package txn

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/manifest"
)

// Limits guard against malicious or broken archives, per StageArchive call.
type Limits struct {
	MaxTotalBytes int64 // bytes actually written; 0 = no limit
}

// LimitsForInstallSize returns the limits for a component whose manifest
// install-size is installSize bytes: 110% of it plus 1 MiB of slack.
func LimitsForInstallSize(installSize int64) Limits {
	installSize = max(installSize, 0)
	return Limits{MaxTotalBytes: installSize + installSize/10 + 1<<20}
}

// StageArchive extracts the zip at zipPath into staging. Entry names are
// taken relative to destPrefix (an FP-relative folder, "" = FP root); legacy
// backslash separators are accepted. It returns the FP-relative paths of the
// files it will place, in archive order.
//
// Directory entries are not created (like the legacy manager). Symlinks and
// other special entries, unsafe names, case-insensitive duplicates and
// archives over lim are rejected; CRC errors surface as zip.ErrChecksum. On
// error nothing from this archive stays staged; earlier archives do.
//
// onProgress (may be nil) gets bytes written so far and the declared total.
func (tx *Tx) StageArchive(ctx context.Context, zipPath, destPrefix string, lim Limits, onProgress func(done, total int64)) ([]string, error) {
	if tx.done {
		return nil, errFinished
	}
	if lim.MaxTotalBytes <= 0 {
		lim.MaxTotalBytes = math.MaxInt64
	}
	// Messages name the archive, not where it was downloaded to.
	zipName := filepath.Base(zipPath)
	prefix := strings.ReplaceAll(destPrefix, `\`, "/")
	if prefix != "" {
		if err := CheckPath(prefix); err != nil {
			return nil, fmt.Errorf("%w: destination folder %q: %w", ErrUnsafeEntry, destPrefix, err)
		}
		prefix += "/"
	}
	// ErrInsecurePath still returns a usable reader; it also flags the
	// backslashes legacy zips use. Every name is validated below.
	zr, err := zip.OpenReader(zipPath)
	if err != nil && !errors.Is(err, zip.ErrInsecurePath) {
		return nil, fmt.Errorf("open archive %s: %w", zipName, sysErr(err))
	}
	defer zr.Close()
	var total int64 // declared, for progress only: the cap is on bytes written
	for _, f := range zr.File {
		total += int64(f.UncompressedSize64)
	}

	var staged []*stagedFile
	mine := map[string]bool{} // lowercased rels of this archive
	undo := func(err error) ([]string, error) {
		for _, sf := range staged {
			tx.root.Remove(sf.staged)
		}
		return nil, err
	}
	var done int64
	for _, f := range zr.File {
		bad := func(why string) ([]string, error) {
			return undo(fmt.Errorf("%w: %s: entry %q: %s", ErrUnsafeEntry, zipName, f.Name, why))
		}
		name := strings.ReplaceAll(f.Name, `\`, "/")
		mode := f.Mode()
		// The raw check catches Unix symlink bits whatever the creator OS.
		if mode&fs.ModeSymlink != 0 || (f.ExternalAttrs>>16)&0o170000 == 0o120000 {
			return bad("symbolic links are not allowed")
		}
		isDir := strings.HasSuffix(name, "/") || mode.IsDir()
		name = strings.TrimSuffix(name, "/")
		if err := CheckPath(name); err != nil {
			return bad(err.Error())
		}
		if isDir {
			continue
		}
		if !mode.IsRegular() {
			return bad("not a regular file")
		}
		rel := prefix + name
		key := strings.ToLower(rel)
		if mine[key] {
			return undo(fmt.Errorf("%w: %s contains %q twice (names differing only in case are the same file on Windows)", ErrConflict, zipName, rel))
		}
		if sf := tx.places[key]; sf != nil && (sf.meta || !tx.replace) {
			return undo(fmt.Errorf("%w: %s (in %s) is also provided by %s", ErrConflict, rel, zipName, sf.source))
		}
		mine[key] = true
		tx.seq++
		sf := &stagedFile{rel: rel, staged: fmt.Sprintf("%s/new/%06d", tx.dir, tx.seq), source: zipName}
		staged = append(staged, sf)
		n, err := tx.extract(ctx, f, sf.staged, lim.MaxTotalBytes-done, func(n int64) {
			if onProgress != nil {
				onProgress(done+n, total)
			}
		})
		done += n
		if err != nil {
			return undo(fmt.Errorf("extract %s from %s: %w", rel, zipName, sysErr(err)))
		}
	}

	// Only now, with the whole archive staged, replace earlier copies.
	files := make([]string, len(staged))
	for i, sf := range staged {
		key := strings.ToLower(sf.rel)
		if old := tx.places[key]; old != nil {
			tx.log.Info("file provided by two archives; the later one wins", "path", sf.rel, "earlier", old.source, "later", zipName)
			tx.root.Remove(old.staged)
		}
		tx.places[key] = sf
		tx.files = append(tx.files, sf)
		files[i] = sf.rel
	}
	return files, nil
}

// extract writes one entry to staged, reading to EOF so archive/zip verifies
// size and CRC. Header sizes are not trusted: at most remaining bytes are
// written. It returns the bytes written.
func (tx *Tx) extract(ctx context.Context, f *zip.File, staged string, remaining int64, progress func(int64)) (int64, error) {
	rc, err := f.Open()
	if err != nil {
		return 0, err
	}
	defer rc.Close()
	out, err := tx.root.OpenFile(staged, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return 0, err
	}
	defer out.Close()
	var written int64
	buf := make([]byte, 256<<10)
	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		n, rerr := rc.Read(buf)
		if written += int64(n); written > remaining {
			return written, fmt.Errorf("%w: it unpacks to more than its size in the manifest allows", ErrArchiveLimit)
		}
		if _, err := out.Write(buf[:n]); err != nil {
			return written, err
		}
		progress(written)
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return written, rerr // includes zip.ErrChecksum and zip.ErrFormat
		}
	}
	if err := out.Close(); err != nil {
		return written, err
	}
	if mod := localDOSTime(f.Modified); !mod.IsZero() {
		_ = tx.root.Chtimes(staged, mod, mod) // nice to have
	}
	return written, nil
}

// localDOSTime fixes archive/zip's reading of an entry with only an MS-DOS
// time: that is local wall-clock time (what zip tools on Windows write),
// but archive/zip returns it as UTC. With an extended timestamp the
// location is a fixed zone instead, and t is right as it is.
func localDOSTime(t time.Time) time.Time {
	if t.IsZero() || t.Location() != time.UTC {
		return t
	}
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, time.Local)
}

// Unstage drops a file that StageArchive staged at rel, so Commit leaves
// whatever is at rel alone (the engine keeps user data such as
// Data/Playlists/Favorites.json this way). It reports whether such a file
// was staged; StageFile files are not affected.
func (tx *Tx) Unstage(rel string) bool {
	key := strings.ToLower(rel)
	sf := tx.places[key]
	if tx.done || sf == nil || sf.meta {
		return false
	}
	delete(tx.places, key)
	tx.root.Remove(sf.staged)
	return true
}

// StageFile stages a small file (e.g. a Components/<id> record; rel clean
// and FP-relative), fsynced. StageFile files are placed after all archive
// files.
func (tx *Tx) StageFile(rel string, data []byte) error {
	if tx.done {
		return errFinished
	}
	key := strings.ToLower(rel)
	if sf := tx.places[key]; sf != nil {
		return fmt.Errorf("%w: %s is staged twice (also by %s)", ErrConflict, rel, sf.source)
	}
	tx.seq++
	sf := &stagedFile{rel: rel, staged: fmt.Sprintf("%s/new/%06d", tx.dir, tx.seq), source: "StageFile", meta: true}
	if err := writeSynced(tx.root, sf.staged, data); err != nil {
		tx.root.Remove(sf.staged)
		return fmt.Errorf("stage %s: %w", rel, sysErr(err))
	}
	tx.places[key] = sf
	tx.files = append(tx.files, sf)
	return nil
}

func writeSynced(r *os.Root, name string, data []byte) error {
	f, err := r.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	_, err = f.Write(data)
	if err == nil {
		err = f.Sync()
	}
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return err // a single *fs.PathError, so sysErr can strip it
}

// sysErr drops the *fs.PathError or *os.LinkError around err: their text
// names the staging or download path, which means nothing to the user.
// Callers name the Flashpoint-relative file instead. errors.Is on the
// underlying error (e.g. a Windows sharing violation) still works.
func sysErr(err error) error {
	switch e := err.(type) {
	case *fs.PathError:
		return e.Err
	case *os.LinkError:
		return e.Err
	}
	return err
}

// CheckPath reports whether rel is a safe FP-relative path in canonical
// form (slash-separated, no empty, "." or ".." segments). The rules are
// manifest.CleanRelPath's, so there is a single validator.
func CheckPath(rel string) error {
	p, err := manifest.CleanRelPath(rel)
	if err != nil {
		return err
	}
	if p == "" || p != rel {
		return fmt.Errorf("%q is not a clean slash-separated path", rel)
	}
	return nil
}

// Package state reads and writes the on-disk state of a Flashpoint install
// that legacy tools and the launcher depend on:
//
//   - <root>/Components/<id>             per-component record (see Record)
//   - <root>/Components/components.bak   raw copy of the last manifest
//   - <root>/Manager/fpm.cfg             line 1: FP root (may be relative), line 2: manifest URL
//
// FP-relative paths are slash-separated internally and converted to the
// legacy backslash form only when (de)serializing.
package state

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/manifest"
)

// Record is the content of Components/<id>. Legacy format (C# manager):
//
//	<hash> <install-size>[ <dep-id>]...\r\n
//	<FP-relative path with backslashes>\r\n
//	...
type Record struct {
	ID          string
	Hash        string // manifest hash at install time (version marker)
	InstallSize int64
	Depends     []string
	Files       []string // slash-separated, FP-relative, in file order

	// Warnings name the path lines ParseRecord skipped because they are
	// unsafe (uninstall must never delete outside the install), for logging.
	Warnings []string
}

// Config is Manager/fpm.cfg.
type Config struct {
	RootRaw string // line 1 as stored, may be relative to Manager (e.g. `..\`)
	Source  string // line 2: the manifest URL
}

const (
	BackupManifestName = "components.bak"
	// PrivateDirName (Components/.fpm/) holds the new manager's own state.
	// It must be a directory: the C# manager treats every top-level file in
	// Components/ as a record but ignores subdirectories.
	PrivateDirName    = ".fpm"
	ComponentsDirName = "Components"
)

// ComponentsDir returns <root>/Components.
func ComponentsDir(root string) string { return filepath.Join(root, ComponentsDirName) }

// RecordPath returns <root>/Components/<id>; id must be a valid ID.
func RecordPath(root, id string) string { return filepath.Join(root, ComponentsDirName, id) }

// PrivateDir returns <root>/Components/.fpm, creating it if needed.
func PrivateDir(root string) (string, error) {
	dir := filepath.Join(ComponentsDir(root), PrivateDirName)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("creating %s: %w", dir, err)
	}
	return dir, nil
}

// ReadBackupManifest returns Components/components.bak, the legacy offline
// copy of the last manifest. It is unsigned: never install from it unverified.
func ReadBackupManifest(root string) ([]byte, error) {
	return readLimited(filepath.Join(ComponentsDir(root), BackupManifestName), manifest.MaxManifestSize)
}

// WriteBackupManifest atomically replaces Components/components.bak, which
// old tools still read.
func WriteBackupManifest(root string, data []byte) error {
	if err := os.MkdirAll(ComponentsDir(root), 0o755); err != nil {
		return err
	}
	return WriteFileAtomic(filepath.Join(ComponentsDir(root), BackupManifestName), data, 0o644)
}

// WriteFileAtomic replaces path with data so that readers and a crash see
// the old or the new content, never a truncated file: temp file in the same
// folder, fsync, rename. The temp name starts with "." so a leftover is
// never mistaken for a component record.
func WriteFileAtomic(path string, data []byte, perm fs.FileMode) error {
	f, err := os.CreateTemp(filepath.Dir(path), "."+filepath.Base(path)+".tmp-*")
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
	if err == nil {
		err = os.Chmod(f.Name(), perm)
	}
	if err == nil {
		err = os.Rename(f.Name(), path)
	}
	if err != nil {
		os.Remove(f.Name())
	}
	return err
}

// readLimited reads a whole file but refuses files larger than limit.
func readLimited(path string, limit int64) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, limit+1))
	if err == nil && int64(len(data)) > limit {
		err = fmt.Errorf("%s is larger than %d bytes", path, limit)
	}
	return data, err
}

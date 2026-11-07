package txn

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

var fixedTime = time.Date(2024, 5, 6, 7, 8, 10, 0, time.UTC)

type zent struct {
	name   string
	data   []byte
	mode   fs.FileMode // 0 = regular 0644
	method uint16      // 0 = Store
	attrs  uint32      // raw ExternalAttrs override (with creator MS-DOS)
}

func writeZip(t *testing.T, dir, name string, ents []zent) string {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, e := range ents {
		h := &zip.FileHeader{Name: e.name, Method: e.method, Modified: fixedTime}
		if e.mode != 0 {
			h.SetMode(e.mode)
		}
		if e.attrs != 0 {
			h.CreatorVersion = 0
			h.ExternalAttrs = e.attrs
		}
		w, err := zw.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		w.Write(e.data)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, buf.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func writeTree(t *testing.T, root string, files map[string]string, dirs ...string) {
	t.Helper()
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(root, filepath.FromSlash(d)), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	for rel, data := range files {
		p := filepath.Join(root, filepath.FromSlash(rel))
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
		os.Chtimes(p, fixedTime, fixedTime)
	}
}

// snapshot maps every path under root (except our working dir) to "dir" or
// the file's hash, plus its mtime if withMtime.
func snapshot(t *testing.T, root string, withMtime bool) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)
		switch {
		case rel == ".":
			return nil
		case rel == "Components/.fpm":
			return filepath.SkipDir
		case d.IsDir():
			out[rel] = "dir"
			return nil
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		v := fmt.Sprintf("%x", sha256.Sum256(data))
		if withMtime {
			fi, _ := d.Info()
			v += " " + fi.ModTime().UTC().Format(time.RFC3339Nano)
		}
		out[rel] = v
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func assertSnap(t *testing.T, label string, got, want map[string]string) {
	t.Helper()
	for _, k := range slices.Sorted(maps.Keys(want)) {
		if got[k] != want[k] {
			t.Errorf("%s: %s: got %q, want %q", label, k, got[k], want[k])
		}
	}
	for k := range got {
		if _, ok := want[k]; !ok {
			t.Errorf("%s: unexpected %s", label, k)
		}
	}
}

func assertNoTxn(t *testing.T, root string) {
	t.Helper()
	if entries, _ := os.ReadDir(filepath.Join(root, filepath.FromSlash(txnBase))); len(entries) > 0 {
		t.Errorf("transaction dirs left: %v", entries)
	}
}

func readFile(t *testing.T, root, rel string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

type fixture struct{ root, zip string }

func newFixture(t *testing.T) fixture {
	t.Helper()
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"Data/keep.txt":        "keep",
		"Data/replace.txt":     "old replace",
		"Data/Sub/old.txt":     "old sub",
		"Launcher/app.exe":     "old exe",
		"Components/comp":      "old record",
		"Components/old-comp":  "old-comp record",
		"Components/unrelated": "unrelated record",
	}, "Data/Empty")
	z := writeZip(t, t.TempDir(), "comp.zip", []zent{
		{name: "Data/New/"},
		{name: "Data/replace.txt", data: []byte("new replace")},
		{name: `Data\New\Deep\new.txt`, data: []byte("new deep"), method: zip.Deflate},
		{name: "Launcher/app.exe", data: []byte("new exe")},
	})
	return fixture{root, z}
}

// run performs the fixture transaction: replace, add (with new folders),
// remove (with a folder that becomes empty), records last. crashAt >= 0
// simulates a crash at that crash point; *points counts the points passed.
func (f fixture) run(t *testing.T, crashAt int, points *int, hook func(*Tx), onLocked func(LockedFile) Decision) error {
	t.Helper()
	tx, err := Begin(f.root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	tx.crash = func() bool {
		n++
		if points != nil {
			*points = n
		}
		return n-1 == crashAt
	}
	tx.retryDelay = time.Millisecond
	tx.quickDelays = []time.Duration{time.Millisecond}
	if hook != nil {
		hook(tx)
	}
	files, err := tx.StageArchive(context.Background(), f.zip, "", Limits{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"Data/replace.txt", "Data/New/Deep/new.txt", "Launcher/app.exe"}; !slices.Equal(files, want) {
		t.Fatalf("files = %q, want %q", files, want)
	}
	tx.StageFile("Components/comp", []byte("new record"))
	tx.StageFile("Components/new-comp", []byte("new-comp record"))
	for _, rel := range []string{"Components/old-comp", "Data/Sub/old.txt", "Data/missing.txt", "Data/Empty/missing.txt", "Data/replace.txt"} {
		tx.Remove(rel)
	}
	return tx.Commit(context.Background(), onLocked)
}

func wantAfter() map[string]string {
	h := func(s string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(s))) }
	return map[string]string{
		"Data":                  "dir",
		"Data/keep.txt":         h("keep"),
		"Data/replace.txt":      h("new replace"),
		"Data/Empty":            "dir", // nothing was removed from it: not ours to prune
		"Data/New":              "dir",
		"Data/New/Deep":         "dir",
		"Data/New/Deep/new.txt": h("new deep"),
		"Launcher":              "dir",
		"Launcher/app.exe":      h("new exe"),
		"Components":            "dir",
		"Components/comp":       h("new record"),
		"Components/new-comp":   h("new-comp record"),
		"Components/unrelated":  h("unrelated record"),
		// Data/Sub pruned, Components/old-comp removed.
	}
}

// renameFails makes renames for which fail(from, to) is true fail with err.
func renameFails(fail func(a, b string) bool, err error) func(*Tx) {
	return func(tx *Tx) {
		real := tx.rename
		tx.rename = func(a, b string) error {
			if fail(a, b) {
				return &os.LinkError{Op: "rename", Old: a, New: b, Err: err}
			}
			return real(a, b)
		}
	}
}

var errBoom = errors.New("disk on fire")

func TestCommitSuccess(t *testing.T) {
	f := newFixture(t)
	if err := f.run(t, -1, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	assertSnap(t, "after", snapshot(t, f.root, false), wantAfter())
	assertNoTxn(t, f.root)
	if fi, _ := os.Stat(filepath.Join(f.root, "Data", "New", "Deep", "new.txt")); !fi.ModTime().Equal(fixedTime) {
		t.Errorf("mtime = %v, want %v from zip", fi.ModTime(), fixedTime)
	}
	if did, err := Recover(f.root, nil); did || err != nil {
		t.Errorf("Recover after clean commit = %v, %v", did, err)
	}
}

// TestCrashConsistency crashes Commit at every crash point (after the
// journal write, after each rename, after "committed"), then recovers, and
// checks that the install is exactly the before or the after state.
func TestCrashConsistency(t *testing.T) {
	ref := newFixture(t)
	before := snapshot(t, ref.root, true)
	var points int
	if err := ref.run(t, -1, &points, nil, nil); err != nil {
		t.Fatal(err)
	}
	committed := points - 1 // the last point is after the committed journal write
	for k := range points {
		t.Run(fmt.Sprint(k), func(t *testing.T) {
			f := newFixture(t)
			if err := f.run(t, k, nil, nil, nil); !errors.Is(err, errCrash) {
				t.Fatalf("Commit = %v, want simulated crash", err)
			}
			want, mtime := before, true
			if k == committed {
				want, mtime = wantAfter(), false
			} else if _, err := Begin(f.root, Options{}); !errors.Is(err, ErrNeedsRecovery) {
				t.Fatalf("Begin with pending txn = %v, want ErrNeedsRecovery", err)
			}
			did, err := Recover(f.root, nil)
			if err != nil || did != (k < committed) {
				t.Fatalf("Recover = %v, %v", did, err)
			}
			assertSnap(t, "recovered", snapshot(t, f.root, mtime), want)
			assertNoTxn(t, f.root)
		})
	}
}

func TestRollback(t *testing.T) {
	f := newFixture(t)
	before := snapshot(t, f.root, true)
	err := f.run(t, -1, nil, renameFails(func(_, b string) bool { return b == "Components/new-comp" }, errBoom), nil)
	if !errors.Is(err, errBoom) || errors.Is(err, ErrRollbackFailed) {
		t.Fatalf("Commit = %v, want wrapped errBoom", err)
	}
	assertSnap(t, "after rollback", snapshot(t, f.root, true), before)
	assertNoTxn(t, f.root)
}

// A staged file that vanishes mid-commit (antivirus) must not make rollback
// move away the live file it was meant to replace.
func TestRollbackKeepsUntouchedLiveFile(t *testing.T) {
	f := newFixture(t)
	before := snapshot(t, f.root, true)
	hook := func(tx *Tx) {
		real := tx.rename
		tx.rename = func(a, b string) error {
			if a == "Data/replace.txt" { // first place op: quarantine a later one, then fail
				tx.root.Remove(tx.places["launcher/app.exe"].staged)
				return errBoom
			}
			return real(a, b)
		}
	}
	if err := f.run(t, -1, nil, hook, nil); !errors.Is(err, errBoom) || errors.Is(err, ErrRollbackFailed) {
		t.Fatalf("Commit = %v, want wrapped errBoom", err)
	}
	assertSnap(t, "after rollback", snapshot(t, f.root, true), before)
	assertNoTxn(t, f.root)
}

func TestRollbackFailureKeepsJournal(t *testing.T) {
	f := newFixture(t)
	before := snapshot(t, f.root, true)
	// Fail the last forward op, and every attempt to restore app.exe.
	hook := renameFails(func(a, b string) bool {
		return b == "Components/new-comp" || b == "Launcher/app.exe" && strings.Contains(a, "/backup/")
	}, errBoom)
	if err := f.run(t, -1, nil, hook, nil); !errors.Is(err, ErrRollbackFailed) || !errors.Is(err, errBoom) {
		t.Fatalf("Commit = %v, want ErrRollbackFailed", err)
	}
	if _, err := Begin(f.root, Options{}); !errors.Is(err, ErrNeedsRecovery) {
		t.Fatalf("Begin = %v, want ErrNeedsRecovery", err)
	}
	if did, err := Recover(f.root, nil); !did || err != nil {
		t.Fatalf("Recover = %v, %v", did, err)
	}
	assertSnap(t, "recovered", snapshot(t, f.root, true), before)
	assertNoTxn(t, f.root)
}

func TestStagedFileMissingBeforeCommit(t *testing.T) {
	f := newFixture(t)
	before := snapshot(t, f.root, true)
	tx, _ := Begin(f.root, Options{})
	if _, err := tx.StageArchive(context.Background(), f.zip, "", Limits{}, nil); err != nil {
		t.Fatal(err)
	}
	tx.root.Remove(tx.places["launcher/app.exe"].staged)
	if err := tx.Commit(context.Background(), nil); err == nil || !strings.Contains(err.Error(), "antivirus") {
		t.Fatalf("Commit = %v, want missing staged file error", err)
	}
	assertSnap(t, "after", snapshot(t, f.root, true), before)
	assertNoTxn(t, f.root)
}

func TestRecoverGarbage(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		txnBase + "/staging-only/new/000001": "x",
		txnBase + "/stray-file":              "x",
		txnBase + "/corrupt/journal.json":    "{not json",
	})
	before := snapshot(t, root, true)
	if did, err := Recover(root, nil); did || err != nil {
		t.Fatalf("Recover = %v, %v; garbage is not an interrupted update", did, err)
	}
	assertSnap(t, "after", snapshot(t, root, true), before)
	assertNoTxn(t, root)

	// Corrupt journal with backups present: keep everything, report.
	writeTree(t, root, map[string]string{
		txnBase + "/corrupt/journal.json":  "{not json",
		txnBase + "/corrupt/backup/000001": "precious",
	})
	if _, err := Recover(root, nil); !errors.Is(err, ErrRollbackFailed) {
		t.Fatalf("Recover = %v, want ErrRollbackFailed", err)
	}
	if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(txnBase), "corrupt", "backup", "000001")); err != nil {
		t.Fatal("backup was deleted")
	}
}

// Like the backup of a running exe on Windows: the leftovers of a committed
// update cannot be deleted. That is not an error, not "recovered", and does
// not block the next Begin.
func TestCleanupFailureIsRetried(t *testing.T) {
	if runtime.GOOS == "windows" || os.Getuid() == 0 {
		t.Skip("needs Unix permissions to make a dir undeletable")
	}
	f := newFixture(t)
	var points int
	f.run(t, -1, &points, nil, nil)
	f = newFixture(t)
	if err := f.run(t, points-1, nil, nil, nil); !errors.Is(err, errCrash) {
		t.Fatal(err)
	}
	dirs, _ := os.ReadDir(filepath.Join(f.root, filepath.FromSlash(txnBase)))
	backup := filepath.Join(f.root, filepath.FromSlash(txnBase), dirs[0].Name(), "backup")
	os.Chmod(backup, 0o500)
	t.Cleanup(func() { os.Chmod(backup, 0o755) })
	if did, err := Recover(f.root, nil); did || err != nil {
		t.Fatalf("Recover = %v, %v; want false, nil", did, err)
	}
	assertSnap(t, "after", snapshot(t, f.root, false), wantAfter())
	if _, err := os.Stat(backup); err != nil {
		t.Fatal("leftover should still be there")
	}
	os.Chmod(backup, 0o755)
	tx, err := Begin(f.root, Options{})
	if err != nil {
		t.Fatal(err)
	}
	tx.Abort()
	assertNoTxn(t, f.root)
}

func TestLocked(t *testing.T) {
	appLocked := func(limit int) func(*Tx) {
		n := 0
		return func(tx *Tx) {
			renameFails(func(a, _ string) bool {
				if a != "Launcher/app.exe" {
					return false
				}
				n++
				return n <= limit
			}, lockedErrno())(tx)
		}
	}
	t.Run("transient", func(t *testing.T) {
		f := newFixture(t)
		// The quick retry absorbs one failure: onLocked is never asked.
		err := f.run(t, -1, nil, appLocked(1), func(LockedFile) Decision { t.Error("onLocked called"); return Abort })
		if err != nil {
			t.Fatal(err)
		}
	})
	t.Run("retry", func(t *testing.T) {
		f := newFixture(t)
		calls := 0
		hook := func(tx *Tx) {
			renameFails(func(a, _ string) bool { return a == "Launcher/app.exe" && calls < 2 }, lockedErrno())(tx)
		}
		err := f.run(t, -1, nil, hook, func(lf LockedFile) Decision {
			calls++
			if lf.Path != "Launcher/app.exe" || !isLocked(lf.Err) {
				t.Errorf("LockedFile = %+v", lf)
			}
			return Retry
		})
		if err != nil || calls != 2 {
			t.Fatalf("Commit = %v, onLocked calls = %d, want 2", err, calls)
		}
		assertSnap(t, "after", snapshot(t, f.root, false), wantAfter())
	})
	for _, onLocked := range []func(LockedFile) Decision{nil, func(LockedFile) Decision { return Abort }} {
		f := newFixture(t)
		before := snapshot(t, f.root, true)
		err := f.run(t, -1, nil, appLocked(1<<30), onLocked)
		if !errors.Is(err, ErrLocked) || !isLocked(err) {
			t.Fatalf("Commit = %v, want ErrLocked", err)
		}
		// The message names the file, never the staging or backup path.
		if msg := err.Error(); !strings.Contains(msg, "Launcher/app.exe") || strings.Contains(msg, ".fpm") {
			t.Errorf("message %q", msg)
		}
		assertSnap(t, "after abort", snapshot(t, f.root, true), before)
		assertNoTxn(t, f.root)
	}
}

// commitOne stages one archive into root and commits.
func commitOne(t *testing.T, root string, ents []zent) error {
	t.Helper()
	tx, _ := Begin(root, Options{})
	if _, err := tx.StageArchive(context.Background(), writeZip(t, t.TempDir(), "x.zip", ents), "", Limits{}, nil); err != nil {
		t.Fatal(err)
	}
	return tx.Commit(context.Background(), nil)
}

func TestCommitConflicts(t *testing.T) {
	for name, ents := range map[string][]zent{
		"file over folder": {{name: "Launcher/app.exe.d", data: []byte("x")}},
		"file under file":  {{name: "Launcher/app.exe/x", data: []byte("x")}},
	} {
		t.Run(name, func(t *testing.T) {
			f := newFixture(t)
			writeTree(t, f.root, nil, "Launcher/app.exe.d")
			before := snapshot(t, f.root, true)
			if err := commitOne(t, f.root, ents); !errors.Is(err, ErrConflict) {
				t.Fatalf("Commit = %v, want ErrConflict", err)
			}
			assertSnap(t, "after", snapshot(t, f.root, true), before)
			assertNoTxn(t, f.root)
		})
	}
}

func TestCommitCancelledBeforeStart(t *testing.T) {
	f := newFixture(t)
	before := snapshot(t, f.root, true)
	tx, _ := Begin(f.root, Options{})
	tx.StageArchive(context.Background(), f.zip, "", Limits{}, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := tx.Commit(ctx, nil); !errors.Is(err, context.Canceled) {
		t.Fatalf("Commit = %v", err)
	}
	assertSnap(t, "after", snapshot(t, f.root, true), before)
	assertNoTxn(t, f.root)
}

func TestReplaceDifferentCase(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "X"), nil, 0o644)
	if _, err := os.Stat(filepath.Join(dir, "x")); err != nil {
		t.Skip("case-sensitive filesystem")
	}
	root := t.TempDir()
	writeTree(t, root, map[string]string{"Launcher/App.EXE": "old"})
	if err := commitOne(t, root, []zent{{name: "launcher/app.exe", data: []byte("new")}}); err != nil {
		t.Fatal(err)
	}
	entries, _ := os.ReadDir(filepath.Join(root, "Launcher"))
	if len(entries) != 1 || readFile(t, root, "Launcher/"+entries[0].Name()) != "new" {
		t.Fatalf("entries = %v", entries)
	}
}

func TestEmptyCommitAndFinish(t *testing.T) {
	root := t.TempDir()
	tx, _ := Begin(root, Options{})
	tx.Remove("nothing/here")
	if err := tx.Commit(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(context.Background(), nil); err == nil {
		t.Error("second Commit succeeded")
	}
	if _, err := tx.StageArchive(context.Background(), "x", "", Limits{}, nil); err == nil {
		t.Error("StageArchive after Commit succeeded")
	}
	tx.Abort() // no-op
	assertNoTxn(t, root)
	if _, err := os.Stat(filepath.Join(root, "nothing")); err == nil {
		t.Error("created folders for a remove")
	}
}

// --- staging -----------------------------------------------------------------

// stageOne stages a single archive into a fresh install and checks that a
// failure leaves nothing behind.
func stageOne(t *testing.T, ents []zent, prefix string, lim Limits) error {
	t.Helper()
	root := t.TempDir()
	writeTree(t, root, map[string]string{"Data/keep.txt": "keep"}, "Components")
	before := snapshot(t, root, true)
	tx, _ := Begin(root, Options{})
	_, err := tx.StageArchive(context.Background(), writeZip(t, t.TempDir(), "c.zip", ents), prefix, lim, nil)
	if err != nil {
		if len(tx.places) != 0 {
			t.Errorf("staged after error: %v", tx.places)
		}
		if staged, _ := fs.ReadDir(tx.root.FS(), tx.dir+"/new"); len(staged) != 0 {
			t.Errorf("staging not cleaned: %v", staged)
		}
	}
	tx.Abort()
	assertSnap(t, "install", snapshot(t, root, true), before)
	assertNoTxn(t, root)
	return err
}

func TestStageRejects(t *testing.T) {
	for name, tc := range map[string]struct {
		ents   []zent
		prefix string
		want   error
	}{
		"dotdot":        {ents: []zent{{name: "ok.txt"}, {name: "../evil"}}, want: ErrUnsafeEntry},
		"backslash":     {ents: []zent{{name: `a\..\..\evil`}}, want: ErrUnsafeEntry},
		"absolute":      {ents: []zent{{name: "/abs"}}, want: ErrUnsafeEntry},
		"drive":         {ents: []zent{{name: `c:\x`}}, want: ErrUnsafeEntry},
		"device":        {ents: []zent{{name: "a/aux"}}, want: ErrUnsafeEntry},
		"dir dotdot":    {ents: []zent{{name: "../dir/"}}, want: ErrUnsafeEntry},
		"bad prefix":    {ents: []zent{{name: "a"}}, prefix: "a/../..", want: ErrUnsafeEntry},
		"symlink":       {ents: []zent{{name: "link", data: []byte("/etc/passwd"), mode: fs.ModeSymlink | 0o777}}, want: ErrUnsafeEntry},
		"symlink attr":  {ents: []zent{{name: "link", data: []byte("../x"), attrs: 0o120777 << 16}}, want: ErrUnsafeEntry},
		"fifo":          {ents: []zent{{name: "fifo", mode: fs.ModeNamedPipe | 0o644}}, want: ErrUnsafeEntry},
		"case dup":      {ents: []zent{{name: "Dir/File.txt"}, {name: `dir\file.TXT`}}, want: ErrConflict},
		"too big total": {ents: []zent{{name: "a", data: make([]byte, 600)}, {name: "b", data: make([]byte, 600)}}, want: ErrArchiveLimit},
	} {
		t.Run(name, func(t *testing.T) {
			if err := stageOne(t, tc.ents, tc.prefix, Limits{MaxTotalBytes: 1000}); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestStageAcceptsLegacyNames(t *testing.T) {
	tx, _ := Begin(t.TempDir(), Options{})
	defer tx.Abort()
	z := writeZip(t, t.TempDir(), "c.zip", []zent{
		{name: `Games\Some Game\data.bin`, data: []byte("1")},
		{name: "Games/"},
		{name: "Games/ünïcode ✓.txt", data: []byte("3")},
		{name: "Games/COM10", data: []byte("5")},
	})
	files, err := tx.StageArchive(context.Background(), z, `FPSoftware\Pkg`, Limits{}, nil)
	want := []string{"FPSoftware/Pkg/Games/Some Game/data.bin", "FPSoftware/Pkg/Games/ünïcode ✓.txt", "FPSoftware/Pkg/Games/COM10"}
	if err != nil || !slices.Equal(files, want) {
		t.Errorf("files = %q, %v; want %q", files, err, want)
	}
}

// The cap is on bytes actually written: the header claims 10 bytes, the
// stream inflates to 4 MiB.
func TestStageLyingHeader(t *testing.T) {
	var comp bytes.Buffer
	zw := zip.NewWriter(&comp)
	w, _ := zw.CreateHeader(&zip.FileHeader{Name: "x", Method: zip.Deflate})
	w.Write(make([]byte, 4<<20))
	zw.Close()
	zr, _ := zip.NewReader(bytes.NewReader(comp.Bytes()), int64(comp.Len()))
	raw, _ := zr.File[0].OpenRaw()
	var rawData bytes.Buffer
	rawData.ReadFrom(raw)

	var buf bytes.Buffer
	zw = zip.NewWriter(&buf)
	w, _ = zw.CreateRaw(&zip.FileHeader{Name: "x", Method: zip.Deflate, CRC32: zr.File[0].CRC32,
		CompressedSize64: uint64(rawData.Len()), UncompressedSize64: 10})
	w.Write(rawData.Bytes())
	zw.Close()
	z := filepath.Join(t.TempDir(), "lie.zip")
	os.WriteFile(z, buf.Bytes(), 0o644)

	tx, _ := Begin(t.TempDir(), Options{})
	defer tx.Abort()
	_, err := tx.StageArchive(context.Background(), z, "", Limits{MaxTotalBytes: 1 << 20}, nil)
	if !errors.Is(err, ErrArchiveLimit) && !errors.Is(err, zip.ErrFormat) {
		t.Fatalf("err = %v, want limit or format error", err)
	}
}

// An entry with only an MS-DOS time (no extended timestamp) carries local
// wall-clock time; archive/zip reads it as UTC.
func TestStageDOSTimeIsLocal(t *testing.T) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	h := &zip.FileHeader{Name: "Data/dos.txt"}
	h.ModifiedDate = uint16((2026-1980)<<9 | 11<<5 | 7) // 2026-11-07
	h.ModifiedTime = uint16(8<<11 | 55<<5 | 42/2)       // 08:55:42
	w, _ := zw.CreateHeader(h)
	w.Write([]byte("x"))
	zw.Close()
	z := filepath.Join(t.TempDir(), "dos.zip")
	os.WriteFile(z, buf.Bytes(), 0o644)

	root := t.TempDir()
	os.Mkdir(filepath.Join(root, "Components"), 0o755)
	tx, _ := Begin(root, Options{})
	if _, err := tx.StageArchive(context.Background(), z, "", Limits{}, nil); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, 11, 7, 8, 55, 42, 0, time.Local)
	if fi, err := os.Stat(filepath.Join(root, "Data", "dos.txt")); err != nil || !fi.ModTime().Equal(want) {
		t.Errorf("mtime = %v, want %v", fi.ModTime(), want)
	}
}

func TestStageChecksum(t *testing.T) {
	content := []byte("hello hello hello hello")
	z := writeZip(t, t.TempDir(), "c.zip", []zent{{name: "a.txt", data: content}})
	data, _ := os.ReadFile(z)
	data[bytes.Index(data, content)+3] ^= 0xff
	os.WriteFile(z, data, 0o644)
	tx, _ := Begin(t.TempDir(), Options{})
	defer tx.Abort()
	if _, err := tx.StageArchive(context.Background(), z, "", Limits{}, nil); !errors.Is(err, zip.ErrChecksum) {
		t.Fatalf("err = %v, want zip.ErrChecksum", err)
	}
}

func TestStageCancel(t *testing.T) {
	root := t.TempDir()
	big := bytes.Repeat([]byte("0123456789abcdef"), 1<<16) // 1 MiB
	z := writeZip(t, t.TempDir(), "c.zip", []zent{{name: "a", data: big}, {name: "b", data: big}, {name: "c", data: big}})
	tx, _ := Begin(root, Options{})
	ctx, cancel := context.WithCancel(context.Background())
	calls := 0
	_, err := tx.StageArchive(ctx, z, "Data", Limits{}, func(done, total int64) {
		calls++
		if total != 3<<20 {
			t.Errorf("total = %d", total)
		}
		cancel() // in the middle of the first entry
	})
	if !errors.Is(err, context.Canceled) || calls != 1 {
		t.Fatalf("err = %v, progress calls = %d", err, calls)
	}
	tx.Abort()
	assertNoTxn(t, root)
}

func TestConflictsAcrossArchives(t *testing.T) {
	zdir := t.TempDir()
	z1 := writeZip(t, zdir, "one.zip", []zent{{name: "x/a.txt"}})
	z2 := writeZip(t, zdir, "two.zip", []zent{{name: "ok.txt"}, {name: "A.TXT"}})
	tx, _ := Begin(t.TempDir(), Options{})
	defer tx.Abort()
	if _, err := tx.StageArchive(context.Background(), z1, "", Limits{}, nil); err != nil {
		t.Fatal(err)
	}
	_, err := tx.StageArchive(context.Background(), z2, "X", Limits{}, nil)
	if !errors.Is(err, ErrConflict) || !strings.Contains(err.Error(), "one.zip") || !strings.Contains(err.Error(), "two.zip") || strings.Contains(err.Error(), zdir) {
		t.Fatalf("err = %v, want conflict naming both archives (without their folder)", err)
	}
	if tx.places["x/ok.txt"] != nil {
		t.Error("failed archive left entries staged")
	}
	if err := tx.StageFile("x/A.txt", nil); !errors.Is(err, ErrConflict) {
		t.Errorf("StageFile dup = %v", err)
	}
}

func TestLaterArchiveWins(t *testing.T) {
	root := t.TempDir()
	zdir := t.TempDir()
	z1 := writeZip(t, zdir, "one.zip", []zent{{name: "shared/gdawp.ini", data: []byte("one")}, {name: "a.txt", data: []byte("a")}})
	z2 := writeZip(t, zdir, "two.zip", []zent{{name: `shared\gdawp.ini`, data: []byte("two")}, {name: "b.txt", data: []byte("b")}})
	z3 := writeZip(t, zdir, "three.zip", []zent{{name: "shared/gdawp.ini", data: []byte("three")}, {name: "../bad", data: []byte("bad")}})
	tx, _ := Begin(root, Options{LaterArchiveWins: true})
	ctx := context.Background()
	if _, err := tx.StageArchive(ctx, z1, "FP", Limits{}, nil); err != nil {
		t.Fatal(err)
	}
	if files, err := tx.StageArchive(ctx, z2, "FP", Limits{}, nil); err != nil || len(files) != 2 || files[0] != "FP/shared/gdawp.ini" {
		t.Fatalf("overlapping archive: %v, %v", files, err)
	}
	// A failing later archive must not take the earlier copy away.
	if _, err := tx.StageArchive(ctx, z3, "FP", Limits{}, nil); !errors.Is(err, ErrUnsafeEntry) {
		t.Fatalf("bad archive: err = %v", err)
	}
	// StageFile files are never replaced, nor replace archive files.
	if err := tx.StageFile("FP/b.txt", []byte("meta")); !errors.Is(err, ErrConflict) {
		t.Fatalf("StageFile over archive file: %v", err)
	}
	if err := tx.Commit(ctx, nil); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, root, "FP/shared/gdawp.ini"); got != "two" {
		t.Errorf("shared file = %q, want the later archive's copy", got)
	}
	if got := readFile(t, root, "FP/a.txt"); got != "a" {
		t.Errorf("a.txt = %q", got)
	}
	assertNoTxn(t, root)
}

func TestUnstage(t *testing.T) {
	root := t.TempDir()
	writeTree(t, root, map[string]string{"Data/Playlists/Favorites.json": "mine"}, "Components")
	z := writeZip(t, t.TempDir(), "p.zip", []zent{{name: "Favorites.json", data: []byte("[]")}, {name: "Other.json", data: []byte("o")}})
	tx, _ := Begin(root, Options{})
	if _, err := tx.StageArchive(context.Background(), z, "Data/Playlists", Limits{}, nil); err != nil {
		t.Fatal(err)
	}
	if !tx.Unstage("data/playlists/favorites.json") {
		t.Fatal("Unstage = false")
	}
	if tx.Unstage("Data/Playlists/Favorites.json") || tx.Unstage("nope") {
		t.Error("second Unstage = true")
	}
	tx.StageFile("Components/x", []byte("rec"))
	if tx.Unstage("Components/x") {
		t.Error("Unstage removed a StageFile file")
	}
	if err := tx.Commit(context.Background(), nil); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, root, "Data/Playlists/Favorites.json"); got != "mine" {
		t.Errorf("Favorites.json = %q, want untouched", got)
	}
	if got := readFile(t, root, "Data/Playlists/Other.json"); got != "o" {
		t.Errorf("Other.json = %q", got)
	}
	assertNoTxn(t, root)
}

func TestCheckPath(t *testing.T) {
	for _, ok := range []string{"a", "Data/Games/x.swf", ".fpm", "a b/c.d", "COM10"} {
		if err := CheckPath(ok); err != nil {
			t.Errorf("CheckPath(%q) = %v", ok, err)
		}
	}
	for _, bad := range []string{"", "a\\b", "/a", "a/", "a//b", "./a", "a/../b", "C:", "PRN.txt"} {
		if err := CheckPath(bad); err == nil {
			t.Errorf("CheckPath(%q) = nil", bad)
		}
	}
}

package state

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// makeFP creates a Flashpoint-like tree under dir and returns its Manager dir.
func makeFP(t *testing.T, dir string, withRecord, withLauncherData bool) string {
	t.Helper()
	mgr := filepath.Join(dir, "Manager")
	must(t, os.MkdirAll(mgr, 0o755))
	must(t, os.MkdirAll(ComponentsDir(dir), 0o755))
	// components.bak alone does not make a root.
	must(t, os.WriteFile(filepath.Join(ComponentsDir(dir), BackupManifestName), []byte("<list/>"), 0o644))
	if withRecord {
		must(t, os.WriteFile(RecordPath(dir, "core-manager"), []byte("0f7ef199 2972160\r\n"), 0o644))
	}
	if withLauncherData {
		must(t, os.MkdirAll(filepath.Join(dir, "Launcher"), 0o755))
		must(t, os.MkdirAll(filepath.Join(dir, "Data"), 0o755))
	}
	return mgr
}

func TestFindRoot(t *testing.T) {
	base := t.TempDir()
	fp := filepath.Join(base, "Flashpoint")
	mgr := makeFP(t, fp, true, false)

	other := filepath.Join(base, "Other")
	makeFP(t, other, false, true)

	empty := filepath.Join(base, "Empty")
	emptyMgr := makeFP(t, empty, false, false)

	tests := []struct {
		name         string
		mgr          string
		cfg          *Config
		want         string
		wantRepaired bool
		wantErr      bool
	}{
		{"relative backslash", mgr, &Config{RootRaw: `..\`}, fp, false, false},
		{"relative slash", mgr, &Config{RootRaw: "../"}, fp, false, false},
		{"missing cfg", mgr, nil, fp, false, false},
		{"empty line", mgr, &Config{RootRaw: "  "}, fp, false, false},
		{"absolute ok", mgr, &Config{RootRaw: fp}, fp, false, false},
		{"absolute other root", mgr, &Config{RootRaw: other}, other, false, false},
		{"stale absolute", mgr, &Config{RootRaw: filepath.Join(base, "gone")}, fp, true, false},
		{"stale windows path", mgr, &Config{RootRaw: `C:\Flashpoint`}, fp, true, false},
		{"relative to nothing", mgr, &Config{RootRaw: `..\..\nope`}, fp, true, false},
		{"no root anywhere", emptyMgr, &Config{RootRaw: `..\`}, "", false, true},
		{"no root stale", emptyMgr, &Config{RootRaw: filepath.Join(base, "gone")}, "", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, repaired, err := FindRoot(tt.mgr, tt.cfg)
			if tt.wantErr {
				if !errors.Is(err, ErrRootNotFound) {
					t.Fatalf("got %q, %v; want ErrRootNotFound", got, err)
				}
				return
			}
			if err != nil || got != tt.want || repaired != tt.wantRepaired {
				t.Errorf("got %q repaired=%v err=%v; want %q repaired=%v", got, repaired, err, tt.want, tt.wantRepaired)
			}
		})
	}
}

func TestIsRoot(t *testing.T) {
	base := t.TempDir()
	makeFP(t, filepath.Join(base, "rec"), true, false)
	makeFP(t, filepath.Join(base, "ld"), false, true)
	makeFP(t, filepath.Join(base, "none"), false, false)
	must(t, os.MkdirAll(filepath.Join(base, "onlylauncher", "Launcher"), 0o755))
	for dir, want := range map[string]bool{"rec": true, "ld": true, "none": false, "onlylauncher": false, "missing": false} {
		if got := IsRoot(filepath.Join(base, dir)); got != want {
			t.Errorf("IsRoot(%s) = %v", dir, got)
		}
	}
}

func TestResolveRoot(t *testing.T) {
	mgr := filepath.Join(string(filepath.Separator)+"fp", "Manager")
	if got := ResolveRoot(mgr, `..\`); got != filepath.Clean(filepath.Join(mgr, "..")) {
		t.Errorf(`..\ -> %q`, got)
	}
	if got := ResolveRoot(mgr, ""); got != filepath.Dir(mgr) {
		t.Errorf(`"" -> %q`, got)
	}
	if got := ResolveRoot(mgr, `sub\dir`); got != filepath.Join(mgr, "sub", "dir") {
		t.Errorf(`sub\dir -> %q`, got)
	}
}

package install

import (
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"testing"
)

// fakeInstall is a small Flashpoint folder that also holds two things of
// the user's; it returns root and the running manager's exe in it.
func fakeInstall(t *testing.T) (string, string) {
	root := filepath.Join(t.TempDir(), "Flashpoint")
	files := map[string]string{
		"Components/core-launcher":      "12345678 10\r\nLauncher\\Flashpoint.exe\r\nFPSoftware\\plugin.dll\r\n",
		"Launcher/Flashpoint.exe":       "x",
		"FPSoftware/plugin.dll":         "x",
		"Data/Playlists/Favorites.json": "x",
		"Manager/FlashpointManager.exe": "x",
		"MyStuff/notes.txt":             "the user's",
		"readme-user.txt":               "the user's",
	}
	for f, data := range files {
		p := filepath.Join(root, filepath.FromSlash(f))
		os.MkdirAll(filepath.Dir(p), 0o755)
		if err := os.WriteFile(p, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return root, filepath.Join(root, "Manager", "FlashpointManager.exe")
}

// left lists what is still in root, slash-separated.
func left(root string) string {
	var out []string
	filepath.WalkDir(root, func(p string, d os.DirEntry, err error) error {
		if err == nil && p != root {
			rel, _ := filepath.Rel(root, p)
			out = append(out, filepath.ToSlash(rel))
		}
		return nil
	})
	return strings.Join(out, " ")
}

func TestUninstall(t *testing.T) {
	root, exe := fakeInstall(t)
	kept, err := Uninstall(root, exe)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(kept, []string{"MyStuff", "readme-user.txt"}) {
		t.Errorf("kept %q", kept)
	}
	if got, want := left(root), "Manager Manager/FlashpointManager.exe MyStuff MyStuff/notes.txt readme-user.txt"; got != want {
		t.Errorf("left %q, want %q", got, want)
	}
}

func TestUninstallRefuses(t *testing.T) {
	notFP := t.TempDir()
	os.WriteFile(filepath.Join(notFP, "keep.txt"), []byte("mine"), 0o644)
	if _, err := Uninstall(notFP, ""); err == nil || !strings.Contains(err.Error(), "not a Flashpoint folder") {
		t.Errorf("non-Flashpoint folder: err = %v", err)
	}
	if left(notFP) != "keep.txt" {
		t.Errorf("deleted files of a non-Flashpoint folder: left %q", left(notFP))
	}
	if runtime.GOOS == "windows" { // a drive that doesn't exist: never a real one
		if err := CheckUninstall(`Q:\`); err == nil || !strings.Contains(err.Error(), "a whole drive") {
			t.Errorf("drive root: err = %v", err)
		}
	}
}

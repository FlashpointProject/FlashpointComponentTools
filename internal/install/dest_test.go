package install

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// testDir returns a temp folder and points the env vars of the protected
// places away from it: t.TempDir is in %TEMP% on Windows, which CheckDest
// rejects.
func testDir(t *testing.T) string {
	dir := t.TempDir()
	for _, k := range []string{"ProgramFiles", "ProgramFiles(x86)", "SystemRoot",
		"OneDrive", "OneDriveConsumer", "OneDriveCommercial"} {
		t.Setenv(k, "")
	}
	t.Setenv("TMP", `Q:\Temp`)
	t.Setenv("TEMP", `Q:\Temp`)
	return dir
}

func mkdirs(t *testing.T, paths ...string) {
	for _, p := range paths {
		if err := os.MkdirAll(p, 0o755); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCheckDest(t *testing.T) {
	base := testDir(t)
	file := filepath.Join(base, "file.txt")
	if err := os.WriteFile(file, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(base, "fp")
	mkdirs(t, filepath.Join(root, "Launcher"), filepath.Join(root, "Data"))
	full := filepath.Join(base, "full")
	mkdirs(t, filepath.Join(full, "sub"))
	empty := filepath.Join(base, "empty")
	mkdirs(t, empty)
	newDir := filepath.Join(base, "new", "Flashpoint") // base/new is missing too
	newChild := filepath.Join(base, "Flashpoint")
	driveRoot := "/"
	if runtime.GOOS == "windows" {
		driveRoot = filepath.VolumeName(base) + `\`
	}

	tests := []struct {
		name, path  string
		want        string // error substring, "" = ok
		abs, create string
	}{
		{"empty", "", "Choose a folder for Flashpoint.", "", ""},
		{"blank", "  ", "Choose a folder for Flashpoint.", "", ""},
		{"relative", filepath.Join("games", "fp"), "Enter a full path", "", ""},
		{"too long", filepath.Join(base, strings.Repeat("a", 192)), "This path is too long", "", ""},
		{"drive root", driveRoot, "not a whole drive", "", ""},
		{"file", file, file + " is a file. Choose a folder.", "", ""},
		{"installed", root, "Flashpoint is already installed in this folder.", "", ""},
		{"not empty", full, "This folder isn't empty.", "", ""},
		{"existing empty folder", empty, "", empty, ""},
		{"new under an existing parent", newChild, "", newChild, newChild},
		{"missing parent", newDir, "", newDir, filepath.Join(base, "new")},
		{"cleaned and quoted", ` "` + filepath.Join(base, "new", "x", "..", "Flashpoint") + `" `, "", newDir, filepath.Join(base, "new")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			abs, create, err := CheckDest(tt.path)
			if tt.want != "" {
				if err == nil || !strings.Contains(err.Error(), tt.want) {
					t.Fatalf("CheckDest(%q) = %v, want error containing %q", tt.path, err, tt.want)
				}
				return
			}
			if err != nil || abs != tt.abs || create != tt.create {
				t.Fatalf("CheckDest(%q) = %q, %q, %v; want %q, %q", tt.path, abs, create, err, tt.abs, tt.create)
			}
		})
	}
}

func TestCleanup(t *testing.T) {
	// create is relative to the test folder; the install is parent/Flashpoint.
	for _, create := range []string{filepath.Join("parent", "Flashpoint"), "parent", ""} {
		base := t.TempDir()
		dir := filepath.Join(base, "parent", "Flashpoint")
		mkdirs(t, filepath.Join(dir, "Components"), filepath.Join(dir, "Temp", "fpm"))
		if err := os.WriteFile(filepath.Join(dir, "Components", ".fpm"), nil, 0o644); err != nil {
			t.Fatal(err)
		}
		if create != "" {
			create = filepath.Join(base, create)
		}
		if err := Cleanup(dir, create); err != nil {
			t.Fatalf("Cleanup(create=%q): %v", create, err)
		}
		if create != "" {
			if _, err := os.Stat(create); !os.IsNotExist(err) {
				t.Errorf("Cleanup(create=%q): still there (%v)", create, err)
			}
		} else if entries, err := os.ReadDir(dir); err != nil || len(entries) != 0 {
			t.Errorf("Cleanup(create=\"\"): want an empty folder, got %d entries, %v", len(entries), err)
		}
	}
	// Folder creation failed: nothing to remove.
	dir := filepath.Join(t.TempDir(), "never-created")
	if err := Cleanup(dir, dir); err != nil {
		t.Errorf("Cleanup(never created): %v", err)
	}
}

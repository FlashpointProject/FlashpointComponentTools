package state

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name, content string
		want          Config
	}{
		{"crlf", "..\\\r\nhttps://nexus-dev.unstable.life/repository/stable/components.xml\r\n",
			Config{`..\`, "https://nexus-dev.unstable.life/repository/stable/components.xml"}},
		{"lf", "C:\\Flashpoint\nhttps://x/components.xml\n", Config{`C:\Flashpoint`, "https://x/components.xml"}},
		{"bom no trailing newline", "\xef\xbb\xbfD:\\FP\r\nhttps://x/c.xml", Config{`D:\FP`, "https://x/c.xml"}},
		{"missing line 2", "D:\\FP\r\n", Config{`D:\FP`, DefaultSource}},
		{"only line 1 no newline", "D:\\FP", Config{`D:\FP`, DefaultSource}},
		{"empty", "", Config{DefaultRootRaw, DefaultSource}},
		{"blank lines", "\r\n  \r\n", Config{DefaultRootRaw, DefaultSource}},
		{"source whitespace", "..\r\n  http://local/c.xml  \r\nextra line\r\n", Config{"..", "http://local/c.xml"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			must(t, os.WriteFile(filepath.Join(dir, ConfigName), []byte(tt.content), 0o644))
			got, err := LoadConfig(dir)
			if err != nil || *got != tt.want {
				t.Errorf("got %+v, %v; want %+v", got, err, tt.want)
			}
		})
	}
	got, err := LoadConfig(t.TempDir())
	if err != nil || *got != (Config{"..", DefaultSource}) {
		t.Errorf("missing file: %+v %v", got, err)
	}
}

func TestSaveConfig(t *testing.T) {
	dir := t.TempDir()
	cfg := &Config{RootRaw: `..\`, Source: "https://nexus-dev.unstable.life/repository/stable/components.xml"}
	must(t, SaveConfig(dir, cfg))
	b, err := os.ReadFile(filepath.Join(dir, ConfigName))
	must(t, err)
	if want := "..\\\r\nhttps://nexus-dev.unstable.life/repository/stable/components.xml\r\n"; string(b) != want {
		t.Errorf("wrote %q", b)
	}
	back, err := LoadConfig(dir)
	if err != nil || *back != *cfg {
		t.Errorf("reload: %+v %v", back, err)
	}
	if err := SaveConfig(dir, &Config{RootRaw: "a\nb", Source: "x"}); err == nil {
		t.Error("accepted a newline")
	}
	assertNoTemp(t, dir)
}

func TestWriteFileAtomic(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "f.json")
	must(t, WriteFileAtomic(p, []byte("one"), 0o600))
	must(t, WriteFileAtomic(p, []byte("two"), 0o644)) // replaces an existing file
	b, err := os.ReadFile(p)
	if err != nil || string(b) != "two" {
		t.Fatalf("got %q %v", b, err)
	}
	if fi, _ := os.Stat(p); runtime.GOOS != "windows" && fi.Mode().Perm() != 0o644 {
		t.Errorf("perm = %v", fi.Mode().Perm())
	}
	if err := WriteFileAtomic(filepath.Join(dir, "missing", "f"), nil, 0o644); err == nil {
		t.Error("wrote into a missing directory")
	}
	assertNoTemp(t, dir)
}

func assertNoTemp(t *testing.T, dir string) {
	t.Helper()
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if e.Name()[0] == '.' {
			t.Errorf("temp file left behind: %s", e.Name())
		}
	}
}

func TestBackupAndPrivateDir(t *testing.T) {
	root := t.TempDir()
	if _, err := ReadBackupManifest(root); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("missing backup: %v", err)
	}
	must(t, WriteBackupManifest(root, []byte("<list/>")))
	must(t, WriteBackupManifest(root, []byte("<list url='x'/>")))
	b, err := ReadBackupManifest(root)
	if err != nil || !bytes.Equal(b, []byte("<list url='x'/>")) {
		t.Errorf("backup = %q %v", b, err)
	}
	dir, err := PrivateDir(root)
	must(t, err)
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() || dir != filepath.Join(root, "Components", ".fpm") {
		t.Errorf("PrivateDir %q: %v", dir, err)
	}
	if _, err := PrivateDir(root); err != nil { // idempotent
		t.Error(err)
	}
	// Neither shows up as a record.
	recs, errs := ReadRecords(root)
	if len(recs) != 0 || len(errs) != 0 {
		t.Errorf("records %v errs %v", recs, errs)
	}
}

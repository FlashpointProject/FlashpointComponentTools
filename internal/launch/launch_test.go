package launch

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFindLauncher(t *testing.T) {
	dir := t.TempDir()
	if _, err := findLauncher(dir); err == nil || !strings.Contains(err.Error(), "launcher was not found") {
		t.Errorf("empty folder: err = %v", err)
	}
	want := filepath.Join(dir, launcherName())
	if err := os.WriteFile(want, nil, 0o755); err != nil {
		t.Fatal(err)
	}
	if got, err := findLauncher(dir); err != nil || got != want {
		t.Errorf("findLauncher = %q, %v; want %q", got, err, want)
	}
}

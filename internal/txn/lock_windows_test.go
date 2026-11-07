//go:build windows

package txn

import (
	"path/filepath"
	"testing"

	"golang.org/x/sys/windows"
)

func lockedErrno() error { return windows.ERROR_SHARING_VIOLATION }

// A file opened with share mode 0 (antivirus, some editors) can't even be
// stat'ed; it must still get the retry + onLocked path, not a plain error.
func TestLockedNoSharing(t *testing.T) {
	f := newFixture(t)
	p, _ := windows.UTF16PtrFromString(filepath.Join(f.root, "Launcher", "app.exe"))
	h, err := windows.CreateFile(p, windows.GENERIC_READ, 0, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(h)
	calls := 0
	err = f.run(t, -1, nil, nil, func(lf LockedFile) Decision {
		calls++
		if lf.Path != "Launcher/app.exe" {
			t.Errorf("LockedFile = %+v", lf)
		}
		windows.CloseHandle(h)
		return Retry
	})
	if err != nil || calls != 1 {
		t.Fatalf("Commit = %v, onLocked calls = %d, want 1", err, calls)
	}
	assertSnap(t, "after", snapshot(t, f.root, false), wantAfter())
}

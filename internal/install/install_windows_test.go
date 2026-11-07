package install

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/sys/windows"
)

func TestVerifySignatureUnsigned(t *testing.T) {
	exe := filepath.Join(t.TempDir(), "unsigned.exe")
	if err := os.WriteFile(exe, []byte("MZ not really an exe"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := verifySignature(exe)
	// TRUST_E_SUBJECT_FORM_UNKNOWN or TRUST_E_NOSIGNATURE: WinVerifyTrust
	// looked at the file, so the call itself is set up right.
	if !errors.Is(err, syscall.Errno(windows.TRUST_E_NOSIGNATURE)) &&
		!errors.Is(err, syscall.Errno(windows.TRUST_E_SUBJECT_FORM_UNKNOWN)) {
		t.Fatalf("verifySignature(unsigned) = %v", err)
	}
}

func TestCreateShortcutsRootOnly(t *testing.T) {
	root := t.TempDir()
	if err := CreateShortcuts(root, false, false); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "Flashpoint.lnk")); err != nil {
		t.Fatal(err)
	}
}

func TestCreateShortcutsMissingFolder(t *testing.T) {
	root := filepath.Join(t.TempDir(), "missing")
	// The plain cause from the exception info, not WScript's generic
	// "Exception occurred. (Unable to save shortcut ...)".
	err := CreateShortcuts(root, false, false)
	if err == nil || !strings.HasPrefix(err.Error(), "Flashpoint folder: ") || strings.Contains(err.Error(), "Unable to save") {
		t.Fatalf("CreateShortcuts(missing root) = %v; want \"Flashpoint folder: <cause>\"", err)
	}
	t.Log(err)
}

// A file in use stays; the rest goes, and the error names the file.
func TestUninstallLockedFile(t *testing.T) {
	root, exe := fakeInstall(t)
	p, _ := windows.UTF16PtrFromString(filepath.Join(root, "Launcher", "Flashpoint.exe"))
	h, err := windows.CreateFile(p, windows.GENERIC_READ, 0, nil, windows.OPEN_EXISTING, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer windows.CloseHandle(h)
	_, err = Uninstall(root, exe)
	if err == nil || !strings.HasPrefix(err.Error(), "Could not delete "+filepath.Join(root, "Launcher", "Flashpoint.exe")+": ") {
		t.Fatalf("err = %v", err)
	}
	if got := left(root); !strings.HasPrefix(got, "Launcher Launcher/Flashpoint.exe Manager Manager/FlashpointManager.exe MyStuff") {
		t.Errorf("left %q", got)
	}
}

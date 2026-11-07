package install

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"golang.org/x/sys/windows"
)

// CreateShortcuts creates Flashpoint.lnk (target <root>\Launcher\Flashpoint.exe,
// working folder <root>\Launcher) in root, and on the Desktop / in the Start
// menu's Programs folder when asked. Each failure is "<place>: <reason>",
// joined with "; ".
func CreateShortcuts(root string, desktop, startMenu bool) error {
	type place struct {
		name, dir string // name is how the user knows it
		err       error
	}
	places := []place{{"Flashpoint folder", root, nil}}
	// KnownFolderPath, not %USERPROFILE%\Desktop: the Desktop is often
	// redirected (OneDrive backup, roaming profiles).
	if desktop {
		dir, err := windows.KnownFolderPath(windows.FOLDERID_Desktop, windows.KF_FLAG_DEFAULT)
		places = append(places, place{"Desktop", dir, err})
	}
	if startMenu {
		dir, err := windows.KnownFolderPath(windows.FOLDERID_Programs, windows.KF_FLAG_DEFAULT)
		places = append(places, place{"Start menu", dir, err})
	}

	launcher := filepath.Join(root, "Launcher")
	return withShell(func(shell *ole.IDispatch) error {
		var failed []string
		for _, p := range places {
			if p.err == nil {
				p.err = writeShortcut(shell, filepath.Join(p.dir, "Flashpoint.lnk"), filepath.Join(launcher, "Flashpoint.exe"), launcher)
			}
			if p.err != nil {
				failed = append(failed, p.name+": "+reason(p.err))
			}
		}
		if failed == nil {
			return nil
		}
		return errors.New(strings.Join(failed, "; "))
	})
}

// removeShortcuts deletes the Flashpoint.lnk on the Desktop and in the
// Start menu (where this installer and the old one put them) when it starts
// a launcher inside root: another Flashpoint's shortcut stays. A missing
// .lnk reads as an empty target.
func removeShortcuts(root string) error {
	return withShell(func(shell *ole.IDispatch) error {
		var errs []error
		for _, id := range []*windows.KNOWNFOLDERID{windows.FOLDERID_Desktop, windows.FOLDERID_Programs, windows.FOLDERID_StartMenu} {
			dir, err := windows.KnownFolderPath(id, windows.KF_FLAG_DEFAULT)
			if err != nil {
				continue
			}
			lnk := filepath.Join(dir, "Flashpoint.lnk")
			if target, err := shortcutTarget(shell, lnk); err == nil && target != "" && within(target, root) {
				errs = append(errs, os.Remove(lnk))
			}
		}
		return errors.Join(errs...)
	})
}

// withShell runs f with a WScript.Shell object, the simplest way to read and
// write a .lnk.
func withShell(f func(shell *ole.IDispatch) error) error {
	// COM is per thread: keep this goroutine on the thread it initialized.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED); err != nil {
		// S_FALSE: COM was already initialized on this thread; still
		// balanced by CoUninitialize.
		if oe, ok := err.(*ole.OleError); !ok || oe.Code() != 1 {
			return err
		}
	}
	defer ole.CoUninitialize()
	unk, err := oleutil.CreateObject("WScript.Shell")
	if err != nil {
		return err
	}
	defer unk.Release()
	shell, err := unk.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return err
	}
	defer shell.Release()
	return f(shell)
}

func shortcutTarget(shell *ole.IDispatch, lnk string) (string, error) {
	v, err := oleutil.CallMethod(shell, "CreateShortcut", lnk)
	if err != nil {
		return "", err
	}
	defer v.Clear()
	t, err := oleutil.GetProperty(v.ToIDispatch(), "TargetPath")
	if err != nil {
		return "", err
	}
	defer t.Clear()
	return t.ToString(), nil
}

func writeShortcut(shell *ole.IDispatch, lnk, target, workDir string) error {
	v, err := oleutil.CallMethod(shell, "CreateShortcut", lnk)
	if err != nil {
		return err
	}
	defer v.Clear()
	sc := v.ToIDispatch()
	for _, p := range [][2]string{{"TargetPath", target}, {"WorkingDirectory", workDir}, {"Description", "Flashpoint"}} {
		if _, err := oleutil.PutProperty(sc, p[0], p[1]); err != nil {
			return err
		}
	}
	_, err = oleutil.CallMethod(sc, "Save")
	return err
}

// reason is the plain cause of err. A failed WScript.Shell call says only
// "Exception occurred. (Unable to save shortcut ...)"; the cause ("Access
// is denied.") is the HRESULT in its exception info.
func reason(err error) string {
	var oe *ole.OleError
	if errors.As(err, &oe) {
		if ei, ok := oe.SubError().(ole.EXCEPINFO); ok && ei.SCODE() != 0 {
			return windows.Errno(ei.SCODE()).Error()
		}
	}
	return err.Error()
}

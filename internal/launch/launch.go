// Package launch finds the Flashpoint launcher and starts programs detached
// from the calling one (used by the manager and the installer).
package launch

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// launcherName is the launcher executable in <FP>/Launcher (electron-builder
// productName "Flashpoint"); the other OSes are for development only.
func launcherName() string {
	if runtime.GOOS == "windows" {
		return "Flashpoint.exe"
	}
	return "flashpoint-launcher"
}

// StartLauncher starts the launcher of the Flashpoint folder root, with
// cwd = <root>/Launcher like the C# manager, and does not wait for it.
func StartLauncher(root string) (exe string, err error) {
	dir := filepath.Join(root, "Launcher")
	exe, err = findLauncher(dir)
	if err != nil {
		return "", err
	}
	if err := Start(exe, nil, dir); err != nil {
		return "", fmt.Errorf("start the launcher: %w", err)
	}
	return exe, nil
}

// findLauncher returns the launcher executable in dir (<FP>/Launcher).
func findLauncher(dir string) (string, error) {
	p := filepath.Join(dir, launcherName())
	if fi, err := os.Stat(p); err != nil || !fi.Mode().IsRegular() {
		return "", fmt.Errorf("the Flashpoint launcher was not found in %s", dir)
	}
	return p, nil
}

// Start starts exe in its own session / process group, without inheriting
// stdio, and does not wait for it ("" dir = current folder).
func Start(exe string, args []string, dir string) error {
	cmd := exec.Command(exe, args...)
	cmd.Dir = dir
	cmd.SysProcAttr = DetachedAttr()
	if err := cmd.Start(); err != nil {
		return err
	}
	go cmd.Wait() // reap it if we are still running when it exits
	return nil
}

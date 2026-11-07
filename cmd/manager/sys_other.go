//go:build !windows

package main

import (
	"errors"
	"os"
	"path/filepath"
	"time"
)

// inJob: Windows job objects only (see sys_windows.go).
func inJob() bool { return false }

func startOutsideJob(string, []string) error { return errors.ErrUnsupported }

// waitParent: Unix can delete the file of a running program.
func waitParent(time.Duration) {}

// deleteAfterExit: Unix can delete a running program, so at once.
func deleteAfterExit(exe, root string) error {
	return errors.Join(os.Remove(exe), os.Remove(filepath.Dir(exe)), os.Remove(root))
}

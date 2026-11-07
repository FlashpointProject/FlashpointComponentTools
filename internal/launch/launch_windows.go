package launch

import (
	"syscall"

	"golang.org/x/sys/windows"
)

// DetachedAttr is the SysProcAttr Start uses: no console, and a process
// group of its own so our Ctrl-C does not reach it. A new value each call,
// so callers may add flags.
func DetachedAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP}
}

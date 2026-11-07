//go:build !windows

package launch

import "syscall"

// DetachedAttr is the SysProcAttr Start uses: a session of its own, so it
// survives our exit and our terminal's signals. A new value each call.
func DetachedAttr() *syscall.SysProcAttr { return &syscall.SysProcAttr{Setsid: true} }

//go:build !windows

package txn

import "syscall"

func lockedErrno() error { return syscall.EBUSY }

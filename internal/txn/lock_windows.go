//go:build windows

package txn

import (
	"errors"

	"golang.org/x/sys/windows"
)

// isLocked reports whether err means "in use by another process": renaming
// a file that is open without FILE_SHARE_DELETE fails with one of these.
// Antivirus and indexers cause them briefly too, hence the quick retries.
// ACCESS_DENIED can also be a real permission problem; the ErrLocked
// message covers both.
func isLocked(err error) bool {
	return errors.Is(err, windows.ERROR_SHARING_VIOLATION) ||
		errors.Is(err, windows.ERROR_LOCK_VIOLATION) ||
		errors.Is(err, windows.ERROR_ACCESS_DENIED)
}

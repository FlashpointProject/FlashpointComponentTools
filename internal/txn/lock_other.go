//go:build !windows

package txn

import (
	"errors"
	"syscall"
)

// isLocked reports whether err means "in use right now, try again later".
// Unix rarely refuses renames of open files; EBUSY and ETXTBSY are the
// closest equivalents.
func isLocked(err error) bool {
	return errors.Is(err, syscall.EBUSY) || errors.Is(err, syscall.ETXTBSY)
}

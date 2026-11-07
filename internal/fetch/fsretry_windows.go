package fetch

import (
	"errors"

	"golang.org/x/sys/windows"
)

// isTransientFSError reports errors a scanner holding the file causes.
func isTransientFSError(err error) bool {
	return errors.Is(err, windows.ERROR_SHARING_VIOLATION) ||
		errors.Is(err, windows.ERROR_ACCESS_DENIED) ||
		errors.Is(err, windows.ERROR_LOCK_VIOLATION)
}

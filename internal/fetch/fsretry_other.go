//go:build !windows

package fetch

// isTransientFSError: only Windows refuses renames of files others hold open.
func isTransientFSError(error) bool { return false }

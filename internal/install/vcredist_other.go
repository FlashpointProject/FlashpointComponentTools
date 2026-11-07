//go:build !windows

package install

import (
	"context"
	"log/slog"
)

// NeedsVCRedist reports whether the Visual C++ 2015-2022 x86 runtime is
// missing (always false outside Windows).
func NeedsVCRedist() bool { return false }

// InstallVCRedist does nothing outside Windows.
func InstallVCRedist(ctx context.Context, log *slog.Logger) (rebootNeeded bool, err error) {
	return false, nil
}

//go:build !windows

package install

// CreateShortcuts does nothing outside Windows (nil).
func CreateShortcuts(root string, desktop, startMenu bool) error { return nil }

// removeShortcuts: no shortcuts outside Windows.
func removeShortcuts(root string) error { return nil }

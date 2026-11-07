package state

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/manifest"
)

const (
	ConfigName = "fpm.cfg"
	// DefaultSource is the C# manager's stable channel. An existing fpm.cfg
	// keeps whatever it says (legacy fpm shipped an unstable.life URL).
	DefaultSource = "https://nexus-dev.flashpointarchive.org/repository/stable/components.xml"
	// DefaultRootRaw is the Manager folder's parent, stored relative so the
	// install survives being moved.
	DefaultRootRaw = ".."
)

// ErrRootNotFound: neither the configured folder nor the Manager's parent
// is a Flashpoint root.
var ErrRootNotFound = errors.New("Flashpoint folder not found")

// LoadConfig reads <managerDir>/fpm.cfg. A missing file or missing/empty
// lines fall back to the defaults: an odd config must never stop the
// launcher's `fpm list`. Only real I/O errors are returned.
func LoadConfig(managerDir string) (*Config, error) {
	cfg := &Config{RootRaw: DefaultRootRaw, Source: DefaultSource}
	path := filepath.Join(managerDir, ConfigName)
	data, err := readLimited(path, 1<<20)
	if errors.Is(err, fs.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", path, err)
	}
	lines := strings.Split(string(bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))), "\n")
	if v := strings.TrimSuffix(lines[0], "\r"); strings.TrimSpace(v) != "" {
		cfg.RootRaw = v
	}
	if len(lines) > 1 && strings.TrimSpace(lines[1]) != "" {
		cfg.Source = strings.TrimSpace(lines[1])
	}
	return cfg, nil
}

// SaveConfig writes <managerDir>/fpm.cfg atomically, as the C#
// File.WriteAllLines did (two CRLF-terminated lines).
func SaveConfig(managerDir string, cfg *Config) error {
	if strings.ContainsAny(cfg.RootRaw+cfg.Source, "\r\n") {
		return errors.New("fpm.cfg values must not contain line breaks")
	}
	path := filepath.Join(managerDir, ConfigName)
	if err := WriteFileAtomic(path, []byte(cfg.RootRaw+"\r\n"+cfg.Source+"\r\n"), 0o644); err != nil {
		return fmt.Errorf("writing %s: %w", path, err)
	}
	return nil
}

// ResolveRoot turns fpm.cfg line 1 into an absolute, clean path. Relative
// values are relative to the Manager folder (the C# cwd). Backslashes are
// accepted on every OS.
func ResolveRoot(managerDir, raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		s = DefaultRootRaw
	}
	s = filepath.FromSlash(strings.ReplaceAll(s, `\`, "/"))
	switch {
	case filepath.IsAbs(s), filepath.VolumeName(s) != "":
		return filepath.Clean(s)
	case strings.HasPrefix(s, string(filepath.Separator)):
		// Windows `\Flashpoint`: rooted on the Manager folder's drive.
		return filepath.Clean(filepath.VolumeName(managerDir) + s)
	default:
		return filepath.Join(managerDir, s)
	}
}

// IsRoot reports whether dir looks like a Flashpoint root: Components/ holds
// at least one record, or (fresh install, or records lost) both Launcher/
// and Data/ exist. It cannot use the manifest: discovery runs before it is
// loaded.
func IsRoot(dir string) bool {
	entries, _ := os.ReadDir(ComponentsDir(dir))
	for _, e := range entries {
		if e.Type().IsRegular() && isRecordName(e.Name()) && manifest.ValidID(e.Name()) {
			return true
		}
	}
	isDir := func(p string) bool { fi, err := os.Stat(filepath.Join(dir, p)); return err == nil && fi.IsDir() }
	return isDir("Launcher") && isDir("Data")
}

// FindRoot locates the Flashpoint root from fpm.cfg (cfg nil = defaults).
// If the configured folder is not a root but the Manager folder's parent is,
// that is returned with repaired = true, and the caller should save `..\`
// as RootRaw: this fixes a moved install or a stale absolute path (e.g. a
// copy still pointing at C:\Flashpoint).
func FindRoot(managerDir string, cfg *Config) (root string, repaired bool, err error) {
	raw := DefaultRootRaw
	if cfg != nil {
		raw = cfg.RootRaw
	}
	if abs, err := filepath.Abs(managerDir); err == nil {
		managerDir = abs
	}
	configured := ResolveRoot(managerDir, raw)
	if IsRoot(configured) {
		return configured, false, nil
	}
	parent := filepath.Dir(managerDir)
	if IsRoot(parent) {
		return parent, true, nil
	}
	return "", false, fmt.Errorf(`%w: checked %q and %q (neither contains Components\ records, or Launcher\ and Data\)`,
		ErrRootNotFound, configured, parent)
}

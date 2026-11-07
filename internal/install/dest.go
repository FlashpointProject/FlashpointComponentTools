// Package install holds the parts of FlashpointInstaller.exe that are not
// GUI: validating the destination, removing an abandoned install, the
// Visual C++ runtime and the shortcuts; and the manager's Uninstall.
package install

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
)

// maxDestLen is the legacy installer's limit: deeper paths inside
// Flashpoint (games, Wine prefixes, PHP) break the 260-character limit.
const maxDestLen = 192

// DefaultDest is <volume of exe>\Flashpoint on Windows ($HOME/Flashpoint elsewhere).
// The volume of the exe is where the user put the installer, typically the
// drive they want Flashpoint on.
func DefaultDest(exe string) string {
	if runtime.GOOS != "windows" {
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Flashpoint")
	}
	return filepath.VolumeName(exe) + `\Flashpoint`
}

// CheckDest validates an install destination typed or picked by the user and
// returns it absolute and cleaned, plus create: the outermost folder that
// doesn't exist yet and will be created for it (abs itself or a missing
// ancestor), or "" when abs exists (and is empty). Errors are user-facing
// sentences.
func CheckDest(path string) (abs, create string, err error) {
	// Explorer's "Copy as path" puts the path in quotes.
	path = strings.Trim(strings.TrimSpace(path), `"`)
	if path == "" {
		return "", "", errors.New("Choose a folder for Flashpoint.")
	}
	// filepath.IsAbs on Windows needs a drive with a backslash (C:\) or a
	// UNC volume; C:foo and \foo depend on the current drive or folder.
	if !filepath.IsAbs(path) {
		return "", "", errors.New(`Enter a full path, such as C:\Flashpoint.`)
	}
	abs = filepath.Clean(path)
	if utf8.RuneCountInString(abs) >= maxDestLen {
		return "", "", errors.New("This path is too long; Flashpoint may not work in it. Choose a shorter one.")
	}
	// Dir of a drive root (C:\, \\server\share, /) is itself.
	if filepath.Dir(abs) == abs {
		return "", "", fmt.Errorf("Choose a folder, not a whole drive, such as %s.", filepath.Join(abs, "Flashpoint"))
	}
	if runtime.GOOS == "windows" {
		if err := checkProtected(abs); err != nil {
			return "", "", err
		}
	}
	return checkExisting(abs)
}

// checkProtected rejects the Windows places Flashpoint must not go in:
// protected folders (writes fail or get virtualized without admin rights),
// the temp folder (Disk Cleanup and Storage Sense empty it) and OneDrive
// (syncing and Files On-Demand break the launcher and the games).
func checkProtected(abs string) error {
	env := func(names ...string) []string {
		var dirs []string
		for _, name := range names {
			dirs = append(dirs, os.Getenv(name))
		}
		return dirs
	}
	places := []struct {
		dirs   []string
		reason string
	}{
		{env("ProgramFiles", "ProgramFiles(x86)"), "in Program Files, because Windows protects it"},
		{env("SystemRoot"), "in the Windows folder, because Windows protects it"},
		{[]string{os.TempDir()}, "in the temporary folder, because Windows may clean it up"},
		{env("OneDrive", "OneDriveConsumer", "OneDriveCommercial"), "in OneDrive, because syncing breaks it"},
	}
	for _, p := range places {
		for _, dir := range p.dirs {
			if dir != "" && within(abs, dir) {
				return fmt.Errorf(`Flashpoint can't be installed %s. Choose another folder, such as %s\Flashpoint.`,
					p.reason, filepath.VolumeName(abs))
			}
		}
	}
	return nil
}

// within reports whether path is dir or inside it, comparing whole path
// segments case-insensitively as Windows does (C:\Program Files2 is not
// inside C:\Program Files).
func within(path, dir string) bool {
	path = filepath.Clean(path)
	dir = strings.TrimSuffix(filepath.Clean(dir), string(filepath.Separator))
	return strings.EqualFold(path, dir) ||
		len(path) > len(dir) && path[len(dir)] == filepath.Separator && strings.EqualFold(path[:len(dir)], dir)
}

// checkExisting applies the rules for a destination that already exists:
// only an empty folder is accepted, so Cleanup can never delete user files.
// For a new one it finds what Cleanup must remove.
func checkExisting(abs string) (string, string, error) {
	fi, err := os.Stat(abs)
	if errors.Is(err, fs.ErrNotExist) {
		// Also the missing parents MkdirAll will create, up to (not
		// including) the volume root.
		create := abs
		for p := filepath.Dir(abs); p != filepath.Dir(p) && missing(p); p = filepath.Dir(p) {
			create = p
		}
		return abs, create, nil
	}
	if err != nil {
		return "", "", fmt.Errorf("Can't open %s: %w", abs, cause(err))
	}
	if !fi.IsDir() {
		return "", "", fmt.Errorf("%s is a file. Choose a folder.", abs)
	}
	if state.IsRoot(abs) {
		return "", "", errors.New("Flashpoint is already installed in this folder. Use its Flashpoint Manager to add or remove components, or choose another folder.")
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return "", "", fmt.Errorf("Can't open %s: %w", abs, cause(err))
	}
	if len(entries) > 0 {
		return "", "", errors.New("This folder isn't empty. Choose a new or empty folder.")
	}
	return abs, "", nil
}

func missing(path string) bool {
	_, err := os.Stat(path)
	return errors.Is(err, fs.ErrNotExist)
}

// cause drops the path and the system call name that *fs.PathError adds
// ("CreateFile X: The network name cannot be found."): the message names
// the path already.
func cause(err error) error {
	var pe *fs.PathError
	if errors.As(err, &pe) {
		return pe.Err
	}
	return err
}

// Cleanup removes an abandoned install: create (all of it) when it is not
// "", else everything inside dir. Only for values CheckDest returned.
func Cleanup(dir, create string) error {
	var errs []error
	if create != "" {
		errs = append(errs, os.RemoveAll(create))
	} else {
		entries, err := os.ReadDir(dir)
		errs = append(errs, err)
		for _, e := range entries {
			errs = append(errs, os.RemoveAll(filepath.Join(dir, e.Name())))
		}
	}
	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("could not remove the unfinished install in %s: %w", dir, err)
	}
	return nil
}

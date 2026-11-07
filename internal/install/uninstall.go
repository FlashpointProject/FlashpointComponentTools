package install

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
)

// flashpointNames are the top-level entries of a Flashpoint folder that no
// component record lists: made by the manager, the launcher or the installer.
var flashpointNames = []string{"Components", "Manager", "Launcher", "Data", "Logs", "Temp", "Curations",
	"preferences.json", "extConfig.json", "Flashpoint.lnk"}

// CheckUninstall says why root must not be uninstalled, nil if it can be.
func CheckUninstall(root string) error {
	root = filepath.Clean(root)
	if filepath.Dir(root) == root { // the old installer allowed D:\ itself
		return fmt.Errorf("Flashpoint is installed in %s, a whole drive, so the manager won't uninstall it. Delete its folders by hand.", root)
	}
	if !state.IsRoot(root) {
		return fmt.Errorf("%s is not a Flashpoint folder.", root)
	}
	return nil
}

// Uninstall deletes Flashpoint from root, games, save data, playlists and
// favorites included, and the Flashpoint.lnk shortcuts that start it. Only
// Flashpoint's entries go: what else is in root (the old installer allowed
// a folder with the user's own files) is kept and returned in kept. keep is
// the running manager, which can't be deleted yet. On error, the rest was
// still deleted; the error names the first thing that wasn't.
func Uninstall(root, keep string) (kept []string, err error) {
	if err := CheckUninstall(root); err != nil {
		return nil, err
	}
	ours := map[string]bool{}
	for _, n := range flashpointNames {
		ours[strings.ToLower(n)] = true
	}
	recs, _ := state.ReadRecords(root) // an unreadable record only adds to kept
	for _, r := range recs {
		for _, f := range r.Files {
			ours[strings.ToLower(strings.Split(f, "/")[0])] = true
		}
	}
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, fmt.Errorf("Could not read %s: %w", root, cause(err))
	}
	for _, e := range entries {
		if !ours[strings.ToLower(e.Name())] {
			kept = append(kept, e.Name())
			continue
		}
		err = cmp.Or(err, removeAllBut(filepath.Join(root, e.Name()), keep))
	}
	err = cmp.Or(err, removeShortcuts(root))
	var pe *fs.PathError
	if errors.As(err, &pe) { // "C:\FP\x.dll: Access is denied.", not "remove C:\FP\x.dll: ..."
		err = fmt.Errorf("Could not delete %s: %w", pe.Path, pe.Err)
	}
	return kept, err
}

// removeAllBut deletes path, or everything in it but keep when keep is in
// it. os.RemoveAll deletes all it can and returns the first error.
func removeAllBut(path, keep string) error {
	if !within(keep, path) {
		return os.RemoveAll(path)
	}
	if within(path, keep) {
		return nil
	}
	entries, err := os.ReadDir(path)
	for _, e := range entries {
		err = cmp.Or(err, removeAllBut(filepath.Join(path, e.Name()), keep))
	}
	return err
}

package engine

import (
	"errors"
	"strings"
)

// ErrPolicy is wrapped when a component would write where it must not.
var ErrPolicy = errors.New("write policy violation")

// reserved reports whether rel (FP-relative, slash-separated, already
// validated by txn) belongs to the manager itself: the records in
// Components/ (a zip entry there would corrupt the install state) and
// fpm.cfg. Which component may write where is otherwise up to the server,
// which publishes core-launcher anyway; Windows compares case-insensitively.
func reserved(rel string) bool {
	top, _, _ := strings.Cut(rel, "/")
	return strings.EqualFold(top, "Components") || strings.EqualFold(rel, "Manager/fpm.cfg")
}

// preserved user-data files are shipped as a starting point and then edited
// by the user: if one exists, installs and updates keep it and removals
// leave it. Stopgap until a v2 manifest carries per-component preserve rules.
var preserved = []string{
	"Data/Playlists/Favorites.json", // playlist-favorites
	"Manager/fpm.cfg",               // never in a zip, but a C#-era record could list it
}

func isPreserved(rel string) bool {
	for _, p := range preserved {
		if strings.EqualFold(p, rel) {
			return true
		}
	}
	return false
}

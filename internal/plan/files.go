package plan

import (
	"path"
	"strings"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
)

// RecordPath is the FP-relative path of a component's record file.
func RecordPath(id string) string { return path.Join("Components", id) }

// pathKey compares FP-relative paths the way Windows does: case-insensitively.
func pathKey(p string) string { return strings.ToLower(strings.ReplaceAll(p, `\`, "/")) }

// FilesToDelete returns the FP-relative paths (spelled as in the old record)
// that applying p must delete: the files of removed components and the old
// files of updated ones, minus every path still owned afterwards by an
// untouched component or by a staged one (legacy tools deleted shared files;
// we don't). staged maps each Install and Update ID to its new file list,
// known only after staging. Removed components' records Components/<id> come
// last. Order: removals, then updates, in plan order; no duplicates.
func FilesToDelete(recs map[string]*state.Record, p *Plan, staged map[string][]string) []string {
	if p == nil {
		return nil
	}
	changing := map[string]bool{}
	for _, c := range p.Update {
		changing[c.ID] = true
	}
	for _, rec := range p.Remove {
		changing[rec.ID] = true
	}
	done := map[string]bool{} // kept or already listed
	for id, rec := range recs {
		if rec != nil && !changing[id] {
			for _, f := range rec.Files {
				done[pathKey(f)] = true
			}
		}
	}
	for _, files := range staged {
		for _, f := range files {
			done[pathKey(f)] = true
		}
	}

	var out []string
	add := func(files ...string) {
		for _, f := range files {
			if k := pathKey(f); !done[k] {
				done[k] = true
				out = append(out, f)
			}
		}
	}
	for _, rec := range p.Remove {
		add(rec.Files...)
	}
	for _, c := range p.Update {
		if rec := recs[c.ID]; rec != nil {
			add(rec.Files...)
		}
	}
	for _, rec := range p.Remove {
		add(RecordPath(rec.ID))
	}
	return out
}

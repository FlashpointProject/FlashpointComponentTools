// Package manifest parses and validates the Flashpoint component manifest
// (components.xml, "v1") into a normalized in-memory model.
//
// Everything in a manifest comes from the network, so Parse validates every
// value that later reaches the filesystem (IDs, paths) and rejects the whole
// manifest when one is unsafe: a broken or malicious manifest must never be
// partially installable.
package manifest

import (
	"errors"
	"strings"
	"time"
)

// Manifest is a parsed, validated component list.
type Manifest struct {
	// BaseURL is the <list url="..."> attribute, always ending in "/":
	// zip URL = BaseURL + ID + ".zip".
	BaseURL string

	Roots      []*Category           // top-level categories in document order
	Components map[string]*Component // by full ID
	Categories map[string]*Category  // by full ID
	Order      []string              // component IDs in document order

	// Warnings are non-fatal problems (dangling dependencies, "-" in an id
	// attribute), for logging.
	Warnings []string
}

// Category is a node in the category tree.
type Category struct {
	ID          string // full ID, e.g. "core-server"
	LocalID     string // the id attribute, e.g. "server"
	Title       string
	Description string
	Required    bool   // same "core" rule as components
	Items       []Item // children in document order
}

// Item is exactly one of Category or Component.
type Item struct {
	Category  *Category
	Component *Component
}

// Component is one installable unit.
type Component struct {
	ID          string // full ID: ancestor category ids + local id joined with "-"
	LocalID     string
	Title       string
	Description string

	// Path is the extraction folder relative to the Flashpoint root, in
	// CleanRelPath form. "" is the root itself.
	Path string

	Depends []string // full IDs as written; may include dangling refs (see Warnings)

	// DownloadSize is the zip size, 0 when absent (the C# manager never reads
	// it); the engine refuses to install an archive without it.
	DownloadSize int64
	InstallSize  int64 // bytes on disk; 0 means metadata-only (see HasArchive)

	// Hash is the "hash" attribute verbatim (CRC32 of the zip as 8 hex chars).
	// It is also the version written into Components/<id> records, so keep
	// the exact string.
	Hash     string
	CRC32    uint32
	SHA256   []byte    // optional "sha256" attribute; nil if absent
	Modified time.Time // date-modified (unix seconds); zero if absent

	Required     bool // "core*": cannot be removed, always installed
	DefaultOn    bool // pre-selected in the installer ("extra*" are off)
	NeverOutdate bool // "core-database": never reported as outdated
}

// ZipName returns the file name of the component archive on the server.
func (c *Component) ZipName() string { return c.ID + ".zip" }

// HasArchive reports whether the component has a zip to download. All legacy
// clients decide this on install-size, not download-size: supportpack-html5
// (download-size 160, install-size 0) is never fetched.
func (c *Component) HasArchive() bool { return c.InstallSize > 0 }

// ErrInvalid wraps every fatal problem found while parsing a manifest.
var ErrInvalid = errors.New("invalid component list")

// DerivePolicy sets the policy flags from the legacy naming conventions
// (C# manager/installer Common.cs; the launcher has the core-database rule):
// the first ID segment "core" is required, a raw "extra" prefix is not
// pre-selected, and exactly "core-database" is never outdated (it holds the
// user's data).
func DerivePolicy(c *Component) {
	c.Required = requiredID(c.ID)
	c.DefaultOn = c.Required || !strings.HasPrefix(c.ID, "extra")
	c.NeverOutdate = c.ID == "core-database"
}

func requiredID(id string) bool {
	first, _, _ := strings.Cut(id, "-")
	return first == "core"
}

// Resolve maps a user argument to components in document order, using
// legacy fpm's rule `ID == id || ID.StartsWith(id + "-")`: a component ID
// yields that component, a category ID everything below it. Matching is
// case-sensitive, as in fpm. ok is false when id names neither.
func (m *Manifest) Resolve(id string) (comps []*Component, ok bool) {
	_, isComp := m.Components[id]
	_, isCat := m.Categories[id]
	if !isComp && !isCat {
		return nil, false
	}
	for _, cid := range m.Order {
		if cid == id || strings.HasPrefix(cid, id+"-") {
			comps = append(comps, m.Components[cid])
		}
	}
	return comps, true
}

// All returns every component in document order.
func (m *Manifest) All() []*Component {
	out := make([]*Component, len(m.Order))
	for i, id := range m.Order {
		out[i] = m.Components[id]
	}
	return out
}

// Package plan turns a user request (install / update / remove components)
// plus the current install state into a validated, ordered, sized Plan. It
// does no I/O.
//
// Legacy references: fpm/src/Handlers.cs (C# fpm CLI) and
// FlashpointManager/src/Common.cs (C# manager GUI).
package plan

import (
	"fmt"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/manifest"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
)

// Status of a component relative to the manifest.
type Status int

const (
	NotInstalled Status = iota
	UpToDate
	Outdated   // installed, record hash != manifest hash (never for NeverOutdate)
	Missing    // required but not installed (legacy: counts as "needs update")
	Deprecated // installed record whose ID is no longer in the manifest
)

func (s Status) String() string {
	if names := []string{"not installed", "up to date", "outdated", "missing", "deprecated"}; s >= 0 && int(s) < len(names) {
		return names[s]
	}
	return fmt.Sprintf("Status(%d)", int(s))
}

// Installed reports whether a record exists for the component.
func (s Status) Installed() bool { return s == UpToDate || s == Outdated || s == Deprecated }

// statusOf follows the legacy manager (Common.cs:333-395): the hash compare
// is ordinal, core-database is never outdated (legacy fpm forgot that), a
// missing required component is Missing, and a record whose ID left the
// manifest is Deprecated. A nil record counts as absent.
func statusOf(m *manifest.Manifest, recs map[string]*state.Record, id string) Status {
	rec, c := recs[id], m.Components[id]
	switch {
	case c == nil && rec != nil:
		return Deprecated
	case c == nil:
		return NotInstalled
	case rec == nil && c.Required:
		return Missing
	case rec == nil:
		return NotInstalled
	case rec.Hash != c.Hash && !c.NeverOutdate:
		return Outdated
	}
	return UpToDate
}

// Statuses returns the status of every manifest component plus a Deprecated
// entry for every record whose ID is not in the manifest.
func Statuses(m *manifest.Manifest, recs map[string]*state.Record) map[string]Status {
	out := make(map[string]Status)
	for id := range m.Components {
		out[id] = statusOf(m, recs, id)
	}
	for id, rec := range recs {
		if rec != nil && m.Components[id] == nil {
			out[id] = Deprecated
		}
	}
	return out
}

// Request describes what the user asked for. IDs may be component or
// category IDs (a category selects all components below it, like fpm).
type Request struct {
	Install          []string
	Remove           []string // may also name deprecated records
	Update           []string
	UpdateAll        bool // every Outdated and Missing component
	RemoveDeprecated bool // with UpdateAll: also remove deprecated records

	// Reinstall re-fetches components even when they are up to date (repair,
	// DESIGN.md §4.7). A not-installed one is installed.
	Reinstall []string

	// AutoDeps adds not-installed dependencies (transitively) instead of
	// failing, and also updates outdated dependencies of components being
	// updated (legacy `fpm update <id>`, Handlers.cs:336-346).
	AutoDeps bool
	// CascadeRemove also removes installed components that depend on a
	// removed one, instead of failing.
	CascadeRemove bool
}

// Plan is the resolved set of operations. Install and Update are in
// dependency order (dependencies first, otherwise manifest order); Remove is
// the reverse (dependents first).
type Plan struct {
	Install []*manifest.Component // not installed -> installed
	Update  []*manifest.Component // installed, replaced by the manifest version
	Remove  []*state.Record       // installed (or deprecated) -> removed

	AddedDeps       []string // added only because of AutoDeps
	AddedDependents []string // removed only because of CascadeRemove

	// Skipped lists requested components that need no work, for the CLI's
	// legacy "... will be skipped" messages.
	Skipped []Skip

	// Warnings are non-fatal problems, e.g. dangling dependencies (ignored,
	// like every legacy client does).
	Warnings []string

	DownloadBytes int64 // sum of DownloadSize over Install+Update with an archive
	DiskDelta     int64 // new InstallSize - old record InstallSize - removed
}

// Empty reports whether the plan does nothing.
func (p *Plan) Empty() bool { return len(p.Install)+len(p.Update)+len(p.Remove) == 0 }

// SkipReason says why a requested component was not acted on.
type SkipReason int

const (
	SkipAlreadyInstalled SkipReason = iota // install requested, already up to date
	SkipNotInstalled                       // update or remove requested, not installed
	SkipUpToDate                           // update requested, already up to date
)

// Skip is one requested component that needs no work.
type Skip struct {
	ID     string
	Reason SkipReason
}

// Error is what Resolve returns for a request that cannot be carried out
// (unknown IDs, removing required components, dependency problems,
// conflicting operations). The message is user-facing.
type Error struct{ Msg string }

func (e *Error) Error() string { return e.Msg }

func refuse(format string, args ...any) error { return &Error{Msg: fmt.Sprintf(format, args...)} }

package plan

import (
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/manifest"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
)

type resolver struct {
	m     *manifest.Manifest
	recs  map[string]*state.Record
	req   Request
	fetch map[string]bool // install (no record) or update (record exists)
	rm    map[string]bool
	plan  *Plan
}

// Resolve turns req into a validated, ordered, sized Plan.
//
// Where this differs from legacy fpm: unknown IDs are an error (the CLI
// filters them first with fpm's message); installing an outdated component
// updates it; removing a required component is refused (the manager refuses
// too); removing a component that installed ones depend on is refused unless
// CascadeRemove. Dependencies are the manifest deps of the new version and
// are satisfied by any installed version. Dangling deps are ignored with a
// warning.
func Resolve(m *manifest.Manifest, recs map[string]*state.Record, req Request) (*Plan, error) {
	if m == nil {
		return nil, errors.New("plan: no component list loaded")
	}
	r := &resolver{m: m, recs: recs, req: req, fetch: map[string]bool{}, rm: map[string]bool{}, plan: &Plan{}}
	if err := r.selectRequested(); err != nil {
		return nil, err
	}
	if err := r.closeDeps(); err != nil {
		return nil, err
	}
	r.finish()
	return r.plan, nil
}

func (r *resolver) status(id string) Status { return statusOf(r.m, r.recs, id) }

// selectRequested expands the requested IDs and classifies them by status.
func (r *resolver) selectRequested() error {
	inst, unk1 := r.expand(r.req.Install, false)
	upd, unk2 := r.expand(r.req.Update, false)
	rem, unk3 := r.expand(r.req.Remove, true)
	rein, unk4 := r.expand(r.req.Reinstall, false)
	if unknown := slices.Concat(unk1, unk2, unk3, unk4); len(unknown) > 0 {
		return refuse("component or category %s does not exist (check the ID with \"fpm list\")", quote(unknown))
	}

	explicit := map[string]bool{}
	for _, id := range inst {
		if r.status(id) == UpToDate {
			r.skip(id, SkipAlreadyInstalled)
		} else {
			r.fetch[id], explicit[id] = true, true
		}
	}
	for _, id := range upd {
		switch r.status(id) {
		case Outdated:
			r.fetch[id], explicit[id] = true, true
		case UpToDate:
			r.skip(id, SkipUpToDate)
		default:
			r.skip(id, SkipNotInstalled) // also Missing: use Install or UpdateAll
		}
	}
	for _, id := range rein {
		r.fetch[id], explicit[id] = true, true
	}
	if r.req.UpdateAll {
		for _, id := range r.m.Order {
			if s := r.status(id); s == Outdated || s == Missing {
				r.fetch[id] = true
			}
		}
	}

	var required []string
	for _, id := range rem {
		switch c := r.m.Components[id]; {
		case !r.status(id).Installed():
			r.skip(id, SkipNotInstalled)
		case c != nil && c.Required:
			required = append(required, id)
		case explicit[id]:
			return refuse("component %q is both installed and removed by this request", id)
		default:
			delete(r.fetch, id) // an explicit removal beats UpdateAll
			r.rm[id] = true
		}
	}
	if len(required) > 0 {
		return refuse("cannot remove required component %s", quote(required))
	}
	if r.req.RemoveDeprecated {
		for id, rec := range r.recs {
			if rec != nil && r.m.Components[id] == nil {
				r.rm[id] = true
			}
		}
	}
	return nil
}

// closeDeps adds dependencies of everything fetched and checks (or, with
// CascadeRemove, removes) installed dependents of everything removed.
func (r *resolver) closeDeps() error {
	var problems []string
	work := r.sorted(r.fetch)
	for i := 0; i < len(work); i++ { // each ID enters fetch once, so cycles terminate
		id := work[i]
		for _, dep := range r.m.Components[id].Depends {
			if r.m.Components[dep] == nil {
				w := fmt.Sprintf("component %s depends on %s, which is not in the component list; ignoring it", id, dep)
				if !slices.Contains(r.plan.Warnings, w) {
					r.plan.Warnings = append(r.plan.Warnings, w)
				}
				continue
			}
			if r.fetch[dep] || r.rm[dep] { // a removed dep is reported below
				continue
			}
			// Any installed version satisfies a dependency, but with AutoDeps
			// an update also updates its outdated deps.
			s := r.status(dep)
			if s.Installed() && !(s == Outdated && r.req.AutoDeps && r.recs[id] != nil) {
				continue
			}
			if !r.req.AutoDeps {
				problems = append(problems, fmt.Sprintf("%s needs %s, which is not installed (install it too)", id, dep))
				continue
			}
			r.fetch[dep] = true
			r.plan.AddedDeps = append(r.plan.AddedDeps, dep)
			work = append(work, dep)
		}
	}

	// Everything that exists after the plan may depend on a removed component.
	cands := maps.Clone(r.fetch)
	for id, rec := range r.recs {
		if rec != nil {
			cands[id] = true
		}
	}
	ids := r.sorted(cands)
	for changed := true; changed; { // cascades can enable further cascades
		changed = false
		for _, id := range ids {
			for _, dep := range r.depsAfter(id) {
				if !r.rm[dep] || r.rm[id] {
					continue
				}
				c := r.m.Components[id]
				switch {
				case r.fetch[id]:
					return refuse("cannot remove %s: %s is being installed or updated and needs it", dep, id)
				case !r.req.CascadeRemove:
					problems = append(problems, fmt.Sprintf("%s is needed by %s (remove that too)", dep, id))
				case c != nil && c.Required:
					return refuse("cannot remove required component %q (it depends on %s)", id, dep)
				default:
					r.rm[id] = true
					r.plan.AddedDependents = append(r.plan.AddedDependents, id)
					changed = true
				}
			}
		}
	}
	if len(problems) > 0 {
		return refuse("dependency problem: %s", strings.Join(problems, "; "))
	}
	return nil
}

// depsAfter returns what id depends on once the plan is applied: its
// manifest deps without dangling ones, even if its record lists others. A
// stale record must not block removals: supportpack-babyz lists the
// deprecated supportpack-netscape, which must stay removable. Only
// deprecated records (always present here) use their record deps.
func (r *resolver) depsAfter(id string) []string {
	if c := r.m.Components[id]; c != nil {
		return slices.DeleteFunc(slices.Clone(c.Depends), func(d string) bool { return r.m.Components[d] == nil })
	}
	return r.recs[id].Depends
}

// finish orders the sets and computes sizes.
func (r *resolver) finish() {
	p := r.plan
	for _, id := range r.order(r.fetch, func(id string) []string { return r.m.Components[id].Depends }) {
		c := r.m.Components[id]
		if c.HasArchive() {
			p.DownloadBytes += c.DownloadSize
		}
		p.DiskDelta += c.InstallSize
		if rec := r.recs[id]; rec != nil {
			p.Update = append(p.Update, c)
			p.DiskDelta -= rec.InstallSize
		} else {
			p.Install = append(p.Install, c)
		}
	}
	rm := r.order(r.rm, r.depsAfter)
	for i := len(rm) - 1; i >= 0; i-- {
		rec := r.recs[rm[i]]
		p.Remove = append(p.Remove, rec)
		p.DiskDelta -= rec.InstallSize
	}
}

// order lists the IDs in set so that each comes after its dependencies in
// the set (DFS post-order, starting in manifest order). Cycles, which Parse
// rejects anyway, are broken wherever the DFS meets them.
func (r *resolver) order(set map[string]bool, deps func(string) []string) []string {
	var out []string
	seen := map[string]bool{}
	var visit func(id string)
	visit = func(id string) {
		if !set[id] || seen[id] {
			return
		}
		seen[id] = true
		for _, d := range deps(id) {
			visit(d)
		}
		out = append(out, id)
	}
	for _, id := range r.sorted(set) {
		visit(id)
	}
	return out
}

// sorted returns the keys of set in manifest order; IDs not in the manifest
// (deprecated records) go last, by name.
func (r *resolver) sorted(set map[string]bool) []string {
	var out, extra []string
	for _, id := range r.m.Order {
		if set[id] {
			out = append(out, id)
		}
	}
	for id := range set {
		if r.m.Components[id] == nil {
			extra = append(extra, id)
		}
	}
	slices.Sort(extra)
	return append(out, extra...)
}

// expand turns requested IDs into component IDs (with records: also
// deprecated record IDs), deduplicated, in request order.
func (r *resolver) expand(ids []string, records bool) (out, unknown []string) {
	for _, id := range ids {
		comps, ok := r.m.Resolve(id)
		if !ok && records && r.recs[id] != nil {
			comps, ok = []*manifest.Component{{ID: id}}, true
		}
		if !ok {
			unknown = append(unknown, id)
		}
		for _, c := range comps {
			if !slices.Contains(out, c.ID) {
				out = append(out, c.ID)
			}
		}
	}
	return out, unknown
}

func (r *resolver) skip(id string, why SkipReason) {
	r.plan.Skipped = append(r.plan.Skipped, Skip{ID: id, Reason: why})
}

func quote(ids []string) string {
	q := make([]string, len(ids))
	for i, id := range ids {
		q[i] = fmt.Sprintf("%q", id)
	}
	return strings.Join(q, ", ")
}

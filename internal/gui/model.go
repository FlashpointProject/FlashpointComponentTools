package gui

import (
	"fmt"
	"slices"
	"strings"

	"gioui.org/widget"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/engine"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/manifest"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/plan"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
)

// model is everything derived from one Snapshot plus the user's selections.
// It knows nothing about layout; the widget.Clickables in it are just the
// per-row click state Gio needs.
type model struct {
	snap     *engine.Snapshot
	roots    []*node
	comps    map[string]*node    // leaves by ID (manifest components and obsolete records)
	neededBy map[string][]string // dependency ID -> manifest components depending on it
	updates  []*updRow

	selected *node
	plans    [2]cachedPlan // by View: Components, Updates
}

// cachedPlan is the last plan of a view: frames come with every mouse
// move, and the plan only changes with the checkboxes.
type cachedPlan struct {
	req plan.Request
	p   *plan.Plan
	err error
}

// node is a category or a component in the tree. Obsolete records (no longer
// in the manifest) are leaves with Comp == nil under a pseudo-category.
type node struct {
	ID, Title, Desc string
	Comp            *manifest.Component
	Rec             *state.Record // installed record, nil if not installed
	Status          plan.Status
	IsCat           bool
	Parent          *node
	Kids            []*node
	Want            bool // leaves: should be installed after Apply
	Open            bool // categories: expanded
	Depth           int

	check, sel, expand widget.Clickable
}

type updRow struct {
	n          *node
	checked    bool
	check, row widget.Clickable
}

const obsoleteID = "\x00obsolete"

func newModel(snap *engine.Snapshot) *model {
	m := &model{snap: snap, comps: map[string]*node{}, neededBy: map[string][]string{}}
	st := plan.Statuses(snap.Manifest, snap.Records)
	var walk func(c *manifest.Category, parent *node, depth int) *node
	walk = func(c *manifest.Category, parent *node, depth int) *node {
		n := &node{ID: c.ID, Title: c.Title, Desc: c.Description, IsCat: true, Parent: parent, Depth: depth}
		for _, it := range c.Items {
			if it.Category != nil {
				n.Kids = append(n.Kids, walk(it.Category, n, depth+1))
				continue
			}
			comp := it.Component
			k := &node{ID: comp.ID, Title: comp.Title, Desc: comp.Description, Comp: comp,
				Rec: snap.Records[comp.ID], Status: st[comp.ID], Parent: n, Depth: depth + 1}
			k.Want = k.Rec != nil || comp.Required
			m.comps[comp.ID] = k
			n.Kids = append(n.Kids, k)
			for _, d := range comp.Depends {
				m.neededBy[d] = append(m.neededBy[d], comp.ID)
			}
		}
		return n
	}
	for _, c := range snap.Manifest.Roots {
		m.roots = append(m.roots, walk(c, nil, 0))
	}

	obs := &node{ID: obsoleteID, Title: "Obsolete components", IsCat: true, Open: true,
		Desc: "These components are installed but no longer in the component list. Uncheck them to remove them."}
	var ids []string
	for id, s := range st {
		if s == plan.Deprecated {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	for _, id := range ids {
		k := &node{ID: id, Title: id, Rec: snap.Records[id], Status: plan.Deprecated, Parent: obs, Depth: 1, Want: true,
			Desc: "This component is no longer offered and can be removed."}
		obs.Kids = append(obs.Kids, k)
		m.comps[id] = k
	}
	if len(obs.Kids) > 0 {
		m.roots = append(m.roots, obs)
	}

	// Updates: outdated and missing (checked, like the old manager), then
	// obsolete (unchecked).
	for _, id := range snap.Manifest.Order {
		if s := st[id]; s == plan.Outdated || s == plan.Missing {
			m.updates = append(m.updates, &updRow{n: m.comps[id], checked: true})
		}
	}
	for _, k := range obs.Kids {
		m.updates = append(m.updates, &updRow{n: k})
	}
	return m
}

// tri is a checkbox state.
type tri int

const (
	unchecked tri = iota
	partial
	checked
)

func (n *node) locked() bool { return n.Comp != nil && n.Comp.Required }

// triOf is the checkbox state for n of total checked.
func triOf(n, total int) tri {
	switch {
	case n == 0:
		return unchecked
	case n == total:
		return checked
	}
	return partial
}

// on is a leaf's checkbox: whether it is installed after Apply. So a
// dependency the plan adds is checked, and a component it also removes
// (because it needs a removed one) is unchecked.
func (m *model) on(n *node) bool {
	if p := m.tree(); p != nil {
		if slices.Contains(p.AddedDeps, n.ID) {
			return true
		}
		if slices.Contains(p.AddedDependents, n.ID) {
			return false
		}
	}
	return n.Want
}

// matches reports whether n matches the search query q (lower case,
// trimmed; "" matches everything).
func matches(n *node, q string) bool {
	return q == "" || strings.Contains(strings.ToLower(n.Title), q) || strings.Contains(n.ID, q)
}

// shown is the leaves below n that the search q shows: all of them below a
// matching category, else only matching ones.
func shown(n *node, q string) []*node {
	hit := false
	for a := n.Parent; a != nil; a = a.Parent {
		hit = hit || matches(a, q)
	}
	var walk func(n *node, hit bool) []*node
	walk = func(n *node, hit bool) (out []*node) {
		hit = hit || matches(n, q)
		if !n.IsCat {
			if hit {
				out = append(out, n)
			}
			return out
		}
		for _, k := range n.Kids {
			out = append(out, walk(k, hit)...)
		}
		return out
	}
	return walk(n, hit)
}

// state is n's checkbox: its shown leaves combined.
func (m *model) state(n *node, q string) tri {
	ls, on := shown(n, q), 0
	for _, l := range ls {
		if m.on(l) {
			on++
		}
	}
	return triOf(on, len(ls))
}

// canToggle reports whether a click on n's checkbox can change something.
func canToggle(n *node, q string) bool {
	return slices.ContainsFunc(shown(n, q), func(l *node) bool { return !l.locked() })
}

// toggle handles a click on n's checkbox. A leaf flips, except that an
// added dependency stays (the details say which component needs it). A
// category checks its shown leaves, or unchecks them when all are checked.
func (m *model) toggle(n *node, q string) {
	if !n.IsCat {
		if !m.on(n) || n.Want {
			m.set(n, !m.on(n))
		}
		return
	}
	v := m.state(n, q) != checked
	for _, l := range shown(n, q) {
		m.set(l, v)
	}
}

// set sets a leaf's Want. Checking also checks the installed components it
// needs that were unchecked: that undoes a cascade removal (and installing
// something whose dependency is being removed is refused anyway). Required
// leaves never change.
func (m *model) set(n *node, v bool) {
	if n.locked() {
		return
	}
	n.Want = v
	if !v {
		return
	}
	deps := []string(nil)
	if n.Comp != nil {
		deps = n.Comp.Depends
	} else if n.Rec != nil {
		deps = n.Rec.Depends
	}
	for _, d := range deps {
		if k := m.comps[d]; k != nil && k.Rec != nil && !k.Want {
			m.set(k, true)
		}
	}
}

func (n *node) leaves(out []*node) []*node {
	if !n.IsCat {
		return append(out, n)
	}
	for _, k := range n.Kids {
		out = k.leaves(out)
	}
	return out
}

// rows flattens the tree for display. With a search query q (lower case,
// trimmed) every category is expanded and only matches (plus everything
// below a matching category, and the ancestors of matches) are shown.
func (m *model) rows(q string) []*node {
	var walk func(ns []*node, inherited bool) []*node
	walk = func(ns []*node, inherited bool) (out []*node) {
		for _, n := range ns {
			hit := inherited || matches(n, q)
			var kids []*node
			if n.IsCat && (n.Open || q != "") {
				kids = walk(n.Kids, hit)
			}
			if hit || len(kids) > 0 {
				out = append(append(out, n), kids...)
			}
		}
		return out
	}
	return walk(m.roots, false)
}

// treeRequest turns the checkboxes into a request: checked and not
// installed = install, unchecked and installed = remove.
func (m *model) treeRequest() plan.Request {
	req := plan.Request{AutoDeps: true, CascadeRemove: true}
	for _, r := range m.roots {
		for _, n := range r.leaves(nil) {
			switch {
			case n.Want && n.Rec == nil:
				req.Install = append(req.Install, n.ID)
			case !n.Want && n.Rec != nil:
				req.Remove = append(req.Remove, n.ID)
			}
		}
	}
	return req
}

func (m *model) updRequest() plan.Request {
	req := plan.Request{AutoDeps: true}
	for _, u := range m.updates {
		if !u.checked {
			continue
		}
		switch u.n.Status {
		case plan.Outdated:
			req.Update = append(req.Update, u.n.ID)
		case plan.Missing:
			req.Install = append(req.Install, u.n.ID)
		case plan.Deprecated:
			req.Remove = append(req.Remove, u.n.ID)
		}
	}
	return req
}

func (m *model) resolve(req plan.Request) (*plan.Plan, error) {
	return plan.Resolve(m.snap.Manifest, m.snap.Records, req)
}

// plan is what Apply would do on view v (Components or Updates).
func (m *model) plan(v View) (*plan.Plan, error) {
	req := m.treeRequest()
	if v == ViewUpdates {
		req = m.updRequest()
	}
	c := &m.plans[v]
	if c.p == nil && c.err == nil || !slices.Equal(c.req.Install, req.Install) ||
		!slices.Equal(c.req.Update, req.Update) || !slices.Equal(c.req.Remove, req.Remove) {
		c.p, c.err = m.resolve(req)
		c.req = req
	}
	return c.p, c.err
}

// tree is the Components view's last plan (nil on error), for the checkboxes.
func (m *model) tree() *plan.Plan { return m.plans[ViewComponents].p }

// updateCount is the number for the Updates tab: what Update selected
// would do, but an obsolete row only counts when it is checked.
func (m *model) updateCount() int {
	n := 0
	for _, u := range m.updates {
		if u.checked || u.n.Status != plan.Deprecated {
			n++
		}
	}
	return n
}

func (m *model) updatesChecked() tri {
	n := 0
	for _, u := range m.updates {
		if u.checked {
			n++
		}
	}
	return triOf(n, len(m.updates))
}

func (m *model) title(id string) string {
	if n := m.comps[id]; n != nil {
		return n.Title
	}
	return id
}

func (m *model) titles(ids []string) []string {
	out := make([]string, len(ids))
	for i, id := range ids {
		out[i] = m.title(id)
	}
	return out
}

// summary describes a plan for the footer: what changes, and the sizes.
// obsolete: the removals are of obsolete components (the Updates view).
func summary(p *plan.Plan, obsolete bool) (string, string) {
	if p == nil || p.Empty() {
		return "No changes selected", ""
	}
	var parts []string
	if n := len(p.Install); n > 0 {
		s := fmt.Sprintf("install %d", n)
		if d := len(p.AddedDeps); d > 0 {
			s += " (incl. " + plural(d, "dependency") + ")"
		}
		parts = append(parts, s)
	}
	if n := len(p.Update); n > 0 {
		parts = append(parts, fmt.Sprintf("update %d", n))
	}
	if n := len(p.Remove); n > 0 {
		s := fmt.Sprintf("remove %d", n)
		if obsolete {
			s += " obsolete"
		} else if d := len(p.AddedDependents); d > 0 {
			s += " (incl. " + plural(d, "dependent") + ")"
		}
		parts = append(parts, s)
	}
	txt := strings.Join(parts, ", ")
	return strings.ToUpper(txt[:1]) + txt[1:], sizes(p)
}

// sizes is a plan's download and disk change, leaving out what is zero.
func sizes(p *plan.Plan) string {
	var parts []string
	if p.DownloadBytes > 0 {
		parts = append(parts, "Download "+size(p.DownloadBytes))
	}
	if p.DiskDelta != 0 {
		parts = append(parts, "Disk space "+signedSize(p.DiskDelta))
	}
	return strings.Join(parts, "  ·  ")
}

// whoNeeds names a component of p that needs id, preferring one the user
// chose ("" if none).
func (m *model) whoNeeds(p *plan.Plan, id string) string {
	for _, chosen := range []bool{true, false} {
		for _, c := range slices.Concat(p.Install, p.Update) {
			if slices.Contains(c.Depends, id) && !(chosen && slices.Contains(p.AddedDeps, c.ID)) {
				return m.title(c.ID)
			}
		}
	}
	return ""
}

// removedDep names a component that p removes and id needs ("" if none).
func (m *model) removedDep(p *plan.Plan, id string) string {
	if c := m.snap.Manifest.Components[id]; c != nil {
		for _, r := range p.Remove {
			if slices.Contains(c.Depends, r.ID) {
				return m.title(r.ID)
			}
		}
	}
	return ""
}

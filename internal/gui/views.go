package gui

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"slices"
	"strings"
	"time"

	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/plan"
)

func (u *ui) layout(gtx C) D {
	paint.Fill(gtx.Ops, colWhite)
	page := func(gtx C) D {
		return vcol(gtx, layout.Rigid(u.header), layout.Rigid(hline), layout.Rigid(u.banner), layout.Flexed(1, u.body))
	}
	if u.modal == nil {
		return page(gtx)
	}
	page(gtx.Disabled())
	return u.modalView(gtx)
}

func (u *ui) body(gtx C) D {
	switch u.screen {
	case scrFatal:
		return u.fatalView(gtx)
	case scrLocate:
		return u.locateView(gtx)
	case scrMain:
		return u.mainView(gtx)
	case scrRun:
		return u.runView(gtx)
	case scrUninstalled:
		return u.uninstalledView(gtx)
	case scrDest:
		return u.destView(gtx)
	case scrFinish:
		return u.finishView(gtx)
	}
	return centered(gtx, 400, func(gtx C) D {
		return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
			layout.Rigid(func(gtx C) D {
				gtx.Constraints = layout.Exact(image.Pt(gtx.Dp(36), gtx.Dp(36)))
				l := material.Loader(u.th)
				l.Color = colAccent
				return l.Layout(gtx)
			}),
			gap(16),
			layout.Rigid(lbl(u.th, 15, u.msg, colMuted).Layout))
	})
}

func (u *ui) header(gtx C) D {
	name := "Manager"
	if u.inst != nil {
		name = "Installer"
	}
	return layout.Inset{Left: 16, Right: 16, Top: 8, Bottom: 8}.Layout(gtx, func(gtx C) D {
		return hrow(gtx,
			layout.Rigid(func(gtx C) D {
				gtx.Constraints = layout.Exact(image.Pt(gtx.Dp(104), gtx.Dp(34)))
				return widget.Image{Src: u.logo, Fit: widget.Contain}.Layout(gtx)
			}),
			// The logo says "Flashpoint", so this reads "Flashpoint Manager".
			gap(12), layout.Rigid(lbl(u.th, 17, name, colMuted).Layout),
			gap(8), layout.Rigid(lbl(u.th, 13, u.opts.Version, colMuted).Layout))
	})
}

func (u *ui) banner(gtx C) D {
	if u.dismiss.Clicked(gtx) {
		u.warnings = nil
	}
	if len(u.warnings) == 0 {
		return D{}
	}
	ws := u.warnings[:min(len(u.warnings), 3)]
	txt := strings.Join(ws, "\n")
	if n := len(u.warnings) - len(ws); n > 0 {
		txt += fmt.Sprintf("\n(and %d more; see the log)", n)
	}
	return bg(gtx, colBanner, func(gtx C) D {
		return layout.Inset{Left: 16, Right: 12, Top: 8, Bottom: 8}.Layout(gtx,
			row(layout.Flexed(1, para(u.th, txt, colInk).Layout), gap(12), btn(u.th, &u.dismiss, "Dismiss", false, true)))
	})
}

// centered lays w out in the middle, at most maxW wide.
func centered(gtx C, maxW unit.Dp, w layout.Widget) D {
	return layout.Center.Layout(gtx, maxWidth(maxW, w))
}

// maxWidth lays w out exactly as wide as the space, at most maxW, with a
// margin.
func maxWidth(maxW unit.Dp, w layout.Widget) layout.Widget {
	return func(gtx C) D {
		gtx.Constraints.Max.X = min(gtx.Constraints.Max.X-gtx.Dp(48), gtx.Dp(maxW))
		gtx.Constraints.Min.X = gtx.Constraints.Max.X
		return w(gtx)
	}
}

// message is a page: a title, text, optional rows, then buttons. It sits at
// a fixed distance below the header, so a line that appears (an error)
// moves only what is below it, not the button under the mouse.
func (u *ui) message(gtx C, title, text string, rows []layout.FlexChild, buttons ...layout.FlexChild) D {
	return layout.Inset{Top: 96}.Layout(gtx, func(gtx C) D {
		return layout.N.Layout(gtx, maxWidth(580, func(gtx C) D {
			kids := []layout.FlexChild{layout.Rigid(bold(lbl(u.th, 18, title, colInk)).Layout), gap(10),
				layout.Rigid(para(u.th, text, colInk).Layout), gap(16)}
			kids = append(kids, rows...)
			kids = append(kids, gap(16), layout.Rigid(row(buttons...)))
			return vcol(gtx, kids...)
		}))
	})
}

func (u *ui) fatalView(gtx C) D {
	if u.retry.Clicked(gtx) {
		u.start("")
	}
	if u.quitBtn.Clicked(gtx) {
		u.quit = true
	}
	return u.message(gtx, u.fatalTitle, u.fatalMsg, nil,
		btn(u.th, &u.retry, "Try again", true, !u.busy), gap(8), btn(u.th, &u.quitBtn, "Quit", false, true))
}

type locateState struct {
	path       widget.Editor
	browse, ok widget.Clickable
	err        string
}

func (u *ui) locateView(gtx C) D {
	l := &u.locate
	path := cleanPath(l.path.Text())
	submit := l.ok.Clicked(gtx)
	for {
		ev, ok := l.path.Update(gtx)
		if !ok {
			break
		}
		switch ev.(type) {
		case widget.SubmitEvent:
			submit = true
		case widget.ChangeEvent:
			l.err = ""
		}
	}
	if submit && path != "" && !u.busy {
		l.err = ""
		u.start(path)
		return D{}
	}
	if l.browse.Clicked(gtx) {
		u.browseInto(&l.path, &l.err, nil)
	}
	if u.quitBtn.Clicked(gtx) {
		u.quit = true
	}
	return u.message(gtx, "Locate your Flashpoint folder",
		"The manager could not find your Flashpoint installation. Choose the folder that contains the Launcher, Data and Components folders.",
		[]layout.FlexChild{
			layout.Rigid(u.folderRow(&l.path, &l.browse, `For example C:\Flashpoint`)),
			layout.Rigid(func(gtx C) D { return layout.Inset{Top: 8}.Layout(gtx, para(u.th, l.err, colAccent).Layout) }),
		},
		btn(u.th, &l.ok, "Continue", true, !u.busy && path != ""), gap(8), btn(u.th, &u.quitBtn, "Quit", false, true))
}

func (u *ui) folderRow(ed *widget.Editor, browse *widget.Clickable, hint string) layout.Widget {
	return row(layout.Flexed(1, textField(u.th, ed, hint)), gap(8), btn(u.th, browse, "Browse…", false, u.expl != nil))
}

// cleanPath trims spaces and the quotes of Explorer's "Copy as path".
func cleanPath(s string) string { return strings.Trim(strings.TrimSpace(s), `"`) }

// browseInto shows the folder picker (it blocks, so not on the UI goroutine).
// adjust, if not nil, changes the chosen folder (on that goroutine too).
func (u *ui) browseInto(ed *widget.Editor, errp *string, adjust func(string) string) {
	go func() {
		p, err := u.expl.ChooseFolder()
		if err == nil && adjust != nil {
			p = adjust(p)
		}
		u.post(func() {
			if err == nil {
				ed.SetText(p)
			} else if !errors.Is(err, explorer.ErrUserDecline) {
				*errp = "The folder picker is not available (" + err.Error() + "); type the path instead."
			}
		})
	}()
}

func (u *ui) mainView(gtx C) D {
	if u.inst != nil {
		return u.installComponentsView(gtx)
	}
	for i := range u.rail {
		if u.rail[i].Clicked(gtx) {
			u.view = View(i)
		}
	}
	// Handle clicks before anything is drawn, so that the plan matches this
	// frame's selection.
	var p *plan.Plan
	var err error
	view := u.settingsView
	switch u.view {
	case ViewComponents:
		rows := u.componentEvents(gtx)
		p, err = u.m.plan(u.view)
		view = func(gtx C) D { return u.componentsView(gtx, rows, p) }
	case ViewUpdates:
		u.updateEvents(gtx)
		p, err = u.m.plan(u.view)
		view = u.updatesView
	}
	return vcol(gtx,
		layout.Flexed(1, func(gtx C) D {
			return layout.Flex{}.Layout(gtx, layout.Rigid(u.railView), layout.Rigid(vline), layout.Flexed(1, view))
		}),
		layout.Rigid(hline), layout.Rigid(func(gtx C) D { return u.footer(gtx, p, err) }))
}

func (u *ui) railView(gtx C) D {
	gtx.Constraints.Min.X, gtx.Constraints.Max.X = gtx.Dp(168), gtx.Dp(168)
	gtx.Constraints.Min.Y = gtx.Constraints.Max.Y
	names := []string{"Components", "Updates", "Settings"}
	if n := u.m.updateCount(); n > 0 {
		names[ViewUpdates] += fmt.Sprintf(" (%d)", n)
	}
	return bg(gtx, colPanel, func(gtx C) D {
		kids := []layout.FlexChild{gap(8)}
		for i, name := range names {
			sel := View(i) == u.view
			kids = append(kids, layout.Rigid(func(gtx C) D {
				return u.rail[i].Layout(gtx, func(gtx C) D {
					l := lbl(u.th, 15, name, colInk)
					if sel {
						l = bold(l)
					}
					item := func(gtx C) D {
						d := layout.Inset{Left: 18, Top: 10, Bottom: 10}.Layout(gtx, l.Layout)
						d.Size.X = gtx.Constraints.Max.X
						return d
					}
					return focused(gtx, &u.rail[i], func(gtx C) D {
						if !sel {
							return item(gtx)
						}
						return layout.Stack{}.Layout(gtx,
							layout.Expanded(func(gtx C) D { return fill(gtx, colWhite) }),
							layout.Expanded(func(gtx C) D { gtx.Constraints.Min.X = gtx.Dp(3); return fill(gtx, colAccent) }),
							layout.Stacked(item))
					})
				})
			}))
		}
		return vcol(gtx, kids...)
	})
}

func (u *ui) footer(gtx C, p *plan.Plan, err error) D {
	label := map[View]string{ViewComponents: "Apply changes", ViewUpdates: "Update selected"}[u.view]
	txt, sub := summary(p, u.view == ViewUpdates)
	col := colInk
	switch {
	case u.view == ViewSettings:
		txt, sub, col = "Flashpoint folder: "+u.root, "", colMuted
	case err != nil:
		txt, sub, col = sentence(err.Error()), "", colAccent
	case u.readOnly != "" && p != nil && !p.Empty():
		sub = "Changes are disabled (see the message at the top)."
	}
	canApply := err == nil && p != nil && !p.Empty() && u.readOnly == "" && !u.busy
	if u.apply.Clicked(gtx) && canApply {
		u.confirm(p, map[View]string{ViewComponents: "Apply these changes?", ViewUpdates: "Install these updates?"}[u.view], label, false)
	}
	if u.launcher.Clicked(gtx) {
		u.openLauncher()
	}
	buttons := []layout.FlexChild{btn(u.th, &u.launcher, "Open Launcher", false, !u.busy)}
	if label != "" {
		buttons = append(buttons, gap(8), btn(u.th, &u.apply, label, true, canApply))
	}
	return u.footerBar(gtx, txt, sub, col, buttons...)
}

// footerBar is the bar under a page: txt (at most two lines) in col, sub
// muted under it, and the buttons on the right.
func (u *ui) footerBar(gtx C, txt, sub string, col color.NRGBA, buttons ...layout.FlexChild) D {
	return bg(gtx, colPanel, func(gtx C) D {
		return layout.Inset{Left: 16, Right: 16, Top: 10, Bottom: 10}.Layout(gtx, func(gtx C) D {
			l := para(u.th, txt, col)
			l.MaxLines = 2
			kids := []layout.FlexChild{layout.Flexed(1, func(gtx C) D {
				return vcol(gtx, layout.Rigid(l.Layout), layout.Rigid(lbl(u.th, 13, sub, colMuted).Layout))
			}), gap(12)}
			return hrow(gtx, append(kids, buttons...)...)
		})
	})
}

// ---- Components ----

// query is the search text, lower case and trimmed.
func (u *ui) query() string { return strings.ToLower(strings.TrimSpace(u.search.Text())) }

// componentEvents handles the tree's clicks and keys (Space toggles the
// selected row, Ctrl+F goes to the search) and returns the rows to show.
func (u *ui) componentEvents(gtx C) []*node {
	m, q := u.m, u.query()
	rows := m.rows(q)
	if !slices.Contains(rows, m.selected) {
		m.selected = nil // hidden by the search
	}
	changed := false
	toggle := func(n *node) {
		if u.readOnly == "" {
			m.toggle(n, q)
			changed = true
		}
	}
	for _, n := range rows {
		if n.check.Clicked(gtx) {
			toggle(n)
			m.selected = n
			gtx.Execute(key.FocusCmd{}) // leave the search, so Space toggles
		}
		if n.expand.Clicked(gtx) {
			n.Open, changed = !n.Open, true
		}
		if n.sel.Clicked(gtx) {
			m.selected = n
			gtx.Execute(key.FocusCmd{})
		}
	}
	for {
		ev, ok := gtx.Event(key.Filter{Name: key.NameSpace}, key.Filter{Name: "F", Required: key.ModShortcut})
		if !ok {
			break
		}
		e, ok := ev.(key.Event)
		switch {
		case !ok || e.State != key.Press:
		case e.Name == "F":
			gtx.Execute(key.FocusCmd{Tag: &u.search})
		case m.selected != nil && gtx.Focused(nil): // not typing, and no button has the focus
			toggle(m.selected)
		}
	}
	if changed {
		rows = m.rows(q)
	}
	return rows
}

// componentsView shows the tree; p is its plan (nil if it has an error).
func (u *ui) componentsView(gtx C, rows []*node, p *plan.Plan) D {
	tree := func(gtx C) D {
		if len(rows) == 0 {
			return layout.UniformInset(16).Layout(gtx, lbl(u.th, 14, "No components match your search.", colMuted).Layout)
		}
		return material.List(u.th, &u.tree).Layout(gtx, len(rows), func(gtx C, i int) D { return u.treeRow(gtx, rows[i], p) })
	}
	return vcol(gtx,
		layout.Rigid(func(gtx C) D {
			return layout.Inset{Left: 12, Right: 12, Top: 10, Bottom: 10}.Layout(gtx, textField(u.th, &u.search, "Search components"))
		}),
		layout.Rigid(hline),
		layout.Flexed(1, func(gtx C) D {
			return layout.Flex{}.Layout(gtx, layout.Flexed(0.6, tree), layout.Rigid(vline), layout.Flexed(0.4, u.detailsView))
		}))
}

func (u *ui) treeRow(gtx C, n *node, p *plan.Plan) D {
	row := func(gtx C) D {
		return layout.Inset{Left: unit.Dp(6 + 18*n.Depth), Right: 12, Top: 3, Bottom: 3}.Layout(gtx, func(gtx C) D {
			return hrow(gtx,
				layout.Rigid(func(gtx C) D {
					sz := image.Pt(gtx.Dp(22), gtx.Dp(22))
					if !n.IsCat {
						return D{Size: sz}
					}
					ic := icRight
					if n.Open || u.query() != "" {
						ic = icDown
					}
					return n.expand.Layout(gtx, func(gtx C) D {
						return focused(gtx, &n.expand, func(gtx C) D {
							gtx.Constraints = layout.Exact(sz)
							return ic.Layout(gtx, colMuted)
						})
					})
				}),
				gap(2),
				layout.Rigid(func(gtx C) D {
					return checkbox(gtx, &n.check, u.m.state(n, u.query()), canToggle(n, u.query()) && u.readOnly == "")
				}),
				gap(6),
				layout.Flexed(1, func(gtx C) D {
					return n.sel.Layout(gtx, func(gtx C) D {
						return focused(gtx, &n.sel, func(gtx C) D { return u.rowLabel(gtx, n, p) })
					})
				}))
		})
	}
	if u.m.selected == n {
		return bg(gtx, colSel, row)
	}
	return row(gtx)
}

// rowLabel is the clickable part of a tree row: title, badges, size.
func (u *ui) rowLabel(gtx C, n *node, p *plan.Plan) D {
	th := u.th
	gtx.Constraints.Min.X = gtx.Constraints.Max.X
	title := lbl(th, 14, n.Title, colInk)
	if n.IsCat {
		title = bold(title)
	}
	var badges []layout.FlexChild
	for _, b := range rowBadges(n, p, u.inst != nil) {
		badges = append(badges, gap(6), layout.Rigid(badge(th, b.text, b.col)))
	}
	// Rigid children are laid out in order: the name comes first but
	// leaves room for the size and a short badge ("Installed"); the
	// badges leave room for the size.
	keep := unit.Dp(84)
	if len(badges) > 0 {
		keep += 70
	}
	return hrow(gtx,
		layout.Rigid(func(gtx C) D {
			gtx.Constraints.Max.X = max(0, gtx.Constraints.Max.X-gtx.Dp(keep))
			return title.Layout(gtx)
		}),
		flexSpace,
		layout.Rigid(func(gtx C) D {
			gtx.Constraints.Max.X = max(0, gtx.Constraints.Max.X-gtx.Dp(84))
			return hrow(gtx, badges...)
		}),
		fixed(84, func(gtx C) D {
			return layout.Inset{Left: 8}.Layout(gtx, rightAligned(lbl(th, 12, u.rowSize(n), colMuted)))
		}))
}

type badgeSpec struct {
	text string
	col  color.NRGBA
}

// rowBadges: what Apply will do to the row (including what the plan p
// added), then its status. The installer leaves out "Will be installed"
// (every checked row would have it) and the status (a fresh install's
// required components have status Missing).
func rowBadges(n *node, p *plan.Plan, installer bool) []badgeSpec {
	if n.IsCat {
		return nil
	}
	var out []badgeSpec
	if p != nil && slices.Contains(p.AddedDeps, n.ID) {
		out = append(out, badgeSpec{"Added as dependency", colBlue})
	} else if p != nil && slices.Contains(p.AddedDependents, n.ID) {
		out = append(out, badgeSpec{"Also removed", colAccent})
	} else if n.Want && n.Rec == nil && !installer {
		out = append(out, badgeSpec{"Will be installed", colBlue})
	} else if !n.Want && n.Rec != nil {
		out = append(out, badgeSpec{"Will be removed", colAccent})
	}
	b, ok := map[plan.Status]badgeSpec{plan.UpToDate: {"Installed", colGreen}, plan.Outdated: {"Update available", colAmber},
		plan.Missing: {"Missing", colAmber}, plan.Deprecated: {"Obsolete", colAccent}}[n.Status]
	if ok && !installer && !(n.Status == plan.UpToDate && len(out) > 0) { // "Will be removed" etc. imply "Installed"
		out = append(out, b)
	}
	if n.locked() {
		out = append(out, badgeSpec{"Required", colMuted})
	}
	return out
}

// counts returns the number of leaves below n, how many are installed, and their size on disk.
func (n *node) counts() (leaves, installed int, bytes int64) {
	for _, k := range n.leaves(nil) {
		leaves++
		if k.Rec != nil {
			installed++
		}
		if k.Comp != nil {
			bytes += k.Comp.InstallSize
		} else {
			bytes += k.Rec.InstallSize
		}
	}
	return
}

// rowSize is a component's size, or a category's installed/total count
// (selected/total in the installer, where nothing is installed yet).
func (u *ui) rowSize(n *node) string {
	all, inst, bytes := n.counts()
	if !n.IsCat {
		return size(bytes)
	}
	if u.inst != nil { // nothing is installed yet: count the selection
		inst = u.m.selectedIn(n)
	}
	return fmt.Sprintf("%d / %d", inst, all)
}

// selectedIn is the number of leaves below n that are checked.
func (m *model) selectedIn(n *node) int {
	k := 0
	for _, l := range n.leaves(nil) {
		if m.on(l) {
			k++
		}
	}
	return k
}

func (u *ui) detailsView(gtx C) D {
	th, m, n := u.th, u.m, u.m.selected
	var ws []layout.Widget
	add := func(w layout.Widget) { ws = append(ws, w) }
	space := func(dp unit.Dp) { add(layout.Spacer{Height: dp}.Layout) }
	prop := func(k, v string) {
		add(func(gtx C) D {
			return layout.Inset{Top: 3, Bottom: 3}.Layout(gtx, func(gtx C) D {
				return layout.Flex{}.Layout(gtx, fixed(100, lbl(th, 13, k, colMuted).Layout), layout.Flexed(1, para(th, v, colInk).Layout))
			})
		})
	}
	if n == nil {
		add(bold(lbl(th, 16, "Choose your components", colInk)).Layout)
		space(8)
		intro := "Check the components you want in your Flashpoint and uncheck the ones you want removed, then press Apply changes."
		if u.inst != nil {
			intro = "Check the components you want in your Flashpoint, then press Install. You can add or remove components later in the Flashpoint Manager."
		}
		add(para(th, intro, colInk).Layout)
		space(8)
		add(para(th, "Click a component or category to learn more about it. Required components can't be removed.", colMuted).Layout)
	} else {
		t := bold(para(th, n.Title, colInk))
		t.TextSize = 17
		add(t.Layout)
		if n.ID != obsoleteID {
			add(lbl(th, 12, n.ID, colMuted).Layout)
		}
		space(10)
		if n.Desc != "" {
			add(para(th, n.Desc, colInk).Layout)
			space(12)
		}
		status := map[plan.Status]string{plan.NotInstalled: "Not installed", plan.UpToDate: "Installed, up to date",
			plan.Outdated: "Installed, update available", plan.Missing: "Not installed (required)",
			plan.Deprecated: "Installed, but no longer in the component list"}[n.Status]
		switch c := n.Comp; {
		case n.IsCat:
			all, inst, bytes := n.counts()
			if u.inst != nil {
				prop("Components", fmt.Sprintf("%d selected of %d", m.selectedIn(n), all))
			} else {
				prop("Components", fmt.Sprintf("%d installed of %d", inst, all))
			}
			prop("Size on disk", size(bytes)+" for all")
		case c != nil:
			if p := m.tree(); p != nil && slices.Contains(p.AddedDeps, n.ID) {
				add(para(th, "Will be installed because "+m.whoNeeds(p, n.ID)+" needs it. To leave it out, uncheck that.", colBlue).Layout)
				space(10)
			} else if p != nil && slices.Contains(p.AddedDependents, n.ID) {
				add(para(th, "Will be removed because it needs "+m.removedDep(p, n.ID)+". Check it to keep both.", colAccent).Layout)
				space(10)
			}
			prop("Status", status)
			if c.HasArchive() {
				prop("Download", size(c.DownloadSize))
			}
			prop("Size on disk", size(c.InstallSize))
			if !c.Modified.IsZero() {
				prop("Updated", date(c.Modified))
			}
			prop("Folder", listOr([]string{winPath(c.Path)}, "Flashpoint folder"))
			prop("Depends on", listOr(m.titles(c.Depends), "Nothing"))
			prop("Needed by", listOr(m.titles(m.neededBy[c.ID]), "Nothing"))
			if c.Required {
				space(8)
				add(para(th, "This component is required and can't be removed.", colMuted).Layout)
			}
		default:
			prop("Status", status)
			prop("Size on disk", size(n.Rec.InstallSize))
			prop("Depends on", listOr(n.Rec.Depends, "Nothing"))
		}
	}
	return layout.UniformInset(16).Layout(gtx, func(gtx C) D {
		return material.List(th, &u.details).Layout(gtx, len(ws), func(gtx C, i int) D { return ws[i](gtx) })
	})
}

// listOr joins s, or returns empty when s has no non-empty element.
func listOr(s []string, empty string) string {
	if j := strings.Join(s, ", "); j != "" {
		return j
	}
	return empty
}

// ---- Updates ----

func (u *ui) updateEvents(gtx C) {
	m := u.m
	if u.selectAll.Clicked(gtx) {
		v := m.updatesChecked() != checked
		for _, r := range m.updates {
			r.checked = v
		}
	}
	for _, r := range m.updates {
		if r.check.Clicked(gtx) || r.row.Clicked(gtx) {
			r.checked = !r.checked
		}
	}
}

func (u *ui) updatesView(gtx C) D {
	th, m := u.th, u.m
	if len(m.updates) == 0 {
		return centered(gtx, 400, func(gtx C) D {
			return layout.Flex{Axis: layout.Vertical, Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(bold(lbl(th, 17, "Everything is up to date", colInk)).Layout), gap(6),
				layout.Rigid(lbl(th, 14, "There are no updates for your installed components.", colMuted).Layout))
		})
	}
	header := func(gtx C) D {
		return layout.Inset{Left: 16, Right: 16, Top: 10, Bottom: 10}.Layout(gtx, func(gtx C) D {
			return hrow(gtx,
				layout.Rigid(func(gtx C) D { return checkbox(gtx, &u.selectAll, m.updatesChecked(), true) }),
				gap(8), layout.Flexed(1, lbl(th, 14, "Select all", colInk).Layout),
				fixed(100, lbl(th, 12, "Updated", colMuted).Layout),
				fixed(90, rightAligned(lbl(th, 12, "Disk change", colMuted))))
		})
	}
	return vcol(gtx, layout.Rigid(header), layout.Rigid(hline), layout.Flexed(1, func(gtx C) D {
		return material.List(th, &u.updList).Layout(gtx, len(m.updates), func(gtx C, i int) D { return u.updateRow(gtx, m.updates[i]) })
	}))
}

func (u *ui) updateRow(gtx C, r *updRow) D {
	th, n := u.th, r.n
	var tag badgeSpec
	var delta int64
	updated := ""
	switch n.Status {
	case plan.Outdated:
		tag, delta = badgeSpec{"Update", colAmber}, n.Comp.InstallSize-n.Rec.InstallSize
	case plan.Missing: // required, no record
		tag, delta = badgeSpec{"Required, not installed", colAmber}, n.Comp.InstallSize
	default: // obsolete: a record, no component
		tag, delta = badgeSpec{"Obsolete: remove", colAccent}, -n.Rec.InstallSize
	}
	if n.Comp != nil {
		updated = date(n.Comp.Modified)
	}
	state := unchecked
	if r.checked {
		state = checked
	}
	return layout.Inset{Left: 16, Right: 16, Top: 6, Bottom: 6}.Layout(gtx, func(gtx C) D {
		return hrow(gtx,
			layout.Rigid(func(gtx C) D { return checkbox(gtx, &r.check, state, true) }),
			gap(8),
			layout.Flexed(1, func(gtx C) D {
				return r.row.Layout(gtx, func(gtx C) D {
					return focused(gtx, &r.row, func(gtx C) D {
						gtx.Constraints.Min.X = gtx.Constraints.Max.X
						return vcol(gtx,
							layout.Rigid(func(gtx C) D {
								return hrow(gtx, layout.Rigid(lbl(th, 14, n.Title, colInk).Layout), gap(8), layout.Rigid(badge(th, tag.text, tag.col)))
							}),
							layout.Rigid(lbl(th, 12, n.Desc, colMuted).Layout))
					})
				})
			}),
			fixed(100, lbl(th, 13, updated, colMuted).Layout),
			fixed(90, rightAligned(lbl(th, 13, signedSize(delta), colInk))))
	})
}

// ---- Modal ----

type modal struct {
	title  string
	body   []string
	labels []string
	btns   []widget.Clickable
	list   widget.List
	onPick func(int) // nil: just close

	enter, escape int // the button the key picks; -1: the key does nothing
}

func newModal(title string, body, labels []string, onPick func(int)) *modal {
	md := &modal{title: title, body: body, labels: labels, btns: make([]widget.Clickable, len(labels)), onPick: onPick, enter: -1, escape: -1}
	md.list.Axis = layout.Vertical
	return md
}

func (u *ui) modalView(gtx C) D {
	md := u.modal
	pick := -1
	for {
		ev, ok := gtx.Event(key.Filter{Name: key.NameEscape}, key.Filter{Name: key.NameReturn}, key.Filter{Name: key.NameEnter})
		if !ok {
			break
		}
		if e, ok := ev.(key.Event); ok && e.State == key.Press {
			pick = map[bool]int{true: md.escape, false: md.enter}[e.Name == key.NameEscape]
		}
	}
	for i := range md.btns {
		if md.btns[i].Clicked(gtx) {
			pick = i
		}
	}
	if pick >= 0 {
		u.modal = nil
		if md.onPick != nil {
			md.onPick(pick)
		}
		gtx.Execute(op.InvalidateCmd{})
		return D{Size: gtx.Constraints.Max}
	}
	gtx.Constraints.Min = gtx.Constraints.Max
	fill(gtx, colScrim)
	return centered(gtx, 540, func(gtx C) D {
		return layout.Stack{}.Layout(gtx,
			layout.Expanded(func(gtx C) D {
				r := image.Rectangle{Max: gtx.Constraints.Min}
				defer clip.UniformRRect(r, gtx.Dp(6)).Push(gtx.Ops).Pop()
				paint.Fill(gtx.Ops, colWhite)
				return D{Size: r.Max}
			}),
			layout.Stacked(func(gtx C) D {
				return layout.UniformInset(22).Layout(gtx, func(gtx C) D {
					gtx.Constraints.Min.X = gtx.Constraints.Max.X
					buttons := []layout.FlexChild{flexSpace}
					for i, l := range md.labels {
						buttons = append(buttons, gap(8), btn(u.th, &md.btns[i], l, i == 0, true))
					}
					return vcol(gtx,
						layout.Rigid(bold(lbl(u.th, 17, md.title, colInk)).Layout),
						gap(12),
						layout.Rigid(func(gtx C) D {
							gtx.Constraints.Max.Y = min(gtx.Constraints.Max.Y, gtx.Dp(330))
							return material.List(u.th, &md.list).Layout(gtx, len(md.body), func(gtx C, i int) D {
								// A line indented with spaces (a list item) stays indented when it wraps.
								txt, in := md.body[i], unit.Dp(0)
								if t, ok := strings.CutPrefix(txt, "    "); ok {
									txt, in = t, 16
								}
								return layout.Inset{Left: in, Bottom: 4}.Layout(gtx, para(u.th, txt, colInk).Layout)
							})
						}),
						gap(18),
						layout.Rigid(row(buttons...)))
				})
			}))
	})
}

// ---- Progress ----

func (u *ui) runView(gtx C) D {
	r := u.run
	if u.cancelB.Clicked(gtx) {
		r.requestCancel()
	}
	if r.finished {
		switch {
		case u.done.Clicked(gtx):
			if u.inst != nil { // back to the components, selection kept
				u.run, u.screen = nil, scrMain
			} else {
				u.reload()
			}
			return D{}
		case u.restart.Clicked(gtx):
			u.restartManager()
			return D{}
		case u.launcher.Clicked(gtx):
			u.openLauncher()
		case u.quitBtn.Clicked(gtx):
			u.quit = true
		}
	} else {
		gtx.Execute(op.InvalidateCmd{At: gtx.Now.Add(500 * time.Millisecond)}) // speeds and countdowns
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	th := u.th

	var kids []layout.FlexChild
	add := func(w layout.Widget) { kids = append(kids, layout.Rigid(w)) }
	if !r.finished {
		title, status, frac := "Applying changes", r.status(), r.fraction()
		if u.download() {
			title = "Installing components"
		} else if u.inst != nil {
			title = "Installing Flashpoint"
		}
		if u.closeAfterRun {
			status += " The window closes when this is finished."
		}
		add(bold(lbl(th, 19, title, colInk)).Layout)
		add(layout.Spacer{Height: 8}.Layout)
		add(lbl(th, 14, status, colMuted).Layout)
		add(layout.Spacer{Height: 10}.Layout)
		add(func(gtx C) D {
			pb := material.ProgressBar(th, frac)
			pb.Color, pb.Height, pb.TrackColor = colAccent, 8, colLine
			return pb.Layout(gtx)
		})
	} else {
		title, detail, failed := r.outcome(u.inst != nil)
		add(bold(lbl(th, 19, title, colInk)).Layout)
		add(layout.Spacer{Height: 8}.Layout)
		if c := r.counts(); c != "" {
			add(para(th, c+".", colInk).Layout)
		}
		if detail != "" {
			add(para(th, detail, map[bool]color.NRGBA{false: colInk, true: colAccent}[failed]).Layout)
		}
		if len(u.dlUnknown) > 0 {
			add(para(th, "Not in the component list: "+strings.Join(u.dlUnknown, ", ")+".", colAccent).Layout)
		}
		if r.managerUpdated() {
			add(para(th, "The manager itself was updated. Restart it to use the new version.", colInk).Layout)
		}
	}
	add(layout.Spacer{Height: 14}.Layout)
	add(hline)
	kids = append(kids, layout.Flexed(1, func(gtx C) D {
		return material.List(th, &r.list).Layout(gtx, len(r.rows), func(gtx C, i int) D { return u.runRow(gtx, r, r.rows[i]) })
	}))
	add(hline)
	add(layout.Spacer{Height: 12}.Layout)
	add(row(u.runButtons(r)...))
	return layout.Inset{Left: 24, Right: 24, Top: 18, Bottom: 14}.Layout(gtx, func(gtx C) D { return vcol(gtx, kids...) })
}

func (u *ui) runButtons(r *run) []layout.FlexChild {
	kids := []layout.FlexChild{flexSpace}
	switch {
	case !r.finished:
		return append(kids, btn(u.th, &u.cancelB, "Cancel", false, !r.committing && !r.cancelling))
	case u.download():
		return append(kids, btn(u.th, &u.quitBtn, "Close", true, true))
	case u.inst != nil: // a success goes on to the finish page
		return append(kids, btn(u.th, &u.quitBtn, "Quit", false, true), gap(8), btn(u.th, &u.done, "Back", true, !u.busy))
	}
	restart := r.managerUpdated()
	kids = append(kids, btn(u.th, &u.launcher, "Open Launcher", false, true), gap(8), btn(u.th, &u.done, "Back to components", !restart, !u.busy))
	if restart {
		kids = append(kids, gap(8), btn(u.th, &u.restart, "Restart manager", true, true))
	}
	return kids
}

// status describes what the run is doing now; r.mu is held.
func (r *run) status() string {
	switch {
	case r.cancelling:
		return "Cancelling…"
	case r.committing:
		return "Moving files into place. This can't be cancelled; it keeps your Flashpoint folder consistent."
	}
	var done, total int64
	var speed float64
	for i, row := range r.rows {
		if row == r.extracting {
			return fmt.Sprintf("Extracting %s (%d of %d)…", row.Title, i+1, len(r.rows))
		}
		done, total = done+min(row.Done, row.Size), total+row.Size
		if row.St == rsDownloading {
			speed += row.speed
		}
	}
	if total == 0 {
		return "Preparing…"
	}
	s := fmt.Sprintf("Downloaded %s of %s", size(done), size(total))
	if speed > 0 {
		s += fmt.Sprintf("  ·  %s/s", size(int64(speed)))
	}
	return s
}

// fraction is the download progress, then the extraction progress.
func (r *run) fraction() float32 {
	if r.committing {
		return 1
	}
	var done, total int64
	extracted, archives := 0, 0
	for _, row := range r.rows {
		done, total = done+min(row.Done, row.Size), total+row.Size
		if row.Size > 0 {
			archives++
			if row.St == rsExtracting {
				extracted++
			}
		}
	}
	switch {
	case extracted > 0:
		return float32(extracted) / float32(archives)
	case total > 0:
		return float32(done) / float32(total)
	}
	return 0
}

// rowText is a progress row's state in words; r.mu is held.
func (r *run) rowText(row *runRow) (string, color.NRGBA) {
	remove := row.Verb == "Remove"
	switch {
	case row.St == rsFailed:
		return "Failed", colAccent
	case row.St == rsSkipped:
		return "Skipped", colAmber
	case r.finished && r.err == nil:
		return map[string]string{"Install": "Installed", "Update": "Updated", "Repair": "Repaired", "Remove": "Removed"}[row.Verb], colGreen
	case r.finished:
		return "Not changed", colMuted
	case r.committing && remove:
		return "Removing…", colInk
	case r.committing:
		return "Installing…", colInk
	case row.St == rsRetrying:
		secs := int(time.Until(row.RetryAt).Round(time.Second).Seconds())
		return fmt.Sprintf("Connection problem, retrying in %d s (attempt %d)", max(secs, 0), row.Attempt+1), colAmber
	case row == r.extracting:
		return "Extracting…", colInk
	case row.St == rsExtracting:
		return "Ready", colMuted
	case row.St == rsDownloading && row.Total > 0 && row.Done >= row.Total:
		return "Downloaded and verified", colMuted
	case row.St == rsDownloading:
		txt := fmt.Sprintf("Downloading %s of %s", size(row.Done), size(max(row.Total, row.Size)))
		if row.speed > 0 {
			txt += fmt.Sprintf(" at %s/s", size(int64(row.speed)))
		}
		return txt, colInk
	case remove:
		return "Will be removed", colMuted
	}
	return "Waiting", colMuted
}

func (u *ui) runRow(gtx C, r *run, row *runRow) D {
	th := u.th
	txt, col := r.rowText(row)
	return layout.Inset{Top: 7, Bottom: 7}.Layout(gtx, func(gtx C) D {
		return hrow(gtx,
			fixed(70, lbl(th, 12, row.Verb, colMuted).Layout),
			layout.Flexed(0.45, lbl(th, 14, row.Title, colInk).Layout),
			gap(12),
			layout.Flexed(0.55, func(gtx C) D {
				return vcol(gtx, layout.Rigid(lbl(th, 13, txt, col).Layout), layout.Rigid(func(gtx C) D {
					if row.St != rsDownloading || row.Size == 0 || row.Done >= row.Size || r.committing || r.finished {
						return D{}
					}
					pb := material.ProgressBar(th, float32(min(row.Done, row.Size))/float32(row.Size))
					pb.Color, pb.Height, pb.TrackColor = colAccent, 4, colLine
					return layout.Inset{Top: 4}.Layout(gtx, pb.Layout)
				}))
			}))
	})
}

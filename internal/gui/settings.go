package gui

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/engine"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/plan"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
)

const devSource = "https://nexus-dev.flashpointarchive.org/repository/development/components.xml"

type settingsState struct {
	folder, custom                    widget.Editor
	source                            widget.Enum
	browse, save, check, repair, logs widget.Clickable
	uninstall                         widget.Clickable
	err                               string

	checking, checked bool
	checkErr          string
	total             int
	broken            []engine.VerifyResult
}

// reset shows the current values. The channel picks the radio, so a legacy
// URL of the stable channel still shows as Stable.
func (s *settingsState) reset(root, source, channel string) {
	s.folder.SetText(root)
	s.custom.SetText("")
	s.err, s.checked, s.checkErr, s.broken = "", false, "", nil
	switch channel {
	case "stable", "development":
		s.source.Value = channel
	default:
		s.source.Value = "custom"
		s.custom.SetText(source)
	}
}

// chosenSource is the source the radios select; the current source while
// its channel is still selected.
func (u *ui) chosenSource() string {
	switch v := u.set.source.Value; {
	case v == u.channel:
		return u.source
	case v == "stable":
		return state.DefaultSource
	case v == "development":
		return devSource
	}
	return strings.TrimSpace(u.set.custom.Text())
}

// settingsEvents handles the page's input and returns why Save is
// disabled ("" if it is not).
func (u *ui) settingsEvents(gtx C) string {
	s := &u.set
	if s.source.Update(gtx) {
		s.err = ""
	}
	for _, ed := range []*widget.Editor{&s.folder, &s.custom} {
		for {
			ev, ok := ed.Update(gtx)
			if !ok {
				break
			}
			if _, ok := ev.(widget.ChangeEvent); ok {
				s.err = "" // it may be about the old text
			}
		}
	}
	if s.browse.Clicked(gtx) {
		u.browseInto(&s.folder, &s.err, nil)
	}
	problem := u.saveProblem()
	if s.save.Clicked(gtx) && u.settingsChanged() && problem == "" && !u.busy {
		// fpm.cfg is small and local: no worker needed.
		if err := u.opts.SaveSettings(cleanPath(s.folder.Text()), u.chosenSource()); err != nil {
			s.err = "Could not save the settings: " + sentence(err.Error())
		} else {
			u.start("")
		}
	}
	if s.check.Clicked(gtx) && !u.busy {
		s.checking, s.checked = true, false
		eng := u.eng
		u.goWork(func(ctx context.Context) func() {
			res, err := eng.Verify(ctx, nil)
			return func() {
				s.checking, s.checked, s.broken, s.total, s.checkErr = false, true, nil, len(res), ""
				if err != nil {
					s.checkErr = sentence(err.Error())
				}
				for _, r := range res {
					if r.Err != nil || len(r.Missing) > 0 {
						s.broken = append(s.broken, r)
					}
				}
			}
		})
	}
	if s.repair.Clicked(gtx) && !u.busy {
		p, err := u.m.resolve(plan.Request{Reinstall: u.repairable(), AutoDeps: true})
		if err != nil {
			u.alert("Cannot repair", err)
		} else {
			u.confirm(p, "Repair these components?", "Repair", true)
		}
	}
	if s.logs.Clicked(gtx) {
		if err := u.opts.OpenLogs(); err != nil {
			u.alert("Could not open the logs folder", err)
		}
	}
	if s.uninstall.Clicked(gtx) && !u.busy {
		if err := u.opts.CanUninstall(u.root); err != nil {
			u.alert("Flashpoint can't be uninstalled here", err)
		} else {
			u.confirmUninstall()
		}
	}
	return problem
}

func (u *ui) confirmUninstall() {
	u.modal = newModal("Uninstall Flashpoint?", []string{
		"Flashpoint will be deleted from " + u.root + ", including downloaded games, save data, playlists and favorites. This can't be undone.",
		"Close Flashpoint first.",
	}, []string{"Uninstall", "Cancel"}, func(i int) {
		if i == 0 {
			u.uninstall()
		}
	})
	u.modal.escape = 1 // and no Enter: deleting everything needs a click
}

// uninstall deletes the Flashpoint folder. The engine is closed first: its
// lock file is in that folder.
func (u *ui) uninstall() {
	root := u.root
	u.closeEngine()
	u.screen, u.msg = scrLoading, "Uninstalling Flashpoint…"
	u.goWork(func(context.Context) func() {
		kept, err := u.opts.Uninstall(root)
		return func() {
			u.logErr("the uninstall was incomplete", err)
			u.screen, u.uninstallKept, u.uninstallErr = scrUninstalled, kept, err
		}
	})
}

func (u *ui) uninstalledView(gtx C) D {
	if u.quitBtn.Clicked(gtx) {
		u.quit = true
	}
	title, text := "Flashpoint was uninstalled", u.root+" was deleted."
	if u.uninstallErr != nil {
		title, text = "Flashpoint was not fully uninstalled", plain(u.uninstallErr, "")+
			" Close Flashpoint, then delete the rest of "+u.root+" by hand."
	}
	var rows []layout.FlexChild
	if len(u.uninstallKept) > 0 {
		if u.uninstallErr == nil {
			text = "Flashpoint was deleted."
		}
		rows = append(rows, layout.Rigid(para(u.th, "Kept in "+u.root+", because they are not part of Flashpoint: "+
			strings.Join(u.uninstallKept, ", ")+".", colMuted).Layout))
	}
	return u.message(gtx, title, text, rows, btn(u.th, &u.quitBtn, "Close", true, true))
}

// settingsChanged reports whether Save has something to do. An empty custom
// URL is not a change (SaveSettings would keep the current source).
func (u *ui) settingsChanged() bool {
	src := u.chosenSource()
	return src != "" && (filepath.Clean(cleanPath(u.set.folder.Text())) != filepath.Clean(u.root) || src != u.source)
}

// saveProblem is what is wrong with the changed values, "" if nothing is.
func (u *ui) saveProblem() string {
	switch f, src := cleanPath(u.set.folder.Text()), u.chosenSource(); {
	case f == "":
		return "Enter your Flashpoint folder."
	case filepath.Clean(f) != filepath.Clean(u.root) && !state.IsRoot(f):
		return notRoot(f)
	case src != u.source && src != "" && !strings.HasPrefix(strings.ToLower(src), "https://"):
		return "Custom URLs must start with https://"
	}
	return ""
}

// repairable is the broken components that are still in the component list.
func (u *ui) repairable() []string {
	var ids []string
	for _, r := range u.set.broken {
		if _, ok := u.m.snap.Manifest.Components[r.ID]; ok {
			ids = append(ids, r.ID)
		}
	}
	return ids
}

func (u *ui) settingsView(gtx C) D {
	problem := u.settingsEvents(gtx)
	th, s := u.th, &u.set
	var ws []layout.Widget
	add := func(w ...layout.Widget) { ws = append(ws, w...) }
	space := func(dp unit.Dp) { add(layout.Spacer{Height: dp}.Layout) }
	heading := func(t string) {
		space(14)
		add(bold(lbl(th, 15, t, colInk)).Layout)
		space(8)
	}

	heading("Flashpoint folder")
	add(u.folderRow(&s.folder, &s.browse, `C:\Flashpoint`))

	heading("Component source")
	add(func(gtx C) D {
		radio := func(key, label string) layout.FlexChild {
			return layout.Rigid(func(gtx C) D {
				rb := material.RadioButton(th, &s.source, key, label)
				rb.IconColor = colAccent
				return rb.Layout(gtx)
			})
		}
		return hrow(gtx, radio("stable", "Stable"), gap(16), radio("development", "Development"), gap(16), radio("custom", "Custom URL"))
	})
	if s.source.Value == "custom" {
		space(6)
		add(textField(th, &s.custom, "https://example.org/components.xml"))
	}
	space(12)
	msg := s.err
	if msg == "" {
		msg = problem
	}
	add(row(btn(th, &s.save, "Save settings", true, u.settingsChanged() && problem == "" && !u.busy), gap(12), layout.Flexed(1, para(th, msg, colAccent).Layout)))

	heading("Check files")
	add(para(th, "Checks that every file of your installed components is still there.", colMuted).Layout)
	space(8)
	label := "Check files"
	switch {
	case s.checking:
		label = "Checking…"
	case s.checked:
		label = "Check again"
	}
	add(row(btn(th, &s.check, label, false, !u.busy), gap(12), layout.Flexed(1, para(th, u.checkSummary(), colInk).Layout)))
	if s.checked && len(s.broken) > 0 {
		space(8)
		for _, r := range s.broken {
			add(u.brokenRow(r))
		}
		space(6)
		add(para(th, "If missing files come back after a repair, your antivirus may be removing them. Add an exception for the Flashpoint folder.", colMuted).Layout)
		if ids := u.repairable(); len(ids) > 0 {
			space(8)
			add(row(btn(th, &s.repair, "Repair "+plural(len(ids), "component"), true, !u.busy && u.readOnly == "")))
		}
	}

	heading("Uninstall")
	add(para(th, "Deletes Flashpoint, with its games, save data, playlists and favorites.", colMuted).Layout)
	space(8)
	add(row(btn(th, &s.uninstall, "Uninstall Flashpoint…", false, !u.busy)))

	heading("About")
	add(lbl(th, 14, "Flashpoint Manager "+u.opts.Version, colInk).Layout)
	space(8)
	add(row(btn(th, &s.logs, "Open logs folder", false, true)))
	space(16)

	return layout.Inset{Left: 24, Right: 24}.Layout(gtx, func(gtx C) D {
		return material.List(th, &u.setList).Layout(gtx, len(ws), func(gtx C, i int) D {
			gtx.Constraints.Max.X = min(gtx.Constraints.Max.X, gtx.Dp(640))
			return ws[i](gtx)
		})
	})
}

func (u *ui) checkSummary() string {
	s := &u.set
	switch {
	case s.checking || !s.checked:
		return ""
	case s.checkErr != "":
		return s.checkErr
	case len(s.broken) == 0:
		return fmt.Sprintf("All %s are complete.", plural(s.total, "installed component"))
	}
	return fmt.Sprintf("%s of %d have problems:", plural(len(s.broken), "component"), s.total)
}

func (u *ui) brokenRow(r engine.VerifyResult) layout.Widget {
	th := u.th
	what := ""
	if r.Err != nil {
		what = sentence(r.Err.Error())
	} else {
		shown := slices.Clone(r.Missing[:min(len(r.Missing), 3)])
		for i, f := range shown {
			shown[i] = winPath(f)
		}
		what = plural(len(r.Missing), "missing file") + ": " + strings.Join(shown, ", ")
		if len(r.Missing) > len(shown) {
			what += fmt.Sprintf(" and %d more", len(r.Missing)-len(shown))
		}
		what += "."
	}
	if _, ok := u.m.snap.Manifest.Components[r.ID]; !ok {
		what += " It is no longer in the component list, so it can't be repaired; remove it on the Components page."
	}
	return func(gtx C) D {
		return layout.Inset{Top: 3, Bottom: 3}.Layout(gtx, func(gtx C) D {
			return layout.Flex{}.Layout(gtx,
				fixed(150, lbl(th, 14, u.m.title(r.ID), colInk).Layout),
				layout.Flexed(1, para(th, what, colMuted).Layout))
		})
	}
}

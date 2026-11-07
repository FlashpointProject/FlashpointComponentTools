package gui

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/widget"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/engine"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
)

// The installer wizard (docs/go-rewrite/M3.md): destination (scrDest),
// components (scrMain without the rail), progress (scrRun), the finishing
// steps (scrLoading) and the finish page (scrFinish).

// InstallOptions are provided by cmd/installer; the GUI does not import
// internal/install, so tests can pass fakes.
type InstallOptions struct {
	DefaultDest string
	// CheckDest validates the destination and returns it absolute. create
	// is the outermost folder that doesn't exist yet and will be created
	// (abs or a missing ancestor), "" when abs exists (empty). Errors are
	// written for the user.
	CheckDest func(path string) (abs, create string, err error)
	// Cleanup removes an abandoned install: create when not "", else
	// everything inside dir.
	Cleanup         func(dir, create string) error
	NeedsVCRedist   func() bool
	InstallVCRedist func(ctx context.Context) (rebootNeeded bool, err error)
	CreateShortcuts func(root string, desktop, startMenu bool) error
}

// vcRedistLink is the runtime installer as shown (the labels can't be
// selected, so short enough to retype; https:// is implied).
const vcRedistLink = "aka.ms/vs/17/release/vc_redist.x86.exe"

type installState struct {
	opts *InstallOptions

	path                               widget.Editor
	browse, next, back, install, close widget.Clickable
	desktopB, startMenuB, launchB      widget.Clickable
	desktop, startMenu, launch         bool
	vcMissing                          bool   // the Visual C++ runtime is installed too (the page says so)
	err                                string // under the path field

	dest      string   // set while there is something to remove if the install is abandoned
	create    string   // for Cleanup: the outermost folder Next creates, "" if dest existed
	installed bool     // Execute succeeded: never remove anything
	notes     []string // for the finish page
}

func newInstallState(opts *InstallOptions) *installState {
	s := &installState{opts: opts, desktop: true, startMenu: true, launch: true, vcMissing: opts.NeedsVCRedist()}
	s.path.SingleLine = true // and no Submit: Enter must not start anything
	s.path.SetText(opts.DefaultDest)
	return s
}

// ---- Destination ----

func (u *ui) destView(gtx C) D {
	s := u.inst
	for {
		ev, ok := s.path.Update(gtx)
		if !ok {
			break
		}
		if _, ok := ev.(widget.ChangeEvent); ok {
			s.err = "" // it is about the old text
		}
	}
	if s.next.Clicked(gtx) && !u.busy {
		u.openDest()
		gtx.Execute(op.InvalidateCmd{})
	}
	if s.browse.Clicked(gtx) {
		u.browseInto(&s.path, &s.err, inFlashpointFolder)
	}
	if u.quitBtn.Clicked(gtx) {
		u.quit = true
	}
	rows := []layout.FlexChild{
		layout.Rigid(u.folderRow(&s.path, &s.browse, `For example D:\Flashpoint`)),
		layout.Rigid(func(gtx C) D { return layout.Inset{Top: 8}.Layout(gtx, para(u.th, s.err, colAccent).Layout) }),
		gap(12),
		layout.Rigid(func(gtx C) D { return u.option(gtx, &s.desktopB, &s.desktop, "Desktop shortcut") }),
		gap(6),
		layout.Rigid(func(gtx C) D { return u.option(gtx, &s.startMenuB, &s.startMenu, "Start menu shortcut") }),
	}
	if s.vcMissing {
		rows = append(rows, gap(14), layout.Rigid(para(u.th,
			"The Visual C++ runtime, which Flashpoint needs, will be installed too (Windows will ask for permission).", colMuted).Layout))
	}
	return u.message(gtx, "Install Flashpoint",
		"Choose a new or empty folder for Flashpoint. You choose the components on the next page.", rows,
		btn(u.th, &s.next, "Next", true, !u.busy), gap(8), btn(u.th, &u.quitBtn, "Quit", false, true))
}

// inFlashpointFolder is the destination Browse… suggests for dir: dir if
// it is empty or a Flashpoint folder (CheckDest then says it is installed
// there), else a Flashpoint folder in it (CheckDest only accepts an empty
// folder, and a picked drive or Games folder is meant as the place).
func inFlashpointFolder(dir string) string {
	if ents, _ := os.ReadDir(dir); len(ents) > 0 && !state.IsRoot(dir) {
		return filepath.Join(dir, "Flashpoint")
	}
	return dir
}

// openDest checks the destination on a worker: CheckDest looks at the
// disk, which blocks for seconds on an unreachable network path.
func (u *ui) openDest() {
	s := u.inst
	path := s.path.Text()
	u.screen, u.msg = scrLoading, "Loading component list…"
	u.goWork(func(context.Context) func() {
		abs, create, err := s.opts.CheckDest(path)
		return func() {
			if err != nil {
				s.err, u.screen = err.Error(), scrDest // written for the user
				return
			}
			u.prepareDest(abs, create)
		}
	})
}

// prepareDest creates the destination, opens and locks an engine on it and
// loads the component list. From the start, closing the window removes the
// destination again (closeInstall); a failure removes it too and shows the
// error on the destination page.
func (u *ui) prepareDest(abs, create string) {
	s := u.inst
	s.err, s.dest, s.create = "", abs, create
	u.goWork(func(ctx context.Context) func() {
		what := "The folder could not be prepared."
		eng, err := u.opts.Open(abs)
		if err == nil {
			_, err = eng.Lock()
		}
		var snap *engine.Snapshot
		if err == nil {
			what = "Could not load the component list."
			lctx, cancel := context.WithTimeout(ctx, loadTimeout)
			snap, err = eng.LoadManifest(lctx, engine.ForChanges)
			cancel()
		}
		if err != nil {
			u.logErr("could not open the destination", err)
			if eng != nil {
				eng.Close() // before Cleanup: the lock file is in the destination
			}
			msg := what + " " + plain(err, "The component list")
			return func() {
				if problem := u.abandon(); problem != "" {
					msg += " " + problem
				}
				s.err, u.screen = msg, scrDest
			}
		}
		return func() {
			u.eng, u.root, u.source, u.channel = eng, eng.Root(), eng.Source(), eng.Channel()
			u.setInstallSnapshot(snap)
		}
	})
}

// setInstallSnapshot shows the components step with the legacy
// installer's defaults: everything but extra*.
func (u *ui) setInstallSnapshot(snap *engine.Snapshot) {
	u.setSnapshot(snap)
	for _, n := range u.m.comps {
		if c := n.Comp; c != nil {
			n.Want = c.Required || c.DefaultOn
		}
	}
}

// backToDest leaves the components step: the destination is removed, and
// Next creates it again.
func (u *ui) backToDest() {
	u.inst.err = u.abandon()
	u.m, u.warnings, u.screen = nil, nil, scrDest
}

// abandon closes the engine and removes what the installer created in the
// destination. It returns the problem in words ("" if none).
func (u *ui) abandon() string {
	s := u.inst
	u.closeEngine() // first: the lock file is in the destination
	err := s.opts.Cleanup(s.dest, s.create)
	s.dest, s.create = "", "" // tried once; a failure is reported, not retried
	if err != nil {
		u.logErr("could not remove the unfinished install", err)
		return sentence(err.Error())
	}
	return ""
}

// closeInstall runs when the window has closed and the engine is closed:
// an install that did not finish Execute is removed, and a finished one
// starts the launcher when the finish page asks for it.
func (u *ui) closeInstall() {
	switch s := u.inst; {
	case s == nil:
	case s.installed:
		if u.screen == scrFinish && s.launch {
			u.logErr("could not start the launcher", u.opts.OpenLauncher())
		}
	case s.dest != "":
		u.abandon()
	}
}

// ---- Components ----

// installComponentsView is the components step: the manager's tree and
// details without the rail, with Back / Install.
func (u *ui) installComponentsView(gtx C) D {
	s := u.inst
	if s.back.Clicked(gtx) && !u.busy {
		u.backToDest()
		gtx.Execute(op.InvalidateCmd{})
		return D{}
	}
	rows := u.componentEvents(gtx)
	p, err := u.m.plan(ViewComponents)
	canInstall := err == nil && p != nil && !p.Empty() && !u.busy
	if s.install.Clicked(gtx) && canInstall {
		u.execute(p, false)
		gtx.Execute(op.InvalidateCmd{})
	}
	txt, sub := summary(p, false)
	if p != nil && !p.Empty() { // no "+": nothing is there yet to compare with
		sub = "Download " + size(p.DownloadBytes) + "  ·  Disk space " + size(p.DiskDelta)
	}
	col := colInk
	if err != nil {
		txt, sub, col = sentence(err.Error()), "", colAccent
	}
	return vcol(gtx,
		layout.Flexed(1, func(gtx C) D { return u.componentsView(gtx, rows, p) }),
		layout.Rigid(hline),
		layout.Rigid(func(gtx C) D {
			return u.footerBar(gtx, txt, sub, col,
				btn(u.th, &s.back, "Back", false, !u.busy), gap(8), btn(u.th, &s.install, "Install", true, canInstall))
		}))
}

// ---- Finishing ----

// installed runs after a successful Execute; from here on the install is
// kept, whatever else fails. It writes fpm.cfg, installs the Visual C++
// runtime if needed and creates the shortcuts, then shows the finish page.
func (u *ui) installed(res *engine.Result) {
	s := u.inst
	// A close asked for during the run waited for it; the run is over, and
	// quitting now would skip fpm.cfg, the runtime and the shortcuts. u.run
	// must go too: while it is set, closing the window is refused (loop).
	s.installed, u.run, u.closeAfterRun = true, nil, false
	if len(res.Skipped) > 0 {
		s.notes = append(s.notes, "Not installed because the download failed: "+strings.Join(u.m.titles(res.Skipped), ", ")+
			". You can add them later in the Flashpoint Manager.")
	}
	// fpm.cfg is small and local: no worker needed.
	if err := u.opts.SaveSettings(u.root, u.source); err != nil {
		u.logErr("could not write fpm.cfg", err)
		s.notes = append(s.notes, `Manager\fpm.cfg could not be written: `+plain(err, "")+" The manager finds the Flashpoint folder without it.")
	}
	if s.vcMissing {
		u.installVCRedist()
	} else {
		u.createShortcuts()
	}
}

func (u *ui) installVCRedist() {
	s := u.inst
	u.screen, u.msg = scrLoading, "Installing the Visual C++ runtime…"
	u.goWork(func(ctx context.Context) func() {
		reboot, err := s.opts.InstallVCRedist(ctx)
		return func() {
			if err == nil {
				if reboot {
					s.notes = append(s.notes, "Restart Windows before you start Flashpoint.")
				}
				u.createShortcuts()
				return
			}
			u.logErr("could not install the Visual C++ runtime", err)
			msg := plain(err, "")
			if offline(err) { // it comes from Microsoft, not from the Flashpoint server
				msg = "Couldn't connect to the Microsoft download server. Check your internet connection and try again."
			}
			u.modal = newModal("The Visual C++ runtime could not be installed",
				[]string{msg, "The Flashpoint launcher needs it. You can install it later from this address:", vcRedistLink},
				[]string{"Try again", "Continue without"}, func(i int) {
					if i == 0 {
						u.installVCRedist()
						return
					}
					s.notes = append(s.notes, "The Visual C++ runtime is not installed. The Flashpoint launcher needs it: install it from "+vcRedistLink+".")
					u.createShortcuts()
				})
			u.modal.enter, u.modal.escape = 0, 1
		}
	})
}

func (u *ui) createShortcuts() {
	s := u.inst
	u.screen, u.msg = scrLoading, "Creating shortcuts…"
	u.goWork(func(context.Context) func() {
		err := s.opts.CreateShortcuts(s.dest, s.desktop, s.startMenu)
		return func() {
			if err != nil {
				u.logErr("could not create the shortcuts", err)
				s.notes = append(s.notes, "Some shortcuts could not be created ("+strings.TrimSuffix(plain(err, ""), ".")+").")
			}
			u.screen = scrFinish
		}
	})
}

// ---- Finish ----

func (u *ui) finishView(gtx C) D {
	s := u.inst
	if s.close.Clicked(gtx) {
		u.quit = true // closeInstall starts the launcher
	}
	var rows []layout.FlexChild
	for _, n := range s.notes {
		rows = append(rows, layout.Rigid(func(gtx C) D { return layout.Inset{Bottom: 8}.Layout(gtx, para(u.th, n, colAmber).Layout) }))
	}
	if len(rows) > 0 {
		rows = append(rows, gap(8))
	}
	rows = append(rows, layout.Rigid(func(gtx C) D { return u.option(gtx, &s.launchB, &s.launch, "Start Flashpoint") }))
	return u.message(gtx, "Flashpoint is installed", "It is in "+s.dest+".", rows,
		btn(u.th, &s.close, "Close", true, true))
}

// option is a checkbox with a label; a click on either toggles *on.
func (u *ui) option(gtx C, c *widget.Clickable, on *bool, label string) D {
	if c.Clicked(gtx) {
		*on = !*on
	}
	st := unchecked
	if *on {
		st = checked
	}
	return c.Layout(gtx, func(gtx C) D {
		return focused(gtx, c, func(gtx C) D {
			return hrow(gtx, layout.Rigid(func(gtx C) D { return checkIcon(gtx, st, true) }), gap(8),
				layout.Rigid(lbl(u.th, 14, label, colInk).Layout))
		})
	})
}

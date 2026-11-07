// Package gui is the Flashpoint Manager window (DESIGN.md §6), built with Gio.
//
// All UI state lives in ui and model and is only touched on the window's
// event goroutine. Long engine calls run on one worker goroutine at a time
// and hand their result back as a closure through ui.inbox, which the event
// loop runs before the next frame.
package gui

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"gioui.org/app"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"
	"gioui.org/x/explorer"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/engine"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/fetch"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/plan"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
)

// View is a page of the main window.
type View int

const (
	ViewComponents View = iota
	ViewUpdates
	ViewSettings
)

// Options are provided by cmd/manager.
type Options struct {
	Version   string
	StartView View
	// DownloadIDs non-empty is download mode (launcher on-demand installs):
	// a small window installs these IDs (already split and lowercased) with
	// their dependencies, shows the progress and closes when done.
	DownloadIDs []string
	// Open opens the install; "" uses fpm.cfg. Called at startup and after
	// the user picked a folder.
	Open func(rootOverride string) (*engine.Engine, error)
	// SaveSettings persists the Flashpoint folder and component source.
	SaveSettings func(root, source string) error
	// OpenLauncher starts Launcher/Flashpoint.exe; the manager then closes.
	OpenLauncher func() error
	// Restart relaunches the manager (after core-manager was updated).
	Restart  func() error
	OpenLogs func() error
	// CanUninstall says why root can't be uninstalled (nil: it can).
	// Uninstall deletes Flashpoint from root (the engine is closed first)
	// and returns what it kept because it isn't Flashpoint's; afterwards
	// the window can only be closed.
	CanUninstall func(root string) error
	Uninstall    func(root string) (kept []string, err error)
	// Log receives a window failure; nil discards it.
	Log *slog.Logger
	// Exit is called with the process exit code when the window has closed
	// (download mode: 0 only if every requested ID ended up installed). It
	// must end the process.
	Exit func(code int)
	// Install non-nil runs the installer wizard (FlashpointInstaller.exe).
	// Open is then called with the destination folder (it creates it),
	// SaveSettings writes the new install's fpm.cfg, and OpenLauncher starts
	// the launcher of the last opened folder.
	Install *InstallOptions
}

// Run opens the window and never returns: Gio must own the main thread on
// macOS, so the window runs on another goroutine and opts.Exit ends the
// process.
func Run(opts Options) {
	go func() {
		w := new(app.Window)
		// 900x600 fits a 1280x720 screen at 150 %.
		width, height, minW, minH := unit.Dp(900), unit.Dp(600), unit.Dp(760), unit.Dp(500)
		if len(opts.DownloadIDs) > 0 {
			width, height, minW, minH = 600, 420, 480, 320
		}
		title := "Flashpoint Manager "
		if opts.Install != nil {
			title = "Flashpoint Installer "
		}
		w.Option(app.Title(strings.TrimSpace(title+opts.Version)), app.Size(width, height), app.MinSize(minW, minH))
		opts.Exit(newUI(opts, w).loop())
	}()
	app.Main()
}

// loadTimeout is the network budget for loading the component list; after
// it the last downloaded copy is used.
const loadTimeout = 15 * time.Second

type screen int

const (
	scrLoading     screen = iota // spinner with ui.msg
	scrFatal                     // error with Retry / Quit
	scrLocate                    // Flashpoint folder not found
	scrMain                      // rail and views
	scrRun                       // progress, then the result
	scrUninstalled               // the result of Uninstall; only Close is left
	scrDest                      // installer: destination folder
	scrFinish                    // installer: done
)

type ui struct {
	opts Options
	w    *app.Window // nil in tests
	th   *material.Theme
	logo paint.ImageOp
	expl *explorer.Explorer

	inbox  chan func()
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	busy   bool // a worker is using the engine

	screen     screen
	view       View
	msg        string // scrLoading
	fatalTitle string // scrFatal
	fatalMsg   string

	eng                   *engine.Engine
	root, source, channel string
	readOnly              string // why changes are disabled, "" if they are not
	m                     *model
	notes                 []string // shown with the next snapshot's warnings
	warnings              []string
	run                   *run
	modal                 *modal
	dlUnknown             []string // download mode: IDs not in the component list
	uninstallKept         []string // scrUninstalled: what Uninstall kept
	uninstallErr          error    // scrUninstalled: what could not be deleted

	closeAfterRun, quit bool
	exitCode            int

	search                          widget.Editor
	tree, details, updList, setList widget.List
	rail                            [3]widget.Clickable
	apply, launcher, retry, quitBtn widget.Clickable
	dismiss, done, restart, cancelB widget.Clickable
	selectAll                       widget.Clickable
	locate                          locateState
	set                             settingsState
	inst                            *installState // installer mode, else nil
}

func newUI(opts Options, w *app.Window) *ui {
	u := &ui{opts: opts, w: w, th: newTheme(), logo: loadLogo(), inbox: make(chan func(), 64), view: opts.StartView}
	u.ctx, u.cancel = context.WithCancel(context.Background())
	if w != nil {
		u.expl = explorer.NewExplorer(w)
	}
	for _, ed := range []*widget.Editor{&u.search, &u.locate.path, &u.set.folder, &u.set.custom} {
		ed.SingleLine = true
	}
	u.locate.path.Submit = true
	for _, l := range []*widget.List{&u.tree, &u.details, &u.updList, &u.setList} {
		l.Axis = layout.Vertical
	}
	if u.download() {
		u.exitCode = 1 // until every requested ID is installed
	}
	if opts.Install != nil {
		u.inst = newInstallState(opts.Install)
		u.screen = scrDest
	}
	return u
}

func (u *ui) download() bool { return len(u.opts.DownloadIDs) > 0 }

func (u *ui) loop() int {
	if u.inst == nil { // the installer starts on its destination page
		u.start("")
	}
	var ops op.Ops
	for {
		e := u.w.Event()
		u.expl.ListenEvents(e)
		switch e := e.(type) {
		case *app.ClosingEvent:
			// Never close in the middle of a change: cancel it (once the
			// commit has started that waits for it) and close afterwards.
			if u.run != nil && !u.run.finished {
				e.Abort()
				u.closeAfterRun = true
				u.run.requestCancel()
				u.w.Invalidate()
			}
		case app.DestroyEvent:
			u.shutdown()
			if e.Err != nil { // e.g. the window could not be created
				if u.opts.Log != nil {
					u.opts.Log.Error("the window failed", "err", e.Err)
				}
				return 1
			}
			return u.exitCode
		case app.FrameEvent:
			u.drain()
			gtx := app.NewContext(&ops, e)
			u.layout(gtx)
			e.Frame(gtx.Ops)
			if u.quit {
				u.quit = false
				u.w.Perform(system.ActionClose)
			}
		}
	}
}

func (u *ui) drain() {
	for {
		select {
		case f := <-u.inbox:
			f()
		default:
			return
		}
	}
}

// post runs f on the UI goroutine before the next frame.
func (u *ui) post(f func()) {
	u.inbox <- f
	u.invalidate()
}

func (u *ui) invalidate() {
	if u.w != nil {
		u.w.Invalidate()
	}
}

// goWork runs f on a worker; the closure it returns runs on the UI goroutine.
func (u *ui) goWork(f func(ctx context.Context) func()) {
	u.busy = true
	u.wg.Add(1)
	go func() {
		defer u.wg.Done()
		done := f(u.ctx)
		u.post(func() {
			u.busy = false
			if done != nil {
				done()
			}
		})
	}()
}

// shutdown cancels workers and waits for them (a commit in progress
// finishes or rolls back), then releases the install lock and finishes the
// installer (closeInstall).
func (u *ui) shutdown() {
	u.cancel()
	idle := make(chan struct{})
	go func() { u.wg.Wait(); close(idle) }()
	for {
		select {
		case <-u.inbox: // the window is gone; results no longer matter
		case <-idle:
			u.closeEngine()
			u.closeInstall()
			return
		}
	}
}

func (u *ui) closeEngine() {
	if u.eng != nil {
		u.eng.Close()
		u.eng = nil
	}
}

// fail returns the closure that shows a fatal error screen.
func (u *ui) fail(title string, err error) func() {
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return func() {
		u.logErr(title, err)
		u.screen, u.fatalTitle, u.fatalMsg, u.exitCode = scrFatal, title, plain(err, ""), 1
	}
}

// logErr keeps the technical text of an error the user sees in plain words.
func (u *ui) logErr(msg string, err error) {
	if u.opts.Log != nil && err != nil {
		u.opts.Log.Warn(msg, "err", err)
	}
}

// start opens the install (rootOverride "" = fpm.cfg), takes the lock and
// loads the component list.
func (u *ui) start(rootOverride string) {
	u.closeEngine()
	u.m, u.run, u.modal = nil, nil, nil
	u.screen, u.msg = scrLoading, "Opening the Flashpoint folder…"
	u.goWork(func(ctx context.Context) func() {
		if rootOverride != "" && !state.IsRoot(rootOverride) {
			return func() {
				u.screen = scrLocate
				u.locate.err = notRoot(rootOverride)
			}
		}
		eng, err := u.opts.Open(rootOverride)
		if err != nil {
			if errors.Is(err, engine.ErrRootNotFound) && !u.download() {
				return func() { u.screen = scrLocate }
			}
			return u.fail("Could not open the Flashpoint folder", err)
		}
		var notes []string
		if rootOverride != "" {
			if err := u.opts.SaveSettings(eng.Root(), eng.Source()); err != nil {
				notes = append(notes, "The folder could not be saved in fpm.cfg, so you will be asked again next time: "+plain(err, ""))
			}
		}
		recovered, err := eng.Lock()
		if err != nil {
			eng.Close()
			if errors.Is(err, engine.ErrBusy) {
				return u.fail("Flashpoint Manager is already running", err)
			}
			return u.fail("Could not open the Flashpoint folder", err)
		}
		if recovered {
			notes = append(notes, "An update that was interrupted earlier has been completed or undone.")
		}
		return func() {
			u.eng, u.root, u.source, u.channel = eng, eng.Root(), eng.Source(), eng.Channel()
			u.notes = append(u.notes, notes...)
			u.reload()
		}
	})
}

// reload reads the component list and records (at startup and after a
// change).
func (u *ui) reload() {
	u.run, u.modal = nil, nil
	u.screen, u.msg = scrLoading, "Loading component list…"
	eng := u.eng
	u.goWork(func(ctx context.Context) func() {
		snap, readOnly, err := load(ctx, eng)
		if err != nil {
			return u.fail("Could not load the component list", err)
		}
		return func() { u.readOnly = readOnly; u.setSnapshot(snap) }
	})
}

// load prefers a manifest that is good enough for changes; if there is
// none, a read-only one lets the user at least browse. Both share one
// network budget: once it is used up, LoadManifest goes straight to the
// saved copies.
func load(ctx context.Context, eng *engine.Engine) (*engine.Snapshot, string, error) {
	lctx, cancel := context.WithTimeout(ctx, loadTimeout)
	defer cancel()
	snap, err := eng.LoadManifest(lctx, engine.ForChanges)
	if err == nil || ctx.Err() != nil {
		return snap, "", err
	}
	if snap, err2 := eng.LoadManifest(lctx, engine.ForReading); err2 == nil {
		return snap, "Changes are disabled: " + plain(err, "The component list"), nil
	}
	return nil, "", err
}

func (u *ui) setSnapshot(snap *engine.Snapshot) {
	u.m = newModel(snap)
	u.warnings = u.notes
	u.notes = nil
	if snap.Origin != engine.FromNetwork {
		u.warnings = append(u.warnings, "Couldn't reach the Flashpoint server; showing the last downloaded list.")
	}
	for _, w := range snap.Warnings {
		if !strings.HasPrefix(w, "could not download the component list") { // replaced by the line above
			u.warnings = append(u.warnings, sentence(clean(w)))
		}
	}
	if u.readOnly != "" {
		u.warnings = append(u.warnings, u.readOnly)
	}
	u.set.reset(u.root, u.source, u.channel)
	u.screen, u.exitCode = scrMain, 0
	if u.download() {
		u.startDownload()
	}
}

// confirm asks before applying p (repair: p reinstalls broken components).
// Enter applies, Escape cancels.
func (u *ui) confirm(p *plan.Plan, title, ok string, repair bool) {
	u.modal = newModal(title, planLines(u.m, p, repair), []string{ok, "Cancel"}, func(i int) {
		if i == 0 {
			u.execute(p, repair)
		}
	})
	u.modal.enter, u.modal.escape = 0, 1
}

func (u *ui) execute(p *plan.Plan, repair bool) {
	r := newRun(u.m, p, repair)
	ctx, cancel := context.WithCancel(u.ctx)
	r.cancel = cancel
	u.run, u.screen = r, scrRun
	eng, hooks := u.eng, u.hooks(r)
	u.goWork(func(context.Context) func() {
		res, err := eng.Execute(ctx, p, hooks)
		cancel()
		return func() { u.finishRun(res, err) }
	})
}

func (u *ui) finishRun(res *engine.Result, err error) {
	u.modal = nil
	if err != nil {
		u.logErr("the change failed", err)
	}
	// Execute checks the space before it starts, so nothing happened yet:
	// keep the selection and say why.
	if errors.Is(err, engine.ErrDiskSpace) && !u.download() && !u.closeAfterRun {
		u.run, u.screen = nil, scrMain
		u.alert("Not enough disk space", errors.New(strings.TrimPrefix(err.Error(), engine.ErrDiskSpace.Error()+": ")))
		return
	}
	if u.inst != nil && err == nil {
		u.installed(res)
		return
	}
	u.run.finish(res, err)
	// Execute without an error or a skip applied the whole plan, which holds
	// every requested component that was not installed yet.
	if u.download() && err == nil && len(res.Skipped) == 0 && len(u.dlUnknown) == 0 {
		u.exitCode, u.quit = 0, true
	}
	if u.closeAfterRun {
		u.quit = true
	}
}

// startDownload runs download mode: install the requested IDs (and their
// dependencies) without asking.
func (u *ui) startDownload() {
	u.exitCode = 1
	var ids []string
	u.dlUnknown = nil
	for _, id := range u.opts.DownloadIDs {
		if _, ok := u.m.snap.Manifest.Resolve(id); ok {
			ids = append(ids, id)
		} else {
			u.dlUnknown = append(u.dlUnknown, id)
		}
	}
	if u.readOnly != "" {
		u.fail("Could not install components", errors.New(u.readOnly))()
		return
	}
	p, err := u.m.resolve(plan.Request{Install: ids, AutoDeps: true})
	switch {
	case err != nil:
		u.fail("Could not install components", err)()
	case len(u.dlUnknown) > 0 && p.Empty():
		u.fail("Could not install components", fmt.Errorf("unknown component %s", strings.Join(u.dlUnknown, ", ")))()
	case p.Empty(): // all installed already
		u.exitCode, u.quit = 0, true
	default:
		u.execute(p, false)
	}
}

func (u *ui) openLauncher() {
	if err := u.opts.OpenLauncher(); err != nil {
		u.alert("Could not start Flashpoint", err)
		return
	}
	u.quit = true
}

func (u *ui) restartManager() {
	u.closeEngine() // the new process needs the lock
	if err := u.opts.Restart(); err != nil {
		u.notes = append(u.notes, "Could not restart the manager: "+plain(err, "")+" Close it and start it again.")
		u.start("")
		return
	}
	u.quit = true
}

func (u *ui) alert(title string, err error) {
	u.modal = newModal(title, []string{plain(err, "")}, []string{"OK"}, nil)
	u.modal.enter, u.modal.escape = 0, 0
}

// plain is err for the user: common failures in plain words, else the
// message without internal paths. what names the download for a 404.
func plain(err error, what string) string {
	switch {
	case offline(err):
		return "Couldn't connect to the Flashpoint server. Check your internet connection and try again."
	case what != "" && errors.Is(err, fetch.ErrNotFound):
		return what + " is missing on the server."
	}
	return sentence(clean(err.Error()))
}

// offline reports whether err is a connection problem. These cover DNS,
// connect, reset, TLS and timeouts; the log has the details. Not the
// net.Error interface: syscall.Errno satisfies it, so every disk error
// would read as a connection problem.
func offline(err error) bool {
	var (
		ue *url.Error
		oe *net.OpError
		de *net.DNSError
	)
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, fetch.ErrStalled) ||
		errors.As(err, &ue) || errors.As(err, &oe) || errors.As(err, &de)
}

var reQuery = regexp.MustCompile(`\?h=[^\s:"]*`)

// clean drops the cache-busting query from an error message, and the
// remark that the result screen makes anyway.
func clean(s string) string {
	return strings.ReplaceAll(reQuery.ReplaceAllString(s, ""), " (no changes were made)", "")
}

// notRoot says that dir is not a Flashpoint folder.
func notRoot(dir string) string {
	return dir + " doesn't contain the Launcher, Data and Components folders."
}

// sentence makes an error message start with a capital letter and end with
// a full stop.
func sentence(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}
	s = strings.ToUpper(s[:1]) + s[1:]
	if !strings.HasSuffix(s, ".") {
		s += "."
	}
	return s
}

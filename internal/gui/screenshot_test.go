package gui

import (
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"gioui.org/gpu/headless"
	"gioui.org/io/input"
	"gioui.org/io/key"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/unit"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/engine"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/fetch"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/manifest"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/plan"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/txn"
)

var shots = flag.String("shots", "", "write screenshots of representative GUI states to this folder")

// fixture is testdata/manifests/small.xml with the legacy C# records, two
// of them made outdated and one obsolete record added.
func fixture(t *testing.T) *engine.Snapshot {
	t.Helper()
	f, err := os.Open("../../testdata/manifests/small.xml")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m, err := manifest.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	dir := "../../testdata/records/csharp"
	ents, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	recs := map[string]*state.Record{}
	for _, e := range ents {
		if e.Name() == state.BackupManifestName {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		if recs[e.Name()], err = state.ParseRecord(e.Name(), data); err != nil {
			t.Fatal(err)
		}
	}
	recs["core-logos"].Hash = "00000000"
	recs["core-logos"].InstallSize -= 300 << 10
	recs["theme-4chan"].Hash = "00000000"
	recs["theme-oldtest"] = &state.Record{ID: "theme-oldtest", Hash: "1234abcd", InstallSize: 120 << 10}
	return &engine.Snapshot{Manifest: m, Records: recs, Origin: engine.FromNetwork}
}

func testUI(t *testing.T, opts Options) *ui {
	u := newUI(opts, nil)
	u.opts.Version = "2.0.0"
	u.root, u.source, u.channel = `C:\Flashpoint`, state.DefaultSource, "stable"
	u.opts.OpenLauncher = func() error { return nil }
	u.opts.Restart = func() error { return nil }
	u.opts.OpenLogs = func() error { return nil }
	u.setSnapshot(fixture(t))
	return u
}

// frame lays u out twice (the second frame sees state settled by the
// first) and returns the ops.
func frame(u *ui, size image.Point) *op.Ops {
	var router input.Router
	var ops op.Ops
	for range 2 {
		ops.Reset()
		gtx := layout.Context{Ops: &ops, Now: time.Now(), Source: router.Source(),
			Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Exact(size)}
		u.layout(gtx)
		router.Frame(&ops)
	}
	return &ops
}

func TestModelRequests(t *testing.T) {
	u := testUI(t, Options{})
	m := u.m
	if p, err := m.plan(ViewComponents); err != nil || !p.Empty() {
		t.Fatalf("unchanged tree should plan nothing (required components are installed), got %+v, %v", p, err)
	}
	m.toggle(m.comps["supportpack-20203dviewer"], "")
	m.toggle(m.comps["theme-colorized"], "")
	p, err := m.plan(ViewComponents)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Install) == 0 || len(p.AddedDeps) == 0 || len(p.Remove) == 0 {
		t.Fatalf("want install with deps and a removal, got %+v", p)
	}
	if lines := planLines(m, p, false); !slices.ContainsFunc(lines, func(l string) bool { return strings.HasPrefix(l, "    15 that need Colorized: ") }) {
		t.Errorf("the 15 themes that need Colorized do not share a line:\n%s", strings.Join(lines, "\n"))
	}
	if m.toggle(m.comps["core-logos"], ""); !m.comps["core-logos"].Want {
		t.Fatal("required component could be unchecked")
	}
	if len(m.updates) != 3 || m.updatesChecked() != partial || m.updateCount() != 2 {
		t.Fatalf("updates: %d rows, state %d", len(m.updates), m.updatesChecked())
	}
	if p, err := m.plan(ViewUpdates); err != nil || len(p.Update) != 2 || len(p.Remove) != 0 {
		t.Fatalf("update plan: %+v, %v", p, err)
	}
	if rows := m.rows("secure"); len(rows) < 3 {
		t.Fatalf("search found %d rows", len(rows))
	}
}

func TestScreenshots(t *testing.T) {
	if *shots == "" {
		t.Skip("pass -shots=<dir> to render screenshots")
	}
	if err := os.MkdirAll(*shots, 0o755); err != nil {
		t.Fatal(err)
	}
	win, err := headless.NewWindow(900, 600)
	if err != nil {
		t.Skip("no GPU for headless rendering:", err)
	}
	defer win.Release()
	shot := func(name string, u *ui) {
		t.Helper()
		writeShot(t, win, name, u)
	}

	u := newUI(Options{Version: "2.0.0"}, nil)
	u.msg = "Loading component list…"
	shot("01-loading", u)
	u.screen = scrLocate
	u.locate.err = notRoot(`D:\Games`)
	shot("02-root-not-found", u)
	u.fail("Flashpoint Manager is already running", engine.ErrBusy)()
	shot("03-busy", u)

	u = testUI(t, Options{})
	shot("10-components", u)

	m := u.m
	for _, id := range []string{"supportpack", "supportpack-common", "theme"} {
		findCat(m, id).Open = true
	}
	m.toggle(m.comps["supportpack-20203dviewer"], "")
	m.toggle(m.comps["theme-colorized"], "")
	m.selected = m.comps["supportpack-20203dviewer"]
	shot("11-components-expanded", u)
	u.tree.Position.First = 25
	m.selected = findCat(m, "supportpack-common")
	shot("12-components-scrolled-category", u)
	u.tree.Position.First = 0

	u.search.SetText("net")
	m.selected = nil
	shot("13-components-search", u)
	u.search.SetText("secure")
	m.toggle(findCat(m, "supportpack-common"), "secure") // only Secure Player
	shot("14-search-category", u)
	m.toggle(findCat(m, "supportpack-common"), "secure")
	u.search.SetText("")
	m.selected = m.comps["supportpack-common-secureplayer"]
	shot("15-dependency-details", u)

	u.view = ViewUpdates
	shot("20-updates", u)
	u.updList.Position.First = 0
	for _, r := range m.updates {
		r.checked = false
	}
	shot("21-updates-none-selected", u)

	u.view = ViewSettings
	u.set.checked, u.set.total = true, 63
	u.set.broken = []engine.VerifyResult{
		{ID: "theme-4chan", Missing: []string{"Data/Themes/4chan/img/check.png", "Data/Themes/4chan/img/cross.png", "Data/Themes/4chan/theme.css", "Data/Themes/4chan/x"}},
		{ID: "theme-oldtest", Missing: []string{"Data/Themes/oldtest/theme.css"}},
	}
	shot("30-settings-check-files", u)
	u.set.source.Value = "custom"
	u.set.custom.SetText("http://example.org/components.xml")
	shot("31-settings-custom", u)
	u.set.reset(u.root, u.source, u.channel)
	u.setList.Position.First = 100 // scrolled to the end: Uninstall
	shot("32-settings-uninstall", u)
	u.setList.Position.First = 0
	u.confirmUninstall()
	shot("33-uninstall-confirm", u)
	u.modal, u.screen = nil, scrUninstalled
	shot("34-uninstalled", u)
	u.uninstallKept = []string{"My Mods", "notes.txt"}
	u.uninstallErr = errors.New(`Could not delete C:\Flashpoint\Launcher\Flashpoint.exe: Access is denied.`)
	shot("35-uninstalled-partly", u)
	u.uninstallKept, u.uninstallErr, u.screen = nil, nil, scrMain

	u.view = ViewComponents
	p, err := m.resolve(m.treeRequest())
	if err != nil {
		t.Fatal(err)
	}
	u.confirm(p, "Apply these changes?", "Apply changes", false)
	shot("40-confirm", u)
	u.modal = nil

	// Progress: a mid-download state with one row retrying.
	p, err = m.resolve(plan.Request{Install: []string{"supportpack-20203dviewer", "supportpack-3dgroovegx", "extra-s3bucketscraper"},
		Update: []string{"core-logos"}, Remove: []string{"theme-oldtest"}, AutoDeps: true})
	if err != nil {
		t.Fatal(err)
	}
	r := newRun(m, p, false)
	r.cancel = func() {}
	u.run, u.screen = r, scrRun
	at := time.Now().Add(-2 * time.Second)
	for _, row := range r.rows {
		row.markT = at
	}
	ev := func(e engine.Event) { r.event(e) }
	ev(engine.Event{Kind: engine.DownloadStarted, ID: "core-logos"})
	ev(engine.Event{Kind: engine.DownloadProgress, ID: "core-logos", Done: 2614191, Total: 2614191})
	ev(engine.Event{Kind: engine.DownloadStarted, ID: "supportpack-20203dviewer"})
	r.byID["supportpack-20203dviewer"].markT = at
	ev(engine.Event{Kind: engine.DownloadProgress, ID: "supportpack-20203dviewer", Done: 3 << 20, Total: 7372114})
	ev(engine.Event{Kind: engine.DownloadStarted, ID: "supportpack-3dgroovegx"})
	ev(engine.Event{Kind: engine.DownloadRetrying, ID: "supportpack-3dgroovegx", Attempt: 2, Wait: 8 * time.Second, Err: errors.New("read: connection reset")})
	shot("50-progress", u)

	dns := &url.Error{Op: "Get", URL: "https://nexus-dev.unstable.life/repository/stable/supportpack-3dgroovegx.zip",
		Err: &net.OpError{Op: "dial", Net: "tcp", Err: &net.DNSError{Err: "no such host", Name: "nexus-dev.unstable.life", IsNotFound: true}}}
	u.modal = newModal("Download failed", []string{"Could not download 3D Groove GX.", plain(dns, "3D Groove GX")},
		[]string{"Retry", "Skip 3D Groove GX", "Abort"}, nil)
	shot("51-error-modal", u)

	r.event(engine.Event{Kind: engine.Extracting, ID: "core-logos"})
	r.event(engine.Event{Kind: engine.Committing})
	u.modal = newModal("A file is in use", []string{`Launcher\Flashpoint.exe is in use or not writable.`,
		txn.LockedAdvice + " Then press Retry. Abort undoes the whole change."}, []string{"Retry", "Abort"}, nil)
	shot("52-locked-modal", u)
	u.modal = nil
	shot("53-committing", u)

	r.finish(&engine.Result{Installed: []string{"supportpack-20203dviewer", "supportpack-common-secureplayer"}, Updated: []string{"core-logos", "core-manager"},
		Removed: []string{"theme-oldtest"}, Skipped: []string{"supportpack-3dgroovegx"}}, nil)
	shot("54-finished", u)
	u.run = newRun(m, p, false)
	u.finishRun(&engine.Result{}, errDiskSpace)
	shot("55-disk-space", u)
	u.modal = nil
	u.run, u.screen = newRun(m, p, false), scrRun
	u.run.lockedPath = `Data\Themes\4chan\theme.css`
	u.run.finish(&engine.Result{}, fmt.Errorf("%w: cannot replace Data/Themes/4chan/theme.css: Sharing violation. (no changes were made)", txn.ErrLocked))
	shot("56-locked-abort", u)

	off := fixture(t)
	off.Origin = engine.FromBackup
	off.Warnings = []string{`could not download the component list (download component list: Get "http://127.0.0.1:8000/small/components.xml": context deadline exceeded); using the copy from the last successful check`}
	u.setSnapshot(off)
	shot("60-offline-banner", u)

	// Download mode in its small window.
	small, err := headless.NewWindow(600, 420)
	if err != nil {
		t.Fatal(err)
	}
	defer small.Release()
	win = small
	d := testUI(t, Options{})
	d.opts.DownloadIDs = []string{"supportpack-20203dviewer"}
	dp, err := d.m.resolve(plan.Request{Install: d.opts.DownloadIDs, AutoDeps: true})
	if err != nil {
		t.Fatal(err)
	}
	d.run, d.screen = newRun(d.m, dp, false), scrRun
	d.run.event(engine.Event{Kind: engine.DownloadStarted, ID: "supportpack-common-secureplayer"})
	d.run.event(engine.Event{Kind: engine.DownloadProgress, ID: "supportpack-common-secureplayer", Done: 1 << 20, Total: 4 << 20})
	shot("70-download-mode", d)

	// The minimum window size, with long badges (a cascade removal).
	tiny, err := headless.NewWindow(760, 500)
	if err != nil {
		t.Fatal(err)
	}
	defer tiny.Release()
	win = tiny
	u = testUI(t, Options{})
	findCat(u.m, "theme").Open = true
	u.m.toggle(u.m.comps["theme-colorized"], "")
	u.m.selected = u.m.comps["theme-bronze"]
	shot("80-small-window", u)

	// A required component that is not installed (a new core component).
	win, err = headless.NewWindow(900, 600)
	if err != nil {
		t.Fatal(err)
	}
	defer win.Release()
	u = testUI(t, Options{})
	snap := fixture(t)
	delete(snap.Records, "core-logos")
	u.setSnapshot(snap)
	u.view = ViewUpdates
	shot("22-updates-missing-core", u)
}

// writeShot renders u in win and saves it as <-shots>/<name>.png.
func writeShot(t *testing.T, win *headless.Window, name string, u *ui) {
	t.Helper()
	if err := win.Frame(frame(u, win.Size())); err != nil {
		t.Fatal(err)
	}
	img := image.NewRGBA(image.Rectangle{Max: win.Size()})
	if err := win.Screenshot(img); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(filepath.Join(*shots, name+".png"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatal(err)
	}
}

var errDiskSpace = fmt.Errorf("%w: this needs about 1.2GB, but only 300.0MB is free on the drive with C:\\Flashpoint. Free up some space and try again.", engine.ErrDiskSpace)

// A refusal for disk space keeps the selection and explains in an alert.
func TestDiskSpaceAlert(t *testing.T) {
	u := testUI(t, Options{})
	u.m.toggle(u.m.comps["supportpack-20203dviewer"], "")
	p, _ := u.m.plan(ViewComponents)
	u.run, u.screen = newRun(u.m, p, false), scrRun
	u.finishRun(&engine.Result{}, errDiskSpace)
	if u.screen != scrMain || u.run != nil || u.modal == nil || u.modal.title != "Not enough disk space" || !u.m.comps["supportpack-20203dviewer"].Want {
		t.Fatalf("screen %d, modal %v", u.screen, u.modal)
	}
}

// The Updates view showed a Missing required component (no record) by
// reading its record and panicked on the first frame.
func TestUpdatesMissingRequired(t *testing.T) {
	u := testUI(t, Options{})
	snap := fixture(t)
	delete(snap.Records, "core-logos")
	u.setSnapshot(snap)
	if n := u.m.comps["core-logos"]; n.Status != plan.Missing || n.Rec != nil {
		t.Fatalf("core-logos: status %v, record %v", n.Status, n.Rec)
	}
	u.view = ViewUpdates
	frame(u, image.Pt(960, 640))
	if p, err := u.m.plan(ViewUpdates); err != nil || len(p.Install) == 0 || p.Install[0].ID != "core-logos" {
		t.Fatalf("plan %+v, %v", p, err)
	}
}

// Closing the window in download mode before anything was installed must
// not report success to the launcher.
func TestDownloadExitCode(t *testing.T) {
	ids := []string{"supportpack-20203dviewer"}
	u := newUI(Options{DownloadIDs: ids}, nil)
	if u.exitCode != 1 {
		t.Fatalf("exit code %d while loading", u.exitCode)
	}
	u.m = newModel(fixture(t))
	p, err := u.m.resolve(plan.Request{Install: ids, AutoDeps: true})
	if err != nil {
		t.Fatal(err)
	}
	u.run = newRun(u.m, p, false)
	u.finishRun(&engine.Result{Skipped: ids}, nil)
	if u.exitCode != 1 || u.quit {
		t.Fatalf("after a skip: exit code %d, quit %v", u.exitCode, u.quit)
	}
	u.run = newRun(u.m, p, false)
	u.finishRun(&engine.Result{Installed: []string{"supportpack-common-secureplayer", ids[0]}}, nil)
	if u.exitCode != 0 || !u.quit {
		t.Fatalf("after success: exit code %d, quit %v", u.exitCode, u.quit)
	}
	// Already installed: nothing to do, success.
	u = testUI(t, Options{DownloadIDs: []string{"core-manager"}})
	if u.run != nil || u.exitCode != 0 || !u.quit {
		t.Fatalf("already installed: run %v, exit code %d, quit %v", u.run != nil, u.exitCode, u.quit)
	}
}

func findCat(m *model, id string) *node {
	var walk func(ns []*node) *node
	walk = func(ns []*node) *node {
		for _, n := range ns {
			if n.IsCat && n.ID == id {
				return n
			}
			if r := walk(n.Kids); r != nil {
				return r
			}
		}
		return nil
	}
	return walk(m.roots)
}

func TestSettingsSource(t *testing.T) {
	u := testUI(t, Options{})
	u.source = "https://nexus-dev.unstable.life/repository/stable/components.xml" // legacy fpm.cfg
	u.set.reset(u.root, u.source, u.channel)
	if u.set.source.Value != "stable" || u.settingsChanged() {
		t.Fatalf("legacy stable URL: radio %q, changed %v", u.set.source.Value, u.settingsChanged())
	}
	u.set.source.Value = "custom"
	if u.settingsChanged() {
		t.Fatal("an empty custom URL enables Save")
	}
	u.set.source.Value = "stable"
	u.set.folder.SetText(`"` + u.root + `"`) // Explorer's "Copy as path"
	if u.settingsChanged() {
		t.Fatal("quoted folder counts as a change")
	}
}

// A category checkbox under a search acts only on the components shown.
func TestFilteredCategoryToggle(t *testing.T) {
	m := testUI(t, Options{}).m
	cat := findCat(m, "supportpack-common")
	if got := len(shown(cat, "secure")); got != 1 {
		t.Fatalf("%d leaves shown for \"secure\"", got)
	}
	m.toggle(cat, "secure")
	for _, l := range cat.leaves(nil) {
		if want := l.Rec != nil || l.ID == "supportpack-common-secureplayer"; l.Want != want {
			t.Errorf("%s: Want %v", l.ID, l.Want)
		}
	}
	if m.state(cat, "secure") != checked || m.state(cat, "") == checked {
		t.Fatal("category state should follow the shown leaves")
	}
}

// Checkboxes show what Apply does: added dependencies checked, cascade
// removals unchecked, and clicking them does the sensible thing.
func TestEffectiveCheckboxes(t *testing.T) {
	m := testUI(t, Options{}).m
	dep, viewer := m.comps["supportpack-common-secureplayer"], m.comps["supportpack-20203dviewer"]
	m.toggle(viewer, "")
	m.plan(ViewComponents)
	if !m.on(dep) || dep.Want {
		t.Fatal("added dependency should show checked")
	}
	m.toggle(dep, "") // blocked: the viewer needs it
	if m.plan(ViewComponents); !m.on(dep) {
		t.Fatal("an added dependency could be unchecked")
	}

	colorized, bronze := m.comps["theme-colorized"], m.comps["theme-bronze"]
	m.toggle(colorized, "")
	m.plan(ViewComponents)
	if m.on(bronze) || !bronze.Want {
		t.Fatal("cascade removal should show unchecked")
	}
	m.toggle(bronze, "") // re-check: keeps Colorized too
	if p, _ := m.plan(ViewComponents); !colorized.Want || slices.ContainsFunc(p.Remove, func(r *state.Record) bool { return r.ID == "theme-colorized" }) {
		t.Fatal("re-checking a cascade removal did not undo it")
	}
}

func TestPlainErrors(t *testing.T) {
	const offline = "Couldn't connect to the Flashpoint server. Check your internet connection and try again."
	for _, c := range []struct {
		err        error
		what, want string
	}{
		{fmt.Errorf("download component list: %w", context.DeadlineExceeded), "", offline},
		{&url.Error{Op: "Get", URL: "https://x/a.zip?h=1", Err: &net.OpError{Op: "dial", Net: "tcp",
			Err: &net.DNSError{Err: "no such host", Name: "x", IsNotFound: true}}}, "X", offline},
		{fmt.Errorf("x: %w", &net.OpError{Op: "dial", Net: "tcp", Err: os.NewSyscallError("connect", syscall.ECONNREFUSED)}), "", offline},
		{&url.Error{Op: "Get", URL: "https://x/", Err: &tls.CertificateVerificationError{Err: errors.New("unknown authority")}}, "", offline},
		{fmt.Errorf("download x: %w", fetch.ErrNotFound), "Macintosh", "Macintosh is missing on the server."},
		{errors.New("get https://example.org/a.zip?h=1234abcd: connection refused"), "", "Get https://example.org/a.zip: connection refused."},
		{errors.New(`file is in use: cannot replace Data/a.css: Sharing violation. (no changes were made)`), "",
			"File is in use: cannot replace Data/a.css: Sharing violation."},
	} {
		if got := plain(c.err, c.what); got != c.want {
			t.Errorf("plain(%v) = %q, want %q", c.err, got, c.want)
		}
	}
	// syscall.Errno satisfies net.Error; a disk error is no connection problem.
	if got := plain(fmt.Errorf("check Data/a.css: %w (no changes were made)", syscall.Errno(32)), ""); got == offline || !strings.HasPrefix(got, "Check Data/a.css: ") {
		t.Errorf("plain(Errno) = %q", got)
	}
	if s, _ := summary(&plan.Plan{Remove: []*state.Record{{ID: "a"}}, DiskDelta: -5}, true); s != "Remove 1 obsolete" {
		t.Errorf("summary %q", s)
	}
}

// keyPresser returns press, which lays u out, presses a key and lays u out
// again (the focus carries over to the next press). It reports whether
// the search field has the focus.
func keyPresser(u *ui) (press func(key.Event) (searchFocused bool)) {
	var router input.Router
	var ops op.Ops
	return func(e key.Event) bool {
		for i := range 2 {
			ops.Reset()
			gtx := layout.Context{Ops: &ops, Now: time.Now(), Source: router.Source(),
				Metric: unit.Metric{PxPerDp: 1, PxPerSp: 1}, Constraints: layout.Exact(image.Pt(900, 600))}
			u.layout(gtx)
			router.Frame(&ops)
			if i == 0 {
				router.Queue(e)
			}
		}
		return router.Source().Focused(&u.search)
	}
}

// Escape picks the modal's safe choice (or nothing), Enter its default.
func TestModalKeys(t *testing.T) {
	u := testUI(t, Options{})
	keys := keyPresser(u)
	press := func(name key.Name) int {
		picked := -2
		u.modal = newModal("t", nil, []string{"OK", "Cancel"}, func(i int) { picked = i })
		u.modal.enter, u.modal.escape = 0, 1
		keys(key.Event{Name: name})
		return picked
	}
	if got := press(key.NameEscape); got != 1 {
		t.Errorf("Escape picked %d", got)
	}
	if got := press(key.NameReturn); got != 0 {
		t.Errorf("Enter picked %d", got)
	}
}

// Space toggles the selected component, Ctrl+F goes to the search.
func TestTreeKeys(t *testing.T) {
	u := testUI(t, Options{})
	n := u.m.comps["supportpack-20203dviewer"]
	for c := n.Parent; c != nil; c = c.Parent {
		c.Open = true
	}
	u.m.selected = n
	press := keyPresser(u)
	press(key.Event{Name: key.NameSpace})
	if !n.Want {
		t.Fatal("Space did not check the selected component")
	}
	if !press(key.Event{Name: "F", Modifiers: key.ModShortcut}) {
		t.Fatal("Ctrl+F did not focus the search")
	}
	if press(key.Event{Name: key.NameSpace}); !n.Want {
		t.Fatal("Space while typing a search toggled the component")
	}
}

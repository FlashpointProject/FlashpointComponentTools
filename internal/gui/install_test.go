package gui

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"gioui.org/gpu/headless"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/engine"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/fetch"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/install"
)

// installFakes records the calls of fake InstallOptions callbacks; Cleanup
// is the real one, so that a lock file still open fails it on Windows. The
// callbacks run on workers, whose results reach the test through the inbox
// (or shutdown's wait), so calls needs no mutex.
type installFakes struct {
	calls []string
}

func (f *installFakes) rec(format string, args ...any) {
	f.calls = append(f.calls, fmt.Sprintf(format, args...))
}

func (f *installFakes) count(prefix string) int {
	n := 0
	for _, c := range f.calls {
		if strings.HasPrefix(c, prefix) {
			n++
		}
	}
	return n
}

func (f *installFakes) has(call string) bool {
	return slices.Contains(f.calls, call)
}

// options are installer Options on a real engine with source as the
// component list; the Visual C++ runtime is missing and fails to install.
func (f *installFakes) options(dest, source string, client *http.Client) Options {
	return Options{
		Install: &InstallOptions{
			DefaultDest: dest,
			CheckDest: func(p string) (string, string, error) {
				if p == "" {
					return "", "", errors.New("Choose a folder for Flashpoint.")
				}
				create := "" // the outermost missing folder
				for d := p; ; d = filepath.Dir(d) {
					if _, err := os.Stat(d); err == nil {
						return p, create, nil
					}
					create = d
				}
			},
			Cleanup: func(dir, create string) error {
				f.rec("cleanup %s %q", dir, create)
				return install.Cleanup(dir, create)
			},
			NeedsVCRedist: func() bool { return true },
			InstallVCRedist: func(context.Context) (bool, error) {
				f.rec("vcredist")
				return false, errors.New("the Visual C++ installer exited with code 1603")
			},
			CreateShortcuts: func(root string, desktop, startMenu bool) error {
				f.rec("shortcuts %s %v %v", root, desktop, startMenu)
				return nil
			},
		},
		Open: func(dest string) (*engine.Engine, error) {
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return nil, err
			}
			return engine.Open(context.Background(), engine.Options{ManagerDir: filepath.Join(dest, "Manager"), RootOverride: dest,
				SourceOverride: source, Fetch: fetch.Options{RetryBudget: time.Millisecond, BaseBackoff: time.Millisecond, HTTPClient: client}})
		},
		SaveSettings: func(root, source string) error {
			return engine.SaveSettings(filepath.Join(root, "Manager"), root, source)
		},
		OpenLauncher: func() error { f.rec("launcher"); return nil },
	}
}

// installUI is newUI with step (runs posted closures until cond holds) and
// pick (answers the modal), shut down at the end of the test.
func installUI(t *testing.T, opts Options) (u *ui, step func(string, func() bool), pick func(int)) {
	u = newUI(opts, nil)
	t.Cleanup(u.shutdown) // before the temp folders are removed
	step = func(what string, cond func() bool) {
		t.Helper()
		for !cond() {
			select {
			case f := <-u.inbox:
				f()
			case <-time.After(10 * time.Second):
				t.Fatalf("timed out waiting for %s (screen %d, error %q)", what, u.screen, u.inst.err)
			}
		}
	}
	pick = func(i int) {
		md := u.modal
		u.modal = nil
		md.onPick(i)
	}
	return u, step, pick
}

// TestInstallFlow drives the installer (no window) against a real engine:
// the destination page, Back, the install, the Visual C++ failure, the
// finish page, and the cleanup of abandoned installs.
func TestInstallFlow(t *testing.T) {
	var zb bytes.Buffer
	zw := zip.NewWriter(&zb)
	w, _ := zw.Create("fpm.exe")
	w.Write([]byte("exe"))
	zw.Close()
	zipCore := zb.Bytes()

	mux := http.NewServeMux()
	srv := httptest.NewTLSServer(mux) // SaveSettings only saves https sources
	defer srv.Close()
	comp := func(id, path string, data []byte) string {
		return fmt.Sprintf(`<component id="%s" title="Comp %s" description="d" download-size="%d" install-size="3" path="%s" hash="%08x" />`,
			id, id, len(data), path, crc32.ChecksumIEEE(data))
	}
	xml := `<list url="` + srv.URL + `/">` +
		`<category id="core" title="Core" description="d">` + comp("a", "Manager", zipCore) + `</category>` +
		`<category id="extra" title="Extras" description="d">` + comp("b", "Data", []byte("not served")) + `</category></list>`
	mux.HandleFunc("/components.xml", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte(xml)) })
	mux.HandleFunc("/core-a.zip", func(w http.ResponseWriter, _ *http.Request) { w.Write(zipCore) })
	source := srv.URL + "/components.xml"

	f := &installFakes{}
	dest := filepath.Join(t.TempDir(), "Flashpoint")
	u, step, pick := installUI(t, f.options(dest, source, srv.Client()))
	s := u.inst
	if u.screen != scrDest || s.path.Text() != dest || !s.vcMissing || !s.desktop || !s.startMenu {
		t.Fatalf("start: screen %d, path %q, runtime %v", u.screen, s.path.Text(), s.vcMissing)
	}

	// A destination problem is shown under the field.
	s.path.SetText("")
	u.openDest()
	step("the destination error", func() bool { return !u.busy })
	if u.screen != scrDest || s.err != "Choose a folder for Flashpoint." {
		t.Fatalf("empty path: screen %d, error %q", u.screen, s.err)
	}

	next := func() {
		t.Helper()
		u.openDest()
		step("the components", func() bool { return !u.busy })
		if u.screen != scrMain {
			t.Fatalf("after Next: screen %d, error %q", u.screen, s.err)
		}
	}
	s.path.SetText(dest)
	next()
	core, extra := u.m.comps["core-a"], u.m.comps["extra-b"]
	if !core.Want || extra.Want {
		t.Fatalf("defaults: core %v, extra %v", core.Want, extra.Want)
	}
	p, _ := u.m.plan(ViewComponents)
	if b := rowBadges(core, p, true); slices.ContainsFunc(b, func(b badgeSpec) bool { return b.text == "Will be installed" }) {
		t.Errorf("badges %v", b)
	}
	if got := u.rowSize(findCat(u.m, "extra")); got != "0 / 1" {
		t.Errorf("category count %q", got)
	}

	// Back removes the new folder (the engine is closed first, or the lock
	// file could not be deleted) and keeps the path.
	u.backToDest()
	if u.screen != scrDest || s.err != "" || !f.has("cleanup "+dest+" "+strconv.Quote(dest)) || s.path.Text() != dest {
		t.Fatalf("Back: screen %d, error %q, calls %v", u.screen, s.err, f.calls)
	}
	if _, err := os.Stat(dest); !os.IsNotExist(err) {
		t.Fatalf("Back left %s: %v", dest, err)
	}

	next()
	s.desktop = false
	p, err := u.m.plan(ViewComponents)
	if err != nil || len(p.Install) != 1 {
		t.Fatalf("plan %+v, %v", p, err)
	}
	u.execute(p, false)
	// The window's X during the commit waits for it; after a success the
	// install still finishes.
	u.closeAfterRun = true
	step("the Visual C++ error", func() bool { return u.modal != nil })
	if u.modal.title != "The Visual C++ runtime could not be installed" {
		t.Fatalf("modal %q", u.modal.title)
	}
	pick(0) // Try again
	step("the second Visual C++ error", func() bool { return u.modal != nil })
	pick(1) // Continue without
	step("the finish page", func() bool { return u.screen == scrFinish })
	if u.quit {
		t.Error("the window closes after the install")
	}

	if n := f.count("vcredist"); n != 2 {
		t.Errorf("InstallVCRedist called %d times", n)
	}
	if !f.has("shortcuts " + dest + " false true") {
		t.Errorf("calls %v", f.calls)
	}
	if !slices.ContainsFunc(s.notes, func(n string) bool { return strings.Contains(n, "Visual C++ runtime is not installed") }) {
		t.Errorf("notes %q", s.notes)
	}
	cfg, err := os.ReadFile(filepath.Join(dest, "Manager", "fpm.cfg"))
	if want := "..\\\r\n" + source + "\r\n"; err != nil || string(cfg) != want {
		t.Errorf("fpm.cfg %q, %v; notes %q", cfg, err, s.notes)
	}

	// Closing the finish page starts the launcher and keeps the install.
	u.shutdown()
	if f.count("cleanup") != 1 || !f.has("launcher") {
		t.Fatalf("after closing: calls %v", f.calls)
	}
	if _, err := os.Stat(filepath.Join(dest, "Manager", "fpm.exe")); err != nil {
		t.Fatal(err)
	}

	// Closing on the components page removes the contents of a folder that
	// existed, but keeps the folder.
	dest2 := filepath.Join(t.TempDir(), "Empty")
	os.Mkdir(dest2, 0o755)
	v, step2, _ := installUI(t, f.options(dest2, source, srv.Client()))
	v.inst.path.SetText(dest2)
	v.openDest()
	step2("the components", func() bool { return !v.busy })
	if v.screen != scrMain {
		t.Fatalf("second install: screen %d, error %q", v.screen, v.inst.err)
	}
	v.shutdown()
	ents, err := os.ReadDir(dest2)
	if !f.has("cleanup "+dest2+` ""`) || err != nil || len(ents) != 0 {
		t.Fatalf("closing before the install: calls %v, entries %v, %v", f.calls, ents, err)
	}

	// A component list that can't be loaded: the error under the field,
	// and the new folder removed.
	dest3 := filepath.Join(t.TempDir(), "Flashpoint")
	x, step3, _ := installUI(t, f.options(dest3, srv.URL+"/missing.xml", srv.Client()))
	x.inst.path.SetText(dest3)
	x.openDest()
	step3("the error", func() bool { return !x.busy })
	if x.screen != scrDest || !strings.HasPrefix(x.inst.err, "Could not load the component list.") || !f.has("cleanup "+dest3+" "+strconv.Quote(dest3)) {
		t.Fatalf("load error: screen %d, error %q, calls %v", x.screen, x.inst.err, f.calls)
	}
	if _, err := os.Stat(dest3); !os.IsNotExist(err) {
		t.Fatalf("load error left %s: %v", dest3, err)
	}

	// A destination whose parent doesn't exist yet: closing on the
	// components page removes the parent too.
	tmp := t.TempDir()
	games := filepath.Join(tmp, "Games")
	dest4 := filepath.Join(games, "Flashpoint")
	y, step4, _ := installUI(t, f.options(dest4, source, srv.Client()))
	y.openDest()
	step4("the components", func() bool { return !y.busy })
	if y.screen != scrMain {
		t.Fatalf("nested install: screen %d, error %q", y.screen, y.inst.err)
	}
	y.shutdown()
	if _, err := os.Stat(games); !f.has("cleanup "+dest4+" "+strconv.Quote(games)) || !os.IsNotExist(err) {
		t.Fatalf("closing a nested install: calls %v, %s: %v", f.calls, games, err)
	}
	if _, err := os.Stat(tmp); err != nil {
		t.Fatal(err)
	}
}

// Browse… suggests a Flashpoint folder in a folder that isn't empty, but
// not in a Flashpoint install (CheckDest then says it is installed there).
func TestInFlashpointFolder(t *testing.T) {
	empty, games, root := t.TempDir(), t.TempDir(), t.TempDir()
	os.WriteFile(filepath.Join(games, "game.exe"), nil, 0o644)
	os.Mkdir(filepath.Join(root, "Launcher"), 0o755)
	os.Mkdir(filepath.Join(root, "Data"), 0o755)
	for dir, want := range map[string]string{empty: empty, games: filepath.Join(games, "Flashpoint"), root: root} {
		if got := inFlashpointFolder(dir); got != want {
			t.Errorf("inFlashpointFolder(%s) = %s, want %s", dir, got, want)
		}
	}
}

func TestScreenshotsInstaller(t *testing.T) {
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

	// The small manifest with nothing installed.
	u := newUI((&installFakes{}).options(`D:\Flashpoint`, "", nil), nil)
	u.opts.Version = "2.0.0"
	u.root, u.source, u.channel = `D:\Flashpoint`, "", "stable"
	u.inst.dest = `D:\Flashpoint`
	shot("90-install-dest", u)
	u.inst.path.SetText(`C:\Users\Someone\OneDrive\Flashpoint`)
	frame(u, win.Size()) // the edit clears the error, as typing does
	u.inst.err = `Flashpoint can't be installed in OneDrive, because syncing breaks it. Choose another folder, such as C:\Flashpoint.`
	u.inst.vcMissing = false
	shot("91-install-dest-error", u)

	snap := fixture(t)
	snap.Records = nil
	u.setInstallSnapshot(snap)
	m := u.m
	findCat(m, "supportpack").Open = true
	findCat(m, "core").Open = true
	m.toggle(m.comps["supportpack-20203dviewer"], "")
	m.selected = m.comps["supportpack-20203dviewer"]
	shot("92-install-components", u)
	m.selected = nil
	shot("93-install-components-intro", u)

	p, err := m.resolve(m.treeRequest())
	if err != nil {
		t.Fatal(err)
	}
	r := newRun(m, p, false)
	r.cancel = func() {}
	u.run, u.screen = r, scrRun
	r.event(engine.Event{Kind: engine.DownloadStarted, ID: "core-logos"})
	r.event(engine.Event{Kind: engine.DownloadProgress, ID: "core-logos", Done: 1 << 20, Total: 2614191})
	shot("94-install-progress", u)
	r.finish(&engine.Result{}, context.Canceled)
	shot("95-install-cancelled", u)

	u.run, u.screen, u.msg = nil, scrLoading, "Installing the Visual C++ runtime…"
	u.modal = newModal("The Visual C++ runtime could not be installed",
		[]string{"The Windows permission prompt was declined, so the Visual C++ runtime was not installed.", "The Flashpoint launcher needs it. You can install it later from this address:", vcRedistLink},
		[]string{"Try again", "Continue without"}, nil)
	shot("96-install-vcredist-modal", u)
	u.modal = nil

	u.screen = scrFinish
	u.inst.notes = []string{
		"The Visual C++ runtime is not installed. The Flashpoint launcher needs it: install it from " + vcRedistLink + ".",
		"Some shortcuts could not be created (Start menu: Access is denied).",
	}
	shot("97-install-finish", u)
	u.inst.notes = nil
	shot("98-install-finish-clean", u)
}

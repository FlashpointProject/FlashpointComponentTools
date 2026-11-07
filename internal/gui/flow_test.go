package gui

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"hash/crc32"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/engine"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/fetch"
)

// TestFlowWithEngine drives the UI logic (no window) against a real engine:
// startup, a change where one download fails and the user picks Skip in the
// error modal, the result, and the reload.
func TestFlowWithEngine(t *testing.T) {
	var zb bytes.Buffer
	zw := zip.NewWriter(&zb)
	w, _ := zw.Create("a/theme.css")
	w.Write([]byte("body{}"))
	zw.Close()
	zipA := zb.Bytes()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()
	comp := func(id, deps string, data []byte) string {
		return fmt.Sprintf(`<component id="%s" title="Theme %s" description="d" download-size="%d" install-size="6" path="Data/Themes"%s hash="%08x" />`,
			id, id, len(data), deps, crc32.ChecksumIEEE(data))
	}
	xml := `<list url="` + srv.URL + `/"><category id="theme" title="Themes" description="d">` +
		comp("a", "", zipA) + comp("b", ` depends="theme-a"`, []byte("missing")) + `</category></list>`
	mux.HandleFunc("/components.xml", func(w http.ResponseWriter, _ *http.Request) { w.Write([]byte(xml)) })
	mux.HandleFunc("/theme-a.zip", func(w http.ResponseWriter, _ *http.Request) { w.Write(zipA) })

	root := t.TempDir()
	for _, d := range []string{"Manager", "Launcher", "Data", "Components"} {
		os.MkdirAll(filepath.Join(root, d), 0o755)
	}
	mgr := filepath.Join(root, "Manager")
	os.WriteFile(filepath.Join(mgr, "fpm.cfg"), []byte("..\\\r\n"+srv.URL+"/components.xml\r\n"), 0o644)

	u := newUI(Options{
		Open: func(r string) (*engine.Engine, error) {
			return engine.Open(context.Background(), engine.Options{ManagerDir: mgr, RootOverride: r,
				Fetch: fetch.Options{RetryBudget: time.Millisecond, BaseBackoff: time.Millisecond}})
		},
		SaveSettings: func(root, source string) error { return engine.SaveSettings(mgr, root, source) },
	}, nil)
	defer u.shutdown()
	// step runs posted closures until cond holds.
	step := func(what string, cond func() bool) {
		t.Helper()
		for !cond() {
			select {
			case f := <-u.inbox:
				f()
			case <-time.After(10 * time.Second):
				t.Fatalf("timed out waiting for %s (screen %d, fatal %q)", what, u.screen, u.fatalMsg)
			}
		}
	}

	u.start("")
	step("startup", func() bool { return !u.busy })
	if u.screen != scrMain {
		t.Fatalf("screen %d: %s %s", u.screen, u.fatalTitle, u.fatalMsg)
	}
	u.m.toggle(u.m.comps["theme-b"], "") // needs theme-a: added as a dependency
	p, err := u.m.plan(ViewComponents)
	if err != nil || len(p.Install) != 2 || len(p.AddedDeps) != 1 {
		t.Fatalf("plan %+v, err %v", p, err)
	}
	u.confirm(p, "Apply these changes?", "Apply", false)
	pick := func(i int) {
		md := u.modal
		u.modal = nil
		md.onPick(i)
	}
	pick(0)
	step("the download error", func() bool { return u.modal != nil })
	if u.modal.title != "Download failed" {
		t.Fatalf("modal %q", u.modal.title)
	}
	pick(1) // Skip
	step("the result", func() bool { return !u.busy })
	r := u.run
	if !r.finished || r.err != nil || !slices.Equal(r.res.Installed, []string{"theme-a"}) || !slices.Equal(r.res.Skipped, []string{"theme-b"}) {
		t.Fatalf("result %+v, err %v", r.res, r.err)
	}
	if title, _, _ := r.outcome(false); title != "Finished, but some components were skipped" {
		t.Fatalf("outcome %q", title)
	}
	u.reload()
	step("reload", func() bool { return !u.busy })
	if u.screen != scrMain || u.m.comps["theme-a"].Rec == nil || u.m.comps["theme-b"].Rec != nil {
		t.Fatalf("after reload: screen %d", u.screen)
	}

	// A second manager on the same folder is refused.
	v := newUI(u.opts, nil)
	defer v.shutdown()
	v.start("")
	for v.busy {
		(<-v.inbox)()
	}
	if v.screen != scrFatal || v.exitCode != 1 {
		t.Fatalf("second instance: screen %d", v.screen)
	}
}

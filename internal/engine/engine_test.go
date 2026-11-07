package engine

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"hash/crc32"
	"html"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/fetch"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/plan"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/trust"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/txn"
)

// --- fixture ---------------------------------------------------------------

type ent struct{ name, data string }

// comp is a fixture component. The ID's first segment is its category.
type comp struct {
	id, path, depends string
	ents              []ent
	meta              bool                // metadata-only: install-size 0, no zip served
	corrupt           func([]byte) []byte // applied to the zip before hashing (manifest matches it)
}

type server struct {
	t     *testing.T
	srv   *httptest.Server
	mu    sync.Mutex
	files map[string][]byte
	fail  bool // every request -> 500
	// hold, if set for a name, makes the handler send half the body and
	// then dribble the rest until the request is cancelled.
	hold     map[string]bool
	requests []string
}

func newServer(t *testing.T) *server {
	s := &server{t: t, files: map[string][]byte{}, hold: map[string]bool{}}
	s.srv = httptest.NewServer(http.HandlerFunc(s.serve))
	t.Cleanup(s.srv.Close)
	return s
}

func (s *server) serve(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	s.requests = append(s.requests, r.URL.Path)
	name := strings.TrimPrefix(r.URL.Path, "/repo/")
	body, ok := s.files[name]
	fail, hold := s.fail, s.hold[name]
	s.mu.Unlock()
	switch {
	case fail:
		http.Error(w, "down", http.StatusInternalServerError)
	case !ok:
		http.NotFound(w, r)
	case hold:
		w.Header().Set("Content-Length", fmt.Sprint(len(body)))
		// Half the body, then a trickle that keeps progress events coming.
		n := len(body) / 2
		w.Write(body[:n])
		w.(http.Flusher).Flush()
		for ; n < len(body)-1; n++ {
			select {
			case <-r.Context().Done():
				return
			case <-time.After(20 * time.Millisecond):
			}
			w.Write(body[n : n+1])
			w.(http.Flusher).Flush()
		}
		<-r.Context().Done()
	default:
		w.Header().Set("ETag", fmt.Sprintf(`"%08x"`, crc32.ChecksumIEEE(body)))
		http.ServeContent(w, r, name, time.Time{}, bytes.NewReader(body))
	}
}

func (s *server) base() string   { return s.srv.URL + "/repo/" }
func (s *server) source() string { return s.base() + "components.xml" }

func (s *server) set(name string, b []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[name] = b
}

func makeZip(t *testing.T, ents []ent) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	dirs := map[string]bool{}
	for _, e := range ents {
		// Real component zips have explicit directory entries.
		for d := filepath.ToSlash(filepath.Dir(e.name)); d != "." && !dirs[d]; d = filepath.ToSlash(filepath.Dir(d)) {
			dirs[d] = true
			if _, err := zw.Create(d + "/"); err != nil {
				t.Fatal(err)
			}
		}
		w, err := zw.CreateHeader(&zip.FileHeader{Name: e.name, Method: zip.Store})
		if err != nil {
			t.Fatal(err)
		}
		w.Write([]byte(e.data))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

// publish builds components.xml for comps and serves it with the zips.
func (s *server) publish(comps []comp) {
	t := s.t
	var cats []string
	byCat := map[string][]string{}
	for _, c := range comps {
		cat, local, _ := strings.Cut(c.id, "-")
		if _, ok := byCat[cat]; !ok {
			cats = append(cats, cat)
		}
		var z []byte
		var install int64
		for _, e := range c.ents {
			install += int64(len(e.data))
		}
		dl := int64(160)
		hash := "00000000"
		if c.meta {
			install = 0
		} else {
			z = makeZip(t, c.ents)
			if c.corrupt != nil {
				z = c.corrupt(z)
			}
			s.set(c.id+".zip", z)
			dl = int64(len(z))
			hash = fmt.Sprintf("%08x", crc32.ChecksumIEEE(z))
		}
		attrs := fmt.Sprintf(`id=%q title=%q description="d" date-modified="1700000000" download-size="%d" install-size="%d" hash=%q`,
			local, "Title of "+c.id, dl, install, hash)
		if c.path != "" {
			attrs += fmt.Sprintf(` path=%q`, html.EscapeString(c.path))
		}
		if c.depends != "" {
			attrs += fmt.Sprintf(` depends=%q`, c.depends)
		}
		byCat[cat] = append(byCat[cat], "<component "+attrs+" />")
	}
	var b strings.Builder
	fmt.Fprintf(&b, "<?xml version='1.0' encoding='utf-8'?>\n<list url=%q>\n", s.base())
	for _, cat := range cats {
		fmt.Fprintf(&b, "<category id=%q title=%q description=\"d\">\n%s\n</category>\n", cat, cat, strings.Join(byCat[cat], "\n"))
	}
	b.WriteString("</list>\n")
	s.set("components.xml", []byte(b.String()))
}

// newRoot creates an empty Flashpoint folder whose fpm.cfg points at src.
func newRoot(t *testing.T, src string) (root, mgr string) {
	root = t.TempDir()
	for _, d := range []string{"Components", "Launcher", "Data", "Manager"} {
		os.MkdirAll(filepath.Join(root, d), 0o755)
	}
	mgr = filepath.Join(root, "Manager")
	os.WriteFile(filepath.Join(mgr, "fpm.cfg"), []byte("..\\\r\n"+src+"\r\n"), 0o644)
	return root, mgr
}

func testOptions(mgr string) Options {
	return Options{
		ManagerDir: mgr,
		Fetch: fetch.Options{BaseBackoff: time.Millisecond, MaxBackoff: 5 * time.Millisecond,
			RetryBudget: 50 * time.Millisecond, StallTimeout: 5 * time.Second},
	}
}

type opErr struct {
	id  string
	err error
}

type hooks struct {
	mu      sync.Mutex
	decide  func(id string, err error) Decision // nil = Abort
	locked  bool
	onEvent func(Event)
	errs    []opErr
}

func (h *hooks) hooks() Hooks {
	return Hooks{
		Progress: func(ev Event) {
			if h.onEvent != nil {
				h.onEvent(ev)
			}
		},
		OnError: func(_ context.Context, id string, err error) Decision {
			h.errs = append(h.errs, opErr{id, err})
			if h.decide == nil {
				return Abort
			}
			return h.decide(id, err)
		},
		OnLocked: func(context.Context, string) bool { return h.locked },
	}
}

// run opens the install, loads the manifest, plans req and executes it.
func run(t *testing.T, opts Options, req plan.Request, h *hooks) (*Result, error) {
	t.Helper()
	ctx := context.Background()
	e, err := Open(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	if _, err := e.LoadManifest(ctx, ForChanges); err != nil {
		t.Fatal(err)
	}
	if _, err := e.Lock(); err != nil {
		t.Fatal(err)
	}
	p, err := e.Plan(req)
	if err != nil {
		t.Fatal(err)
	}
	return e.Execute(ctx, p, h.hooks())
}

// tree returns every file below root (except our private and temp areas)
// with its content.
func tree(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			t.Fatal(err)
		}
		rel, _ := filepath.Rel(root, p)
		rel = filepath.ToSlash(rel)
		if rel == "Temp" || rel == "Components/.fpm" || rel == "Logs" {
			return filepath.SkipDir
		}
		if rel == "Components/components.bak" { // rewritten when the list changes
			return nil
		}
		if d.IsDir() {
			out[rel+"/"] = ""
			return nil
		}
		b, _ := os.ReadFile(p)
		out[rel] = string(b)
		return nil
	})
	return out
}

func diffTrees(t *testing.T, label string, got, want map[string]string) {
	t.Helper()
	for k, v := range want {
		if g, ok := got[k]; !ok {
			t.Errorf("%s: missing %s", label, k)
		} else if g != v {
			t.Errorf("%s: %s = %q, want %q", label, k, g, v)
		}
	}
	for k := range got {
		if _, ok := want[k]; !ok {
			t.Errorf("%s: unexpected %s", label, k)
		}
	}
}

func baseComps() []comp {
	return []comp{
		{id: "core-configuration", ents: []ent{{"Data/services.json", "{}"}, {"version.txt", "v1"}}},
		{id: "core-launcher", path: "Launcher", depends: "core-configuration",
			ents: []ent{{"Flashpoint.exe", "exe1"}, {"resources/app.asar", "asar1"}}},
		{id: "theme-base", path: "Data/Themes", ents: []ent{{"Base/root.css", "base-css"}}},
		{id: "theme-blue", path: "Data/Themes", depends: "theme-base", ents: []ent{{"Blue/theme.css", "blue"}}},
		{id: "supportpack-a", path: "FPSoftware", ents: []ent{
			{"shared/common.ini", "shared"}, {"a/a.exe", "a1"}, {"a/olddir/old.txt", "old"}}},
		{id: "supportpack-b", path: "FPSoftware", ents: []ent{{"shared/common.ini", "shared"}, {"b/b.exe", "b1"}}},
		{id: "supportpack-meta", meta: true},
		{id: "playlist-favorites", path: "Data/Playlists", ents: []ent{{"Favorites.json", "[]"}}},
	}
}

func crcOf(s *server, id string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return fmt.Sprintf("%08x", crc32.ChecksumIEEE(s.files[id+".zip"]))
}

// --- end to end ------------------------------------------------------------

func TestInstallUpdateRemove(t *testing.T) {
	s := newServer(t)
	comps := baseComps()
	s.publish(comps)
	root, mgr := newRoot(t, s.source())
	opts := testOptions(mgr)
	h := &hooks{}

	res, err := run(t, opts, plan.Request{
		Install:  []string{"core", "theme-blue", "supportpack", "playlist-favorites"},
		AutoDeps: true,
	}, h)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Installed) != 8 || len(res.Skipped) != 0 {
		t.Fatalf("result = %+v", res)
	}
	got := tree(t, root)
	for _, f := range []string{"Data/services.json", "version.txt", "Launcher/Flashpoint.exe",
		"Launcher/resources/app.asar", "Data/Themes/Base/root.css", "Data/Themes/Blue/theme.css",
		"FPSoftware/shared/common.ini", "FPSoftware/a/a.exe", "FPSoftware/b/b.exe", "Data/Playlists/Favorites.json"} {
		if _, ok := got[f]; !ok {
			t.Errorf("missing %s", f)
		}
	}
	// Records in the exact legacy byte format.
	wantRec := map[string]string{
		"theme-blue":         crcOf(s, "theme-blue") + " 4 theme-base\r\nData\\Themes\\Blue\\theme.css\r\n",
		"core-launcher":      crcOf(s, "core-launcher") + " 9 core-configuration\r\nLauncher\\Flashpoint.exe\r\nLauncher\\resources\\app.asar\r\n",
		"supportpack-meta":   "00000000 0\r\n",
		"core-configuration": crcOf(s, "core-configuration") + " 4\r\nData\\services.json\r\nversion.txt\r\n",
	}
	for id, want := range wantRec {
		if g := got["Components/"+id]; g != want {
			t.Errorf("record %s = %q, want %q", id, g, want)
		}
	}
	if b, _ := os.ReadFile(filepath.Join(root, "Components", "components.bak")); !bytes.Contains(b, []byte(s.base())) {
		t.Error("components.bak not written")
	}
	if ents, _ := os.ReadDir(filepath.Join(root, "Temp", "fpm", "cache")); len(ents) != 0 {
		t.Errorf("cache not cleaned: %v", ents)
	}

	// The user edits their favourites; then a new release changes
	// supportpack-a (new content, one file and its folder dropped) and
	// playlist-favorites.
	fav := filepath.Join(root, "Data", "Playlists", "Favorites.json")
	os.WriteFile(fav, []byte(`["my game"]`), 0o644)
	comps[4].ents = []ent{{"shared/common.ini", "shared"}, {"a/a.exe", "a2"}}
	comps[7].ents = []ent{{"Favorites.json", "[] "}}
	s.publish(comps)

	res, err = run(t, opts, plan.Request{UpdateAll: true, AutoDeps: true}, h)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(res.Updated, []string{"supportpack-a", "playlist-favorites"}) {
		t.Fatalf("updated = %v", res.Updated)
	}
	got = tree(t, root)
	if got["FPSoftware/a/a.exe"] != "a2" {
		t.Errorf("a.exe = %q", got["FPSoftware/a/a.exe"])
	}
	if _, ok := got["FPSoftware/a/olddir/"]; ok {
		t.Error("emptied folder FPSoftware/a/olddir was not pruned")
	}
	if got["Data/Playlists/Favorites.json"] != `["my game"]` {
		t.Errorf("update overwrote Favorites.json: %q", got["Data/Playlists/Favorites.json"])
	}
	if want := crcOf(s, "playlist-favorites") + " 3\r\nData\\Playlists\\Favorites.json\r\n"; got["Components/playlist-favorites"] != want {
		t.Errorf("favorites record = %q, want %q", got["Components/playlist-favorites"], want)
	}

	// Remove supportpack-a: the shared file stays (supportpack-b lists it),
	// its own folder goes.
	res, err = run(t, opts, plan.Request{Remove: []string{"supportpack-a", "playlist-favorites"}}, h)
	if err != nil {
		t.Fatal(err)
	}
	got = tree(t, root)
	if got["FPSoftware/shared/common.ini"] != "shared" {
		t.Error("shared file deleted although supportpack-b still owns it")
	}
	for _, gone := range []string{"FPSoftware/a/", "FPSoftware/a/a.exe", "Components/supportpack-a", "Components/playlist-favorites"} {
		if _, ok := got[gone]; ok {
			t.Errorf("%s still exists", gone)
		}
	}
	if got["Data/Playlists/Favorites.json"] != `["my game"]` {
		t.Error("removal deleted the user's Favorites.json")
	}

	// Quick verify.
	os.Remove(filepath.Join(root, "FPSoftware", "b", "b.exe"))
	e, _ := Open(context.Background(), opts)
	vr, err := e.Verify(context.Background(), []string{"supportpack-b", "theme-blue", "nope"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vr) != 3 || vr[1].ID != "supportpack-b" || !slices.Equal(vr[1].Missing, []string{"FPSoftware/b/b.exe"}) ||
		len(vr[2].Missing) != 0 || vr[2].Err != nil || vr[0].Err == nil {
		t.Errorf("verify = %+v", vr)
	}
}

// installed returns a fixture with the base components installed.
func installed(t *testing.T) (*server, []comp, string, Options) {
	s := newServer(t)
	comps := baseComps()
	s.publish(comps)
	root, mgr := newRoot(t, s.source())
	opts := testOptions(mgr)
	if _, err := run(t, opts, plan.Request{Install: []string{"core", "theme-base", "supportpack-a"}, AutoDeps: true}, &hooks{}); err != nil {
		t.Fatal(err)
	}
	return s, comps, root, opts
}

func TestDownloadFailure(t *testing.T) {
	for name, d := range map[string]Decision{"abort": Abort, "skip": Skip} {
		t.Run(name, func(t *testing.T) {
			s, comps, root, opts := installed(t)
			before := tree(t, root)
			// supportpack-b arrives corrupted (does not match the manifest);
			// theme-blue is independent and fine.
			s.set("supportpack-b.zip", makeZip(t, []ent{{"b/b.exe", "evil"}}))
			_ = comps
			h := &hooks{decide: func(string, error) Decision { return d }}
			res, err := run(t, opts, plan.Request{Install: []string{"supportpack-b", "theme-blue"}}, h)
			if len(h.errs) != 1 || h.errs[0].id != "supportpack-b" || !errors.Is(h.errs[0].err, fetch.ErrSizeMismatch) && !errors.Is(h.errs[0].err, fetch.ErrHashMismatch) {
				t.Fatalf("OnError calls = %+v", h.errs)
			}
			after := tree(t, root)
			if d == Abort {
				if !errors.Is(err, ErrAborted) {
					t.Fatalf("err = %v", err)
				}
				diffTrees(t, "abort", after, before)
				return
			}
			if err != nil || !slices.Equal(res.Skipped, []string{"supportpack-b"}) || !slices.Equal(res.Installed, []string{"theme-blue"}) {
				t.Fatalf("res = %+v, err = %v", res, err)
			}
			if _, ok := after["FPSoftware/b/b.exe"]; ok {
				t.Error("skipped component was installed")
			}
			delete(after, "Components/theme-blue")
			delete(after, "Data/Themes/Blue/")
			delete(after, "Data/Themes/Blue/theme.css")
			diffTrees(t, "skip", after, before)
		})
	}
}

func TestSkipDropsDependents(t *testing.T) {
	s, comps, root, opts := installed(t)
	_ = comps
	// theme-blue depends on theme-base; make a fresh install where base is broken.
	os.RemoveAll(filepath.Join(root, "Components", "theme-base"))
	os.RemoveAll(filepath.Join(root, "Data", "Themes"))
	before := tree(t, root)
	s.set("theme-base.zip", []byte("garbage"))
	h := &hooks{decide: func(string, error) Decision { return Skip }}
	res, err := run(t, opts, plan.Request{Install: []string{"theme-blue"}, AutoDeps: true}, h)
	if err != nil || !slices.Equal(res.Skipped, []string{"theme-base", "theme-blue"}) || len(res.Installed) != 0 {
		t.Fatalf("res = %+v, err = %v", res, err)
	}
	diffTrees(t, "skip deps", tree(t, root), before)
}

func TestCorruptEntry(t *testing.T) {
	s := newServer(t)
	comps := baseComps()
	// The zip matches the manifest's CRC32 but one entry's data is damaged:
	// staging catches it (zip.ErrChecksum).
	comps[2].corrupt = func(z []byte) []byte {
		z = bytes.Clone(z)
		i := bytes.Index(z, []byte("base-css"))
		z[i] ^= 0xff
		return z
	}
	s.publish(comps)
	root, mgr := newRoot(t, s.source())
	opts := testOptions(mgr)
	before := tree(t, root)
	h := &hooks{}
	_, err := run(t, opts, plan.Request{Install: []string{"core-configuration", "theme-base"}}, h)
	if !errors.Is(err, zip.ErrChecksum) || len(h.errs) != 0 {
		t.Fatalf("err = %v, OnError calls %d", err, len(h.errs))
	}
	diffTrees(t, "corrupt entry", tree(t, root), before)
	// The corrupt archive is deleted, so the next run downloads it again.
	if zips, _ := filepath.Glob(filepath.Join(root, "Temp", "fpm", "cache", "*theme-base*")); len(zips) != 0 {
		t.Errorf("corrupt archive kept: %v", zips)
	}
}

func TestWritePolicyRefused(t *testing.T) {
	evil := []comp{
		{id: "extra-escape", ents: []ent{{"../x", "x"}}},
		{id: "extra-records", ents: []ent{{"Components/core-launcher", "x"}}},
		{id: "extra-lock", path: "components/.fpm", ents: []ent{{"lock", "x"}}},
		{id: "extra-config", path: "Manager", ents: []ent{{"FPM.cfg", "x"}}},
	}
	for _, c := range evil {
		t.Run(c.id, func(t *testing.T) {
			s := newServer(t)
			s.publish(append(baseComps(), c))
			root, mgr := newRoot(t, s.source())
			before := tree(t, root)
			h := &hooks{}
			_, err := run(t, testOptions(mgr), plan.Request{Install: []string{"theme-base", c.id}}, h)
			refused := errors.Is(err, ErrPolicy) || errors.Is(err, txn.ErrUnsafeEntry) // "../x" is caught by txn
			if !refused || len(h.errs) != 0 {
				t.Fatalf("err = %v, OnError = %+v", err, h.errs)
			}
			if !strings.Contains(err.Error(), c.id) {
				t.Errorf("error does not name the component: %v", err)
			}
			diffTrees(t, c.id, tree(t, root), before)
		})
	}
	// Everything else is the server's call, e.g. a supportpack in Launcher.
	for _, rel := range []string{"Launcher/x.dll", "Manager/fpm.exe", "ComponentsX/a", "Data/Components/a", "preferences.json"} {
		if reserved(rel) {
			t.Errorf("%s is reserved", rel)
		}
	}
}

func TestCancelDuringDownload(t *testing.T) {
	s := newServer(t)
	comps := baseComps()
	comps[4].ents = append(comps[4].ents, ent{"a/big.bin", strings.Repeat("x", 256<<10)})
	s.publish(comps)
	s.hold["supportpack-a.zip"] = true
	root, mgr := newRoot(t, s.source())
	opts := testOptions(mgr)
	before := tree(t, root)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e, err := Open(ctx, opts)
	if err != nil {
		t.Fatal(err)
	}
	defer e.Close()
	e.LoadManifest(ctx, ForChanges)
	if _, err := e.Lock(); err != nil {
		t.Fatal(err)
	}
	p, _ := e.Plan(plan.Request{Install: []string{"supportpack-a"}})
	h := &hooks{onEvent: func(ev Event) {
		if ev.Kind == DownloadProgress && ev.Done > 0 {
			cancel()
		}
	}}
	_, err = e.Execute(ctx, p, h.hooks())
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v", err)
	}
	diffTrees(t, "cancel", tree(t, root), before)
	parts, _ := filepath.Glob(filepath.Join(root, "Temp", "fpm", "cache", "*.part"))
	if len(parts) != 1 {
		t.Errorf(".part files = %v, want the partial download kept", parts)
	}
}

func TestLockedFileAbort(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("open files only block renames on Windows")
	}
	s, comps, root, opts := installed(t)
	before := tree(t, root)
	comps[1].ents = []ent{{"Flashpoint.exe", "exe2"}, {"resources/app.asar", "asar2"}}
	s.publish(comps)
	f, err := os.Open(filepath.Join(root, "Launcher", "Flashpoint.exe")) // no FILE_SHARE_DELETE
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	_, err = run(t, opts, plan.Request{Update: []string{"core-launcher"}}, &hooks{})
	if !errors.Is(err, txn.ErrLocked) {
		t.Fatalf("err = %v", err)
	}
	diffTrees(t, "locked", tree(t, root), before)
}

func TestLockBusy(t *testing.T) {
	_, _, _, opts := installed(t)
	a, _ := Open(context.Background(), opts)
	defer a.Close()
	if _, err := a.Lock(); err != nil {
		t.Fatal(err)
	}
	b, _ := Open(context.Background(), opts)
	defer b.Close()
	if _, err := b.Lock(); !errors.Is(err, ErrBusy) {
		t.Fatalf("second Lock = %v, want ErrBusy", err)
	}
	a.Close()
	if _, err := b.Lock(); err != nil {
		t.Fatalf("Lock after release = %v", err)
	}
}

func TestOfflineFallback(t *testing.T) {
	s, _, root, opts := installed(t)
	ctx := context.Background()
	s.mu.Lock()
	s.fail = true
	s.mu.Unlock()
	os.WriteFile(filepath.Join(root, "Components", "theme-broken"), []byte("garbage\r\n"), 0o644)

	load := func(o Options, mode LoadMode) (*Snapshot, error) {
		e, err := Open(ctx, o)
		if err != nil {
			t.Fatal(err)
		}
		return e.LoadManifest(ctx, mode)
	}
	snap, err := load(opts, ForChanges)
	if err != nil || snap.Origin != FromBackup || len(snap.Warnings) != 2 || snap.Records["core-launcher"] == nil {
		t.Fatalf("components.bak: %+v, %v", snap, err)
	}
	if _, ok := snap.Broken["theme-broken"]; !ok || len(snap.Broken) != 1 || snap.Records["theme-broken"] != nil {
		t.Errorf("broken = %v", snap.Broken)
	}
	// Once the channel has keys, components.bak (unsigned) is only good for
	// reading.
	signed := opts
	signed.Keys = []trust.PublicKey{{ID: trust.KeyID{1}, Key: make(ed25519.PublicKey, ed25519.PublicKeySize)}}
	if _, err := load(signed, ForChanges); !errors.Is(err, trust.ErrBadSignature) {
		t.Fatalf("signed channel, changes: err = %v", err)
	}
	if snap, err := load(signed, ForReading); err != nil || snap.Origin != FromBackup || len(snap.Warnings) != 3 {
		t.Fatalf("signed channel, reading: %+v, %v", snap, err)
	}

	os.Remove(filepath.Join(root, "Components", "components.bak"))
	if _, err := load(opts, ForReading); err == nil {
		t.Fatal("no manifest anywhere, but LoadManifest succeeded")
	}
}

// TestOfflineAsksAtOnce: with the server unreachable a moment ago, the first
// download failure goes to OnError instead of retrying for RetryBudget.
func TestOfflineAsksAtOnce(t *testing.T) {
	s, _, _, opts := installed(t)
	s.mu.Lock()
	s.fail = true
	s.mu.Unlock()
	opts.Fetch.RetryBudget = time.Minute
	start := time.Now()
	h := &hooks{}
	_, err := run(t, opts, plan.Request{Install: []string{"theme-blue"}}, h)
	if !errors.Is(err, ErrAborted) || len(h.errs) != 1 || time.Since(start) > 10*time.Second {
		t.Fatalf("err = %v, OnError = %+v after %v", err, h.errs, time.Since(start))
	}
}

func TestRootRepair(t *testing.T) {
	s := newServer(t)
	s.publish(baseComps())
	root, mgr := newRoot(t, s.source())
	os.WriteFile(filepath.Join(mgr, "fpm.cfg"), []byte("C:\\Flashpoint\r\n"+s.source()+"\r\n"), 0o644)
	e, err := Open(context.Background(), Options{ManagerDir: mgr})
	if err != nil {
		t.Fatal(err)
	}
	if e.Root() != root {
		t.Errorf("root = %s", e.Root())
	}
	b, _ := os.ReadFile(filepath.Join(mgr, "fpm.cfg"))
	if string(b) != "..\\\r\n"+s.source()+"\r\n" {
		t.Errorf("fpm.cfg = %q", b)
	}
}

func TestNoSignatureRequestWithoutKeys(t *testing.T) {
	s := newServer(t)
	s.publish(baseComps())
	_, mgr := newRoot(t, s.source())
	e, _ := Open(context.Background(), testOptions(mgr))
	if _, err := e.LoadManifest(context.Background(), ForReading); err != nil {
		t.Fatal(err)
	}
	for _, r := range s.requests {
		if strings.HasSuffix(r, ".minisig") {
			t.Errorf("requested %s although no keys are embedded", r)
		}
	}
}

func TestMirrorsAndChannel(t *testing.T) {
	got := mirrors("https://nexus-dev.unstable.life/repository/stable/")
	want := []string{"https://nexus-dev.unstable.life/repository/stable/", "https://nexus-dev.flashpointarchive.org/repository/stable/"}
	if !slices.Equal(got, want) {
		t.Errorf("mirrors = %v", got)
	}
	if got := mirrors("http://localhost/x/"); len(got) != 1 {
		t.Errorf("mirrors(localhost) = %v", got)
	}
	for in, want := range map[string]string{
		state.DefaultSource: "stable",
		"https://x/repository/development/components.xml": "development",
		"http://localhost/components.xml":                 "custom",
		"https://x/A B/components.xml":                    "custom",
	} {
		e, err := Open(context.Background(), Options{ManagerDir: t.TempDir(), RootOverride: t.TempDir(), SourceOverride: in})
		if err != nil {
			t.Fatal(err)
		}
		if e.channel != want {
			t.Errorf("channel(%s) = %s, want %s", in, e.channel, want)
		}
	}
}

func TestFormatSize(t *testing.T) {
	for n, want := range map[int64]string{
		0: "0.0B", 1: "1.0B", 1023: "1,023.0B", 1024: "1.0KB", 1280: "1.2KB", 1536: "1.5KB",
		2715370: "2.6MB", 3 << 30: "3.0GB", 5 << 40: "5,120.0GB", -2048: "-2.0KB",
	} {
		if got := FormatSize(n); got != want {
			t.Errorf("FormatSize(%d) = %s, want %s", n, got, want)
		}
	}
}

// --- components.bak, signing ------------------------------------------------

func TestBackupWrittenOnlyWhenChanged(t *testing.T) {
	s := newServer(t)
	comps := baseComps()
	s.publish(comps)
	root, mgr := newRoot(t, s.source())
	opts := testOptions(mgr)
	ctx := context.Background()
	bak := filepath.Join(root, "Components", "components.bak")
	load := func() *Snapshot {
		t.Helper()
		e, err := Open(ctx, opts)
		if err != nil {
			t.Fatal(err)
		}
		snap, err := e.LoadManifest(ctx, ForReading)
		if err != nil {
			t.Fatal(err)
		}
		return snap
	}

	// A read-only start writes components.bak once and nothing else.
	load()
	good, err := os.ReadFile(bak)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, "Components", ".fpm")); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Components/.fpm created without Lock: %v", err)
	}
	// Unchanged list: not rewritten.
	old := time.Now().Add(-time.Hour).Truncate(time.Second)
	os.Chtimes(bak, old, old)
	load()
	if fi, _ := os.Stat(bak); !fi.ModTime().Equal(old) {
		t.Error("components.bak rewritten although the list did not change")
	}
	// A broken upload is not saved; the offline copy is used.
	s.set("components.xml", []byte("<list><broken"))
	if snap := load(); snap.Origin != FromBackup {
		t.Fatalf("broken upload: origin %v", snap.Origin)
	}
	if b, _ := os.ReadFile(bak); !bytes.Equal(b, good) {
		t.Fatalf("broken list saved: %q", b)
	}
	// A new list replaces it.
	comps[0].ents[1].data = "v2"
	s.publish(comps)
	if snap := load(); snap.Origin != FromNetwork {
		t.Fatalf("new list: origin %v", snap.Origin)
	}
	if b, _ := os.ReadFile(bak); bytes.Equal(b, good) {
		t.Error("components.bak not updated")
	}
}

func TestSignedManifest(t *testing.T) {
	pub, priv, _ := ed25519.GenerateKey(nil)
	key := trust.PublicKey{ID: trust.KeyID{1, 2, 3, 4, 5, 6, 7, 8}, Key: pub}
	s := newServer(t)
	s.publish(baseComps())
	s.mu.Lock()
	body := s.files["components.xml"]
	s.mu.Unlock()
	s.set("components.xml.minisig", trust.Sign(priv, key.ID, body, trust.FormatComment("repo")))
	_, mgr := newRoot(t, s.source())
	opts := testOptions(mgr)
	opts.Keys = []trust.PublicKey{key}

	if _, err := run(t, opts, plan.Request{Install: []string{"theme-base"}}, &hooks{}); err != nil {
		t.Fatalf("signed: %v", err)
	}
	if e, _ := Open(context.Background(), opts); e != nil {
		if snap, err := e.LoadManifest(context.Background(), ForChanges); err != nil || snap.Trust != trust.Valid || snap.Origin != FromNetwork {
			t.Fatalf("signed: %+v, %v", snap, err)
		}
	}
	// The signature vanishes: the list is not trusted, and the unsigned
	// offline copy is not good enough for changes.
	s.mu.Lock()
	delete(s.files, "components.xml.minisig")
	s.mu.Unlock()
	e, _ := Open(context.Background(), opts)
	if _, err := e.LoadManifest(context.Background(), ForChanges); !errors.Is(err, trust.ErrBadSignature) {
		t.Fatalf("unsigned, for changes: %v", err)
	}
}

func TestOpenRootNotFound(t *testing.T) {
	mgr := t.TempDir() // its parent is not a Flashpoint folder
	if _, err := Open(context.Background(), Options{ManagerDir: mgr}); !errors.Is(err, ErrRootNotFound) {
		t.Errorf("no install: err = %v, want ErrRootNotFound", err)
	}
	_, err := Open(context.Background(), Options{ManagerDir: mgr, RootOverride: filepath.Join(mgr, "nope")})
	if !errors.Is(err, ErrRootNotFound) {
		t.Errorf("missing --path folder: err = %v, want ErrRootNotFound", err)
	}
}

func TestSaveSettings(t *testing.T) {
	root, mgr := newRoot(t, "http://localhost/components.xml")
	other, _ := newRoot(t, "")
	cfg := func() string {
		b, err := os.ReadFile(filepath.Join(mgr, "fpm.cfg"))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	const dev = "https://nexus-dev.flashpointarchive.org/repository/development/components.xml"

	// The existing http source may be kept; root elsewhere is absolute.
	if err := SaveSettings(mgr, other, "http://localhost/components.xml"); err != nil {
		t.Fatal(err)
	}
	if got := cfg(); got != other+"\r\nhttp://localhost/components.xml\r\n" {
		t.Errorf("fpm.cfg = %q", got)
	}
	// The Manager folder's parent (here with a trailing separator) is stored
	// relative.
	if err := SaveSettings(mgr, root+string(filepath.Separator), dev); err != nil {
		t.Fatal(err)
	}
	if got := cfg(); got != "..\\\r\n"+dev+"\r\n" {
		t.Errorf("fpm.cfg = %q", got)
	}
	e, err := Open(context.Background(), Options{ManagerDir: mgr})
	if err != nil || e.Root() != root || e.Source() != dev || e.Channel() != "development" {
		t.Fatalf("reopen: %v root=%s source=%s", err, e.Root(), e.Source())
	}
	// Empty values keep the current ones.
	if err := SaveSettings(mgr, "", ""); err != nil || cfg() != "..\\\r\n"+dev+"\r\n" {
		t.Errorf("keep: err = %v, fpm.cfg = %q", err, cfg())
	}

	before := cfg()
	for _, tc := range []struct{ root, source, err string }{
		{t.TempDir(), "", "not a Flashpoint folder"},
		{filepath.Join(root, "nope"), "", "not a Flashpoint folder"},
		{"", "http://example.org/components.xml", "not a valid https URL"},
		{"", "https:///components.xml", "not a valid https URL"},
		{"", "components.xml", "not a valid https URL"},
		{other, "ftp://x/components.xml", "not a valid https URL"}, // nothing written
	} {
		if err := SaveSettings(mgr, tc.root, tc.source); err == nil || !strings.Contains(err.Error(), tc.err) {
			t.Errorf("SaveSettings(%q, %q) = %v, want %q", tc.root, tc.source, err, tc.err)
		}
	}
	if got := cfg(); got != before {
		t.Errorf("a refused save changed fpm.cfg to %q", got)
	}
}

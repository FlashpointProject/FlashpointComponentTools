package main

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"hash/crc32"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

var update = flag.Bool("update", false, "rewrite the golden files in testdata/")

// --- fixtures --------------------------------------------------------------------

type fakeServer struct {
	*httptest.Server
	mu    sync.Mutex
	files map[string][]byte
	down  bool
}

func newFakeServer(t *testing.T) *fakeServer {
	s := &fakeServer{files: map[string][]byte{}}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		b, ok := s.files[r.URL.Path]
		down := s.down
		s.mu.Unlock()
		switch {
		case down:
			http.Error(w, "down", http.StatusServiceUnavailable)
		case !ok:
			http.NotFound(w, r)
		default:
			w.Write(b)
		}
	}))
	t.Cleanup(s.Close)
	return s
}

func (s *fakeServer) set(path string, b []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.files[path] = b
}

func (s *fakeServer) source() string { return s.URL + "/stable/components.xml" }

func readTestdata(t *testing.T, rel string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", filepath.FromSlash(rel)))
	if err != nil {
		t.Fatal(err)
	}
	return b
}

// csharpInstall is a Flashpoint folder whose Components/ holds the real
// records written by the C# manager (testdata/records/csharp), with
// testdata/manifests/small.xml served as the component list.
//
// So that every status occurs, the served copy gives theme-gold a new hash
// (it is then outdated). servedManifest returns the bytes served.
func csharpInstall(t *testing.T) (root, mgr string, srv *fakeServer) {
	srv = newFakeServer(t)
	srv.set("/stable/components.xml", servedManifest(t))
	root = t.TempDir()
	comps := filepath.Join(root, "Components")
	os.MkdirAll(comps, 0o755)
	src := filepath.Join("..", "..", "testdata", "records", "csharp")
	ents, err := os.ReadDir(src)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range ents {
		b, _ := os.ReadFile(filepath.Join(src, e.Name()))
		os.WriteFile(filepath.Join(comps, e.Name()), b, 0o644)
	}
	mgr = filepath.Join(root, "Manager")
	os.MkdirAll(mgr, 0o755)
	os.WriteFile(filepath.Join(mgr, "fpm.cfg"), []byte("..\\\r\n"+srv.source()+"\r\n"), 0o644)
	return root, mgr, srv
}

func servedManifest(t *testing.T) []byte {
	b := readTestdata(t, "manifests/small.xml")
	i := bytes.Index(b, []byte(`<component id="gold"`))
	j := bytes.Index(b[i:], []byte(`hash="`)) + i + len(`hash="`)
	out := bytes.Clone(b)
	copy(out[j:], "ffffffff")
	return out
}

func runFPM(t *testing.T, mgr, stdin string, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errb bytes.Buffer
	code = run(context.Background(), args, env{
		stdin: strings.NewReader(stdin), stdout: &out, stderr: &errb, managerDir: mgr,
	})
	return code, out.String(), errb.String()
}

func golden(t *testing.T, name, got string) {
	t.Helper()
	p := filepath.Join("testdata", name+".golden")
	if *update {
		os.MkdirAll("testdata", 0o755)
		if err := os.WriteFile(p, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(p)
	if err != nil {
		t.Fatalf("%v (run go test -update to create it)", err)
	}
	if got != string(want) {
		t.Errorf("%s: output differs from %s\ngot:\n%q\nwant:\n%q", name, p, got, want)
	}
}

// --- legacy model ----------------------------------------------------------------

// legacyList is an independent port of what legacy fpm prints for `list`:
// GetComponents/Iterate and the Component constructor (fpm src/Helpers.cs:
// 34-160, 242-288) and ListHandler (src/Handlers.cs:69-97), with .NET's
// Console.WriteLine on Windows ending lines in CRLF. It shares no code with
// the implementation.
func legacyList(t *testing.T, manifestXML []byte, componentsDir string, args []string) string {
	t.Helper()
	d := xml.NewDecoder(bytes.NewReader(manifestXML))
	var stack []string // ids of enclosing non-list elements
	var b strings.Builder
	for {
		tok, err := d.Token()
		if err != nil {
			break
		}
		switch el := tok.(type) {
		case xml.StartElement:
			attr := func(n string) string {
				for _, a := range el.Attr {
					if a.Name.Local == n {
						return a.Value
					}
				}
				return ""
			}
			if el.Name.Local == "component" {
				id := strings.Join(append(append([]string(nil), stack...), attr("id")), "-")
				hash := attr("hash")
				downloaded, outdated := false, false
				if data, err := os.ReadFile(filepath.Join(componentsDir, id)); err == nil {
					downloaded = true
					first := strings.SplitN(strings.TrimPrefix(string(data), "\ufeff"), "\n", 2)[0]
					header := strings.Split(strings.TrimRight(first, "\r"), " ")
					if len(header) >= 2 && header[0] != hash {
						outdated = true
					}
				}
				skip := len(args) > 0 && ((args[0] == "available" && downloaded) ||
					(args[0] == "downloaded" && !downloaded) || (args[0] == "updates" && !outdated))
				if !skip {
					prefix := " "
					if downloaded {
						prefix = "*"
						if outdated {
							prefix = "!"
						}
					}
					line := prefix + " " + id
					if len(args) > 0 && args[len(args)-1] == "verbose" {
						line += " (" + attr("title") + ")"
					}
					b.WriteString(line + "\r\n")
				}
			}
			if el.Name.Local != "list" {
				stack = append(stack, attr("id"))
			}
		case xml.EndElement:
			if el.Name.Local != "list" && len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
	return b.String()
}

// --- list golden tests -------------------------------------------------------------

func TestListGolden(t *testing.T) {
	root, mgr, _ := csharpInstall(t)
	small := servedManifest(t)
	for _, args := range [][]string{
		{},
		{"verbose"},
		{"updates", "verbose"},
		{"updates"},
		{"available"},
		{"downloaded", "verbose"},
		{"verbose", "updates"}, // filter is Args[1], verbose is Args.Last(): no filter, not verbose
	} {
		name := strings.Join(append([]string{"list"}, args...), "_")
		t.Run(name, func(t *testing.T) {
			code, out, errOut := runFPM(t, mgr, "", append([]string{"list"}, args...)...)
			if code != 0 {
				t.Fatalf("exit %d, stderr %s", code, errOut)
			}
			if errOut != "" {
				t.Errorf("unexpected stderr: %q", errOut)
			}
			if want := legacyList(t, small, filepath.Join(root, "Components"), args); out != want {
				t.Errorf("differs from legacy fpm\ngot:\n%s\nwant:\n%s", out, want)
			}
			golden(t, name, out)
		})
	}
	// The fixture must exercise every prefix.
	_, out, _ := runFPM(t, mgr, "", "list")
	for _, p := range []string{"* ", "! ", "  "} {
		if !strings.Contains(out, "\r\n"+p) {
			t.Errorf("fixture has no %q lines", p)
		}
	}
}

// TestListCoreDatabase documents the one deliberate difference from legacy
// fpm: an outdated core-database is listed as up to date.
func TestListCoreDatabase(t *testing.T) {
	root, mgr, srv := csharpInstall(t)
	m := bytes.Replace(servedManifest(t), []byte(`<component id="logos"`),
		[]byte(`<component id="database" title="Database" description="d" date-modified="1" download-size="1" install-size="1" path="Data" hash="12345678" />
		<component id="logos"`), 1)
	srv.set("/stable/components.xml", m)
	os.WriteFile(filepath.Join(root, "Components", "core-database"), []byte("00000000 1\r\nData\\flashpoint.sqlite\r\n"), 0o644)
	code, out, _ := runFPM(t, mgr, "", "list", "verbose")
	if code != 0 || !strings.Contains(out, "\r\n* core-database (Database)\r\n") {
		t.Errorf("core-database line: %q", out)
	}
	// Legacy fpm printed "!"; the launcher turns that into up to date, so it
	// parses the same either way.
	legacy := legacyList(t, m, filepath.Join(root, "Components"), []string{"verbose"})
	if !strings.Contains(legacy, "\r\n! core-database (Database)\r\n") {
		t.Fatal("legacy model does not show the difference")
	}
	lp, op := launcherParse(legacy), launcherParse(out)
	if len(lp) != len(op) {
		t.Fatal("different component count")
	}
	for i := range lp {
		if lp[i] != op[i] {
			t.Errorf("launcher sees %+v from legacy, %+v from us", lp[i], op[i])
		}
	}
}

func TestListBrokenRecord(t *testing.T) {
	// Legacy fpm judges a record by its first line only; an unreadable
	// record still counts as downloaded.
	root, mgr, _ := csharpInstall(t)
	os.WriteFile(filepath.Join(root, "Components", "core-logos"), []byte("deadbeef 5\r\nData\\..\\..\\evil\r\n"), 0o644)
	os.WriteFile(filepath.Join(root, "Components", "theme-red"), []byte("garbage\r\n"), 0o644)
	code, out, errOut := runFPM(t, mgr, "", "list")
	if code != 0 || !strings.Contains(out, "\r\n! core-logos\r\n") || !strings.Contains(out, "\r\n* theme-red\r\n") {
		t.Errorf("exit %d, out %q", code, out)
	}
	if !strings.Contains(errOut, "core-logos") {
		t.Errorf("stderr does not mention the broken record: %q", errOut)
	}
	if want := legacyList(t, servedManifest(t), filepath.Join(root, "Components"), nil); out != want {
		t.Errorf("differs from legacy fpm")
	}
}

func TestListOffline(t *testing.T) {
	root, mgr, srv := csharpInstall(t)
	_, online, _ := runFPM(t, mgr, "", "list", "verbose")
	srv.mu.Lock()
	srv.down = true
	srv.mu.Unlock()

	// The verified copy of the last successful check is used.
	code, out, errOut := runFPM(t, mgr, "", "list", "verbose")
	if code != 0 || out != online || !strings.Contains(errOut, "could not download") {
		t.Errorf("offline with cache: exit %d, same output %v, stderr %q", code, out == online, errOut)
	}
	// Without it, components.bak (here: the C# manager's real one).
	os.RemoveAll(filepath.Join(root, "Components", ".fpm"))
	code, out, errOut = runFPM(t, mgr, "", "list", "verbose")
	if code != 0 || len(out) == 0 || !strings.Contains(errOut, "components.bak") {
		t.Errorf("offline with components.bak: exit %d, stderr %q", code, errOut)
	}
	// Nothing at all: exit 1.
	os.RemoveAll(filepath.Join(root, "Components", ".fpm"))
	os.Remove(filepath.Join(root, "Components", "components.bak"))
	if code, out, _ := runFPM(t, mgr, "", "list", "verbose"); code != 1 || out != "" {
		t.Errorf("no manifest: exit %d, stdout %q", code, out)
	}
}

// --- launcher contract ------------------------------------------------------------

type componentState int

const (
	uninstalled componentState = iota
	upToDate
	needsUpdate
)

type componentStatus struct {
	id, name string
	state    componentState
}

// jsSubstring is JavaScript's String.prototype.substring: NaN/negative
// arguments become 0, both are clamped to the length, and they are swapped
// if start > end.
func jsSubstring(s string, start, end int) string {
	clamp := func(i int) int { return min(max(i, 0), len(s)) }
	start, end = clamp(start), clamp(end)
	if start > end {
		start, end = end, start
	}
	return s[start:end]
}

// launcherParse is a faithful port of the launcher's parser of
// `fpm.exe list verbose` (launcher src/back/index.ts:773-806).
func launcherParse(stdout string) []componentStatus {
	var statuses []componentStatus
	for _, line := range strings.Split(stdout, "\n") { // stdout.split('\n')
		if line == "" { //                                  .filter(line => line !== '')
			continue
		}
		nameIdx := strings.Index(line, "(")                      // line.indexOf('(')
		id := strings.TrimSpace(jsSubstring(line, 1, nameIdx-1)) // line.substring(1, nameIdx - 1).trim()
		name := jsSubstring(line, nameIdx+1, len(line)-2)        // line.substring(nameIdx + 1, line.length - 2)
		var state componentState
		switch line[0] { // line.charAt(0)
		case '*':
			state = upToDate
		case '!':
			state = needsUpdate
		default:
			state = uninstalled
		}
		if id == "core-database" && state == needsUpdate {
			state = upToDate
		}
		statuses = append(statuses, componentStatus{id, name, state})
	}
	return statuses
}

func TestLauncherParsesListVerbose(t *testing.T) {
	root, mgr, _ := csharpInstall(t)
	code, out, _ := runFPM(t, mgr, "", "list", "verbose")
	if code != 0 {
		t.Fatal(code)
	}
	got := launcherParse(out)

	// Expected values straight from the manifest and the record headers.
	small := servedManifest(t)
	var want []componentStatus
	d := xml.NewDecoder(bytes.NewReader(small))
	var stack []string
	for tok, err := d.Token(); err == nil; tok, err = d.Token() {
		switch el := tok.(type) {
		case xml.StartElement:
			a := map[string]string{}
			for _, at := range el.Attr {
				a[at.Name.Local] = at.Value
			}
			if el.Name.Local == "component" {
				id := strings.Join(append(append([]string(nil), stack...), a["id"]), "-")
				st := uninstalled
				if rec, err := os.ReadFile(filepath.Join(root, "Components", id)); err == nil {
					st = upToDate
					if !bytes.HasPrefix(rec, []byte(a["hash"]+" ")) && id != "core-database" {
						st = needsUpdate
					}
				}
				want = append(want, componentStatus{id, a["title"], st})
			}
			if el.Name.Local != "list" {
				stack = append(stack, a["id"])
			}
		case xml.EndElement:
			if el.Name.Local != "list" {
				stack = stack[:len(stack)-1]
			}
		}
	}
	if len(got) != len(want) || len(want) < 200 {
		t.Fatalf("parsed %d components, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("component %d: launcher parsed %+v, want %+v", i, got[i], want[i])
		}
	}

	// Why CRLF matters: with LF the launcher cuts the last title character.
	if lf := launcherParse(strings.ReplaceAll(out, "\r\n", "\n")); lf[0].name == want[0].name {
		t.Error("expected LF output to break the launcher's title parsing")
	}
}

// --- other commands -----------------------------------------------------------------

func makeZip(t *testing.T, files map[string]string) []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, data := range files {
		w, _ := zw.Create(name)
		w.Write([]byte(data))
	}
	zw.Close()
	return buf.Bytes()
}

// tinyInstall serves a one-component manifest over plain http.
func tinyInstall(t *testing.T) (root, mgr string, srv *fakeServer) {
	srv = newFakeServer(t)
	z := makeZip(t, map[string]string{"Test/theme.css": "body{}"})
	srv.set("/stable/theme-test.zip", z)
	srv.set("/stable/components.xml", []byte(fmt.Sprintf(`<?xml version='1.0' encoding='utf-8'?>
<list url="%s/stable/">
	<category id="theme" title="Themes" description="d">
		<component id="test" title="Test Theme" description="A test." date-modified="1699132293" download-size="%d" install-size="6" path="Data/Themes" hash="%08x" />
	</category>
	<category id="core" title="Core" description="d">
		<component id="meta" title="Meta" description="Metadata only." date-modified="1699132293" download-size="0" install-size="0" hash="00000000" />
	</category>
</list>`, srv.URL, len(z), crc32.ChecksumIEEE(z))))
	root = t.TempDir()
	for _, d := range []string{"Launcher", "Data", "Manager"} {
		os.MkdirAll(filepath.Join(root, d), 0o755)
	}
	mgr = filepath.Join(root, "Manager")
	os.WriteFile(filepath.Join(mgr, "fpm.cfg"), []byte("..\\\r\n"+srv.source()+"\r\n"), 0o644)
	return root, mgr, srv
}

func TestDownloadRemoveNonInteractive(t *testing.T) {
	root, mgr, _ := tinyInstall(t)
	// stdin is not a terminal and has no answer: legacy fpm hung here.
	code, out, errOut := runFPM(t, mgr, "", "download", "theme-test")
	if code != 0 {
		t.Fatalf("download: exit %d\n%s%s", code, out, errOut)
	}
	if !strings.Contains(out, "1 component(s) will be downloaded:\r\n  theme-test\r\n") ||
		!strings.Contains(out, "download: Successfully downloaded 1 components\r\n") ||
		!strings.Contains(errOut, "not a terminal") {
		t.Errorf("stdout %q\nstderr %q", out, errOut)
	}
	if b, _ := os.ReadFile(filepath.Join(root, "Data", "Themes", "Test", "theme.css")); string(b) != "body{}" {
		t.Error("file not installed")
	}
	if _, err := os.Stat(filepath.Join(root, "Logs", "fpm.log")); err != nil {
		t.Errorf("no log file: %v", err)
	}

	code, out, errOut = runFPM(t, mgr, "", "download", "theme-test", "nope")
	if code != 0 || !strings.Contains(errOut, "Component or category nope does not exist and will be skipped") ||
		!strings.Contains(errOut, "Component theme-test is already downloaded and will be skipped") ||
		!strings.Contains(out, "download: No components to download") {
		t.Errorf("second download: exit %d\n%s%s", code, out, errOut)
	}

	// update installs the missing core component (metadata-only: header-only record).
	code, out, errOut = runFPM(t, mgr, "", "-y", "update")
	if code != 0 || !strings.Contains(out, "update: Successfully updated 0 components and downloaded 1 components") {
		t.Errorf("update: exit %d\n%s%s", code, out, errOut)
	}
	if b, _ := os.ReadFile(filepath.Join(root, "Components", "core-meta")); string(b) != "00000000 0\r\n" {
		t.Errorf("core-meta record = %q", b)
	}

	if code, _, errOut := runFPM(t, mgr, "", "remove", "core-meta"); code != 1 || !strings.Contains(errOut, "required") {
		t.Errorf("remove required: exit %d %q", code, errOut)
	}
	code, out, errOut = runFPM(t, mgr, "", "remove", "theme-test")
	if code != 0 || !strings.Contains(out, "remove: Successfully removed 1 components") {
		t.Errorf("remove: exit %d\n%s%s", code, out, errOut)
	}
	if _, err := os.Stat(filepath.Join(root, "Data", "Themes")); err == nil {
		t.Error("emptied Data/Themes not pruned")
	}
	if code, out, _ := runFPM(t, mgr, "", "verify"); code != 0 || out != "* core-meta\r\n" {
		t.Errorf("verify: exit %d %q", code, out)
	}
}

func TestInteractiveNo(t *testing.T) {
	_, mgr, _ := tinyInstall(t)
	var out, errb bytes.Buffer
	code := run(context.Background(), []string{"download", "theme-test"}, env{
		stdin: strings.NewReader("maybe\nn\n"), stdout: &out, stderr: &errb, managerDir: mgr, stdinTTY: true,
	})
	if code != 0 || strings.Count(out.String(), "Is this OK? [y/n]: ") != 2 || strings.Contains(out.String(), "Successfully") {
		t.Errorf("exit %d, stdout %q", code, out.String())
	}
}

func TestInfo(t *testing.T) {
	_, mgr, _ := csharpInstall(t)
	code, out, _ := runFPM(t, mgr, "", "info", "theme-bronze")
	if code != 0 {
		t.Fatal(code)
	}
	for _, want := range []string{
		"ID:             theme-bronze\r\n", "Title:          Bronze\r\n", "Download size:  708.0B\r\n",
		"Install size:   747.0B\r\n", "CRC32:          4b392d03\r\n\r\n", "Dependencies: \r\n  theme-colorized\r\n\r\n",
		"Required?       No\r\n", "Downloaded?     Yes\r\n", "Up-to-date?     ",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("info lacks %q:\n%s", want, out)
		}
	}
	if code, _, _ := runFPM(t, mgr, "", "info", "nope"); code != 1 {
		t.Errorf("info nope: exit %d", code)
	}
}

func TestPathSourceUsage(t *testing.T) {
	mgr := t.TempDir()
	if code, out, _ := runFPM(t, mgr, "", "path"); code != 0 || out != filepath.Dir(mgr)+"\r\n" {
		t.Errorf("path default: %d %q", code, out)
	}
	if _, err := os.Stat(filepath.Join(mgr, "fpm.cfg")); err == nil {
		t.Error("reading the path created fpm.cfg")
	}
	if code, _, _ := runFPM(t, mgr, "", "source", "https://example.org/repository/development/components.xml"); code != 0 {
		t.Fatal(code)
	}
	b, _ := os.ReadFile(filepath.Join(mgr, "fpm.cfg"))
	if string(b) != "..\\\r\nhttps://example.org/repository/development/components.xml\r\n" {
		t.Errorf("fpm.cfg = %q", b)
	}
	dir := t.TempDir()
	runFPM(t, mgr, "", "path", dir)
	if code, out, _ := runFPM(t, mgr, "", "path"); code != 0 || out != dir+"\r\n" {
		t.Errorf("path = %q", out)
	}
	if code, out, _ := runFPM(t, mgr, "", "source"); code != 0 || out != "https://example.org/repository/development/components.xml\r\n" {
		t.Errorf("source = %q", out)
	}
	// A bad URL is refused (legacy behaviour) and the old one kept, so the
	// launcher's `fpm list` keeps working.
	if code, _, errOut := runFPM(t, mgr, "", "source", "not a url"); code != 1 || !strings.Contains(errOut, "The specified URL is invalid") {
		t.Errorf("source not a url: exit %d %q", code, errOut)
	}
	if _, out, _ := runFPM(t, mgr, "", "source"); out != "https://example.org/repository/development/components.xml\r\n" {
		t.Errorf("source after a bad URL = %q", out)
	}

	for _, tc := range []struct {
		args []string
		code int
	}{
		{nil, 0}, {[]string{"help"}, 0}, {[]string{"--help"}, 0}, {[]string{"version"}, 0}, {[]string{"--version"}, 0},
		{[]string{"frobnicate"}, 2}, {[]string{"info"}, 2}, {[]string{"remove"}, 2}, {[]string{"list", "--bogus"}, 2},
		{[]string{"--parallel", "2", "list"}, 2}, {[]string{"--path"}, 2},
	} {
		if code, _, _ := runFPM(t, mgr, "", tc.args...); code != tc.code {
			t.Errorf("fpm %v: exit %d, want %d", tc.args, code, tc.code)
		}
	}
	if _, out, _ := runFPM(t, mgr, "", "version"); !strings.HasPrefix(out, "fpm dev (") {
		t.Errorf("version = %q", out)
	}
}

func TestParseArgs(t *testing.T) {
	o, pos, err := parseArgs([]string{"list", "--path=C:\\FP", "-y", "verbose", "--source", "http://x/y.xml", "--", "--literal"})
	if err != nil || o.path != `C:\FP` || !o.yes || o.source != "http://x/y.xml" ||
		strings.Join(pos, " ") != "list verbose --literal" {
		t.Errorf("parseArgs = %+v %v %v", o, pos, err)
	}
}

func TestSplitIDs(t *testing.T) {
	// The launcher passes several IDs as one argument.
	if got := splitIDs([]string{"supportpack-a supportpack-b", " x "}); strings.Join(got, ",") != "supportpack-a,supportpack-b,x" {
		t.Errorf("splitIDs = %q", got)
	}
}

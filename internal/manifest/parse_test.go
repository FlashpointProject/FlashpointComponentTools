package manifest

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func parseFile(t *testing.T, name string) *Manifest {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "testdata", "manifests", name))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func parseString(s string) (*Manifest, error) { return Parse(strings.NewReader(s)) }

func TestParseSnapshots(t *testing.T) {
	// The four dangling references present in both real channels (02-manifest.md §3.3).
	dangling := []string{
		`"supportpack-babyz" depends on "supportpack-netscape"`,
		`"supportpack-quicktime" depends on "supportpack-netscape"`,
		`"supportpack-htmltime" depends on "supportpack-common-activex"`,
		`"supportpack-wildtangent" depends on "supportpack-common-activex"`,
	}
	for _, tt := range []struct {
		file       string
		components int
		base       string
	}{
		{"stable.xml", 236, "https://nexus-dev.unstable.life/repository/stable/"},
		{"development.xml", 247, "https://nexus-dev.unstable.life/repository/development/"},
		{"small.xml", 231, "https://nexus-dev.unstable.life/repository/stable/"},
	} {
		t.Run(tt.file, func(t *testing.T) {
			m := parseFile(t, tt.file)
			if len(m.Components) != tt.components || len(m.Order) != tt.components || len(m.Categories) != 9 || m.BaseURL != tt.base {
				t.Errorf("%d components (order %d), %d categories, base %q", len(m.Components), len(m.Order), len(m.Categories), m.BaseURL)
			}
			if len(m.Warnings) != len(dangling) {
				t.Errorf("warnings = %q", m.Warnings)
			}
			for _, d := range dangling {
				if !slices.ContainsFunc(m.Warnings, func(w string) bool { return strings.Contains(w, d) }) {
					t.Errorf("missing warning for %s", d)
				}
			}
			for _, c := range m.Components {
				if c.Title == "" || c.Hash == "" || c.Modified.IsZero() {
					t.Errorf("%s: incomplete %+v", c.ID, c)
				}
			}
		})
	}
}

func TestStableDetails(t *testing.T) {
	m := parseFile(t, "stable.xml")
	if m.Order[0] != "core-server-automount" { // subcategories come first in the generator's output
		t.Errorf("Order[0] = %q", m.Order[0])
	}
	cfg := m.Components["core-configuration"]
	if cfg == nil || cfg.Path != "" || cfg.Hash != "9d40c19f" || cfg.CRC32 != 0x9d40c19f || cfg.InstallSize != 56883 || cfg.DownloadSize != 15445 {
		t.Errorf("core-configuration = %+v", cfg)
	}
	if c := m.Components["core-launcher"]; c.Path != "Launcher" || len(c.Depends) != 5 {
		t.Errorf("core-launcher = %+v", c)
	}
	if c := m.Components["theme-bronze"]; c.Path != "Data/Themes" || !slices.Equal(c.Depends, []string{"theme-colorized"}) {
		t.Errorf("theme-bronze = %+v", c)
	}
	if m.Components["supportpack-html5"].HasArchive() || !m.Components["supportpack-flash"].HasArchive() {
		t.Error("HasArchive must follow install-size")
	}
	if !m.Categories["core"].Required || !m.Categories["core-server"].Required || m.Categories["extra"].Required {
		t.Error("category Required flags wrong")
	}
	if len(m.Roots) == 0 || m.Roots[0].Items[0].Category == nil {
		t.Error("category tree not built")
	}
	for _, c := range m.Components {
		wantReq, wantOn := strings.HasPrefix(c.ID, "core-"), !strings.HasPrefix(c.ID, "extra-")
		if c.Required != wantReq || c.DefaultOn != wantOn || c.NeverOutdate != (c.ID == "core-database") {
			t.Errorf("%s: policy req=%v on=%v never=%v", c.ID, c.Required, c.DefaultOn, c.NeverOutdate)
		}
	}
}

func TestResolve(t *testing.T) {
	m := parseFile(t, "stable.xml")
	ids := func(cs []*Component) (out []string) {
		for _, c := range cs {
			out = append(out, c.ID)
		}
		return out
	}
	if cs, ok := m.Resolve("core-launcher"); !ok || !slices.Equal(ids(cs), []string{"core-launcher"}) {
		t.Errorf("component: %v %v", ids(cs), ok)
	}
	if cs, ok := m.Resolve("core-server"); !ok || !slices.Equal(ids(cs), []string{"core-server-automount", "core-server-gamezip", "core-server-legacy"}) {
		t.Errorf("subcategory: %v %v", ids(cs), ok)
	}
	if cs, ok := m.Resolve("supportpack"); !ok || len(cs) != 151+12 {
		t.Errorf("supportpack: %d %v", len(cs), ok)
	}
	for _, bad := range []string{"", "cor", "core-", "Core", "core-launcher-x"} {
		if _, ok := m.Resolve(bad); ok {
			t.Errorf("Resolve(%q) found something", bad)
		}
	}
}

// doc wraps extra component attributes into a minimal valid manifest.
func doc(attrs string) string {
	return `<?xml version='1.0' encoding='utf-8'?>
<list url="https://example.org/repo/">
	<category id="cat" title="Cat" description="d">
		<component id="a" title="A" description="d" hash="0011aaff" ` + attrs + `/>
	</category>
</list>`
}

func TestParseRejects(t *testing.T) {
	r := func(old, new string) string { return strings.Replace(doc(""), old, new, 1) }
	comp := func(id, deps string) string {
		return `<component id="` + id + `" title="X" description="d" hash="00000000" depends="` + deps + `"/>`
	}
	tests := []struct{ name, xml, want string }{
		{"traversal", doc(`path="..\..\Windows"`), `".."`},
		{"drive", doc(`path="C:\x"`), "<>:"},
		{"absolute", doc(`path="/etc"`), "absolute"},
		{"UNC", doc(`path="\\server\share"`), "absolute"},
		{"device", doc(`path="Data/nul.txt"`), "reserved"},
		{"trailing dot", doc(`path="Data."`), "dot or space"},
		{"control", doc(`path="a&#9;b"`), "control"},
		{"id slash", r(`id="a"`, `id="a/b"`), "invalid"},
		{"id dash clash", r("</category>", comp("a-b", "")+`</category><category id="cat-a" title="x" description="y">`+comp("b", "")+"</category>"), "duplicate ID"},
		{"id dot", r(`id="a"`, `id="a.b"`), "invalid"},
		{"id upper", r(`id="a"`, `id="A"`), "invalid"},
		{"id missing", r(`id="a"`, ``), "invalid"},
		{"category id", r(`id="cat"`, `id="../x"`), "invalid"},
		{"title missing", r(`title="A"`, ``), "title"},
		{"duplicate", r("</category>", comp("a", "")+"</category>"), "duplicate ID"},
		{"dup cat vs comp", r("</category>", `<category id="a" title="x" description="y"/></category>`), "duplicate ID"},
		{"hash short", r(`hash="0011aaff"`, `hash="0011aaf"`), "hash"},
		{"hash nonhex", r(`hash="0011aaff"`, `hash="0011aafg"`), "hash"},
		{"hash signed", r(`hash="0011aaff"`, `hash="+011aaff"`), "hash"},
		{"hash missing", r(`hash="0011aaff"`, ``), "hash"},
		{"size negative", doc(`install-size="-1"`), "install-size"},
		{"size nonnum", doc(`download-size="12MB"`), "download-size"},
		{"date bad", doc(`date-modified="yesterday"`), "date-modified"},
		{"sha256 bad", doc(`sha256="abcd"`), "sha256"},
		{"dep invalid", doc(`depends="cat-ok ../evil"`), "not a valid component ID"},
		{"self dep", doc(`depends="cat-a"`), "cat-a -> cat-a"},
		{"cycle", strings.Replace(doc(`depends="cat-b"`), "</category>", comp("b", "cat-c")+comp("c", "cat-a")+"</category>", 1), "cat-a -> cat-b -> cat-c -> cat-a"},
		{"no list", `<other url="https://x/"/>`, "expected <list>"},
		{"no url", `<list></list>`, "https"},
		{"http", r("https://", "http://"), "only https"},
		{"ftp", r("https://", "ftp://"), "https"},
		{"relative url", r("https://example.org/repo/", "repo/"), "https"},
		{"url query", r("repo/", "repo/?x=1"), "query"},
		{"truncated", doc("")[:120], ""},
		{"empty", ``, "empty"},
		{"garbage", `not xml at all`, ""},
		{"utf16", `<?xml version="1.0" encoding="utf-16"?><list url="https://x/"/>`, ""},
		{"too large", `<list url="https://x/"><!--` + strings.Repeat("x", MaxManifestSize) + `--></list>`, "larger than"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := parseString(tt.xml)
			if err == nil {
				t.Fatalf("accepted: %+v", m)
			}
			if !errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error %q: want ErrInvalid containing %q", err, tt.want)
			}
		})
	}
}

func TestParseTolerates(t *testing.T) {
	x := "\xef\xbb\xbf" + `<?xml version="1.0" encoding="UTF-8"?>
<!-- leading comment -->
<list url=" https://example.org/repo " generator="v9">
	<?pi something?>
	<notice text="unknown element under list"><nested/></notice>
	<category id="core" title="Core" description="d" newattr="1">
		stray text
		<future/>
		<component id="x" title="Multi&#10;line" description="a&#13;&#10;b" hash="ABCDEF01"
			depends="  core-y	core-y  " path="Data\Sub//Dir/./" sha256="` + strings.Repeat("ab", 32) + `" unknown="1"/>
		<component id="y" title="Y" description="d" hash="00000000" download-size="" install-size=" 5 "><child/></component>
	</category>
	<component id="toplevel" title="T" description="d" hash="00000000"/>
	<category id="extra" title="E" description="d"><component id="z" title="Z" description="d" hash="00000000"/></category>
	<category id="empty" title="E" description="d"></category>
</list>
<!-- trailing comment -->`
	m, err := parseString(x)
	if err != nil {
		t.Fatal(err)
	}
	if m.BaseURL != "https://example.org/repo/" || !slices.Equal(m.Order, []string{"core-x", "core-y", "extra-z"}) || len(m.Warnings) != 0 {
		t.Errorf("BaseURL %q, Order %v, warnings %q", m.BaseURL, m.Order, m.Warnings)
	}
	c := m.Components["core-x"]
	if c.Title != "Multi line" || c.Description != "a\nb" || c.Path != "Data/Sub/Dir" ||
		c.Hash != "ABCDEF01" || c.CRC32 != 0xabcdef01 || len(c.SHA256) != 32 ||
		!slices.Equal(c.Depends, []string{"core-y", "core-y"}) || !c.Modified.IsZero() || !c.Required {
		t.Errorf("core-x = %+v", c)
	}
	if y := m.Components["core-y"]; y.InstallSize != 5 || y.DownloadSize != 0 || !y.HasArchive() {
		t.Errorf("core-y = %+v", y)
	}
	if z := m.Components["extra-z"]; z.DefaultOn || z.Required {
		t.Errorf("extra-z policy = %+v", z)
	}
	if cs, ok := m.Resolve("empty"); !ok || len(cs) != 0 {
		t.Errorf("Resolve(empty) = %v %v", cs, ok)
	}
}

func TestParseDashInID(t *testing.T) {
	// The C# clients accept it, so it must not reject the whole list.
	m, err := parseString(strings.Replace(doc(""), `id="a"`, `id="a-b"`, 1))
	if err != nil {
		t.Fatal(err)
	}
	if m.Components["cat-a-b"] == nil || len(m.Warnings) != 1 || !strings.Contains(m.Warnings[0], `"a-b"`) {
		t.Errorf("components %v, warnings %q", m.Order, m.Warnings)
	}
}

func TestParseAllowInsecure(t *testing.T) {
	x := strings.Replace(doc(""), "https://example.org/repo/", "http://localhost:8080/repo", 1)
	if _, err := parseString(x); err == nil {
		t.Fatal("http accepted without AllowInsecure")
	}
	m, err := ParseWithOptions(strings.NewReader(x), Options{AllowInsecure: true})
	if err != nil || m.BaseURL != "http://localhost:8080/repo/" {
		t.Fatalf("%v %v", m, err)
	}
}

func TestCleanRelPath(t *testing.T) {
	for in, want := range map[string]string{
		"": "", ".": "", "./": "", "Data/": "Data", "Data//x/./y": "Data/x/y",
		`Data\Themes\Bronze\theme.css`: "Data/Themes/Bronze/theme.css",
		" leading space ok":            " leading space ok",
		".preferences.defaults.json":   ".preferences.defaults.json",
		"Jeux vidéo/ünï":               "Jeux vidéo/ünï",
		"console":                      "console", "COM10": "COM10", "connect.txt": "connect.txt", "LPT": "LPT",
	} {
		if got, err := CleanRelPath(in); err != nil || got != want {
			t.Errorf("CleanRelPath(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{
		"..", "../x", `..\..\Windows`, "a/../b", "/etc", `\Windows`, `\\server\share`, "//server/share",
		`C:\x`, "C:x", "file.txt:stream", "a\x00b", "a\nb", "a\x7fb", "a<b", `a"b`, "a|b", "a?b", "a*b",
		"Data.", "Data /x", "x/..../", "CON", "Nul", "nul.txt", "aux.tar.gz", "Data/PRN", "lpt9.log",
		"conin$", "NUL .txt", "CLOCK$.txt", "\xff\xfe", strings.Repeat("a", 256),
	} {
		if got, err := CleanRelPath(in); !errors.Is(err, ErrUnsafePath) {
			t.Errorf("CleanRelPath(%q) = %q, %v; want ErrUnsafePath", in, got, err)
		}
	}
}

func TestValidID(t *testing.T) {
	for _, id := range []string{"core", "core-launcher", "supportpack-common-secureplayer", "a_b-c1", "x3d"} {
		if !ValidID(id) {
			t.Errorf("ValidID(%q) = false", id)
		}
	}
	for _, id := range []string{"", "-", "a-", "-a", "a--b", "A", "a.b", "a/b", "..", "a b", "con", "com1", "_a", "a-_b", strings.Repeat("a", 201)} {
		if ValidID(id) {
			t.Errorf("ValidID(%q) = true", id)
		}
	}
}

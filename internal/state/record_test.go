package state

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
)

const goldenDir = "../../testdata/records/csharp"

func goldenRecords(t testing.TB) map[string][]byte {
	t.Helper()
	entries, err := os.ReadDir(goldenDir)
	if err != nil {
		t.Fatal(err)
	}
	out := make(map[string][]byte)
	for _, e := range entries {
		if e.Name() == BackupManifestName {
			continue
		}
		b, err := os.ReadFile(filepath.Join(goldenDir, e.Name()))
		if err != nil {
			t.Fatal(err)
		}
		out[e.Name()] = b
	}
	return out
}

// Records written by the C# manager must survive parse + MarshalLegacy
// byte for byte, so old and new managers can share an install.
func TestGoldenRoundTrip(t *testing.T) {
	golden := goldenRecords(t)
	if len(golden) != 63 {
		t.Fatalf("expected 63 golden records, found %d", len(golden))
	}
	for id, data := range golden {
		r, err := ParseRecord(id, data)
		if err != nil {
			t.Errorf("%s: %v", id, err)
			continue
		}
		if got := r.MarshalLegacy(); !bytes.Equal(got, data) {
			t.Errorf("%s: round trip differs\n got %q\nwant %q", id, got, data)
		}
		for _, f := range r.Files {
			if strings.Contains(f, `\`) {
				t.Errorf("%s: file %q not slash-separated", id, f)
			}
		}
	}
	r, _ := ParseRecord("theme-bronze", golden["theme-bronze"])
	want := &Record{ID: "theme-bronze", Hash: "4b392d03", InstallSize: 747,
		Depends: []string{"theme-colorized"}, Files: []string{"Data/Themes/Bronze/theme.css"}}
	if !reflect.DeepEqual(r, want) {
		t.Errorf("theme-bronze = %+v", r)
	}
	r, _ = ParseRecord("core-configuration", golden["core-configuration"])
	if r.Files[0] != "Data/services.json" || r.Depends != nil {
		t.Errorf("core-configuration = %+v", r)
	}
}

func TestParseRecordLenient(t *testing.T) {
	want := &Record{ID: "a-b", Hash: "0011aaff", InstallSize: 12, Depends: []string{"x-y", "z-w"},
		Files: []string{"Data/one.txt", "FPSoftware/S3 Thing/two"}}
	for name, in := range map[string]string{
		"crlf":        "0011aaff 12 x-y z-w\r\nData\\one.txt\r\nFPSoftware\\S3 Thing\\two\r\n",
		"lf":          "0011aaff 12 x-y z-w\nData/one.txt\nFPSoftware/S3 Thing/two\n",
		"bom":         "\xef\xbb\xbf0011aaff 12 x-y z-w\r\nData\\one.txt\r\nFPSoftware\\S3 Thing\\two",
		"whitespace":  "0011aaff  12\tx-y z-w  \r\n\r\nData\\one.txt \t\r\n\nFPSoftware/S3 Thing/two\n\n",
		"no newline":  "0011aaff 12 x-y z-w\nData/one.txt\nFPSoftware/S3 Thing/two",
		"mixed slash": "0011aaff 12 x-y z-w\r\nData/./one.txt\r\nFPSoftware\\S3 Thing/two\r\n",
	} {
		got, err := ParseRecord("a-b", []byte(in))
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Errorf("%s: got %+v, %v", name, got, err)
		}
	}
	// Header only: metadata-only component (e.g. supportpack-html5).
	r, err := ParseRecord("supportpack-html5", []byte("bd2d3a2c 0 supportpack-flash\r\n"))
	if err != nil || r.Files != nil || r.InstallSize != 0 {
		t.Errorf("header only: %+v %v", r, err)
	}
	if got := string(r.MarshalLegacy()); got != "bd2d3a2c 0 supportpack-flash\r\n" {
		t.Errorf("marshal header only = %q", got)
	}
}

func TestParseRecordRejects(t *testing.T) {
	for name, tt := range map[string]struct{ id, data string }{
		"empty":       {"a-b", ""},
		"blank":       {"a-b", "\r\n\r\n"},
		"one token":   {"a-b", "0011aaff\r\n"},
		"size nonnum": {"a-b", "0011aaff 12MB\r\n"},
		"size neg":    {"a-b", "0011aaff -1\r\n"},
		"size plus":   {"a-b", "0011aaff +1\r\n"},
		"size huge":   {"a-b", "0011aaff 99999999999999999999\r\n"},
		"bad id":      {"../x", "0011aaff 1\r\n"},
		"bad id 2":    {"core-launcher.txt", "0011aaff 1\r\n"},
	} {
		_, err := ParseRecord(tt.id, []byte(tt.data))
		if err == nil || !errors.Is(err, ErrCorrupt) {
			t.Errorf("%s: got %v, want ErrCorrupt", name, err)
		}
	}
}

func TestParseRecordSkipsUnsafeLines(t *testing.T) {
	for name, line := range map[string]string{
		"traversal": `..\..\Windows\win.ini`,
		"absolute":  `C:\Windows\win.ini`,
		"unix abs":  "/etc/passwd",
		"unc":       `\\server\share\x`,
		"nul":       "Data\\a\x00b",
		"device":    `Data\NUL`,
		"dot line":  ".",
	} {
		r, err := ParseRecord("a-b", []byte("0011aaff 1\r\nData\\ok\r\n"+line+"\r\n"))
		if err != nil || !slices.Equal(r.Files, []string{"Data/ok"}) || len(r.Warnings) != 1 || !strings.Contains(r.Warnings[0], "line 3") {
			t.Errorf("%s: got %+v, %v", name, r, err)
		}
	}
}

func TestReadRecords(t *testing.T) {
	root := t.TempDir()
	comp := ComponentsDir(root)
	must(t, os.MkdirAll(filepath.Join(comp, PrivateDirName), 0o755))
	must(t, os.MkdirAll(filepath.Join(comp, "subdir"), 0o755))
	golden := goldenRecords(t)
	for id, data := range golden {
		must(t, os.WriteFile(filepath.Join(comp, id), data, 0o644))
	}
	bak, err := os.ReadFile(filepath.Join(goldenDir, BackupManifestName))
	must(t, err)
	must(t, os.WriteFile(filepath.Join(comp, BackupManifestName), bak, 0o644))
	// Not records: extension (old-style info file), dot-file, our temp file.
	must(t, os.WriteFile(filepath.Join(comp, "core-launcher.txt"), []byte("garbage"), 0o644))
	must(t, os.WriteFile(filepath.Join(comp, ".hidden"), []byte("garbage"), 0o644))
	must(t, os.WriteFile(filepath.Join(comp, ".components.bak.tmp-123"), []byte("x"), 0o644))
	// Corrupt ones: reported, not fatal. An unsafe line only drops the line.
	must(t, os.WriteFile(filepath.Join(comp, "theme-broken"), []byte("garbage\r\n"), 0o644))
	must(t, os.WriteFile(filepath.Join(comp, "theme-evil"), []byte("00000000 1\r\n..\\..\\x\r\n"), 0o644))
	must(t, os.WriteFile(filepath.Join(comp, "Not-An-ID"), []byte("00000000 1\r\n"), 0o644))

	recs, errs := ReadRecords(root)
	if len(recs) != len(golden)+1 {
		t.Errorf("got %d records, want %d", len(recs), len(golden)+1)
	}
	if r := recs["theme-evil"]; r == nil || r.Files != nil || len(r.Warnings) != 1 {
		t.Errorf("theme-evil = %+v", r)
	}
	for id := range golden {
		if recs[id] == nil || recs[id].ID != id {
			t.Errorf("missing %s", id)
		}
	}
	var bad []string
	for _, err := range errs {
		var re *RecordError
		if !errors.As(err, &re) || !errors.Is(err, ErrCorrupt) {
			t.Errorf("unexpected error type %T: %v", err, err)
			continue
		}
		bad = append(bad, re.ID)
	}
	slices.Sort(bad)
	if !slices.Equal(bad, []string{"Not-An-ID", "theme-broken"}) {
		t.Errorf("errors for %v: %v", bad, errs)
	}

	// Missing Components/: nothing installed, no error.
	recs, errs = ReadRecords(t.TempDir())
	if len(recs) != 0 || errs != nil {
		t.Errorf("empty root: %v %v", recs, errs)
	}

}

func TestIsRecordName(t *testing.T) {
	for name, want := range map[string]bool{
		"core-launcher": true, "foo.": true,
		"components.bak": false, "Components.BAK": false, "core-launcher.txt": false,
		".fpm": false, ".x": false, "": false,
	} {
		if got := isRecordName(name); got != want {
			t.Errorf("isRecordName(%q) = %v", name, got)
		}
	}
}

func must(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

func FuzzParseRecord(f *testing.F) {
	for _, id := range []string{"core-manager", "theme-bronze", "extra-s3bucketscraper"} {
		b, err := os.ReadFile(filepath.Join(goldenDir, id))
		if err != nil {
			f.Fatal(err)
		}
		f.Add(b)
	}
	f.Add([]byte("0 0\n..\\x\n"))
	f.Add([]byte("\xef\xbb\xbfa 1 b c\r\n\r\nC:x\r\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		r, err := ParseRecord("a-b", data)
		if err != nil {
			return
		}
		// Whatever we accept must re-serialize to something we parse back
		// identically (no information lost, no path un-validated).
		r.Warnings = nil
		again, err := ParseRecord("a-b", r.MarshalLegacy())
		if err != nil || !reflect.DeepEqual(r, again) {
			t.Fatalf("not idempotent: %+v vs %+v (%v)", r, again, err)
		}
	})
}

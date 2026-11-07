package plan

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/manifest"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
)

const hNew, hOld = "aaaaaaaa", "bbbbbbbb"

// build makes a manifest from "id:dep,dep" specs. Every component has hash
// hNew, DownloadSize 10, InstallSize 100 and legacy policy flags. Categories
// are only indexed (enough for Resolve), not built into a tree.
func build(specs ...string) *manifest.Manifest {
	m := &manifest.Manifest{Components: map[string]*manifest.Component{}, Categories: map[string]*manifest.Category{}}
	for _, s := range specs {
		id, deps, _ := strings.Cut(s, ":")
		c := &manifest.Component{ID: id, Hash: hNew, DownloadSize: 10, InstallSize: 100}
		if deps != "" {
			c.Depends = strings.Split(deps, ",")
		}
		manifest.DerivePolicy(c)
		m.Components[id] = c
		m.Order = append(m.Order, id)
		parts := strings.Split(id, "-")
		for i := 1; i < len(parts); i++ {
			cid := strings.Join(parts[:i], "-")
			m.Categories[cid] = &manifest.Category{ID: cid}
		}
	}
	return m
}

func baseManifest() *manifest.Manifest {
	return build("core-config", "core-launcher:core-config", "core-database",
		"extra-a:extra-b", "extra-b:extra-c", "extra-c",
		"extra-games-x:core-launcher", "extra-games-y:nowhere-dep", "extra-solo")
}

func rec(id, hash string, size int64, deps []string, files ...string) *state.Record {
	return &state.Record{ID: id, Hash: hash, InstallSize: size, Depends: deps, Files: files}
}

// with returns records for all core components (up to date) plus more.
func with(more ...*state.Record) map[string]*state.Record {
	out := map[string]*state.Record{}
	for _, r := range append([]*state.Record{
		rec("core-config", hNew, 100, nil),
		rec("core-launcher", hNew, 100, []string{"core-config"}),
		rec("core-database", hNew, 100, nil),
	}, more...) {
		out[r.ID] = r
	}
	return out
}

func compIDs(cs []*manifest.Component) []string {
	out := []string{}
	for _, c := range cs {
		out = append(out, c.ID)
	}
	return out
}

func recIDs(rs []*state.Record) []string {
	out := []string{}
	for _, r := range rs {
		out = append(out, r.ID)
	}
	return out
}

func TestStatuses(t *testing.T) {
	m := baseManifest()
	recs := map[string]*state.Record{
		"core-config":   rec("core-config", hNew, 1, nil),
		"core-database": rec("core-database", hOld, 1, nil), // NeverOutdate
		"extra-a":       rec("extra-a", hOld, 1, nil),
		"extra-b":       rec("extra-b", strings.ToUpper(hNew), 1, nil), // ordinal compare
		"extra-c":       nil,                                           // nil counts as absent
		"gone-old":      rec("gone-old", hOld, 1, nil),
	}
	want := map[string]Status{
		"core-config": UpToDate, "core-database": UpToDate, "core-launcher": Missing,
		"extra-a": Outdated, "extra-b": Outdated, "extra-c": NotInstalled, "gone-old": Deprecated,
		"extra-games-x": NotInstalled, "extra-games-y": NotInstalled, "extra-solo": NotInstalled,
	}
	if got := Statuses(m, recs); !reflect.DeepEqual(got, want) {
		t.Errorf("Statuses = %v\nwant %v", got, want)
	}
}

func TestResolve(t *testing.T) {
	type want struct {
		install, update, remove, addedDeps, addedDependents []string
		skipped                                             []Skip
		warnings                                            int
		err                                                 string // substring; "" = success
	}
	tests := []struct {
		name string
		m    *manifest.Manifest
		recs map[string]*state.Record
		req  Request
		want want
	}{
		{"install up to date is a no-op", nil, with(rec("extra-solo", hNew, 100, nil)),
			Request{Install: []string{"extra-solo"}}, want{skipped: []Skip{{"extra-solo", SkipAlreadyInstalled}}}},
		{"install outdated becomes update", nil, with(rec("extra-solo", hOld, 40, nil)),
			Request{Install: []string{"extra-solo"}}, want{update: []string{"extra-solo"}}},
		{"install NeverOutdate with other hash is a no-op", nil, with(rec("core-database", hOld, 1, nil)),
			Request{Install: []string{"core-database"}}, want{skipped: []Skip{{"core-database", SkipAlreadyInstalled}}}},
		{"auto deps transitive, deps first", nil, with(),
			Request{Install: []string{"extra-a"}, AutoDeps: true},
			want{install: []string{"extra-c", "extra-b", "extra-a"}, addedDeps: []string{"extra-b", "extra-c"}}},
		{"missing deps without AutoDeps", nil, with(),
			Request{Install: []string{"extra-a"}}, want{err: "extra-a needs extra-b, which is not installed"}},
		{"installed outdated dep satisfies install", nil, with(rec("extra-c", hOld, 100, nil)),
			Request{Install: []string{"extra-b"}}, want{install: []string{"extra-b"}}},
		{"update with AutoDeps also updates outdated deps", nil, with(rec("extra-b", hOld, 100, nil), rec("extra-c", hOld, 100, nil)),
			Request{Update: []string{"extra-b"}, AutoDeps: true}, want{update: []string{"extra-c", "extra-b"}, addedDeps: []string{"extra-c"}}},
		{"update without AutoDeps leaves outdated deps", nil, with(rec("extra-b", hOld, 100, nil), rec("extra-c", hOld, 100, nil)),
			Request{Update: []string{"extra-b"}}, want{update: []string{"extra-b"}}},
		{"update of not installed / up to date is skipped", nil, with(rec("extra-solo", hNew, 100, nil)),
			Request{Update: []string{"extra-c", "extra-solo", "core-launcher"}},
			want{skipped: []Skip{{"extra-c", SkipNotInstalled}, {"extra-solo", SkipUpToDate}, {"core-launcher", SkipUpToDate}}}},
		{"update all: outdated + missing, deprecated kept, NeverOutdate ignored", nil, map[string]*state.Record{
			"core-config":   rec("core-config", hOld, 50, nil),
			"core-database": rec("core-database", hOld, 100, nil),
			"extra-solo":    rec("extra-solo", hOld, 100, nil),
			"old-thing":     rec("old-thing", hOld, 7, nil),
		}, Request{UpdateAll: true}, want{install: []string{"core-launcher"}, update: []string{"core-config", "extra-solo"}}},
		{"update all + remove deprecated", nil, with(rec("old-thing", hOld, 7, nil), rec("aaa-older", hOld, 7, nil)),
			Request{UpdateAll: true, RemoveDeprecated: true}, want{remove: []string{"old-thing", "aaa-older"}}},
		{"update all + explicit removal: removal wins", nil, with(rec("extra-solo", hOld, 100, nil)),
			Request{UpdateAll: true, Remove: []string{"extra-solo"}}, want{remove: []string{"extra-solo"}}},
		{"category expansion + dangling dep warning", nil, with(),
			Request{Install: []string{"extra-games"}, AutoDeps: true}, want{install: []string{"extra-games-x", "extra-games-y"}, warnings: 1}},
		{"nested category through root category", nil,
			with(rec("extra-a", hNew, 1, nil), rec("extra-b", hNew, 1, nil), rec("extra-c", hNew, 1, nil)),
			Request{Install: []string{"extra"}}, want{
				install:  []string{"extra-games-x", "extra-games-y", "extra-solo"},
				skipped:  []Skip{{"extra-a", SkipAlreadyInstalled}, {"extra-b", SkipAlreadyInstalled}, {"extra-c", SkipAlreadyInstalled}},
				warnings: 1,
			}},
		{"unknown ids", nil, with(), Request{Install: []string{"nope", "extra-solo"}, Remove: []string{"old-thing"}},
			want{err: `component or category "nope", "old-thing" does not exist`}},
		{"remove required", nil, with(), Request{Remove: []string{"core"}},
			want{err: `cannot remove required component "core-config", "core-launcher", "core-database"`}},
		{"remove depended-upon without cascade", nil, with(rec("extra-b", hNew, 1, nil), rec("extra-c", hNew, 1, nil)),
			Request{Remove: []string{"extra-c"}}, want{err: "extra-c is needed by extra-b"}},
		{"remove with cascade, dependents first", nil,
			with(rec("extra-a", hNew, 1, nil), rec("extra-b", hNew, 2, nil), rec("extra-c", hNew, 4, nil)),
			Request{Remove: []string{"extra-c"}, CascadeRemove: true},
			want{remove: []string{"extra-a", "extra-b", "extra-c"}, addedDependents: []string{"extra-b", "extra-a"}}},
		// A stale record (e.g. a renamed dependency): for a manifest component only manifest deps count.
		{"stale record deps of a manifest component are ignored", nil,
			with(rec("extra-solo", hNew, 1, []string{"extra-c"}), rec("extra-c", hNew, 1, nil)),
			Request{Remove: []string{"extra-c"}}, want{remove: []string{"extra-c"}}},
		{"dependents from deprecated record deps", nil, with(rec("old-thing", hOld, 1, []string{"extra-c"}), rec("extra-c", hNew, 1, nil)),
			Request{Remove: []string{"extra-c"}}, want{err: "extra-c is needed by old-thing"}},
		{"cascade into deprecated record", nil, with(rec("old-thing", hOld, 1, []string{"extra-c"}), rec("extra-c", hNew, 1, nil)),
			Request{Remove: []string{"extra-c"}, CascadeRemove: true},
			want{remove: []string{"old-thing", "extra-c"}, addedDependents: []string{"old-thing"}}},
		{"cascade into required component", build("core-config:extra-solo", "extra-solo"),
			map[string]*state.Record{"core-config": rec("core-config", hNew, 1, nil), "extra-solo": rec("extra-solo", hNew, 1, nil)},
			Request{Remove: []string{"extra-solo"}, CascadeRemove: true}, want{err: `cannot remove required component "core-config"`}},
		{"remove not installed is skipped", nil, with(), Request{Remove: []string{"extra-solo"}},
			want{skipped: []Skip{{"extra-solo", SkipNotInstalled}}}},
		{"remove deprecated by id", nil, with(rec("old-thing", hOld, 5, nil)),
			Request{Remove: []string{"old-thing"}}, want{remove: []string{"old-thing"}}},
		{"install and remove same component", nil, with(rec("extra-solo", hOld, 1, nil)),
			Request{Install: []string{"extra-solo"}, Remove: []string{"extra-solo"}}, want{err: "both installed and removed"}},
		{"install something needing a removed dep", nil, with(rec("extra-c", hNew, 1, nil)),
			Request{Install: []string{"extra-b"}, Remove: []string{"extra-c"}, CascadeRemove: true},
			want{err: "cannot remove extra-c: extra-b is being installed"}},
		{"dependency cycle terminates", build("extra-p:extra-q", "extra-q:extra-p"), nil,
			Request{Install: []string{"extra-p"}, AutoDeps: true}, want{install: []string{"extra-q", "extra-p"}, addedDeps: []string{"extra-q"}}},
		{"reinstall: up to date, outdated, NeverOutdate and not installed", nil,
			with(rec("extra-solo", hNew, 100, nil), rec("extra-c", hOld, 100, nil)),
			Request{Reinstall: []string{"extra-solo", "extra-c", "core-database", "extra-b"}},
			want{install: []string{"extra-b"}, update: []string{"core-database", "extra-c", "extra-solo"}}},
		{"reinstall category, deps of an installed one are satisfied", nil,
			with(rec("extra-a", hNew, 1, nil), rec("extra-b", hNew, 1, nil), rec("extra-c", hNew, 1, nil)),
			Request{Reinstall: []string{"extra-a"}}, want{update: []string{"extra-a"}}},
		{"reinstall deprecated is unknown", nil, with(rec("old-thing", hOld, 1, nil)),
			Request{Reinstall: []string{"old-thing"}}, want{err: `"old-thing" does not exist`}},
		{"reinstall and remove same component", nil, with(rec("extra-solo", hNew, 1, nil)),
			Request{Reinstall: []string{"extra-solo"}, Remove: []string{"extra-solo"}}, want{err: "both installed and removed"}},
		{"empty request", nil, with(), Request{}, want{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := tt.m
			if m == nil {
				m = baseManifest()
			}
			p, err := Resolve(m, tt.recs, tt.req)
			if tt.want.err != "" {
				var pe *Error
				if !errors.As(err, &pe) || !strings.Contains(err.Error(), tt.want.err) {
					t.Fatalf("err = %v, want *Error containing %q", err, tt.want.err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve: %v", err)
			}
			got := want{compIDs(p.Install), compIDs(p.Update), recIDs(p.Remove), p.AddedDeps, p.AddedDependents, p.Skipped, len(p.Warnings), ""}
			w := tt.want
			for _, s := range []*[]string{&w.install, &w.update, &w.remove} {
				if *s == nil {
					*s = []string{}
				}
			}
			if !reflect.DeepEqual(got, w) {
				t.Errorf("got  %+v\nwant %+v", got, w)
			}
			if p.Empty() != (len(w.install)+len(w.update)+len(w.remove) == 0) {
				t.Errorf("Empty() = %v", p.Empty())
			}
		})
	}
}

// The real dangling dependency supportpack-babyz -> supportpack-netscape
// (no longer in the list) on an old install that still has a
// supportpack-netscape record, and whose babyz record lists it.
func TestResolveDanglingDepOfDeprecated(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "testdata", "manifests", "stable.xml"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	m, err := manifest.Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	babyz := m.Components["supportpack-babyz"]
	if m.Components["supportpack-netscape"] != nil || !slices.Equal(babyz.Depends, []string{"supportpack-netscape"}) {
		t.Fatal("snapshot changed")
	}
	recs := map[string]*state.Record{
		"supportpack-babyz":    rec("supportpack-babyz", babyz.Hash, babyz.InstallSize, []string{"supportpack-netscape"}),
		"supportpack-netscape": rec("supportpack-netscape", hOld, 5, nil),
	}
	for _, c := range m.Components {
		if c.Required {
			recs[c.ID] = rec(c.ID, c.Hash, c.InstallSize, c.Depends)
		}
	}
	for _, req := range []Request{
		{UpdateAll: true, RemoveDeprecated: true},
		{UpdateAll: true, RemoveDeprecated: true, CascadeRemove: true, AutoDeps: true},
		{Remove: []string{"supportpack-netscape"}, CascadeRemove: true},
	} {
		p, err := Resolve(m, recs, req)
		if err != nil {
			t.Fatalf("%+v: %v", req, err)
		}
		if got := recIDs(p.Remove); !slices.Equal(got, []string{"supportpack-netscape"}) || len(p.AddedDependents)+len(p.Install)+len(p.Update) != 0 {
			t.Errorf("%+v: Remove %v, AddedDependents %v, Install %v, Update %v", req, got, p.AddedDependents, compIDs(p.Install), compIDs(p.Update))
		}
	}
}

func TestResolveSizes(t *testing.T) {
	m := baseManifest()
	m.Components["extra-solo"].DownloadSize = 7
	// Metadata-only: legacy decides on install-size, so download-size is not counted.
	m.Components["extra-c"].InstallSize, m.Components["extra-c"].DownloadSize = 0, 160
	recs := with(rec("extra-solo", hOld, 40, nil), rec("old-thing", hOld, 30, nil))
	p, err := Resolve(m, recs, Request{Install: []string{"extra-c"}, Update: []string{"extra-solo"}, Remove: []string{"old-thing"}})
	if err != nil {
		t.Fatal(err)
	}
	if p.DownloadBytes != 7 || p.DiskDelta != 0+100-40-30 {
		t.Errorf("DownloadBytes = %d, DiskDelta = %d", p.DownloadBytes, p.DiskDelta)
	}
	if _, err := Resolve(nil, nil, Request{}); err == nil {
		t.Error("nil manifest accepted")
	}
}

func TestFilesToDelete(t *testing.T) {
	m := build("extra-a", "extra-b", "extra-c", "extra-d")
	recs := map[string]*state.Record{
		"extra-a": rec("extra-a", hOld, 1, nil, "a/1", "shared/x", "a/old"),
		"extra-b": rec("extra-b", hNew, 1, nil, "b/1", "Shared/X"),
		"extra-c": rec("extra-c", hNew, 1, nil, "c/1", "c/2"),
	}
	tests := []struct {
		name   string
		req    Request
		staged map[string][]string
		want   []string
	}{
		{"update drops stale files, keeps shared and restaged", Request{Update: []string{"extra-a"}, Remove: []string{"extra-c"}},
			map[string][]string{"extra-a": {"A/1", "a/new"}}, []string{"c/1", "c/2", "a/old", "Components/extra-c"}},
		{"remove keeps files still owned by others", Request{Remove: []string{"extra-b"}}, nil,
			[]string{"b/1", "Components/extra-b"}},
		{"file taken over by a new install is kept", Request{Install: []string{"extra-d"}, Remove: []string{"extra-c"}},
			map[string][]string{"extra-d": {"C/1"}}, []string{"c/2", "Components/extra-c"}},
		// Remove is in reverse dependency order, so extra-b (later in the manifest) goes first.
		{"remove both owners of a shared file", Request{Remove: []string{"extra-a", "extra-b"}}, nil,
			[]string{"b/1", "Shared/X", "a/1", "a/old", "Components/extra-b", "Components/extra-a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := Resolve(m, recs, tt.req)
			if err != nil {
				t.Fatal(err)
			}
			if got := FilesToDelete(recs, p, tt.staged); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FilesToDelete = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestResolveRandomDAG checks closure, order and cascade invariants on
// random acyclic manifests (document order unrelated to dependency order).
func TestResolveRandomDAG(t *testing.T) {
	for seed := uint64(0); seed < 200; seed++ {
		rng := rand.New(rand.NewPCG(seed, 99))
		n := 2 + rng.IntN(30)
		specs := make([]string, n)
		for i := range specs {
			var deps []string
			for j := 0; j < i; j++ { // deps only on lower indices => acyclic
				if rng.IntN(6) == 0 {
					deps = append(deps, fmt.Sprintf("extra-c%d", j))
				}
			}
			specs[i] = fmt.Sprintf("extra-c%d:%s", i, strings.Join(deps, ","))
		}
		rng.Shuffle(n, func(i, j int) { specs[i], specs[j] = specs[j], specs[i] })
		m := build(specs...)
		recs := map[string]*state.Record{}
		var want, installed []string
		for _, id := range m.Order {
			if rng.IntN(3) == 0 {
				recs[id] = rec(id, []string{hNew, hOld}[rng.IntN(2)], 50, m.Components[id].Depends)
				installed = append(installed, id)
			}
			if rng.IntN(4) == 0 {
				want = append(want, id)
			}
		}

		p, err := Resolve(m, recs, Request{Install: want, AutoDeps: true})
		if err != nil {
			t.Fatalf("seed %d: %v", seed, err)
		}
		after, seen := map[string]bool{}, map[string]bool{}
		for id := range recs {
			after[id] = true
		}
		for _, list := range [][]*manifest.Component{p.Install, p.Update} {
			for _, c := range list {
				if seen[c.ID] {
					t.Fatalf("seed %d: duplicate %s", seed, c.ID)
				}
				seen[c.ID] = true
				for _, d := range c.Depends { // deps come first and are present
					if !after[d] {
						t.Fatalf("seed %d: %s before or without its dep %s", seed, c.ID, d)
					}
				}
				after[c.ID] = true
				if recs[c.ID] == nil && !slices.Contains(want, c.ID) && !slices.Contains(p.AddedDeps, c.ID) {
					t.Fatalf("seed %d: %s installed for no reason", seed, c.ID)
				}
			}
		}

		if len(installed) == 0 {
			continue
		}
		p, err = Resolve(m, recs, Request{Remove: []string{installed[rng.IntN(len(installed))]}, CascadeRemove: true})
		if err != nil {
			t.Fatalf("seed %d: remove: %v", seed, err)
		}
		removed := map[string]int{}
		for i, r := range p.Remove {
			removed[r.ID] = i
		}
		for id, r := range recs {
			for _, d := range r.Depends {
				if j, gone := removed[d]; gone {
					if i, ok := removed[id]; !ok || i > j {
						t.Fatalf("seed %d: %s not removed before its dep %s", seed, id, d)
					}
				}
			}
		}
	}
}

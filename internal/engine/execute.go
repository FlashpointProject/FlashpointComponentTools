package engine

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/fetch"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/manifest"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/plan"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/txn"
)

// Decision is a UI's answer to a failed download.
type Decision int

const (
	Abort Decision = iota // stop; the install stays unchanged
	Retry                 // download it again
	Skip                  // drop this component (and what depends on it) and go on
)

// EventKind says what an Event reports.
type EventKind int

const (
	DownloadStarted  EventKind = iota // ID
	DownloadProgress                  // ID, Done, Total
	DownloadRetrying                  // ID, Attempt, Wait, Err
	Extracting                        // ID
	Committing                        // moving files into place; no longer cancellable
)

// Event is a progress report from Execute.
type Event struct {
	Kind        EventKind
	ID          string
	Done, Total int64
	Attempt     int
	Wait        time.Duration
	Err         error
}

// Hooks connect Execute to a UI. Every field may be nil. The engine calls
// them one at a time (never concurrently, although downloads run in
// parallel), so they need no locking; Progress should return quickly.
type Hooks struct {
	Progress func(Event)
	// OnError is asked after a download failed for good (fetch already
	// retried with backoff). nil answers Abort. Not called for cancellation.
	OnError func(ctx context.Context, id string, err error) Decision
	// OnLocked is asked during commit when path (FP-relative) is in use, e.g.
	// Launcher/Flashpoint.exe still running. true waits a second and tries
	// again; false (or nil) rolls the whole change back.
	OnLocked func(ctx context.Context, path string) bool
}

// serialized fills in nil hooks and makes Progress safe to call from the
// parallel downloads. OnError and OnLocked only run while nothing else does.
func (h Hooks) serialized() Hooks {
	progress := h.Progress
	if progress == nil {
		progress = func(Event) {}
	}
	var mu sync.Mutex
	h.Progress = func(ev Event) {
		mu.Lock()
		defer mu.Unlock()
		progress(ev)
	}
	if h.OnError == nil {
		h.OnError = func(context.Context, string, error) Decision { return Abort }
	}
	if h.OnLocked == nil {
		h.OnLocked = func(context.Context, string) bool { return false }
	}
	return h
}

// Result is what Execute did.
type Result struct {
	Installed, Updated, Removed []string
	// Skipped lists components dropped after OnError answered Skip, plus the
	// components of the plan that depend on them. The rest was applied.
	Skipped []string
}

var (
	ErrDiskSpace = errors.New("not enough free disk space")
	ErrAborted   = errors.New("aborted") // OnError answered Abort; the install is unchanged
)

// Execute applies p (from Plan, for the current Snapshot) and needs Lock:
// check disk space, download and verify every archive into Temp/fpm/cache
// (the install is untouched; cancelling keeps .part files for resuming),
// then stage everything into one transaction and commit it. On error the
// install is unchanged, unless the error wraps txn.ErrRollbackFailed (the
// next Lock finishes the rollback). Archives are deleted only after success.
func (e *Engine) Execute(ctx context.Context, p *plan.Plan, h Hooks) (*Result, error) {
	res := &Result{}
	if e.lock == nil || e.snap == nil || p == nil {
		return res, errors.New("engine: Execute needs Lock, a loaded manifest and a plan")
	}
	hk := h.serialized()
	changing := append(slices.Clone(p.Install), p.Update...)
	if err := e.checkDiskSpace(changing); err != nil {
		return res, err
	}
	zips, skipped, err := e.download(ctx, changing, hk)
	if err != nil {
		return res, err
	}
	res.Skipped = skipped
	dropped := func(c *manifest.Component) bool { return slices.Contains(skipped, c.ID) }
	final := &plan.Plan{
		Install: slices.DeleteFunc(slices.Clone(p.Install), dropped),
		Update:  slices.DeleteFunc(slices.Clone(p.Update), dropped),
		Remove:  p.Remove,
	}
	if err := e.apply(ctx, final, slices.DeleteFunc(changing, dropped), zips, hk); err != nil {
		return res, err
	}
	for _, c := range final.Install {
		res.Installed = append(res.Installed, c.ID)
	}
	for _, c := range final.Update {
		res.Updated = append(res.Updated, c.ID)
	}
	for _, r := range final.Remove {
		res.Removed = append(res.Removed, r.ID)
	}
	for _, z := range zips {
		os.Remove(z)
	}
	e.log.Info("changes applied", "installed", res.Installed, "updated", res.Updated, "removed", res.Removed, "skipped", res.Skipped)
	return res, nil
}

func (e *Engine) archiveCache() string { return filepath.Join(e.root, "Temp", "fpm", "cache") }

func (e *Engine) item(c *manifest.Component) (fetch.Item, error) {
	if c.DownloadSize <= 0 {
		return fetch.Item{}, fmt.Errorf("the component list gives no download size for %s, so its download cannot be verified", c.ID)
	}
	return fetch.Item{ID: c.ID, Name: c.ZipName(), Size: c.DownloadSize, CRC32: c.CRC32, SHA256: c.SHA256, CacheBust: c.Hash}, nil
}

// checkDiskSpace needs room for the archives not yet cached plus the
// extracted files (staged next to the install), plus 5%.
func (e *Engine) checkDiskSpace(cs []*manifest.Component) error {
	var need int64
	for _, c := range cs {
		need += c.InstallSize
		if !c.HasArchive() {
			continue
		}
		if it, err := e.item(c); err == nil {
			if name, err := fetch.CacheName(it); err == nil {
				if fi, err := os.Stat(filepath.Join(e.archiveCache(), name)); err == nil && fi.Size() == it.Size {
					continue
				}
			}
		}
		need += c.DownloadSize
	}
	need += need / 20
	free, err := freeBytes(e.root)
	if err != nil {
		e.log.Warn("could not determine free disk space", "err", err)
		return nil
	}
	if need > 0 && uint64(need) > free {
		return fmt.Errorf("%w: this needs about %s, but only %s is free on the drive with %s. Free up some space and try again.",
			ErrDiskSpace, FormatSize(need), FormatSize(int64(min(free, 1<<62))), e.root)
	}
	return nil
}

// mirrors returns the manifest's base URL, plus the twin Nexus host (same
// server, once direct and once behind Cloudflare; DESIGN.md §1.3).
func mirrors(base string) []string {
	twins := map[string]string{
		"nexus-dev.unstable.life":         "nexus-dev.flashpointarchive.org",
		"nexus-dev.flashpointarchive.org": "nexus-dev.unstable.life",
	}
	u, err := url.Parse(base)
	if err != nil || twins[strings.ToLower(u.Host)] == "" {
		return []string{base}
	}
	u.Host = twins[strings.ToLower(u.Host)]
	return []string{base, u.String()}
}

// download fetches the archives of cs (fetch.DefaultParallel at a time),
// then asks OnError about each failure. It returns component ID -> zip path
// and the skipped IDs (with their dependents in cs).
func (e *Engine) download(ctx context.Context, cs []*manifest.Component, hk Hooks) (map[string]string, []string, error) {
	client := e.client(mirrors(e.snap.Manifest.BaseURL))
	fetchOne := func(c *manifest.Component) (string, error) {
		it, err := e.item(c)
		if err != nil {
			return "", err
		}
		hk.Progress(Event{Kind: DownloadStarted, ID: c.ID})
		return client.FetchArchive(ctx, it, e.archiveCache(),
			func(p fetch.Progress) {
				hk.Progress(Event{Kind: DownloadProgress, ID: c.ID, Done: p.Done, Total: p.Total})
			},
			func(r fetch.RetryInfo) {
				hk.Progress(Event{Kind: DownloadRetrying, ID: c.ID, Attempt: r.Attempt, Wait: r.Wait, Err: r.Err})
			})
	}

	var archives []*manifest.Component
	for _, c := range cs {
		if c.HasArchive() {
			archives = append(archives, c)
		}
	}
	paths := make([]string, len(archives))
	errs := make([]error, len(archives))
	sem := make(chan struct{}, fetch.DefaultParallel)
	var wg sync.WaitGroup
	for i, c := range archives {
		if ctx.Err() != nil {
			break
		}
		sem <- struct{}{}
		wg.Go(func() {
			defer func() { <-sem }()
			paths[i], errs[i] = fetchOne(c)
		})
	}
	wg.Wait()
	if err := ctx.Err(); err != nil {
		return nil, nil, fmt.Errorf("download cancelled (partial downloads are kept and resume next time): %w", err)
	}

	zips := map[string]string{}
	var skipped []string
	for i, c := range archives {
		for errs[i] != nil && !slices.Contains(skipped, c.ID) {
			if err := ctx.Err(); err != nil {
				return nil, nil, err
			}
			switch hk.OnError(ctx, c.ID, errs[i]) {
			case Retry:
				paths[i], errs[i] = fetchOne(c)
			case Skip:
				e.log.Warn("skipping component", "id", c.ID, "because", errs[i])
				skipped = append(skipped, c.ID)
				skipped = withDependents(cs, skipped)
			default:
				return nil, nil, fmt.Errorf("%w: %w", ErrAborted, errs[i])
			}
		}
		if errs[i] == nil {
			zips[c.ID] = paths[i]
		}
	}
	return zips, skipped, nil
}

// withDependents adds to skip every component of cs that depends on a
// skipped one, transitively.
func withDependents(cs []*manifest.Component, skip []string) []string {
	for changed := true; changed; {
		changed = false
		for _, c := range cs {
			if !slices.Contains(skip, c.ID) && slices.ContainsFunc(c.Depends, func(d string) bool { return slices.Contains(skip, d) }) {
				skip = append(skip, c.ID)
				changed = true
			}
		}
	}
	return skip
}

// apply stages everything into one transaction and commits it.
func (e *Engine) apply(ctx context.Context, p *plan.Plan, changing []*manifest.Component, zips map[string]string, hk Hooks) error {
	tx, err := txn.Begin(e.root, txn.Options{Log: e.log, LaterArchiveWins: true})
	if err != nil {
		return err
	}
	defer tx.Abort()

	staged := map[string][]string{}
	for _, c := range changing {
		// Metadata-only components (install-size 0) get a header-only record,
		// like the C# Manager.
		if c.HasArchive() {
			if staged[c.ID], err = e.stage(ctx, tx, c, zips[c.ID], hk); err != nil {
				return err
			}
		}
	}

	// Delete files no component lists any more and the records of removed
	// components, but never preserved user files.
	recs, errs := state.ReadRecords(e.root)
	if len(errs) > 0 {
		e.log.Warn("unreadable records are ignored for shared-file counting", "errs", errors.Join(errs...))
	}
	for _, rel := range plan.FilesToDelete(recs, p, staged) {
		if isPreserved(rel) {
			e.log.Info("keeping user file", "path", rel)
		} else if err := tx.Remove(rel); err != nil {
			return err
		}
	}

	for _, c := range changing {
		rec := state.Record{ID: c.ID, Hash: c.Hash, InstallSize: c.InstallSize, Depends: c.Depends, Files: staged[c.ID]}
		if err := tx.StageFile(plan.RecordPath(c.ID), rec.MarshalLegacy()); err != nil {
			return err
		}
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	hk.Progress(Event{Kind: Committing})
	return tx.Commit(ctx, func(f txn.LockedFile) txn.Decision {
		if hk.OnLocked(ctx, f.Path) {
			return txn.Retry
		}
		return txn.Abort
	})
}

// stage extracts c's archive into tx, refuses reserved paths, and keeps
// preserved user files that already exist.
func (e *Engine) stage(ctx context.Context, tx *txn.Tx, c *manifest.Component, zipPath string, hk Hooks) ([]string, error) {
	hk.Progress(Event{Kind: Extracting, ID: c.ID})
	files, err := tx.StageArchive(ctx, zipPath, c.Path, txn.LimitsForInstallSize(c.InstallSize), nil)
	if err != nil {
		if errors.Is(err, zip.ErrChecksum) || errors.Is(err, zip.ErrFormat) {
			os.Remove(zipPath) // corrupt although its hash matched: download it again next time
		}
		return nil, fmt.Errorf("%s: %w", c.ID, err)
	}
	for _, rel := range files {
		if reserved(rel) {
			return nil, fmt.Errorf("%w: component %s would overwrite %s, which belongs to the component manager; refusing to install it",
				ErrPolicy, c.ID, rel)
		}
		if isPreserved(rel) {
			if _, err := os.Lstat(filepath.Join(e.root, filepath.FromSlash(rel))); err == nil && tx.Unstage(rel) {
				e.log.Info("keeping existing user file", "id", c.ID, "path", rel)
			}
		}
	}
	return files, nil
}

// FormatSize formats a byte count like legacy fpm's FormatBytes
// (fpm src/Helpers.cs:405-417): binary units B/KB/MB/GB, one decimal,
// rounded half to even, thousands separators (.NET en-US "N1").
func FormatSize(n int64) string {
	units := []string{"B", "KB", "MB", "GB"}
	i := len(units) - 1
	for i > 0 && math.Abs(float64(n)) < float64(int64(1)<<(10*i)) {
		i--
	}
	s := fmt.Sprintf("%.1f", math.RoundToEven(float64(n)/float64(int64(1)<<(10*i))*10)/10)
	sign, s := "", strings.TrimPrefix(s, "-")
	if n < 0 {
		sign = "-"
	}
	intPart, frac, _ := strings.Cut(s, ".")
	for j := len(intPart) - 3; j > 0; j -= 3 {
		intPart = intPart[:j] + "," + intPart[j:]
	}
	return sign + intPart + "." + frac + units[i]
}

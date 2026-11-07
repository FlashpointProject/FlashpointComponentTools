// Package engine is the layer between the core packages (manifest, state,
// plan, fetch, txn, trust) and a user interface (the fpm CLI now, the GUI
// later). It owns the order of operations and the safety rules:
//
//	Open -> LoadManifest -> Plan -> Lock -> Execute
//
// Read-only use (list, info) needs no lock. Execute downloads and verifies
// everything before touching the install, then stages and commits through
// one txn transaction: the install ends up either fully old or fully new.
package engine

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"maps"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/fetch"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/manifest"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/plan"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/trust"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/txn"
)

// Options configure Open.
type Options struct {
	// ManagerDir is the folder of the running executable (normally
	// <FP>/Manager). fpm.cfg is read from there and a relative root resolves
	// against it, not against the working directory.
	ManagerDir     string
	RootOverride   string // --path; fpm.cfg is not changed
	SourceOverride string // --source
	Logger         *slog.Logger
	UserAgent      string

	Fetch fetch.Options     // template for download clients (tests tune backoff)
	Keys  []trust.PublicKey // signing keys; nil = trust.Keys(channel) (tests override)
}

// Engine is one opened Flashpoint install. It is not safe for concurrent use.
type Engine struct {
	opts    Options
	log     *slog.Logger
	root    string
	source  string
	channel string
	notices []string // from Open, repeated as Snapshot warnings
	lock    *os.File
	snap    *Snapshot
}

// ErrBusy: another manager instance holds the install lock.
var ErrBusy = errors.New("another Flashpoint Manager (or fpm) is running on this Flashpoint folder; close it and try again")

var errLockHeld = errors.New("lock held")

// ErrRootNotFound is wrapped by Open when there is no Flashpoint folder: the
// configured one and the Manager folder's parent are not Flashpoint folders,
// or the RootOverride folder does not exist.
var ErrRootNotFound = state.ErrRootNotFound

// Open locates the Flashpoint folder and reads fpm.cfg; no network I/O.
// Root: --path, else fpm.cfg line 1 relative to the Manager folder, else the
// Manager folder's parent. A stale fpm.cfg is repaired to `..\` when the
// parent is valid, so the fix sticks for the launcher and the old tools;
// that is the only thing Open writes.
func Open(ctx context.Context, opts Options) (*Engine, error) {
	if opts.ManagerDir == "" {
		return nil, errors.New("engine: ManagerDir is required")
	}
	mdir, err := filepath.Abs(opts.ManagerDir)
	if err != nil {
		return nil, err
	}
	log := opts.Logger
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	cfg, err := state.LoadConfig(mdir)
	if err != nil {
		return nil, err
	}
	e := &Engine{opts: opts, log: log, source: cfg.Source}

	if opts.RootOverride != "" {
		if e.root, err = filepath.Abs(opts.RootOverride); err != nil {
			return nil, err
		}
		if fi, err := os.Stat(e.root); err != nil || !fi.IsDir() {
			return nil, fmt.Errorf("%w: %s does not exist", ErrRootNotFound, e.root)
		}
	} else {
		root, repaired, err := state.FindRoot(mdir, cfg)
		if err != nil {
			return nil, fmt.Errorf("%w. Set the folder with \"fpm path <folder>\" or --path", err)
		}
		e.root = root
		if repaired {
			msg := fmt.Sprintf("fpm.cfg pointed at %q, which is not a Flashpoint folder; using %s", cfg.RootRaw, root)
			fixed := *cfg
			fixed.RootRaw = `..\`
			if err := state.SaveConfig(mdir, &fixed); err != nil {
				log.Warn("could not repair fpm.cfg", "err", err)
			} else {
				msg += ` and saved it as "..\" in fpm.cfg`
			}
			log.Warn(msg)
			e.notices = append(e.notices, msg)
		}
	}

	if opts.SourceOverride != "" {
		e.source = opts.SourceOverride
	}
	u, err := ValidSource(e.source)
	if err != nil {
		return nil, fmt.Errorf("%w; set it with \"fpm source <url>\"", err)
	}
	e.channel = channelOf(u)
	log.Debug("opened", "root", e.root, "manager", mdir, "source", e.source, "channel", e.channel)
	return e, nil
}

// ValidSource checks a component list URL: http(s) with a host. (http is
// accepted for legacy configs such as http://localhost/...; the GUI only
// saves https, see SaveSettings.)
func ValidSource(s string) (*url.URL, error) {
	u, err := url.Parse(s)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Host == "" {
		return nil, fmt.Errorf("the component list URL %q is not a valid http(s) URL", s)
	}
	return u, nil
}

var channelRE = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,63}$`)

// channelOf names the signing channel of a manifest URL: the folder holding
// components.xml (".../repository/stable/components.xml" -> "stable"), or
// "custom" when that is not a plain name.
func channelOf(u *url.URL) string {
	if dir := path.Base(path.Dir(u.Path)); channelRE.MatchString(dir) {
		return dir
	}
	return "custom"
}

func (e *Engine) client(mirrors []string) *fetch.Client {
	o := e.opts.Fetch
	o.Mirrors, o.UserAgent, o.Logger = mirrors, e.opts.UserAgent, e.log
	if e.snap != nil && e.snap.Origin != FromNetwork {
		// The server was unreachable a moment ago: let the UI ask at the
		// first archive failure instead of retrying for a minute.
		o.RetryBudget = time.Nanosecond
	}
	return fetch.New(o)
}

func (e *Engine) Root() string   { return e.root }
func (e *Engine) Source() string { return e.source }

// Channel is the signing channel of the source: "stable", "development",
// another plain folder name, or "custom".
func (e *Engine) Channel() string { return e.channel }

// SaveSettings validates and writes fpm.cfg in managerDir (the GUI's
// Settings page). An empty value keeps the current one. root must be a
// Flashpoint folder; it is stored as `..\` when it is the Manager folder's
// parent (survives moving the install, DESIGN.md §4.8), else absolute.
// source must be an https URL, unless it is the value already configured
// (legacy configs may hold http://localhost/...). Engines opened before keep
// their old settings: open a new one.
func SaveSettings(managerDir, root, source string) error {
	mdir, err := filepath.Abs(managerDir)
	if err != nil {
		return err
	}
	cfg, err := state.LoadConfig(mdir)
	if err != nil {
		return err
	}
	if root = strings.TrimSpace(root); root != "" {
		if root, err = filepath.Abs(root); err != nil {
			return err
		}
		if !state.IsRoot(root) {
			return fmt.Errorf(`%s is not a Flashpoint folder (it has neither Components\ records nor Launcher\ and Data\ folders)`, root)
		}
		cfg.RootRaw = root
		if sameDir(root, filepath.Dir(mdir)) {
			cfg.RootRaw = `..\`
		}
	}
	if source = strings.TrimSpace(source); source != "" && source != cfg.Source {
		u, err := url.Parse(source)
		if err != nil || u.Scheme != "https" || u.Host == "" {
			return fmt.Errorf("the component list URL %q is not a valid https URL", source)
		}
		cfg.Source = source
	}
	return state.SaveConfig(mdir, cfg)
}

// sameDir reports whether a and b are the same existing folder (this also
// handles case differences on Windows and symlinks).
func sameDir(a, b string) bool {
	fa, err1 := os.Stat(a)
	fb, err2 := os.Stat(b)
	return err1 == nil && err2 == nil && os.SameFile(fa, fb)
}

// Lock takes the single-instance lock (Components/.fpm/lock) that every
// change needs, then finishes or rolls back a transaction a crash
// interrupted; recovered reports whether there was one. Close releases it.
func (e *Engine) Lock() (recovered bool, err error) {
	if e.lock != nil {
		return false, nil
	}
	priv, err := state.PrivateDir(e.root)
	if err != nil {
		return false, fmt.Errorf("cannot write to the Flashpoint folder: %w", err)
	}
	f, err := os.OpenFile(filepath.Join(priv, "lock"), os.O_RDWR|os.O_CREATE, 0o644)
	if err != nil {
		return false, fmt.Errorf("create lock file: %w", err)
	}
	if err := lockFile(f); err != nil {
		f.Close()
		if errors.Is(err, errLockHeld) {
			return false, ErrBusy
		}
		return false, fmt.Errorf("lock %s: %w", f.Name(), err)
	}
	e.lock = f
	if recovered, err = txn.Recover(e.root, e.log); err != nil {
		return recovered, fmt.Errorf("an interrupted update could not be undone: %w", err)
	}
	return recovered, nil
}

// Close releases the lock. The Engine must not be used afterwards.
func (e *Engine) Close() error {
	if e.lock == nil {
		return nil
	}
	err := errors.Join(unlockFile(e.lock), e.lock.Close())
	e.lock = nil
	return err
}

// LoadMode says what a loaded manifest will be used for.
type LoadMode int

const (
	// ForReading (list, info) may fall back to components.bak even when it
	// cannot be verified, with a warning.
	ForReading LoadMode = iota
	// ForChanges (download, update, remove) only accepts components.bak
	// while the channel has no signing keys.
	ForChanges
)

// Origin says where a Snapshot's manifest came from.
type Origin int

const (
	FromNetwork Origin = iota // fetched now
	FromBackup                // network failed; Components/components.bak
)

func (o Origin) String() string { return [...]string{"network", "components.bak"}[o] }

// Snapshot is the manifest plus the local state it was compared with.
type Snapshot struct {
	Manifest *manifest.Manifest
	Records  map[string]*state.Record // parsed records; broken ones are in Broken
	// Broken holds the IDs whose record file exists but could not be read
	// or parsed: installed, but in an unknown state.
	Broken   map[string]error
	Trust    trust.Result
	Origin   Origin
	Warnings []string // for the user (stderr in the CLI)
}

func (s *Snapshot) warn(format string, args ...any) {
	s.Warnings = append(s.Warnings, fmt.Sprintf(format, args...))
}

// LoadManifest downloads and verifies the manifest, falling back to
// components.bak when that fails (ctx's deadline counts as a failure,
// cancellation does not), and reads the records.
func (e *Engine) LoadManifest(ctx context.Context, mode LoadMode) (*Snapshot, error) {
	s := &Snapshot{Warnings: slices.Clone(e.notices)}
	keys := e.opts.Keys
	if keys == nil {
		keys = trust.Keys(e.channel)
	}

	body, res, err := e.fetchManifest(ctx, keys)
	if err == nil {
		if s.Manifest, err = e.parse(body); err == nil {
			s.Trust, s.Origin = res, FromNetwork
			// The offline copy, also read by the old tools. Rewritten only
			// when the list changed, so a launcher start that runs
			// `fpm list` writes nothing.
			if old, err := state.ReadBackupManifest(e.root); err != nil || !bytes.Equal(old, body) {
				if err := state.WriteBackupManifest(e.root, body); err != nil {
					e.log.Warn("could not write components.bak", "err", err)
				}
			}
			return e.finish(s), nil
		}
	}
	if errors.Is(err, context.Canceled) {
		return nil, err
	}
	e.log.Warn("component list download failed", "err", err)
	netErr := fmt.Errorf("could not download the component list (%w)", err)

	raw, err := state.ReadBackupManifest(e.root)
	if errors.Is(err, fs.ErrNotExist) {
		return nil, fmt.Errorf("%w, and there is no offline copy", netErr)
	} else if err != nil {
		return nil, fmt.Errorf("%w, and the offline copy is unreadable: %w", netErr, err)
	}
	// components.bak has no signature: with keys it is only good for reading.
	if _, err := trust.Check(keys, e.channel, raw, nil); err != nil {
		if mode == ForChanges {
			return nil, fmt.Errorf("%w, and the offline copy components.bak cannot be trusted for changes: %w", netErr, err)
		}
		s.warn("components.bak is not verified (%v); showing it anyway", err)
	}
	if s.Manifest, err = e.parse(raw); err != nil {
		return nil, fmt.Errorf("%w, and the offline copy is damaged: %w", netErr, err)
	}
	s.Trust, s.Origin = trust.Unsigned, FromBackup
	s.warn("%v; using the offline copy components.bak", netErr)
	return e.finish(s), nil
}

// fetchManifest downloads the manifest and, if there are keys to check it
// with, its signature (Nexus has a request budget), and verifies it.
func (e *Engine) fetchManifest(ctx context.Context, keys []trust.PublicKey) ([]byte, trust.Result, error) {
	c := e.client(nil)
	body, err := c.FetchManifest(ctx, e.source)
	if err != nil {
		return nil, trust.Unsigned, err
	}
	var sig []byte
	if len(keys) > 0 {
		sig, err = c.FetchSmall(ctx, e.source+".minisig", 64<<10)
		if errors.Is(err, fetch.ErrNotFound) {
			err = nil // Check refuses the missing signature
		}
		if err != nil {
			return nil, trust.Unsigned, fmt.Errorf("download signature: %w", err)
		}
	}
	res, err := trust.Check(keys, e.channel, body, sig)
	return body, res, err
}

func (e *Engine) parse(raw []byte) (*manifest.Manifest, error) {
	// Plain http only when the user configured an http source (legacy fpm
	// allowed that, e.g. http://localhost/components.xml).
	insecure := strings.HasPrefix(strings.ToLower(e.source), "http://")
	m, err := manifest.ParseWithOptions(bytes.NewReader(raw), manifest.Options{AllowInsecure: insecure})
	if err != nil {
		return nil, err
	}
	for _, w := range m.Warnings {
		e.log.Info("component list", "warning", w)
	}
	return m, nil
}

func (e *Engine) finish(s *Snapshot) *Snapshot {
	var err error
	if s.Records, s.Broken, err = e.readRecords(); err != nil {
		s.warn("%v", err)
	}
	for _, id := range slices.Sorted(maps.Keys(s.Broken)) {
		s.warn("%v; reinstall the component to repair it", s.Broken[id])
	}
	// Records with skipped (unsafe) path lines still count as installed.
	for _, id := range slices.Sorted(maps.Keys(s.Records)) {
		for _, w := range s.Records[id].Warnings {
			s.warn("%s; reinstall the component to repair it", w)
		}
	}
	e.snap = s
	e.log.Info("component list loaded", "origin", s.Origin, "trust", s.Trust,
		"components", len(s.Manifest.Order), "installed", len(s.Records), "broken", len(s.Broken))
	return s
}

// readRecords reads Components/. A record that exists but cannot be read
// or parsed goes into broken; other errors (Components/ unreadable) are
// returned.
func (e *Engine) readRecords() (map[string]*state.Record, map[string]error, error) {
	recs, errs := state.ReadRecords(e.root)
	broken := map[string]error{}
	var other []error
	for _, err := range errs {
		if re, ok := errors.AsType[*state.RecordError](err); ok {
			broken[re.ID] = err
		} else {
			other = append(other, err)
		}
	}
	return recs, broken, errors.Join(other...)
}

// Plan resolves a request against the last loaded Snapshot.
func (e *Engine) Plan(req plan.Request) (*plan.Plan, error) {
	if e.snap == nil {
		return nil, errors.New("engine: LoadManifest must be called before Plan")
	}
	return plan.Resolve(e.snap.Manifest, e.snap.Records, req)
}

// VerifyResult is the quick check of one installed component.
type VerifyResult struct {
	ID      string
	Missing []string // record files that do not exist (FP-relative)
	Err     error    // the record is unreadable, or the component not installed
}

// Verify checks that every file listed in each record exists (DESIGN.md
// §4.7 "quick"). ids empty means every installed component; results are in
// ID order. It needs neither the manifest nor the lock.
func (e *Engine) Verify(ctx context.Context, ids []string) ([]VerifyResult, error) {
	recs, broken, err := e.readRecords()
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		for id := range recs {
			ids = append(ids, id)
		}
		for id := range broken {
			ids = append(ids, id)
		}
	}
	slices.Sort(ids)
	var out []VerifyResult
	for _, id := range slices.Compact(ids) {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		r := VerifyResult{ID: id, Err: broken[id]}
		if rec := recs[id]; rec != nil {
			for _, f := range rec.Files {
				if _, err := os.Lstat(filepath.Join(e.root, filepath.FromSlash(f))); err != nil {
					r.Missing = append(r.Missing, f)
				}
			}
		} else if r.Err == nil {
			r.Err = fmt.Errorf("component %s is not installed", id)
		}
		out = append(out, r)
	}
	return out, nil
}

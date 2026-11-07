// Package txn applies a set of file changes to a Flashpoint install so that
// a failure or crash leaves it either fully old or fully new (DESIGN.md §4.5).
//
// It only moves files and knows nothing about components. StageArchive and
// StageFile write new files into <root>/Components/.fpm/txn/<id>/new/ (same
// volume, so placing them is a rename); Remove lists files to delete. Commit
// writes a journal, then moves each replaced or removed file into backup/
// and each staged file into place. Any failure before the journal says
// "committed" rolls back from backup/; Recover does the same at startup
// after a crash, or finishes the cleanup of a committed transaction.
package txn

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"strings"
	"time"
)

// Decision is the caller's answer to a file that is in use.
type Decision int

const (
	Retry Decision = iota
	Abort
)

// LockedFile describes a destination that cannot be replaced right now,
// e.g. Launcher/Flashpoint.exe while the launcher is still running.
type LockedFile struct {
	Path    string // FP-relative, slash-separated
	Err     error
	Holders []string // processes holding the file; currently never filled in
}

var (
	ErrNeedsRecovery  = errors.New("an interrupted update needs to be recovered first")
	ErrUnsafeEntry    = errors.New("unsafe archive entry")
	ErrArchiveLimit   = errors.New("archive exceeds size limits")
	ErrConflict       = errors.New("path conflict") // same path twice, or file vs folder
	ErrLocked         = errors.New("file is in use or not writable")
	ErrRollbackFailed = errors.New("could not undo the partial update") // Recover retries it

	errFinished = errors.New("txn: transaction already committed or aborted")
	errCrash    = errors.New("txn: simulated crash") // tests only
)

// LockedAdvice is what to do about ErrLocked, for prompts that name the file
// themselves. Windows reports a permission problem the same way as a file in
// use (ACCESS_DENIED), so it covers both.
const LockedAdvice = "Close Flashpoint and any program using it; if Flashpoint is installed under Program Files, run the manager as administrator."

// txnBase is the FP-relative directory holding one dir per transaction.
const txnBase = "Components/.fpm/txn"

// Options configure a transaction.
type Options struct {
	Log *slog.Logger // nil = discard

	// LaterArchiveWins lets a later StageArchive provide a file an earlier
	// archive already staged (the later copy is placed) instead of failing
	// with ErrConflict: the legacy result of extracting components one after
	// another, and published components do overlap. StageFile files are
	// never replaced.
	LaterArchiveWins bool
}

// Tx is a single transaction. It is not safe for concurrent use; the engine's
// single-instance lock ensures one transaction per install at a time.
type Tx struct {
	root    *os.Root // install root: every path goes through it
	dir     string   // FP-relative txn dir
	log     *slog.Logger
	replace bool // Options.LaterArchiveWins

	files   []*stagedFile          // in staging order
	places  map[string]*stagedFile // lowercased rel -> file that will be placed
	removes []string
	seq     int
	done    bool

	// Test hooks.
	rename      func(oldname, newname string) error
	retryDelay  time.Duration   // wait after onLocked says Retry
	quickDelays []time.Duration // automatic retries of a locked file first
	crash       func() bool     // true = simulate a crash at this point
}

type stagedFile struct {
	rel    string // FP-relative destination
	staged string // FP-relative staged copy
	source string // archive path or "StageFile", for messages
	meta   bool   // from StageFile: placed last, never replaced
}

// Begin starts a transaction against the install at root. It fails with
// ErrNeedsRecovery if an interrupted transaction exists. Leftovers of
// committed transactions (e.g. the backup of an exe that was still running)
// are deleted on the way, best effort.
func Begin(root string, opts Options) (*Tx, error) {
	log := opts.Log
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, fmt.Errorf("open Flashpoint folder: %w", err)
	}
	entries, err := fs.ReadDir(r.FS(), txnBase)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		r.Close()
		return nil, fmt.Errorf("read %s: %w", txnBase, sysErr(err))
	}
	for _, e := range entries {
		dir := path.Join(txnBase, e.Name())
		if j, err := readJournal(r, dir); err == nil && j.State == stateCommitted {
			finishCommitted(r, dir, j, log)
			continue
		}
		r.Close()
		log.Warn("found an interrupted transaction", "dir", dir)
		return nil, ErrNeedsRecovery
	}
	var b [4]byte
	rand.Read(b[:])
	dir := path.Join(txnBase, time.Now().UTC().Format("20060102T150405")+"-"+hex.EncodeToString(b[:]))
	err = r.MkdirAll(dir+"/new", 0o755)
	if err == nil {
		err = r.Mkdir(dir+"/backup", 0o755)
	}
	if err != nil {
		r.Close()
		return nil, fmt.Errorf("create %s: %w", txnBase, sysErr(err))
	}
	return &Tx{
		root: r, dir: dir, log: log, replace: opts.LaterArchiveWins,
		places:      map[string]*stagedFile{},
		rename:      r.Rename,
		retryDelay:  time.Second,
		quickDelays: quickRetryDelays,
		crash:       func() bool { return false },
	}, nil
}

// quickRetryDelays: about 2.75 s of automatic retries before asking, for an
// antivirus or indexer briefly holding a file we just wrote.
var quickRetryDelays = []time.Duration{
	50 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond, 400 * time.Millisecond,
	500 * time.Millisecond, 500 * time.Millisecond, 500 * time.Millisecond, 500 * time.Millisecond,
}

// Abort discards everything staged. After Commit it is a no-op, so
// `defer tx.Abort()` is fine.
func (tx *Tx) Abort() {
	if !tx.done {
		tx.done = true
		tx.discard()
		tx.root.Close()
	}
}

func (tx *Tx) discard() {
	if err := discardTxn(tx.root, tx.dir); err != nil {
		tx.log.Warn("could not delete staging dir; it will be deleted at next start", "dir", tx.dir, "err", err)
	}
}

// Remove schedules the file rel (FP-relative, clean, slash-separated) for
// deletion. A missing file is fine; a folder at rel is left alone. Removals
// run in call order before any placement; removing a staged path is a no-op.
func (tx *Tx) Remove(rel string) error {
	if tx.done {
		return errFinished
	}
	tx.removes = append(tx.removes, rel)
	return nil
}

// Commit applies the transaction. ctx is honoured until the first change to
// the install and while waiting on a locked file. A file in use is retried
// for about 3 s, then onLocked is asked: Retry waits 1 s and tries again,
// Abort (or a nil onLocked) rolls back with ErrLocked. On error the install
// is unchanged, unless the error wraps ErrRollbackFailed: then the
// transaction stays on disk for Recover.
func (tx *Tx) Commit(ctx context.Context, onLocked func(LockedFile) Decision) error {
	if tx.done {
		return errFinished
	}
	tx.done = true
	defer tx.root.Close()

	j, err := tx.prepare()
	if err == nil && len(j.Ops) > 0 {
		err = ctx.Err()
		if err == nil {
			err = writeJournal(tx.root, tx.dir, j)
		}
	}
	if err != nil || len(j.Ops) == 0 {
		tx.discard()
		return err
	}
	if tx.crash() {
		return errCrash
	}
	for i := range j.Ops {
		if err := tx.apply(ctx, &j.Ops[i], onLocked); err != nil {
			return tx.fail(j, err)
		}
	}
	j.State = stateCommitted
	if err := writeJournal(tx.root, tx.dir, j); err != nil {
		return tx.fail(j, err)
	}
	if tx.crash() {
		return errCrash
	}
	finishCommitted(tx.root, tx.dir, j, tx.log)
	return nil
}

// fail rolls back after a failure while committing.
func (tx *Tx) fail(j *journal, cause error) error {
	if errors.Is(cause, errCrash) {
		return cause
	}
	if err := rollback(tx.root, j, tx.rename, tx.log); err != nil {
		return errors.Join(cause, fmt.Errorf("%w: %w (it will be retried at next start)", ErrRollbackFailed, err))
	}
	tx.discard()
	return fmt.Errorf("%w (no changes were made)", cause)
}

// prepare builds the operation list (removes, archive files, then StageFile
// files so records never claim files not yet in place) and checks it against
// the live tree before anything is touched.
func (tx *Tx) prepare() (*journal, error) {
	j := &journal{State: stateCommitting}
	backup := func() string { return fmt.Sprintf("%s/backup/%06d", tx.dir, len(j.Ops)+1) }
	for _, rel := range tx.removes {
		if tx.places[strings.ToLower(rel)] == nil {
			j.Ops = append(j.Ops, journalOp{Kind: opRemove, Rel: rel, Backup: backup()})
		}
	}
	dirOK := map[string]bool{} // lowercased parent folders already checked
	for _, meta := range []bool{false, true} {
		for _, sf := range tx.files {
			if sf.meta != meta || tx.places[strings.ToLower(sf.rel)] != sf {
				continue // other pass, or unstaged/replaced
			}
			// Antivirus may quarantine a staged file. Fail now, not half-way.
			if _, err := tx.root.Lstat(sf.staged); err != nil {
				return nil, fmt.Errorf("the prepared copy of %s (from %s) has disappeared, perhaps quarantined by an antivirus program; nothing was changed, try again: %w", sf.rel, sf.source, sysErr(err))
			}
			fi, err := tx.root.Lstat(sf.rel)
			if err == nil && fi.IsDir() {
				return nil, fmt.Errorf("%w: cannot install file %s because a folder with that name exists; move it away and try again", ErrConflict, sf.rel)
			}
			for d := path.Dir(sf.rel); d != "." && !dirOK[strings.ToLower(d)]; d = path.Dir(d) {
				dirOK[strings.ToLower(d)] = true
				if fi, err := tx.root.Lstat(d); err == nil && !fi.IsDir() {
					return nil, fmt.Errorf("%w: cannot install %s because %s is a file; move it away and try again", ErrConflict, sf.rel, d)
				}
			}
			j.Ops = append(j.Ops, journalOp{Kind: opPlace, Rel: sf.rel, Staged: sf.staged, Backup: backup(), Existed: err == nil})
		}
	}
	return j, nil
}

// apply performs one operation: move an existing destination to backup,
// then (place ops) move the staged file into place.
func (tx *Tx) apply(ctx context.Context, op *journalOp, onLocked func(LockedFile) Decision) error {
	fi, err := tx.root.Lstat(op.Rel)
	switch {
	case err == nil && fi.IsDir() && op.Kind == opRemove:
		tx.log.Warn("not removing a folder listed as a file", "path", op.Rel)
		return nil
	case err == nil || isLocked(err): // on Windows a file open without sharing can't even be stat'ed
		if err := tx.move(ctx, op.Rel, op.Backup, op.Rel, onLocked); err != nil {
			return err
		}
	case !errors.Is(err, fs.ErrNotExist):
		return fmt.Errorf("check %s: %w", op.Rel, sysErr(err))
	}
	if op.Kind == opRemove {
		return nil
	}
	if err := tx.root.MkdirAll(path.Dir(op.Rel), 0o755); err != nil {
		return fmt.Errorf("create folder for %s: %w", op.Rel, sysErr(err))
	}
	return tx.move(ctx, op.Staged, op.Rel, op.Rel, onLocked)
}

// move renames from -> to, retrying while the file is in use and asking
// onLocked when that takes too long.
func (tx *Tx) move(ctx context.Context, from, to, rel string, onLocked func(LockedFile) Decision) error {
	for {
		err := tx.rename(from, to)
		for _, d := range tx.quickDelays {
			if err == nil || !isLocked(err) || sleep(ctx, d) != nil {
				break
			}
			err = tx.rename(from, to)
		}
		if err == nil {
			if tx.crash() {
				return errCrash
			}
			return nil
		}
		err = sysErr(err) // it names staging paths; rel is what the user knows
		if !isLocked(err) {
			return fmt.Errorf("move %s: %w", rel, err)
		}
		locked := fmt.Errorf("cannot replace %s: %w (%w). %s", rel, ErrLocked, err, LockedAdvice)
		if onLocked == nil || onLocked(LockedFile{Path: rel, Err: err}) != Retry {
			return locked
		}
		if err := sleep(ctx, tx.retryDelay); err != nil {
			return fmt.Errorf("%w: %w", locked, err)
		}
	}
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

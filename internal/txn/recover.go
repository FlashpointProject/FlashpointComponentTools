package txn

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"strings"
)

// journal.json in the txn dir. All paths are FP-relative.
type journal struct {
	State string      `json:"state"`
	Ops   []journalOp `json:"ops"`
}

type journalOp struct {
	Kind   string `json:"kind"`
	Rel    string `json:"rel"`
	Staged string `json:"staged,omitempty"`
	Backup string `json:"backup"`
	// Existed (place ops): the destination existed at prepare time, so a
	// file there without a backup is the untouched original, not ours.
	Existed bool `json:"existed,omitempty"`
}

const (
	journalName     = "journal.json"
	stateCommitting = "committing"
	stateCommitted  = "committed"
	opPlace         = "place"
	opRemove        = "remove"
)

// writeJournal replaces journal.json atomically (temp file, fsync, rename).
// An error means the previous journal is still in place.
func writeJournal(r *os.Root, dir string, j *journal) error {
	data, err := json.Marshal(j)
	if err == nil {
		err = writeSynced(r, dir+"/"+journalName+".tmp", data)
	}
	if err == nil {
		err = r.Rename(dir+"/"+journalName+".tmp", dir+"/"+journalName)
	}
	if err != nil {
		return fmt.Errorf("write transaction journal: %w", sysErr(err))
	}
	return nil
}

func readJournal(r *os.Root, dir string) (*journal, error) {
	data, err := r.ReadFile(dir + "/" + journalName)
	if err != nil {
		return nil, err
	}
	var j journal
	if err := json.Unmarshal(data, &j); err != nil {
		return nil, fmt.Errorf("corrupt journal in %s: %w", dir, err)
	}
	return &j, nil
}

// Recover finishes or rolls back any interrupted transaction of the install
// at root. Call it at startup, under the single-instance lock, before Begin.
// did reports whether an interrupted update was rolled back; deleting
// leftovers (staging, or backups of a committed update such as a then-running
// fpm.exe) is housekeeping and does not count. Every step is idempotent, so a
// crash during Recover is fixed by running it again. An error means a
// rollback could not complete (typically a file still in use); Begin keeps
// failing with ErrNeedsRecovery until a later Recover succeeds.
func Recover(root string, log *slog.Logger) (did bool, err error) {
	if log == nil {
		log = slog.New(slog.DiscardHandler)
	}
	r, err := os.OpenRoot(root)
	if err != nil {
		return false, fmt.Errorf("open Flashpoint folder: %w", err)
	}
	defer r.Close()
	entries, err := fs.ReadDir(r.FS(), txnBase)
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("read %s: %w", txnBase, sysErr(err))
	}
	var errs []error
	for _, e := range entries {
		dir := path.Join(txnBase, e.Name())
		if !e.IsDir() {
			r.Remove(dir) // stray file, not ours to keep
			continue
		}
		rolledBack, err := recoverDir(r, dir, log)
		did = did || rolledBack
		if err != nil {
			errs = append(errs, err)
		}
	}
	return did, errors.Join(errs...)
}

func recoverDir(r *os.Root, dir string, log *slog.Logger) (rolledBack bool, err error) {
	j, err := readJournal(r, dir)
	switch {
	case errors.Is(err, fs.ErrNotExist):
		// Staging only: never touched the install.
		log.Info("deleting unfinished staging of an interrupted update", "dir", dir)
		return false, r.RemoveAll(dir)
	case err != nil:
		// Without the op list backups cannot be put back. If there are
		// none, the dir is just garbage.
		if b, _ := fs.ReadDir(r.FS(), dir+"/backup"); len(b) == 0 {
			log.Warn("deleting transaction dir with unreadable journal (no backups in it)", "dir", dir, "err", err)
			return false, discardTxn(r, dir)
		}
		return false, fmt.Errorf("%w: %w; the original files are kept in %s", ErrRollbackFailed, err, dir+"/backup")
	case j.State == stateCommitted:
		log.Info("finishing cleanup of a completed update", "dir", dir)
		finishCommitted(r, dir, j, log)
		return false, nil
	}
	log.Warn("rolling back an interrupted update", "dir", dir, "ops", len(j.Ops))
	if err := rollback(r, j, r.Rename, log); err != nil {
		return true, fmt.Errorf("%w: %w", ErrRollbackFailed, err)
	}
	return true, discardTxn(r, dir)
}

// rollback undoes all ops in reverse. It continues past a failing op so as
// much as possible is restored, and reports all failures.
func rollback(r *os.Root, j *journal, rename func(string, string) error, log *slog.Logger) error {
	var errs []error
	for i := len(j.Ops) - 1; i >= 0; i-- {
		op := j.Ops[i]
		if err := undoOp(r, op, rename); err != nil {
			log.Error("rollback", "path", op.Rel, "err", err)
			errs = append(errs, err)
		} else if op.Kind == opPlace && !op.Existed {
			prune(r, op.Rel) // folders we may have created for it
		}
	}
	return errors.Join(errs...)
}

// undoOp reverts one op from what exists now, not from a progress log, so it
// is right after a crash at any point and can be interrupted and rerun.
// D = destination, S = staged file, B = backup. Forward, a place op goes
// S,(D) -> S,(B) -> D,(B); a remove op goes D -> B. So:
//   - place, S missing, D present: D is our file if D did not exist before
//     or its original is in B; move it back to S. Otherwise D is the
//     untouched original (S vanished, e.g. quarantined): leave it.
//   - then, if B is present, move it back to D.
func undoOp(r *os.Root, op journalOp, rename func(string, string) error) error {
	// An Lstat error other than "not found" counts as present: the renames
	// below then fail and report it, and nothing is lost.
	exists := func(p string) bool { _, err := r.Lstat(p); return !errors.Is(err, fs.ErrNotExist) }
	d, b := op.Rel, op.Backup
	dE, bE := exists(d), exists(b)
	if op.Kind == opPlace && dE && (!op.Existed || bE) && !exists(op.Staged) {
		if err := rename(d, op.Staged); err != nil {
			return fmt.Errorf("undo install of %s: %w", d, sysErr(err))
		}
		dE = false
	}
	if !bE {
		return nil
	}
	if dE {
		return fmt.Errorf("cannot restore %s: something else was created there; the original is in %s", d, b)
	}
	if err := rename(b, d); err != nil {
		return fmt.Errorf("restore %s: %w", d, sysErr(err))
	}
	return nil
}

// finishCommitted deletes the backups and staging of a committed
// transaction and prunes folders its removals left empty. The journal goes
// last, so a failure (e.g. the backup of a running exe on Windows) is
// retried by the next Begin or Recover. Failures are only logged.
func finishCommitted(r *os.Root, dir string, j *journal, log *slog.Logger) {
	for _, op := range j.Ops {
		if _, err := r.Lstat(op.Backup); err == nil && op.Kind == opRemove {
			prune(r, op.Rel) // only if we did remove a file there
		}
	}
	err := errors.Join(r.RemoveAll(dir+"/new"), r.RemoveAll(dir+"/backup"))
	if err == nil {
		err = discardTxn(r, dir)
	}
	if err != nil {
		log.Warn("update applied, but some leftovers could not be deleted; will retry next time", "dir", dir, "err", err)
	}
}

// prune removes the empty parent folders of rel, stopping at the first
// folder that is not empty, at Components (it marks a valid install) and at
// the root.
func prune(r *os.Root, rel string) {
	for d := path.Dir(rel); d != "." && !strings.EqualFold(d, "Components"); d = path.Dir(d) {
		if fi, err := r.Lstat(d); err != nil || !fi.IsDir() || r.Remove(d) != nil {
			return
		}
	}
}

// discardTxn deletes a txn dir, journal first: a txn dir without a journal
// is safe to delete, so a crash during the RemoveAll is harmless.
func discardTxn(r *os.Root, dir string) error {
	if err := r.Remove(dir + "/" + journalName); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return r.RemoveAll(dir)
}

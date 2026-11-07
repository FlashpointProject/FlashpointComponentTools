package gui

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"sync"
	"time"

	"gioui.org/layout"
	"gioui.org/widget"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/engine"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/manifest"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/plan"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/txn"
)

type rowState int

// The row states the engine reports; the rest (committing, done, not
// changed) follows from the run and is decided when drawing.
const (
	rsWaiting     rowState = iota
	rsDownloading          // also once complete (Done == Total)
	rsRetrying
	rsExtracting // extraction started; run.extracting is the current one
	rsFailed
	rsSkipped
)

// runRow is one component of a running change.
type runRow struct {
	ID, Title, Verb string // Verb: Install, Update, Repair, Remove
	Size            int64  // download size, 0 when nothing is downloaded
	St              rowState
	Done, Total     int64
	Attempt         int
	RetryAt         time.Time

	speed float64 // bytes/s, updated about once a second
	markT time.Time
	markB int64
}

// run is the state of one Execute call. Progress events arrive on engine
// goroutines, so the fields above the line are guarded by mu.
type run struct {
	mu         sync.Mutex
	rows       []*runRow
	byID       map[string]*runRow
	extracting *runRow
	committing bool
	cancelling bool
	lockedPath string // the file in use when the user chose Abort
	abortWhat  string // the component whose download error the user aborted on

	repair bool

	// Set on the UI goroutine.
	cancel   context.CancelFunc
	finished bool
	res      *engine.Result
	err      error
	list     widget.List
}

func newRun(m *model, p *plan.Plan, repair bool) *run {
	r := &run{byID: map[string]*runRow{}, repair: repair}
	add := func(id, verb string, c *manifest.Component) {
		row := &runRow{ID: id, Title: m.title(id), Verb: verb}
		if c != nil && c.HasArchive() {
			row.Size = c.DownloadSize
		}
		r.rows = append(r.rows, row)
		r.byID[id] = row
	}
	for _, c := range p.Install {
		add(c.ID, "Install", c)
	}
	for _, c := range p.Update {
		add(c.ID, map[bool]string{false: "Update", true: "Repair"}[repair], c)
	}
	for _, rec := range p.Remove {
		add(rec.ID, "Remove", nil)
	}
	r.list.Axis = layout.Vertical
	return r
}

func (r *run) event(ev engine.Event) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	if ev.Kind == engine.Committing {
		r.committing = true
		return
	}
	row := r.byID[ev.ID]
	if row == nil {
		return
	}
	switch ev.Kind {
	case engine.DownloadStarted:
		row.St, row.markT, row.markB = rsDownloading, now, row.Done
	case engine.DownloadProgress:
		row.St, row.Done, row.Total = rsDownloading, ev.Done, ev.Total
		if dt := now.Sub(row.markT); dt >= time.Second {
			row.speed = float64(row.Done-row.markB) / dt.Seconds()
			row.markT, row.markB = now, row.Done
		}
		if ev.Total > 0 && ev.Done >= ev.Total {
			row.speed = 0
		}
	case engine.DownloadRetrying:
		row.St, row.Attempt, row.RetryAt, row.speed = rsRetrying, ev.Attempt, now.Add(ev.Wait), 0
	case engine.Extracting:
		row.St, r.extracting = rsExtracting, row
	}
}

func (r *run) setState(id string, st rowState) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if row := r.byID[id]; row != nil {
		row.St = st
	}
}

// requestCancel cancels unless the commit has started (it then finishes).
func (r *run) requestCancel() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if !r.committing && !r.cancelling {
		r.cancelling = true
		r.cancel()
	}
}

func (r *run) finish(res *engine.Result, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.finished, r.res, r.err = true, res, err
	for _, row := range r.rows {
		if slices.Contains(res.Skipped, row.ID) {
			row.St = rsSkipped // also the dependents of a skipped one
		}
	}
}

// managerUpdated reports whether the manager's own files were replaced.
func (r *run) managerUpdated() bool {
	return r.err == nil && (slices.Contains(r.res.Updated, "core-manager") || slices.Contains(r.res.Installed, "core-manager"))
}

// outcome is the result screen's heading and explanation in plain words;
// failed marks an error the user did not choose. In the installer, Quit
// removes the destination with the downloads. r.mu is held.
func (r *run) outcome(installer bool) (title, detail string, failed bool) {
	switch err := r.err; {
	case err == nil && len(r.res.Skipped) > 0:
		return "Finished, but some components were skipped", "", false
	case err == nil:
		return "All done", "", false
	case errors.Is(err, context.Canceled) && installer:
		return "Cancelled", "Nothing was installed. Press Back to try again; the downloads continue where they stopped. Quit removes them.", false
	case errors.Is(err, context.Canceled):
		return "Cancelled", "Nothing was changed. Partial downloads are kept and continue next time.", false
	case errors.Is(err, engine.ErrAborted):
		return "Stopped", "Nothing was changed. " + sentence(strings.TrimPrefix(plain(err, r.abortWhat), "Aborted: ")), false
	case errors.Is(err, txn.ErrRollbackFailed) && installer:
		return "The change could not be completed or undone",
			plain(err, "") + " The installation could not be undone completely; quitting removes the folder.", true
	case errors.Is(err, txn.ErrRollbackFailed):
		return "The change could not be completed or undone",
			plain(err, "") + " Close all Flashpoint programs and start the manager again; it finishes the repair when it starts.", true
	case errors.Is(err, txn.ErrLocked) && r.lockedPath != "":
		return "Cancelled", "Cancelled because " + r.lockedPath + " was in use. Nothing was changed.", false
	default:
		return "The change failed", plain(err, "") + " Nothing was changed.", true
	}
}

func (r *run) counts() string {
	var parts []string
	for _, x := range []struct {
		ids  []string
		verb string
	}{{r.res.Installed, "installed"}, {r.res.Updated, map[bool]string{false: "updated", true: "repaired"}[r.repair]},
		{r.res.Removed, "removed"}, {r.res.Skipped, "skipped"}} {
		if len(x.ids) > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", len(x.ids), x.verb))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return "Components: " + strings.Join(parts, ", ")
}

// hooks connects Execute to the window. OnError and OnLocked block the
// engine goroutine on a modal until the user answers or the run is cancelled.
func (u *ui) hooks(r *run) engine.Hooks {
	m := u.m // not replaced while a run is active
	ask := func(ctx context.Context, title string, body []string, labels []string) int {
		ans := make(chan int, 1)
		u.post(func() { u.modal = newModal(title, body, labels, func(i int) { ans <- i }) })
		select {
		case i := <-ans:
			return i
		case <-ctx.Done():
			return -1
		}
	}
	return engine.Hooks{
		Progress: func(ev engine.Event) { r.event(ev); u.invalidate() },
		OnError: func(ctx context.Context, id string, err error) engine.Decision {
			r.setState(id, rsFailed)
			u.logErr("could not download "+id, err)
			t := m.title(id)
			switch ask(ctx, "Download failed", []string{"Could not download " + t + ".", plain(err, t)},
				[]string{"Retry", "Skip " + t, "Abort"}) {
			case 0:
				r.setState(id, rsWaiting)
				return engine.Retry
			case 1:
				r.setState(id, rsSkipped)
				return engine.Skip
			}
			r.mu.Lock()
			r.abortWhat = t
			r.mu.Unlock()
			return engine.Abort
		},
		OnLocked: func(ctx context.Context, path string) bool {
			if ask(ctx, "A file is in use", []string{
				winPath(path) + " is in use or not writable.",
				txn.LockedAdvice + " Then press Retry. Abort undoes the whole change.",
			}, []string{"Retry", "Abort"}) == 0 {
				return true
			}
			r.mu.Lock()
			r.lockedPath = winPath(path)
			r.mu.Unlock()
			return false
		},
	}
}

// planLines lists a plan for the confirm dialog (repair: its updates are
// repairs).
func planLines(m *model, p *plan.Plan, repair bool) []string {
	var out []string
	// Say why the plan added a component: which selected one needs it, or
	// which removed one it needs. More than 5 removed for one reason (the
	// themes of a removed base theme) share a line.
	notes, reason, why := map[string]string{}, map[string]string{}, map[string][]string{}
	for _, id := range p.AddedDeps {
		if t := m.whoNeeds(p, id); t != "" {
			notes[id] = "  (needed by " + t + ")"
		}
	}
	for _, id := range p.AddedDependents {
		if t := m.removedDep(p, id); t != "" {
			notes[id] = "  (removed because it needs " + t + ")"
			reason[id], why[t] = t, append(why[t], id)
		}
	}
	section := func(head string, ids []string) {
		if len(ids) == 0 {
			return
		}
		out = append(out, fmt.Sprintf("%s (%d):", head, len(ids)))
		for _, id := range ids {
			switch g := why[reason[id]]; {
			case len(g) <= 5:
				out = append(out, "    "+m.title(id)+notes[id])
			case g[0] == id:
				out = append(out, fmt.Sprintf("    %d that need %s: %s", len(g), reason[id], strings.Join(m.titles(g), ", ")))
			}
		}
	}
	var inst, upd, rm []string
	for _, c := range p.Install {
		inst = append(inst, c.ID)
	}
	for _, c := range p.Update {
		upd = append(upd, c.ID)
	}
	for _, r := range p.Remove {
		rm = append(rm, r.ID)
	}
	section("Install", inst)
	section(map[bool]string{false: "Update", true: "Repair"}[repair], upd)
	section("Remove", rm)
	return append(out, "", sizes(p))
}

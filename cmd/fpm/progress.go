package main

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/engine"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/plan"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/txn"
)

// lockedWait is how long a run without a terminal keeps retrying a file in
// use (the launcher exits about 2.5 s after starting the manager).
const lockedWait = 60 * time.Second

// hooks reports progress on stderr (one line per step, and the overall
// download percentage at most every 3 seconds) and asks the user about
// failures when stdin is a terminal. The engine serializes the calls.
func (a *app) hooks(p *plan.Plan) engine.Hooks {
	var total int64
	for _, c := range append(p.Install, p.Update...) {
		if c.HasArchive() {
			total += c.DownloadSize
		}
	}
	done := map[string]int64{}
	last := time.Now()
	lockedSince := map[string]time.Time{}

	return engine.Hooks{
		Progress: func(ev engine.Event) {
			switch ev.Kind {
			case engine.DownloadStarted:
				a.line("Downloading %s...", ev.ID)
			case engine.DownloadRetrying:
				a.line("Retrying %s in %s (attempt %d failed: %v)", ev.ID, ev.Wait.Round(time.Second), ev.Attempt, ev.Err)
			case engine.Extracting:
				a.line("Extracting %s...", ev.ID)
			case engine.Committing:
				a.line("Applying changes...")
			case engine.DownloadProgress:
				done[ev.ID] = ev.Done
				if time.Since(last) < 3*time.Second {
					return
				}
				last = time.Now()
				var sum int64
				for _, d := range done {
					sum += d
				}
				a.line("Downloaded %s of %s", engine.FormatSize(sum), engine.FormatSize(total))
			}
		},
		OnError: func(ctx context.Context, id string, err error) engine.Decision {
			a.msg("%v", err)
			if !a.stdinTTY {
				return engine.Abort // nobody to ask
			}
			for ctx.Err() == nil {
				a.prompt("[r]etry, [s]kip " + id + ", or [a]bort? ")
				ans, ok := a.readLine()
				switch ans = strings.ToLower(ans); ans {
				case "r", "retry":
					return engine.Retry
				case "s", "skip":
					return engine.Skip
				}
				if !ok || ans == "a" || ans == "abort" {
					break
				}
			}
			return engine.Abort
		},
		OnLocked: func(ctx context.Context, path string) bool {
			if a.stdinTTY {
				a.prompt(path + " is in use or not writable. " + txn.LockedAdvice + " Then press Enter to try again (or type a and Enter to abort): ")
				ans, ok := a.readLine()
				return ok && !strings.EqualFold(ans, "a") && !strings.EqualFold(ans, "abort")
			}
			first, seen := lockedSince[path]
			if !seen {
				lockedSince[path] = time.Now()
				a.msg("%s is in use; waiting up to %s for Flashpoint and any program using it to close", path, lockedWait)
				return true
			}
			if time.Since(first) < lockedWait {
				return true
			}
			a.msg("%s is still in use; giving up", path)
			return false
		},
	}
}

func (a *app) prompt(s string) { io.WriteString(a.stderr, s) }

func (a *app) line(format string, args ...any) { fmt.Fprintf(a.stderr, format+"\r\n", args...) }

// readLine reads one answer from stdin; ok is false at EOF.
func (a *app) readLine() (string, bool) {
	line, err := a.in.ReadString('\n')
	if err != nil && line == "" {
		return "", false
	}
	return strings.TrimSpace(line), true
}

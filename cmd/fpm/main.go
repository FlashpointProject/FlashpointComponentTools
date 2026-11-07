// Command fpm is the Flashpoint component manager CLI, a compatible superset
// of the legacy C# fpm.exe: same commands, arguments and output, in
// particular `fpm list verbose`, which the launcher runs at every start and
// parses (DESIGN.md §1.1, §5).
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/engine"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/logfile"
)

// version is set at build time: -ldflags "-X main.version=2.0.0".
var version = "dev"

// Exit codes (DESIGN.md §5).
const (
	exitOK        = 0
	exitError     = 1
	exitUsage     = 2
	exitPartial   = 3 // some components were skipped
	exitCancelled = 4
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	stderr := &lockedWriter{w: os.Stderr}

	// The first Ctrl-C cancels (downloads stop and are kept; a commit in
	// progress still finishes or rolls back), the second exits at once.
	sig := make(chan os.Signal, 2)
	signal.Notify(sig, os.Interrupt)
	go func() {
		<-sig
		fmt.Fprint(stderr, "\r\nCancelling... (press Ctrl-C again to quit immediately)\r\n")
		cancel()
		<-sig
		os.Exit(exitCancelled)
	}()

	mgr := ""
	if exe, err := os.Executable(); err == nil {
		if real, err := filepath.EvalSymlinks(exe); err == nil {
			exe = real
		}
		mgr = filepath.Dir(exe)
	}
	os.Exit(run(ctx, os.Args[1:], env{stdin: os.Stdin, stdout: os.Stdout, stderr: stderr,
		stdinTTY: isTerminal(os.Stdin), managerDir: mgr}))
}

// env is everything run needs from the process, so tests can fake it.
type env struct {
	stdin          io.Reader
	stdout, stderr io.Writer
	stdinTTY       bool
	managerDir     string // folder of the executable: fpm.cfg lives here
}

type options struct {
	yes, verbose, help, version bool
	path, source                string
}

type app struct {
	env
	opts options
	cmd  string
	in   *bufio.Reader
	log  *slog.Logger
	logf *logfile.Sink
}

// usageError is reported with exit code 2.
type usageError struct{ msg string }

func (e *usageError) Error() string { return e.msg }

// parseArgs separates the dash options (anywhere on the line, "--" ends
// them) from the legacy positional arguments.
func parseArgs(args []string) (o options, pos []string, err error) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		if a == "--" {
			return o, append(pos, args[i+1:]...), nil
		}
		if !strings.HasPrefix(a, "-") || a == "-" {
			pos = append(pos, a)
			continue
		}
		name, val, hasVal := strings.Cut(a, "=")
		value := func(dst *string) {
			if !hasVal {
				if i+1 >= len(args) {
					err = &usageError{fmt.Sprintf("option %s needs a value", name)}
					return
				}
				i++
				val = args[i]
			}
			*dst = val
		}
		switch name {
		case "-y", "--yes":
			o.yes = true
		case "-v", "--verbose":
			o.verbose = true
		case "-h", "--help":
			o.help = true
		case "--version":
			o.version = true
		case "--path":
			value(&o.path)
		case "--source":
			value(&o.source)
		default:
			err = &usageError{fmt.Sprintf("unknown option %s", a)}
		}
		if err != nil {
			return o, nil, err
		}
	}
	return o, pos, nil
}

func run(ctx context.Context, args []string, e env) int {
	e.stderr = &lockedWriter{w: e.stderr} // progress, logs and the signal handler share it
	a := &app{env: e, in: bufio.NewReader(e.stdin), logf: &logfile.Sink{}}
	defer a.logf.Close()
	opts, pos, err := parseArgs(args)
	if err != nil {
		fmt.Fprintf(a.stderr, "fpm: %v\r\n(run \"fpm help\" for usage)\r\n", err)
		return exitUsage
	}
	a.opts = opts
	switch {
	case opts.version:
		pos = []string{"version"}
	case opts.help || len(pos) == 0:
		pos = []string{"help"}
	}
	a.cmd, pos = pos[0], pos[1:]

	// Log to <root>/Logs/fpm.log once the root is known, and with --verbose
	// also to stderr.
	var w io.Writer = a.logf
	level := slog.LevelInfo
	if opts.verbose {
		w, level = io.MultiWriter(a.logf, a.stderr), slog.LevelDebug
	}
	a.log = slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))

	switch a.cmd {
	case "help":
		io.WriteString(a.stdout, strings.ReplaceAll(helpText, "\n", "\r\n")+"\r\n")
		return exitOK
	case "version":
		a.out("fpm %s (%s/%s, %s)", version, runtime.GOOS, runtime.GOARCH, runtime.Version())
		return exitOK
	case "path", "source":
		return a.configCmd(pos, a.cmd == "path")
	case "list":
		return a.listCmd(ctx, pos)
	case "info":
		return a.infoCmd(ctx, pos)
	case "download", "remove", "update":
		return a.changeCmd(ctx, pos)
	case "verify":
		return a.verifyCmd(ctx, pos)
	case "recover":
		return a.recoverCmd(ctx)
	}
	fmt.Fprintf(a.stderr, "fpm: unknown command %q (run \"fpm help\" for usage)\r\n", a.cmd)
	return exitUsage
}

// out prints a result line to stdout. Every line ends in CRLF on every OS:
// the launcher strips two characters from each `list` line (launcher
// src/back/index.ts:780), and legacy fpm (.NET on Windows) wrote CRLF.
func (a *app) out(format string, args ...any) {
	fmt.Fprintf(a.stdout, format+"\r\n", args...)
}

// msg prints a diagnostic to stderr in legacy fpm's "<command>: <message>"
// form (fpm src/Helpers.cs:236-240, which used stdout).
func (a *app) msg(format string, args ...any) {
	fmt.Fprintf(a.stderr, a.cmd+": "+format+"\r\n", args...)
}

// fail reports err and returns its exit code.
func (a *app) fail(err error) int {
	var ue *usageError
	switch {
	case errors.As(err, &ue):
		a.msg("%v", err)
		return exitUsage
	case errors.Is(err, context.Canceled):
		a.msg("cancelled; nothing was changed")
		return exitCancelled
	}
	a.msg("%v", err)
	a.log.Error("command failed", "cmd", a.cmd, "err", err)
	return exitError
}

func (a *app) open(ctx context.Context) (*engine.Engine, error) {
	e, err := engine.Open(ctx, engine.Options{
		ManagerDir:     a.managerDir,
		RootOverride:   a.opts.path,
		SourceOverride: a.opts.source,
		Logger:         a.log,
		UserAgent:      fmt.Sprintf("FlashpointManager/%s (fpm; %s/%s)", version, runtime.GOOS, runtime.GOARCH),
	})
	if err != nil {
		return nil, err
	}
	if err := a.logf.Open(filepath.Join(e.Root(), "Logs", "fpm.log")); err != nil {
		a.log.Warn("cannot write the log file", "err", err) // seen only with --verbose
	}
	a.log.Info("fpm started", "version", version, "cmd", a.cmd, "root", e.Root(), "source", e.Source())
	return e, nil
}

// lockedWriter serializes writes from several goroutines.
type lockedWriter struct {
	mu sync.Mutex
	w  io.Writer
}

func (l *lockedWriter) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.w.Write(p)
}

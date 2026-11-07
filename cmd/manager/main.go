// Command manager is FlashpointManager.exe, the Flashpoint component
// manager GUI. It accepts the legacy switches the launcher uses
// (DESIGN.md §1.1, §5):
//
//	FlashpointManager.exe /update /launcher         "Open Manager"
//	FlashpointManager.exe /notemp /download "a b"   on-demand install, exit 0 = installed
//
// plus --path <dir>. Build with -ldflags "-H=windowsgui" on
// Windows; the icon and manifest come from rsrc_windows_*.syso (see
// scripts/build.sh).
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/engine"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/gui"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/install"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/launch"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/logfile"
)

// version is set at build time: -ldflags "-X main.version=2.0.0".
var version = "dev"

func main() {
	args := parseArgs(os.Args[1:])
	// Set in the launcher's back end (Electron running as Node), which
	// spawns us; inherited by the launcher we start, it makes the launcher
	// run as bare Node and exit at once.
	os.Unsetenv("ELECTRON_RUN_AS_NODE")

	// Resolved now: after a self-update os.Executable may name the old,
	// renamed file.
	exe, err := os.Executable()
	if err == nil {
		exe, err = filepath.EvalSymlinks(exe)
	}
	sink := &logfile.Sink{}
	m := &manager{args: args, exe: exe, dir: filepath.Dir(exe), sink: sink,
		log: slog.New(slog.NewTextHandler(sink, nil))}
	if err != nil {
		m.log.Error("cannot locate the executable", "err", err)
		m.dir, _ = os.Getwd()
	}
	m.log.Info("manager started", "version", version, "args", os.Args[1:], "dir", m.dir)
	for _, a := range args.ignored {
		m.log.Warn("ignoring unknown argument", "arg", a)
	}

	// The launcher's job object kills what it spawned when it exits, ~2.5 s
	// after starting us with /update /launcher. Leave it by starting a copy
	// of ourselves, before taking any lock. Not with /notemp: there the
	// launcher waits for our exit code.
	if !args.notemp && !args.relaunch && err == nil && inJob() {
		if err := startOutsideJob(exe, slices.Concat(os.Args[1:], []string{relaunchMarker})); err != nil {
			m.log.Warn("could not relaunch outside the job object", "err", err)
		} else {
			m.log.Info("relaunched outside the job object")
			os.Exit(0)
		}
	}
	if args.relaunch {
		waitParent(3 * time.Second) // it may hold the lock or be the old exe
	}

	start := gui.ViewComponents
	if args.update {
		start = gui.ViewUpdates
	}
	gui.Run(gui.Options{
		Version:      version,
		StartView:    start,
		DownloadIDs:  args.ids,
		Open:         m.open,
		SaveSettings: func(root, source string) error { return engine.SaveSettings(m.dir, root, source) },
		OpenLauncher: m.openLauncher,
		Restart:      m.restart,
		OpenLogs:     m.openLogs,
		CanUninstall: m.canUninstall,
		Uninstall:    m.uninstall,
		Log:          m.log,
		Exit:         m.exit,
	})
}

// exit runs when the window has closed. With /launcher it starts the
// launcher, unless that happened already, a restarted manager will, or
// Flashpoint was uninstalled; then the running exe, kept by Uninstall, is
// deleted once this process has ended.
func (m *manager) exit(code int) {
	m.mu.Lock()
	uninstalled := m.uninstalled
	startLauncher := m.args.launcher && !m.launched && !m.restarted && uninstalled == ""
	m.mu.Unlock()
	if startLauncher {
		if err := m.openLauncher(); err != nil {
			m.log.Error("could not start the launcher", "err", err)
		}
	}
	if uninstalled != "" {
		if err := deleteAfterExit(m.exe, uninstalled); err != nil {
			m.log.Error("could not delete the manager", "err", err)
		}
	}
	m.log.Info("manager exited", "code", code)
	os.Exit(code)
}

// canUninstall says why root can't be uninstalled, nil if it can. Only the
// Flashpoint this manager is in: an old fpm.cfg holds an absolute path, so a
// copied install's manager still manages the original.
func (m *manager) canUninstall(root string) error {
	if !strings.EqualFold(filepath.Clean(root), filepath.Dir(m.dir)) {
		return fmt.Errorf("This manager is in %s but manages %s. Uninstall Flashpoint with the manager in its own folder.", filepath.Dir(m.dir), root)
	}
	return install.CheckUninstall(root)
}

// uninstall deletes Flashpoint from root, after canUninstall. The log file
// is in root\Logs, so it is closed first; the running exe is kept, and exit
// deletes it.
func (m *manager) uninstall(root string) ([]string, error) {
	m.log.Info("uninstalling", "root", root)
	m.sink.Close()
	m.mu.Lock()
	m.uninstalled = root
	m.mu.Unlock()
	return install.Uninstall(root, m.exe)
}

type manager struct {
	args cliArgs
	exe  string // own executable, symlinks resolved
	dir  string // its folder (normally <FP>/Manager): fpm.cfg lives here
	log  *slog.Logger
	sink *logfile.Sink

	mu          sync.Mutex
	root        string // of the last opened engine
	launched    bool   // the launcher was started
	restarted   bool   // a new manager was started
	uninstalled string // the Flashpoint folder Uninstall deleted
}

// open opens the install; rootOverride "" means --path, else fpm.cfg.
func (m *manager) open(rootOverride string) (*engine.Engine, error) {
	if rootOverride == "" {
		rootOverride = m.args.path
	}
	e, err := engine.Open(context.Background(), engine.Options{
		ManagerDir:   m.dir,
		RootOverride: rootOverride,
		Logger:       m.log,
		UserAgent:    fmt.Sprintf("FlashpointManager/%s (gui; %s/%s)", version, runtime.GOOS, runtime.GOARCH),
	})
	if err != nil {
		m.log.Error("cannot open the Flashpoint folder", "err", err)
		return nil, err
	}
	if err := m.sink.Open(filepath.Join(e.Root(), "Logs", "manager.log")); err != nil {
		m.log.Warn("cannot write the log file", "err", err)
	}
	m.log.Info("opened", "root", e.Root(), "source", e.Source(), "channel", e.Channel())
	m.mu.Lock()
	m.root = e.Root()
	m.mu.Unlock()
	return e, nil
}

// openLauncher starts the launcher of the last opened install (else the
// Manager folder's parent) with cwd = its folder, like the C# manager
// (Main.cs:151-161), and does not wait for it.
func (m *manager) openLauncher() error {
	m.mu.Lock()
	root := m.root
	m.mu.Unlock()
	if root == "" {
		root = filepath.Dir(m.dir)
	}
	exe, err := launch.StartLauncher(root)
	if err != nil {
		return err
	}
	m.log.Info("started the launcher", "exe", exe)
	m.mu.Lock()
	m.launched = true
	m.mu.Unlock()
	return nil
}

// openLogs shows <root>/Logs in the file manager.
func (m *manager) openLogs() error {
	m.mu.Lock()
	dir := filepath.Join(m.root, "Logs")
	m.mu.Unlock()
	opener := map[string]string{"windows": "explorer.exe", "darwin": "open"}[runtime.GOOS]
	if opener == "" {
		opener = "xdg-open"
	}
	return launch.Start(opener, []string{dir}, "")
}

// restart starts a new manager with the same arguments minus /download (it
// keeps /launcher, so this one must not start the launcher on exit). The
// caller then quits; release the install lock first. Our child is outside
// any launcher job already, hence relaunchMarker.
func (m *manager) restart() error {
	if err := launch.Start(m.exe, slices.Concat(m.args.restart, []string{relaunchMarker}), ""); err != nil {
		return fmt.Errorf("restart the manager: %w", err)
	}
	m.log.Info("restarted", "args", m.args.restart)
	m.mu.Lock()
	m.restarted = true
	m.mu.Unlock()
	return nil
}

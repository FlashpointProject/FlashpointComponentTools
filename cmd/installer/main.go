// Command installer is FlashpointInstaller.exe, the wizard that installs
// Flashpoint into a new folder (docs/go-rewrite/M3.md). It takes one
// optional switch, for testing and the development channel:
//
//	FlashpointInstaller.exe --source <components.xml URL>
//
// Build with -ldflags "-H=windowsgui" on Windows; the icon and manifest
// come from rsrc_windows_*.syso (see scripts/build.sh).
package main

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/engine"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/gui"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/install"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/launch"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/logfile"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
)

// version is set at build time: -ldflags "-X main.version=2.0.0".
var version = "dev"

func main() {
	// In %TEMP%, not in the destination: an abandoned install's folder is
	// deleted by install.Cleanup.
	sink := &logfile.Sink{}
	in := &installer{log: slog.New(slog.NewTextHandler(sink, nil))}
	if err := sink.Open(filepath.Join(os.TempDir(), "FlashpointInstaller.log")); err != nil {
		in.log.Warn("cannot write the log file", "err", err) // kept in memory only
	}
	var ignored []string
	in.source, ignored = parseArgs(os.Args[1:])
	in.log.Info("installer started", "version", version, "args", os.Args[1:], "source", in.source)
	for _, a := range ignored {
		in.log.Warn("ignoring unknown argument", "arg", a)
	}

	exe, _ := os.Executable()
	gui.Run(gui.Options{
		Version: version,
		Install: &gui.InstallOptions{
			DefaultDest:   install.DefaultDest(exe),
			CheckDest:     install.CheckDest,
			Cleanup:       install.Cleanup,
			NeedsVCRedist: install.NeedsVCRedist,
			InstallVCRedist: func(ctx context.Context) (bool, error) {
				return install.InstallVCRedist(ctx, in.log)
			},
			CreateShortcuts: install.CreateShortcuts,
		},
		Open: in.open,
		// Not engine.SaveSettings: it accepts only https sources, and the
		// source was checked by engine.Open already (http is for testing).
		SaveSettings: func(root, source string) error {
			return state.SaveConfig(filepath.Join(root, "Manager"), &state.Config{RootRaw: `..\`, Source: source})
		},
		OpenLauncher: in.openLauncher,
		Log:          in.log,
		Exit: func(code int) {
			in.log.Info("installer exited", "code", code)
			os.Exit(code)
		},
	})
}

// parseArgs returns the --source value (default state.DefaultSource) and
// the other arguments, which are ignored: the GUI has no console to report
// a usage error on.
func parseArgs(args []string) (source string, ignored []string) {
	source = state.DefaultSource
	for i := 0; i < len(args); i++ {
		if args[i] == "--source" && i+1 < len(args) {
			source = args[i+1]
			i++
		} else {
			ignored = append(ignored, args[i])
		}
	}
	return source, ignored
}

type installer struct {
	source string
	log    *slog.Logger
	// root is the last opened destination; open sets it on the GUI's worker
	// and openLauncher reads it after the window has closed and the
	// workers were waited for, so it needs no lock.
	root string
}

// open creates the destination folder and opens an engine on it: no
// fpm.cfg there yet, so the root and source are given explicitly.
func (in *installer) open(dest string) (*engine.Engine, error) {
	if err := os.MkdirAll(dest, 0o755); err != nil {
		// "Could not create D:\X: Access is denied.", not "mkdir D:/X: ...".
		var pe *fs.PathError
		if errors.As(err, &pe) {
			err = pe.Err
		}
		return nil, fmt.Errorf("could not create %s: %w", dest, err)
	}
	e, err := engine.Open(context.Background(), engine.Options{
		ManagerDir:     filepath.Join(dest, "Manager"),
		RootOverride:   dest,
		SourceOverride: in.source,
		Logger:         in.log,
		UserAgent:      fmt.Sprintf("FlashpointInstaller/%s (gui; %s/%s)", version, runtime.GOOS, runtime.GOARCH),
	})
	if err != nil {
		return nil, err
	}
	in.log.Info("opened", "root", e.Root(), "source", e.Source(), "channel", e.Channel())
	in.root = dest
	return e, nil
}

// openLauncher starts the new install's launcher like the manager does.
func (in *installer) openLauncher() error {
	exe, err := launch.StartLauncher(in.root)
	if err != nil {
		return err
	}
	in.log.Info("started the launcher", "exe", exe)
	return nil
}

package main

import (
	"slices"
	"strings"
)

// relaunchMarker marks the copy started to leave a job object (main.go).
const relaunchMarker = "/relaunched"

// cliArgs are FlashpointManager.exe's command-line switches.
type cliArgs struct {
	update   bool     // /update: start on the Updates view
	launcher bool     // /launcher: start the launcher when the window closes
	ids      []string // /download <ids...>: install these, show only progress
	path     string   // --path <dir>: Flashpoint folder, fpm.cfg unchanged
	notemp   bool     // /notemp: never relaunch (the caller waits for our exit code)
	relaunch bool     // relaunchMarker: this is the relaunched copy

	ignored []string // unknown arguments (logged, never fatal)
	restart []string // the arguments without /download and its IDs
}

// parseArgs reads the legacy switches like the C# manager (Program.cs:160-185;
// /notemp at :24): they are case-insensitive and can be combined, and
// /download takes every following argument up to the next one starting
// with "/". Differences: each /download argument is also split on
// whitespace, because the launcher passes all IDs as one argument
// (launcher GameLauncher.ts: ['/notemp', '/download', ids.join(' ')]), which
// the C# manager failed to find; IDs are lowercased (the C# lookup was
// case-insensitive); /download also stops at a "-" option. /notemp keeps its
// meaning "don't relaunch", though the relaunch is now only to leave the
// launcher's job (main.go), not a %TEMP% copy (DESIGN.md §4.6). The dash
// option --path and relaunchMarker are new. Nothing is fatal: the GUI has
// no console to report a usage error on.
func parseArgs(args []string) cliArgs {
	var a cliArgs
	inDownload := false
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "/") || strings.HasPrefix(arg, "-") {
			inDownload = false
		}
		if inDownload {
			for _, id := range strings.Fields(strings.ToLower(arg)) {
				if !slices.Contains(a.ids, id) {
					a.ids = append(a.ids, id)
				}
			}
			continue
		}
		keep := true
		switch strings.ToLower(arg) {
		case "/update":
			a.update = true
		case "/launcher":
			a.launcher = true
		case "/download":
			inDownload, keep = true, false
		case "/notemp":
			a.notemp = true
		case relaunchMarker:
			a.relaunch, keep = true, false
		case "--path":
			if i+1 < len(args) {
				a.path = args[i+1]
				a.restart = append(a.restart, arg)
				i++
				arg = args[i]
			} else {
				a.ignored = append(a.ignored, arg)
			}
		default:
			if v, ok := strings.CutPrefix(arg, "--path="); ok {
				a.path = v
			} else {
				a.ignored = append(a.ignored, arg)
			}
		}
		if keep {
			a.restart = append(a.restart, arg)
		}
	}
	return a
}

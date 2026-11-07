package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/FlashpointProject/FlashpointComponentTools/internal/engine"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/manifest"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/plan"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/state"
	"github.com/FlashpointProject/FlashpointComponentTools/internal/txn"
)

// listBudget is how long list and info wait for the network before using a
// local copy: the launcher's startup blocks on `fpm list verbose` (legacy
// fpm had a 3 s timeout, fpm src/Main.cs:23).
const listBudget = 3 * time.Second

// configCmd implements path and source (fpm src/Handlers.cs:11-67). Unlike
// legacy fpm, reading a value does not create an empty fpm.cfg.
func (a *app) configCmd(args []string, isPath bool) int {
	_, statErr := os.Stat(filepath.Join(a.managerDir, state.ConfigName))
	cfg, err := state.LoadConfig(a.managerDir)
	if err != nil {
		return a.fail(err)
	}
	switch {
	case len(args) == 0 && isPath && statErr != nil:
		a.out("%s", filepath.Dir(a.managerDir)) // legacy default: GetFullPath(cwd\..)
		return exitOK
	case len(args) == 0 && isPath:
		a.out("%s", cfg.RootRaw)
		return exitOK
	case len(args) == 0:
		a.out("%s", cfg.Source)
		return exitOK
	case isPath:
		abs, err := filepath.Abs(args[0])
		if err != nil {
			a.msg("The specified path is invalid")
			return exitError
		}
		if !state.IsRoot(abs) {
			a.msg("warning: %s does not look like a Flashpoint folder (no Components records, or Launcher and Data folders)", abs)
		}
		cfg.RootRaw = abs
	default:
		if _, err := engine.ValidSource(args[0]); err != nil {
			a.msg("The specified URL is invalid")
			return exitError
		}
		if statErr != nil {
			cfg.RootRaw = `..\`
		}
		cfg.Source = args[0]
	}
	if err := state.SaveConfig(a.managerDir, cfg); err != nil {
		a.msg("Could not write to fpm.cfg (is it open in another program?): %v", err)
		return exitError
	}
	return exitOK
}

// loadForReading opens the install and the manifest within listBudget.
func (a *app) loadForReading(ctx context.Context) (*engine.Engine, *engine.Snapshot, error) {
	e, err := a.open(ctx)
	if err != nil {
		return nil, nil, err
	}
	lctx, cancel := context.WithTimeout(ctx, listBudget)
	defer cancel()
	snap, err := e.LoadManifest(lctx, engine.ForReading)
	if err != nil {
		return nil, nil, err
	}
	for _, w := range snap.Warnings {
		a.msg("%s", w)
	}
	return e, snap, nil
}

// legacyStatus reproduces legacy fpm's flags (fpm src/Helpers.cs:135-159):
// downloaded = Components/<id> exists; outdated = the first token of its
// first line differs from the manifest hash (a header with fewer than two
// tokens is never outdated). Unreadable records are judged the same way.
//
// One deliberate difference: core-database is never outdated (the C#
// Manager's rule, and what `fpm update` does). Legacy fpm printed "!", which
// made `fpm update` overwrite the user's database; the launcher maps that
// "!" to up to date anyway (launcher src/back/index.ts:794-797).
func legacyStatus(root string, c *manifest.Component) (downloaded, outdated bool) {
	data, err := os.ReadFile(state.RecordPath(root, c.ID))
	if err != nil {
		return !errors.Is(err, os.ErrNotExist), false
	}
	first, _, _ := strings.Cut(string(bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))), "\n")
	parts := strings.Split(strings.TrimSuffix(first, "\r"), " ")
	return true, len(parts) >= 2 && parts[0] != c.Hash && !c.NeverOutdate
}

// listCmd: fpm src/Handlers.cs:69-97, byte for byte: "<prefix> <id>[ (<title>)]\r\n"
// in manifest order, prefix "*" (up to date), "!" (outdated) or " ". The
// filter is the first argument, verbose the last one (Args.Last()).
//
// Unlike legacy fpm (exit 1 when offline, which hides the launcher's Manager
// button), a failed download falls back to a local copy and still exits 0.
func (a *app) listCmd(ctx context.Context, args []string) int {
	e, snap, err := a.loadForReading(ctx)
	if err != nil {
		return a.fail(err)
	}
	filter := ""
	if len(args) > 0 {
		filter = args[0]
	}
	verbose := len(args) > 0 && args[len(args)-1] == "verbose"
	var b bytes.Buffer
	for _, c := range snap.Manifest.All() {
		downloaded, outdated := legacyStatus(e.Root(), c)
		if (filter == "available" && downloaded) || (filter == "downloaded" && !downloaded) || (filter == "updates" && !outdated) {
			continue
		}
		prefix := " "
		if outdated {
			prefix = "!"
		} else if downloaded {
			prefix = "*"
		}
		b.WriteString(prefix + " " + c.ID)
		if verbose {
			b.WriteString(" (" + c.Title + ")")
		}
		b.WriteString("\r\n")
	}
	if _, err := a.stdout.Write(b.Bytes()); err != nil { // one write: the launcher reads it whole
		return exitError
	}
	return exitOK
}

// infoCmd: fpm src/Handlers.cs:99-128.
func (a *app) infoCmd(ctx context.Context, args []string) int {
	if len(args) == 0 {
		return a.fail(&usageError{"At least one argument is required"})
	}
	e, snap, err := a.loadForReading(ctx)
	if err != nil {
		return a.fail(err)
	}
	c := snap.Manifest.Components[args[0]]
	if c == nil {
		a.msg("Specified component does not exist")
		return exitError
	}
	downloaded, outdated := legacyStatus(e.Root(), c)
	updated := ""
	if !c.Modified.IsZero() {
		updated = c.Modified.Local().Format("1/2/2006 3:04:05 PM") // .NET en-US short date + long time
	}
	yesNo := map[bool]string{true: "Yes", false: "No"}
	a.out("ID:             %s", c.ID)
	a.out("Title:          %s", c.Title)
	a.out("Description:    %s", c.Description)
	a.out("Download size:  %s", engine.FormatSize(c.DownloadSize))
	a.out("Install size:   %s", engine.FormatSize(c.InstallSize))
	a.out("Last updated:   %s", updated)
	a.out("CRC32:          %s\r\n", c.Hash)
	if len(c.Depends) > 0 {
		a.out("Dependencies: ")
		for _, d := range c.Depends {
			a.out("  %s", d)
		}
		a.out("")
	}
	a.out("Required?       %s", yesNo[strings.HasPrefix(c.ID, "core-")])
	a.out("Downloaded?     %s", yesNo[downloaded])
	if downloaded {
		a.out("Up-to-date?     %s", yesNo[!outdated])
	}
	return exitOK
}

// splitIDs splits every argument on whitespace: the launcher passes several
// IDs as one argument ("a b").
func splitIDs(args []string) []string {
	var out []string
	for _, a := range args {
		out = append(out, strings.Fields(a)...)
	}
	return out
}

// changeCmd implements download, remove and update (fpm src/Handlers.cs:130-434).
func (a *app) changeCmd(ctx context.Context, args []string) int {
	ids := splitIDs(args)
	if a.cmd == "remove" && len(ids) == 0 {
		return a.fail(&usageError{"At least one argument is required"})
	}
	e, err := a.open(ctx)
	if err != nil {
		return a.fail(err)
	}
	defer e.Close()
	if recovered, err := e.Lock(); err != nil { // lock and recover before reading any state
		return a.fail(err)
	} else if recovered {
		a.msg("finished an update that was interrupted earlier")
	}
	snap, err := e.LoadManifest(ctx, engine.ForChanges)
	if err != nil {
		return a.fail(err)
	}
	for _, w := range snap.Warnings {
		a.msg("%s", w)
	}

	var known []string // unknown IDs are skipped with legacy fpm's message
	for _, id := range ids {
		_, ok := snap.Manifest.Resolve(id)
		if ok || a.cmd == "remove" && snap.Records[id] != nil { // deprecated, but removable
			known = append(known, id)
		} else {
			a.msg("Component or category %s does not exist and will be skipped", id)
		}
	}
	req := plan.Request{Install: known, AutoDeps: true}
	switch {
	case a.cmd == "remove":
		req = plan.Request{Remove: known}
	case a.cmd == "update":
		req = plan.Request{Update: known, UpdateAll: len(ids) == 0, AutoDeps: true}
	case len(ids) == 0:
		req.Install = snap.Manifest.Order
	}
	p := &plan.Plan{}
	if len(ids) == 0 || len(known) > 0 {
		if p, err = e.Plan(req); err != nil {
			return a.fail(err)
		}
	}
	for _, s := range p.Skipped {
		a.msg("Component %s %s and will be skipped", s.ID, map[plan.SkipReason]string{
			plan.SkipAlreadyInstalled: "is already downloaded", plan.SkipUpToDate: "is already up-to-date",
			plan.SkipNotInstalled: "is not downloaded"}[s.Reason])
	}
	for _, w := range p.Warnings {
		a.log.Info("plan", "warning", w)
	}
	if p.Empty() {
		a.out("%s: No components to %s", a.cmd, a.cmd)
		return exitOK
	}
	a.printPlan(p)
	if !a.confirm() {
		return exitOK
	}

	res, err := e.Execute(ctx, p, a.hooks(p))
	switch {
	case errors.Is(err, context.Canceled):
		a.msg("cancelled; nothing was changed (partial downloads are kept and resume next time)")
		return exitCancelled
	case errors.Is(err, txn.ErrRollbackFailed):
		a.msg("%v", err)
		a.msg("The Flashpoint folder may be inconsistent. Close all Flashpoint programs and run \"fpm recover\".")
		return exitError
	case err != nil:
		a.log.Error("command failed", "cmd", a.cmd, "err", err)
		a.msg("%s", strings.TrimSuffix(err.Error(), " (no changes were made)")) // said on the next line
		a.msg("No changes were made.")
		return exitError
	}
	msg := map[string]string{"download": "downloaded", "update": "updated", "remove": "removed"}
	n := map[string]int{"download": len(res.Installed), "update": len(res.Updated), "remove": len(res.Removed)}
	line := fmt.Sprintf("Successfully %s %d components", msg[a.cmd], n[a.cmd])
	if a.cmd == "download" && len(res.Updated) > 0 {
		line += fmt.Sprintf(" and updated %d components", len(res.Updated))
	} else if a.cmd == "update" && len(res.Installed) > 0 {
		line += fmt.Sprintf(" and downloaded %d components", len(res.Installed))
	}
	a.out("%s: %s", a.cmd, line)
	if len(res.Skipped) > 0 {
		a.msg("%d component(s) were skipped: %s", len(res.Skipped), strings.Join(res.Skipped, ", "))
		return exitPartial
	}
	return exitOK
}

func (a *app) printPlan(p *plan.Plan) {
	section := func(verb string, ids []string) {
		if len(ids) > 0 {
			a.out("%d component(s) will be %s:", len(ids), verb)
			for _, id := range ids {
				a.out("  %s", id)
			}
			a.out("")
		}
	}
	var inst, upd, rem []string
	var installSize, freed int64
	for _, c := range p.Install {
		inst = append(inst, c.ID)
		installSize += c.InstallSize
	}
	for _, c := range p.Update {
		upd = append(upd, c.ID)
	}
	for _, r := range p.Remove {
		rem = append(rem, r.ID)
		freed += r.InstallSize
	}
	section("updated", upd)
	section("downloaded", inst)
	section("removed", rem)
	switch {
	case a.cmd == "remove":
		a.out("Estimated freed size: %s", engine.FormatSize(freed))
	case a.cmd == "download" && len(upd) == 0:
		a.out("Estimated download size: %s", engine.FormatSize(p.DownloadBytes))
		a.out("Estimated install size:  %s", engine.FormatSize(installSize))
	default:
		a.out("Estimated download size: %s", engine.FormatSize(p.DownloadBytes))
		a.out("Estimated changed size:  %s", engine.FormatSize(p.DiskDelta))
	}
	a.out("")
}

// confirm asks legacy fpm's "Is this OK? [y/n]: " (Handlers.cs:205-212),
// unless --yes is given or stdin is not a terminal: automation such as
// flashpoint-web runs `fpm update` without answering, which hung forever.
func (a *app) confirm() bool {
	if a.opts.yes {
		return true
	}
	if !a.stdinTTY {
		a.msg("input is not a terminal; continuing without asking (use --yes to hide this notice)")
		return true
	}
	for {
		fmt.Fprint(a.stdout, "Is this OK? [y/n]: ")
		line, err := a.in.ReadString('\n')
		switch strings.TrimSpace(line) {
		case "y":
			return true
		case "n":
			return false
		}
		if err != nil {
			a.out("")
			return false
		}
	}
}

func (a *app) verifyCmd(ctx context.Context, args []string) int {
	e, err := a.open(ctx)
	if err != nil {
		return a.fail(err)
	}
	res, err := e.Verify(ctx, splitIDs(args))
	if err != nil {
		return a.fail(err)
	}
	bad := 0
	for _, r := range res {
		switch {
		case r.Err != nil:
			bad++
			a.out("! %s: %v", r.ID, r.Err)
		case len(r.Missing) > 0:
			bad++
			a.out("! %s: %d file(s) missing", r.ID, len(r.Missing))
			for _, f := range r.Missing {
				a.out("    %s", filepath.FromSlash(f))
			}
		default:
			a.out("* %s", r.ID)
		}
	}
	if bad > 0 {
		a.msg("%d of %d component(s) have problems; to repair one, remove it and download it again", bad, len(res))
		return exitError
	}
	return exitOK
}

func (a *app) recoverCmd(ctx context.Context) int {
	e, err := a.open(ctx)
	if err != nil {
		return a.fail(err)
	}
	defer e.Close()
	recovered, err := e.Lock()
	if err != nil {
		return a.fail(err)
	}
	if recovered {
		a.out("recover: finished an interrupted update")
	} else {
		a.out("recover: nothing to recover")
	}
	return exitOK
}

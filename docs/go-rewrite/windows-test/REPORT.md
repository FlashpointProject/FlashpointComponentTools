# Windows real-install test report

Tester: Claude (Windows PC), 2026-11-07. Written as the test runs.

## 1. Environment

- Windows 11 Education, build 26200 (10.0.26200), 64-bit. One display, 2560x1440, 100 % scaling (LogPixels 96). GPU: RTX 4090 (plus a Parsec virtual display adapter).
- `$FP` = `D:\Flashpoint`, "Flashpoint 14.0.3 Infinity - Kingfisher" (`version.txt`), `Launcher\Flashpoint.exe` file version 14.0.3.
- `$FP\Manager` at start:

  | Name | Length | LastWriteTime |
  |---|---|---|
  | FlashpointManager.exe | 1973248 | 2025-02-14 21:01:00 |
  | fpm.cfg | 81 | 2026-11-07 08:41:41 |
  | fpm.exe | 998912 | 2025-02-14 19:51:58 |

- `fpm.cfg` at start (note: line 1 is an absolute path, not `..\`; the stable URL is on `nexus-dev.unstable.life`):
  ```
  D:\Flashpoint
  https://nexus-dev.unstable.life/repository/stable/components.xml
  ```
- 229 records in `Components\`, no subfolders. No `Logs\` and no `Temp\` folder at start.
- Build tested: `dist\VERSION.txt` = `1.3-13-g2767963`; all four SHA256SUMS match. `fpm.exe version` = `fpm 1.3-13-g2767963 (windows/amd64, go1.27.3)`. Copied to `C:\fptest\bin\` and `C:\fptest\bin\selfupdate\`.
- Backup: `C:\fptest\backup\Manager` (3 files), `C:\fptest\backup\Components` (229 files), taken by an earlier session today on this same untouched install (verified identical with `diff -r` before starting).

## 2. Results

| Phase / step | Result | Detail |
|---|---|---|
| A `list` (5 variants) byte for byte | PASS | `fc /b`: no differences for `list`, `list verbose`, `list downloaded verbose`, `list available`, `list updates verbose`. Exit 0, stderr empty on both. Go ~0.28 s each. (No `core-database` `!` difference showed up: legacy didn't mark it outdated on this install.) |
| A `info core-launcher` | PASS (minor diff) | Exit 0, 0.3 s. Differs from legacy in date format and line endings (see UI remarks). |
| A `verify` | PASS | Exit 0, 0.4 s for 229 components, prints a `*` line per component. (Legacy fpm has no `verify`; it prints its help.) |
| A offline `--source https://localhost:1/...` | PASS | 0.09 s, exit 0, stdout identical to the online `list verbose`. stderr: `list: could not download the component list (download component list: Get "https://localhost:1/components.xml": dial tcp [::1]:1: connectex: No connection could be made because the target machine actively refused it.); using the offline copy components.bak` |
| A GUI start, finds Flashpoint | PASS | Opens in ~1 s, `D:\Flashpoint`, stable channel detected for the `nexus-dev.unstable.life` URL. Counts match `list downloaded` (229 installed, Extras 0/7). `a01-components.png` |
| A GUI expand / details / search | PASS | `a02`–`a05`. Ctrl+F focuses search; filtering works. |
| A GUI Updates view | PASS | "Everything is up to date". `a06-updates.png` |
| A GUI Settings | PASS | Radios, http custom URL rejected ("Custom URLs must start with https://"), Save enabled only on change, Check files: "All 229 installed components are complete." in ~1 s. `a07`–`a10` |
| A GUI Browse… | PASS (remark) | Standard Windows folder picker; opens in Documents instead of the current folder. Cancel/Esc leaves the setting alone. `a11-browse.png` |
| A GUI resize / maximise / min size | PASS | Min size 776x539 (window rect); maximise fills the work area correctly. Long titles ellipsized at min size. `a12`, `a13` |
| A GUI window icon / taskbar | FAIL (cosmetic) | Taskbar shows the Flashpoint icon, but the title bar shows the generic Windows app icon. See bug 6. |
| A GUI keyboard | PASS (remark) | Ctrl+F, Space on a selected row (toggles, `a14-space-toggle.png`), Esc in the confirm dialog (cancels) all work. Tab moves focus but no focus indicator is drawn anywhere, so it's impossible to tell what has focus. Esc in the search box does nothing (not clearing). |
| A text sharpness at 100 % | PASS | Crisp. |
| B1 GUI install `extra-youtubetemplates` (0.8 MB) | PASS | Confirm dialog sizes right (`b02`), Enter applies, done in <0.3 s (`b04-finished.png`). Files and `Components\extra-youtubetemplates` written; legacy `list downloaded verbose` shows it. |
| B2 GUI remove it | PASS | Files and the now-empty `Utilities\` folder removed, record gone, legacy agrees (229). `b05-removed.png` |
| B3 CLI `download` / `remove` | PASS | Exit 0 both. With stdin not a terminal: "input is not a terminal; continuing without asking". In a real console the `Is this OK? [y/n]:` prompt works for `n` and `y` (`b06`, `b07`). `remove core-launcher` → `cannot remove required component "core-launcher"`, exit 1. Unknown id / not-installed → message, exit 0 (same as legacy). |
| B3 note: CLI while GUI open | as designed | Every write command fails with `another Flashpoint Manager (or fpm) is running on this Flashpoint folder; close it and try again` (exit 1) while the GUI is open, even idle. `list` still works. |
| B4 Cancel during download | PASS | Removed `supportpack-shockwave` (101 MB download) first with the Go CLI (0.9 s), then installed it in the GUI and pressed Cancel at ~31 MB (`b09`). "Cancelled — Nothing was changed. Partial downloads are kept and continue next time." (`b11`). Snapshot of `FPSoftware\` + `Components\` (names, sizes, mtimes) identical before/after; only `Temp\fpm\cache\6baa6797-supportpack-shockwave.zip.part` (53 MB) + `.validator` added. |
| B4 Resume | PASS | Restarted: 150 ms later the progress already read 60.3 MB of 95.9 MB (`b12-resumed.png`), i.e. it continued from the 53 MB part (at 28 MB/s a restart couldn't be there). `manager.log` says nothing about resuming (no Info-level line exists for it), so this was checked via the `.part` size and progress. |
| B5 Kill during download | PASS | `Stop-Process -Force` mid-download → `.part` 68.8 MB kept; `fpm.exe list verbose` works in between (exit 0, only the Shockwave line differs). Restart resumed (72.6 MB after 150 ms, `b13`) and finished (`b14`). Cache emptied afterwards. Record byte-identical to the backup's, `fpm verify` OK, legacy `list verbose` byte-identical to the start. |
| B6 locked file, share mode None (doc's exact command) | **FAIL** | No "A file is in use" dialog: the removal fails at once with "The change failed — Couldn't connect to the Flashpoint server. Check your internet connection and try again. Nothing was changed." (`b15-file-in-use.png`). Nothing changed on disk. See bugs 2 and 3. |
| B6 locked file, share mode Read | PASS | With `[IO.File]::Open(..., 'Read', 'Read')` the dialog appears (`b16`), Retry while still locked shows it again (`b17`), Retry after `Close()` finishes (`b18`). |
| B6 Abort | PASS | Locked the 2nd file of the component (share Read) so the 1st was already moved to backup; Abort → "Cancelled because Utilities\YouTube Templates\YouTubeAnnotationPlayer.zip was in use. Nothing was changed." (`b20-aborted.png`). Both files' SHA-256 and the record identical to before; no leftovers in `Components\.fpm\txn`. |
| C rehearsal server | PASS | `rehearsal.ps1 -Zip C:\fptest\bin\core-manager.zip -Source https://nexus-dev.unstable.life/repository/stable/components.xml` (this install's own stable URL instead of the default `nexus-dev.flashpointarchive.org`). `core-manager.zip: hash 7c615adc, download-size 9571589, install-size 23147520`. HttpListener on 8642 started without admin. |
| C legacy fpm `source` + `list updates verbose` | PASS | `! core-manager (Manager)` only. |
| C old manager updates tab | PASS | Old manager (v1.3) relaunched itself from `%TEMP%\69McIKvK_FlashpointManager.exe`; Update Components lists only "Manager … +19.2M" (`c02-old-updates.png`). "Install selected updates": GET `/core-manager.zip`, done in ~2 s with no message or dialog; list becomes empty. |
| C result | PASS | `Manager\FlashpointManager.exe` and `fpm.exe` SHA-256 equal to the tested build; `fpm.exe version` = `fpm 1.3-13-g2767963 (windows/amd64, go1.27.3)`; `Components\core-manager` = `7c615adc 23147520` + the two exe lines. `list verbose` byte-identical to phase A's `go_list_verbose.txt` (`core-manager` was already `*` there). |
| C source back to stable | PASS | New fpm `source https://nexus-dev.unstable.life/repository/stable/components.xml` → `fpm.cfg` byte-identical to the backup. As expected the real list now offers `! core-manager (Manager)` (rehearsal hash ≠ real hash); not applied. |
| D1 launcher starts with the Go fpm | PASS | Starts normally; `fpm.log` shows its `list` call (0.25 s). Footer shows "Install 1 software updates" (the real `core-manager`, expected after C) and "25043 Game Info Updates Ready" (metadata, unrelated). `d01-launcher-started.png` |
| D2 launcher → "Install 1 software updates" (`/update /launcher`) | **FAIL** | The Go manager opens on Updates (1) (`d02-manager-from-launcher.png`) and is **killed ~2.4 s later when the launcher exits**; nothing reopens. No "manager exited" log line. See bug 1. |
| D2 `FlashpointManager.exe /update /launcher` started by hand, then closed | PASS | `manager.log`: `started the launcher exe=D:\Flashpoint\Launcher\Flashpoint.exe`, `manager exited code=0`; launcher comes up. |
| D3 on-demand install (`/notemp /download`) | PASS | Removed `supportpack-shockwave` (Go CLI), restarted the launcher, played "Bubble Bobble (Shockwave)": launcher asks "You are missing the Shockwave support pack … Download now?" (`d04`), Yes → small "Installing components" window (`d05-download-window.png`), exit after 5.7 s, the game starts in the Shockwave Projector (`d06-game-running.png`). `manager.log`: `args="[/notemp /download supportpack-shockwave]"`, `manager exited code=0`. Record byte-identical to the original; left installed (it was installed at the start). |
| D4 self-update to `-selfupdate` | PASS (minor issue) | CLI `source http://localhost:8642/...`, rehearsal with the selfupdate zip (`hash 5b432873`). Updates → Manager → apply: "The manager itself was updated. Restart it to use the new version." + "Restart manager" (`d09`). After restart: new pid, header `1.3-13-g2767963-selfupdate` (`d10`), `fpm version` = `fpm 1.3-13-g2767963-selfupdate (windows/amd64, go1.27.3)`, record `5b432873`. Restart race leaves the old exe's backup behind until the next run (bug 4). New exes get mtime +1 h (bug 5). |
| D4 back to the normal build | PASS | Same flow with `C:\fptest\bin\core-manager.zip`: `fpm 1.3-13-g2767963`, record `7c615adc`, exe SHA-256 equal to the build. Same leftover; the next `fpm recover` removed it. Source set back to stable (`fpm.cfg` byte-identical to the backup), server stopped. |

## 3. Bugs

### 1. The launcher kills the Go manager it opens with `/update /launcher` (release blocker)

- Steps: Go manager installed in `$FP\Manager\`; start the launcher; click the footer button "Install 1 software updates" (the launcher spawns `FlashpointManager.exe /update /launcher` and quits ~2.5 s later).
- Expected: the manager stays open; closing it reopens the launcher.
- Actual: the manager window appears on the Updates view, then disappears together with the launcher. No error, no "manager exited" line in `manager.log` (process terminated from outside), no crash in the Windows event log. 100 % reproducible (twice).
- Process trace (100 ms poll, `C:\fptest\watch.log`):
  ```
  09:20:48.441 + FlashpointManager.exe pid=3232 parent=10176 cmd=FlashpointManager.exe /update /launcher
  09:20:50.842 - Flashpoint.exe pid=13932
  09:20:50.974 - Flashpoint.exe pid=10176     <- launcher back end (node), the manager's parent
  09:20:50.975 - FlashpointManager.exe pid=3232
  ```
- Control run with the **legacy** C# exes put back temporarily, same click: `FlashpointManager.exe` (pid 10724) starts `C:\Users\D\AppData\Local\Temp\69McIKvK_FlashpointManager.exe /update /launcher` (pid 3260) and exits; the launcher exits at 09:22:11.2; **pid 3260 survives** and stays open.
- Cause (high confidence): the launcher back end is Node; libuv on Windows puts every child spawned without `detached: true` into a global job object with `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE` (and `SILENT_BREAKAWAY_OK`). When the launcher exits, the job closes and the manager is killed. The C# manager escapes by accident: its `%TEMP%` copy is a grandchild, and with silent breakaway it isn't in the job. `/notemp /download` is unaffected because the launcher stays alive waiting for it (D3 PASS).
- Suggested fix: on Windows, when started with `/launcher` (or whenever the process is in a job), re-launch itself once (e.g. `CreateProcess` with `CREATE_BREAKAWAY_FROM_JOB`, or simply re-exec: a child of a job process with silent breakaway isn't in the job) and exit the first instance before taking the lock. `/notemp` should keep the current behaviour, since the launcher waits for that exit code.
- Note: every process in this test session runs inside a job of the test harness too, so `IsProcessInJob` was true everywhere; the kill timing and the C# control run are what show the launcher's job.

### 2. A file opened without sharing isn't treated as "in use" (no Retry/Abort dialog)

- Steps: install `extra-youtubetemplates`; in another process `[IO.File]::Open("$FP\Utilities\YouTube Templates\YouTubeTemplate.7z", 'Open', 'Read', 'None')`; remove the component (GUI or `fpm -y remove extra-youtubetemplates`).
- Expected: retries for ~3 s, then the "A file is in use" dialog (GUI) / the locked-file advice (CLI).
- Actual: fails immediately (0.5 s), no retry, no dialog. Nothing is changed (good). CLI output:
  ```
  Applying changes...
  remove: check Utilities/YouTube Templates/YouTubeTemplate.7z: The process cannot access the file because it is being used by another process. (no changes were made)
  remove: No changes were made.
  ```
  `manager.log`: `level=WARN msg="the change failed" err="check Utilities/YouTube Templates/YouTubeTemplate.7z: The process cannot access the file because it is being used by another process. (no changes were made)"`
- Cause: `internal/txn/txn.go:282` `apply()` first does `tx.root.Lstat(op.Rel)`. On Windows that stat fails with `ERROR_SHARING_VIOLATION` when another handle has share mode 0 (the `os.Root` stat opens the file), and line 291 returns `check %s` straight away instead of going through `move()`/`isLocked`/`onLocked`. With share mode Read the Lstat succeeds, the rename fails, and the dialog works (B6 PASS rows). Suggested fix: in `apply`, treat `isLocked(err)` from the Lstat like a locked move (retry + `onLocked`), or skip the pre-stat and let `move` report.
- Relevance: antivirus scanners and some programs open files with no sharing; those would hit this path.
- The CLI also says "no changes were made" twice (the error suffix plus "No changes were made.").

### 3. Every OS error is shown as "Couldn't connect to the Flashpoint server"

- Steps: as bug 2, in the GUI. `b15-file-in-use.png`.
- Actual: "The change failed. Couldn't connect to the Flashpoint server. Check your internet connection and try again. Nothing was changed." for a local sharing violation.
- Cause: `internal/gui/gui.go:487` `plain()` uses `errors.As(err, &ne)` with `var ne net.Error`. `syscall.Errno` has `Timeout()` and `Temporary()` methods, so it satisfies `net.Error`; any error chain containing an `Errno` (sharing violation, access denied, disk full, path too long…; `txn.sysErr` unwraps `*PathError` to the bare `Errno`) is reported as a connection problem. Suggested fix: check for `*url.Error` / `*net.OpError` / `*net.DNSError` explicitly, or exclude `syscall.Errno` before the `net.Error` check.

### 4. Self-update restart: new instance can't delete the old exe's backup (minor)

- Steps: D4, press "Restart manager".
- Actual (`manager.log`, same millisecond sequence both times):
  ```
  09:27:44.064 msg="manager started" version=1.3-13-g2767963-selfupdate
  09:27:44.078 msg="finishing cleanup of a completed update" dir=Components/.fpm/txn/20261107T082703-0e1871b9
  09:27:44.079 level=WARN msg="update applied, but some leftovers could not be deleted; will retry next time" ... err="RemoveAll Components/.fpm/txn/20261107T082703-0e1871b9/backup: Access is denied."
  09:27:44.080 msg="manager exited" code=0          <- the OLD process
  ```
  The new process starts before the old one has exited; `backup/000001` (14.8 MB, the old `FlashpointManager.exe`, still mapped) can't be deleted. The next run (`fpm recover`) removes it, so it's self-healing, but it leaves a WARN in every self-update log and 15 MB until the next start.
- Suggestion: in the new instance, wait (bounded, ~2 s) for the parent pid to exit before recovery, or skip the WARN for a backup that's the running/previous exe.
- Nit: `fpm recover` then prints "recover: nothing to recover" although it just finished that cleanup.

### 5. Extracted files get mtimes shifted by the UTC offset (minor)

- The exes extracted by the Go manager have mtime `09:55:42` (an hour in the future at the time) while the zip entries say `2026-11-07 08:55:42` (DOS time, no extended-timestamp field) and the old C# manager extracted the same zip with `08:55:38`.
- Cause: `archive/zip` returns DOS times as UTC when there's no extended timestamp; `internal/txn/stage.go:175` passes `f.Modified` to `Chtimes` unchanged. Fix: for entries without an extended timestamp, reinterpret the wall-clock fields in `time.Local` (what Windows tools do).

### 6. Title bar shows the default Windows icon (cosmetic)

- Steps: start `FlashpointManager.exe`.
- Expected: Flashpoint icon in the title bar (and Alt+Tab).
- Actual: generic default application icon in the title bar; the taskbar button and Explorer show the Flashpoint icon correctly.
- Guess: `cmd/manager/winres/winres.json` stores the icon group under the name `APP`. Explorer/taskbar use the first icon group whatever its name, but Gio's Windows backend (`gioui.org/app/os_windows.go`) loads the window-class icon by resource **ID 1** (`LoadImage(hInst, iconID=1, IMAGE_ICON, ...)`), finds none and falls back to the default. Renaming the group to `"#1"` in winres.json (and regenerating the `.syso`) should fix it. (Not verified: Go isn't installed on this PC.)

## 4. UI and wording remarks

- `fpm info`: legacy prints `Last updated:   2025-10-28 13:25:26` (it follows the Windows short date/long time format; this PC is en-US with a custom `yyyy-MM-dd` pattern), Go prints `Last updated:   10/28/2025 1:25:26 PM` (fixed en-US). Legacy also mixes LF and CRLF line endings in `info` (lines after "Last updated" end in bare LF), Go uses CRLF throughout. Harmless unless something parses `info`.
- Category details: "Components  30, 30 installed" reads oddly (`a03-category-details.png`); "30 (all installed)" or "30 installed of 30" would be clearer.
- Selecting a row and then typing in the search clears the details pane (selection is dropped).
- Settings → Browse… starts in Documents, not in the current Flashpoint folder (`a11-browse.png`).
- No visible keyboard focus indicator on buttons, rows or radios after Tab.
- Maximised at 2560 px, the size column sits ~900 px away from the names (`a13-maximized.png`); a max width for the list or a wider details pane would read better.
- At min size the "Installed" badge truncates to "Inst…" / "In…" (`a12-minsize.png`). At the default size, a row showing "Will be removed" squeezes "Installed" down to "…".
- The component rows show the install size (Shockwave "222.1 MB") while the progress counts the download ("95.9 MB"); both are right, but the jump can surprise.
- After Cancel, the row says "Not changed" but keeps a half-filled red progress bar (`b11-cancelled.png`).
- Nothing at Info level in `manager.log` / `fpm.log` records what was installed, removed or updated, or that a download resumed; only failures are logged. A one-line "applied: +x −y ~z" would help support. Every start also logs the same four "depends on … which does not exist (ignored)" lines from the stable list.
- The old C# manager (phase C) installs the Go manager silently: no message, the list just empties. Fine for the migration, just worth knowing.
- `fpm download`/`remove` print "Successfully downloaded 1 components" (legacy wording, kept for parity).

## 5. Final state and restore

State left (2026-11-07 09:30):

- `D:\Flashpoint\Manager\` = Go `FlashpointManager.exe` (SHA-256 `46f8bc19…`) and `fpm.exe` (`e53ebb6d…`), build `1.3-13-g2767963` windows/amd64. `fpm.cfg` byte-identical to the start (`D:\Flashpoint` / `https://nexus-dev.unstable.life/repository/stable/components.xml`).
- `Components\`: all 229 records byte-identical to the backup except `core-manager` (rehearsal hash `7c615adc`). New: `components.bak` (offline copy) and `.fpm\` (lock + empty `txn\`). `fpm verify` OK.
- **The real stable list now offers `core-manager` as an update** (`! core-manager (Manager)`; the launcher shows "Install 1 software updates"). Applying it would put the old C# manager back; don't, unless that's wanted.
- `supportpack-shockwave` was removed and reinstalled several times during the test (B4/B5/D3); it's installed again with an identical record. Test components (`extra-youtubetemplates`, `extra-copyfolderstructure`) removed; `Utilities\` doesn't exist (as at the start).
- New folders: `D:\Flashpoint\Logs\` (manager.log, fpm.log) and `D:\Flashpoint\Temp\fpm\cache\` (empty). `Manager-go\` removed. No rehearsal server, lock holder or watcher running.
- Because of bug 1, the launcher's "Install … updates" / "Open Flashpoint Manager" buttons currently open the Go manager for ~2 s only. Start `Manager\FlashpointManager.exe` directly to try it.
- Kept: `C:\fptest\backup\` (original Manager + Components), `C:\fptest\bin\` (tested binaries), `C:\fptest\*.txt` (list outputs), `C:\fptest\watch*.log` (process traces).

Restore (undo everything), with the launcher and manager closed:

```powershell
$FP = "D:\Flashpoint"
$B = "C:\fptest\backup"
robocopy "$B\Manager"    "$FP\Manager"    /MIR
robocopy "$B\Components" "$FP\Components" /MIR
Remove-Item -Recurse -Force "$FP\Manager-go", "$FP\Components\.fpm", "$FP\Temp\fpm" -ErrorAction SilentlyContinue
```

## 6. Retest after the fixes (2026-11-07, build `1.3-15-g8f1acbc`)

Fixes in commit `8f1acbc` (see the commit message). Same PC, same install, built on this PC with `scripts/build.sh` (bsdtar fallback). `go vet`, `go test ./...` and `go test -race ./...` pass on Windows.

| Step | Result | Detail |
|---|---|---|
| C old C# manager installs the new build | PASS | Legacy `Manager\` + record restored from the backup, rehearsal with the new zip: the old manager installed it (`fpm 1.3-15-g8f1acbc`, record `fa6a23b8`). |
| D2 launcher footer button (`/update /launcher`) | PASS (bug 1 fixed) | Trace: `FlashpointManager.exe /update /launcher /relaunched` survives the launcher's exit (09:59:57.5) and stays open (`r-d2-manager-after-launcher-exit.png`); closing it starts the launcher, which stays up. A second cause showed up while fixing it: the launcher back end runs with `ELECTRON_RUN_AS_NODE=1`, the manager inherited it, and the launcher it started on exit ran as bare Node and quit after 130 ms. The manager now clears it at start. |
| D3 on-demand install | PASS | As before; `manager.log` now has `msg="changes applied" installed=[supportpack-shockwave] updated=[] removed=[] skipped=[]`. |
| B6 lock with share mode None (doc's command) | PASS (bug 2 fixed) | "A file is in use" dialog (`r-b6-share-none-dialog.png`), Retry while locked shows it again, Retry after release finishes. |
| D4 self-update and back | PASS (bug 4 fixed) | Restarted copy (`args=[/relaunched]`) opens after the old process exited (`.091` → `.112`) and finishes the cleanup; nothing left in `Components\.fpm	xn`. The old instance still logs one WARN at apply time (its own exe is in the backup); accurate, and cleaned up by the restart. |
| Bug 5 mtimes | PASS | Extracted exes carry their build time in local time (09:58), not +1 h. Real server zips keep their dates (e.g. YouTube Templates 2023-06-27). |
| Bug 6 icon | PASS | Flashpoint icon in the title bar. |
| Bug 3 wording | PASS (unit test) | Disk errors no longer read as "Couldn't connect" (`TestPlainErrors`). |
| UI remarks | PASS | Focus outline on rail/rows/checkboxes/arrows (`r-focus2.png`), "30 installed of 30", no stale bar on cancelled rows, no "Installed" next to "Will be removed", badges fit at min size. Browse… still opens in Documents (Gio's folder picker takes no start folder; skipped). |

Final state: `Manager\` = `1.3-15-g8f1acbc` (SHA-256 `64da8aee…` / `422f545b…`), `fpm.cfg` byte-identical to the start, records identical to the backup except `core-manager` (the real list offers it as an update; don't apply). Shockwave reinstalled with an identical record; test components removed.

## Notes for next time

- In this environment `NoDefaultCurrentDirectoryInExePath=1` is set, so `cmd /c "cd /d ... && fpm.exe ..."` from the doc fails with "'fpm.exe' is not recognized" (and `fc` then compares two empty files and says "no differences"!). Use `.\fpm.exe`. Always check output sizes.
- This install already had every theme, logo set and support pack; only 7 small Extras were not installed. For cancel/kill tests a ~100 MB pack has to be removed first (`supportpack-shockwave`: nothing depends on it, 101 MB download). The server delivered ~28 MB/s, so Cancel has to be clicked ~1 s after Apply.
- The doc's locked-file command (share mode `None`) hits bug 2; use share mode `Read` to see the intended dialog.
- `rehearsal.ps1` defaults to `nexus-dev.flashpointarchive.org`; pass `-Source` with the URL from `fpm.cfg` when it differs.
- The launcher's window title is "Flashpoint Launcher" (not the text drawn in its title bar). The old C# manager runs as `%TEMP%\<random>_FlashpointManager.exe`.
- SendKeys: `%` means Alt, so `%errorlevel%` must be sent as `{%}errorlevel{%}`. Guard every key send with a foreground-window check: once the window lookup failed and a few keystrokes went to another app.
- Windows Defender / SmartScreen didn't flag or block the unsigned Go binaries.
- No Syncthing conflict files appeared in the repo during the test.
- No computer-use tool was available; a small PowerShell helper (`System.Drawing` screenshots + `user32` `SetCursorPos`/`mouse_event` + `SendKeys`) worked fine for driving the Gio window.

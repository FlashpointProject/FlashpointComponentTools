# Flashpoint Component Manager v2 (Go) — Design Sketch

Status: draft, 2026-11-06. This replaces `FlashpointManager.exe`, `fpm.exe` and `FlashpointInstaller.exe` with one Go codebase.

Supporting research is in `docs/go-rewrite/research/`. That folder is excluded from git through `.git/info/exclude`, because it contains infrastructure and security details that should not go into this public repo.

| Report | Contents |
|---|---|
| `01-csharp-code.md` | Behavioural spec of the current C# apps, with file:line references |
| `02-manifest.md` | components.xml schema, statistics, how zips are served, the generator |
| `03-ecosystem.md` | Launcher contract, fpm CLI, other consumers, distribution |
| `04-infra.md` | Hosting, publish pipeline, planned infra changes (private) |
| `05-go-stack.md` | UI toolkit comparison and supporting libraries |
| `06-localrun.md` and `localrun/` | Docker + Wine setup running the old apps on macOS; 51 screenshots |
| `ui-spike/` | Throwaway Fyne and Gio tri-state tree prototypes |

---

## 1. What we learned (the constraints)

### 1.1 The surrounding contract we must not break

| Consumer | Contract | Notes |
|---|---|---|
| Launcher startup | `Manager/fpm.exe list verbose`, run with cwd = `Manager` | Prints one line per component: `<*|!| > <id> (<title>)`. Lines must end in **CRLF**, because the launcher strips 2 characters. Stdout must contain component lines only, so diagnostics go to **stderr**. A non-zero exit hides the "Open Manager" button and disables on-demand installs. The launcher's forced self-update depends on `core-launcher` being reported as `!`. |
| Launcher on-demand install | `FlashpointManager.exe /notemp /download "<id1 id2>"` | All IDs arrive as **one argv element**; we must split it on whitespace (C# has a bug here). The launcher waits for the process and treats exit 0 as success. It finds supportpacks by matching the component **title** to the platform name, so titles are effectively an API. |
| Launcher "Open Manager" | `FlashpointManager.exe /update /launcher` | Opens the Updates view and starts `Launcher/Flashpoint.exe` on exit. The launcher quits about 2.5 s after spawning it, so we must wait for and retry locked `Launcher/` files, and leave its kill-on-close job object (§4.6). |
| Disk state | `Components/<id>` | No extension. Line 1 is `<crc32> <install-size> <dep…>`, then FP-relative paths with backslashes, CRLF. "Installed" means this file exists. The third-party `flashpoint-web` reads these too. Also `Components/components.bak` (offline copy of the XML). |
| Config | `Manager/fpm.cfg` | Line 1 is the FP root (may be relative, e.g. `..\`; may be stale), line 2 is the manifest URL. |
| Distribution | GitHub `releases/latest/download/FlashpointInstaller.exe` | The website download button links here (~1.2M downloads), so the asset name is frozen. |
| Self-distribution | `core-manager` component (path `Manager`) | It ships **both** `FlashpointManager.exe` and `fpm.exe`. The old manager updates it by deleting the files in its record and then moving the new files in. |
| Manifest consumers | C# Manager/Installer, C# fpm, fpm-go, flashpoint-web | Old clients crash on **any new element or comment** under `<list>`. Only new attributes are safe to add. |

Implicit rules that live only in code and must be carried over:
- An ID starting with `core` means required.
- The installer leaves `extra*` unchecked by default.
- `core-database` never counts as "outdated".
- `download-size = 0` (or install-size 0) means a metadata-only component with no zip.
- Zip URL = `list@url + id + ".zip"`.
- `hash` = CRC32 of the **whole zip file** (verified against a real zip).

### 1.2 What is wrong today (reliability)

- **No integrity check.** `hash` is only used as a version string, and neither size nor CRC is ever verified.
- **Downloads.** Each zip is held entirely in memory (about a 2 GB ceiling). There is no resume, no read timeout, and the download hangs forever on network loss (observed for 3.5+ minutes).
- **Not transactional.** The manager deletes old files, then moves new ones in. A failure in that window leaves a half-updated install. Component records are appended line by line during extraction, so a crash leaves a record that looks installed. `Temp/` is never cleaned up.
- **No shared-file safety.** Removing a component deletes every file in its record, even files another component also owns.
- **Robustness.**
  - Unhandled exceptions in `async void` handlers crash the app: corrupt zip, cancel during extraction, missing Content-Length, a dependency cycle, an XML comment.
  - The installer hangs forever if the VC++ redist fails to start or the UAC prompt is declined.
  - Message boxes are shown from worker threads.
- **UX bugs.**
  - The update total never resets (922 MB grows to 1.8 GB), and deprecated items are added instead of subtracted.
  - Install size is shown as "download size".
  - Pressing Enter in the path box starts the install.
  - Paths are not validated.
  - The offline indicator is never visible.
  - Cancelling at the download-failed prompt leaves the dialog stuck.
- **Security.**
  - No signatures anywhere: the manifest, zips and executables are all unsigned.
  - The manifest's `path` and IDs are used unsanitized in filesystem paths, so the manifest can write anywhere.
  - The VC++ redist is reused from a predictable `%TEMP%` path without verification.
  - Self-update relaunches from `%TEMP%`.
  - It silently falls back to `components.bak`, unsigned.
  - The manager is effectively the **root of trust** for the launcher, game server, `services.json` (which the launcher executes) and the database.
- **Server side.**
  - The generator validates nothing: 4 dependencies point to components that don't exist, and the development manifest is stale relative to its zips.
  - Zip builds are non-deterministic, so a title-only change still shows up as an "update".
  - Cloudflare caches zips for 4 h at a mutable URL, so a cached old zip can arrive alongside a new manifest.

### 1.3 Infrastructure facts that shape the client

- The manifest and zips come from **Nexus** raw repos: `nexus-dev.flashpointarchive.org` (behind Cloudflare) and `nexus-dev.unstable.life`, which skips Cloudflare. download.flashpointarchive.org is *not* the component source.
- Range requests work (206) and ETag is the file's sha1. The Nexus REST API already exposes sha256 for every blob.
- Everything runs on a single origin server. Nexus is a single replica and restarts on certificate renewals. Nexus Community Edition has a request budget of about **100k requests per day** that may already be exceeded.
  - Patient retries with backoff.
  - One manifest GET (conditional) plus one GET per needed zip: no polling, no HEAD storms.
  - An identifiable User-Agent, because Bot Fight Mode / CrowdSec are being considered.
  - No certificate pinning (certificates rotate between Let's Encrypt and Google Trust Services). Build with Go ≥ 1.27.3 / 1.26.10, after the eMudhra CA distrust.
- Proxying `nexus-dev.unstable.life` through Cloudflare is proposed. Clients should therefore keep a **mirror list** rather than a single hard-coded host.

---

## 2. Tech stack

| Concern | Choice | Why |
|---|---|---|
| Language | Go ≥ 1.27.3 | Single static binary, simple cross-compilation, the project already uses Go (GameServer, UltimateUpdater, FPFSS). |
| GUI | **Gio** (`gioui.org` v0.10.x + `gioui.org/x`) | Pure Go on Windows (`CGO_ENABLED=0`, cross-compiles in seconds, 6.9 MB exe that imports only kernel32). Renders with D3D11 feature level 9_1, which still works on machines stuck on the "Microsoft Basic Display Adapter"; Fyne needs OpenGL 2.1 and fails there. Native folder picker. It is the only toolkit that keeps the door open for a Win7 build via a Go fork. |
| GUI fallback | Fyne v2.8 | The most popular pure-Go toolkit, with tree / tri-state checks / dialogs built in. Choose it if we decide old-GPU Windows doesn't matter and want faster UI work; the cost is CGO on every OS and zig/fyne-cross for Windows. The core is UI-agnostic, so switching costs only the `ui/` package. |
| CLI | stdlib `flag`, or `kong` | `fpm` must be byte-compatible; keep the parser dumb and explicit. |
| HTTP | stdlib `net/http` + our own ~200-line downloader | Range/If-Range resume, stall timeout, backoff with jitter, mirror failover. Libraries (grab, retryablehttp) don't cover resume + verification + mirrors well together. |
| Zip | stdlib `archive/zip` + `os.Root` (Go 1.24+) | Zip-slip prevention enforced by the OS layer; entry CRCs verified automatically. Optionally use `klauspost/compress/flate` for speed. |
| Hashes | stdlib `hash/crc32`, `crypto/sha256` | Computed while streaming via `io.MultiWriter`. |
| Signatures | Ed25519 in minisign format (`aead.dev/minisign`) | Small, auditable, offline key, easy to use in CI (`minisign -S`). Embed 2 public keys for rotation. |
| Single instance | `gofrs/flock` on `Manager/.fpm.lock` | Cross-platform; prevents launcher-spawned and user-spawned managers running at the same time. |
| Windows extras | `golang.org/x/sys/windows` | ShellExecute `runas` only when a folder isn't writable, `.lnk` creation through IShellLink via a small COM shim (or `go-ole`). |
| Logging | stdlib `log/slog` → `Logs/manager.log` (rotated) | Today there are no logs at all. |

Platform floor (**decided 2026-11-06:** Windows 7/8.1 support is dropped, in line with the new launcher):
- Windows 10+, macOS 12+ (Go 1.26) or 13+ (Go 1.27), Linux glibc from Debian 11/Ubuntu 22.04.
- **Windows architecture:** ship **windows/amd64** only (**decided 2026-11-07:** 32-bit is deprecated in the launcher, so the manager doesn't need it either).

---

## 3. Architecture

```
cmd/
  fpm/            console binary  -> fpm.exe              (CGO-free, no UI linked)
  manager/        GUI binary      -> FlashpointManager.exe (-H windowsgui)
  installer/      GUI binary      -> FlashpointInstaller.exe (same UI pkg, install mode)
internal/
  manifest/       parse v1 XML + v2 JSON -> model; validate; legacy XML writer (for tools)
  trust/          minisign verify, embedded keys per channel
  state/          Components/<id> reader/writer (legacy format), fpm.cfg, root discovery
  plan/           selection -> Plan{install, update, remove}; dependency closure, reverse-dependency
                  checks, refcounts, download-size vs disk-delta, free-space check
  fetch/          downloader: cache, resume, verify, retry/backoff, mirrors, rate/stall control
  txn/            transactional apply with on-disk journal; crash recovery
  engine/         orchestrates fetch -> stage -> commit; emits Events; context cancel; Prompt callbacks
  platform/       per-OS: shortcuts, file-lock diagnostics, elevation, self-replace, open launcher
  ui/             Gio views (tree, progress, dialogs, settings); talks only to engine
tools/
  fpm-publish/    CI helper: build sidecar (crc32+sha256), regenerate XML + JSON v2, validate, sign
```

The engine API is the seam between core and UI. The CLI and GUI both drive the same engine:

```go
type Engine interface {
    Load(ctx context.Context) (*Snapshot, error)       // manifest (+trust status), local state, diffs
    Plan(sel Selection) (*Plan, error)                 // resolved, validated, sized
    Execute(ctx context.Context, p *Plan, ui Hooks) error
    Verify(ctx context.Context, ids []string, deep bool) (*Report, error) // "Check files"
    Recover(ctx context.Context) error                 // replay or roll back an interrupted txn
}
type Hooks interface {
    Progress(Event)                                     // typed events: Queued, Downloading{bytes,total,rate},
                                                        // Retrying{in}, Verified, Extracting, Committing...
    OnError(ctx context.Context, e *OpError) Decision   // Retry | Skip | Abort (CLI: policy flags; GUI: dialog)
    OnLocked(ctx context.Context, f LockedFile) Decision // "Launcher.exe is in use by Flashpoint (pid 1234)"
}
```

### 3.1 Data model

```go
type Component struct {
    ID, Title, Description string
    Path                    string   // validated relative path, "" = FP root
    Depends                 []string
    DownloadSize, InstallSize int64
    CRC32                   uint32
    SHA256                  []byte   // nil when only the v1 XML is available
    Modified                time.Time
    Required, DefaultOn     bool     // v2 explicit; v1 derived from "core"/"extra" prefixes
    UpdatePolicy            Policy   // Normal | Never (core-database) | Prompt
    Platforms               []string // v2; v1 = ["windows"]
    Deprecated              bool
}
```

The v1 prefix rules live in **one adapter function** instead of being scattered across forms.

---

## 4. Core algorithms

### 4.1 Startup and root discovery
1. If `--path` is given, use it.
2. Otherwise use `fpm.cfg` line 1, resolved relative to **the exe's directory**. (C# fpm used the exe directory and the C# GUI used the cwd; the launcher sets cwd = Manager, so both agree there.)
3. Then try `..`, then the installer's path argument.
4. A root is valid if it contains `Components/` (or `Launcher/` plus `Data/`).
5. If the stored path is invalid but `..` is valid, **repair `fpm.cfg`**. This fixes the common Ultimate `C:\Flashpoint` case.
6. Take the single-instance lock. Run `txn.Recover()` before anything else.

### 4.2 Manifest loading (fast path for `fpm list`)
- The launcher's startup is blocked on `fpm list`, so use a 3 s budget.
  1. A plain GET of the manifest (and of its `.minisig` when the channel has keys).
  2. On timeout or failure, use `Components/components.bak` (the one offline copy, shared with the legacy tools; rewritten only when the bytes change) and **still exit 0**, with a stderr note. Today fpm exits 1 offline, which hides the Manager button.
- Verification (as built in M2): when keys are embedded for the channel, the signature must exist and verify, and its trusted comment must name the channel. An unverifiable offline copy is then only good for reading, not for changes. With no keys embedded (today) nothing is checked.
  - *Deferred to M4:* anti-rollback (`sequence`) and `expires`. Without them a signed but older manifest could be replayed; it can only point at older genuine components. The M1 store for this was deleted in the M2 review as unused weight.
- Validate everything:
  - IDs must match `^[a-z0-9]+(-[a-z0-9_.]+)*$`.
  - `path` must satisfy `filepath.IsLocal` after normalising `\`.
  - No duplicate IDs.
  - Dangling dependencies → warning, and the component is installable without them (matching today's behaviour), logged.
  - Dependency cycles → reject the cycle members (C# overflows the stack).

### 4.3 Planning
- The user's selection becomes the transitive dependency closure, followed by validation:
  - Unchecking a dependency that another selected/installed component needs gives a clear "X is needed by A, B: remove them too?" (cascade offered, not silently refused).
  - Required components cannot be removed.
  - Platform-incompatible components are hidden.
- Sizes are shown separately:
  - **Download** = sum of download sizes not already in the verified cache.
  - **Disk change** = new install sizes − old install sizes − removed.
  - **Free space required** = staging (download + extracted) + headroom. This is checked against `statfs`/`GetDiskFreeSpaceEx` *before* starting.
- Reference counts: build `path → {owners}` from all records. Removing component A deletes only the paths no other installed component still claims.

### 4.4 Fetch (phase 1, cancellable, never touches the live install)
- Cache: `<FP>/Temp/fpm/cache/<sha256-or-crc>-<id>.zip` (`*.part` while downloading). Being on the same volume as the install makes the later renames atomic and keeps big files off a small system drive.
- Each zip:
  - **URL:** `mirror + id + ".zip?h=<hash>"`. The `?h=` query busts Cloudflare's 4 h cache of the mutable URL while needing no server change; a server that ignores the query still serves the file.
  - **Resume:** `.part` + `Range` + `If-Range: <etag>`.
  - **Verification:** size and CRC32 (+ SHA256 when available), computed while streaming. On mismatch: delete the file, retry once from scratch, then try the next mirror, then report.
  - **Timeouts:** connect 15 s; **stall** if under 1 KB in 30 s. This also fixes the C# app's infinite hang.
- Retries: exponential backoff with full jitter, rotating mirrors on every failure; a mirror that gives a final answer (404, size or hash mismatch) is dropped. Progress resets the budget, so a flaky link still finishes a 330 MB zip. A 1-minute budget without progress rides out a short Nexus restart; when the manifest itself came from the offline copy, the first failure asks at once. After that, `OnError` → Retry / Skip this component (and its dependents) / Abort. The GUI shows "retrying in 8 s…" instead of a modal on the first blip.
- Concurrency: **2 parallel downloads**. That saturates typical links without hammering a single origin.
- Cancel: cancel the context; keep `.part` files for resume next time.
- Metadata-only components (size 0): no fetch, record only.

### 4.5 Stage + commit (phases 2–3, transactional)
Only after **all** downloads are verified:

1. **Stage:** extract each zip into `Components/.fpm/txn/<txid>/new/<component path>/…` through `os.Root`, so zip-slip, absolute paths and `..` are impossible.
   - Reject symlink entries.
   - Cap the total uncompressed size against `install-size` × 1.1, and also cap the entry count.
   - Entry CRCs are verified by `archive/zip`.
   - Build the new file list.
2. **Journal:** write `Components/.fpm/txn/<txid>/journal.json` (fsync): the ordered `(dest, staged, backup, existed)` moves and the removals. State: `committing`. Rollback works from what is actually on disk, and never moves a live file the transaction didn't place.
3. **Commit, per file:**
   1. If the destination exists, move it to `txn/backup/…`.
   2. Move the staged file into place (a rename on the same volume).
   3. On a Windows sharing violation (antivirus, a running launcher), retry for about 3 s, then call `OnLocked` ("close Flashpoint and any program using X"): retry or abort.
4. **Removals:** move files to `backup/`; don't delete them yet.
5. **Records:** write each `Components/<id>` as temp + rename, in the exact legacy format (CRLF, backslashes, `crc size deps` header). Delete the records of removed components. State: `committed`.
6. **Cleanup:** delete `backup/`, staged files and the journal, then prune empty directories (only ones we emptied).
7. **Recovery:** on any failure before `committed`, roll back by moving backups to their original locations. On startup, `Recover()` replays the journal: before `committed` it rolls back, after `committed` it finishes the cleanup. An interrupted update therefore never leaves a half-applied install.

Cancel is allowed until step 3 starts, and then disabled. Commit is fast because it is only renames.

### 4.6 Self-update (`core-manager`)
- On Windows a running exe can be **renamed** but not overwritten. During commit, for our own exe:
  1. Rename `FlashpointManager.exe` → `FlashpointManager.exe.old`.
  2. Place the new exe.
  3. Delete `*.old` on the next start.
  4. Offer "Restart manager".
- This replaces the `%TEMP%` copy-and-relaunch for updating. The manager still relaunches itself once (marker `/relaunched`) when it starts inside a job object without `/notemp`: the launcher (Node/libuv) spawns it in a kill-on-close job and exits ~2.5 s later, which would kill it. The C# manager escaped only through its `%TEMP%` relaunch. With `/notemp` the launcher waits for the exit code, so no relaunch.
- The relaunched or restarted copy waits (≤3 s) for its parent to exit before recovery, so the old exe's backup can be deleted.
- `ELECTRON_RUN_AS_NODE`, inherited from the launcher's back end, is cleared at start; otherwise the launcher started on exit runs as bare Node and quits.
- On macOS, replace the whole `.app` bundle, so code signatures stay valid.

### 4.7 Verify / repair ("Check files")
- **Quick:** do all listed files exist?
- **Deep (new):** compare against the per-file hashes from the v2 manifest (`files` list, optional), else re-download and compare.
- Repair = reinstall the affected components through the normal transaction.
- Unlike today, a pending update doesn't block repair; repair just installs the current version.

### 4.8 Installer mode
Same GUI with the install flow:
1. Pick the destination. Validation: path characters, length, not under Program Files / Temp / OneDrive / a synced folder, writable, free space. Pressing Enter does **not** start the install.
2. Choose components.
3. Install into the empty destination through the same engine (phase-1 cache goes into `<dest>/Temp/fpm`).
4. Write `Manager/fpm.cfg` with `..\` (relative, which survives moving the folder).
5. Create shortcuts.
6. Optionally launch.

On failure or cancel, delete only what *we* created. If the folder was non-empty, nothing pre-existing is touched, because we refuse to overwrite foreign files without explicit confirmation.

VC++ redist:
- Download to a fresh private temp directory and verify its Authenticode signature (WinVerifyTrust, signer must be Microsoft) before running it.
- A missing/declined UAC prompt gives a clear warning with "Continue without" / "Retry". It never hangs.
- Longer-term, ask whether a component can ship the needed DLLs app-locally.

---

## 5. CLI (`fpm`) — compatible superset

Kept byte-compatible:

```
fpm list [available|downloaded|updates] [verbose]   # "<*|!| > id (title)\r\n", stdout only
fpm info <id>
fpm download [ids|categories...]
fpm remove <ids...>
fpm update [ids...]
fpm path [value]
fpm source [value]
```

New, additive:
- `-y/--yes` (no stdin prompt; flashpoint-web and other automation hang on `[y/n]` today).
- `--json` output.
- `verify [--deep] [ids]`.
- `recover`.
- `--channel stable|development`.
- `--path`.
- `--parallel N`.

Exit codes: `0` success, `1` failure, `2` usage error, `3` partial (some components skipped), `4` cancelled. Prompts are skipped automatically when stdin is not a TTY, with a stderr notice.

`FlashpointManager.exe` accepts the legacy switches `/update`, `/launcher`, `/download <ids…>` (split on whitespace **and** across args) and `/notemp`, plus `--path`.
- `/download` without a window: show a small progress-only window (the user should see a 300 MB download happening). Exit 0 only if every requested ID is installed.

---

## 6. GUI sketch

One window (about 900×600, resizable, HiDPI) with a header (logo, channel badge, "offline / unsigned" banner when applicable) and four views on a left rail:

```
+--------------------------------------------------------------------------+
| [FP] Flashpoint Manager           channel: Stable  ● online, signed      |
+-----------+--------------------------------------------------------------+
| Components|  [Search…………………]   [ ] Show installed only                |
| Updates(9)|  [▣] Core Components (required)                 372 MB       |
| Repair    |   ├─[■] Launcher           installed  ✓ up to date            |
| Settings  |   └─[■] Database           installed  ! update  +4 MB         |
|           |  [▣] Support Packs                                            |
|           |   ├─[☐] Shockwave          96 MB  needs: Common Secureplayer  |
|           |  ------------------------------------------------------------ |
|           |  Description: …                                               |
+-----------+--------------------------------------------------------------+
| Download 412 MB · Disk +1.1 GB · Free 38 GB    [Open Launcher] [Apply ▶] |
+--------------------------------------------------------------------------+
```

- **Components:** a tri-state tree combining add and remove. Inline badges show installed / update / deprecated / required. Dependency hints. The details pane shows title, description, sizes, date, dependencies and dependents.
- **Updates:** a flat list with per-item checks, "select all", deprecated items marked "remove". The totals are recomputed from the plan every time, which fixes the never-reset total.
- **Progress** (replaces the view while running):
  - overall bar plus a per-component list with state (queued, downloading at X MB/s, retrying in 8 s, verified, extracting, done, skipped);
  - ETA;
  - Cancel (disabled once commit starts, with a label explaining why);
  - "Show log".
  - Errors appear inline with Retry / Skip / Abort. A modal appears only when auto-retries are exhausted.
- **Repair:** quick/deep scan with results and "Repair selected".
- **Settings:** FP root (folder picker + validate), channel (Stable / Development / Custom URL, https only unless an advanced toggle is set), parallel downloads, open logs folder, uninstall Flashpoint (with an explicit "delete everything including user data" checkbox, off by default).
- **Footer:** **"Close and open Launcher"** (issue #9), always visible.
- **Installer** wizard reuses the components view: Welcome/destination → Components → Progress → Finish (shortcuts, launch).

---

## 7. Manifest v2 and publishing (server side)

Publish **both** formats from one generator run. Upload order: zips first, then XML, then JSON + signature last.

```json
{
  "schema": 2,
  "channel": "stable",
  "sequence": 1042,
  "generated": "2026-11-06T12:00:00Z",
  "expires":   "2027-02-06T12:00:00Z",
  "mirrors": ["https://nexus-dev.flashpointarchive.org/repository/stable/",
              "https://nexus-dev.unstable.life/repository/stable/"],
  "categories": [{"id": "core", "title": "Core Components", "description": "…",
                  "children": [/* categories or components */]}],
  "components": {
    "core-launcher": {
      "title": "Launcher", "description": "…", "path": "Launcher",
      "depends": ["core-configuration"], "platforms": ["windows"],
      "download_size": 239900000, "install_size": 612000000,
      "crc32": "1a2b3c4d", "sha256": "…", "modified": 1761609600,
      "required": true, "default": true, "update_policy": "normal",
      "file": "core-launcher.zip"
    }
  }
}
```

- Detached signature in `components.json.minisig`, signed offline or in a protected CI job. Two embedded public keys allow rotation.
- **XML compatibility:** add new *attributes* only (`sha256="…"`). The Go client also reads them from XML during the transition. Never add elements or comments.
- **Generator hardening** (replace or extend `merge-metadata.py`):
  - compute sha256 (the Nexus API already has it);
  - fail on dangling dependencies, cycles, bad IDs/paths, or size/hash mismatches against the actual blobs;
  - make zips reproducible (sorted entries, fixed mtimes, `7z`/`zip -X`) so title-only changes don't create fake updates;
  - pin CI dependencies (actions, cloned tools) by commit.
- **Later:** content-addressed zip names (`<id>-<sha12>.zip`) for perfect cache behaviour; per-platform components (`platforms`) for Linux/macOS; an optional per-file hash list for deep verify.

---

## 8. Migration plan

1. **Phase 0, server side (no client change):**
   - fix the generator and the 4 dangling dependencies;
   - add `sha256` attributes;
   - start publishing `components.json` (+ signature once keys exist) to development, then stable;
   - confirm Range/ETag through Cloudflare on both hosts.
2. **Phase 1, Go `fpm` + Manager on the development channel:**
   - Publish `core-manager.zip` containing `fpm.exe` + `FlashpointManager.exe` (Go). Never include `fpm.cfg`.
   - The old manager installs it through its normal update path; it runs from the `%TEMP%` copy, so the overwrite works.
   - Gate: the launcher-contract tests in §9 pass on real Windows 10/11 + Wine, and `fpm list verbose` output is byte-diffed against the C# fpm on the same install.
3. **Phase 2, stable `core-manager`:** promote after a soak period on development.
4. **Phase 3, Installer:** replace the `FlashpointInstaller.exe` release asset (same name). Keep v1.3 attached to an archived release for legacy OSes.
5. **Phase 4, mandatory signatures:** a release with `RequireSignedManifest=true`.
6. **Phase 5, Linux/macOS:**
   - per-platform components in v2;
   - a launcher patch so it calls `path.join(cwd, 'fpm' + ext)`, because a bare `fpm.exe` resolves via PATH on Unix;
   - signed/notarized macOS build;
   - Linux tar.gz/AppImage.

**Risk: the push is unconditional.** The old manager can't select by OS. If the Go binary can't start (Win7/8.1 or a 32-bit OS, both accepted because the new launcher drops them too), the launcher's `fpm list` fails. That hides the Manager button and also stops launcher updates, and the user has no GUI left to roll back with. Mitigations:
- (a) push the Go manager only together with (or after) the launcher release that drops 32-bit and Win7/8.1, so nobody loses a working setup earlier than the launcher already takes it away;
- (b) announce the Win10 floor together with the launcher release and keep the C# v1.3 manager/installer downloadable from an archived GitHub release;
- (c) consider keeping the C# v1.3 manager inside the zip as `Manager/legacy/` for one release.

---

## 9. Testing and release

- **Unit:**
  - manifest parsing against the real stable/dev/testing snapshots (including the dangling-dependency and size-0 cases);
  - plan/refcount property tests;
  - `Components/<id>` golden files written by C#;
  - an `fpm list verbose` golden test, plus a test that runs the launcher's actual TypeScript parsing logic against our output.
- **Fault injection:** a local HTTP test server that drops connections mid-body, stalls, returns 5xx / 404 / 416 / wrong ETag / truncated or bit-flipped zips, or is slow (10 KB/s). Assert resume, mirror failover, the hash-mismatch path and no hangs.
- **Crash consistency:** kill the process at every journal step (an env var hook) and assert that `Recover()` restores either the old or the new state exactly, never a mix.
- **E2E:**
  - GitHub Actions `windows-latest` running real install → update → remove → repair against a fixture manifest served from the runner;
  - a contract test that spawns the commands exactly as the launcher does;
  - the Docker + Wine setup in `research/localrun/` gives Mac developers the same loop locally (note: Rosetta crashes about 1 in 4 runs; use an x86 Linux host or CI for reliable runs).
- **CI/release:**
  - Windows from ubuntu (`CGO_ENABLED=0 GOOS=windows GOARCH=amd64`, `-H windowsgui` for the GUI);
  - macOS on a macOS runner;
  - Linux on ubuntu-22.04 (low glibc);
  - reproducible builds (`-trimpath`, pinned toolchain), SBOM, checksums file signed with the release key;
  - Authenticode via SignPath Foundation (free for OSS), never UPX;
  - Apple notarization with `rcodesign` when macOS ships.

---

## 10. Milestones

1. **M1, core + `fpm` parity:** manifest v1, state, plan, fetch, txn, CLI, golden/launcher-contract tests. Shippable as `fpm.exe` alone on the development channel.
2. **M2, GUI manager:** Components/Updates/Repair/Settings, legacy switches, self-update.
3. **M3, installer mode:** VC++ handling, shortcuts, GitHub release.
4. **M4, trust:** v2 JSON + minisign, generator changes, anti-rollback, then mandatory signing.
5. **M5, cross-platform:** platform-tagged components, launcher patch, mac/linux packaging.

## 11. Open decisions

1. ~~**Windows floor.**~~ Decided: Windows 10+.
2. ~~**Gio vs Fyne.**~~ Decided: Gio (a hand-written tri-state tree is a trivial cost).
3. **Signing keys.** A signature is only as strong as the place the private key lives: whoever holds it can make every manager install anything. Proposed:
   - **Stable key:** an offline minisign key held by 1–2 maintainers (plus an encrypted backup). Used when a development manifest is *promoted* to stable, which is already a manual step, so signing adds one command there.
   - **Development key:** a separate key, kept as a CI secret in a protected environment. The manager only accepts it for the development channel, so a CI compromise can't reach stable users.
   - **Rotation:** the binary embeds two public keys per channel (current + next).
   - **Before relying on signatures,** review who and what can publish to the repository (upload credentials, the deploy pipeline).
4. **Is the Nexus publish freeze lifted, and how is promotion to stable done?** (Manual today and undocumented.)
5. ~~**Repository layout.**~~ Decided: Go module at the root of this repo on the `go-rewrite` branch, next to the C# apps for comparison. The C# apps are deleted in the PR that switches over.
6. **Shared VC++ redist.** Keep the system install, or ship the DLLs app-locally inside the components that need them?

## 12. M1 status (2026-11-06)

M1 is done: the `go-rewrite` branch has the core packages plus `cmd/fpm`, about 5.3k lines without tests.

**Verified against the legacy tools** under Wine (amd64 build):
- `fpm list` (every variant) is byte-identical to the legacy C# fpm, and the launcher's parser gives identical results.
- Records written by Go and by the legacy tools are byte-identical, and each side reads the other's state.
- Downloads resume after a kill. Locked files are retried, then rolled back cleanly. Self-update of the running `fpm.exe` works.
- The Windows test binaries pass under Wine.

**Deliberate differences from legacy fpm:**
- Offline `list` exits 0 using the cached manifest.
- A missing zip makes the whole operation refuse with "no changes were made" (legacy crashed half-way).
- Prompts are skipped when stdin isn't a terminal.
- `remove` refuses required components and components that others depend on.

**Hardening included:**
- Path validation of manifest paths, IDs and zip entries.
- A write policy: no archive entry may land in `Components/` or on `Manager/fpm.cfg` (M2 cut the per-component layout rules: the server already controls core-manager and core-launcher, so they protected nothing).
- `Favorites.json` is preserved, and `?h=` cache-busting is added to zip URLs.
- Verified downloads: size and CRC32 always, sha256 when present and required for signed manifests.
- Minisign verification with per-channel keys (M2: verify-only, see §4.2). It is inactive until keys are embedded.

**Known gaps (next):**
- Untested on real Windows. Test it there before shipping.
- No Windows system (WinINET) proxy support; only the proxy environment variables are used.
- Staged files aren't fsynced, so a power cut mid-update may need a repair run.
- No deep verify or repair command yet.
- The GUI is M2.

## 13. M2 status (2026-11-07)

M2 adds the GUI (`internal/gui`, Gio) and `FlashpointManager.exe` (`cmd/manager`) with the legacy switches `/update`, `/launcher`, `/download "a b"`, `/notemp`, self-update restart, "Check files" repair and settings. It was visually tested (rendered screenshots of every view, plus Wine 11.21 against the legacy app). A critical whole-codebase review followed, then a simplification pass:

- **Removed as unneeded weight:** the trust store (rollback, expiry, downgrade), the private manifest cache and conditional GET, `Retry-After` and attempt counting, the per-component write layout, `fpm --offline`, manager `--version`, and two copies of the log-file code (now `internal/logfile`).
- **Fixed:** network errors read as plain sentences in the GUI. Offline downloads ask Retry/Skip/Abort at once instead of after 5 minutes. A record with one unsafe path line still counts as installed (with a warning) in both the CLI and the GUI. A dash inside a component's own ID is a warning, not a rejected manifest. Not enough disk space is an alert, not a failed run. Errors name Flashpoint-relative files, never staging paths. `Open` writes nothing. Settings can't be saved while invalid.
- **Unchanged and still byte-exact:** `fpm list`, the record format and the launcher contract (golden tests).

About 8.0k lines without tests (fpm, manager, GUI and core).

**Known gaps:**
- Real-Windows test (2026-11-07, `windows-test/REPORT.md`): contract, migration through the old manager, resume, rollback and self-update pass. Found: the launcher kills the manager it opens with `/update /launcher` (job object; release blocker), a share-mode-0 lock skips the Retry dialog, OS errors worded as network errors, plus minor issues. Fixed on the Windows PC (commit `8f1acbc`) and retested on the real install (REPORT.md §6), including a second launcher issue (`ELECTRON_RUN_AS_NODE`). Next: a `windows-latest` CI job with an install / self-update / remove end-to-end run and `fpm.exe list verbose` against the golden.
- Gio needs Wine ≥ 11.13 (`EnableMouseInPointer`); Wine is only a comparison tool, not a target.
- Gio on macOS can panic when started while the display is asleep (inside Gio).
- No WinINET proxy support; staged files aren't fsynced.

## 14. M3 status (2026-11-07)

`FlashpointInstaller.exe` (`cmd/installer`) is the manager's GUI in installer mode (`internal/gui/install.go`): destination → components → progress → finish. Spec and decisions: `M3.md`.

- It opens the engine on the new folder, so the install is the normal transactional `Execute`: cancelled or failed changes nothing, downloads resume from `<dest>\Temppm`. Abandoning (Back, Quit, closing the window) removes exactly the folders the installer created, including missing parents.
- Destination rules (`internal/install`): new or empty folder; not Program Files, Windows, Temp, OneDrive or a drive root; not an existing install; under 192 characters.
- After the install: `Managerpm.cfg` (`..\` + source), the VC++ x86 runtime when missing (HTTPS from aka.ms + Authenticode check; it elevates itself), shortcuts (folder, Desktop, Start menu), then "Start Flashpoint".
- Built by Opus subagents in parallel, reviewed by three more (correctness ×2, simplicity), fixed, then tested end to end on this PC against the real stable server: a full default install (229 components, 1.6 GB, 69 s) with cancel and resume, shortcuts, launch, abandon cleanup.
- Then a Fable review (simplicity, rare cases, unsupported platforms, UI): its clear wins were applied (one page layout for all message pages, flatter shortcut code, no 32-bit-only `ProgramW6432`, no own invalid-character rule since Windows reports it, plainer installer wording) and the e2e repeated.

**Known gaps:** the UAC prompt of the VC++ runtime was not exercised (this PC has the runtime; it needs a human click). No uninstaller (the legacy installer had none). The real stable `core-manager` is still the C# manager, so a fresh install gets that until the Go manager is published.

## 15. Uninstall (2026-11-07)

Settings → "Uninstall Flashpoint…" (the C# manager had it; the Go one gained it before the switch-over PR).

- Only the Flashpoint folder the running manager is in: an old `fpm.cfg` holds an absolute path, so a copied install's manager would otherwise manage, and delete, the original. A drive root is refused (the old installer allowed `D:\`).
- Only Flashpoint's entries are deleted: the top-level names in the component records plus a fixed set (`Components`, `Manager`, `Launcher`, `Data`, `Logs`, `Temp`, `Curations`, the launcher's config files). Anything else, the user's own files in a shared folder, is kept and listed on the result page.
- The engine lock and the log file (both inside the folder) are closed first. Files in use stay; the first one is named.
- `Flashpoint.lnk` on the Desktop and in the Start menu is deleted only when it points into this install.
- The running exe is deleted after exit by a hidden `cmd` (`del`, then `rmdir` of the then-empty folders), as the C# manager did; the launcher is not started.
- Reviewed for data safety (junctions and symlinks lose only the link, `cmd` quoting with `& ^ ( ) % ü`), then tested on a real install with user files and shortcuts.

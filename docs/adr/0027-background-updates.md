---
status: accepted
date: 2026-10-21
---
# ADR 0027: Background updates for Zagros and extensions

## Context
ADR 0016 made the update check on demand only and never installed anything, because a background network call is a privacy choice. Users expect what Claude Code does: stay current without asking, say so briefly, and restart into the new version. Extensions had `market.CheckUpdates` but nothing used it, so there was no way to learn that an installed extension had a newer version.

## Decision
- `[updates] background` (default true) turns on a loop in the launcher: first run 45 seconds after start, then every 6 hours. It uses the same GitHub release request and registry download as the manual check, nothing else. `check_on_start` is ignored. Setting `background = false` stops the loop; the `update` keyword and tray item still check on demand.
- Zagros: `update.Stage` downloads the release zip for the CPU, refuses a release without `checksums.txt` or with a mismatching sum, requires the assets to be on `github.com/<repo>/releases/download/`, and unpacks only `zagros.exe` (MZ header, size capped) into `<data dir>/update/<version>`. `update.Apply` renames the running executable to `zagros.exe.old` (Windows allows renaming a running exe), writes the staged file, and restores the old one if writing fails. The new process is started with the hidden `--wait-pid <pid>` and waits for the old one to exit, so the single-instance lock is free; it removes `.old` on start. Development builds (version not semver) never stage or apply.
- Telling the user: a tray balloon and `App.Flash`, a one-off line on the empty launcher for 8 seconds that typing replaces. It is never kept on screen. Once per version per session.
- The "tool" is the existing `update` provider: its rows are "Restart to update to Zagros X" (Enter applies and restarts), "Zagros X is available" (Enter downloads, then the restart row appears), failure with retry, and the extension rows. It reaches the application through `builtin.UpdateHooks`, like `MCPHooks`, so the provider stays free of file and process code.
- Extensions are never updated unattended: a changed version may ask for new permissions and run different code. The loop only counts them and notifies; installing is a click (manager button, `update` row, or `zagros ext update`). `InstallBytes` already replaces in place and prompts for permissions that are new. Folder installs (`source = local`) are excluded from the comparison.
- No extension API and no permission: this is application behaviour, so no stub changes. Authors get updates distributed by bumping `version` in a registry-listed extension.

## Consequences
Zagros makes one GitHub API request and one registry request per 6 hours while it runs, and downloads about 10 MB when a release appears. Users who object turn it off with one line. A failed download or check is logged (Log console) and never shown as an error balloon. The executable folder must be writable; installs in protected folders keep the manual download path.

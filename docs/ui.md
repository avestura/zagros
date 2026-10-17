# UI and app binary

The UI is plain Win32 through `golang.org/x/sys/windows` and raw syscalls
(user32, gdi32, dwmapi, shell32, shcore). No CGO, no webview, no GUI library.
Code lives in `internal/ui`; wiring and the CLI live in `cmd/zagros`.

## System tray

Zagros is a tray application. It starts with no main window and no taskbar
button (the launcher window is a hidden `WS_EX_TOOLWINDOW` popup). The tray
icon is created at startup.

| Interaction | Result |
|---|---|
| Left-click | Toggle the launcher (show if hidden, hide if shown) |
| Right-click | Context menu |
| Hover | Tooltip with state: `Zagros - ready (Alt+Space)`, `Zagros - open`, or `Zagros - hotkey unavailable`, plus a second line such as `3 extension(s) enabled` |

Context menu:

- **Show**
- **Settings** opens `zagros.toml` in the default editor (falls back to Notepad)
- **Reload extensions and themes**
- **Open config folder**
- **Extensions...** opens the extension manager (greyed out when neither the extension manager nor the marketplace is available)
- **Start with Windows** (checkmark reflects the HKCU `Run` entry; toggling it also writes `start_with_windows` to the config through `Options.OnAutostart`)
- **Exit**

Behaviour details:

- **Explorer restarts.** The window registers the `TaskbarCreated` message and re-adds the icon when it arrives.
- **Notifications.** `App.Notify(title, text)` shows a tray balloon. It is thread-safe and is what `ext.Platform.Notify` (extension `notify()`) uses, and what error reporting uses (failed actions, reload results, hotkey conflicts).
- **Tooltip state.** `App.SetTrayState(text)` sets the second tooltip line; the app updates it after extension changes.
- **Single instance.** A named mutex (`Local\ZagrosLauncherSingleton`) guards startup. A second launch finds the first window by class name and posts a show message, then exits. `zagros ext install` posts a reload message to the running instance.
- **Hotkey failure.** If `RegisterHotKey` fails (the combination is taken) the tooltip changes and a balloon explains it; the tray icon still works.

## Keyboard

Launcher:

| Key | Action |
|---|---|
| `Alt+Space` (configurable, `hotkey` in the config) | Show or hide the launcher from anywhere |
| `Esc` | Hide |
| `Up` / `Down` | Move selection (wraps) |
| `PgUp` / `PgDn` | Move one page |
| `Home` / `End` | Caret to start/end of the query. With an empty query, first/last result |
| `Ctrl+Home` / `Ctrl+End` | First/last result, always |
| `Enter` | Run the selected item's action |
| `Ctrl+Enter` | Run the item's first alternate action (`Item.Alt[0]`) |
| `Tab` | Completes the query when the item offers it (for example `g ` on a web search, so you can type the query); otherwise opens the popup menu with the item's alternate actions |
| `Ctrl+1` ... `Ctrl+9` | Run the Nth visible result |
| `Left` / `Right`, with `Ctrl` | Move caret by character or word; `Shift` extends selection |
| `Backspace` / `Delete`, with `Ctrl` | Delete character or word |
| `Ctrl+A` / `Ctrl+C` / `Ctrl+X` / `Ctrl+V` | Select all, copy, cut, paste (single line; newlines become spaces) |
| Mouse | Click a row to select, double-click to run; click or drag in the input to place the caret or select; wheel scrolls |

The launcher hides when it loses focus. Actions of kind `setquery` replace the
query and keep the launcher open; every other kind hides the launcher first and
runs on a goroutine, so a slow action never blocks the UI.

Dialogs: `Esc` closes (permission prompt: counts as Cancel), `Enter` confirms
the permission prompt, wheel scrolls long lists. In the extension manager `Tab`
switches tabs and the arrow keys, `PgUp`/`PgDn`, `Home`/`End` move the selection.

## Windows

- **Launcher** (`launcher.go`, `render.go`). One window created at startup, shown with `SetWindowPos` and hidden with `SW_HIDE`. `WS_EX_TOPMOST | WS_EX_TOOLWINDOW | WS_EX_LAYERED`, borderless. Centred horizontally on the monitor that holds the cursor, 20% from the top. A fade-in of at most 100 ms (`Options.FadeMS`) uses layered alpha.
- **Permission dialog** (`prompt.go`). Implements the ext permission prompter through `App.PromptPermissions`, callable from any goroutine; it blocks until the user decides. It lists the extension name, publisher, source, and every permission with scope and a plain-English explanation, each with Allow/Deny, plus Allow all / Done / Cancel. Undecided permissions are denied. The UI stores nothing; the caller persists the returned grants.
- **Extension manager** (`manager.go`). Installed tab (enable/disable, uninstall with confirmation, click a granted permission to revoke it) Marketplace tab (list and install; each entry shows EXT or THEME, and themes go through `InstallTheme`) and Themes tab (built-in, user and installed themes; "Use" sets `theme = "<name>"` with a minimal edit of the `theme =` line in `zagros.toml`, so comments survive, then reloads). `--selftest` renders all three tabs. Network and install calls run on goroutines and post results back to the UI thread.
- Dialogs are borderless top-level windows drawn with the active theme (`panel.go`): own title bar (drag to move), owner-drawn buttons with hover.

## Rendering

GDI only. The frame is composed in a memory DC and blitted once per `WM_PAINT`
(no flicker). Text uses `CLEARTYPE_QUALITY` fonts from the theme (`font`,
`font_size`). The title shows fuzzy-match highlights (bold plus accent colour),
the subtitle is dim, the category tag sits at the right when `show_tags` is
on. Icons come from `SHGetFileInfo` on a worker goroutine (large icon scaled to the row), cached in
a 128-entry LRU, with a letter tile as the placeholder until the icon arrives and as the fallback for URLs and glyph names.

DPI: the process is Per-Monitor-V2 aware (`SetProcessDpiAwarenessContext` at
runtime, no manifest needed). Every theme metric is in 96-DPI pixels and
scaled by the DPI of the monitor the launcher opens on (`Scale`,
`ComputeMetrics`). Metrics are recomputed on each show and fonts whenever face, size or DPI change, so moving
between monitors with different scaling works.

Corners: `radius > 0` asks DWM for rounded corners (Windows 11); on older
systems a window region is used. `radius = 0` (the Boxy themes) turns rounding
off explicitly.

## Queries

Each keystroke cancels the previous query and starts
`Catalog.QueryStream(ctx, q, emit)` on a goroutine; see "Performance
architecture" below for tiers, snapshots and coalescing.

## Performance architecture

Typing must never wait for anything slower than a repaint. The rules, and
where they live:

1. **The UI thread only draws and edits.** No file, shell, network or process
   call is reachable from `WM_CHAR`/`WM_PAINT`. Actions run on goroutines
   (`activate`), tray commands that open files or write the config run on
   goroutines, queries and icon loads run on workers. (The extension manager
   window still calls `ExtAdmin` synchronously on a click; it is not on the
   typing path.)
2. **Progressive results** (`stream_ui.go`, `core/stream.go`). Each keystroke
   calls `Catalog.QueryStream`. Providers are tiered by the optional
   `core.Tiered` interface (`core/tier.go`): tier 0 (default) is in-memory
   work started immediately; tier 1 (Everything, extensions, recent files) is
   slow I/O, started after a 40 ms debounce with its own 1.5 s budget. Every
   arriving provider result is merged and re-ranked (order is by provider
   registration, so the final order does not depend on arrival order) and sent
   as a `Snapshot`. A tier-0 snapshot is always sent so stale rows of the
   previous query disappear. `Catalog.Query` is a wrapper that returns the last
   snapshot (CLI, tests).
3. **Coalescing** (`perf.go`). The query goroutine writes snapshots into a
   one-slot `latestSlot` and posts a message only when the slot was empty. The
   UI thread applies at most one snapshot per 16 ms (`rateGate`, `timerApply`),
   always the newest, and drops snapshots whose sequence number is not the
   current query. An unchanged list (`SameItems`) causes no repaint.
4. **Stable selection and window.** If the user moved the selection, it follows
   its item ID across updates (`SelectionAfterUpdate`); otherwise row 0 stays
   selected. The window height depends only on the row count (capped by
   `max_results`); `SetWindowPos` and the corner region are only touched when
   it changes.
5. **Async icons** (`iconload.go`, `perf.go`). Painting looks the icon up in an
   `iconLRU` (128 entries, UI-thread only). A miss draws the letter tile and
   queues the path on a single worker goroutine (queue of 24, deduplicated,
   newest first; requests for rows no longer visible are dropped after every
   paint). The worker runs on a COM-initialised locked thread. Network paths
   (UNC, remote drives) and per-extension types (`.txt`, ...) use
   `SHGFI_USEFILEATTRIBUTES`, so the shell never opens them; programs,
   shortcuts and folders get a real lookup, with a fallback to the generic
   icon for missing files. The `HICON` is rendered onto black and white
   backgrounds to recover alpha and stored as a premultiplied 32-bit DIB;
   painting is one `AlphaBlend`. The cache survives hide/show.
6. **GDI objects are created once.** Fonts are keyed by face, size and DPI and
   survive hide/show; fills use `ExtTextOut` with no brushes or pens; the back
   buffer is allocated once at the maximum height. After 5 minutes hidden the
   icon cache is trimmed to 32 entries, the back buffer is freed and the working
   set is trimmed. (Trimming the working set right after every hide made the
   next show fault everything back in, so it is gone.)
7. **Warm start.** After the tray and hotkey exist, `startBackground` starts the
   icon worker (which pays the one-time shell icon initialisation), warms the
   font cache and allocates fonts and back buffer from a goroutine, and
   `Catalog.Warm` loads the apps index (`core.Warmer`). Queries never block on
   an index: if the apps index is not ready the snapshot lists the provider in
   `NotReady` and the status line shows "Indexing apps...". When the index
   lands, the provider's `OnReady` callback makes the catalog fire
   `OnInvalidate` and the launcher re-runs the query, keeping the selection.
8. **Everything** (`builtin/everything*.go`, `evcache.go`) calls the SDK DLL
   in-process. Results of the last 16 searches are kept for 10 s: an identical
   search (backspace) is answered from the cache, and a plain-word search that
   extends a complete earlier one (fewer hits than the limit) is filtered
   locally on file names. Searches superseded while waiting for the DLL lock
   return at once. Only without the DLL is `es.exe` used: after a 60 ms
   debounce, and cancelling the query kills the process.

### Tracing

`zagros --trace` (or `ZAGROS_TRACE=1`) appends timing lines to
`<data dir>\trace.log`; with tracing off each event is one atomic load.
Events: `key_to_first_paint`, `key_to_items_paint` (first paint that shows
results) and `key_to_final_paint` measured from the character being inserted,
`provider` (per provider and tier, with the query), `icon_load`, `paint`,
`fonts`, `warm`, `warm_fonts`, `warm_backbuf`. `core.Trace` also keeps
per-name statistics (`Summary`).

`--selftest` includes a burst test: 50 keystrokes 20 ms apart through the real
pipeline with an instant provider and a 200 ms tier-1 provider, pumping
window messages. It fails if any frame takes more than 50 ms to paint, if
instant results take more than 30 ms to be painted after a keystroke, if the
slow provider's items are missing from the final list, or if the icon worker
loaded nothing, and prints the trace summary. Measured on the development
machine: waiting for all providers (the old behaviour) 201 ms; streamed
results painted within 2.0 ms at worst (0.7 ms average), worst paint 1.4 ms.

## Memory strategy

The target is a small idle footprint (`SetProcessWorkingSetSize` brings the
working set to well under 1 MB right after startup and after a long idle period, and
it grows only while the launcher is in use).

- One window, created once; hide and show, never destroy.
- While hidden: the result list is released; fonts, the back buffer and the icon cache are kept (trimmed after 5 minutes hidden, see "Performance architecture").
- The back buffer is allocated once at the maximum height, not per paint or per show.
- No timers run while hidden (caret blink and fade timers exist only while visible).
- A `debug.FreeOSMemory` and working-set trim runs 8 s after startup and after 5 minutes hidden, not after every hide (that made the next show page-fault).
- Providers load their index in the background after startup (`Catalog.Warm`); a query that arrives earlier shows "Indexing...".

The Go runtime itself keeps some committed address space; "private bytes" in
Task Manager is larger than the working set.

## Pure logic and tests

Layout, editing and parsing are free of Win32 and unit tested
(`internal/ui/pure_test.go`):

- `layout.go`: `Metrics` (scaling, window height, row rectangles, `HitRow`), `Scroll` (selection and scroll window), `CaretIndexAt`, `HScroll`.
- `editor.go`: `Editor` (insert, delete, word moves, selection).
- `hotkey.go`: `ParseHotkey("Ctrl+Shift+K")`.
- `match.go`: fallback highlighter and `SplitRuns`.
- `perf.go` (`perf_test.go`): `SelectionAfterUpdate`, `SameItems`, `rateGate`, `latestSlot`, `iconQueue`, `iconLRU`. Stream merge and tiering are tested in `internal/core/stream_test.go`, Everything result reuse in `internal/builtin/evcache_test.go`.

`zagros --selftest` creates the windows, renders the launcher at 96 and 168
DPI, the permission dialog and the extension manager into offscreen bitmaps,
checks pixels, runs the typing-burst latency test (see "Tracing"), adds and removes the tray icon, and exits 0. It never shows a
window.

## App binary and CLI

```
zagros                       run (tray + hotkey)
zagros --version
zagros --config              print the config path and open it
zagros --selftest
zagros ext install <path|folder|github:owner/repo[@tag]|url|registry-id>
zagros ext list
zagros ext remove <id> [--keep-data]
zagros ext pack <dir> [-o file]
zagros ext validate <dir>
zagros theme list                         built-in, user and installed themes
zagros theme install <registry-id|https url to .toml>
zagros theme remove <name>
zagros market list                        kind, id, description of every entry
zagros market search <query>
```

`zagros ext install <id>` on a theme id stops with a pointer to `zagros theme install`.
Installs and removals post a reload to a running instance.

CLI installs ask for permission consent in the terminal (`y`, `N`, or `each`).
The GUI-subsystem binary attaches to the parent console for CLI output.

Build with `scripts/build.ps1` (output `dist/zagros.exe`, flags
`-H=windowsgui -s -w`, version from `git describe`). The icon is
`assets/zagros.ico`, generated by `go run ./scripts/genicon` and embedded for
the tray and dialogs.

## Wiring

`cmd/zagros/run.go` builds core (config, history, catalog), the built-in
providers, `ext.Manager` (with `guiPrompter`, `platform`, and the executor
chain `Manager.WrapExecutor(executor)`), the market client, and passes small
interfaces to `ui.New` (`ui.Catalog`, `ui.ExtAdmin`, `ui.Market`). Extension
discovery runs from `Options.OnReady`, after the window exists, so permission
prompts can be shown. Providers are re-synchronised when the manager reports a
change.

## Typing and the result list

The input box is painted on its own path: each key is echoed straight to the
window before the query starts, and the result list is repainted separately (and
coalesced while you type fast). Typing therefore never waits on providers or icons.

## Web search fallback

When a query is not claimed by a keyword, the last row offers a web search with the
first engine in `web_search`. Press `Tab` on it to move into that engine's keyword
(`g your query`). Extensions can offer completions too: `item(..., complete="kw ")`.

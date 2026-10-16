---
status: partial
adrs: 0003, 0004
---
# WS02: Native Win32 UI and system tray

- [x] Launcher window, GDI rendering, DPI scaling
- [x] Keyboard navigation, own text editing
- [x] Global hotkey (Alt+Space default)
- [x] Tray icon, menu, Explorer restart recovery, single instance
- [x] Permission consent dialog
- [x] Extension manager window
- [x] CLI and wiring (cmd/zagros)
- [x] Typing performance (see docs/ui.md, "Performance architecture")
  - [x] `--trace` / `ZAGROS_TRACE=1` timing log (`trace.log`)
  - [x] Async icons: letter tile first, one worker, deduplicated bounded queue, premultiplied DIB blit, cache kept across hide/show (128), trimmed after 5 min idle
  - [x] Progressive results: `Catalog.QueryStream`, `core.Tiered` (tier 0 inline, tier 1 debounced 40 ms with 1.5 s budget), stale-snapshot drop, one apply per 16 ms, selection kept by item ID
  - [x] Everything: in-process DLL, result reuse for extended and backspaced queries, `es.exe` only as a debounced, kill-on-cancel fallback
  - [x] Warm start: apps index, fonts, back buffer, shell icon init off the UI thread; "Indexing..." status and automatic re-run when the index lands
  - [x] Stable window: resize only when the row count changes; GDI objects created once per face/size/DPI
  - [x] UI-thread audit: tray file/config work moved to goroutines

Verified: `go build`, `go vet`, unit tests, and `zagros.exe --selftest` (offscreen render at 96 and 168 DPI, a 50-keystroke burst with a 200 ms slow provider that fails on any paint over 50 ms or instant results later than 30 ms, permission dialog, manager, tray icon add/remove). Burst numbers on the dev machine: old wait-for-all behaviour 201 ms until anything shows; now results are painted within 2 ms of a keystroke (0.7 ms average) and the worst paint is 1.4 ms.

Not verified by hand (no interactive session): real keyboard and mouse use, Tab actions menu, IME, focus on show, fade, dialog clicks, right-click tray menu, balloons, Explorer-restart recovery, `ext install` end to end, a live typing session with `--trace` against real Everything and extensions, icons of network paths.

Known gaps: no .syso resource (exe file shows no icon in Explorer), icons use SHGetFileInfo only (no UWP icons), IME composition window not positioned at the caret, manager does not show per-extension settings and calls the extension admin synchronously on click, Ctrl+V reads the clipboard on the UI thread, `systemIsDark` reads the registry on each show when following Windows.

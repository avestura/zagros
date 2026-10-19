---
name: zagros-window-ui
description: Use when adding or changing any Zagros window other than the query box (extension manager, permission prompt, dialogs, settings or help windows in internal/ui), or any screen where users learn what an extension does or how to trigger it. Enforces a modern Boxy look instead of a flat imgui-style window, and requires every feature and extension to be discoverable from the UI.
---

# Zagros window UI

Applies to `internal/ui/manager.go`, `prompt.go`, `dialogs.go`, `panel.go` and any new window. The launcher window follows the same visual rules, but its latency rules in `docs/ui.md` come first.

Two goals, both required: the window looks like Boxy (https://github.com/avestura/boxy-design-skill, vendored in `site/assets/boxy.css`), and a user can find out what something does and how to run it without reading docs.

## 1. Look: Boxy, not a widget toolkit demo

Reference the site CSS tokens when picking values; do not invent new ones. Theme colours come from `theme.Colors`, never hard-coded, so `boxy-light`, `boxy-dark`, `classic` and `nord` all work.

- **Structure from hairlines, not boxes.** 1px borders in `Colors.Border`, no bevels, no 3D buttons, no group boxes with captions. Separate regions with a hairline or a background step (`Background` vs `InputBg`), not both.
- **Radius 0** unless the theme sets one. Never hard-code rounding.
- **4px grid.** Every gap, pad and size is a multiple of 4 at 96 DPI, passed through `p.s()`. Use the Boxy steps (4, 8, 12, 16, 24, 32, 48). No stray `p.s(10)` or `p.s(14)`.
- **Type hierarchy.** At most three levels: title (bold, `Text`), body (`Text`), meta (`TextDim`, small). Do not use size or colour changes for emphasis beyond that. Mono font only for keywords, commands and code.
- **Accent budget under 5% of pixels.** Accent marks selection, focus and the one primary button. Not tab fills, not every "Enabled" tag, not links everywhere.
- **Layout.** A master-detail split beats a list with a footer. List left (fixed width, about 40%), detail right. A full-width list of two-line rows with a strip of chips underneath is the pattern to remove.
- **Controls.** Tabs are an underlined text strip, not a row of buttons. Buttons: one primary per view, the rest ghost (text + hairline). Destructive actions are ghost until hover, then red from the semantic ramp, and always confirm. Toggles show state as a switch or a labelled state, never a button whose label flips between Enable and Disable.
- **Density and whitespace.** Rows 48-56px, 16-24px outer padding, nothing touching an edge. Align on a shared left edge per column.
- **Motion.** Mechanical: linear or short ease, 100-150ms, no bounce, no fades longer than 100ms. Scroll glide stays as it is.
- **States for every control:** rest, hover, pressed, focus (visible 2px accent ring or inset border), disabled (dim text, not removed), busy. Focus must be visible for keyboard users.
- **Title bar.** Own title bar is fine, but it matches the theme (no system-looking caption), has a close button with hover state, and drags.
- **DPI.** Everything scales through `p.s()`. Check 100%, 150% and 200%.

If you catch yourself drawing a rectangle with a label and a border around a cluster of controls, stop and use spacing and a heading instead.

## 2. Function: the window must work, not just look right

Before calling a window done:

- Keyboard-only walk-through: every control reachable (Tab / arrows), `Enter` activates, `Esc` closes or backs out, `Ctrl+F` or typing filters long lists. Show the shortcut next to the control (dim text or tooltip).
- Mouse: hover, click, double-click, wheel, scrollbar drag, drag-and-drop installs. Hit areas at least 28px tall at 100%.
- Long and odd content: 80-character names, empty strings, 0 items, 500 items, no permissions, a failed network call. Text is ellipsised with `textFit`, never overlapped or clipped mid-glyph.
- Slow work never blocks the UI thread (see `docs/ui.md` "Performance architecture"). Busy state is visible and the buttons that cannot run are disabled.
- Errors appear in the window, next to what failed, with what the user can do. Not only in a tray balloon.
- Selection, scroll position and filter survive refresh and tab switches.
- Window opens on the monitor with the cursor, remembers size, minimum size is enforced, resizing reflows the layout.

## 3. Discoverability: find out how it works and how to run it

The current manager shows name, publisher and a permission count. A user cannot tell what an extension does, what to type, or how to configure it. Every window that lists extensions, built-in providers or tools must answer these questions without leaving the window:

1. **What is it?** One-line description, then a longer description.
2. **How do I run it?** The trigger and a copyable example query: keyword (`keywords` in the manifest, shown as mono chips), the first-run phrase, and what the result looks like. A "Try it" action sets the launcher query (`setquery`) and opens the launcher. If an extension has no keyword it says "Shows up in normal search" and lists what it matches.
3. **What can it do?** Its commands, hooks and item actions, with the key for each (`Enter`, `Ctrl+Enter`, `Tab` alternates).
4. **What does it need and why?** Permissions with the reason from `[[permission_detail]]`, in plain words, with revoke.
5. **How do I configure it?** The `[[settings]]` values editable in place, with the file they are stored in printed next to them. Users must never have to guess between `zagros.toml`, the extension folder or the data dir. Settings live in `<data dir>/ext/<id>/settings.json`; the same edit is available as `zagros ext config <id> <key> <value>`.
6. **Where is the full help?** A README or docs link (manifest `homepage`/`readme`), rendered or opened.

Also:

- **Empty and first-run states teach.** An empty launcher shows a short list of things to try. A fresh install of an extension tells the user its keyword.
- **Search the manager.** Filter box at the top of every list, matching name, description, keywords and commands.
- **Marketplace entries** show the same description, keywords, permissions requested and an example before install.
- **Built-in providers and tools** appear in the same discoverable form: a "Features" or "Commands" view listing every keyword and tool with an example, generated from the registries (`core.KeywordProvider`, `internal/tools`), not typed by hand.
- **In the launcher**, `?` or `help` lists keywords and extensions; a keyword typed alone (`w `) shows the extension's usage line as a placeholder row.
- **Errors in extensions** name the extension and the fix.

Data for this comes from the manifest and registries. If `ExtInfo`/`MarketEntry` lack a field (description, keywords, examples, settings, readme path, commands), add it in `internal/ui/types.go` and the `ExtAdmin` implementation; do not invent display text in the UI.

## 4. Parity with extensions (AGENTS.md "Definition of done")

A new manifest field or host mechanism needs the full set: `internal/ext` support, stub in `stubs/zagros/`, `docs/extensions/README.md`, `docs/ui.md`, `stubs/AGENTS.md`, tests. If you add `usage`, `examples` or `commands` to the manifest so the UI can show them:

- Document them in the manifest table of `docs/extensions/README.md` and in `stubs/AGENTS.md` ("write a usage line and one example per keyword").
- `zagros ext init` templates and `examples/` use them.
- `ext validate` warns when an extension has keywords but no description or example.
- Site check: `node scripts/site-build.mjs && node scripts/site-check-links.mjs`.

## 5. Verify with evidence

Do not report a UI change done from the code alone.

1. `go build ./... && go vet ./... && go test ./...`
2. `GOOS=windows go build -o /tmp/z.exe ./cmd/zagros && /tmp/z.exe --selftest` (renders every tab offscreen and times a frame; the frame time must not regress).
3. Look at it: `z.exe --selftest-shots <dir>` saves a PNG per manager state (see `internal/ui/selftest_shots.go`; add a state there for every new view). Read the PNGs with the Read tool. Check each tab, an empty state, an error state, a long-name row, light and dark themes, and 100% and 200% DPI.
4. Compare against the Boxy site in `site-dist/` for spacing, hairlines and type. Anything that looks like a stock Windows form or imgui panel is a defect.
5. Run the section 2 checklist and the section 3 questions against a real extension from `examples/`: from the manager alone, can you tell what it does and trigger it? If not, it is not done.

Report what you measured and which screenshots you checked. Say plainly what you did not verify.

## Anti-patterns to reject

- Rows of equal-weight buttons across the top or bottom
- Labelled boxes around groups of controls
- Accent colour on tabs, tags and buttons at the same time
- A list with a status line as the only feedback
- Chips as the only place a fact is shown
- Hard-coded colours, fonts or pixel sizes
- Features reachable only through `zagros.toml` or the CLI with no hint in the UI
- Text that says what a thing is called but not what it does

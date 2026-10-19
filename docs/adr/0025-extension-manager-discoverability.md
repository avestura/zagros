---
status: accepted
date: 2026-10-20
---
# ADR 0025: The extension manager explains and configures extensions

## Context
The manager listed a name, a publisher and a permission count. Users could not tell what an extension does, what to type to run it, or how to configure it. Settings could only be changed from Go code: there was no window, no command, and no hint of where the values live (users guessed `zagros.toml`). The window itself was a flat list with a row of equal buttons.

## Decision
- The manifest gets `usage` (one line) and `examples` (queries). The manager shows them with the keywords, and each example has a **Try** button that types it into the launcher. `ext validate` warns when an extension has keywords but no examples, or neither keywords nor usage.
- `ExtInfo` carries the manifest data the pane shows (description, homepage, folder, keywords, usage, examples, declared permissions with reasons, settings with values, load error). The UI invents no text.
- Settings are edited in the manager and with `zagros ext config <id> [key [value]]`. Both go through `Extension.SetSetting`, which parses text by the setting's type and stores it in `<data dir>/ext/<id>/settings.json`. Settings stay out of `zagros.toml`, so an extension update keeps them. The manager prints the file path.
- The window is a list plus a detail pane under a tab strip with a live filter, using hairlines, a switch, ghost buttons and one primary action. The visual and discoverability rules are in `.claude/skills/zagros-window-ui`.
- `--selftest-shots <dir>` saves PNGs of the manager states so UI changes are reviewed by eye.

## Consequences
Extension authors should write `description`, `usage`, `examples` and setting labels; the guide for agents says so. A new manifest field that users need to see must be added to `ExtInfo` and the pane. Every control in the manager is reachable with Tab and pressed with Space or Enter; tab switching moved to Ctrl+Tab.

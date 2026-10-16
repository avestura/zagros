# Configuration

Zagros reads a single TOML file, `zagros.toml`, from its config directory. Run `zagros --config` to open it. Changes are picked up the next time the launcher opens.

```toml
hotkey = "Alt+Space"
theme = ""                # empty follows the Windows light/dark setting
max_results = 8
start_with_windows = true
scan_dirs = ["D:\Portable Apps"]

[providers]
clipboard = true
emoji = false             # missing providers are enabled

[aliases]
"proj" = "D:\Projects"
"gh"   = "https://github.com"

[[web_search]]
keyword = "g"
name = "Google"
url = "https://www.google.com/search?q={q}"

[everything]
enabled = true
limit = 20
# dll_path = "C:\Program Files\Everything\Everything64.dll"

[marketplace]
registry_url = "https://avestura.github.io/zagros/registry.json"
```

## Options

| Key | Default | Meaning |
|---|---|---|
| `hotkey` | `Alt+Space` | Global shortcut that toggles the launcher. |
| `theme` | empty | Theme name. Empty follows Windows light/dark. See [Themes](../../docs/themes.md). |
| `max_results` | `8` | Rows shown in the result list. |
| `providers` | all on | Table of provider name to `true` or `false`. |
| `scan_dirs` | none | Extra folders the apps provider scans for shortcuts and executables. |
| `aliases` | none | Keyword to path, URL or command. |
| `web_search` | Google and a few others | Array of tables with `keyword`, `name`, and a `url` containing `{q}`. |
| `everything` | enabled if found | Options for the [Everything](https://www.voidtools.com/) integration. |
| `marketplace.registry_url` | the project registry | Point this at your own `registry.json` to run a private marketplace. |
| `start_with_windows` | `false` | Adds Zagros to your user startup entries. |

## Everything integration

If [Everything](https://www.voidtools.com/) is running, Zagros queries it through its SDK DLL for instant file search. Zagros looks for `Everything64.dll` in the usual places; set `everything.dll_path` if yours lives somewhere else.

# Themes

A theme is a TOML file. Put it in `%APPDATA%\Zagros\themes\` (or `<ZAGROS_HOME>\config\themes\`) and set `theme = "<name>"` in `zagros.toml`. Leave `theme` empty to follow the Windows light/dark setting, which picks `boxy-light` or `boxy-dark`.

Built-in themes: `boxy-light`, `boxy-dark`, `classic` (dark, Keypirinha-like) and `nord`. A user theme whose `name` matches a built-in replaces it.

## Partial themes

Every key is optional. Missing keys come from the built-in Boxy theme that matches the file's `dark` flag, so a theme can be three lines long. If `name` is missing, the file name (without `.toml`) is used.

## Keys

Top level:

| Key | Default | Meaning |
|---|---|---|
| `name` | file name | Theme name used in `zagros.toml` |
| `author` | empty | Credit line |
| `dark` | `false` | Marks a dark theme; selects which Boxy theme fills missing keys |
| `font` | `Segoe UI` | UI font |
| `mono_font` | `Cascadia Mono` | Font for code-like text |
| `font_size` | `11` | Points |
| `width` | `640` | Window width in pixels at 96 DPI |
| `input_height` | `52` | Query box height |
| `row_height` | `44` | Result row height |
| `padding` | `12` | Inner padding |
| `border_width` | `1` | Window and row border, 0 disables |
| `radius` | `0` | Corner radius, 0 is sharp |
| `show_icons` | `true` | Show item icons |
| `show_tags` | `true` | Show the provider tag on each row |

Metrics are in pixels at 96 DPI and are scaled by monitor DPI.

`[colors]` takes `"#rrggbb"` strings (`#rgb` also works):

| Key | Used for |
|---|---|
| `background` | Window background |
| `border` | Window border and hairlines |
| `input_bg`, `input_text`, `placeholder` | Query box |
| `text`, `text_dim` | Result title and subtitle |
| `sel_bg`, `sel_text`, `sel_text_dim` | Selected row |
| `accent` | Selection marker and focus ring only |
| `tag`, `tag_bg` | Provider tag |

## Boxy

The Boxy themes use square corners, 1px hairline borders and a cool gray ramp. Color is reserved for the accent (`#0f62fe` light, `#78a9ff` dark), which marks the selected row and focus and nothing else.

## Example

```toml
name = "midnight"
author = "You"
dark = true
radius = 4
width = 720

[colors]
background   = "#10131a"
border       = "#2a3040"
input_bg     = "#161a23"
input_text   = "#e6e9ef"
placeholder  = "#66708a"
text         = "#e6e9ef"
text_dim     = "#8a93a8"
sel_bg       = "#222839"
sel_text     = "#ffffff"
sel_text_dim = "#b4bcd0"
accent       = "#7aa2f7"
tag          = "#8a93a8"
tag_bg       = "#1b2030"
```

An invalid file is skipped and the remaining themes still load.

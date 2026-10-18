# Getting started

Zagros is a keyboard launcher for Windows. Press a hotkey, type a few letters, press Enter. It is inspired by [Keypirinha](https://keypirinha.com/) and ships as one static `zagros.exe` with no installer and no runtime to install.

## Install

1. Download the latest zip for your CPU (`amd64` or `arm64`) from the [GitHub releases page](https://github.com/avestura/zagros/releases/latest).
2. Unzip it anywhere and run `zagros.exe`. A tray icon appears.
3. Press `Alt+Space` to open the launcher.

Check the `SHA256SUMS` file attached to each release if you want to verify the download.

## First queries

| Type | What happens |
|---|---|
| `fire` | Fuzzy matches installed apps; Enter launches the best match. |
| `2*(3+4)` | The calculator shows `14`; Enter copies the result. |
| `10 km in mi` | The unit converter shows `6.21371 mi`. |
| `pw` | Generates passwords you can copy. |
| `g zagros` | Web search with the `g` keyword. |
| `ext install github:owner/repo` | Installs an extension from GitHub. |

Results are ranked by fuzzy match and by how often you pick them (frecency), so the things you launch most float to the top. Zagros also remembers what you picked for a query: run Firefox after typing `fir` and it is first for `fir` from then on, and inside a Tab scope the tool you ran last leads that scope.

## Keys

- `Alt+Space` opens and closes the launcher (change it in the config).
- `Up` and `Down` move the selection, `Enter` runs the default action.
- `Tab` shows the other actions for the selected result.
- `Esc` hides the window.

## Where files live

Zagros keeps its configuration, themes and extensions in your user profile. Run `zagros --config` to open the config file in your editor. Everything is plain text (TOML for config and themes, Starlark for extensions), so it is easy to back up or put in version control.

## Next steps

- [Configuration](configuration.md) for hotkey, providers, aliases and web searches.
- [Built-in providers](../../docs/builtin-providers.md) for everything that works out of the box.
- [Extensions](../../docs/extensions/README.md) to write your own.

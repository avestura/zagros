<p align="center"><img src="assets/logo.svg" alt="Zagros logo" width="72" height="72"></p>

# Zagros

[![CI](https://github.com/avestura/zagros/actions/workflows/ci.yml/badge.svg)](https://github.com/avestura/zagros/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/avestura/zagros)](https://github.com/avestura/zagros/releases/latest)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
![Platform: Windows](https://img.shields.io/badge/platform-Windows-lightgrey)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8)

A keyboard launcher for Windows. Press a hotkey, type a few letters, press Enter.

Zagros is **inspired by [Keypirinha](https://keypirinha.com/)** by Jean-Charles Lefebvre, the fast, keyboard-driven launcher with a package system. Zagros is an independent project with its own code and is not affiliated with Keypirinha.

Website and docs: <https://avestura.github.io/zagros/>

## Why

- One static `zagros.exe`. No installer, no runtime, no CGO.
- Native Win32 window, drawn with GDI, so it opens instantly.
- Extensions are small Starlark scripts that run in a sandbox and ask for permission, like apps on Android.
- Themes are plain TOML files. The built-in themes follow the [Boxy](https://github.com/avestura/boxy-design-skill) design system.
- The marketplace is a Git repository. No server, no account.

## Features

- Fuzzy search over apps with frecency ranking
- [Everything](https://www.voidtools.com/) integration for instant file search
- Calculator and unit converter
- Password generator
- Text tools (case, encode, hash, and more)
- Clipboard history
- Window switcher
- Web search with your own keywords
- Emoji picker
- Aliases for paths, URLs and commands
- Starlark extensions with an Android-style permission system
- Themes in TOML, light and dark
- Marketplace and `github:owner/repo` installs

## Install

Download the zip for your CPU (`amd64` or `arm64`) from the [latest release](https://github.com/avestura/zagros/releases/latest), unzip it, and run `zagros.exe`. Press `Alt+Space` to open the launcher. Verify downloads against `SHA256SUMS` on the release page.

To build from source (Go 1.26 or later):

```sh
go build -ldflags "-s -w -H=windowsgui" -o zagros.exe ./cmd/zagros
```

## Quick usage

| Type | Result |
|---|---|
| `fire` | Launch Firefox (fuzzy match) |
| `2*(3+4)` | Calculator, Enter copies `14` |
| `10 km in mi` | Unit converter |
| `g zagros` | Web search with the `g` keyword |

Configuration lives in `zagros.toml`. Run `zagros --config` to open it.

```toml
hotkey = "Alt+Space"
theme = ""            # empty follows the Windows light/dark setting
max_results = 8
```

## Extensions

An extension is a folder with a `manifest.toml` and a Starlark entry script (`main.star` by default).

```toml
id = "me/hello"
name = "Hello"
version = "0.1.0"
description = "Says hello"
keywords = ["hi"]
permissions = ["clipboard.write"]
```

Permissions are declared in the manifest and approved by you at install. Scoped forms such as `network:api.github.com` limit what the extension can reach. See the [extensions guide](docs/extensions/README.md) for the script API.

Install from the registry or straight from GitHub:

```sh
zagros ext install me/hello
zagros ext install github:owner/repo
```

To publish, follow [docs/marketplace.md](docs/marketplace.md).

## Documentation

- [Getting started](site/content/getting-started.md) and [configuration](site/content/configuration.md)
- [Architecture](docs/ARCHITECTURE.md)
- [Decisions](docs/adr/) and [workstreams](docs/workstreams/)

The website is built from these files by `node scripts/site-build.mjs` (Node 18 or later, no dependencies) and deployed by GitHub Actions. Output goes to `site-dist/`.

## Credits

- [Keypirinha](https://keypirinha.com/) by Jean-Charles Lefebvre, the inspiration for Zagros.
- [Boxy](https://github.com/avestura/boxy-design-skill), the design system used for the website and built-in themes (MIT).
- [Starlark for Go](https://github.com/google/starlark-go).

## License

[MIT](LICENSE) &copy; Avestura

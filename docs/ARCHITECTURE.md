# Zagros architecture

Zagros is a keyboard launcher for Windows written in Go, inspired by
[Keypirinha](https://keypirinha.com/) by Jean-Charles Lefebvre. Single static
binary, no CGO, native Win32 UI drawn with GDI.

Module: `github.com/avestura/zagros` (Go 1.26). Dependencies stay few:
`golang.org/x/sys/windows`, `go.starlark.net`, `github.com/BurntSushi/toml`.

| Package | Role |
|---|---|
| `internal/core` | `Item`, `Provider`, `Action`; catalog (query fan-out, fuzzy rank), frecency history, remembered picks (`choices.go`), config, data paths |
| `internal/theme` | TOML themes, built-in Boxy themes, layout metrics |
| `internal/ext` | Extension package (.zagros zip): manifest, permissions, Starlark runtime, host API |
| `internal/market` | Registry index + `github:owner/repo` installs, updates |
| `internal/builtin` | Built-in providers: apps, Everything, calc, passwords, ... |
| `internal/ui` | Win32 window, hotkey, tray, rendering, settings and permission dialogs |
| `cmd/zagros` | Wiring and CLI (`zagros ext install ...`, `zagros --config`) |

Shared contract: `internal/core/types.go`. Do not change it without updating
every package that uses it.

Decisions live in `docs/adr/`, delivery tracking in `docs/workstreams/`.

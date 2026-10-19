# Working on Zagros (guide for coding agents)

Zagros is a keyboard launcher for Windows written in Go (no CGO), with a native Win32/GDI UI, Starlark extensions, a marketplace, and a Boxy-designed website. This file is for agents and contributors changing Zagros itself. (`stubs/AGENTS.md` is the different guide that `zagros ext init` writes for people building extensions.)

## Definition of done for any new API

Every time a new API is added (a host function, module, hook, permission, scope/UI mechanism that extensions can use, or a tools function), all of these ship in the same change:

1. **Extension API.** Expose it to the Starlark host (`internal/ext`). If it reaches outside the sandbox (files, network, clipboard, processes, environment, notifications) it needs a permission, declared in the manifest and approved by the user, with the check inside the host function. Interface-only features need none; say so in the docs.
2. **Stub definition.** Add or update the `.pyi` stub in `stubs/zagros/` (signature, types, docstring with the permission it needs, an example). Tools-registry functions generate their stubs: run `go run ./cmd/toolsgen` and commit the output. `TestStubsMatchHost` (`internal/ext/stubs_test.go`) and the generator test fail when the runtime and the stubs differ.
3. **Website documentation.** Update the docs the site is built from: `docs/extensions/README.md` (API reference), the permissions page (`site/content/permissions.md`) when a permission changes, `docs/ui.md` or `docs/builtin-providers.md` for launcher behaviour, and the generated `docs/extensions/tools*.md`. Then check the site: `node scripts/site-build.mjs && node scripts/site-check-links.mjs`.
4. **Agent guide.** If the change affects how an extension is written, update `stubs/AGENTS.md` too, so `zagros ext init` projects teach it.
5. **Tests.** Unit tests for the behaviour, and an extension-level test that calls it from Starlark.
6. **ADR.** A decision that changes the architecture gets a numbered file in `docs/adr/`.

The same rule applies to built-in launcher features: a host mechanism is not done until extensions can use it.

The reverse needs no new work for async results: built-in providers already refresh the list through `core.Tiered`, `Warmer` and `ReadyNotifier`, and extensions get the same effect from `task()` (ADR 0023). A new built-in way to show late results must be reachable from `task` or `item(pending=True)` or say why not.

## Commands

```
go build ./... && go vet ./... && go test ./...        # everything
GOOS=windows go build -o /tmp/z.exe ./cmd/zagros && /tmp/z.exe --selftest   # offscreen UI checks
go run ./cmd/toolsgen                                  # regenerate stubs + tools docs
node scripts/site-build.mjs                            # website into site-dist/
pwsh scripts/build.ps1                                 # dist/zagros.exe
zagros --bench-input                                   # real typing latency (needs a console build)
```

## Layout

- `cmd/zagros` the executable (launcher, `ext`, `theme`, `market`, `update`, `mcp` commands).
- `internal/core` provider/item/scope contracts, catalog, config, history.
- `internal/ui` Win32/GDI window, tray, extension manager (Windows only).
- `internal/builtin` built-in providers. `internal/tools` the pure function registry (one package per group under it) that feeds the launcher, the Starlark modules, the stubs/docs and the MCP search.
- `internal/ext` Starlark runtime, permissions, host API. `internal/market` registry client and installer.
- `stubs/` the `.pyi` API stubs and the extension-author guide. `docs/` Markdown that becomes the website. `site/`, `scripts/` the site build.

## Conventions

- Keep comments sparse and in the surrounding style. No CGO.
- Files may use CRLF on Windows checkouts; edit with tools that preserve line endings.
- Pure tools functions must not read files, the network, the clipboard or the environment; those belong in permissioned host modules.
- Never report a performance or UI fix as done without measuring it (see `--bench-input`, `--selftest`).
- Changing a window other than the query box (extension manager, dialogs) or any screen that explains extensions: follow `.claude/skills/zagros-window-ui/SKILL.md` (Boxy look, working controls, discoverable usage).

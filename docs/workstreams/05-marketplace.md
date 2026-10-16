---
status: partial
adrs: 0009
---
# WS05: Marketplace and registry

- [x] Registry format, schema and generator
- [x] Client: list, search, resolve, download with limits
- [x] github:owner/repo installs and update checks

Open items:
- Not run against real GitHub (tests use httptest servers).
- `tomlite.go` is a small TOML subset parser and `ParseManifest` assumes a manifest layout; reconcile with `internal/ext` and switch to BurntSushi/toml.
- The app adapter from `ext.Manager` to `market.Installer` is not written yet.
- [x] Starter content: 6+ themes and 4+ sample extensions in `marketplace/` with registry entries (ADR 0012)

Starter content (ADR 0012): 7 themes and 6 extensions (wiki, hackernews, npm-search, weather, dictionary, snippets), all with registry entries. Tests validate every theme's contrast and run every extension against stubbed HTTP.

Still open:
- Theme install, `market list|search` and the Themes tab are wired (`zagros theme install`), but only smoke-tested offline.
- HTTP fixtures are written from memory of each service's response shape; no live calls were made.
- Theme release-asset route has no test.

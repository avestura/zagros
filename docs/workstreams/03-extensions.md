---
status: shipped
adrs: 0005, 0006
---
# WS03: Starlark extension runtime and permissions

- [x] Manifest and zip install with zip-slip protection
- [x] Permission store and per-call checks
- [x] Host modules (http, fs, clipboard, proc, storage, crypto)
- [x] Provider adapter, hot reload, pack and validate
- [x] Extension docs and two examples

Notes:
- Tests pass for permission denial, zip-slip, step limit, timeout, install and consent flows. `-race` was not run (needs CGO).
- The UI must supply a `Platform` (clipboard, notifications) and a `PermissionPrompter`.
- Manifest uses `[[permission_detail]]` for reasons and `[[settings]]` for the settings schema.

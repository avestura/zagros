---
status: accepted
date: 2026-10-18
---
# ADR 0019: Wider file, network, archive and database APIs inside the same sandbox

## Context
Extensions could read and write one text file at a time and make GET and POST requests. Real extensions need binary files, folder operations, downloads, archives, a small database and a way to handle errors, and each of those can reach outside the sandbox.

## Decision
- No new permission kinds. `fs.read` and `fs.write` stay folder-scoped and `network` stays host-scoped. Every new function runs the same check on the resolved path or host, inside the host function. Operations on the entry itself (`fs.remove`, `fs.rename`) resolve the parent folder but not the last element, so a link is removed rather than its target. Granted roots, the home folder and drive roots cannot be removed or replaced.
- Trees are never followed through links: `walk`, `glob`, `copy` and archive creation skip symlinks, and archive extraction rejects absolute names, `..`, drive letters and anything that resolves outside the destination. Links and special files in archives are skipped. Totals are capped (`Limits.MaxBulkBytes` 256 MiB, `MaxEntries` 10000) and enforced on the bytes actually written, not on sizes the archive claims.
- `http.request` and `http.download` share one round-trip path with `get` and `post`, so host scoping, per-hop redirect checks and size caps are identical. Cookies live in an in-memory jar per extension, off unless asked for.
- `sqlite` uses `modernc.org/sqlite` (pure Go, no CGO). It needs only `storage`, which is always granted, because the extension cannot name a path: the file is `<DataDir>/ext/<id>/db/<name>.sqlite3` with a name restricted to `[A-Za-z0-9_-]`. The connection has `SQLITE_LIMIT_ATTACHED` set to 0, a `max_page_count` that caps the file at `Limits.MaxDBBytes`, and a statement guard that refuses `ATTACH`, `VACUUM INTO`, `load_extension` and every PRAGMA outside a short read-style list. Rows are capped per query.
- Starlark has no `try`, and `try` is reserved, so the global `attempt(fn, *args, **kwargs)` returns `(value, None)` or `(None, message)`. Cancellation (timeout, step limit) is never caught. Host functions keep raising.
- `binary` (pack/unpack, hex, integers) and the bytes half of `compress` are host modules rather than tools-registry groups, because the registry has no bytes parameter type. They are pure and need no permission. The module is called `binary` because `bytes` is a Starlark builtin that a module would shadow.
- zstd is not provided: the standard library has no implementation and a third-party one is a large dependency for little use.

## Consequences
The Windows binary grows by the SQLite engine, about 3.9 MB (17.4 MB to 22.2 MB with the other new code, 18.3 MB to 22.2 MB for SQLite alone). Extensions get most file and web tasks without a permission change, and a granted folder stays the only place they can touch. The statement guard is lexical, so a new SQLite feature that reads a path would need a new rule; the engine-level ATTACH limit and the fixed file location are the real boundary.

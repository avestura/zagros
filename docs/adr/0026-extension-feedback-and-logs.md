---
status: accepted
date: 2026-10-21
---
# ADR 0026: Loading, failure and log feedback for extensions

## Context
An extension that blocked in `on_query` on a slow or failing request left the launcher blank: no row, no status, and the only trace was a line in `zagros.log`. A folder that failed to load, such as a `min_zagros` check against a `dev` build, was not in the manager at all.

## Decision
- The catalog tells the UI which tier-1 providers it is still waiting for after 300 ms (`Snapshot.Slow`) and, on the last snapshot, which ones gave nothing and why (`Snapshot.Problems`: budget exceeded, or the reason from the optional `core.Failer`). No extension code is needed.
- `ext.LogBuffer` keeps 100 entries per extension in memory by wrapping `Options.Log`. The manager shows them, `zagros ext logs` reads `zagros.log`.
- Folders that fail to load are listed as `broken:<folder>` and can be removed.
- Builds that are not a semver release never fail `min_zagros`.
- `task()` and `item(pending = True)` stay the way to show a loading row. No new extension API and no permission.

## Consequences
Every query to a slow extension may produce one more snapshot at 300 ms. The status line is plain text; it does not say which request is slow, only which extension.

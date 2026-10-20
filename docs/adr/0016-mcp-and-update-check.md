---
status: accepted
date: 2026-10-17
---
# ADR 0016: MCP API search for agents, and an opt-in update check

## Context
People will build extensions with coding agents, which need to search the real API. Separately, users want to know when a new Zagros release exists.

## Decision
MCP: `internal/mcp` implements the Model Context Protocol (JSON-RPC over stdio, or POST /mcp over HTTP) with tools to read the guide, list modules, search, and fetch an item with its permission. It indexes the embedded `.pyi` stubs, so answers equal the runtime API (kept in sync by a test) and no user data is involved. `zagros mcp` serves stdio; the launcher can also serve HTTP (`[mcp] http` or the tray item), loopback only, with Host/Origin checks and an optional bearer token. The tray, the `mcp` keyword and `zagros mcp prompt` hand the user a prompt that teaches an agent how to traverse the API.

Update check: `internal/update` reads the latest GitHub release (public metadata only), compares semantic versions (tolerating `git describe` builds), caches the result, and picks the Windows asset for the CPU. It never downloads or installs. (Superseded in part by ADR 0027: a background check and installs now exist, on by default.) It runs only on demand (`update` keyword, tray item, `zagros update`) unless `[updates] check_on_start` is set, because a background network call is a privacy choice for the user.

## Consequences
Agents get an accurate, permission-aware API map without Zagros reading anything private. Update checks cost nothing unless asked for. Extensions get `host.version()` and `host.compare_versions()`; looking for releases themselves uses the existing `http` module and its network permission.

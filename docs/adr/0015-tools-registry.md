---
status: accepted
date: 2026-10-17
---
# ADR 0015: One registry of pure tools feeds the launcher, extensions, stubs and MCP

## Context
The owner wants a large set of everyday tools (strings, numbers, encodings, hashes, ciphers, random values, IDs, regex, networking, dates, extractors, clipboard changes) in the launcher, available to extensions, and discoverable by coding agents. Writing each tool four times (UI, Starlark binding, stub, docs) would drift.

## Decision
`internal/tools` holds a registry of pure functions. Each registration carries its group, name, title, documentation, typed parameters with defaults, return type, example, search tags and launcher metadata (family, order, input parameter, an input check for decoders, keywords). Everything else is derived: the launcher provider (`internal/builtin/tools.go`) builds families of variants from it; `internal/ext` turns each group into a Starlark module (text, lines, num, math, encode, hash, cipher, rand, uuid, regex, net, dt, extract) with no permission, since the functions touch no files, network, clipboard or environment; `cmd/toolsgen` generates the `.pyi` stubs and the docs pages; the MCP index reads the stubs. Each group is its own Go package registering in `init`, linked together by `internal/tools/all`, so groups are developed and tested independently.

Launcher behaviour: a family keyword lists the likely variants for the typed text; Tab enters the family scope with every variant; `name=value` sets parameters; `a :: b` passes several texts; empty input uses the clipboard; slow functions (bcrypt, scrypt, Argon2, PBKDF2) only run in the scope. The `clip` family applies text-like tools to the clipboard. Tool rows never enter history.

## Consequences
A new tool is one registration and appears everywhere; a test fails if generated files are stale or a function lacks documentation. Values that can exceed 64 bits cross the API as strings. The old `text` provider hands its overlapping keywords to the tools provider.

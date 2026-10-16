---
status: accepted
date: 2026-10-16
---
# ADR 0005: Extensions are Starlark packages in zip archives

## Context
The owner wants extensible plugins that are safe to install from strangers. Keypirinha uses Python in-process, which cannot be sandboxed.

## Decision
An extension is a .zagros zip with manifest.toml plus Starlark files. The Go runtime provides host modules (http, fs, clipboard, proc, storage, crypto, and so on). Starlark has no ambient I/O, so every capability is a host function we control. Execution steps and call time are limited.

## Consequences
Safe by construction and hot-reloadable. Extensions cannot use native libraries. The API surface is ours to maintain and document.

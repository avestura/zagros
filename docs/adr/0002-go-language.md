---
status: accepted
date: 2026-10-16
---
# ADR 0002: Implement Zagros in Go

## Context
The first plan was Rust. The toolchain on the dev machine was Rust 1.86, too old for current Starlark and allocative crates, and the owner then asked for Go. Go 1.26 is installed.

## Decision
Write Zagros in Go with no CGO. Dependencies stay few: golang.org/x/sys/windows, go.starlark.net (the reference Starlark implementation), BurntSushi/toml. Release builds use -ldflags "-s -w -H=windowsgui".

## Consequences
A single static exe and easy cross-compilation to arm64. The Go runtime costs more idle memory than Rust, so the UI frees caches when hidden and avoids large in-memory indexes. Supersedes the initial Rust plan, which was never committed.

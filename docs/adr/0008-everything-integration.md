---
status: accepted
date: 2026-10-16
---
# ADR 0008: Everything search through its DLL, with es.exe fallback

## Context
The owner asked for built-in Everything integration.

## Decision
Load Everything64.dll lazily with syscall (no CGO) from a configured path, beside the exe, or the Everything install. Fall back to es.exe. If neither works or the service is not running the provider returns nothing and reports a status string.

## Consequences
No hard dependency and no bundled third-party binary. Users need Everything installed and running.

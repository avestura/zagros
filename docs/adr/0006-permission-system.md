---
status: accepted
date: 2026-10-16
---
# ADR 0006: Android-style permissions declared in the manifest

## Context
Sensitive tasks (network, file write, process spawn) must be explicit and user-approved.

## Decision
The manifest declares permissions, optionally scoped (network:api.github.com, fs.read:~/Documents). On install and whenever the set changes the UI shows a consent dialog with per-permission allow/deny. Grants persist in permissions.json. Each host function checks its grant at call time. storage is granted by default.

## Consequences
Users see what an extension can do before it runs. Scoping requires care in each host function. Revocation takes effect on the next call.

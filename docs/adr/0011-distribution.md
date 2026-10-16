---
status: accepted
date: 2026-10-16
---
# ADR 0011: Portable zip first, then winget and Scoop

## Context
Single exe, no installer needed. Package managers help updates.

## Decision
Releases ship zip files for windows amd64 and arm64 with checksums. A portable marker file keeps data beside the exe. Winget and Scoop manifests are kept in packaging/.

## Consequences
No installer to maintain. Package manifests need version bumps at each release, scripted in the release workflow where possible.

---
status: accepted
date: 2026-10-16
---
# ADR 0004: Run resident in the system tray

## Context
A launcher must be available instantly and must not clutter the taskbar.

## Decision
The process starts with a tray icon, no taskbar button and a hidden launcher window. Left click toggles the launcher, the right-click menu offers settings, reload, extensions, start with Windows and exit. The icon is recreated when Explorer restarts. A named mutex keeps one instance and a second launch shows the first.

## Consequences
Instant show with no cold start. Needs a Start with Windows toggle (HKCU Run key).

---
status: accepted
date: 2026-10-16
---
# ADR 0003: Native Win32 UI drawn with GDI

## Context
Requirements: native UI, fast, memory efficient, themeable. Options were raw Win32, Slint or egui (in the Rust plan), or a Go GUI toolkit.

## Decision
Use raw Win32 through x/sys/windows syscalls and draw with GDI double-buffering. One borderless window is created once and hidden/shown. Theme metrics scale with per-monitor DPI.

## Consequences
Smallest footprint and full control over Boxy-style sharp visuals. More code for text editing, hit testing and dialogs. Layout math is kept pure so it can be unit tested without a display.

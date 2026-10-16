---
status: accepted
date: 2026-10-16
---
# ADR 0007: Themes and config are TOML

## Context
Keypirinha-style extensible UI and themes; users hand-edit files.

## Decision
Config is zagros.toml. Themes are TOML files in the themes folder, falling back to built-ins (boxy-light, boxy-dark, classic, nord). Colors are #rrggbb strings. The default theme follows the Windows light/dark setting.

## Consequences
Easy to share themes as one file. Layout changes beyond metrics and colors need new theme fields, not arbitrary code.

---
status: accepted
date: 2026-10-16
---
# ADR 0010: Project site uses the Boxy design system, built by GitHub Actions

## Context
The owner wants a GitHub Pages site with docs in the Boxy style.

## Decision
A small dependency-free Node script renders docs/*.md plus hand-written landing and marketplace pages into site-dist, using the vendored Boxy CSS and the boxy-design linter. A workflow deploys to Pages on every push to main.

## Consequences
Docs live next to code and ship with it. We keep Boxy's licence notices.

---
status: accepted
date: 2026-10-17
---
# ADR 0014: Host features reach extensions, and the API ships as stubs

## Context
Scopes were first a host-only mechanism. The owner wants extensions to be first-class users of every host feature, and wants people to build extensions with coding agents that can search the API.

## Decision
A host mechanism is not done until extensions can use it. Every new host feature ships with: the Starlark API and hook, a permission if it reaches outside the sandbox (UI-only features need none, and the docs say so), documentation and website pages, and updated `.pyi` stubs. The stubs live in `stubs/zagros/` (one file per module), are embedded in the binary, and `zagros ext init` / `zagros ext stubs` write them with an `AGENTS.md` guide into a project. The site publishes them under `stubs/`. A Go test compares the runtime's predeclared names with the stubs in both directions, so they cannot drift.

## Consequences
Adding a host function without a stub fails CI. Dev files (`zagros-stubs/`, `AGENTS.md`, `CLAUDE.md`) are excluded from packed archives. The stubs describe a Python dialect for tooling only; extensions still run as Starlark.

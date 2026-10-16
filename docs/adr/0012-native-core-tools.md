---
status: accepted
date: 2026-10-16
---
# ADR 0012: Core tools are native Go, the marketplace holds the extras

## Context
The owner wants fast, necessary functionality (calculator, dev tools, password generation) to be quick and always available, and the marketplace to start with real content: at least five themes and a few plugins.

## Decision
Calculator, unit converter, password and passphrase generation, hashing/encoding/UUID/timestamp tools, system commands, clipboard history, emoji, window switcher and Everything search are built-in Go providers. They do not use the plugin system. Starter marketplace content is separate: six or more themes (installed as a single TOML file into the themes folder) and several Starlark extensions (Wikipedia, Hacker News, npm, weather, dictionary, snippets) that live in marketplace/ in this repo and are listed in the registry with a subdir reference.

## Consequences
Core tools start instantly and work with no permissions or network. Registry entries get a `kind` (theme or extension). The marketplace doubles as worked examples for extension authors.

---
status: accepted
date: 2026-10-18
---
# ADR 0018: Make the extension API findable by agents without embeddings

## Context
Coding agents build extensions from the `.pyi` stubs. With a few hundred functions, an agent that does not know the word `compress` never finds `compress.extract` by typing "unzip". Stubs also looked like Python, so agents sometimes wrote `import encode`, and examples in docstrings could drift into invalid code.

## Decision
- Docstrings carry two search lines: `Keywords:` (synonyms and task phrases, comma separated) and `Task:` (one headline task, which lists the function in a cheat sheet). `Example:` is runnable Starlark, one line or an indented or fenced block. `Needs:` stays the permission.
- `internal/mcp` parses them. `search_api` ranks name above Keywords and Task, Keywords above the doc body, and expands query words with a small built-in synonym table plus light stemming. Lexical only, no embeddings, no network.
- The tools registry gets `Terms` (and `Task`); the generated `Keywords:` line is `Terms` followed by the existing `Tags`, so a function with only Tags still has one. Terms for the older groups are listed in `internal/tools/all/keywords.go`; a new group sets the fields in its own registrations.
- Every stub says in its header that it is a type description, not runnable Starlark.
- `stubs/AGENTS.md` holds a module table and a task cheat sheet between `<!-- BEGIN GENERATED -->` markers, rewritten by `go run ./cmd/toolsgen` from all stubs. `zagros ext init` and `zagros ext stubs` also write `API_INDEX.md`, one grep-able line per function.
- Tests fail when the generated sections are stale, when a function lacks Keywords, an Example or Needs, when a stub header lacks the note, and when an Example does not parse with the interpreter extensions use.

## Consequences
A module added by anyone appears in the table, the cheat sheet, the index and the search after one `toolsgen` run, and CI says so when it was forgotten. Authors write a Keywords line per function. The synonym table is hand-kept and small; groups must stay tight (encode and encrypt are not synonyms).

# Build extensions with coding agents

Zagros ships its extension API as Python type stubs (`.pyi`) and a guide written for coding agents such as Claude Code. An agent can search the stubs instead of guessing, write the extension, validate it, and pack it.

## Start a project

```
zagros ext init my-tool
cd my-tool
zagros ext validate .
zagros --dev
```

`zagros ext init` creates a working extension and everything an agent needs:

| File | Purpose |
|---|---|
| `manifest.toml` | id, name, version, keywords, permissions |
| `main.star` | a small working extension that already uses a scope |
| `AGENTS.md` | the guide for coding agents (workflow, Starlark rules, permissions, UI mechanisms, a module table and a task cheat sheet) |
| `API_INDEX.md` | one line per function: `module.name(args) -> ret \| Needs: permission \| summary \| Keywords: ...`, to grep |
| `CLAUDE.md` | points Claude Code at `AGENTS.md` |
| `zagros-stubs/zagros/*.pyi` | the API stubs, one file per module |

When it works, install it without the marketplace: `zagros ext install .` (or the `install_extension` tool of the [MCP server](mcp.md), which lets the agent do it while you approve the permissions). The extension manager does the same from **File...**, **Folder...** or by dropping a `.zagros` file or folder on its window. A folder install is recorded with the source `local`; install again to update it.

None of the dev files end up in the packed `.zagros` archive. Run `zagros ext stubs .` after upgrading Zagros to refresh `zagros-stubs/`, `AGENTS.md` and `API_INDEX.md`.

## The stubs

Starlark is a Python dialect and nothing is imported: every name in the stubs is already defined when `main.star` runs. Each stub says so in its header: it is a type description, not runnable Starlark. The stubs exist to be read and searched.

| Stub | Describes |
|---|---|
| [`__init__.pyi`](../../stubs/zagros/__init__.pyi) | `item`, `scope`, `action_open`, `action_copy`, `action_run`, `action_call`, `notify`, `settings`, `ctx` |
| [`hooks.pyi`](../../stubs/zagros/hooks.pyi) | `on_load`, `on_query`, `on_scope`, `on_execute`, `on_settings_changed` |
| [`permissions.pyi`](../../stubs/zagros/permissions.pyi) | every permission, its scope and what it allows |
| one file per module | the host modules (`http`, `fs`, `clipboard`, ...) and the pure tool libraries (`text`, `encode`, `hash`, ...). The full list is the generated module table in `AGENTS.md`. |

Every function documents the permission it needs ("Needs: network"), has a runnable `Example:`, and a `Keywords:` line with synonyms and task phrases. The same files are published on this site under `stubs/zagros/` so an agent can fetch them without installing Zagros.

## How an agent finds a function

1. **Grep `API_INDEX.md`.** One line per function, with its Keywords, so `grep -i "unzip\|decompress" API_INDEX.md` lands on the right entry even when the function is called `extract`.
2. **Ask the MCP server.** `search_api` takes a task ("turn text into base64") and understands synonyms; see [MCP](mcp.md).
3. **Use the cheat sheet** in `AGENTS.md`, which maps common tasks to functions.
4. **Open the stub** for the full docstring.

`AGENTS.md` and `API_INDEX.md` are generated from the stubs, so they list every module, including ones added after you created the project. `zagros ext stubs .` refreshes them.

## Docstring conventions

A function docstring ends with these lines, each starting at the left edge of the docstring:

```python
def base64_encode(text: str) -> str:
    """Encodes text as Base64.

    Example: encode.base64_encode("Hello")
    Task: turn text into base64
    Keywords: base64, b64, text to base64, encode as base64
    Needs: none
    """
```

| Line | Meaning |
|---|---|
| `Example:` | Starlark that parses with the real interpreter. Either one line after the colon, or the colon alone followed by an indented or fenced block. A test parses every example. |
| `Task:` | Optional. One headline task phrase; the function is listed under it in the cheat sheet. |
| `Keywords:` | Comma separated synonyms and task phrases ("unzip, decompress, extract zip"). Required on every function. |
| `Needs:` | The permission, or `none`. |

Tool-library functions get these lines from the registry (`Terms`, `Task`, `Tags`); run `go run ./cmd/toolsgen` after changing them.

## Tell your agent

Open the folder in your agent and ask for what you want, for example:

> Read AGENTS.md and the stubs in zagros-stubs/. Write a Zagros extension that looks up npm packages. Use a scope so Tab lists the versions. Declare the narrowest permissions, then run `zagros ext validate .`.

## Host features and the API

Every host feature has a matching extension API, permissions where it reaches outside the sandbox, documentation, and a stub. A test fails when a function exists in the runtime but not in the stubs, or the other way round, so the stubs cannot drift. Features that only change the interface, like scopes and Tab completion, need no permission.

# MCP server for coding agents

Zagros can serve its extension API to coding agents (Claude Code, Cursor, and any other client that speaks the [Model Context Protocol](https://modelcontextprotocol.io)). The agent searches and browses the real API instead of guessing, then writes the extension.

The server only serves the public API documentation: the same `.pyi` stubs and guide that `zagros ext init` writes. It never reads your files, clipboard or settings.

## Give your agent the prompt

Copy the prompt from the tray menu (**Copy coding-agent prompt**), from the launcher (type `mcp`, press Enter), or from the command line:

```
zagros mcp prompt
```

Paste it into your agent and finish the last line with what you want to build. The prompt explains how to use the server: read the guide, list the modules, search, then fetch the exact signature and permission before using a function.

## Connect

**stdio** (the agent starts Zagros itself; nothing to enable):

```
claude mcp add zagros -- "C:\path\to\zagros.exe" mcp
```

`zagros mcp config` prints this command with your real path, plus the JSON for other clients.

**HTTP** (Zagros is already running):

- Turn it on from the tray (**MCP server for coding agents (HTTP)**) for this session, or set it in `zagros.toml` to start with Zagros:

```toml
[mcp]
http = true
addr = "127.0.0.1:7777"
# token = "choose-a-secret"   # clients must then send Authorization: Bearer <token>
```

- Endpoint: `http://127.0.0.1:7777/mcp`. With Claude Code: `claude mcp add --transport http zagros http://127.0.0.1:7777/mcp`.
- `zagros mcp --http` runs the same server in the foreground without the launcher.

The HTTP server only listens on a loopback address, refuses requests whose `Host` or `Origin` is not local (a defence against DNS rebinding), and checks the bearer token when one is set.

## What the agent can call

| Tool | Purpose |
|---|---|
| `get_guide` | The extension author's guide: workflow, Starlark rules, permissions, UI mechanisms. |
| `list_modules` | Every API area with a one-line description and its function count. |
| `list_module` | All functions of one module (or `global` for `item`, `scope`, actions, hooks) with signatures. |
| `search_api` | Ranked search by keyword or task, for example "base64 decode" or "read clipboard". |
| `get_api` | One item in full: signature, parameters, the permission it needs, documentation, example. |
| `get_permissions` | Every permission and what it allows. |
| `get_hooks` | The hook functions an extension defines. |

The stubs are also exposed as resources (`zagros://stubs/<module>.pyi`, `zagros://guide`), and the prompt `write_extension` carries the workflow with your goal.

## Keeping it current

The index is built from the stubs embedded in the binary, and a test fails when the runtime and the stubs differ. A new API is therefore searchable as soon as its stub exists.

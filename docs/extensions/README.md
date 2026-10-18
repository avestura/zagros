# Zagros extensions

Extensions add search results and actions to Zagros. They are written in
[Starlark](https://github.com/bazelbuild/starlark) (a small, deterministic
Python dialect), run in a sandbox, and can only touch the system through
permissions the user grants.

## Format

A `.zagros` file is a zip archive:

```
manifest.toml      required, at the archive root
main.star          entry script (name set by `entry`)
lib.star           other scripts, loaded with load("lib.star", "name")
assets/...         optional files, always readable by the extension
```

An archive whose files all sit inside one top-level folder is also accepted
(for example a GitHub source zip). Installed extensions are unpacked to
`<ExtDir>/<id>/` (an id of the form `owner/name` becomes `owner__name`).

Build one with `zagros ext pack <dir>`, check it with `zagros ext validate <dir>`,
install it with `zagros ext install <file.zagros>`.

Limits: archive 16 MiB, unpacked 64 MiB, 2000 files, 16 MiB per file. Entries
with `..`, absolute paths, drive letters or symlinks are rejected.

## manifest.toml

```toml
id = "dev.example.weather"       # reverse-dns or owner/name; lowercase [a-z0-9._-]
name = "Weather"
version = "1.2.0"                # semver
description = "Forecasts in the launcher."
author = "Jane Doe"
homepage = "https://example.com"
license = "MIT"
min_zagros = "0.1.0"             # optional; install fails on older builds
entry = "main.star"              # default
keywords = ["w", "weather"]      # optional query keywords

permissions = ["network:api.open-meteo.com", "storage"]

[[permission_detail]]            # reason shown in the consent dialog
permission = "network:api.open-meteo.com"
reason = "Fetches the forecast."

[[settings]]                     # user-configurable values
key = "units"
type = "choice"                  # string | bool | int | number | choice
options = ["metric", "imperial"]
default = "metric"
label = "Units"
description = "Temperature units"
```

| Field | Required | Notes |
|---|---|---|
| `id`, `name`, `version` | yes | |
| `description`, `author`, `homepage`, `license` | no | `validate` warns when description/author are empty |
| `min_zagros` | no | semver |
| `entry` | no | relative `.star` path, default `main.star` |
| `keywords` | no | if set, the extension is only queried when the query starts with one of them, and the keyword is removed before `on_query` |
| `permissions` | no | see below |
| `[[permission_detail]]` | no | `permission` + `reason`; `permissions` is a plain array, so details live in their own table array |
| `[[settings]]` | no | `key` (`[a-z_][a-z0-9_]*`), `type`, `default`, `label`, `description`, `options` (choice) |

## Permissions

Declare what you need in `permissions`. The user is asked on first install and
again whenever the declared set changes (a version update with the same set
keeps the earlier decision). Decisions are stored in `permissions.json` in the
data dir, keyed by extension id. A host function called without its permission
raises a Starlark error: `permission denied: <id> requires "<kind>" ...`.
Actions returned by items (open, copy, run) are checked too; a denied action
becomes a no-op and a warning is logged.

| Permission | Allows | Scope |
|---|---|---|
| `network` | `http.get/post/json` | `network:host` allows that host and its subdomains. Redirects are checked per hop. Unscoped means any host |
| `fs.read` | `fs.read_text`, `fs.list`, `fs.exists` | `fs.read:~/Documents` (path prefix, `~` expands, symlinks resolved). The extension's own folder is always readable |
| `fs.write` | `fs.write_text` | same as `fs.read` |
| `clipboard.read` | `clipboard.get` | none |
| `clipboard.write` | `clipboard.set`, `action_copy` | none |
| `process.run` | `proc.run`, `action_run` | `process.run:git` (executable name, `.exe` optional) or a full path |
| `env.read` | `env.get` | `env.read:USERNAME` (one variable) |
| `notifications` | `notify` | none |
| `open` | `action_open` (URLs and files) | none |
| `storage` | `storage.*` | none. Always granted, you need not declare it |

A user may grant fewer permissions than requested, never more than declared.

## Runtime model

- `main.star` is executed once when the extension starts, then its globals are
  frozen. Mutable state belongs in `storage`, not in globals.
- Each hook call runs in a fresh thread. Defaults: 20 million Starlark steps
  and 5 seconds wall clock per call; exceeding either aborts the call with an
  error (logged, the launcher is unaffected). Cancelling a query cancels the
  script.
- Language: top-level `if/for/while`, recursion, `set` and global reassignment
  are enabled.
- `load("other.star", "name")` loads files inside the extension folder.
- Storage quota is 1 MiB per extension; HTTP responses are capped at 8 MiB and
  file reads at 4 MiB.

## Hooks

Define these at the top level of the entry script. Each receives `ctx` first.

| Hook | Called | Returns |
|---|---|---|
| `on_load(ctx)` | once, after the module loads | ignored |
| `on_query(ctx, query)` | on each keystroke while the extension is in scope | list of `item(...)` |
| `on_execute(ctx, item)` | when an item with `action_call` (or no action) is confirmed | `None`, or an action (`action_open`, `action_copy`, `action_run`) to run next |
| `on_scope(ctx, scope_id, query, data)` | on each keystroke while the user is inside a scope your item offered (see below) | list of `item(...)` |
| `on_settings_changed(ctx)` | after the user saves settings | ignored |

`ctx` fields: `id`, `name`, `version`, `dir` (absolute extension folder).

In `on_execute`, `item` has `id`, `title`, `name` (the first argument of
`action_call`, `"default"` when the item had no action) and `data` (the second
argument, round-tripped through JSON).

## API reference

### Items and actions

`item(id, title, subtitle="", icon="", action=None, alt=[], score=0.0, complete="", scope=None)`
Build a result. `id` and `title` are required; `id` must be stable for the same
result (it feeds the usage history). `icon` is a file path (the shell icon is
used) or a glyph name. Without `score`, results keep the order you return
them. Without `action`, confirming calls `on_execute` with name `"default"`.
`alt` is a list of `{"name": ..., "action": ...}` dicts (or `(name, action)`
tuples) shown in the actions menu.

`complete` is the query `Tab` puts in the input box instead of opening
anything (for example `"wiki "`). `scope` makes `Tab` enter a context (below).

### Scopes: Tab contexts

`scope(id, label, placeholder="", query="", data=None)` builds a context that
`Tab` enters when your item is selected. The launcher shows a chip with `label`
in the input box, seeds the input with `query`, and replaces the list with the
items your `on_scope(ctx, scope_id, query, data)` hook returns on each
keystroke (`data` is round-tripped through JSON). `Esc`, or `Backspace` on an
empty input, leaves the scope and restores what was typed before. Items inside
a scope behave like any other item. Items without a `scope` that have `alt`
actions open the built-in "Actions" scope on `Tab` instead.

```python
def on_query(ctx, query):
    return [item(id = "q", title = "Search docs for " + query, subtitle = "Tab to pick a section",
                 scope = scope(id = "sections", label = "Docs", placeholder = "Filter sections", query = query))]

def on_scope(ctx, scope_id, query, data):
    return [item(id = s, title = s, action = action_open("https://example.com/" + s + "?q=" + http.quote(query)))
            for s in ["guide", "api", "blog"]]
```

| Function | Effect | Needs |
|---|---|---|
| `action_open(target)` | open a URL or file with the default handler | `open` |
| `action_copy(text)` | copy text to the clipboard | `clipboard.write` |
| `action_run(cmd, args=[])` | start a program | `process.run` |
| `action_call(name, data=None)` | call `on_execute` | none |

### http (permission `network`)

- `http.get(url, headers=None, timeout=10)` returns a response.
- `http.post(url, body="", json=None, headers=None, timeout=10)`; `json` is
  encoded and sets `Content-Type: application/json`.
- `http.json(url, headers=None, timeout=10)` GETs and decodes JSON, failing on
  non-2xx.
- `http.quote(text)` URL-escapes a query component.

A response has `status`, `ok`, `url`, `headers` (lower-case keys), `text` and
`json()`.

### fs (permissions `fs.read`, `fs.write`)

`fs.read_text(path)`, `fs.write_text(path, text)`, `fs.list(path)` (names, with
a trailing `/` for folders), `fs.exists(path)`.

### clipboard, notify, env

`clipboard.get()`, `clipboard.set(text)`, `notify(title, body="")`,
`env.get(name, default=None)`.

### proc (permission `process.run`)

`proc.run(cmd, args=[], timeout=10, stdin="", cwd="")` runs without a shell and
returns `code`, `ok`, `stdout`, `stderr` (each capped at 1 MiB). A non-zero exit
code is a result, not an error.

### storage (always available)

`storage.get(key, default=None)`, `storage.set(key, value)`,
`storage.delete(key)`, `storage.keys()`. Values are anything `json.encode`
accepts. Data lives in `<DataDir>/ext/<id>/storage.json` and survives updates.

### json, time

Starlark's standard modules: `json.encode/decode/encode_indent/indent` and
`time` (`time.now()`, `time.parse_time(...)`, durations, ...).

### crypto

`crypto.sha256(data)` and `crypto.md5(data)` return hex strings;
`crypto.hmac(key, data, algo="sha256")` supports `sha256`, `sha1`, `sha512`,
`md5`. `data` is a string or bytes.

### host

`host.version()` returns the running Zagros version (`"dev"` for an unversioned
build). `host.compare_versions(a, b)` compares two semantic versions and returns
-1, 0 or 1 (it raises an error for non-semver input). No permission needed.

### Tool modules

`text`, `lines`, `num`, `math`, `encode`, `hash`, `cipher`, `rand`, `uuid`,
`regex`, `net`, `dt` and `extract` are pure function libraries (string changes,
base conversion, encodings, hashes, encryption, random values, IDs, regex,
subnet math, dates, extractors). They need no permission. The reference is
generated from the registry: [Built-in tools API](tools.md). The same functions
power the launcher's [tool families](../tools.md).

### log

`log.info(...)`, `log.warn(...)`, `log.error(...)`; `print()` logs at info.

### settings

A read-only dict of the user's values merged over the manifest defaults,
typed per `type` (`int` gives an int, `number` a float). Saving settings
reloads the module, so the global always reflects the current values.

## Developing

To build extensions with a coding agent such as Claude Code, see
[Build extensions with coding agents](agents.md): `zagros ext init` writes the
`.pyi` API stubs and an `AGENTS.md` guide next to your code.

```
zagros ext init my-tool      # new project: manifest, main.star, API stubs (.pyi), AGENTS.md for coding agents
zagros --dev                 # hot reload: edits to files under the ext dir reload that extension
zagros ext validate .        # manifest, syntax, undefined names, undeclared permission use
zagros ext pack .            # writes <id>-<version>.zagros next to the folder
```

Examples: `examples/extensions/hello` (minimal) and
`examples/extensions/github-repos` (network, settings, alternate actions).

## For the host application

Go package `internal/ext`:

- `ext.NewManager(ext.Options{ExtDir, DataDir, Prompter, Platform, Executor, AppVersion, Log})`
  then `Discover(ctx)`. `Providers()` yields `core.Provider`s that also
  implement `Keywords() []string`.
- Route `core.Action` of kind `"ext-call"` to `Manager.Execute`, or wrap the
  app executor with `Manager.WrapExecutor`.
- The UI implements `ext.PermissionPrompter` (consent dialog) and
  `ext.Platform` (clipboard, toast notifications).
- `Install*`, `Uninstall`, `SetEnabled`, `Reload`, `SetSettings`, `WatchDev`
  back the settings pages and CLI.

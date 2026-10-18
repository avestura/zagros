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
| `network` | `http.get/post/put/patch/delete/head/request/download/json` | `network:host` allows that host and its subdomains. Redirects are checked per hop. Unscoped means any host |
| `fs.read` | `fs.read_text/read_bytes/list/exists/stat/glob/walk`, the source of `fs.copy`, inputs of `compress.create/extract/list` | `fs.read:~/Documents` (path prefix, `~` expands, symlinks resolved). The extension's own folder is always readable |
| `fs.write` | `fs.write_text/write_bytes/append_text/mkdir/remove/rename`, the target of `fs.copy`, `http.download` and `compress.create/extract` outputs | same as `fs.read` |
| `clipboard.read` | `clipboard.get` | none |
| `clipboard.write` | `clipboard.set`, `action_copy` | none |
| `process.run` | `proc.run`, `action_run` | `process.run:git` (executable name, `.exe` optional) or a full path |
| `env.read` | `env.get` | `env.read:USERNAME` (one variable) |
| `notifications` | `notify` | none |
| `open` | `action_open` (URLs and files) | none |
| `secrets` | `secrets.set/get/delete` | none. Entries are private to the extension id |
| `registry.read` | `windows.registry_get/registry_keys/registry_values` | `registry.read:HKCU\Software\Vendor` (key prefix, case-insensitive, matches that key and everything below it). Only `HKCU` and `HKLM`. Unscoped means any key under those roots |
| `system.read` | `windows.services`, `service_status`, `list_windows`, `processes`, `known_folder` | none. Read-only; window titles and process names can reveal what the user is doing |
| `storage` | `storage.*`, `sqlite.*` | none. Always granted, you need not declare it |

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
result (it feeds the usage history and the remembered picks: the row a user runs for a query, or inside a scope, comes first next time, so never put a counter or the typed text in it). `icon` is a file path (the shell icon is
used) or a glyph name. Without `score`, results keep the order you return
them. Without `action`, confirming calls `on_execute` with name `"default"`; if the extension defines no `on_execute`, the row is informational and Enter does nothing.
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

| Function | Effect |
|---|---|
| `http.get(url, headers=None, timeout=10)` | GET, returns a response |
| `http.head(url, headers=None, timeout=10)` | HEAD: status and headers, no body |
| `http.post/put/patch/delete(url, body="", json=None, headers=None, timeout=10)` | `json` is encoded and sets `Content-Type: application/json` |
| `http.request(method, url, headers=None, body=None, json=None, form=None, multipart=None, timeout=10, follow_redirects=True, auth=None, cookies=False, max_bytes=0)` | any method; give at most one of `body` (str or bytes), `json`, `form` (urlencoded dict) and `multipart` |
| `http.download(url, path, headers=None, timeout=60, auth=None, cookies=False, overwrite=True, max_bytes=0)` | stream a URL into a file; needs `network` and `fs.write` for `path` |
| `http.json(url, headers=None, timeout=10)` | GET and decode JSON, failing on non-2xx |
| `http.clear_cookies()` | forget the cookies kept by `cookies=True` |
| `http.quote(text)` | URL-escape a query component |

A response has `status`, `ok` (2xx), `url` (after redirects), `headers`
(lower-case keys, first value), `raw_headers` (lower-case keys, list of every
value, for `set-cookie`), `text`, `bytes` and `json()`. A non-2xx status is a
result, not an error; network failures, timeouts and caps raise.

- `auth` is a token string (sent as `Bearer`) or a `(user, password)` tuple (Basic).
- `multipart` maps field names to a str (text field), bytes (a file part named
  after the field) or a dict `{"filename", "content", "path", "content_type"}`;
  `path` reads a file and needs `fs.read` for it.
- `follow_redirects=False` returns the 3xx response. Redirects are otherwise
  checked per hop against the `network` grant.
- `cookies=True` uses an in-memory cookie jar kept per extension until the
  extension stops or `http.clear_cookies()` runs.
- Caps: response bodies 8 MiB (`max_bytes` lowers it), request bodies 8 MiB,
  downloads 256 MiB. The per-call time limit of the extension also applies.
- Errors raise. Use `attempt(http.get, url)` to get `(response, None)` or `(None, message)`.

```python
def on_query(ctx, q):
    r = http.request("POST", "https://api.example.com/items", json = {"name": q}, auth = "token")
    if not r.ok:
        return [item("err", "Failed: " + str(r.status))]
    return [item("ok", r.json()["id"])]
```

### fs (permissions `fs.read`, `fs.write`)

Paths expand `~`, are resolved through symlinks and must stay inside a granted
folder. Links inside a walked, globbed, copied or archived tree are never
followed. Failures raise (see `attempt`).

| Function | Needs | Effect |
|---|---|---|
| `fs.read_text(path)`, `fs.read_bytes(path)` | read | contents, up to 4 MiB |
| `fs.write_text(path, text)`, `fs.write_bytes(path, data)`, `fs.append_text(path, text)` | write | replace or append, up to 4 MiB per call |
| `fs.list(path)` | read | names, with a trailing `/` for folders |
| `fs.exists(path)` | read | bool |
| `fs.stat(path)` | read | entry with `name`, `path`, `size`, `mtime` (Unix seconds), `is_dir`, `is_file`, `is_symlink` |
| `fs.mkdir(path, parents=True)` | write | create a folder |
| `fs.remove(path, recursive=False, missing_ok=False)` | write | delete; returns whether something was removed. A link is removed, not its target. Granted folder roots, the home folder and drive roots are refused |
| `fs.rename(src, dst, overwrite=False)` | write on both | move or rename |
| `fs.copy(src, dst, overwrite=False)` | read `src`, write `dst` | copy a file or tree to the exact path `dst`; returns bytes copied (caps: 10000 files, 256 MiB) |
| `fs.glob(pattern)` | read | absolute paths; `*`, `?`, `[a-z]`, and `**` across folders. The scope is checked against the folder before the first wildcard |
| `fs.walk(path, max_depth=-1)` | read | entries below `path` with `rel`; at most 10000 |

### attempt

Starlark has no `try`, so host functions that raise (`fs.*`, `http.*`,
`sqlite.*`, permission errors) can be caught with
`value, err = attempt(fn, *args, **kwargs)`: on success `err` is `None`, on a
raised error `value` is `None` and `err` is the message. Timeouts and the step
limit are never caught.

```python
def on_query(ctx, q):
    text, err = attempt(fs.read_text, "~/Documents/todo.txt")
    if err != None:
        return [item("none", "No todo file yet")]
    return [item("t", text.split("\n")[0])]
```

### compress (bytes: no permission; archives: `fs.read` / `fs.write`)

- `compress.gzip/zlib/deflate(data, level=-1)` and
  `compress.gunzip/unzlib/inflate/bunzip2(data, max_bytes=0)` work on bytes.
  Decompression is capped at 4 MiB (or a lower `max_bytes`) and fails beyond it.
- `compress.create(dest, sources, format="")` writes a `zip`, `tar` or `tar.gz`
  archive (format from the extension of `dest`) of files and folders;
  `compress.extract(archive, dest, format="", overwrite=False)` unpacks into a
  folder and returns the file names; `compress.list(archive)` returns entries
  with `name`, `size`, `is_dir`, `mtime`.
- Entries with absolute names, `..` or drive letters are rejected (zip-slip),
  symlinks and special files are skipped, at most 10000 entries and 256 MiB are
  extracted. Sources need `fs.read`, destinations `fs.write`.
- zstd is not provided (no pure Go implementation in the standard library).

### binary (no permission)

Pure helpers for bytes (the name `bytes` is Starlark's own builtin, which you
still use for `bytes("text")`, slicing and `+`): `binary.pack(fmt, *values)`,
`unpack(fmt, data)`, `unpack_from(fmt, data, offset=0)`, `calcsize(fmt)`
(struct-style formats, `<` little endian by default, `>` big), `hex(data, sep="")`,
`unhex(text)`, `to_int(data, byteorder="big", signed=False)`,
`from_int(value, size, byteorder="big", signed=False)`, `xor(data, key)`,
`chunks(data, size)`, `reverse(data)`.

### sqlite (storage, always granted)

`db = sqlite.open(name="main")` opens a private SQLite database stored at
`<DataDir>/ext/<id>/db/<name>.sqlite3`. There is no way to name another path.

- `db.query(sql, params=None)` returns a list of dicts; `db.exec(sql, params=None)`
  returns `rows_affected` and `last_insert_id`. `params` is a list for `?` or a
  dict for `:name`; values are None, bool, int, float, str or bytes.
- `db.transaction(fn)` calls `fn(tx)` with a `tx` that has `query` and `exec`;
  it commits when `fn` returns and rolls back when it raises.
- `db.close()` closes early; databases are also closed when the extension stops.
- Limits: 64 MiB per database, 10000 rows per query (add `LIMIT`). `ATTACH`,
  `VACUUM INTO`, `load_extension` and `PRAGMA`s outside a short read-style list
  (`foreign_keys`, `user_version`, `table_info`, `journal_mode`, ...) are refused,
  so a script cannot reach another file or lift its size cap.
- Pure Go (`modernc.org/sqlite`), no CGO.

```python
def on_query(ctx, q):
    db = sqlite.open("notes")
    db.exec("CREATE TABLE IF NOT EXISTS notes (id INTEGER PRIMARY KEY, title TEXT)")
    if q.startswith("add "):
        db.exec("INSERT INTO notes(title) VALUES (?)", [q[4:]])
    rows = db.query("SELECT id, title FROM notes WHERE title LIKE ? LIMIT 20", ["%" + q + "%"])
    return [item(str(r["id"]), r["title"]) for r in rows]
```

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

### secrets (permission `secrets`)

`secrets.set(key, value)`, `secrets.get(key, default=None)`,
`secrets.delete(key)`. Values live in Windows Credential Manager under
`zagros|<extension id>|<key>`, encrypted for the current user. The id is part
of the name the host builds, so an extension cannot read or overwrite another
extension's entries. Keys are 1 to 256 bytes, values are strings up to 2560
bytes. Values are never logged. On non-Windows systems the calls raise an
error. Use this for tokens and passwords; `storage` and settings are plain
files.

```python
def on_query(ctx, q):
    token = secrets.get("api_token")
    if token == None:
        return [item("login", "Set your API token first")]
    return []
```

### windows (permissions `registry.read`, `system.read`)

Read-only system information, Windows only (other systems raise an error).
There are no write, start, stop or kill operations. To open a file or URL use
`action_open` (permission `open`).

| Function | Permission | Returns |
|---|---|---|
| `windows.registry_get(key, name="")` | `registry.read` | string, int, list of strings or bytes; `None` when the value is missing |
| `windows.registry_keys(key)` | `registry.read` | subkey names |
| `windows.registry_values(key)` | `registry.read` | `{name: value}` |
| `windows.services()` | `system.read` | list of `{name, display_name, state, start_type, pid}` |
| `windows.service_status(name)` | `system.read` | the same dict for one service |
| `windows.list_windows()` | `system.read` | visible titled windows: `{title, process, pid}` |
| `windows.processes()` | `system.read` | `{name, pid}` for each process |
| `windows.known_folder(name)` | `system.read` | path; names: documents, downloads, desktop, pictures, music, videos, appdata, localappdata, programdata, profile |

Registry keys look like `HKCU\Software\Vendor` (`HKEY_CURRENT_USER` and `/`
also work). The scope of `registry.read` is a key prefix, so
`registry.read:HKLM\SOFTWARE\Microsoft\Windows NT` lets the extension read
that key and its subkeys but not `HKLM\SOFTWARE\Other`. Declare the narrowest
prefix that works.

```python
def on_query(ctx, q):
    build = windows.registry_get("HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", "CurrentBuild")
    spooler = windows.service_status("Spooler")["state"]
    return [item("w", "Build " + str(build), subtitle = "Spooler " + spooler)]
```

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

Nine more modules work on structured data and are also pure: `yaml`, `toml`,
`csv`, `xml` and `html` parse text into dicts and lists and write them back;
`url`, `path`, `template` and `jwt` build, take apart and render URLs, file
paths, text templates and JSON Web Tokens. Parsers return dicts that keep the
key order of the source, lists, strings, ints, floats, bools and `None`
(stub type `Any`), and the writers accept the same shapes. Malformed input
raises a normal error with a position, never a crash. `jwt.verify` and
`jwt.sign` read the clock only when `now` is 0. Hash, HMAC, PBKDF2, scrypt,
Argon2 and the RSA, ECDSA and Ed25519 primitives already live in `hash` and
`cipher`, so `jwt` only wraps them in the token format.

### log

`log.info(...)`, `log.warn(...)`, `log.error(...)`; `print()` logs at info.

### settings

A read-only dict of the user's values merged over the manifest defaults,
typed per `type` (`int` gives an int, `number` a float). Saving settings
reloads the module, so the global always reflects the current values.

## Developing

To build extensions with a coding agent such as Claude Code, see
[Build extensions with coding agents](agents.md): `zagros ext init` writes the
`.pyi` API stubs, an `AGENTS.md` guide and a grep-able `API_INDEX.md` (one line
per function, with synonyms) next to your code.

Every stub is a type description, not runnable Starlark: do not `import` it. A
function docstring carries `Example:` (valid Starlark), `Keywords:` (synonyms
and task phrases that the MCP `search_api` and `API_INDEX.md` match), an optional
`Task:` (a headline task for the cheat sheet) and `Needs:` (the permission).
New API functions must have all of them; tests check.

```
zagros ext init my-tool      # new project: manifest, main.star, API stubs (.pyi), AGENTS.md and API_INDEX.md for coding agents
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

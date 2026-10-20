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
usage = "Type 'w <city>' for the forecast."   # shown in the extension manager
examples = ["w oslo", "w new york"]            # queries with a Try button in the manager

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
| `usage` | no | one line telling the user how to run the extension. The extension manager shows it under "How to use". `validate` warns when an extension has neither keywords nor usage |
| `examples` | no | queries the manager lists with a **Try** button that types them into the launcher. Without it the manager offers each keyword. `validate` warns when keywords are set and examples is empty |
| `permissions` | no | see below |
| `[[permission_detail]]` | no | `permission` + `reason`; `permissions` is a plain array, so details live in their own table array |
| `ai_budget_tokens` | no | tokens per day the `ai` module may spend for this extension (input plus output). Default 50000, at most 100000000. Needs the `ai` permission |
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
| `network` | `http.get/post/put/patch/delete/head/request/download/json`, `tcp.*`, `udp.*`, `tls.*`, `dns.lookup` | `network:host` allows that host and its subdomains on any port; `network:host:port` limits `tcp`, `udp` and `tls` to that port (`http` and `dns` check the host only, so a scope with a port does not cover them). Redirects are checked per hop. Unscoped means any host |
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
| `ssh` | `ssh.connect` and the connection's `exec/get/put/download/upload/list/stat/remove/mkdir` | `ssh:user@host:port`, `ssh:user@host` (any port) or `ssh:host` (any user and port). The host matches exactly, subdomains do not count. A login with `auth={"secret": ...}` also needs `secrets` |
| `system.read` |, `service_status`, `list_windows`, `processes`, `known_folder` | none. Read-only; window titles and process names can reveal what the user is doing |
| `sql` | `sql.open` (PostgreSQL, MySQL/MariaDB) | `sql:db.example.com:5432` (host, optional port; a host also covers its subdomains, a bare host covers every port). A `{"secret": name}` password also needs `secrets` |
| `ldap` | `ldap.connect`, `search`, `check_bind` | `ldap:ldap.corp.example.com` (host, optional port). Read-only directory access |
| `ldap.write` | `add`, `modify`, `delete` on a connection | same scope as `ldap`. Separate so a lookup extension never gets write access; it also lets the extension connect, but not search |
| `ai` | `ai.complete/run/embed/providers` | `ai:claude` (a provider name from `zagros.toml`, case-insensitive). Unscoped means any configured provider. Spends the manifest `ai_budget_tokens` per day |
| `storage` | `storage.*`, `sqlite.*` | none. Always granted, you need not declare it |

A user may grant fewer permissions than requested, never more than declared.

## Runtime model

- `main.star` is executed once when the extension starts, then its globals are
  frozen. Mutable state belongs in `storage`, not in globals.
- Each hook call runs in a fresh thread. Defaults: 20 million Starlark steps
  and 5 seconds wall clock per call; exceeding either aborts the call with an
  error (logged, the launcher is unaffected). Cancelling a query cancels the
  script.
- Slow work does not fit that limit. `task(fn, args)` runs a top-level function  in its own thread with 5 minutes and 200 million steps (see Tasks below);  the hook that started it keeps its own limit.
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

`item(id, title, subtitle="", icon="", action=None, alt=[], score=0.0, complete="", scope=None, pending=False)`
Build a result. `id` and `title` are required; `id` must be stable for the same
result (it feeds the usage history and the remembered picks: the row a user runs for a query, or inside a scope, comes first next time, so never put a counter or the typed text in it). `icon` is a file path (the shell icon is
used) or a glyph name. Without `score`, results keep the order you return
them. Without `action`, confirming calls `on_execute` with name `"default"`; if the extension defines no `on_execute`, the row is informational and Enter does nothing.
`alt` is a list of `{"name": ..., "action": ...}` dicts (or `(name, action)`
tuples) shown in the actions menu.

`complete` is the query `Tab` puts in the input box instead of opening
anything (for example `"wiki "`). `scope` makes `Tab` enter a context (below).

### Tasks: slow work without blocking the launcher (no permission)

Hooks stop after 5 seconds and Starlark has no threads, so a slow API call or a
large download cannot run inside `on_query`. `task` moves it to a background
run.

| Function | Effect |
|---|---|
| `task(fn, args=None, ttl=30)` | start `fn(ctx, args)` in the background and return a handle at once. `fn` is the name of a top-level function, `args` any JSON value. `ttl` (2 to 600 seconds) is how long a finished result is reused |
| `item(..., pending=True)` | placeholder row: subtitle "Loading..." unless you give one, Enter does nothing |

The handle has `id`, `done`, `ok` (done and no error), `value` (what `fn`
returned, frozen) and `error` (message or `None`).

How it works:

- The first `task("fn", args)` call starts the run. A call with the same name
  and args while it runs, or within `ttl` after it ended, returns the same run
  instead of starting another. Failures are kept for `ttl` too, so a broken
  service is not retried on every keystroke.
- When a run ends, Zagros runs the current query again. Your `on_query` calls
  `task("fn", args)` again and now gets the finished handle. There is no
  callback to write.
- A task is cancelled when a later query no longer asks for it (the user kept
  typing), when the extension stops or reloads, and when it passes its limit.
  Cancelling cancels its pending HTTP requests. Tasks started from `on_execute`
  or `on_load` are not tied to a query and run to the end.
- Limits: 5 minutes and 200 million steps per run (hooks: 5 seconds and 20
  million), at most 8 runs at once, 64 results kept. A task cannot start a task.
- The function can use any module and each call is checked against the granted
  permissions as usual. Inside a task, `http.download`, `compress.create` and
  `compress.extract` cap their totals at 1 GiB instead of 256 MiB, and `sqlite`
  and `http` calls get the task's time. Row, file-count and database-size caps
  are the same everywhere.

```python
def fetch_repos(ctx, user):
    return http.json("https://api.github.com/users/" + user + "/repos", timeout = 30.0)

def on_query(ctx, query):
    t = task("fetch_repos", query)
    if not t.done:
        return [item(id = "wait", title = "Repos of " + query, pending = True)]
    if not t.ok:
        return [item(id = "err", title = "Failed", subtitle = t.error)]
    return [item(id = r["name"], title = r["name"]) for r in t.value]
```

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
| `http.fetch_all(requests, max_concurrency=4)` | send many requests in parallel; each entry is a URL string or a dict of `http.request` arguments (`method` defaults to GET). Returns a list in the same order |
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
- `http.fetch_all` runs the requests on goroutines while the script waits, at most 100 per call and `max_concurrency` 1 to 8. Each entry has the same network scoping, redirect checks and 8 MiB cap as `http.request`; one call may return 32 MiB in all. A request that fails (permission, timeout, cap, network) does not raise: its entry has `ok = False`, `status = 0`, `url` and `error`, so check `r.ok` for each. A bad argument raises before anything is sent. The hook's 5 second limit covers the whole call, so use it from a `task` when the batch is large.

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

### sql (permission `sql`)

`db = sql.open(driver, host, port=0, database="", user="", password=None, tls=True, timeout=30)`
connects to a PostgreSQL (`"postgres"`, port 5432) or MySQL/MariaDB (`"mysql"`,
3306) server. The permission is scoped by `host[:port]`. `password` is a string
or `{"secret": name}`, read from the extension's own `secrets` (which needs the
`secrets` permission too).

The handle has the same methods as `sqlite`: `db.query(sql, params=None)`,
`db.exec(sql, params=None)`, `db.transaction(fn)` and `db.close()`. Placeholders
are `?` with a list or `:name` with a dict for both servers; the host rewrites
them to `$1..` for PostgreSQL (`??` is a literal `?`, and a query that already
uses `$1` with a list passes through).

- Text columns are `str`, `BLOB`/`BYTEA`/`BINARY` columns are `bytes`, dates and
  times are RFC 3339 strings, and `numeric`/`decimal` are strings.
- TLS is on and the certificate is verified against the system roots. `tls=False`
  turns encryption off; there is no way to skip verification.
- Every statement runs under `timeout` seconds (1 to 300), a query returns at
  most 10000 rows, an extension holds at most 8 connections, and all are closed
  when it stops. The connection is one session; after a network drop, open a new one.
- `last_insert_id` is filled on MySQL only; on PostgreSQL use `RETURNING` with `query`.

```python
def on_query(ctx, q):
    db = sql.open("postgres", "db.example.com", database="app", user="reader",
                  password={"secret": "db_password"})
    rows = db.query("SELECT id, title FROM docs WHERE title ILIKE ? LIMIT 20", ["%" + q + "%"])
    return [item(str(r["id"]), r["title"]) for r in rows]
```

### ldap (permissions `ldap`, `ldap.write`)

`conn = ldap.connect(url, bind_dn="", password=None, start_tls=False, timeout=15)`
opens `ldap://host[:389]` or `ldaps://host[:636]`. Give `bind_dn` and `password`
together (a string or `{"secret": name}`), or neither for an anonymous connection.
The certificate is verified against the system roots for `ldaps://` and
`start_tls=True`, with no way to skip it, and a password is refused over an
unencrypted connection unless the host is this machine.

- `conn.search(base, filter="(objectClass=*)", scope="sub", attributes=None, size_limit=1000)`
  returns `[{"dn": "...", "cn": ["Alice"], "mail": [...]}]`. Every attribute is a
  list; values that are not UTF-8 (`objectGUID`) are `bytes`. `size_limit` is capped at 10000.
- `conn.check_bind(dn, password)` binds as that user on a separate connection and
  returns `True`, or `False` for invalid credentials. An empty password raises.
- `conn.add(dn, attributes)`, `conn.modify(dn, replace=None, add=None, delete=None)`
  (`delete` is a dict of values or a list of attribute names) and `conn.delete(dn)`
  need `ldap.write`.
- `conn.close()`; connections also close when the extension stops (at most 8 open).
- `ldap.escape_filter(value)` and `ldap.escape_dn(value)` are pure and need no
  permission. Use them for any user input placed in a filter or a DN.

```python
def on_query(ctx, q):
    conn = ldap.connect("ldaps://ldap.corp.example.com", bind_dn="cn=svc,dc=example,dc=com",
                        password={"secret": "ldap_password"})
    f = "(&(objectClass=person)(cn=" + ldap.escape_filter(q) + "*))"
    people = conn.search("ou=people,dc=example,dc=com", f, attributes=["cn", "mail"], size_limit=20)
    return [item(p["dn"], p["cn"][0], subtitle=", ".join(p.get("mail", []))) for p in people]
```

### tcp, udp, tls, dns (permission `network`)

Raw sockets for protocols that are not HTTP. They use the `network` permission,
scoped by host or `host:port`.

| Function | Notes |
|---|---|
| `tcp.connect(host, port, timeout=5)` | returns a connection with `send(data)`, `recv(max=4096, timeout=5)`, `recv_until(delim, max=65536, timeout=5)` and `close()`; data is `str` or `bytes`, reads return `bytes` |
| `tcp.request(host, port, data, until=None, max=65536, timeout=5)` | connect, send, read the reply, close |
| `udp.open(host, port)`, `udp.send(host, port, data)`, `udp.request(host, port, data, max=4096, timeout=2)` | datagrams; `open` gives a socket with `send`, `recv`, `close` |
| `tls.connect(host, port=443, timeout=5, server_name=None)` | like `tcp.connect` over TLS 1.2+; the certificate and host name are always verified |
| `tls.cert(host, port=443, timeout=5, server_name=None)` | dict with `subject`, `issuer`, `not_before`, `not_after`, `days_left`, `expired`, `sans`, `serial`, `fingerprint_sha256`, `tls_version`, `verified`, `verify_error`, `chain`; reads expired and self-signed certificates and sends no data |
| `dns.lookup(name, type="A", server=None, timeout=5)` | `A`, `AAAA`, `MX` (dicts with `host`, `pref`), `TXT`, `CNAME`, `NS`, `PTR` (name is an IP). `server="1.1.1.1"` asks only that server, which needs its own `network` grant |

A name is resolved once, the result is checked, and that address is dialled. A
name that resolves to a loopback, private or link-local address is refused when
the grant is scoped, unless the name is `localhost`; to reach a machine on your
network, grant its IP address (`network:192.168.1.20`) and connect by address.
An unscoped `network` grant skips this check. Each connection lives at most two
minutes, carries at most 8 MiB in each direction, and is closed when the call
that opened it ends or the extension stops; at most 8 are open at once. A read
that times out raises (`attempt(conn.recv, 100)`).

```python
def on_query(ctx, q):
    c = tcp.connect("mail.example.com", 25)
    banner = c.recv_until("\r\n")
    c.send("QUIT\r\n")
    c.close()
    info = tls.cert("example.com")
    return [item("1", str(banner), subtitle="cert expires in %d days" % info["days_left"])]
```

### ssh (permission `ssh`)

`conn = ssh.connect(host, user, port=22, auth, timeout=10, host_key=None, trust_new=True)`.
`auth` is a dict and is required: `{"secret": "name"}` (password kept with
`secrets.set`), `{"key_secret": "name"}` (private key PEM kept in secrets) with
an optional `"passphrase_secret"`; the raw strings `"password"`, `"key"` and
`"passphrase"` work too but leave the credential in the extension's files.
Reading a named secret needs the `secrets` permission as well.

| Method | Notes |
|---|---|
| `conn.exec(cmd, timeout=30, stdin=None)` | `{stdout, stderr, code, ok}`; a non-zero exit is not an error; raises on timeout or output over 8 MiB |
| `conn.get(path)`, `conn.put(path, data)` | whole remote file as `bytes` (up to 8 MiB), or write one |
| `conn.download(remote, local)`, `conn.upload(local, remote)` | stream to or from a local file; need `fs.write` / `fs.read` for the local path (cap 256 MiB) |
| `conn.list(path=".")`, `conn.stat(path)` | entries with `name`, `path`, `size`, `mtime`, `is_dir`, `is_file`, `is_symlink`, `mode` |
| `conn.remove(path)`, `conn.mkdir(path, parents=False)` | remove refuses `/` and `~` |
| `conn.close()` | also closed when the call ends |

File operations use SFTP (`github.com/pkg/sftp`) and need the server to offer it.

Host keys are always verified. The first key seen for a host and port is
recorded in the extension's data folder (`ssh_known_hosts.json`, not reachable
through `storage`). A different key later raises an error naming both
fingerprints. `host_key="SHA256:..."` (or a list) pins the fingerprint and
overrides the record, which is the way past a legitimate key change.
`trust_new=False` refuses unknown hosts, and the error shows the fingerprint to
check. `conn.fingerprint` and `conn.host_key_status` (`"pinned"`, `"known"`,
`"new"`) tell you what happened. There is no option to skip verification.

```python
def on_execute(ctx, item_id):
    c = ssh.connect("build.example.com", "deploy", auth={"key_secret": "deploy_key"})
    r = c.exec("systemctl is-active app")
    notify("app", r.stdout.strip() if r.ok else r.stderr)
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

### ai (permission `ai`)

Ask a language model, or let it call your functions. The user configures
providers in `zagros.toml` and stores each key with `zagros ai key set <name>`
(Credential Manager, never the config file). Extensions never see the key, and
errors and logs have it removed.

```toml
[ai]
default = "claude"

[[ai.provider]]
name = "claude"
kind = "anthropic"            # Messages API
model = "claude-sonnet-5-5"

[[ai.provider]]
name = "local"
kind = "openai"               # OpenAI, Ollama, LM Studio, vLLM ...
base_url = "http://localhost:11434/v1"
model = "llama3.2"
embed_model = "nomic-embed-text"
```

Plain `http://` is accepted only for localhost and private addresses (the key
travels in a header). Redirects are never followed. `zagros ai providers` lists
what is configured and `zagros ai test <name>` sends a tiny request.

- `ai.complete(messages, model="", system="", max_tokens=1024, temperature=None, tools=None, tool_choice="", provider="")`
  returns `{text, tool_calls, usage, stop_reason, model, provider}`. `messages`
  is a prompt string or a list of `{role, content}` dicts.
- `ai.run(prompt, tools=None, ..., max_steps=8)` runs the agent loop: it runs
  the tools the model calls, sends the results back and repeats until the model
  answers or `max_steps` (1 to 20) is reached. `tool_calls` in the result is the
  trace of what ran.
- `ai.tool(fn, name="", description="", params=None)` wraps a Starlark function
  for the model. `params` is shorthand (`{"city": "string"}`) or a full JSON
  schema. `ai.registry_tools(["hash.sha256", "encode"])` offers built-in tools
  functions the same way (named `group_function` to the model).
- `ai.embed(texts, model="", provider="")` returns vectors from an `openai`-kind
  provider; an `anthropic` provider raises a clear error.
- `ai.providers()` lists the providers this extension may use.

Limits. Each extension has a daily token budget, `ai_budget_tokens` in the
manifest (default 50000, resets at midnight UTC, kept across restarts). When it
is used up every call raises an error naming the budget; `max_tokens` is lowered
to fit what is left. The 5 second hook limit does not run while a request is in
flight: one request may take 60 seconds, one `ai.run` 120 seconds, and all
waits in one hook call together 180 seconds. Replies over `MaxHTTPBytes` (8 MiB),
prompts over 4 MiB, 200 messages, 64 tools and tool results over 64 KiB are capped or refused. Do not call `ai` from `on_query` on every
keystroke.

Tools run on the extension's thread, so the model can use only what the
extension is already allowed to do: a tool calling `fs.read_text` still needs
`fs.read`, and the model only picks the arguments. A tool that raises is
reported to the model as an error; a timeout aborts the run. `ai.run` cannot be
nested inside a tool. Treat web pages and file text you pass to the model as
untrusted: they can try to steer the tools you offer.

```python
def lookup(word):
    return storage.get("glossary", {}).get(word, "unknown")

def on_execute(ctx, action):
    tools = [ai.tool(lookup, description="Look a term up in the glossary", params={"word": "string"})]
    out = ai.run("Explain 'zagros' using the glossary", tools=tools, max_steps=4)
    notify("Answer", out["text"])
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

The lines are kept in memory only (nothing is written to disk). They show up in the extension manager under Log and live in the log console: right-click the tray icon and choose Log console, or run `zagros logs <id>` in a terminal. Errors your hooks raise and failed actions are logged for you, and the launcher also tells the user when your query failed or ran out of time, so a blocking `http.get` is never silent. For lookups that can take more than a moment use `task()` with `item(pending = True)`, which shows a loading row.

### settings

A read-only dict of the user's values merged over the manifest defaults,
typed per `type` (`int` gives an int, `number` a float). Saving settings
reloads the module, so the global always reflects the current values.

Users change settings in the extension manager (tray, **Extensions...**, select
the extension, **Settings**): booleans flip, choices cycle, other types open
an inline editor. The same is available from the command line:

```
zagros ext config <id>               list settings, their values and where they are stored
zagros ext config <id> <key>         print one value
zagros ext config <id> <key> <value> change it
zagros ext outdated                  list installed extensions that have a newer version in the marketplace
zagros ext update <id>|--all         install the newer version (new permissions are asked for)
```

Values are stored per extension in `<data dir>/ext/<id>/settings.json`. They
are not part of `zagros.toml` and are not in the extension folder, so updating
the extension keeps them. Give every setting a `label` and a `description`:
the manager shows them as written.

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

# Writing a Zagros extension (guide for coding agents)

Zagros is a keyboard launcher for Windows. An extension is a folder with a
`manifest.toml` and a Starlark script (`main.star`). Starlark is a Python
dialect, not Python.

## Find the API first, do not guess

The complete API is in `zagros-stubs/zagros/*.pyi` (next to this file):

| File | Contents |
|---|---|
| `__init__.pyi` | `item`, `scope`, `action_*`, `notify`, `settings`, `ctx` |
| `hooks.pyi` | `on_load`, `on_query`, `on_scope`, `on_execute`, `on_settings_changed` |
| `permissions.pyi` | every permission, what it allows, scopes |
| `http.pyi` `fs.pyi` `clipboard.pyi` `proc.pyi` `env.pyi` `storage.pyi` `crypto.pyi` `log.pyi` `json.pyi` `time.pyi` | one module each |

Search them (`def scope`, `Needs:`) before using a function. Every function
documents the permission it needs ("Needs: network").

## Workflow

1. Edit `manifest.toml` (id, name, version, `keywords`, `permissions`).
2. Edit `main.star`. Define `on_query(ctx, query)` returning `item(...)` rows.
3. `zagros ext validate .` (manifest, syntax, undefined names, undeclared permission use).
4. `zagros --dev` hot-reloads this folder while you edit; install with `zagros ext install .`.
5. `zagros ext pack .` writes `<id>-<version>.zagros`. Publishing is in the marketplace docs.

## Starlark rules

- No `import`, classes, `try/except`, `with`, f-strings or `yield`. Split code across files with `load("other.star", "name")`.
- Available: `for`, `if`, `while`, recursion, functions, lambdas, list/dict/set comprehensions, `"%s" % x` and `.format()` formatting, top-level statements.
- Everything in the stubs is already defined. Do not import it.
- Globals are frozen after load; keep state in `storage` or return it.
- Hooks run with a time and step limit; do slow work (network) sparingly and with short timeouts.

## Permissions

Declare only what you use, as narrowly as possible, in `manifest.toml`:

```toml
permissions = ["network:api.example.com", "clipboard.write"]

[[permission_detail]]
permission = "network:api.example.com"
reason = "Looks up results."
```

The user approves them. Calling a function without its permission fails.
UI features (scopes, Tab completion, alternate actions) need no permission.

## Ready-made function libraries

Before writing helper code, search the stubs: `text`, `lines`, `num`, `math`, `encode`, `hash`, `cipher`, `rand`, `uuid`, `regex`, `net`, `dt` and `extract` are pure libraries (case styles, palindromes, base conversion, base64/32/58/85, URL and HTML escaping, hashes, AES, random values, UUIDs, regex, CIDR math, dates and time zones, extracting emails and IPs). They need no permission. `host.version()` and `host.compare_versions(a, b)` report the Zagros version.

## UI mechanisms you can use

- **Results:** return `item(id, title, subtitle, icon, action, alt, score, complete, scope)`. Keep `id` stable for the same result (never a counter or the typed text): Zagros remembers which row the user ran for a query, or inside a scope, and puts it first next time.
- **Scopes (Tab sub-context):** give an item `scope = scope(id, label, placeholder, query, data)`. Tab enters it: the input shows a chip with `label` and your `on_scope(ctx, scope_id, query, data)` supplies the list. Esc or Backspace on an empty input leaves it.
- **Tab completion:** `complete = "kw "` puts text in the input on Tab.
- **Alternate actions:** `alt = [("Copy", action_copy(x))]` shows in the built-in Actions scope on Tab.
- **Actions:** `action_open`, `action_copy`, `action_run`, `action_call` (calls `on_execute`).

## Minimal example

```python
def on_query(ctx, query):
    return [item(
        id = "hello",
        title = "Hello, " + (query or "world"),
        action = action_copy("Hello, " + (query or "world")),
    )]
```

(Declare `permissions = ["clipboard.write"]` for `action_copy`.)

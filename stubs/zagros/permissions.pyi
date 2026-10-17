"""Permissions: declare in manifest.toml, the user approves them (like Android).

    permissions = ["network:api.github.com", "clipboard.write", "open"]

A permission is "kind" or "kind:scope". Calling a function without its
permission raises an error. `zagros ext validate .` flags undeclared use.

| permission        | scope                    | allows                                   |
|-------------------|--------------------------|------------------------------------------|
| network           | host (api.example.com)   | http.get / post / json to that host      |
| fs.read           | folder                   | fs.read_text, fs.list, fs.exists         |
| fs.write          | folder                   | fs.write_text                            |
| clipboard.read    | -                        | clipboard.get                            |
| clipboard.write   | -                        | clipboard.set, action_copy               |
| process.run       | executable name          | proc.run, action_run                     |
| env.read          | variable name            | env.get                                  |
| notifications     | -                        | notify                                   |
| open              | -                        | action_open                              |
| storage           | -                        | always granted: storage.*                |

No permission needed: item, scope, action_call, settings, log, json, time,
crypto, and every UI mechanism (scopes, completion, alternate actions). When
Zagros gains a host feature that reaches outside the sandbox, it gets a
permission listed here and in the docs; UI-only features do not.
"""

from typing import Literal

PermissionKind = Literal[
    "network",
    "fs.read",
    "fs.write",
    "clipboard.read",
    "clipboard.write",
    "process.run",
    "env.read",
    "notifications",
    "open",
    "storage",
]

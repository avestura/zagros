"""Permissions: declare in manifest.toml, the user approves them (like Android).

    permissions = ["network:api.github.com", "clipboard.write", "open"]

A permission is "kind" or "kind:scope". Calling a function without its
permission raises an error. `zagros ext validate .` flags undeclared use.

Type description, not runnable Starlark: this file lists the permission names for
manifest.toml; nothing here is imported or called in main.star.
Keywords: permission, access, allow, sandbox, consent, manifest permissions, network access, file access

| permission        | scope                    | allows                                   |
|-------------------|--------------------------|------------------------------------------|
| network           | host (api.example.com)   | http.get / post / put / patch / delete / head / request / download / json to that host |
| fs.read           | folder                   | fs.read_text, read_bytes, list, exists, stat, glob, walk, copy (source), compress.create sources, compress.extract and list archives |
| fs.write          | folder                   | fs.write_text, write_bytes, append_text, mkdir, remove, rename, copy (target), http.download, compress.create dest, compress.extract dest |
| clipboard.read    | -                        | clipboard.get                            |
| clipboard.write   | -                        | clipboard.set, action_copy               |
| process.run       | executable name          | proc.run, action_run                     |
| env.read          | variable name            | env.get                                  |
| notifications     | -                        | notify                                   |
| open              | -                        | action_open                              |
| secrets           | -                        | secrets.set, secrets.get, secrets.delete |
| registry.read     | registry key prefix      | windows.registry_get, registry_keys, registry_values |
| system.read       | -                        | windows.services, service_status, list_windows, processes, known_folder |
| storage           | -                        | always granted: storage.*, sqlite.*      |

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
    "secrets",
    "registry.read",
    "system.read",
]

"""Zagros extension API: the names that exist in every extension script.

Extensions are Starlark (a Python dialect), not Python. Nothing is imported:
every name below is already defined (predeclared) when `main.star` runs.
These stubs describe them for people and coding agents. Search this folder
(`def item`, `def scope`, `http.pyi`, ...) instead of guessing.

Modules (each has its own stub file): http, fs, clipboard, proc, env,
storage, crypto, log, json, time, host.

Built-in tool libraries (pure functions, no permission needed; see the
generated stub of each): text, lines, num, math, encode, hash, cipher, rand,
uuid, regex, net, dt, extract.

PERMISSIONS: every sensitive function says "Needs: <permission>". The
permission must be declared in manifest.toml and the user must approve it.
See permissions.pyi. Functions that say "Needs: none" are always allowed.

HOOKS: define these at the top level of main.star (see hooks.pyi).
"""

from typing import Any, Dict, List, Optional, Tuple, Union

from . import clipboard as clipboard
from . import crypto as crypto
from . import env as env
from . import fs as fs
from . import host as host
from . import http as http
from . import json as json
from . import log as log
from . import proc as proc
from . import storage as storage
from . import time as time

class Action:
    """Opaque value built by the action_* functions. Never construct directly."""

class Scope:
    """Opaque value built by scope(). Assign it to item(scope=...)."""

class Item:
    """Opaque value built by item(). Return a list of them from on_query / on_scope."""

class Ctx:
    """First argument of every hook."""

    id: str
    name: str
    version: str
    dir: str
    """Absolute path of the extension folder."""

AltAction = Union[Tuple[str, Action], Dict[str, Any]]
"""("Name", action) or {"name": "Name", "action": action}."""

settings: Dict[str, Any]
"""User settings merged over the manifest defaults, typed per `type`.
Read-only. Saving settings reloads the module. Needs: none."""

def item(
    id: str,
    title: str,
    subtitle: str = "",
    icon: str = "",
    action: Optional[Action] = None,
    alt: List[AltAction] = [],
    score: float = 0.0,
    complete: str = "",
    scope: Optional[Scope] = None,
) -> Item:
    """Build one result row. Needs: none.

    id: stable for the same result (feeds usage history). title: required.
    icon: a file path (the shell icon is used) or a glyph name.
    action: run on Enter. Without it, Enter calls on_execute with name "default".
    alt: secondary actions. With no scope, Tab opens the built-in "Actions" scope
        listing the default action plus these.
    score: without it, results keep the order you return them.
    complete: query that Tab puts in the input (e.g. "kw "), instead of entering
        a scope.
    scope: context that Tab enters (see scope()). Tab inside a scope does nothing.
    """

def scope(
    id: str,
    label: str,
    placeholder: str = "",
    query: str = "",
    data: Any = None,
) -> Scope:
    """Build a sub-context ("scope") that Tab enters. Needs: none.

    The launcher shows a chip with `label` in the input and lists the items
    returned by your on_scope(ctx, scope_id, query, data) hook instead of the
    root results. Esc, or Backspace on an empty input, leaves it.

    id: passed back to on_scope. label: chip text. placeholder: hint while the
    input is empty. query: text the input starts with (usually the root query).
    data: any JSON-able value passed back to on_scope.
    """

def action_open(target: str) -> Action:
    """Open a URL or file with the default handler. Needs: open."""

def action_copy(text: str) -> Action:
    """Copy text to the clipboard. Needs: clipboard.write."""

def action_run(cmd: str, args: List[str] = []) -> Action:
    """Start a program, no shell. Needs: process.run (scoped by executable)."""

def action_call(name: str, data: Any = None) -> Action:
    """Call on_execute(ctx, item) with item.name == name and item.data == data
    (round-tripped through JSON). Needs: none."""

def notify(title: str, body: str = "") -> None:
    """Show a toast notification. Needs: notifications."""

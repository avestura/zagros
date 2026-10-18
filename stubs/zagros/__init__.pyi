"""Zagros extension API: the global names that exist in every extension script (item, scope, action_*, notify, settings).

Type description, not runnable Starlark: do not `import` anything. Extensions are
Starlark (a Python dialect), not Python, and every name in these stubs is already
defined (predeclared) when `main.star` runs. These files describe them for people
and coding agents. Search this folder (`def item`, `def scope`, `http.pyi`, ...)
or API_INDEX.md instead of guessing.

Every module has its own stub file in this folder: host modules (http, fs,
clipboard, ...) and the pure tool libraries (text, encode, hash, ...). The module
table in AGENTS.md lists them all.

Docstring conventions in these stubs: `Example:` is runnable Starlark (one line,
or an indented or fenced block), `Task:` is a headline task phrase, `Keywords:`
lists synonyms and task phrases for search, `Needs:` names the permission.
Needs: none for the names below unless a function says otherwise.
Keywords: extension api, result row, launcher item, sub context, tab, action, notification, settings

PERMISSIONS: every sensitive function says "Needs: <permission>". The
permission must be declared in manifest.toml and the user must approve it.
See permissions.pyi. Functions that say "Needs: none" are always allowed.

HOOKS: define these at the top level of main.star (see hooks.pyi).
"""

from typing import Any, Callable, Dict, List, Optional, Tuple, Union

from . import binary as binary
from . import clipboard as clipboard
from . import compress as compress
from . import crypto as crypto
from . import env as env
from . import fs as fs
from . import host as host
from . import http as http
from . import json as json
from . import log as log
from . import proc as proc
from . import sqlite as sqlite
from . import secrets as secrets
from . import storage as storage
from . import time as time
from . import windows as windows

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
Read-only. Saving settings reloads the module.

Example: key = settings.get("api_key", "")
Keywords: user settings, preferences, configuration, options, manifest settings
Needs: none"""

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
    """Build one result row.

    id: stable for the same result (feeds usage history and remembered picks: the
        row a user runs for a query, or inside a scope, comes first next time; do
        not put counters or typed text in it). title: required.
    icon: a file path (the shell icon is used) or a glyph name.
    action: run on Enter. Without it, Enter calls on_execute with name "default".
    alt: secondary actions. With no scope, Tab opens the built-in "Actions" scope
        listing the default action plus these.
    score: without it, results keep the order you return them.
    complete: query that Tab puts in the input (e.g. "kw "), instead of entering
        a scope.
    scope: context that Tab enters (see scope()). Tab inside a scope does nothing.

    Example:
        item(
            id = "hello",
            title = "Hello, " + query,
            subtitle = "Enter copies it",
            action = action_copy("Hello, " + query),
        )
    Task: return a result row
    Keywords: result, row, list entry, search result, show a result, launcher item, suggestion, hit
    Needs: none
    """

def scope(
    id: str,
    label: str,
    placeholder: str = "",
    query: str = "",
    data: Any = None,
) -> Scope:
    """Build a sub-context ("scope") that Tab enters.

    The launcher shows a chip with `label` in the input and lists the items
    returned by your on_scope(ctx, scope_id, query, data) hook instead of the
    root results. Esc, or Backspace on an empty input, leaves it.

    id: passed back to on_scope. label: chip text. placeholder: hint while the
    input is empty. query: text the input starts with (usually the root query).
    data: any JSON-able value passed back to on_scope.

    Example: scope(id = "styles", label = "Style", placeholder = "Filter styles", query = query)
    Task: open a sub-list when the user presses Tab
    Keywords: tab, sub context, drill down, secondary list, chip, nested results, detail view, filter within
    Needs: none
    """

def action_open(target: str) -> Action:
    """Open a URL or file with the default handler.

    Example: action_open("https://example.com")
    Task: open a URL or file on Enter
    Keywords: open link, browse, launch file, default app, open website, show in browser
    Needs: open
    """

def action_copy(text: str) -> Action:
    """Copy text to the clipboard.

    Example: action_copy("copied text")
    Task: copy a result to the clipboard on Enter
    Keywords: copy, clipboard, put text on clipboard, copy result
    Needs: clipboard.write
    """

def action_run(cmd: str, args: List[str] = []) -> Action:
    """Start a program, no shell.

    Example: action_run("notepad.exe", ["notes.txt"])
    Task: start a program on Enter
    Keywords: launch app, execute program, run command, spawn, open application
    Needs: process.run (scoped by executable)
    """

def action_call(name: str, data: Any = None) -> Action:
    """Call on_execute(ctx, item) with item.name == name and item.data == data
    (round-tripped through JSON).

    Example: action_call("refresh", {"id": 7})
    Task: run your own code when the user confirms a row
    Keywords: callback, custom action, on_execute, handler, do something on enter
    Needs: none
    """

def notify(title: str, body: str = "") -> None:
    """Show a toast notification.

    Example: notify("Done", "3 files processed")
    Task: show a desktop notification
    Keywords: toast, alert, popup, message, tell the user, notification
    Needs: notifications
    """

def attempt(fn: Callable[..., Any], *args: Any, **kwargs: Any) -> Tuple[Any, Optional[str]]:
    """Call fn(*args, **kwargs) and return (value, None), or (None, message) when it raises.

    Starlark has no try/except, and host functions such as fs.read_text or
    http.get raise on failure, so use this to handle errors. A timeout or the
    step limit is never caught. Pass the function itself and its arguments,
    not the result of calling it. (`try` is a reserved word in Starlark.)

    Example: text, err = attempt(fs.read_text, "~/Documents/notes.txt")
    Task: handle an error from a host function
    Keywords: try, catch, except, error handling, pcall, rescue, ignore error, fallback, safe call
    Needs: none
    """

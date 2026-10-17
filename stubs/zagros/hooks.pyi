"""Hooks: top-level functions you define in main.star. Each receives ctx first.

Only on_query is needed to produce results. Define the others when you use
the matching feature. Hooks run in the sandbox with a time and step limit.
"""

from typing import Any, List, Optional

from . import Action, Ctx, Item

class ExecuteItem:
    """The `item` argument of on_execute."""

    id: str
    title: str
    name: str
    """First argument of action_call; "default" when the item had no action."""
    data: Any
    """Second argument of action_call, round-tripped through JSON."""

def on_load(ctx: Ctx) -> None:
    """Once, after the module loads."""

def on_query(ctx: Ctx, query: str) -> List[Item]:
    """Every keystroke while this extension is asked. If manifest `keywords` is set,
    you are only asked when the query starts with a keyword, and the keyword is
    already removed from `query`."""

def on_scope(ctx: Ctx, scope_id: str, query: str, data: Any) -> List[Item]:
    """Every keystroke while the user is inside a scope your item offered with
    scope(...). `scope_id` and `data` are what you passed to scope(); `query` is
    the text typed after the chip. Return the items to show in place of the root
    results. Not defined = the scope shows nothing."""

def on_execute(ctx: Ctx, item: ExecuteItem) -> Optional[Action]:
    """When an item with action_call(...) (or no action) is confirmed. May return
    action_open / action_copy / action_run to run next (not action_call)."""

def on_settings_changed(ctx: Ctx) -> None:
    """After the user saves settings. The module has been reloaded."""

"""Hooks: top-level functions you define in main.star. Each receives ctx first.

Type description, not runnable Starlark: do not `import` anything; these stubs show the signature of each function you define. Only on_query is needed to produce results. Define the others when you use the matching feature. Hooks run in the sandbox with a time and step limit (about 5 seconds and 20M steps per call); work that needs longer goes in a function started with task(), which has its own larger limit and hands the result to the next on_query.
Needs: none.
Keywords: hook, callback, entry point, handler, on_query, lifecycle, main.star functions
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
    """Once, after the module loads.

    Example:
        def on_load(ctx):
            log.info("loaded", ctx.version)
    Keywords: startup, init, initialize, setup, after load, start
    Needs: none
    """

def on_query(ctx: Ctx, query: str) -> List[Item]:
    """Every keystroke while this extension is asked. If manifest `keywords` is set,
    you are only asked when the query starts with a keyword, and the keyword is
    already removed from `query`.

    Example:
        def on_query(ctx, query):
            return [item(id = "echo", title = query, action = action_copy(query))]
    Task: produce launcher results for what the user typed
    Keywords: search handler, results, typed text, keystroke, provide rows, query handler, answer a query
    Needs: none
    """

def on_scope(ctx: Ctx, scope_id: str, query: str, data: Any) -> List[Item]:
    """Every keystroke while the user is inside a scope your item offered with
    scope(...). `scope_id` and `data` are what you passed to scope(); `query` is
    the text typed after the chip. Return the items to show in place of the root
    results. Not defined = the scope shows nothing.

    Example:
        def on_scope(ctx, scope_id, query, data):
            return [item(id = "a", title = "A: " + query)]
    Task: fill a Tab sub-list
    Keywords: tab, sub context, scope handler, drill down results, nested list, chip
    Needs: none
    """

def on_execute(ctx: Ctx, item: ExecuteItem) -> Optional[Action]:
    """When an item with action_call(...) (or no action) is confirmed. May return
    action_open / action_copy / action_run to run next (not action_call).

    Example:
        def on_execute(ctx, it):
            if it.name == "refresh":
                return action_open("https://example.com")
    Task: handle Enter on a row with a custom action
    Keywords: enter, confirm, run, custom action handler, action_call target, click
    Needs: none
    """

def on_settings_changed(ctx: Ctx) -> None:
    """After the user saves settings. The module has been reloaded.

    Example:
        def on_settings_changed(ctx):
            log.info("settings saved")
    Keywords: settings saved, preferences changed, reload, configuration update
    Needs: none
    """

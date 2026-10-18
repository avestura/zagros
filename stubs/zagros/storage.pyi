"""storage: a per-extension key/value store that survives restarts and updates.

Values are anything json.encode accepts. Data lives in <DataDir>/ext/<id>/storage.json.
Type description, not runnable Starlark: do not `import` anything. The name `storage` is predeclared in main.star; call `storage.get(...)` directly.
Needs: none (always granted).
Keywords: persist, remember, save state, key value store, settings, cache, database
"""

from typing import Any, List

def get(key: str, default: Any = None) -> Any:
    """Read a value; returns default (None) when the key is missing.

    Example: n = storage.get("count", 0)
    Task: remember a value between runs
    Keywords: load state, read saved value, recall, persisted value, cache lookup
    Needs: none
    """

def set(key: str, value: Any) -> None:
    """Write a value.

    Example: storage.set("count", 3)
    Keywords: save state, remember, persist value, store data, cache write
    Needs: none
    """

def delete(key: str) -> None:
    """Remove a key.

    Example: storage.delete("count")
    Keywords: forget, remove saved value, clear key, erase state
    Needs: none
    """

def keys() -> List[str]:
    """All keys.

    Example: names = storage.keys()
    Keywords: list saved keys, enumerate storage, what is stored
    Needs: none
    """

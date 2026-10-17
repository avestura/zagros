"""storage module: per-extension key/value store. Needs: none (always granted).

Values are anything json.encode accepts. Data lives in
<DataDir>/ext/<id>/storage.json and survives updates.
"""

from typing import Any, List

def get(key: str, default: Any = None) -> Any:
    """Read a value."""

def set(key: str, value: Any) -> None:
    """Write a value."""

def delete(key: str) -> None:
    """Remove a key."""

def keys() -> List[str]:
    """All keys."""

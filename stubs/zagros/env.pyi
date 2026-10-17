"""env module."""

from typing import Optional

def get(name: str, default: Optional[str] = None) -> Optional[str]:
    """Read an environment variable. Needs: env.read (scoped by name)."""

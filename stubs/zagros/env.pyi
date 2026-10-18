"""env: read environment variables.

Type description, not runnable Starlark: do not `import` anything. The name `env` is predeclared in main.star; call `env.get(...)` directly.
Needs: env.read, scoped by variable name (env.read:HOME).
Keywords: environment variable, getenv, PATH, USERPROFILE, system setting
"""

from typing import Optional

def get(name: str, default: Optional[str] = None) -> Optional[str]:
    """Read an environment variable; returns default (None) when it is not set.

    Example: home = env.get("USERPROFILE", "")
    Task: read an environment variable
    Keywords: getenv, environment variable, os environ, system variable, read variable
    Needs: env.read (scoped by name)
    """

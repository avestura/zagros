"""log: write to <DataDir>/zagros.log; print() logs at info.

Type description, not runnable Starlark: do not `import` anything. The name `log` is predeclared in main.star; call `log.info(...)` directly.
Needs: none.
Keywords: logging, debug, print, trace, diagnostics, console output
"""

from typing import Any

def info(*args: Any) -> None:
    """Log at info level.

    Example: log.info("loaded", 3, "items")
    Task: write a debug message to the log
    Keywords: print, debug, trace, console log, diagnostic message
    Needs: none
    """

def warn(*args: Any) -> None:
    """Log at warning level.

    Example: log.warn("slow response")
    Keywords: warning, caution, log a problem
    Needs: none
    """

def error(*args: Any) -> None:
    """Log at error level.

    Example: log.error("request failed")
    Keywords: failure, log an error, report problem
    Needs: none
    """

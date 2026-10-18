"""proc: run a program without a shell and read its output.

Type description, not runnable Starlark: do not `import` anything. The name `proc` is predeclared in main.star; call `proc.run(...)` directly.
Needs: process.run, scoped by executable name (process.run:git).
Keywords: run command, execute program, subprocess, shell out, cli, exec
"""

from typing import List

class Result:
    code: int
    ok: bool
    stdout: str
    """Capped at 1 MiB."""
    stderr: str
    """Capped at 1 MiB."""

def run(cmd: str, args: List[str] = [], timeout: int = 10, stdin: str = "", cwd: str = "") -> Result:
    """Run a program without a shell. A non-zero exit code is a result, not an
    error.

    Example: r = proc.run("git", ["status", "--short"], cwd = "C:/src/repo")
    Task: run a command and capture its output
    Keywords: execute program, subprocess, shell command, call cli, spawn process, exec, stdout
    Needs: process.run (scoped by executable)
    """

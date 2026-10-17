"""proc module. Needs: process.run, scoped by executable name (process.run:git)."""

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
    error. Needs: process.run."""

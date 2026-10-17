"""fs module. Needs: fs.read / fs.write, scoped by folder (fs.read:~/Documents)."""

from typing import List

def read_text(path: str) -> str:
    """Read a UTF-8 file. Needs: fs.read."""

def write_text(path: str, text: str) -> None:
    """Write a UTF-8 file. Needs: fs.write."""

def list(path: str) -> List[str]:
    """Names in a folder; folders end with "/". Needs: fs.read."""

def exists(path: str) -> bool:
    """Whether the path exists. Needs: fs.read."""

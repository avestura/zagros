"""fs: read, write, copy, move, delete, search and inspect files and folders.

Type description, not runnable Starlark: do not `import` anything. The name `fs` is predeclared in main.star; call `fs.read_text(...)` directly.
Needs: fs.read / fs.write, scoped by folder (fs.read:~/Documents).
Keywords: file, folder, directory, read file, write file, disk, path exists, copy, move, delete, glob, walk

Errors: fs functions raise on failure (missing file, permission denied, over a
size cap). Catch with `value, err = attempt(fs.read_text, path)`, or test first
with `fs.exists(path)`. Paths accept `~`, are resolved through symlinks before
the scope check, and cannot leave the granted folder. Reads and writes are
capped at 4 MiB per call; walk, glob and copy stop at 10000 entries; copy at
256 MiB.
"""

from typing import Any, List

class Entry:
    """Returned by stat and walk."""

    name: str
    path: str
    """Absolute path with forward slashes."""
    rel: str
    """walk only: path relative to the walked folder, forward slashes."""
    size: int
    mtime: int
    """Modification time, Unix seconds."""
    is_dir: bool
    is_file: bool
    is_symlink: bool

def read_text(path: str) -> str:
    """Read a UTF-8 file.

    Example: body = fs.read_text("~/Documents/notes.txt")
    Task: read a text file
    Keywords: open file, load file, cat, file contents, read from disk
    Needs: fs.read
    """

def read_bytes(path: str) -> bytes:
    """Read a file as bytes (up to 4 MiB).

    Example: data = fs.read_bytes("~/Documents/logo.png")
    Task: read a binary file
    Keywords: binary file, raw bytes, read image, load blob, read archive, file bytes
    Needs: fs.read
    """

def write_text(path: str, text: str) -> None:
    """Write a UTF-8 file, replacing it.

    Example: fs.write_text("~/Documents/out.txt", "hello")
    Task: write a text file
    Keywords: save file, create file, store to disk, overwrite file, export text
    Needs: fs.write
    """

def write_bytes(path: str, data: bytes) -> None:
    """Write bytes to a file, replacing it (up to 4 MiB).

    Example: fs.write_bytes("~/Documents/out.bin", binary.unhex("00ff10"))
    Task: write a binary file
    Keywords: save binary, write blob, raw bytes, store image, save download
    Needs: fs.write
    """

def append_text(path: str, text: str) -> None:
    """Append text to a file, creating it when missing.

    Example: fs.append_text("~/Documents/log.txt", "one more line\\n")
    Task: append to a file
    Keywords: add line, log to file, tail, write at end, journal
    Needs: fs.write
    """

def list(path: str) -> List[str]:
    """Names in a folder; folders end with "/".

    Example: names = fs.list("~/Documents")
    Task: list a folder
    Keywords: ls, dir, directory listing, files in folder, browse folder
    Needs: fs.read
    """

def exists(path: str) -> bool:
    """Whether the path exists.

    Example: found = fs.exists("~/Documents/notes.txt")
    Keywords: file exists, path exists, is there a file, check path
    Needs: fs.read
    """

def stat(path: str) -> Entry:
    """Size, modification time and kind of a file or folder; raises when missing.

    Example: info = fs.stat("~/Documents/notes.txt")
    Task: get file size or modified time
    Keywords: file size, modified time, mtime, is directory, file info, metadata, last modified
    Needs: fs.read
    """

def mkdir(path: str, parents: bool = True) -> None:
    """Create a folder. With parents (default) missing parents are created and an existing folder is fine.

    Example: fs.mkdir("~/Documents/reports/2026")
    Task: create a folder
    Keywords: make directory, new folder, mkdir -p, create directory
    Needs: fs.write
    """

def remove(path: str, recursive: bool = False, missing_ok: bool = False) -> bool:
    """Delete a file or folder; returns whether something was removed. A folder needs recursive=True unless empty. A link is removed, not its target. Granted folder roots, the home folder and drive roots are refused.

    Example: fs.remove("~/Documents/old.txt", missing_ok = True)
    Task: delete a file or folder
    Keywords: rm, delete file, remove folder, unlink, clean up, rmdir
    Needs: fs.write
    """

def rename(src: str, dst: str, overwrite: bool = False) -> None:
    """Move or rename a file or folder (same drive). Both paths need fs.write.

    Example: fs.rename("~/Documents/a.txt", "~/Documents/b.txt")
    Task: rename or move a file
    Keywords: mv, move file, rename file, relocate
    Needs: fs.write
    """

def copy(src: str, dst: str, overwrite: bool = False) -> int:
    """Copy a file or a whole folder tree to the exact path dst; returns bytes copied. Links inside a tree are skipped. src needs fs.read, dst needs fs.write.

    Example: fs.copy("~/Documents/a.txt", "~/Documents/backup/a.txt")
    Task: copy a file or folder
    Keywords: cp, duplicate file, backup, copy folder, clone directory
    Needs: fs.read for src, fs.write for dst
    """

def glob(pattern: str) -> List[str]:
    """Absolute paths matching a pattern. `*` and `?` match inside one name, `**` spans folders. The folder part before the first wildcard is what the scope is checked against. Links are not followed.

    Example: logs = fs.glob("~/Documents/**/*.log")
    Task: find files by pattern
    Keywords: find files, wildcard, search files, match extension, *.txt, recursive search
    Needs: fs.read
    """

def walk(path: str, max_depth: int = -1) -> List[Entry]:
    """Every file and folder below path (not path itself), folders before their contents. max_depth -1 is unlimited, 1 is only the direct children. Links are listed but not followed.

    Example: big = [e for e in fs.walk("~/Documents") if e.is_file and e.size > 1000000]
    Task: list a folder recursively
    Keywords: recursive listing, tree, find all files, directory tree, scan folder, du
    Needs: fs.read
    """

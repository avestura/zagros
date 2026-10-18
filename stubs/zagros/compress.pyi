"""compress: gzip, zlib and deflate for bytes, and zip, tar and tar.gz archives of files.

Type description, not runnable Starlark: do not `import` anything. The name `compress` is predeclared in main.star; call `compress.gzip(...)` directly.
Needs: none for the bytes functions; fs.read / fs.write (by path) for create, extract and list.
Keywords: zip, unzip, decompress, archive, tar, gzip, untar, extract archive, compress files, deflate, zlib, bzip2

Archive entries that are absolute, contain `..` or a drive letter are refused
(zip-slip), links and special files are never extracted, and totals are capped:
10000 entries and 256 MiB extracted. Decompressing to memory is capped at 4 MiB
(or a lower max_bytes). zstd is not available. Functions raise on failure.
"""

from typing import List

class ArchiveEntry:
    name: str
    size: int
    is_dir: bool
    mtime: int

def gzip(data: bytes, level: int = -1) -> bytes:
    """Compress to gzip. level 0-9, -1 default, -2 Huffman only.

    Example: packed = compress.gzip(b"hello hello hello")
    Task: gzip some data
    Keywords: compress, gz, deflate, shrink, compress bytes, make smaller
    Needs: none
    """

def gunzip(data: bytes, max_bytes: int = 0) -> bytes:
    """Decompress gzip; fails when the result exceeds the cap.

    Example: raw = compress.gunzip(compress.gzip(b"hello"))
    Task: decompress gzip data
    Keywords: ungzip, decompress, gz, inflate, unzip bytes, decompress response
    Needs: none
    """

def zlib(data: bytes, level: int = -1) -> bytes:
    """Compress to the zlib format.

    Example: packed = compress.zlib(b"hello hello hello")
    Task: zlib-compress data
    Keywords: compress, zlib, deflate stream, shrink
    Needs: none
    """

def unzlib(data: bytes, max_bytes: int = 0) -> bytes:
    """Decompress zlib data.

    Example: raw = compress.unzlib(compress.zlib(b"hello"))
    Task: decompress zlib data
    Keywords: decompress, zlib, inflate
    Needs: none
    """

def deflate(data: bytes, level: int = -1) -> bytes:
    """Compress to a raw deflate stream (no header).

    Example: packed = compress.deflate(b"hello hello hello")
    Task: deflate data
    Keywords: compress, raw deflate, shrink
    Needs: none
    """

def inflate(data: bytes, max_bytes: int = 0) -> bytes:
    """Decompress a raw deflate stream.

    Example: raw = compress.inflate(compress.deflate(b"hello"))
    Task: inflate raw deflate data
    Keywords: decompress, raw deflate, inflate
    Needs: none
    """

def bunzip2(data: bytes, max_bytes: int = 0) -> bytes:
    """Decompress bzip2 data (decompression only).

    Example:
        blob = fs.read_bytes("~/Documents/data.bz2")
        raw = compress.bunzip2(blob)
    Task: decompress bzip2 data
    Keywords: bz2, bzip2, decompress, unbzip2
    Needs: none
    """

def create(dest: str, sources: List[str], format: str = "") -> int:
    """Create an archive from files and folders (folders are added recursively, links skipped); returns the number of entries. format is "zip", "tar" or "tar.gz", taken from the extension of dest when empty. Entries are named after each source's last path element.

    Example: n = compress.create("~/Documents/backup.zip", ["~/Documents/notes", "~/Documents/todo.txt"])
    Task: zip files or folders
    Keywords: zip, make archive, tar, tar.gz, tgz, compress folder, backup, pack files, archive files
    Needs: fs.read for sources, fs.write for dest
    """

def extract(archive: str, dest: str, format: str = "", overwrite: bool = False) -> List[str]:
    """Extract an archive into the folder dest (created when missing); returns the extracted file names. Existing files are an error unless overwrite=True. The format comes from the extension or the file header.

    Example: files = compress.extract("~/Documents/backup.zip", "~/Documents/restored")
    Task: unzip an archive
    Keywords: unzip, untar, extract archive, decompress zip, open tar.gz, unpack, extract files
    Needs: fs.read for archive, fs.write for dest
    """

def list(archive: str, format: str = "") -> List[ArchiveEntry]:
    """List the entries of an archive without extracting.

    Example: names = [e.name for e in compress.list("~/Documents/backup.zip")]
    Task: see what is inside an archive
    Keywords: list zip, archive contents, tar -t, show files in archive, inspect archive
    Needs: fs.read
    """

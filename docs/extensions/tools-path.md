# File paths (`path`)

Join, split, clean and relate file paths in Windows or POSIX flavour, without touching the disk.

No permission is needed: these are pure functions. Generated from the tools registry.

13 functions. Use them as `path.<name>(...)` in any extension.

## path.base

`path.base(path: str, flavor: str = "posix") -> str`

Return the last element: the file or folder name. Trailing separators are ignored; an empty path gives a dot and a root gives the separator.

| Parameter | Type | Description |
|---|---|---|
| `path` | `str` | The path. |
| `flavor` | `str` (default `"posix"`) | Path rules: posix (/ separator), windows (\ or /, drive letters and UNC shares) or auto (windows when the path has a backslash or a drive letter). One of: posix, windows, auto. |

Returns `str`. Example: `path.base("/usr/local/bin/tool.exe")`

Keywords: filename, basename, file name from path, last segment, path, file, base

## path.clean

`path.clean(path: str, flavor: str = "posix") -> str`

Return the shortest equivalent path: repeated separators merged, . removed, .. resolved against the part before it (dropped at the root, kept at the start of a relative path). An empty path becomes a dot.

| Parameter | Type | Description |
|---|---|---|
| `path` | `str` | The path. |
| `flavor` | `str` (default `"posix"`) | Path rules: posix (/ separator), windows (\ or /, drive letters and UNC shares) or auto (windows when the path has a backslash or a drive letter). One of: posix, windows, auto. |

Returns `str`. Example: `path.clean("/a/./b/../c//d/")`

Keywords: normalize path, resolve dots, remove .., simplify path, canonical path, path, filename, file, clean

## path.convert

`path.convert(path: str, to: str = "posix") -> str`

Rewrite the separators of a path for the target flavor: / to \ for windows, \ to / for posix. Nothing else changes, so drive letters stay. Use path.clean afterwards to normalize.

| Parameter | Type | Description |
|---|---|---|
| `path` | `str` | The path. |
| `to` | `str` (default `"posix"`) | The flavor to convert to. One of: posix, windows. |

Returns `str`. Example: `path.convert("C:\\Users\\me\\a.txt", to="posix")`

Keywords: windows to unix path, unix to windows path, backslash to slash, slash to backslash, wsl path, convert separators, path, convert, separator, slash, backslash

## path.dir

`path.dir(path: str, flavor: str = "posix") -> str`

Return everything but the last element, cleaned: the containing folder. A bare name gives a dot.

| Parameter | Type | Description |
|---|---|---|
| `path` | `str` | The path. |
| `flavor` | `str` (default `"posix"`) | Path rules: posix (/ separator), windows (\ or /, drive letters and UNC shares) or auto (windows when the path has a backslash or a drive letter). One of: posix, windows, auto. |

Returns `str`. Example: `path.dir("C:\\Users\\me\\a.txt", flavor="windows")`

Keywords: dirname, parent folder, parent directory, containing folder, path, filename, file, dir

## path.ext

`path.ext(path: str, flavor: str = "posix") -> str`

Return the extension of the last element including the dot, for example .gz for a.tar.gz. Files that only start with a dot (.bashrc) and names that end in a dot have none.

| Parameter | Type | Description |
|---|---|---|
| `path` | `str` | The path. |
| `flavor` | `str` (default `"posix"`) | Path rules: posix (/ separator), windows (\ or /, drive letters and UNC shares) or auto (windows when the path has a backslash or a drive letter). One of: posix, windows, auto. |

Returns `str`. Example: `path.ext("archive.tar.gz")`

Keywords: file extension, suffix, file type, extname, path, filename, file, ext

## path.is_abs

`path.is_abs(path: str, flavor: str = "posix") -> bool`

Report whether the path is absolute: a leading / for posix; a drive with a separator (C:\x) or a UNC share for windows. \x and C:x are not absolute on windows.

| Parameter | Type | Description |
|---|---|---|
| `path` | `str` | The path. |
| `flavor` | `str` (default `"posix"`) | Path rules: posix (/ separator), windows (\ or /, drive letters and UNC shares) or auto (windows when the path has a backslash or a drive letter). One of: posix, windows, auto. |

Returns `bool`. Example: `path.is_abs("C:\\Windows", flavor="windows")`

Keywords: absolute or relative, is relative, check absolute path, rooted, path, filename, file, is_abs

## path.is_within

`path.is_within(parent: str, child: str, flavor: str = "posix") -> bool`

Report whether child is the same as or inside parent after both are cleaned, without touching the disk. Useful to reject .. escapes before using a user-supplied name. Both must be absolute or both relative; names compare case-insensitively on windows.

| Parameter | Type | Description |
|---|---|---|
| `parent` | `str` | The folder. |
| `child` | `str` | The path to test. |
| `flavor` | `str` (default `"posix"`) | Path rules: posix (/ separator), windows (\ or /, drive letters and UNC shares) or auto (windows when the path has a backslash or a drive letter). One of: posix, windows, auto. |

Returns `bool`. Example: `path.is_within("/srv/data", "/srv/data/../etc/passwd")`

Keywords: path traversal, directory traversal check, is subpath, inside folder, sandbox check, starts with, path, traversal, within, subpath, security

## path.join

`path.join(parts: List[str], flavor: str = "posix") -> str`

Join path parts with the flavor's separator and clean the result (dot segments resolved, repeated separators merged). Empty parts are skipped and an absolute part does not discard the ones before it. Returns an empty string when every part is empty.

| Parameter | Type | Description |
|---|---|---|
| `parts` | `List[str]` | The parts to join. |
| `flavor` | `str` (default `"posix"`) | Path rules: posix (/ separator), windows (\ or /, drive letters and UNC shares) or auto (windows when the path has a backslash or a drive letter). One of: posix, windows, auto. |

Returns `str`. Example: `path.join(["C:\\Users", "me", "..", "docs", "a.txt"], flavor="windows")`

Keywords: combine paths, append path, path.join, build path, concatenate path, path, join, combine, append

## path.parts

`path.parts(path: str, flavor: str = "posix") -> List[str]`

Return the elements of the path as a list. The root or drive comes first as its own element (/, C:\, \\host\share\); empty and . elements are dropped and .. is kept.

| Parameter | Type | Description |
|---|---|---|
| `path` | `str` | The path. |
| `flavor` | `str` (default `"posix"`) | Path rules: posix (/ separator), windows (\ or /, drive letters and UNC shares) or auto (windows when the path has a backslash or a drive letter). One of: posix, windows, auto. |

Returns `List[str]`. Example: `path.parts("C:\\a\\b\\c.txt", flavor="windows")`

Keywords: split path into segments, path components, directories list, path segments, path, filename, file, parts

## path.rel

`path.rel(base: str, target: str, flavor: str = "posix") -> str`

Return the path that leads from base to target, for example ../c from /a/b to /a/c. Both paths are cleaned first and must both be absolute or both relative; on windows they must share a drive, and names compare case-insensitively. Raises when no relative path exists.

| Parameter | Type | Description |
|---|---|---|
| `base` | `str` | The starting folder. |
| `target` | `str` | The path to reach. |
| `flavor` | `str` (default `"posix"`) | Path rules: posix (/ separator), windows (\ or /, drive letters and UNC shares) or auto (windows when the path has a backslash or a drive letter). One of: posix, windows, auto. |

Returns `str`. Example: `path.rel("/data/app", "/data/lib/x.so")`

Keywords: relative path, relpath, path from to, make relative, diff paths, path, relative

## path.split

`path.split(path: str, flavor: str = "posix") -> List[str]`

Split after the last separator into \[folder, file\]. The folder keeps its trailing separator, so joining the two gives the original path back. A path with no separator gives an empty folder.

| Parameter | Type | Description |
|---|---|---|
| `path` | `str` | The path. |
| `flavor` | `str` (default `"posix"`) | Path rules: posix (/ separator), windows (\ or /, drive letters and UNC shares) or auto (windows when the path has a backslash or a drive letter). One of: posix, windows, auto. |

Returns `List[str]`. Example: `path.split("/a/b/c.txt")`

Keywords: dirname and basename, separate filename, split path, path, filename, file, split

## path.stem

`path.stem(path: str, flavor: str = "posix") -> str`

Return the last element without its extension (see path.ext): archive.tar.gz gives archive.tar.

| Parameter | Type | Description |
|---|---|---|
| `path` | `str` | The path. |
| `flavor` | `str` (default `"posix"`) | Path rules: posix (/ separator), windows (\ or /, drive letters and UNC shares) or auto (windows when the path has a backslash or a drive letter). One of: posix, windows, auto. |

Returns `str`. Example: `path.stem("/tmp/report.final.pdf")`

Keywords: filename without extension, name without ext, strip extension, path, filename, file, stem

## path.with_ext

`path.with_ext(path: str, ext: str, flavor: str = "posix") -> str`

Replace the extension of the last element. ext may be given with or without the dot; an empty ext removes the extension. Raises when ext contains a separator.

| Parameter | Type | Description |
|---|---|---|
| `path` | `str` | The path. |
| `ext` | `str` | The new extension, for example .txt or txt. |
| `flavor` | `str` (default `"posix"`) | Path rules: posix (/ separator), windows (\ or /, drive letters and UNC shares) or auto (windows when the path has a backslash or a drive letter). One of: posix, windows, auto. |

Returns `str`. Example: `path.with_ext("/tmp/a.txt", "md")`

Keywords: rename extension, change extension, replace suffix, swap extension, path, extension, rename, suffix


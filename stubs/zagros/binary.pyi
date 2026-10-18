"""binary: pack and unpack binary data, hex, integers and byte-slice helpers.

Type description, not runnable Starlark: do not `import` anything. The name `binary` is predeclared in main.star; call `binary.pack(...)` directly. (It is not called `bytes` because that is Starlark's own builtin; use that for `bytes("text")`, slicing and `+`.)
Needs: none. Pure functions.
Keywords: bytes, binary, struct, pack, unpack, hex, endian, byte order, integer to bytes, xor, raw data

Format strings follow Python's struct with standard sizes and no padding. An
optional first character picks the byte order: `<` or `=` or `@` little endian
(default), `>` or `!` big endian. Codes: x pad byte, c one byte, b B 8-bit,
? bool, h H 16-bit, i I l L 32-bit, q Q 64-bit, f 32-bit float, d 64-bit float,
Ns a fixed N-byte field (bytes or str, zero padded). A number before a code
repeats it: "<2H4s".
"""

from typing import Any, List, Tuple, Union

Data = Union[bytes, str]

def pack(fmt: str, *values: Any) -> bytes:
    """Pack values into bytes by a format string.

    Example: header = binary.pack(">HI4s", 1, 4096, "ZAGR")
    Task: build a binary record
    Keywords: struct.pack, serialize binary, build packet, little endian, big endian, encode integers
    Needs: none
    """

def unpack(fmt: str, data: Data) -> Tuple[Any, ...]:
    """Unpack bytes by a format string; the length must match exactly (see calcsize).

    Example: version, size = binary.unpack("<HI", b"\\x01\\x00\\x00\\x10\\x00\\x00")
    Task: parse a binary record
    Keywords: struct.unpack, parse binary, decode packet, read header, read integers from bytes
    Needs: none
    """

def unpack_from(fmt: str, data: Data, offset: int = 0) -> Tuple[Any, ...]:
    """Like unpack but starts at offset and ignores trailing bytes.

    Example: (magic,) = binary.unpack_from("4s", b"ZAGR....", 0)
    Task: parse a field inside a larger buffer
    Keywords: struct.unpack_from, read at offset, parse header of file
    Needs: none
    """

def calcsize(fmt: str) -> int:
    """Number of bytes a format string packs to.

    Example: n = binary.calcsize("<HI4s")
    Keywords: struct.calcsize, record size, format length
    Needs: none
    """

def hex(data: Data, sep: str = "") -> str:
    """Lower-case hex of bytes, optionally with a separator between bytes.

    Example: s = binary.hex(b"\\x01\\xab", sep = ":")
    Task: show bytes as hex
    Keywords: hexdump, bytes to hex, to hex, hex string, mac address format
    Needs: none
    """

def unhex(text: str) -> bytes:
    """Bytes from hex; spaces, colons, dashes and a 0x prefix are ignored.

    Example: data = binary.unhex("01 ab ff")
    Task: parse a hex string
    Keywords: hex to bytes, from hex, decode hex, parse hex dump
    Needs: none
    """

def to_int(data: Data, byteorder: str = "big", signed: bool = False) -> int:
    """Read bytes as one integer (up to 4096 bytes). byteorder is "big" or "little".

    Example: n = binary.to_int(b"\\x01\\x00", "little")
    Task: convert bytes to an integer
    Keywords: bytes to int, from bytes, endian, big integer, read number
    Needs: none
    """

def from_int(value: int, size: int, byteorder: str = "big", signed: bool = False) -> bytes:
    """An integer as exactly size bytes; raises when it does not fit.

    Example: raw = binary.from_int(258, 2)
    Task: convert an integer to bytes
    Keywords: int to bytes, to bytes, endian, serialize number, big integer
    Needs: none
    """

def xor(data: Data, key: Data) -> bytes:
    """XOR data with a repeating key.

    Example: masked = binary.xor(b"hello", b"\\x5a")
    Task: xor bytes
    Keywords: xor cipher, mask bytes, obfuscate, bitwise xor
    Needs: none
    """

def chunks(data: Data, size: int) -> List[bytes]:
    """Split into pieces of size bytes (the last may be shorter).

    Example: parts = binary.chunks(b"abcdefg", 3)
    Task: split bytes into fixed-size blocks
    Keywords: split bytes, slice bytes, blocks, batch, partition
    Needs: none
    """

def reverse(data: Data) -> bytes:
    """Bytes in reverse order, for example to swap endianness of one field.

    Example: swapped = binary.reverse(b"\\x01\\x02\\x03\\x04")
    Task: reverse bytes
    Keywords: swap endian, byte swap, flip bytes
    Needs: none
    """

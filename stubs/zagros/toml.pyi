"""TOML: Parse TOML into dicts and lists, and write values back as TOML.

Generated from the tools registry; do not edit by hand.
Type description, not runnable Starlark: do not `import` anything. The name `toml` is predeclared in main.star; call `toml.<name>(...)` directly.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
Keywords: toml, parse toml, read toml, write toml, config file, cargo, pyproject, serialize
"""

from typing import Any, Dict, List

def dump(value: Any) -> str:
    """Write a dict as TOML text: plain values first, then [tables] and [[arrays of tables]], keeping key order. None has no TOML form and raises, as does a top-level value that is not a dict, a mixed list of tables and values, or nesting deeper than 100 levels. Strings stay strings, so a parsed date is written back as a quoted string.

        value: A dict to write.

    Example: toml.dump({"name": "zagros", "db": {"host": "localhost", "port": 5432}})
    Keywords: dict to toml, json to toml, write toml, save toml, serialize toml, toml, dump, write, serialize
    Needs: none (pure function).
    """

def is_valid(text: str) -> bool:
    """Report whether text is valid TOML. Use toml.parse to get the error message.

        text: The TOML text.

    Example: toml.is_valid("a = 1")
    Keywords: validate toml, check toml, lint toml, toml, validate, check
    Needs: none (pure function).
    """

def parse(text: str) -> Any:
    """Parse TOML text into a dict of dicts, lists, strings, ints, floats and bools. Keys keep their document order. Dates and times become RFC 3339 strings (local dates and times keep their own form). Raises with the line and column on invalid TOML or duplicate keys.

        text: The TOML text.

    Example: toml.parse("name = \"zagros\"\nports = [80, 443]\n[db]\nhost = \"localhost\"")
    Keywords: toml to dict, toml to json, read toml, load toml, config file, cargo.toml, pyproject, deserialize toml, toml, parse, load, config
    Needs: none (pure function).
    """

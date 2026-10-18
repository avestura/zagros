# TOML (`toml`)

Parse TOML into dicts and lists, and write values back as TOML.

No permission is needed: these are pure functions. Generated from the tools registry.

3 functions. Use them as `toml.<name>(...)` in any extension.

## toml.dump

`toml.dump(value: Any) -> str`

Write a dict as TOML text: plain values first, then \[tables\] and \[\[arrays of tables\]\], keeping key order. None has no TOML form and raises, as does a top-level value that is not a dict, a mixed list of tables and values, or nesting deeper than 100 levels. Strings stay strings, so a parsed date is written back as a quoted string.

| Parameter | Type | Description |
|---|---|---|
| `value` | `Any` | A dict to write. |

Returns `str`. Example: `toml.dump({"name": "zagros", "db": {"host": "localhost", "port": 5432}})`

Keywords: dict to toml, json to toml, write toml, save toml, serialize toml, toml, dump, write, serialize

## toml.is_valid

`toml.is_valid(text: str) -> bool`

Report whether text is valid TOML. Use toml.parse to get the error message.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | The TOML text. |

Returns `bool`. Example: `toml.is_valid("a = 1")`

Keywords: validate toml, check toml, lint toml, toml, validate, check

## toml.parse

`toml.parse(text: str) -> Any`

Parse TOML text into a dict of dicts, lists, strings, ints, floats and bools. Keys keep their document order. Dates and times become RFC 3339 strings (local dates and times keep their own form). Raises with the line and column on invalid TOML or duplicate keys.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | The TOML text. |

Returns `Any`. Example: `toml.parse("name = \"zagros\"\nports = [80, 443]\n[db]\nhost = \"localhost\"")`

Keywords: toml to dict, toml to json, read toml, load toml, config file, cargo.toml, pyproject, deserialize toml, toml, parse, load, config


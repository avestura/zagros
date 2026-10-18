# YAML (`yaml`)

Parse YAML into dicts and lists, and write values back as YAML.

No permission is needed: these are pure functions. Generated from the tools registry.

4 functions. Use them as `yaml.<name>(...)` in any extension.

## yaml.dump

`yaml.dump(value: Any, indent: int = 2) -> str`

Write a value (dict, list, string, number, bool, None) as YAML text with the given indentation. Dict keys keep their order. Strings that would read back as another type, or that contain special characters, are quoted; multi-line strings use the literal | style. Raises when the value is nested deeper than 100 levels.

| Parameter | Type | Description |
|---|---|---|
| `value` | `Any` | The value to write. |
| `indent` | `int` (default `2`) | Spaces per level, 1 to 8. |

Returns `str`. Example: `yaml.dump({"name": "zagros", "tags": ["a", "b"]})`

Keywords: dict to yaml, json to yaml, write yaml, save yaml, serialize yaml, yml output, yaml, yml, dump, write, serialize

## yaml.is_valid

`yaml.is_valid(text: str) -> bool`

Report whether text parses as YAML (every document). Use yaml.parse to get the error message.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | The YAML text. |

Returns `bool`. Example: `yaml.is_valid("a: [1, 2")`

Keywords: validate yaml, check yaml, lint yaml, yaml, validate, check

## yaml.parse

`yaml.parse(text: str) -> Any`

Parse the first YAML document in text into dicts, lists, strings, ints, floats, bools and None. Supports block and flow collections, quoted and block scalars, comments, anchors with aliases, merge keys and the core scalar types (true/false are bools; yes/no stay strings). Empty text gives None. Raises with a line number on malformed YAML or a duplicate key.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | The YAML text. |

Returns `Any`. Example: `yaml.parse("name: zagros\ntags: [a, b]\nport: 8080")`

Keywords: yaml to dict, yaml to json, read yaml, load yaml, config file, yml, deserialize yaml, yaml, parse, load, config

## yaml.parse_all

`yaml.parse_all(text: str) -> Any`

Parse every document of a multi-document YAML stream (separated by ---) and return them as a list. Raises with a line number on malformed YAML.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | The YAML text. |

Returns `Any`. Example: `yaml.parse_all("a: 1\n---\nb: 2")`

Keywords: multi document yaml, yaml stream, kubernetes manifests, split yaml documents, yaml, documents, stream, parse


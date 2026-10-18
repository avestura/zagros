"""YAML: Parse YAML into dicts and lists, and write values back as YAML.

Generated from the tools registry; do not edit by hand.
Type description, not runnable Starlark: do not `import` anything. The name `yaml` is predeclared in main.star; call `yaml.<name>(...)` directly.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
Keywords: yaml, yml, parse yaml, read yaml, write yaml, config file, kubernetes manifest, docker compose, serialize
"""

from typing import Any, Dict, List

def dump(value: Any, indent: int = 2) -> str:
    """Write a value (dict, list, string, number, bool, None) as YAML text with the given indentation. Dict keys keep their order. Strings that would read back as another type, or that contain special characters, are quoted; multi-line strings use the literal | style. Raises when the value is nested deeper than 100 levels.

        value: The value to write.
        indent: Spaces per level, 1 to 8.

    Example: yaml.dump({"name": "zagros", "tags": ["a", "b"]})
    Keywords: dict to yaml, json to yaml, write yaml, save yaml, serialize yaml, yml output, yaml, yml, dump, write, serialize
    Needs: none (pure function).
    """

def is_valid(text: str) -> bool:
    """Report whether text parses as YAML (every document). Use yaml.parse to get the error message.

        text: The YAML text.

    Example: yaml.is_valid("a: [1, 2")
    Keywords: validate yaml, check yaml, lint yaml, yaml, validate, check
    Needs: none (pure function).
    """

def parse(text: str) -> Any:
    """Parse the first YAML document in text into dicts, lists, strings, ints, floats, bools and None. Supports block and flow collections, quoted and block scalars, comments, anchors with aliases, merge keys and the core scalar types (true/false are bools; yes/no stay strings). Empty text gives None. Raises with a line number on malformed YAML or a duplicate key.

        text: The YAML text.

    Example: yaml.parse("name: zagros\ntags: [a, b]\nport: 8080")
    Keywords: yaml to dict, yaml to json, read yaml, load yaml, config file, yml, deserialize yaml, yaml, parse, load, config
    Needs: none (pure function).
    """

def parse_all(text: str) -> Any:
    """Parse every document of a multi-document YAML stream (separated by ---) and return them as a list. Raises with a line number on malformed YAML.

        text: The YAML text.

    Example: yaml.parse_all("a: 1\n---\nb: 2")
    Keywords: multi document yaml, yaml stream, kubernetes manifests, split yaml documents, yaml, documents, stream, parse
    Needs: none (pure function).
    """

"""json: the standard Starlark JSON module.

Type description, not runnable Starlark: do not `import` anything. The name `json` is predeclared in main.star; call `json.encode(...)` directly.
Needs: none.
Keywords: json, serialize, deserialize, parse json, stringify, pretty print json
"""

from typing import Any

def encode(x: Any) -> str:
    """Encode a value as compact JSON.

    Example: s = json.encode({"a": [1, 2]})
    Task: turn a value into JSON text
    Keywords: stringify, serialize, to json, dump json, dict to json
    Needs: none
    """

def decode(x: str, default: Any = ...) -> Any:
    """Parse JSON text into dicts, lists, strings and numbers; returns default (when given) on bad input.

    Example: v = json.decode("{\"a\": 1}")
    Task: parse JSON text
    Keywords: parse json, deserialize, from json, load json, json to dict
    Needs: none
    """

def encode_indent(x: Any, prefix: str = "", indent: str = "\t") -> str:
    """Encode a value as indented JSON.

    Example: s = json.encode_indent({"a": 1}, indent = "  ")
    Keywords: pretty print json, format json, beautify json, indent json
    Needs: none
    """

def indent(s: str, prefix: str = "", indent: str = "\t") -> str:
    """Re-indent JSON text.

    Example: s = json.indent("{\"a\":1}", indent = "  ")
    Keywords: pretty print json, format json text, beautify json, reformat
    Needs: none
    """

# Text templates (`template`)

Render text from a safe subset of Go text/template and a dict.

No permission is needed: these are pure functions. Generated from the tools registry.

2 functions. Use them as `template.<name>(...)` in any extension.

## template.is_valid

`template.is_valid(template: str, left: str = "{{", right: str = "}}") -> bool`

Report whether a template parses and uses only the allowed features (see template.render). Use template.render to get the error message.

| Parameter | Type | Description |
|---|---|---|
| `template` | `str` | The template text. |
| `left` | `str` (default `"{{"`) | Opening delimiter. |
| `right` | `str` (default `"}}"`) | Closing delimiter. |

Returns `bool`. Example: `template.is_valid("{{ if .a }}x{{ end }}")`

Keywords: validate template, check template syntax, lint template, template, validate, check, syntax

## template.render

`template.render(template: str, data: Any = None, left: str = "{{", right: str = "}}", strict: bool = False) -> str`

Render text from a template and a dict using Go template syntax: {{ .name }}, {{ if .ok }}...{{ else }}...{{ end }}, {{ range .items }}{{ . }}{{ end }}, {{ with .x }}, variables ($x := ...), pipelines ({{ .name | upper }}) and comments. Besides the Go built-ins (len, index, printf, eq, lt, and, or, not, slice, html, js, urlquery) these functions exist: upper, lower, title, trim, trim_prefix, trim_suffix, replace, split, join, contains, has_prefix, has_suffix, repeat, default, coalesce, quote, json, add, sub, mul, div, mod, int, float, string, list, dict, keys. Nothing can reach outside the sandbox, and {{ define }}, {{ template }}, {{ block }} and call are rejected. A missing key renders as empty text, or raises with strict=True. Raises with the position on syntax or runtime errors, and when the output passes 4 MB or rendering takes over 3 seconds.

| Parameter | Type | Description |
|---|---|---|
| `template` | `str` | The template text. |
| `data` | `Any` (default `None`) | The values, usually a dict; refer to keys as .key. |
| `left` | `str` (default `"{{"`) | Opening delimiter. |
| `right` | `str` (default `"}}"`) | Closing delimiter. |
| `strict` | `bool` (default `False`) | Raise on a missing key instead of rendering it as empty. |

Returns `str`. Example: `template.render("Hello {{ .name }}! {{ range .tags }}[{{ . | upper }}]{{ end }}", {"name": "Ada", "tags": ["a", "b"]})`

Keywords: mustache, handlebars, jinja, fill template, string interpolation, mail merge, placeholders, go template, text template, format string with dict, generate text from data, template, render, format, interpolate, placeholder


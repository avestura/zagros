"""Text templates: Render text from a safe subset of Go text/template and a dict.

Generated from the tools registry; do not edit by hand.
Type description, not runnable Starlark: do not `import` anything. The name `template` is predeclared in main.star; call `template.<name>(...)` directly.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
Keywords: template, render, text template, placeholders, mail merge, format string, go template, interpolate
"""

from typing import Any, Dict, List

def is_valid(template: str, left: str = "{{", right: str = "}}") -> bool:
    """Report whether a template parses and uses only the allowed features (see template.render). Use template.render to get the error message.

        template: The template text.
        left: Opening delimiter.
        right: Closing delimiter.

    Example: template.is_valid("{{ if .a }}x{{ end }}")
    Keywords: validate template, check template syntax, lint template, template, validate, check, syntax
    Needs: none (pure function).
    """

def render(template: str, data: Any = None, left: str = "{{", right: str = "}}", strict: bool = False) -> str:
    """Render text from a template and a dict using Go template syntax: {{ .name }}, {{ if .ok }}...{{ else }}...{{ end }}, {{ range .items }}{{ . }}{{ end }}, {{ with .x }}, variables ($x := ...), pipelines ({{ .name | upper }}) and comments. Besides the Go built-ins (len, index, printf, eq, lt, and, or, not, slice, html, js, urlquery) these functions exist: upper, lower, title, trim, trim_prefix, trim_suffix, replace, split, join, contains, has_prefix, has_suffix, repeat, default, coalesce, quote, json, add, sub, mul, div, mod, int, float, string, list, dict, keys. Nothing can reach outside the sandbox, and {{ define }}, {{ template }}, {{ block }} and call are rejected. A missing key renders as empty text, or raises with strict=True. Raises with the position on syntax or runtime errors, and when the output passes 4 MB or rendering takes over 3 seconds.

        template: The template text.
        data: The values, usually a dict; refer to keys as .key.
        left: Opening delimiter.
        right: Closing delimiter.
        strict: Raise on a missing key instead of rendering it as empty.

    Example: template.render("Hello {{ .name }}! {{ range .tags }}[{{ . | upper }}]{{ end }}", {"name": "Ada", "tags": ["a", "b"]})
    Keywords: mustache, handlebars, jinja, fill template, string interpolation, mail merge, placeholders, go template, text template, format string with dict, generate text from data, template, render, format, interpolate, placeholder
    Needs: none (pure function).
    """

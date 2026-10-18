---
status: accepted
date: 2026-10-18
---
# ADR 0020: Structured values and data-format groups in the tools registry

## Context
The registry only knew strings, numbers, bools, lists of strings or floats and flat string maps. YAML, TOML, CSV, XML and HTML parsers, JWT decoding and URL parsing return nested data, so they could only be offered as text-in text-out helpers.

## Decision
- A new registry type `any` stands for a structured value: dict, list, string, int, float, bool or None. On the Go side it is `*tools.Dict` (an insertion-ordered string map), `[]any`, `string`, `int64`, `float64`, `bool` and `nil`. The Starlark bridge converts it both ways, rejects non-string dict keys and nesting beyond 200 levels, and the generated stubs spell it `Any`.
- Nine groups use it: `yaml`, `toml`, `csv`, `xml`, `html`, `url`, `path`, `template`, `jwt`. All are pure and need no permission. `path` never touches the disk, `template` runs a restricted `text/template` (no `define`, `template`, `block` or `call`, large range counts rejected, 4 MB output and 3 second limits), and `jwt` refuses `alg: none` and never lets an HS token be checked against a PEM key.
- Parsers keep key order so a parsed document can be dumped without reshuffling. The YAML parser is hand-written for the configuration subset (no explicit `?` keys, no custom tags) because no YAML library is in `go.mod`; TOML uses BurntSushi/toml and HTML uses `golang.org/x/net/html`, both already dependencies.
- Hashing, HMAC, key derivation and public-key signatures stay in `hash` and `cipher`; `jwt` uses the standard library for the token format only.

## Consequences
Extensions can read and write config and data formats without a permission or a host round trip. The launcher does not list these functions (no `Family`), since their inputs are structured. Any later registry function can take or return `any`, and the check in `toolstest` accepts it.

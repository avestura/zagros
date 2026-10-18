# Regular expressions (`regex`)

Test, match, replace and explain regular expressions (Go RE2 syntax).

No permission is needed: these are pure functions. Generated from the tools registry.

25 functions. Use them as `regex.<name>(...)` in any extension.

## regex.test

`regex.test(pattern: str, text: str, flags: str = "") -> bool`

Reports whether the pattern matches anywhere in the text. Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Regular expression in Go RE2 syntax. |
| `text` | `str` | Text to search. |
| `flags` | `str` (default `""`) | Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im". |

Returns `bool`. Example: `regex.test("^\\d+$", "12345")`

## regex.match

`regex.match(pattern: str, text: str, flags: str = "") -> str`

Returns the first match of the pattern in the text, or an empty string when nothing matches. Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Regular expression in Go RE2 syntax. |
| `text` | `str` | Text to search. |
| `flags` | `str` (default `""`) | Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im". |

Returns `str`. Example: `regex.match("\\d+", "order 66 shipped")`

## regex.find_all

`regex.find_all(pattern: str, text: str, flags: str = "", limit: int = 0) -> List[str]`

Returns every non-overlapping match, in order. limit caps the result (0 means up to 100000). Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Regular expression in Go RE2 syntax. |
| `text` | `str` | Text to search. |
| `flags` | `str` (default `""`) | Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im". |
| `limit` | `int` (default `0`) | Maximum number of matches; 0 means no explicit limit. |

Returns `List[str]`. Example: `regex.find_all("\\d+", "a1 b22 c333")`

## regex.groups

`regex.groups(pattern: str, text: str, flags: str = "") -> Dict[str, str]`

Returns the capture groups of the first match as a dict. Keys are "0" (whole match), "1", "2"... and, for named groups, also the group name. Groups that did not participate are empty strings; no match gives an empty dict. Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Regular expression in Go RE2 syntax. |
| `text` | `str` | Text to search. |
| `flags` | `str` (default `""`) | Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im". |

Returns `Dict[str, str]`. Example: `regex.groups("(?P<y>\\d{4})-(?P<m>\\d\\d)", "on 2026-10-17")`

## regex.all_groups

`regex.all_groups(pattern: str, text: str, flags: str = "", limit: int = 0) -> List[str]`

Returns one JSON object per match, each mapping "0" (whole match), "1", "2"... and group names to the captured text, e.g. {"0":"a=1","1":"a","2":"1"}. Use limit to cap the number of matches. Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Regular expression in Go RE2 syntax. |
| `text` | `str` | Text to search. |
| `flags` | `str` (default `""`) | Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im". |
| `limit` | `int` (default `0`) | Maximum number of matches; 0 means no explicit limit. |

Returns `List[str]`. Example: `regex.all_groups("(\\w+)=(\\d+)", "a=1 b=2")`

## regex.count

`regex.count(pattern: str, text: str, flags: str = "") -> int`

Counts the non-overlapping matches of the pattern in the text. Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Regular expression in Go RE2 syntax. |
| `text` | `str` | Text to search. |
| `flags` | `str` (default `""`) | Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im". |

Returns `int`. Example: `regex.count("a", "banana")`

## regex.positions

`regex.positions(pattern: str, text: str, flags: str = "", limit: int = 0) -> List[str]`

Lists every match as "start-end: match" with zero-based rune offsets (end is exclusive), so offsets are correct for non-ASCII text. Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Regular expression in Go RE2 syntax. |
| `text` | `str` | Text to search. |
| `flags` | `str` (default `""`) | Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im". |
| `limit` | `int` (default `0`) | Maximum number of matches; 0 means no explicit limit. |

Returns `List[str]`. Example: `regex.positions("an", "banana")`

## regex.replace

`regex.replace(pattern: str, text: str, replacement: str, flags: str = "", limit: int = 0) -> str`

Replaces matches with a template. In the replacement, $1, ${1} and ${name} insert capture groups, $0 is the whole match and $$ is a literal dollar. limit replaces only the first N matches (0 means all). Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Regular expression in Go RE2 syntax. |
| `text` | `str` | Text to search. |
| `replacement` | `str` | Replacement template with $1 / ${name} references. |
| `flags` | `str` (default `""`) | Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im". |
| `limit` | `int` (default `0`) | Replace at most this many matches; 0 means all. |

Returns `str`. Example: `regex.replace("(\\w+)@(\\w+)", "bob@home", "$2:$1")`

## regex.replace_literal

`regex.replace_literal(pattern: str, text: str, replacement: str, flags: str = "", limit: int = 0) -> str`

Replaces matches with the replacement text taken literally, so $ and backslashes have no special meaning. limit replaces only the first N matches (0 means all). Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Regular expression in Go RE2 syntax. |
| `text` | `str` | Text to search. |
| `replacement` | `str` | Literal replacement text. |
| `flags` | `str` (default `""`) | Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im". |
| `limit` | `int` (default `0`) | Replace at most this many matches; 0 means all. |

Returns `str`. Example: `regex.replace_literal("\\d+", "a1b22", "$#")`

## regex.split

`regex.split(pattern: str, text: str, limit: int = 0, flags: str = "") -> List[str]`

Splits the text around matches of the pattern. limit is the maximum number of pieces (0 means no limit); an empty text gives one empty piece. Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Regular expression in Go RE2 syntax. |
| `text` | `str` | Text to search. |
| `limit` | `int` (default `0`) | Maximum number of pieces; 0 means no limit. |
| `flags` | `str` (default `""`) | Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im". |

Returns `List[str]`. Example: `regex.split("[,;]\\s*", "a, b;c")`

## regex.escape

`regex.escape(text: str) -> str`

Escapes every regex metacharacter so the text matches itself literally (Go regexp.QuoteMeta).

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Literal text to escape. |

Returns `str`. Example: `regex.escape("1+1=2 (really?)")`

## regex.is_valid

`regex.is_valid(pattern: str) -> bool`

Reports whether the pattern compiles under Go RE2 syntax. Patterns using lookarounds or backreferences are reported as invalid.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Pattern to check. |

Returns `bool`. Example: `regex.is_valid("(a|b")`

## regex.validate

`regex.validate(pattern: str) -> str`

Compiles the pattern and returns "ok" or the compiler error message (for example "missing closing ): ..."). Lookarounds and backreferences are reported as unsupported syntax.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Pattern to check. |

Returns `str`. Example: `regex.validate("(?<=a)b")`

## regex.explain

`regex.explain(pattern: str) -> List[str]`

Explains a pattern token by token in plain English: anchors, classes, escapes, quantifiers, groups and alternation. Indentation shows group nesting. Covers common constructs, not every RE2 corner.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Pattern to explain. |

Returns `List[str]`. Example: `regex.explain("^(\\d{3})-[a-z]+$")`

## regex.lines_matching

`regex.lines_matching(pattern: str, text: str, invert: bool = False, flags: str = "") -> List[str]`

Returns the lines of the text where the pattern matches (grep). With invert, returns the lines that do not match (grep -v). Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Regular expression in Go RE2 syntax. |
| `text` | `str` | Text to search. |
| `invert` | `bool` (default `False`) | Return the non-matching lines instead. |
| `flags` | `str` (default `""`) | Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im". |

Returns `List[str]`. Example: `regex.lines_matching("^ERROR", "ok\nERROR x\nok")`

## regex.extract_between

`regex.extract_between(text: str, start: str, end: str) -> List[str]`

Returns every piece of text found between the literal start and end delimiters (not regexes), non-overlapping and shortest-first. Nothing is returned when a delimiter is empty or missing.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `start` | `str` | Opening delimiter (literal). |
| `end` | `str` | Closing delimiter (literal). |

Returns `List[str]`. Example: `regex.extract_between("a[1] b[22]", "[", "]")`

## regex.named_group_names

`regex.named_group_names(pattern: str) -> List[str]`

Lists the names of the named capture groups in the pattern, in order. Unnamed groups are not listed.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Pattern to inspect. |

Returns `List[str]`. Example: `regex.named_group_names("(?P<y>\\d+)-(?P<m>\\d+)")`

## regex.literal_prefix

`regex.literal_prefix(pattern: str) -> str`

Returns the literal text every match must start with (empty when there is none). A fully literal pattern returns itself.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Pattern to inspect. |

Returns `str`. Example: `regex.literal_prefix("abc\\d+")`

## regex.from_glob

`regex.from_glob(glob: str) -> str`

Converts a shell glob into an anchored regex: * becomes .* (any characters), ? becomes . (one character), \[abc\] classes are kept, {a,b} becomes (a|b); everything else is escaped.

| Parameter | Type | Description |
|---|---|---|
| `glob` | `str` | Glob such as *.txt or file?.[ch]. |

Returns `str`. Example: `regex.from_glob("*.{jpg,png}")`

## regex.to_glob

`regex.to_glob(pattern: str) -> str`

Converts a simple regex such as ^a.*\.txt$ back into a glob (.* becomes *, . becomes ?, escapes are removed). Fails for regexes using groups, alternation, classes with quantifiers or other constructs a glob cannot express.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Simple regex made of literals, . and .* |

Returns `str`. Example: `regex.to_glob("^img_.*\\.png$")`

## regex.wildcard_match

`regex.wildcard_match(pattern: str, text: str, ignore_case: bool = False) -> bool`

Matches text against a wildcard pattern where * is any run of characters and ? is exactly one character; the whole text must match. Case-sensitive unless ignore_case is true.

| Parameter | Type | Description |
|---|---|---|
| `pattern` | `str` | Wildcard pattern using * and ?. |
| `text` | `str` | Text to search. |
| `ignore_case` | `bool` (default `False`) | Compare case-insensitively. |

Returns `bool`. Example: `regex.wildcard_match("*.txt", "notes.txt")`

## regex.sed_like

`regex.sed_like(text: str, expression: str) -> str`

Applies sed-style substitution expressions such as s/foo/bar/g to the text. One expression per line; any delimiter works (s|a|b|). Flags: g (all), i (ignore case), m, s, or a number N to replace only the Nth match. The replacement can use \1..\9 and & (whole match); use \& for a literal ampersand. Uses RE2, so no lookarounds.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `expression` | `str` | Substitution such as "s/a/b/g" (several allowed, one per line). |

Returns `str`. Example: `regex.sed_like("hello world", "s/o/0/g")`

## regex.presets

`regex.presets(name: str = "") -> Dict[str, str]`

Returns all built-in named patterns (email, url, ipv4, uuid, hex color, semver, jwt and more) as a name to pattern dict. The patterns are unanchored; wrap with ^...$ for a full match.

| Parameter | Type | Description |
|---|---|---|
| `name` | `str` (default `""`) | Ignored; present so the launcher can offer the function. |

Returns `Dict[str, str]`. Example: `regex.presets()`

## regex.preset

`regex.preset(name: str) -> str`

Returns the regex of a built-in preset by name (see regex.presets), ignoring case, spaces and dashes. Fails with the list of known names when the name is unknown.

| Parameter | Type | Description |
|---|---|---|
| `name` | `str` | Preset name such as email, ipv4, uuid or hex_color. |

Returns `str`. Example: `regex.preset("email")`

## regex.test_preset

`regex.test_preset(name: str, text: str, full: bool = False) -> bool`

Tests the text against a built-in preset (see regex.presets). By default it reports whether the preset occurs anywhere in the text; with full true the whole text must match.

| Parameter | Type | Description |
|---|---|---|
| `name` | `str` | Preset name such as email, ipv4, uuid. |
| `text` | `str` | Text to search. |
| `full` | `bool` (default `False`) | Require the whole text to match instead of a substring. |

Returns `bool`. Example: `regex.test_preset("uuid", "123e4567-e89b-12d3-a456-426614174000", True)`


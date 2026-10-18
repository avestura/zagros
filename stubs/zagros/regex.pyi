"""Regular expressions: Test, match, replace and explain regular expressions (Go RE2 syntax).

Generated from the tools registry; do not edit by hand.
Type description, not runnable Starlark: do not `import` anything. The name `regex` is predeclared in main.star; call `regex.<name>(...)` directly.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
Keywords: regex, regular expression, pattern, match, find all, replace, split, capture group, validate pattern, glob, wildcard, explain regex, sed
"""

from typing import Any, Dict, List

def test(pattern: str, text: str, flags: str = "") -> bool:
    """Reports whether the pattern matches anywhere in the text. Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

        pattern: Regular expression in Go RE2 syntax.
        text: Text to search.
        flags: Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im".

    Example: regex.test("^\\d+$", "12345")
    Task: test a regex against text
    Keywords: does it match, regexp, pattern check, boolean match, match, matches, is match, check, boolean
    Needs: none (pure function).
    """

def match(pattern: str, text: str, flags: str = "") -> str:
    """Returns the first match of the pattern in the text, or an empty string when nothing matches. Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

        pattern: Regular expression in Go RE2 syntax.
        text: Text to search.
        flags: Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im".

    Example: regex.match("\\d+", "order 66 shipped")
    Task: get the first regex match
    Keywords: first match, capture groups, find, first, search, capture
    Needs: none (pure function).
    """

def find_all(pattern: str, text: str, flags: str = "", limit: int = 0) -> List[str]:
    """Returns every non-overlapping match, in order. limit caps the result (0 means up to 100000). Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

        pattern: Regular expression in Go RE2 syntax.
        text: Text to search.
        flags: Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im".
        limit: Maximum number of matches; 0 means no explicit limit.

    Example: regex.find_all("\\d+", "a1 b22 c333")
    Task: find all regex matches
    Keywords: every match, global match, scan text, findall, matches, global, every, list
    Needs: none (pure function).
    """

def groups(pattern: str, text: str, flags: str = "") -> Dict[str, str]:
    """Returns the capture groups of the first match as a dict. Keys are "0" (whole match), "1", "2"... and, for named groups, also the group name. Groups that did not participate are empty strings; no match gives an empty dict. Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

        pattern: Regular expression in Go RE2 syntax.
        text: Text to search.
        flags: Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im".

    Example: regex.groups("(?P<y>\\d{4})-(?P<m>\\d\\d)", "on 2026-10-17")
    Keywords: named groups, capture group values, capture, named group, submatch, parts
    Needs: none (pure function).
    """

def all_groups(pattern: str, text: str, flags: str = "", limit: int = 0) -> List[str]:
    """Returns one JSON object per match, each mapping "0" (whole match), "1", "2"... and group names to the captured text, e.g. {"0":"a=1","1":"a","2":"1"}. Use limit to cap the number of matches. Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

        pattern: Regular expression in Go RE2 syntax.
        text: Text to search.
        flags: Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im".
        limit: Maximum number of matches; 0 means no explicit limit.

    Example: regex.all_groups("(\\w+)=(\\d+)", "a=1 b=2")
    Keywords: captures, findall groups, submatches, json
    Needs: none (pure function).
    """

def count(pattern: str, text: str, flags: str = "") -> int:
    """Counts the non-overlapping matches of the pattern in the text. Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

        pattern: Regular expression in Go RE2 syntax.
        text: Text to search.
        flags: Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im".

    Example: regex.count("a", "banana")
    Keywords: number of matches, occurrences, how many
    Needs: none (pure function).
    """

def positions(pattern: str, text: str, flags: str = "", limit: int = 0) -> List[str]:
    """Lists every match as "start-end: match" with zero-based rune offsets (end is exclusive), so offsets are correct for non-ASCII text. Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

        pattern: Regular expression in Go RE2 syntax.
        text: Text to search.
        flags: Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im".
        limit: Maximum number of matches; 0 means no explicit limit.

    Example: regex.positions("an", "banana")
    Keywords: index, offset, where, span, location
    Needs: none (pure function).
    """

def replace(pattern: str, text: str, replacement: str, flags: str = "", limit: int = 0) -> str:
    """Replaces matches with a template. In the replacement, $1, ${1} and ${name} insert capture groups, $0 is the whole match and $$ is a literal dollar. limit replaces only the first N matches (0 means all). Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

        pattern: Regular expression in Go RE2 syntax.
        text: Text to search.
        replacement: Replacement template with $1 / ${name} references.
        flags: Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im".
        limit: Replace at most this many matches; 0 means all.

    Example: regex.replace("(\\w+)@(\\w+)", "bob@home", "$2:$1")
    Task: replace with a regex
    Keywords: substitute, sed s///, rewrite text, regex substitution, sub, rewrite, template, backreference
    Needs: none (pure function).
    """

def replace_literal(pattern: str, text: str, replacement: str, flags: str = "", limit: int = 0) -> str:
    """Replaces matches with the replacement text taken literally, so $ and backslashes have no special meaning. limit replaces only the first N matches (0 means all). Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

        pattern: Regular expression in Go RE2 syntax.
        text: Text to search.
        replacement: Literal replacement text.
        flags: Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im".
        limit: Replace at most this many matches; 0 means all.

    Example: regex.replace_literal("\\d+", "a1b22", "$#")
    Keywords: substitute, raw replacement, no template
    Needs: none (pure function).
    """

def split(pattern: str, text: str, limit: int = 0, flags: str = "") -> List[str]:
    """Splits the text around matches of the pattern. limit is the maximum number of pieces (0 means no limit); an empty text gives one empty piece. Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

        pattern: Regular expression in Go RE2 syntax.
        text: Text to search.
        limit: Maximum number of pieces; 0 means no limit.
        flags: Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im".

    Example: regex.split("[,;]\\s*", "a, b;c")
    Task: split text with a regex
    Keywords: split by pattern, tokenize, divide, separate, cut
    Needs: none (pure function).
    """

def escape(text: str) -> str:
    """Escapes every regex metacharacter so the text matches itself literally (Go regexp.QuoteMeta).

        text: Literal text to escape.

    Example: regex.escape("1+1=2 (really?)")
    Task: escape text for a regex
    Keywords: literal pattern, quote regex special characters, quote, quotemeta, literal, sanitize
    Needs: none (pure function).
    """

def is_valid(pattern: str) -> bool:
    """Reports whether the pattern compiles under Go RE2 syntax. Patterns using lookarounds or backreferences are reported as invalid.

        pattern: Pattern to check.

    Example: regex.is_valid("(a|b")
    Task: check a regex compiles
    Keywords: validate pattern, syntax check regexp, valid, syntax check, compile, lint
    Needs: none (pure function).
    """

def validate(pattern: str) -> str:
    """Compiles the pattern and returns "ok" or the compiler error message (for example "missing closing ): ..."). Lookarounds and backreferences are reported as unsupported syntax.

        pattern: Pattern to check.

    Example: regex.validate("(?<=a)b")
    Keywords: error, syntax, diagnose, compile
    Needs: none (pure function).
    """

def explain(pattern: str) -> List[str]:
    """Explains a pattern token by token in plain English: anchors, classes, escapes, quantifiers, groups and alternation. Indentation shows group nesting. Covers common constructs, not every RE2 corner.

        pattern: Pattern to explain.

    Example: regex.explain("^(\\d{3})-[a-z]+$")
    Task: explain a regex in words
    Keywords: what does this regex do, regex breakdown, describe pattern, describe, understand, breakdown, plain english, tokens
    Needs: none (pure function).
    """

def lines_matching(pattern: str, text: str, invert: bool = False, flags: str = "") -> List[str]:
    """Returns the lines of the text where the pattern matches (grep). With invert, returns the lines that do not match (grep -v). Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported.

        pattern: Regular expression in Go RE2 syntax.
        text: Text to search.
        invert: Return the non-matching lines instead.
        flags: Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im".

    Example: regex.lines_matching("^ERROR", "ok\nERROR x\nok")
    Keywords: grep, filter lines, findstr, select lines
    Needs: none (pure function).
    """

def extract_between(text: str, start: str, end: str) -> List[str]:
    """Returns every piece of text found between the literal start and end delimiters (not regexes), non-overlapping and shortest-first. Nothing is returned when a delimiter is empty or missing.

        text: Text to search.
        start: Opening delimiter (literal).
        end: Closing delimiter (literal).

    Example: regex.extract_between("a[1] b[22]", "[", "]")
    Keywords: text between two markers, substring between delimiters, between, substring, delimiters, brackets, inside
    Needs: none (pure function).
    """

def named_group_names(pattern: str) -> List[str]:
    """Lists the names of the named capture groups in the pattern, in order. Unnamed groups are not listed.

        pattern: Pattern to inspect.

    Example: regex.named_group_names("(?P<y>\\d+)-(?P<m>\\d+)")
    Keywords: groups, capture names, fields
    Needs: none (pure function).
    """

def literal_prefix(pattern: str) -> str:
    """Returns the literal text every match must start with (empty when there is none). A fully literal pattern returns itself.

        pattern: Pattern to inspect.

    Example: regex.literal_prefix("abc\\d+")
    Keywords: prefix, optimize, fixed text, starts with
    Needs: none (pure function).
    """

def from_glob(glob: str) -> str:
    """Converts a shell glob into an anchored regex: * becomes .* (any characters), ? becomes . (one character), [abc] classes are kept, {a,b} becomes (a|b); everything else is escaped.

        glob: Glob such as *.txt or file?.[ch].

    Example: regex.from_glob("*.{jpg,png}")
    Task: turn a glob into a regex
    Keywords: wildcard to regex, file pattern to regexp, glob, shell pattern, fnmatch
    Needs: none (pure function).
    """

def to_glob(pattern: str) -> str:
    """Converts a simple regex such as ^a.*\.txt$ back into a glob (.* becomes *, . becomes ?, escapes are removed). Fails for regexes using groups, alternation, classes with quantifiers or other constructs a glob cannot express.

        pattern: Simple regex made of literals, . and .*

    Example: regex.to_glob("^img_.*\\.png$")
    Keywords: regex to wildcard, reverse glob
    Needs: none (pure function).
    """

def wildcard_match(pattern: str, text: str, ignore_case: bool = False) -> bool:
    """Matches text against a wildcard pattern where * is any run of characters and ? is exactly one character; the whole text must match. Case-sensitive unless ignore_case is true.

        pattern: Wildcard pattern using * and ?.
        text: Text to search.
        ignore_case: Compare case-insensitively.

    Example: regex.wildcard_match("*.txt", "notes.txt")
    Keywords: glob match, like, fnmatch, star, dos wildcard
    Needs: none (pure function).
    """

def sed_like(text: str, expression: str) -> str:
    """Applies sed-style substitution expressions such as s/foo/bar/g to the text. One expression per line; any delimiter works (s|a|b|). Flags: g (all), i (ignore case), m, s, or a number N to replace only the Nth match. The replacement can use \1..\9 and & (whole match); use \& for a literal ampersand. Uses RE2, so no lookarounds.

        text: Text to search.
        expression: Substitution such as "s/a/b/g" (several allowed, one per line).

    Example: regex.sed_like("hello world", "s/o/0/g")
    Keywords: sed, substitute, unix, s///
    Needs: none (pure function).
    """

def presets(name: str = "") -> Dict[str, str]:
    """Returns all built-in named patterns (email, url, ipv4, uuid, hex color, semver, jwt and more) as a name to pattern dict. The patterns are unanchored; wrap with ^...$ for a full match.

        name: Ignored; present so the launcher can offer the function.

    Example: regex.presets()
    Keywords: common patterns, email regex, ready made regex, library, cookbook, list, email, url, uuid
    Needs: none (pure function).
    """

def preset(name: str) -> str:
    """Returns the regex of a built-in preset by name (see regex.presets), ignoring case, spaces and dashes. Fails with the list of known names when the name is unknown.

        name: Preset name such as email, ipv4, uuid or hex_color.

    Example: regex.preset("email")
    Keywords: common pattern, email regex, url regex, uuid regex, lookup
    Needs: none (pure function).
    """

def test_preset(name: str, text: str, full: bool = False) -> bool:
    """Tests the text against a built-in preset (see regex.presets). By default it reports whether the preset occurs anywhere in the text; with full true the whole text must match.

        name: Preset name such as email, ipv4, uuid.
        text: Text to search.
        full: Require the whole text to match instead of a substring.

    Example: regex.test_preset("uuid", "123e4567-e89b-12d3-a456-426614174000", True)
    Keywords: validate, is email, is uuid, is ip, common patterns
    Needs: none (pure function).
    """

// Package regexx registers the "regex" tool group: RE2 pattern testing,
// matching, replacing, explaining and a library of preset patterns.
package regexx

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

const (
	maxText    = 4 << 20
	maxResults = 100000
)

const syntaxNote = " Uses Go RE2 syntax: lookahead, lookbehind and backreferences are not supported."

func req(name, typ, doc string) tools.Param {
	return tools.Param{Name: name, Type: typ, Doc: doc, Required: true}
}

func opt(name, typ, doc string, def any) tools.Param {
	return tools.Param{Name: name, Type: typ, Doc: doc, Default: def}
}

var (
	pPattern = req("pattern", "str", "Regular expression in Go RE2 syntax.")
	pText    = req("text", "str", "Text to search.")
	pFlags   = opt("flags", "str", `Flags: "i" case-insensitive, "m" multiline (^ and $ match at line breaks), "s" dot matches newline, "U" ungreedy. Combine like "im".`, "")
)

func reg(f tools.Func) {
	f.Group = "regex"
	f.Family = "regex"
	if f.Input == "" {
		f.Input = "text"
	}
	if len(f.Keywords) == 0 {
		f.Keywords = []string{"re", "regex"}
	}
	tools.Register(f)
}

func compile(pattern, flags string) (*regexp.Regexp, error) {
	if pattern == "" {
		return nil, fmt.Errorf("empty pattern")
	}
	if len(pattern) > 64<<10 {
		return nil, fmt.Errorf("pattern too long (max 64 KiB)")
	}
	var fl strings.Builder
	for _, c := range flags {
		switch c {
		case 'i', 'm', 's', 'U':
			if !strings.ContainsRune(fl.String(), c) {
				fl.WriteRune(c)
			}
		case ' ', ',', '-', 'g':
			// ignored: "g" is implicit for the *_all functions
		default:
			return nil, fmt.Errorf("unknown flag %q (use i, m, s, U)", string(c))
		}
	}
	if fl.Len() > 0 {
		pattern = "(?" + fl.String() + ")" + pattern
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, cleanErr(err)
	}
	return re, nil
}

func cleanErr(err error) error {
	return fmt.Errorf("%s", strings.TrimPrefix(err.Error(), "error parsing regexp: "))
}

func checkText(t string) error {
	if len(t) > maxText {
		return fmt.Errorf("text too large (max 4 MiB)")
	}
	return nil
}

func prep(a tools.Args) (*regexp.Regexp, string, error) {
	t := a.Str("text")
	if err := checkText(t); err != nil {
		return nil, "", err
	}
	re, err := compile(a.Str("pattern"), a.Str("flags"))
	return re, t, err
}

func clampLimit(n int64) int {
	if n <= 0 || n > maxResults {
		return maxResults
	}
	return int(n)
}

func init() {
	reg(tools.Func{
		Name: "test", Title: "Regex test", Order: 1,
		Doc:     "Reports whether the pattern matches anywhere in the text." + syntaxNote,
		Params:  []tools.Param{pPattern, pText, pFlags},
		Returns: "bool", Example: `regex.test("^\\d+$", "12345")`,
		Tags: []string{"match", "matches", "is match", "check", "boolean", "regexp"},
		Fn: func(a tools.Args) (any, error) {
			re, t, err := prep(a)
			if err != nil {
				return nil, err
			}
			return re.MatchString(t), nil
		},
	})
	reg(tools.Func{
		Name: "match", Title: "First match", Order: 2,
		Doc:     "Returns the first match of the pattern in the text, or an empty string when nothing matches." + syntaxNote,
		Params:  []tools.Param{pPattern, pText, pFlags},
		Returns: "str", Example: `regex.match("\\d+", "order 66 shipped")`,
		Tags: []string{"find", "first", "search", "capture"},
		Fn: func(a tools.Args) (any, error) {
			re, t, err := prep(a)
			if err != nil {
				return nil, err
			}
			return re.FindString(t), nil
		},
	})
	reg(tools.Func{
		Name: "find_all", Title: "All matches", Order: 3,
		Doc:     "Returns every non-overlapping match, in order. limit caps the result (0 means up to 100000)." + syntaxNote,
		Params:  []tools.Param{pPattern, pText, pFlags, opt("limit", "int", "Maximum number of matches; 0 means no explicit limit.", int64(0))},
		Returns: "list[str]", Example: `regex.find_all("\\d+", "a1 b22 c333")`,
		Tags: []string{"findall", "matches", "global", "every", "list"},
		Fn: func(a tools.Args) (any, error) {
			re, t, err := prep(a)
			if err != nil {
				return nil, err
			}
			out := re.FindAllString(t, clampLimit(a.Int("limit")))
			if out == nil {
				out = []string{}
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "groups", Title: "Capture groups", Order: 4,
		Doc: "Returns the capture groups of the first match as a dict. Keys are \"0\" (whole match), \"1\", \"2\"... and, for named groups, also the group name. " +
			"Groups that did not participate are empty strings; no match gives an empty dict." + syntaxNote,
		Params:  []tools.Param{pPattern, pText, pFlags},
		Returns: "dict[str, str]", Example: `regex.groups("(?P<y>\\d{4})-(?P<m>\\d\\d)", "on 2026-10-17")`,
		Tags: []string{"capture", "named group", "submatch", "parts"},
		Fn: func(a tools.Args) (any, error) {
			re, t, err := prep(a)
			if err != nil {
				return nil, err
			}
			return groupMap(re, re.FindStringSubmatch(t)), nil
		},
	})
	reg(tools.Func{
		Name: "all_groups", Title: "Groups of every match", Order: 5,
		Doc: "Returns one JSON object per match, each mapping \"0\" (whole match), \"1\", \"2\"... and group names to the captured text, e.g. {\"0\":\"a=1\",\"1\":\"a\",\"2\":\"1\"}. " +
			"Use limit to cap the number of matches." + syntaxNote,
		Params:  []tools.Param{pPattern, pText, pFlags, opt("limit", "int", "Maximum number of matches; 0 means no explicit limit.", int64(0))},
		Returns: "list[str]", Example: `regex.all_groups("(\\w+)=(\\d+)", "a=1 b=2")`,
		Tags: []string{"captures", "findall groups", "submatches", "json"},
		Fn: func(a tools.Args) (any, error) {
			re, t, err := prep(a)
			if err != nil {
				return nil, err
			}
			out := []string{}
			for _, m := range re.FindAllStringSubmatch(t, clampLimit(a.Int("limit"))) {
				b, _ := json.Marshal(groupMap(re, m))
				out = append(out, string(b))
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "count", Title: "Count matches", Order: 6,
		Doc:     "Counts the non-overlapping matches of the pattern in the text." + syntaxNote,
		Params:  []tools.Param{pPattern, pText, pFlags},
		Returns: "int", Example: `regex.count("a", "banana")`,
		Tags: []string{"number of matches", "occurrences", "how many"},
		Fn: func(a tools.Args) (any, error) {
			re, t, err := prep(a)
			if err != nil {
				return nil, err
			}
			return int64(len(re.FindAllStringIndex(t, -1))), nil
		},
	})
	reg(tools.Func{
		Name: "positions", Title: "Match positions", Order: 7,
		Doc:     "Lists every match as \"start-end: match\" with zero-based rune offsets (end is exclusive), so offsets are correct for non-ASCII text." + syntaxNote,
		Params:  []tools.Param{pPattern, pText, pFlags, opt("limit", "int", "Maximum number of matches; 0 means no explicit limit.", int64(0))},
		Returns: "list[str]", Example: `regex.positions("an", "banana")`,
		Tags: []string{"index", "offset", "where", "span", "location"},
		Fn: func(a tools.Args) (any, error) {
			re, t, err := prep(a)
			if err != nil {
				return nil, err
			}
			out := []string{}
			last, lastRune := 0, 0
			for _, m := range re.FindAllStringIndex(t, clampLimit(a.Int("limit"))) {
				lastRune += utf8.RuneCountInString(t[last:m[0]])
				start := lastRune
				n := utf8.RuneCountInString(t[m[0]:m[1]])
				lastRune += n
				last = m[1]
				out = append(out, fmt.Sprintf("%d-%d: %s", start, start+n, t[m[0]:m[1]]))
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "replace", Title: "Replace", Order: 8,
		Doc: "Replaces matches with a template. In the replacement, $1, ${1} and ${name} insert capture groups, $0 is the whole match and $$ is a literal dollar. " +
			"limit replaces only the first N matches (0 means all)." + syntaxNote,
		Params:  []tools.Param{pPattern, pText, req("replacement", "str", "Replacement template with $1 / ${name} references."), pFlags, opt("limit", "int", "Replace at most this many matches; 0 means all.", int64(0))},
		Returns: "str", Example: `regex.replace("(\\w+)@(\\w+)", "bob@home", "$2:$1")`,
		Tags: []string{"substitute", "sub", "rewrite", "template", "backreference"},
		Fn: func(a tools.Args) (any, error) {
			re, t, err := prep(a)
			if err != nil {
				return nil, err
			}
			return replaceN(re, t, a.Str("replacement"), int(a.Int("limit")), false), nil
		},
	})
	reg(tools.Func{
		Name: "replace_literal", Title: "Replace (literal text)", Order: 9,
		Doc:     "Replaces matches with the replacement text taken literally, so $ and backslashes have no special meaning. limit replaces only the first N matches (0 means all)." + syntaxNote,
		Params:  []tools.Param{pPattern, pText, req("replacement", "str", "Literal replacement text."), pFlags, opt("limit", "int", "Replace at most this many matches; 0 means all.", int64(0))},
		Returns: "str", Example: `regex.replace_literal("\\d+", "a1b22", "$#")`,
		Tags: []string{"substitute", "raw replacement", "no template"},
		Fn: func(a tools.Args) (any, error) {
			re, t, err := prep(a)
			if err != nil {
				return nil, err
			}
			return replaceN(re, t, a.Str("replacement"), int(a.Int("limit")), true), nil
		},
	})
	reg(tools.Func{
		Name: "split", Title: "Split by regex", Order: 10,
		Doc:     "Splits the text around matches of the pattern. limit is the maximum number of pieces (0 means no limit); an empty text gives one empty piece." + syntaxNote,
		Params:  []tools.Param{pPattern, pText, opt("limit", "int", "Maximum number of pieces; 0 means no limit.", int64(0)), pFlags},
		Returns: "list[str]", Example: `regex.split("[,;]\\s*", "a, b;c")`,
		Tags: []string{"divide", "tokenize", "separate", "cut"},
		Fn: func(a tools.Args) (any, error) {
			re, t, err := prep(a)
			if err != nil {
				return nil, err
			}
			n := int(a.Int("limit"))
			if n <= 0 || n > maxResults {
				n = -1
			}
			out := re.Split(t, n)
			if len(out) > maxResults {
				out = out[:maxResults]
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "escape", Title: "Escape for regex", Order: 11,
		Doc:     "Escapes every regex metacharacter so the text matches itself literally (Go regexp.QuoteMeta).",
		Params:  []tools.Param{req("text", "str", "Literal text to escape.")},
		Returns: "str", Example: `regex.escape("1+1=2 (really?)")`,
		Tags: []string{"quote", "quotemeta", "literal", "sanitize"},
		Fn: func(a tools.Args) (any, error) {
			return regexp.QuoteMeta(a.Str("text")), nil
		},
	})
	reg(tools.Func{
		Name: "is_valid", Title: "Is valid pattern", Order: 12, Input: "pattern",
		Doc:     "Reports whether the pattern compiles under Go RE2 syntax. Patterns using lookarounds or backreferences are reported as invalid.",
		Params:  []tools.Param{req("pattern", "str", "Pattern to check.")},
		Returns: "bool", Example: `regex.is_valid("(a|b")`,
		Tags: []string{"valid", "syntax check", "compile", "lint"},
		Fn: func(a tools.Args) (any, error) {
			_, err := compile(a.Str("pattern"), "")
			return err == nil, nil
		},
	})
	reg(tools.Func{
		Name: "validate", Title: "Validate pattern", Order: 13, Input: "pattern",
		Doc:     `Compiles the pattern and returns "ok" or the compiler error message (for example "missing closing ): ..."). Lookarounds and backreferences are reported as unsupported syntax.`,
		Params:  []tools.Param{req("pattern", "str", "Pattern to check.")},
		Returns: "str", Example: `regex.validate("(?<=a)b")`,
		Tags: []string{"error", "syntax", "diagnose", "compile"},
		Fn: func(a tools.Args) (any, error) {
			if _, err := compile(a.Str("pattern"), ""); err != nil {
				return err.Error(), nil
			}
			return "ok", nil
		},
	})
	reg(tools.Func{
		Name: "explain", Title: "Explain pattern", Order: 14, Input: "pattern",
		Doc:     "Explains a pattern token by token in plain English: anchors, classes, escapes, quantifiers, groups and alternation. Indentation shows group nesting. Covers common constructs, not every RE2 corner.",
		Params:  []tools.Param{req("pattern", "str", "Pattern to explain.")},
		Returns: "list[str]", Example: `regex.explain("^(\\d{3})-[a-z]+$")`,
		Tags: []string{"describe", "understand", "breakdown", "plain english", "tokens"},
		Fn: func(a tools.Args) (any, error) {
			return explain(a.Str("pattern")), nil
		},
	})
	reg(tools.Func{
		Name: "lines_matching", Title: "Lines matching", Order: 15,
		Doc:     "Returns the lines of the text where the pattern matches (grep). With invert, returns the lines that do not match (grep -v)." + syntaxNote,
		Params:  []tools.Param{pPattern, pText, opt("invert", "bool", "Return the non-matching lines instead.", false), pFlags},
		Returns: "list[str]", Example: `regex.lines_matching("^ERROR", "ok\nERROR x\nok")`,
		Tags: []string{"grep", "filter lines", "findstr", "select lines"},
		Fn: func(a tools.Args) (any, error) {
			re, t, err := prep(a)
			if err != nil {
				return nil, err
			}
			inv := a.Bool("invert")
			out := []string{}
			for _, ln := range strings.Split(strings.ReplaceAll(t, "\r\n", "\n"), "\n") {
				if re.MatchString(ln) != inv {
					out = append(out, ln)
					if len(out) >= maxResults {
						break
					}
				}
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "extract_between", Title: "Extract between delimiters", Order: 16,
		Doc:     "Returns every piece of text found between the literal start and end delimiters (not regexes), non-overlapping and shortest-first. Nothing is returned when a delimiter is empty or missing.",
		Params:  []tools.Param{pText, req("start", "str", "Opening delimiter (literal)."), req("end", "str", "Closing delimiter (literal).")},
		Returns: "list[str]", Example: `regex.extract_between("a[1] b[22]", "[", "]")`,
		Tags: []string{"between", "substring", "delimiters", "brackets", "inside"},
		Fn: func(a tools.Args) (any, error) {
			t, s, e := a.Str("text"), a.Str("start"), a.Str("end")
			out := []string{}
			if s == "" || e == "" {
				return out, nil
			}
			for len(out) < maxResults {
				i := strings.Index(t, s)
				if i < 0 {
					break
				}
				t = t[i+len(s):]
				j := strings.Index(t, e)
				if j < 0 {
					break
				}
				out = append(out, t[:j])
				t = t[j+len(e):]
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "named_group_names", Title: "Named group names", Order: 17, Input: "pattern",
		Doc:     "Lists the names of the named capture groups in the pattern, in order. Unnamed groups are not listed.",
		Params:  []tools.Param{req("pattern", "str", "Pattern to inspect.")},
		Returns: "list[str]", Example: `regex.named_group_names("(?P<y>\\d+)-(?P<m>\\d+)")`,
		Tags: []string{"groups", "capture names", "fields"},
		Fn: func(a tools.Args) (any, error) {
			re, err := compile(a.Str("pattern"), "")
			if err != nil {
				return nil, err
			}
			out := []string{}
			for _, n := range re.SubexpNames() {
				if n != "" {
					out = append(out, n)
				}
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "literal_prefix", Title: "Literal prefix", Order: 18, Input: "pattern",
		Doc:     "Returns the literal text every match must start with (empty when there is none). A fully literal pattern returns itself.",
		Params:  []tools.Param{req("pattern", "str", "Pattern to inspect.")},
		Returns: "str", Example: `regex.literal_prefix("abc\\d+")`,
		Tags: []string{"prefix", "optimize", "fixed text", "starts with"},
		Fn: func(a tools.Args) (any, error) {
			re, err := compile(a.Str("pattern"), "")
			if err != nil {
				return nil, err
			}
			p, _ := re.LiteralPrefix()
			return p, nil
		},
	})
	reg(tools.Func{
		Name: "from_glob", Title: "Glob to regex", Order: 19, Input: "glob",
		Doc:     "Converts a shell glob into an anchored regex: * becomes .* (any characters), ? becomes . (one character), [abc] classes are kept, {a,b} becomes (a|b); everything else is escaped.",
		Params:  []tools.Param{req("glob", "str", "Glob such as *.txt or file?.[ch].")},
		Returns: "str", Example: `regex.from_glob("*.{jpg,png}")`,
		Tags: []string{"glob", "wildcard to regex", "shell pattern", "fnmatch"},
		Fn:   func(a tools.Args) (any, error) { return globToRegex(a.Str("glob")), nil },
	})
	reg(tools.Func{
		Name: "to_glob", Title: "Regex to glob", Order: 20, Input: "pattern",
		Doc:     "Converts a simple regex such as ^a.*\\.txt$ back into a glob (.* becomes *, . becomes ?, escapes are removed). Fails for regexes using groups, alternation, classes with quantifiers or other constructs a glob cannot express.",
		Params:  []tools.Param{req("pattern", "str", "Simple regex made of literals, . and .*")},
		Returns: "str", Example: `regex.to_glob("^img_.*\\.png$")`,
		Tags: []string{"regex to wildcard", "reverse glob"},
		Fn:   func(a tools.Args) (any, error) { return regexToGlob(a.Str("pattern")) },
	})
	reg(tools.Func{
		Name: "wildcard_match", Title: "Wildcard match", Order: 21,
		Doc:     "Matches text against a wildcard pattern where * is any run of characters and ? is exactly one character; the whole text must match. Case-sensitive unless ignore_case is true.",
		Params:  []tools.Param{req("pattern", "str", "Wildcard pattern using * and ?."), pText, opt("ignore_case", "bool", "Compare case-insensitively.", false)},
		Returns: "bool", Example: `regex.wildcard_match("*.txt", "notes.txt")`,
		Tags: []string{"glob match", "like", "fnmatch", "star", "dos wildcard"},
		Fn: func(a tools.Args) (any, error) {
			if err := checkText(a.Str("text")); err != nil {
				return nil, err
			}
			return wildcard(a.Str("pattern"), a.Str("text"), a.Bool("ignore_case")), nil
		},
	})
	reg(tools.Func{
		Name: "sed_like", Title: "sed substitution", Order: 22,
		Doc: "Applies sed-style substitution expressions such as s/foo/bar/g to the text. One expression per line; any delimiter works (s|a|b|). Flags: g (all), i (ignore case), m, s, or a number N to replace only the Nth match. " +
			"The replacement can use \\1..\\9 and & (whole match); use \\& for a literal ampersand. Uses RE2, so no lookarounds.",
		Params:  []tools.Param{pText, req("expression", "str", `Substitution such as "s/a/b/g" (several allowed, one per line).`)},
		Returns: "str", Example: `regex.sed_like("hello world", "s/o/0/g")`,
		Tags: []string{"sed", "substitute", "unix", "s///"},
		Fn: func(a tools.Args) (any, error) {
			if err := checkText(a.Str("text")); err != nil {
				return nil, err
			}
			return sed(a.Str("text"), a.Str("expression"))
		},
	})
	reg(tools.Func{
		Name: "presets", Title: "Preset patterns", Order: 23, Input: "name",
		Doc:     "Returns all built-in named patterns (email, url, ipv4, uuid, hex color, semver, jwt and more) as a name to pattern dict. The patterns are unanchored; wrap with ^...$ for a full match.",
		Params:  []tools.Param{opt("name", "str", "Ignored; present so the launcher can offer the function.", "")},
		Returns: "dict[str, str]", Example: `regex.presets()`,
		Tags: []string{"library", "common patterns", "cookbook", "list", "email", "url", "uuid"},
		Fn: func(a tools.Args) (any, error) {
			out := make(map[string]string, len(presetList))
			for _, p := range presetList {
				out[p.name] = p.pattern
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "preset", Title: "Preset pattern", Order: 24, Input: "name",
		Doc:     "Returns the regex of a built-in preset by name (see regex.presets), ignoring case, spaces and dashes. Fails with the list of known names when the name is unknown.",
		Params:  []tools.Param{req("name", "str", "Preset name such as email, ipv4, uuid or hex_color.")},
		Returns: "str", Example: `regex.preset("email")`,
		Tags: []string{"common pattern", "email regex", "url regex", "uuid regex", "lookup"},
		Fn: func(a tools.Args) (any, error) {
			p, err := findPreset(a.Str("name"))
			return p.pattern, err
		},
	})
	reg(tools.Func{
		Name: "test_preset", Title: "Test with preset", Order: 25,
		Doc:     "Tests the text against a built-in preset (see regex.presets). By default it reports whether the preset occurs anywhere in the text; with full true the whole text must match.",
		Params:  []tools.Param{req("name", "str", "Preset name such as email, ipv4, uuid."), pText, opt("full", "bool", "Require the whole text to match instead of a substring.", false)},
		Returns: "bool", Example: `regex.test_preset("uuid", "123e4567-e89b-12d3-a456-426614174000", True)`,
		Tags: []string{"validate", "is email", "is uuid", "is ip", "common patterns"},
		Fn: func(a tools.Args) (any, error) {
			if err := checkText(a.Str("text")); err != nil {
				return nil, err
			}
			p, err := findPreset(a.Str("name"))
			if err != nil {
				return nil, err
			}
			pat := p.pattern
			if a.Bool("full") {
				pat = `\A(?:` + pat + `)\z`
			}
			re, err := regexp.Compile(pat)
			if err != nil {
				return nil, err
			}
			return re.MatchString(a.Str("text")), nil
		},
	})
}

func groupMap(re *regexp.Regexp, m []string) map[string]string {
	out := map[string]string{}
	names := re.SubexpNames()
	for i, s := range m {
		out[strconv.Itoa(i)] = s
		if i < len(names) && names[i] != "" {
			out[names[i]] = s
		}
	}
	return out
}

func replaceN(re *regexp.Regexp, t, repl string, limit int, literal bool) string {
	if limit <= 0 {
		if literal {
			return re.ReplaceAllLiteralString(t, repl)
		}
		return re.ReplaceAllString(t, repl)
	}
	var b strings.Builder
	last := 0
	for _, m := range re.FindAllStringSubmatchIndex(t, limit) {
		b.WriteString(t[last:m[0]])
		if literal {
			b.WriteString(repl)
		} else {
			b.Write(re.ExpandString(nil, repl, t, m))
		}
		last = m[1]
	}
	b.WriteString(t[last:])
	return b.String()
}

func wildcard(pat, text string, fold bool) bool {
	if fold {
		pat, text = strings.ToLower(pat), strings.ToLower(text)
	}
	p, s := []rune(pat), []rune(text)
	pi, si, star, mark := 0, 0, -1, 0
	for si < len(s) {
		switch {
		case pi < len(p) && p[pi] == '*':
			star, mark = pi, si
			pi++
		case pi < len(p) && (p[pi] == '?' || p[pi] == s[si]):
			pi++
			si++
		case star >= 0:
			pi = star + 1
			mark++
			si = mark
		default:
			return false
		}
	}
	for pi < len(p) && p[pi] == '*' {
		pi++
	}
	return pi == len(p)
}

func globToRegex(g string) string {
	var b strings.Builder
	b.WriteString("^")
	rs := []rune(g)
	inBrace := 0
	for i := 0; i < len(rs); i++ {
		c := rs[i]
		switch c {
		case '*':
			if i+1 < len(rs) && rs[i+1] == '*' {
				i++
			}
			b.WriteString(".*")
		case '?':
			b.WriteString(".")
		case '[':
			j := i + 1
			if j < len(rs) && (rs[j] == '!' || rs[j] == '^') {
				j++
			}
			if j < len(rs) && rs[j] == ']' {
				j++
			}
			for j < len(rs) && rs[j] != ']' {
				j++
			}
			if j >= len(rs) {
				b.WriteString(`\[`)
				continue
			}
			body := string(rs[i+1 : j])
			if strings.HasPrefix(body, "!") {
				body = "^" + body[1:]
			}
			b.WriteString("[" + strings.ReplaceAll(body, `\`, `\\`) + "]")
			i = j
		case '{':
			inBrace++
			b.WriteString("(?:")
		case '}':
			if inBrace > 0 {
				inBrace--
				b.WriteString(")")
			} else {
				b.WriteString(`\}`)
			}
		case ',':
			if inBrace > 0 {
				b.WriteString("|")
			} else {
				b.WriteString(",")
			}
		case '\\':
			if i+1 < len(rs) {
				i++
				b.WriteString(regexp.QuoteMeta(string(rs[i])))
			} else {
				b.WriteString(`\\`)
			}
		default:
			b.WriteString(regexp.QuoteMeta(string(c)))
		}
	}
	for ; inBrace > 0; inBrace-- {
		b.WriteString(")")
	}
	b.WriteString("$")
	return b.String()
}

func regexToGlob(p string) (string, error) {
	p = strings.TrimPrefix(p, "^")
	p = strings.TrimSuffix(p, "$")
	var b strings.Builder
	rs := []rune(p)
	for i := 0; i < len(rs); i++ {
		c := rs[i]
		switch c {
		case '.':
			if i+1 < len(rs) && rs[i+1] == '*' {
				b.WriteString("*")
				i++
			} else if i+1 < len(rs) && rs[i+1] == '+' {
				b.WriteString("?*")
				i++
			} else {
				b.WriteString("?")
			}
		case '\\':
			if i+1 >= len(rs) {
				return "", fmt.Errorf("dangling backslash")
			}
			i++
			n := rs[i]
			if strings.ContainsRune(`dDwWsSbBAzpPQE`, n) || (n >= '0' && n <= '9') {
				return "", fmt.Errorf("cannot express \\%c as a glob", n)
			}
			if strings.ContainsRune("*?[{", n) {
				b.WriteString(`\`)
			}
			b.WriteRune(n)
		case '*', '+', '?', '(', ')', '|', '{', '}', '[', ']', '^', '$':
			return "", fmt.Errorf("cannot express %q as a glob", string(c))
		default:
			b.WriteRune(c)
		}
	}
	return b.String(), nil
}

var sedFlagRe = regexp.MustCompile(`^[gimsp0-9]*$`)

func sed(text, expr string) (string, error) {
	n := 0
	for _, ln := range strings.Split(strings.ReplaceAll(expr, "\r\n", "\n"), "\n") {
		ln = strings.TrimSpace(ln)
		if ln == "" {
			continue
		}
		n++
		if len(ln) < 4 || ln[0] != 's' {
			return "", fmt.Errorf("expression %q: expected s/pattern/replacement/flags", ln)
		}
		delim := ln[1]
		if delim == '\\' || delim == ' ' {
			return "", fmt.Errorf("expression %q: bad delimiter", ln)
		}
		parts := splitDelim(ln[2:], delim)
		if len(parts) != 3 {
			return "", fmt.Errorf("expression %q: expected s%cpattern%creplacement%cflags", ln, delim, delim, delim)
		}
		pat, repl, fl := parts[0], parts[1], strings.TrimSpace(parts[2])
		if !sedFlagRe.MatchString(fl) {
			return "", fmt.Errorf("expression %q: unknown flags %q", ln, fl)
		}
		cf, digits := "", ""
		global := false
		for _, c := range fl {
			switch {
			case c == 'g':
				global = true
			case c == 'i' || c == 'm' || c == 's':
				cf += string(c)
			case c >= '0' && c <= '9':
				digits += string(c)
			}
		}
		nth := 0
		if digits != "" {
			nth, _ = strconv.Atoi(digits)
		}
		re, err := compile(pat, cf)
		if err != nil {
			return "", err
		}
		tmpl := sedTemplate(repl)
		switch {
		case nth > 0 && global:
			text = replaceFrom(re, text, tmpl, nth)
		case nth > 0:
			text = replaceNth(re, text, tmpl, nth)
		case global:
			text = re.ReplaceAllString(text, tmpl)
		default:
			text = replaceN(re, text, tmpl, 1, false)
		}
	}
	if n == 0 {
		return "", fmt.Errorf("no substitution expression given")
	}
	return text, nil
}

func splitDelim(s string, d byte) []string {
	var parts []string
	var cur strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			if s[i+1] == d {
				cur.WriteByte(d)
			} else {
				cur.WriteByte('\\')
				cur.WriteByte(s[i+1])
			}
			i++
			continue
		}
		if s[i] == d && len(parts) < 2 {
			parts = append(parts, cur.String())
			cur.Reset()
			continue
		}
		cur.WriteByte(s[i])
	}
	return append(parts, cur.String())
}

func sedTemplate(r string) string {
	var b strings.Builder
	for i := 0; i < len(r); i++ {
		c := r[i]
		switch {
		case c == '\\' && i+1 < len(r):
			n := r[i+1]
			i++
			switch {
			case n >= '0' && n <= '9':
				b.WriteString("${" + string(n) + "}")
			case n == 'n':
				b.WriteByte('\n')
			case n == 't':
				b.WriteByte('\t')
			case n == '$':
				b.WriteString("$$")
			default:
				b.WriteByte(n)
			}
		case c == '&':
			b.WriteString("${0}")
		case c == '$':
			b.WriteString("$$")
		default:
			b.WriteByte(c)
		}
	}
	return b.String()
}

func replaceNth(re *regexp.Regexp, t, tmpl string, nth int) string {
	ms := re.FindAllStringSubmatchIndex(t, nth)
	if len(ms) < nth {
		return t
	}
	m := ms[nth-1]
	return t[:m[0]] + string(re.ExpandString(nil, tmpl, t, m)) + t[m[1]:]
}

func replaceFrom(re *regexp.Regexp, t, tmpl string, nth int) string {
	var b strings.Builder
	last := 0
	for i, m := range re.FindAllStringSubmatchIndex(t, -1) {
		if i+1 < nth {
			continue
		}
		b.WriteString(t[last:m[0]])
		b.Write(re.ExpandString(nil, tmpl, t, m))
		last = m[1]
	}
	b.WriteString(t[last:])
	return b.String()
}

type preset struct{ name, pattern string }

func findPreset(name string) (preset, error) {
	key := strings.NewReplacer(" ", "_", "-", "_").Replace(strings.ToLower(strings.TrimSpace(name)))
	for _, p := range presetList {
		if p.name == key {
			return p, nil
		}
	}
	if key != "" {
		for _, p := range presetList {
			if strings.HasPrefix(p.name, key) {
				return p, nil
			}
		}
	}
	names := make([]string, len(presetList))
	for i, p := range presetList {
		names[i] = p.name
	}
	sort.Strings(names)
	return preset{}, fmt.Errorf("unknown preset %q; known: %s", name, strings.Join(names, ", "))
}

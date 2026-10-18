package text

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

// maxOut is the largest string any function in this package will produce.
const maxOut = 4 << 20

// maxCells caps dynamic-programming tables (distance and diff functions).
const maxCells = 50_000_000

type P = tools.Param

func ps(n, d string) P  { return P{Name: n, Type: "str", Doc: d, Required: true} }
func pir(n, d string) P { return P{Name: n, Type: "int", Doc: d, Required: true} }
func pso(n, d, def string, choices ...string) P {
	return P{Name: n, Type: "str", Doc: d, Default: def, Choices: choices}
}
func pi(n, d string, def int64) P { return P{Name: n, Type: "int", Doc: d, Default: def} }
func pb(n, d string, def bool) P  { return P{Name: n, Type: "bool", Doc: d, Default: def} }
func pl(n, d string) P            { return P{Name: n, Type: "list[str]", Doc: d, Required: true} }

type fn = func(a tools.Args) (any, error)

// reg registers a text function. Variants whose extra parameters are all
// optional also become launcher variants of the "text" family.
func reg(name, title, doc, ret, example string, order int, tags string, params []P, f fn) {
	t := tools.Func{
		Group: "text", Name: name, Title: title, Doc: doc, Params: params,
		Returns: ret, Fn: f, Example: example, Tags: strings.Fields(tags), Order: order,
	}
	if launcher(params) {
		t.Family = "text"
		t.Keywords = launcherKeywords[name]
	}
	tools.Register(t)
}

// launcherKeywords are the plain words that select one function from the
// launcher (lower HELLO). Only distinctive words belong here: a common word
// would hijack ordinary queries.
var launcherKeywords = map[string][]string{
	"lower": {"lower", "lowercase"}, "upper": {"upper", "uppercase"}, "slug": {"slug", "slugify"},
	"camel": {"camel", "camelcase"}, "pascal": {"pascal"}, "snake": {"snake"}, "kebab": {"kebab"},
	"is_palindrome": {"palindrome"}, "palindrome_report": {"palindrome"}, "leetspeak": {"leet"},
	"strip_accents": {"unaccent"}, "case_all": {"case"},
}

func launcher(params []P) bool {
	seen := false
	for _, p := range params {
		if !seen && p.Type == "str" {
			seen = true
			if !p.Required {
				return false
			}
			continue
		}
		if p.Required {
			return false
		}
	}
	return seen
}

// simple registers a one-argument string -> string function.
func simple(name, title, doc, example string, order int, tags string, f func(string) string) {
	reg(name, title, doc, "str", example, order, tags, []P{ps("s", "Input text.")},
		func(a tools.Args) (any, error) { return f(a.Str("s")), nil })
}

func checkSize(n int) error {
	if n > maxOut || n < 0 {
		return fmt.Errorf("result would exceed %d MiB", maxOut>>20)
	}
	return nil
}

func isSpace(r rune) bool { return unicode.IsSpace(r) }

// isMark reports combining marks (Mn, Mc, Me).
func isMark(r rune) bool { return unicode.Is(unicode.M, r) }

// normNL converts CRLF and lone CR to LF.
func normNL(s string) string {
	if !strings.ContainsRune(s, '\r') {
		return s
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

func isApos(r rune) bool { return r == '\'' || r == '’' }

// words splits s into words: runs of letters, digits and marks that may
// contain apostrophes and hyphens between them ("don't", "well-known").
func words(s string) []string {
	var out []string
	var cur []rune
	rs := []rune(s)
	flush := func() {
		for len(cur) > 0 && !isWordCore(cur[len(cur)-1]) {
			cur = cur[:len(cur)-1]
		}
		if len(cur) > 0 {
			out = append(out, string(cur))
		}
		cur = cur[:0]
	}
	for i, r := range rs {
		switch {
		case isWordCore(r):
			cur = append(cur, r)
		case (isApos(r) || r == '-') && len(cur) > 0 && i+1 < len(rs) && isWordCore(rs[i+1]):
			cur = append(cur, r)
		default:
			flush()
		}
	}
	flush()
	return out
}

func isWordCore(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) || isMark(r) }

// splitIdent splits an identifier-like string into lower-case words, honouring
// separators, camelCase boundaries and acronyms ("HTMLParser" -> html, parser).
func splitIdent(s string) []string {
	var out []string
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			out = append(out, strings.ToLower(string(cur)))
			cur = cur[:0]
		}
	}
	rs := []rune(s)
	for i, r := range rs {
		if isApos(r) {
			continue
		}
		if !isWordCore(r) {
			flush()
			continue
		}
		if unicode.IsUpper(r) && len(cur) > 0 {
			prev := cur[len(cur)-1]
			nextLower := i+1 < len(rs) && unicode.IsLower(rs[i+1])
			if unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && nextLower) {
				flush()
			}
		}
		cur = append(cur, r)
	}
	flush()
	return out
}

func upperFirst(w string) string {
	r, n := utf8.DecodeRuneInString(w)
	if n == 0 {
		return w
	}
	return string(unicode.ToTitle(r)) + w[n:]
}

func lowerFirst(w string) string {
	r, n := utf8.DecodeRuneInString(w)
	if n == 0 {
		return w
	}
	return string(unicode.ToLower(r)) + w[n:]
}

// graphemes splits s into approximate extended grapheme clusters: a base rune
// followed by combining marks, variation selectors, emoji modifiers, tag
// characters and ZWJ-joined runes; regional indicators pair up; CRLF is one.
func graphemes(s string) []string {
	rs := []rune(s)
	var out []string
	i := 0
	for i < len(rs) {
		start := i
		r := rs[i]
		i++
		if r == '\r' && i < len(rs) && rs[i] == '\n' {
			i++
		} else if isRI(r) && i < len(rs) && isRI(rs[i]) {
			i++
		} else if r != '\n' && r != '\r' {
			for i < len(rs) {
				n := rs[i]
				switch {
				case isMark(n) || (n >= 0x1F3FB && n <= 0x1F3FF) || (n >= 0xE0020 && n <= 0xE007F) || n == 0x200C:
					i++
				case n == 0x200D && i+1 < len(rs):
					i += 2
				default:
					goto done
				}
			}
		}
	done:
		out = append(out, string(rs[start:i]))
	}
	return out
}

func isRI(r rune) bool { return r >= 0x1F1E6 && r <= 0x1F1FF }

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func bstr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

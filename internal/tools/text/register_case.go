package text

import (
	"strings"
	"unicode"

	"github.com/avestura/zagros/internal/tools"
	"golang.org/x/text/unicode/norm"
)

func titleCase(s string) string {
	var b strings.Builder
	prevLetter := false
	rs := []rune(s)
	for i, r := range rs {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if prevLetter {
				b.WriteRune(unicode.ToLower(r))
			} else {
				b.WriteRune(unicode.ToTitle(r))
			}
			prevLetter = true
		case isMark(r):
			b.WriteRune(r)
		case isApos(r) && prevLetter && i+1 < len(rs) && unicode.IsLetter(rs[i+1]):
			b.WriteRune(r)
		default:
			b.WriteRune(r)
			prevLetter = false
		}
	}
	return b.String()
}

func sentenceCase(s string) string {
	rs := []rune(strings.ToLower(s))
	start := true
	for i, r := range rs {
		switch {
		case unicode.IsLetter(r):
			if start {
				rs[i] = unicode.ToTitle(r)
				start = false
			}
		case unicode.IsDigit(r):
			start = false
		case r == '.' || r == '!' || r == '?' || r == '…' || r == '\n':
			if i+1 >= len(rs) || unicode.IsSpace(rs[i+1]) || r == '\n' {
				start = true
			}
		}
	}
	return string(rs)
}

func swapCase(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case unicode.IsUpper(r):
			return unicode.ToLower(r)
		case unicode.IsLower(r):
			return unicode.ToUpper(r)
		}
		return r
	}, s)
}

func alternating(s string) string {
	up := false
	return strings.Map(func(r rune) rune {
		if !unicode.IsLetter(r) {
			return r
		}
		up = !up
		if up {
			return unicode.ToLower(r)
		}
		return unicode.ToUpper(r)
	}, s)
}

func joinWords(ws []string, sep string, f func(i int, w string) string) string {
	out := make([]string, len(ws))
	for i, w := range ws {
		out[i] = f(i, w)
	}
	return strings.Join(out, sep)
}

func camelCase(s string) string {
	return joinWords(splitIdent(s), "", func(i int, w string) string {
		if i == 0 {
			return w
		}
		return upperFirst(w)
	})
}
func pascalCase(s string) string {
	return joinWords(splitIdent(s), "", func(_ int, w string) string { return upperFirst(w) })
}
func snakeCase(s string) string { return strings.Join(splitIdent(s), "_") }
func kebabCase(s string) string { return strings.Join(splitIdent(s), "-") }
func constantCase(s string) string {
	return strings.ToUpper(strings.Join(splitIdent(s), "_"))
}
func dotCase(s string) string  { return strings.Join(splitIdent(s), ".") }
func pathCase(s string) string { return strings.Join(splitIdent(s), "/") }
func trainCase(s string) string {
	return joinWords(splitIdent(s), "-", func(_ int, w string) string { return upperFirst(w) })
}

var asciiSpecial = map[rune]string{
	'ß': "ss", 'ẞ': "SS", 'æ': "ae", 'Æ': "AE", 'œ': "oe", 'Œ': "OE", 'ø': "o", 'Ø': "O",
	'đ': "d", 'Đ': "D", 'ð': "d", 'Ð': "D", 'ł': "l", 'Ł': "L", 'þ': "th", 'Þ': "TH", 'ı': "i", 'ħ': "h", 'Ħ': "H",
}

// foldASCII drops diacritics and transliterates a few Latin letters
// (ß, æ, ø, đ, ł ...) that have no decomposition.
func foldASCII(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if unicode.Is(unicode.Mn, r) {
			continue
		}
		if t, ok := asciiSpecial[r]; ok {
			b.WriteString(t)
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func slugify(s, sep string, ascii bool) string {
	if ascii {
		s = foldASCII(s)
	}
	s = strings.ToLower(s)
	var b strings.Builder
	pending := false
	for _, r := range s {
		ok := unicode.IsLetter(r) || unicode.IsDigit(r) || (!ascii && isMark(r))
		if ascii && r > 127 {
			ok = false
		}
		if ok {
			if pending && b.Len() > 0 {
				b.WriteString(sep)
			}
			pending = false
			b.WriteRune(r)
		} else if !isApos(r) {
			pending = true
		}
	}
	return b.String()
}

func capitalizeWords(s string) string {
	var b strings.Builder
	start := true
	for _, r := range s {
		if unicode.IsSpace(r) {
			start = true
			b.WriteRune(r)
			continue
		}
		if start {
			b.WriteRune(unicode.ToTitle(r))
			start = false
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func caseAll(s string) map[string]string {
	return map[string]string{
		"lower": strings.ToLower(s), "upper": strings.ToUpper(s), "title": titleCase(s),
		"sentence": sentenceCase(s), "swapcase": swapCase(s), "alternating": alternating(s),
		"camel": camelCase(s), "pascal": pascalCase(s), "snake": snakeCase(s), "kebab": kebabCase(s),
		"constant": constantCase(s), "dot": dotCase(s), "path": pathCase(s), "train": trainCase(s),
		"slug": slugify(s, "-", true),
	}
}

func init() {
	simple("lower", "Lower case", "Converts every letter to lower case using Unicode rules (Σ -> σ). Other characters are unchanged.", `text.lower("HELLO World")`, 1, "lowercase lower small letters downcase tolower", strings.ToLower)
	simple("upper", "Upper case", "Converts every letter to upper case using Unicode simple case mapping. Other characters are unchanged.", `text.upper("hello World")`, 2, "uppercase upper capital letters caps toupper", strings.ToUpper)
	simple("title", "Title Case", "Capitalises the first letter of every word and lower-cases the rest. Apostrophes inside a word do not start a new word (don't -> Don't).", `text.title("the quick brown fox")`, 3, "titlecase title capitalize each word headline", titleCase)
	simple("sentence_case", "Sentence case", "Lower-cases the text, then capitalises the first letter after the start and after each . ! ? or line break.", `text.sentence_case("HELLO there. HOW ARE you?")`, 9, "sentence capitalize first letter proper", sentenceCase)
	simple("swapcase", "Swap case", "Inverts the case of every letter: upper becomes lower and lower becomes upper.", `text.swapcase("Hello World")`, 10, "invert toggle flip case swap", swapCase)
	simple("alternating_case", "aLtErNaTiNg case", "Alternates lower and upper case over the letters, starting with lower (the 'spongebob' meme case). Non-letters do not advance the alternation.", `text.alternating_case("hello world")`, 11, "spongebob mocking alternating meme case random", alternating)
	simple("camel", "camelCase", "Converts to camelCase. Words are split on separators, case changes and acronyms (HTMLParser -> htmlParser); apostrophes are dropped.", `text.camel("hello world foo")`, 4, "camelcase camel identifier variable lowerCamelCase", camelCase)
	simple("pascal", "PascalCase", "Converts to PascalCase (UpperCamelCase). Words are split on separators, case changes and acronyms.", `text.pascal("hello world foo")`, 12, "pascalcase uppercamelcase class name identifier", pascalCase)
	simple("snake", "snake_case", "Converts to snake_case: lower-case words joined with underscores. Splits on separators, case changes and acronyms.", `text.snake("Hello World fooBar")`, 5, "snakecase underscore identifier python", snakeCase)
	simple("kebab", "kebab-case", "Converts to kebab-case: lower-case words joined with hyphens. Splits on separators, case changes and acronyms.", `text.kebab("Hello World fooBar")`, 6, "kebabcase dash hyphen css lisp-case identifier", kebabCase)
	simple("constant", "CONSTANT_CASE", "Converts to SCREAMING_SNAKE_CASE: upper-case words joined with underscores.", `text.constant("hello world")`, 13, "screaming snake constant macro upper_snake environment variable", constantCase)
	simple("dot_case", "dot.case", "Converts to dot.case: lower-case words joined with dots.", `text.dot_case("Hello World")`, 14, "dot case namespace package property", dotCase)
	simple("path_case", "path/case", "Converts to path/case: lower-case words joined with forward slashes.", `text.path_case("Hello World")`, 15, "path case slash directory", pathCase)
	simple("train_case", "Train-Case", "Converts to Train-Case: capitalised words joined with hyphens (HTTP header style).", `text.train_case("hello world")`, 16, "train case header http capital hyphen", trainCase)
	simple("capitalize_words", "Capitalize words", "Upper-cases the first character of every whitespace-separated word and leaves the other characters untouched.", `text.capitalize_words("hello wORLD")`, 17, "capitalise first letter each word keep rest", capitalizeWords)
	simple("uncapitalize", "uncapitalize", "Lower-cases only the first character of the text (Hello -> hello); empty input returns empty.", `text.uncapitalize("Hello World")`, 18, "decapitalize lower first letter lcfirst", lowerFirst)
	reg("slug", "URL slug", "Makes a URL slug: lower case, accents folded to ASCII (é -> e, ß -> ss) when ascii is true, every run of other characters replaced by one separator, no leading or trailing separator.",
		"str", `text.slug("Héllo, Wörld! 2024")`, 7, "slugify url permalink seo friendly path accents ascii",
		[]P{ps("s", "Input text."), pso("separator", "Separator placed between words.", "-"), pb("ascii", "Fold accents and drop non-ASCII characters.", true)},
		func(a tools.Args) (any, error) { return slugify(a.Str("s"), a.Str("separator"), a.Bool("ascii")), nil })
	reg("case_all", "All case styles", "Returns every case style of the text as a dict: lower, upper, title, sentence, swapcase, alternating, camel, pascal, snake, kebab, constant, dot, path, train, slug.",
		"dict[str, str]", `text.case_all("hello world")`, 0, "case styles convert all variants naming conventions",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) { return caseAll(a.Str("s")), nil })
}

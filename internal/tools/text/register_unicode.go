package text

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

type block struct {
	lo, hi rune
	name   string
}

// blocks lists the most used Unicode blocks (sorted by start).
var blocks = []block{
	{0x0000, 0x007F, "Basic Latin"}, {0x0080, 0x00FF, "Latin-1 Supplement"}, {0x0100, 0x017F, "Latin Extended-A"},
	{0x0180, 0x024F, "Latin Extended-B"}, {0x0250, 0x02AF, "IPA Extensions"}, {0x02B0, 0x02FF, "Spacing Modifier Letters"},
	{0x0300, 0x036F, "Combining Diacritical Marks"}, {0x0370, 0x03FF, "Greek and Coptic"}, {0x0400, 0x04FF, "Cyrillic"},
	{0x0500, 0x052F, "Cyrillic Supplement"}, {0x0530, 0x058F, "Armenian"}, {0x0590, 0x05FF, "Hebrew"},
	{0x0600, 0x06FF, "Arabic"}, {0x0700, 0x074F, "Syriac"}, {0x0750, 0x077F, "Arabic Supplement"}, {0x0780, 0x07BF, "Thaana"},
	{0x07C0, 0x07FF, "NKo"}, {0x0900, 0x097F, "Devanagari"}, {0x0980, 0x09FF, "Bengali"}, {0x0A00, 0x0A7F, "Gurmukhi"},
	{0x0A80, 0x0AFF, "Gujarati"}, {0x0B00, 0x0B7F, "Oriya"}, {0x0B80, 0x0BFF, "Tamil"}, {0x0C00, 0x0C7F, "Telugu"},
	{0x0C80, 0x0CFF, "Kannada"}, {0x0D00, 0x0D7F, "Malayalam"}, {0x0D80, 0x0DFF, "Sinhala"}, {0x0E00, 0x0E7F, "Thai"},
	{0x0E80, 0x0EFF, "Lao"}, {0x0F00, 0x0FFF, "Tibetan"}, {0x1000, 0x109F, "Myanmar"}, {0x10A0, 0x10FF, "Georgian"},
	{0x1100, 0x11FF, "Hangul Jamo"}, {0x1200, 0x137F, "Ethiopic"}, {0x13A0, 0x13FF, "Cherokee"}, {0x1780, 0x17FF, "Khmer"},
	{0x1800, 0x18AF, "Mongolian"}, {0x1E00, 0x1EFF, "Latin Extended Additional"}, {0x1F00, 0x1FFF, "Greek Extended"},
	{0x2000, 0x206F, "General Punctuation"}, {0x2070, 0x209F, "Superscripts and Subscripts"}, {0x20A0, 0x20CF, "Currency Symbols"},
	{0x20D0, 0x20FF, "Combining Diacritical Marks for Symbols"}, {0x2100, 0x214F, "Letterlike Symbols"}, {0x2150, 0x218F, "Number Forms"},
	{0x2190, 0x21FF, "Arrows"}, {0x2200, 0x22FF, "Mathematical Operators"}, {0x2300, 0x23FF, "Miscellaneous Technical"},
	{0x2400, 0x243F, "Control Pictures"}, {0x2460, 0x24FF, "Enclosed Alphanumerics"}, {0x2500, 0x257F, "Box Drawing"},
	{0x2580, 0x259F, "Block Elements"}, {0x25A0, 0x25FF, "Geometric Shapes"}, {0x2600, 0x26FF, "Miscellaneous Symbols"},
	{0x2700, 0x27BF, "Dingbats"}, {0x27C0, 0x27EF, "Miscellaneous Mathematical Symbols-A"}, {0x2800, 0x28FF, "Braille Patterns"},
	{0x2900, 0x297F, "Supplemental Arrows-B"}, {0x2B00, 0x2BFF, "Miscellaneous Symbols and Arrows"}, {0x2C00, 0x2C5F, "Glagolitic"},
	{0x2E00, 0x2E7F, "Supplemental Punctuation"}, {0x2E80, 0x2FDF, "CJK Radicals"}, {0x3000, 0x303F, "CJK Symbols and Punctuation"},
	{0x3040, 0x309F, "Hiragana"}, {0x30A0, 0x30FF, "Katakana"}, {0x3100, 0x312F, "Bopomofo"}, {0x3130, 0x318F, "Hangul Compatibility Jamo"},
	{0x31F0, 0x31FF, "Katakana Phonetic Extensions"}, {0x3200, 0x32FF, "Enclosed CJK Letters and Months"}, {0x3300, 0x33FF, "CJK Compatibility"},
	{0x3400, 0x4DBF, "CJK Unified Ideographs Extension A"}, {0x4E00, 0x9FFF, "CJK Unified Ideographs"}, {0xA000, 0xA48F, "Yi Syllables"},
	{0xAC00, 0xD7AF, "Hangul Syllables"}, {0xD800, 0xDFFF, "Surrogates"}, {0xE000, 0xF8FF, "Private Use Area"},
	{0xF900, 0xFAFF, "CJK Compatibility Ideographs"}, {0xFB00, 0xFB4F, "Alphabetic Presentation Forms"},
	{0xFB50, 0xFDFF, "Arabic Presentation Forms-A"}, {0xFE00, 0xFE0F, "Variation Selectors"}, {0xFE20, 0xFE2F, "Combining Half Marks"},
	{0xFE70, 0xFEFF, "Arabic Presentation Forms-B"}, {0xFF00, 0xFFEF, "Halfwidth and Fullwidth Forms"}, {0xFFF0, 0xFFFF, "Specials"},
	{0x10000, 0x100FF, "Linear B"}, {0x1D400, 0x1D7FF, "Mathematical Alphanumeric Symbols"}, {0x1F000, 0x1F02F, "Mahjong Tiles"},
	{0x1F0A0, 0x1F0FF, "Playing Cards"}, {0x1F100, 0x1F1FF, "Enclosed Alphanumeric Supplement"}, {0x1F200, 0x1F2FF, "Enclosed Ideographic Supplement"},
	{0x1F300, 0x1F5FF, "Miscellaneous Symbols and Pictographs"}, {0x1F600, 0x1F64F, "Emoticons"}, {0x1F680, 0x1F6FF, "Transport and Map Symbols"},
	{0x1F900, 0x1F9FF, "Supplemental Symbols and Pictographs"}, {0x1FA70, 0x1FAFF, "Symbols and Pictographs Extended-A"},
	{0x20000, 0x2A6DF, "CJK Unified Ideographs Extension B"}, {0xE0000, 0xE007F, "Tags"}, {0xF0000, 0xFFFFF, "Supplementary Private Use Area-A"},
}

func blockOf(r rune) string {
	i := sort.Search(len(blocks), func(i int) bool { return blocks[i].hi >= r })
	if i < len(blocks) && blocks[i].lo <= r {
		return blocks[i].name
	}
	return "Other"
}

var catNames = map[string]string{
	"Lu": "Uppercase Letter", "Ll": "Lowercase Letter", "Lt": "Titlecase Letter", "Lm": "Modifier Letter", "Lo": "Other Letter",
	"Mn": "Nonspacing Mark", "Mc": "Spacing Mark", "Me": "Enclosing Mark", "Nd": "Decimal Number", "Nl": "Letter Number",
	"No": "Other Number", "Pc": "Connector Punctuation", "Pd": "Dash Punctuation", "Ps": "Open Punctuation", "Pe": "Close Punctuation",
	"Pi": "Initial Punctuation", "Pf": "Final Punctuation", "Po": "Other Punctuation", "Sm": "Math Symbol", "Sc": "Currency Symbol",
	"Sk": "Modifier Symbol", "So": "Other Symbol", "Zs": "Space Separator", "Zl": "Line Separator", "Zp": "Paragraph Separator",
	"Cc": "Control", "Cf": "Format", "Cs": "Surrogate", "Co": "Private Use", "Cn": "Unassigned",
}

func category(r rune) string {
	for name := range catNames {
		if t, ok := unicode.Categories[name]; ok && unicode.Is(t, r) {
			return name
		}
	}
	return "Cn"
}

func cpString(r rune) string { return fmt.Sprintf("U+%04X", r) }

func parseCodepoint(t string) (rune, error) {
	t = strings.TrimSpace(t)
	orig := t
	base := 16
	switch {
	case strings.HasPrefix(strings.ToUpper(t), "U+"):
		t = t[2:]
	case strings.HasPrefix(strings.ToLower(t), "0x"):
		t = t[2:]
	case strings.HasPrefix(strings.ToLower(t), `\u`):
		t = strings.Trim(t[2:], "{}")
	case strings.HasPrefix(strings.ToLower(t), "&#x"):
		t = strings.TrimSuffix(t[3:], ";")
	case strings.HasPrefix(t, "&#"):
		t, base = strings.TrimSuffix(t[2:], ";"), 10
	}
	n, err := strconv.ParseInt(t, base, 32)
	if err != nil || n < 0 || n > unicode.MaxRune || (n >= 0xD800 && n <= 0xDFFF) {
		return 0, fmt.Errorf("invalid code point %q", orig)
	}
	return rune(n), nil
}

func firstRune(s string) (rune, error) {
	r, n := utf8.DecodeRuneInString(s)
	if n == 0 {
		return 0, errors.New("ch must not be empty")
	}
	return r, nil
}

func hexBytes(b []byte) string {
	parts := make([]string, len(b))
	for i, x := range b {
		parts[i] = fmt.Sprintf("%02X", x)
	}
	return strings.Join(parts, " ")
}

var loremWords = strings.Fields(`lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua enim ad minim veniam quis nostrud exercitation ullamco laboris nisi aliquip ex ea commodo consequat duis aute irure in reprehenderit voluptate velit esse cillum eu fugiat nulla pariatur excepteur sint occaecat cupidatat non proident sunt culpa qui officia deserunt mollit anim id est laborum`)

const loremStart = "Lorem ipsum dolor sit amet, consectetur adipiscing elit"

// loremGen yields deterministic words: seed 0 begins with the classic opening.
type loremGen struct {
	rng  *rand.Rand
	pos  int
	seed int64
}

func newLorem(seed int64) *loremGen {
	return &loremGen{rng: rand.New(rand.NewPCG(uint64(seed), 0x5a47524f53)), seed: seed}
}

func (g *loremGen) word() string {
	if g.seed == 0 && g.pos < 8 {
		g.pos++
		return loremWords[g.pos-1]
	}
	g.pos++
	return loremWords[g.rng.IntN(len(loremWords))]
}

func (g *loremGen) sentence(n int) string {
	first := g.pos == 0
	ws := make([]string, n)
	for i := range ws {
		ws[i] = g.word()
		if !(first && g.seed == 0) && i > 2 && i < n-2 && g.rng.IntN(9) == 0 {
			ws[i] += ","
		}
	}
	if first && g.seed == 0 && n > 5 && ws[4] == "amet" {
		ws[4] = "amet,"
	}
	return upperFirst(strings.Join(ws, " ")) + "."
}

func genLorem(nwords int, seed int64) string {
	g := newLorem(seed)
	var sents []string
	left := nwords
	for left > 0 {
		n := min(left, 5+g.rng.IntN(10))
		if left-n > 0 && left-n < 4 {
			n = left
		}
		sents = append(sents, g.sentence(n))
		left -= n
	}
	return strings.Join(sents, " ")
}

func init() {
	reg("codepoints", "Code points", "Lists the Unicode code point of every character as U+XXXX (at least four hex digits, upper case). Invalid UTF-8 bytes appear as U+FFFD.", "list[str]", `text.codepoints("aé😀")`, 90, "codepoints unicode code points hex u+ ord scalar",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) {
			out := []string{}
			for _, r := range a.Str("s") {
				out = append(out, cpString(r))
			}
			return out, nil
		})
	reg("from_codepoints", "From code points", "Builds text from code points written as U+0041, 0x41, \\u0041, &#65; (decimal) or plain hex. Invalid values or surrogates are an error.", "str", `text.from_codepoints(["U+0048", "0x69"])`, 91, "from codepoints unicode chr build text hex",
		[]P{pl("points", "Code point strings.")}, func(a tools.Args) (any, error) {
			var b strings.Builder
			for _, p := range a.List("points") {
				r, err := parseCodepoint(p)
				if err != nil {
					return nil, err
				}
				b.WriteRune(r)
			}
			return b.String(), nil
		})
	reg("to_codepoint", "Character to code point", "Returns the decimal Unicode code point of the first character; empty text is an error.", "int", `text.to_codepoint("A")`, 92, "ord codepoint number character decimal unicode",
		[]P{ps("ch", "A character; only the first one is used.")}, func(a tools.Args) (any, error) {
			r, err := firstRune(a.Str("ch"))
			return int64(r), err
		})
	reg("from_codepoint", "Code point to character", "Returns the character for a decimal code point (0 to 0x10FFFF, no surrogates).", "str", `text.from_codepoint(128512)`, 93, "chr from codepoint number character decimal unicode",
		[]P{pir("n", "Decimal code point.")}, func(a tools.Args) (any, error) {
			n := a.Int("n")
			if n < 0 || n > unicode.MaxRune || (n >= 0xD800 && n <= 0xDFFF) {
				return nil, fmt.Errorf("invalid code point %d", n)
			}
			return string(rune(n)), nil
		})
	reg("char_info", "Character info", "Describes the first character as a dict: char, codepoint (U+XXXX), decimal, utf8 (hex bytes), utf8_length, utf16 (hex code units), category (e.g. Lu), category_name, block, upper, lower and printable. Character names are not included.", "dict[str, str]", `text.char_info("é")`, 94, "character info unicode details category block utf8 utf16 lookup inspect",
		[]P{ps("ch", "A character; only the first one is used.")}, func(a tools.Args) (any, error) {
			r, err := firstRune(a.Str("ch"))
			if err != nil {
				return nil, err
			}
			u16 := utf16.Encode([]rune{r})
			parts := make([]string, len(u16))
			for i, u := range u16 {
				parts[i] = fmt.Sprintf("%04X", u)
			}
			u8 := []byte(string(r))
			cat := category(r)
			return map[string]string{
				"char": string(r), "codepoint": cpString(r), "decimal": strconv.Itoa(int(r)), "utf8": hexBytes(u8),
				"utf8_length": strconv.Itoa(len(u8)), "utf16": strings.Join(parts, " "), "category": cat, "category_name": catNames[cat],
				"block": blockOf(r), "upper": string(unicode.ToUpper(r)), "lower": string(unicode.ToLower(r)), "printable": bstr(unicode.IsPrint(r)),
			}, nil
		})
	reg("unicode_blocks", "Unicode blocks used", "Lists the distinct Unicode blocks (Basic Latin, Cyrillic, Emoticons ...) used by the text, in order of first appearance. Characters in blocks not in the built-in table are reported as Other.", "list[str]", `text.unicode_blocks("aб😀")`, 95, "unicode blocks scripts ranges used alphabet detect",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) {
			seen := map[string]bool{}
			out := []string{}
			for _, r := range a.Str("s") {
				b := blockOf(r)
				if !seen[b] {
					seen[b] = true
					out = append(out, b)
				}
			}
			return out, nil
		})

	reg("lorem", "Lorem ipsum words", "Generates lorem ipsum text of exactly the requested number of words (1 to 100000), as sentences. The same seed always gives the same text; seed 0 starts with 'Lorem ipsum dolor sit amet, consectetur adipiscing elit'.", "str", `text.lorem(20)`, 100, "lorem ipsum placeholder dummy filler text generate random words",
		[]P{pi("words", "Number of words.", 50), pi("seed", "Seed for deterministic output; 0 uses the classic opening.", 0)},
		func(a tools.Args) (any, error) {
			n := a.Int("words")
			if n < 1 || n > 100000 {
				return nil, errors.New("words must be between 1 and 100000")
			}
			return genLorem(int(n), a.Int("seed")), nil
		})
	reg("lorem_paragraphs", "Lorem ipsum paragraphs", "Generates lorem ipsum paragraphs (1 to 1000) of the given sentence count (1 to 100) separated by blank lines. The same seed gives the same text.", "str", `text.lorem_paragraphs(2, 4)`, 101, "lorem ipsum paragraphs placeholder dummy filler text generate",
		[]P{pi("paragraphs", "Number of paragraphs.", 3), pi("sentences", "Sentences per paragraph.", 5), pi("seed", "Seed for deterministic output.", 0)},
		func(a tools.Args) (any, error) {
			p, s := a.Int("paragraphs"), a.Int("sentences")
			if p < 1 || p > 1000 || s < 1 || s > 100 {
				return nil, errors.New("paragraphs must be 1-1000 and sentences 1-100")
			}
			g := newLorem(a.Int("seed"))
			paras := make([]string, p)
			for i := range paras {
				sents := make([]string, s)
				for j := range sents {
					sents[j] = g.sentence(6 + g.rng.IntN(9))
				}
				paras[i] = strings.Join(sents, " ")
			}
			return strings.Join(paras, "\n\n"), nil
		})
	reg("counter_string", "Counter string", "Builds a testing counter string of exactly the given length in which each marker is preceded by its own 1-based position, e.g. length 10 gives '*3*5*7*10*' (the marker's position is the number before it). Maximum 1000000.", "str", `text.counter_string(10)`, 102, "counter string test data length generate placeholder fixed length",
		[]P{pir("length", "Total length in characters."), pso("marker", "Single marker character.", "*")},
		func(a tools.Args) (any, error) {
			n := a.Int("length")
			m := a.Str("marker")
			if n < 0 || n > 1_000_000 {
				return nil, errors.New("length must be between 0 and 1000000")
			}
			if utf8.RuneCountInString(m) != 1 {
				return nil, errors.New("marker must be exactly one character")
			}
			// Build backwards: the marker at position p is preceded by the digits of p.
			var parts []string
			pos := int(n)
			for pos > 0 {
				d := strconv.Itoa(pos)
				if len(d)+1 > pos {
					parts = append(parts, strings.Repeat(m, pos))
					break
				}
				parts = append(parts, d+m)
				pos -= len(d) + 1
			}
			for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
				parts[i], parts[j] = parts[j], parts[i]
			}
			return strings.Join(parts, ""), nil
		})
}

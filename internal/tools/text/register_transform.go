package text

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
	"golang.org/x/text/unicode/norm"
)

func filterRunes(s string, keep func(rune) bool) string {
	return strings.Map(func(r rune) rune {
		if keep(r) {
			return r
		}
		return -1
	}, s)
}

func filt(name, title, doc, ex string, order int, tags string, keep func(rune) bool) {
	simple(name, title, doc, ex, order, tags, func(s string) string { return filterRunes(s, keep) })
}

func isPunctRune(r rune) bool { return unicode.IsPunct(r) }

func stripAccents(s string) string {
	var b strings.Builder
	for _, r := range norm.NFD.String(s) {
		if !unicode.Is(unicode.Mn, r) {
			b.WriteRune(r)
		}
	}
	return norm.NFC.String(b.String())
}

func isZeroWidth(r rune) bool {
	switch r {
	case 0x200B, 0x200C, 0x200D, 0x2060, 0xFEFF, 0x00AD, 0x180E, 0x200E, 0x200F, 0x061C, 0x034F:
		return true
	}
	return (r >= 0x202A && r <= 0x202E) || (r >= 0x2066 && r <= 0x2069)
}

func isEmojiRune(r rune) bool {
	switch {
	case r >= 0x1F000 && r <= 0x1FAFF, r >= 0x2600 && r <= 0x27BF, r >= 0xE0020 && r <= 0xE007F:
		return true
	case r == 0x2B50 || r == 0x2B55 || r == 0x2B1B || r == 0x2B1C || r == 0x231A || r == 0x231B:
		return true
	case r >= 0x23E9 && r <= 0x23F3, r >= 0x23F8 && r <= 0x23FA, r == 0x25FD || r == 0x25FE:
		return true
	}
	return false
}

func removeEmoji(s string) string {
	rs := []rune(s)
	var out []rune
	prevEmoji := false
	for i, r := range rs {
		switch {
		case isEmojiRune(r):
			prevEmoji = true
		case r == 0xFE0F || r == 0x20E3:
			// variation selector / keycap: dropped with their emoji
		case r == 0x200D && (prevEmoji || (i+1 < len(rs) && isEmojiRune(rs[i+1]))):
		case r >= 0x1F3FB && r <= 0x1F3FF:
		default:
			out = append(out, r)
			prevEmoji = false
		}
	}
	return string(out)
}

func showInvisibles(s string, spaces bool) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r == ' ':
			if spaces {
				b.WriteString("·")
			} else {
				b.WriteRune(r)
			}
		case unicode.IsControl(r) || isZeroWidth(r) || (unicode.Is(unicode.Zs, r)) || r == 0x2028 || r == 0x2029:
			fmt.Fprintf(&b, "<U+%04X>", r)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}

func smartToStraight(s string, punct bool) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case '‘', '’', '‚', '‛', '′':
			return '\''
		case '“', '”', '„', '‟', '″':
			return '"'
		case '–', '—', '−':
			if punct {
				return '-'
			}
		}
		return r
	}, expandPunct(s, punct))
}

func expandPunct(s string, punct bool) string {
	if !punct {
		return s
	}
	return strings.ReplaceAll(s, "…", "...")
}

func straightToSmart(s string) string {
	rs := []rune(s)
	var b strings.Builder
	for i, r := range rs {
		if r != '"' && r != '\'' {
			b.WriteRune(r)
			continue
		}
		opening := i == 0
		if i > 0 {
			p := rs[i-1]
			opening = unicode.IsSpace(p) || strings.ContainsRune("([{<—–-‘“", p)
		}
		switch {
		case r == '"' && opening:
			b.WriteRune('“')
		case r == '"':
			b.WriteRune('”')
		case opening && i+1 < len(rs) && !unicode.IsSpace(rs[i+1]):
			b.WriteRune('‘')
		default:
			b.WriteRune('’')
		}
	}
	return b.String()
}

func toFullwidth(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == ' ':
			return 0x3000
		case r >= 0x21 && r <= 0x7E:
			return r + 0xFEE0
		}
		return r
	}, s)
}

func toHalfwidth(s string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r == 0x3000:
			return ' '
		case r >= 0xFF01 && r <= 0xFF5E:
			return r - 0xFEE0
		}
		return r
	}, s)
}

var mirrorMap = func() map[rune]rune {
	m := map[rune]rune{}
	pairs := func(from, to string) {
		f, t := []rune(from), []rune(to)
		for i := range f {
			m[f[i]] = t[i]
		}
	}
	pairs("abcdefghijklmnopqrstuvwxyz", "ɐqɔpǝɟƃɥᴉɾʞlɯuodbɹsʇnʌʍxʎz")
	pairs("ABCDEFGHIJKLMNOPQRSTUVWXYZ", "∀qƆpƎℲפHIſʞ˥WNOԀQɹS┴∩ΛMX⅄Z")
	pairs("0123456789", "0ƖᄅƐㄣϛ9ㄥ86")
	pairs(".,'\"?!()[]{}<>_&;", "˙',"+"„¿¡)(][}{><‾⅋؛")
	return m
}()

func mirror(s string) string {
	g := graphemes(s)
	for i, j := 0, len(g)-1; i < j; i, j = i+1, j-1 {
		g[i], g[j] = g[j], g[i]
	}
	var b strings.Builder
	for _, x := range g {
		for _, r := range x {
			if m, ok := mirrorMap[r]; ok {
				r = m
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}

var leetMap = map[rune]rune{'a': '4', 'b': '8', 'e': '3', 'g': '6', 'i': '1', 'o': '0', 's': '5', 't': '7'}

func leet(s string) string {
	return strings.Map(func(r rune) rune {
		if m, ok := leetMap[unicode.ToLower(r)]; ok {
			return m
		}
		return r
	}, s)
}

// idx resolves a possibly negative rune index against n and clamps it.
func idx(i int64, n int) int {
	if i < 0 {
		i += int64(n)
	}
	if i < 0 {
		return 0
	}
	if i > int64(n) {
		return n
	}
	return int(i)
}

func padFill(fill string, n int) string {
	fr := []rune(fill)
	out := make([]rune, n)
	for i := range out {
		out[i] = fr[i%len(fr)]
	}
	return string(out)
}

func padTo(a tools.Args, mode string) (any, error) {
	s, fill, w := a.Str("s"), a.Str("fill"), a.Int("width")
	if fill == "" {
		return nil, errors.New("fill must not be empty")
	}
	if w < 0 || w > maxOut/4 {
		return nil, errors.New("width out of range")
	}
	n := int(w) - utf8.RuneCountInString(s)
	if n <= 0 {
		return s, nil
	}
	switch mode {
	case "left":
		return padFill(fill, n) + s, nil
	case "right":
		return s + padFill(fill, n), nil
	}
	l := n / 2
	return padFill(fill, l) + s + padFill(fill, n-l), nil
}

func wrapText(s string, width int, breakLong bool) string {
	var out []string
	for _, line := range strings.Split(normNL(s), "\n") {
		ws := strings.Fields(line)
		if len(ws) == 0 {
			out = append(out, "")
			continue
		}
		cur := ""
		curN := 0
		for _, w := range ws {
			wn := utf8.RuneCountInString(w)
			for breakLong && wn > width {
				if curN > 0 {
					out = append(out, cur)
					cur, curN = "", 0
				}
				rs := []rune(w)
				out = append(out, string(rs[:width]))
				w = string(rs[width:])
				wn -= width
			}
			switch {
			case curN == 0:
				cur, curN = w, wn
			case curN+1+wn <= width:
				cur += " " + w
				curN += 1 + wn
			default:
				out = append(out, cur)
				cur, curN = w, wn
			}
		}
		if curN > 0 {
			out = append(out, cur)
		}
	}
	return strings.Join(out, "\n")
}

func unindent(s string) string {
	lines := strings.Split(s, "\n")
	common := ""
	first := true
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			continue
		}
		ind := l[:len(l)-len(strings.TrimLeft(l, " \t"))]
		if first {
			common, first = ind, false
			continue
		}
		n := 0
		for n < len(common) && n < len(ind) && common[n] == ind[n] {
			n++
		}
		common = common[:n]
	}
	if common == "" {
		return s
	}
	for i, l := range lines {
		lines[i] = strings.TrimPrefix(l, common)
	}
	return strings.Join(lines, "\n")
}

func replaceCI(s, old, repl string, count int) string {
	if old == "" {
		return s
	}
	ls, lo := strings.ToLower(s), strings.ToLower(old)
	// Lower-casing can change byte lengths; fall back to rune-wise search then.
	if len(ls) != len(s) || len(lo) != len(old) {
		return replaceCIRunes(s, old, repl, count)
	}
	var b strings.Builder
	pos, n := 0, 0
	for count < 0 || n < count {
		i := strings.Index(ls[pos:], lo)
		if i < 0 {
			break
		}
		b.WriteString(s[pos : pos+i])
		b.WriteString(repl)
		pos += i + len(lo)
		n++
	}
	b.WriteString(s[pos:])
	return b.String()
}

func replaceCIRunes(s, old, repl string, count int) string {
	rs, ro := []rune(s), []rune(old)
	eq := func(i int) bool {
		for j, r := range ro {
			if unicode.ToLower(rs[i+j]) != unicode.ToLower(r) {
				return false
			}
		}
		return true
	}
	var b strings.Builder
	n := 0
	for i := 0; i < len(rs); {
		if i+len(ro) <= len(rs) && (count < 0 || n < count) && eq(i) {
			b.WriteString(repl)
			i += len(ro)
			n++
			continue
		}
		b.WriteRune(rs[i])
		i++
	}
	return b.String()
}

func init() {
	reg("reverse", "Reverse text", "Reverses the text by grapheme cluster, so combining marks, skin-tone modifiers, ZWJ emoji sequences and flags stay intact.", "str", `text.reverse("hello")`, 6, "reverse backwards flip mirror string",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) { return reverseGraphemes(a.Str("s")), nil })
	simple("reverse_words", "Reverse word order", "Reverses the order of whitespace-separated words and joins them with single spaces.", `text.reverse_words("one two three")`, 26, "reverse words order backwards sentence", func(s string) string {
		w := strings.Fields(s)
		for i, j := 0, len(w)-1; i < j; i, j = i+1, j-1 {
			w[i], w[j] = w[j], w[i]
		}
		return strings.Join(w, " ")
	})
	reg("sort_chars", "Sort characters", "Sorts the characters by code point (descending when desc is true). Whitespace is kept unless remove_space is true.", "str", `text.sort_chars("hello")`, 27, "sort characters letters alphabetical order",
		[]P{ps("s", "Input text."), pb("desc", "Descending order.", false), pb("remove_space", "Drop whitespace first.", false)},
		func(a tools.Args) (any, error) {
			s := a.Str("s")
			if a.Bool("remove_space") {
				s = filterRunes(s, func(r rune) bool { return !unicode.IsSpace(r) })
			}
			rs := []rune(s)
			sort.Slice(rs, func(i, j int) bool {
				if a.Bool("desc") {
					return rs[i] > rs[j]
				}
				return rs[i] < rs[j]
			})
			return string(rs), nil
		})
	reg("sort_words", "Sort words", "Sorts whitespace-separated words alphabetically and joins them with single spaces. Use lines.sort_lines for line-based sorting.", "str", `text.sort_words("pear apple fig")`, 7, "sort words alphabetical order a-z",
		[]P{ps("s", "Input text."), pb("desc", "Descending order.", false), pb("ignore_case", "Compare case-insensitively.", true)},
		func(a tools.Args) (any, error) {
			w := strings.Fields(a.Str("s"))
			key := func(x string) string {
				if a.Bool("ignore_case") {
					return strings.ToLower(x)
				}
				return x
			}
			sort.SliceStable(w, func(i, j int) bool {
				if a.Bool("desc") {
					return key(w[i]) > key(w[j])
				}
				return key(w[i]) < key(w[j])
			})
			return strings.Join(w, " "), nil
		})
	reg("trim", "Trim", "Removes characters from the ends. mode is both, left or right; cutset lists the characters to remove (empty means any whitespace).", "str", `text.trim("  hi  ")`, 8, "trim strip whitespace spaces ends lstrip rstrip",
		[]P{ps("s", "Input text."), pso("mode", "Which ends to trim.", "both", "both", "left", "right"), pso("cutset", "Characters to remove; empty means whitespace.", "")},
		func(a tools.Args) (any, error) {
			s, cs := a.Str("s"), a.Str("cutset")
			var left, right func(string) string
			if cs == "" {
				left = func(x string) string { return strings.TrimLeftFunc(x, unicode.IsSpace) }
				right = func(x string) string { return strings.TrimRightFunc(x, unicode.IsSpace) }
			} else {
				left = func(x string) string { return strings.TrimLeft(x, cs) }
				right = func(x string) string { return strings.TrimRight(x, cs) }
			}
			switch strings.ToLower(a.Str("mode")) {
			case "both", "":
				return right(left(s)), nil
			case "left":
				return left(s), nil
			case "right":
				return right(s), nil
			}
			return nil, fmt.Errorf("unknown mode %q (use both, left or right)", a.Str("mode"))
		})
	reg("collapse_whitespace", "Collapse whitespace", "Replaces every run of whitespace (spaces, tabs, newlines, Unicode spaces) with one space and trims both ends.", "str", `text.collapse_whitespace("a   b \n c")`, 9, "collapse squeeze whitespace multiple spaces normalize",
		[]P{ps("s", "Input text.")}, func(a tools.Args) (any, error) { return strings.Join(strings.Fields(a.Str("s")), " "), nil })
	filt("remove_whitespace", "Remove whitespace", "Deletes every whitespace character (spaces, tabs, newlines, Unicode spaces).", `text.remove_whitespace("a b\tc")`, 28, "remove strip delete whitespace spaces", func(r rune) bool { return !unicode.IsSpace(r) })
	reg("remove_punctuation", "Remove punctuation", "Deletes Unicode punctuation characters; with symbols=true also deletes symbols such as $ + < = ~.", "str", `text.remove_punctuation("Hi, there!")`, 29, "remove strip punctuation marks commas periods",
		[]P{ps("s", "Input text."), pb("symbols", "Also remove symbol characters.", false)},
		func(a tools.Args) (any, error) {
			sym := a.Bool("symbols")
			return filterRunes(a.Str("s"), func(r rune) bool { return !(unicode.IsPunct(r) || (sym && unicode.IsSymbol(r))) }), nil
		})
	filt("remove_digits", "Remove digits", "Deletes every Unicode decimal digit.", `text.remove_digits("a1b2")`, 30, "remove strip digits numbers", func(r rune) bool { return !unicode.IsDigit(r) })
	filt("only_digits", "Keep only digits", "Keeps only Unicode decimal digits and deletes everything else.", `text.only_digits("tel: 555-1234")`, 31, "only digits numbers extract keep", unicode.IsDigit)
	filt("only_letters", "Keep only letters", "Keeps only letters (and their combining marks) and deletes everything else.", `text.only_letters("a1 b2!")`, 32, "only letters alphabetic extract keep", func(r rune) bool { return unicode.IsLetter(r) || isMark(r) })
	filt("only_alnum", "Keep only letters and digits", "Keeps only letters, combining marks and digits.", `text.only_alnum("a-1 b_2!")`, 33, "only alphanumeric letters digits extract keep", func(r rune) bool { return isWordCore(r) })
	filt("remove_non_ascii", "Remove non-ASCII", "Deletes every character with a code point of 128 or more.", `text.remove_non_ascii("héllo ✓")`, 34, "remove non-ascii unicode strip accents emoji", func(r rune) bool { return r < 128 })
	simple("strip_accents", "Strip accents", "Removes diacritical marks (é -> e, ñ -> n) via NFD decomposition and dropping non-spacing marks. Letters without decomposition (ø, ß, ł) stay unchanged; use slug for transliteration.", `text.strip_accents("Crème brûlée")`, 35, "accents diacritics remove fold ascii normalize deburr", stripAccents)
	reg("normalize", "Unicode normalize", "Applies Unicode normalization form NFC, NFD, NFKC or NFKD (case-insensitive name).", "str", `text.normalize("é", "NFC")`, 36, "unicode normalize nfc nfd nfkc nfkd compose decompose",
		[]P{ps("s", "Input text."), pso("form", "Normalization form.", "NFC", "NFC", "NFD", "NFKC", "NFKD")},
		func(a tools.Args) (any, error) {
			switch strings.ToUpper(a.Str("form")) {
			case "NFC", "":
				return norm.NFC.String(a.Str("s")), nil
			case "NFD":
				return norm.NFD.String(a.Str("s")), nil
			case "NFKC":
				return norm.NFKC.String(a.Str("s")), nil
			case "NFKD":
				return norm.NFKD.String(a.Str("s")), nil
			}
			return nil, fmt.Errorf("unknown form %q (use NFC, NFD, NFKC or NFKD)", a.Str("form"))
		})
	reg("repeat", "Repeat text", "Repeats the text n times with an optional separator between copies. n of 0 gives an empty string; negative n or a result over 4 MiB is an error.", "str", `text.repeat("ab", 3, "-")`, 37, "repeat duplicate times multiply n",
		[]P{ps("s", "Text to repeat."), pi("n", "Number of copies.", 2), pso("sep", "Separator between copies.", "")},
		func(a tools.Args) (any, error) {
			n := a.Int("n")
			if n < 0 {
				return nil, errors.New("n must not be negative")
			}
			if n == 0 {
				return "", nil
			}
			per := int64(len(a.Str("s")) + len(a.Str("sep")))
			if per > 0 && n > int64(maxOut)/per+1 {
				return nil, fmt.Errorf("result would exceed %d MiB", maxOut>>20)
			}
			parts := make([]string, n)
			for i := range parts {
				parts[i] = a.Str("s")
			}
			return strings.Join(parts, a.Str("sep")), nil
		})
	padParams := []P{ps("s", "Input text."), pir("width", "Target width in characters."), pso("fill", "Fill text, cycled as needed.", " ")}
	for _, m := range []struct{ name, title, doc, ex, mode string }{
		{"pad_left", "Pad left", "Pads the start with the fill text up to width characters (right-aligns). Text already that wide is unchanged.", `text.pad_left("7", 3, "0")`, "left"},
		{"pad_right", "Pad right", "Pads the end with the fill text up to width characters (left-aligns). Text already that wide is unchanged.", `text.pad_right("7", 3, ".")`, "right"},
		{"center", "Center text", "Centers the text in width characters using the fill text; an odd leftover puts the extra fill on the right.", `text.center("hi", 6, "*")`, "center"},
	} {
		m := m
		reg(m.name, m.title, m.doc, "str", m.ex, 38, "pad padding fill width align justify "+m.mode, padParams[:3:3],
			func(a tools.Args) (any, error) { return padTo(a, m.mode) })
	}
	reg("truncate", "Truncate", "Shortens the text to at most max characters (runes), ending with the ellipsis when cut. If max is smaller than the ellipsis, only the first max characters are returned.", "str", `text.truncate("Hello world", 8)`, 39, "truncate shorten cut ellipsis limit length",
		[]P{ps("s", "Input text."), pi("max", "Maximum length in characters, including the ellipsis.", 20), pso("ellipsis", "Marker appended when cut.", "…")},
		func(a tools.Args) (any, error) {
			m := a.Int("max")
			if m < 0 {
				return nil, errors.New("max must not be negative")
			}
			rs := []rune(a.Str("s"))
			if int64(len(rs)) <= m {
				return a.Str("s"), nil
			}
			el := []rune(a.Str("ellipsis"))
			if int64(len(el)) >= m {
				return string(rs[:m]), nil
			}
			return string(rs[:m-int64(len(el))]) + string(el), nil
		})
	reg("truncate_words", "Truncate words", "Keeps the first max_words whitespace-separated words and appends the ellipsis if any were dropped; shorter text is returned unchanged.", "str", `text.truncate_words("one two three four", 2)`, 40, "truncate words shorten excerpt summary",
		[]P{ps("s", "Input text."), pi("max_words", "Words to keep.", 10), pso("ellipsis", "Marker appended when cut.", "…")},
		func(a tools.Args) (any, error) {
			if a.Int("max_words") < 0 {
				return nil, errors.New("max_words must not be negative")
			}
			w := strings.Fields(a.Str("s"))
			if int64(len(w)) <= a.Int("max_words") {
				return a.Str("s"), nil
			}
			return strings.Join(w[:a.Int("max_words")], " ") + a.Str("ellipsis"), nil
		})
	reg("wrap", "Word wrap", "Greedy word wrap to width characters per line. Existing line breaks start new paragraphs, runs of spaces collapse, and words longer than width stay whole unless break_long is true.", "str", `text.wrap("the quick brown fox", 10)`, 41, "wrap word wrap hard wrap column width fill reflow",
		[]P{ps("s", "Input text."), pi("width", "Maximum line width in characters.", 72), pb("break_long", "Split words longer than width.", false)},
		func(a tools.Args) (any, error) {
			if a.Int("width") < 1 {
				return nil, errors.New("width must be at least 1")
			}
			return wrapText(a.Str("s"), int(a.Int("width")), a.Bool("break_long")), nil
		})
	reg("indent", "Indent text", "Prefixes every non-empty line with the given text (four spaces by default). Blank lines are left empty.", "str", `text.indent("a\nb", "  ")`, 42, "indent prefix lines spaces tab",
		[]P{ps("s", "Input text."), pso("prefix", "Text added to each non-empty line.", "    ")},
		func(a tools.Args) (any, error) {
			lines := strings.Split(a.Str("s"), "\n")
			for i, l := range lines {
				if strings.TrimSpace(l) != "" {
					lines[i] = a.Str("prefix") + l
				}
			}
			out := strings.Join(lines, "\n")
			return out, checkSize(len(out))
		})
	simple("unindent", "Unindent text", "Removes the leading whitespace shared by all non-blank lines (dedent). Blank lines are ignored when measuring.", "text.unindent(\"  a\\n    b\")", 43, "unindent dedent outdent remove indentation", unindent)
	reg("replace", "Replace text", "Replaces occurrences of old with new, at most count times (-1 = all). An empty old returns the text unchanged.", "str", `text.replace("a-b-c", "-", "+")`, 44, "replace substitute find replace swap substring",
		[]P{ps("s", "Input text."), ps("old", "Text to find."), pso("new", "Replacement text.", ""), pi("count", "Maximum replacements; -1 for all.", -1)},
		func(a tools.Args) (any, error) {
			if a.Str("old") == "" {
				return a.Str("s"), nil
			}
			out := strings.Replace(a.Str("s"), a.Str("old"), a.Str("new"), int(a.Int("count")))
			return out, checkSize(len(out))
		})
	reg("replace_ci", "Replace (ignore case)", "Like replace but matches old case-insensitively; the replacement is inserted as given. An empty old returns the text unchanged.", "str", `text.replace_ci("Hello HELLO", "hello", "bye")`, 45, "replace case insensitive ignore case substitute",
		[]P{ps("s", "Input text."), ps("old", "Text to find (any case)."), pso("new", "Replacement text.", ""), pi("count", "Maximum replacements; -1 for all.", -1)},
		func(a tools.Args) (any, error) {
			out := replaceCI(a.Str("s"), a.Str("old"), a.Str("new"), int(a.Int("count")))
			return out, checkSize(len(out))
		})
	reg("remove", "Remove substring", "Deletes every occurrence of sub. An empty sub returns the text unchanged.", "str", `text.remove("banana", "an")`, 46, "remove delete substring strip occurrences",
		[]P{ps("s", "Input text."), ps("sub", "Text to delete.")},
		func(a tools.Args) (any, error) {
			if a.Str("sub") == "" {
				return a.Str("s"), nil
			}
			return strings.ReplaceAll(a.Str("s"), a.Str("sub"), ""), nil
		})
	reg("insert", "Insert text", "Inserts text at a character index (negative counts from the end; out-of-range values clamp to the ends).", "str", `text.insert("held", 3, "l")`, 47, "insert add text at position index",
		[]P{ps("s", "Input text."), pir("index", "Character index."), ps("text", "Text to insert.")},
		func(a tools.Args) (any, error) {
			rs := []rune(a.Str("s"))
			i := idx(a.Int("index"), len(rs))
			out := string(rs[:i]) + a.Str("text") + string(rs[i:])
			return out, checkSize(len(out))
		})
	reg("delete_range", "Delete range", "Deletes characters from start up to but excluding end (character indexes, negative from the end, clamped). start >= end deletes nothing.", "str", `text.delete_range("hello", 1, 3)`, 48, "delete range remove slice characters between",
		[]P{ps("s", "Input text."), pir("start", "First character index to delete."), pir("end", "Index after the last deleted character.")},
		func(a tools.Args) (any, error) {
			rs := []rune(a.Str("s"))
			st, en := idx(a.Int("start"), len(rs)), idx(a.Int("end"), len(rs))
			if st >= en {
				return a.Str("s"), nil
			}
			return string(rs[:st]) + string(rs[en:]), nil
		})
	reg("substring", "Substring", "Returns length characters starting at start (character-based; negative start counts from the end; length -1 means to the end). Out-of-range values clamp.", "str", `text.substring("héllo wörld", 6, 5)`, 49, "substring substr slice cut extract part",
		[]P{ps("s", "Input text."), pi("start", "Start index; negative counts from the end.", 0), pi("length", "Number of characters; -1 for the rest.", -1)},
		func(a tools.Args) (any, error) {
			rs := []rune(a.Str("s"))
			st := idx(a.Int("start"), len(rs))
			en := len(rs)
			if a.Int("length") >= 0 && a.Int("length") < int64(len(rs)-st) {
				en = st + int(a.Int("length"))
			}
			return string(rs[st:en]), nil
		})
	reg("split_at", "Split at index", "Splits the text into two parts at a character index (negative from the end, clamped). Always returns exactly two strings.", "list[str]", `text.split_at("hello", 2)`, 50, "split at index position two parts halves",
		[]P{ps("s", "Input text."), pir("index", "Character index.")},
		func(a tools.Args) (any, error) {
			rs := []rune(a.Str("s"))
			i := idx(a.Int("index"), len(rs))
			return []string{string(rs[:i]), string(rs[i:])}, nil
		})
	reg("left", "Left characters", "Returns the first n characters (all of them if n is larger; n <= 0 gives empty).", "str", `text.left("hello", 2)`, 51, "left first characters head prefix take",
		[]P{ps("s", "Input text."), pi("n", "Number of characters.", 1)},
		func(a tools.Args) (any, error) {
			rs := []rune(a.Str("s"))
			return string(rs[:idx(max(a.Int("n"), 0), len(rs))]), nil
		})
	reg("right", "Right characters", "Returns the last n characters (all of them if n is larger; n <= 0 gives empty).", "str", `text.right("hello", 2)`, 52, "right last characters tail suffix take",
		[]P{ps("s", "Input text."), pi("n", "Number of characters.", 1)},
		func(a tools.Args) (any, error) {
			rs := []rune(a.Str("s"))
			return string(rs[len(rs)-idx(max(a.Int("n"), 0), len(rs)):]), nil
		})
	reg("remove_prefix", "Remove prefix", "Removes the prefix when the text starts with it; otherwise returns the text unchanged.", "str", `text.remove_prefix("foobar", "foo")`, 53, "remove prefix strip start trimprefix",
		[]P{ps("s", "Input text."), ps("prefix", "Prefix to remove.")},
		func(a tools.Args) (any, error) { return strings.TrimPrefix(a.Str("s"), a.Str("prefix")), nil })
	reg("remove_suffix", "Remove suffix", "Removes the suffix when the text ends with it; otherwise returns the text unchanged.", "str", `text.remove_suffix("foobar", "bar")`, 54, "remove suffix strip end trimsuffix extension",
		[]P{ps("s", "Input text."), ps("suffix", "Suffix to remove.")},
		func(a tools.Args) (any, error) { return strings.TrimSuffix(a.Str("s"), a.Str("suffix")), nil })
	simple("leetspeak", "Leetspeak", "Converts to leetspeak with a fixed map: a4 b8 e3 g6 i1 o0 s5 t7 (case-insensitive); other characters are unchanged.", `text.leetspeak("elite hacker")`, 55, "leet 1337 hacker speak l33t", leet)
	simple("mirror", "Upside-down text", "Flips the text upside down: reverses it and swaps letters, digits and brackets for look-alike rotated characters (a -> ɐ, b -> q, ? -> ¿). Characters without a mapping stay as they are.", `text.mirror("Hello")`, 56, "upside down flip mirror text rotate 180 reverse unicode fancy", mirror)
	filt("remove_zero_width", "Remove zero-width characters", "Deletes invisible characters: zero-width space/non-joiner/joiner, word joiner, BOM, soft hyphen, bidi marks and overrides, Mongolian vowel separator, combining grapheme joiner. Note that ZWNJ/ZWJ are meaningful in some scripts and emoji sequences.", `text.remove_zero_width("a​b")`, 57, "zero width invisible remove zwsp bom soft hyphen hidden characters", func(r rune) bool { return !isZeroWidth(r) })
	reg("show_invisibles", "Show invisible characters", "Makes invisible characters visible: \\n, \\r and \\t as escapes; other control, zero-width and special space characters (including no-break space) as <U+XXXX>; regular spaces become · only when show_spaces is true.", "str", `text.show_invisibles("a\tb​")`, 58, "show invisible hidden whitespace control characters debug escape reveal",
		[]P{ps("s", "Input text."), pb("show_spaces", "Render ordinary spaces as ·.", false)},
		func(a tools.Args) (any, error) { return showInvisibles(a.Str("s"), a.Bool("show_spaces")), nil })
	reg("smart_quotes_to_straight", "Smart quotes to straight", "Replaces curly quotes (‘ ’ “ ” „ ‚ and primes) with ASCII ' and \"; with punctuation=true also converts en/em dashes to - and … to ...", "str", `text.smart_quotes_to_straight("“hi” it’s")`, 59, "smart quotes curly straight ascii typographic replace dumb",
		[]P{ps("s", "Input text."), pb("punctuation", "Also convert dashes and ellipsis.", false)},
		func(a tools.Args) (any, error) { return smartToStraight(a.Str("s"), a.Bool("punctuation")), nil })
	simple("straight_to_smart", "Straight quotes to smart", "Replaces ASCII quotes with typographic ones: opening after the start, whitespace or an opening bracket, closing otherwise; a lone apostrophe inside a word becomes ’.", `text.straight_to_smart("\"hi\" it's")`, 60, "smart quotes curly typographic convert straight apostrophe", straightToSmart)
	simple("remove_emoji", "Remove emoji", "Deletes emoji and pictographs (including skin-tone modifiers, flags, variation selectors and the ZWJ glue between emoji). Covers the main emoji blocks, not every symbol.", `text.remove_emoji("hi 😀 there 👍🏽")`, 61, "emoji remove strip pictographs smileys unicode", removeEmoji)
	simple("to_fullwidth", "Fullwidth characters", "Converts ASCII printable characters to their fullwidth forms (A -> Ａ, space -> U+3000).", `text.to_fullwidth("Hello 123")`, 62, "fullwidth wide aesthetic vaporwave cjk zenkaku", toFullwidth)
	simple("to_halfwidth", "Halfwidth characters", "Converts fullwidth ASCII forms (U+FF01-FF5E and U+3000) back to normal ASCII. Halfwidth katakana is not converted.", `text.to_halfwidth("Ｈｅｌｌｏ")`, 63, "halfwidth narrow ascii hankaku normalize", toHalfwidth)
}

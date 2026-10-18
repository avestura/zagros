package num

import (
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

// ---- Roman numerals ----

var romanVals = []struct {
	v int
	s string
}{{1000, "M"}, {900, "CM"}, {500, "D"}, {400, "CD"}, {100, "C"}, {90, "XC"}, {50, "L"}, {40, "XL"}, {10, "X"}, {9, "IX"}, {5, "V"}, {4, "IV"}, {1, "I"}}

func romanEncode(n int) string {
	var b strings.Builder
	for _, r := range romanVals {
		for n >= r.v {
			b.WriteString(r.s)
			n -= r.v
		}
	}
	return b.String()
}

func romanDecode(s string) (int, error) {
	s = strings.ToUpper(strings.TrimSpace(s))
	if s == "" {
		return 0, errors.New("empty roman numeral")
	}
	if len(s) > 30 {
		return 0, errors.New("roman numeral is too long")
	}
	val := map[byte]int{'I': 1, 'V': 5, 'X': 10, 'L': 50, 'C': 100, 'D': 500, 'M': 1000}
	total := 0
	for i := 0; i < len(s); i++ {
		v, ok := val[s[i]]
		if !ok {
			return 0, fmt.Errorf("invalid roman numeral character %q at position %d", string(s[i]), i+1)
		}
		if i+1 < len(s) && val[s[i+1]] > v {
			total -= v
		} else {
			total += v
		}
	}
	if total < 1 || total > 3999 {
		return 0, fmt.Errorf("roman numeral %s is outside 1..3999", s)
	}
	return total, nil
}

// ---- number words ----

var (
	unitWords = []string{"zero", "one", "two", "three", "four", "five", "six", "seven", "eight", "nine", "ten", "eleven", "twelve", "thirteen", "fourteen", "fifteen", "sixteen", "seventeen", "eighteen", "nineteen"}
	tensWords = []string{"", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"}
	scaleWord = []string{"", "thousand", "million", "billion", "trillion", "quadrillion", "quintillion", "sextillion", "septillion", "octillion", "nonillion", "decillion", "undecillion", "duodecillion"}
)

var maxWords = new(big.Int).Exp(big.NewInt(10), big.NewInt(39), nil)

func below1000(n int) string {
	var parts []string
	if n >= 100 {
		parts = append(parts, unitWords[n/100], "hundred")
		n %= 100
	}
	if n >= 20 {
		w := tensWords[n/10]
		if n%10 != 0 {
			w += "-" + unitWords[n%10]
		}
		parts = append(parts, w)
	} else if n > 0 {
		parts = append(parts, unitWords[n])
	}
	return strings.Join(parts, " ")
}

func numberWords(n *big.Int) (string, error) {
	if new(big.Int).Abs(n).Cmp(maxWords) >= 0 {
		return "", errors.New("number must be below 10^39 in absolute value")
	}
	if n.Sign() == 0 {
		return "zero", nil
	}
	a := new(big.Int).Abs(n)
	var groups []int
	k, m := new(big.Int), big.NewInt(1000)
	for a.Sign() > 0 {
		a.DivMod(a, m, k)
		groups = append(groups, int(k.Int64()))
	}
	var parts []string
	if n.Sign() < 0 {
		parts = append(parts, "minus")
	}
	for i := len(groups) - 1; i >= 0; i-- {
		if groups[i] == 0 {
			continue
		}
		parts = append(parts, below1000(groups[i]))
		if i > 0 {
			parts = append(parts, scaleWord[i])
		}
	}
	return strings.Join(parts, " "), nil
}

var ordSpecial = map[string]string{"one": "first", "two": "second", "three": "third", "five": "fifth", "eight": "eighth", "nine": "ninth", "twelve": "twelfth", "zero": "zeroth"}

func ordinalWords(s string) string {
	i := strings.LastIndexAny(s, " -")
	head, last := s[:i+1], s[i+1:]
	switch {
	case ordSpecial[last] != "":
		last = ordSpecial[last]
	case strings.HasSuffix(last, "y"):
		last = last[:len(last)-1] + "ieth"
	default:
		last += "th"
	}
	return head + last
}

var wordValue = func() map[string]int {
	m := map[string]int{}
	for i, w := range unitWords {
		m[w] = i
	}
	for i, w := range tensWords {
		if w != "" {
			m[w] = i * 10
		}
	}
	m["fourty"] = 40
	for c, o := range ordSpecial {
		m[o] = m[c]
	}
	for i, w := range unitWords {
		if _, ok := ordSpecial[w]; !ok && w != "zero" {
			m[w+"th"] = i
		}
	}
	for i, w := range tensWords {
		if w != "" {
			m[w[:len(w)-1]+"ieth"] = i * 10
		}
	}
	return m
}()

func wordsToNumber(s string) (*big.Int, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return nil, errors.New("empty text")
	}
	if len(s) > 5000 {
		return nil, errors.New("text is too long")
	}
	fields := strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == '-' || r == ',' || r == '\t' })
	neg := false
	total, cur := new(big.Int), 0
	seen := false
	scale := map[string]int{}
	for i, w := range scaleWord {
		if w != "" {
			scale[w] = i
			scale[w+"th"] = i
		}
	}
	for i, w := range fields {
		switch {
		case (w == "minus" || w == "negative") && i == 0:
			neg = true
		case w == "and" || w == "a" && i == 0:
			if w == "a" {
				cur = 1
			}
		case w == "hundred" || w == "hundredth":
			if cur == 0 {
				cur = 1
			}
			cur *= 100
			seen = true
		case scale[w] > 0:
			if cur == 0 {
				cur = 1
			}
			add := new(big.Int).Mul(big.NewInt(int64(cur)), new(big.Int).Exp(big.NewInt(1000), big.NewInt(int64(scale[w])), nil))
			total.Add(total, add)
			cur = 0
			seen = true
		default:
			v, ok := wordValue[w]
			if !ok {
				return nil, fmt.Errorf("unknown number word %q", w)
			}
			cur += v
			seen = true
		}
	}
	if !seen {
		return nil, errors.New("no number words found")
	}
	total.Add(total, big.NewInt(int64(cur)))
	if neg {
		total.Neg(total)
	}
	return total, nil
}

// ---- separators and formatting ----

var plainNumRe = regexp.MustCompile(`^([+-]?)(\d+)(\.\d*)?$`)

func addSeparators(s, sep string, size int) (string, error) {
	t := strings.NewReplacer(",", "", "_", "", " ", "", " ", "", "'", "").Replace(strings.TrimSpace(s))
	m := plainNumRe.FindStringSubmatch(t)
	if m == nil {
		return "", fmt.Errorf("%q is not a plain decimal number", strings.TrimSpace(s))
	}
	return m[1] + groupDigits(m[2], size, sep) + m[3], nil
}

func sciFormat(r *big.Rat, digits int) string {
	prec := uint(digits*4 + 128)
	f := new(big.Float).SetPrec(prec).SetRat(r)
	s := f.Text('e', digits)
	i := strings.IndexByte(s, 'e')
	exp := 0
	fmt.Sscanf(s[i+1:], "%d", &exp)
	sign := "+"
	if exp < 0 {
		sign, exp = "-", -exp
	}
	return fmt.Sprintf("%se%s%d", s[:i], sign, exp)
}

func formatNumber(r *big.Rat, decimals int, thousands, point string) string {
	s := r.FloatString(decimals)
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	intPart, frac := s, ""
	if i := strings.IndexByte(s, '.'); i >= 0 {
		intPart, frac = s[:i], s[i+1:]
	}
	if strings.Trim(intPart+frac, "0") == "" {
		neg = false
	}
	out := groupDigits(intPart, 3, thousands)
	if frac != "" {
		out += point + frac
	}
	if neg {
		out = "-" + out
	}
	return out
}

var suffixMul = map[string]int64{
	"": 1, "k": 1e3, "thousand": 1e3, "m": 1e6, "mm": 1e6, "mil": 1e6, "million": 1e6,
	"b": 1e9, "bn": 1e9, "g": 1e9, "billion": 1e9, "t": 1e12, "tn": 1e12, "trillion": 1e12,
	"q": 1e15, "quadrillion": 1e15,
}

var humanRe = regexp.MustCompile(`^([+-]?(?:\d[\d,_ ]*\.?\d*|\.\d+)(?:[eE][+-]?\d{1,4})?)\s*([A-Za-z%]*)$`)

func parseHuman(s string) (*big.Rat, error) {
	t := strings.TrimSpace(strings.Trim(strings.TrimSpace(s), "$€£¥"))
	m := humanRe.FindStringSubmatch(t)
	if m == nil {
		return nil, fmt.Errorf("%q is not a number like 1.5k, 2M, 3e6 or 1,234", strings.TrimSpace(s))
	}
	r, err := parseRat(m[1])
	if err != nil {
		return nil, err
	}
	suf := strings.ToLower(m[2])
	if suf == "%" {
		return r.Quo(r, big.NewRat(100, 1)), nil
	}
	mul, ok := suffixMul[suf]
	if !ok {
		return nil, fmt.Errorf("unknown suffix %q (use k, M, B, T)", m[2])
	}
	return r.Mul(r, new(big.Rat).SetInt64(mul)), nil
}

var (
	decUnits = []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	binUnits = []string{"B", "KiB", "MiB", "GiB", "TiB", "PiB", "EiB", "ZiB", "YiB"}
)

func humanBytes(r *big.Rat, binary bool, decimals int) string {
	base, units := int64(1000), decUnits
	if binary {
		base, units = 1024, binUnits
	}
	neg := r.Sign() < 0
	v := new(big.Rat).Abs(r)
	b := new(big.Rat).SetInt64(base)
	i := 0
	for i < len(units)-1 && v.Cmp(b) >= 0 {
		v.Quo(v, b)
		i++
	}
	s := ratString(roundRat(v, decimals), decimals)
	if rv := roundRat(v, decimals); rv.Cmp(b) >= 0 && i < len(units)-1 {
		i++
		s = ratString(roundRat(new(big.Rat).Quo(rv, b), decimals), decimals)
	}
	if neg && s != "0" {
		s = "-" + s
	}
	return s + " " + units[i]
}

func roundRat(r *big.Rat, decimals int) *big.Rat {
	out, _ := new(big.Rat).SetString(r.FloatString(decimals))
	return out
}

var bytesRe = regexp.MustCompile(`^([+-]?(?:\d[\d,_ ]*\.?\d*|\.\d+)(?:[eE][+-]?\d{1,3})?)\s*([A-Za-z]*)$`)

func parseBytes(s string, binary bool) (*big.Int, error) {
	m := bytesRe.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return nil, fmt.Errorf("%q is not a size like 1.5 GiB or 200MB", strings.TrimSpace(s))
	}
	r, err := parseRat(m[1])
	if err != nil {
		return nil, err
	}
	u := strings.ToLower(m[2])
	u = strings.TrimSuffix(strings.TrimSuffix(u, "bytes"), "byte")
	if u == "" || u == "b" {
		return ratToIntRound(r)
	}
	exp := strings.Index("kmgtpezy", u[:1])
	if exp < 0 {
		return nil, fmt.Errorf("unknown unit %q", m[2])
	}
	rest := u[1:]
	base := int64(1000)
	switch rest {
	case "", "b":
		if binary {
			base = 1024
		}
	case "ib":
		base = 1024
	default:
		return nil, fmt.Errorf("unknown unit %q", m[2])
	}
	mult := new(big.Int).Exp(big.NewInt(base), big.NewInt(int64(exp+1)), nil)
	return ratToIntRound(r.Mul(r, new(big.Rat).SetInt(mult)))
}

func ratToIntRound(r *big.Rat) (*big.Int, error) {
	out, _ := new(big.Rat).SetString(r.FloatString(0))
	return new(big.Int).Set(out.Num()), nil
}

func init() {
	reg("roman_encode", "Roman numeral", "Converts a whole number 1..3999 to a Roman numeral using standard subtractive notation (1994 gives MCMXCIV). Numbers outside 1..3999 are an error because overline (vinculum) notation is not supported.",
		"str", `num.roman_encode("1994")`, 3, "roman numeral encode number to roman I V X L C M",
		[]P{rs("value", "Decimal integer 1..3999.")},
		func(a tools.Args) (any, error) {
			n, err := argSmall(a, "value", 1, 3999)
			if err != nil {
				return nil, err
			}
			return romanEncode(int(n)), nil
		}, kw("roman"), accepts(func(s string) bool {
			n, err := dec(s)
			return err == nil && n.IsInt64() && n.Int64() >= 1 && n.Int64() <= 3999
		}))

	reg("roman_decode", "Roman to number", "Converts a Roman numeral (case-insensitive, subtractive notation) to a decimal string; MCMXCIV gives 1994. Invalid letters or values outside 1..3999 are errors.",
		"str", `num.roman_decode("MCMXCIV")`, 4, "roman numeral decode roman to number I V X L C M",
		[]P{rs("value", "Roman numeral.")},
		func(a tools.Args) (any, error) {
			n, err := romanDecode(a.Str("value"))
			if err != nil {
				return nil, err
			}
			return fmt.Sprint(n), nil
		}, kw("roman"), accepts(func(s string) bool {
			s = strings.TrimSpace(s)
			if s == "" || len(s) > 30 {
				return false
			}
			_, err := romanDecode(s)
			return err == nil
		}))

	reg("number_to_words", "Number to words", "Spells a whole number as English words (1234 gives \"one thousand two hundred thirty-four\"). Supports negatives (\"minus\") and any size below 10^39 (up to undecillions); pass the number as a decimal string. Set ordinal for ordinal words (\"first\", \"twenty-first\").",
		"str", `num.number_to_words("1234")`, 5, "number words spell out english say text name cardinal ordinal check writing",
		[]P{rs("value", "Integer, decimal string."), os("lang", "Language; only \"en\" is supported.", "en", "en"), ob("ordinal", "Return ordinal words (first, second, ...).", false)},
		func(a tools.Args) (any, error) {
			if l := strings.ToLower(a.Str("lang")); l != "en" && l != "" {
				return nil, fmt.Errorf("unsupported language %q (only en)", a.Str("lang"))
			}
			n, err := argBig(a, "value", 10)
			if err != nil {
				return nil, err
			}
			w, err := numberWords(n)
			if err != nil {
				return nil, err
			}
			if a.Bool("ordinal") {
				w = ordinalWords(w)
			}
			return w, nil
		}, kw("words", "spell"), accepts(parsesDec))

	reg("ordinal_words", "Ordinal words", "Spells a whole number as English ordinal words (1 gives first, 21 gives twenty-first, 100 gives one hundredth). Any size below 10^39.",
		"str", `num.ordinal_words("21")`, 6, "ordinal words first second third spell out",
		[]P{rs("value", "Integer, decimal string.")},
		func(a tools.Args) (any, error) {
			n, err := argBig(a, "value", 10)
			if err != nil {
				return nil, err
			}
			w, err := numberWords(n)
			if err != nil {
				return nil, err
			}
			return ordinalWords(w), nil
		}, accepts(parsesDec))

	reg("words_to_number", "Words to number", "Parses English number words (\"two hundred and five\", \"minus twenty-one\", \"three million four\", ordinals like \"first\") into a decimal string. Unknown words are errors.",
		"str", `num.words_to_number("two hundred and five")`, 7, "words to number parse english text spelled out digits",
		[]P{rs("text", "English number words.")},
		func(a tools.Args) (any, error) {
			n, err := wordsToNumber(a.Str("text"))
			if err != nil {
				return nil, err
			}
			return n.String(), nil
		}, kw("words"), accepts(func(s string) bool {
			if len(s) > 300 || !strings.ContainsAny(s, "aeiou") {
				return false
			}
			_, err := wordsToNumber(s)
			return err == nil
		}))

	reg("ordinal", "Ordinal suffix", "Adds the English ordinal suffix to an integer: 1st, 2nd, 3rd, 4th, 11th, 12th, 13th, 21st, 101st; negatives keep their sign (-1st).",
		"str", `num.ordinal("22")`, 8, "ordinal suffix 1st 2nd 3rd 4th th st nd rd",
		[]P{rs("value", "Integer, decimal string.")},
		func(a tools.Args) (any, error) {
			n, err := argBig(a, "value", 10)
			if err != nil {
				return nil, err
			}
			abs := new(big.Int).Abs(n)
			last2 := int(new(big.Int).Mod(abs, big.NewInt(100)).Int64())
			suf := "th"
			if last2 < 11 || last2 > 13 {
				switch last2 % 10 {
				case 1:
					suf = "st"
				case 2:
					suf = "nd"
				case 3:
					suf = "rd"
				}
			}
			return n.String() + suf, nil
		}, accepts(parsesDec))

	reg("add_thousands_separators", "Add thousands separators", "Groups the integer part of a plain decimal number into threes with a separator: 1234567.891 becomes 1,234,567.891. Existing commas, spaces and underscores are removed first; the fraction is untouched.",
		"str", `num.add_thousands_separators("1234567.891")`, 9, "thousands separator comma group digits format number commas",
		[]P{rs("value", "Plain decimal number."), os("sep", "Separator text.", ","), oi("group", "Digits per group.", 3)},
		func(a tools.Args) (any, error) {
			g, err := intIn("group", a.Int("group"), 1, 100)
			if err != nil {
				return nil, err
			}
			return addSeparators(a.Str("value"), a.Str("sep"), int(g))
		}, kw("commas", "thousands"), accepts(func(s string) bool { _, err := addSeparators(s, ",", 3); return err == nil && len(s) > 3 }))

	reg("strip_separators", "Strip separators", "Normalises a formatted number to plain digits with a dot decimal point by removing thousands marks (commas, spaces, underscores, apostrophes, non-breaking spaces). With decimal_point=\",\" the dots are removed and the comma becomes the decimal point (1.234,56 gives 1234.56).",
		"str", `num.strip_separators("1,234,567.89")`, 10, "strip remove thousands separators commas clean number normalize parse european",
		[]P{rs("value", "Formatted number."), os("decimal_point", "Decimal mark used by the input: \".\" or \",\".", ".", ".", ",")},
		func(a tools.Args) (any, error) {
			dp := a.Str("decimal_point")
			if dp != "." && dp != "," {
				return nil, errors.New("decimal_point must be \".\" or \",\"")
			}
			s := strings.TrimSpace(a.Str("value"))
			s = strings.NewReplacer(" ", "", "_", "", "'", "", "’", "", " ", "", " ", "").Replace(s)
			if dp == "." {
				s = strings.ReplaceAll(s, ",", "")
			} else {
				s = strings.ReplaceAll(strings.ReplaceAll(s, ".", ""), ",", ".")
			}
			if !ratRe.MatchString(s) {
				return nil, fmt.Errorf("%q is not a number", strings.TrimSpace(a.Str("value")))
			}
			return s, nil
		}, accepts(func(s string) bool { return strings.ContainsAny(s, ", _'") && len(s) < 100 }))

	reg("scientific", "Scientific notation", "Writes a number in scientific notation with `digits` digits after the point (123456789 with digits=3 gives 1.235e+8). Accepts decimals, fractions and exponents; works for huge integers.",
		"str", `num.scientific("123456789", 3)`, 11, "scientific notation exponent e standard form engineering mantissa",
		[]P{rs("value", "Number."), oi("digits", "Digits after the decimal point (0..200).", 3)},
		func(a tools.Args) (any, error) {
			d, err := intIn("digits", a.Int("digits"), 0, 200)
			if err != nil {
				return nil, err
			}
			r, err := parseRat(a.Str("value"))
			if err != nil {
				return nil, err
			}
			return sciFormat(r, int(d)), nil
		}, kw("sci", "exponent"), accepts(func(s string) bool { _, err := parseRat(s); return err == nil && len(s) < 200 }))

	reg("format_number", "Format number", "Formats a number with fixed decimals (rounded half away from zero), a thousands separator and a decimal point of your choice: 1234.5 becomes 1,234.50, or 1.234,50 with thousands=\".\" and decimal_point=\",\". Exact for any size.",
		"str", `num.format_number("1234567.891", 2)`, 12, "format number decimals thousands separator decimal point round locale currency",
		[]P{rs("value", "Number."), oi("decimals", "Digits after the decimal point (0..1000).", 2), os("thousands", "Thousands separator (may be empty).", ","), os("decimal_point", "Decimal mark.", ".")},
		func(a tools.Args) (any, error) {
			d, err := intIn("decimals", a.Int("decimals"), 0, 1000)
			if err != nil {
				return nil, err
			}
			r, err := parseRat(a.Str("value"))
			if err != nil {
				return nil, err
			}
			return formatNumber(r, int(d), a.Str("thousands"), a.Str("decimal_point")), nil
		}, accepts(func(s string) bool { _, err := parseRat(s); return err == nil && len(s) < 200 }))

	reg("parse_human_number", "Parse human number", "Parses shorthand like 1.5k, 2M, 3e6, 1,234.5, 2.5 billion or 50% into a plain decimal string (k=10^3, M=10^6, B=10^9, T=10^12, %=1/100). Currency symbols are ignored; unknown suffixes are errors.",
		"str", `num.parse_human_number("1.5k")`, 13, "parse human number shorthand k m b t million billion thousand abbreviation expand",
		[]P{rs("text", "Shorthand number.")},
		func(a tools.Args) (any, error) {
			r, err := parseHuman(a.Str("text"))
			if err != nil {
				return nil, err
			}
			return ratString(r, 12), nil
		}, kw("human", "expand"), accepts(func(s string) bool {
			if len(s) > 100 {
				return false
			}
			_, err := parseHuman(s)
			return err == nil && strings.ContainsAny(s, "kKmMbBtTqQ%eE")
		}))

	reg("humanize_bytes", "Humanize bytes", "Formats a byte count as a readable size: 1536 gives 1.54 KB (decimal, 1000-based) or 1.5 KiB with binary=true (1024-based). Trailing zeros are trimmed; negatives keep the sign.",
		"str", `num.humanize_bytes("1536", binary=True)`, 14, "bytes humanize human readable size kb mb gb kib mib gib file size format",
		[]P{rs("value", "Number of bytes."), ob("binary", "Use 1024-based KiB/MiB units instead of 1000-based KB/MB.", false), oi("decimals", "Digits after the decimal point (0..6).", 2)},
		func(a tools.Args) (any, error) {
			d, err := intIn("decimals", a.Int("decimals"), 0, 6)
			if err != nil {
				return nil, err
			}
			r, err := parseRat(a.Str("value"))
			if err != nil {
				return nil, err
			}
			return humanBytes(r, a.Bool("binary"), int(d)), nil
		}, kw("bytes", "size"), accepts(func(s string) bool { _, err := parseRat(s); return err == nil && len(s) < 40 }))

	reg("parse_bytes", "Parse byte size", "Parses a size like \"1.5 GiB\", \"200MB\" or \"4k\" into a byte count (int). KiB/MiB are 1024-based, KB/MB are 1000-based; bare K/M/G follow the binary flag. Results beyond int64 are an error.",
		"int", `num.parse_bytes("1.5 GiB")`, 15, "parse bytes size gib mib kb mb convert to bytes file size",
		[]P{rs("text", "Size text."), ob("binary", "Treat KB/MB/K/M as 1024-based.", false)},
		func(a tools.Args) (any, error) {
			n, err := parseBytes(a.Str("text"), a.Bool("binary"))
			if err != nil {
				return nil, err
			}
			if !n.IsInt64() {
				return nil, errors.New("size does not fit in an int; use bytes_table for huge values")
			}
			return n.Int64(), nil
		}, kw("bytes", "size"), accepts(func(s string) bool {
			if len(s) > 60 {
				return false
			}
			_, err := parseBytes(s, false)
			return err == nil && strings.ContainsAny(s, "bBkKmMgGtTpP")
		}))

	reg("bytes_table", "Bytes table", "Shows a byte count in every unit: bit, B, KB, MB, GB, TB, PB (1000-based) and KiB, MiB, GiB, TiB, PiB (1024-based), each rounded to 6 decimals. Accepts a number or a size such as \"1.5 GiB\".",
		"dict[str, str]", `num.bytes_table("1073741824")`, 16, "bytes table units kb mb gb tb kib mib gib conversion size chart",
		[]P{rs("value", "Byte count or a size like 1.5GiB.")},
		func(a tools.Args) (any, error) {
			var r *big.Rat
			if n, err := parseBytes(a.Str("value"), false); err == nil {
				r = new(big.Rat).SetInt(n)
			} else if r, err = parseRat(a.Str("value")); err != nil {
				return nil, err
			}
			m := map[string]string{"B": ratString(r, 6), "bit": ratString(new(big.Rat).Mul(r, big.NewRat(8, 1)), 6)}
			for i := 1; i <= 5; i++ {
				d := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(1000), big.NewInt(int64(i)), nil))
				m[decUnits[i]] = ratString(new(big.Rat).Quo(r, d), 6)
				b := new(big.Rat).SetInt(new(big.Int).Exp(big.NewInt(1024), big.NewInt(int64(i)), nil))
				m[binUnits[i]] = ratString(new(big.Rat).Quo(r, b), 6)
			}
			return m, nil
		}, kw("bytes", "size"), accepts(func(s string) bool { _, err := parseRat(s); return err == nil && len(s) < 40 }))

	reg("number_to_binary_bytes", "Number to bytes", "Converts an integer to a list of two-digit hex bytes, big-endian by default (258 gives [\"01\", \"02\"]) or little-endian. size=0 uses the minimal width (negatives as two's complement); a fixed size pads or errors if the value does not fit (max 8192 bytes).",
		"list[str]", `num.number_to_binary_bytes("258")`, 17, "bytes to bytes endian little big hex bytes array serialize integer",
		[]P{rs("value", "Integer."), oi("size", "Width in bytes (0 = minimal).", 0), os("endian", "big or little.", "big", "big", "little"), baseP("base")},
		func(a tools.Args) (any, error) {
			v, err := argBig(a, "value", int(a.Int("base")))
			if err != nil {
				return nil, err
			}
			size, err := intIn("size", a.Int("size"), 0, maxBits/8)
			if err != nil {
				return nil, err
			}
			e := strings.ToLower(a.Str("endian"))
			if e != "big" && e != "little" {
				return nil, errors.New("endian must be big or little")
			}
			if size == 0 {
				bits := v.BitLen()
				if v.Sign() < 0 {
					bits = new(big.Int).Not(v).BitLen() + 1
				}
				size = int64((bits + 7) / 8)
				if size == 0 {
					size = 1
				}
			}
			bits := int(size) * 8
			if v.Sign() < 0 && v.Cmp(new(big.Int).Neg(pow2(bits-1))) < 0 || v.Sign() >= 0 && v.BitLen() > bits {
				return nil, fmt.Errorf("value does not fit in %d bytes", size)
			}
			buf := new(big.Int).And(v, maskOf(bits)).FillBytes(make([]byte, size))
			out := make([]string, len(buf))
			for i := range buf {
				j := i
				if e == "little" {
					j = len(buf) - 1 - i
				}
				out[i] = fmt.Sprintf("%02x", buf[j])
			}
			return out, nil
		}, accepts(parsesAuto))

	reg("bytes_to_number", "Bytes to number", "Combines a list of hex bytes (\"01\", \"0xff\") into one integer, big-endian by default or little-endian, optionally as signed two's complement. Returns a decimal string.",
		"str", `num.bytes_to_number(["01", "02"])`, 18, "bytes to number endian little big hex bytes array integer deserialize",
		[]P{rl("bytes", "Hex byte strings."), os("endian", "big or little.", "big", "big", "little"), ob("signed", "Interpret as two's complement.", false)},
		func(a tools.Args) (any, error) {
			list := a.List("bytes")
			if len(list) == 0 || len(list) > maxBits/8 {
				return nil, errors.New("bytes must contain 1..8192 entries")
			}
			e := strings.ToLower(a.Str("endian"))
			if e != "big" && e != "little" {
				return nil, errors.New("endian must be big or little")
			}
			buf := make([]byte, len(list))
			for i, s := range list {
				n, err := parseBig(s, 16)
				if err != nil || n.Sign() < 0 || n.BitLen() > 8 {
					return nil, fmt.Errorf("bytes[%d] %q is not a byte", i, s)
				}
				j := i
				if e == "little" {
					j = len(list) - 1 - i
				}
				buf[j] = byte(n.Int64())
			}
			v := new(big.Int).SetBytes(buf)
			if a.Bool("signed") {
				v = signedWrap(v, len(buf)*8)
			}
			return v.String(), nil
		})

	reg("digits", "Digits", "Splits the absolute value of a decimal integer into its digits (1203 gives [\"1\", \"2\", \"0\", \"3\"]).",
		"list[str]", `num.digits("1203")`, 19, "digits split digit list decimal places",
		[]P{rs("value", "Integer, decimal string.")},
		func(a tools.Args) (any, error) {
			n, err := argBig(a, "value", 10)
			if err != nil {
				return nil, err
			}
			s := new(big.Int).Abs(n).String()
			out := make([]string, len(s))
			for i := range s {
				out[i] = s[i : i+1]
			}
			return out, nil
		}, accepts(parsesDec))

	digitSum := func(n *big.Int) int64 {
		var sum int64
		for _, c := range new(big.Int).Abs(n).String() {
			sum += int64(c - '0')
		}
		return sum
	}
	reg("digit_sum", "Digit sum", "Adds up the decimal digits of an integer's absolute value (9875 gives 29).",
		"int", `num.digit_sum("9875")`, 20, "digit sum add digits cross sum checksum",
		[]P{rs("value", "Integer, decimal string.")},
		func(a tools.Args) (any, error) {
			n, err := argBig(a, "value", 10)
			if err != nil {
				return nil, err
			}
			return digitSum(n), nil
		}, accepts(parsesDec))

	reg("digital_root", "Digital root", "Repeatedly sums the digits until one digit remains (9875 gives 2); the root of 0 is 0. Uses the absolute value.",
		"int", `num.digital_root("9875")`, 21, "digital root repeated digit sum casting out nines",
		[]P{rs("value", "Integer, decimal string.")},
		func(a tools.Args) (any, error) {
			n, err := argBig(a, "value", 10)
			if err != nil {
				return nil, err
			}
			if n.Sign() == 0 {
				return int64(0), nil
			}
			return 1 + new(big.Int).Mod(new(big.Int).Sub(new(big.Int).Abs(n), big.NewInt(1)), big.NewInt(9)).Int64(), nil
		}, accepts(parsesDec))

	reg("reverse_digits", "Reverse digits", "Reverses the decimal digits of an integer, keeping the sign (1230 gives 321, -45 gives -54).",
		"str", `num.reverse_digits("1230")`, 22, "reverse digits flip number backwards",
		[]P{rs("value", "Integer, decimal string.")},
		func(a tools.Args) (any, error) {
			n, err := argBig(a, "value", 10)
			if err != nil {
				return nil, err
			}
			b := []byte(new(big.Int).Abs(n).String())
			for i, j := 0, len(b)-1; i < j; i, j = i+1, j-1 {
				b[i], b[j] = b[j], b[i]
			}
			r, _ := new(big.Int).SetString(string(b), 10)
			if n.Sign() < 0 {
				r.Neg(r)
			}
			return r.String(), nil
		}, accepts(parsesDec))

	reg("is_palindrome_number", "Palindrome number", "Tells whether an integer reads the same backwards in the given base (default 10; 585 is a palindrome, and so is 9 in base 2 as 1001). Negative numbers are never palindromes.",
		"bool", `num.is_palindrome_number("12321")`, 23, "palindrome number symmetric reads same backwards",
		[]P{rs("value", "Integer."), oi("base", "Base 2..36 used for the digits.", 10)},
		func(a tools.Args) (any, error) {
			n, err := argBig(a, "value", 10)
			if err != nil {
				return nil, err
			}
			b, err := intIn("base", a.Int("base"), 2, 36)
			if err != nil {
				return nil, err
			}
			if n.Sign() < 0 {
				return false, nil
			}
			s := n.Text(int(b))
			for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
				if s[i] != s[j] {
					return false, nil
				}
			}
			return true, nil
		}, accepts(parsesDec))
}

package lines

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

// maxOut is the largest string any function in this package will produce.
const maxOut = 4 << 20

type P = tools.Param

func ps(n, d string) P  { return P{Name: n, Type: "str", Doc: d, Required: true} }
func pir(n, d string) P { return P{Name: n, Type: "int", Doc: d, Required: true} }
func pso(n, d, def string, choices ...string) P {
	return P{Name: n, Type: "str", Doc: d, Default: def, Choices: choices}
}
func pi(n, d string, def int64) P { return P{Name: n, Type: "int", Doc: d, Default: def} }
func pb(n, d string, def bool) P  { return P{Name: n, Type: "bool", Doc: d, Default: def} }
func pls(n, d string, def []string) P {
	return P{Name: n, Type: "list[str]", Doc: d, Default: def}
}

// sIn is the standard input parameter of every lines function.
var sIn = ps("s", "Input text, one item per line.")

// Line model, shared by the functions that return text lines: the input is split
// on LF, CRLF or lone CR; the dominant ending (CRLF if present, else CR if
// present, else LF) and one trailing line break are remembered and restored
// in the output. Empty input has zero lines.
type text struct {
	lines []string
	eol   string
	trail bool
}

func split(s string) text {
	t := text{eol: "\n"}
	switch {
	case strings.Contains(s, "\r\n"):
		t.eol = "\r\n"
	case strings.Contains(s, "\r") && !strings.Contains(s, "\n"):
		t.eol = "\r"
	}
	if s == "" {
		return t
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\r", "\n")
	if strings.HasSuffix(s, "\n") {
		t.trail = true
		s = s[:len(s)-1]
	}
	t.lines = strings.Split(s, "\n")
	return t
}

func (t text) join(ls []string) (string, error) {
	n := 0
	for _, l := range ls {
		n += len(l) + len(t.eol)
	}
	if n > maxOut {
		return "", fmt.Errorf("result would exceed %d MiB", maxOut>>20)
	}
	out := strings.Join(ls, t.eol)
	if t.trail && len(ls) > 0 {
		out += t.eol
	}
	return out, nil
}

type lineFn = func(ls []string, a tools.Args) ([]string, error)

func family(params []P) string {
	for _, p := range params[1:] {
		if p.Required {
			return ""
		}
	}
	return "lines"
}

func regRaw(name, title, doc, ret, example string, order int, tags string, params []P, f func(a tools.Args) (any, error)) {
	tools.Register(tools.Func{
		Group: "lines", Name: name, Title: title, Doc: doc, Params: params, Returns: ret, Fn: f,
		Example: example, Tags: strings.Fields(tags), Order: order, Family: family(params),
	})
}

// reg registers a lines-in, lines-out function (the common case); params must
// start with the text parameter "s".
func reg(name, title, doc, example string, order int, tags string, params []P, f lineFn) {
	regRaw(name, title, doc+" Preserves the input's line ending style and one trailing newline.", "str", example, order, tags, params,
		func(a tools.Args) (any, error) {
			t := split(a.Str("s"))
			out, err := f(t.lines, a)
			if err != nil {
				return nil, err
			}
			return t.join(out)
		})
}

// regList registers a function that returns the lines as a list[str].
func regStr(name, title, doc, ret, example string, order int, tags string, params []P, f func(t text, a tools.Args) (any, error)) {
	regRaw(name, title, doc, ret, example, order, tags, params,
		func(a tools.Args) (any, error) { return f(split(a.Str("s")), a) })
}

func itoa(n int) string { return strconv.Itoa(n) }

func fold(s string, ci bool) string {
	if ci {
		return strings.ToLower(s)
	}
	return s
}

var leadingNum = regexp.MustCompile(`^\s*[-+]?(?:\d+\.?\d*|\.\d+)(?:[eE][-+]?\d+)?`)
var thousands = regexp.MustCompile(`^[-+]?\d{1,3}(?:,\d{3})+(?:\.\d+)?$`)

// parseNumber parses a whole line as a number; thousands separators like
// 1,234.5 are accepted.
func parseNumber(s string) (float64, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, false
	}
	if thousands.MatchString(s) {
		s = strings.ReplaceAll(s, ",", "")
	}
	f, err := strconv.ParseFloat(s, 64)
	return f, err == nil
}

func leadingNumber(s string) (float64, bool) {
	m := leadingNum.FindString(s)
	if m == "" {
		return 0, false
	}
	f, err := strconv.ParseFloat(strings.TrimSpace(m), 64)
	return f, err == nil
}

func isDigit(b byte) bool { return b >= '0' && b <= '9' }

// naturalCmp orders "file2" before "file10" and ignores case in text parts.
func naturalCmp(a, b string) int {
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		if isDigit(a[i]) && isDigit(b[j]) {
			si := i
			for i < len(a) && isDigit(a[i]) {
				i++
			}
			sj := j
			for j < len(b) && isDigit(b[j]) {
				j++
			}
			da, db := strings.TrimLeft(a[si:i], "0"), strings.TrimLeft(b[sj:j], "0")
			if len(da) != len(db) {
				if len(da) < len(db) {
					return -1
				}
				return 1
			}
			if c := strings.Compare(da, db); c != 0 {
				return c
			}
			continue
		}
		ra, na := utf8.DecodeRuneInString(a[i:])
		rb, nb := utf8.DecodeRuneInString(b[j:])
		ra, rb = unicode.ToLower(ra), unicode.ToLower(rb)
		if ra != rb {
			if ra < rb {
				return -1
			}
			return 1
		}
		i += na
		j += nb
	}
	switch {
	case len(a)-i < len(b)-j:
		return -1
	case len(a)-i > len(b)-j:
		return 1
	}
	return 0
}

// compareFor returns a comparison function for a sort mode.
func compareFor(mode string, desc bool) (func(a, b string) int, error) {
	sign := 1
	if desc {
		sign = -1
	}
	switch strings.ToLower(mode) {
	case "text", "":
		return func(a, b string) int { return sign * strings.Compare(a, b) }, nil
	case "case_insensitive", "ci":
		return func(a, b string) int { return sign * strings.Compare(strings.ToLower(a), strings.ToLower(b)) }, nil
	case "natural":
		return func(a, b string) int { return sign * naturalCmp(a, b) }, nil
	case "length":
		return func(a, b string) int {
			return sign * (utf8.RuneCountInString(a) - utf8.RuneCountInString(b))
		}, nil
	case "numeric":
		return func(a, b string) int {
			fa, oka := leadingNumber(a)
			fb, okb := leadingNumber(b)
			switch {
			case !oka && !okb:
				return 0
			case !oka:
				return 1
			case !okb:
				return -1
			case fa < fb:
				return -sign
			case fa > fb:
				return sign
			}
			return 0
		}, nil
	}
	return nil, fmt.Errorf("unknown mode %q (use text, natural, numeric, length or case_insensitive)", mode)
}

var sortModes = []string{"text", "natural", "numeric", "length", "case_insensitive"}

func orderIsDesc(order string) (bool, error) {
	switch strings.ToLower(order) {
	case "asc", "":
		return false, nil
	case "desc":
		return true, nil
	}
	return false, fmt.Errorf("unknown order %q (use asc or desc)", order)
}

// fields splits a line by delimiter; an empty delimiter means runs of whitespace.
func fields(line, delim string) []string {
	if delim == "" {
		return strings.Fields(line)
	}
	return strings.Split(line, delim)
}

func joinDelim(f []string, delim string) string {
	if delim == "" {
		delim = " "
	}
	return strings.Join(f, delim)
}

func runeLen(s string) int { return utf8.RuneCountInString(s) }

package num

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

// Limits that keep every call cheap.
const (
	maxInputLen = 20000   // characters of any numeric input
	maxBits     = 1 << 16 // widest bit field a function will build
	maxShift    = 1 << 20 // largest shift or bit index
)

type P = tools.Param
type fn = func(a tools.Args) (any, error)

func rs(n, d string) P { return P{Name: n, Type: "str", Doc: d, Required: true} }
func os(n, d, def string, choices ...string) P {
	return P{Name: n, Type: "str", Doc: d, Default: def, Choices: choices}
}
func ri(n, d string) P            { return P{Name: n, Type: "int", Doc: d, Required: true} }
func oi(n, d string, def int64) P { return P{Name: n, Type: "int", Doc: d, Default: def} }
func ob(n, d string, def bool) P  { return P{Name: n, Type: "bool", Doc: d, Default: def} }
func rl(n, d string) P            { return P{Name: n, Type: "list[str]", Doc: d, Required: true} }

var outChoices = []string{"dec", "hex", "bin", "oct"}

func outP() P {
	return os("output", "Output format: dec (default), hex (0x prefix), bin (0b prefix) or oct (0o prefix).", "dec", outChoices...)
}

func baseP(name string) P {
	return oi(name, "Input base 2..62. 0 (default) auto-detects: a 0x/0b/0o prefix wins, hex letters mean base 16, otherwise decimal.", 0)
}

// reg registers a num function. A function whose extra parameters are all
// optional also becomes a launcher variant of the "num" family.
func reg(name, title, doc, ret, example string, order int, tags string, params []P, f fn, opts ...func(*tools.Func)) {
	t := tools.Func{
		Group: "num", Name: name, Title: title, Doc: doc, Params: params,
		Returns: ret, Fn: f, Example: example, Tags: strings.Fields(tags), Order: order,
	}
	if launcher(params) {
		t.Family = "num"
	}
	for _, o := range opts {
		o(&t)
	}
	tools.Register(t)
}

func launcher(params []P) bool {
	hasStr := false
	for _, p := range params {
		if p.Type == "str" && !hasStr {
			hasStr = true
			continue
		}
		if p.Required {
			return false
		}
	}
	return hasStr
}

func kw(k ...string) func(*tools.Func) { return func(f *tools.Func) { f.Keywords = k } }
func accepts(fn func(string) bool) func(*tools.Func) {
	return func(f *tools.Func) { f.Accepts = fn }
}

// ---- number parsing ----

func prefixApplies(letter byte, base int) bool {
	switch letter {
	case 'b', 'B':
		return base <= 11
	case 'o', 'O':
		return base <= 24
	case 'x', 'X':
		return base <= 33
	}
	return false
}

// splitPrefix removes the sign and any 0x/0b/0o prefix. pb is the prefix base
// or 0.
func splitPrefix(s string, base int) (neg bool, body string, pb int) {
	t := strings.TrimSpace(s)
	if t != "" && (t[0] == '-' || t[0] == '+') {
		neg = t[0] == '-'
		t = strings.TrimSpace(t[1:])
	}
	if len(t) > 2 && t[0] == '0' {
		switch t[1] {
		case 'x', 'X':
			pb = 16
		case 'b', 'B':
			pb = 2
		case 'o', 'O':
			pb = 8
		}
		if pb != 0 && base != 0 && !prefixApplies(t[1], base) {
			pb = 0
		}
		if pb != 0 {
			t = t[2:]
		}
	}
	return neg, t, pb
}

func isHexDigits(t string) bool {
	if t == "" {
		return false
	}
	for _, r := range t {
		if !(r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F' || r == '_' || r == ' ') {
			return false
		}
	}
	return true
}

func hasHexLetter(t string) bool { return strings.ContainsAny(t, "abcdefABCDEF") }

// parseBig parses s in base (0 = auto-detect). Prefixes, underscores, spaces
// and a leading sign are accepted.
func parseBig(s string, base int) (*big.Int, error) {
	if len(s) > maxInputLen {
		return nil, errors.New("number is too long")
	}
	if strings.TrimSpace(s) == "" {
		return nil, errors.New("empty number")
	}
	if base < 0 || base == 1 || base > 62 {
		return nil, fmt.Errorf("base must be 2..62 (or 0 for auto), got %d", base)
	}
	neg, body, pb := splitPrefix(s, base)
	b := base
	switch {
	case pb != 0:
		b = pb
	case base == 0:
		if isHexDigits(body) && hasHexLetter(body) {
			b = 16
		} else {
			b = 10
		}
	}
	body = strings.Map(func(r rune) rune {
		if r == '_' || r == ' ' || r == ' ' || r == ' ' {
			return -1
		}
		return r
	}, body)
	if body == "" || body[0] == '+' || body[0] == '-' {
		return nil, fmt.Errorf("invalid number %q", s)
	}
	n, ok := new(big.Int).SetString(body, b)
	if !ok {
		return nil, fmt.Errorf("%q is not a valid base-%d number", strings.TrimSpace(s), b)
	}
	if neg {
		n.Neg(n)
	}
	return n, nil
}

func dec(s string) (*big.Int, error) { return parseBig(s, 10) }

// argBig parses the string argument name (prefixes allowed).
func argBig(a tools.Args, name string, base int) (*big.Int, error) {
	n, err := parseBig(a.Str(name), base)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	return n, nil
}

// argSmall parses a decimal string into an int64 in [lo, hi].
func argSmall(a tools.Args, name string, lo, hi int64) (int64, error) {
	n, err := argBig(a, name, 10)
	if err != nil {
		return 0, err
	}
	return smallRange(name, n, lo, hi)
}

func smallRange(name string, n *big.Int, lo, hi int64) (int64, error) {
	if !n.IsInt64() || n.Int64() < lo || n.Int64() > hi {
		return 0, fmt.Errorf("%s must be between %d and %d", name, lo, hi)
	}
	return n.Int64(), nil
}

func intIn(name string, v, lo, hi int64) (int64, error) {
	if v < lo || v > hi {
		return 0, fmt.Errorf("%s must be between %d and %d", name, lo, hi)
	}
	return v, nil
}

// ---- number formatting ----

func groupDigits(s string, size int, sep string) string {
	if size <= 0 || len(s) <= size {
		return s
	}
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	var parts []string
	for len(s) > size {
		parts = append([]string{s[len(s)-size:]}, parts...)
		s = s[:len(s)-size]
	}
	parts = append([]string{s}, parts...)
	out := strings.Join(parts, sep)
	if neg {
		out = "-" + out
	}
	return out
}

func prefixFor(base int) string {
	switch base {
	case 2:
		return "0b"
	case 8:
		return "0o"
	case 16:
		return "0x"
	}
	return ""
}

// fmtBase renders n in base 2..62.
func fmtBase(n *big.Int, base int, prefix, upper bool, group int) string {
	s := n.Text(base)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	if upper && base <= 36 {
		s = strings.ToUpper(s)
	}
	s = groupDigits(s, group, " ")
	if prefix {
		s = prefixFor(base) + s
	}
	if neg {
		s = "-" + s
	}
	return s
}

// outStr renders n for an output parameter: dec, hex, bin or oct.
func outStr(n *big.Int, output string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(output)) {
	case "", "dec", "decimal", "10":
		return n.String(), nil
	case "hex", "16":
		return fmtBase(n, 16, true, false, 0), nil
	case "bin", "2":
		return fmtBase(n, 2, true, false, 0), nil
	case "oct", "8":
		return fmtBase(n, 8, true, false, 0), nil
	}
	return "", fmt.Errorf("output must be dec, hex, bin or oct, got %q", output)
}

func pow2(bits int) *big.Int { return new(big.Int).Lsh(big.NewInt(1), uint(bits)) }

func maskOf(bits int) *big.Int { return new(big.Int).Sub(pow2(bits), big.NewInt(1)) }

func parsesAuto(s string) bool {
	if len(s) > 2000 {
		return false
	}
	_, err := parseBig(s, 0)
	return err == nil
}

func parsesDec(s string) bool {
	if len(s) > 2000 {
		return false
	}
	_, err := parseBig(s, 10)
	return err == nil
}

// autoBase guesses the base of s: prefix, then hex letters, else decimal.
func autoBase(s string) int {
	_, body, pb := splitPrefix(s, 0)
	if pb != 0 {
		return pb
	}
	if isHexDigits(body) && hasHexLetter(body) {
		return 16
	}
	return 10
}

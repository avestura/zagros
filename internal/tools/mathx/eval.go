package mathx

import (
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"

	"github.com/avestura/zagros/internal/tools"
)

const (
	maxExprLen   = 5000
	maxExprDepth = 150
)

type tkind byte

const (
	tNum tkind = iota
	tID
	tFunc
	tOp
	tEnd
)

type token struct {
	kind tkind
	s    string
	v    float64
	pos  int // 1-based character position
}

func isLetter(r rune) bool { return unicode.IsLetter(r) }
func isDigit(r rune) bool  { return r >= '0' && r <= '9' }

func lexExpr(src string) ([]token, error) {
	rs := []rune(src)
	var out []token
	var parens []bool // true for function-call parentheses
	for i := 0; i < len(rs); {
		r := rs[i]
		switch {
		case unicode.IsSpace(r):
			i++
		case isDigit(r) || r == '.' && i+1 < len(rs) && isDigit(rs[i+1]):
			start := i
			funcCtx := len(parens) > 0 && parens[len(parens)-1]
			var b strings.Builder
			grp, hadComma := 0, false
			for i < len(rs) {
				c := rs[i]
				if isDigit(c) {
					b.WriteRune(c)
					grp++
					i++
				} else if c == '_' && i+1 < len(rs) && isDigit(rs[i+1]) && grp > 0 {
					i++
				} else if c == ',' && !funcCtx && i+4 < len(rs)+0 && isDigit(rs[i+1]) && isDigit(rs[i+2]) && isDigit(rs[i+3]) &&
					(i+4 >= len(rs) || !isDigit(rs[i+4])) && ((!hadComma && grp >= 1 && grp <= 3) || (hadComma && grp == 3)) {
					hadComma = true
					grp = 0
					i++
				} else if c == ',' && !funcCtx && i+4 == len(rs) && isDigit(rs[i+1]) && isDigit(rs[i+2]) && isDigit(rs[i+3]) &&
					((!hadComma && grp >= 1 && grp <= 3) || (hadComma && grp == 3)) {
					hadComma = true
					grp = 0
					i++
				} else {
					break
				}
			}
			if i < len(rs) && rs[i] == '.' {
				b.WriteRune('.')
				i++
				for i < len(rs) && (isDigit(rs[i]) || rs[i] == '_' && i+1 < len(rs) && isDigit(rs[i+1])) {
					if rs[i] != '_' {
						b.WriteRune(rs[i])
					}
					i++
				}
			}
			if i < len(rs) && (rs[i] == 'e' || rs[i] == 'E') {
				j := i + 1
				if j < len(rs) && (rs[j] == '+' || rs[j] == '-') {
					j++
				}
				if j < len(rs) && isDigit(rs[j]) {
					for j < len(rs) && isDigit(rs[j]) {
						j++
					}
					b.WriteString(string(rs[i:j]))
					i = j
				}
			}
			v, err := strconv.ParseFloat(b.String(), 64)
			if err != nil && !errors.Is(err, strconv.ErrRange) {
				return nil, fmt.Errorf("invalid number %q at position %d", b.String(), start+1)
			}
			if math.IsInf(v, 0) {
				return nil, fmt.Errorf("number too large at position %d", start+1)
			}
			out = append(out, token{kind: tNum, v: v, s: b.String(), pos: start + 1})
		case isLetter(r) || r == 'π':
			start := i
			for i < len(rs) && (isLetter(rs[i]) || isDigit(rs[i]) || rs[i] == '_') {
				i++
			}
			w := strings.ToLower(string(rs[start:i]))
			j := i
			for j < len(rs) && rs[j] == ' ' {
				j++
			}
			switch {
			case w == "of":
				out = append(out, token{kind: tOp, s: "*", pos: start + 1})
			case j < len(rs) && rs[j] == '(' && w != "pi" && w != "e" && w != "tau" && w != "phi":
				out = append(out, token{kind: tFunc, s: w, pos: start + 1})
			default:
				out = append(out, token{kind: tID, s: w, pos: start + 1})
			}
		default:
			pos := i + 1
			op := ""
			switch r {
			case '+', '-', '/', '%', '^', ',', '!':
				op = string(r)
			case '*':
				op = "*"
				if i+1 < len(rs) && rs[i+1] == '*' {
					op = "^"
					i++
				}
			case '(':
				isFn := len(out) > 0 && out[len(out)-1].kind == tFunc
				parens = append(parens, isFn)
				op = "("
			case ')':
				if len(parens) > 0 {
					parens = parens[:len(parens)-1]
				}
				op = ")"
			case '×', '·', '⋅':
				op = "*"
			case '÷', ':':
				op = "/"
			case '−', '–':
				op = "-"
			case '²':
				out = append(out, token{kind: tOp, s: "^", pos: pos}, token{kind: tNum, v: 2, s: "2", pos: pos})
				i++
				continue
			case '³':
				out = append(out, token{kind: tOp, s: "^", pos: pos}, token{kind: tNum, v: 3, s: "3", pos: pos})
				i++
				continue
			case '√':
				out = append(out, token{kind: tFunc, s: "sqrt", pos: pos})
				i++
				continue
			default:
				return nil, fmt.Errorf("unexpected character %q at position %d", string(r), pos)
			}
			out = append(out, token{kind: tOp, s: op, pos: pos})
			i++
		}
	}
	out = append(out, token{kind: tEnd, pos: len(rs) + 1})
	return out, nil
}

type parser struct {
	toks    []token
	i       int
	depth   int
	degrees bool
}

type val struct {
	v   float64
	pct bool // the term was a bare percent literal like 10%
}

func (p *parser) peek() token { return p.toks[p.i] }
func (p *parser) next() token { t := p.toks[p.i]; p.i++; return t }
func (p *parser) isOp(s string) bool {
	t := p.peek()
	return t.kind == tOp && t.s == s
}

func evalExpr(expr string, degrees bool) (float64, error) {
	if len([]rune(expr)) > maxExprLen {
		return 0, errors.New("expression is too long")
	}
	if strings.TrimSpace(expr) == "" {
		return 0, errors.New("empty expression")
	}
	toks, err := lexExpr(expr)
	if err != nil {
		return 0, err
	}
	p := &parser{toks: toks, degrees: degrees}
	v, err := p.parseSum()
	if err != nil {
		return 0, err
	}
	if t := p.peek(); t.kind != tEnd {
		return 0, fmt.Errorf("unexpected %s at position %d", describe(t), t.pos)
	}
	return v.v, nil
}

func describe(t token) string {
	switch t.kind {
	case tEnd:
		return "end of expression"
	case tNum:
		return "number " + t.s
	}
	return strconv.Quote(t.s)
}

func (p *parser) checkFinite(v float64, pos int, what string) error {
	if math.IsNaN(v) {
		return fmt.Errorf("%s is undefined at position %d", what, pos)
	}
	if math.IsInf(v, 0) {
		return fmt.Errorf("%s overflows at position %d", what, pos)
	}
	return nil
}

func (p *parser) parseSum() (val, error) {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > maxExprDepth {
		return val{}, errors.New("expression is nested too deeply")
	}
	l, err := p.parseTerm()
	if err != nil {
		return val{}, err
	}
	for p.isOp("+") || p.isOp("-") {
		op := p.next()
		r, err := p.parseTerm()
		if err != nil {
			return val{}, err
		}
		rv := r.v
		if r.pct {
			rv = l.v * r.v // "200 + 10%" means 200 + 10% of 200
		}
		if op.s == "+" {
			l = val{v: l.v + rv}
		} else {
			l = val{v: l.v - rv}
		}
		if err := p.checkFinite(l.v, op.pos, "result"); err != nil {
			return val{}, err
		}
	}
	return l, nil
}

func (p *parser) startsOperand() bool {
	t := p.peek()
	return t.kind == tNum || t.kind == tID || t.kind == tFunc || t.kind == tOp && t.s == "("
}

func (p *parser) parseTerm() (val, error) {
	l, err := p.parseUnary()
	if err != nil {
		return val{}, err
	}
	for {
		t := p.peek()
		switch {
		case t.kind == tOp && (t.s == "*" || t.s == "/" || t.s == "%"):
			p.next()
			r, err := p.parseUnary()
			if err != nil {
				return val{}, err
			}
			switch t.s {
			case "*":
				l = val{v: l.v * r.v}
			case "/":
				if r.v == 0 {
					return val{}, fmt.Errorf("division by zero at position %d", t.pos)
				}
				l = val{v: l.v / r.v}
			default:
				if r.v == 0 {
					return val{}, fmt.Errorf("modulo by zero at position %d", t.pos)
				}
				l = val{v: math.Mod(l.v, r.v)}
			}
			if err := p.checkFinite(l.v, t.pos, "result"); err != nil {
				return val{}, err
			}
		case p.startsOperand():
			// implicit multiplication: 2(3+4), 3pi, (1+2)(3+4), 2sin(1)
			if prev := p.toks[p.i-1]; prev.kind == tNum && t.kind == tNum {
				return val{}, fmt.Errorf("unexpected number %s at position %d (missing operator?)", t.s, t.pos)
			}
			r, err := p.parseUnary()
			if err != nil {
				return val{}, err
			}
			l = val{v: l.v * r.v}
			if err := p.checkFinite(l.v, t.pos, "result"); err != nil {
				return val{}, err
			}
		default:
			return l, nil
		}
	}
}

func (p *parser) parseUnary() (val, error) {
	t := p.peek()
	if t.kind == tOp && (t.s == "-" || t.s == "+") {
		p.next()
		p.depth++
		defer func() { p.depth-- }()
		if p.depth > maxExprDepth {
			return val{}, errors.New("expression is nested too deeply")
		}
		v, err := p.parseUnary()
		if err != nil {
			return val{}, err
		}
		if t.s == "-" {
			v.v = -v.v
		}
		return v, nil
	}
	return p.parsePower()
}

func (p *parser) parsePower() (val, error) {
	b, err := p.parsePostfix()
	if err != nil {
		return val{}, err
	}
	if p.isOp("^") {
		t := p.next()
		e, err := p.parseUnary()
		if err != nil {
			return val{}, err
		}
		r := math.Pow(b.v, e.v)
		if err := p.checkFinite(r, t.pos, "power"); err != nil {
			return val{}, err
		}
		return val{v: r}, nil
	}
	return b, nil
}

func (p *parser) parsePostfix() (val, error) {
	v, err := p.parsePrimary()
	if err != nil {
		return val{}, err
	}
	for {
		t := p.peek()
		switch {
		case t.kind == tOp && t.s == "!":
			p.next()
			f, err := factorialF(v.v)
			if err != nil {
				return val{}, fmt.Errorf("%v at position %d", err, t.pos)
			}
			v = val{v: f}
		case t.kind == tOp && t.s == "%":
			// postfix percent unless a number/operand follows (then it is modulo)
			nxt := p.toks[p.i+1]
			if nxt.kind == tNum || nxt.kind == tID || nxt.kind == tFunc || nxt.kind == tOp && nxt.s == "(" {
				return v, nil
			}
			p.next()
			v = val{v: v.v / 100, pct: true}
		default:
			return v, nil
		}
	}
}

var constants = map[string]float64{
	"pi": math.Pi, "π": math.Pi, "e": math.E, "tau": 2 * math.Pi, "phi": (1 + math.Sqrt(5)) / 2,
}

func (p *parser) parsePrimary() (val, error) {
	t := p.next()
	switch t.kind {
	case tNum:
		return val{v: t.v}, nil
	case tID:
		c, ok := constants[t.s]
		if !ok {
			return val{}, fmt.Errorf("unknown name %q at position %d", t.s, t.pos)
		}
		return val{v: c}, nil
	case tFunc:
		p.next() // '('
		var args []float64
		if p.isOp(")") {
			p.next()
		} else {
			for {
				a, err := p.parseSum()
				if err != nil {
					return val{}, err
				}
				args = append(args, a.v)
				if p.isOp(",") {
					p.next()
					continue
				}
				if p.isOp(")") {
					p.next()
					break
				}
				return val{}, fmt.Errorf("expected ',' or ')' at position %d", p.peek().pos)
			}
		}
		r, err := callFunc(t.s, args, p.degrees)
		if err != nil {
			return val{}, fmt.Errorf("%v at position %d", err, t.pos)
		}
		if err := p.checkFinite(r, t.pos, t.s+"()"); err != nil {
			return val{}, err
		}
		return val{v: r}, nil
	case tOp:
		if t.s == "(" {
			v, err := p.parseSum()
			if err != nil {
				return val{}, err
			}
			if !p.isOp(")") {
				return val{}, fmt.Errorf("missing ')' at position %d", p.peek().pos)
			}
			p.next()
			return val{v: v.v}, nil
		}
	case tEnd:
		return val{}, fmt.Errorf("unexpected end of expression at position %d", t.pos)
	}
	return val{}, fmt.Errorf("unexpected %s at position %d", describe(t), t.pos)
}

func factorialF(x float64) (float64, error) {
	if x < 0 || x != math.Floor(x) {
		return 0, errors.New("factorial needs a non-negative integer")
	}
	if x > 170 {
		return 0, errors.New("factorial overflows (n must be <= 170)")
	}
	r := 1.0
	for i := 2.0; i <= x; i++ {
		r *= i
	}
	return r, nil
}

type fnSpec struct {
	min, max int // max -1 = variadic
	f        func(a []float64, deg bool) (float64, error)
}

func f1(f func(float64) float64) fnSpec {
	return fnSpec{1, 1, func(a []float64, _ bool) (float64, error) { return f(a[0]), nil }}
}

func trig(f func(float64) float64) fnSpec {
	return fnSpec{1, 1, func(a []float64, deg bool) (float64, error) {
		if deg {
			return f(a[0] * math.Pi / 180), nil
		}
		return f(a[0]), nil
	}}
}

func itrig(f func(float64) float64) fnSpec {
	return fnSpec{1, 1, func(a []float64, deg bool) (float64, error) {
		if deg {
			return f(a[0]) * 180 / math.Pi, nil
		}
		return f(a[0]), nil
	}}
}

func variadic(f func([]float64) float64) fnSpec {
	return fnSpec{1, -1, func(a []float64, _ bool) (float64, error) { return f(a), nil }}
}

var funcTable map[string]fnSpec

func init() {
	funcTable = map[string]fnSpec{
		"sin": trig(math.Sin), "cos": trig(math.Cos), "tan": trig(math.Tan),
		"asin": itrig(math.Asin), "acos": itrig(math.Acos), "atan": itrig(math.Atan),
		"sinh": f1(math.Sinh), "cosh": f1(math.Cosh), "tanh": f1(math.Tanh),
		"asinh": f1(math.Asinh), "acosh": f1(math.Acosh), "atanh": f1(math.Atanh),
		"sqrt": f1(math.Sqrt), "cbrt": f1(math.Cbrt), "exp": f1(math.Exp), "ln": f1(math.Log),
		"log2": f1(math.Log2), "log10": f1(math.Log10), "abs": f1(math.Abs),
		"floor": f1(math.Floor), "ceil": f1(math.Ceil), "trunc": f1(math.Trunc), "gamma": f1(math.Gamma),
		"sign": f1(func(x float64) float64 {
			switch {
			case x > 0:
				return 1
			case x < 0:
				return -1
			}
			return 0
		}),
		"deg": f1(func(x float64) float64 { return x * 180 / math.Pi }),
		"rad": f1(func(x float64) float64 { return x * math.Pi / 180 }),
		"atan2": {2, 2, func(a []float64, deg bool) (float64, error) {
			r := math.Atan2(a[0], a[1])
			if deg {
				r = r * 180 / math.Pi
			}
			return r, nil
		}},
		"log": {1, 2, func(a []float64, _ bool) (float64, error) {
			if len(a) == 1 {
				return math.Log10(a[0]), nil
			}
			return math.Log(a[0]) / math.Log(a[1]), nil
		}},
		"round": {1, 2, func(a []float64, _ bool) (float64, error) {
			d := 0
			if len(a) == 2 {
				d = int(a[1])
			}
			return roundTo(a[0], d, "half_up")
		}},
		"min": variadic(func(a []float64) float64 {
			m := a[0]
			for _, x := range a[1:] {
				m = math.Min(m, x)
			}
			return m
		}),
		"max": variadic(func(a []float64) float64 {
			m := a[0]
			for _, x := range a[1:] {
				m = math.Max(m, x)
			}
			return m
		}),
		"sum": variadic(func(a []float64) float64 { return sumOf(a) }),
		"avg": variadic(func(a []float64) float64 { return sumOf(a) / float64(len(a)) }),
		"pow": {2, 2, func(a []float64, _ bool) (float64, error) { return math.Pow(a[0], a[1]), nil }},
		"mod": {2, 2, func(a []float64, _ bool) (float64, error) {
			if a[1] == 0 {
				return 0, errors.New("modulo by zero")
			}
			return math.Mod(a[0], a[1]), nil
		}},
		"hypot": variadic(func(a []float64) float64 {
			s := 0.0
			for _, x := range a {
				s += x * x
			}
			return math.Sqrt(s)
		}),
		"factorial": {1, 1, func(a []float64, _ bool) (float64, error) { return factorialF(a[0]) }},
	}
	funcTable["mean"] = funcTable["avg"]
	funcTable["fact"] = funcTable["factorial"]
	funcTable["arcsin"], funcTable["arccos"], funcTable["arctan"] = funcTable["asin"], funcTable["acos"], funcTable["atan"]
	funcTable["lg"] = funcTable["log10"]
}

func sumOf(a []float64) float64 {
	// Kahan summation keeps 0.1+0.2+... tidy.
	var s, c float64
	for _, x := range a {
		y := x - c
		t := s + y
		c = (t - s) - y
		s = t
	}
	return s
}

func callFunc(name string, args []float64, deg bool) (float64, error) {
	spec, ok := funcTable[name]
	if !ok {
		return 0, fmt.Errorf("unknown function %q", name)
	}
	if len(args) < spec.min || spec.max >= 0 && len(args) > spec.max {
		if spec.max < 0 {
			return 0, fmt.Errorf("%s() needs at least %d argument(s)", name, spec.min)
		}
		if spec.min == spec.max {
			return 0, fmt.Errorf("%s() takes %d argument(s), got %d", name, spec.min, len(args))
		}
		return 0, fmt.Errorf("%s() takes %d to %d arguments, got %d", name, spec.min, spec.max, len(args))
	}
	return spec.f(args, deg)
}

func looksLikeMath(s string) bool {
	if len(s) > 300 || strings.TrimSpace(s) == "" {
		return false
	}
	t := strings.TrimSpace(s)
	if _, err := strconv.ParseFloat(t, 64); err == nil {
		return false
	}
	if !strings.ContainsAny(t, "+-*/^%()!×÷√²³") && !strings.ContainsAny(strings.ToLower(t), "abcdefghijklmnopqrstuvwxyzπ") {
		return false
	}
	_, err := evalExpr(t, false)
	return err == nil
}

func init() {
	reg("eval", "Calculate", "Evaluates a math expression with floats and returns the number. Supports + - * / % (remainder) ^ or ** (power), parentheses, implicit multiplication (2(3+4), 3pi), constants pi e tau phi, factorial 5!, thousands separators (1,000), scientific notation (1e3), functions sin cos tan asin acos atan atan2 sinh cosh tanh sqrt cbrt exp ln log(x) [base 10] log(x,base) log2 log10 abs floor ceil round(x[,d]) trunc sign min max sum avg pow hypot gamma factorial deg rad, and the word 'of' as multiplication. Percent: 50% is 0.5, and in a sum 'a + b%' means a plus b percent of a (200 + 10% = 220); a comma is never a decimal point. Errors name the character position.",
		"float", `math.eval("2(3+4) + 10%")`, 1, "calculator calc evaluate expression arithmetic formula compute number sqrt sin cos log percent",
		[]P{rs("expr", "The expression, e.g. 2*(3+4)^2, sqrt(16), 200 + 10%, 15% of 80."), ob("degrees", "Trig functions take and return degrees instead of radians.", false)},
		func(a tools.Args) (any, error) {
			v, err := evalExpr(a.Str("expr"), a.Bool("degrees"))
			if err != nil {
				return nil, err
			}
			return v, nil
		}, accepts(looksLikeMath), kw("calc", "calculate", "="))

	reg("eval_str", "Calculate (text)", "Evaluates a math expression (same syntax as eval) and returns a clean string with up to 12 significant digits and no trailing zeros, so 0.1+0.2 gives \"0.3\" and 1/3 gives \"0.333333333333\". Very large or small results use exponent form (1e+25).",
		"str", `math.eval_str("0.1 + 0.2")`, 0, "calculator calc evaluate expression arithmetic formula compute string result clean",
		[]P{rs("expr", "The expression, e.g. 2*(3+4)^2, sqrt(16), 200 + 10%, 15% of 80."), ob("degrees", "Trig functions take and return degrees instead of radians.", false)},
		func(a tools.Args) (any, error) {
			v, err := evalExpr(a.Str("expr"), a.Bool("degrees"))
			if err != nil {
				return nil, err
			}
			return fmtNum(v), nil
		}, accepts(looksLikeMath), kw("calc", "calculate", "="))
}

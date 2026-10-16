package builtin

import (
	"context"
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/core"
)

type calc struct{ base }

func newCalc() *calc { return &calc{base{"calc"}} }

func (c *calc) Ranked() bool       { return true }
func (c *calc) Keywords() []string { return []string{"="} }

var dateLike = regexp.MustCompile(`^\d{4}-\d{1,2}-\d{1,2}`)

func (c *calc) Query(ctx context.Context, q string) []core.Item {
	q = strings.TrimSpace(q)
	forced := strings.HasPrefix(q, "=")
	if forced {
		q = strings.TrimSpace(q[1:])
	}
	if q == "" {
		return nil
	}
	if !forced && (dateLike.MatchString(q) || !looksLikeExpr(q)) {
		return nil
	}
	v, err := evalExpr(q)
	if err != nil {
		if forced {
			return []core.Item{hint("calc", "err", "Cannot evaluate", err.Error())}
		}
		return nil
	}
	res := formatNumber(v)
	it := core.Item{
		ID: "calc:result", Title: res, Subtitle: q + " =", Category: "Calculator",
		Action: copyAction(res),
	}
	if grouped := groupThousands(res); grouped != res {
		it.Alt = append(it.Alt, core.NamedAction{Name: "Copy with separators", Action: copyAction(grouped)})
	}
	if v == math.Trunc(v) && math.Abs(v) < 1<<53 {
		n := int64(v)
		if n >= 0 {
			it.Alt = append(it.Alt,
				core.NamedAction{Name: "Copy as hex", Action: copyAction("0x" + strconv.FormatInt(n, 16))},
				core.NamedAction{Name: "Copy as binary", Action: copyAction("0b" + strconv.FormatInt(n, 2))})
		}
	}
	return []core.Item{it}
}

// formatNumber prints v with at most 12 significant digits and no float noise.
func formatNumber(v float64) string {
	r, _ := strconv.ParseFloat(strconv.FormatFloat(v, 'g', 12, 64), 64)
	if a := math.Abs(r); a == 0 || (a >= 1e-6 && a < 1e15) {
		return strconv.FormatFloat(r, 'f', -1, 64)
	}
	return strconv.FormatFloat(r, 'g', -1, 64)
}

func groupThousands(s string) string {
	neg := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	if strings.ContainsAny(s, "eE") {
		return s
	}
	ip, fp, hasF := strings.Cut(s, ".")
	if len(ip) <= 3 {
		return map[bool]string{true: "-"}[neg] + s
	}
	var b strings.Builder
	for i, ch := range ip {
		if i > 0 && (len(ip)-i)%3 == 0 {
			b.WriteByte(',')
		}
		b.WriteRune(ch)
	}
	out := b.String()
	if hasF {
		out += "." + fp
	}
	if neg {
		out = "-" + out
	}
	return out
}

// ---- lexer ----

type tokKind int

const (
	tNum tokKind = iota
	tID
	tOp
	tLP
	tRP
	tComma
)

type token struct {
	k tokKind
	n float64
	s string
}

var thousandsRe = regexp.MustCompile(`(\d),(\d{3})(\b|$)`)
var funcCallRe = regexp.MustCompile(`[a-z0-9]+\s*\(`)

func lex(s string) ([]token, error) {
	s = strings.ToLower(s)
	s = strings.NewReplacer("×", "*", "÷", "/", "−", "-", "**", "^", " ", " ", "_", "").Replace(s)
	// "1,000,000" is a thousands separator unless it sits in an argument list.
	if !strings.Contains(s, "min") && !strings.Contains(s, "max") && !strings.Contains(s, "pow") && !strings.Contains(s, "round") {
		for i := 0; i < 3; i++ {
			s = thousandsRe.ReplaceAllString(s, "$1$2$3")
		}
	}
	var toks []token
	r := []rune(s)
	for i := 0; i < len(r); {
		ch := r[i]
		switch {
		case ch == ' ' || ch == '\t':
			i++
		case ch == '0' && i+1 < len(r) && (r[i+1] == 'x' || r[i+1] == 'b'):
			base := 16
			if r[i+1] == 'b' {
				base = 2
			}
			j := i + 2
			for j < len(r) && isDigitIn(r[j], base) {
				j++
			}
			if j == i+2 {
				return nil, errors.New("bad number literal")
			}
			n, err := strconv.ParseUint(string(r[i+2:j]), base, 64)
			if err != nil {
				return nil, err
			}
			toks = append(toks, token{k: tNum, n: float64(n)})
			i = j
		case ch >= '0' && ch <= '9' || ch == '.':
			j := i
			for j < len(r) && (r[j] >= '0' && r[j] <= '9' || r[j] == '.') {
				j++
			}
			if j < len(r) && r[j] == 'e' {
				k := j + 1
				if k < len(r) && (r[k] == '+' || r[k] == '-') {
					k++
				}
				if k < len(r) && r[k] >= '0' && r[k] <= '9' {
					for k < len(r) && r[k] >= '0' && r[k] <= '9' {
						k++
					}
					j = k
				}
			}
			n, err := strconv.ParseFloat(string(r[i:j]), 64)
			if err != nil {
				return nil, fmt.Errorf("bad number %q", string(r[i:j]))
			}
			toks = append(toks, token{k: tNum, n: n})
			i = j
		case ch >= 'a' && ch <= 'z':
			j := i
			for j < len(r) && (r[j] >= 'a' && r[j] <= 'z' || r[j] >= '0' && r[j] <= '9') {
				j++
			}
			toks = append(toks, token{k: tID, s: string(r[i:j])})
			i = j
		case strings.ContainsRune("+-*/%^", ch):
			toks = append(toks, token{k: tOp, s: string(ch)})
			i++
		case ch == '(':
			toks = append(toks, token{k: tLP})
			i++
		case ch == ')':
			toks = append(toks, token{k: tRP})
			i++
		case ch == ',':
			toks = append(toks, token{k: tComma})
			i++
		default:
			return nil, fmt.Errorf("unexpected %q", string(ch))
		}
	}
	return toks, nil
}

func isDigitIn(r rune, base int) bool {
	if base == 2 {
		return r == '0' || r == '1'
	}
	return r >= '0' && r <= '9' || r >= 'a' && r <= 'f'
}

// looksLikeExpr is a cheap filter so ordinary words and bare numbers do not
// produce calculator rows.
func looksLikeExpr(q string) bool {
	toks, err := lex(q)
	if err != nil || len(toks) < 2 {
		return false
	}
	hasVal := false
	for i, t := range toks {
		switch t.k {
		case tNum:
			hasVal = true
		case tID:
			if i+1 < len(toks) && toks[i+1].k == tLP {
				hasVal = true // function call
			} else if t.s == "pi" || t.s == "e" || t.s == "tau" {
				hasVal = true
			} else {
				return false // unknown word
			}
		}
	}
	if !hasVal {
		return false
	}
	// a leading sign alone ("-5", "+3") is not an expression
	if len(toks) == 2 && toks[0].k == tOp && toks[1].k == tNum {
		return false
	}
	return true
}

// ---- parser ----

type parser struct {
	t      []token
	pos    int
	pctEnd int // token index just after the last percent postfix
}

// evalExpr evaluates an arithmetic expression.
func evalExpr(s string) (float64, error) {
	toks, err := lex(s)
	if err != nil {
		return 0, err
	}
	if len(toks) == 0 {
		return 0, errors.New("empty expression")
	}
	p := &parser{t: toks, pctEnd: -1}
	v, err := p.add()
	if err != nil {
		return 0, err
	}
	if p.pos != len(p.t) {
		return 0, errors.New("unexpected token")
	}
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, errors.New("result is not a finite number")
	}
	return v, nil
}

func (p *parser) peekOp(ops string) (string, bool) {
	if p.pos < len(p.t) && p.t[p.pos].k == tOp && strings.Contains(ops, p.t[p.pos].s) {
		return p.t[p.pos].s, true
	}
	return "", false
}

func (p *parser) add() (float64, error) {
	l, err := p.mul()
	if err != nil {
		return 0, err
	}
	for {
		op, ok := p.peekOp("+-")
		if !ok {
			return l, nil
		}
		p.pos++
		r, err := p.mul()
		if err != nil {
			return 0, err
		}
		if p.pctEnd == p.pos { // "100 + 10%" means 10% of 100
			r = l * r
		}
		if op == "+" {
			l += r
		} else {
			l -= r
		}
	}
}

func (p *parser) mul() (float64, error) {
	l, err := p.unary()
	if err != nil {
		return 0, err
	}
	for {
		op, ok := p.peekOp("*/%")
		if !ok {
			return l, nil
		}
		p.pos++
		r, err := p.unary()
		if err != nil {
			return 0, err
		}
		switch op {
		case "*":
			l *= r
		case "/":
			if r == 0 {
				return 0, errors.New("division by zero")
			}
			l /= r
		case "%":
			if r == 0 {
				return 0, errors.New("division by zero")
			}
			l = math.Mod(l, r)
		}
	}
}

func (p *parser) unary() (float64, error) {
	if op, ok := p.peekOp("+-"); ok {
		p.pos++
		v, err := p.unary()
		if op == "-" {
			v = -v
		}
		return v, err
	}
	return p.pow()
}

func (p *parser) pow() (float64, error) {
	b, err := p.postfix()
	if err != nil {
		return 0, err
	}
	if _, ok := p.peekOp("^"); ok {
		p.pos++
		e, err := p.unary() // right associative, allows 2^-1
		if err != nil {
			return 0, err
		}
		return math.Pow(b, e), nil
	}
	return b, nil
}

func (p *parser) operandStartsAt(i int) bool {
	if i >= len(p.t) {
		return false
	}
	switch p.t[i].k {
	case tNum, tID, tLP:
		return true
	}
	return false
}

func (p *parser) postfix() (float64, error) {
	v, err := p.primary()
	if err != nil {
		return 0, err
	}
	for p.pos < len(p.t) && p.t[p.pos].k == tOp && p.t[p.pos].s == "%" && !p.operandStartsAt(p.pos+1) {
		p.pos++
		v /= 100
		p.pctEnd = p.pos
	}
	return v, nil
}

var constants = map[string]float64{"pi": math.Pi, "e": math.E, "tau": 2 * math.Pi}

func (p *parser) primary() (float64, error) {
	if p.pos >= len(p.t) {
		return 0, errors.New("unexpected end")
	}
	t := p.t[p.pos]
	switch t.k {
	case tNum:
		p.pos++
		return t.n, nil
	case tLP:
		p.pos++
		v, err := p.add()
		if err != nil {
			return 0, err
		}
		if p.pos >= len(p.t) || p.t[p.pos].k != tRP {
			return 0, errors.New("missing )")
		}
		p.pos++
		return v, nil
	case tID:
		p.pos++
		if p.pos < len(p.t) && p.t[p.pos].k == tLP {
			p.pos++
			var args []float64
			if p.pos < len(p.t) && p.t[p.pos].k == tRP {
				p.pos++
			} else {
				for {
					v, err := p.add()
					if err != nil {
						return 0, err
					}
					args = append(args, v)
					if p.pos < len(p.t) && p.t[p.pos].k == tComma {
						p.pos++
						continue
					}
					if p.pos < len(p.t) && p.t[p.pos].k == tRP {
						p.pos++
						break
					}
					return 0, errors.New("missing )")
				}
			}
			return callFunc(t.s, args)
		}
		if c, ok := constants[t.s]; ok {
			return c, nil
		}
		return 0, fmt.Errorf("unknown name %q", t.s)
	}
	return 0, errors.New("unexpected token")
}

var funcs1 = map[string]func(float64) float64{
	"sqrt": math.Sqrt, "cbrt": math.Cbrt, "sin": math.Sin, "cos": math.Cos, "tan": math.Tan,
	"asin": math.Asin, "acos": math.Acos, "atan": math.Atan,
	"log": math.Log10, "log10": math.Log10, "log2": math.Log2, "ln": math.Log, "exp": math.Exp,
	"abs": math.Abs, "floor": math.Floor, "ceil": math.Ceil, "trunc": math.Trunc,
	"deg": func(x float64) float64 { return x * 180 / math.Pi },
	"rad": func(x float64) float64 { return x * math.Pi / 180 },
}

func callFunc(name string, a []float64) (float64, error) {
	if f, ok := funcs1[name]; ok {
		if len(a) != 1 {
			return 0, fmt.Errorf("%s takes 1 argument", name)
		}
		return f(a[0]), nil
	}
	switch name {
	case "round":
		if len(a) == 1 {
			return math.Round(a[0]), nil
		}
		if len(a) == 2 {
			m := math.Pow(10, a[1])
			return math.Round(a[0]*m) / m, nil
		}
	case "pow":
		if len(a) == 2 {
			return math.Pow(a[0], a[1]), nil
		}
	case "min", "max":
		if len(a) == 0 {
			return 0, fmt.Errorf("%s needs arguments", name)
		}
		r := a[0]
		for _, x := range a[1:] {
			if name == "min" {
				r = math.Min(r, x)
			} else {
				r = math.Max(r, x)
			}
		}
		return r, nil
	default:
		return 0, fmt.Errorf("unknown function %q", name)
	}
	return 0, fmt.Errorf("wrong argument count for %s", name)
}

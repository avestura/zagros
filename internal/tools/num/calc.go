package num

import (
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

const (
	maxExprLen    = 10000
	maxResultBits = 1 << 20
	maxExprDepth  = 200
)

type tok struct {
	kind byte // 'n' number, 'i' function name, 'o' operator, 'e' end
	s    string
	pos  int // 1-based character offset
}

type calc struct {
	toks  []tok
	i     int
	depth int
	base  int
}

func isWordByte(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '_'
}

func lexCalc(expr string) ([]tok, error) {
	var out []tok
	for i := 0; i < len(expr); {
		c := expr[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case isWordByte(c):
			j := i
			for j < len(expr) && isWordByte(expr[j]) {
				j++
			}
			word := expr[i:j]
			k := j
			for k < len(expr) && expr[k] == ' ' {
				k++
			}
			isFunc := k < len(expr) && expr[k] == '(' && (word[0] >= 'a' && word[0] <= 'z' || word[0] >= 'A' && word[0] <= 'Z')
			kind := byte('n')
			if isFunc {
				kind = 'i'
			}
			out = append(out, tok{kind, word, i + 1})
			i = j
		case strings.HasPrefix(expr[i:], "**"), strings.HasPrefix(expr[i:], "<<"), strings.HasPrefix(expr[i:], ">>"):
			out = append(out, tok{'o', expr[i : i+2], i + 1})
			i += 2
		case strings.IndexByte("+-*/%&|^~(),", c) >= 0:
			out = append(out, tok{'o', string(c), i + 1})
			i++
		default:
			return nil, fmt.Errorf("unexpected character %q at position %d", string(rune(c)), i+1)
		}
	}
	out = append(out, tok{'e', "", len(expr) + 1})
	return out, nil
}

func evalInt(expr string, base int) (*big.Int, error) {
	if len(expr) > maxExprLen {
		return nil, errors.New("expression is too long")
	}
	if strings.TrimSpace(expr) == "" {
		return nil, errors.New("empty expression")
	}
	if base == 0 {
		base = 10
	}
	if base < 2 || base > 36 {
		return nil, errors.New("base must be 0 or 2..36")
	}
	toks, err := lexCalc(expr)
	if err != nil {
		return nil, err
	}
	c := &calc{toks: toks, base: base}
	v, err := c.parseOr()
	if err != nil {
		return nil, err
	}
	if t := c.peek(); t.kind != 'e' {
		return nil, fmt.Errorf("unexpected %q at position %d", t.s, t.pos)
	}
	return v, nil
}

func (c *calc) peek() tok { return c.toks[c.i] }
func (c *calc) next() tok { t := c.toks[c.i]; c.i++; return t }
func (c *calc) isOp(s string) bool {
	t := c.peek()
	return t.kind == 'o' && t.s == s
}

func (c *calc) limit(v *big.Int, pos int) (*big.Int, error) {
	if v.BitLen() > maxResultBits {
		return nil, fmt.Errorf("result is too large at position %d (limit %d bits)", pos, maxResultBits)
	}
	return v, nil
}

func (c *calc) level(ops []string, next func() (*big.Int, error), do func(op string, l, r *big.Int, pos int) (*big.Int, error)) (*big.Int, error) {
	l, err := next()
	if err != nil {
		return nil, err
	}
	for {
		t := c.peek()
		found := false
		if t.kind == 'o' {
			for _, o := range ops {
				if t.s == o {
					found = true
				}
			}
		}
		if !found {
			return l, nil
		}
		c.next()
		r, err := next()
		if err != nil {
			return nil, err
		}
		l, err = do(t.s, l, r, t.pos)
		if err != nil {
			return nil, err
		}
		if l, err = c.limit(l, t.pos); err != nil {
			return nil, err
		}
	}
}

func (c *calc) parseOr() (*big.Int, error) {
	c.depth++
	defer func() { c.depth-- }()
	if c.depth > maxExprDepth {
		return nil, errors.New("expression is nested too deeply")
	}
	return c.level([]string{"|"}, c.parseXor, func(_ string, l, r *big.Int, _ int) (*big.Int, error) {
		return new(big.Int).Or(l, r), nil
	})
}

func (c *calc) parseXor() (*big.Int, error) {
	return c.level([]string{"^"}, c.parseAnd, func(_ string, l, r *big.Int, _ int) (*big.Int, error) {
		return new(big.Int).Xor(l, r), nil
	})
}

func (c *calc) parseAnd() (*big.Int, error) {
	return c.level([]string{"&"}, c.parseShift, func(_ string, l, r *big.Int, _ int) (*big.Int, error) {
		return new(big.Int).And(l, r), nil
	})
}

func (c *calc) parseShift() (*big.Int, error) {
	return c.level([]string{"<<", ">>"}, c.parseAdd, func(op string, l, r *big.Int, pos int) (*big.Int, error) {
		if !r.IsInt64() || r.Int64() < 0 || r.Int64() > maxShift {
			return nil, fmt.Errorf("shift count must be 0..%d at position %d", maxShift, pos)
		}
		if op == "<<" {
			return new(big.Int).Lsh(l, uint(r.Int64())), nil
		}
		return new(big.Int).Rsh(l, uint(r.Int64())), nil
	})
}

func (c *calc) parseAdd() (*big.Int, error) {
	return c.level([]string{"+", "-"}, c.parseMul, func(op string, l, r *big.Int, _ int) (*big.Int, error) {
		if op == "+" {
			return new(big.Int).Add(l, r), nil
		}
		return new(big.Int).Sub(l, r), nil
	})
}

func (c *calc) parseMul() (*big.Int, error) {
	return c.level([]string{"*", "/", "%"}, c.parseUnary, func(op string, l, r *big.Int, pos int) (*big.Int, error) {
		switch op {
		case "*":
			return new(big.Int).Mul(l, r), nil
		case "/":
			if r.Sign() == 0 {
				return nil, fmt.Errorf("division by zero at position %d", pos)
			}
			return new(big.Int).Quo(l, r), nil
		}
		if r.Sign() == 0 {
			return nil, fmt.Errorf("modulo by zero at position %d", pos)
		}
		return new(big.Int).Rem(l, r), nil
	})
}

func (c *calc) parseUnary() (*big.Int, error) {
	t := c.peek()
	if t.kind == 'o' && (t.s == "-" || t.s == "+" || t.s == "~") {
		c.next()
		c.depth++
		defer func() { c.depth-- }()
		if c.depth > maxExprDepth {
			return nil, errors.New("expression is nested too deeply")
		}
		v, err := c.parseUnary()
		if err != nil {
			return nil, err
		}
		switch t.s {
		case "-":
			return new(big.Int).Neg(v), nil
		case "~":
			return new(big.Int).Not(v), nil
		}
		return v, nil
	}
	return c.parsePower()
}

func (c *calc) parsePower() (*big.Int, error) {
	b, err := c.parsePrimary()
	if err != nil {
		return nil, err
	}
	if c.isOp("**") {
		t := c.next()
		e, err := c.parseUnary()
		if err != nil {
			return nil, err
		}
		return powChecked(b, e, nil, t.pos)
	}
	return b, nil
}

func powChecked(b, e, m *big.Int, pos int) (*big.Int, error) {
	if e.Sign() < 0 {
		return nil, fmt.Errorf("negative exponent is not an integer at position %d", pos)
	}
	if m != nil {
		if m.Sign() == 0 {
			return nil, fmt.Errorf("modulus is zero at position %d", pos)
		}
		return new(big.Int).Exp(b, e, new(big.Int).Abs(m)), nil
	}
	if b.CmpAbs(big.NewInt(1)) <= 0 {
		if e.Sign() == 0 {
			return big.NewInt(1), nil
		}
		if b.Sign() == 0 || b.Sign() > 0 {
			return new(big.Int).Set(b), nil
		}
		if e.Bit(0) == 0 {
			return big.NewInt(1), nil
		}
		return big.NewInt(-1), nil
	}
	if !e.IsInt64() || e.Int64() > 1<<22 || int64(b.BitLen())*e.Int64() > maxResultBits {
		return nil, fmt.Errorf("power is too large at position %d (limit %d bits)", pos, maxResultBits)
	}
	return new(big.Int).Exp(b, e, nil), nil
}

func (c *calc) parsePrimary() (*big.Int, error) {
	t := c.next()
	switch {
	case t.kind == 'n':
		n, err := parseBig(t.s, c.base)
		if err != nil {
			return nil, fmt.Errorf("invalid number %q at position %d", t.s, t.pos)
		}
		return n, nil
	case t.kind == 'o' && t.s == "(":
		v, err := c.parseOr()
		if err != nil {
			return nil, err
		}
		if !c.isOp(")") {
			return nil, fmt.Errorf("missing ')' at position %d", c.peek().pos)
		}
		c.next()
		return v, nil
	case t.kind == 'i':
		c.next() // '('
		var args []*big.Int
		if c.isOp(")") {
			c.next()
		} else {
			for {
				v, err := c.parseOr()
				if err != nil {
					return nil, err
				}
				args = append(args, v)
				if c.isOp(",") {
					c.next()
					continue
				}
				if c.isOp(")") {
					c.next()
					break
				}
				return nil, fmt.Errorf("expected ',' or ')' at position %d", c.peek().pos)
			}
		}
		return callCalcFunc(strings.ToLower(t.s), args, t.pos)
	case t.kind == 'e':
		return nil, fmt.Errorf("unexpected end of expression at position %d", t.pos)
	}
	return nil, fmt.Errorf("unexpected %q at position %d", t.s, t.pos)
}

func callCalcFunc(name string, args []*big.Int, pos int) (*big.Int, error) {
	need := func(min, max int) error {
		if len(args) < min || (max >= 0 && len(args) > max) {
			return fmt.Errorf("%s() takes %d..%d arguments at position %d", name, min, max, pos)
		}
		return nil
	}
	switch name {
	case "abs":
		if err := need(1, 1); err != nil {
			return nil, err
		}
		return new(big.Int).Abs(args[0]), nil
	case "min", "max":
		if err := need(1, -1); err != nil {
			return nil, err
		}
		best := args[0]
		for _, v := range args[1:] {
			if (name == "min") == (v.Cmp(best) < 0) && v.Cmp(best) != 0 {
				best = v
			}
		}
		return new(big.Int).Set(best), nil
	case "gcd", "lcm":
		if err := need(1, -1); err != nil {
			return nil, err
		}
		acc := new(big.Int).Abs(args[0])
		for _, v := range args[1:] {
			acc = gcdLcm(name == "lcm", acc, v)
		}
		return acc, nil
	case "pow":
		if err := need(2, 3); err != nil {
			return nil, err
		}
		var m *big.Int
		if len(args) == 3 {
			m = args[2]
		}
		return powChecked(args[0], args[1], m, pos)
	case "bits", "bitlen":
		if err := need(1, 1); err != nil {
			return nil, err
		}
		return big.NewInt(int64(args[0].BitLen())), nil
	case "popcount":
		if err := need(1, 1); err != nil {
			return nil, err
		}
		return big.NewInt(int64(popcountAbs(args[0]))), nil
	case "isqrt", "sqrt":
		if err := need(1, 1); err != nil {
			return nil, err
		}
		if args[0].Sign() < 0 {
			return nil, fmt.Errorf("%s of a negative number at position %d", name, pos)
		}
		return new(big.Int).Sqrt(args[0]), nil
	case "fact", "factorial":
		if err := need(1, 1); err != nil {
			return nil, err
		}
		if !args[0].IsInt64() || args[0].Int64() < 0 || args[0].Int64() > 5000 {
			return nil, fmt.Errorf("factorial argument must be 0..5000 at position %d", pos)
		}
		return new(big.Int).MulRange(1, args[0].Int64()), nil
	}
	return nil, fmt.Errorf("unknown function %q at position %d", name, pos)
}

func gcdLcm(lcm bool, a, b *big.Int) *big.Int {
	g := new(big.Int).GCD(nil, nil, new(big.Int).Abs(a), new(big.Int).Abs(b))
	if !lcm {
		return g
	}
	if g.Sign() == 0 {
		return g
	}
	l := new(big.Int).Mul(new(big.Int).Abs(a), new(big.Int).Abs(b))
	return l.Quo(l, g)
}

func popcountAbs(n *big.Int) int {
	c := 0
	for _, w := range new(big.Int).Abs(n).Bits() {
		for x := uint64(w); x != 0; x &= x - 1 {
			c++
		}
	}
	return c
}

// signedWrap interprets the low bits of n as a two's complement integer.
func signedWrap(n *big.Int, bits int) *big.Int {
	v := new(big.Int).And(n, maskOf(bits))
	if v.Bit(bits-1) == 1 {
		v.Sub(v, pow2(bits))
	}
	return v
}

func init() {
	exprParams := func() []P {
		return []P{rs("expr", "Integer expression: literals (255, 0xFF, 0b1010, 0o17, 1_000), + - * / % ** & | ^ ~ << >> and parentheses."), oi("base", "Base of unprefixed literals, 2..36 (0 = decimal). With base 16, write FF instead of 0xFF.", 0)}
	}
	looksExpr := func(s string) bool {
		if len(s) > 500 || strings.TrimSpace(s) == "" || !strings.ContainsAny(s, "+-*/%&|^~<>(") {
			return false
		}
		_, err := evalInt(s, 0)
		return err == nil
	}
	reg("calc_int", "Integer calculator", "Evaluates an integer expression with arbitrary precision and returns the result as a decimal string (or hex/bin/oct with output). Operators + - * / % ** & | ^ ~ << >> and parentheses; / truncates toward zero, ** is right-associative and size-capped, ^ is XOR. Literals may be hex 0xFF, binary 0b1010, octal 0o17 or decimal with underscores; functions abs, min, max, gcd, lcm, pow(a,b[,mod]), bits, popcount, isqrt, fact. Errors carry the character position.",
		"str", `num.calc_int("0xFF + 1")`, 1, "calculator hex calculator binary calculator programmer calc bigint evaluate expression 0xff arithmetic",
		append(exprParams(), outP()),
		func(a tools.Args) (any, error) {
			v, err := evalInt(a.Str("expr"), int(a.Int("base")))
			if err != nil {
				return nil, err
			}
			return outStr(v, a.Str("output"))
		}, kw("calc", "hexcalc", "bincalc"), accepts(looksExpr))

	reg("calc_all", "Integer calculator (all forms)", "Evaluates an integer expression (same syntax as calc_int) and shows the result in dec, hex, bin and oct, with its bit length and the signed int8/16/32/64 and unsigned uint8/16/32/64 interpretations of its low bits.",
		"dict[str, str]", `num.calc_all("0b1010 << 2")`, 2, "calculator hex bin dec oct programmer signed unsigned int32 int64 twos complement evaluate",
		exprParams(),
		func(a tools.Args) (any, error) {
			v, err := evalInt(a.Str("expr"), int(a.Int("base")))
			if err != nil {
				return nil, err
			}
			m := map[string]string{
				"dec": v.String(), "hex": fmtBase(v, 16, true, true, 0), "bin": fmtBase(v, 2, true, false, 0),
				"oct": fmtBase(v, 8, true, false, 0), "bits": fmt.Sprint(v.BitLen()),
			}
			for _, w := range []int{8, 16, 32, 64} {
				m[fmt.Sprintf("int%d", w)] = signedWrap(v, w).String()
				m[fmt.Sprintf("uint%d", w)] = new(big.Int).And(v, maskOf(w)).String()
			}
			return m, nil
		}, kw("calc", "hexcalc", "bincalc"), accepts(looksExpr))

	arith := func(name, title, kind string, base int) {
		verbs := map[string]string{"add": "adds", "sub": "subtracts", "mul": "multiplies", "div": "divides"}
		op := strings.TrimPrefix(strings.TrimPrefix(name, "hex_"), "bin_")
		ret := "str"
		if op == "div" {
			ret = "dict[str, str]"
		}
		desc := "Takes two " + kind + " numbers (prefix, sign and underscores allowed) and " + verbs[op] + " them with arbitrary precision, returning " + kind + " digits without prefix (negative results keep a minus sign)."
		if op == "div" {
			desc = "Takes two " + kind + " numbers and returns a dict with quotient and remainder as " + kind + " digits (division truncates toward zero; dividing by zero is an error)."
		}
		ex := `num.` + name + `("ff", "1")`
		if base == 2 {
			ex = `num.` + name + `("1010", "11")`
		}
		reg(name, title, desc, ret, ex, 10, kind+" calculator "+op+" arithmetic "+map[string]string{"add": "plus sum", "sub": "minus difference", "mul": "times product", "div": "quotient remainder divide"}[op],
			[]P{rs("a", "First "+kind+" number."), rs("b", "Second "+kind+" number.")},
			func(a tools.Args) (any, error) {
				x, err := argBig(a, "a", base)
				if err != nil {
					return nil, err
				}
				y, err := argBig(a, "b", base)
				if err != nil {
					return nil, err
				}
				f := func(n *big.Int) string { return fmtBase(n, base, false, base == 16, 0) }
				switch op {
				case "add":
					return f(new(big.Int).Add(x, y)), nil
				case "sub":
					return f(new(big.Int).Sub(x, y)), nil
				case "mul":
					if x.BitLen()+y.BitLen() > maxResultBits {
						return nil, errors.New("result is too large")
					}
					return f(new(big.Int).Mul(x, y)), nil
				}
				if y.Sign() == 0 {
					return nil, errors.New("division by zero")
				}
				q, r := new(big.Int).QuoRem(x, y, new(big.Int))
				return map[string]string{"quotient": f(q), "remainder": f(r)}, nil
			})
	}
	arith("hex_add", "Hex add", "hex", 16)
	arith("hex_sub", "Hex subtract", "hex", 16)
	arith("hex_mul", "Hex multiply", "hex", 16)
	arith("hex_div", "Hex divide", "hex", 16)
	arith("bin_add", "Binary add", "binary", 2)
	arith("bin_sub", "Binary subtract", "binary", 2)
	arith("bin_mul", "Binary multiply", "binary", 2)
	arith("bin_div", "Binary divide", "binary", 2)
}

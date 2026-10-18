package mathx

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"regexp"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

// roundTo rounds x to `decimals` places (negative = tens, hundreds...) using
// the shortest decimal form of x, so 2.675 rounds half-up to 2.68.
func roundTo(x float64, decimals int, mode string) (float64, error) {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0, errors.New("cannot round a non-finite number")
	}
	if decimals < -300 || decimals > 300 {
		return 0, errors.New("decimals must be between -300 and 300")
	}
	r, _ := new(big.Rat).SetString(strconv.FormatFloat(x, 'f', -1, 64))
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(absInt(decimals))), nil)
	if decimals >= 0 {
		r.Mul(r, new(big.Rat).SetInt(scale))
	} else {
		r.Quo(r, new(big.Rat).SetInt(scale))
	}
	res, err := roundRat(r, mode)
	if err != nil {
		return 0, err
	}
	out := new(big.Rat).SetInt(res)
	if decimals >= 0 {
		out.Quo(out, new(big.Rat).SetInt(scale))
	} else {
		out.Mul(out, new(big.Rat).SetInt(scale))
	}
	f, _ := out.Float64()
	if f == 0 {
		f = 0 // no negative zero
	}
	return f, nil
}

func absInt(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

var roundModes = []string{"half_up", "half_even", "half_down", "floor", "ceil", "trunc"}

func roundRat(r *big.Rat, mode string) (*big.Int, error) {
	num, den := r.Num(), r.Denom()
	q, rem := new(big.Int).DivMod(num, den, new(big.Int)) // floor division, den > 0
	one := big.NewInt(1)
	up := new(big.Int).Add(q, one)
	neg := num.Sign() < 0
	switch strings.ToLower(strings.TrimSpace(mode)) {
	case "", "half_up", "half-up", "round":
		return halfPick(q, up, rem, den, func() bool { return !neg }), nil
	case "half_down", "half-down":
		return halfPick(q, up, rem, den, func() bool { return neg }), nil
	case "half_even", "half-even", "bankers":
		return halfPick(q, up, rem, den, func() bool { return q.Bit(0) == 1 }), nil
	case "floor":
		return q, nil
	case "ceil":
		if rem.Sign() != 0 {
			return up, nil
		}
		return q, nil
	case "trunc":
		if neg && rem.Sign() != 0 {
			return up, nil
		}
		return q, nil
	}
	return nil, fmt.Errorf("mode must be one of %s", strings.Join(roundModes, ", "))
}

func halfPick(q, up, rem, den *big.Int, tieUp func() bool) *big.Int {
	switch new(big.Int).Lsh(rem, 1).Cmp(den) {
	case -1:
		return q
	case 1:
		return up
	}
	if tieUp() {
		return up
	}
	return q
}

func roundToMultiple(x, m float64, mode string) (float64, error) {
	if m == 0 || math.IsNaN(m) || math.IsInf(m, 0) || math.IsNaN(x) || math.IsInf(x, 0) {
		return 0, errors.New("multiple must be a non-zero finite number")
	}
	rx, _ := new(big.Rat).SetString(strconv.FormatFloat(x, 'f', -1, 64))
	rm, _ := new(big.Rat).SetString(strconv.FormatFloat(math.Abs(m), 'f', -1, 64))
	n, err := roundRat(new(big.Rat).Quo(rx, rm), mode)
	if err != nil {
		return 0, err
	}
	f, _ := new(big.Rat).Mul(new(big.Rat).SetInt(n), rm).Float64()
	return f, nil
}

// ---- fractions ----

var fracRe = regexp.MustCompile(`^[-+]?(\d+(\.\d*)?|\.\d+)(/\d+)?$`)
var mixedRe = regexp.MustCompile(`^([-+]?)(\d+)\s+(\d+)/(\d+)$`)

func parseFrac(s string) (*big.Rat, error) {
	t := strings.TrimSpace(s)
	if len(t) > 400 {
		return nil, errors.New("fraction is too long")
	}
	if m := mixedRe.FindStringSubmatch(t); m != nil {
		w, _ := new(big.Rat).SetString(m[2])
		f, ok := new(big.Rat).SetString(m[3] + "/" + m[4])
		if !ok {
			return nil, fmt.Errorf("%q has a zero denominator", t)
		}
		w.Add(w, f)
		if m[1] == "-" {
			w.Neg(w)
		}
		return w, nil
	}
	t = strings.ReplaceAll(t, " ", "")
	if !fracRe.MatchString(t) {
		return nil, fmt.Errorf("%q is not a fraction like 3/4, 0.75 or 1 1/2", s)
	}
	if i := strings.IndexByte(t, '/'); i >= 0 {
		if strings.Trim(t[i+1:], "0") == "" {
			return nil, fmt.Errorf("%q has a zero denominator", s)
		}
	}
	r, ok := new(big.Rat).SetString(t)
	if !ok {
		return nil, fmt.Errorf("%q is not a fraction", s)
	}
	return r, nil
}

func fracStr(r *big.Rat) string {
	if r.IsInt() {
		return r.Num().String()
	}
	return r.Num().String() + "/" + r.Denom().String()
}

func mixedStr(r *big.Rat) string {
	if r.IsInt() {
		return r.Num().String()
	}
	neg := r.Sign() < 0
	a := new(big.Rat).Abs(r)
	w, rem := new(big.Int).QuoRem(a.Num(), a.Denom(), new(big.Int))
	s := ""
	if w.Sign() != 0 {
		s = w.String() + " "
	}
	s += rem.String() + "/" + a.Denom().String()
	if neg {
		s = "-" + s
	}
	return s
}

func limitDenominator(r *big.Rat, max *big.Int) *big.Rat {
	if r.Denom().Cmp(max) <= 0 {
		return new(big.Rat).Set(r)
	}
	neg := r.Sign() < 0
	abs := new(big.Rat).Abs(r)
	n, d := new(big.Int).Set(abs.Num()), new(big.Int).Set(abs.Denom())
	p0, q0, p1, q1 := big.NewInt(0), big.NewInt(1), big.NewInt(1), big.NewInt(0)
	for d.Sign() != 0 {
		a := new(big.Int).Quo(n, d)
		q2 := new(big.Int).Add(q0, new(big.Int).Mul(a, q1))
		if q2.Cmp(max) > 0 {
			break
		}
		p2 := new(big.Int).Add(p0, new(big.Int).Mul(a, p1))
		p0, q0, p1, q1 = p1, q1, p2, q2
		n, d = d, new(big.Int).Sub(n, new(big.Int).Mul(a, d))
	}
	k := new(big.Int).Quo(new(big.Int).Sub(max, q0), q1)
	b1 := new(big.Rat).SetFrac(new(big.Int).Add(p0, new(big.Int).Mul(k, p1)), new(big.Int).Add(q0, new(big.Int).Mul(k, q1)))
	b2 := new(big.Rat).SetFrac(p1, q1)
	dist := func(x *big.Rat) *big.Rat { return new(big.Rat).Abs(new(big.Rat).Sub(x, abs)) }
	res := b1
	if dist(b2).Cmp(dist(b1)) <= 0 {
		res = b2
	}
	if neg {
		res = new(big.Rat).Neg(res)
	}
	return res
}

// fraction expression evaluator (exact): numbers, + - * / and parentheses.
type fparser struct {
	s     string
	i     int
	depth int
}

func (p *fparser) ws() {
	for p.i < len(p.s) && p.s[p.i] == ' ' {
		p.i++
	}
}

func (p *fparser) sum() (*big.Rat, error) {
	p.depth++
	defer func() { p.depth-- }()
	if p.depth > 100 {
		return nil, errors.New("expression is nested too deeply")
	}
	l, err := p.term()
	if err != nil {
		return nil, err
	}
	for {
		p.ws()
		if p.i >= len(p.s) || (p.s[p.i] != '+' && p.s[p.i] != '-') {
			return l, nil
		}
		op := p.s[p.i]
		p.i++
		r, err := p.term()
		if err != nil {
			return nil, err
		}
		if op == '+' {
			l = new(big.Rat).Add(l, r)
		} else {
			l = new(big.Rat).Sub(l, r)
		}
	}
}

func (p *fparser) term() (*big.Rat, error) {
	l, err := p.unary()
	if err != nil {
		return nil, err
	}
	for {
		p.ws()
		if p.i >= len(p.s) || (p.s[p.i] != '*' && p.s[p.i] != '/' && p.s[p.i] != '÷') {
			return l, nil
		}
		op := p.s[p.i]
		p.i++
		r, err := p.unary()
		if err != nil {
			return nil, err
		}
		if op == '*' {
			l = new(big.Rat).Mul(l, r)
		} else {
			if r.Sign() == 0 {
				return nil, fmt.Errorf("division by zero at position %d", p.i)
			}
			l = new(big.Rat).Quo(l, r)
		}
		if l.Num().BitLen() > 100000 || l.Denom().BitLen() > 100000 {
			return nil, errors.New("result is too large")
		}
	}
}

func (p *fparser) unary() (*big.Rat, error) {
	p.ws()
	if p.i < len(p.s) && (p.s[p.i] == '-' || p.s[p.i] == '+') {
		neg := p.s[p.i] == '-'
		p.i++
		p.depth++
		defer func() { p.depth-- }()
		if p.depth > 100 {
			return nil, errors.New("expression is nested too deeply")
		}
		v, err := p.unary()
		if err != nil {
			return nil, err
		}
		if neg {
			v = new(big.Rat).Neg(v)
		}
		return v, nil
	}
	if p.i < len(p.s) && p.s[p.i] == '(' {
		p.i++
		v, err := p.sum()
		if err != nil {
			return nil, err
		}
		p.ws()
		if p.i >= len(p.s) || p.s[p.i] != ')' {
			return nil, fmt.Errorf("missing ')' at position %d", p.i+1)
		}
		p.i++
		return v, nil
	}
	start := p.i
	for p.i < len(p.s) && (p.s[p.i] >= '0' && p.s[p.i] <= '9' || p.s[p.i] == '.') {
		p.i++
	}
	if start == p.i {
		return nil, fmt.Errorf("expected a number at position %d", start+1)
	}
	r, ok := new(big.Rat).SetString(p.s[start:p.i])
	if !ok {
		return nil, fmt.Errorf("invalid number %q at position %d", p.s[start:p.i], start+1)
	}
	return r, nil
}

// ---- equations ----

func complexStr(re, im float64) string {
	if math.Abs(im) < 1e-15*math.Max(1, math.Abs(re)) && im == 0 {
		return fmtNum(re)
	}
	sign := "+"
	if im < 0 {
		sign, im = "-", -im
	}
	is := fmtNum(im)
	if is == "1" {
		is = ""
	}
	if re == 0 {
		if sign == "+" {
			return is + "i"
		}
		return "-" + is + "i"
	}
	return fmtNum(re) + sign + is + "i"
}

func quadratic(a, b, c float64) ([]string, float64, error) {
	if a == 0 {
		if b == 0 {
			return nil, 0, errors.New("a and b are both zero: not an equation in x")
		}
		return []string{fmtNum(-c / b)}, 0, nil
	}
	d := b*b - 4*a*c
	if math.IsInf(d, 0) || math.IsNaN(d) {
		return nil, 0, errors.New("coefficients are too large")
	}
	if d < 0 {
		re, im := -b/(2*a), math.Sqrt(-d)/(2*math.Abs(a))
		return []string{complexStr(re, -im), complexStr(re, im)}, d, nil
	}
	sq := math.Sqrt(d)
	q := -(b + math.Copysign(sq, b)) / 2
	var x1, x2 float64
	if q == 0 { // b == 0 and d == 0
		x1, x2 = 0, 0
	} else {
		x1, x2 = q/a, c/q
	}
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if d == 0 {
		return []string{fmtNum(x1)}, d, nil
	}
	return []string{fmtNum(x1), fmtNum(x2)}, d, nil
}

func solveSystem(m []float64, n int) ([]float64, error) {
	if len(m) != n*(n+1) {
		return nil, fmt.Errorf("expected %d numbers (augmented matrix, row by row), got %d", n*(n+1), len(m))
	}
	a := make([][]float64, n)
	for i := range a {
		a[i] = append([]float64(nil), m[i*(n+1):(i+1)*(n+1)]...)
	}
	for c := 0; c < n; c++ {
		piv := c
		for r := c + 1; r < n; r++ {
			if math.Abs(a[r][c]) > math.Abs(a[piv][c]) {
				piv = r
			}
		}
		if math.Abs(a[piv][c]) < 1e-12 {
			return nil, errors.New("the system has no unique solution (determinant is 0)")
		}
		a[c], a[piv] = a[piv], a[c]
		for r := c + 1; r < n; r++ {
			f := a[r][c] / a[c][c]
			for k := c; k <= n; k++ {
				a[r][k] -= f * a[c][k]
			}
		}
	}
	x := make([]float64, n)
	for i := n - 1; i >= 0; i-- {
		s := a[i][n]
		for k := i + 1; k < n; k++ {
			s -= a[i][k] * x[k]
		}
		x[i] = s / a[i][i]
		if math.IsNaN(x[i]) || math.IsInf(x[i], 0) {
			return nil, errors.New("the system has no finite solution")
		}
	}
	for i := range x {
		if math.Abs(x[i]) < 1e-12 {
			x[i] = 0
		}
	}
	return x, nil
}

var polyTerm = regexp.MustCompile(`([+-]?)(\d*\.?\d*)(x²|x)?`)

// parsePoly reads "x^2 - 3x + 2 = 0" style input into a, b, c.
func parsePoly(s string) (a, b, c float64, err error) {
	t := strings.ToLower(strings.NewReplacer(" ", "", "^2", "²", "**2", "²", "*", "", "−", "-").Replace(s))
	if i := strings.IndexByte(t, '='); i >= 0 {
		if strings.Trim(t[i+1:], "0.") != "" {
			return 0, 0, 0, errors.New("the right side must be 0")
		}
		t = t[:i]
	}
	pos := 0
	for pos < len(t) {
		loc := polyTerm.FindStringSubmatchIndex(t[pos:])
		if loc == nil || loc[1] == 0 {
			return 0, 0, 0, fmt.Errorf("cannot read %q at position %d", t[pos:], pos+1)
		}
		sign, num, v := t[pos+loc[2]:pos+loc[3]], t[pos+loc[4]:pos+loc[5]], ""
		if loc[6] >= 0 {
			v = t[pos+loc[6] : pos+loc[7]]
		}
		pos += loc[1]
		coef := 1.0
		if num != "" {
			if coef, err = strconv.ParseFloat(num, 64); err != nil {
				return 0, 0, 0, fmt.Errorf("bad number %q", num)
			}
		} else if v == "" {
			return 0, 0, 0, fmt.Errorf("cannot read the equation near position %d", pos)
		}
		if sign == "-" {
			coef = -coef
		}
		switch v {
		case "x²":
			a += coef
		case "x":
			b += coef
		default:
			c += coef
		}
	}
	return
}

// ---- unit helpers ----

func tempUnit(u string) (string, error) {
	t := strings.ToLower(strings.TrimSpace(strings.NewReplacer("°", "", "º", "", "deg", "", "degrees", "", " ", "").Replace(u)))
	switch t {
	case "c", "celsius", "centigrade":
		return "c", nil
	case "f", "fahrenheit":
		return "f", nil
	case "k", "kelvin":
		return "k", nil
	case "r", "ra", "rankine":
		return "r", nil
	}
	return "", fmt.Errorf("unknown temperature unit %q (use c, f, k or r)", u)
}

func toKelvin(v float64, u string) float64 {
	switch u {
	case "c":
		return v + 273.15
	case "f":
		return (v-32)*5/9 + 273.15
	case "r":
		return v * 5 / 9
	}
	return v
}

func fromKelvin(k float64, u string) float64 {
	switch u {
	case "c":
		return k - 273.15
	case "f":
		return (k-273.15)*9/5 + 32
	case "r":
		return k * 9 / 5
	}
	return k
}

func convertTemp(v float64, from, to string) (float64, error) {
	f, err := tempUnit(from)
	if err != nil {
		return 0, err
	}
	t, err := tempUnit(to)
	if err != nil {
		return 0, err
	}
	if f == t {
		return v, nil
	}
	k := toKelvin(v, f)
	if k < -1e-9 {
		return 0, errors.New("temperature is below absolute zero")
	}
	r := fromKelvin(k, t)
	r, _ = roundTo(r, 10, "half_up")
	return r, nil
}

var tempTextRe = regexp.MustCompile(`(?i)^\s*(-?\d+(?:\.\d+)?)\s*°?\s*(c|f|k|r|celsius|fahrenheit|kelvin|rankine)\s*$`)

func parseRatioParts(s string) ([]*big.Rat, error) {
	parts := regexp.MustCompile(`[:/xX,\s]+`).Split(strings.TrimSpace(s), -1)
	if len(parts) < 2 || len(parts) > 20 {
		return nil, errors.New(`ratio needs 2..20 parts like "16:9"`)
	}
	out := make([]*big.Rat, len(parts))
	for i, p := range parts {
		r, ok := new(big.Rat).SetString(p)
		if !ok || len(p) > 60 || strings.ContainsAny(p, "eE") || r.Sign() < 0 {
			return nil, fmt.Errorf("%q is not a valid ratio part", p)
		}
		out[i] = r
	}
	return out, nil
}

func simplifyRatio(parts []*big.Rat) ([]*big.Int, error) {
	l := big.NewInt(1)
	for _, p := range parts {
		d := p.Denom()
		g := new(big.Int).GCD(nil, nil, l, d)
		l.Mul(l, new(big.Int).Quo(d, g))
	}
	ints := make([]*big.Int, len(parts))
	g := new(big.Int)
	for i, p := range parts {
		ints[i] = new(big.Int).Mul(p.Num(), new(big.Int).Quo(l, p.Denom()))
		g.GCD(nil, nil, g, ints[i])
	}
	if g.Sign() == 0 {
		return nil, errors.New("ratio parts are all zero")
	}
	for i := range ints {
		ints[i].Quo(ints[i], g)
	}
	return ints, nil
}

func init() {
	reg("round_to", "Round", "Rounds x to a number of decimals (negative rounds to tens, hundreds) using its decimal form, so 2.675 gives 2.68 with half_up. Modes: half_up (default, away from zero on ties), half_even (banker's), half_down, floor, ceil, trunc.",
		"float", `math.round_to(2.675, 2)`, 70, "round rounding decimals half even bankers floor ceil truncate places",
		[]P{rf("x", "Number to round."), oi("decimals", "Decimal places (-300..300).", 0), os("mode", "Rounding mode.", "half_up", roundModes...)},
		func(a tools.Args) (any, error) { return roundTo(a.Float("x"), int(a.Int("decimals")), a.Str("mode")) })
	reg("round_to_multiple", "Round to multiple", "Rounds x to the nearest multiple of m (round_to_multiple(17, 5) = 15, (18, 5) = 20, 0.1 steps work exactly). Same modes as round_to. m must not be zero.",
		"float", `math.round_to_multiple(18, 5)`, 71, "round nearest multiple step increment quantize snap",
		[]P{rf("x", "Number to round."), rf("multiple", "The step."), os("mode", "Rounding mode.", "half_up", roundModes...)},
		func(a tools.Args) (any, error) {
			return roundToMultiple(a.Float("x"), a.Float("multiple"), a.Str("mode"))
		})
	reg("clamp", "Clamp", "Limits x to the range lo..hi (clamp(15, 0, 10) = 10). lo must not exceed hi.",
		"float", `math.clamp(15, 0, 10)`, 72, "clamp limit bound min max constrain range",
		[]P{rf("x", "Value."), rf("lo", "Lower bound."), rf("hi", "Upper bound.")},
		func(a tools.Args) (any, error) {
			if a.Float("lo") > a.Float("hi") {
				return nil, errors.New("lo must not be greater than hi")
			}
			return math.Min(math.Max(a.Float("x"), a.Float("lo")), a.Float("hi")), nil
		})
	reg("lerp", "Linear interpolation", "Linear interpolation a + (b-a)*t (lerp(10, 20, 0.25) = 12.5); t outside 0..1 extrapolates.",
		"float", `math.lerp(10, 20, 0.25)`, 73, "lerp linear interpolation mix blend between",
		[]P{rf("a", "Start."), rf("b", "End."), rf("t", "Position, 0 = a, 1 = b.")},
		func(a tools.Args) (any, error) {
			return finite("lerp", a.Float("a")+(a.Float("b")-a.Float("a"))*a.Float("t"))
		})
	reg("inverse_lerp", "Inverse lerp", "Returns where x sits between a and b as a fraction ((x-a)/(b-a)). a and b must differ.",
		"float", `math.inverse_lerp(10, 20, 12.5)`, 74, "inverse lerp unlerp fraction position progress",
		[]P{rf("a", "Start."), rf("b", "End."), rf("x", "Value.")},
		func(a tools.Args) (any, error) {
			if a.Float("a") == a.Float("b") {
				return nil, errors.New("a and b must differ")
			}
			return (a.Float("x") - a.Float("a")) / (a.Float("b") - a.Float("a")), nil
		})
	reg("map_range", "Map range", "Maps x from the range in_lo..in_hi to out_lo..out_hi linearly (map_range(5, 0, 10, 0, 100) = 50). in_lo and in_hi must differ; values outside are extrapolated.",
		"float", `math.map_range(5, 0, 10, 0, 100)`, 75, "map range remap rescale scale convert interval",
		[]P{rf("x", "Value."), rf("in_lo", "Input range start."), rf("in_hi", "Input range end."), rf("out_lo", "Output range start."), rf("out_hi", "Output range end.")},
		func(a tools.Args) (any, error) {
			if a.Float("in_lo") == a.Float("in_hi") {
				return nil, errors.New("in_lo and in_hi must differ")
			}
			t := (a.Float("x") - a.Float("in_lo")) / (a.Float("in_hi") - a.Float("in_lo"))
			return finite("map_range", a.Float("out_lo")+(a.Float("out_hi")-a.Float("out_lo"))*t)
		})

	reg("solve_linear", "Solve linear equation", "Solves a*x + b = 0 for x (solve_linear(2, -6) = 3). a must not be zero.",
		"float", `math.solve_linear(2, -6)`, 76, "solve linear equation ax+b root x",
		[]P{rf("a", "Coefficient of x."), rf("b", "Constant term.")},
		func(a tools.Args) (any, error) {
			if a.Float("a") == 0 {
				return nil, errors.New("a must not be zero")
			}
			return finite("solve_linear", -a.Float("b")/a.Float("a"))
		})
	reg("solve_quadratic", "Solve quadratic", "Solves a*x^2 + b*x + c = 0 and returns the roots as strings, real ones ascending (a double root once) and complex ones like \"1+2i\", \"1-2i\". With a=0 it solves the linear equation. x^2-3x+2 gives [\"1\", \"2\"].",
		"list[str]", `math.solve_quadratic(1, -3, 2)`, 5, "quadratic equation solve roots discriminant parabola complex x2 formula",
		[]P{rf("a", "Coefficient of x^2."), rf("b", "Coefficient of x."), rf("c", "Constant term.")},
		func(a tools.Args) (any, error) {
			r, _, err := quadratic(a.Float("a"), a.Float("b"), a.Float("c"))
			return r, err
		})
	reg("quadratic_text", "Quadratic equation", "Solves a quadratic written as text, either \"x^2 - 3x + 2 = 0\" or three coefficients \"1 -3 2\". Returns a, b, c, discriminant and the roots x1, x2 (complex roots as strings).",
		"dict[str, str]", `math.quadratic_text("x^2 - 3x + 2 = 0")`, 6, "quadratic equation solve roots discriminant parabola x2 formula",
		[]P{rs("text", "Equation or coefficients a b c.")},
		func(a tools.Args) (any, error) {
			s := a.Str("text")
			var qa, qb, qc float64
			var err error
			if strings.ContainsAny(s, "xX") {
				qa, qb, qc, err = parsePoly(s)
			} else {
				n := extractNumbers(s)
				if len(n) != 3 {
					return nil, errors.New("expected three coefficients a b c or an equation in x")
				}
				qa, qb, qc = n[0], n[1], n[2]
			}
			if err != nil {
				return nil, err
			}
			roots, d, err := quadratic(qa, qb, qc)
			if err != nil {
				return nil, err
			}
			m := dict("a", qa, "b", qb, "c", qc, "discriminant", d, "x1", roots[0])
			m["x2"] = roots[len(roots)-1]
			return m, nil
		}, accepts(func(s string) bool {
			if len(s) > 80 {
				return false
			}
			if strings.ContainsAny(s, "xX") {
				_, _, _, err := parsePoly(s)
				return err == nil && strings.Contains(strings.ToLower(s), "x")
			}
			return len(extractNumbers(s)) == 3
		}), kw("quadratic", "solve"))
	reg("solve_linear_system_2x2", "Solve 2x2 system", "Solves two linear equations in x and y given as the augmented matrix [a1, b1, c1, a2, b2, c2] meaning a1*x+b1*y=c1 and a2*x+b2*y=c2. Returns [x, y]; singular systems are an error.",
		"list[float]", `math.solve_linear_system_2x2([1, 1, 3, 1, -1, 1])`, 77, "linear system equations simultaneous 2x2 solve cramer gauss",
		[]P{rl("matrix", "Six numbers: a1, b1, c1, a2, b2, c2.")},
		func(a tools.Args) (any, error) { return solveSystem(a.Floats("matrix"), 2) })
	reg("solve_linear_system_3x3", "Solve 3x3 system", "Solves three linear equations in x, y, z given as the row-by-row augmented matrix (12 numbers: a, b, c, d for each equation a*x+b*y+c*z=d). Returns [x, y, z]; singular systems are an error.",
		"list[float]", `math.solve_linear_system_3x3([1,1,1,6, 0,2,5,-4, 2,5,-1,27])`, 78, "linear system equations simultaneous 3x3 solve gauss elimination",
		[]P{rl("matrix", "Twelve numbers, four per equation.")},
		func(a tools.Args) (any, error) { return solveSystem(a.Floats("matrix"), 3) })

	reg("simplify_fraction", "Simplify fraction", "Reduces n/d to lowest terms with a positive denominator (simplify_fraction(6, -8) = \"-3/4\"; whole results print without a denominator). d must not be zero.",
		"str", `math.simplify_fraction(6, 8)`, 79, "simplify fraction reduce lowest terms gcd numerator denominator",
		[]P{ri("n", "Numerator."), ri("d", "Denominator.")},
		func(a tools.Args) (any, error) {
			if a.Int("d") == 0 {
				return nil, errors.New("denominator must not be zero")
			}
			return fracStr(new(big.Rat).SetFrac(big.NewInt(a.Int("n")), big.NewInt(a.Int("d")))), nil
		})
	fop := func(name, title, verb string, f func(x, y *big.Rat) (*big.Rat, error)) {
		reg(name, title, "Exact fraction "+verb+" of a and b (each \"3/4\", \"0.75\", \"2\" or \"1 1/2\"). The result is in lowest terms, printed as n/d or a whole number.",
			"str", `math.`+name+`("1/2", "1/3")`, 80, "fraction "+verb+" exact rational arithmetic",
			[]P{rs("a", "First fraction."), rs("b", "Second fraction.")},
			func(x tools.Args) (any, error) {
				a, err := parseFrac(x.Str("a"))
				if err != nil {
					return nil, err
				}
				b, err := parseFrac(x.Str("b"))
				if err != nil {
					return nil, err
				}
				r, err := f(a, b)
				if err != nil {
					return nil, err
				}
				return fracStr(r), nil
			})
	}
	fop("fraction_add", "Add fractions", "addition", func(x, y *big.Rat) (*big.Rat, error) { return new(big.Rat).Add(x, y), nil })
	fop("fraction_sub", "Subtract fractions", "subtraction", func(x, y *big.Rat) (*big.Rat, error) { return new(big.Rat).Sub(x, y), nil })
	fop("fraction_mul", "Multiply fractions", "multiplication", func(x, y *big.Rat) (*big.Rat, error) { return new(big.Rat).Mul(x, y), nil })
	fop("fraction_div", "Divide fractions", "division", func(x, y *big.Rat) (*big.Rat, error) {
		if y.Sign() == 0 {
			return nil, errors.New("division by zero")
		}
		return new(big.Rat).Quo(x, y), nil
	})
	reg("fraction_calc", "Fraction calculator", "Evaluates an expression exactly with fractions: numbers, decimals, + - * / and parentheses (\"3/4 + 1/6\" gives \"11/12\"). Division is exact, never floating point.",
		"str", `math.fraction_calc("3/4 + 1/6")`, 7, "fraction calculator exact rational add subtract multiply divide frac",
		[]P{rs("expr", "Expression such as 3/4 + 1/6 or (1/2)*(2/3).")},
		func(a tools.Args) (any, error) {
			if len(a.Str("expr")) > 2000 {
				return nil, errors.New("expression is too long")
			}
			p := &fparser{s: strings.TrimSpace(a.Str("expr"))}
			r, err := p.sum()
			if err != nil {
				return nil, err
			}
			p.ws()
			if p.i < len(p.s) {
				return nil, fmt.Errorf("unexpected %q at position %d", string(p.s[p.i]), p.i+1)
			}
			return fracStr(r) + "  (" + mixedStr(r) + ")", nil
		}, accepts(func(s string) bool {
			if len(s) > 200 || !strings.Contains(s, "/") {
				return false
			}
			p := &fparser{s: strings.TrimSpace(s)}
			_, err := p.sum()
			p.ws()
			return err == nil && p.i >= len(p.s)
		}), kw("frac", "fraction"))
	reg("decimal_to_fraction", "Decimal to fraction", "Converts a decimal to a fraction. With max_denominator=0 (default) the conversion is exact for the decimal text (0.125 = 1/8, 0.333 = 333/1000); a positive max_denominator finds the closest fraction up to that denominator (3.14159 with 1000 = 355/113).",
		"str", `math.decimal_to_fraction("0.125")`, 81, "decimal to fraction convert rational numerator denominator",
		[]P{rs("x", "Decimal number or fraction."), oi("max_denominator", "Largest denominator (0 = exact).", 0)},
		func(a tools.Args) (any, error) {
			r, err := parseFrac(a.Str("x"))
			if err != nil {
				return nil, err
			}
			if md := a.Int("max_denominator"); md > 0 {
				r = limitDenominator(r, big.NewInt(md))
			} else if md < 0 {
				return nil, errors.New("max_denominator must not be negative")
			}
			return fracStr(r), nil
		}, accepts(func(s string) bool { _, err := parseFrac(s); return err == nil && strings.Contains(s, ".") }))
	reg("mixed_number", "Mixed number", "Writes an improper fraction or decimal as a mixed number (\"7/2\" gives \"3 1/2\", 2.75 gives \"2 3/4\", -7/2 gives \"-3 1/2\").",
		"str", `math.mixed_number("7/2")`, 82, "mixed number improper fraction whole part convert",
		[]P{rs("x", "Fraction like 7/2 or a decimal.")},
		func(a tools.Args) (any, error) {
			r, err := parseFrac(a.Str("x"))
			if err != nil {
				return nil, err
			}
			return mixedStr(r), nil
		}, accepts(func(s string) bool { _, err := parseFrac(s); return err == nil && strings.Contains(s, "/") }))
	reg("ratio_simplify", "Simplify ratio", "Reduces a ratio to the smallest whole numbers: \"1920:1080\" gives \"16:9\", \"4:6:8\" gives \"2:3:4\", \"0.5:1.5\" gives \"1:3\". Separators : / x , or spaces.",
		"str", `math.ratio_simplify("1920:1080")`, 8, "ratio simplify aspect ratio reduce proportion 16:9 screen",
		[]P{rs("ratio", "Ratio such as 16:9.")},
		func(a tools.Args) (any, error) {
			parts, err := parseRatioParts(a.Str("ratio"))
			if err != nil {
				return nil, err
			}
			ints, err := simplifyRatio(parts)
			if err != nil {
				return nil, err
			}
			s := make([]string, len(ints))
			for i, v := range ints {
				s[i] = v.String()
			}
			return strings.Join(s, ":"), nil
		}, accepts(func(s string) bool {
			if len(s) > 60 || !strings.ContainsAny(s, ":/xX") {
				return false
			}
			_, err := parseRatioParts(s)
			return err == nil
		}), kw("ratio", "aspect"))
	reg("scale_ratio", "Scale ratio", "Scales a ratio so that one part has a known value and returns all parts: scale_ratio(\"16:9\", 1920) gives \"1920:1080\"; position picks which part (1-based) the value belongs to.",
		"str", `math.scale_ratio("16:9", 1920)`, 83, "scale ratio aspect resize proportion width height dimension",
		[]P{rs("ratio", "Ratio such as 16:9."), rf("known", "Value of the chosen part."), oi("position", "1-based index of the part that equals known.", 1)},
		func(a tools.Args) (any, error) {
			parts, err := parseRatioParts(a.Str("ratio"))
			if err != nil {
				return nil, err
			}
			pos, err := intIn("position", a.Int("position"), 1, int64(len(parts)))
			if err != nil {
				return nil, err
			}
			ref, _ := parts[pos-1].Float64()
			if ref == 0 {
				return nil, errors.New("the chosen ratio part is zero")
			}
			out := make([]string, len(parts))
			for i, p := range parts {
				f, _ := p.Float64()
				out[i] = fmtNum(f / ref * a.Float("known"))
			}
			return strings.Join(out, ":"), nil
		})

	reg("temperature", "Temperature convert", "Converts a temperature between c (Celsius), f (Fahrenheit), k (Kelvin) and r (Rankine); full unit names and a degree sign are accepted. Values below absolute zero are an error.",
		"float", `math.temperature(100, "c", "f")`, 84, "temperature convert celsius fahrenheit kelvin rankine degrees",
		[]P{rf("value", "Temperature."), rs("from_unit", "Source unit: c, f, k or r."), rs("to_unit", "Target unit: c, f, k or r.")},
		func(a tools.Args) (any, error) {
			return convertTemp(a.Float("value"), a.Str("from_unit"), a.Str("to_unit"))
		})
	reg("temperature_text", "Temperature", "Converts text like \"72f\", \"-40 C\" or \"300K\" to all four scales and returns a dict with c, f, k and r.",
		"dict[str, str]", `math.temperature_text("72f")`, 9, "temperature convert celsius fahrenheit kelvin rankine degrees weather",
		[]P{rs("text", "Number with a unit letter, e.g. 72f.")},
		func(a tools.Args) (any, error) {
			m := tempTextRe.FindStringSubmatch(a.Str("text"))
			if m == nil {
				return nil, errors.New(`expected a number and a unit, e.g. "72f" or "20 C"`)
			}
			v, _ := strconv.ParseFloat(m[1], 64)
			out := map[string]string{}
			for _, u := range []string{"c", "f", "k", "r"} {
				r, err := convertTemp(v, m[2], u)
				if err != nil {
					return nil, err
				}
				out[u] = fmtNum(r)
			}
			return out, nil
		}, accepts(func(s string) bool { return tempTextRe.MatchString(s) }), kw("temp", "celsius", "fahrenheit"))

	reg("degrees_to_radians", "Degrees to radians", "Converts degrees to radians (180 gives 3.14159...).",
		"float", `math.degrees_to_radians(180)`, 85, "degrees radians convert angle deg rad",
		[]P{rf("degrees", "Angle in degrees.")},
		func(a tools.Args) (any, error) { return a.Float("degrees") * math.Pi / 180, nil })
	reg("radians_to_degrees", "Radians to degrees", "Converts radians to degrees (pi gives 180).",
		"float", `math.radians_to_degrees(3.141592653589793)`, 86, "radians degrees convert angle deg rad",
		[]P{rf("radians", "Angle in radians.")},
		func(a tools.Args) (any, error) { return a.Float("radians") * 180 / math.Pi, nil })
	reg("polar_to_cartesian", "Polar to Cartesian", "Converts polar (r, theta) to Cartesian x and y. theta is in radians unless degrees=true. Returns a dict with x and y (tiny floating-point noise below 1e-12 is cleaned to 0).",
		"dict[str, str]", `math.polar_to_cartesian(2, 90, degrees=True)`, 87, "polar cartesian convert coordinates r theta x y",
		[]P{rf("r", "Radius."), rf("theta", "Angle."), ob("degrees", "theta is in degrees.", false)},
		func(a tools.Args) (any, error) {
			t := a.Float("theta")
			if a.Bool("degrees") {
				t = t * math.Pi / 180
			}
			clean := func(v float64) float64 {
				if math.Abs(v) < 1e-12*math.Max(1, math.Abs(a.Float("r"))) {
					return 0
				}
				return v
			}
			return dict("x", clean(a.Float("r")*math.Cos(t)), "y", clean(a.Float("r")*math.Sin(t))), nil
		})
	reg("cartesian_to_polar", "Cartesian to polar", "Converts Cartesian (x, y) to polar: returns r and theta (radians, or degrees with degrees=true, in the range -pi..pi / -180..180).",
		"dict[str, str]", `math.cartesian_to_polar(1, 1, degrees=True)`, 88, "cartesian polar convert coordinates r theta x y",
		[]P{rf("x", "X."), rf("y", "Y."), ob("degrees", "Return theta in degrees.", false)},
		func(a tools.Args) (any, error) {
			th := math.Atan2(a.Float("y"), a.Float("x"))
			if a.Bool("degrees") {
				th = th * 180 / math.Pi
			}
			return dict("r", math.Hypot(a.Float("x"), a.Float("y")), "theta", th), nil
		})
	reg("distance_2d", "Distance 2D", "Euclidean distance between (x1, y1) and (x2, y2).",
		"float", `math.distance_2d(0, 0, 3, 4)`, 89, "distance 2d euclidean points length between",
		[]P{rf("x1", "X1."), rf("y1", "Y1."), rf("x2", "X2."), rf("y2", "Y2.")},
		func(a tools.Args) (any, error) {
			return finite("distance_2d", math.Hypot(a.Float("x2")-a.Float("x1"), a.Float("y2")-a.Float("y1")))
		})
	reg("distance_3d", "Distance 3D", "Euclidean distance between (x1, y1, z1) and (x2, y2, z2).",
		"float", `math.distance_3d(0, 0, 0, 1, 2, 2)`, 90, "distance 3d euclidean points length between",
		[]P{rf("x1", "X1."), rf("y1", "Y1."), rf("z1", "Z1."), rf("x2", "X2."), rf("y2", "Y2."), rf("z2", "Z2.")},
		func(a tools.Args) (any, error) {
			return finite("distance_3d", math.Sqrt(math.Pow(a.Float("x2")-a.Float("x1"), 2)+math.Pow(a.Float("y2")-a.Float("y1"), 2)+math.Pow(a.Float("z2")-a.Float("z1"), 2)))
		})
	reg("haversine", "Great-circle distance", "Great-circle distance between two latitude/longitude points (degrees) using the haversine formula on a 6371.0088 km sphere. Returns km, mi, nmi and the initial bearing in degrees. Latitudes must be -90..90 and longitudes -180..180.",
		"dict[str, str]", `math.haversine(51.5074, -0.1278, 48.8566, 2.3522)`, 91, "haversine distance great circle latitude longitude gps km miles bearing geo",
		[]P{rf("lat1", "Latitude 1."), rf("lon1", "Longitude 1."), rf("lat2", "Latitude 2."), rf("lon2", "Longitude 2.")},
		func(a tools.Args) (any, error) {
			la1, lo1, la2, lo2 := a.Float("lat1"), a.Float("lon1"), a.Float("lat2"), a.Float("lon2")
			for _, v := range []float64{la1, la2} {
				if math.IsNaN(v) || v < -90 || v > 90 {
					return nil, errors.New("latitude must be between -90 and 90")
				}
			}
			for _, v := range []float64{lo1, lo2} {
				if math.IsNaN(v) || v < -180 || v > 180 {
					return nil, errors.New("longitude must be between -180 and 180")
				}
			}
			rad := math.Pi / 180
			p1, p2, dl := la1*rad, la2*rad, (lo2-lo1)*rad
			h := math.Pow(math.Sin((p2-p1)/2), 2) + math.Cos(p1)*math.Cos(p2)*math.Pow(math.Sin(dl/2), 2)
			km := 2 * 6371.0088 * math.Asin(math.Min(1, math.Sqrt(h)))
			y := math.Sin(dl) * math.Cos(p2)
			x := math.Cos(p1)*math.Sin(p2) - math.Sin(p1)*math.Cos(p2)*math.Cos(dl)
			bearing := math.Mod(math.Atan2(y, x)/rad+360, 360)
			return dict("km", km, "mi", km/1.609344, "nmi", km/1.852, "bearing", bearing), nil
		})
}

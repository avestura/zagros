package num

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

var ratRe = regexp.MustCompile(`^[+-]?(\d+(\.\d*)?|\.\d+)([eE][+-]?\d{1,4})?$`)

// parseRat parses a decimal number (optionally with exponent) or a fraction
// "a/b" exactly. Underscores, spaces and commas as thousands marks are ignored.
func parseRat(s string) (*big.Rat, error) {
	t := strings.TrimSpace(s)
	if len(t) > maxInputLen {
		return nil, errors.New("number is too long")
	}
	t = strings.NewReplacer("_", "", " ", "", ",", "", " ", "").Replace(t)
	if t == "" {
		return nil, errors.New("empty number")
	}
	if i := strings.IndexByte(t, '/'); i > 0 {
		n, err := parseRat(t[:i])
		if err != nil {
			return nil, err
		}
		d, err := parseRat(t[i+1:])
		if err != nil {
			return nil, err
		}
		if d.Sign() == 0 {
			return nil, errors.New("denominator is zero")
		}
		return n.Quo(n, d), nil
	}
	if !ratRe.MatchString(t) {
		return nil, fmt.Errorf("%q is not a number", strings.TrimSpace(s))
	}
	r, ok := new(big.Rat).SetString(t)
	if !ok {
		return nil, fmt.Errorf("%q is not a number", strings.TrimSpace(s))
	}
	return r, nil
}

// ratString renders r exactly if it is an integer, otherwise with up to
// maxDec decimals (trailing zeros trimmed).
func ratString(r *big.Rat, maxDec int) string {
	if r.IsInt() {
		return r.Num().String()
	}
	s := r.FloatString(maxDec)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	if s == "-0" || s == "" {
		return "0"
	}
	return s
}

func fracString(r *big.Rat) string {
	if r.IsInt() {
		return r.Num().String()
	}
	return r.Num().String() + "/" + r.Denom().String()
}

// limitDenominator returns the closest fraction to r with denominator <= max.
func limitDenominator(r *big.Rat, max *big.Int) *big.Rat {
	if r.Denom().Cmp(max) <= 0 {
		return new(big.Rat).Set(r)
	}
	neg := r.Sign() < 0
	n := new(big.Int).Abs(r.Num())
	d := new(big.Int).Set(r.Denom())
	p0, q0, p1, q1 := big.NewInt(0), big.NewInt(1), big.NewInt(1), big.NewInt(0)
	abs := new(big.Rat).Abs(r)
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
	diff := func(x *big.Rat) *big.Rat { return new(big.Rat).Abs(new(big.Rat).Sub(x, abs)) }
	res := b1
	if diff(b2).Cmp(diff(b1)) <= 0 {
		res = b2
	}
	if neg {
		res = new(big.Rat).Neg(res)
	}
	return res
}

func exactFloat(f float64) string {
	r := new(big.Rat)
	if r.SetFloat64(f) == nil {
		return strconv.FormatFloat(f, 'g', -1, 64)
	}
	s := r.FloatString(1080)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	if f == 0 && math.Signbit(f) {
		return "-0"
	}
	return s
}

func floatClass(exp, mant uint64, expMax uint64) string {
	switch {
	case exp == expMax && mant != 0:
		return "nan"
	case exp == expMax:
		return "inf"
	case exp == 0 && mant == 0:
		return "zero"
	case exp == 0:
		return "subnormal"
	}
	return "normal"
}

func init() {
	reg("ieee754_encode", "IEEE 754 encode", "Encodes a decimal number as an IEEE 754 binary32 (bits=32) or binary64 (bits=64) float and returns the hex bits, full binary, sign, biased exponent, mantissa, unbiased exponent, class and the exact stored decimal value. Accepts inf and nan.",
		"dict[str, str]", `num.ieee754_encode("0.1", 32)`, 6, "ieee754 float floating point binary32 binary64 single double hex bits encode representation mantissa exponent sign",
		[]P{rs("value", "Decimal number, e.g. 0.1, -2.5e10, inf, nan."), oi("bits", "32 (single) or 64 (double).", 32)},
		func(a tools.Args) (any, error) {
			bits := a.Int("bits")
			if bits != 32 && bits != 64 {
				return nil, errors.New("bits must be 32 or 64")
			}
			s := strings.TrimSpace(a.Str("value"))
			if len(s) > 400 {
				return nil, errors.New("value is too long")
			}
			f, err := strconv.ParseFloat(strings.ReplaceAll(s, "_", ""), int(bits))
			var ne *strconv.NumError
			if err != nil && !(errors.As(err, &ne) && ne.Err == strconv.ErrRange) {
				return nil, fmt.Errorf("%q is not a number", s)
			}
			var raw uint64
			var eb, mb int
			if bits == 32 {
				raw = uint64(math.Float32bits(float32(f)))
				eb, mb = 8, 23
			} else {
				raw = math.Float64bits(f)
				eb, mb = 11, 52
			}
			sign := raw >> uint(eb+mb)
			exp := (raw >> uint(mb)) & (1<<uint(eb) - 1)
			mant := raw & (1<<uint(mb) - 1)
			bin := fmt.Sprintf("%0*b", int(bits), raw)
			ue := int64(exp) - int64(1<<(uint(eb)-1)-1)
			if exp == 0 {
				ue = 1 - int64(1<<(uint(eb)-1)-1)
			}
			return map[string]string{
				"hex": fmt.Sprintf("%0*x", int(bits)/4, raw), "binary": bin,
				"grouped": bin[:1] + " " + bin[1:1+eb] + " " + bin[1+eb:],
				"sign":    fmt.Sprint(sign), "exponent": bin[1 : 1+eb], "mantissa": bin[1+eb:],
				"exponent_value": fmt.Sprint(ue), "class": floatClass(exp, mant, 1<<uint(eb)-1),
				"value": exactFloat64or32(raw, int(bits)),
			}, nil
		}, kw("ieee", "float", "ieee754"), accepts(func(s string) bool {
			_, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
			return err == nil && len(s) < 100
		}))

	reg("ieee754_decode", "IEEE 754 decode", "Decodes IEEE 754 hex bits (8 hex digits = binary32, 16 = binary64; 0x prefix and spaces ignored) into the decimal value, shortest round-trip form (3dcccccd gives 0.1). Returns NaN, +Inf or -Inf for the special patterns.",
		"str", `num.ieee754_decode("3dcccccd")`, 7, "ieee754 float floating point decode hex bits to number binary32 binary64",
		[]P{rs("hex", "Hex bits, e.g. 3dcccccd or 0x3FB999999999999A."), oi("bits", "32 or 64 (0 = from the digit count).", 0)},
		func(a tools.Args) (any, error) {
			h := strings.NewReplacer(" ", "", "_", "").Replace(strings.TrimSpace(a.Str("hex")))
			h = strings.TrimPrefix(strings.TrimPrefix(h, "0x"), "0X")
			bits := int(a.Int("bits"))
			if bits == 0 {
				switch len(h) {
				case 8:
					bits = 32
				case 16:
					bits = 64
				default:
					return nil, errors.New("expected 8 hex digits (binary32) or 16 (binary64)")
				}
			}
			if bits != 32 && bits != 64 {
				return nil, errors.New("bits must be 32 or 64")
			}
			if len(h) > bits/4 {
				return nil, fmt.Errorf("too many hex digits for %d bits", bits)
			}
			raw, err := strconv.ParseUint(h, 16, 64)
			if err != nil {
				return nil, fmt.Errorf("%q is not hex", h)
			}
			if bits == 32 {
				return strconv.FormatFloat(float64(math.Float32frombits(uint32(raw))), 'g', -1, 32), nil
			}
			return strconv.FormatFloat(math.Float64frombits(raw), 'g', -1, 64), nil
		}, kw("ieee", "float"), accepts(func(s string) bool {
			h := strings.TrimPrefix(strings.NewReplacer(" ", "", "_", "").Replace(strings.TrimSpace(s)), "0x")
			return (len(h) == 8 || len(h) == 16) && isHexDigits(h)
		}))

	reg("float_to_fraction", "Number to fraction", "Finds the closest fraction n/d to a decimal number with denominator at most max_denominator (continued fractions). A finite decimal like 0.125 gives exactly 1/8; 3.14159 with max 1000 gives 355/113. An integer gives just n.",
		"str", `num.float_to_fraction("3.14159", 1000)`, 8, "fraction rational decimal to fraction continued fraction approximate numerator denominator",
		[]P{rs("x", "Decimal number or a fraction a/b."), oi("max_denominator", "Largest allowed denominator (>= 1, up to 10^18).", 1000000)},
		func(a tools.Args) (any, error) {
			r, err := parseRat(a.Str("x"))
			if err != nil {
				return nil, err
			}
			m, err := intIn("max_denominator", a.Int("max_denominator"), 1, math.MaxInt64)
			if err != nil {
				return nil, err
			}
			return fracString(limitDenominator(r, big.NewInt(m))), nil
		}, accepts(func(s string) bool { _, err := parseRat(s); return err == nil && len(s) < 100 }))

	reg("nearest_float32", "Nearest float32", "Rounds a decimal number to the nearest IEEE 754 single-precision float and returns the exact decimal value of that float (0.1 gives 0.100000001490116119384765625). Values beyond float32 range are an error.",
		"str", `num.nearest_float32("0.1")`, 9, "float32 single precision round nearest representable error precision",
		[]P{rs("x", "Decimal number.")},
		func(a tools.Args) (any, error) {
			s := strings.TrimSpace(a.Str("x"))
			if len(s) > 400 {
				return nil, errors.New("value is too long")
			}
			f, err := strconv.ParseFloat(strings.ReplaceAll(s, "_", ""), 32)
			if err != nil {
				return nil, fmt.Errorf("%q is not a float32-range number", s)
			}
			return exactFloat(float64(float32(f))), nil
		}, accepts(func(s string) bool {
			_, err := strconv.ParseFloat(strings.TrimSpace(s), 32)
			return err == nil && len(s) < 100
		}))
}

func exactFloat64or32(raw uint64, bits int) string {
	if bits == 32 {
		return exactFloat(float64(math.Float32frombits(uint32(raw))))
	}
	return exactFloat(math.Float64frombits(raw))
}

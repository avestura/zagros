package mathx

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

type P = tools.Param
type fn = func(a tools.Args) (any, error)

func rs(n, d string) P { return P{Name: n, Type: "str", Doc: d, Required: true} }
func os(n, d, def string, choices ...string) P {
	return P{Name: n, Type: "str", Doc: d, Default: def, Choices: choices}
}
func rf(n, d string) P              { return P{Name: n, Type: "float", Doc: d, Required: true} }
func of(n, d string, def float64) P { return P{Name: n, Type: "float", Doc: d, Default: def} }
func ri(n, d string) P              { return P{Name: n, Type: "int", Doc: d, Required: true} }
func oi(n, d string, def int64) P   { return P{Name: n, Type: "int", Doc: d, Default: def} }
func ob(n, d string, def bool) P    { return P{Name: n, Type: "bool", Doc: d, Default: def} }
func rl(n, d string) P              { return P{Name: n, Type: "list[float]", Doc: d, Required: true} }

func values() P {
	return rl("values", "The numbers (a non-empty list of floats).")
}

// reg registers a math function. A function whose extra parameters are all
// optional also becomes a launcher variant of the "math" family.
func reg(name, title, doc, ret, example string, order int, tags string, params []P, f fn, opts ...func(*tools.Func)) {
	t := tools.Func{
		Group: "math", Name: name, Title: title, Doc: doc, Params: params,
		Returns: ret, Fn: f, Example: example, Tags: strings.Fields(tags), Order: order,
	}
	if launcher(params) {
		t.Family = "math"
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

// fmtNum formats x with at most 12 significant digits, no trailing zeros.
func fmtNum(x float64) string {
	switch {
	case math.IsNaN(x):
		return "NaN"
	case math.IsInf(x, 1):
		return "Infinity"
	case math.IsInf(x, -1):
		return "-Infinity"
	case x == 0:
		return "0"
	}
	r, _ := strconv.ParseFloat(strconv.FormatFloat(x, 'e', 11, 64), 64)
	if a := math.Abs(r); a >= 1e-7 && a < 1e21 {
		return strconv.FormatFloat(r, 'f', -1, 64)
	}
	return strconv.FormatFloat(r, 'e', -1, 64)
}

func money(x float64) string {
	s := strconv.FormatFloat(x, 'f', 2, 64)
	if s == "-0.00" {
		return "0.00"
	}
	return s
}

func finite(name string, x float64) (float64, error) {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return 0, fmt.Errorf("%s: result is not a finite number", name)
	}
	return x, nil
}

func nonEmpty(a tools.Args, name string) ([]float64, error) {
	v := a.Floats(name)
	if len(v) == 0 {
		return nil, fmt.Errorf("%s must not be empty", name)
	}
	if len(v) > 5_000_000 {
		return nil, fmt.Errorf("%s is too long", name)
	}
	for i, x := range v {
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return nil, fmt.Errorf("%s[%d] is not a finite number", name, i)
		}
	}
	return v, nil
}

func intIn(name string, v, lo, hi int64) (int64, error) {
	if v < lo || v > hi {
		return 0, fmt.Errorf("%s must be between %d and %d", name, lo, hi)
	}
	return v, nil
}

func dict(kv ...any) map[string]string {
	m := map[string]string{}
	for i := 0; i+1 < len(kv); i += 2 {
		k := kv[i].(string)
		switch v := kv[i+1].(type) {
		case float64:
			m[k] = fmtNum(v)
		case string:
			m[k] = v
		case int:
			m[k] = strconv.Itoa(v)
		case int64:
			m[k] = strconv.FormatInt(v, 10)
		}
	}
	return m
}

func moneyDict(kv ...any) map[string]string {
	m := map[string]string{}
	for i := 0; i+1 < len(kv); i += 2 {
		k := kv[i].(string)
		switch v := kv[i+1].(type) {
		case float64:
			m[k] = money(v)
		case string:
			m[k] = v
		case int:
			m[k] = strconv.Itoa(v)
		}
	}
	return m
}

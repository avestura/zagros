package builtin

import (
	"math"
	"testing"
)

func TestEvalExpr(t *testing.T) {
	cases := []struct {
		in   string
		want float64
	}{
		{"1+2*3", 7},
		{"(1+2)*3", 9},
		{"2^3^2", 512},
		{"-2^2", -4},
		{"2^-1", 0.5},
		{"10 % 3", 1},
		{"50%", 0.5},
		{"200 + 10%", 220},
		{"200 - 10%", 180},
		{"200 * 10%", 20},
		{"0xff + 1", 256},
		{"0b101 * 2", 10},
		{"1,000 + 1", 1001},
		{"1,000,000 / 4", 250000},
		{"sqrt(16) + abs(-2)", 6},
		{"max(1,200,3)", 200},
		{"min(4, 2)", 2},
		{"round(2.567, 2)", 2.57},
		{"floor(2.9)+ceil(2.1)", 5},
		{"pi*2", 2 * math.Pi},
		{"ln(e)", 1},
		{"log(1000)", 3},
		{"2e3", 2000},
		{"3 × 4 ÷ 2", 6},
		{"2**3", 8},
	}
	for _, c := range cases {
		got, err := evalExpr(c.in)
		if err != nil {
			t.Errorf("%q: %v", c.in, err)
			continue
		}
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("%q = %v, want %v", c.in, got, c.want)
		}
	}
}

func TestEvalErrors(t *testing.T) {
	for _, in := range []string{"1/0", "(1+2", "foo(2)", "sqrt()", "1 +", "7-zip"} {
		if _, err := evalExpr(in); err == nil {
			t.Errorf("%q: expected error", in)
		}
	}
}

func TestLooksLikeExpr(t *testing.T) {
	yes := []string{"1+2", "2*pi", "sqrt(9)", "10%3", "0xff+1"}
	no := []string{"42", "-5", "firefox", "7-zip", "chrome 2", "pi", "hello world"}
	for _, s := range yes {
		if !looksLikeExpr(s) {
			t.Errorf("%q should look like an expression", s)
		}
	}
	for _, s := range no {
		if looksLikeExpr(s) {
			t.Errorf("%q should not look like an expression", s)
		}
	}
}

func TestFormatNumber(t *testing.T) {
	cases := map[float64]string{0.1 + 0.2: "0.3", 1e20: "1e+20", 1234567: "1234567", -2.5: "-2.5", 1.0 / 3: "0.333333333333"}
	for in, want := range cases {
		if got := formatNumber(in); got != want {
			t.Errorf("formatNumber(%v) = %q, want %q", in, got, want)
		}
	}
	if g := groupThousands("-1234567.891"); g != "-1,234,567.891" {
		t.Errorf("groupThousands = %q", g)
	}
}

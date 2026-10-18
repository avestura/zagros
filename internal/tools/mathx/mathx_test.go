package mathx

import (
	"math"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "math") }

type m = map[string]any

func call(name string, args m) (any, error) { return tools.Call("math", name, args) }

func TestEval(t *testing.T) {
	tests := []struct {
		expr string
		want string
		deg  bool
	}{
		{"1+2*3", "7", false},
		{"(1+2)*3", "9", false},
		{"2^10", "1024", false},
		{"2**3**2", "512", false},
		{"-2^2", "-4", false},
		{"2^-1", "0.5", false},
		{"2(3+4)", "14", false},
		{"(1+2)(3+4)", "21", false},
		{"3pi", "9.42477796077", false},
		{"2e3 + 1", "2001", false},
		{"1,000 + 1,234.5", "2234.5", false},
		{"1_000*2", "2000", false},
		{"0.1+0.2", "0.3", false},
		{"1/3", "0.333333333333", false},
		{"7 % 3", "1", false},
		{"50%", "0.5", false},
		{"200 + 10%", "220", false},
		{"200 - 10%", "180", false},
		{"200 * 10%", "20", false},
		{"15% of 80", "12", false},
		{"100 + 10% + 10%", "121", false},
		{"sqrt(16)", "4", false},
		{"cbrt(27)", "3", false},
		{"sin(pi/2)", "1", false},
		{"sin(30)", "0.5", true},
		{"asin(1)", "90", true},
		{"atan2(1,1)", "0.785398163397", false},
		{"cos(pi)", "-1", false},
		{"ln(e)", "1", false},
		{"log(1000)", "3", false},
		{"log(8,2)", "3", false},
		{"log2(1024)", "10", false},
		{"log10(0.001)", "-3", false},
		{"exp(0)", "1", false},
		{"abs(-3)", "3", false},
		{"floor(2.7)+ceil(2.1)+trunc(-2.7)+round(2.5)", "6", false},
		{"round(2.675, 2)", "2.68", false},
		{"sign(-4)", "-1", false},
		{"min(3,1,2)+max(3,1,2)", "4", false},
		{"sum(1,2,3,4)", "10", false},
		{"avg(2,4,9)", "5", false},
		{"pow(2,8)", "256", false},
		{"hypot(3,4)", "5", false},
		{"gamma(5)", "24", false},
		{"factorial(5)", "120", false},
		{"5!", "120", false},
		{"deg(pi)", "180", false},
		{"rad(180)", "3.14159265359", false},
		{"tau", "6.28318530718", false},
		{"phi", "1.61803398875", false},
		{"e", "2.71828182846", false},
		{"2 × 3 ÷ 4", "1.5", false},
		{"2²", "4", false},
		{"max(1,234)", "234", false},
		{"1e15", "1000000000000000", false},
		{"1e25", "1e+25", false},
	}
	for _, tc := range tests {
		v, err := call("eval_str", m{"expr": tc.expr, "degrees": tc.deg})
		if err != nil {
			t.Errorf("eval_str(%q): %v", tc.expr, err)
			continue
		}
		if v.(string) != tc.want {
			t.Errorf("eval_str(%q) = %q, want %q", tc.expr, v, tc.want)
		}
	}
	v, _ := call("eval", m{"expr": "2*21"})
	if v.(float64) != 42 {
		t.Errorf("eval = %v", v)
	}
}

func TestEvalErrors(t *testing.T) {
	for _, e := range []string{"", "1/0", "1%0", "(1+2", "1+", "2 3", "foo", "foo(1)", "sqrt(-1)", "sqrt()", "sqrt(1,2)", "1 $ 2", "log(0)", "10^400", "(-8)^0.5", "ln(0)", "5.5!", "171!", "1e999", strings.Repeat("(", 1000), strings.Repeat("-", 1000) + "1", "max()", "1,2"} {
		if _, err := call("eval", m{"expr": e}); err == nil {
			t.Errorf("eval(%q): expected error", e)
		}
	}
	_, err := call("eval", m{"expr": "1 + $"})
	if err == nil || !strings.Contains(err.Error(), "position 5") {
		t.Errorf("error should name position 5, got %v", err)
	}
}

func approx(t *testing.T, name string, args m, want float64) {
	t.Helper()
	v, err := call(name, args)
	if err != nil {
		t.Errorf("%s(%v): %v", name, args, err)
		return
	}
	got := v.(float64)
	if math.Abs(got-want) > 1e-6*math.Max(1, math.Abs(want)) {
		t.Errorf("%s(%v) = %v, want %v", name, args, got, want)
	}
}

func TestFloats(t *testing.T) {
	approx(t, "percent_of", m{"p": 15.0, "x": 80.0}, 12)
	approx(t, "what_percent", m{"a": 20.0, "b": 50.0}, 40)
	approx(t, "percent_change", m{"old": 80.0, "new": 100.0}, 25)
	approx(t, "percent_change", m{"old": 100.0, "new": 80.0}, -20)
	approx(t, "add_percent", m{"x": 200.0, "p": 10.0}, 220)
	approx(t, "sub_percent", m{"x": 200.0, "p": 10.0}, 180)
	approx(t, "percent_diff", m{"a": 50.0, "b": 100.0}, 66.6666667)
	approx(t, "reverse_percent", m{"final": 120.0, "p": 20.0, "increase": true}, 100)
	approx(t, "reverse_percent", m{"final": 80.0, "p": 20.0, "increase": false}, 100)
	approx(t, "markup", m{"cost": 80.0, "price": 100.0}, 25)
	approx(t, "margin", m{"cost": 80.0, "price": 100.0}, 20)
	approx(t, "price_from_markup", m{"cost": 80.0, "markup_pct": 25.0}, 100)
	approx(t, "price_from_margin", m{"cost": 80.0, "margin_pct": 20.0}, 100)
	approx(t, "tax_add", m{"amount": 100.0, "rate": 20.0}, 120)
	approx(t, "tax_remove", m{"amount": 120.0, "rate": 20.0}, 100)
	approx(t, "sum", m{"values": []float64{0.1, 0.2, 0.3}}, 0.6)
	approx(t, "mean", m{"values": []float64{2, 4, 9}}, 5)
	approx(t, "median", m{"values": []float64{5, 1, 3, 2}}, 2.5)
	approx(t, "min", m{"values": []float64{4, 2, 9}}, 2)
	approx(t, "max", m{"values": []float64{4, 2, 9}}, 9)
	approx(t, "range", m{"values": []float64{4, 2, 9}}, 7)
	approx(t, "geometric_mean", m{"values": []float64{1, 4, 16}}, 4)
	approx(t, "harmonic_mean", m{"values": []float64{40, 60}}, 48)
	approx(t, "variance", m{"values": []float64{2, 4, 4, 4, 5, 5, 7, 9}, "sample": false}, 4)
	approx(t, "stdev", m{"values": []float64{2, 4, 4, 4, 5, 5, 7, 9}, "sample": false}, 2)
	approx(t, "stdev", m{"values": []float64{2, 4, 4, 4, 5, 5, 7, 9}, "sample": true}, 2.138089935)
	approx(t, "quantile", m{"values": []float64{1, 2, 3, 4}, "q": 0.25}, 1.75)
	approx(t, "percentile", m{"values": []float64{1, 2, 3, 4, 5}, "p": 90.0}, 4.6)
	approx(t, "iqr", m{"values": []float64{1, 2, 3, 4, 5, 6, 7, 8}}, 3.5)
	approx(t, "correlation", m{"xs": []float64{1, 2, 3}, "ys": []float64{2, 4, 6}}, 1)
	approx(t, "correlation", m{"xs": []float64{1, 2, 3}, "ys": []float64{3, 2, 1}}, -1)
	approx(t, "round_to", m{"x": 2.675, "decimals": int64(2)}, 2.68)
	approx(t, "round_to", m{"x": 2.5, "mode": "half_even"}, 2)
	approx(t, "round_to", m{"x": 3.5, "mode": "half_even"}, 4)
	approx(t, "round_to", m{"x": -2.5}, -3)
	approx(t, "round_to", m{"x": -2.5, "mode": "half_down"}, -2)
	approx(t, "round_to", m{"x": 2.1, "mode": "ceil"}, 3)
	approx(t, "round_to", m{"x": -2.1, "mode": "ceil"}, -2)
	approx(t, "round_to", m{"x": -2.9, "mode": "floor"}, -3)
	approx(t, "round_to", m{"x": -2.9, "mode": "trunc"}, -2)
	approx(t, "round_to", m{"x": 1234.0, "decimals": int64(-2)}, 1200)
	approx(t, "round_to_multiple", m{"x": 18.0, "multiple": 5.0}, 20)
	approx(t, "round_to_multiple", m{"x": 17.0, "multiple": 5.0}, 15)
	approx(t, "round_to_multiple", m{"x": 0.37, "multiple": 0.05}, 0.35)
	approx(t, "clamp", m{"x": 15.0, "lo": 0.0, "hi": 10.0}, 10)
	approx(t, "lerp", m{"a": 10.0, "b": 20.0, "t": 0.25}, 12.5)
	approx(t, "inverse_lerp", m{"a": 10.0, "b": 20.0, "x": 12.5}, 0.25)
	approx(t, "map_range", m{"x": 5.0, "in_lo": 0.0, "in_hi": 10.0, "out_lo": 0.0, "out_hi": 100.0}, 50)
	approx(t, "solve_linear", m{"a": 2.0, "b": -6.0}, 3)
	approx(t, "temperature", m{"value": 100.0, "from_unit": "c", "to_unit": "f"}, 212)
	approx(t, "temperature", m{"value": 0.0, "from_unit": "K", "to_unit": "°C"}, -273.15)
	approx(t, "temperature", m{"value": -40.0, "from_unit": "f", "to_unit": "c"}, -40)
	approx(t, "degrees_to_radians", m{"degrees": 180.0}, math.Pi)
	approx(t, "radians_to_degrees", m{"radians": math.Pi}, 180)
	approx(t, "distance_2d", m{"x1": 0.0, "y1": 0.0, "x2": 3.0, "y2": 4.0}, 5)
	approx(t, "distance_3d", m{"x1": 0.0, "y1": 0.0, "z1": 0.0, "x2": 1.0, "y2": 2.0, "z2": 2.0}, 3)
	approx(t, "circle_area", m{"r": 1.0}, math.Pi)
	approx(t, "circle_circumference", m{"r": 1.0}, 2*math.Pi)
	approx(t, "rect_area", m{"w": 3.0, "h": 4.0}, 12)
	approx(t, "rect_perimeter", m{"w": 3.0, "h": 4.0}, 14)
	approx(t, "triangle_area", m{"a": 3.0, "b": 4.0, "c": 5.0}, 6)
	approx(t, "triangle_perimeter", m{"a": 3.0, "b": 4.0, "c": 5.0}, 12)
	approx(t, "sphere_volume", m{"r": 1.0}, 4*math.Pi/3)
	approx(t, "sphere_area", m{"r": 1.0}, 4*math.Pi)
	approx(t, "cylinder_volume", m{"r": 1.0, "h": 2.0}, 2*math.Pi)
	approx(t, "cylinder_area", m{"r": 1.0, "h": 2.0}, 6*math.Pi)
	approx(t, "cone_volume", m{"r": 3.0, "h": 4.0}, 12*math.Pi)
	approx(t, "cone_area", m{"r": 3.0, "h": 4.0}, 24*math.Pi)
	approx(t, "pythagoras", m{"a": 3.0, "b": 4.0}, 5)
	approx(t, "pythagoras", m{"b": 12.0, "c": 13.0}, 5)
	approx(t, "pythagoras", m{"a": 5.0, "c": 13.0}, 12)
	approx(t, "future_value", m{"present": 1000.0, "rate": 5.0, "periods": 10.0}, 1628.894627)
	approx(t, "future_value", m{"present": 0.0, "rate": 0.0, "periods": 10.0, "payment": 100.0}, 1000)
	approx(t, "present_value", m{"future": 1628.894627, "rate": 5.0, "periods": 10.0}, 1000)
	approx(t, "cagr", m{"start": 1000.0, "end": 2000.0, "years": 5.0}, 14.869835)
	approx(t, "roi", m{"initial": 1000.0, "final": 1250.0}, 25)
	approx(t, "unit_price", m{"price": 3.5, "qty": 12.0}, 3.5/12)
	approx(t, "average_speed", m{"distance": 150.0, "time": 2.5}, 60)
	approx(t, "travel_time", m{"distance": 150.0, "speed": 60.0}, 2.5)
	approx(t, "travel_distance", m{"speed": 60.0, "time": 2.5}, 150)
	approx(t, "binomial_probability", m{"n": int64(10), "k": int64(5), "p": 0.5}, 0.24609375)
	approx(t, "binomial_probability", m{"n": int64(5), "k": int64(0), "p": 0.0}, 1)
	approx(t, "normal_cdf", m{"x": 0.0}, 0.5)
	approx(t, "normal_cdf", m{"x": 1.96}, 0.9750021)
	approx(t, "normal_cdf", m{"x": 110.0, "mean": 100.0, "sd": 10.0}, 0.8413447)
	approx(t, "normal_pdf", m{"x": 0.0}, 0.3989423)
	approx(t, "dice_probability", m{"sides": int64(6), "count": int64(2), "target": int64(7)}, 1.0/6)
	approx(t, "dice_probability", m{"sides": int64(6), "count": int64(1), "target": int64(7)}, 0)
	approx(t, "dms_to_decimal", m{"dms": "40°26'46\"N"}, 40.446111)
	approx(t, "dms_to_decimal", m{"dms": "73 59 11 W"}, -73.986389)
}

func TestStrings(t *testing.T) {
	tests := []struct {
		name string
		args m
		want string
	}{
		{"solve_quadratic", m{"a": 1.0, "b": -3.0, "c": 2.0}, "1\n2"},
		{"solve_quadratic", m{"a": 1.0, "b": 2.0, "c": 1.0}, "-1"},
		{"solve_quadratic", m{"a": 1.0, "b": 0.0, "c": 1.0}, "-i\ni"},
		{"solve_quadratic", m{"a": 1.0, "b": -2.0, "c": 5.0}, "1-2i\n1+2i"},
		{"solve_quadratic", m{"a": 2.0, "b": 0.0, "c": -8.0}, "-2\n2"},
		{"solve_quadratic", m{"a": 0.0, "b": 2.0, "c": -6.0}, "3"},
		{"simplify_fraction", m{"n": int64(6), "d": int64(-8)}, "-3/4"},
		{"simplify_fraction", m{"n": int64(10), "d": int64(5)}, "2"},
		{"fraction_add", m{"a": "1/2", "b": "1/3"}, "5/6"},
		{"fraction_sub", m{"a": "1/2", "b": "1/2"}, "0"},
		{"fraction_mul", m{"a": "2/3", "b": "3/4"}, "1/2"},
		{"fraction_div", m{"a": "1 1/2", "b": "0.5"}, "3"},
		{"fraction_calc", m{"expr": "3/4 + 1/6"}, "11/12  (11/12)"},
		{"fraction_calc", m{"expr": "(1/2)*(7/2)"}, "7/4  (1 3/4)"},
		{"decimal_to_fraction", m{"x": "0.125"}, "1/8"},
		{"decimal_to_fraction", m{"x": "3.14159", "max_denominator": int64(1000)}, "355/113"},
		{"mixed_number", m{"x": "7/2"}, "3 1/2"},
		{"mixed_number", m{"x": "-7/2"}, "-3 1/2"},
		{"mixed_number", m{"x": "2.75"}, "2 3/4"},
		{"ratio_simplify", m{"ratio": "1920:1080"}, "16:9"},
		{"ratio_simplify", m{"ratio": "4:6:8"}, "2:3:4"},
		{"ratio_simplify", m{"ratio": "0.5:1.5"}, "1:3"},
		{"scale_ratio", m{"ratio": "16:9", "known": 1920.0}, "1920:1080"},
		{"scale_ratio", m{"ratio": "16:9", "known": 1080.0, "position": int64(2)}, "1920:1080"},
		{"time_sum", m{"expr": "1:30 + 2:45"}, "4:15"},
		{"time_sum", m{"expr": "1h 30m + 45m"}, "2:15"},
		{"time_sum", m{"expr": "1:00 - 1:30"}, "-0:30"},
		{"time_sum", m{"expr": "0:45:30 + 0:14:30"}, "1:00"},
		{"time_sum", m{"expr": "0:00:20 + 0:00:50"}, "0:01:10"},
		{"angle_sum", m{"expr": "30 15 20 + 45 50"}, "76°05'20\""},
		{"decimal_to_dms", m{"degrees": 40.446}, "40°26'45.6\""},
		{"combinations", m{"n": int64(52), "k": int64(5)}, "2598960"},
		{"combinations", m{"n": int64(5), "k": int64(9)}, "0"},
		{"permutations", m{"n": int64(10), "k": int64(3)}, "720"},
		{"permutations", m{"n": int64(5), "k": int64(0)}, "1"},
		{"compound_interest", m{"principal": 1000.0, "rate": 5.0, "years": 10.0, "n": int64(1)}, "amount: 1628.89\neffective_annual_rate: 5\ninterest: 628.89"},
		{"loan_payment", m{"principal": 200000.0, "annual_rate": 6.0, "years": 30.0}, "interest: 231676.38\npayment: 1199.10\npayments: 360\ntotal: 431676.38"},
		{"loan_payment", m{"principal": 1200.0, "annual_rate": 0.0, "years": 1.0}, "interest: 0.00\npayment: 100.00\npayments: 12\ntotal: 1200.00"},
		{"simple_interest", m{"principal": 1000.0, "rate": 5.0, "years": 3.0}, "interest: 150.00\ntotal: 1150.00"},
		{"discount", m{"price": 80.0, "pct": 25.0}, "effective_percent: 25\nfinal: 60\nsaved: 20"},
		{"vat", m{"amount": 120.0, "rate": 20.0, "inclusive": true}, "gross: 120\nnet: 100\nvat: 20"},
		{"tip", m{"bill": 85.5, "pct": 18.0, "people": int64(2)}, "bill_per_person: 42.75\nper_person: 50.45\ntip: 15.39\ntip_per_person: 7.70\ntotal: 100.89"},
		{"bmi", m{"kg": 70.0, "m": 1.75}, "bmi: 22.9\ncategory: normal"},
		{"markup_margin", m{"cost": 80.0, "price": 100.0}, "cost: 80\nmargin: 20\nmarkup: 25\nprice: 100\nprofit: 20"},
		{"rule_of_72", m{"rate": 6.0}, "estimate: 12\nexact: 11.8956610459"},
		{"break_even", m{"fixed_costs": 5000.0, "price": 25.0, "variable_cost": 15.0}, "contribution_margin: 10\nrevenue: 12500\nunits: 500\nunits_rounded_up: 500"},
		{"compare_unit_price", m{"a_price": 3.0, "a_qty": 10.0, "b_price": 5.0, "b_qty": 20.0}, "cheaper: b\nsaving_percent: 16.6666666667\nunit_a: 0.3\nunit_b: 0.25"},
		{"mode", m{"values": []float64{1, 2, 2, 3, 3}}, "2\n3"},
		{"normalize", m{"values": []float64{10, 20, 30}}, "0\n0.5\n1"},
		{"cumulative_sum", m{"values": []float64{1, 2, 3, 4}}, "1\n3\n6\n10"},
		{"moving_average", m{"values": []float64{1, 2, 3, 4, 5}, "window": int64(3)}, "2\n3\n4"},
		{"outliers", m{"values": []float64{1, 2, 2, 3, 3, 4, 100}}, "100"},
		{"zscores", m{"values": []float64{1, 2, 3}}, "-1.224744871392\n0\n1.224744871392"},
		{"solve_linear_system_2x2", m{"matrix": []float64{1, 1, 3, 1, -1, 1}}, "2\n1"},
		{"solve_linear_system_3x3", m{"matrix": []float64{1, 1, 1, 6, 0, 2, 5, -4, 2, 5, -1, 27}}, "5\n3\n-2"},
		{"count", m{"values": []float64{1, 2, 3}}, "3"},
		{"percent_text", m{"text": "15% of 80"}, "meaning: 15% of 80\nresult: 12"},
		{"percent_text", m{"text": "20 is what % of 50"}, "meaning: 20 is this percent of 50\nresult: 40"},
		{"percent_text", m{"text": "from 80 to 100"}, "difference: 20\nresult: 25"},
		{"percent_text", m{"text": "50 is 25% of what"}, "meaning: the whole when 50 is 25%\nresult: 200"},
		{"percent_text", m{"text": "200 + 10%"}, "change: 20\nresult: 220"},
		{"temperature_text", m{"text": "72f"}, "c: 22.2222222222\nf: 72\nk: 295.372222222\nr: 531.67"},
		{"quadratic_text", m{"text": "x^2 - 3x + 2 = 0"}, "a: 1\nb: -3\nc: 2\ndiscriminant: 1\nx1: 1\nx2: 2"},
		{"quadratic_text", m{"text": "2x² + 8x"}, "a: 2\nb: 8\nc: 0\ndiscriminant: 64\nx1: -4\nx2: 0"},
		{"quadratic_text", m{"text": "1 -3 2"}, "a: 1\nb: -3\nc: 2\ndiscriminant: 1\nx1: 1\nx2: 2"},
	}
	for _, tc := range tests {
		v, err := call(tc.name, tc.args)
		if err != nil {
			t.Errorf("%s(%v): %v", tc.name, tc.args, err)
			continue
		}
		got := tools.Format(v)
		if got != tc.want {
			t.Errorf("%s(%v) =\n%q\nwant\n%q", tc.name, tc.args, got, tc.want)
		}
	}
}

func TestDescribeAndRegression(t *testing.T) {
	v, err := call("describe", m{"values": []float64{2, 4, 4, 4, 5, 5, 7, 9}})
	if err != nil {
		t.Fatal(err)
	}
	d := v.(map[string]string)
	for k, w := range map[string]string{"count": "8", "sum": "40", "mean": "5", "median": "4.5", "mode": "4", "min": "2", "max": "9", "range": "7", "pop_stdev": "2", "pop_variance": "4"} {
		if d[k] != w {
			t.Errorf("describe[%s] = %q, want %q", k, d[k], w)
		}
	}
	v, _ = call("stats_text", m{"text": "scores: 3, 5, 8 and 1,000 -2"})
	if s := v.(map[string]string); s["count"] != "5" || s["sum"] != "1014" || s["min"] != "-2" {
		t.Errorf("stats_text = %v", s)
	}
	v, err = call("linear_regression", m{"xs": []float64{1, 2, 3, 4}, "ys": []float64{3, 5, 7, 9}})
	if err != nil {
		t.Fatal(err)
	}
	r := v.(map[string]string)
	if r["slope"] != "2" || r["intercept"] != "1" || r["r2"] != "1" {
		t.Errorf("regression = %v", r)
	}
	v, _ = call("haversine", m{"lat1": 51.5074, "lon1": -0.1278, "lat2": 48.8566, "lon2": 2.3522})
	h := v.(map[string]string)
	if !strings.HasPrefix(h["km"], "343.") && !strings.HasPrefix(h["km"], "344.") {
		t.Errorf("haversine London-Paris = %v", h)
	}
	v, _ = call("polar_to_cartesian", m{"r": 2.0, "theta": 90.0, "degrees": true})
	if p := v.(map[string]string); p["x"] != "0" || p["y"] != "2" {
		t.Errorf("polar_to_cartesian = %v", p)
	}
	v, _ = call("cartesian_to_polar", m{"x": 1.0, "y": 1.0, "degrees": true})
	if p := v.(map[string]string); p["theta"] != "45" || p["r"] != "1.41421356237" {
		t.Errorf("cartesian_to_polar = %v", p)
	}
	v, _ = call("geometry", m{"shape": "triangle", "a": 3.0, "b": 4.0, "c": 5.0})
	if g := v.(map[string]string); g["area"] != "6" || g["perimeter"] != "12" {
		t.Errorf("geometry triangle = %v", g)
	}
	v, _ = call("geometry", m{"shape": "cylinder", "a": 1.0, "b": 2.0})
	if g := v.(map[string]string); g["volume"] != "6.28318530718" {
		t.Errorf("geometry cylinder = %v", g)
	}
	v, _ = call("amortization_summary", m{"principal": 200000.0, "annual_rate": 6.0, "years": 30.0, "after": int64(360)})
	if a := v.(map[string]string); a["balance"] != "0.00" || a["payments_left"] != "0" {
		t.Errorf("amortization end = %v", a)
	}
	v, _ = call("amortization_summary", m{"principal": 200000.0, "annual_rate": 6.0, "years": 30.0, "after": int64(1)})
	if a := v.(map[string]string); a["first_interest"] != "1000.00" || a["first_principal"] != "199.10" {
		t.Errorf("amortization first = %v", a)
	}
	v, _ = call("loan_text", m{"text": "200000 6% 30"})
	if l := v.(map[string]string); l["payment"] != "1199.10" {
		t.Errorf("loan_text = %v", l)
	}
	v, _ = call("tip_text", m{"text": "85.50 18% 2"})
	if l := v.(map[string]string); l["per_person"] != "50.45" {
		t.Errorf("tip_text = %v", l)
	}
	if s := strings.Join(tools.Args{}.List("x"), ""); s != "" {
		t.Fail()
	}
}

func TestErrors(t *testing.T) {
	bad := []struct {
		name string
		args m
	}{
		{"what_percent", m{"a": 1.0, "b": 0.0}},
		{"percent_change", m{"old": 0.0, "new": 5.0}},
		{"price_from_margin", m{"cost": 1.0, "margin_pct": 100.0}},
		{"markup", m{"cost": 0.0, "price": 5.0}},
		{"sum", m{"values": []float64{}}},
		{"mean", m{"values": []float64{math.NaN()}}},
		{"variance", m{"values": []float64{1}, "sample": true}},
		{"geometric_mean", m{"values": []float64{1, -1}}},
		{"harmonic_mean", m{"values": []float64{0}}},
		{"quantile", m{"values": []float64{1}, "q": 2.0}},
		{"percentile", m{"values": []float64{1}, "p": -1.0}},
		{"zscores", m{"values": []float64{2, 2}}},
		{"moving_average", m{"values": []float64{1, 2}, "window": int64(5)}},
		{"correlation", m{"xs": []float64{1, 2}, "ys": []float64{1}}},
		{"correlation", m{"xs": []float64{1, 1}, "ys": []float64{1, 2}}},
		{"round_to", m{"x": 1.0, "mode": "weird"}},
		{"round_to", m{"x": math.Inf(1)}},
		{"round_to", m{"x": 1.0, "decimals": int64(9999)}},
		{"round_to_multiple", m{"x": 1.0, "multiple": 0.0}},
		{"clamp", m{"x": 1.0, "lo": 5.0, "hi": 1.0}},
		{"map_range", m{"x": 1.0, "in_lo": 2.0, "in_hi": 2.0, "out_lo": 0.0, "out_hi": 1.0}},
		{"solve_linear", m{"a": 0.0, "b": 1.0}},
		{"solve_quadratic", m{"a": 0.0, "b": 0.0, "c": 1.0}},
		{"solve_quadratic", m{"a": 1e200, "b": 1e200, "c": 1.0}},
		{"solve_linear_system_2x2", m{"matrix": []float64{1, 2, 3}}},
		{"solve_linear_system_2x2", m{"matrix": []float64{1, 1, 1, 2, 2, 2}}},
		{"simplify_fraction", m{"n": int64(1), "d": int64(0)}},
		{"fraction_add", m{"a": "x", "b": "1/2"}},
		{"fraction_add", m{"a": "1/0", "b": "1/2"}},
		{"fraction_div", m{"a": "1/2", "b": "0"}},
		{"fraction_calc", m{"expr": "1/(2-2)"}},
		{"fraction_calc", m{"expr": "1 +"}},
		{"decimal_to_fraction", m{"x": "abc"}},
		{"ratio_simplify", m{"ratio": "16"}},
		{"ratio_simplify", m{"ratio": "a:b"}},
		{"ratio_simplify", m{"ratio": "0:0"}},
		{"scale_ratio", m{"ratio": "16:9", "known": 1.0, "position": int64(5)}},
		{"temperature", m{"value": 1.0, "from_unit": "x", "to_unit": "c"}},
		{"temperature", m{"value": -500.0, "from_unit": "c", "to_unit": "f"}},
		{"temperature_text", m{"text": "hot"}},
		{"haversine", m{"lat1": 100.0, "lon1": 0.0, "lat2": 0.0, "lon2": 0.0}},
		{"geometry", m{"shape": "blob", "a": 1.0}},
		{"geometry", m{"shape": "circle", "a": -1.0}},
		{"geometry", m{"shape": "triangle", "a": 1.0, "b": 1.0, "c": 5.0}},
		{"triangle_area", m{"a": 1.0, "b": 2.0, "c": 10.0}},
		{"circle_area", m{"r": -1.0}},
		{"pythagoras", m{"a": 3.0, "b": 4.0, "c": 5.0}},
		{"pythagoras", m{"a": 0.0, "b": 0.0, "c": 5.0}},
		{"pythagoras", m{"b": 12.0, "c": 5.0}},
		{"compound_interest", m{"principal": 1.0, "rate": 5.0, "years": 1.0, "n": int64(-1)}},
		{"compound_interest", m{"principal": 1e300, "rate": 500.0, "years": 100.0, "n": int64(1)}},
		{"loan_payment", m{"principal": -1.0, "annual_rate": 5.0, "years": 1.0}},
		{"loan_payment", m{"principal": 1.0, "annual_rate": 5.0, "years": 1000.0}},
		{"loan_payment", m{"principal": 1.0, "annual_rate": 5.0, "years": 1.0, "per_year": int64(0)}},
		{"amortization_summary", m{"principal": 1000.0, "annual_rate": 5.0, "years": 1.0, "after": int64(99)}},
		{"cagr", m{"start": 0.0, "end": 1.0, "years": 1.0}},
		{"roi", m{"initial": 0.0, "final": 1.0}},
		{"rule_of_72", m{"rate": 0.0}},
		{"break_even", m{"fixed_costs": 1.0, "price": 1.0, "variable_cost": 2.0}},
		{"unit_price", m{"price": 1.0, "qty": 0.0}},
		{"compare_unit_price", m{"a_price": 1.0, "a_qty": 0.0, "b_price": 1.0, "b_qty": 1.0}},
		{"average_speed", m{"distance": 1.0, "time": 0.0}},
		{"bmi", m{"kg": 0.0, "m": 1.0}},
		{"time_sum", m{"expr": ""}},
		{"time_sum", m{"expr": "abc + 1:00"}},
		{"time_sum", m{"expr": "90 + 1:00"}},
		{"time_sum", m{"expr": "1:00 +"}},
		{"angle_sum", m{"expr": "x + y"}},
		{"dms_to_decimal", m{"dms": "hello"}},
		{"decimal_to_dms", m{"degrees": math.NaN()}},
		{"combinations", m{"n": int64(-1), "k": int64(1)}},
		{"combinations", m{"n": int64(999999), "k": int64(1)}},
		{"permutations", m{"n": int64(5), "k": int64(-1)}},
		{"binomial_probability", m{"n": int64(5), "k": int64(1), "p": 2.0}},
		{"normal_cdf", m{"x": 0.0, "sd": 0.0}},
		{"dice_probability", m{"sides": int64(1), "count": int64(1), "target": int64(1)}},
		{"dice_probability", m{"sides": int64(100), "count": int64(100), "target": int64(1)}},
		{"stats_text", m{"text": "no digits here"}},
		{"percent_text", m{"text": "gibberish"}},
		{"quadratic_text", m{"text": "1 2"}},
		{"quadratic_text", m{"text": "x^2 = 5"}},
		{"tip_text", m{"text": ""}},
		{"loan_text", m{"text": "1 2"}},
		{"tip", m{"bill": 10.0, "people": int64(0)}},
	}
	for _, tc := range bad {
		if _, err := call(tc.name, tc.args); err == nil {
			t.Errorf("%s(%v): expected an error", tc.name, tc.args)
		}
	}
}

func TestGarbage(t *testing.T) {
	inputs := []string{"", " ", "-", "((", "))", "1e999999", "NaN", "inf", "\x00", "ÿ☃", "1/0", "1 2 3", ":", "%", "!!", "0x", "1,,2", "..", "1..2", strings.Repeat("9", 400), strings.Repeat("1+", 3000)}
	for _, f := range tools.All() {
		if f.Group != "math" {
			continue
		}
		for _, in := range inputs {
			args := m{}
			for _, p := range f.Params {
				if p.Required {
					switch p.Type {
					case "str":
						args[p.Name] = in
					case "int":
						args[p.Name] = int64(3)
					case "float":
						args[p.Name] = 1.5
					case "list[float]":
						args[p.Name] = []float64{1, 2, 3}
					}
				}
			}
			if len(f.Params) > 0 && f.Params[0].Type == "str" {
				args[f.Params[0].Name] = in
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s(%q) panicked: %v", f.Key(), in, r)
					}
				}()
				if f.Accepts != nil {
					f.Accepts(in)
				}
				v, err := f.Invoke(args)
				if err == nil {
					switch v.(type) {
					case string, int64, float64, bool, []string, []float64, map[string]string, nil:
					default:
						t.Errorf("%s returned disallowed type %T", f.Key(), v)
					}
				}
			}()
		}
	}
}

func TestExtractNumbers(t *testing.T) {
	got := extractNumbers("a 1-2, -3.5 and 1,234 and 2e3 .5 x4")
	want := []float64{1, 2, -3.5, 1234, 2000, 0.5, 4}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got %v, want %v", got, want)
			break
		}
	}
}

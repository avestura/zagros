package builtin

import (
	"math"
	"testing"
)

func TestConvertUnits(t *testing.T) {
	cases := []struct {
		v        float64
		from, to string
		want     float64
	}{
		{10, "km", "mi", 6.21371192},
		{1, "mi", "km", 1.609344},
		{100, "C", "F", 212},
		{32, "f", "c", 0},
		{0, "c", "k", 273.15},
		{1, "GiB", "MiB", 1024},
		{1, "GB", "MB", 1000},
		{2, "lb", "kg", 0.90718474},
		{90, "min", "h", 1.5},
		{1, "gal", "l", 3.785411784},
		{36, "kmh", "mps", 10},
	}
	for _, c := range cases {
		got, err := convertUnits(c.v, c.from, c.to)
		if err != nil {
			t.Errorf("%v %s->%s: %v", c.v, c.from, c.to, err)
			continue
		}
		if math.Abs(got-c.want) > 1e-6*math.Max(1, math.Abs(c.want)) {
			t.Errorf("%v %s->%s = %v, want %v", c.v, c.from, c.to, got, c.want)
		}
	}
	if _, err := convertUnits(1, "kg", "km"); err == nil {
		t.Error("dimension mismatch must fail")
	}
	if _, err := convertUnits(1, "zz", "km"); err == nil {
		t.Error("unknown unit must fail")
	}
}

func TestParseUnitQuery(t *testing.T) {
	v, f, to, ok := parseUnitQuery("10 km to mi")
	if !ok || v != 10 || f != "km" || to != "mi" {
		t.Errorf("got %v %q %q %v", v, f, to, ok)
	}
	v, f, to, ok = parseUnitQuery("1,500g in kg")
	if !ok || v != 1500 || f != "g" || to != "kg" {
		t.Errorf("got %v %q %q %v", v, f, to, ok)
	}
	if _, _, _, ok := parseUnitQuery("hello world"); ok {
		t.Error("should not parse")
	}
}

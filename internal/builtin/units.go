package builtin

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type unitDef struct {
	dim    string
	factor float64 // multiply by factor to reach the dimension's base unit
}

// Base units: length m, mass g, data B, time s, volume l. Temperature is special.
var unitTable = map[string]unitDef{}

func init() {
	add := func(dim string, factor float64, names ...string) {
		for _, n := range names {
			unitTable[n] = unitDef{dim, factor}
		}
	}
	add("length", 1, "m", "meter", "meters", "metre", "metres")
	add("length", 1e-3, "mm", "millimeter", "millimeters")
	add("length", 1e-2, "cm", "centimeter", "centimeters")
	add("length", 1e3, "km", "kilometer", "kilometers", "kilometre", "kilometres")
	add("length", 1e-6, "um", "µm", "micrometer")
	add("length", 1e-9, "nm", "nanometer")
	add("length", 0.0254, "in", "inch", "inches", "\"")
	add("length", 0.3048, "ft", "foot", "feet")
	add("length", 0.9144, "yd", "yard", "yards")
	add("length", 1609.344, "mi", "mile", "miles")
	add("length", 1852, "nmi", "nauticalmile", "nauticalmiles")

	add("mass", 1, "g", "gram", "grams")
	add("mass", 1e-3, "mg", "milligram", "milligrams")
	add("mass", 1e3, "kg", "kilogram", "kilograms", "kilo", "kilos")
	add("mass", 1e6, "t", "tonne", "tonnes")
	add("mass", 28.349523125, "oz", "ounce", "ounces")
	add("mass", 453.59237, "lb", "lbs", "pound", "pounds")
	add("mass", 6350.29318, "st", "stone")

	add("data", 1, "b", "byte", "bytes")
	add("data", 1e3, "kb")
	add("data", 1e6, "mb")
	add("data", 1e9, "gb")
	add("data", 1e12, "tb")
	add("data", 1e15, "pb")
	add("data", 1024, "kib")
	add("data", 1024*1024, "mib")
	add("data", 1024*1024*1024, "gib")
	add("data", 1024*1024*1024*1024, "tib")
	add("data", 0.125, "bit", "bits")

	add("time", 1, "s", "sec", "secs", "second", "seconds")
	add("time", 1e-3, "ms", "millisecond", "milliseconds")
	add("time", 60, "min", "mins", "minute", "minutes")
	add("time", 3600, "h", "hr", "hrs", "hour", "hours")
	add("time", 86400, "d", "day", "days")
	add("time", 604800, "wk", "week", "weeks")

	add("volume", 1, "l", "liter", "liters", "litre", "litres")
	add("volume", 1e-3, "ml", "milliliter", "milliliters")
	add("volume", 3.785411784, "gal", "gallon", "gallons")
	add("volume", 0.946352946, "qt", "quart", "quarts")
	add("volume", 0.473176473, "pt", "pint", "pints")
	add("volume", 0.2365882365, "cup", "cups")
	add("volume", 0.0295735295625, "floz")

	add("speed", 1, "mps", "m/s")
	add("speed", 1000.0/3600, "kmh", "km/h", "kph")
	add("speed", 0.44704, "mph", "mi/h")
	add("speed", 0.514444, "kn", "knot", "knots")
}

var tempUnits = map[string]string{
	"c": "c", "°c": "c", "celsius": "c", "degc": "c",
	"f": "f", "°f": "f", "fahrenheit": "f", "degf": "f",
	"k": "k", "kelvin": "k",
}

func toKelvin(v float64, u string) float64 {
	switch u {
	case "c":
		return v + 273.15
	case "f":
		return (v-32)*5/9 + 273.15
	}
	return v
}

func fromKelvin(k float64, u string) float64 {
	switch u {
	case "c":
		return k - 273.15
	case "f":
		return (k-273.15)*9/5 + 32
	}
	return k
}

// convertUnits converts v between two named units. Unit names are case
// insensitive except that "mb"/"Mb" style data units are treated as bytes.
func convertUnits(v float64, from, to string) (float64, error) {
	f, t := strings.ToLower(from), strings.ToLower(to)
	if fu, ok := tempUnits[f]; ok {
		tu, ok2 := tempUnits[t]
		if !ok2 {
			return 0, fmt.Errorf("cannot convert %s to %s", from, to)
		}
		return fromKelvin(toKelvin(v, fu), tu), nil
	}
	fd, ok1 := unitTable[f]
	td, ok2 := unitTable[t]
	if !ok1 {
		return 0, fmt.Errorf("unknown unit %q", from)
	}
	if !ok2 {
		return 0, fmt.Errorf("unknown unit %q", to)
	}
	if fd.dim != td.dim {
		return 0, fmt.Errorf("cannot convert %s (%s) to %s (%s)", from, fd.dim, to, td.dim)
	}
	return v * fd.factor / td.factor, nil
}

var unitQueryRe = regexp.MustCompile(`(?i)^\s*(-?[\d.,]+)\s*([^\s\d][^\s]*?)\s+(?:to|in|into|->|=>|=)\s+(\S+)\s*$`)

// parseUnitQuery parses "10 km to mi".
func parseUnitQuery(s string) (v float64, from, to string, ok bool) {
	m := unitQueryRe.FindStringSubmatch(s)
	if m == nil {
		return 0, "", "", false
	}
	n, err := strconv.ParseFloat(strings.ReplaceAll(m[1], ",", ""), 64)
	if err != nil {
		return 0, "", "", false
	}
	return n, m[2], m[3], true
}

// isKnownUnit reports whether name is a unit this converter understands.
func isKnownUnit(name string) bool {
	n := strings.ToLower(name)
	if _, ok := tempUnits[n]; ok {
		return true
	}
	_, ok := unitTable[n]
	return ok
}

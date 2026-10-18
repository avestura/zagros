package mathx

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

var numRe = regexp.MustCompile(`[-+]?(?:\d{1,3}(?:,\d{3})+|\d+)(?:\.\d+)?(?:[eE][-+]?\d+)?|[-+]?\.\d+`)

// extractNumbers pulls every number out of free text. Thousands groups like
// 1,234 are read as one number; a sign only counts when it is not glued to a
// preceding letter, digit or dot (so "1-2" reads as 1 and 2).
func extractNumbers(text string) []float64 {
	var out []float64
	for _, loc := range numRe.FindAllStringIndex(text, 100000) {
		s := text[loc[0]:loc[1]]
		if s[0] == '-' || s[0] == '+' {
			if loc[0] > 0 {
				c := text[loc[0]-1]
				if c >= '0' && c <= '9' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c == '.' || c == ')' {
					s = s[1:]
				}
			}
		}
		v, err := strconv.ParseFloat(strings.ReplaceAll(strings.TrimPrefix(s, "+"), ",", ""), 64)
		if err == nil && !math.IsInf(v, 0) {
			out = append(out, v)
		}
	}
	return out
}

func sortedCopy(v []float64) []float64 {
	s := append([]float64(nil), v...)
	sort.Float64s(s)
	return s
}

func meanOf(v []float64) float64 { return sumOf(v) / float64(len(v)) }

func varianceOf(v []float64, sample bool) (float64, error) {
	n := float64(len(v))
	if sample && len(v) < 2 {
		return 0, errors.New("sample variance needs at least 2 values")
	}
	m := meanOf(v)
	ss := make([]float64, len(v))
	for i, x := range v {
		ss[i] = (x - m) * (x - m)
	}
	d := n
	if sample {
		d = n - 1
	}
	return sumOf(ss) / d, nil
}

// quantileOf uses linear interpolation between closest ranks (NumPy default).
func quantileOf(sorted []float64, q float64) float64 {
	if len(sorted) == 1 {
		return sorted[0]
	}
	pos := q * float64(len(sorted)-1)
	lo := math.Floor(pos)
	hi := math.Ceil(pos)
	return sorted[int(lo)] + (sorted[int(hi)]-sorted[int(lo)])*(pos-lo)
}

func modeOf(v []float64) []float64 {
	counts := map[float64]int{}
	best := 0
	for _, x := range v {
		counts[x]++
		if counts[x] > best {
			best = counts[x]
		}
	}
	var out []float64
	for x, c := range counts {
		if c == best {
			out = append(out, x)
		}
	}
	sort.Float64s(out)
	return out
}

func describeOf(v []float64) map[string]string {
	s := sortedCopy(v)
	sampleVar, err := varianceOf(v, true)
	if err != nil {
		sampleVar = 0
	}
	popVar, _ := varianceOf(v, false)
	modes := modeOf(v)
	ms := make([]string, len(modes))
	for i, x := range modes {
		ms[i] = fmtNum(x)
	}
	q1, q3 := quantileOf(s, 0.25), quantileOf(s, 0.75)
	m := dict("count", len(v), "sum", sumOf(v), "mean", meanOf(v), "median", quantileOf(s, 0.5),
		"min", s[0], "max", s[len(s)-1], "range", s[len(s)-1]-s[0],
		"variance", sampleVar, "stdev", math.Sqrt(sampleVar), "pop_variance", popVar, "pop_stdev", math.Sqrt(popVar),
		"q1", q1, "q3", q3, "iqr", q3-q1)
	m["mode"] = strings.Join(ms, ", ")
	pos := true
	for _, x := range v {
		if x <= 0 {
			pos = false
		}
	}
	if pos {
		m["geometric_mean"] = fmtNum(geoMean(v))
		m["harmonic_mean"] = fmtNum(harmMean(v))
	}
	return m
}

func geoMean(v []float64) float64 {
	l := make([]float64, len(v))
	for i, x := range v {
		l[i] = math.Log(x)
	}
	return math.Exp(meanOf(l))
}

func harmMean(v []float64) float64 {
	r := make([]float64, len(v))
	for i, x := range v {
		r[i] = 1 / x
	}
	return float64(len(v)) / sumOf(r)
}

func pairs(a tools.Args) (xs, ys []float64, err error) {
	xs, err = nonEmpty(a, "xs")
	if err != nil {
		return
	}
	ys, err = nonEmpty(a, "ys")
	if err != nil {
		return
	}
	if len(xs) != len(ys) {
		return nil, nil, fmt.Errorf("xs has %d values but ys has %d", len(xs), len(ys))
	}
	if len(xs) < 2 {
		return nil, nil, errors.New("need at least 2 points")
	}
	return
}

func regress(xs, ys []float64) (slope, intercept, r, r2 float64, err error) {
	mx, my := meanOf(xs), meanOf(ys)
	var sxy, sxx, syy float64
	for i := range xs {
		dx, dy := xs[i]-mx, ys[i]-my
		sxy += dx * dy
		sxx += dx * dx
		syy += dy * dy
	}
	if sxx == 0 {
		return 0, 0, 0, 0, errors.New("xs are all equal; no line can be fitted")
	}
	slope = sxy / sxx
	intercept = my - slope*mx
	if syy == 0 {
		return slope, intercept, 0, 1, nil
	}
	r = sxy / math.Sqrt(sxx*syy)
	return slope, intercept, r, r * r, nil
}

func init() {
	one := func(name, title, doc, ret, ex, tags string, order int, f func(v []float64) (any, error)) {
		reg(name, title, doc, ret, ex, order, tags, []P{values()}, func(a tools.Args) (any, error) {
			v, err := nonEmpty(a, "values")
			if err != nil {
				return nil, err
			}
			return f(v)
		})
	}
	one("sum", "Sum", "Adds up a list of numbers with compensated summation (sum([0.1, 0.2, 0.3]) = 0.6). Empty list is an error.", "float", `math.sum([1, 2, 3.5])`, "sum total add numbers list", 40,
		func(v []float64) (any, error) { return sumOf(v), nil })
	one("mean", "Mean", "Arithmetic mean (average) of a list of numbers.", "float", `math.mean([2, 4, 9])`, "mean average avg arithmetic mean statistics", 41,
		func(v []float64) (any, error) { return meanOf(v), nil })
	one("median", "Median", "Median of a list: the middle value, or the mean of the two middle values for an even count.", "float", `math.median([5, 1, 3, 2])`, "median middle value statistics 50th percentile", 42,
		func(v []float64) (any, error) { return quantileOf(sortedCopy(v), 0.5), nil })
	one("mode", "Mode", "Returns the most frequent value(s) in ascending order; several values are returned when they tie (all values if all are unique).", "list[float]", `math.mode([1, 2, 2, 3, 3])`, "mode most frequent common value statistics multimode", 43,
		func(v []float64) (any, error) { return modeOf(v), nil })
	one("min", "Minimum", "Smallest value in a list.", "float", `math.min([4, 2, 9])`, "min minimum smallest lowest", 44,
		func(v []float64) (any, error) { return sortedCopy(v)[0], nil })
	one("max", "Maximum", "Largest value in a list.", "float", `math.max([4, 2, 9])`, "max maximum largest highest", 45,
		func(v []float64) (any, error) { s := sortedCopy(v); return s[len(s)-1], nil })
	one("range", "Range", "Spread of a list: max minus min.", "float", `math.range([4, 2, 9])`, "range spread max min difference", 46,
		func(v []float64) (any, error) { s := sortedCopy(v); return s[len(s)-1] - s[0], nil })
	one("count", "Count", "Number of values in a list.", "int", `math.count([4, 2, 9])`, "count length number of values size", 47,
		func(v []float64) (any, error) { return int64(len(v)), nil })
	one("geometric_mean", "Geometric mean", "Geometric mean of positive numbers (the n-th root of their product, computed via logs). Zero or negative values are an error.", "float", `math.geometric_mean([1, 4, 16])`, "geometric mean growth average statistics", 48,
		func(v []float64) (any, error) {
			for _, x := range v {
				if x <= 0 {
					return nil, errors.New("geometric mean needs positive values")
				}
			}
			return geoMean(v), nil
		})
	one("harmonic_mean", "Harmonic mean", "Harmonic mean of positive numbers (count divided by the sum of reciprocals), suited to averaging rates. Zero or negative values are an error.", "float", `math.harmonic_mean([40, 60])`, "harmonic mean rate average statistics", 49,
		func(v []float64) (any, error) {
			for _, x := range v {
				if x <= 0 {
					return nil, errors.New("harmonic mean needs positive values")
				}
			}
			return harmMean(v), nil
		})
	one("iqr", "Interquartile range", "Interquartile range: the 75th minus the 25th percentile (linear interpolation).", "float", `math.iqr([1, 2, 3, 4, 5, 6, 7, 8])`, "iqr interquartile range quartile spread statistics", 50,
		func(v []float64) (any, error) {
			s := sortedCopy(v)
			return quantileOf(s, 0.75) - quantileOf(s, 0.25), nil
		})
	one("normalize", "Normalize 0..1", "Rescales values linearly to the range 0..1 (min-max normalisation). If all values are equal every result is 0.", "list[float]", `math.normalize([10, 20, 30])`, "normalize min max scale 0 1 rescale feature scaling", 51,
		func(v []float64) (any, error) {
			s := sortedCopy(v)
			lo, hi := s[0], s[len(s)-1]
			out := make([]float64, len(v))
			if hi > lo {
				for i, x := range v {
					out[i] = (x - lo) / (hi - lo)
				}
			}
			return out, nil
		})
	one("cumulative_sum", "Cumulative sum", "Running total: element i is the sum of the first i+1 values.", "list[float]", `math.cumulative_sum([1, 2, 3, 4])`, "cumulative sum running total cumsum prefix", 52,
		func(v []float64) (any, error) {
			out := make([]float64, len(v))
			var s float64
			for i, x := range v {
				s += x
				out[i] = s
			}
			return out, nil
		})
	one("describe", "Describe data", "Summary statistics as a dict: count, sum, mean, median, mode, min, max, range, sample variance and stdev, pop_variance, pop_stdev, q1, q3, iqr and (for positive data) geometric_mean and harmonic_mean.", "dict[str, str]", `math.describe([2, 4, 4, 4, 5, 5, 7, 9])`, "describe summary statistics stats mean median stdev variance quartiles", 3,
		func(v []float64) (any, error) { return describeOf(v), nil })

	reg("variance", "Variance", "Variance of a list: sample variance (n-1 divisor, default) or population variance with sample=false. Sample variance needs at least 2 values.", "float", `math.variance([2, 4, 4, 4, 5, 5, 7, 9], sample=False)`, 53, "variance spread statistics sample population",
		[]P{values(), ob("sample", "Use the sample (n-1) formula; false gives population variance.", true)},
		func(a tools.Args) (any, error) {
			v, err := nonEmpty(a, "values")
			if err != nil {
				return nil, err
			}
			return varianceOf(v, a.Bool("sample"))
		})
	reg("stdev", "Standard deviation", "Standard deviation: sample (n-1 divisor, default) or population with sample=false. For [2,4,4,4,5,5,7,9] the population value is 2.", "float", `math.stdev([2, 4, 4, 4, 5, 5, 7, 9], sample=False)`, 54, "standard deviation stdev sigma spread statistics sample population",
		[]P{values(), ob("sample", "Use the sample (n-1) formula; false gives population stdev.", true)},
		func(a tools.Args) (any, error) {
			v, err := nonEmpty(a, "values")
			if err != nil {
				return nil, err
			}
			x, err := varianceOf(v, a.Bool("sample"))
			return math.Sqrt(x), err
		})
	reg("zscores", "Z-scores", "Standardises values: (x - mean) / stdev for each value. Uses the population stdev unless sample=true. All-equal data is an error (stdev is 0).", "list[float]", `math.zscores([1, 2, 3, 4, 5])`, 55, "zscore z score standardize standard score statistics",
		[]P{values(), ob("sample", "Use the sample stdev.", false)},
		func(a tools.Args) (any, error) {
			v, err := nonEmpty(a, "values")
			if err != nil {
				return nil, err
			}
			vr, err := varianceOf(v, a.Bool("sample"))
			if err != nil {
				return nil, err
			}
			sd := math.Sqrt(vr)
			if sd == 0 {
				return nil, errors.New("standard deviation is 0")
			}
			m := meanOf(v)
			out := make([]float64, len(v))
			for i, x := range v {
				out[i] = (x - m) / sd
			}
			return out, nil
		})
	reg("quantile", "Quantile", "Quantile q (0..1) with linear interpolation between ranks (NumPy default): quantile(v, 0.5) is the median.", "float", `math.quantile([1, 2, 3, 4], 0.25)`, 56, "quantile quartile percentile interpolation statistics",
		[]P{values(), rf("q", "Fraction between 0 and 1.")},
		func(a tools.Args) (any, error) {
			v, err := nonEmpty(a, "values")
			if err != nil {
				return nil, err
			}
			q := a.Float("q")
			if q < 0 || q > 1 || math.IsNaN(q) {
				return nil, errors.New("q must be between 0 and 1")
			}
			return quantileOf(sortedCopy(v), q), nil
		})
	reg("percentile", "Percentile", "Percentile p (0..100) with linear interpolation: percentile(v, 90) is the value below which 90 percent of the data fall.", "float", `math.percentile([1, 2, 3, 4, 5], 90)`, 57, "percentile rank statistics interpolation",
		[]P{values(), rf("p", "Percent between 0 and 100.")},
		func(a tools.Args) (any, error) {
			v, err := nonEmpty(a, "values")
			if err != nil {
				return nil, err
			}
			p := a.Float("p")
			if p < 0 || p > 100 || math.IsNaN(p) {
				return nil, errors.New("p must be between 0 and 100")
			}
			return quantileOf(sortedCopy(v), p/100), nil
		})
	reg("moving_average", "Moving average", "Simple moving average with the given window; returns len(values)-window+1 averages (only full windows). window must be 1..len(values).", "list[float]", `math.moving_average([1, 2, 3, 4, 5], 3)`, 58, "moving average rolling mean smoothing sma window",
		[]P{values(), oi("window", "Window size.", 3)},
		func(a tools.Args) (any, error) {
			v, err := nonEmpty(a, "values")
			if err != nil {
				return nil, err
			}
			w, err := intIn("window", a.Int("window"), 1, int64(len(v)))
			if err != nil {
				return nil, err
			}
			out := make([]float64, 0, len(v)-int(w)+1)
			for i := 0; i+int(w) <= len(v); i++ {
				out = append(out, meanOf(v[i:i+int(w)]))
			}
			return out, nil
		})
	reg("outliers", "Outliers", "Returns the values outside the Tukey fences Q1 - k*IQR and Q3 + k*IQR (k defaults to 1.5), in input order.", "list[float]", `math.outliers([1, 2, 2, 3, 3, 4, 100])`, 59, "outliers anomalies tukey iqr fences extreme values",
		[]P{values(), of("k", "Fence multiplier.", 1.5)},
		func(a tools.Args) (any, error) {
			v, err := nonEmpty(a, "values")
			if err != nil {
				return nil, err
			}
			s := sortedCopy(v)
			q1, q3 := quantileOf(s, 0.25), quantileOf(s, 0.75)
			lo, hi := q1-a.Float("k")*(q3-q1), q3+a.Float("k")*(q3-q1)
			out := []float64{}
			for _, x := range v {
				if x < lo || x > hi {
					out = append(out, x)
				}
			}
			return out, nil
		})
	reg("correlation", "Correlation", "Pearson correlation coefficient of two equal-length lists (-1..1). Needs at least 2 points and non-constant data.", "float", `math.correlation([1, 2, 3], [2, 4, 6])`, 60, "correlation pearson r statistics relationship",
		[]P{rl("xs", "X values."), rl("ys", "Y values.")},
		func(a tools.Args) (any, error) {
			xs, ys, err := pairs(a)
			if err != nil {
				return nil, err
			}
			_, _, r, _, err := regress(xs, ys)
			if err != nil {
				return nil, err
			}
			return r, nil
		})
	reg("linear_regression", "Linear regression", "Least-squares line y = slope*x + intercept through the points; returns slope, intercept, r and r2 as a dict. xs must not all be equal.", "dict[str, str]", `math.linear_regression([1, 2, 3], [2, 4, 6])`, 61, "linear regression least squares trend line fit slope intercept r2",
		[]P{rl("xs", "X values."), rl("ys", "Y values.")},
		func(a tools.Args) (any, error) {
			xs, ys, err := pairs(a)
			if err != nil {
				return nil, err
			}
			s, i, r, r2, err := regress(xs, ys)
			if err != nil {
				return nil, err
			}
			return dict("slope", s, "intercept", i, "r", r, "r2", r2), nil
		})

	reg("stats_text", "Statistics of text", "Extracts every number from any text (commas, newlines or words in between; 1,234 is one number) and returns summary statistics like describe. Errors if the text has no numbers.",
		"dict[str, str]", `math.stats_text("scores: 3, 5, 8 and 13")`, 2, "stats statistics text numbers extract average sum mean median describe paste list",
		[]P{rs("text", "Text containing numbers.")},
		func(a tools.Args) (any, error) {
			v := extractNumbers(a.Str("text"))
			if len(v) == 0 {
				return nil, errors.New("no numbers found")
			}
			return describeOf(v), nil
		}, accepts(func(s string) bool { return len(extractNumbers(s)) >= 2 }), kw("stats", "avg", "sum", "mean"))
}

package randx

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

func init() {
	reg(tools.Func{
		Name: "int", Title: "Random integer", Order: 1, Input: "args",
		Doc:     "Returns a uniformly random integer between min and max, both inclusive. With one number in args (launcher) the range is 1..n; errors when min is greater than max.",
		Params:  []tools.Param{p("min", "int", "Lowest possible value.", int64(0)), p("max", "int", "Highest possible value.", int64(100)), seedParam(), argsParam("'6' means 1..6, '10 20' means 10..20.")},
		Returns: "int", Example: `rand.int(1, 6)`,
		Tags: []string{"random number", "integer", "between", "range", "rng", "randint"},
		Fn: func(a tools.Args) (any, error) {
			lo, hi := a.Int("min"), a.Int("max")
			if n := numArgs(a.Str("args")); len(n) == 1 {
				lo, hi = 1, n[0]
			} else if len(n) >= 2 {
				lo, hi = n[0], n[1]
			}
			if lo > hi {
				return nil, fmt.Errorf("min (%d) is greater than max (%d)", lo, hi)
			}
			return intRange(newRng(a.Int("seed")), lo, hi), nil
		},
	})
	reg(tools.Func{
		Name: "ints", Title: "Random integers",
		Doc:     "Returns count random integers in [min, max]; with unique=true no value repeats (errors when the range is too small). Values above 2^53 lose precision in the float list.",
		Params:  []tools.Param{p("count", "int", "How many numbers (0..10000).", int64(10)), p("min", "int", "Lowest value.", int64(0)), p("max", "int", "Highest value.", int64(100)), p("unique", "bool", "Forbid repeated values.", false), seedParam()},
		Returns: "list[float]", Example: `rand.ints(5, 1, 49, unique=True)`,
		Tags: []string{"random numbers", "list", "unique", "distinct", "lottery"},
		Fn: func(a tools.Args) (any, error) {
			if err := checkCount(a.Int("count"), "count"); err != nil {
				return nil, err
			}
			lo, hi := a.Int("min"), a.Int("max")
			if lo > hi {
				return nil, fmt.Errorf("min (%d) is greater than max (%d)", lo, hi)
			}
			r := newRng(a.Int("seed"))
			if a.Bool("unique") {
				v, err := distinctInts(r, a.Int("count"), lo, hi)
				return toFloats(v), err
			}
			out := make([]float64, a.Int("count"))
			for i := range out {
				out[i] = float64(intRange(r, lo, hi))
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "float", Title: "Random decimal", Order: 2, Input: "args",
		Doc:     "Returns a random float in [min, max) rounded to the given number of decimals (0..15; negative keeps full precision).",
		Params:  []tools.Param{p("min", "float", "Lower bound.", float64(0)), p("max", "float", "Upper bound.", float64(1)), p("decimals", "int", "Decimal places to round to; -1 means no rounding.", int64(-1)), seedParam(), argsParam("'1 10' means a float between 1 and 10.")},
		Returns: "float", Example: `rand.float(0, 10, 2)`,
		Tags: []string{"random decimal", "float", "double", "uniform"},
		Fn: func(a tools.Args) (any, error) {
			lo, hi := a.Float("min"), a.Float("max")
			if f := strings.Fields(a.Str("args")); len(f) > 0 {
				x, e1 := strconv.ParseFloat(f[0], 64)
				if len(f) == 1 && e1 == nil {
					lo, hi = 0, x
				} else if y, e2 := strconv.ParseFloat(f[1%len(f)], 64); e1 == nil && e2 == nil {
					lo, hi = x, y
				}
			}
			if math.IsNaN(lo) || math.IsNaN(hi) || math.IsInf(lo, 0) || math.IsInf(hi, 0) {
				return nil, fmt.Errorf("bounds must be finite")
			}
			if lo > hi {
				return nil, fmt.Errorf("min (%v) is greater than max (%v)", lo, hi)
			}
			v := lo + newRng(a.Int("seed")).Float64()*(hi-lo)
			if d := a.Int("decimals"); d >= 0 {
				if d > 15 {
					d = 15
				}
				s := math.Pow(10, float64(d))
				v = math.Round(v*s) / s
			}
			return v, nil
		},
	})
	reg(tools.Func{
		Name: "bool", Title: "Random boolean",
		Doc:     "Returns true with the given probability (default 0.5) and false otherwise.",
		Params:  []tools.Param{p("probability", "float", "Chance of true between 0 and 1.", 0.5), seedParam()},
		Returns: "bool", Example: `rand.bool(0.25)`,
		Tags: []string{"random boolean", "true false", "bernoulli", "chance"},
		Fn: func(a tools.Args) (any, error) {
			pr := a.Float("probability")
			if pr < 0 || pr > 1 {
				return nil, fmt.Errorf("probability must be between 0 and 1")
			}
			return newRng(a.Int("seed")).Float64() < pr, nil
		},
	})
	reg(tools.Func{
		Name: "coin_flip", Title: "Flip a coin", Order: 3, Input: "args", Keywords: []string{"coin", "flip", "heads", "tails"},
		Doc:     "Flips one fair coin and returns \"heads\" or \"tails\".",
		Params:  []tools.Param{seedParam(), argsParam("ignored.")},
		Returns: "str", Example: `rand.coin_flip()`,
		Tags: []string{"coin toss", "heads", "tails", "flip"},
		Fn: func(a tools.Args) (any, error) {
			if newRng(a.Int("seed")).IntN(2) == 0 {
				return "heads", nil
			}
			return "tails", nil
		},
	})
	reg(tools.Func{
		Name: "coin_flips", Title: "Flip many coins",
		Doc:     "Flips n fair coins (0..10000) and returns the counts of heads and tails plus the H/T sequence (sequence only when n <= 200).",
		Params:  []tools.Param{p("n", "int", "Number of flips.", int64(10)), seedParam()},
		Returns: "dict[str, str]", Example: `rand.coin_flips(10)`,
		Tags: []string{"coin toss", "heads", "tails", "streak", "binomial"},
		Fn: func(a tools.Args) (any, error) {
			n := a.Int("n")
			if err := checkCount(n, "n"); err != nil {
				return nil, err
			}
			r := newRng(a.Int("seed"))
			var seq strings.Builder
			h := 0
			for i := int64(0); i < n; i++ {
				if r.IntN(2) == 0 {
					h++
					seq.WriteByte('H')
				} else {
					seq.WriteByte('T')
				}
			}
			m := map[string]string{"heads": strconv.Itoa(h), "tails": strconv.FormatInt(n-int64(h), 10), "flips": strconv.FormatInt(n, 10)}
			if n <= 200 {
				m["sequence"] = seq.String()
			}
			return m, nil
		},
	})
	reg(tools.Func{
		Name: "choice", Title: "Random choice", Order: 4,
		Doc:     "Returns one random element of items; errors on an empty list.",
		Params:  []tools.Param{req("items", "list[str]", "Candidates."), seedParam()},
		Returns: "str", Example: `rand.choice(["red", "green", "blue"])`,
		Tags: []string{"pick one", "random element", "select", "choose"},
		Fn: func(a tools.Args) (any, error) {
			it := a.List("items")
			if len(it) == 0 {
				return nil, errEmpty
			}
			return it[newRng(a.Int("seed")).IntN(len(it))], nil
		},
	})
	pick := func(items []string, n int64, unique bool, seed int64) (any, error) {
		if err := checkCount(n, "n"); err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return nil, errEmpty
		}
		r := newRng(seed)
		if unique {
			if int(n) > len(items) {
				return nil, fmt.Errorf("cannot pick %d unique items from %d", n, len(items))
			}
			return shuffled(r, items)[:n], nil
		}
		out := make([]string, n)
		for i := range out {
			out[i] = items[r.IntN(len(items))]
		}
		return out, nil
	}
	reg(tools.Func{
		Name: "pick", Title: "Pick n random items",
		Doc:     "Picks n elements from items. With unique=true (default) no element is picked twice by position and n may not exceed the list length; with unique=false picks may repeat.",
		Params:  []tools.Param{req("items", "list[str]", "Candidates."), p("n", "int", "How many to pick.", int64(1)), p("unique", "bool", "Do not pick the same position twice.", true), seedParam()},
		Returns: "list[str]", Example: `rand.pick(["a", "b", "c", "d"], 2)`,
		Tags: []string{"pick", "select", "winners", "draw", "raffle"},
		Fn: func(a tools.Args) (any, error) {
			return pick(a.List("items"), a.Int("n"), a.Bool("unique"), a.Int("seed"))
		},
	})
	reg(tools.Func{
		Name: "sample", Title: "Random sample",
		Doc:     "Returns a random sample of k distinct positions of items (sampling without replacement), in random order.",
		Params:  []tools.Param{req("items", "list[str]", "Population."), p("k", "int", "Sample size.", int64(1)), seedParam()},
		Returns: "list[str]", Example: `rand.sample(["a", "b", "c"], 2)`,
		Tags: []string{"sample", "without replacement", "subset"},
		Fn:   func(a tools.Args) (any, error) { return pick(a.List("items"), a.Int("k"), true, a.Int("seed")) },
	})
	reg(tools.Func{
		Name: "shuffle", Title: "Shuffle list", Order: 5,
		Doc:     "Returns the items in random order (Fisher-Yates); the input list is not modified.",
		Params:  []tools.Param{req("items", "list[str]", "Items to shuffle."), seedParam()},
		Returns: "list[str]", Example: `rand.shuffle(["a", "b", "c"])`,
		Tags: []string{"randomize", "permute", "reorder", "mix"},
		Fn: func(a tools.Args) (any, error) {
			if len(a.List("items")) > 1000000 {
				return nil, fmt.Errorf("too many items")
			}
			return shuffled(newRng(a.Int("seed")), a.List("items")), nil
		},
	})
	reg(tools.Func{
		Name: "shuffle_text", Title: "Shuffle text", Order: 6, Input: "text", Keywords: []string{"scramble", "shuffle"},
		Doc:     "Randomly reorders the characters, words or lines of text according to by (chars, words, lines). Words are re-joined with single spaces.",
		Params:  []tools.Param{req("text", "str", "Text to shuffle."), p("by", "str", "Unit to shuffle.", "chars"), seedParam()},
		Returns: "str", Example: `rand.shuffle_text("one two three", "words")`,
		Tags: []string{"scramble", "anagram", "randomize text", "mix letters"},
		Fn: func(a tools.Args) (any, error) {
			t := a.Str("text")
			if utf8.RuneCountInString(t) > maxLength {
				return nil, fmt.Errorf("text is longer than %d characters", maxLength)
			}
			r := newRng(a.Int("seed"))
			switch strings.ToLower(a.Str("by")) {
			case "chars", "char", "characters", "":
				rs := []rune(t)
				r.Shuffle(len(rs), func(i, j int) { rs[i], rs[j] = rs[j], rs[i] })
				return string(rs), nil
			case "words", "word":
				return strings.Join(shuffled(r, strings.Fields(t)), " "), nil
			case "lines", "line":
				return strings.Join(shuffled(r, strings.Split(strings.ReplaceAll(t, "\r\n", "\n"), "\n")), "\n"), nil
			}
			return nil, fmt.Errorf("by must be chars, words or lines")
		},
	})
	reg(tools.Func{
		Name: "weighted_choice", Title: "Weighted random choice",
		Doc:     "Picks one item with probability proportional to its weight. Weights must match items in length, be non-negative and not all zero.",
		Params:  []tools.Param{req("items", "list[str]", "Candidates."), req("weights", "list[float]", "One non-negative weight per item."), seedParam()},
		Returns: "str", Example: `rand.weighted_choice(["common", "rare"], [9, 1])`,
		Tags: []string{"weighted", "probability", "loot table", "biased"},
		Fn: func(a tools.Args) (any, error) {
			it, w := a.List("items"), a.Floats("weights")
			if len(it) == 0 {
				return nil, errEmpty
			}
			if len(it) != len(w) {
				return nil, fmt.Errorf("items (%d) and weights (%d) differ in length", len(it), len(w))
			}
			sum := 0.0
			for _, x := range w {
				if x < 0 || math.IsNaN(x) || math.IsInf(x, 0) {
					return nil, fmt.Errorf("weights must be finite and non-negative")
				}
				sum += x
			}
			if sum <= 0 {
				return nil, fmt.Errorf("weights must not all be zero")
			}
			t := newRng(a.Int("seed")).Float64() * sum
			for i, x := range w {
				if t < x {
					return it[i], nil
				}
				t -= x
			}
			for i := len(w) - 1; i >= 0; i-- {
				if w[i] > 0 {
					return it[i], nil
				}
			}
			return it[len(it)-1], nil
		},
	})
	reg(tools.Func{
		Name: "random_item_from_text", Title: "Random item from text", Order: 7, Input: "text", Keywords: []string{"choose", "choice", "draw"},
		Doc:     "Splits text into items and returns one at random. With sep empty it splits by lines, else commas/semicolons, else whitespace; blank items are dropped.",
		Params:  []tools.Param{req("text", "str", "Items separated by newlines, commas or spaces."), p("sep", "str", "Explicit separator; empty auto-detects.", ""), seedParam()},
		Returns: "str", Example: `rand.random_item_from_text("pizza, sushi, tacos")`,
		Tags: []string{"choose from list", "decide", "random line", "pick one"},
		Fn: func(a tools.Args) (any, error) {
			it := splitItems(a.Str("text"), a.Str("sep"))
			if len(it) == 0 {
				return nil, errEmpty
			}
			return it[newRng(a.Int("seed")).IntN(len(it))], nil
		},
	})
	reg(tools.Func{
		Name: "spin_wheel", Title: "Spin the wheel", Order: 8, Input: "text", Keywords: []string{"wheel", "spin", "decide"},
		Doc:     "Spins a wheel of fortune over the items in text (lines, commas or words) and returns the winner. Errors when there are no items.",
		Params:  []tools.Param{req("text", "str", "Items on the wheel."), seedParam()},
		Returns: "str", Example: `rand.spin_wheel("alice, bob, carol")`,
		Tags: []string{"wheel of fortune", "raffle", "winner", "decide", "who goes first"},
		Fn: func(a tools.Args) (any, error) {
			it := splitItems(a.Str("text"), "")
			if len(it) == 0 {
				return nil, errEmpty
			}
			return it[newRng(a.Int("seed")).IntN(len(it))], nil
		},
	})
	reg(tools.Func{
		Name: "yes_no_maybe", Title: "Yes, no or maybe", Order: 9, Input: "args", Keywords: []string{"yesno", "decide"},
		Doc:     "Answers a question with a random \"Yes\", \"No\" or \"Maybe\" (equal odds). With maybe=false only Yes or No.",
		Params:  []tools.Param{p("maybe", "bool", "Include Maybe as an answer.", true), seedParam(), argsParam("the question (ignored).")},
		Returns: "str", Example: `rand.yes_no_maybe()`,
		Tags: []string{"decision", "yes no", "oracle"},
		Fn: func(a tools.Args) (any, error) {
			o := []string{"Yes", "No", "Maybe"}
			if !a.Bool("maybe") {
				o = o[:2]
			}
			return o[newRng(a.Int("seed")).IntN(len(o))], nil
		},
	})
	reg(tools.Func{
		Name: "magic_8_ball", Title: "Magic 8-ball", Order: 10, Input: "question", Keywords: []string{"8ball", "eightball", "oracle"},
		Doc:     "Returns one of the 20 classic Magic 8-Ball answers at random; the question is only for show.",
		Params:  []tools.Param{p("question", "str", "Your question (not used).", ""), seedParam()},
		Returns: "str", Example: `rand.magic_8_ball("Will it ship?")`,
		Tags: []string{"8 ball", "fortune", "oracle", "answer"},
		Fn: func(a tools.Args) (any, error) {
			return ball8[newRng(a.Int("seed")).IntN(len(ball8))], nil
		},
	})
}

var ball8 = []string{
	"It is certain.", "It is decidedly so.", "Without a doubt.", "Yes - definitely.", "You may rely on it.",
	"As I see it, yes.", "Most likely.", "Outlook good.", "Yes.", "Signs point to yes.",
	"Reply hazy, try again.", "Ask again later.", "Better not tell you now.", "Cannot predict now.", "Concentrate and ask again.",
	"Don't count on it.", "My reply is no.", "My sources say no.", "Outlook not so good.", "Very doubtful.",
}

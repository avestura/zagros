package mathx

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

var numPat = `[-+]?\$?[\d][\d,]*\.?\d*|[-+]?\.\d+`

func parseNum(s string) (float64, error) {
	t := strings.NewReplacer(",", "", "$", "", "€", "", "£", "", " ", "").Replace(strings.TrimSpace(s))
	v, err := strconv.ParseFloat(t, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) {
		return 0, fmt.Errorf("%q is not a number", strings.TrimSpace(s))
	}
	return v, nil
}

var (
	rePctOf    = regexp.MustCompile(`(?i)^(?:what\s+is\s+)?(` + numPat + `)\s*(?:%|percent|pct)\s+of\s+(` + numPat + `)\s*\??$`)
	reIsWhat   = regexp.MustCompile(`(?i)^(` + numPat + `)\s+is\s+what\s+(?:%|percent|pct)\s+of\s+(` + numPat + `)\s*\??$`)
	reWhatIs   = regexp.MustCompile(`(?i)^what\s+(?:%|percent|pct)\s+is\s+(` + numPat + `)\s+of\s+(` + numPat + `)\s*\??$`)
	reChange   = regexp.MustCompile(`(?i)^(?:from\s+)?(` + numPat + `)\s*(?:to|->|→|=>)\s*(` + numPat + `)$`)
	reReverse  = regexp.MustCompile(`(?i)^(` + numPat + `)\s+is\s+(` + numPat + `)\s*(?:%|percent|pct)\s+of\s+what\s*\??$`)
	reIncrease = regexp.MustCompile(`(?i)^(` + numPat + `)\s*(\+|-|plus|minus)\s*(` + numPat + `)\s*(?:%|percent|pct)$`)
)

func percentText(s string) (map[string]string, error) {
	t := strings.TrimSpace(s)
	pair := func(m []string) (float64, float64, error) {
		a, err := parseNum(m[1])
		if err != nil {
			return 0, 0, err
		}
		b, err := parseNum(m[2])
		return a, b, err
	}
	if m := rePctOf.FindStringSubmatch(t); m != nil {
		p, x, err := pair(m)
		if err != nil {
			return nil, err
		}
		return dict("result", p/100*x, "meaning", fmtNum(p)+"% of "+fmtNum(x)), nil
	}
	if m := reIsWhat.FindStringSubmatch(t); m != nil {
		a, b, err := pair(m)
		if err != nil {
			return nil, err
		}
		if b == 0 {
			return nil, errors.New("cannot take a percentage of zero")
		}
		return dict("result", a/b*100, "meaning", fmtNum(a)+" is this percent of "+fmtNum(b)), nil
	}
	if m := reWhatIs.FindStringSubmatch(t); m != nil {
		a, b, err := pair(m)
		if err != nil {
			return nil, err
		}
		if b == 0 {
			return nil, errors.New("cannot take a percentage of zero")
		}
		return dict("result", a/b*100, "meaning", fmtNum(a)+" is this percent of "+fmtNum(b)), nil
	}
	if m := reReverse.FindStringSubmatch(t); m != nil {
		a, p, err := pair(m)
		if err != nil {
			return nil, err
		}
		if p == 0 {
			return nil, errors.New("percentage must not be zero")
		}
		return dict("result", a/(p/100), "meaning", "the whole when "+fmtNum(a)+" is "+fmtNum(p)+"%"), nil
	}
	if m := reIncrease.FindStringSubmatch(t); m != nil {
		x, err := parseNum(m[1])
		if err != nil {
			return nil, err
		}
		p, err := parseNum(m[3])
		if err != nil {
			return nil, err
		}
		if m[2] == "-" || strings.EqualFold(m[2], "minus") {
			p = -p
		}
		return dict("result", x*(1+p/100), "change", x*p/100), nil
	}
	if m := reChange.FindStringSubmatch(t); m != nil {
		a, b, err := pair(m)
		if err != nil {
			return nil, err
		}
		if a == 0 {
			return nil, errors.New("percent change from zero is undefined")
		}
		return dict("result", (b-a)/math.Abs(a)*100, "difference", b-a), nil
	}
	return nil, errors.New(`try "15% of 80", "20 is what % of 50", "from 80 to 100", "50 is 25% of what" or "200 + 10%"`)
}

func init() {
	reg("percent_text", "Percentage", "Answers a percentage question typed in words. Understands \"15% of 80\", \"20 is what % of 50\", \"what percent is 20 of 50\", \"50 is 25% of what\", \"from 80 to 100\" (percent change) and \"200 + 10%\" / \"200 - 10%\". Returns a dict with result plus an explanation.",
		"dict[str, str]", `math.percent_text("15% of 80")`, 1, "percent percentage of change increase decrease what percent calculator ratio",
		[]P{rs("text", "The question, e.g. 15% of 80.")},
		func(a tools.Args) (any, error) { return percentText(a.Str("text")) },
		accepts(func(s string) bool {
			if len(s) > 100 {
				return false
			}
			_, err := percentText(s)
			return err == nil
		}), kw("percent", "pct", "%"))

	f2 := func(name, title, doc, ex, tags string, p1, p2 string, order int, f func(a, b float64) (float64, error)) {
		reg(name, title, doc, "float", ex, order, tags,
			[]P{rf(p1, p1+" value."), rf(p2, p2+" value.")},
			func(a tools.Args) (any, error) {
				r, err := f(a.Float(p1), a.Float(p2))
				if err != nil {
					return nil, err
				}
				return finite(name, r)
			})
	}
	f2("percent_of", "Percent of", "Returns p percent of x (percent_of(15, 80) = 12).", `math.percent_of(15, 80)`, "percent of percentage portion share part", "p", "x", 2,
		func(p, x float64) (float64, error) { return p / 100 * x, nil })
	f2("what_percent", "What percent", "Returns what percent a is of b (what_percent(20, 50) = 40). b must not be zero.", `math.what_percent(20, 50)`, "what percent ratio share fraction proportion", "a", "b", 20,
		func(a, b float64) (float64, error) {
			if b == 0 {
				return 0, errors.New("b must not be zero")
			}
			return a / b * 100, nil
		})
	f2("percent_change", "Percent change", "Returns the percentage change from old to new ((new-old)/|old|*100; 80 to 100 is 25). old must not be zero.", `math.percent_change(80, 100)`, "percent change increase decrease growth difference delta", "old", "new", 21,
		func(o, n float64) (float64, error) {
			if o == 0 {
				return 0, errors.New("old must not be zero")
			}
			return (n - o) / math.Abs(o) * 100, nil
		})
	f2("add_percent", "Add percent", "Increases x by p percent (add_percent(200, 10) = 220).", `math.add_percent(200, 10)`, "add percent increase markup plus tax raise", "x", "p", 22,
		func(x, p float64) (float64, error) { return x * (1 + p/100), nil })
	f2("sub_percent", "Subtract percent", "Decreases x by p percent (sub_percent(200, 10) = 180).", `math.sub_percent(200, 10)`, "subtract percent decrease discount minus reduce", "x", "p", 23,
		func(x, p float64) (float64, error) { return x * (1 - p/100), nil })
	f2("percent_diff", "Percent difference", "Returns the symmetric percentage difference between a and b: |a-b| divided by their average (50 and 100 give 66.667). Both zero gives 0.", `math.percent_diff(50, 100)`, "percent difference symmetric relative difference compare", "a", "b", 24,
		func(a, b float64) (float64, error) {
			avg := (math.Abs(a) + math.Abs(b)) / 2
			if avg == 0 {
				return 0, nil
			}
			return math.Abs(a-b) / avg * 100, nil
		})
	f2("markup", "Markup", "Returns the markup percentage of selling at price when the cost is cost: (price-cost)/cost*100. cost must not be zero.", `math.markup(80, 100)`, "markup percent profit cost price retail", "cost", "price", 25,
		func(c, p float64) (float64, error) {
			if c == 0 {
				return 0, errors.New("cost must not be zero")
			}
			return (p - c) / c * 100, nil
		})
	f2("margin", "Margin", "Returns the profit margin percentage of selling at price with cost cost: (price-cost)/price*100. price must not be zero.", `math.margin(80, 100)`, "margin percent profit gross margin cost price", "cost", "price", 26,
		func(c, p float64) (float64, error) {
			if p == 0 {
				return 0, errors.New("price must not be zero")
			}
			return (p - c) / p * 100, nil
		})
	f2("price_from_markup", "Price from markup", "Returns the selling price for a cost and a markup percentage (cost 80, markup 25 gives 100).", `math.price_from_markup(80, 25)`, "price markup cost selling price retail", "cost", "markup_pct", 27,
		func(c, m float64) (float64, error) { return c * (1 + m/100), nil })
	f2("price_from_margin", "Price from margin", "Returns the selling price needed for a target profit margin percentage (cost 80, margin 20 gives 100). margin_pct must be below 100.", `math.price_from_margin(80, 20)`, "price margin cost selling price target profit", "cost", "margin_pct", 28,
		func(c, m float64) (float64, error) {
			if m >= 100 {
				return 0, errors.New("margin_pct must be below 100")
			}
			return c / (1 - m/100), nil
		})
	f2("tax_add", "Add tax", "Adds tax to a net amount: amount * (1 + rate/100) (tax_add(100, 20) = 120).", `math.tax_add(100, 20)`, "tax add vat gst sales tax gross net price", "amount", "rate", 29,
		func(a, r float64) (float64, error) { return a * (1 + r/100), nil })
	f2("tax_remove", "Remove tax", "Removes tax from a gross amount: amount / (1 + rate/100) (tax_remove(120, 20) = 100). rate must be above -100.", `math.tax_remove(120, 20)`, "tax remove vat gst sales tax net gross exclusive", "amount", "rate", 30,
		func(a, r float64) (float64, error) {
			if r <= -100 {
				return 0, errors.New("rate must be above -100")
			}
			return a / (1 + r/100), nil
		})

	reg("reverse_percent", "Reverse percent", "Finds the original value before a percentage was added (increase=true) or taken off (increase=false): reverse_percent(120, 20) = 100, reverse_percent(80, 20, False) = 100.",
		"float", `math.reverse_percent(120, 20)`, 31, "reverse percent original before discount undo increase find original price",
		[]P{rf("final", "The value after the change."), rf("p", "The percentage that was applied."), ob("increase", "True if p was added, false if it was subtracted.", true)},
		func(a tools.Args) (any, error) {
			f := 1 + a.Float("p")/100
			if !a.Bool("increase") {
				f = 1 - a.Float("p")/100
			}
			if f == 0 {
				return nil, errors.New("percentage makes the original undefined")
			}
			return finite("reverse_percent", a.Float("final")/f)
		})

	reg("markup_margin", "Markup and margin", "Returns profit, markup percent and margin percent for a cost and a selling price as a dict.",
		"dict[str, str]", `math.markup_margin(80, 100)`, 32, "markup margin profit cost price dict",
		[]P{rf("cost", "Cost."), rf("price", "Selling price.")},
		func(a tools.Args) (any, error) {
			c, p := a.Float("cost"), a.Float("price")
			if c == 0 || p == 0 {
				return nil, errors.New("cost and price must not be zero")
			}
			return dict("cost", c, "price", p, "profit", p-c, "markup", (p-c)/c*100, "margin", (p-c)/p*100), nil
		})

	reg("discount", "Discount", "Applies a percentage discount (optionally a second stacked one) to a price and returns final price, amount saved and the effective percent as a dict.",
		"dict[str, str]", `math.discount(80, 25)`, 33, "discount sale off price percent saved final coupon",
		[]P{rf("price", "Original price."), rf("pct", "Discount percent."), of("pct2", "Optional second discount percent applied after the first.", 0)},
		func(a tools.Args) (any, error) {
			p := a.Float("price")
			f := (1 - a.Float("pct")/100) * (1 - a.Float("pct2")/100)
			fin := p * f
			return dict("final", fin, "saved", p-fin, "effective_percent", (1-f)*100), nil
		})

	reg("vat", "VAT split", "Splits an amount into net, VAT and gross. With inclusive=false the amount is net and VAT is added; with inclusive=true the amount already includes VAT and is split out.",
		"dict[str, str]", `math.vat(100, 20)`, 34, "vat tax gst split net gross inclusive exclusive sales tax",
		[]P{rf("amount", "The amount."), rf("rate", "VAT rate in percent."), ob("inclusive", "The amount already includes VAT.", false)},
		func(a tools.Args) (any, error) {
			amt, r := a.Float("amount"), a.Float("rate")
			if r <= -100 {
				return nil, errors.New("rate must be above -100")
			}
			net, gross := amt, amt*(1+r/100)
			if a.Bool("inclusive") {
				net, gross = amt/(1+r/100), amt
			}
			return dict("net", net, "vat", gross-net, "gross", gross), nil
		})

	tipFn := func(bill, pct float64, people int64) (map[string]string, error) {
		if people < 1 || people > 100000 {
			return nil, errors.New("people must be between 1 and 100000")
		}
		tip := bill * pct / 100
		total := bill + tip
		n := float64(people)
		return moneyDict("tip", tip, "total", total, "per_person", total/n, "tip_per_person", tip/n, "bill_per_person", bill/n), nil
	}
	reg("tip", "Tip", "Calculates a tip and splits the bill: returns tip, total, per_person, tip_per_person and bill_per_person (rounded to cents) for a bill, a tip percent and a number of people.",
		"dict[str, str]", `math.tip(85.5, 18, 2)`, 35, "tip gratuity bill split restaurant people per person",
		[]P{rf("bill", "Bill before tip."), of("pct", "Tip percent.", 15), oi("people", "Number of people sharing.", 1)},
		func(a tools.Args) (any, error) { return tipFn(a.Float("bill"), a.Float("pct"), a.Int("people")) })

	reg("tip_text", "Tip calculator", "Reads \"bill [tip%] [people]\" from text (\"85.50 18% 2\" or \"85.5 18 2\"; tip defaults to 15%, people to 1) and returns the same dict as tip.",
		"dict[str, str]", `math.tip_text("85.50 18% 2")`, 4, "tip gratuity bill split restaurant people calculator",
		[]P{rs("text", "Bill, optional tip percent, optional number of people.")},
		func(a tools.Args) (any, error) {
			nums := extractNumbers(a.Str("text"))
			if len(nums) == 0 || len(nums) > 3 {
				return nil, errors.New("expected: bill [tip%] [people]")
			}
			pct, people := 15.0, 1.0
			if len(nums) > 1 {
				pct = nums[1]
			}
			if len(nums) > 2 {
				people = nums[2]
			}
			return tipFn(nums[0], pct, int64(people))
		}, accepts(func(s string) bool {
			n := extractNumbers(s)
			return len(n) >= 1 && len(n) <= 3 && len(s) < 60 && strings.ContainsAny(s, "% ")
		}), kw("tip"))
}

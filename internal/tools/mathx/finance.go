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

func loanCalc(principal, annualRate, years float64, perYear int64) (map[string]string, error) {
	if principal <= 0 || math.IsInf(principal, 0) {
		return nil, errors.New("principal must be positive")
	}
	if annualRate < 0 || annualRate > 1000 {
		return nil, errors.New("annual_rate must be between 0 and 1000 percent")
	}
	if years <= 0 || years > 100 {
		return nil, errors.New("years must be between 0 and 100")
	}
	if perYear < 1 || perYear > 366 {
		return nil, errors.New("per_year must be between 1 and 366")
	}
	n := math.Round(years * float64(perYear))
	if n < 1 {
		return nil, errors.New("the loan has no payments")
	}
	r := annualRate / 100 / float64(perYear)
	pay := principal / n
	if r > 0 {
		pay = principal * r / (1 - math.Pow(1+r, -n))
	}
	total := pay * n
	return moneyDict("payment", pay, "total", total, "interest", total-principal, "payments", fmt.Sprint(int64(n))), nil
}

func compound(principal, rate, years float64, n int64) (map[string]string, error) {
	if n < 0 || n > 100000 {
		return nil, errors.New("n must be between 0 (continuous) and 100000")
	}
	if years < 0 {
		return nil, errors.New("years must not be negative")
	}
	r := rate / 100
	var amount float64
	if n == 0 {
		amount = principal * math.Exp(r*years)
	} else {
		amount = principal * math.Pow(1+r/float64(n), float64(n)*years)
	}
	if math.IsInf(amount, 0) || math.IsNaN(amount) {
		return nil, errors.New("result is too large")
	}
	m := moneyDict("amount", amount, "interest", amount-principal)
	if principal > 0 && years > 0 {
		m["effective_annual_rate"] = fmtNum((math.Pow(amount/principal, 1/years) - 1) * 100)
	}
	return m, nil
}

var timeTerm = regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*([hms])`)

func parseDuration(t string) (float64, error) {
	t = strings.TrimSpace(t)
	if t == "" {
		return 0, errors.New("empty time")
	}
	if strings.Contains(t, ":") {
		parts := strings.Split(t, ":")
		if len(parts) > 3 {
			return 0, fmt.Errorf("%q has too many parts", t)
		}
		mult := []float64{3600, 60, 1}
		if len(parts) == 1 {
			return 0, fmt.Errorf("%q needs a colon or unit letters (h, m, s)", t)
		}
		var total float64
		for i, p := range parts {
			v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
			if err != nil || v < 0 {
				return 0, fmt.Errorf("%q is not a time", t)
			}
			total += v * mult[i]
		}
		return total, nil
	}
	locs := timeTerm.FindAllStringSubmatchIndex(t, -1)
	if locs == nil || len(strings.TrimSpace(timeTerm.ReplaceAllString(t, ""))) > 0 {
		return 0, fmt.Errorf("%q is not a time (use h:mm, h:mm:ss or 1h 30m)", t)
	}
	var total float64
	for _, m := range timeTerm.FindAllStringSubmatch(t, -1) {
		v, _ := strconv.ParseFloat(m[1], 64)
		switch strings.ToLower(m[2]) {
		case "h":
			total += v * 3600
		case "m":
			total += v * 60
		default:
			total += v
		}
	}
	return total, nil
}

// splitSigned splits "a + b - c" into signed terms. A leading sign is allowed.
func splitSigned(s string) ([]float64, []string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil, errors.New("empty expression")
	}
	var signs []float64
	var terms []string
	sign, start := 1.0, 0
	flush := func(end int) error {
		t := strings.TrimSpace(s[start:end])
		if t == "" {
			if len(terms) == 0 && end == 0 {
				return nil
			}
			return errors.New("missing value next to an operator")
		}
		signs = append(signs, sign)
		terms = append(terms, t)
		return nil
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == '+' || c == '-' && (i == 0 || s[i-1] == ' ' || s[i-1] == '\t') {
			if i == 0 {
				if c == '-' {
					sign = -1
				}
				start = 1
				continue
			}
			if err := flush(i); err != nil {
				return nil, nil, err
			}
			sign = 1
			if c == '-' {
				sign = -1
			}
			start = i + 1
		}
	}
	if err := flush(len(s)); err != nil {
		return nil, nil, err
	}
	return signs, terms, nil
}

func fmtDuration(sec float64) string {
	neg := sec < 0
	sec = math.Round(math.Abs(sec)*1000) / 1000
	whole := int64(sec)
	h, m, s := whole/3600, whole%3600/60, whole%60
	out := fmt.Sprintf("%d:%02d", h, m)
	if frac := sec - float64(whole); s != 0 || frac != 0 {
		ss := fmtNum(float64(s) + frac)
		if s < 10 {
			ss = "0" + ss
		}
		out += ":" + ss
	}
	if neg && sec != 0 {
		out = "-" + out
	}
	return out
}

var dmsRe = regexp.MustCompile(`(?i)^\s*(-?)\s*(\d+(?:\.\d+)?)\s*(?:°|º|d|deg)?\s*(?:(\d+(?:\.\d+)?)\s*(?:'|′|m|min)?\s*)?(?:(\d+(?:\.\d+)?)\s*(?:"|″|''|s|sec)?\s*)?([NSEW])?\s*$`)

func parseDMS(s string) (float64, error) {
	m := dmsRe.FindStringSubmatch(s)
	if m == nil || len(s) > 60 {
		return 0, fmt.Errorf("%q is not an angle like 40°26'46\"N or 40 26 46", s)
	}
	f := func(x string) float64 { v, _ := strconv.ParseFloat(x, 64); return v }
	deg := f(m[2]) + f(m[3])/60 + f(m[4])/3600
	if m[1] == "-" || strings.EqualFold(m[5], "S") || strings.EqualFold(m[5], "W") {
		deg = -deg
	}
	return deg, nil
}

func formatDMS(deg float64) string {
	neg := deg < 0
	secs := math.Round(math.Abs(deg)*3600*100) / 100
	d := math.Floor(secs / 3600)
	m := math.Floor((secs - d*3600) / 60)
	s := math.Round((secs-d*3600-m*60)*100) / 100
	out := fmt.Sprintf("%d°%02d'%s\"", int64(d), int64(m), strconv.FormatFloat(s, 'f', -1, 64))
	if neg && secs != 0 {
		out = "-" + out
	}
	return out
}

func lgamma(x float64) float64 { v, _ := math.Lgamma(x); return v }

func init() {
	reg("compound_interest", "Compound interest", "Future value with compound interest: principal * (1 + rate/n)^(n*years), rate in percent per year, n compounding periods per year (n=1 annual, 12 monthly, 0 continuous). Returns amount, interest and the effective annual rate as a dict.",
		"dict[str, str]", `math.compound_interest(1000, 5, 10, 1)`, 110, "compound interest investment savings growth apy future value compounding",
		[]P{rf("principal", "Starting amount."), rf("rate", "Annual rate in percent."), rf("years", "Duration in years."), oi("n", "Compounding periods per year (0 = continuous).", 1)},
		func(a tools.Args) (any, error) {
			return compound(a.Float("principal"), a.Float("rate"), a.Float("years"), a.Int("n"))
		})
	reg("simple_interest", "Simple interest", "Simple (non-compounding) interest: principal * rate/100 * years. Returns interest and total as a dict.",
		"dict[str, str]", `math.simple_interest(1000, 5, 3)`, 111, "simple interest loan deposit principal rate years",
		[]P{rf("principal", "Amount."), rf("rate", "Annual rate in percent."), rf("years", "Duration in years.")},
		func(a tools.Args) (any, error) {
			i := a.Float("principal") * a.Float("rate") / 100 * a.Float("years")
			return moneyDict("interest", i, "total", a.Float("principal")+i), nil
		})
	reg("loan_payment", "Loan payment", "Fixed payment of an amortising loan or mortgage: principal, annual rate in percent, term in years and payments per year (default 12). Returns payment, total repaid, total interest and the number of payments, rounded to cents (200000 at 6% for 30 years gives 1199.10 a month).",
		"dict[str, str]", `math.loan_payment(200000, 6, 30)`, 3, "loan payment mortgage monthly emi installment amortization interest car loan",
		[]P{rf("principal", "Amount borrowed."), rf("annual_rate", "Annual interest rate in percent (0 allowed)."), rf("years", "Term in years."), oi("per_year", "Payments per year.", 12)},
		func(a tools.Args) (any, error) {
			return loanCalc(a.Float("principal"), a.Float("annual_rate"), a.Float("years"), a.Int("per_year"))
		})
	reg("loan_text", "Loan calculator", "Reads \"principal rate% years [payments_per_year]\" from text (\"200000 6% 30\") and returns the same dict as loan_payment: payment, total, interest and payments.",
		"dict[str, str]", `math.loan_text("200000 6% 30")`, 5, "loan payment mortgage monthly emi installment calculator interest",
		[]P{rs("text", "Principal, annual rate percent, years, optional payments per year.")},
		func(a tools.Args) (any, error) {
			n := extractNumbers(a.Str("text"))
			if len(n) < 3 || len(n) > 4 {
				return nil, errors.New("expected: principal rate% years [payments per year]")
			}
			py := int64(12)
			if len(n) == 4 {
				py = int64(n[3])
			}
			return loanCalc(n[0], n[1], n[2], py)
		}, accepts(func(s string) bool {
			n := extractNumbers(s)
			return len(s) < 60 && (len(n) == 3 || len(n) == 4) && strings.Contains(s, "%")
		}), kw("loan", "mortgage"))
	reg("amortization_summary", "Amortization summary", "Splits a loan into interest and principal: after `after` payments returns interest_paid, principal_paid, balance, first_interest, first_principal, payment and payments_left (same inputs as loan_payment).",
		"dict[str, str]", `math.amortization_summary(200000, 6, 30, 12, 12)`, 112, "amortization schedule loan balance principal interest split remaining mortgage",
		[]P{rf("principal", "Amount borrowed."), rf("annual_rate", "Annual rate in percent."), rf("years", "Term in years."), oi("per_year", "Payments per year.", 12), oi("after", "Number of payments already made.", 12)},
		func(a tools.Args) (any, error) {
			l, err := loanCalc(a.Float("principal"), a.Float("annual_rate"), a.Float("years"), a.Int("per_year"))
			if err != nil {
				return nil, err
			}
			n := int64(math.Round(a.Float("years") * float64(a.Int("per_year"))))
			after, err := intIn("after", a.Int("after"), 0, n)
			if err != nil {
				return nil, err
			}
			_ = l
			r := a.Float("annual_rate") / 100 / float64(a.Int("per_year"))
			pay := a.Float("principal") / float64(n)
			if r > 0 {
				pay = a.Float("principal") * r / (1 - math.Pow(1+r, -float64(n)))
			}
			bal := a.Float("principal")
			var ip, pp, fi, fp float64
			for i := int64(0); i < after; i++ {
				in := bal * r
				pr := pay - in
				if i == 0 {
					fi, fp = in, pr
				}
				ip += in
				pp += pr
				bal -= pr
			}
			if after == 0 {
				fi = bal * r
				fp = pay - fi
			}
			if bal < 0.005 {
				bal = 0
			}
			return moneyDict("payment", pay, "interest_paid", ip, "principal_paid", pp, "balance", bal, "first_interest", fi, "first_principal", fp, "payments_left", fmt.Sprint(n-after)), nil
		})
	reg("future_value", "Future value", "Future value of a starting amount plus optional regular payments at the end of each period: present*(1+r)^n + payment*((1+r)^n-1)/r, with rate the percent per period.",
		"float", `math.future_value(1000, 5, 10, 100)`, 113, "future value fv investment annuity savings growth periodic payments",
		[]P{rf("present", "Starting amount."), rf("rate", "Rate per period in percent."), rf("periods", "Number of periods."), of("payment", "Payment added at the end of every period.", 0)},
		func(a tools.Args) (any, error) {
			r, n := a.Float("rate")/100, a.Float("periods")
			g := math.Pow(1+r, n)
			ann := a.Float("payment") * n
			if r != 0 {
				ann = a.Float("payment") * (g - 1) / r
			}
			return finite("future_value", a.Float("present")*g+ann)
		})
	reg("present_value", "Present value", "Present value of a future amount plus optional regular payments, discounted at a percent rate per period: future/(1+r)^n + payment*(1-(1+r)^-n)/r.",
		"float", `math.present_value(1628.89, 5, 10)`, 114, "present value pv discount npv annuity worth today",
		[]P{rf("future", "Future amount."), rf("rate", "Discount rate per period in percent."), rf("periods", "Number of periods."), of("payment", "Payment received at the end of every period.", 0)},
		func(a tools.Args) (any, error) {
			r, n := a.Float("rate")/100, a.Float("periods")
			d := math.Pow(1+r, -n)
			ann := a.Float("payment") * n
			if r != 0 {
				ann = a.Float("payment") * (1 - d) / r
			}
			return finite("present_value", a.Float("future")*d+ann)
		})
	reg("cagr", "CAGR", "Compound annual growth rate in percent from a start value to an end value over some years ((end/start)^(1/years)-1). Start and end must be positive; years > 0.",
		"float", `math.cagr(1000, 2000, 5)`, 115, "cagr compound annual growth rate return investment annualized",
		[]P{rf("start", "Start value."), rf("end", "End value."), rf("years", "Years.")},
		func(a tools.Args) (any, error) {
			if a.Float("start") <= 0 || a.Float("end") <= 0 || a.Float("years") <= 0 {
				return nil, errors.New("start, end and years must be positive")
			}
			return finite("cagr", (math.Pow(a.Float("end")/a.Float("start"), 1/a.Float("years"))-1)*100)
		})
	reg("roi", "Return on investment", "Return on investment in percent: (final - initial) / initial * 100. initial must not be zero.",
		"float", `math.roi(1000, 1250)`, 116, "roi return on investment profit percent gain",
		[]P{rf("initial", "Amount invested."), rf("final", "Value at the end.")},
		func(a tools.Args) (any, error) {
			if a.Float("initial") == 0 {
				return nil, errors.New("initial must not be zero")
			}
			return (a.Float("final") - a.Float("initial")) / a.Float("initial") * 100, nil
		})
	reg("rule_of_72", "Rule of 72", "Years to double money at an annual percent rate: returns the rule-of-72 estimate and the exact value ln(2)/ln(1+r) as a dict. rate must be positive.",
		"dict[str, str]", `math.rule_of_72(6)`, 117, "rule of 72 doubling time investment years to double",
		[]P{rf("rate", "Annual rate in percent.")},
		func(a tools.Args) (any, error) {
			if a.Float("rate") <= 0 {
				return nil, errors.New("rate must be positive")
			}
			return dict("estimate", 72/a.Float("rate"), "exact", math.Ln2/math.Log1p(a.Float("rate")/100)), nil
		})
	reg("break_even", "Break-even", "Break-even point: units = fixed_costs / (price - variable_cost). Returns exact units, units rounded up and the revenue at break-even. price must exceed variable_cost.",
		"dict[str, str]", `math.break_even(5000, 25, 15)`, 118, "break even point units fixed costs contribution margin business",
		[]P{rf("fixed_costs", "Fixed costs."), rf("price", "Selling price per unit."), rf("variable_cost", "Variable cost per unit.")},
		func(a tools.Args) (any, error) {
			cm := a.Float("price") - a.Float("variable_cost")
			if cm <= 0 {
				return nil, errors.New("price must exceed variable_cost")
			}
			u := a.Float("fixed_costs") / cm
			return dict("units", u, "units_rounded_up", math.Ceil(u), "revenue", math.Ceil(u)*a.Float("price"), "contribution_margin", cm), nil
		})
	reg("unit_price", "Unit price", "Price per unit: price / quantity (quantity must be positive).",
		"float", `math.unit_price(3.5, 12)`, 119, "unit price per unit cost per item",
		[]P{rf("price", "Total price."), rf("qty", "Quantity.")},
		func(a tools.Args) (any, error) {
			if a.Float("qty") <= 0 {
				return nil, errors.New("qty must be positive")
			}
			return a.Float("price") / a.Float("qty"), nil
		})
	reg("compare_unit_price", "Compare unit prices", "Compares two offers by unit price: returns each unit price, which is cheaper (a, b or same) and the saving percent of the cheaper one.",
		"dict[str, str]", `math.compare_unit_price(3.5, 12, 5, 20)`, 120, "compare unit price better deal cheaper shopping value",
		[]P{rf("a_price", "Price of offer A."), rf("a_qty", "Quantity of offer A."), rf("b_price", "Price of offer B."), rf("b_qty", "Quantity of offer B.")},
		func(a tools.Args) (any, error) {
			if a.Float("a_qty") <= 0 || a.Float("b_qty") <= 0 {
				return nil, errors.New("quantities must be positive")
			}
			ua, ub := a.Float("a_price")/a.Float("a_qty"), a.Float("b_price")/a.Float("b_qty")
			best, save := "same", 0.0
			switch {
			case ua < ub:
				best, save = "a", (1-ua/ub)*100
			case ub < ua:
				best, save = "b", (1-ub/ua)*100
			}
			return dict("unit_a", ua, "unit_b", ub, "cheaper", best, "saving_percent", save), nil
		})
	speed := func(name, title, doc, ex, p1, p2 string, f func(x, y float64) float64) {
		reg(name, title, doc, "float", ex, 121, "speed distance time velocity travel average pace "+strings.ReplaceAll(name, "_", " "),
			[]P{rf(p1, p1+"."), rf(p2, p2+".")},
			func(a tools.Args) (any, error) {
				if a.Float(p2) <= 0 {
					return nil, fmt.Errorf("%s must be positive", p2)
				}
				return finite(name, f(a.Float(p1), a.Float(p2)))
			})
	}
	speed("average_speed", "Average speed", "Average speed = distance / time (any consistent units). time must be positive.", `math.average_speed(150, 2.5)`, "distance", "time", func(d, t float64) float64 { return d / t })
	speed("travel_time", "Travel time", "Time = distance / speed. speed must be positive.", `math.travel_time(150, 60)`, "distance", "speed", func(d, s float64) float64 { return d / s })
	reg("travel_distance", "Travel distance", "Distance = speed * time.",
		"float", `math.travel_distance(60, 2.5)`, 122, "distance speed time travel how far",
		[]P{rf("speed", "Speed."), rf("time", "Time.")},
		func(a tools.Args) (any, error) { return finite("travel_distance", a.Float("speed")*a.Float("time")) })
	reg("bmi", "BMI", "Body mass index = kg / m^2, with the WHO category (underweight, normal, overweight, obese). Weight in kilograms, height in metres; both must be positive.",
		"dict[str, str]", `math.bmi(70, 1.75)`, 123, "bmi body mass index weight height health",
		[]P{rf("kg", "Weight in kilograms."), rf("m", "Height in metres.")},
		func(a tools.Args) (any, error) {
			if a.Float("kg") <= 0 || a.Float("m") <= 0 {
				return nil, errors.New("kg and m must be positive")
			}
			b := a.Float("kg") / (a.Float("m") * a.Float("m"))
			c := "obese"
			switch {
			case b < 18.5:
				c = "underweight"
			case b < 25:
				c = "normal"
			case b < 30:
				c = "overweight"
			}
			rb, _ := roundTo(b, 1, "half_up")
			return dict("bmi", rb, "category", c), nil
		})

	reg("time_sum", "Time sum", "Adds and subtracts clock times: \"1:30 + 2:45\" gives \"4:15\"; also h:mm:ss and unit forms \"1h 30m + 45m\". Hours may exceed 24; a negative total keeps its minus sign; seconds appear only when non-zero.",
		"str", `math.time_sum("1:30 + 2:45")`, 7, "time sum add subtract hours minutes duration clock hh:mm timesheet",
		[]P{rs("expr", "Terms joined by + and -, e.g. 1:30 + 2:45 - 0:20.")},
		func(a tools.Args) (any, error) {
			if len(a.Str("expr")) > 2000 {
				return nil, errors.New("expression is too long")
			}
			signs, terms, err := splitSigned(a.Str("expr"))
			if err != nil {
				return nil, err
			}
			var total float64
			for i, t := range terms {
				d, err := parseDuration(t)
				if err != nil {
					return nil, err
				}
				total += signs[i] * d
			}
			return fmtDuration(total), nil
		}, accepts(func(s string) bool {
			if len(s) > 200 || !strings.ContainsAny(s, ":hms") || !strings.ContainsAny(s, "+-") {
				return false
			}
			_, terms, err := splitSigned(s)
			if err != nil || len(terms) < 2 {
				return false
			}
			for _, t := range terms {
				if _, err := parseDuration(t); err != nil {
					return false
				}
			}
			return true
		}), kw("time", "hours"))
	reg("angle_sum", "Angle sum", "Adds and subtracts angles written as degrees/minutes/seconds: \"30°15'20\\\" + 45°50'\" gives 76°05'20\". Each term accepts 40°26'46\"N style, 40 26 46 or plain degrees. Result is in d°mm'ss\" form.",
		"str", `math.angle_sum("30 15 20 + 45 50")`, 124, "angle sum add subtract degrees minutes seconds dms",
		[]P{rs("expr", "Angles joined by + and -.")},
		func(a tools.Args) (any, error) {
			signs, terms, err := splitSigned(a.Str("expr"))
			if err != nil {
				return nil, err
			}
			var total float64
			for i, t := range terms {
				d, err := parseDMS(t)
				if err != nil {
					return nil, err
				}
				total += signs[i] * d
			}
			return formatDMS(total), nil
		})
	reg("dms_to_decimal", "DMS to decimal degrees", "Converts degrees/minutes/seconds (40°26'46\"N, 40 26 46, -73°59'W) to decimal degrees; S and W (or a leading minus) make it negative.",
		"float", `math.dms_to_decimal("40°26'46\"N")`, 125, "dms degrees minutes seconds decimal degrees latitude longitude gps coordinates",
		[]P{rs("dms", "Angle text.")}, func(a tools.Args) (any, error) { return parseDMS(a.Str("dms")) })
	reg("decimal_to_dms", "Decimal degrees to DMS", "Converts decimal degrees to d°mm'ss.ss\" form (40.446 gives 40°26'45.6\"); the sign is kept as a minus.",
		"str", `math.decimal_to_dms(40.446)`, 126, "decimal degrees dms minutes seconds latitude longitude gps coordinates",
		[]P{rf("degrees", "Decimal degrees.")},
		func(a tools.Args) (any, error) {
			if math.IsNaN(a.Float("degrees")) || math.Abs(a.Float("degrees")) > 1e9 {
				return nil, errors.New("degrees out of range")
			}
			return formatDMS(a.Float("degrees")), nil
		})

	reg("combinations", "Combinations", "Number of ways to choose k items from n without order, C(n, k), as a decimal string (combinations(52, 5) = 2598960). k > n gives 0. n is limited to 0..20000.",
		"str", `math.combinations(52, 5)`, 127, "combinations choose binomial coefficient ncr n choose k counting",
		[]P{ri("n", "Total items."), ri("k", "Items chosen.")},
		func(a tools.Args) (any, error) {
			n, err := intIn("n", a.Int("n"), 0, 20000)
			if err != nil {
				return nil, err
			}
			k, err := intIn("k", a.Int("k"), 0, 1<<40)
			if err != nil {
				return nil, err
			}
			if k > n {
				return "0", nil
			}
			return new(big.Int).Binomial(n, k).String(), nil
		})
	reg("permutations", "Permutations", "Number of ordered arrangements of k items from n, P(n, k) = n!/(n-k)!, as a decimal string (permutations(10, 3) = 720). k > n gives 0. n is limited to 0..20000.",
		"str", `math.permutations(10, 3)`, 128, "permutations arrangements npr ordered counting",
		[]P{ri("n", "Total items."), ri("k", "Items arranged.")},
		func(a tools.Args) (any, error) {
			n, err := intIn("n", a.Int("n"), 0, 20000)
			if err != nil {
				return nil, err
			}
			k, err := intIn("k", a.Int("k"), 0, 1<<40)
			if err != nil {
				return nil, err
			}
			if k > n {
				return "0", nil
			}
			if k == 0 {
				return "1", nil
			}
			return new(big.Int).MulRange(n-k+1, n).String(), nil
		})
	reg("binomial_probability", "Binomial probability", "Probability of exactly k successes in n independent trials with success probability p: C(n,k) p^k (1-p)^(n-k). n up to 10 million; p in 0..1.",
		"float", `math.binomial_probability(10, 5, 0.5)`, 129, "binomial probability distribution trials successes coin flips statistics pmf",
		[]P{ri("n", "Trials."), ri("k", "Successes."), rf("p", "Success probability, 0..1.")},
		func(a tools.Args) (any, error) {
			n, err := intIn("n", a.Int("n"), 0, 10_000_000)
			if err != nil {
				return nil, err
			}
			k, p := a.Int("k"), a.Float("p")
			if p < 0 || p > 1 || math.IsNaN(p) {
				return nil, errors.New("p must be between 0 and 1")
			}
			if k < 0 || k > n {
				return 0.0, nil
			}
			if p == 0 {
				return b2f(k == 0), nil
			}
			if p == 1 {
				return b2f(k == n), nil
			}
			fn, fk := float64(n), float64(k)
			lg := lgamma(fn+1) - lgamma(fk+1) - lgamma(fn-fk+1) + fk*math.Log(p) + (fn-fk)*math.Log1p(-p)
			return math.Exp(lg), nil
		})
	reg("normal_cdf", "Normal CDF", "Cumulative probability P(X <= x) of a normal distribution with the given mean (default 0) and standard deviation sd (default 1, must be positive). normal_cdf(1.96) is about 0.975.",
		"float", `math.normal_cdf(1.96)`, 130, "normal distribution cdf gaussian probability bell curve z score cumulative",
		[]P{rf("x", "Value."), of("mean", "Mean.", 0), of("sd", "Standard deviation.", 1)},
		func(a tools.Args) (any, error) {
			if a.Float("sd") <= 0 {
				return nil, errors.New("sd must be positive")
			}
			return 0.5 * math.Erfc(-(a.Float("x")-a.Float("mean"))/(a.Float("sd")*math.Sqrt2)), nil
		})
	reg("normal_pdf", "Normal PDF", "Probability density of a normal distribution at x (mean default 0, sd default 1, sd must be positive).",
		"float", `math.normal_pdf(0)`, 131, "normal distribution pdf gaussian density bell curve",
		[]P{rf("x", "Value."), of("mean", "Mean.", 0), of("sd", "Standard deviation.", 1)},
		func(a tools.Args) (any, error) {
			if a.Float("sd") <= 0 {
				return nil, errors.New("sd must be positive")
			}
			z := (a.Float("x") - a.Float("mean")) / a.Float("sd")
			return math.Exp(-z*z/2) / (a.Float("sd") * math.Sqrt(2*math.Pi)), nil
		})
	reg("dice_probability", "Dice probability", "Probability that `count` dice with `sides` sides sum to exactly `target` (two six-sided dice summing to 7 gives 1/6). count*sides is limited to 5000.",
		"float", `math.dice_probability(6, 2, 7)`, 132, "dice probability sum roll odds die sides gaming",
		[]P{ri("sides", "Sides per die (>= 2)."), ri("count", "Number of dice (>= 1)."), ri("target", "Target sum.")},
		func(a tools.Args) (any, error) {
			sides, err := intIn("sides", a.Int("sides"), 2, 20000)
			if err != nil {
				return nil, err
			}
			count, err := intIn("count", a.Int("count"), 1, 20000)
			if err != nil {
				return nil, err
			}
			if sides*count > 5000 {
				return nil, errors.New("sides * count must be at most 5000")
			}
			t := a.Int("target")
			if t < count || t > sides*count {
				return 0.0, nil
			}
			cur := []float64{1}
			for d := int64(0); d < count; d++ {
				next := make([]float64, len(cur)+int(sides)-1+1)
				// next[s] = sum_{j=1..sides} cur[s-j] / sides, with cur indexed by sum - d
				for s := range next {
					var sum float64
					for j := int64(0); j < sides; j++ {
						if i := int64(s) - j; i >= 0 && i < int64(len(cur)) {
							sum += cur[i]
						}
					}
					next[s] = sum / float64(sides)
				}
				cur = next[:len(cur)+int(sides)-1]
			}
			return cur[t-count], nil
		})
}

func b2f(b bool) float64 {
	if b {
		return 1
	}
	return 0
}

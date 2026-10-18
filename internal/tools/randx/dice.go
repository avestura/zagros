package randx

import (
	"fmt"
	"math/rand/v2"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

const (
	maxDicePerTerm  = 1000
	maxDiceTotal    = 10000
	maxDiceSides    = 1000000
	maxExplosions   = 100
	statsTrials     = 20000
	statsFixedSeed  = 1
	statsDiceBudget = 2000000
)

type diceTerm struct {
	sign      int64
	isConst   bool
	konst     int64
	count     int
	sides     int
	explode   bool
	mod       string // "", "kh", "kl", "dh", "dl"
	modN      int
	canonical string
}

var diceRe = regexp.MustCompile(`^(\d*)d(\d+|%)(!)?(?:(kh|kl|dh|dl|k|d)(\d*))?$`)

func parseDice(spec string) ([]diceTerm, error) {
	s := strings.ToLower(strings.Join(strings.Fields(spec), ""))
	if s == "" {
		return nil, fmt.Errorf("empty dice spec; try \"2d6+3\"")
	}
	var terms []diceTerm
	sign := int64(1)
	i := 0
	totalDice := 0
	for i < len(s) {
		if s[i] == '+' || s[i] == '-' {
			if s[i] == '-' {
				sign = -1
			} else {
				sign = 1
			}
			i++
		}
		j := i
		for j < len(s) && s[j] != '+' && s[j] != '-' {
			j++
		}
		tok := s[i:j]
		if tok == "" {
			return nil, fmt.Errorf("bad dice spec %q", spec)
		}
		i = j
		if n, err := strconv.ParseInt(tok, 10, 64); err == nil {
			terms = append(terms, diceTerm{sign: sign, isConst: true, konst: n, canonical: tok})
			sign = 1
			continue
		}
		m := diceRe.FindStringSubmatch(tok)
		if m == nil {
			return nil, fmt.Errorf("cannot read dice term %q (examples: d20, 2d6+3, 4d6kh3, 2d20kl1, 3d6!)", tok)
		}
		t := diceTerm{sign: sign, count: 1, canonical: tok}
		if m[1] != "" {
			n, _ := strconv.Atoi(m[1])
			t.count = n
		}
		if m[2] == "%" {
			t.sides = 100
		} else {
			t.sides, _ = strconv.Atoi(m[2])
		}
		t.explode = m[3] == "!"
		if m[4] != "" {
			t.mod = m[4]
			if t.mod == "k" {
				t.mod = "kh"
			} else if t.mod == "d" {
				t.mod = "dl"
			}
			t.modN = 1
			if m[5] != "" {
				t.modN, _ = strconv.Atoi(m[5])
			}
		}
		if t.count < 1 || t.count > maxDicePerTerm {
			return nil, fmt.Errorf("dice count must be 1..%d in %q", maxDicePerTerm, tok)
		}
		if t.sides < 1 || t.sides > maxDiceSides {
			return nil, fmt.Errorf("dice sides must be 1..%d in %q", maxDiceSides, tok)
		}
		if t.explode && t.sides < 2 {
			return nil, fmt.Errorf("a d1 cannot explode")
		}
		totalDice += t.count
		if totalDice > maxDiceTotal {
			return nil, fmt.Errorf("too many dice (max %d)", maxDiceTotal)
		}
		terms = append(terms, t)
		sign = 1
	}
	return terms, nil
}

type diceResult struct {
	total  int64
	detail []string
	rolls  []string
}

func (t diceTerm) kept() int {
	switch t.mod {
	case "kh", "kl":
		return min(t.modN, t.count)
	case "dh", "dl":
		return t.count - min(t.modN, t.count)
	}
	return t.count
}

func rollTerm(r *rand.Rand, t diceTerm) (sum int64, vals []int, keep []bool) {
	vals = make([]int, 0, t.count)
	for i := 0; i < t.count; i++ {
		v := r.IntN(t.sides) + 1
		vals = append(vals, v)
		for ex := 0; t.explode && v == t.sides && ex < maxExplosions; ex++ {
			v = r.IntN(t.sides) + 1
			vals = append(vals, v)
		}
	}
	keep = make([]bool, len(vals))
	for i := range keep {
		keep[i] = true
	}
	if t.mod != "" {
		idx := make([]int, len(vals))
		for i := range idx {
			idx[i] = i
		}
		sort.SliceStable(idx, func(a, b int) bool { return vals[idx[a]] < vals[idx[b]] }) // ascending
		n := min(t.modN, len(vals))
		for i := range keep {
			keep[i] = false
		}
		switch t.mod {
		case "kh":
			for _, k := range idx[len(idx)-n:] {
				keep[k] = true
			}
		case "kl":
			for _, k := range idx[:n] {
				keep[k] = true
			}
		case "dh":
			for _, k := range idx[:len(idx)-n] {
				keep[k] = true
			}
		case "dl":
			for _, k := range idx[n:] {
				keep[k] = true
			}
		}
	}
	for i, v := range vals {
		if keep[i] {
			sum += int64(v)
		}
	}
	return sum, vals, keep
}

func rollDice(r *rand.Rand, terms []diceTerm) diceResult {
	var res diceResult
	for _, t := range terms {
		if t.isConst {
			res.total += t.sign * t.konst
			res.detail = append(res.detail, fmt.Sprintf("%+d", t.sign*t.konst))
			continue
		}
		sum, vals, keep := rollTerm(r, t)
		res.total += t.sign * sum
		parts := make([]string, len(vals))
		for i, v := range vals {
			if keep[i] {
				parts[i] = strconv.Itoa(v)
				res.rolls = append(res.rolls, strconv.Itoa(v))
			} else {
				parts[i] = "(" + strconv.Itoa(v) + ")"
			}
		}
		sg := ""
		if t.sign < 0 {
			sg = "-"
		}
		res.detail = append(res.detail, fmt.Sprintf("%s%s [%s]", sg, t.canonical, strings.Join(parts, ", ")))
	}
	return res
}

func diceBounds(terms []diceTerm) (lo, hi int64, unbounded bool) {
	for _, t := range terms {
		if t.isConst {
			lo += t.sign * t.konst
			hi += t.sign * t.konst
			continue
		}
		k := int64(t.kept())
		tmin, tmax := k, k*int64(t.sides)
		if t.explode {
			unbounded = true
		}
		if t.sign > 0 {
			lo += tmin
			hi += tmax
		} else {
			lo -= tmax
			hi -= tmin
		}
	}
	return
}

func init() {
	reg(tools.Func{
		Name: "dice", Title: "Roll dice", Order: 7, Input: "spec", Keywords: []string{"dice", "roll", "d20", "2d6", "rpg"},
		Doc:     "Rolls dice in standard notation: NdM, sums of terms (2d6+3, 1d8+1d4-1), keep/drop (4d6kh3, 2d20kl1, 5d6dl1) and exploding dice (3d6!). Returns spec, total, the kept rolls and a detail line where dropped dice are in parentheses.",
		Params:  []tools.Param{p("spec", "str", "Dice notation such as d20, 2d6+3, 4d6kh3 or 3d6!. Max 1000 dice per term, 10000 in total.", "d6"), seedParam()},
		Returns: "dict[str, str]", Example: `rand.dice("4d6kh3")`,
		Tags: []string{"dice roller", "dnd", "rpg", "d20", "2d6", "keep highest", "advantage", "exploding", "tabletop"},
		Fn: func(a tools.Args) (any, error) {
			terms, err := parseDice(a.Str("spec"))
			if err != nil {
				return nil, err
			}
			res := rollDice(newRng(a.Int("seed")), terms)
			return map[string]string{
				"spec":   strings.ToLower(strings.Join(strings.Fields(a.Str("spec")), "")),
				"total":  strconv.FormatInt(res.total, 10),
				"rolls":  strings.Join(res.rolls, ", "),
				"detail": strings.Join(res.detail, " "),
			}, nil
		},
	})
	reg(tools.Func{
		Name: "dice_total", Title: "Roll dice (total only)", Order: 8, Input: "spec", Keywords: []string{"roll"},
		Doc:     "Rolls dice notation like dice() but returns only the integer total, handy for further arithmetic.",
		Params:  []tools.Param{p("spec", "str", "Dice notation such as 2d6+3.", "d6"), seedParam()},
		Returns: "int", Example: `rand.dice_total("2d6+3")`,
		Tags: []string{"roll", "sum", "dice total", "2d6"},
		Fn: func(a tools.Args) (any, error) {
			terms, err := parseDice(a.Str("spec"))
			if err != nil {
				return nil, err
			}
			return rollDice(newRng(a.Int("seed")), terms).total, nil
		},
	})
	reg(tools.Func{
		Name: "roll_stats", Title: "Dice statistics",
		Doc:     "Returns min, max and mean of a dice expression. Mean is exact for plain dice and estimated from 20000 deterministic trials when keep/drop or exploding dice are used; max is \"unbounded\" for exploding dice (capped at 100 re-rolls per die when rolling).",
		Params:  []tools.Param{req("spec", "str", "Dice notation such as 4d6kh3.")},
		Returns: "dict[str, str]", Example: `rand.roll_stats("4d6kh3")`,
		Tags: []string{"dice odds", "average", "expected value", "probability", "dnd"},
		Fn: func(a tools.Args) (any, error) {
			terms, err := parseDice(a.Str("spec"))
			if err != nil {
				return nil, err
			}
			lo, hi, unb := diceBounds(terms)
			exact := true
			mean := 0.0
			dice := 0
			for _, t := range terms {
				if t.isConst {
					mean += float64(t.sign * t.konst)
					continue
				}
				dice += t.count
				if t.mod != "" || t.explode {
					exact = false
				} else {
					mean += float64(t.sign) * float64(t.count) * float64(t.sides+1) / 2
				}
			}
			method := "exact"
			if !exact {
				method = fmt.Sprintf("simulation of %d trials", statsTrials)
				trials := statsTrials
				if dice*trials > statsDiceBudget {
					trials = max(1, statsDiceBudget/dice)
					method = fmt.Sprintf("simulation of %d trials", trials)
				}
				r := newRng(statsFixedSeed)
				var sum float64
				for i := 0; i < trials; i++ {
					sum += float64(rollDice(r, terms).total)
				}
				mean = sum / float64(trials)
			}
			maxS := strconv.FormatInt(hi, 10)
			if unb {
				maxS = "unbounded"
			}
			return map[string]string{
				"spec": a.Str("spec"), "min": strconv.FormatInt(lo, 10), "max": maxS,
				"mean": tools.Format(mean), "method": method,
			}, nil
		},
	})

	for _, sides := range []int{4, 6, 8, 10, 12, 20, 100} {
		name := "d" + strconv.Itoa(sides)
		order := 20
		if sides == 20 || sides == 100 || sides == 6 {
			order = 11
		}
		reg(tools.Func{
			Name: name, Title: "Roll a " + name, Order: order, Input: "args", Keywords: []string{name},
			Doc: fmt.Sprintf("Rolls one %d-sided die (1..%d) and adds an optional modifier. In args (launcher) use \"+5\" for a modifier and \"adv\" or \"dis\" to roll twice and keep the higher or lower result.", sides, sides),
			Params: []tools.Param{p("modifier", "int", "Added to the roll.", int64(0)), seedParam(),
				argsParam("'+3', 'adv', 'dis' or 'adv +2'.")},
			Returns: "int", Example: fmt.Sprintf("rand.%s()", name),
			Tags: []string{"die", "roll", "dnd", "tabletop", "advantage", "disadvantage", name},
			Fn: func(a tools.Args) (any, error) {
				mod := a.Int("modifier")
				adv, dis := 0, 0
				for _, f := range strings.Fields(strings.ToLower(a.Str("args"))) {
					switch {
					case strings.HasPrefix(f, "adv"):
						adv = 1
					case strings.HasPrefix(f, "dis"):
						dis = 1
					default:
						if n, err := strconv.ParseInt(strings.TrimPrefix(f, "+"), 10, 64); err == nil {
							mod = n
						}
					}
				}
				r := newRng(a.Int("seed"))
				v := int64(r.IntN(sides) + 1)
				if adv+dis == 1 {
					w := int64(r.IntN(sides) + 1)
					if adv == 1 {
						v = max(v, w)
					} else {
						v = min(v, w)
					}
				}
				return v + mod, nil
			},
		})
	}
}

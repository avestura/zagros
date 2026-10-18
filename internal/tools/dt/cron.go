package dt

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type cronField struct {
	set  map[int]bool
	star bool // field started with * (or ?)
	raw  string
}

type cronSpec struct {
	min, hour, dom, mon, dow cronField
}

var cronMacros = map[string]string{
	"@yearly": "0 0 1 1 *", "@annually": "0 0 1 1 *", "@monthly": "0 0 1 * *", "@weekly": "0 0 * * 0",
	"@daily": "0 0 * * *", "@midnight": "0 0 * * *", "@hourly": "0 * * * *",
}

var cronMonths = map[string]int{"jan": 1, "feb": 2, "mar": 3, "apr": 4, "may": 5, "jun": 6, "jul": 7, "aug": 8, "sep": 9, "oct": 10, "nov": 11, "dec": 12}
var cronDays = map[string]int{"sun": 0, "mon": 1, "tue": 2, "wed": 3, "thu": 4, "fri": 5, "sat": 6}

func cronValue(s string, names map[string]int) (int, error) {
	s = strings.ToLower(s)
	if names != nil {
		if v, ok := names[s]; ok {
			return v, nil
		}
		if len(s) > 3 {
			if v, ok := names[s[:3]]; ok {
				return v, nil
			}
		}
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("invalid value %q", s)
	}
	return v, nil
}

func parseCronField(raw string, lo, hi int, names map[string]int) (cronField, error) {
	f := cronField{set: map[int]bool{}, raw: raw}
	f.star = strings.HasPrefix(raw, "*") || strings.HasPrefix(raw, "?")
	for _, part := range strings.Split(raw, ",") {
		if part == "" {
			return f, fmt.Errorf("empty list item in %q", raw)
		}
		rng, stepS, hasStep := strings.Cut(part, "/")
		step := 1
		if hasStep {
			n, err := strconv.Atoi(stepS)
			if err != nil || n < 1 {
				return f, fmt.Errorf("invalid step in %q", part)
			}
			step = n
		}
		a, b := lo, hi
		switch {
		case rng == "*" || rng == "?":
		case strings.Contains(rng, "-"):
			x, y, _ := strings.Cut(rng, "-")
			var err error
			if a, err = cronValue(x, names); err != nil {
				return f, err
			}
			if b, err = cronValue(y, names); err != nil {
				return f, err
			}
		default:
			v, err := cronValue(rng, names)
			if err != nil {
				return f, err
			}
			a, b = v, v
			if hasStep {
				b = hi
			}
		}
		if a < lo || b > hi || a > b {
			return f, fmt.Errorf("value out of range %d-%d in %q", lo, hi, part)
		}
		for v := a; v <= b; v += step {
			f.set[v] = true
		}
	}
	return f, nil
}

func parseCron(expr string) (cronSpec, error) {
	var c cronSpec
	e := strings.TrimSpace(expr)
	if m, ok := cronMacros[strings.ToLower(e)]; ok {
		e = m
	}
	fs := strings.Fields(e)
	if len(fs) != 5 {
		return c, fmt.Errorf("a cron expression has 5 fields (minute hour day-of-month month day-of-week), got %d", len(fs))
	}
	var err error
	if c.min, err = parseCronField(fs[0], 0, 59, nil); err != nil {
		return c, fmt.Errorf("minute: %w", err)
	}
	if c.hour, err = parseCronField(fs[1], 0, 23, nil); err != nil {
		return c, fmt.Errorf("hour: %w", err)
	}
	if c.dom, err = parseCronField(fs[2], 1, 31, nil); err != nil {
		return c, fmt.Errorf("day of month: %w", err)
	}
	if c.mon, err = parseCronField(fs[3], 1, 12, cronMonths); err != nil {
		return c, fmt.Errorf("month: %w", err)
	}
	if c.dow, err = parseCronField(fs[4], 0, 7, cronDays); err != nil {
		return c, fmt.Errorf("day of week: %w", err)
	}
	if c.dow.set[7] {
		c.dow.set[0] = true
		delete(c.dow.set, 7)
	}
	return c, nil
}

func (c cronSpec) dayMatches(t time.Time) bool {
	domOK, dowOK := c.dom.set[t.Day()], c.dow.set[int(t.Weekday())]
	switch {
	case c.dom.star && c.dow.star:
		return true
	case c.dom.star:
		return dowOK
	case c.dow.star:
		return domOK
	}
	return domOK || dowOK
}

func sortedKeys(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}

func cronNext(expr string, count int, from time.Time) ([]time.Time, error) {
	if count < 1 || count > 1000 {
		return nil, fmt.Errorf("count must be 1..1000")
	}
	c, err := parseCron(expr)
	if err != nil {
		return nil, err
	}
	hours, mins := sortedKeys(c.hour.set), sortedKeys(c.min.set)
	start := from.Truncate(time.Minute).Add(time.Minute)
	loc := from.Location()
	day := time.Date(start.Year(), start.Month(), start.Day(), 12, 0, 0, 0, loc)
	var out []time.Time
	for i := 0; i < 366*9 && len(out) < count; i++ {
		d := day.AddDate(0, 0, i)
		if !c.mon.set[int(d.Month())] || !c.dayMatches(d) {
			continue
		}
		for _, h := range hours {
			for _, m := range mins {
				t := time.Date(d.Year(), d.Month(), d.Day(), h, m, 0, 0, loc)
				if t.Hour() != h || t.Minute() != m || t.Before(start) {
					continue // skipped by a DST gap, or already past
				}
				out = append(out, t)
				if len(out) >= count {
					return out, nil
				}
			}
		}
	}
	return out, nil
}

// ---- explanation ----

func ordinal(n int) string {
	suf := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suf = "st"
		case 2:
			suf = "nd"
		case 3:
			suf = "rd"
		}
	}
	return strconv.Itoa(n) + suf
}

func joinAnd(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	}
	return strings.Join(items[:len(items)-1], ", ") + " and " + items[len(items)-1]
}

// describeSet renders a sorted set as ranges and lists using name.
func describeSet(vals []int, name func(int) string) string {
	var parts []string
	for i := 0; i < len(vals); {
		j := i
		for j+1 < len(vals) && vals[j+1] == vals[j]+1 {
			j++
		}
		if j-i >= 2 {
			parts = append(parts, name(vals[i])+" through "+name(vals[j]))
		} else {
			for k := i; k <= j; k++ {
				parts = append(parts, name(vals[k]))
			}
		}
		i = j + 1
	}
	return joinAnd(parts)
}

func fullSet(f cronField, lo, hi int) bool {
	for v := lo; v <= hi; v++ {
		if !f.set[v] {
			return false
		}
	}
	return true
}

func stepOf(vals []int) int {
	if len(vals) < 3 {
		return 0
	}
	d := vals[1] - vals[0]
	for i := 2; i < len(vals); i++ {
		if vals[i]-vals[i-1] != d {
			return 0
		}
	}
	return d
}

func cronExplain(expr string) (string, error) {
	c, err := parseCron(expr)
	if err != nil {
		return "", err
	}
	mins, hours := sortedKeys(c.min.set), sortedKeys(c.hour.set)
	var parts []string
	allMin, allHour := fullSet(c.min, 0, 59), fullSet(c.hour, 0, 23)
	switch {
	case allMin && allHour:
		parts = append(parts, "Every minute")
	case len(mins)*len(hours) <= 4 && !allMin && !allHour:
		var ts []string
		for _, h := range hours {
			for _, m := range mins {
				ts = append(ts, fmt.Sprintf("%02d:%02d", h, m))
			}
		}
		parts = append(parts, "At "+joinAnd(ts))
	default:
		var mp string
		switch {
		case allMin:
			mp = "Every minute"
		case len(mins) == 1:
			mp = fmt.Sprintf("At minute %d", mins[0])
		case stepOf(mins) > 0 && mins[0] == 0 && c.min.raw != "":
			mp = fmt.Sprintf("Every %d minutes", stepOf(mins))
		default:
			mp = "At minute " + describeSet(mins, strconv.Itoa)
		}
		var hp string
		switch {
		case allHour:
			hp = "past every hour"
		case len(hours) > 1 && hours[len(hours)-1]-hours[0] == len(hours)-1:
			hp = fmt.Sprintf("between %02d:00 and %02d:59", hours[0], hours[len(hours)-1])
		case stepOf(hours) > 0:
			hp = fmt.Sprintf("every %d hours from %02d:00", stepOf(hours), hours[0])
		default:
			hp = "past hour " + describeSet(hours, strconv.Itoa)
		}
		if allHour && !allMin && len(mins) == 1 {
			parts = append(parts, mp+" past every hour")
		} else if allHour {
			parts = append(parts, mp)
		} else {
			parts = append(parts, mp+", "+hp)
		}
	}
	dayName := func(i int) string { return time.Weekday(i).String() }
	monName := func(i int) string { return time.Month(i).String() }
	var dayParts []string
	if !(c.dom.star && fullSet(c.dom, 1, 31)) {
		dayParts = append(dayParts, "on day-of-month "+describeSet(sortedKeys(c.dom.set), strconv.Itoa))
	}
	if !(c.dow.star && fullSet(c.dow, 0, 6)) {
		dayParts = append(dayParts, "on "+describeSet(sortedKeys(c.dow.set), dayName))
	}
	if len(dayParts) == 2 {
		parts = append(parts, strings.Join(dayParts, " and "))
	} else {
		parts = append(parts, dayParts...)
	}
	if !fullSet(c.mon, 1, 12) {
		parts = append(parts, "in "+describeSet(sortedKeys(c.mon.set), monName))
	}
	return strings.Join(parts, ", "), nil
}

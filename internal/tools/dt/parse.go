package dt

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ---- calendar helpers ----

func daysIn(y int, m time.Month) int {
	return time.Date(y, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func isLeap(y int) bool { return y%4 == 0 && (y%100 != 0 || y%400 == 0) }

// addMonths moves t by n calendar months and clamps the day to the target
// month (Jan 31 + 1 month = Feb 28/29).
func addMonths(t time.Time, n int) time.Time {
	y, m, d := t.Date()
	total := y*12 + int(m) - 1 + n
	ny, nm := total/12, total%12
	if nm < 0 {
		nm += 12
		ny--
	}
	mm := time.Month(nm + 1)
	if dm := daysIn(ny, mm); d > dm {
		d = dm
	}
	return time.Date(ny, mm, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}

type delta struct {
	years, months, days int
	dur                 time.Duration
}

func (d delta) apply(t time.Time, sign int) time.Time {
	if n := d.years*12 + d.months; n != 0 {
		t = addMonths(t, sign*n)
	}
	if d.days != 0 {
		t = t.AddDate(0, 0, sign*d.days)
	}
	return t.Add(time.Duration(sign) * d.dur)
}

func (d delta) neg() delta { return delta{-d.years, -d.months, -d.days, -d.dur} }

// seconds converts with 1 year = 365 days and 1 month = 30 days.
func (d delta) seconds() float64 {
	return float64(d.years)*365*86400 + float64(d.months)*30*86400 + float64(d.days)*86400 + d.dur.Seconds()
}

var unitKinds = map[string]string{
	"y": "y", "yr": "y", "yrs": "y", "year": "y", "years": "y",
	"mo": "mo", "mon": "mo", "mos": "mo", "month": "mo", "months": "mo",
	"q": "q", "quarter": "q", "quarters": "q",
	"w": "w", "wk": "w", "wks": "w", "week": "w", "weeks": "w",
	"fortnight": "f", "fortnights": "f",
	"d": "d", "day": "d", "days": "d",
	"h": "h", "hr": "h", "hrs": "h", "hour": "h", "hours": "h",
	"m": "min", "min": "min", "mins": "min", "minute": "min", "minutes": "min",
	"s": "s", "sec": "s", "secs": "s", "second": "s", "seconds": "s",
	"ms": "ms", "msec": "ms", "msecs": "ms", "millisecond": "ms", "milliseconds": "ms",
	"us": "us", "µs": "us", "usec": "us", "microsecond": "us", "microseconds": "us",
	"ns": "ns", "nsec": "ns", "nanosecond": "ns", "nanoseconds": "ns",
}

var numberWords = map[string]float64{
	"a": 1, "an": 1, "one": 1, "two": 2, "three": 3, "four": 4, "five": 5, "six": 6, "seven": 7,
	"eight": 8, "nine": 9, "ten": 10, "eleven": 11, "twelve": 12, "half": 0.5,
}

var specRe = regexp.MustCompile(`(?i)([+\-]?)\s*(\d+(?:[.,]\d+)?|\b(?:an?|one|two|three|four|five|six|seven|eight|nine|ten|eleven|twelve|half)\b)\s*([a-zµ]+)`)
var specGapRe = regexp.MustCompile(`(?i)^[\s,;]*(?:and)?[\s,;]*$`)

// parseSpec reads "1y2mo3w4d5h6m7s", "-1d", "1d-2h", "2 hours and 30 minutes".
func parseSpec(s string) (delta, error) {
	var d delta
	s = strings.TrimSpace(s)
	if s == "" {
		return d, fmt.Errorf("empty duration spec; try \"1d2h\" or \"-3 weeks\"")
	}
	locs := specRe.FindAllStringSubmatchIndex(s, -1)
	if len(locs) == 0 {
		return d, fmt.Errorf("cannot read %q as a duration spec; try \"1y2mo3w4d5h6m7s\"", s)
	}
	prev := 0
	for _, m := range locs {
		if !specGapRe.MatchString(s[prev:m[0]]) {
			return d, fmt.Errorf("cannot read %q in duration spec %q", strings.TrimSpace(s[prev:m[0]]), s)
		}
		prev = m[1]
		sign := 1.0
		if s[m[2]:m[3]] == "-" {
			sign = -1
		}
		numTxt := strings.ToLower(s[m[4]:m[5]])
		var n float64
		if w, ok := numberWords[numTxt]; ok {
			n = w
		} else {
			n, _ = strconv.ParseFloat(strings.Replace(numTxt, ",", ".", 1), 64)
		}
		n *= sign
		unitTxt := strings.ToLower(s[m[6]:m[7]])
		kind, ok := unitKinds[unitTxt]
		if !ok {
			return d, fmt.Errorf("unknown time unit %q", unitTxt)
		}
		if math.Abs(n) > 1e9 {
			return d, fmt.Errorf("number %v is too large", n)
		}
		whole, frac := math.Modf(n)
		w := int(whole)
		switch kind {
		case "y":
			d.years += w
			d.months += int(math.Round(frac * 12))
		case "mo":
			d.months += w
			d.dur += time.Duration(frac * 30 * 24 * float64(time.Hour))
		case "q":
			d.months += 3 * w
			d.months += int(math.Round(frac * 3))
		case "w":
			d.days += 7 * w
			d.dur += time.Duration(frac * 7 * 24 * float64(time.Hour))
		case "f":
			d.days += 14 * w
			d.dur += time.Duration(frac * 14 * 24 * float64(time.Hour))
		case "d":
			d.days += w
			d.dur += time.Duration(frac * 24 * float64(time.Hour))
		case "h":
			d.dur += time.Duration(n * float64(time.Hour))
		case "min":
			d.dur += time.Duration(n * float64(time.Minute))
		case "s":
			d.dur += time.Duration(n * float64(time.Second))
		case "ms":
			d.dur += time.Duration(n * float64(time.Millisecond))
		case "us":
			d.dur += time.Duration(n * float64(time.Microsecond))
		case "ns":
			d.dur += time.Duration(n)
		}
	}
	if !specGapRe.MatchString(s[prev:]) {
		return d, fmt.Errorf("cannot read %q in duration spec %q", strings.TrimSpace(s[prev:]), s)
	}
	return d, nil
}

// ---- weekdays and months ----

var weekdayNames = map[string]time.Weekday{
	"sunday": time.Sunday, "sun": time.Sunday, "monday": time.Monday, "mon": time.Monday,
	"tuesday": time.Tuesday, "tue": time.Tuesday, "tues": time.Tuesday, "wednesday": time.Wednesday, "wed": time.Wednesday,
	"thursday": time.Thursday, "thu": time.Thursday, "thur": time.Thursday, "thurs": time.Thursday,
	"friday": time.Friday, "fri": time.Friday, "saturday": time.Saturday, "sat": time.Saturday,
}

func parseWeekday(s string) (time.Weekday, bool) {
	w, ok := weekdayNames[strings.ToLower(strings.TrimSpace(s))]
	return w, ok
}

func parseMonthName(s string) (time.Month, bool) {
	s = strings.ToLower(strings.Trim(s, ".,"))
	if len(s) < 3 {
		return 0, false
	}
	if s == "sept" {
		return time.September, true
	}
	for m := time.January; m <= time.December; m++ {
		full := strings.ToLower(m.String())
		if s == full || (len(s) == 3 && s == full[:3]) {
			return m, true
		}
	}
	return 0, false
}

// ---- time of day ----

var todRe = regexp.MustCompile(`(?i)^(\d{1,2})(?::(\d{2}))?(?::(\d{2})(?:[.,](\d{1,9}))?)?\s*(am|pm|a\.m\.|p\.m\.)?$`)

type tod struct{ h, m, s, ns int }

func parseTOD(s string) (tod, bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	switch s {
	case "noon", "midday":
		return tod{12, 0, 0, 0}, true
	case "midnight":
		return tod{}, true
	}
	m := todRe.FindStringSubmatch(s)
	if m == nil {
		return tod{}, false
	}
	if m[2] == "" && m[5] == "" {
		return tod{}, false // a bare number is not a time
	}
	h, _ := strconv.Atoi(m[1])
	mi, _ := strconv.Atoi(m[2])
	se, _ := strconv.Atoi(m[3])
	ns := 0
	if m[4] != "" {
		f := (m[4] + "000000000")[:9]
		ns, _ = strconv.Atoi(f)
	}
	if ap := strings.ReplaceAll(m[5], ".", ""); ap != "" {
		if h < 1 || h > 12 {
			return tod{}, false
		}
		if ap == "am" && h == 12 {
			h = 0
		} else if ap == "pm" && h < 12 {
			h += 12
		}
	}
	if h > 23 || mi > 59 || se > 59 {
		return tod{}, false
	}
	return tod{h, mi, se, ns}, true
}

func atTOD(t time.Time, v tod) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, v.h, v.m, v.s, v.ns, t.Location())
}

func midnight(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

// ---- parse entry point ----

type parseOpts struct {
	loc      *time.Location
	now      time.Time // reference instant, already in loc
	dayFirst bool
}

var (
	reTsep    = regexp.MustCompile(`(\d)[tT](\d)`)
	reOrdinal = regexp.MustCompile(`(?i)(\d)(st|nd|rd|th)\b`)
	reDotAP   = regexp.MustCompile(`(?i)\b([ap])\.m\.`)
	reFill    = regexp.MustCompile(`(?i)\b(at|on|the|of)\b`)
	reTime    = regexp.MustCompile(`(?i)(\d{1,2}):(\d{2})(?::(\d{2})(?:[.,](\d{1,9}))?)?(?:\s*(am|pm)\b)?`)
	reHourAP  = regexp.MustCompile(`(?i)\b(\d{1,2})\s*(am|pm)\b`)
	reNum     = regexp.MustCompile(`^[+-]?\d+(?:\.\d+)?$`)
	reHex     = regexp.MustCompile(`(?i)^0x[0-9a-f]+$`)
	reYMD     = regexp.MustCompile(`^(\d{4})[-/.](\d{1,2})[-/.](\d{1,2})$`)
	reDMY     = regexp.MustCompile(`^(\d{1,2})([-/.])(\d{1,2})[-/.](\d{2}|\d{4})$`)
	reCompact = regexp.MustCompile(`^(\d{4})(\d{2})(\d{2})$`)
	reWeek    = regexp.MustCompile(`(?i)^(\d{4})-?w(\d{1,2})(?:-?([1-7]))?$`)
	reOrdDate = regexp.MustCompile(`^(\d{4})-(\d{3})$`)
	reYM      = regexp.MustCompile(`^(\d{4})[-/](\d{1,2})$`)
)

// parseAny reads the many date and time notations Zagros accepts.
func parseAny(text string, o parseOpts) (time.Time, error) {
	s := strings.TrimSpace(text)
	if s == "" {
		return time.Time{}, fmt.Errorf("empty date/time text")
	}
	if t, ok, err := parseRelative(s, o); ok || err != nil {
		return t, err
	}
	if reHex.MatchString(s) {
		n, err := strconv.ParseInt(s[2:], 16, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("hex timestamp %q is too large", s)
		}
		return time.Unix(n, 0).In(o.loc), nil
	}
	if reNum.MatchString(s) {
		return parseNumeric(s, o)
	}
	return parseGeneral(s, o)
}

func parseNumeric(s string, o parseOpts) (time.Time, error) {
	digits := strings.TrimLeft(s, "+-")
	if !strings.Contains(s, ".") && s[0] != '-' && s[0] != '+' {
		switch len(digits) {
		case 4:
			y, _ := strconv.Atoi(digits)
			return time.Date(y, 1, 1, 0, 0, 0, 0, o.loc), nil
		case 8:
			y, _ := strconv.Atoi(digits[:4])
			m, _ := strconv.Atoi(digits[4:6])
			d, _ := strconv.Atoi(digits[6:])
			if y >= 1000 && y <= 2999 && m >= 1 && m <= 12 && d >= 1 && d <= 31 {
				return buildDate(y, time.Month(m), d, tod{}, o.loc)
			}
		case 14:
			y, _ := strconv.Atoi(digits[:4])
			m, _ := strconv.Atoi(digits[4:6])
			d, _ := strconv.Atoi(digits[6:8])
			h, _ := strconv.Atoi(digits[8:10])
			mi, _ := strconv.Atoi(digits[10:12])
			se, _ := strconv.Atoi(digits[12:])
			if y >= 1000 && y <= 2999 && m >= 1 && m <= 12 && d >= 1 && d <= 31 && h < 24 && mi < 60 && se < 60 {
				return buildDate(y, time.Month(m), d, tod{h, mi, se, 0}, o.loc)
			}
		}
	}
	t, _, err := fromUnixAuto(s, "auto", o.loc)
	return t, err
}

// unitScale returns nanoseconds per unit.
func unitScale(unit string) (float64, error) {
	switch strings.ToLower(unit) {
	case "s", "sec", "seconds", "":
		return 1e9, nil
	case "ms", "milliseconds":
		return 1e6, nil
	case "us", "µs", "microseconds":
		return 1e3, nil
	case "ns", "nanoseconds":
		return 1, nil
	}
	return 0, fmt.Errorf("unit must be auto, s, ms, us or ns")
}

// fromUnixAuto converts a Unix number (integer or decimal) in the given unit,
// guessing the unit by magnitude for "auto". It returns the unit it used.
func fromUnixAuto(s, unit string, loc *time.Location) (time.Time, string, error) {
	s = strings.TrimSpace(s)
	if reHex.MatchString(s) {
		n, err := strconv.ParseInt(s[2:], 16, 64)
		if err != nil {
			return time.Time{}, "", fmt.Errorf("hex timestamp %q is too large", s)
		}
		return time.Unix(n, 0).In(loc), "s", nil
	}
	if !reNum.MatchString(s) {
		return time.Time{}, "", fmt.Errorf("%q is not a number", s)
	}
	intPart, fracPart, _ := strings.Cut(strings.TrimPrefix(s, "+"), ".")
	neg := strings.HasPrefix(intPart, "-")
	n, err := strconv.ParseInt(intPart, 10, 64)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("timestamp %q is out of range", s)
	}
	u := strings.ToLower(unit)
	if u == "auto" || u == "" {
		a := n
		if a < 0 {
			a = -a
		}
		switch {
		case a < 100_000_000_000:
			u = "s"
		case a < 100_000_000_000_000:
			u = "ms"
		case a < 100_000_000_000_000_000:
			u = "us"
		default:
			u = "ns"
		}
	}
	scale, err := unitScale(u)
	if err != nil {
		return time.Time{}, "", err
	}
	perSec := int64(1e9 / scale)
	sec := n / perSec
	rem := n % perSec // keeps the sign of n
	ns := rem * int64(scale)
	if fracPart != "" {
		f := (fracPart + "000000000000")[:12]
		fv, _ := strconv.ParseInt(f, 10, 64)
		extra := int64(float64(fv) / 1e12 * scale)
		if neg {
			ns -= extra
		} else {
			ns += extra
		}
	}
	return time.Unix(sec, ns).In(loc), u, nil
}

func buildDate(y int, m time.Month, d int, v tod, loc *time.Location) (time.Time, error) {
	if m < 1 || m > 12 {
		return time.Time{}, fmt.Errorf("month %d is out of range", int(m))
	}
	if d < 1 || d > daysIn(y, m) {
		return time.Time{}, fmt.Errorf("%s %d, %d does not exist", m, d, y)
	}
	return time.Date(y, m, d, v.h, v.m, v.s, v.ns, loc), nil
}

func zoneToken(tok string) (*time.Location, bool) {
	tok = strings.TrimSpace(tok)
	if tok == "" {
		return nil, false
	}
	if off, ok := parseOffset(tok); ok {
		return time.FixedZone(offsetName(off), off), true
	}
	if a, ok := zoneAbbrs[strings.ToUpper(tok)]; ok {
		if strings.EqualFold(tok, "z") || strings.EqualFold(tok, "utc") || strings.EqualFold(tok, "gmt") {
			return time.UTC, true
		}
		return time.FixedZone(strings.ToUpper(tok), a.off), true
	}
	if strings.Contains(tok, "/") {
		if c, ok := canonicalZone(tok); ok {
			if l, err := time.LoadLocation(c); err == nil {
				return l, true
			}
		}
	}
	return nil, false
}

func parseGeneral(orig string, o parseOpts) (time.Time, error) {
	s := reTsep.ReplaceAllString(orig, "$1 $2")
	s = strings.NewReplacer(",", " ", "(", " ", ")", " ").Replace(s)
	s = reOrdinal.ReplaceAllString(s, "$1")
	s = reDotAP.ReplaceAllString(s, "${1}m")
	s = reFill.ReplaceAllString(s, " ")
	s = strings.Join(strings.Fields(s), " ")

	zone := o.loc
	var v tod
	hasTime := false
	var rest string

	take := func(start, end int) {
		before, after := strings.TrimSpace(s[:start]), strings.TrimSpace(s[end:])
		toks := strings.Fields(after)
		i := 0
		var numeric, named *time.Location
		for ; i < len(toks); i++ {
			l, ok := zoneToken(toks[i])
			if !ok {
				break
			}
			if _, isOff := parseOffset(toks[i]); isOff {
				numeric = l
			} else if named == nil {
				named = l
			}
		}
		switch {
		case numeric != nil:
			zone = numeric
		case named != nil:
			zone = named
		}
		rest = strings.TrimSpace(before + " " + strings.Join(toks[i:], " "))
		hasTime = true
	}

	if m := reTime.FindStringSubmatchIndex(s); m != nil {
		h, _ := strconv.Atoi(s[m[2]:m[3]])
		mi, _ := strconv.Atoi(s[m[4]:m[5]])
		se, ns := 0, 0
		if m[6] >= 0 {
			se, _ = strconv.Atoi(s[m[6]:m[7]])
		}
		if m[8] >= 0 {
			ns, _ = strconv.Atoi((s[m[8]:m[9]] + "000000000")[:9])
		}
		if m[10] >= 0 {
			ap := strings.ToLower(strings.ReplaceAll(s[m[10]:m[11]], ".", ""))
			if h < 1 || h > 12 {
				return time.Time{}, fmt.Errorf("hour %d is not valid with am/pm", h)
			}
			if ap == "am" && h == 12 {
				h = 0
			} else if ap == "pm" && h < 12 {
				h += 12
			}
		}
		if h > 23 || mi > 59 || se > 59 {
			return time.Time{}, fmt.Errorf("time %q is out of range", s[m[0]:m[1]])
		}
		v = tod{h, mi, se, ns}
		take(m[0], m[1])
	} else if m := reHourAP.FindStringSubmatchIndex(s); m != nil {
		h, _ := strconv.Atoi(s[m[2]:m[3]])
		ap := strings.ToLower(s[m[4]:m[5]])
		if h < 1 || h > 12 {
			return time.Time{}, fmt.Errorf("hour %d is not valid with am/pm", h)
		}
		if ap == "am" && h == 12 {
			h = 0
		} else if ap == "pm" && h < 12 {
			h += 12
		}
		v = tod{h, 0, 0, 0}
		take(m[0], m[1])
	} else {
		rest = s
	}

	toks := strings.Fields(rest)
	// drop weekday names and a trailing zone when there was no time
	var kept []string
	for _, t := range toks {
		if _, ok := parseWeekday(t); ok && len(toks) > 1 {
			continue
		}
		kept = append(kept, t)
	}
	toks = kept
	if !hasTime && len(toks) > 1 {
		if l, ok := zoneToken(toks[len(toks)-1]); ok {
			if _, isAbbr := zoneAbbrs[strings.ToUpper(toks[len(toks)-1])]; isAbbr || strings.Contains(toks[len(toks)-1], "/") {
				zone = l
				toks = toks[:len(toks)-1]
			}
		}
	}
	if !hasTime && len(toks) == 0 {
		return time.Time{}, fmt.Errorf("cannot read %q as a date or time", orig)
	}
	if len(toks) == 0 {
		n := o.now.In(zone)
		return atTOD(n, v), nil
	}
	ds := strings.Join(toks, " ")
	now := o.now.In(zone)
	if m := reYMD.FindStringSubmatch(ds); m != nil {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		d, _ := strconv.Atoi(m[3])
		return buildDate(y, time.Month(mo), d, v, zone)
	}
	if m := reDMY.FindStringSubmatch(ds); m != nil {
		a, _ := strconv.Atoi(m[1])
		b, _ := strconv.Atoi(m[3])
		y, _ := strconv.Atoi(m[4])
		if len(m[4]) == 2 {
			if y >= 69 {
				y += 1900
			} else {
				y += 2000
			}
		}
		dayFirst := o.dayFirst || m[2] == "."
		var d, mo int
		switch {
		case a > 12:
			d, mo = a, b
		case b > 12:
			mo, d = a, b
		case dayFirst:
			d, mo = a, b
		default:
			mo, d = a, b
		}
		return buildDate(y, time.Month(mo), d, v, zone)
	}
	if m := reCompact.FindStringSubmatch(ds); m != nil {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		d, _ := strconv.Atoi(m[3])
		return buildDate(y, time.Month(mo), d, v, zone)
	}
	if m := reWeek.FindStringSubmatch(ds); m != nil {
		y, _ := strconv.Atoi(m[1])
		w, _ := strconv.Atoi(m[2])
		wd := 1
		if m[3] != "" {
			wd, _ = strconv.Atoi(m[3])
		}
		if w < 1 || w > 53 {
			return time.Time{}, fmt.Errorf("ISO week %d is out of range", w)
		}
		d := isoWeekStart(y, w, zone).AddDate(0, 0, wd-1)
		return atTOD(d, v), nil
	}
	if m := reOrdDate.FindStringSubmatch(ds); m != nil {
		y, _ := strconv.Atoi(m[1])
		n, _ := strconv.Atoi(m[2])
		if n < 1 || n > 365+b2i(isLeap(y)) {
			return time.Time{}, fmt.Errorf("day of year %d is out of range for %d", n, y)
		}
		return atTOD(time.Date(y, 1, n, 0, 0, 0, 0, zone), v), nil
	}
	if m := reYM.FindStringSubmatch(ds); m != nil {
		y, _ := strconv.Atoi(m[1])
		mo, _ := strconv.Atoi(m[2])
		return buildDate(y, time.Month(mo), 1, v, zone)
	}
	if len(toks) == 1 && len(toks[0]) == 4 {
		if y, err := strconv.Atoi(toks[0]); err == nil {
			return time.Date(y, 1, 1, v.h, v.m, v.s, v.ns, zone), nil
		}
	}
	// text months: "Jan 2 2006", "2 January 2006", "31-Jan-2024"
	words := strings.FieldsFunc(ds, func(r rune) bool { return r == ' ' || r == '-' || r == '/' || r == '.' })
	var month time.Month
	var nums []string
	for _, w := range words {
		if m, ok := parseMonthName(w); ok && month == 0 {
			month = m
			continue
		}
		if _, err := strconv.Atoi(w); err == nil {
			nums = append(nums, w)
			continue
		}
		return time.Time{}, fmt.Errorf("cannot read %q as a date or time", orig)
	}
	if month == 0 || len(nums) > 2 {
		return time.Time{}, fmt.Errorf("cannot read %q as a date or time", orig)
	}
	day, year := 0, 0
	for _, n := range nums {
		x, _ := strconv.Atoi(n)
		switch {
		case len(n) >= 3 || x > 31:
			if year != 0 {
				return time.Time{}, fmt.Errorf("cannot read %q as a date or time", orig)
			}
			year = x
		case day == 0:
			day = x
		default:
			year = x
			if len(n) == 2 {
				if x >= 69 {
					year += 1900
				} else {
					year += 2000
				}
			}
		}
	}
	if day == 0 {
		day = 1
	}
	if year == 0 {
		year = now.Year()
	}
	return buildDate(year, month, day, v, zone)
}

func b2i(b bool) int {
	if b {
		return 1
	}
	return 0
}

// isoWeekStart returns the Monday of ISO week w of year y.
func isoWeekStart(y, w int, loc *time.Location) time.Time {
	jan4 := time.Date(y, 1, 4, 0, 0, 0, 0, loc)
	wd := int(jan4.Weekday()+6) % 7 // Monday = 0
	return jan4.AddDate(0, 0, -wd+(w-1)*7)
}

// ---- relative expressions ----

var (
	reRelNow   = regexp.MustCompile(`^(now|today|tomorrow|yesterday|midnight|noon|tonight|day after tomorrow|day before yesterday)(?:\s+(.+))?$`)
	reRelIn    = regexp.MustCompile(`^in\s+(.+)$`)
	reRelAgo   = regexp.MustCompile(`^(.+?)\s+(ago|from now|later|hence|before now|after now)$`)
	reRelNext  = regexp.MustCompile(`^(next|last|previous|prev|this|coming|upcoming)\s+(\S+)(?:\s+(.+))?$`)
	reRelWday  = regexp.MustCompile(`^([a-z]+)(?:\s+(.+))?$`)
	reRelStart = regexp.MustCompile(`^(start|beginning|end)\s+(?:of\s+)?(this\s+|next\s+|last\s+)?(second|minute|hour|day|week|month|quarter|year)$`)
)

func parseRelative(text string, o parseOpts) (time.Time, bool, error) {
	l := strings.ToLower(strings.TrimSpace(text))
	l = strings.Join(strings.Fields(strings.NewReplacer(",", " ").Replace(l)), " ")
	l = reFill.ReplaceAllString(l, " ")
	l = strings.Join(strings.Fields(l), " ")
	now := o.now
	if m := reRelNow.FindStringSubmatch(l); m != nil {
		var base time.Time
		switch m[1] {
		case "now":
			if m[2] != "" {
				return time.Time{}, false, nil
			}
			return now, true, nil
		case "today":
			base = midnight(now)
		case "tomorrow":
			base = midnight(now).AddDate(0, 0, 1)
		case "yesterday":
			base = midnight(now).AddDate(0, 0, -1)
		case "day after tomorrow":
			base = midnight(now).AddDate(0, 0, 2)
		case "day before yesterday":
			base = midnight(now).AddDate(0, 0, -2)
		case "midnight":
			base = midnight(now)
		case "noon":
			base = atTOD(now, tod{12, 0, 0, 0})
		case "tonight":
			base = atTOD(now, tod{20, 0, 0, 0})
		}
		if m[2] == "" {
			return base, true, nil
		}
		v, ok := parseTOD(m[2])
		if !ok {
			return time.Time{}, false, nil
		}
		return atTOD(base, v), true, nil
	}
	if m := reRelStart.FindStringSubmatch(l); m != nil {
		base := now
		switch strings.TrimSpace(m[2]) {
		case "next":
			base = shiftUnit(now, m[3], 1)
		case "last":
			base = shiftUnit(now, m[3], -1)
		}
		if m[1] == "end" {
			return endOf(base, m[3], time.Monday), true, nil
		}
		return startOf(base, m[3], time.Monday), true, nil
	}
	if m := reRelIn.FindStringSubmatch(l); m != nil {
		if d, err := parseSpec(m[1]); err == nil {
			return d.apply(now, 1), true, nil
		}
	}
	if m := reRelAgo.FindStringSubmatch(l); m != nil {
		if d, err := parseSpec(m[1]); err == nil {
			if m[2] == "ago" || m[2] == "before now" {
				return d.apply(now, -1), true, nil
			}
			return d.apply(now, 1), true, nil
		}
	}
	if m := reRelNext.FindStringSubmatch(l); m != nil {
		dir := map[string]int{"next": 1, "coming": 1, "upcoming": 1, "last": -1, "previous": -1, "prev": -1, "this": 0}[m[1]]
		if wd, ok := parseWeekday(m[2]); ok {
			base := midnight(now)
			diff := (int(wd) - int(base.Weekday()) + 7) % 7
			switch dir {
			case 1:
				if diff == 0 {
					diff = 7
				}
			case -1:
				diff -= 7
				if diff == 0 {
					diff = -7
				}
			}
			d := base.AddDate(0, 0, diff)
			if m[3] != "" {
				v, ok := parseTOD(m[3])
				if !ok {
					return time.Time{}, false, nil
				}
				d = atTOD(d, v)
			}
			return d, true, nil
		}
		if k, ok := unitKinds[m[2]]; ok && m[3] == "" {
			switch k {
			case "w", "mo", "y", "d", "q", "h", "min":
				unit := map[string]string{"w": "week", "mo": "month", "y": "year", "d": "day", "q": "quarter", "h": "hour", "min": "minute"}[k]
				return shiftUnit(now, unit, dir), true, nil
			}
		}
	}
	if m := reRelWday.FindStringSubmatch(l); m != nil {
		if wd, ok := parseWeekday(m[1]); ok {
			base := midnight(now)
			d := base.AddDate(0, 0, (int(wd)-int(base.Weekday())+7)%7)
			if m[2] != "" {
				v, ok := parseTOD(m[2])
				if !ok {
					return time.Time{}, false, nil
				}
				d = atTOD(d, v)
			}
			return d, true, nil
		}
	}
	return time.Time{}, false, nil
}

func shiftUnit(t time.Time, unit string, n int) time.Time {
	switch unit {
	case "second":
		return t.Add(time.Duration(n) * time.Second)
	case "minute":
		return t.Add(time.Duration(n) * time.Minute)
	case "hour":
		return t.Add(time.Duration(n) * time.Hour)
	case "day":
		return t.AddDate(0, 0, n)
	case "week":
		return t.AddDate(0, 0, 7*n)
	case "month":
		return addMonths(t, n)
	case "quarter":
		return addMonths(t, 3*n)
	case "year":
		return addMonths(t, 12*n)
	}
	return t
}

// startOf truncates t to the beginning of unit (weeks begin on weekStart).
func startOf(t time.Time, unit string, weekStart time.Weekday) time.Time {
	y, m, d := t.Date()
	loc := t.Location()
	switch unit {
	case "second":
		return t.Truncate(time.Second)
	case "minute":
		return time.Date(y, m, d, t.Hour(), t.Minute(), 0, 0, loc)
	case "hour":
		return time.Date(y, m, d, t.Hour(), 0, 0, 0, loc)
	case "day":
		return time.Date(y, m, d, 0, 0, 0, 0, loc)
	case "week":
		diff := (int(t.Weekday()) - int(weekStart) + 7) % 7
		return time.Date(y, m, d-diff, 0, 0, 0, 0, loc)
	case "month":
		return time.Date(y, m, 1, 0, 0, 0, 0, loc)
	case "quarter":
		return time.Date(y, time.Month((int(m)-1)/3*3+1), 1, 0, 0, 0, 0, loc)
	case "year":
		return time.Date(y, 1, 1, 0, 0, 0, 0, loc)
	}
	return t
}

// endOf is the last nanosecond of unit.
func endOf(t time.Time, unit string, weekStart time.Weekday) time.Time {
	s := startOf(t, unit, weekStart)
	var next time.Time
	switch unit {
	case "second":
		next = s.Add(time.Second)
	case "minute":
		next = s.Add(time.Minute)
	case "hour":
		next = s.Add(time.Hour)
	default:
		next = shiftUnit(s, unit, 1)
	}
	return next.Add(-time.Nanosecond)
}

var unitNames = map[string]string{
	"second": "second", "sec": "second", "s": "second", "minute": "minute", "min": "minute", "hour": "hour", "h": "hour",
	"day": "day", "d": "day", "week": "week", "w": "week", "month": "month", "quarter": "quarter", "q": "quarter", "year": "year", "y": "year",
}

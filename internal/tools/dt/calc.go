package dt

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ---- durations ----

var isoDurRe = regexp.MustCompile(`(?i)^([+-])?P(?:(\d+(?:[.,]\d+)?)Y)?(?:(\d+(?:[.,]\d+)?)M)?(?:(\d+(?:[.,]\d+)?)W)?(?:(\d+(?:[.,]\d+)?)D)?(?:T(?:(\d+(?:[.,]\d+)?)H)?(?:(\d+(?:[.,]\d+)?)M)?(?:(\d+(?:[.,]\d+)?)S)?)?$`)

// isoDurationSeconds reads PnYnMnWnDTnHnMnS (year = 365 days, month = 30 days).
func isoDurationSeconds(s string) (float64, error) {
	s = strings.TrimSpace(s)
	m := isoDurRe.FindStringSubmatch(s)
	if m == nil || strings.EqualFold(strings.TrimLeft(s, "+-"), "P") || strings.HasSuffix(strings.ToUpper(s), "T") {
		return 0, fmt.Errorf("%q is not an ISO 8601 duration (example: PT1H30M, P1DT12H)", s)
	}
	mult := []float64{365 * 86400, 30 * 86400, 7 * 86400, 86400, 3600, 60, 1}
	total := 0.0
	for i, g := range m[2:] {
		if g == "" {
			continue
		}
		v, _ := strconv.ParseFloat(strings.Replace(g, ",", ".", 1), 64)
		total += v * mult[i]
	}
	if m[1] == "-" {
		total = -total
	}
	return total, nil
}

var clockRe = regexp.MustCompile(`^([+-])?(\d+):(\d{1,2})(?::(\d{1,2}(?:[.,]\d+)?))?$`)

// durationSeconds understands units ("1h30m", "90 minutes"), ISO 8601
// (PT1H30M), clock forms (1:30:00 = h:m:s, 1:30 = m:s) and plain seconds.
func durationSeconds(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("empty duration")
	}
	if n, err := strconv.ParseFloat(strings.Replace(s, ",", ".", 1), 64); err == nil {
		return n, nil
	}
	if m := clockRe.FindStringSubmatch(s); m != nil {
		a, _ := strconv.ParseFloat(m[2], 64)
		b, _ := strconv.ParseFloat(m[3], 64)
		var tot float64
		if m[4] == "" {
			tot = a*60 + b
		} else {
			c, _ := strconv.ParseFloat(strings.Replace(m[4], ",", ".", 1), 64)
			tot = a*3600 + b*60 + c
		}
		if m[1] == "-" {
			tot = -tot
		}
		return tot, nil
	}
	if strings.HasPrefix(strings.TrimLeft(s, "+-"), "P") || strings.HasPrefix(strings.TrimLeft(s, "+-"), "p") {
		return isoDurationSeconds(s)
	}
	d, err := parseSpec(s)
	if err != nil {
		return 0, err
	}
	return d.seconds(), nil
}

func plural(n int64, unit string) string {
	if n == 1 || n == -1 {
		return fmt.Sprintf("%d %s", n, unit)
	}
	return fmt.Sprintf("%d %ss", n, unit)
}

func trimFloat(f float64) string {
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(f, 'f', 6, 64), "0"), ".")
}

// splitSeconds breaks |sec| into d/h/m/s plus the fraction (rounded to microseconds).
func splitSeconds(sec float64) (neg bool, d, h, m int64, s float64) {
	if sec < 0 {
		neg, sec = true, -sec
	}
	sec = math.Round(sec*1e6) / 1e6
	whole := int64(sec)
	frac := sec - float64(whole)
	d, h, m = whole/86400, whole%86400/3600, whole%3600/60
	s = float64(whole%60) + frac
	return
}

func formatDuration(sec float64, style string) (string, error) {
	if math.IsNaN(sec) || math.IsInf(sec, 0) || math.Abs(sec) > 1e15 {
		return "", fmt.Errorf("seconds out of range")
	}
	neg, d, h, m, s := splitSeconds(sec)
	sign := ""
	if neg {
		sign = "-"
	}
	switch strings.ToLower(style) {
	case "compact", "":
		var parts []string
		if d > 0 {
			parts = append(parts, fmt.Sprintf("%dd", d))
		}
		if h > 0 {
			parts = append(parts, fmt.Sprintf("%dh", h))
		}
		if m > 0 {
			parts = append(parts, fmt.Sprintf("%dm", m))
		}
		if s > 0 || len(parts) == 0 {
			parts = append(parts, trimFloat(s)+"s")
		}
		return sign + strings.Join(parts, ""), nil
	case "long":
		var parts []string
		if d > 0 {
			parts = append(parts, plural(d, "day"))
		}
		if h > 0 {
			parts = append(parts, plural(h, "hour"))
		}
		if m > 0 {
			parts = append(parts, plural(m, "minute"))
		}
		if s > 0 || len(parts) == 0 {
			if s == math.Trunc(s) {
				parts = append(parts, plural(int64(s), "second"))
			} else {
				parts = append(parts, trimFloat(s)+" seconds")
			}
		}
		return sign + strings.Join(parts, ", "), nil
	case "iso8601", "iso", "iso_8601":
		var b strings.Builder
		b.WriteString(sign + "P")
		if d > 0 {
			fmt.Fprintf(&b, "%dD", d)
		}
		if h > 0 || m > 0 || s > 0 || d == 0 {
			b.WriteString("T")
			if h > 0 {
				fmt.Fprintf(&b, "%dH", h)
			}
			if m > 0 {
				fmt.Fprintf(&b, "%dM", m)
			}
			if s > 0 || (h == 0 && m == 0) {
				b.WriteString(trimFloat(s) + "S")
			}
		}
		return b.String(), nil
	case "clock", "hms":
		hours := d*24 + h
		ss := trimFloat(s)
		if s < 10 {
			ss = "0" + ss
		}
		return fmt.Sprintf("%s%d:%02d:%s", sign, hours, m, ss), nil
	}
	return "", fmt.Errorf("style must be compact, long, iso8601 or clock")
}

// timeMath evaluates "1:30 + 2:45 - 0:20" and returns H:MM[:SS].
func timeMath(expr string) (string, error) {
	e := strings.TrimSpace(expr)
	if e == "" {
		return "", fmt.Errorf("empty expression")
	}
	var terms []string
	var ops []float64
	cur := strings.Builder{}
	sign := 1.0
	flush := func() error {
		t := strings.TrimSpace(cur.String())
		cur.Reset()
		if t == "" {
			return nil
		}
		var v float64
		var err error
		if m := clockRe.FindStringSubmatch(t); m != nil && m[4] == "" {
			a, _ := strconv.ParseFloat(m[2], 64)
			b, _ := strconv.ParseFloat(m[3], 64)
			v = a*3600 + b*60 // two parts mean h:mm in time math
			if m[1] == "-" {
				v = -v
			}
		} else {
			v, err = durationSeconds(t)
		}
		if err != nil {
			return err
		}
		terms = append(terms, t)
		ops = append(ops, sign*v)
		return nil
	}
	prevTerm := false
	for _, r := range e {
		if (r == '+' || r == '-') && prevTerm && cur.Len() > 0 && strings.TrimSpace(cur.String()) != "" {
			// an operator only after a finished term
			if err := flush(); err != nil {
				return "", err
			}
			if r == '-' {
				sign = -1
			} else {
				sign = 1
			}
			prevTerm = false
			continue
		}
		if (r == '+' || r == '-') && !prevTerm && strings.TrimSpace(cur.String()) == "" {
			if r == '-' {
				sign = -sign
			}
			continue
		}
		cur.WriteRune(r)
		if r != ' ' {
			prevTerm = true
		}
	}
	if err := flush(); err != nil {
		return "", err
	}
	if len(ops) == 0 {
		return "", fmt.Errorf("no terms found in %q", expr)
	}
	total := 0.0
	for _, v := range ops {
		total += v
	}
	neg, d, h, m, s := splitSeconds(total)
	hours := d*24 + h
	sg := ""
	if neg {
		sg = "-"
	}
	if s != 0 {
		ss := trimFloat(s)
		if s < 10 {
			ss = "0" + ss
		}
		return fmt.Sprintf("%s%d:%02d:%s", sg, hours, m, ss), nil
	}
	return fmt.Sprintf("%s%d:%02d", sg, hours, m), nil
}

// ---- epochs ----

var (
	epochFiletime = time.Date(1601, 1, 1, 0, 0, 0, 0, time.UTC)
	epochDotnet   = time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC)
	epochMac      = time.Date(2001, 1, 1, 0, 0, 0, 0, time.UTC)
)

const unixJulian = 2440587.5

func julianDay(t time.Time) float64 {
	return float64(t.Unix())/86400 + float64(t.Nanosecond())/86400e9 + unixJulian
}

func fromJulianDay(jd float64) time.Time {
	sec := (jd - unixJulian) * 86400
	whole := math.Floor(sec)
	return time.Unix(int64(whole), int64(math.Round((sec-whole)*1e9))).UTC()
}

// since returns whole units of unit nanoseconds since epoch (floor).
func ticksSince(t, epoch time.Time, nsPerUnit int64) int64 {
	s := t.Unix() - epoch.Unix()
	return s*(1e9/nsPerUnit) + int64(t.Nanosecond())/nsPerUnit
}

func fromTicks(v int64, epoch time.Time, nsPerUnit int64) time.Time {
	per := 1e9 / nsPerUnit
	sec, rem := v/per, v%per
	if rem < 0 {
		sec--
		rem += per
	}
	return time.Unix(epoch.Unix()+sec, rem*nsPerUnit).UTC()
}

// excelSerial: Excel counts 1900-02-29 (which never existed) as day 60.
func excelSerial(t time.Time, date1904 bool) (float64, error) {
	y, m, d := t.Date()
	wall := time.Date(y, m, d, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
	if date1904 {
		return wall.Sub(time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC)).Hours() / 24, nil
	}
	if y < 1900 {
		return 0, fmt.Errorf("Excel serial dates start at 1900-01-01")
	}
	days := wall.Sub(time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)).Seconds() / 86400
	if days < 61 {
		days-- // dates before 1900-03-01 skip the phantom leap day
	}
	return days, nil
}

func fromExcelSerial(v float64, date1904 bool, loc *time.Location) (time.Time, error) {
	if math.IsNaN(v) || math.IsInf(v, 0) || math.Abs(v) > 1e8 {
		return time.Time{}, fmt.Errorf("serial out of range")
	}
	var base time.Time
	switch {
	case date1904:
		base = time.Date(1904, 1, 1, 0, 0, 0, 0, time.UTC)
	case v >= 61:
		base = time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC)
	case v >= 60:
		return time.Time{}, fmt.Errorf("serial 60 is Excel's non-existent 1900-02-29")
	case v >= 0:
		base = time.Date(1899, 12, 31, 0, 0, 0, 0, time.UTC)
	default:
		return time.Time{}, fmt.Errorf("negative serials are not valid in the 1900 date system")
	}
	ns := int64(math.Round(v * 86400e9 / 1000)) // microsecond precision
	w := base.Add(time.Duration(ns) * 1000)
	y, mo, d := w.Date()
	return time.Date(y, mo, d, w.Hour(), w.Minute(), w.Second(), w.Nanosecond(), loc), nil
}

// ---- diff and age ----

type calDiff struct {
	years, months, days, hours, minutes, seconds int
	neg                                          bool
}

func calendarDiff(a, b time.Time) calDiff {
	neg := false
	if b.Before(a) {
		a, b = b, a
		neg = true
	}
	b = b.In(a.Location())
	years := b.Year() - a.Year()
	for years > 0 && addMonths(a, 12*years).After(b) {
		years--
	}
	months := 0
	for months < 11 && !addMonths(a, 12*years+months+1).After(b) {
		months++
	}
	base := addMonths(a, 12*years+months)
	days := 0
	for !base.AddDate(0, 0, days+1).After(b) {
		days++
	}
	rem := b.Sub(base.AddDate(0, 0, days))
	return calDiff{years, months, days, int(rem / time.Hour), int(rem % time.Hour / time.Minute), int(rem % time.Minute / time.Second), neg}
}

func (c calDiff) human() string {
	var parts []string
	add := func(n int, u string) {
		if n > 0 {
			parts = append(parts, plural(int64(n), u))
		}
	}
	add(c.years, "year")
	add(c.months, "month")
	add(c.days, "day")
	add(c.hours, "hour")
	add(c.minutes, "minute")
	add(c.seconds, "second")
	if len(parts) == 0 {
		return "0 seconds"
	}
	return strings.Join(parts, ", ")
}

// humanRelative renders d (positive = future) as "in 3 days" or "2 hours ago".
func humanRelative(d time.Duration) string {
	future := d > 0
	a := d
	if a < 0 {
		a = -a
	}
	sec := int64(a / time.Second)
	var s string
	switch {
	case sec < 10:
		return "just now"
	case sec < 60:
		s = plural(sec, "second")
	case sec < 3600:
		s = plural(sec/60, "minute")
	case sec < 86400:
		s = plural(sec/3600, "hour")
	default:
		days := sec / 86400
		switch {
		case days < 14:
			s = plural(days, "day")
		case days < 56:
			s = plural(days/7, "week")
		case days < 365:
			s = plural(max(int64(float64(days)/30.4375), 2), "month")
		default:
			s = plural(int64(math.Round(float64(days)/365.25)), "year")
		}
	}
	if future {
		return "in " + s
	}
	return s + " ago"
}

// ---- business days ----

func parseWeekend(s string) (map[time.Weekday]bool, error) {
	out := map[time.Weekday]bool{}
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		s = "sat_sun"
	}
	if s == "none" {
		return out, nil
	}
	for _, f := range strings.FieldsFunc(s, func(r rune) bool { return r == '_' || r == ',' || r == ';' || r == ' ' || r == '/' || r == '+' }) {
		w, ok := parseWeekday(f)
		if !ok {
			return nil, fmt.Errorf("weekend: unknown day %q (use e.g. sat_sun, fri_sat, sun or none)", f)
		}
		out[w] = true
	}
	return out, nil
}

func holidaySet(list []string, o parseOpts) (map[string]bool, error) {
	out := map[string]bool{}
	for _, h := range list {
		t, err := parseAny(h, o)
		if err != nil {
			return nil, fmt.Errorf("holiday %q: %w", h, err)
		}
		out[t.Format("2006-01-02")] = true
	}
	return out, nil
}

// ---- Easter ----

func easterWestern(y int) (int, int) {
	a, b, c := y%19, y/100, y%100
	d, e := b/4, b%4
	f := (b + 8) / 25
	g := (b - f + 1) / 3
	h := (19*a + b - d - g + 15) % 30
	i, k := c/4, c%4
	l := (32 + 2*e + 2*i - h - k) % 7
	m := (a + 11*h + 22*l) / 451
	month := (h + l - 7*m + 114) / 31
	day := (h+l-7*m+114)%31 + 1
	return month, day
}

func easterOrthodox(y int) time.Time {
	a, b, c := y%4, y%7, y%19
	d := (19*c + 15) % 30
	e := (2*a + 4*b - d + 34) % 7
	month := (d + e + 114) / 31
	day := (d+e+114)%31 + 1
	julian := time.Date(y, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	offset := y/100 - y/400 - 2
	return julian.AddDate(0, 0, offset)
}

// ---- calendar text ----

func calendarRows(y int, m time.Month, weekStart time.Weekday) [][]int {
	first := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
	lead := (int(first.Weekday()) - int(weekStart) + 7) % 7
	n := daysIn(y, m)
	var rows [][]int
	row := make([]int, 0, 7)
	for i := 0; i < lead; i++ {
		row = append(row, 0)
	}
	for d := 1; d <= n; d++ {
		row = append(row, d)
		if len(row) == 7 {
			rows = append(rows, row)
			row = make([]int, 0, 7)
		}
	}
	if len(row) > 0 {
		for len(row) < 7 {
			row = append(row, 0)
		}
		rows = append(rows, row)
	}
	return rows
}

func calendarText(y int, m time.Month, weekStart time.Weekday) string {
	title := fmt.Sprintf("%s %d", m, y)
	pad := (20 - len(title)) / 2
	if pad < 0 {
		pad = 0
	}
	var b strings.Builder
	b.WriteString(strings.Repeat(" ", pad) + title + "\n")
	hdr := make([]string, 7)
	for i := range hdr {
		hdr[i] = time.Weekday((int(weekStart) + i) % 7).String()[:2]
	}
	b.WriteString(strings.Join(hdr, " "))
	for _, r := range calendarRows(y, m, weekStart) {
		b.WriteString("\n")
		cells := make([]string, 7)
		for i, d := range r {
			if d == 0 {
				cells[i] = "  "
			} else {
				cells[i] = fmt.Sprintf("%2d", d)
			}
		}
		b.WriteString(strings.TrimRight(strings.Join(cells, " "), " "))
	}
	return b.String()
}

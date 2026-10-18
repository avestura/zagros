package dt

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/avestura/zagros/internal/tools"
)

func specP() tools.Param {
	return req("spec", "str", "Duration spec such as 1y2mo3w4d5h6m7s, '-3 days', '2 hours 30 minutes'. Units: y mo w d h m s ms us ns. Months and years are calendar-aware and clamp to month end.")
}

func dateStr(t time.Time) string { return t.Format("2006-01-02") }

// dayNumber is the wall-clock date of t as a UTC midnight, for day counting.
func dayNumber(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
}

func daysBetweenDates(a, b time.Time) int {
	return int(math.Round(dayNumber(b).Sub(dayNumber(a)).Hours() / 24))
}

func weekStartArg(s string) (time.Weekday, error) {
	if strings.TrimSpace(s) == "" {
		return time.Monday, nil
	}
	w, ok := parseWeekday(s)
	if !ok {
		return 0, fmt.Errorf("week_start must be a weekday name such as monday or sunday")
	}
	return w, nil
}

func init() {
	addFn := func(sign int) func(a tools.Args) (any, error) {
		return func(a tools.Args) (any, error) {
			t, _, err := when(a, "text")
			if err != nil {
				return nil, err
			}
			d, err := parseSpec(a.Str("spec"))
			if err != nil {
				return nil, err
			}
			return rfc(d.apply(t, sign)), nil
		}
	}
	reg(tools.Func{
		Name: "add", Title: "Add to date", Order: 6, Input: "text", Keywords: []string{"add", "plus"},
		Doc:     "Adds a duration to a date (empty text = now). The spec mixes units and signs: '1y2mo3w4d5h6m7s', '-90m', '1d-2h'. Month/year steps clamp to the month end (Jan 31 + 1mo = Feb 28/29); day steps keep the wall-clock time across DST.",
		Params:  []tools.Param{req("text", "str", "Date/time text; use now for the current time."), specP(), tzP(), atP(), dayFirstP()},
		Returns: "str", Example: `dt.add("2024-01-31", "1mo", "UTC")`,
		Tags: []string{"date math", "plus days", "in 3 weeks", "date arithmetic", "shift date", "add days"},
		Fn:   addFn(1),
	})
	reg(tools.Func{
		Name: "subtract", Title: "Subtract from date", Input: "text", Keywords: []string{"subtract", "minus"},
		Doc:     "Subtracts a duration from a date (empty text = now); the same spec syntax as add(), so a negative spec adds.",
		Params:  []tools.Param{req("text", "str", "Date/time text; use now for the current time."), specP(), tzP(), atP(), dayFirstP()},
		Returns: "str", Example: `dt.subtract("2024-03-31", "1mo", "UTC")`,
		Tags: []string{"date math", "minus days", "ago", "date arithmetic", "subtract days"},
		Fn:   addFn(-1),
	})
	reg(tools.Func{
		Name: "diff", Title: "Difference between dates", Order: 5, Input: "a", Keywords: []string{"diff", "between", "difference"},
		Doc:     "Compares two dates: total seconds, minutes, hours, days and weeks plus a calendar breakdown (years, months, days, hours, minutes, seconds) and human text. Direction says whether b is after or before a. Empty b means now.",
		Params:  []tools.Param{req("a", "str", "Start date/time."), p("b", "str", "End date/time; empty means now.", ""), tzP(), atP(), dayFirstP()},
		Returns: "dict[str, str]", Example: `dt.diff("2020-02-29", "2024-03-01 12:00", "UTC")`,
		Tags: []string{"difference between dates", "time between", "how long", "duration between", "elapsed", "days between", "calendar difference"},
		Fn: func(a tools.Args) (any, error) {
			o, err := opts(a)
			if err != nil {
				return nil, err
			}
			ta, err := parseAny(a.Str("a"), o)
			if err != nil {
				return nil, fmt.Errorf("a: %w", err)
			}
			tb := o.now
			if s := strings.TrimSpace(a.Str("b")); s != "" {
				if tb, err = parseAny(s, o); err != nil {
					return nil, fmt.Errorf("b: %w", err)
				}
			}
			d := tb.Sub(ta)
			c := calendarDiff(ta, tb)
			dir := "b is after a"
			if d < 0 {
				dir = "b is before a"
			} else if d == 0 {
				dir = "same instant"
			}
			sec := d.Seconds()
			return map[string]string{
				"from": rfc(ta), "to": rfc(tb), "direction": dir,
				"total_seconds": trimFloat(sec), "total_minutes": trimFloat(sec / 60), "total_hours": trimFloat(sec / 3600),
				"total_days": trimFloat(sec / 86400), "total_weeks": trimFloat(sec / 604800),
				"years": strconv.Itoa(c.years), "months": strconv.Itoa(c.months), "days": strconv.Itoa(c.days),
				"hours": strconv.Itoa(c.hours), "minutes": strconv.Itoa(c.minutes), "seconds": strconv.Itoa(c.seconds),
				"human": c.human(),
			}, nil
		},
	})
	reg(tools.Func{
		Name: "age", Title: "Age calculator", Input: "birthdate", Keywords: []string{"age", "birthday"},
		Doc:     "Calculates age from a birth date: years, months and days, total days lived, the next birthday and days until it. A Feb 29 birthday is celebrated on Feb 28 in non-leap years. Errors if the birth date is after the reference date.",
		Params:  []tools.Param{req("birthdate", "str", "Birth date."), p("at", "str", "Date to measure at (empty = now).", ""), tzP(), dayFirstP()},
		Returns: "dict[str, str]", Example: `dt.age("1990-05-17", "2024-01-31", "UTC")`,
		Tags: []string{"how old", "birthday", "years old", "age in days", "next birthday", "date of birth"},
		Fn: func(a tools.Args) (any, error) {
			o, err := opts(a)
			if err != nil {
				return nil, err
			}
			b, err := parseAny(a.Str("birthdate"), o)
			if err != nil {
				return nil, err
			}
			now := o.now
			if now.Before(b) {
				return nil, fmt.Errorf("birth date is after the reference date")
			}
			c := calendarDiff(b, now)
			next := addMonths(b, 12*c.years)
			if daysBetweenDates(now, next) != 0 {
				next = addMonths(b, 12*(c.years+1))
			}
			return map[string]string{
				"years": strconv.Itoa(c.years), "months": strconv.Itoa(c.months), "days": strconv.Itoa(c.days),
				"total_days": strconv.Itoa(daysBetweenDates(b, now)), "total_months": strconv.Itoa(c.years*12 + c.months),
				"next_birthday": dateStr(next), "days_until_birthday": strconv.Itoa(daysBetweenDates(now, next)),
				"next_birthday_weekday": next.Weekday().String(),
				"text":                  fmt.Sprintf("%d years, %d months, %d days", c.years, c.months, c.days),
			}, nil
		},
	})
	boundary := func(end bool) func(a tools.Args) (any, error) {
		return func(a tools.Args) (any, error) {
			t, _, err := when(a, "text")
			if err != nil {
				return nil, err
			}
			u, ok := unitNames[strings.ToLower(strings.TrimSpace(a.Str("unit")))]
			if !ok {
				return nil, fmt.Errorf("unit must be minute, hour, day, week, month, quarter or year")
			}
			ws, err := weekStartArg(a.Str("week_start"))
			if err != nil {
				return nil, err
			}
			if end {
				return rfc(endOf(t, u, ws)), nil
			}
			return rfc(startOf(t, u, ws)), nil
		}
	}
	boundaryParams := func() []tools.Param {
		return []tools.Param{p("text", "str", "Date/time text; empty means now.", ""), p("unit", "str", "minute, hour, day, week, month, quarter or year.", "day"), p("week_start", "str", "First day of the week (monday or sunday...).", "monday"), tzP(), atP(), dayFirstP()}
	}
	reg(tools.Func{
		Name: "start_of", Title: "Start of period", Input: "text", Keywords: []string{"start", "beginning"},
		Doc:    "Returns the first instant of the minute, hour, day, week, month, quarter or year containing the date. Weeks begin on week_start (default Monday).",
		Params: boundaryParams(), Returns: "str", Example: `dt.start_of("2024-05-17 13:45", "month", tz="UTC")`,
		Tags: []string{"beginning of month", "first day of week", "truncate date", "floor date", "midnight"},
		Fn:   boundary(false),
	})
	reg(tools.Func{
		Name: "end_of", Title: "End of period", Input: "text", Keywords: []string{"end"},
		Doc:    "Returns the last nanosecond of the minute, hour, day, week, month, quarter or year containing the date, e.g. 23:59:59.999999999 for a day.",
		Params: boundaryParams(), Returns: "str", Example: `dt.end_of("2024-02-10", "month", tz="UTC")`,
		Tags: []string{"last day of month", "end of week", "end of day", "ceil date"},
		Fn:   boundary(true),
	})
	reg(tools.Func{
		Name: "day_of_week", Title: "Day of the week", Input: "text", Keywords: []string{"weekday"},
		Doc:     "Returns the weekday of a date. style: name (Wednesday), short (Wed), iso (1=Monday..7=Sunday) or number (0=Sunday..6).",
		Params:  []tools.Param{p("text", "str", "Date text; empty means today.", ""), p("style", "str", "name, short, iso or number.", "name"), tzP(), atP(), dayFirstP()},
		Returns: "str", Example: `dt.day_of_week("2024-02-29")`,
		Tags: []string{"weekday", "what day is it", "which day of the week", "monday"},
		Fn: func(a tools.Args) (any, error) {
			t, _, err := when(a, "text")
			if err != nil {
				return nil, err
			}
			switch strings.ToLower(a.Str("style")) {
			case "name", "":
				return t.Weekday().String(), nil
			case "short":
				return t.Weekday().String()[:3], nil
			case "iso":
				return strconv.Itoa((int(t.Weekday())+6)%7 + 1), nil
			case "number":
				return strconv.Itoa(int(t.Weekday())), nil
			}
			return nil, fmt.Errorf("style must be name, short, iso or number")
		},
	})
	intOf := func(f func(t time.Time) int) func(a tools.Args) (any, error) {
		return func(a tools.Args) (any, error) {
			t, _, err := when(a, "text")
			if err != nil {
				return nil, err
			}
			return int64(f(t)), nil
		}
	}
	dtParams := func() []tools.Param {
		return []tools.Param{p("text", "str", "Date text; empty means today.", ""), tzP(), atP(), dayFirstP()}
	}
	reg(tools.Func{
		Name: "day_of_year", Title: "Day of the year", Input: "text",
		Doc: "Returns the day of the year, 1..366 (ordinal date).", Params: dtParams(), Returns: "int", Example: `dt.day_of_year("2024-12-31")`,
		Tags: []string{"julian day number", "ordinal date", "yday", "nth day"},
		Fn:   intOf(func(t time.Time) int { return t.YearDay() }),
	})
	reg(tools.Func{
		Name: "quarter", Title: "Quarter of the year", Input: "text",
		Doc: "Returns the calendar quarter 1..4 of a date.", Params: dtParams(), Returns: "int", Example: `dt.quarter("2024-08-15")`,
		Tags: []string{"q1", "fiscal quarter", "q3", "three months"},
		Fn:   intOf(func(t time.Time) int { return (int(t.Month())-1)/3 + 1 }),
	})
	reg(tools.Func{
		Name: "week_of_year", Title: "Week number", Input: "text", Keywords: []string{"week"},
		Doc:     "Returns the week number of a date. system: iso (ISO 8601, weeks start Monday, week 1 has Jan 4), us (week 1 contains Jan 1, weeks start Sunday), sunday (strftime %U) or monday (strftime %W).",
		Params:  append(dtParams(), p("system", "str", "iso, us, sunday or monday.", "iso")),
		Returns: "int", Example: `dt.week_of_year("2024-12-30", system="iso")`,
		Tags: []string{"iso week", "calendar week", "week number", "kw", "cw"},
		Fn: func(a tools.Args) (any, error) {
			t, _, err := when(a, "text")
			if err != nil {
				return nil, err
			}
			switch strings.ToLower(a.Str("system")) {
			case "iso", "":
				_, w := t.ISOWeek()
				return int64(w), nil
			case "us":
				jan1 := time.Date(t.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
				return int64((t.YearDay()-1+int(jan1.Weekday()))/7 + 1), nil
			case "sunday":
				return int64(weekNumber(t, time.Sunday)), nil
			case "monday":
				return int64(weekNumber(t, time.Monday)), nil
			}
			return nil, fmt.Errorf("system must be iso, us, sunday or monday")
		},
	})
	reg(tools.Func{
		Name: "iso_week_date", Title: "ISO week date", Input: "text", Keywords: []string{"week"},
		Doc:    "Returns the ISO 8601 week date of a date, e.g. 2024-W05-3 (week-based year, week number, weekday 1=Monday). The week-based year can differ from the calendar year around New Year.",
		Params: dtParams(), Returns: "str", Example: `dt.iso_week_date("2024-12-30")`,
		Tags: []string{"iso 8601 week", "week date", "2024-W01", "week-based year"},
		Fn: func(a tools.Args) (any, error) {
			t, _, err := when(a, "text")
			if err != nil {
				return nil, err
			}
			y, w := t.ISOWeek()
			return fmt.Sprintf("%04d-W%02d-%d", y, w, (int(t.Weekday())+6)%7+1), nil
		},
	})
	reg(tools.Func{
		Name: "is_weekend", Title: "Is it a weekend?", Input: "text",
		Doc:     "Returns true when the date falls on a weekend; weekend lists the days (default sat_sun, use fri_sat for Middle Eastern calendars).",
		Params:  append(dtParams(), p("weekend", "str", "sat_sun, fri_sat, sun, ... or none.", "sat_sun")),
		Returns: "bool", Example: `dt.is_weekend("2024-02-03")`,
		Tags: []string{"saturday sunday", "workday", "day off"},
		Fn: func(a tools.Args) (any, error) {
			t, _, err := when(a, "text")
			if err != nil {
				return nil, err
			}
			w, err := parseWeekend(a.Str("weekend"))
			if err != nil {
				return nil, err
			}
			return w[t.Weekday()], nil
		},
	})
	reg(tools.Func{
		Name: "is_leap_year", Title: "Leap year?", Input: "year", Keywords: []string{"leap"},
		Doc:     "Returns true when the year is a leap year in the Gregorian calendar (divisible by 4, except centuries not divisible by 400).",
		Params:  []tools.Param{p("year", "str", "Year such as 2024; empty means this year.", "")},
		Returns: "bool", Example: `dt.is_leap_year("2100")`,
		Tags: []string{"leap", "feb 29", "366 days", "gregorian"},
		Fn: func(a tools.Args) (any, error) {
			y, err := yearArg(a.Str("year"))
			return isLeap(y), err
		},
	})
	reg(tools.Func{
		Name: "days_in_month", Title: "Days in a month",
		Doc:     "Returns the number of days (28..31) in a month of a year, honouring leap years. Month is 1..12.",
		Params:  []tools.Param{req("year", "int", "Year."), req("month", "int", "Month 1..12.")},
		Returns: "int", Example: `dt.days_in_month(2024, 2)`,
		Tags: []string{"month length", "how many days in february", "28 29 30 31"},
		Fn: func(a tools.Args) (any, error) {
			if a.Int("month") < 1 || a.Int("month") > 12 {
				return nil, fmt.Errorf("month must be 1..12")
			}
			return int64(daysIn(int(a.Int("year")), time.Month(a.Int("month")))), nil
		},
	})
	reg(tools.Func{
		Name: "days_in_year", Title: "Days in a year", Input: "year",
		Doc:     "Returns 365 or 366 for the year.",
		Params:  []tools.Param{p("year", "str", "Year such as 2024; empty means this year.", "")},
		Returns: "int", Example: `dt.days_in_year("2024")`,
		Tags: []string{"year length", "365", "366"},
		Fn: func(a tools.Args) (any, error) {
			y, err := yearArg(a.Str("year"))
			return int64(365 + b2i(isLeap(y))), err
		},
	})
	reg(tools.Func{
		Name: "days_between", Title: "Days between dates", Input: "a",
		Doc:     "Returns the number of calendar days from a to b (negative if b is earlier), ignoring the time of day. inclusive counts both end dates. Empty b means today.",
		Params:  []tools.Param{req("a", "str", "Start date."), p("b", "str", "End date; empty means today.", ""), p("inclusive", "bool", "Count both the first and the last day.", false), tzP(), atP(), dayFirstP()},
		Returns: "int", Example: `dt.days_between("2024-01-01", "2024-12-31")`,
		Tags: []string{"number of days", "days until", "days since", "how many days"},
		Fn: func(a tools.Args) (any, error) {
			o, err := opts(a)
			if err != nil {
				return nil, err
			}
			ta, err := parseAny(a.Str("a"), o)
			if err != nil {
				return nil, err
			}
			tb := o.now
			if s := strings.TrimSpace(a.Str("b")); s != "" {
				if tb, err = parseAny(s, o); err != nil {
					return nil, err
				}
			}
			n := daysBetweenDates(ta, tb)
			if a.Bool("inclusive") {
				if n >= 0 {
					n++
				} else {
					n--
				}
			}
			return int64(n), nil
		},
	})
	reg(tools.Func{
		Name: "days_until", Title: "Days until a date", Input: "target",
		Doc:     "Returns the number of calendar days from today (or at) until the target date; negative when it has passed.",
		Params:  []tools.Param{req("target", "str", "Target date."), tzP(), atP(), dayFirstP()},
		Returns: "int", Example: `dt.days_until("2030-01-01", at="2024-01-01")`,
		Tags: []string{"countdown days", "days left", "days to go", "how many days until"},
		Fn: func(a tools.Args) (any, error) {
			o, err := opts(a)
			if err != nil {
				return nil, err
			}
			t, err := parseAny(a.Str("target"), o)
			if err != nil {
				return nil, err
			}
			return int64(daysBetweenDates(o.now, t)), nil
		},
	})
	bizParams := func(first ...tools.Param) []tools.Param {
		return append(first, p("holidays", "list[str]", "Dates to treat as non-working.", []string(nil)),
			p("weekend", "str", "Weekend days: sat_sun (default), fri_sat, sun, none, or a list like fri,sat.", "sat_sun"), tzP(), dayFirstP())
	}
	reg(tools.Func{
		Name: "business_days_between", Title: "Business days between", Input: "a", Keywords: []string{"business", "working"},
		Doc:     "Counts working days from a up to b: weekends and listed holidays are skipped. b itself is excluded unless inclusive is true; the result is negative when b is before a.",
		Params:  bizParams(req("a", "str", "Start date."), req("b", "str", "End date."), p("inclusive", "bool", "Also count b when it is a working day.", false)),
		Returns: "int", Example: `dt.business_days_between("2024-01-01", "2024-01-31", ["2024-01-01"])`,
		Tags: []string{"working days", "workdays", "networkdays", "weekdays between", "excluding weekends", "banking days"},
		Fn: func(a tools.Args) (any, error) {
			o, err := opts(a)
			if err != nil {
				return nil, err
			}
			ta, err := parseAny(a.Str("a"), o)
			if err != nil {
				return nil, err
			}
			tb, err := parseAny(a.Str("b"), o)
			if err != nil {
				return nil, err
			}
			wk, err := parseWeekend(a.Str("weekend"))
			if err != nil {
				return nil, err
			}
			hol, err := holidaySet(a.List("holidays"), o)
			if err != nil {
				return nil, err
			}
			lo, hi, sign := dayNumber(ta), dayNumber(tb), int64(1)
			if hi.Before(lo) {
				lo, hi, sign = hi, lo, -1
			}
			if a.Bool("inclusive") {
				hi = hi.AddDate(0, 0, 1)
			}
			if hi.Sub(lo) > 366*200*24*time.Hour {
				return nil, fmt.Errorf("range is too large (max 200 years)")
			}
			var n int64
			for d := lo; d.Before(hi); d = d.AddDate(0, 0, 1) {
				if !wk[d.Weekday()] && !hol[dateStr(d)] {
					n++
				}
			}
			return sign * n, nil
		},
	})
	reg(tools.Func{
		Name: "add_business_days", Title: "Add business days", Input: "text", Keywords: []string{"business", "working"},
		Doc:     "Moves a date forward (or backward when days is negative) by a number of working days, skipping weekends and holidays; the start day is not counted and the time of day is kept. |days| is limited to 100000.",
		Params:  bizParams(req("text", "str", "Start date; use today for today."), req("days", "int", "Working days to add (negative goes back).")),
		Returns: "str", Example: `dt.add_business_days("2024-01-31", 5, ["2024-02-02"], tz="UTC")`,
		Tags: []string{"workday", "due date", "deadline", "skip weekends", "working days from now", "delivery date"},
		Fn: func(a tools.Args) (any, error) {
			t, o, err := when(a, "text")
			if err != nil {
				return nil, err
			}
			n := a.Int("days")
			if n > 100000 || n < -100000 {
				return nil, fmt.Errorf("days must be within +-100000")
			}
			wk, err := parseWeekend(a.Str("weekend"))
			if err != nil {
				return nil, err
			}
			if len(wk) >= 7 {
				return nil, fmt.Errorf("every day is a weekend")
			}
			hol, err := holidaySet(a.List("holidays"), o)
			if err != nil {
				return nil, err
			}
			step := 1
			if n < 0 {
				step, n = -1, -n
			}
			for n > 0 {
				t = t.AddDate(0, 0, step)
				if !wk[t.Weekday()] && !hol[dateStr(t)] {
					n--
				}
			}
			return rfc(t), nil
		},
	})
	reg(tools.Func{
		Name: "nth_weekday", Title: "Nth weekday of a month",
		Doc:     "Returns the date of the n-th given weekday of a month (n = 1..5), or counted from the end with negative n (-1 = last). Errors when that date does not exist (e.g. a 5th Friday). Example: US Thanksgiving is the 4th Thursday of November.",
		Params:  []tools.Param{req("year", "int", "Year."), req("month", "int", "Month 1..12."), req("weekday", "str", "Weekday name such as thursday."), req("n", "int", "Occurrence 1..5, or -1 for the last.")},
		Returns: "str", Example: `dt.nth_weekday(2024, 11, "thursday", 4)`,
		Tags: []string{"thanksgiving", "second tuesday", "last friday of month", "holiday date", "patch tuesday"},
		Fn: func(a tools.Args) (any, error) {
			wd, ok := parseWeekday(a.Str("weekday"))
			if !ok {
				return nil, fmt.Errorf("unknown weekday %q", a.Str("weekday"))
			}
			if a.Int("month") < 1 || a.Int("month") > 12 {
				return nil, fmt.Errorf("month must be 1..12")
			}
			y, m, n := int(a.Int("year")), time.Month(a.Int("month")), int(a.Int("n"))
			if n == 0 || n > 5 || n < -5 {
				return nil, fmt.Errorf("n must be 1..5 or -1..-5")
			}
			var d time.Time
			if n > 0 {
				first := time.Date(y, m, 1, 0, 0, 0, 0, time.UTC)
				d = first.AddDate(0, 0, (int(wd)-int(first.Weekday())+7)%7+7*(n-1))
			} else {
				last := time.Date(y, m, daysIn(y, m), 0, 0, 0, 0, time.UTC)
				d = last.AddDate(0, 0, -((int(last.Weekday())-int(wd)+7)%7)-7*(-n-1))
			}
			if d.Month() != m {
				return nil, fmt.Errorf("there is no occurrence %d of %s in %s %d", n, wd, m, y)
			}
			return dateStr(d), nil
		},
	})
	reg(tools.Func{
		Name: "easter", Title: "Easter date", Input: "year", Keywords: []string{"easter"},
		Doc:     "Returns the date of Easter Sunday for a year (1583..9999 Gregorian). calendar: western (default) or orthodox (Julian computus, shown in the Gregorian calendar).",
		Params:  []tools.Param{p("year", "str", "Year such as 2025; empty means this year.", ""), p("calendar", "str", "western or orthodox.", "western")},
		Returns: "str", Example: `dt.easter("2025")`,
		Tags: []string{"easter sunday", "holiday", "computus", "movable feast", "good friday", "pascha"},
		Fn: func(a tools.Args) (any, error) {
			y, err := yearArg(a.Str("year"))
			if err != nil {
				return nil, err
			}
			if y < 1583 || y > 9999 {
				return nil, fmt.Errorf("year must be 1583..9999")
			}
			switch strings.ToLower(a.Str("calendar")) {
			case "western", "gregorian", "":
				m, d := easterWestern(y)
				return fmt.Sprintf("%04d-%02d-%02d", y, m, d), nil
			case "orthodox", "eastern", "julian":
				return dateStr(easterOrthodox(y)), nil
			}
			return nil, fmt.Errorf("calendar must be western or orthodox")
		},
	})
	calParams := func() []tools.Param {
		return []tools.Param{req("year", "int", "Year."), req("month", "int", "Month 1..12."), p("week_start", "str", "monday or sunday.", "monday")}
	}
	reg(tools.Func{
		Name: "calendar", Title: "Text calendar",
		Doc:    "Returns a month as a text calendar (like the Unix cal command) with weeks starting on Monday or Sunday.",
		Params: calParams(), Returns: "str", Example: `dt.calendar(2024, 2, "sunday")`,
		Tags: []string{"cal", "month view", "calendar grid", "print calendar"},
		Fn: func(a tools.Args) (any, error) {
			if a.Int("month") < 1 || a.Int("month") > 12 {
				return nil, fmt.Errorf("month must be 1..12")
			}
			ws, err := weekStartArg(a.Str("week_start"))
			if err != nil {
				return nil, err
			}
			return calendarText(int(a.Int("year")), time.Month(a.Int("month")), ws), nil
		},
	})
	reg(tools.Func{
		Name: "month_calendar_rows", Title: "Calendar rows",
		Doc:    "Returns the weeks of a month as list rows; each row has 7 comma-separated cells with the day number, or an empty cell outside the month (',,1,2,3,4,5').",
		Params: calParams(), Returns: "list[str]", Example: `dt.month_calendar_rows(2024, 2)`,
		Tags: []string{"calendar grid", "weeks of month", "month matrix", "date picker"},
		Fn: func(a tools.Args) (any, error) {
			if a.Int("month") < 1 || a.Int("month") > 12 {
				return nil, fmt.Errorf("month must be 1..12")
			}
			ws, err := weekStartArg(a.Str("week_start"))
			if err != nil {
				return nil, err
			}
			var out []string
			for _, r := range calendarRows(int(a.Int("year")), time.Month(a.Int("month")), ws) {
				cells := make([]string, 7)
				for i, d := range r {
					if d != 0 {
						cells[i] = strconv.Itoa(d)
					}
				}
				out = append(out, strings.Join(cells, ","))
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "countdown", Title: "Countdown", Input: "target", Keywords: []string{"countdown", "until"},
		Doc:     "Returns the time remaining until a target as '12 days, 3 hours, 5 minutes, 7 seconds'; for a target in the past it says how long ago it was. Reference is at (default now).",
		Params:  []tools.Param{req("target", "str", "Target date/time."), tzP(), atP(), dayFirstP()},
		Returns: "str", Example: `dt.countdown("2030-01-01", at="2024-01-01")`,
		Tags: []string{"time left", "time until", "remaining", "deadline", "until new year"},
		Fn: func(a tools.Args) (any, error) {
			o, err := opts(a)
			if err != nil {
				return nil, err
			}
			t, err := parseAny(a.Str("target"), o)
			if err != nil {
				return nil, err
			}
			c := calendarDiff(o.now, t)
			if t.Equal(o.now) {
				return "now", nil
			}
			if c.neg {
				return c.human() + " ago", nil
			}
			return c.human(), nil
		},
	})
	reg(tools.Func{
		Name: "relative", Title: "Relative time", Order: 7, Input: "text", Keywords: []string{"relative", "ago", "humanize"},
		Doc:     "Describes a date relative to a reference as '3 hours ago', 'in 2 days' or 'just now' (within 10 seconds). Units step from seconds up to years; reference is at (default now).",
		Params:  []tools.Param{req("text", "str", "Date/time text."), tzP(), atP(), dayFirstP()},
		Returns: "str", Example: `dt.relative("2024-01-29", at="2024-01-31")`,
		Tags: []string{"time ago", "humanize", "fuzzy time", "in 2 days", "natural"},
		Fn: func(a tools.Args) (any, error) {
			t, o, err := when(a, "text")
			if err != nil {
				return nil, err
			}
			return humanRelative(t.Sub(o.now)), nil
		},
	})

	// ---- durations ----
	reg(tools.Func{
		Name: "duration_parse", Title: "Parse duration", Order: 8, Input: "text", Accepts: func(s string) bool { _, err := durationSeconds(s); return err == nil }, Keywords: []string{"duration", "seconds"},
		Doc:     "Parses a duration to seconds: '1h30m', '90 minutes', '1.5h', 'PT1H30M', 'P1DT2H', '1:30:00' (h:m:s) or '1:30' (m:s), or a bare number of seconds. A year counts 365 days and a month 30 days. Negative values are allowed.",
		Params:  []tools.Param{req("text", "str", "Duration text.")},
		Returns: "float", Example: `dt.duration_parse("1h30m")`,
		Tags: []string{"to seconds", "hours to seconds", "timespan", "iso 8601 duration", "convert duration", "human duration"},
		Fn: func(a tools.Args) (any, error) {
			v, err := durationSeconds(a.Str("text"))
			return v, err
		},
	})
	reg(tools.Func{
		Name: "duration_format", Title: "Format duration",
		Doc:     "Formats a number of seconds (decimals allowed) as compact (1h30m), long (1 hour, 30 minutes), iso8601 (PT1H30M) or clock (1:30:00). Days are the largest unit except in clock style.",
		Params:  []tools.Param{req("seconds", "float", "Duration in seconds."), p("style", "str", "compact, long, iso8601 or clock.", "compact")},
		Returns: "str", Example: `dt.duration_format(5400, "long")`,
		Tags: []string{"seconds to hours", "humanize duration", "pretty duration", "timespan", "hh:mm:ss"},
		Fn:   func(a tools.Args) (any, error) { return formatDuration(a.Float("seconds"), a.Str("style")) },
	})
	reg(tools.Func{
		Name: "iso8601_duration_to_seconds", Title: "ISO 8601 duration to seconds", Input: "text",
		Doc:     "Converts an ISO 8601 duration (PnYnMnWnDTnHnMnS, e.g. PT1H30M or P1DT12H) to seconds; years count 365 days and months 30 days. Errors on malformed input.",
		Params:  []tools.Param{req("text", "str", "ISO 8601 duration.")},
		Returns: "float", Example: `dt.iso8601_duration_to_seconds("PT1H30M")`,
		Tags: []string{"pt1h", "xml duration", "youtube duration", "period", "rfc 3339 duration"},
		Fn:   func(a tools.Args) (any, error) { return isoDurationSeconds(a.Str("text")) },
	})
	reg(tools.Func{
		Name: "seconds_to_iso8601_duration", Title: "Seconds to ISO 8601 duration",
		Doc:     "Converts seconds to an ISO 8601 duration such as PT1H30M or P1DT2H (days are the largest unit).",
		Params:  []tools.Param{req("seconds", "float", "Duration in seconds.")},
		Returns: "str", Example: `dt.seconds_to_iso8601_duration(5400)`,
		Tags: []string{"pt1h30m", "xml duration", "iso duration", "period"},
		Fn:   func(a tools.Args) (any, error) { return formatDuration(a.Float("seconds"), "iso8601") },
	})
	reg(tools.Func{
		Name: "sec_to_hms", Title: "Seconds to H:MM:SS",
		Doc:     "Converts seconds to H:MM:SS (hours may exceed 24; fractions kept, e.g. 3725 -> 1:02:05).",
		Params:  []tools.Param{req("seconds", "float", "Duration in seconds.")},
		Returns: "str", Example: `dt.sec_to_hms(3725)`,
		Tags: []string{"hh:mm:ss", "clock", "stopwatch", "timecode", "seconds to time"},
		Fn:   func(a tools.Args) (any, error) { return formatDuration(a.Float("seconds"), "clock") },
	})
	reg(tools.Func{
		Name: "hms_to_sec", Title: "H:MM:SS to seconds", Input: "text",
		Doc:     "Converts H:MM:SS (or MM:SS) to seconds; accepts decimals in the seconds and a leading minus.",
		Params:  []tools.Param{req("text", "str", "Clock duration such as 1:02:05 or 02:05.")},
		Returns: "float", Example: `dt.hms_to_sec("1:02:05")`,
		Tags: []string{"hh:mm:ss to seconds", "timecode", "clock to seconds"},
		Fn: func(a tools.Args) (any, error) {
			if !clockRe.MatchString(strings.TrimSpace(a.Str("text"))) {
				return nil, fmt.Errorf("%q is not H:MM:SS or MM:SS", a.Str("text"))
			}
			return durationSeconds(a.Str("text"))
		},
	})
	reg(tools.Func{
		Name: "time_math", Title: "Time arithmetic", Input: "expr", Keywords: []string{"math", "hours", "minutes"},
		Doc:     "Adds and subtracts times and durations: '1:30 + 2:45 - 0:20' gives 3:55. Terms are H:MM, H:MM:SS, unit forms like 1h30m or 45min, or ISO durations; the result is H:MM (or H:MM:SS when seconds remain), negative with a leading minus.",
		Params:  []tools.Param{req("expr", "str", "Expression with + and -.")},
		Returns: "str", Example: `dt.time_math("1:30 + 2:45 - 0:20")`,
		Tags: []string{"add hours and minutes", "timesheet", "work hours", "sum times", "hour calculator"},
		Fn:   func(a tools.Args) (any, error) { return timeMath(a.Str("expr")) },
	})

	// ---- epoch systems ----
	tzOnly := func(extra ...tools.Param) []tools.Param { return append(extra, tzP()) }
	timeIn := func(name, title, doc, ex string, tags []string, f func(t time.Time) (any, error), ret string) {
		reg(tools.Func{
			Name: name, Title: title, Input: "text", Doc: doc, Returns: ret, Example: ex, Tags: tags,
			Params: append([]tools.Param{p("text", "str", "Date/time text; empty means now.", "")}, tzP(), atP(), dayFirstP()),
			Fn: func(a tools.Args) (any, error) {
				t, _, err := when(a, "text")
				if err != nil {
					return nil, err
				}
				return f(t)
			},
		})
	}
	timeIn("julian_day", "Julian day number", "Returns the Julian day number (days since noon UTC on 4713 BC Jan 1, with fraction) of an instant; the Unix epoch is 2440587.5.",
		`dt.julian_day("2000-01-01 12:00", "UTC")`, []string{"jd", "astronomy", "julian date", "mjd"},
		func(t time.Time) (any, error) { return julianDay(t), nil }, "float")
	timeIn("excel_serial", "Excel serial date", "Returns the Excel/Lotus serial number of a date (1900 system, 1 = 1900-01-01) using the wall-clock time in tz. Excel wrongly treats 1900 as a leap year, so serials from 60 on are one higher than elapsed days; dates before 1900 are an error.",
		`dt.excel_serial("2024-01-31", "UTC")`, []string{"excel date", "spreadsheet date", "xlsx", "serial number", "1900 leap bug", "ole automation"},
		func(t time.Time) (any, error) { return excelSerial(t, false) }, "float")
	timeIn("filetime", "Windows FILETIME", "Returns the Windows FILETIME of an instant: 100-nanosecond intervals since 1601-01-01 UTC.",
		`dt.filetime("2024-01-31", "UTC")`, []string{"win32", "ntfs timestamp", "100 nanoseconds", "1601"},
		func(t time.Time) (any, error) { return ticksSince(t, epochFiletime, 100), nil }, "int")
	timeIn("dotnet_ticks", ".NET ticks", "Returns .NET DateTime ticks: 100-nanosecond intervals since 0001-01-01 UTC.",
		`dt.dotnet_ticks("2024-01-31", "UTC")`, []string{"csharp", "datetime.ticks", "100 nanoseconds", "vb.net"},
		func(t time.Time) (any, error) { return ticksSince(t, epochDotnet, 100), nil }, "int")
	timeIn("webkit_time", "WebKit/Chrome time", "Returns the WebKit (Chrome) timestamp: microseconds since 1601-01-01 UTC, as used in Chrome history and cookie databases.",
		`dt.webkit_time("2024-01-31", "UTC")`, []string{"chrome", "chromium", "cookies", "history", "sqlite"},
		func(t time.Time) (any, error) { return ticksSince(t, epochFiletime, 1000), nil }, "int")
	timeIn("mac_absolute_time", "Mac absolute time", "Returns Apple Core Foundation absolute time: seconds (with fraction) since 2001-01-01 UTC.",
		`dt.mac_absolute_time("2024-01-31", "UTC")`, []string{"cocoa", "cfabsolutetime", "ios", "nsdate", "apple"},
		func(t time.Time) (any, error) {
			return float64(t.Unix()-epochMac.Unix()) + float64(t.Nanosecond())/1e9, nil
		}, "float")
	timeIn("unix_hex", "Unix time in hex", "Returns Unix seconds as 0x-prefixed hexadecimal.",
		`dt.unix_hex("2024-01-31", "UTC")`, []string{"hex timestamp", "epoch hex", "base16"},
		func(t time.Time) (any, error) { return "0x" + strconv.FormatInt(t.Unix(), 16), nil }, "str")
	_ = tzOnly

	fromInt := func(name, title, doc, ex string, tags []string, f func(v int64, loc *time.Location) time.Time) {
		reg(tools.Func{
			Name: name, Title: title, Doc: doc, Returns: "str", Example: ex, Tags: tags,
			Params: []tools.Param{req("value", "int", "The numeric timestamp."), tzP()},
			Fn: func(a tools.Args) (any, error) {
				loc, err := loadLoc(a.Str("tz"))
				if err != nil {
					return nil, err
				}
				return rfc(f(a.Int("value"), loc)), nil
			},
		})
	}
	fromInt("from_filetime", "Windows FILETIME to date", "Converts a Windows FILETIME (100 ns intervals since 1601-01-01 UTC) to RFC 3339.",
		`dt.from_filetime(133511904000000000, "UTC")`, []string{"win32 timestamp", "ntfs", "1601"},
		func(v int64, l *time.Location) time.Time { return fromTicks(v, epochFiletime, 100).In(l) })
	fromInt("from_dotnet_ticks", ".NET ticks to date", "Converts .NET DateTime ticks (100 ns since 0001-01-01) to RFC 3339.",
		`dt.from_dotnet_ticks(638422464000000000, "UTC")`, []string{"csharp", "datetime from ticks"},
		func(v int64, l *time.Location) time.Time { return fromTicks(v, epochDotnet, 100).In(l) })
	fromInt("from_webkit_time", "WebKit/Chrome time to date", "Converts a WebKit/Chrome timestamp (microseconds since 1601-01-01) to RFC 3339.",
		`dt.from_webkit_time(13348377600000000, "UTC")`, []string{"chrome", "chromium", "cookies", "history"},
		func(v int64, l *time.Location) time.Time { return fromTicks(v, epochFiletime, 1000).In(l) })
	reg(tools.Func{
		Name: "from_mac_absolute_time", Title: "Mac absolute time to date",
		Doc:     "Converts Core Foundation absolute time (seconds since 2001-01-01 UTC, decimals allowed) to RFC 3339.",
		Params:  []tools.Param{req("value", "float", "Seconds since 2001-01-01."), tzP()},
		Returns: "str", Example: `dt.from_mac_absolute_time(728438400, "UTC")`,
		Tags: []string{"cocoa", "nsdate", "ios", "apple timestamp"},
		Fn: func(a tools.Args) (any, error) {
			loc, err := loadLoc(a.Str("tz"))
			if err != nil {
				return nil, err
			}
			v := a.Float("value")
			if math.Abs(v) > 1e12 {
				return nil, fmt.Errorf("value out of range")
			}
			s := math.Floor(v)
			return rfc(time.Unix(epochMac.Unix()+int64(s), int64(math.Round((v-s)*1e9))).In(loc)), nil
		},
	})
	reg(tools.Func{
		Name: "from_unix_hex", Title: "Hex Unix time to date", Input: "value",
		Doc:     "Converts hexadecimal Unix seconds (with or without 0x) to RFC 3339.",
		Params:  []tools.Param{req("value", "str", "Hex number such as 0x65b9e1a0."), tzP()},
		Returns: "str", Example: `dt.from_unix_hex("0x65b9e1a0", "UTC")`,
		Tags: []string{"hex timestamp", "epoch hex", "base16"},
		Fn: func(a tools.Args) (any, error) {
			loc, err := loadLoc(a.Str("tz"))
			if err != nil {
				return nil, err
			}
			s := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(a.Str("value"))), "0x")
			n, err := strconv.ParseInt(s, 16, 64)
			if err != nil {
				return nil, fmt.Errorf("%q is not a hex number", a.Str("value"))
			}
			return rfc(time.Unix(n, 0).In(loc)), nil
		},
	})
	reg(tools.Func{
		Name: "from_julian_day", Title: "Julian day to date",
		Doc:     "Converts a Julian day number (e.g. 2460341.5) to RFC 3339.",
		Params:  []tools.Param{req("jd", "float", "Julian day number."), tzP()},
		Returns: "str", Example: `dt.from_julian_day(2451545.0, "UTC")`,
		Tags: []string{"jd to date", "astronomy", "julian date"},
		Fn: func(a tools.Args) (any, error) {
			loc, err := loadLoc(a.Str("tz"))
			if err != nil {
				return nil, err
			}
			if math.Abs(a.Float("jd")) > 1e8 {
				return nil, fmt.Errorf("jd out of range")
			}
			return rfc(fromJulianDay(a.Float("jd")).In(loc)), nil
		},
	})
	reg(tools.Func{
		Name: "from_excel_serial", Title: "Excel serial to date",
		Doc:     "Converts an Excel serial date to RFC 3339 (wall clock in tz). Serial 60 (the non-existent 1900-02-29) is an error; date1904 selects the old Mac system.",
		Params:  []tools.Param{req("serial", "float", "Excel serial number, e.g. 45322.5."), p("date1904", "bool", "Use the 1904 date system.", false), tzP()},
		Returns: "str", Example: `dt.from_excel_serial(45322.5, tz="UTC")`,
		Tags: []string{"excel date", "spreadsheet", "xlsx", "serial to date", "ole automation"},
		Fn: func(a tools.Args) (any, error) {
			loc, err := loadLoc(a.Str("tz"))
			if err != nil {
				return nil, err
			}
			t, err := fromExcelSerial(a.Float("serial"), a.Bool("date1904"), loc)
			if err != nil {
				return nil, err
			}
			return rfc(t), nil
		},
	})

	// ---- cron ----
	reg(tools.Func{
		Name: "cron_next", Title: "Next cron runs", Input: "expr", Keywords: []string{"cron", "schedule"},
		Doc:     "Lists the next count (1..1000) run times of a 5-field cron expression (minute hour day-of-month month day-of-week) after from, as RFC 3339 in tz. Supports *, lists, ranges, steps, month/day names, @daily-style macros; day-of-month and day-of-week combine with OR when both are restricted.",
		Params:  []tools.Param{req("expr", "str", "Cron expression such as '*/15 9-17 * * mon-fri'."), p("count", "int", "How many runs.", int64(5)), p("from", "str", "Start instant (empty = now); runs are strictly after it.", ""), tzP()},
		Returns: "list[str]", Example: `dt.cron_next("0 9 * * mon-fri", 3, "2024-01-31", "UTC")`,
		Tags: []string{"crontab", "schedule", "next run", "cron preview", "cron tester", "job scheduler"},
		Fn: func(a tools.Args) (any, error) {
			o, err := opts(tools.Args{"tz": a.Str("tz"), "at": a.Str("from")})
			if err != nil {
				return nil, err
			}
			ts, err := cronNext(a.Str("expr"), int(a.Int("count")), o.now)
			if err != nil {
				return nil, err
			}
			out := make([]string, len(ts))
			for i, t := range ts {
				out[i] = rfc(t)
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "cron_explain", Title: "Explain cron", Input: "expr", Keywords: []string{"cron", "crontab"},
		Doc:     "Explains a 5-field cron expression in words, e.g. 'At 09:00, on Monday through Friday'. Errors on invalid fields.",
		Params:  []tools.Param{req("expr", "str", "Cron expression or @macro.")},
		Returns: "str", Example: `dt.cron_explain("*/15 9-17 * * mon-fri")`,
		Tags: []string{"crontab", "cron to english", "cron description", "schedule meaning", "decode cron"},
		Fn:   func(a tools.Args) (any, error) { return cronExplain(a.Str("expr")) },
	})
}

func yearArg(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nowFn().Year(), nil
	}
	y, err := strconv.Atoi(s)
	if err != nil || y < 1 || y > 9999 {
		return 0, fmt.Errorf("year must be a number 1..9999, got %q", s)
	}
	return y, nil
}

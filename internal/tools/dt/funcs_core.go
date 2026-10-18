package dt

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/avestura/zagros/internal/tools"
)

// nowFn is the clock; tests may replace it.
var nowFn = time.Now

var dtKeywords = []string{"dt", "date", "time", "ts", "now", "tz", "unix", "epoch", "cron", "age", "diff"}

func p(name, typ, doc string, def any) tools.Param {
	return tools.Param{Name: name, Type: typ, Doc: doc, Default: def}
}

func req(name, typ, doc string) tools.Param {
	return tools.Param{Name: name, Type: typ, Doc: doc, Required: true}
}

func tzP() tools.Param {
	return p("tz", "str", "Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.", "")
}

func atP() tools.Param {
	return p("at", "str", "Reference instant for relative words and countdowns (anything parse() understands); empty means now.", "")
}

func dayFirstP() tools.Param {
	return p("day_first", "bool", "Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.", false)
}

func reg(f tools.Func) {
	f.Group = "dt"
	if f.Order == 0 {
		f.Order = 100 // unset: behind the headline converters in the launcher
	}
	if f.Input != "" {
		f.Family = "dt"
		f.Keywords = append(append([]string(nil), dtKeywords...), f.Keywords...)
	}
	tools.Register(f)
}

func rfc(t time.Time) string { return t.Format(time.RFC3339Nano) }

// opts builds parse options from tz, at and day_first arguments.
func opts(a tools.Args) (parseOpts, error) {
	loc, err := loadLoc(a.Str("tz"))
	if err != nil {
		return parseOpts{}, err
	}
	real := nowFn().In(loc)
	o := parseOpts{loc: loc, now: real}
	if _, ok := a["day_first"]; ok {
		o.dayFirst = a.Bool("day_first")
	}
	if at := strings.TrimSpace(a.Str("at")); at != "" {
		t, err := parseAny(at, o)
		if err != nil {
			return o, fmt.Errorf("at: %w", err)
		}
		o.now = t.In(loc)
	}
	return o, nil
}

// when parses the text param (empty or "now" mean the reference instant).
func when(a tools.Args, name string) (time.Time, parseOpts, error) {
	o, err := opts(a)
	if err != nil {
		return time.Time{}, o, err
	}
	s := strings.TrimSpace(a.Str(name))
	if s == "" {
		return o.now, o, nil
	}
	t, err := parseAny(s, o)
	return t, o, err
}

func acceptsAny(string) bool { return true }

func describe(t, now time.Time) map[string]string {
	u := t.UTC()
	iy, iw := t.ISOWeek()
	_, off := t.Zone()
	xs, _ := excelSerial(t, false)
	m := map[string]string{
		"iso":          rfc(t),
		"utc":          rfc(u),
		"local":        t.Format("2006-01-02 15:04:05 MST"),
		"unix":         strconv.FormatInt(t.Unix(), 10),
		"unix_ms":      strconv.FormatInt(t.UnixMilli(), 10),
		"unix_us":      strconv.FormatInt(t.UnixMicro(), 10),
		"unix_ns":      strconv.FormatInt(t.UnixNano(), 10),
		"unix_hex":     "0x" + strconv.FormatInt(t.Unix(), 16),
		"rfc1123":      u.Format("Mon, 02 Jan 2006 15:04:05 GMT"),
		"rfc2822":      t.Format("Mon, 02 Jan 2006 15:04:05 -0700"),
		"relative":     humanRelative(t.Sub(now)),
		"weekday":      t.Weekday().String(),
		"iso_week":     fmt.Sprintf("%04d-W%02d-%d", iy, iw, (int(t.Weekday())+6)%7+1),
		"day_of_year":  strconv.Itoa(t.YearDay()),
		"quarter":      strconv.Itoa((int(t.Month())-1)/3 + 1),
		"leap_year":    strconv.FormatBool(isLeap(t.Year())),
		"offset":       fmtOffset(off, true),
		"julian_day":   strconv.FormatFloat(julianDay(t), 'f', 6, 64),
		"mjd":          strconv.FormatFloat(julianDay(t)-2400000.5, 'f', 6, 64),
		"excel_serial": strconv.FormatFloat(xs, 'f', 6, 64),
		"filetime":     strconv.FormatInt(ticksSince(t, epochFiletime, 100), 10),
		"dotnet_ticks": strconv.FormatInt(ticksSince(t, epochDotnet, 100), 10),
		"webkit":       strconv.FormatInt(ticksSince(t, epochFiletime, 1000), 10),
		"mac_absolute": strconv.FormatFloat(float64(t.Unix()-epochMac.Unix())+float64(t.Nanosecond())/1e9, 'f', 6, 64),
	}
	if t.Year() < 1900 {
		delete(m, "excel_serial")
	}
	return m
}

func init() {
	reg(tools.Func{
		Name: "now", Title: "Current time", Order: 20, Input: "tz", Keywords: []string{"now", "current"},
		Doc:     "Returns the current time in a zone, formatted with a Go layout, strftime tokens (%Y-%m-%d %H:%M) or a layout name (rfc3339, rfc2822, date, time, unix). Default format is RFC 3339.",
		Params:  []tools.Param{tzP(), p("format", "str", "Go layout, strftime string or layout name.", "rfc3339")},
		Returns: "str", Example: `dt.now("Asia/Tokyo", "%Y-%m-%d %H:%M")`,
		Tags: []string{"current time", "clock", "today", "timestamp", "what time is it", "date now"},
		Fn: func(a tools.Args) (any, error) {
			loc, err := loadLoc(a.Str("tz"))
			if err != nil {
				return nil, err
			}
			return formatTime(nowFn().In(loc), a.Str("format"))
		},
	})
	reg(tools.Func{
		Name: "now_all", Title: "Now in every format", Order: 2, Input: "tz", Keywords: []string{"now", "current"},
		Doc:     "Returns the current moment in many representations at once: ISO 8601, RFC 2822, Unix seconds/ms, UTC, local, ISO week, day of year, quarter, Julian day and more.",
		Params:  []tools.Param{tzP()},
		Returns: "dict[str, str]", Example: `dt.now_all("UTC")`,
		Tags: []string{"current time", "all formats", "timestamp", "iso week", "epoch", "clock"},
		Fn: func(a tools.Args) (any, error) {
			loc, err := loadLoc(a.Str("tz"))
			if err != nil {
				return nil, err
			}
			t := nowFn().In(loc)
			m := describe(t, t)
			m["tz"] = t.Location().String()
			return m, nil
		},
	})
	reg(tools.Func{
		Name: "unix_now", Title: "Unix time now",
		Doc:     "Returns the current Unix time in seconds.",
		Returns: "int", Example: `dt.unix_now()`,
		Tags: []string{"epoch", "timestamp", "seconds since 1970", "current unix time"},
		Fn:   func(a tools.Args) (any, error) { return nowFn().Unix(), nil },
	})
	reg(tools.Func{
		Name: "unix_ms_now", Title: "Unix milliseconds now",
		Doc:     "Returns the current Unix time in milliseconds (JavaScript Date.now()).",
		Returns: "int", Example: `dt.unix_ms_now()`,
		Tags: []string{"epoch", "timestamp", "milliseconds", "date.now", "javascript"},
		Fn:   func(a tools.Args) (any, error) { return nowFn().UnixMilli(), nil },
	})
	reg(tools.Func{
		Name: "parse", Title: "Parse date/time", Order: 3, Input: "text", Accepts: acceptsAny, Keywords: []string{"parse"},
		Doc:     "Parses almost any date text and returns RFC 3339: ISO 8601, RFC 2822, '2024-01-31 15:04', '31/01/2024', 'Jan 2, 2006', Unix seconds/ms/us/ns (guessed by magnitude), today/tomorrow/yesterday, 'in 3 days', '3 weeks ago', 'next friday', 'monday 9am'. Ambiguous numeric dates are month/day unless day_first is set. Errors when nothing fits.",
		Params:  []tools.Param{req("text", "str", "Date/time text."), tzP(), atP(), dayFirstP()},
		Returns: "str", Example: `dt.parse("next friday 5pm", "Europe/Paris")`,
		Tags: []string{"date parser", "natural language date", "iso 8601", "rfc 3339", "relative date", "str to date", "strtotime", "detect format"},
		Fn: func(a tools.Args) (any, error) {
			o, err := opts(a)
			if err != nil {
				return nil, err
			}
			t, err := parseAny(a.Str("text"), o)
			if err != nil {
				return nil, err
			}
			return rfc(t), nil
		},
	})
	reg(tools.Func{
		Name: "is_valid", Title: "Is it a date?", Input: "text",
		Doc:     "Returns true when parse() can read the text as a date or time.",
		Params:  []tools.Param{req("text", "str", "Date/time text."), tzP(), dayFirstP()},
		Returns: "bool", Example: `dt.is_valid("2024-02-30")`,
		Tags: []string{"validate date", "check date", "valid date", "real date"},
		Fn: func(a tools.Args) (any, error) {
			o, err := opts(a)
			if err != nil {
				return nil, err
			}
			_, err = parseAny(a.Str("text"), o)
			return err == nil, nil
		},
	})
	reg(tools.Func{
		Name: "convert", Title: "Date/time converter", Order: 1, Input: "text", Accepts: acceptsAny, Keywords: []string{"convert", "epoch", "timestamp"},
		Doc:     "The all-in-one converter: reads any date/time/timestamp (empty text means now) and returns the instant as Unix s/ms/us/ns, ISO UTC, RFC 1123/2822, local, relative time, weekday, ISO week, day of year, quarter, Julian day, Excel serial, Windows FILETIME, .NET ticks, WebKit/Chrome time, Mac absolute time and Unix hex.",
		Params:  []tools.Param{p("text", "str", "Date/time text or a numeric timestamp; empty means now.", ""), tzP(), atP(), dayFirstP()},
		Returns: "dict[str, str]", Example: `dt.convert("1700000000")`,
		Tags: []string{"timestamp converter", "epoch converter", "unix time", "filetime", "julian", "excel date", "webkit", "ticks", "all formats", "unix to date", "date to unix"},
		Fn: func(a tools.Args) (any, error) {
			t, o, err := when(a, "text")
			if err != nil {
				return nil, err
			}
			return describe(t, o.now), nil
		},
	})
	reg(tools.Func{
		Name: "from_unix", Title: "Unix timestamp to date", Input: "value", Keywords: []string{"unix", "epoch"},
		Doc:     "Converts a Unix timestamp (integer or decimal, or 0x hex) to RFC 3339 in a zone. Unit auto guesses seconds, milliseconds, microseconds or nanoseconds by magnitude; set s, ms, us or ns to force it.",
		Params:  []tools.Param{req("value", "str", "Timestamp such as 1700000000 or 1700000000123."), p("unit", "str", "auto, s, ms, us or ns.", "auto"), tzP()},
		Returns: "str", Example: `dt.from_unix("1700000000", "s", "UTC")`,
		Tags: []string{"epoch to date", "timestamp to date", "unix time converter", "milliseconds", "human readable"},
		Fn: func(a tools.Args) (any, error) {
			loc, err := loadLoc(a.Str("tz"))
			if err != nil {
				return nil, err
			}
			t, _, err := fromUnixAuto(a.Str("value"), a.Str("unit"), loc)
			if err != nil {
				return nil, err
			}
			return rfc(t), nil
		},
	})
	reg(tools.Func{
		Name: "to_unix", Title: "Date to Unix timestamp", Input: "text", Keywords: []string{"unix", "epoch"},
		Doc:     "Converts any date text to a Unix timestamp in seconds (default), ms, us or ns, rounded toward negative infinity. Dates before 1970 give negative values.",
		Params:  []tools.Param{req("text", "str", "Date/time text."), p("unit", "str", "s, ms, us or ns.", "s"), tzP(), atP(), dayFirstP()},
		Returns: "int", Example: `dt.to_unix("2024-01-31 15:04", "s", "UTC")`,
		Tags: []string{"date to epoch", "date to timestamp", "seconds since 1970", "strtotime"},
		Fn: func(a tools.Args) (any, error) {
			t, _, err := when(a, "text")
			if err != nil {
				return nil, err
			}
			switch strings.ToLower(a.Str("unit")) {
			case "s", "", "sec", "seconds":
				return t.Unix(), nil
			case "ms":
				return t.UnixMilli(), nil
			case "us":
				return t.UnixMicro(), nil
			case "ns":
				return t.UnixNano(), nil
			}
			return nil, fmt.Errorf("unit must be s, ms, us or ns")
		},
	})
	reg(tools.Func{
		Name: "convert_tz", Title: "Convert time zone", Order: 4, Input: "text", Keywords: []string{"convert", "timezone", "zone"},
		Doc:     "Converts a date/time from one time zone to another. The text is read in from_tz unless it carries its own zone; zones may be IANA names, cities, abbreviations (EST, PST) or offsets (UTC+3:30). Format is a Go layout, strftime string or name (default RFC 3339).",
		Params:  []tools.Param{p("text", "str", "Date/time text; empty means now.", ""), p("from_tz", "str", "Zone the text is in (empty = local).", ""), p("to_tz", "str", "Target zone.", "UTC"), p("format", "str", "Output layout.", "rfc3339"), atP()},
		Returns: "str", Example: `dt.convert_tz("2024-01-31 15:00", "America/New_York", "Asia/Tehran")`,
		Tags: []string{"timezone converter", "time zone", "est to utc", "pst", "offset", "world time", "meeting time"},
		Fn: func(a tools.Args) (any, error) {
			from, err := loadLoc(a.Str("from_tz"))
			if err != nil {
				return nil, err
			}
			to, err := loadLoc(a.Str("to_tz"))
			if err != nil {
				return nil, err
			}
			args := tools.Args{"tz": a.Str("from_tz"), "at": a.Str("at")}
			_ = from
			t, _, err := when(withText(args, a.Str("text")), "text")
			if err != nil {
				return nil, err
			}
			return formatTime(t.In(to), a.Str("format"))
		},
	})
	reg(tools.Func{
		Name: "tz_info", Title: "Time zone info", Input: "name", Keywords: []string{"timezone", "dst", "offset"},
		Doc:     "Describes a time zone at an instant: UTC offset, abbreviation, whether DST is active, the local time there, and the start and end of the current offset period (the next DST change).",
		Params:  []tools.Param{req("name", "str", "IANA name, city, abbreviation or offset."), p("at", "str", "Instant to inspect (empty = now).", "")},
		Returns: "dict[str, str]", Example: `dt.tz_info("Europe/Paris", "2024-07-01")`,
		Tags: []string{"timezone", "daylight saving", "utc offset", "dst", "zone abbreviation", "gmt offset"},
		Fn: func(a tools.Args) (any, error) {
			loc, err := loadLoc(a.Str("name"))
			if err != nil {
				return nil, err
			}
			o := parseOpts{loc: loc, now: nowFn().In(loc)}
			t := o.now
			if s := strings.TrimSpace(a.Str("at")); s != "" {
				if t, err = parseAny(s, o); err != nil {
					return nil, err
				}
				t = t.In(loc)
			}
			abbr, off := t.Zone()
			m := map[string]string{
				"name": loc.String(), "offset": fmtOffset(off, true), "offset_seconds": strconv.Itoa(off),
				"abbreviation": abbr, "is_dst": strconv.FormatBool(t.IsDST()), "time": t.Format("2006-01-02 15:04:05 MST"),
			}
			start, end := t.ZoneBounds()
			if !start.IsZero() {
				m["period_start"] = rfc(start)
			}
			if !end.IsZero() {
				m["next_change"] = rfc(end)
			}
			return m, nil
		},
	})
	reg(tools.Func{
		Name: "tz_list", Title: "List time zones", Input: "filter", Keywords: []string{"timezones", "zones"},
		Doc:     "Lists the known IANA time zone names, optionally keeping those that contain filter (case-insensitive, spaces match underscores).",
		Params:  []tools.Param{p("filter", "str", "Substring such as asia or new_york; empty lists all.", "")},
		Returns: "list[str]", Example: `dt.tz_list("america/arg")`,
		Tags: []string{"iana", "tz database", "zone names", "olson", "find timezone"},
		Fn:   func(a tools.Args) (any, error) { return filterZones(a.Str("filter")), nil },
	})
	reg(tools.Func{
		Name: "tz_offset", Title: "UTC offset of a zone", Input: "name",
		Doc:     "Returns the UTC offset of a zone at an instant as +hh:mm (DST-aware, e.g. +01:00 or +02:00 for Europe/Paris depending on the date).",
		Params:  []tools.Param{req("name", "str", "IANA name, city, abbreviation or offset."), p("at", "str", "Instant (empty = now).", "")},
		Returns: "str", Example: `dt.tz_offset("Asia/Kolkata")`,
		Tags: []string{"gmt offset", "utc offset", "hours from utc", "timezone difference"},
		Fn: func(a tools.Args) (any, error) {
			loc, err := loadLoc(a.Str("name"))
			if err != nil {
				return nil, err
			}
			o := parseOpts{loc: loc, now: nowFn().In(loc)}
			t := o.now
			if s := strings.TrimSpace(a.Str("at")); s != "" {
				if t, err = parseAny(s, o); err != nil {
					return nil, err
				}
			}
			_, off := t.In(loc).Zone()
			return fmtOffset(off, true), nil
		},
	})
	defaultZones := []string{"UTC", "America/Los_Angeles", "America/New_York", "Europe/London", "Europe/Berlin", "Asia/Dubai", "Asia/Kolkata", "Asia/Tokyo", "Australia/Sydney"}
	reg(tools.Func{
		Name: "world_clock", Title: "World clock", Input: "text", Keywords: []string{"world", "clock", "zones"},
		Doc:     "Shows one instant in several time zones at once (default: a spread of major cities). The text is read in tz (default local); empty means now. Returns a dict zone -> 'YYYY-MM-DD HH:MM ZONE (Weekday)'.",
		Params:  []tools.Param{p("text", "str", "Date/time text; empty means now.", ""), p("zones", "list[str]", "Zones to show; empty uses a default set.", []string(nil)), tzP()},
		Returns: "dict[str, str]", Example: `dt.world_clock("2024-01-31 15:00", ["UTC", "Tokyo", "New York"], "Europe/Paris")`,
		Tags: []string{"time zones", "world time", "multiple timezones", "meeting", "what time is it in"},
		Fn: func(a tools.Args) (any, error) {
			t, _, err := when(a, "text")
			if err != nil {
				return nil, err
			}
			zs := a.List("zones")
			if len(zs) == 0 {
				zs = defaultZones
			}
			out := map[string]string{}
			for _, z := range zs {
				loc, err := loadLoc(z)
				if err != nil {
					return nil, err
				}
				out[z] = t.In(loc).Format("2006-01-02 15:04 MST (Mon)")
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "meeting_planner", Title: "Meeting planner", Input: "time", Keywords: []string{"meeting", "planner"},
		Doc:     "Plans a meeting: takes a time of day (HH:MM) and date in the first zone and shows the local time in every zone with a note: working hours (09-17), early/late (07-09, 17-21) or night.",
		Params:  []tools.Param{req("zones", "list[str]", "Zones; the first one is where the time is given."), p("time", "str", "Time of day in the first zone, HH:MM.", "09:00"), p("date", "str", "Date in the first zone; empty means today there.", "")},
		Returns: "dict[str, str]", Example: `dt.meeting_planner(["Europe/Paris", "America/New_York", "Asia/Tokyo"], "15:00", "2024-03-04")`,
		Tags: []string{"schedule", "overlap", "working hours", "time zone meeting", "find a time"},
		Fn: func(a tools.Args) (any, error) {
			zs := a.List("zones")
			if len(zs) == 0 {
				return nil, fmt.Errorf("zones must list at least one zone")
			}
			base, err := loadLoc(zs[0])
			if err != nil {
				return nil, err
			}
			v, ok := parseTOD(strings.TrimSpace(a.Str("time")))
			if !ok {
				return nil, fmt.Errorf("time must look like 14:30 or 9am")
			}
			o := parseOpts{loc: base, now: nowFn().In(base)}
			day := o.now
			if s := strings.TrimSpace(a.Str("date")); s != "" {
				if day, err = parseAny(s, o); err != nil {
					return nil, err
				}
			}
			start := atTOD(day.In(base), v)
			out := map[string]string{}
			for _, z := range zs {
				loc, err := loadLoc(z)
				if err != nil {
					return nil, err
				}
				l := start.In(loc)
				h := l.Hour()
				note := "night"
				switch {
				case h >= 9 && h < 17:
					note = "working hours"
				case (h >= 7 && h < 9) || (h >= 17 && h < 21):
					note = "early/late"
				}
				out[z] = l.Format("2006-01-02 15:04 MST (Mon)") + " - " + note
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "format", Title: "Format date", Input: "text", Keywords: []string{"format"},
		Doc:     "Formats a date/time with a Go layout (2006-01-02 15:04:05), strftime tokens (%Y-%m-%d %H:%M:%S %A %B %j %U %Z %z, %-d for no padding) or a name (rfc3339, rfc2822, date, datetime, kitchen, unix...). Empty text means now.",
		Params:  []tools.Param{p("text", "str", "Date/time text; empty means now.", ""), p("layout", "str", "Go layout, strftime string or layout name.", "rfc3339"), tzP(), atP(), dayFirstP()},
		Returns: "str", Example: `dt.format("2024-01-31", "%A, %B %-d, %Y")`,
		Tags: []string{"strftime", "date format", "go layout", "date to string", "pattern", "moment format"},
		Fn: func(a tools.Args) (any, error) {
			t, _, err := when(a, "text")
			if err != nil {
				return nil, err
			}
			return formatTime(t, a.Str("layout"))
		},
	})
	reg(tools.Func{
		Name: "layouts", Title: "Format cheatsheet",
		Doc:     "Returns a cheatsheet dict of Go layout elements, strftime tokens and layout names accepted by format() and parse_with().",
		Returns: "dict[str, str]", Example: `dt.layouts()`,
		Tags: []string{"strftime reference", "go time layout", "date format help", "cheat sheet", "tokens"},
		Fn:   func(a tools.Args) (any, error) { return layoutsCheatsheet(), nil },
	})
	reg(tools.Func{
		Name: "parse_with", Title: "Parse with layout", Input: "text",
		Doc:     "Parses text with an explicit layout (Go layout, strftime tokens like %d/%m/%Y %H:%M, or a name) and returns RFC 3339. Use it when parse() would guess wrong; errors when the text does not match.",
		Params:  []tools.Param{req("text", "str", "Date text."), req("layout", "str", "Layout describing the text."), tzP()},
		Returns: "str", Example: `dt.parse_with("31/01/2024 15:04", "%d/%m/%Y %H:%M", "UTC")`,
		Tags: []string{"strptime", "custom date format", "exact format", "parse date with pattern"},
		Fn: func(a tools.Args) (any, error) {
			loc, err := loadLoc(a.Str("tz"))
			if err != nil {
				return nil, err
			}
			t, err := parseWith(a.Str("text"), a.Str("layout"), loc)
			if err != nil {
				return nil, err
			}
			return rfc(t), nil
		},
	})
}

// withText returns a copy of args carrying text.
func withText(a tools.Args, text string) tools.Args {
	b := tools.Args{}
	for k, v := range a {
		b[k] = v
	}
	b["text"] = text
	return b
}

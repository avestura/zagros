package dt

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

var namedLayouts = map[string]string{
	"rfc3339": time.RFC3339, "rfc3339nano": time.RFC3339Nano, "iso": time.RFC3339, "iso8601": time.RFC3339,
	"rfc2822": "Mon, 02 Jan 2006 15:04:05 -0700", "rfc1123": time.RFC1123, "rfc1123z": time.RFC1123Z,
	"rfc822": time.RFC822, "rfc822z": time.RFC822Z, "rfc850": time.RFC850, "http": "Mon, 02 Jan 2006 15:04:05 GMT",
	"ansic": time.ANSIC, "unixdate": time.UnixDate, "rubydate": time.RubyDate, "kitchen": time.Kitchen, "stamp": time.Stamp,
	"date": "2006-01-02", "time": "15:04:05", "datetime": "2006-01-02 15:04:05", "sql": "2006-01-02 15:04:05",
	"us": "01/02/2006", "eu": "02/01/2006", "long": "Monday, January 2, 2006 3:04 PM", "compact": "20060102T150405Z0700",
}

func isoWeekYear(t time.Time) (int, int) { return t.ISOWeek() }

func weekNumber(t time.Time, firstDay time.Weekday) int {
	// week 0 = days before the first firstDay of the year (strftime %U / %W)
	yday := t.YearDay() - 1
	wd := (int(t.Weekday()) - int(firstDay) + 7) % 7
	return (yday + 7 - wd) / 7
}

// formatTime renders t with a Go layout, a named layout or strftime tokens.
func formatTime(t time.Time, layout string) (string, error) {
	if l, ok := namedLayouts[strings.ToLower(strings.TrimSpace(layout))]; ok {
		return t.Format(l), nil
	}
	switch strings.ToLower(strings.TrimSpace(layout)) {
	case "unix":
		return strconv.FormatInt(t.Unix(), 10), nil
	case "unix_ms":
		return strconv.FormatInt(t.UnixMilli(), 10), nil
	}
	if !strings.Contains(layout, "%") {
		return t.Format(layout), nil
	}
	return strftime(t, layout)
}

func strftime(t time.Time, f string) (string, error) {
	var b strings.Builder
	rs := []rune(f)
	for i := 0; i < len(rs); i++ {
		if rs[i] != '%' {
			b.WriteRune(rs[i])
			continue
		}
		i++
		if i >= len(rs) {
			return "", fmt.Errorf("layout ends with a lone %%")
		}
		pad := true
		colon := false
		for i < len(rs) && (rs[i] == '-' || rs[i] == ':' || rs[i] == '_' || rs[i] == '^') {
			switch rs[i] {
			case '-':
				pad = false
			case ':':
				colon = true
			}
			i++
		}
		if i >= len(rs) {
			return "", fmt.Errorf("layout ends inside a %% token")
		}
		num := func(v, w int) string {
			if !pad {
				return strconv.Itoa(v)
			}
			return fmt.Sprintf("%0*d", w, v)
		}
		sp := func(v int) string {
			if !pad {
				return strconv.Itoa(v)
			}
			return fmt.Sprintf("%2d", v)
		}
		h12 := t.Hour() % 12
		if h12 == 0 {
			h12 = 12
		}
		switch c := rs[i]; c {
		case 'Y':
			b.WriteString(strconv.Itoa(t.Year()))
		case 'y':
			b.WriteString(num(t.Year()%100, 2))
		case 'C':
			b.WriteString(num(t.Year()/100, 2))
		case 'm':
			b.WriteString(num(int(t.Month()), 2))
		case 'd':
			b.WriteString(num(t.Day(), 2))
		case 'e':
			b.WriteString(sp(t.Day()))
		case 'H':
			b.WriteString(num(t.Hour(), 2))
		case 'k':
			b.WriteString(sp(t.Hour()))
		case 'I':
			b.WriteString(num(h12, 2))
		case 'l':
			b.WriteString(sp(h12))
		case 'M':
			b.WriteString(num(t.Minute(), 2))
		case 'S':
			b.WriteString(num(t.Second(), 2))
		case 'f':
			fmt.Fprintf(&b, "%06d", t.Nanosecond()/1000)
		case 'L':
			fmt.Fprintf(&b, "%03d", t.Nanosecond()/1e6)
		case 'N':
			fmt.Fprintf(&b, "%09d", t.Nanosecond())
		case 'p':
			b.WriteString(t.Format("PM"))
		case 'P':
			b.WriteString(strings.ToLower(t.Format("PM")))
		case 'A':
			b.WriteString(t.Weekday().String())
		case 'a':
			b.WriteString(t.Weekday().String()[:3])
		case 'B':
			b.WriteString(t.Month().String())
		case 'b', 'h':
			b.WriteString(t.Month().String()[:3])
		case 'j':
			b.WriteString(num(t.YearDay(), 3))
		case 'U':
			b.WriteString(num(weekNumber(t, time.Sunday), 2))
		case 'W':
			b.WriteString(num(weekNumber(t, time.Monday), 2))
		case 'V':
			_, w := t.ISOWeek()
			b.WriteString(num(w, 2))
		case 'G':
			y, _ := t.ISOWeek()
			b.WriteString(strconv.Itoa(y))
		case 'g':
			y, _ := t.ISOWeek()
			b.WriteString(num(y%100, 2))
		case 'u':
			b.WriteString(strconv.Itoa((int(t.Weekday())+6)%7 + 1))
		case 'w':
			b.WriteString(strconv.Itoa(int(t.Weekday())))
		case 'Z':
			b.WriteString(t.Format("MST"))
		case 'z':
			_, off := t.Zone()
			b.WriteString(fmtOffset(off, colon))
		case 's':
			b.WriteString(strconv.FormatInt(t.Unix(), 10))
		case 'q':
			b.WriteString(strconv.Itoa((int(t.Month())-1)/3 + 1))
		case 'n':
			b.WriteByte('\n')
		case 't':
			b.WriteByte('\t')
		case '%':
			b.WriteByte('%')
		case 'D', 'x':
			b.WriteString(t.Format("01/02/06"))
		case 'F':
			b.WriteString(t.Format("2006-01-02"))
		case 'T', 'X':
			b.WriteString(t.Format("15:04:05"))
		case 'R':
			b.WriteString(t.Format("15:04"))
		case 'r':
			b.WriteString(t.Format("03:04:05 PM"))
		case 'c':
			b.WriteString(t.Format("Mon Jan _2 15:04:05 2006"))
		default:
			return "", fmt.Errorf("unsupported strftime token %%%c", c)
		}
	}
	return b.String(), nil
}

// strftimeToGo converts a strftime layout to a Go layout for parsing.
func strftimeToGo(f string) (string, error) {
	var b strings.Builder
	rs := []rune(f)
	for i := 0; i < len(rs); i++ {
		if rs[i] != '%' {
			// escape nothing: Go layouts treat letters like Jan/Mon specially, so
			// only pass through punctuation and digits that cannot clash
			b.WriteRune(rs[i])
			continue
		}
		i++
		if i >= len(rs) {
			return "", fmt.Errorf("layout ends with a lone %%")
		}
		nopad, colon := false, false
		for i < len(rs) && (rs[i] == '-' || rs[i] == ':') {
			if rs[i] == '-' {
				nopad = true
			} else {
				colon = true
			}
			i++
		}
		if i >= len(rs) {
			return "", fmt.Errorf("layout ends inside a %% token")
		}
		pick := func(padded, plain string) {
			if nopad {
				b.WriteString(plain)
			} else {
				b.WriteString(padded)
			}
		}
		switch c := rs[i]; c {
		case 'Y':
			b.WriteString("2006")
		case 'y':
			b.WriteString("06")
		case 'm':
			pick("01", "1")
		case 'd':
			pick("02", "2")
		case 'e':
			b.WriteString("_2")
		case 'H':
			b.WriteString("15")
		case 'I':
			pick("03", "3")
		case 'l':
			b.WriteString("3")
		case 'M':
			pick("04", "4")
		case 'S':
			pick("05", "5")
		case 'f':
			b.WriteString(".000000")
		case 'L':
			b.WriteString(".000")
		case 'N':
			b.WriteString(".000000000")
		case 'p', 'P':
			b.WriteString("PM")
		case 'A':
			b.WriteString("Monday")
		case 'a':
			b.WriteString("Mon")
		case 'B':
			b.WriteString("January")
		case 'b', 'h':
			b.WriteString("Jan")
		case 'j':
			b.WriteString("002")
		case 'Z':
			b.WriteString("MST")
		case 'z':
			if colon {
				b.WriteString("-07:00")
			} else {
				b.WriteString("-0700")
			}
		case '%':
			b.WriteByte('%')
		case 'F':
			b.WriteString("2006-01-02")
		case 'T':
			b.WriteString("15:04:05")
		case 'R':
			b.WriteString("15:04")
		case 'D':
			b.WriteString("01/02/06")
		default:
			return "", fmt.Errorf("strftime token %%%c cannot be used for parsing", c)
		}
	}
	return b.String(), nil
}

// parseWith parses text with a Go layout, named layout or strftime layout.
func parseWith(text, layout string, loc *time.Location) (time.Time, error) {
	l := strings.TrimSpace(layout)
	if l == "%s" || strings.EqualFold(l, "unix") {
		t, _, err := fromUnixAuto(strings.TrimSpace(text), "s", loc)
		return t, err
	}
	if n, ok := namedLayouts[strings.ToLower(l)]; ok {
		l = n
	} else if strings.Contains(layout, "%") {
		g, err := strftimeToGo(layout)
		if err != nil {
			return time.Time{}, err
		}
		l = g
	}
	t, err := time.ParseInLocation(l, strings.TrimSpace(text), loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("text does not match layout %q: %w", layout, err)
	}
	return t, nil
}

func layoutsCheatsheet() map[string]string {
	return map[string]string{
		"go 2006":              "4-digit year",
		"go 06":                "2-digit year",
		"go 01":                "month 01-12",
		"go 1":                 "month 1-12",
		"go Jan":               "month abbreviation",
		"go January":           "month name",
		"go 02":                "day of month 01-31",
		"go _2":                "day of month, space padded",
		"go 2":                 "day of month 1-31",
		"go 002":               "day of year 001-366",
		"go Mon":               "weekday abbreviation",
		"go Monday":            "weekday name",
		"go 15":                "hour 00-23",
		"go 03":                "hour 01-12",
		"go 3":                 "hour 1-12",
		"go 04":                "minute 00-59",
		"go 05":                "second 00-59",
		"go PM":                "AM/PM marker (pm for lowercase)",
		"go .000":              "fractional seconds, fixed digits (.999 trims zeros)",
		"go MST":               "zone abbreviation",
		"go -0700":             "zone offset +hhmm",
		"go -07:00":            "zone offset +hh:mm",
		"go Z07:00":            "Z for UTC, else +hh:mm",
		"strftime %Y":          "4-digit year",
		"strftime %y":          "2-digit year",
		"strftime %m":          "month 01-12",
		"strftime %B":          "month name",
		"strftime %b":          "month abbreviation",
		"strftime %d":          "day of month 01-31",
		"strftime %e":          "day of month, space padded",
		"strftime %j":          "day of year 001-366",
		"strftime %A":          "weekday name",
		"strftime %a":          "weekday abbreviation",
		"strftime %u":          "ISO weekday 1-7 (Monday=1)",
		"strftime %w":          "weekday 0-6 (Sunday=0)",
		"strftime %U":          "week of year, Sunday first, 00-53",
		"strftime %W":          "week of year, Monday first, 00-53",
		"strftime %V":          "ISO week number 01-53",
		"strftime %G":          "ISO week-based year",
		"strftime %H":          "hour 00-23",
		"strftime %I":          "hour 01-12",
		"strftime %M":          "minute",
		"strftime %S":          "second",
		"strftime %f":          "microseconds (6 digits)",
		"strftime %L":          "milliseconds (3 digits)",
		"strftime %N":          "nanoseconds (9 digits)",
		"strftime %p":          "AM or PM",
		"strftime %Z":          "zone abbreviation",
		"strftime %z":          "zone offset +hhmm (%:z gives +hh:mm)",
		"strftime %s":          "Unix seconds",
		"strftime %q":          "quarter 1-4",
		"strftime %F":          "%Y-%m-%d",
		"strftime %T":          "%H:%M:%S",
		"strftime %D":          "%m/%d/%y",
		"strftime %R":          "%H:%M",
		"strftime %-d":         "a leading - drops zero padding (%-d, %-m, %-H)",
		"named rfc3339":        "2006-01-02T15:04:05Z07:00",
		"named rfc2822":        "Mon, 02 Jan 2006 15:04:05 -0700",
		"named date/time/sql":  "date, time, datetime (also: rfc1123, rfc850, ansic, kitchen, stamp, http, us, eu, unix)",
		"reference time in go": "Mon Jan 2 15:04:05 MST 2006 (1 2 3 4 5 6 7)",
	}
}

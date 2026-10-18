package dt

import (
	"strings"
	"testing"
	"time"

	"github.com/avestura/zagros/internal/tools"
	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "dt") }

func init() {
	// Wednesday 2024-01-31 12:00 UTC: all relative tests use this clock.
	nowFn = func() time.Time { return time.Date(2024, 1, 31, 12, 0, 0, 0, time.UTC) }
}

func callE(name string, kv ...any) (any, error) {
	m := map[string]any{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return tools.Call("dt", name, m)
}

func call(t *testing.T, name string, kv ...any) any {
	t.Helper()
	v, err := callE(name, kv...)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return v
}

func expect(t *testing.T, got any, want any, what string) {
	t.Helper()
	if got != want {
		t.Errorf("%s: got %v, want %v", what, got, want)
	}
}

func TestZonesLoad(t *testing.T) {
	for _, n := range zoneNames {
		if _, err := time.LoadLocation(n); err != nil {
			t.Errorf("zone %q does not load: %v", n, err)
		}
	}
	l := call(t, "tz_list", "filter", "tehran").([]string)
	if len(l) != 1 || l[0] != "Asia/Tehran" {
		t.Errorf("%v", l)
	}
	if len(call(t, "tz_list", "filter", "").([]string)) < 400 {
		t.Error("list is too short")
	}
}

func TestParseForms(t *testing.T) {
	cases := map[string]string{
		"2024-01-31T15:04:05Z":                  "2024-01-31T15:04:05Z",
		"2024-01-31T15:04:05.250+03:30":         "2024-01-31T15:04:05.25+03:30",
		"2024-01-31 15:04":                      "2024-01-31T15:04:00Z",
		"2024-01-31":                            "2024-01-31T00:00:00Z",
		"31/01/2024":                            "2024-01-31T00:00:00Z",
		"03/04/2024":                            "2024-03-04T00:00:00Z",
		"31.01.2024 08:30":                      "2024-01-31T08:30:00Z",
		"Jan 2, 2006":                           "2006-01-02T00:00:00Z",
		"January 2nd, 2006 at 3:04 PM":          "2006-01-02T15:04:00Z",
		"2 Jan 2006":                            "2006-01-02T00:00:00Z",
		"31-Jan-2024":                           "2024-01-31T00:00:00Z",
		"Wed, 31 Jan 2024 15:04:05 +0330":       "2024-01-31T15:04:05+03:30",
		"Mon Jan  2 15:04:05 MST 2006":          "2006-01-02T15:04:05-07:00",
		"20240131":                              "2024-01-31T00:00:00Z",
		"20240131150405":                        "2024-01-31T15:04:05Z",
		"2024":                                  "2024-01-01T00:00:00Z",
		"2024-W05-3":                            "2024-01-31T00:00:00Z",
		"2024-031":                              "2024-01-31T00:00:00Z",
		"5pm":                                   "2024-01-31T17:00:00Z",
		"14:30":                                 "2024-01-31T14:30:00Z",
		"1700000000":                            "2023-11-14T22:13:20Z",
		"1700000000123":                         "2023-11-14T22:13:20.123Z",
		"1700000000123456":                      "2023-11-14T22:13:20.123456Z",
		"1700000000123456789":                   "2023-11-14T22:13:20.123456789Z",
		"-1":                                    "1969-12-31T23:59:59Z",
		"-86400.5":                              "1969-12-30T23:59:59.5Z",
		"0x65b9e1a0":                            "2024-01-31T05:58:56Z",
		"now":                                   "2024-01-31T12:00:00Z",
		"today":                                 "2024-01-31T00:00:00Z",
		"tomorrow":                              "2024-02-01T00:00:00Z",
		"yesterday 6pm":                         "2024-01-30T18:00:00Z",
		"in 3 days":                             "2024-02-03T12:00:00Z",
		"3 weeks ago":                           "2024-01-10T12:00:00Z",
		"in 1 month":                            "2024-02-29T12:00:00Z",
		"an hour ago":                           "2024-01-31T11:00:00Z",
		"next friday":                           "2024-02-02T00:00:00Z",
		"last friday":                           "2024-01-26T00:00:00Z",
		"next wednesday":                        "2024-02-07T00:00:00Z",
		"last month":                            "2023-12-31T12:00:00Z",
		"next year":                             "2025-01-31T12:00:00Z",
		"monday 9am":                            "2024-02-05T09:00:00Z",
		"wednesday":                             "2024-01-31T00:00:00Z",
		"end of month":                          "2024-01-31T23:59:59.999999999Z",
		"start of next week":                    "2024-02-05T00:00:00Z",
		"2024-01-31 15:00 America/New_York":     "2024-01-31T15:00:00-05:00",
		"2024-01-31 15:00 EST":                  "2024-01-31T15:00:00-05:00",
		"2024-01-31 15:00 UTC+3:30":             "2024-01-31T15:00:00+03:30",
		"2024-01-31 15:00:00 +0000 UTC":         "2024-01-31T15:00:00Z",
		"2024-01-31 15:00:00.123456789 +0100 X": "",
	}
	for in, want := range cases {
		got, err := callE("parse", "text", in, "tz", "UTC")
		if want == "" {
			if err == nil {
				t.Errorf("parse(%q) should fail, got %v", in, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("parse(%q): %v", in, err)
			continue
		}
		if got != want {
			t.Errorf("parse(%q) = %v, want %s", in, got, want)
		}
	}
	expect(t, call(t, "parse", "text", "03/04/2024", "tz", "UTC", "day_first", true), "2024-04-03T00:00:00Z", "day_first")
	for _, bad := range []string{"", "garbage", "2024-02-30", "13/13/2024", "25:00", "Feb 30 2024", "next smurfday"} {
		if _, err := callE("parse", "text", bad, "tz", "UTC"); err == nil {
			t.Errorf("parse(%q) should fail", bad)
		}
	}
	expect(t, call(t, "is_valid", "text", "2024-02-29"), true, "is_valid leap")
	expect(t, call(t, "is_valid", "text", "2023-02-29"), false, "is_valid non-leap")
}

func TestConvert(t *testing.T) {
	m := call(t, "convert", "text", "1700000000", "tz", "UTC").(map[string]string)
	want := map[string]string{
		"unix": "1700000000", "unix_ms": "1700000000000", "unix_us": "1700000000000000", "unix_ns": "1700000000000000000",
		"iso": "2023-11-14T22:13:20Z", "rfc1123": "Tue, 14 Nov 2023 22:13:20 GMT", "weekday": "Tuesday",
		"iso_week": "2023-W46-2", "day_of_year": "318", "quarter": "4", "filetime": "133444736000000000",
		"dotnet_ticks": "638355968000000000", "webkit": "13344473600000000", "mac_absolute": "721692800.000000",
		"unix_hex": "0x6553f100", "julian_day": "2460263.425926", "excel_serial": "45244.925926", "relative": "2 months ago",
		"leap_year": "false", "offset": "+00:00",
	}
	for k, v := range want {
		if m[k] != v {
			t.Errorf("convert[%s] = %q, want %q", k, m[k], v)
		}
	}
	if m := call(t, "convert", "tz", "UTC").(map[string]string); m["iso"] != "2024-01-31T12:00:00Z" {
		t.Errorf("empty convert is now: %v", m["iso"])
	}
	if _, err := callE("convert", "text", "not a date"); err == nil {
		t.Error("convert should fail on garbage")
	}
	all := call(t, "now_all", "tz", "Asia/Tehran").(map[string]string)
	if all["offset"] != "+03:30" || all["tz"] != "Asia/Tehran" {
		t.Errorf("%v", all)
	}
}

func TestUnix(t *testing.T) {
	expect(t, call(t, "from_unix", "value", "1700000000", "tz", "UTC"), "2023-11-14T22:13:20Z", "from_unix")
	expect(t, call(t, "from_unix", "value", "1700000000000", "tz", "UTC"), "2023-11-14T22:13:20Z", "from_unix ms auto")
	expect(t, call(t, "from_unix", "value", "1700000000000", "unit", "s", "tz", "UTC") != nil, true, "forced unit")
	expect(t, call(t, "from_unix", "value", "-1", "tz", "UTC"), "1969-12-31T23:59:59Z", "negative")
	expect(t, call(t, "from_unix", "value", "0xFF", "tz", "UTC"), "1970-01-01T00:04:15Z", "hex")
	expect(t, call(t, "to_unix", "text", "1970-01-01", "tz", "UTC"), int64(0), "epoch")
	expect(t, call(t, "to_unix", "text", "1969-12-31 23:59:59", "tz", "UTC"), int64(-1), "before epoch")
	expect(t, call(t, "to_unix", "text", "2024-01-31", "unit", "ms", "tz", "UTC"), int64(1706659200000), "ms")
	expect(t, call(t, "unix_now"), int64(1706702400), "unix_now")
	expect(t, call(t, "unix_ms_now"), int64(1706702400000), "unix_ms_now")
	expect(t, call(t, "unix_hex", "text", "2024-01-31 09:49:04", "tz", "UTC"), "0x65ba1790", "unix_hex")
	expect(t, call(t, "from_unix_hex", "value", "65b9e1a0", "tz", "UTC"), "2024-01-31T05:58:56Z", "from_unix_hex")
	if _, err := callE("from_unix", "value", "abc"); err == nil {
		t.Error("bad unix")
	}
	if _, err := callE("from_unix", "value", "1", "unit", "years"); err == nil {
		t.Error("bad unit")
	}
}

func TestEpochSystems(t *testing.T) {
	expect(t, call(t, "filetime", "text", "1970-01-01", "tz", "UTC"), int64(116444736000000000), "filetime epoch")
	expect(t, call(t, "from_filetime", "value", int64(116444736000000000), "tz", "UTC"), "1970-01-01T00:00:00Z", "from_filetime")
	expect(t, call(t, "dotnet_ticks", "text", "0001-01-01", "tz", "UTC"), int64(0), "ticks zero")
	expect(t, call(t, "dotnet_ticks", "text", "1970-01-01", "tz", "UTC"), int64(621355968000000000), "ticks unix")
	expect(t, call(t, "from_dotnet_ticks", "value", int64(621355968000000000), "tz", "UTC"), "1970-01-01T00:00:00Z", "from ticks")
	expect(t, call(t, "webkit_time", "text", "1970-01-01", "tz", "UTC"), int64(11644473600000000), "webkit")
	expect(t, call(t, "from_webkit_time", "value", int64(11644473600000000), "tz", "UTC"), "1970-01-01T00:00:00Z", "from webkit")
	expect(t, call(t, "mac_absolute_time", "text", "2001-01-01", "tz", "UTC"), float64(0), "mac")
	expect(t, call(t, "from_mac_absolute_time", "value", float64(86400.5), "tz", "UTC"), "2001-01-02T00:00:00.5Z", "from mac")
	expect(t, call(t, "julian_day", "text", "2000-01-01 12:00", "tz", "UTC"), 2451545.0, "julian J2000")
	expect(t, call(t, "from_julian_day", "jd", 2451545.0, "tz", "UTC"), "2000-01-01T12:00:00Z", "from julian")
	expect(t, call(t, "julian_day", "text", "1970-01-01", "tz", "UTC"), 2440587.5, "julian unix epoch")
	// Excel 1900 leap-year bug
	expect(t, call(t, "excel_serial", "text", "1900-01-01", "tz", "UTC"), float64(1), "excel 1")
	expect(t, call(t, "excel_serial", "text", "1900-02-28", "tz", "UTC"), float64(59), "excel 59")
	expect(t, call(t, "excel_serial", "text", "1900-03-01", "tz", "UTC"), float64(61), "excel 61")
	expect(t, call(t, "excel_serial", "text", "2024-01-31", "tz", "UTC"), float64(45322), "excel 2024")
	expect(t, call(t, "from_excel_serial", "serial", float64(61), "tz", "UTC"), "1900-03-01T00:00:00Z", "from excel 61")
	expect(t, call(t, "from_excel_serial", "serial", float64(59), "tz", "UTC"), "1900-02-28T00:00:00Z", "from excel 59")
	expect(t, call(t, "from_excel_serial", "serial", float64(45322.5), "tz", "UTC"), "2024-01-31T12:00:00Z", "from excel 2024.5")
	expect(t, call(t, "from_excel_serial", "serial", float64(0), "date1904", true, "tz", "UTC"), "1904-01-01T00:00:00Z", "excel 1904")
	if _, err := callE("from_excel_serial", "serial", float64(60)); err == nil {
		t.Error("serial 60 must fail")
	}
	if _, err := callE("excel_serial", "text", "1899-12-31", "tz", "UTC"); err == nil {
		t.Error("pre-1900 excel must fail")
	}
}

func TestTimeZones(t *testing.T) {
	expect(t, call(t, "convert_tz", "text", "2024-01-31 15:00", "from_tz", "EST", "to_tz", "UTC+3:30"), "2024-01-31T23:30:00+03:30", "EST->UTC+3:30")
	expect(t, call(t, "convert_tz", "text", "2024-07-01 12:00", "from_tz", "Europe/Paris", "to_tz", "America/New_York"), "2024-07-01T06:00:00-04:00", "summer")
	expect(t, call(t, "convert_tz", "text", "2024-01-01 12:00", "from_tz", "Paris", "to_tz", "new york"), "2024-01-01T06:00:00-05:00", "city names")
	expect(t, call(t, "convert_tz", "text", "2024-01-31 15:00", "from_tz", "UTC", "to_tz", "asia/tehran", "format", "%Y-%m-%d %H:%M %Z"), "2024-01-31 18:30 +0330", "format+case-insensitive zone")
	if _, err := callE("convert_tz", "text", "now", "to_tz", "Mars/Olympus"); err == nil {
		t.Error("unknown zone")
	}
	i := call(t, "tz_info", "name", "Europe/Paris", "at", "2024-07-01").(map[string]string)
	if i["offset"] != "+02:00" || i["is_dst"] != "true" || i["abbreviation"] != "CEST" || !strings.HasPrefix(i["next_change"], "2024-10-27T02:00:00") {
		t.Errorf("%v", i)
	}
	i = call(t, "tz_info", "name", "Europe/Paris", "at", "2024-01-15").(map[string]string)
	if i["offset"] != "+01:00" || i["is_dst"] != "false" {
		t.Errorf("%v", i)
	}
	expect(t, call(t, "tz_offset", "name", "Asia/Kolkata"), "+05:30", "kolkata")
	expect(t, call(t, "tz_offset", "name", "Asia/Tehran", "at", "2024-07-01"), "+03:30", "tehran no DST since 2022")
	expect(t, call(t, "tz_offset", "name", "GMT-5"), "-05:00", "GMT-5")
	expect(t, call(t, "tz_offset", "name", "America/Sao_Paulo"), "-03:00", "sao paulo")
	w := call(t, "world_clock", "text", "2024-01-31 12:00", "zones", []string{"UTC", "Tokyo"}, "tz", "UTC").(map[string]string)
	if w["Tokyo"] != "2024-01-31 21:00 JST (Wed)" || w["UTC"] != "2024-01-31 12:00 UTC (Wed)" {
		t.Errorf("%v", w)
	}
	if len(call(t, "world_clock").(map[string]string)) < 5 {
		t.Error("default zones")
	}
	mp := call(t, "meeting_planner", "time", "15:00", "zones", []string{"Europe/Paris", "Asia/Tokyo", "America/New_York"}, "date", "2024-03-04").(map[string]string)
	if !strings.Contains(mp["Asia/Tokyo"], "23:00") || !strings.HasSuffix(mp["Asia/Tokyo"], "night") || !strings.HasSuffix(mp["America/New_York"], "working hours") {
		t.Errorf("%v", mp)
	}
}

func TestDateMath(t *testing.T) {
	expect(t, call(t, "add", "text", "2024-01-31", "spec", "1mo", "tz", "UTC"), "2024-02-29T00:00:00Z", "month clamp")
	expect(t, call(t, "add", "text", "2023-01-31", "spec", "1mo", "tz", "UTC"), "2023-02-28T00:00:00Z", "month clamp non-leap")
	expect(t, call(t, "add", "text", "2024-02-29", "spec", "1y", "tz", "UTC"), "2025-02-28T00:00:00Z", "leap day + year")
	expect(t, call(t, "add", "text", "2024-01-01 00:00", "spec", "1y2mo3w4d5h6m7s", "tz", "UTC"), "2025-03-26T05:06:07Z", "full spec")
	expect(t, call(t, "add", "text", "2024-01-31", "spec", "-1d", "tz", "UTC"), "2024-01-30T00:00:00Z", "negative")
	expect(t, call(t, "add", "text", "2024-01-31 12:00", "spec", "1d-2h", "tz", "UTC"), "2024-02-01T10:00:00Z", "mixed signs")
	expect(t, call(t, "add", "text", "now", "spec", "90 minutes", "tz", "UTC"), "2024-01-31T13:30:00Z", "default now")
	expect(t, call(t, "add", "text", "2024-01-31", "spec", "1.5h", "tz", "UTC"), "2024-01-31T01:30:00Z", "fractional hour")
	expect(t, call(t, "subtract", "text", "2024-03-31", "spec", "1mo", "tz", "UTC"), "2024-02-29T00:00:00Z", "subtract month clamp")
	expect(t, call(t, "add", "text", "1969-12-31 23:59:59", "spec", "2s", "tz", "UTC"), "1970-01-01T00:00:01Z", "across epoch")
	// DST: a calendar day keeps the wall clock, 24h does not
	expect(t, call(t, "add", "text", "2024-03-09 12:00", "spec", "1d", "tz", "America/New_York"), "2024-03-10T12:00:00-04:00", "dst +1d")
	expect(t, call(t, "add", "text", "2024-03-09 12:00", "spec", "24h", "tz", "America/New_York"), "2024-03-10T13:00:00-04:00", "dst +24h")
	for _, bad := range []string{"", "abc", "5 parsecs", "1d garbage"} {
		if _, err := callE("add", "text", "2024-01-01", "spec", bad); err == nil {
			t.Errorf("spec %q should fail", bad)
		}
	}
	d := call(t, "diff", "a", "2020-02-29", "b", "2024-03-01 12:00", "tz", "UTC").(map[string]string)
	if d["years"] != "4" || d["months"] != "0" || d["days"] != "1" || d["hours"] != "12" || d["total_days"] != "1462.5" || d["direction"] != "b is after a" {
		t.Errorf("%v", d)
	}
	d = call(t, "diff", "a", "2024-03-01", "b", "2024-01-31", "tz", "UTC").(map[string]string)
	if d["direction"] != "b is before a" || d["months"] != "1" || d["days"] != "1" || d["total_days"] != "-30" {
		t.Errorf("%v", d)
	}
	if call(t, "diff", "a", "2024-01-31 06:00", "tz", "UTC").(map[string]string)["human"] != "6 hours" {
		t.Error("diff default b is now")
	}
	ag := call(t, "age", "birthdate", "1990-05-17", "tz", "UTC").(map[string]string)
	if ag["years"] != "33" || ag["months"] != "8" || ag["days"] != "14" || ag["next_birthday"] != "2024-05-17" || ag["days_until_birthday"] != "107" {
		t.Errorf("%v", ag)
	}
	ag = call(t, "age", "birthdate", "2000-02-29", "at", "2023-03-01", "tz", "UTC").(map[string]string)
	if ag["years"] != "23" || ag["next_birthday"] != "2024-02-29" {
		t.Errorf("leap birthday %v", ag)
	}
	if _, err := callE("age", "birthdate", "2030-01-01", "tz", "UTC"); err == nil {
		t.Error("future birth date")
	}
	expect(t, call(t, "relative", "text", "2024-01-29", "tz", "UTC"), "2 days ago", "relative past")
	expect(t, call(t, "relative", "text", "2024-01-31 15:00", "tz", "UTC"), "in 3 hours", "relative future")
	expect(t, call(t, "relative", "text", "2024-01-31 12:00:03", "tz", "UTC"), "just now", "just now")
	expect(t, call(t, "relative", "text", "2022-01-31", "tz", "UTC"), "2 years ago", "years")
	expect(t, call(t, "countdown", "target", "2030-01-02", "at", "2024-01-01", "tz", "UTC"), "6 years, 1 day", "countdown")
	expect(t, call(t, "countdown", "target", "2024-01-31 12:00:05", "tz", "UTC"), "5 seconds", "countdown seconds")
	expect(t, call(t, "countdown", "target", "2024-01-30 12:00", "tz", "UTC"), "1 day ago", "countdown past")
}

func TestBoundariesAndCalendar(t *testing.T) {
	expect(t, call(t, "start_of", "text", "2024-05-17 13:45", "unit", "month", "tz", "UTC"), "2024-05-01T00:00:00Z", "start month")
	expect(t, call(t, "start_of", "text", "2024-05-17 13:45", "unit", "week", "tz", "UTC"), "2024-05-13T00:00:00Z", "start week monday")
	expect(t, call(t, "start_of", "text", "2024-05-17 13:45", "unit", "week", "week_start", "sunday", "tz", "UTC"), "2024-05-12T00:00:00Z", "start week sunday")
	expect(t, call(t, "start_of", "text", "2024-05-17", "unit", "quarter", "tz", "UTC"), "2024-04-01T00:00:00Z", "start quarter")
	expect(t, call(t, "start_of", "text", "2024-05-17 13:45:10", "unit", "hour", "tz", "UTC"), "2024-05-17T13:00:00Z", "start hour")
	expect(t, call(t, "end_of", "text", "2024-02-10", "unit", "month", "tz", "UTC"), "2024-02-29T23:59:59.999999999Z", "end month leap")
	expect(t, call(t, "end_of", "text", "2024-12-31 10:00", "unit", "year", "tz", "UTC"), "2024-12-31T23:59:59.999999999Z", "end year")
	expect(t, call(t, "end_of", "text", "2024-05-17", "unit", "week", "tz", "UTC"), "2024-05-19T23:59:59.999999999Z", "end week")
	if _, err := callE("start_of", "text", "2024-01-01", "unit", "decade"); err == nil {
		t.Error("bad unit")
	}
	expect(t, call(t, "day_of_week", "text", "2024-02-29", "tz", "UTC"), "Thursday", "dow")
	expect(t, call(t, "day_of_week", "text", "2024-02-25", "style", "iso", "tz", "UTC"), "7", "iso sunday")
	expect(t, call(t, "day_of_week", "text", "2024-02-25", "style", "number", "tz", "UTC"), "0", "number sunday")
	expect(t, call(t, "day_of_year", "text", "2024-12-31", "tz", "UTC"), int64(366), "doy leap")
	expect(t, call(t, "day_of_year", "text", "2023-12-31", "tz", "UTC"), int64(365), "doy")
	expect(t, call(t, "quarter", "text", "2024-08-15", "tz", "UTC"), int64(3), "quarter")
	expect(t, call(t, "iso_week_date", "text", "2024-12-30", "tz", "UTC"), "2025-W01-1", "iso week new year")
	expect(t, call(t, "iso_week_date", "text", "2021-01-01", "tz", "UTC"), "2020-W53-5", "iso week 53")
	expect(t, call(t, "week_of_year", "text", "2021-01-01", "tz", "UTC"), int64(53), "iso week")
	expect(t, call(t, "week_of_year", "text", "2024-01-01", "system", "us", "tz", "UTC"), int64(1), "us week 1")
	expect(t, call(t, "week_of_year", "text", "2024-01-07", "system", "us", "tz", "UTC"), int64(2), "us week 2")
	expect(t, call(t, "week_of_year", "text", "2024-01-01", "system", "sunday", "tz", "UTC"), int64(0), "%U")
	expect(t, call(t, "week_of_year", "text", "2024-01-01", "system", "monday", "tz", "UTC"), int64(1), "%W")
	expect(t, call(t, "is_leap_year", "year", "2024"), true, "2024")
	expect(t, call(t, "is_leap_year", "year", "1900"), false, "1900")
	expect(t, call(t, "is_leap_year", "year", "2000"), true, "2000")
	expect(t, call(t, "days_in_month", "year", int64(2024), "month", int64(2)), int64(29), "feb 2024")
	expect(t, call(t, "days_in_month", "year", int64(2100), "month", int64(2)), int64(28), "feb 2100")
	expect(t, call(t, "days_in_year", "year", "2024"), int64(366), "days in year")
	if _, err := callE("days_in_month", "year", int64(2024), "month", int64(13)); err == nil {
		t.Error("month 13")
	}
	expect(t, call(t, "days_between", "a", "2024-01-01", "b", "2024-12-31", "tz", "UTC"), int64(365), "days between")
	expect(t, call(t, "days_between", "a", "2024-01-01", "b", "2024-12-31", "inclusive", true, "tz", "UTC"), int64(366), "inclusive")
	expect(t, call(t, "days_between", "a", "2024-03-01", "b", "2024-02-01", "tz", "UTC"), int64(-29), "negative")
	expect(t, call(t, "days_until", "target", "2024-02-29", "tz", "UTC"), int64(29), "days until")
	expect(t, call(t, "is_weekend", "text", "2024-02-03", "tz", "UTC"), true, "saturday")
	expect(t, call(t, "is_weekend", "text", "2024-02-03", "weekend", "fri_sat", "tz", "UTC"), true, "fri_sat saturday")
	expect(t, call(t, "is_weekend", "text", "2024-02-04", "weekend", "fri_sat", "tz", "UTC"), false, "fri_sat sunday")
	expect(t, call(t, "business_days_between", "a", "2024-01-01", "b", "2024-01-08", "tz", "UTC"), int64(5), "bdays")
	expect(t, call(t, "business_days_between", "a", "2024-01-01", "b", "2024-01-08", "holidays", []string{"2024-01-01"}, "tz", "UTC"), int64(4), "bdays holiday")
	expect(t, call(t, "business_days_between", "a", "2024-01-01", "b", "2024-01-05", "inclusive", true, "tz", "UTC"), int64(5), "bdays inclusive")
	expect(t, call(t, "business_days_between", "a", "2024-01-08", "b", "2024-01-01", "tz", "UTC"), int64(-5), "bdays reversed")
	expect(t, call(t, "add_business_days", "text", "2024-01-31", "days", int64(5), "tz", "UTC"), "2024-02-07T00:00:00Z", "add bdays")
	expect(t, call(t, "add_business_days", "text", "2024-01-31", "days", int64(5), "holidays", []string{"2024-02-02"}, "tz", "UTC"), "2024-02-08T00:00:00Z", "add bdays holiday")
	expect(t, call(t, "add_business_days", "text", "2024-02-05", "days", int64(-1), "tz", "UTC"), "2024-02-02T00:00:00Z", "back over weekend")
	expect(t, call(t, "nth_weekday", "year", int64(2024), "month", int64(11), "weekday", "thursday", "n", int64(4)), "2024-11-28", "thanksgiving")
	expect(t, call(t, "nth_weekday", "year", int64(2024), "month", int64(5), "weekday", "monday", "n", int64(-1)), "2024-05-27", "memorial day")
	if _, err := callE("nth_weekday", "year", int64(2024), "month", int64(2), "weekday", "friday", "n", int64(5)); err == nil {
		t.Error("no 5th friday in Feb 2024? (there is: 2,9,16,23 only four)")
	}
	for y, want := range map[string]string{"2024": "2024-03-31", "2025": "2025-04-20", "2000": "2000-04-23", "1961": "1961-04-02", "2038": "2038-04-25"} {
		expect(t, call(t, "easter", "year", y), want, "easter "+y)
	}
	for y, want := range map[string]string{"2024": "2024-05-05", "2025": "2025-04-20", "2023": "2023-04-16"} {
		expect(t, call(t, "easter", "year", y, "calendar", "orthodox"), want, "orthodox "+y)
	}
	cal := call(t, "calendar", "year", int64(2024), "month", int64(2)).(string)
	lines := strings.Split(cal, "\n")
	if len(lines) != 7 || strings.TrimSpace(lines[0]) != "February 2024" || lines[1] != "Mo Tu We Th Fr Sa Su" || lines[2] != "          1  2  3  4" || lines[6] != "26 27 28 29" {
		t.Errorf("calendar:\n%s", cal)
	}
	cal = call(t, "calendar", "year", int64(2024), "month", int64(2), "week_start", "sunday").(string)
	if !strings.Contains(cal, "Su Mo Tu We Th Fr Sa") {
		t.Error(cal)
	}
	rows := call(t, "month_calendar_rows", "year", int64(2024), "month", int64(2)).([]string)
	if len(rows) != 5 || rows[0] != ",,,1,2,3,4" || rows[4] != "26,27,28,29,,," {
		t.Errorf("%q", rows)
	}
}

func TestFormat(t *testing.T) {
	const in = "2024-01-05 03:04:05.123456"
	cases := map[string]string{
		"%Y-%m-%d %H:%M:%S": "2024-01-05 03:04:05", "%A, %B %-d, %Y": "Friday, January 5, 2024", "%j %U %W %V %G": "005 00 01 01 2024",
		"%I:%M %p": "03:04 AM", "%y%m%d": "240105", "%f %L": "123456 123", "%a %b %e": "Fri Jan  5", "%u %w": "5 5",
		"%F %T": "2024-01-05 03:04:05", "%D": "01/05/24", "%s": "1704423845", "%z %:z %Z": "+0000 +00:00 UTC", "%% %q": "% 1",
		"2006-01-02 15:04": "2024-01-05 03:04", "Mon Jan _2": "Fri Jan  5", "rfc2822": "Fri, 05 Jan 2024 03:04:05 +0000",
		"rfc3339": "2024-01-05T03:04:05Z", "date": "2024-01-05", "kitchen": "3:04AM", "unix": "1704423845", "http": "Fri, 05 Jan 2024 03:04:05 GMT",
	}
	for layout, want := range cases {
		if got, err := callE("format", "text", in, "layout", layout, "tz", "UTC"); err != nil || got != want {
			t.Errorf("format(%q) = %v, %v; want %q", layout, got, err, want)
		}
	}
	if _, err := callE("format", "text", in, "layout", "%Ö", "tz", "UTC"); err == nil {
		t.Error("unsupported token")
	}
	if len(call(t, "layouts").(map[string]string)) < 40 {
		t.Error("cheatsheet")
	}
	expect(t, call(t, "parse_with", "text", "31/01/2024 15:04", "layout", "%d/%m/%Y %H:%M", "tz", "UTC"), "2024-01-31T15:04:00Z", "strptime")
	expect(t, call(t, "parse_with", "text", "Jan 2 2006 3:04PM", "layout", "Jan 2 2006 3:04PM", "tz", "UTC"), "2006-01-02T15:04:00Z", "go layout")
	expect(t, call(t, "parse_with", "text", "1700000000", "layout", "%s", "tz", "UTC"), "2023-11-14T22:13:20Z", "%s")
	expect(t, call(t, "parse_with", "text", "5/1/24", "layout", "%-d/%-m/%y", "tz", "UTC"), "2024-01-05T00:00:00Z", "unpadded")
	if _, err := callE("parse_with", "text", "2024-01-31", "layout", "%d/%m/%Y"); err == nil {
		t.Error("mismatch must fail")
	}
}

func TestDurations(t *testing.T) {
	for in, want := range map[string]float64{
		"1h30m": 5400, "90 minutes": 5400, "1.5h": 5400, "PT1H30M": 5400, "P1DT2H": 93600, "1:30:00": 5400, "1:30": 90,
		"90": 90, "-5m": -300, "2 weeks": 1209600, "1d 2h and 3 minutes": 93780, "250ms": 0.25, "1h-30m": 1800, "-PT1M": -60, "1:02:05.5": 3725.5,
	} {
		got, err := callE("duration_parse", "text", in)
		if err != nil || got != want {
			t.Errorf("duration_parse(%q) = %v, %v; want %v", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "abc", "P", "PT", "1h xyz", "5 parsecs"} {
		if _, err := callE("duration_parse", "text", bad); err == nil {
			t.Errorf("duration_parse(%q) should fail", bad)
		}
	}
	for style, want := range map[string]string{"compact": "1h30m", "long": "1 hour, 30 minutes", "iso8601": "PT1H30M", "clock": "1:30:00"} {
		expect(t, call(t, "duration_format", "seconds", float64(5400), "style", style), want, "format "+style)
	}
	expect(t, call(t, "duration_format", "seconds", float64(93784.5), "style", "compact"), "1d2h3m4.5s", "compact days")
	expect(t, call(t, "duration_format", "seconds", float64(0), "style", "long"), "0 seconds", "zero long")
	expect(t, call(t, "duration_format", "seconds", float64(-90), "style", "clock"), "-0:01:30", "negative clock")
	expect(t, call(t, "duration_format", "seconds", float64(90000), "style", "iso8601"), "P1DT1H", "iso days")
	expect(t, call(t, "duration_format", "seconds", float64(0), "style", "iso8601"), "PT0S", "iso zero")
	expect(t, call(t, "duration_format", "seconds", float64(90000), "style", "clock"), "25:00:00", "clock over a day")
	expect(t, call(t, "iso8601_duration_to_seconds", "text", "P1Y"), float64(365*86400), "iso year")
	expect(t, call(t, "iso8601_duration_to_seconds", "text", "P2W"), float64(14*86400), "iso weeks")
	expect(t, call(t, "iso8601_duration_to_seconds", "text", "PT0.5S"), 0.5, "iso fraction")
	if _, err := callE("iso8601_duration_to_seconds", "text", "1h"); err == nil {
		t.Error("not iso")
	}
	expect(t, call(t, "seconds_to_iso8601_duration", "seconds", float64(5400)), "PT1H30M", "to iso")
	expect(t, call(t, "sec_to_hms", "seconds", float64(3725)), "1:02:05", "hms")
	expect(t, call(t, "hms_to_sec", "text", "1:02:05"), float64(3725), "from hms")
	expect(t, call(t, "time_math", "expr", "1:30 + 2:45 - 0:20"), "3:55", "time math")
	expect(t, call(t, "time_math", "expr", "1h30m + 45m"), "2:15", "time math units")
	expect(t, call(t, "time_math", "expr", "0:30 - 1:00"), "-0:30", "time math negative")
	expect(t, call(t, "time_math", "expr", "10:00:30 - 0:00:45"), "9:59:45", "time math seconds")
	expect(t, call(t, "time_math", "expr", "-1:00 + 2:00"), "1:00", "leading minus")
	expect(t, call(t, "time_math", "expr", "23:00 + 5:30"), "28:30", "over 24h")
	if _, err := callE("time_math", "expr", "1:30 + banana"); err == nil {
		t.Error("bad term")
	}
}

func TestCron(t *testing.T) {
	got := call(t, "cron_next", "expr", "0 9 * * mon-fri", "count", int64(4), "from", "2024-01-31 12:00", "tz", "UTC").([]string)
	want := []string{"2024-02-01T09:00:00Z", "2024-02-02T09:00:00Z", "2024-02-05T09:00:00Z", "2024-02-06T09:00:00Z"}
	if strings.Join(got, ",") != strings.Join(want, ",") {
		t.Errorf("%v", got)
	}
	got = call(t, "cron_next", "expr", "*/20 * * * *", "count", int64(3), "from", "2024-01-31 12:00", "tz", "UTC").([]string)
	if got[0] != "2024-01-31T12:20:00Z" || got[2] != "2024-01-31T13:00:00Z" {
		t.Errorf("strictly after from: %v", got)
	}
	got = call(t, "cron_next", "expr", "0 0 29 2 *", "count", int64(2), "from", "2024-03-01", "tz", "UTC").([]string)
	if got[0] != "2028-02-29T00:00:00Z" || got[1] != "2032-02-29T00:00:00Z" {
		t.Errorf("leap day cron: %v", got)
	}
	// dom + dow restricted => OR
	got = call(t, "cron_next", "expr", "0 0 13 * fri", "count", int64(3), "from", "2024-09-01", "tz", "UTC").([]string)
	if got[0] != "2024-09-06T00:00:00Z" || got[1] != "2024-09-13T00:00:00Z" || got[2] != "2024-09-20T00:00:00Z" {
		t.Errorf("dom/dow OR: %v", got)
	}
	got = call(t, "cron_next", "expr", "@hourly", "count", int64(2), "from", "2024-01-31 12:30", "tz", "UTC").([]string)
	if got[0] != "2024-01-31T13:00:00Z" {
		t.Errorf("%v", got)
	}
	got = call(t, "cron_next", "expr", "30 2 * * *", "count", int64(2), "from", "2024-03-09 12:00", "tz", "America/New_York").([]string)
	if got[0] != "2024-03-10T03:30:00-04:00" && got[0] != "2024-03-11T02:30:00-04:00" {
		// 02:30 does not exist on 2024-03-10: that day is skipped
		t.Errorf("dst gap: %v", got)
	}
	if got[0] == "2024-03-10T03:30:00-04:00" {
		t.Errorf("nonexistent local time must be skipped, got %v", got)
	}
	got = call(t, "cron_next", "expr", "0 0 1 jan,jul sun-sat/2", "count", int64(1), "from", "2024-01-31", "tz", "UTC").([]string)
	if len(got) != 1 {
		t.Errorf("%v", got)
	}
	for _, bad := range []string{"", "* * * *", "60 * * * *", "* 24 * * *", "* * 0 * *", "* * * 13 *", "*/0 * * * *", "a b c d e", "5-1 * * * *"} {
		if _, err := callE("cron_next", "expr", bad); err == nil {
			t.Errorf("cron %q should fail", bad)
		}
	}
	if _, err := callE("cron_next", "expr", "* * * * *", "count", int64(1001)); err == nil {
		t.Error("count cap")
	}
	for expr, want := range map[string]string{
		"0 9 * * mon-fri":    "At 09:00, on Monday through Friday",
		"* * * * *":          "Every minute",
		"*/15 * * * *":       "Every 15 minutes",
		"30 2 1 * *":         "At 02:30, on day-of-month 1",
		"0 0 1 1 *":          "At 00:00, on day-of-month 1, in January",
		"@daily":             "At 00:00",
		"0 9,17 * * 1,3,5":   "At 09:00 and 17:00, on Monday, Wednesday and Friday",
		"*/15 9-17 * * 1-5":  "Every 15 minutes, between 09:00 and 17:59, on Monday through Friday",
		"5 * * * *":          "At minute 5 past every hour",
		"0 0 * * 7":          "At 00:00, on Sunday",
		"0 12 * jun-aug sun": "At 12:00, on Sunday, in June through August",
	} {
		if got, err := callE("cron_explain", "expr", expr); err != nil || got != want {
			t.Errorf("cron_explain(%q) = %v, %v; want %q", expr, got, err, want)
		}
	}
}

func TestNow(t *testing.T) {
	expect(t, call(t, "now", "tz", "UTC"), "2024-01-31T12:00:00Z", "now")
	expect(t, call(t, "now", "tz", "Asia/Tokyo", "format", "%Y-%m-%d %H:%M"), "2024-01-31 21:00", "now tokyo")
	expect(t, call(t, "now", "tz", "tehran", "format", "date"), "2024-01-31", "city + named layout")
	if _, err := callE("now", "tz", "Nowhere/Land"); err == nil {
		t.Error("bad zone")
	}
}

"""Date and time: Parse, convert and calculate dates, times, time zones, durations and cron schedules.

Generated from the tools registry; do not edit by hand.
Type description, not runnable Starlark: do not `import` anything. The name `dt` is predeclared in main.star; call `dt.<name>(...)` directly.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
Keywords: date, time, datetime, timestamp, unix time, epoch, time zone, timezone, duration, add days, date difference, age, weekday, calendar, cron, iso 8601, format date, business days
"""

from typing import Any, Dict, List

def convert(text: str = "", tz: str = "", at: str = "", day_first: bool = False) -> Dict[str, str]:
    """The all-in-one converter: reads any date/time/timestamp (empty text means now) and returns the instant as Unix s/ms/us/ns, ISO UTC, RFC 1123/2822, local, relative time, weekday, ISO week, day of year, quarter, Julian day, Excel serial, Windows FILETIME, .NET ticks, WebKit/Chrome time, Mac absolute time and Unix hex.

        text: Date/time text or a numeric timestamp; empty means now.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.convert("1700000000")
    Task: convert a date between formats
    Keywords: reformat date, iso to rfc, timestamp converter, epoch converter, unix time, filetime, julian, excel date, webkit, ticks, all formats, unix to date, date to unix
    Needs: none (pure function).
    """

def now_all(tz: str = "") -> Dict[str, str]:
    """Returns the current moment in many representations at once: ISO 8601, RFC 2822, Unix seconds/ms, UTC, local, ISO week, day of year, quarter, Julian day and more.

        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.

    Example: dt.now_all("UTC")
    Keywords: current time, all formats, timestamp, iso week, epoch, clock
    Needs: none (pure function).
    """

def parse(text: str, tz: str = "", at: str = "", day_first: bool = False) -> str:
    """Parses almost any date text and returns RFC 3339: ISO 8601, RFC 2822, '2024-01-31 15:04', '31/01/2024', 'Jan 2, 2006', Unix seconds/ms/us/ns (guessed by magnitude), today/tomorrow/yesterday, 'in 3 days', '3 weeks ago', 'next friday', 'monday 9am'. Ambiguous numeric dates are month/day unless day_first is set. Errors when nothing fits.

        text: Date/time text.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.parse("next friday 5pm", "Europe/Paris")
    Task: parse a date string
    Keywords: read date, string to datetime, interpret date text, date parser, natural language date, iso 8601, rfc 3339, relative date, str to date, strtotime, detect format
    Needs: none (pure function).
    """

def convert_tz(text: str = "", from_tz: str = "", to_tz: str = "UTC", format: str = "rfc3339", at: str = "") -> str:
    """Converts a date/time from one time zone to another. The text is read in from_tz unless it carries its own zone; zones may be IANA names, cities, abbreviations (EST, PST) or offsets (UTC+3:30). Format is a Go layout, strftime string or name (default RFC 3339).

        text: Date/time text; empty means now.
        from_tz: Zone the text is in (empty = local).
        to_tz: Target zone.
        format: Output layout.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.

    Example: dt.convert_tz("2024-01-31 15:00", "America/New_York", "Asia/Tehran")
    Task: convert a time between time zones
    Keywords: timezone conversion, what time is it in, utc to local, timezone converter, time zone, est to utc, pst, offset, world time, meeting time
    Needs: none (pure function).
    """

def diff(a: str, b: str = "", tz: str = "", at: str = "", day_first: bool = False) -> Dict[str, str]:
    """Compares two dates: total seconds, minutes, hours, days and weeks plus a calendar breakdown (years, months, days, hours, minutes, seconds) and human text. Direction says whether b is after or before a. Empty b means now.

        a: Start date/time.
        b: End date/time; empty means now.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.diff("2020-02-29", "2024-03-01 12:00", "UTC")
    Task: time between two dates
    Keywords: date difference, elapsed time, how long between, difference between dates, time between, how long, duration between, elapsed, days between, calendar difference
    Needs: none (pure function).
    """

def add(text: str, spec: str, tz: str = "", at: str = "", day_first: bool = False) -> str:
    """Adds a duration to a date (empty text = now). The spec mixes units and signs: '1y2mo3w4d5h6m7s', '-90m', '1d-2h'. Month/year steps clamp to the month end (Jan 31 + 1mo = Feb 28/29); day steps keep the wall-clock time across DST.

        text: Date/time text; use now for the current time.
        spec: Duration spec such as 1y2mo3w4d5h6m7s, '-3 days', '2 hours 30 minutes'. Units: y mo w d h m s ms us ns. Months and years are calendar-aware and clamp to month end.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.add("2024-01-31", "1mo", "UTC")
    Task: add time to a date
    Keywords: date plus days, shift date, weeks from now, in 3 days, date math, plus days, in 3 weeks, date arithmetic, add days
    Needs: none (pure function).
    """

def relative(text: str, tz: str = "", at: str = "", day_first: bool = False) -> str:
    """Describes a date relative to a reference as '3 hours ago', 'in 2 days' or 'just now' (within 10 seconds). Units step from seconds up to years; reference is at (default now).

        text: Date/time text.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.relative("2024-01-29", at="2024-01-31")
    Task: describe a time relative to now
    Keywords: time ago, in 5 minutes, humanize date, humanize, fuzzy time, in 2 days, natural
    Needs: none (pure function).
    """

def duration_parse(text: str) -> float:
    """Parses a duration to seconds: '1h30m', '90 minutes', '1.5h', 'PT1H30M', 'P1DT2H', '1:30:00' (h:m:s) or '1:30' (m:s), or a bare number of seconds. A year counts 365 days and a month 30 days. Negative values are allowed.

        text: Duration text.

    Example: dt.duration_parse("1h30m")
    Task: parse a duration
    Keywords: 1h30m to seconds, time span, human duration, to seconds, hours to seconds, timespan, iso 8601 duration, convert duration
    Needs: none (pure function).
    """

def now(tz: str = "", format: str = "rfc3339") -> str:
    """Returns the current time in a zone, formatted with a Go layout, strftime tokens (%Y-%m-%d %H:%M) or a layout name (rfc3339, rfc2822, date, time, unix). Default format is RFC 3339.

        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        format: Go layout, strftime string or layout name.

    Example: dt.now("Asia/Tokyo", "%Y-%m-%d %H:%M")
    Task: get the current time
    Keywords: today, current date, right now, clock, current time, timestamp, what time is it, date now
    Needs: none (pure function).
    """

def add_business_days(text: str, days: int, holidays: List[str] = [], weekend: str = "sat_sun", tz: str = "", day_first: bool = False) -> str:
    """Moves a date forward (or backward when days is negative) by a number of working days, skipping weekends and holidays; the start day is not counted and the time of day is kept. |days| is limited to 100000.

        text: Start date; use today for today.
        days: Working days to add (negative goes back).
        holidays: Dates to treat as non-working.
        weekend: Weekend days: sat_sun (default), fri_sat, sun, none, or a list like fri,sat.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.add_business_days("2024-01-31", 5, ["2024-02-02"], tz="UTC")
    Task: add working days to a date
    Keywords: skip weekends, workdays from now, workday, due date, deadline, working days from now, delivery date
    Needs: none (pure function).
    """

def age(birthdate: str, at: str = "", tz: str = "", day_first: bool = False) -> Dict[str, str]:
    """Calculates age from a birth date: years, months and days, total days lived, the next birthday and days until it. A Feb 29 birthday is celebrated on Feb 28 in non-leap years. Errors if the birth date is after the reference date.

        birthdate: Birth date.
        at: Date to measure at (empty = now).
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.age("1990-05-17", "2024-01-31", "UTC")
    Task: work out an age from a birth date
    Keywords: how old, years since, birthday, years old, age in days, next birthday, date of birth
    Needs: none (pure function).
    """

def business_days_between(a: str, b: str, inclusive: bool = False, holidays: List[str] = [], weekend: str = "sat_sun", tz: str = "", day_first: bool = False) -> int:
    """Counts working days from a up to b: weekends and listed holidays are skipped. b itself is excluded unless inclusive is true; the result is negative when b is before a.

        a: Start date.
        b: End date.
        inclusive: Also count b when it is a working day.
        holidays: Dates to treat as non-working.
        weekend: Weekend days: sat_sun (default), fri_sat, sun, none, or a list like fri,sat.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.business_days_between("2024-01-01", "2024-01-31", ["2024-01-01"])
    Keywords: working days between, weekdays count, working days, workdays, networkdays, weekdays between, excluding weekends, banking days
    Needs: none (pure function).
    """

def calendar(year: int, month: int, week_start: str = "monday") -> str:
    """Returns a month as a text calendar (like the Unix cal command) with weeks starting on Monday or Sunday.

        year: Year.
        month: Month 1..12.
        week_start: monday or sunday.

    Example: dt.calendar(2024, 2, "sunday")
    Keywords: cal, month view, calendar grid, print calendar
    Needs: none (pure function).
    """

def countdown(target: str, tz: str = "", at: str = "", day_first: bool = False) -> str:
    """Returns the time remaining until a target as '12 days, 3 hours, 5 minutes, 7 seconds'; for a target in the past it says how long ago it was. Reference is at (default now).

        target: Target date/time.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.countdown("2030-01-01", at="2024-01-01")
    Keywords: time left, time until, remaining, deadline, until new year
    Needs: none (pure function).
    """

def cron_explain(expr: str) -> str:
    """Explains a 5-field cron expression in words, e.g. 'At 09:00, on Monday through Friday'. Errors on invalid fields.

        expr: Cron expression or @macro.

    Example: dt.cron_explain("*/15 9-17 * * mon-fri")
    Task: explain a cron schedule
    Keywords: cron expression in words, what does this cron do, crontab, cron to english, cron description, schedule meaning, decode cron
    Needs: none (pure function).
    """

def cron_next(expr: str, count: int = 5, from: str = "", tz: str = "") -> List[str]:
    """Lists the next count (1..1000) run times of a 5-field cron expression (minute hour day-of-month month day-of-week) after from, as RFC 3339 in tz. Supports *, lists, ranges, steps, month/day names, @daily-style macros; day-of-month and day-of-week combine with OR when both are restricted.

        expr: Cron expression such as '*/15 9-17 * * mon-fri'.
        count: How many runs.
        from: Start instant (empty = now); runs are strictly after it.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.

    Example: dt.cron_next("0 9 * * mon-fri", 3, "2024-01-31", "UTC")
    Task: list the next cron run times
    Keywords: when will cron run, upcoming schedule, crontab, schedule, next run, cron preview, cron tester, job scheduler
    Needs: none (pure function).
    """

def day_of_week(text: str = "", style: str = "name", tz: str = "", at: str = "", day_first: bool = False) -> str:
    """Returns the weekday of a date. style: name (Wednesday), short (Wed), iso (1=Monday..7=Sunday) or number (0=Sunday..6).

        text: Date text; empty means today.
        style: name, short, iso or number.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.day_of_week("2024-02-29")
    Task: find the weekday of a date
    Keywords: what day is it, monday tuesday, weekday, which day of the week, monday
    Needs: none (pure function).
    """

def day_of_year(text: str = "", tz: str = "", at: str = "", day_first: bool = False) -> int:
    """Returns the day of the year, 1..366 (ordinal date).

        text: Date text; empty means today.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.day_of_year("2024-12-31")
    Keywords: julian day number, ordinal date, yday, nth day
    Needs: none (pure function).
    """

def days_between(a: str, b: str = "", inclusive: bool = False, tz: str = "", at: str = "", day_first: bool = False) -> int:
    """Returns the number of calendar days from a to b (negative if b is earlier), ignoring the time of day. inclusive counts both end dates. Empty b means today.

        a: Start date.
        b: End date; empty means today.
        inclusive: Count both the first and the last day.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.days_between("2024-01-01", "2024-12-31")
    Task: count days between two dates
    Keywords: number of days, days elapsed, days until, days since, how many days
    Needs: none (pure function).
    """

def days_in_month(year: int, month: int) -> int:
    """Returns the number of days (28..31) in a month of a year, honouring leap years. Month is 1..12.

        year: Year.
        month: Month 1..12.

    Example: dt.days_in_month(2024, 2)
    Keywords: month length, how many days in february, 28 29 30 31
    Needs: none (pure function).
    """

def days_in_year(year: str = "") -> int:
    """Returns 365 or 366 for the year.

        year: Year such as 2024; empty means this year.

    Example: dt.days_in_year("2024")
    Keywords: year length, 365, 366
    Needs: none (pure function).
    """

def days_until(target: str, tz: str = "", at: str = "", day_first: bool = False) -> int:
    """Returns the number of calendar days from today (or at) until the target date; negative when it has passed.

        target: Target date.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.days_until("2030-01-01", at="2024-01-01")
    Task: count days until a date
    Keywords: countdown days, how many days left, days left, days to go, how many days until
    Needs: none (pure function).
    """

def dotnet_ticks(text: str = "", tz: str = "", at: str = "", day_first: bool = False) -> int:
    """Returns .NET DateTime ticks: 100-nanosecond intervals since 0001-01-01 UTC.

        text: Date/time text; empty means now.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.dotnet_ticks("2024-01-31", "UTC")
    Keywords: csharp, datetime.ticks, 100 nanoseconds, vb.net
    Needs: none (pure function).
    """

def duration_format(seconds: float, style: str = "compact") -> str:
    """Formats a number of seconds (decimals allowed) as compact (1h30m), long (1 hour, 30 minutes), iso8601 (PT1H30M) or clock (1:30:00). Days are the largest unit except in clock style.

        seconds: Duration in seconds.
        style: compact, long, iso8601 or clock.

    Example: dt.duration_format(5400, "long")
    Keywords: seconds to human time, pretty duration, seconds to hours, humanize duration, timespan, hh:mm:ss
    Needs: none (pure function).
    """

def easter(year: str = "", calendar: str = "western") -> str:
    """Returns the date of Easter Sunday for a year (1583..9999 Gregorian). calendar: western (default) or orthodox (Julian computus, shown in the Gregorian calendar).

        year: Year such as 2025; empty means this year.
        calendar: western or orthodox.

    Example: dt.easter("2025")
    Keywords: easter sunday, holiday, computus, movable feast, good friday, pascha
    Needs: none (pure function).
    """

def end_of(text: str = "", unit: str = "day", week_start: str = "monday", tz: str = "", at: str = "", day_first: bool = False) -> str:
    """Returns the last nanosecond of the minute, hour, day, week, month, quarter or year containing the date, e.g. 23:59:59.999999999 for a day.

        text: Date/time text; empty means now.
        unit: minute, hour, day, week, month, quarter or year.
        week_start: First day of the week (monday or sunday...).
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.end_of("2024-02-10", "month", tz="UTC")
    Keywords: end of day, last day of month, end of week, ceil date
    Needs: none (pure function).
    """

def excel_serial(text: str = "", tz: str = "", at: str = "", day_first: bool = False) -> float:
    """Returns the Excel/Lotus serial number of a date (1900 system, 1 = 1900-01-01) using the wall-clock time in tz. Excel wrongly treats 1900 as a leap year, so serials from 60 on are one higher than elapsed days; dates before 1900 are an error.

        text: Date/time text; empty means now.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.excel_serial("2024-01-31", "UTC")
    Keywords: excel date, spreadsheet date, xlsx, serial number, 1900 leap bug, ole automation
    Needs: none (pure function).
    """

def filetime(text: str = "", tz: str = "", at: str = "", day_first: bool = False) -> int:
    """Returns the Windows FILETIME of an instant: 100-nanosecond intervals since 1601-01-01 UTC.

        text: Date/time text; empty means now.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.filetime("2024-01-31", "UTC")
    Keywords: windows file time, ntfs timestamp, win32, 100 nanoseconds, 1601
    Needs: none (pure function).
    """

def format(text: str = "", layout: str = "rfc3339", tz: str = "", at: str = "", day_first: bool = False) -> str:
    """Formats a date/time with a Go layout (2006-01-02 15:04:05), strftime tokens (%Y-%m-%d %H:%M:%S %A %B %j %U %Z %z, %-d for no padding) or a name (rfc3339, rfc2822, date, datetime, kitchen, unix...). Empty text means now.

        text: Date/time text; empty means now.
        layout: Go layout, strftime string or layout name.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.format("2024-01-31", "%A, %B %-d, %Y")
    Task: format a date
    Keywords: date to string, strftime, pretty print date, date format, go layout, pattern, moment format
    Needs: none (pure function).
    """

def from_dotnet_ticks(value: int, tz: str = "") -> str:
    """Converts .NET DateTime ticks (100 ns since 0001-01-01) to RFC 3339.

        value: The numeric timestamp.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.

    Example: dt.from_dotnet_ticks(638422464000000000, "UTC")
    Keywords: csharp, datetime from ticks
    Needs: none (pure function).
    """

def from_excel_serial(serial: float, date1904: bool = False, tz: str = "") -> str:
    """Converts an Excel serial date to RFC 3339 (wall clock in tz). Serial 60 (the non-existent 1900-02-29) is an error; date1904 selects the old Mac system.

        serial: Excel serial number, e.g. 45322.5.
        date1904: Use the 1904 date system.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.

    Example: dt.from_excel_serial(45322.5, tz="UTC")
    Keywords: excel date, spreadsheet, xlsx, serial to date, ole automation
    Needs: none (pure function).
    """

def from_filetime(value: int, tz: str = "") -> str:
    """Converts a Windows FILETIME (100 ns intervals since 1601-01-01 UTC) to RFC 3339.

        value: The numeric timestamp.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.

    Example: dt.from_filetime(133511904000000000, "UTC")
    Keywords: win32 timestamp, ntfs, 1601
    Needs: none (pure function).
    """

def from_julian_day(jd: float, tz: str = "") -> str:
    """Converts a Julian day number (e.g. 2460341.5) to RFC 3339.

        jd: Julian day number.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.

    Example: dt.from_julian_day(2451545.0, "UTC")
    Keywords: jd to date, astronomy, julian date
    Needs: none (pure function).
    """

def from_mac_absolute_time(value: float, tz: str = "") -> str:
    """Converts Core Foundation absolute time (seconds since 2001-01-01 UTC, decimals allowed) to RFC 3339.

        value: Seconds since 2001-01-01.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.

    Example: dt.from_mac_absolute_time(728438400, "UTC")
    Keywords: cocoa, nsdate, ios, apple timestamp
    Needs: none (pure function).
    """

def from_unix(value: str, unit: str = "auto", tz: str = "") -> str:
    """Converts a Unix timestamp (integer or decimal, or 0x hex) to RFC 3339 in a zone. Unit auto guesses seconds, milliseconds, microseconds or nanoseconds by magnitude; set s, ms, us or ns to force it.

        value: Timestamp such as 1700000000 or 1700000000123.
        unit: auto, s, ms, us or ns.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.

    Example: dt.from_unix("1700000000", "s", "UTC")
    Task: convert a Unix timestamp to a date
    Keywords: epoch to date, timestamp to human time, seconds since 1970, timestamp to date, unix time converter, milliseconds, human readable
    Needs: none (pure function).
    """

def from_unix_hex(value: str, tz: str = "") -> str:
    """Converts hexadecimal Unix seconds (with or without 0x) to RFC 3339.

        value: Hex number such as 0x65b9e1a0.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.

    Example: dt.from_unix_hex("0x65b9e1a0", "UTC")
    Keywords: hex timestamp, epoch hex, base16
    Needs: none (pure function).
    """

def from_webkit_time(value: int, tz: str = "") -> str:
    """Converts a WebKit/Chrome timestamp (microseconds since 1601-01-01) to RFC 3339.

        value: The numeric timestamp.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.

    Example: dt.from_webkit_time(13348377600000000, "UTC")
    Keywords: chrome, chromium, cookies, history
    Needs: none (pure function).
    """

def hms_to_sec(text: str) -> float:
    """Converts H:MM:SS (or MM:SS) to seconds; accepts decimals in the seconds and a leading minus.

        text: Clock duration such as 1:02:05 or 02:05.

    Example: dt.hms_to_sec("1:02:05")
    Keywords: hh:mm:ss to seconds, timecode, clock to seconds
    Needs: none (pure function).
    """

def is_leap_year(year: str = "") -> bool:
    """Returns true when the year is a leap year in the Gregorian calendar (divisible by 4, except centuries not divisible by 400).

        year: Year such as 2024; empty means this year.

    Example: dt.is_leap_year("2100")
    Keywords: leap year check, feb 29, leap, 366 days, gregorian
    Needs: none (pure function).
    """

def is_valid(text: str, tz: str = "", day_first: bool = False) -> bool:
    """Returns true when parse() can read the text as a date or time.

        text: Date/time text.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.is_valid("2024-02-30")
    Keywords: validate date, check date, valid date, real date
    Needs: none (pure function).
    """

def is_weekend(text: str = "", tz: str = "", at: str = "", day_first: bool = False, weekend: str = "sat_sun") -> bool:
    """Returns true when the date falls on a weekend; weekend lists the days (default sat_sun, use fri_sat for Middle Eastern calendars).

        text: Date text; empty means today.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.
        weekend: sat_sun, fri_sat, sun, ... or none.

    Example: dt.is_weekend("2024-02-03")
    Keywords: saturday sunday, workday, day off
    Needs: none (pure function).
    """

def iso8601_duration_to_seconds(text: str) -> float:
    """Converts an ISO 8601 duration (PnYnMnWnDTnHnMnS, e.g. PT1H30M or P1DT12H) to seconds; years count 365 days and months 30 days. Errors on malformed input.

        text: ISO 8601 duration.

    Example: dt.iso8601_duration_to_seconds("PT1H30M")
    Keywords: pt1h, xml duration, youtube duration, period, rfc 3339 duration
    Needs: none (pure function).
    """

def iso_week_date(text: str = "", tz: str = "", at: str = "", day_first: bool = False) -> str:
    """Returns the ISO 8601 week date of a date, e.g. 2024-W05-3 (week-based year, week number, weekday 1=Monday). The week-based year can differ from the calendar year around New Year.

        text: Date text; empty means today.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.iso_week_date("2024-12-30")
    Keywords: iso 8601 week, week date, 2024-W01, week-based year
    Needs: none (pure function).
    """

def julian_day(text: str = "", tz: str = "", at: str = "", day_first: bool = False) -> float:
    """Returns the Julian day number (days since noon UTC on 4713 BC Jan 1, with fraction) of an instant; the Unix epoch is 2440587.5.

        text: Date/time text; empty means now.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.julian_day("2000-01-01 12:00", "UTC")
    Keywords: jd, astronomy, julian date, mjd
    Needs: none (pure function).
    """

def layouts() -> Dict[str, str]:
    """Returns a cheatsheet dict of Go layout elements, strftime tokens and layout names accepted by format() and parse_with().

    Example: dt.layouts()
    Keywords: strftime reference, go time layout, date format help, cheat sheet, tokens
    Needs: none (pure function).
    """

def mac_absolute_time(text: str = "", tz: str = "", at: str = "", day_first: bool = False) -> float:
    """Returns Apple Core Foundation absolute time: seconds (with fraction) since 2001-01-01 UTC.

        text: Date/time text; empty means now.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.mac_absolute_time("2024-01-31", "UTC")
    Keywords: cocoa, cfabsolutetime, ios, nsdate, apple
    Needs: none (pure function).
    """

def meeting_planner(zones: List[str], time: str = "09:00", date: str = "") -> Dict[str, str]:
    """Plans a meeting: takes a time of day (HH:MM) and date in the first zone and shows the local time in every zone with a note: working hours (09-17), early/late (07-09, 17-21) or night.

        zones: Zones; the first one is where the time is given.
        time: Time of day in the first zone, HH:MM.
        date: Date in the first zone; empty means today there.

    Example: dt.meeting_planner(["Europe/Paris", "America/New_York", "Asia/Tokyo"], "15:00", "2024-03-04")
    Task: find a meeting time across time zones
    Keywords: overlap working hours, schedule across zones, schedule, overlap, working hours, time zone meeting, find a time
    Needs: none (pure function).
    """

def month_calendar_rows(year: int, month: int, week_start: str = "monday") -> List[str]:
    """Returns the weeks of a month as list rows; each row has 7 comma-separated cells with the day number, or an empty cell outside the month (',,1,2,3,4,5').

        year: Year.
        month: Month 1..12.
        week_start: monday or sunday.

    Example: dt.month_calendar_rows(2024, 2)
    Keywords: calendar grid, weeks of month, month matrix, date picker
    Needs: none (pure function).
    """

def nth_weekday(year: int, month: int, weekday: str, n: int) -> str:
    """Returns the date of the n-th given weekday of a month (n = 1..5), or counted from the end with negative n (-1 = last). Errors when that date does not exist (e.g. a 5th Friday). Example: US Thanksgiving is the 4th Thursday of November.

        year: Year.
        month: Month 1..12.
        weekday: Weekday name such as thursday.
        n: Occurrence 1..5, or -1 for the last.

    Example: dt.nth_weekday(2024, 11, "thursday", 4)
    Keywords: thanksgiving, second tuesday, last friday of month, holiday date, patch tuesday
    Needs: none (pure function).
    """

def parse_with(text: str, layout: str, tz: str = "") -> str:
    """Parses text with an explicit layout (Go layout, strftime tokens like %d/%m/%Y %H:%M, or a name) and returns RFC 3339. Use it when parse() would guess wrong; errors when the text does not match.

        text: Date text.
        layout: Layout describing the text.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.

    Example: dt.parse_with("31/01/2024 15:04", "%d/%m/%Y %H:%M", "UTC")
    Keywords: strptime, custom date format, exact format, parse date with pattern
    Needs: none (pure function).
    """

def quarter(text: str = "", tz: str = "", at: str = "", day_first: bool = False) -> int:
    """Returns the calendar quarter 1..4 of a date.

        text: Date text; empty means today.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.quarter("2024-08-15")
    Keywords: q1, fiscal quarter, q3, three months
    Needs: none (pure function).
    """

def sec_to_hms(seconds: float) -> str:
    """Converts seconds to H:MM:SS (hours may exceed 24; fractions kept, e.g. 3725 -> 1:02:05).

        seconds: Duration in seconds.

    Example: dt.sec_to_hms(3725)
    Keywords: seconds to hours minutes seconds, format seconds, hh:mm:ss, clock, stopwatch, timecode, seconds to time
    Needs: none (pure function).
    """

def seconds_to_iso8601_duration(seconds: float) -> str:
    """Converts seconds to an ISO 8601 duration such as PT1H30M or P1DT2H (days are the largest unit).

        seconds: Duration in seconds.

    Example: dt.seconds_to_iso8601_duration(5400)
    Keywords: pt1h30m, xml duration, iso duration, period
    Needs: none (pure function).
    """

def start_of(text: str = "", unit: str = "day", week_start: str = "monday", tz: str = "", at: str = "", day_first: bool = False) -> str:
    """Returns the first instant of the minute, hour, day, week, month, quarter or year containing the date. Weeks begin on week_start (default Monday).

        text: Date/time text; empty means now.
        unit: minute, hour, day, week, month, quarter or year.
        week_start: First day of the week (monday or sunday...).
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.start_of("2024-05-17 13:45", "month", tz="UTC")
    Keywords: beginning of day, first of month, truncate date, beginning of month, first day of week, floor date, midnight
    Needs: none (pure function).
    """

def subtract(text: str, spec: str, tz: str = "", at: str = "", day_first: bool = False) -> str:
    """Subtracts a duration from a date (empty text = now); the same spec syntax as add(), so a negative spec adds.

        text: Date/time text; use now for the current time.
        spec: Duration spec such as 1y2mo3w4d5h6m7s, '-3 days', '2 hours 30 minutes'. Units: y mo w d h m s ms us ns. Months and years are calendar-aware and clamp to month end.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.subtract("2024-03-31", "1mo", "UTC")
    Keywords: date minus days, go back in time, days ago, date math, minus days, ago, date arithmetic, subtract days
    Needs: none (pure function).
    """

def time_math(expr: str) -> str:
    """Adds and subtracts times and durations: '1:30 + 2:45 - 0:20' gives 3:55. Terms are H:MM, H:MM:SS, unit forms like 1h30m or 45min, or ISO durations; the result is H:MM (or H:MM:SS when seconds remain), negative with a leading minus.

        expr: Expression with + and -.

    Example: dt.time_math("1:30 + 2:45 - 0:20")
    Keywords: add hours and minutes, timesheet, work hours, sum times, hour calculator
    Needs: none (pure function).
    """

def to_unix(text: str, unit: str = "s", tz: str = "", at: str = "", day_first: bool = False) -> int:
    """Converts any date text to a Unix timestamp in seconds (default), ms, us or ns, rounded toward negative infinity. Dates before 1970 give negative values.

        text: Date/time text.
        unit: s, ms, us or ns.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.to_unix("2024-01-31 15:04", "s", "UTC")
    Task: convert a date to a Unix timestamp
    Keywords: epoch seconds, date to epoch, timestamp, date to timestamp, seconds since 1970, strtotime
    Needs: none (pure function).
    """

def tz_info(name: str, at: str = "") -> Dict[str, str]:
    """Describes a time zone at an instant: UTC offset, abbreviation, whether DST is active, the local time there, and the start and end of the current offset period (the next DST change).

        name: IANA name, city, abbreviation or offset.
        at: Instant to inspect (empty = now).

    Example: dt.tz_info("Europe/Paris", "2024-07-01")
    Keywords: timezone, daylight saving, utc offset, dst, zone abbreviation, gmt offset
    Needs: none (pure function).
    """

def tz_list(filter: str = "") -> List[str]:
    """Lists the known IANA time zone names, optionally keeping those that contain filter (case-insensitive, spaces match underscores).

        filter: Substring such as asia or new_york; empty lists all.

    Example: dt.tz_list("america/arg")
    Keywords: time zone names, iana zones, iana, tz database, zone names, olson, find timezone
    Needs: none (pure function).
    """

def tz_offset(name: str, at: str = "") -> str:
    """Returns the UTC offset of a zone at an instant as +hh:mm (DST-aware, e.g. +01:00 or +02:00 for Europe/Paris depending on the date).

        name: IANA name, city, abbreviation or offset.
        at: Instant (empty = now).

    Example: dt.tz_offset("Asia/Kolkata")
    Keywords: gmt offset, utc offset, hours from utc, timezone difference
    Needs: none (pure function).
    """

def unix_hex(text: str = "", tz: str = "", at: str = "", day_first: bool = False) -> str:
    """Returns Unix seconds as 0x-prefixed hexadecimal.

        text: Date/time text; empty means now.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.unix_hex("2024-01-31", "UTC")
    Keywords: hex timestamp, epoch hex, base16
    Needs: none (pure function).
    """

def unix_ms_now() -> int:
    """Returns the current Unix time in milliseconds (JavaScript Date.now()).

    Example: dt.unix_ms_now()
    Keywords: epoch, timestamp, milliseconds, date.now, javascript
    Needs: none (pure function).
    """

def unix_now() -> int:
    """Returns the current Unix time in seconds.

    Example: dt.unix_now()
    Keywords: current epoch, timestamp now, epoch, timestamp, seconds since 1970, current unix time
    Needs: none (pure function).
    """

def webkit_time(text: str = "", tz: str = "", at: str = "", day_first: bool = False) -> int:
    """Returns the WebKit (Chrome) timestamp: microseconds since 1601-01-01 UTC, as used in Chrome history and cookie databases.

        text: Date/time text; empty means now.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.

    Example: dt.webkit_time("2024-01-31", "UTC")
    Keywords: chrome, chromium, cookies, history, sqlite
    Needs: none (pure function).
    """

def week_of_year(text: str = "", tz: str = "", at: str = "", day_first: bool = False, system: str = "iso") -> int:
    """Returns the week number of a date. system: iso (ISO 8601, weeks start Monday, week 1 has Jan 4), us (week 1 contains Jan 1, weeks start Sunday), sunday (strftime %U) or monday (strftime %W).

        text: Date text; empty means today.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.
        at: Reference instant for relative words and countdowns (anything parse() understands); empty means now.
        day_first: Read ambiguous numeric dates like 03/04/2024 as day/month instead of month/day.
        system: iso, us, sunday or monday.

    Example: dt.week_of_year("2024-12-30", system="iso")
    Keywords: week number, iso week, calendar week, kw, cw
    Needs: none (pure function).
    """

def world_clock(text: str = "", zones: List[str] = [], tz: str = "") -> Dict[str, str]:
    """Shows one instant in several time zones at once (default: a spread of major cities). The text is read in tz (default local); empty means now. Returns a dict zone -> 'YYYY-MM-DD HH:MM ZONE (Weekday)'.

        text: Date/time text; empty means now.
        zones: Zones to show; empty uses a default set.
        tz: Time zone: IANA name (Europe/Paris), city (Tokyo), abbreviation (EST) or offset (UTC+3:30, +05:30). Empty means the local zone.

    Example: dt.world_clock("2024-01-31 15:00", ["UTC", "Tokyo", "New York"], "Europe/Paris")
    Keywords: time in many cities, compare time zones, time zones, world time, multiple timezones, meeting, what time is it in
    Needs: none (pure function).
    """

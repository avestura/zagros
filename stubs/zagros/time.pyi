"""time: the standard Starlark time module (durations and times as time.Duration and time.Time values).

Type description, not runnable Starlark: do not `import` anything. The name `time` is predeclared in main.star; call `time.now()` directly. For string-based date work and time zones use the `dt` module.
Needs: none.
Keywords: clock, now, date, duration, timestamp, parse time, sleep, timezone
"""

from typing import Any

def now() -> Any:
    """Current time.

    Example: t = time.now()
    Task: get the current time as a value
    Keywords: today, current date, clock, right now, timestamp now
    Needs: none
    """

def parse_time(x: str, format: str = ..., location: str = ...) -> Any:
    """Parse a time string (RFC 3339 by default).

    Example: t = time.parse_time("2024-05-01T10:00:00Z")
    Keywords: parse date, string to time, read timestamp, iso 8601
    Needs: none
    """

def parse_duration(d: str) -> Any:
    """Parse a duration such as "1h30m".

    Example: d = time.parse_duration("1h30m")
    Keywords: duration string, time span, hours minutes, interval
    Needs: none
    """

def from_timestamp(sec: int, nsec: int = 0) -> Any:
    """Build a time from Unix seconds.

    Example: t = time.from_timestamp(1700000000)
    Keywords: unix to time, epoch to date, seconds since 1970
    Needs: none
    """

def time(year: int = 1970, month: int = 1, day: int = 1, hour: int = 0, minute: int = 0, second: int = 0, nanosecond: int = 0, location: str = "UTC") -> Any:
    """Build a time from calendar parts.

    Example: t = time.time(year = 2024, month = 5, day = 1)
    Keywords: make date, construct time, calendar date, create datetime
    Needs: none
    """

def is_valid_timezone(loc: str) -> bool:
    """Whether loc is a known IANA time zone name.

    Example: ok = time.is_valid_timezone("Europe/Berlin")
    Keywords: check time zone, valid tz name, iana zone
    Needs: none
    """

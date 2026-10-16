package market

import (
	"strconv"
	"strings"
)

type semver struct {
	nums [3]int
	pre  []string
	ok   bool
}

func parseSemver(s string) semver {
	s = strings.TrimPrefix(strings.TrimSpace(s), "v")
	if i := strings.IndexByte(s, '+'); i >= 0 {
		s = s[:i]
	}
	var v semver
	pre := ""
	if i := strings.IndexByte(s, '-'); i >= 0 {
		s, pre = s[:i], s[i+1:]
	}
	parts := strings.Split(s, ".")
	if len(parts) == 0 || len(parts) > 3 {
		return v
	}
	for i, p := range parts {
		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return semver{}
		}
		v.nums[i] = n
	}
	if pre != "" {
		v.pre = strings.Split(pre, ".")
	}
	v.ok = true
	return v
}

// CompareVersions compares two semver strings (leading "v" allowed, build
// metadata ignored). It returns -1, 0 or 1. Unparsable versions sort lowest.
func CompareVersions(a, b string) int {
	x, y := parseSemver(a), parseSemver(b)
	if !x.ok || !y.ok {
		switch {
		case x.ok:
			return 1
		case y.ok:
			return -1
		}
		return 0
	}
	for i := 0; i < 3; i++ {
		if x.nums[i] != y.nums[i] {
			if x.nums[i] < y.nums[i] {
				return -1
			}
			return 1
		}
	}
	// A version without prerelease is greater than one with.
	if len(x.pre) == 0 || len(y.pre) == 0 {
		switch {
		case len(x.pre) == len(y.pre):
			return 0
		case len(x.pre) == 0:
			return 1
		}
		return -1
	}
	for i := 0; i < len(x.pre) && i < len(y.pre); i++ {
		p, q := x.pre[i], y.pre[i]
		pn, perr := strconv.Atoi(p)
		qn, qerr := strconv.Atoi(q)
		switch {
		case perr == nil && qerr == nil:
			if pn != qn {
				if pn < qn {
					return -1
				}
				return 1
			}
		case perr == nil:
			return -1
		case qerr == nil:
			return 1
		case p != q:
			if p < q {
				return -1
			}
			return 1
		}
	}
	switch {
	case len(x.pre) < len(y.pre):
		return -1
	case len(x.pre) > len(y.pre):
		return 1
	}
	return 0
}

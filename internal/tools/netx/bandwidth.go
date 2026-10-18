package netx

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

// bitsPerUnit returns how many bits a size unit holds (b, kb, MiB, GB...).
func sizeUnitBits(u string) (float64, bool) {
	u = strings.TrimSpace(u)
	for _, w := range []struct{ from, to string }{{"bytes", "B"}, {"byte", "B"}, {"bits", "b"}, {"bit", "b"}} {
		if strings.HasSuffix(strings.ToLower(u), w.from) {
			u = u[:len(u)-len(w.from)] + w.to
			break
		}
	}
	if u == "" {
		return 0, false
	}
	last := u[len(u)-1]
	if last != 'B' && last != 'b' {
		return 0, false
	}
	base := 1.0
	if last == 'B' {
		base = 8
	}
	p := u[:len(u)-1]
	binary := false
	if strings.HasSuffix(p, "i") {
		binary = true
		p = p[:len(p)-1]
	}
	exp := -1
	switch strings.ToLower(p) {
	case "":
		exp = 0
	case "k":
		exp = 1
	case "m":
		exp = 2
	case "g":
		exp = 3
	case "t":
		exp = 4
	case "p":
		exp = 5
	}
	if exp < 0 || (binary && p == "") {
		return 0, false
	}
	// case matters only for b/B; "mb" is treated as megabit/megabyte by the last letter
	f := math.Pow(1000, float64(exp))
	if binary {
		f = math.Pow(1024, float64(exp))
	}
	return base * f, true
}

// rateUnitBits parses bps, Mbps, MB/s, MiB/s, Gbit/s...; result is bits per second.
func rateUnitBits(u string) (float64, error) {
	s := strings.TrimSpace(u)
	t := s
	switch {
	case strings.HasSuffix(t, "/s"):
		t = t[:len(t)-2]
	case strings.HasSuffix(t, "ps"):
		t = t[:len(t)-2]
	case strings.HasSuffix(strings.ToLower(t), "/sec"):
		t = t[:len(t)-4]
	}
	if t == "" {
		return 0, fmt.Errorf("unknown rate unit %q", u)
	}
	if t == "B" || t == "b" || t == "bit" || t == "byte" {
		// bare "bps" / "Bps"
	}
	if b, ok := sizeUnitBits(t); ok {
		return b, nil
	}
	return 0, fmt.Errorf("unknown rate unit %q (use bps, Kbps, Mbps, Gbps, B/s, KB/s, MB/s, MiB/s...)", u)
}

func parseQuantity(s string) (float64, string, error) {
	s = strings.TrimSpace(strings.ReplaceAll(s, ",", ""))
	i := 0
	for i < len(s) && (s[i] >= '0' && s[i] <= '9' || s[i] == '.' || s[i] == '-' || s[i] == '+' || s[i] == 'e' && i > 0 && i+1 < len(s) && (s[i+1] >= '0' && s[i+1] <= '9' || s[i+1] == '-')) {
		i++
	}
	n, err := strconv.ParseFloat(s[:i], 64)
	if err != nil {
		return 0, "", fmt.Errorf("cannot read a number from %q", s)
	}
	return n, strings.TrimSpace(s[i:]), nil
}

func humanDuration(sec float64) string {
	if sec < 0 || math.IsNaN(sec) || math.IsInf(sec, 0) {
		return "n/a"
	}
	if sec < 1 {
		return fmt.Sprintf("%.0f ms", sec*1000)
	}
	if sec < 60 {
		return strings.TrimSuffix(strings.TrimRight(fmt.Sprintf("%.2f", sec), "0"), ".") + " s"
	}
	total := int64(math.Round(sec))
	d, h, m, s := total/86400, total%86400/3600, total%3600/60, total%60
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
		parts = append(parts, fmt.Sprintf("%ds", s))
	}
	return strings.Join(parts, " ")
}

func fmtNum(f float64) string {
	return strconv.FormatFloat(f, 'f', -1, 64)
}

func init() {
	reg(kwBand, tools.Func{
		Name: "bandwidth", Title: "Convert bandwidth units", Order: 120, Input: "from_unit",
		Doc:     "Converts a data rate between units: bps, Kbps, Mbps, Gbps, Tbps (bits, decimal), B/s, KB/s, MB/s, GB/s (bytes, decimal) and KiB/s, MiB/s, GiB/s (bytes, binary). A capital B means bytes, lower-case b means bits; /s or ps both work.",
		Params:  []tools.Param{req("value", "float", "Rate in from_unit."), req("from_unit", "str", `Unit such as "Mbps" or "MB/s".`), req("to_unit", "str", `Target unit such as "MiB/s".`)},
		Returns: "float", Example: `net.bandwidth(100, "Mbps", "MB/s")`,
		Tags: []string{"mbps to mb/s", "megabit", "megabyte per second", "speed conversion", "data rate", "throughput", "download speed"},
		Fn: func(a tools.Args) (any, error) {
			f, err := rateUnitBits(a.Str("from_unit"))
			if err != nil {
				return nil, err
			}
			t, err := rateUnitBits(a.Str("to_unit"))
			if err != nil {
				return nil, err
			}
			return a.Float("value") * f / t, nil
		},
	})
	reg(kwBand, tools.Func{
		Name: "transfer_time", Title: "Transfer time", Order: 121, Input: "size",
		Doc:     `How long a transfer takes: size such as "1.5 GB" or "700 MiB" at a bandwidth such as "100 Mbps" or "12 MB/s". Returns a human duration like "2m 4s". Protocol overhead is ignored (ideal line rate).`,
		Params:  []tools.Param{req("size", "str", `Amount of data, e.g. "4.7 GB".`), req("bandwidth", "str", `Rate, e.g. "50 Mbps".`)},
		Returns: "str", Example: `net.transfer_time("4.7 GB", "100 Mbps")`,
		Tags: []string{"download time", "how long", "eta", "upload time", "duration", "copy time"},
		Accepts: func(s string) bool {
			_, u, err := parseQuantity(s)
			if err != nil {
				return false
			}
			_, ok := sizeUnitBits(u)
			return ok
		},
		Fn: func(a tools.Args) (any, error) {
			n, u, err := parseQuantity(a.Str("size"))
			if err != nil {
				return nil, err
			}
			sb, ok := sizeUnitBits(u)
			if !ok {
				return nil, fmt.Errorf("unknown size unit %q (use B, KB, MB, GB, KiB, MiB, GiB, b, Mb...)", u)
			}
			r, ru, err := parseQuantity(a.Str("bandwidth"))
			if err != nil {
				return nil, err
			}
			rb, err := rateUnitBits(ru)
			if err != nil {
				return nil, err
			}
			if r <= 0 || n < 0 {
				return nil, fmt.Errorf("size must be >= 0 and bandwidth > 0")
			}
			return humanDuration(n * sb / (r * rb)), nil
		},
	})
	reg(kwBand, tools.Func{
		Name: "throughput_needed", Title: "Bandwidth needed", Order: 122, Input: "size",
		Doc:     `Bandwidth required to move a size (e.g. "10 GB") in the given number of seconds. Returns the rate in bps, Kbps, Mbps, Gbps, MB/s and MiB/s as strings.`,
		Params:  []tools.Param{req("size", "str", `Amount of data, e.g. "10 GB".`), req("seconds", "float", "Time allowed in seconds (> 0).")},
		Returns: "dict[str, str]", Example: `net.throughput_needed("10 GB", 600)`,
		Tags: []string{"required speed", "minimum bandwidth", "how fast", "mbps needed"},
		Accepts: func(s string) bool {
			_, u, err := parseQuantity(s)
			if err != nil {
				return false
			}
			_, ok := sizeUnitBits(u)
			return ok
		},
		Fn: func(a tools.Args) (any, error) {
			n, u, err := parseQuantity(a.Str("size"))
			if err != nil {
				return nil, err
			}
			sb, ok := sizeUnitBits(u)
			if !ok {
				return nil, fmt.Errorf("unknown size unit %q", u)
			}
			if a.Float("seconds") <= 0 {
				return nil, fmt.Errorf("seconds must be > 0")
			}
			bps := n * sb / a.Float("seconds")
			rnd := func(f float64) string { return fmtNum(math.Round(f*1000) / 1000) }
			return map[string]string{
				"bps": rnd(bps), "Kbps": rnd(bps / 1e3), "Mbps": rnd(bps / 1e6), "Gbps": rnd(bps / 1e9),
				"MB/s": rnd(bps / 8e6), "MiB/s": rnd(bps / 8 / 1048576),
			}, nil
		},
	})
}

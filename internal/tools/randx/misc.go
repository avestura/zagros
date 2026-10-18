package randx

import (
	"fmt"
	"math"
	"math/rand/v2"
	"net/netip"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/avestura/zagros/internal/tools"
)

func hslToRGB(h, s, l float64) (int, int, int) {
	h = math.Mod(h, 360)
	if h < 0 {
		h += 360
	}
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	f := func(v float64) int { return int(math.Round((v + m) * 255)) }
	return f(r), f(g), f(b)
}

// niceColor returns a pleasant, vivid colour as hue/saturation/lightness.
func niceColor(r *rand.Rand, hue float64) (h, s, l float64) {
	return hue, 0.55 + r.Float64()*0.25, 0.45 + r.Float64()*0.15
}

func hexOf(r, g, b int) string { return fmt.Sprintf("#%02x%02x%02x", r, g, b) }

func colorParams() []tools.Param {
	return []tools.Param{p("pure", "bool", "Pure random RGB instead of a pleasant vivid colour.", false), seedParam()}
}

func drawColor(a tools.Args) (r, g, b int, h, s, l float64) {
	rr := newRng(a.Int("seed"))
	if a.Bool("pure") {
		r, g, b = rr.IntN(256), rr.IntN(256), rr.IntN(256)
		h, s, l = rgbToHSL(r, g, b)
		return
	}
	h, s, l = niceColor(rr, rr.Float64()*360)
	r, g, b = hslToRGB(h, s, l)
	return
}

func rgbToHSL(r, g, b int) (h, s, l float64) {
	rf, gf, bf := float64(r)/255, float64(g)/255, float64(b)/255
	mx, mn := math.Max(rf, math.Max(gf, bf)), math.Min(rf, math.Min(gf, bf))
	l = (mx + mn) / 2
	if mx == mn {
		return 0, 0, l
	}
	d := mx - mn
	if l > 0.5 {
		s = d / (2 - mx - mn)
	} else {
		s = d / (mx + mn)
	}
	switch mx {
	case rf:
		h = (gf - bf) / d
		if gf < bf {
			h += 6
		}
	case gf:
		h = (bf-rf)/d + 2
	default:
		h = (rf-gf)/d + 4
	}
	return h * 60, s, l
}

var colorNames = [][2]string{
	{"coral", "#ff7f50"}, {"crimson", "#dc143c"}, {"gold", "#ffd700"}, {"khaki", "#f0e68c"}, {"orchid", "#da70d6"},
	{"salmon", "#fa8072"}, {"tomato", "#ff6347"}, {"turquoise", "#40e0d0"}, {"violet", "#ee82ee"}, {"teal", "#008080"},
	{"indigo", "#4b0082"}, {"lavender", "#e6e6fa"}, {"lime", "#00ff00"}, {"maroon", "#800000"}, {"navy", "#000080"},
	{"olive", "#808000"}, {"orange", "#ffa500"}, {"plum", "#dda0dd"}, {"sienna", "#a0522d"}, {"skyblue", "#87ceeb"},
	{"slateblue", "#6a5acd"}, {"steelblue", "#4682b4"}, {"tan", "#d2b48c"}, {"thistle", "#d8bfd8"}, {"wheat", "#f5deb3"},
	{"chocolate", "#d2691e"}, {"cornflowerblue", "#6495ed"}, {"firebrick", "#b22222"}, {"forestgreen", "#228b22"}, {"hotpink", "#ff69b4"},
	{"mediumseagreen", "#3cb371"}, {"peru", "#cd853f"}, {"rebeccapurple", "#663399"}, {"seagreen", "#2e8b57"}, {"darkorange", "#ff8c00"},
}

var firstNames = map[string][]string{
	"male":   strings.Fields(`James John Robert Michael William David Richard Joseph Thomas Daniel Matthew Anthony Mark Paul Steven Andrew Kenneth Joshua Kevin Brian George Edward Ryan Jacob Nathan Omar Ali Hassan Liam Noah Ethan Lucas Mateo Hiroshi Arjun Pierre Hans Ivan Marco Dmitri Kaveh Reza`),
	"female": strings.Fields(`Mary Patricia Jennifer Linda Elizabeth Barbara Susan Jessica Sarah Karen Nancy Lisa Margaret Sandra Ashley Emily Donna Michelle Carol Amanda Laura Olivia Emma Sophia Mia Isabella Amelia Chloe Zoe Fatima Aisha Yuki Priya Elena Giulia Anna Sofia Maria Leila Shirin Nora`),
}

var lastNames = strings.Fields(`Smith Johnson Williams Brown Jones Garcia Miller Davis Rodriguez Martinez Hernandez Lopez Gonzalez Wilson Anderson Thomas Taylor Moore Jackson Martin Lee Perez Thompson White Harris Sanchez Clark Ramirez Lewis Robinson Walker Young Allen King Wright Scott Torres Nguyen Hill Flores Green Adams Nelson Baker Hall Rivera Campbell Mitchell Carter Roberts Tanaka Sato Kim Park Chen Wang Patel Singh Khan Ahmadi Rossi Muller Schmidt Dubois Ivanov Kowalski Nilsson Silva Costa`)

var loremWords = strings.Fields(`lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua enim ad minim veniam quis nostrud exercitation ullamco laboris nisi aliquip ex ea commodo consequat duis aute irure in reprehenderit voluptate velit esse cillum fugiat nulla pariatur excepteur sint occaecat cupidatat non proident sunt culpa qui officia deserunt mollit anim id est laborum`)

var emojis = []string{
	"😀", "😂", "😍", "😎", "🤔", "😴", "🥳", "😇", "🤖", "👻", "💩", "🔥", "✨", "🌈", "⭐", "🌙", "☀️", "🌊", "🌸", "🌵",
	"🍎", "🍕", "🍔", "🍣", "🍩", "🍪", "☕", "🍺", "🥑", "🌮", "🐶", "🐱", "🦊", "🐼", "🐸", "🦄", "🐙", "🦋", "🐢", "🦉",
	"⚽", "🏀", "🎮", "🎲", "🎸", "🎨", "📚", "💡", "🔑", "🚀", "✈️", "🚗", "🏝️", "🏔️", "🎁", "🎈", "🎉", "💎", "🧩", "🧠",
}

func init() {
	reg(tools.Func{
		Name: "color_hex", Title: "Random colour (hex)", Order: 12, Input: "args", Keywords: []string{"color", "colour", "hex"},
		Doc:     "Returns a random colour as #rrggbb. By default a pleasant vivid colour; pure=true draws uniform RGB.",
		Params:  append(colorParams(), argsParam("ignored.")),
		Returns: "str", Example: `rand.color_hex()`,
		Tags: []string{"random color", "hex color", "css", "rgb", "swatch"},
		Fn: func(a tools.Args) (any, error) {
			r, g, b, _, _, _ := drawColor(a)
			return hexOf(r, g, b), nil
		},
	})
	reg(tools.Func{
		Name: "color_rgb", Title: "Random colour (rgb)",
		Doc:     "Returns a random colour as a CSS rgb(r, g, b) string; pleasant vivid by default, uniform with pure=true.",
		Params:  colorParams(),
		Returns: "str", Example: `rand.color_rgb()`,
		Tags: []string{"random color", "rgb", "css"},
		Fn: func(a tools.Args) (any, error) {
			r, g, b, _, _, _ := drawColor(a)
			return fmt.Sprintf("rgb(%d, %d, %d)", r, g, b), nil
		},
	})
	reg(tools.Func{
		Name: "color_hsl", Title: "Random colour (hsl)",
		Doc:     "Returns a random colour as a CSS hsl(h, s%, l%) string; pleasant vivid by default, uniform with pure=true.",
		Params:  colorParams(),
		Returns: "str", Example: `rand.color_hsl()`,
		Tags: []string{"random color", "hsl", "css", "hue"},
		Fn: func(a tools.Args) (any, error) {
			_, _, _, h, s, l := drawColor(a)
			return fmt.Sprintf("hsl(%d, %d%%, %d%%)", int(math.Round(h)), int(math.Round(s*100)), int(math.Round(l*100))), nil
		},
	})
	reg(tools.Func{
		Name: "color_palette", Title: "Random colour palette",
		Doc:     "Returns n (1..100) harmonious hex colours: hues are spaced by the golden angle from a random start so neighbours stay distinct.",
		Params:  []tools.Param{p("n", "int", "Number of colours.", int64(5)), seedParam()},
		Returns: "list[str]", Example: `rand.color_palette(5)`,
		Tags: []string{"palette", "color scheme", "theme", "swatches", "design"},
		Fn: func(a tools.Args) (any, error) {
			n := a.Int("n")
			if n < 1 || n > 100 {
				return nil, fmt.Errorf("n must be 1..100")
			}
			r := newRng(a.Int("seed"))
			h := r.Float64() * 360
			out := make([]string, n)
			for i := range out {
				hh, s, l := niceColor(r, h)
				out[i] = hexOf(hslToRGB(hh, s, l))
				h += 137.508
			}
			return out, nil
		},
	})
	reg(tools.Func{
		Name: "color_name", Title: "Random named colour",
		Doc:     "Returns a random CSS colour name with its hex value, e.g. \"coral #ff7f50\".",
		Params:  []tools.Param{seedParam()},
		Returns: "str", Example: `rand.color_name()`,
		Tags: []string{"css color names", "named color", "color"},
		Fn: func(a tools.Args) (any, error) {
			c := colorNames[newRng(a.Int("seed")).IntN(len(colorNames))]
			return c[0] + " " + c[1], nil
		},
	})

	// ---- network ----
	reg(tools.Func{
		Name: "ip_v4", Title: "Random IPv4 address",
		Doc:     "Returns a random IPv4 address. private=true picks from 10/8, 172.16/12 or 192.168/16; otherwise a routable public-looking address that avoids loopback, multicast, link-local, CGNAT, documentation and reserved ranges.",
		Params:  []tools.Param{p("private", "bool", "Use an RFC 1918 private address.", false), seedParam()},
		Returns: "str", Example: `rand.ip_v4(private=True)`,
		Tags: []string{"ip address", "ipv4", "rfc1918", "test data", "network"},
		Fn: func(a tools.Args) (any, error) {
			r := newRng(a.Int("seed"))
			if a.Bool("private") {
				switch r.IntN(3) {
				case 0:
					return fmt.Sprintf("10.%d.%d.%d", r.IntN(256), r.IntN(256), r.IntN(254)+1), nil
				case 1:
					return fmt.Sprintf("172.%d.%d.%d", 16+r.IntN(16), r.IntN(256), r.IntN(254)+1), nil
				}
				return fmt.Sprintf("192.168.%d.%d", r.IntN(256), r.IntN(254)+1), nil
			}
			for {
				b := [4]byte{byte(r.IntN(224)), byte(r.IntN(256)), byte(r.IntN(256)), byte(r.IntN(254) + 1)}
				ip := netip.AddrFrom4(b)
				if b[0] == 0 || ip.IsPrivate() || ip.IsLoopback() || ip.IsMulticast() || ip.IsLinkLocalUnicast() ||
					(b[0] == 100 && b[1]&0xc0 == 64) || (b[0] == 192 && b[1] == 0 && (b[2] == 0 || b[2] == 2)) ||
					(b[0] == 198 && (b[1] == 18 || b[1] == 19)) || (b[0] == 198 && b[1] == 51 && b[2] == 100) ||
					(b[0] == 203 && b[1] == 0 && b[2] == 113) || (b[0] == 192 && b[1] == 88 && b[2] == 99) {
					continue
				}
				return ip.String(), nil
			}
		},
	})
	reg(tools.Func{
		Name: "ip_v6", Title: "Random IPv6 address",
		Doc:     "Returns a random IPv6 address in compressed form. private=true gives a unique-local fd00::/8 address; otherwise a global unicast address in 2000::/3 outside the documentation prefix.",
		Params:  []tools.Param{p("private", "bool", "Use a unique local (fd00::/8) address.", false), seedParam()},
		Returns: "str", Example: `rand.ip_v6()`,
		Tags: []string{"ip address", "ipv6", "ula", "test data", "network"},
		Fn: func(a tools.Args) (any, error) {
			r := newRng(a.Int("seed"))
			for {
				var b [16]byte
				for i := range b {
					b[i] = byte(r.Uint32())
				}
				if a.Bool("private") {
					b[0] = 0xfd
				} else {
					b[0] = 0x20 | b[0]&0x1f
					if b[0] == 0x20 && b[1] == 0x01 && b[2] == 0x0d && b[3] == 0xb8 {
						continue
					}
				}
				return netip.AddrFrom16(b).String(), nil
			}
		},
	})
	reg(tools.Func{
		Name: "mac_address", Title: "Random MAC address",
		Doc:     "Returns a random unicast, locally administered MAC address. Format is colon (aa:bb:..), dash (AA-BB-..), dot (aabb.ccdd.eeff) or plain (aabbccddeeff); upper switches to uppercase.",
		Params:  []tools.Param{p("format", "str", "colon, dash, dot or plain.", "colon"), p("upper", "bool", "Uppercase hex digits.", false), seedParam()},
		Returns: "str", Example: `rand.mac_address("dash", upper=True)`,
		Tags: []string{"mac", "hardware address", "ethernet", "oui", "test data"},
		Fn: func(a tools.Args) (any, error) {
			r := newRng(a.Int("seed"))
			var b [6]byte
			for i := range b {
				b[i] = byte(r.Uint32())
			}
			b[0] = b[0]&0xfc | 0x02
			h := fmt.Sprintf("%02x%02x%02x%02x%02x%02x", b[0], b[1], b[2], b[3], b[4], b[5])
			var s string
			switch strings.ToLower(a.Str("format")) {
			case "colon", "":
				s = h[0:2] + ":" + h[2:4] + ":" + h[4:6] + ":" + h[6:8] + ":" + h[8:10] + ":" + h[10:12]
			case "dash", "hyphen":
				s = h[0:2] + "-" + h[2:4] + "-" + h[4:6] + "-" + h[6:8] + "-" + h[8:10] + "-" + h[10:12]
			case "dot", "cisco":
				s = h[0:4] + "." + h[4:8] + "." + h[8:12]
			case "plain":
				s = h
			default:
				return nil, fmt.Errorf("format must be colon, dash, dot or plain")
			}
			if a.Bool("upper") {
				s = strings.ToUpper(s)
			}
			return s, nil
		},
	})
	reg(tools.Func{
		Name: "port", Title: "Random port number",
		Doc:     "Returns a random TCP/UDP port. kind: any (1..65535), well_known (1..1023), registered (1024..49151), dynamic (49152..65535) or user (1024..65535).",
		Params:  []tools.Param{p("kind", "str", "any, well_known, registered, dynamic or user.", "dynamic"), seedParam()},
		Returns: "int", Example: `rand.port("registered")`,
		Tags: []string{"tcp", "udp", "network port", "ephemeral", "dev server"},
		Fn: func(a tools.Args) (any, error) {
			var lo, hi int64
			switch strings.ToLower(a.Str("kind")) {
			case "any":
				lo, hi = 1, 65535
			case "well_known", "wellknown", "system":
				lo, hi = 1, 1023
			case "registered":
				lo, hi = 1024, 49151
			case "dynamic", "ephemeral", "private", "":
				lo, hi = 49152, 65535
			case "user":
				lo, hi = 1024, 65535
			default:
				return nil, fmt.Errorf("kind must be any, well_known, registered, dynamic or user")
			}
			return intRange(newRng(a.Int("seed")), lo, hi), nil
		},
	})

	// ---- dates and times ----
	reg(tools.Func{
		Name: "random_date", Title: "Random date",
		Doc:     "Returns a random calendar date between start and end inclusive (YYYY-MM-DD or RFC 3339 inputs, UTC). With with_time=true the result is an RFC 3339 timestamp, otherwise YYYY-MM-DD. Errors when start is after end.",
		Params:  []tools.Param{p("start", "str", "Earliest date.", "1970-01-01"), p("end", "str", "Latest date.", "2030-12-31"), p("with_time", "bool", "Include a random time of day.", false), seedParam()},
		Returns: "str", Example: `rand.random_date("2020-01-01", "2020-12-31")`,
		Tags: []string{"random day", "birthday", "test data", "timestamp"},
		Fn: func(a tools.Args) (any, error) {
			s, err := parseDateArg(a.Str("start"))
			if err != nil {
				return nil, err
			}
			e, err := parseDateArg(a.Str("end"))
			if err != nil {
				return nil, err
			}
			if !a.Bool("with_time") {
				e = e.Add(time.Hour * 24)
				s = time.Date(s.Year(), s.Month(), s.Day(), 0, 0, 0, 0, time.UTC)
				e = time.Date(e.Year(), e.Month(), e.Day(), 0, 0, 0, 0, time.UTC)
				days := int64(e.Sub(s).Hours() / 24)
				if days <= 0 {
					return nil, fmt.Errorf("start is after end")
				}
				d := s.AddDate(0, 0, int(intRange(newRng(a.Int("seed")), 0, days-1)))
				return d.Format("2006-01-02"), nil
			}
			if e.Before(s) {
				return nil, fmt.Errorf("start is after end")
			}
			sec := intRange(newRng(a.Int("seed")), s.Unix(), e.Unix())
			return time.Unix(sec, 0).UTC().Format(time.RFC3339), nil
		},
	})
	reg(tools.Func{
		Name: "random_time", Title: "Random time of day",
		Doc:     "Returns a random time of day between start and end (HH:MM or HH:MM:SS, inclusive) as HH:MM, or HH:MM:SS when seconds=true. Errors when start is after end.",
		Params:  []tools.Param{p("start", "str", "Earliest time.", "00:00"), p("end", "str", "Latest time.", "23:59"), p("seconds", "bool", "Include seconds.", false), seedParam()},
		Returns: "str", Example: `rand.random_time("09:00", "17:30")`,
		Tags: []string{"random clock time", "hour minute", "test data"},
		Fn: func(a tools.Args) (any, error) {
			s, err := parseClock(a.Str("start"))
			if err != nil {
				return nil, err
			}
			e, err := parseClock(a.Str("end"))
			if err != nil {
				return nil, err
			}
			if s > e {
				return nil, fmt.Errorf("start is after end")
			}
			if !a.Bool("seconds") {
				s, e = s/60, e/60
				v := intRange(newRng(a.Int("seed")), int64(s), int64(e))
				return fmt.Sprintf("%02d:%02d", v/60, v%60), nil
			}
			v := intRange(newRng(a.Int("seed")), int64(s), int64(e))
			return fmt.Sprintf("%02d:%02d:%02d", v/3600, v/60%60, v%60), nil
		},
	})

	// ---- lottery, sequences ----
	reg(tools.Func{
		Name: "lottery", Title: "Lottery numbers", Order: 15, Input: "args", Keywords: []string{"lotto", "lottery"},
		Doc:     "Draws count distinct numbers from 1..max in ascending order, plus an optional bonus number drawn from the remaining pool. Returns numbers (space separated) and bonus.",
		Params:  []tools.Param{p("count", "int", "Numbers to draw.", int64(6)), p("max", "int", "Highest number in the pool.", int64(49)), p("bonus", "int", "Bonus balls to draw (0 or more) from the rest of the pool.", int64(0)), seedParam(), argsParam("'6 49' or '5 69 1'.")},
		Returns: "dict[str, str]", Example: `rand.lottery(6, 49, 1)`,
		Tags: []string{"lotto", "powerball", "draw", "numbers", "jackpot"},
		Fn: func(a tools.Args) (any, error) {
			count, mx, bonus := a.Int("count"), a.Int("max"), a.Int("bonus")
			if v := numArgs(a.Str("args")); len(v) >= 2 {
				count, mx = v[0], v[1]
				if len(v) >= 3 {
					bonus = v[2]
				}
			}
			if err := checkCount(count, "count"); err != nil {
				return nil, err
			}
			if err := checkCount(bonus, "bonus"); err != nil {
				return nil, err
			}
			if mx < 1 || mx > 1000000 {
				return nil, fmt.Errorf("max must be 1..1000000")
			}
			all, err := distinctInts(newRng(a.Int("seed")), count+bonus, 1, mx)
			if err != nil {
				return nil, err
			}
			main := append([]int64(nil), all[:count]...)
			sort.Slice(main, func(i, j int) bool { return main[i] < main[j] })
			strs := func(v []int64) string {
				p := make([]string, len(v))
				for i, x := range v {
					p[i] = strconv.FormatInt(x, 10)
				}
				return strings.Join(p, " ")
			}
			return map[string]string{"numbers": strs(main), "bonus": strs(all[count:])}, nil
		},
	})
	reg(tools.Func{
		Name: "number_sequence", Title: "Random number sequence",
		Doc:     "Generates count integers as an ascending list, a descending list (kind) or a random walk starting at start with steps of at most max_step. Ascending/descending draw from [start, start + count*max_step].",
		Params:  []tools.Param{p("count", "int", "Length of the sequence (0..10000).", int64(10)), p("start", "int", "First value / lower bound.", int64(0)), p("max_step", "int", "Largest step between neighbours.", int64(10)), p("kind", "str", "ascending, descending or walk.", "ascending"), seedParam()},
		Returns: "list[float]", Example: `rand.number_sequence(8, 0, 5, "walk")`,
		Tags: []string{"random series", "monotonic", "random walk", "test data", "time series"},
		Fn: func(a tools.Args) (any, error) {
			n, start, step := a.Int("count"), a.Int("start"), a.Int("max_step")
			if err := checkCount(n, "count"); err != nil {
				return nil, err
			}
			if step < 0 || step > 1e9 || start > 1e15 || start < -1e15 {
				return nil, fmt.Errorf("max_step must be 0..1e9 and start within +-1e15")
			}
			r := newRng(a.Int("seed"))
			out := make([]float64, n)
			switch strings.ToLower(a.Str("kind")) {
			case "walk":
				v := start
				for i := range out {
					out[i] = float64(v)
					v += intRange(r, -step, step)
				}
			case "ascending", "asc", "descending", "desc", "":
				for i := range out {
					out[i] = float64(intRange(r, start, start+n*step))
				}
				sort.Float64s(out)
				if k := strings.ToLower(a.Str("kind")); strings.HasPrefix(k, "desc") {
					for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
						out[i], out[j] = out[j], out[i]
					}
				}
			default:
				return nil, fmt.Errorf("kind must be ascending, descending or walk")
			}
			return out, nil
		},
	})

	// ---- people and text ----
	nameFn := func(r *rand.Rand, gender string) (string, error) {
		switch strings.ToLower(gender) {
		case "male", "m":
			return firstNames["male"][r.IntN(len(firstNames["male"]))], nil
		case "female", "f":
			return firstNames["female"][r.IntN(len(firstNames["female"]))], nil
		case "any", "":
			g := firstNames["male"]
			if r.IntN(2) == 0 {
				g = firstNames["female"]
			}
			return g[r.IntN(len(g))], nil
		}
		return "", fmt.Errorf("gender must be any, male or female")
	}
	genderP := p("gender", "str", "any, male or female.", "any")
	reg(tools.Func{
		Name: "name_first", Title: "Random first name",
		Doc:     "Returns a random given name from a small built-in list; gender is any, male or female.",
		Params:  []tools.Param{genderP, seedParam()},
		Returns: "str", Example: `rand.name_first("female")`,
		Tags: []string{"given name", "fake name", "test data", "persona"},
		Fn:   func(a tools.Args) (any, error) { return nameFn(newRng(a.Int("seed")), a.Str("gender")) },
	})
	reg(tools.Func{
		Name: "name_last", Title: "Random last name",
		Doc:     "Returns a random family name from a small built-in list.",
		Params:  []tools.Param{seedParam()},
		Returns: "str", Example: `rand.name_last()`,
		Tags: []string{"surname", "family name", "fake name", "test data"},
		Fn:   func(a tools.Args) (any, error) { return lastNames[newRng(a.Int("seed")).IntN(len(lastNames))], nil },
	})
	reg(tools.Func{
		Name: "full_name", Title: "Random full name", Order: 14, Input: "args", Keywords: []string{"name", "person"},
		Doc:     "Returns a random \"First Last\" name from small built-in lists; gender is any, male or female.",
		Params:  []tools.Param{genderP, seedParam(), argsParam("male, female or any.")},
		Returns: "str", Example: `rand.full_name()`,
		Tags: []string{"fake name", "test data", "persona", "dummy user"},
		Fn: func(a tools.Args) (any, error) {
			r := newRng(a.Int("seed"))
			g := a.Str("gender")
			if s := strings.TrimSpace(a.Str("args")); s != "" {
				g = s
			}
			f, err := nameFn(r, g)
			if err != nil {
				return nil, err
			}
			return f + " " + lastNames[r.IntN(len(lastNames))], nil
		},
	})
	reg(tools.Func{
		Name: "email_fake", Title: "Fake email address",
		Doc:     "Returns a plausible fake email such as \"jane.smith42@example.com\". The default domain example.com is reserved for documentation so no real mailbox is hit.",
		Params:  []tools.Param{p("domain", "str", "Domain part.", "example.com"), seedParam()},
		Returns: "str", Example: `rand.email_fake("example.org")`,
		Tags: []string{"dummy email", "test data", "placeholder", "mock"},
		Fn: func(a tools.Args) (any, error) {
			r := newRng(a.Int("seed"))
			f, _ := nameFn(r, "any")
			l := lastNames[r.IntN(len(lastNames))]
			d := a.Str("domain")
			if d == "" {
				d = "example.com"
			}
			return strings.ToLower(f+"."+l) + strconv.Itoa(r.IntN(100)) + "@" + d, nil
		},
	})
	reg(tools.Func{
		Name: "lorem_word", Title: "Random lorem word",
		Doc:     "Returns one random lorem ipsum word. For paragraphs use the text group.",
		Params:  []tools.Param{seedParam()},
		Returns: "str", Example: `rand.lorem_word()`,
		Tags: []string{"lorem ipsum", "placeholder", "dummy text"},
		Fn:   func(a tools.Args) (any, error) { return loremWords[newRng(a.Int("seed")).IntN(len(loremWords))], nil },
	})
	reg(tools.Func{
		Name: "lorem_words", Title: "Random lorem words",
		Doc:     "Returns count random lorem ipsum words (0..10000) joined by single spaces; capitalize upper-cases the first letter.",
		Params:  []tools.Param{p("count", "int", "Number of words.", int64(5)), p("capitalize", "bool", "Capitalize the first word.", false), seedParam()},
		Returns: "str", Example: `rand.lorem_words(6, True)`,
		Tags: []string{"lorem ipsum", "placeholder", "dummy text", "filler"},
		Fn: func(a tools.Args) (any, error) {
			if err := checkCount(a.Int("count"), "count"); err != nil {
				return nil, err
			}
			r := newRng(a.Int("seed"))
			w := make([]string, a.Int("count"))
			for i := range w {
				w[i] = loremWords[r.IntN(len(loremWords))]
			}
			s := strings.Join(w, " ")
			if a.Bool("capitalize") && s != "" {
				s = strings.ToUpper(s[:1]) + s[1:]
			}
			return s, nil
		},
	})
	reg(tools.Func{
		Name: "random_emoji", Title: "Random emoji", Order: 16, Input: "args", Keywords: []string{"emoji"},
		Doc:     "Returns n random emoji (1..1000) from a built-in set of faces, animals, food, objects and symbols, concatenated without separator.",
		Params:  []tools.Param{p("n", "int", "How many emoji.", int64(1)), seedParam(), argsParam("a count such as '5'.")},
		Returns: "str", Example: `rand.random_emoji(3)`,
		Tags: []string{"emoji", "smiley", "unicode", "icon"},
		Fn: func(a tools.Args) (any, error) {
			n := a.Int("n")
			if v := numArgs(a.Str("args")); len(v) > 0 {
				n = v[0]
			}
			if n < 1 || n > 1000 {
				return nil, fmt.Errorf("n must be 1..1000")
			}
			r := newRng(a.Int("seed"))
			var b strings.Builder
			for i := int64(0); i < n; i++ {
				b.WriteString(emojis[r.IntN(len(emojis))])
			}
			return b.String(), nil
		},
	})
}

func parseDateArg(s string) (time.Time, error) {
	s = strings.TrimSpace(s)
	for _, l := range []string{"2006-01-02", time.RFC3339, "2006-01-02 15:04:05", "2006-01-02T15:04:05", "2006-01-02 15:04"} {
		if t, err := time.ParseInLocation(l, s, time.UTC); err == nil {
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot read date %q (use YYYY-MM-DD or RFC 3339)", s)
}

// parseClock returns seconds since midnight.
func parseClock(s string) (int, error) {
	s = strings.TrimSpace(s)
	for _, l := range []string{"15:04", "15:04:05"} {
		if t, err := time.Parse(l, s); err == nil {
			return t.Hour()*3600 + t.Minute()*60 + t.Second(), nil
		}
	}
	return 0, fmt.Errorf("cannot read time %q (use HH:MM or HH:MM:SS)", s)
}

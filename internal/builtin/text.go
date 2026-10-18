package builtin

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/avestura/zagros/internal/core"
)

type text struct {
	base
	// skip lists keywords the tools provider answers instead (it is richer),
	// so both providers never show duplicate rows.
	skip map[string]bool
}

func newText() *text { return &text{base: base{"text"}} }

// handing keywords over to the tools provider
var toolsOwnedKeywords = []string{"uuid", "b64", "b64d", "url", "urld", "hash", "case", "lorem", "ts", "now"}

func (t *text) Ranked() bool { return true }

var textKeywords = []string{"uuid", "b64", "b64d", "url", "urld", "hash", "case", "lorem", "ts", "now", "json", "color", "unit"}

func (t *text) Keywords() []string {
	if len(t.skip) == 0 {
		return textKeywords
	}
	var out []string
	for _, k := range textKeywords {
		if !t.skip[k] {
			out = append(out, k)
		}
	}
	return out
}

type kv struct{ label, value string }

func (t *text) Query(ctx context.Context, q string) []core.Item {
	rest, kw, ok := trigger(q, t.Keywords()...)
	if !ok {
		// bare "10 km to mi"
		if v, f, to, ok := parseUnitQuery(q); ok && isKnownUnit(f) && isKnownUnit(to) {
			return kvItems("unit", unitResults(v, f, to), "Unit")
		}
		return nil
	}
	var rows []kv
	cat := "Text"
	switch strings.ToLower(kw) {
	case "uuid":
		u := newUUID()
		rows = []kv{{"UUID v4", u}, {"UPPERCASE", strings.ToUpper(u)}, {"No dashes", strings.ReplaceAll(u, "-", "")}}
	case "b64":
		if rest != "" {
			rows = []kv{{"Base64", base64.StdEncoding.EncodeToString([]byte(rest))},
				{"Base64 URL-safe", base64.URLEncoding.EncodeToString([]byte(rest))}}
		}
	case "b64d":
		if s, ok := decodeBase64(rest); ok {
			rows = []kv{{"Decoded", s}}
		}
	case "url":
		if rest != "" {
			rows = []kv{{"Query-escaped", url.QueryEscape(rest)}, {"Path-escaped", url.PathEscape(rest)}}
		}
	case "urld":
		if s, err := url.QueryUnescape(rest); err == nil && rest != "" {
			rows = []kv{{"Decoded", s}}
		}
	case "hash":
		if rest != "" {
			rows = hashAll(rest)
		}
	case "case":
		if rest != "" {
			rows = caseAll(rest)
		}
	case "lorem":
		n := 30
		if v, err := strconv.Atoi(rest); err == nil && v > 0 && v <= 2000 {
			n = v
		}
		rows = []kv{{fmt.Sprintf("%d words", n), loremWords(n)}}
	case "ts":
		rows = tsResults(rest, time.Now(), time.Local)
		cat = "Time"
	case "now":
		rows = tsResults("", time.Now(), time.Local)
		cat = "Time"
	case "json":
		src := rest
		if src == "" {
			src, _ = clipText()
		}
		rows = jsonResults(src)
	case "color":
		rows = colorResults(rest)
		cat = "Color"
	case "unit":
		if v, f, to, ok := parseUnitQuery(rest); ok {
			rows = unitResults(v, f, to)
		}
		cat = "Unit"
	}
	if len(rows) == 0 {
		return []core.Item{hint("text", kw, kw+": "+textHelp[strings.ToLower(kw)], "")}
	}
	return kvItems(kw, rows, cat)
}

var textHelp = map[string]string{
	"uuid": "", "b64": "type text to encode", "b64d": "paste base64 to decode", "url": "type text to escape",
	"urld": "paste escaped text", "hash": "type text to hash", "case": "type text to convert",
	"lorem": "optional word count", "ts": "unix time or date (2024-01-31 15:04)", "now": "",
	"json": "copy JSON to the clipboard first", "color": "#rrggbb, rgb(r,g,b) or hsl(h,s%,l%)",
	"unit": "e.g. 10 km to mi",
}

func kvItems(kw string, rows []kv, cat string) []core.Item {
	out := make([]core.Item, 0, len(rows))
	for i, r := range rows {
		// IDs are stable per row label so frecency works, content never enters the ID.
		out = append(out, core.Item{
			ID: fmt.Sprintf("text:%s:%d", strings.ToLower(kw), i), Title: truncate(strings.ReplaceAll(r.value, "\n", " "), 120),
			Subtitle: r.label, Category: cat, Action: copyAction(r.value),
		})
	}
	return out
}

func newUUID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func decodeBase64(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return "", false
	}
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.URLEncoding, base64.RawStdEncoding, base64.RawURLEncoding} {
		if b, err := enc.DecodeString(s); err == nil {
			return string(b), true
		}
	}
	return "", false
}

func hashAll(s string) []kv {
	b := []byte(s)
	m5 := md5.Sum(b)
	s1 := sha1.Sum(b)
	s256 := sha256.Sum256(b)
	s512 := sha512.Sum512(b)
	return []kv{
		{"MD5", hex.EncodeToString(m5[:])},
		{"SHA-1", hex.EncodeToString(s1[:])},
		{"SHA-256", hex.EncodeToString(s256[:])},
		{"SHA-512", hex.EncodeToString(s512[:])},
	}
}

// ---- case conversion ----

// splitWords splits on separators and camelCase boundaries.
func splitWords(s string) []string {
	var words []string
	var cur []rune
	flush := func() {
		if len(cur) > 0 {
			words = append(words, string(cur))
			cur = cur[:0]
		}
	}
	rs := []rune(s)
	for i, r := range rs {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			flush()
			continue
		}
		if i > 0 && unicode.IsUpper(r) && len(cur) > 0 {
			prev := rs[i-1]
			nextLower := i+1 < len(rs) && unicode.IsLower(rs[i+1])
			if unicode.IsLower(prev) || unicode.IsDigit(prev) || (unicode.IsUpper(prev) && nextLower) {
				flush()
			}
		}
		cur = append(cur, r)
	}
	flush()
	return words
}

func capitalize(w string) string {
	r := []rune(strings.ToLower(w))
	if len(r) == 0 {
		return w
	}
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func caseAll(s string) []kv {
	w := splitWords(s)
	lower := make([]string, len(w))
	title := make([]string, len(w))
	for i, x := range w {
		lower[i] = strings.ToLower(x)
		title[i] = capitalize(x)
	}
	camel := ""
	if len(w) > 0 {
		camel = lower[0] + strings.Join(title[1:], "")
	}
	return []kv{
		{"UPPER", strings.ToUpper(s)},
		{"lower", strings.ToLower(s)},
		{"snake_case", strings.Join(lower, "_")},
		{"camelCase", camel},
		{"PascalCase", strings.Join(title, "")},
		{"kebab-case", strings.Join(lower, "-")},
		{"CONSTANT_CASE", strings.ToUpper(strings.Join(lower, "_"))},
		{"Title Case", strings.Join(title, " ")},
	}
}

// ---- lorem ----

var loremBase = strings.Fields(`lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt
ut labore et dolore magna aliqua ut enim ad minim veniam quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea
commodo consequat duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur
excepteur sint occaecat cupidatat non proident sunt in culpa qui officia deserunt mollit anim id est laborum`)

func loremWords(n int) string {
	var b strings.Builder
	for i := 0; i < n; i++ {
		w := loremBase[i%len(loremBase)]
		if i == 0 || strings.HasSuffix(b.String(), ". ") {
			w = capitalize(w)
		}
		b.WriteString(w)
		switch {
		case i == n-1:
			b.WriteByte('.')
		case i%12 == 11:
			b.WriteString(". ")
		default:
			b.WriteByte(' ')
		}
	}
	return b.String()
}

// ---- timestamps ----

var tsLayouts = []string{
	time.RFC3339, "2006-01-02T15:04:05", "2006-01-02 15:04:05", "2006-01-02 15:04", "2006-01-02", "2006/01/02", "02 Jan 2006", "Jan 2 2006",
}

// tsResults converts a unix timestamp (s/ms/us/ns by magnitude) or a date into
// formatted rows. An empty arg means "now".
func tsResults(arg string, now time.Time, loc *time.Location) []kv {
	arg = strings.TrimSpace(arg)
	var t time.Time
	switch {
	case arg == "" || strings.EqualFold(arg, "now"):
		t = now
	default:
		if n, err := strconv.ParseInt(arg, 10, 64); err == nil {
			a := n
			if a < 0 {
				a = -a
			}
			switch {
			case a >= 1e17:
				t = time.Unix(0, n)
			case a >= 1e14:
				t = time.UnixMicro(n)
			case a >= 1e11:
				t = time.UnixMilli(n)
			default:
				t = time.Unix(n, 0)
			}
		} else if f, err := strconv.ParseFloat(arg, 64); err == nil {
			t = time.Unix(int64(f), int64((f-math.Floor(f))*1e9))
		} else {
			found := false
			for _, l := range tsLayouts {
				if p, err := time.ParseInLocation(l, arg, loc); err == nil {
					t, found = p, true
					break
				}
			}
			if !found {
				return nil
			}
		}
	}
	rows := []kv{
		{"Unix seconds", strconv.FormatInt(t.Unix(), 10)},
		{"Unix milliseconds", strconv.FormatInt(t.UnixMilli(), 10)},
		{"UTC (RFC 3339)", t.UTC().Format(time.RFC3339)},
		{"Local (RFC 3339)", t.In(loc).Format(time.RFC3339)},
		{"Readable", t.In(loc).Format("Mon, 02 Jan 2006 15:04:05 MST")},
	}
	if arg != "" && !strings.EqualFold(arg, "now") {
		rows = append(rows, kv{"Relative", relativeTime(t, now)})
	}
	return rows
}

func relativeTime(t, now time.Time) string {
	d := now.Sub(t)
	suffix := " ago"
	if d < 0 {
		d, suffix = -d, " from now"
	}
	switch {
	case d < time.Minute:
		return fmt.Sprintf("%d seconds%s", int(d.Seconds()), suffix)
	case d < time.Hour:
		return fmt.Sprintf("%d minutes%s", int(d.Minutes()), suffix)
	case d < 48*time.Hour:
		return fmt.Sprintf("%.1f hours%s", d.Hours(), suffix)
	case d < 365*24*time.Hour:
		return fmt.Sprintf("%d days%s", int(d.Hours()/24), suffix)
	}
	return fmt.Sprintf("%.1f years%s", d.Hours()/24/365.25, suffix)
}

// ---- JSON ----

func jsonResults(src string) []kv {
	src = strings.TrimSpace(src)
	if src == "" || !json.Valid([]byte(src)) {
		return nil
	}
	var pretty, compact bytes.Buffer
	if json.Indent(&pretty, []byte(src), "", "  ") != nil || json.Compact(&compact, []byte(src)) != nil {
		return nil
	}
	return []kv{{"Pretty", pretty.String()}, {"Minified", compact.String()}}
}

// ---- color ----

func parseColor(s string) (r, g, b uint8, ok bool) {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return
	}
	if strings.HasPrefix(s, "rgb") || strings.HasPrefix(s, "hsl") {
		open, cl := strings.IndexByte(s, '('), strings.IndexByte(s, ')')
		if open < 0 || cl < open {
			return
		}
		parts := strings.FieldsFunc(s[open+1:cl], func(r rune) bool { return r == ',' || r == ' ' || r == '/' })
		if len(parts) < 3 {
			return
		}
		var v [3]float64
		for i := 0; i < 3; i++ {
			p := strings.TrimRight(parts[i], "%°deg")
			f, err := strconv.ParseFloat(p, 64)
			if err != nil {
				return
			}
			v[i] = f
		}
		if strings.HasPrefix(s, "rgb") {
			return clamp8(v[0]), clamp8(v[1]), clamp8(v[2]), true
		}
		r, g, b = hslToRGB(v[0], v[1]/100, v[2]/100)
		return r, g, b, true
	}
	s = strings.TrimPrefix(s, "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return
	}
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return
	}
	return uint8(n >> 16), uint8(n >> 8), uint8(n), true
}

func clamp8(f float64) uint8 {
	return uint8(math.Max(0, math.Min(255, math.Round(f))))
}

func rgbToHSL(r, g, b uint8) (h, s, l float64) {
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

func hslToRGB(h, s, l float64) (uint8, uint8, uint8) {
	h = math.Mod(math.Mod(h, 360)+360, 360) / 360
	if s == 0 {
		v := clamp8(l * 255)
		return v, v, v
	}
	var q float64
	if l < 0.5 {
		q = l * (1 + s)
	} else {
		q = l + s - l*s
	}
	p := 2*l - q
	f := func(t float64) float64 {
		if t < 0 {
			t++
		}
		if t > 1 {
			t--
		}
		switch {
		case t < 1.0/6:
			return p + (q-p)*6*t
		case t < 0.5:
			return q
		case t < 2.0/3:
			return p + (q-p)*(2.0/3-t)*6
		}
		return p
	}
	return clamp8(f(h+1.0/3) * 255), clamp8(f(h) * 255), clamp8(f(h-1.0/3) * 255)
}

func colorResults(s string) []kv {
	r, g, b, ok := parseColor(s)
	if !ok {
		return nil
	}
	h, sat, l := rgbToHSL(r, g, b)
	return []kv{
		{"Hex", fmt.Sprintf("#%02x%02x%02x", r, g, b)},
		{"HEX uppercase", fmt.Sprintf("#%02X%02X%02X", r, g, b)},
		{"RGB", fmt.Sprintf("rgb(%d, %d, %d)", r, g, b)},
		{"HSL", fmt.Sprintf("hsl(%d, %d%%, %d%%)", int(math.Round(h)), int(math.Round(sat*100)), int(math.Round(l*100)))},
	}
}

// ---- units ----

func unitResults(v float64, from, to string) []kv {
	res, err := convertUnits(v, from, to)
	if err != nil {
		return []kv{{"Error", err.Error()}}
	}
	out := formatNumber(res)
	return []kv{
		{fmt.Sprintf("%s %s", formatNumber(v), from), out + " " + to},
		{"Value only", out},
	}
}

package extract

import (
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

var (
	numRe      = regexp.MustCompile(`(?:\d{1,3}(?:,\d{3})+|\d+)(?:\.\d+)?(?:[eE][+-]?\d+)?|\.\d+(?:[eE][+-]?\d+)?`)
	percentRe  = regexp.MustCompile(`\d+(?:[.,]\d+)?[ \x{a0}]?%`)
	hexColorRe = regexp.MustCompile(`#(?:[0-9a-fA-F]{8}|[0-9a-fA-F]{6}|[0-9a-fA-F]{3,4})`)
	rgbRe      = regexp.MustCompile(`(?i)\brgba?\(\s*\d{1,3}%?\s*[, ]\s*\d{1,3}%?\s*[, ]\s*\d{1,3}%?\s*(?:[,/]\s*[\d.]+%?\s*)?\)`)
	hashtagRe  = regexp.MustCompile(`#([\p{L}\p{N}_]+)`)
	mentionRe  = regexp.MustCompile(`@([A-Za-z0-9_]{1,30})`)
	wordRe     = regexp.MustCompile(`[\p{L}\p{M}]+(?:['\x{2019}\-][\p{L}\p{M}]+)*`)
	semverRe   = regexp.MustCompile(`v?(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)\.(?:0|[1-9]\d*)(?:-[0-9A-Za-z\-]+(?:\.[0-9A-Za-z\-]+)*)?(?:\+[0-9A-Za-z\-]+(?:\.[0-9A-Za-z\-]+)*)?`)
	gitHashRe  = regexp.MustCompile(`[0-9a-f]{7,40}`)

	month  = `(?:Jan(?:uary)?|Feb(?:ruary)?|Mar(?:ch)?|Apr(?:il)?|May|June?|July?|Aug(?:ust)?|Sep(?:t(?:ember)?)?|Oct(?:ober)?|Nov(?:ember)?|Dec(?:ember)?)\.?`
	dateRe = regexp.MustCompile(`(?i)\b\d{4}[-/.]\d{1,2}[-/.]\d{1,2}\b|\b\d{1,2}[-/.]\d{1,2}[-/.]\d{4}\b|\b` + month + `[ \t]+\d{1,2}(?:st|nd|rd|th)?,?[ \t]+\d{4}\b|\b\d{1,2}(?:st|nd|rd|th)?[ \t]+(?:of[ \t]+)?` + month + `,?[ \t]+\d{4}\b`)
	timeRe = regexp.MustCompile(`(?i)\b(?:[01]?\d|2[0-3]):[0-5]\d(?::[0-5]\d(?:[.,]\d+)?)?(?:[ \t]?[ap]\.?m\.?)?|\b(?:1[0-2]|0?[1-9])[ \t]?[ap]m\b`)
	dtRe   = regexp.MustCompile(`\b\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}(?::\d{2}(?:[.,]\d+)?)?(?:Z|[+-]\d{2}(?::?\d{2})?)?|\b\d{8}T\d{6}(?:Z|[+-]\d{4})?\b`)

	winPathRe  = regexp.MustCompile(`(?i)\b[a-z]:[\\/](?:[^\s\\/:*?"'<>|]+[\\/])*[^\s\\/:*?"'<>|]*`)
	winQuoted  = regexp.MustCompile(`(?i)["']([a-z]:[\\/][^"'\r\n]*)["']`)
	uncPathRe  = regexp.MustCompile(`\\\\[^\s\\/:*?"'<>|]+(?:\\[^\s\\/:*?"'<>|]+)+`)
	unixPathRe = regexp.MustCompile(`(?:^|[\s"'(=\[,])((?:~|\.{1,2})?/(?:[\w.@%+\-~]+/)*[\w.@%+\-~]+/?)`)
	fileNameRe = regexp.MustCompile(`[A-Za-z0-9_\-][\w\-.()\[\]]*\.[A-Za-z][A-Za-z0-9]{0,9}`)
	urlOrMail  = regexp.MustCompile("(?i)(?:\\b(?:[a-z][a-z0-9+.\\-]*://|www\\.)[^\\s<>\"`]+)|[\\p{L}\\p{N}._%+\\-]+@[\\p{L}\\p{N}.\\-]+")
	dquoteRe   = regexp.MustCompile(`"((?:[^"\\\n]|\\.)*)"`)
	smartRe    = regexp.MustCompile("\u201c([^\u201d]*)\u201d|\u2018([^\u2018\u2019]*)\u2019|\u00ab([^\u00bb]*)\u00bb|\u201e([^\u201c\u201d]*)[\u201c\u201d]")
	btickRe    = regexp.MustCompile("`([^`\n]+)`")
	paraSplit  = regexp.MustCompile(`\r?\n[ \t]*\r?\n`)
)

// ---- numbers

// numberHits scans numbers; kind is "all", "int" or "dec".
func numberHits(text string, kind string, thousands, normalize bool) []hit {
	var out []hit
	ms := numRe.FindAllStringIndex(text, -1)
	skipUntil := 0
	for _, m := range ms {
		s, e := m[0], m[1]
		if s < skipUntil {
			continue
		}
		tok := text[s:e]
		p := prevRune(text, s)
		if isWordRune(p) || p == '.' && s >= 2 && (text[s-2] >= '0' && text[s-2] <= '9') {
			continue
		}
		if tok[0] == '.' && (p == '.' || isAlnum(p)) {
			continue
		}
		// dotted version / IP like 1.2.3 or 10.0.0.1: skip the whole token
		if n := nextRune(text, e); n == '.' && e+1 < len(text) && text[e+1] >= '0' && text[e+1] <= '9' {
			j := e
			for j+1 < len(text) && text[j] == '.' && text[j+1] >= '0' && text[j+1] <= '9' {
				j++
				for j < len(text) && text[j] >= '0' && text[j] <= '9' {
					j++
				}
			}
			skipUntil = j
			continue
		}
		// trailing letters: allow short unit suffixes (5km) but not hex-like runs (3fa8c1)
		if nr := nextRune(text, e); nr == '_' {
			continue
		} else if unicode.IsLetter(nr) {
			j := e
			for j < len(text) && (text[j] < utf8.RuneSelf && (text[j]|0x20 >= 'a' && text[j]|0x20 <= 'z')) {
				j++
			}
			if j-e > 3 || (j < len(text) && (isAlnum(nextRune(text, j)) || text[j] == '_')) {
				continue
			}
			if tok == "0" && (text[e] == 'x' || text[e] == 'X' || text[e] == 'b') {
				continue
			}
		}
		if !thousands && strings.Contains(tok, ",") {
			// treat "1,234" as separate numbers 1 and 234
			for _, part := range strings.Split(tok, ",") {
				i := strings.Index(text[s:e], part)
				out = appendNumber(out, kind, part, s+i, s+i+len(part), p, false)
			}
			continue
		}
		out = appendNumber(out, kind, tok, s, e, p, normalize)
	}
	return out
}

func appendNumber(out []hit, kind, tok string, s, e int, prev rune, normalize bool) []hit {
	isDec := strings.ContainsAny(tok, ".eE")
	if kind == "int" && isDec || kind == "dec" && !isDec {
		return out
	}
	v := tok
	if normalize {
		v = strings.ReplaceAll(v, ",", "")
	}
	return append(out, hit{s, e, v})
}

// withSigns extends number hits with a leading + or - when it is a sign.
func withSigns(text string, hs []hit) []hit {
	for i, h := range hs {
		if h.s == 0 {
			continue
		}
		c := text[h.s-1]
		if c != '-' && c != '+' {
			continue
		}
		if h.s >= 2 {
			pp, _ := utf8.DecodeLastRuneInString(text[:h.s-1])
			if isAlnum(pp) || pp == '.' || pp == ')' || pp == ']' || pp == '_' {
				continue
			}
		}
		hs[i].s--
		hs[i].v = string(c) + h.v
	}
	return hs
}

func numbersFinder(kind string) finder {
	return func(t string, a tools.Args) []hit {
		return withSigns(t, numberHits(t, kind, a.Bool("thousands"), a.Bool("normalize")))
	}
}

func percentHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range percentRe.FindAllStringIndex(text, -1) {
		if p := prevRune(text, m[0]); isWordRune(p) || p == '.' {
			continue
		}
		s := m[0]
		if s > 0 && (text[s-1] == '-' || text[s-1] == '+') {
			if s < 2 || !isAlnum(prevRune(text, s-1)) {
				s--
			}
		}
		out = append(out, hit{s, m[1], text[s:m[1]]})
	}
	return out
}

const curSym = `[$€£¥₹₽₩₺₪₫฿₴₦]`
const curCode = `(?:USD|EUR|GBP|JPY|CNY|INR|CAD|AUD|NZD|CHF|RUB|IRR|TRY|BRL|MXN|KRW|SEK|NOK|DKK|PLN|ZAR|AED|SAR|HKD|SGD|THB|BTC|ETH)`
const curAmount = `\d{1,3}(?:[,.\x{a0}\x{202f}]\d{3})+(?:[.,]\d{1,2})?|\d+(?:[.,]\d{1,2})?`

var currencyRe = regexp.MustCompile(`(?i)(?:(?:US|A|C|NZ|HK|R)?` + curSym + `|\b` + curCode + `\b)[ \x{a0}]?(?:` + curAmount + `)(?:[ \x{a0}]?(?:k|m|bn|b|million|billion|thousand)\b)?|(?:` + curAmount + `)[ \x{a0}]?(?:` + curSym + `|\b` + curCode + `\b|\b(?:dollars?|euros?|pounds?|yen|rupees?|rials?|tomans?|bucks)\b)`)

func currencyHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range currencyRe.FindAllStringIndex(text, -1) {
		if p := prevRune(text, m[0]); isWordRune(p) && !strings.ContainsRune("$€£¥₹₽₩₺₪₫฿₴₦", rune(text[m[0]])) {
			continue
		}
		if p := prevRune(text, m[0]); isAlnum(p) {
			continue
		}
		out = append(out, hit{m[0], m[1], strings.TrimSpace(text[m[0]:m[1]])})
	}
	return out
}

// ---- colors, tags, mentions

func hexColorHits(text string, a tools.Args) []hit {
	var out []hit
	for _, m := range hexColorRe.FindAllStringIndex(text, -1) {
		if p := prevRune(text, m[0]); isWordRune(p) || p == '&' {
			continue
		}
		if n := nextRune(text, m[1]); isWordRune(n) {
			continue
		}
		v := text[m[0]:m[1]]
		if a.Bool("lower") {
			v = strings.ToLower(v)
		}
		out = append(out, hit{m[0], m[1], v})
	}
	return out
}

func rgbHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range rgbRe.FindAllStringIndex(text, -1) {
		out = append(out, hit{m[0], m[1], text[m[0]:m[1]]})
	}
	return out
}

func prefixedHits(re *regexp.Regexp, pre string, text string, a tools.Args, badPrev string) []hit {
	var out []hit
	for _, m := range re.FindAllStringSubmatchIndex(text, -1) {
		if p := prevRune(text, m[0]); isWordRune(p) || strings.ContainsRune(badPrev, p) {
			continue
		}
		name := text[m[2]:m[3]]
		if pre == "#" {
			if _, err := strconv.Atoi(name); err == nil {
				continue
			}
		}
		v := pre + name
		if a.Bool("lower") {
			v = strings.ToLower(v)
		}
		out = append(out, hit{m[0], m[1], v})
	}
	return out
}

// ---- dates and times

func monthNum(s string) int {
	s = strings.ToLower(strings.TrimSuffix(s, "."))
	for i, m := range []string{"jan", "feb", "mar", "apr", "may", "jun", "jul", "aug", "sep", "oct", "nov", "dec"} {
		if strings.HasPrefix(s, m) {
			return i + 1
		}
	}
	return 0
}

func validYMD(y, m, d int) bool {
	if m < 1 || m > 12 || d < 1 || y < 1 {
		return false
	}
	t := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
	return t.Day() == d && int(t.Month()) == m
}

var nonDigit = regexp.MustCompile(`\D+`)

func dateOK(s string) bool {
	if len(s) > 0 && unicode.IsLetter(rune(s[0])) || strings.ContainsAny(s, "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ") {
		// textual month
		mm := regexp.MustCompile(`(?i)` + month).FindString(s)
		if mm == "" {
			return false
		}
		nums := regexp.MustCompile(`\d+`).FindAllString(s, -1)
		if len(nums) < 2 {
			return false
		}
		var d, y int
		for _, n := range nums {
			v, _ := strconv.Atoi(n)
			if len(n) == 4 {
				y = v
			} else if d == 0 {
				d = v
			}
		}
		return validYMD(y, monthNum(mm), d)
	}
	parts := nonDigit.Split(s, -1)
	if len(parts) != 3 {
		return false
	}
	seps := nonDigit.FindAllString(s, -1)
	if len(seps) != 2 || seps[0] != seps[1] {
		return false
	}
	a, _ := strconv.Atoi(parts[0])
	b, _ := strconv.Atoi(parts[1])
	c, _ := strconv.Atoi(parts[2])
	if len(parts[0]) == 4 {
		return validYMD(a, b, c)
	}
	return validYMD(c, b, a) || validYMD(c, a, b)
}

func dateHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range dateRe.FindAllStringIndex(text, -1) {
		v := text[m[0]:m[1]]
		if p := prevRune(text, m[0]); p == '.' || p == '/' || p == '-' || p == ':' {
			if isAlnum(prevRune(text, m[0]-1)) {
				continue
			}
		}
		if n := nextRune(text, m[1]); (n == '.' || n == '/' || n == '-') && m[1]+1 < len(text) && text[m[1]+1] >= '0' && text[m[1]+1] <= '9' {
			continue
		}
		if dateOK(v) {
			out = append(out, hit{m[0], m[1], v})
		}
	}
	return out
}

func timeHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range timeRe.FindAllStringIndex(text, -1) {
		p, n := prevRune(text, m[0]), nextRune(text, m[1])
		if unicode.IsDigit(p) || p == ':' || p == '.' && m[0] >= 2 && unicode.IsDigit(prevRune(text, m[0]-1)) {
			continue
		}
		if n == ':' && m[1]+1 < len(text) && text[m[1]+1] >= '0' && text[m[1]+1] <= '9' || unicode.IsDigit(n) || unicode.IsLetter(n) {
			continue
		}
		v := text[m[0]:m[1]]
		low := strings.ToLower(v)
		if strings.Contains(low, "m") && !strings.Contains(v, ":") {
			// "3pm" style: fine
		}
		out = append(out, hit{m[0], m[1], v})
	}
	return out
}

func datetimeHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range dtRe.FindAllStringIndex(text, -1) {
		v := text[m[0]:m[1]]
		if p := prevRune(text, m[0]); isAlnum(p) {
			continue
		}
		if len(v) >= 10 && v[4] == '-' {
			y, _ := strconv.Atoi(v[0:4])
			mo, _ := strconv.Atoi(v[5:7])
			d, _ := strconv.Atoi(v[8:10])
			if !validYMD(y, mo, d) {
				continue
			}
			hh, _ := strconv.Atoi(v[11:13])
			mi, _ := strconv.Atoi(v[14:16])
			if hh > 23 || mi > 59 {
				continue
			}
		}
		out = append(out, hit{m[0], m[1], v})
	}
	return out
}

// ---- paths and files

var knownRoots = map[string]bool{"usr": true, "etc": true, "var": true, "home": true, "tmp": true, "opt": true, "bin": true, "dev": true,
	"proc": true, "sys": true, "mnt": true, "root": true, "srv": true, "lib": true, "lib64": true, "sbin": true, "run": true, "boot": true,
	"Users": true, "Library": true, "Applications": true, "Volumes": true, "System": true, "private": true, "media": true}

func pathHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range winQuoted.FindAllStringSubmatchIndex(text, -1) {
		out = append(out, hit{m[2], m[3], text[m[2]:m[3]]})
	}
	for _, m := range winPathRe.FindAllStringIndex(text, -1) {
		if p := prevRune(text, m[0]); isAlnum(p) || p == '"' || p == '\'' { // quoted ones are handled above
			continue
		}
		v := trimTrailing(text[m[0]:m[1]])
		if len(v) <= 3 {
			continue
		}
		out = append(out, hit{m[0], m[0] + len(v), v})
	}
	for _, m := range uncPathRe.FindAllStringIndex(text, -1) {
		v := trimTrailing(text[m[0]:m[1]])
		out = append(out, hit{m[0], m[0] + len(v), v})
	}
	for _, m := range unixPathRe.FindAllStringSubmatchIndex(text, -1) {
		s, e := m[2], m[3]
		v := trimTrailing(text[s:e])
		if v == "" || v == "/" {
			continue
		}
		parts := strings.Split(strings.TrimLeft(strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(v, "~"), "../"), "./"), "/"), "/")
		abs := strings.HasPrefix(v, "/")
		if abs && len(parts) < 2 && !knownRoots[parts[0]] {
			continue
		}
		if !abs && len(parts) < 1 {
			continue
		}
		// skip things like "and/or" handled by the required leading slash; skip URLs' // remnants
		if strings.HasPrefix(v, "//") {
			continue
		}
		out = append(out, hit{s, s + len(v), v})
	}
	return out
}

var abbrevNames = map[string]bool{"e.g": true, "i.e": true, "a.m": true, "p.m": true, "u.s": true, "u.k": true, "etc.": true, "vs.": true, "et.al": true}

func fileNameHits(text string, _ tools.Args) []hit {
	// blank out URLs and email addresses so their host names are not file names
	masked := []byte(text)
	for _, m := range urlOrMail.FindAllStringIndex(text, -1) {
		for i := m[0]; i < m[1]; i++ {
			masked[i] = ' '
		}
	}
	mt := string(masked)
	var out []hit
	for _, m := range fileNameRe.FindAllStringIndex(mt, -1) {
		v := mt[m[0]:m[1]]
		if abbrevNames[strings.ToLower(v)] {
			continue
		}
		if n := nextRune(mt, m[1]); n == '@' || n == '.' && m[1]+1 < len(mt) && isAlnum(rune(mt[m[1]+1])) {
			continue
		}
		if p := prevRune(mt, m[0]); isAlnum(p) || p == '.' {
			continue
		}
		ext := v[strings.LastIndex(v, ".")+1:]
		if _, err := strconv.Atoi(ext); err == nil {
			continue
		}
		if len(v) == len(ext)+1 {
			continue
		}
		out = append(out, hit{m[0], m[1], v})
	}
	return out
}

func fileExtHits(text string, a tools.Args) []hit {
	var out []hit
	for _, h := range fileNameHits(text, a) {
		ext := strings.ToLower(h.v[strings.LastIndex(h.v, "."):])
		out = append(out, hit{h.s, h.e, ext})
	}
	return out
}

// ---- quotes, words, sentences

func quotedHits(text string, a tools.Args) []hit {
	var out []hit
	inc := a.Bool("include_quotes")
	emit := func(s, e, is, ie int) {
		if ie <= is {
			return
		}
		if inc {
			out = append(out, hit{s, e, text[s:e]})
		} else {
			out = append(out, hit{s, e, text[is:ie]})
		}
	}
	for _, m := range dquoteRe.FindAllStringSubmatchIndex(text, -1) {
		emit(m[0], m[1], m[2], m[3])
	}
	for _, m := range btickRe.FindAllStringSubmatchIndex(text, -1) {
		emit(m[0], m[1], m[2], m[3])
	}
	for _, m := range smartRe.FindAllStringSubmatchIndex(text, -1) {
		for g := 1; g <= 4; g++ {
			if m[2*g] >= 0 {
				emit(m[0], m[1], m[2*g], m[2*g+1])
			}
		}
	}
	// single quotes: opening quote not after a letter, closing quote not before a letter
	i := 0
	for i < len(text) {
		if text[i] != '\'' || isAlnum(prevRune(text, i)) {
			i++
			continue
		}
		j := i + 1
		found := -1
		for j < len(text) && text[j] != '\n' {
			if text[j] == '\'' && !isAlnum(nextRune(text, j+1)) {
				found = j
				break
			}
			j++
		}
		if found > i+1 {
			emit(i, found+1, i+1, found)
			i = found + 1
			continue
		}
		i++
	}
	return out
}

func wordHits(text string, a tools.Args) []hit {
	var out []hit
	for _, m := range wordRe.FindAllStringIndex(text, -1) {
		v := text[m[0]:m[1]]
		if a.Bool("lower") {
			v = strings.ToLower(v)
		}
		out = append(out, hit{m[0], m[1], v})
	}
	return out
}

var sentenceAbbrev = map[string]bool{"mr": true, "mrs": true, "ms": true, "dr": true, "prof": true, "sr": true, "jr": true, "st": true, "vs": true,
	"etc": true, "e.g": true, "i.e": true, "fig": true, "no": true, "inc": true, "ltd": true, "co": true, "approx": true, "dept": true, "est": true,
	"vol": true, "cf": true, "al": true, "mt": true, "gen": true, "col": true, "capt": true, "lt": true, "sgt": true, "rev": true, "hon": true,
	"u.s": true, "u.k": true, "a.m": true, "p.m": true, "ph.d": true, "jan": true, "feb": true, "mar": true, "apr": true, "aug": true,
	"sep": true, "sept": true, "oct": true, "nov": true, "dec": true}

func sentenceHits(text string, _ tools.Args) []hit {
	var out []hit
	rs := text
	start := 0
	flush := func(end int) {
		seg := strings.TrimSpace(rs[start:end])
		if seg != "" {
			// position of the trimmed segment
			s := start + (len(rs[start:end]) - len(strings.TrimLeft(rs[start:end], " \t\r\n")))
			out = append(out, hit{s, s + len(seg), strings.Join(strings.Fields(seg), " ")})
		}
		start = end
	}
	i := 0
	for i < len(rs) {
		c := rs[i]
		if c == '\n' && i+1 < len(rs) {
			// blank line ends a sentence
			j := i + 1
			for j < len(rs) && (rs[j] == ' ' || rs[j] == '\t' || rs[j] == '\r') {
				j++
			}
			if j < len(rs) && rs[j] == '\n' {
				flush(i)
				i = j
				continue
			}
		}
		if c == '.' || c == '!' || c == '?' {
			j := i + 1
			for j < len(rs) && (rs[j] == '.' || rs[j] == '!' || rs[j] == '?') {
				j++
			}
			for j < len(rs) && strings.ContainsRune(`"')]}`+"\u201d\u2019", rune(rs[j])) {
				j++
			}
			if j >= len(rs) {
				flush(len(rs))
				break
			}
			r, _ := utf8.DecodeRuneInString(rs[j:])
			if unicode.IsSpace(r) {
				ok := true
				if c == '.' {
					k := i
					for k > start && (isWordRune(prevRune(rs, k)) || prevRune(rs, k) == '.') {
						_, sz := utf8.DecodeLastRuneInString(rs[:k])
						k -= sz
					}
					w := strings.ToLower(rs[k:i])
					if sentenceAbbrev[w] || utf8.RuneCountInString(w) == 1 && unicode.IsLetter([]rune(w)[0]) {
						ok = false
					}
					// next word starts with a lower-case letter: not a sentence end
					t := strings.TrimLeft(rs[j:], " \t\r\n")
					if nr, _ := utf8.DecodeRuneInString(t); ok && unicode.IsLower(nr) && c == '.' {
						ok = false
					}
				}
				if ok {
					flush(j)
				}
			}
			i = j
			continue
		}
		i++
	}
	flush(len(rs))
	return out
}

func paragraphHits(text string, _ tools.Args) []hit {
	var out []hit
	last := 0
	emit := func(s, e int) {
		seg := text[s:e]
		t := strings.TrimSpace(seg)
		if t == "" {
			return
		}
		off := s + (len(seg) - len(strings.TrimLeft(seg, " \t\r\n")))
		out = append(out, hit{off, off + len(t), t})
	}
	for _, m := range paraSplit.FindAllStringIndex(text, -1) {
		emit(last, m[0])
		last = m[1]
	}
	emit(last, len(text))
	return out
}

func versionHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range semverRe.FindAllStringIndex(text, -1) {
		s, e := m[0], m[1]
		if p := prevRune(text, s); isAlnum(p) || p == '.' && s >= 2 && text[s-2] >= '0' && text[s-2] <= '9' {
			continue
		}
		// a fourth numeric component means an IP or build number, not semver
		if n := nextRune(text, e); n == '.' && e+1 < len(text) && text[e+1] >= '0' && text[e+1] <= '9' {
			continue
		}
		if n := nextRune(text, e); isAlnum(n) || n == '_' {
			continue
		}
		out = append(out, hit{s, e, text[s:e]})
	}
	return out
}

func gitHashHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range gitHashRe.FindAllStringIndex(text, -1) {
		v := text[m[0]:m[1]]
		if isWordRune(prevRune(text, m[0])) || isWordRune(nextRune(text, m[1])) {
			continue
		}
		hasDigit, hasLetter := false, false
		for _, c := range v {
			if c >= '0' && c <= '9' {
				hasDigit = true
			} else {
				hasLetter = true
			}
		}
		if hasDigit && hasLetter {
			out = append(out, hit{m[0], m[1], v})
		}
	}
	return out
}

func nonASCIIHits(text string, _ tools.Args) []hit {
	var out []hit
	for i, r := range text {
		if r > 127 && r != utf8.RuneError {
			out = append(out, hit{i, i + utf8.RuneLen(r), string(r)})
		}
	}
	return out
}

func capitalizedHits(text string, a tools.Args) []hit {
	type tok struct {
		s, e  int
		w     string
		start bool // sentence-initial
	}
	var toks []tok
	atStart := true
	last := 0
	for _, m := range wordRe.FindAllStringIndex(text, -1) {
		gap := text[last:m[0]]
		if strings.ContainsAny(gap, ".!?\n") {
			atStart = true
		}
		toks = append(toks, tok{m[0], m[1], text[m[0]:m[1]], atStart})
		atStart = false
		last = m[1]
	}
	isCap := func(w string) bool {
		r, n := utf8.DecodeRuneInString(w)
		if !unicode.IsUpper(r) || utf8.RuneCountInString(w) < 2 {
			return false
		}
		rest := w[n:]
		for _, c := range rest {
			if unicode.IsUpper(c) {
				return false
			}
		}
		return true
	}
	skip := a.Bool("skip_sentence_start")
	phrases := a.Bool("phrases")
	var out []hit
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		if !isCap(t.w) || skip && t.start {
			continue
		}
		s, e := t.s, t.e
		if phrases {
			for i+1 < len(toks) && isCap(toks[i+1].w) && strings.TrimSpace(text[e:toks[i+1].s]) == "" && !strings.ContainsAny(text[e:toks[i+1].s], "\n") && !toks[i+1].start {
				i++
				e = toks[i].e
			}
		}
		out = append(out, hit{s, e, text[s:e]})
	}
	return out
}

func acronymHits(text string, a tools.Args) []hit {
	minLen := int(a.Int("min_length"))
	if minLen < 2 {
		minLen = 2
	}
	var out []hit
	i := 0
	for i < len(text) {
		r, n := utf8.DecodeRuneInString(text[i:])
		if !isWordRune(r) {
			i += n
			continue
		}
		j := i
		for j < len(text) {
			r2, n2 := utf8.DecodeRuneInString(text[j:])
			if !isWordRune(r2) {
				break
			}
			j += n2
		}
		w := text[i:j]
		letters, ok := 0, true
		for _, c := range w {
			if unicode.IsUpper(c) {
				letters++
			} else if !(unicode.IsDigit(c) && letters >= 2) {
				ok = false
				break
			}
		}
		if ok && letters >= minLen && !allDigits(w) {
			out = append(out, hit{i, j, w})
		}
		i = j
	}
	return out
}

func allDigits(s string) bool {
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return s != ""
}

func textSpecs() []spec {
	numExtra := []tools.Param{opt("thousands", "bool", `Treat comma groups as thousands separators ("1,234.5" is one number); false splits at commas.`, true), opt("normalize", "bool", "Strip thousands separators from the results.", false)}
	return []spec{
		{name: "numbers", title: "Extract numbers", order: 6, keywords: []string{"numbers"}, extra: numExtra,
			doc:     "Finds integers and decimals with optional sign, thousands separators (1,234.56) and scientific notation (1.5e-3). Digits inside words (abc123), dotted versions and IPs (1.2.3, 10.0.0.1) and hex-like runs are skipped; short unit suffixes such as 5km are accepted. A minus between digits (2026-10-17) is an operator, not a sign.",
			example: `extract.numbers("Paid 1,234.50 for 3 items (-2.5e3 off)")`,
			tags:    []string{"digits", "figures", "integers", "decimals", "floats", "quantities", "scientific notation"}, find: numbersFinder("all")},
		{name: "integers", title: "Extract integers", order: 40, extra: numExtra, noAll: true,
			doc:     "Like extract.numbers but only whole numbers (no decimal point or exponent), with optional sign and thousands separators.",
			example: `extract.integers("1,200 apples, 3.5 pears, -7 plums")`,
			tags:    []string{"ints", "whole numbers", "digits", "counts"}, find: numbersFinder("int")},
		{name: "decimals", title: "Extract decimal numbers", order: 41, extra: numExtra, noAll: true,
			doc:     "Like extract.numbers but only numbers with a fractional part or exponent (3.14, -0.5, 1e6, .75).",
			example: `extract.decimals("pi 3.14159, e 2.718, 42")`,
			tags:    []string{"floats", "fractions", "real numbers", "doubles"}, find: numbersFinder("dec")},
		{name: "percentages", title: "Extract percentages", order: 42,
			doc:     "Finds percentages such as 45%, 12.5 %, -3% or 0,5%, returned as written.",
			example: `extract.percentages("up 12.5% and down 3%")`,
			tags:    []string{"percent", "rates", "ratio", "%"}, find: percentHits},
		{name: "currency_amounts", title: "Extract currency amounts", order: 43,
			doc:     "Finds money amounts: symbol first ($12.50, €1.299,00, £5, US$10), code first (USD 100, EUR 5.5) or after the number (12,50 €, 100 USD, 5 dollars). Heuristic: it does not interpret locale-specific separators, it returns the text as written.",
			example: `extract.currency_amounts("Total $1,299.00 or 12,50 € (USD 100)")`,
			tags:    []string{"money", "price", "prices", "dollars", "euros", "cost", "amount", "invoice"}, find: currencyHits},
		{name: "hex_colors", title: "Extract hex colors", order: 44, extra: []tools.Param{pLower},
			doc:     "Finds CSS hex colors #rgb, #rgba, #rrggbb and #rrggbbaa. HTML entities like &#123; and words glued to the hash are ignored; URL fragments such as #add may still match.",
			example: `extract.hex_colors("color: #FF8800; bg: #fff")`,
			tags:    []string{"color", "css colors", "hex color codes", "rgb hex", "palette"}, find: hexColorHits},
		{name: "rgb_colors", title: "Extract rgb() colors", order: 45,
			doc:     "Finds CSS rgb(...) and rgba(...) color functions with comma or space separated components (rgb(255, 0, 0), rgba(0 0 0 / 50%)).",
			example: `extract.rgb_colors("color: rgb(255, 128, 0); bg: rgba(0,0,0,0.5)")`,
			tags:    []string{"color", "css colors", "rgba", "palette"}, find: rgbHits},
		{name: "hashtags", title: "Extract hashtags", order: 46, keywords: []string{"hashtags"}, extra: []tools.Param{pLower},
			doc:     "Finds #hashtags (Unicode letters, digits, underscore; purely numeric tags like #1 and HTML entities are skipped). Results include the # sign.",
			example: `extract.hashtags("Loving #Go and #open_source! #2026")`,
			tags:    []string{"tags", "social media", "twitter", "instagram", "topics"},
			find:    func(t string, a tools.Args) []hit { return prefixedHits(hashtagRe, "#", t, a, "&#") }},
		{name: "mentions", title: "Extract @mentions", order: 47, extra: []tools.Param{pLower},
			doc:     "Finds @username mentions (letters, digits, underscore, up to 30 characters). The @ in email addresses is not a mention. Results include the @ sign.",
			example: `extract.mentions("thanks @alice and @Bob_42, mail me at x@y.com")`,
			tags:    []string{"handles", "usernames", "twitter", "social media", "at sign"},
			find:    func(t string, a tools.Args) []hit { return prefixedHits(mentionRe, "@", t, a, "@./") }},
		{name: "dates", title: "Extract dates", order: 50,
			doc:     "Finds calendar dates: ISO 2026-10-17 (also with / or .), day/month/year and month/day/year (31/12/2026, 12.31.2026), \"Jan 2, 2006\" and \"2 January 2006\" (month names and ordinal suffixes allowed). Impossible dates such as 2026-02-30 are rejected; numeric dd/mm vs mm/dd ambiguity is accepted when either reading is valid. Results are returned as written.",
			example: `extract.dates("Due 2026-10-17, also 12/25/2026 and March 3rd, 2027")`,
			tags:    []string{"date", "calendar", "iso 8601", "timestamps", "when"}, find: dateHits},
		{name: "times", title: "Extract times", order: 51,
			doc:     "Finds clock times hh:mm, hh:mm:ss(.fff) with optional am/pm (9:30, 14:05:59, 9:30 PM) and bare hours like 3pm. Segments of longer colon sequences (MAC or IPv6 addresses) are ignored.",
			example: `extract.times("Meet at 9:30 am or 14:45, call by 5pm")`,
			tags:    []string{"time", "clock", "hours", "am pm", "timestamps"}, find: timeHits},
		{name: "datetimes", title: "Extract ISO 8601 datetimes", order: 52,
			doc:     "Finds ISO 8601 date-times such as 2026-10-17T14:30:00Z, 2026-10-17 14:30+02:00 and the basic form 20261017T143000Z. Dates and times are range-checked.",
			example: `extract.datetimes("created 2026-10-17T14:30:00Z updated 2026-10-18 09:15")`,
			tags:    []string{"timestamp", "iso 8601", "rfc 3339", "utc", "log timestamps"}, find: datetimeHits},
		{name: "file_paths", title: "Extract file paths", order: 53,
			doc:     "Finds file system paths: Windows drive paths (C:\\Users\\me\\a.txt, also quoted ones with spaces), UNC paths (\\\\server\\share\\x) and Unix paths (/usr/local/bin, ~/docs/a.txt, ./src/main.go). Heuristic: unquoted paths end at whitespace, single-component absolute paths need a well-known root (/etc, /tmp...), and URLs are not reported.",
			example: `extract.file_paths("open C:\\Users\\me\\a.txt and /var/log/syslog or ~/notes/todo.md")`,
			tags:    []string{"paths", "directories", "folders", "unc", "windows path", "unix path", "filesystem"}, find: pathHits},
		{name: "file_names", title: "Extract file names", order: 54,
			doc:     "Finds file names with an extension (report.pdf, archive.tar.gz), including the last component of paths. URLs and email addresses are excluded first. Heuristic: abbreviations like e.g and version numbers are skipped, but a bare domain such as example.com in prose looks like a file name.",
			example: `extract.file_names("see report.pdf and C:\\tmp\\data.csv, not 1.2.3")`,
			tags:    []string{"filenames", "files", "attachments", "extension", "documents"}, find: fileNameHits},
		{name: "file_extensions", title: "Extract file extensions", order: 55,
			doc:     "Returns the extension (lower-case, with leading dot) of every file name found by extract.file_names, e.g. .pdf, .gz.",
			example: `extract.file_extensions("a.PDF b.tar.gz c.txt")`,
			tags:    []string{"suffix", "file types", "format"}, find: fileExtHits, noAll: true},
		{name: "quoted_strings", title: "Extract quoted strings", order: 56, extra: []tools.Param{opt("include_quotes", "bool", "Keep the quote characters in the results.", false)},
			doc:     "Returns text inside double quotes (with backslash escapes), single quotes (apostrophes inside words are not treated as quotes), backticks, curly quotes (\u201c\u201d \u2018\u2019), guillemets and low-9 quotes. Empty strings are skipped.",
			example: `extract.quoted_strings("He said \"hello there\" and 'bye'")`,
			tags:    []string{"quotes", "string literals", "double quotes", "single quotes", "backticks", "citations"}, find: quotedHits},
		{name: "words", title: "Extract words", order: 57, extra: []tools.Param{pLower}, noAll: true,
			doc:     "Splits text into words made of Unicode letters, keeping inner apostrophes and hyphens (don't, well-known). Digits and punctuation separate words. Remember unique defaults to true; pass unique=False to keep every occurrence.",
			example: `extract.words("Hello, wörld! It's well-known.", unique=False)`,
			tags:    []string{"tokens", "tokenize", "vocabulary", "word list", "unicode"}, find: wordHits},
		{name: "sentences", title: "Extract sentences", order: 58, noAll: true,
			doc:     "Splits text into sentences at . ! ? followed by whitespace, treating common abbreviations (Mr., Dr., e.g., U.S.), initials and a following lower-case word as non-boundaries; blank lines also end a sentence. Whitespace inside a sentence is collapsed.",
			example: `extract.sentences("Dr. Smith arrived. He sat down! Did he stay? Yes.")`,
			tags:    []string{"split sentences", "tokenize", "nlp", "segmentation"}, find: sentenceHits},
		{name: "paragraphs", title: "Extract paragraphs", order: 59, noAll: true,
			doc:     "Splits text into paragraphs separated by one or more blank lines; each paragraph is trimmed and keeps its inner line breaks. Remember unique defaults to true.",
			example: `extract.paragraphs("First para\nstill first\n\nSecond para")`,
			tags:    []string{"blocks", "split text", "sections", "chunks"}, find: paragraphHits},
		{name: "versions", title: "Extract version numbers", order: 60,
			doc:     "Finds semantic versions like 1.2.3, v2.0.1, 1.0.0-rc.1 and 1.0.0+build.5. Four-part numbers (IP addresses) and decimals are not reported.",
			example: `extract.versions("upgrade from v1.4.2 to 2.0.0-beta.1")`,
			tags:    []string{"semver", "release", "version", "software version", "changelog"}, find: versionHits},
		{name: "git_commit_hashes", title: "Extract git commit hashes", order: 61,
			doc:     "Finds abbreviated or full git commit hashes: 7 to 40 lower-case hex characters that contain both a digit and a letter (so plain numbers and words like \"facade\" are not reported). Overlaps with extract.hashes for 40-character values.",
			example: `extract.git_commit_hashes("fixed in 388841a and da693da0c1")`,
			tags:    []string{"sha", "commit", "git", "revision", "changeset", "short hash"}, find: gitHashHits},
		{name: "non_ascii_chars", title: "Extract non-ASCII characters", order: 62, noAll: true,
			doc:     "Lists every character outside 7-bit ASCII (accented letters, symbols, emoji, smart quotes, non-breaking spaces), one per result. Useful to find stray Unicode in source files.",
			example: `extract.non_ascii_chars("caf\u00e9 \u2013 na\u00efve \u201cquote\u201d")`,
			tags:    []string{"unicode", "special characters", "accents", "encoding problems", "smart quotes", "bom"}, find: nonASCIIHits},
		{name: "capitalized_words", title: "Extract capitalized words", order: 63, noAll: true,
			extra:   []tools.Param{opt("skip_sentence_start", "bool", "Ignore capitalized words at the start of a sentence.", true), opt("phrases", "bool", "Join consecutive capitalized words (New York City) into one result.", false)},
			doc:     "Proper-noun candidates: words starting with an upper-case letter followed only by lower-case letters (Paris, Alice). By default words that begin a sentence are skipped; phrases joins neighbours like \"New York\". Heuristic, no dictionary.",
			example: `extract.capitalized_words("Yesterday Alice flew to New York with Bob.", phrases=True)`,
			tags:    []string{"proper nouns", "names", "entities", "people", "places", "title case"}, find: capitalizedHits},
		{name: "acronyms", title: "Extract acronyms", order: 64, extra: []tools.Param{opt("min_length", "int", "Minimum number of capital letters (default 2).", int64(2))},
			doc:     "Finds words made only of capital letters (optionally followed by digits), at least min_length letters long: NASA, HTTP, SQL, USB3. Single letters and plain numbers are ignored.",
			example: `extract.acronyms("The NASA and ESA teams use HTTP and SQL")`,
			tags:    []string{"abbreviations", "initialisms", "caps", "all caps", "uppercase words"}, find: acronymHits},
	}
}

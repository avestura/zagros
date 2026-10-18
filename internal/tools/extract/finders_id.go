package extract

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

var (
	uuidRe = regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	hexRe  = regexp.MustCompile(`[0-9a-fA-F]{32,128}`)
	b64Re  = regexp.MustCompile(`[A-Za-z0-9+/]{16,}={0,2}`)
	jwtRe  = regexp.MustCompile(`eyJ[A-Za-z0-9_\-]{5,}\.[A-Za-z0-9_\-]{5,}\.[A-Za-z0-9_\-]*`)
	ccRe   = regexp.MustCompile(`\d(?:[ \-]?\d){12,18}`)
	ssnRe  = regexp.MustCompile(`\d{3}-\d{2}-\d{4}`)
	ibanRe = regexp.MustCompile(`[A-Z]{2}\d{2}`)
	btcRe  = regexp.MustCompile(`[13][a-km-zA-HJ-NP-Z1-9]{25,34}`)
	bechRe = regexp.MustCompile(`bc1[ac-hj-np-z02-9]{11,71}`)
	ethRe  = regexp.MustCompile(`0x[0-9a-fA-F]{40}`)
	envRe  = regexp.MustCompile(`%[A-Za-z_][A-Za-z0-9_]*%|\$\{[A-Za-z_][A-Za-z0-9_]*(?::?[-=+?][^}]*)?\}|\$env:[A-Za-z_][A-Za-z0-9_]*|\$[A-Za-z_][A-Za-z0-9_]*`)
	phRe   = regexp.MustCompile(`\{\{[^{}\n]+?\}\}|\{[A-Za-z_][\w.\-]*\}|<[A-Za-z_][\w.\- ]*>|\[\[[^\[\]\n]+\]\]`)
	emoji  = regexp.MustCompile(`[\x{1F1E6}-\x{1F1FF}]{2}|[0-9#*]\x{FE0F}?\x{20E3}|[\x{1F300}-\x{1FAFF}\x{1F000}-\x{1F2FF}\x{2600}-\x{27BF}\x{2B50}\x{2B55}\x{231A}\x{231B}\x{23E9}-\x{23F3}\x{23F8}-\x{23FA}\x{25B6}\x{25C0}\x{2934}\x{2935}\x{2B05}-\x{2B07}\x{2B1B}\x{2B1C}\x{203C}\x{2049}\x{2194}-\x{2199}\x{21A9}\x{21AA}](?:\x{FE0F})?[\x{1F3FB}-\x{1F3FF}]?(?:\x{200D}[\x{1F300}-\x{1FAFF}\x{2600}-\x{27BF}\x{2640}\x{2642}](?:\x{FE0F})?[\x{1F3FB}-\x{1F3FF}]?)*`)
)

func uuidHits(text string, a tools.Args) []hit {
	var out []hit
	for _, m := range uuidRe.FindAllStringIndex(text, -1) {
		if isAlnum(prevRune(text, m[0])) || isAlnum(nextRune(text, m[1])) || prevRune(text, m[0]) == '-' || nextRune(text, m[1]) == '-' {
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

var hashNames = map[int]string{32: "md5", 40: "sha1", 56: "sha224", 64: "sha256", 96: "sha384", 128: "sha512"}

func hashHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range hexRe.FindAllStringIndex(text, -1) {
		v := text[m[0]:m[1]]
		name, ok := hashNames[len(v)]
		if !ok || isWordRune(prevRune(text, m[0])) || isWordRune(nextRune(text, m[1])) {
			continue
		}
		if p := prevRune(text, m[0]); p == '-' {
			continue
		}
		if n := nextRune(text, m[1]); n == '-' {
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
		if !hasDigit || !hasLetter {
			continue
		}
		out = append(out, hit{m[0], m[1], name + ":" + strings.ToLower(v)})
	}
	return out
}

func base64Hits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range b64Re.FindAllStringIndex(text, -1) {
		v := text[m[0]:m[1]]
		p, n := prevRune(text, m[0]), nextRune(text, m[1])
		if isWordRune(p) || isWordRune(n) || strings.ContainsRune(".-=/+", p) || strings.ContainsRune(".-=", n) {
			continue
		}
		if len(v)%4 != 0 {
			continue
		}
		var lower, upper, digit, sym bool
		for _, c := range strings.TrimRight(v, "=") {
			switch {
			case c >= 'a' && c <= 'z':
				lower = true
			case c >= 'A' && c <= 'Z':
				upper = true
			case c >= '0' && c <= '9':
				digit = true
			default:
				sym = true
			}
		}
		classes := 0
		for _, b := range []bool{lower, upper, digit, sym} {
			if b {
				classes++
			}
		}
		pad := strings.HasSuffix(v, "=")
		if classes < 3 && !(pad && classes >= 2) {
			continue
		}
		if _, err := base64.StdEncoding.DecodeString(v); err != nil {
			continue
		}
		out = append(out, hit{m[0], m[1], v})
	}
	return out
}

func jwtHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range jwtRe.FindAllStringIndex(text, -1) {
		if isWordRune(prevRune(text, m[0])) {
			continue
		}
		v := text[m[0]:m[1]]
		seg := v[:strings.IndexByte(v, '.')]
		if raw, err := base64.RawURLEncoding.DecodeString(seg); err != nil || len(raw) == 0 || raw[0] != '{' {
			continue
		}
		out = append(out, hit{m[0], m[1], v})
	}
	return out
}

func luhn(digits string) bool {
	sum, alt := 0, false
	for i := len(digits) - 1; i >= 0; i-- {
		d := int(digits[i] - '0')
		if alt {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		alt = !alt
	}
	return sum%10 == 0
}

func onlyDigits(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] >= '0' && s[i] <= '9' {
			b.WriteByte(s[i])
		}
	}
	return b.String()
}

func cardHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range ccRe.FindAllStringIndex(text, -1) {
		v := text[m[0]:m[1]]
		if isAlnum(prevRune(text, m[0])) || isAlnum(nextRune(text, m[1])) || prevRune(text, m[0]) == '-' || nextRune(text, m[1]) == '-' {
			continue
		}
		d := onlyDigits(v)
		if len(d) < 13 || len(d) > 19 || d[0] == '0' || d[0] == '1' || strings.Count(d, d[:1]) == len(d) {
			continue
		}
		// consistent separators only
		sp, da := strings.Contains(v, " "), strings.Contains(v, "-")
		if sp && da {
			continue
		}
		if luhn(d) {
			out = append(out, hit{m[0], m[1], v})
		}
	}
	return out
}

func ssnHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range ssnRe.FindAllStringIndex(text, -1) {
		if isAlnum(prevRune(text, m[0])) || isAlnum(nextRune(text, m[1])) || prevRune(text, m[0]) == '-' || nextRune(text, m[1]) == '-' {
			continue
		}
		v := text[m[0]:m[1]]
		area, group, serial := v[0:3], v[4:6], v[7:11]
		if area == "000" || area == "666" || area[0] == '9' || group == "00" || serial == "0000" {
			continue
		}
		out = append(out, hit{m[0], m[1], v})
	}
	return out
}

var ibanLen = map[string]int{"AL": 28, "AD": 24, "AT": 20, "AZ": 28, "BH": 22, "BY": 28, "BE": 16, "BA": 20, "BR": 29, "BG": 22, "CR": 22, "HR": 21,
	"CY": 28, "CZ": 24, "DK": 18, "DO": 28, "EG": 29, "SV": 28, "EE": 20, "FO": 18, "FI": 18, "FR": 27, "GE": 22, "DE": 22, "GI": 23, "GR": 27,
	"GL": 18, "GT": 28, "HU": 28, "IS": 26, "IQ": 23, "IE": 22, "IL": 23, "IT": 27, "JO": 30, "KZ": 20, "XK": 20, "KW": 30, "LV": 21, "LB": 28,
	"LY": 25, "LI": 21, "LT": 20, "LU": 20, "MT": 31, "MR": 27, "MU": 30, "MC": 27, "MD": 24, "ME": 22, "NL": 18, "MK": 19, "NO": 15, "PK": 24,
	"PS": 29, "PL": 28, "PT": 25, "QA": 29, "RO": 24, "LC": 32, "SM": 27, "ST": 25, "SA": 24, "RS": 22, "SC": 31, "SK": 24, "SI": 19, "ES": 24,
	"SD": 18, "SE": 24, "CH": 21, "TL": 23, "TN": 24, "TR": 26, "UA": 29, "AE": 23, "GB": 22, "VA": 22, "VG": 24, "IR": 26}

func ibanValid(s string) bool {
	if len(s) < 15 || len(s) > 34 {
		return false
	}
	r := s[4:] + s[:4]
	n := new(big.Int)
	var sb strings.Builder
	for _, c := range r {
		switch {
		case c >= '0' && c <= '9':
			sb.WriteRune(c)
		case c >= 'A' && c <= 'Z':
			sb.WriteString(itoa(int(c-'A') + 10))
		default:
			return false
		}
	}
	n.SetString(sb.String(), 10)
	return n.Mod(n, big.NewInt(97)).Int64() == 1
}

func itoa(n int) string { return string(rune('0'+n/10)) + string(rune('0'+n%10)) }

func ibanHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range ibanRe.FindAllStringIndex(text, -1) {
		s := m[0]
		if isWordRune(prevRune(text, s)) {
			continue
		}
		want, ok := ibanLen[text[s:s+2]]
		if !ok {
			continue
		}
		var b strings.Builder
		i := s
		for i < len(text) && b.Len() < want {
			c := text[i]
			switch {
			case c >= '0' && c <= '9' || c >= 'A' && c <= 'Z':
				b.WriteByte(c)
				i++
			case c == ' ' && i+1 < len(text) && (text[i+1] >= '0' && text[i+1] <= '9' || text[i+1] >= 'A' && text[i+1] <= 'Z'):
				i++
			default:
				i = len(text) + 1
			}
		}
		if i > len(text) || b.Len() != want || isWordRune(nextRune(text, i)) {
			continue
		}
		if ibanValid(b.String()) {
			out = append(out, hit{s, i, b.String()})
		}
	}
	return out
}

const b58 = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func base58Check(s string) bool {
	n := new(big.Int)
	base := big.NewInt(58)
	for _, c := range s {
		i := strings.IndexRune(b58, c)
		if i < 0 {
			return false
		}
		n.Mul(n, base)
		n.Add(n, big.NewInt(int64(i)))
	}
	raw := n.Bytes()
	for i := 0; i < len(s) && s[i] == '1'; i++ {
		raw = append([]byte{0}, raw...)
	}
	if len(raw) < 5 {
		return false
	}
	body, sum := raw[:len(raw)-4], raw[len(raw)-4:]
	h := sha256.Sum256(body)
	h = sha256.Sum256(h[:])
	return h[0] == sum[0] && h[1] == sum[1] && h[2] == sum[2] && h[3] == sum[3]
}

func cryptoHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range btcRe.FindAllStringIndex(text, -1) {
		if isWordRune(prevRune(text, m[0])) || isWordRune(nextRune(text, m[1])) {
			continue
		}
		if base58Check(text[m[0]:m[1]]) {
			out = append(out, hit{m[0], m[1], text[m[0]:m[1]]})
		}
	}
	for _, m := range bechRe.FindAllStringIndex(text, -1) {
		if isWordRune(prevRune(text, m[0])) || isWordRune(nextRune(text, m[1])) {
			continue
		}
		if l := m[1] - m[0]; l >= 14 && l <= 74 {
			out = append(out, hit{m[0], m[1], text[m[0]:m[1]]})
		}
	}
	for _, m := range ethRe.FindAllStringIndex(text, -1) {
		if isWordRune(prevRune(text, m[0])) || isWordRune(nextRune(text, m[1])) {
			continue
		}
		out = append(out, hit{m[0], m[1], text[m[0]:m[1]]})
	}
	return out
}

func envHits(text string, a tools.Args) []hit {
	var out []hit
	for _, m := range envRe.FindAllStringIndex(text, -1) {
		v := text[m[0]:m[1]]
		if strings.HasPrefix(v, "$") && isWordRune(prevRune(text, m[0])) {
			continue
		}
		if a.Bool("names_only") {
			switch {
			case strings.HasPrefix(v, "%"):
				v = strings.Trim(v, "%")
			case strings.HasPrefix(v, "${"):
				v = strings.TrimSuffix(strings.TrimPrefix(v, "${"), "}")
				if i := strings.IndexAny(v, ":-=+?"); i >= 0 {
					v = v[:i]
				}
			case strings.HasPrefix(v, "$env:"):
				v = v[5:]
			default:
				v = v[1:]
			}
		}
		out = append(out, hit{m[0], m[1], v})
	}
	return out
}

var htmlTagNames = map[string]bool{}

func init() {
	for _, n := range strings.Fields("a abbr address area article aside audio b base bdi bdo blockquote body br button canvas caption cite code col colgroup data datalist dd del details dfn dialog div dl dt em embed fieldset figcaption figure footer form h1 h2 h3 h4 h5 h6 head header hr html i iframe img input ins kbd label legend li link main map mark meta meter nav noscript object ol optgroup option output p param picture pre progress q rp rt ruby s samp script section select slot small source span strong style sub summary sup svg table tbody td template textarea tfoot th thead time title tr track u ul var video wbr center font") {
		htmlTagNames[n] = true
	}
}

func placeholderHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range phRe.FindAllStringIndex(text, -1) {
		v := text[m[0]:m[1]]
		if v[0] == '{' && v[1] != '{' && m[0] > 0 && text[m[0]-1] == '$' {
			continue // ${VAR} is an environment variable
		}
		if v[0] == '<' && htmlTagNames[strings.ToLower(strings.TrimSpace(v[1:len(v)-1]))] {
			continue
		}
		out = append(out, hit{m[0], m[1], v})
	}
	return out
}

func jsonHits(text string, _ tools.Args) []hit {
	var out []hit
	budget := 16*len(text) + 1<<20
	i := 0
	for i < len(text) && budget > 0 {
		c := text[i]
		if c != '{' && c != '[' {
			i++
			continue
		}
		end := balancedEnd(text, i, &budget)
		if end > 0 {
			sn := text[i:end]
			if len(sn) >= 3 && json.Valid([]byte(sn)) && !trivialJSON(sn) {
				out = append(out, hit{i, end, sn})
				i = end
				continue
			}
		}
		i++
	}
	return out
}

func trivialJSON(s string) bool {
	if s[0] != '[' {
		return false
	}
	inner := strings.TrimSpace(s[1 : len(s)-1])
	return !strings.ContainsAny(inner, ",{[\"") // like [1] or [true]: a footnote marker, not data
}

// balancedEnd returns the index just past the bracket matching text[start],
// honouring JSON strings, or -1.
func balancedEnd(text string, start int, budget *int) int {
	depth := 0
	inStr := false
	for i := start; i < len(text); i++ {
		*budget--
		if *budget < 0 {
			return -1
		}
		c := text[i]
		if inStr {
			switch c {
			case '\\':
				i++
			case '"':
				inStr = false
			case '\n':
				// JSON strings cannot contain raw newlines
				return -1
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{', '[':
			depth++
		case '}', ']':
			depth--
			if depth == 0 {
				return i + 1
			}
		}
	}
	return -1
}

func emojiHits(text string, _ tools.Args) []hit {
	var out []hit
	for _, m := range emoji.FindAllStringIndex(text, -1) {
		v := text[m[0]:m[1]]
		if r, _ := utf8.DecodeRuneInString(v); r < 0x80 && len(v) <= 1 {
			continue
		}
		out = append(out, hit{m[0], m[1], v})
	}
	return out
}

func idSpecs() []spec {
	return []spec{
		{name: "uuids", title: "Extract UUIDs", order: 70, extra: []tools.Param{pLower},
			doc:     "Finds UUIDs/GUIDs in the 8-4-4-4-12 hex form (any version, case-insensitive), including those wrapped in braces. Neighbouring hex or dash characters make a candidate invalid.",
			example: `extract.uuids("id=123e4567-e89b-12d3-a456-426614174000")`,
			tags:    []string{"guid", "identifiers", "ids", "uuid v4", "unique id"}, find: uuidHits},
		{name: "hashes", title: "Extract hashes", order: 71,
			doc:     "Finds hexadecimal digests by length and reports them as \"type:hash\" in lower case: md5 (32), sha1 (40), sha224 (56), sha256 (64), sha384 (96), sha512 (128). A token must mix digits and letters, and the type is a guess from the length only (a 32-hex UUID without dashes also looks like md5).",
			example: `extract.hashes("md5 d41d8cd98f00b204e9800998ecf8427e")`,
			tags:    []string{"md5", "sha1", "sha256", "sha512", "checksum", "digest", "fingerprint", "hex"}, find: hashHits},
		{name: "base64_strings", title: "Extract Base64 strings", order: 72,
			doc:     "Finds plausible standard Base64 strings: at least 16 characters, length divisible by 4, valid padding, and a mix of at least three character classes (or padding plus two), so ordinary words, paths and hex hashes are rejected. URL-safe Base64 (-_) and JWT segments are not reported. Heuristic.",
			example: `extract.base64_strings("token SGVsbG8sIFdvcmxkISBIZWxsbyE= end")`,
			tags:    []string{"b64", "encoded", "secrets", "tokens", "payload"}, find: base64Hits},
		{name: "jwt_tokens", title: "Extract JWT tokens", order: 73,
			doc:     "Finds JSON Web Tokens: three dot-separated Base64url segments whose header decodes to a JSON object (starts with eyJ). The signature may be empty (unsecured tokens).",
			example: `extract.jwt_tokens("Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.c2ln")`,
			tags:    []string{"bearer", "json web token", "auth token", "access token", "jws"}, find: jwtHits},
		{name: "credit_cards", title: "Extract credit card numbers", order: 74,
			doc:     "Finds 13-19 digit card numbers (optionally grouped with single spaces or dashes) that pass the Luhn check and do not start with 0 or 1 or repeat one digit. Returned as written. A Luhn match does not prove the number is a real card.",
			example: `extract.credit_cards("card 4111 1111 1111 1111 exp 12/29")`,
			tags:    []string{"card numbers", "pan", "visa", "mastercard", "luhn", "payment", "pci"}, find: cardHits},
		{name: "ssn", title: "Extract US SSNs", order: 75,
			doc:     "Finds US Social Security Numbers in the NNN-NN-NNNN form, rejecting impossible values (area 000, 666 or 9xx, group 00, serial 0000).",
			example: `extract.ssn("SSN 123-45-6789")`,
			tags:    []string{"social security number", "us ssn", "pii", "tax id"}, find: ssnHits},
		{name: "ibans", title: "Extract IBANs", order: 76,
			doc:     "Finds International Bank Account Numbers, with or without spaces, validating country, length and the mod-97 checksum. Results are compact upper-case (no spaces).",
			example: `extract.ibans("pay to GB82 WEST 1234 5698 7654 32 please")`,
			tags:    []string{"bank account", "iban", "swift", "sepa", "account number", "pii"}, find: ibanHits},
		{name: "crypto_addresses", title: "Extract crypto addresses", order: 77,
			doc:     "Finds Bitcoin legacy addresses (Base58Check verified), Bitcoin bech32 addresses (bc1..., shape check only) and Ethereum addresses (0x plus 40 hex digits, checksum casing not verified).",
			example: `extract.crypto_addresses("send to 1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2")`,
			tags:    []string{"bitcoin", "btc", "ethereum", "eth", "wallet", "blockchain", "address"}, find: cryptoHits},
		{name: "env_vars", title: "Extract environment variables", order: 78, extra: []tools.Param{opt("names_only", "bool", "Return just the variable names without %...%, $ or ${...}.", false)},
			doc:     "Finds environment variable references: %VAR% (Windows), $VAR, ${VAR} with optional :-default (POSIX) and $env:VAR (PowerShell). Percentages like \"50% of 20%\" are not matched because names cannot contain spaces.",
			example: `extract.env_vars("cd %USERPROFILE% && echo $HOME ${PATH:-/bin}", names_only=True)`,
			tags:    []string{"environment", "variables", "shell", "path", "windows", "bash", "powershell"}, find: envHits},
		{name: "placeholders", title: "Extract placeholders", order: 79,
			doc:     "Finds template placeholders: {{name}}, {name}, [[name]] and <name>. Single-brace placeholders must be identifier-like (so JSON objects are not matched), ${VAR} is treated as an environment variable instead, and known HTML tag names in angle brackets are skipped.",
			example: `extract.placeholders("Hello {{user}}, order {id} for <customer>")`,
			tags:    []string{"template", "variables", "mustache", "handlebars", "fill in", "merge fields"}, find: placeholderHits},
		{name: "json_snippets", title: "Extract JSON snippets", order: 80, noAll: false,
			doc:     "Finds JSON objects and arrays embedded in text by matching balanced braces/brackets (honouring strings and escapes) and keeping only snippets that parse as valid JSON. Trivial arrays like [1] and unparseable {braces} are skipped; nested snippets are returned as part of their outer one.",
			example: `extract.json_snippets("response: {\"a\": [1, 2]} done")`,
			tags:    []string{"json", "objects", "payload", "logs", "embedded json", "braces"}, find: jsonHits},
		{name: "emojis", title: "Extract emojis", order: 81,
			doc:     "Finds emoji, including skin-tone modifiers, ZWJ sequences (family, professions), flags and keycaps. Based on Unicode ranges, so a few rare symbols may be missed or over-matched.",
			example: `extract.emojis("Great job \U0001F44D\U0001F3FD \U0001F1EE\U0001F1F7 ❤️")`,
			tags:    []string{"emoji", "smileys", "unicode symbols", "emoticons", "flags", "pictographs"}, find: emojiHits},
	}
}

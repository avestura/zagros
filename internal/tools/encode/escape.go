package encode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf16"
	"unicode/utf8"
)

// bsUnescape decodes backslash escapes. mode is "c", "python" or "json".
// \x and octal escapes yield raw bytes; \u and \U yield UTF-8.
func bsUnescape(s, mode string) ([]byte, error) {
	var out []byte
	hexN := func(i, n int) (int, bool) {
		if i+n > len(s) {
			return 0, false
		}
		v, err := strconv.ParseUint(s[i:i+n], 16, 32)
		if err != nil {
			return 0, false
		}
		return int(v), true
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '\\' {
			out = append(out, c)
			continue
		}
		if i+1 >= len(s) {
			out = append(out, c)
			break
		}
		i++
		e := s[i]
		simple := map[byte]byte{'n': '\n', 't': '\t', 'r': '\r', 'b': '\b', 'f': '\f', '\\': '\\', '"': '"', '/': '/'}
		if mode != "json" {
			simple['a'] = 7
			simple['v'] = 11
			simple['\''] = '\''
			delete(simple, '/')
			if mode == "c" {
				simple['?'] = '?'
				simple['e'] = 27
			}
		}
		if v, ok := simple[e]; ok {
			out = append(out, v)
			continue
		}
		switch {
		case e == 'u':
			if mode != "json" && i+1 < len(s) && s[i+1] == '{' {
				if j := strings.IndexByte(s[i:], '}'); j > 0 {
					if v, err := strconv.ParseUint(s[i+2:i+j], 16, 32); err == nil && v <= 0x10FFFF {
						out = utf8.AppendRune(out, rune(v))
						i += j
						continue
					}
				}
			}
			v, ok := hexN(i+1, 4)
			if !ok {
				if mode == "json" {
					return nil, fmt.Errorf("invalid \\u escape at offset %d", i-1)
				}
				out = append(out, '\\', 'u')
				continue
			}
			i += 4
			r := rune(v)
			if utf16.IsSurrogate(r) && r < 0xDC00 && i+6 < len(s)+0 && s[i+1] == '\\' && s[i+2] == 'u' {
				if lo, ok := hexN(i+3, 4); ok && lo >= 0xDC00 && lo <= 0xDFFF {
					r = utf16.DecodeRune(r, rune(lo))
					i += 6
				}
			}
			if utf16.IsSurrogate(r) {
				r = utf8.RuneError
			}
			out = utf8.AppendRune(out, r)
		case e == 'U' && mode != "json":
			v, ok := hexN(i+1, 8)
			if !ok || v > 0x10FFFF {
				out = append(out, '\\', 'U')
				continue
			}
			i += 8
			if v >= 0xD800 && v <= 0xDFFF {
				v = 0xFFFD
			}
			out = utf8.AppendRune(out, rune(v))
		case e == 'x' && mode != "json":
			v, ok := hexN(i+1, 2)
			if !ok {
				out = append(out, '\\', 'x')
				continue
			}
			i += 2
			out = append(out, byte(v))
		case e >= '0' && e <= '7' && mode != "json":
			n := 1
			v := int(e - '0')
			for n < 3 && i+1 < len(s) && s[i+1] >= '0' && s[i+1] <= '7' {
				i++
				v = v*8 + int(s[i]-'0')
				n++
			}
			out = append(out, byte(v))
		case e == '\n' && mode == "python":
			// line continuation
		case e == '\r' && mode == "python":
			if i+1 < len(s) && s[i+1] == '\n' {
				i++
			}
		default:
			if mode == "json" {
				return nil, fmt.Errorf("invalid escape \\%c at offset %d", e, i-1)
			}
			out = append(out, '\\', e)
		}
	}
	return out, nil
}

func bsToText(s, mode string) (string, error) {
	b, err := bsUnescape(s, mode)
	if err != nil {
		return "", err
	}
	if !utf8.Valid(b) {
		return "", fmt.Errorf("the escapes decode to bytes that are not valid UTF-8; use c_unescape with output=\"hex\"")
	}
	return string(b), nil
}

func cEscape(s, control, nonASCII string, quote bool) string {
	var sb strings.Builder
	if quote {
		sb.WriteByte('"')
	}
	byteEsc := func(c byte, style string) {
		if style == "hex" {
			fmt.Fprintf(&sb, "\\x%02x", c)
		} else {
			fmt.Fprintf(&sb, "\\%03o", c)
		}
	}
	for i := 0; i < len(s); {
		r, n := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && n == 1 {
			byteEsc(s[i], "octal")
			i++
			continue
		}
		switch {
		case r == '\\':
			sb.WriteString(`\\`)
		case r == '"':
			sb.WriteString(`\"`)
		case r == '\n':
			sb.WriteString(`\n`)
		case r == '\t':
			sb.WriteString(`\t`)
		case r == '\r':
			sb.WriteString(`\r`)
		case r == 7:
			sb.WriteString(`\a`)
		case r == 8:
			sb.WriteString(`\b`)
		case r == 11:
			sb.WriteString(`\v`)
		case r == 12:
			sb.WriteString(`\f`)
		case r < 0x20 || r == 0x7f:
			byteEsc(byte(r), control)
		case r < 0x80:
			sb.WriteRune(r)
		default:
			switch nonASCII {
			case "octal", "hex":
				for _, c := range []byte(s[i : i+n]) {
					byteEsc(c, nonASCII)
				}
			case "unicode":
				if r > 0xFFFF {
					fmt.Fprintf(&sb, "\\U%08x", r)
				} else {
					fmt.Fprintf(&sb, "\\u%04x", r)
				}
			default:
				sb.WriteString(s[i : i+n])
			}
		}
		i += n
	}
	if quote {
		sb.WriteByte('"')
	}
	return sb.String()
}

func unicodeEscape(s string, all bool, style string, upper bool) string {
	var sb strings.Builder
	f4, f8 := "\\u%04x", "\\U%08x"
	if upper {
		f4, f8 = "\\u%04X", "\\U%08X"
	}
	for _, r := range s {
		if r < 0x80 && !all {
			sb.WriteRune(r)
			continue
		}
		switch {
		case r <= 0xFFFF:
			fmt.Fprintf(&sb, f4, r)
		case style == "python":
			fmt.Fprintf(&sb, f8, r)
		default:
			r1, r2 := utf16.EncodeRune(r)
			fmt.Fprintf(&sb, f4, r1)
			fmt.Fprintf(&sb, f4, r2)
		}
	}
	return sb.String()
}

func codepointEscape(s string, all, upper bool) string {
	var sb strings.Builder
	f := "\\u{%x}"
	if upper {
		f = "\\u{%X}"
	}
	for _, r := range s {
		if r < 0x80 && !all {
			sb.WriteRune(r)
			continue
		}
		fmt.Fprintf(&sb, f, r)
	}
	return sb.String()
}

// unicodeUnescape decodes \uXXXX (with surrogate pairs), \u{X..}, \UXXXXXXXX,
// \xHH and U+XXXX forms, leaving everything else untouched.
func unicodeUnescape(s string) string {
	var sb strings.Builder
	hexAt := func(i, n int) (int, bool) {
		if i+n > len(s) {
			return 0, false
		}
		for _, c := range []byte(s[i : i+n]) {
			if !isHexDigit(c) {
				return 0, false
			}
		}
		v, _ := strconv.ParseUint(s[i:i+n], 16, 32)
		return int(v), true
	}
	for i := 0; i < len(s); {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'u':
				if i+2 < len(s) && s[i+2] == '{' {
					if j := strings.IndexByte(s[i:], '}'); j > 3 && j <= 10 {
						if v, err := strconv.ParseUint(s[i+3:i+j], 16, 32); err == nil && v <= 0x10FFFF {
							sb.WriteRune(rune(v))
							i += j + 1
							continue
						}
					}
				}
				if v, ok := hexAt(i+2, 4); ok {
					r := rune(v)
					i += 6
					if r >= 0xD800 && r < 0xDC00 && i+1 < len(s) && s[i] == '\\' && s[i+1] == 'u' {
						if lo, ok := hexAt(i+2, 4); ok && lo >= 0xDC00 && lo <= 0xDFFF {
							r = utf16.DecodeRune(r, rune(lo))
							i += 6
						}
					}
					if utf16.IsSurrogate(r) {
						r = utf8.RuneError
					}
					sb.WriteRune(r)
					continue
				}
			case 'U':
				if v, ok := hexAt(i+2, 8); ok && v <= 0x10FFFF {
					if v >= 0xD800 && v <= 0xDFFF {
						v = 0xFFFD
					}
					sb.WriteRune(rune(v))
					i += 10
					continue
				}
			case 'x':
				if v, ok := hexAt(i+2, 2); ok {
					sb.WriteRune(rune(v))
					i += 4
					continue
				}
			}
		}
		if s[i] == 'U' && i+2 < len(s) && s[i+1] == '+' && (i == 0 || !isAlnum(s[i-1])) {
			j := i + 2
			for j < len(s) && j < i+8 && isHexDigit(s[j]) {
				j++
			}
			if j-(i+2) >= 4 {
				if v, err := strconv.ParseUint(s[i+2:j], 16, 32); err == nil && v <= 0x10FFFF && !(v >= 0xD800 && v <= 0xDFFF) {
					sb.WriteRune(rune(v))
					i = j
					continue
				}
			}
		}
		sb.WriteByte(s[i])
		i++
	}
	return sb.String()
}

var uniEscPattern = regexp.MustCompile(`\\u[0-9a-fA-F]{4}|\\u\{[0-9a-fA-F]{1,6}\}|\\U[0-9a-fA-F]{8}|\bU\+[0-9a-fA-F]{4,6}`)

const jsKeep = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789@*_+-./"

func jsEscape(s string) string {
	var sb strings.Builder
	for _, u := range utf16.Encode([]rune(s)) {
		switch {
		case u < 0x80 && strings.IndexByte(jsKeep, byte(u)) >= 0:
			sb.WriteByte(byte(u))
		case u < 256:
			fmt.Fprintf(&sb, "%%%02X", u)
		default:
			fmt.Fprintf(&sb, "%%u%04X", u)
		}
	}
	return sb.String()
}

func jsUnescape(s string) string {
	var units []uint16
	hexAt := func(i, n int) (int, bool) {
		if i+n > len(s) {
			return 0, false
		}
		for _, c := range []byte(s[i : i+n]) {
			if !isHexDigit(c) {
				return 0, false
			}
		}
		v, _ := strconv.ParseUint(s[i:i+n], 16, 32)
		return int(v), true
	}
	for i := 0; i < len(s); {
		if s[i] == '%' {
			if i+1 < len(s) && (s[i+1] == 'u' || s[i+1] == 'U') {
				if v, ok := hexAt(i+2, 4); ok {
					units = append(units, uint16(v))
					i += 6
					continue
				}
			}
			if v, ok := hexAt(i+1, 2); ok {
				units = append(units, uint16(v))
				i += 3
				continue
			}
		}
		r, n := utf8.DecodeRuneInString(s[i:])
		units = append(units, utf16.Encode([]rune{r})...)
		i += n
	}
	return string(utf16.Decode(units))
}

func jsonEscape(s string, asciiOnly bool) string {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(s)
	out := strings.TrimSuffix(buf.String(), "\n")
	if asciiOnly {
		var sb strings.Builder
		for _, r := range out {
			if r < 0x80 {
				sb.WriteRune(r)
			} else if r > 0xFFFF {
				r1, r2 := utf16.EncodeRune(r)
				fmt.Fprintf(&sb, "\\u%04x\\u%04x", r1, r2)
			} else {
				fmt.Fprintf(&sb, "\\u%04x", r)
			}
		}
		out = sb.String()
	}
	return out
}

func jsonUnescape(s string) (string, error) {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	return bsToText(s, "json")
}

func pythonRepr(s string) string {
	q := byte('\'')
	if strings.Contains(s, "'") && !strings.Contains(s, "\"") {
		q = '"'
	}
	var sb strings.Builder
	sb.WriteByte(q)
	for i := 0; i < len(s); {
		r, n := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError && n == 1 {
			fmt.Fprintf(&sb, "\\x%02x", s[i])
			i++
			continue
		}
		switch {
		case r == '\\':
			sb.WriteString(`\\`)
		case r == rune(q):
			sb.WriteByte('\\')
			sb.WriteByte(q)
		case r == '\n':
			sb.WriteString(`\n`)
		case r == '\r':
			sb.WriteString(`\r`)
		case r == '\t':
			sb.WriteString(`\t`)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&sb, "\\x%02x", r)
		case r < 0x7f:
			sb.WriteRune(r)
		case unicode.IsPrint(r):
			sb.WriteRune(r)
		case r <= 0xff:
			fmt.Fprintf(&sb, "\\x%02x", r)
		case r <= 0xffff:
			fmt.Fprintf(&sb, "\\u%04x", r)
		default:
			fmt.Fprintf(&sb, "\\U%08x", r)
		}
		i += n
	}
	sb.WriteByte(q)
	return sb.String()
}

func pythonUnrepr(s string) (string, error) {
	t := strings.TrimSpace(s)
	raw := false
	for len(t) > 0 && strings.ContainsRune("rRuUfF", rune(t[0])) && len(t) > 1 && (t[1] == '\'' || t[1] == '"' || strings.ContainsRune("rRuUfF", rune(t[1]))) {
		if t[0] == 'r' || t[0] == 'R' {
			raw = true
		}
		t = t[1:]
	}
	for _, q := range []string{`'''`, `"""`, `'`, `"`} {
		if len(t) >= 2*len(q) && strings.HasPrefix(t, q) && strings.HasSuffix(t, q) {
			t = t[len(q) : len(t)-len(q)]
			break
		}
	}
	if raw {
		return t, nil
	}
	return bsToText(t, "python")
}

func cssEscape(s string) string {
	var sb strings.Builder
	rs := []rune(s)
	for i, r := range rs {
		switch {
		case r == 0:
			sb.WriteRune('�')
		case r >= 1 && r <= 0x1f || r == 0x7f:
			fmt.Fprintf(&sb, "\\%x ", r)
		case i == 0 && r >= '0' && r <= '9', i == 1 && r >= '0' && r <= '9' && rs[0] == '-':
			fmt.Fprintf(&sb, "\\%x ", r)
		case i == 0 && r == '-' && len(rs) == 1:
			sb.WriteString(`\-`)
		case r >= 0x80 || r == '-' || r == '_' || r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z':
			sb.WriteRune(r)
		default:
			sb.WriteByte('\\')
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

func cssUnescape(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); {
		if s[i] != '\\' || i+1 >= len(s) {
			sb.WriteByte(s[i])
			i++
			continue
		}
		i++
		if isHexDigit(s[i]) {
			j := i
			for j < len(s) && j < i+6 && isHexDigit(s[j]) {
				j++
			}
			v, _ := strconv.ParseUint(s[i:j], 16, 32)
			r := rune(v)
			if v == 0 || v > 0x10FFFF || utf16.IsSurrogate(r) {
				r = utf8.RuneError
			}
			sb.WriteRune(r)
			i = j
			if i < len(s) && (s[i] == ' ' || s[i] == '\t' || s[i] == '\n') {
				i++
			} else if i+1 < len(s) && s[i] == '\r' && s[i+1] == '\n' {
				i += 2
			}
			continue
		}
		if s[i] == '\n' {
			i++
			continue
		}
		r, n := utf8.DecodeRuneInString(s[i:])
		sb.WriteRune(r)
		i += n
	}
	return sb.String()
}

func regexUnescape(s string) string {
	var sb strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) && s[i+1] < 0x80 && !isAlnum(s[i+1]) && s[i+1] > ' ' {
			i++
		}
		sb.WriteByte(s[i])
	}
	return sb.String()
}

func sqlEscape(s, dialect string) (string, error) {
	switch normName(dialect) {
	case "", "standard", "ansi", "postgres", "sqlite", "mssql":
		return strings.ReplaceAll(s, "'", "''"), nil
	case "mysql":
		return strings.NewReplacer("\\", "\\\\", "'", "\\'", "\"", "\\\"", "\x00", "\\0", "\n", "\\n", "\r", "\\r", "\x1a", "\\Z").Replace(s), nil
	}
	return "", fmt.Errorf("unknown dialect %q (use standard or mysql)", dialect)
}

func sqlUnescape(s, dialect string) (string, error) {
	switch normName(dialect) {
	case "", "standard", "ansi", "postgres", "sqlite", "mssql":
		return strings.ReplaceAll(s, "''", "'"), nil
	case "mysql":
		s = strings.ReplaceAll(s, "''", "'")
		var sb strings.Builder
		for i := 0; i < len(s); i++ {
			if s[i] == '\\' && i+1 < len(s) {
				i++
				switch s[i] {
				case '0':
					sb.WriteByte(0)
				case 'n':
					sb.WriteByte('\n')
				case 'r':
					sb.WriteByte('\r')
				case 't':
					sb.WriteByte('\t')
				case 'b':
					sb.WriteByte('\b')
				case 'Z':
					sb.WriteByte(0x1a)
				default:
					sb.WriteByte(s[i])
				}
				continue
			}
			sb.WriteByte(s[i])
		}
		return sb.String(), nil
	}
	return "", fmt.Errorf("unknown dialect %q (use standard or mysql)", dialect)
}

var shSafe = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./-]+$`)

func shQuote(s string, always bool) string {
	if s == "" {
		return "''"
	}
	if !always && shSafe.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func psQuote(s string) string {
	var sb strings.Builder
	sb.WriteByte('\'')
	for _, r := range s {
		switch r {
		case '\'', '‘', '’', '‚', '‛':
			sb.WriteRune(r)
		}
		sb.WriteRune(r)
	}
	sb.WriteByte('\'')
	return sb.String()
}

func cmdQuote(s string, meta bool) string {
	needs := s == "" || strings.ContainsAny(s, " \t\n\v\"")
	q := s
	if needs {
		var sb strings.Builder
		sb.WriteByte('"')
		bs := 0
		for _, r := range s {
			switch r {
			case '\\':
				bs++
				continue
			case '"':
				sb.WriteString(strings.Repeat(`\`, bs*2+1))
				sb.WriteByte('"')
			default:
				sb.WriteString(strings.Repeat(`\`, bs))
				sb.WriteRune(r)
			}
			bs = 0
		}
		sb.WriteString(strings.Repeat(`\`, bs*2))
		sb.WriteByte('"')
		q = sb.String()
	}
	if meta {
		var sb strings.Builder
		for _, r := range q {
			if strings.ContainsRune("()%!^\"&|<>", r) {
				sb.WriteByte('^')
			}
			sb.WriteRune(r)
		}
		q = sb.String()
	}
	return q
}

func csvEscape(s, delim string, always bool) string {
	if always || strings.Contains(s, delim) || strings.ContainsAny(s, "\"\r\n") {
		return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
	}
	return s
}

func csvUnescape(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return strings.ReplaceAll(s[1:len(s)-1], "\"\"", "\"")
	}
	return s
}

// ---- quoted-printable ----

func qpEncode(b []byte) string {
	var sb strings.Builder
	lineLen := 0
	emit := func(tok string, soft bool) {
		if soft && lineLen+len(tok) > 75 {
			sb.WriteString("=\r\n")
			lineLen = 0
		}
		sb.WriteString(tok)
		lineLen += len(tok)
	}
	for i := 0; i < len(b); i++ {
		c := b[i]
		switch {
		case c == '\r' && i+1 < len(b) && b[i+1] == '\n':
			sb.WriteString("\r\n")
			i++
			lineLen = 0
		case c == '\n':
			sb.WriteString("\r\n")
			lineLen = 0
		case (c == ' ' || c == '\t') && (i+1 == len(b) || b[i+1] == '\r' || b[i+1] == '\n'):
			emit(fmt.Sprintf("=%02X", c), true)
		case c >= 33 && c <= 126 && c != '=' || c == ' ' || c == '\t':
			emit(string(rune(c)), true)
		default:
			emit(fmt.Sprintf("=%02X", c), true)
		}
	}
	return sb.String()
}

func qpDecode(s string) []byte {
	var out []byte
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c != '=' {
			out = append(out, c)
			continue
		}
		switch {
		case i+2 < len(s) && s[i+1] == '\r' && s[i+2] == '\n':
			i += 2
		case i+1 < len(s) && s[i+1] == '\n':
			i++
		case i+2 < len(s)+0 && isHexDigit(s[i+1]) && isHexDigit(s[i+2]):
			out = append(out, unhex(s[i+1])<<4|unhex(s[i+2]))
			i += 2
		case i+1 == len(s):
			// trailing soft break marker
		default:
			out = append(out, c)
		}
	}
	return out
}

var qpPattern = regexp.MustCompile(`=[0-9A-Fa-f]{2}|=\r?\n`)

// ---- uuencode ----

func uuEncode(b []byte, name, mode string) string {
	var sb strings.Builder
	enc := func(v byte) byte {
		if v == 0 {
			return '`'
		}
		return v + 32
	}
	sb.WriteString("begin " + mode + " " + name + "\n")
	for i := 0; i < len(b); i += 45 {
		end := i + 45
		if end > len(b) {
			end = len(b)
		}
		line := b[i:end]
		sb.WriteByte(enc(byte(len(line))))
		for j := 0; j < len(line); j += 3 {
			var g [3]byte
			copy(g[:], line[j:])
			sb.WriteByte(enc(g[0] >> 2))
			sb.WriteByte(enc((g[0]&3)<<4 | g[1]>>4))
			sb.WriteByte(enc((g[1]&15)<<2 | g[2]>>6))
			sb.WriteByte(enc(g[2] & 63))
		}
		sb.WriteByte('\n')
	}
	sb.WriteString("`\nend\n")
	return sb.String()
}

func uuDecode(s string) ([]byte, error) {
	lines := strings.Split(strings.ReplaceAll(s, "\r\n", "\n"), "\n")
	var out []byte
	started := true
	for _, l := range lines {
		if strings.HasPrefix(l, "begin ") {
			started = true
			continue
		}
		if !started {
			continue
		}
		if l == "end" {
			break
		}
		if l == "" {
			continue
		}
		n := int((l[0] - 32) & 63)
		if n == 0 {
			continue
		}
		data := l[1:]
		var line []byte
		for j := 0; j < len(data); j += 4 {
			var g [4]byte
			for k := 0; k < 4; k++ {
				if j+k < len(data) {
					g[k] = (data[j+k] - 32) & 63
				}
			}
			line = append(line, g[0]<<2|g[1]>>4, g[1]<<4|g[2]>>2, g[2]<<6|g[3])
		}
		if len(line) < n {
			return nil, fmt.Errorf("uuencoded line is shorter than its length byte says")
		}
		out = append(out, line[:n]...)
	}
	return out, nil
}

func unescapeAuto(s string) string {
	t := s
	if strings.Contains(t, "%u") || strings.Contains(t, "%U") {
		t = jsUnescape(t)
	}
	if bsPattern.MatchString(t) {
		if b, err := bsUnescape(t, "c"); err == nil && utf8.Valid(b) {
			t = string(b)
		}
	}
	if entityPattern.MatchString(t) {
		t = html.UnescapeString(t)
	}
	if hasPercentSeq(t) {
		if b := percentDecode(t, false); utf8.Valid(b) {
			t = string(b)
		}
	}
	return t
}

var bsPattern = regexp.MustCompile(`\\(?:[nrtbfv0'"\\/]|x[0-9a-fA-F]{2}|u[0-9a-fA-F]{4}|u\{[0-9a-fA-F]+\}|U[0-9a-fA-F]{8}|[0-7]{3})`)

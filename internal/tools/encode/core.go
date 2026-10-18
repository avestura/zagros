// Package encode registers the "encode" group of the tools registry: Base-N
// codecs, text/number forms, charsets, URL, HTML/XML, punycode, escaping
// pairs, classic encodings and token helpers.
//
// Binary convention: binary data crosses the API as strings. Encoders take
// text (the UTF-8 bytes of the string) and an optional input_encoding
// ("utf-8", "hex", "base64", ...) so arbitrary bytes can be encoded.
// Decoders return text when the bytes are valid UTF-8, and otherwise fail
// unless the output parameter ("hex", "base64", ...) asks for a byte form.
package encode

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

const maxBytes = 8 << 20

var (
	accepts = map[string]func(string) bool{}
	counter int
)

func register(f tools.Func, launcher bool) {
	f.Group = "encode"
	if launcher && f.Family == "" {
		for _, p := range f.Params {
			if p.Type == "str" {
				f.Family = "encode"
				break
			}
		}
	}
	if !launcher {
		f.Family = ""
		f.Accepts = nil
	}
	if f.Order == 0 {
		counter++
		f.Order = 1000 + counter
	}
	if f.Accepts != nil {
		acc := f.Accepts
		accepts[f.Name] = acc
		f.Accepts = func(s string) (ok bool) {
			defer func() {
				if recover() != nil {
					ok = false
				}
			}()
			if len(s) > maxBytes {
				return false
			}
			return acc(s)
		}
	}
	inner := f.Fn
	name := f.Key()
	f.Fn = func(a tools.Args) (res any, err error) {
		defer func() {
			if r := recover(); r != nil {
				res, err = nil, fmt.Errorf("%s: internal error: %v", name, r)
			}
		}()
		for _, v := range a {
			if s, ok := v.(string); ok && len(s) > maxBytes {
				return nil, fmt.Errorf("%s: input is larger than 8 MiB", name)
			}
		}
		res, err = inner(a)
		if err != nil {
			return nil, err
		}
		if s, ok := res.(string); ok && len(s) > maxBytes {
			return nil, fmt.Errorf("%s: result would be larger than 8 MiB", name)
		}
		return res, nil
	}
	tools.Register(f)
}

// reg registers a launcher-visible function.
func reg(f tools.Func) { register(f, true) }

// regAPI registers an API-only function (no launcher family).
func regAPI(f tools.Func) { register(f, false) }

// ---- parameter constructors ----

func pReq(name, doc string) tools.Param {
	return tools.Param{Name: name, Type: "str", Doc: doc, Required: true}
}

func pStr(name, doc, def string, choices ...string) tools.Param {
	return tools.Param{Name: name, Type: "str", Doc: doc, Default: def, Choices: choices}
}

func pBool(name, doc string, def bool) tools.Param {
	return tools.Param{Name: name, Type: "bool", Doc: doc, Default: def}
}

func pInt(name, doc string, def int64) tools.Param {
	return tools.Param{Name: name, Type: "int", Doc: doc, Default: def}
}

func pDict(name, doc string) tools.Param {
	return tools.Param{Name: name, Type: "dict[str, str]", Doc: doc, Required: true}
}

var inEncChoices = []string{"utf-8", "hex", "base64", "percent", "decimal", "binary"}
var outChoices = []string{"text", "hex", "base64", "percent", "decimal", "binary"}

func pInEnc() tools.Param {
	return pStr("input_encoding", "How to read the text argument as bytes: \"utf-8\" (default, the text itself), or a byte form: \"hex\", \"base64\", \"percent\", \"decimal\" (space separated 0-255), \"binary\".", "utf-8", inEncChoices...)
}

func pOut() tools.Param {
	return pStr("output", "Form of the result: \"text\" (default, fails if the bytes are not valid UTF-8), or a byte form \"hex\", \"base64\", \"percent\", \"decimal\", \"binary\".", "text", outChoices...)
}

// ---- byte conversions ----

func normName(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}

// fromRep converts text in the given representation to bytes.
func fromRep(s, enc string) ([]byte, error) {
	switch normName(enc) {
	case "", "utf-8", "utf8", "text":
		return []byte(s), nil
	case "hex", "base16":
		return parseHex(s)
	case "base64", "b64":
		return parseBase64(s)
	case "percent", "url":
		return percentDecode(s, false), nil
	case "decimal", "dec":
		return parseNumberTokens(s, 10)
	case "binary", "bin":
		return parseBinary(s)
	case "octal", "oct":
		return parseNumberTokens(s, 8)
	}
	return nil, fmt.Errorf("unknown input encoding %q (use utf-8, hex, base64, percent, decimal or binary)", enc)
}

// toRep converts bytes to the requested output form.
func toRep(b []byte, out string) (string, error) {
	switch normName(out) {
	case "", "text", "utf-8", "utf8":
		if !utf8.Valid(b) {
			return "", fmt.Errorf("result is not valid UTF-8 text; pass output=\"hex\" or output=\"base64\" to get the bytes")
		}
		return string(b), nil
	case "hex", "base16":
		return hex.EncodeToString(b), nil
	case "base64", "b64":
		return base64.StdEncoding.EncodeToString(b), nil
	case "percent", "url":
		return percentEncodeAll(b), nil
	case "decimal", "dec":
		return joinNums(b, 10, " ", 0), nil
	case "binary", "bin":
		return joinNums(b, 2, " ", 8), nil
	case "octal", "oct":
		return joinNums(b, 8, " ", 3), nil
	}
	return "", fmt.Errorf("unknown output %q (use text, hex, base64, percent, decimal or binary)", out)
}

func joinNums(b []byte, base int, sep string, width int) string {
	parts := make([]string, len(b))
	for i, c := range b {
		s := strconv.FormatInt(int64(c), base)
		if len(s) < width {
			s = strings.Repeat("0", width-len(s)) + s
		}
		parts[i] = s
	}
	return strings.Join(parts, sep)
}

func percentEncodeAll(b []byte) string {
	const digits = "0123456789ABCDEF"
	var sb strings.Builder
	sb.Grow(len(b) * 3)
	for _, c := range b {
		sb.WriteByte('%')
		sb.WriteByte(digits[c>>4])
		sb.WriteByte(digits[c&15])
	}
	return sb.String()
}

func isHexDigit(c byte) bool {
	return c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F'
}

func unhex(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	}
	return c - 'A' + 10
}

// percentDecode decodes %XX sequences (invalid ones stay literal) and
// optionally turns + into a space.
func percentDecode(s string, plus bool) []byte {
	out := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == '%' && i+2 < len(s) && isHexDigit(s[i+1]) && isHexDigit(s[i+2]):
			out = append(out, unhex(s[i+1])<<4|unhex(s[i+2]))
			i += 2
		case c == '+' && plus:
			out = append(out, ' ')
		default:
			out = append(out, c)
		}
	}
	return out
}

// parseHex is tolerant: it ignores whitespace, ":", "-", ",", "_", "%", "0x",
// "\x" and "&#x...;" decorations.
func parseHex(s string) ([]byte, error) {
	r := strings.NewReplacer("0x", "", "0X", "", "\\x", "", "\\X", "", "&#x", "", "&#X", "")
	s = r.Replace(s)
	buf := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case isHexDigit(c):
			buf = append(buf, c)
		case c == ' ' || c == '\t' || c == '\r' || c == '\n' || c == ':' || c == '-' || c == ',' || c == '_' || c == '%' || c == ';' || c == '#':
		default:
			return nil, fmt.Errorf("invalid hex digit %q", string(rune(c)))
		}
	}
	if len(buf)%2 != 0 {
		return nil, fmt.Errorf("hex input has an odd number of digits")
	}
	out := make([]byte, len(buf)/2)
	for i := range out {
		out[i] = unhex(buf[2*i])<<4 | unhex(buf[2*i+1])
	}
	return out, nil
}

// parseBase64 accepts standard and URL-safe alphabets, with or without padding
// and with whitespace or a data: URI prefix.
func parseBase64(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "data:") {
		if i := strings.Index(s, ";base64,"); i >= 0 {
			s = s[i+8:]
		}
	}
	var sb strings.Builder
	isURL, isStd := false, false
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c == ' ' || c == '\t' || c == '\r' || c == '\n':
			continue
		case c == '=':
			continue
		case c == '-' || c == '_':
			isURL = true
		case c == '+' || c == '/':
			isStd = true
		case c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9':
		default:
			return nil, fmt.Errorf("invalid base64 character %q", string(rune(c)))
		}
		sb.WriteByte(c)
	}
	if isURL && isStd {
		return nil, fmt.Errorf("mixed standard and URL-safe base64 alphabets")
	}
	t := sb.String()
	if len(t)%4 == 1 {
		return nil, fmt.Errorf("invalid base64 length")
	}
	enc := base64.RawStdEncoding
	if isURL {
		enc = base64.RawURLEncoding
	}
	return enc.DecodeString(t)
}

func parseNumberTokens(s string, base int) ([]byte, error) {
	var out []byte
	isDigit := func(c rune) bool {
		if base == 8 {
			return c >= '0' && c <= '7'
		}
		return c >= '0' && c <= '9'
	}
	toks := strings.FieldsFunc(s, func(c rune) bool { return !isDigit(c) })
	if len(toks) == 1 && base == 8 && len(toks[0]) > 3 && len(toks[0])%3 == 0 {
		t := toks[0]
		toks = nil
		for i := 0; i < len(t); i += 3 {
			toks = append(toks, t[i:i+3])
		}
	}
	for _, t := range toks {
		v, err := strconv.ParseUint(t, base, 16)
		if err != nil || v > 255 {
			return nil, fmt.Errorf("value %q is not a byte (0-255)", t)
		}
		out = append(out, byte(v))
	}
	for _, r := range s {
		if !isDigit(r) && !strings.ContainsRune(" \t\r\n,;:|\\.-_/[](){}'\"", r) {
			return nil, fmt.Errorf("unexpected character %q", r)
		}
	}
	return out, nil
}

func parseBinary(s string) ([]byte, error) {
	s = strings.NewReplacer("0b", "", "0B", "").Replace(s)
	toks := strings.Fields(strings.NewReplacer(",", " ", ";", " ", "_", " ", "|", " ").Replace(s))
	for _, t := range toks {
		for i := 0; i < len(t); i++ {
			if t[i] != '0' && t[i] != '1' {
				return nil, fmt.Errorf("invalid binary digit %q", string(rune(t[i])))
			}
		}
	}
	width := 8
	if len(toks) > 1 {
		same := true
		for _, t := range toks {
			if len(t) != len(toks[0]) {
				same = false
			}
		}
		if same && len(toks[0]) < 8 && len(toks[0]) >= 5 {
			width = len(toks[0])
		}
	}
	all := strings.Join(toks, "")
	if len(all) == 0 {
		return nil, nil
	}
	if width == 8 && len(all)%8 != 0 {
		return nil, fmt.Errorf("binary input has %d bits, which is not a multiple of 8", len(all))
	}
	var out []byte
	for i := 0; i+width <= len(all); i += width {
		v, _ := strconv.ParseUint(all[i:i+width], 2, 16)
		out = append(out, byte(v))
	}
	return out, nil
}

// ---- Accepts helpers ----

// looksText reports whether b is non-empty valid UTF-8 without stray control
// characters, i.e. something a human would call text.
func looksText(b []byte) bool {
	if len(b) == 0 || !utf8.Valid(b) {
		return false
	}
	for _, r := range string(b) {
		if r < 0x20 && r != '\n' && r != '\r' && r != '\t' || r == 0x7f || r == 0xfffd {
			return false
		}
	}
	return true
}

// acceptText builds an Accepts that requires dec to succeed and produce text.
func acceptText(dec func(string) ([]byte, error)) func(string) bool {
	return func(s string) bool {
		if strings.TrimSpace(s) == "" {
			return false
		}
		b, err := dec(s)
		return err == nil && looksText(b)
	}
}

// acceptStr builds an Accepts from a decoder that returns a string: it must
// succeed and change the text.
func acceptStr(dec func(string) (string, error)) func(string) bool {
	return func(s string) bool {
		if strings.TrimSpace(s) == "" {
			return false
		}
		r, err := dec(s)
		return err == nil && r != s
	}
}

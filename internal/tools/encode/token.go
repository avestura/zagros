package encode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/avestura/zagros/internal/tools"
)

type jwtParts struct {
	header, payload []byte
	sig             string
}

func jwtSplit(tok string) (*jwtParts, error) {
	tok = strings.TrimSpace(tok)
	if len(tok) > 7 && strings.EqualFold(tok[:7], "bearer ") {
		tok = strings.TrimSpace(tok[7:])
	}
	parts := strings.Split(tok, ".")
	if len(parts) == 5 {
		return nil, fmt.Errorf("this looks like an encrypted JWT (JWE, 5 parts); only signed JWTs (3 parts) can be decoded")
	}
	if len(parts) != 3 {
		return nil, fmt.Errorf("a JWT has 3 dot-separated parts, found %d", len(parts))
	}
	dec := func(name, s string) ([]byte, error) {
		b, err := parseBase64(s)
		if err != nil {
			return nil, fmt.Errorf("JWT %s is not valid base64url: %v", name, err)
		}
		b = bytes.TrimSpace(b)
		if len(b) == 0 || b[0] != '{' || !json.Valid(b) {
			return nil, fmt.Errorf("JWT %s is not a JSON object", name)
		}
		return b, nil
	}
	h, err := dec("header", parts[0])
	if err != nil {
		return nil, err
	}
	p, err := dec("payload", parts[1])
	if err != nil {
		return nil, err
	}
	return &jwtParts{h, p, parts[2]}, nil
}

func compactJSON(b []byte) string {
	var buf bytes.Buffer
	if err := json.Compact(&buf, b); err != nil {
		return string(b)
	}
	return buf.String()
}

func jwtDecode(tok string) (map[string]string, error) {
	j, err := jwtSplit(tok)
	if err != nil {
		return nil, err
	}
	res := map[string]string{"header": compactJSON(j.header), "payload": compactJSON(j.payload), "signature": j.sig}
	var h map[string]any
	if json.Unmarshal(j.header, &h) == nil {
		for _, k := range []string{"alg", "typ", "kid"} {
			if v, ok := h[k].(string); ok {
				res[k] = v
			}
		}
	}
	d := json.NewDecoder(bytes.NewReader(j.payload))
	d.UseNumber()
	var p map[string]any
	if d.Decode(&p) == nil {
		for _, k := range []string{"iss", "sub", "jti"} {
			if v, ok := p[k].(string); ok {
				res[k] = v
			}
		}
		switch v := p["aud"].(type) {
		case string:
			res["aud"] = v
		case []any:
			var as []string
			for _, x := range v {
				as = append(as, fmt.Sprint(x))
			}
			res["aud"] = strings.Join(as, ",")
		}
		for _, k := range []string{"exp", "iat", "nbf"} {
			if n, ok := p[k].(json.Number); ok {
				res[k] = n.String()
				if f, err := n.Float64(); err == nil && f > -62135596800 && f < 253402300799 {
					res[k+"_utc"] = time.Unix(int64(f), 0).UTC().Format("2006-01-02 15:04:05 UTC")
				}
			}
		}
	}
	return res, nil
}

type detector struct {
	fn    string
	score int
}

var detectors = []detector{
	{"jwt_decode", 100}, {"data_uri_decode", 100}, {"uudecode", 95}, {"braille_decode", 95},
	{"idna_to_unicode", 90}, {"email_to_unicode", 89}, {"unicode_unescape", 90}, {"js_unescape", 90},
	{"punycode_decode", 85}, {"html_unescape", 85}, {"morse_decode", 85}, {"binary_to_text", 85},
	{"double_url_decode", 82}, {"url_decode", 80}, {"hex_bytes_to_text", 80}, {"nato_decode", 75},
	{"quoted_printable_decode", 70}, {"hex_decode", 70}, {"c_unescape", 68}, {"base58check_decode", 65},
	{"json_unescape", 64}, {"base64url_decode", 62}, {"base64_decode", 60}, {"mojibake_fix", 58},
	{"decimal_bytes_to_text", 55}, {"strip_html_tags", 50}, {"base32_decode", 50}, {"octal_to_text", 50},
	{"a1z26_decode", 50}, {"tap_code_decode", 50}, {"query_string_parse", 45}, {"go_unquote", 40},
	{"python_unrepr", 40}, {"css_unescape", 38}, {"base45_decode", 35}, {"base58_decode", 30},
	{"csv_unescape", 30}, {"base62_decode", 25}, {"base36_decode", 20}, {"base85_decode", 20}, {"base91_decode", 15},
}

func detectEncodings(s string) []string {
	if strings.TrimSpace(s) == "" || len(s) > 1<<20 {
		return nil
	}
	type hit struct {
		name  string
		score int
	}
	var hits []hit
	for _, d := range detectors {
		acc := accepts[d.fn]
		if acc != nil && acc(s) {
			hits = append(hits, hit{d.fn, d.score})
		}
	}
	sort.SliceStable(hits, func(i, j int) bool { return hits[i].score > hits[j].score })
	out := make([]string, len(hits))
	for i, h := range hits {
		out[i] = h.name
	}
	return out
}

func init() {
	reg(tools.Func{
		Name: "jwt_decode", Title: "JWT decode (no verification)",
		Doc:     "Decodes a JSON Web Token WITHOUT verifying its signature: returns header and payload as compact JSON, the raw signature, and alg, typ, kid, iss, sub, aud, jti when present. Numeric exp, iat and nbf are returned as given plus a human UTC time (exp_utc, iat_utc, nbf_utc). A leading \"Bearer \" is ignored; encrypted JWTs (5 parts) are rejected.",
		Params:  []tools.Param{pReq("token", "The JWT, header.payload.signature.")},
		Returns: "dict[str, str]",
		Fn:      func(a tools.Args) (any, error) { return jwtDecode(a.Str("token")) },
		Accepts: func(s string) bool { _, err := jwtSplit(s); return err == nil },
		Example: `encode.jwt_decode("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwiaWF0IjoxNTE2MjM5MDIyfQ.sig")`,
		Tags:    tagsWith("jwt", "json web token", "token", "bearer", "claims", "exp", "iat", "oauth", "decode", "jws"),

		Keywords: []string{"jwt", "token"},
	})
	reg(tools.Func{
		Name: "jwt_parts", Title: "JWT parts (header, payload, signature)",
		Doc:     "Splits a JWT and returns [header JSON, payload JSON, signature] as three strings, the first two base64url-decoded and compacted. No signature verification.",
		Params:  []tools.Param{pReq("token", "The JWT.")},
		Returns: "list[str]",
		Fn: func(a tools.Args) (any, error) {
			j, err := jwtSplit(a.Str("token"))
			if err != nil {
				return nil, err
			}
			return []string{compactJSON(j.header), compactJSON(j.payload), j.sig}, nil
		},
		Accepts: func(s string) bool { _, err := jwtSplit(s); return err == nil },
		Example: `encode.jwt_parts("eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhIn0.sig")`,
		Tags:    tagsWith("jwt", "json web token", "token", "split", "header", "payload", "signature"),

		Keywords: []string{"jwt", "token"},
	})
	reg(tools.Func{
		Name: "detect", Title: "Detect encoding",
		Doc:     "Lists the decoder functions of this module that plausibly apply to the text, best first (for example [\"jwt_decode\"], [\"url_decode\"], [\"hex_decode\", \"base64_decode\"]). Each entry is a function name callable as encode.<name>(text). Heuristic: it only reports decoders whose output would be readable text.",
		Params:  []tools.Param{pReq("text", "Unknown encoded text.")},
		Returns: "list[str]",
		Fn:      func(a tools.Args) (any, error) { return detectEncodings(a.Str("text")), nil },
		Accepts: func(s string) bool { return len(detectEncodings(s)) > 0 },
		Example: `encode.detect("SGVsbG8sIFdvcmxkIQ==")`,
		Tags:    tagsWith("detect", "identify", "auto decode", "which encoding", "guess", "base64", "hex", "url", "html", "morse", "cipher identifier"),

		Keywords: []string{"detect", "auto", "decode"},
		Order:    9,
	})
}

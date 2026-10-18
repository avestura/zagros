package encode

import (
	"encoding/base64"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

const hexUpper = "0123456789ABCDEF"

func isAlnum(c byte) bool {
	return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9'
}

// urlEncodeBytes percent-encodes b keeping the bytes in safe; in form mode a
// space becomes "+".
func urlEncodeBytes(b []byte, safe string, form bool) string {
	var sb strings.Builder
	for _, c := range b {
		switch {
		case isAlnum(c) || strings.IndexByte(safe, c) >= 0:
			sb.WriteByte(c)
		case c == ' ' && form:
			sb.WriteByte('+')
		default:
			sb.WriteByte('%')
			sb.WriteByte(hexUpper[c>>4])
			sb.WriteByte(hexUpper[c&15])
		}
	}
	return sb.String()
}

func urlMode(mode string) (safe string, form bool, err error) {
	switch normName(mode) {
	case "", "component":
		return "-_.!~*'()", false, nil
	case "query", "form":
		return "*-._", true, nil
	case "path":
		return "-_.!~*'()/:@&=+$,;", false, nil
	case "full", "uri":
		return ";,/?:@&=+$-_.!~*'()#", false, nil
	case "rfc3986", "strict":
		return "-._~", false, nil
	}
	return "", false, fmt.Errorf("unknown mode %q (use component, query, path, full or rfc3986)", mode)
}

func urlDecodeString(s string, plus bool, charset string, out string) (string, error) {
	b := percentDecode(s, plus)
	if normName(out) != "" && normName(out) != "text" {
		return toRep(b, out)
	}
	cs := normName(charset)
	if cs == "" || cs == "utf-8" || cs == "utf8" {
		if !utf8.Valid(b) {
			return "", fmt.Errorf("decoded bytes are not valid UTF-8; try charset=\"windows-1252\" or output=\"hex\"")
		}
		return string(b), nil
	}
	return csDecode(b, cs, false)
}

func hasPercentSeq(s string) bool {
	for i := 0; i+2 < len(s); i++ {
		if s[i] == '%' && isHexDigit(s[i+1]) && isHexDigit(s[i+2]) {
			return true
		}
	}
	return false
}

// parseQuery parses a query string into a map, joining repeated keys with ",".
func parseQuery(qs string) (map[string]string, error) {
	if i := strings.IndexByte(qs, '?'); i >= 0 && !strings.Contains(qs[:i], "=") {
		qs = qs[i+1:]
	} else if strings.Contains(qs, "://") {
		qs = ""
	}
	if j := strings.Index(qs, "#"); j >= 0 {
		qs = qs[:j]
	}
	qs = strings.TrimPrefix(qs, "?")
	res := map[string]string{}
	for _, part := range strings.FieldsFunc(qs, func(r rune) bool { return r == '&' || r == ';' }) {
		k, v := part, ""
		if i := strings.IndexByte(part, '='); i >= 0 {
			k, v = part[:i], part[i+1:]
		}
		dk := string(percentDecode(k, true))
		dv := string(percentDecode(v, true))
		if !utf8.ValidString(dk) || !utf8.ValidString(dv) {
			return nil, fmt.Errorf("query contains percent-encoded bytes that are not valid UTF-8")
		}
		if old, ok := res[dk]; ok {
			res[dk] = old + "," + dv
		} else {
			res[dk] = dv
		}
	}
	return res, nil
}

func buildQuery(m map[string]string, prefix bool) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, len(keys))
	for i, k := range keys {
		parts[i] = urlEncodeBytes([]byte(k), "*-._", true) + "=" + urlEncodeBytes([]byte(m[k]), "*-._", true)
	}
	s := strings.Join(parts, "&")
	if prefix && s != "" {
		s = "?" + s
	}
	return s
}

func dataURIEncode(b []byte, mime string, b64 bool, charset string) string {
	if mime == "" {
		mime = "text/plain"
	}
	var sb strings.Builder
	sb.WriteString("data:" + mime)
	if charset != "" {
		sb.WriteString(";charset=" + charset)
	}
	if b64 {
		sb.WriteString(";base64," + base64.StdEncoding.EncodeToString(b))
	} else {
		sb.WriteString("," + urlEncodeBytes(b, "-_.!~*'()", false))
	}
	return sb.String()
}

func dataURIDecode(uri string) (map[string]string, error) {
	uri = strings.TrimSpace(uri)
	if len(uri) < 5 || !strings.EqualFold(uri[:5], "data:") {
		return nil, fmt.Errorf("not a data URI (must start with \"data:\")")
	}
	rest := uri[5:]
	i := strings.IndexByte(rest, ',')
	if i < 0 {
		return nil, fmt.Errorf("data URI has no comma before the data")
	}
	meta, data := rest[:i], rest[i+1:]
	mime, charset, isB64 := "text/plain", "", false
	for n, p := range strings.Split(meta, ";") {
		p = strings.TrimSpace(p)
		switch {
		case strings.EqualFold(p, "base64"):
			isB64 = true
		case strings.HasPrefix(strings.ToLower(p), "charset="):
			charset = p[8:]
		case n == 0 && p != "":
			mime = p
		}
	}
	var b []byte
	if isB64 {
		var err error
		b, err = parseBase64(string(percentDecode(data, false)))
		if err != nil {
			return nil, err
		}
	} else {
		b = percentDecode(data, false)
	}
	res := map[string]string{"mime": mime, "charset": charset, "base64": fmt.Sprint(isB64), "size": fmt.Sprint(len(b)), "data_base64": base64.StdEncoding.EncodeToString(b)}
	if charset != "" {
		if s, err := csDecode(b, charset, false); err == nil {
			res["text"] = s
			return res, nil
		}
	}
	if utf8.Valid(b) {
		res["text"] = string(b)
	} else {
		res["text"] = ""
	}
	return res, nil
}

func init() {
	urlModes := []string{"component", "query", "path", "full", "rfc3986"}
	reg(tools.Func{
		Name: "url_encode", Title: "URL encode", Order: 3,
		Doc:     "Percent-encodes text as UTF-8. Modes: component (encodeURIComponent semantics, default), query (form style: space becomes +), path (keeps / : @ & = + $ , ;), full (encodeURI semantics, keeps URL syntax characters) and rfc3986 (only unreserved characters stay). charset encodes with another charset first (e.g. windows-1252 for %E9).",
		Params:  []tools.Param{pReq("text", "Text to encode."), pStr("mode", "Which characters stay unescaped.", "component", urlModes...), pStr("charset", "Charset used to turn text into bytes.", "utf-8"), pInEnc()},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			safe, form, err := urlMode(a.Str("mode"))
			if err != nil {
				return nil, err
			}
			var b []byte
			if cs := normName(a.Str("charset")); cs != "" && cs != "utf-8" && cs != "utf8" {
				b, err = csEncode(a.Str("text"), cs, false)
			} else {
				b, err = fromRep(a.Str("text"), a.Str("input_encoding"))
			}
			if err != nil {
				return nil, err
			}
			return urlEncodeBytes(b, safe, form), nil
		},
		Example:  `encode.url_encode("a b&c=d/é", mode="component")`,
		Tags:     tagsWith("url", "percent", "percent-encoding", "uri", "encodeURIComponent", "encodeURI", "urlencode", "query string", "rfc3986", "form"),
		Keywords: []string{"url", "uri"},
	})
	reg(tools.Func{
		Name: "url_decode", Title: "URL decode", Order: 4,
		Doc:     "Decodes %XX percent escapes (UTF-8 by default). Tolerant: malformed % sequences stay literal; plus_as_space turns + into a space (form data). charset decodes legacy encodings such as windows-1252; non-UTF-8 bytes are an error unless a charset or output is given.",
		Params:  []tools.Param{pReq("text", "Percent-encoded text."), pBool("plus_as_space", "Treat + as a space (application/x-www-form-urlencoded).", false), pStr("charset", "Charset of the decoded bytes.", "utf-8"), pOut()},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			return urlDecodeString(a.Str("text"), a.Bool("plus_as_space"), a.Str("charset"), a.Str("output"))
		},
		Accepts:  hasPercentSeq,
		Example:  `encode.url_decode("a%20b%26c%3Dd%2F%C3%A9")`,
		Tags:     tagsWith("url", "percent", "percent-decoding", "uri", "decodeURIComponent", "decodeURI", "urldecode", "unquote"),
		Keywords: []string{"url", "uri"},
	})
	reg(tools.Func{
		Name: "url_encode_all", Title: "URL encode every byte",
		Doc:     "Percent-encodes every byte of the text, including letters and digits (\"A\" becomes %41). Useful for obfuscation tests and WAF checks; url_decode reverses it.",
		Params:  []tools.Param{pReq("text", "Text to encode."), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return percentEncodeAll(b), nil
		}),
		Example:  `encode.url_encode_all("admin")`,
		Tags:     tagsWith("url", "percent", "encode all", "obfuscate", "every byte", "waf", "%41"),
		Keywords: []string{"url"},
	})
	reg(tools.Func{
		Name: "double_url_decode", Title: "URL decode repeatedly",
		Doc:     "Percent-decodes up to `times` rounds (default 2), stopping early when nothing changes, for double or triple encoded input such as %2527 or %253Cscript%253E. The final bytes must be valid UTF-8 unless output is hex/base64.",
		Params:  []tools.Param{pReq("text", "Percent-encoded text."), pInt("times", "Maximum number of rounds (1-10).", 2), pBool("plus_as_space", "Treat + as a space.", false), pOut()},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			n := int(a.Int("times"))
			if n < 1 || n > 10 {
				return nil, fmt.Errorf("times must be between 1 and 10")
			}
			b := []byte(a.Str("text"))
			for i := 0; i < n; i++ {
				nb := percentDecode(string(b), a.Bool("plus_as_space"))
				if string(nb) == string(b) {
					break
				}
				b = nb
			}
			return toRep(b, a.Str("output"))
		},
		Accepts: func(s string) bool {
			return hasPercentSeq(s) && hasPercentSeq(string(percentDecode(s, false)))
		},
		Example:  `encode.double_url_decode("%253Cscript%253E")`,
		Tags:     tagsWith("url", "percent", "double encoded", "nested", "decode", "%25", "waf"),
		Keywords: []string{"url"},
	})
	reg(tools.Func{
		Name: "query_string_parse", Title: "Parse query string",
		Doc:     "Parses a query string (leading ? and #fragment ignored; a full URL is accepted and its query used) into a dict. Keys and values are percent-decoded with + as space; repeated keys are joined with \",\"; a key without = maps to \"\". Pairs split on & or ;.",
		Params:  []tools.Param{pReq("qs", "Query string such as \"a=1&b=two&a=3\".")},
		Returns: "dict[str, str]",
		Fn:      func(a tools.Args) (any, error) { return parseQuery(a.Str("qs")) },
		Accepts: func(s string) bool {
			t := strings.TrimSpace(s)
			return strings.Contains(t, "=") && !strings.ContainsAny(t, " \n") && len(t) >= 3
		},
		Example:  `encode.query_string_parse("?a=1&b=two%20words&a=3")`,
		Tags:     tagsWith("query string", "querystring", "url params", "parse", "form", "urlparse", "parse_qs", "get parameters"),
		Keywords: []string{"url", "query", "qs"},
	})
	regAPI(tools.Func{
		Name: "query_string_build", Title: "Build query string",
		Doc:     "Builds a query string from a dict: keys sorted alphabetically, keys and values form-encoded (space becomes +). prefix=true adds a leading ?. An empty dict gives an empty string.",
		Params:  []tools.Param{pDict("params", "Key/value pairs."), pBool("prefix", "Add a leading ?.", false)},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			m, _ := a["params"].(map[string]string)
			return buildQuery(m, a.Bool("prefix")), nil
		},
		Example: `encode.query_string_build({"q": "hello world", "page": "2"})`,
		Tags:    tagsWith("query string", "querystring", "url params", "build", "urlencode", "form", "urlencode dict"),
	})
	regAPI(tools.Func{
		Name: "form_urlencode", Title: "Form URL-encode (application/x-www-form-urlencoded)",
		Doc:     "Encodes a dict as application/x-www-form-urlencoded (sorted keys, space as +). Same as query_string_build without a prefix.",
		Params:  []tools.Param{pDict("params", "Key/value pairs.")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			m, _ := a["params"].(map[string]string)
			return buildQuery(m, false), nil
		},
		Example: `encode.form_urlencode({"name": "Ada Lovelace", "lang": "en"})`,
		Tags:    tagsWith("form", "urlencoded", "application/x-www-form-urlencoded", "post body", "urlencode dict"),
	})
	reg(tools.Func{
		Name: "form_urldecode", Title: "Form URL-decode",
		Doc:     "Decodes an application/x-www-form-urlencoded body into a dict (+ is a space, repeated keys joined with \",\"). Same as query_string_parse.",
		Params:  []tools.Param{pReq("text", "Form body such as \"name=Ada+Lovelace&lang=en\".")},
		Returns: "dict[str, str]",
		Fn:      func(a tools.Args) (any, error) { return parseQuery(a.Str("text")) },
		Accepts: func(s string) bool {
			t := strings.TrimSpace(s)
			return strings.Contains(t, "=") && !strings.ContainsAny(t, " \n") && len(t) >= 3
		},
		Example:  `encode.form_urldecode("name=Ada+Lovelace&lang=en")`,
		Tags:     tagsWith("form", "urlencoded", "application/x-www-form-urlencoded", "post body", "parse"),
		Keywords: []string{"url", "form"},
	})
	reg(tools.Func{
		Name: "data_uri_encode", Title: "Data URI encode",
		Doc:     "Builds a data: URI (RFC 2397) from text, e.g. data:text/plain;base64,SGk=. base64=false percent-encodes the data instead; charset (e.g. \"utf-8\") adds a ;charset= parameter. Use input_encoding=base64 or hex to embed binary data.",
		Params:  []tools.Param{pReq("text", "Content."), pStr("mime", "Media type.", "text/plain"), pBool("base64", "Use base64 (otherwise percent-encoding).", true), pStr("charset", "Optional charset parameter, empty for none.", ""), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return dataURIEncode(b, a.Str("mime"), a.Bool("base64"), a.Str("charset")), nil
		}),
		Example:  `encode.data_uri_encode("<h1>Hi</h1>", mime="text/html")`,
		Tags:     tagsWith("data uri", "data url", "rfc2397", "inline", "base64", "embed", "image"),
		Keywords: []string{"data", "uri"},
	})
	reg(tools.Func{
		Name: "data_uri_decode", Title: "Data URI decode",
		Doc:     "Parses a data: URI into a dict with mime, charset, base64 (\"true\"/\"false\"), size (bytes), text (decoded text, empty when the data is binary) and data_base64 (the raw bytes). Fails when the URI is not a data URI or the payload is malformed.",
		Params:  []tools.Param{pReq("uri", "A data: URI.")},
		Returns: "dict[str, str]",
		Fn:      func(a tools.Args) (any, error) { return dataURIDecode(a.Str("uri")) },
		Accepts: func(s string) bool {
			t := strings.TrimSpace(s)
			if len(t) < 6 || !strings.EqualFold(t[:5], "data:") {
				return false
			}
			_, err := dataURIDecode(t)
			return err == nil
		},
		Example:  `encode.data_uri_decode("data:text/plain;base64,SGVsbG8=")`,
		Tags:     tagsWith("data uri", "data url", "rfc2397", "parse", "base64", "decode"),
		Keywords: []string{"data", "uri"},
	})
}

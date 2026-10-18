package encode

import (
	"reflect"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
)

func call(t *testing.T, name string, kv ...any) any {
	t.Helper()
	m := map[string]any{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	r, err := tools.Call("encode", name, m)
	if err != nil {
		t.Fatalf("%s(%v): %v", name, kv, err)
	}
	return r
}

func callErr(name string, kv ...any) error {
	m := map[string]any{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	_, err := tools.Call("encode", name, m)
	return err
}

func eq(t *testing.T, got, want any, ctx string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Errorf("%s: got %#v, want %#v", ctx, got, want)
	}
}

func TestVectors(t *testing.T) {
	cases := []struct {
		fn, in, want string
		extra        []any
	}{
		{"base64_encode", "foobar", "Zm9vYmFy", nil},
		{"base64_encode", "fooba", "Zm9vYmE=", nil},
		{"base64_encode", "f", "Zg==", nil},
		{"base64_encode", "f", "Zg", []any{"variant", "standard_nopad"}},
		{"base64_encode", "fb ff", "-_8", []any{"variant", "url_nopad", "input_encoding", "hex"}},
		{"base64_encode", "fb ff", "-_8=", []any{"variant", "url", "input_encoding", "hex"}},
		{"base64_encode", "", "", nil},
		{"base64_encode", "😀é", "8J+YgMOp", nil},
		{"base64url_encode", "subjects?_d=1", "c3ViamVjdHM_X2Q9MQ", nil},
		{"base32_encode", "foobar", "MZXW6YTBOI======", nil},
		{"base32_encode", "foobar", "MZXW6YTBOI", []any{"padding", false}},
		{"base32_encode", "foobar", "CPNMUOJ1E8======", []any{"variant", "hex"}},
		{"hex_encode", "Hi", "4869", nil},
		{"hex_encode", "Hi", "0x48:0x69", []any{"separator", ":", "prefix", "0x"}},
		{"hex_encode", "ff", "FF", []any{"case", "upper", "input_encoding", "hex"}},
		{"text_to_hex_bytes", "Hi", "0x48 0x69", nil},
		{"base45_encode", "AB", "BB8", nil},
		{"base45_encode", "Hello!!", "%69 VD92EX0", nil},
		{"base45_encode", "base-45", "UJCLQE7W581", nil},
		{"base45_encode", "ietf!", "QED8WEX0", nil},
		{"base58_encode", "Hello World", "JxF12TrwUP45BMd", nil},
		{"base85_encode", "Man ", "9jqo^", nil},
		{"base85_encode", "Man ", "<~9jqo^~>", []any{"delimiters", true}},
		{"base85_encode", "864fd26fb559f75b", "HelloWorld", []any{"variant", "z85", "input_encoding", "hex"}},
		{"text_to_binary", "Hi", "01001000 01101001", nil},
		{"text_to_octal", "Hi", "110 151", nil},
		{"text_to_decimal_bytes", "Hel", "72 101 108", nil},
		{"text_to_bits", "A", "01000001 00000000", []any{"charset", "utf-16le"}},
		{"url_encode", "a b&c=d/é", "a%20b%26c%3Dd%2F%C3%A9", nil},
		{"url_encode", "a b", "a+b", []any{"mode", "query"}},
		{"url_encode", "http://x/a b?q=é#f", "http://x/a%20b?q=%C3%A9#f", []any{"mode", "full"}},
		{"url_encode", "é", "%E9", []any{"charset", "windows-1252"}},
		{"url_encode_all", "aB", "%61%42", nil},
		{"url_decode", "a%20b+c%zz%C3%A9", "a b+c%zzé", nil},
		{"url_decode", "a+b", "a b", []any{"plus_as_space", true}},
		{"url_decode", "%E9", "é", []any{"charset", "windows-1252"}},
		{"double_url_decode", "%253Cscript%253E", "<script>", nil},
		{"html_escape", `<a href="x">Tom & 'J'</a>`, `&lt;a href=&#34;x&#34;&gt;Tom &amp; &#39;J&#39;&lt;/a&gt;`, nil},
		{"html_unescape", "Tom &amp; Jerry &copy; &#x1F600; &#65;", "Tom & Jerry © 😀 A", nil},
		{"html_escape_all", "café © 😀 <", "caf&eacute; &copy; &#128512; &lt;", nil},
		{"html_numeric_entities", "Hi é", "Hi &#xE9;", []any{"hex", true}},
		{"xml_escape", "<a b='1'>&", "&lt;a b=&apos;1&apos;&gt;&amp;", nil},
		{"xml_unescape", "&lt;&amp;&#65;&nbsp;", "<&A&nbsp;", nil},
		{"strip_html_tags", "<p>Hello <b>world</b> &amp; all</p><script>x()</script>", "Hello world & all", nil},
		{"html_entity_name", "©", "&copy;", nil},
		{"html_entity_name", "😀", "&#128512;", nil},
		{"punycode_encode", "bücher", "bcher-kva", nil},
		{"punycode_encode", "münchen", "xn--mnchen-3ya", []any{"prefix", true}},
		{"punycode_encode", "日本語", "wgv71a119e", nil},
		{"punycode_decode", "xn--bcher-kva", "bücher", nil},
		{"punycode_decode", "wgv71a119e", "日本語", nil},
		{"idna_to_ascii", "München.de", "xn--mnchen-3ya.de", nil},
		{"idna_to_unicode", "xn--mnchen-3ya.de", "münchen.de", nil},
		{"email_to_ascii", "info@bücher.de", "info@xn--bcher-kva.de", nil},
		{"email_to_unicode", "info@xn--bcher-kva.de", "info@bücher.de", nil},
		{"js_escape", "café 😀 a+b", "caf%E9%20%uD83D%uDE00%20a+b", nil},
		{"js_unescape", "caf%E9 %uD83D%uDE00", "café 😀", nil},
		{"json_escape", "a\n\"b\"\\é", `a\n\"b\"\\é`, nil},
		{"json_escape", "é😀", "\\u00e9\\ud83d\\ude00", []any{"ascii_only", true}},
		{"json_quote", "hi", `"hi"`, nil},
		{"json_unescape", `"a\né😀\/"`, "a\né😀/", nil},
		{"c_escape", "a\tb\"\x01é", `a\tb\"\001é`, nil},
		{"c_escape", "é", `\303\251`, []any{"non_ascii", "octal"}},
		{"c_unescape", `a\tb\x41\101é`, "a\tbAAé", nil},
		{"go_quote", "h\né", `"h\né"`, nil},
		{"go_unquote", `"hé\n"`, "hé\n", nil},
		{"python_repr", "it's\t", `"it's\t"`, nil},
		{"python_repr", "a'b\"c", `'a\'b"c'`, nil},
		{"python_unrepr", `'it\'s\x41'`, "it'sA", nil},
		{"unicode_escape", "hé😀", "h\\u00e9\\ud83d\\ude00", nil},
		{"unicode_escape", "😀", `\U0001f600`, []any{"style", "python"}},
		{"unicode_unescape", "h\\u00e9 \\ud83d\\ude00 \\u{1F600} \\U0001F601", "hé 😀 😀 😁", nil},
		{"unicode_codepoint_escape", "a😀", `a\u{1F600}`, nil},
		{"css_escape", "1st .item", `\31 st\ \.item`, nil},
		{"css_unescape", `\31 st\ \.item`, "1st .item", nil},
		{"regex_escape", "a.b*c", `a\.b\*c`, nil},
		{"regex_unescape", `a\.b\*c\d`, `a.b*c\d`, nil},
		{"sql_escape", "O'Brien", "O''Brien", nil},
		{"sql_escape", "a'b\n", `a\'b\n`, []any{"dialect", "mysql"}},
		{"sql_unescape", "O''Brien", "O'Brien", nil},
		{"sql_quote", "O'Brien", "'O''Brien'", nil},
		{"sh_quote", "it's a file", `'it'\''s a file'`, nil},
		{"sh_quote", "safe-name.txt", "safe-name.txt", nil},
		{"sh_quote", "", "''", nil},
		{"powershell_quote", "it's", "'it''s'", nil},
		{"cmd_quote", `a "b" c\`, `"a \"b\" c\\"`, nil},
		{"cmd_quote", "plain", "plain", nil},
		{"csv_escape", `say "hi", ok`, `"say ""hi"", ok"`, nil},
		{"csv_escape", "plain", "plain", nil},
		{"csv_unescape", `"say ""hi"", ok"`, `say "hi", ok`, nil},
		{"quoted_printable_encode", "café = ok", "caf=C3=A9 =3D ok", nil},
		{"quoted_printable_decode", "caf=C3=A9 =3D ok=\r\n!", "café = ok!", nil},
		{"uuencode", "Cat", "begin 644 data\n#0V%T\n`\nend\n", nil},
		{"uudecode", "begin 644 data\n#0V%T\n`\nend\n", "Cat", nil},
		{"morse_encode", "SOS hi", "... --- ... / .... ..", nil},
		{"morse_encode", "é", "..-..", nil},
		{"morse_decode", "... --- ... / .... ..", "SOS HI", nil},
		{"nato_phonetic_encode", "Go 4", "Golf Oscar / Four", nil},
		{"nato_decode", "Golf Oscar / Four", "GO 4", nil},
		{"braille_encode", "Hi 42", "⠠⠓⠊⠀⠼⠙⠃", nil},
		{"braille_decode", "⠠⠓⠊⠀⠼⠙⠃", "Hi 42", nil},
		{"tap_code_encode", "hi", ".. ...  .. ....", nil},
		{"tap_code_decode", ".. ...  .. ....", "HI", nil},
		{"tap_code_decode", "23 24", "HI", nil},
		{"a1z26_encode", "Hello World", "8-5-12-12-15 23-15-18-12-4", nil},
		{"a1z26_decode", "8-5-12-12-15 23-15-18-12-4", "HELLO WORLD", nil},
		{"a1z26_decode", "8 9", "HI", nil},
		{"normalize_utf8", "é", "é", nil},
		{"mojibake_fix", "cafÃ©", "café", nil},
		{"mojibake_fix", "â€™ Ã©", "’ é", nil},
		{"mojibake_fix", "plain café", "plain café", nil},
		{"hex_to_base64", "48656c6c6f", "SGVsbG8=", nil},
		{"base64_to_hex", "SGVsbG8=", "48656c6c6f", nil},
		{"encode_charset", "café", "636166e9", []any{"charset", "windows-1252"}},
		{"encode_charset", "日本", "93fa967b", []any{"charset", "shift_jis"}},
		{"encode_charset", "A", "0041", []any{"charset", "utf-16be"}},
		{"decode_charset", "63 61 66 e9", "café", []any{"charset", "latin1"}},
		{"decode_charset", "93fa967b", "日本", []any{"charset", "sjis"}},
		{"decode_charset", "c4e3bac3", "你好", []any{"charset", "gbk"}},
		{"decode_charset", "c7d1b1db", "한글", []any{"charset", "euc-kr"}},
		{"decode_charset", "e0", "а", []any{"charset", "windows-1251"}},
	}
	for _, c := range cases {
		a := append([]any{}, c.extra...)
		f, _ := tools.Lookup("encode", c.fn)
		a = append(a, f.InputParam(), c.in)
		got := call(t, c.fn, a...)
		if got != c.want {
			t.Errorf("%s(%q %v): got %q, want %q", c.fn, c.in, c.extra, got, c.want)
		}
	}
}

func TestRoundTrips(t *testing.T) {
	samples := []string{"", "a", "ab", "abc", "abcd", "Hello, World!", "héllo wörld", "😀 emoji", "日本語のテキスト", strings.Repeat("xyz ", 50), "line1\nline2\ttab"}
	for _, algo := range baseNames() {
		for _, s := range samples {
			enc := call(t, "base_encode", "algo", algo, "text", s).(string)
			dec := call(t, "base_decode", "algo", algo, "text", enc).(string)
			if dec != s {
				t.Errorf("%s round trip of %q: enc %q dec %q", algo, s, enc, dec)
			}
		}
	}
	pairs := [][2]string{
		{"base64_encode", "base64_decode"}, {"base64url_encode", "base64url_decode"}, {"base32_encode", "base32_decode"},
		{"hex_encode", "hex_decode"}, {"text_to_hex_bytes", "hex_bytes_to_text"}, {"base36_encode", "base36_decode"},
		{"base45_encode", "base45_decode"}, {"base58_encode", "base58_decode"}, {"base62_encode", "base62_decode"},
		{"base85_encode", "base85_decode"}, {"base91_encode", "base91_decode"}, {"text_to_binary", "binary_to_text"},
		{"text_to_octal", "octal_to_text"}, {"text_to_decimal_bytes", "decimal_bytes_to_text"},
		{"url_encode", "url_decode"}, {"url_encode_all", "url_decode"}, {"html_escape", "html_unescape"},
		{"html_escape_all", "html_unescape"}, {"html_numeric_entities", "html_unescape"}, {"xml_escape", "xml_unescape"},
		{"js_escape", "js_unescape"}, {"json_escape", "json_unescape"}, {"json_quote", "json_unescape"}, {"c_escape", "c_unescape"},
		{"go_quote", "go_unquote"}, {"python_repr", "python_unrepr"}, {"unicode_escape", "unicode_unescape"},
		{"unicode_codepoint_escape", "unicode_codepoint_unescape"}, {"css_escape", "css_unescape"},
		{"regex_escape", "regex_unescape"}, {"sql_escape", "sql_unescape"}, {"csv_escape", "csv_unescape"},
		{"quoted_printable_encode", "quoted_printable_decode"}, {"uuencode", "uudecode"}, {"punycode_encode", "punycode_decode"},
	}
	for _, p := range pairs {
		for _, s := range samples {
			if s == "" && p[0] == "uuencode" {
				continue
			}
			if p[0] == "punycode_encode" && (strings.ContainsAny(s, " \n\t") || s == "") {
				continue
			}
			f, _ := tools.Lookup("encode", p[0])
			g, _ := tools.Lookup("encode", p[1])
			enc := call(t, p[0], f.InputParam(), s)
			out := call(t, p[1], g.InputParam(), enc)
			if p[0] == "quoted_printable_encode" {
				out = strings.ReplaceAll(out.(string), "\r\n", "\n")
				if out != strings.ReplaceAll(s, "\r\n", "\n") {
					t.Errorf("%s round trip of %q: enc %q dec %q", p[0], s, enc, out)
				}
				continue
			}
			if out != s {
				t.Errorf("%s/%s round trip of %q: enc %q dec %q", p[0], p[1], s, enc, out)
			}
		}
	}
	bin := "00 01 02 fe ff 80 00"
	for _, algo := range baseNames() {
		enc := call(t, "base_encode", "algo", algo, "text", bin, "input_encoding", "hex").(string)
		dec := call(t, "base_decode", "algo", algo, "text", enc, "output", "hex").(string)
		if dec != "000102feff8000" {
			t.Errorf("%s binary round trip: %q", algo, dec)
		}
	}
	for _, cs := range charsetNames() {
		enc := call(t, "encode_charset", "text", "Hello", "charset", cs).(string)
		dec := call(t, "decode_charset", "data", enc, "charset", cs).(string)
		if dec != "Hello" {
			t.Errorf("charset %s: %q", cs, dec)
		}
	}
}

func TestDecodeErrors(t *testing.T) {
	bad := []struct {
		fn string
		kv []any
	}{
		{"base64_decode", []any{"text", "!!!!"}},
		{"base64_decode", []any{"text", "/w=="}},
		{"hex_decode", []any{"text", "abc"}},
		{"hex_decode", []any{"text", "zz"}},
		{"base32_decode", []any{"text", "1!"}},
		{"base45_decode", []any{"text", "a"}},
		{"base58_decode", []any{"text", "0OIl"}},
		{"base58check_decode", []any{"text", "JxF12TrwUP45BMd"}},
		{"binary_to_text", []any{"bits", "0100"}},
		{"decimal_bytes_to_text", []any{"text", "300"}},
		{"punycode_decode", []any{"text", "ü"}},
		{"json_unescape", []any{"text", `\q`}},
		{"data_uri_decode", []any{"uri", "hello"}},
		{"jwt_decode", []any{"token", "a.b"}},
		{"base_encode", []any{"algo", "nope", "text", "x"}},
		{"url_decode", []any{"text", "%FF"}},
		{"morse_decode", []any{"text", "abc"}},
		{"ascii_table_lookup", []any{"char", "ab"}},
		{"base64_encode", []any{"text", "x", "variant", "nope"}},
		{"base64_decode", []any{"text", "AAAA", "output", "nope"}},
		{"encode_charset", []any{"text", "日", "charset", "iso-8859-1"}},
	}
	for _, b := range bad {
		if callErr(b.fn, b.kv...) == nil {
			t.Errorf("%s(%v): expected error", b.fn, b.kv)
		}
	}
	garbage := []string{"", "\x00\xff\xfe", "%", "%%%", "&", "&#;", "&#xZZ;", "\\", "\\u", "\\ud800", "=", "=\r", "begin", "<", "<~", "~>", "xn--", "xn--zzzz", "...---", "⠀", "1-", "99999999999", "é\xff"}
	for _, f := range tools.All() {
		if f.Group != "encode" {
			continue
		}
		for _, g := range garbage {
			m := map[string]any{}
			for _, p := range f.Params {
				switch p.Type {
				case "str":
					if p.Default == nil || p.Required {
						m[p.Name] = g
					}
				case "dict[str, str]":
					m[p.Name] = map[string]string{"k": g}
				}
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s panicked on %q: %v", f.Key(), g, r)
					}
				}()
				_, _ = f.Invoke(m)
				if f.Accepts != nil {
					f.Accepts(g)
				}
			}()
		}
	}
}

func TestAcceptsRejectsWrongInput(t *testing.T) {
	wrong := map[string][]string{
		"base64_decode":           {"hello world", "not base64!", "", "abc"},
		"base64url_decode":        {"hello", "a+b/"},
		"hex_decode":              {"hello", "xyz", "abc", "", "41 4"},
		"base32_decode":           {"hello", "abc"},
		"base45_decode":           {"hello world", "a"},
		"base58check_decode":      {"hello", "JxF12TrwUP45BMd"},
		"binary_to_text":          {"hello", "0101", "012"},
		"octal_to_text":           {"hello", "999"},
		"decimal_bytes_to_text":   {"hello", "300 400", "1"},
		"url_decode":              {"plain text", "100% sure"},
		"html_unescape":           {"plain text", "a & b"},
		"punycode_decode":         {"hello", "hello-world"},
		"idna_to_unicode":         {"example.com", "münchen.de"},
		"js_unescape":             {"plain", "%41"},
		"unicode_unescape":        {"plain", "\\n"},
		"json_unescape":           {"plain", `\q`},
		"c_unescape":              {"plain"},
		"morse_decode":            {"hello", "12", ""},
		"nato_decode":             {"hello world", "alpha"},
		"braille_decode":          {"hello", ""},
		"jwt_decode":              {"hello", "a.b.c", "x.y"},
		"data_uri_decode":         {"hello", "http://x"},
		"quoted_printable_decode": {"plain", "a=b"},
		"uudecode":                {"hello"},
		"a1z26_decode":            {"hello", "99 100"},
		"tap_code_decode":         {"hello", "99"},
		"csv_unescape":            {"plain"},
		"strip_html_tags":         {"plain text", "a < b"},
		"mojibake_fix":            {"plain", "café"},
		"query_string_parse":      {"plain", ""},
		"sql_unescape":            {"plain"},
	}
	for fn, ins := range wrong {
		f, ok := tools.Lookup("encode", fn)
		if !ok || f.Accepts == nil {
			t.Errorf("%s: missing Accepts", fn)
			continue
		}
		for _, in := range ins {
			if f.Accepts(in) {
				t.Errorf("%s.Accepts(%q) = true, want false", fn, in)
			}
		}
	}
	right := map[string][]string{
		"base64_decode":         {"SGVsbG8sIFdvcmxkIQ=="},
		"hex_decode":            {"48656c6c6f", "48 65 6c 6c 6f"},
		"base32_decode":         {"JBSWY3DPEBLW64TMMQ======"},
		"binary_to_text":        {"01001000 01101001"},
		"decimal_bytes_to_text": {"72 101 108 108 111"},
		"url_decode":            {"a%20b"},
		"html_unescape":         {"a &amp; b"},
		"idna_to_unicode":       {"xn--mnchen-3ya.de"},
		"js_unescape":           {"%u0041%u00e9"},
		"unicode_unescape":      {"\\u00e9"},
		"morse_decode":          {"... --- ..."},
		"nato_decode":           {"Alpha Bravo"},
		"braille_decode":        {"⠓⠊"},
		"jwt_decode":            {"eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhIn0.sig"},
		"data_uri_decode":       {"data:text/plain;base64,SGk="},
		"mojibake_fix":          {"cafÃ©"},
		"a1z26_decode":          {"8-9"},
		"punycode_decode":       {"xn--bcher-kva"},
		"base58check_decode":    {"1QLbz7JHiBTspS962RLKV8GndWFwi5j6Qr"},
	}
	for fn, ins := range right {
		f, _ := tools.Lookup("encode", fn)
		for _, in := range ins {
			if !f.Accepts(in) {
				t.Errorf("%s.Accepts(%q) = false, want true", fn, in)
			}
		}
	}
}

func TestDetect(t *testing.T) {
	cases := map[string]string{
		"SGVsbG8sIFdvcmxkIQ==":                     "base64_decode",
		"48 65 6c 6c 6f 20 57 6f 72 6c 64":         "hex_decode",
		"Hello%20World%21":                         "url_decode",
		"Tom &amp; Jerry":                          "html_unescape",
		"xn--mnchen-3ya.de":                        "idna_to_unicode",
		"eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhIn0.sig": "jwt_decode",
		"... --- ...":                              "morse_decode",
		"01001000 01101001":                        "binary_to_text",
		"caf\\u00e9":                               "unicode_unescape",
		"data:text/plain;base64,SGk=":              "data_uri_decode",
	}
	for in, want := range cases {
		got := call(t, "detect", "text", in).([]string)
		found := false
		for _, g := range got {
			if g == want {
				found = true
			}
		}
		if !found {
			t.Errorf("detect(%q) = %v, want it to contain %s", in, got, want)
		}
	}
	if got := call(t, "detect", "text", "just a plain sentence").([]string); len(got) != 0 {
		t.Errorf("detect(plain) = %v", got)
	}
}

func TestMaps(t *testing.T) {
	eq(t, call(t, "query_string_parse", "qs", "?a=1&b=two%20words&a=3&c&d=x+y#frag"),
		map[string]string{"a": "1,3", "b": "two words", "c": "", "d": "x y"}, "query_string_parse")
	eq(t, call(t, "query_string_parse", "qs", "https://x.com/p?z=1"), map[string]string{"z": "1"}, "parse url")
	eq(t, call(t, "query_string_build", "params", map[string]string{"q": "hello world", "a": "é&"}), "a=%C3%A9%26&q=hello+world", "build")
	eq(t, call(t, "form_urlencode", "params", map[string]string{"x": "1"}), "x=1", "form")
	eq(t, call(t, "form_urldecode", "text", "x=a+b"), map[string]string{"x": "a b"}, "formdec")

	d := call(t, "data_uri_decode", "uri", "data:text/plain;charset=utf-8;base64,SGVsbG8=").(map[string]string)
	eq(t, d["text"], "Hello", "data text")
	eq(t, d["mime"], "text/plain", "data mime")
	eq(t, d["size"], "5", "data size")
	eq(t, call(t, "data_uri_encode", "text", "Hi", "mime", "text/html"), "data:text/html;base64,SGk=", "data enc")
	eq(t, call(t, "data_uri_encode", "text", "a b", "base64", false), "data:text/plain,a%20b", "data enc plain")

	j := call(t, "jwt_decode", "token", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c").(map[string]string)
	eq(t, j["alg"], "HS256", "alg")
	eq(t, j["sub"], "1234567890", "sub")
	eq(t, j["iat_utc"], "2018-01-18 01:30:22 UTC", "iat_utc")
	eq(t, j["payload"], `{"sub":"1234567890","name":"John Doe","iat":1516239022}`, "payload")
	eq(t, call(t, "jwt_parts", "token", "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiJhIn0.sig"), []string{`{"alg":"HS256"}`, `{"sub":"a"}`, "sig"}, "jwt_parts")

	a := call(t, "ascii_table_lookup", "char", "A").(map[string]string)
	eq(t, a["decimal"], "65", "ascii dec")
	eq(t, a["hex"], "41", "ascii hex")
	eq(t, call(t, "ascii_table_lookup", "char", "\n").(map[string]string)["name"], "LF", "LF")

	m := call(t, "mojibake_try_all", "text", "cafÃ©").(map[string]string)
	eq(t, m["windows-1252"], "café", "try_all")

	b58 := call(t, "base58check_parse", "text", "1QLbz7JHiBTspS962RLKV8GndWFwi5j6Qr").(map[string]string)
	eq(t, b58["valid"], "true", "b58check valid")
	eq(t, b58["version"], "0", "b58check version")
	enc := call(t, "base58check_encode", "text", "00112233445566778899aabbccddeeff00112233", "input_encoding", "hex")
	eq(t, call(t, "base58check_decode", "text", enc, "output", "hex"), "00112233445566778899aabbccddeeff00112233", "b58check rt")

	if cs := call(t, "detect_charset", "data_hex", "ef bb bf 68 69").([]string); cs[0] != "utf-8-sig" {
		t.Errorf("detect_charset bom: %v", cs)
	}
	if cs := call(t, "detect_charset", "data_hex", "63 61 66 c3 a9").([]string); cs[0] != "utf-8" {
		t.Errorf("detect_charset utf8: %v", cs)
	}
	if cs := call(t, "detect_charset", "data_hex", "63 61 66 e9").([]string); cs[0] != "windows-1252" && cs[0] != "iso-8859-1" {
		t.Errorf("detect_charset latin: %v", cs)
	}
	if cs := call(t, "detect_charset", "data_hex", "68 00 69 00").([]string); cs[0] != "utf-16le" {
		t.Errorf("detect_charset utf16: %v", cs)
	}
}

func TestEntityTable(t *testing.T) {
	if len(entByRune) < 240 {
		t.Errorf("entity table pruned too much: %d", len(entByRune))
	}
}

func TestMimeAndMisc(t *testing.T) {
	long := strings.Repeat("A", 100)
	got := call(t, "base64_encode", "text", long, "variant", "mime").(string)
	if !strings.Contains(got, "\r\n") || len(strings.SplitN(got, "\r\n", 2)[0]) != 76 {
		t.Errorf("mime wrap: %q", got)
	}
	if call(t, "base64_decode", "text", got) != long {
		t.Error("mime decode")
	}
	if len(call(t, "bases").([]string)) < 15 || len(call(t, "charsets").([]string)) < 40 {
		t.Error("lists too short")
	}
	if err := callErr("base64_encode", "text", strings.Repeat("a", 9<<20)); err == nil {
		t.Error("expected 8 MiB refusal")
	}
	if got := call(t, "unescape_auto", "text", "caf\\u00e9 &amp; %3Cb%3E"); got != "café & <b>" {
		t.Errorf("unescape_auto: %q", got)
	}
}

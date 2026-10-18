package encode

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

var kwEsc = []string{"esc", "escape"}

func kw(extra ...string) []string { return append(append([]string{}, kwEsc...), extra...) }

func init() {
	reg(tools.Func{
		Name: "js_escape", Title: "JavaScript escape() (%uXXXX)",
		Doc:     "Legacy JavaScript escape(): keeps A-Z a-z 0-9 @ * _ + - . / and writes other UTF-16 code units as %XX (below 256) or %uXXXX, upper-case hex. Not the same as URL encoding; deprecated but still seen in old cookies and scripts.",
		Params:  []tools.Param{pReq("text", "Text to escape.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return jsEscape(a.Str("text")), nil },
		Example: `encode.js_escape("café 😀")`,
		Tags:    tagsWith("javascript", "escape", "%u", "unescape", "legacy", "utf-16", "percent"),

		Keywords: kw("js", "javascript"),
	})
	reg(tools.Func{
		Name: "js_unescape", Title: "JavaScript unescape() (%uXXXX)",
		Doc:     "Legacy JavaScript unescape(): decodes %XX (Latin-1 code unit) and %uXXXX sequences, joining surrogate pairs. Invalid sequences stay literal. Unlike url_decode, %E9 means é, not an invalid UTF-8 byte.",
		Params:  []tools.Param{pReq("text", "Escaped text.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return jsUnescape(a.Str("text")), nil },
		Accepts: func(s string) bool { return strings.Contains(s, "%u") && jsUnescape(s) != s },
		Example: `encode.js_unescape("caf%E9 %uD83D%uDE00")`,
		Tags:    tagsWith("javascript", "unescape", "%u", "escape", "legacy", "utf-16", "percent"),

		Keywords: kw("js", "javascript"),
	})
	reg(tools.Func{
		Name: "json_escape", Title: "JSON string escape",
		Doc:     "Escapes text for use inside a JSON string (without the surrounding quotes): \\\" \\\\ \\n \\t \\r, other control characters as \\u00XX. Non-ASCII is kept unless ascii_only=true, which writes \\uXXXX with surrogate pairs. < > & are not escaped.",
		Params:  []tools.Param{pReq("text", "Text to escape."), pBool("ascii_only", "Escape every non-ASCII character as \\uXXXX.", false)},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			s := jsonEscape(a.Str("text"), a.Bool("ascii_only"))
			return s[1 : len(s)-1], nil
		},
		Example: `encode.json_escape("line1\n\"quoted\"")`,
		Tags:    tagsWith("json", "escape", "string", "stringify", "quote", "\\n", "\\u"),

		Keywords: kw("json"),
	})
	reg(tools.Func{
		Name: "json_unescape", Title: "JSON string unescape",
		Doc:     "Decodes JSON string escapes (\\n \\t \\\" \\\\ \\/ \\b \\f \\r \\uXXXX with surrogate pairs). A surrounding pair of double quotes is removed. Fails on invalid escapes such as \\q or truncated \\u sequences.",
		Params:  []tools.Param{pReq("text", "Escaped JSON string contents.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return jsonUnescape(a.Str("text")) },
		Accepts: func(s string) bool { return strings.Contains(s, "\\") && acceptStr(jsonUnescape)(s) },
		Example: `encode.json_unescape("line1\\nline2 \\u00e9")`,
		Tags:    tagsWith("json", "unescape", "string", "parse", "\\n", "\\u"),

		Keywords: kw("json"),
	})
	reg(tools.Func{
		Name: "json_quote", Title: "JSON quote string",
		Doc:     "Returns the text as a complete JSON string literal, including the surrounding double quotes, with all required escapes. ascii_only=true escapes non-ASCII as \\uXXXX.",
		Params:  []tools.Param{pReq("text", "Text to quote."), pBool("ascii_only", "Escape every non-ASCII character.", false)},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return jsonEscape(a.Str("text"), a.Bool("ascii_only")), nil },
		Example: `encode.json_quote("say \"hi\"")`,
		Tags:    tagsWith("json", "quote", "string", "stringify", "literal"),

		Keywords: kw("json", "quote"),
	})
	reg(tools.Func{
		Name: "c_escape", Title: "C string escape",
		Doc:     "Escapes text for a C/C++ string literal: \\\\ \\\" \\n \\t \\r \\a \\b \\v \\f, other control characters as octal (\\001) or hex (\\x01). non_ascii controls what happens to characters above 127: keep, octal or hex (UTF-8 bytes) or unicode (\\uXXXX/\\UXXXXXXXX). quote=true adds surrounding quotes.",
		Params:  []tools.Param{pReq("text", "Text to escape."), pStr("control", "Style for other control characters.", "octal", "octal", "hex"), pStr("non_ascii", "How to treat non-ASCII characters.", "keep", "keep", "octal", "hex", "unicode"), pBool("quote", "Wrap the result in double quotes.", false)},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			return cEscape(a.Str("text"), a.Str("control"), a.Str("non_ascii"), a.Bool("quote")), nil
		},
		Example: `encode.c_escape("tab\there \"q\"")`,
		Tags:    tagsWith("c", "c++", "escape", "string literal", "octal", "\\x", "\\n", "addslashes", "backslash"),

		Keywords: kw("c", "cstring"),
	})
	reg(tools.Func{
		Name: "c_unescape", Title: "C string unescape",
		Doc:     "Decodes C-style backslash escapes: \\n \\t \\r \\a \\b \\f \\v \\e \\\\ \\' \\\" \\?, octal \\NNN (1-3 digits), hex \\xHH, \\uXXXX and \\UXXXXXXXX. \\x and octal produce raw bytes, so the result must be valid UTF-8 unless output is hex/base64. Unknown escapes are kept as written.",
		Params:  []tools.Param{pReq("text", "Text with escapes."), pOut()},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			b, err := bsUnescape(a.Str("text"), "c")
			if err != nil {
				return nil, err
			}
			return toRep(b, a.Str("output"))
		},
		Accepts: func(s string) bool {
			return bsPattern.MatchString(s) && acceptStr(func(x string) (string, error) { return bsToText(x, "c") })(s)
		},
		Example: `encode.c_unescape("tab\\there \\x41\\101")`,
		Tags:    tagsWith("c", "c++", "unescape", "string literal", "octal", "\\x", "\\n", "stripslashes", "backslash"),

		Keywords: kw("c", "cstring"),
	})
	reg(tools.Func{
		Name: "go_quote", Title: "Go string quote",
		Doc:     "Quotes text as a Go double-quoted string literal (strconv.Quote), escaping control characters, quotes and invalid UTF-8. ascii_only=true escapes every non-ASCII character as \\u or \\U.",
		Params:  []tools.Param{pReq("text", "Text to quote."), pBool("ascii_only", "Escape non-ASCII characters.", false)},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			if a.Bool("ascii_only") {
				return strconv.QuoteToASCII(a.Str("text")), nil
			}
			return strconv.Quote(a.Str("text")), nil
		},
		Example: `encode.go_quote("héllo\n")`,
		Tags:    tagsWith("go", "golang", "quote", "string literal", "strconv", "escape"),

		Keywords: kw("go", "golang", "quote"),
	})
	reg(tools.Func{
		Name: "go_unquote", Title: "Go string unquote",
		Doc:     "Parses a Go string literal (\"double-quoted\", `raw` or 'c' character) into its value (strconv.Unquote). Input without quotes is treated as the contents of a double-quoted string. Fails on invalid escapes.",
		Params:  []tools.Param{pReq("text", "Go string literal.")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			s := strings.TrimSpace(a.Str("text"))
			if r, err := strconv.Unquote(s); err == nil {
				return r, nil
			}
			if len(s) > 0 && s[0] != '"' && s[0] != '`' && s[0] != '\'' {
				if r, err := strconv.Unquote(`"` + strings.ReplaceAll(s, `"`, `\"`) + `"`); err == nil {
					return r, nil
				}
			}
			return nil, fmt.Errorf("not a valid Go string literal")
		},
		Accepts: func(s string) bool {
			t := strings.TrimSpace(s)
			if len(t) < 2 || (t[0] != '"' && t[0] != '`') {
				return false
			}
			r, err := strconv.Unquote(t)
			return err == nil && r != t
		},
		Example: `encode.go_unquote("\"h\\u00e9llo\\n\"")`,
		Tags:    tagsWith("go", "golang", "unquote", "string literal", "strconv", "unescape"),

		Keywords: kw("go", "golang", "quote"),
	})
	reg(tools.Func{
		Name: "python_repr", Title: "Python repr() of a string",
		Doc:     "Returns the text as Python would print repr(str): single quotes (double quotes if the text contains ' but no \"), backslash escapes for control characters, non-printable characters as \\xNN, \\uNNNN or \\UNNNNNNNN; printable Unicode is kept.",
		Params:  []tools.Param{pReq("text", "Text to represent.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return pythonRepr(a.Str("text")), nil },
		Example: `encode.python_repr("it's\ta test")`,
		Tags:    tagsWith("python", "repr", "string literal", "quote", "escape"),

		Keywords: kw("python", "py", "repr"),
	})
	reg(tools.Func{
		Name: "python_unrepr", Title: "Python string literal to text",
		Doc:     "Parses a Python string literal ('...', \"...\", triple-quoted, with optional r/u prefix) or bare escaped contents into text: \\n \\t \\xNN \\uNNNN \\UNNNNNNNN octal and line continuations. Raw (r) literals are returned verbatim.",
		Params:  []tools.Param{pReq("text", "Python string literal.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return pythonUnrepr(a.Str("text")) },
		Accepts: func(s string) bool {
			t := strings.TrimSpace(s)
			return len(t) >= 2 && (t[0] == '\'' || t[0] == '"') && t[len(t)-1] == t[0] && acceptStr(pythonUnrepr)(s)
		},
		Example: `encode.python_unrepr("'it\\'s\\ta test'")`,
		Tags:    tagsWith("python", "repr", "string literal", "unquote", "unescape", "literal_eval"),

		Keywords: kw("python", "py", "repr"),
	})
	reg(tools.Func{
		Name: "unicode_escape", Title: "Unicode escape (\\uXXXX)",
		Doc:     "Writes non-ASCII characters as \\uXXXX escapes; characters above U+FFFF become surrogate pairs (\\uD83D\\uDE00, JavaScript/Java/JSON style) or \\UXXXXXXXX with style=python. all=true escapes ASCII too; upper=true uses upper-case hex digits.",
		Params:  []tools.Param{pReq("text", "Text to escape."), pBool("all", "Escape every character, ASCII included.", false), pStr("style", "How to write characters above U+FFFF.", "utf16", "utf16", "python"), pBool("upper", "Upper-case hex digits.", false)},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			return unicodeEscape(a.Str("text"), a.Bool("all"), a.Str("style"), a.Bool("upper")), nil
		},
		Example: `encode.unicode_escape("héllo 😀")`,
		Tags:    tagsWith("unicode", "escape", "\\u", "utf-16", "surrogate pair", "javascript", "java", "json", "ascii only"),

		Keywords: kw("unicode", "u"),
	})
	reg(tools.Func{
		Name: "unicode_unescape", Title: "Unicode unescape (\\uXXXX)",
		Doc:     "Decodes \\uXXXX (surrogate pairs are joined), \\u{1F600}, \\UXXXXXXXX, \\xHH and U+XXXX sequences to characters. Lone surrogates become U+FFFD; anything else is left as written.",
		Params:  []tools.Param{pReq("text", "Text with Unicode escapes.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return unicodeUnescape(a.Str("text")), nil },
		Accepts: func(s string) bool { return uniEscPattern.MatchString(s) },
		Example: `encode.unicode_unescape("h\\u00e9llo \\ud83d\\ude00")`,
		Tags:    tagsWith("unicode", "unescape", "\\u", "utf-16", "surrogate pair", "javascript", "json", "decode"),

		Keywords: kw("unicode", "u"),
	})
	reg(tools.Func{
		Name: "unicode_codepoint_escape", Title: "Unicode code point escape (\\u{...})",
		Doc:     "Writes non-ASCII characters as ES6/Rust style \\u{1F600} code point escapes (no surrogate pairs). all=true escapes ASCII too; upper=true uses upper-case hex.",
		Params:  []tools.Param{pReq("text", "Text to escape."), pBool("all", "Escape every character.", false), pBool("upper", "Upper-case hex digits.", true)},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			return codepointEscape(a.Str("text"), a.Bool("all"), a.Bool("upper")), nil
		},
		Example: `encode.unicode_codepoint_escape("a😀é")`,
		Tags:    tagsWith("unicode", "code point", "\\u{", "es6", "rust", "escape", "emoji"),

		Keywords: kw("unicode", "u"),
	})
	reg(tools.Func{
		Name: "unicode_codepoint_unescape", Title: "Unicode code point unescape (\\u{...})",
		Doc:     "Decodes \\u{1F600} code point escapes (and the other forms handled by unicode_unescape) to characters.",
		Params:  []tools.Param{pReq("text", "Text with \\u{...} escapes.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return unicodeUnescape(a.Str("text")), nil },
		Accepts: func(s string) bool { return strings.Contains(s, "\\u{") && unicodeUnescape(s) != s },
		Example: `encode.unicode_codepoint_unescape("a\\u{1F600}")`,
		Tags:    tagsWith("unicode", "code point", "\\u{", "es6", "rust", "unescape", "emoji"),

		Keywords: kw("unicode", "u"),
	})
	reg(tools.Func{
		Name: "css_escape", Title: "CSS identifier escape",
		Doc:     "Escapes text for use as a CSS identifier or selector part, following CSS.escape(): special characters get a backslash, control characters and a leading digit become hex escapes (\\31 ), NUL becomes U+FFFD.",
		Params:  []tools.Param{pReq("text", "Text to escape.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return cssEscape(a.Str("text")), nil },
		Example: `encode.css_escape("1st .item:hover")`,
		Tags:    tagsWith("css", "escape", "selector", "identifier", "CSS.escape", "queryselector"),

		Keywords: kw("css"),
	})
	reg(tools.Func{
		Name: "css_unescape", Title: "CSS escape decode",
		Doc:     "Decodes CSS escapes: \\ followed by 1-6 hex digits (plus one optional whitespace) is a code point, \\ followed by any other character is that character, \\ at a line end is dropped.",
		Params:  []tools.Param{pReq("text", "CSS escaped text.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return cssUnescape(a.Str("text")), nil },
		Accepts: func(s string) bool { return strings.Contains(s, "\\") && cssUnescape(s) != s },
		Example: `encode.css_unescape("\\31 st\\ .item")`,
		Tags:    tagsWith("css", "unescape", "selector", "identifier", "decode"),

		Keywords: kw("css"),
	})
	reg(tools.Func{
		Name: "regex_escape", Title: "Regex escape",
		Doc:     "Escapes all regular expression metacharacters (\\ . + * ? ( ) | [ ] { } ^ $) with a backslash so the text matches literally (Go regexp.QuoteMeta; also valid for PCRE, JavaScript and Python).",
		Params:  []tools.Param{pReq("text", "Text to escape.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return regexp.QuoteMeta(a.Str("text")), nil },
		Example: `encode.regex_escape("a.b*c(d)")`,
		Tags:    tagsWith("regex", "regexp", "escape", "quotemeta", "re.escape", "literal"),

		Keywords: kw("regex", "re"),
	})
	reg(tools.Func{
		Name: "regex_unescape", Title: "Regex unescape",
		Doc:     "Removes the backslash in front of punctuation (\\. \\* \\( \\\\ ...), reversing regex_escape. Escapes of letters and digits (\\d, \\n, \\1) are left alone because they are not literal escapes.",
		Params:  []tools.Param{pReq("text", "Escaped regex text.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return regexUnescape(a.Str("text")), nil },
		Accepts: func(s string) bool { return regexUnescape(s) != s },
		Example: `encode.regex_unescape("a\\.b\\*c\\(d\\)")`,
		Tags:    tagsWith("regex", "regexp", "unescape", "literal", "unquote"),

		Keywords: kw("regex", "re"),
	})
	reg(tools.Func{
		Name: "sql_escape", Title: "SQL string escape",
		Doc:     "Escapes text for a single-quoted SQL string literal: doubles single quotes (standard SQL, PostgreSQL, SQLite, SQL Server). dialect=\"mysql\" also backslash-escapes \\ \" NUL newline CR and Ctrl-Z. Prefer parameterized queries over escaping.",
		Params:  []tools.Param{pReq("text", "Text to escape."), pStr("dialect", "SQL dialect.", "standard", "standard", "mysql")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return sqlEscape(a.Str("text"), a.Str("dialect")) },
		Example: `encode.sql_escape("O'Brien")`,
		Tags:    tagsWith("sql", "escape", "quote", "single quote", "mysql_real_escape_string", "sqli"),

		Keywords: kw("sql"),
	})
	reg(tools.Func{
		Name: "sql_unescape", Title: "SQL string unescape",
		Doc:     "Reverses sql_escape: turns doubled single quotes back into one (standard) and, with dialect=\"mysql\", decodes backslash escapes as well.",
		Params:  []tools.Param{pReq("text", "Escaped SQL string contents."), pStr("dialect", "SQL dialect.", "standard", "standard", "mysql")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return sqlUnescape(a.Str("text"), a.Str("dialect")) },
		Accepts: func(s string) bool { return strings.Contains(s, "''") },
		Example: `encode.sql_unescape("O''Brien")`,
		Tags:    tagsWith("sql", "unescape", "quote", "single quote"),

		Keywords: kw("sql"),
	})
	reg(tools.Func{
		Name: "sql_quote", Title: "SQL quote string",
		Doc:     "Escapes the text and wraps it in single quotes as a SQL string literal ('O''Brien'). dialect=\"mysql\" uses backslash escapes in addition.",
		Params:  []tools.Param{pReq("text", "Text to quote."), pStr("dialect", "SQL dialect.", "standard", "standard", "mysql")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			s, err := sqlEscape(a.Str("text"), a.Str("dialect"))
			if err != nil {
				return nil, err
			}
			return "'" + s + "'", nil
		},
		Example: `encode.sql_quote("O'Brien")`,
		Tags:    tagsWith("sql", "quote", "string literal", "single quote"),

		Keywords: kw("sql", "quote"),
	})
	reg(tools.Func{
		Name: "sh_quote", Title: "Shell quote (POSIX sh / bash)",
		Doc:     "Quotes text so a POSIX shell reads it as one literal word: wraps it in single quotes and rewrites embedded ' as '\\''. Text made only of safe characters is returned unchanged unless always=true; empty text gives ''.",
		Params:  []tools.Param{pReq("text", "Text to quote."), pBool("always", "Quote even when not necessary.", false)},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return shQuote(a.Str("text"), a.Bool("always")), nil },
		Example: `encode.sh_quote("it's a file.txt")`,
		Tags:    tagsWith("shell", "bash", "sh", "posix", "quote", "escape", "shlex.quote", "escapeshellarg", "argument"),

		Keywords: kw("sh", "shell", "bash", "quote"),
	})
	reg(tools.Func{
		Name: "powershell_quote", Title: "PowerShell quote",
		Doc:     "Quotes text as a PowerShell single-quoted (verbatim) string: wraps it in ' and doubles every ' and typographic single quote, so no variable expansion or escapes apply.",
		Params:  []tools.Param{pReq("text", "Text to quote.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return psQuote(a.Str("text")), nil },
		Example: `encode.powershell_quote("it's $HOME")`,
		Tags:    tagsWith("powershell", "pwsh", "quote", "escape", "single quote", "verbatim", "argument"),

		Keywords: kw("powershell", "ps", "pwsh", "quote"),
	})
	reg(tools.Func{
		Name: "cmd_quote", Title: "Windows command-line quote (cmd.exe)",
		Doc:     "Quotes an argument for the Windows command line using the CommandLineToArgvW rules: wrapped in double quotes when it has spaces or quotes, with backslashes before quotes doubled and quotes written as \\\". for_cmd=true also escapes cmd.exe metacharacters ( ) % ! ^ \" & | < > with ^ for use at a cmd prompt.",
		Params:  []tools.Param{pReq("text", "Argument to quote."), pBool("for_cmd", "Also caret-escape cmd.exe metacharacters.", false)},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return cmdQuote(a.Str("text"), a.Bool("for_cmd")), nil },
		Example: `encode.cmd_quote("C:\\Program Files\\App\\a \"b\".exe")`,
		Tags:    tagsWith("cmd", "cmd.exe", "windows", "command line", "quote", "escape", "argv", "batch"),

		Keywords: kw("cmd", "windows", "quote"),
	})
	reg(tools.Func{
		Name: "csv_escape", Title: "CSV field escape (RFC 4180)",
		Doc:     "Quotes one CSV field when needed (it contains the delimiter, a double quote or a line break) and doubles embedded quotes. always=true quotes every field; delimiter can be \";\" or a tab.",
		Params:  []tools.Param{pReq("text", "Field value."), pStr("delimiter", "Field delimiter.", ","), pBool("always", "Always add quotes.", false)},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			d := a.Str("delimiter")
			if d == "" {
				d = ","
			}
			return csvEscape(a.Str("text"), d, a.Bool("always")), nil
		},
		Example: `encode.csv_escape("say \"hi\", ok")`,
		Tags:    tagsWith("csv", "escape", "quote", "rfc4180", "field", "tsv", "spreadsheet"),

		Keywords: kw("csv"),
	})
	reg(tools.Func{
		Name: "csv_unescape", Title: "CSV field unescape",
		Doc:     "Reverses csv_escape for one field: removes the surrounding double quotes and turns doubled quotes into single quotes. Text without surrounding quotes is returned unchanged.",
		Params:  []tools.Param{pReq("text", "Quoted CSV field.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return csvUnescape(a.Str("text")), nil },
		Accepts: func(s string) bool { return len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' },
		Example: `encode.csv_unescape("\"say \"\"hi\"\", ok\"")`,
		Tags:    tagsWith("csv", "unescape", "unquote", "rfc4180", "field"),

		Keywords: kw("csv"),
	})
	reg(tools.Func{
		Name: "quoted_printable_encode", Title: "Quoted-printable encode",
		Doc:     "Encodes text as MIME quoted-printable (RFC 2045): non-ASCII bytes and '=' become =XX, trailing spaces are protected, lines are soft-wrapped with '=' before 76 columns and newlines become CRLF.",
		Params:  []tools.Param{pReq("text", "Text to encode."), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return qpEncode(b), nil
		}),
		Example:  `encode.quoted_printable_encode("café = good")`,
		Tags:     tagsWith("quoted-printable", "qp", "mime", "email", "rfc2045", "=C3=A9", "encode"),
		Keywords: kw("qp", "quoted-printable", "mime"),
	})
	reg(tools.Func{
		Name: "quoted_printable_decode", Title: "Quoted-printable decode",
		Doc:     "Decodes MIME quoted-printable: =XX bytes are decoded, soft line breaks (= at the end of a line) are removed. Malformed = sequences stay literal. Returns text if valid UTF-8, otherwise an error unless output is hex/base64.",
		Params:  []tools.Param{pReq("text", "Quoted-printable text."), pOut()},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			return toRep(qpDecode(a.Str("text")), a.Str("output"))
		},
		Accepts: func(s string) bool {
			return qpPattern.MatchString(s) && looksText(qpDecode(s))
		},
		Example:  `encode.quoted_printable_decode("caf=C3=A9 =3D good")`,
		Tags:     tagsWith("quoted-printable", "qp", "mime", "email", "rfc2045", "=C3=A9", "decode"),
		Keywords: kw("qp", "quoted-printable", "mime"),
	})
	reg(tools.Func{
		Name: "uuencode", Title: "UUencode",
		Doc:     "Encodes text in the classic Unix uuencode format: a \"begin MODE NAME\" header, lines of up to 45 bytes with a length prefix, a terminating backquote line and \"end\".",
		Params:  []tools.Param{pReq("text", "Text to encode."), pStr("filename", "File name in the header.", "data"), pStr("mode", "Octal permission string in the header.", "644"), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			name := strings.Fields(a.Str("filename") + " ")
			n := "data"
			if len(name) > 0 {
				n = name[0]
			}
			return uuEncode(b, n, a.Str("mode")), nil
		}),
		Example:  `encode.uuencode("Hello", filename="hello.txt")`,
		Tags:     tagsWith("uuencode", "uu", "unix", "begin 644", "encode"),
		Keywords: []string{"uu", "uuencode"},
	})
	reg(tools.Func{
		Name: "uudecode", Title: "UUdecode",
		Doc:     "Decodes uuencoded data. The \"begin\"/\"end\" lines are optional. Returns text if the bytes are valid UTF-8, otherwise an error unless output is hex/base64.",
		Params:  []tools.Param{pReq("text", "UUencoded text."), pOut()},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			b, err := uuDecode(a.Str("text"))
			if err != nil {
				return nil, err
			}
			return toRep(b, a.Str("output"))
		},
		Accepts: func(s string) bool {
			if !strings.HasPrefix(strings.TrimSpace(s), "begin ") {
				return false
			}
			b, err := uuDecode(s)
			return err == nil && len(b) > 0
		},
		Example:  `encode.uudecode(encode.uuencode("Hello"))`,
		Tags:     tagsWith("uuencode", "uudecode", "uu", "unix", "begin 644", "decode"),
		Keywords: []string{"uu", "uuencode"},
	})
	reg(tools.Func{
		Name: "unescape_auto", Title: "Unescape (auto-detect)",
		Doc:     "Applies every common unescaping that matches, in order: JS %uXXXX, backslash escapes (\\n \\xHH \\uXXXX octal), HTML entities, then %XX percent escapes (only if the result is valid UTF-8). Useful when the escaping type is unknown; text without any escapes is returned unchanged.",
		Params:  []tools.Param{pReq("text", "Escaped text.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return unescapeAuto(a.Str("text")), nil },
		Accepts: func(s string) bool { return unescapeAuto(s) != s },
		Example: `encode.unescape_auto("caf\\u00e9 &amp; %3Cb%3E")`,
		Tags:    tagsWith("unescape", "auto", "detect", "backslash", "html", "url", "percent", "decode"),

		Keywords: kw("unescape", "auto"),
	})
}

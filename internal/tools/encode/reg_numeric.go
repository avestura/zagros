package encode

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

func tokensToBytes(s string, base int) ([]byte, error) {
	switch base {
	case 2:
		return parseBinary(s)
	case 8, 10:
		return parseNumberTokens(s, base)
	case 16:
		return parseHex(s)
	}
	return nil, fmt.Errorf("base must be 2, 8, 10 or 16")
}

func numAccepts(base, minTokens int) func(string) bool {
	return func(s string) bool {
		if strings.TrimSpace(s) == "" {
			return false
		}
		if base == 10 || base == 8 {
			if len(strings.FieldsFunc(s, func(r rune) bool { return r < '0' || r > '9' })) < minTokens {
				return false
			}
		}
		b, err := tokensToBytes(s, base)
		return err == nil && len(b) >= 1 && looksText(b)
	}
}

func init() {
	reg(tools.Func{
		Name: "text_to_binary", Title: "Text to binary",
		Doc:     "Writes the UTF-8 bytes of the text as 8-bit binary groups (\"01001000 01101001\"). separator sits between groups; use text_to_bits for UTF-16/UTF-32 or other charsets.",
		Params:  []tools.Param{pReq("text", "Text to convert."), pStr("separator", "Text between 8-bit groups.", " "), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return joinNums(b, 2, a.Str("separator"), 8), nil
		}),
		Example:  `encode.text_to_binary("Hi")`,
		Tags:     tagsWith("binary", "bits", "01", "ascii", "text to binary", "bytes"),
		Keywords: []string{"binary", "bin", "bits"},
	})
	reg(tools.Func{
		Name: "binary_to_text", Title: "Binary to text",
		Doc:     "Reads binary digits (8-bit groups, optional spaces, commas or 0b prefixes; 7-bit groups when every group has 7 digits) and returns the text. Fails when the bit count is not a multiple of 8 or the bytes are not UTF-8.",
		Params:  []tools.Param{pReq("bits", "Binary digits, e.g. \"01001000 01101001\"."), pOut()},
		Returns: "str",
		Fn:      decodeFn("bits", func(a tools.Args, s string) ([]byte, error) { return parseBinary(s) }),
		Accepts: func(s string) bool {
			t := cleanWS(s)
			return len(t) >= 8 && numAccepts(2, 1)(s) && strings.Trim(strings.NewReplacer("0b", "", "0B", "", ",", "", "_", "").Replace(t), "01") == ""
		},
		Example:  `encode.binary_to_text("01001000 01101001")`,
		Tags:     tagsWith("binary", "bits", "01", "ascii", "binary to text", "bytes"),
		Keywords: []string{"binary", "bin", "bits"},
	})
	reg(tools.Func{
		Name: "text_to_octal", Title: "Text to octal bytes",
		Doc:     "Writes each UTF-8 byte of the text as a 3-digit octal number (\"110 151\"), separated by separator. Set prefix to \"\\\\\" for C-style escapes.",
		Params:  []tools.Param{pReq("text", "Text to convert."), pStr("separator", "Text between bytes.", " "), pStr("prefix", "Prefix before each byte.", ""), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			parts := make([]string, len(b))
			for i, c := range b {
				parts[i] = a.Str("prefix") + fmt.Sprintf("%03o", c)
			}
			return strings.Join(parts, a.Str("separator")), nil
		}),
		Example:  `encode.text_to_octal("Hi")`,
		Tags:     tagsWith("octal", "base8", "bytes", "text to octal"),
		Keywords: []string{"octal", "oct"},
	})
	reg(tools.Func{
		Name: "octal_to_text", Title: "Octal bytes to text",
		Doc:     "Reads octal byte values separated by spaces, commas or backslashes (\"110 151\", \"\\\\110\\\\151\") or a run of 3-digit groups, and returns the text. Values above 377 are rejected.",
		Params:  []tools.Param{pReq("text", "Octal bytes."), pOut()},
		Returns: "str",
		Fn:      decodeFn("text", func(a tools.Args, s string) ([]byte, error) { return parseNumberTokens(s, 8) }),
		Accepts: numAccepts(8, 2),
		Example: `encode.octal_to_text("110 151")`,
		Tags:    tagsWith("octal", "base8", "bytes", "octal to text"),

		Keywords: []string{"octal", "oct"},
	})
	reg(tools.Func{
		Name: "text_to_decimal_bytes", Title: "Text to decimal bytes",
		Doc:     "Writes each UTF-8 byte of the text as a decimal number 0-255 (\"72 101 108\"), joined by separator. Equivalent to the byte values of the string.",
		Params:  []tools.Param{pReq("text", "Text to convert."), pStr("separator", "Text between bytes.", " "), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return joinNums(b, 10, a.Str("separator"), 0), nil
		}),
		Example:  `encode.text_to_decimal_bytes("Hel")`,
		Tags:     tagsWith("decimal", "bytes", "ascii codes", "char codes", "text to ascii", "byte values"),
		Keywords: []string{"decimal", "dec", "ascii"},
	})
	reg(tools.Func{
		Name: "decimal_bytes_to_text", Title: "Decimal bytes to text",
		Doc:     "Reads decimal byte values separated by spaces, commas or other non-digits (\"72 101 108\") and returns the text. Values above 255 are rejected.",
		Params:  []tools.Param{pReq("text", "Decimal bytes."), pOut()},
		Returns: "str",
		Fn:      decodeFn("text", func(a tools.Args, s string) ([]byte, error) { return parseNumberTokens(s, 10) }),
		Accepts: numAccepts(10, 2),
		Example: `encode.decimal_bytes_to_text("72 101 108")`,
		Tags:    tagsWith("decimal", "bytes", "ascii codes", "char codes", "ascii to text", "byte values"),

		Keywords: []string{"decimal", "dec", "ascii"},
	})
	reg(tools.Func{
		Name: "text_to_bytes", Title: "Text to byte values (any base)",
		Doc:     "Writes the UTF-8 bytes of the text in base 2, 8, 10 or 16 separated by separator (binary and octal are zero-padded, hex is lower case without prefix).",
		Params:  []tools.Param{pReq("text", "Text to convert."), pInt("base", "Number base: 2, 8, 10 or 16.", 16), pStr("separator", "Text between bytes.", " "), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			base := int(a.Int("base"))
			switch base {
			case 2:
				return joinNums(b, 2, a.Str("separator"), 8), nil
			case 8:
				return joinNums(b, 8, a.Str("separator"), 3), nil
			case 10:
				return joinNums(b, 10, a.Str("separator"), 0), nil
			case 16:
				return hexJoin(b, false, a.Str("separator"), ""), nil
			}
			return "", fmt.Errorf("base must be 2, 8, 10 or 16")
		}),
		Example:  `encode.text_to_bytes("Hi", base=16)`,
		Tags:     tagsWith("bytes", "binary", "octal", "decimal", "hex", "base", "text to bytes"),
		Keywords: []string{"bytes"},
	})
	reg(tools.Func{
		Name: "bytes_to_text", Title: "Byte values to text (any base)",
		Doc:     "Reads byte values written in base 2, 8, 10 or 16 (separated by spaces, commas, etc.) and returns the text. Fails on out-of-range values or invalid UTF-8 unless output is hex/base64.",
		Params:  []tools.Param{pReq("data", "Byte values."), pInt("base", "Number base: 2, 8, 10 or 16.", 16), pOut()},
		Returns: "str",
		Fn: decodeFn("data", func(a tools.Args, s string) ([]byte, error) {
			return tokensToBytes(s, int(a.Int("base")))
		}),
		Accepts: func(s string) bool { return numAccepts(16, 1)(s) && len(strings.Fields(s)) >= 2 },
		Example: `encode.bytes_to_text("48 69", base=16)`,
		Tags:    tagsWith("bytes", "binary", "octal", "decimal", "hex", "base", "bytes to text"),

		Keywords: []string{"bytes"},
	})
	reg(tools.Func{
		Name: "text_to_bits", Title: "Text to bits (any charset)",
		Doc:     "Encodes the text in the given charset (utf-8 default; utf-16le, utf-16be, utf-32le, iso-8859-1, shift_jis, ... see charsets()) and writes every byte as 8 binary digits. Characters the charset cannot represent cause an error.",
		Params:  []tools.Param{pReq("text", "Text to convert."), pStr("charset", "Charset name from charsets().", "utf-8"), pStr("separator", "Text between bytes.", " ")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			b, err := csEncode(a.Str("text"), a.Str("charset"), false)
			if err != nil {
				return nil, err
			}
			return joinNums(b, 2, a.Str("separator"), 8), nil
		},
		Example:  `encode.text_to_bits("Hi", charset="utf-16le")`,
		Tags:     tagsWith("bits", "binary", "utf-16", "utf-32", "charset", "text to bits"),
		Keywords: []string{"bits", "binary"},
	})
	reg(tools.Func{
		Name: "bits_to_text", Title: "Bits to text (any charset)",
		Doc:     "Reads binary digits (groups of 8) and decodes the bytes in the given charset (utf-8 default; utf-16le, utf-32be, windows-1252, ...). Invalid sequences cause an error unless errors=\"replace\".",
		Params:  []tools.Param{pReq("bits", "Binary digits."), pStr("charset", "Charset name from charsets().", "utf-8"), pStr("errors", "\"strict\" or \"replace\" (use U+FFFD).", "strict", "strict", "replace")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			b, err := parseBinary(a.Str("bits"))
			if err != nil {
				return nil, err
			}
			return csDecode(b, a.Str("charset"), a.Str("errors") == "replace")
		},
		Accepts: func(s string) bool {
			t := cleanWS(s)
			return len(t) >= 8 && len(t)%8 == 0 && strings.Trim(t, "01") == "" && numAccepts(2, 1)(s)
		},
		Example:  `encode.bits_to_text("0100100000000000", charset="utf-16le")`,
		Tags:     tagsWith("bits", "binary", "utf-16", "utf-32", "charset", "bits to text"),
		Keywords: []string{"bits", "binary"},
	})
}

var _ = strconv.Itoa

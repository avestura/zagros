package encode

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
	"golang.org/x/text/unicode/norm"
)

func init() {
	regAPI(tools.Func{
		Name: "encode_charset", Title: "Encode text in a charset",
		Doc:     "Converts text to the bytes of a legacy or Unicode charset (utf-8, utf-16le/be, utf-32, iso-8859-x, windows-125x, koi8, ibm437..., shift_jis, euc-jp, gbk, gb18030, big5, euc-kr, ...; see charsets()). The bytes are returned as hex by default (also base64, percent, decimal, binary). Unrepresentable characters are an error unless errors=\"replace\".",
		Params:  []tools.Param{pReq("text", "Text to encode."), pReq("charset", "Charset name from charsets()."), pStr("output", "Byte representation of the result.", "hex", "hex", "base64", "percent", "decimal", "binary"), pStr("errors", "\"strict\" or \"replace\".", "strict", "strict", "replace")},
		Returns: "str", Input: "text",
		Fn: func(a tools.Args) (any, error) {
			b, err := csEncode(a.Str("text"), a.Str("charset"), a.Str("errors") == "replace")
			if err != nil {
				return nil, err
			}
			out := a.Str("output")
			if normName(out) == "text" {
				return nil, fmt.Errorf("output must be a byte form (hex, base64, percent, decimal, binary)")
			}
			return toRep(b, out)
		},
		Example: `encode.encode_charset("café", "windows-1252")`,
		Tags:    tagsWith("charset", "encoding", "codepage", "iso-8859-1", "windows-1252", "utf-16", "shift_jis", "gbk", "big5", "latin1", "text to bytes"),
	})
	regAPI(tools.Func{
		Name: "decode_charset", Title: "Decode bytes from a charset",
		Doc:     "Decodes bytes written as hex (default; also base64, percent, decimal, binary) using a charset from charsets() and returns the text. Invalid byte sequences are an error unless errors=\"replace\", which substitutes U+FFFD.",
		Params:  []tools.Param{pReq("data", "Bytes in the representation named by input."), pReq("charset", "Charset name from charsets()."), pStr("input", "Representation of data.", "hex", "hex", "base64", "percent", "decimal", "binary"), pStr("errors", "\"strict\" or \"replace\".", "strict", "strict", "replace")},
		Returns: "str", Input: "data",
		Fn: func(a tools.Args) (any, error) {
			b, err := fromRep(a.Str("data"), a.Str("input"))
			if err != nil {
				return nil, err
			}
			return csDecode(b, a.Str("charset"), a.Str("errors") == "replace")
		},
		Example: `encode.decode_charset("63 61 66 e9", "windows-1252")`,
		Tags:    tagsWith("charset", "encoding", "codepage", "iso-8859-1", "windows-1252", "utf-16", "shift_jis", "gbk", "big5", "latin1", "bytes to text", "decode"),
	})
	regAPI(tools.Func{
		Name: "charsets", Title: "List supported charsets",
		Doc:     "Lists the canonical charset names accepted by encode_charset, decode_charset, text_to_bits and bits_to_text. Common aliases (latin1, cp1252, sjis, utf8, ...) are accepted too, as are other WHATWG labels.",
		Returns: "list[str]",
		Fn:      func(a tools.Args) (any, error) { return charsetNames(), nil },
		Example: `encode.charsets()`,
		Tags:    tagsWith("charset", "encodings", "codepages", "list", "supported"),
	})
	reg(tools.Func{
		Name: "mojibake_fix", Title: "Fix mojibake (garbled text)",
		Doc:     "Repairs UTF-8 text that was wrongly decoded as windows-1252/latin-1 (\"cafÃ©\" becomes \"café\", \"â€™\" becomes an apostrophe), repeating for double-encoded text. Only runs of non-ASCII characters that form valid UTF-8 are changed; other text is untouched.",
		Params:  []tools.Param{pReq("text", "Garbled text.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return mojibakeFix(a.Str("text")), nil },
		Accepts: func(s string) bool { return mojibakeFix(s) != s },
		Example: `encode.mojibake_fix("cafÃ©")`,
		Tags:    tagsWith("mojibake", "garbled", "fix encoding", "double encoded", "utf-8", "latin1", "windows-1252", "charset", "repair", "ftfy"),

		Keywords: []string{"mojibake", "charset", "enc"},
	})
	reg(tools.Func{
		Name: "mojibake_try_all", Title: "Mojibake: try all repairs",
		Doc:     "Tries to repair garbled text assuming each common source charset (windows-125x, iso-8859-x, koi8, cp437/850/852/866, macintosh) and returns a dict of charset to repaired text, including double-encoding rounds (\"windows-1252 x2\"). Empty when nothing can be repaired.",
		Params:  []tools.Param{pReq("text", "Garbled text.")},
		Returns: "dict[str, str]",
		Fn:      func(a tools.Args) (any, error) { return mojibakeAll(a.Str("text")), nil },
		Accepts: func(s string) bool { return len(mojibakeAll(s)) > 0 },
		Example: `encode.mojibake_try_all("РџСЂРёРІРµС‚")`,
		Tags:    tagsWith("mojibake", "garbled", "fix encoding", "candidates", "charset", "cyrillic", "latin1", "repair"),

		Keywords: []string{"mojibake", "charset", "enc"},
	})
	reg(tools.Func{
		Name: "detect_charset", Title: "Detect charset of bytes",
		Doc:     "Guesses the charset of bytes given as hex (separators and 0x tolerated), best guess first: BOMs, plain ASCII, valid UTF-8, interleaved NULs (UTF-16 without BOM), strict multibyte CJK decodes and single-byte statistics. A heuristic, not a guarantee.",
		Params:  []tools.Param{pReq("data_hex", "Bytes as hex, e.g. \"63 61 66 c3 a9\".")},
		Returns: "list[str]", Input: "data_hex",
		Fn: func(a tools.Args) (any, error) {
			b, err := parseHex(a.Str("data_hex"))
			if err != nil {
				return nil, err
			}
			return detectCharset(b), nil
		},
		Accepts: func(s string) bool {
			b, err := parseHex(s)
			return err == nil && len(b) >= 2 && !utf8.Valid(b)
		},
		Example: `encode.detect_charset("63 61 66 c3 a9")`,
		Tags:    tagsWith("charset", "detect", "guess encoding", "bom", "utf-8", "utf-16", "chardet"),

		Keywords: []string{"charset", "enc"},
	})
	reg(tools.Func{
		Name: "normalize_utf8", Title: "Normalize Unicode (NFC/NFD/NFKC/NFKD)",
		Doc:     "Normalizes Unicode text to NFC (default), NFD, NFKC or NFKD, and replaces invalid UTF-8 bytes with U+FFFD. NFC composes (e + combining acute becomes é); NFKC/NFKD also fold compatibility characters such as ligatures and full-width forms.",
		Params:  []tools.Param{pReq("text", "Text to normalize."), pStr("form", "Normalization form.", "nfc", "nfc", "nfd", "nfkc", "nfkd")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			s := strings.ToValidUTF8(a.Str("text"), "�")
			switch normName(a.Str("form")) {
			case "", "nfc":
				return norm.NFC.String(s), nil
			case "nfd":
				return norm.NFD.String(s), nil
			case "nfkc":
				return norm.NFKC.String(s), nil
			case "nfkd":
				return norm.NFKD.String(s), nil
			}
			return nil, fmt.Errorf("form must be nfc, nfd, nfkc or nfkd")
		},
		Example:  `encode.normalize_utf8("é", form="nfc")`,
		Tags:     tagsWith("unicode", "normalize", "nfc", "nfd", "nfkc", "nfkd", "utf-8", "combining", "diacritics", "compose"),
		Keywords: []string{"normalize", "nfc", "enc"},
	})
}

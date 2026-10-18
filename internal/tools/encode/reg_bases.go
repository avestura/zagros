package encode

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
)

func tagsWith(base ...string) []string { return base }

// textDec wraps a byte decoder into an Fn that honours the output parameter.
func decodeFn(param string, dec func(a tools.Args, s string) ([]byte, error)) func(a tools.Args) (any, error) {
	return func(a tools.Args) (any, error) {
		b, err := dec(a, a.Str(param))
		if err != nil {
			return nil, err
		}
		return toRep(b, a.Str("output"))
	}
}

func encodeFn(param string, enc func(a tools.Args, b []byte) (string, error)) func(a tools.Args) (any, error) {
	return func(a tools.Args) (any, error) {
		b, err := fromRep(a.Str(param), a.Str("input_encoding"))
		if err != nil {
			return nil, err
		}
		return enc(a, b)
	}
}

func init() {
	// ---- base64 ----
	reg(tools.Func{
		Name: "base64_encode", Title: "Base64 encode", Order: 1,
		Doc:     "Encodes text as Base64 (RFC 4648). Variants: standard with padding (default), url (URL-safe alphabet), standard_nopad, url_nopad, mime (wrapped at 76 columns with CRLF); wrap adds line breaks every N characters. Text is encoded as UTF-8 bytes unless input_encoding is hex/base64/etc.",
		Params:  []tools.Param{pReq("text", "Text to encode."), pStr("variant", "Alphabet and padding.", "standard", "standard", "url", "standard_nopad", "url_nopad", "mime"), pInt("wrap", "Insert a line break every N characters (0 = no wrapping; mime always wraps at 76).", 0), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return b64Encode(b, a.Str("variant"), int(a.Int("wrap")))
		}),
		Example:  `encode.base64_encode("Hello, World!")`,
		Tags:     tagsWith("base64", "b64", "rfc4648", "btoa", "mime", "encode", "binary to text", "url-safe", "base64url"),
		Keywords: []string{"b64", "base64"},
	})
	reg(tools.Func{
		Name: "base64_decode", Title: "Base64 decode", Order: 2,
		Doc:     "Decodes Base64 text. Tolerant: accepts the standard and URL-safe alphabets, missing padding, embedded whitespace/line breaks and a data: URI prefix. Returns text when the bytes are valid UTF-8, otherwise an error unless output is hex/base64/etc.",
		Params:  []tools.Param{pReq("text", "Base64 text."), pOut()},
		Returns: "str",
		Fn:      decodeFn("text", func(a tools.Args, s string) ([]byte, error) { return parseBase64(s) }),
		Accepts: func(s string) bool {
			t := cleanWS(s)
			return len(t) >= 4 && acceptText(parseBase64)(s)
		},
		Example:  `encode.base64_decode("SGVsbG8sIFdvcmxkIQ==")`,
		Tags:     tagsWith("base64", "b64", "rfc4648", "atob", "decode", "url-safe", "base64url", "mime"),
		Keywords: []string{"b64", "base64"},
	})
	reg(tools.Func{
		Name: "base64url_encode", Title: "Base64url encode (no padding)",
		Doc:     "Encodes text as URL-safe Base64 without padding (RFC 4648 section 5), the form used in JWTs and URLs. Text is encoded as UTF-8 unless input_encoding says otherwise.",
		Params:  []tools.Param{pReq("text", "Text to encode."), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return base64.RawURLEncoding.EncodeToString(b), nil
		}),
		Example:  `encode.base64url_encode("subjects?_d=1")`,
		Tags:     tagsWith("base64url", "url-safe", "jwt", "websafe", "base64", "b64", "rfc4648"),
		Keywords: []string{"b64", "base64", "base64url"},
	})
	reg(tools.Func{
		Name: "base64url_decode", Title: "Base64url decode",
		Doc:     "Decodes URL-safe Base64 (with or without padding; the standard alphabet is accepted too). Returns text when valid UTF-8, otherwise an error unless output is hex/base64/etc.",
		Params:  []tools.Param{pReq("text", "Base64url text."), pOut()},
		Returns: "str",
		Fn:      decodeFn("text", func(a tools.Args, s string) ([]byte, error) { return parseBase64(s) }),
		Accepts: func(s string) bool {
			return strings.ContainsAny(s, "-_") && !strings.ContainsAny(s, "+/") && len(cleanWS(s)) >= 4 && acceptText(parseBase64)(s)
		},
		Example:  `encode.base64url_decode("c3ViamVjdHM_X2Q9MQ")`,
		Tags:     tagsWith("base64url", "url-safe", "jwt", "websafe", "base64", "b64", "decode"),
		Keywords: []string{"b64", "base64", "base64url"},
	})
	reg(tools.Func{
		Name: "hex_to_base64", Title: "Hex to Base64",
		Doc:     "Converts a hexadecimal byte string (separators, 0x and \\x prefixes tolerated) to Base64, without going through text. Fails on odd length or non-hex characters.",
		Params:  []tools.Param{pReq("hex", "Hex bytes, e.g. \"48656c6c6f\" or \"48 65 6c 6c 6f\"."), pStr("variant", "Base64 variant.", "standard", "standard", "url", "standard_nopad", "url_nopad", "mime")},
		Returns: "str", Input: "hex",
		Fn: func(a tools.Args) (any, error) {
			b, err := parseHex(a.Str("hex"))
			if err != nil {
				return nil, err
			}
			return b64Encode(b, a.Str("variant"), 0)
		},
		Accepts: func(s string) bool {
			b, err := parseHex(s)
			return err == nil && len(b) > 0
		},
		Example:  `encode.hex_to_base64("48656c6c6f")`,
		Tags:     tagsWith("hex", "base64", "convert", "b64", "bytes"),
		Keywords: []string{"b64", "base64", "hex"},
	})
	reg(tools.Func{
		Name: "base64_to_hex", Title: "Base64 to hex",
		Doc:     "Converts Base64 (standard or URL-safe, padding optional) to a hexadecimal byte string, without requiring the bytes to be text. Optional upper-case digits and a separator between bytes.",
		Params:  []tools.Param{pReq("text", "Base64 text."), pStr("case", "Hex digit case.", "lower", "lower", "upper"), pStr("separator", "Text placed between bytes (e.g. \":\" or \" \").", "")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			b, err := parseBase64(a.Str("text"))
			if err != nil {
				return nil, err
			}
			return hexJoin(b, a.Str("case") == "upper", a.Str("separator"), ""), nil
		},
		Accepts: func(s string) bool {
			b, err := parseBase64(s)
			return err == nil && len(b) > 0 && len(cleanWS(s)) >= 4
		},
		Example:  `encode.base64_to_hex("SGVsbG8=")`,
		Tags:     tagsWith("base64", "hex", "convert", "b64", "bytes"),
		Keywords: []string{"b64", "base64", "hex"},
	})

	// ---- hex ----
	reg(tools.Func{
		Name: "hex_encode", Title: "Hex encode", Order: 7,
		Doc:     "Encodes text as hexadecimal bytes (UTF-8 unless input_encoding says otherwise). case chooses lower/upper digits, separator goes between bytes (\":\", \" \", \"-\") and prefix before each byte (\"0x\", \"\\\\x\").",
		Params:  []tools.Param{pReq("text", "Text to encode."), pStr("case", "Hex digit case.", "lower", "lower", "upper"), pStr("separator", "Text placed between bytes, e.g. \":\", \" \" or \", \".", ""), pStr("prefix", "Text placed before each byte, e.g. \"0x\" or \"\\\\x\".", ""), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return hexJoin(b, a.Str("case") == "upper", a.Str("separator"), a.Str("prefix")), nil
		}),
		Example:  `encode.hex_encode("Hello", separator=" ")`,
		Tags:     tagsWith("hex", "base16", "hexadecimal", "bytes", "dump", "rfc4648", "text to hex"),
		Keywords: []string{"hex", "base16"},
	})
	reg(tools.Func{
		Name: "hex_decode", Title: "Hex decode", Order: 8,
		Doc:     "Decodes hexadecimal bytes to text. Tolerant of spaces, colons, dashes, commas and 0x, \\x or % prefixes. Fails on odd digit counts or non-hex characters; returns text if valid UTF-8, otherwise an error unless output is base64/etc.",
		Params:  []tools.Param{pReq("text", "Hex text, e.g. \"48 65 6c 6c 6f\" or \"0x48,0x65\"."), pOut()},
		Returns: "str",
		Fn:      decodeFn("text", func(a tools.Args, s string) ([]byte, error) { return parseHex(s) }),
		Accepts: func(s string) bool {
			t := cleanWS(s)
			if len(t) < 2 {
				return false
			}
			return acceptText(parseHex)(s)
		},
		Example:  `encode.hex_decode("48 65 6c 6c 6f")`,
		Tags:     tagsWith("hex", "base16", "hexadecimal", "bytes", "decode", "hex to text", "unhex"),
		Keywords: []string{"hex", "base16"},
	})
	reg(tools.Func{
		Name: "text_to_hex_bytes", Title: "Text to hex bytes (0x48 0x65 ...)",
		Doc:     "Writes the UTF-8 bytes of the text as hex bytes with a 0x prefix, separated by spaces by default (the common C/array literal style). Change prefix and separator for other styles.",
		Params:  []tools.Param{pReq("text", "Text to convert."), pStr("prefix", "Prefix of each byte.", "0x"), pStr("separator", "Separator between bytes.", " "), pStr("case", "Hex digit case.", "lower", "lower", "upper"), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return hexJoin(b, a.Str("case") == "upper", a.Str("separator"), a.Str("prefix")), nil
		}),
		Example:  `encode.text_to_hex_bytes("Hi", separator=", ")`,
		Tags:     tagsWith("hex", "bytes", "0x", "c array", "byte array", "text to hex"),
		Keywords: []string{"hex"},
	})
	reg(tools.Func{
		Name: "hex_bytes_to_text", Title: "Hex bytes to text",
		Doc:     "Reads hex bytes written in any common style (\"0x48, 0x65\", \"\\\\x48\\\\x65\", \"48:65\", \"4865\") and returns the text. Same rules as hex_decode.",
		Params:  []tools.Param{pReq("text", "Hex bytes."), pOut()},
		Returns: "str",
		Fn:      decodeFn("text", func(a tools.Args, s string) ([]byte, error) { return parseHex(s) }),
		Accepts: func(s string) bool {
			return (strings.Contains(s, "0x") || strings.Contains(s, "\\x")) && acceptText(parseHex)(s)
		},
		Example:  `encode.hex_bytes_to_text("0x48 0x65 0x6c 0x6c 0x6f")`,
		Tags:     tagsWith("hex", "bytes", "0x", "byte array", "hex to text", "c array"),
		Keywords: []string{"hex"},
	})

	// ---- base32 ----
	reg(tools.Func{
		Name: "base32_encode", Title: "Base32 encode",
		Doc:     "Encodes text as Base32. Variants: rfc4648 (default, A-Z2-7, padded), hex (extended hex alphabet), crockford (no padding, upper case), zbase32 (z-base-32, lower case, no padding). padding=false drops '=' for rfc4648/hex.",
		Params:  []tools.Param{pReq("text", "Text to encode."), pStr("variant", "Alphabet.", "rfc4648", "rfc4648", "hex", "crockford", "zbase32"), pBool("padding", "Add '=' padding (rfc4648 and hex only).", true), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return b32Encode(b, a.Str("variant"), a.Bool("padding"))
		}),
		Example:  `encode.base32_encode("Hello")`,
		Tags:     tagsWith("base32", "b32", "rfc4648", "crockford", "z-base-32", "totp", "encode"),
		Keywords: []string{"b32", "base32"},
	})
	reg(tools.Func{
		Name: "base32_decode", Title: "Base32 decode",
		Doc:     "Decodes Base32 in the chosen variant (rfc4648, hex, crockford, zbase32). Case-insensitive, ignores whitespace and padding; Crockford also ignores hyphens and maps I/L to 1 and O to 0. Returns text if valid UTF-8, otherwise an error unless output is hex/base64.",
		Params:  []tools.Param{pReq("text", "Base32 text."), pStr("variant", "Alphabet.", "rfc4648", "rfc4648", "hex", "crockford", "zbase32"), pOut()},
		Returns: "str",
		Fn: decodeFn("text", func(a tools.Args, s string) ([]byte, error) {
			return b32Decode(s, a.Str("variant"))
		}),
		Accepts: func(s string) bool {
			t := cleanWS(s)
			return len(t) >= 8 && t == strings.ToUpper(t) && acceptText(func(x string) ([]byte, error) { return b32Decode(x, "rfc4648") })(s)
		},
		Example:  `encode.base32_decode("JBSWY3DP")`,
		Tags:     tagsWith("base32", "b32", "rfc4648", "crockford", "z-base-32", "decode"),
		Keywords: []string{"b32", "base32"},
	})

	// ---- base36/45/58/62/85/91 ----
	reg(tools.Func{
		Name: "base36_encode", Title: "Base36 encode",
		Doc:     "Encodes the bytes of the text as one big number in base 36 (0-9a-z); leading zero bytes become leading '0' characters. Inputs above 32 KiB are refused. For plain integers use the num group.",
		Params:  []tools.Param{pReq("text", "Text to encode."), pBool("upper", "Use upper-case letters.", false), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return b36Encode(b, a.Bool("upper"))
		}),
		Example:  `encode.base36_encode("Hello")`,
		Tags:     tagsWith("base36", "b36", "alphanumeric", "encode"),
		Keywords: []string{"b36", "base36"},
	})
	reg(tools.Func{
		Name: "base36_decode", Title: "Base36 decode",
		Doc:     "Decodes base 36 text (case-insensitive) into the bytes of a big number, reversing base36_encode. Returns text if valid UTF-8, otherwise an error unless output is hex/base64.",
		Params:  []tools.Param{pReq("text", "Base36 text."), pOut()},
		Returns: "str",
		Fn:      decodeFn("text", func(a tools.Args, s string) ([]byte, error) { return b36Decode(s) }),
		Accepts: func(s string) bool {
			return len(cleanWS(s)) >= 3 && acceptText(b36Decode)(s)
		},
		Example:  `encode.base36_decode("3yud78mn")`,
		Tags:     tagsWith("base36", "b36", "decode"),
		Keywords: []string{"b36", "base36"},
	})
	reg(tools.Func{
		Name: "base45_encode", Title: "Base45 encode",
		Doc:     "Encodes text as Base45 (RFC 9285, used in EU digital COVID certificates and QR codes). Output uses upper-case letters, digits and ' $%*+-./:'.",
		Params:  []tools.Param{pReq("text", "Text to encode."), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return b45Encode(b), nil
		}),
		Example:  `encode.base45_encode("ietf!")`,
		Tags:     tagsWith("base45", "b45", "rfc9285", "qr", "covid", "encode"),
		Keywords: []string{"b45", "base45"},
	})
	reg(tools.Func{
		Name: "base45_decode", Title: "Base45 decode",
		Doc:     "Decodes Base45 (RFC 9285). Upper-case only; fails on characters outside the alphabet, invalid length or out-of-range groups. Returns text if valid UTF-8, otherwise an error unless output is hex/base64.",
		Params:  []tools.Param{pReq("text", "Base45 text."), pOut()},
		Returns: "str",
		Fn:      decodeFn("text", func(a tools.Args, s string) ([]byte, error) { return b45Decode(s) }),
		Accepts: func(s string) bool {
			return len(s) >= 3 && acceptText(b45Decode)(s)
		},
		Example:  `encode.base45_decode("QED8WEX0")`,
		Tags:     tagsWith("base45", "b45", "rfc9285", "qr", "decode"),
		Keywords: []string{"b45", "base45"},
	})
	reg(tools.Func{
		Name: "base58_encode", Title: "Base58 encode",
		Doc:     "Encodes the bytes of the text in Base58 (Bitcoin alphabet by default; flickr and ripple variants available). Leading zero bytes become leading '1' (first alphabet character). Inputs above 32 KiB are refused.",
		Params:  []tools.Param{pReq("text", "Text to encode."), pStr("variant", "Alphabet.", "bitcoin", "bitcoin", "flickr", "ripple"), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return b58Encode(b, a.Str("variant"))
		}),
		Example:  `encode.base58_encode("Hello World")`,
		Tags:     tagsWith("base58", "b58", "bitcoin", "flickr", "ripple", "encode"),
		Keywords: []string{"b58", "base58"},
	})
	reg(tools.Func{
		Name: "base58_decode", Title: "Base58 decode",
		Doc:     "Decodes Base58 (bitcoin, flickr or ripple alphabet). Fails on characters outside the alphabet (0, O, I, l in Bitcoin). Returns text if valid UTF-8, otherwise an error unless output is hex/base64.",
		Params:  []tools.Param{pReq("text", "Base58 text."), pStr("variant", "Alphabet.", "bitcoin", "bitcoin", "flickr", "ripple"), pOut()},
		Returns: "str",
		Fn: decodeFn("text", func(a tools.Args, s string) ([]byte, error) {
			return b58Decode(s, a.Str("variant"))
		}),
		Accepts: func(s string) bool {
			return len(cleanWS(s)) >= 4 && acceptText(func(x string) ([]byte, error) { return b58Decode(x, "bitcoin") })(s)
		},
		Example:  `encode.base58_decode("JxF12TrwUP45BMd")`,
		Tags:     tagsWith("base58", "b58", "bitcoin", "flickr", "ripple", "decode"),
		Keywords: []string{"b58", "base58"},
	})
	reg(tools.Func{
		Name: "base58check_encode", Title: "Base58Check encode",
		Doc:     "Base58Check: prepends a version byte, appends the first 4 bytes of the double SHA-256 checksum and encodes in Base58 (Bitcoin addresses, WIF keys). Default version 0 (P2PKH); payload is the text bytes or, with input_encoding=hex, raw bytes.",
		Params:  []tools.Param{pReq("text", "Payload."), pInt("version", "Version byte 0-255.", 0), pStr("variant", "Alphabet.", "bitcoin", "bitcoin", "flickr", "ripple"), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return b58CheckEncode(b, a.Int("version"), a.Str("variant"))
		}),
		Example:  `encode.base58check_encode("00112233445566778899aabbccddeeff00112233", input_encoding="hex")`,
		Tags:     tagsWith("base58check", "base58", "bitcoin", "address", "wif", "checksum", "b58"),
		Keywords: []string{"b58", "base58"},
	})
	reg(tools.Func{
		Name: "base58check_decode", Title: "Base58Check decode",
		Doc:     "Decodes Base58Check, verifies the 4-byte double SHA-256 checksum and returns the payload without the version byte (use base58check_parse to see the version). Fails with a checksum error on corrupted input; output hex/base64 for binary payloads.",
		Params:  []tools.Param{pReq("text", "Base58Check text."), pStr("variant", "Alphabet.", "bitcoin", "bitcoin", "flickr", "ripple"), pOut()},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			_, p, err := b58CheckDecode(a.Str("text"), a.Str("variant"))
			if err != nil {
				return nil, err
			}
			return toRep(p, a.Str("output"))
		},
		Accepts: func(s string) bool {
			_, _, err := b58CheckDecode(s, "bitcoin")
			return err == nil
		},
		Example:  `encode.base58check_decode("1QLbz7JHiBTspS962RLKV8GndWFwi5j6Qr", output="hex")`,
		Tags:     tagsWith("base58check", "base58", "bitcoin", "address", "wif", "checksum", "b58", "decode"),
		Keywords: []string{"b58", "base58"},
	})
	reg(tools.Func{
		Name: "base58check_parse", Title: "Base58Check parse",
		Doc:     "Parses a Base58Check string (Bitcoin address, WIF) into version, payload_hex, checksum and valid (\"true\"/\"false\"). Does not fail on a bad checksum, only on invalid characters or a too-short string.",
		Params:  []tools.Param{pReq("text", "Base58Check text."), pStr("variant", "Alphabet.", "bitcoin", "bitcoin", "flickr", "ripple")},
		Returns: "dict[str, str]",
		Fn: func(a tools.Args) (any, error) {
			raw, err := b58Decode(a.Str("text"), a.Str("variant"))
			if err != nil {
				return nil, err
			}
			if len(raw) < 5 {
				return nil, fmt.Errorf("too short for base58check")
			}
			body, sum := raw[:len(raw)-4], raw[len(raw)-4:]
			ok := string(doubleSHA(body)[:4]) == string(sum)
			return map[string]string{"version": fmt.Sprint(body[0]), "payload_hex": hex.EncodeToString(body[1:]), "checksum": hex.EncodeToString(sum), "valid": fmt.Sprint(ok)}, nil
		},
		Accepts: func(s string) bool {
			_, _, err := b58CheckDecode(s, "bitcoin")
			return err == nil
		},
		Example:  `encode.base58check_parse("1QLbz7JHiBTspS962RLKV8GndWFwi5j6Qr")`,
		Tags:     tagsWith("base58check", "bitcoin", "address", "version byte", "checksum", "wif"),
		Keywords: []string{"b58", "base58"},
	})
	reg(tools.Func{
		Name: "base62_encode", Title: "Base62 encode",
		Doc:     "Encodes the bytes of the text as one big number in base 62 (0-9A-Za-z by default; gmp = 0-9a-zA-Z, inverted = a-zA-Z0-9). Leading zero bytes become leading '0' characters. Inputs above 32 KiB are refused.",
		Params:  []tools.Param{pReq("text", "Text to encode."), pStr("variant", "Alphabet order.", "standard", "standard", "gmp", "inverted"), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			al, err := b62Alphabet(a.Str("variant"))
			if err != nil {
				return "", err
			}
			return bigEncode(b, al)
		}),
		Example:  `encode.base62_encode("Hello")`,
		Tags:     tagsWith("base62", "b62", "short url", "alphanumeric", "encode"),
		Keywords: []string{"b62", "base62"},
	})
	reg(tools.Func{
		Name: "base62_decode", Title: "Base62 decode",
		Doc:     "Decodes base 62 text (alphabet per variant) back to bytes, reversing base62_encode. Returns text if valid UTF-8, otherwise an error unless output is hex/base64.",
		Params:  []tools.Param{pReq("text", "Base62 text."), pStr("variant", "Alphabet order.", "standard", "standard", "gmp", "inverted"), pOut()},
		Returns: "str",
		Fn: decodeFn("text", func(a tools.Args, s string) ([]byte, error) {
			al, err := b62Alphabet(a.Str("variant"))
			if err != nil {
				return nil, err
			}
			return bigDecode(s, al, nil)
		}),
		Accepts: func(s string) bool {
			return len(cleanWS(s)) >= 3 && acceptText(func(x string) ([]byte, error) { return bigDecode(x, alphaB62, nil) })(s)
		},
		Example:  `encode.base62_decode("5TP3P3v")`,
		Tags:     tagsWith("base62", "b62", "decode"),
		Keywords: []string{"b62", "base62"},
	})
	reg(tools.Func{
		Name: "base85_encode", Title: "Base85 encode",
		Doc:     "Encodes text as Base85: ascii85 (Adobe/btoa style, 'z' for zero groups, optional <~ ~> delimiters), z85 (ZeroMQ) or rfc1924 (the IPv6 alphabet). z85 and rfc1924 handle a partial last group like ascii85 does.",
		Params:  []tools.Param{pReq("text", "Text to encode."), pStr("variant", "Alphabet.", "ascii85", "ascii85", "z85", "rfc1924"), pBool("delimiters", "Wrap ascii85 output in <~ and ~>.", false), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			v := normName(a.Str("variant"))
			if v == "" || v == "ascii85" || v == "btoa" {
				return a85Encode(b, a.Bool("delimiters")), nil
			}
			al, err := b85Alphabet(v)
			if err != nil {
				return "", err
			}
			return b85EncodeAlpha(b, al), nil
		}),
		Example:  `encode.base85_encode("Hello World", variant="ascii85")`,
		Tags:     tagsWith("base85", "ascii85", "b85", "a85", "btoa", "z85", "rfc1924", "adobe", "encode"),
		Keywords: []string{"b85", "base85", "ascii85"},
	})
	reg(tools.Func{
		Name: "base85_decode", Title: "Base85 decode",
		Doc:     "Decodes Base85 (ascii85 with optional <~ ~> delimiters and 'z' groups, z85, or rfc1924). Whitespace is ignored. Returns text if valid UTF-8, otherwise an error unless output is hex/base64.",
		Params:  []tools.Param{pReq("text", "Base85 text."), pStr("variant", "Alphabet.", "ascii85", "ascii85", "z85", "rfc1924"), pOut()},
		Returns: "str",
		Fn: decodeFn("text", func(a tools.Args, s string) ([]byte, error) {
			v := normName(a.Str("variant"))
			if v == "" || v == "ascii85" || v == "btoa" {
				return a85Decode(s)
			}
			al, err := b85Alphabet(v)
			if err != nil {
				return nil, err
			}
			return b85DecodeAlpha(s, al)
		}),
		Accepts: func(s string) bool {
			t := strings.TrimSpace(s)
			return len(t) >= 5 && (strings.HasPrefix(t, "<~") || len(cleanWS(t)) >= 8) && acceptText(a85Decode)(s)
		},
		Example:  `encode.base85_decode("87cURD]i,\"Ebo7")`,
		Tags:     tagsWith("base85", "ascii85", "b85", "a85", "btoa", "z85", "rfc1924", "adobe", "decode"),
		Keywords: []string{"b85", "base85", "ascii85"},
	})
	reg(tools.Func{
		Name: "base91_encode", Title: "Base91 encode",
		Doc:     "Encodes text with basE91 (Joachim Henke), about 23% overhead versus 33% for Base64. Uses 91 printable ASCII characters including quotes and brackets.",
		Params:  []tools.Param{pReq("text", "Text to encode."), pInEnc()},
		Returns: "str",
		Fn: encodeFn("text", func(a tools.Args, b []byte) (string, error) {
			return b91Encode(b), nil
		}),
		Example:  `encode.base91_encode("Hello World")`,
		Tags:     tagsWith("base91", "b91", "basE91", "encode"),
		Keywords: []string{"b91", "base91"},
	})
	reg(tools.Func{
		Name: "base91_decode", Title: "Base91 decode",
		Doc:     "Decodes basE91 text. Fails on characters outside the 91-character alphabet. Returns text if valid UTF-8, otherwise an error unless output is hex/base64.",
		Params:  []tools.Param{pReq("text", "Base91 text."), pOut()},
		Returns: "str",
		Fn:      decodeFn("text", func(a tools.Args, s string) ([]byte, error) { return b91Decode(s) }),
		Accepts: func(s string) bool {
			return len(cleanWS(s)) >= 4 && acceptText(b91Decode)(s)
		},
		Example:  `encode.base91_decode(encode.base91_encode("Hello World"))`,
		Tags:     tagsWith("base91", "b91", "basE91", "decode"),
		Keywords: []string{"b91", "base91"},
	})

	// ---- generic dispatchers ----
	regAPI(tools.Func{
		Name: "base_encode", Title: "Base-N encode (any algorithm)",
		Doc:     "Encodes text with the named base algorithm; see bases() for names (base64, base64url, base32, base32hex, base32_crockford, zbase32, base16, hex, base36, base45, base58, base62, ascii85, z85, base85_rfc1924, base91, ...). Aliases such as b64, b32, base85 work.",
		Params:  []tools.Param{pReq("algo", "Algorithm name from bases()."), pReq("text", "Text to encode."), pInEnc()},
		Returns: "str", Input: "text",
		Fn: func(a tools.Args) (any, error) {
			c, err := lookupCodec(a.Str("algo"))
			if err != nil {
				return nil, err
			}
			b, err := fromRep(a.Str("text"), a.Str("input_encoding"))
			if err != nil {
				return nil, err
			}
			return c.enc(b)
		},
		Example: `encode.base_encode("base58", "Hello")`,
		Tags:    tagsWith("base", "baseN", "base64", "base32", "base58", "base85", "dispatcher", "generic encode"),
	})
	regAPI(tools.Func{
		Name: "base_decode", Title: "Base-N decode (any algorithm)",
		Doc:     "Decodes text with the named base algorithm; see bases() for names. Returns text if the bytes are valid UTF-8, otherwise an error unless output is hex/base64.",
		Params:  []tools.Param{pReq("algo", "Algorithm name from bases()."), pReq("text", "Encoded text."), pOut()},
		Returns: "str", Input: "text",
		Fn: func(a tools.Args) (any, error) {
			c, err := lookupCodec(a.Str("algo"))
			if err != nil {
				return nil, err
			}
			b, err := c.dec(a.Str("text"))
			if err != nil {
				return nil, err
			}
			return toRep(b, a.Str("output"))
		},
		Example: `encode.base_decode("base58", "JxF12TrwUP45BMd")`,
		Tags:    tagsWith("base", "baseN", "base64", "base32", "base58", "base85", "dispatcher", "generic decode"),
	})
	regAPI(tools.Func{
		Name: "bases", Title: "List supported base algorithms",
		Doc:     "Lists the algorithm names accepted by base_encode and base_decode, sorted alphabetically.",
		Returns: "list[str]",
		Fn:      func(a tools.Args) (any, error) { return baseNames(), nil },
		Example: `encode.bases()`,
		Tags:    tagsWith("bases", "list", "algorithms", "base64", "base32", "base58", "supported"),
	})
}

func hexJoin(b []byte, upper bool, sep, prefix string) string {
	const lo, up = "0123456789abcdef", "0123456789ABCDEF"
	d := lo
	if upper {
		d = up
	}
	var sb strings.Builder
	for i, c := range b {
		if i > 0 {
			sb.WriteString(sep)
		}
		sb.WriteString(prefix)
		sb.WriteByte(d[c>>4])
		sb.WriteByte(d[c&15])
	}
	return sb.String()
}

var _ = utf8.Valid

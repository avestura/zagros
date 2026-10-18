package encode

import (
	"crypto/sha256"
	"encoding/ascii85"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"sort"
	"strings"
)

// ---- base64 ----

func cleanWS(s string) string {
	return strings.Map(func(r rune) rune {
		switch r {
		case ' ', '\t', '\r', '\n':
			return -1
		}
		return r
	}, s)
}

func b64Encode(b []byte, variant string, wrap int) (string, error) {
	var enc *base64.Encoding
	eol := "\n"
	switch normName(variant) {
	case "", "standard", "std":
		enc = base64.StdEncoding
	case "url", "urlsafe", "url-safe":
		enc = base64.URLEncoding
	case "standard_nopad", "std_nopad", "nopad", "raw":
		enc = base64.RawStdEncoding
	case "url_nopad", "urlsafe_nopad", "raw_url":
		enc = base64.RawURLEncoding
	case "mime":
		enc = base64.StdEncoding
		wrap = 76
		eol = "\r\n"
	default:
		return "", fmt.Errorf("unknown base64 variant %q (use standard, url, standard_nopad, url_nopad or mime)", variant)
	}
	s := enc.EncodeToString(b)
	if wrap > 0 && len(s) > wrap {
		var sb strings.Builder
		for i := 0; i < len(s); i += wrap {
			if i > 0 {
				sb.WriteString(eol)
			}
			j := i + wrap
			if j > len(s) {
				j = len(s)
			}
			sb.WriteString(s[i:j])
		}
		s = sb.String()
	}
	return s, nil
}

// ---- base32 ----

const (
	b32Std       = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	b32Hex       = "0123456789ABCDEFGHIJKLMNOPQRSTUV"
	b32Crockford = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"
	b32Z         = "ybndrfg8ejkmcpqxot1uwisza345h769"
)

func b32Alphabet(variant string) (alpha string, pads bool, err error) {
	switch normName(variant) {
	case "", "rfc4648", "standard", "std":
		return b32Std, true, nil
	case "hex", "extended-hex", "extended_hex", "base32hex":
		return b32Hex, true, nil
	case "crockford":
		return b32Crockford, false, nil
	case "zbase32", "z-base-32", "z", "zbase":
		return b32Z, false, nil
	}
	return "", false, fmt.Errorf("unknown base32 variant %q (use rfc4648, hex, crockford or zbase32)", variant)
}

func b32Encode(b []byte, variant string, padding bool) (string, error) {
	alpha, pads, err := b32Alphabet(variant)
	if err != nil {
		return "", err
	}
	pad := rune(base32.NoPadding)
	if pads && padding {
		pad = '='
	}
	return base32.NewEncoding(alpha).WithPadding(pad).EncodeToString(b), nil
}

func b32Decode(s, variant string) ([]byte, error) {
	alpha, _, err := b32Alphabet(variant)
	if err != nil {
		return nil, err
	}
	s = cleanWS(s)
	lower := alpha == b32Z
	if alpha == b32Crockford {
		s = strings.ReplaceAll(s, "-", "")
	}
	s = strings.TrimRight(s, "=")
	if lower {
		s = strings.ToLower(s)
	} else {
		s = strings.ToUpper(s)
	}
	if alpha == b32Crockford {
		s = strings.NewReplacer("I", "1", "L", "1", "O", "0").Replace(s)
	}
	for i := 0; i < len(s); i++ {
		if strings.IndexByte(alpha, s[i]) < 0 {
			return nil, fmt.Errorf("invalid base32 character %q", string(rune(s[i])))
		}
	}
	return base32.NewEncoding(alpha).WithPadding(base32.NoPadding).DecodeString(s)
}

// ---- big-number bases (36, 58, 62) ----

const (
	alphaB36      = "0123456789abcdefghijklmnopqrstuvwxyz"
	alphaB62      = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	alphaB62GMP   = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	alphaB62Inv   = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	alphaB58BTC   = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"
	alphaB58Flick = "123456789abcdefghijkmnopqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ"
	alphaB58Rip   = "rpshnaf39wBUDNEGHJKLM4PQRST7VWXYZ2bcdeCg65jkm8oFqi1tuvAxyz"
	bigLimit      = 32 << 10
)

func bigEncode(b []byte, alpha string) (string, error) {
	if len(b) > bigLimit {
		return "", fmt.Errorf("input is too large for big-number base encoding (max 32 KiB)")
	}
	zeros := 0
	for zeros < len(b) && b[zeros] == 0 {
		zeros++
	}
	n := new(big.Int).SetBytes(b)
	base := big.NewInt(int64(len(alpha)))
	mod := new(big.Int)
	var out []byte
	for n.Sign() > 0 {
		n.DivMod(n, base, mod)
		out = append(out, alpha[mod.Int64()])
	}
	for i := 0; i < zeros; i++ {
		out = append(out, alpha[0])
	}
	for i, j := 0, len(out)-1; i < j; i, j = i+1, j-1 {
		out[i], out[j] = out[j], out[i]
	}
	return string(out), nil
}

func bigDecode(s, alpha string, fold func(string) string) ([]byte, error) {
	s = cleanWS(s)
	if fold != nil {
		s = fold(s)
	}
	if len(s) > bigLimit*2 {
		return nil, fmt.Errorf("input is too large for big-number base decoding")
	}
	var idx [256]int
	for i := range idx {
		idx[i] = -1
	}
	for i := 0; i < len(alpha); i++ {
		idx[alpha[i]] = i
	}
	zeros := 0
	for zeros < len(s) && s[zeros] == alpha[0] {
		zeros++
	}
	n := new(big.Int)
	base := big.NewInt(int64(len(alpha)))
	d := new(big.Int)
	for i := 0; i < len(s); i++ {
		v := idx[s[i]]
		if v < 0 {
			return nil, fmt.Errorf("invalid character %q for this alphabet", string(rune(s[i])))
		}
		n.Mul(n, base)
		n.Add(n, d.SetInt64(int64(v)))
	}
	body := n.Bytes()
	out := make([]byte, zeros, zeros+len(body))
	return append(out, body...), nil
}

func b36Encode(b []byte, upper bool) (string, error) {
	s, err := bigEncode(b, alphaB36)
	if upper {
		s = strings.ToUpper(s)
	}
	return s, err
}

func b36Decode(s string) ([]byte, error) { return bigDecode(s, alphaB36, strings.ToLower) }

func b62Alphabet(variant string) (string, error) {
	switch normName(variant) {
	case "", "standard", "std":
		return alphaB62, nil
	case "gmp":
		return alphaB62GMP, nil
	case "inverted", "inv":
		return alphaB62Inv, nil
	}
	return "", fmt.Errorf("unknown base62 variant %q (use standard, gmp or inverted)", variant)
}

func b58Alphabet(variant string) (string, error) {
	switch normName(variant) {
	case "", "bitcoin", "btc":
		return alphaB58BTC, nil
	case "flickr":
		return alphaB58Flick, nil
	case "ripple":
		return alphaB58Rip, nil
	}
	return "", fmt.Errorf("unknown base58 variant %q (use bitcoin, flickr or ripple)", variant)
}

func b58Encode(b []byte, variant string) (string, error) {
	a, err := b58Alphabet(variant)
	if err != nil {
		return "", err
	}
	return bigEncode(b, a)
}

func b58Decode(s, variant string) ([]byte, error) {
	a, err := b58Alphabet(variant)
	if err != nil {
		return nil, err
	}
	return bigDecode(s, a, nil)
}

func doubleSHA(b []byte) []byte {
	h := sha256.Sum256(b)
	h = sha256.Sum256(h[:])
	return h[:]
}

func b58CheckEncode(b []byte, version int64, variant string) (string, error) {
	if version < 0 || version > 255 {
		return "", fmt.Errorf("version must be 0-255")
	}
	p := append([]byte{byte(version)}, b...)
	p = append(p, doubleSHA(p)[:4]...)
	return b58Encode(p, variant)
}

// b58CheckDecode returns version, payload and whether the checksum matches.
func b58CheckDecode(s, variant string) (byte, []byte, error) {
	raw, err := b58Decode(s, variant)
	if err != nil {
		return 0, nil, err
	}
	if len(raw) < 5 {
		return 0, nil, fmt.Errorf("too short for base58check (needs version byte and 4 checksum bytes)")
	}
	body, sum := raw[:len(raw)-4], raw[len(raw)-4:]
	want := doubleSHA(body)[:4]
	if string(want) != string(sum) {
		return 0, nil, fmt.Errorf("base58check checksum mismatch")
	}
	return body[0], body[1:], nil
}

// ---- base45 (RFC 9285) ----

const alphaB45 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ $%*+-./:"

func b45Encode(b []byte) string {
	var sb strings.Builder
	for i := 0; i+1 < len(b); i += 2 {
		n := int(b[i])<<8 | int(b[i+1])
		sb.WriteByte(alphaB45[n%45])
		sb.WriteByte(alphaB45[n/45%45])
		sb.WriteByte(alphaB45[n/45/45])
	}
	if len(b)%2 == 1 {
		n := int(b[len(b)-1])
		sb.WriteByte(alphaB45[n%45])
		sb.WriteByte(alphaB45[n/45])
	}
	return sb.String()
}

func b45Decode(s string) ([]byte, error) {
	s = strings.Trim(s, "\r\n")
	vals := make([]int, len(s))
	for i := 0; i < len(s); i++ {
		v := strings.IndexByte(alphaB45, s[i])
		if v < 0 {
			return nil, fmt.Errorf("invalid base45 character %q", string(rune(s[i])))
		}
		vals[i] = v
	}
	if len(s)%3 == 1 {
		return nil, fmt.Errorf("invalid base45 length")
	}
	var out []byte
	for i := 0; i < len(vals); i += 3 {
		if i+2 < len(vals) {
			n := vals[i] + vals[i+1]*45 + vals[i+2]*45*45
			if n > 0xFFFF {
				return nil, fmt.Errorf("base45 group %q is out of range", s[i:i+3])
			}
			out = append(out, byte(n>>8), byte(n))
		} else {
			n := vals[i] + vals[i+1]*45
			if n > 0xFF {
				return nil, fmt.Errorf("base45 group %q is out of range", s[i:i+2])
			}
			out = append(out, byte(n))
		}
	}
	return out, nil
}

// ---- base85 ----

const (
	alphaZ85     = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ.-:+=^!/*?&<>()[]{}@%$#"
	alphaRFC1924 = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz!#$%&()*+-;<=>?@^_`{|}~"
)

func b85Alphabet(variant string) (string, error) {
	switch normName(variant) {
	case "z85":
		return alphaZ85, nil
	case "rfc1924", "ipv6":
		return alphaRFC1924, nil
	}
	return "", fmt.Errorf("unknown base85 variant %q (use ascii85, z85 or rfc1924)", variant)
}

func b85EncodeAlpha(b []byte, alpha string) string {
	var sb strings.Builder
	for i := 0; i < len(b); i += 4 {
		var chunk [4]byte
		n := copy(chunk[:], b[i:])
		v := uint32(chunk[0])<<24 | uint32(chunk[1])<<16 | uint32(chunk[2])<<8 | uint32(chunk[3])
		var d [5]byte
		for j := 4; j >= 0; j-- {
			d[j] = alpha[v%85]
			v /= 85
		}
		sb.Write(d[:n+1])
	}
	return sb.String()
}

func b85DecodeAlpha(s, alpha string) ([]byte, error) {
	s = cleanWS(s)
	var idx [256]int
	for i := range idx {
		idx[i] = -1
	}
	for i := 0; i < len(alpha); i++ {
		idx[alpha[i]] = i
	}
	var out []byte
	for i := 0; i < len(s); i += 5 {
		end := i + 5
		if end > len(s) {
			end = len(s)
		}
		k := end - i
		if k == 1 {
			return nil, fmt.Errorf("invalid base85 length")
		}
		var v uint64
		for j := 0; j < 5; j++ {
			d := 84
			if j < k {
				d = idx[s[i+j]]
				if d < 0 {
					return nil, fmt.Errorf("invalid base85 character %q", string(rune(s[i+j])))
				}
			}
			v = v*85 + uint64(d)
		}
		if v > 0xFFFFFFFF {
			return nil, fmt.Errorf("base85 group out of range")
		}
		bytes4 := []byte{byte(v >> 24), byte(v >> 16), byte(v >> 8), byte(v)}
		out = append(out, bytes4[:k-1]...)
	}
	return out, nil
}

func a85Encode(b []byte, delim bool) string {
	dst := make([]byte, ascii85.MaxEncodedLen(len(b)))
	n := ascii85.Encode(dst, b)
	s := string(dst[:n])
	if delim {
		s = "<~" + s + "~>"
	}
	return s
}

func a85Decode(s string) ([]byte, error) {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "<~")
	s = strings.TrimSuffix(s, "~>")
	dst := make([]byte, len(s)*4+4)
	nd, _, err := ascii85.Decode(dst, []byte(s), true)
	if err != nil {
		return nil, err
	}
	return dst[:nd], nil
}

// ---- base91 (basE91 by Joachim Henke) ----

const alphaB91 = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!#$%&()*+,./:;<=>?@[]^_`{|}~\""

func b91Encode(data []byte) string {
	var out []byte
	var b, n uint
	for _, c := range data {
		b |= uint(c) << n
		n += 8
		if n > 13 {
			v := b & 8191
			if v > 88 {
				b >>= 13
				n -= 13
			} else {
				v = b & 16383
				b >>= 14
				n -= 14
			}
			out = append(out, alphaB91[v%91], alphaB91[v/91])
		}
	}
	if n > 0 {
		out = append(out, alphaB91[b%91])
		if n > 7 || b > 90 {
			out = append(out, alphaB91[b/91])
		}
	}
	return string(out)
}

func b91Decode(s string) ([]byte, error) {
	s = cleanWS(s)
	var out []byte
	v := -1
	var b, n uint
	for i := 0; i < len(s); i++ {
		d := strings.IndexByte(alphaB91, s[i])
		if d < 0 {
			return nil, fmt.Errorf("invalid base91 character %q", string(rune(s[i])))
		}
		if v < 0 {
			v = d
			continue
		}
		v += d * 91
		b |= uint(v) << n
		if v&8191 > 88 {
			n += 13
		} else {
			n += 14
		}
		for n > 7 {
			out = append(out, byte(b))
			b >>= 8
			n -= 8
		}
		v = -1
	}
	if v >= 0 {
		out = append(out, byte(b|uint(v)<<n))
	}
	return out, nil
}

// ---- codec table for base_encode / base_decode ----

type codec struct {
	name string
	enc  func([]byte) (string, error)
	dec  func(string) ([]byte, error)
}

var codecs = []codec{
	{"base64", func(b []byte) (string, error) { return b64Encode(b, "standard", 0) }, parseBase64},
	{"base64url", func(b []byte) (string, error) { return b64Encode(b, "url_nopad", 0) }, parseBase64},
	{"base64_nopad", func(b []byte) (string, error) { return b64Encode(b, "standard_nopad", 0) }, parseBase64},
	{"base64_mime", func(b []byte) (string, error) { return b64Encode(b, "mime", 0) }, parseBase64},
	{"base32", func(b []byte) (string, error) { return b32Encode(b, "rfc4648", true) }, func(s string) ([]byte, error) { return b32Decode(s, "rfc4648") }},
	{"base32hex", func(b []byte) (string, error) { return b32Encode(b, "hex", true) }, func(s string) ([]byte, error) { return b32Decode(s, "hex") }},
	{"base32_crockford", func(b []byte) (string, error) { return b32Encode(b, "crockford", false) }, func(s string) ([]byte, error) { return b32Decode(s, "crockford") }},
	{"zbase32", func(b []byte) (string, error) { return b32Encode(b, "zbase32", false) }, func(s string) ([]byte, error) { return b32Decode(s, "zbase32") }},
	{"base16", func(b []byte) (string, error) { return strings.ToUpper(hex.EncodeToString(b)), nil }, parseHex},
	{"hex", func(b []byte) (string, error) { return hex.EncodeToString(b), nil }, parseHex},
	{"base36", func(b []byte) (string, error) { return b36Encode(b, false) }, b36Decode},
	{"base45", func(b []byte) (string, error) { return b45Encode(b), nil }, b45Decode},
	{"base58", func(b []byte) (string, error) { return b58Encode(b, "bitcoin") }, func(s string) ([]byte, error) { return b58Decode(s, "bitcoin") }},
	{"base58_flickr", func(b []byte) (string, error) { return b58Encode(b, "flickr") }, func(s string) ([]byte, error) { return b58Decode(s, "flickr") }},
	{"base58_ripple", func(b []byte) (string, error) { return b58Encode(b, "ripple") }, func(s string) ([]byte, error) { return b58Decode(s, "ripple") }},
	{"base62", func(b []byte) (string, error) { return bigEncode(b, alphaB62) }, func(s string) ([]byte, error) { return bigDecode(s, alphaB62, nil) }},
	{"ascii85", func(b []byte) (string, error) { return a85Encode(b, false), nil }, a85Decode},
	{"z85", func(b []byte) (string, error) { return b85EncodeAlpha(b, alphaZ85), nil }, func(s string) ([]byte, error) { return b85DecodeAlpha(s, alphaZ85) }},
	{"base85_rfc1924", func(b []byte) (string, error) { return b85EncodeAlpha(b, alphaRFC1924), nil }, func(s string) ([]byte, error) { return b85DecodeAlpha(s, alphaRFC1924) }},
	{"base91", func(b []byte) (string, error) { return b91Encode(b), nil }, b91Decode},
}

var codecAliases = map[string]string{
	"b64": "base64", "base64std": "base64", "base64standard": "base64", "base64urlsafe": "base64url", "base64_url": "base64url",
	"b32": "base32", "base32rfc4648": "base32", "base32_hex": "base32hex", "crockford": "base32_crockford", "base32crockford": "base32_crockford",
	"z-base-32": "zbase32", "base32_z": "zbase32", "b16": "base16", "b36": "base36", "b45": "base45", "b58": "base58", "b62": "base62",
	"base85": "ascii85", "b85": "ascii85", "a85": "ascii85", "btoa": "ascii85", "rfc1924": "base85_rfc1924", "base85rfc1924": "base85_rfc1924",
	"b91": "base91", "mime": "base64_mime", "base64mime": "base64_mime",
}

func lookupCodec(name string) (codec, error) {
	n := strings.ReplaceAll(normName(name), "-", "_")
	n = strings.ReplaceAll(n, " ", "_")
	if a, ok := codecAliases[n]; ok {
		n = a
	}
	for _, c := range codecs {
		if c.name == n {
			return c, nil
		}
	}
	return codec{}, fmt.Errorf("unknown base algorithm %q; see encode.bases() for the supported names", name)
}

func baseNames() []string {
	out := make([]string, len(codecs))
	for i, c := range codecs {
		out[i] = c.name
	}
	sort.Strings(out)
	return out
}

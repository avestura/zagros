package cipherx

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

func xorBytes(data, key []byte) []byte {
	out := make([]byte, len(data))
	for i, b := range data {
		out[i] = b ^ key[i%len(key)]
	}
	return out
}

// englishFreq is the relative frequency (percent) of letters; space is the
// most common character of English text.
var englishFreq = map[byte]float64{
	' ': 13, 'e': 12.7, 't': 9.1, 'a': 8.2, 'o': 7.5, 'i': 7.0, 'n': 6.7, 's': 6.3, 'h': 6.1, 'r': 6.0,
	'd': 4.3, 'l': 4.0, 'c': 2.8, 'u': 2.8, 'm': 2.4, 'w': 2.4, 'f': 2.2, 'g': 2.0, 'y': 2.0, 'p': 1.9,
	'b': 1.5, 'v': 1.0, 'k': 0.8, 'j': 0.15, 'x': 0.15, 'q': 0.1, 'z': 0.07,
}

func englishScore(b []byte) float64 {
	var s float64
	for _, c := range b {
		lc := c
		if c >= 'A' && c <= 'Z' {
			lc = c + 32
		}
		if f, ok := englishFreq[lc]; ok {
			s += f
		} else if c >= 0x20 && c < 0x7f {
			s += 0.5 // other printable ASCII
		} else if c == '\n' || c == '\r' || c == '\t' {
			s += 0.5
		} else {
			s -= 20
		}
	}
	return s / float64(len(b))
}

func init() {
	reg(tools.Func{
		Name: "xor_cipher", Title: "XOR cipher",
		Doc: "XORs text with a repeating key and returns hex by default (output hex, base64, utf-8, ...). XOR is its own inverse: to decrypt, pass the hex result back with input_encoding=hex and output=utf-8, or use xor_decrypt. An empty key is an error. Not secure encryption.",
		Params: []tools.Param{
			reqStr("text", "Data to XOR (UTF-8 unless input_encoding says hex or base64)."),
			reqStr("key", "Repeating key (UTF-8 unless key_encoding says hex or base64); must not be empty."),
			inEncParam(), encParam("key_encoding", "the key"),
			outParam("hex", "Output format: hex (default), HEX, base64, base64url, base32 or utf-8 (raw bytes as text)."),
		},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			d, err := decodeBytes(a.Str("text"), a.Str("input_encoding"))
			if err != nil {
				return nil, err
			}
			k, err := decodeBytes(a.Str("key"), a.Str("key_encoding"))
			if err != nil {
				return nil, fmt.Errorf("key: %v", err)
			}
			if len(k) == 0 {
				return nil, errors.New("key must not be empty")
			}
			return encodeBytes(xorBytes(d, k), a.Str("output"))
		},
		Example: `cipher.xor_cipher("hello", "key")`,
		Tags:    []string{"xor", "exclusive or", "repeating key", "vernam", "one-time pad", "otp", "obfuscate", "cryptopals"},
		Family:  "cipher", Order: 2, Keywords: []string{"xor", "encrypt"},
	})
	reg(tools.Func{
		Name: "xor_decrypt", Title: "XOR decrypt",
		Doc: "Reverses xor_cipher: decodes data (hex by default), XORs it with the repeating key and returns UTF-8 text by default. An empty key is an error; a wrong key just gives different bytes since XOR has no integrity check.",
		Params: []tools.Param{
			reqStr("data", "XOR-ed data in data_encoding (hex by default)."),
			reqStr("key", "Repeating key (UTF-8 unless key_encoding says hex or base64)."),
			{Name: "data_encoding", Type: "str", Default: "hex", Choices: []string{"hex", "base64", "utf-8"}, Doc: "Encoding of data (default hex)."},
			encParam("key_encoding", "the key"),
			outParam("utf-8", "Output format: utf-8 text (default), hex, HEX, base64, base64url or base32."),
		},
		Returns: "str", Input: "data",
		Fn: func(a tools.Args) (any, error) {
			d, err := decodeBytes(a.Str("data"), a.Str("data_encoding"))
			if err != nil {
				return nil, err
			}
			k, err := decodeBytes(a.Str("key"), a.Str("key_encoding"))
			if err != nil {
				return nil, fmt.Errorf("key: %v", err)
			}
			if len(k) == 0 {
				return nil, errors.New("key must not be empty")
			}
			return encodeBytes(xorBytes(d, k), a.Str("output"))
		},
		Example: `cipher.xor_decrypt("03000a0c0e", "key")`,
		Tags:    []string{"xor", "decrypt", "decode", "repeating key"},
		Family:  "cipher", Order: 3, Keywords: []string{"xor", "decrypt"},
		Accepts: func(s string) bool {
			_, err := decodeBytes(s, "hex")
			return err == nil && len(strings.TrimSpace(s)) >= 2
		},
	})
	reg(tools.Func{
		Name: "xor_brute_single_byte", Title: "Break single-byte XOR",
		Doc: "Tries all 256 single-byte XOR keys on hex data and returns the best `limit` candidates (default 5) ranked by English letter-frequency score, each as '0xKK (c) score=S | plaintext'. Non-printable results are penalised. Works best on English text of 10+ characters.",
		Params: []tools.Param{
			reqStr("hex", "Ciphertext as hex (spaces, colons and a 0x prefix are ignored)."),
			{Name: "limit", Type: "int", Default: int64(5), Doc: "How many candidates to return, 1 to 256 (default 5)."},
		},
		Returns: "list[str]",
		Fn: func(a tools.Args) (any, error) {
			if err := clampInt(a.Int("limit"), 1, 256, "limit"); err != nil {
				return nil, err
			}
			d, err := decodeBytes(a.Str("hex"), "hex")
			if err != nil {
				return nil, err
			}
			if len(d) == 0 {
				return nil, errors.New("no data")
			}
			type cand struct {
				k     int
				score float64
				text  []byte
			}
			cs := make([]cand, 256)
			for k := 0; k < 256; k++ {
				p := xorBytes(d, []byte{byte(k)})
				cs[k] = cand{k, englishScore(p), p}
			}
			sort.SliceStable(cs, func(i, j int) bool { return cs[i].score > cs[j].score })
			var out []string
			for _, c := range cs[:a.Int("limit")] {
				shown := strings.Map(func(r rune) rune {
					if r < 0x20 || r == 0x7f {
						return '.'
					}
					return r
				}, string(c.text))
				ch := "."
				if c.k >= 0x20 && c.k < 0x7f {
					ch = string(rune(c.k))
				}
				out = append(out, fmt.Sprintf("0x%02x (%s) score=%.2f | %s", c.k, ch, c.score, shown))
			}
			return out, nil
		},
		Example: `cipher.xor_brute_single_byte("1b37373331363f78151b7f2b783431333d78397828372d363c78373e783a393b3736")`,
		Tags:    []string{"xor", "break", "crack", "brute force", "single byte", "cryptopals", "frequency analysis", "ctf"},
		Family:  "cipher", Order: 4, Keywords: []string{"xor", "brute"},
		Accepts: func(s string) bool { b, err := decodeBytes(s, "hex"); return err == nil && len(b) >= 4 },
	})
}

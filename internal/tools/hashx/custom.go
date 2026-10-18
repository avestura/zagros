package hashx

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/sha3"

	"github.com/avestura/zagros/internal/tools"
)

func seedParam() tools.Param {
	return tools.Param{Name: "seed", Type: "int", Default: int64(0), Doc: "Seed value (default 0). Truncated to the width of the algorithm."}
}

func lengthParam(def int64, max int64) tools.Param {
	return tools.Param{Name: "length", Type: "int", Default: def, Doc: fmt.Sprintf("Output length in bytes (1 to %d, default %d).", max, def)}
}

func variantParam(names []string, def string) tools.Param {
	sort.Strings(names)
	return tools.Param{Name: "variant", Type: "str", Default: def, Choices: names, Doc: "Parameter set: " + strings.Join(names, ", ") + ". Default " + def + "."}
}

func keys(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

func registerKeyedAndSeeded() {
	basic := func() []tools.Param { return []tools.Param{textParam(), inEncParam()} }
	const kname = "key"
	keyParams := func(what string) []tools.Param {
		return []tools.Param{
			{Name: kname, Type: "str", Default: "", Doc: what + " Empty means unkeyed."},
			encParam("key_encoding", "the key"),
		}
	}
	hdr := func(name, title, doc string, order int, params []tools.Param, fn func(tools.Args, []byte) ([]byte, error), tags ...string) {
		reg(tools.Func{
			Name: name, Title: title, Doc: doc,
			Params:  append(append(params[:0:0], params...), outParam()),
			Returns: "str",
			Fn: func(a tools.Args) (any, error) {
				b, err := inBytes(a)
				if err != nil {
					return nil, err
				}
				d, err := fn(a, b)
				if err != nil {
					return nil, err
				}
				return encodeBytes(d, a.Str("output"))
			},
			Example:  fmt.Sprintf(`hash.%s("hello")`, name),
			Tags:     append([]string{"hash", "digest", name}, tags...),
			Family:   "hash",
			Order:    orderOf(name, order),
			Keywords: keywordsFor(name),
		})
	}
	getKey := func(a tools.Args) ([]byte, error) {
		if a.Str(kname) == "" {
			return nil, nil
		}
		return decodeBytes(a.Str(kname), a.Str("key_encoding"))
	}

	// SHAKE
	for _, s := range []struct {
		name string
		n    func() sha3.ShakeHash
		def  int64
	}{{"shake128", sha3.NewShake128, 32}, {"shake256", sha3.NewShake256, 64}} {
		s := s
		hdr(s.name, strings.ToUpper(s.name), fmt.Sprintf("Computes the %s extendable-output function (SHA-3 family) of text and returns length bytes (default %d, max 65536) encoded per output. Text is UTF-8 unless input_encoding is hex or base64.", strings.ToUpper(s.name), s.def),
			300, append(basic(), lengthParam(s.def, 65536)),
			func(a tools.Args, b []byte) ([]byte, error) {
				n := a.Int("length")
				if err := clampInt(n, 1, 65536, "length"); err != nil {
					return nil, err
				}
				h := s.n()
				h.Write(b)
				out := make([]byte, n)
				h.Read(out)
				return out, nil
			}, "xof", "sha3", "extendable output")
	}

	// BLAKE2b / BLAKE2s with optional key
	for _, s := range []struct {
		name string
		size int
	}{{"blake2b_256", 32}, {"blake2b_384", 48}, {"blake2b_512", 64}} {
		s := s
		hdr(s.name, fmt.Sprintf("BLAKE2b-%d", s.size*8), fmt.Sprintf("Computes the BLAKE2b-%d digest of text, optionally keyed (key up to 64 bytes, a MAC when set; key_encoding utf-8, hex or base64). Text is UTF-8 unless input_encoding is hex or base64; output is hex by default.", s.size*8),
			110, append(basic(), keyParams("Optional MAC key, up to 64 bytes.")...),
			func(a tools.Args, b []byte) ([]byte, error) {
				k, err := getKey(a)
				if err != nil {
					return nil, err
				}
				h, err := blake2b.New(s.size, k)
				if err != nil {
					return nil, err
				}
				h.Write(b)
				return h.Sum(nil), nil
			}, "blake2", "rfc7693", "keyed", "mac")
	}
	hdr("blake2b", "BLAKE2b (any size)", "Computes BLAKE2b with a digest size of 1 to 64 bytes (default 64) and an optional key of up to 64 bytes. Text is UTF-8 unless input_encoding is hex or base64.",
		310, append(append(basic(), lengthParam(64, 64)), keyParams("Optional MAC key, up to 64 bytes.")...),
		func(a tools.Args, b []byte) ([]byte, error) {
			if err := clampInt(a.Int("length"), 1, 64, "length"); err != nil {
				return nil, err
			}
			k, err := getKey(a)
			if err != nil {
				return nil, err
			}
			h, err := blake2b.New(int(a.Int("length")), k)
			if err != nil {
				return nil, err
			}
			h.Write(b)
			return h.Sum(nil), nil
		}, "blake2", "keyed", "mac")
	hdr("blake2s_256", "BLAKE2s-256", "Computes the BLAKE2s-256 digest of text, optionally keyed (key up to 32 bytes). Text is UTF-8 unless input_encoding is hex or base64.",
		111, append(basic(), keyParams("Optional MAC key, up to 32 bytes.")...),
		func(a tools.Args, b []byte) ([]byte, error) {
			k, err := getKey(a)
			if err != nil {
				return nil, err
			}
			h, err := blake2s.New256(k)
			if err != nil {
				return nil, err
			}
			h.Write(b)
			return h.Sum(nil), nil
		}, "blake2", "rfc7693", "keyed", "mac")

	// BLAKE3
	hdr("blake3", "BLAKE3", "Computes the BLAKE3 digest of text (32 bytes by default, any length up to 65536 via length). With a 32-byte key it is a keyed MAC; with context it derives a key from text as key material (key and context are mutually exclusive). Text is UTF-8 unless input_encoding is hex or base64.",
		112, append(append(basic(), lengthParam(32, 65536)),
			tools.Param{Name: "key", Type: "str", Default: "", Doc: "Optional 32-byte key for keyed hashing. Empty means plain hashing."},
			encParam("key_encoding", "the key"),
			tools.Param{Name: "context", Type: "str", Default: "", Doc: "Optional context string; when set, derives a key from text (derive_key mode)."}),
		func(a tools.Args, b []byte) ([]byte, error) {
			n := a.Int("length")
			if err := clampInt(n, 1, 65536, "length"); err != nil {
				return nil, err
			}
			k, err := getKey(a)
			if err != nil {
				return nil, err
			}
			ctx := a.Str("context")
			switch {
			case len(k) > 0 && ctx != "":
				return nil, fmt.Errorf("blake3: use either key or context, not both")
			case len(k) > 0:
				if len(k) != 32 {
					return nil, fmt.Errorf("blake3: key must be exactly 32 bytes (got %d)", len(k))
				}
				return blake3Keyed(k, b, int(n)), nil
			case ctx != "":
				return blake3Derive(ctx, b, int(n)), nil
			}
			return blake3Hash(b, int(n)), nil
		}, "blake", "xof", "keyed", "derive key", "fast")

	// Seeded non-cryptographic hashes
	seeded := func(name, title string, tags []string, f func(b []byte, seed int64) []byte) {
		hdr(name, title, fmt.Sprintf("Computes %s of text with an optional integer seed (default 0). Non-cryptographic: for hash tables and sharding, not security. Text is UTF-8 unless input_encoding is hex or base64; output is hex by default (big-endian).", title),
			200, append(basic(), seedParam()),
			func(a tools.Args, b []byte) ([]byte, error) { return f(b, a.Int("seed")), nil },
			append([]string{"non-cryptographic", "seed"}, tags...)...)
	}
	seeded("murmur3_32", "MurmurHash3 (x86, 32-bit)", []string{"murmur", "murmurhash", "mmh3"},
		func(b []byte, s int64) []byte { return be32(murmur3x86_32(b, uint32(s))) })
	seeded("murmur3_128", "MurmurHash3 (x64, 128-bit)", []string{"murmur", "murmurhash", "mmh3"},
		func(b []byte, s int64) []byte {
			h1, h2 := murmur3x64_128(b, uint32(s))
			return append(be64(h1), be64(h2)...)
		})
	seeded("xxhash32", "xxHash32", []string{"xxh32", "xxhash"},
		func(b []byte, s int64) []byte { return be32(xxh32(b, uint32(s))) })
	seeded("xxhash64", "xxHash64", []string{"xxh64", "xxhash"},
		func(b []byte, s int64) []byte { return be64(xxh64(b, uint64(s))) })

	hdr("siphash", "SipHash-2-4", "Computes SipHash-2-4, a keyed 64-bit hash for hash tables. The key must be 16 bytes (empty means 16 zero bytes); the 8-byte result is output in the reference little-endian byte order. Text is UTF-8 unless input_encoding is hex or base64.",
		210, append(basic(), keyParams("16-byte key.")...),
		func(a tools.Args, b []byte) ([]byte, error) {
			k, err := getKey(a)
			if err != nil {
				return nil, err
			}
			var key [16]byte
			if len(k) != 0 && len(k) != 16 {
				return nil, fmt.Errorf("siphash: key must be 16 bytes (got %d)", len(k))
			}
			copy(key[:], k)
			v := siphash24(key, b)
			out := make([]byte, 8)
			for i := range out {
				out[i] = byte(v >> (8 * uint(i)))
			}
			return out, nil
		}, "siphash24", "keyed", "prf", "hash table")
}

func registerCRCGeneric() {
	for _, c := range []struct {
		name     string
		variants map[string]string
		def      string
		order    int
		doc      string
	}{
		{"crc16", crc16Variants, "arc", 205, "CRC-16"},
		{"crc32", crc32Variants, "ieee", 7, "CRC-32"},
		{"crc64", crc64Variants, "ecma", 206, "CRC-64"},
	} {
		c := c
		reg(tools.Func{
			Name: c.name, Title: c.doc,
			Doc:     fmt.Sprintf("Computes a %s checksum of text; variant picks the parameter set (default %s; the catalogue check value of each variant over \"123456789\" is verified). Text is UTF-8 unless input_encoding is hex or base64; output is fixed-width hex by default, or dec for the plain number. Non-cryptographic.", c.doc, c.def),
			Params:  []tools.Param{textParam(), variantParam(keys(c.variants), c.def), inEncParam(), outParam()},
			Returns: "str",
			Fn: func(a tools.Args) (any, error) {
				b, err := inBytes(a)
				if err != nil {
					return nil, err
				}
				v := strings.ToLower(strings.ReplaceAll(strings.ReplaceAll(a.Str("variant"), "-", "_"), " ", "_"))
				full, ok := c.variants[v]
				if !ok {
					return nil, fmt.Errorf("%s: unknown variant %q", c.name, a.Str("variant"))
				}
				return encodeBytes(crcCatalog[full].bytes(b), a.Str("output"))
			},
			Example:  fmt.Sprintf(`hash.%s("hello")`, c.name),
			Tags:     []string{"hash", "checksum", "crc", "cyclic redundancy check", "error detection", c.name, "zip", "ethernet"},
			Family:   "hash",
			Order:    c.order,
			Keywords: keywordsFor(c.name),
		})
	}
	reg(tools.Func{
		Name: "crc_custom", Title: "Custom CRC",
		Doc: "Computes a CRC with explicit Rocksoft-model parameters: width in bits (8 to 64), polynomial, initial value and final xor as hex strings, and the input/output reflection flags. Useful for protocols not covered by the named variants; the result is fixed-width hex by default.",
		Params: []tools.Param{
			textParam(),
			{Name: "width", Type: "int", Required: true, Doc: "CRC width in bits, 8 to 64."},
			{Name: "poly", Type: "str", Required: true, Doc: "Polynomial in normal (MSB-first) form as hex, e.g. 04C11DB7."},
			{Name: "init", Type: "str", Default: "0", Doc: "Initial register value as hex (default 0)."},
			{Name: "xorout", Type: "str", Default: "0", Doc: "Final xor value as hex (default 0)."},
			{Name: "reflect_in", Type: "bool", Default: false, Doc: "Reflect each input byte (LSB first)."},
			{Name: "reflect_out", Type: "bool", Default: false, Doc: "Reflect the final register."},
			inEncParam(), outParam(),
		},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			if err := clampInt(a.Int("width"), 8, 64, "width"); err != nil {
				return nil, err
			}
			hx := func(name string) (uint64, error) {
				s := strings.TrimPrefix(strings.TrimPrefix(strings.TrimSpace(a.Str(name)), "0x"), "0X")
				if s == "" {
					return 0, nil
				}
				v, err := strconv.ParseUint(s, 16, 64)
				if err != nil {
					return 0, fmt.Errorf("%s must be a hex number: %v", name, err)
				}
				return v, nil
			}
			p, err := hx("poly")
			if err != nil {
				return nil, err
			}
			i, err := hx("init")
			if err != nil {
				return nil, err
			}
			x, err := hx("xorout")
			if err != nil {
				return nil, err
			}
			b, err := inBytes(a)
			if err != nil {
				return nil, err
			}
			spec := crcSpec{uint(a.Int("width")), p, i, a.Bool("reflect_in"), a.Bool("reflect_out"), x}
			return encodeBytes(spec.bytes(b), a.Str("output"))
		},
		Example: `hash.crc_custom("123456789", 16, "1021", init="FFFF")`,
		Tags:    []string{"crc", "custom", "polynomial", "rocksoft", "checksum", "reveng"},
	})
}

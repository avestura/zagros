package hashx

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"hash/adler32"
	"hash/fnv"
	"sort"
	"strings"

	"golang.org/x/crypto/blake2b"
	"golang.org/x/crypto/blake2s"
	"golang.org/x/crypto/md4"
	"golang.org/x/crypto/ripemd160"
	"golang.org/x/crypto/sha3"

	"github.com/avestura/zagros/internal/tools"
)

// algo is one entry of the digest table. newH is set for algorithms that can
// be used in HMAC; sum is used for the rest. custom entries are registered by
// hand (they have extra parameters) but still work through hash().
type algo struct {
	name   string
	title  string
	tags   []string
	newH   func() hash.Hash
	sum    func([]byte) []byte
	custom bool
}

func (a algo) compute(b []byte) []byte {
	if a.sum != nil {
		return a.sum(b)
	}
	h := a.newH()
	h.Write(b)
	return h.Sum(nil)
}

func (a algo) bits() int { return len(a.compute(nil)) * 8 }

func mustB2b(size int) func() hash.Hash {
	return func() hash.Hash { h, _ := blake2b.New(size, nil); return h }
}

func shakeSum(n int, s func() sha3.ShakeHash) func([]byte) []byte {
	return func(b []byte) []byte {
		h := s()
		h.Write(b)
		out := make([]byte, n)
		h.Read(out)
		return out
	}
}

func crcAlgo(name, title string) algo {
	spec := crcCatalog[name]
	return algo{name: name, title: title, sum: spec.bytes, tags: []string{"crc", "checksum", "cyclic redundancy check"}}
}

var algos []algo
var algoIndex = map[string]algo{}

func buildAlgos() []algo {
	std := func(name, title string, n func() hash.Hash, tags ...string) algo {
		return algo{name: name, title: title, newH: n, tags: tags}
	}
	list := []algo{
		std("md4", "MD4", md4.New, "legacy", "insecure", "rfc1320"),
		std("md5", "MD5", md5.New, "legacy", "insecure", "rfc1321", "checksum", "fingerprint"),
		std("sha1", "SHA-1", sha1.New, "legacy", "sha-1", "rfc3174", "git"),
		std("sha224", "SHA-224", sha256.New224, "sha2", "sha-224"),
		std("sha256", "SHA-256", sha256.New, "sha2", "sha-256", "bitcoin", "fips180"),
		std("sha384", "SHA-384", sha512.New384, "sha2", "sha-384"),
		std("sha512", "SHA-512", sha512.New, "sha2", "sha-512"),
		std("sha512_224", "SHA-512/224", sha512.New512_224, "sha2", "sha-512/224"),
		std("sha512_256", "SHA-512/256", sha512.New512_256, "sha2", "sha-512/256"),
		{name: "sha256d", title: "Double SHA-256", tags: []string{"bitcoin", "sha256", "sha256 twice", "block hash", "txid"},
			sum: func(b []byte) []byte { x := sha256.Sum256(b); y := sha256.Sum256(x[:]); return y[:] }},
		{name: "hash160", title: "HASH160 (RIPEMD160 of SHA-256)", tags: []string{"bitcoin", "address", "ripemd160", "sha256"},
			sum: func(b []byte) []byte {
				x := sha256.Sum256(b)
				h := ripemd160.New()
				h.Write(x[:])
				return h.Sum(nil)
			}},
		std("sha3_224", "SHA3-224", sha3.New224, "sha3", "keccak", "fips202"),
		std("sha3_256", "SHA3-256", sha3.New256, "sha3", "keccak", "fips202"),
		std("sha3_384", "SHA3-384", sha3.New384, "sha3", "keccak", "fips202"),
		std("sha3_512", "SHA3-512", sha3.New512, "sha3", "keccak", "fips202"),
		std("keccak256", "Keccak-256", sha3.NewLegacyKeccak256, "ethereum", "keccak", "legacy sha3", "solidity", "evm"),
		std("keccak512", "Keccak-512", sha3.NewLegacyKeccak512, "keccak", "legacy sha3"),
		{name: "shake128", title: "SHAKE128", custom: true, tags: []string{"xof", "sha3", "extendable output"}, sum: shakeSum(32, sha3.NewShake128)},
		{name: "shake256", title: "SHAKE256", custom: true, tags: []string{"xof", "sha3", "extendable output"}, sum: shakeSum(64, sha3.NewShake256)},
		{name: "blake2b_256", title: "BLAKE2b-256", custom: true, newH: mustB2b(32), tags: []string{"blake2", "rfc7693", "keyed"}},
		{name: "blake2b_384", title: "BLAKE2b-384", custom: true, newH: mustB2b(48), tags: []string{"blake2", "rfc7693", "keyed"}},
		{name: "blake2b_512", title: "BLAKE2b-512", custom: true, newH: mustB2b(64), tags: []string{"blake2", "rfc7693", "keyed"}},
		{name: "blake2s_256", title: "BLAKE2s-256", custom: true, tags: []string{"blake2", "rfc7693", "keyed"},
			newH: func() hash.Hash { h, _ := blake2s.New256(nil); return h }},
		{name: "blake3", title: "BLAKE3", custom: true, tags: []string{"blake", "xof", "keyed", "derive key"}, sum: func(b []byte) []byte { return blake3Hash(b, 32) }},
		std("ripemd160", "RIPEMD-160", ripemd160.New, "rmd160", "bitcoin"),
		{name: "sm3", title: "SM3", newH: newSM3, tags: []string{"chinese", "gm/t 0004", "gb/t 32905", "shangmi"}},
		{name: "whirlpool", title: "Whirlpool", newH: newWhirlpool, tags: []string{"iso 10118-3", "512-bit", "veracrypt"}},
		{name: "ntlm", title: "NTLM (NT hash)", sum: ntlm, tags: []string{"windows", "nt hash", "md4", "utf-16le", "password", "legacy"}},
		{name: "lm", title: "LM hash (LAN Manager)", sum: lmHash, tags: []string{"windows", "lanman", "legacy", "insecure", "password", "des"}},
		{name: "mysql41", title: "MySQL PASSWORD() hash", sum: mysql41, tags: []string{"mysql", "mysql_native_password", "sha1 twice", "password"}},
	}
	for _, c := range []struct{ name, title string }{
		{"crc8", "CRC-8 (SMBus)"}, {"crc8_maxim", "CRC-8/MAXIM (Dallas 1-Wire)"},
		{"crc16_arc", "CRC-16/ARC"}, {"crc16_ccitt_false", "CRC-16/CCITT-FALSE"}, {"crc16_xmodem", "CRC-16/XMODEM"},
		{"crc16_kermit", "CRC-16/KERMIT"}, {"crc16_modbus", "CRC-16/MODBUS"}, {"crc16_x25", "CRC-16/X-25 (IBM-SDLC)"},
		{"crc16_usb", "CRC-16/USB"}, {"crc16_genibus", "CRC-16/GENIBUS"}, {"crc16_mcrf4xx", "CRC-16/MCRF4XX"},
		{"crc16_dnp", "CRC-16/DNP"}, {"crc16_buypass", "CRC-16/BUYPASS"}, {"crc16_t10_dif", "CRC-16/T10-DIF"},
		{"crc16_cdma2000", "CRC-16/CDMA2000"}, {"crc16_aug_ccitt", "CRC-16/AUG-CCITT"},
		{"crc24", "CRC-24 (OpenPGP)"},
		{"crc32_castagnoli", "CRC-32C (Castagnoli)"}, {"crc32_koopman", "CRC-32K (Koopman)"},
		{"crc32_bzip2", "CRC-32/BZIP2"}, {"crc32_mpeg2", "CRC-32/MPEG-2"}, {"crc32_posix", "CRC-32/POSIX (cksum)"},
		{"crc32_jamcrc", "CRC-32/JAMCRC"}, {"crc32_q", "CRC-32Q"}, {"crc32_d", "CRC-32D"},
		{"crc64_iso", "CRC-64/ISO (GO-ISO)"}, {"crc64_ecma", "CRC-64/ECMA (XZ)"}, {"crc64_ecma182", "CRC-64/ECMA-182"}, {"crc64_we", "CRC-64/WE"},
	} {
		list = append(list, crcAlgo(c.name, c.title))
	}
	crc32 := crcAlgo("crc32_ieee", "CRC-32 (IEEE)")
	crc32.name, crc32.custom = "crc32", true
	crc16 := crcAlgo("crc16_arc", "CRC-16")
	crc16.name, crc16.custom = "crc16", true
	crc64 := crcAlgo("crc64_ecma", "CRC-64")
	crc64.name, crc64.custom = "crc64", true
	crc32c := crcAlgo("crc32_castagnoli", "CRC-32C (Castagnoli)")
	crc32c.name = "crc32c"
	list = append(list, crc32, crc16, crc64, crc32c)
	list = append(list,
		algo{name: "adler32", title: "Adler-32", tags: []string{"checksum", "zlib", "rfc1950"}, sum: func(b []byte) []byte { return be32(adler32.Checksum(b)) }},
		algo{name: "fletcher16", title: "Fletcher-16", tags: []string{"checksum", "rfc1146"}, sum: fletcher16},
		algo{name: "fletcher32", title: "Fletcher-32", tags: []string{"checksum"}, sum: fletcher32},
		std("fnv32", "FNV-1 32-bit", func() hash.Hash { return fnv.New32() }, "fowler noll vo", "non-cryptographic"),
		std("fnv32a", "FNV-1a 32-bit", func() hash.Hash { return fnv.New32a() }, "fowler noll vo", "non-cryptographic"),
		std("fnv64", "FNV-1 64-bit", func() hash.Hash { return fnv.New64() }, "fowler noll vo", "non-cryptographic"),
		std("fnv64a", "FNV-1a 64-bit", func() hash.Hash { return fnv.New64a() }, "fowler noll vo", "non-cryptographic"),
		std("fnv128", "FNV-1 128-bit", func() hash.Hash { return fnv.New128() }, "fowler noll vo", "non-cryptographic"),
		std("fnv128a", "FNV-1a 128-bit", func() hash.Hash { return fnv.New128a() }, "fowler noll vo", "non-cryptographic"),
		algo{name: "murmur3_32", title: "MurmurHash3 32-bit", custom: true, tags: []string{"murmur", "murmurhash", "non-cryptographic", "mmh3"}, sum: func(b []byte) []byte { return be32(murmur3x86_32(b, 0)) }},
		algo{name: "murmur3_128", title: "MurmurHash3 x64 128-bit", custom: true, tags: []string{"murmur", "murmurhash", "non-cryptographic", "mmh3"}, sum: func(b []byte) []byte {
			h1, h2 := murmur3x64_128(b, 0)
			return append(be64(h1), be64(h2)...)
		}},
		algo{name: "xxhash32", title: "xxHash32", custom: true, tags: []string{"xxh32", "xxhash", "non-cryptographic"}, sum: func(b []byte) []byte { return be32(xxh32(b, 0)) }},
		algo{name: "xxhash64", title: "xxHash64", custom: true, tags: []string{"xxh64", "xxhash", "non-cryptographic"}, sum: func(b []byte) []byte { return be64(xxh64(b, 0)) }},
		algo{name: "siphash", title: "SipHash-2-4", custom: true, tags: []string{"siphash24", "keyed", "hash table", "prf"}, sum: func(b []byte) []byte {
			var k [16]byte
			v := siphash24(k, b)
			out := make([]byte, 8)
			for i := range out {
				out[i] = byte(v >> (8 * uint(i)))
			}
			return out
		}},
		algo{name: "djb2", title: "DJB2 (Bernstein)", tags: []string{"daniel bernstein", "string hash", "non-cryptographic", "times 33"}, sum: djb2},
		algo{name: "djb2a", title: "DJB2a (xor variant)", tags: []string{"daniel bernstein", "string hash", "non-cryptographic", "times 33"}, sum: djb2a},
		algo{name: "sdbm", title: "SDBM", tags: []string{"string hash", "non-cryptographic"}, sum: sdbm},
		algo{name: "jenkins_oaat", title: "Jenkins one-at-a-time", tags: []string{"bob jenkins", "string hash", "non-cryptographic"}, sum: jenkinsOAAT},
	)
	return list
}

func init() {
	algos = buildAlgos()
	for _, a := range algos {
		algoIndex[a.name] = a
	}
	for i, a := range algos {
		if !a.custom {
			registerPlain(a, 100+i)
		}
	}
	registerGeneric()
}

// headline gives the launcher order of the main variants.
var headline = map[string]int{
	"md5": 1, "sha1": 2, "sha256": 3, "sha512": 4, "sha3_256": 5, "blake2b_256": 6, "crc32": 7, "adler32": 8,
}

func orderOf(name string, fallback int) int {
	if o, ok := headline[name]; ok {
		return o
	}
	return fallback
}

func keywordsFor(name string) []string {
	kw := []string{"hash", name}
	for _, p := range []string{"sha", "crc", "blake", "md", "fnv", "murmur", "xxhash", "keccak", "adler", "fletcher", "ripemd"} {
		if strings.HasPrefix(name, p) {
			kw = append(kw, p)
		}
	}
	if strings.HasPrefix(name, "sha3") || strings.HasPrefix(name, "sha5") || strings.HasPrefix(name, "sha2") {
		kw = append(kw, "sha2")
	}
	return kw
}

func reg(f tools.Func) {
	f.Group = group
	tools.Register(f)
}

func digestDoc(title string, bitsN int, extra string) string {
	return fmt.Sprintf("Computes the %s digest of text (%d bits by default). The text is hashed as UTF-8 bytes, or decoded first with input_encoding (hex, base64); the result is lowercase hex unless output asks for HEX, base64, base64url, base32 or dec (decimal number). %s", title, bitsN, extra)
}

func registerPlain(a algo, order int) {
	extra := ""
	switch {
	case strings.HasPrefix(a.name, "crc") || a.name == "adler32" || strings.HasPrefix(a.name, "fletcher") ||
		strings.HasPrefix(a.name, "fnv") || a.name == "djb2" || a.name == "djb2a" || a.name == "sdbm" || a.name == "jenkins_oaat":
		extra = "This is a non-cryptographic checksum: use it for error detection or hash tables, never for security."
	case a.name == "md4" || a.name == "md5" || a.name == "sha1" || a.name == "lm" || a.name == "ntlm" || a.name == "mysql41":
		extra = "Legacy algorithm: fine for fingerprints and compatibility, not for security."
	}
	switch a.name {
	case "ntlm":
		extra = "Windows NT hash: MD4 of the UTF-16LE password. " + extra
	case "lm":
		extra = "Legacy LAN Manager hash: password uppercased, cut to 14 bytes, two DES halves. " + extra
	case "mysql41":
		extra = "MySQL PASSWORD()/mysql_native_password hash: SHA1(SHA1(password)); MySQL stores it as '*' plus uppercase hex. " + extra
	case "sha256d":
		extra = "SHA-256 applied twice, as used by Bitcoin."
	case "hash160":
		extra = "RIPEMD-160 of SHA-256, as used by Bitcoin addresses."
	case "keccak256", "keccak512":
		extra = "Original Keccak padding (as in Ethereum), which differs from NIST SHA3."
	}
	reg(tools.Func{
		Name: a.name, Title: a.title,
		Doc:     digestDoc(a.title, a.bits(), extra),
		Params:  []tools.Param{textParam(), inEncParam(), outParam()},
		Returns: "str",
		Fn: func(args tools.Args) (any, error) {
			b, err := inBytes(args)
			if err != nil {
				return nil, err
			}
			return encodeBytes(a.compute(b), args.Str("output"))
		},
		Example:  fmt.Sprintf(`hash.%s("hello")`, a.name),
		Tags:     append([]string{"hash", "digest", "checksum", a.name}, a.tags...),
		Family:   "hash",
		Order:    orderOf(a.name, order),
		Keywords: keywordsFor(a.name),
	})
}

func lookupAlgo(name string) (algo, error) {
	n := normAlgo(name)
	if a, ok := algoIndex[n]; ok {
		return a, nil
	}
	aliases := map[string]string{
		"sha_1": "sha1", "sha_224": "sha224", "sha_256": "sha256", "sha_384": "sha384", "sha_512": "sha512",
		"sha2_256": "sha256", "sha2_512": "sha512", "sha3": "sha3_256", "keccak": "keccak256", "blake2b": "blake2b_512",
		"blake2s": "blake2s_256", "blake2": "blake2b_512", "ripemd": "ripemd160", "rmd160": "ripemd160", "crc": "crc32",
		"murmur": "murmur3_32", "murmur3": "murmur3_32", "xxh32": "xxhash32", "xxh64": "xxhash64", "xxhash": "xxhash64",
		"nt": "ntlm", "adler": "adler32", "fnv": "fnv64a", "fnv1a": "fnv64a", "fnv1": "fnv64", "crc32_c": "crc32c", "sha512_t256": "sha512_256",
		"sha3_": "sha3_256", "bitcoin": "sha256d", "mysql": "mysql41",
	}
	if t, ok := aliases[n]; ok {
		return algoIndex[t], nil
	}
	return algo{}, fmt.Errorf("unknown hash algorithm %q (see hash.algorithms())", name)
}

func algoNames() []string {
	out := make([]string, 0, len(algos))
	for _, a := range algos {
		out = append(out, a.name)
	}
	sort.Strings(out)
	return out
}

var commonAlgos = []string{
	"md5", "sha1", "sha224", "sha256", "sha384", "sha512", "sha3_256", "sha3_512", "keccak256",
	"blake2b_256", "blake2b_512", "blake2s_256", "blake3", "ripemd160", "crc32", "adler32", "fnv64a", "xxhash64", "murmur3_32",
}

func registerGeneric() {
	reg(tools.Func{
		Name: "hash", Title: "Hash with any algorithm",
		Doc: "Hashes text with the named algorithm (see algorithms(): md5, sha256, sha3_512, blake2b_256, blake3, crc32, xxhash64 and many more; '-' and '/' in names are accepted, so 'SHA-256' works). Algorithms with extra options (seed, key, length) use their defaults here; call their own function to change them. Text is UTF-8 unless input_encoding is hex or base64.",
		Params: []tools.Param{
			{Name: "algo", Type: "str", Required: true, Doc: "Algorithm name, for example sha256, md5, blake2b_512, crc32c."},
			textParam(), inEncParam(), outParam(),
		},
		Returns: "str", Input: "text",
		Fn: func(a tools.Args) (any, error) {
			al, err := lookupAlgo(a.Str("algo"))
			if err != nil {
				return nil, err
			}
			b, err := inBytes(a)
			if err != nil {
				return nil, err
			}
			return encodeBytes(al.compute(b), a.Str("output"))
		},
		Example: `hash.hash("sha256", "hello")`,
		Tags:    []string{"hash", "digest", "checksum", "generic", "any algorithm", "sha", "md5", "crc"},
	})
	reg(tools.Func{
		Name: "algorithms", Title: "List hash algorithms",
		Doc:     "Returns the sorted names accepted by hash.hash(), hash.hash_many() and hash.hash_length_bits(). HMAC accepts the subset that has a streaming form (md5, sha1, sha2, sha3, blake2, ripemd160, sm3, keccak).",
		Returns: "list[str]",
		Fn:      func(tools.Args) (any, error) { return algoNames(), nil },
		Example: `hash.algorithms()`,
		Tags:    []string{"list", "supported", "available", "algorithms", "hash names"},
	})
	reg(tools.Func{
		Name: "hash_all", Title: "All common hashes",
		Doc:     "Hashes text with the common algorithms at once (md5, sha1, sha2 family, sha3, keccak256, blake2, blake3, ripemd160, crc32, adler32, fnv64a, xxhash64, murmur3_32) and returns a dict of algorithm name to digest. Text is UTF-8 unless input_encoding is hex or base64.",
		Params:  []tools.Param{textParam(), inEncParam(), outParam()},
		Returns: "dict[str, str]",
		Fn: func(a tools.Args) (any, error) {
			b, err := inBytes(a)
			if err != nil {
				return nil, err
			}
			m := map[string]string{}
			for _, n := range commonAlgos {
				s, err := encodeBytes(algoIndex[n].compute(b), a.Str("output"))
				if err != nil {
					return nil, err
				}
				m[n] = s
			}
			return m, nil
		},
		Example:  `hash.hash_all("hello")`,
		Tags:     []string{"hash", "all", "every", "checksums", "digests", "md5", "sha1", "sha256", "crc32"},
		Keywords: []string{"checksum", "digest", "all"}, // API-only: it would flood the launcher list; the scope lists each algorithm
	})
	reg(tools.Func{
		Name: "hash_many", Title: "Hash with several algorithms",
		Doc:     "Hashes text with each of the listed algorithms and returns a dict of algorithm name to digest. Unknown algorithm names are an error.",
		Params:  []tools.Param{textParam(), {Name: "algos", Type: "list[str]", Required: true, Doc: "Algorithm names, for example [\"md5\", \"sha256\"]."}, inEncParam(), outParam()},
		Returns: "dict[str, str]",
		Fn: func(a tools.Args) (any, error) {
			b, err := inBytes(a)
			if err != nil {
				return nil, err
			}
			m := map[string]string{}
			for _, n := range a.List("algos") {
				al, err := lookupAlgo(n)
				if err != nil {
					return nil, err
				}
				s, err := encodeBytes(al.compute(b), a.Str("output"))
				if err != nil {
					return nil, err
				}
				m[n] = s
			}
			return m, nil
		},
		Example: `hash.hash_many("hello", ["md5", "sha256"])`,
		Tags:    []string{"hash", "several", "multiple", "batch", "digests"},
	})
	reg(tools.Func{
		Name: "hash_length_bits", Title: "Digest size in bits",
		Doc:     "Returns the default digest size in bits of an algorithm (md5 is 128, sha256 is 256, crc32 is 32). Variable-length ones report their default (shake128 256, shake256 512, blake3 256).",
		Params:  []tools.Param{{Name: "algo", Type: "str", Required: true, Doc: "Algorithm name."}},
		Returns: "int",
		Fn: func(a tools.Args) (any, error) {
			al, err := lookupAlgo(a.Str("algo"))
			if err != nil {
				return nil, err
			}
			return int64(al.bits()), nil
		},
		Example: `hash.hash_length_bits("sha256")`,
		Tags:    []string{"size", "length", "bits", "digest size", "output size"},
	})

	// Algorithms with extra parameters.
	registerKeyedAndSeeded()
	registerCRCGeneric()
}

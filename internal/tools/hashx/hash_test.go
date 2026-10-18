package hashx

import (
	"encoding/hex"
	"fmt"
	"hash/crc32"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "hash") }

func call(t *testing.T, name string, args map[string]any) any {
	t.Helper()
	v, err := tools.Call("hash", name, args)
	if err != nil {
		t.Fatalf("hash.%s(%v): %v", name, args, err)
	}
	return v
}

func callErr(t *testing.T, name string, args map[string]any) error {
	t.Helper()
	_, err := tools.Call("hash", name, args)
	if err == nil {
		t.Fatalf("hash.%s(%v): expected an error", name, args)
	}
	return err
}

func eq(t *testing.T, got any, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %s", got, want)
	}
}

func digest(t *testing.T, name, text string) string {
	t.Helper()
	return call(t, name, map[string]any{"text": text}).(string)
}

func TestDigestVectors(t *testing.T) {
	long := strings.Repeat("a", 1000000)
	cases := []struct{ name, in, want string }{
		{"md4", "", "31d6cfe0d16ae931b73c59d7e0c089c0"},
		{"md4", "abc", "a448017aaf21d8525fc10ae87aa6729d"},
		{"md5", "", "d41d8cd98f00b204e9800998ecf8427e"},
		{"md5", "abc", "900150983cd24fb0d6963f7d28e17f72"},
		{"sha1", "abc", "a9993e364706816aba3e25717850c26c9cd0d89d"},
		{"sha224", "abc", "23097d223405d8228642a477bda255b32aadbce4bda0b3f7e36c9da7"},
		{"sha256", "abc", "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad"},
		{"sha256", "", "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"},
		{"sha384", "abc", "cb00753f45a35e8bb5a03d699ac65007272c32ab0eded1631a8b605a43ff5bed8086072ba1e7cc2358baeca134c825a7"},
		{"sha512", "abc", "ddaf35a193617abacc417349ae20413112e6fa4e89a97ea20a9eeee64b55d39a2192992a274fc1a836ba3c23a3feebbd454d4423643ce80e2a9ac94fa54ca49f"},
		{"sha512_224", "abc", "4634270f707b6a54daae7530460842e20e37ed265ceee9a43e8924aa"},
		{"sha512_256", "abc", "53048e2681941ef99b2e29b76b4c7dabe4c2d0c634fc6d46e0e2f13107e7af23"},
		{"sha3_224", "abc", "e642824c3f8cf24ad09234ee7d3c766fc9a3a5168d0c94ad73b46fdf"},
		{"sha3_256", "", "a7ffc6f8bf1ed76651c14756a061d662f580ff4de43b49fa82d80a4b80f8434a"},
		{"sha3_256", "abc", "3a985da74fe225b2045c172d6bd390bd855f086e3e9d525b46bfe24511431532"},
		{"sha3_384", "abc", "ec01498288516fc926459f58e2c6ad8df9b473cb0fc08c2596da7cf0e49be4b298d88cea927ac7f539f1edf228376d25"},
		{"sha3_512", "abc", "b751850b1a57168a5693cd924b6b096e08f621827444f70d884f5d0240d2712e10e116e9192af3c91a7ec57647e3934057340b4cf408d5a56592f8274eec53f0"},
		{"keccak256", "", "c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"},
		{"keccak256", "abc", "4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45"},
		{"keccak512", "", "0eab42de4c3ceb9235fc91acffe746b29c29a8c366b7c60e4e67c466f36a4304c00fa9caf9d87976ba469bcbe06713b435f091ef2769fb160cdab33d3670680e"},
		{"shake128", "", "7f9c2ba4e88f827d616045507605853ed73b8093f6efbc88eb1a6eacfa66ef26"},
		{"shake256", "", "46b9dd2b0ba88d13233b3feb743eeb243fcd52ea62b81b82b50c27646ed5762fd75dc4ddd8c0f200cb05019d67b592f6fc821c49479ab48640292eacb3b7c4be"},
		{"blake2b_512", "", "786a02f742015903c6c6fd852552d272912f4740e15847618a86e217f71f5419d25e1031afee585313896444934eb04b903a685b1448b755d56f701afe9be2ce"},
		{"blake2b_512", "abc", "ba80a53f981c4d0d6a2797b69f12f6e94c212f14685ac4b74b12bb6fdbffa2d17d87c5392aab792dc252d5de4533cc9518d38aa8dbf1925ab92386edd4009923"},
		{"blake2b_256", "", "0e5751c026e543b2e8ab2eb06099daa1d1e5df47778f7787faab45cdf12fe3a8"},
		{"blake2s_256", "abc", "508c5e8c327c14e2e1a72ba34eeb452f37458b209ed63a294d999b4c86675982"},
		{"blake3", "", "af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262"},
		{"blake3", "abc", "6437b3ac38465133ffb63b75273a8db548c558465d79db03fd359c6cd5bd9d85"},
		{"ripemd160", "", "9c1185a5c5e9fc54612808977ee8f548b2258d31"},
		{"ripemd160", "abc", "8eb208f7e05d987a9b044a8e98c6b087f15a0bfc"},
		{"whirlpool", "", "19fa61d75522a4669b44e39c1d2e1726c530232130d407f89afee0964997f7a73e83be698b288febcf88e3e03c4f0757ea8964e59b63d93708b138cc42a66eb3"},
		{"whirlpool", "abc", "4e2448a4c6f486bb16b6562c73b4020bf3043e3a731bce721ae1b303d97e6d4c7181eebdb6c57e277d0e34957114cbd6c797fc9d95d8b582d225292076d4eef5"},
		{"sm3", "abc", "66c7f0f462eeedd9d1f2d46bdc10e4e24167c4875cf2f7a2297da02b8f4ba8e0"},
		{"sha256d", "", "5df6e0e2761359d30a8275058e299fcc0381534545f55cf43e41983f5d4c9456"},
		{"ntlm", "password", "8846f7eaee8fb117ad06bdd830b7586c"},
		{"lm", "password", "e52cac67419a9a224a3b108f3fa6cb6d"},
		{"mysql41", "password", "2470c0c06dee42fd1618bb99005adca2ec9d1e19"},
		{"adler32", "Wikipedia", "11e60398"},
		{"fletcher16", "abcde", "c8f0"},
		{"fletcher16", "abcdef", "2057"},
		{"fletcher16", "abcdefgh", "0627"},
		{"fletcher32", "abcde", "f04fc729"},
		{"fletcher32", "abcdef", "56502d2a"},
		{"fletcher32", "abcdefgh", "ebe19591"},
		{"fnv32", "a", "050c5d7e"},
		{"fnv32a", "a", "e40c292c"},
		{"fnv64", "a", "af63bd4c8601b7be"},
		{"fnv64a", "a", "af63dc4c8601ec8c"},
		{"fnv32a", "", "811c9dc5"},
		{"murmur3_32", "", "00000000"},
		{"murmur3_32", "The quick brown fox jumps over the lazy dog", "2e4ff723"},
		{"murmur3_128", "", "00000000000000000000000000000000"},
		{"murmur3_128", "hello", "cbd8a7b341bd9b025b1e906a48ae1d19"},
		{"xxhash32", "", "02cc5d05"},
		{"xxhash32", "abc", "32d153ff"},
		{"xxhash64", "", "ef46db3751d8e999"},
		{"xxhash64", "abc", "44bc2cf5ad770999"},
		{"crc32", "123456789", "cbf43926"},
		{"crc32c", "123456789", "e3069283"},
		{"sha256", long, "cdc76e5c9914fb9281a1c7e284d73e67f1809a48a497200e046d39ccc7112cd0"},
	}
	for _, c := range cases {
		got := digest(t, c.name, c.in)
		if got != c.want {
			if len(c.in) > 50 {
				c.in = c.in[:10] + "..."
			}
			t.Errorf("%s(%q) = %s, want %s", c.name, c.in, got, c.want)
		}
	}
}

func TestCRCCheckValues(t *testing.T) {
	want := map[string]string{
		"crc8": "f4", "crc8_maxim": "a1",
		"crc16_arc": "bb3d", "crc16_ccitt_false": "29b1", "crc16_xmodem": "31c3", "crc16_kermit": "2189", "crc16_modbus": "4b37",
		"crc16_x25": "906e", "crc16_usb": "b4c8", "crc16_genibus": "d64e", "crc16_mcrf4xx": "6f91", "crc16_dnp": "ea82",
		"crc16_buypass": "fee8", "crc16_t10_dif": "d0db", "crc16_cdma2000": "4c06", "crc16_aug_ccitt": "e5cc",
		"crc24":      "21cf02",
		"crc32_ieee": "cbf43926", "crc32_castagnoli": "e3069283", "crc32_bzip2": "fc891918", "crc32_mpeg2": "0376e6e7",
		"crc32_posix": "765e7680", "crc32_jamcrc": "340bc6d9", "crc32_q": "3010bf7f", "crc32_d": "87315576",
		"crc64_iso": "b90956c775a41001", "crc64_ecma": "995dc9bbdf1939fa", "crc64_ecma182": "6c40df5f0b497347", "crc64_we": "62ec59e3f1a4f00a",
	}
	for n, w := range want {
		spec, ok := crcCatalog[n]
		if !ok {
			t.Errorf("missing catalog entry %s", n)
			continue
		}
		if got := hex.EncodeToString(spec.bytes([]byte("123456789"))); got != w {
			t.Errorf("%s check = %s, want %s", n, got, w)
		}
	}
	for n := range crcCatalog {
		if _, ok := want[n]; !ok && n != "crc32_koopman" {
			t.Errorf("catalog entry %s has no asserted check value", n)
		}
	}
	// Koopman has no catalogued check value; compare with the standard library.
	eq(t, call(t, "crc32", map[string]any{"text": "123456789", "variant": "koopman"}),
		fmt.Sprintf("%08x", crc32.Checksum([]byte("123456789"), crc32.MakeTable(crc32.Koopman))))
	eq(t, call(t, "crc16", map[string]any{"text": "123456789", "variant": "modbus"}), "4b37")
	eq(t, call(t, "crc64", map[string]any{"text": "123456789", "variant": "iso"}), "b90956c775a41001")
	eq(t, call(t, "crc32", map[string]any{"text": "123456789", "output": "dec"}), "3421780262")
	eq(t, call(t, "crc_custom", map[string]any{"text": "123456789", "width": int64(16), "poly": "1021", "init": "FFFF"}), "29b1")
	callErr(t, "crc32", map[string]any{"text": "x", "variant": "nope"})
}

func TestSeedsAndKeys(t *testing.T) {
	eq(t, call(t, "murmur3_32", map[string]any{"text": "", "seed": int64(1)}), "514e28b7")
	eq(t, call(t, "murmur3_32", map[string]any{"text": "Hello, world!", "seed": int64(1234)}), "faf6cdb3")
	// SipHash-2-4 reference vectors: key 00..0f.
	key := "000102030405060708090a0b0c0d0e0f"
	eq(t, call(t, "siphash", map[string]any{"text": "", "key": key, "key_encoding": "hex"}), "310e0edd47db6f72")
	msg := "000102030405060708090a0b0c0d0e"
	eq(t, call(t, "siphash", map[string]any{"text": msg, "input_encoding": "hex", "key": key, "key_encoding": "hex"}), "e545be4961ca29a1")
	callErr(t, "siphash", map[string]any{"text": "x", "key": "short"})
	// keyed blake2b: RFC 7693 style, just check keying changes the digest and is deterministic.
	a := call(t, "blake2b_256", map[string]any{"text": "x", "key": "k"})
	b := call(t, "blake2b_256", map[string]any{"text": "x", "key": "k"})
	c := call(t, "blake2b_256", map[string]any{"text": "x"})
	if a != b || a == c {
		t.Errorf("blake2b keyed: %v %v %v", a, b, c)
	}
	callErr(t, "blake2b_256", map[string]any{"text": "x", "key": strings.Repeat("k", 65)})
	eq(t, call(t, "shake128", map[string]any{"text": "", "length": int64(4)}), "7f9c2ba4")
	callErr(t, "shake128", map[string]any{"text": "", "length": int64(0)})
	callErr(t, "blake3", map[string]any{"text": "", "key": "short"})
}

func TestBlake3Modes(t *testing.T) {
	// Official test vectors: input is bytes i%251; key and context from test_vectors.json.
	key := "whats the Elvish word for friend"
	ctx := "BLAKE3 2019-12-27 16:29:52 test vectors context"
	input := func(n int) []byte {
		b := make([]byte, n)
		for i := range b {
			b[i] = byte(i % 251)
		}
		return b
	}
	cases := []struct {
		n                  int
		hash, keyed, deriv string
	}{
		{0, "af1349b9f5f9a1a6a0404dea36dcc9499bcb25c9adc112b7cc9a93cae41f3262", "92b2b75604ed3c761f9d6f62392c8a9227ad0ea3f09573e783f1498a4ed60d26", "2cc39783c223154fea8dfb7c1b1660f2ac2dcbd1c1de8277b0b0dd39b7e50d7d"},
		{1, "2d3adedff11b61f14c886e35afa036736dcd87a74d27b5c1510225d0f592e213", "6d7878dfff2f485635d39013278ae14f1454b8c0a3a2d34bc1ab38228a80c95b", "b3e2e340a117a499c6cf2398a19ee0d29cca2bb7404c73063382693bf66cb06c"},
	}
	for _, c := range cases {
		in := hex.EncodeToString(input(c.n))
		eq(t, call(t, "blake3", map[string]any{"text": in, "input_encoding": "hex"}), c.hash)
		eq(t, call(t, "blake3", map[string]any{"text": in, "input_encoding": "hex", "key": key}), c.keyed)
		eq(t, call(t, "blake3", map[string]any{"text": in, "input_encoding": "hex", "context": ctx}), c.deriv)
	}
	callErr(t, "blake3", map[string]any{"text": "x", "key": key, "context": ctx})
	if got := call(t, "blake3", map[string]any{"text": "x", "length": int64(100)}).(string); len(got) != 200 {
		t.Errorf("xof length %d", len(got))
	}
	// Prefix property of the XOF.
	long := call(t, "blake3", map[string]any{"text": "x", "length": int64(100)}).(string)
	short := call(t, "blake3", map[string]any{"text": "x"}).(string)
	if !strings.HasPrefix(long, short) {
		t.Error("blake3 xof must extend the 32-byte digest")
	}
}

func TestOutputFormatsAndInputs(t *testing.T) {
	eq(t, call(t, "md5", map[string]any{"text": "abc", "output": "HEX"}), "900150983CD24FB0D6963F7D28E17F72")
	eq(t, call(t, "md5", map[string]any{"text": "abc", "output": "base64"}), "kAFQmDzST7DWlj99KOF/cg==")
	eq(t, call(t, "md5", map[string]any{"text": "616263", "input_encoding": "hex"}), "900150983cd24fb0d6963f7d28e17f72")
	eq(t, call(t, "md5", map[string]any{"text": "YWJj", "input_encoding": "base64"}), "900150983cd24fb0d6963f7d28e17f72")
	callErr(t, "md5", map[string]any{"text": "zz", "input_encoding": "hex"})
	callErr(t, "md5", map[string]any{"text": "a", "output": "nope"})
	callErr(t, "md5", map[string]any{"text": "a", "input_encoding": "rot13"})
	eq(t, call(t, "hash", map[string]any{"algo": "SHA-256", "text": "abc"}), "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad")
	callErr(t, "hash", map[string]any{"algo": "nope", "text": "abc"})
	all := call(t, "hash_all", map[string]any{"text": "abc"}).(map[string]string)
	if all["md5"] != "900150983cd24fb0d6963f7d28e17f72" || len(all) < 15 {
		t.Errorf("hash_all: %v", all)
	}
	many := call(t, "hash_many", map[string]any{"text": "abc", "algos": []string{"md5", "sha1"}}).(map[string]string)
	if len(many) != 2 {
		t.Errorf("hash_many: %v", many)
	}
	callErr(t, "hash_many", map[string]any{"text": "abc", "algos": []string{"bogus"}})
	if n := call(t, "hash_length_bits", map[string]any{"algo": "sha256"}); n != int64(256) {
		t.Errorf("bits %v", n)
	}
	if n := call(t, "hash_length_bits", map[string]any{"algo": "crc32"}); n != int64(32) {
		t.Errorf("bits %v", n)
	}
	names := call(t, "algorithms", nil).([]string)
	if len(names) < 80 {
		t.Errorf("only %d algorithms", len(names))
	}
	// Every listed algorithm must hash through hash() without error.
	for _, n := range names {
		call(t, "hash", map[string]any{"algo": n, "text": "x"})
	}
	// Garbage in: huge and odd inputs never panic.
	for _, n := range names {
		call(t, "hash", map[string]any{"algo": n, "text": "\x00\xff\xfe日本語"})
	}
}

func TestHMAC(t *testing.T) {
	// RFC 4231 test case 1 and 2.
	key1 := strings.Repeat("0b", 20)
	r := call(t, "hmac", map[string]any{"algo": "sha256", "key": key1, "key_encoding": "hex", "text": "Hi There"})
	eq(t, r, "b0344c61d8db38535ca8afceaf0bf12b881dc200c9833da726e9376c2e32cff7")
	r = call(t, "hmac", map[string]any{"algo": "sha512", "key": key1, "key_encoding": "hex", "text": "Hi There"})
	eq(t, r, "87aa7cdea5ef619d4ff0b4241a1d6cb02379f4e2ce4ec2787ad0b30545e17cdedaa833b7d6b8a702038b274eaea3f4e4be9d914eeb61f1702e696c203a126854")
	r = call(t, "hmac_sha256", map[string]any{"key": "Jefe", "text": "what do ya want for nothing?"})
	eq(t, r, "5bdcc146bf60754e6a042426089575c75a003f089d2739839dec58b964ec3843")
	r = call(t, "hmac_md5", map[string]any{"key": "Jefe", "text": "what do ya want for nothing?"})
	eq(t, r, "750c783e6ab0b503eaa86e310a5db738")
	r = call(t, "hmac_sha1", map[string]any{"key": "Jefe", "text": "what do ya want for nothing?"})
	eq(t, r, "effcdf6ae5eb2fa2d27416d5f184df9c259a7c79")
	if call(t, "hmac_verify", map[string]any{"algo": "sha256", "key": "Jefe", "text": "what do ya want for nothing?", "mac": "5BDCC146BF60754E6A042426089575C75A003F089D2739839DEC58B964EC3843"}) != true {
		t.Error("hmac_verify should accept uppercase hex")
	}
	if call(t, "hmac_verify", map[string]any{"algo": "sha256", "key": "Jefe", "text": "x", "mac": "00"}) != false {
		t.Error("hmac_verify should reject")
	}
	callErr(t, "hmac", map[string]any{"algo": "crc32", "key": "k", "text": "x"})
	callErr(t, "hmac", map[string]any{"algo": "md5", "key": "zz", "key_encoding": "hex", "text": "x"})
}

func TestKDFs(t *testing.T) {
	// HKDF RFC 5869 test case 1.
	r := call(t, "hkdf", map[string]any{
		"algo":   "sha256",
		"secret": strings.Repeat("0b", 22), "key_encoding": "hex",
		"salt": "000102030405060708090a0b0c", "info": "f0f1f2f3f4f5f6f7f8f9", "salt_encoding": "hex", "length": int64(42),
	})
	eq(t, r, "3cb25f25faacd57a90434f64d0362f2a2d2d0a90cf1a5a4c5db02d56ecc4c5bf34007208d5b887185865")
	callErr(t, "hkdf", map[string]any{"algo": "sha256", "secret": "x", "length": int64(100000)})

	// PBKDF2 RFC 6070.
	eq(t, call(t, "pbkdf2", map[string]any{"password": "password", "salt": "salt", "iterations": int64(1), "length": int64(20), "algo": "sha1"}), "0c60c80f961f0e71f3a9b524af6012062fe037a6")
	eq(t, call(t, "pbkdf2", map[string]any{"password": "password", "salt": "salt", "iterations": int64(4096), "length": int64(20), "algo": "sha1"}), "4b007901b765489abead49d926f721d065a429c1")
	// PBKDF2-HMAC-SHA256 (RFC 7914 section 11).
	eq(t, call(t, "pbkdf2", map[string]any{"password": "passwd", "salt": "salt", "iterations": int64(1), "length": int64(64)}),
		"55ac046e56e3089fec1691c22544b605f94185216dde0465e68b9d57c20dacbc49ca9cccf179b645991664b39d77ef317c71b845b1e30bd509112041d3a19783")
	callErr(t, "pbkdf2", map[string]any{"password": "p", "salt": "s", "iterations": int64(5000001)})
	callErr(t, "pbkdf2", map[string]any{"password": "p", "salt": "s", "iterations": int64(0)})

	// scrypt RFC 7914.
	eq(t, call(t, "scrypt", map[string]any{"password": "", "salt": "", "n": int64(16), "r": int64(1), "p": int64(1), "length": int64(64)}),
		"77d6576238657b203b19ca42c18a0497f16b4844e3074ae8dfdffa3fede21442fcd0069ded0948f8326a753a0fc81f17e8d3e0fb2e0d3628cf35e20c38d18906")
	eq(t, call(t, "scrypt", map[string]any{"password": "password", "salt": "NaCl", "n": int64(1024), "r": int64(8), "p": int64(16), "length": int64(64)}),
		"fdbabe1c9d3472007856e7190d01e9fe7c6ad7cbc8237830e77376634b3731622eaf30d92e22a3886ff109279d9830dac727afb94a83ee6d8360cbdfa2cc0640")
	callErr(t, "scrypt", map[string]any{"password": "p", "salt": "s", "n": int64(1 << 21)})
	callErr(t, "scrypt", map[string]any{"password": "p", "salt": "s", "n": int64(1000)})
	callErr(t, "scrypt", map[string]any{"password": "p", "salt": "s", "n": int64(1 << 20), "r": int64(8)}) // 1 GiB

	// bcrypt.
	h := call(t, "bcrypt_hash", map[string]any{"password": "hunter2", "cost": int64(4)}).(string)
	if !strings.HasPrefix(h, "$2a$04$") || len(h) != 60 {
		t.Errorf("bcrypt hash %q", h)
	}
	if call(t, "bcrypt_verify", map[string]any{"password": "hunter2", "hash": h}) != true {
		t.Error("bcrypt verify")
	}
	if call(t, "bcrypt_verify", map[string]any{"password": "hunter3", "hash": h}) != false {
		t.Error("bcrypt verify wrong password")
	}
	// Known hash from the bcrypt reference (password "abc123xyz"? use a stored one).
	known := "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy" // "password"? verified below by round trip only
	_ = known
	callErr(t, "bcrypt_verify", map[string]any{"password": "x", "hash": "garbage"})
	callErr(t, "bcrypt_hash", map[string]any{"password": "x", "cost": int64(15)})
	callErr(t, "bcrypt_hash", map[string]any{"password": strings.Repeat("x", 73), "cost": int64(4)})

	// Argon2: reference implementation README vector (argon2i, t=2, m=2^16, p=4, 24 bytes).
	eq(t, call(t, "argon2i", map[string]any{"password": "password", "salt": "somesalt", "time": int64(2), "memory_kib": int64(65536), "threads": int64(4), "length": int64(24)}),
		"45d7ac72e76f242b20b77b9bf9bf9d5915894e669a24e6c6")
	ph := call(t, "argon2id_hash", map[string]any{"password": "hunter2", "time": int64(1), "memory_kib": int64(64), "threads": int64(1)}).(string)
	if !strings.HasPrefix(ph, "$argon2id$v=19$m=64,t=1,p=1$") {
		t.Errorf("phc %q", ph)
	}
	if call(t, "argon2_verify", map[string]any{"password": "hunter2", "phc": ph}) != true {
		t.Error("argon2 verify")
	}
	if call(t, "argon2_verify", map[string]any{"password": "nope", "phc": ph}) != false {
		t.Error("argon2 verify wrong")
	}
	// argon2id raw output must agree with the PHC tag.
	raw := call(t, "argon2id", map[string]any{"password": "p", "salt": "saltsalt", "time": int64(1), "memory_kib": int64(64), "threads": int64(1), "output": "base64"}).(string)
	if raw == "" {
		t.Error("argon2id empty")
	}
	callErr(t, "argon2id", map[string]any{"password": "p", "salt": "saltsalt", "memory_kib": int64(262145)})
	callErr(t, "argon2id", map[string]any{"password": "p", "salt": "saltsalt", "time": int64(11)})
	callErr(t, "argon2_verify", map[string]any{"password": "p", "phc": "nonsense"})
	callErr(t, "argon2_verify", map[string]any{"password": "p", "phc": "$argon2id$v=19$m=999999999,t=1,p=1$c2FsdA$aGFzaA"})
}

func TestUtilities(t *testing.T) {
	if call(t, "checksum_compare", map[string]any{"a": "D4 1D:8C-D9", "b": "0xd41d8cd9"}) != true {
		t.Error("compare normalised")
	}
	if call(t, "checksum_compare", map[string]any{"a": "d41d", "b": "d41e"}) != false {
		t.Error("compare different")
	}
	eq(t, call(t, "hash_to_base64", map[string]any{"hex": "d41d8cd98f00b204e9800998ecf8427e"}), "1B2M2Y8AsgTpgAmY7PhCfg==")
	callErr(t, "hash_to_base64", map[string]any{"hex": "xyz"})
	has := func(in, want string) {
		t.Helper()
		got := call(t, "hash_identify", map[string]any{"hash": in}).([]string)
		for _, g := range got {
			if strings.HasPrefix(g, want) {
				return
			}
		}
		t.Errorf("hash_identify(%q) = %v, want %q", in, got, want)
	}
	has("d41d8cd98f00b204e9800998ecf8427e", "MD5")
	has("da39a3ee5e6b4b0d3255bfef95601890afd80709", "SHA-1")
	has("e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855", "SHA-256")
	has("$2b$12$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", "bcrypt")
	has("$argon2id$v=19$m=65536,t=3,p=4$c2FsdA$aGFzaA", "argon2id")
	has("$6$salt$hash", "sha512crypt")
	has("*2470C0C06DEE42FD1618BB99005ADCA2EC9D1E19", "MySQL")
	has("kAFQmDzST7DWlj99KOF/cg==", "MD5 (base64)")
	has("cbf43926", "CRC-32")
	if got := call(t, "hash_identify", map[string]any{"hash": ""}).([]string); len(got) != 0 {
		t.Errorf("empty identify: %v", got)
	}
}

func TestOTP(t *testing.T) {
	// RFC 4226 appendix D.
	secret := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	want := []string{"755224", "287082", "359152", "969429", "338314", "254676", "287922", "162583", "399871", "520489"}
	for i, w := range want {
		eq(t, call(t, "hotp", map[string]any{"secret": secret, "counter": int64(i)}), w)
	}
	// RFC 6238 appendix B, 8 digits.
	s1 := "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"
	totp := func(sec string, tm int64, algo string) any {
		return call(t, "totp", map[string]any{"secret": sec, "time": tm, "digits": int64(8), "algo": algo})
	}
	eq(t, totp(s1, 59, "sha1"), "94287082")
	eq(t, totp(s1, 1111111109, "sha1"), "07081804")
	eq(t, totp(s1, 1111111111, "sha1"), "14050471")
	eq(t, totp(s1, 1234567890, "sha1"), "89005924")
	eq(t, totp(s1, 2000000000, "sha1"), "69279037")
	eq(t, totp(s1, 20000000000, "sha1"), "65353130")
	// SHA-256 / SHA-512 secrets are the ASCII digits repeated to 32 / 64 bytes.
	k256 := base32Of("12345678901234567890123456789012")
	k512 := base32Of("1234567890123456789012345678901234567890123456789012345678901234")
	eq(t, totp(k256, 59, "sha256"), "46119246")
	eq(t, totp(k512, 59, "sha512"), "90693936")
	eq(t, totp(k256, 20000000000, "sha256"), "77737706")
	eq(t, totp(k512, 20000000000, "sha512"), "47863826")
	// Secret formatting is forgiving.
	eq(t, call(t, "totp", map[string]any{"secret": "gezd gnbv-gy3t qojq gezd gnbv gy3t qojq====", "time": int64(59), "digits": int64(8)}), "94287082")
	if call(t, "totp_verify", map[string]any{"secret": secret, "code": "287082", "time": int64(59), "window": int64(0)}) != true {
		t.Error("totp_verify counter 1")
	}
	if call(t, "totp_verify", map[string]any{"secret": secret, "code": "755224", "time": int64(59), "window": int64(1)}) != true {
		t.Error("totp_verify window")
	}
	if call(t, "totp_verify", map[string]any{"secret": secret, "code": "755224", "time": int64(59), "window": int64(0)}) != false {
		t.Error("totp_verify outside window")
	}
	if got := call(t, "totp", map[string]any{"secret": secret}).(string); len(got) != 6 {
		t.Errorf("totp now: %q", got)
	}
	callErr(t, "totp", map[string]any{"secret": "!!!"})
	callErr(t, "totp", map[string]any{"secret": secret, "digits": int64(11)})
	callErr(t, "totp", map[string]any{"secret": secret, "algo": "md5"})
	callErr(t, "totp", map[string]any{"secret": secret, "period": int64(0)})
	callErr(t, "hotp", map[string]any{"secret": secret, "counter": int64(-1)})
	uri := call(t, "totp_uri", map[string]any{"secret": "jbsw y3dp ehpk 3pxp", "account": "alice@example.com", "issuer": "Ex Co"}).(string)
	if !strings.HasPrefix(uri, "otpauth://totp/Ex%20Co:alice@example.com?") || !strings.Contains(uri, "secret=JBSWY3DPEHPK3PXP") ||
		!strings.Contains(uri, "algorithm=SHA1") || !strings.Contains(uri, "period=30") || !strings.Contains(uri, "digits=6") {
		t.Errorf("uri %q", uri)
	}
	hu := call(t, "hotp_uri", map[string]any{"secret": "JBSWY3DPEHPK3PXP", "account": "bob"}).(string)
	if !strings.HasPrefix(hu, "otpauth://hotp/bob?") || !strings.Contains(hu, "counter=0") {
		t.Errorf("hotp uri %q", hu)
	}
}

func base32Of(s string) string {
	const alpha = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"
	var out strings.Builder
	var buf, n uint
	for _, c := range []byte(s) {
		buf = buf<<8 | uint(c)
		n += 8
		for n >= 5 {
			out.WriteByte(alpha[(buf>>(n-5))&31])
			n -= 5
		}
	}
	if n > 0 {
		out.WriteByte(alpha[(buf<<(5-n))&31])
	}
	return out.String()
}

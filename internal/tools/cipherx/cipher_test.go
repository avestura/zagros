package cipherx

import (
	"encoding/hex"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "cipher") }

func call(t *testing.T, name string, args map[string]any) any {
	t.Helper()
	v, err := tools.Call("cipher", name, args)
	if err != nil {
		t.Fatalf("cipher.%s(%v): %v", name, args, err)
	}
	return v
}

func callErr(t *testing.T, name string, args map[string]any) error {
	t.Helper()
	_, err := tools.Call("cipher", name, args)
	if err == nil {
		t.Fatalf("cipher.%s(%v): expected an error", name, args)
	}
	return err
}

func eq(t *testing.T, got any, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %s", got, want)
	}
}

func hx(s string) []byte {
	b, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return b
}

func mustScheme(t *testing.T, fn string) scheme {
	t.Helper()
	s, err := findScheme(fn)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestKnownVectors(t *testing.T) {
	// AES-GCM, McGrew-Viega test cases 1 and 3.
	g := mustScheme(t, "aes_gcm")
	ct, _ := g.seal(make([]byte, 16), make([]byte, 12), nil, nil)
	if hex.EncodeToString(ct) != "58e2fccefa7e3061367f1d57a4e7455a" {
		t.Errorf("gcm case 1: %x", ct)
	}
	key := hx("feffe9928665731c6d6a8f9467308308")
	iv := hx("cafebabefacedbaddecaf888")
	pt := hx("d9313225f88406e5a55909c5aff5269a86a7a9531534f7da2e4c303d8a318a721c3c0c95956809532fcf0e2449a6b525b16aedf5aa0de657ba637b391aafd255")
	ct, _ = g.seal(key, iv, pt, nil)
	want := "42831ec2217774244b7221b784d0d49ce3aa212f2c02a4e035c17e2329aca12e21d514b25466931c7d8f6a5aac84aa051ba30b396a0aac973d58e091473f5985" + "4d5c2af327cd64a62cf35abd2ba6fab4"
	if hex.EncodeToString(ct) != want {
		t.Errorf("gcm case 3: %x", ct)
	}
	// Public API with an explicit IV is deterministic: iv|ct|tag.
	out := call(t, "aes_gcm_encrypt", map[string]any{"text": string(pt[:0]) + hex.EncodeToString(pt), "input_encoding": "hex",
		"key": "feffe9928665731c6d6a8f9467308308", "key_encoding": "hex", "iv": "cafebabefacedbaddecaf888", "output": "hex"})
	eq(t, out, "cafebabefacedbaddecaf888"+want)
	eq(t, call(t, "aes_gcm_decrypt", map[string]any{"data": out, "data_encoding": "hex", "key": "feffe9928665731c6d6a8f9467308308", "key_encoding": "hex", "output": "hex"}), hex.EncodeToString(pt))

	// NIST SP 800-38A, AES-128, first block.
	k := hx("2b7e151628aed2a6abf7158809cf4f3c")
	p1 := hx("6bc1bee22e409f96e93d7e117393172a")
	first := func(fn string, nonce string, n int, wantHex string) {
		t.Helper()
		s := mustScheme(t, fn)
		c, err := s.seal(k, hx(nonce), p1, nil)
		if err != nil || hex.EncodeToString(c[:n]) != wantHex {
			t.Errorf("%s: %x %v", fn, c, err)
		}
	}
	first("aes_ecb", "", 16, "3ad77bb40d7a3660a89ecaf32466ef97")
	first("aes_cbc", "000102030405060708090a0b0c0d0e0f", 16, "7649abac8119b246cee98e9b12e9197d")
	first("aes_ctr", "f0f1f2f3f4f5f6f7f8f9fafbfcfdfeff", 16, "874d6191b620e3261bef6864990db6ce")
	first("aes_cfb", "000102030405060708090a0b0c0d0e0f", 16, "3b3fd92eb72dad20333449f8e83cfb4a")
	first("aes_ofb", "000102030405060708090a0b0c0d0e0f", 16, "3b3fd92eb72dad20333449f8e83cfb4a")

	// RFC 8439 2.8.2 ChaCha20-Poly1305.
	cc := mustScheme(t, "chacha20_poly1305")
	rk := hx("808182838485868788898a8b8c8d8e8f909192939495969798999a9b9c9d9e9f")
	msg := []byte("Ladies and Gentlemen of the class of '99: If I could offer you only one tip for the future, sunscreen would be it.")
	c, _ := cc.seal(rk, hx("070000004041424344454647"), msg, hx("50515253c0c1c2c3c4c5c6c7"))
	if !strings.HasPrefix(hex.EncodeToString(c), "d31a8d34648e60db7b86afbc53ef7ec2") || !strings.HasSuffix(hex.EncodeToString(c), "1ae10b594f09e26a7e902ecbd0600691") {
		t.Errorf("chacha20poly1305: %x", c)
	}
	// XChaCha20-Poly1305 draft-irtf-cfrg-xchacha A.3.1.
	xc := mustScheme(t, "xchacha20_poly1305")
	c, _ = xc.seal(rk, hx("404142434445464748494a4b4c4d4e4f5051525354555657"), msg, hx("50515253c0c1c2c3c4c5c6c7"))
	if !strings.HasPrefix(hex.EncodeToString(c), "bd6d179d3e83d43b9576579493c0e939") || !strings.HasSuffix(hex.EncodeToString(c), "c0875924c1c7987947deafd8780acf49") {
		t.Errorf("xchacha20poly1305: %x", c)
	}
	// ChaCha20 keystream, RFC 7539 A.1 test vector 1.
	c20 := mustScheme(t, "chacha20")
	c, _ = c20.seal(make([]byte, 32), make([]byte, 12), make([]byte, 32), nil)
	if hex.EncodeToString(c) != "76b8e0ada0f13d90405d6ae55386bd28bdd219b8a08ded1aa836efcc8b770dc7" {
		t.Errorf("chacha20 keystream: %x", c)
	}
	// RC4 classic vectors.
	eq(t, call(t, "rc4_encrypt", map[string]any{"text": "Plaintext", "key": "Key", "key_encoding": "utf-8", "output": "hex"}), "bbf316e8d940af0ad3")
	eq(t, call(t, "rc4_encrypt", map[string]any{"text": "pedia", "key": "Wiki", "key_encoding": "utf-8", "output": "hex"}), "1021bf0420")
	eq(t, call(t, "rc4_decrypt", map[string]any{"data": "bbf316e8d940af0ad3", "data_encoding": "hex", "key": "Key", "key_encoding": "utf-8"}), "Plaintext")
}

func TestSymmetricRoundTrips(t *testing.T) {
	msgs := []string{"", "hello world", strings.Repeat("0123456789abcdef", 5) + "日本語"}
	for _, s := range schemes {
		for _, m := range msgs {
			// Passphrase mode.
			env := call(t, s.fn+"_encrypt", map[string]any{"text": m, "key": "correct horse"}).(string)
			if !strings.HasPrefix(env, "zg1.") {
				t.Fatalf("%s: envelope %q", s.fn, env)
			}
			eq(t, call(t, s.fn+"_decrypt", map[string]any{"data": env, "key": "correct horse"}), m)
			eq(t, call(t, "decrypt", map[string]any{"data": env, "key": "correct horse"}), m)
			// Wrong passphrase must fail, for every cipher (envelopes are always authenticated).
			callErr(t, s.fn+"_decrypt", map[string]any{"data": env, "key": "wrong horse"})
			// Tampered envelope must fail.
			bad := []byte(env)
			bad[len(bad)-3] ^= 1
			if _, err := tools.Call("cipher", s.fn+"_decrypt", map[string]any{"data": string(bad), "key": "correct horse"}); err == nil && m != "" {
				t.Errorf("%s: tampered envelope accepted", s.fn)
			}
		}
		// Raw key mode.
		klen := 32
		if s.fn == "des3_cbc" {
			klen = 24
		}
		key := strings.Repeat("ab", klen)
		for _, m := range msgs {
			enc := call(t, s.fn+"_encrypt", map[string]any{"text": m, "key": key, "key_encoding": "hex"}).(string)
			eq(t, call(t, s.fn+"_decrypt", map[string]any{"data": enc, "key": key, "key_encoding": "hex"}), m)
			eq(t, call(t, "decrypt", map[string]any{"algo": s.fn, "data": enc, "key": key, "key_encoding": "hex"}), m)
			e2 := call(t, "encrypt", map[string]any{"algo": s.fn, "text": m, "key": key, "key_encoding": "hex", "output": "hex"}).(string)
			eq(t, call(t, "decrypt", map[string]any{"algo": s.fn, "data": e2, "data_encoding": "hex", "key": key, "key_encoding": "hex"}), m)
		}
		// Wrong raw key length is a clear error.
		callErr(t, s.fn+"_encrypt", map[string]any{"text": "x", "key": "", "key_encoding": "hex"})
		// Raw data with a passphrase key and vice versa.
		callErr(t, s.fn+"_decrypt", map[string]any{"data": "AAAAAAAAAAAAAAAAAAAAAAAAAAAA", "key": "pass"})
	}
	// Authenticated ciphers fail in raw mode with a wrong key / tampering.
	for _, fn := range []string{"aes_gcm", "chacha20_poly1305", "xchacha20_poly1305"} {
		k1, k2 := strings.Repeat("11", 32), strings.Repeat("22", 32)
		enc := call(t, fn+"_encrypt", map[string]any{"text": "secret", "key": k1, "key_encoding": "hex", "aad": "ctx"}).(string)
		callErr(t, fn+"_decrypt", map[string]any{"data": enc, "key": k2, "key_encoding": "hex", "aad": "ctx"})
		callErr(t, fn+"_decrypt", map[string]any{"data": enc, "key": k1, "key_encoding": "hex", "aad": "other"})
		eq(t, call(t, fn+"_decrypt", map[string]any{"data": enc, "key": k1, "key_encoding": "hex", "aad": "ctx"}), "secret")
	}
	// Decrypting with a supplied iv (data without an iv prefix).
	enc := call(t, "aes_cbc_encrypt", map[string]any{"text": "hello", "key": strings.Repeat("00", 16), "key_encoding": "hex", "iv": strings.Repeat("01", 16)}).(string)
	raw, _ := decodeBytes(enc, "base64")
	if len(raw) != 32 {
		t.Fatalf("iv|ct length %d", len(raw))
	}
	noIV := hex.EncodeToString(raw[16:])
	eq(t, call(t, "aes_cbc_decrypt", map[string]any{"data": noIV, "data_encoding": "hex", "key": strings.Repeat("00", 16), "key_encoding": "hex", "iv": strings.Repeat("01", 16)}), "hello")
	// Bits.
	e := call(t, "aes_gcm_encrypt", map[string]any{"text": "x", "key": "pw", "bits": int64(128)}).(string)
	if !strings.HasPrefix(e, "zg1.aes128gcm.") {
		t.Errorf("bits=128 label: %s", e)
	}
	eq(t, call(t, "aes_gcm_decrypt", map[string]any{"data": e, "key": "pw"}), "x")
	callErr(t, "aes_gcm_encrypt", map[string]any{"text": "x", "key": "pw", "bits": int64(100)})
	callErr(t, "aes_gcm_encrypt", map[string]any{"text": "x", "key": strings.Repeat("00", 16), "key_encoding": "hex", "bits": int64(256)})
	// Misc errors.
	callErr(t, "aes_gcm_encrypt", map[string]any{"text": "x", "key": ""})
	callErr(t, "aes_gcm_encrypt", map[string]any{"text": "x", "key": "k", "iv": "zz"})
	callErr(t, "aes_gcm_encrypt", map[string]any{"text": "x", "key": "k", "iv": "0102"})
	callErr(t, "aes_gcm_decrypt", map[string]any{"data": "zg1.nope.AAAA", "key": "k"})
	callErr(t, "aes_gcm_decrypt", map[string]any{"data": "zg1.aes256gcm", "key": "k"})
	callErr(t, "aes_gcm_decrypt", map[string]any{"data": "zg1.aes256gcm.AAAA", "key": "k"})
	callErr(t, "decrypt", map[string]any{"data": "AAAA", "key": "k"})
	callErr(t, "encrypt", map[string]any{"algo": "rot13", "text": "x", "key": "k"})
	// A chacha envelope refuses to open as AES.
	cenv := call(t, "chacha20_poly1305_encrypt", map[string]any{"text": "x", "key": "pw"}).(string)
	callErr(t, "aes_gcm_decrypt", map[string]any{"data": cenv, "key": "pw"})
	// Garbage never panics.
	for _, s := range schemes {
		for _, g := range []string{"", "zg1.", "zg1..", "!!!", "zg1.aes256gcm.!!!", "\x00\xff"} {
			tools.Call("cipher", s.fn+"_decrypt", map[string]any{"data": g, "key": "k"})
			tools.Call("cipher", s.fn+"_decrypt", map[string]any{"data": g, "key": strings.Repeat("00", 32), "key_encoding": "hex"})
		}
	}
	if len(call(t, "algorithms", nil).([]string)) != len(schemes) {
		t.Error("algorithms")
	}
	if !looksEncrypted("zg1.aes256gcm.abc") || looksEncrypted("hi") {
		t.Error("looksEncrypted")
	}
}

func TestXOR(t *testing.T) {
	enc := call(t, "xor_cipher", map[string]any{"text": "Burning 'em, if you ain't quick and nimble\nI go crazy when I hear a cymbal", "key": "ICE"})
	eq(t, enc, "0b3637272a2b2e63622c2e69692a23693a2a3c6324202d623d63343c2a26226324272765272a282b2f20430a652e2c652a3124333a653e2b2027630c692b20283165286326302e27282f")
	eq(t, call(t, "xor_decrypt", map[string]any{"data": enc, "key": "ICE"}), "Burning 'em, if you ain't quick and nimble\nI go crazy when I hear a cymbal")
	callErr(t, "xor_cipher", map[string]any{"text": "a", "key": ""})
	callErr(t, "xor_decrypt", map[string]any{"data": "zz", "key": "k"})
	res := call(t, "xor_brute_single_byte", map[string]any{"hex": "1b37373331363f78151b7f2b783431333d78397828372d363c78373e783a393b3736"}).([]string)
	if len(res) != 5 || !strings.HasPrefix(res[0], "0x58 (X)") || !strings.HasSuffix(res[0], "Cooking MC's like a pound of bacon") {
		t.Errorf("brute: %v", res)
	}
	callErr(t, "xor_brute_single_byte", map[string]any{"hex": ""})
	callErr(t, "xor_brute_single_byte", map[string]any{"hex": "xyz"})
}

func TestClassic(t *testing.T) {
	eq(t, call(t, "caesar", map[string]any{"text": "Hello, World!", "shift": int64(3)}), "Khoor, Zruog!")
	eq(t, call(t, "caesar", map[string]any{"text": "Hello", "shift": int64(-3)}), "Ebiil")
	eq(t, call(t, "caesar", map[string]any{"text": "abc", "shift": int64(29)}), "def")
	eq(t, call(t, "caesar_decrypt", map[string]any{"text": "Khoor, Zruog!"}), "Hello, World!")
	br := call(t, "caesar_brute", map[string]any{"text": "Khoor"}).([]string)
	if len(br) != 25 || br[22] != "23: Hello" {
		t.Errorf("brute %v", br)
	}
	eq(t, call(t, "rot13", map[string]any{"text": "Hello"}), "Uryyb")
	eq(t, call(t, "rot13", map[string]any{"text": "Uryyb"}), "Hello")
	eq(t, call(t, "rot5", map[string]any{"text": "a2024"}), "a7579")
	eq(t, call(t, "rot18", map[string]any{"text": "Hello 2024"}), "Uryyb 7579")
	eq(t, call(t, "rot47", map[string]any{"text": "Hello"}), "w6==@")
	eq(t, call(t, "rot47", map[string]any{"text": "w6==@"}), "Hello")
	eq(t, call(t, "atbash", map[string]any{"text": "Wizard"}), "Draziw")
	eq(t, call(t, "vigenere_encrypt", map[string]any{"text": "ATTACKATDAWN", "key": "LEMON"}), "LXFOPVEFRNHR")
	eq(t, call(t, "vigenere_decrypt", map[string]any{"text": "LXFOPVEFRNHR", "key": "LEMON"}), "ATTACKATDAWN")
	eq(t, call(t, "vigenere_encrypt", map[string]any{"text": "Attack at dawn!", "key": "Lemon"}), "Lxfopv ef rnhr!")
	// Autokey (Wikipedia): attack at dawn / queenly.
	eq(t, call(t, "autokey_encrypt", map[string]any{"text": "ATTACKATDAWN", "key": "QUEENLY"}), "QNXEPVYTWTWP")
	eq(t, call(t, "autokey_decrypt", map[string]any{"text": "QNXEPVYTWTWP", "key": "QUEENLY"}), "ATTACKATDAWN")
	b := call(t, "beaufort", map[string]any{"text": "Attack at dawn", "key": "LEMON"}).(string)
	eq(t, call(t, "beaufort", map[string]any{"text": b, "key": "LEMON"}), "Attack at dawn")
	eq(t, call(t, "affine_encrypt", map[string]any{"text": "AFFINE CIPHER", "a": int64(5), "b": int64(8)}), "IHHWVC SWFRCP")
	eq(t, call(t, "affine_decrypt", map[string]any{"text": "IHHWVC SWFRCP", "a": int64(5), "b": int64(8)}), "AFFINE CIPHER")
	callErr(t, "affine_encrypt", map[string]any{"text": "x", "a": int64(2), "b": int64(1)})
	callErr(t, "affine_encrypt", map[string]any{"text": "x", "a": int64(13), "b": int64(1)})
	alpha := "QWERTYUIOPASDFGHJKLZXCVBNM"
	s := call(t, "substitution_encrypt", map[string]any{"text": "Hello", "key": alpha}).(string)
	eq(t, s, "Itssg")
	eq(t, call(t, "substitution_decrypt", map[string]any{"text": s, "key": alpha}), "Hello")
	callErr(t, "substitution_encrypt", map[string]any{"text": "x", "key": "ABC"})
	callErr(t, "substitution_encrypt", map[string]any{"text": "x", "key": strings.Repeat("A", 26)})
	eq(t, call(t, "rail_fence_encrypt", map[string]any{"text": "WEAREDISCOVEREDFLEEATONCE", "rails": int64(3)}), "WECRLTEERDSOEEFEAOCAIVDEN")
	eq(t, call(t, "rail_fence_decrypt", map[string]any{"text": "WECRLTEERDSOEEFEAOCAIVDEN", "rails": int64(3)}), "WEAREDISCOVEREDFLEEATONCE")
	callErr(t, "rail_fence_encrypt", map[string]any{"text": "x", "rails": int64(1)})
	// Playfair (Wikipedia example).
	eq(t, call(t, "playfair_encrypt", map[string]any{"text": "Hide the gold in the tree stump", "key": "Playfair example"}), "BMODZBXDNABEKUDMUIXMMOUVIF")
	eq(t, call(t, "playfair_decrypt", map[string]any{"text": "BMODZBXDNABEKUDMUIXMMOUVIF", "key": "Playfair example"}), "HIDETHEGOLDINTHETREXESTUMP")
	callErr(t, "playfair_decrypt", map[string]any{"text": "ABC", "key": "k"})
	callErr(t, "playfair_encrypt", map[string]any{"text": "abc", "key": "123"})
	eq(t, call(t, "polybius_encrypt", map[string]any{"text": "Hello"}), "23 15 31 31 34")
	eq(t, call(t, "polybius_decrypt", map[string]any{"text": "23 15 31 31 34"}), "HELLO")
	eq(t, call(t, "polybius_decrypt", map[string]any{"text": call(t, "polybius_encrypt", map[string]any{"text": "keyed", "key": "secret"}).(string), "key": "secret"}), "KEYED")
	callErr(t, "polybius_decrypt", map[string]any{"text": "123"})
	eq(t, call(t, "bacon_encrypt", map[string]any{"text": "Hello"}), "AABBB AABAA ABABB ABABB ABBBA")
	eq(t, call(t, "bacon_decrypt", map[string]any{"text": "AABBB AABAA ABABB ABABB ABBBA"}), "HELLO")
	cl := call(t, "bacon_encrypt", map[string]any{"text": "Quiz Jump Vue", "variant": "classic"}).(string)
	eq(t, call(t, "bacon_decrypt", map[string]any{"text": cl, "variant": "classic"}), "QUIZIUMPUUE")
	callErr(t, "bacon_decrypt", map[string]any{"text": "AAB"})
	callErr(t, "bacon_encrypt", map[string]any{"text": "x", "variant": "weird"})
	// Transpositions round-trip on awkward lengths and unicode.
	for _, in := range []string{"", "a", "ab", "Hello, wide world: 日本語!", strings.Repeat("xyz", 17)} {
		for _, k := range []string{"ZEBRAS", "a", "key"} {
			enc := call(t, "columnar_encrypt", map[string]any{"text": in, "key": k}).(string)
			eq(t, call(t, "columnar_decrypt", map[string]any{"text": enc, "key": k}), in)
		}
		for d := int64(1); d < 7; d++ {
			enc := call(t, "scytale_encrypt", map[string]any{"text": in, "diameter": d}).(string)
			eq(t, call(t, "scytale_decrypt", map[string]any{"text": enc, "diameter": d}), in)
			enc = call(t, "rail_fence_encrypt", map[string]any{"text": in, "rails": d + 1}).(string)
			eq(t, call(t, "rail_fence_decrypt", map[string]any{"text": enc, "rails": d + 1}), in)
		}
	}
	callErr(t, "columnar_encrypt", map[string]any{"text": "x", "key": ""})
	callErr(t, "vigenere_encrypt", map[string]any{"text": "x", "key": "123"})
	callErr(t, "scytale_encrypt", map[string]any{"text": "x", "diameter": int64(0)})
	all := call(t, "classic_all", map[string]any{"text": "Hello"}).(map[string]string)
	if all["rot13"] != "Uryyb" || all["atbash"] != "Svool" || all["caesar_1"] != "Ifmmp" || all["reverse"] != "olleH" || len(all) != 9 {
		t.Errorf("classic_all %v", all)
	}
	// Vigenere/affine/caesar round trips with mixed content.
	txt := "The Quick Brown Fox, 123 jumps! Over the lazy dog."
	enc := call(t, "vigenere_encrypt", map[string]any{"text": txt, "key": "Secret Key"}).(string)
	eq(t, call(t, "vigenere_decrypt", map[string]any{"text": enc, "key": "Secret Key"}), txt)
	enc = call(t, "autokey_encrypt", map[string]any{"text": txt, "key": "k"}).(string)
	eq(t, call(t, "autokey_decrypt", map[string]any{"text": enc, "key": "k"}), txt)
	for _, a := range []int64{1, 3, 5, 7, 9, 11, 15, 17, 19, 21, 23, 25, -1} {
		enc = call(t, "affine_encrypt", map[string]any{"text": txt, "a": a, "b": int64(-7)}).(string)
		eq(t, call(t, "affine_decrypt", map[string]any{"text": enc, "a": a, "b": int64(-7)}), txt)
	}
}

func TestAsymmetric(t *testing.T) {
	kp := call(t, "rsa_generate", map[string]any{"bits": int64(2048)}).(map[string]string)
	pub, priv := kp["public_pem"], kp["private_pem"]
	if !strings.Contains(pub, "BEGIN PUBLIC KEY") || !strings.Contains(priv, "BEGIN PRIVATE KEY") || kp["bits"] != "2048" {
		t.Fatalf("rsa keys: %v", kp)
	}
	callErr(t, "rsa_generate", map[string]any{"bits": int64(1024)})
	callErr(t, "rsa_generate", map[string]any{"bits": int64(8192)})
	ct := call(t, "rsa_encrypt", map[string]any{"text": "attack at dawn", "public_pem": pub}).(string)
	eq(t, call(t, "rsa_decrypt", map[string]any{"data": ct, "private_pem": priv}), "attack at dawn")
	callErr(t, "rsa_decrypt", map[string]any{"data": ct, "private_pem": priv, "hash": "sha512"})
	other := call(t, "rsa_generate", nil).(map[string]string)
	callErr(t, "rsa_decrypt", map[string]any{"data": ct, "private_pem": other["private_pem"]})
	callErr(t, "rsa_encrypt", map[string]any{"text": strings.Repeat("x", 300), "public_pem": pub})
	callErr(t, "rsa_encrypt", map[string]any{"text": "x", "public_pem": "nonsense"})
	callErr(t, "rsa_decrypt", map[string]any{"data": ct, "private_pem": pub})
	for _, pad := range []string{"pss", "pkcs1v15"} {
		for _, h := range []string{"sha256", "sha512"} {
			sig := call(t, "rsa_sign", map[string]any{"text": "msg", "private_pem": priv, "padding": pad, "hash": h}).(string)
			if call(t, "rsa_verify", map[string]any{"text": "msg", "signature": sig, "public_pem": pub, "padding": pad, "hash": h}) != true {
				t.Errorf("verify %s %s", pad, h)
			}
			if call(t, "rsa_verify", map[string]any{"text": "msg2", "signature": sig, "public_pem": pub, "padding": pad, "hash": h}) != false {
				t.Errorf("verify tampered %s %s", pad, h)
			}
			if call(t, "rsa_verify", map[string]any{"text": "msg", "signature": sig, "public_pem": other["public_pem"], "padding": pad, "hash": h}) != false {
				t.Errorf("verify wrong key %s %s", pad, h)
			}
		}
	}
	callErr(t, "rsa_sign", map[string]any{"text": "m", "private_pem": priv, "padding": "weird"})

	// Ed25519, RFC 8032 test 1 and 2.
	seed := "9d61b19deffd5a60ba844af492ec2cc44449c5697b326919703bac031cae7f60"
	eq(t, call(t, "ed25519_public_key", map[string]any{"private_key": seed}), "d75a980182b10ab7d54bfed3c964073a0ee172f3daa62325af021a68f707511a")
	sig := call(t, "ed25519_sign", map[string]any{"text": "", "private_key": seed})
	eq(t, sig, "e5564300c360ac729086e2cc806e828a84877f1eb8e5d974d873e065224901555fb8821590a33bacc61e39701cf9b46bd25bf5f0595bbe24655141438e7a100b")
	pk := "d75a980182b10ab7d54bfed3c964073a0ee172f3daa62325af021a68f707511a"
	if call(t, "ed25519_verify", map[string]any{"text": "", "signature": sig, "public_key": pk}) != true {
		t.Error("ed25519 verify")
	}
	if call(t, "ed25519_verify", map[string]any{"text": "x", "signature": sig, "public_key": pk}) != false {
		t.Error("ed25519 verify tampered")
	}
	if call(t, "ed25519_verify", map[string]any{"text": "", "signature": "00", "public_key": pk}) != false {
		t.Error("ed25519 verify short sig")
	}
	callErr(t, "ed25519_verify", map[string]any{"text": "", "signature": sig, "public_key": "abcd"})
	callErr(t, "ed25519_sign", map[string]any{"text": "", "private_key": "abcd"})
	ek := call(t, "ed25519_generate", nil).(map[string]string)
	s2 := call(t, "ed25519_sign", map[string]any{"text": "hi", "private_key": ek["private_key"]}).(string)
	if call(t, "ed25519_verify", map[string]any{"text": "hi", "signature": s2, "public_key": ek["public_key"]}) != true ||
		call(t, "ed25519_verify", map[string]any{"text": "hi", "signature": s2, "public_key": ek["public_pem"]}) != true {
		t.Error("ed25519 generated pair")
	}
	s3 := call(t, "ed25519_sign", map[string]any{"text": "hi", "private_key": ek["private_pem"]}).(string)
	if s3 != s2 {
		t.Error("ed25519 PEM key must sign identically")
	}

	for _, curve := range []string{"p256", "p384", "p521"} {
		ec := call(t, "ecdsa_generate", map[string]any{"curve": curve}).(map[string]string)
		sg := call(t, "ecdsa_sign", map[string]any{"text": "hi", "private_pem": ec["private_pem"]}).(string)
		if call(t, "ecdsa_verify", map[string]any{"text": "hi", "signature": sg, "public_pem": ec["public_pem"]}) != true {
			t.Errorf("ecdsa %s", curve)
		}
		if call(t, "ecdsa_verify", map[string]any{"text": "ho", "signature": sg, "public_pem": ec["public_pem"]}) != false {
			t.Errorf("ecdsa tampered %s", curve)
		}
	}
	callErr(t, "ecdsa_generate", map[string]any{"curve": "p111"})
	callErr(t, "ecdsa_sign", map[string]any{"text": "x", "private_pem": priv})

	// X25519, RFC 7748 section 6.1.
	ash := "77076d0a7318a57d3c16c17251b26645df4c2f87ebc0992ab177fba51db92c2a"
	bpub := "de9edb7d7b7dc1b4d35b61c2ece435373f8343c85b78674dadfc7e146f882b4f"
	eq(t, call(t, "x25519_shared_secret", map[string]any{"private_key": ash, "peer_public_key": bpub}), "4a5d9d5ba4ce2de1728e3bf480350f25e07e21c947d19e3376f09b3c1e161742")
	callErr(t, "x25519_shared_secret", map[string]any{"private_key": ash, "peer_public_key": strings.Repeat("00", 32)})
	callErr(t, "x25519_shared_secret", map[string]any{"private_key": "ab", "peer_public_key": bpub})
	xa := call(t, "x25519_generate", nil).(map[string]string)
	xb := call(t, "x25519_generate", nil).(map[string]string)
	if call(t, "x25519_shared_secret", map[string]any{"private_key": xa["private_key"], "peer_public_key": xb["public_key"]}) !=
		call(t, "x25519_shared_secret", map[string]any{"private_key": xb["private_key"], "peer_public_key": xa["public_key"]}) {
		t.Error("x25519 shared secrets differ")
	}

	// PEM helpers.
	d := call(t, "pem_decode", map[string]any{"pem": pub}).(map[string]string)
	if d["type"] != "PUBLIC KEY" || d["count"] != "1" || d["bytes_hex"] == "" {
		t.Errorf("pem_decode %v", d)
	}
	re := call(t, "pem_encode", map[string]any{"type": d["type"], "hex": d["bytes_hex"]}).(string)
	if strings.TrimSpace(re) != strings.TrimSpace(pub) {
		t.Errorf("pem round trip: %s vs %s", re, pub)
	}
	callErr(t, "pem_decode", map[string]any{"pem": "not pem"})
	callErr(t, "pem_encode", map[string]any{"type": "", "hex": "00"})
	callErr(t, "pem_encode", map[string]any{"type": "X", "hex": "0"})
}

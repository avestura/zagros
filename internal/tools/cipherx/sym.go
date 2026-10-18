package cipherx

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rc4"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/blowfish"
	"golang.org/x/crypto/chacha20"
	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/scrypt"

	"github.com/avestura/zagros/internal/tools"
)

// Passphrase mode: scrypt(N=32768, r=8, p=1) with a random 16-byte salt.
const (
	passSaltLen = 16
	passScryptN = 1 << 15
	envPrefix   = "zg1."
	macLen      = 32
	maxPlain    = 64 << 20
)

var errAuth = errors.New("decryption failed: wrong key or corrupted data")

// scheme describes one symmetric cipher and how it is wrapped.
type scheme struct {
	id      string // generic algo name, e.g. "aes-gcm"
	fn      string // function prefix, e.g. "aes_gcm"
	title   string
	tags    []string
	note    string
	keyOK   func(n int) bool
	keyDesc string
	defPass int // passphrase-mode key length
	nonce   int
	aead    bool
	aes     bool // honours the bits parameter
	label   func(klen int) string
	seal    func(key, nonce, pt, aad []byte) ([]byte, error)
	open    func(key, nonce, ct, aad []byte) ([]byte, error)
}

func pkcs7Pad(b []byte, bs int) []byte {
	n := bs - len(b)%bs
	out := append(append([]byte(nil), b...), make([]byte, n)...)
	for i := len(b); i < len(out); i++ {
		out[i] = byte(n)
	}
	return out
}

func pkcs7Unpad(b []byte, bs int) ([]byte, error) {
	if len(b) == 0 || len(b)%bs != 0 {
		return nil, errors.New("decryption failed: ciphertext is not a whole number of blocks")
	}
	n := int(b[len(b)-1])
	if n == 0 || n > bs || n > len(b) {
		return nil, errors.New("decryption failed: invalid padding (wrong key or corrupted data)")
	}
	for _, c := range b[len(b)-n:] {
		if int(c) != n {
			return nil, errors.New("decryption failed: invalid padding (wrong key or corrupted data)")
		}
	}
	return b[:len(b)-n], nil
}

func cbcSealOpen(nb func([]byte) (cipher.Block, error)) (seal, open func(key, nonce, in, aad []byte) ([]byte, error)) {
	seal = func(key, iv, pt, _ []byte) ([]byte, error) {
		b, err := nb(key)
		if err != nil {
			return nil, err
		}
		p := pkcs7Pad(pt, b.BlockSize())
		cipher.NewCBCEncrypter(b, iv).CryptBlocks(p, p)
		return p, nil
	}
	open = func(key, iv, ct, _ []byte) ([]byte, error) {
		b, err := nb(key)
		if err != nil {
			return nil, err
		}
		if len(ct) == 0 || len(ct)%b.BlockSize() != 0 {
			return nil, errors.New("decryption failed: ciphertext is not a whole number of blocks")
		}
		p := make([]byte, len(ct))
		cipher.NewCBCDecrypter(b, iv).CryptBlocks(p, ct)
		return pkcs7Unpad(p, b.BlockSize())
	}
	return
}

func streamSealOpen(mk func(b cipher.Block, iv []byte, enc bool) cipher.Stream) (seal, open func(key, nonce, in, aad []byte) ([]byte, error)) {
	run := func(enc bool) func(key, iv, in, _ []byte) ([]byte, error) {
		return func(key, iv, in, _ []byte) ([]byte, error) {
			b, err := aes.NewCipher(key)
			if err != nil {
				return nil, err
			}
			out := make([]byte, len(in))
			mk(b, iv, enc).XORKeyStream(out, in)
			return out, nil
		}
	}
	return run(true), run(false)
}

func ecbSeal(key, _, pt, _ []byte) ([]byte, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	p := pkcs7Pad(pt, 16)
	for i := 0; i < len(p); i += 16 {
		b.Encrypt(p[i:i+16], p[i:i+16])
	}
	return p, nil
}

func ecbOpen(key, _, ct, _ []byte) ([]byte, error) {
	b, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(ct) == 0 || len(ct)%16 != 0 {
		return nil, errors.New("decryption failed: ciphertext is not a whole number of blocks")
	}
	p := make([]byte, len(ct))
	for i := 0; i < len(ct); i += 16 {
		b.Decrypt(p[i:i+16], ct[i:i+16])
	}
	return pkcs7Unpad(p, 16)
}

func aeadSealOpen(mk func(key []byte) (cipher.AEAD, error)) (seal, open func(key, nonce, in, aad []byte) ([]byte, error)) {
	seal = func(key, nonce, pt, aad []byte) ([]byte, error) {
		a, err := mk(key)
		if err != nil {
			return nil, err
		}
		return a.Seal(nil, nonce, pt, aad), nil
	}
	open = func(key, nonce, ct, aad []byte) ([]byte, error) {
		a, err := mk(key)
		if err != nil {
			return nil, err
		}
		if len(ct) < a.Overhead() {
			return nil, errors.New("decryption failed: data is too short")
		}
		p, err := a.Open(nil, nonce, ct, aad)
		if err != nil {
			return nil, errAuth
		}
		return p, nil
	}
	return
}

func aesLabel(suffix string) func(int) string {
	return func(k int) string { return fmt.Sprintf("aes%d%s", k*8, suffix) }
}
func fixedLabel(s string) func(int) string { return func(int) string { return s } }

func aesKeyOK(n int) bool { return n == 16 || n == 24 || n == 32 }

func tdesKey(key []byte) []byte {
	if len(key) == 16 {
		return append(append([]byte(nil), key...), key[:8]...)
	}
	return key
}

var schemes = buildSchemes()

func buildSchemes() []scheme {
	gcmSeal, gcmOpen := aeadSealOpen(func(k []byte) (cipher.AEAD, error) {
		b, err := aes.NewCipher(k)
		if err != nil {
			return nil, err
		}
		return cipher.NewGCM(b)
	})
	cbcSeal, cbcOpen := cbcSealOpen(func(k []byte) (cipher.Block, error) { return aes.NewCipher(k) })
	ctrSeal, ctrOpen := streamSealOpen(func(b cipher.Block, iv []byte, _ bool) cipher.Stream { return cipher.NewCTR(b, iv) })
	cfbSeal, cfbOpen := streamSealOpen(func(b cipher.Block, iv []byte, enc bool) cipher.Stream {
		if enc {
			return cipher.NewCFBEncrypter(b, iv)
		}
		return cipher.NewCFBDecrypter(b, iv)
	})
	ofbSeal, ofbOpen := streamSealOpen(func(b cipher.Block, iv []byte, _ bool) cipher.Stream { return cipher.NewOFB(b, iv) })
	ccSeal, ccOpen := aeadSealOpen(func(k []byte) (cipher.AEAD, error) { return chacha20poly1305.New(k) })
	xcSeal, xcOpen := aeadSealOpen(func(k []byte) (cipher.AEAD, error) { return chacha20poly1305.NewX(k) })
	tdSeal, tdOpen := cbcSealOpen(func(k []byte) (cipher.Block, error) { return des.NewTripleDESCipher(tdesKey(k)) })
	bfSeal, bfOpen := cbcSealOpen(func(k []byte) (cipher.Block, error) { return blowfish.NewCipher(k) })
	c20 := func(key, nonce, in, _ []byte) ([]byte, error) {
		c, err := chacha20.NewUnauthenticatedCipher(key, nonce)
		if err != nil {
			return nil, err
		}
		out := make([]byte, len(in))
		c.XORKeyStream(out, in)
		return out, nil
	}
	rc4f := func(key, _, in, _ []byte) ([]byte, error) {
		c, err := rc4.NewCipher(key)
		if err != nil {
			return nil, err
		}
		out := make([]byte, len(in))
		c.XORKeyStream(out, in)
		return out, nil
	}
	aesNote := "AES key must be 16, 24 or 32 bytes."
	return []scheme{
		{id: "aes-gcm", fn: "aes_gcm", title: "AES-GCM", aead: true, aes: true, nonce: 12, defPass: 32, keyOK: aesKeyOK, keyDesc: aesNote,
			label: aesLabel("gcm"), seal: gcmSeal, open: gcmOpen,
			tags: []string{"aes", "gcm", "aead", "authenticated encryption", "nist", "recommended", "aes-256-gcm", "encrypt text"},
			note: "Authenticated (AEAD): a wrong key or tampered data fails with an error. 12-byte nonce; never reuse a nonce with the same key."},
		{id: "aes-cbc", fn: "aes_cbc", title: "AES-CBC", nonce: 16, defPass: 32, keyOK: aesKeyOK, keyDesc: aesNote,
			label: aesLabel("cbc"), seal: cbcSeal, open: cbcOpen, aes: true,
			tags: []string{"aes", "cbc", "pkcs7", "block cipher", "aes-256-cbc", "openssl"},
			note: "PKCS#7 padding, 16-byte IV. Unauthenticated in raw-key mode (tampering is not detected; a wrong key usually fails with a padding error); passphrase envelopes add an HMAC-SHA256 tag."},
		{id: "aes-ctr", fn: "aes_ctr", title: "AES-CTR", nonce: 16, defPass: 32, keyOK: aesKeyOK, keyDesc: aesNote,
			label: aesLabel("ctr"), seal: ctrSeal, open: ctrOpen, aes: true,
			tags: []string{"aes", "ctr", "counter mode", "stream", "aes-256-ctr"},
			note: "Stream mode, 16-byte initial counter block, no padding. Unauthenticated in raw-key mode; never reuse an IV with the same key; passphrase envelopes add an HMAC-SHA256 tag."},
		{id: "aes-cfb", fn: "aes_cfb", title: "AES-CFB", nonce: 16, defPass: 32, keyOK: aesKeyOK, keyDesc: aesNote,
			label: aesLabel("cfb"), seal: cfbSeal, open: cfbOpen, aes: true,
			tags: []string{"aes", "cfb", "cfb128", "stream", "aes-256-cfb"},
			note: "CFB-128 stream mode, 16-byte IV, no padding. Unauthenticated in raw-key mode; passphrase envelopes add an HMAC-SHA256 tag."},
		{id: "aes-ofb", fn: "aes_ofb", title: "AES-OFB", nonce: 16, defPass: 32, keyOK: aesKeyOK, keyDesc: aesNote,
			label: aesLabel("ofb"), seal: ofbSeal, open: ofbOpen, aes: true,
			tags: []string{"aes", "ofb", "stream", "aes-256-ofb"},
			note: "OFB stream mode, 16-byte IV, no padding. Unauthenticated in raw-key mode; passphrase envelopes add an HMAC-SHA256 tag."},
		{id: "aes-ecb", fn: "aes_ecb", title: "AES-ECB (insecure)", nonce: 0, defPass: 32, keyOK: aesKeyOK, keyDesc: aesNote,
			label: aesLabel("ecb"), seal: ecbSeal, open: ecbOpen, aes: true,
			tags: []string{"aes", "ecb", "insecure", "no iv", "legacy", "aes-128-ecb", "penguin"},
			note: "INSECURE: identical plaintext blocks give identical ciphertext blocks and there is no IV. Supported only for compatibility and puzzles; PKCS#7 padding. Unauthenticated in raw-key mode."},
		{id: "chacha20-poly1305", fn: "chacha20_poly1305", title: "ChaCha20-Poly1305", aead: true, nonce: 12, defPass: 32,
			keyOK: func(n int) bool { return n == 32 }, keyDesc: "ChaCha20-Poly1305 key must be 32 bytes.",
			label: fixedLabel("chacha20poly1305"), seal: ccSeal, open: ccOpen,
			tags: []string{"chacha20", "poly1305", "aead", "rfc8439", "authenticated encryption", "chacha"},
			note: "Authenticated (AEAD, RFC 8439): a wrong key or tampered data fails with an error. 12-byte nonce; never reuse a nonce with the same key."},
		{id: "xchacha20-poly1305", fn: "xchacha20_poly1305", title: "XChaCha20-Poly1305", aead: true, nonce: 24, defPass: 32,
			keyOK: func(n int) bool { return n == 32 }, keyDesc: "XChaCha20-Poly1305 key must be 32 bytes.",
			label: fixedLabel("xchacha20poly1305"), seal: xcSeal, open: xcOpen,
			tags: []string{"xchacha20", "poly1305", "aead", "extended nonce", "chacha", "authenticated encryption"},
			note: "Authenticated (AEAD) with a 24-byte nonce that is safe to pick at random. A wrong key or tampered data fails with an error."},
		{id: "chacha20", fn: "chacha20", title: "ChaCha20 stream", nonce: 12, defPass: 32,
			keyOK: func(n int) bool { return n == 32 }, keyDesc: "ChaCha20 key must be 32 bytes.",
			label: fixedLabel("chacha20"), seal: c20, open: c20,
			tags: []string{"chacha20", "stream cipher", "rfc8439", "raw"},
			note: "Raw ChaCha20 keystream XOR (12-byte nonce, counter starts at 0), no authentication in raw-key mode; passphrase envelopes add an HMAC-SHA256 tag."},
		{id: "des3-cbc", fn: "des3_cbc", title: "3DES-CBC (legacy)", nonce: 8, defPass: 24,
			keyOK: func(n int) bool { return n == 16 || n == 24 }, keyDesc: "Triple-DES key must be 16 or 24 bytes.",
			label: fixedLabel("des3cbc"), seal: tdSeal, open: tdOpen,
			tags: []string{"3des", "tdes", "triple des", "des-ede3", "legacy", "des3", "cbc"},
			note: "Legacy Triple-DES in CBC mode (PKCS#7 padding, 8-byte IV), kept for old systems; 16-byte keys mean K1,K2,K1. Unauthenticated in raw-key mode; passphrase envelopes add an HMAC-SHA256 tag."},
		{id: "blowfish-cbc", fn: "blowfish_cbc", title: "Blowfish-CBC (legacy)", nonce: 8, defPass: 32,
			keyOK: func(n int) bool { return n >= 4 && n <= 56 }, keyDesc: "Blowfish key must be 4 to 56 bytes.",
			label: fixedLabel("blowfishcbc"), seal: bfSeal, open: bfOpen,
			tags: []string{"blowfish", "legacy", "cbc", "bruce schneier"},
			note: "Legacy Blowfish in CBC mode (PKCS#7 padding, 8-byte IV). Unauthenticated in raw-key mode; passphrase envelopes add an HMAC-SHA256 tag."},
		{id: "rc4", fn: "rc4", title: "RC4 (broken)", nonce: 0, defPass: 32,
			keyOK: func(n int) bool { return n >= 1 && n <= 256 }, keyDesc: "RC4 key must be 1 to 256 bytes.",
			label: fixedLabel("rc4"), seal: rc4f, open: rc4f,
			tags: []string{"rc4", "arcfour", "arc4", "legacy", "broken", "stream cipher", "wep"},
			note: "BROKEN stream cipher with no IV, kept only for compatibility. Unauthenticated in raw-key mode; passphrase envelopes add an HMAC-SHA256 tag."},
	}
}

var labelIndex = func() map[string]struct {
	s    scheme
	klen int
} {
	m := map[string]struct {
		s    scheme
		klen int
	}{}
	for _, s := range schemes {
		lens := []int{s.defPass}
		if s.aes {
			lens = []int{16, 24, 32}
		}
		for _, l := range lens {
			m[s.label(l)] = struct {
				s    scheme
				klen int
			}{s, l}
		}
	}
	return m
}()

func findScheme(name string) (scheme, error) {
	n := normAlgo(name)
	aliases := map[string]string{
		"aes": "aes_gcm", "aes_256_gcm": "aes_gcm", "aes_128_gcm": "aes_gcm", "gcm": "aes_gcm", "aes256gcm": "aes_gcm",
		"aes_256_cbc": "aes_cbc", "aes_128_cbc": "aes_cbc", "cbc": "aes_cbc", "aes_256_ctr": "aes_ctr", "aes_128_ctr": "aes_ctr",
		"ctr": "aes_ctr", "cfb": "aes_cfb", "ofb": "aes_ofb", "ecb": "aes_ecb", "aes_128_ecb": "aes_ecb",
		"chacha20poly1305": "chacha20_poly1305", "chacha": "chacha20_poly1305", "xchacha20poly1305": "xchacha20_poly1305",
		"xchacha": "xchacha20_poly1305", "3des": "des3_cbc", "3des_cbc": "des3_cbc", "tdes": "des3_cbc", "triple_des": "des3_cbc",
		"des3": "des3_cbc", "blowfish": "blowfish_cbc", "arcfour": "rc4",
	}
	if a, ok := aliases[n]; ok {
		n = a
	}
	for _, s := range schemes {
		if s.fn == n {
			return s, nil
		}
	}
	return scheme{}, fmt.Errorf("unknown cipher %q (see cipher.algorithms())", name)
}

// opts carries the shared encrypt/decrypt options.
type opts struct {
	key    string
	keyEnc string
	bits   int64
	iv     []byte
	hasIV  bool
	aad    []byte
}

func randBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	return b, err
}

func isPassphrase(enc string) bool {
	e := strings.ToLower(strings.TrimSpace(enc))
	return e == "" || e == "passphrase" || e == "password"
}

func macOver(mk []byte, label string, salt, nonce, ct []byte) []byte {
	h := hmac.New(sha256.New, mk)
	h.Write([]byte(label))
	h.Write(salt)
	h.Write(nonce)
	h.Write(ct)
	return h.Sum(nil)
}

func (s scheme) checkKey(k []byte) error {
	if !s.keyOK(len(k)) {
		return fmt.Errorf("%s (got %d bytes); use key_encoding=passphrase to derive a key from a text passphrase", s.keyDesc, len(k))
	}
	return nil
}

// encrypt returns either the passphrase envelope or the raw payload.
func (s scheme) encrypt(pt []byte, o opts, output string) (string, error) {
	if len(pt) > maxPlain {
		return "", fmt.Errorf("input is larger than %d MiB", maxPlain>>20)
	}
	if o.hasIV && len(o.iv) != s.nonce {
		return "", fmt.Errorf("%s needs an iv/nonce of %d bytes (got %d)", s.id, s.nonce, len(o.iv))
	}
	if isPassphrase(o.keyEnc) {
		if o.key == "" {
			return "", errors.New("passphrase is empty")
		}
		klen := s.defPass
		if s.aes && o.bits != 0 {
			if o.bits != 128 && o.bits != 192 && o.bits != 256 {
				return "", fmt.Errorf("bits must be 128, 192 or 256 (got %d)", o.bits)
			}
			klen = int(o.bits / 8)
		}
		salt, err := randBytes(passSaltLen)
		if err != nil {
			return "", err
		}
		nonce := o.iv
		if !o.hasIV {
			if nonce, err = randBytes(s.nonce); err != nil {
				return "", err
			}
		}
		km, err := scrypt.Key([]byte(o.key), salt, passScryptN, 8, 1, klen+macLen)
		if err != nil {
			return "", err
		}
		ct, err := s.seal(km[:klen], nonce, pt, o.aad)
		if err != nil {
			return "", err
		}
		label := s.label(klen)
		if !s.aead {
			ct = append(ct, macOver(km[klen:], label, salt, nonce, ct)...)
		}
		payload := append(append(append([]byte(nil), salt...), nonce...), ct...)
		return envPrefix + label + "." + base64.RawURLEncoding.EncodeToString(payload), nil
	}
	key, err := decodeBytes(o.key, o.keyEnc)
	if err != nil {
		return "", fmt.Errorf("key: %v", err)
	}
	if err := s.checkKey(key); err != nil {
		return "", err
	}
	if s.aes && o.bits != 0 && int(o.bits) != len(key)*8 {
		return "", fmt.Errorf("bits=%d does not match the %d-byte key", o.bits, len(key))
	}
	nonce := o.iv
	if !o.hasIV {
		if nonce, err = randBytes(s.nonce); err != nil {
			return "", err
		}
	}
	ct, err := s.seal(key, nonce, pt, o.aad)
	if err != nil {
		return "", err
	}
	return encodeBytes(append(append([]byte(nil), nonce...), ct...), output)
}

func isEnvelope(data string) bool { return strings.HasPrefix(strings.TrimSpace(data), envPrefix) }

// decryptAny decrypts an envelope or raw payload. want may be a zero scheme
// (matched against the envelope label) in generic decrypt.
func decrypt(want *scheme, data string, dataEnc string, o opts) ([]byte, error) {
	data = strings.TrimSpace(data)
	if isEnvelope(data) {
		if !isPassphrase(o.keyEnc) {
			return nil, errors.New("this is a zg1 passphrase envelope: use key_encoding=passphrase")
		}
		parts := strings.SplitN(data[len(envPrefix):], ".", 2)
		if len(parts) != 2 {
			return nil, errors.New("malformed zg1 envelope (expected zg1.<algo>.<data>)")
		}
		li, ok := labelIndex[parts[0]]
		if !ok {
			return nil, fmt.Errorf("unknown zg1 envelope algorithm %q", parts[0])
		}
		s := li.s
		if want != nil && want.id != s.id {
			return nil, fmt.Errorf("envelope holds %s data, not %s (use decrypt or %s_decrypt)", s.id, want.id, s.fn)
		}
		payload, err := base64.RawURLEncoding.DecodeString(strings.TrimRight(parts[1], "="))
		if err != nil {
			return nil, errors.New("malformed zg1 envelope: bad base64url data")
		}
		min := passSaltLen + s.nonce
		if !s.aead {
			min += macLen
		}
		if len(payload) < min {
			return nil, errors.New("malformed zg1 envelope: data is too short")
		}
		if o.key == "" {
			return nil, errors.New("passphrase is empty")
		}
		salt := payload[:passSaltLen]
		nonce := payload[passSaltLen : passSaltLen+s.nonce]
		ct := payload[passSaltLen+s.nonce:]
		km, err := scrypt.Key([]byte(o.key), salt, passScryptN, 8, 1, li.klen+macLen)
		if err != nil {
			return nil, err
		}
		if !s.aead {
			tag := ct[len(ct)-macLen:]
			ct = ct[:len(ct)-macLen]
			if !hmac.Equal(tag, macOver(km[li.klen:], parts[0], salt, nonce, ct)) {
				return nil, errAuth
			}
		}
		return s.open(km[:li.klen], nonce, ct, o.aad)
	}
	if want == nil {
		return nil, errors.New("data is not a zg1 envelope, so the algorithm cannot be detected: pass algo")
	}
	s := *want
	if isPassphrase(o.keyEnc) {
		return nil, errors.New("data is not a zg1 passphrase envelope: for a raw key set key_encoding to utf-8, hex or base64")
	}
	key, err := decodeBytes(o.key, o.keyEnc)
	if err != nil {
		return nil, fmt.Errorf("key: %v", err)
	}
	if err := s.checkKey(key); err != nil {
		return nil, err
	}
	raw, err := decodeBytes(data, dataEnc)
	if err != nil {
		return nil, fmt.Errorf("data: %v", err)
	}
	nonce := o.iv
	ct := raw
	if o.hasIV {
		if len(nonce) != s.nonce {
			return nil, fmt.Errorf("%s needs an iv/nonce of %d bytes (got %d)", s.id, s.nonce, len(nonce))
		}
	} else {
		if len(raw) < s.nonce {
			return nil, errors.New("decryption failed: data is shorter than the iv/nonce")
		}
		nonce, ct = raw[:s.nonce], raw[s.nonce:]
	}
	return s.open(key, nonce, ct, o.aad)
}

// looksEncrypted is the Accepts check of decryptors.
func looksEncrypted(in string) bool {
	t := strings.TrimSpace(in)
	if isEnvelope(t) {
		return true
	}
	if len(t) < 16 {
		return false
	}
	if _, err := decodeBytes(t, "hex"); err == nil && len(t)%2 == 0 {
		return true
	}
	_, err := decodeB64(t)
	return err == nil
}

func parseOpts(a tools.Args, hasBits, hasAAD bool) (opts, error) {
	o := opts{key: a.Str("key"), keyEnc: a.Str("key_encoding")}
	if hasBits {
		o.bits = a.Int("bits")
	}
	if hasAAD {
		o.aad = []byte(a.Str("aad"))
	}
	if a.Str("iv") != "" {
		iv, err := decodeBytes(a.Str("iv"), a.Str("iv_encoding"))
		if err != nil {
			return o, fmt.Errorf("iv: %v", err)
		}
		o.iv, o.hasIV = iv, true
	}
	return o, nil
}

func commonKeyParams(s scheme, generic bool) []tools.Param {
	ps := []tools.Param{
		reqStr("key", "Passphrase (key_encoding=passphrase, the default) or the raw key ("+s.keyDesc+")."),
		{Name: "key_encoding", Type: "str", Default: "passphrase", Choices: []string{"passphrase", "utf-8", "hex", "base64"},
			Doc: "passphrase (default): key is a password stretched by scrypt with a random salt, output is a zg1 envelope. utf-8, hex or base64: key is the raw key bytes in that encoding."},
	}
	if s.aes || generic {
		ps = append(ps, tools.Param{Name: "bits", Type: "int", Default: int64(0), Doc: "AES key size: 128, 192 or 256. 0 (default) means 256 for passphrases and the key's own length for raw keys. Ignored by non-AES ciphers."})
	}
	ps = append(ps,
		strParam("iv", "Optional IV/nonce, for deterministic output or interop (random by default). Encoded per iv_encoding. On decrypt with a raw key it means the data has no IV prefix.", ""),
		tools.Param{Name: "iv_encoding", Type: "str", Default: "hex", Choices: []string{"hex", "base64", "utf-8"}, Doc: "How iv is turned into bytes: hex (default), base64 or utf-8."},
	)
	if s.aead || generic {
		ps = append(ps, strParam("aad", "Optional additional authenticated data (UTF-8), used by the AEAD ciphers (GCM, ChaCha20-Poly1305) and ignored by the others; it must be the same on decrypt.", ""))
	}
	return ps
}

func encDoc(s scheme) string {
	return fmt.Sprintf("Encrypts text with %s. %s With key_encoding=passphrase (default) the key is a passphrase stretched by scrypt (N=32768) with a random 16-byte salt and the result is the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)>, which %s_decrypt reads back; with key_encoding utf-8, hex or base64 the key is raw and the result is base64(iv|ciphertext|tag) rendered per output (a random IV unless iv is given).", s.title, s.note, s.fn)
}

func decDoc(s scheme) string {
	return fmt.Sprintf("Decrypts data produced by %s_encrypt (%s). Accepts the zg1 envelope (with key_encoding=passphrase) or raw iv|ciphertext|tag data in data_encoding (default base64) with a raw key. %s A wrong key, a wrong passphrase or damaged data returns an error; the plaintext is returned as UTF-8 text unless output asks for hex or base64.", s.fn, s.title, authDoc(s))
}

func authDoc(s scheme) string {
	if s.aead {
		return "Authenticated: tampering is always detected."
	}
	return "Raw-key mode is unauthenticated (a wrong key may yield padding errors or garbage); passphrase envelopes carry an HMAC-SHA256 tag and always fail cleanly on a wrong passphrase."
}

func keywordsFor(s scheme) []string {
	kw := []string{"encrypt", "decrypt", s.fn}
	switch {
	case strings.HasPrefix(s.fn, "aes"):
		kw = append(kw, "aes")
	case strings.Contains(s.fn, "chacha"):
		kw = append(kw, "chacha")
	}
	return kw
}

func headlineOrder(fn string, dec bool) int {
	if fn == "aes_gcm" {
		if dec {
			return 1
		}
		return 0
	}
	return 0
}

func init() {
	for i, s := range schemes {
		s := s
		base := 100 + 2*i
		encOrder, decOrder := base, base+1
		if s.fn == "aes_gcm" {
			encOrder, decOrder = 0, 1
		}
		hasBits := s.aes
		hasAAD := s.aead
		reg(tools.Func{
			Name: s.fn + "_encrypt", Title: s.title + " encrypt", Doc: encDoc(s),
			Params: append(append([]tools.Param{reqStr("text", "Plaintext (UTF-8 unless input_encoding says hex or base64).")}, commonKeyParams(s, false)...),
				inEncParam(), outParam("base64", "Raw-key mode only: how iv|ciphertext|tag is rendered (base64 default). Passphrase mode always returns the zg1 envelope.")),
			Returns: "str",
			Fn: func(a tools.Args) (any, error) {
				pt, err := decodeBytes(a.Str("text"), a.Str("input_encoding"))
				if err != nil {
					return nil, err
				}
				o, err := parseOpts(a, hasBits, hasAAD)
				if err != nil {
					return nil, err
				}
				return s.encrypt(pt, o, a.Str("output"))
			},
			Example: fmt.Sprintf(`cipher.%s_encrypt("secret message", "correct horse battery staple")`, s.fn),
			Tags:    append([]string{"encrypt", "encryption", "cipher", "secret", "password protect"}, s.tags...),
			Family:  "cipher", Order: encOrder, Keywords: keywordsFor(s),
		})
		reg(tools.Func{
			Name: s.fn + "_decrypt", Title: s.title + " decrypt", Doc: decDoc(s),
			Params: append(append([]tools.Param{reqStr("data", "Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.")}, commonKeyParams(s, false)...),
				tools.Param{Name: "data_encoding", Type: "str", Default: "base64", Choices: []string{"base64", "base64url", "hex", "base32"}, Doc: "Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes."},
				outParam("utf-8", "How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32.")),
			Returns: "str", Input: "data",
			Fn: func(a tools.Args) (any, error) {
				o, err := parseOpts(a, false, hasAAD)
				if err != nil {
					return nil, err
				}
				pt, err := decrypt(&s, a.Str("data"), a.Str("data_encoding"), o)
				if err != nil {
					return nil, err
				}
				return encodeBytes(pt, a.Str("output"))
			},
			Example: fmt.Sprintf(`cipher.%s_decrypt("zg1.aes256gcm....", "correct horse battery staple")`, s.fn),
			Tags:    append([]string{"decrypt", "decryption", "cipher", "decipher"}, s.tags...),
			Family:  "cipher", Order: decOrder, Keywords: keywordsFor(s),
			Accepts: looksEncrypted,
		})
	}

	names := func() []string {
		out := make([]string, len(schemes))
		for i, s := range schemes {
			out[i] = s.fn
		}
		return out
	}
	generic := scheme{keyDesc: "key length depends on algo"}
	reg(tools.Func{
		Name: "encrypt", Title: "Encrypt with any cipher",
		Doc: "Encrypts text with the named symmetric cipher (aes_gcm, aes_cbc, aes_ctr, aes_cfb, aes_ofb, aes_ecb, chacha20_poly1305, xchacha20_poly1305, chacha20, des3_cbc, blowfish_cbc, rc4; see algorithms()). Passphrase mode (default) returns the text envelope zg1.<algo>.<base64url(salt|nonce|ciphertext|tag)> using an scrypt-derived key; raw-key mode (key_encoding utf-8, hex or base64) returns base64(iv|ciphertext|tag).",
		Params: append(append([]tools.Param{reqStr("algo", "Cipher name, e.g. aes_gcm or chacha20_poly1305."), reqStr("text", "Plaintext (UTF-8 unless input_encoding says hex or base64).")},
			commonKeyParams(generic, true)...), inEncParam(), outParam("base64", "Raw-key mode only: rendering of iv|ciphertext|tag (base64 default).")),
		Returns: "str", Input: "text",
		Fn: func(a tools.Args) (any, error) {
			s, err := findScheme(a.Str("algo"))
			if err != nil {
				return nil, err
			}
			pt, err := decodeBytes(a.Str("text"), a.Str("input_encoding"))
			if err != nil {
				return nil, err
			}
			o, err := parseOpts(a, true, true)
			if err != nil {
				return nil, err
			}
			if !s.aes {
				o.bits = 0
			}
			if !s.aead {
				o.aad = nil
			}
			return s.encrypt(pt, o, a.Str("output"))
		},
		Example: `cipher.encrypt("aes_gcm", "secret message", "correct horse battery staple")`,
		Tags:    []string{"encrypt", "encryption", "generic", "any cipher", "aes", "chacha20", "password protect"},
		Family:  "cipher", Order: 90, Keywords: []string{"encrypt", "cipher"},
	})
	reg(tools.Func{
		Name: "decrypt", Title: "Decrypt with any cipher",
		Doc: "Decrypts data from encrypt() or any <cipher>_encrypt. For a zg1 envelope the algorithm is read from the envelope, so algo can be omitted; raw-key data needs algo and a key_encoding of utf-8, hex or base64. A wrong key or damaged data is an error (authenticated ciphers and all passphrase envelopes), never silent garbage.",
		Params: append(append([]tools.Param{reqStr("data", "Ciphertext: a zg1 envelope, or raw iv|ciphertext|tag in data_encoding.")},
			commonKeyParams(generic, true)[:2]...),
			append([]tools.Param{strParam("algo", "Cipher name; optional for zg1 envelopes.", "")},
				append(commonKeyParams(generic, true)[2:],
					tools.Param{Name: "data_encoding", Type: "str", Default: "base64", Choices: []string{"base64", "base64url", "hex", "base32"}, Doc: "Encoding of data in raw-key mode (default base64). Ignored for zg1 envelopes."},
					outParam("utf-8", "How the plaintext is returned: utf-8 text (default), hex, HEX, base64, base64url or base32."))...)...),
		Returns: "str", Input: "data",
		Fn: func(a tools.Args) (any, error) {
			var want *scheme
			if a.Str("algo") != "" {
				s, err := findScheme(a.Str("algo"))
				if err != nil {
					return nil, err
				}
				want = &s
			}
			o, err := parseOpts(a, false, true)
			if err != nil {
				return nil, err
			}
			pt, err := decrypt(want, a.Str("data"), a.Str("data_encoding"), o)
			if err != nil {
				return nil, err
			}
			return encodeBytes(pt, a.Str("output"))
		},
		Example: `cipher.decrypt("zg1.aes256gcm....", "correct horse battery staple")`,
		Tags:    []string{"decrypt", "decryption", "generic", "any cipher", "aes", "chacha20", "decipher"},
		Family:  "cipher", Order: 91, Keywords: []string{"decrypt", "cipher"},
		Accepts: looksEncrypted,
	})
	reg(tools.Func{
		Name: "algorithms", Title: "List ciphers",
		Doc:     "Returns the names accepted by encrypt() and decrypt() (symmetric ciphers), e.g. aes_gcm, chacha20_poly1305. The classic ciphers (caesar, vigenere, ...), xor_cipher and the rsa/ed25519/ecdsa functions have their own functions.",
		Returns: "list[str]",
		Fn:      func(tools.Args) (any, error) { return names(), nil },
		Example: `cipher.algorithms()`,
		Tags:    []string{"list", "supported", "available", "algorithms", "ciphers"},
	})
}

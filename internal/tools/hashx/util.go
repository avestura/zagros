package hashx

import (
	"crypto/hmac"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/base32"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"hash"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/avestura/zagros/internal/tools"
)

func cleanHex(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.TrimPrefix(s, "0x")
	return strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "", ":", "", "-", "").Replace(s)
}

func isHex(s string) bool {
	if s == "" {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f' || c >= 'A' && c <= 'F') {
			return false
		}
	}
	return true
}

// identifyHash lists plausible algorithms for a hash string by prefix and length.
func identifyHash(in string) []string {
	s := strings.TrimSpace(in)
	if s == "" {
		return nil
	}
	var out []string
	add := func(v ...string) { out = append(out, v...) }
	prefixes := []struct{ p, name string }{
		{"$2a$", "bcrypt"}, {"$2b$", "bcrypt"}, {"$2y$", "bcrypt"}, {"$2x$", "bcrypt"},
		{"$argon2id$", "argon2id (PHC string)"}, {"$argon2i$", "argon2i (PHC string)"}, {"$argon2d$", "argon2d (PHC string)"},
		{"$1$", "md5crypt (Unix MD5-crypt)"}, {"$apr1$", "Apache apr1 (MD5)"}, {"$5$", "sha256crypt (Unix SHA-256-crypt)"},
		{"$6$", "sha512crypt (Unix SHA-512-crypt)"}, {"$y$", "yescrypt"}, {"$7$", "scrypt (crypt format)"}, {"$scrypt$", "scrypt (passlib)"},
		{"$pbkdf2-sha256$", "PBKDF2-SHA256 (passlib)"}, {"$pbkdf2-sha512$", "PBKDF2-SHA512 (passlib)"}, {"$pbkdf2-sha1$", "PBKDF2-SHA1 (passlib)"},
		{"pbkdf2_sha256$", "PBKDF2-SHA256 (Django)"}, {"pbkdf2_sha1$", "PBKDF2-SHA1 (Django)"},
		{"$P$", "phpass (WordPress/phpBB)"}, {"$H$", "phpass (phpBB/WordPress)"}, {"$S$", "Drupal 7 (SHA-512 phpass)"},
		{"{SHA}", "Netscape LDAP SHA-1 (base64)"}, {"{SSHA}", "Salted SHA-1 LDAP (base64)"}, {"{SSHA512}", "Salted SHA-512 LDAP (base64)"},
		{"{MD5}", "LDAP MD5 (base64)"}, {"sha1$", "Django SHA-1 (salted)"}, {"md5$", "Django MD5 (salted)"},
	}
	for _, p := range prefixes {
		if strings.HasPrefix(s, p.p) {
			add(p.name)
		}
	}
	if len(out) > 0 {
		return out
	}
	if strings.HasPrefix(s, "*") && len(s) == 41 && isHex(s[1:]) {
		return []string{"MySQL 4.1+ PASSWORD() (SHA-1 of SHA-1)"}
	}
	h := cleanHex(s)
	if isHex(h) && len(h)%2 == 0 {
		switch len(h) * 4 {
		case 8:
			add("CRC-8", "Fletcher-8")
		case 16:
			add("CRC-16 (any variant)", "Fletcher-16")
		case 24:
			add("CRC-24 (OpenPGP)")
		case 32:
			add("CRC-32 / CRC-32C", "Adler-32", "Fletcher-32", "FNV-1/FNV-1a 32", "MurmurHash3 32", "xxHash32", "djb2 / sdbm / Jenkins one-at-a-time", "MySQL OLD_PASSWORD (half)")
		case 64:
			add("CRC-64", "FNV-1/FNV-1a 64", "xxHash64", "SipHash-2-4", "MySQL OLD_PASSWORD (323)", "BLAKE2b-64")
		case 128:
			add("MD5", "MD4", "NTLM (NT hash)", "LM hash", "MD2", "RIPEMD-128", "MurmurHash3 x64 128", "FNV-128", "BLAKE2s-128", "Double MD5 / Snefru-128")
		case 160:
			add("SHA-1", "RIPEMD-160", "HASH160 (Bitcoin)", "BLAKE2b-160", "Double SHA-1", "MySQL 4.1 (without '*')")
		case 192:
			add("Tiger-192", "HAVAL-192")
		case 224:
			add("SHA-224", "SHA3-224", "SHA-512/224", "BLAKE2b-224", "BLAKE2s-224", "Keccak-224")
		case 256:
			add("SHA-256", "SHA3-256", "Keccak-256 (Ethereum)", "SHA-512/256", "BLAKE2b-256", "BLAKE2s-256", "BLAKE3", "SM3", "RIPEMD-256", "Double SHA-256 (Bitcoin)", "GOST R 34.11", "Streebog-256", "HMAC-SHA256")
		case 320:
			add("RIPEMD-320")
		case 384:
			add("SHA-384", "SHA3-384", "BLAKE2b-384", "Keccak-384")
		case 512:
			add("SHA-512", "SHA3-512", "BLAKE2b-512", "Keccak-512", "Whirlpool", "Streebog-512", "HMAC-SHA512")
		default:
			add(fmt.Sprintf("unknown %d-bit digest (hex)", len(h)*4))
		}
		return out
	}
	if b, err := decodeB64(s); err == nil && len(b) >= 4 {
		switch len(b) {
		case 16:
			add("MD5 (base64)", "MD4 / NTLM (base64)")
		case 20:
			add("SHA-1 (base64)", "RIPEMD-160 (base64)")
		case 28:
			add("SHA-224 (base64)", "SHA3-224 (base64)")
		case 32:
			add("SHA-256 (base64)", "SHA3-256 (base64)", "BLAKE2b-256 (base64)")
		case 48:
			add("SHA-384 (base64)", "SHA3-384 (base64)")
		case 64:
			add("SHA-512 (base64)", "SHA3-512 (base64)", "BLAKE2b-512 (base64)")
		default:
			add(fmt.Sprintf("unknown %d-byte value (base64)", len(b)))
		}
	}
	return out
}

func init() {
	reg(tools.Func{
		Name: "checksum_compare", Title: "Compare checksums",
		Doc: "Compares two hex digests in constant time after lowercasing and dropping whitespace, ':' and '-' separators and any 0x prefix; returns true when equal. Non-hex input falls back to an exact constant-time string comparison.",
		Params: []tools.Param{
			{Name: "a", Type: "str", Required: true, Doc: "First checksum."},
			{Name: "b", Type: "str", Required: true, Doc: "Second checksum."},
		},
		Returns: "bool",
		Fn: func(a tools.Args) (any, error) {
			x, y := cleanHex(a.Str("a")), cleanHex(a.Str("b"))
			if !isHex(x) || !isHex(y) {
				x, y = strings.TrimSpace(a.Str("a")), strings.TrimSpace(a.Str("b"))
			}
			return subtle.ConstantTimeCompare([]byte(x), []byte(y)) == 1, nil
		},
		Example: `hash.checksum_compare("D41D8CD9 8F00B204", "d41d8cd98f00b204")`,
		Tags:    []string{"compare", "checksum", "verify", "equal", "match", "constant time", "digest", "sha256sum"},
	})
	reg(tools.Func{
		Name: "hash_identify", Title: "Identify hash type",
		Doc:     "Guesses which algorithms could have produced a hash string, from its prefix ($2b$ bcrypt, $argon2id$, $6$ sha512crypt, {SSHA}, ...) or its length (32 hex digits = md5/ntlm/..., 64 = sha256/...). Returns an ordered list of candidates, empty when nothing matches. Only a hint, since many algorithms share a length.",
		Params:  []tools.Param{{Name: "hash", Type: "str", Required: true, Doc: "Hash string, hex, base64 or a modular-crypt/PHC string."}},
		Returns: "list[str]",
		Fn:      func(a tools.Args) (any, error) { return append([]string{}, identifyHash(a.Str("hash"))...), nil },
		Example: `hash.hash_identify("5d41402abc4b2a76b9719d911017c592")`,
		Tags:    []string{"identify", "detect", "guess", "what hash", "hash type", "recognize", "md5 or sha1"},
		Family:  "hash", Order: 600, Keywords: []string{"hash", "identify"},
		Accepts: func(s string) bool { return len(identifyHash(s)) > 0 },
	})
	reg(tools.Func{
		Name: "hash_to_base64", Title: "Hex digest to base64",
		Doc:     "Converts a hex digest (separators and 0x prefix tolerated) to standard base64 with padding, the form used by SRI integrity attributes and HTTP Digest headers. Odd or non-hex input is an error.",
		Params:  []tools.Param{{Name: "hex", Type: "str", Required: true, Doc: "Hex digest."}},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			b, err := decodeBytes(a.Str("hex"), "hex")
			if err != nil {
				return nil, err
			}
			return base64.StdEncoding.EncodeToString(b), nil
		},
		Example: `hash.hash_to_base64("d41d8cd98f00b204e9800998ecf8427e")`,
		Tags:    []string{"hex", "base64", "convert", "sri", "integrity", "digest"},
		Family:  "hash", Order: 601, Keywords: []string{"hash", "base64"},
		Accepts: func(s string) bool { b, err := decodeBytes(s, "hex"); return err == nil && len(b) >= 4 },
	})
	registerOTP()
}

// ---- HOTP / TOTP ----

func otpHasher(algo string) (func() hash.Hash, error) {
	switch normAlgo(algo) {
	case "", "sha1", "sha_1":
		return sha1.New, nil
	case "sha256", "sha_256":
		return sha256.New, nil
	case "sha512", "sha_512":
		return sha512.New, nil
	}
	return nil, fmt.Errorf("otp algorithm must be sha1, sha256 or sha512 (got %q)", algo)
}

func decodeBase32Secret(s string) ([]byte, error) {
	t := strings.ToUpper(strings.NewReplacer(" ", "", "-", "", "\t", "", "\n", "", "\r", "").Replace(s))
	t = strings.TrimRight(t, "=")
	if t == "" {
		return nil, fmt.Errorf("secret is empty")
	}
	b, err := base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(t)
	if err != nil {
		return nil, fmt.Errorf("secret must be base32 (A-Z, 2-7): %v", err)
	}
	return b, nil
}

func hotpCode(key []byte, counter uint64, digits int, nh func() hash.Hash) string {
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], counter)
	m := hmac.New(nh, key)
	m.Write(msg[:])
	sum := m.Sum(nil)
	off := sum[len(sum)-1] & 0x0f
	v := uint64(binary.BigEndian.Uint32(sum[off:off+4]) & 0x7fffffff)
	mod := uint64(1)
	for i := 0; i < digits; i++ {
		mod *= 10
	}
	return fmt.Sprintf("%0*d", digits, v%mod)
}

func registerOTP() {
	secret := tools.Param{Name: "secret", Type: "str", Required: true, Doc: "Shared secret in base32 (spaces, dashes, case and '=' padding ignored)."}
	digits := tools.Param{Name: "digits", Type: "int", Default: int64(6), Doc: "Code length, 1 to 10 (default 6)."}
	algo := tools.Param{Name: "algo", Type: "str", Default: "sha1", Choices: []string{"sha1", "sha256", "sha512"}, Doc: "HMAC hash: sha1 (default, what authenticator apps expect), sha256 or sha512."}
	period := tools.Param{Name: "period", Type: "int", Default: int64(30), Doc: "Time step in seconds (default 30)."}
	timeP := tools.Param{Name: "time", Type: "int", Default: int64(0), Doc: "Unix time in seconds; 0 (default) means now."}
	checkDigits := func(a tools.Args) error { return clampInt(a.Int("digits"), 1, 10, "digits") }
	counterOf := func(a tools.Args) (uint64, error) {
		if a.Int("period") < 1 {
			return 0, fmt.Errorf("period must be at least 1 second")
		}
		t := a.Int("time")
		if t == 0 {
			t = time.Now().Unix()
		}
		if t < 0 {
			return 0, fmt.Errorf("time must not be negative")
		}
		return uint64(t) / uint64(a.Int("period")), nil
	}
	reg(tools.Func{
		Name: "hotp", Title: "HOTP code",
		Doc:     "Computes the HOTP one-time code (RFC 4226) for a base32 secret and counter. Returns a zero-padded string of `digits` digits (default 6, HMAC-SHA1).",
		Params:  []tools.Param{secret, {Name: "counter", Type: "int", Required: true, Doc: "Counter value, 0 or more."}, digits, algo},
		Returns: "str", Input: "secret",
		Fn: func(a tools.Args) (any, error) {
			if err := checkDigits(a); err != nil {
				return nil, err
			}
			if a.Int("counter") < 0 {
				return nil, fmt.Errorf("counter must not be negative")
			}
			key, err := decodeBase32Secret(a.Str("secret"))
			if err != nil {
				return nil, err
			}
			nh, err := otpHasher(a.Str("algo"))
			if err != nil {
				return nil, err
			}
			return hotpCode(key, uint64(a.Int("counter")), int(a.Int("digits")), nh), nil
		},
		Example: `hash.hotp("GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ", 0)`,
		Tags:    []string{"hotp", "otp", "one time password", "2fa", "mfa", "rfc4226", "authenticator", "counter"},
	})
	reg(tools.Func{
		Name: "totp", Title: "TOTP code",
		Doc:     "Computes the TOTP code (RFC 6238) for a base32 secret at a Unix time (0 means now, the only use of the clock in this module). Defaults: 30 s period, 6 digits, HMAC-SHA1, as in Google Authenticator. Returns a zero-padded string.",
		Params:  []tools.Param{secret, timeP, period, digits, algo},
		Returns: "str", Input: "secret",
		Fn: func(a tools.Args) (any, error) {
			if err := checkDigits(a); err != nil {
				return nil, err
			}
			key, err := decodeBase32Secret(a.Str("secret"))
			if err != nil {
				return nil, err
			}
			nh, err := otpHasher(a.Str("algo"))
			if err != nil {
				return nil, err
			}
			c, err := counterOf(a)
			if err != nil {
				return nil, err
			}
			return hotpCode(key, c, int(a.Int("digits")), nh), nil
		},
		Example: `hash.totp("JBSWY3DPEHPK3PXP")`,
		Tags:    []string{"totp", "otp", "2fa", "mfa", "authenticator", "google authenticator", "rfc6238", "one time password", "time based"},
		Family:  "hash", Order: 700, Keywords: []string{"totp", "otp", "2fa"},
		Accepts: func(s string) bool {
			_, err := decodeBase32Secret(s)
			return err == nil && len(strings.TrimSpace(s)) >= 8
		},
	})
	reg(tools.Func{
		Name: "totp_verify", Title: "Verify TOTP code",
		Doc: "Checks a TOTP code against a base32 secret, accepting `window` time steps before and after the given time (default 1, to absorb clock drift). Time 0 means now. Compares in constant time and returns true or false.",
		Params: []tools.Param{secret, {Name: "code", Type: "str", Required: true, Doc: "Code to check."}, timeP, period, digits, algo,
			{Name: "window", Type: "int", Default: int64(1), Doc: "Accepted steps either side of the current one, 0 to 10 (default 1)."}},
		Returns: "bool", Input: "secret",
		Fn: func(a tools.Args) (any, error) {
			if err := checkDigits(a); err != nil {
				return nil, err
			}
			if err := clampInt(a.Int("window"), 0, 10, "window"); err != nil {
				return nil, err
			}
			key, err := decodeBase32Secret(a.Str("secret"))
			if err != nil {
				return nil, err
			}
			nh, err := otpHasher(a.Str("algo"))
			if err != nil {
				return nil, err
			}
			c, err := counterOf(a)
			if err != nil {
				return nil, err
			}
			code := strings.TrimSpace(a.Str("code"))
			ok := 0
			for d := -a.Int("window"); d <= a.Int("window"); d++ {
				cc := int64(c) + d
				if cc < 0 {
					continue
				}
				ok |= subtle.ConstantTimeCompare([]byte(code), []byte(hotpCode(key, uint64(cc), int(a.Int("digits")), nh)))
			}
			return ok == 1, nil
		},
		Example: `hash.totp_verify("JBSWY3DPEHPK3PXP", "123456")`,
		Tags:    []string{"totp", "verify", "check", "2fa", "otp", "validate code"},
	})
	uriParams := func() []tools.Param {
		return []tools.Param{secret,
			{Name: "account", Type: "str", Required: true, Doc: "Account name, e.g. an email address."},
			{Name: "issuer", Type: "str", Default: "", Doc: "Service name shown in the authenticator app."},
			digits, algo}
	}
	buildURI := func(kind string, a tools.Args, extra url.Values, extraLabel string) (string, error) {
		if err := checkDigits(a); err != nil {
			return "", err
		}
		key, err := decodeBase32Secret(a.Str("secret"))
		if err != nil {
			return "", err
		}
		nh, err := otpHasher(a.Str("algo"))
		if err != nil {
			return "", err
		}
		_ = nh
		sec := strings.TrimRight(base32.StdEncoding.EncodeToString(key), "=")
		label := url.PathEscape(a.Str("account"))
		q := url.Values{}
		q.Set("secret", sec)
		if iss := a.Str("issuer"); iss != "" {
			label = url.PathEscape(iss) + ":" + label
			q.Set("issuer", iss)
		}
		q.Set("algorithm", strings.ToUpper(normAlgo(a.Str("algo"))))
		q.Set("digits", strconv.FormatInt(a.Int("digits"), 10))
		for k := range extra {
			q.Set(k, extra.Get(k))
		}
		return "otpauth://" + kind + "/" + label + "?" + q.Encode(), nil
	}
	reg(tools.Func{
		Name: "totp_uri", Title: "TOTP otpauth:// URI",
		Doc:     "Builds the otpauth://totp/ provisioning URI (the content of 2FA QR codes) from a base32 secret, account and issuer. The secret is normalised to uppercase unpadded base32; period is included (default 30).",
		Params:  append(uriParams(), tools.Param{Name: "period", Type: "int", Default: int64(30), Doc: "Time step in seconds (default 30)."}),
		Returns: "str", Input: "secret",
		Fn: func(a tools.Args) (any, error) {
			if a.Int("period") < 1 {
				return nil, fmt.Errorf("period must be at least 1 second")
			}
			return buildURI("totp", a, url.Values{"period": {strconv.FormatInt(a.Int("period"), 10)}}, "")
		},
		Example: `hash.totp_uri("JBSWY3DPEHPK3PXP", "alice@example.com", issuer="Example")`,
		Tags:    []string{"totp", "uri", "otpauth", "qr", "2fa", "provisioning", "authenticator"},
	})
	reg(tools.Func{
		Name: "hotp_uri", Title: "HOTP otpauth:// URI",
		Doc:     "Builds the otpauth://hotp/ provisioning URI from a base32 secret, account, issuer and initial counter.",
		Params:  append(uriParams(), tools.Param{Name: "counter", Type: "int", Default: int64(0), Doc: "Initial counter (default 0)."}),
		Returns: "str", Input: "secret",
		Fn: func(a tools.Args) (any, error) {
			if a.Int("counter") < 0 {
				return nil, fmt.Errorf("counter must not be negative")
			}
			return buildURI("hotp", a, url.Values{"counter": {strconv.FormatInt(a.Int("counter"), 10)}}, "")
		},
		Example: `hash.hotp_uri("JBSWY3DPEHPK3PXP", "alice@example.com", issuer="Example")`,
		Tags:    []string{"hotp", "uri", "otpauth", "qr", "2fa", "provisioning"},
	})
}

var _ = hex.EncodeToString

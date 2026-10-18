package hashx

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"hash"
	"io"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/crypto/scrypt"

	"github.com/avestura/zagros/internal/tools"
)

// Limits that keep one call from hanging the launcher.
const (
	maxPBKDF2Iter  = 5_000_000
	maxBcryptCost  = 14
	maxScryptN     = 1 << 20
	maxKDFMemBytes = 256 << 20
	maxArgonMemKiB = 256 * 1024
	maxArgonTime   = 10
	maxKDFLength   = 1024
)

func hasherFor(name string) (func() hash.Hash, error) {
	a, err := lookupAlgo(name)
	if err != nil {
		return nil, err
	}
	if a.newH == nil {
		return nil, fmt.Errorf("%s cannot be used for HMAC/KDF (use md5, sha1, sha224, sha256, sha384, sha512, sha3_*, blake2*, ripemd160, sm3, keccak)", name)
	}
	return a.newH, nil
}

func macSum(newH func() hash.Hash, key, msg []byte) []byte {
	m := hmac.New(newH, key)
	m.Write(msg)
	return m.Sum(nil)
}

func init() {
	macParams := func(withAlgo bool) []tools.Param {
		var ps []tools.Param
		if withAlgo {
			ps = append(ps, tools.Param{Name: "algo", Type: "str", Required: true, Doc: "Hash algorithm name, e.g. sha256, sha512, sha3_256, blake2b_256."})
		}
		return append(ps,
			tools.Param{Name: "key", Type: "str", Required: true, Doc: "Secret key (UTF-8 unless key_encoding is hex or base64)."},
			textParam(),
			encParam("key_encoding", "the key"),
			inEncParam(), outParam())
	}
	hmacFn := func(fixed string) func(a tools.Args) (any, error) {
		return func(a tools.Args) (any, error) {
			name := fixed
			if name == "" {
				name = a.Str("algo")
			}
			nh, err := hasherFor(name)
			if err != nil {
				return nil, err
			}
			k, err := decodeBytes(a.Str("key"), a.Str("key_encoding"))
			if err != nil {
				return nil, err
			}
			b, err := inBytes(a)
			if err != nil {
				return nil, err
			}
			return encodeBytes(macSum(nh, k, b), a.Str("output"))
		}
	}
	reg(tools.Func{
		Name: "hmac", Title: "HMAC",
		Doc:     "Computes the HMAC (RFC 2104) of text under a secret key with the named hash (sha256, sha512, sha1, md5, sha3_256, blake2b_256, ... see algorithms()). Key and text are UTF-8 unless key_encoding / input_encoding say hex or base64; output is lowercase hex by default.",
		Params:  macParams(true),
		Returns: "str", Input: "text",
		Fn:      hmacFn(""),
		Example: `hash.hmac("sha256", "secret", "message")`,
		Tags:    []string{"hmac", "mac", "message authentication code", "signature", "jwt", "webhook", "rfc2104", "rfc4231"},
	})
	for i, n := range []string{"md5", "sha1", "sha256", "sha384", "sha512"} {
		n := n
		reg(tools.Func{
			Name: "hmac_" + n, Title: "HMAC-" + strings.ToUpper(n),
			Doc:     fmt.Sprintf("Computes HMAC-%s of text under a secret key. Key and text are UTF-8 unless key_encoding / input_encoding say hex or base64; output is lowercase hex by default.", strings.ToUpper(n)),
			Params:  macParams(false),
			Returns: "str", Input: "text",
			Fn:       hmacFn(n),
			Example:  fmt.Sprintf(`hash.hmac_%s("secret", "message")`, n),
			Tags:     []string{"hmac", "mac", n, "signature", "jwt", "webhook"},
			Keywords: []string{"hmac", "hmac_" + n},
			Order:    500 + i,
		})
	}
	reg(tools.Func{
		Name: "hmac_verify", Title: "Verify HMAC",
		Doc: "Recomputes the HMAC of text under key and compares it to mac in constant time; returns true on a match. mac may be hex or base64 (case and whitespace of hex are ignored); an undecodable mac is simply false.",
		Params: []tools.Param{
			{Name: "algo", Type: "str", Required: true, Doc: "Hash algorithm name, e.g. sha256."},
			{Name: "key", Type: "str", Required: true, Doc: "Secret key."},
			textParam(),
			{Name: "mac", Type: "str", Required: true, Doc: "Expected MAC as hex or base64."},
			encParam("key_encoding", "the key"), inEncParam(),
		},
		Returns: "bool", Input: "text",
		Fn: func(a tools.Args) (any, error) {
			nh, err := hasherFor(a.Str("algo"))
			if err != nil {
				return nil, err
			}
			k, err := decodeBytes(a.Str("key"), a.Str("key_encoding"))
			if err != nil {
				return nil, err
			}
			b, err := inBytes(a)
			if err != nil {
				return nil, err
			}
			want := macSum(nh, k, b)
			got, err := decodeBytes(a.Str("mac"), "hex")
			if err != nil {
				got, err = decodeBytes(a.Str("mac"), "base64")
				if err != nil {
					return false, nil
				}
			}
			return hmac.Equal(want, got), nil
		},
		Example: `hash.hmac_verify("sha256", "secret", "message", "8b5f48702995c1598c573db1e21866a9b825d4a794d169d7060a03605796360b")`,
		Tags:    []string{"hmac", "verify", "check", "constant time", "signature"},
	})

	// HKDF
	reg(tools.Func{
		Name: "hkdf", Title: "HKDF",
		Doc: "Derives length bytes (1 to 1024) of key material with HKDF (RFC 5869). secret is decoded with key_encoding; salt and info with salt_encoding (utf-8, hex or base64); empty salt is allowed. Output is hex by default.",
		Params: []tools.Param{
			{Name: "algo", Type: "str", Required: true, Doc: "Hash algorithm, e.g. sha256."},
			{Name: "secret", Type: "str", Required: true, Doc: "Input key material."},
			{Name: "salt", Type: "str", Default: "", Doc: "Optional salt."},
			{Name: "info", Type: "str", Default: "", Doc: "Optional context/application info."},
			{Name: "length", Type: "int", Default: int64(32), Doc: "Output length in bytes (1 to 1024, default 32)."},
			encParam("key_encoding", "the secret"), encParam("salt_encoding", "salt and info"), outParam(),
		},
		Returns: "str", Input: "secret",
		Fn: func(a tools.Args) (any, error) {
			nh, err := hasherFor(a.Str("algo"))
			if err != nil {
				return nil, err
			}
			if err := clampInt(a.Int("length"), 1, maxKDFLength, "length"); err != nil {
				return nil, err
			}
			sec, err := decodeBytes(a.Str("secret"), a.Str("key_encoding"))
			if err != nil {
				return nil, err
			}
			salt, err := decodeBytes(a.Str("salt"), a.Str("salt_encoding"))
			if err != nil {
				return nil, err
			}
			info, err := decodeBytes(a.Str("info"), a.Str("salt_encoding"))
			if err != nil {
				return nil, err
			}
			out := make([]byte, a.Int("length"))
			if _, err := io.ReadFull(hkdf.New(nh, sec, salt, info), out); err != nil {
				return nil, err
			}
			return encodeBytes(out, a.Str("output"))
		},
		Example: `hash.hkdf("sha256", "input key", salt="salt", info="app", length=32)`,
		Tags:    []string{"hkdf", "kdf", "key derivation", "rfc5869", "derive key", "expand"},
	})

	// PBKDF2
	reg(tools.Func{
		Name: "pbkdf2", Title: "PBKDF2",
		Doc: "Derives a key from a password with PBKDF2 (RFC 8018). Iterations default to 600000 and are capped at 5,000,000; length is 1 to 1024 bytes. The password is UTF-8; salt follows salt_encoding. Output is hex by default.",
		Params: []tools.Param{
			{Name: "password", Type: "str", Required: true, Doc: "Password (UTF-8)."},
			{Name: "salt", Type: "str", Required: true, Doc: "Salt, per salt_encoding."},
			{Name: "iterations", Type: "int", Default: int64(600000), Doc: "Iteration count, 1 to 5000000 (default 600000)."},
			{Name: "length", Type: "int", Default: int64(32), Doc: "Derived key length in bytes, 1 to 1024 (default 32)."},
			{Name: "algo", Type: "str", Default: "sha256", Doc: "Underlying hash, default sha256 (sha1, sha512, ... also work)."},
			encParam("salt_encoding", "the salt"), outParam(),
		},
		Returns: "str", Input: "password",
		Fn: func(a tools.Args) (any, error) {
			nh, err := hasherFor(a.Str("algo"))
			if err != nil {
				return nil, err
			}
			if err := clampInt(a.Int("iterations"), 1, maxPBKDF2Iter, "iterations"); err != nil {
				return nil, err
			}
			if err := clampInt(a.Int("length"), 1, maxKDFLength, "length"); err != nil {
				return nil, err
			}
			salt, err := decodeBytes(a.Str("salt"), a.Str("salt_encoding"))
			if err != nil {
				return nil, err
			}
			k := pbkdf2.Key([]byte(a.Str("password")), salt, int(a.Int("iterations")), int(a.Int("length")), nh)
			return encodeBytes(k, a.Str("output"))
		},
		Example: `hash.pbkdf2("password", "salt", iterations=100000)`,
		Tags:    []string{"pbkdf2", "kdf", "password hashing", "key derivation", "rfc2898", "rfc8018", "rfc6070"},
	})

	// scrypt
	reg(tools.Func{
		Name: "scrypt", Title: "scrypt",
		Doc: "Derives a key from a password with scrypt (RFC 7914). Defaults n=32768, r=8, p=1, length=32. Caps: n a power of two up to 2^20, memory (128*n*r bytes) at most 256 MiB, length up to 1024. Password is UTF-8, salt follows salt_encoding; output is hex by default.",
		Params: []tools.Param{
			{Name: "password", Type: "str", Required: true, Doc: "Password (UTF-8)."},
			{Name: "salt", Type: "str", Required: true, Doc: "Salt, per salt_encoding."},
			{Name: "n", Type: "int", Default: int64(32768), Doc: "CPU/memory cost, a power of two from 2 to 1048576 (default 32768)."},
			{Name: "r", Type: "int", Default: int64(8), Doc: "Block size, default 8."},
			{Name: "p", Type: "int", Default: int64(1), Doc: "Parallelization, default 1."},
			{Name: "length", Type: "int", Default: int64(32), Doc: "Derived key length in bytes, 1 to 1024 (default 32)."},
			encParam("salt_encoding", "the salt"), outParam(),
		},
		Returns: "str", Input: "password",
		Fn: func(a tools.Args) (any, error) {
			n, r, p := a.Int("n"), a.Int("r"), a.Int("p")
			if n < 2 || n > maxScryptN || n&(n-1) != 0 {
				return nil, fmt.Errorf("n must be a power of two between 2 and %d", maxScryptN)
			}
			if err := clampInt(r, 1, 256, "r"); err != nil {
				return nil, err
			}
			if err := clampInt(p, 1, 256, "p"); err != nil {
				return nil, err
			}
			if err := clampInt(a.Int("length"), 1, maxKDFLength, "length"); err != nil {
				return nil, err
			}
			if 128*n*r > maxKDFMemBytes || 128*r*p > maxKDFMemBytes {
				return nil, fmt.Errorf("scrypt memory (128*n*r = %d bytes) exceeds the %d MiB cap", 128*n*r, maxKDFMemBytes>>20)
			}
			salt, err := decodeBytes(a.Str("salt"), a.Str("salt_encoding"))
			if err != nil {
				return nil, err
			}
			k, err := scrypt.Key([]byte(a.Str("password")), salt, int(n), int(r), int(p), int(a.Int("length")))
			if err != nil {
				return nil, err
			}
			return encodeBytes(k, a.Str("output"))
		},
		Example: `hash.scrypt("password", "NaCl", n=1024, r=8, p=16)`,
		Tags:    []string{"scrypt", "kdf", "password hashing", "key derivation", "rfc7914", "memory hard"},
	})

	// bcrypt
	reg(tools.Func{
		Name: "bcrypt_hash", Title: "bcrypt hash",
		Doc: "Hashes a password with bcrypt and a random salt, returning the 60-character $2a$ string. cost is 4 to 14 (default 10, each step doubles the time); passwords longer than 72 bytes are an error rather than silently truncated.",
		Params: []tools.Param{
			{Name: "password", Type: "str", Required: true, Doc: "Password, at most 72 bytes."},
			{Name: "cost", Type: "int", Default: int64(10), Doc: "Work factor 4 to 14 (default 10)."},
		},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			if err := clampInt(a.Int("cost"), 4, maxBcryptCost, "cost"); err != nil {
				return nil, err
			}
			h, err := bcrypt.GenerateFromPassword([]byte(a.Str("password")), int(a.Int("cost")))
			if err != nil {
				return nil, err
			}
			return string(h), nil
		},
		Example: `hash.bcrypt_hash("hunter2", cost=10)`,
		Tags:    []string{"bcrypt", "password hashing", "blowfish", "$2a$", "$2b$", "salt", "hash password"},
	})
	reg(tools.Func{
		Name: "bcrypt_verify", Title: "bcrypt verify",
		Doc: "Checks a password against a bcrypt hash ($2a$, $2b$ or $2y$) and returns true on a match, false on a wrong password. A malformed hash string is an error.",
		Params: []tools.Param{
			{Name: "password", Type: "str", Required: true, Doc: "Password to check."},
			{Name: "hash", Type: "str", Required: true, Doc: "bcrypt hash string."},
		},
		Returns: "bool",
		Fn: func(a tools.Args) (any, error) {
			h := strings.TrimSpace(a.Str("hash"))
			if strings.HasPrefix(h, "$2y$") || strings.HasPrefix(h, "$2x$") {
				h = "$2a$" + h[4:]
			}
			err := bcrypt.CompareHashAndPassword([]byte(h), []byte(a.Str("password")))
			if err == nil {
				return true, nil
			}
			if err == bcrypt.ErrMismatchedHashAndPassword {
				return false, nil
			}
			return nil, err
		},
		Example: `hash.bcrypt_verify("hunter2", "$2a$10$...")`,
		Tags:    []string{"bcrypt", "verify", "check password", "compare"},
	})

	// Argon2
	argonParams := func() []tools.Param {
		return []tools.Param{
			{Name: "time", Type: "int", Default: int64(3), Doc: "Passes over memory, 1 to 10 (default 3)."},
			{Name: "memory_kib", Type: "int", Default: int64(65536), Doc: "Memory in KiB, 8 to 262144 (default 65536 = 64 MiB)."},
			{Name: "threads", Type: "int", Default: int64(4), Doc: "Parallelism 1 to 64 (default 4)."},
			{Name: "length", Type: "int", Default: int64(32), Doc: "Tag length in bytes, 4 to 1024 (default 32)."},
		}
	}
	argonCheck := func(a tools.Args) error {
		if err := clampInt(a.Int("time"), 1, maxArgonTime, "time"); err != nil {
			return err
		}
		if err := clampInt(a.Int("memory_kib"), 8, maxArgonMemKiB, "memory_kib"); err != nil {
			return err
		}
		if err := clampInt(a.Int("threads"), 1, 64, "threads"); err != nil {
			return err
		}
		if a.Int("memory_kib") < 8*a.Int("threads") {
			return fmt.Errorf("memory_kib must be at least 8*threads")
		}
		return clampInt(a.Int("length"), 4, maxKDFLength, "length")
	}
	for _, v := range []string{"argon2id", "argon2i"} {
		v := v
		reg(tools.Func{
			Name: v, Title: strings.Replace(strings.ToUpper(v), "ARGON2", "Argon2", 1),
			Doc: fmt.Sprintf("Derives a raw key with %s (RFC 9106, version 1.3). Defaults time=3, memory_kib=65536 (64 MiB), threads=4, length=32; caps time 10, memory 256 MiB, threads 64. Password is UTF-8, salt follows salt_encoding (use at least 8 bytes); output hex by default. For a self-describing stored hash use argon2id_hash.", v),
			Params: append(append([]tools.Param{
				{Name: "password", Type: "str", Required: true, Doc: "Password (UTF-8)."},
				{Name: "salt", Type: "str", Required: true, Doc: "Salt per salt_encoding, at least 8 bytes recommended."},
			}, argonParams()...), encParam("salt_encoding", "the salt"), outParam()),
			Returns: "str", Input: "password",
			Fn: func(a tools.Args) (any, error) {
				if err := argonCheck(a); err != nil {
					return nil, err
				}
				salt, err := decodeBytes(a.Str("salt"), a.Str("salt_encoding"))
				if err != nil {
					return nil, err
				}
				var k []byte
				pw := []byte(a.Str("password"))
				t, m, th, l := uint32(a.Int("time")), uint32(a.Int("memory_kib")), uint8(a.Int("threads")), uint32(a.Int("length"))
				if v == "argon2id" {
					k = argon2.IDKey(pw, salt, t, m, th, l)
				} else {
					k = argon2.Key(pw, salt, t, m, th, l)
				}
				return encodeBytes(k, a.Str("output"))
			},
			Example: fmt.Sprintf(`hash.%s("password", "somesalt", time=2, memory_kib=19456, threads=1)`, v),
			Tags:    []string{"argon2", v, "kdf", "password hashing", "rfc9106", "memory hard", "key derivation"},
		})
	}
	reg(tools.Func{
		Name: "argon2id_hash", Title: "Argon2id password hash",
		Doc:     "Hashes a password with Argon2id and a fresh random 16-byte salt, returning the standard PHC string ($argon2id$v=19$m=...,t=...,p=...$salt$hash). Defaults time=3, memory_kib=65536, threads=4, length=32 (caps: time 10, 256 MiB, 64 threads). Not deterministic; verify with argon2_verify.",
		Params:  append([]tools.Param{{Name: "password", Type: "str", Required: true, Doc: "Password (UTF-8)."}}, argonParams()...),
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			if err := argonCheck(a); err != nil {
				return nil, err
			}
			salt := make([]byte, 16)
			if _, err := rand.Read(salt); err != nil {
				return nil, err
			}
			t, m, th, l := uint32(a.Int("time")), uint32(a.Int("memory_kib")), uint8(a.Int("threads")), uint32(a.Int("length"))
			k := argon2.IDKey([]byte(a.Str("password")), salt, t, m, th, l)
			e := base64.RawStdEncoding
			return fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%s$%s", m, t, th, e.EncodeToString(salt), e.EncodeToString(k)), nil
		},
		Example: `hash.argon2id_hash("hunter2")`,
		Tags:    []string{"argon2", "argon2id", "password hashing", "phc", "random salt", "hash password"},
	})
	reg(tools.Func{
		Name: "argon2_verify", Title: "Argon2 verify",
		Doc: "Checks a password against a PHC string produced by argon2id_hash (or any $argon2id$ / $argon2i$ v=19 hash) and returns true on a match, false on a wrong password. Malformed strings, or parameters above the caps (time 10, 256 MiB), are errors.",
		Params: []tools.Param{
			{Name: "password", Type: "str", Required: true, Doc: "Password to check."},
			{Name: "phc", Type: "str", Required: true, Doc: "PHC string such as $argon2id$v=19$m=65536,t=3,p=4$salt$hash."},
		},
		Returns: "bool",
		Fn: func(a tools.Args) (any, error) {
			parts := strings.Split(strings.TrimSpace(a.Str("phc")), "$")
			if len(parts) != 6 || parts[0] != "" || (parts[1] != "argon2id" && parts[1] != "argon2i") {
				return nil, fmt.Errorf("not an $argon2id$ / $argon2i$ PHC string")
			}
			if parts[2] != "v=19" {
				return nil, fmt.Errorf("unsupported argon2 version %q (only v=19)", parts[2])
			}
			var m, t, p uint64
			for _, kv := range strings.Split(parts[3], ",") {
				k, v, ok := strings.Cut(kv, "=")
				if !ok {
					return nil, fmt.Errorf("bad argon2 parameters %q", parts[3])
				}
				n, err := strconv.ParseUint(v, 10, 32)
				if err != nil {
					return nil, fmt.Errorf("bad argon2 parameter %q", kv)
				}
				switch k {
				case "m":
					m = n
				case "t":
					t = n
				case "p":
					p = n
				}
			}
			if t < 1 || t > maxArgonTime || m < 8 || m > maxArgonMemKiB || p < 1 || p > 64 || m < 8*p {
				return nil, fmt.Errorf("argon2 parameters out of range (time 1-%d, memory 8-%d KiB, threads 1-64)", maxArgonTime, maxArgonMemKiB)
			}
			salt, err := base64.RawStdEncoding.DecodeString(parts[4])
			if err != nil {
				return nil, fmt.Errorf("bad argon2 salt: %v", err)
			}
			want, err := base64.RawStdEncoding.DecodeString(parts[5])
			if err != nil || len(want) < 4 || len(want) > maxKDFLength {
				return nil, fmt.Errorf("bad argon2 hash")
			}
			pw := []byte(a.Str("password"))
			var got []byte
			if parts[1] == "argon2id" {
				got = argon2.IDKey(pw, salt, uint32(t), uint32(m), uint8(p), uint32(len(want)))
			} else {
				got = argon2.Key(pw, salt, uint32(t), uint32(m), uint8(p), uint32(len(want)))
			}
			return subtle.ConstantTimeCompare(got, want) == 1, nil
		},
		Example: `hash.argon2_verify("hunter2", "$argon2id$v=19$m=65536,t=3,p=4$...$...")`,
		Tags:    []string{"argon2", "verify", "check password", "phc", "compare"},
	})
}

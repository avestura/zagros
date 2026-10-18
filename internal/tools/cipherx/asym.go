package cipherx

import (
	"crypto"
	"crypto/ecdh"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"fmt"
	"hash"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

func hashFor(name string) (crypto.Hash, func() hash.Hash, error) {
	switch normAlgo(name) {
	case "", "sha256", "sha_256":
		return crypto.SHA256, sha256.New, nil
	case "sha1", "sha_1":
		return crypto.SHA1, sha1.New, nil
	case "sha384", "sha_384":
		return crypto.SHA384, sha512.New384, nil
	case "sha512", "sha_512":
		return crypto.SHA512, sha512.New, nil
	}
	return 0, nil, fmt.Errorf("unsupported hash %q (use sha256, sha384, sha512 or sha1)", name)
}

func digestOf(nh func() hash.Hash, b []byte) []byte {
	h := nh()
	h.Write(b)
	return h.Sum(nil)
}

func pemBlock(typ string, der []byte) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der}))
}

func parsePrivate(s string) (any, error) {
	blk, _ := pem.Decode([]byte(strings.TrimSpace(s)))
	if blk == nil {
		return nil, errors.New("not a PEM private key (expected -----BEGIN PRIVATE KEY----- ...)")
	}
	if k, err := x509.ParsePKCS8PrivateKey(blk.Bytes); err == nil {
		return k, nil
	}
	if k, err := x509.ParsePKCS1PrivateKey(blk.Bytes); err == nil {
		return k, nil
	}
	if k, err := x509.ParseECPrivateKey(blk.Bytes); err == nil {
		return k, nil
	}
	return nil, errors.New("unsupported or corrupted private key (PKCS#8, PKCS#1 and SEC1 are accepted; encrypted keys are not)")
}

func parsePublic(s string) (any, error) {
	blk, _ := pem.Decode([]byte(strings.TrimSpace(s)))
	if blk == nil {
		return nil, errors.New("not a PEM public key (expected -----BEGIN PUBLIC KEY----- ...)")
	}
	if k, err := x509.ParsePKIXPublicKey(blk.Bytes); err == nil {
		return k, nil
	}
	if k, err := x509.ParsePKCS1PublicKey(blk.Bytes); err == nil {
		return k, nil
	}
	if c, err := x509.ParseCertificate(blk.Bytes); err == nil {
		return c.PublicKey, nil
	}
	// A private key PEM is accepted too: use its public half.
	if k, err := parsePrivate(s); err == nil {
		if s, ok := k.(interface{ Public() crypto.PublicKey }); ok {
			return s.Public(), nil
		}
	}
	return nil, errors.New("unsupported or corrupted public key")
}

func rsaPrivate(s string) (*rsa.PrivateKey, error) {
	k, err := parsePrivate(s)
	if err != nil {
		return nil, err
	}
	r, ok := k.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("the key is not an RSA private key")
	}
	return r, nil
}

func rsaPublic(s string) (*rsa.PublicKey, error) {
	k, err := parsePublic(s)
	if err != nil {
		return nil, err
	}
	r, ok := k.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("the key is not an RSA public key")
	}
	return r, nil
}

func hashParam() tools.Param {
	return tools.Param{Name: "hash", Type: "str", Default: "sha256", Choices: []string{"sha256", "sha384", "sha512", "sha1"}, Doc: "Hash function (default sha256)."}
}

func encodeKeyPair(priv, pub any) (map[string]string, error) {
	pd, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, err
	}
	ud, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return nil, err
	}
	return map[string]string{"private_pem": pemBlock("PRIVATE KEY", pd), "public_pem": pemBlock("PUBLIC KEY", ud)}, nil
}

// keyBytes reads a hex key or, if it looks like PEM, returns nil with pem=true.
func keyArg(s string) (b []byte, isPEM bool, err error) {
	if strings.Contains(s, "-----BEGIN") {
		return nil, true, nil
	}
	b, err = decodeBytes(s, "hex")
	return b, false, err
}

func init() {
	// ---- RSA ----
	reg(tools.Func{
		Name: "rsa_generate", Title: "Generate RSA key pair",
		Doc:     "Generates an RSA key pair and returns a dict with public_pem (PKIX), private_pem (PKCS#8) and bits. bits must be 2048 (default), 3072 or 4096. Slow: 2048 takes about a second, 4096 can take several seconds to a minute; the private key is random every call.",
		Params:  []tools.Param{{Name: "bits", Type: "int", Default: int64(2048), Choices: []string{"2048", "3072", "4096"}, Doc: "Key size: 2048 (default), 3072 or 4096."}},
		Returns: "dict[str, str]",
		Fn: func(a tools.Args) (any, error) {
			b := a.Int("bits")
			if b != 2048 && b != 3072 && b != 4096 {
				return nil, fmt.Errorf("bits must be 2048, 3072 or 4096 (got %d)", b)
			}
			k, err := rsa.GenerateKey(rand.Reader, int(b))
			if err != nil {
				return nil, err
			}
			m, err := encodeKeyPair(k, &k.PublicKey)
			if err != nil {
				return nil, err
			}
			m["bits"] = fmt.Sprint(b)
			return m, nil
		},
		Example: `cipher.rsa_generate(2048)`,
		Tags:    []string{"rsa", "generate", "keypair", "key pair", "pem", "public key", "private key", "openssl genrsa", "asymmetric"},
	})
	reg(tools.Func{
		Name: "rsa_encrypt", Title: "RSA encrypt (OAEP)",
		Doc: "Encrypts text with an RSA public key (PEM) using RSA-OAEP with SHA-256 (or hash). The message must be short (key bytes minus 2*hash size minus 2, i.e. 190 bytes for RSA-2048/SHA-256); longer input is an error. Output is base64 by default.",
		Params: []tools.Param{reqStr("text", "Message (UTF-8 unless input_encoding says hex or base64)."), reqStr("public_pem", "RSA public key in PEM (PKIX, PKCS#1 or a certificate)."),
			inEncParam(), hashParam(), outParam("base64", "Output format (default base64).")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			k, err := rsaPublic(a.Str("public_pem"))
			if err != nil {
				return nil, err
			}
			_, nh, err := hashFor(a.Str("hash"))
			if err != nil {
				return nil, err
			}
			d, err := decodeBytes(a.Str("text"), a.Str("input_encoding"))
			if err != nil {
				return nil, err
			}
			ct, err := rsa.EncryptOAEP(nh(), rand.Reader, k, d, nil)
			if err != nil {
				return nil, err
			}
			return encodeBytes(ct, a.Str("output"))
		},
		Example: `cipher.rsa_encrypt("hi", pem)`,
		Tags:    []string{"rsa", "oaep", "encrypt", "public key", "asymmetric"},
		Family:  "cipher", Order: 200, Keywords: []string{"rsa", "encrypt"},
	})
	reg(tools.Func{
		Name: "rsa_decrypt", Title: "RSA decrypt (OAEP)",
		Doc: "Decrypts RSA-OAEP data with an RSA private key (PEM, PKCS#8 or PKCS#1). hash must match the one used to encrypt (default sha256). A wrong key or damaged data is an error.",
		Params: []tools.Param{reqStr("data", "Ciphertext in data_encoding (base64 default)."), reqStr("private_pem", "RSA private key in PEM."),
			{Name: "data_encoding", Type: "str", Default: "base64", Choices: []string{"base64", "hex"}, Doc: "Encoding of data (default base64)."},
			hashParam(), outParam("utf-8", "How the plaintext is returned (utf-8 default).")},
		Returns: "str", Input: "data",
		Fn: func(a tools.Args) (any, error) {
			k, err := rsaPrivate(a.Str("private_pem"))
			if err != nil {
				return nil, err
			}
			_, nh, err := hashFor(a.Str("hash"))
			if err != nil {
				return nil, err
			}
			d, err := decodeBytes(a.Str("data"), a.Str("data_encoding"))
			if err != nil {
				return nil, err
			}
			pt, err := rsa.DecryptOAEP(nh(), nil, k, d, nil)
			if err != nil {
				return nil, errors.New("decryption failed: wrong key, wrong hash or corrupted data")
			}
			return encodeBytes(pt, a.Str("output"))
		},
		Example: `cipher.rsa_decrypt(ciphertext_b64, private_pem)`,
		Tags:    []string{"rsa", "oaep", "decrypt", "private key", "asymmetric"},
		Family:  "cipher", Order: 201, Keywords: []string{"rsa", "decrypt"},
		Accepts: looksEncrypted,
	})
	padParam := tools.Param{Name: "padding", Type: "str", Default: "pss", Choices: []string{"pss", "pkcs1v15"}, Doc: "Signature scheme: pss (RSASSA-PSS, default) or pkcs1v15 (RSASSA-PKCS1-v1_5)."}
	reg(tools.Func{
		Name: "rsa_sign", Title: "RSA sign",
		Doc:     "Signs text with an RSA private key (PEM), using RSASSA-PSS (default, salt length = hash length) or PKCS#1 v1.5, over the SHA-256 (or hash) digest. Output is base64 by default.",
		Params:  []tools.Param{reqStr("text", "Message to sign."), reqStr("private_pem", "RSA private key in PEM."), padParam, hashParam(), inEncParam(), outParam("base64", "Signature format (default base64).")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			k, err := rsaPrivate(a.Str("private_pem"))
			if err != nil {
				return nil, err
			}
			ch, nh, err := hashFor(a.Str("hash"))
			if err != nil {
				return nil, err
			}
			d, err := decodeBytes(a.Str("text"), a.Str("input_encoding"))
			if err != nil {
				return nil, err
			}
			dg := digestOf(nh, d)
			var sig []byte
			switch strings.ToLower(a.Str("padding")) {
			case "pss":
				sig, err = rsa.SignPSS(rand.Reader, k, ch, dg, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash})
			case "pkcs1v15", "pkcs1", "pkcs1_v1_5":
				sig, err = rsa.SignPKCS1v15(rand.Reader, k, ch, dg)
			default:
				return nil, errors.New("padding must be pss or pkcs1v15")
			}
			if err != nil {
				return nil, err
			}
			return encodeBytes(sig, a.Str("output"))
		},
		Example: `cipher.rsa_sign("message", private_pem)`,
		Tags:    []string{"rsa", "sign", "signature", "pss", "pkcs1", "asymmetric"},
	})
	reg(tools.Func{
		Name: "rsa_verify", Title: "RSA verify",
		Doc: "Verifies an RSA signature (PSS by default, or pkcs1v15) of text against a public key PEM and returns true or false. A malformed key or signature encoding is an error; a bad signature is false.",
		Params: []tools.Param{reqStr("text", "Signed message."), reqStr("signature", "Signature in signature_encoding (base64 default)."), reqStr("public_pem", "RSA public key in PEM."),
			padParam, hashParam(), inEncParam(),
			{Name: "signature_encoding", Type: "str", Default: "base64", Choices: []string{"base64", "hex"}, Doc: "Encoding of signature (default base64)."}},
		Returns: "bool",
		Fn: func(a tools.Args) (any, error) {
			k, err := rsaPublic(a.Str("public_pem"))
			if err != nil {
				return nil, err
			}
			ch, nh, err := hashFor(a.Str("hash"))
			if err != nil {
				return nil, err
			}
			d, err := decodeBytes(a.Str("text"), a.Str("input_encoding"))
			if err != nil {
				return nil, err
			}
			sig, err := decodeBytes(a.Str("signature"), a.Str("signature_encoding"))
			if err != nil {
				return nil, fmt.Errorf("signature: %v", err)
			}
			dg := digestOf(nh, d)
			switch strings.ToLower(a.Str("padding")) {
			case "pss":
				return rsa.VerifyPSS(k, ch, dg, sig, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto}) == nil, nil
			case "pkcs1v15", "pkcs1", "pkcs1_v1_5":
				return rsa.VerifyPKCS1v15(k, ch, dg, sig) == nil, nil
			}
			return nil, errors.New("padding must be pss or pkcs1v15")
		},
		Example: `cipher.rsa_verify("message", signature_b64, public_pem)`,
		Tags:    []string{"rsa", "verify", "signature", "check", "pss", "pkcs1", "asymmetric"},
	})

	// ---- Ed25519 ----
	edPriv := func(s string) (ed25519.PrivateKey, error) {
		b, isPEM, err := keyArg(s)
		if err != nil {
			return nil, err
		}
		if isPEM {
			k, err := parsePrivate(s)
			if err != nil {
				return nil, err
			}
			ek, ok := k.(ed25519.PrivateKey)
			if !ok {
				return nil, errors.New("the key is not an Ed25519 private key")
			}
			return ek, nil
		}
		switch len(b) {
		case ed25519.SeedSize:
			return ed25519.NewKeyFromSeed(b), nil
		case ed25519.PrivateKeySize:
			return ed25519.PrivateKey(b), nil
		}
		return nil, fmt.Errorf("Ed25519 private key must be a 32-byte seed or 64-byte key in hex (got %d bytes)", len(b))
	}
	edPub := func(s string) (ed25519.PublicKey, error) {
		b, isPEM, err := keyArg(s)
		if err != nil {
			return nil, err
		}
		if isPEM {
			k, err := parsePublic(s)
			if err != nil {
				return nil, err
			}
			ek, ok := k.(ed25519.PublicKey)
			if !ok {
				return nil, errors.New("the key is not an Ed25519 public key")
			}
			return ek, nil
		}
		if len(b) != ed25519.PublicKeySize {
			return nil, fmt.Errorf("Ed25519 public key must be 32 bytes in hex (got %d)", len(b))
		}
		return ed25519.PublicKey(b), nil
	}
	reg(tools.Func{
		Name: "ed25519_generate", Title: "Generate Ed25519 key pair",
		Doc:     "Generates an Ed25519 key pair and returns a dict: private_key (32-byte seed, hex), public_key (hex), private_pem (PKCS#8) and public_pem (PKIX). Random every call.",
		Returns: "dict[str, str]",
		Fn: func(tools.Args) (any, error) {
			pub, priv, err := ed25519.GenerateKey(rand.Reader)
			if err != nil {
				return nil, err
			}
			m, err := encodeKeyPair(priv, pub)
			if err != nil {
				return nil, err
			}
			m["private_key"] = hex.EncodeToString(priv.Seed())
			m["public_key"] = hex.EncodeToString(pub)
			return m, nil
		},
		Example: `cipher.ed25519_generate()`,
		Tags:    []string{"ed25519", "generate", "keypair", "eddsa", "signature key", "ssh", "curve25519"},
	})
	reg(tools.Func{
		Name: "ed25519_public_key", Title: "Ed25519 public key",
		Doc:     "Derives the Ed25519 public key (hex) from a private key given as a 32-byte hex seed, a 64-byte hex key or a PEM.",
		Params:  []tools.Param{reqStr("private_key", "Private key: 32-byte seed in hex, 64-byte hex key or PEM.")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			k, err := edPriv(a.Str("private_key"))
			if err != nil {
				return nil, err
			}
			return hex.EncodeToString(k.Public().(ed25519.PublicKey)), nil
		},
		Example: `cipher.ed25519_public_key("9d61b19deffd5a60ba844af492ec2cc44449c5697b326919703bac031cae7f60")`,
		Tags:    []string{"ed25519", "public key", "derive", "seed"},
	})
	reg(tools.Func{
		Name: "ed25519_sign", Title: "Ed25519 sign",
		Doc:     "Signs text with an Ed25519 private key (32-byte hex seed, 64-byte hex key or PEM) and returns the 64-byte signature, hex by default. Ed25519 signatures are deterministic.",
		Params:  []tools.Param{reqStr("text", "Message to sign."), reqStr("private_key", "Private key: 32-byte seed in hex, 64-byte hex key or PEM."), inEncParam(), outParam("hex", "Signature format (hex default).")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			k, err := edPriv(a.Str("private_key"))
			if err != nil {
				return nil, err
			}
			d, err := decodeBytes(a.Str("text"), a.Str("input_encoding"))
			if err != nil {
				return nil, err
			}
			return encodeBytes(ed25519.Sign(k, d), a.Str("output"))
		},
		Example: `cipher.ed25519_sign("message", "9d61b19deffd5a60ba844af492ec2cc44449c5697b326919703bac031cae7f60")`,
		Tags:    []string{"ed25519", "sign", "signature", "eddsa"},
	})
	reg(tools.Func{
		Name: "ed25519_verify", Title: "Ed25519 verify",
		Doc:     "Verifies an Ed25519 signature (hex by default) of text against a public key (32-byte hex or PEM) and returns true or false. Malformed keys are errors; a wrong signature is false.",
		Params:  []tools.Param{reqStr("text", "Signed message."), reqStr("signature", "Signature in signature_encoding (hex default)."), reqStr("public_key", "Public key: 32-byte hex or PEM."), inEncParam(), {Name: "signature_encoding", Type: "str", Default: "hex", Choices: []string{"hex", "base64"}, Doc: "Encoding of signature (hex default)."}},
		Returns: "bool",
		Fn: func(a tools.Args) (any, error) {
			k, err := edPub(a.Str("public_key"))
			if err != nil {
				return nil, err
			}
			d, err := decodeBytes(a.Str("text"), a.Str("input_encoding"))
			if err != nil {
				return nil, err
			}
			sig, err := decodeBytes(a.Str("signature"), a.Str("signature_encoding"))
			if err != nil {
				return nil, fmt.Errorf("signature: %v", err)
			}
			return ed25519.Verify(k, d, sig), nil
		},
		Example: `cipher.ed25519_verify("message", signature_hex, public_key_hex)`,
		Tags:    []string{"ed25519", "verify", "signature", "check", "eddsa"},
	})

	// ---- ECDSA ----
	curveOf := func(name string) (elliptic.Curve, error) {
		switch normAlgo(name) {
		case "", "p256", "p_256", "secp256r1", "prime256v1":
			return elliptic.P256(), nil
		case "p384", "p_384", "secp384r1":
			return elliptic.P384(), nil
		case "p521", "p_521", "secp521r1":
			return elliptic.P521(), nil
		}
		return nil, fmt.Errorf("unsupported curve %q (use p256, p384 or p521)", name)
	}
	reg(tools.Func{
		Name: "ecdsa_generate", Title: "Generate ECDSA key pair",
		Doc:     "Generates an ECDSA key pair on curve p256 (default, a.k.a. secp256r1/prime256v1), p384 or p521 and returns a dict with private_pem (PKCS#8) and public_pem (PKIX). Random every call.",
		Params:  []tools.Param{{Name: "curve", Type: "str", Default: "p256", Choices: []string{"p256", "p384", "p521"}, Doc: "Curve (default p256)."}},
		Returns: "dict[str, str]",
		Fn: func(a tools.Args) (any, error) {
			c, err := curveOf(a.Str("curve"))
			if err != nil {
				return nil, err
			}
			k, err := ecdsa.GenerateKey(c, rand.Reader)
			if err != nil {
				return nil, err
			}
			return encodeKeyPair(k, &k.PublicKey)
		},
		Example: `cipher.ecdsa_generate("p256")`,
		Tags:    []string{"ecdsa", "generate", "keypair", "p256", "secp256r1", "prime256v1", "elliptic curve", "ec"},
	})
	reg(tools.Func{
		Name: "ecdsa_sign", Title: "ECDSA sign",
		Doc:     "Signs text with an ECDSA private key (PEM) over the SHA-256 (or hash) digest and returns the ASN.1 DER signature, base64 by default. Randomized: two calls give different valid signatures.",
		Params:  []tools.Param{reqStr("text", "Message to sign."), reqStr("private_pem", "ECDSA private key in PEM."), hashParam(), inEncParam(), outParam("base64", "Signature format (base64 default).")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			k, err := parsePrivate(a.Str("private_pem"))
			if err != nil {
				return nil, err
			}
			ek, ok := k.(*ecdsa.PrivateKey)
			if !ok {
				return nil, errors.New("the key is not an ECDSA private key")
			}
			_, nh, err := hashFor(a.Str("hash"))
			if err != nil {
				return nil, err
			}
			d, err := decodeBytes(a.Str("text"), a.Str("input_encoding"))
			if err != nil {
				return nil, err
			}
			sig, err := ecdsa.SignASN1(rand.Reader, ek, digestOf(nh, d))
			if err != nil {
				return nil, err
			}
			return encodeBytes(sig, a.Str("output"))
		},
		Example: `cipher.ecdsa_sign("message", private_pem)`,
		Tags:    []string{"ecdsa", "sign", "signature", "p256", "elliptic curve", "der"},
	})
	reg(tools.Func{
		Name: "ecdsa_verify", Title: "ECDSA verify",
		Doc:     "Verifies an ASN.1 DER ECDSA signature (base64 by default) of text against a public key PEM and returns true or false. Malformed keys are errors; a bad signature is false.",
		Params:  []tools.Param{reqStr("text", "Signed message."), reqStr("signature", "DER signature in signature_encoding."), reqStr("public_pem", "ECDSA public key in PEM."), hashParam(), inEncParam(), {Name: "signature_encoding", Type: "str", Default: "base64", Choices: []string{"base64", "hex"}, Doc: "Encoding of signature (base64 default)."}},
		Returns: "bool",
		Fn: func(a tools.Args) (any, error) {
			k, err := parsePublic(a.Str("public_pem"))
			if err != nil {
				return nil, err
			}
			ek, ok := k.(*ecdsa.PublicKey)
			if !ok {
				return nil, errors.New("the key is not an ECDSA public key")
			}
			_, nh, err := hashFor(a.Str("hash"))
			if err != nil {
				return nil, err
			}
			d, err := decodeBytes(a.Str("text"), a.Str("input_encoding"))
			if err != nil {
				return nil, err
			}
			sig, err := decodeBytes(a.Str("signature"), a.Str("signature_encoding"))
			if err != nil {
				return nil, fmt.Errorf("signature: %v", err)
			}
			return ecdsa.VerifyASN1(ek, digestOf(nh, d), sig), nil
		},
		Example: `cipher.ecdsa_verify("message", signature_b64, public_pem)`,
		Tags:    []string{"ecdsa", "verify", "signature", "check", "p256", "elliptic curve"},
	})

	// ---- X25519 ----
	reg(tools.Func{
		Name: "x25519_generate", Title: "Generate X25519 key pair",
		Doc:     "Generates an X25519 (Curve25519 Diffie-Hellman) key pair and returns a dict with private_key and public_key as 32-byte hex. Random every call.",
		Returns: "dict[str, str]",
		Fn: func(tools.Args) (any, error) {
			k, err := ecdh.X25519().GenerateKey(rand.Reader)
			if err != nil {
				return nil, err
			}
			return map[string]string{"private_key": hex.EncodeToString(k.Bytes()), "public_key": hex.EncodeToString(k.PublicKey().Bytes())}, nil
		},
		Example: `cipher.x25519_generate()`,
		Tags:    []string{"x25519", "curve25519", "diffie hellman", "ecdh", "key exchange", "generate"},
	})
	reg(tools.Func{
		Name: "x25519_shared_secret", Title: "X25519 shared secret",
		Doc:     "Computes the X25519 Diffie-Hellman shared secret (32 bytes, hex) from your private key and the peer's public key, both 32-byte hex. Both sides derive the same value. All-zero (low-order) peer keys are rejected.",
		Params:  []tools.Param{reqStr("private_key", "Your private key, 32 bytes hex."), reqStr("peer_public_key", "The peer's public key, 32 bytes hex.")},
		Returns: "str", Input: "private_key",
		Fn: func(a tools.Args) (any, error) {
			pb, err := decodeBytes(a.Str("private_key"), "hex")
			if err != nil {
				return nil, err
			}
			qb, err := decodeBytes(a.Str("peer_public_key"), "hex")
			if err != nil {
				return nil, err
			}
			priv, err := ecdh.X25519().NewPrivateKey(pb)
			if err != nil {
				return nil, fmt.Errorf("private key: %v", err)
			}
			pub, err := ecdh.X25519().NewPublicKey(qb)
			if err != nil {
				return nil, fmt.Errorf("peer public key: %v", err)
			}
			s, err := priv.ECDH(pub)
			if err != nil {
				return nil, err
			}
			return hex.EncodeToString(s), nil
		},
		Example: `cipher.x25519_shared_secret(my_private_hex, peer_public_hex)`,
		Tags:    []string{"x25519", "curve25519", "diffie hellman", "ecdh", "shared secret", "key exchange", "rfc7748"},
	})

	// ---- PEM ----
	reg(tools.Func{
		Name: "pem_decode", Title: "Decode PEM",
		Doc:     "Parses the first PEM block of the text and returns a dict with type (e.g. 'PRIVATE KEY', 'CERTIFICATE'), bytes_hex (the DER body in hex), length (bytes), count (number of PEM blocks found) and any headers as header_<name>. Text with no PEM block is an error.",
		Params:  []tools.Param{reqStr("pem", "PEM text including the BEGIN/END lines.")},
		Returns: "dict[str, str]",
		Fn: func(a tools.Args) (any, error) {
			rest := []byte(a.Str("pem"))
			blk, rest := pem.Decode(rest)
			if blk == nil {
				return nil, errors.New("no PEM block found (expected -----BEGIN ...----- lines)")
			}
			m := map[string]string{"type": blk.Type, "bytes_hex": hex.EncodeToString(blk.Bytes), "length": fmt.Sprint(len(blk.Bytes))}
			n := 1
			for {
				var b *pem.Block
				b, rest = pem.Decode(rest)
				if b == nil {
					break
				}
				n++
			}
			m["count"] = fmt.Sprint(n)
			for k, v := range blk.Headers {
				m["header_"+k] = v
			}
			return m, nil
		},
		Example: `cipher.pem_decode(pem_text)`,
		Tags:    []string{"pem", "decode", "der", "certificate", "key", "base64", "openssl", "parse"},
		Family:  "cipher", Order: 210, Keywords: []string{"pem"},
		Accepts: func(s string) bool { return strings.Contains(s, "-----BEGIN") },
	})
	reg(tools.Func{
		Name: "pem_encode", Title: "Encode PEM",
		Doc:     "Wraps hex bytes in a PEM block of the given type (e.g. 'CERTIFICATE', 'PUBLIC KEY', 'PRIVATE KEY') with 64-column base64 lines. Odd or non-hex input and an empty type are errors.",
		Params:  []tools.Param{reqStr("type", "PEM label, e.g. CERTIFICATE or PUBLIC KEY."), reqStr("hex", "DER bytes as hex.")},
		Returns: "str", Input: "hex",
		Fn: func(a tools.Args) (any, error) {
			if strings.TrimSpace(a.Str("type")) == "" {
				return nil, errors.New("type must not be empty")
			}
			b, err := decodeBytes(a.Str("hex"), "hex")
			if err != nil {
				return nil, err
			}
			return pemBlock(strings.TrimSpace(a.Str("type")), b), nil
		},
		Example: `cipher.pem_encode("PUBLIC KEY", "3059301306...")`,
		Tags:    []string{"pem", "encode", "der", "certificate", "key", "base64", "openssl"},
	})
}

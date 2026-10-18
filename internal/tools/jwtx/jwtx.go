// Package jwtx registers the "jwt" tool group: decode, sign and verify JSON
// Web Tokens. Pure: keys are passed as strings and nothing is fetched.
package jwtx

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	"crypto/hmac"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"crypto/subtle"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"hash"
	"math"
	"math/big"
	"strings"
	"time"

	"github.com/avestura/zagros/internal/tools"
)

var algs = []string{"HS256", "HS384", "HS512", "RS256", "RS384", "RS512", "PS256", "PS384", "PS512", "ES256", "ES384", "ES512", "EdDSA"}

func init() {
	tools.Register(tools.Func{
		Group: "jwt", Name: "decode", Title: "Decode JWT (no verification)",
		Doc:     "Split a JWT into a dict with header, payload (both dicts), signature (the base64url text) and alg. The signature is NOT checked: never trust the payload of a token you have not passed to jwt.verify. Raises when the token does not have three dot-separated base64url parts holding JSON objects.",
		Terms:   []string{"read jwt", "inspect token", "jwt claims", "jwt payload", "bearer token", "id token", "access token", "jwt debugger", "parse jwt"},
		Params:  []tools.Param{tools.Req("token", "str", "The compact JWT; a leading \"Bearer \" is ignored.")},
		Returns: "any",
		Example: `jwt.decode("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.sig")`,
		Tags:    []string{"jwt", "decode", "token", "claims", "inspect"},
		Fn: func(a tools.Args) (any, error) {
			tk, err := split(a.Str("token"))
			if err != nil {
				return nil, err
			}
			return tk.dict(), nil
		},
	})
	tools.Register(tools.Func{
		Group: "jwt", Name: "sign", Title: "Sign a JWT",
		Doc:   "Create a signed JWT from a payload dict. alg is one of HS256/384/512 (key is the shared secret), RS256/384/512 and PS256/384/512 (key is an RSA private key in PEM), ES256/384/512 (EC private key in PEM, curve must match) or EdDSA (Ed25519 private key in PEM). expires_in adds exp, issued_at adds iat, both from now (0 means the current time). header adds or overrides header fields; alg and typ are always set. Raises on a key that does not fit the algorithm or a payload that is not a dict.",
		Terms: []string{"create jwt", "generate token", "make jwt", "encode jwt", "jws", "hs256", "rs256", "access token", "issue token"},
		Params: []tools.Param{
			tools.Req("payload", "any", "The claims, a dict."),
			tools.Req("key", "str", "Shared secret for HS*, or a private key in PEM for the others."),
			{Name: "alg", Type: "str", Default: "HS256", Choices: algs, Doc: "The signing algorithm."},
			tools.Opt("header", "any", "Extra header fields, a dict.", nil),
			tools.Opt("expires_in", "int", "Seconds until the token expires; 0 adds no exp.", int64(0)),
			tools.Opt("issued_at", "bool", "Add the iat claim.", false),
			tools.Opt("now", "int", "Unix time to use for iat and exp; 0 reads the clock.", int64(0)),
			{Name: "key_encoding", Type: "str", Default: "utf-8", Choices: []string{"utf-8", "hex", "base64", "base64url"}, Doc: "How an HS* secret is written; PEM keys ignore it."},
		},
		Returns: "str",
		Example: `jwt.sign({"sub": "123", "name": "Ada"}, "secret", expires_in=3600)`,
		Tags:    []string{"jwt", "sign", "token", "hs256", "rs256", "create"},
		Fn:      sign,
	})
	tools.Register(tools.Func{
		Group: "jwt", Name: "verify", Title: "Verify a JWT",
		Doc:   "Check a JWT's signature and time claims and return what jwt.decode returns. key is the shared secret for HS*, or a public key, certificate or private key in PEM for the others. alg pins the accepted algorithm (strongly advised); empty trusts the token's own header, but \"none\" is always rejected and an HS* token is refused when the key is PEM, which blocks the key-confusion attack. exp and nbf are checked with leeway seconds of slack (check_time=False skips them). Raises with the reason when anything fails.",
		Terms: []string{"validate jwt", "check token", "verify signature", "jwt valid", "authenticate token", "is token expired", "check signature"},
		Params: []tools.Param{
			tools.Req("token", "str", "The compact JWT."),
			tools.Req("key", "str", "Shared secret for HS*, or a PEM public key, certificate or private key."),
			{Name: "alg", Type: "str", Default: "", Choices: algs, Doc: "The only algorithm to accept; empty accepts the token's alg."},
			tools.Opt("check_time", "bool", "Check exp and nbf.", true),
			tools.Opt("leeway", "int", "Seconds of tolerance for exp and nbf.", int64(0)),
			tools.Opt("now", "int", "Unix time to compare against; 0 reads the clock.", int64(0)),
			{Name: "key_encoding", Type: "str", Default: "utf-8", Choices: []string{"utf-8", "hex", "base64", "base64url"}, Doc: "How an HS* secret is written; PEM keys ignore it."},
		},
		Returns: "any",
		Example: `jwt.verify(jwt.sign({"sub": "1"}, "secret"), "secret", alg="HS256")`,
		Tags:    []string{"jwt", "verify", "validate", "signature", "check"},
		Fn: func(a tools.Args) (any, error) {
			tk, err := verify(a)
			if err != nil {
				return nil, err
			}
			return tk.dict(), nil
		},
	})
	tools.Register(tools.Func{
		Group: "jwt", Name: "is_valid", Title: "Is JWT valid",
		Doc:   "Report whether a JWT passes jwt.verify with these arguments, as true or false instead of an error.",
		Terms: []string{"token valid", "check jwt", "signature ok", "is expired"},
		Params: []tools.Param{
			tools.Req("token", "str", "The compact JWT."),
			tools.Req("key", "str", "Shared secret for HS*, or a PEM public key, certificate or private key."),
			{Name: "alg", Type: "str", Default: "", Choices: algs, Doc: "The only algorithm to accept; empty accepts the token's alg."},
			tools.Opt("check_time", "bool", "Check exp and nbf.", true),
			tools.Opt("leeway", "int", "Seconds of tolerance for exp and nbf.", int64(0)),
			tools.Opt("now", "int", "Unix time to compare against; 0 reads the clock.", int64(0)),
			{Name: "key_encoding", Type: "str", Default: "utf-8", Choices: []string{"utf-8", "hex", "base64", "base64url"}, Doc: "How an HS* secret is written; PEM keys ignore it."},
		},
		Returns: "bool",
		Example: `jwt.is_valid("a.b.c", "secret")`,
		Tags:    []string{"jwt", "valid", "verify", "check"},
		Fn: func(a tools.Args) (any, error) {
			_, err := verify(a)
			return err == nil, nil
		},
	})
}

type token struct {
	header, payload *tools.Dict
	signature       string
	signingInput    string
	sigBytes        []byte
}

func (t *token) alg() string {
	v, _ := t.header.Get("alg")
	s, _ := v.(string)
	return s
}

func (t *token) dict() *tools.Dict {
	d := tools.NewDict()
	d.Set("header", t.header)
	d.Set("payload", t.payload)
	d.Set("signature", t.signature)
	d.Set("alg", t.alg())
	return d
}

func b64(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(strings.TrimRight(s, "="))
}

func split(raw string) (*token, error) {
	raw = strings.TrimSpace(raw)
	if i := strings.IndexAny(raw, " \t"); i >= 0 && strings.EqualFold(raw[:i], "bearer") {
		raw = strings.TrimSpace(raw[i:])
	}
	if raw == "" {
		return nil, fmt.Errorf("empty token")
	}
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("a JWT has three dot-separated parts, this has %d", len(parts))
	}
	obj := func(name, s string) (*tools.Dict, error) {
		b, err := b64(s)
		if err != nil {
			return nil, fmt.Errorf("the %s is not base64url", name)
		}
		v, err := tools.ParseJSON(string(b))
		if err != nil {
			return nil, fmt.Errorf("the %s is not valid JSON: %v", name, err)
		}
		d, ok := v.(*tools.Dict)
		if !ok {
			return nil, fmt.Errorf("the %s is not a JSON object", name)
		}
		return d, nil
	}
	h, err := obj("header", parts[0])
	if err != nil {
		return nil, err
	}
	p, err := obj("payload", parts[1])
	if err != nil {
		return nil, err
	}
	t := &token{header: h, payload: p, signature: parts[2], signingInput: parts[0] + "." + parts[1]}
	if parts[2] != "" {
		if t.sigBytes, err = b64(parts[2]); err != nil {
			return nil, fmt.Errorf("the signature is not base64url")
		}
	}
	return t, nil
}

func secret(key, enc string) ([]byte, error) {
	switch enc {
	case "", "utf-8":
		return []byte(key), nil
	case "hex":
		b, err := hex.DecodeString(strings.ReplaceAll(key, " ", ""))
		if err != nil {
			return nil, fmt.Errorf("key is not valid hex")
		}
		return b, nil
	case "base64":
		b, err := base64.StdEncoding.DecodeString(strings.TrimSpace(key))
		if err != nil {
			b, err = base64.RawStdEncoding.DecodeString(strings.TrimSpace(key))
		}
		if err != nil {
			return nil, fmt.Errorf("key is not valid base64")
		}
		return b, nil
	case "base64url":
		b, err := b64(strings.TrimSpace(key))
		if err != nil {
			return nil, fmt.Errorf("key is not valid base64url")
		}
		return b, nil
	}
	return nil, fmt.Errorf("key_encoding must be utf-8, hex, base64 or base64url")
}

type spec struct {
	family string // HS, RS, PS, ES, Ed
	hash   crypto.Hash
	curve  elliptic.Curve
}

func specOf(alg string) (spec, error) {
	h := map[string]crypto.Hash{"256": crypto.SHA256, "384": crypto.SHA384, "512": crypto.SHA512}
	if alg == "EdDSA" {
		return spec{family: "Ed"}, nil
	}
	if len(alg) == 5 {
		if hh, ok := h[alg[2:]]; ok {
			switch alg[:2] {
			case "HS", "RS", "PS":
				return spec{family: alg[:2], hash: hh}, nil
			case "ES":
				c := map[string]elliptic.Curve{"256": elliptic.P256(), "384": elliptic.P384(), "512": elliptic.P521()}[alg[2:]]
				return spec{family: "ES", hash: hh, curve: c}, nil
			}
		}
	}
	return spec{}, fmt.Errorf("unsupported algorithm %q (use one of %s)", alg, strings.Join(algs, ", "))
}

func (s spec) newHash() hash.Hash {
	switch s.hash {
	case crypto.SHA384:
		return sha512.New384()
	case crypto.SHA512:
		return sha512.New()
	}
	return sha256.New()
}

func (s spec) digest(data string) []byte {
	h := s.newHash()
	h.Write([]byte(data))
	return h.Sum(nil)
}

func decodePEM(key string) ([]byte, string, error) {
	blk, _ := pem.Decode([]byte(strings.TrimSpace(key)))
	if blk == nil {
		return nil, "", fmt.Errorf("the key is not PEM (expected -----BEGIN ...-----)")
	}
	return blk.Bytes, blk.Type, nil
}

func privateKey(key string) (any, error) {
	der, _, err := decodePEM(key)
	if err != nil {
		return nil, err
	}
	if k, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		return k, nil
	}
	if k, err := x509.ParsePKCS1PrivateKey(der); err == nil {
		return k, nil
	}
	if k, err := x509.ParseECPrivateKey(der); err == nil {
		return k, nil
	}
	return nil, fmt.Errorf("the PEM is not a supported private key (PKCS#8, PKCS#1 RSA or SEC1 EC)")
}

func publicKey(key string) (any, error) {
	der, typ, err := decodePEM(key)
	if err != nil {
		return nil, err
	}
	if typ == "CERTIFICATE" {
		c, err := x509.ParseCertificate(der)
		if err != nil {
			return nil, fmt.Errorf("the certificate cannot be read: %v", err)
		}
		return c.PublicKey, nil
	}
	if k, err := x509.ParsePKIXPublicKey(der); err == nil {
		return k, nil
	}
	if k, err := x509.ParsePKCS1PublicKey(der); err == nil {
		return k, nil
	}
	pk, err := privateKey(key)
	if err != nil {
		return nil, fmt.Errorf("the PEM is not a public key, certificate or private key")
	}
	switch k := pk.(type) {
	case *rsa.PrivateKey:
		return &k.PublicKey, nil
	case *ecdsa.PrivateKey:
		return &k.PublicKey, nil
	case ed25519.PrivateKey:
		return k.Public(), nil
	}
	return nil, fmt.Errorf("unsupported key type")
}

func sign(a tools.Args) (any, error) {
	payload, ok := a["payload"].(*tools.Dict)
	if !ok {
		return nil, fmt.Errorf("payload must be a dict")
	}
	alg := a.Str("alg")
	sp, err := specOf(alg)
	if err != nil {
		return nil, err
	}
	now := a.Int("now")
	if now == 0 {
		now = time.Now().Unix()
	}
	claims := tools.NewDict()
	for _, k := range payload.Keys() {
		v, _ := payload.Get(k)
		claims.Set(k, v)
	}
	if a.Bool("issued_at") {
		claims.Set("iat", now)
	}
	if n := a.Int("expires_in"); n != 0 {
		if n < 0 {
			return nil, fmt.Errorf("expires_in must not be negative")
		}
		claims.Set("exp", now+n)
	}
	header := tools.NewDict()
	header.Set("alg", alg)
	header.Set("typ", "JWT")
	if h, ok := a["header"]; ok && h != nil {
		hd, ok := h.(*tools.Dict)
		if !ok {
			return nil, fmt.Errorf("header must be a dict")
		}
		for _, k := range hd.Keys() {
			if k == "alg" || k == "typ" {
				continue
			}
			v, _ := hd.Get(k)
			header.Set(k, v)
		}
	}
	hb, err := json.Marshal(header)
	if err != nil {
		return nil, fmt.Errorf("header cannot be encoded: %v", err)
	}
	pb, err := json.Marshal(claims)
	if err != nil {
		return nil, fmt.Errorf("payload cannot be encoded as JSON: %v", err)
	}
	input := base64.RawURLEncoding.EncodeToString(hb) + "." + base64.RawURLEncoding.EncodeToString(pb)
	var sig []byte
	switch sp.family {
	case "HS":
		k, err := secret(a.Str("key"), a.Str("key_encoding"))
		if err != nil {
			return nil, err
		}
		if len(k) == 0 {
			return nil, fmt.Errorf("the secret is empty")
		}
		m := hmac.New(sp.newHash, k)
		m.Write([]byte(input))
		sig = m.Sum(nil)
	default:
		pk, err := privateKey(a.Str("key"))
		if err != nil {
			return nil, err
		}
		if sig, err = signAsym(sp, alg, pk, input); err != nil {
			return nil, err
		}
	}
	return input + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func signAsym(sp spec, alg string, pk any, input string) ([]byte, error) {
	switch sp.family {
	case "RS", "PS":
		k, ok := pk.(*rsa.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("%s needs an RSA private key", alg)
		}
		if k.N.BitLen() < 2048 {
			return nil, fmt.Errorf("RSA keys under 2048 bits are not accepted")
		}
		if sp.family == "RS" {
			return rsa.SignPKCS1v15(crand.Reader, k, sp.hash, sp.digest(input))
		}
		return rsa.SignPSS(crand.Reader, k, sp.hash, sp.digest(input), &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash})
	case "ES":
		k, ok := pk.(*ecdsa.PrivateKey)
		if !ok || k.Curve != sp.curve {
			return nil, fmt.Errorf("%s needs an EC private key on curve %s", alg, sp.curve.Params().Name)
		}
		r, s, err := ecdsa.Sign(crand.Reader, k, sp.digest(input))
		if err != nil {
			return nil, err
		}
		n := (sp.curve.Params().BitSize + 7) / 8
		out := make([]byte, 2*n)
		r.FillBytes(out[:n])
		s.FillBytes(out[n:])
		return out, nil
	case "Ed":
		k, ok := pk.(ed25519.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("EdDSA needs an Ed25519 private key")
		}
		return ed25519.Sign(k, []byte(input)), nil
	}
	return nil, fmt.Errorf("unsupported algorithm")
}

func verify(a tools.Args) (*token, error) {
	tk, err := split(a.Str("token"))
	if err != nil {
		return nil, err
	}
	alg := tk.alg()
	if alg == "" {
		return nil, fmt.Errorf("the token header has no alg")
	}
	if strings.EqualFold(alg, "none") {
		return nil, fmt.Errorf("unsigned tokens (alg none) are never accepted")
	}
	if want := a.Str("alg"); want != "" && want != alg {
		return nil, fmt.Errorf("the token uses %s but %s is required", alg, want)
	}
	sp, err := specOf(alg)
	if err != nil {
		return nil, err
	}
	key := a.Str("key")
	if len(tk.sigBytes) == 0 {
		return nil, fmt.Errorf("the token has no signature")
	}
	ok := false
	switch sp.family {
	case "HS":
		if strings.Contains(key, "-----BEGIN") {
			return nil, fmt.Errorf("an %s token cannot be checked with a PEM key; pin alg to the asymmetric algorithm you expect", alg)
		}
		k, err := secret(key, a.Str("key_encoding"))
		if err != nil {
			return nil, err
		}
		m := hmac.New(sp.newHash, k)
		m.Write([]byte(tk.signingInput))
		ok = subtle.ConstantTimeCompare(m.Sum(nil), tk.sigBytes) == 1
	default:
		pk, err := publicKey(key)
		if err != nil {
			return nil, err
		}
		if ok, err = verifyAsym(sp, alg, pk, tk); err != nil {
			return nil, err
		}
	}
	if !ok {
		return nil, fmt.Errorf("the signature does not match")
	}
	if a.Bool("check_time") {
		now := a.Int("now")
		if now == 0 {
			now = time.Now().Unix()
		}
		leeway := a.Int("leeway")
		if v, has := tk.payload.Get("exp"); has {
			exp, isNum := number(v)
			if !isNum {
				return nil, fmt.Errorf("the exp claim is not a number")
			}
			if float64(now) >= exp+float64(leeway) {
				return nil, fmt.Errorf("the token expired %s ago", ago(float64(now)-exp))
			}
		}
		if v, has := tk.payload.Get("nbf"); has {
			nbf, isNum := number(v)
			if !isNum {
				return nil, fmt.Errorf("the nbf claim is not a number")
			}
			if float64(now)+float64(leeway) < nbf {
				return nil, fmt.Errorf("the token is not valid yet (nbf is %s away)", ago(nbf-float64(now)))
			}
		}
	}
	return tk, nil
}

func number(v any) (float64, bool) {
	switch x := v.(type) {
	case int64:
		return float64(x), true
	case float64:
		return x, !math.IsNaN(x)
	}
	return 0, false
}

func ago(sec float64) string {
	d := time.Duration(sec) * time.Second
	if d < time.Second {
		return "less than a second"
	}
	return d.String()
}

func verifyAsym(sp spec, alg string, pk any, tk *token) (bool, error) {
	switch sp.family {
	case "RS", "PS":
		k, ok := pk.(*rsa.PublicKey)
		if !ok {
			return false, fmt.Errorf("%s needs an RSA public key", alg)
		}
		if k.N.BitLen() < 2048 {
			return false, fmt.Errorf("RSA keys under 2048 bits are not accepted")
		}
		if sp.family == "RS" {
			return rsa.VerifyPKCS1v15(k, sp.hash, sp.digest(tk.signingInput), tk.sigBytes) == nil, nil
		}
		return rsa.VerifyPSS(k, sp.hash, sp.digest(tk.signingInput), tk.sigBytes, &rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthAuto}) == nil, nil
	case "ES":
		k, ok := pk.(*ecdsa.PublicKey)
		if !ok || k.Curve != sp.curve {
			return false, fmt.Errorf("%s needs an EC public key on curve %s", alg, sp.curve.Params().Name)
		}
		n := (sp.curve.Params().BitSize + 7) / 8
		if len(tk.sigBytes) != 2*n {
			return false, nil
		}
		r := new(big.Int).SetBytes(tk.sigBytes[:n])
		s := new(big.Int).SetBytes(tk.sigBytes[n:])
		return ecdsa.Verify(k, sp.digest(tk.signingInput), r, s), nil
	case "Ed":
		k, ok := pk.(ed25519.PublicKey)
		if !ok {
			return false, fmt.Errorf("EdDSA needs an Ed25519 public key")
		}
		return ed25519.Verify(k, []byte(tk.signingInput), tk.sigBytes), nil
	}
	return false, fmt.Errorf("unsupported algorithm")
}

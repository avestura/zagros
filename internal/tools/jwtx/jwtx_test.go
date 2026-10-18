package jwtx

import (
	"crypto/ecdsa"
	"crypto/ed25519"
	"crypto/elliptic"
	crand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"math/rand"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "jwt") }

const sample = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

func TestDecodeAndKnownVector(t *testing.T) {
	got := toolstest.Compact(toolstest.Must(t, "jwt", "decode", "token", "Bearer "+sample))
	want := `{"header":{"alg":"HS256","typ":"JWT"},"payload":{"sub":"1234567890","name":"John Doe","iat":1516239022},` +
		`"signature":"SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c","alg":"HS256"}`
	if strings.ReplaceAll(got, " ", "") != strings.ReplaceAll(want, " ", "") {
		t.Errorf("got  %s\nwant %s", got, want)
	}
	// the jwt.io example signs to exactly this token
	payload := toolstest.JSONValue(t, `{"sub":"1234567890","name":"John Doe","iat":1516239022}`)
	if tok := toolstest.Must(t, "jwt", "sign", "payload", payload, "key", "your-256-bit-secret"); tok != sample {
		t.Errorf("sign differs from the reference token:\n%v", tok)
	}
	if v := toolstest.Must(t, "jwt", "is_valid", "token", sample, "key", "your-256-bit-secret", "alg", "HS256", "check_time", false); v != true {
		t.Error("reference token must verify")
	}
	for _, bad := range []string{"your-256-bit-secreT", ""} {
		if v := toolstest.Must(t, "jwt", "is_valid", "token", sample, "key", bad); v != false {
			t.Errorf("key %q must not verify", bad)
		}
	}
}

func TestDecodeErrors(t *testing.T) {
	enc := func(s string) string { return base64.RawURLEncoding.EncodeToString([]byte(s)) }
	for _, bad := range []string{"", "a.b", "a.b.c.d", "!!.!!.x", enc("[1]") + "." + enc("{}") + ".", enc("{}") + "." + enc("nope") + ".", enc("{}") + "." + enc("{}") + ".@@"} {
		toolstest.Fails(t, "jwt", "decode", "token", bad)
	}
	// unsigned tokens decode but never verify
	unsigned := enc(`{"alg":"none"}`) + "." + enc(`{"a":1}`) + "."
	toolstest.Must(t, "jwt", "decode", "token", unsigned)
	toolstest.Fails(t, "jwt", "verify", "token", unsigned, "key", "")
	toolstest.Fails(t, "jwt", "verify", "token", enc(`{"alg":"NONE"}`)+"."+enc(`{}`)+".x", "key", "k")
	toolstest.Fails(t, "jwt", "verify", "token", enc(`{"alg":"HS999"}`)+"."+enc(`{}`)+".AAAA", "key", "k")
}

func TestTimeClaims(t *testing.T) {
	payload := toolstest.JSONValue(t, `{"sub":"u"}`)
	tok := toolstest.Must(t, "jwt", "sign", "payload", payload, "key", "k", "expires_in", int64(60), "issued_at", true, "now", int64(1000)).(string)
	d := toolstest.Compact(toolstest.Must(t, "jwt", "decode", "token", tok))
	if !strings.Contains(d, `"iat":1000`) || !strings.Contains(d, `"exp":1060`) {
		t.Errorf("claims: %s", d)
	}
	toolstest.Must(t, "jwt", "verify", "token", tok, "key", "k", "now", int64(1059))
	toolstest.Fails(t, "jwt", "verify", "token", tok, "key", "k", "now", int64(1060))
	toolstest.Must(t, "jwt", "verify", "token", tok, "key", "k", "now", int64(1100), "leeway", int64(60))
	toolstest.Must(t, "jwt", "verify", "token", tok, "key", "k", "now", int64(99999), "check_time", false)
	nbf := toolstest.Must(t, "jwt", "sign", "payload", toolstest.JSONValue(t, `{"nbf":5000}`), "key", "k").(string)
	toolstest.Fails(t, "jwt", "verify", "token", nbf, "key", "k", "now", int64(4000))
	toolstest.Must(t, "jwt", "verify", "token", nbf, "key", "k", "now", int64(5000))
	toolstest.Fails(t, "jwt", "verify", "token", tok, "key", "k", "alg", "HS512", "now", int64(1000))
	toolstest.Fails(t, "jwt", "sign", "payload", payload, "key", "k", "expires_in", int64(-1))
}

func TestSignErrors(t *testing.T) {
	p := toolstest.JSONValue(t, `{"a":1}`)
	toolstest.Fails(t, "jwt", "sign", "payload", "x", "key", "k")
	toolstest.Fails(t, "jwt", "sign", "payload", p, "key", "")
	toolstest.Fails(t, "jwt", "sign", "payload", p, "key", "k", "alg", "none")
	toolstest.Fails(t, "jwt", "sign", "payload", p, "key", "not pem", "alg", "RS256")
	toolstest.Fails(t, "jwt", "sign", "payload", p, "key", "k", "header", "x")
	toolstest.Fails(t, "jwt", "sign", "payload", p, "key", "zz", "key_encoding", "hex")
}

func pemOf(typ string, der []byte) string {
	return string(pem.EncodeToMemory(&pem.Block{Type: typ, Bytes: der}))
}

func TestAsymmetric(t *testing.T) {
	rk, _ := rsa.GenerateKey(crand.Reader, 2048)
	rsaPriv := pemOf("RSA PRIVATE KEY", x509.MarshalPKCS1PrivateKey(rk))
	rsaPubDER, _ := x509.MarshalPKIXPublicKey(&rk.PublicKey)
	rsaPub := pemOf("PUBLIC KEY", rsaPubDER)

	ek, _ := ecdsa.GenerateKey(elliptic.P256(), crand.Reader)
	ecDER, _ := x509.MarshalPKCS8PrivateKey(ek)
	ecPriv := pemOf("PRIVATE KEY", ecDER)
	ecPubDER, _ := x509.MarshalPKIXPublicKey(&ek.PublicKey)
	ecPub := pemOf("PUBLIC KEY", ecPubDER)

	e3, _ := ecdsa.GenerateKey(elliptic.P384(), crand.Reader)
	e3DER, _ := x509.MarshalPKCS8PrivateKey(e3)
	e3Priv := pemOf("PRIVATE KEY", e3DER)

	pub, edk, _ := ed25519.GenerateKey(crand.Reader)
	edDER, _ := x509.MarshalPKCS8PrivateKey(edk)
	edPriv := pemOf("PRIVATE KEY", edDER)
	edPubDER, _ := x509.MarshalPKIXPublicKey(pub)
	edPub := pemOf("PUBLIC KEY", edPubDER)

	p := toolstest.JSONValue(t, `{"sub":"u","n":1.5,"list":[1,"a"]}`)
	cases := []struct{ alg, priv, pub string }{
		{"RS256", rsaPriv, rsaPub}, {"RS512", rsaPriv, rsaPub}, {"PS256", rsaPriv, rsaPub}, {"PS384", rsaPriv, rsaPriv},
		{"ES256", ecPriv, ecPub}, {"ES384", e3Priv, e3Priv}, {"EdDSA", edPriv, edPub},
	}
	for _, c := range cases {
		tok := toolstest.Must(t, "jwt", "sign", "payload", p, "key", c.priv, "alg", c.alg).(string)
		got := toolstest.Must(t, "jwt", "verify", "token", tok, "key", c.pub, "alg", c.alg)
		if !strings.Contains(toolstest.Compact(got), `"sub":"u"`) {
			t.Errorf("%s: verify result %v", c.alg, toolstest.Compact(got))
		}
		// flipping a payload character must break the signature
		parts := strings.Split(tok, ".")
		parts[1] = base64.RawURLEncoding.EncodeToString([]byte(`{"sub":"evil"}`))
		toolstest.Fails(t, "jwt", "verify", "token", strings.Join(parts, "."), "key", c.pub)
	}
	// wrong key, wrong curve, wrong family
	otherRSA, _ := rsa.GenerateKey(crand.Reader, 2048)
	otherDER, _ := x509.MarshalPKIXPublicKey(&otherRSA.PublicKey)
	tok := toolstest.Must(t, "jwt", "sign", "payload", p, "key", rsaPriv, "alg", "RS256").(string)
	toolstest.Fails(t, "jwt", "verify", "token", tok, "key", pemOf("PUBLIC KEY", otherDER))
	toolstest.Fails(t, "jwt", "verify", "token", tok, "key", ecPub)
	toolstest.Fails(t, "jwt", "sign", "payload", p, "key", ecPriv, "alg", "ES384")
	toolstest.Fails(t, "jwt", "sign", "payload", p, "key", rsaPriv, "alg", "ES256")
	// key confusion: an HS256 token signed with the public PEM as secret must be refused
	forged := toolstest.Must(t, "jwt", "sign", "payload", p, "key", rsaPub, "alg", "HS256").(string)
	toolstest.Fails(t, "jwt", "verify", "token", forged, "key", rsaPub)
	toolstest.Fails(t, "jwt", "verify", "token", forged, "key", rsaPub, "alg", "RS256")
	toolstest.Fails(t, "jwt", "verify", "token", tok, "key", "plain secret")
}

func TestNeverPanics(t *testing.T) {
	rng := rand.New(rand.NewSource(8))
	alphabet := []byte("abc.=-_ey0Jw{}\":, Bearer\n")
	for i := 0; i < 4000; i++ {
		b := make([]byte, rng.Intn(60))
		for j := range b {
			b[j] = alphabet[rng.Intn(len(alphabet))]
		}
		s := string(b)
		toolstest.Call("jwt", "decode", "token", s)
		toolstest.Call("jwt", "verify", "token", s, "key", s)
		toolstest.Call("jwt", "verify", "token", s, "key", "-----BEGIN PUBLIC KEY-----\n"+s+"\n-----END PUBLIC KEY-----")
		toolstest.Call("jwt", "sign", "payload", toolstest.JSONValue(t, `{}`), "key", s, "alg", "RS256")
	}
	for _, mid := range []string{`{"alg":"RS256"}`, `{"alg":"ES256"}`, `{"alg":"EdDSA"}`, `{"alg":"PS256"}`, `{"alg":5}`, `{"alg":"HS256","exp":"x"}`, `{"alg":"HS256","exp":1e400}`} {
		tok := base64.RawURLEncoding.EncodeToString([]byte(mid)) + "." + base64.RawURLEncoding.EncodeToString([]byte(`{}`)) + ".AAAA"
		toolstest.Call("jwt", "verify", "token", tok, "key", "k")
	}
}

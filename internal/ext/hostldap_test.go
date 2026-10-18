package ext

import (
	"fmt"
	"net"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	ber "github.com/go-asn1-ber/asn1-ber"
	"github.com/go-ldap/ldap/v3"
)

// fakeLDAP is a small in-process LDAP server: simple bind, equality and
// presence search, add, modify and delete over plain TCP on loopback.
type fakeLDAP struct {
	ln      net.Listener
	mu      sync.Mutex
	log     []string
	users   map[string]string // dn -> password
	entries []fakeEntry
	conns   int
	closed  chan struct{}
}

type fakeEntry struct {
	dn    string
	attrs map[string][]string
}

func newFakeLDAP(t *testing.T) *fakeLDAP {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	s := &fakeLDAP{ln: ln, closed: make(chan struct{}, 16), users: map[string]string{"cn=admin,dc=example,dc=com": "adminpw", "uid=alice,ou=people,dc=example,dc=com": "alicepw"}}
	s.entries = []fakeEntry{
		{"uid=alice,ou=people,dc=example,dc=com", map[string][]string{"uid": {"alice"}, "cn": {"Alice A"}, "mail": {"alice@example.com", "a@example.com"}, "jpegPhoto": {"\xff\xd8\x00"}}},
		{"uid=bob,ou=people,dc=example,dc=com", map[string][]string{"uid": {"bob"}, "cn": {"Bob B"}}},
	}
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			s.mu.Lock()
			s.conns++
			s.mu.Unlock()
			go s.serve(c)
		}
	}()
	t.Cleanup(func() { ln.Close() })
	return s
}

func (s *fakeLDAP) port() string { _, p, _ := net.SplitHostPort(s.ln.Addr().String()); return p }
func (s *fakeLDAP) url() string  { return "ldap://127.0.0.1:" + s.port() }
func (s *fakeLDAP) record(f string, a ...any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log = append(s.log, fmt.Sprintf(f, a...))
}
func (s *fakeLDAP) lines() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return strings.Join(s.log, "\n")
}

func ldapReply(c net.Conn, id int64, tag ber.Tag, code int64, msg string) {
	env := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAPMessage")
	env.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, id, "id"))
	op := ber.Encode(ber.ClassApplication, ber.TypeConstructed, tag, nil, "result")
	op.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagEnumerated, code, "code"))
	op.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, "", "matched"))
	op.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, msg, "diagnostic"))
	env.AppendChild(op)
	c.Write(env.Bytes())
}

func (s *fakeLDAP) serve(c net.Conn) {
	defer func() { c.Close(); s.closed <- struct{}{} }()
	for {
		p, err := ber.ReadPacket(c)
		if err != nil {
			return
		}
		id := p.Children[0].Value.(int64)
		op := p.Children[1]
		switch op.Tag {
		case 0: // bind
			dn := op.Children[1].Value.(string)
			pw := string(op.Children[2].Data.Bytes())
			s.record("bind %s", dn)
			if dn == "" || s.users[dn] == pw {
				ldapReply(c, id, 1, 0, "")
			} else {
				ldapReply(c, id, 1, 49, "invalid credentials")
			}
		case 2: // unbind
			return
		case 3: // search
			s.search(c, id, op)
		case 8: // add
			dn := op.Children[0].Value.(string)
			var names []string
			for _, a := range op.Children[1].Children {
				names = append(names, a.Children[0].Value.(string)+"="+valuesOf(a.Children[1]))
			}
			s.record("add %s %s", dn, strings.Join(names, ";"))
			if strings.Contains(dn, "exists") {
				ldapReply(c, id, 9, 68, "entry exists")
				continue
			}
			ldapReply(c, id, 9, 0, "")
		case 6: // modify
			dn := op.Children[0].Value.(string)
			var ch []string
			for _, m := range op.Children[1].Children {
				kind := map[int64]string{0: "add", 1: "delete", 2: "replace"}[m.Children[0].Value.(int64)]
				ch = append(ch, kind+" "+m.Children[1].Children[0].Value.(string)+"="+valuesOf(m.Children[1].Children[1]))
			}
			s.record("modify %s %s", dn, strings.Join(ch, ";"))
			ldapReply(c, id, 7, 0, "")
		case 10: // delete
			s.record("delete %s", string(op.Data.Bytes()))
			ldapReply(c, id, 11, 0, "")
		}
	}
}

func valuesOf(set *ber.Packet) string {
	var v []string
	for _, c := range set.Children {
		v = append(v, c.Value.(string))
	}
	return strings.Join(v, ",")
}

func (s *fakeLDAP) search(c net.Conn, id int64, op *ber.Packet) {
	base := op.Children[0].Value.(string)
	limit := op.Children[3].Value.(int64)
	filter, _ := ldap.DecompileFilter(op.Children[6])
	var want []string
	for _, a := range op.Children[7].Children {
		want = append(want, a.Value.(string))
	}
	s.record("search %s %s limit=%d attrs=%v", base, filter, limit, want)
	n := int64(0)
	code := int64(0)
	for _, en := range s.entries {
		if !matchFilter(en, filter) {
			continue
		}
		if limit > 0 && n >= limit {
			code = 4
			break
		}
		n++
		env := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "LDAPMessage")
		env.AppendChild(ber.NewInteger(ber.ClassUniversal, ber.TypePrimitive, ber.TagInteger, id, "id"))
		e := ber.Encode(ber.ClassApplication, ber.TypeConstructed, 4, nil, "entry")
		e.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, en.dn, "dn"))
		attrs := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "attrs")
		for _, name := range []string{"uid", "cn", "mail", "jpegPhoto"} {
			vals, ok := en.attrs[name]
			if !ok || (len(want) > 0 && !contains(want, name)) {
				continue
			}
			a := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSequence, nil, "attr")
			a.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, name, "type"))
			set := ber.Encode(ber.ClassUniversal, ber.TypeConstructed, ber.TagSet, nil, "vals")
			for _, v := range vals {
				set.AppendChild(ber.NewString(ber.ClassUniversal, ber.TypePrimitive, ber.TagOctetString, v, "val"))
			}
			a.AppendChild(set)
			attrs.AppendChild(a)
		}
		e.AppendChild(attrs)
		env.AppendChild(e)
		c.Write(env.Bytes())
	}
	ldapReply(c, id, 5, code, "")
}

func contains(l []string, s string) bool {
	for _, x := range l {
		if x == s {
			return true
		}
	}
	return false
}

// matchFilter understands (attr=value), (attr=*) and (objectClass=*).
func matchFilter(en fakeEntry, f string) bool {
	f = strings.Trim(f, "()")
	k, v, _ := strings.Cut(f, "=")
	if v == "*" {
		return true
	}
	for _, x := range en.attrs[k] {
		if x == v {
			return true
		}
	}
	return false
}

const ldapPerm = Permission("ldap:127.0.0.1")

func TestLDAPEscapeHelpersNeedNoPermission(t *testing.T) {
	got := mustRun(t, `out = [ldap.escape_filter("a*(b)\\"), ldap.escape_dn("Smith, John+x"), ldap.escape_filter("plain")]`, nil, Options{})
	want := `["a\\2a\\28b\\29\\5c", "Smith\\, John\\+x", "plain"]`
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestLDAPConnectChecks(t *testing.T) {
	s := newFakeLDAP(t)
	for _, c := range []struct {
		granted []Permission
		script  string
		want    string
	}{
		{nil, `ldap.connect("` + s.url() + `")`, `requires "ldap"`},
		{[]Permission{"ldap:other.example.com"}, `ldap.connect("` + s.url() + `")`, `requires "ldap"`},
		{[]Permission{"ldap:127.0.0.1:1"}, `ldap.connect("` + s.url() + `")`, `requires "ldap"`},
		{[]Permission{"network:127.0.0.1", "sql:127.0.0.1"}, `ldap.connect("` + s.url() + `")`, `requires "ldap"`},
		{[]Permission{ldapPerm}, `ldap.connect("http://127.0.0.1")`, "url must look like"},
		{[]Permission{ldapPerm}, `ldap.connect("ldap://127.0.0.1/dc=x")`, "url must look like"},
		{[]Permission{ldapPerm}, `ldap.connect("ldap://u:p@127.0.0.1")`, "url must look like"},
		{[]Permission{ldapPerm}, `ldap.connect("` + s.url() + `", bind_dn="cn=admin,dc=example,dc=com")`, "both bind_dn and password"},
		{[]Permission{ldapPerm}, `ldap.connect("` + s.url() + `", start_tls=True, timeout=0)`, "timeout"},
		{[]Permission{ldapPerm}, `ldap.connect("ldaps://127.0.0.1", start_tls=True)`, "start_tls is for ldap://"},
		{[]Permission{"ldap:ldap.example.com"}, `ldap.connect("ldap://ldap.example.com", bind_dn="cn=a", password="pw")`, "unencrypted"},
	} {
		_, err := runScript(t, "out = "+c.script, c.granted, Options{})
		if err == nil || !strings.Contains(err.Error(), c.want) {
			t.Errorf("%s with %v: want %q, got %v", c.script, c.granted, c.want, err)
		}
	}
}

func TestLDAPSearchAndBindCheck(t *testing.T) {
	s := newFakeLDAP(t)
	got := mustRun(t, `c = ldap.connect("`+s.url()+`", bind_dn="cn=admin,dc=example,dc=com", password="adminpw")
r = c.search("ou=people,dc=example,dc=com", "(uid=alice)", attributes=["cn", "mail", "jpegPhoto"])
all = c.search("dc=example,dc=com", scope="sub")
out = [len(r), r[0]["dn"], r[0]["cn"], r[0]["mail"], r[0]["jpegPhoto"], len(all), str(c), type(c)]`, []Permission{ldapPerm}, Options{})
	want := `[1, "uid=alice,ou=people,dc=example,dc=com", ["Alice A"], ["alice@example.com", "a@example.com"], [b"\xff\xd8\x00"], 2, "<ldap.Connection 127.0.0.1:389>", "ldap.Connection"]`
	want = strings.Replace(want, "389", s.port(), 1)
	if got != want {
		t.Fatalf("got  %s\nwant %s", got, want)
	}
	log := s.lines()
	for _, l := range []string{"bind cn=admin,dc=example,dc=com", "search ou=people,dc=example,dc=com (uid=alice) limit=1000 attrs=[cn mail jpegPhoto]", "search dc=example,dc=com (objectClass=*) limit=1000 attrs=[]"} {
		if !strings.Contains(log, l) {
			t.Errorf("server log missing %q:\n%s", l, log)
		}
	}

	got = mustRun(t, `c = ldap.connect("`+s.url()+`")
out = [len(c.search("dc=example,dc=com", "(uid=*)", size_limit=1)),
       c.check_bind("uid=alice,ou=people,dc=example,dc=com", "alicepw"),
       c.check_bind("uid=alice,ou=people,dc=example,dc=com", "wrong")]`, []Permission{ldapPerm}, Options{})
	if got != "[1, True, False]" {
		t.Fatalf("got %s", got)
	}
	for _, bad := range []string{
		`c.search("x", scope="deep")`,
		`c.search("x", size_limit=0)`,
		`c.search("x", "(((")`,
		`c.search("x", attributes=[1])`,
		`c.check_bind("uid=alice", "")`,
	} {
		if _, err := runScript(t, "c = ldap.connect(\""+s.url()+"\")\nout = "+bad, []Permission{ldapPerm}, Options{}); err == nil {
			t.Errorf("%s: want error", bad)
		}
	}
	if _, err := runScript(t, `c = ldap.connect("`+s.url()+`", bind_dn="cn=admin,dc=example,dc=com", password="nope")
out = 1`, []Permission{ldapPerm}, Options{}); err == nil || !strings.Contains(err.Error(), "bind as") || strings.Contains(err.Error(), "nope") {
		t.Errorf("bad bind: %v", err)
	}
}

func TestLDAPWritesNeedLDAPWrite(t *testing.T) {
	s := newFakeLDAP(t)
	body := `c = ldap.connect("` + s.url() + `")
c.add("uid=new,ou=people,dc=example,dc=com", {"uid": "new", "objectClass": ["top", "person"]})
c.modify("uid=new,ou=people,dc=example,dc=com", replace={"cn": "New"}, add={"mail": ["n@example.com"]}, delete=["description"])
c.delete("uid=new,ou=people,dc=example,dc=com")
out = 1`
	for _, op := range []string{
		`c.add("uid=x", {"uid": "x"})`,
		`c.modify("uid=x", replace={"cn": "X"})`,
		`c.delete("uid=x")`,
	} {
		_, err := runScript(t, "c = ldap.connect(\""+s.url()+"\")\nout = "+op, []Permission{ldapPerm}, Options{})
		if err == nil || !strings.Contains(err.Error(), `requires "ldap.write"`) {
			t.Errorf("%s with read-only: %v", op, err)
		}
	}
	mustRun(t, body, []Permission{ldapPerm, "ldap.write:127.0.0.1"}, Options{})
	log := s.lines()
	for _, l := range []string{
		"add uid=new,ou=people,dc=example,dc=com objectClass=top,person;uid=new",
		"modify uid=new,ou=people,dc=example,dc=com replace cn=New;add mail=n@example.com;delete description=",
		"delete uid=new,ou=people,dc=example,dc=com",
	} {
		if !strings.Contains(log, l) {
			t.Errorf("server log missing %q:\n%s", l, log)
		}
	}
	// ldap.write alone may connect and write but not read
	_, err := runScript(t, `c = ldap.connect("`+s.url()+`")
out = c.search("dc=example,dc=com")`, []Permission{"ldap.write:127.0.0.1"}, Options{})
	if err == nil || !strings.Contains(err.Error(), `requires "ldap"`) {
		t.Errorf("write-only search: %v", err)
	}
	mustRun(t, `c = ldap.connect("`+s.url()+`")
c.delete("uid=w,dc=example,dc=com")
out = 1`, []Permission{"ldap.write:127.0.0.1"}, Options{})
	// server errors come through, and bad arguments are refused before sending
	_, err = runScript(t, `c = ldap.connect("`+s.url()+`")
out = c.add("uid=exists,dc=x", {"uid": "e"})`, []Permission{"ldap.write:127.0.0.1"}, Options{})
	if err == nil || !strings.Contains(err.Error(), "Entry Already Exists") {
		t.Errorf("server error: %v", err)
	}
	for _, bad := range []string{`c.modify("uid=x")`, `c.modify("uid=x", replace="cn")`, `c.add("uid=x", {"a b": "x"})`, `c.add("uid=x", {"n": 5})`, `c.delete("")`} {
		if _, err := runScript(t, "c = ldap.connect(\""+s.url()+"\")\nout = "+bad, []Permission{"ldap.write"}, Options{}); err == nil {
			t.Errorf("%s: want error", bad)
		}
	}
}

func TestLDAPSecretPasswordAndStop(t *testing.T) {
	s := newFakeLDAP(t)
	sec := useFakeSecrets(t)
	sec.m[secretTarget("test.ext", "ldappw")] = "adminpw"
	body := `def on_query(ctx, q):
    c = ldap.connect("` + s.url() + `", bind_dn="cn=admin,dc=example,dc=com", password={"secret": "ldappw"})
    return [item("1", str(len(c.search("dc=example,dc=com"))))]
`
	e := openStarted(t, baseManifest, body, []Permission{ldapPerm, "secrets"}, Options{})
	items, err := e.Query(t.Context(), "")
	if err != nil || len(items) != 1 || items[0].Title != "2" {
		t.Fatalf("query: %v %v", items, err)
	}
	if !strings.Contains(s.lines(), "bind cn=admin,dc=example,dc=com") {
		t.Errorf("secret password was not used:\n%s", s.lines())
	}
	e.Stop()
	<-s.closed // the server saw the connection close
	e2 := openStarted(t, baseManifest, body, []Permission{ldapPerm}, Options{})
	if _, err := e2.Query(t.Context(), ""); err == nil || !strings.Contains(err.Error(), `requires "secrets"`) {
		t.Errorf("secret without permission: %v", err)
	}
}

func TestLDAPSVerifiesCertificate(t *testing.T) {
	srv := httptest.NewUnstartedServer(nil)
	srv.EnableHTTP2 = false
	srv.StartTLS() // a certificate the system does not trust
	defer srv.Close()
	_, port, _ := net.SplitHostPort(srv.Listener.Addr().String())
	_, err := runScript(t, `out = ldap.connect("ldaps://127.0.0.1:`+port+`")`, []Permission{ldapPerm}, Options{})
	if err == nil || !strings.Contains(err.Error(), "certificate") {
		t.Fatalf("want a certificate error, got %v", err)
	}
	// a client that trusts it would get further: check the config is not skipping verification
	c := &ldapConn{host: "127.0.0.1", port: port, ldaps: true, timeout: 2e9}
	if _, err := dialLDAP(t.Context(), c); err == nil {
		t.Fatal("dialLDAP accepted an untrusted certificate")
	}
}

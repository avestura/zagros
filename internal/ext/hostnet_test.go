package ext

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

// runNetScript runs body (which must set r) inside on_query and returns the title.
func runNetScript(t *testing.T, perms []Permission, body string, o Options) string {
	t.Helper()
	var sb strings.Builder
	sb.WriteString("def on_query(ctx, q):\n")
	for _, l := range strings.Split(strings.TrimSpace(body), "\n") {
		sb.WriteString("    " + l + "\n")
	}
	sb.WriteString("    return [item(\"1\", str(r))]\n")
	e := openStarted(t, baseManifest, sb.String(), perms, o)
	items, err := e.Query(context.Background(), "")
	if err != nil {
		return "ERR " + err.Error()
	}
	if len(items) != 1 {
		t.Fatalf("items: %v", items)
	}
	return items[0].Title
}

func listenTCP(t *testing.T, handle func(net.Conn)) (host string, port int) {
	t.Helper()
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go func() { defer c.Close(); handle(c) }()
		}
	}()
	a := l.Addr().(*net.TCPAddr)
	return "127.0.0.1", a.Port
}

func TestNetworkScopeMatching(t *testing.T) {
	ps := func(s string) PermissionSet { return PermissionSet{Permission(s)} }
	cases := []struct {
		scope, target string
		want          bool
	}{
		{"network:example.com", "example.com", true},
		{"network:example.com", "api.example.com", true},
		{"network:example.com", "example.com:25", true},
		{"network:example.com", "badexample.com", false},
		{"network:example.com:25", "example.com:25", true},
		{"network:example.com:25", "mail.example.com:25", true},
		{"network:example.com:25", "example.com:26", false},
		{"network:example.com:25", "example.com", false},
		{"network:127.0.0.1:80", "127.0.0.1:80", true},
		{"network:[::1]:80", "[::1]:80", true},
		{"network:[::1]:80", "[::1]:81", false},
		{"network:::1", "[::1]:80", true},
		{"network:Example.COM", "example.com.", true},
	}
	for _, c := range cases {
		if got := ps(c.scope).Allows(PermNetwork, c.target); got != c.want {
			t.Errorf("%s vs %s: got %v want %v", c.scope, c.target, got, c.want)
		}
	}
	if !ps("network").Allows(PermNetwork, "x.test:1") {
		t.Error("unscoped network allows any host")
	}
}

func TestSSHScopeMatching(t *testing.T) {
	ps := func(s string) PermissionSet { return PermissionSet{Permission(s)} }
	cases := []struct {
		scope, target string
		want          bool
	}{
		{"ssh:deploy@host.example:22", "deploy@host.example:22", true},
		{"ssh:deploy@host.example", "deploy@host.example:2222", true},
		{"ssh:deploy@host.example:22", "deploy@host.example:2222", false},
		{"ssh:deploy@host.example", "root@host.example:22", false},
		{"ssh:host.example", "root@host.example:22", true},
		{"ssh:host.example", "root@sub.host.example:22", false},
		{"ssh:DEPLOY@host.example", "deploy@host.example:22", false},
		{"ssh:deploy@HOST.example", "deploy@host.example:22", true},
		{"ssh:deploy@[::1]:22", "deploy@[::1]:22", true},
	}
	for _, c := range cases {
		if got := ps(c.scope).Allows(PermSSH, c.target); got != c.want {
			t.Errorf("%s vs %s: got %v want %v", c.scope, c.target, got, c.want)
		}
	}
	for _, p := range []Permission{"ssh", "ssh:me@host:22", "network:host:443"} {
		if err := p.Validate(); err != nil {
			t.Errorf("%s: %v", p, err)
		}
	}
	if Permission("ssh:").Validate() == nil {
		t.Error("empty ssh scope should be invalid")
	}
}

func TestTCPRoundTripAndDeny(t *testing.T) {
	host, port := listenTCP(t, func(c net.Conn) {
		buf := make([]byte, 64)
		n, _ := c.Read(buf)
		c.Write([]byte("HELLO " + strings.ToUpper(string(buf[:n]))))
		c.Write([]byte("\r\nsecond line\r\ntail"))
	})
	p := strconv.Itoa(port)
	perm := []Permission{Permission("network:" + host + ":" + p)}

	got := runNetScript(t, perm, `
c = tcp.connect("`+host+`", `+p+`)
c.send("abc")
a = c.recv_until("\r\n")
b = c.recv_until(b"\r\n")
t = c.recv(100)
c.close()
r = [a, b, t, str(c)]`, Options{})
	if !strings.Contains(got, `b"HELLO ABC\r\n"`) || !strings.Contains(got, `b"second line\r\n"`) || !strings.Contains(got, `b"tail"`) {
		t.Fatalf("got %s", got)
	}

	got = runNetScript(t, perm, `r = repr(tcp.request("`+host+`", `+p+`, "xy", until="\r\n"))`, Options{})
	if got != `b"HELLO XY\r\n"` {
		t.Fatalf("request until: %s", got)
	}
	got = runNetScript(t, perm, `r = repr(tcp.request("`+host+`", `+p+`, "xy"))`, Options{})
	if got != `b"HELLO XY\r\nsecond line\r\ntail"` {
		t.Fatalf("request eof: %s", got)
	}

	for name, perms := range map[string][]Permission{
		"none":       nil,
		"wrong port": {Permission("network:" + host + ":" + strconv.Itoa(port+1))},
		"wrong host": {"network:example.com"},
	} {
		got = runNetScript(t, perms, `r = tcp.connect("`+host+`", `+p+`)`, Options{})
		if !strings.Contains(got, "permission denied") {
			t.Errorf("%s: want permission denied, got %s", name, got)
		}
	}
	// the host-only scope covers every port
	if got = runNetScript(t, []Permission{Permission("network:" + host)}, `r = repr(tcp.request("`+host+`", `+p+`, "q", until="\r\n"))`, Options{}); !strings.Contains(got, "HELLO Q") {
		t.Errorf("host scope: %s", got)
	}
}

func TestTCPErrorsAndCaps(t *testing.T) {
	host, port := listenTCP(t, func(c net.Conn) {
		io.WriteString(c, strings.Repeat("x", 5000))
		io.Copy(io.Discard, c)
	})
	p := strconv.Itoa(port)
	perm := []Permission{Permission("network:" + host)}
	small := Options{Limits: Limits{MaxHTTPBytes: 1000}}
	big := Options{Limits: Limits{MaxHTTPBytes: 100000}}

	got := runNetScript(t, perm, `
c = tcp.connect("`+host+`", `+p+`)
r = attempt(c.recv_until, "\n", 500)[1]`, big)
	if !strings.Contains(got, "delimiter not found") {
		t.Errorf("recv_until max: %s", got)
	}
	got = runNetScript(t, perm, `
c = tcp.connect("`+host+`", `+p+`)
r = attempt(c.recv_until, "\n", 1000)[1]`, small)
	if !strings.Contains(got, "bytes received") {
		t.Errorf("recv_until: %s", got)
	}
	got = runNetScript(t, perm, `r = attempt(tcp.request, "`+host+`", `+p+`, "", None, 4000)[1]`, small)
	if !strings.Contains(got, "max must be") {
		t.Errorf("request max: %s", got)
	}
	got = runNetScript(t, perm, `
c = tcp.connect("`+host+`", `+p+`)
c.recv(10)
c.close()
r = attempt(c.recv, 10)[1]`, small)
	if !strings.Contains(got, "closed") {
		t.Errorf("use after close: %s", got)
	}
	got = runNetScript(t, perm, `
c = tcp.connect("`+host+`", `+p+`)
r = attempt(c.send, "x" * 2000)[1]`, small)
	if !strings.Contains(got, "bytes sent") {
		t.Errorf("send cap: %s", got)
	}
	got = runNetScript(t, perm, `
cs = [tcp.connect("`+host+`", `+p+`) for i in range(9)]
r = "no error"`, small)
	if !strings.Contains(got, "connections open") {
		t.Errorf("connection cap: %s", got)
	}
	got = runNetScript(t, perm, `r = attempt(tcp.connect, "`+host+`", 70000)[1]`, small)
	if !strings.Contains(got, "invalid port") {
		t.Errorf("port: %s", got)
	}
	got = runNetScript(t, perm, `r = attempt(tcp.connect, "`+host+`", `+p+`, 0)[1]`, small)
	if !strings.Contains(got, "timeout must be") {
		t.Errorf("timeout: %s", got)
	}
	got = runNetScript(t, perm, `r = attempt(tcp.connect, "`+host+`", 1)[1]`, small)
	if !strings.Contains(got, "tcp.connect") {
		t.Errorf("refused: %s", got)
	}
}

func TestConnectionsCloseWithTheCallAndStop(t *testing.T) {
	host, port := listenTCP(t, func(c net.Conn) { io.Copy(io.Discard, c) })
	p := strconv.Itoa(port)
	e := openStarted(t, baseManifest, `def on_query(ctx, q):
    c = tcp.connect("`+host+`", `+p+`)
    return [item("1", c.addr)]
`, []Permission{Permission("network:" + host)}, Options{})
	if _, err := e.Query(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	open := -1
	for i := 0; i < 100 && open != 0; i++ {
		e.host.net.mu.Lock()
		open = len(e.host.net.conns)
		e.host.net.mu.Unlock()
		time.Sleep(10 * time.Millisecond)
	}
	if open != 0 {
		t.Fatalf("%d connections left after the call", open)
	}
	// a connection still tracked is closed by Stop
	c, err := net.Dial("tcp", net.JoinHostPort(host, p))
	if err != nil {
		t.Fatal(err)
	}
	if err := e.trackConn(c); err != nil {
		t.Fatal(err)
	}
	e.Stop()
	if _, err := c.Write([]byte("x")); err == nil {
		t.Error("Stop should have closed the connection")
	}
}

func TestResolveThenCheck(t *testing.T) {
	old := lookupIPAddr
	defer func() { lookupIPAddr = old }()
	host, port := listenTCP(t, func(c net.Conn) { io.WriteString(c, "hi\n") })
	p := strconv.Itoa(port)
	lookups := 0
	lookupIPAddr = func(_ context.Context, h string) ([]net.IPAddr, error) {
		lookups++
		switch h {
		case "rebind.example.test":
			return []net.IPAddr{{IP: net.ParseIP(host)}}, nil
		case "localhost":
			return []net.IPAddr{{IP: net.ParseIP(host)}}, nil
		}
		return nil, fmt.Errorf("no such host")
	}
	script := `r = attempt(tcp.request, "rebind.example.test", ` + p + `, "", "\n")`
	got := runNetScript(t, []Permission{"network:rebind.example.test"}, script, Options{})
	if !strings.Contains(got, "local or private address") {
		t.Fatalf("scoped name that resolves to loopback must be refused: %s", got)
	}
	// unscoped network may reach it, as it may name the IP directly
	got = runNetScript(t, []Permission{"network"}, script, Options{})
	if !strings.Contains(got, `b"hi\n"`) {
		t.Fatalf("unscoped: %s", got)
	}
	got = runNetScript(t, []Permission{"network:localhost"}, `r = repr(tcp.request("localhost", `+p+`, "", "\n"))`, Options{})
	if got != `b"hi\n"` {
		t.Fatalf("localhost: %s", got)
	}
	// granting the address itself is the way to reach a local service
	got = runNetScript(t, []Permission{Permission("network:" + host)}, `r = repr(tcp.request("`+host+`", `+p+`, "", "\n"))`, Options{})
	if got != `b"hi\n"` {
		t.Fatalf("ip grant: %s", got)
	}
	// the answer is dialled, not looked up again
	lookups = 0
	runNetScript(t, []Permission{"network"}, script, Options{})
	if lookups != 1 {
		t.Errorf("host resolved %d times, want 1", lookups)
	}
	for _, a := range []string{"10.1.2.3", "192.168.0.1", "169.254.1.1", "100.64.0.1", "::1", "fe80::1", "0.0.0.0", "fc00::1"} {
		ad, _ := parseHostIP(a)
		if !nonPublic(ad) {
			t.Errorf("%s should be non-public", a)
		}
	}
	for _, a := range []string{"8.8.8.8", "1.1.1.1", "2606:4700::1111"} {
		ad, _ := parseHostIP(a)
		if nonPublic(ad) {
			t.Errorf("%s should be public", a)
		}
	}
}

func TestUDP(t *testing.T) {
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pc.Close() })
	go func() {
		buf := make([]byte, 2048)
		for {
			n, a, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			pc.WriteTo([]byte("echo:"+string(buf[:n])), a)
		}
	}()
	p := strconv.Itoa(pc.LocalAddr().(*net.UDPAddr).Port)
	perm := []Permission{Permission("network:127.0.0.1:" + p)}
	got := runNetScript(t, perm, `
s = udp.open("127.0.0.1", `+p+`)
s.send("one")
a = s.recv()
b = udp.request("127.0.0.1", `+p+`, b"two")
n = udp.send("127.0.0.1", `+p+`, "three")
r = [a, b, n]`, Options{})
	if got != `[b"echo:one", b"echo:two", 5]` {
		t.Fatalf("got %s", got)
	}
	got = runNetScript(t, perm, `
s = udp.open("127.0.0.1", `+p+`)
r = attempt(s.recv, 10, 0.2)[1]`, Options{})
	if !strings.Contains(got, "timed out") {
		t.Errorf("udp timeout: %s", got)
	}
	got = runNetScript(t, nil, `r = udp.send("127.0.0.1", `+p+`, "x")`, Options{})
	if !strings.Contains(got, "permission denied") {
		t.Errorf("udp deny: %s", got)
	}
}

func TestTLSCertAndVerifiedConnect(t *testing.T) {
	srv := httptest.NewTLSServer(nil)
	defer srv.Close()
	host, ps, _ := net.SplitHostPort(srv.Listener.Addr().String())
	perm := []Permission{Permission("network:" + host)}
	got := runNetScript(t, perm, `
c = tls.cert("`+host+`", `+ps+`)
r = [c["verified"], c["expired"], c["days_left"] > 0, len(c["sans"]) > 0, len(c["fingerprint_sha256"]), c["tls_version"], c["issuer"] != ""]`, Options{})
	if got != `[False, False, True, True, 64, "TLS 1.3", True]` {
		t.Fatalf("cert: %s", got)
	}
	got = runNetScript(t, perm, `
c = tls.cert("`+host+`", `+ps+`)
r = c["verify_error"]`, Options{})
	if !strings.Contains(got, "certificate") {
		t.Errorf("verify_error: %s", got)
	}
	got = runNetScript(t, perm, `r = attempt(tls.connect, "`+host+`", `+ps+`)[1]`, Options{})
	if !strings.Contains(got, "x509") && !strings.Contains(got, "certificate") {
		t.Errorf("tls.connect must verify the certificate: %s", got)
	}
	got = runNetScript(t, nil, `r = tls.cert("`+host+`", `+ps+`)`, Options{})
	if !strings.Contains(got, "permission denied") {
		t.Errorf("tls deny: %s", got)
	}
}

func TestCertInfoFields(t *testing.T) {
	srv := httptest.NewTLSServer(nil)
	defer srv.Close()
	leaf := srv.Certificate()
	st := tls.ConnectionState{PeerCertificates: []*x509.Certificate{leaf}, Version: tls.VersionTLS13}
	v := certInfo(st, "example.com", leaf.NotAfter.Add(48*3600e9))
	s := v.String()
	if !strings.Contains(s, `"expired": True`) || !strings.Contains(s, `"days_left": -2`) {
		t.Errorf("expired cert: %s", s)
	}
	if !strings.Contains(s, "example.com") || !strings.Contains(s, "127.0.0.1") {
		t.Errorf("sans missing: %s", s)
	}
}

// fakeDNS answers A, MX and TXT for example.test over UDP.
func fakeDNS(t *testing.T) string {
	t.Helper()
	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { pc.Close() })
	go func() {
		buf := make([]byte, 1500)
		for {
			n, a, err := pc.ReadFrom(buf)
			if err != nil {
				return
			}
			var m dnsmessage.Message
			if m.Unpack(buf[:n]) != nil || len(m.Questions) == 0 {
				continue
			}
			q := m.Questions[0]
			m.Header.Response, m.Header.Authoritative, m.Header.RecursionAvailable = true, true, true
			hdr := dnsmessage.ResourceHeader{Name: q.Name, Class: dnsmessage.ClassINET, TTL: 60, Type: q.Type}
			switch q.Type {
			case dnsmessage.TypeA:
				m.Answers = []dnsmessage.Resource{{Header: hdr, Body: &dnsmessage.AResource{A: [4]byte{192, 0, 2, 7}}}}
			case dnsmessage.TypeMX:
				m.Answers = []dnsmessage.Resource{{Header: hdr, Body: &dnsmessage.MXResource{Pref: 10, MX: dnsmessage.MustNewName("mail.example.test.")}}}
			case dnsmessage.TypeTXT:
				m.Answers = []dnsmessage.Resource{{Header: hdr, Body: &dnsmessage.TXTResource{TXT: []string{"v=spf1 -all"}}}}
			case dnsmessage.TypeNS:
				m.Answers = []dnsmessage.Resource{{Header: hdr, Body: &dnsmessage.NSResource{NS: dnsmessage.MustNewName("ns1.example.test.")}}}
			default:
				m.Header.RCode = dnsmessage.RCodeNameError
			}
			out, err := m.Pack()
			if err == nil {
				pc.WriteTo(out, a)
			}
		}
	}()
	return pc.LocalAddr().String()
}

func TestDNSLookup(t *testing.T) {
	srv := fakeDNS(t)
	perm := []Permission{"network:127.0.0.1", "network:example.test"}
	q := func(typ string) string {
		return runNetScript(t, perm, `r = dns.lookup("example.test", "`+typ+`", server="`+srv+`", timeout=3)`, Options{})
	}
	if got := q("A"); got != `["192.0.2.7"]` {
		t.Errorf("A: %s", got)
	}
	if got := q("mx"); got != `[{"host": "mail.example.test", "pref": 10}]` {
		t.Errorf("MX: %s", got)
	}
	if got := q("TXT"); got != `["v=spf1 -all"]` {
		t.Errorf("TXT: %s", got)
	}
	if got := q("NS"); got != `["ns1.example.test"]` {
		t.Errorf("NS: %s", got)
	}
	if got := q("AAAA"); !strings.HasPrefix(got, "ERR") && !strings.Contains(got, "no such record") && got != "[]" {
		t.Errorf("AAAA: %s", got)
	}
	if got := q("SRV"); !strings.Contains(got, "unsupported type") {
		t.Errorf("SRV: %s", got)
	}
	// the name and the server are both checked
	for name, perms := range map[string][]Permission{
		"no server grant": {"network:example.test"},
		"no name grant":   {"network:127.0.0.1"},
		"nothing":         nil,
	} {
		got := runNetScript(t, perms, `r = dns.lookup("example.test", "A", server="`+srv+`")`, Options{})
		if !strings.Contains(got, "permission denied") {
			t.Errorf("%s: %s", name, got)
		}
	}
	if got := runNetScript(t, perm, `r = attempt(dns.lookup, "example.test", "PTR")[1]`, Options{}); !strings.Contains(got, "PTR needs an IP") {
		t.Errorf("PTR: %s", got)
	}
}

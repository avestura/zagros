package ext

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

const (
	netMaxConns   = 8               // open sockets and ssh connections per extension
	netMaxLife    = 2 * time.Minute // hard lifetime of one connection
	netMaxTimeout = 300.0           // seconds, for any single timeout argument
	netMaxRecs    = 256             // dns records returned
	udpMaxPacket  = 65507
	netRecvBuf    = 32 * 1024
)

// netState tracks the open connections and ssh host keys of one extension.
type netState struct {
	mu    sync.Mutex
	conns map[io.Closer]struct{}
	known map[string]map[string]string // ssh host keys kept in memory when there is no data folder
}

func (e *Extension) trackConn(c io.Closer) error {
	e.host.net.mu.Lock()
	defer e.host.net.mu.Unlock()
	if len(e.host.net.conns) >= netMaxConns {
		return fmt.Errorf("more than %d connections open; close some first", netMaxConns)
	}
	if e.host.net.conns == nil {
		e.host.net.conns = map[io.Closer]struct{}{}
	}
	e.host.net.conns[c] = struct{}{}
	return nil
}

func (e *Extension) untrackConn(c io.Closer) {
	e.host.net.mu.Lock()
	delete(e.host.net.conns, c)
	e.host.net.mu.Unlock()
}

// closeNet closes every connection the extension still holds (on Stop).
func (e *Extension) closeNet() {
	e.host.net.mu.Lock()
	conns := e.host.net.conns
	e.host.net.conns = nil
	e.host.net.mu.Unlock()
	for c := range conns {
		c.Close()
	}
}

// ---- scopes ----

// splitHostPortLoose splits "host", "host:port", "[v6]" and "[v6]:port"; a
// bare IPv6 address (several colons) is a host.
func splitHostPortLoose(s string) (host, port string) {
	if strings.HasPrefix(s, "[") {
		if h, p, err := net.SplitHostPort(s); err == nil {
			return h, p
		}
		return strings.Trim(s, "[]"), ""
	}
	if strings.Count(s, ":") == 1 {
		if h, p, err := net.SplitHostPort(s); err == nil {
			return h, p
		}
	}
	return s, ""
}

// networkScopeMatches: "host" covers the host and its subdomains on any
// port; "host:port" covers that port only. Targets without a port (http, dns)
// never match a scope that names a port.
func networkScopeMatches(scope, target string) bool {
	sh, sp := splitHostPortLoose(strings.ToLower(scope))
	th, tp := splitHostPortLoose(strings.ToLower(target))
	th, sh = strings.TrimSuffix(th, "."), strings.TrimSuffix(sh, ".")
	if sp != "" && sp != tp {
		return false
	}
	return th == sh || strings.HasSuffix(th, "."+sh)
}

// sshScopeMatches: scope is [user@]host[:port], target is user@host:port.
// The host must match exactly; an omitted user or port means any.
func sshScopeMatches(scope, target string) bool {
	su, sr := "", scope
	if i := strings.LastIndex(scope, "@"); i >= 0 {
		su, sr = scope[:i], scope[i+1:]
	}
	tu, tr := "", target
	if i := strings.LastIndex(target, "@"); i >= 0 {
		tu, tr = target[:i], target[i+1:]
	}
	sh, sp := splitHostPortLoose(sr)
	th, tp := splitHostPortLoose(tr)
	if su != "" && su != tu {
		return false
	}
	if sp != "" && sp != tp {
		return false
	}
	return strings.EqualFold(strings.TrimSuffix(sh, "."), strings.TrimSuffix(th, "."))
}

func (e *Extension) hasUnscoped(kind string) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, p := range e.granted {
		if p.Kind() == kind && p.Scope() == "" {
			return true
		}
	}
	return false
}

// ---- resolve and dial ----

// lookupIPAddr is replaced in tests.
var lookupIPAddr = func(ctx context.Context, host string) ([]net.IPAddr, error) {
	return net.DefaultResolver.LookupIPAddr(ctx, host)
}

var cgnat = netip.MustParsePrefix("100.64.0.0/10")

func nonPublic(a netip.Addr) bool {
	a = a.Unmap()
	return a.IsLoopback() || a.IsPrivate() || a.IsLinkLocalUnicast() || a.IsLinkLocalMulticast() ||
		a.IsMulticast() || a.IsUnspecified() || cgnat.Contains(a) || (a.Is4() && a.As4()[0] == 0)
}

func isLocalName(h string) bool {
	h = strings.ToLower(strings.TrimSuffix(h, "."))
	return h == "localhost" || strings.HasSuffix(h, ".localhost")
}

func parseHostIP(h string) (netip.Addr, bool) {
	a, err := netip.ParseAddr(strings.Trim(h, "[]"))
	return a.Unmap(), err == nil
}

func checkHostName(host string) error {
	if host == "" || len(host) > 253 || strings.ContainsAny(host, " \t\r\n/\\@") {
		return fmt.Errorf("invalid host %q", host)
	}
	return nil
}

func checkPort(port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("invalid port %d", port)
	}
	return nil
}

// resolveHost turns host into addresses. With enforce, a name that resolves to
// a loopback, private or link-local address is refused unless it is
// "localhost": otherwise a host granted by name could be pointed at services
// on this machine or the local network through its DNS record. Reach those by
// granting the IP address itself.
func resolveHost(ctx context.Context, host string, enforce bool) ([]netip.Addr, error) {
	if a, ok := parseHostIP(host); ok {
		return []netip.Addr{a}, nil
	}
	addrs, err := lookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	var out []netip.Addr
	for _, ia := range addrs {
		a, ok := netip.AddrFromSlice(ia.IP)
		if !ok {
			continue
		}
		a = a.Unmap()
		if enforce && !isLocalName(host) && nonPublic(a) {
			return nil, fmt.Errorf("%s resolves to %s, a local or private address; grant the address itself to reach it", host, a)
		}
		if ia.Zone != "" {
			a = a.WithZone(ia.Zone)
		}
		out = append(out, a)
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("no address for %s", host)
	}
	return out, nil
}

// dialHost resolves host once, checks the result and dials the address, so a
// second DNS answer cannot redirect the connection.
func dialHost(ctx context.Context, network, host string, port int, timeout time.Duration, enforce bool) (net.Conn, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	ips, err := resolveHost(ctx, host, enforce)
	if err != nil {
		return nil, err
	}
	var first error
	for _, ip := range ips {
		var d net.Dialer
		c, err := d.DialContext(ctx, network, net.JoinHostPort(ip.String(), strconv.Itoa(port)))
		if err == nil {
			return c, nil
		}
		if first == nil {
			first = err
		}
		if ctx.Err() != nil {
			break
		}
	}
	return nil, first
}

// secs is a timeout argument in seconds that accepts an int or a float.
type secs float64

func (s *secs) Unpack(v starlark.Value) error {
	switch x := v.(type) {
	case starlark.Int:
		f, _ := starlark.AsFloat(x)
		*s = secs(f)
	case starlark.Float:
		*s = secs(x)
	default:
		return fmt.Errorf("got %s, want int or float", v.Type())
	}
	return nil
}

func toTimeout(name string, t secs) (time.Duration, error) {
	if t <= 0 || t > netMaxTimeout {
		return 0, fmt.Errorf("%s: timeout must be between 0 and %d seconds", name, int(netMaxTimeout))
	}
	return time.Duration(float64(t) * float64(time.Second)), nil
}

func dataBytes(v starlark.Value) ([]byte, error) {
	switch x := v.(type) {
	case starlark.String:
		return []byte(x), nil
	case starlark.Bytes:
		return []byte(x), nil
	}
	return nil, fmt.Errorf("data must be a string or bytes, got %s", v.Type())
}

// ---- connection value ----

// netConn is a tcp, tls or udp connection held by a script. It closes when the
// call that opened it ends, when its lifetime runs out or when the extension stops.
type netConn struct {
	e      *Extension
	kind   string // "tcp", "tls", "udp"
	addr   string
	c      net.Conn
	life   time.Time
	cap    int64
	stop   func() bool
	mu     sync.Mutex
	rbuf   []byte
	sent   int64
	got    int64
	closed bool
}

func (e *Extension) newNetConn(th *starlark.Thread, kind, addr string, c net.Conn) (*netConn, error) {
	n := &netConn{e: e, kind: kind, addr: addr, c: c, life: time.Now().Add(netMaxLife), cap: e.opts.Limits.MaxHTTPBytes}
	if err := e.trackConn(n); err != nil {
		c.Close()
		return nil, err
	}
	_ = c.SetDeadline(n.life)
	n.stop = context.AfterFunc(callCtx(th), func() { n.shut() })
	return n, nil
}

func (n *netConn) shut() {
	n.mu.Lock()
	if n.closed {
		n.mu.Unlock()
		return
	}
	n.closed = true
	n.mu.Unlock()
	if n.stop != nil {
		n.stop()
	}
	n.c.Close()
	n.e.untrackConn(n)
}

func (n *netConn) Close() error { n.shut(); return nil }

var _ starlark.HasAttrs = (*netConn)(nil)

func (n *netConn) String() string { return "<" + n.Type() + " " + n.addr + ">" }
func (n *netConn) Type() string {
	if n.kind == "udp" {
		return "udp.Socket"
	}
	return n.kind + ".Conn"
}
func (n *netConn) Freeze()              {}
func (n *netConn) Truth() starlark.Bool { return true }
func (n *netConn) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable type: %s", n.Type())
}

func (n *netConn) AttrNames() []string {
	if n.kind == "udp" {
		return []string{"addr", "close", "recv", "send"}
	}
	return []string{"addr", "close", "recv", "recv_until", "send"}
}

func (n *netConn) Attr(name string) (starlark.Value, error) {
	p := n.kind + "."
	switch name {
	case "addr":
		return starlark.String(n.addr), nil
	case "send":
		return starlark.NewBuiltin(p+"send", n.send), nil
	case "recv":
		return starlark.NewBuiltin(p+"recv", n.recv), nil
	case "close":
		return starlark.NewBuiltin(p+"close", n.closeFn), nil
	case "recv_until":
		if n.kind != "udp" {
			return starlark.NewBuiltin(p+"recv_until", n.recvUntil), nil
		}
	}
	return nil, nil
}

func (n *netConn) deadline(d time.Duration) time.Time {
	t := time.Now().Add(d)
	if t.After(n.life) {
		return n.life
	}
	return t
}

func (n *netConn) check(op string) error {
	if n.closed {
		return fmt.Errorf("%s.%s: connection is closed", n.kind, op)
	}
	return nil
}

func (n *netConn) wrapErr(op string, err error) error {
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return fmt.Errorf("%s.%s: timed out", n.kind, op)
	}
	if n.closed {
		return fmt.Errorf("%s.%s: connection is closed", n.kind, op)
	}
	return fmt.Errorf("%s.%s: %w", n.kind, op, err)
}

func (n *netConn) send(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var data starlark.Value
	timeout := secs(10.0)
	name := n.kind + ".send"
	if err := starlark.UnpackArgs(name, args, kw, "data", &data, "timeout?", &timeout); err != nil {
		return nil, err
	}
	buf, err := dataBytes(data)
	if err != nil {
		return nil, err
	}
	d, err := toTimeout(name, timeout)
	if err != nil {
		return nil, err
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if err := n.check("send"); err != nil {
		return nil, err
	}
	if n.kind == "udp" && len(buf) > udpMaxPacket {
		return nil, fmt.Errorf("udp.send: datagram larger than %d bytes", udpMaxPacket)
	}
	if n.sent+int64(len(buf)) > n.cap {
		return nil, fmt.Errorf("%s: more than %d bytes sent on this connection", name, n.cap)
	}
	_ = n.c.SetWriteDeadline(n.deadline(d))
	w, err := n.c.Write(buf)
	n.sent += int64(w)
	if err != nil {
		return nil, n.wrapErr("send", err)
	}
	return starlark.MakeInt(w), nil
}

func (n *netConn) recv(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	max := 4096
	timeout := secs(5.0)
	name := n.kind + ".recv"
	if err := starlark.UnpackArgs(name, args, kw, "max?", &max, "timeout?", &timeout); err != nil {
		return nil, err
	}
	d, err := toTimeout(name, timeout)
	if err != nil {
		return nil, err
	}
	if max < 1 || int64(max) > n.cap {
		return nil, fmt.Errorf("%s: max must be between 1 and %d", name, n.cap)
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if err := n.check("recv"); err != nil {
		return nil, err
	}
	if len(n.rbuf) > 0 {
		k := min(max, len(n.rbuf))
		out := append([]byte(nil), n.rbuf[:k]...)
		n.rbuf = n.rbuf[k:]
		return starlark.Bytes(out), nil
	}
	buf := make([]byte, min(max, 65535))
	_ = n.c.SetReadDeadline(n.deadline(d))
	r, err := n.c.Read(buf)
	if err := n.account(int64(r)); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	if err != nil && r == 0 && !errors.Is(err, io.EOF) {
		return nil, n.wrapErr("recv", err)
	}
	return starlark.Bytes(buf[:r]), nil
}

func (n *netConn) account(k int64) error {
	n.got += k
	if n.got > n.cap {
		n.c.Close()
		return fmt.Errorf("more than %d bytes received on this connection", n.cap)
	}
	return nil
}

func (n *netConn) recvUntil(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var delim starlark.Value
	max := 65536
	timeout := secs(5.0)
	name := n.kind + ".recv_until"
	if err := starlark.UnpackArgs(name, args, kw, "delim", &delim, "max?", &max, "timeout?", &timeout); err != nil {
		return nil, err
	}
	dl, err := dataBytes(delim)
	if err != nil || len(dl) == 0 {
		return nil, fmt.Errorf("%s: delim must be a non-empty string or bytes", name)
	}
	d, err := toTimeout(name, timeout)
	if err != nil {
		return nil, err
	}
	if max < len(dl) || int64(max) > n.cap {
		return nil, fmt.Errorf("%s: max must be between %d and %d", name, len(dl), n.cap)
	}
	n.mu.Lock()
	defer n.mu.Unlock()
	if err := n.check("recv_until"); err != nil {
		return nil, err
	}
	dd := n.deadline(d)
	buf := make([]byte, netRecvBuf)
	for {
		if i := bytes.Index(n.rbuf, dl); i >= 0 && i+len(dl) <= max {
			out := append([]byte(nil), n.rbuf[:i+len(dl)]...)
			n.rbuf = n.rbuf[i+len(dl):]
			return starlark.Bytes(out), nil
		}
		if len(n.rbuf) >= max {
			return nil, fmt.Errorf("%s: delimiter not found in the first %d bytes", name, max)
		}
		_ = n.c.SetReadDeadline(dd)
		r, err := n.c.Read(buf)
		n.rbuf = append(n.rbuf, buf[:r]...)
		if aerr := n.account(int64(r)); aerr != nil {
			return nil, fmt.Errorf("%s: %w", name, aerr)
		}
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil, fmt.Errorf("%s: connection closed before the delimiter (%d bytes read)", name, len(n.rbuf))
			}
			return nil, n.wrapErr("recv_until", err)
		}
	}
}

func (n *netConn) closeFn(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs(n.kind+".close", args, kw); err != nil {
		return nil, err
	}
	n.shut()
	return starlark.None, nil
}

// readAll reads until EOF, the limit or the deadline. A timeout after some
// data arrived is not an error.
func (n *netConn) readAll(max int, d time.Duration) ([]byte, error) {
	n.mu.Lock()
	defer n.mu.Unlock()
	dd := n.deadline(d)
	out := append([]byte(nil), n.rbuf...)
	n.rbuf = nil
	buf := make([]byte, netRecvBuf)
	for len(out) < max {
		_ = n.c.SetReadDeadline(dd)
		r, err := n.c.Read(buf)
		out = append(out, buf[:r]...)
		if aerr := n.account(int64(r)); aerr != nil {
			return nil, aerr
		}
		if err != nil {
			var ne net.Error
			if errors.Is(err, io.EOF) || (errors.As(err, &ne) && ne.Timeout() && len(out) > 0) {
				break
			}
			return nil, err
		}
	}
	if len(out) > max {
		out = out[:max]
	}
	return out, nil
}

// ---- modules ----

func (e *Extension) tcpModule() *starlarkstruct.Module {
	return module("tcp", starlark.StringDict{
		"connect": e.fn("tcp.connect", e.biTCPConnect),
		"request": e.fn("tcp.request", e.biTCPRequest),
	})
}

func (e *Extension) tlsModule() *starlarkstruct.Module {
	return module("tls", starlark.StringDict{
		"connect": e.fn("tls.connect", e.biTLSConnect),
		"cert":    e.fn("tls.cert", e.biTLSCert),
	})
}

func (e *Extension) udpModule() *starlarkstruct.Module {
	return module("udp", starlark.StringDict{
		"open":    e.fn("udp.open", e.biUDPOpen),
		"send":    e.fn("udp.send", e.biUDPSend),
		"request": e.fn("udp.request", e.biUDPRequest),
	})
}

func (e *Extension) dnsModule() *starlarkstruct.Module {
	return module("dns", starlark.StringDict{"lookup": e.fn("dns.lookup", e.biDNSLookup)})
}

// dialNet checks the network permission for host:port and dials it.
func (e *Extension) dialNet(th *starlark.Thread, name, network, host string, port int, d time.Duration) (net.Conn, error) {
	if err := checkHostName(host); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	if err := checkPort(port); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	if err := e.require(PermNetwork, net.JoinHostPort(strings.ToLower(host), strconv.Itoa(port))); err != nil {
		return nil, err
	}
	c, err := dialHost(callCtx(th), network, host, port, d, !e.hasUnscoped(PermNetwork))
	if err != nil {
		return nil, fmt.Errorf("%s %s: %w", name, net.JoinHostPort(host, strconv.Itoa(port)), err)
	}
	return c, nil
}

func (e *Extension) openTCP(th *starlark.Thread, name, host string, port int, timeout secs) (*netConn, error) {
	d, err := toTimeout(name, timeout)
	if err != nil {
		return nil, err
	}
	c, err := e.dialNet(th, name, "tcp", host, port, d)
	if err != nil {
		return nil, err
	}
	return e.newNetConn(th, "tcp", net.JoinHostPort(host, strconv.Itoa(port)), c)
}

func (e *Extension) biTCPConnect(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var host string
	var port int
	timeout := secs(5.0)
	if err := starlark.UnpackArgs("tcp.connect", args, kw, "host", &host, "port", &port, "timeout?", &timeout); err != nil {
		return nil, err
	}
	n, err := e.openTCP(th, "tcp.connect", host, port, timeout)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (e *Extension) biTCPRequest(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var host string
	var port int
	var data, until starlark.Value = starlark.None, starlark.None
	max := 65536
	timeout := secs(5.0)
	if err := starlark.UnpackArgs("tcp.request", args, kw, "host", &host, "port", &port, "data", &data, "until?", &until, "max?", &max, "timeout?", &timeout); err != nil {
		return nil, err
	}
	return e.requestOn(th, "tcp.request", func() (*netConn, error) { return e.openTCP(th, "tcp.request", host, port, timeout) }, data, until, max, timeout)
}

// requestOn sends data on a fresh connection and reads the reply up to a
// delimiter or until the peer closes, then closes the connection.
func (e *Extension) requestOn(th *starlark.Thread, name string, open func() (*netConn, error), data, until starlark.Value, max int, timeout secs) (starlark.Value, error) {
	buf, err := dataBytes(data)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}
	if max < 1 || int64(max) > e.opts.Limits.MaxHTTPBytes {
		return nil, fmt.Errorf("%s: max must be between 1 and %d", name, e.opts.Limits.MaxHTTPBytes)
	}
	n, err := open()
	if err != nil {
		return nil, err
	}
	defer n.shut()
	if _, err := n.send(th, nil, starlark.Tuple{starlark.Bytes(buf)}, []starlark.Tuple{{starlark.String("timeout"), starlark.Float(float64(timeout))}}); err != nil {
		return nil, err
	}
	if until != starlark.None {
		return n.recvUntil(th, nil, starlark.Tuple{until}, []starlark.Tuple{
			{starlark.String("max"), starlark.MakeInt(max)}, {starlark.String("timeout"), starlark.Float(float64(timeout))}})
	}
	d, _ := toTimeout(name, timeout)
	out, err := n.readAll(max, d)
	if err != nil {
		return nil, n.wrapErr("request", err)
	}
	return starlark.Bytes(out), nil
}

// ---- tls ----

func (e *Extension) tlsHandshake(th *starlark.Thread, name, host string, port int, serverName string, timeout secs, verify bool) (*tls.Conn, error) {
	d, err := toTimeout(name, timeout)
	if err != nil {
		return nil, err
	}
	c, err := e.dialNet(th, name, "tcp", host, port, d)
	if err != nil {
		return nil, err
	}
	if serverName == "" {
		serverName = host
	}
	tc := tls.Client(c, &tls.Config{ServerName: serverName, MinVersion: tls.VersionTLS12, InsecureSkipVerify: !verify})
	ctx, cancel := context.WithTimeout(callCtx(th), d)
	defer cancel()
	if err := tc.HandshakeContext(ctx); err != nil {
		c.Close()
		return nil, fmt.Errorf("%s %s: %w", name, net.JoinHostPort(host, strconv.Itoa(port)), err)
	}
	return tc, nil
}

func (e *Extension) biTLSConnect(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var host, sni string
	port := 443
	timeout := secs(5.0)
	if err := starlark.UnpackArgs("tls.connect", args, kw, "host", &host, "port?", &port, "timeout?", &timeout, "server_name?", &sni); err != nil {
		return nil, err
	}
	tc, err := e.tlsHandshake(th, "tls.connect", host, port, sni, timeout, true)
	if err != nil {
		return nil, err
	}
	n, err := e.newNetConn(th, "tls", net.JoinHostPort(host, strconv.Itoa(port)), tc)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (e *Extension) biTLSCert(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var host, sni string
	port := 443
	timeout := secs(5.0)
	if err := starlark.UnpackArgs("tls.cert", args, kw, "host", &host, "port?", &port, "timeout?", &timeout, "server_name?", &sni); err != nil {
		return nil, err
	}
	tc, err := e.tlsHandshake(th, "tls.cert", host, port, sni, timeout, false)
	if err != nil {
		return nil, err
	}
	defer tc.Close()
	st := tc.ConnectionState()
	if len(st.PeerCertificates) == 0 {
		return nil, errors.New("tls.cert: the server sent no certificate")
	}
	name := sni
	if name == "" {
		name = host
	}
	return certInfo(st, name, time.Now()), nil
}

func certInfo(st tls.ConnectionState, name string, now time.Time) starlark.Value {
	leaf := st.PeerCertificates[0]
	inter := x509.NewCertPool()
	for _, c := range st.PeerCertificates[1:] {
		inter.AddCert(c)
	}
	verified, verr := true, ""
	if _, err := leaf.Verify(x509.VerifyOptions{DNSName: name, Intermediates: inter, CurrentTime: now}); err != nil {
		verified, verr = false, err.Error()
	}
	var sans []string
	sans = append(sans, leaf.DNSNames...)
	for _, ip := range leaf.IPAddresses {
		sans = append(sans, ip.String())
	}
	sans = append(sans, leaf.EmailAddresses...)
	for _, u := range leaf.URIs {
		sans = append(sans, u.String())
	}
	sum := sha256.Sum256(leaf.Raw)
	chain := make([]any, 0, len(st.PeerCertificates))
	for _, c := range st.PeerCertificates {
		chain = append(chain, c.Subject.String())
	}
	left := leaf.NotAfter.Sub(now)
	days := int64(left.Hours() / 24)
	if left < 0 && left.Hours() != float64(days*24) {
		days--
	}
	return goToStarlark(map[string]any{
		"subject":            leaf.Subject.String(),
		"common_name":        leaf.Subject.CommonName,
		"issuer":             leaf.Issuer.String(),
		"issuer_cn":          leaf.Issuer.CommonName,
		"not_before":         leaf.NotBefore.UTC().Format(time.RFC3339),
		"not_after":          leaf.NotAfter.UTC().Format(time.RFC3339),
		"days_left":          days,
		"expired":            now.After(leaf.NotAfter),
		"sans":               sans,
		"serial":             leaf.SerialNumber.Text(16),
		"fingerprint_sha256": hex.EncodeToString(sum[:]),
		"signature":          leaf.SignatureAlgorithm.String(),
		"tls_version":        tls.VersionName(st.Version),
		"verified":           verified,
		"verify_error":       verr,
		"chain":              chain,
	})
}

// ---- udp ----

func (e *Extension) openUDP(th *starlark.Thread, name, host string, port int) (*netConn, error) {
	c, err := e.dialNet(th, name, "udp", host, port, 5*time.Second)
	if err != nil {
		return nil, err
	}
	return e.newNetConn(th, "udp", net.JoinHostPort(host, strconv.Itoa(port)), c)
}

func (e *Extension) biUDPOpen(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var host string
	var port int
	if err := starlark.UnpackArgs("udp.open", args, kw, "host", &host, "port", &port); err != nil {
		return nil, err
	}
	n, err := e.openUDP(th, "udp.open", host, port)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (e *Extension) biUDPSend(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var host string
	var port int
	var data starlark.Value
	if err := starlark.UnpackArgs("udp.send", args, kw, "host", &host, "port", &port, "data", &data); err != nil {
		return nil, err
	}
	n, err := e.openUDP(th, "udp.send", host, port)
	if err != nil {
		return nil, err
	}
	defer n.shut()
	return n.send(th, nil, starlark.Tuple{data}, nil)
}

func (e *Extension) biUDPRequest(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var host string
	var port int
	var data starlark.Value
	max := 4096
	timeout := secs(2.0)
	if err := starlark.UnpackArgs("udp.request", args, kw, "host", &host, "port", &port, "data", &data, "max?", &max, "timeout?", &timeout); err != nil {
		return nil, err
	}
	n, err := e.openUDP(th, "udp.request", host, port)
	if err != nil {
		return nil, err
	}
	defer n.shut()
	if _, err := n.send(th, nil, starlark.Tuple{data}, nil); err != nil {
		return nil, err
	}
	return n.recv(th, nil, nil, []starlark.Tuple{{starlark.String("max"), starlark.MakeInt(max)}, {starlark.String("timeout"), starlark.Float(float64(timeout))}})
}

// ---- dns ----

func (e *Extension) biDNSLookup(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var name, typ, server string
	typ = "A"
	timeout := secs(5.0)
	if err := starlark.UnpackArgs("dns.lookup", args, kw, "name", &name, "type?", &typ, "server?", &server, "timeout?", &timeout); err != nil {
		return nil, err
	}
	d, err := toTimeout("dns.lookup", timeout)
	if err != nil {
		return nil, err
	}
	name = strings.TrimSuffix(strings.TrimSpace(name), ".")
	if err := checkHostName(name); err != nil {
		return nil, fmt.Errorf("dns.lookup: %w", err)
	}
	if err := e.require(PermNetwork, strings.ToLower(name)); err != nil {
		return nil, err
	}
	typ = strings.ToUpper(typ)
	ctx, cancel := context.WithTimeout(callCtx(th), d)
	defer cancel()
	r := net.DefaultResolver
	if server != "" {
		if r, err = e.dnsServer(th, server, d); err != nil {
			return nil, err
		}
	}
	var out []any
	trim := func(s string) string { return strings.TrimSuffix(s, ".") }
	switch typ {
	case "A", "AAAA":
		family := "ip4"
		if typ == "AAAA" {
			family = "ip6"
		}
		ips, err := r.LookupIP(ctx, family, name)
		if err != nil {
			return nil, dnsErr(name, typ, err)
		}
		for _, ip := range ips {
			out = append(out, ip.String())
		}
	case "MX":
		mx, err := r.LookupMX(ctx, name)
		if err != nil {
			return nil, dnsErr(name, typ, err)
		}
		for _, m := range mx {
			out = append(out, map[string]any{"host": trim(m.Host), "pref": int(m.Pref)})
		}
	case "TXT":
		txt, err := r.LookupTXT(ctx, name)
		if err != nil {
			return nil, dnsErr(name, typ, err)
		}
		for _, s := range txt {
			out = append(out, s)
		}
	case "NS":
		ns, err := r.LookupNS(ctx, name)
		if err != nil {
			return nil, dnsErr(name, typ, err)
		}
		for _, s := range ns {
			out = append(out, trim(s.Host))
		}
	case "CNAME":
		c, err := r.LookupCNAME(ctx, name)
		if err != nil {
			return nil, dnsErr(name, typ, err)
		}
		if c = trim(c); !strings.EqualFold(c, name) {
			out = append(out, c)
		}
	case "PTR":
		if _, ok := parseHostIP(name); !ok {
			return nil, errors.New("dns.lookup: PTR needs an IP address as name")
		}
		hosts, err := r.LookupAddr(ctx, name)
		if err != nil {
			return nil, dnsErr(name, typ, err)
		}
		for _, s := range hosts {
			out = append(out, trim(s))
		}
	default:
		return nil, fmt.Errorf("dns.lookup: unsupported type %q (A, AAAA, MX, TXT, CNAME, NS, PTR)", typ)
	}
	if len(out) > netMaxRecs {
		out = out[:netMaxRecs]
	}
	return goToStarlark(out), nil
}

func dnsErr(name, typ string, err error) error {
	var de *net.DNSError
	if errors.As(err, &de) && de.IsNotFound {
		return fmt.Errorf("dns.lookup %s %s: no such record", typ, name)
	}
	return fmt.Errorf("dns.lookup %s %s: %w", typ, name, err)
}

// dnsServer builds a resolver that talks only to the given server (host or
// host:port, port 53 by default). The server needs network permission too.
func (e *Extension) dnsServer(th *starlark.Thread, server string, d time.Duration) (*net.Resolver, error) {
	host, ps := splitHostPortLoose(server)
	port := 53
	if ps != "" {
		p, err := strconv.Atoi(ps)
		if err != nil {
			return nil, fmt.Errorf("dns.lookup: bad server port in %q", server)
		}
		port = p
	}
	if err := checkHostName(host); err != nil {
		return nil, fmt.Errorf("dns.lookup: server: %w", err)
	}
	if err := checkPort(port); err != nil {
		return nil, fmt.Errorf("dns.lookup: server: %w", err)
	}
	if err := e.require(PermNetwork, net.JoinHostPort(strings.ToLower(host), strconv.Itoa(port))); err != nil {
		return nil, err
	}
	enforce := !e.hasUnscoped(PermNetwork)
	return &net.Resolver{PreferGo: true, Dial: func(ctx context.Context, network, _ string) (net.Conn, error) {
		nw := "udp"
		if strings.HasPrefix(network, "tcp") {
			nw = "tcp"
		}
		return dialHost(ctx, nw, host, port, d, enforce)
	}}, nil
}

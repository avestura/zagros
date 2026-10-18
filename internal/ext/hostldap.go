package ext

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/go-ldap/ldap/v3"
	"go.starlark.net/starlark"
)

const (
	maxLDAPConns   = 8
	defLDAPSizeLim = 1000
)

func (e *Extension) ldapModule() starlark.Value {
	return module("ldap", starlark.StringDict{
		"connect":       e.fn("ldap.connect", e.biLDAPConnect),
		"escape_filter": starlark.NewBuiltin("ldap.escape_filter", biLDAPEscapeFilter),
		"escape_dn":     starlark.NewBuiltin("ldap.escape_dn", biLDAPEscapeDN),
	})
}

func biLDAPEscapeFilter(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var s string
	if err := starlark.UnpackArgs("ldap.escape_filter", args, kw, "value", &s); err != nil {
		return nil, err
	}
	return starlark.String(ldap.EscapeFilter(s)), nil
}

func biLDAPEscapeDN(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var s string
	if err := starlark.UnpackArgs("ldap.escape_dn", args, kw, "value", &s); err != nil {
		return nil, err
	}
	return starlark.String(ldap.EscapeDN(s)), nil
}

// ldapConn is one open directory connection of one extension.
type ldapConn struct {
	mu       sync.Mutex
	conn     *ldap.Conn
	ext      *Extension
	target   string // host:port, what the permissions are checked against
	host     string
	port     string
	ldaps    bool
	startTLS bool
	timeout  time.Duration
	max      int
}

// dialLDAP opens a connection that verifies the server certificate (ldaps or
// StartTLS) against the system roots.
func dialLDAP(ctx context.Context, c *ldapConn) (*ldap.Conn, error) {
	d := &net.Dialer{Timeout: c.timeout}
	addr := net.JoinHostPort(c.host, c.port)
	cfg := &tls.Config{ServerName: c.host, MinVersion: tls.VersionTLS12}
	var raw net.Conn
	var err error
	if c.ldaps {
		raw, err = (&tls.Dialer{NetDialer: d, Config: cfg}).DialContext(ctx, "tcp", addr)
	} else {
		raw, err = d.DialContext(ctx, "tcp", addr)
	}
	if err != nil {
		return nil, err
	}
	l := ldap.NewConn(raw, c.ldaps)
	l.Start()
	l.SetTimeout(c.timeout)
	if c.startTLS {
		if err := l.StartTLS(cfg); err != nil {
			l.Close()
			return nil, err
		}
	}
	return l, nil
}

func isLoopbackHost(h string) bool {
	if strings.EqualFold(h, "localhost") {
		return true
	}
	ip := net.ParseIP(h)
	return ip != nil && ip.IsLoopback()
}

func (e *Extension) biLDAPConnect(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var rawURL, bindDN string
	var pw starlark.Value = starlark.None
	startTLS := false
	timeout := 15
	if err := starlark.UnpackArgs("ldap.connect", args, kw, "url", &rawURL, "bind_dn?", &bindDN, "password?", &pw, "start_tls?", &startTLS, "timeout?", &timeout); err != nil {
		return nil, err
	}
	u, err := url.Parse(rawURL)
	if err != nil || (u.Scheme != "ldap" && u.Scheme != "ldaps") || u.Hostname() == "" || (u.Path != "" && u.Path != "/") || u.User != nil || u.RawQuery != "" {
		return nil, fmt.Errorf("ldap.connect: url must look like ldaps://host[:port], got %q", rawURL)
	}
	if err := validHost("ldap.connect", u.Hostname()); err != nil {
		return nil, err
	}
	port := u.Port()
	if port == "" {
		port = "389"
		if u.Scheme == "ldaps" {
			port = "636"
		}
	}
	if n, err := strconv.Atoi(port); err != nil || n < 1 || n > 65535 {
		return nil, fmt.Errorf("ldap.connect: invalid port %q", port)
	}
	if timeout < 1 || timeout > 300 {
		return nil, errors.New("ldap.connect: timeout must be 1 to 300 seconds")
	}
	if startTLS && u.Scheme == "ldaps" {
		return nil, errors.New("ldap.connect: start_tls is for ldap:// urls; ldaps:// is already encrypted")
	}
	target := net.JoinHostPort(u.Hostname(), port)
	if err := e.require(PermLDAP, target); err != nil {
		if !e.allows(PermLDAPWrite, target) {
			return nil, err
		}
	}
	password, err := e.secretValue("ldap.connect", pw)
	if err != nil {
		return nil, err
	}
	if (bindDN == "") != (password == "") {
		return nil, errors.New("ldap.connect: give both bind_dn and password, or neither for an anonymous connection")
	}
	encrypted := u.Scheme == "ldaps" || startTLS
	if password != "" && !encrypted && !isLoopbackHost(u.Hostname()) {
		return nil, errors.New("ldap.connect: refusing to send a password over an unencrypted connection; use ldaps:// or start_tls=True")
	}
	e.host.mu.Lock()
	n := len(e.host.ldaps)
	e.host.mu.Unlock()
	if n >= maxLDAPConns {
		return nil, fmt.Errorf("ldap.connect: at most %d connections per extension; close one first", maxLDAPConns)
	}
	c := &ldapConn{ext: e, target: target, host: u.Hostname(), port: port, ldaps: u.Scheme == "ldaps", startTLS: startTLS,
		timeout: time.Duration(timeout) * time.Second, max: e.opts.Limits.MaxEntries}
	ctx, cancel := context.WithTimeout(callCtx(th), c.timeout)
	defer cancel()
	l, err := dialLDAP(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("ldap.connect: %s: %w", target, err)
	}
	if bindDN != "" {
		if err := l.Bind(bindDN, password); err != nil {
			l.Close()
			return nil, fmt.Errorf("ldap.connect: bind as %q: %w", bindDN, scrubSecret(err, password))
		}
	}
	c.conn = l
	e.host.mu.Lock()
	if e.host.ldaps == nil {
		e.host.ldaps = map[*ldapConn]bool{}
	}
	e.host.ldaps[c] = true
	e.host.mu.Unlock()
	return c, nil
}

func (e *Extension) closeLDAP() {
	e.host.mu.Lock()
	all := e.host.ldaps
	e.host.ldaps = nil
	e.host.mu.Unlock()
	for c := range all {
		c.conn.Close()
	}
}

var _ starlark.HasAttrs = (*ldapConn)(nil)

func (c *ldapConn) String() string        { return "<ldap.Connection " + c.target + ">" }
func (c *ldapConn) Type() string          { return "ldap.Connection" }
func (c *ldapConn) Freeze()               {}
func (c *ldapConn) Truth() starlark.Bool  { return true }
func (c *ldapConn) Hash() (uint32, error) { return 0, errors.New("unhashable type: ldap.Connection") }

func (c *ldapConn) AttrNames() []string {
	return []string{"add", "check_bind", "close", "delete", "modify", "search"}
}

func (c *ldapConn) Attr(name string) (starlark.Value, error) {
	var f func(*starlark.Thread, *starlark.Builtin, starlark.Tuple, []starlark.Tuple) (starlark.Value, error)
	switch name {
	case "search":
		f = c.search
	case "check_bind":
		f = c.checkBind
	case "add":
		f = c.add
	case "modify":
		f = c.modify
	case "delete":
		f = c.del
	case "close":
		f = c.close
	default:
		return nil, nil
	}
	return starlark.NewBuiltin("ldap."+name, f), nil
}

var ldapScopes = map[string]int{"base": ldap.ScopeBaseObject, "one": ldap.ScopeSingleLevel, "sub": ldap.ScopeWholeSubtree}

func (c *ldapConn) search(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var base, filter, scope string
	filter, scope = "(objectClass=*)", "sub"
	var attrs starlark.Value = starlark.None
	limit := defLDAPSizeLim
	if err := starlark.UnpackArgs("ldap.search", args, kw, "base", &base, "filter?", &filter, "scope?", &scope, "attributes?", &attrs, "size_limit?", &limit); err != nil {
		return nil, err
	}
	if err := c.ext.require(PermLDAP, c.target); err != nil {
		return nil, err
	}
	sc, ok := ldapScopes[scope]
	if !ok {
		return nil, fmt.Errorf(`ldap.search: scope must be "base", "one" or "sub", got %q`, scope)
	}
	if limit < 1 {
		return nil, errors.New("ldap.search: size_limit must be at least 1")
	}
	if limit > c.max {
		limit = c.max
	}
	var names []string
	if _, none := attrs.(starlark.NoneType); !none {
		var err error
		if names, err = ldapStrings("ldap.search attributes", attrs); err != nil {
			return nil, err
		}
	}
	req := ldap.NewSearchRequest(base, sc, ldap.NeverDerefAliases, limit, int(c.timeout.Seconds()), false, filter, names, nil)
	c.mu.Lock()
	res, err := c.conn.Search(req)
	c.mu.Unlock()
	if err != nil && !(res != nil && ldap.IsErrorWithCode(err, ldap.LDAPResultSizeLimitExceeded)) {
		return nil, fmt.Errorf("ldap.search: %w", err)
	}
	out := make([]starlark.Value, 0, len(res.Entries))
	for _, en := range res.Entries {
		d := starlark.NewDict(len(en.Attributes) + 1)
		d.SetKey(starlark.String("dn"), starlark.String(en.DN))
		for _, a := range en.Attributes {
			vals := make([]starlark.Value, len(a.ByteValues))
			for i, v := range a.ByteValues {
				if utf8.Valid(v) {
					vals[i] = starlark.String(v)
				} else {
					vals[i] = starlark.Bytes(v)
				}
			}
			d.SetKey(starlark.String(a.Name), starlark.NewList(vals))
		}
		out = append(out, d)
	}
	return starlark.NewList(out), nil
}

func (c *ldapConn) checkBind(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var dn string
	var pw starlark.Value
	if err := starlark.UnpackArgs("ldap.check_bind", args, kw, "dn", &dn, "password", &pw); err != nil {
		return nil, err
	}
	if err := c.ext.require(PermLDAP, c.target); err != nil {
		return nil, err
	}
	password, err := c.ext.secretValue("ldap.check_bind", pw)
	if err != nil {
		return nil, err
	}
	if dn == "" || password == "" {
		return nil, errors.New("ldap.check_bind: dn and password must not be empty (an empty password is an anonymous bind)")
	}
	if !c.ldaps && !c.startTLS && !isLoopbackHost(c.host) {
		return nil, errors.New("ldap.check_bind: refusing to send a password over an unencrypted connection")
	}
	ctx, cancel := context.WithTimeout(callCtx(th), c.timeout)
	defer cancel()
	l, err := dialLDAP(ctx, c) // a separate connection, so this one keeps its identity
	if err != nil {
		return nil, fmt.Errorf("ldap.check_bind: %w", err)
	}
	defer l.Close()
	if err := l.Bind(dn, password); err != nil {
		if ldap.IsErrorWithCode(err, ldap.LDAPResultInvalidCredentials) {
			return starlark.False, nil
		}
		return nil, fmt.Errorf("ldap.check_bind: %w", scrubSecret(err, password))
	}
	return starlark.True, nil
}

func (c *ldapConn) add(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var dn string
	var attrs *starlark.Dict
	if err := starlark.UnpackArgs("ldap.add", args, kw, "dn", &dn, "attributes", &attrs); err != nil {
		return nil, err
	}
	if err := c.ext.require(PermLDAPWrite, c.target); err != nil {
		return nil, err
	}
	m, err := ldapAttrMap("ldap.add", attrs)
	if err != nil {
		return nil, err
	}
	req := ldap.NewAddRequest(dn, nil)
	for _, k := range sortedKeys(m) {
		req.Attribute(k, m[k])
	}
	c.mu.Lock()
	err = c.conn.Add(req)
	c.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("ldap.add: %w", err)
	}
	return starlark.None, nil
}

func (c *ldapConn) modify(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var dn string
	var add, replace, del starlark.Value = starlark.None, starlark.None, starlark.None
	if err := starlark.UnpackArgs("ldap.modify", args, kw, "dn", &dn, "replace?", &replace, "add?", &add, "delete?", &del); err != nil {
		return nil, err
	}
	if err := c.ext.require(PermLDAPWrite, c.target); err != nil {
		return nil, err
	}
	req := ldap.NewModifyRequest(dn, nil)
	changes := 0
	for _, op := range []struct {
		name string
		v    starlark.Value
		fn   func(string, []string)
	}{{"replace", replace, req.Replace}, {"add", add, req.Add}, {"delete", del, req.Delete}} {
		d, ok := op.v.(*starlark.Dict)
		if !ok {
			if _, none := op.v.(starlark.NoneType); none {
				continue
			}
			if l, isList := op.v.(*starlark.List); isList && op.name == "delete" { // delete=["mail"] drops whole attributes
				names, err := ldapStrings("ldap.modify delete", l)
				if err != nil {
					return nil, err
				}
				for _, n := range names {
					req.Delete(n, nil)
					changes++
				}
				continue
			}
			return nil, fmt.Errorf("ldap.modify: %s must be a dict {attribute: value(s)}, got %s", op.name, op.v.Type())
		}
		m, err := ldapAttrMap("ldap.modify "+op.name, d)
		if err != nil {
			return nil, err
		}
		for _, k := range sortedKeys(m) {
			op.fn(k, m[k])
			changes++
		}
	}
	if changes == 0 {
		return nil, errors.New("ldap.modify: nothing to change (give replace, add or delete)")
	}
	c.mu.Lock()
	err := c.conn.Modify(req)
	c.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("ldap.modify: %w", err)
	}
	return starlark.None, nil
}

func (c *ldapConn) del(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var dn string
	if err := starlark.UnpackArgs("ldap.delete", args, kw, "dn", &dn); err != nil {
		return nil, err
	}
	if err := c.ext.require(PermLDAPWrite, c.target); err != nil {
		return nil, err
	}
	if strings.TrimSpace(dn) == "" {
		return nil, errors.New("ldap.delete: dn must not be empty")
	}
	c.mu.Lock()
	err := c.conn.Del(ldap.NewDelRequest(dn, nil))
	c.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("ldap.delete: %w", err)
	}
	return starlark.None, nil
}

func (c *ldapConn) close(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs("ldap.close", args, kw); err != nil {
		return nil, err
	}
	c.ext.host.mu.Lock()
	delete(c.ext.host.ldaps, c)
	c.ext.host.mu.Unlock()
	c.conn.Close()
	return starlark.None, nil
}

func sortedKeys(m map[string][]string) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

// ldapStrings reads a string or a list of strings.
func ldapStrings(what string, v starlark.Value) ([]string, error) {
	switch x := v.(type) {
	case starlark.String:
		return []string{string(x)}, nil
	case starlark.Bytes:
		return []string{string(x)}, nil
	case starlark.Indexable:
		out := make([]string, 0, x.Len())
		for i := 0; i < x.Len(); i++ {
			s, err := ldapStrings(what, x.Index(i))
			if err != nil || len(s) != 1 {
				return nil, fmt.Errorf("%s: want strings, got %s", what, x.Index(i).Type())
			}
			out = append(out, s[0])
		}
		return out, nil
	}
	return nil, fmt.Errorf("%s: want a string or a list of strings, got %s", what, v.Type())
}

func ldapAttrMap(what string, d *starlark.Dict) (map[string][]string, error) {
	out := map[string][]string{}
	for _, kv := range d.Items() {
		k, ok := starlark.AsString(kv[0])
		if !ok || k == "" || strings.ContainsAny(k, " \t\r\n=,") {
			return nil, fmt.Errorf("%s: attribute names must be non-empty strings, got %s", what, kv[0])
		}
		vals, err := ldapStrings(what+" "+k, kv[1])
		if err != nil {
			return nil, err
		}
		out[k] = vals
	}
	return out, nil
}

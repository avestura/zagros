package ext

import (
	"net"
	"strings"
)

// Permissions of the database and directory modules, scoped by host[:port].
const (
	PermSQL       = "sql"
	PermLDAP      = "ldap"
	PermLDAPWrite = "ldap.write"
)

func init() {
	for _, k := range []string{PermSQL, PermLDAP, PermLDAPWrite} {
		scoped[k] = true
		allKinds[k] = true
	}
}

// splitScope splits "host" or "host:port" (also "[::1]:389") without
// treating a bare IPv6 address as host and port.
func splitScope(s string) (host, port string) {
	if h, p, err := net.SplitHostPort(s); err == nil {
		return strings.ToLower(h), p
	}
	return strings.ToLower(strings.Trim(s, "[]")), ""
}

// hostPortMatches reports whether target ("host:port") is covered by scope.
// A scope without a port covers every port; a host also covers its subdomains
// (but an IP address only matches itself).
func hostPortMatches(scope, target string) bool {
	sh, sp := splitScope(scope)
	th, tp := splitScope(target)
	if sp != "" && sp != tp {
		return false
	}
	if th == sh {
		return true
	}
	return net.ParseIP(sh) == nil && strings.HasSuffix(th, "."+sh)
}

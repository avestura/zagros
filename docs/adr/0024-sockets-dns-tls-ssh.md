---
status: accepted
date: 2026-10-19
---
# ADR 0024: Raw sockets, DNS, TLS certificates and SSH for extensions

## Context
Extensions could speak HTTP only. Admin and network tools need other protocols: a banner from a mail server, a certificate expiry check, an MX lookup, a command on a build host. Each of these reaches outside the sandbox and needs a permission with a scope the user can read.

## Decision
- `tcp`, `udp`, `tls` and `dns` use the existing `network` permission. The scope grammar grows to `host` or `host:port`: a bare host still covers the host, its subdomains and every port; a scope with a port covers that port only. `http` and `dns` pass the host without a port, so a scope with a port never matches them.
- The host is resolved once, the answer is checked, and that address is dialled, so a second DNS answer cannot redirect the connection. With a scoped grant, a name that resolves to a loopback, private, link-local, CGNAT or unspecified address is refused unless the name is `localhost`. To reach a machine on the local network the extension is granted its IP address and connects by address. An unscoped `network` grant skips the check, since it could name the address directly. `http` is unchanged and does not do this check yet.
- Connections are bounded: at most 8 open per extension, 2 minutes of life, 8 MiB in each direction (`Limits.MaxHTTPBytes`), timeouts up to 300 seconds. They close when the call that opened them ends (the call context cancels) and when the extension stops. A handle cannot be kept between calls because globals are frozen.
- `tls.connect` always verifies the chain and the host name, with TLS 1.2 as the minimum, and has no switch to turn that off. `tls.cert` accepts any certificate so that expired and self-signed ones can be read, sends no data, and reports `verified` and `verify_error`.
- `dns.lookup` queries the system resolver, or only a given server (pure Go resolver) when `server=` is set; that server needs its own `network` grant. The queried name is the scope target.
- `ssh` is a new permission scoped `[user@]host[:port]`; the host matches exactly (no subdomains) because the grant covers logging in as a user. It stands on its own and does not need `network`. The module uses `golang.org/x/crypto/ssh` and `github.com/pkg/sftp`, both pure Go. Passwords and private keys are read from `secrets` by name (`auth={"secret": ...}`), which also needs the `secrets` permission; raw strings are accepted but discouraged.
- Host keys are verified: trust on first use, recorded per host and port and key type in `ssh_known_hosts.json` in the extension's data folder (not in `storage`, so scripts cannot rewrite it). A changed key is an error naming both fingerprints. `host_key=` pins a fingerprint and overrides the record; `trust_new=False` refuses unknown hosts and shows the fingerprint. There is no option to skip verification.
- `exec` returns stdout, stderr and the exit code, with output capped and a timeout; SFTP covers get, put, download, upload, list, stat, remove and mkdir. `download` and `upload` need `fs.write` and `fs.read` for the local path.

## Consequences
The Windows binary grows by the SSH and SFTP packages (measured in the change that added them). A `network:host` grant now also allows raw TCP and UDP to that host on any port, which the consent text says. Extensions cannot keep a socket open across calls; a long-lived listener or session is out of scope. Server-side forwarding, agent auth, listening sockets and rename over SFTP are not provided. Changing the key of a pinned or recorded host needs `host_key=` from the extension or removing the entry from the data folder.

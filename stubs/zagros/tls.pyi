"""tls: open verified TLS connections and read the certificate of a server.

Type description, not runnable Starlark: do not `import` anything. The name `tls` is predeclared in main.star; call `tls.cert(...)` directly.
Needs: network, scoped by host or host:port (network:example.com:443). Same address rules as `tcp`.
Keywords: tls, ssl, certificate, cert expiry, https certificate, secure socket, san, issuer, check certificate, imaps, smtps

`tls.connect` always verifies the chain and the host name and needs TLS 1.2 or
newer; there is no way to switch verification off. `tls.cert` is the one place
that accepts any certificate, so that expired and self-signed ones can be read:
it only reports (see `verified`) and sends no data. The connection value has the
same send, recv, recv_until and close as `tcp.Conn`.
"""

from typing import Any, Dict, Optional, Union

Data = Union[str, bytes]

class Conn:
    addr: str
    def send(self, data: Data, timeout: float = 10) -> int:
        """Write data over the encrypted connection; returns the byte count.

        Example: n = conn.send("A1 LOGIN user pass\r\n")
        Task: send bytes on a TLS connection
        Keywords: write, transmit
        Needs: none
        """
    def recv(self, max: int = 4096, timeout: float = 5) -> bytes:
        """Read up to max bytes. Empty bytes means the peer closed. Raises on timeout.

        Example: chunk = conn.recv(1024)
        Task: read bytes from a TLS connection
        Keywords: receive, read
        Needs: none
        """
    def recv_until(self, delim: Data, max: int = 65536, timeout: float = 5) -> bytes:
        """Read up to and including delim.

        Example: line = conn.recv_until("\r\n")
        Task: read one line from a TLS connection
        Keywords: read line, delimiter, banner
        Needs: none
        """
    def close(self) -> None:
        """Close the connection early.

        Example: conn.close()
        Needs: none
        """

def connect(host: str, port: int = 443, timeout: float = 5, server_name: Optional[str] = None) -> Conn:
    """Open a TLS connection with certificate and host name verification. server_name overrides the name used for SNI and verification.

    Example: conn = tls.connect("imap.example.com", 993)
    Task: open a verified TLS connection
    Keywords: secure socket, ssl connect, imaps, smtps, tls client, encrypted connection
    Needs: network
    """

def cert(host: str, port: int = 443, timeout: float = 5, server_name: Optional[str] = None) -> Dict[str, Any]:
    """Read the server certificate. Returns subject, common_name, issuer, issuer_cn, not_before and not_after (RFC 3339 UTC), days_left (negative when expired), expired, sans (DNS names, IPs, emails), serial, fingerprint_sha256, signature, tls_version, verified (chain and name check out), verify_error and chain (subjects, leaf first).

    Example: info = tls.cert("example.com")
    Task: check when a certificate expires
    Keywords: ssl expiry, certificate expiration, days left, san, subject alternative names, issuer, fingerprint, inspect certificate, https cert info
    Needs: network
    """

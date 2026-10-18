"""tcp: open TCP connections, send and read raw bytes, or do one request and reply.

Type description, not runnable Starlark: do not `import` anything. The name `tcp` is predeclared in main.star; call `tcp.connect(...)` directly.
Needs: network, scoped by host or host:port (network:mail.example.com:25). A host-only scope covers every port on that host and its subdomains.
Keywords: tcp, socket, raw connection, port check, banner, smtp, redis, telnet, whois, line protocol, send bytes, netcat

A name that resolves to a loopback, private or link-local address is refused
unless the grant is unscoped, the name is localhost, or you connect by IP
address and grant that address. The host is resolved once and the answer is
dialled. Connections close when the call that opened them ends, after two
minutes, or when the extension stops; at most 8 are open at once and each
carries at most 8 MiB in either direction. Errors raise (`attempt(tcp.connect, h, p)`).
"""

from typing import Optional, Union

Data = Union[str, bytes]

class Conn:
    addr: str
    """host:port as given to connect."""
    def send(self, data: Data, timeout: float = 10) -> int:
        """Write data; returns the number of bytes written.

        Example: n = conn.send("PING\r\n")
        Task: send bytes on a connection
        Keywords: write, transmit, push data
        Needs: none
        """
    def recv(self, max: int = 4096, timeout: float = 5) -> bytes:
        """Read up to max bytes (whatever arrives first). Empty bytes means the peer closed. Raises on timeout.

        Example: chunk = conn.recv(1024, timeout=2)
        Task: read bytes from a connection
        Keywords: receive, read socket, get data
        Needs: none
        """
    def recv_until(self, delim: Data, max: int = 65536, timeout: float = 5) -> bytes:
        """Read until delim appears; returns everything up to and including it and keeps any extra for the next read. Raises if max bytes pass without it or the peer closes first.

        Example: line = conn.recv_until("\r\n")
        Task: read one line or message from a connection
        Keywords: read line, readline, delimiter, banner, terminator
        Needs: none
        """
    def close(self) -> None:
        """Close the connection early. It is closed anyway when the call ends.

        Example: conn.close()
        Needs: none
        """

def connect(host: str, port: int, timeout: float = 5) -> Conn:
    """Open a TCP connection. timeout (seconds, at most 300) bounds the connect.

    Example: conn = tcp.connect("mail.example.com", 25)
    Task: open a TCP connection to a host and port
    Keywords: socket connect, open port, telnet, dial, netcat, smtp, redis
    Needs: network
    """

def request(host: str, port: int, data: Data, until: Optional[Data] = None, max: int = 65536, timeout: float = 5) -> bytes:
    """Connect, send data, read the reply, close. With until the reply ends at that delimiter (included); without it the reply ends when the peer closes, max bytes arrive, or the timeout passes after some data came.

    Example: reply = tcp.request("whois.iana.org", 43, "example.com\r\n")
    Task: send one request over TCP and get the reply
    Keywords: request response, whois, one shot, query a server, raw protocol, send and receive
    Needs: network
    """

"""udp: send and receive UDP datagrams.

Type description, not runnable Starlark: do not `import` anything. The name `udp` is predeclared in main.star; call `udp.request(...)` directly.
Needs: network, scoped by host or host:port (network:time.example.com:123). Same address rules as `tcp`.
Keywords: udp, datagram, packet, ntp, syslog, wake on lan, statsd, broadcast, send packet

UDP has no delivery guarantee: a reply that never comes raises "timed out".
A socket only accepts replies from the address it was opened to. A datagram is
at most 65507 bytes; `recv(max)` drops what does not fit in max. Sockets close
when the call that opened them ends or the extension stops.
"""

from typing import Union

Data = Union[str, bytes]

class Socket:
    addr: str
    def send(self, data: Data, timeout: float = 10) -> int:
        """Send one datagram; returns the byte count.

        Example: n = sock.send(b"\x1b")
        Task: send a datagram on an open socket
        Keywords: write packet, transmit datagram
        Needs: none
        """
    def recv(self, max: int = 4096, timeout: float = 5) -> bytes:
        """Wait for one datagram and return it (cut to max bytes). Raises on timeout.

        Example: reply = sock.recv(512, timeout=2)
        Task: receive a datagram
        Keywords: read packet, wait for reply
        Needs: none
        """
    def close(self) -> None:
        """Close the socket early.

        Example: sock.close()
        Needs: none
        """

def open(host: str, port: int) -> Socket:
    """Open a socket that sends to and receives from host:port.

    Example: sock = udp.open("192.0.2.10", 5000)
    Task: open a UDP socket
    Keywords: udp socket, connect udp, datagram socket
    Needs: network
    """

def send(host: str, port: int, data: Data) -> int:
    """Send one datagram and forget it; returns the byte count.

    Example: udp.send("192.0.2.255", 9, b"\xff\xff\xff\xff\xff\xff")
    Task: send a UDP packet without waiting for a reply
    Keywords: fire and forget, wake on lan, magic packet, statsd, syslog, notify
    Needs: network
    """

def request(host: str, port: int, data: Data, max: int = 4096, timeout: float = 2) -> bytes:
    """Send one datagram and wait for one reply.

    Example: reply = udp.request("192.0.2.10", 123, b"\x1b")
    Task: send a UDP packet and read the reply
    Keywords: ntp, query, request response, probe, datagram reply
    Needs: network
    """

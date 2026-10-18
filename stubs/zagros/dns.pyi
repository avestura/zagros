"""dns: look up DNS records (A, AAAA, MX, TXT, CNAME, NS, PTR).

Type description, not runnable Starlark: do not `import` anything. The name `dns` is predeclared in main.star; call `dns.lookup(...)` directly.
Needs: network, scoped by the name being looked up (network:example.com covers its subdomains). A custom server= also needs network for that server (network:1.1.1.1).
Keywords: dns, nslookup, dig, resolve hostname, mx record, txt record, spf, dkim, nameserver, reverse lookup, ip of domain

Without server= the system resolver answers. With server= (host or host:port,
port 53) only that server is asked. Errors raise; a name with no such record
raises "no such record" (`attempt(dns.lookup, name, "MX")`). At most 256 records come back.
"""

from typing import Any, List, Optional

def lookup(name: str, type: str = "A", server: Optional[str] = None, timeout: float = 5) -> List[Any]:
    """Look up records of one type. A and AAAA give address strings, MX gives dicts with host and pref, TXT and NS and CNAME and PTR give strings. CNAME is empty when name is not an alias. PTR takes an IP address as name.

    Example: records = dns.lookup("example.com", "MX", server="1.1.1.1")
    Task: look up DNS records of a domain
    Keywords: resolve, nslookup, dig, mx, txt, spf, ns, cname, ptr, reverse dns, ip address of a domain, nameserver
    Needs: network
    """

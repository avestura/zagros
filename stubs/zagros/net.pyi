"""Networking module (generated from the tools registry; do not edit by hand).

CIDR and subnet calculators, IP and MAC tools, ports, HTTP status, MIME types, URLs and domains.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
"""

from typing import Dict, List

def cidr_info(cidr: str) -> Dict[str, str]:
    """Full subnet calculator for IPv4 and IPv6: network, broadcast (IPv4), netmask, wildcard, first/last host, usable hosts, total addresses (string, can be huge for IPv6), class, type (private, public, loopback, CGNAT...), binary and hex netmask and the reverse DNS zone. Accepts "10.0.0.5/24", "10.0.0.5 255.255.255.0", "10.0.0.5 0.0.0.255" (wildcard mask), IPv6 like "2001:db8::/32", or a bare IP (/32 or /128). /31 and /32 follow RFC 3021.

        cidr: CIDR or address with mask: "10.0.0.5/24", "10.0.0.5 255.255.255.0", "10.0.0.5 0.0.0.255" (wildcard mask), IPv6 like "2001:db8::/32", or a bare IP (/32 or /128).

    Example: net.cidr_info("192.168.1.130/26")
    Needs: none (pure function).
    """

def ip_formats(ip: str) -> Dict[str, str]:
    """Shows an IP address in every common notation: dotted, decimal integer, hex, octal, binary, IPv4-mapped IPv6, and the obfuscated URL forms (http://3232235777/, 0x..., octal) used in phishing and SSRF tests. IPv6 gives compressed, expanded, decimal, hex and binary forms. Also accepts decimal, hex (0x...) and short (127.1) inputs.

        ip: IPv4 or IPv6 address (also decimal, 0x hex or short forms).

    Example: net.ip_formats("192.168.1.1")
    Needs: none (pure function).
    """

def ip_to_int(ip: str) -> str:
    """Converts an IPv4 or IPv6 address to its decimal integer value, returned as a string because IPv6 values exceed 64 bits (192.168.1.1 gives 3232235777).

        ip: IPv4 or IPv6 address.

    Example: net.ip_to_int("192.168.1.1")
    Needs: none (pure function).
    """

def ip_type(ip: str) -> str:
    """Classifies an address: private (RFC 1918), cgnat (100.64/10), loopback, link-local, multicast, unspecified, documentation, benchmark, reserved, broadcast, 6to4, teredo, nat64, ula (IPv6 private), ipv4-mapped or public.

        ip: IPv4 or IPv6 address.

    Example: net.ip_type("100.64.0.1")
    Needs: none (pure function).
    """

def url_parse(url: str) -> Dict[str, str]:
    """Splits a URL into scheme, user, password, host, hostname, port, effective_port (the default port of the scheme when none is given), path, raw query, fragment, origin, host_type (ipv4/ipv6/domain) and file_name. Components keep their percent-encoding; absent parts are empty or missing.

        url: URL such as https://user:pw@example.com:8080/a/b?x=1#top. Input without a scheme is treated as host/path.

    Example: net.url_parse("https://user:pw@example.com:8443/a/b.html?x=1&y=2#top")
    Needs: none (pure function).
    """

def port_service(port_or_name: str) -> Dict[str, str]:
    """Looks up a TCP/UDP port number ("443", "443/tcp") or a service name ("ssh", "postgresql") in an embedded table of about 250 well-known and registered ports. Returns name, protocol, description and whether the port is well-known, registered or dynamic. The table is partial: unknown ports still report their range.

        port_or_name: Port number as a string ("22") or a service name ("ssh").

    Example: net.port_service("5432")
    Needs: none (pure function).
    """

def http_status(code: str) -> Dict[str, str]:
    """Explains an HTTP status code (pass it as a string, e.g. "404"): reason phrase, class (1xx-5xx) and a description. Covers all IANA-registered codes plus the common nginx/Cloudflare ones; unregistered codes still report their class.

        code: Status code such as "404" or "503".

    Example: net.http_status("404")
    Needs: none (pure function).
    """

def mac_normalize(mac: str, format: str = "colon", upper: bool = False) -> str:
    """Rewrites a MAC address in a chosen format: colon (aa:bb:cc:dd:ee:ff, default), dash, dot (Cisco aabb.ccdd.eeff), plain (no separators), upper (colon, upper-case) or lower. upper=true upper-cases any format. Accepts any separators.

        mac: MAC address in any common spelling (aa:bb:cc:dd:ee:ff, aa-bb-..., aabb.ccdd.eeff, aabbccddeeff).
        format: colon (default), dash, dot, plain, upper or lower.
        upper: Upper-case the result.

    Example: net.mac_normalize("00-1A-2B-3C-4D-5E", "dot")
    Needs: none (pure function).
    """

def domain_parts(host: str) -> Dict[str, str]:
    """Splits a host (or the host of a URL) using the public suffix list: public_suffix (e.g. co.uk), tld, registrable_domain (eTLD+1), domain label, subdomain, and whether the suffix is ICANN, private (e.g. github.io) or unknown. Unicode names are converted to punycode; IP addresses are rejected.

        host: Host name, domain or URL (the host part is used).

    Example: net.domain_parts("https://blog.example.co.uk/post")
    Needs: none (pure function).
    """

def cidr_contains(cidr: str, ip: str) -> bool:
    """Reports whether an IP address (or a whole CIDR block) lies inside the CIDR. Mixing IPv4 and IPv6 gives false.

        cidr: The containing network.
        ip: IP address or CIDR to test.

    Example: net.cidr_contains("10.0.0.0/8", "10.20.30.40")
    Needs: none (pure function).
    """

def cidr_overlap(a: str, b: str) -> bool:
    """Reports whether two CIDR blocks share any address (one contains the other, or they are equal).

        a: First CIDR.
        b: Second CIDR.

    Example: net.cidr_overlap("10.0.0.0/16", "10.0.5.0/24")
    Needs: none (pure function).
    """

def cidr_subnets(cidr: str, new_prefix: int, limit: int = 256) -> List[str]:
    """Splits a CIDR into all subnets of the longer prefix new_prefix (for example /16 into /24s). At most limit subnets are returned (default 256, max 65536); a larger split is truncated.

        cidr: Network to divide.
        new_prefix: Target prefix length, longer than the current one.
        limit: Maximum subnets to list (default 256, max 65536).

    Example: net.cidr_subnets("192.168.0.0/24", 26)
    Needs: none (pure function).
    """

def cidr_split(cidr: str, parts: int) -> List[str]:
    """Splits a CIDR into parts equal subnets. parts must be a power of two (2, 4, 8...) and at most 65536.

        cidr: Network to split.
        parts: Number of equal parts (power of two).

    Example: net.cidr_split("10.0.0.0/24", 4)
    Needs: none (pure function).
    """

def cidr_supernet(cidr: str, new_prefix: int) -> str:
    """Returns the larger network with the shorter prefix new_prefix that contains the given CIDR (for example 192.168.5.0/24 with /16 gives 192.168.0.0/16).

        cidr: Network to widen.
        new_prefix: Shorter target prefix length.

    Example: net.cidr_supernet("192.168.5.0/24", 16)
    Needs: none (pure function).
    """

def cidr_summarize(items: List[str]) -> List[str]:
    """Merges a list of CIDRs, IP addresses and ranges ("10.0.0.1-10.0.0.9") into the smallest set of CIDR blocks covering exactly the same addresses (route aggregation). Entries may also be comma or newline separated. IPv4 results come first, then IPv6.

        items: CIDRs, IPs or ranges to merge.

    Example: net.cidr_summarize(["10.0.0.0/25", "10.0.0.128/25", "10.0.1.0/24"])
    Needs: none (pure function).
    """

def cidr_exclude(cidr: str, exclude: str) -> List[str]:
    """Removes the exclude block from the CIDR and returns the remaining space as the minimal list of CIDRs. When they do not overlap the original network is returned; when exclude covers everything the list is empty.

        cidr: Network to cut from.
        exclude: Block to remove.

    Example: net.cidr_exclude("10.0.0.0/24", "10.0.0.64/26")
    Needs: none (pure function).
    """

def range_to_cidrs(start: str, end: str) -> List[str]:
    """Converts an inclusive IP range into the minimal list of CIDR blocks covering it exactly. start and end must be the same IP version; swapped bounds are corrected.

        start: First IP of the range.
        end: Last IP of the range.

    Example: net.range_to_cidrs("10.0.0.5", "10.0.0.20")
    Needs: none (pure function).
    """

def cidr_to_range(cidr: str) -> str:
    """Returns the first and last address of the CIDR as "first - last" (network and broadcast included).

        cidr: CIDR block.

    Example: net.cidr_to_range("192.168.1.0/24")
    Needs: none (pure function).
    """

def cidr_hosts(cidr: str, limit: int = 256) -> List[str]:
    """Lists usable host addresses of the CIDR (network and broadcast excluded for IPv4 up to /30). At most limit addresses are returned (default 256, max 65536).

        cidr: CIDR block.
        limit: Maximum addresses to list (default 256, max 65536).

    Example: net.cidr_hosts("192.168.1.0/29")
    Needs: none (pure function).
    """

def cidr_next(cidr: str) -> str:
    """Returns the next network of the same size directly after the given CIDR (10.0.0.0/24 gives 10.0.1.0/24). Fails at the end of the address space.

        cidr: CIDR block.

    Example: net.cidr_next("10.0.0.0/24")
    Needs: none (pure function).
    """

def cidr_count(cidr: str, usable: bool = False) -> str:
    """Returns the number of addresses in the CIDR as a decimal string (exact, even for huge IPv6 blocks). With usable true it returns the usable host count instead (IPv4 subtracts network and broadcast up to /30).

        cidr: CIDR block.
        usable: Count usable hosts instead of all addresses.

    Example: net.cidr_count("10.0.0.0/22")
    Needs: none (pure function).
    """

def mask_to_prefix(mask: str) -> int:
    """Converts a dotted IPv4 netmask (255.255.255.0) or wildcard mask (0.0.0.255) to its prefix length (24). Fails for non-contiguous masks.

        mask: Dotted netmask or wildcard mask.

    Example: net.mask_to_prefix("255.255.240.0")
    Needs: none (pure function).
    """

def wildcard_to_prefix(wildcard: str) -> int:
    """Converts a wildcard (inverse) mask such as 0.0.0.255 to its prefix length (24), as used in Cisco ACLs and OSPF network statements. Fails for non-contiguous masks.

        wildcard: Wildcard mask, e.g. 0.0.15.255.

    Example: net.wildcard_to_prefix("0.0.15.255")
    Needs: none (pure function).
    """

def prefix_to_mask(prefix: str, version: int = 4, wildcard: bool = False) -> str:
    """Converts a prefix length (24 or /24, given as a string) to a netmask. IPv4 by default (255.255.255.0); version 6 gives the IPv6 mask. With wildcard true it returns the inverse wildcard mask instead.

        prefix: Prefix length such as "24" or "/24".
        version: 4 (default) or 6.
        wildcard: Return the wildcard mask instead.

    Example: net.prefix_to_mask("20")
    Needs: none (pure function).
    """

def subnet_for_hosts(hosts: str, base: str = "") -> Dict[str, str]:
    """Finds the smallest IPv4 subnet that offers at least the needed number of usable hosts (network and broadcast excluded) and returns prefix, netmask and capacity. With base set it also gives the first such subnet inside the base network and how many fit.

        hosts: Needed usable hosts, as a string such as "50".
        base: Optional base network to allocate from, e.g. 10.0.0.0/16.

    Example: net.subnet_for_hosts("50")
    Needs: none (pure function).
    """

def vlsm(base: str, hosts: List[str]) -> List[str]:
    """Variable-length subnet masking: allocates subnets from an IPv4 base network for a list of host requirements, largest first, each aligned on its block size. Each requirement is a number or "name:number". Returns a table (header row first) with block, network, mask, usable range and broadcast; fails when the base is too small.

        base: IPv4 network to allocate from, e.g. 192.168.0.0/24.
        hosts: Host requirements such as ["LAN-A:100", "LAN-B:50", "20"].

    Example: net.vlsm("192.168.0.0/24", ["LAN-A:100", "LAN-B:50", "WAN:2"])
    Needs: none (pure function).
    """

def subnet_table(cidr: str) -> List[str]:
    """Quick-reference table for dividing a CIDR: for each longer prefix up to 8 bits deeper (or the maximum), lists the netmask, number of subnets, addresses per subnet and usable hosts per subnet.

        cidr: Network to tabulate, e.g. 10.0.0.0/16.

    Example: net.subnet_table("192.168.0.0/24")
    Needs: none (pure function).
    """

def int_to_ip(number: str, version: int = 0) -> str:
    """Converts a non-negative integer (decimal string, or 0x hex) to an IP address. version 4 or 6 forces the family; the default 0 picks IPv4 when the number fits in 32 bits and IPv6 otherwise.

        number: Integer as a string, e.g. "3232235777" or "0xC0A80101".
        version: 0 (auto), 4 or 6.

    Example: net.int_to_ip("3232235777")
    Needs: none (pure function).
    """

def ip_to_hex(ip: str) -> str:
    """Converts an IP address to hexadecimal with a 0x prefix (192.168.1.1 gives 0xc0a80101; IPv6 gives 32 hex digits).

        ip: IPv4 or IPv6 address.

    Example: net.ip_to_hex("192.168.1.1")
    Needs: none (pure function).
    """

def ip_to_binary(ip: str) -> str:
    """Writes an IP address in binary: IPv4 as four dotted 8-bit groups, IPv6 as eight colon-separated 16-bit groups.

        ip: IPv4 or IPv6 address.

    Example: net.ip_to_binary("192.168.1.1")
    Needs: none (pure function).
    """

def binary_to_ip(bits: str) -> str:
    """Converts 32 binary digits (IPv4) or 128 binary digits (IPv6) to an address. Dots, colons and spaces between groups are ignored.

        bits: Binary digits, e.g. 11000000.10101000.00000001.00000001.

    Example: net.binary_to_ip("11000000.10101000.00000001.00000001")
    Needs: none (pure function).
    """

def hex_to_ip(hex: str) -> str:
    """Converts 8 hex digits (IPv4) or 32 hex digits (IPv6) to an address. A 0x prefix and separators (. : space) are ignored.

        hex: Hex digits such as C0A80101 or 0xc0a80101.

    Example: net.hex_to_ip("0xC0A80101")
    Needs: none (pure function).
    """

def ip_version(ip: str) -> int:
    """Returns 4 for an IPv4 address and 6 for an IPv6 address; fails for anything else. IPv4-mapped addresses (::ffff:1.2.3.4) report 6.

        ip: Address to inspect.

    Example: net.ip_version("2001:db8::1")
    Needs: none (pure function).
    """

def ip_classify(ip: str) -> Dict[str, str]:
    """Detailed classification of an address as a dict: type, description with RFC, scope, and is_global / is_private / is_loopback / is_multicast / is_link_local / is_unspecified flags (as "true"/"false"). Multicast addresses include their scope; IPv4-mapped ones include the embedded IPv4.

        ip: IPv4 or IPv6 address.

    Example: net.ip_classify("fe80::1")
    Needs: none (pure function).
    """

def ip_normalize(ip: str) -> str:
    """Returns the canonical text form of an address: IPv6 is lower-case and compressed (RFC 5952), IPv4 octets lose leading zeros. Also accepts [brackets], zones, decimal/hex integers and short forms like 127.1.

        ip: Address in any common spelling.

    Example: net.ip_normalize("2001:0DB8:0000:0000:0000:0000:0000:0001")
    Needs: none (pure function).
    """

def ipv6_expand(ip: str) -> str:
    """Expands an IPv6 address to the full eight 4-digit groups (2001:db8::1 becomes 2001:0db8:0000:0000:0000:0000:0000:0001). IPv4 input is returned as its IPv4-mapped expanded form.

        ip: IPv6 address.

    Example: net.ipv6_expand("2001:db8::1")
    Needs: none (pure function).
    """

def ipv6_compress(ip: str) -> str:
    """Compresses an IPv6 address to the shortest RFC 5952 form: leading zeros dropped and the longest run of zero groups replaced by ::. IPv4-mapped addresses keep their dotted tail.

        ip: IPv6 address.

    Example: net.ipv6_compress("2001:0db8:0000:0000:0000:0000:0000:0001")
    Needs: none (pure function).
    """

def ipv4_to_ipv6_mapped(ip: str) -> str:
    """Maps an IPv4 address into IPv6 as ::ffff:a.b.c.d (RFC 4291).

        ip: IPv4 address.

    Example: net.ipv4_to_ipv6_mapped("192.0.2.1")
    Needs: none (pure function).
    """

def ipv6_to_ipv4_mapped(ip: str) -> str:
    """Extracts the IPv4 address from an IPv4-mapped IPv6 address (::ffff:192.0.2.1 gives 192.0.2.1). Fails for IPv6 addresses that are not IPv4-mapped.

        ip: IPv4-mapped IPv6 address.

    Example: net.ipv6_to_ipv4_mapped("::ffff:c000:201")
    Needs: none (pure function).
    """

def ipv4_to_6to4(ip: str) -> str:
    """Derives the 6to4 IPv6 prefix (RFC 3056) of an IPv4 address: 192.0.2.4 gives 2002:c000:204:: (the /48 prefix).

        ip: IPv4 address.

    Example: net.ipv4_to_6to4("192.0.2.4")
    Needs: none (pure function).
    """

def sixtofour_to_ipv4(ip: str) -> str:
    """Extracts the embedded IPv4 address from a 6to4 address (2002:c000:204::1 gives 192.0.2.4). Fails when the address is not inside 2002::/16.

        ip: 6to4 IPv6 address.

    Example: net.sixtofour_to_ipv4("2002:c000:204::1")
    Needs: none (pure function).
    """

def nat64(ip: str, prefix: str = "64:ff9b::/96") -> str:
    """Synthesizes the NAT64 IPv6 address for an IPv4 address (RFC 6052). The default prefix is the well-known 64:ff9b::/96; /32, /40, /48, /56, /64 and /96 prefixes are supported.

        ip: IPv4 address.
        prefix: NAT64 prefix with length, default 64:ff9b::/96.

    Example: net.nat64("192.0.2.33")
    Needs: none (pure function).
    """

def nat64_to_ipv4(ip: str, prefix: str = "64:ff9b::/96") -> str:
    """Extracts the IPv4 address embedded in a NAT64 IPv6 address (RFC 6052). The default prefix is 64:ff9b::/96; the address must lie inside the prefix.

        ip: NAT64 IPv6 address.
        prefix: NAT64 prefix with length, default 64:ff9b::/96.

    Example: net.nat64_to_ipv4("64:ff9b::c000:221")
    Needs: none (pure function).
    """

def eui64_from_mac(mac: str, prefix: str = "fe80::") -> str:
    """Builds the modified EUI-64 IPv6 address (RFC 4291) for a MAC address: inserts ff:fe in the middle and flips the universal/local bit. prefix is the 64-bit network prefix, default fe80:: (link-local).

        mac: MAC address in any common format.
        prefix: IPv6 /64 prefix, e.g. 2001:db8:1:2::

    Example: net.eui64_from_mac("00:1a:2b:3c:4d:5e", "2001:db8::")
    Needs: none (pure function).
    """

def link_local_from_mac(mac: str) -> str:
    """Returns the SLAAC link-local IPv6 address (fe80::/64 plus modified EUI-64) for a MAC address.

        mac: MAC address in any common format.

    Example: net.link_local_from_mac("00:1a:2b:3c:4d:5e")
    Needs: none (pure function).
    """

def ptr_name(ip: str) -> str:
    """Returns the reverse-DNS name for an address: 1.2.3.4 gives 4.3.2.1.in-addr.arpa, IPv6 gives the 32-nibble ip6.arpa name. No lookup is performed.

        ip: IPv4 or IPv6 address.

    Example: net.ptr_name("192.0.2.10")
    Needs: none (pure function).
    """

def ip_compare(a: str, b: str) -> int:
    """Compares two addresses numerically: -1 if a < b, 0 if equal, 1 if a > b. IPv4 sorts before IPv6.

        a: First address.
        b: Second address.

    Example: net.ip_compare("10.0.0.9", "10.0.0.10")
    Needs: none (pure function).
    """

def sort_ips(ips: List[str], descending: bool = False, unique: bool = False) -> List[str]:
    """Sorts addresses numerically (10.0.0.9 before 10.0.0.10, unlike text sorting), IPv4 first. Entries may also be comma or newline separated. descending reverses the order and unique drops duplicates. Invalid entries are an error.

        ips: Addresses to sort.
        descending: Sort high to low.
        unique: Remove duplicates.

    Example: net.sort_ips(["10.0.0.10", "10.0.0.9", "::1"])
    Needs: none (pure function).
    """

def ip_in_ranges(ip: str, ranges: List[str]) -> bool:
    """Reports whether the address lies in at least one of the given CIDRs, single IPs or ranges ("10.0.0.1-10.0.0.50").

        ip: Address to test.
        ranges: CIDRs, IPs or a-b ranges.

    Example: net.ip_in_ranges("10.1.2.3", ["192.168.0.0/16", "10.0.0.0/8"])
    Needs: none (pure function).
    """

def ip_matching_ranges(ip: str, ranges: List[str]) -> List[str]:
    """Returns the entries (CIDRs, IPs or ranges) from the list that contain the address; empty when none do.

        ip: Address to test.
        ranges: CIDRs, IPs or a-b ranges.

    Example: net.ip_matching_ranges("10.1.2.3", ["10.0.0.0/8", "10.1.0.0/16", "172.16.0.0/12"])
    Needs: none (pure function).
    """

def ip_add(ip: str, n: int) -> str:
    """Adds a signed integer to an address (10.0.0.255 + 1 = 10.0.1.0). Fails when the result leaves the address family's range.

        ip: Starting address.
        n: Offset, may be negative.

    Example: net.ip_add("10.0.0.255", 1)
    Needs: none (pure function).
    """

def ip_diff(a: str, b: str) -> str:
    """Returns b minus a as a signed decimal string: the number of addresses from a to b (positive when b is higher). Both must be the same IP version.

        a: Start address.
        b: End address.

    Example: net.ip_diff("10.0.0.1", "10.0.1.1")
    Needs: none (pure function).
    """

def is_valid_ip(ip: str) -> bool:
    """Reports whether the text is a valid IPv4 or IPv6 address (surrounding whitespace, brackets and leading zeros in IPv4 octets are tolerated).

        ip: Text to check.

    Example: net.is_valid_ip("256.1.1.1")
    Needs: none (pure function).
    """

def is_valid_ipv4(ip: str) -> bool:
    """Reports whether the text is a valid dotted-quad IPv4 address.

        ip: Text to check.

    Example: net.is_valid_ipv4("192.168.0.1")
    Needs: none (pure function).
    """

def is_valid_ipv6(ip: str) -> bool:
    """Reports whether the text is a valid IPv6 address (including :: compression and IPv4-mapped forms).

        ip: Text to check.

    Example: net.is_valid_ipv6("2001:db8::1")
    Needs: none (pure function).
    """

def is_valid_cidr(cidr: str) -> bool:
    """Reports whether the text is valid CIDR notation (address/prefix, IPv4 or IPv6). Host bits may be set. Unlike cidr_info it does not accept netmask spellings or a bare address.

        cidr: Text to check.

    Example: net.is_valid_cidr("10.0.0.0/33")
    Needs: none (pure function).
    """

def ipv4_octets(ip: str) -> List[str]:
    """Splits an IPv4 address into its four octets as decimal strings.

        ip: IPv4 address.

    Example: net.ipv4_octets("192.168.1.1")
    Needs: none (pure function).
    """

def mac_is_valid(mac: str) -> bool:
    """Reports whether the text is a valid 48-bit MAC address: exactly 12 hex digits with optional : - . or space separators.

        mac: MAC address in any common spelling (aa:bb:cc:dd:ee:ff, aa-bb-..., aabb.ccdd.eeff, aabbccddeeff).

    Example: net.mac_is_valid("00:1A:2B:3C:4D")
    Needs: none (pure function).
    """

def mac_info(mac: str) -> Dict[str, str]:
    """Describes a MAC address: unicast/multicast/broadcast, universally or locally administered bit, OUI and NIC parts, the EUI-64 form, the SLAAC link-local IPv6 address and, for well-known prefixes, the vendor (the OUI table is partial).

        mac: MAC address in any common spelling (aa:bb:cc:dd:ee:ff, aa-bb-..., aabb.ccdd.eeff, aabbccddeeff).

    Example: net.mac_info("00:50:56:aa:bb:cc")
    Needs: none (pure function).
    """

def oui_vendor(mac: str) -> str:
    """Looks up the manufacturer of a MAC address (or an OUI like 00:50:56) in a small embedded table of about 200 well-known prefixes. The table is intentionally partial: an empty string means unknown, not unassigned. Also recognizes broadcast, multicast and Docker addresses.

        mac: MAC address or its first three bytes.

    Example: net.oui_vendor("b8:27:eb:12:34:56")
    Needs: none (pure function).
    """

def service_ports(name: str) -> List[str]:
    """Lists every port/protocol of a service name ("http" gives 80/tcp, "domain" gives 53/tcp and 53/udp). Matches exact service names first and then service descriptions (e.g. "mail", "database"); an unknown name gives an empty list.

        name: Service name or keyword such as ssh, dns, redis, mail.

    Example: net.service_ports("dns")
    Needs: none (pure function).
    """

def port_range_type(port: str) -> str:
    """Classifies a port number as reserved (0), well-known (1-1023), registered (1024-49151) or dynamic/ephemeral (49152-65535) per IANA. Fails outside 0-65535.

        port: Port number as a string, e.g. "8080".

    Example: net.port_range_type("8080")
    Needs: none (pure function).
    """

def protocol_number(name_or_number: str) -> Dict[str, str]:
    """Looks up an IANA IP protocol number (TCP 6, UDP 17, ICMP 1, GRE 47, ESP 50, OSPF 89...) by number or by name, returning number, keyword and full name. The embedded table covers the commonly seen protocols.

        name_or_number: Protocol keyword ("tcp", "ospf") or number ("47").

    Example: net.protocol_number("gre")
    Needs: none (pure function).
    """

def icmp_type(spec: str, code: int = -1, version: int = 4) -> Dict[str, str]:
    """Names an ICMP message type and code: "3/1" is Destination Unreachable, Host Unreachable. spec is the type, optionally followed by /code; code can also be passed separately (default -1 means none). version 6 uses ICMPv6 (e.g. 135 Neighbor Solicitation). Also lists all codes of the type.

        spec: Type or "type/code", e.g. "8" or "3/4".
        code: Code when not part of spec; -1 for none.
        version: 4 (ICMP, default) or 6 (ICMPv6).

    Example: net.icmp_type("3/4")
    Needs: none (pure function).
    """

def http_status_search(text: str) -> List[str]:
    """Finds status codes by words in their name or description ("redirect", "timeout", "auth"), or by class prefix ("4", "40"). Returns lines like "404 Not Found".

        text: Search words or a code prefix.

    Example: net.http_status_search("redirect")
    Needs: none (pure function).
    """

def http_method_info(method: str) -> Dict[str, str]:
    """Describes an HTTP method (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, TRACE, CONNECT): whether it is safe, idempotent and cacheable, whether it normally carries a body, and what it is for. Case-insensitive.

        method: HTTP method name.

    Example: net.http_method_info("put")
    Needs: none (pure function).
    """

def mime_type(extension_or_filename: str) -> str:
    """Returns the MIME type for a file extension or filename ("png", ".json", "report.pdf") from an embedded table of about 200 common types. Unknown extensions give application/octet-stream. Case-insensitive; for "archive.tar.gz" the last extension counts.

        extension_or_filename: Extension with or without dot, or a file name/path.

    Example: net.mime_type("photo.jpeg")
    Needs: none (pure function).
    """

def extension_for_mime(mime_type: str) -> List[str]:
    """Lists the file extensions (with dot, most common first) for a MIME type such as image/png; parameters like charset are ignored. Unknown types give an empty list.

        mime_type: MIME type, e.g. application/json or text/html; charset=utf-8.

    Example: net.extension_for_mime("image/jpeg")
    Needs: none (pure function).
    """

def content_type_parse(header: str) -> Dict[str, str]:
    """Parses a Content-Type header value such as "text/html; charset=UTF-8" into a dict with type, subtype, full media_type, suffix (+json), charset, boundary and every parameter as param_<name>. Fails for values without a valid media type.

        header: Header value, with or without the "Content-Type:" prefix.

    Example: net.content_type_parse("multipart/form-data; boundary=xyz")
    Needs: none (pure function).
    """

def url_build(parts: Dict[str, str], scheme: str = "https") -> str:
    """Builds a URL from parts. Recognized keys of parts: scheme, user, password, host (or hostname), port, path, query (raw, without ?) and fragment. IPv6 hosts are bracketed automatically and the path is percent-encoded as needed.

        parts: Dict with keys scheme, host, port, path, query, fragment, user, password.
        scheme: Default scheme used when parts has none.

    Example: net.url_build({"scheme": "https", "host": "example.com", "path": "/a b", "query": "x=1"})
    Needs: none (pure function).
    """

def url_join(base: str, ref: str) -> str:
    """Resolves a relative reference against a base URL following RFC 3986, like a browser does for links (../img/a.png against https://x.com/a/b/c.html gives https://x.com/a/img/a.png). An absolute ref wins over the base.

        base: Base URL.
        ref: Relative or absolute reference.

    Example: net.url_join("https://example.com/a/b/c.html", "../img/x.png")
    Needs: none (pure function).
    """

def url_normalize(url: str, sort_query: bool = False, strip_fragment: bool = False) -> str:
    """Normalizes a URL for comparison: lower-cases scheme and host, removes the default port and a trailing host dot, resolves ./ and ../ segments, upper-cases percent escapes, decodes unreserved characters, adds / for an empty path and drops an empty fragment. sort_query orders query parameters; strip_fragment removes #fragment.

        url: URL such as https://user:pw@example.com:8080/a/b?x=1#top. Input without a scheme is treated as host/path.
        sort_query: Sort query parameters alphabetically.
        strip_fragment: Remove the #fragment.

    Example: net.url_normalize("HTTP://Example.COM:80/a/./b/../c?b=2&a=1", True)
    Needs: none (pure function).
    """

def url_query(url: str) -> Dict[str, str]:
    """Returns the query parameters of a URL (or of a bare query string such as a=1&b=2) as a dict with decoded names and values. For repeated keys the values are joined with commas; a key without = has an empty value.

        url: URL or bare query string.

    Example: net.url_query("https://example.com/s?q=go+lang&tag=a&tag=b")
    Needs: none (pure function).
    """

def url_set_query(url: str, params: Dict[str, str]) -> str:
    """Replaces the whole query string of a URL with the given parameters (sorted by name, percent-encoded). An empty dict removes the query.

        url: URL to modify.
        params: New query parameters.

    Example: net.url_set_query("https://example.com/s?old=1", {"q": "go lang", "page": "2"})
    Needs: none (pure function).
    """

def url_add_param(url: str, key: str, value: str) -> str:
    """Appends a query parameter to a URL, keeping existing parameters (a repeated key is added again). Use url_set_param to replace instead.

        url: URL to modify.
        key: Parameter name.
        value: Parameter value (encoded for you).

    Example: net.url_add_param("https://example.com/s?q=1", "page", "2")
    Needs: none (pure function).
    """

def url_set_param(url: str, key: str, value: str) -> str:
    """Sets a query parameter: replaces the first occurrence in place, removes any further occurrences, or appends it when absent.

        url: URL to modify.
        key: Parameter name.
        value: Parameter value (encoded for you).

    Example: net.url_set_param("https://example.com/s?q=1&page=1", "page", "5")
    Needs: none (pure function).
    """

def url_remove_param(url: str, keys: List[str]) -> str:
    """Removes the named query parameters from a URL. A name ending in * is a prefix match (utm_* removes utm_source, utm_medium...). Other parameters keep their original encoding and order.

        url: URL to modify.
        keys: Parameter names to remove, e.g. ["utm_*", "ref"].

    Example: net.url_remove_param("https://example.com/?a=1&b=2", ["b"])
    Needs: none (pure function).
    """

def url_strip_tracking(url: str) -> str:
    """Removes well-known tracking parameters from a URL: utm_*, fbclid, gclid, dclid, msclkid, yclid, mc_cid, mc_eid, igshid, twclid, ttclid, _hsenc, _ga, ref_src, spm, mkt_tok and more. Other parameters and the fragment are kept.

        url: URL to clean.

    Example: net.url_strip_tracking("https://example.com/p?id=7&utm_source=x&fbclid=abc")
    Needs: none (pure function).
    """

def url_origin(url: str) -> str:
    """Returns the web origin (scheme://host[:port]) of a URL, lower-cased, with the port omitted when it is the scheme's default. Fails for URLs without scheme or host.

        url: URL such as https://user:pw@example.com:8080/a/b?x=1#top. Input without a scheme is treated as host/path.

    Example: net.url_origin("https://Example.com:443/a?b=1")
    Needs: none (pure function).
    """

def url_is_valid(url: str) -> bool:
    """Reports whether the text is an absolute URL: a scheme, then for http/https/ftp/ws/wss a valid host (domain, IPv4 or [IPv6]) and an optional valid port. Other schemes (mailto:, file:, data:) only need to parse. Whitespace makes it invalid.

        url: Text to check.

    Example: net.url_is_valid("https://example.com/path")
    Needs: none (pure function).
    """

def is_valid_hostname(host: str, allow_underscore: bool = False) -> bool:
    """Checks RFC 1123 hostname syntax: labels of 1-63 letters, digits and hyphens (no leading or trailing hyphen), at most 253 characters in total. A single label such as localhost is fine. allow_underscore also accepts _ (service names, DKIM).

        host: Hostname to check.
        allow_underscore: Accept underscores in labels.

    Example: net.is_valid_hostname("my-server.local")
    Needs: none (pure function).
    """

def is_valid_domain(host: str) -> bool:
    """Checks that the text is a plausible fully qualified domain name: valid hostname syntax, at least two labels, an alphabetic (or xn--) top-level label and not an IP address. It does not check that the TLD exists.

        host: Domain name to check.

    Example: net.is_valid_domain("example.com")
    Needs: none (pure function).
    """

def email_parse(email: str) -> Dict[str, str]:
    """Splits an email address into local, domain, base_local, +tag, registrable_domain and a normalized form (lower-case; Gmail/Googlemail addresses also lose dots and the +tag, googlemail.com becomes gmail.com). Accepts "Name <a@b.c>" and mailto: prefixes. Fails for invalid addresses.

        email: Email address.

    Example: net.email_parse("John.Doe+news@Gmail.com")
    Needs: none (pure function).
    """

def is_valid_email(email: str) -> bool:
    """Pragmatic RFC 5322 check: a dot-atom or quoted local part (max 64 chars), an @, and a hostname domain with at least two labels or an [IP literal]; total length up to 254. No display names, comments or whitespace. It does not verify the mailbox exists.

        email: Email address.

    Example: net.is_valid_email("user@example.com")
    Needs: none (pure function).
    """

def reverse_domain_name(host: str) -> str:
    """Reverses the labels of a domain: a.b.com becomes com.b.a (Java package / Android application id style). A URL is reduced to its host first.

        host: Domain name.

    Example: net.reverse_domain_name("mail.example.com")
    Needs: none (pure function).
    """

def basic_auth_header(user: str, password: str) -> str:
    """Builds an HTTP Basic Authorization header value: "Basic " plus base64(user:password). The user name must not contain a colon.

        user: User name (no colon).
        password: Password.

    Example: net.basic_auth_header("alice", "s3cret")
    Needs: none (pure function).
    """

def parse_basic_auth(header: str) -> Dict[str, str]:
    """Decodes an HTTP Basic Authorization header (with or without the "Basic " prefix and "Authorization:" name) into user and password. Fails for invalid base64 or a missing colon.

        header: Header value such as Basic YWxpY2U6czNjcmV0.

    Example: net.parse_basic_auth("Basic YWxpY2U6czNjcmV0")
    Needs: none (pure function).
    """

def parse_host_port(address: str) -> Dict[str, str]:
    """Splits host:port, including [IPv6]:port, bare IPv6 (no port), IPv4:port and name:port. Returns host, port (empty when absent), host_type (ipv4, ipv6, domain) and the bracketed IPv6 form. Fails for invalid hosts or ports above 65535.

        address: Address such as example.com:8080 or [::1]:443.

    Example: net.parse_host_port("[::1]:8080")
    Needs: none (pure function).
    """

def hosts_entry(ip: str, names: List[str]) -> str:
    """Formats a hosts-file line (ip, tab, space-separated names) after validating the address and host names, e.g. "127.0.0.1\tapp.local www.app.local".

        ip: IP address.
        names: Host names to map to the address.

    Example: net.hosts_entry("127.0.0.1", ["app.local", "www.app.local"])
    Needs: none (pure function).
    """

def bandwidth(value: float, from_unit: str, to_unit: str) -> float:
    """Converts a data rate between units: bps, Kbps, Mbps, Gbps, Tbps (bits, decimal), B/s, KB/s, MB/s, GB/s (bytes, decimal) and KiB/s, MiB/s, GiB/s (bytes, binary). A capital B means bytes, lower-case b means bits; /s or ps both work.

        value: Rate in from_unit.
        from_unit: Unit such as "Mbps" or "MB/s".
        to_unit: Target unit such as "MiB/s".

    Example: net.bandwidth(100, "Mbps", "MB/s")
    Needs: none (pure function).
    """

def transfer_time(size: str, bandwidth: str) -> str:
    """How long a transfer takes: size such as "1.5 GB" or "700 MiB" at a bandwidth such as "100 Mbps" or "12 MB/s". Returns a human duration like "2m 4s". Protocol overhead is ignored (ideal line rate).

        size: Amount of data, e.g. "4.7 GB".
        bandwidth: Rate, e.g. "50 Mbps".

    Example: net.transfer_time("4.7 GB", "100 Mbps")
    Needs: none (pure function).
    """

def throughput_needed(size: str, seconds: float) -> Dict[str, str]:
    """Bandwidth required to move a size (e.g. "10 GB") in the given number of seconds. Returns the rate in bps, Kbps, Mbps, Gbps, MB/s and MiB/s as strings.

        size: Amount of data, e.g. "10 GB".
        seconds: Time allowed in seconds (> 0).

    Example: net.throughput_needed("10 GB", 600)
    Needs: none (pure function).
    """

# Networking (`net`)

CIDR and subnet calculators, IP and MAC tools, ports, HTTP status, MIME types, URLs and domains.

No permission is needed: these are pure functions. Generated from the tools registry.

92 functions. Use them as `net.<name>(...)` in any extension.

## net.cidr_info

`net.cidr_info(cidr: str) -> Dict[str, str]`

Full subnet calculator for IPv4 and IPv6: network, broadcast (IPv4), netmask, wildcard, first/last host, usable hosts, total addresses (string, can be huge for IPv6), class, type (private, public, loopback, CGNAT...), binary and hex netmask and the reverse DNS zone. Accepts "10.0.0.5/24", "10.0.0.5 255.255.255.0", "10.0.0.5 0.0.0.255" (wildcard mask), IPv6 like "2001:db8::/32", or a bare IP (/32 or /128). /31 and /32 follow RFC 3021.

| Parameter | Type | Description |
|---|---|---|
| `cidr` | `str` | CIDR or address with mask: "10.0.0.5/24", "10.0.0.5 255.255.255.0", "10.0.0.5 0.0.0.255" (wildcard mask), IPv6 like "2001:db8::/32", or a bare IP (/32 or /128). |

Returns `Dict[str, str]`. Example: `net.cidr_info("192.168.1.130/26")`

Keywords: subnet calculator, netmask, network address, broadcast, usable hosts, wildcard, hosts, ipcalc, cidr notation, prefix

## net.ip_formats

`net.ip_formats(ip: str) -> Dict[str, str]`

Shows an IP address in every common notation: dotted, decimal integer, hex, octal, binary, IPv4-mapped IPv6, and the obfuscated URL forms (http://3232235777/, 0x..., octal) used in phishing and SSRF tests. IPv6 gives compressed, expanded, decimal, hex and binary forms. Also accepts decimal, hex (0x...) and short (127.1) inputs.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IPv4 or IPv6 address (also decimal, 0x hex or short forms). |

Returns `Dict[str, str]`. Example: `net.ip_formats("192.168.1.1")`

Keywords: convert ip, decimal ip, hex ip, octal ip, binary ip, obfuscated url, ip notation, ssrf

## net.ip_to_int

`net.ip_to_int(ip: str) -> str`

Converts an IPv4 or IPv6 address to its decimal integer value, returned as a string because IPv6 values exceed 64 bits (192.168.1.1 gives 3232235777).

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IPv4 or IPv6 address. |

Returns `str`. Example: `net.ip_to_int("192.168.1.1")`

Keywords: ip to decimal, ip number, long, integer, numeric ip, inet_aton, bigint

## net.ip_type

`net.ip_type(ip: str) -> str`

Classifies an address: private (RFC 1918), cgnat (100.64/10), loopback, link-local, multicast, unspecified, documentation, benchmark, reserved, broadcast, 6to4, teredo, nat64, ula (IPv6 private), ipv4-mapped or public.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IPv4 or IPv6 address. |

Returns `str`. Example: `net.ip_type("100.64.0.1")`

Keywords: is private ip, public or private, loopback, link local, is private, is public, rfc1918, cgnat, link-local, multicast, ula, documentation, benchmark, classify ip, internal ip

## net.url_parse

`net.url_parse(url: str) -> Dict[str, str]`

Splits a URL into scheme, user, password, host, hostname, port, effective_port (the default port of the scheme when none is given), path, raw query, fragment, origin, host_type (ipv4/ipv6/domain) and file_name. Components keep their percent-encoding; absent parts are empty or missing.

| Parameter | Type | Description |
|---|---|---|
| `url` | `str` | URL such as https://user:pw@example.com:8080/a/b?x=1#top. Input without a scheme is treated as host/path. |

Returns `Dict[str, str]`. Example: `net.url_parse("https://user:pw@example.com:8443/a/b.html?x=1&y=2#top")`

Keywords: scheme host path query, break down url, parse link, uri, split url, components, query string, hostname, port, origin, scheme, fragment, dissect

## net.port_service

`net.port_service(port_or_name: str) -> Dict[str, str]`

Looks up a TCP/UDP port number ("443", "443/tcp") or a service name ("ssh", "postgresql") in an embedded table of about 250 well-known and registered ports. Returns name, protocol, description and whether the port is well-known, registered or dynamic. The table is partial: unknown ports still report their range.

| Parameter | Type | Description |
|---|---|---|
| `port_or_name` | `str` | Port number as a string ("22") or a service name ("ssh"). |

Returns `Dict[str, str]`. Example: `net.port_service("5432")`

Keywords: what runs on port 443, well known ports, port number, what runs on port, iana, well-known ports, service name, tcp, udp, etc/services

## net.http_status

`net.http_status(code: str) -> Dict[str, str]`

Explains an HTTP status code (pass it as a string, e.g. "404"): reason phrase, class (1xx-5xx) and a description. Covers all IANA-registered codes plus the common nginx/Cloudflare ones; unregistered codes still report their class.

| Parameter | Type | Description |
|---|---|---|
| `code` | `str` | Status code such as "404" or "503". |

Returns `Dict[str, str]`. Example: `net.http_status("404")`

Keywords: 404 meaning, status code description, what does 500 mean, status code, http error, 404, 500, 301, reason phrase, response code, rest

## net.mac_normalize

`net.mac_normalize(mac: str, format: str = "colon", upper: bool = False) -> str`

Rewrites a MAC address in a chosen format: colon (aa:bb:cc:dd:ee:ff, default), dash, dot (Cisco aabb.ccdd.eeff), plain (no separators), upper (colon, upper-case) or lower. upper=true upper-cases any format. Accepts any separators.

| Parameter | Type | Description |
|---|---|---|
| `mac` | `str` | MAC address in any common spelling (aa:bb:cc:dd:ee:ff, aa-bb-..., aabb.ccdd.eeff, aabbccddeeff). |
| `format` | `str` (default `"colon"`) | colon (default), dash, dot, plain, upper or lower. |
| `upper` | `bool` (default `False`) | Upper-case the result. |

Returns `str`. Example: `net.mac_normalize("00-1A-2B-3C-4D-5E", "dot")`

Keywords: mac address format, colon or dash mac, mac format, convert mac, cisco format, hyphen, colon, hardware address, ethernet address

## net.domain_parts

`net.domain_parts(host: str) -> Dict[str, str]`

Splits a host (or the host of a URL) using the public suffix list: public_suffix (e.g. co.uk), tld, registrable_domain (eTLD+1), domain label, subdomain, and whether the suffix is ICANN, private (e.g. github.io) or unknown. Unicode names are converted to punycode; IP addresses are rejected.

| Parameter | Type | Description |
|---|---|---|
| `host` | `str` | Host name, domain or URL (the host part is used). |

Returns `Dict[str, str]`. Example: `net.domain_parts("https://blog.example.co.uk/post")`

Keywords: subdomain, registered domain, tld, public suffix, etld+1, registrable domain, second level domain, sld, apex domain, root domain

## net.cidr_contains

`net.cidr_contains(cidr: str, ip: str) -> bool`

Reports whether an IP address (or a whole CIDR block) lies inside the CIDR. Mixing IPv4 and IPv6 gives false.

| Parameter | Type | Description |
|---|---|---|
| `cidr` | `str` | The containing network. |
| `ip` | `str` | IP address or CIDR to test. |

Returns `bool`. Example: `net.cidr_contains("10.0.0.0/8", "10.20.30.40")`

Keywords: ip in range, is ip in cidr, in subnet, belongs, member, within, is ip in network

## net.cidr_overlap

`net.cidr_overlap(a: str, b: str) -> bool`

Reports whether two CIDR blocks share any address (one contains the other, or they are equal).

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First CIDR. |
| `b` | `str` | Second CIDR. |

Returns `bool`. Example: `net.cidr_overlap("10.0.0.0/16", "10.0.5.0/24")`

Keywords: intersect, conflict, collision, overlapping subnets

## net.cidr_subnets

`net.cidr_subnets(cidr: str, new_prefix: int, limit: int = 256) -> List[str]`

Splits a CIDR into all subnets of the longer prefix new_prefix (for example /16 into /24s). At most limit subnets are returned (default 256, max 65536); a larger split is truncated.

| Parameter | Type | Description |
|---|---|---|
| `cidr` | `str` | Network to divide. |
| `new_prefix` | `int` | Target prefix length, longer than the current one. |
| `limit` | `int` (default `256`) | Maximum subnets to list (default 256, max 65536). |

Returns `List[str]`. Example: `net.cidr_subnets("192.168.0.0/24", 26)`

Keywords: subnetting, divide network, split, divide, carve, child networks

## net.cidr_split

`net.cidr_split(cidr: str, parts: int) -> List[str]`

Splits a CIDR into parts equal subnets. parts must be a power of two (2, 4, 8...) and at most 65536.

| Parameter | Type | Description |
|---|---|---|
| `cidr` | `str` | Network to split. |
| `parts` | `int` | Number of equal parts (power of two). |

Returns `List[str]`. Example: `net.cidr_split("10.0.0.0/24", 4)`

Keywords: equal subnets, divide, halve, partition

## net.cidr_supernet

`net.cidr_supernet(cidr: str, new_prefix: int) -> str`

Returns the larger network with the shorter prefix new_prefix that contains the given CIDR (for example 192.168.5.0/24 with /16 gives 192.168.0.0/16).

| Parameter | Type | Description |
|---|---|---|
| `cidr` | `str` | Network to widen. |
| `new_prefix` | `int` | Shorter target prefix length. |

Returns `str`. Example: `net.cidr_supernet("192.168.5.0/24", 16)`

Keywords: parent network, aggregate, widen, summarize one

## net.cidr_summarize

`net.cidr_summarize(items: List[str]) -> List[str]`

Merges a list of CIDRs, IP addresses and ranges ("10.0.0.1-10.0.0.9") into the smallest set of CIDR blocks covering exactly the same addresses (route aggregation). Entries may also be comma or newline separated. IPv4 results come first, then IPv6.

| Parameter | Type | Description |
|---|---|---|
| `items` | `List[str]` | CIDRs, IPs or ranges to merge. |

Returns `List[str]`. Example: `net.cidr_summarize(["10.0.0.0/25", "10.0.0.128/25", "10.0.1.0/24"])`

Keywords: merge cidr ranges, aggregate networks, supernet, aggregate, merge, route summarization, collapse, minimal

## net.cidr_exclude

`net.cidr_exclude(cidr: str, exclude: str) -> List[str]`

Removes the exclude block from the CIDR and returns the remaining space as the minimal list of CIDRs. When they do not overlap the original network is returned; when exclude covers everything the list is empty.

| Parameter | Type | Description |
|---|---|---|
| `cidr` | `str` | Network to cut from. |
| `exclude` | `str` | Block to remove. |

Returns `List[str]`. Example: `net.cidr_exclude("10.0.0.0/24", "10.0.0.64/26")`

Keywords: subtract, remove, carve out, difference, minus

## net.range_to_cidrs

`net.range_to_cidrs(start: str, end: str) -> List[str]`

Converts an inclusive IP range into the minimal list of CIDR blocks covering it exactly. start and end must be the same IP version; swapped bounds are corrected.

| Parameter | Type | Description |
|---|---|---|
| `start` | `str` | First IP of the range. |
| `end` | `str` | Last IP of the range. |

Returns `List[str]`. Example: `net.range_to_cidrs("10.0.0.5", "10.0.0.20")`

Keywords: ip range to cidr, start end address to subnet, range, ip range, from to, cidr list, cover

## net.cidr_to_range

`net.cidr_to_range(cidr: str) -> str`

Returns the first and last address of the CIDR as "first - last" (network and broadcast included).

| Parameter | Type | Description |
|---|---|---|
| `cidr` | `str` | CIDR block. |

Returns `str`. Example: `net.cidr_to_range("192.168.1.0/24")`

Keywords: range, first last, from to, start end

## net.cidr_hosts

`net.cidr_hosts(cidr: str, limit: int = 256) -> List[str]`

Lists usable host addresses of the CIDR (network and broadcast excluded for IPv4 up to /30). At most limit addresses are returned (default 256, max 65536).

| Parameter | Type | Description |
|---|---|---|
| `cidr` | `str` | CIDR block. |
| `limit` | `int` (default `256`) | Maximum addresses to list (default 256, max 65536). |

Returns `List[str]`. Example: `net.cidr_hosts("192.168.1.0/29")`

Keywords: addresses, enumerate, expand, all ips, list ips

## net.cidr_next

`net.cidr_next(cidr: str) -> str`

Returns the next network of the same size directly after the given CIDR (10.0.0.0/24 gives 10.0.1.0/24). Fails at the end of the address space.

| Parameter | Type | Description |
|---|---|---|
| `cidr` | `str` | CIDR block. |

Returns `str`. Example: `net.cidr_next("10.0.0.0/24")`

Keywords: following, adjacent, next subnet, successor

## net.cidr_count

`net.cidr_count(cidr: str, usable: bool = False) -> str`

Returns the number of addresses in the CIDR as a decimal string (exact, even for huge IPv6 blocks). With usable true it returns the usable host count instead (IPv4 subtracts network and broadcast up to /30).

| Parameter | Type | Description |
|---|---|---|
| `cidr` | `str` | CIDR block. |
| `usable` | `bool` (default `False`) | Count usable hosts instead of all addresses. |

Returns `str`. Example: `net.cidr_count("10.0.0.0/22")`

Keywords: size, how many, number of ips, hosts count

## net.mask_to_prefix

`net.mask_to_prefix(mask: str) -> int`

Converts a dotted IPv4 netmask (255.255.255.0) or wildcard mask (0.0.0.255) to its prefix length (24). Fails for non-contiguous masks.

| Parameter | Type | Description |
|---|---|---|
| `mask` | `str` | Dotted netmask or wildcard mask. |

Returns `int`. Example: `net.mask_to_prefix("255.255.240.0")`

Keywords: 255.255.255.0 to /24, netmask to prefix length, netmask to cidr, slash, prefix length, mask bits, wildcard_to_prefix

## net.wildcard_to_prefix

`net.wildcard_to_prefix(wildcard: str) -> int`

Converts a wildcard (inverse) mask such as 0.0.0.255 to its prefix length (24), as used in Cisco ACLs and OSPF network statements. Fails for non-contiguous masks.

| Parameter | Type | Description |
|---|---|---|
| `wildcard` | `str` | Wildcard mask, e.g. 0.0.15.255. |

Returns `int`. Example: `net.wildcard_to_prefix("0.0.15.255")`

Keywords: inverse mask, acl, cisco, ospf

## net.prefix_to_mask

`net.prefix_to_mask(prefix: str, version: int = 4, wildcard: bool = False) -> str`

Converts a prefix length (24 or /24, given as a string) to a netmask. IPv4 by default (255.255.255.0); version 6 gives the IPv6 mask. With wildcard true it returns the inverse wildcard mask instead.

| Parameter | Type | Description |
|---|---|---|
| `prefix` | `str` | Prefix length such as "24" or "/24". |
| `version` | `int` (default `4`) | 4 (default) or 6. |
| `wildcard` | `bool` (default `False`) | Return the wildcard mask instead. |

Returns `str`. Example: `net.prefix_to_mask("20")`

Keywords: /24 to netmask, prefix length to subnet mask, cidr to netmask, slash to mask, subnet mask, wildcard

## net.subnet_for_hosts

`net.subnet_for_hosts(hosts: str, base: str = "") -> Dict[str, str]`

Finds the smallest IPv4 subnet that offers at least the needed number of usable hosts (network and broadcast excluded) and returns prefix, netmask and capacity. With base set it also gives the first such subnet inside the base network and how many fit.

| Parameter | Type | Description |
|---|---|---|
| `hosts` | `str` | Needed usable hosts, as a string such as "50". |
| `base` | `str` (default `""`) | Optional base network to allocate from, e.g. 10.0.0.0/16. |

Returns `Dict[str, str]`. Example: `net.subnet_for_hosts("50")`

Keywords: how big, vlsm, host count, size subnet, smallest prefix

## net.vlsm

`net.vlsm(base: str, hosts: List[str]) -> List[str]`

Variable-length subnet masking: allocates subnets from an IPv4 base network for a list of host requirements, largest first, each aligned on its block size. Each requirement is a number or "name:number". Returns a table (header row first) with block, network, mask, usable range and broadcast; fails when the base is too small.

| Parameter | Type | Description |
|---|---|---|
| `base` | `str` | IPv4 network to allocate from, e.g. 192.168.0.0/24. |
| `hosts` | `List[str]` | Host requirements such as ["LAN-A:100", "LAN-B:50", "20"]. |

Returns `List[str]`. Example: `net.vlsm("192.168.0.0/24", ["LAN-A:100", "LAN-B:50", "WAN:2"])`

Keywords: variable length subnet, allocation, plan, address plan, subnetting exercise

## net.subnet_table

`net.subnet_table(cidr: str) -> List[str]`

Quick-reference table for dividing a CIDR: for each longer prefix up to 8 bits deeper (or the maximum), lists the netmask, number of subnets, addresses per subnet and usable hosts per subnet.

| Parameter | Type | Description |
|---|---|---|
| `cidr` | `str` | Network to tabulate, e.g. 10.0.0.0/16. |

Returns `List[str]`. Example: `net.subnet_table("192.168.0.0/24")`

Keywords: cheat sheet, reference, subnet chart, cidr chart, table

## net.int_to_ip

`net.int_to_ip(number: str, version: int = 0) -> str`

Converts a non-negative integer (decimal string, or 0x hex) to an IP address. version 4 or 6 forces the family; the default 0 picks IPv4 when the number fits in 32 bits and IPv6 otherwise.

| Parameter | Type | Description |
|---|---|---|
| `number` | `str` | Integer as a string, e.g. "3232235777" or "0xC0A80101". |
| `version` | `int` (default `0`) | 0 (auto), 4 or 6. |

Returns `str`. Example: `net.int_to_ip("3232235777")`

Keywords: decimal to ip, integer to dotted quad, long to ip, inet_ntoa, numeric to ip

## net.ip_to_hex

`net.ip_to_hex(ip: str) -> str`

Converts an IP address to hexadecimal with a 0x prefix (192.168.1.1 gives 0xc0a80101; IPv6 gives 32 hex digits).

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IPv4 or IPv6 address. |

Returns `str`. Example: `net.ip_to_hex("192.168.1.1")`

Keywords: hexadecimal, ip hex

## net.ip_to_binary

`net.ip_to_binary(ip: str) -> str`

Writes an IP address in binary: IPv4 as four dotted 8-bit groups, IPv6 as eight colon-separated 16-bit groups.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IPv4 or IPv6 address. |

Returns `str`. Example: `net.ip_to_binary("192.168.1.1")`

Keywords: bits, binary ip, octets binary

## net.binary_to_ip

`net.binary_to_ip(bits: str) -> str`

Converts 32 binary digits (IPv4) or 128 binary digits (IPv6) to an address. Dots, colons and spaces between groups are ignored.

| Parameter | Type | Description |
|---|---|---|
| `bits` | `str` | Binary digits, e.g. 11000000.10101000.00000001.00000001. |

Returns `str`. Example: `net.binary_to_ip("11000000.10101000.00000001.00000001")`

Keywords: from binary, bits to ip

## net.hex_to_ip

`net.hex_to_ip(hex: str) -> str`

Converts 8 hex digits (IPv4) or 32 hex digits (IPv6) to an address. A 0x prefix and separators (. : space) are ignored.

| Parameter | Type | Description |
|---|---|---|
| `hex` | `str` | Hex digits such as C0A80101 or 0xc0a80101. |

Returns `str`. Example: `net.hex_to_ip("0xC0A80101")`

Keywords: from hex, hex address

## net.ip_version

`net.ip_version(ip: str) -> int`

Returns 4 for an IPv4 address and 6 for an IPv6 address; fails for anything else. IPv4-mapped addresses (::ffff:1.2.3.4) report 6.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | Address to inspect. |

Returns `int`. Example: `net.ip_version("2001:db8::1")`

Keywords: ipv4 or ipv6, family, v4, v6

## net.ip_classify

`net.ip_classify(ip: str) -> Dict[str, str]`

Detailed classification of an address as a dict: type, description with RFC, scope, and is_global / is_private / is_loopback / is_multicast / is_link_local / is_unspecified flags (as "true"/"false"). Multicast addresses include their scope; IPv4-mapped ones include the embedded IPv4.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IPv4 or IPv6 address. |

Returns `Dict[str, str]`. Example: `net.ip_classify("fe80::1")`

Keywords: rfc1918, special purpose address, classify, ip info, scope, is global, routable, ip properties

## net.ip_normalize

`net.ip_normalize(ip: str) -> str`

Returns the canonical text form of an address: IPv6 is lower-case and compressed (RFC 5952), IPv4 octets lose leading zeros. Also accepts \[brackets\], zones, decimal/hex integers and short forms like 127.1.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | Address in any common spelling. |

Returns `str`. Example: `net.ip_normalize("2001:0DB8:0000:0000:0000:0000:0000:0001")`

Keywords: canonical, clean, format ip, rfc 5952, leading zeros, shorten ipv6

## net.ipv6_expand

`net.ipv6_expand(ip: str) -> str`

Expands an IPv6 address to the full eight 4-digit groups (2001:db8::1 becomes 2001:0db8:0000:0000:0000:0000:0000:0001). IPv4 input is returned as its IPv4-mapped expanded form.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IPv6 address. |

Returns `str`. Example: `net.ipv6_expand("2001:db8::1")`

Keywords: full ipv6 address, uncompress ipv6, full form, uncompress, long form, zero fill

## net.ipv6_compress

`net.ipv6_compress(ip: str) -> str`

Compresses an IPv6 address to the shortest RFC 5952 form: leading zeros dropped and the longest run of zero groups replaced by ::. IPv4-mapped addresses keep their dotted tail.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IPv6 address. |

Returns `str`. Example: `net.ipv6_compress("2001:0db8:0000:0000:0000:0000:0000:0001")`

Keywords: shorten ipv6, abbreviate ipv6, shorten, abbreviate, short form, zero compression

## net.ipv4_to_ipv6_mapped

`net.ipv4_to_ipv6_mapped(ip: str) -> str`

Maps an IPv4 address into IPv6 as ::ffff:a.b.c.d (RFC 4291).

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IPv4 address. |

Returns `str`. Example: `net.ipv4_to_ipv6_mapped("192.0.2.1")`

Keywords: ::ffff, mapped address, dual stack, v4 in v6

## net.ipv6_to_ipv4_mapped

`net.ipv6_to_ipv4_mapped(ip: str) -> str`

Extracts the IPv4 address from an IPv4-mapped IPv6 address (::ffff:192.0.2.1 gives 192.0.2.1). Fails for IPv6 addresses that are not IPv4-mapped.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IPv4-mapped IPv6 address. |

Returns `str`. Example: `net.ipv6_to_ipv4_mapped("::ffff:c000:201")`

Keywords: unmap, extract ipv4, ::ffff

## net.ipv4_to_6to4

`net.ipv4_to_6to4(ip: str) -> str`

Derives the 6to4 IPv6 prefix (RFC 3056) of an IPv4 address: 192.0.2.4 gives 2002:c000:204:: (the /48 prefix).

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IPv4 address. |

Returns `str`. Example: `net.ipv4_to_6to4("192.0.2.4")`

Keywords: 6to4, 2002, tunnel, transition

## net.sixtofour_to_ipv4

`net.sixtofour_to_ipv4(ip: str) -> str`

Extracts the embedded IPv4 address from a 6to4 address (2002:c000:204::1 gives 192.0.2.4). Fails when the address is not inside 2002::/16.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | 6to4 IPv6 address. |

Returns `str`. Example: `net.sixtofour_to_ipv4("2002:c000:204::1")`

Keywords: 6to4, 2002, extract ipv4

## net.nat64

`net.nat64(ip: str, prefix: str = "64:ff9b::/96") -> str`

Synthesizes the NAT64 IPv6 address for an IPv4 address (RFC 6052). The default prefix is the well-known 64:ff9b::/96; /32, /40, /48, /56, /64 and /96 prefixes are supported.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IPv4 address. |
| `prefix` | `str` (default `"64:ff9b::/96"`) | NAT64 prefix with length, default 64:ff9b::/96. |

Returns `str`. Example: `net.nat64("192.0.2.33")`

Keywords: dns64, 64:ff9b, ipv6 translation, rfc6052

## net.nat64_to_ipv4

`net.nat64_to_ipv4(ip: str, prefix: str = "64:ff9b::/96") -> str`

Extracts the IPv4 address embedded in a NAT64 IPv6 address (RFC 6052). The default prefix is 64:ff9b::/96; the address must lie inside the prefix.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | NAT64 IPv6 address. |
| `prefix` | `str` (default `"64:ff9b::/96"`) | NAT64 prefix with length, default 64:ff9b::/96. |

Returns `str`. Example: `net.nat64_to_ipv4("64:ff9b::c000:221")`

Keywords: dns64, 64:ff9b, extract ipv4

## net.eui64_from_mac

`net.eui64_from_mac(mac: str, prefix: str = "fe80::") -> str`

Builds the modified EUI-64 IPv6 address (RFC 4291) for a MAC address: inserts ff:fe in the middle and flips the universal/local bit. prefix is the 64-bit network prefix, default fe80:: (link-local).

| Parameter | Type | Description |
|---|---|---|
| `mac` | `str` | MAC address in any common format. |
| `prefix` | `str` (default `"fe80::"`) | IPv6 /64 prefix, e.g. 2001:db8:1:2:: |

Returns `str`. Example: `net.eui64_from_mac("00:1a:2b:3c:4d:5e", "2001:db8::")`

Keywords: slaac, interface id, modified eui-64, ipv6 from mac

## net.link_local_from_mac

`net.link_local_from_mac(mac: str) -> str`

Returns the SLAAC link-local IPv6 address (fe80::/64 plus modified EUI-64) for a MAC address.

| Parameter | Type | Description |
|---|---|---|
| `mac` | `str` | MAC address in any common format. |

Returns `str`. Example: `net.link_local_from_mac("00:1a:2b:3c:4d:5e")`

Keywords: fe80, slaac, ipv6 from mac, eui-64

## net.ptr_name

`net.ptr_name(ip: str) -> str`

Returns the reverse-DNS name for an address: 1.2.3.4 gives 4.3.2.1.in-addr.arpa, IPv6 gives the 32-nibble ip6.arpa name. No lookup is performed.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IPv4 or IPv6 address. |

Returns `str`. Example: `net.ptr_name("192.0.2.10")`

Keywords: reverse dns name, in-addr.arpa, reverse dns, ip6.arpa, rdns, ptr record, nibble

## net.ip_compare

`net.ip_compare(a: str, b: str) -> int`

Compares two addresses numerically: -1 if a &lt; b, 0 if equal, 1 if a &gt; b. IPv4 sorts before IPv6.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First address. |
| `b` | `str` | Second address. |

Returns `int`. Example: `net.ip_compare("10.0.0.9", "10.0.0.10")`

Keywords: order, greater, less, equal, sort

## net.sort_ips

`net.sort_ips(ips: List[str], descending: bool = False, unique: bool = False) -> List[str]`

Sorts addresses numerically (10.0.0.9 before 10.0.0.10, unlike text sorting), IPv4 first. Entries may also be comma or newline separated. descending reverses the order and unique drops duplicates. Invalid entries are an error.

| Parameter | Type | Description |
|---|---|---|
| `ips` | `List[str]` | Addresses to sort. |
| `descending` | `bool` (default `False`) | Sort high to low. |
| `unique` | `bool` (default `False`) | Remove duplicates. |

Returns `List[str]`. Example: `net.sort_ips(["10.0.0.10", "10.0.0.9", "::1"])`

Keywords: order ips numerically, order ips, natural sort, dedupe ips, numeric sort

## net.ip_in_ranges

`net.ip_in_ranges(ip: str, ranges: List[str]) -> bool`

Reports whether the address lies in at least one of the given CIDRs, single IPs or ranges ("10.0.0.1-10.0.0.50").

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | Address to test. |
| `ranges` | `List[str]` | CIDRs, IPs or a-b ranges. |

Returns `bool`. Example: `net.ip_in_ranges("10.1.2.3", ["192.168.0.0/16", "10.0.0.0/8"])`

Keywords: allowlist, acl, whitelist, in any subnet, match

## net.ip_matching_ranges

`net.ip_matching_ranges(ip: str, ranges: List[str]) -> List[str]`

Returns the entries (CIDRs, IPs or ranges) from the list that contain the address; empty when none do.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | Address to test. |
| `ranges` | `List[str]` | CIDRs, IPs or a-b ranges. |

Returns `List[str]`. Example: `net.ip_matching_ranges("10.1.2.3", ["10.0.0.0/8", "10.1.0.0/16", "172.16.0.0/12"])`

Keywords: which subnet, lookup, find network, matching cidr

## net.ip_add

`net.ip_add(ip: str, n: int) -> str`

Adds a signed integer to an address (10.0.0.255 + 1 = 10.0.1.0). Fails when the result leaves the address family's range.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | Starting address. |
| `n` | `int` | Offset, may be negative. |

Returns `str`. Example: `net.ip_add("10.0.0.255", 1)`

Keywords: increment, offset, next ip, previous ip, plus

## net.ip_diff

`net.ip_diff(a: str, b: str) -> str`

Returns b minus a as a signed decimal string: the number of addresses from a to b (positive when b is higher). Both must be the same IP version.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | Start address. |
| `b` | `str` | End address. |

Returns `str`. Example: `net.ip_diff("10.0.0.1", "10.0.1.1")`

Keywords: difference, distance, how many addresses, subtract

## net.is_valid_ip

`net.is_valid_ip(ip: str) -> bool`

Reports whether the text is a valid IPv4 or IPv6 address (surrounding whitespace, brackets and leading zeros in IPv4 octets are tolerated).

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | Text to check. |

Returns `bool`. Example: `net.is_valid_ip("256.1.1.1")`

Keywords: validate ip, check ip, ip address valid

## net.is_valid_ipv4

`net.is_valid_ipv4(ip: str) -> bool`

Reports whether the text is a valid dotted-quad IPv4 address.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | Text to check. |

Returns `bool`. Example: `net.is_valid_ipv4("192.168.0.1")`

Keywords: validate ipv4, dotted quad

## net.is_valid_ipv6

`net.is_valid_ipv6(ip: str) -> bool`

Reports whether the text is a valid IPv6 address (including :: compression and IPv4-mapped forms).

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | Text to check. |

Returns `bool`. Example: `net.is_valid_ipv6("2001:db8::1")`

Keywords: validate ipv6

## net.is_valid_cidr

`net.is_valid_cidr(cidr: str) -> bool`

Reports whether the text is valid CIDR notation (address/prefix, IPv4 or IPv6). Host bits may be set. Unlike cidr_info it does not accept netmask spellings or a bare address.

| Parameter | Type | Description |
|---|---|---|
| `cidr` | `str` | Text to check. |

Returns `bool`. Example: `net.is_valid_cidr("10.0.0.0/33")`

Keywords: validate cidr, check subnet, prefix valid

## net.ipv4_octets

`net.ipv4_octets(ip: str) -> List[str]`

Splits an IPv4 address into its four octets as decimal strings.

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IPv4 address. |

Returns `List[str]`. Example: `net.ipv4_octets("192.168.1.1")`

Keywords: split ip, octets, bytes, parts

## net.mac_is_valid

`net.mac_is_valid(mac: str) -> bool`

Reports whether the text is a valid 48-bit MAC address: exactly 12 hex digits with optional : - . or space separators.

| Parameter | Type | Description |
|---|---|---|
| `mac` | `str` | MAC address in any common spelling (aa:bb:cc:dd:ee:ff, aa-bb-..., aabb.ccdd.eeff, aabbccddeeff). |

Returns `bool`. Example: `net.mac_is_valid("00:1A:2B:3C:4D")`

Keywords: validate mac, check mac

## net.mac_info

`net.mac_info(mac: str) -> Dict[str, str]`

Describes a MAC address: unicast/multicast/broadcast, universally or locally administered bit, OUI and NIC parts, the EUI-64 form, the SLAAC link-local IPv6 address and, for well-known prefixes, the vendor (the OUI table is partial).

| Parameter | Type | Description |
|---|---|---|
| `mac` | `str` | MAC address in any common spelling (aa:bb:cc:dd:ee:ff, aa-bb-..., aabb.ccdd.eeff, aabbccddeeff). |

Returns `Dict[str, str]`. Example: `net.mac_info("00:50:56:aa:bb:cc")`

Keywords: oui, vendor, multicast, local bit, eui-64, unicast, mac details

## net.oui_vendor

`net.oui_vendor(mac: str) -> str`

Looks up the manufacturer of a MAC address (or an OUI like 00:50:56) in a small embedded table of about 200 well-known prefixes. The table is intentionally partial: an empty string means unknown, not unassigned. Also recognizes broadcast, multicast and Docker addresses.

| Parameter | Type | Description |
|---|---|---|
| `mac` | `str` | MAC address or its first three bytes. |

Returns `str`. Example: `net.oui_vendor("b8:27:eb:12:34:56")`

Keywords: mac vendor, who makes this nic, manufacturer, vendor, oui, who makes, mac lookup, ieee

## net.service_ports

`net.service_ports(name: str) -> List[str]`

Lists every port/protocol of a service name ("http" gives 80/tcp, "domain" gives 53/tcp and 53/udp). Matches exact service names first and then service descriptions (e.g. "mail", "database"); an unknown name gives an empty list.

| Parameter | Type | Description |
|---|---|---|
| `name` | `str` | Service name or keyword such as ssh, dns, redis, mail. |

Returns `List[str]`. Example: `net.service_ports("dns")`

Keywords: port of service, which port, default port, what port does

## net.port_range_type

`net.port_range_type(port: str) -> str`

Classifies a port number as reserved (0), well-known (1-1023), registered (1024-49151) or dynamic/ephemeral (49152-65535) per IANA. Fails outside 0-65535.

| Parameter | Type | Description |
|---|---|---|
| `port` | `str` | Port number as a string, e.g. "8080". |

Returns `str`. Example: `net.port_range_type("8080")`

Keywords: ephemeral, privileged port, system port, user port, dynamic port

## net.protocol_number

`net.protocol_number(name_or_number: str) -> Dict[str, str]`

Looks up an IANA IP protocol number (TCP 6, UDP 17, ICMP 1, GRE 47, ESP 50, OSPF 89...) by number or by name, returning number, keyword and full name. The embedded table covers the commonly seen protocols.

| Parameter | Type | Description |
|---|---|---|
| `name_or_number` | `str` | Protocol keyword ("tcp", "ospf") or number ("47"). |

Returns `Dict[str, str]`. Example: `net.protocol_number("gre")`

Keywords: ip protocol, iana protocol numbers, next header, ip proto, tcp 6, udp 17, icmp

## net.icmp_type

`net.icmp_type(spec: str, code: int = -1, version: int = 4) -> Dict[str, str]`

Names an ICMP message type and code: "3/1" is Destination Unreachable, Host Unreachable. spec is the type, optionally followed by /code; code can also be passed separately (default -1 means none). version 6 uses ICMPv6 (e.g. 135 Neighbor Solicitation). Also lists all codes of the type.

| Parameter | Type | Description |
|---|---|---|
| `spec` | `str` | Type or "type/code", e.g. "8" or "3/4". |
| `code` | `int` (default `-1`) | Code when not part of spec; -1 for none. |
| `version` | `int` (default `4`) | 4 (ICMP, default) or 6 (ICMPv6). |

Returns `Dict[str, str]`. Example: `net.icmp_type("3/4")`

Keywords: icmp, ping, destination unreachable, echo request, ttl exceeded, icmpv6, message type

## net.http_status_search

`net.http_status_search(text: str) -> List[str]`

Finds status codes by words in their name or description ("redirect", "timeout", "auth"), or by class prefix ("4", "40"). Returns lines like "404 Not Found".

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Search words or a code prefix. |

Returns `List[str]`. Example: `net.http_status_search("redirect")`

Keywords: find status, which status code, list status codes, status by name

## net.http_method_info

`net.http_method_info(method: str) -> Dict[str, str]`

Describes an HTTP method (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, TRACE, CONNECT): whether it is safe, idempotent and cacheable, whether it normally carries a body, and what it is for. Case-insensitive.

| Parameter | Type | Description |
|---|---|---|
| `method` | `str` | HTTP method name. |

Returns `Dict[str, str]`. Example: `net.http_method_info("put")`

Keywords: idempotent, safe method, cacheable, verbs, rest, get vs post

## net.mime_type

`net.mime_type(extension_or_filename: str) -> str`

Returns the MIME type for a file extension or filename ("png", ".json", "report.pdf") from an embedded table of about 200 common types. Unknown extensions give application/octet-stream. Case-insensitive; for "archive.tar.gz" the last extension counts.

| Parameter | Type | Description |
|---|---|---|
| `extension_or_filename` | `str` | Extension with or without dot, or a file name/path. |

Returns `str`. Example: `net.mime_type("photo.jpeg")`

Keywords: content type for extension, media type, file type, content type, extension to mime, mimetype

## net.extension_for_mime

`net.extension_for_mime(mime_type: str) -> List[str]`

Lists the file extensions (with dot, most common first) for a MIME type such as image/png; parameters like charset are ignored. Unknown types give an empty list.

| Parameter | Type | Description |
|---|---|---|
| `mime_type` | `str` | MIME type, e.g. application/json or text/html; charset=utf-8. |

Returns `List[str]`. Example: `net.extension_for_mime("image/jpeg")`

Keywords: mime to extension, suffix, file extension for content type

## net.content_type_parse

`net.content_type_parse(header: str) -> Dict[str, str]`

Parses a Content-Type header value such as "text/html; charset=UTF-8" into a dict with type, subtype, full media_type, suffix (+json), charset, boundary and every parameter as param_&lt;name&gt;. Fails for values without a valid media type.

| Parameter | Type | Description |
|---|---|---|
| `header` | `str` | Header value, with or without the "Content-Type:" prefix. |

Returns `Dict[str, str]`. Example: `net.content_type_parse("multipart/form-data; boundary=xyz")`

Keywords: charset, boundary, media type, header, mime parse, content-type

## net.url_build

`net.url_build(parts: Dict[str, str], scheme: str = "https") -> str`

Builds a URL from parts. Recognized keys of parts: scheme, user, password, host (or hostname), port, path, query (raw, without ?) and fragment. IPv6 hosts are bracketed automatically and the path is percent-encoded as needed.

| Parameter | Type | Description |
|---|---|---|
| `parts` | `Dict[str, str]` | Dict with keys scheme, host, port, path, query, fragment, user, password. |
| `scheme` | `str` (default `"https"`) | Default scheme used when parts has none. |

Returns `str`. Example: `net.url_build({"scheme": "https", "host": "example.com", "path": "/a b", "query": "x=1"})`

Keywords: assemble a url, make link from parts, compose url, assemble, construct, uri builder, join parts

## net.url_join

`net.url_join(base: str, ref: str) -> str`

Resolves a relative reference against a base URL following RFC 3986, like a browser does for links (../img/a.png against https://x.com/a/b/c.html gives https://x.com/a/img/a.png). An absolute ref wins over the base.

| Parameter | Type | Description |
|---|---|---|
| `base` | `str` | Base URL. |
| `ref` | `str` | Relative or absolute reference. |

Returns `str`. Example: `net.url_join("https://example.com/a/b/c.html", "../img/x.png")`

Keywords: resolve relative url, combine url segments, resolve, relative link, combine urls, absolute url, href

## net.url_normalize

`net.url_normalize(url: str, sort_query: bool = False, strip_fragment: bool = False) -> str`

Normalizes a URL for comparison: lower-cases scheme and host, removes the default port and a trailing host dot, resolves ./ and ../ segments, upper-cases percent escapes, decodes unreserved characters, adds / for an empty path and drops an empty fragment. sort_query orders query parameters; strip_fragment removes #fragment.

| Parameter | Type | Description |
|---|---|---|
| `url` | `str` | URL such as https://user:pw@example.com:8080/a/b?x=1#top. Input without a scheme is treated as host/path. |
| `sort_query` | `bool` (default `False`) | Sort query parameters alphabetically. |
| `strip_fragment` | `bool` (default `False`) | Remove the #fragment. |

Returns `str`. Example: `net.url_normalize("HTTP://Example.COM:80/a/./b/../c?b=2&a=1", True)`

Keywords: canonical url, clean up url, canonicalize, dedupe urls, clean url, dot segments, compare urls

## net.url_query

`net.url_query(url: str) -> Dict[str, str]`

Returns the query parameters of a URL (or of a bare query string such as a=1&b=2) as a dict with decoded names and values. For repeated keys the values are joined with commas; a key without = has an empty value.

| Parameter | Type | Description |
|---|---|---|
| `url` | `str` | URL or bare query string. |

Returns `Dict[str, str]`. Example: `net.url_query("https://example.com/s?q=go+lang&tag=a&tag=b")`

Keywords: query string, get parameters, params, args, parse query, querystring

## net.url_set_query

`net.url_set_query(url: str, params: Dict[str, str]) -> str`

Replaces the whole query string of a URL with the given parameters (sorted by name, percent-encoded). An empty dict removes the query.

| Parameter | Type | Description |
|---|---|---|
| `url` | `str` | URL to modify. |
| `params` | `Dict[str, str]` | New query parameters. |

Returns `str`. Example: `net.url_set_query("https://example.com/s?old=1", {"q": "go lang", "page": "2"})`

Keywords: change query, set params, rewrite query

## net.url_add_param

`net.url_add_param(url: str, key: str, value: str) -> str`

Appends a query parameter to a URL, keeping existing parameters (a repeated key is added again). Use url_set_param to replace instead.

| Parameter | Type | Description |
|---|---|---|
| `url` | `str` | URL to modify. |
| `key` | `str` | Parameter name. |
| `value` | `str` | Parameter value (encoded for you). |

Returns `str`. Example: `net.url_add_param("https://example.com/s?q=1", "page", "2")`

Keywords: append param, add query, append query

## net.url_set_param

`net.url_set_param(url: str, key: str, value: str) -> str`

Sets a query parameter: replaces the first occurrence in place, removes any further occurrences, or appends it when absent.

| Parameter | Type | Description |
|---|---|---|
| `url` | `str` | URL to modify. |
| `key` | `str` | Parameter name. |
| `value` | `str` | Parameter value (encoded for you). |

Returns `str`. Example: `net.url_set_param("https://example.com/s?q=1&page=1", "page", "5")`

Keywords: replace param, update query, overwrite

## net.url_remove_param

`net.url_remove_param(url: str, keys: List[str]) -> str`

Removes the named query parameters from a URL. A name ending in * is a prefix match (utm_* removes utm_source, utm_medium...). Other parameters keep their original encoding and order.

| Parameter | Type | Description |
|---|---|---|
| `url` | `str` | URL to modify. |
| `keys` | `List[str]` | Parameter names to remove, e.g. ["utm_*", "ref"]. |

Returns `str`. Example: `net.url_remove_param("https://example.com/?a=1&b=2", ["b"])`

Keywords: delete param, strip query, drop parameter

## net.url_strip_tracking

`net.url_strip_tracking(url: str) -> str`

Removes well-known tracking parameters from a URL: utm_*, fbclid, gclid, dclid, msclkid, yclid, mc_cid, mc_eid, igshid, twclid, ttclid, _hsenc, _ga, ref_src, spm, mkt_tok and more. Other parameters and the fragment are kept.

| Parameter | Type | Description |
|---|---|---|
| `url` | `str` | URL to clean. |

Returns `str`. Example: `net.url_strip_tracking("https://example.com/p?id=7&utm_source=x&fbclid=abc")`

Keywords: clean link, strip utm, remove fbclid, remove utm, privacy, tracking, utm, fbclid, gclid, sanitize url

## net.url_origin

`net.url_origin(url: str) -> str`

Returns the web origin (scheme://host\[:port\]) of a URL, lower-cased, with the port omitted when it is the scheme's default. Fails for URLs without scheme or host.

| Parameter | Type | Description |
|---|---|---|
| `url` | `str` | URL such as https://user:pw@example.com:8080/a/b?x=1#top. Input without a scheme is treated as host/path. |

Returns `str`. Example: `net.url_origin("https://Example.com:443/a?b=1")`

Keywords: same origin, cors, scheme host port, base url

## net.url_is_valid

`net.url_is_valid(url: str) -> bool`

Reports whether the text is an absolute URL: a scheme, then for http/https/ftp/ws/wss a valid host (domain, IPv4 or \[IPv6\]) and an optional valid port. Other schemes (mailto:, file:, data:) only need to parse. Whitespace makes it invalid.

| Parameter | Type | Description |
|---|---|---|
| `url` | `str` | Text to check. |

Returns `bool`. Example: `net.url_is_valid("https://example.com/path")`

Keywords: validate url, check link, well-formed

## net.is_valid_hostname

`net.is_valid_hostname(host: str, allow_underscore: bool = False) -> bool`

Checks RFC 1123 hostname syntax: labels of 1-63 letters, digits and hyphens (no leading or trailing hyphen), at most 253 characters in total. A single label such as localhost is fine. allow_underscore also accepts _ (service names, DKIM).

| Parameter | Type | Description |
|---|---|---|
| `host` | `str` | Hostname to check. |
| `allow_underscore` | `bool` (default `False`) | Accept underscores in labels. |

Returns `bool`. Example: `net.is_valid_hostname("my-server.local")`

Keywords: validate hostname, rfc 1123, dns name

## net.is_valid_domain

`net.is_valid_domain(host: str) -> bool`

Checks that the text is a plausible fully qualified domain name: valid hostname syntax, at least two labels, an alphabetic (or xn--) top-level label and not an IP address. It does not check that the TLD exists.

| Parameter | Type | Description |
|---|---|---|
| `host` | `str` | Domain name to check. |

Returns `bool`. Example: `net.is_valid_domain("example.com")`

Keywords: validate domain, fqdn, check domain

## net.email_parse

`net.email_parse(email: str) -> Dict[str, str]`

Splits an email address into local, domain, base_local, +tag, registrable_domain and a normalized form (lower-case; Gmail/Googlemail addresses also lose dots and the +tag, googlemail.com becomes gmail.com). Accepts "Name &lt;a@b.c&gt;" and mailto: prefixes. Fails for invalid addresses.

| Parameter | Type | Description |
|---|---|---|
| `email` | `str` | Email address. |

Returns `Dict[str, str]`. Example: `net.email_parse("John.Doe+news@Gmail.com")`

Keywords: split email into user and domain, local part, split email, plus addressing, subaddress, gmail normalize, dedupe email

## net.is_valid_email

`net.is_valid_email(email: str) -> bool`

Pragmatic RFC 5322 check: a dot-atom or quoted local part (max 64 chars), an @, and a hostname domain with at least two labels or an \[IP literal\]; total length up to 254. No display names, comments or whitespace. It does not verify the mailbox exists.

| Parameter | Type | Description |
|---|---|---|
| `email` | `str` | Email address. |

Returns `bool`. Example: `net.is_valid_email("user@example.com")`

Keywords: check email syntax, validate email, check email, email syntax

## net.reverse_domain_name

`net.reverse_domain_name(host: str) -> str`

Reverses the labels of a domain: a.b.com becomes com.b.a (Java package / Android application id style). A URL is reduced to its host first.

| Parameter | Type | Description |
|---|---|---|
| `host` | `str` | Domain name. |

Returns `str`. Example: `net.reverse_domain_name("mail.example.com")`

Keywords: package name, bundle id, reverse dns, com.example

## net.basic_auth_header

`net.basic_auth_header(user: str, password: str) -> str`

Builds an HTTP Basic Authorization header value: "Basic " plus base64(user:password). The user name must not contain a colon.

| Parameter | Type | Description |
|---|---|---|
| `user` | `str` | User name (no colon). |
| `password` | `str` | Password. |

Returns `str`. Example: `net.basic_auth_header("alice", "s3cret")`

Keywords: authorization header, username password base64, authorization, http auth, credentials, base64, curl -u

## net.parse_basic_auth

`net.parse_basic_auth(header: str) -> Dict[str, str]`

Decodes an HTTP Basic Authorization header (with or without the "Basic " prefix and "Authorization:" name) into user and password. Fails for invalid base64 or a missing colon.

| Parameter | Type | Description |
|---|---|---|
| `header` | `str` | Header value such as Basic YWxpY2U6czNjcmV0. |

Returns `Dict[str, str]`. Example: `net.parse_basic_auth("Basic YWxpY2U6czNjcmV0")`

Keywords: decode authorization, credentials, base64 decode

## net.parse_host_port

`net.parse_host_port(address: str) -> Dict[str, str]`

Splits host:port, including \[IPv6\]:port, bare IPv6 (no port), IPv4:port and name:port. Returns host, port (empty when absent), host_type (ipv4, ipv6, domain) and the bracketed IPv6 form. Fails for invalid hosts or ports above 65535.

| Parameter | Type | Description |
|---|---|---|
| `address` | `str` | Address such as example.com:8080 or [::1]:443. |

Returns `Dict[str, str]`. Example: `net.parse_host_port("[::1]:8080")`

Keywords: split host port, socket address, endpoint, ipv6 port

## net.hosts_entry

`net.hosts_entry(ip: str, names: List[str]) -> str`

Formats a hosts-file line (ip, tab, space-separated names) after validating the address and host names, e.g. "127.0.0.1\tapp.local www.app.local".

| Parameter | Type | Description |
|---|---|---|
| `ip` | `str` | IP address. |
| `names` | `List[str]` | Host names to map to the address. |

Returns `str`. Example: `net.hosts_entry("127.0.0.1", ["app.local", "www.app.local"])`

Keywords: /etc/hosts, windows hosts, host mapping, local dns

## net.bandwidth

`net.bandwidth(value: float, from_unit: str, to_unit: str) -> float`

Converts a data rate between units: bps, Kbps, Mbps, Gbps, Tbps (bits, decimal), B/s, KB/s, MB/s, GB/s (bytes, decimal) and KiB/s, MiB/s, GiB/s (bytes, binary). A capital B means bytes, lower-case b means bits; /s or ps both work.

| Parameter | Type | Description |
|---|---|---|
| `value` | `float` | Rate in from_unit. |
| `from_unit` | `str` | Unit such as "Mbps" or "MB/s". |
| `to_unit` | `str` | Target unit such as "MiB/s". |

Returns `float`. Example: `net.bandwidth(100, "Mbps", "MB/s")`

Keywords: mbps to mb/s, megabit, megabyte per second, speed conversion, data rate, throughput, download speed

## net.transfer_time

`net.transfer_time(size: str, bandwidth: str) -> str`

How long a transfer takes: size such as "1.5 GB" or "700 MiB" at a bandwidth such as "100 Mbps" or "12 MB/s". Returns a human duration like "2m 4s". Protocol overhead is ignored (ideal line rate).

| Parameter | Type | Description |
|---|---|---|
| `size` | `str` | Amount of data, e.g. "4.7 GB". |
| `bandwidth` | `str` | Rate, e.g. "50 Mbps". |

Returns `str`. Example: `net.transfer_time("4.7 GB", "100 Mbps")`

Keywords: how long to copy, bandwidth and size, download time, how long, eta, upload time, duration, copy time

## net.throughput_needed

`net.throughput_needed(size: str, seconds: float) -> Dict[str, str]`

Bandwidth required to move a size (e.g. "10 GB") in the given number of seconds. Returns the rate in bps, Kbps, Mbps, Gbps, MB/s and MiB/s as strings.

| Parameter | Type | Description |
|---|---|---|
| `size` | `str` | Amount of data, e.g. "10 GB". |
| `seconds` | `float` | Time allowed in seconds (> 0). |

Returns `Dict[str, str]`. Example: `net.throughput_needed("10 GB", 600)`

Keywords: required speed, minimum bandwidth, how fast, mbps needed


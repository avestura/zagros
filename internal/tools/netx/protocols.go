package netx

import (
	"fmt"
	"mime"
	"sort"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

// protoData: number|keyword|name (IANA protocol numbers, selected).
const protoData = `
0|HOPOPT|IPv6 Hop-by-Hop Option
1|ICMP|Internet Control Message Protocol
2|IGMP|Internet Group Management Protocol
3|GGP|Gateway-to-Gateway Protocol
4|IPv4|IPv4 encapsulation (IP-in-IP)
5|ST|Stream Protocol
6|TCP|Transmission Control Protocol
7|CBT|Core-based trees
8|EGP|Exterior Gateway Protocol
9|IGP|Any private interior gateway (Cisco IGRP)
12|PUP|PARC Universal Packet
17|UDP|User Datagram Protocol
20|HMP|Host Monitoring Protocol
22|XNS-IDP|Xerox NS IDP
27|RDP|Reliable Data Protocol
29|ISO-TP4|ISO Transport Protocol Class 4
33|DCCP|Datagram Congestion Control Protocol
36|XTP|Xpress Transport Protocol
41|IPv6|IPv6 encapsulation (6in4)
43|IPv6-Route|Routing Header for IPv6
44|IPv6-Frag|Fragment Header for IPv6
46|RSVP|Reservation Protocol
47|GRE|Generic Routing Encapsulation
50|ESP|IPsec Encapsulating Security Payload
51|AH|IPsec Authentication Header
54|NARP|NBMA Address Resolution Protocol
55|MOBILE|IP Mobility
58|IPv6-ICMP|ICMP for IPv6
59|IPv6-NoNxt|No Next Header for IPv6
60|IPv6-Opts|Destination Options for IPv6
80|ISO-IP|ISO Internet Protocol
83|VINES|Banyan VINES
88|EIGRP|Enhanced Interior Gateway Routing Protocol
89|OSPF|Open Shortest Path First
91|LARP|Locus Address Resolution Protocol
93|AX.25|AX.25 Frames
94|IPIP|IP-within-IP Encapsulation Protocol
97|ETHERIP|Ethernet-within-IP Encapsulation
98|ENCAP|Encapsulation Header
103|PIM|Protocol Independent Multicast
108|IPComp|IP Payload Compression Protocol
111|IPX-in-IP|IPX in IP
112|VRRP|Virtual Router Redundancy Protocol
115|L2TP|Layer Two Tunneling Protocol
118|STP|Schedule Transfer Protocol
123|PTP|Performance Transparency Protocol
124|ISIS|IS-IS over IPv4
126|CRTP|Combat Radio Transport Protocol
132|SCTP|Stream Control Transmission Protocol
133|FC|Fibre Channel
135|MH|Mobility Header
136|UDPLite|Lightweight User Datagram Protocol
137|MPLS-in-IP|MPLS in IP
139|HIP|Host Identity Protocol
140|Shim6|Shim6 Protocol
141|WESP|Wrapped Encapsulating Security Payload
142|ROHC|Robust Header Compression
143|Ethernet|Ethernet (RFC 8986 SRv6)
253|EXP1|Use for experimentation and testing (RFC 3692)
254|EXP2|Use for experimentation and testing (RFC 3692)
255|RESERVED|Reserved
`

type protoEntry struct {
	num       int
	key, name string
}

var protoTable = func() []protoEntry {
	var out []protoEntry
	for _, l := range strings.Split(protoData, "\n") {
		f := strings.SplitN(strings.TrimSpace(l), "|", 3)
		if len(f) != 3 {
			continue
		}
		n, _ := strconv.Atoi(f[0])
		out = append(out, protoEntry{n, f[1], f[2]})
	}
	return out
}()

func lookupProto(s string) (protoEntry, bool) {
	s = strings.TrimSpace(s)
	if n, err := strconv.Atoi(s); err == nil {
		for _, e := range protoTable {
			if e.num == n {
				return e, true
			}
		}
		return protoEntry{}, false
	}
	l := strings.ToLower(s)
	alias := map[string]string{"ospfigp": "ospf", "ip-in-ip": "ipip", "ipv6-icmp": "ipv6-icmp", "icmpv6": "ipv6-icmp", "icmp6": "ipv6-icmp", "ipsec-esp": "esp", "ipsec-ah": "ah", "ip": "ipv4"}
	if a, ok := alias[l]; ok {
		l = a
	}
	for _, e := range protoTable {
		if strings.ToLower(e.key) == l {
			return e, true
		}
	}
	for _, e := range protoTable {
		if strings.EqualFold(e.name, s) {
			return e, true
		}
	}
	return protoEntry{}, false
}

type icmpT struct {
	name  string
	codes map[int]string
}

var icmp4 = map[int]icmpT{
	0:  {"Echo Reply", nil},
	3:  {"Destination Unreachable", map[int]string{0: "Net Unreachable", 1: "Host Unreachable", 2: "Protocol Unreachable", 3: "Port Unreachable", 4: "Fragmentation Needed and DF Set", 5: "Source Route Failed", 6: "Destination Network Unknown", 7: "Destination Host Unknown", 8: "Source Host Isolated", 9: "Network Administratively Prohibited", 10: "Host Administratively Prohibited", 11: "Network Unreachable for TOS", 12: "Host Unreachable for TOS", 13: "Communication Administratively Prohibited", 14: "Host Precedence Violation", 15: "Precedence Cutoff in Effect"}},
	4:  {"Source Quench (deprecated)", nil},
	5:  {"Redirect", map[int]string{0: "Redirect for Network", 1: "Redirect for Host", 2: "Redirect for TOS and Network", 3: "Redirect for TOS and Host"}},
	8:  {"Echo Request", nil},
	9:  {"Router Advertisement", nil},
	10: {"Router Solicitation", nil},
	11: {"Time Exceeded", map[int]string{0: "TTL Exceeded in Transit", 1: "Fragment Reassembly Time Exceeded"}},
	12: {"Parameter Problem", map[int]string{0: "Pointer Indicates the Error", 1: "Missing a Required Option", 2: "Bad Length"}},
	13: {"Timestamp Request", nil},
	14: {"Timestamp Reply", nil},
	15: {"Information Request (deprecated)", nil},
	16: {"Information Reply (deprecated)", nil},
	17: {"Address Mask Request (deprecated)", nil},
	18: {"Address Mask Reply (deprecated)", nil},
	30: {"Traceroute (deprecated)", nil},
	42: {"Extended Echo Request", nil},
	43: {"Extended Echo Reply", map[int]string{0: "No Error", 1: "Malformed Query", 2: "No Such Interface", 3: "No Such Table Entry", 4: "Multiple Interfaces Satisfy Query"}},
}

var icmp6 = map[int]icmpT{
	1:   {"Destination Unreachable", map[int]string{0: "No Route to Destination", 1: "Communication Administratively Prohibited", 2: "Beyond Scope of Source Address", 3: "Address Unreachable", 4: "Port Unreachable", 5: "Source Address Failed Ingress/Egress Policy", 6: "Reject Route to Destination"}},
	2:   {"Packet Too Big", nil},
	3:   {"Time Exceeded", map[int]string{0: "Hop Limit Exceeded in Transit", 1: "Fragment Reassembly Time Exceeded"}},
	4:   {"Parameter Problem", map[int]string{0: "Erroneous Header Field", 1: "Unrecognized Next Header Type", 2: "Unrecognized IPv6 Option"}},
	128: {"Echo Request", nil},
	129: {"Echo Reply", nil},
	130: {"Multicast Listener Query", nil},
	131: {"Multicast Listener Report", nil},
	132: {"Multicast Listener Done", nil},
	133: {"Router Solicitation", nil},
	134: {"Router Advertisement", nil},
	135: {"Neighbor Solicitation", nil},
	136: {"Neighbor Advertisement", nil},
	137: {"Redirect Message", nil},
	138: {"Router Renumbering", nil},
	139: {"ICMP Node Information Query", nil},
	140: {"ICMP Node Information Response", nil},
	141: {"Inverse Neighbor Discovery Solicitation", nil},
	142: {"Inverse Neighbor Discovery Advertisement", nil},
	143: {"Multicast Listener Report v2", nil},
	144: {"Home Agent Address Discovery Request", nil},
	145: {"Home Agent Address Discovery Reply", nil},
	146: {"Mobile Prefix Solicitation", nil},
	147: {"Mobile Prefix Advertisement", nil},
	148: {"Certification Path Solicitation", nil},
	149: {"Certification Path Advertisement", nil},
	151: {"Multicast Router Advertisement", nil},
	152: {"Multicast Router Solicitation", nil},
	153: {"Multicast Router Termination", nil},
	155: {"RPL Control Message", nil},
}

// icmpLookup parses "3", "3/1", "3 1" or "3,1" plus an optional code argument.
func icmpLookup(spec string, code int64, v6 bool) (map[string]string, error) {
	spec = strings.TrimSpace(spec)
	f := strings.FieldsFunc(spec, func(r rune) bool { return r == '/' || r == ',' || r == ' ' || r == ':' })
	if len(f) == 0 || len(f) > 2 {
		return nil, fmt.Errorf("invalid ICMP type %q (use 3 or 3/1)", spec)
	}
	t, err := strconv.Atoi(f[0])
	if err != nil || t < 0 || t > 255 {
		return nil, fmt.Errorf("invalid ICMP type %q", f[0])
	}
	c := int(code)
	if len(f) == 2 {
		c, err = strconv.Atoi(f[1])
		if err != nil || c < 0 || c > 255 {
			return nil, fmt.Errorf("invalid ICMP code %q", f[1])
		}
	}
	table, ver := icmp4, "4"
	if v6 {
		table, ver = icmp6, "6"
	}
	m := map[string]string{"version": ver, "type": strconv.Itoa(t)}
	e, ok := table[t]
	if !ok {
		m["type_name"] = "unassigned or experimental"
	} else {
		m["type_name"] = e.name
	}
	if c >= 0 {
		m["code"] = strconv.Itoa(c)
		switch {
		case ok && e.codes != nil:
			if n, has := e.codes[c]; has {
				m["code_name"] = n
			} else {
				m["code_name"] = "unassigned"
			}
		case c == 0:
			m["code_name"] = "no code (always 0)"
		default:
			m["code_name"] = "unassigned"
		}
	}
	if ok && e.codes != nil {
		var cs []int
		for k := range e.codes {
			cs = append(cs, k)
		}
		sort.Ints(cs)
		parts := make([]string, len(cs))
		for i, k := range cs {
			parts[i] = fmt.Sprintf("%d=%s", k, e.codes[k])
		}
		m["codes"] = strings.Join(parts, "; ")
	}
	return m, nil
}

// ---------------------------------------------------------------- HTTP

type statusT struct {
	code int
	text string
	desc string
}

var statusTable = []statusT{
	{100, "Continue", "The server received the request headers; the client should send the body."},
	{101, "Switching Protocols", "The server agrees to switch protocols as requested (for example to WebSocket)."},
	{102, "Processing", "WebDAV: the server is still processing the request (no final response yet)."},
	{103, "Early Hints", "Preliminary headers (such as Link preload) sent before the final response."},
	{200, "OK", "The request succeeded."},
	{201, "Created", "The request succeeded and a new resource was created."},
	{202, "Accepted", "The request was accepted for processing but is not finished."},
	{203, "Non-Authoritative Information", "The response comes from a transforming proxy, not the origin server."},
	{204, "No Content", "Success with no response body."},
	{205, "Reset Content", "Success; the client should reset the document view (form)."},
	{206, "Partial Content", "Only part of the resource is returned (Range request)."},
	{207, "Multi-Status", "WebDAV: multiple independent status codes in an XML body."},
	{208, "Already Reported", "WebDAV: members were already listed in a previous reply."},
	{226, "IM Used", "The response is the result of instance manipulations (delta encoding)."},
	{300, "Multiple Choices", "Several representations are available; the client may choose."},
	{301, "Moved Permanently", "The resource has a new permanent URL (clients may change POST to GET)."},
	{302, "Found", "Temporary redirect; the URL may change again."},
	{303, "See Other", "Redirect to another URL, which should be fetched with GET."},
	{304, "Not Modified", "The cached version is still valid (conditional request)."},
	{305, "Use Proxy", "Deprecated: the resource must be accessed through a proxy."},
	{307, "Temporary Redirect", "Temporary redirect that preserves the HTTP method and body."},
	{308, "Permanent Redirect", "Permanent redirect that preserves the HTTP method and body."},
	{400, "Bad Request", "The server cannot process the request because of a client error (malformed syntax)."},
	{401, "Unauthorized", "Authentication is required or failed; the client is not identified."},
	{402, "Payment Required", "Reserved for digital payment systems; rarely used."},
	{403, "Forbidden", "The client is known but not allowed to access the resource."},
	{404, "Not Found", "The resource does not exist (or the server hides it)."},
	{405, "Method Not Allowed", "The HTTP method is not supported for this resource; see the Allow header."},
	{406, "Not Acceptable", "No representation matches the client's Accept headers."},
	{407, "Proxy Authentication Required", "The client must authenticate with the proxy first."},
	{408, "Request Timeout", "The server timed out waiting for the request."},
	{409, "Conflict", "The request conflicts with the current state of the resource."},
	{410, "Gone", "The resource used to exist and is permanently removed."},
	{411, "Length Required", "The request needs a Content-Length header."},
	{412, "Precondition Failed", "A precondition in the request headers (If-Match...) evaluated to false."},
	{413, "Content Too Large", "The request body is larger than the server accepts (Payload Too Large)."},
	{414, "URI Too Long", "The request URL is longer than the server accepts."},
	{415, "Unsupported Media Type", "The request body format is not supported."},
	{416, "Range Not Satisfiable", "The requested byte range cannot be served."},
	{417, "Expectation Failed", "The Expect request header cannot be met."},
	{418, "I'm a teapot", "April Fools' joke from RFC 2324 (HTCPCP); the teapot refuses to brew coffee."},
	{421, "Misdirected Request", "The request was sent to a server that cannot produce a response."},
	{422, "Unprocessable Content", "The request is well-formed but semantically invalid (validation error)."},
	{423, "Locked", "WebDAV: the resource is locked."},
	{424, "Failed Dependency", "WebDAV: the request failed because a previous request failed."},
	{425, "Too Early", "The server will not risk processing a request that might be replayed."},
	{426, "Upgrade Required", "The client must switch to another protocol (see Upgrade header)."},
	{428, "Precondition Required", "The server requires the request to be conditional."},
	{429, "Too Many Requests", "Rate limiting: the client sent too many requests; see Retry-After."},
	{431, "Request Header Fields Too Large", "The headers are too large."},
	{451, "Unavailable For Legal Reasons", "Access is denied for legal reasons (censorship, court order)."},
	{500, "Internal Server Error", "Generic server-side failure."},
	{501, "Not Implemented", "The server does not support the functionality required."},
	{502, "Bad Gateway", "A gateway or proxy received an invalid response from upstream."},
	{503, "Service Unavailable", "The server is overloaded or down for maintenance; see Retry-After."},
	{504, "Gateway Timeout", "A gateway or proxy timed out waiting for the upstream server."},
	{505, "HTTP Version Not Supported", "The HTTP version of the request is not supported."},
	{506, "Variant Also Negotiates", "Transparent content negotiation configuration error."},
	{507, "Insufficient Storage", "WebDAV: the server cannot store the representation."},
	{508, "Loop Detected", "WebDAV: an infinite loop was detected while processing."},
	{510, "Not Extended", "Further extensions to the request are required (obsolete)."},
	{511, "Network Authentication Required", "The client must authenticate to gain network access (captive portal)."},
	// widely seen non-standard codes
	{499, "Client Closed Request", "Non-standard (nginx): the client closed the connection before the response."},
	{520, "Web Server Returned an Unknown Error", "Non-standard (Cloudflare): the origin returned an unexpected response."},
	{521, "Web Server Is Down", "Non-standard (Cloudflare): the origin refused the connection."},
	{522, "Connection Timed Out", "Non-standard (Cloudflare): connecting to the origin timed out."},
	{523, "Origin Is Unreachable", "Non-standard (Cloudflare): the origin cannot be reached."},
	{524, "A Timeout Occurred", "Non-standard (Cloudflare): the origin did not reply in time."},
	{525, "SSL Handshake Failed", "Non-standard (Cloudflare): the TLS handshake with the origin failed."},
	{526, "Invalid SSL Certificate", "Non-standard (Cloudflare): the origin certificate is invalid."},
}

func statusClass(c int) string {
	switch c / 100 {
	case 1:
		return "1xx Informational"
	case 2:
		return "2xx Success"
	case 3:
		return "3xx Redirection"
	case 4:
		return "4xx Client Error"
	case 5:
		return "5xx Server Error"
	}
	return "unknown"
}

func statusDict(e statusT) map[string]string {
	std := "standard"
	if e.code == 499 || e.code >= 520 {
		std = "non-standard"
	}
	return map[string]string{"code": strconv.Itoa(e.code), "text": e.text, "class": statusClass(e.code), "description": e.desc, "registry": std}
}

type methodT struct {
	name                                 string
	safe, idempotent, cacheable, hasBody bool
	desc                                 string
}

var methodTable = []methodT{
	{"GET", true, true, true, false, "Retrieve a representation of the resource."},
	{"HEAD", true, true, true, false, "Like GET but returns headers only."},
	{"POST", false, false, true, true, "Submit data for processing; usually creates a resource or triggers an action."},
	{"PUT", false, true, false, true, "Replace the resource with the request body (create if missing)."},
	{"PATCH", false, false, false, true, "Apply a partial modification to the resource."},
	{"DELETE", false, true, false, false, "Remove the resource."},
	{"OPTIONS", true, true, false, false, "Describe the communication options (CORS preflight, Allow header)."},
	{"TRACE", true, true, false, false, "Echo the received request for diagnostics (usually disabled)."},
	{"CONNECT", false, false, false, false, "Open a tunnel through a proxy (HTTPS proxying)."},
}

func yn(b bool) string { return strconv.FormatBool(b) }

func init() {
	reg(kwPort, tools.Func{
		Name: "protocol_number", Title: "IP protocol number", Order: 82, Input: "name_or_number",
		Doc:     "Looks up an IANA IP protocol number (TCP 6, UDP 17, ICMP 1, GRE 47, ESP 50, OSPF 89...) by number or by name, returning number, keyword and full name. The embedded table covers the commonly seen protocols.",
		Params:  []tools.Param{req("name_or_number", "str", `Protocol keyword ("tcp", "ospf") or number ("47").`)},
		Returns: "dict[str, str]", Example: `net.protocol_number("gre")`,
		Tags:    []string{"ip protocol", "iana protocol numbers", "next header", "ip proto", "tcp 6", "udp 17", "icmp"},
		Accepts: func(s string) bool { _, ok := lookupProto(s); return ok },
		Fn: func(a tools.Args) (any, error) {
			e, ok := lookupProto(a.Str("name_or_number"))
			if !ok {
				return nil, fmt.Errorf("unknown IP protocol %q", a.Str("name_or_number"))
			}
			return map[string]string{"number": strconv.Itoa(e.num), "keyword": e.key, "name": e.name}, nil
		},
	})
	reg(kwPort, tools.Func{
		Name: "icmp_type", Title: "ICMP type and code names", Order: 83, Input: "spec",
		Doc:     `Names an ICMP message type and code: "3/1" is Destination Unreachable, Host Unreachable. spec is the type, optionally followed by /code; code can also be passed separately (default -1 means none). version 6 uses ICMPv6 (e.g. 135 Neighbor Solicitation). Also lists all codes of the type.`,
		Params:  []tools.Param{req("spec", "str", `Type or "type/code", e.g. "8" or "3/4".`), opt("code", "int", "Code when not part of spec; -1 for none.", int64(-1)), opt("version", "int", "4 (ICMP, default) or 6 (ICMPv6).", int64(4))},
		Returns: "dict[str, str]", Example: `net.icmp_type("3/4")`,
		Tags:    []string{"icmp", "ping", "destination unreachable", "echo request", "ttl exceeded", "icmpv6", "message type"},
		Accepts: func(s string) bool { _, err := icmpLookup(s, -1, false); return err == nil },
		Fn: func(a tools.Args) (any, error) {
			v := a.Int("version")
			if v != 4 && v != 6 {
				return nil, fmt.Errorf("version must be 4 or 6")
			}
			return icmpLookup(a.Str("spec"), a.Int("code"), v == 6)
		},
	})
	reg(kwStatus, tools.Func{
		Name: "http_status", Title: "HTTP status code", Order: 7, Input: "code",
		Doc:     "Explains an HTTP status code (pass it as a string, e.g. \"404\"): reason phrase, class (1xx-5xx) and a description. Covers all IANA-registered codes plus the common nginx/Cloudflare ones; unregistered codes still report their class.",
		Params:  []tools.Param{req("code", "str", `Status code such as "404" or "503".`)},
		Returns: "dict[str, str]", Example: `net.http_status("404")`,
		Tags: []string{"status code", "http error", "404", "500", "301", "reason phrase", "response code", "rest"},
		Accepts: func(s string) bool {
			n, err := strconv.Atoi(strings.TrimSpace(s))
			return err == nil && n >= 100 && n <= 599
		},
		Fn: func(a tools.Args) (any, error) {
			n, err := strconv.Atoi(strings.TrimSpace(a.Str("code")))
			if err != nil || n < 100 || n > 599 {
				return nil, fmt.Errorf("invalid HTTP status code %q (100-599)", a.Str("code"))
			}
			for _, e := range statusTable {
				if e.code == n {
					return statusDict(e), nil
				}
			}
			return map[string]string{"code": strconv.Itoa(n), "text": "", "class": statusClass(n), "description": "unregistered status code", "registry": "unassigned"}, nil
		},
	})
	reg(kwStatus, tools.Func{
		Name: "http_status_search", Title: "Search HTTP status codes", Order: 84, Input: "text",
		Doc:     `Finds status codes by words in their name or description ("redirect", "timeout", "auth"), or by class prefix ("4", "40"). Returns lines like "404 Not Found".`,
		Params:  []tools.Param{req("text", "str", "Search words or a code prefix.")},
		Returns: "list[str]", Example: `net.http_status_search("redirect")`,
		Tags:    []string{"find status", "which status code", "list status codes", "status by name"},
		Accepts: func(s string) bool { return len(strings.TrimSpace(s)) >= 2 },
		Fn: func(a tools.Args) (any, error) {
			q := strings.ToLower(strings.TrimSpace(a.Str("text")))
			out := []string{}
			if q == "" {
				return out, nil
			}
			for _, e := range statusTable {
				code := strconv.Itoa(e.code)
				hay := strings.ToLower(e.text + " " + e.desc)
				if strings.HasPrefix(code, q) || containsAll(hay, strings.Fields(q)) {
					out = append(out, code+" "+e.text)
				}
			}
			sort.Strings(out)
			return out, nil
		},
	})
	reg(kwStatus, tools.Func{
		Name: "http_method_info", Title: "HTTP method properties", Order: 85, Input: "method",
		Doc:     "Describes an HTTP method (GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS, TRACE, CONNECT): whether it is safe, idempotent and cacheable, whether it normally carries a body, and what it is for. Case-insensitive.",
		Params:  []tools.Param{req("method", "str", "HTTP method name.")},
		Returns: "dict[str, str]", Example: `net.http_method_info("put")`,
		Tags:    []string{"idempotent", "safe method", "cacheable", "verbs", "rest", "get vs post"},
		Accepts: func(s string) bool { return findMethod(s) != nil },
		Fn: func(a tools.Args) (any, error) {
			m := findMethod(a.Str("method"))
			if m == nil {
				return nil, fmt.Errorf("unknown HTTP method %q", a.Str("method"))
			}
			return map[string]string{"method": m.name, "safe": yn(m.safe), "idempotent": yn(m.idempotent), "cacheable": yn(m.cacheable), "request_body": yn(m.hasBody), "description": m.desc}, nil
		},
	})
	reg(kwMime, tools.Func{
		Name: "mime_type", Title: "MIME type of file", Order: 86, Input: "extension_or_filename",
		Doc:     "Returns the MIME type for a file extension or filename (\"png\", \".json\", \"report.pdf\") from an embedded table of about 200 common types. Unknown extensions give application/octet-stream. Case-insensitive; for \"archive.tar.gz\" the last extension counts.",
		Params:  []tools.Param{req("extension_or_filename", "str", "Extension with or without dot, or a file name/path.")},
		Returns: "str", Example: `net.mime_type("photo.jpeg")`,
		Tags: []string{"content type", "media type", "file type", "extension to mime", "mimetype"},
		Accepts: func(s string) bool {
			s = strings.TrimSpace(s)
			return s != "" && !strings.ContainsAny(s, " \t\n") && len(s) < 260
		},
		Fn: func(a tools.Args) (any, error) {
			ext := extOf(a.Str("extension_or_filename"))
			if ext == "" {
				return nil, fmt.Errorf("no file extension in %q", a.Str("extension_or_filename"))
			}
			if t, ok := mimeByExt[ext]; ok {
				return t, nil
			}
			return "application/octet-stream", nil
		},
	})
	reg(kwMime, tools.Func{
		Name: "extension_for_mime", Title: "File extensions of MIME type", Order: 87, Input: "mime_type",
		Doc:     "Lists the file extensions (with dot, most common first) for a MIME type such as image/png; parameters like charset are ignored. Unknown types give an empty list.",
		Params:  []tools.Param{req("mime_type", "str", "MIME type, e.g. application/json or text/html; charset=utf-8.")},
		Returns: "list[str]", Example: `net.extension_for_mime("image/jpeg")`,
		Tags: []string{"mime to extension", "suffix", "file extension for content type"},
		Accepts: func(s string) bool {
			return strings.Contains(s, "/") && !strings.ContainsAny(strings.TrimSpace(s), " ") || strings.Contains(s, ";")
		},
		Fn: func(a tools.Args) (any, error) {
			t := strings.ToLower(strings.TrimSpace(strings.SplitN(a.Str("mime_type"), ";", 2)[0]))
			out := []string{}
			for _, e := range mimeOrder {
				if mimeByExt[e] == t {
					out = append(out, "."+e)
				}
			}
			return out, nil
		},
	})
	reg(kwMime, tools.Func{
		Name: "content_type_parse", Title: "Parse Content-Type header", Order: 88, Input: "header",
		Doc:     `Parses a Content-Type header value such as "text/html; charset=UTF-8" into a dict with type, subtype, full media_type, suffix (+json), charset, boundary and every parameter as param_<name>. Fails for values without a valid media type.`,
		Params:  []tools.Param{req("header", "str", "Header value, with or without the \"Content-Type:\" prefix.")},
		Returns: "dict[str, str]", Example: `net.content_type_parse("multipart/form-data; boundary=xyz")`,
		Tags:    []string{"charset", "boundary", "media type", "header", "mime parse", "content-type"},
		Accepts: func(s string) bool { return strings.Contains(s, "/") },
		Fn: func(a tools.Args) (any, error) {
			h := strings.TrimSpace(a.Str("header"))
			if i := strings.Index(strings.ToLower(h), "content-type:"); i == 0 {
				h = strings.TrimSpace(h[len("content-type:"):])
			}
			mt, params, err := mime.ParseMediaType(h)
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Type %q: %v", h, err)
			}
			typ, sub, found := strings.Cut(mt, "/")
			if !found || typ == "" || sub == "" {
				return nil, fmt.Errorf("invalid Content-Type %q: expected type/subtype", h)
			}
			if c, ok := params["charset"]; ok {
				params["charset"] = strings.ToLower(c)
			}
			out := map[string]string{"media_type": mt, "type": typ, "subtype": sub, "charset": params["charset"], "boundary": params["boundary"]}
			if i := strings.LastIndex(sub, "+"); i >= 0 {
				out["suffix"] = sub[i+1:]
			}
			for k, v := range params {
				out["param_"+k] = v
			}
			return out, nil
		},
	})
}

func containsAll(hay string, words []string) bool {
	for _, w := range words {
		if !strings.Contains(hay, w) {
			return false
		}
	}
	return len(words) > 0
}

func findMethod(s string) *methodT {
	s = strings.ToUpper(strings.TrimSpace(s))
	for i := range methodTable {
		if methodTable[i].name == s {
			return &methodTable[i]
		}
	}
	return nil
}

func extOf(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if i := strings.LastIndexAny(s, `/\`); i >= 0 {
		s = s[i+1:]
	}
	if i := strings.LastIndex(s, "."); i >= 0 {
		s = s[i+1:]
	}
	return s
}

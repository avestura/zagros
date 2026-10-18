package netx

import (
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "net") }

func call(t *testing.T, name string, kv ...any) any {
	t.Helper()
	m := map[string]any{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	v, err := tools.Call("net", name, m)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return v
}

func fails(t *testing.T, name string, kv ...any) {
	t.Helper()
	m := map[string]any{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	if _, err := tools.Call("net", name, m); err == nil {
		t.Errorf("%s %v: expected error", name, kv)
	}
}

func eq(t *testing.T, got any, want string) {
	t.Helper()
	if g := tools.Format(got); g != want {
		t.Errorf("got %q, want %q", g, want)
	}
}

func TestCIDRInfo(t *testing.T) {
	m := call(t, "cidr_info", "cidr", "192.168.1.130/26").(map[string]string)
	want := map[string]string{
		"network": "192.168.1.128", "broadcast": "192.168.1.191", "netmask": "255.255.255.192", "wildcard": "0.0.0.63",
		"first_host": "192.168.1.129", "last_host": "192.168.1.190", "usable_hosts": "62", "total_addresses": "64",
		"ip_class": "C", "type": "private", "prefix": "26", "binary_netmask": "11111111.11111111.11111111.11000000",
		"hex_netmask": "0xffffffc0", "reverse_zone": "128/26.1.168.192.in-addr.arpa",
	}
	for k, v := range want {
		if m[k] != v {
			t.Errorf("%s = %q, want %q", k, m[k], v)
		}
	}
	// alternate spellings
	for _, in := range []string{"10.0.0.5 255.255.255.0", "10.0.0.5 0.0.0.255", "10.0.0.5/255.255.255.0", "10.0.0.5/24"} {
		if c := call(t, "cidr_info", "cidr", in).(map[string]string)["cidr"]; c != "10.0.0.0/24" {
			t.Errorf("%s -> %s", in, c)
		}
	}
	h := call(t, "cidr_info", "cidr", "8.8.8.8").(map[string]string)
	if h["prefix"] != "32" || h["usable_hosts"] != "1" || h["type"] != "public" {
		t.Errorf("host: %v", h)
	}
	p := call(t, "cidr_info", "cidr", "10.0.0.0/31").(map[string]string)
	if p["usable_hosts"] != "2" || p["first_host"] != "10.0.0.0" || p["last_host"] != "10.0.0.1" {
		t.Errorf("/31: %v", p)
	}
	v6 := call(t, "cidr_info", "cidr", "2001:db8::/32").(map[string]string)
	if v6["total_addresses"] != "79228162514264337593543950336" || v6["last_address"] != "2001:db8:ffff:ffff:ffff:ffff:ffff:ffff" || v6["type"] != "documentation" {
		t.Errorf("v6: %v", v6)
	}
	if v6["reverse_zone"] != "8.b.d.0.1.0.0.2.ip6.arpa" {
		t.Errorf("v6 zone %s", v6["reverse_zone"])
	}
	z := call(t, "cidr_info", "cidr", "0.0.0.0/0").(map[string]string)
	if z["total_addresses"] != "4294967296" || z["reverse_zone"] != "in-addr.arpa" {
		t.Errorf("/0: %v", z)
	}
	fails(t, "cidr_info", "cidr", "10.0.0.0/33")
	fails(t, "cidr_info", "cidr", "10.0.0.0 255.0.255.0")
	fails(t, "cidr_info", "cidr", "garbage")
	fails(t, "cidr_info", "cidr", "")
}

func TestCIDROps(t *testing.T) {
	eq(t, call(t, "cidr_contains", "cidr", "10.0.0.0/8", "ip", "10.20.30.40"), "true")
	eq(t, call(t, "cidr_contains", "cidr", "10.0.0.0/8", "ip", "11.0.0.1"), "false")
	eq(t, call(t, "cidr_contains", "cidr", "10.0.0.0/8", "ip", "10.1.0.0/16"), "true")
	eq(t, call(t, "cidr_contains", "cidr", "10.0.0.0/8", "ip", "::1"), "false")
	eq(t, call(t, "cidr_overlap", "a", "10.0.0.0/16", "b", "10.0.5.0/24"), "true")
	eq(t, call(t, "cidr_overlap", "a", "10.0.0.0/16", "b", "10.1.0.0/24"), "false")
	eq(t, call(t, "cidr_subnets", "cidr", "192.168.0.0/24", "new_prefix", int64(26)), "192.168.0.0/26\n192.168.0.64/26\n192.168.0.128/26\n192.168.0.192/26")
	eq(t, call(t, "cidr_subnets", "cidr", "10.0.0.0/8", "new_prefix", int64(24), "limit", int64(2)), "10.0.0.0/24\n10.0.1.0/24")
	fails(t, "cidr_subnets", "cidr", "10.0.0.0/24", "new_prefix", int64(16))
	eq(t, call(t, "cidr_split", "cidr", "10.0.0.0/24", "parts", int64(2)), "10.0.0.0/25\n10.0.0.128/25")
	fails(t, "cidr_split", "cidr", "10.0.0.0/24", "parts", int64(3))
	eq(t, call(t, "cidr_supernet", "cidr", "192.168.5.0/24", "new_prefix", int64(16)), "192.168.0.0/16")
	eq(t, call(t, "cidr_summarize", "items", []string{"10.0.0.0/25", "10.0.0.128/25", "10.0.1.0/24"}), "10.0.0.0/23")
	eq(t, call(t, "cidr_summarize", "items", []string{"10.0.0.1-10.0.0.6", "10.0.0.7", "2001:db8::/33, 2001:db8:8000::/33"}), "10.0.0.1/32\n10.0.0.2/31\n10.0.0.4/30\n2001:db8::/32")
	eq(t, call(t, "cidr_exclude", "cidr", "10.0.0.0/24", "exclude", "10.0.0.64/26"), "10.0.0.0/26\n10.0.0.128/25")
	eq(t, call(t, "cidr_exclude", "cidr", "10.0.0.0/24", "exclude", "10.0.0.0/16"), "")
	eq(t, call(t, "cidr_exclude", "cidr", "10.0.0.0/24", "exclude", "192.168.0.0/16"), "10.0.0.0/24")
	eq(t, call(t, "range_to_cidrs", "start", "10.0.0.5", "end", "10.0.0.20"), "10.0.0.5/32\n10.0.0.6/31\n10.0.0.8/29\n10.0.0.16/30\n10.0.0.20/32")
	eq(t, call(t, "range_to_cidrs", "start", "0.0.0.0", "end", "255.255.255.255"), "0.0.0.0/0")
	eq(t, call(t, "cidr_to_range", "cidr", "192.168.1.0/24"), "192.168.1.0 - 192.168.1.255")
	eq(t, call(t, "cidr_hosts", "cidr", "192.168.1.0/29"), "192.168.1.1\n192.168.1.2\n192.168.1.3\n192.168.1.4\n192.168.1.5\n192.168.1.6")
	eq(t, call(t, "cidr_hosts", "cidr", "255.255.255.255/32"), "255.255.255.255")
	eq(t, call(t, "cidr_next", "cidr", "10.0.0.0/24"), "10.0.1.0/24")
	fails(t, "cidr_next", "cidr", "255.255.255.0/24")
	eq(t, call(t, "cidr_count", "cidr", "10.0.0.0/22"), "1024")
	eq(t, call(t, "cidr_count", "cidr", "10.0.0.0/22", "usable", true), "1022")
	eq(t, call(t, "mask_to_prefix", "mask", "255.255.240.0"), "20")
	eq(t, call(t, "mask_to_prefix", "mask", "0.0.0.255"), "24")
	fails(t, "mask_to_prefix", "mask", "255.0.255.0")
	eq(t, call(t, "wildcard_to_prefix", "wildcard", "0.0.15.255"), "20")
	eq(t, call(t, "prefix_to_mask", "prefix", "/20"), "255.255.240.0")
	eq(t, call(t, "prefix_to_mask", "prefix", "20", "wildcard", true), "0.0.15.255")
	eq(t, call(t, "prefix_to_mask", "prefix", "64", "version", int64(6)), "ffff:ffff:ffff:ffff::")
	fails(t, "prefix_to_mask", "prefix", "40")
	s := call(t, "subnet_for_hosts", "hosts", "50").(map[string]string)
	if s["prefix"] != "26" || s["usable_hosts"] != "62" {
		t.Errorf("subnet_for_hosts: %v", s)
	}
	s = call(t, "subnet_for_hosts", "hosts", "2", "base", "10.0.0.0/24").(map[string]string)
	if s["prefix"] != "30" || s["first_subnet"] != "10.0.0.0/30" || s["subnets_available"] != "64" {
		t.Errorf("with base: %v", s)
	}
	v := call(t, "vlsm", "base", "192.168.0.0/24", "hosts", []string{"LAN-B:50", "LAN-A:100", "WAN:2"}).([]string)
	if len(v) != 4 || !strings.HasPrefix(v[1], "LAN-A | 100 | /25 | 192.168.0.0/25") || !strings.HasPrefix(v[2], "LAN-B | 50 | /26 | 192.168.0.128/26") || !strings.HasPrefix(v[3], "WAN | 2 | /30 | 192.168.0.192/30") {
		t.Errorf("vlsm: %v", v)
	}
	fails(t, "vlsm", "base", "192.168.0.0/25", "hosts", []string{"100", "100"})
	if tb := call(t, "subnet_table", "cidr", "10.0.0.0/24").([]string); len(tb) != 10 || !strings.HasPrefix(tb[1], "/24 | 255.255.255.0 | 1 | 256 | 254") {
		t.Errorf("table: %v", tb)
	}
}

func TestIPTools(t *testing.T) {
	eq(t, call(t, "ip_to_int", "ip", "192.168.1.1"), "3232235777")
	eq(t, call(t, "ip_to_int", "ip", "::1"), "1")
	eq(t, call(t, "int_to_ip", "number", "3232235777"), "192.168.1.1")
	eq(t, call(t, "int_to_ip", "number", "0xC0A80101"), "192.168.1.1")
	eq(t, call(t, "int_to_ip", "number", "1", "version", int64(6)), "::1")
	eq(t, call(t, "int_to_ip", "number", "4294967296"), "::1:0:0")
	fails(t, "int_to_ip", "number", "-1")
	fails(t, "int_to_ip", "number", "99999999999", "version", int64(4))
	eq(t, call(t, "ip_to_hex", "ip", "192.168.1.1"), "0xc0a80101")
	eq(t, call(t, "ip_to_binary", "ip", "192.168.1.1"), "11000000.10101000.00000001.00000001")
	eq(t, call(t, "binary_to_ip", "bits", "11000000.10101000.00000001.00000001"), "192.168.1.1")
	eq(t, call(t, "hex_to_ip", "hex", "0xC0A80101"), "192.168.1.1")
	fails(t, "hex_to_ip", "hex", "C0A801")
	eq(t, call(t, "ip_version", "ip", "2001:db8::1"), "6")
	fails(t, "ip_version", "ip", "1.2.3")
	types := map[string]string{
		"10.1.2.3": "private", "172.16.0.1": "private", "172.32.0.1": "public", "192.168.0.1": "private", "100.64.0.1": "cgnat",
		"127.0.0.1": "loopback", "169.254.1.1": "link-local", "224.0.0.1": "multicast", "0.0.0.0": "unspecified",
		"192.0.2.1": "documentation", "198.18.0.1": "benchmark", "255.255.255.255": "broadcast", "8.8.8.8": "public",
		"::1": "loopback", "::": "unspecified", "fe80::1": "link-local", "fc00::1": "ula", "fd12::1": "ula", "ff02::1": "multicast",
		"2001:db8::1": "documentation", "2002:c000:204::1": "6to4", "2001:0::1": "teredo", "::ffff:1.2.3.4": "ipv4-mapped",
		"2606:4700::1111": "public", "64:ff9b::808:808": "nat64", "240.0.0.1": "reserved",
	}
	for ip, want := range types {
		if got := call(t, "ip_type", "ip", ip); got != want {
			t.Errorf("ip_type(%s) = %v, want %s", ip, got, want)
		}
	}
	c := call(t, "ip_classify", "ip", "ff02::1").(map[string]string)
	if c["scope"] != "link-local" || c["is_multicast"] != "true" {
		t.Errorf("classify: %v", c)
	}
	eq(t, call(t, "ip_normalize", "ip", "2001:0DB8:0000:0000:0000:0000:0000:0001"), "2001:db8::1")
	eq(t, call(t, "ip_normalize", "ip", "010.001.000.001"), "10.1.0.1")
	eq(t, call(t, "ip_normalize", "ip", "127.1"), "127.0.0.1")
	eq(t, call(t, "ip_normalize", "ip", "[::1]"), "::1")
	eq(t, call(t, "ip_normalize", "ip", "0x7f000001"), "127.0.0.1")
	eq(t, call(t, "ipv6_expand", "ip", "2001:db8::1"), "2001:0db8:0000:0000:0000:0000:0000:0001")
	eq(t, call(t, "ipv6_expand", "ip", "::"), "0000:0000:0000:0000:0000:0000:0000:0000")
	eq(t, call(t, "ipv6_compress", "ip", "2001:0db8:0000:0000:0000:0000:0000:0001"), "2001:db8::1")
	eq(t, call(t, "ipv6_compress", "ip", "2001:db8:0:0:1:0:0:1"), "2001:db8::1:0:0:1")
	eq(t, call(t, "ipv4_to_ipv6_mapped", "ip", "192.0.2.1"), "::ffff:192.0.2.1")
	eq(t, call(t, "ipv6_to_ipv4_mapped", "ip", "::ffff:c000:201"), "192.0.2.1")
	fails(t, "ipv6_to_ipv4_mapped", "ip", "2001:db8::1")
	eq(t, call(t, "ipv4_to_6to4", "ip", "192.0.2.4"), "2002:c000:204::")
	eq(t, call(t, "sixtofour_to_ipv4", "ip", "2002:c000:204::1"), "192.0.2.4")
	eq(t, call(t, "nat64", "ip", "192.0.2.33"), "64:ff9b::c000:221")
	eq(t, call(t, "nat64", "ip", "192.0.2.33", "prefix", "2001:db8::/32"), "2001:db8:c000:221::")
	eq(t, call(t, "nat64", "ip", "192.0.2.33", "prefix", "2001:db8:122:344::/64"), "2001:db8:122:344:c0:2:2100:0")
	eq(t, call(t, "nat64_to_ipv4", "ip", "2001:db8:122:344:c0:2:2100:0", "prefix", "2001:db8:122:344::/64"), "192.0.2.33")
	eq(t, call(t, "eui64_from_mac", "mac", "00:1a:2b:3c:4d:5e", "prefix", "2001:db8::"), "2001:db8::21a:2bff:fe3c:4d5e")
	eq(t, call(t, "link_local_from_mac", "mac", "00:1A:2B:3C:4D:5E"), "fe80::21a:2bff:fe3c:4d5e")
	eq(t, call(t, "ptr_name", "ip", "192.0.2.10"), "10.2.0.192.in-addr.arpa")
	eq(t, call(t, "ptr_name", "ip", "2001:db8::567:89ab"), "b.a.9.8.7.6.5.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.0.8.b.d.0.1.0.0.2.ip6.arpa")
	eq(t, call(t, "ip_compare", "a", "10.0.0.9", "b", "10.0.0.10"), "-1")
	eq(t, call(t, "sort_ips", "ips", []string{"10.0.0.10", "10.0.0.9", "::1", "10.0.0.9"}, "unique", true), "10.0.0.9\n10.0.0.10\n::1")
	eq(t, call(t, "sort_ips", "ips", []string{"1.1.1.1", "2.2.2.2"}, "descending", true), "2.2.2.2\n1.1.1.1")
	eq(t, call(t, "ip_in_ranges", "ip", "10.1.2.3", "ranges", []string{"192.168.0.0/16", "10.0.0.0/8"}), "true")
	eq(t, call(t, "ip_in_ranges", "ip", "10.1.2.3", "ranges", []string{"10.1.2.1-10.1.2.2"}), "false")
	eq(t, call(t, "ip_matching_ranges", "ip", "10.1.2.3", "ranges", []string{"10.0.0.0/8", "10.1.0.0/16", "172.16.0.0/12"}), "10.0.0.0/8\n10.1.0.0/16")
	eq(t, call(t, "ip_add", "ip", "10.0.0.255", "n", int64(1)), "10.0.1.0")
	eq(t, call(t, "ip_add", "ip", "10.0.1.0", "n", int64(-1)), "10.0.0.255")
	fails(t, "ip_add", "ip", "255.255.255.255", "n", int64(1))
	eq(t, call(t, "ip_diff", "a", "10.0.0.1", "b", "10.0.1.1"), "256")
	eq(t, call(t, "is_valid_ip", "ip", "256.1.1.1"), "false")
	eq(t, call(t, "is_valid_ipv4", "ip", "1.2.3.4"), "true")
	eq(t, call(t, "is_valid_ipv6", "ip", "1.2.3.4"), "false")
	eq(t, call(t, "is_valid_cidr", "cidr", "10.0.0.0/33"), "false")
	eq(t, call(t, "is_valid_cidr", "cidr", "2001:db8::/32"), "true")
	eq(t, call(t, "ipv4_octets", "ip", "192.168.1.1"), "192\n168\n1\n1")
	f := call(t, "ip_formats", "ip", "192.168.1.1").(map[string]string)
	if f["decimal"] != "3232235777" || f["url_decimal"] != "http://3232235777/" || f["octal"] != "0300.0250.0001.0001" || f["ipv6_mapped"] != "::ffff:192.168.1.1" || f["url_hex"] != "http://0xc0a80101/" {
		t.Errorf("formats: %v", f)
	}
}

func TestMAC(t *testing.T) {
	in := "00-1A-2B-3C-4D-5E"
	eq(t, call(t, "mac_normalize", "mac", in), "00:1a:2b:3c:4d:5e")
	eq(t, call(t, "mac_normalize", "mac", in, "format", "dot"), "001a.2b3c.4d5e")
	eq(t, call(t, "mac_normalize", "mac", in, "format", "plain", "upper", true), "001A2B3C4D5E")
	eq(t, call(t, "mac_normalize", "mac", "001a.2b3c.4d5e", "format", "upper"), "00:1A:2B:3C:4D:5E")
	eq(t, call(t, "mac_normalize", "mac", in, "format", "dash"), "00-1a-2b-3c-4d-5e")
	fails(t, "mac_normalize", "mac", "00:1A:2B:3C:4D")
	fails(t, "mac_normalize", "mac", in, "format", "x")
	eq(t, call(t, "mac_is_valid", "mac", "zz:1A:2B:3C:4D:5E"), "false")
	m := call(t, "mac_info", "mac", "01:00:5e:00:00:fb").(map[string]string)
	if m["cast"] != "multicast" || m["administration"] != "universally administered (UAA)" {
		t.Errorf("info: %v", m)
	}
	m = call(t, "mac_info", "mac", "02:42:ac:11:00:02").(map[string]string)
	if m["administration"] != "locally administered (LAA)" || m["vendor"] == "" {
		t.Errorf("laa: %v", m)
	}
	eq(t, call(t, "oui_vendor", "mac", "B8:27:EB:12:34:56"), "Raspberry Pi Foundation")
	eq(t, call(t, "oui_vendor", "mac", "00:50:56"), "VMware")
	eq(t, call(t, "oui_vendor", "mac", "fe:ed:fa:ce:00:01"), "")
	eq(t, call(t, "oui_vendor", "mac", "ff:ff:ff:ff:ff:ff"), "Broadcast")
	if len(ouiTable) < 250 {
		t.Errorf("oui table too small: %d", len(ouiTable))
	}
}

func TestPortsAndProtocols(t *testing.T) {
	p := call(t, "port_service", "port_or_name", "443").(map[string]string)
	if p["name"] != "https" || p["range"] != "well-known" || !strings.Contains(p["protocol"], "tcp") {
		t.Errorf("443: %v", p)
	}
	p = call(t, "port_service", "port_or_name", "ssh").(map[string]string)
	if p["port"] != "22" {
		t.Errorf("ssh: %v", p)
	}
	p = call(t, "port_service", "port_or_name", "55555").(map[string]string)
	if p["range"] != "dynamic" || p["name"] != "" {
		t.Errorf("55555: %v", p)
	}
	fails(t, "port_service", "port_or_name", "70000x y")
	eq(t, call(t, "service_ports", "name", "domain"), "53/tcp\n53/udp")
	eq(t, call(t, "service_ports", "name", "dns"), "53/tcp\n53/udp")
	eq(t, call(t, "service_ports", "name", "zzzz"), "")
	eq(t, call(t, "port_range_type", "port", "8080"), "registered")
	eq(t, call(t, "port_range_type", "port", "0"), "reserved")
	fails(t, "port_range_type", "port", "65536")
	eq(t, call(t, "protocol_number", "name_or_number", "tcp").(map[string]string)["number"], "6")
	eq(t, call(t, "protocol_number", "name_or_number", "47").(map[string]string)["keyword"], "GRE")
	fails(t, "protocol_number", "name_or_number", "nope")
	i := call(t, "icmp_type", "spec", "3/4").(map[string]string)
	if i["type_name"] != "Destination Unreachable" || i["code_name"] != "Fragmentation Needed and DF Set" {
		t.Errorf("icmp: %v", i)
	}
	i = call(t, "icmp_type", "spec", "135", "version", int64(6)).(map[string]string)
	if i["type_name"] != "Neighbor Solicitation" {
		t.Errorf("icmp6: %v", i)
	}
	fails(t, "icmp_type", "spec", "abc")
	h := call(t, "http_status", "code", "404").(map[string]string)
	if h["text"] != "Not Found" || h["class"] != "4xx Client Error" {
		t.Errorf("404: %v", h)
	}
	if call(t, "http_status", "code", "599").(map[string]string)["registry"] != "unassigned" {
		t.Error("599")
	}
	fails(t, "http_status", "code", "99")
	if !strings.Contains(tools.Format(call(t, "http_status_search", "text", "redirect")), "307 Temporary Redirect") {
		t.Error("search redirect")
	}
	eq(t, call(t, "http_status_search", "text", "41"), "410 Gone\n411 Length Required\n412 Precondition Failed\n413 Content Too Large\n414 URI Too Long\n415 Unsupported Media Type\n416 Range Not Satisfiable\n417 Expectation Failed\n418 I'm a teapot")
	mi := call(t, "http_method_info", "method", "put").(map[string]string)
	if mi["idempotent"] != "true" || mi["safe"] != "false" {
		t.Errorf("PUT: %v", mi)
	}
	fails(t, "http_method_info", "method", "FETCH")
	eq(t, call(t, "mime_type", "extension_or_filename", "photo.JPEG"), "image/jpeg")
	eq(t, call(t, "mime_type", "extension_or_filename", "archive.tar.gz"), "application/gzip")
	eq(t, call(t, "mime_type", "extension_or_filename", ".json"), "application/json")
	eq(t, call(t, "mime_type", "extension_or_filename", "x.unknownext"), "application/octet-stream")
	eq(t, call(t, "extension_for_mime", "mime_type", "image/jpeg; q=1"), ".jpg\n.jpeg\n.jpe")
	c := call(t, "content_type_parse", "header", "Content-Type: text/html; charset=UTF-8").(map[string]string)
	if c["type"] != "text" || c["subtype"] != "html" || c["charset"] != "utf-8" {
		t.Errorf("ct: %v", c)
	}
	c = call(t, "content_type_parse", "header", "application/ld+json").(map[string]string)
	if c["suffix"] != "json" {
		t.Errorf("suffix: %v", c)
	}
	fails(t, "content_type_parse", "header", "garbage")
}

func TestURLs(t *testing.T) {
	u := call(t, "url_parse", "url", "https://user:pw@Example.com:8443/a/b.html?x=1&y=2#top").(map[string]string)
	want := map[string]string{"scheme": "https", "user": "user", "password": "pw", "hostname": "Example.com", "port": "8443", "path": "/a/b.html",
		"query": "x=1&y=2", "fragment": "top", "origin": "https://example.com:8443", "effective_port": "8443", "file_name": "b.html"}
	for k, v := range want {
		if u[k] != v {
			t.Errorf("%s = %q want %q", k, u[k], v)
		}
	}
	u = call(t, "url_parse", "url", "example.com/path?q=1").(map[string]string)
	if u["host"] != "example.com" || u["path"] != "/path" {
		t.Errorf("schemeless: %v", u)
	}
	u = call(t, "url_parse", "url", "http://[::1]:8080/x").(map[string]string)
	if u["hostname"] != "::1" || u["host_type"] != "ipv6" || u["port"] != "8080" {
		t.Errorf("v6: %v", u)
	}
	u = call(t, "url_parse", "url", "https://example.com").(map[string]string)
	if u["effective_port"] != "443" || u["port"] != "" {
		t.Errorf("default port: %v", u)
	}
	u = call(t, "url_parse", "url", "localhost:3000/app").(map[string]string)
	if u["hostname"] != "localhost" || u["port"] != "3000" {
		t.Errorf("localhost: %v", u)
	}
	u = call(t, "url_parse", "url", "mailto:a@b.com").(map[string]string)
	if u["scheme"] != "mailto" || u["opaque"] != "a@b.com" {
		t.Errorf("mailto: %v", u)
	}
	fails(t, "url_parse", "url", "http://exa mple.com")
	eq(t, call(t, "url_build", "parts", map[string]string{"scheme": "https", "host": "example.com", "path": "/a b", "query": "x=1", "fragment": "f"}), "https://example.com/a%20b?x=1#f")
	eq(t, call(t, "url_build", "parts", map[string]string{"host": "::1", "port": "80"}), "https://[::1]:80")
	eq(t, call(t, "url_join", "base", "https://example.com/a/b/c.html", "ref", "../img/x.png"), "https://example.com/a/img/x.png")
	eq(t, call(t, "url_join", "base", "https://example.com/a/", "ref", "//cdn.com/x"), "https://cdn.com/x")
	eq(t, call(t, "url_normalize", "url", "HTTP://Example.COM:80/a/./b/../c?b=2&a=1", "sort_query", true), "http://example.com/a/c?a=1&b=2")
	eq(t, call(t, "url_normalize", "url", "https://example.com"), "https://example.com/")
	eq(t, call(t, "url_normalize", "url", "https://example.com/%7euser/%e2%82%ac#"), "https://example.com/~user/%E2%82%AC")
	eq(t, call(t, "url_normalize", "url", "https://example.com/p#frag", "strip_fragment", true), "https://example.com/p")
	q := call(t, "url_query", "url", "https://example.com/s?q=go+lang&tag=a&tag=b&flag#h").(map[string]string)
	if q["q"] != "go lang" || q["tag"] != "a,b" || q["flag"] != "" {
		t.Errorf("query: %v", q)
	}
	if q := call(t, "url_query", "url", "a=1&b=%20x").(map[string]string); q["a"] != "1" || q["b"] != " x" {
		t.Errorf("bare query: %v", q)
	}
	eq(t, call(t, "url_set_query", "url", "https://e.com/s?old=1", "params", map[string]string{"q": "go lang", "page": "2"}), "https://e.com/s?page=2&q=go+lang")
	eq(t, call(t, "url_add_param", "url", "https://e.com/s?q=1", "key", "page", "value", "2"), "https://e.com/s?q=1&page=2")
	eq(t, call(t, "url_set_param", "url", "https://e.com/s?q=1&page=1&page=3", "key", "page", "value", "5"), "https://e.com/s?q=1&page=5")
	eq(t, call(t, "url_remove_param", "url", "https://e.com/?a=1&utm_x=2&b=3", "keys", []string{"utm_*"}), "https://e.com/?a=1&b=3")
	eq(t, call(t, "url_remove_param", "url", "https://e.com/?a=1", "keys", []string{"a"}), "https://e.com/")
	eq(t, call(t, "url_strip_tracking", "url", "https://e.com/p?id=7&utm_source=x&fbclid=abc&gclid=1#sec"), "https://e.com/p?id=7#sec")
	eq(t, call(t, "url_strip_tracking", "url", "e.com/p?utm_medium=x&q=a%20b"), "e.com/p?q=a%20b")
	eq(t, call(t, "url_origin", "url", "https://Example.com:443/a?b=1"), "https://example.com")
	fails(t, "url_origin", "url", "/just/a/path")
	for in, want := range map[string]string{"https://example.com/path": "true", "example.com": "false", "http://": "false", "http://a b.com": "false",
		"http://localhost:70000": "false", "http://127.0.0.1:80/x": "true", "mailto:a@b.com": "true", "http://-bad.com": "false"} {
		if got := tools.Format(call(t, "url_is_valid", "url", in)); got != want {
			t.Errorf("url_is_valid(%q) = %s want %s", in, got, want)
		}
	}
}

func TestDomains(t *testing.T) {
	d := call(t, "domain_parts", "host", "https://blog.example.co.uk/post").(map[string]string)
	if d["public_suffix"] != "co.uk" || d["registrable_domain"] != "example.co.uk" || d["subdomain"] != "blog" || d["domain"] != "example" || d["suffix_type"] != "icann" || d["tld"] != "uk" {
		t.Errorf("co.uk: %v", d)
	}
	d = call(t, "domain_parts", "host", "user.github.io").(map[string]string)
	if d["suffix_type"] != "private" || d["registrable_domain"] != "user.github.io" {
		t.Errorf("github.io: %v", d)
	}
	d = call(t, "domain_parts", "host", "co.uk").(map[string]string)
	if d["registrable_domain"] != "" {
		t.Errorf("bare suffix: %v", d)
	}
	d = call(t, "domain_parts", "host", "foo.localdomain").(map[string]string)
	if d["suffix_type"] != "unknown" {
		t.Errorf("unknown tld: %v", d)
	}
	d = call(t, "domain_parts", "host", "www.bücher.de").(map[string]string)
	if d["host"] != "www.xn--bcher-kva.de" || d["unicode_host"] != "www.bücher.de" {
		t.Errorf("idn: %v", d)
	}
	fails(t, "domain_parts", "host", "192.168.1.1")
	for in, want := range map[string]string{"my-server.local": "true", "-bad.com": "false", "a..b": "false", "localhost": "true", "exa_mple.com": "false", "": "false", "a.b.": "true"} {
		if got := tools.Format(call(t, "is_valid_hostname", "host", in)); got != want {
			t.Errorf("is_valid_hostname(%q)=%s want %s", in, got, want)
		}
	}
	eq(t, call(t, "is_valid_hostname", "host", "_dmarc.example.com", "allow_underscore", true), "true")
	for in, want := range map[string]string{"example.com": "true", "localhost": "false", "1.2.3.4": "false", "a.b.123": "false", "münchen.de": "true", "a.c": "false"} {
		if got := tools.Format(call(t, "is_valid_domain", "host", in)); got != want {
			t.Errorf("is_valid_domain(%q)=%s want %s", in, got, want)
		}
	}
	e := call(t, "email_parse", "email", "John.Doe+news@Gmail.com").(map[string]string)
	if e["local"] != "John.Doe+news" || e["tag"] != "news" || e["normalized"] != "johndoe@gmail.com" || e["domain"] != "gmail.com" {
		t.Errorf("email: %v", e)
	}
	e = call(t, "email_parse", "email", "Jane <Jane@Example.org>").(map[string]string)
	if e["display_name"] != "Jane" || e["normalized"] != "jane@example.org" {
		t.Errorf("display: %v", e)
	}
	fails(t, "email_parse", "email", "nope")
	for in, want := range map[string]string{"a@b.co": "true", "a.b+c@d.example.com": "true", "a@b": "false", ".a@b.co": "false", "a..b@c.co": "false",
		"a b@c.co": "false", "@b.co": "false", "a@-b.co": "false", "\"a b\"@c.co": "true", "a@[192.168.0.1]": "true", "a@b.co ": "false"} {
		if got := tools.Format(call(t, "is_valid_email", "email", in)); got != want {
			t.Errorf("is_valid_email(%q)=%s want %s", in, got, want)
		}
	}
	eq(t, call(t, "reverse_domain_name", "host", "a.b.com"), "com.b.a")
	h := call(t, "basic_auth_header", "user", "alice", "password", "s3cret")
	eq(t, h, "Basic YWxpY2U6czNjcmV0")
	eq(t, call(t, "parse_basic_auth", "header", "Authorization: Basic YWxpY2U6czNjcmV0").(map[string]string)["password"], "s3cret")
	fails(t, "parse_basic_auth", "header", "Basic !!!")
	fails(t, "basic_auth_header", "user", "a:b", "password", "x")
	for in, want := range map[string][3]string{"[::1]:8080": {"::1", "8080", "ipv6"}, "example.com:80": {"example.com", "80", "domain"}, "::1": {"::1", "", "ipv6"}, "10.0.0.1": {"10.0.0.1", "", "ipv4"}, "[2001:db8::1]": {"2001:db8::1", "", "ipv6"}} {
		m := call(t, "parse_host_port", "address", in).(map[string]string)
		if m["host"] != want[0] || m["port"] != want[1] || m["host_type"] != want[2] {
			t.Errorf("parse_host_port(%s) = %v", in, m)
		}
	}
	fails(t, "parse_host_port", "address", "example.com:99999")
	fails(t, "parse_host_port", "address", "[::1")
	eq(t, call(t, "hosts_entry", "ip", "127.0.0.1", "names", []string{"app.local", "www.app.local"}), "127.0.0.1\tapp.local www.app.local")
	fails(t, "hosts_entry", "ip", "127.0.0.1", "names", []string{"bad name"})
}

func TestBandwidth(t *testing.T) {
	eq(t, call(t, "bandwidth", "value", 100.0, "from_unit", "Mbps", "to_unit", "MB/s"), "12.5")
	eq(t, call(t, "bandwidth", "value", 1.0, "from_unit", "GB/s", "to_unit", "Gbps"), "8")
	eq(t, call(t, "bandwidth", "value", 1.0, "from_unit", "MiB/s", "to_unit", "bps"), "8388608")
	eq(t, call(t, "bandwidth", "value", 8.0, "from_unit", "bps", "to_unit", "Bps"), "1")
	fails(t, "bandwidth", "value", 1.0, "from_unit", "furlongs", "to_unit", "bps")
	eq(t, call(t, "transfer_time", "size", "1 GB", "bandwidth", "80 Mbps"), "1m 40s")
	eq(t, call(t, "transfer_time", "size", "100 MB", "bandwidth", "100 MB/s"), "1 s")
	eq(t, call(t, "transfer_time", "size", "1 TB", "bandwidth", "1 Gbps"), "2h 13m 20s")
	fails(t, "transfer_time", "size", "1 GB", "bandwidth", "0 Mbps")
	fails(t, "transfer_time", "size", "lots", "bandwidth", "1 Mbps")
	m := call(t, "throughput_needed", "size", "10 GB", "seconds", 80.0).(map[string]string)
	if m["Mbps"] != "1000" || m["MB/s"] != "125" {
		t.Errorf("throughput: %v", m)
	}
	fails(t, "throughput_needed", "size", "10 GB", "seconds", 0.0)
}

func TestGarbageDoesNotPanic(t *testing.T) {
	junk := []string{"", " ", "\x00", "::::", "1.2.3.4.5", "/", "999999999999999999999", "http://", "[", "%zz", "a@", "@", strings.Repeat("a", 5000), "💥/💥"}
	for _, f := range tools.All() {
		if f.Group != "net" {
			continue
		}
		for _, j := range junk {
			named := map[string]any{}
			for _, p := range f.Params {
				switch p.Type {
				case "str":
					named[p.Name] = j
				case "int":
					named[p.Name] = int64(-5)
				case "float":
					named[p.Name] = -1.5
				case "bool":
					named[p.Name] = true
				case "list[str]":
					named[p.Name] = []string{j, j}
				case "dict[str, str]":
					named[p.Name] = map[string]string{j: j}
				}
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s(%q) panicked: %v", f.Key(), j, r)
					}
				}()
				f.Invoke(named)
				if f.Accepts != nil {
					f.Accepts(j)
				}
			}()
		}
	}
}

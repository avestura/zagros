package netx

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

// portData: port|protocols|service name|description. A partial table of
// well-known and registered ports (IANA and common practice).
const portData = `
1|tcp|tcpmux|TCP Port Service Multiplexer
7|tcp,udp|echo|Echo protocol
9|tcp,udp|discard|Discard protocol
13|tcp,udp|daytime|Daytime protocol
17|tcp|qotd|Quote of the Day
19|tcp,udp|chargen|Character Generator protocol
20|tcp|ftp-data|FTP data transfer
21|tcp|ftp|File Transfer Protocol control
22|tcp|ssh|Secure Shell, SFTP and SCP
23|tcp|telnet|Telnet remote login (unencrypted)
25|tcp|smtp|Simple Mail Transfer Protocol (server to server)
37|tcp,udp|time|Time protocol
42|tcp,udp|nameserver|Host Name Server (WINS)
43|tcp|whois|WHOIS domain registration lookup
49|tcp,udp|tacacs|TACACS+ authentication
53|tcp,udp|domain|Domain Name System (DNS)
67|udp|dhcp-server|DHCP/BOOTP server
68|udp|dhcp-client|DHCP/BOOTP client
69|udp|tftp|Trivial File Transfer Protocol
70|tcp|gopher|Gopher protocol
79|tcp|finger|Finger user information
80|tcp|http|Hypertext Transfer Protocol
81|tcp|http-alt|Alternate HTTP / Tor / web admin
88|tcp,udp|kerberos|Kerberos authentication
95|tcp|supdup|SUPDUP terminal protocol
102|tcp|iso-tsap|ISO-TSAP / Siemens S7 (MMS)
104|tcp|dicom|DICOM medical imaging (ACR/NEMA)
110|tcp|pop3|Post Office Protocol v3
111|tcp,udp|sunrpc|ONC RPC / portmapper
113|tcp|ident|Ident authentication service
119|tcp|nntp|Network News Transfer Protocol
123|udp|ntp|Network Time Protocol
135|tcp,udp|msrpc|Microsoft RPC endpoint mapper (DCOM)
137|udp|netbios-ns|NetBIOS name service
138|udp|netbios-dgm|NetBIOS datagram service
139|tcp|netbios-ssn|NetBIOS session service (SMB over NetBIOS)
143|tcp|imap|Internet Message Access Protocol
161|udp|snmp|Simple Network Management Protocol
162|udp|snmptrap|SNMP traps
179|tcp|bgp|Border Gateway Protocol
194|tcp|irc|Internet Relay Chat
201|tcp,udp|at-rtmp|AppleTalk Routing Maintenance
389|tcp,udp|ldap|Lightweight Directory Access Protocol
427|tcp,udp|svrloc|Service Location Protocol
443|tcp,udp|https|HTTP over TLS (HTTP/3 uses UDP/QUIC)
445|tcp|microsoft-ds|SMB / CIFS file sharing
464|tcp,udp|kpasswd|Kerberos password change
465|tcp|smtps|SMTP over implicit TLS (submissions)
500|udp|isakmp|IKE / IPsec key exchange
513|tcp|login|rlogin remote login
514|udp|syslog|Syslog (UDP); TCP 514 is rsh
515|tcp|printer|Line Printer Daemon (LPD)
520|udp|rip|Routing Information Protocol
521|udp|ripng|RIPng for IPv6
524|tcp,udp|ncp|NetWare Core Protocol
530|tcp,udp|courier|Courier RPC
543|tcp|klogin|Kerberos login
544|tcp|kshell|Kerberos remote shell
546|udp|dhcpv6-client|DHCPv6 client
547|udp|dhcpv6-server|DHCPv6 server
554|tcp,udp|rtsp|Real Time Streaming Protocol
587|tcp|submission|SMTP mail submission (STARTTLS)
593|tcp|http-rpc-epmap|HTTP RPC endpoint mapper
596|tcp|smsd|SMSD
631|tcp,udp|ipp|Internet Printing Protocol (CUPS)
636|tcp,udp|ldaps|LDAP over TLS
639|tcp,udp|msdp|Multicast Source Discovery Protocol
646|tcp,udp|ldp|Label Distribution Protocol (MPLS)
647|tcp|dhcp-failover|DHCP failover
660|tcp|mac-srvr-admin|macOS Server administration
674|tcp|acap|Application Configuration Access Protocol
691|tcp|msexch-routing|Microsoft Exchange routing
692|tcp|hyperwave-isp|Hyperwave ISP
694|tcp,udp|ha-cluster|Linux-HA heartbeat
749|tcp,udp|kerberos-adm|Kerberos administration
750|tcp,udp|kerberos-iv|Kerberos version IV
751|tcp,udp|kerberos-master|Kerberos authentication
753|tcp,udp|rrh|Reverse Routing Header
754|tcp|krb5_prop|Kerberos propagation
860|tcp|iscsi|iSCSI (alternate)
873|tcp|rsync|rsync file synchronization
902|tcp|vmware-auth|VMware ESXi authentication daemon
989|tcp|ftps-data|FTP data over TLS
990|tcp|ftps|FTP control over TLS
992|tcp|telnets|Telnet over TLS
993|tcp|imaps|IMAP over TLS
994|tcp|ircs|IRC over TLS
995|tcp|pop3s|POP3 over TLS
1080|tcp|socks|SOCKS proxy
1194|tcp,udp|openvpn|OpenVPN
1234|tcp,udp|hotline|Hotline / commonly used for test servers
1241|tcp,udp|nessus|Nessus security scanner
1311|tcp|rxmon|Dell OpenManage / RxMon
1337|tcp|waste|WASTE / commonly used "leet" port
1433|tcp|ms-sql-s|Microsoft SQL Server
1434|udp|ms-sql-m|Microsoft SQL Server Browser
1512|tcp,udp|wins|Microsoft Windows Internet Name Service
1521|tcp|oracle|Oracle database listener
1522|tcp|oracle-alt|Oracle database (alternate)
1701|udp|l2tp|Layer 2 Tunneling Protocol
1720|tcp|h323|H.323 call signaling
1723|tcp|pptp|Point-to-Point Tunneling Protocol
1812|udp|radius|RADIUS authentication
1813|udp|radius-acct|RADIUS accounting
1883|tcp|mqtt|MQTT messaging
1900|udp|ssdp|Simple Service Discovery Protocol (UPnP)
1935|tcp|rtmp|Real-Time Messaging Protocol (Flash streaming)
2049|tcp,udp|nfs|Network File System
2082|tcp|cpanel|cPanel
2083|tcp|cpanel-ssl|cPanel over TLS
2086|tcp|whm|WebHost Manager
2087|tcp|whm-ssl|WebHost Manager over TLS
2181|tcp|zookeeper|Apache ZooKeeper client
2375|tcp|docker|Docker daemon (unencrypted)
2376|tcp|docker-ssl|Docker daemon (TLS)
2377|tcp|docker-swarm|Docker Swarm cluster management
2379|tcp|etcd-client|etcd client
2380|tcp|etcd-server|etcd peer
2483|tcp|oracle-ssl|Oracle database listener over TLS
2484|tcp|oracle-ssl2|Oracle database listener over TLS
2598|tcp|citrix-ica|Citrix ICA session reliability
3000|tcp|dev-server|Common web dev server (Node, Rails, Grafana)
3128|tcp|squid|Squid HTTP proxy
3260|tcp|iscsi-target|iSCSI target
3268|tcp|msft-gc|Active Directory Global Catalog
3269|tcp|msft-gc-ssl|Active Directory Global Catalog over TLS
3306|tcp|mysql|MySQL / MariaDB
3389|tcp,udp|ms-wbt-server|Remote Desktop Protocol (RDP)
3478|tcp,udp|stun|STUN / TURN NAT traversal
3690|tcp|svn|Subversion
3702|udp|ws-discovery|Web Services Dynamic Discovery
4000|tcp|remoteanything|Common development port
4040|tcp|yo-main|Spark UI / ngrok inspector
4369|tcp|epmd|Erlang Port Mapper Daemon
4444|tcp|krb524|Metasploit default handler / Kerberos 5 to 4
4500|udp|ipsec-nat-t|IPsec NAT traversal
4505|tcp|salt-publish|SaltStack publisher
4506|tcp|salt-ret|SaltStack return
4567|tcp|tram|Sinatra default / Tram
4662|tcp|edonkey|eDonkey
4789|udp|vxlan|Virtual Extensible LAN
4840|tcp|opc-ua|OPC UA
5000|tcp|upnp|Flask default / UPnP / Docker registry
5001|tcp|commplex-link|Commonly used by Synology DSM HTTPS
5060|tcp,udp|sip|Session Initiation Protocol
5061|tcp|sips|SIP over TLS
5222|tcp|xmpp-client|XMPP (Jabber) client
5223|tcp|xmpp-ssl|XMPP client over TLS / Apple push
5269|tcp|xmpp-server|XMPP server-to-server
5353|udp|mdns|Multicast DNS (Bonjour, Avahi)
5355|udp|llmnr|Link-Local Multicast Name Resolution
5432|tcp|postgresql|PostgreSQL
5433|tcp|postgresql-alt|PostgreSQL alternate / Vertica
5555|tcp|adb|Android Debug Bridge
5601|tcp|kibana|Kibana
5631|tcp|pcanywhere|pcAnywhere data
5632|udp|pcanywhere-stat|pcAnywhere status
5671|tcp|amqps|AMQP over TLS
5672|tcp|amqp|AMQP (RabbitMQ)
5683|udp|coap|Constrained Application Protocol
5900|tcp|vnc|VNC remote desktop
5938|tcp,udp|teamviewer|TeamViewer
5984|tcp|couchdb|Apache CouchDB
5985|tcp|wsman|Windows Remote Management (WinRM) HTTP
5986|tcp|wsmans|Windows Remote Management (WinRM) HTTPS
6000|tcp|x11|X Window System
6379|tcp|redis|Redis
6443|tcp|kubernetes-api|Kubernetes API server
6660|tcp|irc-alt|IRC (alternate)
6667|tcp|ircu|IRC
6697|tcp|ircs-u|IRC over TLS
6881|tcp,udp|bittorrent|BitTorrent
6969|tcp|bittorrent-tracker|BitTorrent tracker
7001|tcp|weblogic|Oracle WebLogic / Cassandra inter-node TLS
7070|tcp|realserver|RealServer / alternate web
7077|tcp|spark-master|Apache Spark master
7199|tcp|cassandra-jmx|Cassandra JMX
7474|tcp|neo4j|Neo4j HTTP
7687|tcp|bolt|Neo4j Bolt
7946|tcp,udp|docker-swarm-gossip|Docker Swarm node discovery
8000|tcp|http-alt|Common alternate HTTP / Django dev server
8008|tcp|http-alt|Alternate HTTP
8009|tcp|ajp13|Apache JServ Protocol
8080|tcp|http-proxy|Alternate HTTP / proxy / Tomcat
8081|tcp|http-alt|Alternate HTTP
8086|tcp|influxdb|InfluxDB HTTP API
8088|tcp|radan-http|Hadoop YARN / alternate HTTP
8089|tcp|splunk-mgmt|Splunk management
8090|tcp|http-alt|Alternate HTTP
8123|tcp|home-assistant|Home Assistant
8140|tcp|puppet|Puppet agent
8200|tcp|vault|HashiCorp Vault
8300|tcp|consul-server|Consul server RPC
8301|tcp,udp|consul-serf-lan|Consul Serf LAN gossip
8302|tcp,udp|consul-serf-wan|Consul Serf WAN gossip
8443|tcp|https-alt|Alternate HTTPS
8500|tcp|consul|Consul HTTP API
8600|tcp,udp|consul-dns|Consul DNS
8883|tcp|mqtt-tls|MQTT over TLS
8888|tcp|http-alt|Jupyter Notebook / alternate HTTP
9000|tcp|cslistener|SonarQube / PHP-FPM / MinIO / Portainer
9042|tcp|cassandra|Cassandra native protocol (CQL)
9090|tcp|prometheus|Prometheus / Cockpit / web admin
9092|tcp|kafka|Apache Kafka broker
9093|tcp|alertmanager|Prometheus Alertmanager
9100|tcp|jetdirect|HP JetDirect printing / Prometheus node exporter
9160|tcp|cassandra-thrift|Cassandra Thrift
9200|tcp|elasticsearch|Elasticsearch HTTP
9300|tcp|elasticsearch-transport|Elasticsearch node transport
9418|tcp|git|Git protocol (git://)
9443|tcp|https-alt|Alternate HTTPS (VMware, Portainer)
9876|tcp|rocketmq|Apache RocketMQ
9999|tcp|abyss|Common alternate / Urchin
10000|tcp|webmin|Webmin / NDMP
10050|tcp|zabbix-agent|Zabbix agent
10051|tcp|zabbix-trapper|Zabbix server trapper
11211|tcp,udp|memcached|Memcached
15672|tcp|rabbitmq-mgmt|RabbitMQ management UI
17500|tcp,udp|dropbox-lansync|Dropbox LAN sync
19132|udp|minecraft-be|Minecraft Bedrock Edition
20000|tcp|dnp3|DNP3 / Usermin
25565|tcp|minecraft|Minecraft Java Edition
27015|udp|steam|Steam / Source engine game servers
27017|tcp|mongodb|MongoDB
27018|tcp|mongodb-shard|MongoDB shard server
28017|tcp|mongodb-http|MongoDB web status page
30000|tcp|kubernetes-nodeport|Kubernetes NodePort range start
32400|tcp|plex|Plex Media Server
33434|udp|traceroute|traceroute (UDP probes start)
44818|tcp,udp|ethernet-ip|EtherNet/IP industrial protocol
47808|udp|bacnet|BACnet building automation
50000|tcp|db2|IBM DB2 / SAP
51820|udp|wireguard|WireGuard VPN
`

type portEntry struct {
	port       int
	protos     []string
	name, desc string
}

var portTable = buildPorts()

func buildPorts() []portEntry {
	var out []portEntry
	for _, line := range strings.Split(portData, "\n") {
		f := strings.SplitN(strings.TrimSpace(line), "|", 4)
		if len(f) != 4 {
			continue
		}
		n, _ := strconv.Atoi(f[0])
		out = append(out, portEntry{n, strings.Split(f[1], ","), f[2], f[3]})
	}
	return out
}

func portRangeType(p int) string {
	switch {
	case p == 0:
		return "reserved"
	case p <= 1023:
		return "well-known"
	case p <= 49151:
		return "registered"
	default:
		return "dynamic"
	}
}

func portRangeDesc(p int) string {
	switch portRangeType(p) {
	case "reserved":
		return "port 0 is reserved and not used for traffic (binding it asks the OS for a free port)"
	case "well-known":
		return "system port (0-1023), usually needs administrator/root rights to listen on"
	case "registered":
		return "registered/user port (1024-49151)"
	}
	return "dynamic/private/ephemeral port (49152-65535)"
}

func portInfoByNumber(p int) map[string]string {
	m := map[string]string{"port": strconv.Itoa(p), "range": portRangeType(p), "range_description": portRangeDesc(p)}
	var protos, names, descs []string
	for _, e := range portTable {
		if e.port == p {
			names = append(names, e.name)
			descs = append(descs, e.desc)
			protos = append(protos, e.protos...)
		}
	}
	if len(names) == 0 {
		m["name"] = ""
		m["protocol"] = ""
		m["description"] = "no entry in the embedded table (partial list)"
		return m
	}
	m["name"] = names[0]
	m["protocol"] = strings.Join(uniq(protos), ",")
	m["description"] = strings.Join(descs, "; ")
	return m
}

func uniq(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}

func servicePorts(name string) []string {
	n := strings.ToLower(strings.TrimSpace(name))
	if n == "" {
		return nil
	}
	collect := func(match func(e portEntry) bool) []string {
		var out []string
		for _, e := range portTable {
			if match(e) {
				for _, pr := range e.protos {
					out = append(out, fmt.Sprintf("%d/%s", e.port, pr))
				}
			}
		}
		return out
	}
	if out := collect(func(e portEntry) bool { return strings.ToLower(e.name) == n }); out != nil {
		return out
	}
	if len(n) < 3 {
		return nil
	}
	return collect(func(e portEntry) bool {
		return strings.Contains(strings.ToLower(e.name), n) || strings.Contains(strings.ToLower(e.desc), n)
	})
}

func parsePort(s string) (int, bool) {
	s = strings.TrimSuffix(strings.TrimSuffix(strings.ToLower(strings.TrimSpace(s)), "/tcp"), "/udp")
	n, err := strconv.Atoi(s)
	return n, err == nil && n >= 0 && n <= 65535
}

func init() {
	reg(kwPort, tools.Func{
		Name: "port_service", Title: "Port or service lookup", Order: 6, Input: "port_or_name",
		Doc:     "Looks up a TCP/UDP port number (\"443\", \"443/tcp\") or a service name (\"ssh\", \"postgresql\") in an embedded table of about 250 well-known and registered ports. Returns name, protocol, description and whether the port is well-known, registered or dynamic. The table is partial: unknown ports still report their range.",
		Params:  []tools.Param{req("port_or_name", "str", `Port number as a string ("22") or a service name ("ssh").`)},
		Returns: "dict[str, str]", Example: `net.port_service("5432")`,
		Tags: []string{"port number", "what runs on port", "iana", "well-known ports", "service name", "tcp", "udp", "etc/services"},
		Accepts: func(s string) bool {
			s = strings.TrimSpace(s)
			if _, ok := parsePort(s); ok {
				return true
			}
			return len(servicePorts(s)) > 0
		},
		Fn: func(a tools.Args) (any, error) {
			s := a.Str("port_or_name")
			if p, ok := parsePort(s); ok {
				return portInfoByNumber(p), nil
			}
			if strings.TrimSpace(s) == "" || strings.ContainsAny(s, " ") {
				return nil, fmt.Errorf("invalid port or service name %q", s)
			}
			list := servicePorts(s)
			if len(list) == 0 {
				return nil, fmt.Errorf("unknown service %q (the embedded table is partial)", s)
			}
			p, _ := strconv.Atoi(strings.SplitN(list[0], "/", 2)[0])
			m := portInfoByNumber(p)
			m["all_ports"] = strings.Join(list, ", ")
			return m, nil
		},
	})
	reg(kwPort, tools.Func{
		Name: "service_ports", Title: "Ports of a service", Order: 80, Input: "name",
		Doc:     "Lists every port/protocol of a service name (\"http\" gives 80/tcp, \"domain\" gives 53/tcp and 53/udp). Matches exact service names first and then service descriptions (e.g. \"mail\", \"database\"); an unknown name gives an empty list.",
		Params:  []tools.Param{req("name", "str", "Service name or keyword such as ssh, dns, redis, mail.")},
		Returns: "list[str]", Example: `net.service_ports("dns")`,
		Tags: []string{"port of service", "which port", "default port", "what port does"},
		Fn: func(a tools.Args) (any, error) {
			n := strings.ToLower(strings.TrimSpace(a.Str("name")))
			if n == "dns" {
				n = "domain"
			}
			out := servicePorts(n)
			if out == nil {
				out = []string{}
			}
			return out, nil
		},
	})
	reg(kwPort, tools.Func{
		Name: "port_range_type", Title: "Port range type", Order: 81, Input: "port",
		Doc:     "Classifies a port number as reserved (0), well-known (1-1023), registered (1024-49151) or dynamic/ephemeral (49152-65535) per IANA. Fails outside 0-65535.",
		Params:  []tools.Param{req("port", "str", `Port number as a string, e.g. "8080".`)},
		Returns: "str", Example: `net.port_range_type("8080")`,
		Tags:    []string{"ephemeral", "privileged port", "system port", "user port", "dynamic port"},
		Accepts: func(s string) bool { _, ok := parsePort(s); return ok },
		Fn: func(a tools.Args) (any, error) {
			p, ok := parsePort(a.Str("port"))
			if !ok {
				return nil, fmt.Errorf("invalid port %q (0-65535)", a.Str("port"))
			}
			return portRangeType(p), nil
		},
	})
}

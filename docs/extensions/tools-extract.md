# Extractors (`extract`)

Pull emails, URLs, domains, IPs, numbers, phone numbers, dates and more out of any text.

No permission is needed: these are pure functions. Generated from the tools registry.

62 functions. Use them as `extract.<name>(...)` in any extension.

## extract.all

`extract.all(text: str) -> Dict[str, str]`

Headline summary of what a text contains: a dict mapping each extractor name (emails, urls, ips, numbers, dates, hashes...) to the number of matches found, listing only non-zero kinds. Structural kinds (words, sentences, paragraphs, integers, decimals, tlds, html tags...) are left out; use extract.summary for those.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |

Returns `Dict[str, str]`. Example: `extract.all("Mail bob@example.com or call +1 555 123 4567, see https://example.com")`

Keywords: all entities, summary of kinds found, everything, overview, find all, entities, analyze text, counts, inventory

## extract.emails

`extract.emails(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, lower: bool = False) -> List[str]`

Finds email addresses (pragmatic RFC 5322 subset, Unicode letters allowed). Trailing punctuation such as a final dot or comma is not included and "mailto:" prefixes are dropped. Addresses with consecutive dots or a dot before @ are skipped.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `lower` | `bool` (default `False`) | Convert results to lower case. |

Returns `List[str]`. Example: `extract.emails("Mail bob@example.com, or mailto:alice.smith@mail.example.org.")`

Keywords: find emails, scrape email addresses, mail addresses in text, e-mail, email addresses, mailto, contacts, harvest

## extract.urls

`extract.urls(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds http, https, ftp, ws and wss URLs and bare www. links. Trailing sentence punctuation and unbalanced closing brackets are trimmed, so (see https://a.com/x) yields https://a.com/x.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.urls("see https://example.com/a?b=1, and www.test.org.")`

Keywords: find urls, find links, scrape hyperlinks, http addresses, links, hyperlinks, http, https, web addresses, ftp, href

## extract.domains

`extract.domains(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, include_subdomains: bool = True, registrable_only: bool = False) -> List[str]`

Finds host names in URLs and bare domains whose suffix is in the public suffix list (so file names like notes.txt are ignored, but readme.md matches because .md is Moldova's TLD). IP addresses are skipped. Results are lower-case; Unicode names are kept as written.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `include_subdomains` | `bool` (default `True`) | Keep hosts that have subdomains; false keeps only apex (registrable) domains. |
| `registrable_only` | `bool` (default `False`) | Reduce every host to its registrable domain (eTLD+1), e.g. a.b.example.co.uk -> example.co.uk. |

Returns `List[str]`. Example: `extract.domains("visit https://blog.example.co.uk/x or docs.python.org", registrable_only=True)`

Keywords: find domain names, hostnames in text, hostnames, fqdn, websites, sites, tld, registrable domain, etld+1

## extract.ips

`extract.ips(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, exclude_private: bool = False) -> List[str]`

Finds both IPv4 and IPv6 addresses in text order. See extract.ipv4 and extract.ipv6 for the validation rules; exclude_private keeps only public addresses.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `exclude_private` | `bool` (default `False`) | Skip private, loopback, link-local and similar non-public addresses. |

Returns `List[str]`. Example: `extract.ips("a 1.2.3.4 b 2001:db8::1")`

Keywords: find ip addresses, ipv4 in log, scan log for ips, ip addresses, ipv4, ipv6, addresses, log analysis, public ip

## extract.numbers

`extract.numbers(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, thousands: bool = True, normalize: bool = False) -> List[str]`

Finds integers and decimals with optional sign, thousands separators (1,234.56) and scientific notation (1.5e-3). Digits inside words (abc123), dotted versions and IPs (1.2.3, 10.0.0.1) and hex-like runs are skipped; short unit suffixes such as 5km are accepted. A minus between digits (2026-10-17) is an operator, not a sign.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `thousands` | `bool` (default `True`) | Treat comma groups as thousands separators ("1,234.5" is one number); false splits at commas. |
| `normalize` | `bool` (default `False`) | Strip thousands separators from the results. |

Returns `List[str]`. Example: `extract.numbers("Paid 1,234.50 for 3 items (-2.5e3 off)")`

Keywords: find digits, numeric values in text, digits, figures, integers, decimals, floats, quantities, scientific notation

## extract.phone_numbers

`extract.phone_numbers(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds international and national phone numbers: optional + or 00, digits grouped by spaces, dashes, dots or parentheses, 7 to 15 digits in total. Heuristic: plain digit runs without + or separators, ISO dates, SSNs, IPs and dotted version numbers are skipped, while long dash-separated ids may still match.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.phone_numbers("Call +1 (555) 123-4567 or 020 7946 0958")`

Keywords: find telephone numbers, phone in text, phone, telephone, mobile, cell, tel, contact numbers, e164, international

## extract.entities

`extract.entities(text: str) -> Dict[str, str]`

The most useful entities of a text as a dict of kind to a comma-separated list of unique values: emails, urls, domains, ips, phone_numbers, hashtags, mentions and dates. Kinds with no match are omitted; each list is capped at 50 values.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |

Returns `Dict[str, str]`. Example: `extract.entities("Contact bob@example.com #news @alice 2026-10-17")`

Keywords: named entities, contacts, ner, key info, overview

## extract.ipv4

`extract.ipv4(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, exclude_private: bool = False) -> List[str]`

Finds dotted-quad IPv4 addresses. Octets must be 0-255 without leading zeros; parts of longer dotted numbers (1.2.3.4.5, version strings) are ignored. exclude_private keeps only publicly routable addresses.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `exclude_private` | `bool` (default `False`) | Skip private, loopback, link-local, CGNAT, documentation and multicast addresses. |

Returns `List[str]`. Example: `extract.ipv4("from 10.0.0.1 to 8.8.8.8", exclude_private=True)`

Keywords: ip address, ipv4, dotted quad, log analysis, public ip

## extract.ipv6

`extract.ipv6(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, exclude_private: bool = False) -> List[str]`

Finds valid IPv6 addresses including :: compression and IPv4-mapped forms. Candidates are verified by a real parser; MAC addresses, times and C++ scope operators like std::vector are not reported.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `exclude_private` | `bool` (default `False`) | Skip loopback, link-local, ULA and multicast addresses. |

Returns `List[str]`. Example: `extract.ipv6("host 2001:db8::1 and ::1")`

Keywords: ip address, ipv6, colon hex, log analysis

## extract.summary

`extract.summary(text: str) -> Dict[str, str]`

Text statistics (characters, bytes, lines, words, sentences, paragraphs) plus the number of matches of every extractor with at least one match, as a dict of strings.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |

Returns `Dict[str, str]`. Example: `extract.summary("Hello world. Mail me at a@b.co")`

Keywords: stats, statistics, word count, character count, line count, analyze, profile

## extract.cidrs

`extract.cidrs(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds IPv4 and IPv6 CIDR blocks such as 10.0.0.0/8 or 2001:db8::/32. The address and the prefix length (0-32 / 0-128) are validated.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.cidrs("allow 10.0.0.0/8 and 2001:db8::/32")`

Keywords: subnets, networks, ip ranges, prefix, acl

## extract.redact

`extract.redact(text: str, kinds: List[str] = [], mask: str = "[REDACTED]") -> str`

Replaces matches of the chosen kinds with a mask. kinds defaults to emails, ips, phone_numbers and credit_cards; any extractor name works (see extract.kinds). A one-character mask is repeated to the match length ("*"), otherwise the mask is inserted as is and {kind} is replaced by the kind name. Overlapping matches are replaced once.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `kinds` | `List[str]` (default `[]`) | Extractor names to redact; empty uses the default four. |
| `mask` | `str` (default `"[REDACTED]"`) | Replacement text, default "[REDACTED]"; "*" masks character by character; "{kind}" is substituted. |

Returns `str`. Example: `extract.redact("Mail bob@example.com from 10.0.0.1", ["emails", "ips"], "<{kind}>")`

Keywords: mask secrets, censor, anonymize, scrub personal data, pii, mask, scrub, sanitize, privacy, hide

## extract.mask_credit_cards

`extract.mask_credit_cards(text: str, mask: str = "*", keep_last: int = 4) -> str`

Masks Luhn-valid 13-19 digit card numbers found in the text, keeping separators and the last keep_last digits (default 4): 4111 1111 1111 1111 becomes **** **** **** 1111.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `mask` | `str` (default `"*"`) | Mask character(s) for each hidden digit. |
| `keep_last` | `int` (default `4`) | Digits left visible at the end (0-8). |

Returns `str`. Example: `extract.mask_credit_cards("card 4111 1111 1111 1111")`

Keywords: hide card numbers, pci masking, pan, pci, hide card, redact card, obfuscate, payment

## extract.ports

`extract.ports(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds host:port endpoints where the host is an IPv4 address, \[IPv6\], localhost or a domain name and the port is 1-65535 (example.com:8080, 10.0.0.1:22, \[::1\]:443). Times like 12:30 are not matched.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.ports("connect to 10.0.0.5:8080 and localhost:3000")`

Keywords: endpoints, sockets, host port, address:port, netstat

## extract.count

`extract.count(text: str, kind: str, unique: bool = False) -> int`

Counts how many items of one kind (see extract.kinds, e.g. emails, urls, ips) appear in the text. By default every occurrence counts; with unique true only distinct values.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `kind` | `str` | Extractor name such as emails, urls or ips. |
| `unique` | `bool` (default `False`) | Count distinct values only. |

Returns `int`. Example: `extract.count("a@b.co a@b.co c@d.org", "emails")`

Keywords: how many, number of, occurrences, tally

## extract.mac_addresses

`extract.mac_addresses(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds MAC addresses written as aa:bb:cc:dd:ee:ff, aa-bb-cc-dd-ee-ff or Cisco aabb.ccdd.eeff. Longer hex/colon sequences (IPv6, EUI-64) are not reported.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.mac_addresses("eth0 00:1A:2B:3C:4D:5E")`

Keywords: mac, hardware address, ethernet, arp, physical address

## extract.extract

`extract.extract(kind: str, text: str, unique: bool = True, sort: bool = False) -> List[str]`

Generic dispatcher: runs the extractor named kind (see extract.kinds) with default options and returns its list. Equivalent to calling extract.&lt;kind&gt;(text); unknown kinds fail with the list of valid names.

| Parameter | Type | Description |
|---|---|---|
| `kind` | `str` | Extractor name such as emails, urls, ips, dates. |
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |

Returns `List[str]`. Example: `extract.extract("emails", "write to bob@example.com")`

Keywords: generic extractor, pull out by type, dispatch, generic, by name, any kind

## extract.kinds

`extract.kinds(filter: str = "") -> List[str]`

Lists the names of all extractors usable with extract.extract, extract.count and extract.redact, optionally filtered by a substring.

| Parameter | Type | Description |
|---|---|---|
| `filter` | `str` (default `""`) | Only kinds containing this text. |

Returns `List[str]`. Example: `extract.kinds("ip")`

Keywords: what can be extracted, list extractors, available, help, names

## extract.lines_containing

`extract.lines_containing(text: str, substr: str, ignore_case: bool = False, invert: bool = False, limit: int = 10000) -> List[str]`

Returns the lines of the text that contain a literal substring (grep -F). ignore_case compares case-insensitively and invert returns the lines that do not contain it. An empty substring matches every line.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `substr` | `str` | Literal text to look for. |
| `ignore_case` | `bool` (default `False`) | Case-insensitive match. |
| `invert` | `bool` (default `False`) | Return the non-matching lines instead. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.lines_containing("ok\nERROR x\nok", "error", True)`

Keywords: grep, find lines, filter lines, search, contains, findstr

## extract.links_markdown

`extract.links_markdown(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, with_text: bool = False) -> List[str]`

Returns the targets of Markdown links \[text\](url), reference definitions (\[id\]: url) and &lt;autolinks&gt;. Images !\[alt\](src) are skipped. With with_text each result is "text -&gt; url".

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `with_text` | `bool` (default `False`) | Return "text -> url" instead of only the url. |

Returns `List[str]`. Example: `extract.links_markdown("[Go](https://go.dev) and ![img](a.png)")`

Keywords: links from markdown, \[text\](url), markdown, md links, anchor text, reference links

## extract.links_html

`extract.links_html(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Returns href values from &lt;a&gt;, &lt;area&gt;, &lt;link&gt; and &lt;base&gt; tags in HTML, with HTML entities decoded (&amp; becomes &). Single, double and unquoted attribute values are supported.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.links_html("<a href='/x?a=1&amp;b=2'>x</a>")`

Keywords: hrefs from html, anchor links, href, anchors, html links, a tag, scrape links

## extract.images_html

`extract.images_html(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Returns the src values of &lt;img&gt; tags in HTML (entities decoded). srcset and CSS background images are not parsed.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.images_html("<img src=\"a.png\"> <img src='b.jpg'>")`

Keywords: img, src, pictures, image urls, scrape images

## extract.html_tags

`extract.html_tags(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, full: bool = False) -> List[str]`

Lists the HTML opening tags found, as lower-case tag names (div, a, img...) or, with full, the complete opening tags including attributes. Comments, doctype and closing tags are ignored.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `full` | `bool` (default `False`) | Return whole opening tags instead of names. |

Returns `List[str]`. Example: `extract.html_tags("<div class='x'><p>hi</p></div>")`

Keywords: tags, elements, markup, html elements

## extract.html_attributes

`extract.html_attributes(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, name: str = "class") -> List[str]`

Returns the values of one attribute (name, default class) from every HTML tag, e.g. all id or alt values. Entities are decoded; attributes without a value are skipped.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `name` | `str` (default `"class"`) | Attribute name such as class, id, alt, data-id. |

Returns `List[str]`. Example: `extract.html_attributes("<p id='a' class='x y'>", "class")`

Keywords: attribute, class, id, alt, title, data attributes

## extract.xml_tags

`extract.xml_tags(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Lists XML element names (with namespace prefix, case preserved) from opening and self-closing tags. Comments, processing instructions, CDATA and closing tags are ignored.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.xml_tags("<root><ns:item id='1'/></root>")`

Keywords: xml, elements, svg, tag names, soap

## extract.tlds

`extract.tlds(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Returns the top-level domain (last label, e.g. com, uk) of every domain or URL host found, using the same detection as extract.domains.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.tlds("a.example.com b.example.co.uk")`

Keywords: top level domain, extensions, country codes, suffix

## extract.integers

`extract.integers(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, thousands: bool = True, normalize: bool = False) -> List[str]`

Like extract.numbers but only whole numbers (no decimal point or exponent), with optional sign and thousands separators.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `thousands` | `bool` (default `True`) | Treat comma groups as thousands separators ("1,234.5" is one number); false splits at commas. |
| `normalize` | `bool` (default `False`) | Strip thousands separators from the results. |

Returns `List[str]`. Example: `extract.integers("1,200 apples, 3.5 pears, -7 plums")`

Keywords: ints, whole numbers, digits, counts

## extract.decimals

`extract.decimals(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, thousands: bool = True, normalize: bool = False) -> List[str]`

Like extract.numbers but only numbers with a fractional part or exponent (3.14, -0.5, 1e6, .75).

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `thousands` | `bool` (default `True`) | Treat comma groups as thousands separators ("1,234.5" is one number); false splits at commas. |
| `normalize` | `bool` (default `False`) | Strip thousands separators from the results. |

Returns `List[str]`. Example: `extract.decimals("pi 3.14159, e 2.718, 42")`

Keywords: floats, fractions, real numbers, doubles

## extract.percentages

`extract.percentages(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds percentages such as 45%, 12.5 %, -3% or 0,5%, returned as written.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.percentages("up 12.5% and down 3%")`

Keywords: percent, rates, ratio, %

## extract.currency_amounts

`extract.currency_amounts(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds money amounts: symbol first ($12.50, €1.299,00, £5, US$10), code first (USD 100, EUR 5.5) or after the number (12,50 €, 100 USD, 5 dollars). Heuristic: it does not interpret locale-specific separators, it returns the text as written.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.currency_amounts("Total $1,299.00 or 12,50 € (USD 100)")`

Keywords: money, price, prices, dollars, euros, cost, amount, invoice

## extract.hex_colors

`extract.hex_colors(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, lower: bool = False) -> List[str]`

Finds CSS hex colors #rgb, #rgba, #rrggbb and #rrggbbaa. HTML entities like &#123; and words glued to the hash are ignored; URL fragments such as #add may still match.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `lower` | `bool` (default `False`) | Convert results to lower case. |

Returns `List[str]`. Example: `extract.hex_colors("color: #FF8800; bg: #fff")`

Keywords: find css colors, color codes in text, color, css colors, hex color codes, rgb hex, palette

## extract.rgb_colors

`extract.rgb_colors(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds CSS rgb(...) and rgba(...) color functions with comma or space separated components (rgb(255, 0, 0), rgba(0 0 0 / 50%)).

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.rgb_colors("color: rgb(255, 128, 0); bg: rgba(0,0,0,0.5)")`

Keywords: color, css colors, rgba, palette

## extract.hashtags

`extract.hashtags(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, lower: bool = False) -> List[str]`

Finds #hashtags (Unicode letters, digits, underscore; purely numeric tags like #1 and HTML entities are skipped). Results include the # sign.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `lower` | `bool` (default `False`) | Convert results to lower case. |

Returns `List[str]`. Example: `extract.hashtags("Loving #Go and #open_source! #2026")`

Keywords: find hash tags, social tags, tags, social media, twitter, instagram, topics

## extract.mentions

`extract.mentions(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, lower: bool = False) -> List[str]`

Finds @username mentions (letters, digits, underscore, up to 30 characters). The @ in email addresses is not a mention. Results include the @ sign.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `lower` | `bool` (default `False`) | Convert results to lower case. |

Returns `List[str]`. Example: `extract.mentions("thanks @alice and @Bob_42, mail me at x@y.com")`

Keywords: handles, usernames, twitter, social media, at sign

## extract.dates

`extract.dates(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds calendar dates: ISO 2026-10-17 (also with / or .), day/month/year and month/day/year (31/12/2026, 12.31.2026), "Jan 2, 2006" and "2 January 2006" (month names and ordinal suffixes allowed). Impossible dates such as 2026-02-30 are rejected; numeric dd/mm vs mm/dd ambiguity is accepted when either reading is valid. Results are returned as written.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.dates("Due 2026-10-17, also 12/25/2026 and March 3rd, 2027")`

Keywords: find dates in text, date strings, date, calendar, iso 8601, timestamps, when

## extract.times

`extract.times(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds clock times hh:mm, hh:mm:ss(.fff) with optional am/pm (9:30, 14:05:59, 9:30 PM) and bare hours like 3pm. Segments of longer colon sequences (MAC or IPv6 addresses) are ignored.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.times("Meet at 9:30 am or 14:45, call by 5pm")`

Keywords: time, clock, hours, am pm, timestamps

## extract.datetimes

`extract.datetimes(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds ISO 8601 date-times such as 2026-10-17T14:30:00Z, 2026-10-17 14:30+02:00 and the basic form 20261017T143000Z. Dates and times are range-checked.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.datetimes("created 2026-10-17T14:30:00Z updated 2026-10-18 09:15")`

Keywords: timestamp, iso 8601, rfc 3339, utc, log timestamps

## extract.file_paths

`extract.file_paths(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds file system paths: Windows drive paths (C:\Users\me\a.txt, also quoted ones with spaces), UNC paths (\\server\share\x) and Unix paths (/usr/local/bin, ~/docs/a.txt, ./src/main.go). Heuristic: unquoted paths end at whitespace, single-component absolute paths need a well-known root (/etc, /tmp...), and URLs are not reported.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.file_paths("open C:\\Users\\me\\a.txt and /var/log/syslog or ~/notes/todo.md")`

Keywords: find paths in text, windows or unix path, paths, directories, folders, unc, windows path, unix path, filesystem

## extract.file_names

`extract.file_names(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds file names with an extension (report.pdf, archive.tar.gz), including the last component of paths. URLs and email addresses are excluded first. Heuristic: abbreviations like e.g and version numbers are skipped, but a bare domain such as example.com in prose looks like a file name.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.file_names("see report.pdf and C:\\tmp\\data.csv, not 1.2.3")`

Keywords: filenames, files, attachments, extension, documents

## extract.file_extensions

`extract.file_extensions(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Returns the extension (lower-case, with leading dot) of every file name found by extract.file_names, e.g. .pdf, .gz.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.file_extensions("a.PDF b.tar.gz c.txt")`

Keywords: suffix, file types, format

## extract.quoted_strings

`extract.quoted_strings(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, include_quotes: bool = False) -> List[str]`

Returns text inside double quotes (with backslash escapes), single quotes (apostrophes inside words are not treated as quotes), backticks, curly quotes (“” ‘’), guillemets and low-9 quotes. Empty strings are skipped.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `include_quotes` | `bool` (default `False`) | Keep the quote characters in the results. |

Returns `List[str]`. Example: `extract.quoted_strings("He said \"hello there\" and 'bye'")`

Keywords: quotes, string literals, double quotes, single quotes, backticks, citations

## extract.words

`extract.words(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, lower: bool = False) -> List[str]`

Splits text into words made of Unicode letters, keeping inner apostrophes and hyphens (don't, well-known). Digits and punctuation separate words. Remember unique defaults to true; pass unique=False to keep every occurrence.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `lower` | `bool` (default `False`) | Convert results to lower case. |

Returns `List[str]`. Example: `extract.words("Hello, wörld! It's well-known.", unique=False)`

Keywords: split into words, tokenize text, tokens, tokenize, vocabulary, word list, unicode

## extract.sentences

`extract.sentences(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Splits text into sentences at . ! ? followed by whitespace, treating common abbreviations (Mr., Dr., e.g., U.S.), initials and a following lower-case word as non-boundaries; blank lines also end a sentence. Whitespace inside a sentence is collapsed.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.sentences("Dr. Smith arrived. He sat down! Did he stay? Yes.")`

Keywords: split sentences, tokenize, nlp, segmentation

## extract.paragraphs

`extract.paragraphs(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Splits text into paragraphs separated by one or more blank lines; each paragraph is trimmed and keeps its inner line breaks. Remember unique defaults to true.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.paragraphs("First para\nstill first\n\nSecond para")`

Keywords: blocks, split text, sections, chunks

## extract.versions

`extract.versions(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds semantic versions like 1.2.3, v2.0.1, 1.0.0-rc.1 and 1.0.0+build.5. Four-part numbers (IP addresses) and decimals are not reported.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.versions("upgrade from v1.4.2 to 2.0.0-beta.1")`

Keywords: semver, release, version, software version, changelog

## extract.git_commit_hashes

`extract.git_commit_hashes(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds abbreviated or full git commit hashes: 7 to 40 lower-case hex characters that contain both a digit and a letter (so plain numbers and words like "facade" are not reported). Overlaps with extract.hashes for 40-character values.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.git_commit_hashes("fixed in 388841a and da693da0c1")`

Keywords: sha, commit, git, revision, changeset, short hash

## extract.non_ascii_chars

`extract.non_ascii_chars(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Lists every character outside 7-bit ASCII (accented letters, symbols, emoji, smart quotes, non-breaking spaces), one per result. Useful to find stray Unicode in source files.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.non_ascii_chars("caf\u00e9 \u2013 na\u00efve \u201cquote\u201d")`

Keywords: unicode, special characters, accents, encoding problems, smart quotes, bom

## extract.capitalized_words

`extract.capitalized_words(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, skip_sentence_start: bool = True, phrases: bool = False) -> List[str]`

Proper-noun candidates: words starting with an upper-case letter followed only by lower-case letters (Paris, Alice). By default words that begin a sentence are skipped; phrases joins neighbours like "New York". Heuristic, no dictionary.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `skip_sentence_start` | `bool` (default `True`) | Ignore capitalized words at the start of a sentence. |
| `phrases` | `bool` (default `False`) | Join consecutive capitalized words (New York City) into one result. |

Returns `List[str]`. Example: `extract.capitalized_words("Yesterday Alice flew to New York with Bob.", phrases=True)`

Keywords: proper nouns, names, entities, people, places, title case

## extract.acronyms

`extract.acronyms(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, min_length: int = 2) -> List[str]`

Finds words made only of capital letters (optionally followed by digits), at least min_length letters long: NASA, HTTP, SQL, USB3. Single letters and plain numbers are ignored.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `min_length` | `int` (default `2`) | Minimum number of capital letters (default 2). |

Returns `List[str]`. Example: `extract.acronyms("The NASA and ESA teams use HTTP and SQL")`

Keywords: abbreviations, initialisms, caps, all caps, uppercase words

## extract.uuids

`extract.uuids(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, lower: bool = False) -> List[str]`

Finds UUIDs/GUIDs in the 8-4-4-4-12 hex form (any version, case-insensitive), including those wrapped in braces. Neighbouring hex or dash characters make a candidate invalid.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `lower` | `bool` (default `False`) | Convert results to lower case. |

Returns `List[str]`. Example: `extract.uuids("id=123e4567-e89b-12d3-a456-426614174000")`

Keywords: find guids in text, ids in log, guid, identifiers, ids, uuid v4, unique id

## extract.hashes

`extract.hashes(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds hexadecimal digests by length and reports them as "type:hash" in lower case: md5 (32), sha1 (40), sha224 (56), sha256 (64), sha384 (96), sha512 (128). A token must mix digits and letters, and the type is a guess from the length only (a 32-hex UUID without dashes also looks like md5).

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.hashes("md5 d41d8cd98f00b204e9800998ecf8427e")`

Keywords: md5, sha1, sha256, sha512, checksum, digest, fingerprint, hex

## extract.base64_strings

`extract.base64_strings(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds plausible standard Base64 strings: at least 16 characters, length divisible by 4, valid padding, and a mix of at least three character classes (or padding plus two), so ordinary words, paths and hex hashes are rejected. URL-safe Base64 (-_) and JWT segments are not reported. Heuristic.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.base64_strings("token SGVsbG8sIFdvcmxkISBIZWxsbyE= end")`

Keywords: b64, encoded, secrets, tokens, payload

## extract.jwt_tokens

`extract.jwt_tokens(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds JSON Web Tokens: three dot-separated Base64url segments whose header decodes to a JSON object (starts with eyJ). The signature may be empty (unsecured tokens).

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.jwt_tokens("Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.c2ln")`

Keywords: bearer, json web token, auth token, access token, jws

## extract.credit_cards

`extract.credit_cards(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds 13-19 digit card numbers (optionally grouped with single spaces or dashes) that pass the Luhn check and do not start with 0 or 1 or repeat one digit. Returned as written. A Luhn match does not prove the number is a real card.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.credit_cards("card 4111 1111 1111 1111 exp 12/29")`

Keywords: card numbers, pan, visa, mastercard, luhn, payment, pci

## extract.ssn

`extract.ssn(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds US Social Security Numbers in the NNN-NN-NNNN form, rejecting impossible values (area 000, 666 or 9xx, group 00, serial 0000).

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.ssn("SSN 123-45-6789")`

Keywords: social security number, us ssn, pii, tax id

## extract.ibans

`extract.ibans(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds International Bank Account Numbers, with or without spaces, validating country, length and the mod-97 checksum. Results are compact upper-case (no spaces).

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.ibans("pay to GB82 WEST 1234 5698 7654 32 please")`

Keywords: bank account, iban, swift, sepa, account number, pii

## extract.crypto_addresses

`extract.crypto_addresses(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds Bitcoin legacy addresses (Base58Check verified), Bitcoin bech32 addresses (bc1..., shape check only) and Ethereum addresses (0x plus 40 hex digits, checksum casing not verified).

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.crypto_addresses("send to 1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2")`

Keywords: bitcoin, btc, ethereum, eth, wallet, blockchain, address

## extract.env_vars

`extract.env_vars(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, names_only: bool = False) -> List[str]`

Finds environment variable references: %VAR% (Windows), $VAR, ${VAR} with optional :-default (POSIX) and $env:VAR (PowerShell). Percentages like "50% of 20%" are not matched because names cannot contain spaces.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |
| `names_only` | `bool` (default `False`) | Return just the variable names without %...%, $ or ${...}. |

Returns `List[str]`. Example: `extract.env_vars("cd %USERPROFILE% && echo $HOME ${PATH:-/bin}", names_only=True)`

Keywords: environment, variables, shell, path, windows, bash, powershell

## extract.placeholders

`extract.placeholders(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds template placeholders: {{name}}, {name}, \[\[name\]\] and &lt;name&gt;. Single-brace placeholders must be identifier-like (so JSON objects are not matched), ${VAR} is treated as an environment variable instead, and known HTML tag names in angle brackets are skipped.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.placeholders("Hello {{user}}, order {id} for <customer>")`

Keywords: template, variables, mustache, handlebars, fill in, merge fields

## extract.json_snippets

`extract.json_snippets(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds JSON objects and arrays embedded in text by matching balanced braces/brackets (honouring strings and escapes) and keeping only snippets that parse as valid JSON. Trivial arrays like \[1\] and unparseable {braces} are skipped; nested snippets are returned as part of their outer one.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.json_snippets("response: {\"a\": [1, 2]} done")`

Keywords: find json in text, embedded json, json, objects, payload, logs, braces

## extract.emojis

`extract.emojis(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]`

Finds emoji, including skin-tone modifiers, ZWJ sequences (family, professions), flags and keycaps. Based on Unicode ranges, so a few rare symbols may be missed or over-matched.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to search. |
| `unique` | `bool` (default `True`) | Drop duplicates, keeping first-seen order. |
| `sort` | `bool` (default `False`) | Sort the result alphabetically. |
| `limit` | `int` (default `10000`) | Maximum number of results (default 10000, max 100000). |

Returns `List[str]`. Example: `extract.emojis("Great job \U0001F44D\U0001F3FD \U0001F1EE\U0001F1F7 ❤️")`

Keywords: emoji, smileys, unicode symbols, emoticons, flags, pictographs


"""Extractors module (generated from the tools registry; do not edit by hand).

Pull emails, URLs, domains, IPs, numbers, phone numbers, dates and more out of any text.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
"""

from typing import Dict, List

def all(text: str) -> Dict[str, str]:
    """Headline summary of what a text contains: a dict mapping each extractor name (emails, urls, ips, numbers, dates, hashes...) to the number of matches found, listing only non-zero kinds. Structural kinds (words, sentences, paragraphs, integers, decimals, tlds, html tags...) are left out; use extract.summary for those.

        text: Text to search.

    Example: extract.all("Mail bob@example.com or call +1 555 123 4567, see https://example.com")
    Needs: none (pure function).
    """

def emails(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, lower: bool = False) -> List[str]:
    """Finds email addresses (pragmatic RFC 5322 subset, Unicode letters allowed). Trailing punctuation such as a final dot or comma is not included and "mailto:" prefixes are dropped. Addresses with consecutive dots or a dot before @ are skipped.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        lower: Convert results to lower case.

    Example: extract.emails("Mail bob@example.com, or mailto:alice.smith@mail.example.org.")
    Needs: none (pure function).
    """

def urls(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds http, https, ftp, ws and wss URLs and bare www. links. Trailing sentence punctuation and unbalanced closing brackets are trimmed, so (see https://a.com/x) yields https://a.com/x.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.urls("see https://example.com/a?b=1, and www.test.org.")
    Needs: none (pure function).
    """

def domains(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, include_subdomains: bool = True, registrable_only: bool = False) -> List[str]:
    """Finds host names in URLs and bare domains whose suffix is in the public suffix list (so file names like notes.txt are ignored, but readme.md matches because .md is Moldova's TLD). IP addresses are skipped. Results are lower-case; Unicode names are kept as written.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        include_subdomains: Keep hosts that have subdomains; false keeps only apex (registrable) domains.
        registrable_only: Reduce every host to its registrable domain (eTLD+1), e.g. a.b.example.co.uk -> example.co.uk.

    Example: extract.domains("visit https://blog.example.co.uk/x or docs.python.org", registrable_only=True)
    Needs: none (pure function).
    """

def ips(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, exclude_private: bool = False) -> List[str]:
    """Finds both IPv4 and IPv6 addresses in text order. See extract.ipv4 and extract.ipv6 for the validation rules; exclude_private keeps only public addresses.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        exclude_private: Skip private, loopback, link-local and similar non-public addresses.

    Example: extract.ips("a 1.2.3.4 b 2001:db8::1")
    Needs: none (pure function).
    """

def numbers(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, thousands: bool = True, normalize: bool = False) -> List[str]:
    """Finds integers and decimals with optional sign, thousands separators (1,234.56) and scientific notation (1.5e-3). Digits inside words (abc123), dotted versions and IPs (1.2.3, 10.0.0.1) and hex-like runs are skipped; short unit suffixes such as 5km are accepted. A minus between digits (2026-10-17) is an operator, not a sign.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        thousands: Treat comma groups as thousands separators ("1,234.5" is one number); false splits at commas.
        normalize: Strip thousands separators from the results.

    Example: extract.numbers("Paid 1,234.50 for 3 items (-2.5e3 off)")
    Needs: none (pure function).
    """

def phone_numbers(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds international and national phone numbers: optional + or 00, digits grouped by spaces, dashes, dots or parentheses, 7 to 15 digits in total. Heuristic: plain digit runs without + or separators, ISO dates, SSNs, IPs and dotted version numbers are skipped, while long dash-separated ids may still match.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.phone_numbers("Call +1 (555) 123-4567 or 020 7946 0958")
    Needs: none (pure function).
    """

def entities(text: str) -> Dict[str, str]:
    """The most useful entities of a text as a dict of kind to a comma-separated list of unique values: emails, urls, domains, ips, phone_numbers, hashtags, mentions and dates. Kinds with no match are omitted; each list is capped at 50 values.

        text: Text to search.

    Example: extract.entities("Contact bob@example.com #news @alice 2026-10-17")
    Needs: none (pure function).
    """

def ipv4(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, exclude_private: bool = False) -> List[str]:
    """Finds dotted-quad IPv4 addresses. Octets must be 0-255 without leading zeros; parts of longer dotted numbers (1.2.3.4.5, version strings) are ignored. exclude_private keeps only publicly routable addresses.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        exclude_private: Skip private, loopback, link-local, CGNAT, documentation and multicast addresses.

    Example: extract.ipv4("from 10.0.0.1 to 8.8.8.8", exclude_private=True)
    Needs: none (pure function).
    """

def ipv6(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, exclude_private: bool = False) -> List[str]:
    """Finds valid IPv6 addresses including :: compression and IPv4-mapped forms. Candidates are verified by a real parser; MAC addresses, times and C++ scope operators like std::vector are not reported.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        exclude_private: Skip loopback, link-local, ULA and multicast addresses.

    Example: extract.ipv6("host 2001:db8::1 and ::1")
    Needs: none (pure function).
    """

def summary(text: str) -> Dict[str, str]:
    """Text statistics (characters, bytes, lines, words, sentences, paragraphs) plus the number of matches of every extractor with at least one match, as a dict of strings.

        text: Text to search.

    Example: extract.summary("Hello world. Mail me at a@b.co")
    Needs: none (pure function).
    """

def cidrs(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds IPv4 and IPv6 CIDR blocks such as 10.0.0.0/8 or 2001:db8::/32. The address and the prefix length (0-32 / 0-128) are validated.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.cidrs("allow 10.0.0.0/8 and 2001:db8::/32")
    Needs: none (pure function).
    """

def redact(text: str, kinds: List[str] = [], mask: str = "[REDACTED]") -> str:
    """Replaces matches of the chosen kinds with a mask. kinds defaults to emails, ips, phone_numbers and credit_cards; any extractor name works (see extract.kinds). A one-character mask is repeated to the match length ("*"), otherwise the mask is inserted as is and {kind} is replaced by the kind name. Overlapping matches are replaced once.

        text: Text to search.
        kinds: Extractor names to redact; empty uses the default four.
        mask: Replacement text, default "[REDACTED]"; "*" masks character by character; "{kind}" is substituted.

    Example: extract.redact("Mail bob@example.com from 10.0.0.1", ["emails", "ips"], "<{kind}>")
    Needs: none (pure function).
    """

def mask_credit_cards(text: str, mask: str = "*", keep_last: int = 4) -> str:
    """Masks Luhn-valid 13-19 digit card numbers found in the text, keeping separators and the last keep_last digits (default 4): 4111 1111 1111 1111 becomes **** **** **** 1111.

        text: Text to search.
        mask: Mask character(s) for each hidden digit.
        keep_last: Digits left visible at the end (0-8).

    Example: extract.mask_credit_cards("card 4111 1111 1111 1111")
    Needs: none (pure function).
    """

def ports(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds host:port endpoints where the host is an IPv4 address, [IPv6], localhost or a domain name and the port is 1-65535 (example.com:8080, 10.0.0.1:22, [::1]:443). Times like 12:30 are not matched.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.ports("connect to 10.0.0.5:8080 and localhost:3000")
    Needs: none (pure function).
    """

def count(text: str, kind: str, unique: bool = False) -> int:
    """Counts how many items of one kind (see extract.kinds, e.g. emails, urls, ips) appear in the text. By default every occurrence counts; with unique true only distinct values.

        text: Text to search.
        kind: Extractor name such as emails, urls or ips.
        unique: Count distinct values only.

    Example: extract.count("a@b.co a@b.co c@d.org", "emails")
    Needs: none (pure function).
    """

def mac_addresses(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds MAC addresses written as aa:bb:cc:dd:ee:ff, aa-bb-cc-dd-ee-ff or Cisco aabb.ccdd.eeff. Longer hex/colon sequences (IPv6, EUI-64) are not reported.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.mac_addresses("eth0 00:1A:2B:3C:4D:5E")
    Needs: none (pure function).
    """

def extract(kind: str, text: str, unique: bool = True, sort: bool = False) -> List[str]:
    """Generic dispatcher: runs the extractor named kind (see extract.kinds) with default options and returns its list. Equivalent to calling extract.<kind>(text); unknown kinds fail with the list of valid names.

        kind: Extractor name such as emails, urls, ips, dates.
        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.

    Example: extract.extract("emails", "write to bob@example.com")
    Needs: none (pure function).
    """

def kinds(filter: str = "") -> List[str]:
    """Lists the names of all extractors usable with extract.extract, extract.count and extract.redact, optionally filtered by a substring.

        filter: Only kinds containing this text.

    Example: extract.kinds("ip")
    Needs: none (pure function).
    """

def lines_containing(text: str, substr: str, ignore_case: bool = False, invert: bool = False, limit: int = 10000) -> List[str]:
    """Returns the lines of the text that contain a literal substring (grep -F). ignore_case compares case-insensitively and invert returns the lines that do not contain it. An empty substring matches every line.

        text: Text to search.
        substr: Literal text to look for.
        ignore_case: Case-insensitive match.
        invert: Return the non-matching lines instead.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.lines_containing("ok\nERROR x\nok", "error", True)
    Needs: none (pure function).
    """

def links_markdown(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, with_text: bool = False) -> List[str]:
    """Returns the targets of Markdown links [text](url), reference definitions ([id]: url) and <autolinks>. Images ![alt](src) are skipped. With with_text each result is "text -> url".

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        with_text: Return "text -> url" instead of only the url.

    Example: extract.links_markdown("[Go](https://go.dev) and ![img](a.png)")
    Needs: none (pure function).
    """

def links_html(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Returns href values from <a>, <area>, <link> and <base> tags in HTML, with HTML entities decoded (&amp; becomes &). Single, double and unquoted attribute values are supported.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.links_html("<a href='/x?a=1&amp;b=2'>x</a>")
    Needs: none (pure function).
    """

def images_html(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Returns the src values of <img> tags in HTML (entities decoded). srcset and CSS background images are not parsed.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.images_html("<img src=\"a.png\"> <img src='b.jpg'>")
    Needs: none (pure function).
    """

def html_tags(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, full: bool = False) -> List[str]:
    """Lists the HTML opening tags found, as lower-case tag names (div, a, img...) or, with full, the complete opening tags including attributes. Comments, doctype and closing tags are ignored.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        full: Return whole opening tags instead of names.

    Example: extract.html_tags("<div class='x'><p>hi</p></div>")
    Needs: none (pure function).
    """

def html_attributes(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, name: str = "class") -> List[str]:
    """Returns the values of one attribute (name, default class) from every HTML tag, e.g. all id or alt values. Entities are decoded; attributes without a value are skipped.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        name: Attribute name such as class, id, alt, data-id.

    Example: extract.html_attributes("<p id='a' class='x y'>", "class")
    Needs: none (pure function).
    """

def xml_tags(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Lists XML element names (with namespace prefix, case preserved) from opening and self-closing tags. Comments, processing instructions, CDATA and closing tags are ignored.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.xml_tags("<root><ns:item id='1'/></root>")
    Needs: none (pure function).
    """

def tlds(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Returns the top-level domain (last label, e.g. com, uk) of every domain or URL host found, using the same detection as extract.domains.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.tlds("a.example.com b.example.co.uk")
    Needs: none (pure function).
    """

def integers(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, thousands: bool = True, normalize: bool = False) -> List[str]:
    """Like extract.numbers but only whole numbers (no decimal point or exponent), with optional sign and thousands separators.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        thousands: Treat comma groups as thousands separators ("1,234.5" is one number); false splits at commas.
        normalize: Strip thousands separators from the results.

    Example: extract.integers("1,200 apples, 3.5 pears, -7 plums")
    Needs: none (pure function).
    """

def decimals(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, thousands: bool = True, normalize: bool = False) -> List[str]:
    """Like extract.numbers but only numbers with a fractional part or exponent (3.14, -0.5, 1e6, .75).

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        thousands: Treat comma groups as thousands separators ("1,234.5" is one number); false splits at commas.
        normalize: Strip thousands separators from the results.

    Example: extract.decimals("pi 3.14159, e 2.718, 42")
    Needs: none (pure function).
    """

def percentages(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds percentages such as 45%, 12.5 %, -3% or 0,5%, returned as written.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.percentages("up 12.5% and down 3%")
    Needs: none (pure function).
    """

def currency_amounts(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds money amounts: symbol first ($12.50, €1.299,00, £5, US$10), code first (USD 100, EUR 5.5) or after the number (12,50 €, 100 USD, 5 dollars). Heuristic: it does not interpret locale-specific separators, it returns the text as written.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.currency_amounts("Total $1,299.00 or 12,50 € (USD 100)")
    Needs: none (pure function).
    """

def hex_colors(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, lower: bool = False) -> List[str]:
    """Finds CSS hex colors #rgb, #rgba, #rrggbb and #rrggbbaa. HTML entities like &#123; and words glued to the hash are ignored; URL fragments such as #add may still match.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        lower: Convert results to lower case.

    Example: extract.hex_colors("color: #FF8800; bg: #fff")
    Needs: none (pure function).
    """

def rgb_colors(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds CSS rgb(...) and rgba(...) color functions with comma or space separated components (rgb(255, 0, 0), rgba(0 0 0 / 50%)).

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.rgb_colors("color: rgb(255, 128, 0); bg: rgba(0,0,0,0.5)")
    Needs: none (pure function).
    """

def hashtags(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, lower: bool = False) -> List[str]:
    """Finds #hashtags (Unicode letters, digits, underscore; purely numeric tags like #1 and HTML entities are skipped). Results include the # sign.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        lower: Convert results to lower case.

    Example: extract.hashtags("Loving #Go and #open_source! #2026")
    Needs: none (pure function).
    """

def mentions(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, lower: bool = False) -> List[str]:
    """Finds @username mentions (letters, digits, underscore, up to 30 characters). The @ in email addresses is not a mention. Results include the @ sign.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        lower: Convert results to lower case.

    Example: extract.mentions("thanks @alice and @Bob_42, mail me at x@y.com")
    Needs: none (pure function).
    """

def dates(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds calendar dates: ISO 2026-10-17 (also with / or .), day/month/year and month/day/year (31/12/2026, 12.31.2026), "Jan 2, 2006" and "2 January 2006" (month names and ordinal suffixes allowed). Impossible dates such as 2026-02-30 are rejected; numeric dd/mm vs mm/dd ambiguity is accepted when either reading is valid. Results are returned as written.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.dates("Due 2026-10-17, also 12/25/2026 and March 3rd, 2027")
    Needs: none (pure function).
    """

def times(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds clock times hh:mm, hh:mm:ss(.fff) with optional am/pm (9:30, 14:05:59, 9:30 PM) and bare hours like 3pm. Segments of longer colon sequences (MAC or IPv6 addresses) are ignored.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.times("Meet at 9:30 am or 14:45, call by 5pm")
    Needs: none (pure function).
    """

def datetimes(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds ISO 8601 date-times such as 2026-10-17T14:30:00Z, 2026-10-17 14:30+02:00 and the basic form 20261017T143000Z. Dates and times are range-checked.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.datetimes("created 2026-10-17T14:30:00Z updated 2026-10-18 09:15")
    Needs: none (pure function).
    """

def file_paths(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds file system paths: Windows drive paths (C:\Users\me\a.txt, also quoted ones with spaces), UNC paths (\\server\share\x) and Unix paths (/usr/local/bin, ~/docs/a.txt, ./src/main.go). Heuristic: unquoted paths end at whitespace, single-component absolute paths need a well-known root (/etc, /tmp...), and URLs are not reported.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.file_paths("open C:\\Users\\me\\a.txt and /var/log/syslog or ~/notes/todo.md")
    Needs: none (pure function).
    """

def file_names(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds file names with an extension (report.pdf, archive.tar.gz), including the last component of paths. URLs and email addresses are excluded first. Heuristic: abbreviations like e.g and version numbers are skipped, but a bare domain such as example.com in prose looks like a file name.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.file_names("see report.pdf and C:\\tmp\\data.csv, not 1.2.3")
    Needs: none (pure function).
    """

def file_extensions(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Returns the extension (lower-case, with leading dot) of every file name found by extract.file_names, e.g. .pdf, .gz.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.file_extensions("a.PDF b.tar.gz c.txt")
    Needs: none (pure function).
    """

def quoted_strings(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, include_quotes: bool = False) -> List[str]:
    """Returns text inside double quotes (with backslash escapes), single quotes (apostrophes inside words are not treated as quotes), backticks, curly quotes (“” ‘’), guillemets and low-9 quotes. Empty strings are skipped.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        include_quotes: Keep the quote characters in the results.

    Example: extract.quoted_strings("He said \"hello there\" and 'bye'")
    Needs: none (pure function).
    """

def words(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, lower: bool = False) -> List[str]:
    """Splits text into words made of Unicode letters, keeping inner apostrophes and hyphens (don't, well-known). Digits and punctuation separate words. Remember unique defaults to true; pass unique=False to keep every occurrence.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        lower: Convert results to lower case.

    Example: extract.words("Hello, wörld! It's well-known.", unique=False)
    Needs: none (pure function).
    """

def sentences(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Splits text into sentences at . ! ? followed by whitespace, treating common abbreviations (Mr., Dr., e.g., U.S.), initials and a following lower-case word as non-boundaries; blank lines also end a sentence. Whitespace inside a sentence is collapsed.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.sentences("Dr. Smith arrived. He sat down! Did he stay? Yes.")
    Needs: none (pure function).
    """

def paragraphs(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Splits text into paragraphs separated by one or more blank lines; each paragraph is trimmed and keeps its inner line breaks. Remember unique defaults to true.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.paragraphs("First para\nstill first\n\nSecond para")
    Needs: none (pure function).
    """

def versions(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds semantic versions like 1.2.3, v2.0.1, 1.0.0-rc.1 and 1.0.0+build.5. Four-part numbers (IP addresses) and decimals are not reported.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.versions("upgrade from v1.4.2 to 2.0.0-beta.1")
    Needs: none (pure function).
    """

def git_commit_hashes(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds abbreviated or full git commit hashes: 7 to 40 lower-case hex characters that contain both a digit and a letter (so plain numbers and words like "facade" are not reported). Overlaps with extract.hashes for 40-character values.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.git_commit_hashes("fixed in 388841a and da693da0c1")
    Needs: none (pure function).
    """

def non_ascii_chars(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Lists every character outside 7-bit ASCII (accented letters, symbols, emoji, smart quotes, non-breaking spaces), one per result. Useful to find stray Unicode in source files.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.non_ascii_chars("caf\u00e9 \u2013 na\u00efve \u201cquote\u201d")
    Needs: none (pure function).
    """

def capitalized_words(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, skip_sentence_start: bool = True, phrases: bool = False) -> List[str]:
    """Proper-noun candidates: words starting with an upper-case letter followed only by lower-case letters (Paris, Alice). By default words that begin a sentence are skipped; phrases joins neighbours like "New York". Heuristic, no dictionary.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        skip_sentence_start: Ignore capitalized words at the start of a sentence.
        phrases: Join consecutive capitalized words (New York City) into one result.

    Example: extract.capitalized_words("Yesterday Alice flew to New York with Bob.", phrases=True)
    Needs: none (pure function).
    """

def acronyms(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, min_length: int = 2) -> List[str]:
    """Finds words made only of capital letters (optionally followed by digits), at least min_length letters long: NASA, HTTP, SQL, USB3. Single letters and plain numbers are ignored.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        min_length: Minimum number of capital letters (default 2).

    Example: extract.acronyms("The NASA and ESA teams use HTTP and SQL")
    Needs: none (pure function).
    """

def uuids(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, lower: bool = False) -> List[str]:
    """Finds UUIDs/GUIDs in the 8-4-4-4-12 hex form (any version, case-insensitive), including those wrapped in braces. Neighbouring hex or dash characters make a candidate invalid.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        lower: Convert results to lower case.

    Example: extract.uuids("id=123e4567-e89b-12d3-a456-426614174000")
    Needs: none (pure function).
    """

def hashes(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds hexadecimal digests by length and reports them as "type:hash" in lower case: md5 (32), sha1 (40), sha224 (56), sha256 (64), sha384 (96), sha512 (128). A token must mix digits and letters, and the type is a guess from the length only (a 32-hex UUID without dashes also looks like md5).

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.hashes("md5 d41d8cd98f00b204e9800998ecf8427e")
    Needs: none (pure function).
    """

def base64_strings(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds plausible standard Base64 strings: at least 16 characters, length divisible by 4, valid padding, and a mix of at least three character classes (or padding plus two), so ordinary words, paths and hex hashes are rejected. URL-safe Base64 (-_) and JWT segments are not reported. Heuristic.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.base64_strings("token SGVsbG8sIFdvcmxkISBIZWxsbyE= end")
    Needs: none (pure function).
    """

def jwt_tokens(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds JSON Web Tokens: three dot-separated Base64url segments whose header decodes to a JSON object (starts with eyJ). The signature may be empty (unsecured tokens).

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.jwt_tokens("Authorization: Bearer eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.c2ln")
    Needs: none (pure function).
    """

def credit_cards(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds 13-19 digit card numbers (optionally grouped with single spaces or dashes) that pass the Luhn check and do not start with 0 or 1 or repeat one digit. Returned as written. A Luhn match does not prove the number is a real card.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.credit_cards("card 4111 1111 1111 1111 exp 12/29")
    Needs: none (pure function).
    """

def ssn(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds US Social Security Numbers in the NNN-NN-NNNN form, rejecting impossible values (area 000, 666 or 9xx, group 00, serial 0000).

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.ssn("SSN 123-45-6789")
    Needs: none (pure function).
    """

def ibans(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds International Bank Account Numbers, with or without spaces, validating country, length and the mod-97 checksum. Results are compact upper-case (no spaces).

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.ibans("pay to GB82 WEST 1234 5698 7654 32 please")
    Needs: none (pure function).
    """

def crypto_addresses(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds Bitcoin legacy addresses (Base58Check verified), Bitcoin bech32 addresses (bc1..., shape check only) and Ethereum addresses (0x plus 40 hex digits, checksum casing not verified).

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.crypto_addresses("send to 1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2")
    Needs: none (pure function).
    """

def env_vars(text: str, unique: bool = True, sort: bool = False, limit: int = 10000, names_only: bool = False) -> List[str]:
    """Finds environment variable references: %VAR% (Windows), $VAR, ${VAR} with optional :-default (POSIX) and $env:VAR (PowerShell). Percentages like "50% of 20%" are not matched because names cannot contain spaces.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).
        names_only: Return just the variable names without %...%, $ or ${...}.

    Example: extract.env_vars("cd %USERPROFILE% && echo $HOME ${PATH:-/bin}", names_only=True)
    Needs: none (pure function).
    """

def placeholders(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds template placeholders: {{name}}, {name}, [[name]] and <name>. Single-brace placeholders must be identifier-like (so JSON objects are not matched), ${VAR} is treated as an environment variable instead, and known HTML tag names in angle brackets are skipped.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.placeholders("Hello {{user}}, order {id} for <customer>")
    Needs: none (pure function).
    """

def json_snippets(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds JSON objects and arrays embedded in text by matching balanced braces/brackets (honouring strings and escapes) and keeping only snippets that parse as valid JSON. Trivial arrays like [1] and unparseable {braces} are skipped; nested snippets are returned as part of their outer one.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.json_snippets("response: {\"a\": [1, 2]} done")
    Needs: none (pure function).
    """

def emojis(text: str, unique: bool = True, sort: bool = False, limit: int = 10000) -> List[str]:
    """Finds emoji, including skin-tone modifiers, ZWJ sequences (family, professions), flags and keycaps. Based on Unicode ranges, so a few rare symbols may be missed or over-matched.

        text: Text to search.
        unique: Drop duplicates, keeping first-seen order.
        sort: Sort the result alphabetically.
        limit: Maximum number of results (default 10000, max 100000).

    Example: extract.emojis("Great job \U0001F44D\U0001F3FD \U0001F1EE\U0001F1F7 ❤️")
    Needs: none (pure function).
    """

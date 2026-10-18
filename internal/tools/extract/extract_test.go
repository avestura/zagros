package extract

import (
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "extract") }

func ex(t *testing.T, name, text string, kv ...any) []string {
	t.Helper()
	m := map[string]any{"text": text}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	v, err := tools.Call("extract", name, m)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	l, ok := v.([]string)
	if !ok {
		t.Fatalf("%s: result %T", name, v)
	}
	return l
}

func want(t *testing.T, got []string, exp ...string) {
	t.Helper()
	if strings.Join(got, "\x00") != strings.Join(exp, "\x00") {
		t.Errorf("got %q, want %q", got, exp)
	}
}

func TestNetwork(t *testing.T) {
	want(t, ex(t, "emails", "Mail Bob.Smith+x@Example.com, or mailto:alice@mail.example.org. Not a@b, nor .x@y.com or a..b@c.com"),
		"Bob.Smith+x@Example.com", "alice@mail.example.org", "x@y.com")
	want(t, ex(t, "emails", "A@B.CO a@b.co", "lower", true), "a@b.co")
	want(t, ex(t, "urls", "see https://a.com/x?y=1, (http://b.org/p_(1)) and www.test.org. ftp://f.net/z; done!"),
		"https://a.com/x?y=1", "http://b.org/p_(1)", "www.test.org", "ftp://f.net/z")
	want(t, ex(t, "urls", "http:// nothing, www. nothing"))
	want(t, ex(t, "links_markdown", "[Go](https://go.dev \"t\") ![img](a.png) <https://auto.link/x>\n[ref]: https://r.example/y"),
		"https://go.dev", "https://auto.link/x", "https://r.example/y")
	want(t, ex(t, "links_markdown", "[Go](https://go.dev)", "with_text", true), "Go -> https://go.dev")
	want(t, ex(t, "links_html", `<a href="/a?x=1&amp;y=2">a</a> <A HREF='b.html'>b</A> <a href=c.html>c</a> <img src="i.png">`), "/a?x=1&y=2", "b.html", "c.html")
	want(t, ex(t, "images_html", `<img alt=x src="a.png"/><IMG SRC='b.jpg'><a href="no.html">`), "a.png", "b.jpg")
	want(t, ex(t, "html_tags", `<div class="x"><p>hi</p><br/><!-- c --></div>`), "div", "p", "br")
	want(t, ex(t, "html_attributes", `<p id="a" class="x y"><span class='z'>`, "name", "class"), "x y", "z")
	want(t, ex(t, "xml_tags", `<?xml version="1.0"?><root><ns:Item id="1"/><!-- c --></root>`), "root", "ns:Item")

	want(t, ex(t, "domains", "visit https://Blog.Example.co.uk:8080/x, docs.python.org, notes.txt and user@mail.example.com; 10.0.0.1 http://localhost/x"),
		"blog.example.co.uk", "docs.python.org", "mail.example.com", "localhost")
	want(t, ex(t, "domains", "a.b.example.co.uk x.github.io", "registrable_only", true), "example.co.uk", "x.github.io")
	want(t, ex(t, "domains", "www.example.com example.com", "include_subdomains", false), "example.com")
	want(t, ex(t, "tlds", "a.example.com b.example.co.uk"), "com", "uk")

	want(t, ex(t, "ipv4", "ips 10.0.0.1, 8.8.8.8 256.1.1.1 1.2.3.4.5 01.2.3.4 v1.2.3.4x"), "10.0.0.1", "8.8.8.8")
	want(t, ex(t, "ipv4", "10.0.0.1 192.168.1.1 127.0.0.1 8.8.8.8 100.64.0.1", "exclude_private", true), "8.8.8.8")
	want(t, ex(t, "ipv6", "a 2001:db8::1, ::1 fe80::1% std::vector 12:30:45 aa:bb:cc:dd:ee:ff end"), "2001:db8::1", "::1", "fe80::1")
	want(t, ex(t, "ipv6", "::ffff:192.0.2.1."), "::ffff:192.0.2.1")
	want(t, ex(t, "ips", "1.1.1.1 and 2001:db8::2 and 1.1.1.1"), "1.1.1.1", "2001:db8::2")
	want(t, ex(t, "cidrs", "allow 10.0.0.0/8, 192.168.0.0/33 and 2001:db8::/32."), "10.0.0.0/8", "2001:db8::/32")
	want(t, ex(t, "ports", "connect 10.0.0.5:8080, localhost:3000 example.com:443 at 12:30 and a.b:99999"), "10.0.0.5:8080", "localhost:3000", "example.com:443")
	want(t, ex(t, "mac_addresses", "eth 00:1A:2B:3C:4D:5E and aa-bb-cc-dd-ee-ff, 001a.2b3c.4d5e; no 2001:db8:85a3:0:0:8a2e:370:7334"),
		"00:1A:2B:3C:4D:5E", "aa-bb-cc-dd-ee-ff", "001a.2b3c.4d5e")
}

func TestNumbersAndFormats(t *testing.T) {
	want(t, ex(t, "numbers", "Paid 1,234.50 for 3 items (-2.5e3 off), 5km, abc123, 1.2.3, 2026-10-17, 0x1F, .75"),
		"1,234.50", "3", "-2.5e3", "5", "2026", "10", "17", ".75")
	want(t, ex(t, "numbers", "1,234 and 5", "normalize", true), "1234", "5")
	want(t, ex(t, "numbers", "1,234", "thousands", false), "1", "234")
	want(t, ex(t, "integers", "1,200 apples, 3.5 pears, -7 plums, 1e5"), "1,200", "-7")
	want(t, ex(t, "decimals", "pi 3.14159, 42, -0.5, 1e6"), "3.14159", "-0.5", "1e6")
	want(t, ex(t, "percentages", "up 12.5% down -3 % and 100%"), "12.5%", "-3 %", "100%")
	want(t, ex(t, "currency_amounts", "Total $1,299.00 or 12,50 € (USD 100), 5 dollars, EUR 5.5 and £5"),
		"$1,299.00", "12,50 €", "USD 100", "5 dollars", "EUR 5.5", "£5")
	want(t, ex(t, "hex_colors", "color: #FF8800; bg: #fff; &#123; #12345 #abcd"), "#FF8800", "#fff", "#abcd")
	want(t, ex(t, "rgb_colors", "rgb(255, 128, 0) rgba(0,0,0,0.5) rgb(1 2 3 / 50%) rgb(1,2)"), "rgb(255, 128, 0)", "rgba(0,0,0,0.5)", "rgb(1 2 3 / 50%)")
	want(t, ex(t, "hashtags", "Loving #Go and #open_source! #2026 a#b &#39; #Ünï", "lower", true), "#go", "#open_source", "#ünï")
	want(t, ex(t, "mentions", "thanks @alice and @Bob_42, mail x@y.com @"), "@alice", "@Bob_42")
	want(t, ex(t, "uuids", "id=123E4567-e89b-12d3-a456-426614174000 {123e4567-e89b-12d3-a456-426614174001} x123e4567-e89b-12d3-a456-426614174002"),
		"123E4567-e89b-12d3-a456-426614174000", "123e4567-e89b-12d3-a456-426614174001")
	want(t, ex(t, "hashes", "md5 D41D8CD98F00B204E9800998ECF8427E sha1 da39a3ee5e6b4b0d3255bfef95601890afd80709 facadefacadefacadefacadefacadefac 123456789012345678901234567890ab1"),
		"md5:d41d8cd98f00b204e9800998ecf8427e", "sha1:da39a3ee5e6b4b0d3255bfef95601890afd80709")
	want(t, ex(t, "base64_strings", "token SGVsbG8sIFdvcmxkISBIZWxsbyE= end /usr/local/bin/something internationalizationlong abcd"),
		"SGVsbG8sIFdvcmxkISBIZWxsbyE=")
	jwt := "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxIn0.c2lnbmF0dXJl"
	want(t, ex(t, "jwt_tokens", "Bearer "+jwt+" eyJmoo.bar.baz"), jwt)
	want(t, ex(t, "base64_strings", jwt))
}

func TestDatesTimes(t *testing.T) {
	want(t, ex(t, "dates", "Due 2026-10-17, 31/12/2026, 12.31.2026, March 3rd, 2027, 2 January 2006 but 2026-02-30 and 99/99/2026 and 1.2.2026.5"),
		"2026-10-17", "31/12/2026", "12.31.2026", "March 3rd, 2027", "2 January 2006")
	want(t, ex(t, "dates", "Jan 2, 2006 and Sept. 9 2020"), "Jan 2, 2006", "Sept. 9 2020")
	want(t, ex(t, "times", "Meet at 9:30 am or 14:45, call by 5pm; 10:20:30:40:50:60 25:00 12:60 3:1"), "9:30 am", "14:45", "5pm")
	want(t, ex(t, "times", "at 23:59:59.123 and 7:05"), "23:59:59.123", "7:05")
	want(t, ex(t, "datetimes", "created 2026-10-17T14:30:00Z updated 2026-10-18 09:15 odd 2026-13-01T10:00 basic 20261017T143000Z tz 2026-10-17T14:30:00+02:00"),
		"2026-10-17T14:30:00Z", "2026-10-18 09:15", "20261017T143000Z", "2026-10-17T14:30:00+02:00")
}

func TestPathsAndFiles(t *testing.T) {
	want(t, ex(t, "file_paths", `open C:\Users\me\a.txt, "D:\My Files\b c.doc" and \\server\share\x.txt then /var/log/syslog, ~/notes/todo.md ./src/main.go and /tmp see https://example.com/a/b and and/or`),
		`C:\Users\me\a.txt`, `D:\My Files\b c.doc`, `\\server\share\x.txt`, "/var/log/syslog", "~/notes/todo.md", "./src/main.go", "/tmp")
	want(t, ex(t, "file_paths", "x /etc/hosts. y /a"), "/etc/hosts")
	want(t, ex(t, "file_names", `see report.pdf and C:\tmp\data.csv, archive.tar.gz; not 1.2.3, e.g. foo, https://example.com/x.html or me@host.com`),
		"report.pdf", "data.csv", "archive.tar.gz")
	want(t, ex(t, "file_extensions", "a.PDF b.tar.gz c.txt d.pdf"), ".pdf", ".gz", ".txt")
}

func TestText(t *testing.T) {
	want(t, ex(t, "quoted_strings", "He said \"hello there\" and 'bye' don't `code` \u201csmart\u201d \u2018single\u2019 \u00abguil\u00bb \"\" it's fine"),
		"hello there", "bye", "code", "smart", "single", "guil")
	want(t, ex(t, "quoted_strings", `say "hi"`, "include_quotes", true), `"hi"`)
	want(t, ex(t, "words", "Hello, wörld! It's well-known 123.", "lower", true), "hello", "wörld", "it's", "well-known")
	want(t, ex(t, "words", "a b a", "unique", false), "a", "b", "a")
	want(t, ex(t, "sentences", "Dr. Smith arrived at 3.30 p.m. He sat down! Did he stay? Yes. e.g. this one. Ok"),
		"Dr. Smith arrived at 3.30 p.m. He sat down!", "Did he stay?", "Yes. e.g. this one.", "Ok")
	want(t, ex(t, "paragraphs", "First para\nstill first\n\n\nSecond para\r\n\r\nThird"), "First para\nstill first", "Second para", "Third")
	want(t, ex(t, "versions", "upgrade from v1.4.2 to 2.0.0-beta.1+b5; ip 1.2.3.4 and 1.2 and 01.2.3"), "v1.4.2", "2.0.0-beta.1+b5")
	want(t, ex(t, "git_commit_hashes", "fixed in 388841a and da693da0c1, not facade or 1234567 or 0xdeadbeef"), "388841a", "da693da0c1")
	want(t, ex(t, "emojis", "Great \U0001F44D\U0001F3FD flag \U0001F1EE\U0001F1F7 heart \u2764\uFE0F fam \U0001F468\u200D\U0001F469\u200D\U0001F467 1\uFE0F\u20E3 plain 1 #"),
		"\U0001F44D\U0001F3FD", "\U0001F1EE\U0001F1F7", "\u2764\uFE0F", "\U0001F468\u200D\U0001F469\u200D\U0001F467", "1\uFE0F\u20E3")
	want(t, ex(t, "non_ascii_chars", "caf\u00e9 \u2013 caf\u00e9"), "\u00e9", "\u2013")
	want(t, ex(t, "capitalized_words", "Yesterday Alice flew to New York with Bob. Paris is NASA free."), "Alice", "New", "York", "Bob")
	want(t, ex(t, "capitalized_words", "Yesterday Alice flew to New York.", "phrases", true), "Alice", "New York")
	want(t, ex(t, "acronyms", "The NASA and ESA teams use HTTP, SQL, A and USB3 and 1234."), "NASA", "ESA", "HTTP", "SQL", "USB3")
}

func TestIdentifiers(t *testing.T) {
	want(t, ex(t, "credit_cards", "card 4111 1111 1111 1111, 4111-1111-1111-1111 bad 4111 1111 1111 1112 and 1111111111111111 or 4111 1111-1111 1111"),
		"4111 1111 1111 1111", "4111-1111-1111-1111")
	want(t, ex(t, "ssn", "SSN 123-45-6789, 000-12-3456 666-12-3456 900-12-3456 123-00-6789 123-45-0000"), "123-45-6789")
	want(t, ex(t, "ibans", "pay GB82 WEST 1234 5698 7654 32 or de89370400440532013000 not GB82 WEST 1234 5698 7654 33"),
		"GB82WEST12345698765432")
	want(t, ex(t, "crypto_addresses", "btc 1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2 bad 1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN3 eth 0x52908400098527886E0F7030069857D2E4169EE7 bech bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq"),
		"1BvBMSEYstWetqTFn5Au4m4GFg7xJaNVN2", "0x52908400098527886E0F7030069857D2E4169EE7", "bc1qar0srrr7xfkvy5l643lydnw9re59gtzzwf5mdq")
	want(t, ex(t, "env_vars", `cd %USERPROFILE% && echo $HOME ${PATH:-/bin} $env:TEMP 50% of 20% cost $5`), "%USERPROFILE%", "$HOME", "${PATH:-/bin}", "$env:TEMP")
	want(t, ex(t, "env_vars", `%A% $B ${C:-x}`, "names_only", true), "A", "B", "C")
	want(t, ex(t, "placeholders", `Hello {{user}}, order {id} for <customer> in <div> and ${HOME} {"a": 1} [[slot]]`), "{{user}}", "{id}", "<customer>", "[[slot]]")
	want(t, ex(t, "json_snippets", `resp: {"a": [1, 2], "s": "}{"} then [1] and {not json} and [{"b": null}, 2] end {}`), `{"a": [1, 2], "s": "}{"}`, `[{"b": null}, 2]`)
	want(t, ex(t, "phone_numbers", "Call +1 (555) 123-4567 or 020 7946 0958, 555-123-4567; not 2026-10-17, 10.0.0.1, 1234567, 123-45-6789 or 1.2.3.4.5"),
		"+1 (555) 123-4567", "020 7946 0958", "555-123-4567")
	want(t, ex(t, "phone_numbers", "intl +4930123456 and 0044 20 7946 0958"), "+4930123456", "0044 20 7946 0958")
}

func TestAggregates(t *testing.T) {
	text := "Mail bob@example.com or call +1 555 123 4567, see https://example.com/x #news @alice 2026-10-17 from 10.0.0.1"
	call := func(name string, kv ...any) any {
		m := map[string]any{}
		for i := 0; i < len(kv); i += 2 {
			m[kv[i].(string)] = kv[i+1]
		}
		v, err := tools.Call("extract", name, m)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		return v
	}
	all := call("all", "text", text).(map[string]string)
	for _, k := range []string{"emails", "urls", "domains", "ips", "phone_numbers", "hashtags", "mentions", "dates"} {
		if all[k] == "" {
			t.Errorf("all() missing %s: %v", k, all)
		}
	}
	if _, bad := all["words"]; bad {
		t.Error("all() must skip words")
	}
	if len(call("all", "text", "").(map[string]string)) != 0 {
		t.Error("all of empty text must be empty")
	}
	en := call("entities", "text", text).(map[string]string)
	if en["emails"] != "bob@example.com" || en["domains"] != "example.com" || en["hashtags"] != "#news" || en["dates"] != "2026-10-17" {
		t.Errorf("entities: %v", en)
	}
	sm := call("summary", "text", "Hello world. Mail a@b.co\n\nBye").(map[string]string)
	if sm["lines"] != "3" || sm["words"] != "7" || sm["paragraphs"] != "2" || sm["emails"] != "1" {
		t.Errorf("summary: %v", sm)
	}
	red := call("redact", "text", "Mail bob@example.com from 10.0.0.1 card 4111 1111 1111 1111").(string)
	if red != "Mail [REDACTED] from [REDACTED] card [REDACTED]" {
		t.Errorf("redact: %q", red)
	}
	if r := call("redact", "text", "x bob@a.co y", "kinds", []string{"emails"}, "mask", "*").(string); r != "x ******** y" {
		t.Errorf("redact star: %q", r)
	}
	if r := call("redact", "text", "x bob@a.co 1.2.3.4", "kinds", []string{"emails", "ips"}, "mask", "<{kind}>").(string); r != "x <emails> <ips>" {
		t.Errorf("redact kind: %q", r)
	}
	if _, err := tools.Call("extract", "redact", map[string]any{"text": "x", "kinds": []string{"nope"}}); err == nil {
		t.Error("unknown kind must fail")
	}
	if m := call("mask_credit_cards", "text", "card 4111 1111 1111 1111 ok").(string); m != "card **** **** **** 1111 ok" {
		t.Errorf("mask: %q", m)
	}
	if n := call("count", "text", "a@b.co a@b.co c@d.org", "kind", "emails").(int64); n != 3 {
		t.Errorf("count: %d", n)
	}
	if n := call("count", "text", "a@b.co a@b.co c@d.org", "kind", "email", "unique", true).(int64); n != 2 {
		t.Errorf("count unique: %d", n)
	}
	want(t, call("extract", "kind", "Emails", "text", "x@y.com x@y.com b@a.org", "sort", true).([]string), "b@a.org", "x@y.com")
	if _, err := tools.Call("extract", "extract", map[string]any{"kind": "zzz", "text": "x"}); err == nil {
		t.Error("unknown kind")
	}
	if k := call("kinds", "filter", "ip").([]string); len(k) < 2 {
		t.Errorf("kinds: %v", k)
	}
	if len(call("kinds").([]string)) < 50 {
		t.Error("expected many kinds")
	}
	want(t, call("lines_containing", "text", "ok\nERROR x\nok", "substr", "error", "ignore_case", true).([]string), "ERROR x")
	want(t, call("lines_containing", "text", "ok\nERROR x\nok2", "substr", "ok", "invert", true).([]string), "ERROR x")
}

func TestLimitsAndGarbage(t *testing.T) {
	want(t, ex(t, "numbers", "1 2 3 4 5", "limit", int64(2)), "1", "2")
	want(t, ex(t, "words", "b a c", "sort", true), "a", "b", "c")
	big := strings.Repeat("a", 9<<20)
	if _, err := tools.Call("extract", "emails", map[string]any{"text": big}); err == nil {
		t.Error("oversized text must fail")
	}
	junk := []string{"", "\x00\xff\xfe", "@", "#", "[", "{", "<", "%", "$", "::", "http://", "\\\\", "C:\\", "((((", strings.Repeat("{", 5000), strings.Repeat("[\"", 3000), strings.Repeat("a.", 3000), "💥" + strings.Repeat("9", 4000), "\u202e\ufeff"}
	for _, f := range tools.All() {
		if f.Group != "extract" {
			continue
		}
		for _, j := range junk {
			named := map[string]any{}
			for _, p := range f.Params {
				switch p.Type {
				case "str":
					named[p.Name] = j
				case "list[str]":
					named[p.Name] = []string{"emails"}
				}
			}
			if f.Name == "extract" || f.Name == "count" || f.Name == "redact" {
				named["kind"] = "emails"
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s(%q) panicked: %v", f.Key(), j, r)
					}
				}()
				f.Invoke(named)
			}()
		}
	}
}

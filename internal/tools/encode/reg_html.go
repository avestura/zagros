package encode

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/avestura/zagros/internal/tools"
	xhtml "golang.org/x/net/html"
)

// htmlEntityTable is the HTML 4 named character reference set (name:codepoint).
const htmlEntityTable = "nbsp:160 iexcl:161 cent:162 pound:163 curren:164 yen:165 brvbar:166 sect:167 uml:168 copy:169 ordf:170 laquo:171 not:172 shy:173 reg:174 macr:175 deg:176 plusmn:177 sup2:178 sup3:179 acute:180 micro:181 para:182 middot:183 cedil:184 sup1:185 ordm:186 raquo:187 frac14:188 frac12:189 frac34:190 iquest:191 " +
	"Agrave:192 Aacute:193 Acirc:194 Atilde:195 Auml:196 Aring:197 AElig:198 Ccedil:199 Egrave:200 Eacute:201 Ecirc:202 Euml:203 Igrave:204 Iacute:205 Icirc:206 Iuml:207 ETH:208 Ntilde:209 Ograve:210 Oacute:211 Ocirc:212 Otilde:213 Ouml:214 times:215 Oslash:216 Ugrave:217 Uacute:218 Ucirc:219 Uuml:220 Yacute:221 THORN:222 szlig:223 " +
	"agrave:224 aacute:225 acirc:226 atilde:227 auml:228 aring:229 aelig:230 ccedil:231 egrave:232 eacute:233 ecirc:234 euml:235 igrave:236 iacute:237 icirc:238 iuml:239 eth:240 ntilde:241 ograve:242 oacute:243 ocirc:244 otilde:245 ouml:246 divide:247 oslash:248 ugrave:249 uacute:250 ucirc:251 uuml:252 yacute:253 thorn:254 yuml:255 " +
	"OElig:338 oelig:339 Scaron:352 scaron:353 Yuml:376 fnof:402 circ:710 tilde:732 " +
	"Alpha:913 Beta:914 Gamma:915 Delta:916 Epsilon:917 Zeta:918 Eta:919 Theta:920 Iota:921 Kappa:922 Lambda:923 Mu:924 Nu:925 Xi:926 Omicron:927 Pi:928 Rho:929 Sigma:931 Tau:932 Upsilon:933 Phi:934 Chi:935 Psi:936 Omega:937 " +
	"alpha:945 beta:946 gamma:947 delta:948 epsilon:949 zeta:950 eta:951 theta:952 iota:953 kappa:954 lambda:955 mu:956 nu:957 xi:958 omicron:959 pi:960 rho:961 sigmaf:962 sigma:963 tau:964 upsilon:965 phi:966 chi:967 psi:968 omega:969 thetasym:977 upsih:978 piv:982 " +
	"ensp:8194 emsp:8195 thinsp:8201 zwnj:8204 zwj:8205 lrm:8206 rlm:8207 ndash:8211 mdash:8212 lsquo:8216 rsquo:8217 sbquo:8218 ldquo:8220 rdquo:8221 bdquo:8222 dagger:8224 Dagger:8225 bull:8226 hellip:8230 permil:8240 prime:8242 Prime:8243 lsaquo:8249 rsaquo:8250 oline:8254 frasl:8260 euro:8364 image:8465 weierp:8472 real:8476 trade:8482 alefsym:8501 " +
	"larr:8592 uarr:8593 rarr:8594 darr:8595 harr:8596 crarr:8629 lArr:8656 uArr:8657 rArr:8658 dArr:8659 hArr:8660 " +
	"forall:8704 part:8706 exist:8707 empty:8709 nabla:8711 isin:8712 notin:8713 ni:8715 prod:8719 sum:8721 minus:8722 lowast:8727 radic:8730 prop:8733 infin:8734 ang:8736 and:8743 or:8744 cap:8745 cup:8746 int:8747 there4:8756 sim:8764 cong:8773 asymp:8776 ne:8800 equiv:8801 le:8804 ge:8805 sub:8834 sup:8835 nsub:8836 sube:8838 supe:8839 oplus:8853 otimes:8855 perp:8869 sdot:8901 " +
	"lceil:8968 rceil:8969 lfloor:8970 rfloor:8971 loz:9674 spades:9824 clubs:9827 hearts:9829 diams:9830"

var (
	entByRune = map[rune]string{}
	entByName = map[string]rune{}
)

func init() {
	for _, kv := range strings.Fields(htmlEntityTable) {
		i := strings.IndexByte(kv, ':')
		n, _ := strconv.Atoi(kv[i+1:])
		name := kv[:i]
		// keep only names that the standard library unescapes to the same rune
		if html.UnescapeString("&"+name+";") != string(rune(n)) {
			continue
		}
		entByRune[rune(n)] = name
		entByName[name] = rune(n)
	}
}

var entityPattern = regexp.MustCompile(`&(#[0-9]{1,8}|#[xX][0-9a-fA-F]{1,6}|[A-Za-z][A-Za-z0-9]{1,31});`)

// htmlEscapeAll escapes the markup characters and every non-ASCII character,
// preferring named entities.
func htmlEscapeAll(s string, hexNum bool) string {
	var sb strings.Builder
	for _, r := range s {
		switch {
		case r == '&':
			sb.WriteString("&amp;")
		case r == '<':
			sb.WriteString("&lt;")
		case r == '>':
			sb.WriteString("&gt;")
		case r == '"':
			sb.WriteString("&quot;")
		case r == '\'':
			sb.WriteString("&#39;")
		case r < 0x80:
			sb.WriteRune(r)
		default:
			if n, ok := entByRune[r]; ok {
				sb.WriteString("&" + n + ";")
			} else {
				sb.WriteString(numericEntity(r, hexNum))
			}
		}
	}
	return sb.String()
}

func numericEntity(r rune, hexNum bool) string {
	if hexNum {
		return fmt.Sprintf("&#x%X;", r)
	}
	return fmt.Sprintf("&#%d;", r)
}

func xmlUnescape(s string) string {
	return entityPattern.ReplaceAllStringFunc(s, func(m string) string {
		body := m[1 : len(m)-1]
		if body[0] == '#' {
			var n int64
			var err error
			if body[1] == 'x' || body[1] == 'X' {
				n, err = strconv.ParseInt(body[2:], 16, 32)
			} else {
				n, err = strconv.ParseInt(body[1:], 10, 32)
			}
			if err != nil || n <= 0 || n > 0x10FFFF || n >= 0xD800 && n <= 0xDFFF {
				return m
			}
			return string(rune(n))
		}
		switch body {
		case "amp":
			return "&"
		case "lt":
			return "<"
		case "gt":
			return ">"
		case "quot":
			return "\""
		case "apos":
			return "'"
		}
		return m
	})
}

var blockTags = map[string]bool{"p": true, "div": true, "br": true, "li": true, "tr": true, "h1": true, "h2": true, "h3": true, "h4": true, "h5": true, "h6": true, "ul": true, "ol": true, "table": true, "section": true, "article": true, "header": true, "footer": true, "blockquote": true, "pre": true, "hr": true, "dt": true, "dd": true, "title": true}

func stripTags(s string, breaks bool) string {
	z := xhtml.NewTokenizer(strings.NewReader(s))
	var sb strings.Builder
	skip := ""
	nl := func() {
		if breaks && sb.Len() > 0 && !strings.HasSuffix(sb.String(), "\n") {
			sb.WriteByte('\n')
		}
	}
	for {
		switch z.Next() {
		case xhtml.ErrorToken:
			return strings.TrimSpace(sb.String())
		case xhtml.TextToken:
			if skip == "" {
				sb.Write(z.Text())
			}
		case xhtml.StartTagToken, xhtml.SelfClosingTagToken:
			name, _ := z.TagName()
			n := string(name)
			if skip == "" && (n == "script" || n == "style") {
				skip = n
			}
			if blockTags[n] {
				nl()
			}
		case xhtml.EndTagToken:
			name, _ := z.TagName()
			n := string(name)
			if n == skip {
				skip = ""
			}
			if blockTags[n] && skip == "" {
				nl()
			}
		}
	}
}

var tagPattern = regexp.MustCompile(`<[A-Za-z/!][^<>]*>`)

func init() {
	reg(tools.Func{
		Name: "html_escape", Title: "HTML escape", Order: 5,
		Doc:     "Escapes the five markup-significant characters & < > \" ' as &amp; &lt; &gt; &quot; &#39; so text is safe in HTML text and attribute values. quotes=false leaves quotes alone (text content only). Non-ASCII characters are kept; use html_escape_all to escape them too.",
		Params:  []tools.Param{pReq("text", "Text to escape."), pBool("quotes", "Also escape \" and '.", true)},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			s := a.Str("text")
			if a.Bool("quotes") {
				return html.EscapeString(s), nil
			}
			return strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;").Replace(s), nil
		},
		Example:  `encode.html_escape("<a href=\"x\">Tom & Jerry</a>")`,
		Tags:     tagsWith("html", "escape", "entities", "xss", "sanitize", "&amp;", "htmlspecialchars", "encode"),
		Keywords: []string{"html", "esc", "escape"},
	})
	reg(tools.Func{
		Name: "html_unescape", Title: "HTML unescape", Order: 6,
		Doc:     "Decodes HTML character references: named entities (the full HTML5 table, with or without the trailing semicolon where HTML allows), decimal &#169; and hex &#xA9; numeric references. Unknown entities are left as they are.",
		Params:  []tools.Param{pReq("text", "Text containing entities.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return html.UnescapeString(a.Str("text")), nil },
		Accepts: func(s string) bool {
			return entityPattern.MatchString(s) && html.UnescapeString(s) != s
		},
		Example:  `encode.html_unescape("Tom &amp; Jerry &copy; &#x1F600;")`,
		Tags:     tagsWith("html", "unescape", "entities", "decode", "&amp;", "htmlentities", "html_entity_decode"),
		Keywords: []string{"html", "esc", "escape", "entities"},
	})
	reg(tools.Func{
		Name: "html_escape_all", Title: "HTML escape (all non-ASCII)",
		Doc:     "Escapes markup characters and every non-ASCII character: named entities where an HTML 4 name exists (&eacute; &euro; &hearts;), numeric references otherwise (&#128512;, or &#x1F600; with hex=true). Output is pure ASCII.",
		Params:  []tools.Param{pReq("text", "Text to escape."), pBool("hex", "Use hexadecimal numeric references for characters without a name.", false)},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return htmlEscapeAll(a.Str("text"), a.Bool("hex")), nil },
		Example: `encode.html_escape_all("café © 😀")`,
		Tags:    tagsWith("html", "escape", "entities", "named entities", "htmlentities", "ascii only", "non-ascii", "encode"),

		Keywords: []string{"html", "esc", "entities"},
	})
	reg(tools.Func{
		Name: "html_numeric_entities", Title: "HTML numeric entities",
		Doc:     "Converts characters to numeric character references (&#233; decimal, &#xE9; with hex=true). By default converts non-ASCII characters and the markup characters & < > \" '; all=true converts every character including letters.",
		Params:  []tools.Param{pReq("text", "Text to convert."), pBool("hex", "Hexadecimal references.", false), pBool("all", "Convert every character, even plain ASCII.", false)},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			var sb strings.Builder
			for _, r := range a.Str("text") {
				if a.Bool("all") || r >= 0x80 || strings.ContainsRune("&<>\"'", r) {
					sb.WriteString(numericEntity(r, a.Bool("hex")))
				} else {
					sb.WriteRune(r)
				}
			}
			return sb.String(), nil
		},
		Example:  `encode.html_numeric_entities("Hi é", hex=True)`,
		Tags:     tagsWith("html", "numeric", "entities", "&#", "decimal", "hex", "character reference", "ncr"),
		Keywords: []string{"html", "entities"},
	})
	reg(tools.Func{
		Name: "xml_escape", Title: "XML escape",
		Doc:     "Escapes & < > \" ' as &amp; &lt; &gt; &quot; &apos; for XML text and attributes, and removes characters that are not allowed in XML 1.0 (most control characters).",
		Params:  []tools.Param{pReq("text", "Text to escape.")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			var sb strings.Builder
			for _, r := range strings.ToValidUTF8(a.Str("text"), "") {
				switch r {
				case '&':
					sb.WriteString("&amp;")
				case '<':
					sb.WriteString("&lt;")
				case '>':
					sb.WriteString("&gt;")
				case '"':
					sb.WriteString("&quot;")
				case '\'':
					sb.WriteString("&apos;")
				default:
					if r == '\t' || r == '\n' || r == '\r' || r >= 0x20 && r <= 0xD7FF || r >= 0xE000 && r <= 0xFFFD || r >= 0x10000 {
						sb.WriteRune(r)
					}
				}
			}
			return sb.String(), nil
		},
		Example:  `encode.xml_escape("<a b='1'>x & y</a>")`,
		Tags:     tagsWith("xml", "escape", "entities", "&apos;", "cdata", "encode"),
		Keywords: []string{"xml", "esc", "escape"},
	})
	reg(tools.Func{
		Name: "xml_unescape", Title: "XML unescape",
		Doc:     "Decodes the five predefined XML entities (&amp; &lt; &gt; &quot; &apos;) and numeric references (&#65; &#x41;). Other named entities and invalid numbers are left untouched; use html_unescape for HTML names.",
		Params:  []tools.Param{pReq("text", "Text containing XML entities.")},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return xmlUnescape(a.Str("text")), nil },
		Accepts: func(s string) bool { return xmlUnescape(s) != s },
		Example: `encode.xml_unescape("&lt;a&gt; &amp; &#65;")`,
		Tags:    tagsWith("xml", "unescape", "entities", "&apos;", "decode"),

		Keywords: []string{"xml", "esc", "escape"},
	})
	reg(tools.Func{
		Name: "strip_html_tags", Title: "Strip HTML tags",
		Doc:     "Removes HTML tags and comments, drops <script> and <style> contents, decodes entities and keeps the text. keep_breaks=true (default) turns block tags such as <p>, <br>, <li>, <div> into line breaks. Plain text without tags is returned trimmed.",
		Params:  []tools.Param{pReq("text", "HTML text."), pBool("keep_breaks", "Insert line breaks at block-level tags.", true)},
		Returns: "str",
		Fn:      func(a tools.Args) (any, error) { return stripTags(a.Str("text"), a.Bool("keep_breaks")), nil },
		Accepts: func(s string) bool { return tagPattern.MatchString(s) },
		Example: `encode.strip_html_tags("<p>Hello <b>world</b> &amp; all</p>")`,
		Tags:    tagsWith("html", "strip tags", "remove tags", "plain text", "sanitize", "striptags", "text content"),

		Keywords: []string{"html", "strip", "tags"},
	})
	reg(tools.Func{
		Name: "html_entity_name", Title: "HTML entity of a character",
		Doc:     "Returns the HTML entity for a character: the named form (\"&copy;\", \"&amp;\", \"&euro;\") when an HTML 4 name exists, otherwise the decimal numeric reference (\"&#128512;\"). Uses the first character if several are given; empty input is an error.",
		Params:  []tools.Param{pReq("char", "A single character.")},
		Returns: "str",
		Fn: func(a tools.Args) (any, error) {
			r, n := utf8.DecodeRuneInString(a.Str("char"))
			if n == 0 {
				return nil, fmt.Errorf("char is empty")
			}
			switch r {
			case '&':
				return "&amp;", nil
			case '<':
				return "&lt;", nil
			case '>':
				return "&gt;", nil
			case '"':
				return "&quot;", nil
			}
			if name, ok := entByRune[r]; ok {
				return "&" + name + ";", nil
			}
			return numericEntity(r, false), nil
		},
		Example:  `encode.html_entity_name("©")`,
		Tags:     tagsWith("html", "entity", "named entity", "character reference", "char to entity", "lookup"),
		Keywords: []string{"html", "entity"},
	})
}

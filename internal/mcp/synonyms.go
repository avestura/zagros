package mcp

import "strings"

// synonymGroups lists words that mean the same thing when someone searches
// the API. A query word is expanded to its group at a lower weight than the
// word itself. Keep the groups tight: words that differ in meaning (encode and
// encrypt, hash and encrypt, extract an archive and extract text) must not
// share a group.
var synonymGroups = [][]string{
	{"unzip", "decompress", "uncompress", "unpack", "inflate", "untar", "unarchive"},
	{"zip", "compress", "archive", "deflate", "gzip", "tar", "pack"},
	{"encrypt", "encipher", "seal"},
	{"decrypt", "decipher", "unseal"},
	{"hash", "digest", "checksum", "fingerprint"},
	{"uuid", "guid"},
	{"random", "rng", "randomize", "randomise"},
	{"uppercase", "upper", "capitals"},
	{"lowercase", "lower", "downcase"},
	{"trim", "strip"},
	{"remove", "delete", "erase", "drop"},
	{"find", "search", "locate", "lookup"},
	{"replace", "substitute"},
	{"duplicate", "dedupe", "deduplicate", "unique", "distinct", "uniq"},
	{"sort", "order", "alphabetize"},
	{"length", "size", "len"},
	{"timestamp", "epoch", "unix"},
	{"timezone", "tz"},
	{"url", "uri", "link", "hyperlink"},
	{"email", "mail"},
	{"password", "passwd", "pwd"},
	{"base64", "b64"},
	{"hex", "hexadecimal"},
	{"binary", "bin"},
	{"octal", "oct"},
	{"percent", "percentage", "pct"},
	{"average", "mean", "avg"},
	{"calculate", "compute", "calc", "evaluate", "eval"},
	{"join", "concatenate", "concat", "merge", "combine"},
	{"compare", "diff", "difference"},
	{"escape", "sanitize", "sanitise"},
	{"unescape", "unquote"},
	{"phone", "telephone"},
	{"clipboard", "pasteboard", "paste"},
	{"read", "load", "open"},
	{"write", "save", "persist"},
	{"store", "storage", "remember", "persist"},
	{"fetch", "download", "request"},
	{"http", "https", "web"},
	{"notify", "notification", "toast", "alert"},
	{"run", "execute", "launch", "spawn"},
	{"env", "environment"},
	{"mask", "redact", "censor", "obfuscate"},
	{"folder", "directory", "dir"},
	{"parse", "interpret"},
	{"convert", "transform", "turn", "change"},
	{"regex", "regexp", "regular"},
	{"split", "divide", "tokenize"},
	{"duration", "span", "interval"},
	{"color", "colour"},
	{"number", "numeric", "digit"},
	{"text", "string"},
}

var synonymOf = func() map[string][]string {
	m := map[string][]string{}
	for _, g := range synonymGroups {
		for _, w := range g {
			for _, o := range g {
				if o != w {
					m[w] = appendUnique(m[w], o)
				}
			}
		}
	}
	return m
}()

func appendUnique(s []string, v string) []string {
	for _, x := range s {
		if x == v {
			return s
		}
	}
	return append(s, v)
}

// Synonyms returns the words the search treats as meaning the same as w.
func Synonyms(w string) []string { return append([]string(nil), synonymOf[strings.ToLower(w)]...) }

var stopwords = map[string]bool{}

func init() {
	for _, w := range strings.Fields("a an the to into of in on for with from and or my some this that is it as at by be how do i me you your can want need using use make get what which then so") {
		stopwords[w] = true
	}
}

// stem trims the plural, verb and silent-e endings that would stop
// "decompressing" or "emails" from matching "decompress" and "email". Both
// sides of every comparison are stemmed, so it only has to be consistent.
func stem(w string) string {
	switch {
	case len(w) > 6 && strings.HasSuffix(w, "ing"):
		w = w[:len(w)-3]
	case len(w) > 5 && strings.HasSuffix(w, "ed"):
		w = w[:len(w)-2]
	case len(w) > 4 && strings.HasSuffix(w, "ies"):
		w = w[:len(w)-3] + "y"
	case len(w) > 4 && (strings.HasSuffix(w, "ches") || strings.HasSuffix(w, "shes") || strings.HasSuffix(w, "xes") || strings.HasSuffix(w, "sses")):
		w = w[:len(w)-2]
	case len(w) > 3 && strings.HasSuffix(w, "s") && !strings.HasSuffix(w, "ss"):
		w = w[:len(w)-1]
	}
	if len(w) > 4 && strings.HasSuffix(w, "e") {
		w = w[:len(w)-1]
	}
	return w
}

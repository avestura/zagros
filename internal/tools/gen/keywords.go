package gen

import "strings"

// GroupKeywords are the Keywords of each module's docstring: the words and
// task phrases someone would use before they know the module exists. A group
// without an entry falls back to its name and title.
var GroupKeywords = map[string]string{
	"text":     "string, case, uppercase, lowercase, camelCase, snake_case, slug, palindrome, anagram, trim, pad, truncate, similarity, levenshtein, unicode, codepoint, invisible characters",
	"lines":    "sort lines, remove duplicate lines, dedupe, join lines, split text into lines, filter lines, csv row, list, column, diff lines, line endings, markdown table",
	"num":      "number, base conversion, hex, binary, octal, decimal, roman numerals, number to words, prime, factorial, gcd, bit operations, ieee754, checksum digit, luhn, iban, isbn, human bytes",
	"math":     "calculator, expression, evaluate, percentage, statistics, mean, median, stdev, algebra, geometry, area, volume, finance, loan, interest, discount, tip, fraction, unit price",
	"encode":   "encode, decode, escape, unescape, base64, base32, base58, base85, hex, url encode, percent encoding, html entities, punycode, charset, mojibake, quote string, jwt, morse, data uri, turn text into base64",
	"hash":     "hash, digest, checksum, fingerprint, md5, sha1, sha256, sha512, crc32, hmac, password hash, bcrypt, argon2, pbkdf2, scrypt, totp, one-time password, file integrity",
	"cipher":   "encrypt, decrypt, cipher, aes, chacha20, rsa, ed25519, ecdsa, sign, verify signature, keypair, pem, xor, caesar, rot13, vigenere, secret, symmetric, asymmetric",
	"rand":     "random, generate random, dice, password generator, passphrase, token, pick, shuffle, sample, random color, random date, fake data, lorem, coin flip, lottery",
	"uuid":     "uuid, guid, unique id, ulid, ksuid, snowflake, nanoid, object id, typeid, identifier, generate id, parse id, v4, v7",
	"regex":    "regex, regular expression, pattern, match, find all, replace, split, capture group, validate pattern, glob, wildcard, explain regex, sed",
	"net":      "network, ip address, cidr, subnet, netmask, ipv4, ipv6, mac address, port, http status, mime type, content type, url parse, domain, hostname, email validate, bandwidth",
	"dt":       "date, time, datetime, timestamp, unix time, epoch, time zone, timezone, duration, add days, date difference, age, weekday, calendar, cron, iso 8601, format date, business days",
	"extract":  "extract, pull out, find all, scrape, emails, urls, domains, ip addresses, phone numbers, numbers, dates, hashtags, mentions, redact, mask, personal data, pii",
	"yaml":     "yaml, yml, parse yaml, read yaml, write yaml, config file, kubernetes manifest, docker compose, serialize",
	"toml":     "toml, parse toml, read toml, write toml, config file, cargo, pyproject, serialize",
	"csv":      "csv, tsv, spreadsheet, parse csv, read csv, write csv, delimiter, rows, columns, table, export, import",
	"xml":      "xml, parse xml, read xml, write xml, svg, rss, atom, soap, xpath, element tree, serialize",
	"html":     "html, parse html, css selector, query selector, scrape, web page, dom, extract text, links, attributes, strip tags",
	"url":      "url, uri, link, parse url, build url, join url, query string, query parameters, normalize url, scheme host path",
	"path":     "path, file path, filename, extension, basename, dirname, join path, split path, normalize path, windows path, posix path",
	"template": "template, render, text template, placeholders, mail merge, format string, go template, interpolate",
	"jwt":      "jwt, json web token, bearer token, decode token, sign token, verify token, claims, hs256, rs256, oauth",
}

func groupKeywords(group string) string {
	if kw := strings.TrimSpace(GroupKeywords[group]); kw != "" {
		return kw
	}
	if t := GroupTitle[group]; t != "" {
		return group + ", " + strings.ToLower(t)
	}
	return group
}

// TypeNote is the header sentence every stub carries so nobody mistakes it
// for runnable code.
func TypeNote(group string) string {
	return "Type description, not runnable Starlark: do not `import` anything. The name `" + group + "` is predeclared in main.star; call `" + group + ".<name>(...)` directly."
}

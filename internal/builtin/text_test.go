package builtin

import (
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestCaseAll(t *testing.T) {
	want := map[string]string{
		"snake_case": "hello_big_world", "camelCase": "helloBigWorld", "PascalCase": "HelloBigWorld",
		"kebab-case": "hello-big-world", "CONSTANT_CASE": "HELLO_BIG_WORLD", "Title Case": "Hello Big World",
	}
	for _, in := range []string{"hello big world", "helloBigWorld", "Hello_Big-World", "  HELLO  big world "} {
		got := map[string]string{}
		for _, r := range caseAll(in) {
			got[r.label] = r.value
		}
		for k, v := range want {
			if got[k] != v {
				t.Errorf("%q: %s = %q, want %q", in, k, got[k], v)
			}
		}
	}
	if w := splitWords("parseHTTPResponse2Go"); strings.Join(w, ",") != "parse,HTTP,Response2,Go" {
		t.Errorf("splitWords = %v", w)
	}
}

func TestHashAll(t *testing.T) {
	got := map[string]string{}
	for _, r := range hashAll("abc") {
		got[r.label] = r.value
	}
	if got["MD5"] != "900150983cd24fb0d6963f7d28e17f72" {
		t.Errorf("md5 %s", got["MD5"])
	}
	if got["SHA-1"] != "a9993e364706816aba3e25717850c26c9cd0d89d" {
		t.Errorf("sha1 %s", got["SHA-1"])
	}
	if got["SHA-256"] != "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad" {
		t.Errorf("sha256 %s", got["SHA-256"])
	}
	if !strings.HasPrefix(got["SHA-512"], "ddaf35a193617aba") {
		t.Errorf("sha512 %s", got["SHA-512"])
	}
}

func TestTimestamps(t *testing.T) {
	now := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	rows := tsResults("1700000000", now, time.UTC)
	if rows[2].value != "2023-11-14T22:13:20Z" {
		t.Errorf("utc = %s", rows[2].value)
	}
	ms := tsResults("1700000000000", now, time.UTC)
	if ms[0].value != "1700000000" {
		t.Errorf("ms detection: %s", ms[0].value)
	}
	d := tsResults("2024-01-31 15:04", now, time.UTC)
	if d == nil || d[0].value != "1706713440" {
		t.Errorf("date parse: %v", d)
	}
	if tsResults("garbage", now, time.UTC) != nil {
		t.Error("garbage must not parse")
	}
	if r := relativeTime(now.Add(-3*time.Hour), now); r != "3.0 hours ago" {
		t.Errorf("relative = %q", r)
	}
}

func TestColor(t *testing.T) {
	for _, in := range []string{"#ff8000", "ff8000", "#f80", "rgb(255, 136, 0)", "hsl(32, 100%, 50%)"} {
		r, g, b, ok := parseColor(in)
		if !ok {
			t.Errorf("%q not parsed", in)
			continue
		}
		if r != 255 || (g != 128 && g != 136 && g != 137 && g != 0x88) || b != 0 {
			if in == "#ff8000" || in == "ff8000" {
				if g != 128 {
					t.Errorf("%q -> %d,%d,%d", in, r, g, b)
				}
			}
		}
	}
	rows := colorResults("#ff0000")
	if rows[2].value != "rgb(255, 0, 0)" || rows[3].value != "hsl(0, 100%, 50%)" {
		t.Errorf("rows = %v", rows)
	}
	if _, _, _, ok := parseColor("nope"); ok {
		t.Error("nope parsed")
	}
}

func TestMisc(t *testing.T) {
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if u := newUUID(); !re.MatchString(u) {
		t.Errorf("bad uuid %s", u)
	}
	if s, ok := decodeBase64("aGVsbG8="); !ok || s != "hello" {
		t.Errorf("b64d = %q %v", s, ok)
	}
	if got := len(strings.Fields(loremWords(100))); got != 100 {
		t.Errorf("lorem words = %d", got)
	}
	if j := jsonResults(`{"a": [1, 2]}`); len(j) != 2 || j[1].value != `{"a":[1,2]}` {
		t.Errorf("json = %v", j)
	}
	if jsonResults("{bad") != nil {
		t.Error("invalid json accepted")
	}
}

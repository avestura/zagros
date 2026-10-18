package randx

import (
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "rand") }

func call(t *testing.T, name string, kv ...any) any {
	t.Helper()
	m := map[string]any{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	v, err := tools.Call("rand", name, m)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return v
}

func callErr(name string, kv ...any) error {
	m := map[string]any{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	_, err := tools.Call("rand", name, m)
	return err
}

func TestSeedDeterministic(t *testing.T) {
	a := call(t, "int", "min", int64(1), "max", int64(1000000), "seed", int64(42))
	b := call(t, "int", "min", int64(1), "max", int64(1000000), "seed", int64(42))
	if a != b {
		t.Fatalf("seeded results differ: %v %v", a, b)
	}
	x := call(t, "shuffle", "items", []string{"a", "b", "c", "d", "e", "f"}, "seed", int64(7)).([]string)
	y := call(t, "shuffle", "items", []string{"a", "b", "c", "d", "e", "f"}, "seed", int64(7)).([]string)
	if strings.Join(x, "") != strings.Join(y, "") {
		t.Fatal("shuffle not reproducible")
	}
}

func TestIntRange(t *testing.T) {
	for i := 0; i < 500; i++ {
		v := call(t, "int", "min", int64(-3), "max", int64(3)).(int64)
		if v < -3 || v > 3 {
			t.Fatalf("out of range %d", v)
		}
	}
	if callErr("int", "min", int64(5), "max", int64(1)) == nil {
		t.Fatal("want error")
	}
	if v := call(t, "int", "args", "6").(int64); v < 1 || v > 6 {
		t.Fatal(v)
	}
}

func TestIntsUnique(t *testing.T) {
	v := call(t, "ints", "count", int64(10), "min", int64(1), "max", int64(10), "unique", true).([]float64)
	seen := map[float64]bool{}
	for _, x := range v {
		seen[x] = true
	}
	if len(seen) != 10 {
		t.Fatalf("not unique: %v", v)
	}
	if callErr("ints", "count", int64(11), "min", int64(1), "max", int64(10), "unique", true) == nil {
		t.Fatal("want error")
	}
	if callErr("ints", "count", int64(10001)) == nil {
		t.Fatal("cap")
	}
}

func TestDice(t *testing.T) {
	m := call(t, "dice", "spec", "4d6kh3+2", "seed", int64(3)).(map[string]string)
	total, _ := strconv.Atoi(m["total"])
	if total < 5 || total > 20 {
		t.Fatalf("bad total %v", m)
	}
	if n := len(strings.Split(m["rolls"], ", ")); n != 3 {
		t.Fatalf("kept rolls %d: %v", n, m)
	}
	if !strings.Contains(m["detail"], "(") {
		t.Fatalf("dropped die not marked: %v", m)
	}
	s := call(t, "roll_stats", "spec", "2d6+3").(map[string]string)
	if s["min"] != "5" || s["max"] != "15" || s["mean"] != "10" {
		t.Fatalf("stats %v", s)
	}
	s = call(t, "roll_stats", "spec", "4d6kh3").(map[string]string)
	mean, _ := strconv.ParseFloat(s["mean"], 64)
	if s["min"] != "3" || s["max"] != "18" || mean < 12.0 || mean > 13.0 { // true mean 12.24
		t.Fatalf("stats %v", s)
	}
	s = call(t, "roll_stats", "spec", "d6!").(map[string]string)
	if s["max"] != "unbounded" {
		t.Fatalf("%v", s)
	}
	for _, bad := range []string{"", "2x6", "0d6", "d0", "1001d6", "d1!", "+"} {
		if callErr("dice", "spec", bad) == nil {
			t.Errorf("spec %q should fail", bad)
		}
	}
	m = call(t, "dice", "spec", "d2!-1d4", "seed", int64(1)).(map[string]string)
	if m["total"] == "" {
		t.Fatal("no total")
	}
	for i := 0; i < 200; i++ {
		v := call(t, "d20").(int64)
		if v < 1 || v > 20 {
			t.Fatal(v)
		}
	}
	if v := call(t, "d6", "modifier", int64(100)).(int64); v < 101 {
		t.Fatal(v)
	}
}

func TestPickAndWeights(t *testing.T) {
	if callErr("choice", "items", []string{}) == nil {
		t.Fatal("empty choice")
	}
	p := call(t, "pick", "items", []string{"a", "b", "c"}, "n", int64(3)).([]string)
	if len(p) != 3 {
		t.Fatal(p)
	}
	if callErr("pick", "items", []string{"a"}, "n", int64(2)) == nil {
		t.Fatal("unique overflow")
	}
	for i := 0; i < 50; i++ {
		if call(t, "weighted_choice", "items", []string{"a", "b"}, "weights", []float64{0, 1}) != "b" {
			t.Fatal("zero weight picked")
		}
	}
	if callErr("weighted_choice", "items", []string{"a"}, "weights", []float64{1, 2}) == nil {
		t.Fatal("length mismatch")
	}
	w := call(t, "shuffle_text", "text", "one two three", "by", "words").(string)
	if len(strings.Fields(w)) != 3 {
		t.Fatal(w)
	}
	c := call(t, "coin_flips", "n", int64(50)).(map[string]string)
	h, _ := strconv.Atoi(c["heads"])
	tl, _ := strconv.Atoi(c["tails"])
	if h+tl != 50 || len(c["sequence"]) != 50 {
		t.Fatal(c)
	}
	if call(t, "random_item_from_text", "text", "x, y, z", "seed", int64(1)) == "" {
		t.Fatal("empty")
	}
}

func TestPassword(t *testing.T) {
	for i := 0; i < 100; i++ {
		s := call(t, "password", "length", int64(12)).(string)
		if len(s) != 12 {
			t.Fatal(s)
		}
		for _, re := range []string{`[a-z]`, `[A-Z]`, `[0-9]`, `[^a-zA-Z0-9]`} {
			if !regexp.MustCompile(re).MatchString(s) {
				t.Fatalf("%q lacks %s", s, re)
			}
		}
	}
	s := call(t, "password", "length", int64(200), "symbols", false, "exclude_ambiguous", true).(string)
	if len(s) != 128 || strings.ContainsAny(s, "0O1lI") || regexp.MustCompile(`[^a-zA-Z0-9]`).MatchString(s) {
		t.Fatal(s)
	}
	if s := call(t, "password", "args", "30 nosym noupper nonum").(string); len(s) != 30 || strings.Trim(s, pwLower) != "" {
		t.Fatal(s)
	}
	if got := call(t, "passwords", "count", int64(7)).([]string); len(got) != 7 {
		t.Fatal(got)
	}
	pp := call(t, "passphrase", "words", int64(4), "capitalize", true, "number", true).(string)
	if !regexp.MustCompile(`^([A-Z][a-z]+-){4}\d+$`).MatchString(pp) {
		t.Fatal(pp)
	}
	if pin := call(t, "pin", "digits", int64(8)).(string); !regexp.MustCompile(`^\d{8}$`).MatchString(pin) {
		t.Fatal(pin)
	}
}

func TestTokensAndStrings(t *testing.T) {
	if s := call(t, "token_hex", "bytes", int64(16)).(string); len(s) != 32 {
		t.Fatal(s)
	}
	if s := call(t, "token_base64url", "bytes", int64(32)).(string); len(s) != 43 || strings.ContainsAny(s, "+/=") {
		t.Fatal(s)
	}
	if s := call(t, "token_base32", "bytes", int64(5)).(string); len(s) != 8 {
		t.Fatal(s)
	}
	if s := call(t, "string", "length", int64(40), "charset", "xyz", "seed", int64(5)).(string); len(s) != 40 || strings.Trim(s, "xyz") != "" {
		t.Fatal(s)
	}
	if s := call(t, "string", "length", int64(20), "charset", "hex").(string); !regexp.MustCompile(`^[0-9a-f]{20}$`).MatchString(s) {
		t.Fatal(s)
	}
	if callErr("string", "length", int64(100001)) == nil {
		t.Fatal("cap")
	}
	if callErr("token_hex", "bytes", int64(0)) == nil {
		t.Fatal("zero bytes")
	}
}

func TestMisc(t *testing.T) {
	if c := call(t, "color_hex").(string); !regexp.MustCompile(`^#[0-9a-f]{6}$`).MatchString(c) {
		t.Fatal(c)
	}
	if got := call(t, "color_palette", "n", int64(6)).([]string); len(got) != 6 {
		t.Fatal(got)
	}
	if c := call(t, "color_hsl").(string); !strings.HasPrefix(c, "hsl(") {
		t.Fatal(c)
	}
	if ip := call(t, "ip_v4", "private", true).(string); !(strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.") || strings.HasPrefix(ip, "192.168.")) {
		t.Fatal(ip)
	}
	if ip := call(t, "ip_v4").(string); strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "127.") {
		t.Fatal(ip)
	}
	if ip := call(t, "ip_v6", "private", true).(string); !strings.HasPrefix(ip, "fd") {
		t.Fatal(ip)
	}
	if m := call(t, "mac_address", "format", "dash", "upper", true).(string); !regexp.MustCompile(`^[0-9A-F]{2}(-[0-9A-F]{2}){5}$`).MatchString(m) {
		t.Fatal(m)
	}
	if m := call(t, "mac_address").(string); m[1] != '2' && m[1] != '6' && m[1] != 'a' && m[1] != 'e' {
		t.Fatalf("not locally administered unicast: %s", m)
	}
	if v := call(t, "port", "kind", "well_known").(int64); v < 1 || v > 1023 {
		t.Fatal(v)
	}
	d := call(t, "random_date", "start", "2024-02-28", "end", "2024-03-01").(string)
	if d != "2024-02-28" && d != "2024-02-29" && d != "2024-03-01" {
		t.Fatal(d)
	}
	if call(t, "random_date", "start", "2024-02-29", "end", "2024-02-29") != "2024-02-29" {
		t.Fatal("single day")
	}
	if callErr("random_date", "start", "2025-01-01", "end", "2024-01-01") == nil {
		t.Fatal("reversed")
	}
	if tm := call(t, "random_time", "start", "09:00", "end", "09:00").(string); tm != "09:00" {
		t.Fatal(tm)
	}
	l := call(t, "lottery", "count", int64(6), "max", int64(49), "bonus", int64(1)).(map[string]string)
	if len(strings.Fields(l["numbers"])) != 6 || len(strings.Fields(l["bonus"])) != 1 {
		t.Fatal(l)
	}
	if callErr("lottery", "count", int64(10), "max", int64(5)) == nil {
		t.Fatal("pool too small")
	}
	seq := call(t, "number_sequence", "count", int64(20), "kind", "descending", "seed", int64(2)).([]float64)
	for i := 1; i < len(seq); i++ {
		if seq[i] > seq[i-1] {
			t.Fatal(seq)
		}
	}
	if e := call(t, "email_fake").(string); !strings.HasSuffix(e, "@example.com") {
		t.Fatal(e)
	}
	if call(t, "full_name", "args", "female") == "" || call(t, "name_first") == "" || call(t, "name_last") == "" {
		t.Fatal("names")
	}
	if got := call(t, "lorem_words", "count", int64(4), "capitalize", true).(string); len(strings.Fields(got)) != 4 || got[0] < 'A' || got[0] > 'Z' {
		t.Fatal(got)
	}
	if call(t, "random_emoji", "n", int64(3)) == "" || call(t, "lorem_word") == "" || call(t, "color_name") == "" || call(t, "color_rgb") == "" {
		t.Fatal("misc")
	}
	for _, f := range []string{"yes_no_maybe", "magic_8_ball", "coin_flip"} {
		if call(t, f) == "" {
			t.Fatal(f)
		}
	}
	if call(t, "spin_wheel", "text", "only") != "only" {
		t.Fatal("wheel")
	}
	if call(t, "bool", "probability", float64(1)) != true {
		t.Fatal("bool")
	}
	if f := call(t, "float", "min", float64(1), "max", float64(2), "decimals", int64(2)).(float64); f < 1 || f > 2 {
		t.Fatal(f)
	}
	if s := call(t, "bytes_hex", "count", int64(4), "seed", int64(9)); s != call(t, "bytes_hex", "count", int64(4), "seed", int64(9)) {
		t.Fatal("bytes seed")
	}
}

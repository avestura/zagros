package uuid

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/avestura/zagros/internal/tools"
	"github.com/avestura/zagros/internal/tools/toolstest"
)

func TestRegistry(t *testing.T) { toolstest.Check(t, "uuid") }

func call(t *testing.T, name string, kv ...any) any {
	t.Helper()
	v, err := callE(name, kv...)
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	return v
}

func callE(name string, kv ...any) (any, error) {
	m := map[string]any{}
	for i := 0; i < len(kv); i += 2 {
		m[kv[i].(string)] = kv[i+1]
	}
	return tools.Call("uuid", name, m)
}

func TestRFC9562Vectors(t *testing.T) {
	node := [6]byte{0x9F, 0x6B, 0xDE, 0xCE, 0xD8, 0x46}
	if got := buildV1(0x1EC9414C232AB00, 0x33C8, node).String(); got != "c232ab00-9414-11ec-b3c8-9f6bdeced846" {
		t.Errorf("v1 vector: %s", got)
	}
	if got := buildV6(0x1EC9414C232AB00, 0x33C8, node).String(); got != "1ec9414c-232a-6b00-b3c8-9f6bdeced846" {
		t.Errorf("v6 vector: %s", got)
	}
	if got := buildV7(0x017F22E279B0, 0xCC3, 0x18C4DC0C0C07398F).String(); got != "017f22e2-79b0-7cc3-98c4-dc0c0c07398f" {
		t.Errorf("v7 vector: %s", got)
	}
	// 2022-02-22 19:22:22 UTC
	want := time.Date(2022, 2, 22, 19, 22, 22, 0, time.UTC)
	for _, s := range []string{"C232AB00-9414-11EC-B3C8-9F6BDECED846", "1EC9414C-232A-6B00-B3C8-9F6BDECED846", "017F22E2-79B0-7CC3-98C4-DC0C0C07398F"} {
		m := call(t, "uuid_parse", "u", s).(map[string]string)
		ts, err := time.Parse(time.RFC3339Nano, m["timestamp"])
		if err != nil || !ts.Equal(want) {
			t.Errorf("%s: timestamp %q want %v", s, m["timestamp"], want)
		}
	}
	m := call(t, "uuid_parse", "u", "C232AB00-9414-11EC-B3C8-9F6BDECED846").(map[string]string)
	if m["version"] != "1" || m["node"] != "9f:6b:de:ce:d8:46" || m["clock_seq"] != strconv.Itoa(0x33C8) || m["variant"] != "RFC 4122/9562" {
		t.Errorf("v1 parse: %v", m)
	}
}

func TestNameBased(t *testing.T) {
	cases := []struct{ fn, ns, name, want string }{
		{"uuid_v5", "dns", "www.example.com", "2ed6657d-e927-568b-95e1-2665a8aea6a2"},
		{"uuid_v5", "dns", "python.org", "886313e1-3b8a-5372-9b90-0c9aee199e5d"},
		{"uuid_v3", "dns", "python.org", "6fa459ea-ee8a-3ca4-894e-db77e160355e"},
		{"uuid_v3", "6ba7b810-9dad-11d1-80b4-00c04fd430c8", "www.example.com", "5df41881-3aed-3515-88a7-2f4a814cf09e"},
	}
	for _, c := range cases {
		if got := call(t, c.fn, "namespace", c.ns, "name", c.name); got != c.want {
			t.Errorf("%s(%s,%s) = %v, want %s", c.fn, c.ns, c.name, got, c.want)
		}
	}
	if _, err := callE("uuid_v5", "namespace", "bogus", "name", "x"); err == nil {
		t.Error("bad namespace should fail")
	}
	ns := call(t, "namespaces").(map[string]string)
	if ns["dns"] != "6ba7b810-9dad-11d1-80b4-00c04fd430c8" || ns["x500"] != "6ba7b814-9dad-11d1-80b4-00c04fd430c8" {
		t.Errorf("%v", ns)
	}
}

func TestGenerators(t *testing.T) {
	u := call(t, "uuid_v4").(string)
	if !regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`).MatchString(u) {
		t.Errorf("v4 %s", u)
	}
	if u := call(t, "uuid_v4", "style", "braces").(string); !strings.HasPrefix(u, "{") {
		t.Error(u)
	}
	a := call(t, "uuid_v7", "time", "2022-02-22T19:22:22Z", "seed", int64(5)).(string)
	b := call(t, "uuid_v7", "time", "2022-02-22T19:22:22Z", "seed", int64(5)).(string)
	if a != b || !strings.HasPrefix(a, "017f22e2-79b0-7") {
		t.Errorf("v7 %s %s", a, b)
	}
	v1 := call(t, "uuid_v1", "time", "2022-02-22T19:22:22Z").(string)
	if !strings.HasPrefix(v1, "c232ab00-9414-11ec-") {
		t.Errorf("v1 %s", v1)
	}
	m := call(t, "uuid_parse", "u", v1).(map[string]string)
	if m["node_multicast"] != "true" {
		t.Errorf("node must carry the multicast bit: %v", m)
	}
	if v6 := call(t, "uuid_v6", "time", "2022-02-22T19:22:22Z").(string); !strings.HasPrefix(v6, "1ec9414c-232a-6b00-") {
		t.Errorf("v6 %s", v6)
	}
	if call(t, "uuid_nil") != "00000000-0000-0000-0000-000000000000" || call(t, "uuid_max") != "ffffffff-ffff-ffff-ffff-ffffffffffff" {
		t.Error("nil/max")
	}
	if v8 := call(t, "uuid_v8", "data", "00112233445566778899aabbccddeeff").(string); v8 != "00112233-4455-8677-8899-aabbccddeeff" {
		t.Errorf("v8 %s", v8)
	}
	bulk := call(t, "uuid_bulk", "count", int64(500), "version", "7").([]string)
	for i := 1; i < len(bulk); i++ {
		if bulk[i] <= bulk[i-1] {
			t.Fatalf("v7 bulk not increasing at %d: %s <= %s", i, bulk[i], bulk[i-1])
		}
	}
	if _, err := callE("uuid_bulk", "count", int64(10001)); err == nil {
		t.Error("count cap")
	}
	if _, err := callE("uuid_bulk", "version", "9"); err == nil {
		t.Error("bad version")
	}
}

func TestParseFormat(t *testing.T) {
	const canon = "3b241101-e2bb-4255-8caf-4136c566a962"
	for _, in := range []string{canon, strings.ToUpper(canon), "{" + canon + "}", "urn:uuid:" + canon, strings.ReplaceAll(canon, "-", ""), "  " + canon + "\n"} {
		if got := call(t, "uuid_format", "u", in); got != canon {
			t.Errorf("%q -> %v", in, got)
		}
		if call(t, "uuid_validate", "u", in) != true {
			t.Errorf("%q should validate", in)
		}
	}
	for _, bad := range []string{"", "xyz", "3b241101-e2bb-4255-8caf-4136c566a96", "3b241101e2bb-4255-8caf-4136c566a9622", "3b241101-e2bb-4255-8caf-4136c566a96g"} {
		if call(t, "uuid_validate", "u", bad) != false {
			t.Errorf("%q should not validate", bad)
		}
	}
	if call(t, "uuid_validate", "u", canon, "version", int64(4)) != true || call(t, "uuid_validate", "u", canon, "version", int64(7)) != false {
		t.Error("version check")
	}
	styles := map[string]string{
		"upper": "3B241101-E2BB-4255-8CAF-4136C566A962", "no_dashes": "3b241101e2bb42558caf4136c566a962",
		"braces": "{" + canon + "}", "urn": "urn:uuid:" + canon, "base64": "OyQRAeK7QlWMr0E2xWapYg",
		"base32": "HMSBCAPCXNBFLDFPIE3MKZVJMI",
	}
	for st, want := range styles {
		if got := call(t, "uuid_format", "u", canon, "style", st); got != want {
			t.Errorf("style %s: %v want %s", st, got, want)
		}
	}
	short := call(t, "uuid_format", "u", canon, "style", "base58").(string)
	if len(short) != 22 || call(t, "short_uuid_decode", "s", short) != canon || call(t, "short_uuid", "u", canon) != short {
		t.Errorf("base58 %s", short)
	}
	if call(t, "short_uuid_decode", "s", "1111111111111111111111") != "00000000-0000-0000-0000-000000000000" {
		t.Error("base58 zero")
	}
	if _, err := callE("uuid_format", "u", canon, "style", "nope"); err == nil {
		t.Error("bad style")
	}
	n := call(t, "uuid_to_int", "u", canon).(string)
	if call(t, "uuid_from_int", "value", n) != canon {
		t.Error("int roundtrip")
	}
	if call(t, "uuid_to_int", "u", "00000000-0000-0000-0000-0000000000ff") != "255" {
		t.Error("to_int 255")
	}
	if call(t, "uuid_from_int", "value", "0xff") != "00000000-0000-0000-0000-0000000000ff" {
		t.Error("hex int")
	}
	if _, err := callE("uuid_from_int", "value", "340282366920938463463374607431768211456"); err == nil {
		t.Error("2^128 should overflow")
	}
	if _, err := callE("uuid_from_int", "value", "-1"); err == nil {
		t.Error("negative")
	}
	nilm := call(t, "uuid_parse", "u", "00000000-0000-0000-0000-000000000000").(map[string]string)
	if nilm["version"] != "nil" {
		t.Error(nilm)
	}
}

func TestCompareSort(t *testing.T) {
	a := "017f22e2-79b0-7cc3-98c4-dc0c0c07398f"
	b := "017f22e2-79b1-7000-8000-000000000000"
	if call(t, "uuid_compare", "a", a, "b", b) != int64(-1) || call(t, "uuid_compare", "a", b, "b", a) != int64(1) || call(t, "uuid_compare", "a", a, "b", a) != int64(0) {
		t.Error("compare")
	}
	got := call(t, "uuid_sort", "uuids", []string{b, a}).([]string)
	if got[0] != a || got[1] != b {
		t.Errorf("%v", got)
	}
	got = call(t, "uuid_sort", "uuids", []string{a, b}, "reverse", true, "by", "time").([]string)
	if got[0] != b {
		t.Errorf("%v", got)
	}
	// v1 sorts by time, not bytes
	older := call(t, "uuid_v1", "time", "2020-01-01T00:00:00Z", "seed", int64(1)).(string)
	newer := call(t, "uuid_v1", "time", "2021-01-01T00:00:00Z", "seed", int64(1)).(string)
	if call(t, "uuid_compare", "a", older, "b", newer, "by", "time") != int64(-1) {
		t.Error("v1 time order")
	}
	if _, err := callE("uuid_sort", "uuids", []string{"nope"}); err == nil {
		t.Error("invalid uuid in sort")
	}
}

func TestULID(t *testing.T) {
	const spec = "01ARYZ6S41TSV4RRFFQ69G5FAV"
	m := call(t, "ulid_parse", "ulid", spec).(map[string]string)
	if m["unix_ms"] != "1469918176385" || m["timestamp"] != "2016-07-30T22:36:16.385Z" {
		t.Errorf("%v", m)
	}
	if call(t, "ulid_parse", "ulid", strings.ToLower(spec)).(map[string]string)["canonical"] != spec {
		t.Error("lowercase")
	}
	// Crockford aliases
	if call(t, "ulid_parse", "ulid", "O1ARYZ6S41TSV4RRFFQ69G5FAV").(map[string]string)["canonical"] != "01ARYZ6S41TSV4RRFFQ69G5FAV" {
		t.Error("alias O")
	}
	for _, bad := range []string{"", "01ARYZ6S41TSV4RRFFQ69G5FA", "81ARYZ6S41TSV4RRFFQ69G5FAV", "01ARYZ6S41TSV4RRFFQ69G5FAU"} {
		if _, err := callE("ulid_parse", "ulid", bad); err == nil {
			t.Errorf("%q should fail", bad)
		}
	}
	u := call(t, "ulid_to_uuid", "ulid", spec).(string)
	if u != "01563df3-6481-d676-4c61-efb99302bd5b" {
		t.Errorf("ulid->uuid %s", u)
	}
	if call(t, "uuid_to_ulid", "u", u) != spec {
		t.Error("roundtrip")
	}
	id := call(t, "ulid_new", "time", "2016-07-30T22:36:16.385Z", "seed", int64(1)).(string)
	if !strings.HasPrefix(id, spec[:10]) || len(id) != 26 {
		t.Errorf("ulid_new %s", id)
	}
	if call(t, "ulid_new", "time", "1469918176385", "seed", int64(1)) != id {
		t.Error("ms time input")
	}
	bulk := call(t, "ulid_bulk", "count", int64(300)).([]string)
	for i := 1; i < len(bulk); i++ {
		if bulk[i] <= bulk[i-1] {
			t.Fatalf("ulid bulk not monotonic at %d", i)
		}
	}
	if _, err := callE("ulid_new", "time", "1960-01-01"); err == nil {
		t.Error("pre-1970 ULID")
	}
}

func TestKSUID(t *testing.T) {
	m := call(t, "ksuid_parse", "ksuid", "0ujtsYcgvSTl8PAuAdqWYSMnLOv").(map[string]string)
	if m["ksuid_seconds"] != "107608047" || m["payload"] != "B5A1CD34B5F99D1154FB6853345C9735" || m["timestamp"] != "2017-10-10T04:00:47.000Z" {
		t.Errorf("%v", m)
	}
	id := call(t, "ksuid_new", "time", "2017-10-10T04:00:47Z", "seed", int64(3)).(string)
	if len(id) != 27 || !strings.HasPrefix(id, "0ujts") {
		t.Errorf("ksuid_new %s", id)
	}
	if call(t, "ksuid_parse", "ksuid", id).(map[string]string)["unix"] != "1507608047" {
		t.Error("roundtrip")
	}
	if _, err := callE("ksuid_new", "time", "2000-01-01"); err == nil {
		t.Error("before epoch")
	}
	if _, err := callE("ksuid_parse", "ksuid", "short"); err == nil {
		t.Error("short")
	}
	if _, err := callE("ksuid_parse", "ksuid", "zzzzzzzzzzzzzzzzzzzzzzzzzzz"); err == nil {
		t.Error("overflow")
	}
}

func TestSnowflake(t *testing.T) {
	m := call(t, "snowflake_decode", "id", "175928847299117063", "epoch", "discord").(map[string]string)
	if m["timestamp"] != "2016-04-30T11:18:25.796Z" || m["worker"] != "1" || m["process"] != "0" || m["sequence"] != "7" {
		t.Errorf("%v", m)
	}
	enc := call(t, "snowflake_encode", "time", "2016-04-30T11:18:25.796Z", "worker", int64(1), "sequence", int64(7), "epoch", "discord")
	if enc != "175928847299117063" {
		t.Errorf("encode %v", enc)
	}
	// Tweet 1382350606417817604 (Twitter epoch)
	tm := call(t, "snowflake_decode", "id", "1382350606417817604").(map[string]string)
	if !strings.HasPrefix(tm["timestamp"], "2021-04") {
		t.Errorf("%v", tm)
	}
	custom := call(t, "snowflake_encode", "time", "1500000000", "epoch", "0").(string)
	if call(t, "snowflake_decode", "id", custom, "epoch", "0").(map[string]string)["unix_ms"] != "1500000000000" {
		t.Errorf("custom epoch (%s)", custom)
	}
	if _, err := callE("snowflake_encode", "time", "2000-01-01", "epoch", "discord"); err == nil {
		t.Error("before epoch")
	}
	if _, err := callE("snowflake_encode", "time", "2020-01-01", "worker", int64(32)); err == nil {
		t.Error("worker range")
	}
	if _, err := callE("snowflake_decode", "id", "abc"); err == nil {
		t.Error("bad id")
	}
	if _, err := callE("snowflake_decode", "id", "1", "epoch", "nope"); err == nil {
		t.Error("bad epoch")
	}
}

func TestObjectID(t *testing.T) {
	m := call(t, "objectid_parse", "id", "507f1f77bcf86cd799439011").(map[string]string)
	if m["timestamp"] != "2012-10-17T21:13:27.000Z" || m["counter"] != strconv.Itoa(0x439011) || m["random"] != "bcf86cd799" {
		t.Errorf("%v", m)
	}
	id := call(t, "objectid_new", "time", "2012-10-17T21:13:27Z", "seed", int64(2)).(string)
	if len(id) != 24 || !strings.HasPrefix(id, "507f1f77") {
		t.Errorf("objectid_new %s", id)
	}
	if _, err := callE("objectid_parse", "id", "507f1f77bcf86cd79943901"); err == nil {
		t.Error("short")
	}
	if _, err := callE("objectid_parse", "id", "zzzzzzzzzzzzzzzzzzzzzzzz"); err == nil {
		t.Error("non hex")
	}
}

func TestNanoidAndTypeID(t *testing.T) {
	id := call(t, "nanoid").(string)
	if !nanoidRe.MatchString(id) {
		t.Errorf("%s", id)
	}
	if s := call(t, "nanoid", "size", int64(30), "alphabet", "ab", "seed", int64(1)).(string); len(s) != 30 || strings.Trim(s, "ab") != "" {
		t.Error(s)
	}
	if call(t, "nanoid", "size", int64(10), "seed", int64(4)) != call(t, "nanoid", "size", int64(10), "seed", int64(4)) {
		t.Error("seed")
	}
	if _, err := callE("nanoid", "alphabet", "a"); err == nil {
		t.Error("single char alphabet")
	}
	if _, err := callE("nanoid", "size", int64(0)); err == nil {
		t.Error("size 0")
	}
	if c := call(t, "cuid2").(string); len(c) != 24 || c[0] < 'a' || c[0] > 'z' {
		t.Error(c)
	}
	if call(t, "uuid_to_typeid", "u", "00000000-0000-0000-0000-000000000000", "prefix", "") != "00000000000000000000000000" {
		t.Error("typeid nil")
	}
	if call(t, "uuid_to_typeid", "u", "ffffffff-ffff-ffff-ffff-ffffffffffff", "prefix", "x") != "x_7zzzzzzzzzzzzzzzzzzzzzzzzz" {
		t.Error("typeid max")
	}
	tid := call(t, "typeid_new", "prefix", "user", "time", "2022-02-22T19:22:22Z", "seed", int64(1)).(string)
	m := call(t, "typeid_parse", "typeid", tid).(map[string]string)
	if m["prefix"] != "user" || m["version"] != "7" || !strings.HasPrefix(m["timestamp"], "2022-02-22T19:22:22") {
		t.Errorf("%s %v", tid, m)
	}
	for _, bad := range []string{"User_00000000000000000000000000", "user_8zzzzzzzzzzzzzzzzzzzzzzzzz", "user_0000", "_00000000000000000000000000"} {
		if _, err := callE("typeid_parse", "typeid", bad); err == nil {
			t.Errorf("%q should fail", bad)
		}
	}
	if _, err := callE("typeid_new", "prefix", "Bad_"); err == nil {
		t.Error("bad prefix")
	}
}

func TestDetect(t *testing.T) {
	has := func(s, want string) {
		t.Helper()
		for _, g := range call(t, "id_detect", "s", s).([]string) {
			if strings.Contains(g, want) {
				return
			}
		}
		t.Errorf("id_detect(%q) lacks %q: %v", s, want, call(t, "id_detect", "s", s))
	}
	has("3b241101-e2bb-4255-8caf-4136c566a962", "uuid v4")
	has("017f22e2-79b0-7cc3-98c4-dc0c0c07398f", "uuid v7 (2022-02-22")
	has("01ARZ3NDEKTSV4RRFFQ69G5FAV", "ulid")
	has("0ujtsYcgvSTl8PAuAdqWYSMnLOv", "ksuid")
	has("507f1f77bcf86cd799439011", "objectid")
	has("175928847299117063", "snowflake, discord")
	has("user_01h455vb4pex5vsknk084sn02q", "typeid")
	has("V1StGXR8_Z5jdHi6B-myT", "nanoid")
	has("3b241101e2bb42558caf4136c566a962", "without dashes")
	if got := call(t, "id_detect", "s", "hello world").([]string); len(got) != 0 {
		t.Errorf("%v", got)
	}
	if got := call(t, "id_detect", "s", "").([]string); len(got) != 0 {
		t.Errorf("%v", got)
	}
}

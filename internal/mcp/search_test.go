package mcp

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/avestura/zagros/stubs"
)

const compressStub = `"""compress: zip, gzip and tar archives.

Type description, not runnable Starlark: do not ` + "`import`" + ` anything.
Needs: fs.read / fs.write.
Keywords: archive, zip, tar, gzip
"""

def extract(archive: str, dest: str) -> int:
    """Unpack an archive into a folder.

    Example:
        n = compress.extract("a.zip", "out")
    Task: unzip a file
    Keywords: unzip, decompress, extract zip, unpack archive, untar
    Needs: fs.read, fs.write
    """

def create(paths: List[str], archive: str) -> int:
    """Pack files into a zip archive.

    Example: compress.create(["a.txt"], "a.zip")
    Keywords: zip files, compress, make archive, pack folder
    Needs: fs.read, fs.write
    """

def lines(text: str) -> List[str]:
    """Split text into lines and extract the non-empty ones.

    Example: compress.lines("a\nb")
    Keywords: split text
    Needs: none
    """
`

func compressIndex(t *testing.T) *Index {
	t.Helper()
	idx, err := BuildIndex(fstest.MapFS{"zagros/compress.pyi": {Data: []byte(compressStub)}})
	if err != nil {
		t.Fatal(err)
	}
	return idx
}

func TestKeywordsTaskAndExampleBlocksAreParsed(t *testing.T) {
	idx := compressIndex(t)
	e, ok := idx.Get("compress.extract")
	if !ok {
		t.Fatal("compress.extract missing")
	}
	if e.Task != "unzip a file" || len(e.Keywords) != 5 || e.Keywords[0] != "unzip" || e.Keywords[2] != "extract zip" {
		t.Fatalf("task/keywords: %q %q", e.Task, e.Keywords)
	}
	if e.Example != `n = compress.extract("a.zip", "out")` || e.Needs != "fs.read, fs.write" {
		t.Fatalf("example/needs: %q %q", e.Example, e.Needs)
	}
	c, _ := idx.Get("compress.create")
	if c.Example != `compress.create(["a.txt"], "a.zip")` {
		t.Fatalf("one-line example: %q", c.Example)
	}
}

func TestExampleFencedBlock(t *testing.T) {
	doc := "Does a thing.\n\nExample:\n    ```\n    a = 1\n    b = a + 1\n    ```\nNeeds: none"
	if got := parseExample(doc); got != "a = 1\nb = a + 1" {
		t.Fatalf("fenced: %q", got)
	}
	doc = "Does a thing.\n\nExample:\n    x = 1\n\n    y = 2\nKeywords: a, b\nNeeds: none"
	if got := parseExample(doc); got != "x = 1\n\ny = 2" {
		t.Fatalf("indented: %q", got)
	}
}

func rank(m []Match, q string) int {
	for i, x := range m {
		if x.Qualified == q {
			return i
		}
	}
	return -1
}

func TestSearchUsesKeywordsAndSynonyms(t *testing.T) {
	idx := compressIndex(t)
	for _, q := range []string{"unzip a file", "decompress", "uncompress archive", "unpack", "extract zip", "how do I unzip files", "untar"} {
		if r := rank(idx.Search(q, "", 5), "compress.extract"); r != 0 {
			t.Errorf("%q: compress.extract ranked %d: %+v", q, r, idx.Search(q, "", 5))
		}
	}
	for _, q := range []string{"zip files", "compress a folder", "make an archive", "deflate"} {
		if r := rank(idx.Search(q, "", 5), "compress.create"); r != 0 {
			t.Errorf("%q: compress.create ranked %d: %+v", q, r, idx.Search(q, "", 5))
		}
	}
}

func TestSynonymsAreSane(t *testing.T) {
	has := func(w, s string) bool {
		for _, x := range Synonyms(w) {
			if x == s {
				return true
			}
		}
		return false
	}
	if !has("unzip", "decompress") || !has("decompress", "unzip") || !has("zip", "compress") {
		t.Fatal("archive words should be synonyms")
	}
	for _, bad := range [][2]string{{"encode", "encrypt"}, {"encrypt", "decrypt"}, {"encode", "decode"}, {"hash", "encrypt"}, {"zip", "unzip"}, {"compress", "decompress"}} {
		if has(bad[0], bad[1]) || has(bad[1], bad[0]) {
			t.Errorf("%s and %s must not be synonyms", bad[0], bad[1])
		}
	}
	for _, g := range synonymGroups {
		for _, w := range g {
			if w != strings.ToLower(w) || strings.Contains(w, " ") {
				t.Errorf("synonym %q must be one lower-case word", w)
			}
		}
	}
}

func TestSearchRealStubsByTask(t *testing.T) {
	idx, err := BuildIndex(stubs.FS())
	if err != nil {
		t.Fatal(err)
	}
	cases := map[string]string{
		"turn text into base64":             "encode.base64_encode",
		"decode base64":                     "encode.base64_decode",
		"make a uuid":                       "uuid.uuid_v4",
		"generate a guid":                   "uuid.uuid_v4",
		"checksum of text":                  "hash.",
		"how many days until a date":        "dt.days_until",
		"remove duplicate lines":            "lines.unique_lines",
		"read the clipboard":                "clipboard.get",
		"paste":                             "clipboard.get",
		"remember a value between runs":     "storage.get",
		"fetch a web page":                  "http.get",
		"unix timestamp to date":            "dt.from_unix",
		"escape text for html":              "encode.html_escape",
		"percent-encode text for a url":     "encode.url_encode",
		"open a sub list when pressing tab": "scope",
	}
	for q, want := range cases {
		m := idx.Search(q, "", 5)
		found := false
		for i, x := range m {
			if i < 3 && strings.HasPrefix(x.Qualified, want) {
				found = true
			}
		}
		if !found {
			names := []string{}
			for _, x := range m {
				names = append(names, x.Qualified)
			}
			t.Errorf("%q: want %s in the top 3, got %v", q, want, names)
		}
	}
	// encode is not encrypt: a query for one must not lead with the other.
	if m := idx.Search("encrypt text", "", 3); len(m) == 0 || !strings.HasPrefix(m[0].Qualified, "cipher.") {
		t.Errorf("encrypt text: %+v", m)
	}
}

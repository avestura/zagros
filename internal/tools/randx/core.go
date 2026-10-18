package randx

import (
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

const (
	maxCount  = 10000
	maxLength = 100000
)

// cryptoSrc is a math/rand/v2 Source backed by crypto/rand.
type cryptoSrc struct{}

func (cryptoSrc) Uint64() uint64 {
	var b [8]byte
	_, _ = crand.Read(b[:])
	return binary.LittleEndian.Uint64(b[:])
}

// newRng returns a crypto-random generator for seed 0 and a deterministic PCG
// generator for any other seed.
func newRng(seed int64) *rand.Rand {
	if seed == 0 {
		return rand.New(cryptoSrc{})
	}
	return rand.New(rand.NewPCG(uint64(seed), 0x9e3779b97f4a7c15))
}

// cryptoRng always uses crypto/rand (passwords, tokens).
func cryptoRng() *rand.Rand { return rand.New(cryptoSrc{}) }

func p(name, typ, doc string, def any) tools.Param {
	return tools.Param{Name: name, Type: typ, Doc: doc, Default: def}
}

func req(name, typ, doc string) tools.Param {
	return tools.Param{Name: name, Type: typ, Doc: doc, Required: true}
}

func seedParam() tools.Param {
	return p("seed", "int", "0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.", int64(0))
}

func argsParam(doc string) tools.Param {
	return p("args", "str", "Launcher shortcut, free text: "+doc+" When non-empty it overrides the matching parameters.", "")
}

// Every variant answers rand/random; specific keywords (roll, flip, shuffle...) are on the functions themselves.
var defaultKeywords = []string{"rand", "random"}

// reg registers f in group rand. Functions with an explicit Input are also
// shown in the launcher family "rand".
func reg(f tools.Func) {
	f.Group = "rand"
	if f.Input != "" {
		f.Family = "rand"
		f.Keywords = append(append([]string(nil), defaultKeywords...), f.Keywords...)
	}
	tools.Register(f)
}

// numArgs extracts the integers from free text.
func numArgs(s string) []int64 {
	var out []int64
	for _, f := range strings.FieldsFunc(s, func(r rune) bool { return r == ' ' || r == ',' || r == ';' }) {
		if n, err := strconv.ParseInt(f, 10, 64); err == nil {
			out = append(out, n)
		}
	}
	return out
}

func checkCount(n int64, what string) error {
	if n < 0 {
		return fmt.Errorf("%s must not be negative", what)
	}
	if n > maxCount {
		return fmt.Errorf("%s must be at most %d", what, maxCount)
	}
	return nil
}

func checkLength(n int64, what string) error {
	if n < 0 {
		return fmt.Errorf("%s must not be negative", what)
	}
	if n > maxLength {
		return fmt.Errorf("%s must be at most %d", what, maxLength)
	}
	return nil
}

var errEmpty = errors.New("the list is empty")

// intRange draws a uniform integer in [lo, hi].
func intRange(r *rand.Rand, lo, hi int64) int64 {
	span := uint64(hi) - uint64(lo)
	if span == ^uint64(0) {
		return int64(r.Uint64())
	}
	return int64(uint64(lo) + r.Uint64N(span+1))
}

// distinctInts draws count distinct integers in [lo, hi].
func distinctInts(r *rand.Rand, count, lo, hi int64) ([]int64, error) {
	span := uint64(hi) - uint64(lo)
	if span != ^uint64(0) && uint64(count) > span+1 {
		return nil, fmt.Errorf("cannot draw %d unique values from [%d, %d]", count, lo, hi)
	}
	seen := make(map[int64]bool, count)
	out := make([]int64, 0, count)
	for int64(len(out)) < count {
		v := intRange(r, lo, hi)
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out, nil
}

func shuffled(r *rand.Rand, items []string) []string {
	out := append([]string(nil), items...)
	r.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

func toFloats(v []int64) []float64 {
	out := make([]float64, len(v))
	for i, x := range v {
		out[i] = float64(x)
	}
	return out
}

// splitItems splits free text into items: by sep when given, else by lines,
// else commas/semicolons, else whitespace.
func splitItems(text, sep string) []string {
	var parts []string
	switch {
	case sep != "":
		parts = strings.Split(text, sep)
	case strings.Contains(text, "\n"):
		parts = strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	case strings.ContainsAny(text, ",;"):
		parts = strings.FieldsFunc(text, func(r rune) bool { return r == ',' || r == ';' })
	default:
		parts = strings.Fields(text)
	}
	var out []string
	for _, s := range parts {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	return out
}

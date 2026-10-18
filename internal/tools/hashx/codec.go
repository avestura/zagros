package hashx

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

const group = "hash"

// Binary data crosses the API as strings. decodeBytes reads one using an
// encoding name: "utf-8" (default; the text's own bytes), "hex" or "base64".
func decodeBytes(s, enc string) ([]byte, error) {
	switch strings.ToLower(strings.TrimSpace(enc)) {
	case "", "utf-8", "utf8", "text", "raw":
		return []byte(s), nil
	case "hex":
		t := strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "", ":", "", "-", "").Replace(s)
		t = strings.TrimPrefix(strings.TrimPrefix(t, "0x"), "0X")
		b, err := hex.DecodeString(t)
		if err != nil {
			return nil, fmt.Errorf("invalid hex input: %v", err)
		}
		return b, nil
	case "base64", "b64":
		return decodeB64(s)
	}
	return nil, fmt.Errorf("unknown encoding %q (use utf-8, hex or base64)", enc)
}

func decodeB64(s string) ([]byte, error) {
	t := strings.NewReplacer(" ", "", "\t", "", "\r", "", "\n", "").Replace(s)
	t = strings.TrimRight(t, "=")
	if b, err := base64.RawStdEncoding.DecodeString(t); err == nil {
		return b, nil
	}
	if b, err := base64.RawURLEncoding.DecodeString(t); err == nil {
		return b, nil
	}
	return nil, fmt.Errorf("invalid base64 input")
}

// encodeBytes renders digest bytes: hex (default), HEX, base64, base64url,
// base32 or dec (big-endian unsigned decimal).
func encodeBytes(b []byte, out string) (string, error) {
	switch strings.TrimSpace(out) {
	case "", "hex", "lower":
		return hex.EncodeToString(b), nil
	case "HEX", "upper":
		return strings.ToUpper(hex.EncodeToString(b)), nil
	case "base64", "b64":
		return base64.StdEncoding.EncodeToString(b), nil
	case "base64url":
		return base64.URLEncoding.EncodeToString(b), nil
	case "base32":
		return base32.StdEncoding.EncodeToString(b), nil
	case "dec", "decimal":
		return new(big.Int).SetBytes(b).String(), nil
	}
	return "", fmt.Errorf("unknown output %q (use hex, HEX, base64, base64url, base32 or dec)", out)
}

func textParam() tools.Param {
	return tools.Param{Name: "text", Type: "str", Required: true, Doc: "Input text (UTF-8 bytes unless input_encoding says otherwise)."}
}

func inEncParam() tools.Param {
	return tools.Param{Name: "input_encoding", Type: "str", Default: "utf-8", Choices: []string{"utf-8", "hex", "base64"},
		Doc: "How text is turned into bytes: utf-8 (default), hex or base64."}
}

func outParam() tools.Param {
	return tools.Param{Name: "output", Type: "str", Default: "hex", Choices: []string{"hex", "HEX", "base64", "base64url", "base32", "dec"},
		Doc: "Output format: hex (default, lowercase), HEX (uppercase), base64, base64url, base32 or dec (big-endian decimal number)."}
}

func encParam(name, what string) tools.Param {
	return tools.Param{Name: name, Type: "str", Default: "utf-8", Choices: []string{"utf-8", "hex", "base64"},
		Doc: "How " + what + " is turned into bytes: utf-8 (default), hex or base64."}
}

func clampInt(v, lo, hi int64, name string) error {
	if v < lo || v > hi {
		return fmt.Errorf("%s must be between %d and %d (got %d)", name, lo, hi, v)
	}
	return nil
}

func normAlgo(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.NewReplacer("-", "_", "/", "_", " ", "_").Replace(s)
	return s
}

// inBytes reads the standard text + input_encoding pair.
func inBytes(a tools.Args) ([]byte, error) {
	return decodeBytes(a.Str("text"), a.Str("input_encoding"))
}

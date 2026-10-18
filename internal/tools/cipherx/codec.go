package cipherx

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

const group = "cipher"

func reg(f tools.Func) {
	f.Group = group
	tools.Register(f)
}

// Binary data crosses the API as strings. decodeBytes reads one using an
// encoding name: utf-8 (the text's own bytes), hex or base64.
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
	case "base64url":
		return decodeB64(s)
	case "base32":
		t := strings.ToUpper(strings.NewReplacer(" ", "", "\n", "", "\r", "", "\t", "").Replace(s))
		b, err := base32.StdEncoding.DecodeString(t)
		if err != nil {
			b, err = base32.StdEncoding.WithPadding(base32.NoPadding).DecodeString(strings.TrimRight(t, "="))
		}
		if err != nil {
			return nil, fmt.Errorf("invalid base32 input")
		}
		return b, nil
	}
	return nil, fmt.Errorf("unknown encoding %q (use utf-8, hex, base64, base64url or base32)", enc)
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

// encodeBytes renders bytes: hex, HEX, base64, base64url, base32, or utf-8
// (the bytes as text).
func encodeBytes(b []byte, out string) (string, error) {
	switch strings.TrimSpace(out) {
	case "hex", "lower":
		return hex.EncodeToString(b), nil
	case "HEX", "upper":
		return strings.ToUpper(hex.EncodeToString(b)), nil
	case "base64", "b64":
		return base64.StdEncoding.EncodeToString(b), nil
	case "base64url":
		return base64.URLEncoding.EncodeToString(b), nil
	case "base32":
		return base32.StdEncoding.EncodeToString(b), nil
	case "utf-8", "utf8", "text", "raw":
		return string(b), nil
	}
	return "", fmt.Errorf("unknown output %q (use hex, HEX, base64, base64url, base32 or utf-8)", out)
}

func clampInt(v, lo, hi int64, name string) error {
	if v < lo || v > hi {
		return fmt.Errorf("%s must be between %d and %d (got %d)", name, lo, hi, v)
	}
	return nil
}

func normAlgo(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	return strings.NewReplacer("-", "_", "/", "_", " ", "_").Replace(s)
}

func strParam(name, doc string, def string) tools.Param {
	return tools.Param{Name: name, Type: "str", Default: def, Doc: doc}
}

func reqStr(name, doc string) tools.Param {
	return tools.Param{Name: name, Type: "str", Required: true, Doc: doc}
}

func outParam(def string, doc string) tools.Param {
	return tools.Param{Name: "output", Type: "str", Default: def, Choices: []string{"base64", "base64url", "hex", "HEX", "base32", "utf-8"}, Doc: doc}
}

func inEncParam() tools.Param {
	return tools.Param{Name: "input_encoding", Type: "str", Default: "utf-8", Choices: []string{"utf-8", "hex", "base64"},
		Doc: "How text is turned into bytes: utf-8 (default), hex or base64."}
}

func encParam(name, what string) tools.Param {
	return tools.Param{Name: name, Type: "str", Default: "utf-8", Choices: []string{"utf-8", "hex", "base64"},
		Doc: "How " + what + " is turned into bytes: utf-8 (default), hex or base64."}
}

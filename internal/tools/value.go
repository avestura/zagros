package tools

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// Dict is an insertion-ordered string-keyed map, the structured value that
// parsers return for the "any" type (it becomes a Starlark dict that keeps its
// key order). Lists are []any, scalars are string, int64, float64, bool or nil.
type Dict struct {
	keys []string
	vals map[string]any
}

// NewDict returns an empty Dict.
func NewDict() *Dict { return &Dict{vals: map[string]any{}} }

// Set adds or replaces key; a new key goes last.
func (d *Dict) Set(k string, v any) {
	if _, ok := d.vals[k]; !ok {
		d.keys = append(d.keys, k)
	}
	d.vals[k] = v
}

// Get returns the value of key.
func (d *Dict) Get(k string) (any, bool) { v, ok := d.vals[k]; return v, ok }

// Keys returns the keys in insertion order.
func (d *Dict) Keys() []string { return append([]string(nil), d.keys...) }

// Len is the number of keys.
func (d *Dict) Len() int { return len(d.keys) }

// MarshalJSON keeps the key order.
func (d *Dict) MarshalJSON() ([]byte, error) {
	var b bytes.Buffer
	b.WriteByte('{')
	for i, k := range d.keys {
		if i > 0 {
			b.WriteByte(',')
		}
		kb, _ := marshalRaw(k)
		b.Write(kb)
		b.WriteByte(':')
		vb, err := marshalRaw(d.vals[k])
		if err != nil {
			return nil, err
		}
		b.Write(vb)
	}
	b.WriteByte('}')
	return b.Bytes(), nil
}

func marshalRaw(v any) ([]byte, error) {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return nil, err
	}
	return bytes.TrimRight(b.Bytes(), "\n"), nil
}

// FromMap builds a Dict from a plain map with sorted keys.
func FromMap(m map[string]any) *Dict {
	d := NewDict()
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		d.Set(k, m[k])
	}
	return d
}

// Plain converts structured values to plain Go maps and slices (for
// encoders that want map[string]any), recursively.
func Plain(v any) any {
	switch x := v.(type) {
	case *Dict:
		m := make(map[string]any, len(x.keys))
		for _, k := range x.keys {
			m[k] = Plain(x.vals[k])
		}
		return m
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = Plain(e)
		}
		return out
	case []string:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = e
		}
		return out
	case map[string]string:
		m := make(map[string]any, len(x))
		for k, e := range x {
			m[k] = e
		}
		return m
	case map[string]any:
		m := make(map[string]any, len(x))
		for k, e := range x {
			m[k] = Plain(e)
		}
		return m
	}
	return v
}

// JSON renders a structured value as indented JSON, for Format.
func JSON(v any) string {
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		return err.Error()
	}
	return strings.TrimRight(b.String(), "\n")
}

// Req declares a required parameter.
func Req(name, typ, doc string) Param { return Param{Name: name, Type: typ, Doc: doc, Required: true} }

// Opt declares an optional parameter with a default (nil means the zero value).
func Opt(name, typ, doc string, def any) Param {
	return Param{Name: name, Type: typ, Doc: doc, Default: def}
}

// Scalar renders a scalar structured value as text for formats that only
// hold strings (CSV cells, query values). Lists and dicts return ok=false.
func Scalar(v any) (string, bool) {
	switch x := v.(type) {
	case nil:
		return "", true
	case string:
		return x, true
	case bool:
		if x {
			return "true", true
		}
		return "false", true
	case int64:
		return Format(x), true
	case int:
		return Format(x), true
	case float64:
		return Format(x), true
	}
	return "", false
}

// ParseJSON decodes JSON text into structured values: objects keep their key
// order (*Dict), arrays are []any, whole numbers are int64, other numbers
// float64. Nesting deeper than 200 levels is an error.
func ParseJSON(s string) (any, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	v, err := decodeValue(dec, 0)
	if err != nil {
		return nil, err
	}
	if dec.More() {
		return nil, fmt.Errorf("unexpected data after the JSON value")
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("unexpected data after the JSON value")
	}
	return v, nil
}

func decodeValue(dec *json.Decoder, depth int) (any, error) {
	if depth > 200 {
		return nil, fmt.Errorf("JSON is nested deeper than 200 levels")
	}
	tok, err := dec.Token()
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}
	switch x := tok.(type) {
	case json.Delim:
		if x == '[' {
			out := []any{}
			for dec.More() {
				v, err := decodeValue(dec, depth+1)
				if err != nil {
					return nil, err
				}
				out = append(out, v)
			}
			_, err := dec.Token()
			return out, err
		}
		d := NewDict()
		for dec.More() {
			kt, err := dec.Token()
			if err != nil {
				return nil, err
			}
			v, err := decodeValue(dec, depth+1)
			if err != nil {
				return nil, err
			}
			d.Set(kt.(string), v)
		}
		_, err := dec.Token()
		return d, err
	case json.Number:
		if i, err := x.Int64(); err == nil {
			return i, nil
		}
		f, err := x.Float64()
		return f, err
	}
	return tok, nil
}

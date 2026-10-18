package ext

import (
	"fmt"
	"math"
	"sort"

	"github.com/avestura/zagros/internal/tools"
	"go.starlark.net/starlark"
)

// toolModules exposes the pure utility registry (internal/tools) as Starlark
// modules: text, lines, num, math, encode, hash, cipher, rand, uuid, regex,
// net, dt, extract. They need no permission because the functions are pure
// (no file, network, clipboard or environment access). The modules are built
// from the registry, so a new tools function reaches extensions, the stubs and
// the docs without further wiring.
func toolModules() starlark.StringDict {
	groups := map[string]starlark.StringDict{}
	for _, f := range tools.All() {
		m := groups[f.Group]
		if m == nil {
			m = starlark.StringDict{}
			groups[f.Group] = m
		}
		m[f.Name] = toolBuiltin(f)
	}
	out := starlark.StringDict{}
	for g, members := range groups {
		out[g] = module(g, members)
	}
	return out
}

func toolBuiltin(f tools.Func) *starlark.Builtin {
	return starlark.NewBuiltin(f.Key(), func(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
		vals := make([]starlark.Value, len(f.Params))
		pairs := make([]any, 0, 2*len(f.Params))
		for i, p := range f.Params {
			name := p.Name
			if !p.Required {
				name += "?"
			}
			pairs = append(pairs, name, &vals[i])
		}
		if err := starlark.UnpackArgs(f.Key(), args, kw, pairs...); err != nil {
			return nil, err
		}
		named := map[string]any{}
		for i, p := range f.Params {
			if vals[i] == nil || vals[i] == starlark.None {
				continue // omitted: the registry applies the default
			}
			v, err := toGo(p.Type, vals[i])
			if err != nil {
				return nil, fmt.Errorf("%s: argument %q: %w", f.Key(), p.Name, err)
			}
			named[p.Name] = v
		}
		res, err := f.Invoke(named)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", f.Key(), err)
		}
		return fromGo(res)
	})
}

// toGo converts a Starlark argument to the Go type of a tools param type.
func toGo(typ string, v starlark.Value) (any, error) {
	switch typ {
	case "str":
		switch x := v.(type) {
		case starlark.String:
			return string(x), nil
		case starlark.Bytes:
			return string(x), nil
		}
		return nil, fmt.Errorf("want str, got %s", v.Type())
	case "int":
		if i, ok := v.(starlark.Int); ok {
			n, ok := i.Int64()
			if !ok {
				return nil, fmt.Errorf("integer too large for int64; pass big numbers as strings")
			}
			return n, nil
		}
		return nil, fmt.Errorf("want int, got %s", v.Type())
	case "float":
		switch x := v.(type) {
		case starlark.Float:
			return float64(x), nil
		case starlark.Int:
			return float64(x.Float()), nil
		}
		return nil, fmt.Errorf("want float, got %s", v.Type())
	case "bool":
		if b, ok := v.(starlark.Bool); ok {
			return bool(b), nil
		}
		return nil, fmt.Errorf("want bool, got %s", v.Type())
	case "list[str]", "list[float]":
		it, ok := v.(starlark.Iterable)
		if !ok {
			return nil, fmt.Errorf("want a list, got %s", v.Type())
		}
		iter := it.Iterate()
		defer iter.Done()
		var strs []string
		var floats []float64
		var x starlark.Value
		for iter.Next(&x) {
			if typ == "list[str]" {
				s, ok := x.(starlark.String)
				if !ok {
					return nil, fmt.Errorf("list element must be str, got %s", x.Type())
				}
				strs = append(strs, string(s))
				continue
			}
			f, err := toGo("float", x)
			if err != nil {
				return nil, err
			}
			floats = append(floats, f.(float64))
		}
		if typ == "list[str]" {
			if strs == nil {
				strs = []string{}
			}
			return strs, nil
		}
		if floats == nil {
			floats = []float64{}
		}
		return floats, nil
	case "dict[str, str]":
		d, ok := v.(*starlark.Dict)
		if !ok {
			return nil, fmt.Errorf("want a dict, got %s", v.Type())
		}
		out := map[string]string{}
		for _, kv := range d.Items() {
			k, ok1 := kv[0].(starlark.String)
			s, ok2 := kv[1].(starlark.String)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("dict keys and values must be str")
			}
			out[string(k)] = string(s)
		}
		return out, nil
	}
	return nil, fmt.Errorf("unsupported param type %q", typ)
}

// fromGo converts a tools result to a Starlark value.
func fromGo(v any) (starlark.Value, error) {
	switch x := v.(type) {
	case nil:
		return starlark.None, nil
	case string:
		return starlark.String(x), nil
	case int64:
		return starlark.MakeInt64(x), nil
	case int:
		return starlark.MakeInt(x), nil
	case float64:
		if math.IsNaN(x) || math.IsInf(x, 0) {
			return starlark.Float(x), nil
		}
		return starlark.Float(x), nil
	case bool:
		return starlark.Bool(x), nil
	case []string:
		l := make([]starlark.Value, len(x))
		for i, s := range x {
			l[i] = starlark.String(s)
		}
		return starlark.NewList(l), nil
	case []float64:
		l := make([]starlark.Value, len(x))
		for i, f := range x {
			l[i] = starlark.Float(f)
		}
		return starlark.NewList(l), nil
	case map[string]string:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		d := starlark.NewDict(len(keys))
		for _, k := range keys {
			if err := d.SetKey(starlark.String(k), starlark.String(x[k])); err != nil {
				return nil, err
			}
		}
		return d, nil
	}
	return nil, fmt.Errorf("tools returned unsupported type %T", v)
}

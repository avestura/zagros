// Package toolstest holds the quality checks every tools group must pass.
// Each group package calls Check from one test.
package toolstest

import (
	"strings"
	"testing"

	"github.com/avestura/zagros/internal/tools"
)

var validReturns = map[string]bool{
	"str": true, "int": true, "float": true, "bool": true, "None": true,
	"list[str]": true, "list[float]": true, "dict[str, str]": true, "any": true,
}

var validParamTypes = map[string]bool{
	"str": true, "int": true, "float": true, "bool": true,
	"list[str]": true, "list[float]": true, "dict[str, str]": true, "any": true,
}

// Groups are the allowed Func.Group values.
var Groups = map[string]bool{
	"text": true, "lines": true, "num": true, "math": true, "encode": true, "hash": true, "cipher": true,
	"rand": true, "uuid": true, "regex": true, "net": true, "dt": true, "extract": true,
	"yaml": true, "toml": true, "csv": true, "xml": true, "html": true,
	"url": true, "path": true, "template": true, "jwt": true,
}

// Check verifies the registrations of one group (all groups when group is "").
// It enforces what the generated docs, stubs, launcher and MCP search rely on.
// It also fails when the group registered nothing.
func Check(t *testing.T, group string) {
	t.Helper()
	n := 0
	for _, f := range tools.All() {
		if group != "" && f.Group != group {
			continue
		}
		n++
		k := f.Key()
		if !Groups[f.Group] {
			t.Errorf("%s: unknown group %q", k, f.Group)
		}
		if len(f.Doc) < 20 {
			t.Errorf("%s: Doc is too short to search by: %q", k, f.Doc)
		}
		if !validReturns[f.Returns] {
			t.Errorf("%s: bad Returns %q", k, f.Returns)
		}
		if !strings.Contains(f.Example, f.Group+"."+f.Name+"(") {
			t.Errorf("%s: Example must call %s.%s(...): %q", k, f.Group, f.Name, f.Example)
		}
		if len(f.Tags) == 0 {
			t.Errorf("%s: Tags are required (search words)", k)
		}
		sawOptional := false
		for _, p := range f.Params {
			if !p.Required {
				sawOptional = true
			} else if sawOptional {
				t.Errorf("%s: required parameter %s comes after an optional one (Starlark and the stubs need required parameters first)", k, p.Name)
			}
			if !validParamTypes[p.Type] {
				t.Errorf("%s: parameter %s has bad type %q", k, p.Name, p.Type)
			}
			if p.Doc == "" {
				t.Errorf("%s: parameter %s needs a Doc", k, p.Name)
			}
			if p.Required && p.Default != nil {
				t.Errorf("%s: required parameter %s has a default", k, p.Name)
			}
			if !p.Required && p.Default != nil && !defaultMatches(p) {
				t.Errorf("%s: default of %s has the wrong Go type %T for %s", k, p.Name, p.Default, p.Type)
			}
		}
		if f.Family != "" && f.InputParam() == "" {
			t.Errorf("%s: a launcher variant needs a str input parameter", k)
		}
		if f.Input != "" {
			found := false
			for _, p := range f.Params {
				if p.Name == f.Input && p.Type == "str" {
					found = true
				}
			}
			if !found {
				t.Errorf("%s: Input %q is not a str parameter", k, f.Input)
			}
		}
	}
	if n == 0 {
		t.Errorf("group %q registered no functions", group)
	}
}

func defaultMatches(p tools.Param) bool {
	switch p.Type {
	case "str":
		_, ok := p.Default.(string)
		return ok
	case "int":
		_, ok := p.Default.(int64)
		return ok
	case "float":
		_, ok := p.Default.(float64)
		return ok
	case "bool":
		_, ok := p.Default.(bool)
		return ok
	case "list[str]":
		_, ok := p.Default.([]string)
		return ok
	case "list[float]":
		_, ok := p.Default.([]float64)
		return ok
	case "dict[str, str]":
		_, ok := p.Default.(map[string]string)
		return ok
	case "any":
		return true
	}
	return false
}

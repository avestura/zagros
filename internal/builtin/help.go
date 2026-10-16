package builtin

import (
	"context"
	"sort"
	"strings"

	"github.com/avestura/zagros/internal/core"
)

type help struct {
	base
	providers []core.Provider
}

func newHelp(ps []core.Provider) *help {
	return &help{base{"help"}, append([]core.Provider(nil), ps...)}
}

func (h *help) Ranked() bool       { return true }
func (h *help) Keywords() []string { return []string{"?", "help"} }

var keywordDocs = map[string]string{
	"f": "search files with Everything", "=": "calculator (also plain expressions like 2*(3+4))",
	"pw": "generate passwords: pw 24 nosym nonum noupper nolower noamb", "pass": "same as pw",
	"pp": "passphrases: pp 6 cap num", "pin": "numeric PINs: pin 6",
	"uuid": "random UUID v4", "b64": "Base64 encode", "b64d": "Base64 decode", "url": "URL-encode", "urld": "URL-decode",
	"hash": "MD5/SHA-1/SHA-256/SHA-512", "case": "case conversions", "lorem": "lorem ipsum: lorem 50",
	"ts": "timestamp converter: ts 1700000000 | ts 2024-01-31", "now": "current time formats",
	"json": "pretty/minify JSON from the clipboard", "color": "color conversion: color #ff8800", "unit": "unit converter: unit 10 km to mi",
	"ip": "local IP addresses", "cb": "clipboard history (opt-in)", "w": "switch to an open window",
	"kill": "kill a process by name", ":": "emoji picker: :smile", "emoji": "emoji picker", ">": "run a command in cmd.exe",
	"ps": "run a command in PowerShell", "recent": "recent files", "?": "this list",
}

func (h *help) Query(ctx context.Context, q string) []core.Item {
	rest, _, ok := trigger(q, "?", "help")
	if !ok {
		if t := strings.TrimSpace(q); t != "?" && t != "help" {
			return nil
		}
	}
	type row struct{ kw, owner string }
	var rows []row
	for _, p := range h.providers {
		kp, ok := p.(keywordProvider)
		if !ok {
			continue
		}
		for _, k := range kp.Keywords() {
			rows = append(rows, row{k, p.Name()})
		}
	}
	rows = append(rows, row{"?", "help"})
	sort.Slice(rows, func(i, j int) bool { return rows[i].kw < rows[j].kw })
	var out []core.Item
	for _, r := range rows {
		desc := keywordDocs[r.kw]
		if desc == "" {
			desc = r.owner
		}
		if rest != "" && !containsFold(r.kw+" "+desc, rest) {
			continue
		}
		out = append(out, core.Item{
			ID: "help:" + r.kw, Title: r.kw, Subtitle: desc, Category: "Help",
			Action: core.Action{Kind: "setquery", Target: r.kw + " "},
		})
	}
	return out
}

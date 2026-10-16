package builtin

import (
	"context"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/avestura/zagros/internal/core"
)

type websearch struct {
	base
	cfg *core.Config
}

func newWebSearch(cfg *core.Config) *websearch { return &websearch{base{"websearch"}, cfg} }

func (w *websearch) Ranked() bool { return true }

func (w *websearch) Keywords() []string {
	var k []string
	for _, s := range w.cfg.WebSearch {
		k = append(k, s.Keyword)
	}
	for a := range w.cfg.Aliases {
		k = append(k, a)
	}
	sort.Strings(k)
	return k
}

var (
	schemeRe = regexp.MustCompile(`(?i)^(https?|ftp)://\S+$`)
	hostRe   = regexp.MustCompile(`(?i)^(?:localhost|(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+([a-z]{2,24})|(?:\d{1,3}\.){3}\d{1,3})(?::\d{1,5})?(?:[/?#]\S*)?$`)
)

var knownTLDs = map[string]bool{}

func init() {
	for _, t := range strings.Fields(`com org net io dev app co ai me edu gov info biz xyz tv us uk de fr ru ir ca au in jp cn br it es nl se no
		fi pl ch at be eu cc ly sh so to gg fm im cloud tech online site store blog wiki news page link`) {
		knownTLDs[t] = true
	}
}

// bareURL returns an openable URL when q looks like one ("example.com/a").
func bareURL(q string) (string, bool) {
	if strings.ContainsAny(q, " \t") {
		return "", false
	}
	if schemeRe.MatchString(q) {
		return q, true
	}
	m := hostRe.FindStringSubmatch(q)
	if m == nil {
		return "", false
	}
	if m[1] != "" && !knownTLDs[strings.ToLower(m[1])] {
		return "", false
	}
	return "https://" + q, true
}

func (w *websearch) Query(ctx context.Context, q string) []core.Item {
	q = strings.TrimSpace(q)
	if q == "" {
		return nil
	}
	var out []core.Item
	if u, ok := bareURL(q); ok {
		out = append(out, core.Item{ID: "websearch:url", Title: "Open " + u, Subtitle: "URL", Category: "Web",
			Action: core.Action{Kind: "open", Target: u}})
	}
	first, rest, _ := strings.Cut(q, " ")
	rest = strings.TrimSpace(rest)
	for _, s := range w.cfg.WebSearch {
		if !strings.EqualFold(first, s.Keyword) {
			continue
		}
		if rest == "" {
			out = append(out, hint("websearch", s.Keyword, "Search "+s.Name, "type your query after \""+s.Keyword+" \""))
			continue
		}
		out = append(out, core.Item{
			ID: "websearch:" + strings.ToLower(s.Keyword), Title: "Search " + s.Name + " for \"" + rest + "\"",
			Subtitle: s.URL, Category: "Web",
			Action: core.Action{Kind: "open", Target: strings.ReplaceAll(s.URL, "{q}", url.QueryEscape(rest))},
		})
	}
	for key, target := range w.cfg.Aliases {
		if !strings.EqualFold(first, key) {
			continue
		}
		out = append(out, aliasItem(key, target, rest))
	}
	return out
}

func aliasItem(key, target, rest string) core.Item {
	t := target
	if strings.Contains(t, "{q}") {
		esc := rest
		if isURLish(t) {
			esc = url.QueryEscape(rest)
		}
		t = strings.ReplaceAll(t, "{q}", esc)
	}
	it := core.Item{ID: "websearch:alias:" + strings.ToLower(key), Title: key, Subtitle: t, Category: "Alias"}
	if isURLish(t) || fileExists(t) {
		it.Action = core.Action{Kind: "open", Target: t}
	} else {
		it.Action = core.Action{Kind: "run", Target: t}
	}
	return it
}

func isURLish(s string) bool {
	return schemeRe.MatchString(s) || strings.HasPrefix(strings.ToLower(s), "shell:") || strings.Contains(s, "://") || strings.HasPrefix(s, "ms-")
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

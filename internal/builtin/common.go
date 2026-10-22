// Package builtin contains Zagros' built-in providers.
package builtin

import (
	"strings"

	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/tools"
)

// keywordProvider mirrors core.KeywordProvider so this package does not depend
// on it landing first. Providers with a trigger implement it.
type keywordProvider interface{ Keywords() []string }

// All returns the built-in providers enabled by cfg. A provider is enabled
// unless cfg.Providers[name] is false; "clipboard" is opt-in and needs an
// explicit true; "everything" additionally needs cfg.Everything.Enabled.
// Nothing is indexed until the first query.
func All(cfg *core.Config) []core.Provider {
	if cfg == nil {
		cfg = &core.Config{}
	}
	on := func(n string) bool { v, ok := cfg.Providers[n]; return !ok || v }
	var ps []core.Provider
	add := func(name string, mk func() core.Provider) {
		if on(name) {
			ps = append(ps, mk())
		}
	}
	add("apps", func() core.Provider {
		a := newApps(cfg)
		theApps.Store(a)
		return a
	})
	if cfg.Everything.Enabled {
		add("everything", func() core.Provider { return newEverything(cfg) })
	}
	add("calc", func() core.Provider { return newCalc() })
	add("password", func() core.Provider { return newPassword() })
	toolsOn := on("tools") && len(tools.All()) > 0
	add("text", func() core.Provider {
		t := newText()
		if toolsOn {
			t.skip = map[string]bool{}
			for _, k := range toolsOwnedKeywords {
				t.skip[k] = true
			}
		}
		return t
	})
	if toolsOn {
		add("tools", func() core.Provider { return newTools() })
	}
	add("system", func() core.Provider { return newSystem() })
	add("websearch", func() core.Provider { return newWebSearch(cfg) })
	if cfg.Providers["clipboard"] {
		ps = append(ps, newClipboard())
	}
	add("windows", func() core.Provider { return newWindows() })
	add("processes", func() core.Provider { return newProcesses() })
	add("emoji", func() core.Provider { return newEmoji() })
	add("runner", func() core.Provider { return newRunner() })
	add("update", func() core.Provider { return newUpdate() })
	if on("apps") {
		add("reindex", func() core.Provider { return newReindex() })
	}
	add("mcp", func() core.Provider { return newMCP() })
	add("recent", func() core.Provider { return newRecent() })
	if on("help") {
		ps = append(ps, newHelp(ps))
	}
	return ps
}

type base struct{ name string }

func (b base) Name() string { return b.name }

// trigger reports whether q starts with one of the keywords as a whole word
// and returns the remaining argument text (trimmed) and the matched keyword.
func trigger(q string, kws ...string) (rest, kw string, ok bool) {
	q = strings.TrimLeft(q, " \t")
	end := strings.IndexAny(q, " \t")
	first, rest := q, ""
	if end >= 0 {
		first, rest = q[:end], strings.TrimSpace(q[end:])
	}
	for _, k := range kws {
		if strings.EqualFold(first, k) {
			return rest, k, true
		}
	}
	return "", "", false
}

func copyAction(s string) core.Action { return core.Action{Kind: "copy", Target: s} }

// noHistory marks an action whose item must never be stored in history.
func noHistoryAction(a core.Action) core.Action {
	if a.Extra == nil {
		a.Extra = map[string]string{}
	}
	a.Extra["nohistory"] = "1"
	return a
}

// subseq reports whether the runes of needle occur in order in hay (case-insensitive).
func subseq(needle, hay string) bool {
	n := []rune(strings.ToLower(needle))
	i := 0
	for _, r := range strings.ToLower(hay) {
		if i < len(n) && r == n[i] {
			i++
		}
	}
	return i == len(n)
}

func containsFold(hay, needle string) bool {
	return strings.Contains(strings.ToLower(hay), strings.ToLower(needle))
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n-1]) + "…"
}

func hint(provider, id, title, sub string) core.Item {
	return core.Item{ID: provider + ":" + id, Title: title, Subtitle: sub, Category: provider, Action: core.Action{Kind: "none"}}
}

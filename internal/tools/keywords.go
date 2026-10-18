package tools

import (
	"fmt"
	"strings"
)

// Annotate adds search metadata to an already registered function: task is
// the headline task phrase ("unzip a file") that lists it in the generated
// cheat sheet; terms are synonyms and task phrases for the stub's Keywords
// line. It panics on an unknown key so a rename cannot leave stale metadata.
func Annotate(key, task string, terms ...string) {
	mu.Lock()
	defer mu.Unlock()
	i, ok := index[key]
	if !ok {
		panic(fmt.Sprintf("tools: Annotate of unknown function %q", key))
	}
	f := &funcs[i]
	if task != "" {
		f.Task = task
	}
	f.Terms = append(f.Terms, terms...)
}

// SearchTerms is the Keywords line of f: Terms first (phrases written for
// search), then Tags, without duplicates. It is never empty for a function
// that has Tags, so older registrations keep working without Terms.
func (f Func) SearchTerms() []string {
	seen := map[string]bool{}
	var out []string
	for _, group := range [][]string{f.Terms, f.Tags} {
		for _, t := range group {
			t = strings.TrimSpace(strings.ReplaceAll(t, ",", " "))
			k := strings.ToLower(t)
			if t == "" || seen[k] {
				continue
			}
			seen[k] = true
			out = append(out, t)
		}
	}
	return out
}

package mcp

import (
	"sort"
	"strings"
	"sync"
)

// Field weights, highest first: the name, then Keywords and Task phrases,
// then the documentation body. A synonym of a query word scores lower than
// the word itself.
const (
	wName, wNamePrefix, wNameSyn = 20, 8, 12
	wKey, wKeySyn                = 12, 8
	wDoc, wDocSyn                = 5, 3
	wPhrase, wPhraseExact        = 25, 15
)

type qWord struct {
	stem string
	alts []string // stemmed synonyms
}

// queryWords turns a query into stemmed words with their synonyms. Filler
// words ("a", "the", "into") are dropped unless nothing else is left.
func queryWords(query string) []qWord {
	var all, kept []string
	for _, t := range tokens(query) {
		all = append(all, t)
		if !stopwords[t] {
			kept = append(kept, t)
		}
	}
	if len(kept) == 0 {
		kept = all
	}
	var out []qWord
	seen := map[string]bool{}
	for _, t := range kept {
		w := qWord{stem: stem(t)}
		if seen[w.stem] {
			continue
		}
		seen[w.stem] = true
		for _, s := range synonymOf[t] {
			if st := stem(s); st != w.stem {
				w.alts = appendUnique(w.alts, st)
			}
		}
		out = append(out, w)
	}
	return out
}

type meta struct {
	name    []string // stemmed name and module words, repeats kept
	key     map[string]bool
	phrases [][]string // stemmed Keywords and Task phrases without filler words
	doc     map[string]bool
}

var (
	metaMu sync.Mutex
)

func stems(s string, dropStop bool) []string {
	var out []string
	for _, t := range tokens(s) {
		if dropStop && stopwords[t] {
			continue
		}
		out = append(out, stem(t))
	}
	return out
}

func buildMeta(e Entry) meta {
	m := meta{key: map[string]bool{}, doc: map[string]bool{}}
	m.name = stems(strings.ReplaceAll(e.Qualified, "_", " ")+" "+e.Module, false)
	phrases := append([]string(nil), e.Keywords...)
	if e.Task != "" {
		phrases = append(phrases, e.Task)
	}
	for _, p := range phrases {
		ws := stems(p, true)
		if len(ws) == 0 {
			continue
		}
		m.phrases = append(m.phrases, ws)
		for _, w := range ws {
			m.key[w] = true
		}
	}
	for _, w := range stems(e.Doc+" "+e.Signature, false) {
		m.doc[w] = true
	}
	return m
}

func (x *Index) metaFor() []meta {
	metaMu.Lock()
	defer metaMu.Unlock()
	if x.metas == nil || len(x.metas) != len(x.Entries) {
		x.metas = make([]meta, len(x.Entries))
		for i, e := range x.Entries {
			x.metas[i] = buildMeta(e)
		}
	}
	return x.metas
}

func has(set []string, w string) bool {
	for _, s := range set {
		if s == w {
			return true
		}
	}
	return false
}

// inPhrase reports whether the query word or one of its synonyms is in p.
func (w qWord) inPhrase(p []string) bool {
	if has(p, w.stem) {
		return true
	}
	for _, a := range w.alts {
		if has(p, a) {
			return true
		}
	}
	return false
}

// Search ranks entries for a free-text query: exact names first, then name and
// module words, then Keywords and Task phrases (synonyms and task wording),
// then words in the documentation. Query words are stemmed and expanded with
// a small synonym table. module narrows to one module ("" for any, "global"
// for the host-level names).
func (x *Index) Search(query, module string, limit int) []Match {
	if limit <= 0 {
		limit = 10
	}
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return nil
	}
	qw := queryWords(query)
	metas := x.metaFor()
	var out []Match
	for i, e := range x.Entries {
		if e.Kind == "module" && module != "" {
			continue
		}
		if module != "" && !(strings.EqualFold(e.Module, module) || (module == "global" && e.Module == "")) {
			continue
		}
		m := metas[i]
		score := 0
		name, qual := strings.ToLower(e.Name), strings.ToLower(e.Qualified)
		switch {
		case q == qual || q == name:
			score += 100
		case strings.Contains(qual, q):
			score += 40
		}
		covered := 0
		for _, w := range qw {
			hit := false
			// name
			direct, syn := 0, 0
			for _, n := range m.name {
				if n == w.stem {
					direct += wName
				} else if len(w.stem) >= 4 && strings.HasPrefix(n, w.stem) {
					direct += wNamePrefix
				} else if has(w.alts, n) {
					syn = wNameSyn
				}
			}
			if direct > 0 {
				score += direct
				hit = true
			} else if syn > 0 {
				score += syn
				hit = true
			}
			// keywords
			switch {
			case m.key[w.stem]:
				score += wKey
				hit = true
			default:
				for _, a := range w.alts {
					if m.key[a] {
						score += wKeySyn
						hit = true
						break
					}
				}
			}
			// documentation
			switch {
			case m.doc[w.stem]:
				score += wDoc
				hit = true
			default:
				for _, a := range w.alts {
					if m.doc[a] {
						score += wDocSyn
						hit = true
						break
					}
				}
			}
			if hit {
				covered++
			}
		}
		// A Keywords phrase that holds every query word is the strongest sign
		// the entry is the answer to a task-style query.
		best := 0
		for _, p := range m.phrases {
			all := true
			for _, w := range qw {
				if !w.inPhrase(p) {
					all = false
					break
				}
			}
			if !all {
				continue
			}
			b := wPhrase
			if len(qw) == 1 {
				b = wKey
			}
			if len(p) == len(qw) {
				b += wPhraseExact
			}
			if b > best {
				best = b
			}
		}
		score += best
		// Entries that account for every query word beat those that match one
		// word loosely; this is what makes task-style queries work.
		if len(qw) > 1 && covered == len(qw) {
			score += 30 * covered
		}
		if score > 0 {
			if e.Kind == "module" && score < 100 {
				score -= 8 // a function that answers the task beats its module
			}
			out = append(out, Match{Entry: e, Score: score})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Score != out[j].Score {
			return out[i].Score > out[j].Score
		}
		return out[i].Qualified < out[j].Qualified
	})
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

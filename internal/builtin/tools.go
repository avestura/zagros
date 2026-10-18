package builtin

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/avestura/zagros/internal/core"
	"github.com/avestura/zagros/internal/tools"
)

// The "tools" provider is the launcher face of the tools registry
// (internal/tools). Every registered function that names a Family appears as
// a variant of that family:
//
//	b64 hello          the likely variants for the text, best first
//	hash hello         every cheap hash of the text
//	enc  hello         all the encodings
//	Tab on a row       enters the family scope: every variant for the same text,
//	                   with "@word" to narrow the list
//
// Syntax shared by all families:
//
//	key=value          sets a named parameter (salt=abc iterations=1000)
//	a :: b             arguments for functions that take several strings
//	                   (re \d+ :: abc 123, cidr 10.0.0.0/8 :: 10.1.2.3)
//	empty input        uses the clipboard text
//
// The "clip" family applies the text/lines/encode/extract tools to the
// clipboard and puts the result back with Enter (it copies it).

type toolFamily struct {
	id, title, placeholder string
	keywords               []string
	// noClip families never fall back to the clipboard when the input is empty.
	noClip bool
}

var toolFamilies = []toolFamily{
	{id: "text", title: "Text", placeholder: "Text to change or measure, @filter to narrow", keywords: []string{"text", "t", "str", "string"}},
	{id: "lines", title: "Lines", placeholder: "Lines of text, @filter to narrow", keywords: []string{"lines"}},
	{id: "num", title: "Numbers", placeholder: "A number or 0xFF + 0b101, @filter to narrow", keywords: []string{"num", "number", "n", "base", "bits"}},
	{id: "math", title: "Math", placeholder: "An expression or numbers, @filter to narrow", keywords: []string{"math", "stats"}},
	{id: "encode", title: "Encode", placeholder: "Text to encode or decode, @filter to narrow", keywords: []string{"enc", "encode", "dec", "decode"}},
	{id: "hash", title: "Hash", placeholder: "Text to hash, salt=... key=... for keyed ones", keywords: []string{"hash"}},
	{id: "cipher", title: "Cipher", placeholder: "Text, then key=secret", keywords: []string{"crypt", "cipher", "encrypt", "decrypt"}},
	{id: "rand", title: "Random", placeholder: "Press Tab again for new values", keywords: []string{"rand", "random"}, noClip: true},
	{id: "uuid", title: "IDs", placeholder: "Press Tab again for new IDs, or paste an ID to parse it", keywords: []string{"uuid", "guid", "id"}, noClip: true},
	{id: "regex", title: "Regex", placeholder: "pattern :: text", keywords: []string{"re", "regex"}},
	{id: "net", title: "Network", placeholder: "An IP, CIDR, URL, MAC or port, @filter to narrow", keywords: []string{"net", "network"}},
	{id: "dt", title: "Date and time", placeholder: "A date, timestamp or duration, @filter to narrow", keywords: []string{"dt", "date", "time"}},
	{id: "extract", title: "Extract", placeholder: "Text to pull things out of", keywords: []string{"ex", "extract"}},
	{id: "clip", title: "Clipboard", placeholder: "Filter the transformations, e.g. upper, sort, base64", keywords: []string{"clip"}},
}

// clipSources are the families whose variants the clip family applies to the
// clipboard text.
var clipSources = []string{"text", "lines", "encode", "extract"}

// aliases keep the keywords of the old text provider working.
var toolAliases = []struct {
	kw, group, name string
}{
	{"case", "text", "case_all"},
	{"slug", "text", "slug"},
	{"ts", "dt", "convert"},
	{"now", "dt", "now_all"},
	{"epoch", "dt", "convert"},
	{"flip", "rand", "coin_flip"},
}

var slowRe = regexp.MustCompile(`bcrypt|scrypt|argon2|pbkdf2|rsa_generate|rsa_sign|rsa_encrypt|rsa_decrypt`)

const (
	toolsRootRows  = 12
	toolsScopeRows = 400
)

type toolIndex struct {
	families map[string]*toolFamily
	famByKw  map[string]*toolFamily
	byFamily map[string][]tools.Func
	byKw     map[string][]tools.Func
	byName   map[string][]tools.Func
}

type toolsProvider struct {
	base
	once sync.Once
	idx  *toolIndex
}

func newTools() *toolsProvider { return &toolsProvider{base: base{"tools"}} }

func (p *toolsProvider) Ranked() bool { return true }

func (p *toolsProvider) index() *toolIndex {
	p.once.Do(func() {
		ix := &toolIndex{families: map[string]*toolFamily{}, famByKw: map[string]*toolFamily{},
			byFamily: map[string][]tools.Func{}, byKw: map[string][]tools.Func{}, byName: map[string][]tools.Func{}}
		for i := range toolFamilies {
			f := &toolFamilies[i]
			ix.families[f.id] = f
			for _, k := range f.keywords {
				ix.famByKw[k] = f
			}
		}
		for _, f := range tools.All() { // ordered by group, Order, name
			if f.Family == "" {
				continue
			}
			ix.byFamily[f.Family] = append(ix.byFamily[f.Family], f)
			for _, k := range f.Keywords {
				k = strings.ToLower(k)
				ix.byKw[k] = append(ix.byKw[k], f)
			}
			ix.byName[strings.ToLower(f.Name)] = append(ix.byName[strings.ToLower(f.Name)], f)
		}
		for id := range ix.byFamily {
			fs := ix.byFamily[id]
			sort.SliceStable(fs, func(i, j int) bool { return fs[i].Order < fs[j].Order })
		}
		p.idx = ix
	})
	return p.idx
}

// Keywords are the launcher entry points: family and function keywords.
func (p *toolsProvider) Keywords() []string {
	ix := p.index()
	seen := map[string]bool{}
	var out []string
	add := func(k string) {
		if k != "" && !seen[k] {
			seen[k] = true
			out = append(out, k)
		}
	}
	for k := range ix.famByKw {
		add(k)
	}
	for k := range ix.byKw {
		add(k)
	}
	for _, a := range toolAliases {
		add(a.kw)
	}
	sort.Strings(out)
	return out
}

// ---- input parsing ----

var paramTokenRe = regexp.MustCompile(`(?:^|\s)([a-z_][a-z0-9_]*)=("[^"]*"|\S+)`)

// toolInput is the parsed text after the keyword.
type toolInput struct {
	raw    string            // exactly what the user typed after the keyword
	text   string            // the main input (without params)
	args   []string          // segments split by " :: " (len>=2 only when used)
	params map[string]string // name=value tokens
	filter []string          // @words that narrow a scope list
	clip   bool              // text came from the clipboard
}

func parseToolInput(raw string, knownParams map[string]bool) toolInput {
	in := toolInput{raw: raw, params: map[string]string{}}
	rest := raw
	// @filter words at the start
	for {
		rest = strings.TrimLeft(rest, " ")
		if strings.HasPrefix(rest, "@") {
			w, tail, _ := strings.Cut(rest[1:], " ")
			if w != "" {
				in.filter = append(in.filter, strings.ToLower(w))
			}
			rest = tail
			continue
		}
		break
	}
	// name=value tokens (only for parameters some variant of the family knows)
	rest = paramTokenRe.ReplaceAllStringFunc(rest, func(m string) string {
		sub := paramTokenRe.FindStringSubmatch(m)
		if !knownParams[sub[1]] {
			return m
		}
		in.params[sub[1]] = strings.Trim(sub[2], `"`)
		return " "
	})
	rest = strings.TrimSpace(rest)
	if strings.Contains(rest, " :: ") {
		for _, s := range strings.Split(rest, " :: ") {
			in.args = append(in.args, strings.TrimSpace(s))
		}
		in.text = in.args[len(in.args)-1]
	} else {
		in.text = rest
	}
	return in
}

func knownParamNames(fs []tools.Func) map[string]bool {
	m := map[string]bool{}
	for _, f := range fs {
		in := f.InputParam()
		for _, p := range f.Params {
			if p.Name != in {
				m[p.Name] = true
			}
		}
	}
	return m
}

// convert turns a typed string into the Go type of a param.
func convertParam(typ, s string) (any, error) {
	switch typ {
	case "str":
		return s, nil
	case "int":
		n, err := strconv.ParseInt(strings.ReplaceAll(s, "_", ""), 0, 64)
		return n, err
	case "float":
		return strconv.ParseFloat(s, 64)
	case "bool":
		switch strings.ToLower(s) {
		case "1", "true", "yes", "y", "on":
			return true, nil
		case "0", "false", "no", "n", "off":
			return false, nil
		}
		return nil, fmt.Errorf("not a bool: %q", s)
	case "list[str]":
		var out []string
		for _, f := range strings.Split(s, ",") {
			out = append(out, strings.TrimSpace(f))
		}
		return out, nil
	case "list[float]":
		var out []float64
		for _, f := range strings.FieldsFunc(s, func(r rune) bool { return r == ',' || r == ' ' || r == ';' }) {
			v, err := strconv.ParseFloat(f, 64)
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	}
	return nil, fmt.Errorf("unsupported parameter type %s", typ)
}

// runVariant calls f for the input. missing lists required parameters the
// user did not supply (the variant is then not shown).
func runVariant(f tools.Func, in toolInput) (res any, missing []string, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()
	named := map[string]any{}
	inName := f.InputParam()
	// "a :: b" arguments fill the str parameters in declared order.
	var strParams []string
	for _, p := range f.Params {
		if p.Type == "str" && (p.Required || p.Name == inName) {
			strParams = append(strParams, p.Name)
		}
	}
	if len(in.args) >= 2 {
		for i, seg := range in.args {
			if i < len(strParams) {
				named[strParams[i]] = seg
			}
		}
	} else if inName != "" {
		named[inName] = in.text
	}
	for _, p := range f.Params {
		if v, ok := in.params[p.Name]; ok && p.Name != inName {
			cv, cerr := convertParam(p.Type, v)
			if cerr != nil {
				return nil, nil, fmt.Errorf("%s: %w", p.Name, cerr)
			}
			named[p.Name] = cv
		}
		if _, ok := named[p.Name]; !ok && p.Required {
			missing = append(missing, p.Name)
		}
	}
	if len(missing) > 0 {
		return nil, missing, nil
	}
	res, err = f.Invoke(named)
	return res, nil, err
}

func oneLine(s string) string {
	s = strings.ReplaceAll(strings.ReplaceAll(s, "\r\n", "\n"), "\n", " ⏎ ")
	return truncate(s, 120)
}

// rowsFor turns a result into launcher rows. Dicts give one row per key.
func (p *toolsProvider) rowsFor(f tools.Func, res any, fam *toolFamily, in toolInput, sc *core.Scope, prefer string) []core.Item {
	mk := func(id, title, sub, value string) core.Item {
		return core.Item{ID: "tools:" + id, Title: oneLine(title), Subtitle: sub, Category: fam.title,
			Action: noHistoryAction(copyAction(value)), Scope: sc}
	}
	label := f.Title
	if in.clip {
		label += " (clipboard)"
	}
	switch x := res.(type) {
	case map[string]string:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var out []core.Item
		for _, k := range keys {
			if x[k] == "" || x[k] == in.text {
				continue
			}
			out = append(out, mk(f.Key()+":"+k, x[k], k+" - "+label, x[k]))
		}
		return out
	case []string:
		if len(x) == 0 {
			return nil
		}
		joined := strings.Join(x, "\n")
		sub := label
		if len(x) > 1 {
			sub = fmt.Sprintf("%s (%d)", label, len(x))
		}
		return []core.Item{mk(f.Key(), strings.Join(x, ", "), sub, joined)}
	}
	v := tools.Format(res)
	if v == "" {
		return nil
	}
	if s, ok := res.(string); ok && s == in.text && in.text != "" {
		return nil // unchanged text is noise
	}
	return []core.Item{mk(f.Key(), v, label, v)}
}

func (p *toolsProvider) scopeFor(fam *toolFamily, in toolInput, prefer string) *core.Scope {
	return &core.Scope{ID: "tools." + fam.id, Label: fam.title, Placeholder: fam.placeholder, Provider: "tools",
		Data: map[string]string{"family": fam.id, "prefer": prefer}, Query: in.raw}
}

// ---- resolving a keyword ----

type toolSel struct {
	fam    *toolFamily
	funcs  []tools.Func
	prefer string // the keyword that selected a subset, listed first in scopes
	// viaFamily is true when the user typed a family keyword (text, enc, net);
	// only then may an unanswerable input show an explanatory hint row.
	viaFamily bool
	// explicit is true for family keywords and the legacy aliases (now, ts,
	// case): only those may answer a bare keyword with no input typed.
	explicit bool
}

func (p *toolsProvider) resolve(kw string) (toolSel, bool) {
	ix := p.index()
	kw = strings.ToLower(kw)
	if fam, ok := ix.famByKw[kw]; ok {
		sel := p.familySel(fam, "")
		sel.viaFamily, sel.explicit = true, true
		return sel, true
	}
	for _, a := range toolAliases {
		if a.kw == kw {
			if f, ok := tools.Lookup(a.group, a.name); ok && f.Family != "" {
				return toolSel{fam: ix.families[f.Family], funcs: []tools.Func{f}, prefer: kw, explicit: true}, true
			}
		}
	}
	if fs := ix.byKw[kw]; len(fs) > 0 {
		// A keyword shared by several families ("hex": number conversion and
		// text encoding) lists the more specific family first.
		fs = append([]tools.Func(nil), fs...)
		sort.SliceStable(fs, func(i, j int) bool { return famRank(fs[i].Family) < famRank(fs[j].Family) })
		return toolSel{fam: ix.families[fs[0].Family], funcs: fs, prefer: kw}, true
	}
	// A bare function name is a keyword only when it is technical (sha256,
	// base64_encode); plain words such as "reverse" or "count" would hijack
	// ordinary launcher queries, so those need an explicit Func.Keywords entry.
	if fs := ix.byName[kw]; len(fs) == 1 && strings.ContainsAny(kw, "_0123456789") {
		return toolSel{fam: ix.families[fs[0].Family], funcs: fs, prefer: kw}, true
	}
	return toolSel{}, false
}

// famRank orders families when one keyword belongs to several of them.
func famRank(id string) int {
	for i, f := range []string{"num", "math", "dt", "net", "uuid", "encode", "hash", "cipher", "text", "lines", "regex", "extract", "rand"} {
		if f == id {
			return i
		}
	}
	return 99
}

// familySel returns the variants of fam. The clip family is the union of the
// text-like families.
func (p *toolsProvider) familySel(fam *toolFamily, prefer string) toolSel {
	ix := p.index()
	if fam.id != "clip" {
		return toolSel{fam: fam, funcs: ix.byFamily[fam.id], prefer: prefer}
	}
	var fs []tools.Func
	for _, id := range clipSources {
		for _, f := range ix.byFamily[id] {
			if singleInput(f) {
				fs = append(fs, f)
			}
		}
	}
	return toolSel{fam: fam, funcs: fs, prefer: prefer}
}

// singleInput reports whether f needs nothing but the input text.
func singleInput(f tools.Func) bool {
	in := f.InputParam()
	if in == "" {
		return false
	}
	for _, p := range f.Params {
		if p.Required && p.Name != in {
			return false
		}
	}
	return true
}

// ---- Query ----

func (p *toolsProvider) Query(ctx context.Context, q string) []core.Item {
	if raw := core.RawQuery(ctx); raw != "" {
		q = raw // keeps the trailing space: "b64 " means "use the clipboard"
	}
	q = strings.TrimLeft(q, " \t")
	if q == "" {
		return nil
	}
	kw, rest, _ := strings.Cut(q, " ")
	sel, ok := p.resolve(kw)
	if !ok {
		return p.autoDetect(ctx, q)
	}
	bare := !strings.Contains(q, " ")
	if bare && !sel.explicit {
		return nil // "format", "age": ordinary words alone are not tool queries
	}
	return p.run(ctx, sel, rest, false, bare)
}

// run evaluates the variants of sel for the raw input. all=true is the
// scope list (slow variants included, more rows).
func (p *toolsProvider) run(ctx context.Context, sel toolSel, rawInput string, all, bare bool) []core.Item {
	fam := sel.fam
	in := parseToolInput(rawInput, knownParamNames(sel.funcs))
	funcs := sel.funcs
	if fam.id == "clip" { // the clipboard is the input; typed words filter the variants
		in.filter = append(in.filter, strings.Fields(strings.ToLower(in.text))...)
		in.text = ""
	}
	// A bare keyword (no space typed yet) never reads the clipboard or shows a
	// hint: it may just be the start of an ordinary query like "date" or "age".
	if in.text == "" && !fam.noClip && !bare {
		if s, ok := clipText(); ok && strings.TrimSpace(s) != "" {
			in.text, in.clip = s, true
		}
	}
	if len(sel.funcs) == 0 {
		return nil
	}
	// Scope lists put the variants of the typed keyword first.
	if all && sel.prefer != "" {
		sorted := append([]tools.Func(nil), funcs...)
		sort.SliceStable(sorted, func(i, j int) bool { return matchesKw(sorted[i], sel.prefer) && !matchesKw(sorted[j], sel.prefer) })
		funcs = sorted
	}
	limit := toolsRootRows
	if all {
		limit = toolsScopeRows
	}
	sc := p.scopeFor(fam, in, sel.prefer)
	var out []core.Item
	missing := map[string]bool{}
	for _, f := range funcs {
		if ctx.Err() != nil || len(out) >= limit {
			break
		}
		if !all && slowRe.MatchString(f.Name) {
			continue
		}
		if f.InputParam() == "" && !fam.noClip {
			continue // listings (extract.kinds) are API-only outside the generator families
		}
		if len(in.filter) > 0 && !filterMatch(f, in.filter) {
			continue
		}
		if f.Accepts != nil && f.InputParam() != "" && !f.Accepts(in.text) {
			continue
		}
		if f.InputParam() != "" && in.text == "" && len(in.args) == 0 && inputRequired(f) {
			continue // nothing to work on
		}
		fin := in
		if in.clip && !inputRequired(f) {
			fin.text, fin.clip = "", false // an optional input (now) must not swallow the clipboard
		}
		res, miss, err := runVariant(f, fin)
		for _, m := range miss {
			missing[m] = true
		}
		if err != nil || res == nil {
			continue
		}
		if b, isBool := res.(bool); isBool && !b && !all {
			continue // "false" answers to is_* questions are noise in the live list
		}
		out = append(out, p.rowsFor(f, res, fam, fin, sc, sel.prefer)...)
	}
	out = dedupeRows(out)
	if len(out) == 0 {
		// Explain what to type only after "<keyword> " with nothing else; text
		// that merely starts with a keyword ("date and time settings") stays quiet.
		if bare || (!all && (!sel.viaFamily || in.text != "" || len(in.args) > 0)) {
			return nil
		}
		return []core.Item{p.hintRow(fam, in, missing, sc)}
	}
	if len(out) > limit {
		out = out[:limit]
	}
	return out
}

// inputRequired reports whether f cannot run without its input text; an
// optional input (dt.now_all) runs on empty input.
func inputRequired(f tools.Func) bool {
	in := f.InputParam()
	for _, p := range f.Params {
		if p.Name == in {
			return p.Required
		}
	}
	return true
}

// dedupeRows drops rows that show the same value, so a result produced by two
// variants (a dict entry and the function of the same name) appears once.
func dedupeRows(rows []core.Item) []core.Item {
	seen := map[string]bool{}
	out := rows[:0]
	for _, r := range rows {
		k := r.Title + "|" + r.Action.Target
		if seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, r)
	}
	return out
}

func matchesKw(f tools.Func, kw string) bool {
	if strings.EqualFold(f.Name, kw) {
		return true
	}
	for _, k := range f.Keywords {
		if strings.EqualFold(k, kw) {
			return true
		}
	}
	return false
}

func filterMatch(f tools.Func, words []string) bool {
	hay := strings.ToLower(f.Name + " " + f.Title + " " + strings.Join(f.Tags, " "))
	for _, w := range words {
		if !strings.Contains(hay, w) {
			return false
		}
	}
	return true
}

// hintRow explains what to type when nothing could be produced.
func (p *toolsProvider) hintRow(fam *toolFamily, in toolInput, missing map[string]bool, sc *core.Scope) core.Item {
	sub := fam.placeholder
	if len(missing) > 0 {
		var names []string
		for m := range missing {
			names = append(names, m+"=...")
		}
		sort.Strings(names)
		if len(names) > 4 {
			names = names[:4]
		}
		sub = "add " + strings.Join(names, " ") + " (or a :: b for several texts)"
	}
	title := fam.title + " tools"
	if in.text == "" {
		title += ": type or copy something"
	} else {
		title += ": nothing applies to this input"
	}
	return core.Item{ID: "tools:hint:" + fam.id, Title: title, Subtitle: sub, Category: fam.title,
		Action: core.Action{Kind: "none"}, Scope: sc}
}

// ---- Scope ----

// QueryScope lists every variant of the family for the typed text.
func (p *toolsProvider) QueryScope(ctx context.Context, s core.Scope, q string) []core.Item {
	ix := p.index()
	fam, ok := ix.families[s.Data["family"]]
	if !ok {
		return nil
	}
	return p.run(ctx, p.familySel(fam, s.Data["prefer"]), q, true, false)
}

// ---- auto detection (no keyword) ----

var (
	ipRe    = regexp.MustCompile(`^(\d{1,3}\.){3}\d{1,3}(/\d{1,2})?$`)
	ip6Re   = regexp.MustCompile(`(?i)^[0-9a-f:]*:[0-9a-f:]*(/\d{1,3})?$`)
	hexLit  = regexp.MustCompile(`(?i)^[-~(]*0[xbo][0-9a-f_]+[\s0-9a-fxbo_+\-*/%&|^<>~()]*$`)
	uuidRe  = regexp.MustCompile(`(?i)^[{(]?[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}[})]?$`)
	epochRe = regexp.MustCompile(`^\d{10}(\d{3})?$`)
	jwtRe   = regexp.MustCompile(`^eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]*$`)
)

// autoDetect shows a calculator-style answer for input that is clearly one
// thing: an IP or CIDR, a hex/binary expression, a UUID, a unix timestamp or
// a JWT. Everything else needs a keyword, so ordinary typing stays quiet.
func (p *toolsProvider) autoDetect(ctx context.Context, q string) []core.Item {
	q = strings.TrimSpace(q)
	var picks [][2]string
	switch {
	case ipRe.MatchString(q) || (strings.Count(q, ":") >= 2 && ip6Re.MatchString(q)):
		picks = [][2]string{{"net", "cidr_info"}, {"net", "ip_formats"}}
	case hexLit.MatchString(q):
		picks = [][2]string{{"num", "calc_all"}}
	case uuidRe.MatchString(q):
		picks = [][2]string{{"uuid", "uuid_parse"}}
	case epochRe.MatchString(q):
		picks = [][2]string{{"dt", "convert"}}
	case jwtRe.MatchString(q):
		picks = [][2]string{{"encode", "jwt_decode"}}
	default:
		return nil
	}
	ix := p.index()
	var out []core.Item
	for _, pk := range picks {
		f, ok := tools.Lookup(pk[0], pk[1])
		if !ok || f.Family == "" {
			continue
		}
		fam := ix.families[f.Family]
		in := toolInput{raw: q, text: q, params: map[string]string{}}
		res, _, err := runVariant(f, in)
		if err != nil || res == nil {
			continue
		}
		out = append(out, p.rowsFor(f, res, fam, in, p.scopeFor(fam, in, ""), "")...)
		if ctx.Err() != nil {
			break
		}
	}
	if len(out) > toolsRootRows {
		out = out[:toolsRootRows]
	}
	return out
}

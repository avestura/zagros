package ext

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/avestura/zagros/internal/core"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

// KindExtCall is the core.Action kind for actions that call back into an
// extension's on_execute hook.
const KindExtCall = "ext-call"

func mkAction(kind, target string, args []string, name string, data starlark.Value) starlark.Value {
	d := starlark.StringDict{
		"kind":   starlark.String(kind),
		"target": starlark.String(target),
		"args":   goToStarlark(args),
		"name":   starlark.String(name),
		"data":   data,
	}
	if data == nil {
		d["data"] = starlark.None
	}
	return starlarkstruct.FromStringDict(starlark.String("action"), d)
}

func biActionOpen(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var t string
	if err := starlark.UnpackArgs("action_open", args, kw, "target", &t); err != nil {
		return nil, err
	}
	return mkAction("open", t, nil, "", nil), nil
}

func biActionCopy(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var t string
	if err := starlark.UnpackArgs("action_copy", args, kw, "text", &t); err != nil {
		return nil, err
	}
	return mkAction("copy", t, nil, "", nil), nil
}

func biActionRun(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var cmd string
	var a starlark.Value = starlark.None
	if err := starlark.UnpackArgs("action_run", args, kw, "cmd", &cmd, "args?", &a); err != nil {
		return nil, err
	}
	list, err := stringList(a)
	if err != nil {
		return nil, err
	}
	return mkAction("run", cmd, list, "", nil), nil
}

func biActionCall(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var name string
	var data starlark.Value = starlark.None
	if err := starlark.UnpackArgs("action_call", args, kw, "name", &name, "data?", &data); err != nil {
		return nil, err
	}
	return mkAction("call", "", nil, name, data), nil
}

func biItem(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var id, title, sub, icon string
	var action, alt starlark.Value = starlark.None, starlark.None
	score := 0.0
	if err := starlark.UnpackArgs("item", args, kw, "id", &id, "title", &title, "subtitle?", &sub, "icon?", &icon, "action?", &action, "alt?", &alt, "score?", &score); err != nil {
		return nil, err
	}
	if id == "" || title == "" {
		return nil, fmt.Errorf("item: id and title must not be empty")
	}
	return starlarkstruct.FromStringDict(starlark.String("item"), starlark.StringDict{
		"id": starlark.String(id), "title": starlark.String(title), "subtitle": starlark.String(sub),
		"icon": starlark.String(icon), "action": action, "alt": alt, "score": starlark.Float(score),
	}), nil
}

// field reads a named field from a struct or dict.
func field(v starlark.Value, name string) (starlark.Value, bool) {
	switch x := v.(type) {
	case *starlarkstruct.Struct:
		f, err := x.Attr(name)
		if err != nil || f == nil {
			return nil, false
		}
		return f, true
	case *starlark.Dict:
		f, ok, _ := x.Get(starlark.String(name))
		return f, ok
	}
	return nil, false
}

func strField(v starlark.Value, name string) string {
	f, ok := field(v, name)
	if !ok {
		return ""
	}
	s, _ := starlark.AsString(f)
	return s
}

type actionSpec struct {
	kind, target, name string
	args               []string
	data               starlark.Value
}

func parseAction(v starlark.Value) (actionSpec, error) {
	if v == nil || v == starlark.None {
		return actionSpec{kind: "call", name: "default"}, nil
	}
	if _, ok := field(v, "kind"); !ok {
		return actionSpec{}, fmt.Errorf("action must come from action_open/action_copy/action_run/action_call, got %s", v.Type())
	}
	a := actionSpec{kind: strField(v, "kind"), target: strField(v, "target"), name: strField(v, "name")}
	if f, ok := field(v, "args"); ok {
		l, err := stringList(f)
		if err != nil {
			return a, err
		}
		a.args = l
	}
	if f, ok := field(v, "data"); ok {
		a.data = f
	}
	switch a.kind {
	case "open", "copy", "run", "call", "none":
	default:
		return a, fmt.Errorf("unknown action kind %q", a.kind)
	}
	return a, nil
}

// toCore converts an action, enforcing the permission that its effect needs.
func (e *Extension) toCore(th *starlark.Thread, a actionSpec, itemID, title string) (core.Action, error) {
	switch a.kind {
	case "open":
		if err := e.require(PermOpen, ""); err != nil {
			return core.Action{Kind: "none"}, err
		}
		return core.Action{Kind: "open", Target: a.target}, nil
	case "copy":
		if err := e.require(PermClipboardWrite, ""); err != nil {
			return core.Action{Kind: "none"}, err
		}
		return core.Action{Kind: "copy", Target: a.target}, nil
	case "run":
		if err := e.require(PermProcessRun, a.target); err != nil {
			return core.Action{Kind: "none"}, err
		}
		return core.Action{Kind: "run", Target: a.target, Args: a.args}, nil
	case "call":
		data := ""
		if a.data != nil && a.data != starlark.None {
			enc, err := jsonCall(th, "encode", a.data)
			if err != nil {
				return core.Action{Kind: "none"}, err
			}
			data = string(enc.(starlark.String))
		}
		return core.Action{Kind: KindExtCall, Extra: map[string]string{
			"ext": e.Manifest.ID, "name": a.name, "data": data, "item": itemID, "title": title,
		}}, nil
	}
	return core.Action{Kind: "none"}, nil
}

// Query runs on_query and converts the returned items. Items whose action
// needs a permission that is not granted keep their row but get a no-op action.
func (e *Extension) Query(ctx context.Context, q string) ([]core.Item, error) {
	if !e.Enabled() || e.runtime() == nil {
		return nil, nil
	}
	rt := e.runtime()
	res, err := e.callHook(ctx, "on_query", starlark.String(q))
	if err != nil {
		if err == errNoHook {
			return nil, nil
		}
		return nil, err
	}
	if res == starlark.None {
		return nil, nil
	}
	it, ok := res.(starlark.Iterable)
	if !ok {
		return nil, fmt.Errorf("on_query must return a list of item(), got %s", res.Type())
	}
	th := e.newThread(ctx, "convert", nil)
	_ = rt
	var out []core.Item
	iter := it.Iterate()
	defer iter.Done()
	var v starlark.Value
	for iter.Next(&v) {
		ci, err := e.convertItem(th, v)
		if err != nil {
			return out, err
		}
		out = append(out, ci)
		if len(out) >= 200 {
			break
		}
	}
	for i := range out {
		if out[i].Score == 0 {
			out[i].Score = 1 - float64(i)/float64(len(out)+1)
		}
	}
	return out, nil
}

func (e *Extension) convertItem(th *starlark.Thread, v starlark.Value) (core.Item, error) {
	id, title := strField(v, "id"), strField(v, "title")
	if id == "" || title == "" {
		return core.Item{}, fmt.Errorf("on_query returned %s without id/title; use item(...)", v.Type())
	}
	ci := core.Item{
		ID:       "ext:" + e.Manifest.ID + ":" + id,
		Title:    title,
		Subtitle: strField(v, "subtitle"),
		Icon:     strField(v, "icon"),
		Category: e.Manifest.Name,
	}
	if f, ok := field(v, "score"); ok {
		if x, ok := f.(starlark.Float); ok {
			ci.Score = float64(x)
		}
	}
	af, _ := field(v, "action")
	spec, err := parseAction(af)
	if err != nil {
		return ci, fmt.Errorf("item %q: %w", id, err)
	}
	ci.Action = e.safeAction(th, spec, ci.ID, title)
	if altv, ok := field(v, "alt"); ok && altv != starlark.None {
		it, ok := altv.(starlark.Iterable)
		if !ok {
			return ci, fmt.Errorf("item %q: alt must be a list", id)
		}
		iter := it.Iterate()
		defer iter.Done()
		var a starlark.Value
		for iter.Next(&a) {
			var name string
			var act starlark.Value
			if t, ok := a.(starlark.Tuple); ok && len(t) == 2 {
				name, _ = starlark.AsString(t[0])
				act = t[1]
			} else {
				name = strField(a, "name")
				act, _ = field(a, "action")
			}
			sp, err := parseAction(act)
			if err != nil || name == "" {
				return ci, fmt.Errorf("item %q: alt entries need a name and an action", id)
			}
			ci.Alt = append(ci.Alt, core.NamedAction{Name: name, Action: e.safeAction(th, sp, ci.ID, title)})
		}
	}
	return ci, nil
}

func (e *Extension) safeAction(th *starlark.Thread, s actionSpec, id, title string) core.Action {
	a, err := e.toCore(th, s, id, title)
	if err != nil {
		e.opts.Log(e.Manifest.ID, "warn", err.Error())
	}
	return a
}

// Execute handles a core.Action of Kind "ext-call" by calling on_execute.
// If on_execute returns an action it is run through Options.Executor.
func (e *Extension) Execute(ctx context.Context, a core.Action) error {
	if a.Kind != KindExtCall {
		return fmt.Errorf("ext: cannot execute action kind %q", a.Kind)
	}
	if !e.Enabled() || e.runtime() == nil {
		return fmt.Errorf("extension %s is not running", e.Manifest.ID)
	}
	th := e.newThread(ctx, "execute-args", nil)
	var data starlark.Value = starlark.None
	if s := a.Extra["data"]; s != "" {
		d, err := jsonCall(th, "decode", starlark.String(s))
		if err != nil {
			return err
		}
		data = d
	}
	itemVal := starlarkstruct.FromStringDict(starlark.String("item"), starlark.StringDict{
		"id":    starlark.String(a.Extra["item"]),
		"title": starlark.String(a.Extra["title"]),
		"name":  starlark.String(a.Extra["name"]),
		"data":  data,
	})
	res, err := e.callHook(ctx, "on_execute", itemVal)
	if err != nil {
		if err == errNoHook {
			return fmt.Errorf("extension %s has no on_execute hook", e.Manifest.ID)
		}
		return err
	}
	if res == nil || res == starlark.None {
		return nil
	}
	spec, err := parseAction(res)
	if err != nil {
		return fmt.Errorf("on_execute result: %w", err)
	}
	if spec.kind == "call" {
		return fmt.Errorf("on_execute cannot return action_call")
	}
	ca, err := e.toCore(th, spec, "", "")
	if err != nil {
		return err
	}
	if ca.Kind == "none" || e.opts.Executor == nil {
		return nil
	}
	return e.opts.Executor.Execute(ca)
}

// ExecData is the decoded form of an ext-call action, for UI/debugging.
func ExecData(a core.Action) (extID, name string, data json.RawMessage) {
	return a.Extra["ext"], a.Extra["name"], json.RawMessage(a.Extra["data"])
}

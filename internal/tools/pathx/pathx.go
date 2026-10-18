// Package pathx registers the "path" tool group: join, split, clean and
// compare file paths in Windows or POSIX flavour. The functions work on text
// only; they never touch the disk and do not depend on the host OS.
package pathx

import (
	"fmt"
	"strings"

	"github.com/avestura/zagros/internal/tools"
)

var flavorChoices = []string{"posix", "windows", "auto"}

func flavorParam() tools.Param {
	return tools.Param{Name: "flavor", Type: "str", Default: "posix", Choices: flavorChoices,
		Doc: "Path rules: posix (/ separator), windows (\\ or /, drive letters and UNC shares) or auto (windows when the path has a backslash or a drive letter)."}
}

func init() {
	one := func(name, title, doc, example string, kw []string, ret string, fn func(f flavor, p string) (any, error)) {
		tools.Register(tools.Func{
			Group: "path", Name: name, Title: title,
			Doc:     doc,
			Terms:   kw,
			Params:  []tools.Param{tools.Req("path", "str", "The path."), flavorParam()},
			Returns: ret,
			Example: example,
			Tags:    append([]string{"path", "filename", "file"}, name),
			Fn: func(a tools.Args) (any, error) {
				f, err := pickFlavor(a.Str("flavor"), a.Str("path"))
				if err != nil {
					return nil, err
				}
				return fn(f, a.Str("path"))
			},
		})
	}

	tools.Register(tools.Func{
		Group: "path", Name: "join", Title: "Join path parts",
		Doc:     "Join path parts with the flavor's separator and clean the result (dot segments resolved, repeated separators merged). Empty parts are skipped and an absolute part does not discard the ones before it. Returns an empty string when every part is empty.",
		Terms:   []string{"combine paths", "append path", "path.join", "build path", "concatenate path"},
		Params:  []tools.Param{tools.Req("parts", "list[str]", "The parts to join."), flavorParam()},
		Returns: "str",
		Example: `path.join(["C:\\Users", "me", "..", "docs", "a.txt"], flavor="windows")`,
		Tags:    []string{"path", "join", "combine", "append"},
		Fn: func(a tools.Args) (any, error) {
			parts := a.List("parts")
			f, err := pickFlavor(a.Str("flavor"), strings.Join(parts, ""))
			if err != nil {
				return nil, err
			}
			var keep []string
			for _, p := range parts {
				if p != "" {
					keep = append(keep, p)
				}
			}
			if len(keep) == 0 {
				return "", nil
			}
			return f.clean(strings.Join(keep, string(f.sep))), nil
		},
	})
	one("clean", "Clean path", "Return the shortest equivalent path: repeated separators merged, . removed, .. resolved against the part before it (dropped at the root, kept at the start of a relative path). An empty path becomes a dot.",
		`path.clean("/a/./b/../c//d/")`, []string{"normalize path", "resolve dots", "remove ..", "simplify path", "canonical path"}, "str",
		func(f flavor, p string) (any, error) { return f.clean(p), nil })
	one("base", "File name of a path", "Return the last element: the file or folder name. Trailing separators are ignored; an empty path gives a dot and a root gives the separator.",
		`path.base("/usr/local/bin/tool.exe")`, []string{"filename", "basename", "file name from path", "last segment"}, "str",
		func(f flavor, p string) (any, error) { return f.base(p), nil })
	one("dir", "Folder of a path", "Return everything but the last element, cleaned: the containing folder. A bare name gives a dot.",
		`path.dir("C:\\Users\\me\\a.txt", flavor="windows")`, []string{"dirname", "parent folder", "parent directory", "containing folder"}, "str",
		func(f flavor, p string) (any, error) { return f.dir(p), nil })
	one("ext", "File extension", "Return the extension of the last element including the dot, for example .gz for a.tar.gz. Files that only start with a dot (.bashrc) and names that end in a dot have none.",
		`path.ext("archive.tar.gz")`, []string{"file extension", "suffix", "file type", "extname"}, "str",
		func(f flavor, p string) (any, error) { b := f.base(p); return b[len(b)-len(extOf(b)):], nil })
	one("stem", "File name without extension", "Return the last element without its extension (see path.ext): archive.tar.gz gives archive.tar.",
		`path.stem("/tmp/report.final.pdf")`, []string{"filename without extension", "name without ext", "strip extension"}, "str",
		func(f flavor, p string) (any, error) { b := f.base(p); return b[:len(b)-len(extOf(b))], nil })
	one("split", "Split into folder and file", "Split after the last separator into [folder, file]. The folder keeps its trailing separator, so joining the two gives the original path back. A path with no separator gives an empty folder.",
		`path.split("/a/b/c.txt")`, []string{"dirname and basename", "separate filename", "split path"}, "list[str]",
		func(f flavor, p string) (any, error) { d, b := f.split(p); return []string{d, b}, nil })
	one("parts", "Path elements", "Return the elements of the path as a list. The root or drive comes first as its own element (/, C:\\, \\\\host\\share\\); empty and . elements are dropped and .. is kept.",
		`path.parts("C:\\a\\b\\c.txt", flavor="windows")`, []string{"split path into segments", "path components", "directories list", "path segments"}, "list[str]",
		func(f flavor, p string) (any, error) { return f.parts(p), nil })
	one("is_abs", "Is absolute path", "Report whether the path is absolute: a leading / for posix; a drive with a separator (C:\\x) or a UNC share for windows. \\x and C:x are not absolute on windows.",
		`path.is_abs("C:\\Windows", flavor="windows")`, []string{"absolute or relative", "is relative", "check absolute path", "rooted"}, "bool",
		func(f flavor, p string) (any, error) { return f.isAbs(p), nil })

	tools.Register(tools.Func{
		Group: "path", Name: "rel", Title: "Relative path between two paths",
		Doc:     "Return the path that leads from base to target, for example ../c from /a/b to /a/c. Both paths are cleaned first and must both be absolute or both relative; on windows they must share a drive, and names compare case-insensitively. Raises when no relative path exists.",
		Terms:   []string{"relative path", "relpath", "path from to", "make relative", "diff paths"},
		Params:  []tools.Param{tools.Req("base", "str", "The starting folder."), tools.Req("target", "str", "The path to reach."), flavorParam()},
		Returns: "str",
		Example: `path.rel("/data/app", "/data/lib/x.so")`,
		Tags:    []string{"path", "relative", "relpath"},
		Fn: func(a tools.Args) (any, error) {
			f, err := pickFlavor(a.Str("flavor"), a.Str("base")+a.Str("target"))
			if err != nil {
				return nil, err
			}
			return f.rel(a.Str("base"), a.Str("target"))
		},
	})
	tools.Register(tools.Func{
		Group: "path", Name: "is_within", Title: "Is path inside a folder",
		Doc:     "Report whether child is the same as or inside parent after both are cleaned, without touching the disk. Useful to reject .. escapes before using a user-supplied name. Both must be absolute or both relative; names compare case-insensitively on windows.",
		Terms:   []string{"path traversal", "directory traversal check", "is subpath", "inside folder", "sandbox check", "starts with"},
		Params:  []tools.Param{tools.Req("parent", "str", "The folder."), tools.Req("child", "str", "The path to test."), flavorParam()},
		Returns: "bool",
		Example: `path.is_within("/srv/data", "/srv/data/../etc/passwd")`,
		Tags:    []string{"path", "traversal", "within", "subpath", "security"},
		Fn: func(a tools.Args) (any, error) {
			f, err := pickFlavor(a.Str("flavor"), a.Str("parent")+a.Str("child"))
			if err != nil {
				return nil, err
			}
			r, err := f.rel(a.Str("parent"), a.Str("child"))
			if err != nil {
				return false, nil
			}
			return r != ".." && !strings.HasPrefix(r, ".."+string(f.sep)), nil
		},
	})
	tools.Register(tools.Func{
		Group: "path", Name: "with_ext", Title: "Change file extension",
		Doc:     "Replace the extension of the last element. ext may be given with or without the dot; an empty ext removes the extension. Raises when ext contains a separator.",
		Terms:   []string{"rename extension", "change extension", "replace suffix", "swap extension"},
		Params:  []tools.Param{tools.Req("path", "str", "The path."), tools.Req("ext", "str", "The new extension, for example .txt or txt."), flavorParam()},
		Returns: "str",
		Example: `path.with_ext("/tmp/a.txt", "md")`,
		Tags:    []string{"path", "extension", "rename", "suffix"},
		Fn: func(a tools.Args) (any, error) {
			f, err := pickFlavor(a.Str("flavor"), a.Str("path"))
			if err != nil {
				return nil, err
			}
			ext := a.Str("ext")
			if ext != "" && !strings.HasPrefix(ext, ".") {
				ext = "." + ext
			}
			if strings.ContainsAny(ext, "/\\") || ext == "." {
				return nil, fmt.Errorf("invalid extension %q", a.Str("ext"))
			}
			p := a.Str("path")
			b := f.base(p)
			if b == "." || b == string(f.sep) || b == ".." {
				return nil, fmt.Errorf("path %q has no file name", p)
			}
			trimmed := strings.TrimRight(p, "/\\")
			if f.win {
				return trimmed[:len(trimmed)-len(extOf(b))] + ext, nil
			}
			trimmed = strings.TrimRight(p, "/")
			return trimmed[:len(trimmed)-len(extOf(b))] + ext, nil
		},
	})
	tools.Register(tools.Func{
		Group: "path", Name: "convert", Title: "Convert path separators",
		Doc:     "Rewrite the separators of a path for the target flavor: / to \\ for windows, \\ to / for posix. Nothing else changes, so drive letters stay. Use path.clean afterwards to normalize.",
		Terms:   []string{"windows to unix path", "unix to windows path", "backslash to slash", "slash to backslash", "wsl path", "convert separators"},
		Params:  []tools.Param{tools.Req("path", "str", "The path."), {Name: "to", Type: "str", Default: "posix", Choices: []string{"posix", "windows"}, Doc: "The flavor to convert to."}},
		Returns: "str",
		Example: `path.convert("C:\\Users\\me\\a.txt", to="posix")`,
		Tags:    []string{"path", "convert", "separator", "slash", "backslash"},
		Fn: func(a tools.Args) (any, error) {
			switch a.Str("to") {
			case "posix":
				return strings.ReplaceAll(a.Str("path"), "\\", "/"), nil
			case "windows":
				return strings.ReplaceAll(a.Str("path"), "/", "\\"), nil
			}
			return nil, fmt.Errorf("to must be posix or windows, got %q", a.Str("to"))
		},
	})
}

type flavor struct {
	win bool
	sep byte
}

func pickFlavor(name, sample string) (flavor, error) {
	switch name {
	case "", "posix":
		return flavor{false, '/'}, nil
	case "windows":
		return flavor{true, '\\'}, nil
	case "auto":
		if strings.Contains(sample, "\\") || (len(sample) >= 2 && sample[1] == ':' && isLetter(sample[0])) {
			return flavor{true, '\\'}, nil
		}
		return flavor{false, '/'}, nil
	}
	return flavor{}, fmt.Errorf("flavor must be posix, windows or auto, got %q", name)
}

func isLetter(c byte) bool { return c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' }

func (f flavor) isSep(c byte) bool { return c == '/' || (f.win && c == '\\') }

// volume returns the length of the drive or UNC prefix.
func (f flavor) volume(p string) int {
	if !f.win {
		return 0
	}
	if len(p) >= 2 && p[1] == ':' && isLetter(p[0]) {
		return 2
	}
	if len(p) >= 5 && f.isSep(p[0]) && f.isSep(p[1]) && !f.isSep(p[2]) && p[2] != '.' {
		// \\host\share
		n := 2
		for n < len(p) && !f.isSep(p[n]) {
			n++
		}
		if n < len(p) {
			m := n + 1
			for m < len(p) && !f.isSep(p[m]) {
				m++
			}
			if m > n+1 {
				return m
			}
		}
	}
	return 0
}

func (f flavor) split3(p string) (vol string, rooted bool, elems []string) {
	n := f.volume(p)
	vol, rest := p[:n], p[n:]
	vol = strings.ReplaceAll(vol, "/", "\\")
	rooted = len(rest) > 0 && f.isSep(rest[0])
	if len(vol) > 2 {
		rooted = true // a UNC share is always rooted
	}
	for _, e := range strings.FieldsFunc(rest, func(r rune) bool { return r < 0x80 && f.isSep(byte(r)) }) {
		elems = append(elems, e)
	}
	return
}

func (f flavor) clean(p string) string {
	if p == "" {
		return "."
	}
	vol, rooted, elems := f.split3(p)
	var out []string
	for _, e := range elems {
		switch e {
		case ".":
		case "..":
			switch {
			case len(out) > 0 && out[len(out)-1] != "..":
				out = out[:len(out)-1]
			case !rooted:
				out = append(out, "..")
			}
		default:
			out = append(out, e)
		}
	}
	s := string(f.sep)
	body := strings.Join(out, s)
	switch {
	case rooted && len(vol) > 2 && body == "":
		return vol + s
	case rooted:
		return vol + s + body
	case body == "":
		return vol + "."
	}
	return vol + body
}

func (f flavor) split(p string) (dir, file string) {
	n := f.volume(p)
	i := len(p) - 1
	for i >= n && !f.isSep(p[i]) {
		i--
	}
	return p[:i+1], p[i+1:]
}

func (f flavor) base(p string) string {
	if p == "" {
		return "."
	}
	for len(p) > 0 && f.isSep(p[len(p)-1]) {
		p = p[:len(p)-1]
	}
	n := f.volume(p)
	p = p[n:]
	for i := len(p) - 1; i >= 0; i-- {
		if f.isSep(p[i]) {
			p = p[i+1:]
			break
		}
	}
	if p == "" {
		return string(f.sep)
	}
	return p
}

func (f flavor) dir(p string) string {
	n := f.volume(p)
	i := len(p) - 1
	for i >= n && !f.isSep(p[i]) {
		i--
	}
	d := f.clean(p[n : i+1])
	if d == "." && n > 2 {
		return p[:n]
	}
	return strings.ReplaceAll(p[:n], "/", "\\") + d
}

// extOf returns the extension of a bare file name.
func extOf(b string) string {
	i := strings.LastIndexByte(b, '.')
	if i <= 0 || i == len(b)-1 {
		return ""
	}
	return b[i:]
}

func (f flavor) parts(p string) []string {
	vol, rooted, elems := f.split3(p)
	out := []string{}
	if rooted || vol != "" {
		root := vol
		if rooted {
			root += string(f.sep)
		}
		out = append(out, root)
	}
	for _, e := range elems {
		if e != "." {
			out = append(out, e)
		}
	}
	return out
}

func (f flavor) isAbs(p string) bool {
	vol, rooted, _ := f.split3(p)
	if f.win {
		return rooted && vol != ""
	}
	return rooted
}

func (f flavor) eq(a, b string) bool {
	if f.win {
		return strings.EqualFold(a, b)
	}
	return a == b
}

func (f flavor) rel(base, target string) (string, error) {
	bv, br, be := f.split3(f.clean(base))
	tv, tr, te := f.split3(f.clean(target))
	if !f.eq(bv, tv) {
		return "", fmt.Errorf("cannot make %q relative to %q: different drives", target, base)
	}
	if br != tr {
		return "", fmt.Errorf("cannot make %q relative to %q: one is absolute and the other is not", target, base)
	}
	norm := func(e []string) []string {
		if len(e) == 1 && e[0] == "." {
			return nil
		}
		return e
	}
	be, te = norm(be), norm(te)
	i := 0
	for i < len(be) && i < len(te) && f.eq(be[i], te[i]) {
		i++
	}
	for _, e := range be[i:] {
		if e == ".." {
			return "", fmt.Errorf("cannot make %q relative to %q: base goes above its start", target, base)
		}
	}
	var out []string
	for range be[i:] {
		out = append(out, "..")
	}
	out = append(out, te[i:]...)
	if len(out) == 0 {
		return ".", nil
	}
	return strings.Join(out, string(f.sep)), nil
}

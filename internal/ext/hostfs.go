package ext

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func (e *Extension) fsModule() *starlarkstruct.Module {
	return module("fs", starlark.StringDict{
		"read_text":   e.fn("fs.read_text", e.biFSRead),
		"read_bytes":  e.fn("fs.read_bytes", e.biFSReadBytes),
		"write_text":  e.fn("fs.write_text", e.biFSWrite),
		"write_bytes": e.fn("fs.write_bytes", e.biFSWriteBytes),
		"append_text": e.fn("fs.append_text", e.biFSAppend),
		"list":        e.fn("fs.list", e.biFSList),
		"exists":      e.fn("fs.exists", e.biFSExists),
		"stat":        e.fn("fs.stat", e.biFSStat),
		"mkdir":       e.fn("fs.mkdir", e.biFSMkdir),
		"remove":      e.fn("fs.remove", e.biFSRemove),
		"rename":      e.fn("fs.rename", e.biFSRename),
		"copy":        e.fn("fs.copy", e.biFSCopy),
		"glob":        e.fn("fs.glob", e.biFSGlob),
		"walk":        e.fn("fs.walk", e.biFSWalk),
	})
}

// lexicalPath expands ~ and makes p absolute without touching the disk; it is
// the spelling returned to scripts, while I/O always uses the resolved path.
func lexicalPath(p string) (string, error) {
	abs, err := filepath.Abs(expandHome(p))
	if err != nil {
		return "", err
	}
	return filepath.Clean(abs), nil
}

// checkFSNoFollow is checkFS for operations on the entry itself (remove,
// rename): the parent folder is resolved through links but the last element
// is not, so a link can be removed or renamed without touching its target.
func (e *Extension) checkFSNoFollow(kind, p string) (string, error) {
	abs, err := lexicalPath(p)
	if err != nil {
		return "", err
	}
	dir := filepath.Dir(abs)
	if dir == abs {
		return "", fmt.Errorf("%s is a filesystem root", p)
	}
	parent, err := resolvePath(dir)
	if err != nil {
		return "", err
	}
	r := filepath.Join(parent, filepath.Base(abs))
	if kind == PermFSRead {
		if own, _ := resolvePath(e.Dir); pathWithin(own, r) {
			return r, nil
		}
	}
	return r, e.require(kind, r)
}

// protectedRoot reports paths that bulk operations must never remove or
// replace: volume roots, the home folder and the roots of granted fs.write scopes.
func (e *Extension) protectedRoot(r string) bool {
	if filepath.Dir(r) == r {
		return true
	}
	if h, err := os.UserHomeDir(); err == nil {
		if hr, err := resolvePath(h); err == nil && samePath(hr, r) {
			return true
		}
	}
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, g := range e.granted {
		if g.Kind() != PermFSWrite || g.Scope() == "" {
			continue
		}
		if sr, err := resolvePath(g.Scope()); err == nil && samePath(sr, r) {
			return true
		}
	}
	return false
}

func (e *Extension) capWrite(n int) error {
	if int64(n) > e.opts.Limits.MaxFileBytes {
		return fmt.Errorf("data larger than %d bytes", e.opts.Limits.MaxFileBytes)
	}
	return nil
}

func (e *Extension) biFSReadBytes(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	if err := starlark.UnpackArgs("fs.read_bytes", args, kw, "path", &p); err != nil {
		return nil, err
	}
	r, err := e.checkFS(PermFSRead, p)
	if err != nil {
		return nil, err
	}
	b, err := e.readFileCapped(r)
	if err != nil {
		return nil, err
	}
	return starlark.Bytes(b), nil
}

func (e *Extension) biFSWriteBytes(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	var data starlark.Value
	if err := starlark.UnpackArgs("fs.write_bytes", args, kw, "path", &p, "data", &data); err != nil {
		return nil, err
	}
	b, err := asBytes(data)
	if err != nil {
		return nil, err
	}
	if err := e.capWrite(len(b)); err != nil {
		return nil, err
	}
	r, err := e.checkFS(PermFSWrite, p)
	if err != nil {
		return nil, err
	}
	return starlark.None, os.WriteFile(r, b, 0o644)
}

func (e *Extension) biFSAppend(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	var data starlark.Value
	if err := starlark.UnpackArgs("fs.append_text", args, kw, "path", &p, "text", &data); err != nil {
		return nil, err
	}
	b, err := asBytes(data)
	if err != nil {
		return nil, err
	}
	if err := e.capWrite(len(b)); err != nil {
		return nil, err
	}
	r, err := e.checkFS(PermFSWrite, p)
	if err != nil {
		return nil, err
	}
	f, err := os.OpenFile(r, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	_, err = f.Write(b)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	return starlark.None, err
}

func entryValue(name, display, rel string, info fs.FileInfo) starlark.Value {
	d := starlark.StringDict{
		"name":       starlark.String(name),
		"path":       starlark.String(filepath.ToSlash(display)),
		"size":       starlark.MakeInt64(info.Size()),
		"mtime":      starlark.MakeInt64(info.ModTime().Unix()),
		"is_dir":     starlark.Bool(info.IsDir()),
		"is_file":    starlark.Bool(info.Mode().IsRegular()),
		"is_symlink": starlark.Bool(info.Mode()&fs.ModeSymlink != 0),
	}
	if rel != "" {
		d["rel"] = starlark.String(rel)
	}
	return starlarkstruct.FromStringDict(starlark.String("entry"), d)
}

func (e *Extension) biFSStat(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	if err := starlark.UnpackArgs("fs.stat", args, kw, "path", &p); err != nil {
		return nil, err
	}
	r, err := e.checkFS(PermFSRead, p)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(r)
	if err != nil {
		return nil, err
	}
	disp, _ := lexicalPath(p)
	return entryValue(info.Name(), disp, "", info), nil
}

func (e *Extension) biFSMkdir(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	parents := true
	if err := starlark.UnpackArgs("fs.mkdir", args, kw, "path", &p, "parents?", &parents); err != nil {
		return nil, err
	}
	r, err := e.checkFS(PermFSWrite, p)
	if err != nil {
		return nil, err
	}
	if parents {
		return starlark.None, os.MkdirAll(r, 0o755)
	}
	return starlark.None, os.Mkdir(r, 0o755)
}

func (e *Extension) biFSRemove(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	var recursive, missingOK bool
	if err := starlark.UnpackArgs("fs.remove", args, kw, "path", &p, "recursive?", &recursive, "missing_ok?", &missingOK); err != nil {
		return nil, err
	}
	r, err := e.checkFSNoFollow(PermFSWrite, p)
	if err != nil {
		return nil, err
	}
	if e.protectedRoot(r) {
		return nil, fmt.Errorf("refusing to remove %s (a granted folder, home or drive root)", p)
	}
	info, err := os.Lstat(r)
	if errors.Is(err, fs.ErrNotExist) && missingOK {
		return starlark.False, nil
	}
	if err != nil {
		return nil, err
	}
	if info.IsDir() && recursive {
		err = os.RemoveAll(r)
	} else {
		err = os.Remove(r)
	}
	if err != nil {
		return nil, err
	}
	return starlark.True, nil
}

func (e *Extension) biFSRename(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var src, dst string
	var overwrite bool
	if err := starlark.UnpackArgs("fs.rename", args, kw, "src", &src, "dst", &dst, "overwrite?", &overwrite); err != nil {
		return nil, err
	}
	s, err := e.checkFSNoFollow(PermFSWrite, src)
	if err != nil {
		return nil, err
	}
	d, err := e.checkFSNoFollow(PermFSWrite, dst)
	if err != nil {
		return nil, err
	}
	if e.protectedRoot(s) || e.protectedRoot(d) {
		return nil, errors.New("refusing to move a granted folder, home or drive root")
	}
	if _, err := os.Lstat(s); err != nil {
		return nil, err
	}
	if _, err := os.Lstat(d); err == nil && !overwrite && !samePath(s, d) {
		return nil, fmt.Errorf("%s already exists (pass overwrite = True)", dst)
	}
	return starlark.None, os.Rename(s, d)
}

type copyBudget struct {
	files int
	bytes int64
}

func (e *Extension) biFSCopy(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var src, dst string
	var overwrite bool
	if err := starlark.UnpackArgs("fs.copy", args, kw, "src", &src, "dst", &dst, "overwrite?", &overwrite); err != nil {
		return nil, err
	}
	s, err := e.checkFS(PermFSRead, src)
	if err != nil {
		return nil, err
	}
	d, err := e.checkFSNoFollow(PermFSWrite, dst)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(s)
	if err != nil {
		return nil, err
	}
	if info.IsDir() && pathWithin(s, d) {
		return nil, errors.New("cannot copy a folder into itself")
	}
	if e.protectedRoot(d) {
		return nil, errors.New("refusing to replace a granted folder, home or drive root")
	}
	var bud copyBudget
	if err := e.copyTree(s, d, info, overwrite, &bud); err != nil {
		return nil, err
	}
	return starlark.MakeInt64(bud.bytes), nil
}

func (e *Extension) copyTree(src, dst string, info fs.FileInfo, overwrite bool, bud *copyBudget) error {
	bud.files++
	if bud.files > e.opts.Limits.MaxEntries {
		return fmt.Errorf("more than %d files to copy", e.opts.Limits.MaxEntries)
	}
	if info.IsDir() {
		if err := os.MkdirAll(dst, 0o755); err != nil {
			return err
		}
		ents, err := os.ReadDir(src)
		if err != nil {
			return err
		}
		for _, d := range ents {
			ci, err := d.Info()
			if err != nil {
				return err
			}
			if ci.Mode()&fs.ModeSymlink != 0 || !(ci.IsDir() || ci.Mode().IsRegular()) {
				continue // links and devices inside a tree are not followed or copied
			}
			if err := e.copyTree(filepath.Join(src, d.Name()), filepath.Join(dst, d.Name()), ci, overwrite, bud); err != nil {
				return err
			}
		}
		return nil
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}
	if !overwrite {
		if _, err := os.Lstat(dst); err == nil {
			return fmt.Errorf("%s already exists (pass overwrite = True)", dst)
		}
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp := dst + ".part"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}
	remain := e.opts.Limits.MaxBulkBytes - bud.bytes
	n, err := io.Copy(out, io.LimitReader(in, remain+1))
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err == nil && n > remain {
		err = fmt.Errorf("copy larger than %d bytes", e.opts.Limits.MaxBulkBytes)
	}
	if err == nil {
		os.Remove(dst)
		err = os.Rename(tmp, dst)
	}
	if err != nil {
		os.Remove(tmp)
		return err
	}
	bud.bytes += n
	return nil
}

const maxScan = 200000

func (e *Extension) biFSWalk(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	maxDepth := -1
	if err := starlark.UnpackArgs("fs.walk", args, kw, "path", &p, "max_depth?", &maxDepth); err != nil {
		return nil, err
	}
	r, err := e.checkFS(PermFSRead, p)
	if err != nil {
		return nil, err
	}
	disp, _ := lexicalPath(p)
	var out []starlark.Value
	scanned := 0
	var rec func(dir, rel string, depth int) error
	rec = func(dir, rel string, depth int) error {
		ents, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, d := range ents {
			scanned++
			if scanned > maxScan {
				return fmt.Errorf("more than %d entries to scan", maxScan)
			}
			info, err := d.Info()
			if err != nil {
				continue
			}
			rl := path.Join(rel, d.Name())
			if len(out) >= e.opts.Limits.MaxEntries {
				return fmt.Errorf("more than %d entries (narrow the folder or set max_depth)", e.opts.Limits.MaxEntries)
			}
			out = append(out, entryValue(d.Name(), filepath.Join(disp, filepath.FromSlash(rl)), rl, info))
			if info.IsDir() && (maxDepth < 0 || depth < maxDepth) {
				if err := rec(filepath.Join(dir, d.Name()), rl, depth+1); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := rec(r, "", 1); err != nil {
		return nil, err
	}
	return starlark.NewList(out), nil
}

func hasMeta(s string) bool { return strings.ContainsAny(s, "*?[") }

func matchSegs(p, n []string) bool {
	for len(p) > 0 {
		if p[0] == "**" {
			for len(p) > 1 && p[1] == "**" {
				p = p[1:]
			}
			if len(p) == 1 {
				return true
			}
			for i := 0; i <= len(n); i++ {
				if matchSegs(p[1:], n[i:]) {
					return true
				}
			}
			return false
		}
		if len(n) == 0 {
			return false
		}
		if ok, err := path.Match(p[0], n[0]); err != nil || !ok {
			return false
		}
		p, n = p[1:], n[1:]
	}
	return len(n) == 0
}

func (e *Extension) biFSGlob(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var pattern string
	if err := starlark.UnpackArgs("fs.glob", args, kw, "pattern", &pattern); err != nil {
		return nil, err
	}
	pat := filepath.ToSlash(expandHome(pattern))
	segs := strings.Split(pat, "/")
	i := 0
	for i < len(segs) && !hasMeta(segs[i]) {
		i++
	}
	if i == len(segs) { // no wildcard: behaves like exists
		r, err := e.checkFS(PermFSRead, pat)
		if err != nil {
			return nil, err
		}
		if _, err := os.Stat(r); err != nil {
			return starlark.NewList(nil), nil
		}
		d, _ := lexicalPath(pat)
		return goToStarlark([]string{filepath.ToSlash(d)}), nil
	}
	base := strings.Join(segs[:i], "/")
	if base == "" && i == 0 {
		base = "."
	} else if base == "" {
		base = "/"
	}
	if len(base) == 2 && base[1] == ':' {
		base += "/"
	}
	rest := segs[i:]
	r, err := e.checkFS(PermFSRead, base)
	if err != nil {
		return nil, err
	}
	disp, _ := lexicalPath(base)
	fold := runtime.GOOS == "windows"
	if fold {
		for k := range rest {
			rest[k] = strings.ToLower(rest[k])
		}
	}
	var out []string
	scanned := 0
	var rec func(dir, rel string) error
	rec = func(dir, rel string) error {
		ents, err := os.ReadDir(dir)
		if err != nil {
			return err
		}
		for _, d := range ents {
			scanned++
			if scanned > maxScan {
				return fmt.Errorf("more than %d entries to scan (use a narrower pattern)", maxScan)
			}
			rl := path.Join(rel, d.Name())
			cmp := rl
			if fold {
				cmp = strings.ToLower(rl)
			}
			if matchSegs(rest, strings.Split(cmp, "/")) {
				if len(out) >= e.opts.Limits.MaxEntries {
					return fmt.Errorf("more than %d matches (use a narrower pattern)", e.opts.Limits.MaxEntries)
				}
				out = append(out, filepath.ToSlash(filepath.Join(disp, filepath.FromSlash(rl))))
			}
			if d.IsDir() && d.Type()&fs.ModeSymlink == 0 {
				if err := rec(filepath.Join(dir, d.Name()), rl); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := rec(r, ""); err != nil {
		return nil, err
	}
	sort.Strings(out)
	return goToStarlark(out), nil
}

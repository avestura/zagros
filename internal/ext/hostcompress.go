package ext

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/bzip2"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
)

func (e *Extension) compressModule() *starlarkstruct.Module {
	return module("compress", starlark.StringDict{
		"gzip":    e.fn("compress.gzip", e.compressFn("gzip")),
		"gunzip":  e.fn("compress.gunzip", e.decompressFn("gunzip")),
		"zlib":    e.fn("compress.zlib", e.compressFn("zlib")),
		"unzlib":  e.fn("compress.unzlib", e.decompressFn("unzlib")),
		"deflate": e.fn("compress.deflate", e.compressFn("deflate")),
		"inflate": e.fn("compress.inflate", e.decompressFn("inflate")),
		"bunzip2": e.fn("compress.bunzip2", e.decompressFn("bunzip2")),
		"create":  e.fn("compress.create", e.biArchiveCreate),
		"extract": e.fn("compress.extract", e.biArchiveExtract),
		"list":    e.fn("compress.list", e.biArchiveList),
	})
}

func (e *Extension) compressFn(kind string) bfn {
	name := "compress." + kind
	return func(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
		var data starlark.Value
		level := -1
		if err := starlark.UnpackArgs(name, args, kw, "data", &data, "level?", &level); err != nil {
			return nil, err
		}
		in, err := asBytes(data)
		if err != nil {
			return nil, err
		}
		if level < -2 || level > 9 {
			return nil, errors.New("level must be -2..9 (-1 default)")
		}
		var buf bytes.Buffer
		var w io.WriteCloser
		switch kind {
		case "gzip":
			w, err = gzip.NewWriterLevel(&buf, level)
		case "zlib":
			w, err = zlib.NewWriterLevel(&buf, level)
		default:
			w, err = flate.NewWriter(&buf, level)
		}
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(in); err != nil {
			return nil, err
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
		return starlark.Bytes(buf.String()), nil
	}
}

func (e *Extension) decompressFn(kind string) bfn {
	name := "compress." + kind
	return func(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
		var data starlark.Value
		var maxBytes int64
		if err := starlark.UnpackArgs(name, args, kw, "data", &data, "max_bytes?", &maxBytes); err != nil {
			return nil, err
		}
		in, err := asBytes(data)
		if err != nil {
			return nil, err
		}
		var r io.Reader
		br := bytes.NewReader(in)
		switch kind {
		case "gunzip":
			gr, gerr := gzip.NewReader(br)
			if gerr != nil {
				return nil, gerr
			}
			defer gr.Close()
			r = gr
		case "unzlib":
			zr, zerr := zlib.NewReader(br)
			if zerr != nil {
				return nil, zerr
			}
			defer zr.Close()
			r = zr
		case "inflate":
			fr := flate.NewReader(br)
			defer fr.Close()
			r = fr
		default:
			r = bzip2.NewReader(br)
		}
		limit := e.opts.Limits.MaxFileBytes
		if maxBytes > 0 && maxBytes < limit {
			limit = maxBytes
		}
		out, err := io.ReadAll(io.LimitReader(r, limit+1))
		if err != nil {
			return nil, err
		}
		if int64(len(out)) > limit {
			return nil, fmt.Errorf("decompressed data larger than %d bytes", limit)
		}
		return starlark.Bytes(out), nil
	}
}

// ---- archives ----

func archiveFormat(format, name string, sniff []byte) (string, error) {
	f := strings.ToLower(format)
	if f == "" {
		n := strings.ToLower(name)
		switch {
		case strings.HasSuffix(n, ".zip"):
			f = "zip"
		case strings.HasSuffix(n, ".tar.gz"), strings.HasSuffix(n, ".tgz"):
			f = "tar.gz"
		case strings.HasSuffix(n, ".tar"):
			f = "tar"
		case len(sniff) >= 4 && string(sniff[:2]) == "PK":
			f = "zip"
		case len(sniff) >= 2 && sniff[0] == 0x1f && sniff[1] == 0x8b:
			f = "tar.gz"
		}
	}
	switch f {
	case "zip", "tar", "tar.gz":
		return f, nil
	case "tgz":
		return "tar.gz", nil
	}
	return "", fmt.Errorf("cannot tell the archive format of %q; pass format = \"zip\", \"tar\" or \"tar.gz\"", name)
}

func sniffFile(p string) []byte {
	f, err := os.Open(p)
	if err != nil {
		return nil
	}
	defer f.Close()
	b := make([]byte, 4)
	n, _ := io.ReadFull(f, b)
	return b[:n]
}

type archSrc struct {
	disk string
	name string
	info fs.FileInfo
}

func (e *Extension) collectSources(srcs []string, skip string) ([]archSrc, error) {
	var out []archSrc
	var total int64
	var walk func(disk, name string, info fs.FileInfo) error
	walk = func(disk, name string, info fs.FileInfo) error {
		if samePath(disk, skip) {
			return nil
		}
		if len(out) >= e.opts.Limits.MaxEntries {
			return fmt.Errorf("more than %d files to archive", e.opts.Limits.MaxEntries)
		}
		if info.IsDir() {
			out = append(out, archSrc{disk, name + "/", info})
			ents, err := os.ReadDir(disk)
			if err != nil {
				return err
			}
			for _, d := range ents {
				ci, err := d.Info()
				if err != nil {
					return err
				}
				if ci.Mode()&fs.ModeSymlink != 0 || !(ci.IsDir() || ci.Mode().IsRegular()) {
					continue
				}
				if err := walk(filepath.Join(disk, d.Name()), name+"/"+d.Name(), ci); err != nil {
					return err
				}
			}
			return nil
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("%s is not a regular file", disk)
		}
		total += info.Size()
		if total > e.opts.Limits.MaxBulkBytes {
			return fmt.Errorf("sources larger than %d bytes", e.opts.Limits.MaxBulkBytes)
		}
		out = append(out, archSrc{disk, name, info})
		return nil
	}
	for _, s := range srcs {
		r, err := e.checkFS(PermFSRead, s)
		if err != nil {
			return nil, err
		}
		info, err := os.Stat(r)
		if err != nil {
			return nil, err
		}
		lex, _ := lexicalPath(s)
		if err := walk(r, filepath.Base(lex), info); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func (e *Extension) biArchiveCreate(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var dest, format string
	var srcv starlark.Value
	if err := starlark.UnpackArgs("compress.create", args, kw, "dest", &dest, "sources", &srcv, "format?", &format); err != nil {
		return nil, err
	}
	srcs, err := stringList(srcv)
	if err != nil {
		return nil, err
	}
	if len(srcs) == 0 {
		return nil, errors.New("sources is empty")
	}
	f, err := archiveFormat(format, dest, nil)
	if err != nil {
		return nil, err
	}
	d, err := e.checkFS(PermFSWrite, dest)
	if err != nil {
		return nil, err
	}
	files, err := e.collectSources(srcs, d)
	if err != nil {
		return nil, err
	}
	tmp := d + ".part"
	out, err := os.Create(tmp)
	if err != nil {
		return nil, err
	}
	err = writeArchive(out, f, files)
	if cerr := out.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		os.Remove(d)
		err = os.Rename(tmp, d)
	}
	if err != nil {
		os.Remove(tmp)
		return nil, err
	}
	return starlark.MakeInt(len(files)), nil
}

func writeArchive(out io.Writer, format string, files []archSrc) error {
	if format == "zip" {
		zw := zip.NewWriter(out)
		for _, s := range files {
			h, err := zip.FileInfoHeader(s.info)
			if err != nil {
				return err
			}
			h.Name = s.name
			if !s.info.IsDir() {
				h.Method = zip.Deflate
			}
			w, err := zw.CreateHeader(h)
			if err != nil {
				return err
			}
			if s.info.IsDir() {
				continue
			}
			if err := copyFileTo(w, s.disk); err != nil {
				return err
			}
		}
		return zw.Close()
	}
	var gz *gzip.Writer
	w := out
	if format == "tar.gz" {
		gz = gzip.NewWriter(out)
		w = gz
	}
	tw := tar.NewWriter(w)
	for _, s := range files {
		h, err := tar.FileInfoHeader(s.info, "")
		if err != nil {
			return err
		}
		h.Name = s.name
		h.Uname, h.Gname, h.Uid, h.Gid = "", "", 0, 0
		if err := tw.WriteHeader(h); err != nil {
			return err
		}
		if s.info.IsDir() {
			continue
		}
		if err := copyFileTo(tw, s.disk); err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	if gz != nil {
		return gz.Close()
	}
	return nil
}

func copyFileTo(w io.Writer, p string) error {
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(w, f)
	return err
}

// archiveEntry is one member of an archive being read.
type archiveEntry struct {
	name  string
	size  int64
	dir   bool
	skip  bool // links and special files: never extracted
	mtime int64
	open  func() (io.Reader, error)
}

// walkArchive calls fn for each entry; reading the reader is only valid inside fn.
func (e *Extension) walkArchive(p, format string, fn func(archiveEntry) error) error {
	n := 0
	count := func() error {
		n++
		if n > e.opts.Limits.MaxEntries {
			return fmt.Errorf("archive has more than %d entries", e.opts.Limits.MaxEntries)
		}
		return nil
	}
	if format == "zip" {
		zr, err := zip.OpenReader(p)
		if err != nil {
			return err
		}
		defer zr.Close()
		if len(zr.File) > e.opts.Limits.MaxEntries {
			return fmt.Errorf("archive has more than %d entries", e.opts.Limits.MaxEntries)
		}
		for _, f := range zr.File {
			ent := archiveEntry{name: f.Name, size: int64(f.UncompressedSize64), dir: f.FileInfo().IsDir(), mtime: f.Modified.Unix(),
				skip: !f.Mode().IsRegular() && !f.FileInfo().IsDir()}
			ent.open = func() (io.Reader, error) { return f.Open() }
			if err := fn(ent); err != nil {
				return err
			}
		}
		return nil
	}
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	var r io.Reader = f
	if format == "tar.gz" {
		gr, err := gzip.NewReader(f)
		if err != nil {
			return err
		}
		defer gr.Close()
		r = gr
	}
	tr := tar.NewReader(r)
	for {
		h, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		if err := count(); err != nil {
			return err
		}
		ent := archiveEntry{name: h.Name, size: h.Size, dir: h.Typeflag == tar.TypeDir, mtime: h.ModTime.Unix(),
			skip: h.Typeflag != tar.TypeReg && h.Typeflag != tar.TypeDir}
		ent.open = func() (io.Reader, error) { return tr, nil }
		if err := fn(ent); err != nil {
			return err
		}
	}
}

func (e *Extension) biArchiveList(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p, format string
	if err := starlark.UnpackArgs("compress.list", args, kw, "archive", &p, "format?", &format); err != nil {
		return nil, err
	}
	r, err := e.checkFS(PermFSRead, p)
	if err != nil {
		return nil, err
	}
	f, err := archiveFormat(format, p, sniffFile(r))
	if err != nil {
		return nil, err
	}
	var out []starlark.Value
	err = e.walkArchive(r, f, func(a archiveEntry) error {
		out = append(out, starlarkstruct.FromStringDict(starlark.String("entry"), starlark.StringDict{
			"name": starlark.String(a.name), "size": starlark.MakeInt64(a.size), "is_dir": starlark.Bool(a.dir), "mtime": starlark.MakeInt64(a.mtime),
		}))
		return nil
	})
	if err != nil {
		return nil, err
	}
	return starlark.NewList(out), nil
}

func (e *Extension) biArchiveExtract(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p, dest, format string
	var overwrite bool
	if err := starlark.UnpackArgs("compress.extract", args, kw, "archive", &p, "dest", &dest, "format?", &format, "overwrite?", &overwrite); err != nil {
		return nil, err
	}
	src, err := e.checkFS(PermFSRead, p)
	if err != nil {
		return nil, err
	}
	d, err := e.checkFS(PermFSWrite, dest)
	if err != nil {
		return nil, err
	}
	f, err := archiveFormat(format, p, sniffFile(src))
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(d, 0o755); err != nil {
		return nil, err
	}
	root, err := resolvePath(d)
	if err != nil {
		return nil, err
	}
	var names []string
	var total int64
	err = e.walkArchive(src, f, func(a archiveEntry) error {
		if a.skip {
			return nil
		}
		clean, err := safeEntryName(a.name)
		if err != nil {
			return err
		}
		if clean == "" {
			return nil
		}
		target := filepath.Join(root, filepath.FromSlash(clean))
		if res, err := resolvePath(target); err != nil || !pathWithin(root, res) {
			return fmt.Errorf("archive entry %q escapes the destination", a.name)
		}
		if a.dir {
			return os.MkdirAll(target, 0o755)
		}
		if a.size > e.opts.Limits.MaxBulkBytes-total {
			return fmt.Errorf("archive larger than %d bytes when extracted", e.opts.Limits.MaxBulkBytes)
		}
		if _, err := os.Lstat(target); err == nil && !overwrite {
			return fmt.Errorf("%s already exists (pass overwrite = True)", clean)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if res, _ := resolvePath(target); !pathWithin(root, res) {
			return fmt.Errorf("archive entry %q escapes the destination", a.name)
		}
		rd, err := a.open()
		if err != nil {
			return err
		}
		if c, ok := rd.(io.Closer); ok {
			defer c.Close()
		}
		out, err := os.Create(target)
		if err != nil {
			return err
		}
		remain := e.opts.Limits.MaxBulkBytes - total
		n, err := io.Copy(out, io.LimitReader(rd, remain+1))
		if cerr := out.Close(); err == nil {
			err = cerr
		}
		if err == nil && n > remain {
			err = fmt.Errorf("archive larger than %d bytes when extracted", e.opts.Limits.MaxBulkBytes)
		}
		if err != nil {
			os.Remove(target)
			return err
		}
		total += n
		names = append(names, clean)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return goToStarlark(names), nil
}

// safeEntryName rejects absolute names, drive letters and ".." so an archive
// entry can only land below the destination.
func safeEntryName(name string) (string, error) {
	n := strings.ReplaceAll(name, `\`, "/")
	n = strings.TrimSuffix(n, "/")
	if n == "" || n == "." {
		return "", nil
	}
	if !filepath.IsLocal(filepath.FromSlash(n)) || strings.Contains(n, ":") || strings.HasPrefix(n, "/") {
		return "", fmt.Errorf("unsafe archive entry name %q", name)
	}
	return path.Clean(n), nil
}

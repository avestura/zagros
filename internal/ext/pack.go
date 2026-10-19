package ext

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"go.starlark.net/starlark"
)

// Archive limits for installs and packs.
const (
	MaxArchiveBytes  = 16 << 20 // compressed .zagros size
	MaxUnpackedBytes = 64 << 20 // sum of all file sizes
	MaxArchiveFiles  = 2000
	MaxSingleFile    = 16 << 20
)

// Extension is the file extension for packed extensions.
const PackExt = ".zagros"

// StubsDir is the folder `zagros ext init` fills with the .pyi API stubs; it is
// never packed.
const StubsDir = "zagros-stubs"

// Pack zips the extension in dir (manifest.toml at the archive root). Hidden
// files and folders (names starting with ".") and *.zagros files are skipped.
// The manifest is validated first. Output is deterministic.
func Pack(dir string) ([]byte, error) {
	mb, err := os.ReadFile(filepath.Join(dir, ManifestFile))
	if err != nil {
		return nil, fmt.Errorf("pack: %w", err)
	}
	if _, err := ParseManifest(mb); err != nil {
		return nil, err
	}
	var files []string
	err = filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if p == dir {
			return nil
		}
		name := d.Name()
		if strings.HasPrefix(name, ".") {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			if name == StubsDir { // the agent stubs from `zagros ext init` are dev-only
				return filepath.SkipDir
			}
			return nil
		}
		if p == filepath.Join(dir, "AGENTS.md") || p == filepath.Join(dir, "CLAUDE.md") || p == filepath.Join(dir, "API_INDEX.md") {
			return nil
		}
		if strings.HasSuffix(name, PackExt) {
			return nil
		}
		if !d.Type().IsRegular() {
			return fmt.Errorf("pack: %s is not a regular file", p)
		}
		rel, _ := filepath.Rel(dir, p)
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(files)
	if len(files) > MaxArchiveFiles {
		return nil, fmt.Errorf("pack: %d files exceeds limit %d", len(files), MaxArchiveFiles)
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	var total int64
	for _, rel := range files {
		b, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(rel)))
		if err != nil {
			return nil, err
		}
		total += int64(len(b))
		if int64(len(b)) > MaxSingleFile || total > MaxUnpackedBytes {
			return nil, fmt.Errorf("pack: size limit exceeded at %s", rel)
		}
		w, err := zw.CreateHeader(&zip.FileHeader{Name: rel, Method: zip.Deflate})
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(b); err != nil {
			return nil, err
		}
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	if buf.Len() > MaxArchiveBytes {
		return nil, fmt.Errorf("pack: archive is %d bytes, limit is %d", buf.Len(), MaxArchiveBytes)
	}
	return buf.Bytes(), nil
}

// PackFile packs dir into outPath (default "<dir-name>-<version>.zagros" next
// to dir when outPath is empty) and returns the path written.
func PackFile(dir, outPath string) (string, error) {
	data, err := Pack(dir)
	if err != nil {
		return "", err
	}
	if outPath == "" {
		mb, _ := os.ReadFile(filepath.Join(dir, ManifestFile))
		m, _ := ParseManifest(mb)
		abs, _ := filepath.Abs(dir)
		outPath = filepath.Join(filepath.Dir(abs), dirName(m.ID)+"-"+m.Version+PackExt)
	}
	return outPath, os.WriteFile(outPath, data, 0o644)
}

// safeRel validates an archive entry name and returns the cleaned
// slash-separated relative path.
func safeRel(name string) (string, error) {
	n := strings.ReplaceAll(name, "\\", "/")
	if n == "" || strings.HasPrefix(n, "/") || strings.Contains(n, ":") {
		return "", fmt.Errorf("unsafe path %q in archive", name)
	}
	for _, seg := range strings.Split(strings.TrimSuffix(n, "/"), "/") {
		if seg == ".." || seg == "" {
			return "", fmt.Errorf("unsafe path %q in archive", name)
		}
	}
	c := path.Clean(n)
	if c == "." || strings.HasPrefix(c, "..") {
		return "", fmt.Errorf("unsafe path %q in archive", name)
	}
	return c, nil
}

// extractZip unpacks data into dest (which must exist) with zip-slip, symlink
// and size protection. A single wrapping top-level folder is stripped. It
// returns the manifest bytes.
func extractZip(data []byte, dest string) ([]byte, error) {
	if len(data) > MaxArchiveBytes {
		return nil, fmt.Errorf("archive is %d bytes, limit is %d", len(data), MaxArchiveBytes)
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, fmt.Errorf("not a valid zip: %w", err)
	}
	if len(zr.File) > MaxArchiveFiles {
		return nil, fmt.Errorf("archive has %d entries, limit is %d", len(zr.File), MaxArchiveFiles)
	}
	prefix, found := "", false
	for _, f := range zr.File {
		n, err := safeRel(f.Name)
		if err != nil {
			return nil, err
		}
		if n == ManifestFile {
			prefix, found = "", true
			break
		}
		if strings.Count(n, "/") == 1 && strings.HasSuffix(n, "/"+ManifestFile) && !found {
			prefix, found = strings.TrimSuffix(n, ManifestFile), true
		}
	}
	if !found {
		return nil, fmt.Errorf("archive has no %s", ManifestFile)
	}
	var total int64
	var manifest []byte
	absDest, _ := filepath.Abs(dest)
	for _, f := range zr.File {
		n, err := safeRel(f.Name)
		if err != nil {
			return nil, err
		}
		if f.Mode()&os.ModeSymlink != 0 {
			return nil, fmt.Errorf("symlink %q not allowed in archive", f.Name)
		}
		if prefix != "" {
			if !strings.HasPrefix(n+"/", prefix) && !strings.HasPrefix(n, prefix) {
				return nil, fmt.Errorf("entry %q is outside the top-level folder %q", f.Name, prefix)
			}
			n = strings.TrimPrefix(n, prefix)
			if n == "" || n == "." {
				continue
			}
		}
		if strings.HasPrefix(n, "__MACOSX/") {
			continue
		}
		target := filepath.Join(absDest, filepath.FromSlash(n))
		if !pathWithin(absDest, target) {
			return nil, fmt.Errorf("unsafe path %q in archive", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return nil, err
			}
			continue
		}
		if f.UncompressedSize64 > MaxSingleFile {
			return nil, fmt.Errorf("%q is too large", f.Name)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return nil, err
		}
		rc, err := f.Open()
		if err != nil {
			return nil, err
		}
		// Enforce the limit on actual bytes, not the declared size.
		b, err := io.ReadAll(io.LimitReader(rc, MaxSingleFile+1))
		rc.Close()
		if err != nil {
			return nil, err
		}
		total += int64(len(b))
		if int64(len(b)) > MaxSingleFile || total > MaxUnpackedBytes {
			return nil, fmt.Errorf("archive exceeds unpacked size limit (%d bytes)", MaxUnpackedBytes)
		}
		if err := os.WriteFile(target, b, 0o644); err != nil {
			return nil, err
		}
		if n == ManifestFile {
			manifest = b
		}
	}
	if manifest == nil {
		return nil, fmt.Errorf("archive has no %s", ManifestFile)
	}
	return manifest, nil
}

// Issue is one finding of Validate.
type Issue struct {
	Error bool
	Msg   string
}

func (i Issue) String() string {
	if i.Error {
		return "error: " + i.Msg
	}
	return "warning: " + i.Msg
}

// HasErrors reports whether any issue is an error.
func HasErrors(is []Issue) bool {
	for _, i := range is {
		if i.Error {
			return true
		}
	}
	return false
}

var hostModules = map[string]string{
	"http.":                PermNetwork,
	"tcp.":                 PermNetwork,
	"udp.":                 PermNetwork,
	"tls.":                 PermNetwork,
	"dns.":                 PermNetwork,
	"ssh.":                 PermSSH,
	"fs.write_text":        PermFSWrite,
	"fs.":                  PermFSRead,
	"clipboard.get":        PermClipboardRead,
	"clipboard.set":        PermClipboardWrite,
	"proc.":                PermProcessRun,
	"env.":                 PermEnvRead,
	"secrets.":             PermSecrets,
	"sql.open":             PermSQL,
	"ldap.connect":         PermLDAP,
	"ai.complete(":         PermAI,
	"ai.run(":              PermAI,
	"ai.embed(":            PermAI,
	"ai.providers(":        PermAI,
	"windows.registry":     PermRegistryRead,
	"windows.service":      PermSystemRead,
	"windows.list_windows": PermSystemRead,
	"windows.processes":    PermSystemRead,
	"windows.known_folder": PermSystemRead,
	"notify(":              PermNotifications,
	"action_open(":         PermOpen,
	"action_run(":          PermProcessRun,
	"action_copy(":         PermClipboardWrite,
}

// Validate statically checks the extension in dir: manifest, entry script
// (syntax and unresolved names), hooks, and permission usage. It runs no code.
func Validate(dir string) []Issue {
	var out []Issue
	errf := func(f string, a ...any) { out = append(out, Issue{true, fmt.Sprintf(f, a...)}) }
	warnf := func(f string, a ...any) { out = append(out, Issue{false, fmt.Sprintf(f, a...)}) }

	mb, err := os.ReadFile(filepath.Join(dir, ManifestFile))
	if err != nil {
		errf("%v", err)
		return out
	}
	m, warns, err := parseManifest(mb)
	for _, w := range warns {
		warnf("%s", w)
	}
	if err != nil {
		errf("%v", err)
		if m == nil || m.Entry == "" {
			return out
		}
	}
	if m.Description == "" {
		warnf("description is empty")
	}
	if m.Author == "" {
		warnf("author is empty")
	}
	if len(m.Keywords) > 0 && len(m.Examples) == 0 {
		warnf("keywords are set but examples is empty: the extension manager cannot show users what to type")
	}
	if len(m.Keywords) == 0 && m.Usage == "" {
		warnf("no keywords and no usage: users cannot tell how to trigger the extension")
	}
	src, err := os.ReadFile(filepath.Join(dir, filepath.FromSlash(m.Entry)))
	if err != nil {
		errf("entry %s: %v", m.Entry, err)
		return out
	}
	e := &Extension{Manifest: m, Dir: dir, opts: &Options{}, kv: newKV("", 0)}
	e.opts.defaults()
	pre := e.predeclared()
	_, prog, err := starlark.SourceProgramOptions(fileOpts, m.Entry, src, pre.Has)
	_ = prog
	if err != nil {
		errf("%s: %v", m.Entry, err)
		return out
	}
	text := string(src)
	for _, hook := range []string{"on_query", "on_execute"} {
		if !strings.Contains(text, "def "+hook+"(") {
			if hook == "on_query" {
				warnf("%s does not define on_query: the extension will never produce results", m.Entry)
			}
		}
	}
	declared := map[string]bool{}
	for _, p := range m.Permissions {
		declared[p.Kind()] = true
	}
	reported := map[string]bool{}
	for needle, kind := range hostModules {
		if strings.Contains(text, needle) && !declared[kind] && !reported[kind] {
			reported[kind] = true
			warnf("script appears to use %q but permission %q is not declared", strings.TrimSuffix(needle, "("), kind)
		}
	}
	for _, p := range m.Permissions {
		if p.Kind() == PermNetwork && p.Scope() == "" {
			warnf("unscoped %q lets the extension contact any host; prefer network:<host>", PermNetwork)
		}
	}
	return out
}

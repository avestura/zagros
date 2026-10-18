package ext

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/sftp"
	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"golang.org/x/crypto/ssh"
)

const knownHostsFile = "ssh_known_hosts.json"

func (e *Extension) sshModule() *starlarkstruct.Module {
	return module("ssh", starlark.StringDict{"connect": e.fn("ssh.connect", e.biSSHConnect)})
}

// ---- host keys ----

// sshKnown returns the recorded key fingerprints of host:port by key type.
// They live in the extension's data folder, outside storage, so a script
// cannot rewrite them; without a data folder they stay in memory.
func (e *Extension) sshKnown(key string) map[string]string {
	e.host.net.mu.Lock()
	defer e.host.net.mu.Unlock()
	return copyMap(e.loadKnownLocked()[key])
}

func copyMap(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func (e *Extension) loadKnownLocked() map[string]map[string]string {
	if e.dataDir == "" {
		if e.host.net.known == nil {
			e.host.net.known = map[string]map[string]string{}
		}
		return e.host.net.known
	}
	all := map[string]map[string]string{}
	if b, err := os.ReadFile(filepath.Join(e.dataDir, knownHostsFile)); err == nil {
		_ = json.Unmarshal(b, &all)
	}
	return all
}

func (e *Extension) sshRemember(key, typ, fp string) error {
	e.host.net.mu.Lock()
	defer e.host.net.mu.Unlock()
	all := e.loadKnownLocked()
	if all[key] == nil {
		all[key] = map[string]string{}
	}
	all[key][typ] = fp
	if e.dataDir == "" {
		return nil
	}
	b, err := json.MarshalIndent(all, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(filepath.Join(e.dataDir, knownHostsFile), b)
}

// keyAlgos maps recorded key types to the algorithms to offer, so the server
// presents a key of a type already on record instead of triggering a false alarm.
func keyAlgos(types map[string]string) []string {
	var out []string
	for t := range types {
		if t == ssh.KeyAlgoRSA {
			out = append(out, ssh.KeyAlgoRSASHA512, ssh.KeyAlgoRSASHA256)
		} else {
			out = append(out, t)
		}
	}
	sort.Strings(out)
	return out
}

func normFingerprint(s string) string {
	s = strings.TrimSpace(s)
	if !strings.HasPrefix(s, "SHA256:") {
		s = "SHA256:" + s
	}
	return strings.TrimRight(s, "=")
}

type hostKeyState struct {
	fp, status string
}

// hostKeyCallback verifies the server key against the pins, or the recorded
// keys, or records it when the host is new and trust_new allows. It never
// accepts a key that differs from what is on record.
func (e *Extension) hostKeyCallback(id string, pins []string, trustNew bool, known map[string]string, st *hostKeyState) ssh.HostKeyCallback {
	return func(_ string, _ net.Addr, key ssh.PublicKey) error {
		fp := strings.TrimRight(ssh.FingerprintSHA256(key), "=")
		st.fp = fp
		if len(pins) > 0 {
			for _, p := range pins {
				if p == fp {
					st.status = "pinned"
					return nil
				}
			}
			return fmt.Errorf("ssh: host key of %s does not match host_key: the server presented %s", id, fp)
		}
		if want, ok := known[key.Type()]; ok {
			if want == fp {
				st.status = "known"
				return nil
			}
			return fmt.Errorf("ssh: the host key of %s has changed (recorded %s, server presented %s). Someone may be intercepting the connection. If the server key was replaced on purpose, check the fingerprint with the server's owner and pass host_key=%q", id, want, fp, fp)
		}
		if len(known) > 0 {
			return fmt.Errorf("ssh: %s presented a %s key (%s) but other key types are on record; pass host_key=%q once you have checked it", id, key.Type(), fp, fp)
		}
		if !trustNew {
			return fmt.Errorf("ssh: %s is not a known host (key %s). Check this fingerprint with the server's owner, then pass host_key=%q", id, fp, fp)
		}
		if err := e.sshRemember(id, key.Type(), fp); err != nil {
			return fmt.Errorf("ssh: could not record the host key: %w", err)
		}
		st.status = "new"
		return nil
	}
}

// ---- auth ----

func (e *Extension) sshSecret(name string) (string, error) {
	if err := e.require(PermSecrets, ""); err != nil {
		return "", err
	}
	if err := secretKey(name); err != nil {
		return "", err
	}
	v, ok, err := secretBackend.Get(secretTarget(e.Manifest.ID, name))
	if err != nil {
		return "", fmt.Errorf("secret %q: %w", name, err)
	}
	if !ok {
		return "", fmt.Errorf("secret %q is not set", name)
	}
	return v, nil
}

// sshAuth builds the auth methods from {"secret"|"password", "key_secret"|"key",
// "passphrase_secret"|"passphrase"}.
func (e *Extension) sshAuth(d *starlark.Dict) ([]ssh.AuthMethod, error) {
	if d == nil || d.Len() == 0 {
		return nil, errors.New(`ssh.connect: auth is required, e.g. auth={"secret": "name"} or auth={"key_secret": "name"}`)
	}
	m, err := stringMap(d)
	if err != nil {
		return nil, errors.New("ssh.connect: auth must be a dict of strings")
	}
	for k := range m {
		switch k {
		case "secret", "password", "key_secret", "key", "passphrase_secret", "passphrase":
		default:
			return nil, fmt.Errorf("ssh.connect: unknown auth field %q", k)
		}
	}
	pick := func(secretField, rawField string) (string, bool, error) {
		if s, ok := m[secretField]; ok {
			v, err := e.sshSecret(s)
			return v, true, err
		}
		if v, ok := m[rawField]; ok {
			return v, true, nil
		}
		return "", false, nil
	}
	var out []ssh.AuthMethod
	pw, hasPW, err := pick("secret", "password")
	if err != nil {
		return nil, err
	}
	pem, hasKey, err := pick("key_secret", "key")
	if err != nil {
		return nil, err
	}
	pass, hasPass, err := pick("passphrase_secret", "passphrase")
	if err != nil {
		return nil, err
	}
	if hasKey {
		var signer ssh.Signer
		if hasPass {
			signer, err = ssh.ParsePrivateKeyWithPassphrase([]byte(pem), []byte(pass))
		} else {
			signer, err = ssh.ParsePrivateKey([]byte(pem))
		}
		var miss *ssh.PassphraseMissingError
		if errors.As(err, &miss) {
			return nil, errors.New("ssh.connect: the private key is encrypted; add passphrase_secret to auth")
		}
		if err != nil {
			return nil, fmt.Errorf("ssh.connect: private key: %w", err)
		}
		out = append(out, ssh.PublicKeys(signer))
	}
	if hasPW {
		out = append(out, ssh.Password(pw), ssh.KeyboardInteractive(func(_, _ string, qs []string, _ []bool) ([]string, error) {
			ans := make([]string, len(qs))
			for i := range ans {
				ans[i] = pw
			}
			return ans, nil
		}))
	}
	if len(out) == 0 {
		return nil, errors.New("ssh.connect: auth needs a password or a key (secret, password, key_secret, key)")
	}
	return out, nil
}

// ---- connect ----

type sshConn struct {
	e      *Extension
	cl     *ssh.Client
	id     string
	user   string
	fp     string
	status string
	stop   func() bool
	mu     sync.Mutex
	sftpc  *sftp.Client
	closed bool
}

func (e *Extension) biSSHConnect(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var host, user string
	port := 22
	var auth *starlark.Dict
	var hostKey starlark.Value = starlark.None
	timeout := secs(10)
	trustNew := true
	if err := starlark.UnpackArgs("ssh.connect", args, kw, "host", &host, "user", &user, "port?", &port, "auth", &auth,
		"timeout?", &timeout, "host_key?", &hostKey, "trust_new?", &trustNew); err != nil {
		return nil, err
	}
	if err := checkHostName(host); err != nil {
		return nil, fmt.Errorf("ssh.connect: %w", err)
	}
	if err := checkPort(port); err != nil {
		return nil, fmt.Errorf("ssh.connect: %w", err)
	}
	if user == "" || strings.ContainsAny(user, " \t\r\n:/") {
		return nil, fmt.Errorf("ssh.connect: invalid user %q", user)
	}
	d, err := toTimeout("ssh.connect", timeout)
	if err != nil {
		return nil, err
	}
	hp := net.JoinHostPort(strings.ToLower(host), strconv.Itoa(port))
	if err := e.require(PermSSH, user+"@"+hp); err != nil {
		return nil, err
	}
	var pins []string
	switch x := hostKey.(type) {
	case starlark.NoneType:
	case starlark.String:
		pins = []string{normFingerprint(string(x))}
	default:
		l, err := stringList(hostKey)
		if err != nil || len(l) == 0 {
			return nil, errors.New("ssh.connect: host_key must be a fingerprint string or a list of them")
		}
		for _, s := range l {
			pins = append(pins, normFingerprint(s))
		}
	}
	methods, err := e.sshAuth(auth)
	if err != nil {
		return nil, err
	}
	known := e.sshKnown(hp)
	st := &hostKeyState{}
	cfg := &ssh.ClientConfig{
		User:            user,
		Auth:            methods,
		HostKeyCallback: e.hostKeyCallback(hp, pins, trustNew, known, st),
		Timeout:         d,
	}
	if len(pins) == 0 && len(known) > 0 {
		cfg.HostKeyAlgorithms = keyAlgos(known)
	}
	c, err := dialHost(callCtx(th), "tcp", host, port, d, !e.hasUnscoped(PermSSH))
	if err != nil {
		return nil, fmt.Errorf("ssh.connect %s: %w", hp, err)
	}
	_ = c.SetDeadline(time.Now().Add(d))
	cc, chans, reqs, err := ssh.NewClientConn(c, hp, cfg)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("ssh.connect %s@%s: %w", user, hp, err)
	}
	_ = c.SetDeadline(time.Now().Add(netMaxLife))
	s := &sshConn{e: e, cl: ssh.NewClient(cc, chans, reqs), id: user + "@" + hp, user: user, fp: st.fp, status: st.status}
	if err := e.trackConn(s); err != nil {
		s.cl.Close()
		return nil, err
	}
	s.stop = context.AfterFunc(callCtx(th), func() { s.Close() })
	return s, nil
}

func (s *sshConn) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	sc := s.sftpc
	s.mu.Unlock()
	if s.stop != nil {
		s.stop()
	}
	if sc != nil {
		sc.Close()
	}
	s.cl.Close()
	s.e.untrackConn(s)
	return nil
}

var _ starlark.HasAttrs = (*sshConn)(nil)

func (s *sshConn) String() string        { return "<ssh.Conn " + s.id + ">" }
func (s *sshConn) Type() string          { return "ssh.Conn" }
func (s *sshConn) Freeze()               {}
func (s *sshConn) Truth() starlark.Bool  { return true }
func (s *sshConn) Hash() (uint32, error) { return 0, errors.New("unhashable type: ssh.Conn") }
func (s *sshConn) AttrNames() []string {
	return []string{"close", "download", "exec", "fingerprint", "get", "host_key_status", "list", "mkdir", "put", "remove", "stat", "upload", "user"}
}

func (s *sshConn) Attr(name string) (starlark.Value, error) {
	m := map[string]bfn{
		"exec": s.exec, "get": s.get, "put": s.put, "download": s.download, "upload": s.upload,
		"list": s.list, "stat": s.stat, "remove": s.remove, "mkdir": s.mkdir, "close": s.closeFn,
	}
	if f, ok := m[name]; ok {
		return starlark.NewBuiltin("ssh."+name, func(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
			return f(th, args, kw)
		}), nil
	}
	switch name {
	case "fingerprint":
		return starlark.String(s.fp), nil
	case "host_key_status":
		return starlark.String(s.status), nil
	case "user":
		return starlark.String(s.user), nil
	}
	return nil, nil
}

func (s *sshConn) closeFn(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs("ssh.close", args, kw); err != nil {
		return nil, err
	}
	s.Close()
	return starlark.None, nil
}

func (s *sshConn) alive() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("ssh: connection is closed")
	}
	return nil
}

// ---- exec ----

// sshCapBuf keeps up to n bytes and signals when more arrives; the excess is
// dropped so the remote side never stalls before the session is closed.
type sshCapBuf struct {
	mu   sync.Mutex
	b    bytes.Buffer
	n    int64
	over chan struct{}
	hit  bool
}

func (c *sshCapBuf) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if room := c.n - int64(c.b.Len()); int64(len(p)) > room {
		if room > 0 {
			c.b.Write(p[:room])
		}
		if !c.hit {
			c.hit = true
			close(c.over)
		}
		return len(p), nil
	}
	return c.b.Write(p)
}

func (s *sshConn) exec(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var cmd string
	var stdin starlark.Value = starlark.None
	timeout := secs(30)
	if err := starlark.UnpackArgs("ssh.exec", args, kw, "cmd", &cmd, "timeout?", &timeout, "stdin?", &stdin); err != nil {
		return nil, err
	}
	d, err := toTimeout("ssh.exec", timeout)
	if err != nil {
		return nil, err
	}
	if cmd == "" {
		return nil, errors.New("ssh.exec: cmd is empty")
	}
	if err := s.alive(); err != nil {
		return nil, err
	}
	sess, err := s.cl.NewSession()
	if err != nil {
		return nil, fmt.Errorf("ssh.exec: %w", err)
	}
	defer sess.Close()
	limit := s.e.opts.Limits.MaxHTTPBytes
	over := make(chan struct{})
	so, se := &sshCapBuf{n: limit, over: over}, &sshCapBuf{n: limit, over: make(chan struct{})}
	sess.Stdout, sess.Stderr = so, se
	if stdin != starlark.None {
		in, err := dataBytes(stdin)
		if err != nil {
			return nil, fmt.Errorf("ssh.exec: stdin: %w", err)
		}
		sess.Stdin = bytes.NewReader(in)
	}
	done := make(chan error, 1)
	go func() { done <- sess.Run(cmd) }()
	ctx := callCtx(th)
	timer := time.NewTimer(d)
	defer timer.Stop()
	var runErr error
	select {
	case runErr = <-done:
	case <-timer.C:
		sess.Close()
		return nil, fmt.Errorf("ssh.exec: command did not finish within %s", d)
	case <-ctx.Done():
		sess.Close()
		return nil, fmt.Errorf("ssh.exec: %w", ctx.Err())
	case <-so.over:
		sess.Close()
		return nil, fmt.Errorf("ssh.exec: stdout larger than %d bytes", limit)
	case <-se.over:
		sess.Close()
		return nil, fmt.Errorf("ssh.exec: stderr larger than %d bytes", limit)
	}
	code := 0
	var xe *ssh.ExitError
	var me *ssh.ExitMissingError
	switch {
	case runErr == nil:
	case errors.As(runErr, &xe):
		code = xe.ExitStatus()
	case errors.As(runErr, &me):
		code = -1
	default:
		return nil, fmt.Errorf("ssh.exec: %w", runErr)
	}
	so.mu.Lock()
	se.mu.Lock()
	defer so.mu.Unlock()
	defer se.mu.Unlock()
	return starlarkstruct.FromStringDict(starlark.String("ssh_result"), starlark.StringDict{
		"stdout": starlark.String(so.b.String()),
		"stderr": starlark.String(se.b.String()),
		"code":   starlark.MakeInt(code),
		"ok":     starlark.Bool(code == 0),
	}), nil
}

// ---- sftp ----

func (s *sshConn) sftp() (*sftp.Client, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil, errors.New("ssh: connection is closed")
	}
	if s.sftpc == nil {
		c, err := sftp.NewClient(s.cl)
		if err != nil {
			return nil, fmt.Errorf("ssh: the server has no sftp subsystem: %w", err)
		}
		s.sftpc = c
	}
	return s.sftpc, nil
}

func remotePath(name, p string) (string, error) {
	if p == "" || strings.ContainsRune(p, 0) {
		return "", fmt.Errorf("%s: invalid path %q", name, p)
	}
	return p, nil
}

func (s *sshConn) readRemote(name, p string, limit int64, w io.Writer) (int64, error) {
	c, err := s.sftp()
	if err != nil {
		return 0, err
	}
	f, err := c.Open(p)
	if err != nil {
		return 0, fmt.Errorf("%s %s: %w", name, p, err)
	}
	defer f.Close()
	n, err := io.Copy(w, io.LimitReader(f, limit+1))
	if err == nil && n > limit {
		err = fmt.Errorf("file larger than %d bytes", limit)
	}
	if err != nil {
		return n, fmt.Errorf("%s %s: %w", name, p, err)
	}
	return n, nil
}

func (s *sshConn) writeRemote(name, p string, r io.Reader, limit int64) (int64, error) {
	c, err := s.sftp()
	if err != nil {
		return 0, err
	}
	f, err := c.OpenFile(p, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
	if err != nil {
		return 0, fmt.Errorf("%s %s: %w", name, p, err)
	}
	n, err := io.Copy(f, io.LimitReader(r, limit+1))
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil && n > limit {
		err = fmt.Errorf("data larger than %d bytes", limit)
	}
	if err != nil {
		return n, fmt.Errorf("%s %s: %w", name, p, err)
	}
	return n, nil
}

func (s *sshConn) get(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	if err := starlark.UnpackArgs("ssh.get", args, kw, "path", &p); err != nil {
		return nil, err
	}
	if _, err := remotePath("ssh.get", p); err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if _, err := s.readRemote("ssh.get", p, s.e.opts.Limits.MaxHTTPBytes, &buf); err != nil {
		return nil, err
	}
	return starlark.Bytes(buf.Bytes()), nil
}

func (s *sshConn) put(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	var data starlark.Value
	if err := starlark.UnpackArgs("ssh.put", args, kw, "path", &p, "data", &data); err != nil {
		return nil, err
	}
	if _, err := remotePath("ssh.put", p); err != nil {
		return nil, err
	}
	b, err := dataBytes(data)
	if err != nil {
		return nil, fmt.Errorf("ssh.put: %w", err)
	}
	n, err := s.writeRemote("ssh.put", p, bytes.NewReader(b), s.e.opts.Limits.MaxHTTPBytes)
	if err != nil {
		return nil, err
	}
	return starlark.MakeInt64(n), nil
}

// download copies a remote file to a local one the extension may write
// (fs.write), through a temporary file so a failed transfer leaves nothing half-written.
func (s *sshConn) download(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var remote, local string
	if err := starlark.UnpackArgs("ssh.download", args, kw, "remote", &remote, "local", &local); err != nil {
		return nil, err
	}
	if _, err := remotePath("ssh.download", remote); err != nil {
		return nil, err
	}
	dst, err := s.e.checkFS(PermFSWrite, local)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return nil, err
	}
	tmp := dst + ".part"
	f, err := os.Create(tmp)
	if err != nil {
		return nil, err
	}
	n, err := s.readRemote("ssh.download", remote, s.e.opts.Limits.MaxBulkBytes, f)
	if cerr := f.Close(); err == nil {
		err = cerr
	}
	if err == nil {
		err = os.Rename(tmp, dst)
	}
	if err != nil {
		os.Remove(tmp)
		return nil, err
	}
	return starlark.MakeInt64(n), nil
}

func (s *sshConn) upload(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var local, remote string
	if err := starlark.UnpackArgs("ssh.upload", args, kw, "local", &local, "remote", &remote); err != nil {
		return nil, err
	}
	if _, err := remotePath("ssh.upload", remote); err != nil {
		return nil, err
	}
	src, err := s.e.checkFS(PermFSRead, local)
	if err != nil {
		return nil, err
	}
	f, err := os.Open(src)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	n, err := s.writeRemote("ssh.upload", remote, f, s.e.opts.Limits.MaxBulkBytes)
	if err != nil {
		return nil, err
	}
	return starlark.MakeInt64(n), nil
}

func remoteEntry(dir, name string, fi os.FileInfo) starlark.Value {
	p := name
	if dir != "" {
		p = strings.TrimSuffix(dir, "/") + "/" + name
	}
	return starlarkstruct.FromStringDict(starlark.String("entry"), starlark.StringDict{
		"name":       starlark.String(name),
		"path":       starlark.String(p),
		"size":       starlark.MakeInt64(fi.Size()),
		"mtime":      starlark.MakeInt64(fi.ModTime().Unix()),
		"is_dir":     starlark.Bool(fi.IsDir()),
		"is_file":    starlark.Bool(fi.Mode().IsRegular()),
		"is_symlink": starlark.Bool(fi.Mode()&os.ModeSymlink != 0),
		"mode":       starlark.String(fi.Mode().Perm().String()),
	})
}

func (s *sshConn) list(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	p := "."
	if err := starlark.UnpackArgs("ssh.list", args, kw, "path?", &p); err != nil {
		return nil, err
	}
	if _, err := remotePath("ssh.list", p); err != nil {
		return nil, err
	}
	c, err := s.sftp()
	if err != nil {
		return nil, err
	}
	infos, err := c.ReadDir(p)
	if err != nil {
		return nil, fmt.Errorf("ssh.list %s: %w", p, err)
	}
	if len(infos) > s.e.opts.Limits.MaxEntries {
		return nil, fmt.Errorf("ssh.list %s: more than %d entries", p, s.e.opts.Limits.MaxEntries)
	}
	out := make([]starlark.Value, len(infos))
	for i, fi := range infos {
		out[i] = remoteEntry(p, fi.Name(), fi)
	}
	return starlark.NewList(out), nil
}

func (s *sshConn) stat(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	if err := starlark.UnpackArgs("ssh.stat", args, kw, "path", &p); err != nil {
		return nil, err
	}
	if _, err := remotePath("ssh.stat", p); err != nil {
		return nil, err
	}
	c, err := s.sftp()
	if err != nil {
		return nil, err
	}
	fi, err := c.Lstat(p)
	if err != nil {
		return nil, fmt.Errorf("ssh.stat %s: %w", p, err)
	}
	return remoteEntry(filepath.ToSlash(filepath.Dir(p)), fi.Name(), fi), nil
}

// remove deletes a file or an empty directory.
func (s *sshConn) remove(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	if err := starlark.UnpackArgs("ssh.remove", args, kw, "path", &p); err != nil {
		return nil, err
	}
	if _, err := remotePath("ssh.remove", p); err != nil {
		return nil, err
	}
	if t := strings.Trim(p, "/."); t == "" || p == "~" {
		return nil, fmt.Errorf("ssh.remove: refusing to remove %q", p)
	}
	c, err := s.sftp()
	if err != nil {
		return nil, err
	}
	if err := c.Remove(p); err != nil {
		return nil, fmt.Errorf("ssh.remove %s: %w", p, err)
	}
	return starlark.None, nil
}

func (s *sshConn) mkdir(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var p string
	parents := false
	if err := starlark.UnpackArgs("ssh.mkdir", args, kw, "path", &p, "parents?", &parents); err != nil {
		return nil, err
	}
	if _, err := remotePath("ssh.mkdir", p); err != nil {
		return nil, err
	}
	c, err := s.sftp()
	if err != nil {
		return nil, err
	}
	if parents {
		err = c.MkdirAll(p)
	} else {
		err = c.Mkdir(p)
	}
	if err != nil {
		return nil, fmt.Errorf("ssh.mkdir %s: %w", p, err)
	}
	return starlark.None, nil
}

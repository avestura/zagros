package ext

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/pem"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type testSSH struct {
	host string
	port int
	fp   string // SHA256:... without padding
	l    net.Listener
}

func newSigner(t *testing.T) ssh.Signer {
	t.Helper()
	_, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	s, err := ssh.NewSignerFromKey(priv)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

// startSSH runs an in-process server on addr ("" = any port) that accepts the
// password "pw" or the public key authorized, runs a few canned commands and
// serves sftp over the real filesystem.
func startSSH(t *testing.T, addr string, hostKey ssh.Signer, authorized ssh.PublicKey) *testSSH {
	t.Helper()
	if addr == "" {
		addr = "127.0.0.1:0"
	}
	cfg := &ssh.ServerConfig{
		PasswordCallback: func(c ssh.ConnMetadata, p []byte) (*ssh.Permissions, error) {
			if string(p) == "pw" {
				return nil, nil
			}
			return nil, fmt.Errorf("bad password")
		},
		PublicKeyCallback: func(c ssh.ConnMetadata, k ssh.PublicKey) (*ssh.Permissions, error) {
			if authorized != nil && string(k.Marshal()) == string(authorized.Marshal()) {
				return nil, nil
			}
			return nil, fmt.Errorf("unknown key")
		},
	}
	cfg.AddHostKey(hostKey)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { l.Close() })
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go serveSSH(c, cfg)
		}
	}()
	a := l.Addr().(*net.TCPAddr)
	return &testSSH{host: "127.0.0.1", port: a.Port, fp: strings.TrimRight(ssh.FingerprintSHA256(hostKey.PublicKey()), "="), l: l}
}

func serveSSH(nc net.Conn, cfg *ssh.ServerConfig) {
	defer nc.Close()
	conn, chans, reqs, err := ssh.NewServerConn(nc, cfg)
	if err != nil {
		return
	}
	defer conn.Close()
	go ssh.DiscardRequests(reqs)
	for nch := range chans {
		if nch.ChannelType() != "session" {
			nch.Reject(ssh.UnknownChannelType, "no")
			continue
		}
		ch, creqs, err := nch.Accept()
		if err != nil {
			return
		}
		go func() {
			defer ch.Close()
			for r := range creqs {
				switch r.Type {
				case "exec":
					cmd := string(r.Payload[4:])
					r.Reply(true, nil)
					code := uint32(0)
					switch {
					case cmd == "whoami":
						io.WriteString(ch, conn.User()+"\n")
					case cmd == "fail":
						io.WriteString(ch.Stderr(), "bad thing\n")
						code = 3
					case cmd == "cat":
						io.Copy(ch, ch)
					case cmd == "flood":
						chunk := []byte(strings.Repeat("z", 4096))
						for i := 0; i < 64; i++ {
							if _, err := ch.Write(chunk); err != nil {
								return
							}
						}
					case cmd == "hang":
						time.Sleep(3 * time.Second)
						return
					default:
						io.WriteString(ch, "ran: "+cmd+"\n")
					}
					ch.SendRequest("exit-status", false, ssh.Marshal(struct{ C uint32 }{code}))
					return
				case "subsystem":
					if string(r.Payload[4:]) != "sftp" {
						r.Reply(false, nil)
						continue
					}
					r.Reply(true, nil)
					srv, err := sftp.NewServer(ch)
					if err == nil {
						srv.Serve()
					}
					return
				default:
					r.Reply(false, nil)
				}
			}
		}()
	}
}

func sshPerms(s *testSSH, extra ...Permission) []Permission {
	return append([]Permission{Permission(fmt.Sprintf("ssh:alice@%s:%d", s.host, s.port)), "secrets"}, extra...)
}

func connectLine(s *testSSH, extra string) string {
	return fmt.Sprintf(`c = ssh.connect("%s", "alice", port=%d, auth={"secret": "pw"}%s)`, s.host, s.port, extra)
}

func TestSSHExecAndTOFU(t *testing.T) {
	f := useFakeSecrets(t)
	f.m[secretTarget("test.ext", "pw")] = "pw"
	srv := startSSH(t, "", newSigner(t), nil)
	data := t.TempDir()
	opts := Options{DataDir: data}

	got := runNetScript(t, sshPerms(srv), connectLine(srv, "")+`
a = c.exec("whoami")
b = c.exec("fail")
d = c.exec("cat", stdin="piped")
r = [a.stdout, a.code, a.ok, b.stderr, b.code, b.ok, d.stdout, c.host_key_status, c.fingerprint]`, opts)
	want := fmt.Sprintf(`["alice\n", 0, True, "bad thing\n", 3, False, "piped", "new", "%s"]`, srv.fp)
	if got != want {
		t.Fatalf("got  %s\nwant %s", got, want)
	}
	if _, err := os.Stat(filepath.Join(data, "ext", "test-ext", knownHostsFile)); err != nil {
		matches, _ := filepath.Glob(filepath.Join(data, "ext", "*", knownHostsFile))
		if len(matches) != 1 {
			t.Fatalf("known hosts file not written: %v", err)
		}
	}
	// second connection: the recorded key is used
	got = runNetScript(t, sshPerms(srv), connectLine(srv, "")+"\nr = c.host_key_status", opts)
	if got != `"known"` && got != "known" {
		t.Fatalf("status: %s", got)
	}

	// a different key on the same address is refused, and the error says how to proceed
	srv.l.Close()
	evil := startSSH(t, fmt.Sprintf("%s:%d", srv.host, srv.port), newSigner(t), nil)
	got = runNetScript(t, sshPerms(evil), "r = attempt(ssh.connect, \""+evil.host+"\", \"alice\", "+strconv.Itoa(evil.port)+", {\"secret\": \"pw\"})[1]", opts)
	if !strings.Contains(got, "host key") || !strings.Contains(got, "has changed") || !strings.Contains(got, evil.fp) || !strings.Contains(got, srv.fp) {
		t.Fatalf("changed key: %s", got)
	}
	// pinning the new fingerprint is the explicit way through
	got = runNetScript(t, sshPerms(evil), connectLine(evil, `, host_key="`+evil.fp+`"`)+"\nr = [c.host_key_status, c.exec('x').stdout]", opts)
	if !strings.Contains(got, "pinned") {
		t.Fatalf("pin: %s", got)
	}
}

func TestSSHPinAndNoTrustNew(t *testing.T) {
	f := useFakeSecrets(t)
	f.m[secretTarget("test.ext", "pw")] = "pw"
	srv := startSSH(t, "", newSigner(t), nil)

	got := runNetScript(t, sshPerms(srv), "r = attempt(ssh.connect, \""+srv.host+"\", \"alice\", "+strconv.Itoa(srv.port)+", {\"secret\": \"pw\"}, trust_new=False)[1]", Options{})
	if !strings.Contains(got, "not a known host") || !strings.Contains(got, srv.fp) {
		t.Fatalf("trust_new=False: %s", got)
	}
	got = runNetScript(t, sshPerms(srv), connectLine(srv, `, trust_new=False, host_key="`+strings.TrimPrefix(srv.fp, "SHA256:")+`"`)+"\nr = c.host_key_status", Options{})
	if !strings.Contains(got, "pinned") {
		t.Fatalf("pin without prefix: %s", got)
	}
	got = runNetScript(t, sshPerms(srv), "r = attempt(ssh.connect, \""+srv.host+"\", \"alice\", "+strconv.Itoa(srv.port)+", {\"secret\": \"pw\"}, host_key=[\"SHA256:AAAA\", \"BBBB\"])[1]", Options{})
	if !strings.Contains(got, "does not match host_key") || !strings.Contains(got, srv.fp) {
		t.Fatalf("wrong pin: %s", got)
	}
	// in memory: a first connect records the key and a changed one fails
	e := openStarted(t, baseManifest, "def on_query(ctx, q):\n    return []\n", sshPerms(srv), Options{})
	if err := e.sshRemember("h:22", "ssh-ed25519", "SHA256:x"); err != nil {
		t.Fatal(err)
	}
	if e.sshKnown("h:22")["ssh-ed25519"] != "SHA256:x" || len(e.sshKnown("other:22")) != 0 {
		t.Error("in-memory known hosts")
	}
}

func TestSSHPermissionAndAuth(t *testing.T) {
	f := useFakeSecrets(t)
	f.m[secretTarget("test.ext", "pw")] = "pw"
	f.m[secretTarget("test.ext", "bad")] = "nope"
	srv := startSSH(t, "", newSigner(t), nil)
	call := func(perms []Permission, auth string, extra string) string {
		return runNetScript(t, perms, fmt.Sprintf(`r = attempt(ssh.connect, "%s", "alice", %d, %s%s)[1]`, srv.host, srv.port, auth, extra), Options{})
	}
	if got := call(nil, `{"secret": "pw"}`, ""); !strings.Contains(got, "permission denied") || !strings.Contains(got, `"ssh"`) {
		t.Errorf("no grant: %s", got)
	}
	if got := call([]Permission{Permission(fmt.Sprintf("ssh:bob@%s:%d", srv.host, srv.port)), "secrets"}, `{"secret": "pw"}`, ""); !strings.Contains(got, "permission denied") {
		t.Errorf("other user: %s", got)
	}
	if got := call([]Permission{Permission(fmt.Sprintf("ssh:alice@%s:%d", srv.host, srv.port+1)), "secrets"}, `{"secret": "pw"}`, ""); !strings.Contains(got, "permission denied") {
		t.Errorf("other port: %s", got)
	}
	// network permission does not stand in for ssh
	if got := call([]Permission{"network", "secrets"}, `{"secret": "pw"}`, ""); !strings.Contains(got, "permission denied") {
		t.Errorf("network instead of ssh: %s", got)
	}
	// reading the secret needs the secrets permission as well
	if got := call([]Permission{Permission(fmt.Sprintf("ssh:alice@%s:%d", srv.host, srv.port))}, `{"secret": "pw"}`, ""); !strings.Contains(got, "permission denied") || !strings.Contains(got, "secrets") {
		t.Errorf("no secrets: %s", got)
	}
	p := sshPerms(srv)
	if got := call(p, `{"secret": "missing"}`, ""); !strings.Contains(got, "not set") {
		t.Errorf("missing secret: %s", got)
	}
	if got := call(p, `{"secret": "bad"}`, ""); !strings.Contains(got, "unable to authenticate") {
		t.Errorf("bad password: %s", got)
	}
	if got := call(p, `{"hunter2": "x"}`, ""); !strings.Contains(got, "unknown auth field") {
		t.Errorf("unknown field: %s", got)
	}
	if got := call(p, `{}`, ""); !strings.Contains(got, "auth is required") {
		t.Errorf("empty auth: %s", got)
	}
	// a raw password is accepted but never needs the secrets permission
	if got := runNetScript(t, []Permission{Permission(fmt.Sprintf("ssh:alice@%s:%d", srv.host, srv.port))},
		fmt.Sprintf(`c = ssh.connect("%s", "alice", %d, {"password": "pw"})
r = c.exec("whoami").stdout`, srv.host, srv.port), Options{}); got != `alice\n` && !strings.Contains(got, "alice") {
		t.Errorf("raw password: %s", got)
	}
}

func TestSSHKeyAuth(t *testing.T) {
	f := useFakeSecrets(t)
	_, priv, _ := ed25519.GenerateKey(rand.Reader)
	blk, err := ssh.MarshalPrivateKey(priv, "")
	if err != nil {
		t.Fatal(err)
	}
	f.m[secretTarget("test.ext", "key")] = string(pem.EncodeToMemory(blk))
	enc, err := ssh.MarshalPrivateKeyWithPassphrase(priv, "", []byte("pass"))
	if err != nil {
		t.Fatal(err)
	}
	f.m[secretTarget("test.ext", "ekey")] = string(pem.EncodeToMemory(enc))
	f.m[secretTarget("test.ext", "pp")] = "pass"
	signer, _ := ssh.NewSignerFromKey(priv)
	srv := startSSH(t, "", newSigner(t), signer.PublicKey())
	run := func(auth string) string {
		return runNetScript(t, sshPerms(srv), fmt.Sprintf(`r = attempt(ssh.connect, "%s", "alice", %d, %s)`, srv.host, srv.port, auth), Options{})
	}
	if got := run(`{"key_secret": "key"}`); !strings.Contains(got, "<ssh.Conn alice@") {
		t.Errorf("key: %s", got)
	}
	if got := run(`{"key_secret": "ekey", "passphrase_secret": "pp"}`); !strings.Contains(got, "<ssh.Conn alice@") {
		t.Errorf("encrypted key: %s", got)
	}
	if got := run(`{"key_secret": "ekey"}`); !strings.Contains(got, "passphrase_secret") {
		t.Errorf("missing passphrase: %s", got)
	}
}

func TestSSHCapsAndTimeouts(t *testing.T) {
	f := useFakeSecrets(t)
	f.m[secretTarget("test.ext", "pw")] = "pw"
	srv := startSSH(t, "", newSigner(t), nil)
	small := Options{Limits: Limits{MaxHTTPBytes: 10000}}
	got := runNetScript(t, sshPerms(srv), connectLine(srv, "")+`
r = attempt(c.exec, "flood")[1]`, small)
	if !strings.Contains(got, "stdout larger than 10000 bytes") {
		t.Errorf("flood: %s", got)
	}
	got = runNetScript(t, sshPerms(srv), connectLine(srv, "")+`
r = attempt(c.exec, "hang", 0.3)[1]`, small)
	if !strings.Contains(got, "did not finish") {
		t.Errorf("hang: %s", got)
	}
	got = runNetScript(t, sshPerms(srv), connectLine(srv, "")+`
c.close()
r = attempt(c.exec, "whoami")[1]`, small)
	if !strings.Contains(got, "closed") {
		t.Errorf("closed: %s", got)
	}
}

func TestSSHSFTP(t *testing.T) {
	f := useFakeSecrets(t)
	f.m[secretTarget("test.ext", "pw")] = "pw"
	srv := startSSH(t, "", newSigner(t), nil)
	remote := filepath.ToSlash(t.TempDir())
	localDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(localDir, "up.txt"), []byte("uploaded"), 0o644); err != nil {
		t.Fatal(err)
	}
	ld := filepath.ToSlash(localDir)
	perms := sshPerms(srv, Permission("fs.read:"+ld), Permission("fs.write:"+ld))
	got := runNetScript(t, perms, connectLine(srv, "")+`
n = c.put("`+remote+`/a.txt", "hello")
c.mkdir("`+remote+`/sub/deep", parents=True)
c.upload("`+ld+`/up.txt", "`+remote+`/sub/up.txt")
c.download("`+remote+`/a.txt", "`+ld+`/down.txt")
data = c.get("`+remote+`/a.txt")
names = sorted([e.name for e in c.list("`+remote+`")])
st = c.stat("`+remote+`/a.txt")
c.remove("`+remote+`/a.txt")
gone = attempt(c.stat, "`+remote+`/a.txt")[1] != None
r = [n, data, names, st.size, st.is_file, st.is_dir, gone]`, Options{})
	if got != `[5, b"hello", ["a.txt", "sub"], 5, True, False, True]` {
		t.Fatalf("got %s", got)
	}
	if b, _ := os.ReadFile(filepath.Join(localDir, "down.txt")); string(b) != "hello" {
		t.Errorf("download wrote %q", b)
	}
	if b, _ := os.ReadFile(filepath.Join(remote, "sub", "up.txt")); string(b) != "uploaded" {
		t.Errorf("upload wrote %q", b)
	}

	// local files need the fs permissions, as everywhere else
	noFS := sshPerms(srv)
	for _, call := range []string{`c.download("` + remote + `/sub/up.txt", "` + ld + `/x.txt")`, `c.upload("` + ld + `/up.txt", "` + remote + `/y.txt")`} {
		if got := runNetScript(t, noFS, connectLine(srv, "")+"\nr = attempt(lambda: "+call+")[1]", Options{}); !strings.Contains(got, "permission denied") {
			t.Errorf("%s: %s", call, got)
		}
	}
	// size cap on get
	big := filepath.Join(remote, "big.bin")
	os.WriteFile(big, make([]byte, 5000), 0o644)
	got = runNetScript(t, sshPerms(srv), connectLine(srv, "")+`
r = attempt(c.get, "`+remote+`/big.bin")[1]`, Options{Limits: Limits{MaxHTTPBytes: 1000}})
	if !strings.Contains(got, "larger than 1000 bytes") {
		t.Errorf("get cap: %s", got)
	}
	got = runNetScript(t, sshPerms(srv), connectLine(srv, "")+`
r = attempt(c.remove, "/")[1]`, Options{})
	if !strings.Contains(got, "refusing") {
		t.Errorf("remove root: %s", got)
	}
}

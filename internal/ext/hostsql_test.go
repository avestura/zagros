package ext

import (
	"context"
	"database/sql/driver"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"go.starlark.net/starlark"
)

// fakeDB is a database/sql driver that records what it is asked and answers
// a few canned queries, so the sql module runs without a server.
type fakeDB struct {
	mu     sync.Mutex
	log    []string
	cfgs   []sqlConfig
	closed int
}

func (f *fakeDB) record(s string, a ...any) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.log = append(f.log, fmt.Sprintf(s, a...))
}

func (f *fakeDB) lines() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return strings.Join(f.log, "\n")
}

type fakeConnector struct{ f *fakeDB }

func (c fakeConnector) Connect(context.Context) (driver.Conn, error) { return &fakeConn{c.f}, nil }
func (c fakeConnector) Driver() driver.Driver                        { return nil }

type fakeConn struct{ f *fakeDB }

func (c *fakeConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not supported") }
func (c *fakeConn) Close() error                        { c.f.mu.Lock(); c.f.closed++; c.f.mu.Unlock(); return nil }
func (c *fakeConn) Begin() (driver.Tx, error) {
	return c.BeginTx(context.Background(), driver.TxOptions{})
}
func (c *fakeConn) BeginTx(context.Context, driver.TxOptions) (driver.Tx, error) {
	c.f.record("BEGIN")
	return fakeTx{c.f}, nil
}
func (c *fakeConn) Ping(context.Context) error { return nil }

func (c *fakeConn) ExecContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Result, error) {
	c.f.record("EXEC %s %v", q, namedVals(args))
	if strings.Contains(q, "FAIL") {
		return nil, errors.New("boom")
	}
	return driver.RowsAffected(3), nil
}

func (c *fakeConn) QueryContext(ctx context.Context, q string, args []driver.NamedValue) (driver.Rows, error) {
	c.f.record("QUERY %s %v", q, namedVals(args))
	switch {
	case strings.Contains(q, "SLEEP"):
		<-ctx.Done()
		return nil, ctx.Err()
	case strings.Contains(q, "MANY"):
		return &fakeRows{n: 20000}, nil
	}
	return &fakeRows{n: 2}, nil
}

func namedVals(a []driver.NamedValue) []any {
	out := make([]any, len(a))
	for i, v := range a {
		out[i] = v.Value
	}
	return out
}

type fakeTx struct{ f *fakeDB }

func (t fakeTx) Commit() error   { t.f.record("COMMIT"); return nil }
func (t fakeTx) Rollback() error { t.f.record("ROLLBACK"); return nil }

type fakeRows struct{ n, i int }

func (r *fakeRows) Columns() []string { return []string{"id", "name", "data", "ratio", "at"} }
func (r *fakeRows) Close() error      { return nil }
func (r *fakeRows) ColumnTypeDatabaseTypeName(i int) string {
	return []string{"INT", "VARCHAR", "BLOB", "DOUBLE", "DATETIME"}[i]
}
func (r *fakeRows) Next(dest []driver.Value) error {
	if r.i >= r.n {
		return io.EOF
	}
	r.i++
	dest[0], dest[1], dest[2], dest[3] = int64(r.i), []byte("name"+fmt.Sprint(r.i)), []byte{0xff, 0x00}, 1.5
	dest[4] = time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	return nil
}

// useFakeSQL registers the drivers "fakeq" (? placeholders) and "faked" ($n).
func useFakeSQL(t *testing.T) *fakeDB {
	t.Helper()
	f := &fakeDB{}
	mk := func(style bindStyle) sqlDriver {
		return sqlDriver{defPort: 1111, style: style, textBytes: true, connector: func(c sqlConfig) (driver.Connector, error) {
			f.mu.Lock()
			f.cfgs = append(f.cfgs, c)
			f.mu.Unlock()
			if c.Host == "down.example.com" {
				return nil, errors.New("no route, tried " + c.Password)
			}
			return fakeConnector{f}, nil
		}}
	}
	sqlDrivers["fakeq"], sqlDrivers["faked"] = mk(bindQuestion), mk(bindDollar)
	t.Cleanup(func() { delete(sqlDrivers, "fakeq"); delete(sqlDrivers, "faked") })
	return f
}

func TestHostPortScopes(t *testing.T) {
	cases := []struct {
		scope, target string
		want          bool
	}{
		{"db.example.com", "db.example.com:5432", true},
		{"DB.example.com", "db.EXAMPLE.com:5432", true},
		{"example.com", "db.example.com:3306", true},
		{"db.example.com:5432", "db.example.com:5432", true},
		{"db.example.com:5432", "db.example.com:5433", false},
		{"example.com", "evilexample.com:5432", false},
		{"db.example.com", "example.com:5432", false},
		{"10.0.0.1", "10.0.0.1:389", true},
		{"0.0.0.1", "10.0.0.1:389", false},
		{"[::1]:389", "[::1]:389", true},
		{"[::1]:389", "[::1]:636", false},
		{"::1", "[::1]:389", true},
	}
	for _, c := range cases {
		if got := hostPortMatches(c.scope, c.target); got != c.want {
			t.Errorf("hostPortMatches(%q, %q) = %v, want %v", c.scope, c.target, got, c.want)
		}
	}
	for _, k := range []string{"sql", "ldap", "ldap.write"} {
		if err := Permission(k + ":db.example.com:5432").Validate(); err != nil {
			t.Errorf("%s scoped: %v", k, err)
		}
		if err := Permission(k).Validate(); err != nil {
			t.Errorf("%s unscoped: %v", k, err)
		}
		if err := Permission(k + ":").Validate(); err == nil {
			t.Errorf("%s with empty scope must fail", k)
		}
	}
	ps := PermissionSet{"sql:db.example.com:5432", "ldap:ldap.corp.example.com", "ldap.write:ldap.corp.example.com:636"}
	if !ps.Allows(PermSQL, "db.example.com:5432") || ps.Allows(PermSQL, "db.example.com:5433") || ps.Allows(PermSQL, "other:5432") {
		t.Error("sql scope")
	}
	if !ps.Allows(PermLDAP, "ldap.corp.example.com:636") || ps.Allows(PermLDAPWrite, "ldap.corp.example.com:389") || !ps.Allows(PermLDAPWrite, "ldap.corp.example.com:636") {
		t.Error("ldap scopes")
	}
	if ps.Allows(PermLDAP, "x:1") || (PermissionSet{"sql"}).Allows(PermLDAP, "x:1") || (PermissionSet{"network:db.example.com"}).Allows(PermSQL, "db.example.com:5432") {
		t.Error("kinds must not cross")
	}
}

func TestBindSQL(t *testing.T) {
	list := func(v ...starlark.Value) starlark.Value { return starlark.NewList(v) }
	dict := func(kv ...any) starlark.Value {
		d := starlark.NewDict(len(kv) / 2)
		for i := 0; i < len(kv); i += 2 {
			d.SetKey(starlark.String(kv[i].(string)), kv[i+1].(starlark.Value))
		}
		return d
	}
	S, I := func(s string) starlark.Value { return starlark.String(s) }, func(n int) starlark.Value { return starlark.MakeInt(n) }
	cases := []struct {
		name  string
		q     string
		p     starlark.Value
		style bindStyle
		want  string
		args  []any
		err   string
	}{
		{"mysql list", "SELECT * FROM t WHERE a = ? AND b = ?", list(I(1), S("x")), bindQuestion, "SELECT * FROM t WHERE a = ? AND b = ?", []any{int64(1), "x"}, ""},
		{"pg list", "SELECT * FROM t WHERE a = ? AND b = ?", list(I(1), S("x")), bindDollar, "SELECT * FROM t WHERE a = $1 AND b = $2", []any{int64(1), "x"}, ""},
		{"pg native", "SELECT $1, $2", list(I(1), I(2)), bindDollar, "SELECT $1, $2", []any{int64(1), int64(2)}, ""},
		{"quoted ? kept", "SELECT '?', \"?\", a -- ?\n FROM /* ? */ t WHERE b = ?", list(I(1)), bindDollar, "SELECT '?', \"?\", a -- ?\n FROM /* ? */ t WHERE b = $1", []any{int64(1)}, ""},
		{"pg ?? literal", "SELECT data ?? 'k' FROM t WHERE a = ?", list(I(1)), bindDollar, "SELECT data ? 'k' FROM t WHERE a = $1", []any{int64(1)}, ""},
		{"pg dollar quote", "SELECT $$a ? b$$, ?", list(I(7)), bindDollar, "SELECT $$a ? b$$, $1", []any{int64(7)}, ""},
		{"named mysql", "UPDATE t SET a = :v WHERE b = :v OR c = :w", dict("v", I(1), "w", I(2)), bindQuestion, "UPDATE t SET a = ? WHERE b = ? OR c = ?", []any{int64(1), int64(1), int64(2)}, ""},
		{"named pg reuses", "UPDATE t SET a = :v::int WHERE b = :v OR c = :w", dict("v", I(1), "w", I(2)), bindDollar, "UPDATE t SET a = $1::int WHERE b = $1 OR c = $2", []any{int64(1), int64(2)}, ""},
		{"named strips colon", "SELECT :v", dict(":v", I(1)), bindQuestion, "SELECT ?", []any{int64(1)}, ""},
		{"no params", "SELECT 1", starlark.None, bindQuestion, "SELECT 1", nil, ""},
		{"too few", "SELECT ?, ?", list(I(1)), bindQuestion, "", nil, "more ? placeholders"},
		{"too many", "SELECT ?", list(I(1), I(2)), bindQuestion, "", nil, "1 ? placeholders but 2"},
		{"missing name", "SELECT :nope", dict("v", I(1)), bindQuestion, "", nil, "no value for :nope"},
		{"mixed", "SELECT ?", dict("v", I(1)), bindQuestion, "", nil, "use :name"},
		{"bad type", "SELECT ?", list(list()), bindQuestion, "", nil, "cannot bind a list"},
		{"bad params", "SELECT ?", S("x"), bindQuestion, "", nil, "params must be"},
	}
	for _, c := range cases {
		q, args, err := bindSQL(c.q, c.p, c.style)
		if c.err != "" {
			if err == nil || !strings.Contains(err.Error(), c.err) {
				t.Errorf("%s: err = %v, want %q", c.name, err, c.err)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if q != c.want || !reflect.DeepEqual(args, c.args) {
			t.Errorf("%s: got %q %v, want %q %v", c.name, q, args, c.want, c.args)
		}
	}
}

func TestSQLRealDriversBuildConnectors(t *testing.T) {
	for _, name := range []string{"postgres", "mysql"} {
		d := sqlDrivers[name]
		c, err := d.connector(sqlConfig{Host: "db.example.com", Port: d.defPort, Database: "d b", User: "u@x", Password: "p:/?#@ss", TLS: true, Timeout: time.Second})
		if err != nil || c == nil {
			t.Errorf("%s: %v", name, err)
		}
	}
	// an unreachable server surfaces an error that does not leak the password
	_, err := runScript(t, `out = sql.open("postgres", "127.0.0.1", port=1, database="d", user="u", password="s3cret-pw", timeout=2)`,
		[]Permission{"sql:127.0.0.1"}, Options{})
	if err == nil || strings.Contains(err.Error(), "s3cret-pw") {
		t.Fatalf("want a connect error without the password, got %v", err)
	}
}

const sqlPerm = Permission("sql:db.example.com")

func TestSQLOpenNeedsPermission(t *testing.T) {
	useFakeSQL(t)
	for _, granted := range [][]Permission{
		nil,
		{"sql:other.example.com"},
		{"sql:db.example.com:1112"},
		{"network:db.example.com", "ldap:db.example.com"},
	} {
		_, err := runScript(t, `out = sql.open("fakeq", "db.example.com")`, granted, Options{})
		if err == nil || !strings.Contains(err.Error(), `requires "sql"`) {
			t.Errorf("%v: want permission error, got %v", granted, err)
		}
	}
	if _, err := runScript(t, `out = sql.open("fakeq", "db.example.com")`, []Permission{"sql:db.example.com:1111"}, Options{}); err != nil {
		t.Errorf("default port must match the scope: %v", err)
	}
	if _, err := runScript(t, `out = sql.open("oracle", "db.example.com")`, []Permission{sqlPerm}, Options{}); err == nil || !strings.Contains(err.Error(), "unknown driver") {
		t.Errorf("unknown driver: %v", err)
	}
	for _, bad := range []string{`host="a b"`, `host="db.example.com", port=70000`, `host="db.example.com", timeout=0`} {
		if _, err := runScript(t, `out = sql.open("fakeq", `+bad+`)`, []Permission{sqlPerm}, Options{}); err == nil {
			t.Errorf("%s: want error", bad)
		}
	}
}

func TestSQLQueryExecTransaction(t *testing.T) {
	f := useFakeSQL(t)
	got := mustRun(t, `db = sql.open("fakeq", "db.example.com", database="app", user="u", password="pw")
rows = db.query("SELECT id, name FROM t WHERE id > ?", [0])
r = db.exec("UPDATE t SET name = :n", {"n": "x"})
out = [rows[0]["id"], rows[1]["name"], rows[0]["data"], rows[0]["ratio"], rows[0]["at"], r.rows_affected, str(db), type(db)]`, []Permission{sqlPerm}, Options{})
	want := `[1, "name2", b"\xff\x00", 1.5, "2026-01-02T03:04:05Z", 3, "<sql.Database db.example.com:1111/app>", "sql.Database"]`
	if got != want {
		t.Fatalf("got %s\nwant %s", got, want)
	}
	log := f.lines()
	for _, s := range []string{"QUERY SELECT id, name FROM t WHERE id > ? [0]", "EXEC UPDATE t SET name = ? [x]"} {
		if !strings.Contains(log, s) {
			t.Errorf("log missing %q:\n%s", s, log)
		}
	}
	if c := f.cfgs[0]; c.Host != "db.example.com" || c.Port != 1111 || c.Database != "app" || c.User != "u" || c.Password != "pw" || !c.TLS {
		t.Errorf("config: %+v", c)
	}

	f2 := useFakeSQL(t)
	mustRun(t, `db = sql.open("faked", "db.example.com", tls=False)
db.transaction(lambda tx: tx.exec("INSERT INTO t VALUES (?)", [1]))
_, err = attempt(db.transaction, lambda tx: tx.exec("FAIL ?", [2]))
out = err`, []Permission{sqlPerm}, Options{})
	log = f2.lines()
	if !strings.Contains(log, "BEGIN\nEXEC INSERT INTO t VALUES ($1) [1]\nCOMMIT") || !strings.Contains(log, "ROLLBACK") || f2.cfgs[0].TLS {
		t.Errorf("transaction log:\n%s", log)
	}
	// the database handle is refused inside a transaction
	_, err := runScript(t, `db = sql.open("fakeq", "db.example.com")
out = db.transaction(lambda tx: db.query("SELECT 1"))`, []Permission{sqlPerm}, Options{})
	if err == nil || !strings.Contains(err.Error(), "use the tx argument") {
		t.Errorf("db inside tx: %v", err)
	}
	if _, err := runScript(t, `db = sql.open("fakeq", "db.example.com")
out = db.exec("SELECT ?", [1, 2])`, []Permission{sqlPerm}, Options{}); err == nil || !strings.Contains(err.Error(), "placeholders") {
		t.Errorf("param count: %v", err)
	}
}

func TestSQLRowCapTimeoutAndClose(t *testing.T) {
	f := useFakeSQL(t)
	_, err := runScript(t, `db = sql.open("fakeq", "db.example.com")
out = db.query("SELECT MANY")`, []Permission{sqlPerm}, Options{})
	if err == nil || !strings.Contains(err.Error(), "more than 10000 rows") {
		t.Errorf("row cap: %v", err)
	}
	start := time.Now()
	_, err = runScript(t, `db = sql.open("fakeq", "db.example.com", timeout=1)
out = db.query("SELECT SLEEP")`, []Permission{sqlPerm}, Options{})
	if err == nil || !strings.Contains(err.Error(), "deadline") || time.Since(start) > 10*time.Second {
		t.Errorf("statement timeout: %v after %v", err, time.Since(start))
	}
	// connections are closed on Stop, and an explicit close is safe to repeat
	f.mu.Lock()
	f.closed = 0
	f.mu.Unlock()
	e := openStarted(t, baseManifest, `def on_query(ctx, q):
    db = sql.open("fakeq", "db.example.com")
    db.close()
    db.close()
    sql.open("fakeq", "db.example.com")
    return [item("1", "x")]
`, []Permission{sqlPerm}, Options{})
	if _, err := e.Query(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	if e.countSQL() != 1 {
		t.Fatalf("open connections: %d", e.countSQL())
	}
	e.Stop()
	f.mu.Lock()
	closed := f.closed
	f.mu.Unlock()
	if closed < 2 || e.countSQL() != 0 {
		t.Errorf("closed %d connections, %d still tracked", closed, e.countSQL())
	}
	_, err = runScript(t, `for i in range(9):
    sql.open("fakeq", "db.example.com")
out = 1`, []Permission{sqlPerm}, Options{})
	if err == nil || !strings.Contains(err.Error(), "at most 8") {
		t.Errorf("connection cap: %v", err)
	}
}

func TestSQLPasswordFromSecret(t *testing.T) {
	f := useFakeSQL(t)
	sec := useFakeSecrets(t)
	sec.m[secretTarget("test.ext", "dbpw")] = "from-vault"
	mustRun(t, `db = sql.open("fakeq", "db.example.com", password={"secret": "dbpw"})
out = 1`, []Permission{sqlPerm, "secrets"}, Options{})
	if f.cfgs[0].Password != "from-vault" {
		t.Errorf("password: %q", f.cfgs[0].Password)
	}
	_, err := runScript(t, `out = sql.open("fakeq", "db.example.com", password={"secret": "dbpw"})`, []Permission{sqlPerm}, Options{})
	if err == nil || !strings.Contains(err.Error(), `requires "secrets"`) {
		t.Errorf("secret without permission: %v", err)
	}
	_, err = runScript(t, `out = sql.open("fakeq", "db.example.com", password={"secret": "nope"})`, []Permission{sqlPerm, "secrets"}, Options{})
	if err == nil || !strings.Contains(err.Error(), "not set") {
		t.Errorf("missing secret: %v", err)
	}
	_, err = runScript(t, `out = sql.open("fakeq", "db.example.com", password={"key": "x"})`, []Permission{sqlPerm, "secrets"}, Options{})
	if err == nil || !strings.Contains(err.Error(), `{"secret": name}`) {
		t.Errorf("bad dict: %v", err)
	}
	_, err = runScript(t, `out = sql.open("fakeq", "down.example.com", password="hunter2-pw")`, []Permission{"sql:down.example.com"}, Options{})
	if err == nil || strings.Contains(err.Error(), "hunter2-pw") {
		t.Errorf("connector error must not echo the password: %v", err)
	}
}

func TestValidateFlagsUndeclaredSQLAndLDAP(t *testing.T) {
	for needle, kind := range map[string]string{`sql.open("postgres", "h")`: "sql", `ldap.connect("ldaps://h")`: "ldap"} {
		dir := filepath.Join(t.TempDir(), "x")
		writeExt(t, dir, baseManifest, "def on_query(ctx, q):\n    "+needle+"\n    return []\n", nil)
		warned := false
		for _, i := range Validate(dir) {
			if !i.Error && strings.Contains(i.Msg, `permission "`+kind+`" is not declared`) {
				warned = true
			}
		}
		if !warned {
			t.Errorf("%s: want an undeclared-permission warning", needle)
		}
	}
}

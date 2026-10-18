package ext

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"go.starlark.net/starlark"
	"go.starlark.net/starlarkstruct"
	"modernc.org/sqlite"
)

func (e *Extension) sqliteModule() *starlarkstruct.Module {
	return module("sqlite", starlark.StringDict{"open": e.fn("sqlite.open", e.biSQLiteOpen)})
}

var dbNameRe = regexp.MustCompile(`^[A-Za-z0-9_-]{1,64}$`)

// sqliteDB is one open database of one extension: a single pinned connection
// with ATTACH disabled and a size cap, so scripts cannot reach other files.
type sqliteDB struct {
	db      *sql.DB
	conn    *sql.Conn
	max     int
	txMu    sync.Mutex // held for a whole transaction
	ownerMu sync.Mutex
	owner   *starlark.Thread
}

func (e *Extension) biSQLiteOpen(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	name := "main"
	if err := starlark.UnpackArgs("sqlite.open", args, kw, "name?", &name); err != nil {
		return nil, err
	}
	if err := e.require(PermStorage, ""); err != nil {
		return nil, err
	}
	if !dbNameRe.MatchString(name) {
		return nil, fmt.Errorf("sqlite.open: name %q must be 1-64 letters, digits, - or _", name)
	}
	e.host.mu.Lock()
	defer e.host.mu.Unlock()
	if d := e.host.dbs[name]; d != nil {
		return &sqliteHandle{db: d, ext: e, name: name}, nil
	}
	dsn := ":memory:"
	if e.dataDir != "" {
		dir := filepath.Join(e.dataDir, "db")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		dsn = filepath.Join(dir, name+".sqlite3")
	}
	d, err := e.openSQLite(callCtx(th), dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlite.open: %w", err)
	}
	if e.host.dbs == nil {
		e.host.dbs = map[string]*sqliteDB{}
	}
	e.host.dbs[name] = d
	return &sqliteHandle{db: d, ext: e, name: name}, nil
}

func (e *Extension) openSQLite(ctx context.Context, dsn string) (*sqliteDB, error) {
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	conn, err := db.Conn(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}
	fail := func(err error) (*sqliteDB, error) {
		conn.Close()
		db.Close()
		return nil, err
	}
	if _, err := sqlite.Limit(conn, 7, 0); err != nil { // SQLITE_LIMIT_ATTACHED: no ATTACH
		return fail(err)
	}
	var pageSize int
	if err := conn.QueryRowContext(ctx, "PRAGMA page_size").Scan(&pageSize); err != nil || pageSize <= 0 {
		pageSize = 4096
	}
	pages := int(e.opts.Limits.MaxDBBytes) / pageSize
	for _, p := range []string{"PRAGMA busy_timeout = 5000", "PRAGMA foreign_keys = ON", "PRAGMA trusted_schema = OFF", fmt.Sprintf("PRAGMA max_page_count = %d", pages)} {
		if _, err := conn.ExecContext(ctx, p); err != nil {
			return fail(err)
		}
	}
	return &sqliteDB{db: db, conn: conn, max: e.opts.Limits.MaxEntries}, nil
}

func (e *Extension) closeDBs() {
	e.host.mu.Lock()
	dbs := e.host.dbs
	e.host.dbs = nil
	e.host.mu.Unlock()
	for _, d := range dbs {
		d.close()
	}
}

func (d *sqliteDB) close() {
	d.txMu.Lock()
	defer d.txMu.Unlock()
	d.conn.Close()
	d.db.Close()
}

// sqliteHandle is what scripts hold: the database (tx == nil) or a
// transaction passed to the callback of transaction().
type sqliteHandle struct {
	db   *sqliteDB
	ext  *Extension
	name string
	tx   *sql.Tx
}

var _ starlark.HasAttrs = (*sqliteHandle)(nil)

func (h *sqliteHandle) String() string {
	if h.tx != nil {
		return "<sqlite.Transaction>"
	}
	return "<sqlite.Database " + h.name + ">"
}
func (h *sqliteHandle) Type() string {
	if h.tx != nil {
		return "sqlite.Transaction"
	}
	return "sqlite.Database"
}
func (h *sqliteHandle) Freeze()               {}
func (h *sqliteHandle) Truth() starlark.Bool  { return true }
func (h *sqliteHandle) Hash() (uint32, error) { return 0, fmt.Errorf("unhashable type: %s", h.Type()) }

func (h *sqliteHandle) AttrNames() []string {
	if h.tx != nil {
		return []string{"exec", "query"}
	}
	return []string{"close", "exec", "query", "transaction"}
}

func (h *sqliteHandle) Attr(name string) (starlark.Value, error) {
	switch name {
	case "query":
		return starlark.NewBuiltin("sqlite.query", h.query), nil
	case "exec":
		return starlark.NewBuiltin("sqlite.exec", h.exec), nil
	}
	if h.tx == nil {
		switch name {
		case "transaction":
			return starlark.NewBuiltin("sqlite.transaction", h.transaction), nil
		case "close":
			return starlark.NewBuiltin("sqlite.close", h.close), nil
		}
	}
	return nil, nil
}

type sqlRunner interface {
	QueryContext(ctx context.Context, q string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, q string, args ...any) (sql.Result, error)
}

// enter picks the runner and holds the transaction lock for plain calls.
func (h *sqliteHandle) enter(th *starlark.Thread) (sqlRunner, func(), error) {
	if h.tx != nil {
		return h.tx, func() {}, nil
	}
	h.db.ownerMu.Lock()
	inside := h.db.owner == th
	h.db.ownerMu.Unlock()
	if inside {
		return nil, nil, errors.New("inside transaction() use the tx argument, not the database")
	}
	h.db.txMu.Lock()
	return h.db.conn, h.db.txMu.Unlock, nil
}

func (h *sqliteHandle) query(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	q, params, err := h.sqlArgs("query", args, kw)
	if err != nil {
		return nil, err
	}
	run, leave, err := h.enter(th)
	if err != nil {
		return nil, err
	}
	defer leave()
	rows, err := run.QueryContext(callCtx(th), q, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}
	var out []starlark.Value
	for rows.Next() {
		if len(out) >= h.db.max {
			return nil, fmt.Errorf("more than %d rows (add a LIMIT)", h.db.max)
		}
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		d := starlark.NewDict(len(cols))
		for i, c := range cols {
			if err := d.SetKey(starlark.String(c), sqlToStarlark(vals[i])); err != nil {
				return nil, err
			}
		}
		out = append(out, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return starlark.NewList(out), nil
}

func (h *sqliteHandle) exec(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	q, params, err := h.sqlArgs("exec", args, kw)
	if err != nil {
		return nil, err
	}
	run, leave, err := h.enter(th)
	if err != nil {
		return nil, err
	}
	defer leave()
	res, err := run.ExecContext(callCtx(th), q, params...)
	if err != nil {
		return nil, err
	}
	n, _ := res.RowsAffected()
	id, _ := res.LastInsertId()
	return starlarkstruct.FromStringDict(starlark.String("result"), starlark.StringDict{
		"rows_affected": starlark.MakeInt64(n), "last_insert_id": starlark.MakeInt64(id),
	}), nil
}

func (h *sqliteHandle) transaction(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var fv starlark.Value
	if err := starlark.UnpackArgs("sqlite.transaction", args, kw, "fn", &fv); err != nil {
		return nil, err
	}
	fn, ok := fv.(starlark.Callable)
	if !ok {
		return nil, fmt.Errorf("sqlite.transaction: want a function, got %s", fv.Type())
	}
	h.db.ownerMu.Lock()
	nested := h.db.owner == th
	h.db.ownerMu.Unlock()
	if nested {
		return nil, errors.New("transactions cannot be nested")
	}
	h.db.txMu.Lock()
	defer h.db.txMu.Unlock()
	tx, err := h.db.conn.BeginTx(callCtx(th), nil)
	if err != nil {
		return nil, err
	}
	h.db.ownerMu.Lock()
	h.db.owner = th
	h.db.ownerMu.Unlock()
	defer func() {
		h.db.ownerMu.Lock()
		h.db.owner = nil
		h.db.ownerMu.Unlock()
	}()
	res, err := starlark.Call(th, fn, starlark.Tuple{&sqliteHandle{db: h.db, ext: h.ext, name: h.name, tx: tx}}, nil)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return res, nil
}

func (h *sqliteHandle) close(th *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	if err := starlark.UnpackArgs("sqlite.close", args, kw); err != nil {
		return nil, err
	}
	e := h.ext
	e.host.mu.Lock()
	if e.host.dbs[h.name] == h.db {
		delete(e.host.dbs, h.name)
		e.host.mu.Unlock()
		h.db.close()
		return starlark.None, nil
	}
	e.host.mu.Unlock()
	return starlark.None, nil
}

func (h *sqliteHandle) sqlArgs(name string, args starlark.Tuple, kw []starlark.Tuple) (string, []any, error) {
	var q string
	var p starlark.Value = starlark.None
	if err := starlark.UnpackArgs("sqlite."+name, args, kw, "sql", &q, "params?", &p); err != nil {
		return "", nil, err
	}
	if err := sqlGuard(q); err != nil {
		return "", nil, err
	}
	params, err := sqlParams(p)
	return q, params, err
}

func sqlParams(p starlark.Value) ([]any, error) {
	switch x := p.(type) {
	case starlark.NoneType:
		return nil, nil
	case *starlark.Dict:
		var out []any
		for _, kv := range x.Items() {
			k, ok := starlark.AsString(kv[0])
			if !ok {
				return nil, errors.New("named parameters need string keys")
			}
			v, err := sqlValue(kv[1])
			if err != nil {
				return nil, err
			}
			out = append(out, sql.Named(strings.TrimLeft(k, ":@$"), v))
		}
		return out, nil
	case starlark.Indexable:
		var out []any
		for i := 0; i < x.Len(); i++ {
			v, err := sqlValue(x.Index(i))
			if err != nil {
				return nil, err
			}
			out = append(out, v)
		}
		return out, nil
	}
	return nil, fmt.Errorf("params must be a list, tuple or dict, got %s", p.Type())
}

func sqlValue(v starlark.Value) (any, error) {
	switch x := v.(type) {
	case starlark.NoneType:
		return nil, nil
	case starlark.Bool:
		return bool(x), nil
	case starlark.Int:
		n, ok := x.Int64()
		if !ok {
			return nil, errors.New("integer too large for sqlite (64-bit)")
		}
		return n, nil
	case starlark.Float:
		return float64(x), nil
	case starlark.String:
		return string(x), nil
	case starlark.Bytes:
		return []byte(x), nil
	}
	return nil, fmt.Errorf("cannot bind a %s as an sql parameter", v.Type())
}

func sqlToStarlark(v any) starlark.Value {
	switch x := v.(type) {
	case nil:
		return starlark.None
	case int64:
		return starlark.MakeInt64(x)
	case float64:
		return starlark.Float(x)
	case bool:
		return starlark.Bool(x)
	case string:
		return starlark.String(x)
	case []byte:
		return starlark.Bytes(x)
	case time.Time:
		return starlark.String(x.Format(time.RFC3339Nano))
	}
	return starlark.String(fmt.Sprint(v))
}

var safePragmas = map[string]bool{
	"foreign_keys": true, "user_version": true, "application_id": true, "table_info": true, "table_xinfo": true, "table_list": true,
	"index_list": true, "index_info": true, "index_xinfo": true, "foreign_key_list": true, "integrity_check": true, "quick_check": true,
	"foreign_key_check": true, "journal_mode": true, "synchronous": true, "page_count": true, "freelist_count": true, "optimize": true,
	"wal_checkpoint": true, "encoding": true, "recursive_triggers": true, "busy_timeout": true, "secure_delete": true,
	"case_sensitive_like": true, "defer_foreign_keys": true, "schema_version": true, "data_version": true, "collation_list": true,
}

// sqlGuard rejects the statements that could reach outside the extension's
// own database file: ATTACH, VACUUM INTO, load_extension and any PRAGMA that
// is not on the safe list (the size cap and trusted_schema stay fixed).
func sqlGuard(q string) error {
	for _, st := range sqlStatements(q) {
		if len(st) == 0 {
			continue
		}
		for _, w := range st {
			switch w {
			case "ATTACH", "LOAD_EXTENSION":
				return fmt.Errorf("sqlite: %s is not allowed", strings.ToLower(w))
			}
		}
		switch st[0] {
		case "VACUUM":
			for _, w := range st[1:] {
				if w == "INTO" {
					return errors.New("sqlite: VACUUM INTO is not allowed")
				}
			}
		case "PRAGMA":
			if len(st) < 2 {
				return errors.New("sqlite: malformed PRAGMA")
			}
			n := strings.ToLower(st[1])
			if i := strings.LastIndexByte(n, '.'); i >= 0 {
				n = n[i+1:]
			}
			if !safePragmas[n] {
				return fmt.Errorf("sqlite: PRAGMA %s is not allowed", n)
			}
		}
	}
	return nil
}

func isWordByte(c byte) bool {
	return c == '_' || c == '.' || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
}

// sqlStatements splits q into statements of upper-cased words, skipping
// string literals, quoted identifiers and comments.
func sqlStatements(q string) [][]string {
	var out [][]string
	var cur []string
	flush := func() { out = append(out, cur); cur = nil }
	for i := 0; i < len(q); {
		c := q[i]
		switch {
		case c == '-' && i+1 < len(q) && q[i+1] == '-':
			for i < len(q) && q[i] != '\n' {
				i++
			}
		case c == '/' && i+1 < len(q) && q[i+1] == '*':
			j := strings.Index(q[i+2:], "*/")
			if j < 0 {
				i = len(q)
			} else {
				i += j + 4
			}
		case c == '\'' || c == '"' || c == '`':
			i++
			for i < len(q) {
				if q[i] == c {
					if i+1 < len(q) && q[i+1] == c {
						i += 2
						continue
					}
					break
				}
				i++
			}
			i++
		case c == '[':
			for i < len(q) && q[i] != ']' {
				i++
			}
			i++
		case c == ';':
			flush()
			i++
		case isWordByte(c):
			j := i
			for j < len(q) && isWordByte(q[j]) {
				j++
			}
			cur = append(cur, strings.ToUpper(q[i:j]))
			i = j
		default:
			i++
		}
	}
	flush()
	return out
}

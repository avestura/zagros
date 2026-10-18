package ext

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net"
	"net/url"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"go.starlark.net/starlark"
)

// bindStyle is how a driver spells query parameters.
type bindStyle int

const (
	bindQuestion bindStyle = iota // ? (mysql)
	bindDollar                    // $1 (postgres)
)

// sqlConfig is what sql.open collects before a driver builds a connector.
type sqlConfig struct {
	Host, Database, User, Password string
	Port                           int
	TLS                            bool
	Timeout                        time.Duration
}

type sqlDriver struct {
	defPort   int
	style     bindStyle
	textBytes bool
	connector func(sqlConfig) (driver.Connector, error)
}

// sqlDrivers is the closed list a script can name; tests add a fake.
var sqlDrivers = map[string]sqlDriver{
	"postgres": {5432, bindDollar, false, postgresConnector},
	"mysql":    {3306, bindQuestion, true, mysqlConnector},
}

const maxSQLConns = 8

func postgresConnector(c sqlConfig) (driver.Connector, error) {
	u := url.URL{Scheme: "postgres", Host: net.JoinHostPort(c.Host, strconv.Itoa(c.Port)), Path: "/" + c.Database}
	if c.User != "" {
		u.User = url.UserPassword(c.User, c.Password)
	}
	mode := "verify-full"
	if !c.TLS {
		mode = "disable"
	}
	q := url.Values{}
	q.Set("sslmode", mode)
	q.Set("connect_timeout", strconv.Itoa(int(c.Timeout.Seconds())))
	q.Set("passfile", os.DevNull) // never fall back to the user's .pgpass
	u.RawQuery = q.Encode()
	cfg, err := pgx.ParseConfig(u.String())
	if err != nil {
		return nil, errors.New("invalid connection settings")
	}
	return stdlib.GetConnector(*cfg), nil
}

func mysqlConnector(c sqlConfig) (driver.Connector, error) {
	cfg := mysql.NewConfig()
	cfg.User, cfg.Passwd = c.User, c.Password
	cfg.Net = "tcp"
	cfg.Addr = net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
	cfg.DBName = c.Database
	cfg.Timeout = c.Timeout
	cfg.ParseTime = true
	cfg.TLSConfig = "false"
	if c.TLS {
		cfg.TLSConfig = "true" // verifies the certificate chain and host name
	}
	return mysql.NewConnector(cfg)
}

func (e *Extension) sqlModule() starlark.Value {
	return module("sql", starlark.StringDict{"open": e.fn("sql.open", e.biSQLOpen)})
}

func (e *Extension) biSQLOpen(th *starlark.Thread, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	var drv, host, dbname, user string
	var port int
	var pw starlark.Value = starlark.None
	tlsOn := true
	timeout := 30
	if err := starlark.UnpackArgs("sql.open", args, kw, "driver", &drv, "host", &host, "port?", &port, "database?", &dbname,
		"user?", &user, "password?", &pw, "tls?", &tlsOn, "timeout?", &timeout); err != nil {
		return nil, err
	}
	d, ok := sqlDrivers[drv]
	if !ok {
		return nil, fmt.Errorf("sql.open: unknown driver %q (use %s)", drv, strings.Join(sqlDriverNames(), " or "))
	}
	if err := validHost("sql.open", host); err != nil {
		return nil, err
	}
	if port == 0 {
		port = d.defPort
	}
	if port < 1 || port > 65535 {
		return nil, fmt.Errorf("sql.open: port %d out of range", port)
	}
	if timeout < 1 || timeout > 300 {
		return nil, errors.New("sql.open: timeout must be 1 to 300 seconds")
	}
	if err := e.require(PermSQL, net.JoinHostPort(host, strconv.Itoa(port))); err != nil {
		return nil, err
	}
	password, err := e.secretValue("sql.open", pw)
	if err != nil {
		return nil, err
	}
	if n := e.countSQL(); n >= maxSQLConns {
		return nil, fmt.Errorf("sql.open: at most %d connections per extension; close one first", maxSQLConns)
	}
	to := time.Duration(timeout) * time.Second
	conn, err := d.connector(sqlConfig{Host: host, Port: port, Database: dbname, User: user, Password: password, TLS: tlsOn, Timeout: to})
	if err != nil {
		return nil, fmt.Errorf("sql.open: %w", scrubSecret(err, password))
	}
	db := sql.OpenDB(conn)
	db.SetMaxOpenConns(1)
	ctx, cancel := context.WithTimeout(callCtx(th), to)
	defer cancel()
	c, err := db.Conn(ctx)
	if err == nil {
		err = c.PingContext(ctx)
		if err != nil {
			c.Close()
		}
	}
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("sql.open: %s:%d: %w", host, port, scrubSecret(err, password))
	}
	sd := &sqliteDB{db: db, conn: c, max: e.opts.Limits.MaxEntries, label: "sql", timeout: to, style: d.style, textBytes: d.textBytes}
	e.host.mu.Lock()
	if e.host.sqls == nil {
		e.host.sqls = map[*sqliteDB]bool{}
	}
	e.host.sqls[sd] = true
	e.host.mu.Unlock()
	return &sqliteHandle{db: sd, ext: e, name: net.JoinHostPort(host, strconv.Itoa(port)) + "/" + dbname}, nil
}

func sqlDriverNames() []string {
	var n []string
	for k := range sqlDrivers {
		n = append(n, strconv.Quote(k))
	}
	sort.Strings(n)
	return n
}

func validHost(fn, host string) error {
	if host == "" || strings.ContainsAny(host, " \t\r\n/\\@?#") {
		return fmt.Errorf("%s: invalid host %q", fn, host)
	}
	return nil
}

// scrubSecret keeps a password out of an error message.
func scrubSecret(err error, secret string) error {
	if secret == "" || !strings.Contains(err.Error(), secret) {
		return err
	}
	return errors.New(strings.ReplaceAll(err.Error(), secret, "***"))
}

func (e *Extension) countSQL() int {
	e.host.mu.Lock()
	defer e.host.mu.Unlock()
	return len(e.host.sqls)
}

func (e *Extension) dropSQL(d *sqliteDB) {
	e.host.mu.Lock()
	_, ok := e.host.sqls[d]
	delete(e.host.sqls, d)
	e.host.mu.Unlock()
	if ok {
		d.close()
	}
}

func (e *Extension) closeSQL() {
	e.host.mu.Lock()
	all := e.host.sqls
	e.host.sqls = nil
	e.host.mu.Unlock()
	for d := range all {
		d.close()
	}
}

// secretValue turns a script's password argument into a string: a plain
// string, None, or {"secret": name} read from the extension's own secrets.
func (e *Extension) secretValue(fn string, v starlark.Value) (string, error) {
	switch x := v.(type) {
	case starlark.NoneType:
		return "", nil
	case starlark.String:
		return string(x), nil
	case *starlark.Dict:
		if x.Len() != 1 {
			return "", fmt.Errorf(`%s: password dict must be {"secret": name}`, fn)
		}
		nv, ok, _ := x.Get(starlark.String("secret"))
		name, isStr := starlark.AsString(nv)
		if !ok || !isStr {
			return "", fmt.Errorf(`%s: password dict must be {"secret": name}`, fn)
		}
		if err := e.require(PermSecrets, ""); err != nil {
			return "", err
		}
		if err := secretKey(name); err != nil {
			return "", err
		}
		s, found, err := secretBackend.Get(secretTarget(e.Manifest.ID, name))
		if err != nil {
			return "", fmt.Errorf("%s: secret %q: %w", fn, name, err)
		}
		if !found {
			return "", fmt.Errorf("%s: secret %q is not set", fn, name)
		}
		return s, nil
	}
	return "", fmt.Errorf("%s: password must be a string or {\"secret\": name}, got %s", fn, v.Type())
}

// textColumns marks the columns whose bytes are text, not blobs.
func textColumns(rows *sql.Rows) []bool {
	ts, err := rows.ColumnTypes()
	if err != nil {
		return nil
	}
	out := make([]bool, len(ts))
	for i, t := range ts {
		n := strings.ToUpper(t.DatabaseTypeName())
		out[i] = !strings.Contains(n, "BLOB") && !strings.Contains(n, "BINARY") && !strings.Contains(n, "BYTEA")
	}
	return out
}

// bindSQL turns the script's query and params into what the driver wants.
// Placeholders are ? with a list or :name with a dict; for postgres they are
// rewritten to $1.. (and ?? stands for a literal ?). A query that already uses
// $1 with a list passes through untouched. Quoted text and comments are skipped.
func bindSQL(q string, p starlark.Value, style bindStyle) (string, []any, error) {
	var list []any
	var named map[string]any
	switch x := p.(type) {
	case starlark.NoneType:
	case *starlark.Dict:
		named = map[string]any{}
		for _, kv := range x.Items() {
			k, ok := starlark.AsString(kv[0])
			if !ok {
				return "", nil, errors.New("named parameters need string keys")
			}
			v, err := sqlValue(kv[1])
			if err != nil {
				return "", nil, err
			}
			named[strings.TrimLeft(k, ":@$")] = v
		}
	case starlark.String, starlark.Bytes:
		return "", nil, fmt.Errorf("params must be a list, tuple or dict, got %s", p.Type())
	case starlark.Indexable:
		for i := 0; i < x.Len(); i++ {
			v, err := sqlValue(x.Index(i))
			if err != nil {
				return "", nil, err
			}
			list = append(list, v)
		}
	default:
		return "", nil, fmt.Errorf("params must be a list, tuple or dict, got %s", p.Type())
	}

	var out strings.Builder
	var args []any
	index := map[string]int{}
	marks := 0
	emit := func(v any, name string) {
		if style == bindDollar {
			n, seen := index[name]
			if !seen || name == "" {
				args = append(args, v)
				n = len(args)
				index[name] = n
			}
			out.WriteString("$" + strconv.Itoa(n))
			return
		}
		args = append(args, v)
		out.WriteByte('?')
	}
	for i := 0; i < len(q); {
		c := q[i]
		switch {
		case c == '-' && i+1 < len(q) && q[i+1] == '-':
			j := strings.IndexByte(q[i:], '\n')
			if j < 0 {
				j = len(q) - i
			}
			out.WriteString(q[i : i+j])
			i += j
		case c == '/' && i+1 < len(q) && q[i+1] == '*':
			j := strings.Index(q[i+2:], "*/")
			if j < 0 {
				j = len(q) - i - 2
			} else {
				j += 2
			}
			out.WriteString(q[i : i+2+j])
			i += 2 + j
		case c == '\'' || c == '"' || c == '`':
			j := i + 1
			for j < len(q) {
				if q[j] == '\\' && c == '\'' && style == bindQuestion {
					j += 2
					continue
				}
				if q[j] == c {
					if j+1 < len(q) && q[j+1] == c {
						j += 2
						continue
					}
					break
				}
				j++
			}
			if j >= len(q) {
				j = len(q) - 1
			}
			out.WriteString(q[i : j+1])
			i = j + 1
		case c == '$' && style == bindDollar && i+1 < len(q) && (q[i+1] == '$' || q[i+1] == '_' || isAlpha(q[i+1])):
			j := i + 1
			for j < len(q) && (q[j] == '_' || isAlpha(q[j]) || (q[j] >= '0' && q[j] <= '9')) {
				j++
			}
			if j < len(q) && q[j] == '$' {
				tag := q[i : j+1]
				end := strings.Index(q[j+1:], tag)
				if end < 0 {
					end = len(q) - j - 1
				} else {
					end += len(tag)
				}
				out.WriteString(q[i : j+1+end])
				i = j + 1 + end
				break
			}
			out.WriteByte(c)
			i++
		case c == '?':
			if style == bindDollar && i+1 < len(q) && q[i+1] == '?' {
				out.WriteByte('?')
				i += 2
				break
			}
			if named != nil {
				return "", nil, errors.New("use :name placeholders with a dict, not ?")
			}
			if marks >= len(list) {
				return "", nil, fmt.Errorf("query has more ? placeholders than the %d parameters given", len(list))
			}
			emit(list[marks], "")
			marks++
			i++
		case c == ':' && named != nil && (i == 0 || q[i-1] != ':') && i+1 < len(q) && (q[i+1] == '_' || isAlpha(q[i+1])):
			j := i + 1
			for j < len(q) && (q[j] == '_' || isAlpha(q[j]) || (q[j] >= '0' && q[j] <= '9')) {
				j++
			}
			name := q[i+1 : j]
			v, ok := named[name]
			if !ok {
				return "", nil, fmt.Errorf("no value for :%s in params", name)
			}
			emit(v, name)
			i = j
		default:
			out.WriteByte(c)
			i++
		}
	}
	if named == nil {
		if marks == 0 && len(list) > 0 && style == bindDollar && strings.Contains(q, "$") {
			return q, list, nil
		}
		if marks != len(list) {
			return "", nil, fmt.Errorf("query has %d ? placeholders but %d parameters were given", marks, len(list))
		}
	}
	return out.String(), args, nil
}

func isAlpha(c byte) bool { return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') }

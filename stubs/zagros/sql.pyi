"""sql: talk to a PostgreSQL or MySQL/MariaDB server. Same shape as sqlite.

Type description, not runnable Starlark: do not `import` anything. The name `sql` is predeclared in main.star; call `sql.open()` directly.
Needs: sql, scoped by server host[:port] (for example `sql:db.example.com:5432`); a password given as {"secret": name} also needs secrets.
Keywords: database, sql, postgres, postgresql, mysql, mariadb, remote database, query server, table, insert, select, transaction, connect database

Drivers are "postgres" and "mysql". `db.query`, `db.exec`, `db.transaction` and
`db.close` behave like the `sqlite` module: rows come back as dicts keyed by
column name, `exec` returns rows_affected, SQL errors raise (catch with
`attempt(db.exec, sql)`). Placeholders are `?` with a list or `:name` with a
dict, for both servers (postgres `$1` also works with a list; write `??` for a
literal `?`). Never build SQL with string formatting.

Value mapping: NULL is None; integers, floats and booleans map directly; text
is str; binary columns (BLOB, BYTEA, BINARY) are bytes; dates and times are
RFC 3339 strings; postgres numeric and mysql decimal are strings, so no
precision is lost. Only int, float, str, bool, bytes and None can be bound.

Limits: TLS is on by default and the server certificate is verified against
the system roots; there is no option to skip verification (pass tls=False only
for a server you trust on a private network). Each statement runs under the
`timeout` given to open (default 30 s, at most 300 s). A query returns at most
10000 rows (add LIMIT). An extension holds at most 8 connections at once; they
are closed when the extension stops. The connection is a single session: after
a network drop, close it and open a new one.
"""

from typing import Any, Callable, Dict, List, Optional, Union

Params = Union[List[Any], Dict[str, Any], None]
Password = Union[str, Dict[str, str], None]

class Result:
    rows_affected: int
    last_insert_id: int  # mysql only; 0 on postgres (use RETURNING with query)

class Transaction:
    def query(self, sql: str, params: Params = None) -> List[Dict[str, Any]]: ...
    def exec(self, sql: str, params: Params = None) -> Result: ...

class Database:
    def query(self, sql: str, params: Params = None) -> List[Dict[str, Any]]:
        """Run a SELECT (or anything that returns rows) and return a list of dicts.

        Needs: sql
        Example: rows = db.query("SELECT id, email FROM users WHERE active = ? LIMIT 50", [True])
        Task: read rows from a PostgreSQL or MySQL server
        Keywords: select, sql query, fetch rows, read table, search records, postgres query, mysql query
        """
    def exec(self, sql: str, params: Params = None) -> Result:
        """Run INSERT, UPDATE, DELETE, DDL and similar; returns rows_affected.

        Needs: sql
        Example: db.exec("UPDATE users SET active = :a WHERE id = :id", {"a": False, "id": 7})
        Task: change data on a PostgreSQL or MySQL server
        Keywords: insert, update, delete, create table, write rows, sql execute, migrate schema
        """
    def transaction(self, fn: Callable[[Transaction], Any]) -> Any:
        """Call fn(tx); commit when it returns (its value is returned), roll back when it raises. Inside fn use tx, not the database.

        Needs: sql
        Example: db.transaction(lambda tx: tx.exec("UPDATE accounts SET balance = balance - ? WHERE id = ?", [10, 1]))
        Task: run several changes atomically on a server database
        Keywords: begin commit rollback, atomic, all or nothing, unit of work
        """
    def close(self) -> None:
        """Close the connection early. Connections are also closed when the extension stops.

        Needs: none
        Example: db.close()
        """

def open(driver: str, host: str, port: int = 0, database: str = "", user: str = "", password: Password = None, tls: bool = True, timeout: int = 30) -> Database:
    """Connect to a PostgreSQL ("postgres", default port 5432) or MySQL/MariaDB ("mysql", 3306) server.

    `password` is a string or {"secret": name} to read it from the `secrets` module (keeps it out of the source and out of settings).
    Needs: sql scoped to the host (and port when the scope has one)
    Example: db = sql.open("postgres", "db.example.com", database="app", user="reader", password={"secret": "db_password"})
    Task: connect to a PostgreSQL or MySQL database
    Keywords: sql open, connect database, postgres connect, mysql connect, mariadb, remote database, database login, connection string
    """

"""sqlite: a private SQL database per extension (pure Go SQLite, no setup).

Type description, not runnable Starlark: do not `import` anything. The name `sqlite` is predeclared in main.star; call `sqlite.open()` directly.
Needs: storage, which is always granted. Databases live in <DataDir>/ext/<id>/db/<name>.sqlite3 and survive updates; there is no way to open another path.
Keywords: database, sql, sqlite, table, query, insert, select, transaction, persistent data, records, db

Limits: one database is capped at 64 MiB, a query returns at most 10000 rows
(add LIMIT), ATTACH, VACUUM INTO and most PRAGMAs are refused (foreign_keys,
user_version, table_info, journal_mode and other read-style ones work). Use
`?` placeholders with a list, or `:name` with a dict, never string formatting.
SQL errors raise; catch with `attempt(db.exec, sql)`. Values: None, bool, int,
float, str and bytes map to NULL, integer, real, text and blob; rows come back
as dicts keyed by column name.
"""

from typing import Any, Callable, Dict, List, Union

Params = Union[List[Any], Dict[str, Any], None]

class Result:
    rows_affected: int
    last_insert_id: int

class Transaction:
    def query(self, sql: str, params: Params = None) -> List[Dict[str, Any]]: ...
    def exec(self, sql: str, params: Params = None) -> Result: ...

class Database:
    def query(self, sql: str, params: Params = None) -> List[Dict[str, Any]]:
        """Run a SELECT (or anything that returns rows) and return a list of dicts.

        Example: rows = db.query("SELECT id, title FROM notes WHERE title LIKE ?", ["%todo%"])
        Task: read rows from the database
        Keywords: select, sql query, fetch rows, read table, search records
        Needs: none
        """
    def exec(self, sql: str, params: Params = None) -> Result:
        """Run INSERT, UPDATE, DELETE, CREATE TABLE and similar; returns rows_affected and last_insert_id.

        Example: db.exec("INSERT INTO notes(title) VALUES (:t)", {"t": "buy milk"})
        Task: change the database
        Keywords: insert, update, delete, create table, write rows, sql execute, migrate schema
        Needs: none
        """
    def transaction(self, fn: Callable[[Transaction], Any]) -> Any:
        """Call fn(tx); commit when it returns (its value is returned), roll back when it raises. Inside fn use tx, not the database.

        Example: db.transaction(lambda tx: tx.exec("UPDATE notes SET done = 1 WHERE id = ?", [1]))
        Task: run several changes atomically
        Keywords: begin commit rollback, atomic, all or nothing, unit of work
        Needs: none
        """
    def close(self) -> None:
        """Close the database early; reopen with sqlite.open. Databases are also closed when the extension stops.

        Example: db.close()
        Needs: none
        """

def open(name: str = "main") -> Database:
    """Open (creating when needed) the database called name: 1-64 letters, digits, - or _.

    Example: db = sqlite.open("notes")
    Task: open the extension's database
    Keywords: sqlite open, database, connect, create database, persistent store, local db
    Needs: none
    """

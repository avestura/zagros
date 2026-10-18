---
status: accepted
date: 2026-10-19
---
# ADR 0021: Server databases and directories as host-scoped permissions

## Context
Extensions had a private SQLite file and HTTP, but no way to query a company PostgreSQL or MySQL server or look people up in LDAP and Active Directory. Both reach other machines with credentials, so they need consent, and a database connection is more than a network request: it can run any statement the account allows.

## Decision
- Three new permission kinds, all scoped by `host[:port]`: `sql`, `ldap` and `ldap.write`. A scope without a port covers every port, a host also covers its subdomains, and an IP address matches only itself. The check runs inside `sql.open` and `ldap.connect` on the host and port that will be dialled, and again on each LDAP operation.
- LDAP reads and writes are separate kinds. A lookup extension should not hold write access, and `ldap.write` is easy to show as its own line in the consent dialog. `ldap.write` can connect and modify but cannot search. SQL has one kind: a read-only grant cannot be enforced from the client because the server decides what a statement does, so the account on the server is the real limit and the docs say so.
- Drivers: `jackc/pgx/v5` (its `database/sql` adapter), `go-sql-driver/mysql` and `go-ldap/ldap/v3`, all pure Go. The script cannot name a driver outside the closed list `postgres` and `mysql`.
- The `sql` handle reuses the `sqlite` handle code (query, exec, transaction, close, the row cap, the single pinned connection and the transaction ownership rules), with a per-statement timeout and a placeholder rewriter. Scripts write `?` or `:name` for both servers; the host turns them into `$1..` for PostgreSQL, skipping quoted text and comments. One shape to learn.
- TLS is on by default, certificates are verified against the system roots, and there is no option to skip verification. `tls=False` is allowed for SQL on a private network. LDAP uses `ldaps://` or `start_tls=True`, and a bind password is refused over plain `ldap://` unless the host is loopback. Connection settings are built from fields, never from a script-supplied connection string, and the PostgreSQL driver is told not to read the user's `.pgpass`.
- Passwords are a string or `{"secret": name}`, resolved from the extension's own entries in `secrets`, which then needs the `secrets` permission. Error messages have the password scrubbed.
- Limits: 8 connections per extension, closed on stop; 10000 rows per query and per LDAP search; timeouts of 1 to 300 seconds. `ldap.check_bind` uses a second connection so the main one keeps its identity, and refuses an empty password, which servers treat as an anonymous bind.
- `ldap.escape_filter` and `ldap.escape_dn` are pure and need no permission.

## Consequences
Extensions can reach company databases and directories after the user approves a named host. A granted `sql` extension can do whatever its database account can; the consent text says it can change or delete data. The binary grows by the three drivers (see the size in the change that added them). Tests use a fake `database/sql` driver and an in-process LDAP server speaking real BER over loopback; nothing here was run against a live PostgreSQL, MySQL or directory server.

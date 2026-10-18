# Writing a Zagros extension (guide for coding agents)

Zagros is a keyboard launcher for Windows. An extension is a folder with a
`manifest.toml` and a Starlark script (`main.star`). Starlark is a Python
dialect, not Python.

## Find the API first, do not guess

The complete API is in `zagros-stubs/zagros/*.pyi` (next to this file). The
stubs are type descriptions, not runnable Starlark: never `import` them, every
name in them is already defined in `main.star`.

How to find a function, fastest first:

1. `API_INDEX.md` (next to this file) has one line per function:
   `module.name(args) -> ret | Needs: permission | summary | Keywords: ...`.
   Grep it with the words you would use for the task, for example
   `grep -i "unzip\|decompress" API_INDEX.md` or `grep -i "base64" API_INDEX.md`.
   The Keywords hold synonyms and task phrases, so try two or three wordings.
2. If the user connected the Zagros MCP server, call `search_api` with a task
   ("turn text into base64", "read the clipboard"), then `get_api` for the
   parameters and an example. `list_modules` shows the module table below.
3. The cheat sheet below maps common tasks to functions.
4. Open the stub (`zagros-stubs/zagros/<module>.pyi`) for the full docstring.

Every function documents the permission it needs (`Needs: network`) and has a
runnable `Example:`. Do not guess names: if grep finds nothing, try a synonym
before writing the helper yourself.

### Modules

<!-- BEGIN GENERATED: modules -->

| Stub file | Permission | What it covers | Functions |
|---|---|---|---|
| `__init__.pyi` (global names) | none | Zagros extension API: the global names that exist in every extension script (item, scope, action_*, notify, settings). | 9 |
| `hooks.pyi` (global names) | none | Hooks: top-level functions you define in main.star. Each receives ctx first. | 5 |
| `permissions.pyi` (global names) | per function | Permissions: declare in manifest.toml, the user approves them (like Android). | - |
| `ai.pyi` (module `ai`) | ai | Ask a language model, or let it call your functions. Needs: ai, scoped by provider name. | 6 |
| `binary.pyi` (module `binary`) | none | Pack and unpack binary data, hex, integers and byte-slice helpers. | 11 |
| `cipher.pyi` (module `cipher`) | none | Encryption: AES, ChaCha20, RSA, Ed25519, XOR and classic ciphers, each with a decrypt counterpart. | 76 |
| `clipboard.pyi` (module `clipboard`) | clipboard.read / clipboard.write | Read and replace the system clipboard text. | 2 |
| `compress.pyi` (module `compress`) | none | Gzip, zlib and deflate for bytes, and zip, tar and tar.gz archives of files. | 10 |
| `crypto.pyi` (module `crypto`) | none | The three hashing helpers every extension has (use the `hash` module for more algorithms). | 3 |
| `csv.pyi` (module `csv`) | none | CSV: Parse CSV and TSV into rows or dicts, auto-detect the delimiter, and write rows back. | 3 |
| `dns.pyi` (module `dns`) | network | Look up DNS records (A, AAAA, MX, TXT, CNAME, NS, PTR). | 1 |
| `dt.pyi` (module `dt`) | none | Date and time: Parse, convert and calculate dates, times, time zones, durations and cron schedules. | 66 |
| `encode.pyi` (module `encode`) | none | Encoding and escaping: Base16/32/58/62/64/85 and more, charsets, URL, HTML, punycode, escape and unescape, JWT decode. | 117 |
| `env.pyi` (module `env`) | env.read | Read environment variables. | 1 |
| `extract.pyi` (module `extract`) | none | Extractors: Pull emails, URLs, domains, IPs, numbers, phone numbers, dates and more out of any text. | 62 |
| `fs.pyi` (module `fs`) | fs.read / fs.write | Read, write, copy, move, delete, search and inspect files and folders. | 14 |
| `hash.pyi` (module `hash`) | none | Hashes: Digests (MD5, SHA-1/2/3, BLAKE, CRC, Adler, FNV, xxHash), HMAC, PBKDF2, bcrypt, scrypt, Argon2, TOTP. | 112 |
| `host.pyi` (module `host`) | none | Facts about the running Zagros, to adapt to its version. | 2 |
| `html.pyi` (module `html`) | none | HTML parsing: Parse HTML, select elements with CSS selectors, extract text and attributes. | 5 |
| `http.pyi` (module `http`) | network | Make HTTP requests, upload and download files, read JSON from web APIs. | 12 |
| `json.pyi` (module `json`) | none | The standard Starlark JSON module. | 4 |
| `jwt.pyi` (module `jwt`) | none | JSON Web Tokens: Decode JSON Web Tokens, and sign or verify them (HS, RS, ES and EdDSA). | 4 |
| `ldap.pyi` (module `ldap`) | ldap | Search and edit an LDAP directory such as Active Directory or OpenLDAP. | 3 |
| `lines.pyi` (module `lines`) | none | Lines: Work on multi-line text: sort, unique, filter, join, columns, CSV and list conversions. | 60 |
| `log.pyi` (module `log`) | none | Write to <DataDir>/zagros.log; print() logs at info. | 3 |
| `math.pyi` (module `math`) | none | Math and statistics: An expression evaluator, percentages, statistics, algebra, geometry and finance helpers. | 116 |
| `net.pyi` (module `net`) | none | Networking: CIDR and subnet calculators, IP and MAC tools, ports, HTTP status, MIME types, URLs and domains. | 92 |
| `num.pyi` (module `num`) | none | Numbers and bases: Base conversion, the hex and binary calculator, bit operations, IEEE 754, roman numerals, number words, primes and checksums. | 111 |
| `path.pyi` (module `path`) | none | File paths: Join, split, clean and relate file paths in Windows or POSIX flavour, without touching the disk. | 13 |
| `proc.pyi` (module `proc`) | process.run | Run a program without a shell and read its output. | 1 |
| `rand.pyi` (module `rand`) | none | Random generation: Random numbers, dice, strings, passwords, passphrases, colors and picks, with optional seeds. | 55 |
| `regex.pyi` (module `regex`) | none | Regular expressions: Test, match, replace and explain regular expressions (Go RE2 syntax). | 25 |
| `secrets.pyi` (module `secrets`) | secrets | Per-extension secrets in Windows Credential Manager. Needs: secrets. | 3 |
| `sql.pyi` (module `sql`) | sql | Talk to a PostgreSQL or MySQL/MariaDB server. Same shape as sqlite. | 1 |
| `sqlite.pyi` (module `sqlite`) | storage | A private SQL database per extension (pure Go SQLite, no setup). | 1 |
| `ssh.pyi` (module `ssh`) | ssh | Run commands and transfer files on a server over SSH (with SFTP). | 1 |
| `storage.pyi` (module `storage`) | none | A per-extension key/value store that survives restarts and updates. | 4 |
| `tcp.pyi` (module `tcp`) | network | Open TCP connections, send and read raw bytes, or do one request and reply. | 2 |
| `template.pyi` (module `template`) | none | Text templates: Render text from a safe subset of Go text/template and a dict. | 2 |
| `text.pyi` (module `text`) | none | Text: Change, measure and analyse strings: case styles, palindromes, lengths, normalisation, similarity, Unicode. | 107 |
| `time.pyi` (module `time`) | none | The standard Starlark time module (durations and times as time.Duration and time.Time values). | 6 |
| `tls.pyi` (module `tls`) | network | Open verified TLS connections and read the certificate of a server. | 2 |
| `toml.pyi` (module `toml`) | none | TOML: Parse TOML into dicts and lists, and write values back as TOML. | 3 |
| `udp.pyi` (module `udp`) | network | Send and receive UDP datagrams. | 3 |
| `url.pyi` (module `url`) | none | URLs: Parse, build, join and normalize URLs and query strings as structured values. | 6 |
| `uuid.pyi` (module `uuid`) | none | UUIDs and IDs: UUID v1-v7, ULID, KSUID, Snowflake, ObjectID, nanoid: generate, parse, convert. | 37 |
| `windows.pyi` (module `windows`) | registry.read for the registry functions | Read-only Windows system information. Windows only. | 8 |
| `xml.pyi` (module `xml`) | none | XML: Parse XML into a tree of dicts, query it with simple paths, and write it back. | 5 |
| `yaml.pyi` (module `yaml`) | none | YAML: Parse YAML into dicts and lists, and write values back as YAML. | 4 |

<!-- END GENERATED: modules -->

### Task cheat sheet

<!-- BEGIN GENERATED: cheatsheet -->

- add time to a date: `dt.add`
- add working days to a date: `dt.add_business_days`
- append to a file: `fs.append_text` (needs fs.write)
- apply ROT13: `cipher.rot13`
- ask a language model a question: `ai.complete` (needs ai)
- average of numbers: `math.mean`
- build a Basic auth header: `net.basic_auth_header`
- build a binary record: `binary.pack`
- build a query string: `encode.query_string_build`
- calculate a tip: `math.tip`
- check a password against its hash: `hash.bcrypt_verify`
- check a regex compiles: `regex.is_valid`
- check a URL or read its headers: `http.head` (needs network)
- check for a palindrome: `text.is_palindrome`
- check for a prime number: `num.is_prime`
- check for an anagram: `text.is_anagram`
- check if an IP is inside a subnet: `net.cidr_contains`
- check if text is a UUID: `uuid.uuid_validate`
- check if text is an IP address: `net.is_valid_ip`
- check when a certificate expires: `tls.cert` (needs network)
- checksum text with CRC-32: `hash.crc32`
- classify an IP address: `net.ip_type`
- compare two checksums: `hash.checksum_compare`
- compare two texts line by line: `lines.lines_diff`
- compare two version numbers: `host.compare_versions`
- connect to a PostgreSQL or MySQL database: `sql.open` (needs sql scoped to the host)
- connect to a server over SSH: `ssh.connect` (needs ssh)
- connect to an LDAP or Active Directory server: `ldap.connect` (needs ldap scoped to the host)
- convert a date between formats: `dt.convert`
- convert a date to a Unix timestamp: `dt.to_unix`
- convert a number between bases: `num.to_base`
- convert a number to binary: `num.to_bin`
- convert a number to hex: `num.to_hex`
- convert a time between time zones: `dt.convert_tz`
- convert a Unix timestamp to a date: `dt.from_unix`
- convert an integer to bytes: `binary.from_int`
- convert an IP address to an integer: `net.ip_to_int`
- convert bytes to an integer: `binary.to_int`
- convert hex to a number: `num.from_hex`
- convert temperature: `math.temperature`
- convert to camelCase: `text.camel`
- convert to kebab-case: `text.kebab`
- convert to roman numerals: `num.roman_encode`
- convert to snake_case: `text.snake`
- copy a file or folder: `fs.copy` (needs fs.read for src, fs.write for dst)
- copy a result to the clipboard on Enter: `action_copy` (needs clipboard.write)
- count characters: `text.length`
- count days between two dates: `dt.days_between`
- count days until a date: `dt.days_until`
- count lines: `lines.count_lines`
- count words: `text.word_count`
- create a folder: `fs.mkdir` (needs fs.write)
- decode a percent-encoded string: `encode.url_decode`
- decode base64: `encode.base64_decode`
- decode hex to text: `encode.hex_decode`
- decompress bzip2 data: `compress.bunzip2`
- decompress gzip data: `compress.gunzip`
- decompress zlib data: `compress.unzlib`
- decrypt AES-GCM text: `cipher.aes_gcm_decrypt`
- decrypt text with a named cipher: `cipher.decrypt`
- deflate data: `compress.deflate`
- delete a file or folder: `fs.remove` (needs fs.write)
- delete a resource on a web API: `http.delete` (needs network)
- describe a CIDR block: `net.cidr_info`
- describe a function as a model tool: `ai.tool`
- describe a time relative to now: `dt.relative`
- distance between coordinates: `math.haversine`
- download a file from the internet: `http.download` (needs network and fs.write for the path)
- edit distance between strings: `text.levenshtein`
- encrypt text with a named cipher: `cipher.encrypt`
- encrypt text with AES: `cipher.aes_gcm_encrypt`
- escape text for a JSON string: `encode.json_escape`
- escape text for a regex: `regex.escape`
- escape text for HTML: `encode.html_escape`
- estimate a download time: `net.transfer_time`
- evaluate a math expression: `math.eval`
- explain a cron schedule: `dt.cron_explain`
- explain a regex in words: `regex.explain`
- extract by kind name: `extract.extract`
- factor a number: `num.prime_factors`
- fetch a web page: `http.get` (needs network)
- fetch several URLs at the same time: `http.fetch_all` (needs network)
- fill a Tab sub-list: `on_scope`
- find a meeting time across time zones: `dt.meeting_planner`
- find all regex matches: `regex.find_all`
- find files by pattern: `fs.glob` (needs fs.read)
- find the MIME type of a file name: `net.mime_type`
- find the service for a port: `net.port_service`
- find the weekday of a date: `dt.day_of_week`
- fix garbled text: `encode.mojibake_fix`
- flip a coin: `rand.coin_flip`
- format a byte count: `num.humanize_bytes`
- format a date: `dt.format`
- generate a passphrase: `rand.passphrase`
- generate a password: `rand.password`
- generate a random string: `rand.string`
- generate a short random id: `uuid.nanoid`
- generate a sortable UUID: `uuid.uuid_v7`
- generate a ULID: `uuid.ulid_new`
- generate a UUID: `uuid.uuid_v4`
- generate an Ed25519 key pair: `cipher.ed25519_generate`
- generate an RSA key pair: `cipher.rsa_generate`
- generate many UUIDs: `uuid.uuid_bulk`
- generate placeholder text: `text.lorem`
- get file size or modified time: `fs.stat` (needs fs.read)
- get the current time: `dt.now`
- get the current time as a value: `time.now`
- get the first regex match: `regex.match`
- give a language model the built-in tools: `ai.registry_tools`
- greatest common divisor: `num.gcd`
- guess how text is encoded: `encode.detect`
- gzip some data: `compress.gzip`
- handle an error from a host function: `attempt`
- handle Enter on a row with a custom action: `on_execute`
- hash a password: `hash.bcrypt_hash`
- hash text with any algorithm: `hash.hash`
- hash text with MD5: `hash.md5`
- hash text with SHA-256: `hash.sha256`
- hide sensitive data in text: `extract.redact`
- identify what kind of ID this is: `uuid.id_detect`
- inflate raw deflate data: `compress.inflate`
- inspect a UUID: `uuid.uuid_parse`
- join a base URL and a path: `net.url_join`
- join lines into one: `lines.join_lines`
- keep lines containing text: `lines.filter_contains`
- least common multiple: `num.lcm`
- let a language model call my functions: `ai.run` (needs ai)
- list a folder: `fs.list` (needs fs.read)
- list a folder recursively: `fs.walk` (needs fs.read)
- list the next cron run times: `dt.cron_next`
- look up an HTTP status code: `net.http_status`
- look up DNS records of a domain: `dns.lookup` (needs network)
- make a custom HTTP request: `http.request` (needs network)
- make a one-time password code: `hash.totp`
- make a random hex token: `rand.token_hex`
- make a URL slug: `text.slug`
- make text lowercase: `text.lower`
- make text uppercase: `text.upper`
- median of numbers: `math.median`
- monthly loan payment: `math.loan_payment`
- normalize line endings: `lines.to_lf`
- open a sub-list when the user presses Tab: `scope`
- open a TCP connection to a host and port: `tcp.connect` (needs network)
- open a UDP socket: `udp.open` (needs network)
- open a URL or file on Enter: `action_open` (needs open)
- open a verified TLS connection: `tls.connect` (needs network)
- open the extension's database: `sqlite.open`
- parse a binary record: `binary.unpack`
- parse a date string: `dt.parse`
- parse a duration: `dt.duration_parse`
- parse a field inside a larger buffer: `binary.unpack_from`
- parse a hex string: `binary.unhex`
- parse a query string: `encode.query_string_parse`
- parse JSON text: `json.decode`
- partially update a resource on a web API: `http.patch` (needs network)
- percent change between two values: `math.percent_change`
- percent-encode text for a URL: `encode.url_encode`
- pick a random integer: `rand.int`
- pick one at random: `rand.choice`
- price after discount: `math.discount`
- produce launcher results for what the user typed: `on_query`
- pull email addresses out of text: `extract.emails`
- pull IP addresses out of text: `extract.ips`
- pull links out of text: `extract.urls`
- pull numbers out of text: `extract.numbers`
- put text on the clipboard: `clipboard.set` (needs clipboard.write)
- put user input safely into an LDAP DN: `ldap.escape_dn`
- put user input safely into an LDAP filter: `ldap.escape_filter`
- quote an argument for the shell: `encode.sh_quote`
- random color: `rand.color_hex`
- read a binary file: `fs.read_bytes` (needs fs.read)
- read a Snowflake ID: `uuid.snowflake_decode`
- read a text file: `fs.read_text` (needs fs.read)
- read an environment variable: `env.get` (needs env.read)
- read JSON from a web API: `http.json` (needs network)
- read the claims of a JWT: `encode.jwt_decode`
- read the clipboard: `clipboard.get` (needs clipboard.read)
- read the Zagros version: `host.version`
- remember a value between runs: `storage.get`
- remove accents: `text.strip_accents`
- remove blank lines: `lines.remove_empty_lines`
- remove duplicate lines: `lines.unique_lines`
- remove tracking parameters from a URL: `net.url_strip_tracking`
- rename or move a file: `fs.rename` (needs fs.write)
- replace a resource on a web API: `http.put` (needs network)
- replace text: `text.replace`
- replace with a regex: `regex.replace`
- return a result row: `item`
- reverse a string: `text.reverse`
- reverse bytes: `binary.reverse`
- roll dice: `rand.dice`
- round a number: `math.round_to`
- run a command and capture its output: `proc.run` (needs process.run)
- run slow work without blocking the launcher: `task`
- run your own code when the user confirms a row: `action_call`
- see every common hash of a text: `hash.hash_all`
- see everything extractable from a text: `extract.all`
- see what is inside an archive: `compress.list` (needs fs.read)
- see which AI providers are configured: `ai.providers` (needs ai)
- send a UDP packet and read the reply: `udp.request` (needs network)
- send a UDP packet without waiting for a reply: `udp.send` (needs network)
- send data to a web API: `http.post` (needs network)
- send one request over TCP and get the reply: `tcp.request` (needs network)
- shift letters with a Caesar cipher: `cipher.caesar`
- shorten text with an ellipsis: `text.truncate`
- show a desktop notification: `notify` (needs notifications)
- show bytes as hex: `binary.hex`
- shuffle a list: `rand.shuffle`
- sign data with RSA: `cipher.rsa_sign`
- sign text with a secret key: `hash.hmac`
- simplify a fraction: `math.simplify_fraction`
- solve a quadratic equation: `math.solve_quadratic`
- sort IP addresses: `net.sort_ips`
- sort lines: `lines.sort_lines`
- split a domain name: `net.domain_parts`
- split a network into subnets: `net.cidr_subnets`
- split a URL into parts: `net.url_parse`
- split bytes into fixed-size blocks: `binary.chunks`
- split text into lines: `lines.split_to_lines`
- split text with a regex: `regex.split`
- standard deviation: `math.stdev`
- start a program on Enter: `action_run` (needs process.run)
- strip HTML tags from text: `encode.strip_html_tags`
- sum numbers: `math.sum`
- take a percentage of a number: `math.percent_of`
- take the first lines: `lines.head`
- take the last lines: `lines.tail`
- test a regex against text: `regex.test`
- time between two dates: `dt.diff`
- title case text: `text.title`
- trim whitespace: `text.trim`
- turn a glob into a regex: `regex.from_glob`
- turn a value into JSON text: `json.encode`
- turn text into base64: `encode.base64_encode`
- turn text into embedding vectors: `ai.embed` (needs ai)
- turn text into hex: `encode.hex_encode`
- unzip an archive: `compress.extract` (needs fs.read for archive, fs.write for dest)
- validate a card number checksum: `num.luhn_check`
- validate an email address: `net.is_valid_email`
- validate an IBAN: `num.iban_check`
- validate an ISBN: `num.isbn13_check`
- work out an age from a birth date: `dt.age`
- work out which hash this is: `hash.hash_identify`
- write a binary file: `fs.write_bytes` (needs fs.write)
- write a debug message to the log: `log.info`
- write a number in words: `num.number_to_words`
- write a text file: `fs.write_text` (needs fs.write)
- xor bytes: `binary.xor`
- zip files or folders: `compress.create` (needs fs.read for sources, fs.write for dest)
- zlib-compress data: `compress.zlib`

<!-- END GENERATED: cheatsheet -->

Both lists above are generated from the stubs (`go run ./cmd/toolsgen` in the
Zagros repository); a function missing here has no `Task:` line, but it is
still in `API_INDEX.md`.

## Workflow

1. Edit `manifest.toml` (id, name, version, `keywords`, `permissions`).
2. Edit `main.star`. Define `on_query(ctx, query)` returning `item(...)` rows.
3. `zagros ext validate .` (manifest, syntax, undefined names, undeclared permission use).
4. `zagros --dev` hot-reloads this folder while you edit; install with `zagros ext install .` (or the `install_extension` MCP tool when the user connected the Zagros MCP server). Never try to grant permissions: Zagros asks the user.
5. `zagros ext pack .` writes `<id>-<version>.zagros`. Publishing is in the marketplace docs.

## Starlark rules

- No `import`, classes, `try/except`, `with`, f-strings or `yield`. Split code across files with `load("other.star", "name")`.
- Host functions (`fs.*`, `http.*`, `sqlite.*`, `sql.*`, `ldap.*`, permission errors) raise on failure. There is no `try`; catch with the global `value, err = attempt(fn, arg1, arg2)` (pass the function, not its result), or test first with `fs.exists`. `r.ok` is false for a non-2xx HTTP status, which is not an error.
- Available: `for`, `if`, `while`, recursion, functions, lambdas, list/dict/set comprehensions, `"%s" % x` and `.format()` formatting, top-level statements.
- Everything in the stubs is already defined. Do not import it.
- Globals are frozen after load; keep state in `storage` or return it.
- Hooks run with a time and step limit (5 seconds). For several requests use `http.fetch_all([...])` (parallel, results in order, failed entries have `ok = False` and `error`). For anything slow (large downloads, big archives, slow APIs) put the work in a top-level function and start it with `t = task("fn_name", args)`: it returns a handle at once, runs with a 5 minute limit, and when it finishes Zagros runs the query again so `task("fn_name", args)` returns the same handle with `done`, `ok`, `value` and `error`. Return `item(id, title, pending = True)` while `not t.done`. A task is cancelled when the next query stops asking for it, and it cannot start another task.

## Permissions

Declare only what you use, as narrowly as possible, in `manifest.toml`:

```toml
permissions = ["network:api.example.com", "clipboard.write"]

[[permission_detail]]
permission = "network:api.example.com"
reason = "Looks up results."
```

The user approves them. Calling a function without its permission fails.
UI features (scopes, Tab completion, alternate actions) need no permission.
Files are scoped by folder (`fs.read:~/Documents`, `fs.write:~/Documents/out`); `fs.copy`, `fs.rename`, `fs.glob`, `fs.walk`, `http.download` and `compress.create/extract` all check the same scopes, so declare the narrowest folder. `sqlite.open(name)` gives a private database in the extension's data folder and needs no declaration (only `storage`, always granted); use `?` placeholders, never string formatting. `binary` and the bytes half of `compress` are pure.

Raw sockets and DNS (`tcp`, `udp`, `tls`, `dns`) use `network`, scoped `network:host` or `network:host:port` (a port in the scope does not cover `http` or `dns`). A scoped name that resolves to a private or loopback address is refused; grant the IP address to reach a local machine. Connections close when the call ends, so open, use and close inside one hook. `tls.cert(host)` returns `days_left`, `sans`, `issuer`, `verified`; `tls.connect` always verifies. `ssh.connect(host, user, auth={"secret": "name"})` needs `ssh:user@host[:port]` (and `secrets` for a named secret or `key_secret`); host keys are remembered on first use, a changed key raises, and `host_key="SHA256:..."` pins one. `exec` returns `{stdout, stderr, code, ok}` and a non-zero exit is not an error. Wrap all of these in `attempt()`.

Database and directory servers have their own permissions, scoped by host and optional port: `sql:db.example.com:5432` for `sql.open("postgres" | "mysql", host, ...)`, `ldap:ldap.corp.example.com` for `ldap.connect` (search, check_bind) and `ldap.write:ldap.corp.example.com` for add, modify and delete. The `sql` handle has the same `query`, `exec`, `transaction` and `close` as `sqlite`, with `?` or `:name` placeholders. Keep passwords out of the source: `password={"secret": "name"}` reads from `secrets` (declare `secrets` too). TLS is on and verified, and there is no skip option; a password is refused over plain `ldap://` unless `start_tls=True`. Put user input into LDAP filters with `ldap.escape_filter` and into DNs with `ldap.escape_dn`. Wrap network calls in `attempt()`, since servers are down more often than files are missing.
Language models: `ai` (permission `ai`, scoped by provider name such as `ai:claude`; the user configures providers in zagros.toml and stores keys with `zagros ai key set`, so you never handle a key). Declare `ai_budget_tokens` in manifest.toml (default 50000 a day); calls raise a clear error when it is used up. Use `ai.complete` for one answer and `ai.run` with `ai.tool(fn, description=..., params=...)` or `ai.registry_tools([...])` for tool use. Tools run with your extension's own permissions only, so give them the smallest job and treat text from the web or files as untrusted. Call `ai` from `on_execute` or an alternate action, not on every `on_query` keystroke. The 5 second hook limit pauses while a provider request is in flight (60 s per request, 180 s per hook call).
Keep API tokens and passwords in `secrets` (permission `secrets`, Windows Credential Manager, private to the extension), not in `storage` or settings. The `windows` module reads system state on Windows only: `registry.read:HKCU\Software\X` (scoped by key prefix) for registry reads, `system.read` for services, windows, processes and known folders. It cannot change anything.

## Ready-made function libraries

Before writing helper code, search `API_INDEX.md` or `search_api`. The modules whose permission column says `none` in the table above are pure libraries (case styles, base conversion, base64/32/58/85, URL and HTML escaping, hashes, AES, random values, UUIDs, regex, CIDR math, dates and time zones, extracting emails and IPs, and more as they are added). They need no permission. `host.version()` and `host.compare_versions(a, b)` report the Zagros version.

## Examples in the stubs

Each `Example:` in a stub is valid Starlark that parses with the real interpreter (a test enforces it). Names such as `query`, `ctx`, `data` and `scope_id` stand for the hook arguments. Copy the shape, not the placeholder values.

## UI mechanisms you can use

- **Results:** return `item(id, title, subtitle, icon, action, alt, score, complete, scope, pending)`. Keep `id` stable for the same result (never a counter or the typed text): Zagros remembers which row the user ran for a query, or inside a scope, and puts it first next time.
- **Slow results:** `task(fn, args)` plus `item(..., pending = True)` (see Starlark rules). No permission of its own; the function it runs needs its own.
- **Scopes (Tab sub-context):** give an item `scope = scope(id, label, placeholder, query, data)`. Tab enters it: the input shows a chip with `label` and your `on_scope(ctx, scope_id, query, data)` supplies the list. Esc or Backspace on an empty input leaves it.
- **Tab completion:** `complete = "kw "` puts text in the input on Tab.
- **Alternate actions:** `alt = [("Copy", action_copy(x))]` shows in the built-in Actions scope on Tab.
- **Actions:** `action_open`, `action_copy`, `action_run`, `action_call` (calls `on_execute`).

## Minimal example

```python
def on_query(ctx, query):
    return [item(
        id = "hello",
        title = "Hello, " + (query or "world"),
        action = action_copy("Hello, " + (query or "world")),
    )]
```

(Declare `permissions = ["clipboard.write"]` for `action_copy`.)

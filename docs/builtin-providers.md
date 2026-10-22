# Built-in providers

Constructor: `builtin.All(cfg *core.Config) []core.Provider` (package `internal/builtin`).
A provider is on unless `cfg.Providers[name] = false`. `clipboard` is opt-in (needs `true`),
`everything` also needs `cfg.Everything.Enabled`. Providers with a trigger implement
`Keywords() []string`. Nothing is indexed until the first query.

Items that must never enter history (passwords, clipboard entries) carry
`Action.Extra["nohistory"] = "1"`; the app should skip frecency recording for them.
Destructive actions carry `Action.Extra["confirm"] = "1"`.

| Provider name | Trigger | What it does |
|---|---|---|
| `apps` | any text | Start Menu (user + common) `.lnk`/`.url`, `scan_dirs` (`.exe` too, depth 3), Store apps. |
| `everything` | `f <query>`, or plain text of 3+ chars (up to 5 extra rows) | File search through Everything. Enter opens, Alt actions: reveal in Explorer, copy path. |
| `calc` | `=` prefix, or expression-looking input | Calculator. |
| `password` | `pw`, `pass`, `pp`, `pin` | Password, passphrase and PIN generators. |
| `text` | see below | Text, time, color and unit tools. |
| `system` | fuzzy on action names, `ip` | Lock, power, shell shortcuts, dark mode, DNS, local IPs. |
| `websearch` | keywords from `[[web_search]]`, aliases, bare URLs | Web search, aliases, open URLs. |
| `clipboard` | `cb [filter]` | In-memory clipboard history (opt-in). |
| `windows` | `w [title]` | Switch to an open window. |
| `processes` | `kill <name>` | Kill a process (confirm). |
| `emoji` | `:name` or `emoji name` | Copy an emoji (about 530 curated entries). |
| `runner` | `> cmd`, `ps <cmd>` | Run via cmd.exe or PowerShell. |
| `recent` | `recent [filter]` | Windows Recent files. |
| `tools` | family keywords such as `text`, `enc`, `hash`, `net`, `dt`, `clip`, plus function names | The tool families (strings, numbers, encoding, hashes, ciphers, random, regex, network, date/time, extractors, clipboard). See [Built-in tools](tools.md). |
| `reindex` | `reindex`, `rescan` | Rebuilds the app index now; Enter starts the background scan. |
| `update` | `update`, `upgrade` | Checks GitHub releases for a newer Zagros; Enter opens the release page, Tab offers the download and a re-check. |
| `mcp` | `mcp`, `agent` | Copies the prompt that teaches a coding agent to search the extension API; Tab: other setup choices. |
| `help` | `?` or `help [filter]` | Lists every trigger; Enter fills the keyword into the query. |

## apps

Index sources: `%ProgramData%` and `%APPDATA%` Start Menu `Programs`, every directory in
`scan_dirs`, and Store apps from `Get-StartApps` (entries with a `Package!App` id, opened as
`shell:AppsFolder\<AUMID>`). Shortcut targets are resolved with `IShellLinkW` through raw
COM calls (no CGO) and used for icons. Entries named "uninstall" and shortcuts to documents
are skipped. The index is cached in `<DataDir>\apps-index.json`, loaded on the first query,
and rescanned in the background when older than 12 hours. Type `reindex` (or `rescan`) and press Enter to rescan immediately; extensions can call `host.reindex_apps()`. Matching is coarse (in-order
letters); the catalog does fuzzy ranking.

## everything

Loads `Everything64.dll` (`Everything32.dll` / `EverythingARM64.dll` per architecture) from, in
order: `everything.dll_path`, next to the exe, `Program Files\Everything` (and `Everything 1.5a`),
then `PATH`. Only absolute paths are loaded. Falls back to `es.exe` when found. If neither
exists, or the Everything service is not running, the provider returns nothing; `Status()`
(on the concrete type, via interface `interface{ Status() string }`) reports why. Result
count is `everything.limit` (default 20).

## calc

`+ - * / % ^ **`, parentheses, unary minus (`-2^2 = -4`), `pi e tau`, `0xff`, `0b101`,
`1e3`, thousands separators (`1,000,000`), `×` and `÷`.
Percent: `50%` is 0.5, `200 + 10%` is 220, `200 * 10%` is 20. `%` followed by an operand is modulo.
Functions: `sqrt cbrt sin cos tan asin acos atan log log10 log2 ln exp abs floor ceil trunc round(x[,n]) pow min max deg rad`.
Enter copies the result; Alt copies with separators, hex, or binary. Bare numbers, words and
dates never produce a row; `= ...` always evaluates and shows errors.

## password

| Query | Result |
|---|---|
| `pw`, `pass` | 5 passwords, default 20 chars. Options: a number (4-128), `nosym`, `nonum`, `noupper`, `nolower`, `noamb`. Example: `pw 24 nosym`. Each enabled class appears at least once. |
| `pp [n] [cap] [num]` | Passphrases of n words (default 5, 2-20) from a 540-word list, joined by `-`. |
| `pin [n]` | Numeric PINs (default 6). |

All randomness is `crypto/rand`. Subtitle shows the entropy estimate.

## text

| Trigger | Result |
|---|---|
| `uuid` | UUID v4 (lower, upper, no dashes) |
| `b64 <text>` / `b64d <b64>` | Base64 encode (std and URL-safe) / decode |
| `url <text>` / `urld <text>` | Percent-encode (query and path style) / decode |
| `hash <text>` | MD5, SHA-1, SHA-256, SHA-512 |
| `case <text>` | UPPER, lower, snake_case, camelCase, PascalCase, kebab-case, CONSTANT_CASE, Title Case |
| `lorem [n]` | n words of lorem ipsum (default 30, max 2000) |
| `ts [value]` | Unix seconds/ms/us/ns (detected by size) or a date (`2024-01-31`, `2024-01-31 15:04`, RFC 3339) to all formats; a date gives the Unix time |
| `now` | Current time in Unix, UTC, local and readable forms |
| `json [text]` | Pretty and minified JSON from the argument, or from the clipboard when empty |
| `color <value>` | `#rgb`, `#rrggbb`, `rrggbb`, `rgb(r,g,b)`, `hsl(h,s%,l%)` to hex, rgb, hsl |
| `unit <n> <from> to <to>` | Also works without the keyword: `10 km to mi` |

Units: length (mm cm m km in ft yd mi nmi), mass (mg g kg t oz lb st), data (b kb mb gb tb, kib mib gib tib, bits),
time (ms s min h d wk), volume (ml l gal qt pt cup floz), speed (m/s km/h mph kn), temperature (C F K).
`kb`/`mb` are decimal, `kib`/`mib` binary.

## system

Type part of an action name (two or more letters). Actions marked * carry `confirm`.

Lock, Sleep, Hibernate*, Shut down*, Restart*, Sign out*, Empty Recycle Bin*, Task Manager,
Settings, Control Panel, Services, Device Manager, Toggle dark mode, Flush DNS cache.
`ip` lists local non-loopback addresses; Enter copies one.

## websearch

- `<keyword> <query>` for each `[[web_search]]` entry; `{q}` in the URL is replaced by the URL-escaped query.
- `cfg.Aliases`: typing the alias (and optionally arguments, substituted for `{q}`) opens the target. URLs, `shell:` and
  existing paths are opened; anything else is run as a command line.
- Bare URLs (`https://...` or `host.tld/path` with a known TLD, `localhost`, IPv4) get an "Open" row.

## clipboard (`cb`)

Opt in with `[providers] clipboard = true`. Polls `GetClipboardSequenceNumber` every 500 ms, keeps
the last 50 text entries in memory only (max 64 KiB each, never written to disk), de-duplicates, and skips content
flagged `ExcludeClipboardContentFromMonitorProcessing` (password managers). `cb foo` filters.
The provider has a `Close()` method that stops the poller.

## windows (`w`)

Lists visible, un-cloaked top-level windows, filtered by title substring. Enter restores and focuses.

## processes (`kill`)

`kill chrome` lists matching processes (max 20, never PID 0-4 or Zagros itself). Enter terminates after confirmation.

## emoji

`:smile`, `:thumbs up` or `emoji heart`. All words must match the name or keywords. Enter copies the glyph.

## runner

`> ipconfig /all` opens cmd.exe with `/k` (Alt: `/c`, window closes). `ps Get-Process` opens PowerShell with `-NoExit`.

## recent

`recent` lists shortcuts from `%APPDATA%\Microsoft\Windows\Recent`, newest first (cached for 10 s).

## tools

The launcher face of the tools registry; see [Built-in tools](tools.md) for the families, keywords and the Tab scope. It takes over the keywords `uuid`, `b64`, `b64d`, `url`, `urld`, `hash`, `case`, `lorem`, `ts` and `now` from the `text` provider, which keeps `json`, `color` and `unit`. Without a keyword it adds a `Number 14` or `String hello` row after the other results; Tab on it opens that family.

## update

`update` shows the state of Zagros and of your extensions.

**Zagros itself.** While Zagros runs it checks `api.github.com/repos/avestura/zagros/releases/latest` in the background (45 seconds after start, then every 6 hours), downloads the Windows zip for your CPU, verifies it against the release's `checksums.txt` (a release without one is refused, and the files must come from `github.com/avestura/zagros`), and unpacks `zagros.exe` into `<data dir>/update`. It then tells you once, briefly: a tray balloon and the line `✔ Update installed · Restart to update` on the empty launcher for a few seconds. Nothing stays on screen. Typing `update` shows **Restart to update to Zagros X**; Enter swaps in the new executable (the old one is kept as `zagros.exe.old` until the next start) and starts it. If the folder is not writable the row says so and nothing changes. With the check off, or when the download has not run yet, the row reads "Zagros X is available" and Enter downloads and verifies it now. Development builds (`dev`) never update. Turn the background check off with `updates.background = false`; `update` and the tray item then check on demand.

**Extensions.** The same check compares installed extensions with the marketplace registry. `update` lists "N extension updates available" with the versions; Enter updates all, Tab updates one. The extension manager marks them with an `update` tag and an **Update to X** button. An update that asks for new permissions prompts for them like an install. Extensions are never updated without a click. Extensions installed from a folder with `zagros ext install <path>` are yours and never offered a registry version. `zagros ext outdated` lists them and `zagros ext update <id>|--all` installs from a terminal.

`zagros update` still checks from a terminal and never installs (`--json` for scripts).

## mcp

`mcp` copies the prompt for coding agents. Tab: the command that adds the stdio server to Claude Code, the HTTP endpoint and its prompt, a switch for the HTTP server, and the docs. See [MCP server for coding agents](extensions/mcp.md).

## Platform notes

Windows-only code lives in `platform_windows.go`, `lnk_windows.go` and `everything_windows.go`;
`platform_other.go` provides stubs so the pure logic (calc, units, passwords, text tools, parsing)
builds and tests on any OS. The package uses only the standard `syscall` package, so it adds no
module dependencies.

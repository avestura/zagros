# Built-in tools

Zagros ships a large set of everyday tools, all native and instant: change and measure text, convert numbers and bases, encode and decode, hash and encrypt, generate random values and IDs, test regular expressions, calculate subnets, convert dates and time zones, extract emails and IPs from text, and transform the clipboard.

They are organised in **families**. Type a family keyword and some text, and the best variants for that text appear. Press `Tab` on any row to open the whole family for the same text.

## Families and keywords

| Family | Keywords | Examples |
|---|---|---|
| Text | `text`, `t`, `str`, `case`, `slug` | `text Hello World`, `case helloWorld` |
| Lines | `lines` | `lines` with a list on the clipboard: sort, unique, join, CSV |
| Numbers | `num`, `number`, `n`, `base`, `bits` | `num 255`, `num 0xff + 0b101`, `hex ff` |
| Math | `math`, `stats`, `percent` | `math 2(3+4)^2`, `stats 3 5 8 13`, `percent 15% of 80` |
| Encode | `enc`, `encode`, `dec`, `decode`, `b64`, `url`, `html`, `hex`, `punycode`, `esc` | `b64 hello`, `url a b&c`, `punycode m端nchen.de` |
| Hash | `hash`, `md5`, `sha256`, `crc` | `hash hello`, `sha256 hello` (digests and checksums; HMAC, PBKDF2, bcrypt, scrypt and Argon2 are in the [extension API](extensions/tools-hash.md) only) |
| Cipher | `crypt`, `cipher`, `encrypt`, `decrypt`, `aes`, `xor`, `rot13`, `caesar` | `crypt hello key=secret` |
| Random | `rand`, `random`, `roll`, `flip`, `shuffle` | `roll 2d6+3`, `rand 1 100`, `flip` |
| IDs | `uuid`, `guid`, `id`, `ulid`, `nanoid` | `uuid`, or paste a UUID to parse it |
| Regex | `re`, `regex` | `re \d+ :: order 66 of 100` |
| Network | `net`, `cidr`, `ip`, `subnet`, `mac`, `port`, `status`, `mime`, `domain` | `10.0.0.5/24`, `port 443`, `status 418` |
| Date and time | `dt`, `date`, `time`, `ts`, `now`, `tz`, `epoch`, `cron`, `diff` | `ts 1700000000`, `dt tomorrow 9am` |
| Extract | `ex`, `extract`, `emails`, `domains`, `ips`, `urls`, `numbers` | `emails` with text on the clipboard |
| Clipboard | `clip` | `clip upper`, `clip sort`, `clip base64` |

Many single functions have a keyword of their own (`lower HELLO`, `roman 1994`, `palindrome racecar`, `words 42`), and technical function names such as `sha256` or `base64_encode` work as keywords. Plain words alone (`date`, `age`, `format`) never trigger a tool, so ordinary searches stay clean.

## Input rules

- **Empty input uses the clipboard.** `b64 ` (keyword, then a space, nothing else) shows the encodings of whatever you copied. Rows say "(clipboard)". The keyword alone, without the space, stays quiet.
- **`name=value` sets a parameter.** `crypt hello key=secret`. Names are the function parameters; text that merely looks like `a=b` stays text unless some variant of that family has such a parameter.
- **`a :: b` passes several texts.** `re \d+ :: abc 123`, `net 10.0.0.0/8 :: 10.1.2.3`. The parts fill the function's string parameters in order.
- **Enter copies the result.** Tool rows never enter the usage history, because the text you typed may be a secret.
- **Automatic answers.** Without a keyword, an IP or CIDR, a `0x`/`0b` expression, a UUID, a 10 or 13 digit unix timestamp, or a JWT is recognised and answered directly.

## No keyword needed

Type a number or any text and a few rows at the bottom of the list offer the tool families for it, for example `String hello`, `Encode hello` and `Hash hello` (or `Number 14`, `Encode 14`, `Hash 14`). Press `Tab` on one to open that family for the input, with every variant, so Base64 and the checksums are one keypress away without knowing `b64` or `sha256`. When the input looks like something specific the matching family comes first: `Network` for an IP, CIDR, MAC, domain or URL, `Date` for an ISO date or unix timestamp, `ID` for a UUID, `Lines` for several lines. At most four such rows share the bottom of the list with the web search row, and they never push out more than half of a short list. The keywords above are shortcuts to the same place. The rows are interface only: extensions get the same functions through the tools modules and the scope mechanism, so no new API or permission is involved.

## Tab: the family scope

`Tab` on a result enters the family scope. The input shows a chip with the family name and keeps your text. The list shows every variant for that text:

- variants of the keyword you typed come first (after `b64 hello`, the base64 ones, then the rest of Encode);
- variants that are too slow to run on every keystroke only run here;
- type `@word` before the text to narrow the list: `@sha hello`, `@decode SGVsbG8=`;
- the tool you ran last in a family leads that family the next time (see below);
- inside a generator family (Random, IDs), `Tab` again makes new values;
- `Esc` or `Backspace` on an empty input goes back.

## Remembered picks

Zagros remembers what you run. Type `fir`, run Firefox, and Firefox is the first answer to `fir` (and to `fire`, which extends it) from then on. Inside a Tab scope the pick is remembered for the scope: run `Uppercase` in the String scope and it leads the String scope next time, whatever text you type, and the same goes for the Hash, Encode and every other scope, for web search engines and for extension scopes. Tool rows are remembered by tool, never with the text you typed, and rows that are marked as never entering history (tool results at the root) are not remembered at all.

The memory is small and cleans up after itself, in `choices.json` next to `history.json`:

- at most 300 queries, each with at most 6 items; the least recently used query goes first;
- a pick loses half its weight every 30 days and is deleted when it has faded, so unused picks disappear; an empty store deletes its file;
- an item that stops showing up for the exact query it was learned for (an uninstalled app) is deleted after three misses in a row, and an item whose action fails is deleted at once. Providers that are off, still indexing or timed out never count as a miss;
- queries over 40 characters, and items whose IDs are not stable (clipboard history entries, the calculator result, windows, processes), are never stored.

Extension items get this automatically, because it works on the item `id` that every extension item has. Give items stable ids for it to work: `id="open-docs"`, not an id that includes a counter or the typed text.

## The clipboard family

`clip` applies the text, lines, encode and extract tools to the clipboard. Words after `clip` filter the transformations (`clip upper`, `clip sort`, `clip trim`). Enter puts the result on the clipboard, which modifies it in place.

## For extensions

Every function behind these families is available to extensions as a module (`text`, `lines`, `num`, `math`, `encode`, `hash`, `cipher`, `rand`, `uuid`, `regex`, `net`, `dt`, `extract`, plus `yaml`, `toml`, `csv`, `xml`, `html`, `url`, `path`, `template`, `jwt` for structured data) with no permission needed, because they are pure functions. See the [tools API](extensions/tools.md).

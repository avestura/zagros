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

Type a number or any text and a row at the bottom of the list reads `Number 14` or `String hello`. Press `Tab` on it to open the Numbers or Text family for that input, with every variant. The keywords above are shortcuts to the same place; you never have to know them. The row is interface only: extensions get the same functions through the tools modules and the scope mechanism, so no new API or permission is involved.

## Tab: the family scope

`Tab` on a result enters the family scope. The input shows a chip with the family name and keeps your text. The list shows every variant for that text:

- variants of the keyword you typed come first (after `b64 hello`, the base64 ones, then the rest of Encode);
- variants that are too slow to run on every keystroke only run here;
- type `@word` before the text to narrow the list: `@sha hello`, `@decode SGVsbG8=`;
- inside a generator family (Random, IDs), `Tab` again makes new values;
- `Esc` or `Backspace` on an empty input goes back.

## The clipboard family

`clip` applies the text, lines, encode and extract tools to the clipboard. Words after `clip` filter the transformations (`clip upper`, `clip sort`, `clip trim`). Enter puts the result on the clipboard, which modifies it in place.

## For extensions

Every function behind these families is available to extensions as a module (`text`, `lines`, `num`, `math`, `encode`, `hash`, `cipher`, `rand`, `uuid`, `regex`, `net`, `dt`, `extract`) with no permission needed, because they are pure functions. See the [tools API](extensions/tools.md).

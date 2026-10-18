# Random generation (`rand`)

Random numbers, dice, strings, passwords, passphrases, colors and picks, with optional seeds.

No permission is needed: these are pure functions. Generated from the tools registry.

55 functions. Use them as `rand.<name>(...)` in any extension.

## rand.bool

`rand.bool(probability: float = 0.5, seed: int = 0) -> bool`

Returns true with the given probability (default 0.5) and false otherwise.

| Parameter | Type | Description |
|---|---|---|
| `probability` | `float` (default `0.5`) | Chance of true between 0 and 1. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `bool`. Example: `rand.bool(0.25)`

## rand.bytes_hex

`rand.bytes_hex(count: int = 16, seed: int = 0) -> str`

Returns count random bytes as hex. Uses crypto/rand when seed is 0, otherwise a reproducible pseudo-random stream (not for secrets).

| Parameter | Type | Description |
|---|---|---|
| `count` | `int` (default `16`) | Number of bytes (1..50000). |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.bytes_hex(8)`

## rand.coin_flips

`rand.coin_flips(n: int = 10, seed: int = 0) -> Dict[str, str]`

Flips n fair coins (0..10000) and returns the counts of heads and tails plus the H/T sequence (sequence only when n &lt;= 200).

| Parameter | Type | Description |
|---|---|---|
| `n` | `int` (default `10`) | Number of flips. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `Dict[str, str]`. Example: `rand.coin_flips(10)`

## rand.color_hsl

`rand.color_hsl(pure: bool = False, seed: int = 0) -> str`

Returns a random colour as a CSS hsl(h, s%, l%) string; pleasant vivid by default, uniform with pure=true.

| Parameter | Type | Description |
|---|---|---|
| `pure` | `bool` (default `False`) | Pure random RGB instead of a pleasant vivid colour. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.color_hsl()`

## rand.color_name

`rand.color_name(seed: int = 0) -> str`

Returns a random CSS colour name with its hex value, e.g. "coral #ff7f50".

| Parameter | Type | Description |
|---|---|---|
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.color_name()`

## rand.color_palette

`rand.color_palette(n: int = 5, seed: int = 0) -> List[str]`

Returns n (1..100) harmonious hex colours: hues are spaced by the golden angle from a random start so neighbours stay distinct.

| Parameter | Type | Description |
|---|---|---|
| `n` | `int` (default `5`) | Number of colours. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `List[str]`. Example: `rand.color_palette(5)`

## rand.color_rgb

`rand.color_rgb(pure: bool = False, seed: int = 0) -> str`

Returns a random colour as a CSS rgb(r, g, b) string; pleasant vivid by default, uniform with pure=true.

| Parameter | Type | Description |
|---|---|---|
| `pure` | `bool` (default `False`) | Pure random RGB instead of a pleasant vivid colour. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.color_rgb()`

## rand.email_fake

`rand.email_fake(domain: str = "example.com", seed: int = 0) -> str`

Returns a plausible fake email such as "jane.smith42@example.com". The default domain example.com is reserved for documentation so no real mailbox is hit.

| Parameter | Type | Description |
|---|---|---|
| `domain` | `str` (default `"example.com"`) | Domain part. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.email_fake("example.org")`

## rand.ints

`rand.ints(count: int = 10, min: int = 0, max: int = 100, unique: bool = False, seed: int = 0) -> List[float]`

Returns count random integers in \[min, max\]; with unique=true no value repeats (errors when the range is too small). Values above 2^53 lose precision in the float list.

| Parameter | Type | Description |
|---|---|---|
| `count` | `int` (default `10`) | How many numbers (0..10000). |
| `min` | `int` (default `0`) | Lowest value. |
| `max` | `int` (default `100`) | Highest value. |
| `unique` | `bool` (default `False`) | Forbid repeated values. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `List[float]`. Example: `rand.ints(5, 1, 49, unique=True)`

## rand.ip_v4

`rand.ip_v4(private: bool = False, seed: int = 0) -> str`

Returns a random IPv4 address. private=true picks from 10/8, 172.16/12 or 192.168/16; otherwise a routable public-looking address that avoids loopback, multicast, link-local, CGNAT, documentation and reserved ranges.

| Parameter | Type | Description |
|---|---|---|
| `private` | `bool` (default `False`) | Use an RFC 1918 private address. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.ip_v4(private=True)`

## rand.ip_v6

`rand.ip_v6(private: bool = False, seed: int = 0) -> str`

Returns a random IPv6 address in compressed form. private=true gives a unique-local fd00::/8 address; otherwise a global unicast address in 2000::/3 outside the documentation prefix.

| Parameter | Type | Description |
|---|---|---|
| `private` | `bool` (default `False`) | Use a unique local (fd00::/8) address. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.ip_v6()`

## rand.lorem_word

`rand.lorem_word(seed: int = 0) -> str`

Returns one random lorem ipsum word. For paragraphs use the text group.

| Parameter | Type | Description |
|---|---|---|
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.lorem_word()`

## rand.lorem_words

`rand.lorem_words(count: int = 5, capitalize: bool = False, seed: int = 0) -> str`

Returns count random lorem ipsum words (0..10000) joined by single spaces; capitalize upper-cases the first letter.

| Parameter | Type | Description |
|---|---|---|
| `count` | `int` (default `5`) | Number of words. |
| `capitalize` | `bool` (default `False`) | Capitalize the first word. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.lorem_words(6, True)`

## rand.mac_address

`rand.mac_address(format: str = "colon", upper: bool = False, seed: int = 0) -> str`

Returns a random unicast, locally administered MAC address. Format is colon (aa:bb:..), dash (AA-BB-..), dot (aabb.ccdd.eeff) or plain (aabbccddeeff); upper switches to uppercase.

| Parameter | Type | Description |
|---|---|---|
| `format` | `str` (default `"colon"`) | colon, dash, dot or plain. |
| `upper` | `bool` (default `False`) | Uppercase hex digits. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.mac_address("dash", upper=True)`

## rand.name_first

`rand.name_first(gender: str = "any", seed: int = 0) -> str`

Returns a random given name from a small built-in list; gender is any, male or female.

| Parameter | Type | Description |
|---|---|---|
| `gender` | `str` (default `"any"`) | any, male or female. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.name_first("female")`

## rand.name_last

`rand.name_last(seed: int = 0) -> str`

Returns a random family name from a small built-in list.

| Parameter | Type | Description |
|---|---|---|
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.name_last()`

## rand.number_sequence

`rand.number_sequence(count: int = 10, start: int = 0, max_step: int = 10, kind: str = "ascending", seed: int = 0) -> List[float]`

Generates count integers as an ascending list, a descending list (kind) or a random walk starting at start with steps of at most max_step. Ascending/descending draw from \[start, start + count*max_step\].

| Parameter | Type | Description |
|---|---|---|
| `count` | `int` (default `10`) | Length of the sequence (0..10000). |
| `start` | `int` (default `0`) | First value / lower bound. |
| `max_step` | `int` (default `10`) | Largest step between neighbours. |
| `kind` | `str` (default `"ascending"`) | ascending, descending or walk. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `List[float]`. Example: `rand.number_sequence(8, 0, 5, "walk")`

## rand.passwords

`rand.passwords(count: int = 5, length: int = 20, lower: bool = True, upper: bool = True, digits: bool = True, symbols: bool = True, exclude_ambiguous: bool = False, args: str = "") -> List[str]`

Generates count passwords (1..1000) with crypto/rand using the same rules as password().

| Parameter | Type | Description |
|---|---|---|
| `count` | `int` (default `5`) | How many passwords. |
| `length` | `int` (default `20`) | Characters per password, clamped to 4..128. |
| `lower` | `bool` (default `True`) | Include lowercase letters. |
| `upper` | `bool` (default `True`) | Include uppercase letters. |
| `digits` | `bool` (default `True`) | Include digits. |
| `symbols` | `bool` (default `True`) | Include symbols !@#$%^&*()-_=+[]{};:,.? |
| `exclude_ambiguous` | `bool` (default `False`) | Drop look-alike characters 0 O 1 l I \| ` ' ". |
| `args` | `str` (default `""`) | Launcher shortcut, free text: '24 nosym nonum noupper nolower noamb'. When non-empty it overrides the matching parameters. |

Returns `List[str]`. Example: `rand.passwords(5, 16)`

## rand.pick

`rand.pick(items: List[str], n: int = 1, unique: bool = True, seed: int = 0) -> List[str]`

Picks n elements from items. With unique=true (default) no element is picked twice by position and n may not exceed the list length; with unique=false picks may repeat.

| Parameter | Type | Description |
|---|---|---|
| `items` | `List[str]` | Candidates. |
| `n` | `int` (default `1`) | How many to pick. |
| `unique` | `bool` (default `True`) | Do not pick the same position twice. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `List[str]`. Example: `rand.pick(["a", "b", "c", "d"], 2)`

## rand.port

`rand.port(kind: str = "dynamic", seed: int = 0) -> int`

Returns a random TCP/UDP port. kind: any (1..65535), well_known (1..1023), registered (1024..49151), dynamic (49152..65535) or user (1024..65535).

| Parameter | Type | Description |
|---|---|---|
| `kind` | `str` (default `"dynamic"`) | any, well_known, registered, dynamic or user. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `int`. Example: `rand.port("registered")`

## rand.random_date

`rand.random_date(start: str = "1970-01-01", end: str = "2030-12-31", with_time: bool = False, seed: int = 0) -> str`

Returns a random calendar date between start and end inclusive (YYYY-MM-DD or RFC 3339 inputs, UTC). With with_time=true the result is an RFC 3339 timestamp, otherwise YYYY-MM-DD. Errors when start is after end.

| Parameter | Type | Description |
|---|---|---|
| `start` | `str` (default `"1970-01-01"`) | Earliest date. |
| `end` | `str` (default `"2030-12-31"`) | Latest date. |
| `with_time` | `bool` (default `False`) | Include a random time of day. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.random_date("2020-01-01", "2020-12-31")`

## rand.random_time

`rand.random_time(start: str = "00:00", end: str = "23:59", seconds: bool = False, seed: int = 0) -> str`

Returns a random time of day between start and end (HH:MM or HH:MM:SS, inclusive) as HH:MM, or HH:MM:SS when seconds=true. Errors when start is after end.

| Parameter | Type | Description |
|---|---|---|
| `start` | `str` (default `"00:00"`) | Earliest time. |
| `end` | `str` (default `"23:59"`) | Latest time. |
| `seconds` | `bool` (default `False`) | Include seconds. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.random_time("09:00", "17:30")`

## rand.roll_stats

`rand.roll_stats(spec: str) -> Dict[str, str]`

Returns min, max and mean of a dice expression. Mean is exact for plain dice and estimated from 20000 deterministic trials when keep/drop or exploding dice are used; max is "unbounded" for exploding dice (capped at 100 re-rolls per die when rolling).

| Parameter | Type | Description |
|---|---|---|
| `spec` | `str` | Dice notation such as 4d6kh3. |

Returns `Dict[str, str]`. Example: `rand.roll_stats("4d6kh3")`

## rand.sample

`rand.sample(items: List[str], k: int = 1, seed: int = 0) -> List[str]`

Returns a random sample of k distinct positions of items (sampling without replacement), in random order.

| Parameter | Type | Description |
|---|---|---|
| `items` | `List[str]` | Population. |
| `k` | `int` (default `1`) | Sample size. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `List[str]`. Example: `rand.sample(["a", "b", "c"], 2)`

## rand.token_base32

`rand.token_base32(bytes: int = 32) -> str`

Returns bytes cryptographically random bytes as unpadded uppercase RFC 4648 base32, handy for TOTP secrets.

| Parameter | Type | Description |
|---|---|---|
| `bytes` | `int` (default `32`) | Number of random bytes (1..50000). |

Returns `str`. Example: `rand.token_base32(20)`

## rand.token_base64url

`rand.token_base64url(bytes: int = 32) -> str`

Returns bytes cryptographically random bytes as unpadded URL-safe base64.

| Parameter | Type | Description |
|---|---|---|
| `bytes` | `int` (default `32`) | Number of random bytes (1..50000). |

Returns `str`. Example: `rand.token_base64url(32)`

## rand.token_hex

`rand.token_hex(bytes: int = 32) -> str`

Returns bytes cryptographically random bytes as lowercase hex (2 characters per byte).

| Parameter | Type | Description |
|---|---|---|
| `bytes` | `int` (default `32`) | Number of random bytes (1..50000). |

Returns `str`. Example: `rand.token_hex(16)`

## rand.weighted_choice

`rand.weighted_choice(items: List[str], weights: List[float], seed: int = 0) -> str`

Picks one item with probability proportional to its weight. Weights must match items in length, be non-negative and not all zero.

| Parameter | Type | Description |
|---|---|---|
| `items` | `List[str]` | Candidates. |
| `weights` | `List[float]` | One non-negative weight per item. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.weighted_choice(["common", "rare"], [9, 1])`

## rand.int

`rand.int(min: int = 0, max: int = 100, seed: int = 0, args: str = "") -> int`

Returns a uniformly random integer between min and max, both inclusive. With one number in args (launcher) the range is 1..n; errors when min is greater than max.

| Parameter | Type | Description |
|---|---|---|
| `min` | `int` (default `0`) | Lowest possible value. |
| `max` | `int` (default `100`) | Highest possible value. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: '6' means 1..6, '10 20' means 10..20. When non-empty it overrides the matching parameters. |

Returns `int`. Example: `rand.int(1, 6)`

## rand.float

`rand.float(min: float = 0.0, max: float = 1.0, decimals: int = -1, seed: int = 0, args: str = "") -> float`

Returns a random float in \[min, max) rounded to the given number of decimals (0..15; negative keeps full precision).

| Parameter | Type | Description |
|---|---|---|
| `min` | `float` (default `0.0`) | Lower bound. |
| `max` | `float` (default `1.0`) | Upper bound. |
| `decimals` | `int` (default `-1`) | Decimal places to round to; -1 means no rounding. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: '1 10' means a float between 1 and 10. When non-empty it overrides the matching parameters. |

Returns `float`. Example: `rand.float(0, 10, 2)`

## rand.coin_flip

`rand.coin_flip(seed: int = 0, args: str = "") -> str`

Flips one fair coin and returns "heads" or "tails".

| Parameter | Type | Description |
|---|---|---|
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: ignored. When non-empty it overrides the matching parameters. |

Returns `str`. Example: `rand.coin_flip()`

## rand.choice

`rand.choice(items: List[str], seed: int = 0) -> str`

Returns one random element of items; errors on an empty list.

| Parameter | Type | Description |
|---|---|---|
| `items` | `List[str]` | Candidates. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.choice(["red", "green", "blue"])`

## rand.string

`rand.string(length: int = 16, charset: str = "alnum", seed: int = 0) -> str`

Returns a random string of length characters drawn from a named charset (alnum, alpha, digits, hex, upperhex, lower, upper, symbols, base32, base58, base64url, binary, all) or from any custom characters you pass. Not for secrets unless seed is 0.

| Parameter | Type | Description |
|---|---|---|
| `length` | `int` (default `16`) | Characters (0..100000). |
| `charset` | `str` (default `"alnum"`) | Charset name or the literal characters to draw from. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.string(12, "hex")`

## rand.password

`rand.password(length: int = 20, lower: bool = True, upper: bool = True, digits: bool = True, symbols: bool = True, exclude_ambiguous: bool = False, args: str = "") -> str`

Generates a password with crypto/rand, guaranteeing at least one character of every enabled class (lowercase, uppercase, digits, symbols). Length is clamped to 4..128 and if every class is disabled lowercase is used.

| Parameter | Type | Description |
|---|---|---|
| `length` | `int` (default `20`) | Characters per password, clamped to 4..128. |
| `lower` | `bool` (default `True`) | Include lowercase letters. |
| `upper` | `bool` (default `True`) | Include uppercase letters. |
| `digits` | `bool` (default `True`) | Include digits. |
| `symbols` | `bool` (default `True`) | Include symbols !@#$%^&*()-_=+[]{};:,.? |
| `exclude_ambiguous` | `bool` (default `False`) | Drop look-alike characters 0 O 1 l I \| ` ' ". |
| `args` | `str` (default `""`) | Launcher shortcut, free text: '24 nosym nonum noupper nolower noamb'. When non-empty it overrides the matching parameters. |

Returns `str`. Example: `rand.password(24, symbols=False)`

## rand.shuffle

`rand.shuffle(items: List[str], seed: int = 0) -> List[str]`

Returns the items in random order (Fisher-Yates); the input list is not modified.

| Parameter | Type | Description |
|---|---|---|
| `items` | `List[str]` | Items to shuffle. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `List[str]`. Example: `rand.shuffle(["a", "b", "c"])`

## rand.passphrase

`rand.passphrase(words: int = 5, separator: str = "-", capitalize: bool = False, number: bool = False, args: str = "") -> str`

Generates a passphrase of random common words using crypto/rand, e.g. "gold-tree-moon-hand-wave". Optionally capitalizes words and appends a two-digit number.

| Parameter | Type | Description |
|---|---|---|
| `words` | `int` (default `5`) | Number of words (1..64). |
| `separator` | `str` (default `"-"`) | Text between words. |
| `capitalize` | `bool` (default `False`) | Capitalize each word. |
| `number` | `bool` (default `False`) | Append a number 0..99. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: a word count such as '6'. When non-empty it overrides the matching parameters. |

Returns `str`. Example: `rand.passphrase(4, " ", True)`

## rand.shuffle_text

`rand.shuffle_text(text: str, by: str = "chars", seed: int = 0) -> str`

Randomly reorders the characters, words or lines of text according to by (chars, words, lines). Words are re-joined with single spaces.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Text to shuffle. |
| `by` | `str` (default `"chars"`) | Unit to shuffle. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.shuffle_text("one two three", "words")`

## rand.dice

`rand.dice(spec: str = "d6", seed: int = 0) -> Dict[str, str]`

Rolls dice in standard notation: NdM, sums of terms (2d6+3, 1d8+1d4-1), keep/drop (4d6kh3, 2d20kl1, 5d6dl1) and exploding dice (3d6!). Returns spec, total, the kept rolls and a detail line where dropped dice are in parentheses.

| Parameter | Type | Description |
|---|---|---|
| `spec` | `str` (default `"d6"`) | Dice notation such as d20, 2d6+3, 4d6kh3 or 3d6!. Max 1000 dice per term, 10000 in total. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `Dict[str, str]`. Example: `rand.dice("4d6kh3")`

## rand.pin

`rand.pin(digits: int = 6, args: str = "") -> str`

Generates a numeric PIN of the given number of digits (1..64) using crypto/rand; leading zeros are kept.

| Parameter | Type | Description |
|---|---|---|
| `digits` | `int` (default `6`) | Number of digits. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: a digit count such as '8'. When non-empty it overrides the matching parameters. |

Returns `str`. Example: `rand.pin(6)`

## rand.random_item_from_text

`rand.random_item_from_text(text: str, sep: str = "", seed: int = 0) -> str`

Splits text into items and returns one at random. With sep empty it splits by lines, else commas/semicolons, else whitespace; blank items are dropped.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Items separated by newlines, commas or spaces. |
| `sep` | `str` (default `""`) | Explicit separator; empty auto-detects. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.random_item_from_text("pizza, sushi, tacos")`

## rand.dice_total

`rand.dice_total(spec: str = "d6", seed: int = 0) -> int`

Rolls dice notation like dice() but returns only the integer total, handy for further arithmetic.

| Parameter | Type | Description |
|---|---|---|
| `spec` | `str` (default `"d6"`) | Dice notation such as 2d6+3. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `int`. Example: `rand.dice_total("2d6+3")`

## rand.spin_wheel

`rand.spin_wheel(text: str, seed: int = 0) -> str`

Spins a wheel of fortune over the items in text (lines, commas or words) and returns the winner. Errors when there are no items.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | Items on the wheel. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.spin_wheel("alice, bob, carol")`

## rand.yes_no_maybe

`rand.yes_no_maybe(maybe: bool = True, seed: int = 0, args: str = "") -> str`

Answers a question with a random "Yes", "No" or "Maybe" (equal odds). With maybe=false only Yes or No.

| Parameter | Type | Description |
|---|---|---|
| `maybe` | `bool` (default `True`) | Include Maybe as an answer. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: the question (ignored). When non-empty it overrides the matching parameters. |

Returns `str`. Example: `rand.yes_no_maybe()`

## rand.magic_8_ball

`rand.magic_8_ball(question: str = "", seed: int = 0) -> str`

Returns one of the 20 classic Magic 8-Ball answers at random; the question is only for show.

| Parameter | Type | Description |
|---|---|---|
| `question` | `str` (default `""`) | Your question (not used). |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |

Returns `str`. Example: `rand.magic_8_ball("Will it ship?")`

## rand.d100

`rand.d100(modifier: int = 0, seed: int = 0, args: str = "") -> int`

Rolls one 100-sided die (1..100) and adds an optional modifier. In args (launcher) use "+5" for a modifier and "adv" or "dis" to roll twice and keep the higher or lower result.

| Parameter | Type | Description |
|---|---|---|
| `modifier` | `int` (default `0`) | Added to the roll. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: '+3', 'adv', 'dis' or 'adv +2'. When non-empty it overrides the matching parameters. |

Returns `int`. Example: `rand.d100()`

## rand.d20

`rand.d20(modifier: int = 0, seed: int = 0, args: str = "") -> int`

Rolls one 20-sided die (1..20) and adds an optional modifier. In args (launcher) use "+5" for a modifier and "adv" or "dis" to roll twice and keep the higher or lower result.

| Parameter | Type | Description |
|---|---|---|
| `modifier` | `int` (default `0`) | Added to the roll. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: '+3', 'adv', 'dis' or 'adv +2'. When non-empty it overrides the matching parameters. |

Returns `int`. Example: `rand.d20()`

## rand.d6

`rand.d6(modifier: int = 0, seed: int = 0, args: str = "") -> int`

Rolls one 6-sided die (1..6) and adds an optional modifier. In args (launcher) use "+5" for a modifier and "adv" or "dis" to roll twice and keep the higher or lower result.

| Parameter | Type | Description |
|---|---|---|
| `modifier` | `int` (default `0`) | Added to the roll. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: '+3', 'adv', 'dis' or 'adv +2'. When non-empty it overrides the matching parameters. |

Returns `int`. Example: `rand.d6()`

## rand.color_hex

`rand.color_hex(pure: bool = False, seed: int = 0, args: str = "") -> str`

Returns a random colour as #rrggbb. By default a pleasant vivid colour; pure=true draws uniform RGB.

| Parameter | Type | Description |
|---|---|---|
| `pure` | `bool` (default `False`) | Pure random RGB instead of a pleasant vivid colour. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: ignored. When non-empty it overrides the matching parameters. |

Returns `str`. Example: `rand.color_hex()`

## rand.full_name

`rand.full_name(gender: str = "any", seed: int = 0, args: str = "") -> str`

Returns a random "First Last" name from small built-in lists; gender is any, male or female.

| Parameter | Type | Description |
|---|---|---|
| `gender` | `str` (default `"any"`) | any, male or female. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: male, female or any. When non-empty it overrides the matching parameters. |

Returns `str`. Example: `rand.full_name()`

## rand.lottery

`rand.lottery(count: int = 6, max: int = 49, bonus: int = 0, seed: int = 0, args: str = "") -> Dict[str, str]`

Draws count distinct numbers from 1..max in ascending order, plus an optional bonus number drawn from the remaining pool. Returns numbers (space separated) and bonus.

| Parameter | Type | Description |
|---|---|---|
| `count` | `int` (default `6`) | Numbers to draw. |
| `max` | `int` (default `49`) | Highest number in the pool. |
| `bonus` | `int` (default `0`) | Bonus balls to draw (0 or more) from the rest of the pool. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: '6 49' or '5 69 1'. When non-empty it overrides the matching parameters. |

Returns `Dict[str, str]`. Example: `rand.lottery(6, 49, 1)`

## rand.random_emoji

`rand.random_emoji(n: int = 1, seed: int = 0, args: str = "") -> str`

Returns n random emoji (1..1000) from a built-in set of faces, animals, food, objects and symbols, concatenated without separator.

| Parameter | Type | Description |
|---|---|---|
| `n` | `int` (default `1`) | How many emoji. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: a count such as '5'. When non-empty it overrides the matching parameters. |

Returns `str`. Example: `rand.random_emoji(3)`

## rand.d10

`rand.d10(modifier: int = 0, seed: int = 0, args: str = "") -> int`

Rolls one 10-sided die (1..10) and adds an optional modifier. In args (launcher) use "+5" for a modifier and "adv" or "dis" to roll twice and keep the higher or lower result.

| Parameter | Type | Description |
|---|---|---|
| `modifier` | `int` (default `0`) | Added to the roll. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: '+3', 'adv', 'dis' or 'adv +2'. When non-empty it overrides the matching parameters. |

Returns `int`. Example: `rand.d10()`

## rand.d12

`rand.d12(modifier: int = 0, seed: int = 0, args: str = "") -> int`

Rolls one 12-sided die (1..12) and adds an optional modifier. In args (launcher) use "+5" for a modifier and "adv" or "dis" to roll twice and keep the higher or lower result.

| Parameter | Type | Description |
|---|---|---|
| `modifier` | `int` (default `0`) | Added to the roll. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: '+3', 'adv', 'dis' or 'adv +2'. When non-empty it overrides the matching parameters. |

Returns `int`. Example: `rand.d12()`

## rand.d4

`rand.d4(modifier: int = 0, seed: int = 0, args: str = "") -> int`

Rolls one 4-sided die (1..4) and adds an optional modifier. In args (launcher) use "+5" for a modifier and "adv" or "dis" to roll twice and keep the higher or lower result.

| Parameter | Type | Description |
|---|---|---|
| `modifier` | `int` (default `0`) | Added to the roll. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: '+3', 'adv', 'dis' or 'adv +2'. When non-empty it overrides the matching parameters. |

Returns `int`. Example: `rand.d4()`

## rand.d8

`rand.d8(modifier: int = 0, seed: int = 0, args: str = "") -> int`

Rolls one 8-sided die (1..8) and adds an optional modifier. In args (launcher) use "+5" for a modifier and "adv" or "dis" to roll twice and keep the higher or lower result.

| Parameter | Type | Description |
|---|---|---|
| `modifier` | `int` (default `0`) | Added to the roll. |
| `seed` | `int` (default `0`) | 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible. |
| `args` | `str` (default `""`) | Launcher shortcut, free text: '+3', 'adv', 'dis' or 'adv +2'. When non-empty it overrides the matching parameters. |

Returns `int`. Example: `rand.d8()`


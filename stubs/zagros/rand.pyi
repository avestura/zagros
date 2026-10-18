"""Random generation: Random numbers, dice, strings, passwords, passphrases, colors and picks, with optional seeds.

Generated from the tools registry; do not edit by hand.
Type description, not runnable Starlark: do not `import` anything. The name `rand` is predeclared in main.star; call `rand.<name>(...)` directly.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
Keywords: random, generate random, dice, password generator, passphrase, token, pick, shuffle, sample, random color, random date, fake data, lorem, coin flip, lottery
"""

from typing import Any, Dict, List

def bool(probability: float = 0.5, seed: int = 0) -> bool:
    """Returns true with the given probability (default 0.5) and false otherwise.

        probability: Chance of true between 0 and 1.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.bool(0.25)
    Keywords: random boolean, true false, bernoulli, chance
    Needs: none (pure function).
    """

def bytes_hex(count: int = 16, seed: int = 0) -> str:
    """Returns count random bytes as hex. Uses crypto/rand when seed is 0, otherwise a reproducible pseudo-random stream (not for secrets).

        count: Number of bytes (1..50000).
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.bytes_hex(8)
    Keywords: random bytes, hex dump, nonce, iv, salt
    Needs: none (pure function).
    """

def coin_flips(n: int = 10, seed: int = 0) -> Dict[str, str]:
    """Flips n fair coins (0..10000) and returns the counts of heads and tails plus the H/T sequence (sequence only when n <= 200).

        n: Number of flips.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.coin_flips(10)
    Keywords: coin toss, heads, tails, streak, binomial
    Needs: none (pure function).
    """

def color_hsl(pure: bool = False, seed: int = 0) -> str:
    """Returns a random colour as a CSS hsl(h, s%, l%) string; pleasant vivid by default, uniform with pure=true.

        pure: Pure random RGB instead of a pleasant vivid colour.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.color_hsl()
    Keywords: random color, hsl, css, hue
    Needs: none (pure function).
    """

def color_name(seed: int = 0) -> str:
    """Returns a random CSS colour name with its hex value, e.g. "coral #ff7f50".

        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.color_name()
    Keywords: css color names, named color, color
    Needs: none (pure function).
    """

def color_palette(n: int = 5, seed: int = 0) -> List[str]:
    """Returns n (1..100) harmonious hex colours: hues are spaced by the golden angle from a random start so neighbours stay distinct.

        n: Number of colours.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.color_palette(5)
    Keywords: palette, color scheme, theme, swatches, design
    Needs: none (pure function).
    """

def color_rgb(pure: bool = False, seed: int = 0) -> str:
    """Returns a random colour as a CSS rgb(r, g, b) string; pleasant vivid by default, uniform with pure=true.

        pure: Pure random RGB instead of a pleasant vivid colour.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.color_rgb()
    Keywords: random color, rgb, css
    Needs: none (pure function).
    """

def email_fake(domain: str = "example.com", seed: int = 0) -> str:
    """Returns a plausible fake email such as "jane.smith42@example.com". The default domain example.com is reserved for documentation so no real mailbox is hit.

        domain: Domain part.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.email_fake("example.org")
    Keywords: fake email, test data, dummy address, dummy email, placeholder, mock
    Needs: none (pure function).
    """

def ints(count: int = 10, min: int = 0, max: int = 100, unique: bool = False, seed: int = 0) -> List[float]:
    """Returns count random integers in [min, max]; with unique=true no value repeats (errors when the range is too small). Values above 2^53 lose precision in the float list.

        count: How many numbers (0..10000).
        min: Lowest value.
        max: Highest value.
        unique: Forbid repeated values.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.ints(5, 1, 49, unique=True)
    Keywords: random numbers, list, unique, distinct, lottery
    Needs: none (pure function).
    """

def ip_v4(private: bool = False, seed: int = 0) -> str:
    """Returns a random IPv4 address. private=true picks from 10/8, 172.16/12 or 192.168/16; otherwise a routable public-looking address that avoids loopback, multicast, link-local, CGNAT, documentation and reserved ranges.

        private: Use an RFC 1918 private address.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.ip_v4(private=True)
    Keywords: ip address, ipv4, rfc1918, test data, network
    Needs: none (pure function).
    """

def ip_v6(private: bool = False, seed: int = 0) -> str:
    """Returns a random IPv6 address in compressed form. private=true gives a unique-local fd00::/8 address; otherwise a global unicast address in 2000::/3 outside the documentation prefix.

        private: Use a unique local (fd00::/8) address.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.ip_v6()
    Keywords: ip address, ipv6, ula, test data, network
    Needs: none (pure function).
    """

def lorem_word(seed: int = 0) -> str:
    """Returns one random lorem ipsum word. For paragraphs use the text group.

        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.lorem_word()
    Keywords: lorem ipsum, placeholder, dummy text
    Needs: none (pure function).
    """

def lorem_words(count: int = 5, capitalize: bool = False, seed: int = 0) -> str:
    """Returns count random lorem ipsum words (0..10000) joined by single spaces; capitalize upper-cases the first letter.

        count: Number of words.
        capitalize: Capitalize the first word.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.lorem_words(6, True)
    Keywords: lorem ipsum, placeholder, dummy text, filler
    Needs: none (pure function).
    """

def mac_address(format: str = "colon", upper: bool = False, seed: int = 0) -> str:
    """Returns a random unicast, locally administered MAC address. Format is colon (aa:bb:..), dash (AA-BB-..), dot (aabb.ccdd.eeff) or plain (aabbccddeeff); upper switches to uppercase.

        format: colon, dash, dot or plain.
        upper: Uppercase hex digits.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.mac_address("dash", upper=True)
    Keywords: mac, hardware address, ethernet, oui, test data
    Needs: none (pure function).
    """

def name_first(gender: str = "any", seed: int = 0) -> str:
    """Returns a random given name from a small built-in list; gender is any, male or female.

        gender: any, male or female.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.name_first("female")
    Keywords: fake first name, test data, random name, given name, fake name, persona
    Needs: none (pure function).
    """

def name_last(seed: int = 0) -> str:
    """Returns a random family name from a small built-in list.

        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.name_last()
    Keywords: surname, family name, fake name, test data
    Needs: none (pure function).
    """

def number_sequence(count: int = 10, start: int = 0, max_step: int = 10, kind: str = "ascending", seed: int = 0) -> List[float]:
    """Generates count integers as an ascending list, a descending list (kind) or a random walk starting at start with steps of at most max_step. Ascending/descending draw from [start, start + count*max_step].

        count: Length of the sequence (0..10000).
        start: First value / lower bound.
        max_step: Largest step between neighbours.
        kind: ascending, descending or walk.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.number_sequence(8, 0, 5, "walk")
    Keywords: random series, monotonic, random walk, test data, time series
    Needs: none (pure function).
    """

def passwords(count: int = 5, length: int = 20, lower: bool = True, upper: bool = True, digits: bool = True, symbols: bool = True, exclude_ambiguous: bool = False, args: str = "") -> List[str]:
    """Generates count passwords (1..1000) with crypto/rand using the same rules as password().

        count: How many passwords.
        length: Characters per password, clamped to 4..128.
        lower: Include lowercase letters.
        upper: Include uppercase letters.
        digits: Include digits.
        symbols: Include symbols !@#$%^&*()-_=+[]{};:,.?
        exclude_ambiguous: Drop look-alike characters 0 O 1 l I | ` ' ".
        args: Launcher shortcut, free text: '24 nosym nonum noupper nolower noamb'. When non-empty it overrides the matching parameters.

    Example: rand.passwords(5, 16)
    Keywords: bulk passwords, many passwords, batch
    Needs: none (pure function).
    """

def pick(items: List[str], n: int = 1, unique: bool = True, seed: int = 0) -> List[str]:
    """Picks n elements from items. With unique=true (default) no element is picked twice by position and n may not exceed the list length; with unique=false picks may repeat.

        items: Candidates.
        n: How many to pick.
        unique: Do not pick the same position twice.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.pick(["a", "b", "c", "d"], 2)
    Keywords: random item from list, choose one, pick, select, winners, draw, raffle
    Needs: none (pure function).
    """

def port(kind: str = "dynamic", seed: int = 0) -> int:
    """Returns a random TCP/UDP port. kind: any (1..65535), well_known (1..1023), registered (1024..49151), dynamic (49152..65535) or user (1024..65535).

        kind: any, well_known, registered, dynamic or user.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.port("registered")
    Keywords: tcp, udp, network port, ephemeral, dev server
    Needs: none (pure function).
    """

def random_date(start: str = "1970-01-01", end: str = "2030-12-31", with_time: bool = False, seed: int = 0) -> str:
    """Returns a random calendar date between start and end inclusive (YYYY-MM-DD or RFC 3339 inputs, UTC). With with_time=true the result is an RFC 3339 timestamp, otherwise YYYY-MM-DD. Errors when start is after end.

        start: Earliest date.
        end: Latest date.
        with_time: Include a random time of day.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.random_date("2020-01-01", "2020-12-31")
    Keywords: random timestamp, fake date, random day, birthday, test data, timestamp
    Needs: none (pure function).
    """

def random_time(start: str = "00:00", end: str = "23:59", seconds: bool = False, seed: int = 0) -> str:
    """Returns a random time of day between start and end (HH:MM or HH:MM:SS, inclusive) as HH:MM, or HH:MM:SS when seconds=true. Errors when start is after end.

        start: Earliest time.
        end: Latest time.
        seconds: Include seconds.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.random_time("09:00", "17:30")
    Keywords: random clock time, hour minute, test data
    Needs: none (pure function).
    """

def roll_stats(spec: str) -> Dict[str, str]:
    """Returns min, max and mean of a dice expression. Mean is exact for plain dice and estimated from 20000 deterministic trials when keep/drop or exploding dice are used; max is "unbounded" for exploding dice (capped at 100 re-rolls per die when rolling).

        spec: Dice notation such as 4d6kh3.

    Example: rand.roll_stats("4d6kh3")
    Keywords: dice odds, average, expected value, probability, dnd
    Needs: none (pure function).
    """

def sample(items: List[str], k: int = 1, seed: int = 0) -> List[str]:
    """Returns a random sample of k distinct positions of items (sampling without replacement), in random order.

        items: Population.
        k: Sample size.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.sample(["a", "b", "c"], 2)
    Keywords: random subset, pick n at random, sample, without replacement, subset
    Needs: none (pure function).
    """

def token_base32(bytes: int = 32) -> str:
    """Returns bytes cryptographically random bytes as unpadded uppercase RFC 4648 base32, handy for TOTP secrets.

        bytes: Number of random bytes (1..50000).

    Example: rand.token_base32(20)
    Keywords: token, secret, api key, random bytes, crypto, session id, csprng
    Needs: none (pure function).
    """

def token_base64url(bytes: int = 32) -> str:
    """Returns bytes cryptographically random bytes as unpadded URL-safe base64.

        bytes: Number of random bytes (1..50000).

    Example: rand.token_base64url(32)
    Keywords: url safe token, session id, random token, token, secret, api key, random bytes, crypto, csprng
    Needs: none (pure function).
    """

def token_hex(bytes: int = 32) -> str:
    """Returns bytes cryptographically random bytes as lowercase hex (2 characters per byte).

        bytes: Number of random bytes (1..50000).

    Example: rand.token_hex(16)
    Task: make a random hex token
    Keywords: api key, secret key, random bytes as hex, nonce, token, secret, random bytes, crypto, session id, csprng
    Needs: none (pure function).
    """

def weighted_choice(items: List[str], weights: List[float], seed: int = 0) -> str:
    """Picks one item with probability proportional to its weight. Weights must match items in length, be non-negative and not all zero.

        items: Candidates.
        weights: One non-negative weight per item.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.weighted_choice(["common", "rare"], [9, 1])
    Keywords: weighted, probability, loot table, biased
    Needs: none (pure function).
    """

def int(min: int = 0, max: int = 100, seed: int = 0, args: str = "") -> int:
    """Returns a uniformly random integer between min and max, both inclusive. With one number in args (launcher) the range is 1..n; errors when min is greater than max.

        min: Lowest possible value.
        max: Highest possible value.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: '6' means 1..6, '10 20' means 10..20. When non-empty it overrides the matching parameters.

    Example: rand.int(1, 6)
    Task: pick a random integer
    Keywords: random number, random between, rng, integer, between, range, randint
    Needs: none (pure function).
    """

def float(min: float = 0.0, max: float = 1.0, decimals: int = -1, seed: int = 0, args: str = "") -> float:
    """Returns a random float in [min, max) rounded to the given number of decimals (0..15; negative keeps full precision).

        min: Lower bound.
        max: Upper bound.
        decimals: Decimal places to round to; -1 means no rounding.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: '1 10' means a float between 1 and 10. When non-empty it overrides the matching parameters.

    Example: rand.float(0, 10, 2)
    Keywords: random decimal, random fraction, float, double, uniform
    Needs: none (pure function).
    """

def coin_flip(seed: int = 0, args: str = "") -> str:
    """Flips one fair coin and returns "heads" or "tails".

        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: ignored. When non-empty it overrides the matching parameters.

    Example: rand.coin_flip()
    Task: flip a coin
    Keywords: heads or tails, coin toss, heads, tails, flip
    Needs: none (pure function).
    """

def choice(items: List[str], seed: int = 0) -> str:
    """Returns one random element of items; errors on an empty list.

        items: Candidates.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.choice(["red", "green", "blue"])
    Task: pick one at random
    Keywords: random pick, random element, choose randomly, pick one, select, choose
    Needs: none (pure function).
    """

def string(length: int = 16, charset: str = "alnum", seed: int = 0) -> str:
    """Returns a random string of length characters drawn from a named charset (alnum, alpha, digits, hex, upperhex, lower, upper, symbols, base32, base58, base64url, binary, all) or from any custom characters you pass. Not for secrets unless seed is 0.

        length: Characters (0..100000).
        charset: Charset name or the literal characters to draw from.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.string(12, "hex")
    Task: generate a random string
    Keywords: random text, random characters, random id, alphanumeric, generator, id
    Needs: none (pure function).
    """

def password(length: int = 20, lower: bool = True, upper: bool = True, digits: bool = True, symbols: bool = True, exclude_ambiguous: bool = False, args: str = "") -> str:
    """Generates a password with crypto/rand, guaranteeing at least one character of every enabled class (lowercase, uppercase, digits, symbols). Length is clamped to 4..128 and if every class is disabled lowercase is used.

        length: Characters per password, clamped to 4..128.
        lower: Include lowercase letters.
        upper: Include uppercase letters.
        digits: Include digits.
        symbols: Include symbols !@#$%^&*()-_=+[]{};:,.?
        exclude_ambiguous: Drop look-alike characters 0 O 1 l I | ` ' ".
        args: Launcher shortcut, free text: '24 nosym nonum noupper nolower noamb'. When non-empty it overrides the matching parameters.

    Example: rand.password(24, symbols=False)
    Task: generate a password
    Keywords: password generator, strong password, random password, secret, secure password, generator, credentials
    Needs: none (pure function).
    """

def shuffle(items: List[str], seed: int = 0) -> List[str]:
    """Returns the items in random order (Fisher-Yates); the input list is not modified.

        items: Items to shuffle.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.shuffle(["a", "b", "c"])
    Task: shuffle a list
    Keywords: randomize order, scramble list, randomize, permute, reorder, mix
    Needs: none (pure function).
    """

def passphrase(words: int = 5, separator: str = "-", capitalize: bool = False, number: bool = False, args: str = "") -> str:
    """Generates a passphrase of random common words using crypto/rand, e.g. "gold-tree-moon-hand-wave". Optionally capitalizes words and appends a two-digit number.

        words: Number of words (1..64).
        separator: Text between words.
        capitalize: Capitalize each word.
        number: Append a number 0..99.
        args: Launcher shortcut, free text: a word count such as '6'. When non-empty it overrides the matching parameters.

    Example: rand.passphrase(4, " ", True)
    Task: generate a passphrase
    Keywords: diceware, random words, xkcd, memorable password, words
    Needs: none (pure function).
    """

def shuffle_text(text: str, by: str = "chars", seed: int = 0) -> str:
    """Randomly reorders the characters, words or lines of text according to by (chars, words, lines). Words are re-joined with single spaces.

        text: Text to shuffle.
        by: Unit to shuffle.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.shuffle_text("one two three", "words")
    Keywords: scramble, anagram, randomize text, mix letters
    Needs: none (pure function).
    """

def dice(spec: str = "d6", seed: int = 0) -> Dict[str, str]:
    """Rolls dice in standard notation: NdM, sums of terms (2d6+3, 1d8+1d4-1), keep/drop (4d6kh3, 2d20kl1, 5d6dl1) and exploding dice (3d6!). Returns spec, total, the kept rolls and a detail line where dropped dice are in parentheses.

        spec: Dice notation such as d20, 2d6+3, 4d6kh3 or 3d6!. Max 1000 dice per term, 10000 in total.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.dice("4d6kh3")
    Task: roll dice
    Keywords: d6, d20, tabletop, ndm notation, dice roller, dnd, rpg, 2d6, keep highest, advantage, exploding
    Needs: none (pure function).
    """

def pin(digits: int = 6, args: str = "") -> str:
    """Generates a numeric PIN of the given number of digits (1..64) using crypto/rand; leading zeros are kept.

        digits: Number of digits.
        args: Launcher shortcut, free text: a digit count such as '8'. When non-empty it overrides the matching parameters.

    Example: rand.pin(6)
    Keywords: numeric code, otp, verification code, passcode
    Needs: none (pure function).
    """

def random_item_from_text(text: str, sep: str = "", seed: int = 0) -> str:
    """Splits text into items and returns one at random. With sep empty it splits by lines, else commas/semicolons, else whitespace; blank items are dropped.

        text: Items separated by newlines, commas or spaces.
        sep: Explicit separator; empty auto-detects.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.random_item_from_text("pizza, sushi, tacos")
    Keywords: choose from list, decide, random line, pick one
    Needs: none (pure function).
    """

def dice_total(spec: str = "d6", seed: int = 0) -> int:
    """Rolls dice notation like dice() but returns only the integer total, handy for further arithmetic.

        spec: Dice notation such as 2d6+3.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.dice_total("2d6+3")
    Keywords: roll, sum, dice total, 2d6
    Needs: none (pure function).
    """

def spin_wheel(text: str, seed: int = 0) -> str:
    """Spins a wheel of fortune over the items in text (lines, commas or words) and returns the winner. Errors when there are no items.

        text: Items on the wheel.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.spin_wheel("alice, bob, carol")
    Keywords: wheel of fortune, raffle, winner, decide, who goes first
    Needs: none (pure function).
    """

def yes_no_maybe(maybe: bool = True, seed: int = 0, args: str = "") -> str:
    """Answers a question with a random "Yes", "No" or "Maybe" (equal odds). With maybe=false only Yes or No.

        maybe: Include Maybe as an answer.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: the question (ignored). When non-empty it overrides the matching parameters.

    Example: rand.yes_no_maybe()
    Keywords: decide for me, magic answer, random answer, decision, yes no, oracle
    Needs: none (pure function).
    """

def magic_8_ball(question: str = "", seed: int = 0) -> str:
    """Returns one of the 20 classic Magic 8-Ball answers at random; the question is only for show.

        question: Your question (not used).
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.

    Example: rand.magic_8_ball("Will it ship?")
    Keywords: 8 ball, fortune, oracle, answer
    Needs: none (pure function).
    """

def d100(modifier: int = 0, seed: int = 0, args: str = "") -> int:
    """Rolls one 100-sided die (1..100) and adds an optional modifier. In args (launcher) use "+5" for a modifier and "adv" or "dis" to roll twice and keep the higher or lower result.

        modifier: Added to the roll.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: '+3', 'adv', 'dis' or 'adv +2'. When non-empty it overrides the matching parameters.

    Example: rand.d100()
    Keywords: die, roll, dnd, tabletop, advantage, disadvantage, d100
    Needs: none (pure function).
    """

def d20(modifier: int = 0, seed: int = 0, args: str = "") -> int:
    """Rolls one 20-sided die (1..20) and adds an optional modifier. In args (launcher) use "+5" for a modifier and "adv" or "dis" to roll twice and keep the higher or lower result.

        modifier: Added to the roll.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: '+3', 'adv', 'dis' or 'adv +2'. When non-empty it overrides the matching parameters.

    Example: rand.d20()
    Keywords: die, roll, dnd, tabletop, advantage, disadvantage, d20
    Needs: none (pure function).
    """

def d6(modifier: int = 0, seed: int = 0, args: str = "") -> int:
    """Rolls one 6-sided die (1..6) and adds an optional modifier. In args (launcher) use "+5" for a modifier and "adv" or "dis" to roll twice and keep the higher or lower result.

        modifier: Added to the roll.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: '+3', 'adv', 'dis' or 'adv +2'. When non-empty it overrides the matching parameters.

    Example: rand.d6()
    Keywords: die, roll, dnd, tabletop, advantage, disadvantage, d6
    Needs: none (pure function).
    """

def color_hex(pure: bool = False, seed: int = 0, args: str = "") -> str:
    """Returns a random colour as #rrggbb. By default a pleasant vivid colour; pure=true draws uniform RGB.

        pure: Pure random RGB instead of a pleasant vivid colour.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: ignored. When non-empty it overrides the matching parameters.

    Example: rand.color_hex()
    Task: random color
    Keywords: random hex color, random colour, random color, hex color, css, rgb, swatch
    Needs: none (pure function).
    """

def full_name(gender: str = "any", seed: int = 0, args: str = "") -> str:
    """Returns a random "First Last" name from small built-in lists; gender is any, male or female.

        gender: any, male or female.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: male, female or any. When non-empty it overrides the matching parameters.

    Example: rand.full_name()
    Keywords: fake name, test data, persona, dummy user
    Needs: none (pure function).
    """

def lottery(count: int = 6, max: int = 49, bonus: int = 0, seed: int = 0, args: str = "") -> Dict[str, str]:
    """Draws count distinct numbers from 1..max in ascending order, plus an optional bonus number drawn from the remaining pool. Returns numbers (space separated) and bonus.

        count: Numbers to draw.
        max: Highest number in the pool.
        bonus: Bonus balls to draw (0 or more) from the rest of the pool.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: '6 49' or '5 69 1'. When non-empty it overrides the matching parameters.

    Example: rand.lottery(6, 49, 1)
    Keywords: lotto numbers, unique random numbers, lotto, powerball, draw, numbers, jackpot
    Needs: none (pure function).
    """

def random_emoji(n: int = 1, seed: int = 0, args: str = "") -> str:
    """Returns n random emoji (1..1000) from a built-in set of faces, animals, food, objects and symbols, concatenated without separator.

        n: How many emoji.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: a count such as '5'. When non-empty it overrides the matching parameters.

    Example: rand.random_emoji(3)
    Keywords: emoji, smiley, unicode, icon
    Needs: none (pure function).
    """

def d10(modifier: int = 0, seed: int = 0, args: str = "") -> int:
    """Rolls one 10-sided die (1..10) and adds an optional modifier. In args (launcher) use "+5" for a modifier and "adv" or "dis" to roll twice and keep the higher or lower result.

        modifier: Added to the roll.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: '+3', 'adv', 'dis' or 'adv +2'. When non-empty it overrides the matching parameters.

    Example: rand.d10()
    Keywords: die, roll, dnd, tabletop, advantage, disadvantage, d10
    Needs: none (pure function).
    """

def d12(modifier: int = 0, seed: int = 0, args: str = "") -> int:
    """Rolls one 12-sided die (1..12) and adds an optional modifier. In args (launcher) use "+5" for a modifier and "adv" or "dis" to roll twice and keep the higher or lower result.

        modifier: Added to the roll.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: '+3', 'adv', 'dis' or 'adv +2'. When non-empty it overrides the matching parameters.

    Example: rand.d12()
    Keywords: die, roll, dnd, tabletop, advantage, disadvantage, d12
    Needs: none (pure function).
    """

def d4(modifier: int = 0, seed: int = 0, args: str = "") -> int:
    """Rolls one 4-sided die (1..4) and adds an optional modifier. In args (launcher) use "+5" for a modifier and "adv" or "dis" to roll twice and keep the higher or lower result.

        modifier: Added to the roll.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: '+3', 'adv', 'dis' or 'adv +2'. When non-empty it overrides the matching parameters.

    Example: rand.d4()
    Keywords: die, roll, dnd, tabletop, advantage, disadvantage, d4
    Needs: none (pure function).
    """

def d8(modifier: int = 0, seed: int = 0, args: str = "") -> int:
    """Rolls one 8-sided die (1..8) and adds an optional modifier. In args (launcher) use "+5" for a modifier and "adv" or "dis" to roll twice and keep the higher or lower result.

        modifier: Added to the roll.
        seed: 0 (default) draws from crypto/rand on every call; any other value makes the result reproducible.
        args: Launcher shortcut, free text: '+3', 'adv', 'dis' or 'adv +2'. When non-empty it overrides the matching parameters.

    Example: rand.d8()
    Keywords: die, roll, dnd, tabletop, advantage, disadvantage, d8
    Needs: none (pure function).
    """

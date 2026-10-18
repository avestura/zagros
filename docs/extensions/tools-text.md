# Text (`text`)

Change, measure and analyse strings: case styles, palindromes, lengths, normalisation, similarity, Unicode.

No permission is needed: these are pure functions. Generated from the tools registry.

107 functions. Use them as `text.<name>(...)` in any extension.

## text.case_all

`text.case_all(s: str) -> Dict[str, str]`

Returns every case style of the text as a dict: lower, upper, title, sentence, swapcase, alternating, camel, pascal, snake, kebab, constant, dot, path, train, slug.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `Dict[str, str]`. Example: `text.case_all("hello world")`

## text.stats

`text.stats(s: str) -> Dict[str, str]`

Returns a dict with chars, bytes, graphemes, words, lines, sentences, unique_chars, avg_word_length, longest_word, reading_time_seconds and reading_time (at 200 words per minute).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `Dict[str, str]`. Example: `text.stats("Hello world. Bye.")`

## text.lower

`text.lower(s: str) -> str`

Converts every letter to lower case using Unicode rules (Σ -&gt; σ). Other characters are unchanged.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.lower("HELLO World")`

## text.palindrome_report

`text.palindrome_report(s: str, ignore_case: bool = True, ignore_non_alnum: bool = True) -> Dict[str, str]`

Returns a dict with is_palindrome, cleaned (the compared text), reversed and longest_substring (longest palindromic substring of the cleaned text).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `ignore_case` | `bool` (default `True`) | Ignore letter case. |
| `ignore_non_alnum` | `bool` (default `True`) | Ignore spaces, punctuation and symbols. |

Returns `Dict[str, str]`. Example: `text.palindrome_report("Never odd or even")`

## text.upper

`text.upper(s: str) -> str`

Converts every letter to upper case using Unicode simple case mapping. Other characters are unchanged.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.upper("hello World")`

## text.title

`text.title(s: str) -> str`

Capitalises the first letter of every word and lower-cases the rest. Apostrophes inside a word do not start a new word (don't -&gt; Don't).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.title("the quick brown fox")`

## text.camel

`text.camel(s: str) -> str`

Converts to camelCase. Words are split on separators, case changes and acronyms (HTMLParser -&gt; htmlParser); apostrophes are dropped.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.camel("hello world foo")`

## text.snake

`text.snake(s: str) -> str`

Converts to snake_case: lower-case words joined with underscores. Splits on separators, case changes and acronyms.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.snake("Hello World fooBar")`

## text.kebab

`text.kebab(s: str) -> str`

Converts to kebab-case: lower-case words joined with hyphens. Splits on separators, case changes and acronyms.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.kebab("Hello World fooBar")`

## text.reverse

`text.reverse(s: str) -> str`

Reverses the text by grapheme cluster, so combining marks, skin-tone modifiers, ZWJ emoji sequences and flags stay intact.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.reverse("hello")`

## text.slug

`text.slug(s: str, separator: str = "-", ascii: bool = True) -> str`

Makes a URL slug: lower case, accents folded to ASCII (é -&gt; e, ß -&gt; ss) when ascii is true, every run of other characters replaced by one separator, no leading or trailing separator.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `separator` | `str` (default `"-"`) | Separator placed between words. |
| `ascii` | `bool` (default `True`) | Fold accents and drop non-ASCII characters. |

Returns `str`. Example: `text.slug("Héllo, Wörld! 2024")`

## text.sort_words

`text.sort_words(s: str, desc: bool = False, ignore_case: bool = True) -> str`

Sorts whitespace-separated words alphabetically and joins them with single spaces. Use lines.sort_lines for line-based sorting.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `desc` | `bool` (default `False`) | Descending order. |
| `ignore_case` | `bool` (default `True`) | Compare case-insensitively. |

Returns `str`. Example: `text.sort_words("pear apple fig")`

## text.trim

`text.trim(s: str, mode: str = "both", cutset: str = "") -> str`

Removes characters from the ends. mode is both, left or right; cutset lists the characters to remove (empty means any whitespace).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `mode` | `str` (default `"both"`) | Which ends to trim. One of: both, left, right. |
| `cutset` | `str` (default `""`) | Characters to remove; empty means whitespace. |

Returns `str`. Example: `text.trim("  hi  ")`

## text.word_count

`text.word_count(s: str) -> int`

Counts words: runs of letters, digits and combining marks, which may contain apostrophes or hyphens between them (don't and well-known are one word each). Empty or punctuation-only text gives 0.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `int`. Example: `text.word_count("Hello, big world!")`

## text.collapse_whitespace

`text.collapse_whitespace(s: str) -> str`

Replaces every run of whitespace (spaces, tabs, newlines, Unicode spaces) with one space and trims both ends.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.collapse_whitespace("a   b \n c")`

## text.sentence_case

`text.sentence_case(s: str) -> str`

Lower-cases the text, then capitalises the first letter after the start and after each . ! ? or line break.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.sentence_case("HELLO there. HOW ARE you?")`

## text.swapcase

`text.swapcase(s: str) -> str`

Inverts the case of every letter: upper becomes lower and lower becomes upper.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.swapcase("Hello World")`

## text.alternating_case

`text.alternating_case(s: str) -> str`

Alternates lower and upper case over the letters, starting with lower (the 'spongebob' meme case). Non-letters do not advance the alternation.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.alternating_case("hello world")`

## text.pascal

`text.pascal(s: str) -> str`

Converts to PascalCase (UpperCamelCase). Words are split on separators, case changes and acronyms.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.pascal("hello world foo")`

## text.constant

`text.constant(s: str) -> str`

Converts to SCREAMING_SNAKE_CASE: upper-case words joined with underscores.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.constant("hello world")`

## text.dot_case

`text.dot_case(s: str) -> str`

Converts to dot.case: lower-case words joined with dots.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.dot_case("Hello World")`

## text.path_case

`text.path_case(s: str) -> str`

Converts to path/case: lower-case words joined with forward slashes.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.path_case("Hello World")`

## text.train_case

`text.train_case(s: str) -> str`

Converts to Train-Case: capitalised words joined with hyphens (HTTP header style).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.train_case("hello world")`

## text.capitalize_words

`text.capitalize_words(s: str) -> str`

Upper-cases the first character of every whitespace-separated word and leaves the other characters untouched.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.capitalize_words("hello wORLD")`

## text.uncapitalize

`text.uncapitalize(s: str) -> str`

Lower-cases only the first character of the text (Hello -&gt; hello); empty input returns empty.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.uncapitalize("Hello World")`

## text.length

`text.length(s: str) -> int`

Number of characters (Unicode code points / runes) in the text. Emoji with modifiers or combining marks count each code point; see grapheme_length.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `int`. Example: `text.length("héllo")`

## text.byte_length

`text.byte_length(s: str) -> int`

Number of bytes of the text encoded as UTF-8.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `int`. Example: `text.byte_length("héllo")`

## text.grapheme_length

`text.grapheme_length(s: str) -> int`

Number of user-perceived characters. Approximation of extended grapheme clusters: a base rune plus following combining marks, variation selectors, skin-tone modifiers, tag characters and ZWJ-joined runes forms one grapheme; regional-indicator pairs (flags) and CRLF are one grapheme.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `int`. Example: `text.grapheme_length("é👨‍👩‍👧")`

## text.line_count

`text.line_count(s: str) -> int`

Number of lines. CRLF, LF and lone CR all separate lines; a single trailing line break does not add an empty line; empty text has 0 lines.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `int`. Example: `text.line_count("a\nb\nc")`

## text.sentence_count

`text.sentence_count(s: str) -> int`

Counts sentences ended by . ! ? or ... followed by whitespace or the end; trailing text without a terminator counts as one sentence. Decimals like 3.14 do not split.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `int`. Example: `text.sentence_count("Hi. How are you? Fine")`

## text.char_count

`text.char_count(s: str, ch: str, ignore_case: bool = False) -> int`

Counts non-overlapping occurrences of a character or substring. An empty needle is an error.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Text to search. |
| `ch` | `str` | Character or substring to count. |
| `ignore_case` | `bool` (default `False`) | Compare case-insensitively. |

Returns `int`. Example: `text.char_count("banana", "a")`

## text.reverse_words

`text.reverse_words(s: str) -> str`

Reverses the order of whitespace-separated words and joins them with single spaces.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.reverse_words("one two three")`

## text.sort_chars

`text.sort_chars(s: str, desc: bool = False, remove_space: bool = False) -> str`

Sorts the characters by code point (descending when desc is true). Whitespace is kept unless remove_space is true.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `desc` | `bool` (default `False`) | Descending order. |
| `remove_space` | `bool` (default `False`) | Drop whitespace first. |

Returns `str`. Example: `text.sort_chars("hello")`

## text.remove_whitespace

`text.remove_whitespace(s: str) -> str`

Deletes every whitespace character (spaces, tabs, newlines, Unicode spaces).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.remove_whitespace("a b\tc")`

## text.remove_punctuation

`text.remove_punctuation(s: str, symbols: bool = False) -> str`

Deletes Unicode punctuation characters; with symbols=true also deletes symbols such as $ + &lt; = ~.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `symbols` | `bool` (default `False`) | Also remove symbol characters. |

Returns `str`. Example: `text.remove_punctuation("Hi, there!")`

## text.is_palindrome

`text.is_palindrome(s: str, ignore_case: bool = True, ignore_non_alnum: bool = True) -> bool`

True when the text reads the same forwards and backwards. By default case and every non-letter/non-digit are ignored; empty text counts as a palindrome.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `ignore_case` | `bool` (default `True`) | Ignore letter case. |
| `ignore_non_alnum` | `bool` (default `True`) | Ignore spaces, punctuation and symbols. |

Returns `bool`. Example: `text.is_palindrome("A man, a plan, a canal: Panama")`

## text.remove_digits

`text.remove_digits(s: str) -> str`

Deletes every Unicode decimal digit.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.remove_digits("a1b2")`

## text.longest_palindromic_substring

`text.longest_palindromic_substring(s: str) -> str`

Returns the longest substring (case-sensitive, rune-based, first one on ties) that reads the same in both directions; empty text returns empty.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.longest_palindromic_substring("babad")`

## text.only_digits

`text.only_digits(s: str) -> str`

Keeps only Unicode decimal digits and deletes everything else.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.only_digits("tel: 555-1234")`

## text.is_anagram

`text.is_anagram(a: str, b: str, ignore_case: bool = True) -> bool`

True when both texts use the same letters and digits the same number of times. Spaces and punctuation are ignored; case is ignored by default.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First text. |
| `b` | `str` | Second text. |
| `ignore_case` | `bool` (default `True`) | Ignore letter case. |

Returns `bool`. Example: `text.is_anagram("Listen", "Silent")`

## text.only_letters

`text.only_letters(s: str) -> str`

Keeps only letters (and their combining marks) and deletes everything else.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.only_letters("a1 b2!")`

## text.anagram_key

`text.anagram_key(s: str) -> str`

Returns the letters and digits of the text lower-cased and sorted; two texts are anagrams when their keys match.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.anagram_key("Listen")`

## text.only_alnum

`text.only_alnum(s: str) -> str`

Keeps only letters, combining marks and digits.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.only_alnum("a-1 b_2!")`

## text.is_pangram

`text.is_pangram(s: str, alphabet: str = "abcdefghijklmnopqrstuvwxyz") -> bool`

True when the text contains every letter of the alphabet (default a-z, case-insensitive). Pass another alphabet to test other languages.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `alphabet` | `str` (default `"abcdefghijklmnopqrstuvwxyz"`) | Letters that must all appear. |

Returns `bool`. Example: `text.is_pangram("The quick brown fox jumps over the lazy dog")`

## text.remove_non_ascii

`text.remove_non_ascii(s: str) -> str`

Deletes every character with a code point of 128 or more.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.remove_non_ascii("héllo ✓")`

## text.strip_accents

`text.strip_accents(s: str) -> str`

Removes diacritical marks (é -&gt; e, ñ -&gt; n) via NFD decomposition and dropping non-spacing marks. Letters without decomposition (ø, ß, ł) stay unchanged; use slug for transliteration.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.strip_accents("Crème brûlée")`

## text.normalize

`text.normalize(s: str, form: str = "NFC") -> str`

Applies Unicode normalization form NFC, NFD, NFKC or NFKD (case-insensitive name).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `form` | `str` (default `"NFC"`) | Normalization form. One of: NFC, NFD, NFKC, NFKD. |

Returns `str`. Example: `text.normalize("é", "NFC")`

## text.repeat

`text.repeat(s: str, n: int = 2, sep: str = "") -> str`

Repeats the text n times with an optional separator between copies. n of 0 gives an empty string; negative n or a result over 4 MiB is an error.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Text to repeat. |
| `n` | `int` (default `2`) | Number of copies. |
| `sep` | `str` (default `""`) | Separator between copies. |

Returns `str`. Example: `text.repeat("ab", 3, "-")`

## text.center

`text.center(s: str, width: int, fill: str = " ") -> str`

Centers the text in width characters using the fill text; an odd leftover puts the extra fill on the right.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `width` | `int` | Target width in characters. |
| `fill` | `str` (default `" "`) | Fill text, cycled as needed. |

Returns `str`. Example: `text.center("hi", 6, "*")`

## text.pad_left

`text.pad_left(s: str, width: int, fill: str = " ") -> str`

Pads the start with the fill text up to width characters (right-aligns). Text already that wide is unchanged.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `width` | `int` | Target width in characters. |
| `fill` | `str` (default `" "`) | Fill text, cycled as needed. |

Returns `str`. Example: `text.pad_left("7", 3, "0")`

## text.pad_right

`text.pad_right(s: str, width: int, fill: str = " ") -> str`

Pads the end with the fill text up to width characters (left-aligns). Text already that wide is unchanged.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `width` | `int` | Target width in characters. |
| `fill` | `str` (default `" "`) | Fill text, cycled as needed. |

Returns `str`. Example: `text.pad_right("7", 3, ".")`

## text.truncate

`text.truncate(s: str, max: int = 20, ellipsis: str = "…") -> str`

Shortens the text to at most max characters (runes), ending with the ellipsis when cut. If max is smaller than the ellipsis, only the first max characters are returned.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `max` | `int` (default `20`) | Maximum length in characters, including the ellipsis. |
| `ellipsis` | `str` (default `"…"`) | Marker appended when cut. |

Returns `str`. Example: `text.truncate("Hello world", 8)`

## text.is_alpha

`text.is_alpha(s: str) -> bool`

True when the text is non-empty and every character is a letter (combining marks allowed).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `bool`. Example: `text.is_alpha("héllo")`

## text.is_ascii

`text.is_ascii(s: str) -> bool`

True when every character is ASCII (code point below 128). Empty text is ASCII.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `bool`. Example: `text.is_ascii("hello")`

## text.is_blank

`text.is_blank(s: str) -> bool`

True when the text is empty or contains only whitespace (including Unicode spaces).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `bool`. Example: `text.is_blank("  \t")`

## text.is_numeric

`text.is_numeric(s: str) -> bool`

True when the text is non-empty and every character is a Unicode decimal digit (Nd). Signs, dots and spaces make it false.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `bool`. Example: `text.is_numeric("12345")`

## text.truncate_words

`text.truncate_words(s: str, max_words: int = 10, ellipsis: str = "…") -> str`

Keeps the first max_words whitespace-separated words and appends the ellipsis if any were dropped; shorter text is returned unchanged.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `max_words` | `int` (default `10`) | Words to keep. |
| `ellipsis` | `str` (default `"…"`) | Marker appended when cut. |

Returns `str`. Example: `text.truncate_words("one two three four", 2)`

## text.consonant_count

`text.consonant_count(s: str) -> int`

Counts letters that are not vowels (y counts as a consonant); digits and punctuation are ignored. Non-Latin letters count as consonants.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `int`. Example: `text.consonant_count("Education")`

## text.vowel_count

`text.vowel_count(s: str) -> int`

Counts the vowels a e i o u, including their accented forms (é, ü ...), case-insensitive.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `int`. Example: `text.vowel_count("Education")`

## text.wrap

`text.wrap(s: str, width: int = 72, break_long: bool = False) -> str`

Greedy word wrap to width characters per line. Existing line breaks start new paragraphs, runs of spaces collapse, and words longer than width stay whole unless break_long is true.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `width` | `int` (default `72`) | Maximum line width in characters. |
| `break_long` | `bool` (default `False`) | Split words longer than width. |

Returns `str`. Example: `text.wrap("the quick brown fox", 10)`

## text.char_frequency

`text.char_frequency(s: str, include_space: bool = False) -> Dict[str, str]`

Returns a dict mapping each character to how many times it occurs (values are counts as text). Grapheme-level frequency; whitespace is skipped unless include_space is true. Use char_frequency_sorted for ordered output.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `include_space` | `bool` (default `False`) | Count whitespace too. |

Returns `Dict[str, str]`. Example: `text.char_frequency("hello")`

## text.indent

`text.indent(s: str, prefix: str = "    ") -> str`

Prefixes every non-empty line with the given text (four spaces by default). Blank lines are left empty.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `prefix` | `str` (default `"    "`) | Text added to each non-empty line. |

Returns `str`. Example: `text.indent("a\nb", "  ")`

## text.char_frequency_sorted

`text.char_frequency_sorted(s: str, include_space: bool = False) -> List[str]`

Returns lines like 'l: 2' ordered by count descending, then by character ascending. Grapheme-level frequency; whitespace is skipped unless include_space is true.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `include_space` | `bool` (default `False`) | Count whitespace too. |

Returns `List[str]`. Example: `text.char_frequency_sorted("hello")`

## text.unindent

`text.unindent(s: str) -> str`

Removes the leading whitespace shared by all non-blank lines (dedent). Blank lines are ignored when measuring.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.unindent("  a\n    b")`

## text.replace

`text.replace(s: str, old: str, new: str = "", count: int = -1) -> str`

Replaces occurrences of old with new, at most count times (-1 = all). An empty old returns the text unchanged.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `old` | `str` | Text to find. |
| `new` | `str` (default `""`) | Replacement text. |
| `count` | `int` (default `-1`) | Maximum replacements; -1 for all. |

Returns `str`. Example: `text.replace("a-b-c", "-", "+")`

## text.unique_chars

`text.unique_chars(s: str) -> str`

Returns each distinct character once, in order of first appearance (whitespace skipped).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.unique_chars("hello")`

## text.duplicate_chars

`text.duplicate_chars(s: str) -> str`

Returns the characters that occur more than once, each listed once in order of first appearance (whitespace skipped).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.duplicate_chars("hello")`

## text.replace_ci

`text.replace_ci(s: str, old: str, new: str = "", count: int = -1) -> str`

Like replace but matches old case-insensitively; the replacement is inserted as given. An empty old returns the text unchanged.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `old` | `str` | Text to find (any case). |
| `new` | `str` (default `""`) | Replacement text. |
| `count` | `int` (default `-1`) | Maximum replacements; -1 for all. |

Returns `str`. Example: `text.replace_ci("Hello HELLO", "hello", "bye")`

## text.most_common_char

`text.most_common_char(s: str) -> str`

Returns the most frequent character (whitespace skipped); ties go to the smallest character; empty text returns empty.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.most_common_char("hello")`

## text.remove

`text.remove(s: str, sub: str) -> str`

Deletes every occurrence of sub. An empty sub returns the text unchanged.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `sub` | `str` | Text to delete. |

Returns `str`. Example: `text.remove("banana", "an")`

## text.insert

`text.insert(s: str, index: int, text: str) -> str`

Inserts text at a character index (negative counts from the end; out-of-range values clamp to the ends).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `index` | `int` | Character index. |
| `text` | `str` | Text to insert. |

Returns `str`. Example: `text.insert("held", 3, "l")`

## text.delete_range

`text.delete_range(s: str, start: int, end: int) -> str`

Deletes characters from start up to but excluding end (character indexes, negative from the end, clamped). start &gt;= end deletes nothing.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `start` | `int` | First character index to delete. |
| `end` | `int` | Index after the last deleted character. |

Returns `str`. Example: `text.delete_range("hello", 1, 3)`

## text.substring

`text.substring(s: str, start: int = 0, length: int = -1) -> str`

Returns length characters starting at start (character-based; negative start counts from the end; length -1 means to the end). Out-of-range values clamp.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `start` | `int` (default `0`) | Start index; negative counts from the end. |
| `length` | `int` (default `-1`) | Number of characters; -1 for the rest. |

Returns `str`. Example: `text.substring("héllo wörld", 6, 5)`

## text.split_at

`text.split_at(s: str, index: int) -> List[str]`

Splits the text into two parts at a character index (negative from the end, clamped). Always returns exactly two strings.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `index` | `int` | Character index. |

Returns `List[str]`. Example: `text.split_at("hello", 2)`

## text.left

`text.left(s: str, n: int = 1) -> str`

Returns the first n characters (all of them if n is larger; n &lt;= 0 gives empty).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `n` | `int` (default `1`) | Number of characters. |

Returns `str`. Example: `text.left("hello", 2)`

## text.right

`text.right(s: str, n: int = 1) -> str`

Returns the last n characters (all of them if n is larger; n &lt;= 0 gives empty).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `n` | `int` (default `1`) | Number of characters. |

Returns `str`. Example: `text.right("hello", 2)`

## text.remove_prefix

`text.remove_prefix(s: str, prefix: str) -> str`

Removes the prefix when the text starts with it; otherwise returns the text unchanged.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `prefix` | `str` | Prefix to remove. |

Returns `str`. Example: `text.remove_prefix("foobar", "foo")`

## text.remove_suffix

`text.remove_suffix(s: str, suffix: str) -> str`

Removes the suffix when the text ends with it; otherwise returns the text unchanged.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `suffix` | `str` | Suffix to remove. |

Returns `str`. Example: `text.remove_suffix("foobar", "bar")`

## text.leetspeak

`text.leetspeak(s: str) -> str`

Converts to leetspeak with a fixed map: a4 b8 e3 g6 i1 o0 s5 t7 (case-insensitive); other characters are unchanged.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.leetspeak("elite hacker")`

## text.mirror

`text.mirror(s: str) -> str`

Flips the text upside down: reverses it and swaps letters, digits and brackets for look-alike rotated characters (a -&gt; ɐ, b -&gt; q, ? -&gt; ¿). Characters without a mapping stay as they are.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.mirror("Hello")`

## text.remove_zero_width

`text.remove_zero_width(s: str) -> str`

Deletes invisible characters: zero-width space/non-joiner/joiner, word joiner, BOM, soft hyphen, bidi marks and overrides, Mongolian vowel separator, combining grapheme joiner. Note that ZWNJ/ZWJ are meaningful in some scripts and emoji sequences.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.remove_zero_width("a​b")`

## text.show_invisibles

`text.show_invisibles(s: str, show_spaces: bool = False) -> str`

Makes invisible characters visible: \n, \r and \t as escapes; other control, zero-width and special space characters (including no-break space) as &lt;U+XXXX&gt;; regular spaces become · only when show_spaces is true.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `show_spaces` | `bool` (default `False`) | Render ordinary spaces as ·. |

Returns `str`. Example: `text.show_invisibles("a\tb​")`

## text.smart_quotes_to_straight

`text.smart_quotes_to_straight(s: str, punctuation: bool = False) -> str`

Replaces curly quotes (‘ ’ “ ” „ ‚ and primes) with ASCII ' and "; with punctuation=true also converts en/em dashes to - and … to ...

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |
| `punctuation` | `bool` (default `False`) | Also convert dashes and ellipsis. |

Returns `str`. Example: `text.smart_quotes_to_straight("“hi” it’s")`

## text.straight_to_smart

`text.straight_to_smart(s: str) -> str`

Replaces ASCII quotes with typographic ones: opening after the start, whitespace or an opening bracket, closing otherwise; a lone apostrophe inside a word becomes ’.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.straight_to_smart("\"hi\" it's")`

## text.remove_emoji

`text.remove_emoji(s: str) -> str`

Deletes emoji and pictographs (including skin-tone modifiers, flags, variation selectors and the ZWJ glue between emoji). Covers the main emoji blocks, not every symbol.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.remove_emoji("hi 😀 there 👍🏽")`

## text.to_fullwidth

`text.to_fullwidth(s: str) -> str`

Converts ASCII printable characters to their fullwidth forms (A -&gt; Ａ, space -&gt; U+3000).

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.to_fullwidth("Hello 123")`

## text.to_halfwidth

`text.to_halfwidth(s: str) -> str`

Converts fullwidth ASCII forms (U+FF01-FF5E and U+3000) back to normal ASCII. Halfwidth katakana is not converted.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `str`. Example: `text.to_halfwidth("Ｈｅｌｌｏ")`

## text.levenshtein

`text.levenshtein(a: str, b: str) -> int`

Edit distance between two texts: the minimum number of single-character insertions, deletions and substitutions (rune-based). Fails for very large inputs (over 50 million table cells).

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First text. |
| `b` | `str` | Second text. |

Returns `int`. Example: `text.levenshtein("kitten", "sitting")`

## text.damerau_levenshtein

`text.damerau_levenshtein(a: str, b: str) -> int`

Edit distance that also counts a swap of two adjacent characters as one edit (optimal string alignment variant, rune-based). ca -&gt; ac is 1.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First text. |
| `b` | `str` | Second text. |

Returns `int`. Example: `text.damerau_levenshtein("ca", "ac")`

## text.hamming

`text.hamming(a: str, b: str) -> int`

Number of positions at which two texts of equal length (in characters) differ. Different lengths are an error.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First text. |
| `b` | `str` | Second text. |

Returns `int`. Example: `text.hamming("karolin", "kathrin")`

## text.jaro_winkler

`text.jaro_winkler(a: str, b: str) -> float`

Jaro-Winkler similarity from 0 (nothing in common) to 1 (identical), favouring a shared prefix of up to 4 characters; two empty texts give 1.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First text. |
| `b` | `str` | Second text. |

Returns `float`. Example: `text.jaro_winkler("martha", "marhta")`

## text.similarity

`text.similarity(a: str, b: str) -> float`

Similarity from 0 to 1 computed as 1 - levenshtein / longer length; two empty texts give 1.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First text. |
| `b` | `str` | Second text. |

Returns `float`. Example: `text.similarity("kitten", "sitting")`

## text.common_prefix

`text.common_prefix(a: str, b: str) -> str`

Longest prefix shared by both texts (character-based); empty when they differ from the first character.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First text. |
| `b` | `str` | Second text. |

Returns `str`. Example: `text.common_prefix("interview", "internet")`

## text.common_suffix

`text.common_suffix(a: str, b: str) -> str`

Longest suffix shared by both texts (character-based); empty when the last characters differ.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First text. |
| `b` | `str` | Second text. |

Returns `str`. Example: `text.common_suffix("testing", "running")`

## text.longest_common_substring

`text.longest_common_substring(a: str, b: str) -> str`

Longest contiguous piece of text found in both inputs (case-sensitive, first one in the first text on ties). Fails for very large inputs.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First text. |
| `b` | `str` | Second text. |

Returns `str`. Example: `text.longest_common_substring("abcdef", "zcdemn")`

## text.diff_summary

`text.diff_summary(a: str, b: str) -> Dict[str, str]`

Character-level summary of how a becomes b using the longest common subsequence: dict with inserted, deleted, unchanged, distance (inserted + deleted) and similarity (0..1). Fails for very large inputs.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First text. |
| `b` | `str` | Second text. |

Returns `Dict[str, str]`. Example: `text.diff_summary("kitten", "sitting")`

## text.compare_ci

`text.compare_ci(a: str, b: str) -> int`

Compares two texts ignoring case: returns -1 if a sorts before b, 0 if equal, 1 if after.

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First text. |
| `b` | `str` | Second text. |

Returns `int`. Example: `text.compare_ci("Apple", "apple")`

## text.equals_ci

`text.equals_ci(a: str, b: str) -> bool`

True when both texts are equal ignoring case (Unicode simple folding).

| Parameter | Type | Description |
|---|---|---|
| `a` | `str` | First text. |
| `b` | `str` | Second text. |

Returns `bool`. Example: `text.equals_ci("Hello", "hELLO")`

## text.codepoints

`text.codepoints(s: str) -> List[str]`

Lists the Unicode code point of every character as U+XXXX (at least four hex digits, upper case). Invalid UTF-8 bytes appear as U+FFFD.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `List[str]`. Example: `text.codepoints("aé😀")`

## text.from_codepoints

`text.from_codepoints(points: List[str]) -> str`

Builds text from code points written as U+0041, 0x41, \u0041, &#65; (decimal) or plain hex. Invalid values or surrogates are an error.

| Parameter | Type | Description |
|---|---|---|
| `points` | `List[str]` | Code point strings. |

Returns `str`. Example: `text.from_codepoints(["U+0048", "0x69"])`

## text.to_codepoint

`text.to_codepoint(ch: str) -> int`

Returns the decimal Unicode code point of the first character; empty text is an error.

| Parameter | Type | Description |
|---|---|---|
| `ch` | `str` | A character; only the first one is used. |

Returns `int`. Example: `text.to_codepoint("A")`

## text.from_codepoint

`text.from_codepoint(n: int) -> str`

Returns the character for a decimal code point (0 to 0x10FFFF, no surrogates).

| Parameter | Type | Description |
|---|---|---|
| `n` | `int` | Decimal code point. |

Returns `str`. Example: `text.from_codepoint(128512)`

## text.char_info

`text.char_info(ch: str) -> Dict[str, str]`

Describes the first character as a dict: char, codepoint (U+XXXX), decimal, utf8 (hex bytes), utf8_length, utf16 (hex code units), category (e.g. Lu), category_name, block, upper, lower and printable. Character names are not included.

| Parameter | Type | Description |
|---|---|---|
| `ch` | `str` | A character; only the first one is used. |

Returns `Dict[str, str]`. Example: `text.char_info("é")`

## text.unicode_blocks

`text.unicode_blocks(s: str) -> List[str]`

Lists the distinct Unicode blocks (Basic Latin, Cyrillic, Emoticons ...) used by the text, in order of first appearance. Characters in blocks not in the built-in table are reported as Other.

| Parameter | Type | Description |
|---|---|---|
| `s` | `str` | Input text. |

Returns `List[str]`. Example: `text.unicode_blocks("aб😀")`

## text.lorem

`text.lorem(words: int = 50, seed: int = 0) -> str`

Generates lorem ipsum text of exactly the requested number of words (1 to 100000), as sentences. The same seed always gives the same text; seed 0 starts with 'Lorem ipsum dolor sit amet, consectetur adipiscing elit'.

| Parameter | Type | Description |
|---|---|---|
| `words` | `int` (default `50`) | Number of words. |
| `seed` | `int` (default `0`) | Seed for deterministic output; 0 uses the classic opening. |

Returns `str`. Example: `text.lorem(20)`

## text.lorem_paragraphs

`text.lorem_paragraphs(paragraphs: int = 3, sentences: int = 5, seed: int = 0) -> str`

Generates lorem ipsum paragraphs (1 to 1000) of the given sentence count (1 to 100) separated by blank lines. The same seed gives the same text.

| Parameter | Type | Description |
|---|---|---|
| `paragraphs` | `int` (default `3`) | Number of paragraphs. |
| `sentences` | `int` (default `5`) | Sentences per paragraph. |
| `seed` | `int` (default `0`) | Seed for deterministic output. |

Returns `str`. Example: `text.lorem_paragraphs(2, 4)`

## text.counter_string

`text.counter_string(length: int, marker: str = "*") -> str`

Builds a testing counter string of exactly the given length in which each marker is preceded by its own 1-based position, e.g. length 10 gives '*3*5*7*10*' (the marker's position is the number before it). Maximum 1000000.

| Parameter | Type | Description |
|---|---|---|
| `length` | `int` | Total length in characters. |
| `marker` | `str` (default `"*"`) | Single marker character. |

Returns `str`. Example: `text.counter_string(10)`


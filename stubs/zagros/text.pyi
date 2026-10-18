"""Text module (generated from the tools registry; do not edit by hand).

Change, measure and analyse strings: case styles, palindromes, lengths, normalisation, similarity, Unicode.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
"""

from typing import Dict, List

def case_all(s: str) -> Dict[str, str]:
    """Returns every case style of the text as a dict: lower, upper, title, sentence, swapcase, alternating, camel, pascal, snake, kebab, constant, dot, path, train, slug.

        s: Input text.

    Example: text.case_all("hello world")
    Needs: none (pure function).
    """

def stats(s: str) -> Dict[str, str]:
    """Returns a dict with chars, bytes, graphemes, words, lines, sentences, unique_chars, avg_word_length, longest_word, reading_time_seconds and reading_time (at 200 words per minute).

        s: Input text.

    Example: text.stats("Hello world. Bye.")
    Needs: none (pure function).
    """

def lower(s: str) -> str:
    """Converts every letter to lower case using Unicode rules (Σ -> σ). Other characters are unchanged.

        s: Input text.

    Example: text.lower("HELLO World")
    Needs: none (pure function).
    """

def palindrome_report(s: str, ignore_case: bool = True, ignore_non_alnum: bool = True) -> Dict[str, str]:
    """Returns a dict with is_palindrome, cleaned (the compared text), reversed and longest_substring (longest palindromic substring of the cleaned text).

        s: Input text.
        ignore_case: Ignore letter case.
        ignore_non_alnum: Ignore spaces, punctuation and symbols.

    Example: text.palindrome_report("Never odd or even")
    Needs: none (pure function).
    """

def upper(s: str) -> str:
    """Converts every letter to upper case using Unicode simple case mapping. Other characters are unchanged.

        s: Input text.

    Example: text.upper("hello World")
    Needs: none (pure function).
    """

def title(s: str) -> str:
    """Capitalises the first letter of every word and lower-cases the rest. Apostrophes inside a word do not start a new word (don't -> Don't).

        s: Input text.

    Example: text.title("the quick brown fox")
    Needs: none (pure function).
    """

def camel(s: str) -> str:
    """Converts to camelCase. Words are split on separators, case changes and acronyms (HTMLParser -> htmlParser); apostrophes are dropped.

        s: Input text.

    Example: text.camel("hello world foo")
    Needs: none (pure function).
    """

def snake(s: str) -> str:
    """Converts to snake_case: lower-case words joined with underscores. Splits on separators, case changes and acronyms.

        s: Input text.

    Example: text.snake("Hello World fooBar")
    Needs: none (pure function).
    """

def kebab(s: str) -> str:
    """Converts to kebab-case: lower-case words joined with hyphens. Splits on separators, case changes and acronyms.

        s: Input text.

    Example: text.kebab("Hello World fooBar")
    Needs: none (pure function).
    """

def reverse(s: str) -> str:
    """Reverses the text by grapheme cluster, so combining marks, skin-tone modifiers, ZWJ emoji sequences and flags stay intact.

        s: Input text.

    Example: text.reverse("hello")
    Needs: none (pure function).
    """

def slug(s: str, separator: str = "-", ascii: bool = True) -> str:
    """Makes a URL slug: lower case, accents folded to ASCII (é -> e, ß -> ss) when ascii is true, every run of other characters replaced by one separator, no leading or trailing separator.

        s: Input text.
        separator: Separator placed between words.
        ascii: Fold accents and drop non-ASCII characters.

    Example: text.slug("Héllo, Wörld! 2024")
    Needs: none (pure function).
    """

def sort_words(s: str, desc: bool = False, ignore_case: bool = True) -> str:
    """Sorts whitespace-separated words alphabetically and joins them with single spaces. Use lines.sort_lines for line-based sorting.

        s: Input text.
        desc: Descending order.
        ignore_case: Compare case-insensitively.

    Example: text.sort_words("pear apple fig")
    Needs: none (pure function).
    """

def trim(s: str, mode: str = "both", cutset: str = "") -> str:
    """Removes characters from the ends. mode is both, left or right; cutset lists the characters to remove (empty means any whitespace).

        s: Input text.
        mode: Which ends to trim. One of: both, left, right.
        cutset: Characters to remove; empty means whitespace.

    Example: text.trim("  hi  ")
    Needs: none (pure function).
    """

def word_count(s: str) -> int:
    """Counts words: runs of letters, digits and combining marks, which may contain apostrophes or hyphens between them (don't and well-known are one word each). Empty or punctuation-only text gives 0.

        s: Input text.

    Example: text.word_count("Hello, big world!")
    Needs: none (pure function).
    """

def collapse_whitespace(s: str) -> str:
    """Replaces every run of whitespace (spaces, tabs, newlines, Unicode spaces) with one space and trims both ends.

        s: Input text.

    Example: text.collapse_whitespace("a   b \n c")
    Needs: none (pure function).
    """

def sentence_case(s: str) -> str:
    """Lower-cases the text, then capitalises the first letter after the start and after each . ! ? or line break.

        s: Input text.

    Example: text.sentence_case("HELLO there. HOW ARE you?")
    Needs: none (pure function).
    """

def swapcase(s: str) -> str:
    """Inverts the case of every letter: upper becomes lower and lower becomes upper.

        s: Input text.

    Example: text.swapcase("Hello World")
    Needs: none (pure function).
    """

def alternating_case(s: str) -> str:
    """Alternates lower and upper case over the letters, starting with lower (the 'spongebob' meme case). Non-letters do not advance the alternation.

        s: Input text.

    Example: text.alternating_case("hello world")
    Needs: none (pure function).
    """

def pascal(s: str) -> str:
    """Converts to PascalCase (UpperCamelCase). Words are split on separators, case changes and acronyms.

        s: Input text.

    Example: text.pascal("hello world foo")
    Needs: none (pure function).
    """

def constant(s: str) -> str:
    """Converts to SCREAMING_SNAKE_CASE: upper-case words joined with underscores.

        s: Input text.

    Example: text.constant("hello world")
    Needs: none (pure function).
    """

def dot_case(s: str) -> str:
    """Converts to dot.case: lower-case words joined with dots.

        s: Input text.

    Example: text.dot_case("Hello World")
    Needs: none (pure function).
    """

def path_case(s: str) -> str:
    """Converts to path/case: lower-case words joined with forward slashes.

        s: Input text.

    Example: text.path_case("Hello World")
    Needs: none (pure function).
    """

def train_case(s: str) -> str:
    """Converts to Train-Case: capitalised words joined with hyphens (HTTP header style).

        s: Input text.

    Example: text.train_case("hello world")
    Needs: none (pure function).
    """

def capitalize_words(s: str) -> str:
    """Upper-cases the first character of every whitespace-separated word and leaves the other characters untouched.

        s: Input text.

    Example: text.capitalize_words("hello wORLD")
    Needs: none (pure function).
    """

def uncapitalize(s: str) -> str:
    """Lower-cases only the first character of the text (Hello -> hello); empty input returns empty.

        s: Input text.

    Example: text.uncapitalize("Hello World")
    Needs: none (pure function).
    """

def length(s: str) -> int:
    """Number of characters (Unicode code points / runes) in the text. Emoji with modifiers or combining marks count each code point; see grapheme_length.

        s: Input text.

    Example: text.length("héllo")
    Needs: none (pure function).
    """

def byte_length(s: str) -> int:
    """Number of bytes of the text encoded as UTF-8.

        s: Input text.

    Example: text.byte_length("héllo")
    Needs: none (pure function).
    """

def grapheme_length(s: str) -> int:
    """Number of user-perceived characters. Approximation of extended grapheme clusters: a base rune plus following combining marks, variation selectors, skin-tone modifiers, tag characters and ZWJ-joined runes forms one grapheme; regional-indicator pairs (flags) and CRLF are one grapheme.

        s: Input text.

    Example: text.grapheme_length("é👨‍👩‍👧")
    Needs: none (pure function).
    """

def line_count(s: str) -> int:
    """Number of lines. CRLF, LF and lone CR all separate lines; a single trailing line break does not add an empty line; empty text has 0 lines.

        s: Input text.

    Example: text.line_count("a\nb\nc")
    Needs: none (pure function).
    """

def sentence_count(s: str) -> int:
    """Counts sentences ended by . ! ? or ... followed by whitespace or the end; trailing text without a terminator counts as one sentence. Decimals like 3.14 do not split.

        s: Input text.

    Example: text.sentence_count("Hi. How are you? Fine")
    Needs: none (pure function).
    """

def char_count(s: str, ch: str, ignore_case: bool = False) -> int:
    """Counts non-overlapping occurrences of a character or substring. An empty needle is an error.

        s: Text to search.
        ch: Character or substring to count.
        ignore_case: Compare case-insensitively.

    Example: text.char_count("banana", "a")
    Needs: none (pure function).
    """

def reverse_words(s: str) -> str:
    """Reverses the order of whitespace-separated words and joins them with single spaces.

        s: Input text.

    Example: text.reverse_words("one two three")
    Needs: none (pure function).
    """

def sort_chars(s: str, desc: bool = False, remove_space: bool = False) -> str:
    """Sorts the characters by code point (descending when desc is true). Whitespace is kept unless remove_space is true.

        s: Input text.
        desc: Descending order.
        remove_space: Drop whitespace first.

    Example: text.sort_chars("hello")
    Needs: none (pure function).
    """

def remove_whitespace(s: str) -> str:
    """Deletes every whitespace character (spaces, tabs, newlines, Unicode spaces).

        s: Input text.

    Example: text.remove_whitespace("a b\tc")
    Needs: none (pure function).
    """

def remove_punctuation(s: str, symbols: bool = False) -> str:
    """Deletes Unicode punctuation characters; with symbols=true also deletes symbols such as $ + < = ~.

        s: Input text.
        symbols: Also remove symbol characters.

    Example: text.remove_punctuation("Hi, there!")
    Needs: none (pure function).
    """

def is_palindrome(s: str, ignore_case: bool = True, ignore_non_alnum: bool = True) -> bool:
    """True when the text reads the same forwards and backwards. By default case and every non-letter/non-digit are ignored; empty text counts as a palindrome.

        s: Input text.
        ignore_case: Ignore letter case.
        ignore_non_alnum: Ignore spaces, punctuation and symbols.

    Example: text.is_palindrome("A man, a plan, a canal: Panama")
    Needs: none (pure function).
    """

def remove_digits(s: str) -> str:
    """Deletes every Unicode decimal digit.

        s: Input text.

    Example: text.remove_digits("a1b2")
    Needs: none (pure function).
    """

def longest_palindromic_substring(s: str) -> str:
    """Returns the longest substring (case-sensitive, rune-based, first one on ties) that reads the same in both directions; empty text returns empty.

        s: Input text.

    Example: text.longest_palindromic_substring("babad")
    Needs: none (pure function).
    """

def only_digits(s: str) -> str:
    """Keeps only Unicode decimal digits and deletes everything else.

        s: Input text.

    Example: text.only_digits("tel: 555-1234")
    Needs: none (pure function).
    """

def is_anagram(a: str, b: str, ignore_case: bool = True) -> bool:
    """True when both texts use the same letters and digits the same number of times. Spaces and punctuation are ignored; case is ignored by default.

        a: First text.
        b: Second text.
        ignore_case: Ignore letter case.

    Example: text.is_anagram("Listen", "Silent")
    Needs: none (pure function).
    """

def only_letters(s: str) -> str:
    """Keeps only letters (and their combining marks) and deletes everything else.

        s: Input text.

    Example: text.only_letters("a1 b2!")
    Needs: none (pure function).
    """

def anagram_key(s: str) -> str:
    """Returns the letters and digits of the text lower-cased and sorted; two texts are anagrams when their keys match.

        s: Input text.

    Example: text.anagram_key("Listen")
    Needs: none (pure function).
    """

def only_alnum(s: str) -> str:
    """Keeps only letters, combining marks and digits.

        s: Input text.

    Example: text.only_alnum("a-1 b_2!")
    Needs: none (pure function).
    """

def is_pangram(s: str, alphabet: str = "abcdefghijklmnopqrstuvwxyz") -> bool:
    """True when the text contains every letter of the alphabet (default a-z, case-insensitive). Pass another alphabet to test other languages.

        s: Input text.
        alphabet: Letters that must all appear.

    Example: text.is_pangram("The quick brown fox jumps over the lazy dog")
    Needs: none (pure function).
    """

def remove_non_ascii(s: str) -> str:
    """Deletes every character with a code point of 128 or more.

        s: Input text.

    Example: text.remove_non_ascii("héllo ✓")
    Needs: none (pure function).
    """

def strip_accents(s: str) -> str:
    """Removes diacritical marks (é -> e, ñ -> n) via NFD decomposition and dropping non-spacing marks. Letters without decomposition (ø, ß, ł) stay unchanged; use slug for transliteration.

        s: Input text.

    Example: text.strip_accents("Crème brûlée")
    Needs: none (pure function).
    """

def normalize(s: str, form: str = "NFC") -> str:
    """Applies Unicode normalization form NFC, NFD, NFKC or NFKD (case-insensitive name).

        s: Input text.
        form: Normalization form. One of: NFC, NFD, NFKC, NFKD.

    Example: text.normalize("é", "NFC")
    Needs: none (pure function).
    """

def repeat(s: str, n: int = 2, sep: str = "") -> str:
    """Repeats the text n times with an optional separator between copies. n of 0 gives an empty string; negative n or a result over 4 MiB is an error.

        s: Text to repeat.
        n: Number of copies.
        sep: Separator between copies.

    Example: text.repeat("ab", 3, "-")
    Needs: none (pure function).
    """

def center(s: str, width: int, fill: str = " ") -> str:
    """Centers the text in width characters using the fill text; an odd leftover puts the extra fill on the right.

        s: Input text.
        width: Target width in characters.
        fill: Fill text, cycled as needed.

    Example: text.center("hi", 6, "*")
    Needs: none (pure function).
    """

def pad_left(s: str, width: int, fill: str = " ") -> str:
    """Pads the start with the fill text up to width characters (right-aligns). Text already that wide is unchanged.

        s: Input text.
        width: Target width in characters.
        fill: Fill text, cycled as needed.

    Example: text.pad_left("7", 3, "0")
    Needs: none (pure function).
    """

def pad_right(s: str, width: int, fill: str = " ") -> str:
    """Pads the end with the fill text up to width characters (left-aligns). Text already that wide is unchanged.

        s: Input text.
        width: Target width in characters.
        fill: Fill text, cycled as needed.

    Example: text.pad_right("7", 3, ".")
    Needs: none (pure function).
    """

def truncate(s: str, max: int = 20, ellipsis: str = "…") -> str:
    """Shortens the text to at most max characters (runes), ending with the ellipsis when cut. If max is smaller than the ellipsis, only the first max characters are returned.

        s: Input text.
        max: Maximum length in characters, including the ellipsis.
        ellipsis: Marker appended when cut.

    Example: text.truncate("Hello world", 8)
    Needs: none (pure function).
    """

def is_alpha(s: str) -> bool:
    """True when the text is non-empty and every character is a letter (combining marks allowed).

        s: Input text.

    Example: text.is_alpha("héllo")
    Needs: none (pure function).
    """

def is_ascii(s: str) -> bool:
    """True when every character is ASCII (code point below 128). Empty text is ASCII.

        s: Input text.

    Example: text.is_ascii("hello")
    Needs: none (pure function).
    """

def is_blank(s: str) -> bool:
    """True when the text is empty or contains only whitespace (including Unicode spaces).

        s: Input text.

    Example: text.is_blank("  \t")
    Needs: none (pure function).
    """

def is_numeric(s: str) -> bool:
    """True when the text is non-empty and every character is a Unicode decimal digit (Nd). Signs, dots and spaces make it false.

        s: Input text.

    Example: text.is_numeric("12345")
    Needs: none (pure function).
    """

def truncate_words(s: str, max_words: int = 10, ellipsis: str = "…") -> str:
    """Keeps the first max_words whitespace-separated words and appends the ellipsis if any were dropped; shorter text is returned unchanged.

        s: Input text.
        max_words: Words to keep.
        ellipsis: Marker appended when cut.

    Example: text.truncate_words("one two three four", 2)
    Needs: none (pure function).
    """

def consonant_count(s: str) -> int:
    """Counts letters that are not vowels (y counts as a consonant); digits and punctuation are ignored. Non-Latin letters count as consonants.

        s: Input text.

    Example: text.consonant_count("Education")
    Needs: none (pure function).
    """

def vowel_count(s: str) -> int:
    """Counts the vowels a e i o u, including their accented forms (é, ü ...), case-insensitive.

        s: Input text.

    Example: text.vowel_count("Education")
    Needs: none (pure function).
    """

def wrap(s: str, width: int = 72, break_long: bool = False) -> str:
    """Greedy word wrap to width characters per line. Existing line breaks start new paragraphs, runs of spaces collapse, and words longer than width stay whole unless break_long is true.

        s: Input text.
        width: Maximum line width in characters.
        break_long: Split words longer than width.

    Example: text.wrap("the quick brown fox", 10)
    Needs: none (pure function).
    """

def char_frequency(s: str, include_space: bool = False) -> Dict[str, str]:
    """Returns a dict mapping each character to how many times it occurs (values are counts as text). Grapheme-level frequency; whitespace is skipped unless include_space is true. Use char_frequency_sorted for ordered output.

        s: Input text.
        include_space: Count whitespace too.

    Example: text.char_frequency("hello")
    Needs: none (pure function).
    """

def indent(s: str, prefix: str = "    ") -> str:
    """Prefixes every non-empty line with the given text (four spaces by default). Blank lines are left empty.

        s: Input text.
        prefix: Text added to each non-empty line.

    Example: text.indent("a\nb", "  ")
    Needs: none (pure function).
    """

def char_frequency_sorted(s: str, include_space: bool = False) -> List[str]:
    """Returns lines like 'l: 2' ordered by count descending, then by character ascending. Grapheme-level frequency; whitespace is skipped unless include_space is true.

        s: Input text.
        include_space: Count whitespace too.

    Example: text.char_frequency_sorted("hello")
    Needs: none (pure function).
    """

def unindent(s: str) -> str:
    """Removes the leading whitespace shared by all non-blank lines (dedent). Blank lines are ignored when measuring.

        s: Input text.

    Example: text.unindent("  a\n    b")
    Needs: none (pure function).
    """

def replace(s: str, old: str, new: str = "", count: int = -1) -> str:
    """Replaces occurrences of old with new, at most count times (-1 = all). An empty old returns the text unchanged.

        s: Input text.
        old: Text to find.
        new: Replacement text.
        count: Maximum replacements; -1 for all.

    Example: text.replace("a-b-c", "-", "+")
    Needs: none (pure function).
    """

def unique_chars(s: str) -> str:
    """Returns each distinct character once, in order of first appearance (whitespace skipped).

        s: Input text.

    Example: text.unique_chars("hello")
    Needs: none (pure function).
    """

def duplicate_chars(s: str) -> str:
    """Returns the characters that occur more than once, each listed once in order of first appearance (whitespace skipped).

        s: Input text.

    Example: text.duplicate_chars("hello")
    Needs: none (pure function).
    """

def replace_ci(s: str, old: str, new: str = "", count: int = -1) -> str:
    """Like replace but matches old case-insensitively; the replacement is inserted as given. An empty old returns the text unchanged.

        s: Input text.
        old: Text to find (any case).
        new: Replacement text.
        count: Maximum replacements; -1 for all.

    Example: text.replace_ci("Hello HELLO", "hello", "bye")
    Needs: none (pure function).
    """

def most_common_char(s: str) -> str:
    """Returns the most frequent character (whitespace skipped); ties go to the smallest character; empty text returns empty.

        s: Input text.

    Example: text.most_common_char("hello")
    Needs: none (pure function).
    """

def remove(s: str, sub: str) -> str:
    """Deletes every occurrence of sub. An empty sub returns the text unchanged.

        s: Input text.
        sub: Text to delete.

    Example: text.remove("banana", "an")
    Needs: none (pure function).
    """

def insert(s: str, index: int, text: str) -> str:
    """Inserts text at a character index (negative counts from the end; out-of-range values clamp to the ends).

        s: Input text.
        index: Character index.
        text: Text to insert.

    Example: text.insert("held", 3, "l")
    Needs: none (pure function).
    """

def delete_range(s: str, start: int, end: int) -> str:
    """Deletes characters from start up to but excluding end (character indexes, negative from the end, clamped). start >= end deletes nothing.

        s: Input text.
        start: First character index to delete.
        end: Index after the last deleted character.

    Example: text.delete_range("hello", 1, 3)
    Needs: none (pure function).
    """

def substring(s: str, start: int = 0, length: int = -1) -> str:
    """Returns length characters starting at start (character-based; negative start counts from the end; length -1 means to the end). Out-of-range values clamp.

        s: Input text.
        start: Start index; negative counts from the end.
        length: Number of characters; -1 for the rest.

    Example: text.substring("héllo wörld", 6, 5)
    Needs: none (pure function).
    """

def split_at(s: str, index: int) -> List[str]:
    """Splits the text into two parts at a character index (negative from the end, clamped). Always returns exactly two strings.

        s: Input text.
        index: Character index.

    Example: text.split_at("hello", 2)
    Needs: none (pure function).
    """

def left(s: str, n: int = 1) -> str:
    """Returns the first n characters (all of them if n is larger; n <= 0 gives empty).

        s: Input text.
        n: Number of characters.

    Example: text.left("hello", 2)
    Needs: none (pure function).
    """

def right(s: str, n: int = 1) -> str:
    """Returns the last n characters (all of them if n is larger; n <= 0 gives empty).

        s: Input text.
        n: Number of characters.

    Example: text.right("hello", 2)
    Needs: none (pure function).
    """

def remove_prefix(s: str, prefix: str) -> str:
    """Removes the prefix when the text starts with it; otherwise returns the text unchanged.

        s: Input text.
        prefix: Prefix to remove.

    Example: text.remove_prefix("foobar", "foo")
    Needs: none (pure function).
    """

def remove_suffix(s: str, suffix: str) -> str:
    """Removes the suffix when the text ends with it; otherwise returns the text unchanged.

        s: Input text.
        suffix: Suffix to remove.

    Example: text.remove_suffix("foobar", "bar")
    Needs: none (pure function).
    """

def leetspeak(s: str) -> str:
    """Converts to leetspeak with a fixed map: a4 b8 e3 g6 i1 o0 s5 t7 (case-insensitive); other characters are unchanged.

        s: Input text.

    Example: text.leetspeak("elite hacker")
    Needs: none (pure function).
    """

def mirror(s: str) -> str:
    """Flips the text upside down: reverses it and swaps letters, digits and brackets for look-alike rotated characters (a -> ɐ, b -> q, ? -> ¿). Characters without a mapping stay as they are.

        s: Input text.

    Example: text.mirror("Hello")
    Needs: none (pure function).
    """

def remove_zero_width(s: str) -> str:
    """Deletes invisible characters: zero-width space/non-joiner/joiner, word joiner, BOM, soft hyphen, bidi marks and overrides, Mongolian vowel separator, combining grapheme joiner. Note that ZWNJ/ZWJ are meaningful in some scripts and emoji sequences.

        s: Input text.

    Example: text.remove_zero_width("a​b")
    Needs: none (pure function).
    """

def show_invisibles(s: str, show_spaces: bool = False) -> str:
    """Makes invisible characters visible: \n, \r and \t as escapes; other control, zero-width and special space characters (including no-break space) as <U+XXXX>; regular spaces become · only when show_spaces is true.

        s: Input text.
        show_spaces: Render ordinary spaces as ·.

    Example: text.show_invisibles("a\tb​")
    Needs: none (pure function).
    """

def smart_quotes_to_straight(s: str, punctuation: bool = False) -> str:
    """Replaces curly quotes (‘ ’ “ ” „ ‚ and primes) with ASCII ' and "; with punctuation=true also converts en/em dashes to - and … to ...

        s: Input text.
        punctuation: Also convert dashes and ellipsis.

    Example: text.smart_quotes_to_straight("“hi” it’s")
    Needs: none (pure function).
    """

def straight_to_smart(s: str) -> str:
    """Replaces ASCII quotes with typographic ones: opening after the start, whitespace or an opening bracket, closing otherwise; a lone apostrophe inside a word becomes ’.

        s: Input text.

    Example: text.straight_to_smart("\"hi\" it's")
    Needs: none (pure function).
    """

def remove_emoji(s: str) -> str:
    """Deletes emoji and pictographs (including skin-tone modifiers, flags, variation selectors and the ZWJ glue between emoji). Covers the main emoji blocks, not every symbol.

        s: Input text.

    Example: text.remove_emoji("hi 😀 there 👍🏽")
    Needs: none (pure function).
    """

def to_fullwidth(s: str) -> str:
    """Converts ASCII printable characters to their fullwidth forms (A -> Ａ, space -> U+3000).

        s: Input text.

    Example: text.to_fullwidth("Hello 123")
    Needs: none (pure function).
    """

def to_halfwidth(s: str) -> str:
    """Converts fullwidth ASCII forms (U+FF01-FF5E and U+3000) back to normal ASCII. Halfwidth katakana is not converted.

        s: Input text.

    Example: text.to_halfwidth("Ｈｅｌｌｏ")
    Needs: none (pure function).
    """

def levenshtein(a: str, b: str) -> int:
    """Edit distance between two texts: the minimum number of single-character insertions, deletions and substitutions (rune-based). Fails for very large inputs (over 50 million table cells).

        a: First text.
        b: Second text.

    Example: text.levenshtein("kitten", "sitting")
    Needs: none (pure function).
    """

def damerau_levenshtein(a: str, b: str) -> int:
    """Edit distance that also counts a swap of two adjacent characters as one edit (optimal string alignment variant, rune-based). ca -> ac is 1.

        a: First text.
        b: Second text.

    Example: text.damerau_levenshtein("ca", "ac")
    Needs: none (pure function).
    """

def hamming(a: str, b: str) -> int:
    """Number of positions at which two texts of equal length (in characters) differ. Different lengths are an error.

        a: First text.
        b: Second text.

    Example: text.hamming("karolin", "kathrin")
    Needs: none (pure function).
    """

def jaro_winkler(a: str, b: str) -> float:
    """Jaro-Winkler similarity from 0 (nothing in common) to 1 (identical), favouring a shared prefix of up to 4 characters; two empty texts give 1.

        a: First text.
        b: Second text.

    Example: text.jaro_winkler("martha", "marhta")
    Needs: none (pure function).
    """

def similarity(a: str, b: str) -> float:
    """Similarity from 0 to 1 computed as 1 - levenshtein / longer length; two empty texts give 1.

        a: First text.
        b: Second text.

    Example: text.similarity("kitten", "sitting")
    Needs: none (pure function).
    """

def common_prefix(a: str, b: str) -> str:
    """Longest prefix shared by both texts (character-based); empty when they differ from the first character.

        a: First text.
        b: Second text.

    Example: text.common_prefix("interview", "internet")
    Needs: none (pure function).
    """

def common_suffix(a: str, b: str) -> str:
    """Longest suffix shared by both texts (character-based); empty when the last characters differ.

        a: First text.
        b: Second text.

    Example: text.common_suffix("testing", "running")
    Needs: none (pure function).
    """

def longest_common_substring(a: str, b: str) -> str:
    """Longest contiguous piece of text found in both inputs (case-sensitive, first one in the first text on ties). Fails for very large inputs.

        a: First text.
        b: Second text.

    Example: text.longest_common_substring("abcdef", "zcdemn")
    Needs: none (pure function).
    """

def diff_summary(a: str, b: str) -> Dict[str, str]:
    """Character-level summary of how a becomes b using the longest common subsequence: dict with inserted, deleted, unchanged, distance (inserted + deleted) and similarity (0..1). Fails for very large inputs.

        a: First text.
        b: Second text.

    Example: text.diff_summary("kitten", "sitting")
    Needs: none (pure function).
    """

def compare_ci(a: str, b: str) -> int:
    """Compares two texts ignoring case: returns -1 if a sorts before b, 0 if equal, 1 if after.

        a: First text.
        b: Second text.

    Example: text.compare_ci("Apple", "apple")
    Needs: none (pure function).
    """

def equals_ci(a: str, b: str) -> bool:
    """True when both texts are equal ignoring case (Unicode simple folding).

        a: First text.
        b: Second text.

    Example: text.equals_ci("Hello", "hELLO")
    Needs: none (pure function).
    """

def codepoints(s: str) -> List[str]:
    """Lists the Unicode code point of every character as U+XXXX (at least four hex digits, upper case). Invalid UTF-8 bytes appear as U+FFFD.

        s: Input text.

    Example: text.codepoints("aé😀")
    Needs: none (pure function).
    """

def from_codepoints(points: List[str]) -> str:
    """Builds text from code points written as U+0041, 0x41, \u0041, &#65; (decimal) or plain hex. Invalid values or surrogates are an error.

        points: Code point strings.

    Example: text.from_codepoints(["U+0048", "0x69"])
    Needs: none (pure function).
    """

def to_codepoint(ch: str) -> int:
    """Returns the decimal Unicode code point of the first character; empty text is an error.

        ch: A character; only the first one is used.

    Example: text.to_codepoint("A")
    Needs: none (pure function).
    """

def from_codepoint(n: int) -> str:
    """Returns the character for a decimal code point (0 to 0x10FFFF, no surrogates).

        n: Decimal code point.

    Example: text.from_codepoint(128512)
    Needs: none (pure function).
    """

def char_info(ch: str) -> Dict[str, str]:
    """Describes the first character as a dict: char, codepoint (U+XXXX), decimal, utf8 (hex bytes), utf8_length, utf16 (hex code units), category (e.g. Lu), category_name, block, upper, lower and printable. Character names are not included.

        ch: A character; only the first one is used.

    Example: text.char_info("é")
    Needs: none (pure function).
    """

def unicode_blocks(s: str) -> List[str]:
    """Lists the distinct Unicode blocks (Basic Latin, Cyrillic, Emoticons ...) used by the text, in order of first appearance. Characters in blocks not in the built-in table are reported as Other.

        s: Input text.

    Example: text.unicode_blocks("aб😀")
    Needs: none (pure function).
    """

def lorem(words: int = 50, seed: int = 0) -> str:
    """Generates lorem ipsum text of exactly the requested number of words (1 to 100000), as sentences. The same seed always gives the same text; seed 0 starts with 'Lorem ipsum dolor sit amet, consectetur adipiscing elit'.

        words: Number of words.
        seed: Seed for deterministic output; 0 uses the classic opening.

    Example: text.lorem(20)
    Needs: none (pure function).
    """

def lorem_paragraphs(paragraphs: int = 3, sentences: int = 5, seed: int = 0) -> str:
    """Generates lorem ipsum paragraphs (1 to 1000) of the given sentence count (1 to 100) separated by blank lines. The same seed gives the same text.

        paragraphs: Number of paragraphs.
        sentences: Sentences per paragraph.
        seed: Seed for deterministic output.

    Example: text.lorem_paragraphs(2, 4)
    Needs: none (pure function).
    """

def counter_string(length: int, marker: str = "*") -> str:
    """Builds a testing counter string of exactly the given length in which each marker is preceded by its own 1-based position, e.g. length 10 gives '*3*5*7*10*' (the marker's position is the number before it). Maximum 1000000.

        length: Total length in characters.
        marker: Single marker character.

    Example: text.counter_string(10)
    Needs: none (pure function).
    """

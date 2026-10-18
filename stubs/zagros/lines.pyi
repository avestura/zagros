"""Lines: Work on multi-line text: sort, unique, filter, join, columns, CSV and list conversions.

Generated from the tools registry; do not edit by hand.
Type description, not runnable Starlark: do not `import` anything. The name `lines` is predeclared in main.star; call `lines.<name>(...)` directly.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
Keywords: sort lines, remove duplicate lines, dedupe, join lines, split text into lines, filter lines, csv row, list, column, diff lines, line endings, markdown table
"""

from typing import Any, Dict, List

def sort_lines(s: str, order: str = "asc", mode: str = "text", reverse: bool = False) -> str:
    """Sorts lines. mode is text (code point order), natural (file2 before file10), numeric (by leading number; lines without one go last), length (in characters) or case_insensitive; order is asc or desc; reverse flips the result. The sort is stable. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        order: asc or desc. One of: asc, desc.
        mode: How lines are compared. One of: text, natural, numeric, length, case_insensitive.
        reverse: Reverse the sorted result afterwards.

    Example: lines.sort_lines("b\na\nc")
    Task: sort lines
    Keywords: alphabetize, order lines, sort a list, sort, order, alphabetical, natural, numeric, length, lines, arrange, a-z
    Needs: none (pure function).
    """

def unique_lines(s: str, case_insensitive: bool = False) -> str:
    """Removes duplicate lines, keeping the first occurrence and the original order. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        case_insensitive: Treat lines differing only in case as equal.

    Example: lines.unique_lines("a\nb\na")
    Task: remove duplicate lines
    Keywords: dedupe, deduplicate, distinct lines, uniq, unique, duplicates, remove, distinct, lines
    Needs: none (pure function).
    """

def reverse_lines(s: str) -> str:
    """Reverses the order of the lines (last line first). Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.

    Example: lines.reverse_lines("a\nb\nc")
    Keywords: reverse, order, flip, lines, tac, backwards
    Needs: none (pure function).
    """

def count_lines(s: str, skip_empty: bool = False) -> int:
    """Counts lines. A single trailing line break does not add a line; empty text has 0 lines. skip_empty ignores blank lines.

        s: Input text, one item per line.
        skip_empty: Do not count blank lines.

    Example: lines.count_lines("a\n\nb")
    Task: count lines
    Keywords: number of lines, wc -l, count, lines, number, wc, -l, how, many
    Needs: none (pure function).
    """

def number_lines(s: str, start: int = 1, sep: str = ". ", align: bool = False) -> str:
    """Prefixes each line with its number followed by sep (default '. '). With align true the numbers are right-aligned to the same width. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        start: First number.
        sep: Text between number and line.
        align: Right-align numbers to equal width.

    Example: lines.number_lines("a\nb", 1, ". ")
    Keywords: add line numbers, prefix numbers, number, lines, index, enumerate, line, numbers, nl, cat, -n
    Needs: none (pure function).
    """

def remove_empty_lines(s: str, whitespace_only: bool = True) -> str:
    """Deletes blank lines. By default lines containing only whitespace count as blank; set whitespace_only to false to delete only truly empty lines. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        whitespace_only: Also remove lines containing only whitespace.

    Example: lines.remove_empty_lines("a\n\nb")
    Task: remove blank lines
    Keywords: delete empty lines, strip blank lines, remove, empty, blank, lines, delete, clean, squeeze
    Needs: none (pure function).
    """

def trim_lines(s: str, mode: str = "both") -> str:
    """Trims whitespace from each line (mode both, left or right). Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        mode: Which ends to trim. One of: both, left, right.

    Example: lines.trim_lines("  a \n b")
    Keywords: trim, strip, whitespace, lines, each, leading, trailing
    Needs: none (pure function).
    """

def join_lines(s: str, sep: str = " ", skip_empty: bool = False) -> str:
    """Joins all lines into one line, replacing line breaks with sep (default a space). skip_empty drops blank lines first. The result has no trailing line break.

        s: Input text, one item per line.
        sep: Text placed between lines.
        skip_empty: Drop blank lines.

    Example: lines.join_lines("a\nb\nc", ", ")
    Task: join lines into one
    Keywords: concatenate lines, merge lines with a separator, comma separated, join, lines, merge, combine, concatenate, implode, comma, separate, single, line
    Needs: none (pure function).
    """

def split_to_lines(s: str, sep: str = ",", trim: bool = False) -> str:
    """Replaces every occurrence of sep with a line break (default comma); trim removes surrounding whitespace from each piece. An empty sep is an error.

        s: Input text, one item per line.
        sep: Separator to split on.
        trim: Trim each piece.

    Example: lines.split_to_lines("a, b, c", ",", True)
    Task: split text into lines
    Keywords: break by separator, one per line, split, lines, separator, explode, comma, break, newline, delimiter
    Needs: none (pure function).
    """

def prefix_lines(s: str, prefix: str = "> ", skip_empty: bool = False) -> str:
    """Adds text to the start of every line. skip_empty leaves blank lines untouched. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        prefix: Text to add.
        skip_empty: Skip blank lines.

    Example: lines.prefix_lines("a\nb", "> ")
    Keywords: prefix, add, start, each, line, prepend, quote
    Needs: none (pure function).
    """

def suffix_lines(s: str, suffix: str = ";", skip_empty: bool = False) -> str:
    """Adds text to the end of every line. skip_empty leaves blank lines untouched. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        suffix: Text to add.
        skip_empty: Skip blank lines.

    Example: lines.suffix_lines("a\nb", ";")
    Keywords: suffix, add, end, each, line, append, semicolon
    Needs: none (pure function).
    """

def surround_lines(s: str, prefix: str = "\"", suffix: str = "\"", skip_empty: bool = False) -> str:
    """Wraps every line between a prefix and a suffix (default double quotes). skip_empty leaves blank lines untouched. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        prefix: Text before each line.
        suffix: Text after each line.
        skip_empty: Skip blank lines.

    Example: lines.surround_lines("a\nb", "'", "'")
    Keywords: surround, wrap, quote, each, line, double, quotes, single, brackets
    Needs: none (pure function).
    """

def indent_lines(s: str, width: int = 4, use_tabs: bool = False) -> str:
    """Indents non-empty lines by width spaces (or tabs when use_tabs is true). Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        width: Spaces (or tabs) to add.
        use_tabs: Indent with tab characters.

    Example: lines.indent_lines("a\nb", 2)
    Keywords: indent, tab, spaces, lines, shift, right
    Needs: none (pure function).
    """

def unindent_lines(s: str, width: int = 0) -> str:
    """Removes up to width leading spaces (or one tab per step) from each line; with width 0 removes the indentation common to all non-blank lines. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        width: Columns to remove; 0 removes the common indentation.

    Example: lines.unindent_lines("    a\n    b", 4)
    Keywords: unindent, dedent, outdent, remove, indentation, lines
    Needs: none (pure function).
    """

def filter_contains(s: str, substr: str, invert: bool = False, ignore_case: bool = False) -> str:
    """Keeps the lines that contain substr (or, with invert, the lines that do not). Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        substr: Text to look for.
        invert: Keep the lines that do NOT match.
        ignore_case: Ignore letter case.

    Example: lines.filter_contains("apple\nbanana", "an")
    Task: keep lines containing text
    Keywords: grep, filter lines, search lines, filter, contains, match, search, keep, lines
    Needs: none (pure function).
    """

def filter_prefix(s: str, prefix: str, invert: bool = False, ignore_case: bool = False) -> str:
    """Keeps the lines that start with prefix (or, with invert, the lines that do not). Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        prefix: Text to look for.
        invert: Keep the lines that do NOT match.
        ignore_case: Ignore letter case.

    Example: lines.filter_prefix("#a\nb", "#")
    Keywords: filter, starts, with, prefix, begin, keep, lines
    Needs: none (pure function).
    """

def filter_suffix(s: str, suffix: str, invert: bool = False, ignore_case: bool = False) -> str:
    """Keeps the lines that end with suffix (or, with invert, the lines that do not). Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        suffix: Text to look for.
        invert: Keep the lines that do NOT match.
        ignore_case: Ignore letter case.

    Example: lines.filter_suffix("a.go\nb.py", ".go")
    Keywords: filter, ends, with, suffix, extension, keep, lines
    Needs: none (pure function).
    """

def head(s: str, n: int = 10) -> str:
    """Keeps the first n lines (all of them if fewer). n of 0 gives empty text. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        n: Number of lines.

    Example: lines.head("a\nb\nc", 2)
    Task: take the first lines
    Keywords: first n lines, top lines, head, first, lines, top, take, n
    Needs: none (pure function).
    """

def tail(s: str, n: int = 10) -> str:
    """Keeps the last n lines (all of them if fewer). n of 0 gives empty text. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        n: Number of lines.

    Example: lines.tail("a\nb\nc", 2)
    Task: take the last lines
    Keywords: last n lines, bottom lines, tail, last, lines, bottom, end, take, n
    Needs: none (pure function).
    """

def nth_line(s: str, n: int = 1) -> str:
    """Returns line n, counting from 1; a negative n counts from the end (-1 is the last line). Out of range is an error.

        s: Input text, one item per line.
        n: 1-based line number, negative from the end.

    Example: lines.nth_line("a\nb\nc", 2)
    Keywords: nth, line, get, number, pick, index, row
    Needs: none (pure function).
    """

def line_range(s: str, from: int = 1, to: int = 0) -> str:
    """Keeps lines from..to, 1-based and inclusive; to of 0 means up to the last line; negative numbers count from the end. Out-of-range bounds are clamped. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        from: First line (1-based).
        to: Last line (inclusive); 0 for the end.

    Example: lines.line_range("a\nb\nc\nd", 2, 3)
    Keywords: line, range, slice, lines, from, to, between, sed
    Needs: none (pure function).
    """

def duplicate_lines(s: str, case_insensitive: bool = False) -> str:
    """Keeps one copy of each line that appears more than once, in order of first appearance. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        case_insensitive: Treat lines differing only in case as equal.

    Example: lines.duplicate_lines("a\nb\na")
    Keywords: duplicates, repeated, find, lines, uniq, -d
    Needs: none (pure function).
    """

def sort_unique(s: str, order: str = "asc", mode: str = "text", reverse: bool = False, case_insensitive: bool = False) -> str:
    """Sorts the lines (same options as sort_lines) and removes duplicates, like sort -u. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        order: asc or desc. One of: asc, desc.
        mode: How lines are compared. One of: text, natural, numeric, length, case_insensitive.
        reverse: Reverse the sorted result afterwards.
        case_insensitive: Dedupe ignoring case.

    Example: lines.sort_unique("b\na\nb")
    Keywords: sort -u, sorted distinct, sort, unique, dedupe, -u, distinct
    Needs: none (pure function).
    """

def shuffle_lines(s: str, seed: int = 0) -> str:
    """Shuffles the lines pseudo-randomly. The function is pure: the same seed always gives the same order (seed 0 is a valid fixed seed); change the seed for a different shuffle. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        seed: Seed value.

    Example: lines.shuffle_lines("a\nb\nc\nd", 42)
    Keywords: randomize order, scramble lines, shuffle, random, order, lines, permute
    Needs: none (pure function).
    """

def skip(s: str, n: int = 1) -> str:
    """Drops the first n lines and keeps the rest. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        n: Lines to skip.

    Example: lines.skip("a\nb\nc", 1)
    Keywords: skip, drop, first, lines, offset, tail, from
    Needs: none (pure function).
    """

def take_every(s: str, n: int = 2, offset: int = 0) -> str:
    """Keeps every nth line starting with line offset+1 (offset defaults to 0, so n=2 keeps lines 1, 3, 5...). n must be at least 1. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        n: Step.
        offset: Lines to skip before the first kept one.

    Example: lines.take_every("a\nb\nc\nd", 2)
    Keywords: every, nth, line, step, sample, stride, pick, alternate
    Needs: none (pure function).
    """

def dedupe_consecutive(s: str, case_insensitive: bool = False) -> str:
    """Collapses runs of identical adjacent lines into one (like uniq without -d); non-adjacent duplicates stay. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        case_insensitive: Compare ignoring case.

    Example: lines.dedupe_consecutive("a\na\nb\na")
    Keywords: dedupe, consecutive, adjacent, duplicates, uniq, repeated, runs
    Needs: none (pure function).
    """

def min_max_lines(s: str, mode: str = "length") -> Dict[str, str]:
    """Finds the extreme lines by mode: length (shortest and longest, in characters), text (first and last in code point order) or numeric (smallest and largest leading number). Returns a dict with count, min and max (empty strings for empty input; lines without a number are ignored in numeric mode).

        s: Input text, one item per line.
        mode: How lines are compared. One of: length, text, numeric.

    Example: lines.min_max_lines("bb\na\nccc")
    Keywords: min, max, shortest, longest, smallest, largest, lines, extreme
    Needs: none (pure function).
    """

def transpose(s: str, delimiter: str = ",") -> str:
    """Turns rows into columns: each line is split by the delimiter and the table is flipped. Short rows are padded with empty cells. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        delimiter: Column delimiter; empty means runs of whitespace.

    Example: lines.transpose("a,b\nc,d", ",")
    Keywords: swap rows and columns, pivot, transpose, rotate, matrix, rows, columns, flip, table, swap, axes
    Needs: none (pure function).
    """

def columns(s: str, delimiter: str = ",", indexes: List[str] = [], output_delimiter: str = "") -> str:
    """Keeps selected columns of each line. indexes are 1-based: '2', '-1' (last), '2-4' or '3-' (to the end); comma-separated entries also work. Rows lacking a column skip it. Output is joined with output_delimiter (default: the input delimiter, or a space for whitespace splitting). Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        delimiter: Column delimiter; empty means runs of whitespace.
        indexes: Columns to keep.
        output_delimiter: Joiner for the kept columns; empty reuses the delimiter.

    Example: lines.columns("a,b,c", ",", ["1", "3"])
    Keywords: columns, pick, cut, select, fields, awk, extract, column, csv
    Needs: none (pure function).
    """

def align_columns(s: str, delimiter: str = ",", separator: str = "  ", align: str = "left") -> str:
    """Pads the cells of each column to equal width so the lines form a table. Widths count characters; trailing padding is removed. align is left or right. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        delimiter: Column delimiter; empty means runs of whitespace.
        separator: Text between aligned columns.
        align: left or right. One of: left, right.

    Example: lines.align_columns("a,bbb\ncc,d", ",", " | ")
    Keywords: align, columns, table, pad, format, pretty, print, column
    Needs: none (pure function).
    """

def swap_columns(s: str, col_a: int = 1, col_b: int = 2, delimiter: str = ",") -> str:
    """Swaps two 1-based columns (negative counts from the end) on every line. Lines that lack either column are left unchanged. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        col_a: First column.
        col_b: Second column.
        delimiter: Column delimiter; empty means runs of whitespace.

    Example: lines.swap_columns("a,b,c", 1, 3, ",")
    Keywords: swap, columns, exchange, reorder, fields, switch
    Needs: none (pure function).
    """

def sort_by_column(s: str, column: int = 1, delimiter: str = "", mode: str = "text", order: str = "asc") -> str:
    """Sorts lines by one 1-based column (negative from the end); lines lacking the column sort as empty. mode and order as in sort_lines. The sort is stable. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        column: Column to sort by.
        delimiter: Column delimiter; empty means runs of whitespace.
        mode: How keys are compared. One of: text, natural, numeric, length, case_insensitive.
        order: asc or desc. One of: asc, desc.

    Example: lines.sort_by_column("b,2\na,1", 2, ",", "numeric")
    Keywords: sort table by column, sort, by, column, field, key, csv, table, order
    Needs: none (pure function).
    """

def sum_numbers(s: str, strict: bool = False) -> float:
    """Adds up the numbers written one per line (thousands separators like 1,234.5 are accepted). Blank and non-numeric lines are ignored unless strict is true, in which case they are an error. No numbers gives 0.

        s: Input text, one item per line.
        strict: Fail on lines that are not numbers.

    Example: lines.sum_numbers("1\n2.5\n3")
    Keywords: sum, total, add, numbers, lines, calculate
    Needs: none (pure function).
    """

def average_numbers(s: str, strict: bool = False) -> float:
    """Arithmetic mean of the numbers written one per line. Blank and non-numeric lines are ignored unless strict is true; with no numbers it is an error.

        s: Input text, one item per line.
        strict: Fail on lines that are not numbers.

    Example: lines.average_numbers("1\n2\n6")
    Keywords: average, mean, avg, numbers, lines, calculate
    Needs: none (pure function).
    """

def list_to_csv_row(s: str, delimiter: str = ",", skip_empty: bool = False) -> str:
    """Joins the lines into one RFC 4180 CSV row, quoting fields that contain the delimiter, a quote or a line break (quotes are doubled). skip_empty drops blank lines.

        s: Input text, one item per line.
        delimiter: Field delimiter.
        skip_empty: Drop blank lines.

    Example: lines.list_to_csv_row("a\nb,c")
    Keywords: comma separated values, lines to csv, csv, row, list, join, comma, separated, rfc4180, quote
    Needs: none (pure function).
    """

def csv_row_to_lines(s: str, delimiter: str = ",") -> str:
    """Parses CSV text (normally one row) and returns each field on its own line. Handles quoted fields with embedded delimiters, quotes and line breaks; with several rows, all fields are listed in order.

        s: Input text, one item per line.
        delimiter: Field delimiter (one character).

    Example: lines.csv_row_to_lines("a,\"b,c\",d")
    Keywords: csv, row, split, fields, lines, parse, explode
    Needs: none (pure function).
    """

def csv_to_tsv(s: str, delimiter: str = ",") -> str:
    """Converts CSV to tab-separated text. Tabs and line breaks inside fields are replaced by single spaces because TSV has no quoting. delimiter is the CSV delimiter (one character).

        s: Input text, one item per line.
        delimiter: CSV delimiter.

    Example: lines.csv_to_tsv("a,\"b,c\"\n1,2")
    Keywords: csv, tsv, convert, tab, separated, spreadsheet
    Needs: none (pure function).
    """

def tsv_to_csv(s: str, delimiter: str = ",") -> str:
    """Converts tab-separated text to RFC 4180 CSV: fields containing the delimiter, quotes or CR are quoted with doubled quotes. delimiter is the output delimiter (one character).

        s: Input text, one item per line.
        delimiter: CSV delimiter to produce.

    Example: lines.tsv_to_csv("a\tb,c\n1\t2")
    Keywords: tsv, csv, convert, tab, separated, spreadsheet
    Needs: none (pure function).
    """

def markdown_table_from_csv(s: str, delimiter: str = ",", header: bool = True) -> str:
    """Builds a GitHub-style Markdown table from CSV (or any single-character delimiter such as a tab). With header true the first row is the header; otherwise the header is Column 1, Column 2 ... Short rows are padded and | in cells is escaped.

        s: Input text, one item per line.
        delimiter: CSV delimiter.
        header: First row is the header.

    Example: lines.markdown_table_from_csv("name,age\nAnn,30")
    Keywords: csv to markdown table, make a table, markdown, table, csv, convert, github, gfm, md
    Needs: none (pure function).
    """

def markdown_bullets(s: str, marker: str = "-") -> str:
    """Turns each non-empty line into a Markdown bullet by prefixing the marker and a space (marker - by default; * and + also work). Blank lines are kept as they are. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        marker: Bullet marker. One of: -, *, +.

    Example: lines.markdown_bullets("a\nb")
    Keywords: markdown, bullets, list, unordered, dash, asterisk, md
    Needs: none (pure function).
    """

def numbered_list(s: str, start: int = 1, sep: str = ". ") -> str:
    """Turns each non-empty line into a numbered item ('1. text'); blank lines are kept and do not consume a number. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        start: First number.
        sep: Text between number and line.

    Example: lines.numbered_list("a\nb")
    Keywords: numbered, list, ordered, markdown, enumerate, md, 1.
    Needs: none (pure function).
    """

def remove_markdown_bullets(s: str) -> str:
    """Strips list markers from the start of each line: - * + and bullet characters, numbered markers like 1. or 2), and task boxes [ ] / [x]. Indentation is kept; lines without a marker are unchanged. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.

    Example: lines.remove_markdown_bullets("- a\n2. b")
    Keywords: remove, bullets, markdown, list, markers, strip, numbers, unordered
    Needs: none (pure function).
    """

def lines_diff(a: str, b: str, ignore_case: bool = False) -> Dict[str, str]:
    """Compares two texts as sets of lines and returns a dict with only_in_a, only_in_b and common (each a newline-joined list in order of first appearance, without duplicates).

        a: First text, one item per line.
        b: Second text, one item per line.
        ignore_case: Compare lines ignoring case.

    Example: lines.lines_diff("a\nb", "b\nc")
    Task: compare two texts line by line
    Keywords: diff, what changed, difference between lists, compare, lines, difference, set, only, in, common, intersect
    Needs: none (pure function).
    """

def intersect_lines(a: str, b: str, ignore_case: bool = False) -> str:
    """Lines of a that also appear in b (duplicates removed, order of a). Uses the line ending style and trailing newline of a.

        a: First text, one item per line.
        b: Second text, one item per line.
        ignore_case: Compare lines ignoring case.

    Example: lines.intersect_lines("a\nb", "b\nc")
    Keywords: common lines, lines in both lists, intersect, common, both, lines, set, and
    Needs: none (pure function).
    """

def union_lines(a: str, b: str, ignore_case: bool = False) -> str:
    """All distinct lines of a followed by the new lines of b. Uses the line ending style and trailing newline of a.

        a: First text, one item per line.
        b: Second text, one item per line.
        ignore_case: Compare lines ignoring case.

    Example: lines.union_lines("a\nb", "b\nc")
    Keywords: union, merge, combine, lines, set, or, distinct
    Needs: none (pure function).
    """

def subtract_lines(a: str, b: str, ignore_case: bool = False) -> str:
    """Distinct lines of a that do not appear in b. Uses the line ending style and trailing newline of a.

        a: First text, one item per line.
        b: Second text, one item per line.
        ignore_case: Compare lines ignoring case.

    Example: lines.subtract_lines("a\nb", "b\nc")
    Keywords: lines only in first list, set difference, subtract, difference, except, minus, remove, lines, set, not, in
    Needs: none (pure function).
    """

def wrap_each(s: str, width: int = 72) -> str:
    """Splits every line into chunks of exactly width characters (the last chunk may be shorter), ignoring word boundaries. Blank lines are kept. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        width: Maximum line width in characters.

    Example: lines.wrap_each("abcdefgh", 3)
    Keywords: wrap, hard, chunk, split, width, columns, fixed, each, line
    Needs: none (pure function).
    """

def word_wrap_lines(s: str, width: int = 72, break_long: bool = False) -> str:
    """Word-wraps every line separately to width characters, breaking only between words (runs of whitespace collapse to one space); blank lines are kept, over-long words stay whole unless break_long is true. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        width: Maximum line width in characters.
        break_long: Split words longer than width.

    Example: lines.word_wrap_lines("the quick brown fox", 10)
    Keywords: word, wrap, reflow, width, paragraph, fill, lines, soft
    Needs: none (pure function).
    """

def to_json_array(s: str, skip_empty: bool = False, pretty: bool = False) -> str:
    """Converts the lines to a JSON array of strings (valid UTF-8, no HTML escaping). skip_empty drops blank lines; pretty indents one item per line.

        s: Input text, one item per line.
        skip_empty: Drop blank lines.
        pretty: Indent the output.

    Example: lines.to_json_array("a\nb")
    Keywords: lines to json, list to json array, json, array, lines, convert, list, strings, stringify
    Needs: none (pure function).
    """

def from_json_array(s: str) -> str:
    """Reads a JSON array and puts each element on its own line: strings unquoted, other values (numbers, booleans, null, objects) as compact JSON. Anything that is not a JSON array is an error.

        s: Input text, one item per line.

    Example: lines.from_json_array('["a", "b", 3]')
    Keywords: json, array, to, lines, parse, list, explode, strings
    Needs: none (pure function).
    """

def to_sql_in_list(s: str, sep: str = ",", parens: bool = False, numbers: bool = False, skip_empty: bool = True) -> str:
    """Turns the lines into a SQL list such as 'a','b' with single quotes doubled for escaping. numbers=true leaves values unquoted when every item is a number; parens wraps the list in (...); blank lines are dropped unless skip_empty is false.

        s: Input text, one item per line.
        sep: Separator between items.
        parens: Wrap the list in parentheses.
        numbers: Do not quote numeric items.
        skip_empty: Drop blank lines.

    Example: lines.to_sql_in_list("a\nb")
    Keywords: sql in clause, quote values for sql, sql, in, list, quote, values, where, clause, database, query, strings
    Needs: none (pure function).
    """

def to_quoted_csv(s: str, delimiter: str = ",", quote: str = "\"", skip_empty: bool = False) -> str:
    """Joins the lines into one CSV row in which every field is quoted (quotes inside are doubled), e.g. "a","b". skip_empty drops blank lines.

        s: Input text, one item per line.
        delimiter: Field delimiter.
        quote: Quote character.
        skip_empty: Drop blank lines.

    Example: lines.to_quoted_csv("a\nb")
    Keywords: quoted, csv, row, join, comma, quote, all, fields, list
    Needs: none (pure function).
    """

def to_lf(s: str) -> str:
    """Converts every line ending (CRLF and lone CR) to LF (Unix); all other text is untouched.

        s: Input text, one item per line.

    Example: lines.to_lf("a\r\nb")
    Task: normalize line endings
    Keywords: crlf to lf, unix line endings, dos2unix, lf, unix, line, endings, newline, convert, crlf, normalize
    Needs: none (pure function).
    """

def to_crlf(s: str) -> str:
    """Converts every line ending (LF and lone CR) to CRLF (Windows); existing CRLF is kept as one break.

        s: Input text, one item per line.

    Example: lines.to_crlf("a\nb")
    Keywords: lf to crlf, windows line endings, unix2dos, crlf, windows, line, endings, newline, convert, normalize
    Needs: none (pure function).
    """

def detect_line_endings(s: str) -> str:
    """Reports the line ending style: LF, CRLF, CR, mixed (more than one style) or none (no line break at all).

        s: Input text, one item per line.

    Example: lines.detect_line_endings("a\r\nb")
    Keywords: detect, line, endings, crlf, lf, cr, newline, type, windows, unix, mixed
    Needs: none (pure function).
    """

def strip_trailing_whitespace(s: str) -> str:
    """Removes spaces, tabs and other Unicode whitespace from the end of every line. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.

    Example: lines.strip_trailing_whitespace("a  \nb\t")
    Keywords: trim line ends, remove trailing spaces, trailing, whitespace, strip, trim, right, end, spaces, clean, rstrip
    Needs: none (pure function).
    """

def tabs_to_spaces(s: str, width: int = 4) -> str:
    """Replaces tab characters with spaces, aligning to tab stops every width columns (counted in characters). width is 1 to 16. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        width: Tab stop width.

    Example: lines.tabs_to_spaces("a\tb", 4)
    Keywords: tabs, spaces, convert, expandtabs, indent, untabify
    Needs: none (pure function).
    """

def spaces_to_tabs(s: str, width: int = 4) -> str:
    """Converts leading indentation to tabs: every width spaces at the start of a line become one tab; leftover spaces and spaces elsewhere are kept. width is 1 to 16. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        width: Spaces per tab.

    Example: lines.spaces_to_tabs("        x", 4)
    Keywords: spaces, tabs, convert, indent, tabify, unexpand
    Needs: none (pure function).
    """

def remove_comments(s: str, prefixes: List[str] = [], drop_empty: bool = False) -> str:
    """Deletes lines whose first non-blank characters match one of the prefixes (default # and //). Only whole-line comments are removed, not trailing ones; drop_empty also removes blank lines. Preserves the input's line ending style and one trailing newline.

        s: Input text, one item per line.
        prefixes: Comment markers.
        drop_empty: Also remove blank lines.

    Example: lines.remove_comments("# note\ncode\n// x")
    Keywords: remove, comments, strip, hash, slash, config, code, lines, delete
    Needs: none (pure function).
    """

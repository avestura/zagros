"""CSV: Parse CSV and TSV into rows or dicts, auto-detect the delimiter, and write rows back.

Generated from the tools registry; do not edit by hand.
Type description, not runnable Starlark: do not `import` anything. The name `csv` is predeclared in main.star; call `csv.<name>(...)` directly.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
Keywords: csv, tsv, spreadsheet, parse csv, read csv, write csv, delimiter, rows, columns, table, export, import
"""

from typing import Any, Dict, List

def detect_delimiter(text: str) -> str:
    """Guess the separator of delimited text: comma, semicolon, tab or pipe, chosen by how consistently each appears across the first lines. Returns a comma when nothing fits.

        text: The delimited text.

    Example: csv.detect_delimiter("a;b;c\n1;2;3")
    Keywords: which delimiter, separator, sniff csv, csv, delimiter, sniff
    Needs: none (pure function).
    """

def dump(rows: Any, delimiter: str = ",", columns: List[str] = [], header: bool = True, crlf: bool = False) -> str:
    """Write rows as CSV text. Rows are lists of values or dicts; for dicts the columns are the keys in first-seen order unless columns is given. Cells that need it are quoted. None becomes an empty cell. Raises when a cell is a list or dict.

        rows: A list of lists, or a list of dicts.
        delimiter: Single character separator (use \t for tab).
        columns: Column order for dict rows; empty derives it from the rows.
        header: Write a header line when rows are dicts or columns is set.
        crlf: End lines with CRLF instead of LF.

    Example: csv.dump([{"name": "ada", "age": 36}, {"name": "bob", "age": 41}])
    Keywords: list to csv, table to csv, export csv, tsv output, rows to text, dicts to csv, csv, tsv, write, export, table, rows
    Needs: none (pure function).
    """

def parse(text: str, delimiter: str = "", header: bool = False, trim: bool = False, comment: str = "", strict: bool = False) -> Any:
    """Parse CSV or TSV text. Returns a list of rows (lists of strings), or with header=True a list of dicts keyed by the first row. An empty delimiter detects comma, semicolon, tab or pipe from the first lines. Quotes may be loose unless strict=True. A UTF-8 BOM is dropped. Raises on malformed quoting in strict mode.

        text: The CSV text.
        delimiter: Single character separator (use \t for tab); empty detects it.
        header: Use the first row as keys and return dicts.
        trim: Trim spaces around each cell.
        comment: A character that starts a comment line to skip; empty keeps all lines.
        strict: Reject bare quotes inside unquoted cells.

    Example: csv.parse("name,age\nada,36\nbob,41", header=True)
    Keywords: csv to list, tsv, spreadsheet, read table, split columns, comma separated, parse delimited text, rows to dicts, csv, parse, table, rows, delimiter
    Needs: none (pure function).
    """

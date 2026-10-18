# CSV (`csv`)

Parse CSV and TSV into rows or dicts, auto-detect the delimiter, and write rows back.

No permission is needed: these are pure functions. Generated from the tools registry.

3 functions. Use them as `csv.<name>(...)` in any extension.

## csv.detect_delimiter

`csv.detect_delimiter(text: str) -> str`

Guess the separator of delimited text: comma, semicolon, tab or pipe, chosen by how consistently each appears across the first lines. Returns a comma when nothing fits.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | The delimited text. |

Returns `str`. Example: `csv.detect_delimiter("a;b;c\n1;2;3")`

Keywords: which delimiter, separator, sniff csv, csv, delimiter, sniff

## csv.dump

`csv.dump(rows: Any, delimiter: str = ",", columns: List[str] = [], header: bool = True, crlf: bool = False) -> str`

Write rows as CSV text. Rows are lists of values or dicts; for dicts the columns are the keys in first-seen order unless columns is given. Cells that need it are quoted. None becomes an empty cell. Raises when a cell is a list or dict.

| Parameter | Type | Description |
|---|---|---|
| `rows` | `Any` | A list of lists, or a list of dicts. |
| `delimiter` | `str` (default `","`) | Single character separator (use \t for tab). |
| `columns` | `List[str]` (default `[]`) | Column order for dict rows; empty derives it from the rows. |
| `header` | `bool` (default `True`) | Write a header line when rows are dicts or columns is set. |
| `crlf` | `bool` (default `False`) | End lines with CRLF instead of LF. |

Returns `str`. Example: `csv.dump([{"name": "ada", "age": 36}, {"name": "bob", "age": 41}])`

Keywords: list to csv, table to csv, export csv, tsv output, rows to text, dicts to csv, csv, tsv, write, export, table, rows

## csv.parse

`csv.parse(text: str, delimiter: str = "", header: bool = False, trim: bool = False, comment: str = "", strict: bool = False) -> Any`

Parse CSV or TSV text. Returns a list of rows (lists of strings), or with header=True a list of dicts keyed by the first row. An empty delimiter detects comma, semicolon, tab or pipe from the first lines. Quotes may be loose unless strict=True. A UTF-8 BOM is dropped. Raises on malformed quoting in strict mode.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | The CSV text. |
| `delimiter` | `str` (default `""`) | Single character separator (use \t for tab); empty detects it. |
| `header` | `bool` (default `False`) | Use the first row as keys and return dicts. |
| `trim` | `bool` (default `False`) | Trim spaces around each cell. |
| `comment` | `str` (default `""`) | A character that starts a comment line to skip; empty keeps all lines. |
| `strict` | `bool` (default `False`) | Reject bare quotes inside unquoted cells. |

Returns `Any`. Example: `csv.parse("name,age\nada,36\nbob,41", header=True)`

Keywords: csv to list, tsv, spreadsheet, read table, split columns, comma separated, parse delimited text, rows to dicts, csv, parse, table, rows, delimiter


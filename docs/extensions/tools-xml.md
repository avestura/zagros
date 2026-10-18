# XML (`xml`)

Parse XML into a tree of dicts, query it with simple paths, and write it back.

No permission is needed: these are pure functions. Generated from the tools registry.

5 functions. Use them as `xml.<name>(...)` in any extension.

## xml.dump

`xml.dump(node: Any, indent: str = "  ", declaration: bool = False) -> str`

Write a tree (the shape xml.parse returns) as XML text. Only tag is required; attrs, text and children are optional. Text and attribute values are escaped. Raises on an invalid tag or attribute name or a malformed tree. With indent empty the output is on one line.

| Parameter | Type | Description |
|---|---|---|
| `node` | `Any` | A dict with tag, and optionally attrs, text and children. |
| `indent` | `str` (default `"  "`) | Indentation per level; empty writes a single line. |
| `declaration` | `bool` (default `False`) | Start with the <?xml ...?> declaration. |

Returns `str`. Example: `xml.dump({"tag": "a", "attrs": {"id": "1"}, "children": [{"tag": "b", "text": "hi"}]})`

Keywords: dict to xml, json to xml, write xml, serialize xml, build xml, pretty print xml, xml, dump, write, serialize, pretty

## xml.is_valid

`xml.is_valid(text: str) -> bool`

Report whether text is well-formed XML with a single root element.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | The XML text. |

Returns `bool`. Example: `xml.is_valid("<a><b></a>")`

Keywords: validate xml, check xml, well formed, xml, validate, check

## xml.parse

`xml.parse(text: str) -> Any`

Parse XML text into a tree. Every element is a dict with tag (the name as written, prefix included), attrs (dict of strings), text (the element's own text, trimmed) and children (list of elements). Comments, processing instructions and the DOCTYPE are dropped; CDATA becomes text; the five standard entities and numeric references are decoded. Raises on malformed XML, more than one root, or nesting beyond 256 levels.

| Parameter | Type | Description |
|---|---|---|
| `text` | `str` | The XML text. |

Returns `Any`. Example: `xml.parse("<a id='1'><b>hi</b></a>")`

Keywords: xml to dict, xml to json, read xml, load xml, rss, atom, svg, sitemap, deserialize xml, xml, parse, tree, dom

## xml.query

`xml.query(source: Any, path: str) -> Any`

Select parts of an XML document with a simple path. source is XML text or a tree from xml.parse. The path starts at the document, so the first step names the root: /library/book/title. Steps: a name or *, // for any depth, \[n\] position (1-based), \[last()\], \[@attr\], \[@attr='v'\], \[child='v'\]. A final @attr returns attribute values and text() returns text, both as strings; otherwise the result is a list of elements. Raises on a bad path or XML.

| Parameter | Type | Description |
|---|---|---|
| `source` | `Any` | XML text, or a tree returned by xml.parse. |
| `path` | `str` | The path, for example //item[@id='2']/title or /feed/entry/@href. |

Returns `Any`. Example: `xml.query("<l><b id='1'>x</b><b id='2'>y</b></l>", "/l/b[@id='2']/text()")`

Keywords: xpath, find elements, select nodes, xml search, get attribute, extract from xml, rss items, xml, query, select, find

## xml.text

`xml.text(source: Any, path: str = "", separator: str = " ") -> str`

Return all the text inside an element and its descendants, in document order, joined by the separator. source is XML text or a tree from xml.parse; path is optional and selects the first matching element.

| Parameter | Type | Description |
|---|---|---|
| `source` | `Any` | XML text, or a tree returned by xml.parse. |
| `path` | `str` (default `""`) | Optional path of the element to read; empty uses the root. |
| `separator` | `str` (default `" "`) | Text between pieces. |

Returns `str`. Example: `xml.text("<a>one <b>two</b></a>")`

Keywords: strip tags, inner text, all text, flatten xml, xml, text, strip, inner


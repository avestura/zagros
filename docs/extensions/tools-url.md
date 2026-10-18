# URLs (`url`)

Parse, build, join and normalize URLs and query strings as structured values.

No permission is needed: these are pure functions. Generated from the tools registry.

6 functions. Use them as `url.<name>(...)` in any extension.

## url.build

`url.build(parts: Any) -> str`

Assemble a URL from a dict with any of scheme, user, password, host or hostname plus port, path, query and fragment. query is a dict (values may be strings, numbers, bools or lists of them; None is skipped) or a ready query string. Parts are escaped as needed; a path gets a leading slash when there is a host. Raises when a part has the wrong type or the port is not a number.

| Parameter | Type | Description |
|---|---|---|
| `parts` | `Any` | A dict of URL parts. |

Returns `str`. Example: `url.build({"scheme": "https", "host": "example.com", "path": "/search", "query": {"q": "go lang", "page": 2}})`

Keywords: make url, construct url, dict to url, assemble url, add query params, compose link, url, build, construct, compose

## url.join

`url.join(base: str, ref: str) -> str`

Resolve ref against base the way a browser resolves a link: ../x, /root, ?q=1, #frag, //other.host/x and absolute URLs all work, and dot segments are removed. Raises when either side is malformed.

| Parameter | Type | Description |
|---|---|---|
| `base` | `str` | The base URL. |
| `ref` | `str` | The relative or absolute reference. |

Returns `str`. Example: `url.join("https://example.com/a/b/c", "../d?x=1")`

Keywords: relative url, absolute url, resolve link, combine url, urljoin, href to absolute, url, join, resolve, relative, absolute

## url.normalize

`url.normalize(url: str, sort_query: bool = True, strip_fragment: bool = False, strip_trailing_slash: bool = False) -> str`

Rewrite a URL into a canonical form so equal URLs compare equal: lower-case scheme and host, default port removed, dot segments resolved, percent escapes upper-cased and unnecessary ones decoded, an empty path made /, and the query sorted (sort_query). Options also drop the fragment and a trailing slash. Raises on malformed URLs.

| Parameter | Type | Description |
|---|---|---|
| `url` | `str` | The URL. |
| `sort_query` | `bool` (default `True`) | Sort query parameters by key. |
| `strip_fragment` | `bool` (default `False`) | Remove the #fragment. |
| `strip_trailing_slash` | `bool` (default `False`) | Remove a trailing slash from the path (except for /). |

Returns `str`. Example: `url.normalize("HTTP://Example.com:80/a/./b/../c?b=2&a=1")`

Keywords: canonical url, dedupe urls, clean url, compare urls, same url, canonicalize, url, normalize, canonical, dedupe

## url.parse

`url.parse(url: str, default_scheme: str = "") -> Any`

Split a URL into a dict with scheme, user, password, host (with port), hostname, port, path, query (dict of key to list of values), query_string, fragment and origin. Missing parts are empty strings. With default_scheme set, a URL without a scheme such as example.com/a gets that scheme first. Raises on whitespace, control characters or malformed escapes.

| Parameter | Type | Description |
|---|---|---|
| `url` | `str` | The URL. |
| `default_scheme` | `str` (default `""`) | Scheme to assume when the URL has none, for example https. |

Returns `Any`. Example: `url.parse("https://u:p@example.com:8080/a/b?x=1&x=2#top")`

Keywords: split url, url parts, url to dict, query string, hostname from url, get domain, parse link, url components, url, parse, split, components, hostname, query

## url.query_encode

`url.query_encode(params: Any, sort: bool = False, plus: bool = True) -> str`

Encode a dict as a query string without the leading ?. Values may be strings, numbers, bools or lists (one key=value pair per item); None values are skipped. Keys keep their order unless sort=True. Spaces become + by default, or %20 with plus=False. Raises on nested dicts.

| Parameter | Type | Description |
|---|---|---|
| `params` | `Any` | A dict of parameters. |
| `sort` | `bool` (default `False`) | Sort keys alphabetically. |
| `plus` | `bool` (default `True`) | Write spaces as + instead of %20. |

Returns `str`. Example: `url.query_encode({"q": "go lang", "tags": ["a", "b"], "n": 2})`

Keywords: dict to query string, urlencode, form encode, build query, params to string, url, query, encode, form

## url.query_parse

`url.query_parse(query: str, multi: bool = False) -> Any`

Parse a query string (a leading ? is ignored) into a dict in order of first appearance. Values are percent-decoded and + becomes a space. By default a repeated key keeps its last value as a string; with multi=True every value is a list. A key without = has an empty value. Raises on a malformed percent escape.

| Parameter | Type | Description |
|---|---|---|
| `query` | `str` | The query string, with or without the leading ?. |
| `multi` | `bool` (default `False`) | Return a list of values for every key. |

Returns `Any`. Example: `url.query_parse("?a=1&b=x%20y&a=2", multi=True)`

Keywords: get params, query params to dict, parse query, querystring, url parameters, search params, url, query, params, parse


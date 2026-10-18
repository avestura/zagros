"""URLs: Parse, build, join and normalize URLs and query strings as structured values.

Generated from the tools registry; do not edit by hand.
Type description, not runnable Starlark: do not `import` anything. The name `url` is predeclared in main.star; call `url.<name>(...)` directly.
Needs: none. These are pure functions: no file, network, clipboard or environment access.
Keywords: url, uri, link, parse url, build url, join url, query string, query parameters, normalize url, scheme host path
"""

from typing import Any, Dict, List

def build(parts: Any) -> str:
    """Assemble a URL from a dict with any of scheme, user, password, host or hostname plus port, path, query and fragment. query is a dict (values may be strings, numbers, bools or lists of them; None is skipped) or a ready query string. Parts are escaped as needed; a path gets a leading slash when there is a host. Raises when a part has the wrong type or the port is not a number.

        parts: A dict of URL parts.

    Example: url.build({"scheme": "https", "host": "example.com", "path": "/search", "query": {"q": "go lang", "page": 2}})
    Keywords: make url, construct url, dict to url, assemble url, add query params, compose link, url, build, construct, compose
    Needs: none (pure function).
    """

def join(base: str, ref: str) -> str:
    """Resolve ref against base the way a browser resolves a link: ../x, /root, ?q=1, #frag, //other.host/x and absolute URLs all work, and dot segments are removed. Raises when either side is malformed.

        base: The base URL.
        ref: The relative or absolute reference.

    Example: url.join("https://example.com/a/b/c", "../d?x=1")
    Keywords: relative url, absolute url, resolve link, combine url, urljoin, href to absolute, url, join, resolve, relative, absolute
    Needs: none (pure function).
    """

def normalize(url: str, sort_query: bool = True, strip_fragment: bool = False, strip_trailing_slash: bool = False) -> str:
    """Rewrite a URL into a canonical form so equal URLs compare equal: lower-case scheme and host, default port removed, dot segments resolved, percent escapes upper-cased and unnecessary ones decoded, an empty path made /, and the query sorted (sort_query). Options also drop the fragment and a trailing slash. Raises on malformed URLs.

        url: The URL.
        sort_query: Sort query parameters by key.
        strip_fragment: Remove the #fragment.
        strip_trailing_slash: Remove a trailing slash from the path (except for /).

    Example: url.normalize("HTTP://Example.com:80/a/./b/../c?b=2&a=1")
    Keywords: canonical url, dedupe urls, clean url, compare urls, same url, canonicalize, url, normalize, canonical, dedupe
    Needs: none (pure function).
    """

def parse(url: str, default_scheme: str = "") -> Any:
    """Split a URL into a dict with scheme, user, password, host (with port), hostname, port, path, query (dict of key to list of values), query_string, fragment and origin. Missing parts are empty strings. With default_scheme set, a URL without a scheme such as example.com/a gets that scheme first. Raises on whitespace, control characters or malformed escapes.

        url: The URL.
        default_scheme: Scheme to assume when the URL has none, for example https.

    Example: url.parse("https://u:p@example.com:8080/a/b?x=1&x=2#top")
    Keywords: split url, url parts, url to dict, query string, hostname from url, get domain, parse link, url components, url, parse, split, components, hostname, query
    Needs: none (pure function).
    """

def query_encode(params: Any, sort: bool = False, plus: bool = True) -> str:
    """Encode a dict as a query string without the leading ?. Values may be strings, numbers, bools or lists (one key=value pair per item); None values are skipped. Keys keep their order unless sort=True. Spaces become + by default, or %20 with plus=False. Raises on nested dicts.

        params: A dict of parameters.
        sort: Sort keys alphabetically.
        plus: Write spaces as + instead of %20.

    Example: url.query_encode({"q": "go lang", "tags": ["a", "b"], "n": 2})
    Keywords: dict to query string, urlencode, form encode, build query, params to string, url, query, encode, form
    Needs: none (pure function).
    """

def query_parse(query: str, multi: bool = False) -> Any:
    """Parse a query string (a leading ? is ignored) into a dict in order of first appearance. Values are percent-decoded and + becomes a space. By default a repeated key keeps its last value as a string; with multi=True every value is a list. A key without = has an empty value. Raises on a malformed percent escape.

        query: The query string, with or without the leading ?.
        multi: Return a list of values for every key.

    Example: url.query_parse("?a=1&b=x%20y&a=2", multi=True)
    Keywords: get params, query params to dict, parse query, querystring, url parameters, search params, url, query, params, parse
    Needs: none (pure function).
    """

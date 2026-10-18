"""http: make HTTP requests, upload and download files, read JSON from web APIs.

Type description, not runnable Starlark: do not `import` anything. The name `http` is predeclared in main.star; call `http.get(...)` directly.
Needs: network, scoped by host (network:api.github.com). download also needs fs.write for the target path.
Keywords: web request, fetch url, download, upload, REST API, GET, POST, PUT, PATCH, DELETE, JSON API, internet

Every function raises on network errors, timeouts, a redirect to a host that
is not granted, and bodies over the cap (8 MiB, downloads 256 MiB). A non-2xx
status is not an error: check `r.ok`. Catch errors with `attempt(http.get, url)`.
The per-call time limit of the extension (5 seconds by default) also applies.
"""

from typing import Any, Dict, List, Optional, Tuple, Union

class Response:
    status: int
    ok: bool
    """True for 2xx."""
    url: str
    """Final URL after redirects."""
    headers: Dict[str, str]
    """Lower-case keys, first value of each header."""
    raw_headers: Dict[str, List[str]]
    """Lower-case keys, every value (use for set-cookie)."""
    text: str
    bytes: bytes
    def json(self) -> Any: ...

class Download:
    status: int
    ok: bool
    url: str
    headers: Dict[str, str]
    path: Optional[str]
    """Absolute path written, None when the status was not 2xx (nothing is written then)."""
    size: int

Auth = Union[str, Tuple[str, str]]
"""A string is sent as `Authorization: Bearer <token>`; a (user, password) tuple as Basic."""

def get(url: str, headers: Optional[Dict[str, str]] = None, timeout: int = 10) -> Response:
    """GET a URL.

    Example: r = http.get("https://api.github.com/zen", timeout = 5)
    Task: fetch a web page
    Keywords: download, http get, fetch url, request, call api, web request
    Needs: network
    """

def head(url: str, headers: Optional[Dict[str, str]] = None, timeout: int = 10) -> Response:
    """HEAD: status and headers without a body.

    Example: r = http.head("https://example.com/file.zip")
    Task: check a URL or read its headers
    Keywords: content-length, file size, link check, url alive, ping url, http head
    Needs: network
    """

def post(
    url: str,
    body: str = "",
    json: Any = None,
    headers: Optional[Dict[str, str]] = None,
    timeout: int = 10,
) -> Response:
    """POST. `json` is encoded and sets Content-Type: application/json.

    Example: r = http.post("https://example.com/api", json = {"q": "x"})
    Task: send data to a web API
    Keywords: http post, submit, upload, send json, webhook, call api
    Needs: network
    """

def put(
    url: str,
    body: str = "",
    json: Any = None,
    headers: Optional[Dict[str, str]] = None,
    timeout: int = 10,
) -> Response:
    """PUT, same arguments as post.

    Example: r = http.put("https://example.com/api/items/1", json = {"name": "x"})
    Task: replace a resource on a web API
    Keywords: http put, update resource, upsert, rest update, replace
    Needs: network
    """

def patch(
    url: str,
    body: str = "",
    json: Any = None,
    headers: Optional[Dict[str, str]] = None,
    timeout: int = 10,
) -> Response:
    """PATCH, same arguments as post.

    Example: r = http.patch("https://example.com/api/items/1", json = {"done": True})
    Task: partially update a resource on a web API
    Keywords: http patch, partial update, modify resource, rest update
    Needs: network
    """

def delete(
    url: str,
    body: str = "",
    json: Any = None,
    headers: Optional[Dict[str, str]] = None,
    timeout: int = 10,
) -> Response:
    """DELETE, same arguments as post.

    Example: r = http.delete("https://example.com/api/items/1")
    Task: delete a resource on a web API
    Keywords: http delete, remove resource, rest delete, destroy
    Needs: network
    """

def request(
    method: str,
    url: str,
    headers: Optional[Dict[str, str]] = None,
    body: Union[str, bytes, None] = None,
    json: Any = None,
    form: Optional[Dict[str, str]] = None,
    multipart: Optional[Dict[str, Any]] = None,
    timeout: float = 10,
    follow_redirects: bool = True,
    auth: Optional[Auth] = None,
    cookies: bool = False,
    max_bytes: int = 0,
) -> Response:
    """Any method with full control. Give at most one of body, json, form (urlencoded) and multipart.

    multipart maps field names to a str (text field), bytes (file part named after the field), or a dict
    {"filename": str, "content": str | bytes, "path": str, "content_type": str}; "path" reads that file
    (needs fs.read). follow_redirects=False returns the 3xx response itself. cookies=True keeps cookies in
    this extension's in-memory jar across calls (see clear_cookies). max_bytes lowers the response cap.

    Example: r = http.request("POST", "https://example.com/login", form = {"user": "me", "pass": "pw"}, cookies = True, follow_redirects = False)
    Task: make a custom HTTP request
    Keywords: http request, any method, basic auth, bearer token, form post, multipart upload, file upload, cookies, session, no redirect, options
    Needs: network (and fs.read when a multipart part uses "path")
    """

def download(
    url: str,
    path: str,
    headers: Optional[Dict[str, str]] = None,
    timeout: float = 60,
    auth: Optional[Auth] = None,
    cookies: bool = False,
    overwrite: bool = True,
    max_bytes: int = 0,
) -> Download:
    """Stream a URL into a file (via a .part file renamed on success). Nothing is written for a non-2xx status. Capped at 256 MiB (1 GiB inside a task) or max_bytes. Long downloads belong in a task(): a hook stops after 5 seconds, a task after 5 minutes.

    Example: d = http.download("https://example.com/report.pdf", "~/Documents/report.pdf")
    Task: download a file from the internet
    Keywords: save url to file, wget, curl -o, fetch file, download file, save attachment
    Needs: network and fs.write for the path
    """

class FetchError:
    """The entry http.fetch_all returns for a request that failed. It has no body fields."""

    ok: bool
    """Always False."""
    status: int
    """Always 0."""
    url: str
    error: str
    """Why it failed: permission denied, timeout, size cap, network error."""

def fetch_all(requests: List[Union[str, Dict[str, Any]]], max_concurrency: int = 4) -> List[Union[Response, FetchError]]:
    """Send many requests in parallel and return the results in the same order.

    Each entry is a URL string (GET) or a dict with the arguments of request()
    (url, method defaulting to "GET", headers, body, json, form, multipart,
    timeout, follow_redirects, auth, cookies, max_bytes). A failed entry does not
    raise: it comes back as a FetchError, so check `r.ok` on every result. Network
    scoping, redirect checks and the 8 MiB response cap are per entry; the
    responses of one call may add up to 32 MiB. At most 100 requests and
    max_concurrency 8. The time limit of the hook (5 seconds) covers the whole
    call, so inside on_query keep the batch small or run it from a task().

    Example:
        rs = http.fetch_all(["https://api.github.com/zen", {"url": "https://example.com/api", "method": "POST", "json": {"q": "x"}}], max_concurrency = 2)
        zen = rs[0].text if rs[0].ok else rs[0].error
    Task: fetch several URLs at the same time
    Keywords: parallel requests, concurrent, batch, many urls, fan out, multiple requests, gather, speed up http
    Needs: network
    """

def clear_cookies() -> None:
    """Forget the cookies kept by cookies=True requests.

    Example: http.clear_cookies()
    Keywords: logout, reset session, forget cookies
    Needs: none
    """

def json(url: str, headers: Optional[Dict[str, str]] = None, timeout: int = 10) -> Any:
    """GET and decode JSON; fails on non-2xx.

    Example: data = http.json("https://api.github.com/repos/avestura/zagros")
    Task: read JSON from a web API
    Keywords: fetch json, rest api, get json, call api, parse response
    Needs: network
    """

def quote(text: str) -> str:
    """URL-escape a query component.

    Example: q = http.quote("a b&c")
    Keywords: urlencode, percent encode, escape query string, encode url parameter
    Needs: none
    """

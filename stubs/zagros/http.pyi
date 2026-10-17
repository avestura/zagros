"""http module. Needs: network (scoped by host, e.g. network:api.github.com)."""

from typing import Any, Dict, Optional

class Response:
    status: int
    ok: bool
    url: str
    headers: Dict[str, str]
    """Lower-case keys."""
    text: str
    def json(self) -> Any: ...

def get(url: str, headers: Optional[Dict[str, str]] = None, timeout: int = 10) -> Response:
    """GET a URL. Needs: network."""

def post(
    url: str,
    body: str = "",
    json: Any = None,
    headers: Optional[Dict[str, str]] = None,
    timeout: int = 10,
) -> Response:
    """POST. `json` is encoded and sets Content-Type: application/json. Needs: network."""

def json(url: str, headers: Optional[Dict[str, str]] = None, timeout: int = 10) -> Any:
    """GET and decode JSON; fails on non-2xx. Needs: network."""

def quote(text: str) -> str:
    """URL-escape a query component. Needs: none."""

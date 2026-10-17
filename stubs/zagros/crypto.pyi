"""crypto module. Needs: none."""

from typing import Union

def sha256(data: Union[str, bytes]) -> str:
    """Hex SHA-256."""

def md5(data: Union[str, bytes]) -> str:
    """Hex MD5."""

def hmac(key: Union[str, bytes], data: Union[str, bytes], algo: str = "sha256") -> str:
    """Hex HMAC. algo: sha256, sha1, sha512 or md5."""

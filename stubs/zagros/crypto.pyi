"""crypto: the three hashing helpers every extension has (use the `hash` module for more algorithms).

Type description, not runnable Starlark: do not `import` anything. The name `crypto` is predeclared in main.star; call `crypto.sha256(...)` directly.
Needs: none.
Keywords: hash, sha256, md5, hmac, digest, checksum, signature
"""

from typing import Union

def sha256(data: Union[str, bytes]) -> str:
    """Hex SHA-256 of a string or bytes.

    Example: crypto.sha256("hello")
    Keywords: sha-256, digest, checksum, fingerprint, hash text
    Needs: none
    """

def md5(data: Union[str, bytes]) -> str:
    """Hex MD5 of a string or bytes.

    Example: crypto.md5("hello")
    Keywords: md5sum, digest, checksum, fingerprint, hash text
    Needs: none
    """

def hmac(key: Union[str, bytes], data: Union[str, bytes], algo: str = "sha256") -> str:
    """Hex HMAC. algo: sha256, sha1, sha512 or md5.

    Example: crypto.hmac("secret", "message", algo = "sha256")
    Keywords: keyed hash, message authentication code, sign with a secret, webhook signature
    Needs: none
    """

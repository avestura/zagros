"""secrets module: per-extension secrets in Windows Credential Manager. Needs: secrets.

Use it for API tokens and passwords instead of storage or settings, which are
plain files. Each extension has its own namespace (keyed by its manifest id),
so one extension cannot read or overwrite another's secrets. Entries appear in
Windows Credential Manager under the name "zagros|<extension id>|<key>".
Values are strings up to 2560 bytes; keys are 1 to 256 bytes. Windows only:
on other systems every call raises an error. Secret values are never logged.

Type description, not runnable Starlark: do not `import` anything. The name `secrets` is predeclared in main.star; call `secrets.get(...)` directly.
Keywords: password, api token, credential, keychain, credential manager, vault, store secret, remember login
"""

from typing import Optional

def set(key: str, value: str) -> None:
    """Save a secret, replacing any earlier value for the key.

    Needs: secrets.
    Example:
        secrets.set("api_token", "abc123")
    Keywords: save password, store token, api key, credential, keychain, remember login
    """

def get(key: str, default: Optional[str] = None) -> Optional[str]:
    """Read a secret, or `default` when the key was never set.

    Needs: secrets.
    Example:
        token = secrets.get("api_token", "")
    Keywords: read password, load token, api key, credential, keychain, saved login
    """

def delete(key: str) -> None:
    """Remove a secret. Deleting a missing key is not an error.

    Needs: secrets.
    Example:
        secrets.delete("api_token")
    Keywords: forget password, remove token, sign out, clear credential, revoke
    """

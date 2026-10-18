"""host module: facts about the running Zagros. Needs: none.

Use it to adapt to the Zagros version an extension runs on. To look for a
newer Zagros release yourself, call http.json("https://api.github.com/repos/avestura/zagros/releases/latest")
(needs the permission network:api.github.com) and compare the tag with
compare_versions().
"""

def version() -> str:
    """The running Zagros version, for example "0.2.0", or "dev" for an unversioned build."""

def compare_versions(a: str, b: str) -> int:
    """Compare two semantic versions ("1.2.3", "v1.2.3", "1.2.3-rc.1"): -1 if a < b,
    0 if equal, 1 if a > b. A pre-release is older than its release. Raises an error
    when either argument is not a semantic version."""

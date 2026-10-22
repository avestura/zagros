"""host: facts about the running Zagros, to adapt to its version.

Type description, not runnable Starlark: do not `import` anything. The name `host` is predeclared in main.star; call `host.version()` directly.
To look for a newer Zagros release yourself, call http.json("https://api.github.com/repos/avestura/zagros/releases/latest")
(needs the permission network:api.github.com) and compare the tag with compare_versions().
Needs: none.
Keywords: version, zagros version, semver, release, update check, compatibility
"""

def version() -> str:
    """The running Zagros version, for example "0.2.0", or "dev" for an unversioned build.

    Example: v = host.version()
    Task: read the Zagros version
    Keywords: current version, app version, which zagros, build version
    Needs: none
    """

def reindex_apps() -> bool:
    """Rescan the launcher's app index (Start Menu, scan_dirs, Store apps) in the background now,
    instead of waiting for the 12 hour cache to expire. Returns True when a scan started, False
    when one is already running or the apps provider is off. Use it after your extension installs
    or removes an application.

    Example: started = host.reindex_apps()
    Task: refresh the launcher app list
    Keywords: reindex, rescan apps, refresh app index, app cache, start menu scan
    Needs: none
    """

def compare_versions(a: str, b: str) -> int:
    """Compare two semantic versions ("1.2.3", "v1.2.3", "1.2.3-rc.1"): -1 if a < b,
    0 if equal, 1 if a > b. A pre-release is older than its release. Raises an error
    when either argument is not a semantic version.

    Example: newer = host.compare_versions("1.2.0", host.version()) > 0
    Task: compare two version numbers
    Keywords: semver compare, is newer, version order, greater than version, update available
    Needs: none
    """

"""windows module: read-only Windows system information. Windows only.

On other systems every call raises an error. There are no write, start, stop
or kill operations. Registry reads need `registry.read` (scoped by key
prefix, for example `registry.read:HKCU\\Software\\Microsoft`); everything
else needs `system.read`. To open a file or URL use action_open (needs `open`).
Registry keys are written "HKCU\\Software\\X" or "HKLM\\..."; only HKCU and
HKLM are supported, and forward slashes are accepted.

Type description, not runnable Starlark: do not `import` anything. The name `windows` is predeclared in main.star; call `windows.processes()` directly.
Needs: registry.read for the registry functions, system.read for the rest.
Keywords: windows, registry, services, processes, running apps, task manager, system info, known folders, window list
"""

from typing import Any, Dict, List, Optional

def registry_get(key: str, name: str = "") -> Any:
    """Read one registry value: a string, an int (DWORD/QWORD), a list of strings (MULTI_SZ) or bytes. An empty `name` reads the key's default value. Returns None when the value does not exist; raises when the key cannot be opened.

    Needs: registry.read (scoped by key prefix, e.g. registry.read:HKCU\\Software\\Microsoft).
    Example:
        build = windows.registry_get("HKLM\\SOFTWARE\\Microsoft\\Windows NT\\CurrentVersion", "CurrentBuild")
    Keywords: read registry, regedit, registry value, hklm, hkcu, windows version, installed software setting
    """

def registry_keys(key: str) -> List[str]:
    """Names of the subkeys of a registry key.

    Needs: registry.read (scoped by key prefix).
    Example:
        names = windows.registry_keys("HKCU\\Software\\Microsoft\\Windows\\CurrentVersion")
    Keywords: list registry keys, subkeys, browse registry, regedit, enumerate registry
    """

def registry_values(key: str) -> Dict[str, Any]:
    """All values of a registry key as {name: value}, typed as in registry_get. Values that cannot be read are skipped.

    Needs: registry.read (scoped by key prefix).
    Example:
        vals = windows.registry_values("HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Run")
    Keywords: list registry values, startup programs, autorun, dump registry key, regedit
    """

def services() -> List[Dict[str, Any]]:
    """All Windows services as dicts with name, display_name, state (stopped, running, paused, start_pending, ...), start_type (automatic, manual, disabled, boot, system) and pid (0 when not running).

    Needs: system.read.
    Example:
        running = [s["display_name"] for s in windows.services() if s["state"] == "running"]
    Keywords: list services, services.msc, running services, background services, daemons, service manager
    """

def service_status(name: str) -> Dict[str, Any]:
    """Status of one service by its short name (for example "Spooler"): the same dict as services(). Raises when the service does not exist.

    Needs: system.read.
    Example:
        st = windows.service_status("Spooler")["state"]
    Keywords: is service running, check service, service state, print spooler, service health
    """

def list_windows() -> List[Dict[str, Any]]:
    """Visible top-level windows that have a title, as dicts with title, process (executable name) and pid. Tool windows and cloaked windows (other virtual desktops, suspended apps) are left out.

    Needs: system.read.
    Example:
        titles = [w["title"] for w in windows.list_windows()]
    Keywords: open windows, window switcher, alt tab, running apps, window titles, task list
    """

def processes() -> List[Dict[str, Any]]:
    """Running processes as dicts with name (executable file name) and pid.

    Needs: system.read.
    Example:
        names = [p["name"] for p in windows.processes()]
    Keywords: running processes, task manager, process list, is app running, pid, tasklist
    """

def known_folder(name: str) -> str:
    """Absolute path of a known folder. Names: documents, downloads, desktop, pictures, music, videos, appdata (roaming), localappdata, programdata, profile.

    Needs: system.read.
    Example:
        downloads = windows.known_folder("downloads")
    Keywords: documents folder, downloads folder, appdata, user profile path, special folders, home directory
    """

"""ssh: run commands and transfer files on a server over SSH (with SFTP).

Type description, not runnable Starlark: do not `import` anything. The name `ssh` is predeclared in main.star; call `ssh.connect(...)` directly.
Needs: ssh, scoped by [user@]host[:port] (ssh:deploy@build.example.com:22). The host must match exactly; no user means any user, no port means any port. auth={"secret": ...} also needs secrets; download needs fs.write and upload needs fs.read for the local path.
Keywords: ssh, remote command, remote shell, sftp, scp, server, deploy, upload file, download file, run on server, remote exec

Host keys are verified. The first key seen for a host:port is recorded in the
extension's data folder (trust on first use), and a different key later raises
an error that names both fingerprints. Pass host_key="SHA256:..." to pin the
fingerprint instead (it overrides the record), or trust_new=False to refuse
hosts that are not yet on record; the error then shows the fingerprint to check.
Verification cannot be switched off. Credentials: prefer auth={"secret": name}
or {"key_secret": name} (read from `secrets`) over raw strings. Output of exec
and file reads are capped at 8 MiB (downloads and uploads 256 MiB); a connection
lives at most two minutes and closes when the call ends. Errors raise.
"""

from typing import Any, Dict, List, Optional, Union

Data = Union[str, bytes]

class Result:
    stdout: str
    stderr: str
    code: int
    """Exit status; -1 when the server reported none."""
    ok: bool
    """True when code is 0."""

class Entry:
    name: str
    path: str
    size: int
    mtime: int
    is_dir: bool
    is_file: bool
    is_symlink: bool
    mode: str

class Conn:
    user: str
    fingerprint: str
    """SHA256 fingerprint of the server host key."""
    host_key_status: str
    """"pinned", "known" or "new" (recorded just now)."""
    def exec(self, cmd: str, timeout: float = 30, stdin: Optional[Data] = None) -> Result:
        """Run a command and wait for it. A non-zero exit is not an error: check ok or code. Raises on timeout or output over the cap.

        Example: r = conn.exec("df -h /", timeout=10)
        Task: run a command on the server
        Keywords: remote command, shell, execute, run script, uptime, systemctl, deploy step
        Needs: none
        """
    def get(self, path: str) -> bytes:
        """Read a remote file into memory (up to 8 MiB).

        Example: data = conn.get("/etc/hostname")
        Task: read a remote file
        Keywords: sftp read, cat file, fetch remote file
        Needs: none
        """
    def put(self, path: str, data: Data) -> int:
        """Write data to a remote file, replacing it; returns the byte count.

        Example: conn.put("/tmp/note.txt", "hello")
        Task: write a remote file
        Keywords: sftp write, upload text, create remote file
        Needs: none
        """
    def download(self, remote: str, local: str) -> int:
        """Copy a remote file to a local path; returns the size. Needs fs.write for local.

        Example: n = conn.download("/var/log/app.log", "~/Downloads/app.log")
        Task: download a file from the server
        Keywords: scp, sftp get, copy from server, fetch file
        Needs: fs.write
        """
    def upload(self, local: str, remote: str) -> int:
        """Copy a local file to the server; returns the size. Needs fs.read for local.

        Example: n = conn.upload("~/Documents/site.zip", "/srv/site.zip")
        Task: upload a file to the server
        Keywords: scp, sftp put, copy to server, deploy file
        Needs: fs.read
        """
    def list(self, path: str = ".") -> List[Entry]:
        """List a remote folder.

        Example: entries = conn.list("/var/log")
        Task: list files on the server
        Keywords: ls, dir, sftp list, remote folder
        Needs: none
        """
    def stat(self, path: str) -> Entry:
        """Describe one remote file or folder (not following links). Raises when missing.

        Example: info = conn.stat("/etc/hostname")
        Task: get remote file size and time
        Keywords: file info, exists, size, modified time
        Needs: none
        """
    def remove(self, path: str) -> None:
        """Delete a remote file or an empty folder. Refuses "/" and "~".

        Example: conn.remove("/tmp/note.txt")
        Task: delete a remote file
        Keywords: rm, delete, unlink, rmdir
        Needs: none
        """
    def mkdir(self, path: str, parents: bool = False) -> None:
        """Create a remote folder; parents=True also creates missing parents.

        Example: conn.mkdir("/tmp/work/out", parents=True)
        Task: create a remote folder
        Keywords: mkdir -p, make directory
        Needs: none
        """
    def close(self) -> None:
        """Close the connection early.

        Example: conn.close()
        Needs: none
        """

def connect(host: str, user: str, port: int = 22, auth: Optional[Dict[str, str]] = None, timeout: float = 10, host_key: Union[str, List[str], None] = None, trust_new: bool = True) -> Conn:
    """Open an SSH connection. auth is required and takes: "secret" (password stored with secrets.set, by name), "key_secret" (private key PEM stored in secrets), "passphrase_secret" (for an encrypted key); or the raw "password", "key", "passphrase" strings.

    Example: conn = ssh.connect("build.example.com", "deploy", auth={"key_secret": "deploy_key"})
    Task: connect to a server over SSH
    Keywords: ssh login, remote server, private key, password auth, host key, known hosts, fingerprint, tofu, pin host key
    Needs: ssh
    """

# Permissions

Extensions run in a Starlark sandbox. By default an extension can do nothing except return results. Anything else needs a permission that it declares in its manifest, in the style of Android: Zagros shows a consent dialog when you install the extension, and you can review or revoke grants later.

## The model

- **Declared up front.** Permissions live in the extension manifest, so the marketplace can show them before you install.
- **Granted by you.** Install shows each permission in plain language. Declining installs nothing.
- **Enforced by the host.** The host API checks the grant on every call. A call without a grant fails with an error the extension can catch; it is never silently allowed.
- **Re-asked on change.** If an update adds a permission, you are asked again before the update applies.

## What needs no permission

Interface mechanisms stay inside the launcher and need no permission: results (`item`), scopes (the Tab sub-context), Tab completion, alternate actions, `settings`, `storage`, `log`, `json`, `time` and `crypto`. A host feature that reaches outside the sandbox always gets its own permission, listed in the stubs (`permissions.pyi`) and in the extensions guide. Coding agents can read the full list in the [API stubs](../../docs/extensions/agents.md).

## Files, downloads and archives

- `fs.read` and `fs.write` are scoped to a folder, for example `fs.write:~/Documents`. The host resolves symlinks and `..` before it checks the folder, so a link inside a granted folder cannot lead out of it. Reading, listing, searching (`fs.glob`, `fs.walk`), copying and archiving need `fs.read`; writing, appending, creating folders, renaming, deleting and copying into a place need `fs.write`. Deleting or moving the granted folder itself, your home folder or a drive root is refused.
- `http.download` and the file parts of `http.request` use both permissions: `network` for the host and `fs.write` or `fs.read` for the file.
- `compress.create`, `compress.extract` and `compress.list` work on files, so they need `fs.read` for the archive or sources and `fs.write` for what they create. Archive entries that try to leave the destination are rejected, and extraction stops at fixed size and entry limits.
- `sqlite` gives each extension its own database file under its data folder. It needs only the always-granted `storage` permission because the extension cannot name another path; `ATTACH` and similar statements are blocked.
- Byte helpers (`binary`, the in-memory `compress.gzip` family) and `attempt` touch nothing outside the sandbox and need no permission.

## Secrets and system information

- `secrets` lets an extension keep tokens and passwords in Windows Credential Manager. Each extension only sees its own entries.
- `registry.read` lets an extension read the registry (`HKCU` and `HKLM` only). It is scoped to a key prefix, for example `registry.read:HKCU\Software\Vendor`, and the consent dialog shows that prefix.
- `system.read` lets an extension list services, running processes, open window titles and the paths of known folders such as Downloads. It cannot start, stop or change anything. Window titles and process names can show what you are working on, so grant it only to extensions that need it.

## Consent dialog

The dialog lists the extension name, its author, the install source (`github:owner/repo` or a registry id), and each requested permission with a short explanation. Dangerous permissions, such as running programs, are marked with a warning.

See the [extensions guide](../../docs/extensions/README.md) for the exact permission names and the host API they unlock, and [marketplace](../../docs/marketplace.md) for how permissions are shown before install.

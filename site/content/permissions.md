# Permissions

Extensions run in a Starlark sandbox. By default an extension can do nothing except return results. Anything else needs a permission that it declares in its manifest, in the style of Android: Zagros shows a consent dialog when you install the extension, and you can review or revoke grants later.

## The model

- **Declared up front.** Permissions live in the extension manifest, so the marketplace can show them before you install.
- **Granted by you.** Install shows each permission in plain language. Declining installs nothing.
- **Enforced by the host.** The host API checks the grant on every call. A call without a grant fails with an error the extension can catch; it is never silently allowed.
- **Re-asked on change.** If an update adds a permission, you are asked again before the update applies.

## What needs no permission

Interface mechanisms stay inside the launcher and need no permission: results (`item`), scopes (the Tab sub-context), Tab completion, alternate actions, `settings`, `storage`, `log`, `json`, `time` and `crypto`. A host feature that reaches outside the sandbox always gets its own permission, listed in the stubs (`permissions.pyi`) and in the extensions guide. Coding agents can read the full list in the [API stubs](../../docs/extensions/agents.md).

## Consent dialog

The dialog lists the extension name, its author, the install source (`github:owner/repo` or a registry id), and each requested permission with a short explanation. Dangerous permissions, such as running programs, are marked with a warning.

See the [extensions guide](../../docs/extensions/README.md) for the exact permission names and the host API they unlock, and [marketplace](../../docs/marketplace.md) for how permissions are shown before install.

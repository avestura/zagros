# Marketplace

Zagros has no marketplace server. Extensions live in their authors' GitHub
repositories. A static index, `registry.json`, is built from pull requests to
this repo and served from GitHub Pages:

    https://avestura.github.io/zagros/registry.json

The registry is only a discovery aid. Anything can also be installed straight
from GitHub or a URL without being listed.

## Install references

| Syntax | Meaning |
| --- | --- |
| `hello-world` | Registry id. Looked up in `registry.json`. |
| `github:owner/repo` | Latest GitHub release of the repo. |
| `github:owner/repo@v1.2.0` | A specific release tag. If no release has that tag, the repo is fetched at that tag, branch or commit. |
| `owner/repo[@tag]` | Same as the `github:` form. |
| `https://github.com/owner/repo` | Same as `github:owner/repo`. `/tree/<ref>` and `/releases/tag/<tag>` are understood. |
| `https://host/path/ext.zagros` | A direct archive. Host must be on the allowlist below. |

Resolution order for GitHub refs:

1. Release asset ending in `.zagros`.
2. Release asset ending in `.zip` that has `manifest.toml` at its root.
3. The repository zipball at the ref (default branch if none). It is repacked
   and accepted when `manifest.toml` is at the repo root, or in `subdir` when
   the registry entry sets one.

Downloads are capped at 20 MB, use a 30 s timeout, follow at most 5 redirects,
and must be `https`. Allowed hosts: `github.com`, `api.github.com`,
`objects.githubusercontent.com`, `codeload.github.com`, and the registry host.
The SHA-256 of the downloaded bytes is computed and returned to the caller. If
the registry entry has a `sha256` pin, a mismatch aborts the install.

Set `GITHUB_TOKEN` to raise the GitHub API rate limit (60 requests per hour
unauthenticated). The token is only sent to the API host.

## Publishing an extension

1. Put `manifest.toml` at the root of your repo (or a subfolder, see `subdir`).
2. Tag a release (`v1.0.0`) and attach the packaged `.zagros` file (a zip with
   `manifest.toml` at its root). The tag should match the manifest `version`.
3. Add the topic `zagros-extension` to the repo so people can find it on GitHub.
4. To be listed in the registry, open a PR adding
   `registry/extensions/<id>.toml`. The file name must equal the extension id
   and the `id` in your manifest.

```toml
repo = "you/zagros-hello"       # required
tags = ["demo", "text"]         # optional
# ref    = "v1.0.0"             # pin a tag/branch instead of the latest release
# subdir = "extensions/hello"   # manifest.toml is not at the repo root
# kind   = "theme"              # single-file theme, see Themes below
# asset  = "hello-*.zagros"     # glob when a release has several assets
# sha256 = "<64 hex chars>"     # pin the exact archive
```

CI runs `go run ./cmd/registrygen` on every PR. It fetches the latest release
and `manifest.toml` of each listed repo, checks that the ids match and the
version is semver, and fails the PR on any error. After merge, CI writes
`site/registry.json` and publishes it. Name, description, author, permissions
and `latest` in the registry always come from your manifest; do not put them in
the TOML (they are only read by `registrygen -offline`).

Try it locally:

    go run ./cmd/registrygen -offline -out /tmp/registry.json   # no network
    go run ./cmd/registrygen                                    # queries GitHub

## registry.json

    {"schema": 1, "generated": "...", "extensions": [
      {"id", "name", "description", "author", "repo": "owner/name", "tags",
       "latest", "permissions", "verified", "kind"?, "downloads"?, "ref"?, "subdir"?,
       "asset"?, "sha256"?}]}

See `registry/schema.json`. The client caches the file in `<DataDir>/cache`
with its ETag and uses the cached copy when offline.

## Trust model

- Extensions run in a sandbox and can only use the permissions declared in
  their manifest. Before every install or update that adds permissions, Zagros
  shows the list and asks for consent.
- The registry listing is not a security review. Anyone can send a PR.
- `verified = true` means a maintainer reviewed the extension's code at the
  listed version and its declared permissions fit what it does. It is set by
  maintainers in the TOML file only. A new release is not covered until the
  maintainers re-verify; updates to verified extensions that add permissions
  drop the flag.
- Installs from `github:` refs and URLs bypass the registry, so they never
  carry the verified flag. Treat them like any downloaded program.
- Authors can pin an archive with `sha256` so a moved tag cannot change what is
  installed.

## Themes

A registry entry with `kind = "theme"` (default `"extension"`) installs a single
`.toml` theme file into the themes folder (`core.ThemeDir()`, as
`<id>.toml`) instead of unpacking an extension. Use `subdir` for the file's path in the
repo, or `asset` for a glob matching a release asset. `ref` and `sha256` work as
for extensions.

```toml
kind   = "theme"
repo   = "you/zagros-themes"
subdir = "themes/midnight.toml"
tags   = ["theme", "dark"]
```

`Client.InstallTheme(ctx, themeDir, ref)` accepts a registry id or an https
URL ending in `.toml`. The file must be at most 64 KB and parse with
`theme.Parse`; otherwise nothing is written. Themes need no permissions and
have no code. Theme entries have an empty `permissions` list in
`registry.json` and carry `"kind": "theme"`.

Commands:

    zagros market list                    kind, id, description
    zagros market search <query>
    zagros theme install <id|https url to .toml>
    zagros theme remove <name>            user and installed themes only
    zagros theme list                     built-in / user / installed

Themes land in the themes folder as `<id>.toml`. Which ones came from the
marketplace is recorded in `<DataDir>/installed-themes.json`, which is what
separates "installed" from "user". Installing does not switch themes: pick it
in the extension manager's Themes tab ("Use"), or set `theme = "<name>"` in
`zagros.toml`. A running instance reloads after each install or removal.
`zagros ext install <id>` on a theme id fails and points to `theme install`.
In the extension manager, the Marketplace tab labels entries EXT or THEME and
installs themes through the same button.

## Starter content (`marketplace/`)

This repo ships starter content that installs straight from the monorepo; the
entries in `registry/extensions/` set `repo = "avestura/zagros"` and a
`subdir` under `marketplace/`.

| Folder | Contents |
| --- | --- |
| `marketplace/themes/<id>.toml` | solarized-dark, solarized-light, dracula, gruvbox-dark, catppuccin-mocha, tokyo-night, boxy-blueprint |
| `marketplace/extensions/<id>/` | wiki, hackernews, npm-search, weather, dictionary, snippets (each has `manifest.toml`, `main.star`, `README.md`) |

Entries without a pinned `ref` or `asset` and with a `subdir` install from the
default branch of the repo; the repo's own application releases are ignored.

`go test ./marketplace` parses every theme (text on background at least 4.5:1,
selected text at least 3:1), validates every extension with the `ext`
validator, runs `on_query` against stubbed HTTP, and checks that each theme and
extension has a registry entry.

`pwsh scripts/pack-marketplace.ps1` writes `dist/marketplace/<id>.zagros` for
each extension and copies the themes next to them, ready to attach to a release.

`go run ./cmd/registrygen -offline` reads `marketplace/` locally (flag `-root`
sets the repo root) and fills name, description, author, version and
permissions from the manifests and theme files.

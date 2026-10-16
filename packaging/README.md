# Packaging

Manifests for Scoop and winget. The 0.1.0 manifests contain placeholder
SHA256 values (all zeros, marked `PLACEHOLDER` in the winget files) and a
placeholder winget `ReleaseDate`. Replace them before publishing.

Release assets are `zagros_<version>_windows_amd64.zip` and
`zagros_<version>_windows_arm64.zip`, each containing `zagros.exe`, plus
`checksums.txt`.

## Updating

```powershell
./scripts/update-packaging.ps1 -Version 0.2.0 -ChecksumsFile path\to\checksums.txt
# or
./scripts/update-packaging.ps1 -Version 0.2.0 -Sha256Amd64 <hex> -Sha256Arm64 <hex>
```

This rewrites `scoop/zagros.json` and renames and updates the winget folder
`winget/manifests/a/Avestura/Zagros/<version>/`. Pass `-ReleaseDate yyyy-MM-dd`
to override today's date. The script drops the `PLACEHOLDER` comments once it
writes real hashes.

## Scoop

`scoop/zagros.json` has `checkver` and `autoupdate` configured, so it can live
in a bucket and be maintained by Scoop's Excavator. Test locally with
`scoop install .\packaging\scoop\zagros.json`.

## winget

The manifests use schema 1.9.0, identifier `Avestura.Zagros`, a zip installer
with a nested portable `zagros.exe` (command alias `zagros`). Validate with
`winget validate packaging\winget\manifests\a\Avestura\Zagros\<version>`, then
submit the folder to `microsoft/winget-pkgs` (or use `wingetcreate`).

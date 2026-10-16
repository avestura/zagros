<#
.SYNOPSIS
  Rewrites the Scoop and winget manifests for a new Zagros release.
.EXAMPLE
  ./scripts/update-packaging.ps1 -Version 0.2.0 -ChecksumsFile dist/checksums.txt
  ./scripts/update-packaging.ps1 -Version 0.2.0 -Sha256Amd64 <hex> -Sha256Arm64 <hex>
#>
[CmdletBinding()]
param(
    [Parameter(Mandatory)][string]$Version,
    [string]$Sha256Amd64,
    [string]$Sha256Arm64,
    [string]$ChecksumsFile,
    [string]$ReleaseDate = (Get-Date -Format 'yyyy-MM-dd')
)
$ErrorActionPreference = 'Stop'
$Version = $Version.TrimStart('v')
$root = Split-Path -Parent $PSScriptRoot
$amdZip = "zagros_${Version}_windows_amd64.zip"
$armZip = "zagros_${Version}_windows_arm64.zip"

if ($ChecksumsFile) {
    foreach ($line in Get-Content $ChecksumsFile) {
        if ($line -match '^([0-9a-fA-F]{64})\s+\*?(.+)$') {
            if ($Matches[2].Trim() -eq $amdZip) { $Sha256Amd64 = $Matches[1] }
            if ($Matches[2].Trim() -eq $armZip) { $Sha256Arm64 = $Matches[1] }
        }
    }
}
foreach ($h in @($Sha256Amd64, $Sha256Arm64)) {
    if ($h -notmatch '^[0-9a-fA-F]{64}$') {
        throw 'Provide valid -Sha256Amd64 and -Sha256Arm64, or a -ChecksumsFile containing both zips.'
    }
}
$Sha256Amd64 = $Sha256Amd64.ToUpper(); $Sha256Arm64 = $Sha256Arm64.ToUpper()
$base = "https://github.com/avestura/zagros/releases/download/v$Version"
$utf8 = New-Object System.Text.UTF8Encoding($false)

# Scoop
$scoopPath = Join-Path $root 'packaging/scoop/zagros.json'
$s = Get-Content $scoopPath -Raw | ConvertFrom-Json
$s.version = $Version
$s.architecture.'64bit'.url = "$base/$amdZip"
$s.architecture.'64bit'.hash = $Sha256Amd64.ToLower()
$s.architecture.arm64.url = "$base/$armZip"
$s.architecture.arm64.hash = $Sha256Arm64.ToLower()
$json = $s | ConvertTo-Json -Depth 10
[IO.File]::WriteAllText($scoopPath, $json + "`n", $utf8)

# winget: rename the version folder, then rewrite the files
$wingetRoot = Join-Path $root 'packaging/winget/manifests/a/Avestura/Zagros'
$target = Join-Path $wingetRoot $Version
$existing = Get-ChildItem $wingetRoot -Directory | Select-Object -First 1
if ($existing.FullName -ne $target) {
    if (Test-Path $target) { Remove-Item $target -Recurse -Force }
    Move-Item $existing.FullName $target
}
foreach ($f in Get-ChildItem $target -Filter *.yaml) {
    $t = (Get-Content $f.FullName -Raw) -replace "`r`n", "`n"
    $t = $t -replace '(?m)^PackageVersion: .*$', "PackageVersion: $Version"
    if ($f.Name -like '*.installer.yaml') {
        $t = $t -replace '(?m)^ReleaseDate: .*$', "ReleaseDate: $ReleaseDate"
        $arch = ''
        $lines = $t -split "`n"
        for ($i = 0; $i -lt $lines.Count; $i++) {
            if ($lines[$i] -match 'Architecture: (\w+)') { $arch = $Matches[1] }
            if ($lines[$i] -match '^  InstallerUrl:') {
                $z = if ($arch -eq 'arm64') { $armZip } else { $amdZip }
                $lines[$i] = "  InstallerUrl: $base/$z"
            }
            if ($lines[$i] -match '^  InstallerSha256:') {
                $h = if ($arch -eq 'arm64') { $Sha256Arm64 } else { $Sha256Amd64 }
                $lines[$i] = "  InstallerSha256: $h"
            }
        }
        $t = $lines -join "`n"
    }
    [IO.File]::WriteAllText($f.FullName, $t, $utf8)
}
Write-Host "Updated packaging manifests to $Version"

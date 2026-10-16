<#
.SYNOPSIS
  Packs every extension in marketplace/extensions into dist/marketplace/<id>.zagros
  and copies the themes to dist/marketplace/themes/<id>.toml.
.EXAMPLE
  pwsh scripts/pack-marketplace.ps1
  pwsh scripts/pack-marketplace.ps1 -Out C:\temp\market
#>
[CmdletBinding()]
param(
    [string]$Root = (Join-Path $PSScriptRoot '..'),
    [string]$Out = ''
)
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.IO.Compression
Add-Type -AssemblyName System.IO.Compression.FileSystem

$Root = (Resolve-Path $Root).Path
if (-not $Out) { $Out = Join-Path $Root 'dist/marketplace' }
New-Item -ItemType Directory -Force -Path $Out | Out-Null

$count = 0
foreach ($dir in Get-ChildItem (Join-Path $Root 'marketplace/extensions') -Directory) {
    if (-not (Test-Path (Join-Path $dir.FullName 'manifest.toml'))) { continue }
    $dest = Join-Path $Out ($dir.Name + '.zagros')
    if (Test-Path $dest) { Remove-Item $dest }
    $zip = [System.IO.Compression.ZipFile]::Open($dest, 'Create')
    try {
        $base = $dir.FullName.TrimEnd('\', '/')
        foreach ($f in Get-ChildItem $dir.FullName -Recurse -File | Sort-Object FullName) {
            # Entry names use forward slashes with manifest.toml at the archive root.
            $name = $f.FullName.Substring($base.Length + 1).Replace('\', '/')
            [void][System.IO.Compression.ZipFileExtensions]::CreateEntryFromFile($zip, $f.FullName, $name, 'Optimal')
        }
    } finally { $zip.Dispose() }
    Write-Host "packed $dest"
    $count++
}

$themeOut = Join-Path $Out 'themes'
New-Item -ItemType Directory -Force -Path $themeOut | Out-Null
foreach ($t in Get-ChildItem (Join-Path $Root 'marketplace/themes') -Filter *.toml) {
    Copy-Item $t.FullName $themeOut -Force
}
Write-Host "$count extensions packed to $Out"

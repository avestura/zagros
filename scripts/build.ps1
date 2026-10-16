<#
.SYNOPSIS
  Builds dist/zagros.exe (windowsgui, stripped).
.PARAMETER Version
  Version string baked into the binary. Defaults to `git describe`, or "dev".
.PARAMETER Test
  Also run go vet and go test before building.
#>
param(
  [string]$Version = "",
  [switch]$Test
)
$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

if (-not $Version) {
  $ErrorActionPreference = "Continue"
  $Version = (git describe --tags --always --dirty 2>$null)
  $ErrorActionPreference = "Stop"
  if (-not $Version -or $LASTEXITCODE -ne 0) { $Version = "dev" }
}

if (-not (Test-Path "assets/zagros.ico")) { go run ./scripts/genicon }
if ($Test) { go vet ./...; go test ./... }

New-Item -ItemType Directory -Force dist | Out-Null
$env:CGO_ENABLED = "0"
$env:GOOS = "windows"
if (-not $env:GOARCH) { $env:GOARCH = "amd64" }
go build -trimpath -ldflags "-H=windowsgui -s -w -X main.version=$Version" -o dist/zagros.exe ./cmd/zagros
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
$size = [math]::Round((Get-Item dist/zagros.exe).Length / 1MB, 2)
Write-Host "dist/zagros.exe $Version ($size MB)"

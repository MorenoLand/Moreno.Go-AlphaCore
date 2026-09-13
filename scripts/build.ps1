param([string]$OutputDir = "bin")

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot
$output = [IO.Path]::GetFullPath((Join-Path $root $OutputDir))
New-Item -ItemType Directory -Force -Path $output | Out-Null
Push-Location $root
try {
    & go build -trimpath -ldflags "-s -w" -o (Join-Path $output "Moreno.AlphaCore.exe") .
    if ($LASTEXITCODE -ne 0) { throw "go build failed with exit code $LASTEXITCODE" }
} finally {
    Pop-Location
}

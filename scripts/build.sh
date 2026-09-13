#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
output="${1:-$root/bin}"
mkdir -p "$output"
go build -trimpath -ldflags "-s -w" -o "$output/Moreno.AlphaCore" "$root"

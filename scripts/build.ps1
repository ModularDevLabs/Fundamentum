$ErrorActionPreference = "Stop"
$Name = "modbot"
$OutDir = "dist"
New-Item -ItemType Directory -Force -Path $OutDir | Out-Null

Write-Host "Building local binary ./$Name for normal dashboard runs..."
$env:CGO_ENABLED = "0"
go build -o $Name ./cmd/modbot
Write-Host "Built ./$Name"

function Build($os, $arch, $ext) {
  $outfile = Join-Path $OutDir "${Name}_${os}_${arch}${ext}"
  $env:GOOS = $os
  $env:GOARCH = $arch
  $env:CGO_ENABLED = "0"
  go build -o $outfile ./cmd/modbot
  Write-Host "Built $outfile"
}

Build "windows" "amd64" ".exe"
Build "linux" "amd64" ""
Build "linux" "arm64" ""
Build "darwin" "amd64" ""
Build "darwin" "arm64" ""

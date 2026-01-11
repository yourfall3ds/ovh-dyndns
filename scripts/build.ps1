# PowerShell script to build OVH DynDNS for Windows and Linux

Write-Host "Starting compilation..." -ForegroundColor Green

# Configuration
$buildDir = "build"

# Ensure build directory exists
if (-Not (Test-Path $buildDir)) {
    New-Item -ItemType Directory -Path $buildDir | Out-Null
}

# Clean build directory
Remove-Item -Path (Join-Path $buildDir "*") -Force -ErrorAction SilentlyContinue

# Helper function for building
function Build-Target ($os, $arch, $output) {
    Write-Host "Building for $os ($arch)..." -ForegroundColor Cyan
    $env:GOOS = $os
    $env:GOARCH = $arch

    $outputPath = Join-Path $buildDir $output
    go build -o $outputPath ./cmd/ovh-dyndns

    if ($LASTEXITCODE -eq 0) {
        Write-Host "$os build: OK" -ForegroundColor Green
    } else {
        Write-Host "$os build: ERROR" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""

# Build for Windows
Build-Target "windows" "amd64" "windows_amd64.exe"

# Build for Linux
Build-Target "linux" "amd64" "linux_amd64"

# Cleanup environment variables
Remove-Item Env:\GOOS -ErrorAction SilentlyContinue
Remove-Item Env:\GOARCH -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "Compilation completed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Generated files in '$buildDir/':" -ForegroundColor Yellow
Get-ChildItem -Path $buildDir | ForEach-Object {
    $size = [math]::Round($_.Length / 1MB, 2)
    Write-Host ("  - {0} ({1} MB)" -f $_.Name, $size)
}

Write-Host ""
Write-Host "Instructions:" -ForegroundColor Cyan
$windowsPath = Join-Path "." (Join-Path $buildDir "windows_amd64.exe")
$linuxPath = Join-Path "." (Join-Path $buildDir "linux_amd64")
Write-Host "  Windows: $windowsPath"
Write-Host "  Linux:   $linuxPath"
Write-Host ""
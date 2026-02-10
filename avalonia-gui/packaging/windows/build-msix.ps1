# Windows MSIX package builder for ARM Emulator
# Requires: Windows SDK (for MakeAppx.exe)

param(
    [string]$Configuration = "Release",
    [string]$Version = "1.0.0"
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectDir = Join-Path $ScriptDir "..\..\" | Resolve-Path
$OutputDir = Join-Path $ProjectDir "dist\windows"
$AppName = "ARMEmulator"
$RID = "win-x64"

Write-Host "Building ARM Emulator for Windows..." -ForegroundColor Cyan
Write-Host "Project directory: $ProjectDir"
Write-Host "Output directory: $OutputDir"

# Clean previous builds
if (Test-Path $OutputDir) {
    Remove-Item -Recurse -Force $OutputDir
}
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

# Build the .NET application
Write-Host "Publishing .NET application..." -ForegroundColor Yellow
Set-Location $ProjectDir
dotnet publish ARMEmulator\ARMEmulator.csproj `
    --configuration $Configuration `
    --runtime $RID `
    --self-contained true `
    -p:PublishSingleFile=true `
    -p:IncludeNativeLibrariesForSelfExtract=true `
    --output "$OutputDir\app"

if ($LASTEXITCODE -ne 0) {
    Write-Error "Build failed!"
    exit 1
}

# Copy backend binary
$BackendSource = Join-Path $ProjectDir "..\arm-emulator.exe"
if (Test-Path $BackendSource) {
    Write-Host "Copying backend binary..." -ForegroundColor Yellow
    Copy-Item $BackendSource -Destination "$OutputDir\app\arm-emulator.exe"
} else {
    Write-Warning "Backend binary not found at $BackendSource"
    Write-Warning "The app will not be able to start the backend automatically."
}

# Create MSIX manifest
$ManifestDir = Join-Path $OutputDir "manifest"
New-Item -ItemType Directory -Force -Path $ManifestDir | Out-Null

$Manifest = @"
<?xml version="1.0" encoding="utf-8"?>
<Package xmlns="http://schemas.microsoft.com/appx/manifest/foundation/windows10"
         xmlns:uap="http://schemas.microsoft.com/appx/manifest/uap/windows10">
  <Identity Name="ARMEmulator"
            Publisher="CN=ARMEmulator"
            Version="$Version.0" />
  <Properties>
    <DisplayName>ARM Emulator</DisplayName>
    <PublisherDisplayName>ARM Emulator Project</PublisherDisplayName>
  </Properties>
  <Dependencies>
    <TargetDeviceFamily Name="Windows.Desktop" MinVersion="10.0.19041.0" MaxVersionTested="10.0.22621.0" />
  </Dependencies>
  <Resources>
    <Resource Language="en-us" />
  </Resources>
  <Applications>
    <Application Id="ARMEmulator" Executable="ARMEmulator.exe" EntryPoint="Windows.FullTrustApplication">
      <uap:VisualElements DisplayName="ARM Emulator"
                         Description="ARM Assembly Language Emulator"
                         BackgroundColor="transparent"
                         Square150x150Logo="Assets\Square150x150Logo.png"
                         Square44x44Logo="Assets\Square44x44Logo.png">
      </uap:VisualElements>
    </Application>
  </Applications>
</Package>
"@

Set-Content -Path (Join-Path $ManifestDir "AppxManifest.xml") -Value $Manifest

Write-Host "MSIX package structure created" -ForegroundColor Green
Write-Host ""
Write-Host "Build complete!" -ForegroundColor Green
Write-Host "Application files: $OutputDir\app"
Write-Host ""
Write-Host "To create an MSIX package, run MakeAppx.exe from the Windows SDK:"
Write-Host "  MakeAppx.exe pack /d '$OutputDir\app' /p '$OutputDir\ARMEmulator-$Version-$RID.msix'"
Write-Host ""
Write-Host "Note: You will need to sign the MSIX with a certificate to install it."

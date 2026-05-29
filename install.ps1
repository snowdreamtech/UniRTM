<#
.SYNOPSIS
    UniRTM installer for Windows (PowerShell).

.DESCRIPTION
    Downloads and installs the unirtm binary from GitHub Releases.

.PARAMETER Version
    Target version to install (e.g. "v0.0.10"). Defaults to latest.

.PARAMETER InstallDir
    Directory to install the binary. Defaults to "$HOME\.unirtm\bin".

.PARAMETER NoProxy
    Disable the GitHub proxy.

.EXAMPLE
    # Install latest
    irm https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/install.ps1 | iex

    # Install specific version
    $env:UNIRTM_VERSION="v0.0.10"; irm https://raw.githubusercontent.com/snowdreamtech/UniRTM/main/install.ps1 | iex
#>
[CmdletBinding()]
param(
    [string]$Version     = $env:UNIRTM_VERSION,
    [string]$InstallDir  = "",
    [string]$GithubProxy = $env:GITHUB_PROXY,
    [switch]$NoProxy
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# ---------------------------------------------------------------------------
# Configuration
# ---------------------------------------------------------------------------
$Repo   = "snowdreamtech/UniRTM"
$Binary = "unirtm"

if (-not $GithubProxy) {
    $GithubProxy = "https://gh-proxy.sn0wdr1am.com/"
}
if ($NoProxy) {
    $GithubProxy = ""
}

# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------
function Write-Info  { param($Msg) Write-Host "[INFO]  $Msg" -ForegroundColor Green }
function Write-Warn  { param($Msg) Write-Host "[WARN]  $Msg" -ForegroundColor Yellow }
function Write-Err   { param($Msg) Write-Host "[ERROR] $Msg" -ForegroundColor Red }
function Abort       { param($Msg) Write-Err $Msg; exit 1 }

function Apply-Proxy {
    param([string]$Url)
    if ($GithubProxy -and ($Url -match "^https://github\.com/" -or $Url -match "^https://objects\.githubusercontent\.com/" -or $Url -match "^https://raw\.githubusercontent\.com/")) {
        return "${GithubProxy}${Url}"
    }
    return $Url
}

function Invoke-DownloadWithRetry {
    param(
        [string]$Url,
        [string]$OutFile,
        [int]$MaxRetries = 5,
        [int]$RetryDelay = 3
    )

    $ProxiedUrl = Apply-Proxy $Url
    $Urls = @($ProxiedUrl)
    if ($ProxiedUrl -ne $Url) { $Urls += $Url }

    foreach ($TargetUrl in $Urls) {
        $attempt = 0
        while ($attempt -lt $MaxRetries) {
            try {
                $attempt++
                Write-Info "Downloading from: $TargetUrl (attempt $attempt/$MaxRetries)"
                $webClient = New-Object System.Net.WebClient
                $webClient.Headers.Add("User-Agent", "unirtm-installer/1.0")
                if ($OutFile) {
                    $webClient.DownloadFile($TargetUrl, $OutFile)
                    if ((Get-Item $OutFile).Length -gt 0) { return }
                    throw "Downloaded file is empty"
                } else {
                    return $webClient.DownloadString($TargetUrl)
                }
            } catch {
                Write-Warn "Attempt $attempt failed: $_"
                if ($attempt -lt $MaxRetries) {
                    $sleep = $RetryDelay * $attempt
                    Write-Info "Retrying in ${sleep}s..."
                    Start-Sleep -Seconds $sleep
                }
            }
        }
        if ($TargetUrl -eq $ProxiedUrl -and $ProxiedUrl -ne $Url) {
            Write-Warn "Proxy downloads failed, switching to direct download..."
        } else {
            Abort "All download attempts failed for: $Url"
        }
    }
}

# ---------------------------------------------------------------------------
# Detect Architecture
# ---------------------------------------------------------------------------
function Detect-Arch {
    $arch = $env:PROCESSOR_ARCHITECTURE
    # Also check PROCESSOR_ARCHITEW6432 for WOW64
    if ($env:PROCESSOR_ARCHITEW6432) { $arch = $env:PROCESSOR_ARCHITEW6432 }
    switch -Wildcard ($arch) {
        "AMD64"  { return "x86_64" }
        "ARM64"  { return "arm64" }
        "x86"    { return "i386" }
        default  { Abort "Unsupported architecture: $arch" }
    }
}

# ---------------------------------------------------------------------------
# Resolve version
# ---------------------------------------------------------------------------
function Resolve-Version {
    if ($Version) {
        if (-not $Version.StartsWith("v")) { $Version = "v$Version" }
        Write-Info "Target version: $Version"
        return $Version
    }

    Write-Info "Fetching latest release from GitHub API..."
    $ApiUrl = "https://api.github.com/repos/$Repo/releases/latest"
    $ProxiedApiUrl = Apply-Proxy $ApiUrl

    $urls = @($ProxiedApiUrl)
    if ($ProxiedApiUrl -ne $ApiUrl) { $urls += $ApiUrl }

    foreach ($url in $urls) {
        try {
            $wc = New-Object System.Net.WebClient
            $wc.Headers.Add("User-Agent", "unirtm-installer/1.0")
            $json = $wc.DownloadString($url)
            if ($json -match '"tag_name"\s*:\s*"([^"]+)"') {
                $tag = $Matches[1]
                Write-Info "Latest version: $tag"
                return $tag
            }
        } catch {
            Write-Warn "Failed to fetch from ${url}: $_"
        }
    }
    Abort "Could not determine latest version. Please use -Version to specify one."
}

# ---------------------------------------------------------------------------
# Download and verify
# ---------------------------------------------------------------------------
function Download-And-Verify {
    param([string]$ResolvedVersion, [string]$Arch)

    $ArchiveName  = "${Binary}_Windows_${Arch}.zip"
    $ArchiveUrl   = "https://github.com/${Repo}/releases/download/${ResolvedVersion}/${ArchiveName}"
    $ChecksumUrl  = "https://github.com/${Repo}/releases/download/${ResolvedVersion}/checksums.txt"

    $TmpDir = Join-Path $env:TEMP "unirtm_install_$([System.IO.Path]::GetRandomFileName())"
    New-Item -ItemType Directory -Path $TmpDir | Out-Null

    $ArchivePath  = Join-Path $TmpDir $ArchiveName
    $ChecksumPath = Join-Path $TmpDir "checksums.txt"

    Write-Info "Downloading $ArchiveName..."
    Invoke-DownloadWithRetry -Url $ArchiveUrl -OutFile $ArchivePath

    if (-not (Test-Path $ArchivePath) -or (Get-Item $ArchivePath).Length -eq 0) {
        Abort "Downloaded archive is empty or missing: $ArchivePath"
    }

    # Try to download and verify checksum
    try {
        Invoke-DownloadWithRetry -Url $ChecksumUrl -OutFile $ChecksumPath
        $checksumContent = Get-Content $ChecksumPath -Raw
        $entry = ($checksumContent -split "`r?`n") | Where-Object { $_ -match ('\s' + [regex]::Escape($ArchiveName) + '$') }
        if ($entry) {
            $expected = ($entry -split '\s+')[0].Trim()
            $actual   = (Get-FileHash -Algorithm SHA256 $ArchivePath).Hash.ToLower()
            if ($expected -ne $actual) {
                Abort "Checksum mismatch! Expected: $expected, Got: $actual"
            }
            Write-Info "Checksum verified OK."
        } else {
            Write-Warn "Checksum entry not found for $ArchiveName, skipping verification."
        }
    } catch {
        Write-Warn "Could not verify checksum: $_"
    }

    return @{ TmpDir = $TmpDir; ArchivePath = $ArchivePath }
}

# ---------------------------------------------------------------------------
# Install binary
# ---------------------------------------------------------------------------
function Install-Binary {
    param([hashtable]$Download, [string]$InstallDirPath)

    $TmpDir      = $Download.TmpDir
    $ArchivePath = $Download.ArchivePath

    Write-Info "Extracting archive..."
    Expand-Archive -Path $ArchivePath -DestinationPath $TmpDir -Force

    # Find the binary
    $BinaryFile = Get-ChildItem -Path $TmpDir -Recurse -Filter "${Binary}.exe" | Select-Object -First 1
    if (-not $BinaryFile) {
        Abort "Binary '${Binary}.exe' not found in archive."
    }

    # Ensure install dir exists
    if (-not (Test-Path $InstallDirPath)) {
        New-Item -ItemType Directory -Path $InstallDirPath | Out-Null
    }

    $Destination = Join-Path $InstallDirPath "${Binary}.exe"
    Copy-Item -Path $BinaryFile.FullName -Destination $Destination -Force
    Write-Info "Installed ${Binary}.exe to $Destination"

    # Cleanup
    Remove-Item -Recurse -Force $TmpDir -ErrorAction SilentlyContinue

    return $Destination
}

# ---------------------------------------------------------------------------
# Update user PATH
# ---------------------------------------------------------------------------
function Update-UserPath {
    param([string]$DirToAdd)

    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -notlike "*${DirToAdd}*") {
        [Environment]::SetEnvironmentVariable("PATH", "${currentPath};${DirToAdd}", "User")
        $env:PATH = "${env:PATH};${DirToAdd}"
        Write-Info "Added $DirToAdd to your user PATH."
        Write-Warn "Restart your terminal for PATH changes to take effect."
    } else {
        Write-Info "$DirToAdd is already in PATH."
    }
}

# ---------------------------------------------------------------------------
# Post-install verification
# ---------------------------------------------------------------------------
function Verify-Install {
    param([string]$BinaryPath)

    if (-not (Test-Path $BinaryPath)) {
        Abort "Verification failed: binary not found at $BinaryPath"
    }

    try {
        $verOutput = & $BinaryPath version 2>&1 | Select-Object -First 1
        Write-Info "Installed version: $verOutput"
    } catch {
        Write-Warn "Could not verify installed version: $_"
    }

    Write-Info "Installation complete!"
}

# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
function Main {
    $arch = Detect-Arch
    Write-Info "Detected architecture: $arch"

    $resolvedVersion = Resolve-Version

    if (-not $InstallDir) {
        $InstallDir = Join-Path $HOME ".unirtm\bin"
    }

    $download    = Download-And-Verify -ResolvedVersion $resolvedVersion -Arch $arch
    $binaryPath  = Install-Binary -Download $download -InstallDirPath $InstallDir

    Update-UserPath -DirToAdd $InstallDir
    Verify-Install -BinaryPath $binaryPath
}

Main

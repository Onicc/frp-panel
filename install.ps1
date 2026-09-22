#Requires -Version 5.1
[CmdletBinding()]
param(
    [string]$Version = "latest",
    [string]$GitHubProxy = "",
    [Parameter(ValueFromRemainingArguments = $true)]
    [string[]]$AgentArguments
)

$ErrorActionPreference = "Stop"
if ($Version -ne "latest" -and $Version -notmatch '^v[0-9]+\.[0-9]+\.[0-9]+$') {
    throw "Version must be latest or a stable vX.X.X release; edge is retired"
}
$repository = "Onicc/frp-panel"
$architecture = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
switch ($architecture) {
    "x64" { $assetArchitecture = "amd64" }
    "arm64" { $assetArchitecture = "arm64" }
    default { throw "Unsupported Windows architecture: $architecture" }
}

$asset = "frp-panel-agent-windows-$assetArchitecture.exe"
if ($Version -eq "latest") {
    $releaseUrl = "https://github.com/$repository/releases/latest/download"
} else {
    $releaseUrl = "https://github.com/$repository/releases/download/$Version"
}
if ($GitHubProxy) {
    $releaseUrl = $GitHubProxy.TrimEnd('/') + "/" + $releaseUrl
}

$tempDirectory = Join-Path ([System.IO.Path]::GetTempPath()) ("frp-panel-" + [Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tempDirectory | Out-Null
try {
    $binary = Join-Path $tempDirectory $asset
    $checksum = Join-Path $tempDirectory "checksums.txt"
    Invoke-WebRequest -UseBasicParsing -Uri "$releaseUrl/$asset" -OutFile $binary
    Invoke-WebRequest -UseBasicParsing -Uri "$releaseUrl/checksums.txt" -OutFile $checksum
    $checksumLine = Get-Content $checksum | Where-Object { $_ -match "\s+$([Regex]::Escape($asset))$" } | Select-Object -First 1
    if (-not $checksumLine) { throw "Asset is missing from checksums.txt" }
    $expected = (($checksumLine.Trim()) -split '\s+')[0].ToLowerInvariant()
    $actual = (Get-FileHash -Algorithm SHA256 -Path $binary).Hash.ToLowerInvariant()
    if ($actual -ne $expected) {
        throw "Checksum verification failed"
    }
    & $binary service install @AgentArguments
    if ($LASTEXITCODE -ne 0) {
        throw "Agent service installation failed with exit code $LASTEXITCODE"
    }
    Write-Host "frp-panel-agent installed successfully; no files were written to $((Get-Location).Path)."
} finally {
    Remove-Item -LiteralPath $tempDirectory -Recurse -Force -ErrorAction SilentlyContinue
}

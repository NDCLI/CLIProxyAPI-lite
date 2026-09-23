$ErrorActionPreference = 'Stop'

$sourceDir = $PSScriptRoot
$installDir = Join-Path $env:LOCALAPPDATA 'CLIProxyAPI'
$requiredFiles = @('cli-proxy-api.exe', 'cliproxy.exe', 'config.example.yaml')

foreach ($name in $requiredFiles) {
    if (-not (Test-Path -LiteralPath (Join-Path $sourceDir $name) -PathType Leaf)) {
        throw "Required release file is missing: $name"
    }
}

New-Item -ItemType Directory -Path $installDir -Force | Out-Null
foreach ($name in $requiredFiles) {
    $sourcePath = [IO.Path]::GetFullPath((Join-Path $sourceDir $name))
    $targetPath = [IO.Path]::GetFullPath((Join-Path $installDir $name))
    if (-not [string]::Equals($sourcePath, $targetPath, [StringComparison]::OrdinalIgnoreCase)) {
        Copy-Item -LiteralPath $sourcePath -Destination $targetPath -Force
    }
}

$configPath = Join-Path $installDir 'config.yaml'
if (-not (Test-Path -LiteralPath $configPath -PathType Leaf)) {
    Copy-Item -LiteralPath (Join-Path $installDir 'config.example.yaml') -Destination $configPath
}

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
$pathEntries = @($userPath -split ';' | Where-Object { $_.Trim() })
$alreadyOnPath = $pathEntries | Where-Object {
    [string]::Equals($_.Trim().TrimEnd('\'), $installDir.TrimEnd('\'), [StringComparison]::OrdinalIgnoreCase)
}
if (-not $alreadyOnPath) {
    $updatedPath = (@($pathEntries) + $installDir) -join ';'
    [Environment]::SetEnvironmentVariable('Path', $updatedPath, 'User')
}

Write-Host "Installed to $installDir"
Write-Host 'Close and reopen CMD or PowerShell, then run: cliproxy'
Write-Host "Edit configuration: $configPath"

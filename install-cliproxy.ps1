$ErrorActionPreference = 'Stop'

$sourceDir = $PSScriptRoot
$installDir = Join-Path $env:LOCALAPPDATA 'Lumina'
$legacyDir = Join-Path $env:LOCALAPPDATA 'CLIProxyAPI'
$requiredFiles = @('cli-proxy-api.exe', 'lumina.exe', 'config.example.yaml')

foreach ($name in $requiredFiles) {
    if (-not (Test-Path -LiteralPath (Join-Path $sourceDir $name) -PathType Leaf)) {
        throw "Required release file is missing: $name"
    }
}

New-Item -ItemType Directory -Path $installDir -Force | Out-Null

if ((Test-Path -LiteralPath $legacyDir -PathType Container) -and -not (Test-Path -LiteralPath (Join-Path $installDir 'config.yaml') -PathType Leaf)) {
    foreach ($name in @('config.yaml', '.env')) {
        $from = Join-Path $legacyDir $name
        $to = Join-Path $installDir $name
        if ((Test-Path -LiteralPath $from) -and -not (Test-Path -LiteralPath $to)) {
            Copy-Item -LiteralPath $from -Destination $to -Force
        }
    }
    foreach ($name in @('auths', 'certs')) {
        $from = Join-Path $legacyDir $name
        $to = Join-Path $installDir $name
        if ((Test-Path -LiteralPath $from -PathType Container) -and -not (Test-Path -LiteralPath $to)) {
            Copy-Item -LiteralPath $from -Destination $to -Recurse -Force
        }
    }
}

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
$normalizedInstall = $installDir.TrimEnd('\')
$normalizedLegacy = $legacyDir.TrimEnd('\')
$pathEntries = @($pathEntries | Where-Object {
    -not [string]::Equals($_.Trim().TrimEnd('\'), $normalizedLegacy, [StringComparison]::OrdinalIgnoreCase)
})
$alreadyOnPath = $pathEntries | Where-Object {
    [string]::Equals($_.Trim().TrimEnd('\'), $normalizedInstall, [StringComparison]::OrdinalIgnoreCase)
}
if (-not $alreadyOnPath) {
    $pathEntries = @($normalizedInstall) + @($pathEntries)
}
[Environment]::SetEnvironmentVariable('Path', ($pathEntries -join ';'), 'User')

Write-Host "Installed Lumina to $installDir"
Write-Host 'Close and reopen CMD or PowerShell, then run: lumina'
Write-Host "Edit configuration: $configPath"

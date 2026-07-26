[CmdletBinding()]
param()

$ErrorActionPreference = "Continue"
$RepoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path

function Get-ToolStatus {
    param(
        [Parameter(Mandatory = $true)][string]$Name,
        [Parameter(Mandatory = $true)][string[]]$VersionArguments
    )

    $Command = Get-Command $Name -ErrorAction SilentlyContinue
    if (-not $Command) {
        return [pscustomobject]@{ Tool = $Name; Available = $false; Location = ""; Version = "" }
    }

    $Version = (& $Name @VersionArguments 2>&1 | Select-Object -First 1 | Out-String).Trim()
    return [pscustomobject]@{
        Tool = $Name
        Available = $true
        Location = $Command.Source
        Version = $Version
    }
}

Write-Host "ClipSense repository doctor"
Write-Host "Repository: $RepoRoot"
Write-Host "Operating system: $([System.Environment]::OSVersion.VersionString)"
Write-Host "Architecture: $env:PROCESSOR_ARCHITECTURE"

if ($RepoRoot -match "[\\/]OneDrive[\\/]") {
    Write-Warning "Repository is inside OneDrive. Builds, watchers, and large media operations may contend with synchronization."
}

$Tools = @(
    Get-ToolStatus -Name "git" -VersionArguments @("--version")
    Get-ToolStatus -Name "docker" -VersionArguments @("--version")
    Get-ToolStatus -Name "node" -VersionArguments @("--version")
    Get-ToolStatus -Name "npm" -VersionArguments @("--version")
    Get-ToolStatus -Name "go" -VersionArguments @("version")
    Get-ToolStatus -Name "python" -VersionArguments @("--version")
    Get-ToolStatus -Name "ffmpeg" -VersionArguments @("-version")
    Get-ToolStatus -Name "pwsh" -VersionArguments @("--version")
    Get-ToolStatus -Name "gh" -VersionArguments @("--version")
)

$Tools | Format-Table -AutoSize

$Docker = $Tools | Where-Object Tool -eq "docker"
if ($Docker.Available) {
    & docker compose version
} else {
    Write-Warning "Docker Compose not checked because docker is unavailable on PATH."
}

$EnvironmentPaths = @(
    (Join-Path $RepoRoot ".env"),
    (Join-Path $RepoRoot "apps/web/.env.local")
)
foreach ($Path in $EnvironmentPaths) {
    Write-Host ("Environment file {0}: {1}" -f $Path, (Test-Path -LiteralPath $Path))
}

Write-Host "Required local ports:"
$Ports = 3000, 5432, 6333, 6379, 8080
foreach ($Port in $Ports) {
    $Listener = Get-NetTCPConnection -State Listen -LocalPort $Port -ErrorAction SilentlyContinue
    if ($Listener) {
        Write-Warning "Port $Port is already in use."
    } else {
        Write-Host "Port $Port appears available."
    }
}

Write-Host "Doctor completed. No tools or files were installed or modified."


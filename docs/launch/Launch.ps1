#!/usr/bin/env pwsh
#Requires -Version 5.1

<#
.SYNOPSIS
    Cross-platform bootstrapper that installs prerequisites (if available) and starts every backend service plus the Vite frontend.

.DESCRIPTION
    This replaces the original Windows-only helper. It works on PowerShell 7+ (pwsh) for both Windows and Linux/Mac.
    The script will:
      1. Ensure Go, Node.js/npm, and a C toolchain are present (auto-install when possible).
      2. Create the SQLite database file if it does not exist.
      3. Download Go modules and install frontend dependencies.
      4. Start Auth, Users, Posts, Groups, Chat, and Notifications microservices.
      5. Start the Vite dev server so the Vue frontend can talk to the backend.

    Use Ctrl+C to stop everything. View logs with Receive-Job -Name "<service name>" -Keep.
#>
[CmdletBinding()]
param(
    [switch]$SkipInstall,
    [switch]$SkipFrontend,
    [string]$FrontendHost = "localhost",
    [int]$FrontendPort = 5173
)

$runtimeInfo = [System.Runtime.InteropServices.RuntimeInformation]
if (-not (Test-Path variable:IsWindows)) {
    Set-Variable -Scope Script -Name IsWindows -Value ($runtimeInfo::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Windows))
}
if (-not (Test-Path variable:IsLinux)) {
    Set-Variable -Scope Script -Name IsLinux -Value ($runtimeInfo::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::Linux))
}
if (-not (Test-Path variable:IsMacOS)) {
    Set-Variable -Scope Script -Name IsMacOS -Value ($runtimeInfo::IsOSPlatform([System.Runtime.InteropServices.OSPlatform]::OSX))
}

$ErrorActionPreference = 'Stop'

$script:RepoRoot = $PSScriptRoot
$script:DatabasePath = Join-Path $RepoRoot "social_network.db"
$script:FrontendDir = Join-Path $RepoRoot "frontend"
$script:AuthServiceUrl = "http://localhost:8081"
$script:ServiceJobs = @()

function Write-Section {
    param([string]$Text)
    Write-Host ""
    Write-Host "=== $Text ===" -ForegroundColor Magenta
}

function Write-Info {
    param([string]$Text)
    Write-Host "  - $Text" -ForegroundColor Gray
}

function Test-IsRoot {
    if ($IsWindows) { return $true }
    try {
        $uid = & id -u 2>$null
        return ([int]$uid -eq 0)
    } catch {
        return $false
    }
}

function Invoke-CommandElevated {
    param(
        [Parameter(Mandatory)] [string]$Executable,
        [Parameter(Mandatory)] [string[]]$Arguments
    )

    if (($IsLinux -or $IsMacOS) -and -not (Test-IsRoot)) {
        if (-not (Get-Command sudo -ErrorAction SilentlyContinue)) {
            throw "sudo is required to install $Executable but was not found. Re-run as root or install sudo."
        }
        & sudo $Executable @Arguments
    } else {
        & $Executable @Arguments
    }
}

function Ensure-SystemTool {
    param(
        [Parameter(Mandatory)] [string]$Command,
        [Parameter(Mandatory)] [string]$FriendlyName,
        [Parameter(Mandatory)] [hashtable]$Packages
    )

    if (Get-Command $Command -ErrorAction SilentlyContinue) {
        Write-Info "$FriendlyName already installed."
        return
    }

    if ($SkipInstall) {
        throw "$FriendlyName is required but missing. Install it manually or rerun without -SkipInstall."
    }

    Write-Info "$FriendlyName not detected. Attempting automatic install..."
    $installed = $false

    if ($IsWindows) {
        if ($Packages.WindowsWinget -and (Get-Command winget -ErrorAction SilentlyContinue)) {
            $args = @(
                "install", "-e", "--id", $Packages.WindowsWinget,
                "--accept-package-agreements", "--accept-source-agreements", "--silent"
            )
            Start-Process -FilePath "winget" -ArgumentList $args -Wait
            $installed = $true
        } elseif ($Packages.WindowsChoco -and (Get-Command choco -ErrorAction SilentlyContinue)) {
            Start-Process -FilePath "choco" -ArgumentList @("install", $Packages.WindowsChoco, "-y") -Wait
            $installed = $true
        }
    } elseif ($IsLinux) {
        if ($Packages.LinuxApt -and (Get-Command apt-get -ErrorAction SilentlyContinue)) {
            Invoke-CommandElevated -Executable "apt-get" -Arguments @("update")
            $pkg = $Packages.LinuxApt -split '\s+'
            Invoke-CommandElevated -Executable "apt-get" -Arguments @("install", "-y") + $pkg
            $installed = $true
        } elseif ($Packages.LinuxDnf -and (Get-Command dnf -ErrorAction SilentlyContinue)) {
            $pkg = $Packages.LinuxDnf -split '\s+'
            Invoke-CommandElevated -Executable "dnf" -Arguments @("install", "-y") + $pkg
            $installed = $true
        } elseif ($Packages.LinuxPacman -and (Get-Command pacman -ErrorAction SilentlyContinue)) {
            $pkg = $Packages.LinuxPacman -split '\s+'
            Invoke-CommandElevated -Executable "pacman" -Arguments @("-S", "--noconfirm") + $pkg
            $installed = $true
        }
    } elseif ($IsMacOS) {
        if ($Packages.MacBrew -and (Get-Command brew -ErrorAction SilentlyContinue)) {
            & brew install $Packages.MacBrew
            $installed = $true
        }
    }

    if (-not $installed) {
        throw "Unable to auto-install $FriendlyName. Install it manually and rerun the script."
    }

    if (-not (Get-Command $Command -ErrorAction SilentlyContinue)) {
        throw "$FriendlyName installation finished but the command '$Command' is still unavailable."
    }
}

function Ensure-DatabaseFile {
    if (-not (Test-Path $script:DatabasePath)) {
        Write-Info "Creating SQLite database at $($script:DatabasePath)"
        New-Item -Path $script:DatabasePath -ItemType File -Force | Out-Null
    } else {
        Write-Info "Using existing database at $($script:DatabasePath)"
    }
}

function Ensure-GoDependencies {
  Write-Section "Downloading Go modules"
  Push-Location $script:RepoRoot
  try {
    go mod download
  } finally {
    Pop-Location
  }
}

function Ensure-RootNodeDependencies {
  Write-Section "Installing shared Node.js dependencies"
  Push-Location $script:RepoRoot
  try {
    if (Test-Path package-lock.json) {
      npm install --no-fund --no-audit | Out-Null
    } elseif (Test-Path package.json) {
      npm install --no-fund --no-audit | Out-Null
    }
  } finally {
    Pop-Location
  }
}

function Ensure-FrontendDependencies {
  if ($SkipFrontend) { return }
  Write-Section "Installing frontend dependencies"
  Push-Location $script:FrontendDir
  try {
        npm install --no-fund --no-audit
    } finally {
        Pop-Location
    }
}

function Start-GoService {
    param(
        [Parameter(Mandatory)] [hashtable]$Service
    )

    $servicePath = Join-Path $script:RepoRoot $Service.Path
    if (-not (Test-Path $servicePath)) {
        Write-Info "Skipping $($Service.Name) because $servicePath was not found."
        return
    }

    Write-Host ("++ Starting {0} (:{1})" -f $Service.Name, $Service.Port) -ForegroundColor Cyan
    $envMap = @{}
    if ($Service.Env) {
        foreach ($key in $Service.Env.Keys) {
            $envMap[$key] = $Service.Env[$key]
        }
    }

    $job = Start-Job -Name $Service.Name -ScriptBlock {
        param($WorkingDir, $DbPath, $EnvMap)
        Set-Location $WorkingDir
        $env:DATABASE_PATH = $DbPath
        foreach ($entry in $EnvMap.GetEnumerator()) {
            if ($entry.Key -and $entry.Value) {
                Set-Item -Path ("env:{0}" -f $entry.Key) -Value $entry.Value
            }
        }
        go run main.go
    } -ArgumentList $servicePath, $script:DatabasePath, $envMap

    $script:ServiceJobs += $job
    Start-Sleep -Milliseconds 300
}

function Start-Frontend {
    if ($SkipFrontend) {
        Write-Info "Skipping frontend (requested)."
        return
    }

    Write-Section "Starting frontend (Vite)"
    Write-Info ("Dev server will be available at http://{0}:{1}" -f $FrontendHost, $FrontendPort)
    $job = Start-Job -Name "Frontend" -ScriptBlock {
        param($WorkingDir, $HostName, $Port)
        Set-Location $WorkingDir
        npm run dev -- --host $HostName --port $Port
    } -ArgumentList $script:FrontendDir, $FrontendHost, $FrontendPort

    $script:ServiceJobs += $job
}

function Stop-ServiceJobs {
    if (-not $script:ServiceJobs) { return }
    Write-Host ""
    Write-Host "Stopping services..." -ForegroundColor Yellow
    foreach ($job in $script:ServiceJobs) {
        try {
            if ($job.State -eq 'Running') {
                Stop-Job -Job $job -Force
                Wait-Job -Job $job -ErrorAction SilentlyContinue | Out-Null
            }
        } catch {
            Write-Verbose $_
        } finally {
            Remove-Job -Job $job -Force -ErrorAction SilentlyContinue
        }
    }
}

function Show-ServiceSummary {
    Write-Section "Services running"
    $services = @(
        @{ Name = "Auth"; URL = "http://localhost:8081"; Description = "Login & tokens" },
        @{ Name = "Users"; URL = "http://localhost:8082"; Description = "Profiles & follows" },
        @{ Name = "Posts"; URL = "http://localhost:8083"; Description = "Posts & comments" },
        @{ Name = "Groups"; URL = "http://localhost:8084"; Description = "Groups & events" },
        @{ Name = "Chat"; URL = "http://localhost:8085"; Description = "Realtime chat" },
        @{ Name = "Notifications"; URL = "http://localhost:8086"; Description = "Realtime notifications" }
    )

    foreach ($svc in $services) {
        Write-Info ("{0,-15} {1} ({2})" -f ($svc.Name + ":"), $svc.URL, $svc.Description)
    }

    if (-not $SkipFrontend) {
        Write-Info ("Frontend:      http://{0}:{1} (Vite dev server)" -f $FrontendHost, $FrontendPort)
    }

    Write-Host ""
  Write-Host "Use 'Get-Job' to inspect running jobs." -ForegroundColor DarkGray
  Write-Host "Stream logs via: Receive-Job -Name 'Auth Service' -Keep" -ForegroundColor DarkGray
  Write-Host "Press Ctrl+C to stop all services." -ForegroundColor Yellow
}

function Open-FrontendBrowser {
  if ($SkipFrontend) { return }
  try {
    $url = "http://$FrontendHost`:$FrontendPort"
    Write-Host "Opening browser at $url" -ForegroundColor Cyan
    Start-Process $url | Out-Null
  } catch {
    Write-Host "Failed to auto-open browser. Please open http://$FrontendHost`:$FrontendPort manually." -ForegroundColor Yellow
  }
}

Write-Section "Checking prerequisites"
Ensure-SystemTool -Command "go" -FriendlyName "Go" -Packages @{
    WindowsWinget = "GoLang.Go"
    WindowsChoco  = "golang"
    LinuxApt      = "golang-go"
    LinuxDnf      = "golang"
    LinuxPacman   = "go"
    MacBrew       = "go"
}
Ensure-SystemTool -Command "node" -FriendlyName "Node.js (LTS)" -Packages @{
    WindowsWinget = "OpenJS.NodeJS.LTS"
    WindowsChoco  = "nodejs-lts"
    LinuxApt      = "nodejs npm"
    LinuxDnf      = "nodejs npm"
    LinuxPacman   = "nodejs npm"
    MacBrew       = "node"
}
Ensure-SystemTool -Command "npm" -FriendlyName "npm" -Packages @{
    WindowsWinget = "OpenJS.NodeJS.LTS"
    WindowsChoco  = "nodejs-lts"
    LinuxApt      = "nodejs npm"
    LinuxDnf      = "nodejs npm"
    LinuxPacman   = "nodejs npm"
    MacBrew       = "node"
}
Ensure-SystemTool -Command "gcc" -FriendlyName "C/C++ toolchain (required for go-sqlite3)" -Packages @{
    WindowsChoco = "mingw"
    LinuxApt     = "build-essential"
    LinuxDnf     = "gcc"
    LinuxPacman  = "base-devel"
    MacBrew      = "gcc"
}
Ensure-DatabaseFile
Ensure-GoDependencies
Ensure-RootNodeDependencies
Ensure-FrontendDependencies

$services = @(
    @{ Name = "Auth Service";          Port = 8081; Path = "services/auth";           Env = @{} },
    @{ Name = "Users Service";         Port = 8082; Path = "services/users";          Env = @{ AUTH_SERVICE_URL = $script:AuthServiceUrl } },
    @{ Name = "Posts Service";         Port = 8083; Path = "services/posts";          Env = @{ AUTH_SERVICE_URL = $script:AuthServiceUrl; PORT = "8083" } },
    @{ Name = "Groups Service";        Port = 8084; Path = "services/groups";         Env = @{ AUTH_SERVICE_URL = $script:AuthServiceUrl } },
    @{ Name = "Chat Service";          Port = 8085; Path = "services/chat";           Env = @{ AUTH_SERVICE_URL = $script:AuthServiceUrl } },
    @{ Name = "Notifications Service"; Port = 8086; Path = "services/notifications";  Env = @{ AUTH_SERVICE_URL = $script:AuthServiceUrl } }
)

Write-Section "Starting backend services"
foreach ($svc in $services) {
    Start-GoService -Service $svc
}

Start-Frontend
Show-ServiceSummary
Open-FrontendBrowser

try {
    while ($true) {
        Start-Sleep -Seconds 1
    }
} finally {
    Stop-ServiceJobs
}


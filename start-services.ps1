# Start all backend services in separate PowerShell windows (cross-platform)
# This script resolves paths relative to the repo so it works on any machine

$repoRoot = $PSScriptRoot
$dbPath = Join-Path $repoRoot "social_network.db"
$authServiceUrl = "http://localhost:8081"

if (-not (Test-Path $dbPath)) {
    Write-Error "Database file not found at '$dbPath'. Make sure migrations have been run or adjust the path."
    exit 1
}

$shellExecutable = if (Get-Command pwsh -ErrorAction SilentlyContinue) {
    "pwsh"
} elseif (Get-Command powershell -ErrorAction SilentlyContinue) {
    "powershell"
} else {
    Write-Error "Neither 'pwsh' nor 'powershell' was found in PATH."
    exit 1
}

if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go compiler not found. Install Go and ensure 'go' is available in PATH."
    exit 1
}

# Service configurations
$services = @(
    @{
        Name = "Auth Service"
        Port = 8081
        Path = "services\auth"
        NeedsAuthUrl = $false
    },
    @{
        Name = "Users Service"
        Port = 8082
        Path = "services\users"
        NeedsAuthUrl = $true
    },
    @{
        Name = "Posts Service"
        Port = 8083
        Path = "services\posts"
        NeedsAuthUrl = $true
    },
    @{
        Name = "Groups Service"
        Port = 8084
        Path = "services\groups"
        NeedsAuthUrl = $true
    },
    @{
        Name = "Chat Service"
        Port = 8085
        Path = "services\chat"
        NeedsAuthUrl = $true
    },
    @{
        Name = "Notifications Service"
        Port = 8086
        Path = "services\notifications"
        NeedsAuthUrl = $true
    }
)

Write-Host "Starting all backend services..." -ForegroundColor Green
Write-Host ""

foreach ($service in $services) {
    Write-Host "Starting $($service.Name) on port $($service.Port)..." -ForegroundColor Cyan
    
    $servicePath = Join-Path $repoRoot $service.Path
    if (-not (Test-Path $servicePath)) {
        Write-Warning "  Skipping $($service.Name) - path '$servicePath' not found."
        continue
    }

    $envAssignments = "`$env:DATABASE_PATH = '$dbPath';"
    if ($service.NeedsAuthUrl) {
        $envAssignments += " `$env:AUTH_SERVICE_URL = '$authServiceUrl';"
    }

    $command = "& { Set-Location -LiteralPath '$servicePath'; $envAssignments go run main.go }"
    
    # Start in new PowerShell window
    Start-Process $shellExecutable -ArgumentList "-NoExit", "-Command", $command
    
    # Small delay between starts
    Start-Sleep -Milliseconds 500
}

Write-Host ""
Write-Host "All services started!" -ForegroundColor Green
Write-Host ""
Write-Host "Services running on:" -ForegroundColor Yellow
Write-Host "  - Auth:          http://localhost:8081" -ForegroundColor White
Write-Host "  - Users:         http://localhost:8082" -ForegroundColor White
Write-Host "  - Posts:         http://localhost:8083" -ForegroundColor White
Write-Host "  - Groups:        http://localhost:8084" -ForegroundColor White
Write-Host "  - Chat:          http://localhost:8085" -ForegroundColor White
Write-Host "  - Notifications: http://localhost:8086" -ForegroundColor White
Write-Host ""
Write-Host "Press any key to close this window..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

# Database Migration Script (PowerShell)
# This script applies all pending migrations to bring the database to the latest version

$ErrorActionPreference = "Stop"

# Configuration
$DB_PATH = if ($env:DATABASE_PATH) { $env:DATABASE_PATH } else { "..\social_network.db" }
$MIGRATIONS_DIR = ".\migrations"

# Convert Windows paths to forward slashes for migrate tool
$DB_PATH = $DB_PATH -replace '\\', '/'
$MIGRATIONS_DIR = $MIGRATIONS_DIR -replace '\\', '/'

Write-Host "=== Database Migration Script ===" -ForegroundColor Cyan
Write-Host "Database: $DB_PATH"
Write-Host "Migrations: $MIGRATIONS_DIR"
Write-Host ""

# Check if migrate tool is installed
try {
    $null = Get-Command migrate -ErrorAction Stop
} catch {
    Write-Host "Error: 'migrate' command not found." -ForegroundColor Red
    Write-Host "Please install golang-migrate:"
    Write-Host "  Windows: scoop install migrate"
    Write-Host "  or: go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    Write-Host "  macOS/Linux: brew install golang-migrate"
    exit 1
}

# Run migrations
Write-Host "Applying migrations..." -ForegroundColor Yellow

$dbUri = "sqlite3://$DB_PATH"
migrate -path $MIGRATIONS_DIR -database $dbUri up

if ($LASTEXITCODE -eq 0 -or $LASTEXITCODE -eq $null) {
    Write-Host ""
    Write-Host "[SUCCESS] Migrations completed successfully!" -ForegroundColor Green
    
    # Show current version (suppress stderr output)
    $ErrorActionPreference = "SilentlyContinue"
    $version = (migrate -path $MIGRATIONS_DIR -database $dbUri version) 2>$null
    $ErrorActionPreference = "Stop"
    if ($version) {
        Write-Host "Current database version: $version" -ForegroundColor Cyan
    }
} else {
    Write-Host ""
    Write-Host "[ERROR] Migration failed with exit code: $LASTEXITCODE" -ForegroundColor Red
    exit 1
}

# PowerShell Database migration script using golang-migrate

param(
    [Parameter(Position=0)]
    [string]$Command = "up",
    
    [Parameter(Position=1)]
    [string]$Arg = ""
)

$ErrorActionPreference = 'Stop'

# Load .env file if exists
$envFile = Join-Path $PSScriptRoot ".." ".env"
if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            [Environment]::SetEnvironmentVariable($name, $value, "Process")
        }
    }
}

# Default values
$MIGRATIONS_PATH = if ($env:MIGRATIONS_PATH) { $env:MIGRATIONS_PATH } else { "./migrations" }
$DB_HOST = if ($env:DB_HOST) { $env:DB_HOST } else { "localhost" }
$DB_PORT = if ($env:DB_PORT) { $env:DB_PORT } else { "5432" }
$DB_USER = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
$DB_PASSWORD = if ($env:DB_PASSWORD) { $env:DB_PASSWORD } else { "postgres" }
$DB_NAME = if ($env:DB_NAME) { $env:DB_NAME } else { "filesharing" }
$DB_SSLMODE = if ($env:DB_SSLMODE) { $env:DB_SSLMODE } else { "disable" }

# Construct database URL
$DATABASE_URL = "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

# Check if migrate tool is installed
if (-not (Get-Command migrate -ErrorAction SilentlyContinue)) {
    Write-Host "Error: golang-migrate is not installed" -ForegroundColor Red
    Write-Host ""
    Write-Host "Install it with:" -ForegroundColor Yellow
    Write-Host "  Chocolatey: choco install golang-migrate"
    Write-Host "  Scoop: scoop install migrate"
    Write-Host "  Or with Go: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    Write-Host ""
    Write-Host "Download from: https://github.com/golang-migrate/migrate/releases"
    exit 1
}

# Execute command
switch ($Command.ToLower()) {
    "up" {
        Write-Host "Running migrations UP..." -ForegroundColor Yellow
        migrate -path $MIGRATIONS_PATH -database $DATABASE_URL up
    }
    "down" {
        $steps = if ($Arg) { $Arg } else { "1" }
        Write-Host "Running migrations DOWN (steps: $steps)..." -ForegroundColor Yellow
        migrate -path $MIGRATIONS_PATH -database $DATABASE_URL down $steps
    }
    "drop" {
        Write-Host "WARNING: This will drop all tables!" -ForegroundColor Red
        $confirm = Read-Host "Are you sure? (yes/no)"
        if ($confirm -eq "yes") {
            Write-Host "Dropping all migrations..." -ForegroundColor Yellow
            migrate -path $MIGRATIONS_PATH -database $DATABASE_URL drop -f
        } else {
            Write-Host "Cancelled." -ForegroundColor Green
            exit 0
        }
    }
    "force" {
        if (-not $Arg) {
            Write-Host "Error: version required for force command" -ForegroundColor Red
            Write-Host "Usage: .\migrate.ps1 force <version>"
            exit 1
        }
        Write-Host "Forcing migration version to $Arg..." -ForegroundColor Yellow
        migrate -path $MIGRATIONS_PATH -database $DATABASE_URL force $Arg
    }
    "version" {
        Write-Host "Current migration version:" -ForegroundColor Yellow
        migrate -path $MIGRATIONS_PATH -database $DATABASE_URL version
    }
    "create" {
        if (-not $Arg) {
            Write-Host "Error: migration name required" -ForegroundColor Red
            Write-Host "Usage: .\migrate.ps1 create <migration_name>"
            exit 1
        }
        Write-Host "Creating new migration: $Arg" -ForegroundColor Yellow
        migrate create -ext sql -dir $MIGRATIONS_PATH -seq $Arg
    }
    default {
        Write-Host "Usage: .\migrate.ps1 {up|down|drop|force|version|create} [args]" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Commands:"
        Write-Host "  up              - Apply all pending migrations"
        Write-Host "  down [N]        - Revert last N migrations (default: 1)"
        Write-Host "  drop            - Drop all tables (requires confirmation)"
        Write-Host "  force <version> - Set migration version without running migrations"
        Write-Host "  version         - Show current migration version"
        Write-Host "  create <name>   - Create new migration files"
        exit 1
    }
}

Write-Host "Done!" -ForegroundColor Green

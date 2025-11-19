# =============================================================================
# PowerShell Development Helper Script for Windows
# =============================================================================
# Usage: .\dev.ps1 [command]
# Alternative to Makefile for Windows users without Make
# =============================================================================

param(
    [Parameter(Position=0)]
    [string]$Command = "help",
    
    [Parameter(Position=1)]
    [string]$Arg = ""
)

$ErrorActionPreference = 'Stop'

# Colors
function Write-Success { param([string]$msg) Write-Host $msg -ForegroundColor Green }
function Write-Info { param([string]$msg) Write-Host $msg -ForegroundColor Cyan }
function Write-Warn { param([string]$msg) Write-Host $msg -ForegroundColor Yellow }
function Write-Error { param([string]$msg) Write-Host $msg -ForegroundColor Red }

# Load .env file
function Load-Env {
    if (Test-Path .env) {
        Get-Content .env | ForEach-Object {
            if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
                $name = $matches[1].Trim()
                $value = $matches[2].Trim()
                [Environment]::SetEnvironmentVariable($name, $value, "Process")
            }
        }
    }
}

# Database URL
function Get-DatabaseUrl {
    Load-Env
    $dbUser = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
    $dbPass = if ($env:DB_PASSWORD) { $env:DB_PASSWORD } else { "postgres" }
    $dbHost = "postgres"
    $dbPort = "5432"
    $dbName = if ($env:DB_NAME) { $env:DB_NAME } else { "file_sharing_db" }
    $dbSsl = if ($env:DB_SSLMODE) { $env:DB_SSLMODE } else { "disable" }
    
    return "postgres://${dbUser}:${dbPass}@${dbHost}:${dbPort}/${dbName}?sslmode=${dbSsl}"
}

# Commands
switch ($Command.ToLower()) {
    "setup" {
        Write-Info "════════════════════════════════════════════════════════════════"
        Write-Info "  Starting Complete Setup"
        Write-Info "════════════════════════════════════════════════════════════════"
        Write-Host ""
        
        Write-Info "Step 1/4: Checking .env file..."
        if (-not (Test-Path .env)) {
            if (Test-Path .env.example) {
                Copy-Item .env.example .env
                Write-Success "✓ .env file created from .env.example"
            }
        } else {
            Write-Success "✓ .env file exists"
        }
        Write-Host ""
        
        Write-Info "Step 2/4: Starting Docker services..."
        docker-compose up -d postgres
        Write-Success "✓ PostgreSQL started"
        Write-Host ""
        
        Write-Info "Step 3/4: Waiting for database to be ready..."
        Start-Sleep -Seconds 5
        Write-Host ""
        
        Write-Info "Step 4/4: Running migrations..."
        $dbUrl = Get-DatabaseUrl
        docker-compose run --rm migrate -path=/migrations -database $dbUrl up
        Write-Host ""
        
        Write-Success "════════════════════════════════════════════════════════════════"
        Write-Success "  Setup Complete! 🎉"
        Write-Success "════════════════════════════════════════════════════════════════"
        Write-Host ""
        Write-Warn "Next steps:"
        Write-Host "  • Start development: " -NoNewline; Write-Success ".\dev.ps1 dev"
        Write-Host "  • View database: " -NoNewline; Write-Success ".\dev.ps1 db-shell"
        Write-Host "  • Check status: " -NoNewline; Write-Success ".\dev.ps1 verify"
        Write-Host ""
    }
    
    "dev" {
        Write-Info "Starting development environment..."
        docker-compose --profile dev up -d
        Write-Success "✓ Dev environment started!"
        Write-Warn "API running at: http://localhost:8080"
        Write-Warn "View logs: .\dev.ps1 logs"
    }
    
    "prod" {
        Write-Info "Starting production environment..."
        docker-compose --profile prod up -d
        Write-Success "✓ Production environment started!"
    }
    
    "down" {
        Write-Info "Stopping all services..."
        docker-compose down
        Write-Success "✓ All services stopped!"
    }
    
    "restart" {
        Write-Info "Restarting services..."
        docker-compose down
        docker-compose up -d postgres
        Write-Success "✓ Services restarted!"
    }
    
    "logs" {
        docker-compose logs -f app-dev
    }
    
    "logs-db" {
        docker-compose logs -f postgres
    }
    
    "logs-all" {
        docker-compose logs -f
    }
    
    "ps" {
        docker-compose ps
    }
    
    "db-shell" {
        Write-Info "Connecting to PostgreSQL..."
        Load-Env
        $dbUser = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
        $dbName = if ($env:DB_NAME) { $env:DB_NAME } else { "file_sharing_db" }
        docker-compose exec postgres psql -U $dbUser -d $dbName
    }
    
    "db-status" {
        Write-Info "Checking database status..."
        Load-Env
        $dbUser = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
        docker-compose exec postgres pg_isready -U $dbUser
    }
    
    "migrate-up" {
        Write-Info "Running migrations UP..."
        $dbUrl = Get-DatabaseUrl
        docker-compose run --rm migrate -path=/migrations -database $dbUrl up
        Write-Success "✓ Migrations completed!"
    }
    
    "migrate-down" {
        Write-Warn "Rolling back last migration..."
        $dbUrl = Get-DatabaseUrl
        docker-compose run --rm migrate -path=/migrations -database $dbUrl down 1
        Write-Success "✓ Rollback completed!"
    }
    
    "migrate-version" {
        Write-Info "Current migration version:"
        $dbUrl = Get-DatabaseUrl
        docker-compose run --rm migrate -path=/migrations -database $dbUrl version
    }
    
    "migrate-create" {
        if (-not $Arg) {
            Write-Error "Migration name required"
            Write-Warn "Usage: .\dev.ps1 migrate-create my_migration"
            exit 1
        }
        Write-Info "Creating migration: $Arg"
        docker-compose run --rm migrate create -ext sql -dir /migrations -seq $Arg
        Write-Success "✓ Migration files created!"
    }
    
    "migrate-force" {
        if (-not $Arg) {
            Write-Error "Version required"
            Write-Warn "Usage: .\dev.ps1 migrate-force 1"
            exit 1
        }
        Write-Warn "Forcing migration version to $Arg..."
        $dbUrl = Get-DatabaseUrl
        docker-compose run --rm migrate -path=/migrations -database $dbUrl force $Arg
        Write-Success "✓ Version forced to $Arg!"
    }
    
    "verify" {
        Write-Info "Verifying setup..."
        Write-Host ""
        
        Write-Warn "1. Database Status:"
        Load-Env
        $dbUser = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
        docker-compose exec postgres pg_isready -U $dbUser 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "   ✓ Database is ready"
        } else {
            Write-Error "   ✗ Database is not ready"
        }
        Write-Host ""
        
        Write-Warn "2. Migration Version:"
        $dbUrl = Get-DatabaseUrl
        docker-compose run --rm migrate -path=/migrations -database $dbUrl version 2>$null
        Write-Host ""
        
        Write-Warn "3. Tables:"
        $dbName = if ($env:DB_NAME) { $env:DB_NAME } else { "file_sharing_db" }
        $tableCount = docker-compose exec postgres psql -U $dbUser -d $dbName -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>$null
        if ($tableCount) {
            Write-Host "   Tables found: $($tableCount.Trim())"
        }
        Write-Host ""
    }
    
    "check" {
        Write-Info "Running health checks..."
        Write-Host ""
        
        Write-Warn "Database:"
        Load-Env
        $dbUser = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
        docker-compose exec postgres pg_isready -U $dbUser 2>$null
        if ($LASTEXITCODE -eq 0) {
            Write-Success "  ✓ Ready"
        } else {
            Write-Error "  ✗ Not ready"
        }
        Write-Host ""
        
        Write-Warn "Migration Version:"
        $dbUrl = Get-DatabaseUrl
        docker-compose run --rm migrate -path=/migrations -database $dbUrl version 2>$null | Select-Object -First 1
        Write-Host ""
    }
    
    "clean" {
        Write-Error "WARNING: This will remove all data!"
        $confirm = Read-Host "Type 'yes' to confirm"
        if ($confirm -eq "yes") {
            docker-compose down -v --remove-orphans
            Write-Success "✓ Cleaned up!"
        } else {
            Write-Warn "Cancelled."
        }
    }
    
    "reset" {
        Write-Warn "Resetting everything..."
        docker-compose down -v --remove-orphans
        Write-Host ""
        & $PSCommandPath setup
    }
    
    "tools" {
        Write-Info "Starting tools..."
        docker-compose --profile tools up -d
        Write-Success "✓ Tools started!"
        Write-Warn "Adminer: http://localhost:8081"
    }
    
    default {
        Write-Host ""
        Write-Success "════════════════════════════════════════════════════════════════"
        Write-Success "  File Sharing API - Development Helper (Windows)"
        Write-Success "════════════════════════════════════════════════════════════════"
        Write-Host ""
        Write-Warn "Usage: .\dev.ps1 [command]"
        Write-Host ""
        Write-Warn "Common Commands:"
        Write-Host "  setup              - Complete setup (first time)"
        Write-Host "  dev                - Start development environment"
        Write-Host "  down               - Stop all services"
        Write-Host "  restart            - Restart services"
        Write-Host ""
        Write-Warn "Database Migrations:"
        Write-Host "  migrate-up         - Apply pending migrations"
        Write-Host "  migrate-down       - Rollback last migration"
        Write-Host "  migrate-version    - Show current version"
        Write-Host "  migrate-create <name> - Create new migration"
        Write-Host "  migrate-force <ver>   - Force version"
        Write-Host ""
        Write-Warn "Database Operations:"
        Write-Host "  db-shell           - Open PostgreSQL shell"
        Write-Host "  db-status          - Check database status"
        Write-Host ""
        Write-Warn "Monitoring:"
        Write-Host "  logs               - View application logs"
        Write-Host "  logs-db            - View database logs"
        Write-Host "  logs-all           - View all logs"
        Write-Host "  ps                 - Show running containers"
        Write-Host ""
        Write-Warn "Maintenance:"
        Write-Host "  verify             - Verify complete setup"
        Write-Host "  check              - Quick health check"
        Write-Host "  tools              - Start Adminer (DB UI)"
        Write-Host "  clean              - Remove all data (careful!)"
        Write-Host "  reset              - Reset everything"
        Write-Host ""
        Write-Warn "Examples:"
        Write-Host "  .\dev.ps1 setup"
        Write-Host "  .\dev.ps1 dev"
        Write-Host "  .\dev.ps1 migrate-create add_new_table"
        Write-Host ""
        Write-Info "For Linux/Mac users: Use 'make' commands instead"
        Write-Info "See SETUP.md for more details"
        Write-Host ""
    }
}


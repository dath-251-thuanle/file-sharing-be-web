# PowerShell Database setup and migration script for Windows
# This script automates PostgreSQL setup and migrations
# Usage: .\setup-db.ps1 [setup|migrate|up|down|version|create]

param(
    [Parameter(Position=0)]
    [string]$Command = "setup",
    
    [Parameter(Position=1)]
    [string]$Arg = ""
)

$ErrorActionPreference = 'Stop'

# Colors for output
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Load environment variables
function Load-Env {
    $envFile = Join-Path $PSScriptRoot ".." ".env"
    $envExampleFile = Join-Path $PSScriptRoot ".." ".env.example"
    
    # Check if .env file exists
    if (-not (Test-Path $envFile)) {
        if (Test-Path $envExampleFile) {
            Write-Warn ".env file not found"
            Write-Info "Creating .env from .env.example..."
            Copy-Item $envExampleFile $envFile
            Write-Info "✓ .env file created from .env.example"
            Write-Warn "⚠ Please review and update .env file with your settings (especially DB_PASSWORD)"
            Write-Warn "⚠ Edit .env file now, then press Enter to continue..."
            Read-Host "Press Enter after updating .env file"
        } else {
            Write-Error ".env file is required but not found"
            Write-Error ".env.example file is also not found"
            Write-Error "Please create .env file manually or ensure .env.example exists"
            exit 1
        }
    }
    
    # Load .env file
    Write-Info "Loading environment variables from .env"
    
    Get-Content $envFile | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+)=(.*)$') {
            $name = $matches[1].Trim()
            $value = $matches[2].Trim()
            [Environment]::SetEnvironmentVariable($name, $value, "Process")
        }
    }
    
    # Validate required variables
    $missingVars = @()
    if (-not $env:DB_HOST) { $missingVars += "DB_HOST" }
    if (-not $env:DB_PORT) { $missingVars += "DB_PORT" }
    if (-not $env:DB_USER) { $missingVars += "DB_USER" }
    if (-not $env:DB_PASSWORD) { $missingVars += "DB_PASSWORD" }
    if (-not $env:DB_NAME) { $missingVars += "DB_NAME" }
    
    if ($missingVars.Count -gt 0) {
        Write-Error "Missing required environment variables in .env:"
        foreach ($var in $missingVars) {
            Write-Error "  - $var"
        }
        exit 1
    }
    
    Write-Info "✓ Environment variables loaded successfully"
}

# Check if PostgreSQL is installed
function Test-PostgreSQL {
    if (-not (Get-Command psql -ErrorAction SilentlyContinue)) {
        Write-Warn "PostgreSQL is not installed"
        $install = Read-Host "Do you want to install PostgreSQL? (y/n)"
        if ($install -eq "y" -or $install -eq "Y") {
            Write-Info "Installing PostgreSQL..."
            Write-Info "Please install PostgreSQL from: https://www.postgresql.org/download/windows/"
            Write-Info "Or use Chocolatey: choco install postgresql"
            Write-Error "PostgreSQL installation required. Exiting."
            exit 1
        } else {
            Write-Error "PostgreSQL is required. Exiting."
            exit 1
        }
    }
    Write-Info "PostgreSQL is installed"
}

# Start PostgreSQL service
function Start-PostgreSQL {
    Write-Info "Starting PostgreSQL service..."
    
    $service = Get-Service -Name "postgresql*" -ErrorAction SilentlyContinue
    if ($service) {
        if ($service.Status -ne "Running") {
            Start-Service $service.Name
            Write-Info "PostgreSQL service started"
        } else {
            Write-Info "PostgreSQL service is already running"
        }
    } else {
        Write-Warn "PostgreSQL service not found. Please start it manually."
    }
    
    # Wait a bit for PostgreSQL to start
    Start-Sleep -Seconds 3
    
    # Check if PostgreSQL is running
    try {
        $result = & psql -h localhost -U postgres -c "SELECT 1;" 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Info "PostgreSQL is running"
        } else {
            Write-Error "Failed to connect to PostgreSQL"
            exit 1
        }
    } catch {
        Write-Error "Failed to verify PostgreSQL connection"
        exit 1
    }
}

# Setup PostgreSQL password
function Set-PostgreSQLPassword {
    $dbPassword = if ($env:DB_PASSWORD) { $env:DB_PASSWORD } else { "postgres" }
    $dbUser = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
    
    Write-Info "Setting up password for PostgreSQL user: $dbUser"
    
    # Set password using psql
    $sql = "ALTER USER $dbUser WITH PASSWORD '$dbPassword';"
    try {
        & psql -h localhost -U postgres -c $sql 2>&1 | Out-Null
        Write-Info "Password set successfully for user $dbUser"
    } catch {
        Write-Warn "Failed to set password (may already be set or using different auth method)"
    }
}

# Create database
function New-Database {
    $dbName = if ($env:DB_NAME) { $env:DB_NAME } else { "file_sharing_db" }
    $dbUser = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
    
    Write-Info "Creating database: $dbName"
    
    # Check if database exists
    $checkDb = & psql -h localhost -U postgres -lqt 2>&1 | Select-String $dbName
    if ($checkDb) {
        Write-Warn "Database $dbName already exists"
        $recreate = Read-Host "Do you want to drop and recreate it? (y/n)"
        if ($recreate -eq "y" -or $recreate -eq "Y") {
            Write-Info "Dropping existing database..."
            & psql -h localhost -U postgres -c "DROP DATABASE IF EXISTS $dbName;" 2>&1 | Out-Null
            & createdb -h localhost -U postgres $dbName 2>&1 | Out-Null
            Write-Info "Database $dbName created"
        } else {
            Write-Info "Using existing database"
        }
    } else {
        & createdb -h localhost -U postgres $dbName 2>&1 | Out-Null
        Write-Info "Database $dbName created"
    }
    
    # Set password for postgres user if DB_PASSWORD is provided
    if ($env:DB_PASSWORD -and $dbUser -eq "postgres") {
        Write-Info "Setting password for postgres user"
        $sql = "ALTER USER postgres WITH PASSWORD '$($env:DB_PASSWORD)';"
        & psql -h localhost -U postgres -c $sql 2>&1 | Out-Null
    }
    
    # Create user if specified and different from postgres
    if ($dbUser -ne "postgres") {
        Write-Info "Creating user: $dbUser"
        & createuser -h localhost -U postgres -s $dbUser 2>&1 | Out-Null
        
        if ($env:DB_PASSWORD) {
            Write-Info "Setting password for user $dbUser"
            $sql = "ALTER USER $dbUser WITH PASSWORD '$($env:DB_PASSWORD)';"
            & psql -h localhost -U postgres -c $sql 2>&1 | Out-Null
        }
        
        # Grant privileges
        $sql = "GRANT ALL PRIVILEGES ON DATABASE $dbName TO $dbUser;"
        & psql -h localhost -U postgres -c $sql 2>&1 | Out-Null
    }
}

# Check if golang-migrate is installed
function Test-Migrate {
    if (-not (Get-Command migrate -ErrorAction SilentlyContinue)) {
        Write-Warn "golang-migrate is not installed"
        $install = Read-Host "Do you want to install golang-migrate? (y/n)"
        if ($install -eq "y" -or $install -eq "Y") {
            Write-Info "Installing golang-migrate..."
            Write-Info "Please install using one of these methods:"
            Write-Info "  Chocolatey: choco install golang-migrate"
            Write-Info "  Scoop: scoop install migrate"
            Write-Info "  Or with Go: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
            Write-Error "golang-migrate is required. Exiting."
            exit 1
        } else {
            Write-Error "golang-migrate is required. Exiting."
            exit 1
        }
    }
    Write-Info "golang-migrate is installed"
}

# Run migrations
function Invoke-Migrations {
    param(
        [string]$MigrateCommand = "up",
        [string]$MigrateArg = ""
    )
    
    $migrationsPath = if ($env:MIGRATIONS_PATH) { $env:MIGRATIONS_PATH } else { "./migrations" }
    
    if (-not (Test-Path $migrationsPath)) {
        Write-Error "Migrations directory not found: $migrationsPath"
        exit 1
    }
    
    # Construct database URL
    $dbHost = if ($env:DB_HOST) { $env:DB_HOST } else { "localhost" }
    $dbPort = if ($env:DB_PORT) { $env:DB_PORT } else { "5432" }
    $dbUser = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
    $dbPassword = if ($env:DB_PASSWORD) { $env:DB_PASSWORD } else { "postgres" }
    $dbName = if ($env:DB_NAME) { $env:DB_NAME } else { "file_sharing_db" }
    $dbSslMode = if ($env:DB_SSLMODE) { $env:DB_SSLMODE } else { "disable" }
    
    $databaseUrl = "postgres://${dbUser}:${dbPassword}@${dbHost}:${dbPort}/${dbName}?sslmode=${dbSslMode}"
    
    switch ($MigrateCommand.ToLower()) {
        "up" {
            Write-Info "Running migrations UP..."
            & migrate -path $migrationsPath -database $databaseUrl up
        }
        "down" {
            $steps = if ($MigrateArg) { $MigrateArg } else { "1" }
            Write-Info "Running migrations DOWN (steps: $steps)..."
            & migrate -path $migrationsPath -database $databaseUrl down $steps
        }
        "drop" {
            Write-Warn "Dropping all migrations..."
            $confirm = Read-Host "Are you sure? This will drop all tables! (yes/no)"
            if ($confirm -eq "yes") {
                & migrate -path $migrationsPath -database $databaseUrl drop -f
            } else {
                Write-Info "Cancelled."
            }
        }
        "force" {
            if (-not $MigrateArg) {
                Write-Error "Version required for force command"
                Write-Error "Usage: .\setup-db.ps1 force <version>"
                exit 1
            }
            Write-Info "Forcing migration version to $MigrateArg..."
            & migrate -path $migrationsPath -database $databaseUrl force $MigrateArg
        }
        "version" {
            Write-Info "Current migration version:"
            & migrate -path $migrationsPath -database $databaseUrl version
        }
        "create" {
            if (-not $MigrateArg) {
                Write-Error "Migration name required"
                Write-Error "Usage: .\setup-db.ps1 create <migration_name>"
                exit 1
            }
            Write-Info "Creating new migration: $MigrateArg"
            & migrate create -ext sql -dir $migrationsPath -seq $MigrateArg
        }
        default {
            Write-Error "Unknown migration command: $MigrateCommand"
            Write-Info "Available commands: up, down, drop, force, version, create"
            exit 1
        }
    }
}

# Verify setup
function Test-Setup {
    Write-Info "Verifying database setup..."
    
    $dbName = if ($env:DB_NAME) { $env:DB_NAME } else { "file_sharing_db" }
    
    # Check if tables exist
    $tableCount = & psql -h localhost -U postgres -d $dbName -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>&1
    $tableCount = $tableCount.Trim()
    
    if ($tableCount -gt 0) {
        Write-Info "Found $tableCount tables in database"
        
        # Check for demo data
        $userCount = & psql -h localhost -U postgres -d $dbName -t -c "SELECT COUNT(*) FROM users;" 2>&1
        $userCount = $userCount.Trim()
        
        $fileCount = & psql -h localhost -U postgres -d $dbName -t -c "SELECT COUNT(*) FROM files;" 2>&1
        $fileCount = $fileCount.Trim()
        
        Write-Info "Users: $userCount"
        Write-Info "Files: $fileCount"
        
        Write-Info "✓ Database schema is ready (no demo data is seeded by default)"
    } else {
        Write-Error "No tables found in database"
        exit 1
    }
}

# Run demo queries (optional)
function Invoke-DemoQueries {
    $demoFile = "pkg/database/demo_queries.sql"
    
    if (-not (Test-Path $demoFile)) {
        Write-Warn "Demo queries file not found: $demoFile"
        return
    }
    
    Write-Host ""
    Write-Warn "Demo queries are deprecated and will not be executed automatically."
    Write-Info "If you still want to use them, run manually:"
    Write-Info "  psql -h localhost -U postgres -d file_sharing_db -f $demoFile"
}

# Full setup (create DB, migrate)
function Start-Setup {
    Write-Info "Starting database setup for Windows..."
    Write-Host ""
    
    Load-Env
    Test-PostgreSQL
    Start-PostgreSQL
    Set-PostgreSQLPassword
    New-Database
    Test-Migrate
    Invoke-Migrations "up"
    Test-Setup
    # Demo queries are no longer run automatically
    
    Write-Host ""
    $dbHost = if ($env:DB_HOST) { $env:DB_HOST } else { "localhost" }
    $dbUser = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
    $dbName = if ($env:DB_NAME) { $env:DB_NAME } else { "file_sharing_db" }
    Write-Info "Setup complete! You can now:"
    Write-Info "  1. Connect to database: psql -h $dbHost -U $dbUser -d $dbName"
    Write-Info "  2. Check migrations: .\setup-db.ps1 version"
    Write-Info "  3. Start the backend server: go run cmd/api/main.go"
    Write-Host ""
}

# Main function
function Main {
    param([string]$Cmd, [string]$CmdArg)
    
    switch ($Cmd.ToLower()) {
        "setup" {
            Start-Setup
        }
        "migrate" {
            Load-Env
            Test-Migrate
            Invoke-Migrations "up"
        }
        "up" {
            Load-Env
            Test-Migrate
            Invoke-Migrations "up"
        }
        "down" {
            Load-Env
            Test-Migrate
            Invoke-Migrations "down" $CmdArg
        }
        "version" {
            Load-Env
            Test-Migrate
            Invoke-Migrations "version"
        }
        "create" {
            Load-Env
            Test-Migrate
            Invoke-Migrations "create" $CmdArg
        }
        "force" {
            Load-Env
            Test-Migrate
            Invoke-Migrations "force" $CmdArg
        }
        "drop" {
            Load-Env
            Test-Migrate
            Invoke-Migrations "drop"
        }
        default {
            Write-Host "Usage: .\setup-db.ps1 {setup|migrate|up|down|version|create|force|drop} [args]" -ForegroundColor Yellow
            Write-Host ""
            Write-Host "Commands:"
            Write-Host "  setup           - Full database setup (create DB, migrate)"
            Write-Host "  migrate|up      - Apply all pending migrations"
            Write-Host "  down [N]        - Revert last N migrations (default: 1)"
            Write-Host "  drop            - Drop all tables (requires confirmation)"
            Write-Host "  force <version> - Set migration version without running migrations"
            Write-Host "  version         - Show current migration version"
            Write-Host "  create <name>   - Create new migration files"
            exit 1
        }
    }
}

# Run main function
Main -Cmd $Command -CmdArg $Arg


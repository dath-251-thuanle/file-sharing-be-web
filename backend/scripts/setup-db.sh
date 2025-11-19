#!/bin/bash
# Database setup and migration script for Linux/WSL
# This script automates PostgreSQL setup and migrations
# Usage: ./setup-db.sh [setup|migrate|up|down|version|create]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Print colored message
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Determine if we need sudo (not needed when running as root)
SUDO_CMD=""
if [ "$EUID" -ne 0 ]; then
    SUDO_CMD="sudo"
fi

# Function to run command as postgres user
run_as_postgres() {
    if [ "$EUID" -eq 0 ]; then
        # Running as root, use runuser or su
        if command -v runuser &> /dev/null; then
            runuser -u postgres -- "$@"
        else
            # Use su with proper quoting
            su - postgres -c "$(printf '%q ' "$@")"
        fi
    else
        # Not root, use sudo
        sudo -u postgres "$@"
    fi
}

# Load environment variables
load_env() {
    # Check if .env file exists
    if [ ! -f .env ]; then
        if [ -f .env.example ]; then
            print_warn ".env file not found"
            print_info "Creating .env from .env.example..."
            cp .env.example .env
            print_info "✓ .env file created from .env.example"
            print_warn "⚠ Please review and update .env file with your settings (especially DB_PASSWORD)"
            print_warn "⚠ Edit .env file now, then press Enter to continue..."
            read -p "Press Enter after updating .env file..."
        else
            print_error ".env file is required but not found"
            print_error ".env.example file is also not found"
            print_error "Please create .env file manually or ensure .env.example exists"
            exit 1
        fi
    fi
    
    # Load .env file
    print_info "Loading environment variables from .env"
    
    # Safe way to load .env - only process lines with valid variable format
    set -a  # Automatically export all variables
    while IFS='=' read -r key value; do
        # Skip comments and empty lines
        [[ $key =~ ^#.*$ ]] && continue
        [[ -z "$key" ]] && continue
        
        # Remove leading/trailing whitespace and carriage return (for Windows line endings)
        key=$(echo "$key" | tr -d '\r' | xargs)
        value=$(echo "$value" | tr -d '\r' | xargs)
        
        # Only export if key looks like a valid variable name
        if [[ $key =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]]; then
            export "$key=$value"
        fi
    done < .env
    set +a
    
    # Check required variables
    local missing_vars=()
    [ -z "$DB_HOST" ] && missing_vars+=("DB_HOST")
    [ -z "$DB_PORT" ] && missing_vars+=("DB_PORT")
    [ -z "$DB_USER" ] && missing_vars+=("DB_USER")
    [ -z "$DB_PASSWORD" ] && missing_vars+=("DB_PASSWORD")
    [ -z "$DB_NAME" ] && missing_vars+=("DB_NAME")
    
    if [ ${#missing_vars[@]} -gt 0 ]; then
        print_error "Missing required environment variables in .env:"
        for var in "${missing_vars[@]}"; do
            print_error "  - $var"
        done
        exit 1
    fi
    
    print_info "✓ Environment variables loaded successfully"
}

# Check if running as root
check_root() {
    if [ "$EUID" -eq 0 ]; then 
        print_warn "Running as root user"
        print_warn "This is allowed but not recommended for security reasons"
        read -p "Do you want to continue? (y/n): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Exiting. Please run as a regular user or use 'su - <username>' to switch user"
            exit 0
        fi
    fi
}

# Check if PostgreSQL is installed
check_postgresql() {
    if ! command -v psql &> /dev/null; then
        print_warn "PostgreSQL is not installed"
        read -p "Do you want to install PostgreSQL? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_info "Installing PostgreSQL..."
            $SUDO_CMD apt update
            $SUDO_CMD apt install postgresql postgresql-contrib -y
        else
            print_error "PostgreSQL is required. Exiting."
            exit 1
        fi
    fi
    print_info "PostgreSQL is installed"
}

# Start PostgreSQL service
start_postgresql() {
    print_info "Starting PostgreSQL service..."
    
    # Try systemctl first (WSL2/Linux)
    if command -v systemctl &> /dev/null; then
        $SUDO_CMD systemctl start postgresql 2>/dev/null || true
        $SUDO_CMD systemctl enable postgresql 2>/dev/null || true
    fi
    
    # Try service command (WSL1/older systems)
    $SUDO_CMD service postgresql start 2>/dev/null || true
    
    # Wait a bit for PostgreSQL to start
    sleep 3
    
    # Check if PostgreSQL is running
    if run_as_postgres pg_isready &> /dev/null; then
        print_info "PostgreSQL is running"
    else
        print_error "Failed to start PostgreSQL"
        exit 1
    fi
}

# Setup PostgreSQL password for postgres user
setup_postgres_password() {
    local db_password="${DB_PASSWORD:-postgres}"
    local db_user="${DB_USER:-postgres}"
    
    print_info "Setting up password for PostgreSQL user: $db_user"
    
    # Set password for postgres user
    if run_as_postgres psql -c "ALTER USER $db_user WITH PASSWORD '$db_password';" 2>/dev/null; then
        print_info "Password set successfully for user $db_user"
    else
        print_warn "Failed to set password (may already be set or using different auth method)"
    fi
}

# Create database
create_database() {
    local db_name="${DB_NAME:-file_sharing_db}"
    local db_user="${DB_USER:-postgres}"
    
    print_info "Creating database: $db_name"
    
    # Check if database exists
    if run_as_postgres psql -lqt | cut -d \| -f 1 | grep -qw "$db_name"; then
        print_warn "Database $db_name already exists"
        read -p "Do you want to drop and recreate it? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_info "Dropping existing database..."
            run_as_postgres dropdb "$db_name" 2>/dev/null || true
            run_as_postgres createdb "$db_name"
            print_info "Database $db_name created"
        else
            print_info "Using existing database"
        fi
    else
        run_as_postgres createdb "$db_name"
        print_info "Database $db_name created"
    fi
    
    # Always set password for postgres user if DB_PASSWORD is provided
    if [ -n "$DB_PASSWORD" ] && [ "$db_user" = "postgres" ]; then
        print_info "Setting password for postgres user"
        run_as_postgres psql -c "ALTER USER postgres WITH PASSWORD '$DB_PASSWORD';" 2>/dev/null || print_warn "Could not set password (may already be set)"
    fi
    
    # Create user if specified and different from postgres
    if [ "$db_user" != "postgres" ]; then
        print_info "Creating user: $db_user"
        run_as_postgres createuser -s "$db_user" 2>/dev/null || print_warn "User $db_user may already exist"
        
        # Set password if provided
        if [ -n "$DB_PASSWORD" ]; then
            print_info "Setting password for user $db_user"
            run_as_postgres psql -c "ALTER USER $db_user WITH PASSWORD '$DB_PASSWORD';"
        fi
        
        # Grant privileges
        run_as_postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE $db_name TO $db_user;"
    fi
}

# Check if golang-migrate is installed
check_migrate() {
    if ! command -v migrate &> /dev/null; then
        print_warn "golang-migrate is not installed"
        read -p "Do you want to install golang-migrate? (y/n): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            print_info "Installing golang-migrate..."
            
            # Try download binary
            if curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz -o /tmp/migrate.tar.gz 2>/dev/null; then
                tar -xzf /tmp/migrate.tar.gz -C /tmp
                $SUDO_CMD mv /tmp/migrate /usr/local/bin/
                rm /tmp/migrate.tar.gz
                print_info "golang-migrate installed"
            else
                print_error "Failed to download golang-migrate"
                print_info "You can install manually:"
                print_info "  curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz"
                print_info "  sudo mv migrate /usr/local/bin/"
                exit 1
            fi
        else
            print_error "golang-migrate is required. Exiting."
            exit 1
        fi
    fi
    print_info "golang-migrate is installed"
}

# Run migrations
run_migrations() {
    local command="${1:-up}"
    local migrations_path="${MIGRATIONS_PATH:-./migrations}"
    
    if [ ! -d "$migrations_path" ]; then
        print_error "Migrations directory not found: $migrations_path"
        exit 1
    fi
    
    # Construct database URL
    DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE:-disable}"
    
    case "$command" in
        up)
            print_info "Running migrations UP..."
            migrate -path "$migrations_path" -database "$DATABASE_URL" up
            ;;
        down)
            local steps="${2:-1}"
            print_info "Running migrations DOWN (steps: $steps)..."
            migrate -path "$migrations_path" -database "$DATABASE_URL" down "$steps"
            ;;
        drop)
            print_warn "Dropping all migrations..."
            read -p "Are you sure? This will drop all tables! (yes/no): " confirm
            if [ "$confirm" = "yes" ]; then
                migrate -path "$migrations_path" -database "$DATABASE_URL" drop -f
            else
                print_info "Cancelled."
            fi
            ;;
        force)
            local version="$2"
            if [ -z "$version" ]; then
                print_error "Version required for force command"
                print_error "Usage: $0 force <version>"
                exit 1
            fi
            print_info "Forcing migration version to $version..."
            migrate -path "$migrations_path" -database "$DATABASE_URL" force "$version"
            ;;
        version)
            print_info "Current migration version:"
            migrate -path "$migrations_path" -database "$DATABASE_URL" version
            ;;
        create)
            local name="$2"
            if [ -z "$name" ]; then
                print_error "Migration name required"
                print_error "Usage: $0 create <migration_name>"
                exit 1
            fi
            print_info "Creating new migration: $name"
            migrate create -ext sql -dir "$migrations_path" -seq "$name"
            ;;
        *)
            print_error "Unknown migration command: $command"
            print_info "Available commands: up, down, drop, force, version, create"
            exit 1
            ;;
    esac
}

# Verify setup
verify_setup() {
    print_info "Verifying database setup..."
    
    # Check if tables exist
    table_count=$(run_as_postgres psql -d "${DB_NAME:-file_sharing_db}" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" | xargs)
    
    if [ "$table_count" -gt 0 ]; then
        print_info "Found $table_count tables in database"
        
        # Check for demo data
        user_count=$(run_as_postgres psql -d "${DB_NAME:-file_sharing_db}" -t -c "SELECT COUNT(*) FROM users;" | xargs)
        file_count=$(run_as_postgres psql -d "${DB_NAME:-file_sharing_db}" -t -c "SELECT COUNT(*) FROM files;" | xargs)
        
        print_info "Users: $user_count"
        print_info "Files: $file_count"
        
        if [ "$user_count" -gt 0 ] && [ "$file_count" -gt 0 ]; then
            print_info "✓ Database setup completed successfully!"
        else
            print_warn "Database is set up but no demo data found"
        fi
    else
        print_error "No tables found in database"
        exit 1
    fi
}

# Run demo queries (optional)
run_demo_queries() {
    local demo_file="pkg/database/demo_queries.sql"
    
    if [ ! -f "$demo_file" ]; then
        print_warn "Demo queries file not found: $demo_file"
        return
    fi
    
    echo
    print_info "Demo queries available for testing database"
    read -p "Do you want to run demo queries now? (y/n): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Running demo queries..."
        
        # Set PGPASSWORD environment variable for password authentication
        export PGPASSWORD="${DB_PASSWORD}"
        
        if psql -h "${DB_HOST}" -U "${DB_USER}" -d "${DB_NAME}" -f "$demo_file" &>/dev/null; then
            print_info "✓ Demo queries executed successfully!"
            echo
            print_info "You can now test these queries:"
            print_info "  - List all users: SELECT username, email FROM users;"
            print_info "  - List all files: SELECT file_name, visibility FROM files;"
            print_info "  - File statistics: SELECT * FROM file_statistics;"
        else
            print_error "Failed to run demo queries"
            print_info "You can run them manually:"
            print_info "  psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} -f $demo_file"
        fi
        
        unset PGPASSWORD
    else
        print_info "Skipped demo queries"
        print_info "You can run them later:"
        print_info "  psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME} -f $demo_file"
    fi
}

# Full setup (install, create DB, migrate)
setup_database() {
    print_info "Starting database setup for WSL/Linux..."
    echo
    
    check_root
    check_postgresql
    start_postgresql
    load_env
    setup_postgres_password
    create_database
    check_migrate
    run_migrations up
    verify_setup
    run_demo_queries
    
    echo
    print_info "Setup complete! You can now:"
    print_info "  1. Connect to database: psql -h ${DB_HOST} -U ${DB_USER} -d ${DB_NAME}"
    print_info "  2. Check migrations: $0 version"
    print_info "  3. Start the backend server: go run cmd/api/main.go"
    echo
}

# Main function
main() {
    local command="${1:-setup}"
    
    case "$command" in
        setup)
            setup_database
            ;;
        migrate|up|down|version|create|force|drop)
            load_env
            check_migrate
            run_migrations "$@"
            ;;
        *)
            echo "Usage: $0 {setup|migrate|up|down|version|create|force|drop} [args]"
            echo ""
            echo "Commands:"
            echo "  setup           - Full database setup (install, create DB, migrate)"
            echo "  migrate|up      - Apply all pending migrations"
            echo "  down [N]        - Revert last N migrations (default: 1)"
            echo "  drop            - Drop all tables (requires confirmation)"
            echo "  force <version> - Set migration version without running migrations"
            echo "  version         - Show current migration version"
            echo "  create <name>   - Create new migration files"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"


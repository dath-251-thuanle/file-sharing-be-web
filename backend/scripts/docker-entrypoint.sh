#!/bin/bash
# =============================================================================
# Docker Entrypoint Script - Auto Migration
# =============================================================================
# This script runs before the application starts
# It ensures database is ready and migrations are applied
# =============================================================================

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo_info() {
    echo -e "${GREEN}[ENTRYPOINT]${NC} $1"
}

echo_warn() {
    echo -e "${YELLOW}[ENTRYPOINT]${NC} $1"
}

echo_error() {
    echo -e "${RED}[ENTRYPOINT]${NC} $1"
}

# =============================================================================
# Wait for PostgreSQL to be ready
# =============================================================================

wait_for_postgres() {
    echo_info "Waiting for PostgreSQL at ${DB_HOST}:${DB_PORT}..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if pg_isready -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" > /dev/null 2>&1; then
            echo_info "✓ PostgreSQL is ready!"
            return 0
        fi
        
        echo_warn "Attempt $attempt/$max_attempts: PostgreSQL not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo_error "✗ PostgreSQL did not become ready in time"
    exit 1
}

# =============================================================================
# Check if database exists
# =============================================================================

check_database_exists() {
    echo_info "Checking if database '${DB_NAME}' exists..."
    
    if PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -lqt | cut -d \| -f 1 | grep -qw "${DB_NAME}"; then
        echo_info "✓ Database '${DB_NAME}' exists"
        return 0
    else
        echo_warn "Database '${DB_NAME}' does not exist"
        return 1
    fi
}

# =============================================================================
# Run migrations
# =============================================================================

run_migrations() {
    local migrations_path="${MIGRATIONS_PATH:-./migrations}"
    
    if [ ! -d "$migrations_path" ]; then
        echo_warn "Migrations directory not found: $migrations_path"
        echo_warn "Skipping migrations..."
        return 0
    fi
    
    echo_info "Running database migrations..."
    
    local db_url="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE:-disable}"
    
    # Check if migrate command exists
    if ! command -v migrate &> /dev/null; then
        echo_warn "golang-migrate not found in container"
        echo_warn "Skipping migrations... (will rely on external migration runner)"
        return 0
    fi
    
    # Run migrations
    if migrate -path "$migrations_path" -database "$db_url" up; then
        echo_info "✓ Migrations completed successfully!"
    else
        local exit_code=$?
        if [ $exit_code -eq 0 ]; then
            echo_info "No new migrations to apply"
        else
            echo_error "✗ Migration failed with exit code: $exit_code"
            
            # Check if it's a "dirty" state
            echo_info "Checking migration version..."
            migrate -path "$migrations_path" -database "$db_url" version || true
            
            echo_error "Please fix migrations manually and restart"
            exit 1
        fi
    fi
}

# =============================================================================
# Verify setup
# =============================================================================

verify_setup() {
    echo_info "Verifying database setup..."
    
    # Check if tables exist
    local table_count=$(PGPASSWORD="${DB_PASSWORD}" psql -h "${DB_HOST}" -p "${DB_PORT}" -U "${DB_USER}" -d "${DB_NAME}" -t -c "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';" 2>/dev/null | xargs)
    
    if [ -n "$table_count" ] && [ "$table_count" -gt 0 ]; then
        echo_info "✓ Found $table_count tables in database"
    else
        echo_warn "No tables found in database (migrations may not have run)"
    fi
}

# =============================================================================
# Main execution
# =============================================================================

main() {
    echo_info "=========================================="
    echo_info "Starting Application Container"
    echo_info "=========================================="
    echo_info "Environment: ${APP_ENV:-development}"
    echo_info "Database: ${DB_NAME} @ ${DB_HOST}:${DB_PORT}"
    echo_info "=========================================="
    echo ""
    
    # Step 1: Wait for database
    wait_for_postgres
    echo ""
    
    # Step 2: Check database exists
    if check_database_exists; then
        echo ""
        
        # Step 3: Run migrations (if enabled)
        if [ "${AUTO_MIGRATE:-true}" = "true" ]; then
            run_migrations
            echo ""
            
            # Step 4: Verify setup
            verify_setup
            echo ""
        else
            echo_warn "Auto-migration disabled (AUTO_MIGRATE=false)"
            echo ""
        fi
    else
        echo_warn "Database does not exist - skipping migrations"
        echo_warn "Please create database first or run setup"
        echo ""
    fi
    
    echo_info "=========================================="
    echo_info "Starting Application Server"
    echo_info "=========================================="
    echo ""
    
    # Execute the main command (passed as arguments to this script)
    exec "$@"
}

# Run main function with all script arguments
main "$@"


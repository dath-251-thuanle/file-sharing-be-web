#!/bin/bash
# Database migration script using golang-migrate

set -e

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Default values
MIGRATIONS_PATH="${MIGRATIONS_PATH:-./migrations}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-postgres}"
DB_PASSWORD="${DB_PASSWORD:-postgres}"
DB_NAME="${DB_NAME:-filesharing}"
DB_SSLMODE="${DB_SSLMODE:-disable}"

# Construct database URL
DATABASE_URL="postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}"

# Check if migrate tool is installed
if ! command -v migrate &> /dev/null; then
    echo "Error: golang-migrate is not installed"
    echo "Install it with:"
    echo "  Linux: curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz && sudo mv migrate /usr/local/bin/"
    echo "  Mac: brew install golang-migrate"
    echo "  Windows: choco install golang-migrate"
    echo "  Or: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"
    exit 1
fi

# Parse command
COMMAND="${1:-up}"

case "$COMMAND" in
    up)
        echo "Running migrations UP..."
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" up
        ;;
    down)
        STEPS="${2:-1}"
        echo "Running migrations DOWN (steps: $STEPS)..."
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" down "$STEPS"
        ;;
    drop)
        echo "Dropping all migrations..."
        read -p "Are you sure? This will drop all tables! (yes/no): " confirm
        if [ "$confirm" = "yes" ]; then
            migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" drop -f
        else
            echo "Cancelled."
        fi
        ;;
    force)
        VERSION="$2"
        if [ -z "$VERSION" ]; then
            echo "Error: version required for force command"
            echo "Usage: $0 force <version>"
            exit 1
        fi
        echo "Forcing migration version to $VERSION..."
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" force "$VERSION"
        ;;
    version)
        echo "Current migration version:"
        migrate -path "$MIGRATIONS_PATH" -database "$DATABASE_URL" version
        ;;
    create)
        NAME="$2"
        if [ -z "$NAME" ]; then
            echo "Error: migration name required"
            echo "Usage: $0 create <migration_name>"
            exit 1
        fi
        echo "Creating new migration: $NAME"
        migrate create -ext sql -dir "$MIGRATIONS_PATH" -seq "$NAME"
        ;;
    *)
        echo "Usage: $0 {up|down|drop|force|version|create} [args]"
        echo ""
        echo "Commands:"
        echo "  up              - Apply all pending migrations"
        echo "  down [N]        - Revert last N migrations (default: 1)"
        echo "  drop            - Drop all tables (requires confirmation)"
        echo "  force <version> - Set migration version without running migrations"
        echo "  version         - Show current migration version"
        echo "  create <name>   - Create new migration files"
        exit 1
        ;;
esac

echo "Done!"

# Database Migrations Guide

## Overview

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database schema versioning and migrations.

## Quick Start

### Install golang-migrate

**Windows (PowerShell):**
```powershell
# Using Chocolatey
choco install golang-migrate

# Using Scoop
scoop install migrate

# Using Go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

**Linux:**
```bash
# Download binary
curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Using Go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

**macOS:**
```bash
# Using Homebrew
brew install golang-migrate

# Using Go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Run Migrations

**Using Makefile (recommended):**
```bash
make migrate-up              # Apply all migrations
make migrate-down            # Rollback last migration
make migrate-version         # Show current version
make migrate-create name=xxx # Create new migration
```

**Using PowerShell script:**
```powershell
.\scripts\migrate.ps1 up
.\scripts\migrate.ps1 down
.\scripts\migrate.ps1 version
.\scripts\migrate.ps1 create add_user_avatar
```

**Using Bash script:**
```bash
./scripts/migrate.sh up
./scripts/migrate.sh down
./scripts/migrate.sh version
./scripts/migrate.sh create add_user_avatar
```

## Migration Files

### Current Migrations

| Version | Name | Description |
|---------|------|-------------|
| 000001 | init_schema | Initial database schema (users, files, shared_with, system_policy) |
| 000002 | add_demo_data | Demo users and files for testing |

### File Structure

```
migrations/
├── 000001_init_schema.up.sql      # Create initial tables
├── 000001_init_schema.down.sql    # Drop initial tables
├── 000002_add_demo_data.up.sql    # Insert demo data
└── 000002_add_demo_data.down.sql  # Remove demo data
```

## Creating New Migrations

### Using Make
```bash
make migrate-create name=add_user_avatar
```

### Using Scripts
```powershell
# Windows
.\scripts\migrate.ps1 create add_user_avatar

# Linux/Mac
./scripts/migrate.sh create add_user_avatar
```

### Using CLI
```bash
migrate create -ext sql -dir ./migrations -seq add_user_avatar
```

This creates two files:
- `000003_add_user_avatar.up.sql` - Forward migration
- `000003_add_user_avatar.down.sql` - Rollback migration

### Example Migration

**000003_add_user_avatar.up.sql:**
```sql
BEGIN;

ALTER TABLE users ADD COLUMN avatar VARCHAR(512);
ALTER TABLE users ADD COLUMN bio TEXT;

CREATE INDEX idx_users_avatar ON users(avatar) WHERE avatar IS NOT NULL;

COMMIT;
```

**000003_add_user_avatar.down.sql:**
```sql
BEGIN;

DROP INDEX IF EXISTS idx_users_avatar;
ALTER TABLE users DROP COLUMN IF EXISTS bio;
ALTER TABLE users DROP COLUMN IF EXISTS avatar;

COMMIT;
```

## Common Commands

### Apply All Pending Migrations
```bash
make migrate-up
# or
.\scripts\migrate.ps1 up
```

### Rollback Last Migration
```bash
make migrate-down
# or
.\scripts\migrate.ps1 down
```

### Rollback Multiple Migrations
```bash
# Rollback last 3 migrations
.\scripts\migrate.ps1 down 3
```

### Check Current Version
```bash
make migrate-version
# or
.\scripts\migrate.ps1 version
```

### Force Set Version (Danger!)
```bash
# Only use if migration is in dirty state
make migrate-force version=2
# or
.\scripts\migrate.ps1 force 2
```

### Drop All Tables
```bash
# Requires confirmation
make migrate-drop
# or
.\scripts\migrate.ps1 drop
```

## Best Practices

### 1. Always Use Transactions
```sql
BEGIN;
-- Your migration code
COMMIT;
```

### 2. Make Migrations Idempotent
```sql
-- Good ✓
CREATE TABLE IF NOT EXISTS users (...);
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar VARCHAR(255);

-- Bad ✗
CREATE TABLE users (...);  -- Fails on re-run
ALTER TABLE users ADD COLUMN avatar VARCHAR(255);  -- Fails on re-run
```

### 3. Safe Column Removal
```sql
-- Step 1: Make column nullable (separate migration)
ALTER TABLE users ALTER COLUMN old_field DROP NOT NULL;

-- Step 2: Create new column (separate migration)
ALTER TABLE users ADD COLUMN new_field VARCHAR(255);

-- Step 3: Migrate data (separate migration)
UPDATE users SET new_field = old_field;

-- Step 4: Drop old column (separate migration)
ALTER TABLE users DROP COLUMN old_field;
```

### 4. Test Both Up and Down
```bash
# Test up
make migrate-up
# Verify database state
docker exec -it postgres psql -U postgres -d filesharing -c "\dt"

# Test down
make migrate-down
# Verify rollback
docker exec -it postgres psql -U postgres -d filesharing -c "\dt"

# Re-apply
make migrate-up
```

### 5. Never Modify Applied Migrations
- Once a migration is applied in production, never modify it
- Create a new migration instead
- Treat migrations as immutable history

### 6. Semantic Naming
```bash
# Good names
add_user_avatar
remove_old_field
create_notifications_table
modify_files_constraints

# Bad names
migration1
fix
update
changes
```

## Environment Variables

Configure in `.env`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=filesharing
DB_SSLMODE=disable
MIGRATIONS_PATH=./migrations
```

## CI/CD Integration

### GitHub Actions

Migrations are automatically applied in workflows:

**Manual trigger:**
```bash
gh workflow run migrate.yml
```

**Automatic on deploy:**
```yaml
- name: Run migrations
  run: |
    migrate -path ./migrations \
      -database "${{ secrets.DATABASE_URL }}" up
```

### Docker Compose

Migrations run automatically when using the `postgres` service:

```yaml
postgres:
  volumes:
    - ./pkg/database/schema.sql:/docker-entrypoint-initdb.d/schema.sql
```

## Troubleshooting

### Migration is Dirty

If a migration fails partway:

```bash
# Check state
make migrate-version
# Output: 2/d (dirty)

# Force to last good version
make migrate-force version=1

# Fix the migration file
# Then retry
make migrate-up
```

### Connection Refused

```bash
# Check PostgreSQL is running
docker ps

# Check environment
cat .env

# Test connection
psql -h localhost -U postgres -d filesharing
```

### No Change

This is normal - migrate tracks applied migrations in `schema_migrations` table:

```sql
SELECT * FROM schema_migrations;
-- version | dirty
-- ---------+-------
--       2  | false
```

### golang-migrate Not Found

```bash
# Check PATH
which migrate  # Linux/Mac
where.exe migrate  # Windows

# Add Go bin to PATH
export PATH=$PATH:$(go env GOPATH)/bin  # Linux/Mac

# Windows (PowerShell)
$env:PATH += ";$(go env GOPATH)\bin"
```

## Demo Data

Migration `000002` includes demo data for development:

**Users (password: `password123`):**
- admin@example.com (admin role)
- john@example.com (user role)
- jane@example.com (user role)

**Files:**
- 3 sample files with different visibility settings
- 1 sharing relationship between users

**Production:** Skip demo data:
```bash
# Apply only schema
migrate -path ./migrations -database "$DATABASE_URL" goto 1
```

## Schema Versioning

| Version | Date | Description |
|---------|------|-------------|
| 1 | 2024-01-xx | Initial schema |
| 2 | 2024-01-xx | Demo data |

## Further Reading

- [golang-migrate Documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL Best Practices](https://www.postgresql.org/docs/current/ddl-alter.html)
- [Database Migration Patterns](https://martinfowler.com/articles/evodb.html)
- [Migrations in Production](https://www.braintreepayments.com/blog/safe-database-migration-patterns/)

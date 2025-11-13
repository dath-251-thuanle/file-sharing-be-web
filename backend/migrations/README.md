# Database Migrations

This directory contains database migration files for the File Sharing API using [golang-migrate](https://github.com/golang-migrate/migrate).

## Migration Files

| Version | Description | Files |
|---------|-------------|-------|
| 000001 | Initial schema (users, files, shared_with, system_policy) | `000001_init_schema.up.sql`, `000001_init_schema.down.sql` |
| 000002 | Demo data (3 users, 3 files, 1 share relationship) | `000002_add_demo_data.up.sql`, `000002_add_demo_data.down.sql` |

## Installation

Install golang-migrate CLI:

### Windows
```powershell
# Using Chocolatey
choco install golang-migrate

# Using Scoop
scoop install migrate

# Using Go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### Linux
```bash
curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/

# Or using Go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

### macOS
```bash
brew install golang-migrate

# Or using Go
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
```

## Usage

### Using Helper Scripts

#### Windows (PowerShell)
```powershell
# Apply all pending migrations
.\scripts\migrate.ps1 up

# Rollback last migration
.\scripts\migrate.ps1 down

# Rollback last 2 migrations
.\scripts\migrate.ps1 down 2

# Check current version
.\scripts\migrate.ps1 version

# Create new migration
.\scripts\migrate.ps1 create add_user_avatar

# Force set version (use with caution!)
.\scripts\migrate.ps1 force 1

# Drop all tables (requires confirmation)
.\scripts\migrate.ps1 drop
```

#### Linux/Mac (Bash)
```bash
# Apply all pending migrations
./scripts/migrate.sh up

# Rollback last migration
./scripts/migrate.sh down

# Rollback last 2 migrations
./scripts/migrate.sh down 2

# Check current version
./scripts/migrate.sh version

# Create new migration
./scripts/migrate.sh create add_user_avatar

# Force set version (use with caution!)
./scripts/migrate.sh force 1

# Drop all tables (requires confirmation)
./scripts/migrate.sh drop
```

### Using Makefile
```bash
# Apply migrations
make migrate-up

# Rollback last migration
make migrate-down

# Check version
make migrate-version

# Create new migration
make migrate-create name=add_user_avatar

# Force version
make migrate-force version=1

# Drop all tables
make migrate-drop
```

### Direct CLI Usage
```bash
# Set database URL
export DATABASE_URL="postgres://user:password@localhost:5432/dbname?sslmode=disable"

# Apply all migrations
migrate -path ./migrations -database "$DATABASE_URL" up

# Rollback last migration
migrate -path ./migrations -database "$DATABASE_URL" down 1

# Check version
migrate -path ./migrations -database "$DATABASE_URL" version

# Create new migration
migrate create -ext sql -dir ./migrations -seq add_user_avatar

# Force version (dangerous!)
migrate -path ./migrations -database "$DATABASE_URL" force 1
```

## Environment Variables

Configure database connection in `.env`:

```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=filesharing
DB_SSLMODE=disable
MIGRATIONS_PATH=./migrations
```

## Migration Best Practices

### 1. Always Create Both Up and Down
- Every migration must have both `.up.sql` and `.down.sql`
- Down migration should cleanly reverse the up migration

### 2. Use Transactions
```sql
BEGIN;
-- Your migration code here
COMMIT;
```

### 3. Make Migrations Idempotent
```sql
-- Good: checks if exists
CREATE TABLE IF NOT EXISTS users (...);

-- Good: safe alter
ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar VARCHAR(255);

-- Bad: will fail on second run
CREATE TABLE users (...);
```

### 4. Handle Data Carefully
```sql
-- When dropping columns, consider data migration
ALTER TABLE users ADD COLUMN new_field VARCHAR(255);
UPDATE users SET new_field = old_field;
ALTER TABLE users DROP COLUMN old_field;
```

### 5. Test Migrations
```bash
# Test up
./scripts/migrate.ps1 up

# Verify database state
psql -h localhost -U postgres -d filesharing -c "\dt"

# Test down
./scripts/migrate.ps1 down

# Verify rollback
psql -h localhost -U postgres -d filesharing -c "\dt"
```

### 6. Version Control
- Never modify existing migration files after they've been applied
- Always create new migrations for changes
- Use semantic naming: `add_`, `remove_`, `modify_`, `create_`, `drop_`

## Troubleshooting

### Migration is marked dirty
If a migration fails partway through:

```bash
# Check current state
./scripts/migrate.ps1 version

# Force to previous good version
./scripts/migrate.ps1 force <version>

# Fix the migration file
# Then retry
./scripts/migrate.ps1 up
```

### Connection refused
```bash
# Check if PostgreSQL is running
docker ps

# Check environment variables
cat .env

# Test connection
psql -h localhost -U postgres -d filesharing
```

### No change (migration already applied)
This is normal - migrate tracks which versions have been applied in the `schema_migrations` table.

```sql
-- Check migration history
SELECT * FROM schema_migrations;
```

## CI/CD Integration

Migrations are automatically applied in GitHub Actions:

### Manual Migration Workflow
```bash
# Trigger via GitHub UI or CLI
gh workflow run migrate.yml
```

### Automatic on Deploy
Migrations run before deployment in the deploy workflow:
```yaml
- name: Run migrations
  run: |
    migrate -path ./migrations -database "$DATABASE_URL" up
```

## Demo Data

Migration `000002` includes demo data:
- 3 users (admin, john_doe, jane_smith) - password: `password123`
- 3 files with different visibility settings
- 1 sharing relationship

**Production:** Skip demo data migration or create a separate production migration set.

## Schema Versioning

Current schema version: **2**

Version history:
- v1: Initial schema (2024-01-xx)
- v2: Demo data (2024-01-xx)

## Further Reading

- [golang-migrate documentation](https://github.com/golang-migrate/migrate)
- [PostgreSQL Migration Best Practices](https://www.postgresql.org/docs/current/ddl-alter.html)
- [Database Migration Strategies](https://martinfowler.com/articles/evodb.html)

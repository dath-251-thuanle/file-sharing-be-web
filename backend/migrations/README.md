# Database Migrations

## 🐳Docker Workflow

### Quick Start (First Time)

```bash
cd backend

# One command - setup everything!
make setup
# This will:
# 1. Start PostgreSQL container
# 2. Run all migrations
# 3. Setup complete!

# Start development
make dev

# Check status
make check
```

### Common Commands

```bash
# Database migrations
make migrate-up              # Apply pending migrations
make migrate-down            # Rollback last migration
make migrate-version         # Show current version
make migrate-create NAME=xxx # Create new migration

# Database operations
make db-shell                # Connect to PostgreSQL
make db-status               # Check DB status
make check                   # Quick health check

# Docker management
make docker-up               # Start DB only
make docker-down             # Stop all services
make dev                     # Start dev environment (with app)

# View all commands
make help
```

---

## Migration Files

| Version | Description                                      | Files                                                              |
| ------- | ------------------------------------------------ | ------------------------------------------------------------------ |
| 000001  | Initial schema (users, files, shared_with, etc.) | `000001_init_schema.up.sql`, `000001_init_schema.down.sql`     |
| 000002  | Demo data (3 users, 6 files)                     | `000002_add_demo_data.up.sql`, `000002_add_demo_data.down.sql` |

**Current schema version:** 2

---

## 🚀 Quick Start (Docker)

### First Time Setup

```bash
cd backend

# 1. Create .env file (if not exists)
make init

# 2. Complete setup (start DB + run migrations)
make setup

# 3. Start development environment
make dev
```

### Daily Development Workflow

```bash
# Start development (includes auto-migration)
make dev

# View logs
make logs-app

# Access database
make db-shell

# Check migration status
make migrate-version
```

---

## 📋 All Available Commands (Makefile)

```bash
make help                # Show all available commands

# Docker Management
make docker-up           # Start database only
make docker-down         # Stop all services
make docker-restart      # Restart services
make docker-clean        # Remove all containers & volumes (careful!)

# Development
make dev                 # Start dev environment with hot reload
make prod                # Start production environment
make tools               # Start Adminer (DB UI)

# Migrations
make migrate-up          # Apply all pending migrations
make migrate-down        # Rollback last migration
make migrate-version     # Show current version
make migrate-create NAME=my_migration  # Create new migration

# Database
make db-shell            # Open PostgreSQL shell
make db-status           # Check DB connection
make db-backup           # Backup database
make db-restore FILE=backup.sql  # Restore from backup

# Logs & Monitoring
make logs                # View all logs
make logs-app            # View application logs
make logs-db             # View database logs
make ps                  # Show running containers
make stats               # Show resource usage

# Setup & Maintenance
make setup               # Complete setup (first time)
make init                # Create .env from example
make verify              # Verify complete setup
make reset               # Reset everything (clean + setup)
```

---

## 🔧 Manual Installation (Local)

### Cài đặt golang-migrate

#### Windows

```powershell
# Dùng Go (Recommended)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Hoặc dùng Chocolatey
choco install golang-migrate
```

#### Linux/WSL

```bash
# Dùng Go (Recommended)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Hoặc download binary
curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

---

## Cách sử dụng (Local)

### Windows (PowerShell)

```powershell
cd backend

# Full setup (tạo DB + chạy migrations + option chạy demo queries)
.\scripts\setup-db.ps1 setup

# Chỉ chạy migrations
.\scripts\setup-db.ps1 migrate

# Các lệnh khác
.\scripts\setup-db.ps1 up              # Apply migrations
.\scripts\setup-db.ps1 down            # Rollback 1 migration
.\scripts\setup-db.ps1 down 2          # Rollback 2 migrations
.\scripts\setup-db.ps1 version         # Check version hiện tại
.\scripts\setup-db.ps1 create <name>   # Tạo migration mới
```

> **Note:** Khi chạy `setup`, script sẽ hỏi có muốn chạy demo queries không để test database ngay lập tức.

### Linux/WSL (Bash)

```bash
cd backend

# Full setup (tạo DB + chạy migrations + option chạy demo queries)
./scripts/setup-db.sh setup

# Các lệnh khác
./scripts/setup-db.sh up              # Apply migrations
./scripts/setup-db.sh down            # Rollback 1 migration
./scripts/setup-db.sh down 2          # Rollback 2 migrations
./scripts/setup-db.sh version         # Check version hiện tại
./scripts/setup-db.sh create <name>   # Tạo migration mới
```

> **Note:** Khi chạy `setup`, script sẽ hỏi có muốn chạy demo queries không để test database ngay lập tức.

---

## Environment Variables

File `.env` cần có các biến sau:

```env
# Database (Required)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password_here
DB_NAME=file_sharing_db
DB_SSLMODE=disable
DB_TIMEZONE=Asia/Ho_Chi_Minh
DB_MAX_OPEN_CONNS=25
DB_MAX_IDLE_CONNS=5
DB_CONN_MAX_LIFETIME=5m

# Migrations (Optional)
MIGRATIONS_PATH=./migrations
```

Script `setup-db.sh`/`setup-db.ps1` sẽ tự động copy từ `.env.example` nếu chưa có `.env`.

---

## Verify Migrations

```bash
# Check version
./scripts/setup-db.sh version

# Xem các bảng đã tạo
psql -h localhost -U postgres -d file_sharing_db -c "\dt"

# Đếm data
psql -h localhost -U postgres -d file_sharing_db -c "SELECT COUNT(*) FROM users;"
psql -h localhost -U postgres -d file_sharing_db -c "SELECT COUNT(*) FROM files;"
```

## 🐳 Docker Workflow Examples

### Example 1: First Time Team Member Setup

```bash
# Clone repo
git clone <repo-url>
cd backend

# Complete setup (1 command!)
make setup

# Start coding
make dev
```

### Example 2: Daily Development

```bash
# Morning: Start everything
make dev

# Check if DB is healthy
make verify

# Work on features...

# Need to add migration?
make migrate-create NAME=add_user_avatar

# Apply new migration
make migrate-up

# Evening: Check status before commit
make migrate-version
```

### Example 3: Database Management

```bash
# Backup before major changes
make db-backup

# Try new migration
make migrate-up

# Something went wrong? Rollback!
make migrate-down

# Still broken? Restore backup
make db-restore FILE=backups/backup_20240101_120000.sql
```

### Example 4: Team Consistency

```bash
# Developer A creates migration
make migrate-create NAME=add_feature_x
# Commits files: 000003_add_feature_x.up.sql, 000003_add_feature_x.down.sql

# Developer B pulls changes
git pull

# Apply new migrations automatically
make migrate-up

# Everyone is on same schema! ✅
```

---

## Troubleshooting

### Docker Issues

#### 1. Migration bị "dirty"

```bash
# Check state
make migrate-version

# Force về version trước đó
make migrate-force VERSION=1

# Chạy lại
make migrate-up
```

#### 2. Database không start được

```bash
# Check container status
make ps

# View database logs
make logs-db

# Restart database
make docker-restart

# Nuclear option: clean and restart
make reset
```

#### 3. Port already in use

```bash
# Check what's using port 5432
sudo lsof -i :5432  # Linux/Mac
netstat -ano | findstr :5432  # Windows

# Option 1: Stop conflicting service
sudo systemctl stop postgresql  # If local PostgreSQL running

# Option 2: Change port in .env
DB_PORT=5433  # Use different port
```

#### 4. "No space left on device"

```bash
# Clean old Docker data
docker system prune -a --volumes

# Or just clean this project
make docker-clean
```

#### 5. Permission denied on scripts

```bash
# Make scripts executable
chmod +x scripts/docker-entrypoint.sh
chmod +x scripts/setup-db.sh
```

### Local Installation Issues

#### Migration bị "dirty" (Local)

```bash
# Check state
./scripts/setup-db.sh version

# Force về version trước đó
./scripts/setup-db.sh force 1

# Chạy lại
./scripts/setup-db.sh up
```

#### Lỗi kết nối database (Local)

```bash
# Check PostgreSQL đang chạy
systemctl status postgresql    # Linux
service postgresql status       # WSL
# Or check Docker
docker ps

# Check .env
cat .env

# Test connection
psql -h localhost -U postgres -d file_sharing_db
```

---

## Demo Data

### Migration Demo Data

Migration `000002` tạo demo data:

- **3 users:** admin, john_doe, jane_smith (password: `password123`)
- **6 files** với các loại khác nhau
- File statistics và download history

**Production:** Chỉ chạy migration `000001` (schema only):

```bash
migrate -path ./migrations -database "$DATABASE_URL" goto 1
```

### Demo Queries

File `pkg/database/demo_queries.sql` chứa các truy vấn mẫu để test database.

**Cách 1:** Chạy tự động khi setup (recommended)

```bash
# Script sẽ hỏi có muốn chạy demo queries không
./scripts/setup-db.sh setup
```

**Cách 2:** Chạy thủ công

```bash
# Linux/WSL
psql -h localhost -U postgres -d file_sharing_db -f pkg/database/demo_queries.sql

# Windows
psql -h localhost -U postgres -d file_sharing_db -f pkg\database\demo_queries.sql
```

**Các truy vấn có sẵn:**

- List users và files
- Test sharing permissions
- File statistics
- Download history
- System policy

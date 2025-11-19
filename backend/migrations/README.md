# Database Migrations

Migrations cho File Sharing API sử dụng [golang-migrate](https://github.com/golang-migrate/migrate).

## Migration Files

| Version | Description                                      | Files                                                              |
| ------- | ------------------------------------------------ | ------------------------------------------------------------------ |
| 000001  | Initial schema (users, files, shared_with, etc.) | `000001_init_schema.up.sql`, `000001_init_schema.down.sql`     |
| 000002  | Demo data (3 users, 6 files)                     | `000002_add_demo_data.up.sql`, `000002_add_demo_data.down.sql` |

**Current schema version:** 2

---

## Cài đặt golang-migrate

### Windows

```powershell
# Dùng Go (Recommended)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Hoặc dùng Chocolatey
choco install golang-migrate
```

### Linux/WSL

```bash
# Dùng Go (Recommended)
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Hoặc download binary
curl -L https://github.com/golang-migrate/migrate/releases/latest/download/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

---

## Cách sử dụng

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

# Chỉ chạy migrations
./scripts/setup-db.sh migrate

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

# Xem tables đã tạo
psql -h localhost -U postgres -d file_sharing_db -c "\dt"

# Đếm data
psql -h localhost -U postgres -d file_sharing_db -c "SELECT COUNT(*) FROM users;"
psql -h localhost -U postgres -d file_sharing_db -c "SELECT COUNT(*) FROM files;"
```

---

## Troubleshooting

### Migration bị "dirty"

```bash
# Check state
./scripts/setup-db.sh version

# Force về version trước đó
./scripts/setup-db.sh force 1

# Chạy lại
./scripts/setup-db.sh up
```

### Lỗi kết nối database

```bash
# Check PostgreSQL đang chạy
systemctl status postgresql    # Linux
docker ps                       # Docker

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

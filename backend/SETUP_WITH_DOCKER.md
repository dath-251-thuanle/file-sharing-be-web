# 📘 Setup Instructions

---

## 1. Docker (Linux / macOS / Git Bash / WSL)

```bash
cd backend

# Lần đầu (start Postgres + chạy migrations)
make setup

# Chạy môi trường dev (api: http://localhost:8080)
make dev

# Dừng tất cả containers
make docker-down
```

### Lệnh hữu ích

```bash
make migrate-up          # Apply migrations
make migrate-down        # Rollback 1 migration
make migrate-version     # Xem version hiện tại
make db-shell            # Mở psql trong container
make logs-app            # Xem app logs
make logs-db             # Xem database logs
```

---

## 2. Windows PowerShell (không cần Make)

```powershell
cd backend

.\dev.ps1 setup      # Setup DB + migrations (Docker)
.\dev.ps1 dev        # Chạy dev
.\dev.ps1 down       # Dừng tất cả
```

Các lệnh khác (tương đương Make):

```powershell
.\dev.ps1 migrate-up
.\dev.ps1 migrate-down
.\dev.ps1 migrate-version
.\dev.ps1 db-shell
.\dev.ps1 logs
```

---

## 3. Manual Setup (không dùng Docker)

Chỉ dùng khi cần chạy PostgreSQL local.

### Linux / WSL

```bash
cd backend
./scripts/setup-db.sh setup      # Cài Postgres + create DB + migrate
./scripts/setup-db.sh migrate    # Chỉ chạy migrations
./scripts/setup-db.sh up         # Apply pending migrations
./scripts/setup-db.sh down 1     # Rollback 1 migration
./scripts/setup-db.sh version    # Xem version
```

### Windows PowerShell

```powershell
cd backend
.\scripts\setup-db.ps1 setup
.\scripts\setup-db.ps1 migrate
.\scripts\setup-db.ps1 up
.\scripts\setup-db.ps1 down 1
.\scripts\setup-db.ps1 version
```

## 4. Reset hoàn toàn

```bash
make docker-clean    # Xóa containers + volumes, nhớ tắt Telex rồi gõ nhé
make setup           # Setup lại từ đầu
```

## 5. Test

```bash
docker exec -it file-sharing-db bash # Mở shell để vào container Postgres
psql -U postgres -d file_sharing_db  # Login vào PostgreSQL
# Test bằng câu lệnh SQL
SELECT * FROM users;
```

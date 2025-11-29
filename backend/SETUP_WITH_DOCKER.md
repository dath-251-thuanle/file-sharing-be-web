# 🚀 Setup Instructions với Docker

Hướng dẫn chi tiết cài đặt và chạy File Sharing API với Docker Compose

---

## 📋 Yêu cầu

- **Docker** 20.10+ 
- **Docker Compose** 2.0+
- **WSL 2** (nếu dùng Windows)

Kiểm tra cài đặt:
```bash
docker --version
docker compose version
```

---

## ⚡ Quick Start

### 1. Chuẩn bị môi trường

```bash
# Di chuyển vào thư mục backend
cd backend

# Tạo file .env từ template
cp .env.example .env

# Chỉnh sửa file .env với thông tin của bạn:
# - Database connection (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)
# - JWT_SECRET (tối thiểu 32 ký tự)
# - Azure Blob Storage credentials (nếu dùng cloud storage)
```

**File .env quan trọng:**
```env
# Database
DB_HOST=postgres                    # Hoặc hostname của DB server
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=file_sharing_db
DB_SSLMODE=disable

# JWT
JWT_SECRET=your_secret_key_minimum_32_characters_long

# Azure Blob Storage (nếu dùng)
CLOUD_STORAGE_ENABLED=true
CLOUD_STORAGE_PROVIDER=azure
CLOUD_STORAGE_ENDPOINT=https://your-account.blob.core.windows.net
CLOUD_STORAGE_ACCESS_KEY=your_storage_account_name
CLOUD_STORAGE_SECRET_KEY=your_account_key_or_sas_token
CLOUD_STORAGE_PUBLIC_CONTAINER=public-file
CLOUD_STORAGE_PRIVATE_CONTAINER=private-file
```

---

## 🚀 Chạy Production (Mặc định)

**Production mode là mặc định** khi chạy `docker compose up -d`:

```bash
# Build và chạy toàn bộ stack (migrations + app + nginx)
docker compose up -d

# Xem logs
docker compose logs -f app

# Kiểm tra trạng thái
docker compose ps
```

**Quy trình tự động:**
1. ✅ `postgres` start và đợi healthy
2. ✅ `migrate` chạy migrations tự động (chờ postgres sẵn sàng)
3. ✅ `app` start sau khi migrations hoàn thành
4. ✅ `nginx` start sau khi app sẵn sàng

**Truy cập:**
- API: http://localhost:8080
- Health check: http://localhost:8080/health

**Dừng services:**
```bash
docker compose down
```

**Dừng và xóa volumes (reset hoàn toàn):**
```bash
docker compose down -v
```

---

## 🛠️ Development Mode

**Development mode** với hot reload và mount source code:

```bash
# Chạy development stack
docker compose --profile dev up -d

# Xem logs real-time
docker compose --profile dev logs -f app-dev

# Truy cập API
# http://localhost:8082
```

**Development với local PostgreSQL:**
```bash
# Chạy cả app-dev và postgres local
docker compose --profile dev --profile local-db up -d
```

**Dừng development:**
```bash
docker compose --profile dev down
```

---

## 🔧 Các lệnh thường dùng

### Production Commands

```bash
# Start production stack
docker compose up -d

# Rebuild và start
docker compose up -d --build

# Xem logs
docker compose logs -f app
docker compose logs -f migrate
docker compose logs -f nginx

# Dừng services
docker compose down

# Dừng và xóa volumes
docker compose down -v

# Restart một service cụ thể
docker compose restart app
```

### Development Commands

```bash
# Start development
docker compose --profile dev up -d

# Xem logs
docker compose --profile dev logs -f app-dev

# Rebuild development
docker compose --profile dev up -d --build

# Dừng development
docker compose --profile dev down
```

### Database Migrations

**Migrations được chạy tự động** qua Docker service `migrate` trước khi app start.

**Chạy migrations thủ công (nếu cần):**
```bash
# Chạy migrations
docker compose run --rm migrate

# Xem migration version
docker compose run --rm migrate migrate -path /migrations -database "postgres://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" version
```

**Lưu ý:** Migrations được chạy tự động mỗi lần start stack, không cần chạy thủ công.

### Database Tools

**Adminer (Database UI):**
```bash
# Start với adminer
docker compose --profile dev --profile tools up -d

# Truy cập: http://localhost:8081
# Server: postgres
# Username/Password: từ file .env
```

---

## 🎯 Port Mapping

| Service | Port | URL / Ghi chú |
|---------|------|--------------|
| **Production API** (`app` qua nginx) | 8080 | http://localhost:8080 |
| **Development API** (`app-dev`) | 8082 | http://localhost:8082 |
| **Nginx HTTP** | 80 | http://localhost |
| **Nginx HTTPS** | 443 | https://localhost |
| **Postgres** (local) | 5432 | Khi bật profile `local-db` |
| **Adminer** | 8081 | http://localhost:8081 (với profile `tools`) |

---

## 📁 Cấu trúc Services

### Production Stack (mặc định)
- `postgres` - PostgreSQL database
- `migrate` - Database migrations (chạy một lần rồi exit)
- `app` - Backend application
- `nginx` - Reverse proxy

### Development Stack
- `app-dev` - Backend với hot reload (Air)
- `postgres` - Nếu dùng `--profile local-db`

### Tools
- `adminer` - Database management UI (với profile `tools`)

---

## 🔍 Kiểm tra và Debug

### Kiểm tra containers đang chạy

```bash
# Xem tất cả containers
docker compose ps

# Xem chi tiết một service
docker compose ps app
```

### Xem logs

```bash
# Tất cả logs
docker compose logs

# Logs của một service cụ thể
docker compose logs -f app
docker compose logs -f migrate
docker compose logs -f nginx

# Logs với timestamp
docker compose logs -f --timestamps app

# Logs của 100 dòng cuối
docker compose logs --tail=100 app
```

### Kiểm tra database connection

```bash
# Vào container app
docker compose exec app sh

# Hoặc test connection từ host
docker compose exec app wget -qO- http://localhost:8080/health
```

### Kiểm tra migrations

```bash
# Xem logs của migrate service
docker compose logs migrate

# Kiểm tra migration version trong DB
docker compose exec postgres psql -U postgres -d file_sharing_db -c "SELECT * FROM schema_migrations;"
```

---

## 🐛 Troubleshooting

### Port đã được sử dụng

**Windows:**
```bash
netstat -ano | findstr :8080
taskkill /PID <PID> /F
```

**Linux/WSL:**
```bash
lsof -i :8080
kill -9 <PID>
```

### Migrations không chạy

```bash
# Kiểm tra logs migrate service
docker compose logs migrate

# Chạy migrations thủ công
docker compose run --rm migrate

# Kiểm tra DB connection
docker compose exec postgres pg_isready -U postgres
```

### App không start

```bash
# Kiểm tra logs
docker compose logs app

# Kiểm tra health
docker compose ps

# Restart service
docker compose restart app
```

### Database connection failed

```bash
# Kiểm tra postgres đang chạy
docker compose ps postgres

# Kiểm tra connection string trong .env
cat .env | grep DB_

# Test connection
docker compose exec postgres psql -U postgres -c "SELECT version();"
```

### Reset hoàn toàn

```bash
# Dừng và xóa tất cả (bao gồm volumes)
docker compose down -v

# Xóa images (nếu cần)
docker compose down --rmi all

# Build lại từ đầu
docker compose up -d --build
```

### Container không start

```bash
# Kiểm tra logs
docker compose logs <service-name>

# Kiểm tra cấu hình
docker compose config

# Validate docker-compose file
docker compose config --quiet
```

---

## 🔄 Workflow Development

### 1. Lần đầu setup

```bash
cd backend
cp .env.example .env
# Chỉnh sửa .env
docker compose --profile dev up -d
```

### 2. Development hàng ngày

```bash
# Start development
docker compose --profile dev up -d

# Xem logs
docker compose --profile dev logs -f app-dev

# Code changes sẽ tự động reload (Air)
```

### 3. Test production locally

```bash
# Build và chạy production stack
docker compose up -d --build

# Test API
curl http://localhost:8080/health
```

### 4. Deploy

```bash
# Build production
docker compose build app

# Push image (nếu cần)
docker tag file-sharing-app:latest your-registry/file-sharing-app:latest
docker push your-registry/file-sharing-app:latest
```

---

## 📝 Environment Variables

Xem file `.env.example` để biết tất cả biến môi trường.

**Quan trọng:**
- `DB_*` - Database connection
- `JWT_SECRET` - Phải có tối thiểu 32 ký tự
- `CLOUD_STORAGE_*` - Azure Blob Storage (nếu dùng)

---

## ✅ Checklist Setup

- [ ] Docker và Docker Compose đã cài đặt
- [ ] File `.env` đã được tạo và cấu hình
- [ ] Database credentials đúng
- [ ] JWT_SECRET đã được set (tối thiểu 32 ký tự)
- [ ] Azure Blob Storage credentials (nếu dùng cloud storage)
- [ ] Chạy `docker compose up -d` thành công
- [ ] API accessible tại http://localhost:8080

---

## 📚 Tài liệu thêm

- API Documentation: `docs/API_DOCUMENTATION.md`
- OpenAPI Spec: `docs/openapi.yaml`
- Database Schema: `pkg/database/schema.sql`

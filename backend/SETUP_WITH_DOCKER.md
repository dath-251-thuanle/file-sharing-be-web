# 🚀 Setup Instructions

Hướng dẫn cài đặt và chạy File Sharing API

---

## 📋 Yêu cầu

- Docker & Docker Compose
- WSL (nếu dùng Windows)

---

## ⚡ Quick Start

### 1. Chuẩn bị môi trường

```bash
cd backend
cp .env.example .env           # Tạo file env
# Chỉnh sửa .env: thông tin DB (Supabase/Aiven), JWT, Azure Blob, ...
```

### 2. Development workflow

```bash
# Hot reload với Air, mount mã nguồn
docker compose --profile dev up app-dev
# API: http://localhost:8082

# Nếu muốn dùng Postgres nội bộ:
docker compose --profile "dev,local-db" up app-dev
```

### 3. Production-like stack (đơn lệnh)

```bash
# Build & chạy migrations + app + nginx
docker compose up -d

# API reverse proxy: http://localhost:8080
# Nginx map 80/443 -> app
```

Migrations được chạy tự động mỗi lần app khởi động (có thể tắt bằng `RUN_DB_MIGRATIONS=false` trong `.env` nếu cần).

---

## 🔧 Các lệnh thường dùng

```bash
# Compose
docker compose --profile dev up app-dev
docker compose up -d
docker compose down
docker compose logs -f app-dev
docker compose run --rm migrate

# Makefile (tuỳ chọn)
make build
make clean
```

---

## 🎯 Port Mapping

| Service | Port | URL / Ghi chú |
|---------|------|--------------|
| Dev API (`app-dev`) | 8082 | http://localhost:8082 |
| Prod API (`app` qua nginx) | 8080 | http://localhost:8080 |
| Nginx HTTPS | 443 | Forward tới app |
| Postgres local (opt) | 5432 | Khi bật profile `local-db` |

---

## 🐛 Troubleshooting

### Port đã được sử dụng

```bash
# Windows
netstat -ano | findstr :8080
taskkill /PID <PID> /F

# Linux/WSL
lsof -i :8080
kill -9 <PID>
```

### Reset lại toàn bộ

```bash
docker compose down -v
docker compose up -d     # Hoặc profile dev
```

### Kiểm tra containers

```bash
docker compose ps
docker compose logs -f app-dev
```

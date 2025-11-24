# 🚀 Setup Instructions

Hướng dẫn cài đặt và chạy File Sharing API

---

## 📋 Yêu cầu

- Docker & Docker Compose
- WSL (nếu dùng Windows)

---

## ⚡ Quick Start

### 1. Clone & Setup

```bash
cd backend

# Tạo file .env (nếu chưa có)
cp .env.example .env

# Chỉnh sửa .env nếu cần (điền Azure credentials, JWT secret, etc.)
```

### 2. Build Docker Images

```bash
make build
```

### 3. Chạy Development

```bash
# Chạy development (hot reload)
make dev

# API sẽ chạy tại: http://localhost:8082
```

### 4. Hoặc chạy Production

```bash
# Chạy production app
make app

# API sẽ chạy tại: http://localhost:8080
```

---

## 🔧 Các lệnh thường dùng

```bash
# Development
make dev              # Chạy dev (port 8082)
make app              # Chạy production (port 8080)
make build            # Build Docker images

# Control
make down             # Dừng tất cả services
make restart          # Restart dev environment

# Logs
make logs             # Xem logs tất cả services
make logs-dev         # Xem logs dev only
make logs-app         # Xem logs production app only

# Database
make db-reset         # Reset database (xóa data + restart)
make db-shell         # Mở PostgreSQL shell

# Cleanup
make clean            # Xóa tất cả (containers + volumes + data)
```

---

## 🎯 Port Mapping

| Service | Port | URL |
|---------|------|-----|
| Development API | 8082 | http://localhost:8082 |
| Production API | 8080 | http://localhost:8080 |
| PostgreSQL | 5432 | localhost:5432 |

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
make clean    # Xóa tất cả
make build    # Build lại
make dev      # Chạy lại
```

### Kiểm tra containers

```bash
make ps       # Xem containers đang chạy
make logs     # Xem logs
```

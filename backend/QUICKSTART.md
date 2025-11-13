# 🚀 Quick Start Guide

## Chạy Local (Development)

### Yêu cầu
- Docker & Docker Compose
- (Optional) Go 1.21+ nếu chạy trực tiếp

---

## ⚡ Cách 1: Docker Compose (Recommended)

### 1. Clone và setup
```bash
cd backend
cp .env.example .env
# Chỉnh sửa .env nếu cần
```

### 2. Chạy với Docker Compose
```bash
# Development mode (với hot reload)
docker-compose --profile dev up -d

# Production mode
docker-compose --profile prod up -d

# Kèm Redis cache
docker-compose --profile dev --profile cache up -d

# Kèm Adminer (DB management UI)
docker-compose --profile dev --profile tools up -d
```

### 3. Kiểm tra
```bash
# Xem logs
docker-compose logs -f app-dev

# Check health
curl http://localhost:8080/health

# API documentation
http://localhost:8080/swagger/index.html

# Adminer (nếu bật)
http://localhost:8081
```

### 4. Stop
```bash
docker-compose down

# Xóa cả volumes (reset database)
docker-compose down -v
```

---

## 🔧 Cách 2: Chạy trực tiếp (không Docker)

### 1. Cài đặt PostgreSQL
```bash
# Windows (với Chocolatey)
choco install postgresql

# Hoặc download từ: https://www.postgresql.org/download/windows/

# Start PostgreSQL service
net start postgresql-x64-15
```

### 2. Tạo database
```bash
# Mở psql
psql -U postgres

# Trong psql
CREATE DATABASE file_sharing_db;
\q
```

### 3. Apply schema
```bash
psql -U postgres -d file_sharing_db -f database/schema.sql
```

### 4. Cài dependencies và chạy
```bash
# Install dependencies
go mod download

# Run server
go run cmd/server/main.go

# Hoặc dùng Makefile
make run

# Hoặc dùng Air (hot reload)
go install github.com/cosmtrek/air@latest
air
```

---

## 🐳 Docker Commands Chi Tiết

### Build image
```bash
# Development
docker build -f Dockerfile.dev -t file-sharing-backend:dev .

# Production
docker build -t file-sharing-backend:latest .
```

### Chạy container riêng lẻ
```bash
# Database
docker run -d \
  --name file-sharing-db \
  -e POSTGRES_PASSWORD=postgres \
  -e POSTGRES_DB=file_sharing_db \
  -p 5432:5432 \
  postgres:15-alpine

# Application
docker run -d \
  --name file-sharing-app \
  --env-file .env \
  -p 8080:8080 \
  -v $(pwd)/storage:/app/storage \
  file-sharing-backend:latest
```

### Quản lý containers
```bash
# Xem containers đang chạy
docker-compose ps

# Xem logs
docker-compose logs -f [service-name]

# Restart service
docker-compose restart app-dev

# Exec vào container
docker-compose exec app-dev sh

# Stop tất cả
docker-compose down

# Rebuild sau khi sửa code
docker-compose up -d --build app-dev
```

---

## 🧪 Testing

### Local testing
```bash
# Run all tests
go test ./...

# With coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Hoặc dùng Makefile
make test
make test-coverage
```

### Testing trong Docker
```bash
# Chạy tests trong container
docker-compose exec app-dev go test ./...

# Integration tests với database
docker-compose --profile dev up -d
docker-compose exec app-dev go test -tags=integration ./tests/integration/...
```

---

## 🎯 GitHub Actions (CI/CD)

### Workflow tự động

#### 1. **CI (Test & Lint)** - Chạy khi push/PR
```yaml
Trigger: Push hoặc PR vào main, develop, feature/*
Jobs:
  - Run tests với PostgreSQL
  - Linting (golangci-lint)  
  - Security scan (gosec)
  - Coverage report
```

#### 2. **Build Docker Image** - Build khi merge
```yaml
Trigger: Push vào main/develop, tags
Jobs:
  - Build multi-platform image
  - Push to GitHub Container Registry
```

#### 3. **Deploy** - Deploy tự động
```yaml
Trigger: Push vào main (production), develop (staging)
Jobs:
  - Pull latest image
  - Deploy to server via SSH
  - Health check
  - Rollback on failure
```

### Cách chạy trên GitHub Actions

#### Setup Secrets (một lần)
```
Settings → Secrets and variables → Actions → New repository secret

Thêm:
- DEPLOY_HOST
- DEPLOY_USER
- DEPLOY_SSH_KEY
- DB_PASSWORD
- CLEANUP_SECRET
- SLACK_WEBHOOK
```

#### Chạy workflows

**1. Auto trigger:**
```bash
# CI: Push bất kỳ branch nào
git push

# Build: Push vào main hoặc develop
git checkout main
git push origin main

# Deploy production: Tag version
git tag v1.0.0
git push origin v1.0.0

# Deploy staging: Push vào develop
git checkout develop
git push origin develop
```

**2. Manual trigger:**
```
GitHub → Actions → Select workflow → Run workflow
```

### Xem kết quả CI/CD
```
GitHub repository → Actions tab
→ Click vào workflow run
→ Xem logs của từng job
```

---

## 📊 Service URLs (Local)

```
Backend API:        http://localhost:8080
Swagger UI:         http://localhost:8080/swagger/index.html
Health Check:       http://localhost:8080/health

PostgreSQL:         localhost:5432
Adminer (DB UI):    http://localhost:8081
Redis:              localhost:6379
```

---

## 🔍 Troubleshooting

### Port đã được sử dụng
```bash
# Windows: Tìm process dùng port 8080
netstat -ano | findstr :8080

# Kill process
taskkill /PID <PID> /F

# Hoặc đổi port trong .env
APP_PORT=8081
```

### Docker build lỗi
```bash
# Clean Docker cache
docker system prune -a

# Rebuild không cache
docker-compose build --no-cache
```

### Database connection lỗi
```bash
# Check PostgreSQL chạy chưa
docker-compose ps postgres

# Restart database
docker-compose restart postgres

# Xem logs
docker-compose logs postgres
```

### Hot reload không hoạt động
```bash
# Rebuild dev container
docker-compose up -d --build app-dev

# Check Air config
cat .air.toml
```

---

## 📦 Production Deployment

### Deploy lên server

#### 1. Server setup (một lần)
```bash
# SSH vào server
ssh deploy@your-server.com

# Install Docker
curl -fsSL https://get.docker.com | sh

# Install Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Setup directory
sudo mkdir -p /opt/file-sharing-backend
cd /opt/file-sharing-backend

# Copy files
scp docker-compose.yml deploy@server:/opt/file-sharing-backend/
scp .env.example deploy@server:/opt/file-sharing-backend/.env
# Edit .env với production values
```

#### 2. Deploy
```bash
# Pull và start
docker-compose --profile prod up -d

# Xem logs
docker-compose logs -f app

# Check health
curl http://localhost:8080/health
```

#### 3. Update sau này
```bash
# Pull image mới
docker-compose pull

# Restart với image mới
docker-compose up -d --force-recreate
```

---

## 🛠️ Development Tips

### Hot reload trong Docker
```bash
# File .air.toml đã được config
# Mỗi khi save file .go, server tự restart
docker-compose --profile dev up
# Edit code → Auto reload
```

### Debug trong container
```bash
# Exec vào container
docker-compose exec app-dev sh

# Check environment
env | grep DB_

# Test database connection
ping postgres
```

### Database management
```bash
# Access Adminer
http://localhost:8081
# Server: postgres
# Username: postgres
# Password: postgres
# Database: file_sharing_db

# Hoặc dùng psql
docker-compose exec postgres psql -U postgres -d file_sharing_db
```

### Generate Swagger docs
```bash
# Cài swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate
swag init -g cmd/server/main.go -o docs/swagger

# Hoặc dùng Makefile
make swagger
```

---

## 📚 Tài liệu khác

- [API Documentation](docs/API_DOCUMENTATION.md)
- [CI/CD Setup](.github/CI_CD_SETUP.md)
- [OpenAPI Spec](docs/openapi.yaml)
- [Database Schema](database/schema.sql)

---

## 🎓 Learning Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose](https://docs.docker.com/compose/)
- [Gin Framework](https://gin-gonic.com/docs/)
- [GitHub Actions](https://docs.github.com/en/actions)

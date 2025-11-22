# File Sharing System - Backend

Backend API cho hệ thống chia sẻ file tạm thời, được xây dựng bằng **Golang** với **Gin Framework** và **PostgreSQL**.

## Quick Start

Toàn bộ hướng dẫn setup/chạy (Docker, Windows, manual) đã gộp tại [`SETUP.md`](./SETUP.md).
Làm theo file đó để khởi chạy hệ thống.

## Danh Sách Thành Viên

| STT | Họ và Tên              | Mã Số Sinh Viên |
| --- | ------------------------- | ------------------ |
| 1   | Nguyễn Nhật Huy         | 2311197            |
| 2   | Tô Thế Hưng            | 2211384            |
| 3   | Nguyễn Phạm Mạnh Dũng | 2310559            |
| 4   | Đặng Thành Duy Đan    | 2310615            |
| 5   | Phan Đình Khang         | 2311459            |
| 6   | Võ Tiến Nam             | 2312205            |
| 7   | Nguyễn Hữu Minh Khôi   | 2352614            |
| 8   | Nguyễn Huỳnh Gia Đại  | 2310624            |

## 🚀 Tính năng

- ✅ Upload file (có hoặc không cần đăng nhập)
- ✅ Tạo link chia sẻ duy nhất cho mỗi file
- ✅ Thiết lập thời gian hiệu lực linh hoạt (from/to)
- ✅ Bảo vệ file bằng mật khẩu (bcrypt)
- ✅ Xác thực 2FA với TOTP (Google Authenticator)
- ✅ Chia sẻ với danh sách người dùng cụ thể
- ✅ Tự động xóa file hết hạn (cron job)
- ✅ JWT authentication
- ✅ Admin dashboard & system policy management

## 📋 Yêu cầu

- Go 1.21 hoặc cao hơn
- PostgreSQL 14+
- Docker & Docker Compose (optional)

## 🛠️ Cài đặt

Các bước cài đặt/khởi chạy (Docker + manual) → xem [`SETUP.md`](./SETUP.md).

## 📚 API Documentation

### API Specs

- OpenAPI YAML: `docs/openapi.yaml`
- Swagger JSON: `docs/swagger.json`
- Markdown: `docs/API_DOCUMENTATION.md`

### Generate Swagger docs

```bash
# Cài swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
make swagger

# Hoặc
swag init -g cmd/server/main.go -o docs/swagger
```

## 🔧 Makefile Commands

```bash
# Development
make run              # Chạy server (development mode)
make build            # Build binary
make test             # Chạy tests
make test-coverage    # Test với coverage report

# Database
make migrate-up       # Apply migrations
make migrate-down     # Rollback migrations
make db-seed          # Seed sample data

# Docker
make docker-build     # Build Docker image
make docker-run       # Run Docker container
make docker-up        # Docker compose up
make docker-down      # Docker compose down

# Code quality
make lint             # Run linter
make fmt              # Format code
make vet              # Run go vet

# Documentation
make swagger          # Generate Swagger docs

# Cleanup
make clean            # Clean build artifacts
```

## 📁 Cấu trúc thư mục

```
backend/
├── cmd/
│   └── server/           # Application entry point
│       └── main.go       # Main file (TODO: implement)
├── internal/
│   ├── config/           # Configuration loader (TODO)
│   ├── handlers/         # HTTP handlers (TODO)
│   ├── middleware/       # Middleware (auth, cors, etc.) (TODO)
│   ├── models/           # Domain models (TODO)
│   ├── repositories/     # Data access layer (TODO)
│   ├── services/         # Business logic (TODO)
│   └── utils/            # Utilities (TODO)
├── pkg/
│   ├── database/         # Database connection & queries
│   │   ├── schema.sql    # Database schema
│   │   └── demo_queries.sql
│   └── logger/           # Logger setup (TODO)
├── api/
│   └── routes/           # Route definitions (TODO)
├── configs/              # Config files
├── migrations/           # Database migrations
├── scripts/              # Helper scripts
├── storage/
│   └── uploads/          # File storage directory
├── tests/
│   ├── integration/      # Integration tests (TODO)
│   └── unit/             # Unit tests (TODO)
├── docs/
│   ├── openapi.yaml      # OpenAPI specification ✅
│   ├── swagger.json      # Swagger JSON ✅
│   └── API_DOCUMENTATION.md  # API guide ✅
├── .github/
│   └── workflows/        # CI/CD workflows
├── .env.example          # Environment template ✅
├── .env                  # Environment variables ✅
├── .gitignore            # Git ignore ✅
├── .dockerignore         # Docker ignore
├── Dockerfile            # Docker configuration
├── docker-compose.yml    # Docker Compose
├── go.mod                # Go modules ✅
├── go.sum                # Go dependencies checksum
├── Makefile              # Build commands ✅
└── README.md             # This file ✅
```

## 🔐 Environment Variables

Xem file `.env.example` để biết tất cả biến môi trường cần thiết.

### Quan trọng:

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=file_sharing_db

# JWT Secret (đổi thành giá trị bảo mật)
JWT_SECRET=change_this_to_secure_secret_minimum_32_characters_long

# File Storage
STORAGE_PATH=./storage/uploads
MAX_FILE_SIZE_MB=50

# Cleanup Secret (cho cron job)
CLEANUP_SECRET=change_this_cleanup_secret
```

## 🧪 Testing

```bash
# Chạy tất cả tests
make test

# Test với coverage
make test-coverage

# Test specific package
go test ./internal/services/...

# Verbose mode
go test -v ./...
```

## 📊 Database Schema

Database schema được định nghĩa trong `database/schema.sql`:

### Tables:

- `users` - User accounts với TOTP 2FA
- `files` - File metadata với validity period
- `shared_with` - Many-to-many relationship (files ↔ users)
- `system_policy` - Global configuration (singleton)

### Key Features:

- UUID primary keys
- Citext cho email/username (case-insensitive)
- Bcrypt password hashing
- TOTP secret storage
- File validity period với constraints
- Indexes cho performance

## 🔄 CI/CD

GitHub Actions workflows:

- **Test**: Chạy tests tự động khi push/PR
- **Build**: Build và push Docker image
- **Deploy**: Deploy lên production (khi merge vào main)

Xem `.github/workflows/` để biết chi tiết.

## 📝 Development Workflow

1. **Tạo feature branch**

   ```bash
   git checkout -b feature/your-feature
   ```
2. **Code & test**

   ```bash
   # Implement your feature
   # Write tests
   make test
   ```
3. **Format & lint**

   ```bash
   make fmt
   make lint
   ```
4. **Commit & push**

   ```bash
   git add .
   git commit -m "feat: your feature description"
   git push origin feature/your-feature
   ```
5. **Create Pull Request**

## 🐛 Troubleshooting

### Database connection error

```bash
# Kiểm tra PostgreSQL đang chạy
pg_isready -h localhost -p 5432

# Kiểm tra credentials trong .env
```

### Port already in use

```bash
# Tìm process đang dùng port 8080
netstat -ano | findstr :8080  # Windows
lsof -i :8080                  # Linux/Mac

# Kill process hoặc đổi port trong .env
```

### Module not found

```bash
# Download dependencies
go mod download
go mod tidy
```

# File Sharing System - Backend

Backend API cho hệ thống chia sẻ file tạm thời, được xây dựng bằng **Golang** với **Gin Framework** và **PostgreSQL**.

## Quick Start

```bash
cd backend
cp .env.example .env    # Tạo file .env từ template

# Development (hot reload, profile dev)
docker compose --profile dev up app-dev
# API: http://localhost:8082

# Production-like stack (migrate + app + nginx)
docker compose up -d
# API: http://localhost:8080
```

Tài liệu vận hành chi tiết xem tại [`backend/SETUP_WITH_DOCKER.md`](./backend/SETUP_WITH_DOCKER.md)

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

## Tính năng

- Upload file (có hoặc không cần đăng nhập)
- Tạo link chia sẻ duy nhất cho mỗi file
- Thiết lập thời gian hiệu lực linh hoạt (from/to)
- Bảo vệ file bằng mật khẩu (bcrypt)
- Xác thực 2FA với TOTP (Google Authenticator)
- Chia sẻ với danh sách người dùng cụ thể
- Tự động xóa file hết hạn (cron job)
- JWT authentication
- Admin dashboard & system policy management

## Yêu cầu

- Go 1.21 hoặc cao hơn
- PostgreSQL 14+
- Docker & Docker Compose (optional)

## Cài đặt

Xem hướng dẫn chi tiết tại [`backend/SETUP_WITH_DOCKER.md`](./backend/SETUP_WITH_DOCKER.md)

## Documentation & Reports

### Documents

Thư mục `docs/` chứa tài liệu hướng dẫn cho dự án:

- `backend/docs/API_DOCUMENTATION.md` - Hướng dẫn sử dụng API
- `backend/docs/openapi.yaml` - OpenAPI specification
- `backend/docs/swagger/` - Swagger documentation
- `backend/SETUP_WITH_DOCKER.md` - Hướng dẫn deploy với Docker
- `DEPLOYMENT.md` - Hướng dẫn deploy production
- `SOPS_SETUP.md` - Hướng dẫn setup SOPS cho environment variables

### Reports

Thư mục `reports/` chứa báo cáo cho môn học:

- `backend/report/Report.pdf` - Báo cáo dự án

### Generate Swagger docs

```bash
go install github.com/swaggo/swag/cmd/swag@latest
make swagger
swag init -g cmd/server/main.go -o docs/swagger
```

## Make Commands

```bash
# Compose-based workflow (khuyến nghị)
docker compose --profile dev up app-dev        # Dev (hot reload)
docker compose up -d                           # Prod stack (migrate + app + nginx)
docker compose down                            # Stop tất cả
docker compose logs -f app-dev                 # Xem log dev
docker compose run --rm migrate                # Chạy migrations thủ công (nếu cần)

# Makefile vẫn còn nhưng ưu tiên Compose để đồng nhất quy trình
make build
make clean
```

## Cấu trúc thư mục

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
├── docs/                 # Documentation
│   ├── openapi.yaml
│   ├── swagger.json
│   └── API_DOCUMENTATION.md
├── reports/              # Báo cáo môn học
├── .github/
│   └── workflows/
├── .env.example
├── .env
├── .gitignore
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Environment Variables

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

### Decrypt Environment Files (SOPS)

```bash
# Decrypt production
sops -d --input-type dotenv --output-type dotenv prod.enc > .env

# Decrypt development
sops -d --input-type dotenv --output-type dotenv dev.enc > .env
```

## Testing

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

## Database Schema

Database schema được định nghĩa trong `database/schema.sql`:

### Tables:

- `users` - User accounts với TOTP 2FA
- `files` - File metadata với validity period (includes `shared_with_emails` JSONB column for whitelist)
- `system_policy` - Global configuration (singleton)

### Key Features:

- UUID primary keys
- Citext cho email/username (case-insensitive)
- Bcrypt password hashing
- File validity period với constraints
- Indexes cho performance

## CI/CD

GitHub Actions workflows:

- **Test**: Chạy tests tự động khi push/PR
- **Build**: Build và push Docker image
- **Deploy**: Deploy lên production (khi merge vào main)

Xem `.github/workflows/` để biết chi tiết.

## Development Workflow

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

## Troubleshooting

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

### Xem logs để debug

```bash
make logs         # Tất cả logs
make logs-dev     # Dev logs
make ps           # Containers đang chạy
```

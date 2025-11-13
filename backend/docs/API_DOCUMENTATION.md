# API Documentation

## Swagger/OpenAPI Files

Project này có 2 file documentation:

### 1. OpenAPI YAML (`openapi.yaml`)
- Format: YAML
- Dễ đọc, dễ chỉnh sửa
- Sử dụng để import vào Swagger Editor/Postman
- **Recommended**: Sử dụng file này để phát triển và cập nhật API spec

### 2. Swagger JSON (`swagger.json`) 
- Format: JSON
- Tương thích với nhiều tool
- Có thể generate từ YAML

## Xem Documentation

### Online Tools

1. **Swagger Editor**: https://editor.swagger.io/
   - Copy nội dung `openapi.yaml` vào editor
   - Xem live preview và validate

2. **Swagger UI**: 
   - Deploy cùng backend để có interactive API docs
   - URL: http://localhost:8080/swagger/

3. **Postman**:
   - Import file `openapi.yaml` hoặc `swagger.json`
   - Tự động generate API collection

### Local Setup

```bash
# Cài đặt Swagger UI (sẽ implement trong code)
# Truy cập: http://localhost:8080/swagger/index.html
```

## API Overview

### Base URL
- Development: `http://localhost:8080/api`
- Production: `https://api.filesharing.com/api`

### Authentication
- Type: Bearer Token (JWT)
- Header: `Authorization: Bearer <token>`

### Endpoints Summary

#### Authentication
- `POST /auth/register` - Đăng ký
- `POST /auth/login` - Đăng nhập
- `POST /auth/login/totp` - Xác thực TOTP
- `POST /auth/totp/setup` - Thiết lập TOTP
- `POST /auth/totp/verify` - Xác minh TOTP
- `POST /auth/logout` - Đăng xuất

#### Files
- `POST /files/upload` - Upload file
- `GET /files/:shareToken` - Lấy thông tin file
- `GET /files/:shareToken/download` - Tải file
- `DELETE /files/:id` - Xóa file
- `GET /files/my` - Danh sách file của user

#### Admin
- `POST /admin/cleanup` - Xóa file hết hạn
- `GET /admin/policy` - Lấy cấu hình hệ thống
- `PATCH /admin/policy` - Cập nhật cấu hình

## Response Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Success |
| 201 | Created | Upload thành công |
| 400 | Bad Request | Validation error |
| 401 | Unauthorized | Cần đăng nhập |
| 403 | Forbidden | Không có quyền |
| 404 | Not Found | Không tìm thấy |
| 410 | Gone | File đã hết hạn |
| 413 | Payload Too Large | File quá lớn |
| 423 | Locked | File chưa đến thời gian hiệu lực |

## File Status

| Status | Description |
|--------|-------------|
| `pending` | Chưa đến thời gian `availableFrom` |
| `active` | Đang trong thời gian hiệu lực |
| `expired` | Đã hết hạn (`availableTo` đã qua) |

## Validity Period Logic

| Input | Result |
|-------|--------|
| FROM + TO | Hiệu lực từ FROM đến TO |
| Chỉ TO | Hiệu lực từ hiện tại đến TO |
| Chỉ FROM | Hiệu lực từ FROM đến FROM + 7 ngày |
| Không có | Hiệu lực từ hiện tại đến +7 ngày |

## Examples

### Upload File (Anonymous)
```bash
curl -X POST http://localhost:8080/api/files/upload \
  -F "file=@document.pdf" \
  -F "isPublic=true" \
  -F "availableTo=2025-11-20T00:00:00Z"
```

### Upload File (Authenticated)
```bash
curl -X POST http://localhost:8080/api/files/upload \
  -H "Authorization: Bearer eyJhbG..." \
  -F "file=@secret.pdf" \
  -F "isPublic=false" \
  -F "password=secret123" \
  -F 'sharedWith=["user1@example.com"]'
```

### Download File (Public)
```bash
curl -O http://localhost:8080/api/files/abc123def456/download
```

### Download File (Password Protected)
```bash
curl -O http://localhost:8080/api/files/abc123def456/download?password=secret123
```

### Download File (Private - Authenticated)
```bash
curl -O http://localhost:8080/api/files/abc123def456/download \
  -H "Authorization: Bearer eyJhbG..."
```

## Testing with Postman

1. Import `openapi.yaml` hoặc `swagger.json`
2. Set environment variables:
   - `base_url`: http://localhost:8080/api
   - `access_token`: (sau khi login)
3. Test các endpoints

## Updating Documentation

Khi thêm/sửa API:
1. Cập nhật `openapi.yaml`
2. Validate tại https://editor.swagger.io/
3. Commit changes
4. Update backend code tương ứng

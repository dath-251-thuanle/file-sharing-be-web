# API Documentation

## Table of Contents

- [API Specification](#api-specification)
- [API Overview](#api-overview)
- [Endpoints Summary](#endpoints-summary)
- [Response Codes](#response-codes)
- [Database Tables](#database-tables)
- [TOTP/2FA Flow](#totp2fa-flow)
- [File Statistics &amp; Analytics](#file-statistics--analytics)
- [File Status](#file-status)
- [Validity Period Logic](#validity-period-logic)
- [Security](#security)
- [Download Access Control](#download-access-control)
- [Quick Reference](#quick-reference)

---

## API Specification

Project sử dụng **OpenAPI 3.0.3** để định nghĩa API:

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

- `POST /auth/register` - Đăng ký tài khoản mới
- `POST /auth/login` - Đăng nhập (trả về token hoặc yêu cầu TOTP)
- `POST /auth/login/totp` - Xác thực TOTP để hoàn tất đăng nhập
- `POST /auth/totp/setup` - Thiết lập TOTP cho user (yêu cầu Bearer token)
- `POST /auth/totp/verify` - Xác minh mã TOTP để kích hoạt 2FA (yêu cầu Bearer token)
- `POST /auth/logout` - Đăng xuất
- `GET /user` - Lấy thông tin user và danh sách file

#### Files

- `POST /files/upload` - Upload file (hỗ trợ enableTOTP, trả về QR code nếu bật)
- `GET /files/{id}` - Lấy thông tin file theo UUID (chỉ owner/admin)
- `GET /files/{id}/stats` - Lấy thống kê download của file (chỉ owner/admin)
- `GET /files/{id}/download-history` - Lấy lịch sử download chi tiết (chỉ owner/admin)
- `GET /files/{shareToken}` - Lấy thông tin file qua share token (public)
- `GET /files/{shareToken}/download` - Tải file (hỗ trợ password, downloadToken)
- `DELETE /files/{id}` - Xóa file (chỉ owner)
- `POST /files/{shareToken}/totp/validate` - Xác thực mã TOTP để lấy download token

#### Admin

- `POST /admin/cleanup` - Xóa file hết hạn
- `GET /admin/policy` - Lấy cấu hình hệ thống
- `PATCH /admin/policy` - Cập nhật cấu hình

## Response Codes

| Code | Meaning           | Description                            |
| ---- | ----------------- | -------------------------------------- |
| 200  | OK                | Success                                |
| 201  | Created           | Upload thành công                    |
| 400  | Bad Request       | Validation error / Invalid token       |
| 401  | Unauthorized      | Cần đăng nhập / Token expired      |
| 403  | Forbidden         | Không có quyền / Wrong password     |
| 404  | Not Found         | Không tìm thấy resource             |
| 409  | Conflict          | Email/username đã tồn tại          |
| 410  | Gone              | File đã hết hạn                    |
| 413  | Payload Too Large | File quá lớn                         |
| 423  | Locked            | File chưa đến thời gian hiệu lực |

## Database Tables

| Table                     | Description               | Key Features                           |
| ------------------------- | ------------------------- | -------------------------------------- |
| `users`                 | User accounts             | TOTP support, roles (user/admin)       |
| `files`                 | Uploaded files metadata   | Share tokens, TOTP, password, validity |
| `shared_with`           | File sharing whitelist    | Many-to-many: files ↔ users           |
| `file_statistics`       | Aggregated download stats | Download count, unique users           |
| `download_history`      | Detailed download log     | Audit trail, anonymous support         |
| `system_policy`         | System configuration      | File size limits, validity rules       |

**Schema:** Xem `pkg/database/schema.sql`
**Migrations:** Xem `migrations/` folder
**Demo Queries:** Xem `pkg/database/demo_queries.sql`

## TOTP/2FA Flow

### User TOTP (2FA for Account Login)

**Luồng bật TOTP:**

1. User đăng ký tài khoản: `POST /auth/register`
2. User đăng nhập lần đầu: `POST /auth/login` → nhận `accessToken`
3. User muốn bật 2FA: `POST /auth/totp/setup` (cần Bearer token) → nhận `secret` + `qrCode`
4. User quét QR code bằng Google Authenticator/Authy
5. User xác minh mã: `POST /auth/totp/verify` (cần Bearer token) → tài khoản được đánh dấu `totpEnabled=true`

**Luồng đăng nhập với TOTP:**

1. User nhập email/password: `POST /auth/login` → trả về `requireTOTP: true`
2. User nhập mã 6 số từ app: `POST /auth/login/totp` → nhận `accessToken`

### File TOTP (2FA for File Download)

**Luồng upload file với TOTP:**

1. Owner upload file với `enableTOTP=true`: `POST /files/upload`
2. Backend tự động sinh `secret` + `qrCode` trong response
3. Owner quét QR code vào Google Authenticator/Authy
4. File được đánh dấu `totpEnabled=true` ngay sau khi upload

**Luồng download file có TOTP:**

1. User truy cập `/files/{shareToken}` → thấy `file.totpEnabled = true`
2. User liên hệ owner để lấy mã 6 số hiện tại từ Google Authenticator
3. User gọi `POST /files/{shareToken}/totp/validate` với mã → nhận `downloadToken` (valid 5 phút)
4. User download: `GET /files/{shareToken}/download?downloadToken=xxx`

**Use case:** Owner nghi ngờ password bị leak, bật TOTP để chỉ những người được owner cung cấp mã mới tải được.

**Lưu ý:** File TOTP hoạt động với cả anonymous upload (không cần Bearer token).

## File Statistics & Analytics

### GET /files//stats

Lấy thống kê download của file (chỉ owner/admin).

**Dữ liệu trả về:**

- `downloadCount` - Tổng số lượt download
- `uniqueDownloaders` - Số người download khác nhau (authenticated users only)
- `lastDownloadedAt` - Thời điểm download gần nhất

**Source:** Bảng `file_statistics`

**Note:** Anonymous uploads không có statistics

### GET /files//download-history

Lấy lịch sử download chi tiết (chỉ owner/admin).

**Dữ liệu trả về:**

- Danh sách downloads với: downloader info, timestamp, completed status
- Downloader = null nếu là anonymous download
- Hỗ trợ pagination: `?page=1&limit=50`

**Source:** Bảng `download_history`

**Use case:** Audit trail, xem ai đã download file, khi nào, thành công hay bị gián đoạn

## File Status

| Status      | Description                               |
| ----------- | ----------------------------------------- |
| `pending` | Chưa đến thời gian `availableFrom`  |
| `active`  | Đang trong thời gian hiệu lực         |
| `expired` | Đã hết hạn (`availableTo` đã qua) |

## Validity Period Logic

| Input      | Result                                   |
| ---------- | ---------------------------------------- |
| FROM + TO  | Hiệu lực từ FROM đến TO             |
| Chỉ TO    | Hiệu lực từ hiện tại đến TO       |
| Chỉ FROM  | Hiệu lực từ FROM đến FROM + 7 ngày |
| Không có | Hiệu lực từ hiện tại đến +7 ngày |

## Security

### Bearer Token (JWT)

- Lấy từ: `POST /auth/login` hoặc `POST /auth/login/totp`
- Format: `Authorization: Bearer <token>`
- Dùng cho: Tất cả authenticated endpoints

### X-Cron-Secret

- Static secret key cho cron job
- Cấu hình qua environment variable: `CRON_SECRET`
- Dùng cho endpoint `/admin/cleanup`
- Gửi qua header: `X-Cron-Secret: <secret>`

### Download Token

- Token tạm thời (valid 5 phút) sau khi validate TOTP
- Lấy từ: `POST /files/{shareToken}/totp/validate`
- Dùng cho: `GET /files/{shareToken}/download?downloadToken=xxx`
- Format: Query parameter `?downloadToken=xxx`

## Download Access Control

Các endpoint tải file hỗ trợ nhiều lớp bảo mật đồng thời. Backend kiểm tra theo thứ tự:

1. **File status**: nếu hết hạn → `410`, nếu chưa đến thời gian → `423`
2. **Whitelist**: nếu file có `sharedWith` → yêu cầu Bearer token. Thiếu token → `401`, user không nằm trong whitelist → `403`
3. **Password**: nếu cấu hình password → yêu cầu query/body `password`. Thiếu hoặc sai → `403` (download) / `400` (validate)
4. **TOTP**: nếu `totpEnabled=true` → phải có `downloadToken` (được cấp sau khi validate). Thiếu hoặc sai → `400`

### `/files/{shareToken}/download`

| HTTP code | Case                     | Description                                   |
| --------- | ------------------------ | --------------------------------------------- |
| `200`   | Success                  | Trả file binary                              |
| `400`   | `missingDownloadToken` | File có TOTP nhưng thiếu `downloadToken` |
| `400`   | `invalidDownloadToken` | Token sai hoặc hết hạn                     |
| `401`   | `missingAuth`          | File private nhưng thiếu Bearer token       |
| `403`   | `wrongPassword`        | Password sai                                  |
| `403`   | `missingPassword`      | File có password nhưng không gửi          |
| `403`   | `notWhitelisted`       | User không nằm trong danh sách chia sẻ    |
| `404`   | `notFound`             | Share token không tồn tại                  |
| `410`   | `expired`              | File đã hết hạn                           |
| `423`   | `pending`              | File chưa đến thời gian hiệu lực        |

### `/files/{shareToken}/totp/validate`

| HTTP code | Case                | Description                            |
| --------- | ------------------- | -------------------------------------- |
| `200`   | Success             | Trả `downloadToken` (valid 5 phút) |
| `400`   | `invalidTOTPCode` | Mã TOTP sai/hết hạn                 |
| `400`   | `wrongPassword`   | Password sai (nếu file có password)  |
| `400`   | `missingPassword` | Quên gửi password                    |
| `400`   | `totpNotEnabled`  | File không bật TOTP                  |
| `404`   | `notFound`        | Share token không tồn tại           |
| `410`   | `expired`         | File đã hết hạn                    |

## Quick Reference

### Common Use Cases

#### 1. Anonymous Upload + Share

```
POST /files/upload
→ Nhận shareToken
→ Chia sẻ link: https://domain.com/f/{shareToken}
```

#### 2. Upload với Password Protection

```
POST /files/upload
Body: { file, password: "secret123" }
→ Người download cần password
```

#### 3. Upload với TOTP

```
POST /files/upload
Body: { file, enableTOTP: true }
→ Nhận QR code
→ Quét vào Google Authenticator
→ Người download phải liên hệ owner để lấy mã 6 số
```

#### 4. Share với Whitelist

```
POST /files/upload
Body: { 
  file, 
  isPublic: false,
  sharedWith: ["user1@gmail.com", "user2@gmail.com"]
}
→ Chỉ user1 và user2 có thể download (cần đăng nhập)
```

#### 5. Owner Xem Ai Đã Download File

```
1. GET /files/{id}/stats → Tổng quan
2. GET /files/{id}/download-history → Chi tiết từng lượt download
```

#### 6. Download File Có Nhiều Lớp Bảo Mật

```
File có: password + TOTP + whitelist

1. Đăng nhập (để pass whitelist check)
2. Gọi POST /files/{shareToken}/totp/validate
   Body: { code: "123456", password: "secret123" }
   → Nhận downloadToken
3. GET /files/{shareToken}/download?downloadToken=xxx
```

### Migration Commands

```bash
# Apply migrations
make migrate-up

# Check current version
make migrate-version

# Rollback if needed
make migrate-down

# Create new migration
make migrate-create NAME=my_migration
```

### Testing with Demo Data

```bash
# Run demo queries
psql -h localhost -U postgres -d file_sharing_db -f pkg/database/demo_queries.sql

# Or with Docker
make db-shell
# Then paste queries from demo_queries.sql
```

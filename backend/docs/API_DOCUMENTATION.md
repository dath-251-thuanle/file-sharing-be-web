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

Project sử dụng **OpenAPI 3.0.4** để định nghĩa API:

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
- Production: `https://sharefilehcmut.azurewebsites.net/api`

### Authentication

- Type: Bearer Token (JWT)
- Header: `Authorization: Bearer <token>`

### Endpoints Summary

#### Authentication

- `POST /auth/register` – Đăng ký tài khoản mới (username/email/password required; password minimum 8 characters). Trả về `userId` khi thành công. `409 Conflict` nếu email/username đã dùng.
- `POST /auth/login` – Đăng nhập bằng email và password. Nếu user chưa bật TOTP, trả về `accessToken`. Nếu đã bật thì trả về `requireTOTP: true` cùng `cid`/thông tin để gọi `/auth/login/totp`.
- `POST /auth/login/totp` – Hoàn tất đăng nhập khi TOTP được yêu cầu (không cần Bearer token). Yêu cầu `cid` + `code` để đổi lấy `accessToken`.
- `POST /auth/totp/setup` – Sinh secret + QR code để bật TOTP (cần Bearer token). Trả về `totpSetup` payload.
- `POST /auth/totp/verify` – Xác minh mã TOTP 6 chữ số để kích hoạt 2FA. Cần Bearer token.
- `POST /auth/logout` – Đăng xuất (client chỉ cần xóa token).
- `GET /user` – Lấy profile user hiện tại (id, username, email, role, totpEnabled). Yêu cầu Bearer token.

#### Files

- `POST /files/upload` – Upload file (multipart form-data với `file`, `isPublic`, `password`, `availableFrom`, `availableTo`, `sharedWith`). Anonymous upload chỉ được public. Private uploads yêu cầu Bearer token. Hỗ trợ whitelist email và password validation, thời gian hiệu lực theo `system_policy`.
- `GET /files/my` – Lấy danh sách file của user hiện tại có pagination (`page`, `limit`, `status`, `sortBy`, `order`) và `summary` trạng thái.
- `GET /files/info/{id}` – Lấy metadata file đầy đủ theo UUID (owner hoặc admin). Trả về `sharedWith`, owner info, status, `hoursRemaining`.
- `DELETE /files/info/{id}` – Xóa file theo UUID (owner hoặc admin).
- `GET /files/stats/{id}` – Lấy thống kê download (owner/admin) từ bảng `file_statistics`.
- `GET /files/download-history/{id}` – Lấy lịch sử download chi tiết với pagination (owner/admin).
- `GET /files/{shareToken}` – Lấy metadata giới hạn qua share token (public). Không trả `sharedWith`.
- `GET /files/{shareToken}/download` – Tải file (binary). Kiểm tra theo thứ tự: trạng thái (`expired/pending`), whitelist (nếu có), password (`X-File-Password`). Có thể sử dụng Bearer token và credential tương ứng.
- `GET /files/{shareToken}/preview` – Xem inline (PDF/image/video) (áp dụng cùng logic bảo mật như download).

#### Admin

- `POST /admin/cleanup` – Xóa file hết hạn. Yêu cầu Bearer admin token hoặc header `X-Cron-Secret`.
- `GET /admin/policy` – Lấy system policy (admin token). Trả về giới hạn file size, validity, password length.
- `PATCH /admin/policy` – Cập nhật system policy (admin token). Yêu cầu payload hợp lệ (`maxValidityDays >= minValidityHours`, ...).

#### Public Policy

- `GET /policy/limits` – Trả về `maxFileSizeMB` và `requirePasswordMinLength` để client validate trước khi upload (public endpoint).
- `PATCH /policy/limits` – (Admin-only) Cập nhật giới hạn policy công khai; sau khi cập nhật, client có thể đọc lại qua `GET /policy/limits`.

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

## Service Tests

Các test của từng service nằm trong thư mục `backend/services_test`. Tài liệu chi tiết có thể xem trong [`backend/services_test/README.md`](../backend/services_test/README.md).

- **Auth service**: kiểm thử tất cả luồng login/register/TOTP (thành công, sai password, duplicate user, setup/verify TOTP, generate token, lấy profile).
- **File service**: giả lập storage để test upload/download/delete, pagination, search, file status (pending/expired), policy defaults và rollback khi storage hoặc database lỗi.
- **Helpers**: `newTestDB` reset database sau mỗi test, seed `system_policy`.

Chạy toàn bộ test:

```bash
cd backend
go test ./services_test/...
```


## Database Tables

| Table                | Description               | Key Features                     |
| -------------------- | ------------------------- | -------------------------------- |
| `users`            | User accounts             | TOTP support, roles (user/admin) |
| `files`            | Uploaded files metadata   | Share tokens, password, validity, shared_with_emails (JSONB) |
| `file_statistics`  | Aggregated download stats | Download count, unique users     |
| `download_history` | Detailed download log     | Audit trail, anonymous support   |
| `system_policy`    | System configuration      | File size limits, validity rules |

**Schema:** Xem `pkg/database/schema.sql`
**Migrations:** Xem `migrations/` folder
**Demo Queries:** Xem `pkg/database/demo_queries.sql`

## Cloud Storage

Backend sử dụng Azure Blob Storage:
- **Provider**: Azure
- **Max File Size**: 50MB
- **Containers**: Public (file public), Private (file protected)

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
| `pending` | Chưa đến thời gian `availableFrom`  (owner có thể preview bằng JWT, người khác nhận 423) |
| `active`  | Đang trong thời gian hiệu lực         |
| `expired` | Đã hết hạn (`availableTo` đã qua) |

## Validity Period Logic

| Input      | Result                                   |
| ---------- | ---------------------------------------- |
| FROM + TO  | Hiệu lực từ FROM đến TO             |
| Chỉ TO    | Hiệu lực từ hiện tại đến TO       |
| Chỉ FROM  | Hiệu lực từ FROM đến FROM + 7 ngày |
| Không có | Hiệu lực từ hiện tại đến +7 ngày |

**Validation bổ sung:**
- `availableFrom` ≤ `availableTo`.
- `availableTo` không nằm trong quá khứ tại thời điểm upload và không vượt quá `system_policy.maxValidityDays`.
- Tổng thời gian hiệu lực phải nằm trong giới hạn policy; vi phạm → backend trả lỗi `invalidValidityRange`.

**System Policy:**
- Default validity: 7 days
- Min validity: 1 hour, Max validity: 30 days
- Min password length: 6 characters
- Admin có thể thay đổi qua `PATCH /admin/policy`

## Security

### Bearer Token (JWT)

- Lấy từ: `POST /auth/login` hoặc `POST /auth/login/totp`
- Format: `Authorization: Bearer <token>`
- Dùng cho: Tất cả authenticated endpoints

### X-Cron-Secret

- Secret key cho cron job (lưu trong env)
- Dùng cho endpoint `/admin/cleanup`
- Header: `X-Cron-Secret: <secret>`

### Rate Limiting

- General API: 60 requests/minute
- File Upload: 10 uploads/hour

### CORS

Production origin: `https://sharefilehcmut.azurewebsites.net`
Custom headers: `X-Cron-Secret`, `X-File-Password`

## Download Access Control

Các endpoint tải file hỗ trợ nhiều lớp bảo mật đồng thời. Backend kiểm tra theo thứ tự:

1. **File status**: nếu hết hạn → `410`, nếu chưa đến thời gian → `423`
2. **Whitelist**: nếu file có `sharedWith` → yêu cầu Bearer token. Thiếu token → `401`, user không nằm trong whitelist → `403`
3. **Password**: nếu cấu hình password → yêu cầu query/body `password`. Thiếu hoặc sai → `403`

### `/files/{shareToken}/download`

| HTTP code | Case                | Description                                |
| --------- | ------------------- | ------------------------------------------ |
| `200`   | Success             | Trả file binary                           |
| `401`   | `missingAuth`     | File private nhưng thiếu Bearer token    |
| `403`   | `wrongPassword`   | Password sai                               |
| `403`   | `missingPassword` | File có password nhưng không gửi       |
| `403`   | `notWhitelisted`  | User không nằm trong danh sách chia sẻ |
| `404`   | `notFound`        | Share token không tồn tại               |
| `410`   | `expired`         | File đã hết hạn                        |
| `423`   | `pending`         | File chưa đến thời gian hiệu lực     |

**Owner preview & notification:**
- Chủ file (JWT hợp lệ, `sub` = ownerId) có thể bypass trạng thái `pending` để kiểm thử link; người khác vẫn nhận `423` cho tới khi `availableFrom` đến.
- Khuyến nghị cấu hình cron/background job gửi email/SMS/webhook khi file chuyển từ `pending` sang `active` cho owner và whitelist; endpoint này không tự gửi thông báo.

## Quick Reference

### Common Use Cases

#### 1. Anonymous Upload + Share

```
POST /files/upload
→ Nhận shareToken
→ Chia sẻ link: https://exampledomain.com/f/{shareToken}
```

**Lưu ý:** Anonymous chỉ được upload file public, không đặt whitelist/password nâng cao và không thể chỉnh sửa/xóa sau khi upload. Muốn private → đăng nhập trước.

#### 2. Upload với Password Protection

```
POST /files/upload
Body: { file, password: "secret123" }
→ Người download cần password
```

#### 3. Share với Whitelist

```
POST /files/upload
Body: { 
  file, 
  isPublic: false,
  sharedWith: ["user1@gmail.com", "user2@gmail.com"]
}
→ Chỉ user1 và user2 có thể download (cần đăng nhập)
```

#### 4. Owner Xem Ai Đã Download File

```
1. GET /files/{id}/stats → Tổng quan
2. GET /files/{id}/download-history → Chi tiết từng lượt download
```

#### 5. Owner Xem Danh Sách File Của Mình

```
GET /files/my?status=all&page=1&limit=20
→ Nhận danh sách file + pagination + summary (active/pending/expired/deleted)
```

#### 6. Download File Có Nhiều Lớp Bảo Mật

```
File có: password + whitelist

1. Đăng nhập (để pass whitelist check)
2. GET /files/{shareToken}/download?password=secret123
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

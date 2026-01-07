# Service Test Overview

Document này mô tả những gì được kiểm thử ở tầng service trong `backend/services_test`.

## Test Helpers

- `newTestDB(t *testing.T)` tạo một database PostgreSQL tạm thời theo biến môi trường (mặc định `file_sharing_db`) và reset toàn bộ bảng trước/sau mỗi test.
- `resetTestDatabase` truncate tất cả bảng liên quan (`users`, `files`, `system_policy`, v.v.) rồi seed policy mặc định.
- `seedDefaultSystemPolicy` đảm bảo `system_policy` luôn tồn tại với các giới hạn chuẩn (55MB max, 7 ngày default, 6 ký tự password tối thiểu).

## Authentication Service Tests (`auth_service_test.go`)

Các test focus vào flow đăng nhập/đăng ký/TOTP:

- `Login`: kiểm thử thành công (TOTP bật/tắt), sai password, user không tồn tại, repository lỗi; đảm bảo error đúng type (`ErrInvalidCredentials`, `gorm.ErrRecordNotFound`).
- `Register`: kiểm thử đăng ký thành công, username/email trùng, password quá ngắn.
- `TOTP`: setup tạo secret/qrCode, verify đúng và sai mã.
- `Token & Profile`: generate JWT hợp lệ và gọi `GetProfile` thành công/không tìm thấy.
- Các mocks sử dụng struct `mockUserRepo` với function pointer cho từng method để dễ inject lỗi.

## File Service Tests (`file_service_test.go`)

Các test dùng fake storage để mô phỏng upload/download/delete:

- **Upload**: public/private, bắt buộc owner khi private, sharedWith email, availability window, password protection, rollback khi storage/db lỗi, sử dụng default validity từ policy.
- **Download**: thành công, path rỗng, không có storage; test thứ tự kiểm tra (storage config, status, whitelist, password).
- **Queries**: `GetByID`, `GetByShareToken`, `GetByOwnerID` (pagination), `GetPublicFiles`, `SearchFiles`.
- **Lifecycle**: Delete, expired files, pending files, hệ thống policy trả về giá trị hợp lệ.
- Fake storage (`fakeStorage`) track file, cho phép inject lỗi upload/download/delete.

## Running the Tests

 Từ thư mục `backend`:

```bash
go test ./services_test/...           # tất cả test service
go test ./services_test/auth_service_test.go
go test ./services_test/file_service_test.go
```

Có thể thêm `-v`, `-race`, `-coverprofile` để kiểm tra sâu hơn.

## Khi cần mở rộng

- Đoạn này là điểm bắt đầu khi thêm test cho các service khác (user, download_history, statistics) hoặc khi cần test integration/controller.



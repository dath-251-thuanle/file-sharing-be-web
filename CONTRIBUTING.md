# Quy Ước Đóng Góp (Convention)

## 📑 Mục Lục

- [Quy Ước Đặt Tên Nhánh](#quy-ước-đặt-tên-nhánh)
- [Quy Ước Commit Message](#quy-ước-commit-message)
- [Quy Trình Làm Việc](#quy-trình-làm-việc)
- [Lưu Ý Quan Trọng](#lưu-ý-quan-trọng)

---

## 🌿 Quy Ước Đặt Tên Nhánh

### Định Dạng

```
<loại>/<mô-tả-ngắn-gọn>
```

### Các Loại Nhánh

| Loại        | Mô Tả               | Ví Dụ                     |
| ------------| ------------------- | ------------------------- |
| `feat/`     | Tính năng mới       | `feature/upload-file`     |
| `fix/`      | Sửa lỗi             | `fix/loi-dang-nhap`       |
| `refactor/` | Tái cấu trúc code   | `refactor/database-query` |
| `docs/`     | Cập nhật tài liệu   | `docs/huong-dan-setup`    |
| `test/`     | Thêm tests          | `test/user-service`       |
| `chore/`    | Bảo trì, cập nhật   | `chore/cap-nhat-thu-vien` |

### Quy Tắc

- ✅ Dùng chữ thường
- ✅ Dùng dấu gạch ngang `-` để phân tách từ
- ✅ Ngắn gọn nhưng rõ nghĩa
- ❌ Không dùng ký tự đặc biệt

---

## 💬 Quy Ước Commit Message

Sử dụng chuẩn **Conventional Commits**.

### Định Dạng

```
<type>: <mô tả ngắn>

<nội dung chi tiết (tùy chọn)>

<footer (tùy chọn)>
```

### Các Type

| Type       | Mô Tả                 | Ví Dụ                               |
| ---------- | --------------------- | ----------------------------------- |
| `feat`     | Tính năng mới         | `feat: thêm đăng nhập JWT`          |
| `fix`      | Sửa lỗi               | `fix: sửa lỗi upload file lớn`      |
| `docs`     | Cập nhật tài liệu     | `docs(readme): cập nhật hướng dẫn`  |
| `style`    | Format code           | `style: định dạng lại code`         |
| `refactor` | Tái cấu trúc          | `refactor: tối ưu database query`   |
| `perf`     | Cải thiện hiệu năng   | `perf: tối ưu tốc độ tải file`      |
| `test`     | Thêm/sửa tests        | `test: thêm test cho user service`  |
| `chore`    | Bảo trì               | `chore: cập nhật dependencies`      |
| `ci`       | CI/CD                 | `ci: thêm GitHub Actions`           |
| `build`    | Build system          | `build: cập nhật Docker config`     |

### Ví Dụ Tốt

```
feat: thêm xác thực JWT token

Triển khai hệ thống xác thực JWT với refresh token.
- Thêm tạo và validate token
- Thêm middleware cho protected routes

Closes #123
```

```
fix(upload): sửa lỗi validation kích thước file

Sửa lỗi file lớn hơn 10MB bị từ chối không đúng.
```

### Ví Dụ Không Tốt

```
❌ update code
❌ fix bug
❌ WIP
❌ sửa lỗi
```

### Nguyên Tắc

1. **Dòng đầu** (tiêu đề):

   - Giữ dưới 50 ký tự
   - Dùng động từ nguyên thể ("thêm" không phải "đã thêm")
   - Không kết thúc bằng dấu chấm
2. **Nội dung** (tùy chọn):

   - Giải thích **cái gì** và **tại sao**
   - Cách dòng tiêu đề một dòng trống
3. **Footer** (tùy chọn):

   - Link issue: `Closes #123`, `Fixes #456`
   - Breaking change: `BREAKING CHANGE: ...`

---

## 🔄 Quy Trình Làm Việc

### Khi Bắt Đầu Feature Mới

```bash
# Bước 1: Cập nhật nhánh main
git checkout main
git pull origin main

# Bước 2: Tạo nhánh mới từ main
git checkout -b feat/ten-tinh-nang

# Bước 3: Làm việc và commit
git add .
git commit -m "feat: thêm chức năng upload file"

# Bước 4: Push nhánh lên GitHub
git push -u origin feat/ten-tinh-nang

# Bước 5: Tạo Pull Request trên GitHub
```

### Trước Khi Tạo Pull Request

```bash
# 1. Cập nhật code mới nhất từ main
git checkout main
git pull origin main
git checkout feature/ten-tinh-nang
git merge main

# 2. Giải quyết conflicts (nếu có)

# 3. Chạy tests
go test ./...

# 4. Format code
go fmt ./...
go vet ./...

# 5. Push code đã cập nhật
git push
```

### Yêu Cầu Trước Khi Merge

- [ ] Code tuân theo coding style
- [ ] Tất cả tests pass
- [ ] Đã thêm tests cho tính năng mới
- [ ] Commit messages tuân theo convention
- [ ] Không có merge conflicts
- [ ] Có ít nhất 1 reviewer approve

---

## ⚠️ Lưu Ý Quan Trọng

### ❌ KHÔNG BAO GIỜ

#### 1. Commit trực tiếp lên main

```bash
# KHÔNG LÀM NHƯ NÀY
git checkout main
git commit -m "fix"
git push
```

#### 2. Quên pull thay đổi mới nhất

```bash
# KHÔNG LÀM NHƯ NÀY
git checkout -b feat/xyz  # mà không pull main trước
```

#### 3. Viết commit message không rõ

```bash
# KHÔNG LÀM NHƯ NÀY
git commit -m "fix"
git commit -m "update"
git commit -m "WIP"
```

### ✅ LUÔN LUÔN

#### 1. Pull main trước khi tạo nhánh mới

```bash
git checkout main
git pull origin main
git checkout -b feat/xyz
```

#### 2. Viết commit message rõ ràng

```bash
git commit -m "feat(upload): thêm validation kích thước file"
```

#### 3. Test code trước khi push

```bash
go test ./...
go fmt ./...
git push
```

#### 4. Cập nhật nhánh trước khi tạo PR

```bash
git checkout main
git pull origin main
git checkout feature/xyz
git merge main
```

---

## 🎯 Tóm Tắt

**Quy trình chuẩn khi làm tính năng mới:**

```bash
# 1. Pull main
git checkout main && git pull origin main

# 2. Tạo nhánh
git checkout -b feature/ten-tinh-nang

# 3. Code + Commit
git add .
git commit -m "feat: mô tả ngắn gọn"

# 4. Push
git push -u origin feature/ten-tinh-nang

# 5. Tạo PR trên GitHub
```

**Nhớ:**

- 🌿 Tên nhánh: `<type>/<mô-tả>`
- 💬 Commit: `<type>: <mô tả>`
- 🔄 Luôn pull main trước khi tạo nhánh mới
- ✅ Test trước khi push

---

Có câu hỏi? Tạo issue với label `question` hoặc đi hỏi. Trong trường hợp bị conflict khi mở pull request, từ từ mà fix nhé

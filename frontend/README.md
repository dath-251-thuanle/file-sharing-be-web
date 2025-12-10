# File Sharing Frontend Project

## Tổng Quan Dự Án

Đây là repository chứa mã nguồn **Front-end** (Next.js App Router) cho hệ thống chia sẻ file bảo mật. Hệ thống cho phép người dùng upload file tạm thời, chia sẻ qua link, bảo vệ bằng mật khẩu/TOTP và quản lý thời gian hiệu lực linh hoạt.

Frontend sẽ tương tác chặt chẽ với Backend API để xử lý các nghiệp vụ: Authentication, File Management, Access Control và System Administration.

### Tech Stack

  * **Framework:** Next.js 14+ (App Router)
  * **Language:** TypeScript
  * **Styling:** Tailwind CSS
  * **State Management:** React Hooks / Context API
  * **HTTP Client:** Axios (Khuyên dùng để xử lý Interceptor dễ dàng hơn)

-----

## Phân Chia Công Việc (Team Assignments)

Công việc được chia theo các module chức năng chính. Mỗi thành viên chịu trách nhiệm từ giao diện (UI) đến logic gọi API (Integration) cho phần của mình.

| STT | Module | Người Phụ Trách | Chi Tiết Nhiệm Vụ & Logic Frontend | API Endpoints |
| :-- | :--- | :--- | :--- | :--- |
| **1** | **Auth & TOTP** | **Bảo Minh** | - **Login:** Xử lý flow đăng nhập thường & đăng nhập 2 bước (check `requireTOTP: true`).<br>- **Register:** Form đăng ký validation.<br>- **TOTP Setup:** Hiển thị QR Code, xác thực mã OTP kích hoạt.<br>- **Lưu trữ:** Quản lý Token/Session trong LocalStorage/Cookies. | `/api/auth/register`<br>`api/auth/login`<br>`api/auth/login/totp`<br>`api/auth/totp/setup`<br>`api/auth/totp/verify` |
| **2** | **User Dashboard** | **Bảo Minh** | - **Danh sách file:** Hiển thị dạng bảng, phân trang (`page`, `limit`).<br>- **Bộ lọc:** Filter file theo trạng thái (`active`, `expired`, `pending`).<br>- **Xử lý UI:** Hiển thị `hoursRemaining`, nút Copy Link, nút Xóa file.<br>- **State Management:** Đồng bộ trạng thái khi user xóa file hoặc logout. | `/api/files/my`<br>`/api/files/:id` (DELETE) |
| **3** | **Upload File** | **Khánh** | - **Form Upload:** Xử lý Multipart/form-data.<br>- **Cấu hình:** Toggle Password, Date Picker (`AvailableFrom` \< `AvailableTo`), nhập email share.<br>- **Validation:** Check file size, extension trước khi upload.<br>- **UI/UX:** Hiển thị progress bar khi upload. | `/api/files/upload` |
| **4** | **Access & Download** | **Minh Thức** | - **Trang Download (`/f/:token`):** Gọi API lấy metadata.<br>- **UI Trạng thái:**<br>  + 🟢 Active: Hiện nút download.<br>  + 🟡 Pending: Hiện đồng hồ đếm ngược.<br>  + 🔴 Expired: Hiện thông báo lỗi.<br>- **Security:** Popup nhập Password/TOTP nếu file yêu cầu.<br>- **Action:** Gọi API download (xử lý Blob/Stream). | `/api/files/:shareToken`<br>`/api/files/:shareToken/download` |
| **5** | **Admin System** | **Trung Kiên** | - **Admin Dashboard:** Trang quản trị (Check Role Admin).<br>- **System Policy:** Cấu hình hệ thống (Max size, expire days).<br>- **Cleanup:** UI trigger dọn dẹp file rác.<br>- **Global Config:** Setup Axios Interceptor (gắn Bearer Token tự động cho toàn app). | `/api/admin/policy`<br>`/api/admin/cleanup` |

Bạn Minh Quân xin rút khỏi nhóm.

-----

## 📂 Cấu Trúc Thư Mục (App Router)

Để đảm bảo code gọn gàng và dễ merge, thống nhất cấu trúc như sau:

```bash
app/
├── (auth)/                 # Route Group cho Authentication (Bảo Minh)
│   ├── login/page.tsx
│   ├── register/page.tsx
│   └── totp-setup
│       ├── page.tsx
│       └── layout.tsx
├── dashboard/              # Route Group cho User đã login (Bảo Minh)
│   ├── page.tsx
├── (public)/               # Public Access (Minh Thức)
│   └── f/
│       └── [token]/page.tsx
├── admin/                  # Admin Routes (Trung Kiên)
│   ├── cleanup/page.tsx
│   ├── policy/page.tsx
│   ├── templates/notadmin.tsx
│   ├── page.tsx
│   └── layout.tsx
├── upload/                 # Upload Page (Khánh)
│   └── page.tsx
├── layout.tsx              # Root Layout
└── page.tsx                # Homepage
```

-----

## Quy Tắc Phát Triển (Development Rules)

### 1\. Xử lý API Response & HTTP Codes

Mọi người **BẮT BUỘC** phải xử lý các mã lỗi HTTP đặc thù từ Backend, không chỉ check `status === 200`.

  * **401 Unauthorized:** Redirect về trang Login ngay lập tức.
  * **403 Forbidden:**
      * *Module Download:* Hiển thị input nhập Password hoặc thông báo "Bạn không có quyền".
      * *Module Khác:* Thông báo lỗi toast "Access Denied".
  * **423 Locked (Module Download):** Hiển thị UI "File chưa đến giờ mở" ( kèm thời gian `availableFrom`).
  * **410 Gone (Module Download):** Hiển thị UI "File đã hết hạn hoặc bị xóa".

### 2\. Components & Hooks

  * **API Call:** Không gọi `fetch/axios` trực tiếp trong Component. Hãy tạo file trong `src/services/` (ví dụ: `authService.ts`, `fileService.ts`).
  * **UI Components:** Sử dụng lại các component chung trong `src/components/ui` (Button, Input, Modal...) để đồng bộ giao diện.

### 3\. Git Workflow & Commit Convention

Sử dụng prefix rõ ràng để biết commit thuộc về  module nào:

  * `auth: ...` (Bảo Minh)
  * `dash: ...` (Minh Quan)
  * `upload: ...` (Khánh)
  * `access: ...` (Minh Thức)
  * `admin: ...` (Trung Kiên)

Ví dụ:

> `upload: add validation for availableFrom date`
> `access: handle 410 gone error ui`

-----

## Getting Started

1.  **Clone repository:**
    ```bash
    git clone <repo-url>
    ```
2.  **Install dependencies:**
    ```bash
    npm install
    ```
3.  **Setup Environment:** (Setup later)
    Tạo file `.env.local`:
    ```env
    NEXT_PUBLIC_API_URL=http://localhost:8080/api
    ```
4.  **Run Development Server:**
    ```bash
    npm run dev
    ```
## Usage for Docker

```bash
docker compose up -d
```

To specify the backend api, you can update NEXT_PUBLIC_API_URL in the docker-compose.yaml

## Notes
- Should factor to components
- MUST create pull request for the commit
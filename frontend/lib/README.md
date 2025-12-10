# (New pls read) schemas
Nhất quán chung thư viện schema theo docs openapi mới của Minh viết trong thư mục `@/docs/`. Không bỏ các schema trong component-specific api library, tập trung hết ở `@/lib/components/schemas.ts`.

# Shared axios interceptor
Mục đích: đồng nhất việc request xuống backend, thay vì định nghĩa tạo request trên page.

## client.ts

### Helper
Các phương thức truy cập client side chung, hạn chế tự check trong các page.
```
function safeGetItem(key: string): string | null;
function safeSetItem(key: string, value: string);
function safeRemoveItem(key: string);
```

Các function export:
```
function getAccessToken(): string | null;
function setAccessToken(token: string);
function clearAuth();
function getCurrentUser(): User | null;
function setCurrentUser(user: User);
```

Các function check cơ bản:
```
function _isLoggedIn(): boolean;
function _isAdmin(): boolean;
function isFormData(value: unknown): value is FormData;
```
Local storage này sẽ lưu ở client side, nạp lên rút xuống có nghĩa là server đang kêu client tự đưa, đổi.

### Axios instance
Dùng làm người giao tiếp với backend cho front end. Hạn chế bloat code page bằng implement fetch ở thư mục /lib/api (có ví dụ ở admin.ts)

```
export const adminApi = {
  getPolicy(): Promise<SystemPolicy> {
    return api.get<SystemPolicy>("/admin/policy");
  },

  updatePolicy(payload: SystemPolicyUpdate): Promise<UpdatePolicyResponse> {
    return api.patch<UpdatePolicyResponse>("/admin/policy", payload);
  },

  cleanupExpiredFiles(): Promise<CleanupResponse> {
    return api.post<CleanupResponse>("/admin/cleanup");
  },
};
```

Thống nhất đống cách thức request xuống backend server `const res = await adminApi.cleanupExpiredFiles();`
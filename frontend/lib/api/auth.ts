import { authClient, adminClient } from "./client";
import { clearAuth } from "./helper";
import type {
  LoginRequest,
  LoginResponse,
  RegisterRequest,
  RegisterSuccessResponse,
  TotpSetupResponse,
  TotpVerifyRequest,
  TotpVerifyResponse,
  TotpLoginRequest,
  LoginSuccessResponse,
  UserResponse,
  ChangePasswordRequest,
} from "../components/schemas";

/* ============================================================
 * Helper types
 * ============================================================
 */

type AuthRequest<T> = Promise<T>;

/* ============================================================
 * Validation helpers (FE-side)
 * ============================================================
 */

function validateLoginPayload(payload: LoginRequest): void {
  if (!payload) {
    throw new Error("Login payload is required");
  }

  if (!payload.username || payload.username.trim() === "") {
    throw new Error("Username is required");
  }

  if (!payload.password || payload.password.trim() === "") {
    throw new Error("Password is required");
  }
}

function validateRegisterPayload(payload: RegisterRequest): void {
  if (!payload) {
    throw new Error("Register payload is required");
  }

  if (!payload.username || payload.username.trim() === "") {
    throw new Error("Username is required");
  }

  if (!payload.password || payload.password.length < 6) {
    throw new Error("Password must be at least 6 characters");
  }
}

function validateTotpVerifyPayload(payload: TotpVerifyRequest): void {
  if (!payload || !payload.code) {
    throw new Error("TOTP code is required");
  }
}

function validateChangePasswordPayload(
  payload: ChangePasswordRequest
): void {
  if (!payload) {
    throw new Error("ChangePassword payload is required");
  }

  if (!payload.oldPassword || !payload.newPassword) {
    throw new Error("Old password and new password are required");
  }

  if (payload.newPassword.length < 6) {
    throw new Error("New password must be at least 6 characters");
  }
}

/* ============================================================
 * Response guards (runtime safety)
 * ============================================================
 */

function assertLoginResponse(
  data: LoginResponse
): LoginResponse {
  if (!data || typeof data.requiresTotp !== "boolean") {
    throw new Error("Invalid LoginResponse format");
  }
  return data;
}

function assertLoginSuccessResponse(
  data: LoginSuccessResponse
): LoginSuccessResponse {
  if (!data || typeof data.success !== "boolean") {
    throw new Error("Invalid LoginSuccessResponse format");
  }
  return data;
}

function assertUserResponse(
  data: UserResponse
): UserResponse {
  if (!data || typeof data.id !== "number") {
    throw new Error("Invalid UserResponse format");
  }
  return data;
}

/* ============================================================
 * Auth API implementation
 * ============================================================
 */

/**
 * ------------------------------------------------------------
 * POST /api/auth/login
 * Đăng nhập bằng username/password
 * ------------------------------------------------------------
 */
export async function login(
  payload: LoginRequest
): AuthRequest<LoginResponse> {
  validateLoginPayload(payload);

  const response = await authClient.post<LoginResponse>(
    "/api/auth/login",
    payload
  );

  return assertLoginResponse(response);
}

/**
 * ------------------------------------------------------------
 * POST /api/auth/register
 * Đăng ký tài khoản mới
 * ------------------------------------------------------------
 */
export async function register(
  payload: RegisterRequest
): AuthRequest<RegisterSuccessResponse> {
  validateRegisterPayload(payload);

  const response =
    await authClient.post<RegisterSuccessResponse>(
      "/api/auth/register",
      payload
    );

  return response;
}

/**
 * ------------------------------------------------------------
 * POST /api/auth/totp/setup
 * Khởi tạo TOTP (2FA)
 * ------------------------------------------------------------
 */
export async function setupTotp(): AuthRequest<TotpSetupResponse> {
  const response =
    await authClient.post<TotpSetupResponse>(
      "/api/auth/totp/setup",
      {}
    );

  return response;
}

/**
 * ------------------------------------------------------------
 * POST /api/auth/totp/verify
 * Xác thực TOTP sau khi setup
 * ------------------------------------------------------------
 */
export async function verifyTotp(
  payload: TotpVerifyRequest
): AuthRequest<TotpVerifyResponse> {
  validateTotpVerifyPayload(payload);

  const response =
    await authClient.post<TotpVerifyResponse>(
      "/api/auth/totp/verify",
      payload
    );

  return response;
}

/**
 * ------------------------------------------------------------
 * POST /api/auth/login/totp
 * Đăng nhập bằng TOTP
 * ------------------------------------------------------------
 */
export async function loginTotp(
  payload: TotpLoginRequest
): AuthRequest<LoginSuccessResponse> {
  validateTotpVerifyPayload(payload);

  const response =
    await authClient.post<LoginSuccessResponse>(
      "/api/auth/login/totp",
      payload
    );

  return assertLoginSuccessResponse(response);
}

/**
 * ------------------------------------------------------------
 * GET /api/user
 * Lấy profile user hiện tại
 * ------------------------------------------------------------
 */
export async function getUserProfile(): AuthRequest<UserResponse> {
  const response =
    await authClient.get<UserResponse>("/api/user");

  return assertUserResponse(response);
}

/**
 * ------------------------------------------------------------
 * POST /api/auth/totp/disable
 * Tắt TOTP
 * ------------------------------------------------------------
 */
export async function disableTotp(
  code: string
): AuthRequest<void> {
  if (!code || code.trim() === "") {
    throw new Error("TOTP code is required");
  }

  await authClient.post(
    "/api/auth/totp/disable",
    { code }
  );
}

/**
 * ------------------------------------------------------------
 * POST /api/auth/password/change
 * Đổi mật khẩu
 * ------------------------------------------------------------
 */
export async function changePassword(
  payload: ChangePasswordRequest
): AuthRequest<void> {
  validateChangePasswordPayload(payload);

  await authClient.post(
    "/api/auth/password/change",
    payload
  );
}

/**
 * ------------------------------------------------------------
 * Logout (FE-side)
 * ------------------------------------------------------------
 */
export function logout(): void {
  clearAuth();
}

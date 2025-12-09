import { authClient, adminClient } from "./client";
import { clearAuth } from "./helper";
import { 
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
  ChangePasswordRequest
} from "../components/schemas";

export const login = (payload: LoginRequest) =>
  authClient.post<LoginResponse>("/api/auth/login", payload);
  
export const register = (payload: RegisterRequest) =>
  authClient.post<RegisterSuccessResponse>("/api/auth/register", payload);

export const setupTotp = () => 
  authClient.post<TotpSetupResponse>("/api/auth/totp/setup", {});

export const verifyTotp = (payload: TotpVerifyRequest) =>
  authClient.post<TotpVerifyResponse>("/api/auth/totp/verify", payload);

export const loginTotp = (payload: TotpLoginRequest) =>
  authClient.post<LoginSuccessResponse>("/api/auth/login/totp", payload);

export const getUserProfile = () => 
  authClient.get<UserResponse>("/api/user");

export const disableTotp = (code: string) =>
  authClient.post<any>("/api/auth/totp/disable", { code });

export const changePassword = (payload: ChangePasswordRequest) =>
  authClient.post<any>("/api/auth/password/change", payload);

export const logout = () => {
  clearAuth();
};

import axios, {
  AxiosError,
  InternalAxiosRequestConfig,
  AxiosResponse,
  AxiosInstance,
  AxiosRequestConfig,
} from "axios";
import { toast } from "sonner";
import { getAccessToken, getAdminToken, clearAuth } from "@/lib/api/helper";


const getBaseUrl = (): string => {
  const envUrl = typeof process !== 'undefined' && process.env?.NEXT_PUBLIC_API_URL;
  if (!envUrl || envUrl.trim() === '') {
    return ''; // Empty string = relative path (same origin)
  }
  // Remove trailing slash if present
  return envUrl.endsWith('/') ? envUrl.slice(0, -1) : envUrl;
};

const BASE_URL = getBaseUrl();
const DEFAULT_TIMEOUT = 15000; // 15s for normal requests
const UPLOAD_TIMEOUT = 10 * 60 * 1000; // 10 Minutes for large files [cite: 361, 373]

interface ErrorResponse {
  message: string;
}

interface ApiClient extends AxiosInstance {
  get<T>(url: string, config?: AxiosRequestConfig): Promise<T>;
  post<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>;
  put<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>;
  patch<T>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T>;
  delete<T>(url: string, config?: AxiosRequestConfig): Promise<T>;
}

// 1a. Token Injector for user-protected APIs
const attachTokenInterceptor = (config: InternalAxiosRequestConfig) => {
  const token = getAccessToken();
  if (token) {
    config.headers.set("Authorization", `Bearer ${token}`);
  }
  return config;
};

// 1b. Token Injector for admin APIs (shared secret)
const attachAdminTokenInterceptor = (config: InternalAxiosRequestConfig) => {
  const adminToken = getAdminToken();
  if (adminToken) {
    config.headers.set("Authorization", `Bearer ${adminToken}`);
  }
  return config;
};

// 2. Global Error Handler (Used by all clients)
const errorResponseInterceptor = (error: AxiosError<ErrorResponse>) => {
  if (error.response) {
    const { status, data } = error.response;

    if (status === 401) {
      // Token expired handling - only redirect if not on public pages
      const currentPath = typeof window !== 'undefined' ? window.location.pathname : '';
      const publicPaths = ['/login', '/register', '/', '/f/'];
      const isPublicPage = publicPaths.some(path => currentPath.startsWith(path));
      
      clearAuth();
      
      // Only redirect and show toast if on protected page
      if (!isPublicPage) {
        toast.error("Session expired. Please login again.");
        window.location.href = "/login";
      }
      // Silently fail on public pages (user not logged in is expected)
    } else if (status === 403) {
      toast.error(data?.message || "Access denied.");
    } else if (status === 413) {
      toast.error("File is too large."); // Specific to upload [cite: 173]
    } else if (status >= 500) {
      toast.error("System error. Please try again later.");
    }
  } else {
    toast.error("Network error. Please check your connection.");
  }
  return Promise.reject(error);
};

// 3. Data Unwrap (Optional, purely preference)
const responseDataInterceptor = (response: AxiosResponse) => response.data;

// --- Client Definitions ---


const authClient: ApiClient = axios.create({
  baseURL: BASE_URL,
  timeout: DEFAULT_TIMEOUT,
  headers: {
    "Content-Type": "application/json",
  },
});
// Attach token so /api/user, etc. carry Authorization
authClient.interceptors.request.use(attachTokenInterceptor, Promise.reject);
authClient.interceptors.response.use(responseDataInterceptor, errorResponseInterceptor);

/**
 * 2. Admin Client
 * - REQUIRES Admin Token (shared secret)
 * - Enforces JSON Content-Type
 * - Standard Timeout
 * - Use for: /api/admin/*
 */
const adminClient: ApiClient = axios.create({
  baseURL: BASE_URL,
  timeout: DEFAULT_TIMEOUT,
  headers: {
    "Content-Type": "application/json",
  },
});
adminClient.interceptors.request.use(attachAdminTokenInterceptor, Promise.reject);
adminClient.interceptors.response.use(responseDataInterceptor, errorResponseInterceptor);


/**
 * 3. Upload Client
 * - REQUIRES Authorization Token
 * - Enforces Multipart/Form-Data
 * - LONG Timeout (for 50MB-100MB files) 
 * - Use for: /api/files/upload [cite: 126]
 */
const uploadClient: ApiClient = axios.create({
  baseURL: BASE_URL,
  timeout: UPLOAD_TIMEOUT,
  headers: {
    // Explicitly set multipart (though Axios can detect FormData, this ensures safety)
    "Content-Type": "multipart/form-data",
  },
});
uploadClient.interceptors.request.use(attachTokenInterceptor, Promise.reject);
uploadClient.interceptors.response.use(responseDataInterceptor, errorResponseInterceptor);

export {
  authClient,
  adminClient,
  uploadClient,
};
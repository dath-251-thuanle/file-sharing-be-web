import { User } from "@/lib/components/schemas";
import Cookies from "js-cookie";

const ACCESS_TOKEN_KEY = "fs_access_token";
const USER_KEY = "fs_user";
const ACCESS_TOKEN_COOKIE = "fs_access_token"; // Cookie name for server-side access
const ADMIN_TOKEN_KEY = "fs_admin_token";
const CID_KEY = "fs_login_cid";

/**
 * Helpers: localStorage-safe (avoid SSR crash)
 */
function safeGetItem(key: string): string | null {
    if (typeof window === "undefined") return null;
    return localStorage.getItem(key);
}

function safeSetItem(key: string, value: string) {
    if (typeof window === "undefined") return;
    localStorage.setItem(key, value);
}

function safeRemoveItem(key: string) {
    if (typeof window === "undefined") return;
    localStorage.removeItem(key);
}

/**
 * Auth storage helpers
 */
export function getAccessToken(): string | null {
    return safeGetItem(ACCESS_TOKEN_KEY);
}

export function setAccessToken(token: string) {
    // Store in localStorage for client-side access
    safeSetItem(ACCESS_TOKEN_KEY, token);
    // Store in cookie for server-side access (API routes)
    // Set cookie with 7 days expiry, httpOnly=false so client can read it
    // SameSite=Lax for CSRF protection, Secure in production
    Cookies.set(ACCESS_TOKEN_COOKIE, token, {
        expires: 7, // 7 days
        sameSite: 'lax',
        secure: process.env.NODE_ENV === 'production',
        path: '/',
    });
}

// Admin token helpers (for calling /api/admin/* with shared secret)
export function getAdminToken(): string | null {
    return safeGetItem(ADMIN_TOKEN_KEY);
}

export function setAdminToken(token: string) {
    safeSetItem(ADMIN_TOKEN_KEY, token);
}

export function clearAdminToken() {
    safeRemoveItem(ADMIN_TOKEN_KEY);
}

export function clearAuth() {
    safeRemoveItem(ACCESS_TOKEN_KEY);
    safeRemoveItem(USER_KEY);
    safeRemoveItem(CID_KEY);
    // Remove cookie
    Cookies.remove(ACCESS_TOKEN_COOKIE, { path: '/' });
}

export function setLoginChallengeId(cid: string) {
    safeSetItem(CID_KEY, cid);
}

export function getLoginChallengeId(): string | null {
    return safeGetItem(CID_KEY);
}

export function clearLoginChallengeId() {
    safeRemoveItem(CID_KEY);
}

export function getCurrentUser(): User | null {
    const raw = safeGetItem(USER_KEY);
    if (!raw) return null;
    try {
        return JSON.parse(raw) as User;
    } catch {
        return null;
    }
}

export function setCurrentUser(user: User) {
    safeSetItem(USER_KEY, JSON.stringify(user));
}

export function _isLoggedIn(): boolean {
    // return true;
    return !!getAccessToken();
}

export function _isAdmin(): boolean {
    // return true;
    const user = getCurrentUser();
    console.log(user);
    return user?.role === "admin";
}

export function getErrorMessage(error: any, defaultMessage: string = "An error occurred"): string {
    if (!error) return defaultMessage;

    if (typeof error === "string") return error;

    // Axios error with response
    if (error.response) {
        const data = error.response.data;
        // If data is simple string
        if (typeof data === "string") return data;
        
        // If data is object
        if (data && typeof data === "object") {
            // Prioritize 'message' then 'error'
            if (data.message) return data.message;
            if (data.error) return data.error;
        }
    }

    // Fallback to error message or default
    return error.message || defaultMessage;
}

export function isFormData(value: unknown): value is FormData {
    if (typeof FormData === "undefined" || !value) {
        return false;
    }

    return value instanceof FormData;
}

/**
 * Enhanced fetch that automatically adds Authorization token if available
 * Use this instead of native fetch() to ensure token is always sent
 */
export async function authenticatedFetch(
    url: string | URL | Request,
    init?: RequestInit
): Promise<Response> {
    const token = getAccessToken();
    
    // Convert headers to Headers object if needed
    const headers = new Headers();
    
    // Copy existing headers
    if (init?.headers) {
        if (init.headers instanceof Headers) {
            init.headers.forEach((value, key) => {
                headers.set(key, value);
            });
        } else if (Array.isArray(init.headers)) {
            init.headers.forEach(([key, value]) => {
                headers.set(key, value);
            });
        } else {
            // Plain object
            Object.entries(init.headers).forEach(([key, value]) => {
                if (value) {
                    headers.set(key, value);
                }
            });
        }
    }
    
    // Automatically add Authorization header if token exists and not already set
    if (token && !headers.has('Authorization')) {
        headers.set('Authorization', `Bearer ${token}`);
        console.log('[authenticatedFetch] Added Authorization token automatically');
        console.log('[authenticatedFetch] Token preview:', token.substring(0, 20) + '...');
    } else if (token && headers.has('Authorization')) {
        console.log('[authenticatedFetch] Authorization header already exists, not overwriting');
    } else if (!token) {
        console.log('[authenticatedFetch] No token available, not adding Authorization header');
    }
    
    // Log all headers being sent
    console.log('[authenticatedFetch] Final headers:', Array.from(headers.entries()).map(([k, v]) => [k, k === 'Authorization' ? v.substring(0, 20) + '...' : v]));
    
    // Merge with existing init options
    const enhancedInit: RequestInit = {
        ...init,
        headers: headers,
        credentials: init?.credentials || 'include',
    };
    
    return fetch(url, enhancedInit);
}
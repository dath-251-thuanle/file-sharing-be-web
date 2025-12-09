import { uploadClient, adminClient, authClient } from "@/lib/api/client";
import { getAccessToken } from "@/lib/api/helper";
import type { FileUploadResponse, UploadedFileSummary, UserFilesResponse, FileInfoResponse } from "@/lib/components/schemas";

const downloadLinkFromToken = (shareToken?: string): string => {
  if (!shareToken) return "";
  // Get base URL from current window location (works in browser)
  const baseUrl = typeof window !== 'undefined' 
    ? window.location.origin 
    : '';
  return `${baseUrl}/api/files/${shareToken}/download`;
};

export class ApiError extends Error {
  status: number;
  data: unknown;

  constructor(status: number, message: string, data?: unknown) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.data = data;
  }
}

type AxiosErrorLike = {
  response?: {
    status?: number;
    data?: unknown;
  };
};

function isAxiosErrorLike(error: unknown): error is AxiosErrorLike {
  return typeof error === "object" && error !== null && "response" in error;
}

type LegacyUploadResponse = {
  fileId?: string;
  fileName?: string;
  shareLink?: string;
  availableFrom?: string;
  availableTo?: string;
  sharedWith?: string[];
  expiresAt?: string;
};

type ModernUploadResponse = {
  success?: boolean;
  message?: string;
  file?: Partial<UploadedFileSummary>;
};

function isModernUploadResponse(payload: unknown): payload is ModernUploadResponse {
  if (typeof payload !== "object" || payload === null) {
    return false;
  }

  const candidate = payload as ModernUploadResponse;
  return typeof candidate.file === "object" && candidate.file !== null;
}

function toUploadedFileSummary(data: Partial<UploadedFileSummary>): UploadedFileSummary {
  return {
    id: data.id ?? data.shareToken ?? "",
    fileName: data.fileName ?? "",
    shareLink: data.shareLink ?? downloadLinkFromToken(data.shareToken),
    shareToken: data.shareToken,
    isPublic: data.isPublic,
    hasPassword: data.hasPassword,
    availableFrom: data.availableFrom,
    availableTo: data.availableTo,
    sharedWith: data.sharedWith,
    expiresAt: data.expiresAt,
  };
}

function toUploadFileResponse(raw: unknown): FileUploadResponse {
  if (isModernUploadResponse(raw)) {
    return {
      success: raw.success ?? true,
      message: raw.message,
      file: toUploadedFileSummary(raw.file ?? {}),
    };
  }

  const legacy = raw as LegacyUploadResponse | null;
  if (legacy && (legacy.fileId || legacy.fileName || legacy.shareLink)) {
    return {
      success: true,
      file: {
        id: legacy.fileId ?? legacy.shareLink ?? "",
        fileName: legacy.fileName ?? "",
        shareLink: legacy.shareLink ?? "",
        availableFrom: legacy.availableFrom,
        availableTo: legacy.availableTo,
        sharedWith: legacy.sharedWith,
        expiresAt: legacy.expiresAt,
      },
    };
  }

  throw new ApiError(500, "Phản hồi upload không hợp lệ", raw);
}

async function uploadFile(formData: FormData): Promise<FileUploadResponse> {
  try {
    const response = await uploadClient.post("/api/files/upload", formData, {
      headers: { "Content-Type": "multipart/form-data" },
      withCredentials: true,
    });

    return toUploadFileResponse(response);
  } catch (error: unknown) {
    if (isAxiosErrorLike(error)) {
      const status = error.response?.status ?? 500;
      const data = error.response?.data;
      const message =
        typeof data === "object" && data !== null && "message" in data && typeof (data as { message?: unknown }).message === "string"
          ? String((data as { message?: unknown }).message)
          : "Upload file thất bại";

      throw new ApiError(status, message, data);
    }

    if (error instanceof ApiError) {
      throw error;
    }

    const fallbackMessage = error instanceof Error ? error.message : "Upload file thất bại";
    throw new ApiError(500, fallbackMessage, error);
  }
}

async function deleteFile(fileId: string): Promise<any> {
  // Backend defines DELETE /api/files/info/:id for owner/admin delete
  return authClient.delete(`/api/files/info/${fileId}`);
}

async function getUserFiles(params?: {
  status?: string;
  page?: number;
  limit?: number;
  sortBy?: string;
  order?: string;
}): Promise<UserFilesResponse> {
  return authClient.get<UserFilesResponse>("/api/files/my", { params });
}

async function getFileInfoById(fileId: string): Promise<FileInfoResponse> {
  return authClient.get<FileInfoResponse>(`/api/files/info/${fileId}`);
}

async function getFileInfoByToken(shareToken: string): Promise<FileInfoResponse> {
  // Try with auth token first (to get owner details like sharedWith)
  // If no token, fallback to public endpoint
  const token = getAccessToken();
  if (token) {
    try {
      return authClient.get<FileInfoResponse>(`/api/files/${shareToken}`);
    } catch (err: any) {
      // If auth fails (401), fallback to public endpoint
      // This handles cases where token is invalid/expired but we still want to show file info
      if (err?.response?.status === 401) {
        console.log('[getFileInfoByToken] Auth failed, falling back to public endpoint');
        return authClient.get<FileInfoResponse>(`/api/files/${shareToken}`);
      }
      throw err;
    }
  }
  return authClient.get<FileInfoResponse>(`/api/files/${shareToken}`);
}

export const fileApi = {
  upload: uploadFile,
  delete: deleteFile,
  getUserFiles: getUserFiles,
  getFileInfoById: getFileInfoById,
  getFileInfoByToken: getFileInfoByToken,
};

export { uploadFile, deleteFile, getUserFiles, getFileInfoById, getFileInfoByToken };

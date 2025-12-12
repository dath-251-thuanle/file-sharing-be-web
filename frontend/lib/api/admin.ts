import { adminClient } from "@/lib/api/client";
import type {
  SystemPolicy,
  SystemPolicyUpdate,
  UpdatePolicyResponse,
  CleanupResponse,
} from "../components/schemas";

/* ============================================================
 * Helper types
 * ============================================================
 */

/**
 * Wrapper chung cho mọi admin request
 * Giúp dễ mở rộng logging / retry sau này
 */
type AdminRequest<T> = Promise<T>;

/* ============================================================
 * Validation helpers (FE-side safety)
 * ============================================================
 */

/**
 * Validate payload trước khi gửi lên backend
 * (giảm rủi ro 400, đặc biệt khi form phức tạp)
 */
function validatePolicyUpdate(payload: SystemPolicyUpdate): void {
  if (!payload) {
    throw new Error("SystemPolicyUpdate payload is required");
  }

  // Ví dụ một số check phổ biến (tùy schema của bạn)
  if (
    typeof payload.maxFileSizeMB === "number" &&
    payload.maxFileSizeMB <= 0
  ) {
    throw new Error("maxFileSizeMB must be greater than 0");
  }

  if (
    typeof payload.fileTTLHours === "number" &&
    payload.fileTTLHours <= 0
  ) {
    throw new Error("fileTTLHours must be greater than 0");
  }
}

/* ============================================================
 * Response guards (optional nhưng rất nên có)
 * ============================================================
 */

function assertUpdatePolicyResponse(
  data: UpdatePolicyResponse
): UpdatePolicyResponse {
  if (!data || typeof data.success !== "boolean") {
    throw new Error("Invalid UpdatePolicyResponse format");
  }
  return data;
}

function assertCleanupResponse(
  data: CleanupResponse
): CleanupResponse {
  if (!data || typeof data.deletedCount !== "number") {
    throw new Error("Invalid CleanupResponse format");
  }
  return data;
}

/* ============================================================
 * Admin API implementation
 * ============================================================
 */

export const adminApi = {
  /**
   * ------------------------------------------------------------
   * GET /api/admin/policy
   * Lấy system policy hiện tại
   * ------------------------------------------------------------
   */
  async getPolicy(): AdminRequest<SystemPolicy> {
    // Gửi request
    const response = await adminClient.get<SystemPolicy>(
      "/api/admin/policy"
    );

    // Có thể thêm transform nếu backend đổi format
    return response;
  },

  /**
   * ------------------------------------------------------------
   * PATCH /api/admin/policy
   * Cập nhật system policy
   * ------------------------------------------------------------
   */
  async updatePolicy(
    payload: SystemPolicyUpdate
  ): AdminRequest<UpdatePolicyResponse> {
    // Validate FE-side
    validatePolicyUpdate(payload);

    // Gửi request
    const response = await adminClient.patch<UpdatePolicyResponse>(
      "/api/admin/policy",
      payload
    );

    // Check format response
    return assertUpdatePolicyResponse(response);
  },

  /**
   * ------------------------------------------------------------
   * POST /api/admin/cleanup
   * Cleanup các file đã hết hạn
   * ------------------------------------------------------------
   */
  async cleanupExpiredFiles(): AdminRequest<CleanupResponse> {
    // Gửi request
    const response = await adminClient.post<CleanupResponse>(
      "/api/admin/cleanup"
    );

    // Validate response
    return assertCleanupResponse(response);
  },
};

/* ============================================================
 * Default export (tuỳ style project)
 * ============================================================
 */

export default adminApi;

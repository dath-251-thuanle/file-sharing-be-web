"use client";

import { useState } from "react";
import { adminApi } from "@/lib/api/admin";
import { toast } from "sonner";

type CleanupButtonProps = {
  variant?: "solid" | "outline";
  className?: string;
};

export function CleanupButton({
  variant = "outline",
  className = "",
}: CleanupButtonProps) {
  const [loading, setLoading] = useState(false);

  const handleCleanup = async () => {
    setLoading(true);

    try {
      const res = await adminApi.cleanupExpiredFiles();
      const timeStr = new Date(res.timestamp).toLocaleString();
      toast.success(
        `${res.message} • Đã xoá: ${res.deletedFiles} file • ${timeStr}`
      );
    } catch (err: any) {
      console.error(err);
      const msg =
        err?.response?.data?.message ||
        "Không thể cleanup file hết hạn. Vui lòng thử lại.";
      toast.error(msg);
    } finally {
      setLoading(false);
    }
  };

  const baseClasses =
    "rounded-md px-3 py-2 text-sm font-medium shadow-sm disabled:cursor-not-allowed disabled:opacity-60";
  const variantClasses =
    variant === "solid"
      ? "bg-blue-600 text-white hover:bg-blue-700"
      : "border hover:bg-gray-50";

  return (
    <button
      type="button"
      onClick={handleCleanup}
      disabled={loading}
      className={`${baseClasses} ${variantClasses} ${className}`}
    >
      {loading ? "Đang cleanup..." : "Cleanup file hết hạn"}
    </button>
  );
}

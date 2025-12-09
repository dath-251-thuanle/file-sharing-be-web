"use client";

import { AdminShell } from "@/components/admin/AdminShell";
import { CleanupButton } from "@/components/admin/CleanupButton";

export default function AdminCleanupPage() {
  return (
    <AdminShell
      title="Cleanup file hết hạn"
      description="Thực hiện dọn dẹp thủ công các file đã hết hạn lưu trữ trong hệ thống."
    >
      <div className="space-y-4 rounded-xl border bg-white p-5 shadow-sm">
        <p className="text-s text-gray-600">
          Hệ thống đã có cơ chế auto-cleanup theo lịch trình.
        </p>

        <p className="text-sm text-gray-500">
          Nhấn nút này để kiểm tra và xóa các file đã quá hạn.
        </p>

        <div className="pt-2">
          <CleanupButton variant="solid" />
        </div>
      </div>
    </AdminShell>
  );
}

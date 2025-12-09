"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { SystemPolicy, SystemPolicyUpdate } from "@/lib/components/schemas"
import { adminApi } from "@/lib/api/admin";
import { toast } from "sonner";
import { AdminShell } from "@/components/admin/AdminShell";
import { AdminTokenBar } from "@/components/admin/AdminTokenBar";
import { NumberField } from "@/components/admin/NumberField";
import { CleanupButton } from "@/components/admin/CleanupButton";

const initialForm = (policy: SystemPolicy): SystemPolicyUpdate => ({
  maxFileSizeMB: policy.maxFileSizeMB,
  minValidityHours: policy.minValidityHours,
  maxValidityDays: policy.maxValidityDays,
  defaultValidityDays: policy.defaultValidityDays,
  requirePasswordMinLength: policy.requirePasswordMinLength,
});

export default function AdminPolicyPage() {
  const router = useRouter();
  const [policy, setPolicy] = useState<SystemPolicy | null>(null);
  const [form, setForm] = useState<SystemPolicyUpdate | null>(null);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);

  // Check admin + load policy
  useEffect(() => {
    adminApi
      .getPolicy()
      .then((data) => {
        console.log(data);
        setPolicy(data);
        setForm(initialForm(data));
      })
      .catch((err) => {
        console.error(err);
        const status = err?.response?.status;
        if (status === 401 || status === 403) {
          toast.error("Nhập ADMIN_API_TOKEN (Bearer) rồi bấm Lưu token, sau đó thử lại.");
        } else {
          toast.error("Không thể tải System Policy. Vui lòng thử lại.");
        }
      })
      .finally(() => setLoading(false));
  }, [router]);

  const handleChange =
    (field: keyof SystemPolicyUpdate) =>
      (value: number | undefined) => {
        setForm((prev) => ({
          ...(prev ?? {}),
          [field]: value,
        }));
      };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!form) return;

    setSaving(true);

    try {
      const res = await adminApi.updatePolicy(form);
      setPolicy(res.policy);
      setForm(initialForm(res.policy));
      toast.success(res.message || "Cập nhật System Policy thành công.");
    } catch (err: any) {
      console.error(err);
      const msg =
        err?.response?.data?.message ||
        "Cập nhật thất bại. Vui lòng kiểm tra dữ liệu và thử lại.";
      toast.error(msg);
    } finally {
      setSaving(false);
    }
  };

  if (loading) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <p className="text-sm text-gray-500">Đang tải System Policy...</p>
      </div>
    );
  }

  if (!policy || !form) {
    return (
      <div className="p-6 space-y-4">
        <h1 className="mb-2 text-xl font-semibold">System Policy</h1>
        <AdminTokenBar />
        <p className="text-sm text-red-500">
          Không thể tải System Policy. Vui lòng nhập ADMIN_API_TOKEN và thử lại.
        </p>
        <button
          onClick={() => window.location.reload()}
          className="inline-flex items-center px-4 py-2 rounded-md bg-blue-600 text-white text-sm font-medium hover:bg-blue-700"
        >
          Thử lại
        </button>
      </div>
    );
  }

  return (
    <AdminShell
      title="System Policy"
      description="Cấu hình giới hạn hệ thống cho dịch vụ chia sẻ file."
      actions={<CleanupButton />}
    >
      <form
        onSubmit={handleSubmit}
        className="space-y-5 rounded-xl border bg-white p-5 shadow-sm"
      >
        <NumberField
          label="Max file size (MB)"
          description="Giới hạn kích thước file tối đa mà user có thể upload."
          value={form.maxFileSizeMB}
          min={1}
          onChange={handleChange("maxFileSizeMB")}
        />

        <NumberField
          label="Min validity (giờ)"
          description="Số giờ tối thiểu mà một link có thể tồn tại."
          value={form.minValidityHours}
          min={1}
          onChange={handleChange("minValidityHours")}
        />

        <NumberField
          label="Max validity (ngày)"
          description="Số ngày tối đa mà một link có thể tồn tại."
          value={form.maxValidityDays}
          min={1}
          onChange={handleChange("maxValidityDays")}
        />

        <NumberField
          label="Default validity (ngày)"
          description="Giá trị mặc định khi user tạo link mới (nếu không chọn)."
          value={form.defaultValidityDays}
          min={1}
          onChange={handleChange("defaultValidityDays")}
        />

        <NumberField
          label="Min password length (nếu có mật khẩu)"
          description="Độ dài tối thiểu mật khẩu bảo vệ link chia sẻ."
          value={form.requirePasswordMinLength}
          min={1}
          onChange={handleChange("requirePasswordMinLength")}
        />

        <div className="flex items-center justify-end gap-3">
          <button
            type="button"
            onClick={() => policy && setForm(initialForm(policy))}
            className="rounded-md border px-3 py-2 text-sm font-medium hover:bg-gray-50"
          >
            Reset
          </button>
          <button
            type="submit"
            disabled={saving}
            className="rounded-md bg-blue-600 px-4 py-2 text-sm font-semibold text-white shadow-sm hover:bg-blue-700 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {saving ? "Đang lưu..." : "Lưu System Policy"}
          </button>
        </div>
      </form>
    </AdminShell>
  );
}

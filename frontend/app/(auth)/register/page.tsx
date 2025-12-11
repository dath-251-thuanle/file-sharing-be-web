"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import RegisterForm, {
  RegisterFormData,
} from "@/components/auth/RegisterForm";
import { register } from "@/lib/api/auth";
import { getErrorMessage } from "@/lib/api/helper";
import { getPolicyLimits } from "@/lib/api/policy";
import type { PolicyLimits } from "@/lib/components/schemas";

export default function RegisterPage() {
  const [formData, setFormData] = useState<RegisterFormData>({
    username: "",
    email: "",
    password: "",
    confirmPassword: "",
  });
  const [policy, setPolicy] = useState<PolicyLimits | null>(null);

  const router = useRouter();

  const updateField = (field: keyof RegisterFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  useEffect(() => {
    getPolicyLimits()
      .then((p) => setPolicy(p))
      .catch(() => {
        toast.error("Không tải được giới hạn hệ thống, dùng mặc định.");
      });
  }, []);

  const minPasswordLength = policy?.requirePasswordMinLength ?? 8;

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (formData.password !== formData.confirmPassword) {
      toast.error("Mật khẩu và xác nhận mật khẩu không khớp.");
      return;
    }
    if (formData.password.length < minPasswordLength) {
      toast.error(`Mật khẩu phải có tối thiểu ${minPasswordLength} ký tự.`);
      return;
    }
    try {
        await register({
          username: formData.username,
          email: formData.email,
          password: formData.password,
        });

        toast.success("Đăng ký thành công! Vui lòng đăng nhập.");
        router.push("/login");
    } catch (err: unknown) {
        const msg = getErrorMessage(err, "Không thể đăng ký. Vui lòng kiểm tra thông tin và thử lại.");
        toast.error(msg);
    }
  };

  return (
    <RegisterForm
      formData={formData}
      updateField={updateField}
      handleSubmit={handleSubmit}
      minPasswordLength={minPasswordLength}
    />
  );
}
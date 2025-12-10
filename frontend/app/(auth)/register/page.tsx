"use client";

import { useState, Component } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import RegisterForm, {
  RegisterFormData,
} from "@/components/auth/RegisterForm";
import { register } from "@/lib/api/auth";
import { getErrorMessage } from "@/lib/api/helper";

export default function RegisterPage() {
  const [formData, setFormData] = useState<RegisterFormData>({
    username: "",
    email: "",
    password: "",
    confirmPassword: "",
  });

  const router = useRouter();

  const updateField = (field: keyof RegisterFormData, value: any) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (formData.password !== formData.confirmPassword) {
      toast.error("Mật khẩu và xác nhận mật khẩu không khớp.");
      return;
    }
    try {
        const res = await register({
          username: formData.username,
          email: formData.email,
          password: formData.password,
        });

        toast.success("Đăng ký thành công! Vui lòng đăng nhập.");
        router.push("/login");
    } catch (err: any) {
        const msg = getErrorMessage(err, "Sai thông tin đăng nhập. Vui lòng thử lại.");
        toast.error(msg);
    }
  };

  return (
    <RegisterForm
      formData={formData}
      updateField={updateField}
      handleSubmit={handleSubmit}
    />
  );
}
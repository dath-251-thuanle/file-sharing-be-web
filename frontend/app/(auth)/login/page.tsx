"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { login } from "@/lib/api/auth";
import { setAccessToken, setCurrentUser, getErrorMessage, setLoginChallengeId } from "@/lib/api/helper";
import LoginForm, { LoginFormData } from "@/components/auth/LoginForm";

export default function LoginPage() {
        const router = useRouter();
        const [formData, setFormData] = useState<LoginFormData>({
            email: "",
            password: "",
        });

        const updateField = (field: keyof LoginFormData, value: any) => {
            setFormData((prev) => ({ ...prev, [field]: value }));
        };

        const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
            e.preventDefault();
            try {
                const res = await login({
                    email: formData.email,
                    password: formData.password,
                });

                if ("requireTOTP" in res) {
                    // Handle TOTP verification
                    setLoginChallengeId(res.cid);
                    router.push(`/login/totp?email=${encodeURIComponent(formData.email)}`);
                }
                else if ("accessToken" in res) {
                    setAccessToken(res.accessToken);
                    setCurrentUser(res.user);
                    toast.success("Đăng nhập thành công!");
                    router.push("/dashboard");
                }
                else {
                    toast.error("Invalid credentials. Please try again.");
                }
            } catch (err: any) {
                const msg = getErrorMessage(err, "Sai thông tin đăng nhập. Vui lòng thử lại.");
                toast.error(msg);
            }
        };
        return (
            <LoginForm
                formData={formData}
                updateField={updateField}
                handleSubmit={handleSubmit}
            />
        );
    
}
"use client";

import { useState, useEffect, Suspense } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { toast } from "sonner";
import { loginTotp } from "@/lib/api/auth";
import { setAccessToken, setCurrentUser, getErrorMessage, getLoginChallengeId, clearLoginChallengeId } from "@/lib/api/helper";
import LoginTotpForm, {
  LoginTotpFormData,
} from "@/components/auth/LoginTotpForm";

function LoginTotpPageContent() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const email = searchParams.get("email") || "";
  
  const [cid, setCid] = useState<string | null>(null);
  const [formData, setFormData] = useState<LoginTotpFormData>({
    code: "",
  });
  const [verifying, setVerifying] = useState(false);

  useEffect(() => {
    const storedCid = getLoginChallengeId();
    if (!storedCid) {
      toast.error("Invalid session. Please login again.");
      router.push("/login");
    } else {
      setCid(storedCid);
    }
  }, [router]);

  const updateField = (field: keyof LoginTotpFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    if (!cid) {
      toast.error("Invalid session. Please login again.");
      router.push("/login");
      return;
    }

    setVerifying(true);
    try {
      const res = await loginTotp({
        cid,
        code: formData.code,
      });

      if ("accessToken" in res) {
        setAccessToken(res.accessToken);
        setCurrentUser(res.user);
        clearLoginChallengeId();
        toast.success("Đăng nhập thành công!");
        router.push("/dashboard");
      } else {
        toast.error("Invalid credentials. Please try again.");
      }
    } catch (err: any) {
      const msg = getErrorMessage(err, "Sai mã TOTP hoặc đã hết hạn. Vui lòng thử lại.");
      toast.error(msg);
    } finally {
      setVerifying(false);
    }
  };

  return (
    <LoginTotpForm
      email={email}
      formData={formData}
      updateField={updateField}
      handleSubmit={handleSubmit}
      verifying={verifying}
    />
  );
}

export default function LoginTotpPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <LoginTotpPageContent />
    </Suspense>
  );
}

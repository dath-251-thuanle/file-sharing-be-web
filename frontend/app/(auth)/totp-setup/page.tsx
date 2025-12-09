"use client";

import { useState, useEffect } from "react";
import { useRouter } from "next/navigation";
import { toast } from "sonner";
import { setupTotp, verifyTotp } from "@/lib/api/auth";
import { getAccessToken, getErrorMessage } from "@/lib/api/helper";
import TotpSetupForm,{ TotpSetupFormData } from "@/components/auth/TotpSetupForm";

export default function TotpSetupPage() {
  const router = useRouter();
  const [qrCode, setQrCode] = useState<string | null>(null);
  const [secret, setSecret] = useState<string | null>(null);
  const [formData, setFormData] = useState<TotpSetupFormData>({
    totpCode: "",
  });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [verifying, setVerifying] = useState(false);

  const updateField = (field: keyof TotpSetupFormData, value: string) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  useEffect(() => {
    const fetchTotpSetup = async () => {
      const token = getAccessToken();

      if (!token) {
        toast.error("You must be logged in to set up TOTP.");
        router.push("/login");
        return;
      }

      try {
        setLoading(true);
        const data = await setupTotp();
        setQrCode(data.totpSetup.qrCode);
        setSecret(data.totpSetup.secret);
        setError(null);
      } catch (err: any) {
        const msg = getErrorMessage(err, "Failed to fetch TOTP setup information.");
        setError(msg);
        toast.error(msg);
      } finally {
        setLoading(false);
      }
    };

    fetchTotpSetup();
  }, [router]);

  const handleVerify = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setVerifying(true);
    setError(null);

    try {
      const res = await verifyTotp({ code: formData.totpCode });
      if (res.totpEnabled) {
        toast.success("Two-Factor Authentication enabled successfully!");
        router.push("/dashboard"); // Redirect to dashboard or home
      } else {
        throw new Error(res.message || "Verification failed. Please try again.");
      }
    } catch (err: any) {
      const msg = getErrorMessage(err, "Invalid TOTP code. Please try again.");
      setError(msg);
      toast.error(msg);
    }
    finally {
      setVerifying(false);
    }
  };

  return (
    <TotpSetupForm
      formData={formData}
      updateField={updateField}
      handleSubmit={handleVerify}
      qrCode={qrCode}
      secret={secret}
      loading={loading}
      error={error}
      verifying={verifying}
    />
  );
}

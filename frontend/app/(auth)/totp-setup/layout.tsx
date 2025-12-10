"use client";

import type { ReactNode } from "react";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { _isLoggedIn } from "@/lib/api/helper";

interface AuthTOTPLayoutProps {
  children: ReactNode;
}

export default function AdminLayout({ children }: AuthTOTPLayoutProps) {
    const router = useRouter();
    const [isChecking, setIsChecking] = useState(true);

    useEffect(() => {
        if (typeof window !== 'undefined') {
            if (!_isLoggedIn()) {
                router.push("/login");
            } else {
                setIsChecking(false);
            }
        }
    }, [router]);

    if (isChecking) {
        return (
            <div className="min-h-screen flex items-center justify-center">
                <p className="text-gray-600">Đang kiểm tra...</p>
            </div>
        );
    }

    return <>{children}</>;
}

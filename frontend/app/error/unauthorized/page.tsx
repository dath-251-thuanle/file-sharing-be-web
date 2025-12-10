"use client";

import { Suspense } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import Link from "next/link";
import { LockKeyhole, LogIn, Home } from "lucide-react";

function UnauthorizedContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const message = searchParams.get("message") || "Bạn cần đăng nhập để truy cập tài nguyên này.";
  const redirectUrl = searchParams.get("redirect");

  const handleLogin = () => {
    if (redirectUrl) {
      router.push(`/login?redirect=${encodeURIComponent(redirectUrl)}`);
    } else {
      router.push("/login");
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-amber-50 to-orange-50 flex items-center justify-center px-4">
      <div className="max-w-2xl w-full">
        {/* Icon và Title */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-amber-100 mb-6">
            <LockKeyhole className="w-10 h-10 text-amber-600" />
          </div>
          <h1 className="text-4xl font-bold text-gray-900 mb-2">
            Yêu cầu đăng nhập
          </h1>
          <p className="text-lg text-gray-600">
            401 Unauthorized
          </p>
        </div>

        {/* Error Box */}
        <div className="bg-white rounded-2xl shadow-xl border border-amber-100 p-8 mb-6">
          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-6 h-6 rounded-full bg-amber-100 flex items-center justify-center mt-0.5">
                <LockKeyhole className="w-4 h-4 text-amber-600" />
              </div>
              <div className="flex-1">
                <p className="text-base text-amber-700 font-medium mb-2">
                  {message}
                </p>
                <p className="text-sm text-gray-600">
                  Vui lòng đăng nhập để tiếp tục truy cập tài nguyên này.
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <button
            onClick={handleLogin}
            className="inline-flex items-center justify-center gap-2 px-6 py-3 rounded-xl bg-blue-600 text-white font-medium hover:bg-blue-700 transition-colors shadow-lg hover:shadow-xl"
          >
            <LogIn className="w-5 h-5" />
            Đăng nhập
          </button>
          
          <Link
            href="/"
            className="inline-flex items-center justify-center gap-2 px-6 py-3 rounded-xl border-2 border-gray-300 text-gray-700 font-medium hover:bg-gray-50 transition-colors"
          >
            <Home className="w-5 h-5" />
            Quay về trang chủ
          </Link>
        </div>

        {/* Additional Info */}
        <div className="mt-8 text-center">
          <p className="text-sm text-gray-500 mb-2">
            Chưa có tài khoản?
          </p>
          <Link
            href="/register"
            className="text-blue-600 hover:text-blue-700 font-medium"
          >
            Đăng ký ngay
          </Link>
        </div>
      </div>
    </div>
  );
}

export default function UnauthorizedPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-gradient-to-br from-amber-50 to-orange-50 flex items-center justify-center">
        <div className="text-center">
          <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-amber-100 mb-6">
            <LockKeyhole className="w-10 h-10 text-amber-600" />
          </div>
          <p className="text-gray-600">Đang tải...</p>
        </div>
      </div>
    }>
      <UnauthorizedContent />
    </Suspense>
  );
}


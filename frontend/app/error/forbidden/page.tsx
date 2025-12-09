"use client";

import { Suspense } from "react";
import { useSearchParams } from "next/navigation";
import Link from "next/link";
import { ShieldX, Home, Mail } from "lucide-react";

function ForbiddenContent() {
  const searchParams = useSearchParams();
  const message = searchParams.get("message") || "Bạn không có quyền truy cập tài nguyên này.";
  const reason = searchParams.get("reason") || "not_whitelisted";

  return (
    <div className="min-h-screen bg-gradient-to-br from-red-50 to-orange-50 flex items-center justify-center px-4">
      <div className="max-w-2xl w-full">
        {/* Icon và Title */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-red-100 mb-6">
            <ShieldX className="w-10 h-10 text-red-600" />
          </div>
          <h1 className="text-4xl font-bold text-gray-900 mb-2">
            Không có quyền truy cập
          </h1>
          <p className="text-lg text-gray-600">
            403 Forbidden
          </p>
        </div>

        {/* Error Box */}
        <div className="bg-white rounded-2xl shadow-xl border border-red-100 p-8 mb-6">
          <div className="space-y-4">
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-6 h-6 rounded-full bg-red-100 flex items-center justify-center mt-0.5">
                <span className="text-red-600 text-sm font-bold">!</span>
              </div>
              <div className="flex-1">
                <p className="text-base text-red-700 font-medium mb-2">
                  {message}
                </p>
                
                {reason === "not_whitelisted" && (
                  <div className="space-y-2 text-sm text-gray-600">
                    <p className="flex items-center gap-2">
                      <Mail className="w-4 h-4 text-gray-400" />
                      Email của bạn không nằm trong danh sách được phép tải file này.
                    </p>
                    <p className="mt-3 text-gray-700">
                      Vui lòng liên hệ với người chia sẻ file để được thêm vào danh sách.
                    </p>
                  </div>
                )}

                {reason === "password_required" && (
                  <div className="space-y-2 text-sm text-gray-600">
                    <p>File này được bảo vệ bằng mật khẩu.</p>
                    <p className="text-gray-700">
                      Vui lòng nhập mật khẩu để tải xuống.
                    </p>
                  </div>
                )}
              </div>
            </div>
          </div>
        </div>

        {/* Actions */}
        <div className="flex flex-col sm:flex-row gap-4 justify-center">
          <Link
            href="/"
            className="inline-flex items-center justify-center gap-2 px-6 py-3 rounded-xl bg-blue-600 text-white font-medium hover:bg-blue-700 transition-colors shadow-lg hover:shadow-xl"
          >
            <Home className="w-5 h-5" />
            Quay về trang chủ
          </Link>
          
          <Link
            href="/dashboard"
            className="inline-flex items-center justify-center gap-2 px-6 py-3 rounded-xl border-2 border-gray-300 text-gray-700 font-medium hover:bg-gray-50 transition-colors"
          >
            Xem file của tôi
          </Link>
        </div>

        {/* Additional Info */}
        <div className="mt-8 text-center">
          <p className="text-sm text-gray-500">
            Nếu bạn cho rằng đây là lỗi, vui lòng liên hệ với quản trị viên.
          </p>
        </div>
      </div>
    </div>
  );
}

export default function ForbiddenPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-gradient-to-br from-red-50 to-orange-50 flex items-center justify-center">
        <div className="text-center">
          <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-red-100 mb-6">
            <ShieldX className="w-10 h-10 text-red-600" />
          </div>
          <p className="text-gray-600">Đang tải...</p>
        </div>
      </div>
    }>
      <ForbiddenContent />
    </Suspense>
  );
}


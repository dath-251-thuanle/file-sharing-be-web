"use client";

import { Suspense } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { useEffect } from "react";
import { ArrowRight, LogIn } from "lucide-react";

function RedirectingContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const redirectTo = searchParams.get("to");
  const message = searchParams.get("message") || "Đang chuyển hướng...";

  useEffect(() => {
    if (redirectTo) {
      const timer = setTimeout(() => {
        router.push(redirectTo);
      }, 2000);
      return () => clearTimeout(timer);
    }
  }, [redirectTo, router]);

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-50 flex items-center justify-center px-4">
      <div className="max-w-md w-full">
        {/* Icon và Title */}
        <div className="text-center mb-8">
          <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-blue-100 mb-6 animate-pulse">
            <LogIn className="w-10 h-10 text-blue-600" />
          </div>
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            {message}
          </h1>
          <p className="text-gray-600">
            Vui lòng đợi trong giây lát...
          </p>
        </div>

        {/* Progress Box */}
        <div className="bg-white rounded-2xl shadow-xl border border-blue-100 p-8 mb-6">
          <div className="space-y-4">
            {/* Loading Animation */}
            <div className="flex items-center justify-center gap-3 text-blue-600">
              <div className="w-2 h-2 bg-blue-600 rounded-full animate-bounce" style={{ animationDelay: '0ms' }}></div>
              <div className="w-2 h-2 bg-blue-600 rounded-full animate-bounce" style={{ animationDelay: '150ms' }}></div>
              <div className="w-2 h-2 bg-blue-600 rounded-full animate-bounce" style={{ animationDelay: '300ms' }}></div>
            </div>
            
            {/* Progress Bar */}
            <div className="w-full bg-gray-200 rounded-full h-2 overflow-hidden">
              <div className="bg-blue-600 h-2 rounded-full animate-progress"></div>
            </div>
            
            <p className="text-center text-sm text-gray-600">
              Đang xử lý yêu cầu của bạn...
            </p>
          </div>
        </div>

        {/* Manual Link */}
        {redirectTo && (
          <div className="text-center">
            <p className="text-sm text-gray-500 mb-3">
              Nếu không tự động chuyển hướng:
            </p>
            <button
              onClick={() => router.push(redirectTo)}
              className="inline-flex items-center gap-2 text-blue-600 hover:text-blue-700 font-medium transition-colors"
            >
              Nhấn vào đây
              <ArrowRight className="w-4 h-4" />
            </button>
          </div>
        )}
      </div>

      <style jsx>{`
        @keyframes progress {
          0% {
            width: 0%;
          }
          100% {
            width: 100%;
          }
        }
        .animate-progress {
          animation: progress 2s ease-in-out;
        }
      `}</style>
    </div>
  );
}

export default function RedirectingPage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-50 flex items-center justify-center">
        <div className="text-center">
          <div className="inline-flex items-center justify-center w-20 h-20 rounded-full bg-blue-100 mb-6">
            <LogIn className="w-10 h-10 text-blue-600" />
          </div>
          <p className="text-gray-600">Đang tải...</p>
        </div>
      </div>
    }>
      <RedirectingContent />
    </Suspense>
  );
}


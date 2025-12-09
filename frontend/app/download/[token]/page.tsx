"use client";

import React, { useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { toast } from "sonner";
import { getAccessToken } from "@/lib/api/helper";

export default function PasswordPage() {
  const params = useParams();
  const router = useRouter();
  const { token } = params as { token: string };
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleDownload = async () => {
    if (!password.trim()) {
      setError("Vui lòng nhập mật khẩu");
      toast.error("Vui lòng nhập mật khẩu");
      return;
    }

    setLoading(true);
    setError(null);

    try {
      const tokenFromStorage = getAccessToken();
      const headers: HeadersInit = {
        'X-File-Password': password,
      };
      
      if (tokenFromStorage) {
        headers['Authorization'] = `Bearer ${tokenFromStorage}`;
      }

      const response = await fetch(`/api/files/${token}/download`, {
        method: 'GET',
        headers,
        credentials: 'include',
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        const errorMessage = errorData.message || errorData.error || "Mật khẩu không đúng";
        
        // Check if requires login
        if (response.status === 401 && errorData.requiresLogin) {
          toast.error(errorMessage);
          setTimeout(() => {
            router.push("/login");
          }, 2000);
          return;
        }
        
        setError(errorMessage);
        toast.error(errorMessage);
        return;
      }

      // Download successful
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      
      // Get filename from Content-Disposition header
      const contentDisposition = response.headers.get('Content-Disposition');
      let filename = `file-${token}`;
      if (contentDisposition) {
        const filenameMatch = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/);
        if (filenameMatch && filenameMatch[1]) {
          filename = filenameMatch[1].replace(/['"]/g, '');
        }
      }
      
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      
      toast.success("Đã bắt đầu tải file!");
    } catch (e: any) {
      const errorMessage = e?.message || "Không thể tải xuống";
      setError(errorMessage);
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center py-12 px-4">
      <div className="w-full max-w-md">
        <div className="bg-white rounded-2xl shadow-xl p-8">
          <div className="text-center mb-6">
            <h1 className="text-3xl font-bold text-gray-800 mb-2">
              File được bảo vệ bằng mật khẩu
            </h1>
            <p className="text-gray-600">
              Vui lòng nhập mật khẩu để tải xuống file
            </p>
          </div>

          {error && (
            <div className="mb-4 p-4 bg-red-100 border border-red-300 text-red-800 rounded-lg">
              {error}
            </div>
          )}

          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Mật khẩu
            </label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  handleDownload();
                }
              }}
              className="w-full px-4 py-3 border border-gray-300 rounded-lg focus:ring-2 focus:ring-indigo-500 focus:border-indigo-500"
              placeholder="Nhập mật khẩu..."
              autoFocus
            />
          </div>

          <button
            onClick={handleDownload}
            disabled={loading}
            className={`w-full py-3 px-6 rounded-lg text-white font-semibold transition-all ${
              loading
                ? "bg-gray-400 cursor-not-allowed"
                : "bg-indigo-600 hover:bg-indigo-700"
            }`}
          >
            {loading ? "Đang tải..." : "⬇️ Tải xuống"}
          </button>

          <button
            onClick={() => router.back()}
            className="w-full mt-4 py-2 px-4 text-gray-600 hover:text-gray-800 transition"
          >
            ← Quay lại
          </button>
        </div>
      </div>
    </div>
  );
}


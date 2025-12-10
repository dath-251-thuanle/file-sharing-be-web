"use client";

import { useEffect, useState } from "react";
import { setAdminToken, getAdminToken, clearAdminToken } from "@/lib/api/helper";
import { toast } from "sonner";

type AdminTokenBarProps = {
  className?: string;
};

export function AdminTokenBar({ className = "" }: AdminTokenBarProps) {
  const [token, setToken] = useState("");

  useEffect(() => {
    const existing = getAdminToken();
    if (existing) {
      setToken(existing);
    }
  }, []);

  const handleSave = () => {
    if (!token.trim()) {
      toast.error("Vui lòng nhập ADMIN_API_TOKEN");
      return;
    }
    setAdminToken(token.trim());
    toast.success("Đã lưu ADMIN_API_TOKEN");
  };

  const handleClear = () => {
    clearAdminToken();
    setToken("");
    toast.success("Đã xoá ADMIN_API_TOKEN");
  };

  return (
    <div
      className={`mb-4 rounded-lg border border-gray-200 bg-gray-50 p-4 shadow-sm ${className}`}
    >
      <div className="flex flex-col gap-3 md:flex-row md:items-center">
        <div className="flex-1">
          <label className="block text-sm font-medium text-gray-700 mb-1">
            ADMIN_API_TOKEN
          </label>
          <input
            type="password"
            value={token}
            onChange={(e) => setToken(e.target.value)}
            placeholder="Nhập ADMIN_API_TOKEN"
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:ring-1 focus:ring-blue-500"
          />
          <p className="mt-1 text-xs text-gray-500">
            Token này sẽ được gửi kèm Authorization header cho các API /admin/*
          </p>
        </div>
        <div className="flex gap-2">
          <button
            onClick={handleSave}
            className="rounded-md bg-blue-600 px-3 py-2 text-sm font-medium text-white shadow-sm hover:bg-blue-700"
          >
            Lưu token
          </button>
          <button
            onClick={handleClear}
            className="rounded-md border px-3 py-2 text-sm font-medium hover:bg-gray-100"
          >
            Xóa
          </button>
        </div>
      </div>
    </div>
  );
}


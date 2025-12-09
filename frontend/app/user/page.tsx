"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import { User } from "@/lib/components/schemas";
import { getUserProfile } from "@/lib/api/auth";
import { Loader, ShieldCheck, ShieldOff } from "lucide-react";
import { setCurrentUser } from "@/lib/api/helper";
import { toast } from "sonner";
import Link from "next/link";

export default function UserProfilePage() {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const router = useRouter();

  useEffect(() => {
    const fetchProfile = async () => {
      setIsLoading(true);
      try {
        const response = await getUserProfile();
        setCurrentUser(response.user);
        setUser(response.user);
      } catch (err: any) {
        if (err.message?.includes("Unauthorized") || err.status === 401) {
          router.push("/login");
        } else {
          setError("Không thể tải thông tin profile.");
          toast.error("Không thể tải thông tin profile.");
        }
      } finally {
        setIsLoading(false);
      }
    };

    fetchProfile();
  }, [router]);

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Loader className="animate-spin h-8 w-8 text-gray-500" />
        <p className="ml-2 text-gray-500">Đang tải thông tin profile...</p>
      </div>
    );
  }

  if (error || !user) {
    return (
      <div className="container mx-auto p-4 sm:p-6">
        <div className="bg-white shadow-md rounded-lg p-6 text-center">
          <p className="text-red-500">{error || "Không thể tải thông tin profile"}</p>
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto p-4 sm:p-6">
      <div className="bg-white shadow-md rounded-lg p-6 mb-8">
        <h1 className="text-2xl font-bold mb-6">Thông tin tài khoản</h1>
        
        <div className="space-y-4 mb-6">
          <div>
            <label className="text-sm font-medium text-gray-500">Username</label>
            <p className="text-lg text-gray-900">{user.username}</p>
          </div>
          <div>
            <label className="text-sm font-medium text-gray-500">Email</label>
            <p className="text-lg text-gray-900">{user.email}</p>
          </div>
          <div>
            <label className="text-sm font-medium text-gray-500">Vai trò</label>
            <p className="text-lg text-gray-900 capitalize">{user.role}</p>
          </div>
        </div>

        <div className="mt-6 pt-6 border-t border-gray-200">
          <h2 className="text-lg font-semibold mb-4">Two-Factor Authentication (2FA)</h2>
          {user.totpEnabled ? (
            <div className="flex items-center gap-4">
              <span className="flex items-center text-green-600">
                <ShieldCheck className="h-5 w-5 mr-1" />
                2FA đã được bật
              </span>
              <Link 
                href="/totp-setup" 
                className="px-4 py-2 text-sm font-medium text-white bg-yellow-600 rounded-md hover:bg-yellow-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-yellow-500"
              >
                Quản lý 2FA
              </Link>
            </div>
          ) : (
            <div className="flex items-center gap-4">
              <span className="flex items-center text-red-600">
                <ShieldOff className="h-5 w-5 mr-1" />
                2FA chưa được bật
              </span>
              <Link 
                href="/totp-setup" 
                className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                Bật 2FA
              </Link>
            </div>
          )}
        </div>

        <div className="mt-6 pt-6 border-t border-gray-200">
          <h2 className="text-lg font-semibold mb-4">Thao tác</h2>
          <div className="flex flex-wrap gap-4">
            <Link
              href="/dashboard"
              className="px-4 py-2 text-sm font-medium text-white bg-gray-600 rounded-md hover:bg-gray-700"
            >
              Dashboard
            </Link>
            <Link
              href="/files/my"
              className="px-4 py-2 text-sm font-medium text-white bg-blue-600 rounded-md hover:bg-blue-700"
            >
              File của tôi
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}


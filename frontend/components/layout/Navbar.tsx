"use client";

import { useState, useEffect, useCallback } from "react";
import Link from "next/link";
import { usePathname, useRouter } from "next/navigation";
import { HardDrive } from "lucide-react";
import { logout, getUserProfile } from "@/lib/api/auth";
import { setCurrentUser, getAccessToken } from "@/lib/api/helper";

export default function Navbar() {
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const [isAdmin, setIsAdmin] = useState(false);
  const [mounted, setMounted] = useState(false);
  const router = useRouter();
  const pathname = usePathname();

  const fetchUser = useCallback(async () => {
    // Only fetch user profile if token exists
    const token = getAccessToken();
    if (!token) {
      setIsLoggedIn(false);
      setIsAdmin(false);
      setMounted(true);
      return;
    }

    try {
      const res = await getUserProfile();
      setCurrentUser(res.user);
      setIsLoggedIn(true);
      setIsAdmin(res.user.role === "admin");
    } catch {
      setIsLoggedIn(false);
      setIsAdmin(false);
    } finally {
      setMounted(true);
    }
  }, []);

  useEffect(() => {
    fetchUser();
  }, [fetchUser, pathname]);

  useEffect(() => {
    const onFocus = () => fetchUser();
    window.addEventListener("focus", onFocus);
    return () => {
      window.removeEventListener("focus", onFocus);
    };
  }, [fetchUser]);

  const handleLogout = () => {
    logout();
    setIsLoggedIn(false);
    setIsAdmin(false);
    router.push("/login");
  };

  return (
    <nav className="fixed top-0 left-0 right-0 h-16 bg-white border-b border-gray-200 z-50 px-4 sm:px-6 lg:px-8 flex items-center justify-between">
      <Link href="/" className="flex items-center gap-2">
        <div className="bg-blue-600 p-1.5 rounded text-white" suppressHydrationWarning>
          <HardDrive size={20} />
        </div>
        <span className="font-bold text-xl text-gray-900">SecureShare</span>
      </Link>

      <div className="flex items-center gap-4" suppressHydrationWarning>
        <Link href="/upload" className="text-sm font-medium text-gray-600 hover:text-blue-600 transition-colors">
          Upload Mới
        </Link>
        
        {mounted && isLoggedIn ? (
          <>
            <Link href="/files/my" className="text-sm font-medium text-gray-600 hover:text-blue-600 transition-colors">
              File của tôi
            </Link>
            {isAdmin && (
              <Link
                href="/admin"
                className="text-sm font-medium text-red-600 hover:text-red-700"
              >
                Admin Dashboard
              </Link>
            )}
            <Link href="/dashboard" className="text-sm font-medium text-gray-900 hover:text-gray-700">
              Dashboard
            </Link>
            <Link href="/user" className="text-sm font-medium text-gray-600 hover:text-gray-900">
              Profile
            </Link>
            <button 
              onClick={handleLogout}
              className="text-sm font-medium text-gray-600 hover:text-gray-900"
            >
              Logout
            </button>
          </>
        ) : (
          <>
            <Link href="/login" className="text-sm font-medium text-gray-600 hover:text-gray-900">
              Đăng nhập
            </Link>
            <Link 
              href="/register" 
              className="text-sm font-medium bg-gray-900 text-white px-4 py-2 rounded-md hover:bg-gray-800 transition-colors"
            >
              Đăng ký
            </Link>
          </>
        )}
      </div>
    </nav>
  );
}
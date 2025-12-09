"use client";

import type { ReactNode } from "react";

interface AdminLayoutProps {
  children: ReactNode;
}

/**
 * Admin pages are protected bằng ADMIN_API_TOKEN (Bearer) chứ không dựa vào role user.
 * Vì vậy không chặn render theo role; để người dùng nhập token trong AdminTokenBar.
 */
export default function AdminLayout({ children }: AdminLayoutProps) {
  return <>{children}</>;
}

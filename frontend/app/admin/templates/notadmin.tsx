import Link from "next/link";
import { ShieldOff, ArrowLeft } from "lucide-react";

interface NotAdminPageProps {
  message?: string;
}

export default function NotAdminPage({ message = "Bạn không có quyền truy cập trang này." }: NotAdminPageProps) {
  return (
    <div className="flex h-screen">
      <div className="m-auto">
        <div className="bg-white border border-gray-200 rounded-2xl shadow-sm p-8 text-center">
          <div className="m-auto mb-4 flex h-16 w-16 items-center justify-center rounded-full bg-red-50">
            <ShieldOff className="h-8 w-8 text-red-500" />
          </div>

          <h1 className="text-2xl font-bold text-gray-900 mb-2">
            Truy cập bị từ chối
          </h1>
          <p className="text-gray-600 mb-6">Bạn không có quyền truy cập trang này.</p>

          <div className="flex flex-col sm:flex-row gap-3 justify-center">
            <Link
              href="/"
              className="inline-flex items-center justify-center px-5 py-2.5 rounded-md border border-gray-300 bg-white text-sm font-medium text-gray-700 hover:bg-gray-50 transition"
            >
              <ArrowLeft className="w-4 h-4 mr-2" />
              Về trang chủ
            </Link>

            <Link
              href="/login"
              className="inline-flex items-center justify-center px-5 py-2.5 rounded-md text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 shadow-sm hover:shadow-md transition"
            >
              Đăng nhập với quyền Admin
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
}
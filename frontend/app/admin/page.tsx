import Link from "next/link";
import { ShieldCheck, Clock, UploadCloud, Zap, ArrowLeft } from "lucide-react";
import { AdminTokenBar } from "@/components/admin/AdminTokenBar";

export default function AdminHome() {
  return (
    <div className="w-full flex flex-col items-center">
      <div className="w-full max-w-7xl px-4 sm:px-6 lg:px-8 mt-6">
        <AdminTokenBar />
      </div>

      {/* Header */}
      <section className="w-full border-b border-gray-200 bg-white">
        <div className="max-w-7xl mx-auto px-4 py-10 sm:px-6 lg:px-8">
          <div className="flex items-center gap-3 mb-4">
            <span className="inline-flex h-10 w-10 items-center justify-center rounded-lg bg-blue-100">
              <ShieldCheck className="w-6 h-6 text-blue-600" />
            </span>
            <div>
              <h1 className="text-2xl sm:text-3xl font-bold text-gray-900">
                Bảng điều khiển Admin
              </h1>
              <p className="text-sm text-gray-500">
                Quản lý cấu hình hệ thống, dọn dẹp file và giám sát hoạt động.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Main content */}
      <section className="w-full max-w-7xl mx-auto px-4 py-10 sm:px-6 lg:px-8">
        <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
          {/* System Policy */}
          <div className="bg-white border border-gray-100 rounded-xl shadow-sm p-6 flex flex-col justify-between">
            <div>
              <div className="w-12 h-12 rounded-lg bg-blue-100 flex items-center justify-center mb-4">
                <Clock className="w-6 h-6 text-blue-600" />
              </div>
              <h2 className="text-lg font-semibold mb-2">System Policy</h2>
              <p className="text-sm text-gray-600 mb-4">
                Cấu hình giới hạn hệ thống: dung lượng tối đa, thời gian hiệu
                lực, độ dài mật khẩu, v.v.
              </p>
            </div>
            <Link
              href="/admin/policy"
              className="inline-flex items-center justify-center px-4 py-2 rounded-md bg-blue-600 text-white text-sm font-medium hover:bg-blue-700 transition"
            >
              <ShieldCheck className="w-4 h-4 mr-2" />
              Quản lý System Policy
            </Link>
          </div>

          {/* Cleanup */}
          <div className="bg-white border border-gray-100 rounded-xl shadow-sm p-6 flex flex-col justify-between">
            <div>
              <div className="w-12 h-12 rounded-lg bg-amber-100 flex items-center justify-center mb-4">
                <Zap className="w-6 h-6 text-amber-500" />
              </div>
              <h2 className="text-lg font-semibold mb-2">Cleanup</h2>
              <p className="text-sm text-gray-600 mb-4">
                Thực hiện dọn dẹp file hết hạn thủ công để demo hoặc kiểm tra
                hệ thống.
              </p>
            </div>
            <Link
              href="/admin/cleanup"
              className="inline-flex items-center justify-center px-4 py-2 rounded-md bg-amber-500 text-white text-sm font-medium hover:bg-amber-600 transition"
            >
              <Zap className="w-4 h-4 mr-2" />
              Chạy Cleanup
            </Link>
          </div>

          {/* My Files / Upload shortcut */}
          <div className="bg-white border border-gray-100 rounded-xl shadow-sm p-6 flex flex-col justify-between">
            <div>
              <div className="w-12 h-12 rounded-lg bg-green-100 flex items-center justify-center mb-4">
                <UploadCloud className="w-6 h-6 text-green-600" />
              </div>
              <h2 className="text-lg font-semibold mb-2">Files & Upload</h2>
              <p className="text-sm text-gray-600 mb-4">
                Xem danh sách file của bạn hoặc chuyển sang trang upload để
                kiểm thử nhanh.
              </p>
            </div>
            <div className="flex flex-col sm:flex-row gap-2">
              <Link
                href="/dashboard"
                className="flex-1 inline-flex items-center justify-center px-4 py-2 rounded-md border border-gray-200 text-sm font-medium text-gray-700 hover:bg-gray-50 transition"
              >
                Danh sách file
              </Link>
              <Link
                href="/upload"
                className="flex-1 inline-flex items-center justify-center px-4 py-2 rounded-md bg-green-600 text-white text-sm font-medium hover:bg-green-700 transition"
              >
                <UploadCloud className="w-4 h-4 mr-2" />
                Upload mới
              </Link>
            </div>
          </div>
        </div>

        {/* Back to home */}
        <div className="mt-10">
          <Link
            href="/"
            className="inline-flex items-center text-sm text-gray-500 hover:text-gray-700"
          >
            <ArrowLeft className="w-4 h-4 mr-1" />
            Về trang landing
          </Link>
        </div>
      </section>
    </div>
  );
}
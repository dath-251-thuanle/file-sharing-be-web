"use client";

import { Component } from "react";
import Link from "next/link";
import { ShieldCheck, Clock, Lock, UploadCloud } from "lucide-react";

export default class Home extends Component {
  render() {
    return (
      <div className="flex flex-col items-center">
        
        {/* HERO SECTION */}
        <section className="w-full bg-white border-b border-gray-200">
          <div className="max-w-7xl mx-auto px-4 py-20 sm:px-6 lg:px-8 text-center">
            <h1 className="text-4xl sm:text-6xl font-extrabold text-gray-900 tracking-tight mb-6">
              Chia sẻ file <span className="text-blue-600">An toàn</span> & <span className="text-blue-600">Nhanh chóng</span>
            </h1>
            <p className="text-lg sm:text-xl text-gray-500 max-w-2xl mx-auto mb-10">
              Upload file không cần tài khoản. Bảo vệ bằng mật khẩu, xác thực 2 lớp (TOTP) và tự động xóa khi hết hạn.
            </p>
            
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              {/* CTA: Upload Ngay (Dành cho Anonymous - Khánh phụ trách) */}
              <Link 
                href="/upload" 
                className="inline-flex items-center justify-center px-8 py-3 border border-transparent text-base font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 md:text-lg transition-all shadow-lg hover:shadow-xl"
              >
                <UploadCloud className="w-5 h-5 mr-2" />
                Upload Ngay
              </Link>
              
              {/* CTA: Quản lý file (Dành cho Dashboard - Minh Quan phụ trách) */}
              <Link 
                href="/login" 
                className="inline-flex items-center justify-center px-8 py-3 border border-gray-300 text-base font-medium rounded-md text-gray-700 bg-white hover:bg-gray-50 md:text-lg transition-all"
              >
                Đăng nhập quản lý
              </Link>
            </div>
          </div>
        </section>

        {/* FEATURE SECTION */}
        <section className="max-w-7xl mx-auto px-4 py-16 sm:px-6 lg:px-8">
          <div className="text-center mb-12">
            <h2 className="text-3xl font-bold text-gray-900">Tính năng nổi bật</h2>
            <p className="mt-4 text-gray-500">Hệ thống được thiết kế để bảo vệ dữ liệu của bạn tối đa.</p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
            {/* Feature 1: Time Validity */}
            <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100 hover:shadow-md transition-shadow">
              <div className="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-4">
                <Clock className="w-6 h-6 text-blue-600" />
              </div>
              <h3 className="text-xl font-semibold mb-2">Hẹn giờ tự hủy</h3>
              <p className="text-gray-600">
                Thiết lập thời gian hiệu lực (From - To). File sẽ tự động bị xóa vĩnh viễn khỏi hệ thống khi hết hạn.
              </p>
            </div>

            {/* Feature 2: Security */}
            <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100 hover:shadow-md transition-shadow">
              <div className="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mb-4">
                <ShieldCheck className="w-6 h-6 text-green-600" />
              </div>
              <h3 className="text-xl font-semibold mb-2">Bảo mật đa lớp</h3>
              <p className="text-gray-600">
                Hỗ trợ đặt mật khẩu riêng cho file và xác thực 2 lớp (TOTP) để đảm bảo chỉ người nhận mới mở được.
              </p>
            </div>

            {/* Feature 3: Access Control */}
            <div className="bg-white p-6 rounded-xl shadow-sm border border-gray-100 hover:shadow-md transition-shadow">
              <div className="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mb-4">
                <Lock className="w-6 h-6 text-purple-600" />
              </div>
              <h3 className="text-xl font-semibold mb-2">Quyền truy cập</h3>
              <p className="text-gray-600">
                Chế độ Public, Password-protected hoặc chia sẻ riêng tư cho danh sách email cụ thể.
              </p>
            </div>
          </div>
        </section>

        {/* HOW IT WORKS SECTION */}
        <section className="w-full bg-gray-900 text-white py-16">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <h2 className="text-3xl font-bold text-center mb-12">Cách thức hoạt động</h2>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-8 text-center">
              <div className="flex flex-col items-center">
                <div className="w-16 h-16 bg-gray-800 rounded-full flex items-center justify-center mb-4 text-2xl font-bold text-blue-400">1</div>
                <h4 className="text-xl font-semibold mb-2">Tải file lên</h4>
                <p className="text-gray-400">Kéo thả file, đặt mật khẩu và thời gian hết hạn.</p>
              </div>
              <div className="flex flex-col items-center">
                <div className="w-16 h-16 bg-gray-800 rounded-full flex items-center justify-center mb-4 text-2xl font-bold text-blue-400">2</div>
                <h4 className="text-xl font-semibold mb-2">Nhận Link chia sẻ</h4>
                <p className="text-gray-400">Hệ thống tạo ra link duy nhất (Share Token) cho file của bạn.</p>
              </div>
              <div className="flex flex-col items-center">
                <div className="w-16 h-16 bg-gray-800 rounded-full flex items-center justify-center mb-4 text-2xl font-bold text-blue-400">3</div>
                <h4 className="text-xl font-semibold mb-2">Người nhận tải về</h4>
                <p className="text-gray-400">Truy cập link, nhập mật khẩu (nếu có) và tải file an toàn.</p>
              </div>
            </div>
          </div>
        </section>
      </div>
    );
  }
}
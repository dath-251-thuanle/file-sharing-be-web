"use client";

import { Component, FormEvent } from "react";
import { Mail, Lock } from "lucide-react";
import Link from "next/link";

export interface LoginFormData {
  email: string;
  password: string;
}

interface LoginFormProps {
  formData: LoginFormData;
  updateField: (field: keyof LoginFormData, value: any) => void;
  handleSubmit: (e: FormEvent<HTMLFormElement>) => void;
}

export default class LoginForm extends Component<LoginFormProps> {
  render() {
    const { formData, updateField, handleSubmit } = this.props;

    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <div className="w-full max-w-md p-8 space-y-8 bg-white rounded-xl shadow-lg">
          <div className="text-center">
            <h1 className="text-3xl font-bold text-gray-900">Đăng nhập</h1>
            <p className="mt-2 text-gray-500">
              Chào mừng trở lại! Đăng nhập để quản lý file của bạn.
            </p>
          </div>
          <form className="space-y-6" onSubmit={handleSubmit}>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                <Mail className="w-5 h-5 text-gray-400" />
              </div>
              <input
                type="email"
                id="email"
                name="email"
                value={formData.email}
                onChange={(e) => updateField("email", e.target.value)}
                className="w-full py-3 pl-10 pr-4 text-gray-900 bg-gray-100 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:bg-white"
                placeholder="Email"
                required
              />
            </div>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                <Lock className="w-5 h-5 text-gray-400" />
              </div>
              <input
                type="password"
                id="password"
                name="password"
                value={formData.password}
                onChange={(e) => updateField("password", e.target.value)}
                className="w-full py-3 pl-10 pr-4 text-gray-900 bg-gray-100 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:bg-white"
                placeholder="Mật khẩu"
                required
              />
            </div>
            <button
              type="submit"
              className="w-full px-4 py-3 font-semibold text-white bg-blue-600 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
            >
              Đăng nhập
            </button>
          </form>
          <div className="text-center text-gray-500">
            Chưa có tài khoản?{" "}
            <Link href="/register" className="font-medium text-blue-600 hover:underline">
              Đăng ký ngay
            </Link>
          </div>
        </div>
      </div>
    );
  }
}

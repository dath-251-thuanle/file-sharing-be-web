"use client";

import { Component, FormEvent } from "react";
import { Mail, KeyRound } from "lucide-react";

export interface LoginTotpFormData {
  code: string;
}

interface LoginTotpFormProps {
  email: string;
  formData: LoginTotpFormData;
  updateField: (field: keyof LoginTotpFormData, value: string) => void;
  handleSubmit: (e: FormEvent<HTMLFormElement>) => void;
  verifying: boolean;
}

export default class LoginTotpForm extends Component<LoginTotpFormProps> {
  render() {
    const { email, formData, updateField, handleSubmit, verifying } =
      this.props;

    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <div className="w-full max-w-md p-8 space-y-8 bg-white rounded-xl shadow-lg">
          <div className="text-center">
            <h1 className="text-3xl font-bold text-gray-900">
              Two-Factor Authentication
            </h1>
            <p className="mt-2 text-gray-500">
              Enter the code from your authenticator app.
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
                value={email}
                className="w-full py-3 pl-10 pr-4 text-gray-900 bg-gray-200 border border-gray-300 rounded-md focus:outline-none"
                placeholder="Email"
                disabled
              />
            </div>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                <KeyRound className="w-5 h-5 text-gray-400" />
              </div>
              <input
                type="text"
                id="code"
                name="code"
                value={formData.code}
                onChange={(e) => updateField("code", e.target.value)}
                className="w-full py-3 pl-10 pr-4 text-gray-900 bg-gray-100 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:bg-white"
                placeholder="6-digit code"
                required
                maxLength={6}
                pattern="\d{6}"
              />
            </div>
            <button
              type="submit"
              className="w-full px-4 py-3 font-semibold text-white bg-blue-600 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:bg-gray-400"
              disabled={verifying || formData.code.length !== 6}
            >
              Verify & Login
            </button>
          </form>
        </div>
      </div>
    );
  }
}

"use client";

import { Component, FormEvent } from "react";
import Link from "next/link";
import { KeyRound, Loader2, ShieldAlert } from "lucide-react";

export interface TotpSetupFormData {
  totpCode: string;
}

interface TotpSetupFormProps {
  formData: TotpSetupFormData;
  updateField: (field: keyof TotpSetupFormData, value: string) => void;
  handleSubmit: (e: FormEvent<HTMLFormElement>) => void;
  qrCode: string | null;
  secret: string | null;
  loading: boolean;
  error: string | null;
  verifying: boolean;
}

export default class TotpSetupForm extends Component<TotpSetupFormProps> {
  render() {
    const {
      formData,
      updateField,
      handleSubmit,
      qrCode,
      secret,
      loading,
      error,
      verifying,
    } = this.props;

    if (loading) {
      return (
        <div className="flex flex-col items-center justify-center min-h-screen bg-gray-50 text-gray-700">
          <Loader2 className="w-12 h-12 animate-spin mb-4" />
          <p className="text-lg">Loading Security Setup...</p>
        </div>
      );
    }

    if (error && !qrCode) {
      return (
        <div className="flex flex-col items-center justify-center min-h-screen bg-gray-50 text-red-600">
          <ShieldAlert className="w-12 h-12 mb-4" />
          <p className="text-lg font-semibold">Error</p>
          <p className="mb-4">{error}</p>
          <Link href="/" className="text-blue-600 hover:underline">
            Go back to Home
          </Link>
        </div>
      );
    }

    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50 py-12">
        <div className="w-full max-w-lg p-8 space-y-8 bg-white rounded-xl shadow-lg">
          <div className="text-center">
            <h1 className="text-3xl font-bold text-gray-900">
              Set Up Two-Factor Authentication
            </h1>
            <p className="mt-2 text-gray-500">
              Scan the QR code with your authenticator app.
            </p>
          </div>

          <div className="flex flex-col items-center gap-8">
            <div className="p-4 bg-white border rounded-lg">
              {qrCode ? (
                <img src={qrCode} alt="TOTP QR Code" width={200} height={200} />
              ) : (
                <div className="w-[200px] h-[200px] bg-gray-200 animate-pulse"></div>
              )}
            </div>
            <div className="text-center">
              <p className="text-gray-600">Or enter this key manually:</p>
              <div className="mt-2 font-mono tracking-widest text-lg bg-gray-100 text-gray-800 px-4 py-2 rounded-md inline-block">
                {secret || "LOADING..."}
              </div>
            </div>
          </div>

          <form className="space-y-6" onSubmit={handleSubmit}>
            <div className="relative">
              <div className="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                <KeyRound className="w-5 h-5 text-gray-400" />
              </div>
              <input
                type="text"
                id="totpCode"
                name="totpCode"
                value={formData.totpCode}
                onChange={(e) => updateField("totpCode", e.target.value)}
                className="w-full py-3 pl-10 pr-4 text-gray-900 bg-gray-100 border border-gray-200 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:bg-white"
                placeholder="6-digit code"
                required
                maxLength={6}
                pattern="\d{6}"
              />
            </div>

            {error && (
              <p className="text-sm text-red-500 text-center">{error}</p>
            )}

            <button
              type="submit"
              className="w-full px-4 py-3 font-semibold text-white bg-blue-600 rounded-md hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:bg-gray-400 flex items-center justify-center"
              disabled={verifying || formData.totpCode.length !== 6}
            >
              {verifying && <Loader2 className="w-5 h-5 mr-2 animate-spin" />}
              Verify & Enable
            </button>
          </form>
        </div>
      </div>
    );
  }
}

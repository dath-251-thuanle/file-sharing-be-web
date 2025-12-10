import type { Metadata } from "next";
import { Inter } from "next/font/google";
import "./globals.css";
import Navbar from "@/components/layout/Navbar";
import Footer from "@/components/layout/Footer";
import { Toaster } from "sonner";

const inter = Inter({ subsets: ["latin"] });

export const metadata: Metadata = {
  title: "SecureShare",
  description: "Nền tảng chia sẻ tài liệu bảo mật",
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="vi">
      <body
        className={`${inter.className} min-h-screen flex flex-col antialiased bg-gray-50 text-gray-900`}
        suppressHydrationWarning
      >
        <Navbar />
        
        <main className="flex-grow pt-16">
          {children}
        </main>

        <Footer />
        
        {/* Toast notifications */}
        <Toaster position="top-center" richColors />
      </body>
    </html>
  );
}
"use client";

import { useEffect, useState } from "react";
import { useParams, useRouter } from "next/navigation";
import { File as FileType, FileInfoResponse } from "@/lib/components/schemas";
import { getFileInfoById, getFileInfoByToken } from "@/lib/api/file";
import { getCurrentUser, authenticatedFetch } from "@/lib/api/helper";
import { Loader, Download, Copy, Clock, Users, Lock, Calendar } from "lucide-react";
import { toast } from "sonner";
import Link from "next/link";

export default function FileInfoPage() {
  const params = useParams();
  const router = useRouter();
  const { id } = params as { id: string };
  
  const [file, setFile] = useState<FileType | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isOwner, setIsOwner] = useState(false);

  useEffect(() => {
    const fetchFileInfo = async () => {
      setIsLoading(true);
      try {
        const currentUser = getCurrentUser();
        
        // This page is for owner to view their file details
        // Only use /files/info/{id} endpoint (requires authentication)
        // Do NOT fallback to shareToken endpoint - that's for public downloaders
        const fileInfo = await getFileInfoById(id);
        
        setFile(fileInfo.file);
        
        // Check if current user is the owner
        if (currentUser && fileInfo.file.owner) {
          const isFileOwner = currentUser.id === fileInfo.file.owner.id;
          setIsOwner(isFileOwner);
        } else {
          setIsOwner(false);
        }
      } catch (err: any) {
        if (err.status === 401) {
          setError("Vui lòng đăng nhập để xem thông tin file.");
          router.push("/login");
        } else if (err.status === 403) {
          setError("Bạn không có quyền truy cập file này. Chỉ owner hoặc admin mới có thể xem thông tin chi tiết.");
        } else if (err.status === 404) {
          setError("File không tồn tại.");
        } else {
          setError("Không thể tải thông tin file.");
          toast.error("Không thể tải thông tin file.");
        }
      } finally {
        setIsLoading(false);
      }
    };

    fetchFileInfo();
  }, [id, router]);

  const getDownloadLink = () => {
    if (!file) return "";
    if (file.shareLink && file.shareLink.startsWith("http")) {
      return file.shareLink;
    }
    const origin = typeof window !== "undefined" ? window.location.origin : "";
    return `${origin}/api/files/${file.shareToken}/download`;
  };

  const handleCopyLink = async () => {
    const link = getDownloadLink();
    if (!link) return;
    try {
      await navigator.clipboard.writeText(link);
      toast.success("Đã sao chép link chia sẻ!");
    } catch (err) {
      toast.error("Không thể sao chép link.");
    }
  };

  const handleDownload = async () => {
    if (!file) return;
    
    try {
      const downloadLink = getDownloadLink();
      if (!downloadLink) {
        toast.error("Không thể tải file.");
        return;
      }

      // Use authenticatedFetch to ensure token is included
      const response = await authenticatedFetch(downloadLink, {
        method: 'GET',
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        const errorMessage = errorData.message || errorData.error || "Không thể tải file";
        
        if (response.status === 401) {
          toast.error("Vui lòng đăng nhập để tải file.");
          router.push("/login");
        } else if (response.status === 403) {
          toast.error(errorMessage);
        } else {
          toast.error(errorMessage);
        }
        return;
      }

      // Create blob and trigger download
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      
      // Get filename from Content-Disposition header
      const contentDisposition = response.headers.get('Content-Disposition');
      let filename = file.fileName || `file-${file.shareToken}`;
      if (contentDisposition) {
        const filenameMatch = contentDisposition.match(/filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/);
        if (filenameMatch && filenameMatch[1]) {
          filename = filenameMatch[1].replace(/['"]/g, '');
        }
      }
      
      a.download = filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      
      toast.success("Đã bắt đầu tải file!");
    } catch (error: any) {
      console.error("Download error:", error);
      toast.error("Không thể tải file. Vui lòng thử lại.");
    }
  };

  const humanFileSize = (bytes: number) => {
    const units = ["B", "KB", "MB", "GB"];
    let i = 0;
    while (bytes >= 1024 && i < units.length - 1) {
      bytes /= 1024;
      i++;
    }
    return `${bytes.toFixed(1)} ${units[i]}`;
  };

  // Filter out owner email from sharedWith list for display
  const getSharedWithFiltered = () => {
    if (!file || !isOwner || !file.sharedWith || file.sharedWith.length === 0) {
      return [];
    }
    const ownerEmail = file.owner?.email?.toLowerCase();
    return file.sharedWith.filter(
      (email) => email.toLowerCase() !== ownerEmail
    );
  };

  const formatDate = (dateString?: string) => {
    if (!dateString) return "Không có";
    return new Date(dateString).toLocaleString('vi-VN');
  };

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Loader className="animate-spin h-8 w-8 text-gray-500" />
        <p className="ml-2 text-gray-500">Đang tải thông tin file...</p>
      </div>
    );
  }

  if (error || !file) {
    return (
      <div className="container mx-auto p-4 sm:p-6">
        <div className="bg-white shadow-md rounded-lg p-6 text-center">
          <p className="text-red-500">{error || "File không tồn tại"}</p>
          <Link href="/files/my" className="mt-4 inline-block text-blue-600 hover:text-blue-800">
            Quay lại danh sách file
          </Link>
        </div>
      </div>
    );
  }

  const now = new Date();
  const expiresAt = file.availableTo ? new Date(file.availableTo) : null;
  const availableFrom = file.availableFrom ? new Date(file.availableFrom) : null;
  const isExpired = expiresAt ? now > expiresAt : false;
  const isPending = availableFrom ? now < availableFrom : false;
  const isActive = !isExpired && !isPending;
  
  // Get filtered sharedWith list (excluding owner)
  const sharedWithFiltered = getSharedWithFiltered();

  return (
    <div className="container mx-auto p-4 sm:p-6">
      <div className="bg-white shadow-md rounded-lg p-6">
        <div className="flex justify-between items-start mb-6">
          <div>
            <h1 className="text-2xl font-bold mb-2">{file.fileName}</h1>
            <p className="text-gray-600">
              {humanFileSize(file.fileSize)} • {file.mimeType}
            </p>
          </div>
          <div className="flex gap-2">
            <button
              onClick={handleCopyLink}
              className="inline-flex items-center gap-2 px-4 py-2 bg-gray-100 hover:bg-gray-200 rounded-md text-sm font-medium"
            >
              <Copy className="h-4 w-4" />
              Sao chép link
            </button>
            {isActive && (
              <button
                onClick={handleDownload}
                className="inline-flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-sm font-medium"
              >
                <Download className="h-4 w-4" />
                Tải xuống
              </button>
            )}
          </div>
        </div>

        {/* Status Badge */}
        <div className="mb-6">
          {isExpired && (
            <div className="p-4 bg-red-100 text-red-700 rounded-lg font-medium text-center">
              🔴 File đã hết hạn và bị xóa
            </div>
          )}
          {isPending && availableFrom && (
            <div className="p-4 bg-yellow-100 text-yellow-700 rounded-lg font-medium text-center">
              🟡 File chưa đến thời gian mở khóa
              <div className="text-sm mt-2">
                Sẽ mở vào: {formatDate(file.availableFrom)}
              </div>
            </div>
          )}
          {isActive && (
            <div className="p-4 bg-green-100 text-green-700 rounded-lg font-medium text-center">
              ✅ File đang hoạt động
            </div>
          )}
        </div>

        {/* File Details */}
        <div className="grid grid-cols-1 md:grid-cols-2 gap-6 mb-6">
          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-medium text-gray-500 mb-2">Thông tin cơ bản</h3>
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  <span className="text-sm text-gray-600">Share Token:</span>
                  <code className="text-sm bg-gray-100 px-2 py-1 rounded">{file.shareToken}</code>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-sm text-gray-600">Trạng thái:</span>
                  <span className={`px-2 py-1 text-xs font-semibold rounded-full ${
                    file.status === 'active' ? 'bg-green-100 text-green-800' :
                    file.status === 'expired' ? 'bg-yellow-100 text-yellow-800' :
                    'bg-gray-100 text-gray-800'
                  }`}>
                    {file.status === 'active' ? 'Đang hoạt động' : 
                     file.status === 'expired' ? 'Hết hạn' : 'Chờ mở'}
                  </span>
                </div>
                <div className="flex items-center gap-2">
                  <Calendar className="h-4 w-4 text-gray-400" />
                  <span className="text-sm text-gray-600">Ngày tạo: {formatDate(file.createdAt)}</span>
                </div>
              </div>
            </div>
          </div>

          <div className="space-y-4">
            <div>
              <h3 className="text-sm font-medium text-gray-500 mb-2">Bảo mật</h3>
              <div className="space-y-2">
                <div className="flex items-center gap-2">
                  {file.hasPassword ? (
                    <>
                      <Lock className="h-4 w-4 text-red-500" />
                      <span className="text-sm text-gray-600">Có mật khẩu</span>
                    </>
                  ) : (
                    <>
                      <Lock className="h-4 w-4 text-gray-400" />
                      <span className="text-sm text-gray-600">Không có mật khẩu</span>
                    </>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  {file.isPublic ? (
                    <>
                      <Users className="h-4 w-4 text-green-500" />
                      <span className="text-sm text-gray-600">Công khai</span>
                    </>
                  ) : (
                    <>
                      <Users className="h-4 w-4 text-red-500" />
                      <span className="text-sm text-gray-600">Riêng tư</span>
                    </>
                  )}
                </div>
              </div>
            </div>
          </div>
        </div>

        {/* Time Validity (if owner) */}
        {isOwner && (
          <div className="mb-6 p-4 bg-gray-50 rounded-lg">
            <h3 className="text-sm font-medium text-gray-700 mb-3 flex items-center gap-2">
              <Clock className="h-4 w-4" />
              Thời gian hiệu lực
            </h3>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <span className="text-sm text-gray-600">Từ:</span>
                <p className="text-sm font-medium">{formatDate(file.availableFrom)}</p>
              </div>
              <div>
                <span className="text-sm text-gray-600">Đến:</span>
                <p className="text-sm font-medium">{formatDate(file.availableTo)}</p>
              </div>
            </div>
            {file.hoursRemaining !== undefined && (
              <div className="mt-2">
                <span className="text-sm text-gray-600">Còn lại: </span>
                <span className="text-sm font-medium">{file.hoursRemaining.toFixed(1)} giờ</span>
              </div>
            )}
          </div>
        )}

        {/* Shared With (if owner) */}
        {sharedWithFiltered.length > 0 && (
          <div className="mb-6 p-4 bg-gray-50 rounded-lg">
            <h3 className="text-sm font-medium text-gray-700 mb-3 flex items-center gap-2">
              <Users className="h-4 w-4" />
              Chia sẻ với ({sharedWithFiltered.length} người)
            </h3>
            <div className="flex flex-wrap gap-2">
              {sharedWithFiltered.map((email) => (
                <span
                  key={email}
                  className="inline-flex items-center gap-2 rounded-full bg-blue-50 px-3 py-1 text-sm text-blue-700"
                >
                  {email}
                </span>
              ))}
            </div>
          </div>
        )}

        {/* Owner Info */}
        {file.owner && (
          <div className="mb-6 p-4 bg-gray-50 rounded-lg">
            <h3 className="text-sm font-medium text-gray-700 mb-2">Người upload</h3>
            <p className="text-sm text-gray-600">
              {file.owner.username} ({file.owner.email})
            </p>
          </div>
        )}

        {/* Actions */}
        <div className="flex gap-4">
          <Link
            href="/files/my"
            className="px-4 py-2 bg-gray-200 hover:bg-gray-300 rounded-md text-sm font-medium"
          >
            Quay lại danh sách
          </Link>
          {file.shareLink && (
            <Link
              href={`/f/${file.shareToken}`}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md text-sm font-medium"
            >
              Xem file chia sẻ
            </Link>
          )}
        </div>
      </div>
    </div>
  );
}


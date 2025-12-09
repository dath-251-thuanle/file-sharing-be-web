"use client";

import React, { useState, useEffect, useMemo } from "react";
import { useParams, useRouter, useSearchParams } from "next/navigation";
import { toast } from "sonner";
import { getFileInfoByToken } from "@/lib/api/file";
import { getAccessToken, authenticatedFetch } from "@/lib/api/helper";

export default function Page() {
    const params = useParams();
    const router = useRouter();
    const searchParams = useSearchParams();
    const { token } = params as { token: string };
    const autoDownload = searchParams.get("autoDownload") === "true";
    const [password, setPassword] = useState("");
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState<string | null>(null);
    const [downloaded, setDownloaded] = useState(false);
    const [fileInfo, setFileInfo] = useState<any>(null);
    const [checkingAuth, setCheckingAuth] = useState(true);

    const downloadLink = useMemo(() => {
        const origin = typeof window !== "undefined" ? window.location.origin : "";
        const pwd = password ? `?password=${encodeURIComponent(password)}` : "";
        return `${origin}/api/files/${token}/download${pwd}`;
    }, [token, password]);

    const [previewError, setPreviewError] = useState(false);
    
    // Build preview URL with authentication headers
    const buildPreviewUrl = () => {
        const origin = typeof window !== "undefined" ? window.location.origin : "";
        const accessToken = getAccessToken();
        
        // Build URL with headers encoded as query params for iframe/img src
        // Note: For authenticated requests, we'll need to use fetch + blob URL
        let url = `${origin}/api/files/${token}/preview`;
        const params = new URLSearchParams();
        
        if (accessToken) {
            // For images/iframes, we can't set Authorization header directly
            // We'll need to fetch with credentials and create blob URL
            return null; // Signal to use fetch approach
        }
        
        return url;
    };

    const previewLink = useMemo(() => {
        return buildPreviewUrl();
    }, [token, password]);

    // For authenticated preview, fetch and create blob URL
    const [previewBlobUrl, setPreviewBlobUrl] = useState<string | null>(null);
    
    useEffect(() => {
        if (!fileInfo?.file?.mimeType) return;
        
        const accessToken = getAccessToken();
        if (!accessToken && !password) {
            // Public file without password
            setPreviewBlobUrl(`${window.location.origin}/api/files/${token}/preview`);
            return;
        }

        // Need to fetch with authentication
        const fetchPreview = async () => {
            try {
                const headers: HeadersInit = {};
                if (accessToken) {
                    headers['Authorization'] = `Bearer ${accessToken}`;
                }
                if (password) {
                    headers['X-File-Password'] = password;
                }

                const response = await fetch(`${window.location.origin}/api/files/${token}/preview`, {
                    headers,
                });

                if (!response.ok) {
                    setPreviewError(true);
                    return;
                }

                const blob = await response.blob();
                const blobUrl = URL.createObjectURL(blob);
                setPreviewBlobUrl(blobUrl);
            } catch (err) {
                console.error('Preview fetch error:', err);
                setPreviewError(true);
            }
        };

        fetchPreview();

        // Cleanup blob URL on unmount
        return () => {
            if (previewBlobUrl) {
                URL.revokeObjectURL(previewBlobUrl);
            }
        };
    }, [token, password, fileInfo?.file?.mimeType]);

    // Copy link chia sẻ trang public
    const copyLink = () => {
        const link = `${window.location.origin}/f/${token}`;
        navigator.clipboard.writeText(link).then(() => {
            toast.success("Đã sao chép link chia sẻ");
        }).catch(() => {
            toast.error("Không thể sao chép link");
        });
    };

    // Trạng thái đơn giản để giữ UI
    const [isActive, setIsActive] = useState(true);

    // Check file info and auth requirement on mount
    useEffect(() => {
        const normalizeFileInfo = (info: any) => {
            if (!info || !info.file) return info;
            const f = info.file;
            return {
                ...info,
                file: {
                    ...f,
                    fileSize: f.fileSize ?? f.file_size,
                    mimeType: f.mimeType ?? f.mime_type,
                }
            };
        };

        const checkFileAuth = async () => {
            setCheckingAuth(true);
            try {
                const info = await getFileInfoByToken(token);
                setFileInfo(normalizeFileInfo(info));
                setError(null);
            } catch (err: any) {
                const errorMessage = err?.response?.data?.message || err?.message || "";
                
                // GetFileInfo endpoint should not return 401 for whitelist files
                // It only returns basic file info. If we get 401, it means token is invalid
                // But we should still show the page and let user try to download
                if (err?.response?.status === 401) {
                    // Token might be invalid, but we still want to show file info if possible
                    // Don't redirect immediately - let user see the page and try download
                    // Try to get file info without auth
                    try {
                        const publicInfo = await getFileInfoByToken(token);
                        setFileInfo(normalizeFileInfo(publicInfo));
                        setError(null);
                    } catch (publicErr: any) {
                        // If public endpoint also fails, show error but don't redirect
                        setError("Không thể tải thông tin file. Vui lòng thử lại.");
                    }
                } else if (err?.response?.status === 403) {
                    // 403 Forbidden - user is logged in but not authorized (not in whitelist)
                    setError(errorMessage || "Bạn không có quyền truy cập file này. Email của bạn không nằm trong danh sách được phép.");
                    toast.error(errorMessage || "Bạn không có quyền truy cập file này.");
                } else {
                    setError(errorMessage || "Không thể tải thông tin file");
                }
            } finally {
                setCheckingAuth(false);
            }
        };

        if (token) {
            checkFileAuth();
        }
    }, [token, router, autoDownload]);

    // Auto-download after login if autoDownload parameter is set and file info is loaded
    // Use ref to prevent multiple triggers
    const autoDownloadTriggered = React.useRef(false);
    useEffect(() => {
        if (autoDownload && fileInfo && getAccessToken() && !loading && !downloaded && !autoDownloadTriggered.current) {
            autoDownloadTriggered.current = true;
            // Small delay to ensure UI is ready
            const timer = setTimeout(() => {
                download();
            }, 500);
            return () => clearTimeout(timer);
        }
    }, [autoDownload, fileInfo, loading, downloaded]);

    const download = async () => {
        if (!isActive) return;
        // Prevent multiple simultaneous downloads
        if (loading) {
            return;
        }
        setLoading(true);
        setError(null);
        try {
            // Prepare headers - only password if provided
            // authenticatedFetch will automatically add Authorization header
            const headers: HeadersInit = {};
            if (password) {
                headers['X-File-Password'] = password;
            }
            
            // Use authenticatedFetch to ensure token is automatically included
            // authenticatedFetch will add Authorization header automatically if token exists
            let response: Response;
            try {
                response = await authenticatedFetch(downloadLink, {
                    method: 'GET',
                    headers: headers, // Only password header if provided
                });
            } catch (fetchError: any) {
                throw new Error(`Network error: ${fetchError?.message || 'Failed to fetch'}`);
            }

            // Check if response is OK
            if (!response.ok) {
                // Try to parse error message
                let errorMessage = "Không thể tải xuống";
                let errorData: any = {};
                try {
                    const responseText = await response.text();
                    try {
                        errorData = JSON.parse(responseText);
                        errorMessage = errorData.message || errorData.error || errorMessage;
                    } catch {
                        // Not JSON, use response text as error message
                        errorMessage = responseText || response.statusText || errorMessage;
                    }
                } catch {
                    // If can't read response, use status text
                    errorMessage = response.statusText || errorMessage;
                }

                // Handle 401 Unauthorized - file requires authentication
                if (response.status === 401) {
                    // Check if user already has token in localStorage
                    const hasToken = !!getAccessToken();
                    
                    if (!hasToken) {
                        // No token in localStorage - redirect to login
                        toast.error("File này yêu cầu đăng nhập. Vui lòng đăng nhập để tải xuống.");
                        setTimeout(() => {
                            const redirectUrl = `/f/${token}?autoDownload=true`;
                            router.push(`/login?redirect=${encodeURIComponent(redirectUrl)}`);
                        }, 2000);
                        return;
                    } else if (errorData.noTokenInRequest) {
                        try {
                            const retryResponse = await authenticatedFetch(downloadLink, {
                                method: 'GET',
                                headers: password ? { 'X-File-Password': password } : {},
                            });
                            
                            if (!retryResponse.ok) {
                                // Still failed, handle error
                                const retryErrorData = await retryResponse.json().catch(() => ({}));
                                const retryErrorMessage = retryErrorData.message || retryErrorData.error || "Không thể tải xuống";
                                
                                if (retryResponse.status === 401) {
                                    setError("Token không hợp lệ hoặc đã hết hạn. Vui lòng đăng nhập lại.");
                                    toast.error("Token không hợp lệ. Vui lòng đăng nhập lại.");
                                    setTimeout(() => {
                                        const redirectUrl = `/f/${token}?autoDownload=true`;
                                        router.push(`/login?redirect=${encodeURIComponent(redirectUrl)}`);
                                    }, 2000);
                                    return;
                                } else if (retryResponse.status === 403) {
                                    setError(retryErrorMessage || "Bạn không có quyền truy cập file này.");
                                    toast.error(retryErrorMessage || "Bạn không có quyền truy cập file này.");
                                    return;
                                } else {
                                    setError(retryErrorMessage);
                                    toast.error(retryErrorMessage);
                                    return;
                                }
                            }
                            
                            // Retry successful, continue with download
                            response = retryResponse;
                        } catch (retryError: any) {
                            setError("Không thể tải file. Vui lòng thử lại.");
                            toast.error("Không thể tải file. Vui lòng thử lại.");
                            return;
                        }
                    } else {
                        // Has token but still 401 - token might be invalid or expired
                        setError("Token không hợp lệ hoặc đã hết hạn. Vui lòng đăng nhập lại.");
                        toast.error("Token không hợp lệ. Vui lòng đăng nhập lại.");
                        setTimeout(() => {
                            const redirectUrl = `/f/${token}?autoDownload=true`;
                            router.push(`/login?redirect=${encodeURIComponent(redirectUrl)}`);
                        }, 2000);
                        return;
                    }
                }

                // Handle 403 Forbidden - user is logged in but not authorized (not in whitelist and not owner)
                if (response.status === 403) {
                    // User has token but email is not in whitelist - show error, don't redirect to login
                    setError(errorMessage || "Bạn không có quyền truy cập file này. Email của bạn không nằm trong danh sách được phép.");
                    toast.error(errorMessage || "Bạn không có quyền truy cập file này.");
                    return;
                }

                // Other errors
                setError(errorMessage);
                toast.error(errorMessage);
                return;
            }

            // If OK, trigger download
            // Check if response is HTML (error page) instead of file
            const contentType = response.headers.get('Content-Type') || '';
            if (contentType.includes('text/html')) {
                setError('Không thể tải file. Backend trả về HTML thay vì file binary.');
                toast.error('Lỗi: Backend trả về HTML. Vui lòng kiểm tra backend logs.');
                return;
            }
            
            let blob: Blob;
            try {
                blob = await response.blob();
            } catch (blobError: any) {
                throw new Error(`Failed to create blob: ${blobError?.message || 'Unknown error'}`);
            }
            
            const url = window.URL.createObjectURL(blob);
            const a = document.createElement('a');
            a.href = url;
            
            // Get filename from Content-Disposition header or use token
            const contentDisposition = response.headers.get('Content-Disposition');
            let filename = `file-${token}`;
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
            
            setDownloaded(true);
            toast.success("Đã bắt đầu tải file!");
        } catch (e: any) {
            const errorMessage = e?.response?.data?.message || e?.message || "Không thể tải xuống";
            setError(errorMessage);
            toast.error(errorMessage);
        } finally {
            setLoading(false);
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

    // Note: sharedWith (whitelist) is NOT shown on public download page
    // Only owner can see whitelist via /files/info/{id} endpoint

    return (
        <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-start justify-center py-12 px-4">
            <div className="w-full max-w-5xl">

                {/* HEADER */}
                <div className="text-center mb-10">
                    <h1 className="text-4xl font-bold text-gray-800 mb-2">
                        File được chia sẻ với bạn
                    </h1>
                    <p className="text-gray-600">Token: {token}</p>
                    {checkingAuth && (
                        <p className="text-blue-600 mt-2">Đang kiểm tra quyền truy cập...</p>
                    )}
                </div>

                {/* SUCCESS */}
                {downloaded && (
                    <div className="mb-6 p-4 bg-green-100 border border-green-300 text-green-800 rounded-lg text-center font-medium">
                        ✅ Đã bắt đầu tải file thành công!
                    </div>
                )}

                {/* ERROR */}
                {error && (
                    <div className="mb-6 p-4 bg-red-100 border border-red-300 text-red-800 rounded-lg">
                        {error}
                    </div>
                )}

                <div className="grid md:grid-cols-2 gap-8">

                    {/* LEFT PREVIEW */}
                    <div className="order-2 md:order-1">
                        <div className="bg-white rounded-2xl shadow-xl overflow-hidden">
                            <div className="bg-indigo-600 text-white p-4 text-center font-medium">
                                Xem trước file
                            </div>
                            <div className="p-6">
                                <div className="bg-gray-50 border-2 border-dashed rounded-xl w-full h-96 flex flex-col items-center justify-center text-gray-600 overflow-hidden">
                                    {previewError ? (
                                        <>
                                            <div className="text-8xl mb-4">⚠️</div>
                                            <p className="text-lg font-medium">Không thể tải preview</p>
                                            <p className="text-sm mt-2">Vui lòng kiểm tra quyền truy cập</p>
                                        </>
                                    ) : !previewBlobUrl ? (
                                        <>
                                            <div className="text-8xl mb-4">⏳</div>
                                            <p className="text-lg font-medium">Đang tải preview...</p>
                                        </>
                                    ) : fileInfo?.file?.mimeType ? (
                                        <>
                                            {/* Image Preview */}
                                            {fileInfo.file.mimeType.startsWith('image/') && (
                                                <img 
                                                    src={previewBlobUrl} 
                                                    alt={fileInfo.file.fileName || "Preview"} 
                                                    className="max-w-full max-h-96 object-contain"
                                                    onError={(e: React.SyntheticEvent<HTMLImageElement>) => {
                                                        setPreviewError(true);
                                                    }}
                                                />
                                            )}
                                            
                                            {/* PDF Preview */}
                                            {fileInfo.file.mimeType === 'application/pdf' && (
                                                <iframe 
                                                    src={previewBlobUrl} 
                                                    className="w-full h-96 border-0"
                                                    title="PDF Preview"
                                                />
                                            )}
                                            
                                            {/* Video Preview */}
                                            {fileInfo.file.mimeType.startsWith('video/') && (
                                                <video 
                                                    src={previewBlobUrl} 
                                                    controls 
                                                    className="max-w-full max-h-96"
                                                >
                                                    Trình duyệt không hỗ trợ video
                                                </video>
                                            )}
                                            
                                            {/* Audio Preview */}
                                            {fileInfo.file.mimeType.startsWith('audio/') && (
                                                <div className="flex flex-col items-center">
                                                    <div className="text-8xl mb-4">🎵</div>
                                                    <audio src={previewBlobUrl} controls className="w-full max-w-md">
                                                        Trình duyệt không hỗ trợ audio
                                                    </audio>
                                                </div>
                                            )}
                                            
                                            {/* Text Preview */}
                                            {(fileInfo.file.mimeType.startsWith('text/') || 
                                              fileInfo.file.mimeType === 'application/json' ||
                                              fileInfo.file.mimeType === 'application/xml') && (
                                                <iframe 
                                                    src={previewBlobUrl} 
                                                    className="w-full h-96 border-0 bg-white p-4"
                                                    title="Text Preview"
                                                />
                                            )}
                                            
                                            {/* Unsupported file types */}
                                            {!fileInfo.file.mimeType.startsWith('image/') && 
                                             !fileInfo.file.mimeType.startsWith('video/') && 
                                             !fileInfo.file.mimeType.startsWith('audio/') && 
                                             !fileInfo.file.mimeType.startsWith('text/') && 
                                             fileInfo.file.mimeType !== 'application/pdf' &&
                                             fileInfo.file.mimeType !== 'application/json' &&
                                             fileInfo.file.mimeType !== 'application/xml' && (
                                                <>
                                                    <div className="text-8xl mb-4">📄</div>
                                                    <p className="text-lg font-medium">Preview không hỗ trợ</p>
                                                    <p className="text-sm mt-2">Loại file: {fileInfo.file.mimeType}</p>
                                                </>
                                            )}
                                        </>
                                    ) : (
                                        <>
                                            <div className="text-8xl mb-4">📄</div>
                                            <p className="text-lg font-medium">Đang tải thông tin file...</p>
                                        </>
                                    )}
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* RIGHT PANEL */}
                    <div className="order-1 md:order-2">
                        <div className="bg-white rounded-2xl shadow-xl p-8">

                            {/* Copy Link */}
                            <button
                                onClick={copyLink}
                                className="mb-4 w-full py-3 rounded-lg bg-gray-200 hover:bg-gray-300 transition"
                            >
                                📋 Sao chép link chia sẻ
                            </button>

                            {/* PASSWORD (nếu cần) */}
                            <div className="mb-6">
                                <input
                                    type="password"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    className="w-full px-4 py-3 border rounded-lg focus:ring-2"
                                    placeholder="Nhập mật khẩu (nếu có)..."
                                />
                            </div>

                            {/* DOWNLOAD BTN */}
                            <button
                                onClick={download}
                                disabled={loading || checkingAuth}
                                className={`w-full py-4 px-6 rounded-xl text-white font-semibold transition-all 
                                            ${loading || checkingAuth ? "bg-gray-400" : "bg-indigo-600 hover:bg-indigo-700"}`}
                            >
                                {checkingAuth ? "Đang kiểm tra..." : loading ? "Đang chuẩn bị..." : "⬇️ Tải xuống"}
                            </button>

                            {/* Thông tin file */}
                            {fileInfo && (
                                <div className="mt-6 text-sm text-gray-600">
                                    <p className="font-semibold">Thông tin file:</p>
                                    <p>Tên: {fileInfo.file?.fileName || "N/A"}</p>
                                    <p>Kích thước: {fileInfo.file?.fileSize ? humanFileSize(fileInfo.file.fileSize) : "N/A"}</p>
                                    <p>Định dạng: {fileInfo.file?.mimeType || "N/A"}</p>
                                </div>
                            )}
                            
                            {!fileInfo && !checkingAuth && (
                                <div className="mt-6 text-sm text-gray-600">
                                    <p className="font-semibold">Thông tin (tham khảo):</p>
                                    <p>Kích thước: không xác định</p>
                                    <p>Định dạng: không xác định</p>
                                </div>
                            )}
                        </div>
                    </div>
                </div>

            </div>
        </div>
    );
}

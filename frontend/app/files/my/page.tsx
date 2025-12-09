"use client";

import { useEffect, useState } from "react";
import Link from "next/link";
import { useRouter } from "next/navigation";
import { File as FileType, UserFilesResponse } from "@/lib/components/schemas";
import { getUserFiles } from "@/lib/api/file";
import { Loader, Trash2, Eye } from "lucide-react";
import { toast } from "sonner";
import { deleteFile } from "@/lib/api/file";

export default function MyFilesPage() {
  const [files, setFiles] = useState<FileType[]>([]);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState("all");
  const [sortBy, setSortBy] = useState("createdAt");
  const [sortOrder, setSortOrder] = useState("desc");
  const [currentPage, setCurrentPage] = useState(1);
  const [pagination, setPagination] = useState({
    currentPage: 1,
    totalPages: 1,
    totalFiles: 0,
    limit: 20,
  });
  const [summary, setSummary] = useState({
    activeFiles: 0,
    pendingFiles: 0,
    expiredFiles: 0,
    deletedFiles: 0,
  });

  const router = useRouter();

  useEffect(() => {
    const fetchFiles = async () => {
      setIsLoading(true);
      try {
        const response: UserFilesResponse = await getUserFiles({
          status: statusFilter,
          page: currentPage,
          limit: pagination.limit,
          sortBy,
          order: sortOrder,
        });
        
        setFiles(response.files);
        setPagination(response.pagination);
        setSummary(response.summary);
      } catch (err: any) {
        if (err.message?.includes("Unauthorized") || err.status === 401) {
          router.push("/login");
        } else {
          setError("Không thể tải danh sách file.");
          toast.error("Không thể tải danh sách file.");
        }
      } finally {
        setIsLoading(false);
      }
    };

    fetchFiles();
  }, [router, statusFilter, sortBy, sortOrder, currentPage, pagination.limit]);

  const handleDelete = async (fileId: string) => {
    if (window.confirm("Bạn có chắc chắn muốn xóa file này? Hành động này không thể hoàn tác.")) {
      try {
        await deleteFile(fileId);
        toast.success("File đã được xóa thành công");
        // Re-fetch files after deletion
        setCurrentPage(1);
      } catch (err: any) {
        toast.error(`Không thể xóa file: ${err.message}`);
      }
    }
  };

  if (isLoading) {
    return (
      <div className="flex justify-center items-center h-64">
        <Loader className="animate-spin h-8 w-8 text-gray-500" />
        <p className="ml-2 text-gray-500">Đang tải danh sách file...</p>
      </div>
    );
  }

  if (error) {
    return <div className="text-center text-red-500">{error}</div>;
  }

  return (
    <div className="container mx-auto p-4 sm:p-6">
      <div className="bg-white shadow-md rounded-lg p-6 mb-8">
        <h1 className="text-2xl font-bold mb-4">File của tôi</h1>
        
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-4 text-center mb-6">
          <div className="p-4 bg-blue-100 rounded-lg">
            <p className="text-2xl font-bold">{summary.activeFiles}</p>
            <p className="text-sm text-blue-800">Đang hoạt động</p>
          </div>
          <div className="p-4 bg-gray-100 rounded-lg">
            <p className="text-2xl font-bold">{summary.pendingFiles}</p>
            <p className="text-sm text-gray-800">Chờ mở</p>
          </div>
          <div className="p-4 bg-yellow-100 rounded-lg">
            <p className="text-2xl font-bold">{summary.expiredFiles}</p>
            <p className="text-sm text-yellow-800">Hết hạn</p>
          </div>
          <div className="p-4 bg-red-100 rounded-lg">
            <p className="text-2xl font-bold">{summary.deletedFiles}</p>
            <p className="text-sm text-red-800">Đã xóa</p>
          </div>
        </div>

        <div className="flex justify-between items-center mb-4">
          <div className="flex items-center gap-4">
            <div>
              <label htmlFor="status-filter" className="sr-only">Lọc theo trạng thái</label>
              <select
                id="status-filter"
                value={statusFilter}
                onChange={(e) => setStatusFilter(e.target.value)}
                className="block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
              >
                <option value="all">Tất cả trạng thái</option>
                <option value="active">Đang hoạt động</option>
                <option value="pending">Chờ mở</option>
                <option value="expired">Hết hạn</option>
              </select>
            </div>
            <div>
              <label htmlFor="sort-by" className="sr-only">Sắp xếp theo</label>
              <select
                id="sort-by"
                value={`${sortBy}-${sortOrder}`}
                onChange={(e) => {
                  const [newSortBy, newSortOrder] = e.target.value.split('-');
                  setSortBy(newSortBy);
                  setSortOrder(newSortOrder);
                }}
                className="block w-full pl-3 pr-10 py-2 text-base border-gray-300 focus:outline-none focus:ring-indigo-500 focus:border-indigo-500 sm:text-sm rounded-md"
              >
                <option value="createdAt-desc">Mới nhất</option>
                <option value="createdAt-asc">Cũ nhất</option>
                <option value="fileName-asc">Tên (A-Z)</option>
                <option value="fileName-desc">Tên (Z-A)</option>
              </select>
            </div>
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="min-w-full divide-y divide-gray-200">
            <thead className="bg-gray-50">
              <tr>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Tên file
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Trạng thái
                </th>
                <th scope="col" className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Ngày tạo
                </th>
                <th scope="col" className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                  Thao tác
                </th>
              </tr>
            </thead>
            <tbody className="bg-white divide-y divide-gray-200">
              {files.length > 0 ? (
                files.map((file) => (
                  <tr key={file.id}>
                    <td className="px-6 py-4 whitespace-nowrap text-sm font-medium">
                      <Link href={`/files/info/${file.id}`} className="text-indigo-600 hover:text-indigo-900">
                        {file.fileName}
                      </Link>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm">
                      <span className={`px-2 inline-flex text-xs leading-5 font-semibold rounded-full ${
                        file.status === 'active' ? 'bg-green-100 text-green-800' :
                        file.status === 'expired' ? 'bg-yellow-100 text-yellow-800' :
                        'bg-gray-100 text-gray-800'
                      }`}>
                        {file.status === 'active' ? 'Đang hoạt động' : 
                         file.status === 'expired' ? 'Hết hạn' : 'Chờ mở'}
                      </span>
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                      {new Date(file.createdAt).toLocaleString('vi-VN')}
                    </td>
                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                      <div className="flex justify-end gap-2">
                        <Link 
                          href={`/files/info/${file.id}`}
                          className="text-blue-600 hover:text-blue-900"
                          title="Xem chi tiết"
                        >
                          <Eye className="h-5 w-5" />
                        </Link>
                        <button 
                          onClick={() => handleDelete(file.id)} 
                          className="text-red-600 hover:text-red-900"
                          title="Xóa file"
                        >
                          <Trash2 className="h-5 w-5" />
                        </button>
                      </div>
                    </td>
                  </tr>
                ))
              ) : (
                <tr>
                  <td colSpan={4} className="px-6 py-4 text-center text-sm text-gray-500">
                    Bạn chưa upload file nào.
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <div className="mt-4 flex justify-between items-center">
          <div>
            <p className="text-sm text-gray-700">
              Hiển thị <span className="font-medium">{(currentPage - 1) * pagination.limit + 1}</span> đến{' '}
              <span className="font-medium">{Math.min(currentPage * pagination.limit, pagination.totalFiles)}</span> trong{' '}
              <span className="font-medium">{pagination.totalFiles}</span> kết quả
            </p>
          </div>
          <div>
            <nav className="relative z-0 inline-flex rounded-md shadow-sm -space-x-px" aria-label="Pagination">
              <button
                onClick={() => setCurrentPage(currentPage - 1)}
                disabled={currentPage === 1}
                className="relative inline-flex items-center px-2 py-2 rounded-l-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:bg-gray-100"
              >
                Trước
              </button>
              {Array.from({ length: pagination.totalPages }, (_, i) => i + 1).map((page) => (
                <button
                  key={page}
                  onClick={() => setCurrentPage(page)}
                  className={`relative inline-flex items-center px-4 py-2 border border-gray-300 bg-white text-sm font-medium ${
                    currentPage === page ? 'text-indigo-600 bg-indigo-50' : 'text-gray-700 hover:bg-gray-50'
                  }`}
                >
                  {page}
                </button>
              ))}
              <button
                onClick={() => setCurrentPage(currentPage + 1)}
                disabled={currentPage === pagination.totalPages}
                className="relative inline-flex items-center px-2 py-2 rounded-r-md border border-gray-300 bg-white text-sm font-medium text-gray-500 hover:bg-gray-50 disabled:bg-gray-100"
              >
                Sau
              </button>
            </nav>
          </div>
        </div>
      </div>
    </div>
  );
}


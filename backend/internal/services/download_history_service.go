package services

import (
	"time"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ repositories.DownloadHistoryRepository = (*DownloadHistoryService)(nil)

type DownloadHistoryService struct {
	db *gorm.DB
}

func NewDownloadHistoryService(db *gorm.DB) *DownloadHistoryService {
	return &DownloadHistoryService{db: db}
}

// Lấy chi tiết 1 dòng lịch sử cụ thể
func (s *DownloadHistoryService) GetByID(id uuid.UUID) (*models.DownloadHistory, error) {
	var history models.DownloadHistory
	err := s.db.Preload("Downloader").Preload("File").First(&history, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &history, nil
}

// Create: Ghi lại hành động download (Dùng cho API Download)
func (s *DownloadHistoryService) Create(history *models.DownloadHistory) error {
	return s.db.Create(history).Error
}

// Lấy danh sách ai đã tải file này (Dùng cho API History)
func (s *DownloadHistoryService) GetByFileID(fileID uuid.UUID, limit, offset int) ([]models.DownloadHistory, int64, error) {
	var history []models.DownloadHistory
	var total int64

	// Đếm tổng
	if err := s.db.Model(&models.DownloadHistory{}).Where("file_id = ?", fileID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Lấy dữ liệu (Preload Downloader để biết tên người tải)
	err := s.db.Where("file_id = ?", fileID).
		Preload("Downloader").
		Order("downloaded_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&history).Error

	if err != nil {
		return nil, 0, err
	}

	return history, total, nil
}

// GetByDownloaderID: Xem "Tôi đã tải những file nào" (Cho trang Profile User)
func (s *DownloadHistoryService) GetByDownloaderID(downloaderID uuid.UUID, limit, offset int) ([]models.DownloadHistory, int64, error) {
	var history []models.DownloadHistory
	var total int64

	if err := s.db.Model(&models.DownloadHistory{}).Where("downloader_id = ?", downloaderID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Preload File để hiện tên file đã tải
	err := s.db.Where("downloader_id = ?", downloaderID).
		Preload("File").
		Order("downloaded_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&history).Error

	if err != nil {
		return nil, 0, err
	}

	return history, total, nil
}

// GetByFileIDAndDownloaderID: Kiểm tra xem User A đã từng tải File B chưa?
func (s *DownloadHistoryService) GetByFileIDAndDownloaderID(fileID uuid.UUID, downloaderID uuid.UUID) ([]models.DownloadHistory, error) {
	var history []models.DownloadHistory
	err := s.db.Where("file_id = ? AND downloader_id = ?", fileID, downloaderID).
		Find(&history).Error
	return history, err
}

// GetDownloadCount: Đếm tổng lượt tải từ bảng History (để đối chiếu với bảng Stats)
func (s *DownloadHistoryService) GetDownloadCount(fileID uuid.UUID) (int64, error) {
	var count int64
	err := s.db.Model(&models.DownloadHistory{}).
		Where("file_id = ? AND download_completed = ?", fileID, true).
		Count(&count).Error
	return count, err
}

// GetUniqueDownloaderCount: Đếm số người tải unique thực tế
func (s *DownloadHistoryService) GetUniqueDownloaderCount(fileID uuid.UUID) (int64, error) {
	var count int64
	// Đếm số lượng downloader_id khác nhau (loại bỏ null/anonymous)
	err := s.db.Model(&models.DownloadHistory{}).
		Where("file_id = ? AND downloader_id IS NOT NULL", fileID).
		Distinct("downloader_id").
		Count(&count).Error
	return count, err
}

// GetDownloadsInRange: Thống kê tải theo ngày/tháng (Cho biểu đồ Admin)
func (s *DownloadHistoryService) GetDownloadsInRange(fileID uuid.UUID, startTime, endTime time.Time) ([]models.DownloadHistory, error) {
	var history []models.DownloadHistory
	err := s.db.Where("file_id = ? AND downloaded_at BETWEEN ? AND ?", fileID, startTime, endTime).
		Find(&history).Error
	return history, err
}

// Delete: Xóa 1 dòng lịch sử
func (s *DownloadHistoryService) Delete(id uuid.UUID) error {
	return s.db.Delete(&models.DownloadHistory{}, "id = ?", id).Error
}

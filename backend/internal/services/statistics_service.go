package services

import (
	"time"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/repositories"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ repositories.StatisticsRepository = (*StatisticsService)(nil)

type StatisticsService struct {
	db *gorm.DB
}

func NewStatisticsService(db *gorm.DB) *StatisticsService {
	return &StatisticsService{db: db}
}

func (s *StatisticsService) GetByFileID(fileID uuid.UUID) (*models.FileStatistics, error) {
	var stats models.FileStatistics
	err := s.db.Where("file_id = ?", fileID).First(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// Tạo mới thống kê (Gọi khi vừa Upload file xong)
func (s *StatisticsService) Create(stats *models.FileStatistics) error {
	return s.db.Create(stats).Error
}

func (s *StatisticsService) Update(stats *models.FileStatistics) error {
	return s.db.Save(stats).Error
}

// Hàm này sẽ được gọi bên trong API Download
func (s *StatisticsService) IncrementDownloadCount(fileID uuid.UUID) error {
	// Dùng SQL raw hoặc GORM update expression để tránh race condition
	return s.db.Model(&models.FileStatistics{}).
		Where("file_id = ?", fileID).
		UpdateColumn("download_count", gorm.Expr("download_count + ?", 1)).
		Error
}

func (s *StatisticsService) IncrementUniqueDownloaders(fileID uuid.UUID) error {
	return s.db.Model(&models.FileStatistics{}).
		Where("file_id = ?", fileID).
		UpdateColumn("unique_downloaders", gorm.Expr("unique_downloaders + ?", 1)).
		Error
}

func (s *StatisticsService) UpdateLastDownloadedAt(fileID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&models.FileStatistics{}).
		Where("file_id = ?", fileID).
		Update("last_downloaded_at", now).
		Error
}

func (s *StatisticsService) GetByFileIDs(fileIDs []uuid.UUID) ([]models.FileStatistics, error) {
	var stats []models.FileStatistics
	err := s.db.Where("file_id IN ?", fileIDs).Find(&stats).Error
	return stats, err
}

func (s *StatisticsService) Delete(fileID uuid.UUID) error {
	return s.db.Where("file_id = ?", fileID).Delete(&models.FileStatistics{}).Error
}

package services

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StatisticsService struct {
	db *gorm.DB
}

func NewStatisticsService(db *gorm.DB) *StatisticsService {
	return &StatisticsService{db: db}
}

// GetByFileID lấy thông tin thống kê của file từ database
func (s *StatisticsService) GetByFileID(fileID uuid.UUID) (*models.FileStatistics, error) {
	var stats models.FileStatistics

	// Tìm trong bảng file_statistics bản ghi có file_id trùng khớp
	// Preload("File") là tùy chọn nếu bạn muốn lấy luôn thông tin file đi kèm
	err := s.db.Where("file_id = ?", fileID).First(&stats).Error
	if err != nil {
		return nil, err
	}

	return &stats, nil
}

package repositories

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/google/uuid"
)

type StatisticsRepository interface {
	GetByFileID(fileID uuid.UUID) (*models.FileStatistics, error)
	Create(stats *models.FileStatistics) error
	Update(stats *models.FileStatistics) error
	IncrementDownloadCount(fileID uuid.UUID) error
	IncrementUniqueDownloaders(fileID uuid.UUID) error
	UpdateLastDownloadedAt(fileID uuid.UUID) error
	GetByFileIDs(fileIDs []uuid.UUID) ([]models.FileStatistics, error)
	Delete(fileID uuid.UUID) error
}
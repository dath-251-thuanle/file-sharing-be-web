package repositories

import (
	"time"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/google/uuid"
)

type DownloadHistoryRepository interface {
	GetByID(id uuid.UUID) (*models.DownloadHistory, error)
	Create(history *models.DownloadHistory) error
	GetByFileID(fileID uuid.UUID, limit, offset int) ([]models.DownloadHistory, int64, error)
	GetByDownloaderID(downloaderID uuid.UUID, limit, offset int) ([]models.DownloadHistory, int64, error)
	GetByFileIDAndDownloaderID(fileID uuid.UUID, downloaderID uuid.UUID) ([]models.DownloadHistory, error)
	GetDownloadCount(fileID uuid.UUID) (int64, error)
	GetUniqueDownloaderCount(fileID uuid.UUID) (int64, error)
	GetDownloadsInRange(fileID uuid.UUID, startTime, endTime time.Time) ([]models.DownloadHistory, error)
	Delete(id uuid.UUID) error
}


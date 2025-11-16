package repositories

import (
	"time"

	"github.com/google/uuid"
	"github.com/yourusername/file-sharing-backend/internal/models"
	"gorm.io/gorm"
)

type StatisticsRepository interface {
	// File Statistics
	GetFileStatistics(fileID uuid.UUID) (*models.FileStatistics, error)
	CreateFileStatistics(fileID uuid.UUID) error
	IncrementDownloadCount(fileID uuid.UUID, downloaderID *uuid.UUID) error
	IncrementViewCount(fileID uuid.UUID) error
	
	// Download History
	RecordDownload(params *models.DownloadHistoryParams) error
	GetDownloadHistory(fileID uuid.UUID, limit, offset int) ([]models.DownloadHistory, int64, error)
	GetUserDownloadHistory(userID uuid.UUID, limit, offset int) ([]models.DownloadHistory, int64, error)
	
	// Analytics
	GetTopDownloadedFiles(ownerID uuid.UUID, limit int) ([]FileStatsWithFile, error)
	GetDailyDownloads(fileID uuid.UUID, days int) ([]DailyDownloadStats, error)
}

type statisticsRepository struct {
	db *gorm.DB
}

func NewStatisticsRepository(db *gorm.DB) StatisticsRepository {
	return &statisticsRepository{db: db}
}

// GetFileStatistics retrieves statistics for a file
func (r *statisticsRepository) GetFileStatistics(fileID uuid.UUID) (*models.FileStatistics, error) {
	var stats models.FileStatistics
	err := r.db.Where("file_id = ?", fileID).First(&stats).Error
	if err != nil {
		return nil, err
	}
	return &stats, nil
}

// CreateFileStatistics initializes statistics for a new file
func (r *statisticsRepository) CreateFileStatistics(fileID uuid.UUID) error {
	stats := &models.FileStatistics{
		FileID:        fileID,
		DownloadCount: 0,
		ViewCount:     0,
	}
	return r.db.Create(stats).Error
}

// IncrementDownloadCount updates download statistics
// This should be called after successfully recording a download
func (r *statisticsRepository) IncrementDownloadCount(fileID uuid.UUID, downloaderID *uuid.UUID) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Upsert file_statistics
		now := time.Now()
		
		// Try to update existing record
		result := tx.Model(&models.FileStatistics{}).
			Where("file_id = ?", fileID).
			Updates(map[string]interface{}{
				"download_count":      gorm.Expr("download_count + 1"),
				"last_downloaded_at":  now,
				"updated_at":          now,
			})
		
		// If no record exists, create one
		if result.RowsAffected == 0 {
			stats := &models.FileStatistics{
				FileID:           fileID,
				DownloadCount:    1,
				LastDownloadedAt: &now,
				UpdatedAt:        now,
			}
			if err := tx.Create(stats).Error; err != nil {
				return err
			}
		}
		
		// Update unique_downloaders count if downloader is authenticated
		if downloaderID != nil {
			var count int64
			err := tx.Model(&models.DownloadHistory{}).
				Where("file_id = ? AND downloader_id IS NOT NULL", fileID).
				Distinct("downloader_id").
				Count(&count).Error
			
			if err != nil {
				return err
			}
			
			tx.Model(&models.FileStatistics{}).
				Where("file_id = ?", fileID).
				Update("unique_downloaders", count)
		}
		
		return nil
	})
}

// IncrementViewCount updates view statistics
func (r *statisticsRepository) IncrementViewCount(fileID uuid.UUID) error {
	now := time.Now()
	
	// Try to update existing record
	result := r.db.Model(&models.FileStatistics{}).
		Where("file_id = ?", fileID).
		Updates(map[string]interface{}{
			"view_count":     gorm.Expr("view_count + 1"),
			"last_viewed_at": now,
			"updated_at":     now,
		})
	
	// If no record exists, create one
	if result.RowsAffected == 0 {
		stats := &models.FileStatistics{
			FileID:       fileID,
			ViewCount:    1,
			LastViewedAt: &now,
			UpdatedAt:    now,
		}
		return r.db.Create(stats).Error
	}
	
	return result.Error
}

// RecordDownload records a download event in history
func (r *statisticsRepository) RecordDownload(params *models.DownloadHistoryParams) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		// Create download history record
		history := &models.DownloadHistory{
			FileID:            params.FileID,
			DownloaderID:      params.DownloaderID,
			IPAddress:         params.IPAddress,
			UserAgent:         params.UserAgent,
			Referer:           params.Referer,
			DownloadCompleted: params.DownloadCompleted,
			BytesTransferred:  params.BytesTransferred,
			AccessMethod:      params.AccessMethod,
			PasswordProtected: params.PasswordProtected,
		}
		
		if err := tx.Create(history).Error; err != nil {
			return err
		}
		
		// Update statistics (handled by separate function)
		repo := &statisticsRepository{db: tx}
		return repo.IncrementDownloadCount(params.FileID, params.DownloaderID)
	})
}

// GetDownloadHistory retrieves paginated download history for a file
func (r *statisticsRepository) GetDownloadHistory(fileID uuid.UUID, limit, offset int) ([]models.DownloadHistory, int64, error) {
	var history []models.DownloadHistory
	var total int64
	
	// Count total
	if err := r.db.Model(&models.DownloadHistory{}).Where("file_id = ?", fileID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	err := r.db.Where("file_id = ?", fileID).
		Order("downloaded_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("Downloader").
		Find(&history).Error
	
	return history, total, err
}

// GetUserDownloadHistory retrieves paginated download history for a user
func (r *statisticsRepository) GetUserDownloadHistory(userID uuid.UUID, limit, offset int) ([]models.DownloadHistory, int64, error) {
	var history []models.DownloadHistory
	var total int64
	
	// Count total
	if err := r.db.Model(&models.DownloadHistory{}).Where("downloader_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	
	// Get paginated results
	err := r.db.Where("downloader_id = ?", userID).
		Order("downloaded_at DESC").
		Limit(limit).
		Offset(offset).
		Preload("File").
		Find(&history).Error
	
	return history, total, err
}

// GetTopDownloadedFiles retrieves top downloaded files for a user
func (r *statisticsRepository) GetTopDownloadedFiles(ownerID uuid.UUID, limit int) ([]FileStatsWithFile, error) {
	var results []FileStatsWithFile
	
	err := r.db.Table("files").
		Select("files.id, files.file_name, files.created_at, file_statistics.download_count, file_statistics.view_count, file_statistics.last_downloaded_at").
		Joins("LEFT JOIN file_statistics ON files.id = file_statistics.file_id").
		Where("files.owner_id = ?", ownerID).
		Order("file_statistics.download_count DESC NULLS LAST").
		Limit(limit).
		Scan(&results).Error
	
	return results, err
}

// GetDailyDownloads retrieves daily download statistics for a file
func (r *statisticsRepository) GetDailyDownloads(fileID uuid.UUID, days int) ([]DailyDownloadStats, error) {
	var results []DailyDownloadStats
	
	err := r.db.Table("download_history").
		Select("DATE(downloaded_at) as date, COUNT(*) as count").
		Where("file_id = ? AND downloaded_at >= NOW() - INTERVAL '? days'", fileID, days).
		Group("DATE(downloaded_at)").
		Order("date DESC").
		Scan(&results).Error
	
	return results, err
}

// Helper types for query results
type FileStatsWithFile struct {
	ID               uuid.UUID  `json:"id"`
	FileName         string     `json:"file_name"`
	CreatedAt        time.Time  `json:"created_at"`
	DownloadCount    int        `json:"download_count"`
	ViewCount        int        `json:"view_count"`
	LastDownloadedAt *time.Time `json:"last_downloaded_at"`
}

type DailyDownloadStats struct {
	Date  time.Time `json:"date"`
	Count int       `json:"count"`
}


package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/repositories"
)

type StatisticsService interface {
	// Get statistics (owner verification required by handler)
	GetFileStatistics(fileID uuid.UUID) (*models.FileStatistics, error)
	GetDownloadHistory(fileID uuid.UUID, page, pageSize int) (*DownloadHistoryResponse, error)
	GetUserDashboard(userID uuid.UUID) (*UserDashboard, error)
	
	// Record events
	RecordDownload(fileID uuid.UUID, downloaderID *uuid.UUID, ipAddress, userAgent, referer *string, passwordProtected bool) error
	RecordView(fileID uuid.UUID) error
}

type statisticsService struct {
	statsRepo repositories.StatisticsRepository
	fileRepo  repositories.FileRepository
}

func NewStatisticsService(statsRepo repositories.StatisticsRepository, fileRepo repositories.FileRepository) StatisticsService {
	return &statisticsService{
		statsRepo: statsRepo,
		fileRepo:  fileRepo,
	}
}

// GetFileStatistics retrieves statistics for a file
func (s *statisticsService) GetFileStatistics(fileID uuid.UUID) (*models.FileStatistics, error) {
	stats, err := s.statsRepo.GetFileStatistics(fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get file statistics: %w", err)
	}
	return stats, nil
}

// GetDownloadHistory retrieves paginated download history
func (s *statisticsService) GetDownloadHistory(fileID uuid.UUID, page, pageSize int) (*DownloadHistoryResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}
	
	offset := (page - 1) * pageSize
	
	history, total, err := s.statsRepo.GetDownloadHistory(fileID, pageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get download history: %w", err)
	}
	
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	
	return &DownloadHistoryResponse{
		Downloads:  history,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// GetUserDashboard retrieves dashboard statistics for a user
func (s *statisticsService) GetUserDashboard(userID uuid.UUID) (*UserDashboard, error) {
	// Get top files
	topFiles, err := s.statsRepo.GetTopDownloadedFiles(userID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get top files: %w", err)
	}
	
	// Calculate totals
	var totalDownloads, totalViews int
	for _, file := range topFiles {
		totalDownloads += file.DownloadCount
		totalViews += file.ViewCount
	}
	
	// Get total files count (would need file repository)
	// For now, use length of topFiles as approximation
	totalFiles := len(topFiles)
	
	return &UserDashboard{
		TotalFiles:     totalFiles,
		TotalDownloads: totalDownloads,
		TotalViews:     totalViews,
		TopFiles:       topFiles,
	}, nil
}

// RecordDownload records a download event
func (s *statisticsService) RecordDownload(
	fileID uuid.UUID,
	downloaderID *uuid.UUID,
	ipAddress, userAgent, referer *string,
	passwordProtected bool,
) error {
	// Verify file exists
	file, err := s.fileRepo.FindByID(fileID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}
	
	// Only track statistics for files with owners (authenticated uploads)
	if file.OwnerID == nil {
		// Anonymous file upload - don't track statistics
		return nil
	}
	
	// Determine access method
	accessMethod := models.AccessMethodDirect
	if passwordProtected {
		accessMethod = models.AccessMethodSharedLink
	}
	
	// Record download
	params := &models.DownloadHistoryParams{
		FileID:            fileID,
		DownloaderID:      downloaderID,
		IPAddress:         ipAddress,
		UserAgent:         userAgent,
		Referer:           referer,
		DownloadCompleted: true,
		AccessMethod:      accessMethod,
		PasswordProtected: passwordProtected,
	}
	
	if err := s.statsRepo.RecordDownload(params); err != nil {
		return fmt.Errorf("failed to record download: %w", err)
	}
	
	return nil
}

// RecordView records a file view (info page view, not download)
func (s *statisticsService) RecordView(fileID uuid.UUID) error {
	// Verify file exists and has owner
	file, err := s.fileRepo.FindByID(fileID)
	if err != nil {
		return fmt.Errorf("file not found: %w", err)
	}
	
	// Only track statistics for files with owners
	if file.OwnerID == nil {
		return nil
	}
	
	if err := s.statsRepo.IncrementViewCount(fileID); err != nil {
		return fmt.Errorf("failed to record view: %w", err)
	}
	
	return nil
}

// Response types
type DownloadHistoryResponse struct {
	Downloads  []models.DownloadHistory `json:"downloads"`
	Total      int64                    `json:"total"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	TotalPages int                      `json:"total_pages"`
}

type UserDashboard struct {
	TotalFiles     int                                  `json:"total_files"`
	TotalDownloads int                                  `json:"total_downloads"`
	TotalViews     int                                  `json:"total_views"`
	TopFiles       []repositories.FileStatsWithFile     `json:"top_files"`
}


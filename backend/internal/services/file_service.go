package services

import (
	"context"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/repositories"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/storage"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var _ repositories.FileRepository = (*FileService)(nil)

type FileService struct {
	db      *gorm.DB
	storage storage.Storage
}

func NewFileService(db *gorm.DB, st storage.Storage) *FileService {
	return &FileService{
		db:      db,
		storage: st,
	}
}

func (s *FileService) GetSystemPolicy(ctx context.Context) (*models.SystemPolicy, error) {
	var policy models.SystemPolicy
	if err := s.db.WithContext(ctx).First(&policy, 1).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return &models.SystemPolicy{
				ID:                       1,
				MaxFileSizeMB:            50,
				MinValidityHours:         1,
				MaxValidityDays:          30,
				DefaultValidityDays:      7,
				RequirePasswordMinLength: 8,
			}, nil
		}
		return nil, err
	}

	if policy.MaxFileSizeMB <= 0 {
		policy.MaxFileSizeMB = 50
	}
	if policy.MinValidityHours <= 0 {
		policy.MinValidityHours = 1
	}
	if policy.MaxValidityDays <= 0 {
		policy.MaxValidityDays = 30
	}
	if policy.DefaultValidityDays <= 0 {
		policy.DefaultValidityDays = 7
	}
	if policy.RequirePasswordMinLength < 4 {
		policy.RequirePasswordMinLength = 8
	}

	return &policy, nil
}

type UploadInput struct {
	FileName         string
	ContentType      string
	Size             int64
	Reader           io.Reader
	IsPublic         *bool
	OwnerID          *uuid.UUID
	PasswordHash     *string
	AvailableFrom    *time.Time
	AvailableTo      *time.Time
	SharedWithEmails []string
}

func (in *UploadInput) container() storage.ContainerType {
	if in != nil && in.IsPublic != nil && *in.IsPublic {
		return storage.ContainerPublic
	}
	return storage.ContainerPrivate
}

func (in *UploadInput) sanitizedFileName() string {
	if in == nil || in.FileName == "" {
		return uuid.NewString()
	}
	name := filepath.Base(in.FileName)
	name = strings.TrimSpace(name)
	if name == "" || name == "." {
		return uuid.NewString()
	}
	return name
}

func (s *FileService) UploadFile(ctx context.Context, input *UploadInput) (*models.File, error) {
	if input == nil || input.Reader == nil {
		return nil, fmt.Errorf("file service: invalid upload input")
	}
	if s.storage == nil {
		return nil, fmt.Errorf("file service: storage backend is not configured")
	}
	isPrivate := input.IsPublic != nil && !*input.IsPublic
	if isPrivate && input.OwnerID == nil {
		return nil, fmt.Errorf("anonymous private uploads require authentication")
	}

	fileName := input.sanitizedFileName()
	storageName := fmt.Sprintf("%s-%s", uuid.NewString(), fileName)
	obj := &storage.Object{
		Name:        storageName,
		Container:   input.container(),
		ContentType: input.ContentType,
		Size:        input.Size,
		Reader:      input.Reader,
	}

	loc, err := s.storage.Upload(ctx, obj)
	if err != nil {
		return nil, err
	}

	availableFrom, availableTo, err := s.resolveAvailability(ctx, input)
	if err != nil {
		_ = s.storage.Delete(ctx, loc)
		return nil, err
	}

	file := &models.File{
		FileName:      fileName,
		FilePath:      loc.Path,
		FileSize:      input.Size,
		MimeType:      optionalString(input.ContentType),
		OwnerID:       input.OwnerID,
		IsPublic:      input.IsPublic,
		PasswordHash:  input.PasswordHash,
		AvailableFrom: availableFrom,
		AvailableTo:   availableTo,
	}

	// Normalize sharedWith emails and exclude owner's email
	var sharedWithEmails []string
	if len(input.SharedWithEmails) > 0 && input.OwnerID != nil {
		// Get owner email to exclude from sharedWith
		var ownerEmail string
		if input.OwnerID != nil {
			var owner models.User
			if err := s.db.WithContext(ctx).First(&owner, "id = ?", *input.OwnerID).Error; err == nil {
				ownerEmail = owner.Email
			}
		}

		// Normalize and filter emails
		for _, email := range input.SharedWithEmails {
			email = strings.TrimSpace(email)
			if email != "" {
				// Skip owner's email (owner has access via owner_id)
				if ownerEmail == "" || !strings.EqualFold(email, ownerEmail) {
					sharedWithEmails = append(sharedWithEmails, email)
				}
			}
		}
	}

	// Set sharedWithEmails directly to file (JSONB column)
	file.SharedWithEmails = models.StringArray(sharedWithEmails)

	txErr := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create file with SharedWithEmails (JSONB column)
		if err := tx.Omit("Owner", "Statistics").Create(file).Error; err != nil {
			return err
		}

		// Preload Owner
		return tx.Preload("Owner").First(file, "id = ?", file.ID).Error
	})
	if txErr != nil {
		_ = s.storage.Delete(ctx, loc)
		return nil, txErr
	}

	// Reload file with Owner
	if err := s.db.Preload("Owner").First(file, "id = ?", file.ID).Error; err != nil {
		return nil, err
	}

	return file, nil
}

func optionalString(val string) *string {
	if val == "" {
		return nil
	}
	v := val
	return &v
}

func containerFromFile(file *models.File) storage.ContainerType {
	if file != nil && file.IsPublic != nil && *file.IsPublic {
		return storage.ContainerPublic
	}
	return storage.ContainerPrivate
}

func (s *FileService) resolveAvailability(ctx context.Context, input *UploadInput) (*time.Time, *time.Time, error) {
	if input.AvailableFrom != nil || input.AvailableTo != nil {
		return input.AvailableFrom, input.AvailableTo, nil
	}

	days, err := s.defaultValidityDays(ctx)
	if err != nil {
		return nil, nil, err
	}
	if days <= 0 {
		days = 7
	}

	now := time.Now()
	expiryTime := now.AddDate(0, 0, days)

	return &now, &expiryTime, nil
}

func (s *FileService) defaultValidityDays(ctx context.Context) (int, error) {
	var policy models.SystemPolicy
	if err := s.db.WithContext(ctx).First(&policy).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 7, nil
		}
		return 0, err
	}

	if policy.DefaultValidityDays <= 0 {
		return 7, nil
	}
	return policy.DefaultValidityDays, nil
}

func (s *FileService) GetByID(id uuid.UUID) (*models.File, error) {
	var file models.File
	err := s.db.Preload("Owner").Preload("Statistics").First(&file, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *FileService) GetByShareToken(token string) (*models.File, error) {
	var file models.File
	err := s.db.Preload("Owner").Preload("Statistics").Where("share_token = ?", token).First(&file).Error
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (s *FileService) Create(file *models.File) error {
	return s.db.Create(file).Error
}

func (s *FileService) Update(file *models.File) error {
	return s.db.Save(file).Error
}

func (s *FileService) Delete(id uuid.UUID) error {
	var file models.File
	if err := s.db.First(&file, "id = ?", id).Error; err != nil {
		return err
	}

	if s.storage != nil && file.FilePath != "" {
		loc := &storage.Location{
			Container: containerFromFile(&file),
			Path:      file.FilePath,
		}
		_ = s.storage.Delete(context.Background(), loc)
	}

	return s.db.Delete(&models.File{}, "id = ?", id).Error
}

func (s *FileService) GetByOwnerID(ownerID uuid.UUID, limit, offset int) ([]models.File, int64, error) {
	var files []models.File
	var total int64

	query := s.db.Model(&models.File{}).Where("owner_id = ?", ownerID)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Owner").Preload("Statistics").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&files).Error

	if err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

func (s *FileService) GetPublicFiles(limit, offset int) ([]models.File, int64, error) {
	var files []models.File
	var total int64

	query := s.db.Model(&models.File{}).Where("is_public = ?", true)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Owner").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&files).Error

	if err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

func (s *FileService) SearchFiles(query string, limit, offset int) ([]models.File, int64, error) {
	var files []models.File
	var total int64

	searchQuery := s.db.Model(&models.File{}).
		Where("file_name ILIKE ?", "%"+query+"%")

	if err := searchQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := searchQuery.Preload("Owner").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&files).Error

	if err != nil {
		return nil, 0, err
	}

	return files, total, nil
}

func (s *FileService) GetExpiredFiles() ([]models.File, error) {
	var files []models.File

	err := s.db.Where("available_to IS NOT NULL AND available_to < CURRENT_TIMESTAMP").
		Find(&files).Error

	if err != nil {
		return nil, err
	}

	return files, nil
}

func (s *FileService) GetPendingFiles() ([]models.File, error) {
	var files []models.File

	err := s.db.Where("available_from IS NOT NULL AND available_from > CURRENT_TIMESTAMP").
		Find(&files).Error

	if err != nil {
		return nil, err
	}

	return files, nil
}

func (s *FileService) Download(ctx context.Context, filePath *string, container storage.ContainerType) (*storage.DownloadResult, error) {
	if filePath == nil || *filePath == "" {
		return nil, fmt.Errorf("file service: file path is empty")
	}

	if s.storage == nil {
		return nil, fmt.Errorf("file service: storage backend is not configured")
	}

	loc := &storage.Location{
		Container: container,
		Path:      *filePath,
	}

	return s.storage.Download(ctx, loc)
}

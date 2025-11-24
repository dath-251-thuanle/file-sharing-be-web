package services

import (
	"context"
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

type UploadInput struct {
	FileName      string
	ContentType   string
	Size          int64
	Reader        io.Reader
	IsPublic      *bool
	OwnerID       *uuid.UUID
	PasswordHash  *string
	AvailableFrom *time.Time
	AvailableTo   *time.Time
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

	// Ensure at least one date is set (constraint requirement)
	availableFrom := input.AvailableFrom
	availableTo := input.AvailableTo
	
	if availableFrom == nil && availableTo == nil {
		// Default: available from now, expire after 7 days
		now := time.Now()
		defaultDays := 7 // TODO: fetch from system_policy.default_validity_days
		expiryTime := now.AddDate(0, 0, defaultDays)
		availableFrom = &now
		availableTo = &expiryTime
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

	if err := s.db.WithContext(ctx).Create(file).Error; err != nil {
		_ = s.storage.Delete(ctx, loc)
		return nil, err
	}

	// Reload to get actual values from database (including defaults)
	if err := s.db.WithContext(ctx).First(file, "id = ?", file.ID).Error; err != nil {
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
	err := s.db.Preload("Owner").Preload("Statistics").
		Where("share_token = ?", token).First(&file).Error
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
		// Best-effort delete to avoid leaving orphaned blobs. Ignore errors and continue.
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

	err := query.Preload("Statistics").
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

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

func (s *FileService) UploadFile(ctx context.Context, input *UploadInput) (*models.File, error) {}

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

func (s *FileService) GetByID(id uuid.UUID) (*models.File, error) {}

func (s *FileService) GetByShareToken(token string) (*models.File, error) {}

func (s *FileService) Create(file *models.File) error {}

func (s *FileService) Update(file *models.File) error {}

func (s *FileService) Delete(id uuid.UUID) error {}

func (s *FileService) GetByOwnerID(ownerID uuid.UUID, limit, offset int) ([]models.File, int64, error) {}

func (s *FileService) GetPublicFiles(limit, offset int) ([]models.File, int64, error) {}

func (s *FileService) SearchFiles(query string, limit, offset int) ([]models.File, int64, error) {}

func (s *FileService) GetExpiredFiles() ([]models.File, error) {}

func (s *FileService) GetPendingFiles() ([]models.File, error) {}

func (s *FileService) Download(ctx context.Context, filePath *string, container storage.ContainerType) (*storage.DownloadResult, error) {}

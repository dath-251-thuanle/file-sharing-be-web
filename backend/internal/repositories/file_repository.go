package repositories

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/google/uuid"
)

type FileRepository interface {
	GetByID(id uuid.UUID) (*models.File, error)
	GetByShareToken(token string) (*models.File, error)
	Create(file *models.File) error
	Update(file *models.File) error
	Delete(id uuid.UUID) error
	GetByOwnerID(ownerID uuid.UUID, limit, offset int) ([]models.File, int64, error)
	GetPublicFiles(limit, offset int) ([]models.File, int64, error)
	SearchFiles(query string, limit, offset int) ([]models.File, int64, error)
	GetExpiredFiles() ([]models.File, error)
	GetPendingFiles() ([]models.File, error)
}


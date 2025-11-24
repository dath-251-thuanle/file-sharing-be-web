package repositories

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/google/uuid"
)

type UserRepository interface {
	GetByID(id uuid.UUID) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
	Delete(id uuid.UUID) error
	GetAll(limit, offset int) ([]models.User, int64, error)
	ExistsByUsername(username string) (bool, error)
	ExistsByEmail(email string) (bool, error)
}


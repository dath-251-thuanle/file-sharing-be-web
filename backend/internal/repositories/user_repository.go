package repositories

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
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

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByID(id uuid.UUID) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "username = ?", username).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, "email = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

func (r *userRepository) Delete(id uuid.UUID) error {
	return r.db.Delete(&models.User{}, "id = ?", id).Error
}

func (r *userRepository) GetAll(limit, offset int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	r.db.Model(&models.User{}).Count(&total)
	err := r.db.Limit(limit).Offset(offset).Find(&users).Error

	return users, total, err
}

func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

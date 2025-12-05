package repositories

import (
	"time"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LoginSessionRepository interface {
	Create(userID uuid.UUID, ttl time.Duration) (*models.LoginSession, error)
	GetActiveByID(id uuid.UUID, now time.Time) (*models.LoginSession, error)
	IncrementFailedAttempts(id uuid.UUID) error
	MarkConsumed(id uuid.UUID, consumedAt time.Time) error
}

type loginSessionRepository struct {
	db *gorm.DB
}

func NewLoginSessionRepository(db *gorm.DB) LoginSessionRepository {
	return &loginSessionRepository{db: db}
}

func (r *loginSessionRepository) Create(userID uuid.UUID, ttl time.Duration) (*models.LoginSession, error) {
	now := time.Now().UTC()
	session := &models.LoginSession{
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(ttl),
	}
	if err := r.db.Create(session).Error; err != nil {
		return nil, err
	}
	return session, nil
}

func (r *loginSessionRepository) GetActiveByID(id uuid.UUID, now time.Time) (*models.LoginSession, error) {
	var session models.LoginSession
	err := r.db.
		Where("id = ? AND consumed_at IS NULL AND expires_at > ?", id, now).
		First(&session).Error
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *loginSessionRepository) IncrementFailedAttempts(id uuid.UUID) error {
	return r.db.Model(&models.LoginSession{}).
		Where("id = ?", id).
		UpdateColumn("failed_attempts", gorm.Expr("failed_attempts + 1")).Error
}

func (r *loginSessionRepository) MarkConsumed(id uuid.UUID, consumedAt time.Time) error {
	return r.db.Model(&models.LoginSession{}).
		Where("id = ?", id).
		Update("consumed_at", consumedAt).Error
}



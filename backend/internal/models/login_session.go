package models

import (
	"time"

	"github.com/google/uuid"
)

type LoginSession struct {
	ID             uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID         uuid.UUID  `gorm:"type:uuid;not null;index"`
	CreatedAt      time.Time  `gorm:"type:timestamptz;not null;default:now()"`
	ExpiresAt      time.Time  `gorm:"type:timestamptz;not null;index"`
	ConsumedAt     *time.Time `gorm:"type:timestamptz"`
	FailedAttempts int        `gorm:"not null;default:0"`
}



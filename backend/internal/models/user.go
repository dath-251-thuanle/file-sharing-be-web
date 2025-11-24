package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Username     string    `gorm:"type:citext;uniqueIndex;not null" json:"username"`
	Email        string    `gorm:"type:citext;uniqueIndex;not null" json:"email"`
	Role         UserRole  `gorm:"type:user_role;default:'user'" json:"role"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	TOTPSecret   *string   `gorm:"type:varchar(32)" json:"-"`
	TOTPEnabled  *bool     `gorm:"default:false" json:"totp_enabled"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	// Relationships
	OwnedFiles  []File       `gorm:"foreignKey:OwnerID" json:"-"`
	SharedFiles []SharedWith `gorm:"foreignKey:UserID" json:"-"`
}

func (User) TableName() string {
	return "users"
}

// BeforeCreate hook to generate UUID
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}


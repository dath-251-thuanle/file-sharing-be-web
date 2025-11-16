package models

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SharedWith struct {
	ID     uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	FileID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_file_user" json:"file_id"`
	UserID uuid.UUID `gorm:"type:uuid;not null;index;uniqueIndex:idx_file_user" json:"user_id"`

	// Relationships
	File File `gorm:"foreignKey:FileID" json:"file,omitempty"`
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (SharedWith) TableName() string {
	return "shared_with"
}

// BeforeCreate hook to generate UUID
func (s *SharedWith) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}


package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type File struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	ShareToken    string     `gorm:"type:varchar(32);uniqueIndex;not null" json:"share_token"`
	FileName      string     `gorm:"type:varchar(255);not null" json:"file_name"`
	FilePath      string     `gorm:"type:varchar(512);not null" json:"file_path"`
	FileSize      int64      `gorm:"type:bigint;not null" json:"file_size"`
	MimeType      *string    `gorm:"type:varchar(100)" json:"mime_type"`
	OwnerID       *uuid.UUID `gorm:"type:uuid;index" json:"owner_id"`
	IsPublic      bool       `gorm:"default:true" json:"is_public"`
	PasswordHash  *string    `gorm:"type:varchar(255)" json:"-"`
	AvailableFrom *time.Time `gorm:"type:timestamp with time zone" json:"available_from"`
	AvailableTo   *time.Time `gorm:"type:timestamp with time zone" json:"available_to"`
	CreatedAt     time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	// Relationships
	Owner      *User           `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	SharedWith []SharedWith    `gorm:"foreignKey:FileID" json:"shared_with,omitempty"`
	Statistics *FileStatistics `gorm:"foreignKey:FileID" json:"statistics,omitempty"`
}

func (File) TableName() string {
	return "files"
}

// BeforeCreate hook to generate UUID and share token
func (f *File) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	if f.ShareToken == "" {
		f.ShareToken = generateShareToken()
	}
	return nil
}

// AfterCreate hook to initialize statistics for files with owner
func (f *File) AfterCreate(tx *gorm.DB) error {
	// Only create statistics for files uploaded by authenticated users
	if f.OwnerID != nil {
		return InitializeForFile(tx, f.ID)
	}
	return nil
}

func (f *File) IsAvailable() bool {
	now := time.Now()

	if f.AvailableFrom != nil && now.Before(*f.AvailableFrom) {
		return false // Not yet available
	}

	if f.AvailableTo != nil && now.After(*f.AvailableTo) {
		return false // Expired
	}

	return true
}
// GetStatus returns the current status of the file
func (f *File) GetStatus() string {
	if !f.IsAvailable() {
		if f.AvailableFrom != nil && time.Now().Before(*f.AvailableFrom) {
			return "pending"
		}
		if f.AvailableTo != nil && time.Now().After(*f.AvailableTo) {
			return "expired"
		}
	}
	return "active"
}

func (f *File) HasStatistics() bool {
	return f.OwnerID != nil
}

func generateShareToken() string {
	return uuid.New().String()[:32]
}


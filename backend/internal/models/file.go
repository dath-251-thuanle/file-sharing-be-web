package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StringArray is a custom type for PostgreSQL JSONB array of strings
type StringArray []string

// Value implements driver.Valuer interface for saving to database
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return "[]", nil
	}
	return json.Marshal(a)
}

// Scan implements sql.Scanner interface for reading from database
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = []string{}
		return nil
	}

	var bytes []byte
	switch v := value.(type) {
	case []byte:
		bytes = v
	case string:
		bytes = []byte(v)
	default:
		return errors.New("failed to unmarshal JSONB value")
	}

	if len(bytes) == 0 || string(bytes) == "null" {
		*a = []string{}
		return nil
	}

	return json.Unmarshal(bytes, a)
}

type File struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	ShareToken    string     `gorm:"type:varchar(32);uniqueIndex;not null" json:"share_token"`
	FileName      string     `gorm:"type:varchar(255);not null" json:"file_name"`
	FilePath      string     `gorm:"type:varchar(512);not null" json:"file_path"`
	FileSize      int64      `gorm:"type:bigint;not null" json:"file_size"`
	MimeType      *string    `gorm:"type:varchar(100)" json:"mime_type"`
	OwnerID       *uuid.UUID `gorm:"type:uuid;index" json:"owner_id"`
	IsPublic      *bool       `gorm:"default:true" json:"is_public"`
	PasswordHash  *string    `gorm:"type:varchar(255)" json:"-"`
	AvailableFrom    *time.Time `gorm:"type:timestamp with time zone" json:"available_from"`
	AvailableTo      *time.Time `gorm:"type:timestamp with time zone" json:"available_to"`
	SharedWithEmails StringArray `gorm:"type:jsonb;default:'[]'" json:"shared_with,omitempty"`  // Multi-valued attribute (whitelist)
	CreatedAt        time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`

	Owner      *User           `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Statistics *FileStatistics `gorm:"foreignKey:FileID" json:"statistics,omitempty"`
}

func (File) TableName() string {
	return "files"
}

func (f *File) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	if f.ShareToken == "" {
		f.ShareToken = generateShareToken()
	}
	return nil
}

func (f *File) AfterCreate(tx *gorm.DB) error {
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

// HasPassword checks if the file has a password protection
func (f *File) HasPassword() bool {
	return f.PasswordHash != nil && *f.PasswordHash != ""
}

func generateShareToken() string {
	return uuid.New().String()[:32]
}


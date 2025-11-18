package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileStatistics struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	FileID            uuid.UUID  `gorm:"type:uuid;uniqueIndex;not null" json:"file_id"`
	DownloadCount     int        `gorm:"default:0" json:"download_count"`
	UniqueDownloaders int        `gorm:"default:0" json:"unique_downloaders"`
	LastDownloadedAt  *time.Time `gorm:"type:timestamp with time zone" json:"last_downloaded_at"`
	CreatedAt         time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt         time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	File File `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"file,omitempty"`
}

func (FileStatistics) TableName() string {
	return "file_statistics"
}

func (fs *FileStatistics) BeforeCreate(tx *gorm.DB) error {
	if fs.ID == uuid.Nil {
		fs.ID = uuid.New()
	}
	return nil
}

func InitializeForFile(tx *gorm.DB, fileID uuid.UUID) error {
	stats := &FileStatistics{
		FileID:        fileID,
		DownloadCount: 0,
	}
	return tx.Create(stats).Error
}


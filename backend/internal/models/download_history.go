package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type DownloadHistory struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	FileID            uuid.UUID  `gorm:"type:uuid;not null;index" json:"file_id"`
	DownloaderID      *uuid.UUID `gorm:"type:uuid;index" json:"downloader_id,omitempty"`
	DownloadedAt      time.Time  `gorm:"default:CURRENT_TIMESTAMP;index" json:"downloaded_at"`
	DownloadCompleted *bool      `gorm:"default:true" json:"download_completed"`

	File       File  `gorm:"foreignKey:FileID;constraint:OnDelete:CASCADE" json:"file,omitempty"`
	Downloader *User `gorm:"foreignKey:DownloaderID;constraint:OnDelete:SET NULL" json:"downloader,omitempty"`
}

func (DownloadHistory) TableName() string {
	return "download_history"
}

func (dh *DownloadHistory) BeforeCreate(tx *gorm.DB) error {
	if dh.ID == uuid.Nil {
		dh.ID = uuid.New()
	}
	return nil
}

func RecordDownload(tx *gorm.DB, fileID uuid.UUID, downloaderID *uuid.UUID, completed bool) error {
	history := &DownloadHistory{
		FileID:            fileID,
		DownloaderID:      downloaderID,
		DownloadCompleted: &completed,
	}
	return tx.Create(history).Error
}


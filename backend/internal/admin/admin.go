package admin

import (
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/storage"
)

// --- MODELS ---

type SystemPolicy struct {
	ID                       uint `json:"id" gorm:"primaryKey"`
	MaxFileSizeMB            int  `json:"maxFileSizeMB"`
	MinValidityHours         int  `json:"minValidityHours"`
	MaxValidityDays          int  `json:"maxValidityDays"`
	DefaultValidityDays      int  `json:"defaultValidityDays"`
	RequirePasswordMinLength int  `json:"requirePasswordMinLength"`
}

func (SystemPolicy) TableName() string { return "system_policy" }

// File (Partial struct for cleanup)
// We need IsPublic to determine the ContainerType
type File struct {
	ID          string    `gorm:"primaryKey"`
	FilePath    string    `gorm:"column:file_path"`
	IsPublic    bool      `gorm:"column:is_public"` // Added this field
	AvailableTo time.Time `gorm:"column:available_to"`
}

func (File) TableName() string { return "files" }

// --- SETUP ---

func Setup(router *gin.Engine, db *gorm.DB, store storage.Storage) {
	ensurePolicyExists(db)

	group := router.Group("/api/admin")
	{
		group.GET("/policy", getPolicy(db))
		group.PATCH("/policy", updatePolicy(db))
		group.POST("/cleanup", cleanupFiles(db, store))
	}
}

// --- LOGIC ---

func ensurePolicyExists(db *gorm.DB) {
	var count int64
	db.Model(&SystemPolicy{}).Where("id = ?", 1).Count(&count)
	if count == 0 {
		defaultPolicy := SystemPolicy{
			ID:                       1,
			MaxFileSizeMB:            50,
			MinValidityHours:         1,
			MaxValidityDays:          30,
			DefaultValidityDays:      7,
			RequirePasswordMinLength: 8,
		}
		db.Create(&defaultPolicy)
	}
}

func getPolicy(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var policy SystemPolicy
		if err := db.First(&policy, 1).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Policy not found"})
			return
		}
		c.JSON(http.StatusOK, policy)
	}
}

func updatePolicy(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input SystemPolicy
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}
		
		if err := db.Model(&SystemPolicy{ID: 1}).Updates(input).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
			return
		}
		
		var updated SystemPolicy
		db.First(&updated, 1)
		c.JSON(http.StatusOK, gin.H{"message": "Updated", "policy": updated})
	}
}

func cleanupFiles(db *gorm.DB, store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := c.GetHeader("X-Cron-Secret")
		envSecret := os.Getenv("CLEANUP_SECRET")
		if secret == "" || secret != envSecret {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid or missing X-Cron-Secret"})
			return
		}

		var expiredFiles []File
		// Find files where available_to < NOW()
		if err := db.Where("available_to < ?", time.Now()).Find(&expiredFiles).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		if len(expiredFiles) == 0 {
			c.JSON(http.StatusOK, gin.H{"message": "No files to clean"})
			return
		}

		deletedCount := 0
		for _, file := range expiredFiles {
			// 1. Determine Container
			container := storage.ContainerPrivate
			if file.IsPublic {
				container = storage.ContainerPublic
			}

			// 2. Construct Location object
			location := &storage.Location{
				Container: container,
				Path:      file.FilePath,
			}
			
			// 3. Delete from Storage (Pass Context)
			// We ignore error if file is already missing (idempotent)
			err := store.Delete(c.Request.Context(), location)
			
			// 4. Delete from DB if storage delete worked or file was missing
			if err == nil {
				if err := db.Delete(&file).Error; err == nil {
					deletedCount++
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Cleanup complete",
			"files_found":  len(expiredFiles),
			"files_deleted": deletedCount,
		})
	}
}

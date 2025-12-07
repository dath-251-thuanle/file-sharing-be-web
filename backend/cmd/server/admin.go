package admin

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/storage"
)

// Rate limiting state
var (
	cleanupRateMutex            sync.Mutex
	cleanupWindowStart          time.Time
	cleanupRequestCount         int
	cleanupWindowDuration       = time.Minute
	cleanupMaxRequestsPerWindow = 10
)

//########################
//## 0. SETUP MODULE   ###
//########################
func Setup(router *gin.Engine, db *gorm.DB, store storage.Storage) {
	// 1. Ensure DB has default policy
	ensure_policy_exists(db)

	// 2. Register Routes with Auth Middleware
	admin := router.Group("/admin")
	admin.Use(admin_auth_middleware())
	{
		admin.GET("/policy", get_policy(db))
		admin.PATCH("/policy", update_policy(db))
		admin.POST("/cleanup", cleanup_files(db, store))
	}
}

//########################
//## 1. AUTH MIDDLEWARE ##
//########################
func admin_auth_middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Let OPTIONS through without auth (CORS preflight)
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		envToken := os.Getenv("ADMIN_API_TOKEN")
		if envToken == "" {
			log.Println("[Admin] ❌ ADMIN_API_TOKEN is not set in .env. Admin API is disabled.")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin API configuration missing"})
			return
		}

		authHeader := c.GetHeader("Authorization")
		
		if strings.HasPrefix(path, "/admin/cleanup") {
			cronSecretHeader := c.GetHeader("X-Cron-Secret")
			envSecret := os.Getenv("CLEANUP_SECRET")
			if cronSecretHeader != "" && envSecret != "" && cronSecretHeader == envSecret {
				c.Set("adminAuth", "cron")
				c.Next()
				return
			}
		}

		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Bearer token is required",
			})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid token format. Use 'Bearer <token>'",
			})
			return
		}

		if parts[1] != envToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid Admin Token",
			})
			return
		}

		c.Set("adminAuth", "token")
		c.Next()
	}
}

//########################
//## 2. GET POLICY     ###
//########################
func get_policy(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var policy models.SystemPolicy
		if err := db.First(&policy, 1).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal error",
				"message": "Policy not found",
			})
			return
		}
		c.JSON(http.StatusOK, policy)
	}
}

//########################
//## 3. UPDATE POLICY  ###
//########################
func update_policy(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		type policyUpdateRequest struct {
			MaxFileSizeMB            *int `json:"maxFileSizeMB"`
			MinValidityHours         *int `json:"minValidityHours"`
			MaxValidityDays          *int `json:"maxValidityDays"`
			DefaultValidityDays      *int `json:"defaultValidityDays"`
			RequirePasswordMinLength *int `json:"requirePasswordMinLength"`
		}

		var input policyUpdateRequest
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"message": "Invalid input data",
			})
			return
		}

		// Validation logic preserved...
		if input.MaxFileSizeMB != nil && *input.MaxFileSizeMB < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error", "message": "maxFileSizeMB must be >= 1"})
			return
		}

		updates := map[string]interface{}{}
		if input.MaxFileSizeMB != nil { updates["max_file_size_mb"] = *input.MaxFileSizeMB }
		if input.MinValidityHours != nil { updates["min_validity_hours"] = *input.MinValidityHours }
		if input.MaxValidityDays != nil { updates["max_validity_days"] = *input.MaxValidityDays }
		if input.DefaultValidityDays != nil { updates["default_validity_days"] = *input.DefaultValidityDays }
		if input.RequirePasswordMinLength != nil { updates["require_password_min_length"] = *input.RequirePasswordMinLength }

		if len(updates) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Validation error", "message": "No fields provided"})
			return
		}

		if err := db.Model(&models.SystemPolicy{ID: 1}).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error", "message": "Update failed"})
			return
		}

		var updated models.SystemPolicy
		db.First(&updated, 1)
		c.JSON(http.StatusOK, gin.H{"message": "Policy updated", "policy": updated})
	}
}

//########################
//## 4. CLEANUP FILES  ###
//########################
func cleanup_files(db *gorm.DB, store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !allowCleanupRequest() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Cleanup endpoint is rate limited.",
			})
			return
		}

		startTime := time.Now().UTC()
		var expiredFiles []models.File
		
		// Find expired files
		if err := db.Where("available_to < ?", time.Now()).Find(&expiredFiles).Error; err != nil {
			log.Printf("[Admin] error querying expired files: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error", "message": "Database query failed"})
			return
		}

		deletedCount := 0
		for _, file := range expiredFiles {
			// Determine container
			container := storage.ContainerPrivate
			if file.IsPublic != nil && *file.IsPublic {
				container = storage.ContainerPublic
			}

			// Delete from Storage
			loc := &storage.Location{Container: container, Path: file.FilePath}
			err := store.Delete(c.Request.Context(), loc)

			// Delete from DB if storage delete OK or file missing
			if err == nil {
				if err := db.Delete(&file).Error; err == nil {
					deletedCount++
				}
			} else {
				log.Printf("[Admin] failed to delete storage file %s: %v", file.ID, err)
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":      "Cleanup complete",
			"files_found":  len(expiredFiles),
			"files_deleted": deletedCount,
			"timestamp":    startTime.Format(time.RFC3339),
		})
	}
}

//########################
//## 5. INIT HELPERS   ###
//########################
func ensure_policy_exists(db *gorm.DB) {
	var count int64
	db.Model(&models.SystemPolicy{}).Where("id = ?", 1).Count(&count)
	if count == 0 {
		log.Println("[Admin] Initializing Default System Policy...")
		defaultPolicy := models.SystemPolicy{
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

func allowCleanupRequest() bool {
	cleanupRateMutex.Lock()
	defer cleanupRateMutex.Unlock()

	now := time.Now()
	if cleanupWindowStart.IsZero() || now.Sub(cleanupWindowStart) > cleanupWindowDuration {
		cleanupWindowStart = now
		cleanupRequestCount = 0
	}

	if cleanupRequestCount >= cleanupMaxRequestsPerWindow {
		return false
	}
	cleanupRequestCount++
	return true
}

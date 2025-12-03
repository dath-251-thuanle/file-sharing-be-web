package admin

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
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

// Global state for the rotating token
var (
	current_admin_token string
	token_mutex         sync.RWMutex
)

//########################
//## 0. SETUP MODULE   ###
//########################
func Setup(router *gin.Engine, db *gorm.DB, store storage.Storage) {
	// 1. Ensure DB has default policy
	ensure_policy_exists(db)

	// 2. Start the 5-minute token rotation loop
	go rotate_token_job()

	// 3. Register Routes with Auth Middleware
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
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing Authorization header"})
			return
		}

		// Format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token format. Use 'Bearer <token>'"})
			return
		}

		token_mutex.RLock()
		validToken := current_admin_token
		token_mutex.RUnlock()

		if parts[1] != validToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired admin token"})
			return
		}

		c.Next()
	}
}

//########################
//## 2. TOKEN ROTATOR  ###
//########################
func rotate_token_job() {
	// Generate immediately on startup
	generate_new_token()

	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		generate_new_token()
	}
}

func generate_new_token() {
	bytes := make([]byte, 16) // 16 bytes = 32 hex chars
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("[Admin] NOT OK! Failed to generate token: %v", err)
		return
	}
	newToken := hex.EncodeToString(bytes)

	token_mutex.Lock()
	current_admin_token = newToken
	token_mutex.Unlock()

	// PRINT TO CONSOLE so you can copy-paste it
	fmt.Println("\n=======================================================")
	fmt.Printf("[Admin] NEW ACCESS TOKEN (Valid 5 mins): %s\n", newToken)
	fmt.Println("=======================================================")
}

//########################
//## 3. GET POLICY     ###
//########################
func get_policy(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var policy models.SystemPolicy
		// ID 1 is the singleton configuration
		if err := db.First(&policy, 1).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Policy not found", "details": err.Error()})
			return
		}
		c.JSON(http.StatusOK, policy)
	}
}

//########################
//## 4. UPDATE POLICY  ###
//########################
func update_policy(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var input models.SystemPolicy
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// Update fields provided in JSON (ignoring zero values)
		if err := db.Model(&models.SystemPolicy{ID: 1}).Updates(input).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
			return
		}

		// Return fresh state
		var updated models.SystemPolicy
		db.First(&updated, 1)
		c.JSON(http.StatusOK, gin.H{"message": "Policy updated", "policy": updated})
	}
}

//########################
//## 5. CLEANUP FILES  ###
//########################
func cleanup_files(db *gorm.DB, store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Verify X-Cron-Secret (Second layer of security for automated jobs)
		secret := c.GetHeader("X-Cron-Secret")
		envSecret := os.Getenv("CLEANUP_SECRET")
		
		// If calling manually with Bearer token, we might skip this check,
		// but let's keep it strict: You need BOTH Bearer (Authentication) AND Secret (Authorization for this action)
		if secret == "" || secret != envSecret {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid or missing X-Cron-Secret"})
			return
		}

		var expiredFiles []models.File
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
			// Determine container (Public vs Private)
			container := storage.ContainerPrivate
			if file.IsPublic != nil && *file.IsPublic {
				container = storage.ContainerPublic
			}

			// Delete from Storage
			location := &storage.Location{
				Container: container,
				Path:      file.FilePath,
			}
			err := store.Delete(c.Request.Context(), location)

			// Delete from DB (Hard delete)
			if err == nil {
				if err := db.Delete(&file).Error; err == nil {
					deletedCount++
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":       "Cleanup complete",
			"files_found":   len(expiredFiles),
			"files_deleted": deletedCount,
		})
	}
}

//########################
//## 6. INIT HELPERS   ###
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

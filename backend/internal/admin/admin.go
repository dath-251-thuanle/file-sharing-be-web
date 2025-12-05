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
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		authHeader := c.GetHeader("Authorization")
		cronSecretHeader := c.GetHeader("X-Cron-Secret")
		envSecret := os.Getenv("CLEANUP_SECRET")

		// Special auth logic for /admin/cleanup:
		// - Allow either valid admin token OR valid X-Cron-Secret (docs: "hoặc").
		if strings.HasPrefix(path, "/admin/cleanup") {
			// 1. Try admin Bearer token path if provided
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" {
					token_mutex.RLock()
					validToken := current_admin_token
					token_mutex.RUnlock()

					if parts[1] == validToken {
						c.Set("adminAuth", "token")
						c.Next()
						return
					}
				}
				// Authorization header provided but invalid token
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":   "Forbidden",
					"message": "You don't have permission to perform cleanup",
				})
				return
			}

			// 2. Fallback: Cron secret header
			if cronSecretHeader != "" {
				if cronSecretHeader == envSecret && envSecret != "" {
					c.Set("adminAuth", "cron")
					c.Next()
					return
				}
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"error":   "Forbidden",
					"message": "Invalid cron secret",
				})
				return
			}

			// 3. Neither token nor secret header provided 
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "X-Cron-Secret header is required",
			})
			return
		}

		// Default admin auth for other /admin endpoints: require Bearer token
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Bearer token is required",
			})
			return
		}

		// Format: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid token format. Use 'Bearer <token>'",
			})
			return
		}

		token_mutex.RLock()
		validToken := current_admin_token
		token_mutex.RUnlock()

		if parts[1] != validToken {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid or expired admin token",
			})
			return
		}

		c.Set("adminAuth", "token")
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
//## 4. UPDATE POLICY  ###
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

		// Get current policy to validate relationships between fields if needed
		var current models.SystemPolicy
		if err := db.First(&current, 1).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal error",
				"message": "Policy not found",
			})
			return
		}

		// Minimum value according to SystemPolicyUpdate in docs
		if input.MaxFileSizeMB != nil && *input.MaxFileSizeMB < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"message": "maxFileSizeMB must be greater than or equal to 1",
			})
			return
		}
		if input.MinValidityHours != nil && *input.MinValidityHours < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"message": "minValidityHours must be greater than or equal to 1",
			})
			return
		}
		if input.MaxValidityDays != nil && *input.MaxValidityDays < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"message": "maxValidityDays must be greater than or equal to 1",
			})
			return
		}
		if input.DefaultValidityDays != nil && *input.DefaultValidityDays < 1 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"message": "defaultValidityDays must be greater than or equal to 1",
			})
			return
		}
		if input.RequirePasswordMinLength != nil && *input.RequirePasswordMinLength < 4 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"message": "requirePasswordMinLength must be greater than or equal to 4",
			})
			return
		}

		// Validate relationship maxValidityDays >= minValidityHours (according to docs)
		// Use new value if provided, otherwise use current
		minHours := current.MinValidityHours
		maxDays := current.MaxValidityDays
		if input.MinValidityHours != nil {
			minHours = *input.MinValidityHours
		}
		if input.MaxValidityDays != nil {
			maxDays = *input.MaxValidityDays
		}
		if maxDays < minHours {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"message": "maxValidityDays must be greater than or equal to minValidityHours",
			})
			return
		}

		// Build map only with fields sent (partial update)
		updates := map[string]interface{}{}
		if input.MaxFileSizeMB != nil {
			updates["max_file_size_mb"] = *input.MaxFileSizeMB
		}
		if input.MinValidityHours != nil {
			updates["min_validity_hours"] = *input.MinValidityHours
		}
		if input.MaxValidityDays != nil {
			updates["max_validity_days"] = *input.MaxValidityDays
		}
		if input.DefaultValidityDays != nil {
			updates["default_validity_days"] = *input.DefaultValidityDays
		}
		if input.RequirePasswordMinLength != nil {
			updates["require_password_min_length"] = *input.RequirePasswordMinLength
		}

		if len(updates) == 0 {
			// No fields provided for update
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"message": "No fields provided for update",
			})
			return
		}

		if err := db.Model(&models.SystemPolicy{ID: 1}).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal error",
				"message": "Update failed",
			})
			return
		}

		// Return fresh policy
		var updated models.SystemPolicy
		db.First(&updated, 1)
		c.JSON(http.StatusOK, gin.H{
			"message": "Policy updated",
			"policy":  updated,
		})
	}
}

//########################
//## 5. CLEANUP FILES  ###
//########################
func cleanup_files(db *gorm.DB, store storage.Storage) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Rate limiting: avoid DOS on /admin/cleanup endpoint
		if !allowCleanupRequest() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Too many requests",
				"message": "Cleanup endpoint is rate limited. Please try again later.",
			})
			return
		}

		startTime := time.Now().UTC()

		var expiredFiles []models.File
		if err := db.Where("available_to < ?", time.Now()).Find(&expiredFiles).Error; err != nil {
			log.Printf("[AdminCleanup] error querying expired files: %v (ip=%s)", err, c.ClientIP())
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal error",
				"message": "Database query failed",
			})
			return
		}

		deletedCount := 0
		for _, file := range expiredFiles {
			// Determine container (public vs private)
			container := storage.ContainerPrivate
			if file.IsPublic != nil && *file.IsPublic {
				container = storage.ContainerPublic
			}

			// Delete from storage
			location := &storage.Location{
				Container: container,
				Path:      file.FilePath,
			}
			err := store.Delete(c.Request.Context(), location)

			// Delete from DB (hard delete)
			if err == nil {
				if err := db.Delete(&file).Error; err == nil {
					deletedCount++
				}
			} else {
				log.Printf("[AdminCleanup] failed to delete file from storage (id=%s, path=%s): %v", file.ID, file.FilePath, err)
			}
		}

		timestamp := startTime.Format(time.RFC3339)

		// Audit log (timestamp, source, result)
		authType, _ := c.Get("adminAuth")
		log.Printf(
			"[AdminCleanup] ts=%s ip=%s auth=%v files_found=%d files_deleted=%d",
			timestamp,
			c.ClientIP(),
			authType,
			len(expiredFiles),
			deletedCount,
		)

		c.JSON(http.StatusOK, gin.H{
			"message":      "Expired files removed",
			"deletedFiles": deletedCount,
			"timestamp":    timestamp,
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

func allowCleanupRequest() bool {
	cleanupRateMutex.Lock()
	defer cleanupRateMutex.Unlock()

	now := time.Now()
	if cleanupWindowStart.IsZero() || now.Sub(cleanupWindowStart) > cleanupWindowDuration {
		// Reset window (time window for rate limiting)
		cleanupWindowStart = now
		cleanupRequestCount = 0
	}

	if cleanupRequestCount >= cleanupMaxRequestsPerWindow {
		return false
	}

	cleanupRequestCount++
	return true
}

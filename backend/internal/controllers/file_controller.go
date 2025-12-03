package controllers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/services"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileController struct {
	fileService    *services.FileService
	statsService   *services.StatisticsService
	historyService *services.DownloadHistoryService
}

func NewFileController(
	fileService *services.FileService,
	statsService *services.StatisticsService,
	historyService *services.DownloadHistoryService) *FileController {
	return &FileController{
		fileService:    fileService,
		statsService:   statsService,
		historyService: historyService,
	}
}

// UploadFile handles file upload
// POST /files/upload
func (fc *FileController) UploadFile(c *gin.Context) {
	// Get current user (optional - for authenticated uploads)
	currentUserID := getUserIDFromContext(c)

	// Parse file from form
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"message": "File is required",
		})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Cannot open uploaded file",
		})
		return
	}
	defer file.Close()

	// Parse form fields
	isPublicStr := c.PostForm("isPublic")
	var isPublic bool = true

	if isPublicStr != "" {
		isPublicStrLower := strings.ToLower(strings.TrimSpace(isPublicStr))
		isPublic = (isPublicStrLower == "true" || isPublicStrLower == "1" || isPublicStrLower == "yes")
	}
	password := c.PostForm("password")
	availableFromStr := c.PostForm("availableFrom")
	availableToStr := c.PostForm("availableTo")
	sharedWithEmails := c.PostFormArray("sharedWith")

	if currentUserID == nil {
		if !isPublic {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Private uploads (isPublic=false) require authentication",
			})
			return
		}
		if password != "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Password protection requires authentication",
			})
			return
		}
		if len(sharedWithEmails) > 0 {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Whitelist (sharedWith) requires authentication",
			})
			return
		}
	}

	if password != "" {
		policy, err := getSystemPolicy(c.Request.Context(), fc.fileService)
		if err == nil && len(password) < policy.RequirePasswordMinLength {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"message": fmt.Sprintf("Password must have at least %d characters", policy.RequirePasswordMinLength),
			})
			return
		}
		// Default minimum if policy not found
		if len(password) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"message": "Password must have at least 8 characters",
			})
			return
		}
	}

	// Hash password if provided
	var passwordHash *string
	if password != "" {
		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal server error",
				"message": "Failed to hash password",
			})
			return
		}
		hashStr := string(hash)
		passwordHash = &hashStr
	}

	// Parse availability dates
	var availableFrom, availableTo *time.Time
	if availableFromStr != "" {
		from, err := time.Parse(time.RFC3339, availableFromStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"message": "Invalid availableFrom format. Use RFC3339 format (e.g., 2025-11-10T00:00:00Z)",
			})
			return
		}
		availableFrom = &from
	}
	if availableToStr != "" {
		to, err := time.Parse(time.RFC3339, availableToStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"message": "Invalid availableTo format. Use RFC3339 format (e.g., 2025-11-17T00:00:00Z)",
			})
			return
		}
		availableTo = &to
	}

	// Validate availability dates
	if availableFrom != nil && availableTo != nil && !availableFrom.Before(*availableTo) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"message": "availableFrom must be before availableTo and within allowed policy window",
		})
		return
	}

	// Get Content-Type from header, or detect from file extension
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		// Try to detect from file extension
		ext := filepath.Ext(fileHeader.Filename)
		contentType = detectContentType(ext)
	}

	// Check file size against system policy
	policy, err := getSystemPolicy(c.Request.Context(), fc.fileService)
	if err == nil {
		maxSizeBytes := int64(policy.MaxFileSizeMB) * 1024 * 1024
		if fileHeader.Size > maxSizeBytes {
			c.JSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "Payload too large",
				"message": fmt.Sprintf("File size exceeds the system limit of %d MB", policy.MaxFileSizeMB),
			})
			return
		}
	}

	uploadInput := &services.UploadInput{
		FileName:      fileHeader.Filename,
		ContentType:   contentType,
		Size:          fileHeader.Size,
		Reader:        file,
		IsPublic:      &isPublic,
		OwnerID:       currentUserID,
		PasswordHash:  passwordHash,
		AvailableFrom: availableFrom,
		AvailableTo:   availableTo,
	}

	storedFile, err := fc.fileService.UploadFile(c.Request.Context(), uploadInput)
	if err != nil {
		// Check for specific error types
		if strings.Contains(err.Error(), "anonymous private uploads") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": err.Error(),
			})
			return
		}
		if strings.Contains(err.Error(), "validation") || strings.Contains(err.Error(), "invalid") {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Validation error",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": err.Error(),
		})
		return
	}

	response := gin.H{
		"success": true,
		"message": "File uploaded successfully",
		"file": gin.H{
			"id":         storedFile.ID,
			"fileName":   storedFile.FileName,
			"shareToken": storedFile.ShareToken,
			"isPublic":   storedFile.IsPublic,
		},
	}

	c.JSON(http.StatusCreated, response)
}

// GetFileInfo returns basic file metadata without downloading (public endpoint)
// GET /files/:shareToken
// Only returns minimal information: id, fileName, shareToken, status, isPublic, hasPassword
func (fc *FileController) GetFileInfo(c *gin.Context) {
	shareToken := c.Param("shareToken")

	// Get file metadata from database
	file, err := fc.fileService.GetByShareToken(shareToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not found",
				"message": "File not found",
			})
			return
		}
		return
	}

	// Check file status
	status := file.GetStatus()
	if status == "expired" {
		c.JSON(http.StatusGone, gin.H{
			"error":   "File expired",
			"message": "File has expired",
		})
		return
	}

	// Check if password hash exists and is not empty
	hasPassword := file.PasswordHash != nil && 
		*file.PasswordHash != "" && 
		len(*file.PasswordHash) >= 60 && 
		strings.HasPrefix(*file.PasswordHash, "$2")

	response := gin.H{
		"file": gin.H{
			"id":         file.ID,
			"fileName":   file.FileName,
			"shareToken": file.ShareToken,
			"status":     status,
			"isPublic":   file.IsPublic,
			"hasPassword": hasPassword,
		},
	}

	c.JSON(http.StatusOK, response)
}

// GetFileByID returns detailed file information by UUID (owner/admin only)
// GET /files/:id
func (fc *FileController) GetFileByID(c *gin.Context) {
	// CHECK 401: Kiểm tra đăng nhập
	currentUserID := getUserIDFromContext(c)
	if currentUserID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid or missing authentication token",
		})
		return
	}

	// CHECK 400: Validate Input - Must be UUID
	idStr := c.Param("id")
	fileID, err := uuid.Parse(idStr)
	if err != nil {
		// Not a valid UUID - this route should not match, but handle gracefully
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"message": "Invalid file ID format (Must be UUID)",
		})
		return
	}

	// CHECK 404: Tìm file
	file, err := fc.fileService.GetByID(fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not found",
				"message": "File not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to retrieve file",
		})
		return
	}

	// CHECK 403: Kiểm tra quyền truy cập (chỉ owner hoặc admin)
	currentUserRole := getUserRoleFromContext(c)
	isOwner := file.OwnerID != nil && *currentUserID == *file.OwnerID
	isAdmin := currentUserRole == models.RoleAdmin
	if !isOwner && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "You don't have permission to access this file",
		})
		return
	}

	// Build response with full metadata (including sensitive info)
	status := file.GetStatus()
	response := gin.H{
		"file": gin.H{
			"id":         file.ID,
			"fileName":   file.FileName,
			"fileSize":   file.FileSize,
			"mimeType":   file.MimeType,
			"shareToken": file.ShareToken,
			"status":     status,
			"isPublic":   file.IsPublic,
			"createdAt":  file.CreatedAt,
		},
	}

	// Add share link
	if file.ShareToken != "" {
		response["file"].(gin.H)["shareLink"] = fmt.Sprintf("/f/%s", file.ShareToken)
	}

	// Add availability info
	if file.AvailableFrom != nil {
		response["file"].(gin.H)["availableFrom"] = file.AvailableFrom
	}
	if file.AvailableTo != nil {
		response["file"].(gin.H)["availableTo"] = file.AvailableTo
		// Calculate hours remaining if active
		if status == "active" {
			hoursRemaining := time.Until(*file.AvailableTo).Hours()
			if hoursRemaining > 0 {
				response["file"].(gin.H)["hoursRemaining"] = hoursRemaining
			}
		}
	}

	// Add password protection indicator
	hasPassword := false
	if file.PasswordHash != nil {
		trimmed := strings.TrimSpace(*file.PasswordHash)
		hasPassword = trimmed != ""
	}
	response["file"].(gin.H)["hasPassword"] = hasPassword

	// Add sharedWith list (email addresses)
	if len(file.SharedWith) > 0 {
		sharedWithEmails := make([]string, 0, len(file.SharedWith))
		for _, sw := range file.SharedWith {
			if sw.User.Email != "" {
				sharedWithEmails = append(sharedWithEmails, sw.User.Email)
			}
		}
		response["file"].(gin.H)["sharedWith"] = sharedWithEmails
	}

	// Add owner info (full details)
	if file.Owner != nil {
		response["file"].(gin.H)["owner"] = gin.H{
			"id":       file.Owner.ID,
			"username": file.Owner.Username,
			"email":    file.Owner.Email,
			"role":     file.Owner.Role,
		}
	}

	c.JSON(http.StatusOK, response)
}

// DeleteFile deletes a file by UUID (owner/admin only)
// DELETE /files/info/:id
func (fc *FileController) DeleteFile(c *gin.Context) {
	// CHECK 401: Kiểm tra đăng nhập
	currentUserID := getUserIDFromContext(c)
	if currentUserID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid or missing authentication token",
		})
		return
	}

	// CHECK 400: Validate Input - Must be UUID
	idStr := c.Param("id")
	fileID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"message": "Invalid file ID format (Must be UUID)",
		})
		return
	}

	// CHECK 404: Tìm file
	file, err := fc.fileService.GetByID(fileID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not found",
				"message": "File not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to retrieve file",
		})
		return
	}

	// CHECK 403: Kiểm tra quyền truy cập (chỉ owner hoặc admin)
	// Anonymous uploads (OwnerID == nil) không thể xóa
	if file.OwnerID == nil {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "Anonymous uploads cannot be deleted",
		})
		return
	}

	currentUserRole := getUserRoleFromContext(c)
	isOwner := *currentUserID == *file.OwnerID
	isAdmin := currentUserRole == models.RoleAdmin
	if !isOwner && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "You don't have permission to delete this file",
		})
		return
	}

	// Delete file
	err = fc.fileService.Delete(fileID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to delete file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
		"fileId":  fileID,
	})
}

// DownloadFile handles file download by share token
// GET /files/:shareToken/download
func (fc *FileController) DownloadFile(c *gin.Context) {
	shareToken := c.Param("shareToken")

	// Get current user (optional - for authenticated downloads)
	currentUserID := getUserIDFromContext(c)
	currentUserEmail := getUserEmailFromContext(c)

	// Get file metadata from database with relationships
	file, err := fc.fileService.GetByShareToken(shareToken)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "Not found",
				"message": "File not found",
			})
			return
		}
		return
	}

	// Security check 1: File status (expired/pending)
	// Owner can bypass pending status for preview
	isOwner := currentUserID != nil && file.OwnerID != nil && *currentUserID == *file.OwnerID
	status := file.GetStatus()

	if status == "expired" {
		c.JSON(http.StatusGone, gin.H{
			"error":     "File expired",
			"expiredAt": file.AvailableTo,
		})
		return
	}

	if status == "pending" && !isOwner {
		// Calculate hours until available
		hoursUntilAvailable := 0.0
		if file.AvailableFrom != nil {
			hoursUntilAvailable = time.Until(*file.AvailableFrom).Hours()
		}
		c.JSON(http.StatusLocked, gin.H{
			"error":               "File not yet available",
			"availableFrom":       file.AvailableFrom,
			"hoursUntilAvailable": hoursUntilAvailable,
		})
		return
	}

	// Security check 2: Whitelist (sharedWith)
	// If file has sharedWith list, require authentication and verify email
	if len(file.SharedWith) > 0 {
		if currentUserID == nil || currentUserEmail == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "This file requires authentication. Please provide a Bearer token",
			})
			return
		}

		// Check if user email is in whitelist
		isWhitelisted := false
		for _, shared := range file.SharedWith {
			if shared.User.Email != "" && strings.EqualFold(shared.User.Email, currentUserEmail) {
				isWhitelisted = true
				break
			}
		}

		// Owner is always allowed
		if !isWhitelisted && !isOwner {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Access denied",
				"message": "You are not allowed to download this file. Your email is not in the shared list",
			})
			return
		}
	}

	// Security check 3: Password verification
	if file.PasswordHash != nil && *file.PasswordHash != "" {
		password := strings.TrimSpace(c.GetHeader("X-File-Password"))

		if password == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Password required",
				"message": "This file is password-protected. Please provide the password via X-File-Password header",
			})
			return
		}

		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(*file.PasswordHash), []byte(password)); err != nil {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Incorrect password",
				"message": "The file password is incorrect",
			})
			return
		}
	}

	// Determine container type based on file's isPublic field
	container := containerFromFile(file)

	// Download file from storage
	downloadResult, err := fc.fileService.Download(c.Request.Context(), &file.FilePath, container)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Download failed",
			"message": err.Error(),
		})
		return
	}
	defer downloadResult.Reader.Close()

	// Set response headers
	c.Header("Content-Disposition", "attachment; filename=\""+file.FileName+"\"")
	c.Header("Content-Type", downloadResult.ContentType)
	c.Header("Content-Length", fmt.Sprintf("%d", downloadResult.Size))

	// Stream file to response
	c.Status(http.StatusOK)
	bytesCopied, copyError := io.Copy(c.Writer, downloadResult.Reader)

	isCompleted := copyError == nil && bytesCopied == downloadResult.Size

	fileID := file.ID
	userID := currentUserID

	go func() {
		// GHI LỊCH SỬ
		err := fc.historyService.Create(&models.DownloadHistory{
			FileID:            fileID,
			DownloaderID:      userID, // nil nếu là khách
			DownloadCompleted: &isCompleted,
			DownloadedAt:      time.Now(),
		})
		if err != nil {
			fmt.Printf("Failed to record history: %v\n", err)
		}

		if isCompleted {
			_ = fc.statsService.IncrementDownloadCount(fileID)
			_ = fc.statsService.UpdateLastDownloadedAt(fileID)

			// XỬ LÝ UNIQUE DOWNLOADER
			if userID != nil {
				history, _ := fc.historyService.GetByFileIDAndDownloaderID(fileID, *userID)

				successCount := 0
				for _, h := range history {
					if h.DownloadCompleted != nil && *h.DownloadCompleted {
						successCount++
					}
				}
				if successCount == 1 {
					_ = fc.statsService.IncrementUniqueDownloaders(fileID)
				}
			}
		}
	}()
}

func containerFromFile(file *models.File) storage.ContainerType {
	if file != nil && file.IsPublic != nil && *file.IsPublic {
		return storage.ContainerPublic
	}
	return storage.ContainerPrivate
}

// getUserIDFromContext extracts user ID from JWT token in context
// Returns nil if user is not authenticated
func getUserIDFromContext(c *gin.Context) *uuid.UUID {
	// Try to get user ID from context (set by auth middleware)
	userIDVal, exists := c.Get("userID")
	if !exists {
		return nil
	}

	userID, ok := userIDVal.(uuid.UUID)
	if !ok {
		// Try string conversion
		if userIDStr, ok := userIDVal.(string); ok {
			parsed, err := uuid.Parse(userIDStr)
			if err == nil {
				return &parsed
			}
		}
		return nil
	}

	return &userID
}

// getUserEmailFromContext extracts user email from JWT token in context
// Returns empty string if user is not authenticated
func getUserEmailFromContext(c *gin.Context) string {
	emailVal, exists := c.Get("userEmail")
	if !exists {
		return ""
	}

	email, ok := emailVal.(string)
	if !ok {
		return ""
	}

	return email
}

// getSystemPolicy retrieves system policy from database
// Returns default policy if not found or on error
func getSystemPolicy(ctx context.Context, fileService *services.FileService) (*SystemPolicy, error) {
	// TODO: Add GetSystemPolicy method to FileService to access db
	// For now, return default policy
	// In production, this should query the database
	return &SystemPolicy{
		MaxFileSizeMB:            50,
		MinValidityHours:         1,
		MaxValidityDays:          30,
		DefaultValidityDays:      7,
		RequirePasswordMinLength: 8,
	}, nil
}

// SystemPolicy represents system configuration
type SystemPolicy struct {
	MaxFileSizeMB            int
	MinValidityHours         int
	MaxValidityDays          int
	DefaultValidityDays      int
	RequirePasswordMinLength int
}

// detectContentType detects MIME type from file extension
func detectContentType(ext string) string {
	// Common MIME types mapping
	mimeTypes := map[string]string{
		// Microsoft Office (legacy)
		".doc": "application/msword",
		".xls": "application/vnd.ms-excel",
		".ppt": "application/vnd.ms-powerpoint",

		// Microsoft Office (modern - Office Open XML)
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",

		// PDF
		".pdf": "application/pdf",

		// Images
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".bmp":  "image/bmp",
		".webp": "image/webp",
		".svg":  "image/svg+xml",

		// Text
		".txt":  "text/plain",
		".csv":  "text/csv",
		".html": "text/html",
		".css":  "text/css",
		".js":   "application/javascript",
		".json": "application/json",
		".xml":  "application/xml",

		// Archives
		".zip": "application/zip",
		".rar": "application/x-rar-compressed",
		".7z":  "application/x-7z-compressed",
		".tar": "application/x-tar",
		".gz":  "application/gzip",

		// Audio
		".mp3": "audio/mpeg",
		".wav": "audio/wav",
		".ogg": "audio/ogg",
		".m4a": "audio/mp4",

		// Video
		".mp4":  "video/mp4",
		".avi":  "video/x-msvideo",
		".mov":  "video/quicktime",
		".wmv":  "video/x-ms-wmv",
		".flv":  "video/x-flv",
		".webm": "video/webm",
	}

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}

	// Fallback to Go's built-in mime detection
	if detectedType := mime.TypeByExtension(ext); detectedType != "" {
		return detectedType
	}

	// Default fallback
	return "application/octet-stream"
}

// GET /files/stats/:id
func (fc *FileController) GetFileStats(c *gin.Context) {
	//CHECK 401: Kiểm tra đăng nhập
	currentUserID := getUserIDFromContext(c)
	if currentUserID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid or missing authentication token",
		})
		return
	}

	// CHECK 400: Validate Input
	idStr := c.Param("id")
	fileID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"message": "Invalid file ID format (Must be UUID)",
		})
		return
	}

	// CHECK 404: Tìm file
	file, err := fc.fileService.GetByID(fileID)
	if err != nil || file.OwnerID == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not found",
			"message": "File not found or statistics not available (anonymous upload)",
		})
		return
	}

	//CHECK 403: Kiểm tra chính chủ
	currentUserRole := getUserRoleFromContext(c)
	isOwner := *currentUserID == *file.OwnerID
	isAdmin := currentUserRole == models.RoleAdmin
	if !isOwner && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "You don't have permission to view statistics for this file",
		})
		return
	}

	stats, err := fc.statsService.GetByFileID(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not found",
			"message": "Statistics data not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"fileId":   file.ID,
		"fileName": file.FileName,
		"statistics": gin.H{
			"downloadCount":     stats.DownloadCount,
			"uniqueDownloaders": stats.UniqueDownloaders,
			"lastDownloadedAt":  stats.LastDownloadedAt,
			"createdAt":         stats.CreatedAt,
		},
	})
}

func getUserRoleFromContext(c *gin.Context) models.UserRole {
	roleVal, exists := c.Get("userRole")
	if !exists {
		return models.RoleUser // Default
	}

	role, ok := roleVal.(models.UserRole)
	if !ok {
		// Try string conversion
		if roleStr, ok := roleVal.(string); ok {
			if roleStr == "admin" {
				return models.RoleAdmin
			}
		}
		return models.RoleUser
	}

	return role
}

// GetMyFiles handles GET /files/my - List files owned by current user
func (fc *FileController) GetMyFiles(c *gin.Context) {
	// CHECK 401: Require authentication
	currentUserID := getUserIDFromContext(c)
	if currentUserID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid or missing authentication token",
		})
		return
	}

	// Parse query parameters
	status := c.DefaultQuery("status", "all")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	sortBy := c.DefaultQuery("sortBy", "createdAt")
	order := c.DefaultQuery("order", "desc")

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	offset := (page - 1) * limit

	// Get files from service
	files, total, err := fc.fileService.GetByOwnerID(*currentUserID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to retrieve files",
		})
		return
	}

	// Filter by status if needed
	var filteredFiles []models.File
	if status != "all" {
		for _, file := range files {
			fileStatus := file.GetStatus()
			if fileStatus == status {
				filteredFiles = append(filteredFiles, file)
			}
		}
	} else {
		filteredFiles = files
	}

	// Calculate summary (count by status)
	activeCount := 0
	pendingCount := 0
	expiredCount := 0
	deletedCount := 0

	// Get all files for summary (not just current page)
	allFiles, _, err := fc.fileService.GetByOwnerID(*currentUserID, 10000, 0)
	if err == nil {
		for _, file := range allFiles {
			fileStatus := file.GetStatus()
			switch fileStatus {
			case "active":
				activeCount++
			case "pending":
				pendingCount++
			case "expired":
				expiredCount++
			}
		}
	}

	summary := gin.H{
		"activeFiles":  activeCount,
		"pendingFiles": pendingCount,
		"expiredFiles": expiredCount,
		"deletedFiles": deletedCount,
	}

	// Sort filtered files
	if sortBy == "fileName" {
		// Simple string sort
		for i := 0; i < len(filteredFiles)-1; i++ {
			for j := i + 1; j < len(filteredFiles); j++ {
				if order == "asc" && filteredFiles[i].FileName > filteredFiles[j].FileName {
					filteredFiles[i], filteredFiles[j] = filteredFiles[j], filteredFiles[i]
				} else if order == "desc" && filteredFiles[i].FileName < filteredFiles[j].FileName {
					filteredFiles[i], filteredFiles[j] = filteredFiles[j], filteredFiles[i]
				}
			}
		}
	}
	// createdAt sort is already handled by service (Order("created_at DESC"))

	// Build response
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	response := gin.H{
		"files": filteredFiles,
		"pagination": gin.H{
			"currentPage": page,
			"totalPages":  totalPages,
			"totalFiles":  int(total),
			"limit":       limit,
		},
		"summary": summary,
	}

	c.JSON(http.StatusOK, response)
}

// GET /files/download-history/:id
func (fc *FileController) GetDownloadHistory(c *gin.Context) {
	//AUTH CHECK (401)
	currentUserID := getUserIDFromContext(c)
	if currentUserID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Invalid or missing authentication token",
		})
		return
	}

	// VALIDATE ID (400)
	idStr := c.Param("id")
	fileID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"message": "Invalid file ID format",
		})
		return
	}

	// CHECK FILE (404)
	file, err := fc.fileService.GetByID(fileID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not found",
			"message": "File not found",
		})
		return
	}

	// CHECK PERMISSION (403)
	currentUserRole := getUserRoleFromContext(c)
	isOwner := file.OwnerID != nil && *currentUserID == *file.OwnerID
	isAdmin := currentUserRole == models.RoleAdmin

	if !isOwner && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "Forbidden",
			"message": "You don't have permission to view download history for this file",
		})
		return
	}

	// PAGINATION LOGIC
	pageStr := c.DefaultQuery("page", "1")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}

	// Lấy limit (default 50, max 100 theo Swagger)
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	offset := (page - 1) * limit

	// GỌI SERVICE
	histories, totalRecords, err := fc.historyService.GetByFileID(fileID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to retrieve history",
		})
		return
	}

	// MAP RESPONSE (Biến đổi struct DB thành JSON Swagger)
	var historyResponse []gin.H
	for _, h := range histories {
		// Xử lý object downloader (Null nếu là anonymous)
		var downloaderInfo interface{} = nil
		if h.Downloader != nil {
			downloaderInfo = gin.H{
				"username": h.Downloader.Username,
				"email":    h.Downloader.Email,
			}
		}

		historyResponse = append(historyResponse, gin.H{
			"id":                h.ID,
			"downloader":        downloaderInfo,
			"downloadedAt":      h.DownloadedAt,
			"downloadCompleted": h.DownloadCompleted,
		})
	}

	// Tính tổng số trang
	totalPages := int(math.Ceil(float64(totalRecords) / float64(limit)))
	if totalPages == 0 {
		totalPages = 1
	}

	c.JSON(http.StatusOK, gin.H{
		"fileId":   file.ID,
		"fileName": file.FileName,
		"history":  historyResponse,
		"pagination": gin.H{
			"currentPage":  page,
			"totalPages":   totalPages,
			"totalRecords": totalRecords,
			"limit":        limit,
		},
	})
}

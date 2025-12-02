package controllers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"
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
	fileService *services.FileService
}

func NewFileController(fileService *services.FileService) *FileController {
	return &FileController{
		fileService: fileService,
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

// GetFileInfo returns file metadata without downloading
// GET /files/:shareToken
func (fc *FileController) GetFileInfo(c *gin.Context) {
	shareToken := c.Param("shareToken")
	if shareToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"message": "Share token is required",
		})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to retrieve file",
		})
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

	// Build response with safe metadata (no sensitive info like password hash)
	response := gin.H{
		"file": gin.H{
			"id":         file.ID,
			"fileName":   file.FileName,
			"fileSize":   file.FileSize,
			"mimeType":   file.MimeType,
			"shareToken": file.ShareToken,
			"isPublic":   file.IsPublic,
			"status":     status,
			"createdAt":  file.CreatedAt,
		},
	}

	// Add availability info if present
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

	// Add password protection indicator (don't expose actual hash)
	hasPassword := file.PasswordHash != nil && *file.PasswordHash != ""
	response["file"].(gin.H)["hasPassword"] = hasPassword

	// Add owner info if present (basic info only)
	if file.Owner != nil {
		response["file"].(gin.H)["owner"] = gin.H{
			"id":       file.Owner.ID,
			"username": file.Owner.Username,
		}
	}

	c.JSON(http.StatusOK, response)
}

// DownloadFile handles file download by share token
// GET /files/:shareToken/download
func (fc *FileController) DownloadFile(c *gin.Context) {
	shareToken := c.Param("shareToken")
	if shareToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"message": "Share token is required",
		})
		return
	}

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
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Failed to retrieve file",
		})
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
			"error":                "File not yet available",
			"availableFrom":        file.AvailableFrom,
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
		password := c.Query("password")
		if password == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Password required",
				"message": "This file is password-protected. Please provide the password parameter",
			})
			return
		}

		// Verify password
		err := bcrypt.CompareHashAndPassword([]byte(*file.PasswordHash), []byte(password))
		if err != nil {
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
	io.Copy(c.Writer, downloadResult.Reader)
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
		".doc":  "application/msword",
		".xls":  "application/vnd.ms-excel",
		".ppt":  "application/vnd.ms-powerpoint",
		
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
		".mp3":  "audio/mpeg",
		".wav":  "audio/wav",
		".ogg":  "audio/ogg",
		".m4a":  "audio/mp4",
		
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


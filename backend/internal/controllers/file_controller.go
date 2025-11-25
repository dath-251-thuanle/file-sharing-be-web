package controllers

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/services"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/storage"
	"github.com/gin-gonic/gin"
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
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "cannot open uploaded file"})
		return
	}
	defer file.Close()

	isPublic := c.DefaultPostForm("is_public", "true") == "true"

	// Get Content-Type from header, or detect from file extension
	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" || contentType == "application/octet-stream" {
		// Try to detect from file extension
		ext := filepath.Ext(fileHeader.Filename)
		contentType = detectContentType(ext)
	}

	uploadInput := &services.UploadInput{
		FileName:    fileHeader.Filename,
		ContentType: contentType,
		Size:        fileHeader.Size,
		Reader:      file,
		IsPublic:    &isPublic,
	}

	storedFile, err := fc.fileService.UploadFile(c.Request.Context(), uploadInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "File uploaded successfully",
		"file": gin.H{
			"id":         storedFile.ID,
			"file_name":  storedFile.FileName,
			"shareToken": storedFile.ShareToken,
			"is_public":  storedFile.IsPublic,
		},
	})
}

// DownloadFile handles file download by share token
// GET /files/:shareToken/download
func (fc *FileController) DownloadFile(c *gin.Context) {
	shareToken := c.Param("shareToken")
	if shareToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "share token is required"})
		return
	}

	// Get file metadata from database
	file, err := fc.fileService.GetByShareToken(shareToken)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "Not found",
			"message": "File not found",
		})
		return
	}

	// Check file availability (status: pending/active/expired)
	if !file.IsAvailable() {
		status := file.GetStatus()
		if status == "expired" {
			c.JSON(http.StatusGone, gin.H{
				"error":     "File expired",
				"expiredAt": file.AvailableTo,
			})
			return
		}
		if status == "pending" {
			c.JSON(http.StatusLocked, gin.H{
				"error":         "File not yet available",
				"availableFrom": file.AvailableFrom,
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

// Helper function to determine container type from file
func containerFromFile(file *models.File) storage.ContainerType {
	if file != nil && file.IsPublic != nil && *file.IsPublic {
		return storage.ContainerPublic
	}
	return storage.ContainerPrivate
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


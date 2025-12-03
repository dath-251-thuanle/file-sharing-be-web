package routes

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/gin-gonic/gin"
)

// RegisterFileRoutes registers file-related routes under a given router group.
// The passed router is expected to be something like router.Group("/files").
func RegisterFileRoutes(router *gin.RouterGroup, fileController *controllers.FileController, authMiddleware gin.HandlerFunc) {
	// Public endpoints
	// POST /files/upload - Upload a file
	router.POST("/upload", fileController.UploadFile)

	// GET /files/:shareToken - Get file metadata (public, no auth required)
	router.GET("/:shareToken", fileController.GetFileInfo)

	// GET /files/:shareToken/download - Download a file
	router.GET("/:shareToken/download", fileController.DownloadFile)

	// Authenticated routes group
	authenticated := router.Group("")
	authenticated.Use(authMiddleware)
	{
		// GET /files/info/:id - Get file info by UUID (owner/admin only)
		authenticated.GET("/info/:id", fileController.GetFileByID)

		// DELETE /files/info/:id - Delete file by UUID (owner/admin only)
		authenticated.DELETE("/info/:id", fileController.DeleteFile)

		// GET /files/stats/:id - Get file statistics
		stats := authenticated.Group("/stats")
		{
			stats.GET("/:id", fileController.GetFileStats)
		}

		// GET /files/download-history/:id - Get download history
		downloadHistory := authenticated.Group("/download-history")
		{
			downloadHistory.GET("/:id", fileController.GetDownloadHistory)
		}
	}
}
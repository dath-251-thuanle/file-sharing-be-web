package routes

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterFileRoutes registers file-related routes
func RegisterFileRoutes(router *gin.RouterGroup, fileController *controllers.FileController, cfg *config.Config) {
	// POST /api/files/upload - Upload a file
	router.POST("/upload", fileController.UploadFile)

	// GET /api/files/:shareToken - Get file metadata (public, no auth required)
	router.GET("/:shareToken", fileController.GetFileInfo)

	// GET /api/files/:shareToken/download - Download a file
	router.GET("/:shareToken/download", fileController.DownloadFile)

	// Authenticated routes group
	authenticated := router.Group("")
	authenticated.Use(middleware.RequireAuth(cfg))
	{
		// GET /api/files/info/:id - Get file info by UUID (owner/admin only)
		authenticated.GET("/info/:id", fileController.GetFileByID)

		// GET /api/files/stats/:id - Get file statistics
		stats := authenticated.Group("/stats")
		{
			stats.GET("/:id", fileController.GetFileStats)
		}

		// GET /api/files/download-history/:id - Get download history
		downloadHistory := authenticated.Group("/download-history")
		{
			downloadHistory.GET("/:id", fileController.GetDownloadHistory)
		}
	}
}
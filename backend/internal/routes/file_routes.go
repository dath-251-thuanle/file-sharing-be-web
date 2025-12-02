package routes

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterFileRoutes(router *gin.RouterGroup, fileController *controllers.FileController, cfg *config.Config) {
	// POST /files/upload - Upload a file
	router.POST("/upload", fileController.UploadFile)

	router.GET("/:shareToken", fileController.GetFileInfo)

	authenticated := router.Group("")
	// GET /files/info/:id - Get file info by UUID (owner/admin only, requires auth)
	authenticated.Use(middleware.RequireAuth(cfg))
	{
		authenticated.GET("/info/:id", fileController.GetFileByID)
	}

	// GET /files/stats/:id - Get file statistics
	stats := router.Group("/stats")
	{
		stats.GET("/:id", fileController.GetFileStats)
	}

	// GET /files/download-history/:id - Get download history
	download_history := router.Group("/download-history")
	{
		download_history.GET("/:id", fileController.GetDownloadHistory)
	}

	// GET /files/:shareToken/download - Download a file
	router.GET("/:shareToken/download", fileController.DownloadFile)
}
package routes

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/gin-gonic/gin"
)

// RegisterFileRoutes registers all file-related routes
func RegisterFileRoutes(router *gin.Engine, fileController *controllers.FileController) {
	files := router.Group("/files")
	{
		// POST /files/upload - Upload a file
		files.POST("/upload", fileController.UploadFile)

		// GET /files/stats/:id - Get file statistics
		stats := files.Group("/stats")
		{
			stats.GET("/:id", fileController.GetFileStats)
		}

		// GET /files/download-history/:id - Get download history
		download_history := files.Group("/download-history")
		{
			download_history.GET("/:id", fileController.GetDownloadHistory)
		}

		// Route download giữ nguyên
		files.GET("/:shareToken/download", fileController.DownloadFile)
		// => /files/:shareToken/download
	}
}

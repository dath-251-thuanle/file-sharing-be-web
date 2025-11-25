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

		// GET /files/:shareToken/download - Download a file by share token
		files.GET("/:shareToken/download", fileController.DownloadFile)
	}
}


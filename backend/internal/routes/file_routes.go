package routes

import (
	"strings"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterFileRoutes(router *gin.RouterGroup, fileController *controllers.FileController, authMiddleware gin.HandlerFunc) {
	// Public endpoints
	// POST /files/upload - Upload a file
	router.POST("/upload", optionalAuth(authMiddleware), fileController.UploadFile)

	// GET /files/:shareToken - Get file metadata (public, no auth required)
	router.GET("/:shareToken", fileController.GetFileInfo)

	// GET /files/:shareToken/download - Download a file (requires valid Bearer token)
	router.GET("/:shareToken/download", optionalAuth(authMiddleware), fileController.DownloadFile)

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

// optionalAuth wraps the required auth middleware so that:
// - If there is no Bearer token, the request continues as anonymous.
// - If a Bearer token is present, it must be valid; otherwise, 401 is returned.
func optionalAuth(authMiddleware gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			// No token provided -> continue as anonymous request
			c.Next()
			return
		}

		// Token present -> delegate to full auth middleware
		authMiddleware(c)
	}
}
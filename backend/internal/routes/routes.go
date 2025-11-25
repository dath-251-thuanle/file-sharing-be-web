package routes

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/gin-gonic/gin"
)

// SetupRoutes registers all application routes
func SetupRoutes(router *gin.Engine, fileController *controllers.FileController) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Register file routes
	RegisterFileRoutes(router, fileController)

	// TODO: Add more routes here as needed
	// - Auth routes (register, login, logout, TOTP)
	// - Admin routes (cleanup, policy)
	// - User routes (profile, my files)
}


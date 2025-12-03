package routes

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/gin-gonic/gin"
)

// SetupRoutes registers all application routes
func SetupRoutes(
	router *gin.Engine,
	fileController *controllers.FileController,
	authController *controllers.AuthController,
	authMiddleware gin.HandlerFunc,
) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes
	auth := router.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/login/totp", authController.LoginTOTP)
		auth.POST("/logout", authController.Logout)

		protected := auth.Group("/")
		protected.Use(authMiddleware)
		{
			protected.POST("/totp/setup", authController.TOTPSetup)
			protected.POST("/totp/verify", authController.TOTPVerify)
		}
	}

	// User profile
	userGroup := router.Group("/")
	userGroup.Use(authMiddleware)
	{
		userGroup.GET("/user", authController.Profile)
	}

	// Register file routes
	RegisterFileRoutes(router, fileController)

	// TODO: Add more routes here as needed
	// - Auth routes (register, login, logout, TOTP)
	// - Admin routes (cleanup, policy)
	// - User routes (profile, my files)
}


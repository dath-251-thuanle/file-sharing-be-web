package routes

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/gin-gonic/gin"
)

// SetupRoutes registers all application routes.
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

	// API routes group with /api prefix
	api := router.Group("/api")

	// Public policy limits (max file size, password length)
	api.GET("/policy/limits", fileController.GetPolicyLimits)

	// Auth routes: /api/auth/*
	authGroup := api.Group("/auth")
	RegisterAuthRoutes(authGroup, authController, authMiddleware)

	// User profile route: /api/user
	userGroup := api.Group("/user")
	userGroup.Use(authMiddleware)
	{
		userGroup.GET("", authController.Profile)
	}

	// File routes: /api/files/*
	filesGroup := api.Group("/files")
	RegisterFileRoutes(filesGroup, fileController, authMiddleware)
}


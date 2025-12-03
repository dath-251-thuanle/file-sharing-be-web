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

	// Base API group (no /api prefix to match docs: /auth, /user, /files, ...)
	api := router.Group("/")

	// Auth routes: /auth/*
	authGroup := api.Group("/auth")
	RegisterAuthRoutes(authGroup, authController, authMiddleware)

	// User profile route: /user
	userGroup := api.Group("/user")
	userGroup.Use(authMiddleware)
	{
		userGroup.GET("", authController.Profile)
	}

	// File routes: /files/*
	filesGroup := api.Group("/files")
	RegisterFileRoutes(filesGroup, fileController, authMiddleware)
}


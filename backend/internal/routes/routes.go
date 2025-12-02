package routes

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/database"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/middleware"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/repositories"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/services"
)

func SetupRoutes(router *gin.Engine, fileController *controllers.FileController) {
	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	db := database.GetDB()

	cfg, err := config.Load()
	if err != nil {
		log.Printf("Warning: Failed to load config in routes, defaults will be used: %v", err)
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)

	// Initialize services
	userService := services.NewUserService(userRepo, cfg)

	// Initialize controllers
	authController := controllers.NewAuthController(userService)
	userController := controllers.NewUserController(userRepo)

	// Register auth routes (public)
	auth := router.Group("/api/auth")
	RegisterAuthRoutes(auth, authController)

	// Register user routes (with optional JWT middleware)
	user := router.Group("/api/user")
	user.Use(middleware.JWTAuthMiddleware(cfg))
	RegisterUserRoutes(user, userController, cfg)

	// Register file routes (with optional JWT middleware)
	files := router.Group("/api/files")
	files.Use(middleware.JWTAuthMiddleware(cfg))
	RegisterFileRoutes(files, fileController, cfg)
}

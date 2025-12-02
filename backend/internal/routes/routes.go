package routes

import (
	"log"

	"github.com/gin-gonic/gin"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/database"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/repositories"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/services"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/middleware"
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

	userRepo := repositories.NewUserRepository(db)
	userService := services.NewUserService(userRepo, cfg)
	authController := controllers.NewAuthController(userService)

	// Đăng ký group route cho auth
	auth := router.Group("/api/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.GET("/profile", middleware.JWTAuthMiddleware(cfg), authController.GetProfile) // Route lấy thông tin profile
		auth.POST("/logout", authController.Logout) // Route đăng xuất
	}

	// Đăng ký các route file
	RegisterFileRoutes(router, fileController)
}

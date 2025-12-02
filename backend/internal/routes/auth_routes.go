package routes

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/gin-gonic/gin"
)

// RegisterAuthRoutes registers authentication routes
func RegisterAuthRoutes(router *gin.RouterGroup, authController *controllers.AuthController) {
	// POST /api/auth/register - Register new user
	router.POST("/register", authController.Register)
	
	// POST /api/auth/login - Login user
	router.POST("/login", authController.Login)
	
	// POST /api/auth/logout - Logout user
	router.POST("/logout", authController.Logout)
}

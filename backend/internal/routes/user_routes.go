package routes

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes registers user-related routes
func RegisterUserRoutes(router *gin.RouterGroup, userController *controllers.UserController, cfg *config.Config) {
	// GET /api/user - Get current user profile (requires auth)
	router.GET("", middleware.RequireAuth(cfg), userController.GetProfile)
}

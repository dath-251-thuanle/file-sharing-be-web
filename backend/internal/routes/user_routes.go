package routes

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/gin-gonic/gin"
)

// RegisterUserRoutes registers user-related routes under a given router group.
// The passed router is expected to be something like router.Group("/user").
func RegisterUserRoutes(router *gin.RouterGroup, userController *controllers.UserController, authMiddleware gin.HandlerFunc) {
	protected := router.Group("")
	protected.Use(authMiddleware)
	{
		// GET /user - Get current user profile (requires auth)
		protected.GET("", userController.GetProfile)
	}
}

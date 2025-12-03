package routes

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(router *gin.RouterGroup, userController *controllers.UserController, authMiddleware gin.HandlerFunc) {
	protected := router.Group("")
	protected.Use(authMiddleware)
	{
		// GET /user - Get current user profile (requires auth)
		protected.GET("", userController.GetProfile)
	}
}

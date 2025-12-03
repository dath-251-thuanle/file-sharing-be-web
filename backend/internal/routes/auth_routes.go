package routes

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(router *gin.RouterGroup, authController *controllers.AuthController, authMiddleware gin.HandlerFunc) {
	// Public auth endpoints
	// POST /auth/register - Register new user
	router.POST("/register", authController.Register)

	// POST /auth/login - Login user (password step)
	router.POST("/login", authController.Login)

	// POST /auth/login/totp - Login with TOTP after password step
	router.POST("/login/totp", authController.LoginTOTP)

	// Protected auth endpoints (require valid JWT)
	protected := router.Group("")
	protected.Use(authMiddleware)
	{
		// POST /auth/totp/setup - Generate TOTP secret and QR
		protected.POST("/totp/setup", authController.TOTPSetup)

		// POST /auth/totp/verify - Verify TOTP and enable it
		protected.POST("/totp/verify", authController.TOTPVerify)

		// POST /auth/totp/disable - Disable TOTP (requires TOTP code)
		protected.POST("/totp/disable", authController.DisableTOTP)

		// POST /auth/password/change - Change password (requires old password or TOTP code)
		protected.POST("/password/change", authController.ChangePassword)

		// POST /auth/logout - Logout user
		protected.POST("/logout", authController.Logout)
	}
}

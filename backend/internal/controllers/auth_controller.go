package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// AuthController - Controller cho các API liên quan đến Auth
type AuthController struct {}

// NewAuthController - Khởi tạo AuthController
func NewAuthController() *AuthController {
	return &AuthController{}
}

// Logout - Đăng xuất người dùng
// POST /auth/logout
func (ac *AuthController) Logout(c *gin.Context) {
	// Xóa JWT token khỏi client (ví dụ xóa cookie, hoặc thông báo đăng xuất)
	c.JSON(http.StatusOK, gin.H{
		"message": "User logged out successfully",
	})
}


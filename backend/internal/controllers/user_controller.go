package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/repositories"
	"github.com/google/uuid"
)

// UserController - Controller cho các API liên quan đến User
type UserController struct {
	userRepo repositories.UserRepository
}

// NewUserController - Khởi tạo UserController
func NewUserController(userRepo repositories.UserRepository) *UserController {
	return &UserController{userRepo: userRepo}
}

// GetProfile - Lấy thông tin profile của người dùng
// GET /user
func (uc *UserController) GetProfile(c *gin.Context) {
	// Lấy user ID từ context (middleware xác thực)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Unauthorized",
			"message": "Missing or invalid authentication token",
		})
		return
	}

	// Lấy thông tin người dùng từ repository
	user, err := uc.userRepo.GetByID(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal server error",
			"message": "Unable to retrieve user profile",
		})
		return
	}

	// Trả về thông tin người dùng
	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
			"totpEnabled": user.TOTPEnabled,
		},
	})
}

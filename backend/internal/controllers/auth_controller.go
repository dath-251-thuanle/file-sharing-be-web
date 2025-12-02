package controllers

import (
	"fmt"
	"net/http"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type AuthController struct {
	userService *services.UserService
}

func NewAuthController(userService *services.UserService) *AuthController {
	return &AuthController{
		userService: userService,
	}
}

func getValidationMessage(err error) string {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			switch fieldError.Tag() {
			case "required":
				return fmt.Sprintf("%s is required", fieldError.Field())
			case "email":
				return fmt.Sprintf("%s format is invalid", fieldError.Field())
			case "min":
				return fmt.Sprintf("%s must have at least %s characters", fieldError.Field(), fieldError.Param())
			}
		}
	}
	return err.Error()
}

func (ac *AuthController) Register(c *gin.Context) {
	var input services.RegisterInput

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Validation error",
			"message": getValidationMessage(err),
		})
		return
	}

	user, err := ac.userService.Register(input)
	if err != nil {
		if err.Error() == "email already exists" || err.Error() == "username already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "Conflict",
				"message": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Internal Server Error",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
		"userId":  user.ID,
	})
}

func (ac *AuthController) Login(c *gin.Context) {
	var input services.LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	token, user, err := ac.userService.Login(input)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Login successful",
		"access_token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

package controllers

import (
	"net/http"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

type registerRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type totpLoginRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type totpVerifyRequest struct {
	Code string `json:"code" binding:"required,len=6"`
}

// Register handles POST /auth/register
func (a *AuthController) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	user, err := a.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		if err == services.ErrUserExists {
			writeError(c, http.StatusConflict, "Conflict", "Email or username already exists")
			return
		}
		writeError(c, http.StatusBadRequest, "Validation error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User registered successfully",
		"userId":  user.ID,
	})
}

// Login handles POST /auth/login
func (a *AuthController) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	user, totpEnabled, err := a.authService.Login(req.Email, req.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			writeError(c, http.StatusUnauthorized, "Unauthorized", "Invalid email or password")
			return
		}
		writeError(c, http.StatusInternalServerError, "Internal error", err.Error())
		return
	}

	if totpEnabled {
		c.JSON(http.StatusOK, gin.H{
			"requireTOTP": true,
			"message":     "TOTP verification required",
		})
		return
	}

	token, err := a.authService.GenerateAccessToken(user)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Internal error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken": token,
		"user":        sanitizeUser(user),
	})
}

// LoginTOTP handles POST /auth/login/totp
func (a *AuthController) LoginTOTP(c *gin.Context) {
	var req totpLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	user, err := a.authService.LoginWithTOTP(req.Email, req.Code)
	if err != nil {
		switch err {
		case services.ErrInvalidCredentials, services.ErrInvalidTOTPCode, services.ErrTOTPNotEnabled, services.ErrTOTPSecretNotCreated:
			writeError(c, http.StatusUnauthorized, "Unauthorized", "Invalid or expired TOTP code")
			return
		default:
			writeError(c, http.StatusInternalServerError, "Internal error", err.Error())
			return
		}
	}

	token, err := a.authService.GenerateAccessToken(user)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Internal error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accessToken": token,
		"user":        sanitizeUser(user),
	})
}

// TOTPSetup handles POST /auth/totp/setup
func (a *AuthController) TOTPSetup(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "Unauthorized", "Invalid user context")
		return
	}

	setup, err := a.authService.SetupTOTP(userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Internal error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "TOTP secret generated",
		"totpSetup": gin.H{
			"secret": setup.Secret,
			"qrCode": setup.QRCode,
		},
	})
}

// TOTPVerify handles POST /auth/totp/verify
func (a *AuthController) TOTPVerify(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "Unauthorized", "Invalid user context")
		return
	}

	var req totpVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeValidationError(c, err)
		return
	}

	if err := a.authService.VerifyTOTP(userID, req.Code); err != nil {
		if err == services.ErrInvalidTOTPCode {
			writeError(c, http.StatusBadRequest, "Invalid TOTP code", "The provided code is incorrect or expired")
			return
		}
		writeError(c, http.StatusInternalServerError, "Internal error", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "TOTP verified successfully",
		"totpEnabled": true,
	})
}

// Profile handles GET /user
func (a *AuthController) Profile(c *gin.Context) {
	userID, ok := userIDFromContext(c)
	if !ok {
		writeError(c, http.StatusUnauthorized, "Unauthorized", "Invalid user context")
		return
	}

	user, err := a.authService.GetProfile(userID)
	if err != nil {
		writeError(c, http.StatusInternalServerError, "Internal error", err.Error())
		return
	}
	if user == nil {
		writeError(c, http.StatusUnauthorized, "Unauthorized", "User not found")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": sanitizeUser(user),
	})
}

// Logout handles POST /auth/logout
func (a *AuthController) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "User logged out",
	})
}

func sanitizeUser(user *models.User) gin.H {
	return gin.H{
		"id":          user.ID,
		"username":    user.Username,
		"email":       user.Email,
		"role":        user.Role,
		"totpEnabled": user.TOTPEnabled != nil && *user.TOTPEnabled,
	}
}

func writeError(c *gin.Context, status int, errStr, msg string) {
	c.JSON(status, gin.H{
		"error":   errStr,
		"message": msg,
	})
}

func writeValidationError(c *gin.Context, err error) {
	writeError(c, http.StatusBadRequest, "Validation error", err.Error())
}

func userIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		return uuid.Nil, false
	}
	userIDStr, ok := userIDVal.(string)
	if !ok {
		return uuid.Nil, false
	}
	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, false
	}
	return uid, true
}

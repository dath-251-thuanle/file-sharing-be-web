package middleware

import (
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/dgrijalva/jwt-go"
	"net/http"
	"strings"
)

// AuthMiddleware - Middleware xác thực người dùng bằng JWT
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lấy Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Kiểm tra token
		tokenString := strings.Split(authHeader, "Bearer ")[1]
		claims := &models.Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("your_secret_key"), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Lưu thông tin người dùng vào context
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}

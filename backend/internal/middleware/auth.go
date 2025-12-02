package middleware

import (
	"net/http"
	"strings"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)


func JWTAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Invalid format - continue without setting user context
			c.Next()
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			c.Next()
			return
		}

		// Parse and verify JWT token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			c.Next()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.Next()
			return
		}

		// Extract user ID (sub claim)
		sub, ok := claims["sub"].(string)
		if ok && sub != "" {
			userID, err := uuid.Parse(sub)
			if err == nil {
				c.Set("userID", userID)
			}
		}

		// Extract email
		if email, ok := claims["email"].(string); ok && email != "" {
			c.Set("userEmail", email)
		}

		// Extract role
		if roleStr, ok := claims["role"].(string); ok {
			role := models.RoleUser
			if roleStr == "admin" {
				role = models.RoleAdmin
			}
			c.Set("userRole", role)
		}

		// Extract username (optional, for convenience)
		if username, ok := claims["username"].(string); ok && username != "" {
			c.Set("username", username)
		}

		c.Next()
	}
}

// RequireAuth creates a middleware that requires valid JWT token.
// Returns 401 if token is missing or invalid.
func RequireAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Bearer token is required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid token format. Use 'Bearer <token>'",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Bearer token is required",
			})
			c.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid or expired access token",
			})
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "Invalid token claims",
			})
			c.Abort()
			return
		}

		// Set user info in context
		sub, ok := claims["sub"].(string)
		if ok && sub != "" {
			userID, err := uuid.Parse(sub)
			if err == nil {
				c.Set("userID", userID)
			}
		}

		if email, ok := claims["email"].(string); ok && email != "" {
			c.Set("userEmail", email)
		}

		if roleStr, ok := claims["role"].(string); ok {
			role := models.RoleUser
			if roleStr == "admin" {
				role = models.RoleAdmin
			}
			c.Set("userRole", role)
		}

		if username, ok := claims["username"].(string); ok && username != "" {
			c.Set("username", username)
		}

		c.Next()
	}
}


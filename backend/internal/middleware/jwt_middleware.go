package middleware

import (
	"errors"
	"strings"

	"github.com/dath-251-thuanle/file-sharing-be-web/internal/config"
	"github.com/dath-251-thuanle/file-sharing-be-web/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func JWTAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(401, gin.H{"error": "Authorization token is required"})
			c.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.JSON(401, gin.H{"error": "Bearer token is required"})
			c.Abort()
			return
		}

		// Parse JWT Token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHS256); !ok {
				return nil, errors.New("Invalid signing method")
			}
			return []byte(cfg.JWT.Secret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(401, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Attach user information to the context
		userID, _ := claims["sub"].(string)
		username, _ := claims["username"].(string)
		email, _ := claims["email"].(string)
		role, _ := claims["role"].(string)

		user := &services.User{
			ID:       userID,
			Username: username,
			Email:    email,
			Role:     role,
		}

		// Save user to context
		c.Set("user", user)

		c.Next()
	}
}

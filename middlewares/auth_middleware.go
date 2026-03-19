package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"shopcart-api/utils"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthenticated: Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthenticated: Invalid Authorization header format"})
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthenticated: Invalid or expired token"})
			c.Abort()
			return
		}

		c.Set("user_id", uint(claims["user_id"].(float64)))
		c.Set("role", claims["role"].(string))
		c.Next()
	}
}

func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
			c.Abort()
			return
		}
		roleStr := role.(string)
		if roleStr != "ADMIN" && roleStr != "SUPERADMIN" {
			c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden: Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// Management allows ADMIN, MANAGER, SUPERVISOR
func Management() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden"})
			c.Abort()
			return
		}
		roleStr := role.(string)
		if roleStr != "ADMIN" && roleStr != "MANAGER" && roleStr != "SUPERVISOR" {
			c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden: Management role required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

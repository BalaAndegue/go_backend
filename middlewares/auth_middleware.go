package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"shopcart-api/config"
	"shopcart-api/models"
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

		// Only access tokens may be used to authenticate API requests.
		if t, _ := claims["type"].(string); t != "access" {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthenticated: Invalid token type"})
			c.Abort()
			return
		}

		userID := uint(claims["user_id"].(float64))

		// Verify the token version against the user to honour revocation.
		var user models.User
		if err := config.DB.Select("id", "role", "token_version").First(&user, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthenticated: User not found"})
			c.Abort()
			return
		}
		if ver, ok := claims["ver"].(float64); !ok || int(ver) != user.TokenVersion {
			c.JSON(http.StatusUnauthorized, gin.H{"message": "Unauthenticated: Token revoked"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Set("role", user.Role)
		c.Next()
	}
}

func Admin() gin.HandlerFunc {
	return requireRoles("Admin access required", "ADMIN", "SUPERADMIN")
}

// Management allows ADMIN, SUPERADMIN, MANAGER, SUPERVISOR
func Management() gin.HandlerFunc {
	return requireRoles("Management role required", "ADMIN", "SUPERADMIN", "MANAGER", "SUPERVISOR")
}

// ProductManagement allows the roles permitted to manage the product catalogue:
// ADMIN, SUPERADMIN and VENDOR. It exists because product CRUD has a different
// access policy from the broader Management endpoints.
func ProductManagement() gin.HandlerFunc {
	return requireRoles("Product management role required", "ADMIN", "SUPERADMIN", "VENDOR")
}

// requireRoles builds a middleware that allows only the given roles.
func requireRoles(message string, roles ...string) gin.HandlerFunc {
	allowed := make(map[string]bool, len(roles))
	for _, r := range roles {
		allowed[r] = true
	}
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || !allowed[role.(string)] {
			c.JSON(http.StatusForbidden, gin.H{"message": "Forbidden: " + message})
			c.Abort()
			return
		}
		c.Next()
	}
}

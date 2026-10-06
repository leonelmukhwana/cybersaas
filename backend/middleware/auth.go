package middleware

import (
	"net/http"
	"strings"

	"cybersaas/backend/auth"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID   = "user_id"
	ContextTenantID = "tenant_id"
	ContextRole     = "role"
)

// AuthMiddleware validates the JWT and stores the authenticated
// user's identity in the Gin context.
func AuthMiddleware(tokenManager *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := strings.TrimSpace(c.GetHeader("Authorization"))

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "authorization header is required",
			})
			c.Abort()
			return
		}

		parts := strings.Fields(authHeader)

		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid authorization header",
			})
			c.Abort()
			return
		}

		tokenString := strings.TrimSpace(parts[1])

		claims, err := tokenManager.Validate(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"message": "invalid or expired token",
			})
			c.Abort()
			return
		}

		c.Set(ContextUserID, claims.UserID)
		c.Set(ContextRole, claims.Role)

		if claims.TenantID != "" {
			c.Set(ContextTenantID, claims.TenantID)
		}

		c.Next()
	}
}

// CurrentUserID returns the authenticated user's ID.
func CurrentUserID(c *gin.Context) (string, bool) {
	value, exists := c.Get(ContextUserID)
	if !exists {
		return "", false
	}

	userID, ok := value.(string)

	return userID, ok && strings.TrimSpace(userID) != ""
}

// CurrentTenantID returns the authenticated user's tenant ID.
func CurrentTenantID(c *gin.Context) (string, bool) {
	value, exists := c.Get(ContextTenantID)
	if !exists {
		return "", false
	}

	tenantID, ok := value.(string)

	return tenantID, ok && strings.TrimSpace(tenantID) != ""
}

// CurrentRole returns the authenticated user's role.
func CurrentRole(c *gin.Context) (string, bool) {
	value, exists := c.Get(ContextRole)
	if !exists {
		return "", false
	}

	role, ok := value.(string)

	return role, ok && strings.TrimSpace(role) != ""
}

package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RequireRole allows only the specified roles to access a route.
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, ok := CurrentRole(c)

		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "role information is missing",
			})
			c.Abort()
			return
		}

		for _, allowedRole := range allowedRoles {
			if role == allowedRole {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{
			"message": "you do not have permission to access this resource",
		})
		c.Abort()
	}
}

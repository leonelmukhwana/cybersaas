package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type TerminalAuthenticator interface {
	AuthenticateTerminal(
		ctx context.Context,
		credentialHash string,
	) (string, string, string, error)
}

func TerminalAuthMiddleware(
	authenticator TerminalAuthenticator,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		if authenticator == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "terminal authenticator uninitialized",
			})
			c.Abort()
			return
		}

		authHeader := c.GetHeader("Authorization")

		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "terminal authorization required",
			})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)

		if len(parts) != 2 ||
			!strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid terminal authorization",
			})
			c.Abort()
			return
		}

		credential := strings.TrimSpace(parts[1])

		if credential == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "terminal credential required",
			})
			c.Abort()
			return
		}

		hash := sha256.Sum256([]byte(credential))
		credentialHash := hex.EncodeToString(hash[:])

		terminalID, tenantID, branchID, err :=
			authenticator.AuthenticateTerminal(
				c.Request.Context(),
				credentialHash,
			)

		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or revoked terminal credential",
			})
			c.Abort()
			return
		}

		c.Set("terminal_id", terminalID)
		c.Set("terminal_tenant_id", tenantID)
		c.Set("terminal_branch_id", branchID)

		c.Next()
	}
}

func CurrentTerminalID(c *gin.Context) (string, bool) {
	value, exists := c.Get("terminal_id")
	if !exists {
		return "", false
	}

	id, ok := value.(string)
	return id, ok
}

func CurrentTerminalTenantID(c *gin.Context) (string, bool) {
	value, exists := c.Get("terminal_tenant_id")
	if !exists {
		return "", false
	}

	id, ok := value.(string)
	return id, ok
}

func CurrentTerminalBranchID(c *gin.Context) (string, bool) {
	value, exists := c.Get("terminal_branch_id")
	if !exists {
		return "", false
	}

	id, ok := value.(string)
	return id, ok
}

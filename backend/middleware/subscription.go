package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SubscriptionAccessChecker interface {
	CheckWebAccess(
		ctx context.Context,
		tenantID string,
	) (bool, string, error)
}

func RequireActiveSubscription(
	checker SubscriptionAccessChecker,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, ok := CurrentTenantID(c)

		if !ok || tenantID == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "tenant ID is required",
			})
			c.Abort()
			return
		}

		if checker == nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "subscription service is unavailable",
			})
			c.Abort()
			return
		}

		allowed, reason, err := checker.CheckWebAccess(
			c.Request.Context(),
			tenantID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "unable to verify subscription",
			})
			c.Abort()
			return
		}

		if allowed {
			c.Next()
			return
		}

		switch reason {
		case "trial_expired":
			c.JSON(http.StatusForbidden, gin.H{
				"error": "free trial has expired",
				"code":  "TRIAL_EXPIRED",
			})

		case "subscription_expired":
			c.JSON(http.StatusForbidden, gin.H{
				"error": "subscription has expired",
				"code":  "SUBSCRIPTION_EXPIRED",
			})

		case "subscription_inactive":
			c.JSON(http.StatusForbidden, gin.H{
				"error": "subscription is not active",
				"code":  "SUBSCRIPTION_INACTIVE",
			})

		default:
			c.JSON(http.StatusForbidden, gin.H{
				"error": "subscription access is not available",
				"code":  "SUBSCRIPTION_ACCESS_DENIED",
			})
		}

		c.Abort()
	}
}

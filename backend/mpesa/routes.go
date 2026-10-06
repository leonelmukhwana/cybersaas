package mpesa

import (
	"cybersaas/backend/auth"
	"cybersaas/backend/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	tokenManager *auth.TokenManager,
	requireActiveSubscription gin.HandlerFunc,
) {
	mpesa := router.Group("/mpesa")

	mpesa.Use(middleware.AuthMiddleware(tokenManager))

	// ============================================================
	// CYBER OWNER / BRANCH M-PESA
	// ============================================================

	branch := mpesa.Group("/branches/:branchID")

	// Branch configuration is owner-only and requires
	// an active subscription.
	branchConfig := branch.Group("")
	branchConfig.Use(
		requireActiveSubscription,
		middleware.RequireRole("owner"),
	)

	branchConfig.GET("/config", handler.GetBranch)
	branchConfig.PUT("/config", handler.SaveBranch)
	branchConfig.DELETE("/config", handler.DeleteBranch)

	// Customer payment STK Push is a normal business operation,
	// so it requires an active subscription.
	branch.Use(requireActiveSubscription)

	branch.POST(
		"/stk",
		middleware.RequireRole("owner", "attendant"),
		handler.InitiateSTKPush,
	)

	// ============================================================
	// CYBERSAAS SUBSCRIPTION M-PESA
	// ============================================================

	// IMPORTANT:
	// Do NOT apply requireActiveSubscription here.
	// An expired trial/subscription must still be able to
	// initiate payment to activate/reactivate the subscription.
	subscription := mpesa.Group("/subscription")
	subscription.Use(middleware.RequireRole("owner"))

	subscription.POST("/stk", handler.InitiateSubscriptionSTKPush)

	// ============================================================
	// SAAS PLATFORM M-PESA
	// ============================================================

	platform := mpesa.Group("/platform")
	platform.Use(middleware.RequireRole("platform_admin"))

	platform.GET("/config", handler.GetPlatform)
	platform.PUT("/config", handler.SavePlatform)
}

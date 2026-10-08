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
	// ============================================================
	// SAFARICOM DARАJA CALLBACK
	// ============================================================
	// This endpoint must be public because Safaricom does not
	// send our application's JWT.
	callback := router.Group("/mpesa")
	callback.POST("/callback", handler.Callback)

	// ============================================================
	// AUTHENTICATED M-PESA APPLICATION ROUTES
	// ============================================================
	mpesa := router.Group("/mpesa")
	mpesa.Use(middleware.AuthMiddleware(tokenManager))

	// ============================================================
	// CYBER OWNER / BRANCH M-PESA
	// ============================================================

	branch := mpesa.Group("/branches/:branchID")

	branchConfig := branch.Group("")
	branchConfig.Use(
		requireActiveSubscription,
		middleware.RequireRole("owner"),
	)

	branchConfig.GET("/config", handler.GetBranch)
	branchConfig.PUT("/config", handler.SaveBranch)
	branchConfig.DELETE("/config", handler.DeleteBranch)

	branch.Use(requireActiveSubscription)

	branch.POST(
		"/stk",
		middleware.RequireRole("owner", "attendant"),
		handler.InitiateSTKPush,
	)

	// ============================================================
	// CYBERSAAS SUBSCRIPTION M-PESA
	// ============================================================

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

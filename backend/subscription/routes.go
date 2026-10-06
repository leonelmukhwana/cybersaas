package subscription

import (
	"cybersaas/backend/auth"
	"cybersaas/backend/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	tokenManager *auth.TokenManager,
) {
	// ------------------------------------------------------------
	// PUBLIC SUBSCRIPTION PLAN LIST
	// ------------------------------------------------------------
	// Used by landing page / registration / subscription screens.
	router.GET(
		"/subscription/plans",
		handler.Plans,
	)

	// ------------------------------------------------------------
	// PLATFORM ADMIN — SUBSCRIPTION PRICING
	// ------------------------------------------------------------
	// Only SaaS Owner / Platform Admin can manage packages.
	platform := router.Group("/platform")
	platform.Use(middleware.AuthMiddleware(tokenManager))
	platform.Use(middleware.RequireRole("platform_admin"))

	// View all plans, including inactive plans.
	platform.GET(
		"/subscription-plans",
		handler.PlatformPlans,
	)

	// Create a new subscription package.
	platform.POST(
		"/subscription-plans",
		handler.CreatePlatformPlan,
	)

	// Update an existing subscription package.
	platform.PATCH(
		"/subscription-plans/:id",
		handler.UpdatePlatformPlan,
	)

	// ------------------------------------------------------------
	// PROTECTED TENANT SUBSCRIPTION ROUTES
	// ------------------------------------------------------------

	subscription := router.Group("/subscription")
	subscription.Use(
		middleware.AuthMiddleware(tokenManager),
	)

	subscription.GET(
		"",
		middleware.RequireRole("owner", "attendant"),
		handler.Overview,
	)

	subscription.POST(
		"",
		middleware.RequireRole("owner"),
		handler.Create,
	)

	subscription.GET(
		"/payments",
		middleware.RequireRole("owner"),
		handler.Payments,
	)

	subscription.GET(
		"/ledger",
		middleware.RequireRole("owner"),
		handler.Ledger,
	)

	subscription.POST(
		"/payments",
		middleware.RequireRole("owner"),
		handler.CreatePayment,
	)
}

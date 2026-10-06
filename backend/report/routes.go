package report

import (
	"cybersaas/backend/auth"
	"cybersaas/backend/middleware"

	"github.com/gin-gonic/gin"
)

type RouteDependencies struct {
	Handler *Handler
}

func RegisterRoutes(
	router *gin.RouterGroup,
	deps RouteDependencies,
	tokenManager *auth.TokenManager,
	requireActiveSubscription gin.HandlerFunc,
) {
	reports := router.Group("/reports")

	// Cyber Attendant compliance report routes.
	// Requires authentication and an active subscription.
	reports.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
	)

	reports.POST(
		"/compliance",
		deps.Handler.GenerateComplianceReport,
	)

	reports.GET(
		"/compliance",
		deps.Handler.ListComplianceReports,
	)

	reports.GET(
		"/compliance/:id",
		deps.Handler.GetComplianceReport,
	)
}

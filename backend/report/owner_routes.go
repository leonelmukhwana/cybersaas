package report

import (
	"github.com/gin-gonic/gin"

	"cybersaas/backend/auth"
	"cybersaas/backend/middleware"
)

func RegisterOwnerRoutes(
	router *gin.RouterGroup,
	handler *OwnerReportHandler,
	tokenManager *auth.TokenManager,
	requireActiveSubscription gin.HandlerFunc,
) {
	ownerReports := router.Group("/reports/owner")

	ownerReports.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
		middleware.RequireRole("owner"),
	)

	ownerReports.GET(
		"/summary",
		handler.Summary,
	)

	ownerReports.GET(
		"/daily/download",
		handler.DownloadDaily,
	)

	ownerReports.GET(
		"/monthly/download",
		handler.DownloadMonthly,
	)
}

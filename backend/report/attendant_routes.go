package report

import (
	"github.com/gin-gonic/gin"

	"cybersaas/backend/auth"
	"cybersaas/backend/middleware"
)

func RegisterAttendantRoutes(
	router *gin.RouterGroup,
	handler *AttendantReportHandler,
	tokenManager *auth.TokenManager,
	requireActiveSubscription gin.HandlerFunc,
) {
	attendantReports := router.Group(
		"/reports/attendant",
	)

	attendantReports.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
		middleware.RequireRole("attendant"),
	)

	attendantReports.GET(
		"/summary",
		handler.Summary,
	)
}

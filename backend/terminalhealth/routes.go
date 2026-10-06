package terminalhealth

import (
	"github.com/gin-gonic/gin"

	"cybersaas/backend/auth"
	"cybersaas/backend/middleware"
	"cybersaas/backend/terminal"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	tokenManager *auth.TokenManager,
	terminalService *terminal.Service,
	requireActiveSubscription gin.HandlerFunc,
) {
	// ============================================================
	// TERMINAL CLIENT
	// ============================================================

	terminalRoutes := router.Group("/terminals")
	terminalRoutes.Use(
		middleware.TerminalAuthMiddleware(terminalService),
	)

	terminalRoutes.POST("/health", handler.ReportHealth)

	// ============================================================
	// CYBER ATTENDANT
	// ============================================================

	attendantHealthRoutes := router.Group("/attendant/health")
	attendantHealthRoutes.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
		middleware.RequireRole("attendant"),
	)

	attendantHealthRoutes.GET(
		"/terminals",
		handler.ListAttendantTerminals,
	)

	attendantHealthRoutes.GET(
		"/terminals/:terminal_id",
		handler.GetAttendantTerminalHealth,
	)
}

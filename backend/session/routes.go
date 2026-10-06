package session

import (
	"cybersaas/backend/auth"
	"cybersaas/backend/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	tokenManager *auth.TokenManager,
	terminalAuthenticator middleware.TerminalAuthenticator,
	requireActiveSubscription gin.HandlerFunc,
) {
	// ========================================================
	// OWNER / ATTENDANT SESSION API
	// ========================================================

	sessions := router.Group("/sessions")

	sessions.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
		middleware.RequireRole("owner", "attendant"),
	)

	sessions.POST("", handler.Start)
	sessions.GET("", handler.List)
	sessions.GET("/:id", handler.Get)
	sessions.GET("/:id/billing-preview", handler.Preview)

	// ========================================================
	// ATTENDANT / OWNER SESSION CONTROLS
	// ========================================================

	sessions.POST("/:id/pause", handler.Pause)
	sessions.POST("/:id/resume", handler.Resume)
	sessions.POST("/:id/end", handler.End)
	sessions.POST("/:id/cancel", handler.Cancel)

	// ========================================================
	// TERMINAL SESSION API
	//
	// IMPORTANT:
	// No subscription middleware here.
	// Terminals must continue operating after subscription
	// expiry.
	// ========================================================

	terminalSessions := router.Group("/terminal-sessions")

	terminalSessions.Use(
		middleware.TerminalAuthMiddleware(
			terminalAuthenticator,
		),
	)

	terminalSessions.POST("", handler.TerminalStart)

	// /active must be registered before /:id.
	terminalSessions.GET("/active", handler.TerminalActive)

	terminalSessions.GET("/:id", handler.TerminalGet)

	terminalSessions.GET(
		"/:id/billing-preview",
		handler.TerminalPreview,
	)

	terminalSessions.POST(
		"/:id/pause",
		handler.TerminalPause,
	)

	terminalSessions.POST(
		"/:id/resume",
		handler.TerminalResume,
	)

	terminalSessions.POST(
		"/:id/end",
		handler.TerminalEnd,
	)

	terminalSessions.GET(
		"/billing-config",
		handler.TerminalBillingConfig,
	)
}

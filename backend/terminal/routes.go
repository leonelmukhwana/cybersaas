package terminal

import (
	"cybersaas/backend/auth"
	"cybersaas/backend/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	tokenManager *auth.TokenManager,
	authenticator middleware.TerminalAuthenticator,
	requireActiveSubscription gin.HandlerFunc,
) {
	// ========================================================
	// PC CLIENT SELF-REGISTRATION
	//
	// Must remain available so a terminal can register using
	// a valid licence key.
	// ========================================================

	router.POST(
		"/terminals/register",
		handler.Register,
	)

	// ========================================================
	// TERMINAL AUTHENTICATION
	//
	// Must remain available independently of subscription
	// status.
	// ========================================================

	router.POST(
		"/terminals/auth",
		handler.Authenticate,
	)

	// ========================================================
	// AUTHENTICATED TERMINAL COMMUNICATION
	//
	// IMPORTANT:
	// No SaaS subscription middleware here.
	// Terminals must continue operating even when the web
	// subscription has expired.
	// ========================================================

	terminalAPI := router.Group("/terminals")

	terminalAPI.Use(
		middleware.TerminalAuthMiddleware(authenticator),
	)

	terminalAPI.POST(
		"/heartbeat",
		handler.Heartbeat,
	)

	terminalAPI.GET(
		"/control-state",
		handler.ControlState,
	)

	terminalAPI.GET(
		"/commands/pending",
		handler.PendingCommand,
	)

	terminalAPI.POST(
		"/commands/:id/ack",
		handler.AcknowledgeCommand,
	)

	// ========================================================
	// OWNER TERMINAL MANAGEMENT
	//
	// These are web-dashboard operations and require an active
	// subscription.
	// ========================================================

	terminals := router.Group("/terminals")

	terminals.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
		middleware.RequireRole("owner"),
	)

	terminals.GET("", handler.List)
	terminals.GET("/:id", handler.Get)

	terminals.PATCH(
		"/:id/name",
		handler.Rename,
	)

	terminals.PATCH(
		"/:id/status",
		handler.ChangeStatus,
	)

	terminals.PATCH(
		"/:id/branch",
		handler.Move,
	)

	terminals.PATCH(
		"/:id/lock-state",
		handler.SetLockState,
	)

	terminals.POST(
		"/:id/restart",
		handler.Restart,
	)

	terminals.POST(
		"/:id/shutdown",
		handler.Shutdown,
	)

	terminals.DELETE(
		"/:id",
		handler.Deregister,
	)

	// ========================================================
	// LICENCE KEY MANAGEMENT
	// ========================================================

	terminals.POST(
		"/licence-keys",
		handler.GenerateLicenceKey,
	)

	terminals.GET(
		"/licence-keys",
		handler.ListLicenceKeys,
	)

	terminals.DELETE(
		"/licence-keys/:id",
		handler.RevokeLicenceKey,
	)

	// ========================================================
	// ATTENDANT TERMINAL MANAGEMENT
	//
	// Web-dashboard operations require an active subscription.
	// ========================================================

	attendantTerminals := router.Group("/attendant/terminals")

	attendantTerminals.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
		middleware.RequireRole("attendant"),
	)

	attendantTerminals.GET(
		"",
		handler.ListAttendantTerminals,
	)

	attendantTerminals.GET(
		"/:id",
		handler.GetAttendantTerminal,
	)

	attendantTerminals.PATCH(
		"/:id/lock-state",
		handler.SetAttendantLockState,
	)

	attendantTerminals.POST(
		"/:id/restart",
		handler.RestartAttendantTerminal,
	)

	attendantTerminals.POST(
		"/:id/shutdown",
		handler.ShutdownAttendantTerminal,
	)
}

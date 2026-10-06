package customer

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
	// Web dashboard customer APIs.
	// These require an active SaaS subscription.
	customers := router.Group("/customers")

	customers.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
		middleware.RequireRole("owner", "attendant"),
	)

	customers.POST(
		"",
		handler.Create,
	)

	customers.GET(
		"",
		handler.List,
	)

	customers.GET(
		"/:id",
		handler.Get,
	)

	customers.PUT(
		"/:id",
		handler.Update,
	)

	// Terminal customer API.
	// This must remain operational even when the web subscription
	// has expired because terminals continue operating locally.
	terminalCustomers := router.Group("/terminal-customers")

	terminalCustomers.Use(
		middleware.TerminalAuthMiddleware(
			terminalAuthenticator,
		),
	)

	terminalCustomers.POST(
		"/lookup",
		handler.TerminalLookup,
	)
}

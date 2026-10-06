package expenses

import (
	"github.com/gin-gonic/gin"

	"cybersaas/backend/auth"
	"cybersaas/backend/middleware"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	tokenManager *auth.TokenManager,
	requireActiveSubscription gin.HandlerFunc,
) {
	expenses := router.Group("/expenses")

	expenses.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
	)

	// Attendants can create expenses.
	expenses.POST(
		"",
		middleware.RequireRole("attendant"),
		handler.Create,
	)

	// Cyber Owners can view expenses for their tenant/branches.
	expenses.GET(
		"",
		middleware.RequireRole("owner"),
		handler.ListOwner,
	)

	// Attendants can view expenses for their assigned branch.
	expenses.GET(
		"/my-branch",
		middleware.RequireRole("attendant"),
		handler.ListAttendant,
	)

	// Both owner and attendant can view a specific expense,
	// subject to tenant/branch authorization in the handler/service.
	expenses.GET(
		"/:id",
		middleware.RequireRole("owner", "attendant"),
		handler.Get,
	)
}

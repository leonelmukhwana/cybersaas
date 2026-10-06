package attendant

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
	attendants := router.Group("/attendants")

	attendants.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
	)

	attendants.POST(
		"",
		middleware.RequireRole("owner"),
		handler.Create,
	)

	attendants.GET(
		"",
		middleware.RequireRole("owner"),
		handler.List,
	)

	attendants.GET(
		"/me",
		middleware.RequireRole("attendant"),
		handler.Me,
	)

	attendants.GET(
		"/:id",
		middleware.RequireRole("owner"),
		handler.Get,
	)

	attendants.PUT(
		"/:id",
		middleware.RequireRole("owner"),
		handler.Update,
	)

	attendants.PATCH(
		"/:id/status",
		middleware.RequireRole("owner"),
		handler.ChangeStatus,
	)

	attendants.PATCH(
		"/:id/branch",
		middleware.RequireRole("owner"),
		handler.AssignBranch,
	)

	attendants.DELETE(
		"/:id/branch",
		middleware.RequireRole("owner"),
		handler.UnassignBranch,
	)
}

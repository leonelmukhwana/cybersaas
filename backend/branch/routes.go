package branch

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
	branches := router.Group("/branches")

	branches.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
	)

	// Cyber Owner only.
	branches.POST(
		"",
		middleware.RequireRole("owner"),
		handler.Create,
	)

	branches.GET(
		"",
		middleware.RequireRole("owner"),
		handler.List,
	)

	branches.GET(
		"/:id",
		middleware.RequireRole("owner"),
		handler.Get,
	)

	branches.PUT(
		"/:id",
		middleware.RequireRole("owner"),
		handler.Update,
	)

	branches.PATCH(
		"/:id/status",
		middleware.RequireRole("owner"),
		handler.ChangeStatus,
	)
}

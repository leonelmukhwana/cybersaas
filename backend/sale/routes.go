package sale

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
	sales := router.Group("/sales")

	sales.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
		middleware.RequireRole(
			"owner",
			"attendant",
		),
	)

	sales.POST(
		"",
		handler.Create,
	)

	sales.GET(
		"",
		handler.List,
	)

	sales.GET(
		"/:id",
		handler.Get,
	)

	sales.POST(
		"/:id/void",
		middleware.RequireRole("owner"),
		handler.Void,
	)
}

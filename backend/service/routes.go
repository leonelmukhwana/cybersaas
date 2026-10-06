package service

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
	services := router.Group("/services")

	services.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
		middleware.RequireRole("owner", "attendant"),
	)

	// Owner + Attendant
	services.GET("", handler.List)
	services.GET("/:id", handler.Get)

	// Owner only
	services.POST(
		"",
		middleware.RequireRole("owner"),
		handler.Create,
	)

	services.PUT(
		"/:id",
		middleware.RequireRole("owner"),
		handler.Update,
	)

	services.PATCH(
		"/:id/status",
		middleware.RequireRole("owner"),
		handler.ChangeStatus,
	)
}

package payment

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
	payments := router.Group("/payments")

	payments.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
		middleware.RequireRole(
			"owner",
			"attendant",
		),
	)

	payments.POST(
		"",
		handler.Create,
	)

	payments.GET(
		"",
		handler.List,
	)

	payments.GET(
		"/:id",
		handler.Get,
	)

	payments.PATCH(
		"/:id/confirm",
		handler.Confirm,
	)

	payments.PATCH(
		"/:id/fail",
		handler.Fail,
	)

	payments.PATCH(
		"/:id/refund",
		middleware.RequireRole("owner"),
		handler.Refund,
	)
}

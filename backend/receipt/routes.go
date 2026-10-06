package receipt

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
	receipts := router.Group("/receipts")

	receipts.Use(
		middleware.AuthMiddleware(tokenManager),
		requireActiveSubscription,
		middleware.RequireRole(
			"owner",
			"attendant",
		),
	)

	receipts.POST(
		"",
		handler.Create,
	)

	receipts.GET(
		"",
		handler.List,
	)

	receipts.GET(
		"/:id",
		handler.Get,
	)

	receipts.POST(
		"/:id/print",
		handler.Print,
	)

	receipts.POST(
		"/:id/reprint",
		handler.Reprint,
	)
}

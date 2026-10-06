package tenant

import (
	"cybersaas/backend/auth"
	"cybersaas/backend/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	tokenManager *auth.TokenManager,
) {

	tenant := router.Group("/tenant")

	tenant.Use(
		middleware.AuthMiddleware(tokenManager),
	)

	// Create the business/tenant.
	tenant.POST(
		"",
		middleware.RequireRole("owner"),
		handler.Create,
	)

	// Get the currently authenticated owner's tenant.
	tenant.GET(
		"/me",
		middleware.RequireRole("owner"),
		handler.Me,
	)

	// Get tenant by ID, but service verifies ownership.
	tenant.GET(
		"/:tenantID",
		middleware.RequireRole("owner"),
		handler.Get,
	)

	// Update business name.
	tenant.PUT(
		"",
		middleware.RequireRole("owner"),
		handler.Update,
	)
}

package platform

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
	platform := router.Group("/platform")

	platform.Use(middleware.AuthMiddleware(tokenManager))
	platform.Use(middleware.RequireRole("platform_admin"))

	platform.GET("/dashboard", handler.Dashboard)

	platform.GET("/owners", handler.CyberOwners)

	platform.GET("/subscriptions", handler.Subscriptions)

	platform.GET("/revenue", handler.Revenue)

	platform.GET("/audit-logs", handler.AuditLogs)
}

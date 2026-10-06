package sync

import (
	"cybersaas/backend/middleware"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(
	router *gin.RouterGroup,
	handler *Handler,
	terminalAuthenticator middleware.TerminalAuthenticator,
) {
	syncAPI := router.Group("/sync")
	syncAPI.Use(
		middleware.TerminalAuthMiddleware(terminalAuthenticator),
	)

	syncAPI.POST("/events", handler.Push)
	syncAPI.GET("/events", handler.Pull)
	syncAPI.GET("/state", handler.State)
	syncAPI.POST("/successful", handler.Successful)
}

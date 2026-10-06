package sync

import (
	"net/http"
	"strconv"
	"time"

	"cybersaas/backend/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Push(c *gin.Context) {
	terminalID, ok := middleware.CurrentTerminalID(c)
	if !ok || terminalID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal authentication required",
		})
		return
	}

	tenantID, ok := middleware.CurrentTerminalTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal tenant is required",
		})
		return
	}

	branchID, ok := middleware.CurrentTerminalBranchID(c)
	if !ok || branchID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal branch is required",
		})
		return
	}

	var request PushEventRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	response, err := h.service.PushEvent(
		c.Request.Context(),
		tenantID,
		branchID,
		terminalID,
		request,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusAccepted, response)
}

func (h *Handler) Pull(c *gin.Context) {
	terminalID, ok := middleware.CurrentTerminalID(c)
	if !ok || terminalID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal authentication required",
		})
		return
	}

	tenantID, ok := middleware.CurrentTerminalTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal tenant is required",
		})
		return
	}

	branchID, ok := middleware.CurrentTerminalBranchID(c)
	if !ok || branchID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal branch is required",
		})
		return
	}

	var since *time.Time

	sinceValue := c.Query("since")

	if sinceValue != "" {
		parsed, err := time.Parse(time.RFC3339Nano, sinceValue)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid since timestamp",
			})
			return
		}

		since = &parsed
	}

	limit := 100

	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid limit",
			})
			return
		}

		limit = parsed
	}

	response, err := h.service.PullEvents(
		c.Request.Context(),
		tenantID,
		branchID,
		terminalID,
		since,
		limit,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) State(c *gin.Context) {
	terminalID, ok := middleware.CurrentTerminalID(c)
	if !ok || terminalID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal authentication required",
		})
		return
	}

	state, err := h.service.GetState(
		c.Request.Context(),
		terminalID,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, state)
}

func (h *Handler) Successful(c *gin.Context) {
	terminalID, ok := middleware.CurrentTerminalID(c)
	if !ok || terminalID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal authentication required",
		})
		return
	}

	if err := h.service.MarkSyncSuccessful(
		c.Request.Context(),
		terminalID,
	); err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status": "synced",
	})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch err {
	case ErrInvalidEventID,
		ErrInvalidEntityID,
		ErrInvalidEventType,
		ErrInvalidEntityType,
		ErrInvalidLimit:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "sync operation failed",
		})
	}
}

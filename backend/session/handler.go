package session

import (
	"errors"
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
	return &Handler{
		service: service,
	}
}

// ============================================================
// OWNER / ATTENDANT SESSION API
// ============================================================

func (h *Handler) Start(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user context is required",
		})
		return
	}

	role, ok := middleware.CurrentRole(c)
	if !ok || role == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user role is required",
		})
		return
	}

	var request StartSessionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	result, err := h.service.Start(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		request,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *Handler) Get(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user context is required",
		})
		return
	}

	role, ok := middleware.CurrentRole(c)
	if !ok || role == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user role is required",
		})
		return
	}

	branchID := c.Query("branch_id")

	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch_id is required",
		})
		return
	}

	result, err := h.service.Get(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		branchID,
		c.Param("id"),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) Preview(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user context is required",
		})
		return
	}

	role, ok := middleware.CurrentRole(c)
	if !ok || role == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user role is required",
		})
		return
	}

	branchID := c.Query("branch_id")

	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch_id is required",
		})
		return
	}

	at := time.Now().UTC()

	if value := c.Query("at"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid at timestamp",
			})
			return
		}

		at = parsed.UTC()
	}

	result, err := h.service.PreviewBilling(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		branchID,
		c.Param("id"),
		at,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) End(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user context is required",
		})
		return
	}

	role, ok := middleware.CurrentRole(c)
	if !ok || role == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user role is required",
		})
		return
	}

	branchID := c.Query("branch_id")

	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch_id is required",
		})
		return
	}

	var request EndSessionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	result, err := h.service.End(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		branchID,
		c.Param("id"),
		request,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) Cancel(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user context is required",
		})
		return
	}

	role, ok := middleware.CurrentRole(c)
	if !ok || role == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user role is required",
		})
		return
	}

	branchID := c.Query("branch_id")

	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch_id is required",
		})
		return
	}

	var request CancelSessionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	result, err := h.service.Cancel(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		branchID,
		c.Param("id"),
		request.Reason,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) List(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user context is required",
		})
		return
	}

	role, ok := middleware.CurrentRole(c)
	if !ok || role == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user role is required",
		})
		return
	}

	branchID := c.Query("branch_id")

	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch_id is required",
		})
		return
	}

	limit := 50
	offset := 0

	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid limit",
			})
			return
		}

		limit = parsed
	}

	if value := c.Query("offset"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid offset",
			})
			return
		}

		offset = parsed
	}

	result, err := h.service.List(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		branchID,
		c.Query("status"),
		limit,
		offset,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// ============================================================
// TERMINAL SESSION API
// ============================================================

// TerminalStart starts a session directly from an authenticated
// CyberSaaS terminal.
//
// Authentication comes from TerminalAuthMiddleware.
//
// The terminal does NOT provide:
//   - tenant_id
//   - branch_id
//   - terminal_id
//   - rate
//   - amount due
//
// Those values are controlled by the authenticated terminal and
// backend billing configuration.
func (h *Handler) TerminalStart(c *gin.Context) {
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
			"error": "terminal tenant context is required",
		})
		return
	}

	branchID, ok := middleware.CurrentTerminalBranchID(c)

	if !ok || branchID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal branch context is required",
		})
		return
	}

	var request TerminalStartSessionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	result, err := h.service.StartFromTerminal(
		c.Request.Context(),
		terminalID,
		tenantID,
		branchID,
		request,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "terminal session started successfully",
		"data":    result,
	})
}

func (h *Handler) TerminalGet(c *gin.Context) {
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
			"error": "terminal tenant context is required",
		})
		return
	}

	branchID, ok := middleware.CurrentTerminalBranchID(c)

	if !ok || branchID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal branch context is required",
		})
		return
	}

	result, err := h.service.GetFromTerminal(
		c.Request.Context(),
		terminalID,
		tenantID,
		branchID,
		c.Param("id"),
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

func (h *Handler) TerminalPreview(c *gin.Context) {
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
			"error": "terminal tenant context is required",
		})
		return
	}

	branchID, ok := middleware.CurrentTerminalBranchID(c)

	if !ok || branchID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal branch context is required",
		})
		return
	}

	at := time.Now().UTC()

	if value := c.Query("at"); value != "" {
		parsed, err := time.Parse(time.RFC3339, value)

		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid at timestamp",
			})
			return
		}

		at = parsed.UTC()
	}

	result, err := h.service.PreviewFromTerminal(
		c.Request.Context(),
		terminalID,
		tenantID,
		branchID,
		c.Param("id"),
		at,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

func (h *Handler) TerminalEnd(c *gin.Context) {
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
			"error": "terminal tenant context is required",
		})
		return
	}

	branchID, ok := middleware.CurrentTerminalBranchID(c)

	if !ok || branchID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal branch context is required",
		})
		return
	}

	var request TerminalEndSessionRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	result, err := h.service.EndFromTerminal(
		c.Request.Context(),
		terminalID,
		tenantID,
		branchID,
		c.Param("id"),
		request,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "terminal session ended successfully",
		"data":    result,
	})
}

// ============================================================
// ERROR HANDLING
// ============================================================

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrSessionNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "session not found",
		})

	case errors.Is(err, ErrBranchAccessDenied):
		c.JSON(http.StatusForbidden, gin.H{
			"error": "branch access denied",
		})

	case errors.Is(err, ErrTerminalNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "terminal not found or inactive",
		})

	case errors.Is(err, ErrCustomerNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "customer not found in this branch",
		})

	case errors.Is(err, ErrActiveSessionExists):
		c.JSON(http.StatusConflict, gin.H{
			"error": "terminal already has an active session",
		})

	case errors.Is(err, ErrInvalidSessionState):
		c.JSON(http.StatusConflict, gin.H{
			"error": "session is not active",
		})

	case errors.Is(err, ErrInvalidSessionType),
		errors.Is(err, ErrInvalidAmount),
		errors.Is(err, ErrInvalidMinutes),
		errors.Is(err, ErrPrepaidAmountMissing),
		errors.Is(err, ErrPrepaidMinutesMissing),
		errors.Is(err, ErrInvalidSessionID),
		errors.Is(err, ErrInvalidTerminalID),
		errors.Is(err, ErrInvalidCustomerID),
		errors.Is(err, ErrInvalidOperationID),
		errors.Is(err, ErrInvalidBranchID):

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrDuplicateOperation):
		c.JSON(http.StatusConflict, gin.H{
			"error": "client operation has already been processed",
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}

func (h *Handler) TerminalActive(c *gin.Context) {
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
			"error": "terminal tenant context missing",
		})
		return
	}

	branchID, ok := middleware.CurrentTerminalBranchID(c)
	if !ok || branchID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal branch context missing",
		})
		return
	}

	result, err := h.service.GetActiveFromTerminal(
		c.Request.Context(),
		terminalID,
		tenantID,
		branchID,
	)

	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			c.JSON(http.StatusOK, gin.H{
				"data": nil,
			})
			return
		}

		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

// ============================================================
// TERMINAL PAUSE / RESUME
// ============================================================

func (h *Handler) TerminalPause(c *gin.Context) {
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
			"error": "terminal tenant context is required",
		})
		return
	}

	branchID, ok := middleware.CurrentTerminalBranchID(c)
	if !ok || branchID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal branch context is required",
		})
		return
	}

	result, err := h.service.PauseFromTerminal(
		c.Request.Context(),
		terminalID,
		tenantID,
		branchID,
		c.Param("id"),
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "session paused successfully",
		"data":    result,
	})
}

func (h *Handler) TerminalResume(c *gin.Context) {
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
			"error": "terminal tenant context is required",
		})
		return
	}

	branchID, ok := middleware.CurrentTerminalBranchID(c)
	if !ok || branchID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal branch context is required",
		})
		return
	}

	result, err := h.service.ResumeFromTerminal(
		c.Request.Context(),
		terminalID,
		tenantID,
		branchID,
		c.Param("id"),
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "session resumed successfully",
		"data":    result,
	})
}

func (h *Handler) TerminalBillingConfig(c *gin.Context) {
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
			"error": "terminal tenant context is required",
		})
		return
	}

	branchID, ok := middleware.CurrentTerminalBranchID(c)

	if !ok || branchID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal branch context is required",
		})
		return
	}

	result, err := h.service.GetBillingConfigFromTerminal(
		c.Request.Context(),
		terminalID,
		tenantID,
		branchID,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

func (h *Handler) Pause(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user context is required",
		})
		return
	}

	role, ok := middleware.CurrentRole(c)
	if !ok || role == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user role is required",
		})
		return
	}

	branchID := c.Query("branch_id")

	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch_id is required",
		})
		return
	}

	result, err := h.service.Pause(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		branchID,
		c.Param("id"),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "session paused successfully",
		"data":    result,
	})
}

func (h *Handler) Resume(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user context is required",
		})
		return
	}

	role, ok := middleware.CurrentRole(c)
	if !ok || role == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user role is required",
		})
		return
	}

	branchID := c.Query("branch_id")

	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch_id is required",
		})
		return
	}

	result, err := h.service.Resume(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		branchID,
		c.Param("id"),
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "session resumed successfully",
		"data":    result,
	})
}

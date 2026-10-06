package terminal

import (
	"net/http"
	"strconv"
	"strings"

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
// PUBLIC PC SELF-REGISTRATION
// ============================================================

// Register is called by the C# PC client.
//
// No owner JWT is required.
//
// The licence key determines the tenant and branch.
func (h *Handler) Register(c *gin.Context) {
	var req RegisterTerminalRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	result, err := h.service.RegisterTerminal(
		c.Request.Context(),
		req,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "terminal registered successfully",
		"data":    result,
	})
}

// ============================================================
// TERMINAL AUTHENTICATED API
// ============================================================

// Heartbeat is called by the installed C# PC client.
//
// Authentication is handled by TerminalAuthMiddleware.
//
// The terminal sends its current machine information and the
// server responds with the latest desired lock/unlock state.
//
// This endpoint does NOT check SaaS subscription status.
// Subscription expiry must never automatically lock a terminal.
func (h *Handler) Heartbeat(c *gin.Context) {
	terminalID, ok := middleware.CurrentTerminalID(c)

	if !ok || terminalID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal authentication required",
		})
		return
	}

	var req TerminalHeartbeatRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid heartbeat request",
			"details": err.Error(),
		})
		return
	}

	result, err := h.service.Heartbeat(
		c.Request.Context(),
		terminalID,
		req,
	)

	if err != nil {
		// Temporary diagnostic response so we can see the
		// actual repository/database error.
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to process heartbeat",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

// ControlState returns the latest lock/unlock command for the
// authenticated terminal.
//
// The C# client can use this endpoint independently from
// heartbeat polling if required.
func (h *Handler) ControlState(c *gin.Context) {
	terminalID, ok := middleware.CurrentTerminalID(c)

	if !ok || terminalID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal authentication required",
		})
		return
	}

	result, err := h.service.GetControlState(
		c.Request.Context(),
		terminalID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load terminal control state",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

// ============================================================
// OWNER LICENCE KEY MANAGEMENT
// ============================================================

func (h *Handler) GenerateLicenceKey(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)

	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	var req GenerateLicenceKeyRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	result, err := h.service.GenerateLicenceKey(
		c.Request.Context(),
		tenantID,
		userID,
		req,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "licence key generated successfully",
		"data":    result,
	})
}

func (h *Handler) ListLicenceKeys(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	limit, offset := parsePagination(c)

	result, err := h.service.ListLicenceKeys(
		c.Request.Context(),
		tenantID,
		limit,
		offset,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load licence keys",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) RevokeLicenceKey(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	keyID := c.Param("id")

	if err := h.service.RevokeLicenceKey(
		c.Request.Context(),
		tenantID,
		keyID,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "licence key revoked successfully",
	})
}

// ============================================================
// OWNER TERMINAL MANAGEMENT
// ============================================================

func (h *Handler) List(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	limit, offset := parsePagination(c)

	result, err := h.service.ListTerminals(
		c.Request.Context(),
		tenantID,
		limit,
		offset,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load terminals",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) Get(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	terminalID := c.Param("id")

	result, err := h.service.GetTerminal(
		c.Request.Context(),
		tenantID,
		terminalID,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

func (h *Handler) Rename(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	var req RenameTerminalRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	err := h.service.RenameTerminal(
		c.Request.Context(),
		tenantID,
		c.Param("id"),
		req,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "terminal renamed successfully",
	})
}

func (h *Handler) ChangeStatus(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	var req ChangeStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	err := h.service.ChangeStatus(
		c.Request.Context(),
		tenantID,
		c.Param("id"),
		req,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "terminal status updated successfully",
	})
}

func (h *Handler) Move(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	var req MoveTerminalRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	err := h.service.MoveTerminal(
		c.Request.Context(),
		tenantID,
		c.Param("id"),
		req,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "terminal moved successfully",
	})
}

// ============================================================
// OWNER TERMINAL LOCK / UNLOCK
// ============================================================

// SetLockState allows the Cyber Owner to request a terminal
// to lock or unlock.
//
// This endpoint uses the owner's JWT because it is a dashboard
// operation.
//
// It only changes the desired terminal state.
//
// It does NOT directly control the Windows PC. The C# client
// receives the state through heartbeat/control-state polling.
func (h *Handler) SetLockState(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	terminalID := c.Param("id")

	if terminalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "terminal id is required",
		})
		return
	}

	var req ChangeTerminalLockStateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid lock state request",
		})
		return
	}

	result, err := h.service.SetControlState(
		c.Request.Context(),
		tenantID,
		terminalID,
		req.State,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "terminal lock state updated successfully",
		"data":    result,
	})
}

// ============================================================
// ATTENDANT TERMINAL MANAGEMENT
// ============================================================

// ListAttendantTerminals returns only terminals belonging to
// the branch currently assigned to the authenticated attendant.
//
// Attendants can view terminals but cannot register, rename,
// move, disable, or delete them.
func (h *Handler) ListAttendantTerminals(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)

	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	limit, offset := parsePagination(c)

	result, err := h.service.ListAttendantTerminals(
		c.Request.Context(),
		tenantID,
		userID,
		limit,
		offset,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetAttendantTerminal returns one terminal only when it belongs
// to the attendant's currently assigned branch.
func (h *Handler) GetAttendantTerminal(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)

	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	terminalID := c.Param("id")

	if terminalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "terminal id is required",
		})
		return
	}

	result, err := h.service.GetAttendantTerminal(
		c.Request.Context(),
		tenantID,
		userID,
		terminalID,
	)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": result,
	})
}

// SetAttendantLockState allows an attendant to lock or unlock
// a terminal belonging to their assigned branch.
//
// It does not directly control Windows. The terminal client
// receives the desired state through its authenticated polling.
func (h *Handler) SetAttendantLockState(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)

	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	terminalID := c.Param("id")

	if terminalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "terminal id is required",
		})
		return
	}

	var req ChangeTerminalLockStateRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid lock state request",
		})
		return
	}

	result, err := h.service.SetAttendantControlState(
		c.Request.Context(),
		tenantID,
		userID,
		terminalID,
		req.State,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "terminal lock state updated successfully",
		"data":    result,
	})
}

// RestartAttendantTerminal queues a restart command for a
// terminal belonging to the attendant's assigned branch.
func (h *Handler) RestartAttendantTerminal(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)

	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	terminalID := c.Param("id")

	if terminalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "terminal id is required",
		})
		return
	}

	result, err := h.service.QueueAttendantCommand(
		c.Request.Context(),
		tenantID,
		userID,
		terminalID,
		CommandRestart,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "terminal restart command queued successfully",
		"data":    result,
	})
}

// ShutdownAttendantTerminal queues a shutdown command for a
// terminal belonging to the attendant's assigned branch.
func (h *Handler) ShutdownAttendantTerminal(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)

	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	terminalID := c.Param("id")

	if terminalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "terminal id is required",
		})
		return
	}

	result, err := h.service.QueueAttendantCommand(
		c.Request.Context(),
		tenantID,
		userID,
		terminalID,
		CommandShutdown,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "terminal shutdown command queued successfully",
		"data":    result,
	})
}

// deregister
func (h *Handler) Deregister(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	terminalID := c.Param("id")

	if terminalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "terminal id is required",
		})
		return
	}

	err := h.service.DeregisterTerminal(
		c.Request.Context(),
		tenantID,
		terminalID,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "terminal deregistered successfully",
	})
}

func (h *Handler) Authenticate(c *gin.Context) {
	var req TerminalAuthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	result, err := h.service.AuthenticateTerminalRequest(
		c.Request.Context(),
		req,
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "terminal authenticated successfully",
		"data":    result,
	})
}

// ============================================================
// TERMINAL POWER COMMANDS
// ============================================================

// Restart requests the authenticated terminal to restart Windows.
//
// The endpoint only creates a pending command.
// The C# client receives the command through its polling/heartbeat
// mechanism and performs the actual restart.
func (h *Handler) Restart(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	terminalID := c.Param("id")

	if terminalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "terminal id is required",
		})
		return
	}

	result, err := h.service.QueueCommand(
		c.Request.Context(),
		tenantID,
		terminalID,
		CommandRestart,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "terminal restart command queued successfully",
		"data":    result,
	})
}

// Shutdown requests the authenticated terminal to shut down Windows.
//
// The endpoint only creates a pending command.
// The C# client performs the actual shutdown.
func (h *Handler) Shutdown(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	terminalID := c.Param("id")

	if terminalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "terminal id is required",
		})
		return
	}

	result, err := h.service.QueueCommand(
		c.Request.Context(),
		tenantID,
		terminalID,
		CommandShutdown,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "terminal shutdown command queued successfully",
		"data":    result,
	})
}

func (h *Handler) PendingCommand(c *gin.Context) {
	terminalID, ok := middleware.CurrentTerminalID(c)
	if !ok || terminalID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal context is required",
		})
		return
	}

	command, err := h.service.GetPendingCommand(
		c.Request.Context(),
		terminalID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": command,
	})
}

func (h *Handler) AcknowledgeCommand(c *gin.Context) {
	terminalID, ok := middleware.CurrentTerminalID(c)
	if !ok || terminalID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal context is required",
		})
		return
	}

	commandID := c.Param("id")

	if commandID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "command id is required",
		})
		return
	}

	var req struct {
		Status       string  `json:"status"`
		ErrorMessage *string `json:"error_message,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	status := strings.ToLower(strings.TrimSpace(req.Status))

	if status != "executed" && status != "failed" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "status must be executed or failed",
		})
		return
	}

	result, err := h.service.AcknowledgeCommand(
		c.Request.Context(),
		terminalID,
		commandID,
		status,
		req.ErrorMessage,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "terminal command acknowledged successfully",
		"data":    result,
	})
}

// ============================================================
// PAGINATION
// ============================================================

func parsePagination(c *gin.Context) (int, int) {
	limit := 50
	offset := 0

	if value := c.Query("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			limit = parsed
		}
	}

	if value := c.Query("offset"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			offset = parsed
		}
	}

	return limit, offset
}

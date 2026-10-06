package terminalhealth

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"

	"cybersaas/backend/middleware"
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
// TERMINAL CLIENT
// ============================================================

// ReportHealth receives a health report from an authenticated
// terminal.
//
// The terminal ID is taken from terminal authentication context.
// It is never trusted from the request body.
func (h *Handler) ReportHealth(c *gin.Context) {
	terminalID := c.GetString("terminal_id")

	if terminalID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal authentication required",
		})
		return
	}

	var req ReportHealthRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid health report",
		})
		return
	}

	health, err := h.service.ReportHealth(
		c.Request.Context(),
		terminalID,
		req,
	)

	if err != nil {
		switch err {
		case ErrInvalidTerminalID:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return

		case ErrInvalidHealth:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to store terminal health",
		})
		return
	}

	c.JSON(http.StatusOK, TerminalHealthResponse{
		Health: *health,
	})
}

// ============================================================
// OWNER - SINGLE TERMINAL
// ============================================================

func (h *Handler) GetTerminalHealth(c *gin.Context) {
	tenantID := c.GetString(middleware.ContextTenantID)

	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context required",
		})
		return
	}

	terminalID := c.Param("terminalID")

	if terminalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "terminal id is required",
		})
		return
	}

	health, err := h.service.GetTerminalHealth(
		c.Request.Context(),
		tenantID,
		terminalID,
	)

	if err != nil {
		if err.Error() == "terminal health not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "terminal health not found",
			})
			return
		}

		if err.Error() == "invalid tenant id" ||
			err == ErrInvalidTerminalID {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve terminal health",
		})
		return
	}

	c.JSON(http.StatusOK, TerminalHealthResponse{
		Health: *health,
	})
}

// ============================================================
// OWNER - BRANCH TERMINALS
// ============================================================

func (h *Handler) ListBranchHealth(c *gin.Context) {
	tenantID := c.GetString(middleware.ContextTenantID)

	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context required",
		})
		return
	}

	branchID := c.Param("branchID")

	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch id is required",
		})
		return
	}

	terminals, total, err := h.service.ListBranchHealth(
		c.Request.Context(),
		tenantID,
		branchID,
	)

	if err != nil {
		if err.Error() == "invalid tenant id" ||
			err.Error() == "invalid branch id" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve branch terminal health",
		})
		return
	}

	c.JSON(http.StatusOK, TerminalHealthListResponse{
		Terminals: terminals,
		Total:     total,
	})
}

func (h *Handler) GetAttendantHealthSummary(c *gin.Context) {
	tenantID := c.GetString(middleware.ContextTenantID)

	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context required",
		})
		return
	}

	attendantID := c.GetString(middleware.ContextUserID)

	if attendantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "attendant context required",
		})
		return
	}

	summary, err := h.service.GetAttendantHealthSummary(
		c.Request.Context(),
		tenantID,
		attendantID,
	)

	if err != nil {
		if err.Error() == "invalid tenant id" ||
			err.Error() == "invalid attendant id" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve attendant health summary",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"summary": summary,
	})
}

func (h *Handler) ListAttendantTerminals(c *gin.Context) {
	tenantID := c.GetString(middleware.ContextTenantID)
	attendantID := c.GetString(middleware.ContextUserID)

	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context required",
		})
		return
	}

	if attendantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "attendant context required",
		})
		return
	}

	terminals, err := h.service.ListAttendantTerminals(
		c.Request.Context(),
		tenantID,
		attendantID,
	)
	if err != nil {
		if err.Error() == "invalid tenant id" ||
			err.Error() == "invalid attendant id" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to retrieve attendant terminals",
		})
		return
	}

	c.JSON(http.StatusOK, AttendantTerminalListResponse{
		Terminals: terminals,
		Total:     int64(len(terminals)),
	})
}
func (h *Handler) GetAttendantTerminalHealth(c *gin.Context) {
	tenantID := c.GetString(middleware.ContextTenantID)
	attendantID := c.GetString(middleware.ContextUserID)
	terminalID := c.Param("terminal_id")

	health, err := h.service.GetAttendantTerminalHealth(
		c.Request.Context(),
		tenantID,
		attendantID,
		terminalID,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "terminal health not found",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, TerminalHealthResponse{
		Health: *health,
	})
}

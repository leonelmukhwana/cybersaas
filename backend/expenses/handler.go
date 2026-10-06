package expenses

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

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

func (h *Handler) Create(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
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

	var request CreateExpenseRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	expense, err := h.service.Create(
		c.Request.Context(),
		tenantID,
		userID,
		request,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, expense)
}

func (h *Handler) ListOwner(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	branchID := strings.TrimSpace(
		c.Query("branch_id"),
	)

	expenses, total, err := h.service.ListForOwner(
		c.Request.Context(),
		tenantID,
		branchID,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ExpenseListResponse{
		Expenses: expenses,
		Total:    total,
	})
}

func (h *Handler) ListAttendant(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
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

	branchID := strings.TrimSpace(
		c.Query("branch_id"),
	)

	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch_id is required",
		})
		return
	}

	expenses, total, err := h.service.ListForAttendant(
		c.Request.Context(),
		tenantID,
		userID,
		branchID,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, ExpenseListResponse{
		Expenses: expenses,
		Total:    total,
	})
}

func (h *Handler) Get(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	id := strings.TrimSpace(
		c.Param("id"),
	)

	expense, err := h.service.Get(
		c.Request.Context(),
		tenantID,
		id,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, expense)
}

func (h *Handler) handleError(
	c *gin.Context,
	err error,
) {
	switch {
	case errors.Is(err, ErrInvalidTenantID),
		errors.Is(err, ErrInvalidUserID),
		errors.Is(err, ErrInvalidBranchID),
		errors.Is(err, ErrInvalidExpenseID),
		errors.Is(err, ErrInvalidCategory),
		errors.Is(err, ErrInvalidAmount),
		errors.Is(err, ErrInvalidPayment),
		errors.Is(err, ErrInvalidExpenseDate):

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	case errors.Is(err, ErrBranchAccess):
		c.JSON(http.StatusForbidden, gin.H{
			"error": "branch access denied",
		})

	case errors.Is(err, ErrNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "expense not found",
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}

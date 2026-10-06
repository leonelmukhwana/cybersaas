package payment

import (
	"errors"
	"net/http"
	"strconv"

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

func (h *Handler) Create(c *gin.Context) {
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

	var request CreatePaymentRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	payment, err := h.service.Create(
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

	c.JSON(http.StatusCreated, payment)
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

	payment, err := h.service.Get(
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

	c.JSON(http.StatusOK, payment)
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

func (h *Handler) Confirm(c *gin.Context) {
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

	var request ConfirmPaymentRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	payment, err := h.service.Confirm(
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

	c.JSON(http.StatusOK, payment)
}

func (h *Handler) Fail(c *gin.Context) {
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

	var request FailPaymentRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := h.service.Fail(
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

	c.JSON(http.StatusOK, gin.H{
		"message": "payment marked as failed",
	})
}

func (h *Handler) Refund(c *gin.Context) {
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

	var request RefundPaymentRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := h.service.Refund(
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

	c.JSON(http.StatusOK, gin.H{
		"message": "payment refunded successfully",
	})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrPaymentNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "payment not found",
		})

	case errors.Is(err, ErrSaleNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "sale not found",
		})

	case errors.Is(err, ErrBranchAccessDenied):
		c.JSON(http.StatusForbidden, gin.H{
			"error": "branch access denied",
		})

	case errors.Is(err, ErrPaymentAlreadyFinal):
		c.JSON(http.StatusConflict, gin.H{
			"error": "payment is already in a final state",
		})

	case errors.Is(err, ErrDuplicatePayment):
		c.JSON(http.StatusConflict, gin.H{
			"error": "duplicate payment operation",
		})

	case errors.Is(err, ErrPaymentReferenceUsed):
		c.JSON(http.StatusConflict, gin.H{
			"error": "payment reference already exists",
		})

	case errors.Is(err, ErrInvalidPaymentID),
		errors.Is(err, ErrInvalidSaleID),
		errors.Is(err, ErrInvalidBranchID),
		errors.Is(err, ErrInvalidAmount),
		errors.Is(err, ErrInvalidMethod),
		errors.Is(err, ErrInvalidStatus):

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}

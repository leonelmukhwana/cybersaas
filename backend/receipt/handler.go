package receipt

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

type CreateReceiptRequest struct {
	BranchID  string  `json:"branch_id" binding:"required"`
	SaleID    string  `json:"sale_id" binding:"required"`
	PaymentID *string `json:"payment_id"`
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

	var request CreateReceiptRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	receipt, err := h.service.Create(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		request.BranchID,
		request.SaleID,
		request.PaymentID,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, receipt)
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

	receipt, err := h.service.Get(
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

	c.JSON(http.StatusOK, receipt)
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
		limit,
		offset,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) Print(c *gin.Context) {
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

	receipt, err := h.service.MarkPrinted(
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

	c.JSON(http.StatusOK, receipt)
}

func (h *Handler) Reprint(c *gin.Context) {
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

	var request ReprintReceiptRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	receipt, err := h.service.Reprint(
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

	c.JSON(http.StatusOK, receipt)
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrReceiptNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "receipt not found",
		})

	case errors.Is(err, ErrSaleNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "sale not found",
		})

	case errors.Is(err, ErrPaymentNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "payment not found",
		})

	case errors.Is(err, ErrPaymentNotConfirmed):
		c.JSON(http.StatusConflict, gin.H{
			"error": "receipt can only be created for a confirmed payment",
		})

	case errors.Is(err, ErrReceiptAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{
			"error": "receipt already exists for this sale",
		})

	case errors.Is(err, ErrBranchAccessDenied):
		c.JSON(http.StatusForbidden, gin.H{
			"error": "branch access denied",
		})

	case errors.Is(err, ErrInvalidReceiptID),
		errors.Is(err, ErrInvalidBranchID):

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}

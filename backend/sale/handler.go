package sale

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

	var request CreateSaleRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	sale, err := h.service.Create(
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

	c.JSON(http.StatusCreated, sale)
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
	saleID := c.Param("id")

	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch_id is required",
		})
		return
	}

	sale, err := h.service.Get(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		branchID,
		saleID,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, sale)
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

	status := c.DefaultQuery("status", "")

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
		status,
		limit,
		offset,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) Void(c *gin.Context) {
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
	saleID := c.Param("id")

	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch_id is required",
		})
		return
	}

	var request VoidSaleRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	err := h.service.Void(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		branchID,
		saleID,
		request.Reason,
	)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "sale voided successfully",
	})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrSaleNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "sale not found",
		})

	case errors.Is(err, ErrBranchAccessDenied):
		c.JSON(http.StatusForbidden, gin.H{
			"error": "branch access denied",
		})

	case errors.Is(err, ErrBranchNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "branch not found",
		})

	case errors.Is(err, ErrSaleAlreadyVoided):
		c.JSON(http.StatusConflict, gin.H{
			"error": "sale is already voided",
		})

	case errors.Is(err, ErrInvalidSaleID),
		errors.Is(err, ErrInvalidBranchID),
		errors.Is(err, ErrInvalidCustomerID),
		errors.Is(err, ErrInvalidSessionID),
		errors.Is(err, ErrInvalidTerminalID),
		errors.Is(err, ErrInvalidAttendantID),
		errors.Is(err, ErrInvalidSaleItem),
		errors.Is(err, ErrInvalidQuantity),
		errors.Is(err, ErrInvalidPrice),
		errors.Is(err, ErrInvalidDiscount),
		errors.Is(err, ErrInvalidDiscountType),
		errors.Is(err, ErrNoSaleItems):

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}

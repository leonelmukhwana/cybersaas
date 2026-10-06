package customer

import (
	"errors"
	"net/http"
	"strconv"

	"cybersaas/backend/middleware"

	"github.com/gin-gonic/gin"
)

const clientOperationIDHeader = "X-Client-Operation-ID"

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
		c.JSON(http.StatusUnauthorized, gin.H{
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

	clientOperationID := c.GetHeader(clientOperationIDHeader)

	if clientOperationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "X-Client-Operation-ID header is required",
		})
		return
	}

	var request CreateCustomerRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	customer, err := h.service.Create(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		clientOperationID,
		request,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"customer": customer,
	})
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
		c.JSON(http.StatusUnauthorized, gin.H{
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

	customerID := c.Param("id")
	branchID := c.Query("branch_id")

	if customerID == "" || branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "customer id and branch_id are required",
		})
		return
	}

	customer, err := h.service.Get(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		customerID,
		branchID,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customer": customer,
	})
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
		c.JSON(http.StatusUnauthorized, gin.H{
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

	search := c.Query("search")

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

	response, err := h.service.List(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		branchID,
		search,
		limit,
		offset,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

func (h *Handler) Update(c *gin.Context) {
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

	role, ok := middleware.CurrentRole(c)

	if !ok || role == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "user role is required",
		})
		return
	}

	customerID := c.Param("id")
	branchID := c.Query("branch_id")

	if customerID == "" || branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "customer id and branch_id are required",
		})
		return
	}

	clientOperationID := c.GetHeader(clientOperationIDHeader)

	if clientOperationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "X-Client-Operation-ID header is required",
		})
		return
	}

	var request UpdateCustomerRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	customer, err := h.service.Update(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		customerID,
		branchID,
		clientOperationID,
		request,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customer": customer,
	})
}

func (h *Handler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrCustomerNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "customer not found",
		})

	case errors.Is(err, ErrBranchNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "branch not found",
		})

	case errors.Is(err, ErrBranchAccessDenied):
		c.JSON(http.StatusForbidden, gin.H{
			"error": "you do not have access to this branch",
		})

	case errors.Is(err, ErrCustomerAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{
			"error": "customer is already registered at this branch",
		})

	default:
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
	}
}

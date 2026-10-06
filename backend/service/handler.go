package service

import (
	"errors"
	"net/http"
	"strconv"

	"cybersaas/backend/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *ServiceManager
}

func NewHandler(service *ServiceManager) *Handler {
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

	var request CreateServiceRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	service, err := h.service.Create(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		request,
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusCreated, service)
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

	branchID := c.Query("branch_id")

	service, err := h.service.Get(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		c.Param("id"),
		branchID,
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, service)
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

	activeOnly := true

	if value := c.Query("active_only"); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "active_only must be true or false",
			})
			return
		}

		activeOnly = parsed
	}

	limit := 50

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

	offset := 0

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
		activeOnly,
		limit,
		offset,
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
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

	var request UpdateServiceRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	branchID := c.Query("branch_id")

	service, err := h.service.Update(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		c.Param("id"),
		branchID,
		request,
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, service)
}

func (h *Handler) ChangeStatus(c *gin.Context) {
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

	var request ChangeStatusRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	branchID := c.Query("branch_id")

	service, err := h.service.ChangeStatus(
		c.Request.Context(),
		tenantID,
		userID,
		role,
		c.Param("id"),
		branchID,
		request.Active,
	)
	if err != nil {
		handleServiceError(c, err)
		return
	}

	c.JSON(http.StatusOK, service)
}

func handleServiceError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrServiceNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "service not found",
		})

	case errors.Is(err, ErrServiceAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{
			"error": "a service with this name already exists in this branch",
		})

	case errors.Is(err, ErrBranchNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": "branch not found",
		})

	case errors.Is(err, ErrBranchAccessDenied):
		c.JSON(http.StatusForbidden, gin.H{
			"error": "branch access denied",
		})

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
	}
}

package tenant

import (
	"errors"
	"net/http"

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

// ------------------------------------------------------------
// CREATE TENANT
// ------------------------------------------------------------

func (h *Handler) Create(c *gin.Context) {

	userID, ok := middleware.CurrentUserID(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	role, ok := middleware.CurrentRole(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "role context is required",
		})
		return
	}

	if role != "owner" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "only a cyber owner can create a tenant",
		})
		return
	}

	var request CreateTenantRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tenant, err := h.service.CreateTenant(
		c.Request.Context(),
		userID,
		request,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrTenantAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{
				"error": "owner already has a tenant",
			})

		case errors.Is(err, ErrOwnerNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "owner not found",
			})

		case errors.Is(err, ErrNotOwner):
			c.JSON(http.StatusForbidden, gin.H{
				"error": "user is not a cyber owner",
			})

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}

		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "tenant created successfully",
		"tenant":  tenant,
	})
}

// ------------------------------------------------------------
// GET MY TENANT
// ------------------------------------------------------------

func (h *Handler) Me(c *gin.Context) {

	userID, ok := middleware.CurrentUserID(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	tenant, err := h.service.GetMyTenant(
		c.Request.Context(),
		userID,
	)

	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "tenant has not been created yet",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant": tenant,
	})
}

// ------------------------------------------------------------
// GET TENANT BY ID
// ------------------------------------------------------------

func (h *Handler) Get(c *gin.Context) {

	userID, ok := middleware.CurrentUserID(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	tenantID := c.Param("tenantID")

	tenant, err := h.service.GetTenant(
		c.Request.Context(),
		userID,
		tenantID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrTenantNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "tenant not found",
			})

		case errors.Is(err, ErrInvalidTenantID):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid tenant id",
			})

		default:
			c.JSON(http.StatusForbidden, gin.H{
				"error": err.Error(),
			})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tenant": tenant,
	})
}

// ------------------------------------------------------------
// UPDATE
// ------------------------------------------------------------

func (h *Handler) Update(c *gin.Context) {

	userID, ok := middleware.CurrentUserID(c)

	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	var request UpdateTenantRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	tenant, err := h.service.UpdateTenant(
		c.Request.Context(),
		userID,
		request,
	)

	if err != nil {
		if errors.Is(err, ErrTenantNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "tenant not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "tenant updated successfully",
		"tenant":  tenant,
	})
}

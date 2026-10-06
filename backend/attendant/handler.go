package attendant

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
	return &Handler{service: service}
}

func tenantID(c *gin.Context) (string, bool) {
	id, ok := middleware.CurrentTenantID(c)

	if !ok || id == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return "", false
	}

	return id, true
}

func ownerID(c *gin.Context) (string, bool) {
	id, ok := middleware.CurrentUserID(c)

	if !ok || id == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return "", false
	}

	return id, true
}

func (h *Handler) Create(c *gin.Context) {
	tenant, ok := tenantID(c)
	if !ok {
		return
	}

	owner, ok := ownerID(c)
	if !ok {
		return
	}

	var req CreateAttendantRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	attendant, err := h.service.Create(
		c.Request.Context(),
		tenant,
		owner,
		req,
	)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, attendant)
}

// get the branch by attendant
func (h *Handler) Me(c *gin.Context) {
	tenant, ok := tenantID(c)
	if !ok {
		return
	}

	userID, ok := middleware.CurrentUserID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	attendant, err := h.service.Get(
		c.Request.Context(),
		tenant,
		userID,
	)

	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "attendant not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load attendant",
		})
		return
	}

	c.JSON(http.StatusOK, attendant)
}

func (h *Handler) List(c *gin.Context) {
	tenant, ok := tenantID(c)
	if !ok {
		return
	}

	result, err := h.service.List(
		c.Request.Context(),
		tenant,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load attendants",
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handler) Get(c *gin.Context) {
	tenant, ok := tenantID(c)
	if !ok {
		return
	}

	id := c.Param("id")

	attendant, err := h.service.Get(
		c.Request.Context(),
		tenant,
		id,
	)

	if errors.Is(err, ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "attendant not found",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load attendant",
		})
		return
	}

	c.JSON(http.StatusOK, attendant)
}

func (h *Handler) Update(c *gin.Context) {
	tenant, ok := tenantID(c)
	if !ok {
		return
	}

	owner, ok := ownerID(c)
	if !ok {
		return
	}

	id := c.Param("id")

	var req UpdateAttendantRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	if err := h.service.Update(
		c.Request.Context(),
		tenant,
		owner,
		id,
		req,
	); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "attendant not found",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "attendant updated successfully",
	})
}

func (h *Handler) ChangeStatus(c *gin.Context) {
	tenant, ok := tenantID(c)
	if !ok {
		return
	}

	owner, ok := ownerID(c)
	if !ok {
		return
	}

	id := c.Param("id")

	var req ChangeStatusRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	if err := h.service.ChangeStatus(
		c.Request.Context(),
		tenant,
		owner,
		id,
		req.Status,
	); err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "attendant not found",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "attendant status updated successfully",
	})
}

func (h *Handler) AssignBranch(c *gin.Context) {
	tenant, ok := tenantID(c)
	if !ok {
		return
	}

	owner, ok := ownerID(c)
	if !ok {
		return
	}

	id := c.Param("id")

	var req AssignBranchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	if err := h.service.AssignBranch(
		c.Request.Context(),
		tenant,
		owner,
		id,
		req.BranchID,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "attendant assigned to branch successfully",
	})
}

func (h *Handler) UnassignBranch(c *gin.Context) {
	tenant, ok := tenantID(c)
	if !ok {
		return
	}

	owner, ok := ownerID(c)
	if !ok {
		return
	}

	id := c.Param("id")

	if err := h.service.UnassignBranch(
		c.Request.Context(),
		tenant,
		owner,
		id,
	); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "attendant unassigned successfully",
	})
}

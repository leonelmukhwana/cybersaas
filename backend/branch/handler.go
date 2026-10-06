package branch

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

// -----------------------------------------------------------------------------
// CREATE
// -----------------------------------------------------------------------------

func (h *Handler) Create(c *gin.Context) {
	userID, userOK := middleware.CurrentUserID(c)

	tenantID, hasTenant := middleware.CurrentTenantID(c)

	if !userOK || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	var req CreateBranchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	ip := c.ClientIP()

	var ipAddress *string
	if ip != "" {
		ipAddress = &ip
	}

	// The tenant may be absent for a newly registered owner.
	// CreateBranch handles first-time tenant onboarding.
	_ = tenantID
	_ = hasTenant

	b, err := h.service.CreateBranch(
		c.Request.Context(),
		userID,
		req.BusinessName,
		req.Name,
		req.Address,
		req.Phone,
		ipAddress,
	)

	if err != nil {
		status := http.StatusBadRequest

		if errors.Is(err, ErrAlreadyExists) {
			status = http.StatusConflict
		}

		c.JSON(status, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"branch": b,
	})
}

// -----------------------------------------------------------------------------
// LIST
// -----------------------------------------------------------------------------

func (h *Handler) List(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	branches, err := h.service.ListBranches(
		c.Request.Context(),
		tenantID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load branches",
		})
		return
	}

	c.JSON(http.StatusOK, BranchListResponse{
		Branches: branches,
		Total:    int64(len(branches)),
	})
}

// -----------------------------------------------------------------------------
// GET
// -----------------------------------------------------------------------------

func (h *Handler) Get(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	branchID := c.Param("id")

	b, err := h.service.GetBranch(
		c.Request.Context(),
		tenantID,
		branchID,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "branch not found",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"branch": b,
	})
}

// -----------------------------------------------------------------------------
// UPDATE
// -----------------------------------------------------------------------------

func (h *Handler) Update(c *gin.Context) {
	userID, userOK := middleware.CurrentUserID(c)
	if !userOK || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	tenantID, tenantOK := middleware.CurrentTenantID(c)
	if !tenantOK || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	var req UpdateBranchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	ip := c.ClientIP()

	var ipAddress *string
	if ip != "" {
		ipAddress = &ip
	}

	b, err := h.service.UpdateBranch(
		c.Request.Context(),
		userID,
		tenantID,
		c.Param("id"),
		req.Name,
		req.Address,
		req.Phone,
		ipAddress,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "branch not found",
			})
			return
		}

		if errors.Is(err, ErrAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{
				"error": "a branch with that name already exists",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"branch": b,
	})
}

// -----------------------------------------------------------------------------
// STATUS
// -----------------------------------------------------------------------------

func (h *Handler) ChangeStatus(c *gin.Context) {
	userID, userOK := middleware.CurrentUserID(c)
	if !userOK || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authentication required",
		})
		return
	}

	tenantID, tenantOK := middleware.CurrentTenantID(c)
	if !tenantOK || tenantID == "" {
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

	ip := c.ClientIP()

	var ipAddress *string
	if ip != "" {
		ipAddress = &ip
	}

	b, err := h.service.ChangeStatus(
		c.Request.Context(),
		userID,
		tenantID,
		c.Param("id"),
		req.Status,
		ipAddress,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "branch not found",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"branch": b,
	})
}

package subscription

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

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
// GET /api/subscription/plans
// ------------------------------------------------------------

func (h *Handler) Plans(c *gin.Context) {
	plans, err := h.service.GetPlans(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load subscription plans",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plans": plans,
	})
}

// ------------------------------------------------------------
// GET /api/subscription
// ------------------------------------------------------------

func (h *Handler) Overview(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	overview, err := h.service.GetOverview(
		c.Request.Context(),
		tenantID,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "subscription not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load subscription",
		})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// ------------------------------------------------------------
// POST /api/subscription
// ------------------------------------------------------------

type CreateSubscriptionRequest struct {
	PlanID string `json:"plan_id" binding:"required"`
}

func (h *Handler) Create(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	var req CreateSubscriptionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "plan_id is required",
		})
		return
	}

	subscription, err := h.service.CreateSubscription(
		c.Request.Context(),
		tenantID,
		req.PlanID,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "subscription plan not found",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, subscription)
}

// ------------------------------------------------------------
// GET /api/subscription/payments
// ------------------------------------------------------------

func (h *Handler) Payments(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	limit, offset := pagination(c)

	payments, err := h.service.GetPayments(
		c.Request.Context(),
		tenantID,
		limit,
		offset,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load subscription payments",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"payments": payments,
		"limit":    limit,
		"offset":   offset,
	})
}

// ------------------------------------------------------------
// GET /api/subscription/ledger
// ------------------------------------------------------------

func (h *Handler) Ledger(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	limit, offset := pagination(c)

	entries, err := h.service.GetLedger(
		c.Request.Context(),
		tenantID,
		limit,
		offset,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load subscription ledger",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ledger": entries,
		"limit":  limit,
		"offset": offset,
	})
}

// ------------------------------------------------------------
// POST /api/subscription/payments
// ------------------------------------------------------------

type CreatePaymentRequest struct {
	PhoneNumber   string `json:"phone_number" binding:"required"`
	PaymentMethod string `json:"payment_method"`
}

func (h *Handler) CreatePayment(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	var req CreatePaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	payment, err := h.service.CreatePayment(
		c.Request.Context(),
		tenantID,
		req.PhoneNumber,
		req.PaymentMethod,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "subscription not found",
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, payment)
}

// ------------------------------------------------------------
// GET /api/platform/subscription-plans
// PLATFORM ADMIN ONLY
// ------------------------------------------------------------

func (h *Handler) PlatformPlans(c *gin.Context) {
	plans, err := h.service.GetAllPlans(
		c.Request.Context(),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to load subscription plans",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"plans": plans,
	})
}

// ------------------------------------------------------------
// POST /api/platform/subscription-plans
// PLATFORM ADMIN ONLY
// ------------------------------------------------------------

type CreatePlatformPlanRequest struct {
	Name              string `json:"name" binding:"required"`
	IncludedBranches  int    `json:"included_branches"`
	IncludedTerminals int    `json:"included_terminals"`
	ExtraBranchRate   string `json:"extra_branch_rate" binding:"required"`
	ExtraTerminalRate string `json:"extra_terminal_rate" binding:"required"`
	MonthlyPrice      string `json:"monthly_price" binding:"required"`
	IsLifetime        bool   `json:"is_lifetime"`
	IsActive          bool   `json:"is_active"`
}

func (h *Handler) CreatePlatformPlan(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)

	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authenticated user is required",
		})
		return
	}

	var req CreatePlatformPlanRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	plan := Plan{
		Name:              req.Name,
		IncludedBranches:  req.IncludedBranches,
		IncludedTerminals: req.IncludedTerminals,
		ExtraBranchRate:   req.ExtraBranchRate,
		ExtraTerminalRate: req.ExtraTerminalRate,
		MonthlyPrice:      req.MonthlyPrice,
		IsLifetime:        req.IsLifetime,
		IsActive:          req.IsActive,
	}

	created, err := h.service.CreatePlan(
		c.Request.Context(),
		plan,
		userID,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "subscription plan not found",
			})
			return
		}

		if strings.Contains(
			err.Error(),
			"already exists",
		) {
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, created)
}

// ------------------------------------------------------------
// PATCH /api/platform/subscription-plans/:id
// PLATFORM ADMIN ONLY
// ------------------------------------------------------------

type UpdatePlatformPlanRequest struct {
	Name              string `json:"name" binding:"required"`
	IncludedBranches  int    `json:"included_branches"`
	IncludedTerminals int    `json:"included_terminals"`
	ExtraBranchRate   string `json:"extra_branch_rate" binding:"required"`
	ExtraTerminalRate string `json:"extra_terminal_rate" binding:"required"`
	MonthlyPrice      string `json:"monthly_price" binding:"required"`
	IsLifetime        bool   `json:"is_lifetime"`
	IsActive          bool   `json:"is_active"`
}

func (h *Handler) UpdatePlatformPlan(c *gin.Context) {
	planID := c.Param("id")

	if planID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "plan ID is required",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)

	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "authenticated user is required",
		})
		return
	}

	var req UpdatePlatformPlanRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	plan := Plan{
		ID:                planID,
		Name:              req.Name,
		IncludedBranches:  req.IncludedBranches,
		IncludedTerminals: req.IncludedTerminals,
		ExtraBranchRate:   req.ExtraBranchRate,
		ExtraTerminalRate: req.ExtraTerminalRate,
		MonthlyPrice:      req.MonthlyPrice,
		IsLifetime:        req.IsLifetime,
		IsActive:          req.IsActive,
	}

	updated, err := h.service.UpdatePlan(
		c.Request.Context(),
		planID,
		plan,
		userID,
	)

	if err != nil {
		if errors.Is(err, ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "subscription plan not found",
			})
			return
		}

		if strings.Contains(
			err.Error(),
			"already exists",
		) {
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// ------------------------------------------------------------
// PAGINATION
// ------------------------------------------------------------

func pagination(c *gin.Context) (int, int) {
	limit := 50
	offset := 0

	if value := c.Query("limit"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			limit = parsed
		}
	}

	if value := c.Query("offset"); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			offset = parsed
		}
	}

	if limit <= 0 {
		limit = 50
	}

	if limit > 100 {
		limit = 100
	}

	if offset < 0 {
		offset = 0
	}

	return limit, offset
}

package terminalpayment

import (
	"net/http"

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

// ============================================================
// CASH PAYMENT
// ============================================================

func (h *Handler) CreateCashPayment(c *gin.Context) {
	var req CreateCashPaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	terminalID, exists := c.Get("terminal_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal authentication required",
		})
		return
	}

	tenantID, exists := c.Get("terminal_tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal tenant context missing",
		})
		return
	}

	branchID, exists := c.Get("terminal_branch_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal branch context missing",
		})
		return
	}

	terminalIDString, ok := terminalID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid terminal context",
		})
		return
	}

	tenantIDString, ok := tenantID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid tenant context",
		})
		return
	}

	branchIDString, ok := branchID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid branch context",
		})
		return
	}

	result, err := h.service.CreateCashPayment(
		c.Request.Context(),
		tenantIDString,
		terminalIDString,
		branchIDString,
		req,
	)
	if err != nil {
		switch err {
		case ErrSessionNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		case ErrSessionNotCompleted:
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, result)
}

// ============================================================
// M-PESA PAYMENT
// ============================================================

func (h *Handler) CreateMpesaPayment(c *gin.Context) {
	var req CreateMpesaPaymentRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request",
		})
		return
	}

	terminalID, exists := c.Get("terminal_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal authentication required",
		})
		return
	}

	tenantID, exists := c.Get("terminal_tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal tenant context missing",
		})
		return
	}

	branchID, exists := c.Get("terminal_branch_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal branch context missing",
		})
		return
	}

	terminalIDString, ok := terminalID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid terminal context",
		})
		return
	}

	tenantIDString, ok := tenantID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid tenant context",
		})
		return
	}

	branchIDString, ok := branchID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid branch context",
		})
		return
	}

	result, err := h.service.CreateMpesaPayment(
		c.Request.Context(),
		tenantIDString,
		terminalIDString,
		branchIDString,
		req,
	)
	if err != nil {
		switch err {
		case ErrSessionNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		case ErrSessionNotCompleted:
			c.JSON(http.StatusConflict, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusCreated, result)
}

// ============================================================
// M-PESA PAYMENT STATUS
// ============================================================

func (h *Handler) GetMpesaPaymentStatus(c *gin.Context) {
	paymentID := c.Param("paymentID")

	if paymentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "payment id is required",
		})
		return
	}

	terminalID, exists := c.Get("terminal_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal authentication required",
		})
		return
	}

	tenantID, exists := c.Get("terminal_tenant_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal tenant context missing",
		})
		return
	}

	branchID, exists := c.Get("terminal_branch_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "terminal branch context missing",
		})
		return
	}

	terminalIDString, ok := terminalID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid terminal context",
		})
		return
	}

	tenantIDString, ok := tenantID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid tenant context",
		})
		return
	}

	branchIDString, ok := branchID.(string)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid branch context",
		})
		return
	}

	result, err := h.service.GetMpesaPaymentStatus(
		c.Request.Context(),
		tenantIDString,
		terminalIDString,
		branchIDString,
		paymentID,
	)
	if err != nil {
		switch err {
		case ErrPaymentNotFound:
			c.JSON(http.StatusNotFound, gin.H{
				"error": err.Error(),
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

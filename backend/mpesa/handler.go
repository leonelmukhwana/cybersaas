package mpesa

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"cybersaas/backend/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// GetBranch returns the M-Pesa configuration for a branch.
// Secrets are never returned by the service/repository.
func (h *Handler) GetBranch(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context missing",
		})
		return
	}

	branchID := strings.TrimSpace(c.Param("branchID"))
	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch id is required",
		})
		return
	}

	config, err := h.service.GetBranchConfiguration(
		c.Request.Context(),
		tenantID,
		branchID,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrBranchNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "branch not found",
			})
		case errors.Is(err, ErrConfigurationNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "mpesa configuration not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to get mpesa configuration",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"configuration": config,
	})
}

// SaveBranch creates or updates the M-Pesa configuration for a branch.
func (h *Handler) SaveBranch(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context missing",
		})
		return
	}

	var request SaveConfigurationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	if strings.TrimSpace(request.BranchID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch id is required",
		})
		return
	}

	config, err := h.service.SaveBranchConfiguration(
		c.Request.Context(),
		tenantID,
		request,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidBranchID):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid branch id",
			})
		case errors.Is(err, ErrInvalidProvider):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid mpesa provider",
			})
		case errors.Is(err, ErrInvalidEnvironment):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid mpesa environment",
			})
		case errors.Is(err, ErrBranchNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "branch not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to save mpesa configuration",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "mpesa configuration saved successfully",
		"configuration": config,
	})
}

// DeleteBranch removes the M-Pesa configuration for a branch.
func (h *Handler) DeleteBranch(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context missing",
		})
		return
	}

	branchID := strings.TrimSpace(c.Param("branchID"))
	if branchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch id is required",
		})
		return
	}

	err := h.service.DeleteBranchConfiguration(
		c.Request.Context(),
		tenantID,
		branchID,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidBranchID):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid branch id",
			})
		case errors.Is(err, ErrConfigurationNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "mpesa configuration not found",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to delete mpesa configuration",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "mpesa configuration deleted successfully",
	})
}

// GetPlatform returns the platform M-Pesa configuration.
// Only platform_admin can reach this handler through the routes middleware.
func (h *Handler) GetPlatform(c *gin.Context) {
	config, err := h.service.GetPlatformConfiguration(
		c.Request.Context(),
	)
	if err != nil {
		if errors.Is(err, ErrConfigurationNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "platform mpesa configuration not found",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to get platform mpesa configuration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"configuration": config,
	})
}

// SavePlatform creates or updates the platform M-Pesa configuration.
func (h *Handler) SavePlatform(c *gin.Context) {
	var request PlatformSaveConfigurationRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	config, err := h.service.SavePlatformConfiguration(
		c.Request.Context(),
		request,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidProvider):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid mpesa provider",
			})
		case errors.Is(err, ErrInvalidEnvironment):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid mpesa environment",
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to save platform mpesa configuration",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "platform mpesa configuration saved successfully",
		"configuration": config,
	})
}

// InitiateSTKPush starts an M-Pesa STK Push for a sale.
func (h *Handler) InitiateSTKPush(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context missing",
		})
		return
	}

	userID, ok := middleware.CurrentUserID(c)
	if !ok || userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context missing",
		})
		return
	}

	var request InitiateSTKRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	request.BranchID = strings.TrimSpace(request.BranchID)
	request.SaleID = strings.TrimSpace(request.SaleID)
	request.PhoneNumber = strings.TrimSpace(request.PhoneNumber)
	request.Amount = strings.TrimSpace(request.Amount)
	request.AccountReference = strings.TrimSpace(request.AccountReference)
	request.TransactionDescription = strings.TrimSpace(request.TransactionDescription)

	if request.BranchID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "branch id is required",
		})
		return
	}

	if request.SaleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "sale id is required",
		})
		return
	}

	if request.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "phone number is required",
		})
		return
	}

	if request.Amount == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "amount is required",
		})
		return
	}

	if request.AccountReference == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "account reference is required",
		})
		return
	}

	if request.TransactionDescription == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "transaction description is required",
		})
		return
	}

	stkRequest, err := h.service.InitiateSTKPush(
		c.Request.Context(),
		tenantID,
		userID,
		request,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidBranchID):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid branch id",
			})
		case errors.Is(err, ErrInvalidProvider):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid mpesa provider",
			})
		case errors.Is(err, ErrBranchNotFound):
			c.JSON(http.StatusNotFound, gin.H{
				"error": "branch not found",
			})
		case errors.Is(err, ErrConfigurationNotFound):
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "mpesa configuration not found",
			})
		default:
			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "stk push initiated successfully",
		"stk_request": stkRequest,
	})
}

// InitiateSubscriptionSTKPush starts an M-Pesa STK Push
// for a CyberSaaS subscription payment.
func (h *Handler) InitiateSubscriptionSTKPush(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)
	if !ok || tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context missing",
		})
		return
	}

	var request struct {
		PhoneNumber string `json:"phone_number" binding:"required"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "phone number is required",
		})
		return
	}

	request.PhoneNumber = strings.TrimSpace(request.PhoneNumber)

	if request.PhoneNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "phone number is required",
		})
		return
	}

	stkRequest, err := h.service.InitiateSubscriptionSTKPush(
		c.Request.Context(),
		tenantID,
		request.PhoneNumber,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "subscription STK push initiated successfully",
		"stk_request": stkRequest,
	})
}

func (h *Handler) Callback(c *gin.Context) {
	var request STKCallbackRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"ResultCode": 1,
			"ResultDesc": "Invalid callback payload",
		})
		return
	}

	result := request.Body.StkCallback

	stkRequest, err := h.service.ProcessSTKCallback(
		c.Request.Context(),
		result,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrSTKRequestNotFound):
			// Return HTTP 200 so the callback endpoint itself remains
			// available without exposing internal database information.
			c.JSON(http.StatusOK, gin.H{
				"ResultCode": 0,
				"ResultDesc": "Callback received",
			})
			return

		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"ResultCode": 1,
				"ResultDesc": "Callback processing failed",
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"ResultCode":  0,
		"ResultDesc":  "Callback processed",
		"stk_request": stkRequest,
	})
}

package report

import (
	"errors"
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

// GenerateComplianceReport
//
// POST /api/reports/compliance
func (h *Handler) GenerateComplianceReport(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	userRole := c.GetString("role")

	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	var request GenerateComplianceReportRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	result, err := h.service.GenerateComplianceReport(
		c.Request.Context(),
		tenantID,
		userID,
		userRole,
		request,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTenantID),
			errors.Is(err, ErrInvalidUserID),
			errors.Is(err, ErrInvalidBranchID),
			errors.Is(err, ErrInvalidPeriod),
			errors.Is(err, ErrInvalidReportType):

			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, ErrReportNotFound):

			c.JSON(http.StatusNotFound, gin.H{
				"error": "branch or compliance report not found",
			})

		case errors.Is(err, ErrReportAccessDenied):

			c.JSON(http.StatusForbidden, gin.H{
				"error": "you do not have permission to access this branch report",
			})

		case errors.Is(err, ErrInvalidReportRole):

			c.JSON(http.StatusForbidden, gin.H{
				"error": "you do not have permission to generate compliance reports",
			})

		default:

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
		}

		return
	}

	c.JSON(http.StatusCreated, ComplianceReportResponse{
		Snapshot: result,
	})
}

// GetComplianceReport
//
// GET /api/reports/compliance/:id
func (h *Handler) GetComplianceReport(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	userRole := c.GetString("role")

	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	snapshotID := c.Param("id")

	result, err := h.service.GetComplianceReport(
		c.Request.Context(),
		tenantID,
		userID,
		userRole,
		snapshotID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTenantID),
			errors.Is(err, ErrInvalidUserID):

			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, ErrReportNotFound):

			c.JSON(http.StatusNotFound, gin.H{
				"error": "compliance report not found",
			})

		case errors.Is(err, ErrReportAccessDenied):

			c.JSON(http.StatusForbidden, gin.H{
				"error": "you do not have permission to access this branch report",
			})

		case errors.Is(err, ErrInvalidReportRole):

			c.JSON(http.StatusForbidden, gin.H{
				"error": "you do not have permission to access compliance reports",
			})

		default:

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to retrieve compliance report",
			})
		}

		return
	}

	c.JSON(http.StatusOK, ComplianceReportResponse{
		Snapshot: result,
	})
}

// ListComplianceReports
//
// GET /api/reports/compliance
//
// Owner:
//
//	branch_id is required.
//
// Attendant:
//
//	branch_id must NOT be supplied.
//	The backend automatically resolves the attendant's assigned branch.
func (h *Handler) ListComplianceReports(c *gin.Context) {
	tenantID := c.GetString("tenant_id")
	userID := c.GetString("user_id")
	userRole := c.GetString("role")

	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	branchID := c.Query("branch_id")

	results, err := h.service.ListComplianceReports(
		c.Request.Context(),
		tenantID,
		userID,
		userRole,
		branchID,
	)

	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidTenantID),
			errors.Is(err, ErrInvalidUserID),
			errors.Is(err, ErrInvalidBranchID):

			c.JSON(http.StatusBadRequest, gin.H{
				"error": err.Error(),
			})

		case errors.Is(err, ErrReportAccessDenied):

			c.JSON(http.StatusForbidden, gin.H{
				"error": "you do not have permission to access this branch report",
			})

		case errors.Is(err, ErrInvalidReportRole):

			c.JSON(http.StatusForbidden, gin.H{
				"error": "you do not have permission to access compliance reports",
			})

		default:

			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to retrieve compliance reports",
			})
		}

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reports": results,
	})
}

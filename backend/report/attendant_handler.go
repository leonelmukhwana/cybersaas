package report

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"cybersaas/backend/middleware"
)

type AttendantReportHandler struct {
	service *AttendantReportService
}

func NewAttendantReportHandler(
	service *AttendantReportService,
) *AttendantReportHandler {
	return &AttendantReportHandler{
		service: service,
	}
}

func (h *AttendantReportHandler) Summary(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	attendantID, ok := middleware.CurrentUserID(c)

	if !ok || attendantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user context is required",
		})
		return
	}

	role, ok := middleware.CurrentRole(c)

	if !ok || role != "attendant" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "attendant access is required",
		})
		return
	}

	periodStart, periodEnd, err := parseAttendantReportPeriod(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	result, err := h.service.GetSummary(
		c.Request.Context(),
		tenantID,
		attendantID,
		periodStart,
		periodEnd,
	)

	if err != nil {
		c.Error(err)

		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

func parseAttendantReportPeriod(
	c *gin.Context,
) (time.Time, time.Time, error) {
	startValue := c.Query("period_start")
	endValue := c.Query("period_end")

	if startValue == "" {
		return time.Time{}, time.Time{},
			&reportValidationError{
				"period_start is required",
			}
	}

	if endValue == "" {
		return time.Time{}, time.Time{},
			&reportValidationError{
				"period_end is required",
			}
	}

	start, err := time.Parse(
		time.RFC3339,
		startValue,
	)

	if err != nil {
		return time.Time{}, time.Time{},
			&reportValidationError{
				"period_start must be a valid RFC3339 date",
			}
	}

	end, err := time.Parse(
		time.RFC3339,
		endValue,
	)

	if err != nil {
		return time.Time{}, time.Time{},
			&reportValidationError{
				"period_end must be a valid RFC3339 date",
			}
	}

	if !end.After(start) {
		return time.Time{}, time.Time{},
			&reportValidationError{
				"period_end must be after period_start",
			}
	}

	return start, end, nil
}

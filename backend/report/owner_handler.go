package report

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"cybersaas/backend/middleware"
)

type OwnerReportHandler struct {
	service *OwnerReportService
}

func NewOwnerReportHandler(
	service *OwnerReportService,
) *OwnerReportHandler {
	return &OwnerReportHandler{
		service: service,
	}
}

func (h *OwnerReportHandler) Summary(c *gin.Context) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	periodStart, periodEnd, err := parseOwnerReportPeriod(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var branchID *string

	if value := c.Query("branch_id"); value != "" {
		branchID = &value
	}

	result, err := h.service.GetSummary(
		c.Request.Context(),
		tenantID,
		branchID,
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

func (h *OwnerReportHandler) DownloadDaily(c *gin.Context) {
	h.downloadReport(c, "daily")
}

func (h *OwnerReportHandler) DownloadMonthly(c *gin.Context) {
	h.downloadReport(c, "monthly")
}

func (h *OwnerReportHandler) downloadReport(
	c *gin.Context,
	reportType string,
) {
	tenantID, ok := middleware.CurrentTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "tenant context is required",
		})
		return
	}

	periodStart, periodEnd, err := parseOwnerReportPeriod(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	var branchID *string

	if value := c.Query("branch_id"); value != "" {
		branchID = &value
	}

	data, err := h.service.DownloadReport(
		c.Request.Context(),
		tenantID,
		branchID,
		periodStart,
		periodEnd,
	)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	datePart := periodStart.Format("2006-01-02")

	filename := fmt.Sprintf(
		"cybersaas-owner-%s-report-%s.csv",
		reportType,
		datePart,
	)

	c.Header(
		"Content-Disposition",
		fmt.Sprintf(`attachment; filename="%s"`, filename),
	)

	c.Data(
		http.StatusOK,
		"text/csv; charset=utf-8",
		data,
	)
}

func parseOwnerReportPeriod(
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

	start, err := time.Parse(time.RFC3339, startValue)
	if err != nil {
		return time.Time{}, time.Time{},
			&reportValidationError{
				"period_start must be a valid RFC3339 date",
			}
	}

	end, err := time.Parse(time.RFC3339, endValue)
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

type reportValidationError struct {
	message string
}

func (e *reportValidationError) Error() string {
	return e.message
}

package platform

import (
	"net/http"
	"strconv"
	"strings"

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
// DASHBOARD
// ============================================================

func (h *Handler) Dashboard(c *gin.Context) {
	stats, err := h.service.GetDashboardStats(
		c.Request.Context(),
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to load platform dashboard",
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// ============================================================
// PAGINATION
// ============================================================

func parsePlatformPagination(c *gin.Context) (string, int, int, error) {
	search := strings.TrimSpace(c.Query("search"))

	limit := 20
	offset := 0

	if value := c.Query("limit"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err != nil || parsed < 1 {
			return "", 0, 0, err
		}

		limit = parsed
	}

	if value := c.Query("offset"); value != "" {
		parsed, err := strconv.Atoi(value)

		if err != nil || parsed < 0 {
			return "", 0, 0, err
		}

		offset = parsed
	}

	return search, limit, offset, nil
}

// ============================================================
// CYBER OWNERS
// ============================================================

func (h *Handler) CyberOwners(c *gin.Context) {
	search, limit, offset, err := parsePlatformPagination(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid pagination parameters",
		})
		return
	}

	owners, err := h.service.GetCyberOwners(
		c.Request.Context(),
		search,
		limit,
		offset,
	)

	if err != nil {
		// Temporary diagnostic response.
		// We will remove the database error from the API response
		// once the problem is fixed.
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to load cyber owners",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, owners)
}

// ============================================================
// SUBSCRIPTIONS
// ============================================================

func (h *Handler) Subscriptions(c *gin.Context) {
	search, limit, offset, err := parsePlatformPagination(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid pagination parameters",
		})
		return
	}

	subscriptions, err := h.service.GetSubscriptions(
		c.Request.Context(),
		search,
		limit,
		offset,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to load subscriptions",
		})
		return
	}

	c.JSON(http.StatusOK, subscriptions)
}

// ============================================================
// REVENUE
// ============================================================

func (h *Handler) Revenue(c *gin.Context) {
	search, limit, offset, err := parsePlatformPagination(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid pagination parameters",
		})
		return
	}

	revenue, err := h.service.GetRevenue(
		c.Request.Context(),
		search,
		limit,
		offset,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to load platform revenue",
		})
		return
	}

	c.JSON(http.StatusOK, revenue)
}

// ============================================================
// AUDIT LOGS
// ============================================================

func (h *Handler) AuditLogs(c *gin.Context) {
	search, limit, offset, err := parsePlatformPagination(c)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "invalid pagination parameters",
		})
		return
	}

	logs, err := h.service.GetAuditLogs(
		c.Request.Context(),
		search,
		limit,
		offset,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "failed to load audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, logs)
}

package customer

import (
	"net/http"
	"strings"

	"cybersaas/backend/middleware"

	"github.com/gin-gonic/gin"
)

func (h *Handler) TerminalLookup(c *gin.Context) {
	tenantID, ok := middleware.CurrentTerminalTenantID(c)

	if !ok || tenantID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "terminal tenant context is required",
		})
		return
	}

	branchID, ok := middleware.CurrentTerminalBranchID(c)

	if !ok || branchID == "" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "terminal branch context is required",
		})
		return
	}

	var request TerminalCustomerLookupRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid customer lookup request",
		})
		return
	}

	request.SearchType =
		strings.ToLower(strings.TrimSpace(request.SearchType))

	request.SearchTerm =
		strings.TrimSpace(request.SearchTerm)

	if request.SearchType != "adult" &&
		request.SearchType != "child" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "search_type must be adult or child",
		})
		return
	}

	if request.SearchTerm == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "search_term is required",
		})
		return
	}

	// Adult lookup requires an ID number.
	if request.SearchType == "adult" {
		if request.IDNumber == nil ||
			strings.TrimSpace(*request.IDNumber) == "" {
			// Use search_term as the adult ID when the client
			// supplies the normalized generic search field.
			request.IDNumber = &request.SearchTerm
		}
	}

	// Child lookup deliberately does NOT require an ID number.
	if request.SearchType == "child" {
		request.FullName = &request.SearchTerm
	}

	response, err := h.service.LookupForTerminal(
		c.Request.Context(),
		tenantID,
		branchID,
		request.SearchType,
		request.SearchTerm,
	)

	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, response)
}

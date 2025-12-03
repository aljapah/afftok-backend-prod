package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminLogsHandler handles admin logs API endpoints
type AdminLogsHandler struct {
	observability *services.ObservabilityService
}

// NewAdminLogsHandler creates a new admin logs handler
func NewAdminLogsHandler() *AdminLogsHandler {
	return &AdminLogsHandler{
		observability: services.NewObservabilityService(),
	}
}

// LogsResponse represents logs response
type LogsResponse struct {
	CorrelationID string                 `json:"correlation_id"`
	Timestamp     string                 `json:"timestamp"`
	Logs          []services.LogEvent    `json:"logs"`
	Count         int                    `json:"count"`
	Filters       map[string]interface{} `json:"filters,omitempty"`
}

// GetRecentLogs returns recent structured logs
// GET /api/admin/logs/recent?limit=100&category=&user_id=&offer_id=&ip=
func (h *AdminLogsHandler) GetRecentLogs(c *gin.Context) {
	correlationID := generateCorrelationID()

	// Parse parameters
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}

	category := c.Query("category")
	userID := c.Query("user_id")
	offerID := c.Query("offer_id")
	ip := c.Query("ip")

	logs := h.observability.GetRecentLogs(limit, category, userID, offerID, ip)

	response := LogsResponse{
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Logs:          logs,
		Count:         len(logs),
		Filters: map[string]interface{}{
			"limit":    limit,
			"category": category,
			"user_id":  userID,
			"offer_id": offerID,
			"ip":       ip,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}

// GetErrorLogs returns only error logs
// GET /api/admin/logs/errors?limit=100
func (h *AdminLogsHandler) GetErrorLogs(c *gin.Context) {
	correlationID := generateCorrelationID()

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	logs := h.observability.GetLogsFromRedis(services.LogCategoryErrorEvent, limit)

	response := LogsResponse{
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Logs:          logs,
		Count:         len(logs),
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}

// GetFraudLogs returns only fraud detection logs
// GET /api/admin/logs/fraud?limit=100
func (h *AdminLogsHandler) GetFraudLogs(c *gin.Context) {
	correlationID := generateCorrelationID()

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	logs := h.observability.GetFraudLogs(limit)

	response := LogsResponse{
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Logs:          logs,
		Count:         len(logs),
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}

// GetLogCategories returns all available log categories
// GET /api/admin/logs/categories
func (h *AdminLogsHandler) GetLogCategories(c *gin.Context) {
	correlationID := generateCorrelationID()

	categories := []map[string]string{
		{"id": services.LogCategoryClickEvent, "name": "Click Events", "description": "Click tracking events"},
		{"id": services.LogCategoryPostbackEvent, "name": "Postback Events", "description": "Postback/callback events"},
		{"id": services.LogCategoryFraudDetection, "name": "Fraud Detection", "description": "Fraud and bot detection events"},
		{"id": services.LogCategoryRateLimitBlock, "name": "Rate Limit", "description": "Rate limiting events"},
		{"id": services.LogCategoryAdminAccess, "name": "Admin Access", "description": "Admin panel access events"},
		{"id": services.LogCategoryAuthEvent, "name": "Authentication", "description": "Login/logout/register events"},
		{"id": services.LogCategoryEarningsUpdate, "name": "Earnings", "description": "Earnings update events"},
		{"id": services.LogCategoryConversionEvent, "name": "Conversions", "description": "Conversion events"},
		{"id": services.LogCategorySystemEvent, "name": "System", "description": "System events"},
		{"id": services.LogCategoryErrorEvent, "name": "Errors", "description": "Error events"},
		{"id": services.LogCategoryPerformance, "name": "Performance", "description": "Performance metrics"},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": map[string]interface{}{
			"categories": categories,
			"count":      len(categories),
		},
	})
}

// GetLogsByCategory returns logs filtered by category
// GET /api/admin/logs/category/:category?limit=100
func (h *AdminLogsHandler) GetLogsByCategory(c *gin.Context) {
	correlationID := generateCorrelationID()
	category := c.Param("category")

	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Category is required",
		})
		return
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	logs := h.observability.GetLogsFromRedis(category, limit)

	response := LogsResponse{
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Logs:          logs,
		Count:         len(logs),
		Filters: map[string]interface{}{
			"category": category,
			"limit":    limit,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}

// GetLogsByIP returns logs filtered by IP address
// GET /api/admin/logs/ip/:ip?limit=100
func (h *AdminLogsHandler) GetLogsByIP(c *gin.Context) {
	correlationID := generateCorrelationID()
	ip := c.Param("ip")

	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "IP address is required",
		})
		return
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	logs := h.observability.GetRecentLogs(limit, "", "", "", ip)

	response := LogsResponse{
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Logs:          logs,
		Count:         len(logs),
		Filters: map[string]interface{}{
			"ip":    ip,
			"limit": limit,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}

// GetLogsByUser returns logs filtered by user ID
// GET /api/admin/logs/user/:id?limit=100
func (h *AdminLogsHandler) GetLogsByUser(c *gin.Context) {
	correlationID := generateCorrelationID()
	userID := c.Param("id")

	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "User ID is required",
		})
		return
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	logs := h.observability.GetRecentLogs(limit, "", userID, "", "")

	response := LogsResponse{
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Logs:          logs,
		Count:         len(logs),
		Filters: map[string]interface{}{
			"user_id": userID,
			"limit":   limit,
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}


package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
)

// ObservabilityHandler handles observability and monitoring endpoints
type ObservabilityHandler struct {
	observability *services.ObservabilityService
}

// NewObservabilityHandler creates a new ObservabilityHandler
func NewObservabilityHandler() *ObservabilityHandler {
	return &ObservabilityHandler{
		observability: services.NewObservabilityService(),
	}
}

// GetRecentLogs returns recent logs with filtering
// GET /api/admin/logs/recent?category=click_event&user_id=xxx&ip=xxx&limit=100
func (h *ObservabilityHandler) GetRecentLogs(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"logs":    logs,
		"count":   len(logs),
		"filters": gin.H{
			"category": category,
			"user_id":  userID,
			"offer_id": offerID,
			"ip":       ip,
			"limit":    limit,
		},
	})
}

// GetFraudLogs returns recent fraud detection logs
// GET /api/admin/logs/fraud?limit=100
func (h *ObservabilityHandler) GetFraudLogs(c *gin.Context) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	logs := h.observability.GetFraudLogs(limit)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"logs":    logs,
		"count":   len(logs),
	})
}

// GetLogsByCategory returns logs for a specific category
// GET /api/admin/logs/category/:category
func (h *ObservabilityHandler) GetLogsByCategory(c *gin.Context) {
	category := c.Param("category")
	if category == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Category required"})
		return
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	logs := h.observability.GetLogsFromRedis(category, limit)

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"category": category,
		"logs":     logs,
		"count":    len(logs),
	})
}

// GetMetrics returns system metrics
// GET /api/admin/metrics
func (h *ObservabilityHandler) GetMetrics(c *gin.Context) {
	metrics := h.observability.GetMetrics()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"metrics": metrics,
	})
}

// GetSystemHealth returns system health status
// GET /api/admin/health
func (h *ObservabilityHandler) GetSystemHealth(c *gin.Context) {
	health := h.observability.GetSystemHealth()

	status := "healthy"
	if dbHealthy, ok := health["database_healthy"].(bool); ok && !dbHealthy {
		status = "degraded"
	}
	if redisHealthy, ok := health["redis_healthy"].(bool); ok && !redisHealthy {
		status = "degraded"
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    status,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"checks":    health,
	})
}

// GetHealth returns simplified health for public endpoint
// GET /health
func (h *ObservabilityHandler) GetHealth(c *gin.Context) {
	health := h.observability.GetSystemHealth()

	status := "ok"
	statusCode := http.StatusOK

	if dbHealthy, ok := health["database_healthy"].(bool); ok && !dbHealthy {
		status = "degraded"
	}
	if redisHealthy, ok := health["redis_healthy"].(bool); ok && !redisHealthy {
		status = "degraded"
	}

	c.JSON(statusCode, gin.H{
		"status":  status,
		"message": "AffTok API is running",
	})
}

// GetConnections returns connection information
// GET /api/admin/connections
func (h *ObservabilityHandler) GetConnections(c *gin.Context) {
	health := h.observability.GetSystemHealth()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"database": gin.H{
			"healthy":     health["database_healthy"],
			"latency_ms":  health["database_latency_ms"],
			"connections": health["database_connections"],
		},
		"redis": gin.H{
			"healthy":    health["redis_healthy"],
			"latency_ms": health["redis_latency_ms"],
			"keys":       health["redis_keys"],
			"memory":     health["redis_memory"],
		},
	})
}

// GetFraudInsights returns fraud intelligence data
// GET /api/admin/fraud/insights
func (h *ObservabilityHandler) GetFraudInsights(c *gin.Context) {
	insights := h.observability.GetFraudInsights()

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"insights": insights,
	})
}

// GetRedisDiagnostics returns Redis diagnostics
// GET /api/admin/diagnostics/redis
func (h *ObservabilityHandler) GetRedisDiagnostics(c *gin.Context) {
	diagnostics := h.observability.GetRedisDiagnostics()

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"diagnostics": diagnostics,
	})
}

// GetDBDiagnostics returns database diagnostics
// GET /api/admin/diagnostics/db
func (h *ObservabilityHandler) GetDBDiagnostics(c *gin.Context) {
	health := h.observability.GetSystemHealth()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"diagnostics": gin.H{
			"healthy":          health["database_healthy"],
			"latency_ms":       health["database_latency_ms"],
			"open_connections": health["database_connections"],
		},
	})
}

// GetDashboardSummary returns a summary for admin dashboard
// GET /api/admin/dashboard
func (h *ObservabilityHandler) GetDashboardSummary(c *gin.Context) {
	metrics := h.observability.GetMetrics()
	health := h.observability.GetSystemHealth()
	fraud := h.observability.GetFraudInsights()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"summary": gin.H{
			// Health
			"db_healthy":    health["database_healthy"],
			"redis_healthy": health["redis_healthy"],
			
			// Clicks
			"total_clicks":        metrics.TotalClicks,
			"blocked_clicks":      metrics.ClicksBlocked,
			"bot_clicks":          metrics.ClicksFromBots,
			"duplicate_clicks":    metrics.ClicksDuplicate,
			"rate_limited_clicks": metrics.ClicksRateLimited,
			"avg_click_time_ms":   metrics.AvgClickProcessingMs,
			
			// Conversions
			"total_conversions":     metrics.TotalConversions,
			"approved_conversions":  metrics.ConversionsApproved,
			"rejected_conversions":  metrics.ConversionsRejected,
			
			// Postbacks
			"total_postbacks":    metrics.TotalPostbacks,
			"valid_postbacks":    metrics.PostbacksValid,
			"invalid_postbacks":  metrics.PostbacksInvalid,
			"replayed_postbacks": metrics.PostbacksReplayed,
			
			// Fraud
			"fraud_attempts":   metrics.FraudAttempts,
			"ips_blocked":      metrics.IPsBlocked,
			"bots_detected":    metrics.BotsDetected,
			"risky_ips_count":  len(fraud.TopRiskyIPs),
			
			// Auth
			"total_logins":        metrics.TotalLogins,
			"failed_logins":       metrics.FailedLogins,
			"total_registrations": metrics.TotalRegistrations,
			
			// Errors
			"total_errors": metrics.TotalErrors,
			"error_4xx":    metrics.Error4xx,
			"error_5xx":    metrics.Error5xx,
			
			// Performance
			"memory_usage_mb":  health["memory_mb"],
			"goroutines":       health["goroutines"],
			"uptime_seconds":   health["uptime_seconds"],
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

// GetLogCategories returns available log categories
// GET /api/admin/logs/categories
func (h *ObservabilityHandler) GetLogCategories(c *gin.Context) {
	categories := []string{
		services.LogCategoryClickEvent,
		services.LogCategoryPostbackEvent,
		services.LogCategoryFraudDetection,
		services.LogCategoryRateLimitBlock,
		services.LogCategoryAdminAccess,
		services.LogCategoryAuthEvent,
		services.LogCategoryEarningsUpdate,
		services.LogCategoryConversionEvent,
		services.LogCategorySystemEvent,
		services.LogCategoryErrorEvent,
		services.LogCategoryPerformance,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"categories": categories,
	})
}

// ExportMetrics exports all metrics as JSON
// GET /api/admin/metrics/export
func (h *ObservabilityHandler) ExportMetrics(c *gin.Context) {
	metrics := h.observability.GetMetrics()
	health := h.observability.GetSystemHealth()
	fraud := h.observability.GetFraudInsights()

	export := gin.H{
		"metrics":   metrics,
		"health":    health,
		"fraud":     fraud,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename=afftok-metrics.json")
	c.JSON(http.StatusOK, export)
}

// GetLogsByIP returns logs for a specific IP
// GET /api/admin/logs/ip/:ip
func (h *ObservabilityHandler) GetLogsByIP(c *gin.Context) {
	ip := c.Param("ip")
	if ip == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "IP address required"})
		return
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	logs := h.observability.GetRecentLogs(limit, "", "", "", ip)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"ip":      ip,
		"logs":    logs,
		"count":   len(logs),
	})
}

// GetLogsByUser returns logs for a specific user
// GET /api/admin/logs/user/:user_id
func (h *ObservabilityHandler) GetLogsByUser(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	logs := h.observability.GetRecentLogs(limit, "", userID, "", "")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"user_id": userID,
		"logs":    logs,
		"count":   len(logs),
	})
}

// GetErrorLogs returns recent error logs
// GET /api/admin/logs/errors
func (h *ObservabilityHandler) GetErrorLogs(c *gin.Context) {
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 500 {
			limit = parsed
		}
	}

	logs := h.observability.GetLogsFromRedis(services.LogCategoryErrorEvent, limit)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"logs":    logs,
		"count":   len(logs),
	})
}

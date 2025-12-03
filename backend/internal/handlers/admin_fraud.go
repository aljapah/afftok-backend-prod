package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminFraudHandler handles admin fraud API endpoints
type AdminFraudHandler struct {
	observability   *services.ObservabilityService
	securityService *services.SecurityService
}

// NewAdminFraudHandler creates a new admin fraud handler
func NewAdminFraudHandler() *AdminFraudHandler {
	return &AdminFraudHandler{
		observability:   services.NewObservabilityService(),
		securityService: services.NewSecurityService(),
	}
}

// FraudInsightsResponse represents fraud insights response
type FraudInsightsResponse struct {
	CorrelationID    string              `json:"correlation_id"`
	Timestamp        string              `json:"timestamp"`
	Summary          FraudSummary        `json:"summary"`
	TopRiskyIPs      []RiskyIPInfo       `json:"top_risky_ips"`
	RecentAttempts   []services.LogEvent `json:"recent_attempts"`
	HourlyHistogram  map[string]int64    `json:"hourly_histogram"`
	RiskIndicators   []RiskIndicator     `json:"risk_indicators"`
}

// FraudSummary represents fraud summary
type FraudSummary struct {
	TotalAttempts     int64   `json:"total_attempts"`
	BlockedAttempts   int64   `json:"blocked_attempts"`
	IPsBlocked        int64   `json:"ips_blocked"`
	BotsDetected      int64   `json:"bots_detected"`
	ReplayAttempts    int64   `json:"replay_attempts"`
	RateLimitBlocks   int64   `json:"rate_limit_blocks"`
	BlockRate         float64 `json:"block_rate_percent"`
}

// RiskyIPInfo represents information about a risky IP
type RiskyIPInfo struct {
	IP           string    `json:"ip"`
	RiskScore    int       `json:"risk_score"`
	Attempts     int64     `json:"attempts"`
	LastSeen     string    `json:"last_seen"`
	Indicators   []string  `json:"indicators"`
	IsBlocked    bool      `json:"is_blocked"`
	Country      string    `json:"country,omitempty"`
}

// RiskIndicator represents a risk indicator
type RiskIndicator struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Count       int64  `json:"count"`
	Severity    string `json:"severity"`
}

// GetFraudInsights returns comprehensive fraud intelligence
// GET /api/admin/fraud/insights
func (h *AdminFraudHandler) GetFraudInsights(c *gin.Context) {
	correlationID := generateCorrelationID()

	metrics := h.observability.GetMetrics()
	fraudInsights := h.observability.GetFraudInsights()
	fraudLogs := h.observability.GetFraudLogs(50)

	// Build summary
	totalAttempts := metrics.FraudAttempts + metrics.ClicksFromBots + metrics.ClicksRateLimited
	blockedAttempts := metrics.ClicksBlocked + metrics.IPsBlocked
	blockRate := float64(0)
	if totalAttempts > 0 {
		blockRate = float64(blockedAttempts) / float64(totalAttempts) * 100
	}

	summary := FraudSummary{
		TotalAttempts:   totalAttempts,
		BlockedAttempts: blockedAttempts,
		IPsBlocked:      metrics.IPsBlocked,
		BotsDetected:    metrics.BotsDetected,
		ReplayAttempts:  metrics.PostbacksReplayed,
		RateLimitBlocks: metrics.ClicksRateLimited,
		BlockRate:       blockRate,
	}

	// Build risky IPs list
	riskyIPs := make([]RiskyIPInfo, 0)
	for _, ip := range fraudInsights.TopRiskyIPs {
		riskyIPs = append(riskyIPs, RiskyIPInfo{
			IP:         ip.IP,
			RiskScore:  ip.RiskScore,
			Attempts:   ip.Attempts,
			LastSeen:   ip.LastSeen,
			Indicators: []string{}, // Indicators not available in source
			IsBlocked:  ip.RiskScore >= 80,
		})
	}

	// Build hourly histogram from Redis
	hourlyHistogram := h.getHourlyFraudHistogram()

	// Build risk indicators
	indicators := []RiskIndicator{
		{
			Name:        "bot_detection",
			Description: "Requests identified as bot traffic",
			Count:       metrics.BotsDetected,
			Severity:    getSeverity(metrics.BotsDetected, 100, 500),
		},
		{
			Name:        "rate_limit_violations",
			Description: "Requests blocked by rate limiting",
			Count:       metrics.ClicksRateLimited,
			Severity:    getSeverity(metrics.ClicksRateLimited, 50, 200),
		},
		{
			Name:        "duplicate_clicks",
			Description: "Duplicate click attempts detected",
			Count:       metrics.ClicksDuplicate,
			Severity:    getSeverity(metrics.ClicksDuplicate, 100, 500),
		},
		{
			Name:        "replay_attacks",
			Description: "Postback replay attempts",
			Count:       metrics.PostbacksReplayed,
			Severity:    getSeverity(metrics.PostbacksReplayed, 10, 50),
		},
		{
			Name:        "invalid_postbacks",
			Description: "Invalid postback requests",
			Count:       metrics.PostbacksInvalid,
			Severity:    getSeverity(metrics.PostbacksInvalid, 20, 100),
		},
		{
			Name:        "failed_auth",
			Description: "Failed authentication attempts",
			Count:       metrics.FailedLogins,
			Severity:    getSeverity(metrics.FailedLogins, 50, 200),
		},
	}

	response := FraudInsightsResponse{
		CorrelationID:   correlationID,
		Timestamp:       time.Now().UTC().Format(time.RFC3339),
		Summary:         summary,
		TopRiskyIPs:     riskyIPs,
		RecentAttempts:  fraudLogs,
		HourlyHistogram: hourlyHistogram,
		RiskIndicators:  indicators,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}

// getHourlyFraudHistogram returns fraud attempts by hour
func (h *AdminFraudHandler) getHourlyFraudHistogram() map[string]int64 {
	histogram := make(map[string]int64)
	
	if cache.RedisClient == nil {
		return histogram
	}

	ctx := context.Background()
	now := time.Now()

	// Get last 24 hours
	for i := 0; i < 24; i++ {
		t := now.Add(-time.Duration(i) * time.Hour)
		hourKey := services.NSFraud + "hourly:" + t.Format("2006010215")
		
		if count, err := cache.RedisClient.Get(ctx, hourKey).Int64(); err == nil {
			histogram[t.Format("15:00")] = count
		} else {
			histogram[t.Format("15:00")] = 0
		}
	}

	return histogram
}

// getSeverity returns severity based on count thresholds
func getSeverity(count int64, warningThreshold, criticalThreshold int64) string {
	if count >= criticalThreshold {
		return "critical"
	}
	if count >= warningThreshold {
		return "warning"
	}
	return "normal"
}

// BlockIP blocks an IP address
// POST /api/admin/fraud/block-ip
func (h *AdminFraudHandler) BlockIP(c *gin.Context) {
	correlationID := generateCorrelationID()

	var req struct {
		IP       string `json:"ip" binding:"required"`
		Reason   string `json:"reason"`
		Duration int    `json:"duration_hours"` // 0 = permanent
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request",
		})
		return
	}

	// Store blocked IP in Redis
	if cache.RedisClient != nil {
		ctx := context.Background()
		key := services.NSFraud + "blocked_ip:" + req.IP
		
		duration := time.Duration(0)
		if req.Duration > 0 {
			duration = time.Duration(req.Duration) * time.Hour
		} else {
			duration = 365 * 24 * time.Hour // 1 year for "permanent"
		}

		blockData := map[string]interface{}{
			"ip":         req.IP,
			"reason":     req.Reason,
			"blocked_at": time.Now().UTC().Format(time.RFC3339),
			"blocked_by": c.GetString("userID"),
		}

		cache.RedisClient.HSet(ctx, key, blockData)
		cache.RedisClient.Expire(ctx, key, duration)

		// Log the action
		h.observability.LogFraud(req.IP, "", "manual_block: "+req.Reason, 100, 1.0, []string{"admin_blocked"}, nil)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": map[string]interface{}{
			"ip":       req.IP,
			"blocked":  true,
			"duration": req.Duration,
		},
	})
}

// UnblockIP unblocks an IP address
// POST /api/admin/fraud/unblock-ip
func (h *AdminFraudHandler) UnblockIP(c *gin.Context) {
	correlationID := generateCorrelationID()

	var req struct {
		IP string `json:"ip" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request",
		})
		return
	}

	// Remove blocked IP from Redis
	if cache.RedisClient != nil {
		ctx := context.Background()
		key := services.NSFraud + "blocked_ip:" + req.IP
		cache.RedisClient.Del(ctx, key)
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": map[string]interface{}{
			"ip":        req.IP,
			"unblocked": true,
		},
	})
}

// GetBlockedIPs returns list of blocked IPs
// GET /api/admin/fraud/blocked-ips
func (h *AdminFraudHandler) GetBlockedIPs(c *gin.Context) {
	correlationID := generateCorrelationID()

	blockedIPs := make([]map[string]interface{}, 0)

	if cache.RedisClient != nil {
		ctx := context.Background()
		
		// Scan for blocked IP keys
		var cursor uint64
		for {
			keys, nextCursor, err := cache.RedisClient.Scan(ctx, cursor, services.NSFraud+"blocked_ip:*", 100).Result()
			if err != nil {
				break
			}

			for _, key := range keys {
				data, err := cache.RedisClient.HGetAll(ctx, key).Result()
				if err == nil && len(data) > 0 {
					blockedIPs = append(blockedIPs, map[string]interface{}{
						"ip":         data["ip"],
						"reason":     data["reason"],
						"blocked_at": data["blocked_at"],
						"blocked_by": data["blocked_by"],
					})
				}
			}

			cursor = nextCursor
			if cursor == 0 {
				break
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": map[string]interface{}{
			"blocked_ips": blockedIPs,
			"count":       len(blockedIPs),
		},
	})
}


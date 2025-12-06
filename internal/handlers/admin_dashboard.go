package handlers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/database"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminDashboardHandler handles admin dashboard API endpoints
type AdminDashboardHandler struct {
	observability *services.ObservabilityService
	cacheService  *services.CacheService
}

// NewAdminDashboardHandler creates a new admin dashboard handler
func NewAdminDashboardHandler() *AdminDashboardHandler {
	return &AdminDashboardHandler{
		observability: services.NewObservabilityService(),
		cacheService:  services.NewCacheService(),
	}
}

// generateCorrelationID creates a unique correlation ID for request tracing
func generateCorrelationID() string {
	return uuid.New().String()[:8]
}

// DashboardResponse represents the complete dashboard data
type DashboardResponse struct {
	CorrelationID string                 `json:"correlation_id"`
	Timestamp     string                 `json:"timestamp"`
	Health        HealthStatus           `json:"health"`
	Clicks        ClicksStats            `json:"clicks"`
	Conversions   ConversionsStats       `json:"conversions"`
	Postbacks     PostbacksStats         `json:"postbacks"`
	Fraud         FraudStats             `json:"fraud"`
	Auth          AuthStats              `json:"auth"`
	Errors        ErrorStats             `json:"errors"`
	System        SystemStats            `json:"system"`
	Performance   PerformanceStats       `json:"performance"`
}

// HealthStatus represents system health
type HealthStatus struct {
	DatabaseHealthy bool   `json:"database_healthy"`
	DatabaseLatency int64  `json:"database_latency_ms"`
	RedisHealthy    bool   `json:"redis_healthy"`
	RedisLatency    int64  `json:"redis_latency_ms"`
	OverallStatus   string `json:"overall_status"`
}

// ClicksStats represents click statistics
type ClicksStats struct {
	Total         int64 `json:"total"`
	Blocked       int64 `json:"blocked"`
	FromBots      int64 `json:"from_bots"`
	Duplicate     int64 `json:"duplicate"`
	RateLimited   int64 `json:"rate_limited"`
	Unique        int64 `json:"unique"`
	AvgProcessMs  int64 `json:"avg_process_ms"`
}

// ConversionsStats represents conversion statistics
type ConversionsStats struct {
	Total    int64 `json:"total"`
	Approved int64 `json:"approved"`
	Rejected int64 `json:"rejected"`
	Pending  int64 `json:"pending"`
}

// PostbacksStats represents postback statistics
type PostbacksStats struct {
	Total    int64 `json:"total"`
	Valid    int64 `json:"valid"`
	Invalid  int64 `json:"invalid"`
	Replayed int64 `json:"replayed"`
	AvgMs    int64 `json:"avg_process_ms"`
}

// FraudStats represents fraud detection statistics
type FraudStats struct {
	Attempts     int64 `json:"attempts"`
	IPsBlocked   int64 `json:"ips_blocked"`
	BotsDetected int64 `json:"bots_detected"`
	RiskyIPs     int   `json:"risky_ips_count"`
}

// AuthStats represents authentication statistics
type AuthStats struct {
	TotalLogins        int64 `json:"total_logins"`
	FailedLogins       int64 `json:"failed_logins"`
	TotalRegistrations int64 `json:"total_registrations"`
}

// ErrorStats represents error statistics
type ErrorStats struct {
	Total    int64 `json:"total"`
	Error4xx int64 `json:"error_4xx"`
	Error5xx int64 `json:"error_5xx"`
}

// SystemStats represents system statistics
type SystemStats struct {
	MemoryUsageMB float64 `json:"memory_usage_mb"`
	MemoryAllocMB float64 `json:"memory_alloc_mb"`
	MemorySysMB   float64 `json:"memory_sys_mb"`
	Goroutines    int     `json:"goroutines"`
	NumCPU        int     `json:"num_cpu"`
	UptimeSeconds int64   `json:"uptime_seconds"`
	StartTime     string  `json:"start_time"`
}

// PerformanceStats represents performance statistics
type PerformanceStats struct {
	DBConnections     int     `json:"db_connections"`
	DBMaxOpen         int     `json:"db_max_open"`
	DBInUse           int     `json:"db_in_use"`
	DBIdle            int     `json:"db_idle"`
	RedisPoolSize     uint32  `json:"redis_pool_size"`
	RedisActiveConns  uint32  `json:"redis_active_conns"`
	RedisIdleConns    uint32  `json:"redis_idle_conns"`
	AvgClickMs        float64 `json:"avg_click_ms"`
	AvgPostbackMs     float64 `json:"avg_postback_ms"`
	AvgDBQueryMs      float64 `json:"avg_db_query_ms"`
}

// GetDashboard returns complete dashboard data
// GET /api/admin/dashboard
func (h *AdminDashboardHandler) GetDashboard(c *gin.Context) {
	correlationID := generateCorrelationID()
	
	metrics := h.observability.GetMetrics()
	health := h.observability.GetSystemHealth()
	fraud := h.observability.GetFraudInsights()

	// Build health status
	healthStatus := HealthStatus{
		OverallStatus: "healthy",
	}
	
	if dbHealthy, ok := health["database_healthy"].(bool); ok {
		healthStatus.DatabaseHealthy = dbHealthy
		if !dbHealthy {
			healthStatus.OverallStatus = "degraded"
		}
	}
	if dbLatency, ok := health["database_latency_ms"].(int64); ok {
		healthStatus.DatabaseLatency = dbLatency
	}
	if redisHealthy, ok := health["redis_healthy"].(bool); ok {
		healthStatus.RedisHealthy = redisHealthy
		if !redisHealthy {
			healthStatus.OverallStatus = "degraded"
		}
	}
	if redisLatency, ok := health["redis_latency_ms"].(int64); ok {
		healthStatus.RedisLatency = redisLatency
	}

	// Build clicks stats
	clicksStats := ClicksStats{
		Total:        metrics.TotalClicks,
		Blocked:      metrics.ClicksBlocked,
		FromBots:     metrics.ClicksFromBots,
		Duplicate:    metrics.ClicksDuplicate,
		RateLimited:  metrics.ClicksRateLimited,
		Unique:       metrics.TotalClicks - metrics.ClicksDuplicate,
		AvgProcessMs: metrics.AvgClickProcessingMs,
	}

	// Build conversions stats
	conversionsStats := ConversionsStats{
		Total:    metrics.TotalConversions,
		Approved: metrics.ConversionsApproved,
		Rejected: metrics.ConversionsRejected,
		Pending:  metrics.TotalConversions - metrics.ConversionsApproved - metrics.ConversionsRejected,
	}

	// Build postbacks stats
	postbacksStats := PostbacksStats{
		Total:    metrics.TotalPostbacks,
		Valid:    metrics.PostbacksValid,
		Invalid:  metrics.PostbacksInvalid,
		Replayed: metrics.PostbacksReplayed,
		AvgMs:    metrics.AvgPostbackMs,
	}

	// Build fraud stats
	fraudStats := FraudStats{
		Attempts:     metrics.FraudAttempts,
		IPsBlocked:   metrics.IPsBlocked,
		BotsDetected: metrics.BotsDetected,
		RiskyIPs:     len(fraud.TopRiskyIPs),
	}

	// Build auth stats
	authStats := AuthStats{
		TotalLogins:        metrics.TotalLogins,
		FailedLogins:       metrics.FailedLogins,
		TotalRegistrations: metrics.TotalRegistrations,
	}

	// Build error stats
	errorStats := ErrorStats{
		Total:    metrics.TotalErrors,
		Error4xx: metrics.Error4xx,
		Error5xx: metrics.Error5xx,
	}

	// Build system stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	systemStats := SystemStats{
		MemoryUsageMB: float64(memStats.Alloc) / 1024 / 1024,
		MemoryAllocMB: float64(memStats.TotalAlloc) / 1024 / 1024,
		MemorySysMB:   float64(memStats.Sys) / 1024 / 1024,
		Goroutines:    runtime.NumGoroutine(),
		NumCPU:        runtime.NumCPU(),
		StartTime:     metrics.StartTime.Format(time.RFC3339),
	}
	if uptime, ok := health["uptime_seconds"].(int64); ok {
		systemStats.UptimeSeconds = uptime
	}

	// Build performance stats
	perfStats := PerformanceStats{}
	
	// Database stats
	if database.DB != nil {
		if sqlDB, err := database.DB.DB(); err == nil {
			stats := sqlDB.Stats()
			perfStats.DBConnections = stats.OpenConnections
			perfStats.DBMaxOpen = stats.MaxOpenConnections
			perfStats.DBInUse = stats.InUse
			perfStats.DBIdle = stats.Idle
		}
	}

	// Redis stats
	if poolStats := cache.GetPoolStats(); poolStats != nil {
		perfStats.RedisPoolSize = poolStats.TotalConns
		perfStats.RedisIdleConns = poolStats.IdleConns
	}

	// Build response
	response := DashboardResponse{
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Health:        healthStatus,
		Clicks:        clicksStats,
		Conversions:   conversionsStats,
		Postbacks:     postbacksStats,
		Fraud:         fraudStats,
		Auth:          authStats,
		Errors:        errorStats,
		System:        systemStats,
		Performance:   perfStats,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}


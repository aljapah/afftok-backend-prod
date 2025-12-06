package handlers

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/database"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminMetricsHandler handles admin metrics API endpoints
type AdminMetricsHandler struct {
	observability *services.ObservabilityService
}

// NewAdminMetricsHandler creates a new admin metrics handler
func NewAdminMetricsHandler() *AdminMetricsHandler {
	return &AdminMetricsHandler{
		observability: services.NewObservabilityService(),
	}
}

// MetricsResponse represents complete metrics data
type MetricsResponse struct {
	CorrelationID string                 `json:"correlation_id"`
	Timestamp     string                 `json:"timestamp"`
	Counters      MetricsCounters        `json:"counters"`
	Performance   MetricsPerformance     `json:"performance"`
	System        MetricsSystem          `json:"system"`
	WorkerPools   map[string]interface{} `json:"worker_pools"`
	Cache         map[string]interface{} `json:"cache"`
}

// MetricsCounters represents all metric counters
type MetricsCounters struct {
	// Clicks
	TotalClicks       int64 `json:"total_clicks"`
	ClicksBlocked     int64 `json:"clicks_blocked"`
	ClicksFromBots    int64 `json:"clicks_from_bots"`
	ClicksDuplicate   int64 `json:"clicks_duplicate"`
	ClicksRateLimited int64 `json:"clicks_rate_limited"`
	
	// Postbacks
	TotalPostbacks    int64 `json:"total_postbacks"`
	PostbacksValid    int64 `json:"postbacks_valid"`
	PostbacksInvalid  int64 `json:"postbacks_invalid"`
	PostbacksReplayed int64 `json:"postbacks_replayed"`
	
	// Conversions
	TotalConversions    int64 `json:"total_conversions"`
	ConversionsApproved int64 `json:"conversions_approved"`
	ConversionsRejected int64 `json:"conversions_rejected"`
	
	// Auth
	TotalLogins        int64 `json:"total_logins"`
	FailedLogins       int64 `json:"failed_logins"`
	TotalRegistrations int64 `json:"total_registrations"`
	
	// Fraud
	FraudAttempts int64 `json:"fraud_attempts"`
	IPsBlocked    int64 `json:"ips_blocked"`
	BotsDetected  int64 `json:"bots_detected"`
	
	// Errors
	TotalErrors int64 `json:"total_errors"`
	Error4xx    int64 `json:"error_4xx"`
	Error5xx    int64 `json:"error_5xx"`
	
	// System
	DBQueryCount    int64 `json:"db_query_count"`
	DBSlowQueries   int64 `json:"db_slow_queries"`
	RedisOperations int64 `json:"redis_operations"`
	RedisErrors     int64 `json:"redis_errors"`
}

// MetricsPerformance represents performance metrics
type MetricsPerformance struct {
	AvgClickProcessingMs    int64   `json:"avg_click_processing_ms"`
	AvgPostbackProcessingMs int64   `json:"avg_postback_processing_ms"`
	DBConnectionsOpen       int     `json:"db_connections_open"`
	DBConnectionsInUse      int     `json:"db_connections_in_use"`
	DBConnectionsIdle       int     `json:"db_connections_idle"`
	DBMaxOpenConnections    int     `json:"db_max_open_connections"`
	DBWaitCount             int64   `json:"db_wait_count"`
	DBWaitDuration          string  `json:"db_wait_duration"`
	RedisPoolSize           uint32  `json:"redis_pool_size"`
	RedisActiveConns        uint32  `json:"redis_active_conns"`
	RedisIdleConns          uint32  `json:"redis_idle_conns"`
	RedisMisses             uint32  `json:"redis_misses"`
	RedisTimeouts           uint32  `json:"redis_timeouts"`
}

// MetricsSystem represents system metrics
type MetricsSystem struct {
	GoVersion     string  `json:"go_version"`
	NumCPU        int     `json:"num_cpu"`
	NumGoroutine  int     `json:"num_goroutine"`
	MemoryAllocMB float64 `json:"memory_alloc_mb"`
	MemoryTotalMB float64 `json:"memory_total_mb"`
	MemorySysMB   float64 `json:"memory_sys_mb"`
	MemoryHeapMB  float64 `json:"memory_heap_mb"`
	MemoryStackMB float64 `json:"memory_stack_mb"`
	GCPauseNs     uint64  `json:"gc_pause_ns"`
	GCNumGC       uint32  `json:"gc_num_gc"`
	UptimeSeconds int64   `json:"uptime_seconds"`
}

// GetMetrics returns all performance metrics
// GET /api/admin/metrics
func (h *AdminMetricsHandler) GetMetrics(c *gin.Context) {
	correlationID := generateCorrelationID()
	metrics := h.observability.GetMetrics()

	// Build counters
	counters := MetricsCounters{
		TotalClicks:         metrics.TotalClicks,
		ClicksBlocked:       metrics.ClicksBlocked,
		ClicksFromBots:      metrics.ClicksFromBots,
		ClicksDuplicate:     metrics.ClicksDuplicate,
		ClicksRateLimited:   metrics.ClicksRateLimited,
		TotalPostbacks:      metrics.TotalPostbacks,
		PostbacksValid:      metrics.PostbacksValid,
		PostbacksInvalid:    metrics.PostbacksInvalid,
		PostbacksReplayed:   metrics.PostbacksReplayed,
		TotalConversions:    metrics.TotalConversions,
		ConversionsApproved: metrics.ConversionsApproved,
		ConversionsRejected: metrics.ConversionsRejected,
		TotalLogins:         metrics.TotalLogins,
		FailedLogins:        metrics.FailedLogins,
		TotalRegistrations:  metrics.TotalRegistrations,
		FraudAttempts:       metrics.FraudAttempts,
		IPsBlocked:          metrics.IPsBlocked,
		BotsDetected:        metrics.BotsDetected,
		TotalErrors:         metrics.TotalErrors,
		Error4xx:            metrics.Error4xx,
		Error5xx:            metrics.Error5xx,
		DBQueryCount:        metrics.DBQueryCount,
		DBSlowQueries:       metrics.DBSlowQueries,
		RedisOperations:     metrics.RedisOperations,
		RedisErrors:         metrics.RedisErrors,
	}

	// Build performance
	performance := MetricsPerformance{
		AvgClickProcessingMs:    metrics.AvgClickProcessingMs,
		AvgPostbackProcessingMs: metrics.AvgPostbackMs,
	}

	// Database stats
	if database.DB != nil {
		if sqlDB, err := database.DB.DB(); err == nil {
			stats := sqlDB.Stats()
			performance.DBConnectionsOpen = stats.OpenConnections
			performance.DBConnectionsInUse = stats.InUse
			performance.DBConnectionsIdle = stats.Idle
			performance.DBMaxOpenConnections = stats.MaxOpenConnections
			performance.DBWaitCount = stats.WaitCount
			performance.DBWaitDuration = stats.WaitDuration.String()
		}
	}

	// Redis stats
	if poolStats := cache.GetPoolStats(); poolStats != nil {
		performance.RedisPoolSize = poolStats.TotalConns
		performance.RedisIdleConns = poolStats.IdleConns
		performance.RedisMisses = poolStats.Misses
		performance.RedisTimeouts = poolStats.Timeouts
	}

	// Build system
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	system := MetricsSystem{
		GoVersion:     runtime.Version(),
		NumCPU:        runtime.NumCPU(),
		NumGoroutine:  runtime.NumGoroutine(),
		MemoryAllocMB: float64(memStats.Alloc) / 1024 / 1024,
		MemoryTotalMB: float64(memStats.TotalAlloc) / 1024 / 1024,
		MemorySysMB:   float64(memStats.Sys) / 1024 / 1024,
		MemoryHeapMB:  float64(memStats.HeapAlloc) / 1024 / 1024,
		MemoryStackMB: float64(memStats.StackInuse) / 1024 / 1024,
		GCPauseNs:     memStats.PauseNs[(memStats.NumGC+255)%256],
		GCNumGC:       memStats.NumGC,
		UptimeSeconds: int64(time.Since(metrics.StartTime).Seconds()),
	}

	// Worker pools stats
	workerPools := services.GetAllPoolStats()

	// Cache stats
	cacheStats := services.NewCacheService().GetCacheStats()

	response := MetricsResponse{
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Counters:      counters,
		Performance:   performance,
		System:        system,
		WorkerPools:   workerPools,
		Cache:         cacheStats,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}

// ExportMetrics exports all metrics as downloadable JSON
// GET /api/admin/metrics/export
func (h *AdminMetricsHandler) ExportMetrics(c *gin.Context) {
	correlationID := generateCorrelationID()
	metrics := h.observability.GetMetrics()
	health := h.observability.GetSystemHealth()
	fraud := h.observability.GetFraudInsights()
	workerPools := services.GetAllPoolStats()
	cacheStats := services.NewCacheService().GetCacheStats()

	// Build complete export
	export := map[string]interface{}{
		"correlation_id": correlationID,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"metrics":        metrics,
		"health":         health,
		"fraud_insights": fraud,
		"worker_pools":   workerPools,
		"cache":          cacheStats,
		"system": map[string]interface{}{
			"go_version":    runtime.Version(),
			"num_cpu":       runtime.NumCPU(),
			"num_goroutine": runtime.NumGoroutine(),
		},
	}

	// Add memory stats
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	export["memory"] = map[string]interface{}{
		"alloc_mb":       float64(memStats.Alloc) / 1024 / 1024,
		"total_alloc_mb": float64(memStats.TotalAlloc) / 1024 / 1024,
		"sys_mb":         float64(memStats.Sys) / 1024 / 1024,
		"heap_mb":        float64(memStats.HeapAlloc) / 1024 / 1024,
		"stack_mb":       float64(memStats.StackInuse) / 1024 / 1024,
		"gc_num":         memStats.NumGC,
	}

	// Add database stats
	if database.DB != nil {
		if sqlDB, err := database.DB.DB(); err == nil {
			stats := sqlDB.Stats()
			export["database"] = map[string]interface{}{
				"open_connections": stats.OpenConnections,
				"in_use":           stats.InUse,
				"idle":             stats.Idle,
				"max_open":         stats.MaxOpenConnections,
				"wait_count":       stats.WaitCount,
				"wait_duration":    stats.WaitDuration.String(),
			}
		}
	}

	// Add Redis stats
	if poolStats := cache.GetPoolStats(); poolStats != nil {
		export["redis"] = map[string]interface{}{
			"total_conns": poolStats.TotalConns,
			"idle_conns":  poolStats.IdleConns,
			"stale_conns": poolStats.StaleConns,
			"hits":        poolStats.Hits,
			"misses":      poolStats.Misses,
			"timeouts":    poolStats.Timeouts,
		}
	}

	// Format as JSON
	jsonData, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to export metrics",
		})
		return
	}

	// Set headers for download
	filename := "afftok-metrics-" + time.Now().Format("20060102-150405") + ".json"
	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("X-Correlation-ID", correlationID)
	c.Data(http.StatusOK, "application/json", jsonData)
}


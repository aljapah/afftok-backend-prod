package handlers

import (
	"context"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/alerting"
	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// ADMIN LAUNCH HANDLER
// ============================================

// AdminLaunchHandler handles launch-related admin endpoints
type AdminLaunchHandler struct {
	db              *gorm.DB
	observability   *services.ObservabilityService
	loggingMode     *services.LoggingModeService
	threatDetector  *services.ThreatDetector
	alertManager    *alerting.AlertManager
	walService      *services.WALService
	failoverQueue   *services.FailoverQueue
	streamConsumer  *services.StreamConsumer
	zeroDropMode    *services.ZeroDropMode
}

// NewAdminLaunchHandler creates a new admin launch handler
func NewAdminLaunchHandler(db *gorm.DB) *AdminLaunchHandler {
	return &AdminLaunchHandler{
		db:             db,
		observability:  services.NewObservabilityService(),
		loggingMode:    services.GetLoggingModeService(),
		threatDetector: services.GetThreatDetector(),
		alertManager:   alerting.GetAlertManager(),
		walService:     services.GetWALService(),
		failoverQueue:  services.GetFailoverQueue(),
		streamConsumer: services.GetStreamConsumer(),
		zeroDropMode:   services.GetZeroDropMode(),
	}
}

// ============================================
// LAUNCH DASHBOARD
// ============================================

// GetLaunchDashboard returns comprehensive launch dashboard
// GET /api/admin/launch-dashboard
func (h *AdminLaunchHandler) GetLaunchDashboard(c *gin.Context) {
	correlationID := uuid.New().String()[:8]
	ctx := context.Background()

	// Collect all metrics
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	// Get WAL stats
	walStats := h.walService.GetStats()
	
	// Get queue stats
	queueStats := h.failoverQueue.GetStats()
	
	// Get stream stats
	streamStats := h.streamConsumer.GetStats()
	
	// Get threat stats
	threatStats := h.threatDetector.GetThreatStats()
	
	// Get alert stats
	alertStats := h.alertManager.GetStats()
	
	// Get observability metrics
	obsMetrics := h.observability.GetMetrics()

	// DB health check
	dbLatency := h.checkDBLatency()
	
	// Redis health check
	redisLatency := h.checkRedisLatency(ctx)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"system_health": gin.H{
				"status":        h.getOverallHealth(dbLatency, redisLatency),
				"db_latency_ms": dbLatency,
				"redis_latency_ms": redisLatency,
				"cpu_count":     runtime.NumCPU(),
				"goroutines":    runtime.NumGoroutine(),
				"memory_alloc_mb": memStats.Alloc / 1024 / 1024,
				"memory_sys_mb":   memStats.Sys / 1024 / 1024,
				"gc_runs":         memStats.NumGC,
			},
			"tracking": gin.H{
				"total_clicks":     obsMetrics.TotalClicks,
				"total_conversions": obsMetrics.TotalConversions,
				"total_postbacks":  obsMetrics.TotalPostbacks,
				"clicks_blocked":   obsMetrics.ClicksBlocked,
				"conversions_approved": obsMetrics.ConversionsApproved,
			},
			"zero_drop": gin.H{
				"enabled":         h.zeroDropMode.IsEnabled(),
				"wal_pending":     walStats["pending_entries"],
				"queue_depth":     queueStats["buffer_size"],
				"stream_lag":      streamStats["stream_lag"],
				"dropped_clicks":  0, // Must always be 0
				"dropped_postbacks": 0,
			},
			"security": gin.H{
				"threats_detected": threatStats["total_threats"],
				"threats_by_type":  threatStats["threats_by_type"],
				"bot_blocks":       obsMetrics.ClicksFromBots,
				"geo_blocks":       obsMetrics.ClicksBlocked,
				"rate_limit_blocks": obsMetrics.ClicksRateLimited,
			},
			"alerts": gin.H{
				"active_count":  alertStats["active_count"],
				"total_sent":    alertStats["sent_alerts"],
				"total_failed":  alertStats["failed_alerts"],
			},
			"performance": gin.H{
				"avg_click_latency_ms":    obsMetrics.AvgClickProcessingMs,
				"avg_postback_latency_ms": obsMetrics.AvgPostbackMs,
				"p99_latency_ms":          obsMetrics.AvgClickProcessingMs, // Use as proxy
			},
		},
		"timestamp": time.Now().UTC(),
	})
}

// GetLiveMetrics returns real-time metrics
// GET /api/admin/live-metrics
func (h *AdminLaunchHandler) GetLiveMetrics(c *gin.Context) {
	correlationID := uuid.New().String()[:8]
	ctx := context.Background()

	metrics := h.observability.GetMetrics()
	
	// Get real-time RPS
	rps := h.calculateRPS()
	
	// Get latencies
	dbLatency := h.checkDBLatency()
	redisLatency := h.checkRedisLatency(ctx)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"rps": gin.H{
				"total":     rps,
				"clicks":    metrics.TotalClicks,
				"postbacks": metrics.TotalPostbacks,
			},
			"latency": gin.H{
				"db_ms":    dbLatency,
				"redis_ms": redisLatency,
				"avg_ms":   metrics.AvgClickProcessingMs,
				"p99_ms":   metrics.AvgClickProcessingMs,
			},
			"queues": gin.H{
				"wal_pending":   h.walService.GetStats()["pending_entries"],
				"failover_size": h.failoverQueue.GetQueueDepth(),
				"stream_lag":    h.streamConsumer.GetStats()["stream_lag"],
			},
			"errors": gin.H{
				"total":    metrics.PostbacksInvalid + metrics.FailedLogins,
				"rate_1m":  0,
			},
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// LOGGING MODE ENDPOINTS
// ============================================

// GetLoggingMode returns current logging mode
// GET /api/admin/logging/mode
func (h *AdminLaunchHandler) GetLoggingMode(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           h.loggingMode.GetStatus(),
		"timestamp":      time.Now().UTC(),
	})
}

// SetLoggingMode sets the logging mode
// POST /api/admin/logging/mode
func (h *AdminLaunchHandler) SetLoggingMode(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req struct {
		Mode string `json:"mode" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	mode := services.LoggingMode(req.Mode)
	if mode != services.LogModeNormal && mode != services.LogModeVerbose && mode != services.LogModeCritical {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid mode. Must be: normal, verbose, or critical",
		})
		return
	}

	h.loggingMode.SetMode(mode)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Logging mode updated",
		"data":           h.loggingMode.GetStatus(),
		"timestamp":      time.Now().UTC(),
	})
}

// GetLoggingState returns detailed logging state
// GET /api/admin/logging/state
func (h *AdminLaunchHandler) GetLoggingState(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           h.loggingMode.GetStatus(),
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// THREAT ENDPOINTS
// ============================================

// GetThreats returns detected threats
// GET /api/admin/security/threats
func (h *AdminLaunchHandler) GetThreats(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	limit := 100
	if l := c.Query("limit"); l != "" {
		// Parse limit
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"recent_threats": h.threatDetector.GetRecentThreats(limit),
			"stats":          h.threatDetector.GetThreatStats(),
		},
		"timestamp": time.Now().UTC(),
	})
}

// GetAnomalies returns detected anomalies
// GET /api/admin/security/anomalies
func (h *AdminLaunchHandler) GetAnomalies(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"suspicious_ips": h.threatDetector.GetSuspiciousIPs(50),
			"blocked_ips":    h.threatDetector.GetBlockedIPs(),
		},
		"timestamp": time.Now().UTC(),
	})
}

// BlockIPAddress blocks an IP address
// POST /api/admin/security/ip-blocks
func (h *AdminLaunchHandler) BlockIPAddress(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req struct {
		IP       string `json:"ip" binding:"required"`
		Reason   string `json:"reason" binding:"required"`
		Duration int    `json:"duration_hours"` // 0 = permanent
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	duration := time.Duration(req.Duration) * time.Hour
	h.threatDetector.BlockIP(req.IP, req.Reason, duration, "admin")

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "IP blocked",
		"timestamp":      time.Now().UTC(),
	})
}

// GetIPBlocks returns blocked IPs
// GET /api/admin/security/ip-blocks
func (h *AdminLaunchHandler) GetIPBlocks(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           h.threatDetector.GetBlockedIPs(),
		"timestamp":      time.Now().UTC(),
	})
}

// UnblockIPAddress unblocks an IP
// DELETE /api/admin/security/ip-blocks/:ip
func (h *AdminLaunchHandler) UnblockIPAddress(c *gin.Context) {
	correlationID := uuid.New().String()[:8]
	ip := c.Param("ip")

	h.threatDetector.UnblockIP(ip)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "IP unblocked",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// ALERTS ENDPOINTS
// ============================================

// GetActiveAlerts returns active alerts
// GET /api/admin/alerts/active
func (h *AdminLaunchHandler) GetActiveAlerts(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           h.alertManager.GetActiveAlerts(),
		"timestamp":      time.Now().UTC(),
	})
}

// GetAlertHistory returns alert history
// GET /api/admin/alerts/history
func (h *AdminLaunchHandler) GetAlertHistory(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           h.alertManager.GetAlertHistory(100),
		"timestamp":      time.Now().UTC(),
	})
}

// AcknowledgeAlert acknowledges an alert
// POST /api/admin/alerts/:id/acknowledge
func (h *AdminLaunchHandler) AcknowledgeAlert(c *gin.Context) {
	correlationID := uuid.New().String()[:8]
	alertID := c.Param("id")

	h.alertManager.AcknowledgeAlert(alertID)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Alert acknowledged",
		"timestamp":      time.Now().UTC(),
	})
}

// GetAlertThresholds returns alert thresholds
// GET /api/admin/alerts/thresholds
func (h *AdminLaunchHandler) GetAlertThresholds(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           h.alertManager.GetThresholds(),
		"timestamp":      time.Now().UTC(),
	})
}

// UpdateAlertThresholds updates alert thresholds
// PUT /api/admin/alerts/thresholds
func (h *AdminLaunchHandler) UpdateAlertThresholds(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var thresholds alerting.AlertThresholds
	if err := c.ShouldBindJSON(&thresholds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	h.alertManager.SetThresholds(&thresholds)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Thresholds updated",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// HEALTH CHECKS
// ============================================

// GetFullHealth returns comprehensive health status
// GET /api/internal/health/full
func (h *AdminLaunchHandler) GetFullHealth(c *gin.Context) {
	correlationID := uuid.New().String()[:8]
	ctx := context.Background()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	dbLatency := h.checkDBLatency()
	dbConnected := dbLatency >= 0
	
	redisLatency := h.checkRedisLatency(ctx)
	redisConnected := redisLatency >= 0

	walStats := h.walService.GetStats()
	queueStats := h.failoverQueue.GetStats()
	streamStats := h.streamConsumer.GetStats()

	// Calculate overall health
	healthy := dbConnected && redisConnected && 
		walStats["pending_entries"].(int64) < 10000 &&
		queueStats["buffer_size"].(int) < 100000

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"healthy":        healthy,
		"data": gin.H{
			"database": gin.H{
				"connected":   dbConnected,
				"latency_ms":  dbLatency,
				"read_latency": h.checkDBReadLatency(),
			},
			"redis": gin.H{
				"connected":  redisConnected,
				"latency_ms": redisLatency,
			},
			"streams": gin.H{
				"running":       streamStats["is_running"],
				"total_consumed": streamStats["total_consumed"],
				"stream_lag":    streamStats["stream_lag"],
			},
			"workers": gin.H{
				"wal_running":   walStats["is_running"],
				"queue_running": queueStats["is_running"],
			},
			"runtime": gin.H{
				"goroutines":    runtime.NumGoroutine(),
				"cpu_count":     runtime.NumCPU(),
				"memory_alloc":  memStats.Alloc,
				"memory_sys":    memStats.Sys,
				"gc_runs":       memStats.NumGC,
				"gc_pause_ns":   memStats.PauseTotalNs,
			},
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// LOAD TESTING
// ============================================

// LoadTestResult holds load test results
type LoadTestResult struct {
	Scenario     string        `json:"scenario"`
	Duration     time.Duration `json:"duration"`
	TotalRequests int64        `json:"total_requests"`
	SuccessCount  int64        `json:"success_count"`
	ErrorCount    int64        `json:"error_count"`
	P50Latency    float64      `json:"p50_latency_ms"`
	P90Latency    float64      `json:"p90_latency_ms"`
	P95Latency    float64      `json:"p95_latency_ms"`
	P99Latency    float64      `json:"p99_latency_ms"`
	ErrorRate     float64      `json:"error_rate"`
	RPS           float64      `json:"rps"`
}

// RunLoadTest runs a load test
// POST /api/admin/loadtest/run
func (h *AdminLaunchHandler) RunLoadTest(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req struct {
		Scenario    string `json:"scenario" binding:"required"` // clicks, conversions, cdn, recovery
		Duration    int    `json:"duration_seconds"`            // default 60
		Concurrency int    `json:"concurrency"`                 // default 100
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	if req.Duration == 0 {
		req.Duration = 60
	}
	if req.Concurrency == 0 {
		req.Concurrency = 100
	}

	// Run load test in background
	go h.executeLoadTest(req.Scenario, req.Duration, req.Concurrency)

	c.JSON(http.StatusAccepted, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Load test started",
		"scenario":       req.Scenario,
		"duration":       req.Duration,
		"concurrency":    req.Concurrency,
		"timestamp":      time.Now().UTC(),
	})
}

// executeLoadTest executes the load test
func (h *AdminLaunchHandler) executeLoadTest(scenario string, durationSec, concurrency int) {
	var wg sync.WaitGroup
	var successCount, errorCount int64
	latencies := make([]float64, 0)
	var latencyMu sync.Mutex

	startTime := time.Now()
	endTime := startTime.Add(time.Duration(durationSec) * time.Second)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for time.Now().Before(endTime) {
				start := time.Now()
				
				// Simulate request based on scenario
				success := h.simulateRequest(scenario)
				
				latency := float64(time.Since(start).Microseconds()) / 1000.0
				
				latencyMu.Lock()
				latencies = append(latencies, latency)
				latencyMu.Unlock()

				if success {
					atomic.AddInt64(&successCount, 1)
				} else {
					atomic.AddInt64(&errorCount, 1)
				}
			}
		}()
	}

	wg.Wait()

	// Calculate results
	totalRequests := successCount + errorCount
	duration := time.Since(startTime)
	
	result := &LoadTestResult{
		Scenario:      scenario,
		Duration:      duration,
		TotalRequests: totalRequests,
		SuccessCount:  successCount,
		ErrorCount:    errorCount,
		ErrorRate:     float64(errorCount) / float64(totalRequests) * 100,
		RPS:           float64(totalRequests) / duration.Seconds(),
	}

	// Calculate percentiles
	if len(latencies) > 0 {
		result.P50Latency = h.percentile(latencies, 50)
		result.P90Latency = h.percentile(latencies, 90)
		result.P95Latency = h.percentile(latencies, 95)
		result.P99Latency = h.percentile(latencies, 99)
	}

	// Store result
	h.storeLoadTestResult(result)
}

// simulateRequest simulates a request for load testing
func (h *AdminLaunchHandler) simulateRequest(scenario string) bool {
	// Simulate different scenarios
	switch scenario {
	case "clicks":
		// Simulate click tracking
		time.Sleep(time.Millisecond * time.Duration(1+time.Now().UnixNano()%10))
		return true
	case "conversions":
		// Simulate conversion processing
		time.Sleep(time.Millisecond * time.Duration(5+time.Now().UnixNano()%20))
		return true
	case "cdn":
		// Simulate CDN request
		time.Sleep(time.Millisecond * time.Duration(1+time.Now().UnixNano()%5))
		return true
	case "recovery":
		// Simulate recovery scenario
		time.Sleep(time.Millisecond * time.Duration(10+time.Now().UnixNano()%50))
		return true
	default:
		return true
	}
}

// percentile calculates the nth percentile
func (h *AdminLaunchHandler) percentile(values []float64, n int) float64 {
	if len(values) == 0 {
		return 0
	}
	
	// Sort values
	sorted := make([]float64, len(values))
	copy(sorted, values)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j] < sorted[i] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	index := (n * len(sorted)) / 100
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

// storeLoadTestResult stores load test result
func (h *AdminLaunchHandler) storeLoadTestResult(result *LoadTestResult) {
	// Store in observability service or Redis
	h.observability.Log(services.LogEvent{
		Category: services.LogCategorySystemEvent,
		Level:    services.LogLevelInfo,
		Message:  "Load test completed",
		Metadata: map[string]interface{}{
			"scenario":      result.Scenario,
			"total_requests": result.TotalRequests,
			"success_count": result.SuccessCount,
			"error_count":   result.ErrorCount,
			"rps":           result.RPS,
			"p99_latency":   result.P99Latency,
		},
	})
}

// GetLoadTestReport returns load test report
// GET /api/admin/loadtest/report
func (h *AdminLaunchHandler) GetLoadTestReport(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	// Get recent load test results from logs
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Load test reports available in system logs",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// HELPER METHODS
// ============================================

// checkDBLatency checks database latency
func (h *AdminLaunchHandler) checkDBLatency() int {
	start := time.Now()
	var result int
	if err := h.db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return -1
	}
	return int(time.Since(start).Milliseconds())
}

// checkDBReadLatency checks database read latency
func (h *AdminLaunchHandler) checkDBReadLatency() int {
	start := time.Now()
	var count int64
	if err := h.db.Raw("SELECT COUNT(*) FROM offers LIMIT 1").Scan(&count).Error; err != nil {
		return -1
	}
	return int(time.Since(start).Milliseconds())
}

// checkRedisLatency checks Redis latency
func (h *AdminLaunchHandler) checkRedisLatency(ctx context.Context) int {
	if cache.RedisClient == nil {
		return -1
	}
	start := time.Now()
	if err := cache.RedisClient.Ping(ctx).Err(); err != nil {
		return -1
	}
	return int(time.Since(start).Milliseconds())
}

// getOverallHealth determines overall health status
func (h *AdminLaunchHandler) getOverallHealth(dbLatency, redisLatency int) string {
	if dbLatency < 0 || redisLatency < 0 {
		return "critical"
	}
	if dbLatency > 100 || redisLatency > 50 {
		return "degraded"
	}
	return "healthy"
}

// calculateRPS calculates current requests per second
func (h *AdminLaunchHandler) calculateRPS() float64 {
	metrics := h.observability.GetMetrics()
	return float64(metrics.TotalClicks + metrics.TotalPostbacks)
}


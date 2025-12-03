package services

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/google/uuid"
)

// ============================================
// LOG CATEGORIES
// ============================================

const (
	LogCategoryClickEvent      = "click_event"
	LogCategoryPostbackEvent   = "postback_event"
	LogCategoryFraudDetection  = "fraud_detection"
	LogCategoryRateLimitBlock  = "rate_limit_block"
	LogCategoryAdminAccess     = "admin_access"
	LogCategoryAuthEvent       = "auth_event"
	LogCategoryEarningsUpdate  = "earnings_update"
	LogCategoryConversionEvent = "conversion_event"
	LogCategorySystemEvent     = "system_event"
	LogCategoryErrorEvent      = "error_event"
	LogCategoryPerformance     = "performance"
)

// Log severity levels
const (
	LogLevelDebug = "DEBUG"
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
	LogLevelFatal = "FATAL"
)

// ============================================
// STRUCTURED LOG EVENT
// ============================================

// LogEvent represents a structured log entry
type LogEvent struct {
	// Core fields
	Timestamp     time.Time `json:"timestamp"`
	Level         string    `json:"level"`
	Category      string    `json:"category"`
	Message       string    `json:"message"`
	CorrelationID string    `json:"correlation_id,omitempty"`

	// Entity IDs
	UserID      string `json:"user_id,omitempty"`
	OfferID     string `json:"offer_id,omitempty"`
	UserOfferID string `json:"user_offer_id,omitempty"`
	ClickID     string `json:"click_id,omitempty"`
	ConversionID string `json:"conversion_id,omitempty"`

	// Request context
	IP           string `json:"ip,omitempty"`
	UserAgent    string `json:"user_agent,omitempty"`
	Device       string `json:"device,omitempty"`
	TrackingCode string `json:"tracking_code,omitempty"`
	Fingerprint  string `json:"fingerprint,omitempty"`
	Endpoint     string `json:"endpoint,omitempty"`
	Method       string `json:"method,omitempty"`

	// Security/Fraud
	RiskScore    int      `json:"risk_score,omitempty"`
	BotScore     float64  `json:"bot_score,omitempty"`
	Indicators   []string `json:"indicators,omitempty"`
	IsBlocked    bool     `json:"is_blocked,omitempty"`
	BlockReason  string   `json:"block_reason,omitempty"`

	// Performance
	DurationMs   int64 `json:"duration_ms,omitempty"`
	DBQueryMs    int64 `json:"db_query_ms,omitempty"`
	RedisLatency int64 `json:"redis_latency_ms,omitempty"`

	// Error details
	ErrorCode    string `json:"error_code,omitempty"`
	ErrorMessage string `json:"error_message,omitempty"`
	StackTrace   string `json:"stack_trace,omitempty"`

	// Additional metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ============================================
// METRICS
// ============================================

// Metrics holds system-wide counters
type Metrics struct {
	// Click metrics
	TotalClicks          int64 `json:"total_clicks"`
	ClicksBlocked        int64 `json:"clicks_blocked"`
	ClicksFromBots       int64 `json:"clicks_from_bots"`
	ClicksDuplicate      int64 `json:"clicks_duplicate"`
	ClicksRateLimited    int64 `json:"clicks_rate_limited"`
	AvgClickProcessingMs int64 `json:"avg_click_processing_ms"`

	// Postback metrics
	TotalPostbacks       int64 `json:"total_postbacks"`
	PostbacksValid       int64 `json:"postbacks_valid"`
	PostbacksInvalid     int64 `json:"postbacks_invalid"`
	PostbacksReplayed    int64 `json:"postbacks_replayed"`
	AvgPostbackMs        int64 `json:"avg_postback_ms"`

	// Conversion metrics
	TotalConversions     int64 `json:"total_conversions"`
	ConversionsApproved  int64 `json:"conversions_approved"`
	ConversionsRejected  int64 `json:"conversions_rejected"`

	// Auth metrics
	TotalLogins          int64 `json:"total_logins"`
	FailedLogins         int64 `json:"failed_logins"`
	TotalRegistrations   int64 `json:"total_registrations"`

	// System metrics
	ActiveConnections    int64 `json:"active_connections"`
	DBQueryCount         int64 `json:"db_query_count"`
	DBSlowQueries        int64 `json:"db_slow_queries"`
	RedisOperations      int64 `json:"redis_operations"`
	RedisErrors          int64 `json:"redis_errors"`

	// Error metrics
	TotalErrors          int64 `json:"total_errors"`
	Error4xx             int64 `json:"error_4xx"`
	Error5xx             int64 `json:"error_5xx"`

	// Fraud metrics
	FraudAttempts        int64 `json:"fraud_attempts"`
	IPsBlocked           int64 `json:"ips_blocked"`
	BotsDetected         int64 `json:"bots_detected"`

	// Timestamps
	StartTime            time.Time `json:"start_time"`
	LastUpdated          time.Time `json:"last_updated"`
}

// ============================================
// OBSERVABILITY SERVICE
// ============================================

// ObservabilityService handles logging, metrics, and monitoring
type ObservabilityService struct {
	metrics     *Metrics
	logBuffer   []LogEvent
	bufferMutex sync.RWMutex
	maxBuffer   int
}

var (
	observabilityInstance *ObservabilityService
	observabilityOnce     sync.Once
)

// NewObservabilityService creates a singleton ObservabilityService
func NewObservabilityService() *ObservabilityService {
	observabilityOnce.Do(func() {
		observabilityInstance = &ObservabilityService{
			metrics: &Metrics{
				StartTime:   time.Now(),
				LastUpdated: time.Now(),
			},
			logBuffer: make([]LogEvent, 0, 10000),
			maxBuffer: 10000,
		}
	})
	return observabilityInstance
}

// ============================================
// LOGGING METHODS
// ============================================

// Log creates a structured log entry
func (o *ObservabilityService) Log(event LogEvent) {
	// Set defaults
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	if event.Level == "" {
		event.Level = LogLevelInfo
	}
	if event.CorrelationID == "" {
		event.CorrelationID = uuid.New().String()[:8]
	}

	// Output as JSON
	jsonBytes, _ := json.Marshal(event)
	fmt.Println(string(jsonBytes))

	// Store in buffer for recent logs endpoint
	o.bufferMutex.Lock()
	if len(o.logBuffer) >= o.maxBuffer {
		// Remove oldest 10%
		o.logBuffer = o.logBuffer[o.maxBuffer/10:]
	}
	o.logBuffer = append(o.logBuffer, event)
	o.bufferMutex.Unlock()

	// Store in Redis for persistence (optional)
	o.storeLogInRedis(event)
}

// storeLogInRedis stores log in Redis for persistence
func (o *ObservabilityService) storeLogInRedis(event LogEvent) {
	if cache.RedisClient == nil {
		return
	}

	ctx := context.Background()
	
	// Store in category-specific list
	key := fmt.Sprintf("logs:%s", event.Category)
	jsonBytes, _ := json.Marshal(event)
	
	// Use LPUSH + LTRIM to maintain a rolling log
	cache.RedisClient.LPush(ctx, key, string(jsonBytes))
	cache.RedisClient.LTrim(ctx, key, 0, 999) // Keep last 1000 per category
	cache.RedisClient.Expire(ctx, key, 24*time.Hour)

	// Store fraud events separately for analysis
	if event.Category == LogCategoryFraudDetection {
		fraudKey := fmt.Sprintf("logs:fraud:%s", time.Now().Format("2006-01-02"))
		cache.RedisClient.LPush(ctx, fraudKey, string(jsonBytes))
		cache.RedisClient.Expire(ctx, fraudKey, 7*24*time.Hour) // Keep 7 days
	}

	// Track risky IPs
	if event.RiskScore > 50 && event.IP != "" {
		ipKey := fmt.Sprintf("risky_ip:%s", event.IP)
		cache.Increment(ctx, ipKey)
		cache.RedisClient.Expire(ctx, ipKey, 24*time.Hour)
	}
}

// ============================================
// CONVENIENCE LOGGING METHODS
// ============================================

// LogClick logs a click event
func (o *ObservabilityService) LogClick(userOfferID, ip, userAgent, device, trackingCode, fingerprint string, riskScore int, blocked bool, blockReason string, durationMs int64) {
	o.Log(LogEvent{
		Level:        LogLevelInfo,
		Category:     LogCategoryClickEvent,
		Message:      "Click tracked",
		UserOfferID:  userOfferID,
		IP:           ip,
		UserAgent:    userAgent,
		Device:       device,
		TrackingCode: trackingCode,
		Fingerprint:  fingerprint,
		RiskScore:    riskScore,
		IsBlocked:    blocked,
		BlockReason:  blockReason,
		DurationMs:   durationMs,
	})

	// Update metrics
	atomic.AddInt64(&o.metrics.TotalClicks, 1)
	if blocked {
		atomic.AddInt64(&o.metrics.ClicksBlocked, 1)
	}
}

// LogPostback logs a postback event
func (o *ObservabilityService) LogPostback(userOfferID, networkID, externalID, ip string, valid bool, reason string, durationMs int64) {
	level := LogLevelInfo
	if !valid {
		level = LogLevelWarn
	}

	o.Log(LogEvent{
		Level:       level,
		Category:    LogCategoryPostbackEvent,
		Message:     "Postback received",
		UserOfferID: userOfferID,
		IP:          ip,
		IsBlocked:   !valid,
		BlockReason: reason,
		DurationMs:  durationMs,
		Metadata: map[string]interface{}{
			"network_id":  networkID,
			"external_id": externalID,
		},
	})

	// Update metrics
	atomic.AddInt64(&o.metrics.TotalPostbacks, 1)
	if valid {
		atomic.AddInt64(&o.metrics.PostbacksValid, 1)
	} else {
		atomic.AddInt64(&o.metrics.PostbacksInvalid, 1)
	}
}

// LogFraud logs a fraud detection event
func (o *ObservabilityService) LogFraud(ip, userAgent, reason string, riskScore int, botScore float64, indicators []string, metadata map[string]interface{}) {
	o.Log(LogEvent{
		Level:      LogLevelWarn,
		Category:   LogCategoryFraudDetection,
		Message:    "Fraud detected",
		IP:         ip,
		UserAgent:  userAgent,
		RiskScore:  riskScore,
		BotScore:   botScore,
		Indicators: indicators,
		IsBlocked:  true,
		BlockReason: reason,
		Metadata:   metadata,
	})

	// Update metrics
	atomic.AddInt64(&o.metrics.FraudAttempts, 1)
	if botScore > 0.8 {
		atomic.AddInt64(&o.metrics.BotsDetected, 1)
	}
}

// LogAuth logs an authentication event
func (o *ObservabilityService) LogAuth(userID, username, ip, action string, success bool, errorMsg string) {
	level := LogLevelInfo
	if !success {
		level = LogLevelWarn
	}

	o.Log(LogEvent{
		Level:        level,
		Category:     LogCategoryAuthEvent,
		Message:      fmt.Sprintf("Auth: %s", action),
		UserID:       userID,
		IP:           ip,
		IsBlocked:    !success,
		ErrorMessage: errorMsg,
		Metadata: map[string]interface{}{
			"username": username,
			"action":   action,
		},
	})

	// Update metrics
	if action == "login" {
		atomic.AddInt64(&o.metrics.TotalLogins, 1)
		if !success {
			atomic.AddInt64(&o.metrics.FailedLogins, 1)
		}
	} else if action == "register" {
		atomic.AddInt64(&o.metrics.TotalRegistrations, 1)
	}
}

// LogError logs an error event
func (o *ObservabilityService) LogError(errorCode, errorMsg, endpoint, method, ip string, statusCode int, correlationID string) {
	level := LogLevelError
	if statusCode >= 400 && statusCode < 500 {
		level = LogLevelWarn
	}

	o.Log(LogEvent{
		Level:         level,
		Category:      LogCategoryErrorEvent,
		Message:       "Error occurred",
		CorrelationID: correlationID,
		IP:            ip,
		Endpoint:      endpoint,
		Method:        method,
		ErrorCode:     errorCode,
		ErrorMessage:  errorMsg,
		Metadata: map[string]interface{}{
			"status_code": statusCode,
		},
	})

	// Update metrics
	atomic.AddInt64(&o.metrics.TotalErrors, 1)
	if statusCode >= 400 && statusCode < 500 {
		atomic.AddInt64(&o.metrics.Error4xx, 1)
	} else if statusCode >= 500 {
		atomic.AddInt64(&o.metrics.Error5xx, 1)
	}
}

// LogRateLimit logs a rate limit event
func (o *ObservabilityService) LogRateLimit(ip, endpoint, reason string) {
	o.Log(LogEvent{
		Level:       LogLevelWarn,
		Category:    LogCategoryRateLimitBlock,
		Message:     "Rate limit exceeded",
		IP:          ip,
		Endpoint:    endpoint,
		IsBlocked:   true,
		BlockReason: reason,
	})
}

// LogAdminAccess logs an admin access event
func (o *ObservabilityService) LogAdminAccess(userID, ip, endpoint, method string, success bool) {
	level := LogLevelInfo
	if !success {
		level = LogLevelWarn
	}

	o.Log(LogEvent{
		Level:     level,
		Category:  LogCategoryAdminAccess,
		Message:   "Admin access",
		UserID:    userID,
		IP:        ip,
		Endpoint:  endpoint,
		Method:    method,
		IsBlocked: !success,
	})
}

// LogConversion logs a conversion event
func (o *ObservabilityService) LogConversion(conversionID, userOfferID, userID string, amount, commission int, status string) {
	o.Log(LogEvent{
		Level:        LogLevelInfo,
		Category:     LogCategoryConversionEvent,
		Message:      "Conversion recorded",
		ConversionID: conversionID,
		UserOfferID:  userOfferID,
		UserID:       userID,
		Metadata: map[string]interface{}{
			"amount":     amount,
			"commission": commission,
			"status":     status,
		},
	})

	// Update metrics
	atomic.AddInt64(&o.metrics.TotalConversions, 1)
	if status == "approved" {
		atomic.AddInt64(&o.metrics.ConversionsApproved, 1)
	} else if status == "rejected" {
		atomic.AddInt64(&o.metrics.ConversionsRejected, 1)
	}
}

// LogPerformance logs a performance metric
func (o *ObservabilityService) LogPerformance(operation string, durationMs, dbQueryMs, redisLatencyMs int64) {
	o.Log(LogEvent{
		Level:        LogLevelDebug,
		Category:     LogCategoryPerformance,
		Message:      fmt.Sprintf("Performance: %s", operation),
		DurationMs:   durationMs,
		DBQueryMs:    dbQueryMs,
		RedisLatency: redisLatencyMs,
	})

	// Track slow queries
	if dbQueryMs > 100 {
		atomic.AddInt64(&o.metrics.DBSlowQueries, 1)
	}
}

// ============================================
// METRICS METHODS
// ============================================

// GetMetrics returns current metrics
func (o *ObservabilityService) GetMetrics() *Metrics {
	o.metrics.LastUpdated = time.Now()
	return o.metrics
}

// GetSystemHealth returns system health information
func (o *ObservabilityService) GetSystemHealth() map[string]interface{} {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	health := map[string]interface{}{
		"status":           "healthy",
		"uptime_seconds":   time.Since(o.metrics.StartTime).Seconds(),
		"goroutines":       runtime.NumGoroutine(),
		"memory_alloc_mb":  memStats.Alloc / 1024 / 1024,
		"memory_sys_mb":    memStats.Sys / 1024 / 1024,
		"gc_cycles":        memStats.NumGC,
		"cpu_cores":        runtime.NumCPU(),
	}

	// Check Redis health
	if cache.RedisClient != nil {
		ctx := context.Background()
		start := time.Now()
		if err := cache.RedisClient.Ping(ctx).Err(); err != nil {
			health["redis_status"] = "unhealthy"
			health["redis_error"] = err.Error()
		} else {
			health["redis_status"] = "healthy"
			health["redis_latency_ms"] = time.Since(start).Milliseconds()
		}
	} else {
		health["redis_status"] = "disabled"
	}

	return health
}

// ============================================
// LOG RETRIEVAL METHODS
// ============================================

// GetRecentLogs returns recent logs from buffer
func (o *ObservabilityService) GetRecentLogs(limit int, category, userID, offerID, ip string) []LogEvent {
	o.bufferMutex.RLock()
	defer o.bufferMutex.RUnlock()

	var filtered []LogEvent
	
	// Iterate in reverse (newest first)
	for i := len(o.logBuffer) - 1; i >= 0 && len(filtered) < limit; i-- {
		log := o.logBuffer[i]
		
		// Apply filters
		if category != "" && log.Category != category {
			continue
		}
		if userID != "" && log.UserID != userID {
			continue
		}
		if offerID != "" && log.OfferID != offerID {
			continue
		}
		if ip != "" && log.IP != ip {
			continue
		}
		
		filtered = append(filtered, log)
	}

	return filtered
}

// GetFraudLogs returns fraud-related logs
func (o *ObservabilityService) GetFraudLogs(limit int) []LogEvent {
	return o.GetRecentLogs(limit, LogCategoryFraudDetection, "", "", "")
}

// GetLogsFromRedis retrieves logs from Redis
func (o *ObservabilityService) GetLogsFromRedis(category string, limit int) []LogEvent {
	if cache.RedisClient == nil {
		return []LogEvent{}
	}

	ctx := context.Background()
	key := fmt.Sprintf("logs:%s", category)
	
	results, err := cache.RedisClient.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return []LogEvent{}
	}

	var logs []LogEvent
	for _, r := range results {
		var log LogEvent
		if err := json.Unmarshal([]byte(r), &log); err == nil {
			logs = append(logs, log)
		}
	}

	return logs
}

// ============================================
// FRAUD INTELLIGENCE
// ============================================

// FraudInsight represents fraud analysis insight
type FraudInsight struct {
	TopRiskyIPs       []IPRiskInfo    `json:"top_risky_ips"`
	SuspiciousPatterns []PatternInfo  `json:"suspicious_patterns"`
	DailyOffenders    []OffenderInfo  `json:"daily_offenders"`
	Summary           FraudSummary    `json:"summary"`
}

type IPRiskInfo struct {
	IP          string `json:"ip"`
	Attempts    int64  `json:"attempts"`
	RiskScore   int    `json:"risk_score"`
	LastSeen    string `json:"last_seen"`
}

type PatternInfo struct {
	Pattern     string `json:"pattern"`
	Count       int64  `json:"count"`
	Description string `json:"description"`
}

type OffenderInfo struct {
	IP          string `json:"ip"`
	UserAgent   string `json:"user_agent"`
	Attempts    int64  `json:"attempts"`
	BlockedAt   string `json:"blocked_at"`
}

type FraudSummary struct {
	TotalAttempts24h  int64   `json:"total_attempts_24h"`
	UniqueIPs         int     `json:"unique_ips"`
	BotsBlocked       int64   `json:"bots_blocked"`
	ReplayAttempts    int64   `json:"replay_attempts"`
	BlockRate         float64 `json:"block_rate_percent"`
}

// GetFraudInsights returns fraud intelligence data
func (o *ObservabilityService) GetFraudInsights() FraudInsight {
	insight := FraudInsight{
		TopRiskyIPs:       []IPRiskInfo{},
		SuspiciousPatterns: []PatternInfo{},
		DailyOffenders:    []OffenderInfo{},
	}

	// Get from Redis if available
	if cache.RedisClient != nil {
		ctx := context.Background()
		
		// Get risky IPs
		keys, _ := cache.RedisClient.Keys(ctx, "risky_ip:*").Result()
		for _, key := range keys {
			ip := key[len("risky_ip:"):]
			count, _ := cache.RedisClient.Get(ctx, key).Int64()
			if count > 5 {
				insight.TopRiskyIPs = append(insight.TopRiskyIPs, IPRiskInfo{
					IP:       ip,
					Attempts: count,
				})
			}
		}
	}

	// Calculate summary from metrics
	insight.Summary = FraudSummary{
		TotalAttempts24h: o.metrics.FraudAttempts,
		BotsBlocked:      o.metrics.BotsDetected,
		ReplayAttempts:   o.metrics.PostbacksReplayed,
	}

	if o.metrics.TotalClicks > 0 {
		insight.Summary.BlockRate = float64(o.metrics.ClicksBlocked) / float64(o.metrics.TotalClicks) * 100
	}

	return insight
}

// ============================================
// REDIS DIAGNOSTICS
// ============================================

// RedisDiagnostics contains Redis health information
type RedisDiagnostics struct {
	Connected     bool              `json:"connected"`
	MemoryUsedMB  float64           `json:"memory_used_mb"`
	KeyCount      int64             `json:"key_count"`
	KeysByPrefix  map[string]int64  `json:"keys_by_prefix"`
	LatencyMs     int64             `json:"latency_ms"`
	Errors        []string          `json:"errors,omitempty"`
}

// GetRedisDiagnostics returns Redis health information
func (o *ObservabilityService) GetRedisDiagnostics() RedisDiagnostics {
	diag := RedisDiagnostics{
		Connected:    false,
		KeysByPrefix: make(map[string]int64),
	}

	if cache.RedisClient == nil {
		diag.Errors = append(diag.Errors, "Redis client not initialized")
		return diag
	}

	ctx := context.Background()
	
	// Test connection
	start := time.Now()
	if err := cache.RedisClient.Ping(ctx).Err(); err != nil {
		diag.Errors = append(diag.Errors, err.Error())
		return diag
	}
	diag.Connected = true
	diag.LatencyMs = time.Since(start).Milliseconds()

	// Get memory info
	if _, memErr := cache.RedisClient.Info(ctx, "memory").Result(); memErr == nil {
		// Parse used_memory from info string
		// This is simplified; in production, parse properly
		diag.MemoryUsedMB = 0 // Would parse from info
	}

	// Count keys by prefix
	prefixes := []string{"tracking:", "click_fp:", "ratelimit:", "logs:", "user_stats:", "risky_ip:"}
	for _, prefix := range prefixes {
		keys, _ := cache.RedisClient.Keys(ctx, prefix+"*").Result()
		diag.KeysByPrefix[prefix] = int64(len(keys))
	}

	// Total key count
	dbSize, _ := cache.RedisClient.DBSize(ctx).Result()
	diag.KeyCount = dbSize

	return diag
}

// ============================================
// CORRELATION ID
// ============================================

// GenerateCorrelationID creates a unique correlation ID for request tracing
func GenerateCorrelationID() string {
	return uuid.New().String()[:12]
}


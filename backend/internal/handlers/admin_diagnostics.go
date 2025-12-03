package handlers

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/database"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
)

// AdminDiagnosticsHandler handles admin diagnostics API endpoints
type AdminDiagnosticsHandler struct {
	observability *services.ObservabilityService
}

// NewAdminDiagnosticsHandler creates a new admin diagnostics handler
func NewAdminDiagnosticsHandler() *AdminDiagnosticsHandler {
	return &AdminDiagnosticsHandler{
		observability: services.NewObservabilityService(),
	}
}

// RedisDiagnosticsResponse represents Redis diagnostics
type RedisDiagnosticsResponse struct {
	CorrelationID string                 `json:"correlation_id"`
	Timestamp     string                 `json:"timestamp"`
	Connected     bool                   `json:"connected"`
	LatencyMs     int64                  `json:"latency_ms"`
	Memory        RedisMemoryInfo        `json:"memory"`
	Keys          RedisKeysInfo          `json:"keys"`
	Pool          RedisPoolInfo          `json:"pool"`
	Info          map[string]interface{} `json:"info"`
	Errors        []string               `json:"errors,omitempty"`
}

// RedisMemoryInfo represents Redis memory info
type RedisMemoryInfo struct {
	UsedMemory       string `json:"used_memory"`
	UsedMemoryPeak   string `json:"used_memory_peak"`
	UsedMemoryRSS    string `json:"used_memory_rss"`
	MaxMemory        string `json:"max_memory"`
	MemoryFragRatio  string `json:"memory_frag_ratio"`
}

// RedisKeysInfo represents Redis keys info
type RedisKeysInfo struct {
	TotalKeys    int64            `json:"total_keys"`
	KeysByPrefix map[string]int64 `json:"keys_by_prefix"`
	ExpiringKeys int64            `json:"expiring_keys"`
}

// RedisPoolInfo represents Redis pool info
type RedisPoolInfo struct {
	TotalConns uint32 `json:"total_conns"`
	IdleConns  uint32 `json:"idle_conns"`
	StaleConns uint32 `json:"stale_conns"`
	Hits       uint32 `json:"hits"`
	Misses     uint32 `json:"misses"`
	Timeouts   uint32 `json:"timeouts"`
}

// GetRedisDiagnostics returns Redis diagnostics
// GET /api/admin/diagnostics/redis
func (h *AdminDiagnosticsHandler) GetRedisDiagnostics(c *gin.Context) {
	correlationID := generateCorrelationID()

	response := RedisDiagnosticsResponse{
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Connected:     false,
		Errors:        make([]string, 0),
		Keys: RedisKeysInfo{
			KeysByPrefix: make(map[string]int64),
		},
		Info: make(map[string]interface{}),
	}

	if cache.RedisClient == nil {
		response.Errors = append(response.Errors, "Redis client not initialized")
		c.JSON(http.StatusOK, gin.H{
			"success":        true,
			"correlation_id": correlationID,
			"data":           response,
		})
		return
	}

	ctx := context.Background()

	// Test connection
	start := time.Now()
	if err := cache.RedisClient.Ping(ctx).Err(); err != nil {
		response.Errors = append(response.Errors, "Ping failed: "+err.Error())
	} else {
		response.Connected = true
	}
	response.LatencyMs = time.Since(start).Milliseconds()

	// Get memory info
	if memInfo, err := cache.RedisClient.Info(ctx, "memory").Result(); err == nil {
		response.Memory = parseRedisMemoryInfo(memInfo)
	}

	// Get key counts by prefix
	prefixes := []string{
		services.NSTracking,
		services.NSStats,
		services.NSOffer,
		services.NSUser,
		services.NSFraud,
		services.NSLogs,
		services.NSRateLimit,
		services.NSCounter,
	}

	for _, prefix := range prefixes {
		keys, err := cache.RedisClient.Keys(ctx, prefix+"*").Result()
		if err == nil {
			response.Keys.KeysByPrefix[prefix] = int64(len(keys))
		}
	}

	// Get total keys
	if dbSize, err := cache.RedisClient.DBSize(ctx).Result(); err == nil {
		response.Keys.TotalKeys = dbSize
	}

	// Get pool stats
	if poolStats := cache.GetPoolStats(); poolStats != nil {
		response.Pool = RedisPoolInfo{
			TotalConns: poolStats.TotalConns,
			IdleConns:  poolStats.IdleConns,
			StaleConns: poolStats.StaleConns,
			Hits:       poolStats.Hits,
			Misses:     poolStats.Misses,
			Timeouts:   poolStats.Timeouts,
		}
	}

	// Get server info
	if serverInfo, err := cache.RedisClient.Info(ctx, "server").Result(); err == nil {
		response.Info["redis_version"] = parseRedisInfoValue(serverInfo, "redis_version")
		response.Info["uptime_seconds"] = parseRedisInfoValue(serverInfo, "uptime_in_seconds")
		response.Info["connected_clients"] = parseRedisInfoValue(serverInfo, "connected_clients")
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}

// DatabaseDiagnosticsResponse represents database diagnostics
type DatabaseDiagnosticsResponse struct {
	CorrelationID string                 `json:"correlation_id"`
	Timestamp     string                 `json:"timestamp"`
	Connected     bool                   `json:"connected"`
	LatencyMs     int64                  `json:"latency_ms"`
	Version       string                 `json:"version"`
	Pool          DatabasePoolInfo       `json:"pool"`
	Tables        []TableInfo            `json:"tables"`
	Indexes       []IndexInfo            `json:"indexes"`
	Stats         map[string]interface{} `json:"stats"`
	Errors        []string               `json:"errors,omitempty"`
}

// DatabasePoolInfo represents database pool info
type DatabasePoolInfo struct {
	MaxOpen           int    `json:"max_open"`
	Open              int    `json:"open"`
	InUse             int    `json:"in_use"`
	Idle              int    `json:"idle"`
	WaitCount         int64  `json:"wait_count"`
	WaitDuration      string `json:"wait_duration"`
	MaxIdleClosed     int64  `json:"max_idle_closed"`
	MaxLifetimeClosed int64  `json:"max_lifetime_closed"`
}

// TableInfo represents table information
type TableInfo struct {
	Name       string `json:"name"`
	RowCount   int64  `json:"row_count"`
	SizeBytes  int64  `json:"size_bytes"`
	SizeHuman  string `json:"size_human"`
}

// IndexInfo represents index information
type IndexInfo struct {
	TableName  string `json:"table_name"`
	IndexName  string `json:"index_name"`
	IndexType  string `json:"index_type"`
	IsUnique   bool   `json:"is_unique"`
	SizeBytes  int64  `json:"size_bytes"`
}

// GetDBDiagnostics returns database diagnostics
// GET /api/admin/diagnostics/db
func (h *AdminDiagnosticsHandler) GetDBDiagnostics(c *gin.Context) {
	correlationID := generateCorrelationID()

	response := DatabaseDiagnosticsResponse{
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Connected:     false,
		Tables:        make([]TableInfo, 0),
		Indexes:       make([]IndexInfo, 0),
		Stats:         make(map[string]interface{}),
		Errors:        make([]string, 0),
	}

	if database.DB == nil {
		response.Errors = append(response.Errors, "Database not initialized")
		c.JSON(http.StatusOK, gin.H{
			"success":        true,
			"correlation_id": correlationID,
			"data":           response,
		})
		return
	}

	sqlDB, err := database.DB.DB()
	if err != nil {
		response.Errors = append(response.Errors, "Failed to get DB connection: "+err.Error())
		c.JSON(http.StatusOK, gin.H{
			"success":        true,
			"correlation_id": correlationID,
			"data":           response,
		})
		return
	}

	// Test connection
	start := time.Now()
	if err := sqlDB.Ping(); err != nil {
		response.Errors = append(response.Errors, "Ping failed: "+err.Error())
	} else {
		response.Connected = true
	}
	response.LatencyMs = time.Since(start).Milliseconds()

	// Get version
	var version string
	database.DB.Raw("SELECT version()").Scan(&version)
	response.Version = version

	// Get pool stats
	stats := sqlDB.Stats()
	response.Pool = DatabasePoolInfo{
		MaxOpen:           stats.MaxOpenConnections,
		Open:              stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration.String(),
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
	}

	// Get table sizes
	tables := []string{"clicks", "conversions", "user_offers", "offers", "afftok_users", "tracking_events"}
	for _, table := range tables {
		var count int64
		database.DB.Table(table).Count(&count)
		
		var sizeBytes int64
		database.DB.Raw("SELECT pg_total_relation_size(?)", table).Scan(&sizeBytes)
		
		response.Tables = append(response.Tables, TableInfo{
			Name:      table,
			RowCount:  count,
			SizeBytes: sizeBytes,
			SizeHuman: formatBytes(sizeBytes),
		})
	}

	// Get index info
	type indexRow struct {
		TableName string
		IndexName string
		IndexType string
		IsUnique  bool
	}
	var indexes []indexRow
	database.DB.Raw(`
		SELECT 
			tablename as table_name,
			indexname as index_name,
			indexdef as index_type,
			indisunique as is_unique
		FROM pg_indexes 
		JOIN pg_class ON pg_indexes.indexname = pg_class.relname
		JOIN pg_index ON pg_class.oid = pg_index.indexrelid
		WHERE schemaname = 'public'
		ORDER BY tablename, indexname
	`).Scan(&indexes)

	for _, idx := range indexes {
		response.Indexes = append(response.Indexes, IndexInfo{
			TableName: idx.TableName,
			IndexName: idx.IndexName,
			IndexType: idx.IndexType,
			IsUnique:  idx.IsUnique,
		})
	}

	// Get additional stats
	var dbSize int64
	database.DB.Raw("SELECT pg_database_size(current_database())").Scan(&dbSize)
	response.Stats["database_size_bytes"] = dbSize
	response.Stats["database_size_human"] = formatBytes(dbSize)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}

// parseRedisMemoryInfo parses Redis memory info from INFO output
func parseRedisMemoryInfo(info string) RedisMemoryInfo {
	return RedisMemoryInfo{
		UsedMemory:      parseRedisInfoValue(info, "used_memory_human"),
		UsedMemoryPeak:  parseRedisInfoValue(info, "used_memory_peak_human"),
		UsedMemoryRSS:   parseRedisInfoValue(info, "used_memory_rss_human"),
		MaxMemory:       parseRedisInfoValue(info, "maxmemory_human"),
		MemoryFragRatio: parseRedisInfoValue(info, "mem_fragmentation_ratio"),
	}
}

// parseRedisInfoValue extracts a value from Redis INFO output
func parseRedisInfoValue(info, key string) string {
	lines := splitInfoLines(info)
	for _, line := range lines {
		if len(line) > len(key)+1 && line[:len(key)] == key {
			return line[len(key)+1:]
		}
	}
	return ""
}

// formatBytes formats bytes to human readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return string(rune(bytes)) + " B"
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return string(rune(bytes/div)) + " " + string("KMGTPE"[exp]) + "B"
}

// GetSystemDiagnostics returns system diagnostics
// GET /api/admin/diagnostics/system
func (h *AdminDiagnosticsHandler) GetSystemDiagnostics(c *gin.Context) {
	correlationID := generateCorrelationID()

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	response := map[string]interface{}{
		"correlation_id": correlationID,
		"timestamp":      time.Now().UTC().Format(time.RFC3339),
		"runtime": map[string]interface{}{
			"go_version":    runtime.Version(),
			"go_os":         runtime.GOOS,
			"go_arch":       runtime.GOARCH,
			"num_cpu":       runtime.NumCPU(),
			"num_goroutine": runtime.NumGoroutine(),
			"gomaxprocs":    runtime.GOMAXPROCS(0),
		},
		"memory": map[string]interface{}{
			"alloc_mb":       float64(memStats.Alloc) / 1024 / 1024,
			"total_alloc_mb": float64(memStats.TotalAlloc) / 1024 / 1024,
			"sys_mb":         float64(memStats.Sys) / 1024 / 1024,
			"heap_alloc_mb":  float64(memStats.HeapAlloc) / 1024 / 1024,
			"heap_sys_mb":    float64(memStats.HeapSys) / 1024 / 1024,
			"heap_idle_mb":   float64(memStats.HeapIdle) / 1024 / 1024,
			"heap_inuse_mb":  float64(memStats.HeapInuse) / 1024 / 1024,
			"heap_released_mb": float64(memStats.HeapReleased) / 1024 / 1024,
			"heap_objects":   memStats.HeapObjects,
			"stack_inuse_mb": float64(memStats.StackInuse) / 1024 / 1024,
			"stack_sys_mb":   float64(memStats.StackSys) / 1024 / 1024,
		},
		"gc": map[string]interface{}{
			"num_gc":        memStats.NumGC,
			"pause_total_ns": memStats.PauseTotalNs,
			"pause_ns":      memStats.PauseNs[(memStats.NumGC+255)%256],
			"gc_cpu_fraction": memStats.GCCPUFraction,
			"enable_gc":     memStats.EnableGC,
		},
		"worker_pools": services.GetAllPoolStats(),
		"cache":        services.NewCacheService().GetCacheStats(),
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}


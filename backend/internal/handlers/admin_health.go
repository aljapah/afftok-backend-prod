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

// AdminHealthHandler handles admin health API endpoints
type AdminHealthHandler struct {
	observability *services.ObservabilityService
}

// NewAdminHealthHandler creates a new admin health handler
func NewAdminHealthHandler() *AdminHealthHandler {
	return &AdminHealthHandler{
		observability: services.NewObservabilityService(),
	}
}

// HealthResponse represents complete health status
type HealthResponse struct {
	CorrelationID string        `json:"correlation_id"`
	Timestamp     string        `json:"timestamp"`
	Status        string        `json:"status"`
	Database      DatabaseHealth `json:"database"`
	Redis         RedisHealth    `json:"redis"`
	Runtime       RuntimeHealth  `json:"runtime"`
	Checks        []HealthCheck  `json:"checks"`
}

// DatabaseHealth represents database health
type DatabaseHealth struct {
	Healthy       bool   `json:"healthy"`
	LatencyMs     int64  `json:"latency_ms"`
	Message       string `json:"message,omitempty"`
	Version       string `json:"version,omitempty"`
}

// RedisHealth represents Redis health
type RedisHealth struct {
	Healthy   bool   `json:"healthy"`
	LatencyMs int64  `json:"latency_ms"`
	Message   string `json:"message,omitempty"`
	Version   string `json:"version,omitempty"`
}

// RuntimeHealth represents Go runtime health
type RuntimeHealth struct {
	Healthy       bool    `json:"healthy"`
	Goroutines    int     `json:"goroutines"`
	MemoryMB      float64 `json:"memory_mb"`
	GCPauseMs     float64 `json:"gc_pause_ms"`
	Message       string  `json:"message,omitempty"`
}

// HealthCheck represents a single health check
type HealthCheck struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	LatencyMs int64  `json:"latency_ms,omitempty"`
	Message   string `json:"message,omitempty"`
}

// GetHealth returns complete system health
// GET /api/admin/health
func (h *AdminHealthHandler) GetHealth(c *gin.Context) {
	correlationID := generateCorrelationID()
	
	response := HealthResponse{
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
		Status:        "healthy",
		Checks:        make([]HealthCheck, 0),
	}

	// Check Database
	dbCheck := HealthCheck{Name: "database", Status: "healthy"}
	dbStart := time.Now()
	
	if database.DB != nil {
		if sqlDB, err := database.DB.DB(); err == nil {
			if err := sqlDB.Ping(); err != nil {
				dbCheck.Status = "unhealthy"
				dbCheck.Message = err.Error()
				response.Status = "degraded"
			}
			dbCheck.LatencyMs = time.Since(dbStart).Milliseconds()
			
			// Get version
			var version string
			database.DB.Raw("SELECT version()").Scan(&version)
			response.Database = DatabaseHealth{
				Healthy:   dbCheck.Status == "healthy",
				LatencyMs: dbCheck.LatencyMs,
				Version:   version,
			}
		} else {
			dbCheck.Status = "unhealthy"
			dbCheck.Message = "Failed to get DB connection"
			response.Status = "unhealthy"
		}
	} else {
		dbCheck.Status = "unhealthy"
		dbCheck.Message = "Database not initialized"
		response.Status = "unhealthy"
	}
	response.Checks = append(response.Checks, dbCheck)

	// Check Redis
	redisCheck := HealthCheck{Name: "redis", Status: "healthy"}
	redisStart := time.Now()
	
	if cache.RedisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		
		if err := cache.RedisClient.Ping(ctx).Err(); err != nil {
			redisCheck.Status = "unhealthy"
			redisCheck.Message = err.Error()
			response.Status = "degraded"
		}
		redisCheck.LatencyMs = time.Since(redisStart).Milliseconds()
		
		// Get version
		var version string
		if info, err := cache.RedisClient.Info(ctx, "server").Result(); err == nil {
			// Parse redis_version from info
			version = parseRedisVersion(info)
		}
		
		response.Redis = RedisHealth{
			Healthy:   redisCheck.Status == "healthy",
			LatencyMs: redisCheck.LatencyMs,
			Version:   version,
		}
	} else {
		redisCheck.Status = "unhealthy"
		redisCheck.Message = "Redis not initialized"
		response.Status = "degraded"
	}
	response.Checks = append(response.Checks, redisCheck)

	// Check Runtime
	runtimeCheck := HealthCheck{Name: "runtime", Status: "healthy"}
	
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	
	goroutines := runtime.NumGoroutine()
	memoryMB := float64(memStats.Alloc) / 1024 / 1024
	gcPauseMs := float64(memStats.PauseNs[(memStats.NumGC+255)%256]) / 1000000
	
	// Check for warning conditions
	if goroutines > 10000 {
		runtimeCheck.Status = "warning"
		runtimeCheck.Message = "High goroutine count"
	}
	if memoryMB > 1024 {
		runtimeCheck.Status = "warning"
		runtimeCheck.Message = "High memory usage"
	}
	
	response.Runtime = RuntimeHealth{
		Healthy:    runtimeCheck.Status == "healthy",
		Goroutines: goroutines,
		MemoryMB:   memoryMB,
		GCPauseMs:  gcPauseMs,
		Message:    runtimeCheck.Message,
	}
	response.Checks = append(response.Checks, runtimeCheck)

	// Set HTTP status based on health
	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}

// ConnectionsResponse represents connection pool status
type ConnectionsResponse struct {
	CorrelationID string             `json:"correlation_id"`
	Timestamp     string             `json:"timestamp"`
	Database      DatabaseConnStatus `json:"database"`
	Redis         RedisConnStatus    `json:"redis"`
}

// DatabaseConnStatus represents database connection status
type DatabaseConnStatus struct {
	MaxOpen         int    `json:"max_open"`
	Open            int    `json:"open"`
	InUse           int    `json:"in_use"`
	Idle            int    `json:"idle"`
	WaitCount       int64  `json:"wait_count"`
	WaitDuration    string `json:"wait_duration"`
	MaxIdleClosed   int64  `json:"max_idle_closed"`
	MaxLifetimeClosed int64 `json:"max_lifetime_closed"`
}

// RedisConnStatus represents Redis connection status
type RedisConnStatus struct {
	PoolSize    uint32 `json:"pool_size"`
	TotalConns  uint32 `json:"total_conns"`
	IdleConns   uint32 `json:"idle_conns"`
	StaleConns  uint32 `json:"stale_conns"`
	Hits        uint32 `json:"hits"`
	Misses      uint32 `json:"misses"`
	Timeouts    uint32 `json:"timeouts"`
}

// GetConnections returns connection pool status
// GET /api/admin/connections
func (h *AdminHealthHandler) GetConnections(c *gin.Context) {
	correlationID := generateCorrelationID()

	response := ConnectionsResponse{
		CorrelationID: correlationID,
		Timestamp:     time.Now().UTC().Format(time.RFC3339),
	}

	// Database connections
	if database.DB != nil {
		if sqlDB, err := database.DB.DB(); err == nil {
			stats := sqlDB.Stats()
			response.Database = DatabaseConnStatus{
				MaxOpen:           stats.MaxOpenConnections,
				Open:              stats.OpenConnections,
				InUse:             stats.InUse,
				Idle:              stats.Idle,
				WaitCount:         stats.WaitCount,
				WaitDuration:      stats.WaitDuration.String(),
				MaxIdleClosed:     stats.MaxIdleClosed,
				MaxLifetimeClosed: stats.MaxLifetimeClosed,
			}
		}
	}

	// Redis connections
	if poolStats := cache.GetPoolStats(); poolStats != nil {
		response.Redis = RedisConnStatus{
			TotalConns: poolStats.TotalConns,
			IdleConns:  poolStats.IdleConns,
			StaleConns: poolStats.StaleConns,
			Hits:       poolStats.Hits,
			Misses:     poolStats.Misses,
			Timeouts:   poolStats.Timeouts,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}

// parseRedisVersion extracts Redis version from INFO output
func parseRedisVersion(info string) string {
	lines := splitInfoLines(info)
	for _, line := range lines {
		if len(line) > 14 && line[:14] == "redis_version:" {
			return line[14:]
		}
	}
	return ""
}

func splitInfoLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			line := s[start:i]
			if len(line) > 0 && line[len(line)-1] == '\r' {
				line = line[:len(line)-1]
			}
			lines = append(lines, line)
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}


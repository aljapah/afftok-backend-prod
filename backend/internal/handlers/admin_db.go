package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/database"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ============================================
// ADMIN DATABASE HANDLER
// ============================================

// AdminDBHandler handles database administration endpoints
type AdminDBHandler struct {
	dbStats   *services.DBStatsService
	partition *services.PartitionService
}

// NewAdminDBHandler creates a new admin DB handler
func NewAdminDBHandler(dbStats *services.DBStatsService, partition *services.PartitionService) *AdminDBHandler {
	return &AdminDBHandler{
		dbStats:   dbStats,
		partition: partition,
	}
}

// ============================================
// BACKUP INFO (PITR)
// ============================================

// GetBackupInfo returns backup/PITR configuration info
// GET /api/admin/db/backup-info
func (h *AdminDBHandler) GetBackupInfo(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	info, err := h.dbStats.GetBackupInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to get backup info: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           info,
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// VACUUM / ANALYZE
// ============================================

// GetVacuumPlan returns a recommended vacuum/analyze plan
// GET /api/admin/db/vacuum-plan
func (h *AdminDBHandler) GetVacuumPlan(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	plan, err := h.dbStats.GetVacuumPlan()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to get vacuum plan: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           plan,
		"timestamp":      time.Now().UTC(),
	})
}

// GetTableStats returns statistics for all tables
// GET /api/admin/db/stats
func (h *AdminDBHandler) GetTableStats(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	stats, err := h.dbStats.GetTableStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to get table stats: " + err.Error(),
		})
		return
	}

	// Get DB size
	dbSize, _ := h.dbStats.GetDBSize()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"tables":       stats,
			"table_count":  len(stats),
			"database":     dbSize,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// INDEX PROFILING
// ============================================

// GetIndexes returns index statistics
// GET /api/admin/db/indexes
func (h *AdminDBHandler) GetIndexes(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	indexes, err := h.dbStats.GetIndexStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to get index stats: " + err.Error(),
		})
		return
	}

	// Get unused indexes
	unused, _ := h.dbStats.GetUnusedIndexes()

	// Calculate totals
	var totalSizeMB float64
	var totalScans int64
	for _, idx := range indexes {
		totalSizeMB += idx.SizeMB
		totalScans += idx.IdxScans
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"indexes":         indexes,
			"index_count":     len(indexes),
			"unused_indexes":  unused,
			"unused_count":    len(unused),
			"total_size_mb":   totalSizeMB,
			"total_scans":     totalScans,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// PARTITIONING
// ============================================

// GetPartitionStatus returns partitioning status for clicks table
// GET /api/admin/db/partitions
func (h *AdminDBHandler) GetPartitionStatus(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	status, err := h.partition.GetPartitionStatus("clicks")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to get partition status: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           status,
		"timestamp":      time.Now().UTC(),
	})
}

// CreatePartition creates a monthly partition for clicks
// POST /api/admin/db/partition/create?year=2025&month=01
func (h *AdminDBHandler) CreatePartition(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	yearStr := c.Query("year")
	monthStr := c.Query("month")

	if yearStr == "" || monthStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "year and month query parameters are required",
		})
		return
	}

	year, err := strconv.Atoi(yearStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid year format",
		})
		return
	}

	month, err := strconv.Atoi(monthStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid month format",
		})
		return
	}

	result, err := h.partition.CreateMonthlyPartition(year, month)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to create partition: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        result.Success,
		"correlation_id": correlationID,
		"data":           result,
		"timestamp":      time.Now().UTC(),
	})
}

// EnsurePartitions ensures partitions exist for current and future months
// POST /api/admin/db/partitions/ensure
func (h *AdminDBHandler) EnsurePartitions(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	results, err := h.partition.EnsurePartitionsExist()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to ensure partitions: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           results,
		"timestamp":      time.Now().UTC(),
	})
}

// GetMigrationPlan returns a plan for migrating clicks to partitioned table
// GET /api/admin/db/partition/migration-plan
func (h *AdminDBHandler) GetMigrationPlan(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	plan, err := h.partition.GetMigrationPlan()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to get migration plan: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           plan,
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// LATENCY
// ============================================

// GetDBLatency returns database latency statistics
// GET /api/admin/db/latency
func (h *AdminDBHandler) GetDBLatency(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	stats := database.GetLatencyStats()

	// Also get live latency
	start := time.Now()
	var result int
	database.DB.Raw("SELECT 1").Scan(&result)
	liveLatency := time.Since(start).Milliseconds()

	// Get replica status if available
	var replicaStatus interface{}
	if router := database.GetDBRouter(); router != nil {
		replicaStatus = router.GetReplicaStatus()
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"live_latency_ms": liveLatency,
			"read_stats":      stats["read"],
			"write_stats":     stats["write"],
			"replica_status":  replicaStatus,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// SLOW QUERIES
// ============================================

// GetSlowQueries returns slow queries from pg_stat_statements
// GET /api/admin/db/slow-queries?limit=20
func (h *AdminDBHandler) GetSlowQueries(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	limitStr := c.DefaultQuery("limit", "20")
	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	queries, err := h.dbStats.GetSlowQueries(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          err.Error(),
			"note":           "pg_stat_statements extension may not be installed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"queries":     queries,
			"query_count": len(queries),
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// CONNECTION POOL
// ============================================

// GetConnectionPool returns connection pool statistics
// GET /api/admin/db/pool
func (h *AdminDBHandler) GetConnectionPool(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	sqlDB, err := database.DB.DB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to get DB stats: " + err.Error(),
		})
		return
	}

	stats := sqlDB.Stats()

	// Get replica pool if available
	var replicaStats interface{}
	if router := database.GetDBRouter(); router != nil && router.HasReplica() {
		if replicaDB := router.Replica(); replicaDB != nil {
			if sqlReplica, err := replicaDB.DB(); err == nil {
				replicaStats = sqlReplica.Stats()
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"primary": gin.H{
				"max_open_connections": stats.MaxOpenConnections,
				"open_connections":     stats.OpenConnections,
				"in_use":               stats.InUse,
				"idle":                 stats.Idle,
				"wait_count":           stats.WaitCount,
				"wait_duration_ms":     stats.WaitDuration.Milliseconds(),
				"max_idle_closed":      stats.MaxIdleClosed,
				"max_idle_time_closed": stats.MaxIdleTimeClosed,
				"max_lifetime_closed":  stats.MaxLifetimeClosed,
			},
			"replica": replicaStats,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// DATABASE SIZE
// ============================================

// GetDBSize returns database size information
// GET /api/admin/db/size
func (h *AdminDBHandler) GetDBSize(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	size, err := h.dbStats.GetDBSize()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to get DB size: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           size,
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// FULL DB REPORT
// ============================================

// GetDBReport returns a comprehensive database report
// GET /api/admin/db/report
func (h *AdminDBHandler) GetDBReport(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	// Gather all data
	tableStats, _ := h.dbStats.GetTableStats()
	indexStats, _ := h.dbStats.GetIndexStats()
	unusedIndexes, _ := h.dbStats.GetUnusedIndexes()
	vacuumPlan, _ := h.dbStats.GetVacuumPlan()
	backupInfo, _ := h.dbStats.GetBackupInfo()
	dbSize, _ := h.dbStats.GetDBSize()
	partitionStatus, _ := h.partition.GetPartitionStatus("clicks")
	latencyStats := database.GetLatencyStats()

	// Get connection pool stats
	var poolStats interface{}
	if sqlDB, err := database.DB.DB(); err == nil {
		poolStats = sqlDB.Stats()
	}

	// Get replica status
	var replicaStatus interface{}
	if router := database.GetDBRouter(); router != nil {
		replicaStatus = router.GetReplicaStatus()
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"generated_at":     time.Now().UTC(),
			"database_size":    dbSize,
			"tables":           tableStats,
			"indexes":          indexStats,
			"unused_indexes":   unusedIndexes,
			"vacuum_plan":      vacuumPlan,
			"backup_info":      backupInfo,
			"partitions":       partitionStatus,
			"connection_pool":  poolStats,
			"replica_status":   replicaStatus,
			"latency_stats":    latencyStats,
		},
		"timestamp": time.Now().UTC(),
	})
}


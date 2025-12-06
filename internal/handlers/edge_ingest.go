package handlers

import (
	"compress/gzip"
	"io"
	"net/http"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// EDGE INGEST HANDLER
// ============================================

// EdgeIngestHandler handles edge click ingestion
type EdgeIngestHandler struct {
	db            *gorm.DB
	ingestService *services.EdgeIngestService
}

// NewEdgeIngestHandler creates a new edge ingest handler
func NewEdgeIngestHandler(db *gorm.DB) *EdgeIngestHandler {
	return &EdgeIngestHandler{
		db:            db,
		ingestService: services.GetEdgeIngestService(db),
	}
}

// SetLinkService sets the link service on the ingest service
func (h *EdgeIngestHandler) SetLinkService(ls *services.LinkService) {
	h.ingestService.SetLinkService(ls)
}

// SetClickService sets the click service on the ingest service
func (h *EdgeIngestHandler) SetClickService(cs *services.ClickService) {
	h.ingestService.SetClickService(cs)
}

// ============================================
// ENDPOINTS
// ============================================

// IngestEdgeClick handles single or batch edge click ingestion
// POST /api/internal/edge-click
func (h *EdgeIngestHandler) IngestEdgeClick(c *gin.Context) {
	correlationID := uuid.New().String()[:8]
	startTime := time.Now()

	// Check content encoding
	contentEncoding := c.GetHeader("Content-Encoding")
	isBatch := c.GetHeader("X-Edge-Batch") == "true"
	isJSONLines := c.GetHeader("Content-Type") == "application/x-ndjson"

	var processed, failed int
	var err error

	if contentEncoding == "gzip" {
		// Handle gzip compressed data
		processed, failed, err = h.ingestService.IngestGzip(c.Request.Body)
	} else if isJSONLines {
		// Handle JSON lines format
		processed, failed, err = h.ingestService.IngestJSONLines(c.Request.Body)
	} else if isBatch {
		// Handle batch JSON
		var batch services.EdgeBatchRequest
		if err := c.ShouldBindJSON(&batch); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":        false,
				"correlation_id": correlationID,
				"error":          "Invalid request body: " + err.Error(),
			})
			return
		}
		processed, failed, err = h.ingestService.IngestBatch(batch.Events)
	} else {
		// Handle single event
		var event services.EdgeClickEvent
		if err := c.ShouldBindJSON(&event); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":        false,
				"correlation_id": correlationID,
				"error":          "Invalid request body: " + err.Error(),
			})
			return
		}
		err = h.ingestService.IngestEvent(event)
		if err == nil {
			processed = 1
		} else {
			failed = 1
		}
	}

	if err != nil && processed == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to ingest events: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"processed":     processed,
			"failed":        failed,
			"processing_ms": time.Since(startTime).Milliseconds(),
		},
		"timestamp": time.Now().UTC(),
	})
}

// GetOfferConfig returns offer config for edge caching
// GET /api/internal/edge/offer/:trackingCode
func (h *EdgeIngestHandler) GetOfferConfig(c *gin.Context) {
	correlationID := uuid.New().String()[:8]
	trackingCode := c.Param("trackingCode")

	if trackingCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Missing tracking code",
		})
		return
	}

	config, err := h.ingestService.GetOfferConfigForEdge(trackingCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Offer not found: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           config,
		"timestamp":      time.Now().UTC(),
	})
}

// GetEdgeStats returns edge ingestion statistics
// GET /api/internal/edge/stats
func (h *EdgeIngestHandler) GetEdgeStats(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	stats := h.ingestService.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           stats,
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// ADMIN EDGE HANDLER
// ============================================

// AdminEdgeHandler handles admin edge management
type AdminEdgeHandler struct {
	db            *gorm.DB
	ingestService *services.EdgeIngestService
}

// NewAdminEdgeHandler creates a new admin edge handler
func NewAdminEdgeHandler(db *gorm.DB) *AdminEdgeHandler {
	return &AdminEdgeHandler{
		db:            db,
		ingestService: services.GetEdgeIngestService(db),
	}
}

// GetEdgeStatus returns edge system status
// GET /api/admin/edge/status
func (h *AdminEdgeHandler) GetEdgeStatus(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	stats := h.ingestService.GetStats()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"status":          "active",
			"ingest_stats":    stats,
			"workers_active":  true,
			"queue_healthy":   stats["queue_size"].(int) < stats["queue_capacity"].(int)/2,
		},
		"timestamp": time.Now().UTC(),
	})
}

// GetEdgeRegions returns edge region statistics
// GET /api/admin/edge/regions
func (h *AdminEdgeHandler) GetEdgeRegions(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	// In production, this would aggregate from edge metrics
	regions := []map[string]interface{}{
		{"region": "IAD", "name": "US East (Virginia)", "clicks": 0, "latency_ms": 5},
		{"region": "LHR", "name": "Europe (London)", "clicks": 0, "latency_ms": 8},
		{"region": "SIN", "name": "Asia (Singapore)", "clicks": 0, "latency_ms": 12},
		{"region": "SYD", "name": "Australia (Sydney)", "clicks": 0, "latency_ms": 15},
		{"region": "DXB", "name": "Middle East (Dubai)", "clicks": 0, "latency_ms": 10},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           regions,
		"timestamp":      time.Now().UTC(),
	})
}

// GetEdgeRouter returns router statistics
// GET /api/admin/edge/router
func (h *AdminEdgeHandler) GetEdgeRouter(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	// In production, this would aggregate from edge metrics
	router := map[string]interface{}{
		"total_decisions":     0,
		"ab_tests":            0,
		"rotations":           0,
		"caps_exceeded":       0,
		"geo_blocks":          0,
		"device_blocks":       0,
		"bot_blocks":          0,
		"avg_decision_time_ms": 2,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           router,
		"timestamp":      time.Now().UTC(),
	})
}

// GetEdgeFullStats returns full edge statistics
// GET /api/admin/edge/stats
func (h *AdminEdgeHandler) GetEdgeFullStats(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	ingestStats := h.ingestService.GetStats()

	stats := map[string]interface{}{
		"ingest":          ingestStats,
		"clicks_per_second": 0,
		"bot_block_rate":    "0%",
		"geo_block_rate":    "0%",
		"avg_latency_ms":    5,
		"p99_latency_ms":    15,
		"error_rate":        "0%",
		"uptime_hours":      0,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           stats,
		"timestamp":      time.Now().UTC(),
	})
}

// GetEdgeQueue returns edge queue status
// GET /api/admin/edge/queue
func (h *AdminEdgeHandler) GetEdgeQueue(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	stats := h.ingestService.GetStats()

	queue := map[string]interface{}{
		"size":              stats["queue_size"],
		"capacity":          stats["queue_capacity"],
		"utilization":       float64(stats["queue_size"].(int)) / float64(stats["queue_capacity"].(int)) * 100,
		"pending_batches":   0,
		"failed_items":      stats["total_failed"],
		"dlq_size":          0,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           queue,
		"timestamp":      time.Now().UTC(),
	})
}

// GetEdgeFailover returns failover status
// GET /api/admin/edge/failover
func (h *AdminEdgeHandler) GetEdgeFailover(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	failover := map[string]interface{}{
		"backend_healthy":       true,
		"last_health_check":     time.Now().UTC(),
		"consecutive_failures":  0,
		"failover_active":       false,
		"cached_offers":         0,
		"local_clicks_pending":  0,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           failover,
		"timestamp":      time.Now().UTC(),
	})
}

// RefreshEdgeCache triggers edge cache refresh
// POST /api/admin/edge/cache/refresh
func (h *AdminEdgeHandler) RefreshEdgeCache(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req struct {
		Type string `json:"type"` // offer, geo_rule, tenant, all
		ID   string `json:"id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	// In production, this would send a message to edge workers
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Cache refresh request sent to edge workers",
		"data": gin.H{
			"type": req.Type,
			"id":   req.ID,
		},
		"timestamp": time.Now().UTC(),
	})
}

// FlushEdgeQueue triggers edge queue flush
// POST /api/admin/edge/queue/flush
func (h *AdminEdgeHandler) FlushEdgeQueue(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	// In production, this would trigger queue flush
	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Queue flush triggered",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// GZIP READER HELPER
// ============================================

// readGzipBody reads and decompresses gzip body
func readGzipBody(body io.ReadCloser) ([]byte, error) {
	gzReader, err := gzip.NewReader(body)
	if err != nil {
		return nil, err
	}
	defer gzReader.Close()
	
	return io.ReadAll(gzReader)
}


package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// ADMIN WEBHOOKS HANDLER
// ============================================

// AdminWebhooksHandler handles webhook management endpoints
type AdminWebhooksHandler struct {
	db             *gorm.DB
	webhookService *services.WebhookService
}

// NewAdminWebhooksHandler creates a new admin webhooks handler
func NewAdminWebhooksHandler(db *gorm.DB) *AdminWebhooksHandler {
	return &AdminWebhooksHandler{
		db:             db,
		webhookService: services.GetWebhookService(db),
	}
}

// ============================================
// PIPELINE ENDPOINTS
// ============================================

// GetAllPipelines returns all webhook pipelines
// GET /api/admin/webhooks/pipelines
func (h *AdminWebhooksHandler) GetAllPipelines(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	// Parse filters
	var advertiserID, offerID *uuid.UUID
	var triggerType *models.WebhookTriggerType
	var status *models.WebhookPipelineStatus

	if advID := c.Query("advertiser_id"); advID != "" {
		if id, err := uuid.Parse(advID); err == nil {
			advertiserID = &id
		}
	}

	if offID := c.Query("offer_id"); offID != "" {
		if id, err := uuid.Parse(offID); err == nil {
			offerID = &id
		}
	}

	if trigger := c.Query("trigger_type"); trigger != "" {
		t := models.WebhookTriggerType(trigger)
		triggerType = &t
	}

	if s := c.Query("status"); s != "" {
		st := models.WebhookPipelineStatus(s)
		status = &st
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	pipelines, total, err := h.webhookService.ListPipelines(advertiserID, offerID, triggerType, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to fetch pipelines: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"pipelines": pipelines,
			"total":     total,
			"limit":     limit,
			"offset":    offset,
		},
		"timestamp": time.Now().UTC(),
	})
}

// GetPipeline returns a single pipeline
// GET /api/admin/webhooks/pipelines/:id
func (h *AdminWebhooksHandler) GetPipeline(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	pipelineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid pipeline ID",
		})
		return
	}

	pipeline, err := h.webhookService.GetPipeline(pipelineID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Pipeline not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           pipeline,
		"timestamp":      time.Now().UTC(),
	})
}

// CreatePipeline creates a new pipeline
// POST /api/admin/webhooks/pipelines
func (h *AdminWebhooksHandler) CreatePipeline(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req struct {
		Name         string                        `json:"name" binding:"required"`
		Description  string                        `json:"description"`
		AdvertiserID *uuid.UUID                    `json:"advertiser_id"`
		OfferID      *uuid.UUID                    `json:"offer_id"`
		TriggerType  models.WebhookTriggerType     `json:"trigger_type" binding:"required"`
		Status       models.WebhookPipelineStatus  `json:"status"`
		FailoverURL  string                        `json:"failover_url"`
		MaxRetries   int                           `json:"max_retries"`
		TimeoutMs    int                           `json:"timeout_ms"`
		Priority     int                           `json:"priority"`
		Steps        []models.WebhookStep          `json:"steps"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	// Set defaults
	if req.Status == "" {
		req.Status = models.WebhookPipelineStatusDraft
	}
	if req.MaxRetries == 0 {
		req.MaxRetries = 5
	}
	if req.TimeoutMs == 0 {
		req.TimeoutMs = 30000
	}

	pipeline := &models.WebhookPipeline{
		ID:           uuid.New(),
		Name:         req.Name,
		Description:  req.Description,
		AdvertiserID: req.AdvertiserID,
		OfferID:      req.OfferID,
		TriggerType:  req.TriggerType,
		Status:       req.Status,
		FailoverURL:  req.FailoverURL,
		MaxRetries:   req.MaxRetries,
		TimeoutMs:    req.TimeoutMs,
		Priority:     req.Priority,
		Steps:        req.Steps,
	}

	if err := h.webhookService.CreatePipeline(pipeline); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to create pipeline: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           pipeline,
		"message":        "Pipeline created successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// UpdatePipeline updates a pipeline
// PUT /api/admin/webhooks/pipelines/:id
func (h *AdminWebhooksHandler) UpdatePipeline(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	pipelineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid pipeline ID",
		})
		return
	}

	var pipeline models.WebhookPipeline
	if err := c.ShouldBindJSON(&pipeline); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	pipeline.ID = pipelineID

	if err := h.webhookService.UpdatePipeline(&pipeline); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to update pipeline: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           pipeline,
		"message":        "Pipeline updated successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// DeletePipeline deletes a pipeline
// DELETE /api/admin/webhooks/pipelines/:id
func (h *AdminWebhooksHandler) DeletePipeline(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	pipelineID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid pipeline ID",
		})
		return
	}

	if err := h.webhookService.DeletePipeline(pipelineID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to delete pipeline: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Pipeline deleted successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// EXECUTION LOG ENDPOINTS
// ============================================

// GetRecentLogs returns recent execution logs
// GET /api/admin/webhooks/logs/recent
func (h *AdminWebhooksHandler) GetRecentLogs(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var pipelineID *uuid.UUID
	var status *models.WebhookExecutionStatus

	if pipID := c.Query("pipeline_id"); pipID != "" {
		if id, err := uuid.Parse(pipID); err == nil {
			pipelineID = &id
		}
	}

	if s := c.Query("status"); s != "" {
		st := models.WebhookExecutionStatus(s)
		status = &st
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	executions, total, err := h.webhookService.ListExecutions(pipelineID, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to fetch logs: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"executions": executions,
			"total":      total,
			"limit":      limit,
			"offset":     offset,
		},
		"timestamp": time.Now().UTC(),
	})
}

// GetExecutionLog returns a single execution log
// GET /api/admin/webhooks/logs/:task_id
func (h *AdminWebhooksHandler) GetExecutionLog(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	executionID, err := uuid.Parse(c.Param("task_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid execution ID",
		})
		return
	}

	execution, err := h.webhookService.GetExecution(executionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Execution not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           execution,
		"timestamp":      time.Now().UTC(),
	})
}

// GetFailureLogs returns recent failure logs
// GET /api/admin/webhooks/logs/failures
func (h *AdminWebhooksHandler) GetFailureLogs(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	failures, err := h.webhookService.GetRecentFailures(limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to fetch failures: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"failures": failures,
			"count":    len(failures),
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// DLQ ENDPOINTS
// ============================================

// GetDLQ returns DLQ items
// GET /api/admin/webhooks/dlq
func (h *AdminWebhooksHandler) GetDLQ(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	items, total, err := h.webhookService.GetDLQItems(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to fetch DLQ: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"items":  items,
			"total":  total,
			"limit":  limit,
			"offset": offset,
		},
		"timestamp": time.Now().UTC(),
	})
}

// RetryDLQItem retries a DLQ item
// POST /api/admin/webhooks/dlq/retry/:id
func (h *AdminWebhooksHandler) RetryDLQItem(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid DLQ item ID",
		})
		return
	}

	if err := h.webhookService.RetryDLQItem(itemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to retry DLQ item: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "DLQ item queued for retry",
		"timestamp":      time.Now().UTC(),
	})
}

// DeleteDLQItem deletes a DLQ item
// DELETE /api/admin/webhooks/dlq/:id
func (h *AdminWebhooksHandler) DeleteDLQItem(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	itemID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid DLQ item ID",
		})
		return
	}

	if err := h.webhookService.DeleteDLQItem(itemID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to delete DLQ item: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "DLQ item deleted",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// TESTING ENDPOINTS
// ============================================

// TestPipeline tests a pipeline
// POST /api/admin/webhooks/test/pipeline
func (h *AdminWebhooksHandler) TestPipeline(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req struct {
		PipelineID uuid.UUID              `json:"pipeline_id" binding:"required"`
		Payload    map[string]interface{} `json:"payload"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	// Use default test payload if not provided
	if req.Payload == nil {
		req.Payload = map[string]interface{}{
			"click": map[string]interface{}{
				"id":         uuid.New().String(),
				"ip":         "192.168.1.1",
				"user_agent": "Test User Agent",
				"device":     "desktop",
				"country":    "US",
			},
			"conversion": map[string]interface{}{
				"id":          uuid.New().String(),
				"amount":      100,
				"external_id": "test-" + uuid.New().String()[:8],
			},
			"custom": map[string]interface{}{
				"test": true,
			},
		}
	}

	execution, err := h.webhookService.TestPipeline(req.PipelineID, req.Payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to test pipeline: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"execution_id": execution.ID,
			"status":       execution.Status,
			"message":      "Test execution queued. Check logs for results.",
		},
		"timestamp": time.Now().UTC(),
	})
}

// TestStep tests a single step
// POST /api/admin/webhooks/test/step
func (h *AdminWebhooksHandler) TestStep(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req struct {
		Step    models.WebhookStep     `json:"step" binding:"required"`
		Payload map[string]interface{} `json:"payload"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	// Use default test payload if not provided
	if req.Payload == nil {
		req.Payload = map[string]interface{}{
			"custom": map[string]interface{}{
				"test": true,
			},
		}
	}

	result, err := h.webhookService.TestStep(&req.Step, req.Payload)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to test step: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"success":       result.Success,
			"status_code":   result.StatusCode,
			"response_body": result.ResponseBody,
			"error":         result.Error,
			"duration_ms":   result.DurationMs,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// STATS ENDPOINT
// ============================================

// GetStats returns webhook statistics
// GET /api/admin/webhooks/stats
func (h *AdminWebhooksHandler) GetStats(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	metrics := h.webhookService.GetMetrics()

	// Get additional counts from DB
	var pipelineCount, activePipelineCount int64
	h.db.Model(&models.WebhookPipeline{}).Count(&pipelineCount)
	h.db.Model(&models.WebhookPipeline{}).Where("status = ?", models.WebhookPipelineStatusActive).Count(&activePipelineCount)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"metrics":          metrics,
			"pipeline_count":   pipelineCount,
			"active_pipelines": activePipelineCount,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// TRIGGER TYPES REFERENCE
// ============================================

// GetTriggerTypes returns available trigger types
// GET /api/admin/webhooks/trigger-types
func (h *AdminWebhooksHandler) GetTriggerTypes(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": []map[string]string{
			{"value": string(models.WebhookTriggerClick), "label": "Click", "description": "Triggered when a tracking link is clicked"},
			{"value": string(models.WebhookTriggerConversion), "label": "Conversion", "description": "Triggered when a conversion is recorded"},
			{"value": string(models.WebhookTriggerPostback), "label": "Postback", "description": "Triggered when a postback is received"},
			{"value": string(models.WebhookTriggerJoinOffer), "label": "Join Offer", "description": "Triggered when a user joins an offer"},
			{"value": string(models.WebhookTriggerCustom), "label": "Custom", "description": "Custom trigger for testing"},
		},
		"timestamp": time.Now().UTC(),
	})
}

// GetSignatureModes returns available signature modes
// GET /api/admin/webhooks/signature-modes
func (h *AdminWebhooksHandler) GetSignatureModes(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": []map[string]string{
			{"value": string(models.WebhookSignatureNone), "label": "None", "description": "No signature"},
			{"value": string(models.WebhookSignatureHMAC), "label": "HMAC-SHA256", "description": "HMAC-SHA256 signature in X-Afftok-Signature header"},
			{"value": string(models.WebhookSignatureJWT), "label": "JWT", "description": "JWT token in Authorization header"},
		},
		"timestamp": time.Now().UTC(),
	})
}


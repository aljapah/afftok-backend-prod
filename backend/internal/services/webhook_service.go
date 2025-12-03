package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// WEBHOOK SERVICE
// ============================================

// WebhookService is the main service for webhook management
type WebhookService struct {
	db             *gorm.DB
	queueService   *WebhookQueueService
	workerPool     *WebhookWorkerPool
	templateEngine *TemplateEngine
	signingService *WebhookSigningService
	observability  *ObservabilityService
	mutex          sync.RWMutex
}

// NewWebhookService creates a new webhook service
func NewWebhookService(db *gorm.DB) *WebhookService {
	service := &WebhookService{
		db:             db,
		queueService:   GetWebhookQueueService(),
		templateEngine: NewTemplateEngine(),
		signingService: NewWebhookSigningService(),
		observability:  NewObservabilityService(),
	}

	// Initialize worker pool
	service.workerPool = NewWebhookWorkerPool(db)

	return service
}

// Start starts the webhook service
func (s *WebhookService) Start() {
	s.workerPool.Start()
}

// Stop stops the webhook service
func (s *WebhookService) Stop() {
	s.workerPool.Stop()
	s.queueService.Stop()
}

// ============================================
// PIPELINE MANAGEMENT
// ============================================

// CreatePipeline creates a new webhook pipeline
func (s *WebhookService) CreatePipeline(pipeline *models.WebhookPipeline) error {
	if pipeline.ID == uuid.Nil {
		pipeline.ID = uuid.New()
	}

	// Validate steps
	for i := range pipeline.Steps {
		if pipeline.Steps[i].ID == uuid.Nil {
			pipeline.Steps[i].ID = uuid.New()
		}
		pipeline.Steps[i].PipelineID = pipeline.ID
		pipeline.Steps[i].StepOrder = i
	}

	return s.db.Create(pipeline).Error
}

// UpdatePipeline updates an existing pipeline
func (s *WebhookService) UpdatePipeline(pipeline *models.WebhookPipeline) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Update pipeline
		if err := tx.Save(pipeline).Error; err != nil {
			return err
		}

		// Delete old steps
		if err := tx.Where("pipeline_id = ?", pipeline.ID).Delete(&models.WebhookStep{}).Error; err != nil {
			return err
		}

		// Create new steps
		for i := range pipeline.Steps {
			if pipeline.Steps[i].ID == uuid.Nil {
				pipeline.Steps[i].ID = uuid.New()
			}
			pipeline.Steps[i].PipelineID = pipeline.ID
			pipeline.Steps[i].StepOrder = i
			if err := tx.Create(&pipeline.Steps[i]).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

// DeletePipeline deletes a pipeline
func (s *WebhookService) DeletePipeline(pipelineID uuid.UUID) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Delete steps first (cascade should handle this, but be explicit)
		if err := tx.Where("pipeline_id = ?", pipelineID).Delete(&models.WebhookStep{}).Error; err != nil {
			return err
		}

		// Delete pipeline
		return tx.Delete(&models.WebhookPipeline{}, "id = ?", pipelineID).Error
	})
}

// GetPipeline gets a pipeline by ID
func (s *WebhookService) GetPipeline(pipelineID uuid.UUID) (*models.WebhookPipeline, error) {
	var pipeline models.WebhookPipeline
	err := s.db.Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_order ASC")
	}).First(&pipeline, "id = ?", pipelineID).Error

	if err != nil {
		return nil, err
	}

	return &pipeline, nil
}

// ListPipelines lists all pipelines with optional filters
func (s *WebhookService) ListPipelines(
	advertiserID *uuid.UUID,
	offerID *uuid.UUID,
	triggerType *models.WebhookTriggerType,
	status *models.WebhookPipelineStatus,
	limit, offset int,
) ([]models.WebhookPipeline, int64, error) {
	query := s.db.Model(&models.WebhookPipeline{})

	if advertiserID != nil {
		query = query.Where("advertiser_id = ?", *advertiserID)
	}
	if offerID != nil {
		query = query.Where("offer_id = ?", *offerID)
	}
	if triggerType != nil {
		query = query.Where("trigger_type = ?", *triggerType)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	var total int64
	query.Count(&total)

	var pipelines []models.WebhookPipeline
	err := query.Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_order ASC")
	}).Order("priority DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&pipelines).Error

	return pipelines, total, err
}

// ============================================
// WEBHOOK TRIGGERING
// ============================================

// TriggerWebhook triggers webhooks for a specific event
func (s *WebhookService) TriggerWebhook(
	triggerType models.WebhookTriggerType,
	triggerID string,
	advertiserID *uuid.UUID,
	offerID *uuid.UUID,
	payload map[string]interface{},
) error {
	// Find matching pipelines
	pipelines, err := s.findMatchingPipelines(triggerType, advertiserID, offerID)
	if err != nil {
		return err
	}

	if len(pipelines) == 0 {
		return nil // No webhooks configured
	}

	correlationID := uuid.New().String()[:8]

	// Create tasks for each pipeline
	for _, pipeline := range pipelines {
		task := CreateWebhookTask(
			uuid.New(),        // execution ID
			pipeline.ID,
			pipeline.AdvertiserID,
			pipeline.OfferID,
			0,                 // start from step 0
			payload,
			pipeline.MaxRetries,
			pipeline.Priority,
		)
		task.CorrelationID = correlationID

		// Enqueue task
		if err := s.queueService.EnqueuePrimary(task); err != nil {
			s.observability.Log(LogEvent{
				Category:      "webhook_trigger_error",
				Level:         LogLevelError,
				Message:       "Failed to enqueue webhook task",
				CorrelationID: correlationID,
				Metadata: map[string]interface{}{
					"pipeline_id": pipeline.ID.String(),
					"trigger_id":  triggerID,
					"error":       err.Error(),
				},
			})
			continue
		}

		s.observability.Log(LogEvent{
			Category:      "webhook_triggered",
			Level:         LogLevelInfo,
			Message:       "Webhook triggered",
			CorrelationID: correlationID,
			Metadata: map[string]interface{}{
				"pipeline_id":  pipeline.ID.String(),
				"pipeline_name": pipeline.Name,
				"trigger_type": string(triggerType),
				"trigger_id":   triggerID,
			},
		})
	}

	return nil
}

// findMatchingPipelines finds pipelines that match the trigger criteria
func (s *WebhookService) findMatchingPipelines(
	triggerType models.WebhookTriggerType,
	advertiserID *uuid.UUID,
	offerID *uuid.UUID,
) ([]models.WebhookPipeline, error) {
	var pipelines []models.WebhookPipeline

	query := s.db.Where("trigger_type = ? AND status = ?", triggerType, models.WebhookPipelineStatusActive)

	// Match by specificity: offer > advertiser > global
	if offerID != nil {
		// Try offer-specific first
		var offerPipelines []models.WebhookPipeline
		s.db.Where("trigger_type = ? AND status = ? AND offer_id = ?",
			triggerType, models.WebhookPipelineStatusActive, *offerID).
			Order("priority DESC").
			Find(&offerPipelines)
		
		if len(offerPipelines) > 0 {
			return offerPipelines, nil
		}
	}

	if advertiserID != nil {
		// Try advertiser-specific
		var advertiserPipelines []models.WebhookPipeline
		s.db.Where("trigger_type = ? AND status = ? AND advertiser_id = ? AND offer_id IS NULL",
			triggerType, models.WebhookPipelineStatusActive, *advertiserID).
			Order("priority DESC").
			Find(&advertiserPipelines)
		
		if len(advertiserPipelines) > 0 {
			return advertiserPipelines, nil
		}
	}

	// Fall back to global pipelines
	query.Where("advertiser_id IS NULL AND offer_id IS NULL").
		Order("priority DESC").
		Find(&pipelines)

	return pipelines, nil
}

// ============================================
// TRIGGER HELPERS
// ============================================

// TriggerClickWebhook triggers webhooks for a click event
func (s *WebhookService) TriggerClickWebhook(click *models.Click, userOffer *models.UserOffer, offer *models.Offer) error {
	payload := map[string]interface{}{
		"click": map[string]interface{}{
			"id":         click.ID.String(),
			"ip":         click.IPAddress,
			"user_agent": click.UserAgent,
			"device":     click.Device,
			"browser":    click.Browser,
			"os":         click.OS,
			"country":    click.Country,
			"city":       click.City,
			"clicked_at": click.ClickedAt,
		},
		"user_offer": map[string]interface{}{
			"id":      userOffer.ID.String(),
			"user_id": userOffer.UserID.String(),
		},
		"offer": map[string]interface{}{
			"id":    offer.ID.String(),
			"title": offer.Title,
		},
	}

	// Use NetworkID as advertiser reference
	var advertiserID *uuid.UUID
	if offer.NetworkID != nil {
		advertiserID = offer.NetworkID
	}

	return s.TriggerWebhook(
		models.WebhookTriggerClick,
		click.ID.String(),
		advertiserID,
		&offer.ID,
		payload,
	)
}

// TriggerConversionWebhook triggers webhooks for a conversion event
func (s *WebhookService) TriggerConversionWebhook(conversion *models.Conversion, userOffer *models.UserOffer, offer *models.Offer) error {
	payload := map[string]interface{}{
		"conversion": map[string]interface{}{
			"id":                     conversion.ID.String(),
			"external_conversion_id": conversion.ExternalConversionID,
			"amount":                 conversion.Amount,
			"status":                 conversion.Status,
			"network_id":             conversion.NetworkID,
			"converted_at":           conversion.ConvertedAt,
		},
		"user_offer": map[string]interface{}{
			"id":      userOffer.ID.String(),
			"user_id": userOffer.UserID.String(),
		},
		"offer": map[string]interface{}{
			"id":    offer.ID.String(),
			"title": offer.Title,
		},
	}

	// Use NetworkID as advertiser reference
	var advertiserID *uuid.UUID
	if offer.NetworkID != nil {
		advertiserID = offer.NetworkID
	}

	return s.TriggerWebhook(
		models.WebhookTriggerConversion,
		conversion.ID.String(),
		advertiserID,
		&offer.ID,
		payload,
	)
}

// TriggerPostbackWebhook triggers webhooks for a postback event
func (s *WebhookService) TriggerPostbackWebhook(postbackData map[string]interface{}, advertiserID *uuid.UUID, offerID *uuid.UUID) error {
	payload := map[string]interface{}{
		"postback": postbackData,
	}

	triggerID := ""
	if id, ok := postbackData["id"].(string); ok {
		triggerID = id
	}

	return s.TriggerWebhook(
		models.WebhookTriggerPostback,
		triggerID,
		advertiserID,
		offerID,
		payload,
	)
}

// TriggerJoinOfferWebhook triggers webhooks when a user joins an offer
func (s *WebhookService) TriggerJoinOfferWebhook(userOffer *models.UserOffer, offer *models.Offer, user map[string]interface{}) error {
	payload := map[string]interface{}{
		"user_offer": map[string]interface{}{
			"id":             userOffer.ID.String(),
			"user_id":        userOffer.UserID.String(),
			"offer_id":       userOffer.OfferID.String(),
			"affiliate_link": userOffer.AffiliateLink,
			"joined_at":      userOffer.JoinedAt,
		},
		"offer": map[string]interface{}{
			"id":    offer.ID.String(),
			"title": offer.Title,
		},
		"user": user,
	}

	// Use NetworkID as advertiser reference
	var advertiserID *uuid.UUID
	if offer.NetworkID != nil {
		advertiserID = offer.NetworkID
	}

	return s.TriggerWebhook(
		models.WebhookTriggerJoinOffer,
		userOffer.ID.String(),
		advertiserID,
		&offer.ID,
		payload,
	)
}

// ============================================
// EXECUTION LOGS
// ============================================

// GetExecution gets an execution by ID
func (s *WebhookService) GetExecution(executionID uuid.UUID) (*models.WebhookExecution, error) {
	var execution models.WebhookExecution
	err := s.db.Preload("Pipeline").Preload("StepResults").
		First(&execution, "id = ?", executionID).Error
	
	if err != nil {
		return nil, err
	}
	
	return &execution, nil
}

// ListExecutions lists executions with filters
func (s *WebhookService) ListExecutions(
	pipelineID *uuid.UUID,
	status *models.WebhookExecutionStatus,
	limit, offset int,
) ([]models.WebhookExecution, int64, error) {
	query := s.db.Model(&models.WebhookExecution{})

	if pipelineID != nil {
		query = query.Where("pipeline_id = ?", *pipelineID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	var total int64
	query.Count(&total)

	var executions []models.WebhookExecution
	err := query.Preload("Pipeline").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&executions).Error

	return executions, total, err
}

// GetRecentFailures gets recent failed executions
func (s *WebhookService) GetRecentFailures(limit int) ([]models.WebhookExecution, error) {
	var executions []models.WebhookExecution
	err := s.db.Where("status IN ?", []models.WebhookExecutionStatus{
		models.WebhookExecutionFailed,
		models.WebhookExecutionDLQ,
	}).Preload("Pipeline").
		Order("created_at DESC").
		Limit(limit).
		Find(&executions).Error

	return executions, err
}

// ============================================
// DLQ MANAGEMENT
// ============================================

// GetDLQItems gets items from the dead letter queue
func (s *WebhookService) GetDLQItems(limit, offset int) ([]models.WebhookDLQItem, int64, error) {
	var items []models.WebhookDLQItem
	var total int64

	s.db.Model(&models.WebhookDLQItem{}).Count(&total)

	err := s.db.Preload("Pipeline").Preload("Execution").
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error

	return items, total, err
}

// RetryDLQItem retries a DLQ item
func (s *WebhookService) RetryDLQItem(dlqItemID uuid.UUID) error {
	var dlqItem models.WebhookDLQItem
	if err := s.db.First(&dlqItem, "id = ?", dlqItemID).Error; err != nil {
		return err
	}

	if !dlqItem.CanRetry {
		return fmt.Errorf("DLQ item cannot be retried")
	}

	// Parse task data
	var task models.WebhookTask
	if err := json.Unmarshal(dlqItem.TaskData, &task); err != nil {
		return fmt.Errorf("failed to parse task data: %w", err)
	}

	// Reset task
	task.Attempts = 0
	task.LastError = ""
	task.NextRetryAt = time.Time{}

	// Re-enqueue
	if err := s.queueService.EnqueuePrimary(&task); err != nil {
		return err
	}

	// Update DLQ item
	now := time.Now()
	return s.db.Model(&dlqItem).Updates(map[string]interface{}{
		"retried_at": &now,
		"can_retry":  false,
	}).Error
}

// DeleteDLQItem deletes a DLQ item
func (s *WebhookService) DeleteDLQItem(dlqItemID uuid.UUID) error {
	return s.db.Delete(&models.WebhookDLQItem{}, "id = ?", dlqItemID).Error
}

// ============================================
// TESTING
// ============================================

// TestPipeline tests a pipeline with sample data
func (s *WebhookService) TestPipeline(pipelineID uuid.UUID, testPayload map[string]interface{}) (*models.WebhookExecution, error) {
	pipeline, err := s.GetPipeline(pipelineID)
	if err != nil {
		return nil, err
	}

	// Create test execution
	executionID := uuid.New()
	now := time.Now()
	
	execution := &models.WebhookExecution{
		ID:            executionID,
		PipelineID:    pipelineID,
		TriggerType:   models.WebhookTriggerCustom,
		TriggerID:     "test-" + uuid.New().String()[:8],
		CorrelationID: uuid.New().String()[:8],
		Status:        models.WebhookExecutionPending,
		TotalSteps:    len(pipeline.Steps),
		MaxAttempts:   1,
		StartedAt:     &now,
	}
	
	payloadJSON, _ := json.Marshal(testPayload)
	execution.Payload = payloadJSON

	if err := s.db.Create(execution).Error; err != nil {
		return nil, err
	}

	// Create and enqueue task
	task := CreateWebhookTask(
		executionID,
		pipelineID,
		pipeline.AdvertiserID,
		pipeline.OfferID,
		0,
		testPayload,
		1, // Only 1 attempt for tests
		10, // High priority
	)
	task.CorrelationID = execution.CorrelationID

	if err := s.queueService.EnqueuePrimary(task); err != nil {
		return nil, err
	}

	return execution, nil
}

// TestStep tests a single step
func (s *WebhookService) TestStep(step *models.WebhookStep, testPayload map[string]interface{}) (*StepExecutionResult, error) {
	ctx := NewTemplateContext()
	ctx.CorrelationID = uuid.New().String()[:8]
	ctx.TaskID = "test-" + uuid.New().String()[:8]

	// Populate context from payload
	if click, ok := testPayload["click"].(map[string]interface{}); ok {
		ctx.Click = click
	}
	if conversion, ok := testPayload["conversion"].(map[string]interface{}); ok {
		ctx.Conversion = conversion
	}
	if custom, ok := testPayload["custom"].(map[string]interface{}); ok {
		ctx.Custom = custom
	}

	task := &models.WebhookTask{
		ID:            ctx.TaskID,
		ExecutionID:   uuid.New(),
		PipelineID:    step.PipelineID,
		Payload:       testPayload,
		CorrelationID: ctx.CorrelationID,
	}

	return s.workerPool.executeStep(step, ctx, task, 0), nil
}

// ============================================
// METRICS
// ============================================

// GetMetrics returns comprehensive webhook metrics
func (s *WebhookService) GetMetrics() *models.WebhookMetrics {
	queueMetrics := s.queueService.GetMetrics()
	workerMetrics := s.workerPool.GetMetrics()
	queueSizes := s.queueService.GetQueueSizes()

	// Calculate average latency
	var avgLatency float64
	if workerMetrics.TasksSucceeded > 0 {
		avgLatency = float64(workerMetrics.TotalLatencyMs) / float64(workerMetrics.TasksSucceeded)
	}

	// Count DLQ items in DB
	var dlqCount int64
	s.db.Model(&models.WebhookDLQItem{}).Count(&dlqCount)

	// Count pending/running tasks
	var pendingCount, runningCount int64
	s.db.Model(&models.WebhookExecution{}).Where("status = ?", models.WebhookExecutionPending).Count(&pendingCount)
	s.db.Model(&models.WebhookExecution{}).Where("status = ?", models.WebhookExecutionRunning).Count(&runningCount)

	return &models.WebhookMetrics{
		TotalTasks:        queueMetrics.TotalEnqueued,
		TotalSuccess:      workerMetrics.TasksSucceeded,
		TotalFailures:     workerMetrics.TasksFailed,
		TotalRetries:      workerMetrics.TasksRetried,
		FailoverTriggered: queueMetrics.TotalFailover,
		DLQItems:          dlqCount,
		AvgLatencyMs:      avgLatency,
		StepSuccessCount:  workerMetrics.StepsSucceeded,
		StepFailureCount:  workerMetrics.StepsFailed,
		PendingTasks:      pendingCount,
		RunningTasks:      runningCount,
		QueueSizes:        queueSizes,
	}
}

// ============================================
// GLOBAL INSTANCE
// ============================================

var (
	webhookServiceInstance *WebhookService
	webhookServiceOnce     sync.Once
)

// GetWebhookService returns the global webhook service instance
func GetWebhookService(db *gorm.DB) *WebhookService {
	webhookServiceOnce.Do(func() {
		webhookServiceInstance = NewWebhookService(db)
	})
	return webhookServiceInstance
}


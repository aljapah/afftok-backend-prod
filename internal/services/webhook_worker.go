package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// WEBHOOK WORKER POOL SERVICE
// ============================================

// WebhookWorkerPool manages webhook execution workers
type WebhookWorkerPool struct {
	db              *gorm.DB
	queueService    *WebhookQueueService
	signingService  *WebhookSigningService
	templateEngine  *TemplateEngine
	observability   *ObservabilityService
	httpClient      *http.Client

	// Worker counts
	primaryWorkers  int
	failoverWorkers int
	dlqWorkers      int

	// Control
	ctx        context.Context
	cancel     context.CancelFunc
	wg         sync.WaitGroup
	running    bool
	mutex      sync.RWMutex

	// Metrics
	metrics *webhookWorkerMetrics
}

// webhookWorkerMetrics tracks worker metrics
type webhookWorkerMetrics struct {
	TasksProcessed   int64
	TasksSucceeded   int64
	TasksFailed      int64
	TasksRetried     int64
	StepsExecuted    int64
	StepsSucceeded   int64
	StepsFailed      int64
	TotalLatencyMs   int64
	ActiveWorkers    int64
}

// NewWebhookWorkerPool creates a new webhook worker pool
func NewWebhookWorkerPool(db *gorm.DB) *WebhookWorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	cpuCount := runtime.NumCPU()

	return &WebhookWorkerPool{
		db:              db,
		queueService:    GetWebhookQueueService(),
		signingService:  NewWebhookSigningService(),
		templateEngine:  NewTemplateEngine(),
		observability:   NewObservabilityService(),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
		primaryWorkers:  cpuCount * 4,
		failoverWorkers: cpuCount * 2,
		dlqWorkers:      cpuCount,
		ctx:             ctx,
		cancel:          cancel,
		running:         false,
		metrics:         &webhookWorkerMetrics{},
	}
}

// ============================================
// WORKER LIFECYCLE
// ============================================

// Start starts all worker pools
func (p *WebhookWorkerPool) Start() {
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.running {
		return
	}

	p.running = true

	// Start primary workers
	for i := 0; i < p.primaryWorkers; i++ {
		p.wg.Add(1)
		go p.primaryWorker(i)
	}

	// Start failover workers
	for i := 0; i < p.failoverWorkers; i++ {
		p.wg.Add(1)
		go p.failoverWorker(i)
	}

	// Start DLQ workers
	for i := 0; i < p.dlqWorkers; i++ {
		p.wg.Add(1)
		go p.dlqWorker(i)
	}

	fmt.Printf("[WebhookWorker] Started %d primary, %d failover, %d DLQ workers\n",
		p.primaryWorkers, p.failoverWorkers, p.dlqWorkers)
}

// Stop stops all worker pools
func (p *WebhookWorkerPool) Stop() {
	p.mutex.Lock()
	p.running = false
	p.mutex.Unlock()

	p.cancel()
	p.wg.Wait()

	fmt.Println("[WebhookWorker] All workers stopped")
}

// ============================================
// WORKER IMPLEMENTATIONS
// ============================================

// primaryWorker processes tasks from the primary queue
func (p *WebhookWorkerPool) primaryWorker(id int) {
	defer p.wg.Done()
	atomic.AddInt64(&p.metrics.ActiveWorkers, 1)
	defer atomic.AddInt64(&p.metrics.ActiveWorkers, -1)

	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			task, err := p.queueService.DequeuePrimary(100 * time.Millisecond)
			if err != nil || task == nil {
				continue
			}

			// Check if task should wait for retry
			if task.NextRetryAt.After(time.Now()) {
				// Re-enqueue for later
				p.queueService.EnqueuePrimary(task)
				time.Sleep(100 * time.Millisecond)
				continue
			}

			p.processTask(task)
		}
	}
}

// failoverWorker processes tasks from the failover queue
func (p *WebhookWorkerPool) failoverWorker(id int) {
	defer p.wg.Done()
	atomic.AddInt64(&p.metrics.ActiveWorkers, 1)
	defer atomic.AddInt64(&p.metrics.ActiveWorkers, -1)

	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			task, err := p.queueService.DequeueFailover(200 * time.Millisecond)
			if err != nil || task == nil {
				continue
			}

			p.processFailoverTask(task)
		}
	}
}

// dlqWorker processes tasks from the DLQ (for manual retry)
func (p *WebhookWorkerPool) dlqWorker(id int) {
	defer p.wg.Done()
	atomic.AddInt64(&p.metrics.ActiveWorkers, 1)
	defer atomic.AddInt64(&p.metrics.ActiveWorkers, -1)

	for {
		select {
		case <-p.ctx.Done():
			return
		default:
			// DLQ items are processed less frequently
			time.Sleep(1 * time.Second)
			
			task, err := p.queueService.DequeueDLQ(500 * time.Millisecond)
			if err != nil || task == nil {
				continue
			}

			// Store in database for admin review
			p.storeDLQItem(task)
		}
	}
}

// ============================================
// TASK PROCESSING
// ============================================

// processTask processes a webhook task
func (p *WebhookWorkerPool) processTask(task *models.WebhookTask) {
	startTime := time.Now()
	atomic.AddInt64(&p.metrics.TasksProcessed, 1)

	// Log task start
	p.observability.Log(LogEvent{
		Category:      "webhook_task_start",
		Level:         LogLevelInfo,
		Message:       "Processing webhook task",
		CorrelationID: task.CorrelationID,
		Metadata: map[string]interface{}{
			"task_id":      task.ID,
			"pipeline_id":  task.PipelineID.String(),
			"step_index":   task.StepIndex,
			"attempt":      task.Attempts,
		},
	})

	// Load pipeline and steps
	var pipeline models.WebhookPipeline
	if err := p.db.Preload("Steps", func(db *gorm.DB) *gorm.DB {
		return db.Order("step_order ASC")
	}).First(&pipeline, "id = ?", task.PipelineID).Error; err != nil {
		p.handleTaskError(task, fmt.Errorf("pipeline not found: %w", err))
		return
	}

	// Get or create execution
	var execution models.WebhookExecution
	if err := p.db.First(&execution, "id = ?", task.ExecutionID).Error; err != nil {
		// Create new execution
		execution = models.WebhookExecution{
			ID:            task.ExecutionID,
			PipelineID:    task.PipelineID,
			TriggerType:   pipeline.TriggerType,
			CorrelationID: task.CorrelationID,
			Status:        models.WebhookExecutionRunning,
			TotalSteps:    len(pipeline.Steps),
			MaxAttempts:   task.MaxAttempts,
		}
		now := time.Now()
		execution.StartedAt = &now
		payloadJSON, _ := json.Marshal(task.Payload)
		execution.Payload = payloadJSON
		p.db.Create(&execution)
	} else {
		// Update status
		p.db.Model(&execution).Update("status", models.WebhookExecutionRunning)
	}

	// Build template context
	ctx := p.buildTemplateContext(task)

	// Execute steps starting from current step
	success := true
	for i := task.StepIndex; i < len(pipeline.Steps); i++ {
		step := pipeline.Steps[i]
		
		stepResult := p.executeStep(&step, ctx, task, i)
		
		// Store step result
		p.storeStepResult(&execution, &step, stepResult, i, task.Attempts)

		if !stepResult.Success {
			success = false
			task.LastError = stepResult.Error

			if step.StopOnFailure {
				break
			}
		}

		// Update current step
		p.db.Model(&execution).Update("current_step", i+1)
	}

	// Update execution status
	durationMs := time.Since(startTime).Milliseconds()
	atomic.AddInt64(&p.metrics.TotalLatencyMs, durationMs)

	if success {
		atomic.AddInt64(&p.metrics.TasksSucceeded, 1)
		now := time.Now()
		p.db.Model(&execution).Updates(map[string]interface{}{
			"status":       models.WebhookExecutionSuccess,
			"completed_at": &now,
			"duration_ms":  durationMs,
		})

		p.observability.Log(LogEvent{
			Category:      "webhook_task_success",
			Level:         LogLevelInfo,
			Message:       "Webhook task completed successfully",
			CorrelationID: task.CorrelationID,
			DurationMs:    durationMs,
		})
	} else {
		p.handleTaskError(task, fmt.Errorf(task.LastError))
	}
}

// processFailoverTask processes a task from the failover queue
func (p *WebhookWorkerPool) processFailoverTask(task *models.WebhookTask) {
	// Load pipeline to check for failover URL
	var pipeline models.WebhookPipeline
	if err := p.db.First(&pipeline, "id = ?", task.PipelineID).Error; err != nil {
		p.queueService.EnqueueDLQ(task)
		return
	}

	// Try failover URL if available
	if pipeline.FailoverURL != "" {
		ctx := p.buildTemplateContext(task)
		
		// Create a virtual step for failover
		failoverStep := models.WebhookStep{
			URL:           pipeline.FailoverURL,
			Method:        models.WebhookMethodPOST,
			TimeoutMs:     pipeline.TimeoutMs,
			SignatureMode: models.WebhookSignatureHMAC,
		}

		result := p.executeStep(&failoverStep, ctx, task, -1)
		
		if result.Success {
			p.observability.Log(LogEvent{
				Category:      "webhook_failover_success",
				Level:         LogLevelInfo,
				Message:       "Failover webhook succeeded",
				CorrelationID: task.CorrelationID,
			})
			return
		}
	}

	// Failover failed, move to DLQ
	p.queueService.EnqueueDLQ(task)
	
	p.observability.Log(LogEvent{
		Category:      "webhook_failover_failed",
		Level:         LogLevelError,
		Message:       "Failover failed, moved to DLQ",
		CorrelationID: task.CorrelationID,
	})
}

// ============================================
// STEP EXECUTION
// ============================================

// StepExecutionResult represents the result of executing a step
type StepExecutionResult struct {
	Success      bool
	StatusCode   int
	ResponseBody string
	Error        string
	DurationMs   int64
}

// executeStep executes a single webhook step
func (p *WebhookWorkerPool) executeStep(
	step *models.WebhookStep,
	ctx *TemplateContext,
	task *models.WebhookTask,
	stepIndex int,
) *StepExecutionResult {
	startTime := time.Now()
	atomic.AddInt64(&p.metrics.StepsExecuted, 1)

	result := &StepExecutionResult{}

	// Render URL
	url, err := p.templateEngine.RenderURL(step.URL, ctx)
	if err != nil {
		result.Error = fmt.Sprintf("failed to render URL: %v", err)
		atomic.AddInt64(&p.metrics.StepsFailed, 1)
		return result
	}

	// Render body
	var body []byte
	if step.BodyTemplate != "" {
		renderedBody, err := p.templateEngine.RenderBody(step.BodyTemplate, ctx)
		if err != nil {
			result.Error = fmt.Sprintf("failed to render body: %v", err)
			atomic.AddInt64(&p.metrics.StepsFailed, 1)
			return result
		}
		body = []byte(renderedBody)
	}

	// Create request
	req, err := http.NewRequest(string(step.Method), url, bytes.NewReader(body))
	if err != nil {
		result.Error = fmt.Sprintf("failed to create request: %v", err)
		atomic.AddInt64(&p.metrics.StepsFailed, 1)
		return result
	}

	// Render and add headers
	var headers map[string]string
	if step.Headers != nil {
		json.Unmarshal(step.Headers, &headers)
	}
	if headers == nil {
		headers = make(map[string]string)
	}

	renderedHeaders, _ := p.templateEngine.RenderHeaders(headers, ctx)
	for k, v := range renderedHeaders {
		req.Header.Set(k, v)
	}

	// Sign request
	signedReq, err := p.signingService.SignRequest(
		SigningMode(step.SignatureMode),
		body,
		step.SigningKey,
		task.ID,
		task.AdvertiserID,
		task.PipelineID,
		task.ExecutionID,
		stepIndex,
	)
	if err != nil {
		result.Error = fmt.Sprintf("failed to sign request: %v", err)
		atomic.AddInt64(&p.metrics.StepsFailed, 1)
		return result
	}

	// Add signature headers
	for k, v := range signedReq.Headers {
		req.Header.Set(k, v)
	}

	// Add metadata headers
	p.signingService.AddWebhookMetadataHeaders(
		nil,
		task.ID,
		task.PipelineID,
		task.ExecutionID,
		stepIndex,
		task.Attempts,
	)

	// Set timeout
	httpCtx, cancel := context.WithTimeout(context.Background(), time.Duration(step.TimeoutMs)*time.Millisecond)
	defer cancel()
	req = req.WithContext(httpCtx)

	// Execute request
	resp, err := p.httpClient.Do(req)
	result.DurationMs = time.Since(startTime).Milliseconds()

	if err != nil {
		result.Error = fmt.Sprintf("request failed: %v", err)
		atomic.AddInt64(&p.metrics.StepsFailed, 1)
		
		p.observability.Log(LogEvent{
			Category:      "webhook_step_failure",
			Level:         LogLevelError,
			Message:       "Webhook step failed",
			CorrelationID: task.CorrelationID,
			Metadata: map[string]interface{}{
				"step_id":    step.ID.String(),
				"step_order": stepIndex,
				"error":      err.Error(),
				"url":        url,
			},
			DurationMs: result.DurationMs,
		})
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode

	// Read response body
	respBody, _ := io.ReadAll(resp.Body)
	result.ResponseBody = string(respBody)

	// Check status code
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		result.Success = true
		atomic.AddInt64(&p.metrics.StepsSucceeded, 1)
		
		p.observability.Log(LogEvent{
			Category:      "webhook_step_success",
			Level:         LogLevelInfo,
			Message:       "Webhook step succeeded",
			CorrelationID: task.CorrelationID,
			Metadata: map[string]interface{}{
				"step_id":     step.ID.String(),
				"step_order":  stepIndex,
				"status_code": resp.StatusCode,
				"url":         url,
			},
			DurationMs: result.DurationMs,
		})
	} else {
		result.Error = fmt.Sprintf("unexpected status code: %d", resp.StatusCode)
		atomic.AddInt64(&p.metrics.StepsFailed, 1)
		
		p.observability.Log(LogEvent{
			Category:      "webhook_step_failure",
			Level:         LogLevelError,
			Message:       "Webhook step returned error status",
			CorrelationID: task.CorrelationID,
			Metadata: map[string]interface{}{
				"step_id":       step.ID.String(),
				"step_order":    stepIndex,
				"status_code":   resp.StatusCode,
				"response_body": string(respBody[:min(500, len(respBody))]),
				"url":           url,
			},
			DurationMs: result.DurationMs,
		})
	}

	return result
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// buildTemplateContext builds a template context from task payload
func (p *WebhookWorkerPool) buildTemplateContext(task *models.WebhookTask) *TemplateContext {
	ctx := NewTemplateContext()
	ctx.TaskID = task.ID
	ctx.CorrelationID = task.CorrelationID

	// Extract data from payload
	if click, ok := task.Payload["click"].(map[string]interface{}); ok {
		ctx.Click = click
	}
	if conversion, ok := task.Payload["conversion"].(map[string]interface{}); ok {
		ctx.Conversion = conversion
	}
	if userOffer, ok := task.Payload["user_offer"].(map[string]interface{}); ok {
		ctx.UserOffer = userOffer
	}
	if offer, ok := task.Payload["offer"].(map[string]interface{}); ok {
		ctx.Offer = offer
	}
	if user, ok := task.Payload["user"].(map[string]interface{}); ok {
		ctx.User = user
	}
	if postback, ok := task.Payload["postback"].(map[string]interface{}); ok {
		ctx.Postback = postback
	}
	if custom, ok := task.Payload["custom"].(map[string]interface{}); ok {
		ctx.Custom = custom
	}

	return ctx
}

// handleTaskError handles a task error
func (p *WebhookWorkerPool) handleTaskError(task *models.WebhookTask, err error) {
	atomic.AddInt64(&p.metrics.TasksFailed, 1)
	task.LastError = err.Error()

	// Update execution status
	p.db.Model(&models.WebhookExecution{}).
		Where("id = ?", task.ExecutionID).
		Updates(map[string]interface{}{
			"status":     models.WebhookExecutionRetrying,
			"attempts":   task.Attempts + 1,
			"last_error": task.LastError,
		})

	// Schedule retry
	policy := DefaultRetryPolicy()
	if err := p.queueService.ScheduleRetry(task, policy); err != nil {
		// Retry scheduling failed, move to DLQ
		p.queueService.EnqueueDLQ(task)
	} else {
		atomic.AddInt64(&p.metrics.TasksRetried, 1)
		
		p.observability.Log(LogEvent{
			Category:      "webhook_retry",
			Level:         LogLevelWarn,
			Message:       "Webhook task scheduled for retry",
			CorrelationID: task.CorrelationID,
			Metadata: map[string]interface{}{
				"task_id":       task.ID,
				"attempt":       task.Attempts,
				"next_retry_at": task.NextRetryAt,
				"error":         task.LastError,
			},
		})
	}
}

// storeStepResult stores the result of a step execution
func (p *WebhookWorkerPool) storeStepResult(
	execution *models.WebhookExecution,
	step *models.WebhookStep,
	result *StepExecutionResult,
	stepOrder int,
	attempt int,
) {
	status := models.WebhookExecutionSuccess
	if !result.Success {
		status = models.WebhookExecutionFailed
	}

	now := time.Now()
	stepResult := models.WebhookStepResult{
		ID:           uuid.New(),
		ExecutionID:  execution.ID,
		StepID:       step.ID,
		StepOrder:    stepOrder,
		Status:       status,
		Attempt:      attempt,
		RequestURL:   step.URL,
		ResponseCode: result.StatusCode,
		ResponseBody: result.ResponseBody,
		ErrorMessage: result.Error,
		DurationMs:   result.DurationMs,
		StartedAt:    &now,
		CompletedAt:  &now,
	}

	p.db.Create(&stepResult)
}

// storeDLQItem stores a task in the DLQ table
func (p *WebhookWorkerPool) storeDLQItem(task *models.WebhookTask) {
	taskData, _ := json.Marshal(task)

	dlqItem := models.WebhookDLQItem{
		ID:            uuid.New(),
		ExecutionID:   task.ExecutionID,
		PipelineID:    task.PipelineID,
		TaskData:      taskData,
		FailureReason: task.LastError,
		Attempts:      task.Attempts,
		CanRetry:      true,
	}

	p.db.Create(&dlqItem)

	// Update execution status
	p.db.Model(&models.WebhookExecution{}).
		Where("id = ?", task.ExecutionID).
		Update("status", models.WebhookExecutionDLQ)

	p.observability.Log(LogEvent{
		Category:      "webhook_dlq",
		Level:         LogLevelError,
		Message:       "Task moved to DLQ",
		CorrelationID: task.CorrelationID,
		Metadata: map[string]interface{}{
			"task_id":     task.ID,
			"pipeline_id": task.PipelineID.String(),
			"attempts":    task.Attempts,
			"error":       task.LastError,
		},
	})
}

// GetMetrics returns worker pool metrics
func (p *WebhookWorkerPool) GetMetrics() *webhookWorkerMetrics {
	return &webhookWorkerMetrics{
		TasksProcessed: atomic.LoadInt64(&p.metrics.TasksProcessed),
		TasksSucceeded: atomic.LoadInt64(&p.metrics.TasksSucceeded),
		TasksFailed:    atomic.LoadInt64(&p.metrics.TasksFailed),
		TasksRetried:   atomic.LoadInt64(&p.metrics.TasksRetried),
		StepsExecuted:  atomic.LoadInt64(&p.metrics.StepsExecuted),
		StepsSucceeded: atomic.LoadInt64(&p.metrics.StepsSucceeded),
		StepsFailed:    atomic.LoadInt64(&p.metrics.StepsFailed),
		TotalLatencyMs: atomic.LoadInt64(&p.metrics.TotalLatencyMs),
		ActiveWorkers:  atomic.LoadInt64(&p.metrics.ActiveWorkers),
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


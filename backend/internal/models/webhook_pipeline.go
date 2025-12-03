package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ============================================
// WEBHOOK PIPELINE MODELS
// ============================================

// WebhookPipelineStatus represents pipeline status
type WebhookPipelineStatus string

const (
	WebhookPipelineStatusActive   WebhookPipelineStatus = "active"
	WebhookPipelineStatusInactive WebhookPipelineStatus = "inactive"
	WebhookPipelineStatusDraft    WebhookPipelineStatus = "draft"
)

// WebhookTriggerType represents when the webhook fires
type WebhookTriggerType string

const (
	WebhookTriggerClick      WebhookTriggerType = "click"
	WebhookTriggerConversion WebhookTriggerType = "conversion"
	WebhookTriggerPostback   WebhookTriggerType = "postback"
	WebhookTriggerJoinOffer  WebhookTriggerType = "join_offer"
	WebhookTriggerCustom     WebhookTriggerType = "custom"
)

// WebhookPipeline represents a multi-step webhook pipeline
type WebhookPipeline struct {
	ID           uuid.UUID             `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name         string                `json:"name" gorm:"size:255;not null"`
	Description  string                `json:"description" gorm:"size:1000"`
	AdvertiserID *uuid.UUID            `json:"advertiser_id,omitempty" gorm:"type:uuid;index"`
	OfferID      *uuid.UUID            `json:"offer_id,omitempty" gorm:"type:uuid;index"`
	TriggerType  WebhookTriggerType    `json:"trigger_type" gorm:"size:50;not null;index"`
	Status       WebhookPipelineStatus `json:"status" gorm:"size:20;default:'draft';index"`
	FailoverURL  string                `json:"failover_url,omitempty" gorm:"size:2048"`
	MaxRetries   int                   `json:"max_retries" gorm:"default:5"`
	TimeoutMs    int                   `json:"timeout_ms" gorm:"default:30000"`
	Priority     int                   `json:"priority" gorm:"default:0;index"`
	Metadata     datatypes.JSON        `json:"metadata,omitempty" gorm:"type:jsonb"`
	CreatedAt    time.Time             `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time             `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Steps []WebhookStep `json:"steps,omitempty" gorm:"foreignKey:PipelineID;constraint:OnDelete:CASCADE"`
}

func (WebhookPipeline) TableName() string {
	return "webhook_pipelines"
}

// ============================================
// WEBHOOK STEP
// ============================================

// WebhookStepMethod represents HTTP method
type WebhookStepMethod string

const (
	WebhookMethodGET  WebhookStepMethod = "GET"
	WebhookMethodPOST WebhookStepMethod = "POST"
	WebhookMethodPUT  WebhookStepMethod = "PUT"
)

// WebhookSignatureMode represents signature type
type WebhookSignatureMode string

const (
	WebhookSignatureNone WebhookSignatureMode = "none"
	WebhookSignatureHMAC WebhookSignatureMode = "hmac"
	WebhookSignatureJWT  WebhookSignatureMode = "jwt"
)

// WebhookBackoffMode represents retry backoff strategy
type WebhookBackoffMode string

const (
	WebhookBackoffFixed       WebhookBackoffMode = "fixed"
	WebhookBackoffExponential WebhookBackoffMode = "exponential"
)

// WebhookStep represents a single step in a pipeline
type WebhookStep struct {
	ID            uuid.UUID            `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PipelineID    uuid.UUID            `json:"pipeline_id" gorm:"type:uuid;not null;index"`
	StepOrder     int                  `json:"step_order" gorm:"not null;index"`
	Name          string               `json:"name" gorm:"size:255;not null"`
	URL           string               `json:"url" gorm:"size:2048;not null"`
	Method        WebhookStepMethod    `json:"method" gorm:"size:10;default:'POST'"`
	Headers       datatypes.JSON       `json:"headers,omitempty" gorm:"type:jsonb"`
	BodyTemplate  string               `json:"body_template,omitempty" gorm:"type:text"`
	TimeoutMs     int                  `json:"timeout_ms" gorm:"default:10000"`
	MaxAttempts   int                  `json:"max_attempts" gorm:"default:3"`
	BackoffMode   WebhookBackoffMode   `json:"backoff_mode" gorm:"size:20;default:'exponential'"`
	BackoffBaseMs int                  `json:"backoff_base_ms" gorm:"default:5000"`
	StopOnFailure bool                 `json:"stop_on_failure" gorm:"default:true"`
	SignatureMode WebhookSignatureMode `json:"signature_mode" gorm:"size:20;default:'none'"`
	SigningKey    string               `json:"signing_key,omitempty" gorm:"size:512"`
	Conditions    datatypes.JSON       `json:"conditions,omitempty" gorm:"type:jsonb"`
	CreatedAt     time.Time            `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time            `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Pipeline *WebhookPipeline `json:"pipeline,omitempty" gorm:"foreignKey:PipelineID"`
}

func (WebhookStep) TableName() string {
	return "webhook_steps"
}

// ============================================
// WEBHOOK EXECUTION
// ============================================

// WebhookExecutionStatus represents execution status
type WebhookExecutionStatus string

const (
	WebhookExecutionPending    WebhookExecutionStatus = "pending"
	WebhookExecutionRunning    WebhookExecutionStatus = "running"
	WebhookExecutionSuccess    WebhookExecutionStatus = "success"
	WebhookExecutionFailed     WebhookExecutionStatus = "failed"
	WebhookExecutionRetrying   WebhookExecutionStatus = "retrying"
	WebhookExecutionFailover   WebhookExecutionStatus = "failover"
	WebhookExecutionDLQ        WebhookExecutionStatus = "dlq"
	WebhookExecutionCancelled  WebhookExecutionStatus = "cancelled"
)

// WebhookExecution represents a pipeline execution instance
type WebhookExecution struct {
	ID            uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	PipelineID    uuid.UUID              `json:"pipeline_id" gorm:"type:uuid;not null;index"`
	TriggerType   WebhookTriggerType     `json:"trigger_type" gorm:"size:50;not null;index"`
	TriggerID     string                 `json:"trigger_id" gorm:"size:100;index"`
	CorrelationID string                 `json:"correlation_id" gorm:"size:50;index"`
	Status        WebhookExecutionStatus `json:"status" gorm:"size:20;default:'pending';index"`
	CurrentStep   int                    `json:"current_step" gorm:"default:0"`
	TotalSteps    int                    `json:"total_steps" gorm:"default:0"`
	Attempts      int                    `json:"attempts" gorm:"default:0"`
	MaxAttempts   int                    `json:"max_attempts" gorm:"default:5"`
	Payload       datatypes.JSON         `json:"payload" gorm:"type:jsonb"`
	LastError     string                 `json:"last_error,omitempty" gorm:"type:text"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	NextRetryAt   *time.Time             `json:"next_retry_at,omitempty" gorm:"index"`
	DurationMs    int64                  `json:"duration_ms" gorm:"default:0"`
	CreatedAt     time.Time              `json:"created_at" gorm:"autoCreateTime;index"`
	UpdatedAt     time.Time              `json:"updated_at" gorm:"autoUpdateTime"`

	// Relations
	Pipeline    *WebhookPipeline     `json:"pipeline,omitempty" gorm:"foreignKey:PipelineID"`
	StepResults []WebhookStepResult  `json:"step_results,omitempty" gorm:"foreignKey:ExecutionID;constraint:OnDelete:CASCADE"`
}

func (WebhookExecution) TableName() string {
	return "webhook_executions"
}

// ============================================
// WEBHOOK STEP RESULT
// ============================================

// WebhookStepResult represents the result of a single step execution
type WebhookStepResult struct {
	ID           uuid.UUID              `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExecutionID  uuid.UUID              `json:"execution_id" gorm:"type:uuid;not null;index"`
	StepID       uuid.UUID              `json:"step_id" gorm:"type:uuid;not null;index"`
	StepOrder    int                    `json:"step_order" gorm:"not null"`
	Status       WebhookExecutionStatus `json:"status" gorm:"size:20;default:'pending'"`
	Attempt      int                    `json:"attempt" gorm:"default:1"`
	RequestURL   string                 `json:"request_url" gorm:"size:2048"`
	RequestBody  string                 `json:"request_body,omitempty" gorm:"type:text"`
	ResponseCode int                    `json:"response_code" gorm:"default:0"`
	ResponseBody string                 `json:"response_body,omitempty" gorm:"type:text"`
	ErrorMessage string                 `json:"error_message,omitempty" gorm:"type:text"`
	DurationMs   int64                  `json:"duration_ms" gorm:"default:0"`
	StartedAt    *time.Time             `json:"started_at,omitempty"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	CreatedAt    time.Time              `json:"created_at" gorm:"autoCreateTime"`

	// Relations
	Execution *WebhookExecution `json:"execution,omitempty" gorm:"foreignKey:ExecutionID"`
	Step      *WebhookStep      `json:"step,omitempty" gorm:"foreignKey:StepID"`
}

func (WebhookStepResult) TableName() string {
	return "webhook_step_results"
}

// ============================================
// WEBHOOK TASK (Queue Item)
// ============================================

// WebhookTaskQueue represents which queue the task is in
type WebhookTaskQueue string

const (
	WebhookQueuePrimary  WebhookTaskQueue = "primary"
	WebhookQueueFailover WebhookTaskQueue = "failover"
	WebhookQueueDLQ      WebhookTaskQueue = "dlq"
)

// WebhookTask represents a task in the webhook queue
type WebhookTask struct {
	ID            string           `json:"id"`
	ExecutionID   uuid.UUID        `json:"execution_id"`
	PipelineID    uuid.UUID        `json:"pipeline_id"`
	AdvertiserID  *uuid.UUID       `json:"advertiser_id,omitempty"`
	OfferID       *uuid.UUID       `json:"offer_id,omitempty"`
	StepIndex     int              `json:"step_index"`
	Payload       map[string]interface{} `json:"payload"`
	Attempts      int              `json:"attempts"`
	MaxAttempts   int              `json:"max_attempts"`
	Queue         WebhookTaskQueue `json:"queue"`
	Priority      int              `json:"priority"`
	CreatedAt     time.Time        `json:"created_at"`
	LastError     string           `json:"last_error,omitempty"`
	NextRetryAt   time.Time        `json:"next_retry_at"`
	CorrelationID string           `json:"correlation_id"`
}

// ============================================
// WEBHOOK METRICS
// ============================================

// WebhookMetrics holds webhook system metrics
type WebhookMetrics struct {
	TotalTasks        int64   `json:"total_tasks"`
	TotalSuccess      int64   `json:"total_success"`
	TotalFailures     int64   `json:"total_failures"`
	TotalRetries      int64   `json:"total_retries"`
	FailoverTriggered int64   `json:"failover_triggered"`
	DLQItems          int64   `json:"dlq_items"`
	AvgLatencyMs      float64 `json:"avg_latency_ms"`
	StepSuccessCount  int64   `json:"step_success_count"`
	StepFailureCount  int64   `json:"step_failure_count"`
	PendingTasks      int64   `json:"pending_tasks"`
	RunningTasks      int64   `json:"running_tasks"`
	QueueSizes        map[string]int64 `json:"queue_sizes"`
}

// ============================================
// WEBHOOK DLQ ITEM (Dead Letter Queue)
// ============================================

// WebhookDLQItem represents an item in the dead letter queue
type WebhookDLQItem struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ExecutionID   uuid.UUID      `json:"execution_id" gorm:"type:uuid;not null;index"`
	PipelineID    uuid.UUID      `json:"pipeline_id" gorm:"type:uuid;not null;index"`
	TaskData      datatypes.JSON `json:"task_data" gorm:"type:jsonb"`
	FailureReason string         `json:"failure_reason" gorm:"type:text"`
	Attempts      int            `json:"attempts" gorm:"default:0"`
	CanRetry      bool           `json:"can_retry" gorm:"default:true"`
	RetriedAt     *time.Time     `json:"retried_at,omitempty"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime;index"`

	// Relations
	Execution *WebhookExecution `json:"execution,omitempty" gorm:"foreignKey:ExecutionID"`
	Pipeline  *WebhookPipeline  `json:"pipeline,omitempty" gorm:"foreignKey:PipelineID"`
}

func (WebhookDLQItem) TableName() string {
	return "webhook_dlq_items"
}


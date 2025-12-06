package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
)

// ============================================
// WEBHOOK QUEUE SERVICE
// ============================================

// WebhookQueueService manages webhook task queues
type WebhookQueueService struct {
	// In-memory queues (L1)
	primaryQueue  chan *models.WebhookTask
	failoverQueue chan *models.WebhookTask
	dlqQueue      chan *models.WebhookTask

	// Queue sizes
	primarySize  int
	failoverSize int
	dlqSize      int

	// Metrics
	metrics *webhookQueueMetrics

	// State
	running bool
	mutex   sync.RWMutex
}

// webhookQueueMetrics tracks queue metrics
type webhookQueueMetrics struct {
	TotalEnqueued     int64
	TotalDequeued     int64
	TotalFailover     int64
	TotalDLQ          int64
	TotalDropped      int64
	PrimaryQueueSize  int64
	FailoverQueueSize int64
	DLQQueueSize      int64
}

// Redis queue keys
const (
	RedisKeyPrimaryQueue  = "webhook:queue:primary"
	RedisKeyFailoverQueue = "webhook:queue:failover"
	RedisKeyDLQQueue      = "webhook:queue:dlq"
	RedisKeyTaskPrefix    = "webhook:task:"
)

// NewWebhookQueueService creates a new webhook queue service
func NewWebhookQueueService() *WebhookQueueService {
	cpuCount := runtime.NumCPU()

	return &WebhookQueueService{
		primaryQueue:  make(chan *models.WebhookTask, cpuCount*10000),
		failoverQueue: make(chan *models.WebhookTask, cpuCount*5000),
		dlqQueue:      make(chan *models.WebhookTask, cpuCount*2000),
		primarySize:   cpuCount * 10000,
		failoverSize:  cpuCount * 5000,
		dlqSize:       cpuCount * 2000,
		metrics:       &webhookQueueMetrics{},
		running:       true,
	}
}

// ============================================
// ENQUEUE OPERATIONS
// ============================================

// EnqueuePrimary adds a task to the primary queue
func (q *WebhookQueueService) EnqueuePrimary(task *models.WebhookTask) error {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	if !q.running {
		return fmt.Errorf("queue service is not running")
	}

	task.Queue = models.WebhookQueuePrimary
	task.CreatedAt = time.Now()

	// Try L1 (in-memory) first
	select {
	case q.primaryQueue <- task:
		atomic.AddInt64(&q.metrics.TotalEnqueued, 1)
		atomic.AddInt64(&q.metrics.PrimaryQueueSize, 1)
		return nil
	default:
		// L1 full, try L2 (Redis)
		return q.enqueueRedis(RedisKeyPrimaryQueue, task)
	}
}

// EnqueueFailover adds a task to the failover queue
func (q *WebhookQueueService) EnqueueFailover(task *models.WebhookTask) error {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	if !q.running {
		return fmt.Errorf("queue service is not running")
	}

	task.Queue = models.WebhookQueueFailover

	// Try L1 first
	select {
	case q.failoverQueue <- task:
		atomic.AddInt64(&q.metrics.TotalFailover, 1)
		atomic.AddInt64(&q.metrics.FailoverQueueSize, 1)
		return nil
	default:
		// L1 full, try L2
		return q.enqueueRedis(RedisKeyFailoverQueue, task)
	}
}

// EnqueueDLQ adds a task to the dead letter queue
func (q *WebhookQueueService) EnqueueDLQ(task *models.WebhookTask) error {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	if !q.running {
		return fmt.Errorf("queue service is not running")
	}

	task.Queue = models.WebhookQueueDLQ

	// Try L1 first
	select {
	case q.dlqQueue <- task:
		atomic.AddInt64(&q.metrics.TotalDLQ, 1)
		atomic.AddInt64(&q.metrics.DLQQueueSize, 1)
		return nil
	default:
		// L1 full, try L2
		return q.enqueueRedis(RedisKeyDLQQueue, task)
	}
}

// enqueueRedis adds a task to Redis queue (L2)
func (q *WebhookQueueService) enqueueRedis(queueKey string, task *models.WebhookTask) error {
	ctx := context.Background()

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	// Use Redis sorted set as queue
	if cache.RedisClient != nil {
		score := float64(task.Priority)*1e12 + float64(task.CreatedAt.UnixNano())
		err = cache.RedisClient.ZAdd(ctx, queueKey, cache.RedisZ{
			Score:  score,
			Member: string(data),
		}).Err()
		if err != nil {
			atomic.AddInt64(&q.metrics.TotalDropped, 1)
			return fmt.Errorf("failed to enqueue to Redis: %w", err)
		}
	}

	atomic.AddInt64(&q.metrics.TotalEnqueued, 1)
	return nil
}

// ============================================
// DEQUEUE OPERATIONS
// ============================================

// DequeuePrimary gets a task from the primary queue
func (q *WebhookQueueService) DequeuePrimary(timeout time.Duration) (*models.WebhookTask, error) {
	// Try L1 first
	select {
	case task := <-q.primaryQueue:
		atomic.AddInt64(&q.metrics.TotalDequeued, 1)
		atomic.AddInt64(&q.metrics.PrimaryQueueSize, -1)
		return task, nil
	case <-time.After(timeout / 2):
		// Try L2
		return q.dequeueRedis(RedisKeyPrimaryQueue)
	}
}

// DequeueFailover gets a task from the failover queue
func (q *WebhookQueueService) DequeueFailover(timeout time.Duration) (*models.WebhookTask, error) {
	select {
	case task := <-q.failoverQueue:
		atomic.AddInt64(&q.metrics.TotalDequeued, 1)
		atomic.AddInt64(&q.metrics.FailoverQueueSize, -1)
		return task, nil
	case <-time.After(timeout / 2):
		return q.dequeueRedis(RedisKeyFailoverQueue)
	}
}

// DequeueDLQ gets a task from the DLQ
func (q *WebhookQueueService) DequeueDLQ(timeout time.Duration) (*models.WebhookTask, error) {
	select {
	case task := <-q.dlqQueue:
		atomic.AddInt64(&q.metrics.TotalDequeued, 1)
		atomic.AddInt64(&q.metrics.DLQQueueSize, -1)
		return task, nil
	case <-time.After(timeout / 2):
		return q.dequeueRedis(RedisKeyDLQQueue)
	}
}

// dequeueRedis gets a task from Redis queue (L2)
func (q *WebhookQueueService) dequeueRedis(queueKey string) (*models.WebhookTask, error) {
	ctx := context.Background()

	if cache.RedisClient == nil {
		return nil, nil
	}

	// Pop highest priority item
	results, err := cache.RedisClient.ZPopMax(ctx, queueKey, 1).Result()
	if err != nil || len(results) == 0 {
		return nil, nil
	}

	var task models.WebhookTask
	if err := json.Unmarshal([]byte(results[0].Member.(string)), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	atomic.AddInt64(&q.metrics.TotalDequeued, 1)
	return &task, nil
}

// ============================================
// RETRY LOGIC
// ============================================

// RetryPolicy defines retry behavior
type RetryPolicy struct {
	MaxAttempts   int
	BackoffMode   models.WebhookBackoffMode
	BackoffBaseMs int
	MaxBackoffMs  int
	JitterPercent int // 0-100
}

// DefaultRetryPolicy returns the default retry policy
func DefaultRetryPolicy() *RetryPolicy {
	return &RetryPolicy{
		MaxAttempts:   5,
		BackoffMode:   models.WebhookBackoffExponential,
		BackoffBaseMs: 5000,  // 5 seconds
		MaxBackoffMs:  300000, // 5 minutes
		JitterPercent: 20,
	}
}

// CalculateBackoff calculates the next retry delay
func (p *RetryPolicy) CalculateBackoff(attempt int) time.Duration {
	var delayMs int

	switch p.BackoffMode {
	case models.WebhookBackoffFixed:
		delayMs = p.BackoffBaseMs
	case models.WebhookBackoffExponential:
		// Exponential: base * 2^(attempt-1)
		// 5s → 10s → 20s → 40s → 80s → capped at max
		delayMs = p.BackoffBaseMs * int(math.Pow(2, float64(attempt-1)))
		if delayMs > p.MaxBackoffMs {
			delayMs = p.MaxBackoffMs
		}
	default:
		delayMs = p.BackoffBaseMs
	}

	// Add jitter
	if p.JitterPercent > 0 {
		jitter := delayMs * p.JitterPercent / 100
		delayMs += rand.Intn(jitter*2) - jitter
	}

	return time.Duration(delayMs) * time.Millisecond
}

// ShouldRetry determines if a task should be retried
func (p *RetryPolicy) ShouldRetry(attempt int) bool {
	return attempt < p.MaxAttempts
}

// ScheduleRetry schedules a task for retry
func (q *WebhookQueueService) ScheduleRetry(task *models.WebhookTask, policy *RetryPolicy) error {
	task.Attempts++
	
	if !policy.ShouldRetry(task.Attempts) {
		// Max retries exceeded, move to failover
		return q.EnqueueFailover(task)
	}

	// Calculate next retry time
	backoff := policy.CalculateBackoff(task.Attempts)
	task.NextRetryAt = time.Now().Add(backoff)

	// Re-enqueue to primary queue
	return q.EnqueuePrimary(task)
}

// ============================================
// QUEUE MANAGEMENT
// ============================================

// GetQueueSizes returns current queue sizes
func (q *WebhookQueueService) GetQueueSizes() map[string]int64 {
	return map[string]int64{
		"primary":  atomic.LoadInt64(&q.metrics.PrimaryQueueSize) + q.getRedisQueueSize(RedisKeyPrimaryQueue),
		"failover": atomic.LoadInt64(&q.metrics.FailoverQueueSize) + q.getRedisQueueSize(RedisKeyFailoverQueue),
		"dlq":      atomic.LoadInt64(&q.metrics.DLQQueueSize) + q.getRedisQueueSize(RedisKeyDLQQueue),
	}
}

// getRedisQueueSize gets the size of a Redis queue
func (q *WebhookQueueService) getRedisQueueSize(queueKey string) int64 {
	ctx := context.Background()
	
	if cache.RedisClient == nil {
		return 0
	}

	count, err := cache.RedisClient.ZCard(ctx, queueKey).Result()
	if err != nil {
		return 0
	}

	return count
}

// GetMetrics returns queue metrics
func (q *WebhookQueueService) GetMetrics() *webhookQueueMetrics {
	return &webhookQueueMetrics{
		TotalEnqueued:     atomic.LoadInt64(&q.metrics.TotalEnqueued),
		TotalDequeued:     atomic.LoadInt64(&q.metrics.TotalDequeued),
		TotalFailover:     atomic.LoadInt64(&q.metrics.TotalFailover),
		TotalDLQ:          atomic.LoadInt64(&q.metrics.TotalDLQ),
		TotalDropped:      atomic.LoadInt64(&q.metrics.TotalDropped),
		PrimaryQueueSize:  atomic.LoadInt64(&q.metrics.PrimaryQueueSize),
		FailoverQueueSize: atomic.LoadInt64(&q.metrics.FailoverQueueSize),
		DLQQueueSize:      atomic.LoadInt64(&q.metrics.DLQQueueSize),
	}
}

// ClearQueue clears a specific queue
func (q *WebhookQueueService) ClearQueue(queueType models.WebhookTaskQueue) error {
	ctx := context.Background()

	switch queueType {
	case models.WebhookQueuePrimary:
		// Drain L1
		for len(q.primaryQueue) > 0 {
			<-q.primaryQueue
		}
		atomic.StoreInt64(&q.metrics.PrimaryQueueSize, 0)
		// Clear L2
		if cache.RedisClient != nil {
			cache.RedisClient.Del(ctx, RedisKeyPrimaryQueue)
		}

	case models.WebhookQueueFailover:
		for len(q.failoverQueue) > 0 {
			<-q.failoverQueue
		}
		atomic.StoreInt64(&q.metrics.FailoverQueueSize, 0)
		if cache.RedisClient != nil {
			cache.RedisClient.Del(ctx, RedisKeyFailoverQueue)
		}

	case models.WebhookQueueDLQ:
		for len(q.dlqQueue) > 0 {
			<-q.dlqQueue
		}
		atomic.StoreInt64(&q.metrics.DLQQueueSize, 0)
		if cache.RedisClient != nil {
			cache.RedisClient.Del(ctx, RedisKeyDLQQueue)
		}
	}

	return nil
}

// Stop stops the queue service
func (q *WebhookQueueService) Stop() {
	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.running = false
}

// ============================================
// TASK HELPERS
// ============================================

// CreateTask creates a new webhook task
func CreateWebhookTask(
	executionID, pipelineID uuid.UUID,
	advertiserID, offerID *uuid.UUID,
	stepIndex int,
	payload map[string]interface{},
	maxAttempts int,
	priority int,
) *models.WebhookTask {
	return &models.WebhookTask{
		ID:            uuid.New().String(),
		ExecutionID:   executionID,
		PipelineID:    pipelineID,
		AdvertiserID:  advertiserID,
		OfferID:       offerID,
		StepIndex:     stepIndex,
		Payload:       payload,
		Attempts:      0,
		MaxAttempts:   maxAttempts,
		Queue:         models.WebhookQueuePrimary,
		Priority:      priority,
		CreatedAt:     time.Now(),
		CorrelationID: uuid.New().String()[:8],
	}
}

// ============================================
// GLOBAL QUEUE INSTANCE
// ============================================

var (
	webhookQueueInstance *WebhookQueueService
	webhookQueueOnce     sync.Once
)

// GetWebhookQueueService returns the global webhook queue service instance
func GetWebhookQueueService() *WebhookQueueService {
	webhookQueueOnce.Do(func() {
		webhookQueueInstance = NewWebhookQueueService()
	})
	return webhookQueueInstance
}


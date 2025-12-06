package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
)

// ============================================
// FAILOVER QUEUE
// ============================================

// FailoverQueueEvent represents an event in the failover queue
type FailoverQueueEvent struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"` // click, conversion, postback
	TenantID    string                 `json:"tenant_id"`
	Data        map[string]interface{} `json:"data"`
	Timestamp   time.Time              `json:"timestamp"`
	Attempts    int                    `json:"attempts"`
	LastAttempt *time.Time             `json:"last_attempt,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Source      string                 `json:"source"` // edge, backend, api
}

// FailoverQueue provides local failover queuing with retry
type FailoverQueue struct {
	mu            sync.RWMutex
	walService    *WALService
	observability *ObservabilityService
	
	// In-memory buffer
	buffer        []*FailoverQueueEvent
	bufferSize    int
	maxBufferSize int
	
	// Retry settings
	baseRetryMs   int
	maxRetryMs    int
	maxAttempts   int
	
	// State
	isRunning     bool
	stopChan      chan struct{}
	wg            sync.WaitGroup
	
	// Metrics
	totalQueued   int64
	totalSent     int64
	totalFailed   int64
	totalDropped  int64
	
	// Callbacks
	onSend        func(event *FailoverQueueEvent) error
}

// FailoverQueueConfig holds failover queue configuration
type FailoverQueueConfig struct {
	MaxBufferSize int
	BaseRetryMs   int
	MaxRetryMs    int
	MaxAttempts   int
}

// DefaultFailoverQueueConfig returns default configuration
func DefaultFailoverQueueConfig() FailoverQueueConfig {
	return FailoverQueueConfig{
		MaxBufferSize: 500000,
		BaseRetryMs:   100,
		MaxRetryMs:    60000, // 60 seconds
		MaxAttempts:   0,     // unlimited
	}
}

// NewFailoverQueue creates a new failover queue
func NewFailoverQueue(config FailoverQueueConfig, walService *WALService) *FailoverQueue {
	return &FailoverQueue{
		walService:    walService,
		observability: NewObservabilityService(),
		buffer:        make([]*FailoverQueueEvent, 0, 10000),
		maxBufferSize: config.MaxBufferSize,
		baseRetryMs:   config.BaseRetryMs,
		maxRetryMs:    config.MaxRetryMs,
		maxAttempts:   config.MaxAttempts,
		stopChan:      make(chan struct{}),
	}
}

// ============================================
// QUEUE OPERATIONS
// ============================================

// Enqueue adds an event to the failover queue
func (q *FailoverQueue) Enqueue(event *FailoverQueueEvent) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check buffer size
	if q.bufferSize >= q.maxBufferSize {
		atomic.AddInt64(&q.totalDropped, 1)
		return fmt.Errorf("failover queue full: %d events", q.maxBufferSize)
	}

	// Set defaults
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}

	// Write to WAL first for durability
	if q.walService != nil {
		walEventType := WALEventType(event.Type)
		if _, err := q.walService.Append(walEventType, event.TenantID, event.Data); err != nil {
			q.observability.Log(LogEvent{
				Category: LogCategoryErrorEvent,
				Level:    LogLevelError,
				Message:  "Failed to write to WAL",
				Metadata: map[string]interface{}{
					"event_id": event.ID,
					"error":    err.Error(),
				},
			})
		}
	}

	// Add to in-memory buffer
	q.buffer = append(q.buffer, event)
	q.bufferSize++
	atomic.AddInt64(&q.totalQueued, 1)

	// Also persist to Redis for distributed access
	q.persistToRedis(event)

	return nil
}

// EnqueueClick is a convenience method for click events
func (q *FailoverQueue) EnqueueClick(tenantID string, data map[string]interface{}) error {
	event := &FailoverQueueEvent{
		ID:        fmt.Sprintf("click_%d_%d", time.Now().UnixNano(), rand.Int63()),
		Type:      "click",
		TenantID:  tenantID,
		Data:      data,
		Timestamp: time.Now().UTC(),
		Source:    "backend",
	}
	return q.Enqueue(event)
}

// EnqueueConversion is a convenience method for conversion events
func (q *FailoverQueue) EnqueueConversion(tenantID string, data map[string]interface{}) error {
	event := &FailoverQueueEvent{
		ID:        fmt.Sprintf("conv_%d_%d", time.Now().UnixNano(), rand.Int63()),
		Type:      "conversion",
		TenantID:  tenantID,
		Data:      data,
		Timestamp: time.Now().UTC(),
		Source:    "backend",
	}
	return q.Enqueue(event)
}

// EnqueuePostback is a convenience method for postback events
func (q *FailoverQueue) EnqueuePostback(tenantID string, data map[string]interface{}) error {
	event := &FailoverQueueEvent{
		ID:        fmt.Sprintf("pb_%d_%d", time.Now().UnixNano(), rand.Int63()),
		Type:      "postback",
		TenantID:  tenantID,
		Data:      data,
		Timestamp: time.Now().UTC(),
		Source:    "backend",
	}
	return q.Enqueue(event)
}

// persistToRedis persists event to Redis for distributed access
func (q *FailoverQueue) persistToRedis(event *FailoverQueueEvent) {
	ctx := context.Background()
	key := fmt.Sprintf("failover_queue:%s:%s", event.Type, event.ID)
	
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	
	cache.Set(ctx, key, string(data), 24*time.Hour)
}

// ============================================
// RETRY LOGIC
// ============================================

// SetSendCallback sets the callback for sending events
func (q *FailoverQueue) SetSendCallback(callback func(event *FailoverQueueEvent) error) {
	q.onSend = callback
}

// ProcessQueue processes queued events with retry
func (q *FailoverQueue) ProcessQueue() {
	q.mu.Lock()
	if len(q.buffer) == 0 {
		q.mu.Unlock()
		return
	}

	// Take a batch
	batchSize := 100
	if len(q.buffer) < batchSize {
		batchSize = len(q.buffer)
	}
	batch := q.buffer[:batchSize]
	q.buffer = q.buffer[batchSize:]
	q.bufferSize -= batchSize
	q.mu.Unlock()

	// Process batch
	for _, event := range batch {
		if err := q.processEvent(event); err != nil {
			// Re-queue for retry
			q.requeue(event, err)
		}
	}
}

// processEvent processes a single event
func (q *FailoverQueue) processEvent(event *FailoverQueueEvent) error {
	if q.onSend == nil {
		return fmt.Errorf("no send callback configured")
	}

	event.Attempts++
	now := time.Now()
	event.LastAttempt = &now

	err := q.onSend(event)
	if err != nil {
		event.Error = err.Error()
		return err
	}

	atomic.AddInt64(&q.totalSent, 1)
	
	// Mark as processed in WAL
	if q.walService != nil {
		q.walService.MarkProcessed(event.ID)
	}

	return nil
}

// requeue re-queues an event for retry
func (q *FailoverQueue) requeue(event *FailoverQueueEvent, err error) {
	// Check max attempts
	if q.maxAttempts > 0 && event.Attempts >= q.maxAttempts {
		atomic.AddInt64(&q.totalFailed, 1)
		q.observability.Log(LogEvent{
			Category: LogCategoryErrorEvent,
			Level:    LogLevelError,
			Message:  "Event failed after max attempts",
			Metadata: map[string]interface{}{
				"event_id": event.ID,
				"type":     event.Type,
				"attempts": event.Attempts,
				"error":    err.Error(),
			},
		})
		return
	}

	// Calculate backoff with jitter
	backoffMs := q.calculateBackoff(event.Attempts)
	
	// Schedule retry
	go func() {
		time.Sleep(time.Duration(backoffMs) * time.Millisecond)
		q.mu.Lock()
		q.buffer = append(q.buffer, event)
		q.bufferSize++
		q.mu.Unlock()
	}()
}

// calculateBackoff calculates exponential backoff with jitter
func (q *FailoverQueue) calculateBackoff(attempt int) int {
	// Exponential backoff: base * 2^attempt
	backoff := float64(q.baseRetryMs) * math.Pow(2, float64(attempt))
	
	// Cap at max
	if backoff > float64(q.maxRetryMs) {
		backoff = float64(q.maxRetryMs)
	}
	
	// Add jitter (±20%)
	jitter := backoff * 0.2 * (rand.Float64()*2 - 1)
	backoff += jitter
	
	return int(backoff)
}

// ============================================
// LIFECYCLE
// ============================================

// Start starts the failover queue workers
func (q *FailoverQueue) Start() {
	q.isRunning = true

	// Process worker
	q.wg.Add(1)
	go q.processWorker()

	// Redis sync worker
	q.wg.Add(1)
	go q.redisSyncWorker()
}

// Stop stops the failover queue
func (q *FailoverQueue) Stop() {
	q.isRunning = false
	close(q.stopChan)
	q.wg.Wait()
}

// processWorker continuously processes the queue
func (q *FailoverQueue) processWorker() {
	defer q.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-q.stopChan:
			// Drain remaining events
			q.ProcessQueue()
			return
		case <-ticker.C:
			q.ProcessQueue()
		}
	}
}

// redisSyncWorker syncs queue state with Redis
func (q *FailoverQueue) redisSyncWorker() {
	defer q.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-q.stopChan:
			return
		case <-ticker.C:
			q.syncFromRedis()
		}
	}
}

// syncFromRedis loads pending events from Redis
func (q *FailoverQueue) syncFromRedis() {
	// In production, scan Redis for pending events and add to local buffer
	// This enables distributed failover across multiple backend instances
}

// ============================================
// METRICS
// ============================================

// GetStats returns queue statistics
func (q *FailoverQueue) GetStats() map[string]interface{} {
	q.mu.RLock()
	bufferSize := q.bufferSize
	q.mu.RUnlock()

	return map[string]interface{}{
		"buffer_size":      bufferSize,
		"max_buffer_size":  q.maxBufferSize,
		"total_queued":     atomic.LoadInt64(&q.totalQueued),
		"total_sent":       atomic.LoadInt64(&q.totalSent),
		"total_failed":     atomic.LoadInt64(&q.totalFailed),
		"total_dropped":    atomic.LoadInt64(&q.totalDropped),
		"is_running":       q.isRunning,
		"buffer_utilization": float64(bufferSize) / float64(q.maxBufferSize) * 100,
	}
}

// GetQueueDepth returns the current queue depth
func (q *FailoverQueue) GetQueueDepth() int {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return q.bufferSize
}

// ============================================
// GLOBAL INSTANCE
// ============================================

var (
	failoverQueueInstance *FailoverQueue
	failoverQueueOnce     sync.Once
)

// GetFailoverQueue returns the global failover queue instance
func GetFailoverQueue() *FailoverQueue {
	failoverQueueOnce.Do(func() {
		config := DefaultFailoverQueueConfig()
		walService := GetWALService()
		failoverQueueInstance = NewFailoverQueue(config, walService)
	})
	return failoverQueueInstance
}


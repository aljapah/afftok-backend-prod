package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/google/uuid"
)

// ============================================
// POSTBACK QUEUE ITEM
// ============================================

// PostbackQueueItem represents a queued outbound postback
type PostbackQueueItem struct {
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	AdvertiserID string                 `json:"advertiser_id"`
	URL          string                 `json:"url"`
	Method       string                 `json:"method"`
	Headers      map[string]string      `json:"headers"`
	Body         string                 `json:"body"`
	Timestamp    time.Time              `json:"timestamp"`
	Attempts     int                    `json:"attempts"`
	LastAttempt  *time.Time             `json:"last_attempt,omitempty"`
	LastError    string                 `json:"last_error,omitempty"`
	StatusCode   int                    `json:"status_code,omitempty"`
	Response     string                 `json:"response,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// PostbackQueueStatus represents item status
type PostbackQueueStatus string

const (
	PostbackStatusPending   PostbackQueueStatus = "pending"
	PostbackStatusSent      PostbackQueueStatus = "sent"
	PostbackStatusFailed    PostbackQueueStatus = "failed"
	PostbackStatusDLQ       PostbackQueueStatus = "dlq"
)

// ============================================
// POSTBACK QUEUE SERVICE
// ============================================

// PostbackQueueService manages outbound postback queue
type PostbackQueueService struct {
	mu            sync.RWMutex
	walService    *WALService
	observability *ObservabilityService
	httpClient    *http.Client
	
	// Queues
	pendingQueue  []*PostbackQueueItem
	dlq           []*PostbackQueueItem
	
	// Configuration
	maxRetries    int
	baseRetryMs   int
	maxRetryMs    int
	requestTimeout time.Duration
	
	// State
	isRunning     bool
	stopChan      chan struct{}
	wg            sync.WaitGroup
	
	// Metrics
	totalQueued   int64
	totalSent     int64
	totalFailed   int64
	totalDLQ      int64
}

// PostbackQueueConfig holds configuration
type PostbackQueueConfig struct {
	MaxRetries     int
	BaseRetryMs    int
	MaxRetryMs     int
	RequestTimeout time.Duration
}

// DefaultPostbackQueueConfig returns default configuration
func DefaultPostbackQueueConfig() PostbackQueueConfig {
	return PostbackQueueConfig{
		MaxRetries:     10,
		BaseRetryMs:    1000,    // 1 second
		MaxRetryMs:     300000,  // 5 minutes
		RequestTimeout: 30 * time.Second,
	}
}

// NewPostbackQueueService creates a new postback queue service
func NewPostbackQueueService(config PostbackQueueConfig) *PostbackQueueService {
	return &PostbackQueueService{
		walService:     GetWALService(),
		observability:  NewObservabilityService(),
		httpClient: &http.Client{
			Timeout: config.RequestTimeout,
		},
		pendingQueue:   make([]*PostbackQueueItem, 0),
		dlq:            make([]*PostbackQueueItem, 0),
		maxRetries:     config.MaxRetries,
		baseRetryMs:    config.BaseRetryMs,
		maxRetryMs:     config.MaxRetryMs,
		requestTimeout: config.RequestTimeout,
		stopChan:       make(chan struct{}),
	}
}

// ============================================
// QUEUE OPERATIONS
// ============================================

// Enqueue adds a postback to the queue
func (s *PostbackQueueService) Enqueue(item *PostbackQueueItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set defaults
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	if item.Timestamp.IsZero() {
		item.Timestamp = time.Now().UTC()
	}
	if item.Method == "" {
		item.Method = "POST"
	}

	// Write to WAL first
	if s.walService != nil {
		data := map[string]interface{}{
			"id":            item.ID,
			"url":           item.URL,
			"method":        item.Method,
			"headers":       item.Headers,
			"body":          item.Body,
			"advertiser_id": item.AdvertiserID,
		}
		s.walService.Append(WALEventPostback, item.TenantID, data)
	}

	// Add to pending queue
	s.pendingQueue = append(s.pendingQueue, item)
	atomic.AddInt64(&s.totalQueued, 1)

	// Persist to Redis
	s.persistToRedis(item)

	return nil
}

// EnqueuePostback is a convenience method
func (s *PostbackQueueService) EnqueuePostback(
	tenantID, advertiserID, url, method string,
	headers map[string]string,
	body string,
	metadata map[string]interface{},
) error {
	item := &PostbackQueueItem{
		ID:           uuid.New().String(),
		TenantID:     tenantID,
		AdvertiserID: advertiserID,
		URL:          url,
		Method:       method,
		Headers:      headers,
		Body:         body,
		Timestamp:    time.Now().UTC(),
		Metadata:     metadata,
	}
	return s.Enqueue(item)
}

// persistToRedis persists item to Redis
func (s *PostbackQueueService) persistToRedis(item *PostbackQueueItem) {
	ctx := context.Background()
	key := fmt.Sprintf("postback_queue:%s", item.ID)
	
	data, err := json.Marshal(item)
	if err != nil {
		return
	}
	
	cache.Set(ctx, key, string(data), 24*time.Hour)
}

// ============================================
// PROCESSING
// ============================================

// ProcessQueue processes pending postbacks
func (s *PostbackQueueService) ProcessQueue() {
	s.mu.Lock()
	if len(s.pendingQueue) == 0 {
		s.mu.Unlock()
		return
	}

	// Take batch
	batchSize := 10
	if len(s.pendingQueue) < batchSize {
		batchSize = len(s.pendingQueue)
	}
	batch := s.pendingQueue[:batchSize]
	s.pendingQueue = s.pendingQueue[batchSize:]
	s.mu.Unlock()

	// Process batch
	for _, item := range batch {
		s.processItem(item)
	}
}

// processItem processes a single postback item
func (s *PostbackQueueService) processItem(item *PostbackQueueItem) {
	item.Attempts++
	now := time.Now()
	item.LastAttempt = &now

	// Send request
	statusCode, response, err := s.sendPostback(item)
	item.StatusCode = statusCode
	item.Response = response

	if err != nil {
		item.LastError = err.Error()
		s.handleFailure(item)
		return
	}

	// Success
	atomic.AddInt64(&s.totalSent, 1)
	
	// Mark as processed in WAL
	if s.walService != nil {
		s.walService.MarkProcessed(item.ID)
	}

	// Remove from Redis
	ctx := context.Background()
	cache.Delete(ctx, fmt.Sprintf("postback_queue:%s", item.ID))

	s.observability.Log(LogEvent{
		Category: LogCategoryPostbackEvent,
		Level:    LogLevelInfo,
		Message:  "Postback sent successfully",
		Metadata: map[string]interface{}{
			"postback_id":   item.ID,
			"advertiser_id": item.AdvertiserID,
			"url":           item.URL,
			"status_code":   statusCode,
			"attempts":      item.Attempts,
		},
	})
}

// sendPostback sends the HTTP request
func (s *PostbackQueueService) sendPostback(item *PostbackQueueItem) (int, string, error) {
	// Create request
	var bodyReader io.Reader
	if item.Body != "" {
		bodyReader = bytes.NewBufferString(item.Body)
	}

	req, err := http.NewRequest(item.Method, item.URL, bodyReader)
	if err != nil {
		return 0, "", fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range item.Headers {
		req.Header.Set(key, value)
	}

	// Set default content type
	if req.Header.Get("Content-Type") == "" && item.Body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Send request
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return 0, "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	respBody, _ := io.ReadAll(resp.Body)
	response := string(respBody)

	// Check status
	if resp.StatusCode >= 400 {
		return resp.StatusCode, response, fmt.Errorf("HTTP %d: %s", resp.StatusCode, response)
	}

	return resp.StatusCode, response, nil
}

// handleFailure handles a failed postback
func (s *PostbackQueueService) handleFailure(item *PostbackQueueItem) {
	// Check if max retries exceeded
	if item.Attempts >= s.maxRetries {
		s.moveToDLQ(item)
		return
	}

	// Calculate backoff
	backoffMs := s.calculateBackoff(item.Attempts)

	s.observability.Log(LogEvent{
		Category: LogCategoryPostbackEvent,
		Level:    LogLevelWarn,
		Message:  "Postback failed, scheduling retry",
		Metadata: map[string]interface{}{
			"postback_id": item.ID,
			"attempts":    item.Attempts,
			"error":       item.LastError,
			"retry_in_ms": backoffMs,
		},
	})

	// Schedule retry
	go func() {
		time.Sleep(time.Duration(backoffMs) * time.Millisecond)
		s.mu.Lock()
		s.pendingQueue = append(s.pendingQueue, item)
		s.mu.Unlock()
	}()
}

// moveToDLQ moves item to Dead Letter Queue
func (s *PostbackQueueService) moveToDLQ(item *PostbackQueueItem) {
	s.mu.Lock()
	s.dlq = append(s.dlq, item)
	s.mu.Unlock()

	atomic.AddInt64(&s.totalFailed, 1)
	atomic.AddInt64(&s.totalDLQ, 1)

	// Persist to Redis DLQ
	ctx := context.Background()
	key := fmt.Sprintf("postback_dlq:%s", item.ID)
	data, _ := json.Marshal(item)
	cache.Set(ctx, key, string(data), 7*24*time.Hour) // 7 days

	s.observability.Log(LogEvent{
		Category: LogCategoryPostbackEvent,
		Level:    LogLevelError,
		Message:  "Postback moved to DLQ after max retries",
		Metadata: map[string]interface{}{
			"postback_id": item.ID,
			"attempts":    item.Attempts,
			"last_error":  item.LastError,
		},
	})
}

// calculateBackoff calculates exponential backoff with jitter
func (s *PostbackQueueService) calculateBackoff(attempt int) int {
	backoff := float64(s.baseRetryMs) * math.Pow(2, float64(attempt-1))
	
	if backoff > float64(s.maxRetryMs) {
		backoff = float64(s.maxRetryMs)
	}
	
	// Add jitter (±20%)
	jitter := backoff * 0.2 * (rand.Float64()*2 - 1)
	backoff += jitter
	
	return int(backoff)
}

// ============================================
// DLQ OPERATIONS
// ============================================

// GetDLQ returns all DLQ items
func (s *PostbackQueueService) GetDLQ() []*PostbackQueueItem {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	result := make([]*PostbackQueueItem, len(s.dlq))
	copy(result, s.dlq)
	return result
}

// RetryDLQItem retries a specific DLQ item
func (s *PostbackQueueService) RetryDLQItem(itemID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and remove from DLQ
	for i, item := range s.dlq {
		if item.ID == itemID {
			// Remove from DLQ
			s.dlq = append(s.dlq[:i], s.dlq[i+1:]...)
			
			// Reset attempts and add to pending
			item.Attempts = 0
			item.LastError = ""
			s.pendingQueue = append(s.pendingQueue, item)
			
			atomic.AddInt64(&s.totalDLQ, -1)
			
			// Remove from Redis DLQ
			ctx := context.Background()
			cache.Delete(ctx, fmt.Sprintf("postback_dlq:%s", itemID))
			
			return nil
		}
	}

	return fmt.Errorf("item not found in DLQ: %s", itemID)
}

// RetryAllDLQ retries all DLQ items
func (s *PostbackQueueService) RetryAllDLQ() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	count := len(s.dlq)
	
	for _, item := range s.dlq {
		item.Attempts = 0
		item.LastError = ""
		s.pendingQueue = append(s.pendingQueue, item)
		
		// Remove from Redis DLQ
		ctx := context.Background()
		cache.Delete(ctx, fmt.Sprintf("postback_dlq:%s", item.ID))
	}

	s.dlq = make([]*PostbackQueueItem, 0)
	atomic.StoreInt64(&s.totalDLQ, 0)

	return count
}

// DeleteDLQItem deletes a DLQ item permanently
func (s *PostbackQueueService) DeleteDLQItem(itemID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, item := range s.dlq {
		if item.ID == itemID {
			s.dlq = append(s.dlq[:i], s.dlq[i+1:]...)
			atomic.AddInt64(&s.totalDLQ, -1)
			
			// Remove from Redis
			ctx := context.Background()
			cache.Delete(ctx, fmt.Sprintf("postback_dlq:%s", itemID))
			
			return nil
		}
	}

	return fmt.Errorf("item not found in DLQ: %s", itemID)
}

// ============================================
// LIFECYCLE
// ============================================

// Start starts the postback queue workers
func (s *PostbackQueueService) Start() {
	s.isRunning = true

	// Process worker
	s.wg.Add(1)
	go s.processWorker()

	// Redis sync worker
	s.wg.Add(1)
	go s.redisSyncWorker()
}

// Stop stops the postback queue
func (s *PostbackQueueService) Stop() {
	s.isRunning = false
	close(s.stopChan)
	s.wg.Wait()
}

// processWorker continuously processes the queue
func (s *PostbackQueueService) processWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			// Drain remaining
			s.ProcessQueue()
			return
		case <-ticker.C:
			s.ProcessQueue()
		}
	}
}

// redisSyncWorker syncs with Redis
func (s *PostbackQueueService) redisSyncWorker() {
	defer s.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.syncFromRedis()
		}
	}
}

// syncFromRedis loads pending items from Redis
func (s *PostbackQueueService) syncFromRedis() {
	// In production, scan Redis for pending postbacks and add to local queue
}

// ============================================
// METRICS
// ============================================

// GetStats returns queue statistics
func (s *PostbackQueueService) GetStats() map[string]interface{} {
	s.mu.RLock()
	pendingCount := len(s.pendingQueue)
	dlqCount := len(s.dlq)
	s.mu.RUnlock()

	return map[string]interface{}{
		"pending_count":  pendingCount,
		"dlq_count":      dlqCount,
		"total_queued":   atomic.LoadInt64(&s.totalQueued),
		"total_sent":     atomic.LoadInt64(&s.totalSent),
		"total_failed":   atomic.LoadInt64(&s.totalFailed),
		"total_dlq":      atomic.LoadInt64(&s.totalDLQ),
		"is_running":     s.isRunning,
	}
}

// GetPendingCount returns the number of pending postbacks
func (s *PostbackQueueService) GetPendingCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.pendingQueue)
}

// ============================================
// GLOBAL INSTANCE
// ============================================

var (
	postbackQueueInstance *PostbackQueueService
	postbackQueueOnce     sync.Once
)

// GetPostbackQueueService returns the global postback queue service
func GetPostbackQueueService() *PostbackQueueService {
	postbackQueueOnce.Do(func() {
		config := DefaultPostbackQueueConfig()
		postbackQueueInstance = NewPostbackQueueService(config)
	})
	return postbackQueueInstance
}


package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/google/uuid"
)

// ============================================
// STREAM NAMES
// ============================================

const (
	StreamClicks      = "stream:clicks"
	StreamConversions = "stream:conversions"
	StreamPostbacks   = "stream:postbacks"
	StreamEdgeEvents  = "stream:edge_events"
)

// ============================================
// STREAM MESSAGE
// ============================================

// StreamMessage represents a message in a Redis stream
type StreamMessage struct {
	ID        string                 `json:"id"`
	StreamID  string                 `json:"stream_id"`
	Timestamp time.Time              `json:"timestamp"`
	Type      string                 `json:"type"`
	TenantID  string                 `json:"tenant_id"`
	Data      map[string]interface{} `json:"data"`
	Processed bool                   `json:"processed"`
}

// ============================================
// STREAM PRODUCER
// ============================================

// StreamProducer produces messages to Redis streams
type StreamProducer struct {
	observability *ObservabilityService
}

// NewStreamProducer creates a new stream producer
func NewStreamProducer() *StreamProducer {
	return &StreamProducer{
		observability: NewObservabilityService(),
	}
}

// Publish publishes a message to a stream
func (p *StreamProducer) Publish(ctx context.Context, stream string, msg *StreamMessage) error {
	if cache.RedisClient == nil {
		return fmt.Errorf("Redis not available")
	}

	// Set defaults
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now().UTC()
	}

	// Serialize data
	data, err := json.Marshal(msg.Data)
	if err != nil {
		return fmt.Errorf("failed to serialize message data: %w", err)
	}

	// Add to stream
	args := map[string]interface{}{
		"id":        msg.ID,
		"type":      msg.Type,
		"tenant_id": msg.TenantID,
		"timestamp": msg.Timestamp.Format(time.RFC3339Nano),
		"data":      string(data),
	}

	result := cache.RedisClient.XAdd(ctx, &cache.RedisXAddArgs{
		Stream: stream,
		Values: args,
	})

	if result.Err() != nil {
		return fmt.Errorf("failed to add to stream: %w", result.Err())
	}

	msg.StreamID = result.Val()
	return nil
}

// PublishClick publishes a click event
func (p *StreamProducer) PublishClick(ctx context.Context, tenantID string, data map[string]interface{}) error {
	msg := &StreamMessage{
		Type:     "click",
		TenantID: tenantID,
		Data:     data,
	}
	return p.Publish(ctx, StreamClicks, msg)
}

// PublishConversion publishes a conversion event
func (p *StreamProducer) PublishConversion(ctx context.Context, tenantID string, data map[string]interface{}) error {
	msg := &StreamMessage{
		Type:     "conversion",
		TenantID: tenantID,
		Data:     data,
	}
	return p.Publish(ctx, StreamConversions, msg)
}

// PublishPostback publishes a postback event
func (p *StreamProducer) PublishPostback(ctx context.Context, tenantID string, data map[string]interface{}) error {
	msg := &StreamMessage{
		Type:     "postback",
		TenantID: tenantID,
		Data:     data,
	}
	return p.Publish(ctx, StreamPostbacks, msg)
}

// PublishEdgeEvent publishes an edge event
func (p *StreamProducer) PublishEdgeEvent(ctx context.Context, tenantID string, data map[string]interface{}) error {
	msg := &StreamMessage{
		Type:     "edge_event",
		TenantID: tenantID,
		Data:     data,
	}
	return p.Publish(ctx, StreamEdgeEvents, msg)
}

// ============================================
// STREAM CONSUMER
// ============================================

// StreamConsumer consumes messages from Redis streams
type StreamConsumer struct {
	mu            sync.RWMutex
	groupName     string
	consumerName  string
	streams       []string
	handlers      map[string]StreamHandler
	observability *ObservabilityService
	walService    *WALService
	
	// State
	isRunning     bool
	stopChan      chan struct{}
	wg            sync.WaitGroup
	
	// Metrics
	totalConsumed int64
	totalAcked    int64
	totalFailed   int64
	streamLag     map[string]int64
}

// StreamHandler handles messages from a stream
type StreamHandler func(ctx context.Context, msg *StreamMessage) error

// StreamConsumerConfig holds consumer configuration
type StreamConsumerConfig struct {
	GroupName    string
	ConsumerName string
	Streams      []string
}

// DefaultStreamConsumerConfig returns default configuration
func DefaultStreamConsumerConfig() StreamConsumerConfig {
	return StreamConsumerConfig{
		GroupName:    "afftok-consumers",
		ConsumerName: fmt.Sprintf("consumer-%s", uuid.New().String()[:8]),
		Streams:      []string{StreamClicks, StreamConversions, StreamPostbacks, StreamEdgeEvents},
	}
}

// NewStreamConsumer creates a new stream consumer
func NewStreamConsumer(config StreamConsumerConfig) *StreamConsumer {
	return &StreamConsumer{
		groupName:     config.GroupName,
		consumerName:  config.ConsumerName,
		streams:       config.Streams,
		handlers:      make(map[string]StreamHandler),
		observability: NewObservabilityService(),
		walService:    GetWALService(),
		stopChan:      make(chan struct{}),
		streamLag:     make(map[string]int64),
	}
}

// RegisterHandler registers a handler for a stream
func (c *StreamConsumer) RegisterHandler(stream string, handler StreamHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.handlers[stream] = handler
}

// ============================================
// CONSUMER LIFECYCLE
// ============================================

// Start starts the stream consumer
func (c *StreamConsumer) Start() error {
	if cache.RedisClient == nil {
		return fmt.Errorf("Redis not available")
	}

	// Create consumer groups
	ctx := context.Background()
	for _, stream := range c.streams {
		// Create group (ignore error if already exists)
		cache.RedisClient.XGroupCreateMkStream(ctx, stream, c.groupName, "0")
	}

	c.isRunning = true

	// Start consumer workers
	for _, stream := range c.streams {
		c.wg.Add(1)
		go c.consumeStream(stream)
	}

	// Start lag monitor
	c.wg.Add(1)
	go c.monitorLag()

	// Start pending processor
	c.wg.Add(1)
	go c.processPending()

	return nil
}

// Stop stops the stream consumer
func (c *StreamConsumer) Stop() {
	c.isRunning = false
	close(c.stopChan)
	c.wg.Wait()
}

// ============================================
// STREAM CONSUMPTION
// ============================================

// consumeStream consumes messages from a single stream
func (c *StreamConsumer) consumeStream(stream string) {
	defer c.wg.Done()

	ctx := context.Background()

	for {
		select {
		case <-c.stopChan:
			return
		default:
		}

		// Read from stream
		result := cache.RedisClient.XReadGroup(ctx, &cache.RedisXReadGroupArgs{
			Group:    c.groupName,
			Consumer: c.consumerName,
			Streams:  []string{stream, ">"},
			Count:    100,
			Block:    time.Second,
		})

		if result.Err() != nil {
			if result.Err().Error() != "redis: nil" {
				c.observability.Log(LogEvent{
					Category: LogCategoryErrorEvent,
					Level:    LogLevelError,
					Message:  "Stream read error",
					Metadata: map[string]interface{}{
						"stream": stream,
						"error":  result.Err().Error(),
					},
				})
			}
			continue
		}

		// Process messages
		for _, streamResult := range result.Val() {
			for _, message := range streamResult.Messages {
				c.processMessage(ctx, stream, message)
			}
		}
	}
}

// processMessage processes a single message
func (c *StreamConsumer) processMessage(ctx context.Context, stream string, message cache.RedisXMessage) {
	atomic.AddInt64(&c.totalConsumed, 1)

	// Parse message
	msg, err := c.parseMessage(message)
	if err != nil {
		c.observability.Log(LogEvent{
			Category: LogCategoryErrorEvent,
			Level:    LogLevelError,
			Message:  "Failed to parse stream message",
			Metadata: map[string]interface{}{
				"stream":    stream,
				"message_id": message.ID,
				"error":     err.Error(),
			},
		})
		// Still ack to avoid infinite loop
		cache.RedisClient.XAck(ctx, stream, c.groupName, message.ID)
		return
	}

	// Write to WAL before processing
	if c.walService != nil {
		walType := WALEventType(msg.Type)
		c.walService.Append(walType, msg.TenantID, msg.Data)
	}

	// Get handler
	c.mu.RLock()
	handler, exists := c.handlers[stream]
	c.mu.RUnlock()

	if !exists {
		// No handler, just ack
		cache.RedisClient.XAck(ctx, stream, c.groupName, message.ID)
		atomic.AddInt64(&c.totalAcked, 1)
		return
	}

	// Process with handler
	if err := handler(ctx, msg); err != nil {
		atomic.AddInt64(&c.totalFailed, 1)
		c.observability.Log(LogEvent{
			Category: LogCategoryErrorEvent,
			Level:    LogLevelError,
			Message:  "Stream handler error",
			Metadata: map[string]interface{}{
				"stream":    stream,
				"message_id": message.ID,
				"error":     err.Error(),
			},
		})
		// Don't ack - will be retried
		return
	}

	// Ack message
	cache.RedisClient.XAck(ctx, stream, c.groupName, message.ID)
	atomic.AddInt64(&c.totalAcked, 1)

	// Mark WAL entry as processed
	if c.walService != nil {
		c.walService.MarkProcessed(msg.ID)
	}
}

// parseMessage parses a Redis stream message
func (c *StreamConsumer) parseMessage(message cache.RedisXMessage) (*StreamMessage, error) {
	msg := &StreamMessage{
		StreamID: message.ID,
	}

	if id, ok := message.Values["id"].(string); ok {
		msg.ID = id
	}
	if msgType, ok := message.Values["type"].(string); ok {
		msg.Type = msgType
	}
	if tenantID, ok := message.Values["tenant_id"].(string); ok {
		msg.TenantID = tenantID
	}
	if ts, ok := message.Values["timestamp"].(string); ok {
		if t, err := time.Parse(time.RFC3339Nano, ts); err == nil {
			msg.Timestamp = t
		}
	}
	if data, ok := message.Values["data"].(string); ok {
		if err := json.Unmarshal([]byte(data), &msg.Data); err != nil {
			return nil, fmt.Errorf("failed to parse data: %w", err)
		}
	}

	return msg, nil
}

// ============================================
// PENDING PROCESSING
// ============================================

// processPending processes pending (unacked) messages
func (c *StreamConsumer) processPending() {
	defer c.wg.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.claimPending()
		}
	}
}

// claimPending claims and reprocesses pending messages
func (c *StreamConsumer) claimPending() {
	ctx := context.Background()

	for _, stream := range c.streams {
		// Get pending messages older than 1 minute
		result := cache.RedisClient.XPendingExt(ctx, &cache.RedisXPendingExtArgs{
			Stream: stream,
			Group:  c.groupName,
			Start:  "-",
			End:    "+",
			Count:  100,
		})

		if result.Err() != nil {
			continue
		}

		for _, pending := range result.Val() {
			// Only claim messages idle for more than 1 minute
			if pending.Idle < time.Minute {
				continue
			}

			// Claim the message
			claimResult := cache.RedisClient.XClaim(ctx, &cache.RedisXClaimArgs{
				Stream:   stream,
				Group:    c.groupName,
				Consumer: c.consumerName,
				MinIdle:  time.Minute,
				Messages: []string{pending.ID},
			})

			if claimResult.Err() != nil {
				continue
			}

			// Reprocess claimed messages
			for _, message := range claimResult.Val() {
				c.processMessage(ctx, stream, message)
			}
		}
	}
}

// ============================================
// LAG MONITORING
// ============================================

// monitorLag monitors stream lag
func (c *StreamConsumer) monitorLag() {
	defer c.wg.Done()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.stopChan:
			return
		case <-ticker.C:
			c.updateLag()
		}
	}
}

// updateLag updates stream lag metrics
func (c *StreamConsumer) updateLag() {
	ctx := context.Background()

	for _, stream := range c.streams {
		// Get stream length
		lenResult := cache.RedisClient.XLen(ctx, stream)
		if lenResult.Err() != nil {
			continue
		}

		// Get pending count
		pendingResult := cache.RedisClient.XPending(ctx, stream, c.groupName)
		if pendingResult.Err() != nil {
			continue
		}

		c.mu.Lock()
		c.streamLag[stream] = pendingResult.Val().Count
		c.mu.Unlock()
	}
}

// ============================================
// METRICS
// ============================================

// GetStats returns consumer statistics
func (c *StreamConsumer) GetStats() map[string]interface{} {
	c.mu.RLock()
	lagCopy := make(map[string]int64)
	for k, v := range c.streamLag {
		lagCopy[k] = v
	}
	c.mu.RUnlock()

	return map[string]interface{}{
		"group_name":     c.groupName,
		"consumer_name":  c.consumerName,
		"streams":        c.streams,
		"total_consumed": atomic.LoadInt64(&c.totalConsumed),
		"total_acked":    atomic.LoadInt64(&c.totalAcked),
		"total_failed":   atomic.LoadInt64(&c.totalFailed),
		"stream_lag":     lagCopy,
		"is_running":     c.isRunning,
	}
}

// GetStreamLag returns the lag for a specific stream
func (c *StreamConsumer) GetStreamLag(stream string) int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.streamLag[stream]
}

// ============================================
// GLOBAL INSTANCES
// ============================================

var (
	streamProducerInstance *StreamProducer
	streamProducerOnce     sync.Once
	streamConsumerInstance *StreamConsumer
	streamConsumerOnce     sync.Once
)

// GetStreamProducer returns the global stream producer
func GetStreamProducer() *StreamProducer {
	streamProducerOnce.Do(func() {
		streamProducerInstance = NewStreamProducer()
	})
	return streamProducerInstance
}

// GetStreamConsumer returns the global stream consumer
func GetStreamConsumer() *StreamConsumer {
	streamConsumerOnce.Do(func() {
		config := DefaultStreamConsumerConfig()
		streamConsumerInstance = NewStreamConsumer(config)
	})
	return streamConsumerInstance
}


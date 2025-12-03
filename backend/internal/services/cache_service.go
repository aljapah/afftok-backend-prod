package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
)

// ============================================
// REDIS KEY NAMESPACES
// ============================================

const (
	// Tracking namespace - for click/conversion tracking
	NSTracking = "tracking:"
	
	// Stats namespace - for aggregated statistics
	NSStats = "stats:"
	
	// Offer namespace - for offer data caching
	NSOffer = "offer:"
	
	// User namespace - for user data caching
	NSUser = "user:"
	
	// Fraud namespace - for fraud detection data
	NSFraud = "fraud:"
	
	// Logs namespace - for log storage
	NSLogs = "logs:"
	
	// Rate limit namespace
	NSRateLimit = "ratelimit:"
	
	// Session namespace
	NSSession = "session:"
	
	// Counter namespace - for atomic counters
	NSCounter = "counter:"
)

// ============================================
// CACHE SERVICE
// ============================================

// CacheService provides high-performance caching with namespaces
type CacheService struct {
	localCache   sync.Map // In-memory L1 cache
	localTTL     time.Duration
	mutex        sync.RWMutex
	
	// Write buffer for batching
	writeBuffer  chan writeOp
	bufferSize   int
	flushInterval time.Duration
	
	ctx          context.Context
	cancel       context.CancelFunc
}

type writeOp struct {
	key   string
	value interface{}
	ttl   time.Duration
}

type localCacheEntry struct {
	value     interface{}
	expiresAt time.Time
}

var (
	cacheServiceInstance *CacheService
	cacheServiceOnce     sync.Once
)

// NewCacheService creates a singleton CacheService
func NewCacheService() *CacheService {
	cacheServiceOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		
		cacheServiceInstance = &CacheService{
			localTTL:      30 * time.Second,
			writeBuffer:   make(chan writeOp, 10000),
			bufferSize:    100,
			flushInterval: 100 * time.Millisecond,
			ctx:           ctx,
			cancel:        cancel,
		}
		
		// Start background flush worker
		go cacheServiceInstance.flushWorker()
		
		// Start local cache cleanup
		go cacheServiceInstance.cleanupWorker()
	})
	return cacheServiceInstance
}

// ============================================
// BASIC OPERATIONS
// ============================================

// Get retrieves a value from cache (L1 -> L2)
func (c *CacheService) Get(key string) (string, error) {
	// Try L1 (local) cache first
	if entry, ok := c.localCache.Load(key); ok {
		if e, ok := entry.(*localCacheEntry); ok {
			if time.Now().Before(e.expiresAt) {
				if str, ok := e.value.(string); ok {
					return str, nil
				}
			}
			// Expired, remove from L1
			c.localCache.Delete(key)
		}
	}

	// Try L2 (Redis) cache
	if cache.RedisClient == nil {
		return "", fmt.Errorf("redis not available")
	}

	ctx, cancel := context.WithTimeout(c.ctx, 100*time.Millisecond)
	defer cancel()

	result, err := cache.RedisClient.Get(ctx, key).Result()
	if err != nil {
		return "", err
	}

	// Store in L1 for future reads
	c.setLocal(key, result, c.localTTL)

	return result, nil
}

// Set stores a value in cache
func (c *CacheService) Set(key string, value interface{}, ttl time.Duration) error {
	// Set in L1 immediately
	c.setLocal(key, value, ttl)

	// Queue for L2 (Redis) write
	select {
	case c.writeBuffer <- writeOp{key: key, value: value, ttl: ttl}:
		return nil
	default:
		// Buffer full, write directly
		return c.setRedis(key, value, ttl)
	}
}

// SetImmediate stores a value immediately without buffering
func (c *CacheService) SetImmediate(key string, value interface{}, ttl time.Duration) error {
	c.setLocal(key, value, ttl)
	return c.setRedis(key, value, ttl)
}

// Delete removes a value from cache
func (c *CacheService) Delete(key string) error {
	c.localCache.Delete(key)
	
	if cache.RedisClient == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(c.ctx, 100*time.Millisecond)
	defer cancel()

	return cache.RedisClient.Del(ctx, key).Err()
}

// ============================================
// ATOMIC OPERATIONS
// ============================================

// Increment atomically increments a counter
func (c *CacheService) Increment(key string) (int64, error) {
	if cache.RedisClient == nil {
		return 0, fmt.Errorf("redis not available")
	}

	ctx, cancel := context.WithTimeout(c.ctx, 100*time.Millisecond)
	defer cancel()

	return cache.RedisClient.Incr(ctx, key).Result()
}

// IncrementBy atomically increments a counter by a value
func (c *CacheService) IncrementBy(key string, value int64) (int64, error) {
	if cache.RedisClient == nil {
		return 0, fmt.Errorf("redis not available")
	}

	ctx, cancel := context.WithTimeout(c.ctx, 100*time.Millisecond)
	defer cancel()

	return cache.RedisClient.IncrBy(ctx, key, value).Result()
}

// IncrementFloat atomically increments a float counter
func (c *CacheService) IncrementFloat(key string, value float64) (float64, error) {
	if cache.RedisClient == nil {
		return 0, fmt.Errorf("redis not available")
	}

	ctx, cancel := context.WithTimeout(c.ctx, 100*time.Millisecond)
	defer cancel()

	return cache.RedisClient.IncrByFloat(ctx, key, value).Result()
}

// SetNX sets a value only if it doesn't exist (for locking)
func (c *CacheService) SetNX(key string, value interface{}, ttl time.Duration) (bool, error) {
	if cache.RedisClient == nil {
		return false, fmt.Errorf("redis not available")
	}

	ctx, cancel := context.WithTimeout(c.ctx, 100*time.Millisecond)
	defer cancel()

	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	default:
		bytes, _ := json.Marshal(v)
		strValue = string(bytes)
	}

	return cache.RedisClient.SetNX(ctx, key, strValue, ttl).Result()
}

// ============================================
// NAMESPACED OPERATIONS
// ============================================

// GetTracking gets a tracking value
func (c *CacheService) GetTracking(key string) (string, error) {
	return c.Get(NSTracking + key)
}

// SetTracking sets a tracking value
func (c *CacheService) SetTracking(key string, value interface{}, ttl time.Duration) error {
	return c.Set(NSTracking+key, value, ttl)
}

// GetStats gets a stats value
func (c *CacheService) GetStats(key string) (string, error) {
	return c.Get(NSStats + key)
}

// SetStats sets a stats value
func (c *CacheService) SetStats(key string, value interface{}, ttl time.Duration) error {
	return c.Set(NSStats+key, value, ttl)
}

// IncrementCounter increments a namespaced counter
func (c *CacheService) IncrementCounter(key string) (int64, error) {
	return c.Increment(NSCounter + key)
}

// GetOffer gets cached offer data
func (c *CacheService) GetOffer(offerID string) (string, error) {
	return c.Get(NSOffer + offerID)
}

// SetOffer caches offer data
func (c *CacheService) SetOffer(offerID string, value interface{}, ttl time.Duration) error {
	return c.Set(NSOffer+offerID, value, ttl)
}

// GetUser gets cached user data
func (c *CacheService) GetUser(userID string) (string, error) {
	return c.Get(NSUser + userID)
}

// SetUser caches user data
func (c *CacheService) SetUser(userID string, value interface{}, ttl time.Duration) error {
	return c.Set(NSUser+userID, value, ttl)
}

// ============================================
// BATCH OPERATIONS
// ============================================

// MGet gets multiple values
func (c *CacheService) MGet(keys []string) ([]interface{}, error) {
	if cache.RedisClient == nil {
		return nil, fmt.Errorf("redis not available")
	}

	ctx, cancel := context.WithTimeout(c.ctx, 200*time.Millisecond)
	defer cancel()

	return cache.RedisClient.MGet(ctx, keys...).Result()
}

// MSet sets multiple values
func (c *CacheService) MSet(pairs map[string]interface{}) error {
	if cache.RedisClient == nil {
		return fmt.Errorf("redis not available")
	}

	ctx, cancel := context.WithTimeout(c.ctx, 200*time.Millisecond)
	defer cancel()

	// Convert to slice for MSet
	args := make([]interface{}, 0, len(pairs)*2)
	for k, v := range pairs {
		args = append(args, k, v)
		c.setLocal(k, v, c.localTTL)
	}

	return cache.RedisClient.MSet(ctx, args...).Err()
}

// ============================================
// LIST OPERATIONS (for logs)
// ============================================

// LPush pushes to a list
func (c *CacheService) LPush(key string, values ...interface{}) error {
	if cache.RedisClient == nil {
		return fmt.Errorf("redis not available")
	}

	ctx, cancel := context.WithTimeout(c.ctx, 100*time.Millisecond)
	defer cancel()

	return cache.RedisClient.LPush(ctx, key, values...).Err()
}

// LRange gets a range from a list
func (c *CacheService) LRange(key string, start, stop int64) ([]string, error) {
	if cache.RedisClient == nil {
		return nil, fmt.Errorf("redis not available")
	}

	ctx, cancel := context.WithTimeout(c.ctx, 200*time.Millisecond)
	defer cancel()

	return cache.RedisClient.LRange(ctx, key, start, stop).Result()
}

// LTrim trims a list
func (c *CacheService) LTrim(key string, start, stop int64) error {
	if cache.RedisClient == nil {
		return fmt.Errorf("redis not available")
	}

	ctx, cancel := context.WithTimeout(c.ctx, 100*time.Millisecond)
	defer cancel()

	return cache.RedisClient.LTrim(ctx, key, start, stop).Err()
}

// ============================================
// PIPELINE OPERATIONS
// ============================================

// Pipeline executes multiple commands in a pipeline
func (c *CacheService) Pipeline(fn func(pipe cache.Pipeliner) error) error {
	if cache.RedisClient == nil {
		return fmt.Errorf("redis not available")
	}

	ctx, cancel := context.WithTimeout(c.ctx, 500*time.Millisecond)
	defer cancel()

	pipe := cache.RedisClient.Pipeline()
	if err := fn(pipe); err != nil {
		return err
	}

	_, err := pipe.Exec(ctx)
	return err
}

// ============================================
// INTERNAL METHODS
// ============================================

func (c *CacheService) setLocal(key string, value interface{}, ttl time.Duration) {
	c.localCache.Store(key, &localCacheEntry{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	})
}

func (c *CacheService) setRedis(key string, value interface{}, ttl time.Duration) error {
	if cache.RedisClient == nil {
		return nil
	}

	ctx, cancel := context.WithTimeout(c.ctx, 100*time.Millisecond)
	defer cancel()

	var strValue string
	switch v := value.(type) {
	case string:
		strValue = v
	case []byte:
		strValue = string(v)
	default:
		bytes, err := json.Marshal(v)
		if err != nil {
			return err
		}
		strValue = string(bytes)
	}

	return cache.RedisClient.Set(ctx, key, strValue, ttl).Err()
}

// flushWorker batches writes to Redis
func (c *CacheService) flushWorker() {
	ticker := time.NewTicker(c.flushInterval)
	defer ticker.Stop()

	batch := make([]writeOp, 0, c.bufferSize)

	for {
		select {
		case <-c.ctx.Done():
			// Flush remaining
			if len(batch) > 0 {
				c.flushBatch(batch)
			}
			return
		case op := <-c.writeBuffer:
			batch = append(batch, op)
			if len(batch) >= c.bufferSize {
				c.flushBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				c.flushBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

func (c *CacheService) flushBatch(batch []writeOp) {
	if cache.RedisClient == nil || len(batch) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(c.ctx, 500*time.Millisecond)
	defer cancel()

	pipe := cache.RedisClient.Pipeline()
	
	for _, op := range batch {
		var strValue string
		switch v := op.value.(type) {
		case string:
			strValue = v
		default:
			bytes, _ := json.Marshal(v)
			strValue = string(bytes)
		}
		pipe.Set(ctx, op.key, strValue, op.ttl)
	}

	pipe.Exec(ctx)
}

// cleanupWorker periodically cleans expired L1 cache entries
func (c *CacheService) cleanupWorker() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			now := time.Now()
			c.localCache.Range(func(key, value interface{}) bool {
				if entry, ok := value.(*localCacheEntry); ok {
					if now.After(entry.expiresAt) {
						c.localCache.Delete(key)
					}
				}
				return true
			})
		}
	}
}

// Stop stops the cache service
func (c *CacheService) Stop() {
	c.cancel()
}

// GetStats returns cache statistics
func (c *CacheService) GetCacheStats() map[string]interface{} {
	localCount := 0
	c.localCache.Range(func(key, value interface{}) bool {
		localCount++
		return true
	})

	return map[string]interface{}{
		"local_cache_size": localCount,
		"write_buffer_len": len(c.writeBuffer),
		"local_ttl_seconds": c.localTTL.Seconds(),
	}
}


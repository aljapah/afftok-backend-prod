package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/config"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

// Pipeliner is an alias for redis.Pipeliner
type Pipeliner = redis.Pipeliner

// RedisZ is an alias for redis.Z (sorted set member)
type RedisZ = redis.Z

// RedisXAddArgs is an alias for redis.XAddArgs
type RedisXAddArgs = redis.XAddArgs

// RedisXReadGroupArgs is an alias for redis.XReadGroupArgs
type RedisXReadGroupArgs = redis.XReadGroupArgs

// RedisXMessage is an alias for redis.XMessage
type RedisXMessage = redis.XMessage

// RedisXPendingExtArgs is an alias for redis.XPendingExtArgs
type RedisXPendingExtArgs = redis.XPendingExtArgs

// RedisXClaimArgs is an alias for redis.XClaimArgs
type RedisXClaimArgs = redis.XClaimArgs

// RedisConfig holds Redis connection pool configuration
type RedisConfig struct {
	PoolSize        int
	MinIdleConns    int
	MaxIdleConns    int
	ConnMaxIdleTime time.Duration
	ConnMaxLifetime time.Duration
	PoolTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

// DefaultRedisConfig returns optimized Redis configuration
// NOTE: For Redis Cloud free tier (30MB), max connections is ~30
// For production, increase these values based on your Redis plan
func DefaultRedisConfig() RedisConfig {
	// Use conservative values for free tier Redis
	// Upgrade these for production Redis plans
	poolSize := 10        // Max 10 connections (free tier safe)
	minIdle := 2          // Keep 2 connections ready
	maxIdle := 5          // Max 5 idle connections
	
	return RedisConfig{
		PoolSize:        poolSize,
		MinIdleConns:    minIdle,
		MaxIdleConns:    maxIdle,
		ConnMaxIdleTime: 5 * time.Minute,    // Close idle connections after 5 min
		ConnMaxLifetime: 30 * time.Minute,   // Recycle connections every 30 min
		PoolTimeout:     4 * time.Second,    // Wait for connection from pool
		ReadTimeout:     3 * time.Second,    // Read timeout
		WriteTimeout:    3 * time.Second,    // Write timeout
	}
}

func ConnectRedis(cfg *config.Config) (*redis.Client, error) {
	redisURL := cfg.RedisURL
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is not configured")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Apply high-performance configuration
	redisCfg := DefaultRedisConfig()
	opt.PoolSize = redisCfg.PoolSize
	opt.MinIdleConns = redisCfg.MinIdleConns
	opt.MaxIdleConns = redisCfg.MaxIdleConns
	opt.ConnMaxIdleTime = redisCfg.ConnMaxIdleTime
	opt.ConnMaxLifetime = redisCfg.ConnMaxLifetime
	opt.PoolTimeout = redisCfg.PoolTimeout
	opt.ReadTimeout = redisCfg.ReadTimeout
	opt.WriteTimeout = redisCfg.WriteTimeout

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	RedisClient = client
	log.Printf("✅ Redis connected successfully (pool_size=%d, min_idle=%d)", 
		redisCfg.PoolSize, redisCfg.MinIdleConns)
	return client, nil
}

func CloseRedis(client *redis.Client) error {
	if client == nil {
		return nil
	}
	return client.Close()
}

func GetRedisClient() *redis.Client {
	return RedisClient
}

func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.Set(ctx, key, value, expiration).Err()
}

func Get(ctx context.Context, key string) (string, error) {
	if RedisClient == nil {
		return "", fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.Get(ctx, key).Result()
}

func Delete(ctx context.Context, keys ...string) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.Del(ctx, keys...).Err()
}

func Exists(ctx context.Context, keys ...string) (int64, error) {
	if RedisClient == nil {
		return 0, fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.Exists(ctx, keys...).Result()
}

func Increment(ctx context.Context, key string) (int64, error) {
	if RedisClient == nil {
		return 0, fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.Incr(ctx, key).Result()
}

func Decrement(ctx context.Context, key string) (int64, error) {
	if RedisClient == nil {
		return 0, fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.Decr(ctx, key).Result()
}

// Expire sets expiration on a key
func Expire(ctx context.Context, key string, expiration time.Duration) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.Expire(ctx, key, expiration).Err()
}

// IncrWithExpire increments and sets expiration atomically
func IncrWithExpire(ctx context.Context, key string, expiration time.Duration) (int64, error) {
	if RedisClient == nil {
		return 0, fmt.Errorf("Redis client not initialized")
	}
	pipe := RedisClient.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, expiration)
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	return incr.Val(), nil
}

// ============================================
// BATCH OPERATIONS
// ============================================

// MGet gets multiple keys at once
func MGet(ctx context.Context, keys ...string) ([]interface{}, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.MGet(ctx, keys...).Result()
}

// MSet sets multiple key-value pairs at once
func MSet(ctx context.Context, values ...interface{}) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.MSet(ctx, values...).Err()
}

// Pipeline returns a new pipeline for batching commands
func Pipeline() redis.Pipeliner {
	if RedisClient == nil {
		return nil
	}
	return RedisClient.Pipeline()
}

// TxPipeline returns a transactional pipeline
func TxPipeline() redis.Pipeliner {
	if RedisClient == nil {
		return nil
	}
	return RedisClient.TxPipeline()
}

// ============================================
// ATOMIC OPERATIONS
// ============================================

// IncrBy increments by a specific value
func IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	if RedisClient == nil {
		return 0, fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.IncrBy(ctx, key, value).Result()
}

// SetNX sets only if key doesn't exist (for distributed locks)
func SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if RedisClient == nil {
		return false, fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.SetNX(ctx, key, value, expiration).Result()
}

// GetSet atomically sets a new value and returns the old value
func GetSet(ctx context.Context, key string, value interface{}) (string, error) {
	if RedisClient == nil {
		return "", fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.GetSet(ctx, key, value).Result()
}

// ============================================
// LIST OPERATIONS (for queues)
// ============================================

// LPush pushes to the left of a list
func LPush(ctx context.Context, key string, values ...interface{}) (int64, error) {
	if RedisClient == nil {
		return 0, fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.LPush(ctx, key, values...).Result()
}

// RPop pops from the right of a list
func RPop(ctx context.Context, key string) (string, error) {
	if RedisClient == nil {
		return "", fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.RPop(ctx, key).Result()
}

// BRPop blocking pop from the right of a list
func BRPop(ctx context.Context, timeout time.Duration, keys ...string) ([]string, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.BRPop(ctx, timeout, keys...).Result()
}

// LLen returns the length of a list
func LLen(ctx context.Context, key string) (int64, error) {
	if RedisClient == nil {
		return 0, fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.LLen(ctx, key).Result()
}

// LRange returns a range of elements from a list
func LRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	if RedisClient == nil {
		return nil, fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.LRange(ctx, key, start, stop).Result()
}

// LTrim trims a list to the specified range
func LTrim(ctx context.Context, key string, start, stop int64) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}
	return RedisClient.LTrim(ctx, key, start, stop).Err()
}

// ============================================
// POOL STATS
// ============================================

// GetPoolStats returns Redis connection pool statistics
func GetPoolStats() *redis.PoolStats {
	if RedisClient == nil {
		return nil
	}
	return RedisClient.PoolStats()
}

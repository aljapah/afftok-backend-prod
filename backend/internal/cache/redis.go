package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/afftok/backend/internal/config"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client

func ConnectRedis(cfg *config.Config) (*redis.Client, error) {
	redisURL := cfg.RedisURL
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is not configured")
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	RedisClient = client
	log.Println("✅ Redis connected successfully")
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

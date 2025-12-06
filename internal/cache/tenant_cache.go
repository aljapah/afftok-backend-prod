package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ============================================
// TENANT-SCOPED CACHE
// ============================================

// TenantCache provides tenant-scoped caching operations
type TenantCache struct {
	tenantID uuid.UUID
	prefix   string
}

// NewTenantCache creates a new tenant-scoped cache
func NewTenantCache(tenantID uuid.UUID) *TenantCache {
	return &TenantCache{
		tenantID: tenantID,
		prefix:   fmt.Sprintf("tenant:%s:", tenantID.String()),
	}
}

// ============================================
// KEY NAMESPACES
// ============================================

// Namespace constants for tenant-scoped keys
const (
	NSStats      = "stats:"
	NSOffers     = "offers:"
	NSUsers      = "users:"
	NSClicks     = "clicks:"
	NSFraud      = "fraud:"
	NSRateLimit  = "ratelimit:"
	NSGeoRules   = "geo:"
	NSWebhooks   = "webhooks:"
	NSAPIKeys    = "apikeys:"
	NSSettings   = "settings:"
	NSAnalytics  = "analytics:"
)

// ============================================
// KEY BUILDERS
// ============================================

// Key builds a tenant-scoped key
func (c *TenantCache) Key(namespace, key string) string {
	return c.prefix + namespace + key
}

// StatsKey builds a stats key
func (c *TenantCache) StatsKey(key string) string {
	return c.Key(NSStats, key)
}

// OffersKey builds an offers key
func (c *TenantCache) OffersKey(key string) string {
	return c.Key(NSOffers, key)
}

// UsersKey builds a users key
func (c *TenantCache) UsersKey(key string) string {
	return c.Key(NSUsers, key)
}

// ClicksKey builds a clicks key
func (c *TenantCache) ClicksKey(key string) string {
	return c.Key(NSClicks, key)
}

// FraudKey builds a fraud key
func (c *TenantCache) FraudKey(key string) string {
	return c.Key(NSFraud, key)
}

// RateLimitKey builds a rate limit key
func (c *TenantCache) RateLimitKey(key string) string {
	return c.Key(NSRateLimit, key)
}

// GeoRulesKey builds a geo rules key
func (c *TenantCache) GeoRulesKey(key string) string {
	return c.Key(NSGeoRules, key)
}

// WebhooksKey builds a webhooks key
func (c *TenantCache) WebhooksKey(key string) string {
	return c.Key(NSWebhooks, key)
}

// APIKeysKey builds an API keys key
func (c *TenantCache) APIKeysKey(key string) string {
	return c.Key(NSAPIKeys, key)
}

// SettingsKey builds a settings key
func (c *TenantCache) SettingsKey(key string) string {
	return c.Key(NSSettings, key)
}

// AnalyticsKey builds an analytics key
func (c *TenantCache) AnalyticsKey(key string) string {
	return c.Key(NSAnalytics, key)
}

// ============================================
// BASIC OPERATIONS
// ============================================

// Set sets a value with tenant scope
func (c *TenantCache) Set(ctx context.Context, namespace, key string, value interface{}, expiration time.Duration) error {
	fullKey := c.Key(namespace, key)
	
	var data string
	switch v := value.(type) {
	case string:
		data = v
	case []byte:
		data = string(v)
	default:
		bytes, err := json.Marshal(value)
		if err != nil {
			return err
		}
		data = string(bytes)
	}
	
	return Set(ctx, fullKey, data, expiration)
}

// Get gets a value with tenant scope
func (c *TenantCache) Get(ctx context.Context, namespace, key string) (string, error) {
	fullKey := c.Key(namespace, key)
	return Get(ctx, fullKey)
}

// GetJSON gets a JSON value and unmarshals it
func (c *TenantCache) GetJSON(ctx context.Context, namespace, key string, dest interface{}) error {
	data, err := c.Get(ctx, namespace, key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// Delete deletes a value with tenant scope
func (c *TenantCache) Delete(ctx context.Context, namespace, key string) error {
	fullKey := c.Key(namespace, key)
	return Delete(ctx, fullKey)
}

// Exists checks if a key exists
func (c *TenantCache) Exists(ctx context.Context, namespace, key string) (int64, error) {
	fullKey := c.Key(namespace, key)
	return Exists(ctx, fullKey)
}

// ============================================
// COUNTER OPERATIONS
// ============================================

// Increment increments a counter
func (c *TenantCache) Increment(ctx context.Context, namespace, key string) (int64, error) {
	fullKey := c.Key(namespace, key)
	return Increment(ctx, fullKey)
}

// IncrementBy increments a counter by a specific value
func (c *TenantCache) IncrementBy(ctx context.Context, namespace, key string, value int64) (int64, error) {
	fullKey := c.Key(namespace, key)
	if RedisClient == nil {
		return 0, fmt.Errorf("Redis not available")
	}
	return RedisClient.IncrBy(ctx, fullKey, value).Result()
}

// Decrement decrements a counter
func (c *TenantCache) Decrement(ctx context.Context, namespace, key string) (int64, error) {
	fullKey := c.Key(namespace, key)
	if RedisClient == nil {
		return 0, fmt.Errorf("Redis not available")
	}
	return RedisClient.Decr(ctx, fullKey).Result()
}

// GetCounter gets a counter value
func (c *TenantCache) GetCounter(ctx context.Context, namespace, key string) (int64, error) {
	data, err := c.Get(ctx, namespace, key)
	if err != nil {
		return 0, err
	}
	
	var count int64
	fmt.Sscanf(data, "%d", &count)
	return count, nil
}

// ============================================
// STATS OPERATIONS
// ============================================

// IncrementClicks increments click count
func (c *TenantCache) IncrementClicks(ctx context.Context) (int64, error) {
	return c.Increment(ctx, NSStats, "clicks:total")
}

// IncrementDailyClicks increments daily click count
func (c *TenantCache) IncrementDailyClicks(ctx context.Context, date string) (int64, error) {
	return c.Increment(ctx, NSStats, "clicks:daily:"+date)
}

// IncrementConversions increments conversion count
func (c *TenantCache) IncrementConversions(ctx context.Context) (int64, error) {
	return c.Increment(ctx, NSStats, "conversions:total")
}

// IncrementEarnings increments earnings
func (c *TenantCache) IncrementEarnings(ctx context.Context, amount int64) (int64, error) {
	return c.IncrementBy(ctx, NSStats, "earnings:total", amount)
}

// GetTotalClicks gets total clicks
func (c *TenantCache) GetTotalClicks(ctx context.Context) (int64, error) {
	return c.GetCounter(ctx, NSStats, "clicks:total")
}

// GetTotalConversions gets total conversions
func (c *TenantCache) GetTotalConversions(ctx context.Context) (int64, error) {
	return c.GetCounter(ctx, NSStats, "conversions:total")
}

// GetTotalEarnings gets total earnings
func (c *TenantCache) GetTotalEarnings(ctx context.Context) (int64, error) {
	return c.GetCounter(ctx, NSStats, "earnings:total")
}

// ============================================
// FRAUD OPERATIONS
// ============================================

// IncrementFraudBlocked increments fraud blocked count
func (c *TenantCache) IncrementFraudBlocked(ctx context.Context) (int64, error) {
	return c.Increment(ctx, NSFraud, "blocked:total")
}

// AddRiskyIP adds an IP to risky list
func (c *TenantCache) AddRiskyIP(ctx context.Context, ip string, score int) error {
	return c.Set(ctx, NSFraud, "risky_ip:"+ip, score, 24*time.Hour)
}

// IsRiskyIP checks if an IP is risky
func (c *TenantCache) IsRiskyIP(ctx context.Context, ip string) (bool, error) {
	exists, err := c.Exists(ctx, NSFraud, "risky_ip:"+ip)
	return exists > 0, err
}

// ============================================
// RATE LIMIT OPERATIONS
// ============================================

// CheckRateLimit checks and increments rate limit
func (c *TenantCache) CheckRateLimit(ctx context.Context, key string, limit int64, window time.Duration) (bool, int64, error) {
	fullKey := c.RateLimitKey(key)
	
	count, err := Increment(ctx, fullKey)
	if err != nil {
		return true, 0, err // Allow on error
	}
	
	if count == 1 {
		// First request, set expiry
		Expire(ctx, fullKey, window)
	}
	
	return count <= limit, count, nil
}

// ============================================
// OFFER CACHE OPERATIONS
// ============================================

// CacheOffer caches an offer
func (c *TenantCache) CacheOffer(ctx context.Context, offerID string, data interface{}) error {
	return c.Set(ctx, NSOffers, offerID, data, 5*time.Minute)
}

// GetCachedOffer gets a cached offer
func (c *TenantCache) GetCachedOffer(ctx context.Context, offerID string, dest interface{}) error {
	return c.GetJSON(ctx, NSOffers, offerID, dest)
}

// InvalidateOffer invalidates an offer cache
func (c *TenantCache) InvalidateOffer(ctx context.Context, offerID string) error {
	return c.Delete(ctx, NSOffers, offerID)
}

// ============================================
// USER CACHE OPERATIONS
// ============================================

// CacheUserStats caches user stats
func (c *TenantCache) CacheUserStats(ctx context.Context, userID string, stats interface{}) error {
	return c.Set(ctx, NSUsers, "stats:"+userID, stats, 5*time.Minute)
}

// GetCachedUserStats gets cached user stats
func (c *TenantCache) GetCachedUserStats(ctx context.Context, userID string, dest interface{}) error {
	return c.GetJSON(ctx, NSUsers, "stats:"+userID, dest)
}

// ============================================
// BULK OPERATIONS
// ============================================

// DeleteByPattern deletes all keys matching a pattern
func (c *TenantCache) DeleteByPattern(ctx context.Context, namespace, pattern string) error {
	if RedisClient == nil {
		return fmt.Errorf("Redis not available")
	}
	
	fullPattern := c.Key(namespace, pattern)
	
	var cursor uint64
	for {
		keys, nextCursor, err := RedisClient.Scan(ctx, cursor, fullPattern, 100).Result()
		if err != nil {
			return err
		}
		
		if len(keys) > 0 {
			RedisClient.Del(ctx, keys...)
		}
		
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}
	
	return nil
}

// ClearTenantCache clears all cache for this tenant
func (c *TenantCache) ClearTenantCache(ctx context.Context) error {
	return c.DeleteByPattern(ctx, "", "*")
}

// ============================================
// ANALYTICS OPERATIONS
// ============================================

// CacheDailyStats caches daily stats
func (c *TenantCache) CacheDailyStats(ctx context.Context, date string, stats interface{}) error {
	return c.Set(ctx, NSAnalytics, "daily:"+date, stats, 1*time.Hour)
}

// GetCachedDailyStats gets cached daily stats
func (c *TenantCache) GetCachedDailyStats(ctx context.Context, date string, dest interface{}) error {
	return c.GetJSON(ctx, NSAnalytics, "daily:"+date, dest)
}

// CacheHourlyStats caches hourly stats
func (c *TenantCache) CacheHourlyStats(ctx context.Context, hour string, stats interface{}) error {
	return c.Set(ctx, NSAnalytics, "hourly:"+hour, stats, 30*time.Minute)
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// TenantKey builds a tenant-scoped key (static helper)
func TenantKey(tenantID uuid.UUID, namespace, key string) string {
	return fmt.Sprintf("tenant:%s:%s%s", tenantID.String(), namespace, key)
}

// ParseTenantKey parses a tenant key to extract tenant ID
func ParseTenantKey(key string) (uuid.UUID, string, error) {
	// Format: tenant:{uuid}:{namespace}{key}
	var tenantIDStr string
	var rest string
	
	n, err := fmt.Sscanf(key, "tenant:%s:", &tenantIDStr)
	if err != nil || n != 1 {
		return uuid.Nil, "", fmt.Errorf("invalid tenant key format")
	}
	
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return uuid.Nil, "", err
	}
	
	// Extract the rest of the key
	prefixLen := len("tenant:") + len(tenantIDStr) + 1
	if len(key) > prefixLen {
		rest = key[prefixLen:]
	}
	
	return tenantID, rest, nil
}


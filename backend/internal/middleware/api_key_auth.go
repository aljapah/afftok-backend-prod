package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// Ensure models is used
var _ = models.APIKeyStatusActive

// ============================================
// API KEY AUTHENTICATION MIDDLEWARE
// ============================================

// Context keys for API key auth
const (
	ContextAPIKeyID      = "api_key_id"
	ContextAdvertiserID  = "advertiser_id"
	ContextAPIKeyName    = "api_key_name"
	ContextAPIKeyPerms   = "api_key_permissions"
	ContextAuthMethod    = "auth_method"
	AuthMethodAPIKey     = "api_key"
	AuthMethodJWT        = "jwt"
)

// API Key rate limit configuration
const (
	apiKeyRateLimitPerMinute = 60
	apiKeyRateLimitPerIP     = 20
	apiKeyRateLimitBurst     = 10
	apiKeyBlockDuration      = 5 * time.Minute
)

// API Key fraud indicators
const (
	FraudIndicatorInvalidAPIKey    = "invalid_api_key"
	FraudIndicatorRevokedAPIKey    = "revoked_api_key"
	FraudIndicatorExpiredAPIKey    = "expired_api_key"
	FraudIndicatorAPIKeyRateLimit  = "api_key_rate_limit"
	FraudIndicatorAPIKeyIPViolation = "api_key_ip_violation"
	FraudIndicatorAPIKeyBruteForce = "api_key_brute_force"
)

// APIKeyMetrics tracks API key usage metrics
type APIKeyMetrics struct {
	TotalAttempts    int64
	SuccessAttempts  int64
	FailedAttempts   int64
	BlockedAttempts  int64
	RateLimitBlocks  int64
	IPViolations     int64
}

var (
	apiKeyMetrics = &APIKeyMetrics{}
	apiKeyService *services.APIKeyService
	apiKeyObsService *services.ObservabilityService
	apiKeyServiceMutex sync.RWMutex
)

// InitAPIKeyMiddleware initializes the API key middleware with required services
func InitAPIKeyMiddleware(keyService *services.APIKeyService, obsService *services.ObservabilityService) {
	apiKeyServiceMutex.Lock()
	defer apiKeyServiceMutex.Unlock()
	apiKeyService = keyService
	apiKeyObsService = obsService
}

// ============================================
// MAIN MIDDLEWARE
// ============================================

// APIKeyAuthMiddleware authenticates requests using API keys
// Accepts key via:
// - Authorization: Bearer <key>
// - X-API-Key: <key>
func APIKeyAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		ip := c.ClientIP()
		endpoint := c.Request.URL.Path
		method := c.Request.Method

		atomic.AddInt64(&apiKeyMetrics.TotalAttempts, 1)

		// Extract API key
		apiKey := extractAPIKey(c)
		if apiKey == "" {
			// No API key provided - allow through for JWT auth fallback
			c.Next()
			return
		}

		// Check for brute force
		if isAPIKeyBruteForce(ip) {
			atomic.AddInt64(&apiKeyMetrics.BlockedAttempts, 1)
			logAPIKeyFraud(ip, "", FraudIndicatorAPIKeyBruteForce, "Too many failed attempts")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Too many failed attempts. Please try again later.",
				"code":    "API_KEY_BLOCKED",
			})
			return
		}

		// Validate API key
		apiKeyServiceMutex.RLock()
		service := apiKeyService
		apiKeyServiceMutex.RUnlock()

		if service == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "API key service not initialized",
			})
			return
		}

		keyInfo, err := service.ValidateAPIKeyWithIP(apiKey, ip)
		if err != nil {
			atomic.AddInt64(&apiKeyMetrics.FailedAttempts, 1)
			recordAPIKeyFailure(ip)

			// Determine fraud indicator
			indicator := FraudIndicatorInvalidAPIKey
			errorMsg := "Invalid API key"
			
			if strings.Contains(err.Error(), "revoked") {
				indicator = FraudIndicatorRevokedAPIKey
				errorMsg = "API key has been revoked"
			} else if strings.Contains(err.Error(), "expired") {
				indicator = FraudIndicatorExpiredAPIKey
				errorMsg = "API key has expired"
			} else if strings.Contains(err.Error(), "IP not allowed") {
				indicator = FraudIndicatorAPIKeyIPViolation
				errorMsg = "IP address not allowed for this API key"
				atomic.AddInt64(&apiKeyMetrics.IPViolations, 1)
			}

			logAPIKeyFraud(ip, apiKey[:min(20, len(apiKey))], indicator, err.Error())
			logAPIKeyUsage(uuid.Nil, uuid.Nil, ip, endpoint, method, c.GetHeader("User-Agent"), false, http.StatusUnauthorized, errorMsg, time.Since(startTime).Milliseconds())

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   errorMsg,
				"code":    "INVALID_API_KEY",
			})
			return
		}

		// Check rate limits
		if !checkAPIKeyRateLimit(keyInfo.ID.String(), ip, keyInfo.RateLimitPerMinute) {
			atomic.AddInt64(&apiKeyMetrics.RateLimitBlocks, 1)
			logAPIKeyFraud(ip, keyInfo.KeyHint, FraudIndicatorAPIKeyRateLimit, "Rate limit exceeded")
			logAPIKeyUsage(keyInfo.ID, keyInfo.AdvertiserID, ip, endpoint, method, c.GetHeader("User-Agent"), false, http.StatusTooManyRequests, "Rate limit exceeded", time.Since(startTime).Milliseconds())

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Rate limit exceeded",
				"code":    "API_KEY_RATE_LIMIT",
			})
			return
		}

		// Success!
		atomic.AddInt64(&apiKeyMetrics.SuccessAttempts, 1)
		clearAPIKeyFailures(ip)

		// Set context values
		c.Set(ContextAPIKeyID, keyInfo.ID.String())
		c.Set(ContextAdvertiserID, keyInfo.AdvertiserID.String())
		c.Set(ContextAPIKeyName, keyInfo.Name)
		c.Set(ContextAuthMethod, AuthMethodAPIKey)

		// Parse and set permissions
		var permissions []string
		if keyInfo.Permissions != nil && len(keyInfo.Permissions) > 0 {
			if err := json.Unmarshal(keyInfo.Permissions, &permissions); err != nil {
				permissions = []string{"*"} // Default to all if parse fails
			}
		}
		c.Set(ContextAPIKeyPerms, permissions)

		// Increment usage (async)
		go service.IncrementUsage(keyInfo.ID, ip)

		// Log usage (async)
		go logAPIKeyUsage(keyInfo.ID, keyInfo.AdvertiserID, ip, endpoint, method, c.GetHeader("User-Agent"), true, http.StatusOK, "", time.Since(startTime).Milliseconds())

		c.Next()
	}
}

// APIKeyRequiredMiddleware requires API key authentication (no JWT fallback)
func APIKeyRequiredMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		ip := c.ClientIP()

		atomic.AddInt64(&apiKeyMetrics.TotalAttempts, 1)

		// Extract API key
		apiKey := extractAPIKey(c)
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "API key required. Use Authorization: Bearer <key> or X-API-Key header",
				"code":    "API_KEY_REQUIRED",
			})
			return
		}

		// Check for brute force
		if isAPIKeyBruteForce(ip) {
			atomic.AddInt64(&apiKeyMetrics.BlockedAttempts, 1)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Too many failed attempts",
			})
			return
		}

		// Validate
		apiKeyServiceMutex.RLock()
		service := apiKeyService
		apiKeyServiceMutex.RUnlock()

		if service == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "API key service not initialized",
			})
			return
		}

		keyInfo, err := service.ValidateAPIKeyWithIP(apiKey, ip)
		if err != nil {
			atomic.AddInt64(&apiKeyMetrics.FailedAttempts, 1)
			recordAPIKeyFailure(ip)

			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or unauthorized API key",
				"code":    "INVALID_API_KEY",
			})
			return
		}

		// Check rate limits
		if !checkAPIKeyRateLimit(keyInfo.ID.String(), ip, keyInfo.RateLimitPerMinute) {
			atomic.AddInt64(&apiKeyMetrics.RateLimitBlocks, 1)
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "Rate limit exceeded",
			})
			return
		}

		atomic.AddInt64(&apiKeyMetrics.SuccessAttempts, 1)
		clearAPIKeyFailures(ip)

		// Set context
		c.Set(ContextAPIKeyID, keyInfo.ID.String())
		c.Set(ContextAdvertiserID, keyInfo.AdvertiserID.String())
		c.Set(ContextAPIKeyName, keyInfo.Name)
		c.Set(ContextAuthMethod, AuthMethodAPIKey)

		go service.IncrementUsage(keyInfo.ID, ip)
		go logAPIKeyUsage(keyInfo.ID, keyInfo.AdvertiserID, ip, c.Request.URL.Path, c.Request.Method, c.GetHeader("User-Agent"), true, http.StatusOK, "", time.Since(startTime).Milliseconds())

		c.Next()
	}
}

// APIKeyOrJWTMiddleware accepts either API key or JWT authentication
func APIKeyOrJWTMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try API key first
		apiKey := extractAPIKey(c)
		if apiKey != "" && strings.HasPrefix(apiKey, "afftok_") {
			// Use API key auth
			APIKeyAuthMiddleware()(c)
			if c.IsAborted() {
				return
			}
			// Check if API key was validated
			if _, exists := c.Get(ContextAPIKeyID); exists {
				return // API key auth successful
			}
		}

		// Fall back to JWT auth
		AuthMiddleware()(c)
	}
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// extractAPIKey extracts the API key from request headers
func extractAPIKey(c *gin.Context) string {
	// Try X-API-Key header first
	if key := c.GetHeader("X-API-Key"); key != "" {
		return key
	}

	// Try Authorization: Bearer header
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		token := strings.TrimPrefix(auth, "Bearer ")
		// Only return if it looks like an API key (not a JWT)
		if strings.HasPrefix(token, "afftok_") {
			return token
		}
	}

	return ""
}

// ============================================
// RATE LIMITING
// ============================================

// checkAPIKeyRateLimit checks if the request is within rate limits
func checkAPIKeyRateLimit(keyID, ip string, customLimit int) bool {
	ctx := context.Background()
	limit := customLimit
	if limit <= 0 {
		limit = apiKeyRateLimitPerMinute
	}

	// Per-key rate limit
	keyRateKey := fmt.Sprintf("ratelimit:apikey:%s", keyID)
	keyCount, _ := cache.Increment(ctx, keyRateKey)
	if keyCount == 1 {
		// Set expiry on first increment
		cache.Expire(ctx, keyRateKey, time.Minute)
	}
	if keyCount > int64(limit) {
		return false
	}

	// Per-IP per-key rate limit (more restrictive)
	ipKeyRateKey := fmt.Sprintf("ratelimit:apikey:%s:ip:%s", keyID, ip)
	ipCount, _ := cache.Increment(ctx, ipKeyRateKey)
	if ipCount == 1 {
		cache.Expire(ctx, ipKeyRateKey, time.Minute)
	}
	if ipCount > apiKeyRateLimitPerIP {
		return false
	}

	return true
}

// ============================================
// BRUTE FORCE PROTECTION
// ============================================

// recordAPIKeyFailure records a failed API key attempt
func recordAPIKeyFailure(ip string) {
	ctx := context.Background()
	key := fmt.Sprintf("apikey:failures:%s", ip)
	count, _ := cache.Increment(ctx, key)
	if count == 1 {
		cache.Expire(ctx, key, 10*time.Minute)
	}
	
	// Block after 10 failures
	if count >= 10 {
		blockKey := fmt.Sprintf("apikey:blocked:%s", ip)
		cache.Set(ctx, blockKey, "1", apiKeyBlockDuration)
	}
}

// isAPIKeyBruteForce checks if an IP is blocked due to brute force
func isAPIKeyBruteForce(ip string) bool {
	ctx := context.Background()
	blockKey := fmt.Sprintf("apikey:blocked:%s", ip)
	blocked, _ := cache.Get(ctx, blockKey)
	return blocked == "1"
}

// clearAPIKeyFailures clears failure count for an IP
func clearAPIKeyFailures(ip string) {
	ctx := context.Background()
	key := fmt.Sprintf("apikey:failures:%s", ip)
	cache.Delete(ctx, key)
}

// ============================================
// LOGGING & OBSERVABILITY
// ============================================

// logAPIKeyUsage logs API key usage
func logAPIKeyUsage(keyID, advertiserID uuid.UUID, ip, endpoint, method, userAgent string, success bool, statusCode int, errorReason string, latencyMs int64) {
	apiKeyServiceMutex.RLock()
	service := apiKeyService
	obsService := apiKeyObsService
	apiKeyServiceMutex.RUnlock()

	// Log to API key service
	if service != nil && keyID != uuid.Nil {
		service.LogUsage(keyID, advertiserID, ip, endpoint, method, userAgent, success, statusCode, errorReason, latencyMs)
	}

	// Log to observability
	if obsService != nil {
		obsService.Log(services.LogEvent{
			Timestamp:   time.Now(),
			Level:       getLogLevel(success),
			Category:    "api_key_event",
			Message:     getMessage(success, errorReason),
			UserID:      advertiserID.String(),
			IP:          ip,
			UserAgent:   userAgent,
			DurationMs:  latencyMs,
			Metadata: map[string]interface{}{
				"api_key_id": keyID.String(),
				"endpoint":   endpoint,
				"method":     method,
				"success":    success,
				"status_code": statusCode,
				"error":      errorReason,
			},
		})
	}
}

// logAPIKeyFraud logs fraud detection for API keys
func logAPIKeyFraud(ip, keyHint, indicator, reason string) {
	apiKeyServiceMutex.RLock()
	obsService := apiKeyObsService
	apiKeyServiceMutex.RUnlock()

	if obsService != nil {
		obsService.LogFraud(ip, "", reason, 80, 0.8, []string{indicator}, map[string]interface{}{
			"key_hint":  keyHint,
			"indicator": indicator,
		})
	}
}

func getLogLevel(success bool) string {
	if success {
		return services.LogLevelInfo
	}
	return services.LogLevelWarn
}

func getMessage(success bool, errorReason string) string {
	if success {
		return "API key authentication successful"
	}
	if errorReason != "" {
		return "API key authentication failed: " + errorReason
	}
	return "API key authentication failed"
}

// ============================================
// METRICS
// ============================================

// GetAPIKeyMetrics returns API key metrics
func GetAPIKeyMetrics() *APIKeyMetrics {
	return &APIKeyMetrics{
		TotalAttempts:   atomic.LoadInt64(&apiKeyMetrics.TotalAttempts),
		SuccessAttempts: atomic.LoadInt64(&apiKeyMetrics.SuccessAttempts),
		FailedAttempts:  atomic.LoadInt64(&apiKeyMetrics.FailedAttempts),
		BlockedAttempts: atomic.LoadInt64(&apiKeyMetrics.BlockedAttempts),
		RateLimitBlocks: atomic.LoadInt64(&apiKeyMetrics.RateLimitBlocks),
		IPViolations:    atomic.LoadInt64(&apiKeyMetrics.IPViolations),
	}
}

// ============================================
// PERMISSION CHECKING
// ============================================

// RequirePermission middleware checks for specific permission
func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if authenticated via API key
		permsInterface, exists := c.Get(ContextAPIKeyPerms)
		if !exists {
			// Not API key auth, allow through (JWT has full access)
			c.Next()
			return
		}

		perms, ok := permsInterface.([]string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Invalid permissions",
			})
			return
		}

		// Check for permission
		for _, p := range perms {
			if p == permission || p == "*" {
				c.Next()
				return
			}
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   fmt.Sprintf("Missing required permission: %s", permission),
			"code":    "PERMISSION_DENIED",
		})
	}
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}


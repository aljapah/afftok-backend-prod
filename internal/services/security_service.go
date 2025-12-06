package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SecurityService handles security, anti-fraud, and rate limiting
type SecurityService struct {
	hmacSecret      []byte
	rateLimitWindow time.Duration
	mutex           sync.RWMutex
}

var (
	securityServiceInstance *SecurityService
	securityServiceOnce     sync.Once
)

// NewSecurityService creates a singleton SecurityService
func NewSecurityService() *SecurityService {
	securityServiceOnce.Do(func() {
		secret := getSecuritySecret()
		securityServiceInstance = &SecurityService{
			hmacSecret:      []byte(secret),
			rateLimitWindow: time.Minute,
		}
	})
	return securityServiceInstance
}

func getSecuritySecret() string {
	// TODO: Load from environment variable in production
	return "afftok-security-hmac-secret-2025-v2"
}

// ============================================
// TRACKING CODE SECURITY
// ============================================

// SecureTrackingCode represents a tamper-proof tracking code
type SecureTrackingCode struct {
	UserOfferID uuid.UUID
	Timestamp   int64
	Nonce       string
	Signature   string
}

// GenerateSecureTrackingCode creates a tamper-proof tracking code with timestamp and nonce
func (s *SecurityService) GenerateSecureTrackingCode(userOfferID uuid.UUID) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Generate random nonce (8 bytes)
	nonceBytes := make([]byte, 8)
	if _, err := rand.Read(nonceBytes); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonce := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(nonceBytes)

	// Current timestamp (Unix seconds)
	timestamp := time.Now().Unix()

	// Create data to sign: userOfferID|timestamp|nonce
	dataToSign := fmt.Sprintf("%s|%d|%s", userOfferID.String(), timestamp, nonce)

	// Generate HMAC signature
	signature := s.sign(dataToSign)

	// Encode: base64(userOfferID_first8)|timestamp_hex|nonce|signature_first12
	// This creates a compact but secure code
	code := fmt.Sprintf("%s-%s-%s-%s",
		userOfferID.String()[:8],
		strconv.FormatInt(timestamp, 36), // Base36 for compactness
		nonce[:8],
		signature[:12],
	)

	// Cache the mapping for fast lookup
	ctx := context.Background()
	cacheKey := fmt.Sprintf("secure_tracking:%s", code)
	cacheValue := fmt.Sprintf("%s|%d", userOfferID.String(), timestamp)
	
	if cache.RedisClient != nil {
		cache.Set(ctx, cacheKey, cacheValue, 365*24*time.Hour)
	}

	return code, nil
}

// ValidateSecureTrackingCode validates a tracking code and returns the userOfferID
func (s *SecurityService) ValidateSecureTrackingCode(code string, maxAge time.Duration) (uuid.UUID, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("secure_tracking:%s", code)

	// Try Redis first
	if cache.RedisClient != nil {
		result, err := cache.Get(ctx, cacheKey)
		if err == nil && result != "" {
			parts := strings.Split(result, "|")
			if len(parts) >= 2 {
				// Check timestamp expiration
				timestamp, _ := strconv.ParseInt(parts[1], 10, 64)
				if maxAge > 0 && time.Since(time.Unix(timestamp, 0)) > maxAge {
					return uuid.Nil, fmt.Errorf("tracking code expired")
				}
				return uuid.Parse(parts[0])
			}
		}
	}

	return uuid.Nil, fmt.Errorf("invalid or expired tracking code")
}

// sign creates an HMAC-SHA256 signature
func (s *SecurityService) sign(data string) string {
	h := hmac.New(sha256.New, s.hmacSecret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// ============================================
// CLICK FINGERPRINTING & ANTI-FRAUD
// ============================================

// ClickFingerprint represents a unique click identifier for deduplication
type ClickFingerprint struct {
	Hash      string
	IP        string
	UserAgent string
	Timestamp int64
	Salt      string
}

// GenerateClickFingerprint creates a secure fingerprint for click deduplication
func (s *SecurityService) GenerateClickFingerprint(userOfferID uuid.UUID, ip, userAgent string) string {
	// Generate random salt
	saltBytes := make([]byte, 4)
	rand.Read(saltBytes)
	salt := hex.EncodeToString(saltBytes)

	// Create fingerprint data
	// Include: userOfferID + IP + UserAgent hash + time window + salt
	uaHash := s.hashUserAgent(userAgent)
	timeWindow := time.Now().Unix() / 60 // 1-minute window

	data := fmt.Sprintf("%s:%s:%s:%d:%s",
		userOfferID.String(),
		s.hashIP(ip),
		uaHash,
		timeWindow,
		salt,
	)

	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))[:24]
}

// hashIP creates a privacy-preserving hash of IP address
func (s *SecurityService) hashIP(ip string) string {
	h := sha256.New()
	h.Write([]byte(ip + string(s.hmacSecret[:8])))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// hashUserAgent creates a normalized hash of user agent
func (s *SecurityService) hashUserAgent(ua string) string {
	// Normalize user agent (remove version numbers for consistency)
	normalized := regexp.MustCompile(`[\d.]+`).ReplaceAllString(strings.ToLower(ua), "")
	h := sha256.New()
	h.Write([]byte(normalized))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// IsClickDuplicate checks if a click is a duplicate within the time window
func (s *SecurityService) IsClickDuplicate(fingerprint string, window time.Duration) bool {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("click_fp:%s", fingerprint)

	if cache.RedisClient == nil {
		return false
	}

	exists, err := cache.Exists(ctx, cacheKey)
	if err != nil {
		return false
	}

	if exists > 0 {
		return true
	}

	// Mark as seen
	cache.Set(ctx, cacheKey, "1", window)
	return false
}

// ============================================
// BOT DETECTION
// ============================================

// BotDetectionResult represents the result of bot detection
type BotDetectionResult struct {
	IsBot       bool
	Confidence  float64
	Reason      string
	RiskScore   int // 0-100
}

// Known bot patterns
var botPatterns = []string{
	"bot", "crawler", "spider", "scraper", "curl", "wget", "python",
	"java/", "httpclient", "okhttp", "axios", "node-fetch", "go-http",
	"headless", "phantom", "selenium", "puppeteer", "playwright",
	"googlebot", "bingbot", "yandex", "baiduspider", "facebookexternalhit",
	"twitterbot", "linkedinbot", "slackbot", "telegrambot",
}

// DetectBot analyzes request for bot-like behavior
func (s *SecurityService) DetectBot(c *gin.Context) BotDetectionResult {
	result := BotDetectionResult{
		IsBot:      false,
		Confidence: 0,
		RiskScore:  0,
	}

	ua := strings.ToLower(c.Request.UserAgent())
	
	// Check for empty user agent
	if ua == "" {
		result.IsBot = true
		result.Confidence = 0.9
		result.Reason = "empty_user_agent"
		result.RiskScore = 90
		return result
	}

	// Check for known bot patterns
	for _, pattern := range botPatterns {
		if strings.Contains(ua, pattern) {
			result.IsBot = true
			result.Confidence = 0.95
			result.Reason = fmt.Sprintf("bot_pattern:%s", pattern)
			result.RiskScore = 95
			return result
		}
	}

	// Check for missing common headers
	if c.GetHeader("Accept-Language") == "" {
		result.RiskScore += 20
	}
	if c.GetHeader("Accept-Encoding") == "" {
		result.RiskScore += 15
	}
	if c.GetHeader("Accept") == "" {
		result.RiskScore += 15
	}

	// Check for suspicious user agent patterns
	if len(ua) < 20 {
		result.RiskScore += 25
	}
	if !strings.Contains(ua, "mozilla") && !strings.Contains(ua, "opera") {
		result.RiskScore += 20
	}

	// Check for datacenter IP ranges (simplified)
	ip := c.ClientIP()
	if s.isDatacenterIP(ip) {
		result.RiskScore += 30
	}

	// Determine if bot based on risk score
	if result.RiskScore >= 70 {
		result.IsBot = true
		result.Confidence = float64(result.RiskScore) / 100
		result.Reason = fmt.Sprintf("high_risk_score:%d", result.RiskScore)
	}

	return result
}

// isDatacenterIP checks if IP belongs to known datacenter ranges
func (s *SecurityService) isDatacenterIP(ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	// Common datacenter CIDR ranges (simplified)
	datacenterRanges := []string{
		"104.16.0.0/12",   // Cloudflare
		"172.64.0.0/13",   // Cloudflare
		"34.0.0.0/8",      // Google Cloud
		"35.0.0.0/8",      // Google Cloud
		"52.0.0.0/8",      // AWS
		"54.0.0.0/8",      // AWS
		"13.0.0.0/8",      // Azure
		"20.0.0.0/8",      // Azure
		"157.240.0.0/16",  // Facebook
	}

	for _, cidr := range datacenterRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}

	return false
}

// ============================================
// RATE LIMITING
// ============================================

// RateLimitResult represents the result of a rate limit check
type RateLimitResult struct {
	Allowed   bool
	Remaining int
	ResetAt   time.Time
	Reason    string
}

// CheckRateLimit checks if a request is within rate limits
func (s *SecurityService) CheckRateLimit(key string, limit int, window time.Duration) RateLimitResult {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("ratelimit:%s", key)

	result := RateLimitResult{
		Allowed:   true,
		Remaining: limit,
		ResetAt:   time.Now().Add(window),
	}

	if cache.RedisClient == nil {
		return result // Allow if Redis unavailable
	}

	// Get current count
	countStr, err := cache.Get(ctx, cacheKey)
	if err != nil {
		// First request in window
		cache.Set(ctx, cacheKey, "1", window)
		result.Remaining = limit - 1
		return result
	}

	count, _ := strconv.Atoi(countStr)
	
	if count >= limit {
		result.Allowed = false
		result.Remaining = 0
		result.Reason = "rate_limit_exceeded"
		
		// Get TTL for reset time
		if ttl, err := cache.RedisClient.TTL(ctx, cacheKey).Result(); err == nil {
			result.ResetAt = time.Now().Add(ttl)
		}
		
		return result
	}

	// Increment counter
	cache.Increment(ctx, cacheKey)
	result.Remaining = limit - count - 1

	return result
}

// Rate limit configurations
const (
	RateLimitClicksPerMinute    = 30
	RateLimitClicksPerHour      = 300
	RateLimitStatsPerMinute     = 60
	RateLimitPostbackPerMinute  = 100
	RateLimitAuthPerMinute      = 100
	RateLimitJoinOfferPerMinute = 20
)

// CheckClickRateLimit checks click-specific rate limits
func (s *SecurityService) CheckClickRateLimit(ip string, userOfferID uuid.UUID) RateLimitResult {
	// Per-IP limit
	ipKey := fmt.Sprintf("click:ip:%s", s.hashIP(ip))
	ipResult := s.CheckRateLimit(ipKey, RateLimitClicksPerMinute, time.Minute)
	if !ipResult.Allowed {
		return ipResult
	}

	// Per-IP-per-offer limit (stricter)
	ipOfferKey := fmt.Sprintf("click:ip_offer:%s:%s", s.hashIP(ip), userOfferID.String()[:8])
	ipOfferResult := s.CheckRateLimit(ipOfferKey, 5, time.Minute)
	if !ipOfferResult.Allowed {
		ipOfferResult.Reason = "too_many_clicks_same_offer"
		return ipOfferResult
	}

	return ipResult
}

// ============================================
// REQUEST VALIDATION
// ============================================

// ValidateRequest performs security validation on incoming request
func (s *SecurityService) ValidateRequest(c *gin.Context) error {
	// Check for required headers
	if c.Request.UserAgent() == "" {
		return fmt.Errorf("missing user agent")
	}

	// Check for suspicious headers
	if c.GetHeader("X-Forwarded-For") != "" {
		// Validate X-Forwarded-For format
		xff := c.GetHeader("X-Forwarded-For")
		if len(xff) > 500 {
			return fmt.Errorf("invalid X-Forwarded-For header")
		}
	}

	// Check content length for POST requests
	if c.Request.Method == "POST" && c.Request.ContentLength > 1024*1024 {
		return fmt.Errorf("request body too large")
	}

	return nil
}

// ============================================
// POSTBACK SECURITY
// ============================================

// PostbackValidationResult represents postback validation result
type PostbackValidationResult struct {
	Valid    bool
	Reason   string
	NetworkID string
}

// ValidatePostback validates an incoming postback request
func (s *SecurityService) ValidatePostback(c *gin.Context, expectedToken string) PostbackValidationResult {
	result := PostbackValidationResult{Valid: true}

	// Check for token/signature
	token := c.Query("token")
	if token == "" {
		token = c.Query("sig")
	}
	if token == "" {
		token = c.GetHeader("X-Postback-Token")
	}

	// If expected token is set, validate it
	if expectedToken != "" && token != expectedToken {
		result.Valid = false
		result.Reason = "invalid_token"
		return result
	}

	// Check for replay attack (same external_id within short window)
	externalID := c.Query("external_id")
	if externalID == "" {
		externalID = c.Query("transaction_id")
	}
	
	if externalID != "" {
		ctx := context.Background()
		replayKey := fmt.Sprintf("postback_replay:%s", externalID)
		
		if cache.RedisClient != nil {
			exists, _ := cache.Exists(ctx, replayKey)
			if exists > 0 {
				result.Valid = false
				result.Reason = "replay_detected"
				return result
			}
			// Mark as processed
			cache.Set(ctx, replayKey, "1", 24*time.Hour)
		}
	}

	// Validate IP if allowlist is configured
	// TODO: Implement IP allowlist from config

	return result
}

// ============================================
// AUDIT LOGGING
// ============================================

// AuditEvent represents a security audit event
type AuditEvent struct {
	Timestamp   time.Time
	EventType   string
	UserID      string
	IP          string
	UserAgent   string
	Resource    string
	Action      string
	Success     bool
	Details     map[string]interface{}
}

// LogAuditEvent logs a security audit event
func (s *SecurityService) LogAuditEvent(event AuditEvent) {
	// For now, just print to stdout
	// In production, this should go to a dedicated audit log
	fmt.Printf("[AUDIT] %s | type=%s user=%s ip=%s resource=%s action=%s success=%v\n",
		event.Timestamp.Format(time.RFC3339),
		event.EventType,
		event.UserID,
		event.IP,
		event.Resource,
		event.Action,
		event.Success,
	)

	// TODO: Store in database or send to logging service
}

// ============================================
// INPUT SANITIZATION
// ============================================

// SanitizeString removes potentially dangerous characters
func (s *SecurityService) SanitizeString(input string, maxLen int) string {
	if len(input) > maxLen {
		input = input[:maxLen]
	}
	
	// Remove null bytes
	input = strings.ReplaceAll(input, "\x00", "")
	
	// Remove control characters
	input = regexp.MustCompile(`[\x00-\x1f\x7f]`).ReplaceAllString(input, "")
	
	return strings.TrimSpace(input)
}

// ValidateUUID validates and parses a UUID string
func (s *SecurityService) ValidateUUID(input string) (uuid.UUID, error) {
	input = strings.TrimSpace(input)
	if len(input) != 36 {
		return uuid.Nil, fmt.Errorf("invalid UUID length")
	}
	return uuid.Parse(input)
}


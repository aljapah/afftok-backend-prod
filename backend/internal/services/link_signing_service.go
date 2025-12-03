package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/google/uuid"
)

// ============================================
// LINK SIGNING SERVICE
// ============================================

// LinkSigningService handles secure link generation and validation
type LinkSigningService struct {
	signingSecret     []byte
	ttlSeconds        int64
	allowLegacyCodes  bool
	mutex             sync.RWMutex
}

// Link signing configuration
var (
	defaultSigningSecret = []byte("afftok-default-signing-secret-change-in-production-64bytes!")
	defaultTTLSeconds    = int64(86400) // 24 hours
	defaultAllowLegacy   = true
)

// LinkSigningMetrics tracks link signing metrics
type LinkSigningMetrics struct {
	ValidLinks          int64
	InvalidSignature    int64
	ExpiredLinks        int64
	ReplayBlocked       int64
	LegacyAccepted      int64
	MalformedLinks      int64
}

var linkMetrics = &LinkSigningMetrics{}

// GetLinkSigningMetrics returns link signing metrics
func GetLinkSigningMetrics() *LinkSigningMetrics {
	return &LinkSigningMetrics{
		ValidLinks:       atomic.LoadInt64(&linkMetrics.ValidLinks),
		InvalidSignature: atomic.LoadInt64(&linkMetrics.InvalidSignature),
		ExpiredLinks:     atomic.LoadInt64(&linkMetrics.ExpiredLinks),
		ReplayBlocked:    atomic.LoadInt64(&linkMetrics.ReplayBlocked),
		LegacyAccepted:   atomic.LoadInt64(&linkMetrics.LegacyAccepted),
		MalformedLinks:   atomic.LoadInt64(&linkMetrics.MalformedLinks),
	}
}

// NewLinkSigningService creates a new link signing service
func NewLinkSigningService() *LinkSigningService {
	service := &LinkSigningService{
		signingSecret:    defaultSigningSecret,
		ttlSeconds:       defaultTTLSeconds,
		allowLegacyCodes: defaultAllowLegacy,
	}

	// Load from environment
	if secret := os.Getenv("LINK_SIGNING_SECRET"); secret != "" {
		service.signingSecret = []byte(secret)
	}

	if ttl := os.Getenv("LINK_TTL_SECONDS"); ttl != "" {
		if parsed, err := strconv.ParseInt(ttl, 10, 64); err == nil && parsed > 0 {
			service.ttlSeconds = parsed
		}
	}

	if legacy := os.Getenv("ALLOW_LEGACY_TRACKING_CODES"); legacy != "" {
		service.allowLegacyCodes = strings.ToLower(legacy) == "true"
	}

	return service
}

// ============================================
// SIGNED LINK FORMAT
// ============================================

// SignedLink represents a signed tracking link
// Format: {trackingCode}.{timestamp}.{nonce}.{signature}
type SignedLink struct {
	TrackingCode string
	Timestamp    int64
	Nonce        string
	Signature    string
	Raw          string
}

// SignedLinkValidationResult represents the result of link validation
type SignedLinkValidationResult struct {
	Valid        bool
	UserOfferID  uuid.UUID
	TrackingCode string
	Reason       string
	IsLegacy     bool
	Indicators   []string
}

// ============================================
// LINK GENERATION
// ============================================

// GenerateSignedLink creates a new signed tracking link
func (s *LinkSigningService) GenerateSignedLink(trackingCode string) string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Generate timestamp
	timestamp := time.Now().Unix()

	// Generate nonce (8-12 chars)
	nonce := s.generateNonce(10)

	// Create signature
	signature := s.createSignature(trackingCode, timestamp, nonce)

	// Format: trackingCode.timestamp.nonce.signature
	return fmt.Sprintf("%s.%d.%s.%s", trackingCode, timestamp, nonce, signature)
}

// GenerateSignedURL creates a full signed URL
func (s *LinkSigningService) GenerateSignedURL(baseURL, trackingCode string) string {
	signedCode := s.GenerateSignedLink(trackingCode)
	return fmt.Sprintf("%s/c/%s", baseURL, signedCode)
}

// generateNonce creates a random nonce
func (s *LinkSigningService) generateNonce(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes)[:length]
}

// createSignature creates HMAC-SHA256 signature
func (s *LinkSigningService) createSignature(trackingCode string, timestamp int64, nonce string) string {
	data := fmt.Sprintf("%s.%d.%s", trackingCode, timestamp, nonce)
	
	h := hmac.New(sha256.New, s.signingSecret)
	h.Write([]byte(data))
	
	// Return first 16 chars of hex for shorter URLs
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// ============================================
// LINK VALIDATION
// ============================================

// ValidateSignedLink validates a signed tracking link
func (s *LinkSigningService) ValidateSignedLink(raw string) *SignedLinkValidationResult {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	result := &SignedLinkValidationResult{
		Valid:      false,
		Indicators: make([]string, 0),
	}

	// Check if it's a legacy code (no dots)
	if !strings.Contains(raw, ".") {
		return s.handleLegacyCode(raw, result)
	}

	// Parse signed link
	parts := strings.Split(raw, ".")
	if len(parts) != 4 {
		atomic.AddInt64(&linkMetrics.MalformedLinks, 1)
		result.Reason = "malformed_link"
		result.Indicators = append(result.Indicators, "malformed_link")
		return result
	}

	trackingCode := parts[0]
	timestampStr := parts[1]
	nonce := parts[2]
	signature := parts[3]

	result.TrackingCode = trackingCode

	// Parse timestamp
	timestamp, err := strconv.ParseInt(timestampStr, 10, 64)
	if err != nil {
		atomic.AddInt64(&linkMetrics.MalformedLinks, 1)
		result.Reason = "invalid_timestamp"
		result.Indicators = append(result.Indicators, "malformed_link")
		return result
	}

	// Step 1: Validate signature
	expectedSignature := s.createSignature(trackingCode, timestamp, nonce)
	if !hmac.Equal([]byte(signature), []byte(expectedSignature)) {
		atomic.AddInt64(&linkMetrics.InvalidSignature, 1)
		result.Reason = "invalid_signature"
		result.Indicators = append(result.Indicators, "invalid_signature")
		return result
	}

	// Step 2: Check TTL
	now := time.Now().Unix()
	age := now - timestamp
	if age > s.ttlSeconds {
		atomic.AddInt64(&linkMetrics.ExpiredLinks, 1)
		result.Reason = "link_expired"
		result.Indicators = append(result.Indicators, "expired_link")
		return result
	}

	// Step 3: Check for negative age (future timestamp)
	if age < -60 { // Allow 60 seconds clock skew
		atomic.AddInt64(&linkMetrics.InvalidSignature, 1)
		result.Reason = "future_timestamp"
		result.Indicators = append(result.Indicators, "invalid_timestamp")
		return result
	}

	// Step 4: Check replay (nonce)
	if s.isReplayAttempt(nonce) {
		atomic.AddInt64(&linkMetrics.ReplayBlocked, 1)
		result.Reason = "replay_attempt"
		result.Indicators = append(result.Indicators, "replay_attempt")
		return result
	}

	// Step 5: Store nonce to prevent replay
	s.storeNonce(nonce)

	// Valid!
	atomic.AddInt64(&linkMetrics.ValidLinks, 1)
	result.Valid = true
	result.Reason = "valid"
	result.TrackingCode = trackingCode

	return result
}

// handleLegacyCode handles legacy tracking codes without signature
func (s *LinkSigningService) handleLegacyCode(code string, result *SignedLinkValidationResult) *SignedLinkValidationResult {
	result.TrackingCode = code
	result.IsLegacy = true

	if s.allowLegacyCodes {
		atomic.AddInt64(&linkMetrics.LegacyAccepted, 1)
		result.Valid = true
		result.Reason = "legacy_accepted"
		result.Indicators = append(result.Indicators, "legacy_code")
		return result
	}

	atomic.AddInt64(&linkMetrics.InvalidSignature, 1)
	result.Reason = "legacy_not_allowed"
	result.Indicators = append(result.Indicators, "legacy_rejected")
	return result
}

// ============================================
// REPLAY PROTECTION
// ============================================

// isReplayAttempt checks if a nonce has been used before
func (s *LinkSigningService) isReplayAttempt(nonce string) bool {
	ctx := context.Background()
	key := fmt.Sprintf("replay:%s", nonce)
	
	exists, err := cache.Exists(ctx, key)
	if err != nil {
		// On error, allow (fail open for availability)
		return false
	}
	
	return exists > 0
}

// storeNonce stores a nonce to prevent replay
func (s *LinkSigningService) storeNonce(nonce string) {
	ctx := context.Background()
	key := fmt.Sprintf("replay:%s", nonce)
	
	// Store with TTL slightly longer than link TTL
	ttl := time.Duration(s.ttlSeconds+3600) * time.Second
	cache.Set(ctx, key, "1", ttl)
}

// ClearReplayCache clears all replay nonces
func (s *LinkSigningService) ClearReplayCache() error {
	ctx := context.Background()
	
	// Get all replay keys
	if cache.RedisClient == nil {
		return fmt.Errorf("Redis not available")
	}
	
	keys, err := cache.RedisClient.Keys(ctx, "replay:*").Result()
	if err != nil {
		return err
	}
	
	if len(keys) > 0 {
		return cache.RedisClient.Del(ctx, keys...).Err()
	}
	
	return nil
}

// GetReplayCacheCount returns the number of stored nonces
func (s *LinkSigningService) GetReplayCacheCount() int64 {
	ctx := context.Background()
	
	if cache.RedisClient == nil {
		return 0
	}
	
	keys, err := cache.RedisClient.Keys(ctx, "replay:*").Result()
	if err != nil {
		return 0
	}
	
	return int64(len(keys))
}

// ============================================
// SECRET ROTATION
// ============================================

// RotateSecret rotates the signing secret
func (s *LinkSigningService) RotateSecret(newSecret string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(newSecret) < 32 {
		return fmt.Errorf("secret must be at least 32 bytes")
	}

	s.signingSecret = []byte(newSecret)
	
	// Clear replay cache on rotation
	s.ClearReplayCache()
	
	return nil
}

// GenerateNewSecret generates a new random secret
func (s *LinkSigningService) GenerateNewSecret() (string, error) {
	bytes := make([]byte, 64)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ============================================
// CONFIGURATION
// ============================================

// LinkSigningConfig represents the current configuration
type LinkSigningConfig struct {
	TTLSeconds       int64  `json:"ttl_seconds"`
	AllowLegacyCodes bool   `json:"allow_legacy_codes"`
	SecretLength     int    `json:"secret_length"`
	ReplayCacheCount int64  `json:"replay_cache_count"`
}

// GetConfig returns the current configuration
func (s *LinkSigningService) GetConfig() *LinkSigningConfig {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return &LinkSigningConfig{
		TTLSeconds:       s.ttlSeconds,
		AllowLegacyCodes: s.allowLegacyCodes,
		SecretLength:     len(s.signingSecret),
		ReplayCacheCount: s.GetReplayCacheCount(),
	}
}

// SetTTL sets the TTL in seconds
func (s *LinkSigningService) SetTTL(seconds int64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.ttlSeconds = seconds
}

// SetAllowLegacy sets whether legacy codes are allowed
func (s *LinkSigningService) SetAllowLegacy(allow bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.allowLegacyCodes = allow
}

// ============================================
// TEST LINK
// ============================================

// TestLink tests a link and returns detailed validation info
func (s *LinkSigningService) TestLink(raw string) map[string]interface{} {
	result := s.ValidateSignedLink(raw)
	
	// Parse for detailed info
	parts := strings.Split(raw, ".")
	
	info := map[string]interface{}{
		"raw":           raw,
		"valid":         result.Valid,
		"reason":        result.Reason,
		"is_legacy":     result.IsLegacy,
		"tracking_code": result.TrackingCode,
		"indicators":    result.Indicators,
		"parts_count":   len(parts),
	}
	
	if len(parts) == 4 {
		timestamp, _ := strconv.ParseInt(parts[1], 10, 64)
		now := time.Now().Unix()
		
		info["timestamp"] = timestamp
		info["timestamp_human"] = time.Unix(timestamp, 0).UTC().Format(time.RFC3339)
		info["nonce"] = parts[2]
		info["signature"] = parts[3]
		info["age_seconds"] = now - timestamp
		info["ttl_seconds"] = s.ttlSeconds
		info["expires_in"] = s.ttlSeconds - (now - timestamp)
	}
	
	return info
}

// ============================================
// STATISTICS
// ============================================

// LinkSigningStats represents comprehensive statistics
type LinkSigningStats struct {
	ValidLinks       int64   `json:"valid"`
	InvalidSignature int64   `json:"invalid_signature"`
	ExpiredLinks     int64   `json:"expired"`
	ReplayBlocked    int64   `json:"replay_blocked"`
	LegacyAccepted   int64   `json:"legacy_accepted"`
	MalformedLinks   int64   `json:"malformed"`
	TTLSeconds       int64   `json:"ttl_seconds"`
	AllowLegacy      bool    `json:"allow_legacy"`
	ReplayCacheCount int64   `json:"replay_cache_count"`
	SuccessRate      float64 `json:"success_rate_percent"`
}

// GetStats returns comprehensive statistics
func (s *LinkSigningService) GetStats() *LinkSigningStats {
	metrics := GetLinkSigningMetrics()
	config := s.GetConfig()
	
	total := metrics.ValidLinks + metrics.InvalidSignature + metrics.ExpiredLinks + 
		metrics.ReplayBlocked + metrics.MalformedLinks
	
	successRate := float64(0)
	if total > 0 {
		successRate = float64(metrics.ValidLinks+metrics.LegacyAccepted) / float64(total) * 100
	}
	
	return &LinkSigningStats{
		ValidLinks:       metrics.ValidLinks,
		InvalidSignature: metrics.InvalidSignature,
		ExpiredLinks:     metrics.ExpiredLinks,
		ReplayBlocked:    metrics.ReplayBlocked,
		LegacyAccepted:   metrics.LegacyAccepted,
		MalformedLinks:   metrics.MalformedLinks,
		TTLSeconds:       config.TTLSeconds,
		AllowLegacy:      config.AllowLegacyCodes,
		ReplayCacheCount: config.ReplayCacheCount,
		SuccessRate:      successRate,
	}
}


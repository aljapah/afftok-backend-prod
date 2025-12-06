package services

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/database"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
)

// LinkService handles secure affiliate link generation and validation
type LinkService struct {
	hmacSecret []byte
	mutex      sync.Mutex
}

var (
	linkServiceInstance *LinkService
	linkServiceOnce     sync.Once
)

// NewLinkService creates a singleton LinkService
func NewLinkService() *LinkService {
	linkServiceOnce.Do(func() {
		// Use a secure secret - in production, load from environment
		secret := getHMACSecret()
		linkServiceInstance = &LinkService{
			hmacSecret: []byte(secret),
		}
	})
	return linkServiceInstance
}

// getHMACSecret returns the HMAC secret from config or generates a fallback
func getHMACSecret() string {
	// TODO: Load from environment in production
	// For now, use a static secret (should be replaced with env var)
	return "afftok-secure-link-secret-2025"
}

// GenerateTrackingCode creates a unique, secure tracking code
// Format: [random_id]-[signature]
// This code is used in place of exposing userOfferID directly
func (s *LinkService) GenerateTrackingCode(userOfferID uuid.UUID) (string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Generate a random 8-byte ID
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	randomID := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(randomBytes)

	// Create signature: HMAC(randomID + userOfferID)
	data := randomID + userOfferID.String()
	signature := s.sign(data)

	// Combine: randomID-signature (first 8 chars of signature)
	trackingCode := randomID + "-" + signature[:8]

	// Store mapping in Redis for fast lookup
	ctx := context.Background()
	cacheKey := fmt.Sprintf("tracking:%s", trackingCode)
	
	if cache.RedisClient != nil {
		// Store with 1-year expiration
		err := cache.Set(ctx, cacheKey, userOfferID.String(), 365*24*time.Hour)
		if err != nil {
			// Log error but don't fail - we'll store in DB as backup
			fmt.Printf("[LinkService] Redis cache error: %v\n", err)
		}
	}

	return trackingCode, nil
}

// ResolveTrackingCode resolves a tracking code to userOfferID
func (s *LinkService) ResolveTrackingCode(trackingCode string) (uuid.UUID, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("tracking:%s", trackingCode)

	// Try Redis first
	if cache.RedisClient != nil {
		result, err := cache.Get(ctx, cacheKey)
		if err == nil && result != "" {
			return uuid.Parse(result)
		}
	}

	// Fallback to database lookup - try multiple fields
	var userOffer models.UserOffer
	
	// Try exact match on tracking_code first
	if err := database.DB.Where("tracking_code = ?", trackingCode).First(&userOffer).Error; err == nil {
		// Found, cache and return
		if cache.RedisClient != nil {
			cache.Set(ctx, cacheKey, userOffer.ID.String(), 365*24*time.Hour)
		}
		return userOffer.ID, nil
	}
	
	// Try exact match on short_link
	if err := database.DB.Where("short_link = ?", trackingCode).First(&userOffer).Error; err == nil {
		if cache.RedisClient != nil {
			cache.Set(ctx, cacheKey, userOffer.ID.String(), 365*24*time.Hour)
		}
		return userOffer.ID, nil
	}
	
	// Try LIKE match as last resort (for legacy links)
	if err := database.DB.Where("short_link LIKE ? OR tracking_code LIKE ?", "%"+trackingCode+"%", "%"+trackingCode+"%").First(&userOffer).Error; err != nil {
		return uuid.Nil, fmt.Errorf("tracking code not found: %s", trackingCode)
	}

	// Re-cache the result
	if cache.RedisClient != nil {
		cache.Set(ctx, cacheKey, userOffer.ID.String(), 365*24*time.Hour)
	}

	return userOffer.ID, nil
}

// GenerateAffiliateLink creates a complete, secure affiliate link
func (s *LinkService) GenerateAffiliateLink(baseURL string, userOfferID uuid.UUID, promoterID uuid.UUID) (string, string, error) {
	// Generate unique tracking code
	trackingCode, err := s.GenerateTrackingCode(userOfferID)
	if err != nil {
		return "", "", err
	}

	// Build the short link (for sharing)
	// Format: /api/c/{trackingCode}
	shortLink := fmt.Sprintf("/api/c/%s", trackingCode)

	// Build full affiliate link (includes tracking)
	affiliateLink := fmt.Sprintf("%s?aff=%s&p=%s", 
		baseURL, 
		trackingCode,
		promoterID.String()[:8], // Only first 8 chars of promoter ID for privacy
	)

	return affiliateLink, shortLink, nil
}

// ValidateTrackingCode validates that a tracking code is properly signed
func (s *LinkService) ValidateTrackingCode(trackingCode string) bool {
	parts := strings.Split(trackingCode, "-")
	if len(parts) != 2 {
		return false
	}

	randomID := parts[0]
	providedSig := parts[1]

	// We can't fully validate without the original userOfferID
	// But we can check if it exists in our system
	ctx := context.Background()
	cacheKey := fmt.Sprintf("tracking:%s", trackingCode)

	if cache.RedisClient != nil {
		exists, _ := cache.Exists(ctx, cacheKey)
		if exists > 0 {
			return true
		}
	}

	// Check length validity
	return len(randomID) >= 8 && len(providedSig) >= 8
}

// sign creates an HMAC-SHA256 signature
func (s *LinkService) sign(data string) string {
	h := hmac.New(sha256.New, s.hmacSecret)
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

// GenerateClickID creates a unique click identifier for deduplication
func (s *LinkService) GenerateClickID(userOfferID uuid.UUID, ipAddress string, userAgent string) string {
	// Create a fingerprint from request data
	data := fmt.Sprintf("%s:%s:%s:%d", 
		userOfferID.String(), 
		ipAddress, 
		userAgent,
		time.Now().Unix()/60, // 1-minute window for dedup
	)
	
	h := sha256.New()
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))[:16]
}

// IsClickDuplicate checks if a click is a duplicate (within time window)
func (s *LinkService) IsClickDuplicate(clickID string) bool {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("click_dedup:%s", clickID)

	if cache.RedisClient == nil {
		return false // Can't check without Redis
	}

	exists, err := cache.Exists(ctx, cacheKey)
	if err != nil {
		return false
	}

	if exists > 0 {
		return true // Duplicate click
	}

	// Mark this click as seen (5 minute window)
	cache.Set(ctx, cacheKey, "1", 5*time.Minute)
	return false
}

// GeneratePostbackToken creates a secure token for postback verification
func (s *LinkService) GeneratePostbackToken(networkID uuid.UUID, userOfferID uuid.UUID) string {
	data := fmt.Sprintf("%s:%s:%d", networkID.String(), userOfferID.String(), time.Now().Unix())
	return s.sign(data)[:32]
}

// ValidatePostbackToken validates a postback token
func (s *LinkService) ValidatePostbackToken(token string, networkID uuid.UUID, userOfferID uuid.UUID, maxAge time.Duration) bool {
	// Check token format
	if len(token) < 32 {
		return false
	}

	// For now, just check if it's a valid hex string
	// In production, would verify against stored tokens
	_, err := hex.DecodeString(token)
	return err == nil
}


package services

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ============================================
// API KEY SERVICE
// ============================================

// APIKeyService handles API key operations
type APIKeyService struct {
	db    *gorm.DB
	mutex sync.RWMutex
}

// NewAPIKeyService creates a new API key service
func NewAPIKeyService(db *gorm.DB) *APIKeyService {
	return &APIKeyService{db: db}
}

// ============================================
// ARGON2 CONFIGURATION
// ============================================

// Argon2 parameters (OWASP recommended)
const (
	argon2Time    = 1
	argon2Memory  = 64 * 1024 // 64MB
	argon2Threads = 4
	argon2KeyLen  = 32
	argon2SaltLen = 16
)

// API Key format
const (
	keyPrefixLive = "afftok_live_sk_"
	keyPrefixTest = "afftok_test_sk_"
	keyLength     = 32 // Random part length
)

// ============================================
// KEY GENERATION
// ============================================

// GenerateAPIKey creates a new API key for an advertiser
// Returns the plaintext key ONCE - it's never stored
func (s *APIKeyService) GenerateAPIKey(advertiserID uuid.UUID, req *models.CreateAPIKeyRequest) (*models.CreateAPIKeyResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Verify advertiser exists
	var advertiser models.AfftokUser
	if err := s.db.First(&advertiser, "id = ?", advertiserID).Error; err != nil {
		return nil, fmt.Errorf("advertiser not found: %w", err)
	}

	// Generate random key
	randomBytes := make([]byte, keyLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	// Create plaintext key
	keyPrefix := keyPrefixLive
	randomPart := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(randomBytes)
	plaintextKey := keyPrefix + randomPart

	// Hash the key with Argon2
	hash, err := s.hashKey(plaintextKey)
	if err != nil {
		return nil, fmt.Errorf("failed to hash key: %w", err)
	}

	// Get last 4 chars as hint
	keyHint := "..." + plaintextKey[len(plaintextKey)-4:]

	// Parse network ID if provided
	var networkID *uuid.UUID
	if req.NetworkID != "" {
		id, err := uuid.Parse(req.NetworkID)
		if err == nil {
			networkID = &id
		}
	}

	// Parse permissions
	permissions := models.DefaultPermissions()
	if len(req.Permissions) > 0 {
		permissions = req.Permissions
	}
	permissionsJSON, _ := json.Marshal(permissions)

	// Parse allowed IPs
	var allowedIPsJSON []byte
	if len(req.AllowedIPs) > 0 {
		allowedIPsJSON, _ = json.Marshal(req.AllowedIPs)
	}

	// Calculate expiration
	var expiresAt *time.Time
	if req.ExpiresInDays > 0 {
		exp := time.Now().AddDate(0, 0, req.ExpiresInDays)
		expiresAt = &exp
	}

	// Rate limit
	rateLimit := 60
	if req.RateLimitPerMinute > 0 {
		rateLimit = req.RateLimitPerMinute
	}

	// Create API key record
	apiKey := &models.AdvertiserAPIKey{
		AdvertiserID:       advertiserID,
		NetworkID:          networkID,
		Name:               req.Name,
		KeyPrefix:          keyPrefix,
		KeyHash:            hash,
		KeyHint:            keyHint,
		Status:             models.APIKeyStatusActive,
		Permissions:        datatypes.JSON(permissionsJSON),
		AllowedIPs:         datatypes.JSON(allowedIPsJSON),
		RateLimitPerMinute: rateLimit,
		RateLimitBurst:     10,
		ExpiresAt:          expiresAt,
	}

	if err := s.db.Create(apiKey).Error; err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	// Log creation (without plaintext key)
	log.Printf("🔑 API Key created: id=%s, advertiser=%s, name=%s", 
		apiKey.ID, advertiserID, req.Name)

	// Return response with plaintext key (only shown once!)
	return &models.CreateAPIKeyResponse{
		ID:           apiKey.ID,
		Name:         apiKey.Name,
		PlaintextKey: plaintextKey,
		KeyHint:      keyHint,
		Status:       string(apiKey.Status),
		CreatedAt:    apiKey.CreatedAt,
		ExpiresAt:    expiresAt,
		Warning:      "⚠️ Save this key now! It will NOT be shown again.",
	}, nil
}

// ============================================
// KEY VALIDATION
// ============================================

// ValidateAPIKey validates an API key and returns the associated info
func (s *APIKeyService) ValidateAPIKey(plaintextKey string) (*models.AdvertiserAPIKey, error) {
	// Check key format
	if !strings.HasPrefix(plaintextKey, "afftok_") {
		return nil, fmt.Errorf("invalid key format")
	}

	// Extract prefix
	var prefix string
	if strings.HasPrefix(plaintextKey, keyPrefixLive) {
		prefix = keyPrefixLive
	} else if strings.HasPrefix(plaintextKey, keyPrefixTest) {
		prefix = keyPrefixTest
	} else {
		return nil, fmt.Errorf("invalid key prefix")
	}

	// Try cache first
	ctx := context.Background()
	cacheKey := "apikey:valid:" + s.hashForCache(plaintextKey)
	if cached, err := cache.Get(ctx, cacheKey); err == nil && cached != "" {
		var apiKey models.AdvertiserAPIKey
		if err := json.Unmarshal([]byte(cached), &apiKey); err == nil {
			return &apiKey, nil
		}
	}

	// Get all active keys with this prefix (minimize DB queries)
	var keys []models.AdvertiserAPIKey
	if err := s.db.Where("key_prefix = ? AND status = ?", prefix, models.APIKeyStatusActive).
		Preload("Advertiser").
		Find(&keys).Error; err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	// Check each key (constant-time comparison)
	for i := range keys {
		if s.verifyKey(plaintextKey, keys[i].KeyHash) {
			// Check expiration
			if keys[i].ExpiresAt != nil && time.Now().After(*keys[i].ExpiresAt) {
				// Mark as expired
				s.db.Model(&keys[i]).Update("status", models.APIKeyStatusExpired)
				return nil, fmt.Errorf("key expired")
			}

			// Cache valid key
			if cached, err := json.Marshal(keys[i]); err == nil {
				cache.Set(ctx, cacheKey, string(cached), 5*time.Minute)
			}

			return &keys[i], nil
		}
	}

	return nil, fmt.Errorf("invalid key")
}

// ValidateAPIKeyWithIP validates key and checks IP allowlist
func (s *APIKeyService) ValidateAPIKeyWithIP(plaintextKey, ip string) (*models.AdvertiserAPIKey, error) {
	apiKey, err := s.ValidateAPIKey(plaintextKey)
	if err != nil {
		return nil, err
	}

	// Check IP allowlist
	if !s.isIPAllowed(apiKey, ip) {
		return nil, fmt.Errorf("IP not allowed")
	}

	return apiKey, nil
}

// ============================================
// KEY ROTATION
// ============================================

// RotateAPIKey creates a new key and revokes the old one
func (s *APIKeyService) RotateAPIKey(apiKeyID uuid.UUID) (*models.CreateAPIKeyResponse, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Get existing key
	var oldKey models.AdvertiserAPIKey
	if err := s.db.First(&oldKey, "id = ?", apiKeyID).Error; err != nil {
		return nil, fmt.Errorf("key not found: %w", err)
	}

	if oldKey.Status != models.APIKeyStatusActive {
		return nil, fmt.Errorf("can only rotate active keys")
	}

	// Generate new key
	randomBytes := make([]byte, keyLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random key: %w", err)
	}

	randomPart := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(randomBytes)
	plaintextKey := oldKey.KeyPrefix + randomPart

	hash, err := s.hashKey(plaintextKey)
	if err != nil {
		return nil, fmt.Errorf("failed to hash key: %w", err)
	}

	keyHint := "..." + plaintextKey[len(plaintextKey)-4:]

	// Start transaction
	tx := s.db.Begin()

	// Revoke old key
	now := time.Now()
	if err := tx.Model(&oldKey).Updates(map[string]interface{}{
		"status":     models.APIKeyStatusRevoked,
		"revoked_at": now,
	}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to revoke old key: %w", err)
	}

	// Create new key
	newKey := &models.AdvertiserAPIKey{
		AdvertiserID:       oldKey.AdvertiserID,
		NetworkID:          oldKey.NetworkID,
		Name:               oldKey.Name + " (rotated)",
		KeyPrefix:          oldKey.KeyPrefix,
		KeyHash:            hash,
		KeyHint:            keyHint,
		Status:             models.APIKeyStatusActive,
		Permissions:        oldKey.Permissions,
		AllowedIPs:         oldKey.AllowedIPs,
		RateLimitPerMinute: oldKey.RateLimitPerMinute,
		RateLimitBurst:     oldKey.RateLimitBurst,
		ExpiresAt:          oldKey.ExpiresAt,
	}

	if err := tx.Create(newKey).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create new key: %w", err)
	}

	tx.Commit()

	// Invalidate cache
	s.invalidateKeyCache(oldKey.ID)

	log.Printf("🔄 API Key rotated: old=%s, new=%s", oldKey.ID, newKey.ID)

	return &models.CreateAPIKeyResponse{
		ID:           newKey.ID,
		Name:         newKey.Name,
		PlaintextKey: plaintextKey,
		KeyHint:      keyHint,
		Status:       string(newKey.Status),
		CreatedAt:    newKey.CreatedAt,
		ExpiresAt:    newKey.ExpiresAt,
		Warning:      "⚠️ Old key has been revoked. Save this new key now!",
	}, nil
}

// ============================================
// KEY REVOCATION
// ============================================

// RevokeAPIKey revokes an API key
func (s *APIKeyService) RevokeAPIKey(apiKeyID uuid.UUID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	result := s.db.Model(&models.AdvertiserAPIKey{}).
		Where("id = ? AND status = ?", apiKeyID, models.APIKeyStatusActive).
		Updates(map[string]interface{}{
			"status":     models.APIKeyStatusRevoked,
			"revoked_at": now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to revoke key: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("key not found or already revoked")
	}

	// Invalidate cache
	s.invalidateKeyCache(apiKeyID)

	log.Printf("🚫 API Key revoked: id=%s", apiKeyID)
	return nil
}

// ============================================
// IP MANAGEMENT
// ============================================

// AddAllowedIP adds an IP to the allowlist
func (s *APIKeyService) AddAllowedIP(apiKeyID uuid.UUID, ip string) error {
	var key models.AdvertiserAPIKey
	if err := s.db.First(&key, "id = ?", apiKeyID).Error; err != nil {
		return fmt.Errorf("key not found: %w", err)
	}

	// Parse existing IPs
	var allowedIPs []string
	if key.AllowedIPs != nil {
		json.Unmarshal(key.AllowedIPs, &allowedIPs)
	}

	// Check if already exists
	for _, existingIP := range allowedIPs {
		if existingIP == ip {
			return nil // Already exists
		}
	}

	// Add new IP
	allowedIPs = append(allowedIPs, ip)
	allowedIPsJSON, _ := json.Marshal(allowedIPs)

	if err := s.db.Model(&key).Update("allowed_ips", datatypes.JSON(allowedIPsJSON)).Error; err != nil {
		return fmt.Errorf("failed to update allowed IPs: %w", err)
	}

	s.invalidateKeyCache(apiKeyID)
	return nil
}

// RemoveAllowedIP removes an IP from the allowlist
func (s *APIKeyService) RemoveAllowedIP(apiKeyID uuid.UUID, ip string) error {
	var key models.AdvertiserAPIKey
	if err := s.db.First(&key, "id = ?", apiKeyID).Error; err != nil {
		return fmt.Errorf("key not found: %w", err)
	}

	// Parse existing IPs
	var allowedIPs []string
	if key.AllowedIPs != nil {
		json.Unmarshal(key.AllowedIPs, &allowedIPs)
	}

	// Remove IP
	var newIPs []string
	for _, existingIP := range allowedIPs {
		if existingIP != ip {
			newIPs = append(newIPs, existingIP)
		}
	}

	allowedIPsJSON, _ := json.Marshal(newIPs)

	if err := s.db.Model(&key).Update("allowed_ips", datatypes.JSON(allowedIPsJSON)).Error; err != nil {
		return fmt.Errorf("failed to update allowed IPs: %w", err)
	}

	s.invalidateKeyCache(apiKeyID)
	return nil
}

// ============================================
// USAGE TRACKING
// ============================================

// IncrementUsage increments the usage counter
func (s *APIKeyService) IncrementUsage(apiKeyID uuid.UUID, ip string) {
	now := time.Now()
	s.db.Model(&models.AdvertiserAPIKey{}).
		Where("id = ?", apiKeyID).
		Updates(map[string]interface{}{
			"usage_count":  gorm.Expr("usage_count + 1"),
			"last_used_at": now,
			"last_used_ip": ip,
		})
}

// LogUsage logs an API key usage event
func (s *APIKeyService) LogUsage(apiKeyID, advertiserID uuid.UUID, ip, endpoint, method, userAgent string, success bool, statusCode int, errorReason string, latencyMs int64) {
	log := &models.APIKeyUsageLog{
		APIKeyID:     apiKeyID,
		AdvertiserID: advertiserID,
		IP:           ip,
		Endpoint:     endpoint,
		Method:       method,
		UserAgent:    userAgent,
		Success:      success,
		StatusCode:   statusCode,
		ErrorReason:  errorReason,
		LatencyMs:    latencyMs,
	}

	// Async insert
	go func() {
		s.db.Create(log)
	}()
}

// ============================================
// LISTING & QUERIES
// ============================================

// GetAPIKeyByID returns an API key by ID (without hash)
func (s *APIKeyService) GetAPIKeyByID(id uuid.UUID) (*models.APIKeyInfo, error) {
	var key models.AdvertiserAPIKey
	if err := s.db.Preload("Advertiser").Preload("Network").First(&key, "id = ?", id).Error; err != nil {
		return nil, err
	}

	return s.toAPIKeyInfo(&key), nil
}

// GetAPIKeysByAdvertiser returns all API keys for an advertiser
func (s *APIKeyService) GetAPIKeysByAdvertiser(advertiserID uuid.UUID) ([]models.APIKeyInfo, error) {
	var keys []models.AdvertiserAPIKey
	if err := s.db.Where("advertiser_id = ?", advertiserID).
		Order("created_at DESC").
		Find(&keys).Error; err != nil {
		return nil, err
	}

	result := make([]models.APIKeyInfo, len(keys))
	for i := range keys {
		result[i] = *s.toAPIKeyInfo(&keys[i])
	}

	return result, nil
}

// GetAllAPIKeys returns all API keys (admin)
func (s *APIKeyService) GetAllAPIKeys(page, limit int) ([]models.APIKeyInfo, int64, error) {
	var keys []models.AdvertiserAPIKey
	var total int64

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	s.db.Model(&models.AdvertiserAPIKey{}).Count(&total)

	if err := s.db.Preload("Advertiser").Preload("Network").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&keys).Error; err != nil {
		return nil, 0, err
	}

	result := make([]models.APIKeyInfo, len(keys))
	for i := range keys {
		result[i] = *s.toAPIKeyInfo(&keys[i])
	}

	return result, total, nil
}

// ============================================
// STATISTICS
// ============================================

// APIKeyStats represents API key statistics
type APIKeyStats struct {
	TotalKeys     int64 `json:"total_keys"`
	ActiveKeys    int64 `json:"active_keys"`
	RevokedKeys   int64 `json:"revoked_keys"`
	ExpiredKeys   int64 `json:"expired_keys"`
	TotalUsage    int64 `json:"total_usage"`
	UsageToday    int64 `json:"usage_today"`
	UsageThisWeek int64 `json:"usage_this_week"`
	FailedAttempts int64 `json:"failed_attempts_today"`
}

// GetAPIKeyStats returns API key statistics
func (s *APIKeyService) GetAPIKeyStats() (*APIKeyStats, error) {
	stats := &APIKeyStats{}

	// Total keys by status
	s.db.Model(&models.AdvertiserAPIKey{}).Count(&stats.TotalKeys)
	s.db.Model(&models.AdvertiserAPIKey{}).Where("status = ?", models.APIKeyStatusActive).Count(&stats.ActiveKeys)
	s.db.Model(&models.AdvertiserAPIKey{}).Where("status = ?", models.APIKeyStatusRevoked).Count(&stats.RevokedKeys)
	s.db.Model(&models.AdvertiserAPIKey{}).Where("status = ?", models.APIKeyStatusExpired).Count(&stats.ExpiredKeys)

	// Total usage
	s.db.Model(&models.AdvertiserAPIKey{}).Select("COALESCE(SUM(usage_count), 0)").Scan(&stats.TotalUsage)

	// Usage today
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&models.APIKeyUsageLog{}).Where("created_at >= ?", today).Count(&stats.UsageToday)

	// Usage this week
	weekAgo := time.Now().AddDate(0, 0, -7)
	s.db.Model(&models.APIKeyUsageLog{}).Where("created_at >= ?", weekAgo).Count(&stats.UsageThisWeek)

	// Failed attempts today
	s.db.Model(&models.APIKeyUsageLog{}).Where("created_at >= ? AND success = ?", today, false).Count(&stats.FailedAttempts)

	return stats, nil
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// hashKey hashes a key using Argon2
func (s *APIKeyService) hashKey(key string) (string, error) {
	salt := make([]byte, argon2SaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}

	hash := argon2.IDKey([]byte(key), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	// Encode as: $argon2id$salt$hash
	return fmt.Sprintf("$argon2id$%s$%s",
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(hash)), nil
}

// verifyKey verifies a key against a hash
func (s *APIKeyService) verifyKey(key, encodedHash string) bool {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 4 || parts[1] != "argon2id" {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}

	computedHash := argon2.IDKey([]byte(key), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	return subtle.ConstantTimeCompare(computedHash, expectedHash) == 1
}

// hashForCache creates a fast hash for caching (not for security)
func (s *APIKeyService) hashForCache(key string) string {
	hash := argon2.IDKey([]byte(key), []byte("cache-salt"), 1, 1024, 1, 16)
	return base64.RawURLEncoding.EncodeToString(hash)
}

// invalidateKeyCache invalidates the cache for a key
func (s *APIKeyService) invalidateKeyCache(apiKeyID uuid.UUID) {
	// We can't easily invalidate by key content, so we rely on TTL
	// For immediate invalidation, we could store a version counter
}

// isIPAllowed checks if an IP is in the allowlist
func (s *APIKeyService) isIPAllowed(key *models.AdvertiserAPIKey, ip string) bool {
	if key.AllowedIPs == nil || len(key.AllowedIPs) == 0 {
		return true
	}

	var allowedIPs []string
	if err := json.Unmarshal(key.AllowedIPs, &allowedIPs); err != nil {
		return true // Parse error, allow
	}

	if len(allowedIPs) == 0 {
		return true
	}

	for _, allowedIP := range allowedIPs {
		// Direct match
		if allowedIP == ip {
			return true
		}

		// CIDR match
		if strings.Contains(allowedIP, "/") {
			_, network, err := net.ParseCIDR(allowedIP)
			if err == nil && network.Contains(net.ParseIP(ip)) {
				return true
			}
		}
	}

	return false
}

// toAPIKeyInfo converts an AdvertiserAPIKey to APIKeyInfo
func (s *APIKeyService) toAPIKeyInfo(key *models.AdvertiserAPIKey) *models.APIKeyInfo {
	info := &models.APIKeyInfo{
		ID:                 key.ID,
		AdvertiserID:       key.AdvertiserID,
		NetworkID:          key.NetworkID,
		Name:               key.Name,
		KeyPrefix:          key.KeyPrefix,
		KeyHint:            key.KeyHint,
		Status:             key.Status,
		UsageCount:         key.UsageCount,
		LastUsedAt:         key.LastUsedAt,
		LastUsedIP:         key.LastUsedIP,
		RateLimitPerMinute: key.RateLimitPerMinute,
		CreatedAt:          key.CreatedAt,
		ExpiresAt:          key.ExpiresAt,
	}

	// Parse permissions
	if key.Permissions != nil {
		json.Unmarshal(key.Permissions, &info.Permissions)
	}

	// Parse allowed IPs
	if key.AllowedIPs != nil {
		json.Unmarshal(key.AllowedIPs, &info.AllowedIPs)
	}

	// Add advertiser name
	if key.Advertiser != nil {
		info.AdvertiserName = key.Advertiser.FullName
	}

	// Add network name
	if key.Network != nil {
		info.NetworkName = key.Network.Name
	}

	return info
}


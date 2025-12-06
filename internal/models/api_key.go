package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ============================================
// ADVERTISER API KEY MODEL
// ============================================

// APIKeyStatus represents the status of an API key
type APIKeyStatus string

const (
	APIKeyStatusActive  APIKeyStatus = "active"
	APIKeyStatusRevoked APIKeyStatus = "revoked"
	APIKeyStatusExpired APIKeyStatus = "expired"
)

// AdvertiserAPIKey represents an API key for advertiser/network authentication
type AdvertiserAPIKey struct {
	ID           uuid.UUID          `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	AdvertiserID uuid.UUID          `gorm:"type:uuid;not null;index:idx_api_keys_advertiser" json:"advertiser_id"`
	NetworkID    *uuid.UUID         `gorm:"type:uuid;index:idx_api_keys_network" json:"network_id,omitempty"`
	
	// Key info
	Name         string             `gorm:"size:100;not null" json:"name"`
	KeyPrefix    string             `gorm:"size:20;not null;index:idx_api_keys_prefix" json:"key_prefix"` // e.g., "afftok_live_sk_"
	KeyHash      string             `gorm:"size:255;not null" json:"-"` // Argon2/bcrypt hash - NEVER expose
	KeyHint      string             `gorm:"size:10" json:"key_hint"` // Last 4 chars for identification
	
	// Status
	Status       APIKeyStatus       `gorm:"size:20;default:'active';index:idx_api_keys_status" json:"status"`
	
	// Permissions & Security
	Permissions  datatypes.JSON     `gorm:"type:jsonb" json:"permissions,omitempty"` // e.g., {"postback:write", "stats:read"}
	AllowedIPs   datatypes.JSON     `gorm:"type:jsonb" json:"allowed_ips,omitempty"` // e.g., ["192.168.1.1", "10.0.0.0/8"]
	
	// Usage tracking
	UsageCount   int64              `gorm:"default:0" json:"usage_count"`
	LastUsedAt   *time.Time         `gorm:"index:idx_api_keys_last_used" json:"last_used_at,omitempty"`
	LastUsedIP   string             `gorm:"size:45" json:"last_used_ip,omitempty"`
	
	// Rate limiting
	RateLimitPerMinute int          `gorm:"default:60" json:"rate_limit_per_minute"`
	RateLimitBurst     int          `gorm:"default:10" json:"rate_limit_burst"`
	
	// Expiration
	ExpiresAt    *time.Time         `json:"expires_at,omitempty"`
	
	// Timestamps
	CreatedAt    time.Time          `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time          `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	RevokedAt    *time.Time         `json:"revoked_at,omitempty"`
	
	// Relations
	Advertiser   *AfftokUser        `gorm:"foreignKey:AdvertiserID" json:"advertiser,omitempty"`
	Network      *Network           `gorm:"foreignKey:NetworkID" json:"network,omitempty"`
}

// TableName returns the table name for GORM
func (AdvertiserAPIKey) TableName() string {
	return "advertiser_api_keys"
}

// IsActive returns true if the key is active and not expired
func (k *AdvertiserAPIKey) IsActive() bool {
	if k.Status != APIKeyStatusActive {
		return false
	}
	if k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt) {
		return false
	}
	return true
}

// IsIPAllowed checks if the given IP is allowed
func (k *AdvertiserAPIKey) IsIPAllowed(ip string) bool {
	if k.AllowedIPs == nil || len(k.AllowedIPs) == 0 {
		return true // No restrictions
	}
	
	var allowedIPs []string
	if err := k.AllowedIPs.UnmarshalJSON([]byte(k.AllowedIPs)); err != nil {
		return true // Parse error, allow
	}
	
	for _, allowedIP := range allowedIPs {
		if allowedIP == ip {
			return true
		}
		// TODO: Add CIDR matching for ranges like "10.0.0.0/8"
	}
	
	return false
}

// HasPermission checks if the key has a specific permission
func (k *AdvertiserAPIKey) HasPermission(permission string) bool {
	if k.Permissions == nil || len(k.Permissions) == 0 {
		return true // No restrictions = all permissions
	}
	
	var permissions []string
	if err := k.Permissions.UnmarshalJSON([]byte(k.Permissions)); err != nil {
		return true // Parse error, allow
	}
	
	for _, perm := range permissions {
		if perm == permission || perm == "*" {
			return true
		}
	}
	
	return false
}

// ============================================
// API KEY USAGE LOG
// ============================================

// APIKeyUsageLog tracks API key usage
type APIKeyUsageLog struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	APIKeyID    uuid.UUID  `gorm:"type:uuid;not null;index:idx_api_usage_key" json:"api_key_id"`
	AdvertiserID uuid.UUID `gorm:"type:uuid;not null;index:idx_api_usage_advertiser" json:"advertiser_id"`
	
	// Request info
	IP          string     `gorm:"size:45;index:idx_api_usage_ip" json:"ip"`
	Endpoint    string     `gorm:"size:255" json:"endpoint"`
	Method      string     `gorm:"size:10" json:"method"`
	UserAgent   string     `gorm:"type:text" json:"user_agent,omitempty"`
	
	// Result
	Success     bool       `gorm:"index:idx_api_usage_success" json:"success"`
	StatusCode  int        `json:"status_code"`
	ErrorReason string     `gorm:"size:255" json:"error_reason,omitempty"`
	
	// Performance
	LatencyMs   int64      `json:"latency_ms"`
	
	// Timestamp
	CreatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP;index:idx_api_usage_created" json:"created_at"`
}

// TableName returns the table name for GORM
func (APIKeyUsageLog) TableName() string {
	return "api_key_usage_logs"
}

// ============================================
// API KEY PERMISSIONS
// ============================================

// Common permissions
const (
	PermissionPostbackWrite = "postback:write"
	PermissionPostbackRead  = "postback:read"
	PermissionStatsRead     = "stats:read"
	PermissionOffersRead    = "offers:read"
	PermissionAllAccess     = "*"
)

// DefaultPermissions returns default permissions for new API keys
func DefaultPermissions() []string {
	return []string{
		PermissionPostbackWrite,
		PermissionPostbackRead,
	}
}

// ============================================
// API KEY CREATE/UPDATE DTOs
// ============================================

// CreateAPIKeyRequest represents a request to create an API key
type CreateAPIKeyRequest struct {
	Name               string   `json:"name" binding:"required,min=3,max=100"`
	NetworkID          string   `json:"network_id,omitempty"`
	Permissions        []string `json:"permissions,omitempty"`
	AllowedIPs         []string `json:"allowed_ips,omitempty"`
	RateLimitPerMinute int      `json:"rate_limit_per_minute,omitempty"`
	ExpiresInDays      int      `json:"expires_in_days,omitempty"` // 0 = no expiration
}

// CreateAPIKeyResponse represents the response after creating an API key
type CreateAPIKeyResponse struct {
	ID          uuid.UUID `json:"id"`
	Name        string    `json:"name"`
	PlaintextKey string   `json:"api_key"` // Only shown ONCE
	KeyHint     string    `json:"key_hint"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	Warning     string    `json:"warning"`
}

// APIKeyInfo represents API key info for listing (without sensitive data)
type APIKeyInfo struct {
	ID                 uuid.UUID    `json:"id"`
	AdvertiserID       uuid.UUID    `json:"advertiser_id"`
	AdvertiserName     string       `json:"advertiser_name,omitempty"`
	NetworkID          *uuid.UUID   `json:"network_id,omitempty"`
	NetworkName        string       `json:"network_name,omitempty"`
	Name               string       `json:"name"`
	KeyPrefix          string       `json:"key_prefix"`
	KeyHint            string       `json:"key_hint"`
	Status             APIKeyStatus `json:"status"`
	Permissions        []string     `json:"permissions,omitempty"`
	AllowedIPs         []string     `json:"allowed_ips,omitempty"`
	UsageCount         int64        `json:"usage_count"`
	LastUsedAt         *time.Time   `json:"last_used_at,omitempty"`
	LastUsedIP         string       `json:"last_used_ip,omitempty"`
	RateLimitPerMinute int          `json:"rate_limit_per_minute"`
	CreatedAt          time.Time    `json:"created_at"`
	ExpiresAt          *time.Time   `json:"expires_at,omitempty"`
}


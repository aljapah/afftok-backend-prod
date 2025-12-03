package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// ============================================
// TENANT MODEL
// ============================================

// TenantStatus represents tenant status
type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusPending   TenantStatus = "pending"
	TenantStatusDeleted   TenantStatus = "deleted"
)

// TenantPlan represents tenant subscription plan
type TenantPlan string

const (
	TenantPlanFree       TenantPlan = "free"
	TenantPlanPro        TenantPlan = "pro"
	TenantPlanEnterprise TenantPlan = "enterprise"
)

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name             string         `json:"name" gorm:"size:255;not null"`
	Slug             string         `json:"slug" gorm:"size:100;uniqueIndex;not null"`
	Status           TenantStatus   `json:"status" gorm:"size:20;default:'pending';index"`
	Plan             TenantPlan     `json:"plan" gorm:"size:20;default:'free';index"`
	
	// Contact & Admin
	AdminEmail       string         `json:"admin_email" gorm:"size:255"`
	AdminUserID      *uuid.UUID     `json:"admin_user_id,omitempty" gorm:"type:uuid"`
	
	// Settings
	Settings         datatypes.JSON `json:"settings,omitempty" gorm:"type:jsonb"`
	
	// Branding
	LogoURL          string         `json:"logo_url,omitempty" gorm:"size:500"`
	FaviconURL       string         `json:"favicon_url,omitempty" gorm:"size:500"`
	PrimaryColor     string         `json:"primary_color,omitempty" gorm:"size:20;default:'#3B82F6'"`
	SecondaryColor   string         `json:"secondary_color,omitempty" gorm:"size:20;default:'#1E40AF'"`
	
	// Domain Configuration
	AllowedDomains   datatypes.JSON `json:"allowed_domains,omitempty" gorm:"type:jsonb"`
	CustomDomain     string         `json:"custom_domain,omitempty" gorm:"size:255;index"`
	
	// Default Resources
	DefaultNetworkID *uuid.UUID     `json:"default_network_id,omitempty" gorm:"type:uuid"`
	
	// Limits (based on plan)
	MaxUsers         int            `json:"max_users" gorm:"default:10"`
	MaxOffers        int            `json:"max_offers" gorm:"default:50"`
	MaxClicksPerDay  int            `json:"max_clicks_per_day" gorm:"default:10000"`
	MaxAPIKeys       int            `json:"max_api_keys" gorm:"default:5"`
	MaxWebhooks      int            `json:"max_webhooks" gorm:"default:10"`
	
	// Feature Flags
	Features         datatypes.JSON `json:"features,omitempty" gorm:"type:jsonb"`
	
	// Billing
	BillingEmail     string         `json:"billing_email,omitempty" gorm:"size:255"`
	StripeCustomerID string         `json:"stripe_customer_id,omitempty" gorm:"size:100"`
	
	// Timestamps
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	SuspendedAt      *time.Time     `json:"suspended_at,omitempty"`
	DeletedAt        *time.Time     `json:"deleted_at,omitempty" gorm:"index"`
}

func (Tenant) TableName() string {
	return "tenants"
}

// ============================================
// TENANT SETTINGS
// ============================================

// TenantSettings represents tenant-specific settings
type TenantSettings struct {
	// Tracking Settings
	DefaultLinkTTL       int    `json:"default_link_ttl"`        // seconds
	AllowLegacyLinks     bool   `json:"allow_legacy_links"`
	EnableBotDetection   bool   `json:"enable_bot_detection"`
	EnableGeoRules       bool   `json:"enable_geo_rules"`
	EnableFraudDetection bool   `json:"enable_fraud_detection"`
	
	// Webhook Settings
	WebhookRetryCount    int    `json:"webhook_retry_count"`
	WebhookTimeoutMs     int    `json:"webhook_timeout_ms"`
	
	// API Settings
	APIRateLimitPerMin   int    `json:"api_rate_limit_per_min"`
	
	// Notification Settings
	NotifyOnConversion   bool   `json:"notify_on_conversion"`
	NotifyOnFraud        bool   `json:"notify_on_fraud"`
	NotificationEmail    string `json:"notification_email"`
	
	// Timezone
	Timezone             string `json:"timezone"`
}

// DefaultTenantSettings returns default settings for a new tenant
func DefaultTenantSettings() TenantSettings {
	return TenantSettings{
		DefaultLinkTTL:       86400,  // 24 hours
		AllowLegacyLinks:     true,
		EnableBotDetection:   true,
		EnableGeoRules:       true,
		EnableFraudDetection: true,
		WebhookRetryCount:    5,
		WebhookTimeoutMs:     30000,
		APIRateLimitPerMin:   60,
		NotifyOnConversion:   true,
		NotifyOnFraud:        true,
		Timezone:             "UTC",
	}
}

// ============================================
// TENANT FEATURES
// ============================================

// TenantFeatures represents feature flags for a tenant
type TenantFeatures struct {
	AdvancedAnalytics  bool `json:"advanced_analytics"`
	CustomBranding     bool `json:"custom_branding"`
	APIAccess          bool `json:"api_access"`
	WebhooksEnabled    bool `json:"webhooks_enabled"`
	GeoRulesEnabled    bool `json:"geo_rules_enabled"`
	TeamManagement     bool `json:"team_management"`
	WhiteLabelEnabled  bool `json:"white_label_enabled"`
	PrioritySupport    bool `json:"priority_support"`
	CustomDomain       bool `json:"custom_domain"`
	SSO                bool `json:"sso"`
}

// GetFeaturesForPlan returns features based on plan
func GetFeaturesForPlan(plan TenantPlan) TenantFeatures {
	switch plan {
	case TenantPlanFree:
		return TenantFeatures{
			AdvancedAnalytics: false,
			CustomBranding:    false,
			APIAccess:         true,
			WebhooksEnabled:   false,
			GeoRulesEnabled:   false,
			TeamManagement:    false,
			WhiteLabelEnabled: false,
			PrioritySupport:   false,
			CustomDomain:      false,
			SSO:               false,
		}
	case TenantPlanPro:
		return TenantFeatures{
			AdvancedAnalytics: true,
			CustomBranding:    true,
			APIAccess:         true,
			WebhooksEnabled:   true,
			GeoRulesEnabled:   true,
			TeamManagement:    true,
			WhiteLabelEnabled: false,
			PrioritySupport:   false,
			CustomDomain:      true,
			SSO:               false,
		}
	case TenantPlanEnterprise:
		return TenantFeatures{
			AdvancedAnalytics: true,
			CustomBranding:    true,
			APIAccess:         true,
			WebhooksEnabled:   true,
			GeoRulesEnabled:   true,
			TeamManagement:    true,
			WhiteLabelEnabled: true,
			PrioritySupport:   true,
			CustomDomain:      true,
			SSO:               true,
		}
	default:
		return GetFeaturesForPlan(TenantPlanFree)
	}
}

// GetLimitsForPlan returns resource limits based on plan
func GetLimitsForPlan(plan TenantPlan) (maxUsers, maxOffers, maxClicks, maxAPIKeys, maxWebhooks int) {
	switch plan {
	case TenantPlanFree:
		return 5, 10, 5000, 2, 0
	case TenantPlanPro:
		return 25, 100, 100000, 10, 20
	case TenantPlanEnterprise:
		return 1000, 10000, 10000000, 100, 500
	default:
		return GetLimitsForPlan(TenantPlanFree)
	}
}

// ============================================
// TENANT DOMAIN MAPPING
// ============================================

// TenantDomain represents a domain mapping for a tenant
type TenantDomain struct {
	ID         uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID   uuid.UUID  `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Domain     string     `json:"domain" gorm:"size:255;uniqueIndex;not null"`
	IsPrimary  bool       `json:"is_primary" gorm:"default:false"`
	IsVerified bool       `json:"is_verified" gorm:"default:false"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at" gorm:"autoCreateTime"`
	
	// Relations
	Tenant     *Tenant    `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

func (TenantDomain) TableName() string {
	return "tenant_domains"
}

// ============================================
// TENANT STATS
// ============================================

// TenantStats represents aggregated stats for a tenant
type TenantStats struct {
	TenantID          uuid.UUID `json:"tenant_id"`
	TotalUsers        int64     `json:"total_users"`
	TotalOffers       int64     `json:"total_offers"`
	TotalClicks       int64     `json:"total_clicks"`
	TotalConversions  int64     `json:"total_conversions"`
	TotalEarnings     int64     `json:"total_earnings"`
	TodayClicks       int64     `json:"today_clicks"`
	TodayConversions  int64     `json:"today_conversions"`
	ActiveAPIKeys     int64     `json:"active_api_keys"`
	ActiveWebhooks    int64     `json:"active_webhooks"`
	ActiveGeoRules    int64     `json:"active_geo_rules"`
	FraudBlocked      int64     `json:"fraud_blocked"`
	LastActivityAt    time.Time `json:"last_activity_at"`
}

// ============================================
// TENANT AUDIT LOG
// ============================================

// TenantAuditAction represents audit action types
type TenantAuditAction string

const (
	TenantAuditCreated       TenantAuditAction = "created"
	TenantAuditUpdated       TenantAuditAction = "updated"
	TenantAuditSuspended     TenantAuditAction = "suspended"
	TenantAuditActivated     TenantAuditAction = "activated"
	TenantAuditPlanChanged   TenantAuditAction = "plan_changed"
	TenantAuditDomainAdded   TenantAuditAction = "domain_added"
	TenantAuditDomainRemoved TenantAuditAction = "domain_removed"
	TenantAuditSettingsChanged TenantAuditAction = "settings_changed"
)

// TenantAuditLog represents an audit log entry for tenant changes
type TenantAuditLog struct {
	ID        uuid.UUID         `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TenantID  uuid.UUID         `json:"tenant_id" gorm:"type:uuid;not null;index"`
	Action    TenantAuditAction `json:"action" gorm:"size:50;not null;index"`
	ActorID   *uuid.UUID        `json:"actor_id,omitempty" gorm:"type:uuid"`
	ActorType string            `json:"actor_type" gorm:"size:20"` // admin, system, api
	OldValue  datatypes.JSON    `json:"old_value,omitempty" gorm:"type:jsonb"`
	NewValue  datatypes.JSON    `json:"new_value,omitempty" gorm:"type:jsonb"`
	IPAddress string            `json:"ip_address,omitempty" gorm:"size:45"`
	UserAgent string            `json:"user_agent,omitempty" gorm:"size:500"`
	CreatedAt time.Time         `json:"created_at" gorm:"autoCreateTime;index"`
	
	// Relations
	Tenant    *Tenant           `json:"tenant,omitempty" gorm:"foreignKey:TenantID"`
}

func (TenantAuditLog) TableName() string {
	return "tenant_audit_logs"
}

// ============================================
// TENANT-SCOPED INTERFACE
// ============================================

// TenantScoped interface for models that belong to a tenant
type TenantScoped interface {
	GetTenantID() uuid.UUID
	SetTenantID(tenantID uuid.UUID)
}

// ============================================
// DEFAULT TENANT (for migration)
// ============================================

// DefaultTenantID is the ID for the default tenant (used for migration)
var DefaultTenantID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// DefaultTenant returns the default tenant for migration purposes
func DefaultTenant() *Tenant {
	return &Tenant{
		ID:              DefaultTenantID,
		Name:            "Default Tenant",
		Slug:            "default",
		Status:          TenantStatusActive,
		Plan:            TenantPlanEnterprise,
		MaxUsers:        1000000,
		MaxOffers:       1000000,
		MaxClicksPerDay: 100000000,
		MaxAPIKeys:      10000,
		MaxWebhooks:     10000,
	}
}


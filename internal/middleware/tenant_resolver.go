package middleware

import (
	"context"
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
	"gorm.io/gorm"
)

// ============================================
// TENANT CONTEXT KEYS
// ============================================

const (
	TenantIDKey      = "tenant_id"
	TenantKey        = "tenant"
	TenantSlugKey    = "tenant_slug"
	IsSuperAdminKey  = "is_super_admin"
)

// ============================================
// TENANT RESOLVER MIDDLEWARE
// ============================================

var (
	tenantService *services.TenantService
	tenantDB      *gorm.DB
	tenantOnce    sync.Once
)

// InitTenantMiddleware initializes the tenant middleware with dependencies
func InitTenantMiddleware(db *gorm.DB) {
	tenantOnce.Do(func() {
		tenantDB = db
		tenantService = services.NewTenantService(db)
	})
}

// TenantResolverMiddleware resolves tenant from request
func TenantResolverMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip for public endpoints
		if isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Try to resolve tenant
		tenant, err := resolveTenant(c)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "Tenant resolution failed",
				"details": err.Error(),
			})
			return
		}

		// Check if tenant is active
		if tenant.Status == models.TenantStatusSuspended {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Tenant is suspended",
				"code":  "TENANT_SUSPENDED",
			})
			return
		}

		if tenant.Status == models.TenantStatusDeleted {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "Tenant not found",
				"code":  "TENANT_NOT_FOUND",
			})
			return
		}

		// Attach tenant to context
		c.Set(TenantIDKey, tenant.ID)
		c.Set(TenantKey, tenant)
		c.Set(TenantSlugKey, tenant.Slug)

		// Track tenant activity
		go trackTenantActivity(tenant.ID)

		c.Next()
	}
}

// resolveTenant resolves tenant from various sources
func resolveTenant(c *gin.Context) (*models.Tenant, error) {
	// 1. Try X-Tenant-ID header
	if tenantID := c.GetHeader("X-Tenant-ID"); tenantID != "" {
		id, err := uuid.Parse(tenantID)
		if err != nil {
			return nil, fmt.Errorf("invalid tenant ID format")
		}
		return tenantService.GetTenant(id)
	}

	// 2. Try X-Tenant-Slug header
	if slug := c.GetHeader("X-Tenant-Slug"); slug != "" {
		return tenantService.GetTenantBySlug(slug)
	}

	// 3. Try domain/subdomain resolution
	host := c.Request.Host
	if host != "" {
		// Remove port if present
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}

		// Try subdomain resolution (e.g., company1.afftok.com)
		parts := strings.Split(host, ".")
		if len(parts) >= 3 {
			subdomain := parts[0]
			tenant, err := tenantService.GetTenantBySlug(subdomain)
			if err == nil {
				return tenant, nil
			}
		}

		// Try custom domain resolution
		tenant, err := tenantService.GetTenantByDomain(host)
		if err == nil {
			return tenant, nil
		}
	}

	// 4. Try from API Key (if already resolved by API Key middleware)
	if tenantID, exists := c.Get("api_key_tenant_id"); exists {
		if id, ok := tenantID.(uuid.UUID); ok {
			return tenantService.GetTenant(id)
		}
	}

	// 5. Try from JWT claims (if already authenticated)
	if tenantID, exists := c.Get("jwt_tenant_id"); exists {
		if id, ok := tenantID.(uuid.UUID); ok {
			return tenantService.GetTenant(id)
		}
	}

	// 6. Default tenant (for backward compatibility during migration)
	return tenantService.GetTenant(models.DefaultTenantID)
}

// isPublicEndpoint checks if the endpoint is public
func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/api/auth/login",
		"/api/auth/register",
		"/api/c/", // Click tracking (tenant resolved from tracking code)
		"/api/postback",
		"/api/internal/",
	}

	for _, p := range publicPaths {
		if strings.HasPrefix(path, p) {
			return true
		}
	}

	return false
}

// ============================================
// TENANT RATE LIMITING
// ============================================

// TenantRateLimitMiddleware applies rate limits per tenant
func TenantRateLimitMiddleware(requestsPerMinute int) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get(TenantIDKey)
		if !exists {
			c.Next()
			return
		}

		tid := tenantID.(uuid.UUID)
		key := fmt.Sprintf("tenant_ratelimit:%s:%d", tid.String(), time.Now().Minute())

		ctx := context.Background()
		count, err := cache.Increment(ctx, key)
		if err == nil && count == 1 {
			// Set expiry on first request
			cache.Expire(ctx, key, 2*time.Minute)
		}

		if count > int64(requestsPerMinute) {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Tenant rate limit exceeded",
				"code":  "TENANT_RATE_LIMIT",
			})
			return
		}

		c.Next()
	}
}

// ============================================
// TENANT CLICK LIMIT MIDDLEWARE
// ============================================

// TenantClickLimitMiddleware checks tenant's daily click limit
func TenantClickLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tenantID, exists := c.Get(TenantIDKey)
		if !exists {
			c.Next()
			return
		}

		tid := tenantID.(uuid.UUID)
		
		allowed, _, err := tenantService.CheckDailyClickLimit(tid)
		if err != nil {
			c.Next()
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Daily click limit exceeded for this tenant",
				"code":  "CLICK_LIMIT_EXCEEDED",
			})
			return
		}

		c.Next()
	}
}

// ============================================
// SUPER ADMIN BYPASS
// ============================================

// SuperAdminBypassMiddleware allows super admins to access any tenant
func SuperAdminBypassMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is super admin
		isSuperAdmin, exists := c.Get(IsSuperAdminKey)
		if exists && isSuperAdmin.(bool) {
			// Allow super admin to specify any tenant
			if tenantID := c.GetHeader("X-Admin-Tenant-ID"); tenantID != "" {
				id, err := uuid.Parse(tenantID)
				if err == nil {
					tenant, err := tenantService.GetTenant(id)
					if err == nil {
						c.Set(TenantIDKey, tenant.ID)
						c.Set(TenantKey, tenant)
						c.Set(TenantSlugKey, tenant.Slug)
					}
				}
			}
		}

		c.Next()
	}
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// GetTenantID gets tenant ID from context
func GetTenantID(c *gin.Context) uuid.UUID {
	if tenantID, exists := c.Get(TenantIDKey); exists {
		if id, ok := tenantID.(uuid.UUID); ok {
			return id
		}
	}
	return models.DefaultTenantID
}

// GetTenant gets tenant from context
func GetTenant(c *gin.Context) *models.Tenant {
	if tenant, exists := c.Get(TenantKey); exists {
		if t, ok := tenant.(*models.Tenant); ok {
			return t
		}
	}
	return nil
}

// MustGetTenantID gets tenant ID or panics
func MustGetTenantID(c *gin.Context) uuid.UUID {
	tenantID := GetTenantID(c)
	if tenantID == uuid.Nil {
		panic("tenant ID not found in context")
	}
	return tenantID
}

// IsSuperAdmin checks if current user is super admin
func IsSuperAdmin(c *gin.Context) bool {
	if isSuperAdmin, exists := c.Get(IsSuperAdminKey); exists {
		return isSuperAdmin.(bool)
	}
	return false
}

// ============================================
// TENANT METRICS
// ============================================

var tenantMetrics = struct {
	sync.RWMutex
	RequestCounts map[string]*int64
}{
	RequestCounts: make(map[string]*int64),
}

// trackTenantActivity tracks tenant activity metrics
func trackTenantActivity(tenantID uuid.UUID) {
	key := tenantID.String()

	tenantMetrics.Lock()
	if tenantMetrics.RequestCounts[key] == nil {
		var count int64
		tenantMetrics.RequestCounts[key] = &count
	}
	tenantMetrics.Unlock()

	tenantMetrics.RLock()
	atomic.AddInt64(tenantMetrics.RequestCounts[key], 1)
	tenantMetrics.RUnlock()

	// Also update Redis for persistence
	ctx := context.Background()
	redisKey := fmt.Sprintf("tenant_activity:%s:%s", tenantID.String(), time.Now().Format("2006-01-02"))
	cache.Increment(ctx, redisKey)
	cache.Expire(ctx, redisKey, 48*time.Hour)
}

// GetTenantRequestCount gets request count for a tenant
func GetTenantRequestCount(tenantID uuid.UUID) int64 {
	key := tenantID.String()

	tenantMetrics.RLock()
	defer tenantMetrics.RUnlock()

	if count := tenantMetrics.RequestCounts[key]; count != nil {
		return atomic.LoadInt64(count)
	}
	return 0
}

// ============================================
// TENANT FEATURE CHECK MIDDLEWARE
// ============================================

// RequireFeature creates a middleware that checks if tenant has a feature
func RequireFeature(feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tenant := GetTenant(c)
		if tenant == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Tenant not found",
			})
			return
		}

		// Parse features
		var features models.TenantFeatures
		if tenant.Features != nil {
			if err := tenant.Features.Scan(&features); err != nil {
				// If parsing fails, use plan defaults
				features = models.GetFeaturesForPlan(tenant.Plan)
			}
		} else {
			features = models.GetFeaturesForPlan(tenant.Plan)
		}

		// Check feature
		hasFeature := false
		switch feature {
		case "advanced_analytics":
			hasFeature = features.AdvancedAnalytics
		case "custom_branding":
			hasFeature = features.CustomBranding
		case "api_access":
			hasFeature = features.APIAccess
		case "webhooks":
			hasFeature = features.WebhooksEnabled
		case "geo_rules":
			hasFeature = features.GeoRulesEnabled
		case "team_management":
			hasFeature = features.TeamManagement
		case "white_label":
			hasFeature = features.WhiteLabelEnabled
		case "custom_domain":
			hasFeature = features.CustomDomain
		case "sso":
			hasFeature = features.SSO
		}

		if !hasFeature {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "Feature not available on your plan",
				"feature": feature,
				"plan":    tenant.Plan,
				"code":    "FEATURE_NOT_AVAILABLE",
			})
			return
		}

		c.Next()
	}
}


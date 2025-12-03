package services

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// TENANT SERVICE
// ============================================

// TenantService handles tenant management
type TenantService struct {
	db            *gorm.DB
	cache         sync.Map // In-memory cache for tenant lookups
	observability *ObservabilityService
}

// NewTenantService creates a new tenant service
func NewTenantService(db *gorm.DB) *TenantService {
	return &TenantService{
		db:            db,
		observability: NewObservabilityService(),
	}
}

// ============================================
// TENANT CRUD
// ============================================

// CreateTenant creates a new tenant
func (s *TenantService) CreateTenant(tenant *models.Tenant) error {
	if tenant.ID == uuid.Nil {
		tenant.ID = uuid.New()
	}

	// Validate slug
	if err := s.validateSlug(tenant.Slug); err != nil {
		return err
	}

	// Set defaults based on plan
	maxUsers, maxOffers, maxClicks, maxAPIKeys, maxWebhooks := models.GetLimitsForPlan(tenant.Plan)
	tenant.MaxUsers = maxUsers
	tenant.MaxOffers = maxOffers
	tenant.MaxClicksPerDay = maxClicks
	tenant.MaxAPIKeys = maxAPIKeys
	tenant.MaxWebhooks = maxWebhooks

	// Set default features
	features := models.GetFeaturesForPlan(tenant.Plan)
	featuresJSON, _ := json.Marshal(features)
	tenant.Features = featuresJSON

	// Set default settings
	settings := models.DefaultTenantSettings()
	settingsJSON, _ := json.Marshal(settings)
	tenant.Settings = settingsJSON

	if err := s.db.Create(tenant).Error; err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}

	// Log audit
	s.logAudit(tenant.ID, models.TenantAuditCreated, nil, nil, tenant)

	// Invalidate cache
	s.invalidateCache(tenant.ID)

	return nil
}

// GetTenant gets a tenant by ID
func (s *TenantService) GetTenant(tenantID uuid.UUID) (*models.Tenant, error) {
	// Try cache first
	if cached, ok := s.cache.Load(tenantID.String()); ok {
		return cached.(*models.Tenant), nil
	}

	// Try Redis
	tenant, err := s.getTenantFromRedis(tenantID)
	if err == nil && tenant != nil {
		s.cache.Store(tenantID.String(), tenant)
		return tenant, nil
	}

	// Load from DB
	tenant = &models.Tenant{}
	if err := s.db.First(tenant, "id = ?", tenantID).Error; err != nil {
		return nil, err
	}

	// Cache it
	s.cacheTenant(tenant)

	return tenant, nil
}

// GetTenantBySlug gets a tenant by slug
func (s *TenantService) GetTenantBySlug(slug string) (*models.Tenant, error) {
	cacheKey := "slug:" + slug

	// Try cache first
	if cached, ok := s.cache.Load(cacheKey); ok {
		return cached.(*models.Tenant), nil
	}

	// Load from DB
	tenant := &models.Tenant{}
	if err := s.db.First(tenant, "slug = ?", slug).Error; err != nil {
		return nil, err
	}

	// Cache it
	s.cache.Store(cacheKey, tenant)
	s.cacheTenant(tenant)

	return tenant, nil
}

// GetTenantByDomain gets a tenant by domain
func (s *TenantService) GetTenantByDomain(domain string) (*models.Tenant, error) {
	cacheKey := "domain:" + strings.ToLower(domain)

	// Try cache first
	if cached, ok := s.cache.Load(cacheKey); ok {
		return cached.(*models.Tenant), nil
	}

	// Try custom domain first
	tenant := &models.Tenant{}
	if err := s.db.First(tenant, "custom_domain = ?", domain).Error; err == nil {
		s.cache.Store(cacheKey, tenant)
		s.cacheTenant(tenant)
		return tenant, nil
	}

	// Try domain mapping
	var tenantDomain models.TenantDomain
	if err := s.db.Preload("Tenant").First(&tenantDomain, "domain = ?", domain).Error; err != nil {
		return nil, err
	}

	s.cache.Store(cacheKey, tenantDomain.Tenant)
	s.cacheTenant(tenantDomain.Tenant)

	return tenantDomain.Tenant, nil
}

// UpdateTenant updates a tenant
func (s *TenantService) UpdateTenant(tenant *models.Tenant) error {
	// Get old value for audit
	oldTenant, _ := s.GetTenant(tenant.ID)

	if err := s.db.Save(tenant).Error; err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}

	// Log audit
	s.logAudit(tenant.ID, models.TenantAuditUpdated, nil, oldTenant, tenant)

	// Invalidate cache
	s.invalidateCache(tenant.ID)

	return nil
}

// DeleteTenant soft-deletes a tenant
func (s *TenantService) DeleteTenant(tenantID uuid.UUID) error {
	now := time.Now()
	if err := s.db.Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Updates(map[string]interface{}{
			"status":     models.TenantStatusDeleted,
			"deleted_at": &now,
		}).Error; err != nil {
		return err
	}

	s.invalidateCache(tenantID)
	return nil
}

// ============================================
// TENANT STATUS MANAGEMENT
// ============================================

// SuspendTenant suspends a tenant
func (s *TenantService) SuspendTenant(tenantID uuid.UUID, reason string) error {
	now := time.Now()
	if err := s.db.Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Updates(map[string]interface{}{
			"status":       models.TenantStatusSuspended,
			"suspended_at": &now,
		}).Error; err != nil {
		return err
	}

	// Log audit
	s.logAudit(tenantID, models.TenantAuditSuspended, nil, nil, map[string]string{"reason": reason})

	s.invalidateCache(tenantID)
	return nil
}

// ActivateTenant activates a tenant
func (s *TenantService) ActivateTenant(tenantID uuid.UUID) error {
	if err := s.db.Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Updates(map[string]interface{}{
			"status":       models.TenantStatusActive,
			"suspended_at": nil,
		}).Error; err != nil {
		return err
	}

	// Log audit
	s.logAudit(tenantID, models.TenantAuditActivated, nil, nil, nil)

	s.invalidateCache(tenantID)
	return nil
}

// IsTenantActive checks if a tenant is active
func (s *TenantService) IsTenantActive(tenantID uuid.UUID) bool {
	tenant, err := s.GetTenant(tenantID)
	if err != nil {
		return false
	}
	return tenant.Status == models.TenantStatusActive
}

// ============================================
// TENANT PLAN MANAGEMENT
// ============================================

// ChangePlan changes a tenant's plan
func (s *TenantService) ChangePlan(tenantID uuid.UUID, newPlan models.TenantPlan) error {
	tenant, err := s.GetTenant(tenantID)
	if err != nil {
		return err
	}

	oldPlan := tenant.Plan
	tenant.Plan = newPlan

	// Update limits
	maxUsers, maxOffers, maxClicks, maxAPIKeys, maxWebhooks := models.GetLimitsForPlan(newPlan)
	tenant.MaxUsers = maxUsers
	tenant.MaxOffers = maxOffers
	tenant.MaxClicksPerDay = maxClicks
	tenant.MaxAPIKeys = maxAPIKeys
	tenant.MaxWebhooks = maxWebhooks

	// Update features
	features := models.GetFeaturesForPlan(newPlan)
	featuresJSON, _ := json.Marshal(features)
	tenant.Features = featuresJSON

	if err := s.db.Save(tenant).Error; err != nil {
		return err
	}

	// Log audit
	s.logAudit(tenantID, models.TenantAuditPlanChanged, nil,
		map[string]string{"old_plan": string(oldPlan)},
		map[string]string{"new_plan": string(newPlan)})

	s.invalidateCache(tenantID)
	return nil
}

// ============================================
// TENANT DOMAIN MANAGEMENT
// ============================================

// AddDomain adds a domain to a tenant
func (s *TenantService) AddDomain(tenantID uuid.UUID, domain string, isPrimary bool) error {
	tenantDomain := &models.TenantDomain{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Domain:    strings.ToLower(domain),
		IsPrimary: isPrimary,
	}

	if err := s.db.Create(tenantDomain).Error; err != nil {
		return err
	}

	// Log audit
	s.logAudit(tenantID, models.TenantAuditDomainAdded, nil, nil, map[string]string{"domain": domain})

	s.invalidateCache(tenantID)
	return nil
}

// RemoveDomain removes a domain from a tenant
func (s *TenantService) RemoveDomain(tenantID uuid.UUID, domain string) error {
	if err := s.db.Where("tenant_id = ? AND domain = ?", tenantID, domain).
		Delete(&models.TenantDomain{}).Error; err != nil {
		return err
	}

	// Log audit
	s.logAudit(tenantID, models.TenantAuditDomainRemoved, nil, nil, map[string]string{"domain": domain})

	s.invalidateCache(tenantID)
	return nil
}

// GetDomains gets all domains for a tenant
func (s *TenantService) GetDomains(tenantID uuid.UUID) ([]models.TenantDomain, error) {
	var domains []models.TenantDomain
	err := s.db.Where("tenant_id = ?", tenantID).Find(&domains).Error
	return domains, err
}

// VerifyDomain marks a domain as verified
func (s *TenantService) VerifyDomain(tenantID uuid.UUID, domain string) error {
	now := time.Now()
	return s.db.Model(&models.TenantDomain{}).
		Where("tenant_id = ? AND domain = ?", tenantID, domain).
		Updates(map[string]interface{}{
			"is_verified": true,
			"verified_at": &now,
		}).Error
}

// ============================================
// TENANT SETTINGS
// ============================================

// GetSettings gets tenant settings
func (s *TenantService) GetSettings(tenantID uuid.UUID) (*models.TenantSettings, error) {
	tenant, err := s.GetTenant(tenantID)
	if err != nil {
		return nil, err
	}

	var settings models.TenantSettings
	if tenant.Settings != nil {
		json.Unmarshal(tenant.Settings, &settings)
	} else {
		settings = models.DefaultTenantSettings()
	}

	return &settings, nil
}

// UpdateSettings updates tenant settings
func (s *TenantService) UpdateSettings(tenantID uuid.UUID, settings *models.TenantSettings) error {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return err
	}

	if err := s.db.Model(&models.Tenant{}).
		Where("id = ?", tenantID).
		Update("settings", settingsJSON).Error; err != nil {
		return err
	}

	// Log audit
	s.logAudit(tenantID, models.TenantAuditSettingsChanged, nil, nil, settings)

	s.invalidateCache(tenantID)
	return nil
}

// ============================================
// TENANT STATS
// ============================================

// GetStats gets aggregated stats for a tenant
func (s *TenantService) GetStats(tenantID uuid.UUID) (*models.TenantStats, error) {
	stats := &models.TenantStats{
		TenantID: tenantID,
	}

	// Count users
	s.db.Model(&models.AfftokUser{}).Where("tenant_id = ?", tenantID).Count(&stats.TotalUsers)

	// Count offers
	s.db.Model(&models.Offer{}).Where("tenant_id = ?", tenantID).Count(&stats.TotalOffers)

	// Count clicks
	s.db.Model(&models.Click{}).Where("tenant_id = ?", tenantID).Count(&stats.TotalClicks)

	// Count conversions
	s.db.Model(&models.Conversion{}).Where("tenant_id = ?", tenantID).Count(&stats.TotalConversions)

	// Today's clicks
	today := time.Now().Truncate(24 * time.Hour)
	s.db.Model(&models.Click{}).
		Where("tenant_id = ? AND clicked_at >= ?", tenantID, today).
		Count(&stats.TodayClicks)

	// Today's conversions
	s.db.Model(&models.Conversion{}).
		Where("tenant_id = ? AND converted_at >= ?", tenantID, today).
		Count(&stats.TodayConversions)

	// Active API keys
	s.db.Model(&models.AdvertiserAPIKey{}).
		Where("tenant_id = ? AND status = ?", tenantID, "active").
		Count(&stats.ActiveAPIKeys)

	// Active webhooks
	s.db.Model(&models.WebhookPipeline{}).
		Where("tenant_id = ? AND status = ?", tenantID, "active").
		Count(&stats.ActiveWebhooks)

	// Active geo rules
	s.db.Model(&models.GeoRule{}).
		Where("tenant_id = ? AND status = ?", tenantID, "active").
		Count(&stats.ActiveGeoRules)

	stats.LastActivityAt = time.Now()

	return stats, nil
}

// ============================================
// TENANT LISTING
// ============================================

// ListTenants lists all tenants with filters
func (s *TenantService) ListTenants(
	status *models.TenantStatus,
	plan *models.TenantPlan,
	search string,
	limit, offset int,
) ([]models.Tenant, int64, error) {
	query := s.db.Model(&models.Tenant{}).Where("deleted_at IS NULL")

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if plan != nil {
		query = query.Where("plan = ?", *plan)
	}

	if search != "" {
		query = query.Where("name ILIKE ? OR slug ILIKE ? OR admin_email ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	var total int64
	query.Count(&total)

	var tenants []models.Tenant
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&tenants).Error

	return tenants, total, err
}

// ============================================
// LIMIT CHECKS
// ============================================

// CheckUserLimit checks if tenant can add more users
func (s *TenantService) CheckUserLimit(tenantID uuid.UUID) (bool, error) {
	tenant, err := s.GetTenant(tenantID)
	if err != nil {
		return false, err
	}

	var count int64
	s.db.Model(&models.AfftokUser{}).Where("tenant_id = ?", tenantID).Count(&count)

	return int(count) < tenant.MaxUsers, nil
}

// CheckOfferLimit checks if tenant can add more offers
func (s *TenantService) CheckOfferLimit(tenantID uuid.UUID) (bool, error) {
	tenant, err := s.GetTenant(tenantID)
	if err != nil {
		return false, err
	}

	var count int64
	s.db.Model(&models.Offer{}).Where("tenant_id = ?", tenantID).Count(&count)

	return int(count) < tenant.MaxOffers, nil
}

// CheckDailyClickLimit checks if tenant has exceeded daily click limit
func (s *TenantService) CheckDailyClickLimit(tenantID uuid.UUID) (bool, int64, error) {
	tenant, err := s.GetTenant(tenantID)
	if err != nil {
		return false, 0, err
	}

	today := time.Now().Truncate(24 * time.Hour)
	var count int64
	s.db.Model(&models.Click{}).
		Where("tenant_id = ? AND clicked_at >= ?", tenantID, today).
		Count(&count)

	return int(count) < tenant.MaxClicksPerDay, count, nil
}

// ============================================
// CACHING
// ============================================

func (s *TenantService) cacheTenant(tenant *models.Tenant) {
	s.cache.Store(tenant.ID.String(), tenant)
	s.cache.Store("slug:"+tenant.Slug, tenant)

	// Cache in Redis
	ctx := context.Background()
	key := fmt.Sprintf("tenant:%s", tenant.ID.String())
	data, _ := json.Marshal(tenant)
	cache.Set(ctx, key, string(data), 5*time.Minute)
}

func (s *TenantService) getTenantFromRedis(tenantID uuid.UUID) (*models.Tenant, error) {
	ctx := context.Background()
	key := fmt.Sprintf("tenant:%s", tenantID.String())
	
	data, err := cache.Get(ctx, key)
	if err != nil || data == "" {
		return nil, fmt.Errorf("not found")
	}

	var tenant models.Tenant
	if err := json.Unmarshal([]byte(data), &tenant); err != nil {
		return nil, err
	}

	return &tenant, nil
}

func (s *TenantService) invalidateCache(tenantID uuid.UUID) {
	// Clear in-memory cache
	s.cache.Delete(tenantID.String())

	// Clear Redis
	ctx := context.Background()
	key := fmt.Sprintf("tenant:%s", tenantID.String())
	cache.Delete(ctx, key)
}

// ============================================
// VALIDATION
// ============================================

func (s *TenantService) validateSlug(slug string) error {
	if slug == "" {
		return fmt.Errorf("slug is required")
	}

	if len(slug) < 3 || len(slug) > 50 {
		return fmt.Errorf("slug must be between 3 and 50 characters")
	}

	// Only lowercase alphanumeric and hyphens
	matched, _ := regexp.MatchString(`^[a-z0-9][a-z0-9-]*[a-z0-9]$`, slug)
	if !matched {
		return fmt.Errorf("slug must contain only lowercase letters, numbers, and hyphens")
	}

	// Check uniqueness
	var count int64
	s.db.Model(&models.Tenant{}).Where("slug = ?", slug).Count(&count)
	if count > 0 {
		return fmt.Errorf("slug already exists")
	}

	return nil
}

// ============================================
// AUDIT LOGGING
// ============================================

func (s *TenantService) logAudit(tenantID uuid.UUID, action models.TenantAuditAction, actorID *uuid.UUID, oldValue, newValue interface{}) {
	var oldJSON, newJSON []byte
	if oldValue != nil {
		oldJSON, _ = json.Marshal(oldValue)
	}
	if newValue != nil {
		newJSON, _ = json.Marshal(newValue)
	}

	log := &models.TenantAuditLog{
		ID:        uuid.New(),
		TenantID:  tenantID,
		Action:    action,
		ActorID:   actorID,
		ActorType: "system",
		OldValue:  oldJSON,
		NewValue:  newJSON,
	}

	s.db.Create(log)
}

// GetAuditLogs gets audit logs for a tenant
func (s *TenantService) GetAuditLogs(tenantID uuid.UUID, limit int) ([]models.TenantAuditLog, error) {
	var logs []models.TenantAuditLog
	err := s.db.Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// ============================================
// GLOBAL INSTANCE
// ============================================

var (
	tenantServiceInstance *TenantService
	tenantServiceOnce     sync.Once
)

// GetTenantService returns the global tenant service instance
func GetTenantService(db *gorm.DB) *TenantService {
	tenantServiceOnce.Do(func() {
		tenantServiceInstance = NewTenantService(db)
	})
	return tenantServiceInstance
}


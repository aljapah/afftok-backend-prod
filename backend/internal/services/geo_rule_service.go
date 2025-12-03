package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ============================================
// GEO RULE SERVICE
// ============================================

// GeoRuleService handles geo rule operations
type GeoRuleService struct {
	db    *gorm.DB
	mutex sync.RWMutex
}

// NewGeoRuleService creates a new geo rule service
func NewGeoRuleService(db *gorm.DB) *GeoRuleService {
	return &GeoRuleService{db: db}
}

// Cache configuration
const (
	geoRuleCacheTTL     = 10 * time.Minute
	geoRuleCachePrefix  = "georules:"
)

// ============================================
// GEO RULE METRICS
// ============================================

// GeoRuleMetrics tracks geo rule metrics
type GeoRuleMetrics struct {
	TotalChecks      int64
	BlockedByRule    int64
	AllowedByRule    int64
	NoRuleApplied    int64
	CacheHits        int64
	CacheMisses      int64
}

var geoMetrics = &GeoRuleMetrics{}

// GetGeoRuleMetrics returns geo rule metrics
func GetGeoRuleMetrics() *GeoRuleMetrics {
	return &GeoRuleMetrics{
		TotalChecks:   atomic.LoadInt64(&geoMetrics.TotalChecks),
		BlockedByRule: atomic.LoadInt64(&geoMetrics.BlockedByRule),
		AllowedByRule: atomic.LoadInt64(&geoMetrics.AllowedByRule),
		NoRuleApplied: atomic.LoadInt64(&geoMetrics.NoRuleApplied),
		CacheHits:     atomic.LoadInt64(&geoMetrics.CacheHits),
		CacheMisses:   atomic.LoadInt64(&geoMetrics.CacheMisses),
	}
}

// ============================================
// CORE GEO CHECK LOGIC
// ============================================

// GetEffectiveGeoRule checks if a country is allowed for a given offer/advertiser
// Resolution order: Offer -> Advertiser -> Global -> Default Allow
func (s *GeoRuleService) GetEffectiveGeoRule(offerID, advertiserID *uuid.UUID, countryCode string) *models.GeoCheckResult {
	atomic.AddInt64(&geoMetrics.TotalChecks, 1)
	
	// Normalize country code
	countryCode = strings.ToUpper(strings.TrimSpace(countryCode))
	if countryCode == "" {
		// No country = allow (can't block unknown)
		return &models.GeoCheckResult{
			Allowed: true,
			Reason:  "no_country",
		}
	}

	ctx := context.Background()

	// 1. Check offer-level rule
	if offerID != nil {
		rule := s.getRule(ctx, models.GeoRuleScopeOffer, *offerID)
		if rule != nil && rule.IsActive() {
			result := s.matchCountry(rule, countryCode)
			if !result.Allowed {
				atomic.AddInt64(&geoMetrics.BlockedByRule, 1)
			} else {
				atomic.AddInt64(&geoMetrics.AllowedByRule, 1)
			}
			return result
		}
	}

	// 2. Check advertiser-level rule
	if advertiserID != nil {
		rule := s.getRule(ctx, models.GeoRuleScopeAdvertiser, *advertiserID)
		if rule != nil && rule.IsActive() {
			result := s.matchCountry(rule, countryCode)
			if !result.Allowed {
				atomic.AddInt64(&geoMetrics.BlockedByRule, 1)
			} else {
				atomic.AddInt64(&geoMetrics.AllowedByRule, 1)
			}
			return result
		}
	}

	// 3. Check global rule
	globalRule := s.getGlobalRule(ctx)
	if globalRule != nil && globalRule.IsActive() {
		result := s.matchCountry(globalRule, countryCode)
		if !result.Allowed {
			atomic.AddInt64(&geoMetrics.BlockedByRule, 1)
		} else {
			atomic.AddInt64(&geoMetrics.AllowedByRule, 1)
		}
		return result
	}

	// 4. Default: Allow
	atomic.AddInt64(&geoMetrics.NoRuleApplied, 1)
	return &models.GeoCheckResult{
		Allowed: true,
		Reason:  "no_rule",
	}
}

// matchCountry checks if a country matches a rule
func (s *GeoRuleService) matchCountry(rule *models.GeoRule, countryCode string) *models.GeoCheckResult {
	countries := s.parseCountries(rule.Countries)
	isInList := false
	
	for _, c := range countries {
		if strings.EqualFold(c, countryCode) {
			isInList = true
			break
		}
	}

	result := &models.GeoCheckResult{
		Rule: rule,
	}

	switch rule.Mode {
	case models.GeoRuleModeAllow:
		// Allow mode: country must be in list
		if isInList {
			result.Allowed = true
			result.Reason = "allowed"
		} else {
			result.Allowed = false
			result.Reason = "blocked_by_rule"
		}
	case models.GeoRuleModeBlock:
		// Block mode: country must NOT be in list
		if isInList {
			result.Allowed = false
			result.Reason = "blocked_by_rule"
		} else {
			result.Allowed = true
			result.Reason = "allowed"
		}
	default:
		// Unknown mode = allow
		result.Allowed = true
		result.Reason = "unknown_mode"
	}

	return result
}

// parseCountries parses JSON countries array
func (s *GeoRuleService) parseCountries(data datatypes.JSON) []string {
	if data == nil || len(data) == 0 {
		return []string{}
	}
	
	var countries []string
	if err := json.Unmarshal(data, &countries); err != nil {
		return []string{}
	}
	return countries
}

// ============================================
// CACHE OPERATIONS
// ============================================

// getRule gets a rule from cache or database
func (s *GeoRuleService) getRule(ctx context.Context, scopeType models.GeoRuleScopeType, scopeID uuid.UUID) *models.GeoRule {
	cacheKey := fmt.Sprintf("%s%s:%s", geoRuleCachePrefix, scopeType, scopeID.String())
	
	// Try cache first
	if cached, err := cache.Get(ctx, cacheKey); err == nil && cached != "" {
		atomic.AddInt64(&geoMetrics.CacheHits, 1)
		var rule models.GeoRule
		if err := json.Unmarshal([]byte(cached), &rule); err == nil {
			return &rule
		}
	}
	
	atomic.AddInt64(&geoMetrics.CacheMisses, 1)

	// Query database
	var rule models.GeoRule
	err := s.db.Where("scope_type = ? AND scope_id = ? AND status = ?", 
		scopeType, scopeID, models.GeoRuleStatusActive).
		Order("priority ASC").
		First(&rule).Error
	
	if err != nil {
		// Cache negative result
		cache.Set(ctx, cacheKey, "null", geoRuleCacheTTL)
		return nil
	}

	// Cache result
	if jsonBytes, err := json.Marshal(rule); err == nil {
		cache.Set(ctx, cacheKey, string(jsonBytes), geoRuleCacheTTL)
	}

	return &rule
}

// getGlobalRule gets the global rule
func (s *GeoRuleService) getGlobalRule(ctx context.Context) *models.GeoRule {
	cacheKey := geoRuleCachePrefix + "global"
	
	// Try cache first
	if cached, err := cache.Get(ctx, cacheKey); err == nil && cached != "" {
		atomic.AddInt64(&geoMetrics.CacheHits, 1)
		if cached == "null" {
			return nil
		}
		var rule models.GeoRule
		if err := json.Unmarshal([]byte(cached), &rule); err == nil {
			return &rule
		}
	}
	
	atomic.AddInt64(&geoMetrics.CacheMisses, 1)

	// Query database
	var rule models.GeoRule
	err := s.db.Where("scope_type = ? AND status = ?", 
		models.GeoRuleScopeGlobal, models.GeoRuleStatusActive).
		Order("priority ASC").
		First(&rule).Error
	
	if err != nil {
		cache.Set(ctx, cacheKey, "null", geoRuleCacheTTL)
		return nil
	}

	// Cache result
	if jsonBytes, err := json.Marshal(rule); err == nil {
		cache.Set(ctx, cacheKey, string(jsonBytes), geoRuleCacheTTL)
	}

	return &rule
}

// InvalidateCache invalidates cache for a rule
func (s *GeoRuleService) InvalidateCache(scopeType models.GeoRuleScopeType, scopeID *uuid.UUID) {
	ctx := context.Background()
	
	if scopeType == models.GeoRuleScopeGlobal {
		cache.Delete(ctx, geoRuleCachePrefix+"global")
	} else if scopeID != nil {
		cacheKey := fmt.Sprintf("%s%s:%s", geoRuleCachePrefix, scopeType, scopeID.String())
		cache.Delete(ctx, cacheKey)
	}
}

// ============================================
// CRUD OPERATIONS
// ============================================

// CreateGeoRule creates a new geo rule
func (s *GeoRuleService) CreateGeoRule(req *models.CreateGeoRuleRequest) (*models.GeoRule, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Validate country codes
	valid, invalid := models.ValidateCountryCodes(req.Countries)
	if len(invalid) > 0 {
		return nil, fmt.Errorf("invalid country codes: %v", invalid)
	}
	if len(valid) == 0 {
		return nil, fmt.Errorf("at least one valid country code is required")
	}

	// Parse scope ID
	var scopeID *uuid.UUID
	if req.ScopeType != "global" {
		if req.ScopeID == "" {
			return nil, fmt.Errorf("scope_id is required for %s rules", req.ScopeType)
		}
		id, err := uuid.Parse(req.ScopeID)
		if err != nil {
			return nil, fmt.Errorf("invalid scope_id: %w", err)
		}
		scopeID = &id

		// Verify scope exists
		if req.ScopeType == "offer" {
			var offer models.Offer
			if err := s.db.First(&offer, "id = ?", id).Error; err != nil {
				return nil, fmt.Errorf("offer not found")
			}
		} else if req.ScopeType == "advertiser" {
			var user models.AfftokUser
			if err := s.db.First(&user, "id = ?", id).Error; err != nil {
				return nil, fmt.Errorf("advertiser not found")
			}
		}
	}

	// Set default priority
	priority := req.Priority
	if priority <= 0 {
		priority = 100
	}

	// Create countries JSON
	countriesJSON, _ := json.Marshal(valid)

	rule := &models.GeoRule{
		ScopeType:   models.GeoRuleScopeType(req.ScopeType),
		ScopeID:     scopeID,
		Mode:        models.GeoRuleMode(req.Mode),
		Countries:   datatypes.JSON(countriesJSON),
		Priority:    priority,
		Status:      models.GeoRuleStatusActive,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := s.db.Create(rule).Error; err != nil {
		return nil, fmt.Errorf("failed to create geo rule: %w", err)
	}

	// Invalidate cache
	s.InvalidateCache(rule.ScopeType, scopeID)

	return rule, nil
}

// UpdateGeoRule updates an existing geo rule
func (s *GeoRuleService) UpdateGeoRule(ruleID uuid.UUID, req *models.UpdateGeoRuleRequest) (*models.GeoRule, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var rule models.GeoRule
	if err := s.db.First(&rule, "id = ?", ruleID).Error; err != nil {
		return nil, fmt.Errorf("rule not found")
	}

	// Update fields
	if req.Mode != nil {
		rule.Mode = models.GeoRuleMode(*req.Mode)
	}
	if req.Countries != nil && len(req.Countries) > 0 {
		valid, invalid := models.ValidateCountryCodes(req.Countries)
		if len(invalid) > 0 {
			return nil, fmt.Errorf("invalid country codes: %v", invalid)
		}
		countriesJSON, _ := json.Marshal(valid)
		rule.Countries = datatypes.JSON(countriesJSON)
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}
	if req.Status != nil {
		rule.Status = models.GeoRuleStatus(*req.Status)
	}
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Description != nil {
		rule.Description = *req.Description
	}

	rule.UpdatedAt = time.Now().UTC()

	if err := s.db.Save(&rule).Error; err != nil {
		return nil, fmt.Errorf("failed to update geo rule: %w", err)
	}

	// Invalidate cache
	s.InvalidateCache(rule.ScopeType, rule.ScopeID)

	return &rule, nil
}

// DeleteGeoRule deletes a geo rule
func (s *GeoRuleService) DeleteGeoRule(ruleID uuid.UUID) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var rule models.GeoRule
	if err := s.db.First(&rule, "id = ?", ruleID).Error; err != nil {
		return fmt.Errorf("rule not found")
	}

	if err := s.db.Delete(&rule).Error; err != nil {
		return fmt.Errorf("failed to delete geo rule: %w", err)
	}

	// Invalidate cache
	s.InvalidateCache(rule.ScopeType, rule.ScopeID)

	return nil
}

// GetGeoRuleByID returns a geo rule by ID
func (s *GeoRuleService) GetGeoRuleByID(ruleID uuid.UUID) (*models.GeoRuleInfo, error) {
	var rule models.GeoRule
	if err := s.db.First(&rule, "id = ?", ruleID).Error; err != nil {
		return nil, fmt.Errorf("rule not found")
	}

	return s.toGeoRuleInfo(&rule), nil
}

// GetAllGeoRules returns all geo rules with pagination
func (s *GeoRuleService) GetAllGeoRules(page, limit int, scopeType string) ([]models.GeoRuleInfo, int64, error) {
	var rules []models.GeoRule
	var total int64

	offset := (page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	query := s.db.Model(&models.GeoRule{})
	if scopeType != "" {
		query = query.Where("scope_type = ?", scopeType)
	}

	query.Count(&total)

	if err := query.Order("priority ASC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&rules).Error; err != nil {
		return nil, 0, err
	}

	result := make([]models.GeoRuleInfo, len(rules))
	for i := range rules {
		result[i] = *s.toGeoRuleInfo(&rules[i])
	}

	return result, total, nil
}

// GetGeoRulesByOffer returns geo rules for a specific offer
func (s *GeoRuleService) GetGeoRulesByOffer(offerID uuid.UUID) ([]models.GeoRuleInfo, error) {
	var rules []models.GeoRule
	if err := s.db.Where("scope_type = ? AND scope_id = ?", 
		models.GeoRuleScopeOffer, offerID).
		Order("priority ASC").
		Find(&rules).Error; err != nil {
		return nil, err
	}

	result := make([]models.GeoRuleInfo, len(rules))
	for i := range rules {
		result[i] = *s.toGeoRuleInfo(&rules[i])
	}

	return result, nil
}

// GetGeoRulesByAdvertiser returns geo rules for a specific advertiser
func (s *GeoRuleService) GetGeoRulesByAdvertiser(advertiserID uuid.UUID) ([]models.GeoRuleInfo, error) {
	var rules []models.GeoRule
	if err := s.db.Where("scope_type = ? AND scope_id = ?", 
		models.GeoRuleScopeAdvertiser, advertiserID).
		Order("priority ASC").
		Find(&rules).Error; err != nil {
		return nil, err
	}

	result := make([]models.GeoRuleInfo, len(rules))
	for i := range rules {
		result[i] = *s.toGeoRuleInfo(&rules[i])
	}

	return result, nil
}

// toGeoRuleInfo converts GeoRule to GeoRuleInfo
func (s *GeoRuleService) toGeoRuleInfo(rule *models.GeoRule) *models.GeoRuleInfo {
	info := &models.GeoRuleInfo{
		ID:          rule.ID,
		ScopeType:   rule.ScopeType,
		ScopeID:     rule.ScopeID,
		Mode:        rule.Mode,
		Countries:   s.parseCountries(rule.Countries),
		Priority:    rule.Priority,
		Status:      rule.Status,
		Name:        rule.Name,
		Description: rule.Description,
		CreatedAt:   rule.CreatedAt,
		UpdatedAt:   rule.UpdatedAt,
	}

	// Get scope name
	if rule.ScopeID != nil {
		switch rule.ScopeType {
		case models.GeoRuleScopeOffer:
			var offer models.Offer
			if s.db.Select("title").First(&offer, "id = ?", rule.ScopeID).Error == nil {
				info.ScopeName = offer.Title
			}
		case models.GeoRuleScopeAdvertiser:
			var user models.AfftokUser
			if s.db.Select("full_name").First(&user, "id = ?", rule.ScopeID).Error == nil {
				info.ScopeName = user.FullName
			}
		}
	}

	return info
}

// ============================================
// STATISTICS
// ============================================

// GeoRuleStats represents geo rule statistics
type GeoRuleStats struct {
	TotalRules       int64            `json:"total_rules"`
	ActiveRules      int64            `json:"active_rules"`
	DisabledRules    int64            `json:"disabled_rules"`
	OfferRules       int64            `json:"offer_rules"`
	AdvertiserRules  int64            `json:"advertiser_rules"`
	GlobalRules      int64            `json:"global_rules"`
	AllowRules       int64            `json:"allow_rules"`
	BlockRules       int64            `json:"block_rules"`
	TopBlockedCountries map[string]int64 `json:"top_blocked_countries"`
}

// GetGeoRuleStats returns geo rule statistics
func (s *GeoRuleService) GetGeoRuleStats() (*GeoRuleStats, error) {
	stats := &GeoRuleStats{
		TopBlockedCountries: make(map[string]int64),
	}

	// Count by status
	s.db.Model(&models.GeoRule{}).Count(&stats.TotalRules)
	s.db.Model(&models.GeoRule{}).Where("status = ?", models.GeoRuleStatusActive).Count(&stats.ActiveRules)
	s.db.Model(&models.GeoRule{}).Where("status = ?", models.GeoRuleStatusDisabled).Count(&stats.DisabledRules)

	// Count by scope type
	s.db.Model(&models.GeoRule{}).Where("scope_type = ?", models.GeoRuleScopeOffer).Count(&stats.OfferRules)
	s.db.Model(&models.GeoRule{}).Where("scope_type = ?", models.GeoRuleScopeAdvertiser).Count(&stats.AdvertiserRules)
	s.db.Model(&models.GeoRule{}).Where("scope_type = ?", models.GeoRuleScopeGlobal).Count(&stats.GlobalRules)

	// Count by mode
	s.db.Model(&models.GeoRule{}).Where("mode = ?", models.GeoRuleModeAllow).Count(&stats.AllowRules)
	s.db.Model(&models.GeoRule{}).Where("mode = ?", models.GeoRuleModeBlock).Count(&stats.BlockRules)

	return stats, nil
}


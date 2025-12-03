package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ============================================
// ADMIN GEO RULES HANDLER
// ============================================

// AdminGeoRulesHandler handles geo rule management endpoints
type AdminGeoRulesHandler struct {
	geoRuleService *services.GeoRuleService
}

// NewAdminGeoRulesHandler creates a new admin geo rules handler
func NewAdminGeoRulesHandler(geoRuleService *services.GeoRuleService) *AdminGeoRulesHandler {
	return &AdminGeoRulesHandler{
		geoRuleService: geoRuleService,
	}
}

// ============================================
// LIST GEO RULES
// ============================================

// GetAllGeoRules returns all geo rules with pagination
// GET /api/admin/geo-rules?page=1&limit=20&scope_type=offer
func (h *AdminGeoRulesHandler) GetAllGeoRules(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	scopeType := c.Query("scope_type")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	rules, total, err := h.geoRuleService.GetAllGeoRules(page, limit, scopeType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to fetch geo rules: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"geo_rules":   rules,
			"total":       total,
			"page":        page,
			"limit":       limit,
			"total_pages": (total + int64(limit) - 1) / int64(limit),
		},
		"timestamp": time.Now().UTC(),
	})
}

// GetGeoRuleByID returns a single geo rule by ID
// GET /api/admin/geo-rules/:id
func (h *AdminGeoRulesHandler) GetGeoRuleByID(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid geo rule ID",
		})
		return
	}

	rule, err := h.geoRuleService.GetGeoRuleByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Geo rule not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           rule,
		"timestamp":      time.Now().UTC(),
	})
}

// GetGeoRulesByOffer returns geo rules for a specific offer
// GET /api/admin/offers/:id/geo-rules
func (h *AdminGeoRulesHandler) GetGeoRulesByOffer(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	offerIDStr := c.Param("id")
	offerID, err := uuid.Parse(offerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid offer ID",
		})
		return
	}

	rules, err := h.geoRuleService.GetGeoRulesByOffer(offerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to fetch geo rules: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"geo_rules": rules,
			"count":     len(rules),
			"offer_id":  offerID,
		},
		"timestamp": time.Now().UTC(),
	})
}

// GetGeoRulesByAdvertiser returns geo rules for a specific advertiser
// GET /api/admin/advertisers/:id/geo-rules
func (h *AdminGeoRulesHandler) GetGeoRulesByAdvertiser(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	advertiserIDStr := c.Param("id")
	advertiserID, err := uuid.Parse(advertiserIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid advertiser ID",
		})
		return
	}

	rules, err := h.geoRuleService.GetGeoRulesByAdvertiser(advertiserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to fetch geo rules: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"geo_rules":     rules,
			"count":         len(rules),
			"advertiser_id": advertiserID,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// CREATE GEO RULE
// ============================================

// CreateGeoRule creates a new geo rule
// POST /api/admin/geo-rules
func (h *AdminGeoRulesHandler) CreateGeoRule(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req models.CreateGeoRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate scope_type
	validScopeTypes := map[string]bool{"offer": true, "advertiser": true, "global": true}
	if !validScopeTypes[req.ScopeType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid scope_type. Must be: offer, advertiser, or global",
		})
		return
	}

	// Validate mode
	validModes := map[string]bool{"allow": true, "block": true}
	if !validModes[req.Mode] {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid mode. Must be: allow or block",
		})
		return
	}

	// Validate countries
	if len(req.Countries) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "At least one country code is required",
		})
		return
	}

	valid, invalid := models.ValidateCountryCodes(req.Countries)
	if len(invalid) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid country codes",
			"invalid_codes":  invalid,
			"valid_codes":    valid,
		})
		return
	}

	rule, err := h.geoRuleService.CreateGeoRule(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to create geo rule: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           rule,
		"message":        "Geo rule created successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// UPDATE GEO RULE
// ============================================

// UpdateGeoRule updates an existing geo rule
// PUT /api/admin/geo-rules/:id
func (h *AdminGeoRulesHandler) UpdateGeoRule(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid geo rule ID",
		})
		return
	}

	var req models.UpdateGeoRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate mode if provided
	if req.Mode != nil {
		validModes := map[string]bool{"allow": true, "block": true}
		if !validModes[*req.Mode] {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":        false,
				"correlation_id": correlationID,
				"error":          "Invalid mode. Must be: allow or block",
			})
			return
		}
	}

	// Validate status if provided
	if req.Status != nil {
		validStatuses := map[string]bool{"active": true, "disabled": true}
		if !validStatuses[*req.Status] {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":        false,
				"correlation_id": correlationID,
				"error":          "Invalid status. Must be: active or disabled",
			})
			return
		}
	}

	// Validate countries if provided
	if len(req.Countries) > 0 {
		valid, invalid := models.ValidateCountryCodes(req.Countries)
		if len(invalid) > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success":        false,
				"correlation_id": correlationID,
				"error":          "Invalid country codes",
				"invalid_codes":  invalid,
				"valid_codes":    valid,
			})
			return
		}
	}

	rule, err := h.geoRuleService.UpdateGeoRule(id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to update geo rule: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           rule,
		"message":        "Geo rule updated successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// DELETE GEO RULE
// ============================================

// DeleteGeoRule deletes a geo rule
// DELETE /api/admin/geo-rules/:id
func (h *AdminGeoRulesHandler) DeleteGeoRule(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid geo rule ID",
		})
		return
	}

	if err := h.geoRuleService.DeleteGeoRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to delete geo rule: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"message":        "Geo rule deleted successfully",
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// STATISTICS
// ============================================

// GetGeoRuleStats returns geo rule statistics
// GET /api/admin/geo-rules/stats
func (h *AdminGeoRulesHandler) GetGeoRuleStats(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	stats, err := h.geoRuleService.GetGeoRuleStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Failed to get geo rule stats: " + err.Error(),
		})
		return
	}

	// Get metrics
	metrics := services.GetGeoRuleMetrics()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"rules": stats,
			"enforcement": gin.H{
				"total_checks":    metrics.TotalChecks,
				"blocked_by_rule": metrics.BlockedByRule,
				"allowed_by_rule": metrics.AllowedByRule,
				"no_rule_applied": metrics.NoRuleApplied,
				"cache_hits":      metrics.CacheHits,
				"cache_misses":    metrics.CacheMisses,
			},
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// COUNTRY CODES REFERENCE
// ============================================

// GetCountryCodes returns all valid country codes
// GET /api/admin/geo-rules/countries
func (h *AdminGeoRulesHandler) GetCountryCodes(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	// Convert map to array for easier frontend use
	type CountryInfo struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}

	countries := make([]CountryInfo, 0, len(models.ValidCountryCodes))
	for code, name := range models.ValidCountryCodes {
		countries = append(countries, CountryInfo{
			Code: code,
			Name: name,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"countries": countries,
			"count":     len(countries),
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// TEST GEO RULE
// ============================================

// TestGeoRule tests a geo rule against a country
// POST /api/admin/geo-rules/test
func (h *AdminGeoRulesHandler) TestGeoRule(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req struct {
		OfferID      string `json:"offer_id,omitempty"`
		AdvertiserID string `json:"advertiser_id,omitempty"`
		CountryCode  string `json:"country_code" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate country code
	if !models.IsValidCountryCode(req.CountryCode) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid country code: " + req.CountryCode,
		})
		return
	}

	var offerID, advertiserID *uuid.UUID

	if req.OfferID != "" {
		id, err := uuid.Parse(req.OfferID)
		if err == nil {
			offerID = &id
		}
	}

	if req.AdvertiserID != "" {
		id, err := uuid.Parse(req.AdvertiserID)
		if err == nil {
			advertiserID = &id
		}
	}

	result := h.geoRuleService.GetEffectiveGeoRule(offerID, advertiserID, req.CountryCode)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"allowed":      result.Allowed,
			"reason":       result.Reason,
			"rule_applied": result.Rule != nil,
			"rule":         result.Rule,
			"country_code": req.CountryCode,
			"country_name": models.GetCountryName(req.CountryCode),
		},
		"timestamp": time.Now().UTC(),
	})
}


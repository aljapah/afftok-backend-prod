package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostbackHandler struct {
	db                   *gorm.DB
	securityService      *services.SecurityService
	observabilityService *services.ObservabilityService
	apiKeyService        *services.APIKeyService
	geoRuleService       *services.GeoRuleService
}

func NewPostbackHandler(db *gorm.DB) *PostbackHandler {
	return &PostbackHandler{
		db:                   db,
		securityService:      services.NewSecurityService(),
		observabilityService: services.NewObservabilityService(),
		apiKeyService:        services.NewAPIKeyService(db),
		geoRuleService:       services.NewGeoRuleService(db),
	}
}

// SetAPIKeyService sets the API key service (for dependency injection)
func (h *PostbackHandler) SetAPIKeyService(service *services.APIKeyService) {
	h.apiKeyService = service
}

// SetGeoRuleService sets the geo rule service (for dependency injection)
func (h *PostbackHandler) SetGeoRuleService(service *services.GeoRuleService) {
	h.geoRuleService = service
}

// PostbackRequest represents incoming postback data from advertisers
type PostbackRequest struct {
	// Required fields
	UserOfferID string `json:"user_offer_id" form:"user_offer_id"`
	
	// Alternative identifiers (for different networks)
	TrackingCode string `json:"tracking_code" form:"tracking_code" query:"aff"`
	SubID        string `json:"sub_id" form:"sub_id" query:"sub_id"`
	ClickID      string `json:"click_id" form:"click_id" query:"click_id"`
	
	// Conversion details
	Amount       int    `json:"amount" form:"amount" query:"amount"`
	Commission   int    `json:"commission" form:"commission" query:"commission"`
	Currency     string `json:"currency" form:"currency" query:"currency"`
	Status       string `json:"status" form:"status" query:"status"`
	
	// External identifiers
	ExternalID   string `json:"external_id" form:"external_id" query:"external_id"`
	TransactionID string `json:"transaction_id" form:"transaction_id" query:"transaction_id"`
	
	// Network identification
	NetworkID    string `json:"network_id" form:"network_id" query:"network_id"`
	NetworkName  string `json:"network_name" form:"network_name" query:"network_name"`
	
	// Geo information (optional - for geo rule enforcement)
	Country      string `json:"country" form:"country" query:"country"`
	
	// Verification
	Signature    string `json:"signature" form:"signature" query:"sig"`
	Token        string `json:"token" form:"token" query:"token"`
}

// GeoEnforceOnPostback controls whether geo rules are enforced on postbacks
// Can be set via environment variable GEO_ENFORCE_ON_POSTBACK
var GeoEnforceOnPostback = false

// HandlePostback processes incoming postback/callback from advertisers
// Supports both API key auth (advertisers) and JWT auth (internal)
func (h *PostbackHandler) HandlePostback(c *gin.Context) {
	ip := c.ClientIP()
	startTime := time.Now()
	
	// Check authentication method
	authMethod := c.GetString("auth_method")
	apiKeyID := c.GetString("api_key_id")
	advertiserID := c.GetString("advertiser_id")
	
	// If API key auth, get the advertiser context
	var authenticatedAdvertiserID *uuid.UUID
	if authMethod == "api_key" && advertiserID != "" {
		if id, err := uuid.Parse(advertiserID); err == nil {
			authenticatedAdvertiserID = &id
		}
	}
	
	// Security: Rate limiting (different limits for API key vs IP)
	var rateLimitKey string
	var rateLimit int
	if apiKeyID != "" {
		rateLimitKey = "postback:apikey:" + apiKeyID
		rateLimit = 100 // Higher limit for authenticated API keys
	} else {
		rateLimitKey = "postback:ip:" + ip
		rateLimit = services.RateLimitPostbackPerMinute
	}
	
	rateLimitResult := h.securityService.CheckRateLimit(rateLimitKey, rateLimit, time.Minute)
	if !rateLimitResult.Allowed {
		h.observabilityService.LogRateLimit(ip, "/api/postback", "postback_rate_limit")
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many postback requests"})
		return
	}
	
	// Log authentication context
	if apiKeyID != "" {
		h.observabilityService.Log(services.LogEvent{
			Timestamp: time.Now(),
			Level:     services.LogLevelInfo,
			Category:  "api_key_event",
			Message:   "Postback received via API key",
			IP:        ip,
			Metadata: map[string]interface{}{
				"api_key_id":    apiKeyID,
				"advertiser_id": advertiserID,
				"auth_method":   authMethod,
			},
		})
	}
	
	_ = authenticatedAdvertiserID // Will be used for validation in future

	var req PostbackRequest
	
	// Support both JSON and form/query params
	if c.ContentType() == "application/json" {
		if err := c.ShouldBindJSON(&req); err != nil {
			h.securityService.LogAuditEvent(services.AuditEvent{
				Timestamp: time.Now(),
				EventType: "postback_invalid_json",
				IP:        ip,
				Resource:  c.Request.URL.Path,
				Action:    "postback",
				Success:   false,
				Details: map[string]interface{}{
					"error": err.Error(),
				},
			})
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}
	} else {
		// Bind from query params and form data
		if err := c.ShouldBindQuery(&req); err != nil {
			// Try form binding
			c.ShouldBind(&req)
		}
	}

	// Security: Validate input lengths
	if len(req.UserOfferID) > 50 || len(req.TrackingCode) > 100 || len(req.ExternalID) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameter length"})
		return
	}

	// Log incoming postback for debugging
	postbackJSON, _ := json.Marshal(req)
	fmt.Printf("[Postback] Received from %s: %s\n", ip, string(postbackJSON))
	
	// Log postback received
	h.observabilityService.LogPostback(
		req.UserOfferID,
		req.NetworkID,
		req.ExternalID,
		ip,
		true,
		"",
		0, // duration will be set later
	)

	// Resolve user offer ID from various sources
	userOfferID, err := h.resolveUserOfferID(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate user offer exists
	var userOffer models.UserOffer
	if err := h.db.Preload("Offer").First(&userOffer, "id = ?", userOfferID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User offer not found"})
		return
	}

	// Generate unique external ID if not provided
	externalID := req.ExternalID
	if externalID == "" {
		externalID = req.TransactionID
	}
	if externalID == "" {
		// Generate one based on request data
		externalID = fmt.Sprintf("auto_%s_%d", userOfferID.String()[:8], time.Now().UnixNano())
	}

	// Check for duplicate conversion
	if externalID != "" {
		var existingConversion models.Conversion
		if err := h.db.Where("external_conversion_id = ?", externalID).First(&existingConversion).Error; err == nil {
			// Duplicate found - return success but don't create new
			c.JSON(http.StatusOK, gin.H{
				"message":    "Conversion already recorded",
				"duplicate":  true,
				"conversion": existingConversion,
			})
			return
		}
	}

	// Resolve click ID if provided
	var clickID *uuid.UUID
	if req.ClickID != "" {
		if id, err := uuid.Parse(req.ClickID); err == nil {
			clickID = &id
		}
	}

	// Resolve network ID
	var networkID *uuid.UUID
	if req.NetworkID != "" {
		if id, err := uuid.Parse(req.NetworkID); err == nil {
			networkID = &id
		}
	}

	// Determine status
	status := req.Status
	if status == "" {
		status = models.ConversionStatusPending
	}

	// Determine currency
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	// Calculate commission if not provided
	commission := req.Commission
	if commission == 0 && userOffer.Offer != nil {
		commission = userOffer.Offer.Commission
	}

	// Store postback data for audit
	postbackData, _ := json.Marshal(req)
	now := time.Now().UTC()

	// Create conversion record
	conversion := models.Conversion{
		ID:                   uuid.New(),
		UserOfferID:          userOfferID,
		ClickID:              clickID,
		ExternalConversionID: externalID,
		NetworkID:            networkID,
		Amount:               req.Amount,
		Commission:           commission,
		Currency:             currency,
		Status:               status,
		ConvertedAt:          now,
		PostbackData:         string(postbackData),
		PostbackReceivedAt:   &now,
	}

	// Use transaction for atomic updates
	err = h.db.Transaction(func(tx *gorm.DB) error {
		// 1. Create conversion
		if err := tx.Create(&conversion).Error; err != nil {
			return fmt.Errorf("failed to create conversion: %w", err)
		}

		// 2. Update UserOffer stats (conversion counter + earnings + updated_at)
		userOfferUpdates := map[string]interface{}{
			"total_conversions": gorm.Expr("total_conversions + 1"),
			"updated_at":        now,
		}
		if status == models.ConversionStatusApproved {
			userOfferUpdates["earnings"] = gorm.Expr("earnings + ?", commission)
		}
		if err := tx.Model(&models.UserOffer{}).
			Where("id = ?", userOfferID).
			UpdateColumns(userOfferUpdates).Error; err != nil {
			return fmt.Errorf("failed to update user offer: %w", err)
		}

		// 3. Update Offer total conversions
		if err := tx.Model(&models.Offer{}).
			Where("id = ?", userOffer.OfferID).
			UpdateColumn("total_conversions", gorm.Expr("total_conversions + 1")).Error; err != nil {
			return fmt.Errorf("failed to update offer conversions: %w", err)
		}

		// 4. Update User stats
		if err := tx.Model(&models.AfftokUser{}).
			Where("id = ?", userOffer.UserID).
			UpdateColumn("total_conversions", gorm.Expr("total_conversions + 1")).Error; err != nil {
			return fmt.Errorf("failed to update user conversions: %w", err)
		}

		// 5. If approved, update earnings
		if status == models.ConversionStatusApproved {
			if err := tx.Model(&models.AfftokUser{}).
				Where("id = ?", userOffer.UserID).
				UpdateColumn("total_earnings", gorm.Expr("total_earnings + ?", commission)).Error; err != nil {
				return fmt.Errorf("failed to update user earnings: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process postback: " + err.Error()})
		return
	}

	fmt.Printf("[Postback] Conversion created: %s for user offer %s\n", conversion.ID.String(), userOfferID.String())

	// Log conversion with full observability
	durationMs := time.Since(startTime).Milliseconds()
	h.observabilityService.LogConversion(
		conversion.ID.String(),
		userOfferID.String(),
		userOffer.UserID.String(),
		req.Amount,
		commission,
		status,
	)
	h.observabilityService.LogPerformance("postback_processing", durationMs, 0, 0)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Conversion recorded successfully",
		"conversion": conversion,
	})
}

// resolveUserOfferID resolves the user offer ID from various request parameters
func (h *PostbackHandler) resolveUserOfferID(req *PostbackRequest) (uuid.UUID, error) {
	// 1. Direct user_offer_id
	if req.UserOfferID != "" {
		id, err := uuid.Parse(req.UserOfferID)
		if err == nil {
			return id, nil
		}
	}

	// 2. From tracking code (short link)
	if req.TrackingCode != "" {
		var userOffer models.UserOffer
		if err := h.db.Where("short_link LIKE ?", "%"+req.TrackingCode+"%").First(&userOffer).Error; err == nil {
			return userOffer.ID, nil
		}
	}

	// 3. From sub_id (some networks use this)
	if req.SubID != "" {
		id, err := uuid.Parse(req.SubID)
		if err == nil {
			var userOffer models.UserOffer
			if err := h.db.First(&userOffer, "id = ?", id).Error; err == nil {
				return userOffer.ID, nil
			}
		}
	}

	return uuid.Nil, fmt.Errorf("unable to resolve user offer ID from request")
}

// ApproveConversion approves a pending conversion
func (h *PostbackHandler) ApproveConversion(c *gin.Context) {
	conversionID := c.Param("id")

	var conversion models.Conversion
	if err := h.db.First(&conversion, "id = ?", conversionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversion not found"})
		return
	}

	if !conversion.CanApprove() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversion cannot be approved (current status: " + conversion.Status + ")"})
		return
	}

	now := time.Now().UTC()
	
	// Use transaction for atomic update
	err := h.db.Transaction(func(tx *gorm.DB) error {
		// Get user offer for user ID
		var userOffer models.UserOffer
		if err := tx.First(&userOffer, "id = ?", conversion.UserOfferID).Error; err != nil {
			return err
		}

		// Update conversion status
		conversion.Status = models.ConversionStatusApproved
		conversion.ApprovedAt = &now
		if err := tx.Save(&conversion).Error; err != nil {
			return err
		}

		// Update user earnings
		if err := tx.Model(&models.AfftokUser{}).
			Where("id = ?", userOffer.UserID).
			UpdateColumn("total_earnings", gorm.Expr("total_earnings + ?", conversion.Commission)).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve conversion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Conversion approved successfully",
		"conversion": conversion,
	})
}

// RejectConversion rejects a pending conversion
func (h *PostbackHandler) RejectConversion(c *gin.Context) {
	conversionID := c.Param("id")

	type RejectRequest struct {
		Reason string `json:"reason"`
	}

	var req RejectRequest
	c.ShouldBindJSON(&req)

	var conversion models.Conversion
	if err := h.db.First(&conversion, "id = ?", conversionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversion not found"})
		return
	}

	if !conversion.CanReject() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversion cannot be rejected (current status: " + conversion.Status + ")"})
		return
	}

	conversion.Status = models.ConversionStatusRejected
	conversion.RejectionReason = req.Reason

	if err := h.db.Save(&conversion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject conversion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Conversion rejected successfully",
		"conversion": conversion,
	})
}

// GetConversions returns conversions with filtering
func (h *PostbackHandler) GetConversions(c *gin.Context) {
	status := c.Query("status")
	userOfferID := c.Query("user_offer_id")
	networkID := c.Query("network_id")

	query := h.db.Model(&models.Conversion{}).Order("converted_at DESC")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if userOfferID != "" {
		query = query.Where("user_offer_id = ?", userOfferID)
	}
	if networkID != "" {
		query = query.Where("network_id = ?", networkID)
	}

	var conversions []models.Conversion
	if err := query.Limit(100).Find(&conversions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversions"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversions": conversions,
		"count":       len(conversions),
	})
}

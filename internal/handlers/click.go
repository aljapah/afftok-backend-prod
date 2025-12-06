package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClickHandler struct {
	db                   *gorm.DB
	clickService         *services.ClickService
	linkService          *services.LinkService
	securityService      *services.SecurityService
	observabilityService *services.ObservabilityService
	geoRuleService       *services.GeoRuleService
	linkSigningService   *services.LinkSigningService
}

func NewClickHandler(db *gorm.DB) *ClickHandler {
	return &ClickHandler{
		db:                   db,
		clickService:         services.NewClickService(),
		linkService:          services.NewLinkService(),
		securityService:      services.NewSecurityService(),
		observabilityService: services.NewObservabilityService(),
		geoRuleService:       services.NewGeoRuleService(db),
		linkSigningService:   services.NewLinkSigningService(),
	}
}

// SetGeoRuleService sets the geo rule service (for dependency injection)
func (h *ClickHandler) SetGeoRuleService(service *services.GeoRuleService) {
	h.geoRuleService = service
}

// SetLinkSigningService sets the link signing service (for dependency injection)
func (h *ClickHandler) SetLinkSigningService(service *services.LinkSigningService) {
	h.linkSigningService = service
}

// TrackClick handles click tracking and redirect
// Supports multiple URL formats:
// - /api/c/{offerID}?promoter={promoterID} (legacy)
// - /api/c/{trackingCode} (legacy secure format)
// - /api/c/{trackingCode}.{timestamp}.{nonce}.{signature} (new signed format)
func (h *ClickHandler) TrackClick(c *gin.Context) {
	rawCode := c.Param("id")
	promoterID := c.Query("promoter")
	ip := c.ClientIP()
	
	fmt.Printf("[Click] Tracking request: raw=%s, promoter=%s, ip=%s\n", rawCode, promoterID, ip)

	startTime := time.Now()

	// Security Check 0: Link Signature Validation
	signResult := h.linkSigningService.ValidateSignedLink(rawCode)
	if !signResult.Valid && !signResult.IsLegacy {
		// Log fraud event for invalid signature
		h.observabilityService.LogFraud(
			ip,
			c.Request.UserAgent(),
			signResult.Reason,
			90, // High risk score for signature issues
			0.95,
			signResult.Indicators,
			map[string]interface{}{
				"raw":           rawCode,
				"reason":        signResult.Reason,
				"tracking_code": signResult.TrackingCode,
			},
		)
		
		fmt.Printf("[Click] Link validation failed: %s, reason=%s\n", rawCode, signResult.Reason)
		
		// Handle invalid link - try to redirect anyway
		h.handleInvalidLink(c, signResult.TrackingCode)
		return
	}
	
	idOrCode := signResult.TrackingCode
	signatureValid := signResult.Valid
	isLegacyCode := signResult.IsLegacy
	
	// Log legacy code usage for monitoring
	if isLegacyCode && signatureValid {
		fmt.Printf("[Click] Legacy code accepted: %s\n", idOrCode)
	}

	_ = signatureValid // Used for logging

	// Security Check 1: Bot Detection
	botResult := h.securityService.DetectBot(c)
	if botResult.IsBot && botResult.Confidence > 0.85 {
		// Log fraud detection
		h.observabilityService.LogFraud(
			ip,
			c.Request.UserAgent(),
			botResult.Reason,
			botResult.RiskScore,
			botResult.Confidence,
			[]string{"bot_detected", botResult.Reason},
			map[string]interface{}{
				"tracking_code": idOrCode,
			},
		)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Security Check 2: Rate Limiting
	// Create a dummy UUID for rate limiting if we don't have the real one yet
	rateLimitResult := h.securityService.CheckClickRateLimit(ip, uuid.Nil)
	if !rateLimitResult.Allowed {
		h.observabilityService.LogRateLimit(ip, "/api/c/"+idOrCode, rateLimitResult.Reason)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Too many requests"})
		return
	}

	// Security Check 3: Validate input
	if len(idOrCode) > 100 || len(promoterID) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parameters"})
		return
	}

	var userOffer models.UserOffer
	var offer models.Offer

	// Try to resolve as tracking code first
	if strings.Contains(idOrCode, "-") {
		// New format: tracking code
		userOfferID, err := h.linkService.ResolveTrackingCode(idOrCode)
		if err == nil {
			if err := h.db.Preload("Offer").First(&userOffer, "id = ?", userOfferID).Error; err == nil {
				offer = *userOffer.Offer
				goto trackAndRedirect
			}
		}
	}

	// Try as offer ID (legacy format)
	{
		offerID, err := uuid.Parse(idOrCode)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tracking ID"})
			return
		}

		// Get the offer
		if err := h.db.First(&offer, "id = ?", offerID).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
			return
		}

		// Find or create user offer based on promoter
		if promoterID != "" {
			promoterUUID, err := uuid.Parse(promoterID)
			if err == nil {
				// Try to find existing user offer
				if err := h.db.Where("user_id = ? AND offer_id = ?", promoterUUID, offerID).First(&userOffer).Error; err != nil {
					// Create new user offer with secure tracking code
					affiliateLink, shortLink, err := h.linkService.GenerateAffiliateLink(
						offer.DestinationURL, 
						uuid.New(), // Will be set properly after creation
						promoterUUID,
					)
					if err != nil {
						affiliateLink = fmt.Sprintf("%s?ref=%s", offer.DestinationURL, promoterUUID.String())
						shortLink = ""
					}

					userOffer = models.UserOffer{
						ID:            uuid.New(),
						UserID:        promoterUUID,
						OfferID:       offerID,
						AffiliateLink: affiliateLink,
						ShortLink:     shortLink,
						Status:        "active",
					}
					
					if err := h.db.Create(&userOffer).Error; err != nil {
						fmt.Printf("[Click] Failed to create user offer: %v\n", err)
					} else {
						// Update users_count on offer
						h.db.Model(&offer).UpdateColumn("users_count", gorm.Expr("users_count + 1"))
					}
				}
			}
		}
	}

trackAndRedirect:
	// Track the click if we have a valid user offer
	if userOffer.ID != uuid.Nil {
		// Security Check 4: Geo Rule Check
		// Get country from IP (using existing click service or header)
		countryCode := h.getCountryFromRequest(c)
		
		// Check geo rules
		geoResult := h.geoRuleService.GetEffectiveGeoRule(&offer.ID, &userOffer.UserID, countryCode)
		if !geoResult.Allowed {
			// Log geo block event
			h.observabilityService.LogFraud(
				ip,
				c.Request.UserAgent(),
				"geo_block",
				80, // High risk score for geo violations
				0.9,
				[]string{"geo_block", "geo_rule_violation"},
				map[string]interface{}{
					"country":       countryCode,
					"offer_id":      offer.ID.String(),
					"advertiser_id": userOffer.UserID.String(),
					"rule_id":       getRuleID(geoResult.Rule),
					"mode":          getRuleMode(geoResult.Rule),
					"reason":        geoResult.Reason,
				},
			)
			
			fmt.Printf("[Click] Geo blocked: country=%s, offer=%s, reason=%s\n", 
				countryCode, offer.ID.String(), geoResult.Reason)
			
			// Return safe response (no click recorded)
			// Still redirect to avoid revealing the block
			goto redirectOnly
		}
		
		// Security Check 5: Click Fingerprinting & Deduplication
		fingerprint := h.securityService.GenerateClickFingerprint(userOffer.ID, ip, c.Request.UserAgent())
		if h.securityService.IsClickDuplicate(fingerprint, 5*time.Minute) {
			fmt.Printf("[Click] Duplicate click detected for user offer %s\n", userOffer.ID.String())
			// Still redirect, but don't count the click
			goto redirectOnly
		}

		click, err := h.clickService.TrackClick(c, userOffer.ID)
		durationMs := time.Since(startTime).Milliseconds()
		
		if err != nil {
			fmt.Printf("[Click] Error tracking click: %v\n", err)
			h.observabilityService.LogError(
				"CLICK_TRACK_ERROR",
				err.Error(),
				"/api/c/"+idOrCode,
				"GET",
				ip,
				500,
				services.GenerateCorrelationID(),
			)
			// Continue with redirect even if tracking fails
		} else {
			fmt.Printf("[Click] Click tracked: %s for user offer %s\n", click.ID.String(), userOffer.ID.String())
			
			// Log successful click with full observability
			h.observabilityService.LogClick(
				userOffer.ID.String(),
				ip,
				c.Request.UserAgent(),
				"", // device will be parsed
				idOrCode,
				fingerprint,
				botResult.RiskScore,
				false,
				"",
				durationMs,
			)
		}
	} else {
		fmt.Printf("[Click] No user offer to track for offer %s\n", offer.ID.String())
	}

redirectOnly:

	// Redirect to destination
	destinationURL := offer.DestinationURL
	if destinationURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Offer has no destination URL"})
		return
	}

	fmt.Printf("[Click] Redirecting to: %s\n", destinationURL)
	c.Redirect(http.StatusFound, destinationURL)
}

// handleInvalidLink handles invalid/tampered links
// It tries to redirect to the destination anyway (but doesn't count the click)
func (h *ClickHandler) handleInvalidLink(c *gin.Context, trackingCode string) {
	// Try to resolve the tracking code anyway for redirect (but don't count click)
	if trackingCode != "" {
		// Try to find the offer for redirect
		userOfferID, err := h.linkService.ResolveTrackingCode(trackingCode)
		if err == nil {
			var uo models.UserOffer
			if h.db.Preload("Offer").First(&uo, "id = ?", userOfferID).Error == nil && uo.Offer != nil {
				if uo.Offer.DestinationURL != "" {
					fmt.Printf("[Click] Invalid link, redirecting anyway: %s\n", uo.Offer.DestinationURL)
					c.Redirect(http.StatusFound, uo.Offer.DestinationURL)
					return
				}
			}
		}
	}
	
	// Fallback: return error or redirect to homepage
	fallbackURL := os.Getenv("FALLBACK_REDIRECT_URL")
	if fallbackURL != "" {
		c.Redirect(http.StatusFound, fallbackURL)
		return
	}
	
	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tracking link"})
}

// GetClickStats returns click statistics for a specific user offer
func (h *ClickHandler) GetClickStats(c *gin.Context) {
	userOfferID := c.Param("id")
	userID, _ := c.Get("userID")

	// Verify ownership
	var userOffer models.UserOffer
	if err := h.db.Where("id = ? AND user_id = ?", userOfferID, userID).First(&userOffer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User offer not found or access denied"})
		return
	}

	// Get stats from service
	stats, err := h.clickService.GetClickStats(userOffer.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch click stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetMyClicks returns all clicks for the current user's offers
func (h *ClickHandler) GetMyClicks(c *gin.Context) {
	userID, _ := c.Get("userID")

	// Get all user offers
	var userOffers []models.UserOffer
	if err := h.db.Where("user_id = ?", userID).Find(&userOffers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user offers"})
		return
	}

	if len(userOffers) == 0 {
		c.JSON(http.StatusOK, gin.H{
			"clicks": []models.Click{},
			"total":  0,
		})
		return
	}

	// Get user offer IDs
	offerIDs := make([]uuid.UUID, len(userOffers))
	for i, uo := range userOffers {
		offerIDs[i] = uo.ID
	}

	// Get recent clicks
	var clicks []models.Click
	if err := h.db.Where("user_offer_id IN ?", offerIDs).
		Order("clicked_at DESC").
		Limit(100).
		Find(&clicks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch clicks"})
		return
	}

	// Get total count
	var total int64
	h.db.Model(&models.Click{}).Where("user_offer_id IN ?", offerIDs).Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"clicks": clicks,
		"total":  total,
	})
}

// GetClicksByOffer returns clicks grouped by offer
func (h *ClickHandler) GetClicksByOffer(c *gin.Context) {
	userID, _ := c.Get("userID")

	type OfferClickStats struct {
		OfferID      uuid.UUID `json:"offer_id"`
		OfferTitle   string    `json:"offer_title"`
		TotalClicks  int64     `json:"total_clicks"`
		UniqueClicks int64     `json:"unique_clicks"`
		Conversions  int64     `json:"conversions"`
	}

	var stats []OfferClickStats

	// Get all user offers with stats
	var userOffers []models.UserOffer
	if err := h.db.Preload("Offer").Where("user_id = ?", userID).Find(&userOffers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user offers"})
		return
	}

	for _, uo := range userOffers {
		var totalClicks, uniqueClicks, conversions int64

		// Total clicks
		h.db.Model(&models.Click{}).Where("user_offer_id = ?", uo.ID).Count(&totalClicks)

		// Unique clicks (by IP)
		h.db.Model(&models.Click{}).
			Where("user_offer_id = ?", uo.ID).
			Distinct("ip_address").
			Count(&uniqueClicks)

		// Conversions
		h.db.Model(&models.Conversion{}).Where("user_offer_id = ?", uo.ID).Count(&conversions)

		offerTitle := ""
		if uo.Offer != nil {
			offerTitle = uo.Offer.Title
		}

		stats = append(stats, OfferClickStats{
			OfferID:      uo.OfferID,
			OfferTitle:   offerTitle,
			TotalClicks:  totalClicks,
			UniqueClicks: uniqueClicks,
			Conversions:  conversions,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"stats": stats,
	})
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// getCountryFromRequest extracts country code from request
func (h *ClickHandler) getCountryFromRequest(c *gin.Context) string {
	// 1. Try CF-IPCountry header (Cloudflare)
	if country := c.GetHeader("CF-IPCountry"); country != "" {
		return strings.ToUpper(country)
	}
	
	// 2. Try X-Country header (custom)
	if country := c.GetHeader("X-Country"); country != "" {
		return strings.ToUpper(country)
	}
	
	// 3. Try X-Geo-Country header
	if country := c.GetHeader("X-Geo-Country"); country != "" {
		return strings.ToUpper(country)
	}
	
	// 4. Try to get from click data (if available)
	// This would require a GeoIP service, for now return empty
	// In production, integrate with MaxMind or similar
	
	return ""
}

// getRuleID safely gets rule ID
func getRuleID(rule *models.GeoRule) string {
	if rule == nil {
		return ""
	}
	return rule.ID.String()
}

// getRuleMode safely gets rule mode
func getRuleMode(rule *models.GeoRule) string {
	if rule == nil {
		return ""
	}
	return string(rule.Mode)
}

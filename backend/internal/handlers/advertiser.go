package handlers

import (
	"net/http"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdvertiserHandler handles advertiser-specific operations
type AdvertiserHandler struct {
	db *gorm.DB
}

// NewAdvertiserHandler creates a new advertiser handler
func NewAdvertiserHandler(db *gorm.DB) *AdvertiserHandler {
	return &AdvertiserHandler{db: db}
}

// CreateOfferRequest represents the request body for creating an offer
type CreateOfferRequest struct {
	Title          string `json:"title" binding:"required"`
	TitleAr        string `json:"title_ar"`
	Description    string `json:"description"`
	DescriptionAr  string `json:"description_ar"`
	TermsAr        string `json:"terms_ar"`
	ImageURL       string `json:"image_url"`
	LogoURL        string `json:"logo_url"`
	DestinationURL string `json:"destination_url" binding:"required"`
	Category       string `json:"category"`
	Payout         int    `json:"payout"`
	Commission     int    `json:"commission"`
	PayoutType     string `json:"payout_type"`
}

// CreateOffer creates a new offer for the advertiser (pending approval)
// POST /api/advertiser/offers
func (h *AdvertiserHandler) CreateOffer(c *gin.Context) {
	// Get advertiser ID from JWT
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	advertiserID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Verify user is an advertiser
	var user models.AfftokUser
	if err := h.db.First(&user, "id = ?", advertiserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	if user.Role != "advertiser" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only advertisers can create offers"})
		return
	}

	var req CreateOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default payout type
	payoutType := req.PayoutType
	if payoutType == "" {
		payoutType = "cpa"
	}

	// Create offer with pending status
	offer := models.Offer{
		AdvertiserID:   &advertiserID,
		Title:          req.Title,
		TitleAr:        req.TitleAr,
		Description:    req.Description,
		DescriptionAr:  req.DescriptionAr,
		TermsAr:        req.TermsAr,
		ImageURL:       req.ImageURL,
		LogoURL:        req.LogoURL,
		DestinationURL: req.DestinationURL,
		Category:       req.Category,
		Payout:         req.Payout,
		Commission:     req.Commission,
		PayoutType:     payoutType,
		Status:         "pending", // Always pending for advertiser-created offers
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := h.db.Create(&offer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create offer"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Offer created successfully and pending approval",
		"offer":   offer,
	})
}

// GetMyOffers returns all offers created by the advertiser
// GET /api/advertiser/offers
func (h *AdvertiserHandler) GetMyOffers(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	advertiserID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	var offers []models.Offer
	if err := h.db.Where("advertiser_id = ?", advertiserID).
		Order("created_at DESC").
		Find(&offers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch offers"})
		return
	}

	// Calculate stats for each offer
	type OfferWithStats struct {
		models.Offer
		PromotersCount int `json:"promoters_count"`
	}

	var offersWithStats []OfferWithStats
	for _, offer := range offers {
		var promotersCount int64
		h.db.Model(&models.UserOffer{}).Where("offer_id = ?", offer.ID).Count(&promotersCount)

		offersWithStats = append(offersWithStats, OfferWithStats{
			Offer:          offer,
			PromotersCount: int(promotersCount),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"offers": offersWithStats,
		"total":  len(offersWithStats),
	})
}

// GetOfferStats returns detailed stats for a specific offer
// GET /api/advertiser/offers/:id/stats
func (h *AdvertiserHandler) GetOfferStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	advertiserID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	offerIDStr := c.Param("id")
	offerID, err := uuid.Parse(offerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
		return
	}

	// Verify ownership
	var offer models.Offer
	if err := h.db.First(&offer, "id = ? AND advertiser_id = ?", offerID, advertiserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found or not owned by you"})
		return
	}

	// Get promoters count
	var promotersCount int64
	h.db.Model(&models.UserOffer{}).Where("offer_id = ?", offerID).Count(&promotersCount)

	// Get today's stats
	today := time.Now().Truncate(24 * time.Hour)
	var todayClicks int64
	var todayConversions int64

	h.db.Model(&models.Click{}).
		Joins("JOIN user_offers ON clicks.user_offer_id = user_offers.id").
		Where("user_offers.offer_id = ? AND clicks.clicked_at >= ?", offerID, today).
		Count(&todayClicks)

	h.db.Model(&models.Conversion{}).
		Joins("JOIN user_offers ON conversions.user_offer_id = user_offers.id").
		Where("user_offers.offer_id = ? AND conversions.converted_at >= ?", offerID, today).
		Count(&todayConversions)

	// Get weekly stats
	weekAgo := time.Now().AddDate(0, 0, -7)
	var weeklyClicks int64
	var weeklyConversions int64

	h.db.Model(&models.Click{}).
		Joins("JOIN user_offers ON clicks.user_offer_id = user_offers.id").
		Where("user_offers.offer_id = ? AND clicks.clicked_at >= ?", offerID, weekAgo).
		Count(&weeklyClicks)

	h.db.Model(&models.Conversion{}).
		Joins("JOIN user_offers ON conversions.user_offer_id = user_offers.id").
		Where("user_offers.offer_id = ? AND conversions.converted_at >= ?", offerID, weekAgo).
		Count(&weeklyConversions)

	c.JSON(http.StatusOK, gin.H{
		"offer":             offer,
		"promoters_count":   promotersCount,
		"total_clicks":      offer.TotalClicks,
		"total_conversions": offer.TotalConversions,
		"today_clicks":      todayClicks,
		"today_conversions": todayConversions,
		"weekly_clicks":     weeklyClicks,
		"weekly_conversions": weeklyConversions,
		"conversion_rate":   offer.ConversionRate(),
	})
}

// UpdateOffer updates an advertiser's offer (only if pending)
// PUT /api/advertiser/offers/:id
func (h *AdvertiserHandler) UpdateOffer(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	advertiserID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	offerIDStr := c.Param("id")
	offerID, err := uuid.Parse(offerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
		return
	}

	// Find offer and verify ownership
	var offer models.Offer
	if err := h.db.First(&offer, "id = ? AND advertiser_id = ?", offerID, advertiserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found or not owned by you"})
		return
	}

	// Only allow updates to pending or rejected offers
	if offer.Status != "pending" && offer.Status != "rejected" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot update active offers. Please contact support."})
		return
	}

	var req CreateOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	updates := map[string]interface{}{
		"title":           req.Title,
		"title_ar":        req.TitleAr,
		"description":     req.Description,
		"description_ar":  req.DescriptionAr,
		"terms_ar":        req.TermsAr,
		"image_url":       req.ImageURL,
		"logo_url":        req.LogoURL,
		"destination_url": req.DestinationURL,
		"category":        req.Category,
		"payout":          req.Payout,
		"commission":      req.Commission,
		"status":          "pending", // Reset to pending for re-review
		"rejection_reason": "",       // Clear rejection reason
		"updated_at":      time.Now(),
	}

	if req.PayoutType != "" {
		updates["payout_type"] = req.PayoutType
	}

	if err := h.db.Model(&offer).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update offer"})
		return
	}

	// Reload offer
	h.db.First(&offer, offerID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Offer updated and resubmitted for approval",
		"offer":   offer,
	})
}

// DeleteOffer deletes an advertiser's offer (only if pending)
// DELETE /api/advertiser/offers/:id
func (h *AdvertiserHandler) DeleteOffer(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	advertiserID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	offerIDStr := c.Param("id")
	offerID, err := uuid.Parse(offerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
		return
	}

	// Find offer and verify ownership
	var offer models.Offer
	if err := h.db.First(&offer, "id = ? AND advertiser_id = ?", offerID, advertiserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found or not owned by you"})
		return
	}

	// Only allow deletion of pending offers
	if offer.Status != "pending" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot delete active offers. Please contact support to deactivate."})
		return
	}

	if err := h.db.Delete(&offer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete offer"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Offer deleted successfully",
	})
}

// GetDashboard returns advertiser dashboard summary
// GET /api/advertiser/dashboard
func (h *AdvertiserHandler) GetDashboard(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	advertiserID, ok := userID.(uuid.UUID)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
		return
	}

	// Verify user is an advertiser
	var user models.AfftokUser
	if err := h.db.First(&user, "id = ?", advertiserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	// Count offers by status
	var totalOffers int64
	var pendingOffers int64
	var activeOffers int64
	var rejectedOffers int64

	h.db.Model(&models.Offer{}).Where("advertiser_id = ?", advertiserID).Count(&totalOffers)
	h.db.Model(&models.Offer{}).Where("advertiser_id = ? AND status = ?", advertiserID, "pending").Count(&pendingOffers)
	h.db.Model(&models.Offer{}).Where("advertiser_id = ? AND status = ?", advertiserID, "active").Count(&activeOffers)
	h.db.Model(&models.Offer{}).Where("advertiser_id = ? AND status = ?", advertiserID, "rejected").Count(&rejectedOffers)

	// Get total promoters across all offers
	var totalPromoters int64
	h.db.Model(&models.UserOffer{}).
		Joins("JOIN offers ON user_offers.offer_id = offers.id").
		Where("offers.advertiser_id = ?", advertiserID).
		Count(&totalPromoters)

	// Get total clicks and conversions across all offers
	var offers []models.Offer
	h.db.Where("advertiser_id = ?", advertiserID).Find(&offers)

	var totalClicks, totalConversions int
	for _, offer := range offers {
		totalClicks += offer.TotalClicks
		totalConversions += offer.TotalConversions
	}

	// Get today's stats
	today := time.Now().Truncate(24 * time.Hour)
	var todayClicks int64
	var todayConversions int64

	h.db.Model(&models.Click{}).
		Joins("JOIN user_offers ON clicks.user_offer_id = user_offers.id").
		Joins("JOIN offers ON user_offers.offer_id = offers.id").
		Where("offers.advertiser_id = ? AND clicks.clicked_at >= ?", advertiserID, today).
		Count(&todayClicks)

	h.db.Model(&models.Conversion{}).
		Joins("JOIN user_offers ON conversions.user_offer_id = user_offers.id").
		Joins("JOIN offers ON user_offers.offer_id = offers.id").
		Where("offers.advertiser_id = ? AND conversions.converted_at >= ?", advertiserID, today).
		Count(&todayConversions)

	c.JSON(http.StatusOK, gin.H{
		"advertiser": gin.H{
			"id":           user.ID,
			"company_name": user.CompanyName,
			"full_name":    user.FullName,
			"email":        user.Email,
		},
		"offers": gin.H{
			"total":    totalOffers,
			"pending":  pendingOffers,
			"active":   activeOffers,
			"rejected": rejectedOffers,
		},
		"stats": gin.H{
			"total_promoters":    totalPromoters,
			"total_clicks":       totalClicks,
			"total_conversions":  totalConversions,
			"today_clicks":       todayClicks,
			"today_conversions":  todayConversions,
		},
	})
}


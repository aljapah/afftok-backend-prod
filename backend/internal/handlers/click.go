package handlers

import (
	"net/http"

	"github.com/afftok/backend/internal/models"
	"github.com/afftok/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClickHandler struct {
	db           *gorm.DB
	clickService *services.ClickService
}

func NewClickHandler(db *gorm.DB) *ClickHandler {
	return &ClickHandler{
		db:           db,
		clickService: services.NewClickService(),
	}
}

func (h *ClickHandler) TrackClick(c *gin.Context) {
	offerID := c.Param("id")
	promoterID := c.Query("promoter")

	id, err := uuid.Parse(offerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
		return
	}

	var offer models.Offer
	if err := h.db.First(&offer, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
		return
	}

	var userOffer models.UserOffer
	if promoterID != "" {
		promoterUUID, err := uuid.Parse(promoterID)
		if err == nil {
			h.db.FirstOrCreate(&userOffer, models.UserOffer{
				UserID:  promoterUUID,
				OfferID: id,
			})
		}
	}

	if userOffer.ID != uuid.Nil {
		_, err = h.clickService.TrackClick(c, userOffer.ID)
		if err != nil {
			c.Error(err)
		}
	}

	c.Redirect(http.StatusFound, offer.DestinationURL)
}

func (h *ClickHandler) GetClickStats(c *gin.Context) {
	userOfferID := c.Param("id")
	userID, _ := c.Get("userID")

	var userOffer models.UserOffer
	if err := h.db.Where("id = ? AND user_id = ?", userOfferID, userID).First(&userOffer).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User offer not found"})
		return
	}

	var clicks []models.Click
	h.db.Where("user_offer_id = ?", userOfferID).Order("created_at DESC").Limit(100).Find(&clicks)

	deviceStats := make(map[string]int)
	browserStats := make(map[string]int)
	osStats := make(map[string]int)

	for _, click := range clicks {
		deviceStats[click.Device]++
		browserStats[click.Browser]++
		osStats[click.OS]++
	}

	c.JSON(http.StatusOK, gin.H{
		"total_clicks":  len(clicks),
		"device_stats":  deviceStats,
		"browser_stats": browserStats,
		"os_stats":      osStats,
		"recent_clicks": clicks,
	})
}

func (h *ClickHandler) GetMyClicks(c *gin.Context) {
	userID, _ := c.Get("userID")

	var userOffers []models.UserOffer
	h.db.Where("user_id = ?", userID).Find(&userOffers)

	offerIDs := make([]uuid.UUID, len(userOffers))
	for i, uo := range userOffers {
		offerIDs[i] = uo.ID
	}

	var clicks []models.Click
	h.db.Where("user_offer_id IN ?", offerIDs).Order("created_at DESC").Limit(100).Find(&clicks)

	c.JSON(http.StatusOK, gin.H{
		"clicks": clicks,
	})
}

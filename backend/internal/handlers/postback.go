package handlers

import (
	"net/http"
	"time"

	"github.com/afftok/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PostbackHandler struct {
	db *gorm.DB
}

func NewPostbackHandler(db *gorm.DB) *PostbackHandler {
	return &PostbackHandler{db: db}
}

func (h *PostbackHandler) HandlePostback(c *gin.Context) {
	type PostbackRequest struct {
		UserOfferID string `json:"user_offer_id" binding:"required"`
		ClickID     string `json:"click_id"`
		Amount      int    `json:"amount"`
		Commission  int    `json:"commission"`
		Status      string `json:"status"`
		ExternalID  string `json:"external_id"`
	}

	var req PostbackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userOfferID, err := uuid.Parse(req.UserOfferID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user offer ID"})
		return
	}

	var userOffer models.UserOffer
	if err := h.db.First(&userOffer, "id = ?", userOfferID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User offer not found"})
		return
	}

	var clickID *uuid.UUID
	if req.ClickID != "" {
		id, err := uuid.Parse(req.ClickID)
		if err == nil {
			clickID = &id
		}
	}

	conversion := models.Conversion{
		ID:                   uuid.New(),
		UserOfferID:          userOfferID,
		ClickID:              clickID,
		ExternalConversionID: req.ExternalID,
		Amount:               req.Amount,
		Commission:           req.Commission,
		Status:               req.Status,
	}

	if req.Status == "" {
		conversion.Status = "pending"
	}

	if err := h.db.Create(&conversion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversion"})
		return
	}

	h.db.Model(&userOffer).UpdateColumn("conversions", h.db.Raw("conversions + 1"))
	if conversion.Status == "approved" {
		h.db.Model(&userOffer).UpdateColumn("earnings", h.db.Raw("earnings + ?", conversion.Commission))
	}

	h.db.Model(&models.Offer{}).Where("id = ?", userOffer.OfferID).UpdateColumn("total_conversions", h.db.Raw("total_conversions + 1"))

	h.db.Model(&models.AfftokUser{}).Where("id = ?", userOffer.UserID).UpdateColumn("total_conversions", h.db.Raw("total_conversions + 1"))
	if conversion.Status == "approved" {
		h.db.Model(&models.AfftokUser{}).Where("id = ?", userOffer.UserID).UpdateColumn("total_earnings", h.db.Raw("total_earnings + ?", conversion.Commission))
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Conversion recorded successfully",
		"conversion": conversion,
	})
}

func (h *PostbackHandler) ApproveConversion(c *gin.Context) {
	conversionID := c.Param("id")

	var conversion models.Conversion
	if err := h.db.First(&conversion, "id = ?", conversionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversion not found"})
		return
	}

	if conversion.Status == "approved" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Conversion already approved"})
		return
	}

	now := time.Now()
	conversion.Status = "approved"
	conversion.ApprovedAt = &now

	if err := h.db.Save(&conversion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve conversion"})
		return
	}

	var userOffer models.UserOffer
	if err := h.db.First(&userOffer, "id = ?", conversion.UserOfferID).Error; err == nil {
		h.db.Model(&userOffer).UpdateColumn("earnings", h.db.Raw("earnings + ?", conversion.Commission))

		h.db.Model(&models.AfftokUser{}).Where("id = ?", userOffer.UserID).UpdateColumn("total_earnings", h.db.Raw("total_earnings + ?", conversion.Commission))
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Conversion approved successfully",
		"conversion": conversion,
	})
}

func (h *PostbackHandler) RejectConversion(c *gin.Context) {
	conversionID := c.Param("id")

	var conversion models.Conversion
	if err := h.db.First(&conversion, "id = ?", conversionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversion not found"})
		return
	}

	conversion.Status = "rejected"

	if err := h.db.Save(&conversion).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject conversion"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Conversion rejected successfully",
		"conversion": conversion,
	})
}

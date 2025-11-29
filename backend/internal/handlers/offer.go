package handlers

import (
    "net/http"

    "github.com/afftok/backend/internal/models"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type OfferHandler struct {
    db *gorm.DB
}

func NewOfferHandler(db *gorm.DB) *OfferHandler {
    return &OfferHandler{db: db}
}

func (h *OfferHandler) GetAllOffers(c *gin.Context) {
    var offers []models.Offer

    page := 1
    limit := 20
    offset := (page - 1) * limit

    query := h.db

    status := c.Query("status")
    if status != "" {
        query = query.Where("status = ?", status)
    } else {
        query = query.Where("status = ?", "active")
    }

    category := c.Query("category")
    if category != "" {
        query = query.Where("category = ?", category)
    }

    sortBy := c.DefaultQuery("sort", "created_at")
    order := c.DefaultQuery("order", "desc")
    query = query.Order(sortBy + " " + order)

    if err := query.Limit(limit).Offset(offset).Find(&offers).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch offers"})
        return
    }

    var total int64
    h.db.Model(&models.Offer{}).Where("status = ?", "active").Count(&total)

    c.JSON(http.StatusOK, gin.H{
        "offers": offers,
        "pagination": gin.H{
            "page":  page,
            "limit": limit,
            "total": total,
        },
    })
}

func (h *OfferHandler) GetOffer(c *gin.Context) {
    offerID := c.Param("id")

    var offer models.Offer
    if err := h.db.First(&offer, "id = ?", offerID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "offer": offer,
    })
}

func (h *OfferHandler) CreateOffer(c *gin.Context) {
    type CreateOfferRequest struct {
        Title          string `json:"title" binding:"required"`
        Description    string `json:"description"`
        ImageURL       string `json:"image_url"`
        LogoURL        string `json:"logo_url"`
        DestinationURL string `json:"destination_url" binding:"required"`
        Category       string `json:"category"`
        Payout         int    `json:"payout"`
        Commission     int    `json:"commission"`
        PayoutType     string `json:"payout_type"`
    }

    var req CreateOfferRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    offer := models.Offer{
        ID:             uuid.New(),
        Title:          req.Title,
        Description:    req.Description,
        ImageURL:       req.ImageURL,
        LogoURL:        req.LogoURL,
        DestinationURL: req.DestinationURL,
        Category:       req.Category,
        Payout:         req.Payout,
        Commission:     req.Commission,
        PayoutType:     req.PayoutType,
        Status:         "active",
    }

    if err := h.db.Create(&offer).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create offer"})
        return
    }

    c.JSON(http.StatusCreated, gin.H{
        "message": "Offer created successfully",
        "offer":   offer,
    })
}

func (h *OfferHandler) UpdateOffer(c *gin.Context) {
    offerID := c.Param("id")

    type UpdateOfferRequest struct {
        Title          string `json:"title"`
        Description    string `json:"description"`
        ImageURL       string `json:"image_url"`
        LogoURL        string `json:"logo_url"`
        DestinationURL string `json:"destination_url"`
        Category       string `json:"category"`
        Payout         int    `json:"payout"`
        Commission     int    `json:"commission"`
        Status         string `json:"status"`
    }

    var req UpdateOfferRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }

    updates := map[string]interface{}{}
    if req.Title != "" {
        updates["title"] = req.Title
    }
    if req.Description != "" {
        updates["description"] = req.Description
    }
    if req.ImageURL != "" {
        updates["image_url"] = req.ImageURL
    }
    if req.LogoURL != "" {
        updates["logo_url"] = req.LogoURL
    }
    if req.DestinationURL != "" {
        updates["destination_url"] = req.DestinationURL
    }
    if req.Category != "" {
        updates["category"] = req.Category
    }
    if req.Payout > 0 {
        updates["payout"] = req.Payout
    }
    if req.Commission > 0 {
        updates["commission"] = req.Commission
    }
    if req.Status != "" {
        updates["status"] = req.Status
    }

    if err := h.db.Model(&models.Offer{}).Where("id = ?", offerID).Updates(updates).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update offer"})
        return
    }

    var offer models.Offer
    h.db.First(&offer, "id = ?", offerID)

    c.JSON(http.StatusOK, gin.H{
        "message": "Offer updated successfully",
        "offer":   offer,
    })
}

func (h *OfferHandler) DeleteOffer(c *gin.Context) {
    offerID := c.Param("id")

    id, err := uuid.Parse(offerID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
        return
    }

    if err := h.db.Delete(&models.Offer{}, "id = ?", id).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete offer"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "message": "Offer deleted successfully",
    })
}

func (h *OfferHandler) JoinOffer(c *gin.Context) {
    offerID := c.Param("id")
    userID, _ := c.Get("userID")

    var offer models.Offer
    if err := h.db.First(&offer, "id = ?", offerID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
        return
    }

    var existingUserOffer models.UserOffer
    if err := h.db.Where("user_id = ? AND offer_id = ?", userID, offerID).First(&existingUserOffer).Error; err == nil {
        c.JSON(http.StatusConflict, gin.H{"error": "You already joined this offer"})
        return
    }

    affiliateLink := offer.DestinationURL + "?ref=" + userID.(uuid.UUID).String()

    userOffer := models.UserOffer{
        ID:            uuid.New(),
        UserID:        userID.(uuid.UUID),
        OfferID:       offer.ID,
        AffiliateLink: affiliateLink,
        Status:        "active",
    }

    if err := h.db.Create(&userOffer).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join offer"})
        return
    }

    h.db.Model(&offer).Update("users_count", offer.UsersCount+1)

    c.JSON(http.StatusCreated, gin.H{
        "message":        "Joined offer successfully",
        "user_offer":     userOffer,
        "affiliate_link": affiliateLink,
    })
}

func (h *OfferHandler) GetMyOffers(c *gin.Context) {
    userID, _ := c.Get("userID")

    var userOffers []models.UserOffer
    if err := h.db.Preload("Offer").Where("user_id = ?", userID).Find(&userOffers).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch your offers"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "offers": userOffers,
    })
}

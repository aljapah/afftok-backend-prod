package handlers

import (
    "fmt"
    "net/http"
    "strings"

    "github.com/aljapah/afftok-backend-prod/internal/models"
    "github.com/aljapah/afftok-backend-prod/internal/services"
    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"
)

type OfferHandler struct {
    db                 *gorm.DB
    linkService        *services.LinkService
    linkSigningService *services.LinkSigningService
}

func NewOfferHandler(db *gorm.DB) *OfferHandler {
    return &OfferHandler{
        db:                 db,
        linkService:        services.NewLinkService(),
        linkSigningService: services.NewLinkSigningService(),
    }
}

// SetLinkSigningService sets the link signing service (for dependency injection)
func (h *OfferHandler) SetLinkSigningService(service *services.LinkSigningService) {
    h.linkSigningService = service
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
    userUUID := userID.(uuid.UUID)

    var offer models.Offer
    if err := h.db.First(&offer, "id = ?", offerID).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
        return
    }

    // Check for existing user offer
    var existingUserOffer models.UserOffer
    if err := h.db.Where("user_id = ? AND offer_id = ?", userUUID, offerID).First(&existingUserOffer).Error; err == nil {
        c.JSON(http.StatusConflict, gin.H{
            "error":          "You already joined this offer",
            "user_offer":     existingUserOffer,
            "affiliate_link": existingUserOffer.AffiliateLink,
        })
        return
    }

    // Create new user offer ID first
    userOfferID := uuid.New()

    // Generate secure affiliate link and tracking code
    affiliateLink, shortLink, err := h.linkService.GenerateAffiliateLink(
        offer.DestinationURL,
        userOfferID,
        userUUID,
    )
    if err != nil {
        // Fallback to simple format if link generation fails
        affiliateLink = fmt.Sprintf("%s?ref=%s", offer.DestinationURL, userUUID.String())
        shortLink = ""
        fmt.Printf("[JoinOffer] Link generation error: %v, using fallback\n", err)
    }

    // Extract tracking code from short link
    trackingCode := ""
    if shortLink != "" {
        parts := strings.Split(shortLink, "/")
        if len(parts) > 0 {
            trackingCode = parts[len(parts)-1]
        }
    }

    userOffer := models.UserOffer{
        ID:            userOfferID,
        UserID:        userUUID,
        OfferID:       offer.ID,
        AffiliateLink: affiliateLink,
        ShortLink:     trackingCode, // Store just the code, not the full path
        TrackingCode:  trackingCode,
        Status:        "active",
    }

    // Use transaction for atomic operation
    err = h.db.Transaction(func(tx *gorm.DB) error {
        // Create user offer
        if err := tx.Create(&userOffer).Error; err != nil {
            return err
        }

        // Increment users_count on offer
        if err := tx.Model(&models.Offer{}).
            Where("id = ?", offer.ID).
            UpdateColumn("users_count", gorm.Expr("users_count + 1")).Error; err != nil {
            return err
        }

        return nil
    })

    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join offer"})
        return
    }

    // Generate tracking URL for mobile app
    // Use signed link for security
    var trackingURL string
    var signedLink string
    
    if trackingCode != "" && h.linkSigningService != nil {
        // Generate signed tracking link
        signedLink = h.linkSigningService.GenerateSignedLink(trackingCode)
        trackingURL = fmt.Sprintf("/api/c/%s", signedLink)
    } else if shortLink != "" {
        // Fallback to unsigned link (legacy mode)
        trackingURL = fmt.Sprintf("/api/c/%s", shortLink)
        signedLink = shortLink
    } else {
        // Ultimate fallback
        trackingURL = fmt.Sprintf("/api/c/%s?promoter=%s", offer.ID.String(), userUUID.String())
        signedLink = ""
    }

    c.JSON(http.StatusCreated, gin.H{
        "success":        true,
        "message":        "Joined offer successfully",
        "user_offer":     userOffer,
        "affiliate_link": affiliateLink,
        "tracking_url":   trackingURL,
        "short_link":     trackingCode,        // Original tracking code
        "signed_link":    signedLink,          // Signed version
    })
}

// GetPendingOffers returns all pending offers for admin review
// GET /api/admin/offers/pending
func (h *OfferHandler) GetPendingOffers(c *gin.Context) {
    var offers []models.Offer
    if err := h.db.Preload("Advertiser").Where("status = ?", "pending").Order("created_at DESC").Find(&offers).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending offers"})
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "offers": offers,
        "total":  len(offers),
    })
}

// ApproveOffer approves a pending offer
// POST /api/admin/offers/:id/approve
func (h *OfferHandler) ApproveOffer(c *gin.Context) {
    offerID := c.Param("id")

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

    if offer.Status != "pending" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Offer is not pending"})
        return
    }

    // Update status to active
    if err := h.db.Model(&offer).Updates(map[string]interface{}{
        "status":           "active",
        "rejection_reason": "",
    }).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve offer"})
        return
    }

    // Reload offer
    h.db.First(&offer, id)

    c.JSON(http.StatusOK, gin.H{
        "message": "Offer approved successfully",
        "offer":   offer,
    })
}

// RejectOffer rejects a pending offer with a reason
// POST /api/admin/offers/:id/reject
func (h *OfferHandler) RejectOffer(c *gin.Context) {
    offerID := c.Param("id")

    id, err := uuid.Parse(offerID)
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid offer ID"})
        return
    }

    type RejectRequest struct {
        Reason string `json:"reason" binding:"required"`
    }

    var req RejectRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Rejection reason is required"})
        return
    }

    var offer models.Offer
    if err := h.db.First(&offer, "id = ?", id).Error; err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Offer not found"})
        return
    }

    if offer.Status != "pending" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Offer is not pending"})
        return
    }

    // Update status to rejected
    if err := h.db.Model(&offer).Updates(map[string]interface{}{
        "status":           "rejected",
        "rejection_reason": req.Reason,
    }).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject offer"})
        return
    }

    // Reload offer
    h.db.First(&offer, id)

    c.JSON(http.StatusOK, gin.H{
        "message": "Offer rejected successfully",
        "offer":   offer,
    })
}

func (h *OfferHandler) GetMyOffers(c *gin.Context) {
    userID, _ := c.Get("userID")

    var userOffers []models.UserOffer
    if err := h.db.Preload("Offer").Where("user_id = ?", userID).Order("joined_at DESC").Find(&userOffers).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch your offers"})
        return
    }

    // Build response with stats for each user offer
    type UserOfferResponse struct {
        ID            uuid.UUID      `json:"id"`
        UserID        uuid.UUID      `json:"user_id"`
        OfferID       uuid.UUID      `json:"offer_id"`
        AffiliateLink string         `json:"affiliate_link"`
        ShortLink     string         `json:"short_link,omitempty"`
        TrackingCode  string         `json:"tracking_code,omitempty"`
        TrackingURL   string         `json:"tracking_url,omitempty"`
        Status        string         `json:"status"`
        Earnings      int            `json:"earnings"`
        JoinedAt      string         `json:"joined_at"`
        Offer         *models.Offer  `json:"offer,omitempty"`
        Stats         map[string]int `json:"stats"`
    }

    var response []UserOfferResponse
    for _, uo := range userOffers {
        // Use cached stats from UserOffer model (updated atomically during clicks/conversions)
        // Fall back to DB count if cached values are 0 (for backwards compatibility)
        clickCount := uo.TotalClicks
        conversionCount := uo.TotalConversions
        
        // If cached values are 0, query DB (for existing data before migration)
        if clickCount == 0 {
            var count int64
            h.db.Model(&models.Click{}).Where("user_offer_id = ?", uo.ID).Count(&count)
            clickCount = int(count)
        }
        if conversionCount == 0 {
            var count int64
            h.db.Model(&models.Conversion{}).Where("user_offer_id = ?", uo.ID).Count(&count)
            conversionCount = int(count)
        }

        // Build tracking URL
        trackingURL := ""
        if uo.ShortLink != "" {
            trackingURL = "/api/c/" + uo.ShortLink
        } else if uo.TrackingCode != "" {
            trackingURL = "/api/c/" + uo.TrackingCode
        } else {
            trackingURL = fmt.Sprintf("/api/c/%s?promoter=%s", uo.OfferID.String(), uo.UserID.String())
        }

        response = append(response, UserOfferResponse{
            ID:            uo.ID,
            UserID:        uo.UserID,
            OfferID:       uo.OfferID,
            AffiliateLink: uo.AffiliateLink,
            ShortLink:     uo.ShortLink,
            TrackingCode:  uo.TrackingCode,
            TrackingURL:   trackingURL,
            Status:        uo.Status,
            Earnings:      uo.Earnings,
            JoinedAt:      uo.JoinedAt.Format("2006-01-02T15:04:05Z07:00"),
            Offer:         uo.Offer,
            Stats: map[string]int{
                "clicks":      clickCount,
                "conversions": conversionCount,
                "shares":      0,
            },
        })
    }

    c.JSON(http.StatusOK, gin.H{
        "offers": response,
    })
}

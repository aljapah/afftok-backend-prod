package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConversionWebhookHandler handles incoming conversion webhooks from various platforms
type ConversionWebhookHandler struct {
	db *gorm.DB
}

// NewConversionWebhookHandler creates a new conversion webhook handler
func NewConversionWebhookHandler(db *gorm.DB) *ConversionWebhookHandler {
	return &ConversionWebhookHandler{db: db}
}

// ============================================
// GENERIC POSTBACK (for custom integrations)
// ============================================

// HandlePostback handles generic postback from any source
// GET/POST /api/postback?click_id=xxx&amount=100&order_id=yyy&status=approved
func (h *ConversionWebhookHandler) HandlePostback(c *gin.Context) {
	clickID := c.Query("click_id")
	if clickID == "" {
		clickID = c.PostForm("click_id")
	}
	
	if clickID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "click_id is required"})
		return
	}

	amount := c.DefaultQuery("amount", c.DefaultPostForm("amount", "0"))
	orderID := c.DefaultQuery("order_id", c.DefaultPostForm("order_id", ""))
	status := c.DefaultQuery("status", c.DefaultPostForm("status", "pending"))
	currency := c.DefaultQuery("currency", c.DefaultPostForm("currency", "USD"))

	amountFloat, _ := strconv.ParseFloat(amount, 64)

	// Find user offer by click_id
	userOffer, err := h.findUserOfferByClickID(clickID)
	if err != nil {
		fmt.Printf("[Postback] UserOffer not found for click_id: %s\n", clickID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid click_id"})
		return
	}

	// Create conversion record
	conversion := models.Conversion{
		ID:                   uuid.New(),
		UserOfferID:          userOffer.ID,
		ExternalConversionID: orderID,
		Amount:               int(amountFloat * 100), // Store in cents
		Currency:             currency,
		Status:               status,
		PostbackData:         fmt.Sprintf(`{"order_id":"%s","raw_amount":"%s","source":"postback"}`, orderID, amount),
		ConvertedAt:          time.Now(),
	}

	if err := h.db.Create(&conversion).Error; err != nil {
		fmt.Printf("[Postback] Failed to create conversion: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record conversion"})
		return
	}

	// Update stats
	h.updateConversionStats(userOffer.ID, userOffer.OfferID)

	fmt.Printf("[Postback] Conversion recorded: click_id=%s, amount=%s, order_id=%s\n", clickID, amount, orderID)
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"conversion_id": conversion.ID.String(),
		"message":       "Conversion recorded successfully",
	})
}

// ============================================
// SHOPIFY WEBHOOK
// ============================================

// ShopifyOrder represents a Shopify order webhook payload
type ShopifyOrder struct {
	ID              int64   `json:"id"`
	OrderNumber     int     `json:"order_number"`
	TotalPrice      string  `json:"total_price"`
	Currency        string  `json:"currency"`
	FinancialStatus string  `json:"financial_status"`
	NoteAttributes  []struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	} `json:"note_attributes"`
	LandingSite string `json:"landing_site"`
}

// HandleShopifyWebhook handles Shopify order webhooks
// POST /api/webhook/shopify/:advertiser_id
func (h *ConversionWebhookHandler) HandleShopifyWebhook(c *gin.Context) {
	advertiserID := c.Param("advertiser_id")
	
	// Verify webhook signature (if HMAC secret is set)
	// shopifyHMAC := c.GetHeader("X-Shopify-Hmac-SHA256")
	// TODO: Verify signature with advertiser's webhook secret

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	var order ShopifyOrder
	if err := json.Unmarshal(body, &order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Extract click_id from note_attributes or landing_site URL
	clickID := ""
	for _, attr := range order.NoteAttributes {
		if attr.Name == "click_id" || attr.Name == "aff_id" {
			clickID = attr.Value
			break
		}
	}

	// Try to extract from landing_site URL if not in note_attributes
	if clickID == "" && order.LandingSite != "" {
		clickID = extractClickIDFromURL(order.LandingSite)
	}

	if clickID == "" {
		fmt.Printf("[Shopify] No click_id found for order %d\n", order.OrderNumber)
		c.JSON(http.StatusOK, gin.H{"message": "No affiliate tracking found"})
		return
	}

	// Find user offer
	userOffer, err := h.findUserOfferByClickID(clickID)
	if err != nil {
		fmt.Printf("[Shopify] UserOffer not found for click_id: %s\n", clickID)
		c.JSON(http.StatusOK, gin.H{"message": "Affiliate not found"})
		return
	}

	// Verify advertiser owns this offer
	var offer models.Offer
	if err := h.db.First(&offer, "id = ?", userOffer.OfferID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "Offer not found"})
		return
	}

	if offer.AdvertiserID != nil && offer.AdvertiserID.String() != advertiserID {
		fmt.Printf("[Shopify] Advertiser mismatch: expected %s, got %s\n", advertiserID, offer.AdvertiserID.String())
	}

	// Parse amount
	amountFloat, _ := strconv.ParseFloat(order.TotalPrice, 64)

	// Create conversion
	conversion := models.Conversion{
		ID:                   uuid.New(),
		UserOfferID:          userOffer.ID,
		ExternalConversionID: fmt.Sprintf("shopify_%d", order.ID),
		Amount:               int(amountFloat * 100),
		Currency:             order.Currency,
		Status:               mapShopifyStatus(order.FinancialStatus),
		PostbackData:         string(body),
		ConvertedAt:          time.Now(),
	}

	if err := h.db.Create(&conversion).Error; err != nil {
		fmt.Printf("[Shopify] Failed to create conversion: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record conversion"})
		return
	}

	h.updateConversionStats(userOffer.ID, userOffer.OfferID)

	fmt.Printf("[Shopify] Conversion recorded: order=%d, click_id=%s, amount=%s\n", order.OrderNumber, clickID, order.TotalPrice)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ============================================
// SALLA WEBHOOK (سلة)
// ============================================

// SallaOrder represents a Salla order webhook payload
type SallaOrder struct {
	Event string `json:"event"`
	Data  struct {
		ID          int    `json:"id"`
		ReferenceID string `json:"reference_id"`
		Total       struct {
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
		} `json:"total"`
		Status   string `json:"status"`
		Metadata []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"metadata"`
		URLs struct {
			Customer string `json:"customer"`
		} `json:"urls"`
	} `json:"data"`
}

// HandleSallaWebhook handles Salla order webhooks
// POST /api/webhook/salla/:advertiser_id
func (h *ConversionWebhookHandler) HandleSallaWebhook(c *gin.Context) {
	advertiserID := c.Param("advertiser_id")
	_ = advertiserID // Used for verification

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	var order SallaOrder
	if err := json.Unmarshal(body, &order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Only process order.created or order.updated events
	if order.Event != "order.created" && order.Event != "order.updated" {
		c.JSON(http.StatusOK, gin.H{"message": "Event ignored"})
		return
	}

	// Extract click_id from metadata
	clickID := ""
	for _, meta := range order.Data.Metadata {
		if meta.Key == "click_id" || meta.Key == "aff_id" {
			clickID = meta.Value
			break
		}
	}

	// Try to extract from customer URL if not in metadata
	if clickID == "" && order.Data.URLs.Customer != "" {
		clickID = extractClickIDFromURL(order.Data.URLs.Customer)
	}

	if clickID == "" {
		fmt.Printf("[Salla] No click_id found for order %s\n", order.Data.ReferenceID)
		c.JSON(http.StatusOK, gin.H{"message": "No affiliate tracking found"})
		return
	}

	// Find user offer
	userOffer, err := h.findUserOfferByClickID(clickID)
	if err != nil {
		fmt.Printf("[Salla] UserOffer not found for click_id: %s\n", clickID)
		c.JSON(http.StatusOK, gin.H{"message": "Affiliate not found"})
		return
	}

	// Create conversion
	conversion := models.Conversion{
		ID:                   uuid.New(),
		UserOfferID:          userOffer.ID,
		ExternalConversionID: fmt.Sprintf("salla_%d", order.Data.ID),
		Amount:               int(order.Data.Total.Amount * 100),
		Currency:             order.Data.Total.Currency,
		Status:               mapSallaStatus(order.Data.Status),
		PostbackData:         string(body),
		ConvertedAt:          time.Now(),
	}

	if err := h.db.Create(&conversion).Error; err != nil {
		fmt.Printf("[Salla] Failed to create conversion: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record conversion"})
		return
	}

	h.updateConversionStats(userOffer.ID, userOffer.OfferID)

	fmt.Printf("[Salla] Conversion recorded: order=%s, click_id=%s, amount=%.2f\n", order.Data.ReferenceID, clickID, order.Data.Total.Amount)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ============================================
// ZID WEBHOOK (زد)
// ============================================

// ZidOrder represents a Zid order webhook payload
type ZidOrder struct {
	ID         string  `json:"id"`
	OrderID    string  `json:"order_id"`
	TotalPrice float64 `json:"total_price"`
	Currency   string  `json:"currency"`
	Status     string  `json:"status"`
	Meta       map[string]string `json:"meta"`
	RefURL     string  `json:"ref_url"`
}

// HandleZidWebhook handles Zid order webhooks
// POST /api/webhook/zid/:advertiser_id
func (h *ConversionWebhookHandler) HandleZidWebhook(c *gin.Context) {
	advertiserID := c.Param("advertiser_id")
	_ = advertiserID

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read body"})
		return
	}

	var order ZidOrder
	if err := json.Unmarshal(body, &order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON"})
		return
	}

	// Extract click_id from meta or ref_url
	clickID := ""
	if order.Meta != nil {
		if id, ok := order.Meta["click_id"]; ok {
			clickID = id
		} else if id, ok := order.Meta["aff_id"]; ok {
			clickID = id
		}
	}

	if clickID == "" && order.RefURL != "" {
		clickID = extractClickIDFromURL(order.RefURL)
	}

	if clickID == "" {
		fmt.Printf("[Zid] No click_id found for order %s\n", order.OrderID)
		c.JSON(http.StatusOK, gin.H{"message": "No affiliate tracking found"})
		return
	}

	userOffer, err := h.findUserOfferByClickID(clickID)
	if err != nil {
		fmt.Printf("[Zid] UserOffer not found for click_id: %s\n", clickID)
		c.JSON(http.StatusOK, gin.H{"message": "Affiliate not found"})
		return
	}

	conversion := models.Conversion{
		ID:                   uuid.New(),
		UserOfferID:          userOffer.ID,
		ExternalConversionID: fmt.Sprintf("zid_%s", order.OrderID),
		Amount:               int(order.TotalPrice * 100),
		Currency:             order.Currency,
		Status:               mapZidStatus(order.Status),
		PostbackData:         string(body),
		ConvertedAt:          time.Now(),
	}

	if err := h.db.Create(&conversion).Error; err != nil {
		fmt.Printf("[Zid] Failed to create conversion: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record conversion"})
		return
	}

	h.updateConversionStats(userOffer.ID, userOffer.OfferID)

	fmt.Printf("[Zid] Conversion recorded: order=%s, click_id=%s, amount=%.2f\n", order.OrderID, clickID, order.TotalPrice)
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ============================================
// PIXEL TRACKING (JavaScript Pixel)
// ============================================

// HandlePixelConversion handles conversion from JavaScript pixel
// POST /api/pixel/convert
func (h *ConversionWebhookHandler) HandlePixelConversion(c *gin.Context) {
	type PixelRequest struct {
		ClickID  string  `json:"click_id"`
		Amount   float64 `json:"amount"`
		Currency string  `json:"currency"`
		OrderID  string  `json:"order_id"`
	}

	var req PixelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Try to get from query params (for img pixel)
		req.ClickID = c.Query("click_id")
		req.Amount, _ = strconv.ParseFloat(c.Query("amount"), 64)
		req.Currency = c.DefaultQuery("currency", "USD")
		req.OrderID = c.Query("order_id")
	}

	if req.ClickID == "" {
		// Try to get from cookie
		req.ClickID, _ = c.Cookie("afftok_click_id")
	}

	if req.ClickID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "click_id is required"})
		return
	}

	userOffer, err := h.findUserOfferByClickID(req.ClickID)
	if err != nil {
		fmt.Printf("[Pixel] UserOffer not found for click_id: %s\n", req.ClickID)
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid click_id"})
		return
	}

	conversion := models.Conversion{
		ID:                   uuid.New(),
		UserOfferID:          userOffer.ID,
		ExternalConversionID: req.OrderID,
		Amount:               int(req.Amount * 100),
		Currency:             req.Currency,
		Status:               "pending",
		PostbackData:         fmt.Sprintf(`{"order_id":"%s","ip":"%s","source":"pixel"}`, req.OrderID, c.ClientIP()),
		ConvertedAt:          time.Now(),
	}

	if err := h.db.Create(&conversion).Error; err != nil {
		fmt.Printf("[Pixel] Failed to create conversion: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to record conversion"})
		return
	}

	h.updateConversionStats(userOffer.ID, userOffer.OfferID)

	fmt.Printf("[Pixel] Conversion recorded: click_id=%s, amount=%.2f\n", req.ClickID, req.Amount)
	
	// Return 1x1 transparent GIF for img pixel
	if c.Query("img") == "1" {
		c.Header("Content-Type", "image/gif")
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		// 1x1 transparent GIF
		gif := []byte{0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0x00, 0x00, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x21, 0xf9, 0x04, 0x01, 0x00, 0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00, 0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3b}
		c.Data(http.StatusOK, "image/gif", gif)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"conversion_id": conversion.ID.String(),
	})
}

// ============================================
// HELPER FUNCTIONS
// ============================================

// findUserOfferByClickID finds user offer by click_id (which is the UserOffer ID)
func (h *ConversionWebhookHandler) findUserOfferByClickID(clickID string) (*models.UserOffer, error) {
	var userOffer models.UserOffer

	// Try as UUID first
	userOfferID, err := uuid.Parse(clickID)
	if err == nil {
		if err := h.db.First(&userOffer, "id = ?", userOfferID).Error; err == nil {
			return &userOffer, nil
		}
	}

	// Try as tracking code
	if err := h.db.Where("tracking_code = ? OR short_link = ?", clickID, clickID).First(&userOffer).Error; err == nil {
		return &userOffer, nil
	}

	return nil, fmt.Errorf("user offer not found for click_id: %s", clickID)
}

// updateConversionStats updates conversion counts on user_offer and offer
func (h *ConversionWebhookHandler) updateConversionStats(userOfferID, offerID uuid.UUID) {
	// Update user_offer stats
	h.db.Model(&models.UserOffer{}).Where("id = ?", userOfferID).
		UpdateColumn("total_conversions", gorm.Expr("total_conversions + 1"))

	// Update offer stats
	h.db.Model(&models.Offer{}).Where("id = ?", offerID).
		UpdateColumn("total_conversions", gorm.Expr("total_conversions + 1"))
}

// extractClickIDFromURL extracts click_id from URL query parameters
func extractClickIDFromURL(rawURL string) string {
	// Simple extraction - look for click_id= or aff_id=
	if idx := strings.Index(rawURL, "click_id="); idx != -1 {
		value := rawURL[idx+9:]
		if endIdx := strings.IndexAny(value, "&# "); endIdx != -1 {
			return value[:endIdx]
		}
		return value
	}
	if idx := strings.Index(rawURL, "aff_id="); idx != -1 {
		value := rawURL[idx+7:]
		if endIdx := strings.IndexAny(value, "&# "); endIdx != -1 {
			return value[:endIdx]
		}
		return value
	}
	return ""
}

// mapShopifyStatus maps Shopify financial status to our status
func mapShopifyStatus(status string) string {
	switch status {
	case "paid":
		return "approved"
	case "pending":
		return "pending"
	case "refunded", "voided":
		return "rejected"
	default:
		return "pending"
	}
}

// mapSallaStatus maps Salla order status to our status
func mapSallaStatus(status string) string {
	switch status {
	case "completed", "delivered":
		return "approved"
	case "pending", "processing":
		return "pending"
	case "cancelled", "refunded":
		return "rejected"
	default:
		return "pending"
	}
}

// mapZidStatus maps Zid order status to our status
func mapZidStatus(status string) string {
	switch status {
	case "completed", "delivered":
		return "approved"
	case "pending", "processing":
		return "pending"
	case "cancelled", "refunded":
		return "rejected"
	default:
		return "pending"
	}
}

// verifyHMAC verifies HMAC signature
func verifyHMAC(message, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	expectedMAC := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expectedMAC), []byte(signature))
}


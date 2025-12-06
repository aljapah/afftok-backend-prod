package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type InvoiceHandler struct {
	db *gorm.DB
}

func NewInvoiceHandler(db *gorm.DB) *InvoiceHandler {
	return &InvoiceHandler{db: db}
}

// GetMyInvoices returns invoices for the authenticated advertiser
func (h *InvoiceHandler) GetMyInvoices(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var invoices []models.Invoice
	if err := h.db.Where("advertiser_id = ?", userID).
		Order("year DESC, month DESC").
		Find(&invoices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invoices"})
		return
	}

	// Calculate summary
	var summary models.InvoiceSummary
	for _, inv := range invoices {
		summary.TotalInvoices++
		summary.TotalAmount += inv.PlatformAmount
		switch inv.Status {
		case "paid":
			summary.PaidAmount += inv.PlatformAmount
		case "pending":
			summary.PendingAmount += inv.PlatformAmount
			summary.PendingCount++
		case "overdue":
			summary.OverdueAmount += inv.PlatformAmount
			summary.OverdueCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"invoices": invoices,
		"summary":  summary,
	})
}

// GetInvoice returns a specific invoice
func (h *InvoiceHandler) GetInvoice(c *gin.Context) {
	invoiceID := c.Param("id")
	userID, _ := c.Get("userID")

	var invoice models.Invoice
	if err := h.db.Where("id = ? AND advertiser_id = ?", invoiceID, userID).
		First(&invoice).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	// Get invoice items
	var items []models.InvoiceItem
	h.db.Where("invoice_id = ?", invoiceID).Find(&items)

	c.JSON(http.StatusOK, gin.H{
		"invoice": invoice,
		"items":   items,
	})
}

// ConfirmPayment allows advertiser to confirm payment and upload proof
func (h *InvoiceHandler) ConfirmPayment(c *gin.Context) {
	invoiceID := c.Param("id")
	userID, _ := c.Get("userID")

	var req struct {
		PaymentProof  string `json:"payment_proof"`
		PaymentMethod string `json:"payment_method"`
		PaymentNote   string `json:"payment_note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var invoice models.Invoice
	if err := h.db.Where("id = ? AND advertiser_id = ?", invoiceID, userID).
		First(&invoice).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	if invoice.Status == "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invoice already paid"})
		return
	}

	// Update invoice with payment info (status remains pending until admin confirms)
	now := time.Now()
	invoice.PaymentProof = req.PaymentProof
	invoice.PaymentMethod = req.PaymentMethod
	invoice.PaymentNote = req.PaymentNote
	invoice.PaidAt = &now
	invoice.Status = "pending_confirmation" // Waiting for admin to confirm

	if err := h.db.Save(&invoice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invoice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment submitted for confirmation",
		"invoice": invoice,
	})
}

// ============ ADMIN ENDPOINTS ============

// AdminGetAllInvoices returns all invoices for admin
func (h *InvoiceHandler) AdminGetAllInvoices(c *gin.Context) {
	status := c.Query("status")
	month := c.Query("month")
	year := c.Query("year")

	query := h.db.Preload("Advertiser").Order("created_at DESC")

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if month != "" {
		if m, err := strconv.Atoi(month); err == nil {
			query = query.Where("month = ?", m)
		}
	}
	if year != "" {
		if y, err := strconv.Atoi(year); err == nil {
			query = query.Where("year = ?", y)
		}
	}

	var invoices []models.Invoice
	if err := query.Find(&invoices).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch invoices"})
		return
	}

	// Calculate summary
	var summary models.InvoiceSummary
	for _, inv := range invoices {
		summary.TotalInvoices++
		summary.TotalAmount += inv.PlatformAmount
		switch inv.Status {
		case "paid":
			summary.PaidAmount += inv.PlatformAmount
		case "pending", "pending_confirmation":
			summary.PendingAmount += inv.PlatformAmount
			summary.PendingCount++
		case "overdue":
			summary.OverdueAmount += inv.PlatformAmount
			summary.OverdueCount++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"invoices": invoices,
		"summary":  summary,
	})
}

// AdminConfirmPayment confirms an invoice payment
func (h *InvoiceHandler) AdminConfirmPayment(c *gin.Context) {
	invoiceID := c.Param("id")
	adminID, _ := c.Get("userID")

	var req struct {
		ReviewNote string `json:"review_note"`
	}
	c.ShouldBindJSON(&req)

	var invoice models.Invoice
	if err := h.db.First(&invoice, "id = ?", invoiceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	now := time.Now()
	adminUUID := adminID.(uuid.UUID)
	invoice.Status = "paid"
	invoice.ReviewedBy = &adminUUID
	invoice.ReviewedAt = &now
	invoice.ReviewNote = req.ReviewNote

	if err := h.db.Save(&invoice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invoice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment confirmed",
		"invoice": invoice,
	})
}

// AdminRejectPayment rejects a payment (e.g., invalid proof)
func (h *InvoiceHandler) AdminRejectPayment(c *gin.Context) {
	invoiceID := c.Param("id")
	adminID, _ := c.Get("userID")

	var req struct {
		ReviewNote string `json:"review_note" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Review note is required"})
		return
	}

	var invoice models.Invoice
	if err := h.db.First(&invoice, "id = ?", invoiceID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invoice not found"})
		return
	}

	now := time.Now()
	adminUUID := adminID.(uuid.UUID)
	invoice.Status = "pending" // Back to pending
	invoice.PaymentProof = ""
	invoice.PaidAt = nil
	invoice.ReviewedBy = &adminUUID
	invoice.ReviewedAt = &now
	invoice.ReviewNote = req.ReviewNote

	if err := h.db.Save(&invoice).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update invoice"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Payment rejected",
		"invoice": invoice,
	})
}

// AdminGenerateMonthlyInvoices generates invoices for all advertisers
func (h *InvoiceHandler) AdminGenerateMonthlyInvoices(c *gin.Context) {
	var req struct {
		Month int `json:"month" binding:"required,min=1,max=12"`
		Year  int `json:"year" binding:"required,min=2024"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid month/year"})
		return
	}

	// Get all advertisers
	var advertisers []models.AfftokUser
	if err := h.db.Where("role = ? AND status = ?", "advertiser", "active").
		Find(&advertisers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch advertisers"})
		return
	}

	platformRate := 0.10 // 10%
	createdCount := 0
	skippedCount := 0

	for _, advertiser := range advertisers {
		// Check if invoice already exists
		var existing models.Invoice
		if err := h.db.Where("advertiser_id = ? AND month = ? AND year = ?",
			advertiser.ID, req.Month, req.Year).First(&existing).Error; err == nil {
			skippedCount++
			continue
		}

		// Calculate total conversions and payouts for this advertiser's offers
		var result struct {
			TotalConversions int
			TotalPayout      float64
		}
		
		// Get conversions for advertiser's offers in this period
		periodStart := time.Date(req.Year, time.Month(req.Month), 1, 0, 0, 0, 0, time.UTC)
		periodEnd := periodStart.AddDate(0, 1, 0).Add(-time.Second)
		
		h.db.Table("conversions").
			Select("COUNT(*) as total_conversions, COALESCE(SUM(payout), 0) as total_payout").
			Joins("JOIN offers ON conversions.offer_id = offers.id").
			Where("offers.advertiser_id = ? AND conversions.created_at BETWEEN ? AND ?",
				advertiser.ID, periodStart, periodEnd).
			Scan(&result)

		// Skip if no conversions
		if result.TotalConversions == 0 {
			skippedCount++
			continue
		}

		// Create invoice
		invoice := models.Invoice{
			AdvertiserID:        advertiser.ID,
			Month:               req.Month,
			Year:                req.Year,
			PeriodStart:         periodStart,
			PeriodEnd:           periodEnd,
			TotalConversions:    result.TotalConversions,
			TotalPromoterPayout: result.TotalPayout,
			PlatformRate:        platformRate,
			PlatformAmount:      result.TotalPayout * platformRate,
			Currency:            "KWD",
			Status:              "pending",
			DueDate:             periodEnd.AddDate(0, 0, 7), // Due 7 days after period end
		}

		if err := h.db.Create(&invoice).Error; err != nil {
			continue
		}
		createdCount++
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Invoice generation completed",
		"created_count": createdCount,
		"skipped_count": skippedCount,
	})
}

// AdminGetInvoiceSummary returns overall invoice statistics
func (h *InvoiceHandler) AdminGetInvoiceSummary(c *gin.Context) {
	var summary struct {
		TotalInvoices    int64   `json:"total_invoices"`
		TotalAmount      float64 `json:"total_amount"`
		PaidAmount       float64 `json:"paid_amount"`
		PendingAmount    float64 `json:"pending_amount"`
		OverdueAmount    float64 `json:"overdue_amount"`
		ThisMonthAmount  float64 `json:"this_month_amount"`
		ThisMonthPending int64   `json:"this_month_pending"`
	}

	h.db.Model(&models.Invoice{}).Count(&summary.TotalInvoices)
	h.db.Model(&models.Invoice{}).Select("COALESCE(SUM(platform_amount), 0)").Scan(&summary.TotalAmount)
	h.db.Model(&models.Invoice{}).Where("status = ?", "paid").
		Select("COALESCE(SUM(platform_amount), 0)").Scan(&summary.PaidAmount)
	h.db.Model(&models.Invoice{}).Where("status IN ?", []string{"pending", "pending_confirmation"}).
		Select("COALESCE(SUM(platform_amount), 0)").Scan(&summary.PendingAmount)
	h.db.Model(&models.Invoice{}).Where("status = ?", "overdue").
		Select("COALESCE(SUM(platform_amount), 0)").Scan(&summary.OverdueAmount)

	// This month
	now := time.Now()
	h.db.Model(&models.Invoice{}).
		Where("month = ? AND year = ?", int(now.Month()), now.Year()).
		Select("COALESCE(SUM(platform_amount), 0)").Scan(&summary.ThisMonthAmount)
	h.db.Model(&models.Invoice{}).
		Where("month = ? AND year = ? AND status IN ?", int(now.Month()), now.Year(), []string{"pending", "pending_confirmation"}).
		Count(&summary.ThisMonthPending)

	c.JSON(http.StatusOK, summary)
}


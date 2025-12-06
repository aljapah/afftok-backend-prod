package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Invoice represents a monthly invoice for an advertiser
type Invoice struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key" json:"id"`
	AdvertiserID    uuid.UUID  `gorm:"type:uuid;not null;index" json:"advertiser_id"`
	Advertiser      *AfftokUser `gorm:"foreignKey:AdvertiserID" json:"advertiser,omitempty"`
	
	// Invoice period
	Month           int        `json:"month"` // 1-12
	Year            int        `json:"year"`
	PeriodStart     time.Time  `json:"period_start"`
	PeriodEnd       time.Time  `json:"period_end"`
	
	// Financial details
	TotalConversions    int     `json:"total_conversions"`
	TotalPromoterPayout float64 `json:"total_promoter_payout"` // Total paid to promoters
	PlatformRate        float64 `json:"platform_rate"`          // 0.10 = 10%
	PlatformAmount      float64 `json:"platform_amount"`        // Amount owed to platform
	Currency            string  `gorm:"default:'KWD'" json:"currency"`
	
	// Status
	Status          string     `gorm:"default:'pending'" json:"status"` // pending, paid, overdue, cancelled
	DueDate         time.Time  `json:"due_date"`
	PaidAt          *time.Time `json:"paid_at,omitempty"`
	PaymentProof    string     `json:"payment_proof,omitempty"` // URL to uploaded receipt
	PaymentMethod   string     `json:"payment_method,omitempty"` // bank_transfer, etc.
	PaymentNote     string     `json:"payment_note,omitempty"`
	
	// Admin review
	ReviewedBy      *uuid.UUID `gorm:"type:uuid" json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time `json:"reviewed_at,omitempty"`
	ReviewNote      string     `json:"review_note,omitempty"`
	
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// BeforeCreate generates UUID before creating
func (i *Invoice) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}

// InvoiceItem represents a line item in an invoice (optional detail)
type InvoiceItem struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	InvoiceID    uuid.UUID `gorm:"type:uuid;not null;index" json:"invoice_id"`
	OfferID      uuid.UUID `gorm:"type:uuid" json:"offer_id"`
	OfferTitle   string    `json:"offer_title"`
	Conversions  int       `json:"conversions"`
	PromoterPayout float64 `json:"promoter_payout"`
	PlatformAmount float64 `json:"platform_amount"`
	CreatedAt    time.Time `json:"created_at"`
}

func (ii *InvoiceItem) BeforeCreate(tx *gorm.DB) error {
	if ii.ID == uuid.Nil {
		ii.ID = uuid.New()
	}
	return nil
}

// InvoiceSummary for dashboard display
type InvoiceSummary struct {
	TotalInvoices     int     `json:"total_invoices"`
	TotalAmount       float64 `json:"total_amount"`
	PaidAmount        float64 `json:"paid_amount"`
	PendingAmount     float64 `json:"pending_amount"`
	OverdueAmount     float64 `json:"overdue_amount"`
	PendingCount      int     `json:"pending_count"`
	OverdueCount      int     `json:"overdue_count"`
}


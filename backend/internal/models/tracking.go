package models

import (
	"time"

	"github.com/google/uuid"
)

// Click represents a single click on an affiliate link
type Click struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserOfferID uuid.UUID  `gorm:"type:uuid;not null;index:idx_clicks_user_offer" json:"user_offer_id"`
	IPAddress   string     `gorm:"type:varchar(45);index:idx_clicks_ip" json:"ip_address,omitempty"`
	UserAgent   string     `gorm:"type:text" json:"user_agent,omitempty"`
	Device      string     `gorm:"type:varchar(50);index:idx_clicks_device" json:"device,omitempty"`
	Browser     string     `gorm:"type:varchar(50)" json:"browser,omitempty"`
	OS          string     `gorm:"type:varchar(50)" json:"os,omitempty"`
	Country     string     `gorm:"type:varchar(2);index:idx_clicks_country" json:"country,omitempty"`
	City        string     `gorm:"type:varchar(100)" json:"city,omitempty"`
	Referrer    string     `gorm:"type:text" json:"referrer,omitempty"`
	ClickedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP;index:idx_clicks_time" json:"clicked_at"`
	
	// Deduplication and tracking
	Fingerprint string     `gorm:"type:varchar(64);index:idx_clicks_fingerprint" json:"fingerprint,omitempty"`
	IsUnique    bool       `gorm:"default:true" json:"is_unique"`
	
	// Relationships
	UserOffer   *UserOffer `gorm:"foreignKey:UserOfferID" json:"user_offer,omitempty"`
}

func (Click) TableName() string {
	return "clicks"
}

// Conversion represents a successful conversion from a click
type Conversion struct {
	ID                   uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserOfferID          uuid.UUID  `gorm:"type:uuid;not null;index:idx_conv_user_offer" json:"user_offer_id"`
	ClickID              *uuid.UUID `gorm:"type:uuid;index:idx_conv_click" json:"click_id,omitempty"`
	
	// External tracking
	ExternalConversionID string     `gorm:"type:varchar(100);uniqueIndex:idx_conv_external_unique" json:"external_conversion_id,omitempty"`
	NetworkID            *uuid.UUID `gorm:"type:uuid;index:idx_conv_network" json:"network_id,omitempty"`
	
	// Financial data
	Amount               int        `gorm:"default:0" json:"amount"`
	Commission           int        `gorm:"default:0" json:"commission"`
	Currency             string     `gorm:"type:varchar(3);default:'USD'" json:"currency"`
	
	// Status tracking
	Status               string     `gorm:"type:varchar(20);default:'pending';index:idx_conv_status" json:"status"`
	RejectionReason      string     `gorm:"type:text" json:"rejection_reason,omitempty"`
	
	// Timestamps
	ConvertedAt          time.Time  `gorm:"default:CURRENT_TIMESTAMP;index:idx_conv_time" json:"converted_at"`
	ApprovedAt           *time.Time `json:"approved_at,omitempty"`
	PaidAt               *time.Time `json:"paid_at,omitempty"`
	
	// Postback data
	PostbackData         string     `gorm:"type:jsonb" json:"postback_data,omitempty"`
	PostbackReceivedAt   *time.Time `json:"postback_received_at,omitempty"`
	
	// Relationships
	UserOffer            *UserOffer `gorm:"foreignKey:UserOfferID" json:"user_offer,omitempty"`
	Click                *Click     `gorm:"foreignKey:ClickID" json:"click,omitempty"`
}

func (Conversion) TableName() string {
	return "conversions"
}

// ConversionStatus constants
const (
	ConversionStatusPending  = "pending"
	ConversionStatusApproved = "approved"
	ConversionStatusRejected = "rejected"
	ConversionStatusPaid     = "paid"
)

// IsValid checks if conversion status is valid
func (c *Conversion) IsValid() bool {
	validStatuses := map[string]bool{
		ConversionStatusPending:  true,
		ConversionStatusApproved: true,
		ConversionStatusRejected: true,
		ConversionStatusPaid:     true,
	}
	return validStatuses[c.Status]
}

// CanApprove checks if conversion can be approved
func (c *Conversion) CanApprove() bool {
	return c.Status == ConversionStatusPending
}

// CanReject checks if conversion can be rejected
func (c *Conversion) CanReject() bool {
	return c.Status == ConversionStatusPending
}

// Note: Team and TeamMember are defined in team.go

// Badge represents an achievement badge
type Badge struct {
	ID            uuid.UUID   `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name          string      `gorm:"type:varchar(100);not null" json:"name"`
	Description   string      `gorm:"type:text" json:"description,omitempty"`
	IconURL       string      `gorm:"type:text;column:icon_url" json:"icon_url,omitempty"`
	Criteria      string      `gorm:"type:text" json:"criteria,omitempty"`
	PointsReward  int         `gorm:"default:0;column:points_reward" json:"points_reward"`
	RequiredValue int         `gorm:"default:0;column:required_value" json:"required_value"`
	Points        int         `gorm:"default:0" json:"points"`
	CreatedAt     time.Time   `gorm:"default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
	UserBadges    []UserBadge `gorm:"foreignKey:BadgeID" json:"user_badges,omitempty"`
}

func (Badge) TableName() string {
	return "badges"
}

// UserBadge represents a badge earned by a user
type UserBadge struct {
	ID       uuid.UUID   `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID   uuid.UUID   `gorm:"type:uuid;not null;column:user_id;index:idx_user_badge_user" json:"user_id"`
	BadgeID  uuid.UUID   `gorm:"type:uuid;not null;column:badge_id;index:idx_user_badge_badge" json:"badge_id"`
	EarnedAt time.Time   `gorm:"default:CURRENT_TIMESTAMP;column:earned_at" json:"earned_at"`
	User     *AfftokUser `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Badge    *Badge      `gorm:"foreignKey:BadgeID" json:"badge,omitempty"`
}

func (UserBadge) TableName() string {
	return "user_badges"
}

// TrackingEvent represents a generic tracking event for analytics
type TrackingEvent struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	EventType   string     `gorm:"type:varchar(50);not null;index:idx_event_type" json:"event_type"`
	UserID      *uuid.UUID `gorm:"type:uuid;index:idx_event_user" json:"user_id,omitempty"`
	OfferID     *uuid.UUID `gorm:"type:uuid;index:idx_event_offer" json:"offer_id,omitempty"`
	UserOfferID *uuid.UUID `gorm:"type:uuid;index:idx_event_user_offer" json:"user_offer_id,omitempty"`
	Data        string     `gorm:"type:jsonb" json:"data,omitempty"`
	IPAddress   string     `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
	UserAgent   string     `gorm:"type:text" json:"user_agent,omitempty"`
	CreatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP;index:idx_event_time" json:"created_at"`
}

func (TrackingEvent) TableName() string {
	return "tracking_events"
}

// Event type constants
const (
	EventTypeClick      = "click"
	EventTypeVisit      = "visit"
	EventTypeConversion = "conversion"
	EventTypeShare      = "share"
	EventTypeView       = "view"
)

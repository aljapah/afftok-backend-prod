package models

import (
	"time"

	"github.com/google/uuid"
)

type Network struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	LogoURL     string    `gorm:"type:text" json:"logo_url,omitempty"`
	APIURL      string    `gorm:"type:text" json:"api_url,omitempty"`
	APIKey      string    `gorm:"type:text" json:"-"`
	PostbackURL string    `gorm:"type:text" json:"postback_url,omitempty"`
	HMACSecret  string    `gorm:"type:text" json:"-"`
	Status      string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Offers      []Offer   `gorm:"foreignKey:NetworkID" json:"offers,omitempty"`
}

func (Network) TableName() string {
	return "networks"
}

type Offer struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	NetworkID        *uuid.UUID `gorm:"type:uuid" json:"network_id,omitempty"`
	AdvertiserID     *uuid.UUID `gorm:"type:uuid;index" json:"advertiser_id,omitempty"` // NEW: Link to advertiser user
	ExternalOfferID  string     `gorm:"type:varchar(100)" json:"external_offer_id,omitempty"`
	
	// English Fields
	Title            string     `gorm:"type:varchar(255);not null" json:"title"`
	Description      string     `gorm:"type:text" json:"description,omitempty"`
	
	// Arabic Fields (للمستخدم العربي)
	TitleAr          string     `gorm:"type:varchar(255)" json:"title_ar,omitempty"`
	DescriptionAr    string     `gorm:"type:text" json:"description_ar,omitempty"`
	TermsAr          string     `gorm:"type:text" json:"terms_ar,omitempty"`
	
	// Common Fields
	ImageURL         string     `gorm:"type:text" json:"image_url,omitempty"`
	LogoURL          string     `gorm:"type:text" json:"logo_url,omitempty"`
	DestinationURL   string     `gorm:"type:text;not null" json:"destination_url"`
	Category         string     `gorm:"type:varchar(50)" json:"category,omitempty"`
	Payout           int        `gorm:"default:0" json:"payout"`
	Commission       int        `gorm:"default:0" json:"commission"`
	PayoutType       string     `gorm:"type:varchar(20);default:'cpa'" json:"payout_type"`
	Rating           float64    `gorm:"type:decimal(3,2);default:0.0" json:"rating"`
	UsersCount       int        `gorm:"default:0" json:"users_count"`
	Status           string     `gorm:"type:varchar(20);default:'pending'" json:"status"` // pending, active, rejected, paused
	RejectionReason  string     `gorm:"type:text" json:"rejection_reason,omitempty"`      // NEW: Reason if rejected
	TotalClicks      int        `gorm:"default:0" json:"total_clicks"`
	TotalConversions int        `gorm:"default:0" json:"total_conversions"`
	CreatedAt        time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Network          *Network    `gorm:"foreignKey:NetworkID" json:"network,omitempty"`
	Advertiser       *AfftokUser `gorm:"foreignKey:AdvertiserID" json:"advertiser,omitempty"` // NEW: Advertiser relationship
	UserOffers       []UserOffer `gorm:"foreignKey:OfferID" json:"user_offers,omitempty"`
}

// GetTitle returns title based on language preference
func (o *Offer) GetTitle(lang string) string {
	if lang == "ar" && o.TitleAr != "" {
		return o.TitleAr
	}
	return o.Title
}

// GetDescription returns description based on language preference
func (o *Offer) GetDescription(lang string) string {
	if lang == "ar" && o.DescriptionAr != "" {
		return o.DescriptionAr
	}
	return o.Description
}

func (Offer) TableName() string {
	return "offers"
}

func (o *Offer) ConversionRate() float64 {
	if o.TotalClicks == 0 {
		return 0
	}
	return (float64(o.TotalConversions) / float64(o.TotalClicks)) * 100
}

type UserOffer struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;not null;index:idx_user_offers_user" json:"user_id"`
	OfferID       uuid.UUID `gorm:"type:uuid;not null;index:idx_user_offers_offer" json:"offer_id"`
	AffiliateLink string    `gorm:"type:text;not null" json:"affiliate_link"`
	ShortLink     string    `gorm:"type:text;index:idx_user_offers_short_link" json:"short_link,omitempty"`
	TrackingCode  string    `gorm:"type:varchar(32);index:idx_user_offers_tracking" json:"tracking_code,omitempty"`
	Status        string    `gorm:"type:varchar(20);default:'active';index:idx_user_offers_status" json:"status"`
	Earnings      int       `gorm:"default:0" json:"earnings"`
	TotalClicks   int       `gorm:"default:0" json:"total_clicks"`
	TotalConversions int    `gorm:"default:0" json:"total_conversions"`
	JoinedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	UpdatedAt     time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	User          *AfftokUser  `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Offer         *Offer       `gorm:"foreignKey:OfferID" json:"offer,omitempty"`
	Clicks        []Click      `gorm:"foreignKey:UserOfferID" json:"clicks,omitempty"`
	Conversions   []Conversion `gorm:"foreignKey:UserOfferID" json:"conversions,omitempty"`
}

func (UserOffer) TableName() string {
	return "user_offers"
}

func (uo *UserOffer) ConversionRate() float64 {
	totalClicks := len(uo.Clicks)
	totalConversions := len(uo.Conversions)
	if totalClicks == 0 {
		return 0
	}
	return (float64(totalConversions) / float64(totalClicks)) * 100
}

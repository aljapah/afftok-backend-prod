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
	ExternalOfferID  string     `gorm:"type:varchar(100)" json:"external_offer_id,omitempty"`
	Title            string     `gorm:"type:varchar(255);not null" json:"title"`
	Description      string     `gorm:"type:text" json:"description,omitempty"`
	ImageURL         string     `gorm:"type:text" json:"image_url,omitempty"`
	LogoURL          string     `gorm:"type:text" json:"logo_url,omitempty"`
	DestinationURL   string     `gorm:"type:text;not null" json:"destination_url"`
	Category         string     `gorm:"type:varchar(50)" json:"category,omitempty"`
	Payout           int        `gorm:"default:0" json:"payout"`
	Commission       int        `gorm:"default:0" json:"commission"`
	PayoutType       string     `gorm:"type:varchar(20);default:'cpa'" json:"payout_type"`
	Rating           float64    `gorm:"type:decimal(3,2);default:0.0" json:"rating"`
	UsersCount       int        `gorm:"default:0" json:"users_count"`
	Status           string     `gorm:"type:varchar(20);default:'pending'" json:"status"`
	TotalClicks      int        `gorm:"default:0" json:"total_clicks"`
	TotalConversions int        `gorm:"default:0" json:"total_conversions"`
	CreatedAt        time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	Network          *Network   `gorm:"foreignKey:NetworkID" json:"network,omitempty"`
	UserOffers       []UserOffer `gorm:"foreignKey:OfferID" json:"user_offers,omitempty"`
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
	UserID        uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	OfferID       uuid.UUID `gorm:"type:uuid;not null" json:"offer_id"`
	AffiliateLink string    `gorm:"type:text;not null" json:"affiliate_link"`
	ShortLink     string    `gorm:"type:text" json:"short_link,omitempty"`
	Status        string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	Earnings      int       `gorm:"default:0" json:"earnings"`
	JoinedAt      time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
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

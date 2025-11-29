package models

import (
	"time"

	"github.com/google/uuid"
)

// AfftokUser represents an affiliate marketer
type AfftokUser struct {
	ID               uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Username         string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email            string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash     string    `gorm:"type:varchar(255);not null" json:"-"`
	FullName         string    `gorm:"type:varchar(100)" json:"full_name,omitempty"`
	AvatarURL        string    `gorm:"type:text" json:"avatar_url,omitempty"`
	Bio              string    `gorm:"type:text" json:"bio,omitempty"`
	Role             string    `gorm:"type:varchar(20);default:'user'" json:"role"`
	Status           string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	Points           int       `gorm:"default:0" json:"points"`
	Level            int       `gorm:"default:1" json:"level"`
	TotalClicks      int       `gorm:"default:0" json:"total_clicks"`
	TotalConversions int       `gorm:"default:0" json:"total_conversions"`
	TotalEarnings    int       `gorm:"default:0" json:"total_earnings"`
	CreatedAt        time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt        time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	UserOffers []UserOffer  `gorm:"foreignKey:UserID" json:"user_offers,omitempty"`
	TeamMember *TeamMember  `gorm:"foreignKey:UserID" json:"team_member,omitempty"`
	UserBadges []UserBadge  `gorm:"foreignKey:UserID" json:"user_badges,omitempty"`
}

// TableName specifies the table name
func (AfftokUser) TableName() string {
	return "afftok_users"
}

// UserLevel returns the user level based on conversions
func (u *AfftokUser) UserLevel() string {
	switch {
	case u.TotalConversions >= 500:
		return "legend"
	case u.TotalConversions >= 201:
		return "master"
	case u.TotalConversions >= 51:
		return "expert"
	case u.TotalConversions >= 11:
		return "pro"
	default:
		return "rookie"
	}
}

// UserLevelEmoji returns emoji for user level
func (u *AfftokUser) UserLevelEmoji() string {
	level := u.UserLevel()
	switch level {
	case "legend":
		return "🏆"
	case "master":
		return "👑"
	case "expert":
		return "💎"
	case "pro":
		return "⭐"
	default:
		return "🌱"
	}
}

// ConversionRate calculates the conversion rate
func (u *AfftokUser) ConversionRate() float64 {
	if u.TotalClicks == 0 {
		return 0
	}
	return (float64(u.TotalConversions) / float64(u.TotalClicks)) * 100
}

// PersonalLink returns the user's personal link
func (u *AfftokUser) PersonalLink() string {
	return "afftok.com/u/" + u.Username
}

// AdminUser represents an admin panel user
type AdminUser struct {
	ID           int       `gorm:"primary_key;auto_increment" json:"id"`
	OpenID       string    `gorm:"type:varchar(64);uniqueIndex;not null" json:"open_id"`
	Name         string    `gorm:"type:text" json:"name,omitempty"`
	Email        string    `gorm:"type:varchar(320)" json:"email,omitempty"`
	LoginMethod  string    `gorm:"type:varchar(64)" json:"login_method,omitempty"`
	Role         string    `gorm:"type:varchar(20);default:'user'" json:"role"`
	CreatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt    time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	LastSignedIn time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"last_signed_in"`
}

// TableName specifies the table name
func (AdminUser) TableName() string {
	return "admin_users"
}

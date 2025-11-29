package models

import (
    "time"

    "github.com/google/uuid"
)

type Click struct {
    ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    UserOfferID uuid.UUID  `gorm:"type:uuid;not null" json:"user_offer_id"`
    IPAddress   string     `gorm:"type:varchar(45)" json:"ip_address,omitempty"`
    UserAgent   string     `gorm:"type:text" json:"user_agent,omitempty"`
    Device      string     `gorm:"type:varchar(50)" json:"device,omitempty"`
    Browser     string     `gorm:"type:varchar(50)" json:"browser,omitempty"`
    OS          string     `gorm:"type:varchar(50)" json:"os,omitempty"`
    Country     string     `gorm:"type:varchar(2)" json:"country,omitempty"`
    City        string     `gorm:"type:varchar(100)" json:"city,omitempty"`
    Referrer    string     `gorm:"type:text" json:"referrer,omitempty"`
    ClickedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP;index" json:"clicked_at"`
    UserOffer   *UserOffer `gorm:"foreignKey:UserOfferID" json:"user_offer,omitempty"`
}

func (Click) TableName() string {
    return "clicks"
}

type Conversion struct {
    ID                   uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    UserOfferID          uuid.UUID  `gorm:"type:uuid;not null" json:"user_offer_id"`
    ClickID              *uuid.UUID `gorm:"type:uuid" json:"click_id,omitempty"`
    ExternalConversionID string     `gorm:"type:varchar(100)" json:"external_conversion_id,omitempty"`
    Amount               int        `gorm:"default:0" json:"amount"`
    Commission           int        `gorm:"default:0" json:"commission"`
    Status               string     `gorm:"type:varchar(20);default:'pending';index" json:"status"`
    ConvertedAt          time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"converted_at"`
    ApprovedAt           *time.Time `json:"approved_at,omitempty"`
    UserOffer            *UserOffer `gorm:"foreignKey:UserOfferID" json:"user_offer,omitempty"`
    Click                *Click     `gorm:"foreignKey:ClickID" json:"click,omitempty"`
}

func (Conversion) TableName() string {
    return "conversions"
}

type Team struct {
    ID          uuid.UUID    `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    Name        string       `gorm:"type:varchar(100);not null" json:"name"`
    Description string       `gorm:"type:text" json:"description,omitempty"`
    LogoURL     string       `gorm:"type:text;column:logo_url" json:"logo_url,omitempty"`
    OwnerID     uuid.UUID    `gorm:"type:uuid;not null;column:owner_id" json:"owner_id"`
    MaxMembers  int          `gorm:"default:10;column:max_members" json:"max_members"`
    TotalPoints int          `gorm:"default:0;column:total_points" json:"total_points"`
    MemberCount int          `gorm:"default:1;column:member_count" json:"member_count"`
    Status      string       `gorm:"type:varchar(20);default:'active'" json:"status"`
    CreatedAt   time.Time    `gorm:"default:CURRENT_TIMESTAMP;column:created_at" json:"created_at"`
    UpdatedAt   time.Time    `gorm:"default:CURRENT_TIMESTAMP;column:updated_at" json:"updated_at"`
    Owner       *AfftokUser  `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
    Members     []TeamMember `gorm:"foreignKey:TeamID" json:"members,omitempty"`
}

func (Team) TableName() string {
    return "teams"
}

type TeamMember struct {
    ID       uuid.UUID   `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    TeamID   uuid.UUID   `gorm:"type:uuid;not null;column:team_id" json:"team_id"`
    UserID   uuid.UUID   `gorm:"type:uuid;not null;column:user_id" json:"user_id"`
    Role     string      `gorm:"type:varchar(20);default:'member'" json:"role"`
    Points   int         `gorm:"default:0" json:"points"`
    JoinedAt time.Time   `gorm:"default:CURRENT_TIMESTAMP;column:joined_at" json:"joined_at"`
    Team     *Team       `gorm:"foreignKey:TeamID" json:"team,omitempty"`
    User     *AfftokUser `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (TeamMember) TableName() string {
    return "team_members"
}

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

type UserBadge struct {
    ID       uuid.UUID   `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
    UserID   uuid.UUID   `gorm:"type:uuid;not null;column:user_id" json:"user_id"`
    BadgeID  uuid.UUID   `gorm:"type:uuid;not null;column:badge_id" json:"badge_id"`
    EarnedAt time.Time   `gorm:"default:CURRENT_TIMESTAMP;column:earned_at" json:"earned_at"`
    User     *AfftokUser `gorm:"foreignKey:UserID" json:"user,omitempty"`
    Badge    *Badge      `gorm:"foreignKey:BadgeID" json:"badge,omitempty"`
}

func (UserBadge) TableName() string {
    return "user_badges"
}

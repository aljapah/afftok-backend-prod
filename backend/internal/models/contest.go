package models

import (
	"time"

	"github.com/google/uuid"
)

// Contest represents a competition/challenge for teams or individuals
type Contest struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Title       string     `gorm:"type:varchar(255);not null" json:"title"`
	TitleAr     string     `gorm:"type:varchar(255)" json:"title_ar,omitempty"`
	Description string     `gorm:"type:text" json:"description,omitempty"`
	DescriptionAr string   `gorm:"type:text" json:"description_ar,omitempty"`
	ImageURL    string     `gorm:"type:text" json:"image_url,omitempty"`
	
	// Prize Information
	PrizeTitle       string  `gorm:"type:varchar(255)" json:"prize_title,omitempty"`
	PrizeTitleAr     string  `gorm:"type:varchar(255)" json:"prize_title_ar,omitempty"`
	PrizeDescription string  `gorm:"type:text" json:"prize_description,omitempty"`
	PrizeDescriptionAr string `gorm:"type:text" json:"prize_description_ar,omitempty"`
	PrizeAmount      float64 `gorm:"type:decimal(10,2);default:0" json:"prize_amount"`
	PrizeCurrency    string  `gorm:"type:varchar(10);default:'USD'" json:"prize_currency"`
	
	// Contest Type & Target
	ContestType   string     `gorm:"type:varchar(20);default:'team'" json:"contest_type"` // team, individual
	TargetType    string     `gorm:"type:varchar(20);default:'clicks'" json:"target_type"` // clicks, conversions, referrals, points
	TargetValue   int        `gorm:"default:100" json:"target_value"`
	
	// Conditions
	MinClicks      int `gorm:"default:0" json:"min_clicks"`
	MinConversions int `gorm:"default:0" json:"min_conversions"`
	MinMembers     int `gorm:"default:1" json:"min_members"` // For team contests
	MaxParticipants int `gorm:"default:0" json:"max_participants"` // 0 = unlimited
	
	// Dates
	StartDate   time.Time  `gorm:"not null" json:"start_date"`
	EndDate     time.Time  `gorm:"not null" json:"end_date"`
	
	// Status
	Status      string     `gorm:"type:varchar(20);default:'draft'" json:"status"` // draft, active, ended, cancelled
	
	// Stats
	ParticipantsCount int `gorm:"default:0" json:"participants_count"`
	
	CreatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	
	// Relationships
	Participants []ContestParticipant `gorm:"foreignKey:ContestID" json:"participants,omitempty"`
}

func (Contest) TableName() string {
	return "contests"
}

// IsActive checks if contest is currently active
func (c *Contest) IsActive() bool {
	now := time.Now()
	return c.Status == "active" && now.After(c.StartDate) && now.Before(c.EndDate)
}

// IsEnded checks if contest has ended
func (c *Contest) IsEnded() bool {
	return time.Now().After(c.EndDate) || c.Status == "ended"
}

// TimeLeft returns duration until end
func (c *Contest) TimeLeft() time.Duration {
	return c.EndDate.Sub(time.Now())
}

// ContestParticipant represents a team or user participating in a contest
type ContestParticipant struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	ContestID   uuid.UUID  `gorm:"type:uuid;not null" json:"contest_id"`
	TeamID      *uuid.UUID `gorm:"type:uuid" json:"team_id,omitempty"` // For team contests
	UserID      *uuid.UUID `gorm:"type:uuid" json:"user_id,omitempty"` // For individual contests
	
	// Progress
	CurrentClicks      int `gorm:"default:0" json:"current_clicks"`
	CurrentConversions int `gorm:"default:0" json:"current_conversions"`
	CurrentPoints      int `gorm:"default:0" json:"current_points"`
	Progress           int `gorm:"default:0" json:"progress"` // Percentage
	
	// Ranking
	Rank        int        `gorm:"default:0" json:"rank"`
	
	// Status
	Status      string     `gorm:"type:varchar(20);default:'active'" json:"status"` // active, winner, completed, disqualified
	
	JoinedAt    time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	UpdatedAt   time.Time  `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`
	
	// Relationships
	Contest Contest    `gorm:"foreignKey:ContestID" json:"contest,omitempty"`
	Team    *Team      `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	User    *AfftokUser `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (ContestParticipant) TableName() string {
	return "contest_participants"
}

// Contest Status Constants
const (
	ContestStatusDraft     = "draft"
	ContestStatusActive    = "active"
	ContestStatusEnded     = "ended"
	ContestStatusCancelled = "cancelled"
)

// Contest Type Constants
const (
	ContestTypeTeam       = "team"
	ContestTypeIndividual = "individual"
)

// Target Type Constants
const (
	TargetTypeClicks      = "clicks"
	TargetTypeConversions = "conversions"
	TargetTypeReferrals   = "referrals"
	TargetTypePoints      = "points"
)


package models

import (
	"time"

	"github.com/google/uuid"
)

// Team represents a group of affiliate marketers
type Team struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	Name        string    `gorm:"type:varchar(100);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	LogoURL     string    `gorm:"type:text" json:"logo_url,omitempty"`
	OwnerID     uuid.UUID `gorm:"type:uuid;not null" json:"owner_id"`
	MaxMembers  int       `gorm:"default:10" json:"max_members"`
	MemberCount int       `gorm:"default:1" json:"member_count"`
	TotalPoints int       `gorm:"default:0" json:"total_points"`
	TotalClicks int       `gorm:"default:0" json:"total_clicks"`
	TotalConversions int  `gorm:"default:0" json:"total_conversions"`
	Status      string    `gorm:"type:varchar(20);default:'active'" json:"status"`
	
	// Invite system
	InviteCode  string    `gorm:"type:varchar(20);uniqueIndex" json:"invite_code,omitempty"`
	InviteURL   string    `gorm:"type:text" json:"invite_url,omitempty"`
	
	CreatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt   time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"updated_at"`

	// Relationships
	Owner   AfftokUser   `gorm:"foreignKey:OwnerID" json:"owner,omitempty"`
	Members []TeamMember `gorm:"foreignKey:TeamID" json:"members,omitempty"`
}

// TableName specifies the table name
func (Team) TableName() string {
	return "teams"
}

// TeamMember represents a user's membership in a team
type TeamMember struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	TeamID    uuid.UUID `gorm:"type:uuid;not null" json:"team_id"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"user_id"`
	Role      string    `gorm:"type:varchar(20);default:'member'" json:"role"` // owner, admin, member
	Status    string    `gorm:"type:varchar(20);default:'active'" json:"status"` // active, pending, rejected
	Points    int       `gorm:"default:0" json:"points"`
	JoinedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"joined_at"`
	
	// Relationships
	Team Team       `gorm:"foreignKey:TeamID" json:"team,omitempty"`
	User AfftokUser `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// TableName specifies the table name
func (TeamMember) TableName() string {
	return "team_members"
}

// TeamMember Status Constants
const (
	TeamMemberStatusActive   = "active"
	TeamMemberStatusPending  = "pending"
	TeamMemberStatusRejected = "rejected"
)

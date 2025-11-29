package models

import (
	"time"

	"github.com/google/uuid"
)

type PromoterRating struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;default:uuid_generate_v4()" json:"id"`
	PromoterID uuid.UUID `gorm:"type:uuid;not null" json:"promoter_id"`
	VisitorIP  string    `gorm:"type:varchar(45)" json:"visitor_ip"`
	Rating     int       `gorm:"type:integer;check:rating >= 1 AND rating <= 5" json:"rating"`
	CreatedAt  time.Time `gorm:"default:CURRENT_TIMESTAMP" json:"created_at"`
}

func (PromoterRating) TableName() string {
	return "promoter_ratings"
}

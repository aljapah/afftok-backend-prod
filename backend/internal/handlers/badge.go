package handlers

import (
	"net/http"

	"github.com/afftok/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BadgeHandler struct {
	db *gorm.DB
}

func NewBadgeHandler(db *gorm.DB) *BadgeHandler {
	return &BadgeHandler{db: db}
}

func (h *BadgeHandler) GetAllBadges(c *gin.Context) {
	var badges []models.Badge

	if err := h.db.Order("required_value ASC").Find(&badges).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch badges"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"badges": badges,
	})
}

func (h *BadgeHandler) GetMyBadges(c *gin.Context) {
	userID, _ := c.Get("userID")

	var userBadges []models.UserBadge
	if err := h.db.Preload("Badge").Where("user_id = ?", userID).Order("earned_at DESC").Find(&userBadges).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch your badges"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"badges": userBadges,
	})
}

func (h *BadgeHandler) CheckAndAwardBadges(userID uuid.UUID) error {
	var user models.AfftokUser
	if err := h.db.First(&user, "id = ?", userID).Error; err != nil {
		return err
	}

	var badges []models.Badge
	h.db.Find(&badges)

	for _, badge := range badges {
		var existingUserBadge models.UserBadge
		if err := h.db.Where("user_id = ? AND badge_id = ?", userID, badge.ID).First(&existingUserBadge).Error; err == nil {
			continue
		}

		earned := false
		switch badge.Criteria {
		case "conversions":
			if user.TotalConversions >= badge.RequiredValue {
				earned = true
			}
		case "clicks":
			if user.TotalClicks >= badge.RequiredValue {
				earned = true
			}
		case "earnings":
			if user.TotalEarnings >= badge.RequiredValue {
				earned = true
			}
		case "points":
			if user.Points >= badge.RequiredValue {
				earned = true
			}
		}

		if earned {
			userBadge := models.UserBadge{
				ID:      uuid.New(),
				UserID:  userID,
				BadgeID: badge.ID,
			}
			h.db.Create(&userBadge)

			h.db.Model(&user).UpdateColumn("points", h.db.Raw("points + ?", badge.Points))
		}
	}

	return nil
}

func (h *BadgeHandler) CreateBadge(c *gin.Context) {
	type CreateBadgeRequest struct {
		Name          string `json:"name" binding:"required"`
		Description   string `json:"description"`
		IconURL       string `json:"icon_url"`
		Criteria      string `json:"criteria" binding:"required"`
		RequiredValue int    `json:"required_value" binding:"required"`
		Points        int    `json:"points"`
	}

	var req CreateBadgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	badge := models.Badge{
		ID:            uuid.New(),
		Name:          req.Name,
		Description:   req.Description,
		IconURL:       req.IconURL,
		Criteria:      req.Criteria,
		RequiredValue: req.RequiredValue,
		Points:        req.Points,
	}

	if err := h.db.Create(&badge).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create badge"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Badge created successfully",
		"badge":   badge,
	})
}

func (h *BadgeHandler) UpdateBadge(c *gin.Context) {
	badgeID := c.Param("id")

	type UpdateBadgeRequest struct {
		Name          string `json:"name"`
		Description   string `json:"description"`
		IconURL       string `json:"icon_url"`
		Criteria      string `json:"criteria"`
		RequiredValue int    `json:"required_value"`
		Points        int    `json:"points"`
	}

	var req UpdateBadgeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.IconURL != "" {
		updates["icon_url"] = req.IconURL
	}
	if req.Criteria != "" {
		updates["criteria"] = req.Criteria
	}
	if req.RequiredValue > 0 {
		updates["required_value"] = req.RequiredValue
	}
	if req.Points > 0 {
		updates["points"] = req.Points
	}

	if err := h.db.Model(&models.Badge{}).Where("id = ?", badgeID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update badge"})
		return
	}

	var badge models.Badge
	h.db.First(&badge, "id = ?", badgeID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Badge updated successfully",
		"badge":   badge,
	})
}

func (h *BadgeHandler) DeleteBadge(c *gin.Context) {
	badgeID := c.Param("id")

	id, err := uuid.Parse(badgeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid badge ID"})
		return
	}

	if err := h.db.Delete(&models.Badge{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete badge"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Badge deleted successfully",
	})
}

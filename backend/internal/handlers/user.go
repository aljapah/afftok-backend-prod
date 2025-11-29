package handlers

import (
	"net/http"

	"github.com/afftok/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	var users []models.AfftokUser

	page := 1
	limit := 20
	offset := (page - 1) * limit

	query := h.db.Select("id, username, email, full_name, avatar_url, role, status, points, level, total_clicks, total_conversions, total_earnings, created_at")

	sortBy := c.DefaultQuery("sort", "created_at")
	order := c.DefaultQuery("order", "desc")
	query = query.Order(sortBy + " " + order)

	if err := query.Limit(limit).Offset(offset).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	var total int64
	h.db.Model(&models.AfftokUser{}).Count(&total)

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (h *UserHandler) GetUser(c *gin.Context) {
	userID := c.Param("id")

	var user models.AfftokUser
	if err := h.db.
		Preload("UserBadges.Badge").
		Select("id, username, email, full_name, avatar_url, bio, role, status, points, level, total_clicks, total_conversions, total_earnings, created_at").
		First(&user, "id = ?", userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	type UpdateProfileRequest struct {
		FullName  string `json:"full_name"`
		Bio       string `json:"bio"`
		AvatarURL string `json:"avatar_url"`
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.FullName != "" {
		updates["full_name"] = req.FullName
	}
	if req.Bio != "" {
		updates["bio"] = req.Bio
	}
	if req.AvatarURL != "" {
		updates["avatar_url"] = req.AvatarURL
	}

	if err := h.db.Model(&models.AfftokUser{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	var user models.AfftokUser
	h.db.First(&user, "id = ?", userID)
	user.PasswordHash = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"user":    user,
	})
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	type UpdateUserRequest struct {
		Status string `json:"status"`
		Role   string `json:"role"`
		Points int    `json:"points"`
		Level  int    `json:"level"`
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updates := map[string]interface{}{}
	if req.Status != "" {
		updates["status"] = req.Status
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}
	if req.Points > 0 {
		updates["points"] = req.Points
	}
	if req.Level > 0 {
		updates["level"] = req.Level
	}

	if err := h.db.Model(&models.AfftokUser{}).Where("id = ?", userID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	var user models.AfftokUser
	h.db.First(&user, "id = ?", userID)
	user.PasswordHash = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"user":    user,
	})
}

func (h *UserHandler) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := h.db.Delete(&models.AfftokUser{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
	})
}

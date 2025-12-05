package handlers

import (
	"net/http"
	"strconv"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserHandler struct {
	db               *gorm.DB
	analyticsService *services.AnalyticsService
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{
		db:               db,
		analyticsService: services.NewAnalyticsService(),
	}
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

// GetMyStats returns detailed statistics for the current user
func (h *UserHandler) GetMyStats(c *gin.Context) {
	userID, _ := c.Get("userID")

	stats, err := h.analyticsService.GetUserStats(userID.(uuid.UUID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"stats":   stats,
	})
}

// GetDailyStats returns daily click/conversion data for charts
func (h *UserHandler) GetDailyStats(c *gin.Context) {
	userID, _ := c.Get("userID")

	days := 7 // Default to 7 days
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 && parsed <= 90 {
			days = parsed
		}
	}

	stats, err := h.analyticsService.GetDailyStats(userID.(uuid.UUID), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch daily stats"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
		"days":    days,
	})
}

// GetLeaderboard returns top 10 promoters and current user's rank
func (h *UserHandler) GetLeaderboard(c *gin.Context) {
	userID, _ := c.Get("userID")
	currentUserID := userID.(uuid.UUID)

	limit := 10
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	// Get top users by points (clicks * 2 + conversions * 20)
	var leaderboard []struct {
		ID               uuid.UUID `json:"id"`
		Username         string    `json:"username"`
		FullName         string    `json:"full_name"`
		AvatarURL        string    `json:"avatar_url"`
		Country          string    `json:"country"`
		TotalClicks      int       `json:"total_clicks"`
		TotalConversions int       `json:"total_conversions"`
		Points           int       `json:"points"`
		Rank             int       `json:"rank"`
	}

	// Query top users ordered by calculated points
	err := h.db.Model(&models.AfftokUser{}).
		Select(`
			id, 
			username, 
			full_name, 
			avatar_url, 
			COALESCE(country, '') as country,
			total_clicks, 
			total_conversions,
			(total_clicks * 2 + total_conversions * 20) as points
		`).
		Where("role = ? OR role IS NULL OR role = ''", "user"). // Only promoters
		Where("status = ?", "active").
		Order("points DESC").
		Limit(limit).
		Scan(&leaderboard).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch leaderboard",
		})
		return
	}

	// Add rank numbers
	for i := range leaderboard {
		leaderboard[i].Rank = i + 1
	}

	// Get current user's rank
	var myRank struct {
		Rank             int    `json:"rank"`
		TotalClicks      int    `json:"total_clicks"`
		TotalConversions int    `json:"total_conversions"`
		Points           int    `json:"points"`
		Country          string `json:"country"`
	}

	// Get current user's stats
	var currentUser models.AfftokUser
	if err := h.db.First(&currentUser, "id = ?", currentUserID).Error; err == nil {
		myRank.TotalClicks = currentUser.TotalClicks
		myRank.TotalConversions = currentUser.TotalConversions
		myRank.Points = currentUser.TotalClicks*2 + currentUser.TotalConversions*20
		myRank.Country = currentUser.Country

		// Calculate rank
		var usersAbove int64
		h.db.Model(&models.AfftokUser{}).
			Where("(total_clicks * 2 + total_conversions * 20) > ?", myRank.Points).
			Where("role = ? OR role IS NULL OR role = ''", "user").
			Where("status = ?", "active").
			Count(&usersAbove)
		myRank.Rank = int(usersAbove) + 1
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"leaderboard": leaderboard,
		"my_rank":     myRank,
	})
}

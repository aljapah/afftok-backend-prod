package handlers

import (
	"net/http"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ContestHandler struct {
	db *gorm.DB
}

func NewContestHandler(db *gorm.DB) *ContestHandler {
	return &ContestHandler{db: db}
}

// ========== Public Endpoints ==========

// GetActiveContests returns all active contests
func (h *ContestHandler) GetActiveContests(c *gin.Context) {
	var contests []models.Contest
	
	now := time.Now()
	if err := h.db.Where("status = ? AND start_date <= ? AND end_date >= ?", 
		models.ContestStatusActive, now, now).
		Order("end_date ASC").
		Find(&contests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contests": contests,
		"count":    len(contests),
	})
}

// GetContest returns a single contest with participants
func (h *ContestHandler) GetContest(c *gin.Context) {
	contestID := c.Param("id")

	var contest models.Contest
	if err := h.db.Preload("Participants.Team").Preload("Participants.User").
		First(&contest, "id = ?", contestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contest not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"contest": contest,
	})
}

// GetContestLeaderboard returns top participants
func (h *ContestHandler) GetContestLeaderboard(c *gin.Context) {
	contestID := c.Param("id")

	var participants []models.ContestParticipant
	if err := h.db.Preload("Team").Preload("User").
		Where("contest_id = ?", contestID).
		Order("progress DESC, current_clicks DESC").
		Limit(50).
		Find(&participants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch leaderboard"})
		return
	}

	// Update ranks
	for i := range participants {
		participants[i].Rank = i + 1
	}

	c.JSON(http.StatusOK, gin.H{
		"leaderboard": participants,
	})
}

// JoinContest allows a team or user to join a contest
func (h *ContestHandler) JoinContest(c *gin.Context) {
	contestID := c.Param("id")
	userID, _ := c.Get("userID")

	var contest models.Contest
	if err := h.db.First(&contest, "id = ?", contestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contest not found"})
		return
	}

	if !contest.IsActive() {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Contest is not active"})
		return
	}

	// Check max participants
	if contest.MaxParticipants > 0 && contest.ParticipantsCount >= contest.MaxParticipants {
		c.JSON(http.StatusConflict, gin.H{"error": "Contest is full"})
		return
	}

	// Check if already participating
	var existingParticipant models.ContestParticipant
	if contest.ContestType == models.ContestTypeIndividual {
		if err := h.db.Where("contest_id = ? AND user_id = ?", contestID, userID).
			First(&existingParticipant).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "You are already participating in this contest"})
			return
		}
	} else {
		// Team contest - check if user's team is already participating
		var member models.TeamMember
		if err := h.db.Where("user_id = ? AND status = ?", userID, "active").
			First(&member).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "You must be in a team to join this contest"})
			return
		}

		if err := h.db.Where("contest_id = ? AND team_id = ?", contestID, member.TeamID).
			First(&existingParticipant).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Your team is already participating in this contest"})
			return
		}
	}

	// Create participant
	participant := models.ContestParticipant{
		ID:        uuid.New(),
		ContestID: uuid.MustParse(contestID),
		Status:    "active",
	}

	if contest.ContestType == models.ContestTypeIndividual {
		uid := userID.(uuid.UUID)
		participant.UserID = &uid
	} else {
		var member models.TeamMember
		h.db.Where("user_id = ? AND status = ?", userID, "active").First(&member)
		participant.TeamID = &member.TeamID
	}

	if err := h.db.Create(&participant).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join contest"})
		return
	}

	// Increment participants count
	h.db.Model(&contest).UpdateColumn("participants_count", gorm.Expr("participants_count + 1"))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Successfully joined the contest",
	})
}

// GetMyContests returns contests the user is participating in
func (h *ContestHandler) GetMyContests(c *gin.Context) {
	userID, _ := c.Get("userID")

	var participants []models.ContestParticipant
	if err := h.db.Preload("Contest").
		Where("user_id = ?", userID).
		Find(&participants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contests"})
		return
	}

	// Also check team contests
	var member models.TeamMember
	if err := h.db.Where("user_id = ? AND status = ?", userID, "active").First(&member).Error; err == nil {
		var teamParticipants []models.ContestParticipant
		h.db.Preload("Contest").Where("team_id = ?", member.TeamID).Find(&teamParticipants)
		participants = append(participants, teamParticipants...)
	}

	c.JSON(http.StatusOK, gin.H{
		"contests": participants,
	})
}

// ========== Admin Endpoints ==========

// AdminGetAllContests returns all contests for admin
func (h *ContestHandler) AdminGetAllContests(c *gin.Context) {
	var contests []models.Contest
	
	if err := h.db.Order("created_at DESC").Find(&contests).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch contests"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"contests": contests,
	})
}

// AdminCreateContest creates a new contest
func (h *ContestHandler) AdminCreateContest(c *gin.Context) {
	type CreateContestRequest struct {
		Title            string  `json:"title" binding:"required"`
		TitleAr          string  `json:"title_ar"`
		Description      string  `json:"description"`
		DescriptionAr    string  `json:"description_ar"`
		ImageURL         string  `json:"image_url"`
		PrizeTitle       string  `json:"prize_title"`
		PrizeTitleAr     string  `json:"prize_title_ar"`
		PrizeDescription string  `json:"prize_description"`
		PrizeAmount      float64 `json:"prize_amount"`
		PrizeCurrency    string  `json:"prize_currency"`
		ContestType      string  `json:"contest_type"`
		TargetType       string  `json:"target_type"`
		TargetValue      int     `json:"target_value"`
		MinClicks        int     `json:"min_clicks"`
		MinConversions   int     `json:"min_conversions"`
		MinMembers       int     `json:"min_members"`
		MaxParticipants  int     `json:"max_participants"`
		StartDate        string  `json:"start_date" binding:"required"`
		EndDate          string  `json:"end_date" binding:"required"`
		Status           string  `json:"status"`
	}

	var req CreateContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	startDate, err := time.Parse(time.RFC3339, req.StartDate)
	if err != nil {
		startDate, _ = time.Parse("2006-01-02", req.StartDate)
	}

	endDate, err := time.Parse(time.RFC3339, req.EndDate)
	if err != nil {
		endDate, _ = time.Parse("2006-01-02", req.EndDate)
	}

	// Defaults
	if req.ContestType == "" {
		req.ContestType = models.ContestTypeIndividual
	}
	if req.TargetType == "" {
		req.TargetType = models.TargetTypeClicks
	}
	if req.Status == "" {
		req.Status = models.ContestStatusDraft
	}
	if req.TargetValue == 0 {
		req.TargetValue = 100
	}
	if req.PrizeCurrency == "" {
		req.PrizeCurrency = "USD"
	}
	if req.MinMembers == 0 {
		req.MinMembers = 1
	}

	contest := models.Contest{
		ID:               uuid.New(),
		Title:            req.Title,
		TitleAr:          req.TitleAr,
		Description:      req.Description,
		DescriptionAr:    req.DescriptionAr,
		ImageURL:         req.ImageURL,
		PrizeTitle:       req.PrizeTitle,
		PrizeTitleAr:     req.PrizeTitleAr,
		PrizeDescription: req.PrizeDescription,
		PrizeAmount:      req.PrizeAmount,
		PrizeCurrency:    req.PrizeCurrency,
		ContestType:      req.ContestType,
		TargetType:       req.TargetType,
		TargetValue:      req.TargetValue,
		MinClicks:        req.MinClicks,
		MinConversions:   req.MinConversions,
		MinMembers:       req.MinMembers,
		MaxParticipants:  req.MaxParticipants,
		StartDate:        startDate,
		EndDate:          endDate,
		Status:           req.Status,
	}

	if err := h.db.Create(&contest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create contest"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Contest created successfully",
		"contest": contest,
	})
}

// AdminUpdateContest updates a contest
func (h *ContestHandler) AdminUpdateContest(c *gin.Context) {
	contestID := c.Param("id")

	var contest models.Contest
	if err := h.db.First(&contest, "id = ?", contestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contest not found"})
		return
	}

	type UpdateContestRequest struct {
		Title            *string  `json:"title"`
		TitleAr          *string  `json:"title_ar"`
		Description      *string  `json:"description"`
		DescriptionAr    *string  `json:"description_ar"`
		ImageURL         *string  `json:"image_url"`
		PrizeTitle       *string  `json:"prize_title"`
		PrizeTitleAr     *string  `json:"prize_title_ar"`
		PrizeDescription *string  `json:"prize_description"`
		PrizeAmount      *float64 `json:"prize_amount"`
		PrizeCurrency    *string  `json:"prize_currency"`
		ContestType      *string  `json:"contest_type"`
		TargetType       *string  `json:"target_type"`
		TargetValue      *int     `json:"target_value"`
		MinClicks        *int     `json:"min_clicks"`
		MinConversions   *int     `json:"min_conversions"`
		MinMembers       *int     `json:"min_members"`
		MaxParticipants  *int     `json:"max_participants"`
		StartDate        *string  `json:"start_date"`
		EndDate          *string  `json:"end_date"`
		Status           *string  `json:"status"`
	}

	var req UpdateContestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update fields
	if req.Title != nil {
		contest.Title = *req.Title
	}
	if req.TitleAr != nil {
		contest.TitleAr = *req.TitleAr
	}
	if req.Description != nil {
		contest.Description = *req.Description
	}
	if req.DescriptionAr != nil {
		contest.DescriptionAr = *req.DescriptionAr
	}
	if req.ImageURL != nil {
		contest.ImageURL = *req.ImageURL
	}
	if req.PrizeTitle != nil {
		contest.PrizeTitle = *req.PrizeTitle
	}
	if req.PrizeTitleAr != nil {
		contest.PrizeTitleAr = *req.PrizeTitleAr
	}
	if req.PrizeDescription != nil {
		contest.PrizeDescription = *req.PrizeDescription
	}
	if req.PrizeAmount != nil {
		contest.PrizeAmount = *req.PrizeAmount
	}
	if req.PrizeCurrency != nil {
		contest.PrizeCurrency = *req.PrizeCurrency
	}
	if req.ContestType != nil {
		contest.ContestType = *req.ContestType
	}
	if req.TargetType != nil {
		contest.TargetType = *req.TargetType
	}
	if req.TargetValue != nil {
		contest.TargetValue = *req.TargetValue
	}
	if req.MinClicks != nil {
		contest.MinClicks = *req.MinClicks
	}
	if req.MinConversions != nil {
		contest.MinConversions = *req.MinConversions
	}
	if req.MinMembers != nil {
		contest.MinMembers = *req.MinMembers
	}
	if req.MaxParticipants != nil {
		contest.MaxParticipants = *req.MaxParticipants
	}
	if req.StartDate != nil {
		startDate, err := time.Parse(time.RFC3339, *req.StartDate)
		if err != nil {
			startDate, _ = time.Parse("2006-01-02", *req.StartDate)
		}
		contest.StartDate = startDate
	}
	if req.EndDate != nil {
		endDate, err := time.Parse(time.RFC3339, *req.EndDate)
		if err != nil {
			endDate, _ = time.Parse("2006-01-02", *req.EndDate)
		}
		contest.EndDate = endDate
	}
	if req.Status != nil {
		contest.Status = *req.Status
	}

	contest.UpdatedAt = time.Now()

	if err := h.db.Save(&contest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update contest"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Contest updated successfully",
		"contest": contest,
	})
}

// AdminDeleteContest deletes a contest
func (h *ContestHandler) AdminDeleteContest(c *gin.Context) {
	contestID := c.Param("id")

	var contest models.Contest
	if err := h.db.First(&contest, "id = ?", contestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Contest not found"})
		return
	}

	// Delete participants first
	h.db.Where("contest_id = ?", contestID).Delete(&models.ContestParticipant{})

	// Delete contest
	if err := h.db.Delete(&contest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete contest"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Contest deleted successfully",
	})
}

// AdminGetContestParticipants returns all participants for a contest
func (h *ContestHandler) AdminGetContestParticipants(c *gin.Context) {
	contestID := c.Param("id")

	var participants []models.ContestParticipant
	if err := h.db.Preload("Team").Preload("User").
		Where("contest_id = ?", contestID).
		Order("progress DESC").
		Find(&participants).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch participants"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"participants": participants,
	})
}


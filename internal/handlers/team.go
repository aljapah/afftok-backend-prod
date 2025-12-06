package handlers

import (
	"net/http"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeamHandler struct {
	db *gorm.DB
}

func NewTeamHandler(db *gorm.DB) *TeamHandler {
	return &TeamHandler{db: db}
}

func (h *TeamHandler) GetAllTeams(c *gin.Context) {
	var teams []models.Team

	query := h.db.Preload("Owner").Preload("Members.User")

	status := c.Query("status")
	if status != "" {
		query = query.Where("status = ?", status)
	}

	sortBy := c.DefaultQuery("sort", "total_points")
	order := c.DefaultQuery("order", "desc")
	query = query.Order(sortBy + " " + order)

	if err := query.Find(&teams).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teams"})
		return
	}

	// Calculate team totals from members' actual stats
	for i := range teams {
		var totalClicks, totalConversions, totalPoints int
		for _, member := range teams[i].Members {
			if member.Status == "active" {
				totalClicks += member.User.TotalClicks
				totalConversions += member.User.TotalConversions
				totalPoints += member.User.Points
			}
		}
		teams[i].TotalClicks = totalClicks
		teams[i].TotalConversions = totalConversions
		teams[i].TotalPoints = totalPoints
	}

	c.JSON(http.StatusOK, gin.H{
		"teams": teams,
	})
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamID := c.Param("id")

	var team models.Team
	if err := h.db.Preload("Owner").Preload("Members.User").First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Calculate team totals from members' actual stats
	var totalClicks, totalConversions, totalPoints int
	for _, member := range team.Members {
		if member.Status == "active" {
			totalClicks += member.User.TotalClicks
			totalConversions += member.User.TotalConversions
			totalPoints += member.User.Points
		}
	}
	team.TotalClicks = totalClicks
	team.TotalConversions = totalConversions
	team.TotalPoints = totalPoints

	c.JSON(http.StatusOK, gin.H{
		"team": team,
	})
}

func (h *TeamHandler) CreateTeam(c *gin.Context) {
	userID, _ := c.Get("userID")

	type CreateTeamRequest struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		LogoURL     string `json:"logo_url"`
		MaxMembers  int    `json:"max_members"`
	}

	var req CreateTeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.MaxMembers == 0 {
		req.MaxMembers = 10
	}

	var existingMember models.TeamMember
	if err := h.db.Where("user_id = ?", userID).First(&existingMember).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You are already in a team"})
		return
	}

	inviteCode := generateInviteCode()
	team := models.Team{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		LogoURL:     req.LogoURL,
		OwnerID:     userID.(uuid.UUID),
		MaxMembers:  req.MaxMembers,
		TotalPoints: 0,
		MemberCount: 1,
		Status:      "active",
		InviteCode:  inviteCode,
		InviteURL:   "https://afftok.com/join/" + inviteCode,
	}

	if err := h.db.Create(&team).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team"})
		return
	}

	member := models.TeamMember{
		ID:     uuid.New(),
		TeamID: team.ID,
		UserID: userID.(uuid.UUID),
		Role:   "owner",
		Status: "active",
		Points: 0,
	}

	h.db.Create(&member)

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Team created successfully",
		"team":        team,
		"invite_code": team.InviteCode,
		"invite_url":  team.InviteURL,
	})
}

func (h *TeamHandler) JoinTeam(c *gin.Context) {
	teamID := c.Param("id")
	userID, _ := c.Get("userID")

	var team models.Team
	if err := h.db.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if team.MemberCount >= team.MaxMembers {
		c.JSON(http.StatusConflict, gin.H{"error": "Team is full"})
		return
	}

	var existingMember models.TeamMember
	if err := h.db.Where("user_id = ?", userID).First(&existingMember).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You are already in a team"})
		return
	}

	member := models.TeamMember{
		ID:     uuid.New(),
		TeamID: team.ID,
		UserID: userID.(uuid.UUID),
		Role:   "member",
		Points: 0,
	}

	if err := h.db.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join team"})
		return
	}

	h.db.Model(&team).UpdateColumn("member_count", h.db.Raw("member_count + 1"))

	c.JSON(http.StatusOK, gin.H{
		"message": "Joined team successfully",
		"team":    team,
	})
}

func (h *TeamHandler) LeaveTeam(c *gin.Context) {
	teamID := c.Param("id")
	userID, _ := c.Get("userID")

	var member models.TeamMember
	if err := h.db.Where("team_id = ? AND user_id = ?", teamID, userID).First(&member).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "You are not in this team"})
		return
	}

	if member.Role == "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Team owner cannot leave. Please delete the team or transfer ownership."})
		return
	}

	if err := h.db.Delete(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave team"})
		return
	}

	var team models.Team
	if err := h.db.First(&team, "id = ?", teamID).Error; err == nil {
		h.db.Model(&team).UpdateColumn("member_count", h.db.Raw("member_count - 1"))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Left team successfully",
	})
}

// GetMyTeam returns the current user's team with all details
func (h *TeamHandler) GetMyTeam(c *gin.Context) {
	userID, _ := c.Get("userID")

	// Find the user's team membership
	var member models.TeamMember
	if err := h.db.Where("user_id = ?", userID).First(&member).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "You are not in any team"})
		return
	}

	// Load the team with all details
	var team models.Team
	if err := h.db.Preload("Owner").Preload("Members.User").First(&team, "id = ?", member.TeamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Calculate team totals from members' actual stats
	var totalClicks, totalConversions, totalPoints int
	for _, m := range team.Members {
		if m.Status == "active" {
			totalClicks += m.User.TotalClicks
			totalConversions += m.User.TotalConversions
			totalPoints += m.User.Points
		}
	}
	team.TotalClicks = totalClicks
	team.TotalConversions = totalConversions
	team.TotalPoints = totalPoints

	// Check if user is owner
	isOwner := team.OwnerID == userID.(uuid.UUID)

	// Get pending members if owner
	var pendingMembers []models.TeamMember
	if isOwner {
		h.db.Preload("User").Where("team_id = ? AND status = ?", team.ID, "pending").Find(&pendingMembers)
	}

	c.JSON(http.StatusOK, gin.H{
		"team":            team,
		"is_owner":        isOwner,
		"pending_members": pendingMembers,
		"invite_code":     team.InviteCode,
		"invite_url":      team.InviteURL,
	})
}

// JoinTeamByInviteCode allows joining a team using invite code
func (h *TeamHandler) JoinTeamByInviteCode(c *gin.Context) {
	code := c.Param("code")
	userID, _ := c.Get("userID")

	// Check if user is already in a team
	var existingMember models.TeamMember
	if err := h.db.Where("user_id = ?", userID).First(&existingMember).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "You are already in a team"})
		return
	}

	// Find team by invite code
	var team models.Team
	if err := h.db.Where("invite_code = ?", code).First(&team).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Invalid invite code"})
		return
	}

	// Check if team is full
	if team.MemberCount >= team.MaxMembers {
		c.JSON(http.StatusConflict, gin.H{"error": "Team is full"})
		return
	}

	// Create pending member
	member := models.TeamMember{
		ID:     uuid.New(),
		TeamID: team.ID,
		UserID: userID.(uuid.UUID),
		Role:   "member",
		Status: "pending",
		Points: 0,
	}

	if err := h.db.Create(&member).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send join request"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Join request sent successfully",
		"status":    "pending",
		"team_name": team.Name,
	})
}

// ApproveMember approves a pending member (owner only)
func (h *TeamHandler) ApproveMember(c *gin.Context) {
	teamID := c.Param("id")
	memberID := c.Param("memberId")
	userID, _ := c.Get("userID")

	// Verify owner
	var team models.Team
	if err := h.db.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if team.OwnerID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team owner can approve members"})
		return
	}

	// Find and update member
	var member models.TeamMember
	if err := h.db.Where("id = ? AND team_id = ? AND status = ?", memberID, teamID, "pending").First(&member).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pending member not found"})
		return
	}

	member.Status = "active"
	h.db.Save(&member)
	h.db.Model(&team).UpdateColumn("member_count", gorm.Expr("member_count + 1"))

	c.JSON(http.StatusOK, gin.H{
		"message": "Member approved successfully",
	})
}

// RejectMember rejects a pending member (owner only)
func (h *TeamHandler) RejectMember(c *gin.Context) {
	teamID := c.Param("id")
	memberID := c.Param("memberId")
	userID, _ := c.Get("userID")

	// Verify owner
	var team models.Team
	if err := h.db.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if team.OwnerID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team owner can reject members"})
		return
	}

	// Delete pending member
	result := h.db.Where("id = ? AND team_id = ? AND status = ?", memberID, teamID, "pending").Delete(&models.TeamMember{})
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Pending member not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Member rejected",
	})
}

// RemoveMember removes an active member (owner only)
func (h *TeamHandler) RemoveMember(c *gin.Context) {
	teamID := c.Param("id")
	memberID := c.Param("memberId")
	userID, _ := c.Get("userID")

	// Verify owner
	var team models.Team
	if err := h.db.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if team.OwnerID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team owner can remove members"})
		return
	}

	// Find member
	var member models.TeamMember
	if err := h.db.Where("id = ? AND team_id = ?", memberID, teamID).First(&member).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Member not found"})
		return
	}

	// Cannot remove owner
	if member.Role == "owner" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Cannot remove team owner"})
		return
	}

	h.db.Delete(&member)
	h.db.Model(&team).UpdateColumn("member_count", gorm.Expr("member_count - 1"))

	c.JSON(http.StatusOK, gin.H{
		"message": "Member removed successfully",
	})
}

// GetPendingRequests returns pending join requests (owner only)
func (h *TeamHandler) GetPendingRequests(c *gin.Context) {
	teamID := c.Param("id")
	userID, _ := c.Get("userID")

	// Verify owner
	var team models.Team
	if err := h.db.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if team.OwnerID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team owner can view pending requests"})
		return
	}

	var pendingMembers []models.TeamMember
	h.db.Preload("User").Where("team_id = ? AND status = ?", teamID, "pending").Find(&pendingMembers)

	c.JSON(http.StatusOK, gin.H{
		"pending_members": pendingMembers,
	})
}

// RegenerateInviteCode generates a new invite code (owner only)
func (h *TeamHandler) RegenerateInviteCode(c *gin.Context) {
	teamID := c.Param("id")
	userID, _ := c.Get("userID")

	// Verify owner
	var team models.Team
	if err := h.db.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if team.OwnerID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team owner can regenerate invite code"})
		return
	}

	// Generate new invite code
	newCode := generateInviteCode()
	team.InviteCode = newCode
	team.InviteURL = "https://afftok.com/join/" + newCode
	h.db.Save(&team)

	c.JSON(http.StatusOK, gin.H{
		"invite_code": team.InviteCode,
		"invite_url":  team.InviteURL,
	})
}

// DeleteTeam deletes a team (owner only)
func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	teamID := c.Param("id")
	userID, _ := c.Get("userID")

	// Verify owner
	var team models.Team
	if err := h.db.First(&team, "id = ?", teamID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	if team.OwnerID != userID.(uuid.UUID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only team owner can delete the team"})
		return
	}

	// Delete all members first
	h.db.Where("team_id = ?", teamID).Delete(&models.TeamMember{})

	// Delete team
	h.db.Delete(&team)

	c.JSON(http.StatusOK, gin.H{
		"message": "Team deleted successfully",
	})
}

// Helper function to generate invite code
func generateInviteCode() string {
	return uuid.New().String()[:8]
}

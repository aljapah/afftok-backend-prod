package handlers

import (
	"net/http"

	"github.com/afftok/backend/internal/models"
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
		Points: 0,
	}

	h.db.Create(&member)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Team created successfully",
		"team":    team,
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

package handlers

import (
	"fmt"
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
		InviteURL:   "https://go.afftokapp.com/api/invite/" + inviteCode,
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
	team.InviteURL = "https://go.afftokapp.com/api/invite/" + newCode
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

// GetTeamLandingPage serves the team invite landing page (public)
func (h *TeamHandler) GetTeamLandingPage(c *gin.Context) {
	code := c.Param("code")

	// Find team by invite code
	var team models.Team
	if err := h.db.Preload("Owner").Preload("Members.User").Where("invite_code = ?", code).First(&team).Error; err != nil {
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(`
<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>رابط غير صالح - AffTok</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
        }
        .container { text-align: center; padding: 40px; }
        h1 { font-size: 48px; margin-bottom: 20px; }
        p { font-size: 18px; opacity: 0.8; }
    </style>
</head>
<body>
    <div class="container">
        <h1>❌</h1>
        <h2>رابط الدعوة غير صالح</h2>
        <p>هذا الرابط غير موجود أو منتهي الصلاحية</p>
    </div>
</body>
</html>`))
		return
	}

	// Calculate team stats
	var totalClicks, totalConversions int
	activeMembers := 0
	for _, member := range team.Members {
		if member.Status == "active" {
			totalClicks += member.User.TotalClicks
			totalConversions += member.User.TotalConversions
			activeMembers++
		}
	}

	// Serve beautiful landing page
	html := `
<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>انضم لفريق ` + team.Name + ` - AffTok</title>
    <meta property="og:title" content="انضم لفريق ` + team.Name + ` على AffTok">
    <meta property="og:description" content="` + team.Description + `">
    <meta property="og:image" content="` + team.LogoURL + `">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { 
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a1a2e 0%, #16213e 50%, #0f3460 100%);
            min-height: 100vh;
            color: white;
        }
        .hero {
            padding: 60px 20px;
            text-align: center;
            background: linear-gradient(180deg, rgba(255,0,110,0.2) 0%, transparent 100%);
        }
        .team-logo {
            width: 120px;
            height: 120px;
            border-radius: 30px;
            background: linear-gradient(135deg, #ff006e, #ff4d00);
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 24px;
            font-size: 48px;
            box-shadow: 0 20px 60px rgba(255,0,110,0.4);
        }
        .team-logo img {
            width: 100%;
            height: 100%;
            border-radius: 30px;
            object-fit: cover;
        }
        h1 { font-size: 32px; margin-bottom: 12px; }
        .description { font-size: 16px; opacity: 0.8; max-width: 400px; margin: 0 auto 30px; }
        .stats {
            display: flex;
            justify-content: center;
            gap: 40px;
            margin: 30px 0;
        }
        .stat { text-align: center; }
        .stat-value { 
            font-size: 36px; 
            font-weight: bold; 
            background: linear-gradient(135deg, #ff006e, #ff4d00);
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
        }
        .stat-label { font-size: 14px; opacity: 0.7; }
        .join-btn {
            background: linear-gradient(135deg, #ff006e, #ff4d00);
            color: white;
            border: none;
            padding: 18px 60px;
            font-size: 20px;
            border-radius: 50px;
            cursor: pointer;
            margin: 30px 0;
            text-decoration: none;
            display: inline-block;
            box-shadow: 0 10px 40px rgba(255,0,110,0.4);
            transition: transform 0.3s, box-shadow 0.3s;
        }
        .join-btn:hover {
            transform: translateY(-3px);
            box-shadow: 0 15px 50px rgba(255,0,110,0.5);
        }
        .members {
            padding: 40px 20px;
            max-width: 500px;
            margin: 0 auto;
        }
        .members h3 { text-align: center; margin-bottom: 20px; opacity: 0.9; }
        .member {
            display: flex;
            align-items: center;
            gap: 15px;
            padding: 15px;
            background: rgba(255,255,255,0.05);
            border-radius: 16px;
            margin-bottom: 12px;
        }
        .member-avatar {
            width: 50px;
            height: 50px;
            border-radius: 50%;
            background: linear-gradient(135deg, #667eea, #764ba2);
        }
        .member-info { flex: 1; }
        .member-name { font-weight: 600; }
        .member-role { font-size: 12px; opacity: 0.6; }
        .owner-badge {
            background: gold;
            color: black;
            padding: 4px 12px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: bold;
        }
        .download-section {
            background: rgba(0,0,0,0.3);
            padding: 50px 20px;
            text-align: center;
        }
        .download-section h3 { margin-bottom: 15px; }
        .download-section p { opacity: 0.7; margin-bottom: 25px; }
        .store-buttons { display: flex; justify-content: center; gap: 15px; flex-wrap: wrap; }
        .store-btn {
            background: white;
            color: black;
            padding: 12px 24px;
            border-radius: 12px;
            text-decoration: none;
            display: flex;
            align-items: center;
            gap: 10px;
            font-weight: 600;
        }
        .footer {
            text-align: center;
            padding: 30px;
            opacity: 0.6;
            font-size: 14px;
        }
        .footer a { color: #ff006e; text-decoration: none; }
    </style>
</head>
<body>
    <div class="hero">
        <div class="team-logo">
            ` + func() string {
		if team.LogoURL != "" {
			return `<img src="` + team.LogoURL + `" alt="` + team.Name + `">`
		}
		return "👥"
	}() + `
        </div>
        <h1>` + team.Name + `</h1>
        <p class="description">` + team.Description + `</p>
        
        <div class="stats">
            <div class="stat">
                <div class="stat-value">` + fmt.Sprintf("%d", activeMembers) + `</div>
                <div class="stat-label">الأعضاء</div>
            </div>
            <div class="stat">
                <div class="stat-value">` + fmt.Sprintf("%d", totalConversions) + `</div>
                <div class="stat-label">التحويلات</div>
            </div>
            <div class="stat">
                <div class="stat-value">` + fmt.Sprintf("%d", totalClicks) + `</div>
                <div class="stat-label">النقرات</div>
            </div>
        </div>
        
        <a href="afftok://join/` + code + `" class="join-btn">🚀 انضم للفريق الآن</a>
    </div>
    
    <div class="members">
        <h3>أعضاء الفريق</h3>
        ` + func() string {
		membersHTML := ""
		for _, member := range team.Members {
			if member.Status == "active" {
				roleHTML := ""
				if member.Role == "owner" {
					roleHTML = `<span class="owner-badge">👑 القائد</span>`
				}
				membersHTML += `
                <div class="member">
                    <div class="member-avatar"></div>
                    <div class="member-info">
                        <div class="member-name">` + member.User.FullName + `</div>
                        <div class="member-role">@` + member.User.Username + `</div>
                    </div>
                    ` + roleHTML + `
                </div>`
			}
		}
		return membersHTML
	}() + `
    </div>
    
    <div class="download-section">
        <h3>📱 حمّل تطبيق AffTok</h3>
        <p>انضم لآلاف المروجين واكسب من عروض الأفلييت</p>
        <div class="store-buttons">
            <a href="https://apps.apple.com/app/afftok" class="store-btn">
                🍎 App Store
            </a>
            <a href="https://play.google.com/store/apps/details?id=com.afftok.app" class="store-btn">
                ▶️ Google Play
            </a>
        </div>
    </div>
    
    <div class="footer">
        <p>© 2025 AffTok - جميع الحقوق محفوظة</p>
        <p><a href="https://afftokapp.com">afftokapp.com</a></p>
    </div>
    
    <script>
        // Try to open app, fallback to store
        document.querySelector('.join-btn').addEventListener('click', function(e) {
            e.preventDefault();
            var appUrl = 'afftok://join/` + code + `';
            var storeUrl = /iPhone|iPad|iPod/i.test(navigator.userAgent) 
                ? 'https://apps.apple.com/app/afftok'
                : 'https://play.google.com/store/apps/details?id=com.afftok.app';
            
            var start = Date.now();
            window.location = appUrl;
            
            setTimeout(function() {
                if (Date.now() - start < 2000) {
                    window.location = storeUrl;
                }
            }, 1500);
        });
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

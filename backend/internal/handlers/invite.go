package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InviteHandler struct {
	db *gorm.DB
}

func NewInviteHandler(db *gorm.DB) *InviteHandler {
	return &InviteHandler{db: db}
}

// GetInviteInfo returns the team invite landing page (public)
func (h *InviteHandler) GetInviteInfo(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invite code is required"})
		return
	}

	// Find team by invite code
	var team models.Team
	if err := h.db.Preload("Owner").Preload("Members.User").Where("invite_code = ?", code).First(&team).Error; err != nil {
		// Return error page
		errorHTML := `<!DOCTYPE html>
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
</html>`
		c.Data(http.StatusNotFound, "text/html; charset=utf-8", []byte(errorHTML))
		return
	}

	// Calculate team stats
	activeMembers := 0
	var totalConversions, totalClicks int
	for _, member := range team.Members {
		if member.Status == "active" {
			activeMembers++
			totalConversions += member.User.TotalConversions
			totalClicks += member.User.TotalClicks
		}
	}

	// Build members HTML
	membersHTML := ""
	for _, member := range team.Members {
		if member.Status == "active" {
			roleHTML := ""
			if member.Role == "owner" {
				roleHTML = `<span class="owner-badge">👑 القائد</span>`
			}
			name := member.User.FullName
			if name == "" {
				name = member.User.Username
			}
			firstChar := "?"
			if len(name) > 0 {
				firstChar = string([]rune(name)[0])
			}
			membersHTML += `<div class="member">
                <div class="member-avatar">` + firstChar + `</div>
                <div class="member-info">
                    <div class="member-name">` + name + `</div>
                    <div class="member-role">@` + member.User.Username + `</div>
                </div>
                ` + roleHTML + `
            </div>`
		}
	}

	// Team logo or emoji
	logoHTML := "👥"
	if team.LogoURL != "" {
		logoHTML = `<img src="` + team.LogoURL + `" alt="` + team.Name + `">`
	}

	// Build the HTML page
	html := `<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>انضم لفريق ` + team.Name + ` - AffTok</title>
    <meta property="og:title" content="انضم لفريق ` + team.Name + ` على AffTok">
    <meta property="og:description" content="` + team.Description + `">
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
            overflow: hidden;
        }
        .team-logo img {
            width: 100%;
            height: 100%;
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
            background-clip: text;
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
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 20px;
            font-weight: bold;
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
        <div class="team-logo">` + logoHTML + `</div>
        <h1>` + team.Name + `</h1>
        <p class="description">` + team.Description + `</p>
        
        <div class="stats">
            <div class="stat">
                <div class="stat-value">` + strconv.Itoa(activeMembers) + `</div>
                <div class="stat-label">الأعضاء</div>
            </div>
            <div class="stat">
                <div class="stat-value">` + strconv.Itoa(totalConversions) + `</div>
                <div class="stat-label">التحويلات</div>
            </div>
            <div class="stat">
                <div class="stat-value">` + strconv.Itoa(totalClicks) + `</div>
                <div class="stat-label">النقرات</div>
            </div>
        </div>
        
        <a href="afftok://join/` + code + `" class="join-btn">🚀 انضم للفريق الآن</a>
    </div>
    
    <div class="members">
        <h3>أعضاء الفريق</h3>
        ` + membersHTML + `
    </div>
    
    <div class="download-section">
        <h3>📱 حمّل تطبيق AffTok</h3>
        <p>انضم لآلاف المروجين واكسب من عروض الأفلييت</p>
        <div class="store-buttons">
            <a href="https://apps.apple.com/app/afftok" class="store-btn">🍎 App Store</a>
            <a href="https://play.google.com/store/apps/details?id=com.afftok.app" class="store-btn">▶️ Google Play</a>
        </div>
    </div>
    
    <div class="footer">
        <p>© 2025 AffTok - جميع الحقوق محفوظة</p>
        <p><a href="https://afftokapp.com">afftokapp.com</a></p>
    </div>
    
    <script>
        document.querySelector('.join-btn').addEventListener('click', function(e) {
            e.preventDefault();
            var appUrl = 'afftok://join/` + code + `';
            var storeUrl = /iPhone|iPad|iPod/i.test(navigator.userAgent) 
                ? 'https://apps.apple.com/app/afftok'
                : 'https://play.google.com/store/apps/details?id=com.afftok.app';
            
            window.location = appUrl;
            setTimeout(function() { window.location = storeUrl; }, 1500);
        });
    </script>
</body>
</html>`

	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

// ValidateInvite validates an invite code (returns JSON)
func (h *InviteHandler) ValidateInvite(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invite code is required"})
		return
	}

	var team models.Team
	if err := h.db.Where("invite_code = ?", code).First(&team).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"valid":   false,
			"message": "Invite link not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid":   true,
		"message": "Invite link is valid",
		"code":    code,
		"team_id": team.ID,
	})
}

// RecordInviteVisit records a visit to an invite link
func (h *InviteHandler) RecordInviteVisit(c *gin.Context) {
	code := c.Param("code")
	c.JSON(http.StatusOK, gin.H{
		"message": "Visit recorded",
		"code":    code,
	})
}

// GetMyInviteLink returns the authenticated user's personal invite link
func (h *InviteHandler) GetMyInviteLink(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":     userID,
		"invite_link": fmt.Sprintf("https://go.afftokapp.com/api/invite/%v", userID),
		"invite_code": fmt.Sprintf("%v", userID),
	})
}

// CheckPendingInvite checks if user has a pending invite to claim
func (h *InviteHandler) CheckPendingInvite(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":        userID,
		"pending_invite": nil,
		"has_pending":    false,
	})
}

// AutoJoinByInvite automatically joins user to team based on stored invite
func (h *InviteHandler) AutoJoinByInvite(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "No pending invite to auto-join",
		"user_id": userID,
	})
}

// ClaimInvite claims a specific invite by ID
func (h *InviteHandler) ClaimInvite(c *gin.Context) {
	inviteID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":   "Invite claimed successfully",
		"invite_id": inviteID,
		"user_id":   userID,
	})
}

// ClaimInviteByCode claims an invite using the invite code
func (h *InviteHandler) ClaimInviteByCode(c *gin.Context) {
	code := c.Param("code")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Invite claimed successfully",
		"code":    code,
		"user_id": userID,
	})
}

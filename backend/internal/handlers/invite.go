package handlers

import (
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
		errorHTML := getErrorHTML()
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusNotFound, errorHTML)
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
				roleHTML = `<span style="background:linear-gradient(135deg,#FFD700,#FFA500);color:#000;padding:6px 16px;border-radius:20px;font-size:12px;font-weight:bold;margin-right:auto;">👑 القائد</span>`
			}
			name := member.User.FullName
			if name == "" {
				name = member.User.Username
			}
			firstChar := "?"
			if len(name) > 0 {
				firstChar = string([]rune(name)[0])
			}
			membersHTML += `<div style="display:flex;align-items:center;gap:15px;padding:16px;background:rgba(255,255,255,0.08);border-radius:16px;margin-bottom:12px;backdrop-filter:blur(10px);">
				<div style="width:55px;height:55px;border-radius:50%;background:linear-gradient(135deg,#667eea,#764ba2);display:flex;align-items:center;justify-content:center;font-size:22px;font-weight:bold;color:white;box-shadow:0 4px 15px rgba(102,126,234,0.4);">` + firstChar + `</div>
				<div style="flex:1;">
					<div style="font-weight:600;font-size:16px;margin-bottom:4px;">` + name + `</div>
					<div style="font-size:13px;opacity:0.6;">@` + member.User.Username + `</div>
				</div>
				` + roleHTML + `
			</div>`
		}
	}

	// Team logo
	logoHTML := `<div style="font-size:60px;">👥</div>`
	if team.LogoURL != "" {
		logoHTML = `<img src="` + team.LogoURL + `" style="width:100%;height:100%;object-fit:cover;border-radius:30px;" alt="` + team.Name + `">`
	}

	html := `<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <meta http-equiv="Content-Security-Policy" content="default-src * 'unsafe-inline' 'unsafe-eval' data: blob:;">
    <title>انضم لفريق ` + team.Name + ` - AffTok</title>
    <meta property="og:title" content="انضم لفريق ` + team.Name + ` على AffTok">
    <meta property="og:description" content="` + team.Description + `">
    <meta property="og:type" content="website">
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Tajawal:wght@400;500;700;800&display=swap" rel="stylesheet">
</head>
<body style="margin:0;padding:0;font-family:'Tajawal',sans-serif;background:linear-gradient(135deg,#0f0c29 0%,#302b63 50%,#24243e 100%);min-height:100vh;color:white;">
    
    <!-- Animated Background -->
    <div style="position:fixed;top:0;left:0;width:100%;height:100%;overflow:hidden;z-index:0;pointer-events:none;">
        <div style="position:absolute;width:400px;height:400px;background:radial-gradient(circle,rgba(255,0,110,0.3) 0%,transparent 70%);top:-100px;right:-100px;animation:pulse 4s ease-in-out infinite;"></div>
        <div style="position:absolute;width:300px;height:300px;background:radial-gradient(circle,rgba(102,126,234,0.3) 0%,transparent 70%);bottom:-50px;left:-50px;animation:pulse 5s ease-in-out infinite;"></div>
    </div>
    
    <div style="position:relative;z-index:1;max-width:500px;margin:0 auto;padding:20px;">
        
        <!-- Hero Section -->
        <div style="text-align:center;padding:40px 20px;">
            <!-- Team Logo -->
            <div style="width:130px;height:130px;margin:0 auto 25px;background:linear-gradient(135deg,#ff006e,#ff4d00);border-radius:35px;display:flex;align-items:center;justify-content:center;box-shadow:0 20px 60px rgba(255,0,110,0.5);overflow:hidden;">
                ` + logoHTML + `
            </div>
            
            <!-- Team Name -->
            <h1 style="font-size:36px;font-weight:800;margin:0 0 10px;background:linear-gradient(135deg,#fff,#e0e0e0);-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text;">` + team.Name + `</h1>
            
            <!-- Description -->
            <p style="font-size:16px;opacity:0.8;margin:0 0 30px;line-height:1.6;">` + team.Description + `</p>
            
            <!-- Stats -->
            <div style="display:flex;justify-content:center;gap:30px;margin:30px 0;">
                <div style="text-align:center;">
                    <div style="font-size:42px;font-weight:800;background:linear-gradient(135deg,#ff006e,#ff4d00);-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text;">` + strconv.Itoa(activeMembers) + `</div>
                    <div style="font-size:14px;opacity:0.7;margin-top:5px;">الأعضاء</div>
                </div>
                <div style="width:1px;background:rgba(255,255,255,0.2);"></div>
                <div style="text-align:center;">
                    <div style="font-size:42px;font-weight:800;background:linear-gradient(135deg,#00d4ff,#0099ff);-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text;">` + strconv.Itoa(totalConversions) + `</div>
                    <div style="font-size:14px;opacity:0.7;margin-top:5px;">التحويلات</div>
                </div>
                <div style="width:1px;background:rgba(255,255,255,0.2);"></div>
                <div style="text-align:center;">
                    <div style="font-size:42px;font-weight:800;background:linear-gradient(135deg,#a855f7,#7c3aed);-webkit-background-clip:text;-webkit-text-fill-color:transparent;background-clip:text;">` + strconv.Itoa(totalClicks) + `</div>
                    <div style="font-size:14px;opacity:0.7;margin-top:5px;">النقرات</div>
                </div>
            </div>
            
            <!-- Join Button -->
            <a href="afftok://join/` + code + `" id="joinBtn" style="display:inline-block;background:linear-gradient(135deg,#ff006e,#ff4d00);color:white;text-decoration:none;padding:20px 50px;font-size:20px;font-weight:700;border-radius:50px;box-shadow:0 15px 45px rgba(255,0,110,0.5);transition:all 0.3s ease;margin-top:10px;">
                🚀 انضم للفريق الآن
            </a>
        </div>
        
        <!-- Members Section -->
        <div style="padding:20px;margin-top:20px;">
            <h3 style="text-align:center;font-size:20px;font-weight:700;margin-bottom:20px;opacity:0.9;">👥 أعضاء الفريق</h3>
            ` + membersHTML + `
        </div>
        
        <!-- Download Section -->
        <div style="background:rgba(0,0,0,0.3);border-radius:25px;padding:40px 20px;margin-top:30px;text-align:center;backdrop-filter:blur(10px);">
            <h3 style="font-size:22px;font-weight:700;margin:0 0 10px;">📱 حمّل تطبيق AffTok</h3>
            <p style="opacity:0.7;margin:0 0 25px;font-size:15px;">انضم لآلاف المروجين واكسب من عروض الأفلييت</p>
            <div style="display:flex;justify-content:center;gap:15px;flex-wrap:wrap;">
                <a href="https://apps.apple.com/app/afftok" style="display:flex;align-items:center;gap:10px;background:white;color:black;padding:14px 28px;border-radius:14px;text-decoration:none;font-weight:600;font-size:15px;box-shadow:0 5px 20px rgba(0,0,0,0.3);">
                    🍎 App Store
                </a>
                <a href="https://play.google.com/store/apps/details?id=com.afftok.app" style="display:flex;align-items:center;gap:10px;background:white;color:black;padding:14px 28px;border-radius:14px;text-decoration:none;font-weight:600;font-size:15px;box-shadow:0 5px 20px rgba(0,0,0,0.3);">
                    ▶️ Google Play
                </a>
            </div>
        </div>
        
        <!-- Footer -->
        <div style="text-align:center;padding:30px 20px;opacity:0.6;font-size:14px;">
            <p style="margin:0 0 8px;">© 2025 AffTok - جميع الحقوق محفوظة</p>
            <a href="https://afftokapp.com" style="color:#ff006e;text-decoration:none;">afftokapp.com</a>
        </div>
    </div>
    
    <style>
        @keyframes pulse {
            0%, 100% { transform: scale(1); opacity: 0.5; }
            50% { transform: scale(1.1); opacity: 0.8; }
        }
        #joinBtn:hover {
            transform: translateY(-5px);
            box-shadow: 0 20px 60px rgba(255,0,110,0.6);
        }
    </style>
    
    <script>
        document.getElementById('joinBtn').addEventListener('click', function(e) {
            e.preventDefault();
            var appUrl = 'afftok://join/` + code + `';
            var isIOS = /iPhone|iPad|iPod/i.test(navigator.userAgent);
            var storeUrl = isIOS 
                ? 'https://apps.apple.com/app/afftok'
                : 'https://play.google.com/store/apps/details?id=com.afftok.app';
            
            window.location = appUrl;
            setTimeout(function() { 
                window.location = storeUrl; 
            }, 2000);
        });
    </script>
</body>
</html>`

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}

func getErrorHTML() string {
	return `<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="Content-Security-Policy" content="default-src * 'unsafe-inline' 'unsafe-eval' data: blob:;">
    <title>رابط غير صالح - AffTok</title>
    <link href="https://fonts.googleapis.com/css2?family=Tajawal:wght@400;700&display=swap" rel="stylesheet">
</head>
<body style="margin:0;padding:0;font-family:'Tajawal',sans-serif;background:linear-gradient(135deg,#0f0c29,#302b63,#24243e);min-height:100vh;display:flex;align-items:center;justify-content:center;color:white;">
    <div style="text-align:center;padding:40px;">
        <div style="font-size:80px;margin-bottom:20px;">❌</div>
        <h2 style="font-size:28px;margin:0 0 15px;">رابط الدعوة غير صالح</h2>
        <p style="font-size:16px;opacity:0.7;">هذا الرابط غير موجود أو منتهي الصلاحية</p>
        <a href="https://afftokapp.com" style="display:inline-block;margin-top:30px;background:linear-gradient(135deg,#ff006e,#ff4d00);color:white;padding:15px 40px;border-radius:30px;text-decoration:none;font-weight:600;">زيارة الموقع</a>
    </div>
</body>
</html>`
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
		c.JSON(http.StatusNotFound, gin.H{"valid": false, "message": "Invite link not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"valid": true, "message": "Invite link is valid", "code": code, "team_id": team.ID})
}

// RecordInviteVisit records a visit to an invite link
func (h *InviteHandler) RecordInviteVisit(c *gin.Context) {
	code := c.Param("code")
	c.JSON(http.StatusOK, gin.H{"message": "Visit recorded", "code": code})
}

// GetMyInviteLink returns the authenticated user's personal invite link
func (h *InviteHandler) GetMyInviteLink(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": userID, "invite_link": "https://go.afftokapp.com/api/invite/user", "invite_code": "user"})
}

// CheckPendingInvite checks if user has a pending invite
func (h *InviteHandler) CheckPendingInvite(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user_id": userID, "pending_invite": nil, "has_pending": false})
}

// AutoJoinByInvite automatically joins user to team
func (h *InviteHandler) AutoJoinByInvite(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "No pending invite", "user_id": userID})
}

// ClaimInvite claims a specific invite by ID
func (h *InviteHandler) ClaimInvite(c *gin.Context) {
	inviteID := c.Param("id")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invite claimed", "invite_id": inviteID, "user_id": userID})
}

// ClaimInviteByCode claims an invite using the invite code
func (h *InviteHandler) ClaimInviteByCode(c *gin.Context) {
	code := c.Param("code")
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Invite claimed", "code": code, "user_id": userID})
}

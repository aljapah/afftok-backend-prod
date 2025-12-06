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
				roleHTML = `<span class="owner-badge">👑 <span data-lang="owner">القائد</span></span>`
			}
			name := member.User.FullName
			if name == "" {
				name = member.User.Username
			}
			firstChar := "?"
			if len(name) > 0 {
				firstChar = string([]rune(name)[0])
			}
			membersHTML += `<div class="member-card">
				<div class="member-avatar">` + firstChar + `</div>
				<div class="member-info">
					<div class="member-name">` + name + `</div>
					<div class="member-username">@` + member.User.Username + `</div>
				</div>
				` + roleHTML + `
			</div>`
		}
	}

	// Team logo
	logoHTML := `<div class="team-emoji">👥</div>`
	if team.LogoURL != "" {
		logoHTML = `<img src="` + team.LogoURL + `" alt="` + team.Name + `" class="team-logo-img">`
	}

	html := `<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + team.Name + ` - AffTok</title>
    <meta property="og:title" content="انضم لفريق ` + team.Name + ` على AffTok">
    <meta property="og:description" content="` + team.Description + `">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        
        :root {
            --primary: #FF006E;
            --secondary: #FF4D00;
            --accent: #8E2DE2;
            --bg-dark: #0a0a0a;
            --bg-card: #1a1a1a;
            --bg-light: #242424;
            --border-color: #333333;
            --text-primary: #ffffff;
            --text-secondary: #a8a8a8;
        }
        
        html { scroll-behavior: smooth; }
        
        body {
            font-family: 'Segoe UI', 'Roboto', -apple-system, BlinkMacSystemFont, sans-serif;
            background: linear-gradient(135deg, var(--bg-dark) 0%, #0f0f0f 100%);
            color: var(--text-primary);
            line-height: 1.6;
            min-height: 100vh;
        }
        
        body.en { direction: ltr; }
        
        .container { max-width: 800px; margin: 0 auto; padding: 0 20px; }
        
        /* Header */
        header {
            background: rgba(15, 15, 15, 0.9);
            border-bottom: 1px solid var(--border-color);
            padding: 12px 0;
            position: sticky;
            top: 0;
            z-index: 100;
            backdrop-filter: blur(10px);
        }
        
        .header-content {
            display: flex;
            justify-content: space-between;
            align-items: center;
        }
        
        .logo {
            font-size: 28px;
            font-weight: 700;
            background: linear-gradient(135deg, var(--primary), var(--secondary));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        
        .language-toggle {
            padding: 8px 16px;
            background: var(--bg-card);
            border: 1px solid var(--border-color);
            border-radius: 8px;
            color: var(--text-primary);
            cursor: pointer;
            font-weight: 600;
            font-size: 13px;
            transition: all 0.3s ease;
        }
        
        .language-toggle:hover {
            background: rgba(255, 0, 110, 0.1);
            border-color: var(--primary);
            color: var(--primary);
        }
        
        /* Hero Section */
        .hero {
            text-align: center;
            padding: 50px 20px;
            background: linear-gradient(180deg, rgba(255, 0, 110, 0.08) 0%, transparent 100%);
            border-bottom: 1px solid var(--border-color);
        }
        
        .team-logo {
            width: 120px;
            height: 120px;
            margin: 0 auto 20px;
            background: linear-gradient(135deg, var(--primary), var(--secondary));
            border-radius: 30px;
            display: flex;
            align-items: center;
            justify-content: center;
            box-shadow: 0 20px 50px rgba(255, 0, 110, 0.4);
            overflow: hidden;
        }
        
        .team-logo-img { width: 100%; height: 100%; object-fit: cover; }
        .team-emoji { font-size: 50px; }
        
        .team-name {
            font-size: 36px;
            font-weight: 700;
            margin-bottom: 10px;
            letter-spacing: -0.5px;
        }
        
        .team-description {
            color: var(--text-secondary);
            font-size: 16px;
            margin-bottom: 30px;
            max-width: 500px;
            margin-left: auto;
            margin-right: auto;
        }
        
        /* Stats */
        .stats {
            display: flex;
            justify-content: center;
            gap: 40px;
            margin: 30px 0;
            flex-wrap: wrap;
        }
        
        .stat {
            text-align: center;
            padding: 0 20px;
        }
        
        .stat:not(:last-child) {
            border-left: 1px solid var(--border-color);
        }
        
        body.en .stat:not(:last-child) {
            border-left: none;
            border-right: 1px solid var(--border-color);
        }
        
        .stat-value {
            font-size: 40px;
            font-weight: 800;
            background: linear-gradient(135deg, var(--primary), var(--secondary));
            -webkit-background-clip: text;
            -webkit-text-fill-color: transparent;
            background-clip: text;
        }
        
        .stat-label {
            font-size: 14px;
            color: var(--text-secondary);
            margin-top: 5px;
        }
        
        /* Join Button */
        .join-btn {
            display: inline-block;
            background: linear-gradient(135deg, var(--primary), var(--secondary));
            color: white;
            text-decoration: none;
            padding: 18px 50px;
            font-size: 18px;
            font-weight: 700;
            border-radius: 50px;
            box-shadow: 0 10px 40px rgba(255, 0, 110, 0.4);
            transition: all 0.3s ease;
            margin-top: 20px;
            border: none;
            cursor: pointer;
        }
        
        .join-btn:hover {
            transform: translateY(-3px);
            box-shadow: 0 15px 50px rgba(255, 0, 110, 0.5);
        }
        
        /* Members Section */
        .members-section {
            padding: 50px 0;
        }
        
        .section-title {
            font-size: 24px;
            font-weight: 700;
            text-align: center;
            margin-bottom: 30px;
        }
        
        .members-grid {
            display: grid;
            gap: 15px;
            max-width: 500px;
            margin: 0 auto;
        }
        
        .member-card {
            display: flex;
            align-items: center;
            gap: 15px;
            padding: 16px;
            background: var(--bg-card);
            border: 1px solid var(--border-color);
            border-radius: 12px;
            transition: all 0.3s ease;
        }
        
        .member-card:hover {
            border-color: var(--primary);
            transform: translateY(-2px);
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
            color: white;
        }
        
        .member-info { flex: 1; }
        .member-name { font-weight: 600; font-size: 16px; }
        .member-username { font-size: 13px; color: var(--text-secondary); }
        
        .owner-badge {
            background: linear-gradient(135deg, #FFD700, #FFA500);
            color: #000;
            padding: 6px 14px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: 600;
        }
        
        /* Download Section */
        .download-section {
            background: linear-gradient(135deg, rgba(255, 0, 110, 0.08), rgba(255, 77, 0, 0.08));
            border: 1px solid var(--border-color);
            border-radius: 12px;
            padding: 40px;
            text-align: center;
            margin: 40px 20px;
        }
        
        .download-title { font-size: 22px; font-weight: 700; margin-bottom: 10px; }
        .download-subtitle { color: var(--text-secondary); margin-bottom: 24px; font-size: 14px; }
        
        .download-buttons {
            display: flex;
            gap: 16px;
            justify-content: center;
            flex-wrap: wrap;
        }
        
        .download-btn {
            background: white;
            color: black;
            padding: 12px 24px;
            border-radius: 10px;
            border: none;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.3s ease;
            display: flex;
            align-items: center;
            gap: 8px;
            font-size: 14px;
            text-decoration: none;
        }
        
        .download-btn:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 20px rgba(0, 0, 0, 0.3);
        }
        
        /* Social Section */
        .social-section {
            text-align: center;
            padding: 40px 0;
            border-top: 1px solid var(--border-color);
            border-bottom: 1px solid var(--border-color);
        }
        
        .social-title { font-size: 18px; font-weight: 700; margin-bottom: 20px; }
        
        .social-icons {
            display: flex;
            justify-content: center;
            gap: 20px;
            flex-wrap: wrap;
        }
        
        .social-link {
            color: var(--primary);
            font-size: 28px;
            transition: all 0.3s ease;
            text-decoration: none;
        }
        
        .social-link:hover {
            transform: scale(1.2);
            color: var(--secondary);
        }
        
        /* Footer */
        footer {
            background: rgba(15, 15, 15, 0.5);
            border-top: 1px solid var(--border-color);
            padding: 30px 0;
            text-align: center;
        }
        
        .support-email {
            color: var(--text-secondary);
            font-size: 13px;
            margin-bottom: 20px;
        }
        
        .support-email a {
            color: var(--primary);
            text-decoration: none;
            font-weight: 600;
        }
        
        .footer-links {
            display: flex;
            gap: 24px;
            justify-content: center;
            flex-wrap: wrap;
            margin-bottom: 20px;
        }
        
        .footer-link {
            color: var(--text-secondary);
            text-decoration: none;
            font-size: 13px;
            transition: color 0.3s ease;
        }
        
        .footer-link:hover { color: var(--primary); }
        
        .copyright {
            color: var(--text-secondary);
            font-size: 12px;
        }
        
        @media (max-width: 600px) {
            .stats { gap: 20px; }
            .stat { padding: 0 15px; }
            .stat-value { font-size: 32px; }
            .team-name { font-size: 28px; }
            .download-buttons { flex-direction: column; }
            .download-btn { width: 100%; justify-content: center; }
        }
    </style>
</head>
<body>
    <header>
        <div class="container">
            <div class="header-content">
                <div style="width: 80px;"></div>
                <div class="logo">AffTok</div>
                <button class="language-toggle" onclick="toggleLanguage()" id="lang-btn">English</button>
            </div>
        </div>
    </header>

    <div class="hero">
        <div class="container">
            <div class="team-logo">` + logoHTML + `</div>
            <h1 class="team-name">` + team.Name + `</h1>
            <p class="team-description">` + team.Description + `</p>
            
            <div class="stats">
                <div class="stat">
                    <div class="stat-value">` + strconv.Itoa(activeMembers) + `</div>
                    <div class="stat-label" data-lang="members">الأعضاء</div>
                </div>
                <div class="stat">
                    <div class="stat-value">` + strconv.Itoa(totalConversions) + `</div>
                    <div class="stat-label" data-lang="conversions">التحويلات</div>
                </div>
                <div class="stat">
                    <div class="stat-value">` + strconv.Itoa(totalClicks) + `</div>
                    <div class="stat-label" data-lang="clicks">النقرات</div>
                </div>
            </div>
            
            <a href="afftok://join/` + code + `" class="join-btn" id="joinBtn">
                🚀 <span data-lang="join_now">انضم للفريق الآن</span>
            </a>
        </div>
    </div>

    <div class="members-section">
        <div class="container">
            <h2 class="section-title">👥 <span data-lang="team_members">أعضاء الفريق</span></h2>
            <div class="members-grid">` + membersHTML + `</div>
        </div>
    </div>

    <div class="container">
        <div class="download-section">
            <h3 class="download-title" data-lang="download_title">📱 حمّل تطبيق AffTok</h3>
            <p class="download-subtitle" data-lang="download_subtitle">انضم لآلاف المروجين واكسب من عروض الأفلييت</p>
            <div class="download-buttons">
                <a href="https://apps.apple.com/app/afftok" class="download-btn">
                    <i class="fab fa-apple"></i>
                    <span>App Store</span>
                </a>
                <a href="https://play.google.com/store/apps/details?id=com.afftok.app" class="download-btn">
                    <i class="fab fa-google-play"></i>
                    <span>Google Play</span>
                </a>
            </div>
        </div>
    </div>

    <div class="container">
        <div class="social-section">
            <h3 class="social-title" data-lang="follow_us">تابعنا على وسائل التواصل</h3>
            <div class="social-icons">
                <a href="https://afftokapp.com" target="_blank" class="social-link"><i class="fas fa-globe"></i></a>
                <a href="https://www.instagram.com/afftok_app" target="_blank" class="social-link"><i class="fab fa-instagram"></i></a>
                <a href="https://www.tiktok.com/@afftok_app" target="_blank" class="social-link"><i class="fab fa-tiktok"></i></a>
                <a href="https://twitter.com/afftokapp" target="_blank" class="social-link">𝕏</a>
                <a href="https://www.youtube.com/@afftok" target="_blank" class="social-link"><i class="fab fa-youtube"></i></a>
            </div>
        </div>
    </div>

    <footer>
        <div class="container">
            <div class="support-email">
                <span data-lang="support">للدعم والاستفسارات:</span>
                <a href="mailto:support@afftokapp.com">support@afftokapp.com</a>
            </div>
            <div class="footer-links">
                <a href="https://afftokapp.com/privacy.html" class="footer-link" data-lang="privacy" target="_blank">سياسة الخصوصية</a>
                <a href="https://afftokapp.com/terms.html" class="footer-link" data-lang="terms" target="_blank">شروط الاستخدام</a>
            </div>
            <p class="copyright" data-lang="copyright">© 2025 AffTok. جميع الحقوق محفوظة.</p>
        </div>
    </footer>

    <script>
        let currentLang = 'ar';
        
        const translations = {
            ar: {
                lang_btn: 'English',
                members: 'الأعضاء',
                conversions: 'التحويلات',
                clicks: 'النقرات',
                join_now: 'انضم للفريق الآن',
                team_members: 'أعضاء الفريق',
                owner: 'القائد',
                download_title: '📱 حمّل تطبيق AffTok',
                download_subtitle: 'انضم لآلاف المروجين واكسب من عروض الأفلييت',
                follow_us: 'تابعنا على وسائل التواصل',
                support: 'للدعم والاستفسارات:',
                privacy: 'سياسة الخصوصية',
                terms: 'شروط الاستخدام',
                copyright: '© 2025 AffTok. جميع الحقوق محفوظة.'
            },
            en: {
                lang_btn: 'العربية',
                members: 'Members',
                conversions: 'Conversions',
                clicks: 'Clicks',
                join_now: 'Join Team Now',
                team_members: 'Team Members',
                owner: 'Owner',
                download_title: '📱 Download AffTok App',
                download_subtitle: 'Join thousands of promoters and earn from affiliate offers',
                follow_us: 'Follow us on social media',
                support: 'For support and inquiries:',
                privacy: 'Privacy Policy',
                terms: 'Terms of Use',
                copyright: '© 2025 AffTok. All rights reserved.'
            }
        };
        
        function toggleLanguage() {
            currentLang = currentLang === 'ar' ? 'en' : 'ar';
            document.documentElement.lang = currentLang;
            document.documentElement.dir = currentLang === 'ar' ? 'rtl' : 'ltr';
            document.body.classList.toggle('en', currentLang === 'en');
            updateTranslations();
        }
        
        function updateTranslations() {
            document.getElementById('lang-btn').textContent = translations[currentLang].lang_btn;
            document.querySelectorAll('[data-lang]').forEach(el => {
                const key = el.getAttribute('data-lang');
                if (translations[currentLang][key]) {
                    el.textContent = translations[currentLang][key];
                }
            });
        }
        
        document.getElementById('joinBtn').addEventListener('click', function(e) {
            e.preventDefault();
            var appUrl = 'afftok://join/` + code + `';
            var isIOS = /iPhone|iPad|iPod/i.test(navigator.userAgent);
            var storeUrl = isIOS 
                ? 'https://apps.apple.com/app/afftok'
                : 'https://play.google.com/store/apps/details?id=com.afftok.app';
            
            window.location = appUrl;
            setTimeout(function() { window.location = storeUrl; }, 2000);
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
    <title>رابط غير صالح - AffTok</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Segoe UI', sans-serif;
            background: linear-gradient(135deg, #0a0a0a, #0f0f0f);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
        }
        .container { text-align: center; padding: 40px; }
        .icon { font-size: 80px; margin-bottom: 20px; }
        h2 { font-size: 28px; margin-bottom: 15px; }
        p { color: #a8a8a8; font-size: 16px; margin-bottom: 30px; }
        .btn {
            display: inline-block;
            background: linear-gradient(135deg, #FF006E, #FF4D00);
            color: white;
            padding: 15px 40px;
            border-radius: 30px;
            text-decoration: none;
            font-weight: 600;
            transition: all 0.3s ease;
        }
        .btn:hover { transform: translateY(-3px); box-shadow: 0 10px 30px rgba(255,0,110,0.4); }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">❌</div>
        <h2>رابط الدعوة غير صالح</h2>
        <p>هذا الرابط غير موجود أو منتهي الصلاحية</p>
        <a href="https://afftokapp.com" class="btn">زيارة الموقع</a>
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

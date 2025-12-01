package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PromoterHandler struct {
	db *gorm.DB
}

func NewPromoterHandler(db *gorm.DB) *PromoterHandler {
	return &PromoterHandler{db: db}
}

func (h *PromoterHandler) GetPromoterPage(c *gin.Context) {
	userID := c.Param("id")

	id, err := uuid.Parse(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var user models.AfftokUser
	if err := h.db.Where("id = ?", id).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	h.servePromoterPage(c, user)
}

func (h *PromoterHandler) GetPromoterPageByUsername(c *gin.Context) {
	username := c.Param("username")

	var user models.AfftokUser
	if err := h.db.Where("username = ?", username).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	h.servePromoterPage(c, user)
}

func (h *PromoterHandler) servePromoterPage(c *gin.Context, user models.AfftokUser) {

	var offers []models.Offer
	if err := h.db.Where("status = ?", "active").Order("created_at DESC").Find(&offers).Error; err != nil {
		offers = []models.Offer{}
	}

	var totalClicks int64
	var totalOffers int64

	h.db.Model(&models.Click{}).
		Joins("JOIN user_offers ON clicks.user_offer_id = user_offers.id").
		Where("user_offers.user_id = ? AND user_offers.status = ?", user.ID, "active").
		Count(&totalClicks)

	h.db.Model(&models.UserOffer{}).
		Where("user_id = ? AND status = ?", user.ID, "active").
		Count(&totalOffers)

	html := h.generateHTML(user, offers, int(totalOffers), int(totalClicks))
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func (h *PromoterHandler) generateHTML(user models.AfftokUser, offers []models.Offer, totalOffers int, totalClicks int) string {
	filePath := "public/promoter_landing.html"
	if _, err := os.Stat(filePath); err == nil {
		data, err := ioutil.ReadFile(filePath)
		if err == nil {
			return string(data)
		}
	}

	promoterID := user.ID.String()

	offersHTML := ""
	for _, offer := range offers {
		commissionText := ""
		if offer.PayoutType == "cpa" {
			commissionText = fmt.Sprintf("CPA $%.0f", float64(offer.Payout))
		} else if offer.PayoutType == "percentage" {
			commissionText = fmt.Sprintf("%.0f%% Cashback", float64(offer.Payout))
		} else {
			commissionText = fmt.Sprintf("$%.2f", float64(offer.Payout))
		}

		categoryBadge := getCategoryBadge(offer.Category)

		imageURL := offer.ImageURL
		if imageURL == "" {
			imageURL = "https://via.placeholder.com/400x200?text=" + offer.Title
		}

		offersHTML += fmt.Sprintf(`
			<div class="offer-card">
				<div class="offer-image" style="background-image: url('%s')">
					<div class="category-badge %s">%s</div>
				</div>
				<div class="offer-content">
					<h3 class="offer-title">%s</h3>
					<p class="offer-description">%s</p>
					<div class="offer-meta">
						<div class="offer-commission">%s</div>
					</div>
				<a href="/api/c/%s?promoter=%s" class="get-link-btn" target="_blank">
					<span class="btn-icon">🔗</span>
					<span class="btn-text">Get Link</span>
				</a>
				</div>
			</div>
		`, imageURL, categoryBadge, offer.Category, offer.Title, offer.Description,
			commissionText, offer.ID, promoterID)
	}

	bioText := user.Bio
	if bioText == "" {
		bioText = "Welcome to my page! I share the best offers and deals to help you save money and earn rewards. Check out my curated selection below!"
	}

	bioHTML := fmt.Sprintf(`
		<div class="bio-section">
			<p class="bio-text">%s</p>
		</div>
	`, bioText)

	fullName := user.FullName
	if fullName == "" {
		fullName = user.Username
	}

	avatarLetter := "👤"
	if len(user.Username) > 0 {
		avatarLetter = string(user.Username[0])
	}

	promoterRating := h.GetPromoterRating(user.ID)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html lang="ar" dir="rtl">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s - AffTok</title>
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.4.0/css/all.min.css">
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}

		:root {
			--primary: #FF006E;
			--secondary: #FF4D00;
			--accent: #8E2DE2;
			--dark-accent: #4A00E0;
			--bg-dark: #0a0a0a;
			--bg-card: #1a1a1a;
			--bg-light: #242424;
			--border-color: #333333;
			--text-primary: #ffffff;
			--text-secondary: #a8a8a8;
		}

		html {
			scroll-behavior: smooth;
		}

		body {
			font-family: 'Segoe UI', 'Roboto', 'Helvetica Neue', -apple-system, BlinkMacSystemFont, Arial, sans-serif;
			background: linear-gradient(135deg, var(--bg-dark) 0%%, #0f0f0f 100%%);
			color: var(--text-primary);
			line-height: 1.6;
			min-height: 100vh;
		}

		body.en {
			direction: ltr;
		}

		body.en * {
			direction: ltr;
		}

		.container {
			max-width: 1200px;
			margin: 0 auto;
			padding: 0 20px;
		}

		header {
			background: rgba(15, 15, 15, 0.8);
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
			font-size: 32px;
			font-weight: 700;
			letter-spacing: -0.5px;
			background: linear-gradient(135deg, var(--primary), var(--secondary));
			-webkit-background-clip: text;
			-webkit-text-fill-color: transparent;
			background-clip: text;
			order: 2;
		}

		.language-toggle {
			padding: 8px 16px;
			background: var(--bg-card);
			border: 1px solid var(--border-color);
			border-radius: 8px;
			color: var(--text-primary);
			cursor: pointer;
			transition: all 0.3s ease;
			font-weight: 600;
			font-size: 13px;
		}

		.language-toggle:hover {
			background: rgba(255, 0, 110, 0.1);
			border-color: var(--primary);
			color: var(--primary);
		}

		.profile-section {
			text-align: center;
			padding: 40px 0;
			background: linear-gradient(180deg, rgba(255, 0, 110, 0.05) 0%%, transparent 100%%);
			border-bottom: 1px solid var(--border-color);
		}

		.profile-image {
			width: 80px;
			height: 80px;
			border-radius: 50%%;
			margin: 0 auto 16px;
			border: 3px solid var(--primary);
			object-fit: cover;
			box-shadow: 0 0 20px rgba(255, 0, 110, 0.3);
		}

		.profile-name {
			font-size: 28px;
			font-weight: 700;
			margin-bottom: 8px;
			letter-spacing: -0.5px;
		}

		.profile-username {
			color: var(--text-secondary);
			font-size: 14px;
			margin-bottom: 12px;
		}

		.profile-tagline {
			color: var(--primary);
			font-size: 14px;
			font-weight: 600;
			margin-bottom: 12px;
		}

		.member-badge {
			display: inline-block;
			padding: 6px 14px;
			border: 2px solid var(--secondary);
			border-radius: 20px;
			color: var(--secondary);
			font-size: 12px;
			font-weight: 600;
			margin-bottom: 16px;
		}

		.rating-section {
			display: flex;
			justify-content: center;
			align-items: center;
			gap: 8px;
			margin-bottom: 20px;
		}

		.rating-stars {
			color: #FFD700;
			font-size: 18px;
		}

		.rating-value {
			font-weight: 600;
			color: var(--text-primary);
		}

		.stats-grid {
			display: grid;
			grid-template-columns: repeat(3, 1fr);
			gap: 16px;
			margin-top: 24px;
		}

		.stat-card {
			background: var(--bg-card);
			border: 1px solid var(--border-color);
			border-radius: 12px;
			padding: 16px;
			text-align: center;
		}

		.stat-value {
			font-size: 24px;
			font-weight: 700;
			color: var(--primary);
		}

		.stat-label {
			font-size: 12px;
			color: var(--text-secondary);
			margin-top: 8px;
		}

		.content-section {
			padding: 40px 0;
		}

		.section-title {
			font-size: 24px;
			font-weight: 700;
			margin-bottom: 24px;
			color: var(--text-primary);
		}

		.bio-section {
			background: var(--bg-card);
			border: 1px solid var(--border-color);
			border-radius: 12px;
			padding: 24px;
			margin-bottom: 40px;
		}

		.bio-text {
			color: var(--text-secondary);
			line-height: 1.8;
		}

		.offers-grid {
			display: grid;
			grid-template-columns: repeat(3, 1fr);
			gap: 20px;
		}

		.offer-card {
			background: var(--bg-card);
			border: 1px solid var(--border-color);
			border-radius: 12px;
			overflow: hidden;
			transition: all 0.3s ease;
		}

		.offer-card:hover {
			transform: translateY(-4px);
			border-color: var(--primary);
			box-shadow: 0 8px 24px rgba(255, 0, 110, 0.2);
		}

		.offer-image {
			width: 100%%;
			height: 160px;
			background-size: cover;
			background-position: center;
			position: relative;
		}

		.category-badge {
			position: absolute;
			top: 12px;
			right: 12px;
			background: rgba(0, 0, 0, 0.7);
			color: var(--text-primary);
			padding: 6px 12px;
			border-radius: 20px;
			font-size: 11px;
			font-weight: 600;
			backdrop-filter: blur(10px);
		}

		.offer-content {
			padding: 16px;
		}

		.offer-title {
			font-size: 16px;
			font-weight: 600;
			margin-bottom: 8px;
			color: var(--text-primary);
			line-height: 1.3;
		}

		.offer-description {
			font-size: 13px;
			color: var(--text-secondary);
			margin-bottom: 12px;
			line-height: 1.4;
			display: -webkit-box;
			-webkit-line-clamp: 2;
			-webkit-box-orient: vertical;
			overflow: hidden;
		}

		.offer-commission {
			background: linear-gradient(135deg, var(--accent), var(--dark-accent));
			color: var(--text-primary);
			padding: 8px 12px;
			border-radius: 8px;
			font-size: 12px;
			font-weight: 600;
			margin-bottom: 12px;
			display: inline-block;
		}

		.get-link-btn {
			width: 100%%;
			padding: 10px;
			background: linear-gradient(135deg, var(--accent), var(--dark-accent));
			color: var(--text-primary);
			border: none;
			border-radius: 8px;
			font-size: 14px;
			font-weight: 600;
			cursor: pointer;
			transition: all 0.3s ease;
			text-decoration: none;
			display: inline-block;
			text-align: center;
		}

		.get-link-btn:hover {
			transform: scale(1.02);
			box-shadow: 0 4px 12px rgba(142, 45, 226, 0.4);
		}

		footer {
			background: var(--bg-dark);
			border-top: 1px solid var(--border-color);
			padding: 40px 0;
			text-align: center;
			color: var(--text-secondary);
			font-size: 14px;
		}

		@media (max-width: 768px) {
			.offers-grid {
				grid-template-columns: repeat(2, 1fr);
			}

			.stats-grid {
				grid-template-columns: repeat(2, 1fr);
			}

			.profile-name {
				font-size: 24px;
			}
		}

		@media (max-width: 480px) {
			.offers-grid {
				grid-template-columns: 1fr;
			}

			.stats-grid {
				grid-template-columns: 1fr;
			}

			.profile-name {
				font-size: 20px;
			}

			.section-title {
				font-size: 20px;
			}
		}
	</style>
</head>
<body>
	<header>
		<div class="container">
			<div class="header-content">
				<button class="language-toggle" onclick="toggleLanguage()">
					<span id="lang-text">English</span>
				</button>
				<div class="logo">AffTok</div>
				<div style="width: 100px;"></div>
			</div>
		</div>
	</header>

	<div class="profile-section">
		<div class="container">
			<div style="width: 80px; height: 80px; border-radius: 50%%; margin: 0 auto 16px; background: linear-gradient(135deg, var(--accent), var(--dark-accent)); display: flex; align-items: center; justify-content: center; font-size: 36px; border: 3px solid var(--primary);">%s</div>
			<h1 class="profile-name">%s</h1>
			<p class="profile-username">@%s</p>
			<div class="rating-section">
				<span class="rating-stars">★★★★★</span>
				<span class="rating-value">%.1f</span>
			</div>
			<div class="stats-grid">
				<div class="stat-card">
					<div class="stat-value">%d</div>
					<div class="stat-label">Offers</div>
				</div>
				<div class="stat-card">
					<div class="stat-value">%d</div>
					<div class="stat-label">Clicks</div>
				</div>
				<div class="stat-card">
					<div class="stat-value">Gold</div>
					<div class="stat-label">Member</div>
				</div>
			</div>
		</div>
	</div>

	<div class="content-section">
		<div class="container">
			%s
			<h2 class="section-title">Available Offers</h2>
			<div class="offers-grid">
				%s
			</div>
		</div>
	</div>

	<footer>
		<div class="container">
			<p>© 2024 AffTok. All rights reserved.</p>
		</div>
	</footer>

	<script>
		let currentLang = 'ar';

		function toggleLanguage() {
			currentLang = currentLang === 'ar' ? 'en' : 'ar';
			const html = document.documentElement;
			const body = document.body;
			const langBtn = document.getElementById('lang-text');

			if (currentLang === 'ar') {
				html.setAttribute('lang', 'ar');
				html.setAttribute('dir', 'rtl');
				body.classList.remove('en');
				langBtn.textContent = 'English';
			} else {
				html.setAttribute('lang', 'en');
				html.setAttribute('dir', 'ltr');
				body.classList.add('en');
				langBtn.textContent = 'العربية';
			}
		}
	</script>
</body>
</html>
		`, fullName, avatarLetter, user.Username,
		promoterRating, totalOffers, totalClicks, bioHTML, offersHTML)
}

func getCategoryBadge(category string) string {
	switch category {
	case "Finance":
		return "badge-finance"
	case "E-commerce":
		return "badge-ecommerce"
	case "Crypto":
		return "badge-crypto"
	case "Travel":
		return "badge-travel"
	default:
		return "badge-default"
	}
}

func (h *PromoterHandler) RatePromoter(c *gin.Context) {
	var req struct {
		PromoterID string `json:"promoter_id"`
		Rating     int    `json:"rating"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Rating < 1 || req.Rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Rating must be between 1 and 5"})
		return
	}

	promoterID, err := uuid.Parse(req.PromoterID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid promoter ID"})
		return
	}

	visitorIP := c.ClientIP()

	var existingRating models.PromoterRating
	result := h.db.Where("promoter_id = ? AND visitor_ip = ?", promoterID, visitorIP).First(&existingRating)

	if result.Error == nil {
		existingRating.Rating = req.Rating
		h.db.Save(&existingRating)
	} else {
		newRating := models.PromoterRating{
			PromoterID: promoterID,
			VisitorIP:  visitorIP,
			Rating:     req.Rating,
		}
		h.db.Create(&newRating)
	}

	var avgRating float64
	h.db.Model(&models.PromoterRating{}).
		Where("promoter_id = ?", promoterID).
		Select("COALESCE(AVG(rating), 0)").
		Scan(&avgRating)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"average_rating": avgRating,
	})
}

func (h *PromoterHandler) GetPromoterRating(promoterID uuid.UUID) float64 {
	var avgRating float64
	h.db.Model(&models.PromoterRating{}).
		Where("promoter_id = ?", promoterID).
		Select("COALESCE(AVG(rating), 4.5)").
		Scan(&avgRating)
	return avgRating
}

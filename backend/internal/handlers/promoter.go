package handlers

import (
	"fmt"
	"net/http"

	"github.com/afftok/backend/internal/models"
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

	var offers []models.Offer
	if err := h.db.Where("status = ?", "active").Order("created_at DESC").Find(&offers).Error; err != nil {
		offers = []models.Offer{}
	}

	var totalClicks int64
	var totalOffers int64

	h.db.Model(&models.Click{}).
		Joins("JOIN user_offers ON clicks.user_offer_id = user_offers.id").
		Where("user_offers.user_id = ? AND user_offers.status = ?", id, "active").
		Count(&totalClicks)

	h.db.Model(&models.UserOffer{}).
		Where("user_id = ? AND status = ?", id, "active").
		Count(&totalOffers)

	html := h.generateHTML(user, offers, int(totalOffers), int(totalClicks))
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
}

func (h *PromoterHandler) generateHTML(user models.AfftokUser, offers []models.Offer, totalOffers int, totalClicks int) string {

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
					<span class="btn-text" data-en="Get Link" data-ar="احصل على الرابط">Get Link</span>
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
			<p class="bio-text" data-en="Welcome to my page! I share the best offers and deals to help you save money and earn rewards. Check out my curated selection below!" data-ar="مرحباً بك في صفحتي! أشارك أفضل العروض والصفقات لمساعدتك على توفير المال وكسب المكافآت. تفقد مجموعتي المختارة أدناه!">%s</p>
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
<html lang="en" dir="ltr">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>%s - AffTok Promoter</title>
	<style>
		* {
			margin: 0;
			padding: 0;
			box-sizing: border-box;
		}

		body {
			font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
			background: #000000;
			color: #ffffff;
			min-height: 100vh;
			display: flex;
			flex-direction: column;
			box-shadow: 0 0 0 8px #8E2DE2, 0 0 0 16px #4A00E0;
		}

		.top-bar {
			background: #000000;
			padding: 15px 20px;
			display: flex;
			justify-content: space-between;
			align-items: center;
			border-bottom: 1px solid #2d2d2d;
		}

		.logo-container {
			display: flex;
			align-items: center;
			gap: 12px;
		}

		.logo {
			width: 40px;
			height: 40px;
		}

		.logo-text {
			font-size: 24px;
			font-weight: bold;
			color: #FF0000;
		}

		.lang-toggle {
			background: #1a1a1a;
			border: 1px solid #2d2d2d;
			color: #ffffff;
			padding: 8px 16px;
			border-radius: 20px;
			cursor: pointer;
			font-size: 14px;
			font-weight: 600;
			transition: all 0.3s ease;
		}

		.lang-toggle:hover {
			background: linear-gradient(135deg, #8E2DE2 0%%, #4A00E0 100%%);
			border-color: transparent;
		}

		.header {
			background: linear-gradient(135deg, #8E2DE2 0%%, #4A00E0 100%%);
			padding: 60px 20px 40px;
			text-align: center;
		}

		.avatar {
			width: 100px;
			height: 100px;
			border-radius: 50%%;
			background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
			display: flex;
			align-items: center;
			justify-content: center;
			font-size: 48px;
			margin: 0 auto 20px;
			border: 4px solid rgba(255, 255, 255, 0.2);
		}

		.username {
			font-size: 32px;
			font-weight: bold;
			margin-bottom: 10px;
		}

		.fullname {
			font-size: 18px;
			opacity: 0.9;
			margin-bottom: 15px;
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 10px;
		}

		.rating-modal {
			position: fixed;
			top: 0;
			left: 0;
			width: 100%%;
			height: 100%%;
			background: rgba(0, 0, 0, 0.8);
			display: none;
			align-items: center;
			justify-content: center;
			z-index: 2000;
		}

		.rating-modal.show {
			display: flex;
		}

		.rating-box {
			background: linear-gradient(135deg, #8E2DE2 0%%, #4A00E0 100%%);
			padding: 40px;
			border-radius: 20px;
			text-align: center;
			box-shadow: 0 20px 60px rgba(142, 45, 226, 0.6);
		}

		.rating-box h3 {
			margin-bottom: 30px;
			font-size: 24px;
		}

		.rating-stars {
			display: flex;
			gap: 15px;
			justify-content: center;
			margin-bottom: 20px;
		}

		.rating-stars span {
			font-size: 48px;
			cursor: pointer;
			color: #4A00E0;
			transition: all 0.2s ease;
		}

		.rating-stars span:hover,
		.rating-stars span.active {
			color: #FFC700;
			transform: scale(1.2);
		}

		.close-modal {
			background: rgba(255, 255, 255, 0.2);
			border: none;
			color: white;
			padding: 10px 30px;
			border-radius: 10px;
			cursor: pointer;
			font-size: 16px;
			font-weight: 600;
		}

		.rate-prompt {
			font-size: 14px;
			opacity: 0.8;
			margin-bottom: 30px;
			cursor: pointer;
			text-decoration: underline;
		}

		.stats {
			display: flex;
			justify-content: center;
			gap: 40px;
			flex-wrap: wrap;
		}

		.stat-item {
			text-align: center;
		}

		.stat-value {
			font-size: 28px;
			font-weight: bold;
		}

		.stat-label {
			font-size: 14px;
			opacity: 0.8;
			margin-top: 5px;
		}

		.container {
			max-width: 1200px;
			margin: 0 auto;
			padding: 40px 20px;
			flex: 1;
		}

		.bio-section {
			background: #1a1a1a;
			border-radius: 16px;
			padding: 30px;
			margin-bottom: 40px;
			border: 1px solid #2d2d2d;
		}

		.bio-section h2 {
			font-size: 24px;
			margin-bottom: 15px;
			color: #ffffff;
		}

		.bio-section p {
			font-size: 16px;
			line-height: 1.6;
			color: #cccccc;
		}

		.section-title {
			font-size: 28px;
			margin-bottom: 30px;
			text-align: center;
		}

		.offers-grid {
			display: grid;
			grid-template-columns: repeat(3, 1fr);
			gap: 24px;
		}

		.offer-card {
			background: #1a1a1a;
			border-radius: 16px;
			overflow: hidden;
			border: 1px solid #2d2d2d;
			transition: transform 0.3s ease, box-shadow 0.3s ease;
		}

		.offer-card:hover {
			transform: translateY(-5px);
			box-shadow: 0 10px 30px rgba(142, 45, 226, 0.3);
		}

		.offer-image {
			height: 180px;
			background-size: cover;
			background-position: center;
			position: relative;
		}

		.category-badge {
			position: absolute;
			top: 12px;
			right: 12px;
			padding: 6px 12px;
			border-radius: 20px;
			font-size: 12px;
			font-weight: 600;
			backdrop-filter: blur(10px);
		}

		.badge-finance { background: rgba(34, 197, 94, 0.9); }
		.badge-ecommerce { background: rgba(59, 130, 246, 0.9); }
		.badge-crypto { background: rgba(251, 146, 60, 0.9); }
		.badge-travel { background: rgba(168, 85, 247, 0.9); }
		.badge-default { background: rgba(107, 114, 128, 0.9); }

		.offer-content {
			padding: 20px;
		}

		.offer-title {
			font-size: 18px;
			font-weight: 600;
			margin-bottom: 10px;
			line-height: 1.3;
			min-height: 48px;
			display: flex;
			align-items: center;
		}

		.offer-description {
			font-size: 14px;
			color: #999999;
			margin-bottom: 15px;
			line-height: 1.5;
			display: -webkit-box;
			-webkit-line-clamp: 2;
			-webkit-box-orient: vertical;
			overflow: hidden;
		}

		.offer-meta {
			display: flex;
			justify-content: space-between;
			align-items: center;
			margin-bottom: 15px;
		}

		.offer-commission {
			background: linear-gradient(135deg, #8E2DE2 0%%, #4A00E0 100%%);
			padding: 6px 12px;
			border-radius: 8px;
			font-size: 14px;
			font-weight: 600;
		}

		.get-link-btn {
			width: 100%%;
			padding: 12px;
			background: linear-gradient(135deg, #8E2DE2 0%%, #4A00E0 100%%);
			border: none;
			border-radius: 10px;
			color: white;
			font-size: 16px;
			font-weight: 600;
			cursor: pointer;
			display: flex;
			align-items: center;
			justify-content: center;
			gap: 8px;
			transition: transform 0.2s ease;
			text-decoration: none;
		}

		.get-link-btn:hover {
			transform: scale(1.02);
		}

		.get-link-btn:active {
			transform: scale(0.98);
		}

		.btn-icon {
			font-size: 18px;
		}

		.toast {
			position: fixed;
			bottom: 30px;
			left: 50%%;
			transform: translateX(-50%%) translateY(100px);
			background: linear-gradient(135deg, #8E2DE2 0%%, #4A00E0 100%%);
			color: white;
			padding: 16px 32px;
			border-radius: 50px;
			font-weight: 600;
			box-shadow: 0 10px 40px rgba(142, 45, 226, 0.5);
			transition: transform 0.3s ease;
			z-index: 1000;
		}

		.toast.show {
			transform: translateX(-50%%) translateY(0);
		}

		.footer {
			background: #0a0a0a;
			border-top: 1px solid #2d2d2d;
			padding: 40px 20px;
			margin-top: 60px;
		}

		.footer-content {
			max-width: 1200px;
			margin: 0 auto;
			text-align: center;
		}

		.footer-logo {
			width: 60px;
			height: 60px;
			margin: 0 auto 20px;
		}

		.footer-title {
			font-size: 24px;
			font-weight: bold;
			color: #FF0000;
			margin-bottom: 20px;
		}

		.footer-text {
			font-size: 14px;
			color: #999999;
			line-height: 1.6;
		}

		.copyright {
			margin-top: 30px;
			padding-top: 30px;
			border-top: 1px solid #2d2d2d;
			font-size: 12px;
			color: #666666;
		}

		@media (max-width: 768px) {
			.offers-grid {
				grid-template-columns: repeat(2, 1fr);
			}
		}

		@media (max-width: 480px) {
			.offers-grid {
				grid-template-columns: 1fr;
			}
			.username {
				font-size: 24px;
			}
		}
	</style>
</head>
<body>
	<div class="top-bar">
		<div class="logo-container">
			<div class="logo-text">🎬</div>
			<span class="logo-text">AffTok</span>
		</div>
		<button class="lang-toggle" id="lang-btn" onclick="toggleLanguage()">
			<span id="lang-btn-text">العربية</span>
		</button>
	</div>

	<div class="header">
		<div class="avatar">%s</div>
		<div class="username">%s</div>
		<div class="fullname">
			<span>%s</span>
			<span>⭐ %.1f</span>
		</div>
		<div class="stats">
			<div class="stat-item">
				<div class="stat-value">%d</div>
				<div class="stat-label" data-en="Offers" data-ar="العروض">Offers</div>
			</div>
			<div class="stat-item">
				<div class="stat-value">%d</div>
				<div class="stat-label" data-en="Clicks" data-ar="النقرات">Clicks</div>
			</div>
		</div>
		<div class="rate-prompt" onclick="showRatingModal()" data-en="Rate this promoter" data-ar="قيّم هذا المروج">Rate this promoter</div>
	</div>

	<div class="container">
		%s
		<div class="section-title" data-en="Available Offers" data-ar="العروض المتاحة">Available Offers</div>
		<div class="offers-grid">
			%s
		</div>
	</div>

	<div class="footer">
		<div class="footer-content">
			<div class="footer-title">🎬 AffTok</div>
			<p class="footer-text" data-en="The ultimate affiliate marketing platform" data-ar="منصة التسويق بالعمولة الشاملة">The ultimate affiliate marketing platform</p>
			<div class="copyright en" style="display: block;">© 2024 AffTok. All rights reserved.</div>
			<div class="copyright ar" style="display: none;">© 2024 AffTok. جميع الحقوق محفوظة.</div>
		</div>
	</div>

	<div class="rating-modal" id="ratingModal">
		<div class="rating-box">
			<h3 data-en="Rate this promoter" data-ar="قيّم هذا المروج">Rate this promoter</h3>
			<div class="rating-stars">
				<span data-rating="1">⭐</span>
				<span data-rating="2">⭐</span>
				<span data-rating="3">⭐</span>
				<span data-rating="4">⭐</span>
				<span data-rating="5">⭐</span>
			</div>
			<button class="close-modal" onclick="closeRatingModal()" data-en="Close" data-ar="إغلاق">Close</button>
		</div>
	</div>

	<div class="toast" id="toast"></div>

	<script>
		let currentLang = 'en';
		const promoterID = '%s';

		function toggleLanguage() {
			currentLang = currentLang === 'en' ? 'ar' : 'en';
			const html = document.documentElement;

			if (currentLang === 'ar') {
				html.setAttribute('lang', 'ar');
				html.setAttribute('dir', 'rtl');
				document.getElementById('lang-btn-text').textContent = 'English';
			} else {
				html.setAttribute('lang', 'en');
				html.setAttribute('dir', 'ltr');
				document.getElementById('lang-btn-text').textContent = 'العربية';
			}

			document.querySelectorAll('[data-en][data-ar]').forEach(el => {
				el.textContent = el.getAttribute('data-' + currentLang);
			});

			const copyrights = document.querySelectorAll('.copyright');
			copyrights.forEach(cr => {
				if (cr.classList.contains('ar')) {
					cr.style.display = currentLang === 'ar' ? 'block' : 'none';
				} else {
					cr.style.display = currentLang === 'en' ? 'block' : 'none';
				}
			});
		}

		function showRatingModal() {
			document.getElementById('ratingModal').classList.add('show');
		}

		function closeRatingModal() {
			document.getElementById('ratingModal').classList.remove('show');
		}

		document.querySelectorAll('.rating-stars span').forEach(star => {
			star.addEventListener('click', function() {
				const rating = parseInt(this.getAttribute('data-rating'));
				ratePromoter(rating);
				closeRatingModal();
			});
			star.addEventListener('mouseenter', function() {
				const rating = parseInt(this.getAttribute('data-rating'));
				document.querySelectorAll('.rating-stars span').forEach((s, i) => {
					if (i < rating) s.classList.add('active');
					else s.classList.remove('active');
				});
			});
		});

		function ratePromoter(rating) {
			fetch('/api/rate-promoter', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				body: JSON.stringify({ promoter_id: promoterID, rating: rating })
			})
			.then(res => res.json())
			.then(data => {
				if (data.success) {
					showToast(currentLang === 'ar' ? 'شكراً لتقييمك!' : 'Thank you for rating!');
				}
			})
			.catch(err => {
				console.error('Rating failed:', err);
			});
		}

		function showToast(message) {
			const toast = document.getElementById('toast');
			toast.textContent = message || 'Link copied to clipboard! 🎉';
			toast.classList.add('show');
			setTimeout(() => {
				toast.classList.remove('show');
			}, 3000);
		}
	</script>
</body>
</html>
		`, fullName, avatarLetter, user.Username,
			promoterRating, totalOffers, totalClicks, bioHTML, offersHTML, promoterID)
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

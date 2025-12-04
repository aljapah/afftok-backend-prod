package handlers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

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
	// Read the template file from public folder
	filePath := "public/promoter_landing.html"
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		// Try alternate path
		filePath = "./public/promoter_landing.html"
		data, err = ioutil.ReadFile(filePath)
		if err != nil {
			// Return error page if template not found
			return fmt.Sprintf(`<!DOCTYPE html>
<html><head><title>Error</title></head>
<body style="background:#0a0a0a;color:#fff;font-family:sans-serif;text-align:center;padding:50px;">
<h1>Template not found</h1>
<p>Path: %s</p>
<p>Error: %s</p>
</body></html>`, filePath, err.Error())
		}
	}
	
	html := string(data)
	
	// Build offers JSON
	offersJSON := "["
	for i, offer := range offers {
		if i > 0 {
			offersJSON += ","
		}
		// Escape special characters in strings
		title := strings.ReplaceAll(offer.Title, `"`, `\"`)
		title = strings.ReplaceAll(title, "\n", " ")
		desc := strings.ReplaceAll(offer.Description, `"`, `\"`)
		desc = strings.ReplaceAll(desc, "\n", " ")
		
		offersJSON += fmt.Sprintf(`{"id":"%s","title":"%s","description":"%s","image_url":"%s","category":"%s","payout":%d,"payout_type":"%s"}`,
			offer.ID.String(),
			title,
			desc,
			offer.ImageURL,
			offer.Category,
			offer.Payout,
			offer.PayoutType)
	}
	offersJSON += "]"
	
	// Get promoter rating
	promoterRating := h.GetPromoterRating(user.ID)
	
	// Escape user data
	fullName := strings.ReplaceAll(user.FullName, `"`, `\"`)
	if fullName == "" {
		fullName = user.Username
	}
	bio := strings.ReplaceAll(user.Bio, `"`, `\"`)
	bio = strings.ReplaceAll(bio, "\n", " ")
	
	// Inject promoter data into the page
	dataScript := fmt.Sprintf(`<script>
		window.promoterId = "%s";
		window.promoterData = {
			id: "%s",
			name: "%s",
			username: "%s",
			bio: "%s",
			avatar: "%s",
			rating: %.1f,
			totalClicks: %d,
			totalOffers: %d
		};
		window.offersData = %s;
	</script>`,
		user.ID.String(),
		user.ID.String(),
		fullName,
		user.Username,
		bio,
		user.AvatarURL,
		promoterRating,
		totalClicks,
		totalOffers,
		offersJSON)
	
	// Insert the data script before </head>
	html = strings.Replace(html, "</head>", dataScript+"</head>", 1)
	return html
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

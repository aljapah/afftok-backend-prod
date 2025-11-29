package services

import (
	"strings"

	"github.com/afftok/backend/internal/database"
	"github.com/afftok/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ClickService struct{}

func NewClickService() *ClickService {
	return &ClickService{}
}

// TrackClick records a click on an affiliate link
func (s *ClickService) TrackClick(c *gin.Context, userOfferID uuid.UUID) (*models.Click, error) {
	// Extract device info from user agent
	userAgent := c.Request.UserAgent()
	device, browser, os := parseUserAgent(userAgent)

	// Get IP address
	ipAddress := c.ClientIP()

	// Get referrer
	referrer := c.Request.Referer()

	// Create click record
	click := models.Click{
		ID:          uuid.New(),
		UserOfferID: userOfferID,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		Device:      device,
		Browser:     browser,
		OS:          os,
		Referrer:    referrer,
	}

	if err := database.DB.Create(&click).Error; err != nil {
		return nil, err
	}

	// Update user offer clicks count
	database.DB.Model(&models.UserOffer{}).Where("id = ?", userOfferID).UpdateColumn("clicks", database.DB.Raw("clicks + 1"))

	// Update offer total clicks
	var userOffer models.UserOffer
	if err := database.DB.First(&userOffer, "id = ?", userOfferID).Error; err == nil {
		database.DB.Model(&models.Offer{}).Where("id = ?", userOffer.OfferID).UpdateColumn("total_clicks", database.DB.Raw("total_clicks + 1"))
		
		// Update user total clicks
		database.DB.Model(&models.AfftokUser{}).Where("id = ?", userOffer.UserID).UpdateColumn("total_clicks", database.DB.Raw("total_clicks + 1"))
	}

	return &click, nil
}

// parseUserAgent extracts device, browser, and OS from user agent string
func parseUserAgent(ua string) (device, browser, os string) {
	ua = strings.ToLower(ua)

	// Detect device
	if strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone") {
		device = "mobile"
	} else if strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad") {
		device = "tablet"
	} else {
		device = "desktop"
	}

	// Detect browser
	switch {
	case strings.Contains(ua, "edg"):
		browser = "Edge"
	case strings.Contains(ua, "chrome"):
		browser = "Chrome"
	case strings.Contains(ua, "safari"):
		browser = "Safari"
	case strings.Contains(ua, "firefox"):
		browser = "Firefox"
	case strings.Contains(ua, "opera"):
		browser = "Opera"
	default:
		browser = "Other"
	}

	// Detect OS
	switch {
	case strings.Contains(ua, "windows"):
		os = "Windows"
	case strings.Contains(ua, "mac"):
		os = "macOS"
	case strings.Contains(ua, "linux"):
		os = "Linux"
	case strings.Contains(ua, "android"):
		os = "Android"
	case strings.Contains(ua, "ios") || strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad"):
		os = "iOS"
	default:
		os = "Other"
	}

	return
}

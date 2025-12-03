package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/database"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClickService struct {
	linkService *LinkService
}

func NewClickService() *ClickService {
	return &ClickService{
		linkService: NewLinkService(),
	}
}

// TrackClick records a click on an affiliate link with atomic operations
func (s *ClickService) TrackClick(c *gin.Context, userOfferID uuid.UUID) (*models.Click, error) {
	// Extract device info from user agent
	userAgent := c.Request.UserAgent()
	device, browser, os := parseUserAgent(userAgent)

	// Get IP address
	ipAddress := c.ClientIP()

	// Get referrer
	referrer := c.Request.Referer()

	// Generate click fingerprint for deduplication
	clickID := s.linkService.GenerateClickID(userOfferID, ipAddress, userAgent)

	// Check for duplicate click (within time window)
	if s.linkService.IsClickDuplicate(clickID) {
		// Return existing click without creating new one
		var existingClick models.Click
		if err := database.DB.Where("user_offer_id = ? AND ip_address = ?", userOfferID, ipAddress).
			Order("clicked_at DESC").
			First(&existingClick).Error; err == nil {
			return &existingClick, nil
		}
		// If not found in DB, still proceed (Redis might be stale)
	}

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
		ClickedAt:   time.Now().UTC(),
	}

	// Use transaction for atomic updates
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Create click record
		if err := tx.Create(&click).Error; err != nil {
			return fmt.Errorf("failed to create click: %w", err)
		}

		// 2. Get user offer to find related IDs
		var userOffer models.UserOffer
		if err := tx.First(&userOffer, "id = ?", userOfferID).Error; err != nil {
			return fmt.Errorf("user offer not found: %w", err)
		}

		// 3. Atomic increment on UserOffer (clicks counter + updated_at)
		if err := tx.Model(&models.UserOffer{}).
			Where("id = ?", userOfferID).
			UpdateColumns(map[string]interface{}{
				"total_clicks": gorm.Expr("total_clicks + 1"),
				"updated_at":   time.Now().UTC(),
			}).Error; err != nil {
			return fmt.Errorf("failed to update user offer: %w", err)
		}

		// 4. Atomic increment on Offer
		if err := tx.Model(&models.Offer{}).
			Where("id = ?", userOffer.OfferID).
			UpdateColumn("total_clicks", gorm.Expr("total_clicks + 1")).Error; err != nil {
			return fmt.Errorf("failed to update offer clicks: %w", err)
		}

		// 5. Atomic increment on User
		if err := tx.Model(&models.AfftokUser{}).
			Where("id = ?", userOffer.UserID).
			UpdateColumn("total_clicks", gorm.Expr("total_clicks + 1")).Error; err != nil {
			return fmt.Errorf("failed to update user clicks: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Update Redis counters asynchronously (non-blocking)
	go s.updateRedisCounters(userOfferID, click.ID)

	return &click, nil
}

// updateRedisCounters updates click counters in Redis for fast reads
func (s *ClickService) updateRedisCounters(userOfferID uuid.UUID, clickID uuid.UUID) {
	if cache.RedisClient == nil {
		return
	}

	ctx := context.Background()
	today := time.Now().UTC().Format("2006-01-02")

	// Keys for different time windows
	keys := []string{
		fmt.Sprintf("clicks:total:%s", userOfferID.String()),
		fmt.Sprintf("clicks:daily:%s:%s", userOfferID.String(), today),
		fmt.Sprintf("clicks:hourly:%s:%s:%d", userOfferID.String(), today, time.Now().Hour()),
	}

	for _, key := range keys {
		if _, err := cache.Increment(ctx, key); err != nil {
			fmt.Printf("[ClickService] Redis increment error for %s: %v\n", key, err)
		}
	}

	// Set expiration for daily/hourly keys
	cache.RedisClient.Expire(ctx, keys[1], 48*time.Hour)  // Daily: 48 hours
	cache.RedisClient.Expire(ctx, keys[2], 2*time.Hour)   // Hourly: 2 hours
}

// GetClickStats returns click statistics for a user offer
func (s *ClickService) GetClickStats(userOfferID uuid.UUID) (map[string]interface{}, error) {
	stats := make(map[string]interface{})
	ctx := context.Background()

	// Try Redis first for fast read
	if cache.RedisClient != nil {
		totalKey := fmt.Sprintf("clicks:total:%s", userOfferID.String())
		if total, err := cache.Get(ctx, totalKey); err == nil {
			stats["total_from_cache"] = total
		}
	}

	// Get detailed stats from database
	var clicks []models.Click
	if err := database.DB.Where("user_offer_id = ?", userOfferID).
		Order("clicked_at DESC").
		Limit(100).
		Find(&clicks).Error; err != nil {
		return nil, err
	}

	// Aggregate stats
	deviceStats := make(map[string]int)
	browserStats := make(map[string]int)
	osStats := make(map[string]int)
	countryStats := make(map[string]int)
	hourlyStats := make(map[int]int)

	for _, click := range clicks {
		deviceStats[click.Device]++
		browserStats[click.Browser]++
		osStats[click.OS]++
		if click.Country != "" {
			countryStats[click.Country]++
		}
		hourlyStats[click.ClickedAt.Hour()]++
	}

	stats["total_clicks"] = len(clicks)
	stats["device_breakdown"] = deviceStats
	stats["browser_breakdown"] = browserStats
	stats["os_breakdown"] = osStats
	stats["country_breakdown"] = countryStats
	stats["hourly_distribution"] = hourlyStats
	stats["recent_clicks"] = clicks

	return stats, nil
}

// GetUserTotalClicks returns total clicks for a user across all offers
func (s *ClickService) GetUserTotalClicks(userID uuid.UUID) (int64, error) {
	var total int64

	// Get all user offers
	var userOfferIDs []uuid.UUID
	if err := database.DB.Model(&models.UserOffer{}).
		Where("user_id = ?", userID).
		Pluck("id", &userOfferIDs).Error; err != nil {
		return 0, err
	}

	if len(userOfferIDs) == 0 {
		return 0, nil
	}

	// Count clicks
	if err := database.DB.Model(&models.Click{}).
		Where("user_offer_id IN ?", userOfferIDs).
		Count(&total).Error; err != nil {
		return 0, err
	}

	return total, nil
}

// parseUserAgent extracts device, browser, and OS from user agent string
func parseUserAgent(ua string) (device, browser, os string) {
	ua = strings.ToLower(ua)

	// Detect device
	switch {
	case strings.Contains(ua, "mobile") || strings.Contains(ua, "android") || strings.Contains(ua, "iphone"):
		device = "mobile"
	case strings.Contains(ua, "tablet") || strings.Contains(ua, "ipad"):
		device = "tablet"
	default:
		device = "desktop"
	}

	// Detect browser (order matters - check specific first)
	switch {
	case strings.Contains(ua, "edg"):
		browser = "Edge"
	case strings.Contains(ua, "opr") || strings.Contains(ua, "opera"):
		browser = "Opera"
	case strings.Contains(ua, "firefox"):
		browser = "Firefox"
	case strings.Contains(ua, "chrome") && !strings.Contains(ua, "edg"):
		browser = "Chrome"
	case strings.Contains(ua, "safari") && !strings.Contains(ua, "chrome"):
		browser = "Safari"
	default:
		browser = "Other"
	}

	// Detect OS
	switch {
	case strings.Contains(ua, "windows"):
		os = "Windows"
	case strings.Contains(ua, "mac") && !strings.Contains(ua, "iphone") && !strings.Contains(ua, "ipad"):
		os = "macOS"
	case strings.Contains(ua, "linux") && !strings.Contains(ua, "android"):
		os = "Linux"
	case strings.Contains(ua, "android"):
		os = "Android"
	case strings.Contains(ua, "iphone") || strings.Contains(ua, "ipad") || strings.Contains(ua, "ios"):
		os = "iOS"
	default:
		os = "Other"
	}

	return
}

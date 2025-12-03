package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/database"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
)

// AnalyticsService handles analytics aggregation and caching
type AnalyticsService struct{}

func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{}
}

// UserStats represents aggregated user statistics
type UserStats struct {
	TotalClicks      int64   `json:"total_clicks"`
	TotalConversions int64   `json:"total_conversions"`
	TotalEarnings    int64   `json:"total_earnings"`
	ConversionRate   float64 `json:"conversion_rate"`
	ActiveOffers     int64   `json:"active_offers"`
	
	// Time-based stats
	ClicksToday      int64 `json:"clicks_today"`
	ClicksThisWeek   int64 `json:"clicks_this_week"`
	ClicksThisMonth  int64 `json:"clicks_this_month"`
	
	ConversionsToday     int64 `json:"conversions_today"`
	ConversionsThisWeek  int64 `json:"conversions_this_week"`
	ConversionsThisMonth int64 `json:"conversions_this_month"`
}

// OfferStats represents aggregated offer statistics
type OfferStats struct {
	TotalClicks      int64   `json:"total_clicks"`
	TotalConversions int64   `json:"total_conversions"`
	UniqueClicks     int64   `json:"unique_clicks"`
	ConversionRate   float64 `json:"conversion_rate"`
	TotalUsers       int64   `json:"total_users"`
	TotalEarnings    int64   `json:"total_earnings"`
}

// GetUserStats returns aggregated stats for a user
func (s *AnalyticsService) GetUserStats(userID uuid.UUID) (*UserStats, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("user_stats:%s", userID.String())

	// Try cache first
	if cache.RedisClient != nil {
		if cached, err := cache.Get(ctx, cacheKey); err == nil {
			var stats UserStats
			if err := json.Unmarshal([]byte(cached), &stats); err == nil {
				return &stats, nil
			}
		}
	}

	// Calculate from database
	stats := &UserStats{}

	// Get user offer IDs
	var userOfferIDs []uuid.UUID
	if err := database.DB.Model(&models.UserOffer{}).
		Where("user_id = ?", userID).
		Pluck("id", &userOfferIDs).Error; err != nil {
		return nil, err
	}

	// Active offers count
	stats.ActiveOffers = int64(len(userOfferIDs))

	if len(userOfferIDs) == 0 {
		return stats, nil
	}

	// Total clicks
	database.DB.Model(&models.Click{}).
		Where("user_offer_id IN ?", userOfferIDs).
		Count(&stats.TotalClicks)

	// Total conversions
	database.DB.Model(&models.Conversion{}).
		Where("user_offer_id IN ?", userOfferIDs).
		Count(&stats.TotalConversions)

	// Total earnings (only approved conversions)
	var earnings struct {
		Total int64
	}
	database.DB.Model(&models.Conversion{}).
		Select("COALESCE(SUM(commission), 0) as total").
		Where("user_offer_id IN ? AND status = ?", userOfferIDs, models.ConversionStatusApproved).
		Scan(&earnings)
	stats.TotalEarnings = earnings.Total

	// Conversion rate
	if stats.TotalClicks > 0 {
		stats.ConversionRate = float64(stats.TotalConversions) / float64(stats.TotalClicks) * 100
	}

	// Time-based stats
	now := time.Now().UTC()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	startOfWeek := startOfDay.AddDate(0, 0, -int(now.Weekday()))
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)

	// Clicks today
	database.DB.Model(&models.Click{}).
		Where("user_offer_id IN ? AND clicked_at >= ?", userOfferIDs, startOfDay).
		Count(&stats.ClicksToday)

	// Clicks this week
	database.DB.Model(&models.Click{}).
		Where("user_offer_id IN ? AND clicked_at >= ?", userOfferIDs, startOfWeek).
		Count(&stats.ClicksThisWeek)

	// Clicks this month
	database.DB.Model(&models.Click{}).
		Where("user_offer_id IN ? AND clicked_at >= ?", userOfferIDs, startOfMonth).
		Count(&stats.ClicksThisMonth)

	// Conversions today
	database.DB.Model(&models.Conversion{}).
		Where("user_offer_id IN ? AND converted_at >= ?", userOfferIDs, startOfDay).
		Count(&stats.ConversionsToday)

	// Conversions this week
	database.DB.Model(&models.Conversion{}).
		Where("user_offer_id IN ? AND converted_at >= ?", userOfferIDs, startOfWeek).
		Count(&stats.ConversionsThisWeek)

	// Conversions this month
	database.DB.Model(&models.Conversion{}).
		Where("user_offer_id IN ? AND converted_at >= ?", userOfferIDs, startOfMonth).
		Count(&stats.ConversionsThisMonth)

	// Cache the result
	if cache.RedisClient != nil {
		if data, err := json.Marshal(stats); err == nil {
			cache.Set(ctx, cacheKey, string(data), 5*time.Minute)
		}
	}

	return stats, nil
}

// GetOfferStats returns aggregated stats for an offer
func (s *AnalyticsService) GetOfferStats(offerID uuid.UUID) (*OfferStats, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf("offer_stats:%s", offerID.String())

	// Try cache first
	if cache.RedisClient != nil {
		if cached, err := cache.Get(ctx, cacheKey); err == nil {
			var stats OfferStats
			if err := json.Unmarshal([]byte(cached), &stats); err == nil {
				return &stats, nil
			}
		}
	}

	stats := &OfferStats{}

	// Get user offer IDs for this offer
	var userOfferIDs []uuid.UUID
	if err := database.DB.Model(&models.UserOffer{}).
		Where("offer_id = ?", offerID).
		Pluck("id", &userOfferIDs).Error; err != nil {
		return nil, err
	}

	stats.TotalUsers = int64(len(userOfferIDs))

	if len(userOfferIDs) == 0 {
		return stats, nil
	}

	// Total clicks
	database.DB.Model(&models.Click{}).
		Where("user_offer_id IN ?", userOfferIDs).
		Count(&stats.TotalClicks)

	// Unique clicks (by IP)
	database.DB.Model(&models.Click{}).
		Where("user_offer_id IN ?", userOfferIDs).
		Distinct("ip_address").
		Count(&stats.UniqueClicks)

	// Total conversions
	database.DB.Model(&models.Conversion{}).
		Where("user_offer_id IN ?", userOfferIDs).
		Count(&stats.TotalConversions)

	// Total earnings
	var earnings struct {
		Total int64
	}
	database.DB.Model(&models.Conversion{}).
		Select("COALESCE(SUM(commission), 0) as total").
		Where("user_offer_id IN ? AND status = ?", userOfferIDs, models.ConversionStatusApproved).
		Scan(&earnings)
	stats.TotalEarnings = earnings.Total

	// Conversion rate
	if stats.TotalClicks > 0 {
		stats.ConversionRate = float64(stats.TotalConversions) / float64(stats.TotalClicks) * 100
	}

	// Cache the result
	if cache.RedisClient != nil {
		if data, err := json.Marshal(stats); err == nil {
			cache.Set(ctx, cacheKey, string(data), 5*time.Minute)
		}
	}

	return stats, nil
}

// InvalidateUserStatsCache clears the cache for a user's stats
func (s *AnalyticsService) InvalidateUserStatsCache(userID uuid.UUID) error {
	if cache.RedisClient == nil {
		return nil
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("user_stats:%s", userID.String())
	return cache.Delete(ctx, cacheKey)
}

// InvalidateOfferStatsCache clears the cache for an offer's stats
func (s *AnalyticsService) InvalidateOfferStatsCache(offerID uuid.UUID) error {
	if cache.RedisClient == nil {
		return nil
	}

	ctx := context.Background()
	cacheKey := fmt.Sprintf("offer_stats:%s", offerID.String())
	return cache.Delete(ctx, cacheKey)
}

// RecordEvent records a tracking event for analytics
func (s *AnalyticsService) RecordEvent(eventType string, userID, offerID, userOfferID *uuid.UUID, data map[string]interface{}, ipAddress, userAgent string) error {
	dataJSON, _ := json.Marshal(data)

	event := models.TrackingEvent{
		ID:          uuid.New(),
		EventType:   eventType,
		UserID:      userID,
		OfferID:     offerID,
		UserOfferID: userOfferID,
		Data:        string(dataJSON),
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		CreatedAt:   time.Now().UTC(),
	}

	return database.DB.Create(&event).Error
}

// GetDailyStats returns click and conversion counts for the last N days
func (s *AnalyticsService) GetDailyStats(userID uuid.UUID, days int) ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	// Get user offer IDs
	var userOfferIDs []uuid.UUID
	if err := database.DB.Model(&models.UserOffer{}).
		Where("user_id = ?", userID).
		Pluck("id", &userOfferIDs).Error; err != nil {
		return nil, err
	}

	if len(userOfferIDs) == 0 {
		return results, nil
	}

	now := time.Now().UTC()

	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i)
		startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
		endOfDay := startOfDay.Add(24 * time.Hour)

		var clicks, conversions int64

		database.DB.Model(&models.Click{}).
			Where("user_offer_id IN ? AND clicked_at >= ? AND clicked_at < ?", userOfferIDs, startOfDay, endOfDay).
			Count(&clicks)

		database.DB.Model(&models.Conversion{}).
			Where("user_offer_id IN ? AND converted_at >= ? AND converted_at < ?", userOfferIDs, startOfDay, endOfDay).
			Count(&conversions)

		results = append(results, map[string]interface{}{
			"date":        startOfDay.Format("2006-01-02"),
			"clicks":      clicks,
			"conversions": conversions,
		})
	}

	return results, nil
}


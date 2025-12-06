package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/database"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// HIGH-PERFORMANCE CLICK SERVICE V2
// ============================================

// ClickServiceV2 provides high-throughput click tracking
type ClickServiceV2 struct {
	db            *gorm.DB
	cacheService  *CacheService
	observability *ObservabilityService
	
	// Micro-batching for DB writes
	clickBuffer   chan *ClickData
	bufferSize    int
	flushInterval time.Duration
	
	// Counter aggregation
	counterBuffer map[string]int64
	counterMutex  sync.Mutex
	
	ctx           context.Context
	cancel        context.CancelFunc
}

// ClickData represents click data for async processing
type ClickData struct {
	ID           uuid.UUID
	UserOfferID  uuid.UUID
	IP           string
	UserAgent    string
	Device       string
	Browser      string
	OS           string
	Country      string
	City         string
	Fingerprint  string
	Referrer     string
	ClickedAt    time.Time
	IsUnique     bool
	RiskScore    int
}

var (
	clickServiceV2Instance *ClickServiceV2
	clickServiceV2Once     sync.Once
)

// NewClickServiceV2 creates a high-performance click service
func NewClickServiceV2() *ClickServiceV2 {
	clickServiceV2Once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		
		clickServiceV2Instance = &ClickServiceV2{
			db:            database.DB,
			cacheService:  NewCacheService(),
			observability: NewObservabilityService(),
			clickBuffer:   make(chan *ClickData, 50000),
			bufferSize:    100,
			flushInterval: 100 * time.Millisecond,
			counterBuffer: make(map[string]int64),
			ctx:           ctx,
			cancel:        cancel,
		}
		
		// Start background workers
		go clickServiceV2Instance.clickFlushWorker()
		go clickServiceV2Instance.counterFlushWorker()
	})
	return clickServiceV2Instance
}

// TrackClickAsync tracks a click asynchronously (non-blocking)
func (s *ClickServiceV2) TrackClickAsync(data *ClickData) bool {
	// Generate ID if not set
	if data.ID == uuid.Nil {
		data.ID = uuid.New()
	}
	if data.ClickedAt.IsZero() {
		data.ClickedAt = time.Now().UTC()
	}

	// Update Redis counters immediately (fast path)
	go s.updateRedisCounters(data)

	// Queue for DB write (slow path)
	select {
	case s.clickBuffer <- data:
		return true
	default:
		// Buffer full, track as dropped
		s.observability.GetMetrics().ClicksBlocked++
		return false
	}
}

// TrackClickSync tracks a click synchronously (blocking)
func (s *ClickServiceV2) TrackClickSync(data *ClickData) (*models.Click, error) {
	if data.ID == uuid.Nil {
		data.ID = uuid.New()
	}
	if data.ClickedAt.IsZero() {
		data.ClickedAt = time.Now().UTC()
	}

	// Update Redis counters
	s.updateRedisCounters(data)

	// Create click record
	click := &models.Click{
		ID:          data.ID,
		UserOfferID: data.UserOfferID,
		IPAddress:   data.IP,
		UserAgent:   data.UserAgent,
		Device:      data.Device,
		Browser:     data.Browser,
		OS:          data.OS,
		Country:     data.Country,
		City:        data.City,
		Fingerprint: data.Fingerprint,
		Referrer:    data.Referrer,
		ClickedAt:   data.ClickedAt,
	}

	// Single atomic transaction
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Insert click
		if err := tx.Create(click).Error; err != nil {
			return err
		}

		// Update UserOffer counter atomically
		if err := tx.Model(&models.UserOffer{}).
			Where("id = ?", data.UserOfferID).
			UpdateColumn("total_clicks", gorm.Expr("total_clicks + 1")).
			Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return click, nil
}

// updateRedisCounters updates Redis counters for real-time stats
func (s *ClickServiceV2) updateRedisCounters(data *ClickData) {
	ctx := context.Background()
	now := time.Now()
	
	// Use pipeline for atomic batch updates
	if cache.RedisClient == nil {
		return
	}

	pipe := cache.Pipeline()
	if pipe == nil {
		return
	}

	userOfferKey := data.UserOfferID.String()[:8]
	
	// Total clicks
	pipe.Incr(ctx, NSCounter+"clicks:total")
	pipe.Incr(ctx, NSCounter+"clicks:offer:"+userOfferKey)
	
	// Hourly clicks
	hourKey := fmt.Sprintf("clicks:hourly:%s:%s", userOfferKey, now.Format("2006010215"))
	pipe.Incr(ctx, NSCounter+hourKey)
	pipe.Expire(ctx, NSCounter+hourKey, 48*time.Hour)
	
	// Daily clicks
	dayKey := fmt.Sprintf("clicks:daily:%s:%s", userOfferKey, now.Format("20060102"))
	pipe.Incr(ctx, NSCounter+dayKey)
	pipe.Expire(ctx, NSCounter+dayKey, 31*24*time.Hour)
	
	// Country stats
	if data.Country != "" {
		pipe.Incr(ctx, NSCounter+"clicks:country:"+data.Country)
	}
	
	// Device stats
	if data.Device != "" {
		pipe.Incr(ctx, NSCounter+"clicks:device:"+data.Device)
	}

	pipe.Exec(ctx)
}

// clickFlushWorker batches click writes to database
func (s *ClickServiceV2) clickFlushWorker() {
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()

	batch := make([]*ClickData, 0, s.bufferSize)

	for {
		select {
		case <-s.ctx.Done():
			if len(batch) > 0 {
				s.flushClickBatch(batch)
			}
			return
		case data := <-s.clickBuffer:
			batch = append(batch, data)
			if len(batch) >= s.bufferSize {
				s.flushClickBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.flushClickBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// flushClickBatch writes a batch of clicks to database
func (s *ClickServiceV2) flushClickBatch(batch []*ClickData) {
	if len(batch) == 0 {
		return
	}

	start := time.Now()

	// Convert to Click models
	clicks := make([]models.Click, len(batch))
	counterUpdates := make(map[uuid.UUID]int)

	for i, data := range batch {
		clicks[i] = models.Click{
			ID:          data.ID,
			UserOfferID: data.UserOfferID,
			IPAddress:   data.IP,
			UserAgent:   data.UserAgent,
			Device:      data.Device,
			Browser:     data.Browser,
			OS:          data.OS,
			Country:     data.Country,
			City:        data.City,
			Fingerprint: data.Fingerprint,
			Referrer:    data.Referrer,
			ClickedAt:   data.ClickedAt,
		}
		counterUpdates[data.UserOfferID]++
	}

	// Batch insert clicks
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// Bulk insert clicks
		if err := tx.CreateInBatches(clicks, 100).Error; err != nil {
			return err
		}

		// Update counters for each UserOffer
		for userOfferID, count := range counterUpdates {
			if err := tx.Model(&models.UserOffer{}).
				Where("id = ?", userOfferID).
				UpdateColumn("total_clicks", gorm.Expr("total_clicks + ?", count)).
				Error; err != nil {
				return err
			}
		}

		return nil
	})

	duration := time.Since(start)
	
	if err != nil {
		fmt.Printf("[ClickServiceV2] Batch flush failed: %v\n", err)
		s.observability.GetMetrics().TotalErrors++
	} else {
		fmt.Printf("[ClickServiceV2] Flushed %d clicks in %v\n", len(batch), duration)
	}
}

// counterFlushWorker periodically syncs in-memory counters to database
func (s *ClickServiceV2) counterFlushWorker() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.flushCounters()
		}
	}
}

// flushCounters syncs counters from Redis to database
func (s *ClickServiceV2) flushCounters() {
	// This can be used for periodic reconciliation
	// For now, counters are updated in real-time via transactions
}

// GetClickCount returns the current click count for a UserOffer
func (s *ClickServiceV2) GetClickCount(userOfferID uuid.UUID) (int64, error) {
	ctx := context.Background()
	key := NSCounter + "clicks:offer:" + userOfferID.String()[:8]
	
	// Try Redis first
	if cache.RedisClient != nil {
		count, err := cache.RedisClient.Get(ctx, key).Int64()
		if err == nil {
			return count, nil
		}
	}

	// Fallback to database
	var count int64
	err := s.db.Model(&models.Click{}).
		Where("user_offer_id = ?", userOfferID).
		Count(&count).Error
	
	return count, err
}

// GetHourlyClicks returns hourly click counts for a UserOffer
func (s *ClickServiceV2) GetHourlyClicks(userOfferID uuid.UUID, hours int) (map[string]int64, error) {
	ctx := context.Background()
	result := make(map[string]int64)
	userOfferKey := userOfferID.String()[:8]

	now := time.Now()
	for i := 0; i < hours; i++ {
		t := now.Add(-time.Duration(i) * time.Hour)
		hourKey := NSCounter + fmt.Sprintf("clicks:hourly:%s:%s", userOfferKey, t.Format("2006010215"))
		
		if cache.RedisClient != nil {
			count, err := cache.RedisClient.Get(ctx, hourKey).Int64()
			if err == nil {
				result[t.Format("2006-01-02 15:00")] = count
			}
		}
	}

	return result, nil
}

// GetDailyClicks returns daily click counts for a UserOffer
func (s *ClickServiceV2) GetDailyClicks(userOfferID uuid.UUID, days int) (map[string]int64, error) {
	ctx := context.Background()
	result := make(map[string]int64)
	userOfferKey := userOfferID.String()[:8]

	now := time.Now()
	for i := 0; i < days; i++ {
		t := now.AddDate(0, 0, -i)
		dayKey := NSCounter + fmt.Sprintf("clicks:daily:%s:%s", userOfferKey, t.Format("20060102"))
		
		if cache.RedisClient != nil {
			count, err := cache.RedisClient.Get(ctx, dayKey).Int64()
			if err == nil {
				result[t.Format("2006-01-02")] = count
			}
		}
	}

	return result, nil
}

// GetClickStats returns aggregated click statistics
func (s *ClickServiceV2) GetClickStats() map[string]interface{} {
	ctx := context.Background()
	stats := make(map[string]interface{})

	if cache.RedisClient == nil {
		return stats
	}

	// Total clicks
	if total, err := cache.RedisClient.Get(ctx, NSCounter+"clicks:total").Int64(); err == nil {
		stats["total_clicks"] = total
	}

	// Get device breakdown
	devices := []string{"mobile", "desktop", "tablet"}
	deviceStats := make(map[string]int64)
	for _, device := range devices {
		if count, err := cache.RedisClient.Get(ctx, NSCounter+"clicks:device:"+device).Int64(); err == nil {
			deviceStats[device] = count
		}
	}
	stats["by_device"] = deviceStats

	return stats
}

// Stop stops the click service
func (s *ClickServiceV2) Stop() {
	s.cancel()
}

// ============================================
// CLICK DEDUPLICATION
// ============================================

// IsClickDuplicate checks if a click is a duplicate
func (s *ClickServiceV2) IsClickDuplicate(fingerprint string, window time.Duration) bool {
	ctx := context.Background()
	key := NSFraud + "click_fp:" + fingerprint

	if cache.RedisClient == nil {
		return false
	}

	// Try to set with NX (only if not exists)
	set, err := cache.SetNX(ctx, key, "1", window)
	if err != nil {
		return false
	}

	// If set succeeded, it's not a duplicate
	return !set
}

// ============================================
// PRECOMPUTED STATS CACHE
// ============================================

// CachedOfferStats represents cached offer statistics
type CachedOfferStats struct {
	TotalClicks      int64     `json:"total_clicks"`
	UniqueClicks     int64     `json:"unique_clicks"`
	TotalConversions int64     `json:"total_conversions"`
	ConversionRate   float64   `json:"conversion_rate"`
	LastUpdated      time.Time `json:"last_updated"`
}

// GetCachedOfferStats returns cached statistics for an offer
func (s *ClickServiceV2) GetCachedOfferStats(userOfferID uuid.UUID) (*CachedOfferStats, error) {
	ctx := context.Background()
	key := NSStats + "offer:" + userOfferID.String()

	// Try cache first
	if cache.RedisClient != nil {
		data, err := cache.RedisClient.Get(ctx, key).Result()
		if err == nil {
			var stats CachedOfferStats
			if json.Unmarshal([]byte(data), &stats) == nil {
				return &stats, nil
			}
		}
	}

	// Compute from database
	stats, err := s.computeOfferStats(userOfferID)
	if err != nil {
		return nil, err
	}

	// Cache for 1 minute
	if cache.RedisClient != nil {
		data, _ := json.Marshal(stats)
		cache.Set(ctx, key, string(data), time.Minute)
	}

	return stats, nil
}

// computeOfferStats computes statistics from database
func (s *ClickServiceV2) computeOfferStats(userOfferID uuid.UUID) (*CachedOfferStats, error) {
	var userOffer models.UserOffer
	if err := s.db.First(&userOffer, "id = ?", userOfferID).Error; err != nil {
		return nil, err
	}

	stats := &CachedOfferStats{
		TotalClicks:      int64(userOffer.TotalClicks),
		TotalConversions: int64(userOffer.TotalConversions),
		LastUpdated:      time.Now(),
	}

	if stats.TotalClicks > 0 {
		stats.ConversionRate = float64(stats.TotalConversions) / float64(stats.TotalClicks) * 100
	}

	return stats, nil
}

// RefreshOfferStatsCache refreshes the cache for an offer
func (s *ClickServiceV2) RefreshOfferStatsCache(userOfferID uuid.UUID) error {
	stats, err := s.computeOfferStats(userOfferID)
	if err != nil {
		return err
	}

	ctx := context.Background()
	key := NSStats + "offer:" + userOfferID.String()
	data, _ := json.Marshal(stats)
	
	return cache.Set(ctx, key, string(data), time.Minute)
}


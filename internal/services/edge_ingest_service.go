package services

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// EDGE CLICK EVENT
// ============================================

// EdgeClickEvent represents a click event from the edge
type EdgeClickEvent struct {
	TrackingCode     string                 `json:"tracking_code"`
	TenantID         string                 `json:"tenant_id"`
	OfferID          string                 `json:"offer_id"`
	UserOfferID      string                 `json:"user_offer_id"`
	Country          string                 `json:"country"`
	Region           string                 `json:"region"`
	City             string                 `json:"city"`
	Device           string                 `json:"device"`
	Browser          string                 `json:"browser"`
	OS               string                 `json:"os"`
	IP               string                 `json:"ip"`
	Timestamp        string                 `json:"timestamp"`
	UserAgent        string                 `json:"user_agent"`
	Referer          string                 `json:"referer"`
	AcceptLanguage   string                 `json:"accept_language"`
	EdgeLocation     string                 `json:"edge_location"`
	LatencyMs        int                    `json:"latency_ms"`
	RouterDecision   string                 `json:"router_decision"`
	FinalDestination string                 `json:"final_destination"`
	Meta             map[string]interface{} `json:"meta"`
}

// EdgeBatchRequest represents a batch of edge click events
type EdgeBatchRequest struct {
	Events []EdgeClickEvent `json:"events"`
}

// ============================================
// EDGE INGEST SERVICE
// ============================================

// EdgeIngestService handles ingestion of edge click events
type EdgeIngestService struct {
	db            *gorm.DB
	observability *ObservabilityService
	linkService   *LinkService
	clickService  *ClickService
	
	// Async processing
	eventQueue    chan EdgeClickEvent
	batchSize     int
	flushInterval time.Duration
	
	// Metrics
	totalReceived  int64
	totalProcessed int64
	totalFailed    int64
	
	// Worker control
	workerWg      sync.WaitGroup
	stopChan      chan struct{}
}

// NewEdgeIngestService creates a new edge ingest service
func NewEdgeIngestService(db *gorm.DB) *EdgeIngestService {
	service := &EdgeIngestService{
		db:            db,
		observability: NewObservabilityService(),
		eventQueue:    make(chan EdgeClickEvent, 10000),
		batchSize:     100,
		flushInterval: 5 * time.Second,
		stopChan:      make(chan struct{}),
	}
	
	return service
}

// SetLinkService sets the link service
func (s *EdgeIngestService) SetLinkService(ls *LinkService) {
	s.linkService = ls
}

// SetClickService sets the click service
func (s *EdgeIngestService) SetClickService(cs *ClickService) {
	s.clickService = cs
}

// Start starts the edge ingest workers
func (s *EdgeIngestService) Start(workerCount int) {
	for i := 0; i < workerCount; i++ {
		s.workerWg.Add(1)
		go s.worker(i)
	}
	
	// Start batch flusher
	s.workerWg.Add(1)
	go s.batchFlusher()
}

// Stop stops the edge ingest workers
func (s *EdgeIngestService) Stop() {
	close(s.stopChan)
	s.workerWg.Wait()
}

// ============================================
// EVENT INGESTION
// ============================================

// IngestEvent ingests a single edge click event
func (s *EdgeIngestService) IngestEvent(event EdgeClickEvent) error {
	atomic.AddInt64(&s.totalReceived, 1)
	
	select {
	case s.eventQueue <- event:
		return nil
	default:
		// Queue full, process synchronously
		return s.processEvent(event)
	}
}

// IngestBatch ingests a batch of edge click events
func (s *EdgeIngestService) IngestBatch(events []EdgeClickEvent) (int, int, error) {
	processed := 0
	failed := 0
	
	for _, event := range events {
		if err := s.IngestEvent(event); err != nil {
			failed++
		} else {
			processed++
		}
	}
	
	return processed, failed, nil
}

// IngestGzip ingests gzip-compressed events
func (s *EdgeIngestService) IngestGzip(r io.Reader) (int, int, error) {
	gzReader, err := gzip.NewReader(r)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to create gzip reader: %w", err)
	}
	defer gzReader.Close()
	
	var batch EdgeBatchRequest
	if err := json.NewDecoder(gzReader).Decode(&batch); err != nil {
		return 0, 0, fmt.Errorf("failed to decode gzip data: %w", err)
	}
	
	return s.IngestBatch(batch.Events)
}

// IngestJSONLines ingests JSON lines format
func (s *EdgeIngestService) IngestJSONLines(r io.Reader) (int, int, error) {
	decoder := json.NewDecoder(r)
	processed := 0
	failed := 0
	
	for {
		var event EdgeClickEvent
		if err := decoder.Decode(&event); err != nil {
			if err == io.EOF {
				break
			}
			failed++
			continue
		}
		
		if err := s.IngestEvent(event); err != nil {
			failed++
		} else {
			processed++
		}
	}
	
	return processed, failed, nil
}

// ============================================
// WORKERS
// ============================================

// worker processes events from the queue
func (s *EdgeIngestService) worker(id int) {
	defer s.workerWg.Done()
	
	for {
		select {
		case <-s.stopChan:
			return
		case event := <-s.eventQueue:
			if err := s.processEvent(event); err != nil {
				atomic.AddInt64(&s.totalFailed, 1)
				s.observability.Log(LogEvent{
					Category: LogCategoryErrorEvent,
					Level:    LogLevelError,
					Message:  "Failed to process edge event",
					Metadata: map[string]interface{}{
						"worker_id":     id,
						"tracking_code": event.TrackingCode,
						"error":         err.Error(),
					},
				})
			} else {
				atomic.AddInt64(&s.totalProcessed, 1)
			}
		}
	}
}

// batchFlusher periodically flushes events
func (s *EdgeIngestService) batchFlusher() {
	defer s.workerWg.Done()
	
	ticker := time.NewTicker(s.flushInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.flushBatch()
		}
	}
}

// flushBatch flushes accumulated events
func (s *EdgeIngestService) flushBatch() {
	// Collect events from queue
	events := make([]EdgeClickEvent, 0, s.batchSize)
	
	for i := 0; i < s.batchSize; i++ {
		select {
		case event := <-s.eventQueue:
			events = append(events, event)
		default:
			break
		}
	}
	
	if len(events) == 0 {
		return
	}
	
	// Process batch
	for _, event := range events {
		if err := s.processEvent(event); err != nil {
			atomic.AddInt64(&s.totalFailed, 1)
		} else {
			atomic.AddInt64(&s.totalProcessed, 1)
		}
	}
}

// ============================================
// EVENT PROCESSING
// ============================================

// processEvent processes a single edge click event
func (s *EdgeIngestService) processEvent(event EdgeClickEvent) error {
	startTime := time.Now()
	
	// Parse timestamp
	clickedAt, err := time.Parse(time.RFC3339, event.Timestamp)
	if err != nil {
		clickedAt = time.Now()
	}
	
	// Resolve user offer ID from tracking code
	var userOfferID uuid.UUID
	var promoterID uuid.UUID
	
	if s.linkService != nil {
		uoID, err := s.linkService.ResolveTrackingCode(event.TrackingCode)
		if err != nil {
			return fmt.Errorf("failed to resolve tracking code: %w", err)
		}
		userOfferID = uoID
		// Get promoter ID from user offer
		var userOffer models.UserOffer
		if err := s.db.First(&userOffer, "id = ?", userOfferID).Error; err == nil {
			promoterID = userOffer.UserID
		}
	} else {
		// Try to parse as UUID directly
		if event.UserOfferID != "" {
			userOfferID, _ = uuid.Parse(event.UserOfferID)
		}
	}
	
	if userOfferID == uuid.Nil {
		return fmt.Errorf("could not resolve user offer ID")
	}
	
	// Parse tenant ID
	tenantID := models.DefaultTenantID
	if event.TenantID != "" {
		if parsed, err := uuid.Parse(event.TenantID); err == nil {
			tenantID = parsed
		}
	}
	
	// Create click record
	click := &models.Click{
		ID:          uuid.New(),
		UserOfferID: userOfferID,
		IPAddress:   event.IP,
		UserAgent:   event.UserAgent,
		Device:      event.Device,
		Browser:     event.Browser,
		OS:          event.OS,
		Country:     event.Country,
		City:        event.City,
		ClickedAt:   clickedAt,
		Fingerprint: s.generateFingerprint(event),
	}
	
	// Process in transaction
	err = s.db.Transaction(func(tx *gorm.DB) error {
		// Create click
		if err := tx.Create(click).Error; err != nil {
			return err
		}
		
		// Update UserOffer stats
		if err := tx.Model(&models.UserOffer{}).
			Where("id = ?", userOfferID).
			Updates(map[string]interface{}{
				"total_clicks": gorm.Expr("total_clicks + 1"),
				"updated_at":   time.Now(),
			}).Error; err != nil {
			return err
		}
		
		// Update Offer stats
		if err := tx.Model(&models.Offer{}).
			Where("id = (SELECT offer_id FROM user_offers WHERE id = ?)", userOfferID).
			Update("total_clicks", gorm.Expr("total_clicks + 1")).Error; err != nil {
			return err
		}
		
		// Update User stats
		if promoterID != uuid.Nil {
			if err := tx.Model(&models.AfftokUser{}).
				Where("id = ?", promoterID).
				Updates(map[string]interface{}{
					"total_clicks":   gorm.Expr("total_clicks + 1"),
					"monthly_clicks": gorm.Expr("monthly_clicks + 1"),
				}).Error; err != nil {
				return err
			}
		}
		
		return nil
	})
	
	if err != nil {
		return fmt.Errorf("failed to process click: %w", err)
	}
	
	// Update Redis counters
	ctx := context.Background()
	cache.Increment(ctx, fmt.Sprintf("clicks:total:%s", userOfferID.String()))
	cache.Increment(ctx, fmt.Sprintf("clicks:daily:%s:%s", userOfferID.String(), time.Now().Format("2006-01-02")))
	cache.Increment(ctx, fmt.Sprintf("tenant:%s:clicks:total", tenantID.String()))
	
	// Log event
	s.observability.Log(LogEvent{
		Category:    LogCategoryClickEvent,
		Level:       LogLevelInfo,
		Message:     "Edge click processed",
		UserOfferID: userOfferID.String(),
		IP:          event.IP,
		UserAgent:   event.UserAgent,
		Device:      event.Device,
		Metadata: map[string]interface{}{
			"country":          event.Country,
			"edge_location":    event.EdgeLocation,
			"edge_latency_ms":  event.LatencyMs,
			"router_decision":  event.RouterDecision,
			"processing_ms":    time.Since(startTime).Milliseconds(),
			"source":           "edge",
		},
	})
	
	return nil
}

// generateFingerprint generates a fingerprint for the click
func (s *EdgeIngestService) generateFingerprint(event EdgeClickEvent) string {
	data := fmt.Sprintf("%s|%s|%s|%s", event.IP, event.UserAgent, event.TrackingCode, event.Timestamp[:10])
	return fmt.Sprintf("%x", data)
}

// ============================================
// METRICS
// ============================================

// GetStats returns ingestion statistics
func (s *EdgeIngestService) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_received":  atomic.LoadInt64(&s.totalReceived),
		"total_processed": atomic.LoadInt64(&s.totalProcessed),
		"total_failed":    atomic.LoadInt64(&s.totalFailed),
		"queue_size":      len(s.eventQueue),
		"queue_capacity":  cap(s.eventQueue),
	}
}

// ============================================
// OFFER CONFIG FOR EDGE
// ============================================

// EdgeOfferConfig represents offer config for edge
type EdgeOfferConfig struct {
	ID            string                 `json:"id"`
	TenantID      string                 `json:"tenant_id"`
	AdvertiserID  string                 `json:"advertiser_id"`
	LandingURL    string                 `json:"landing_url"`
	FallbackURL   string                 `json:"fallback_url,omitempty"`
	DailyCap      int                    `json:"daily_cap,omitempty"`
	TotalCap      int                    `json:"total_cap,omitempty"`
	CurrentClicks int                    `json:"current_clicks,omitempty"`
	Status        string                 `json:"status"`
	RoutingRules  []EdgeRoutingRule      `json:"routing_rules,omitempty"`
	ABTest        *EdgeABTestConfig      `json:"ab_test,omitempty"`
	Rotation      *EdgeRotationConfig    `json:"rotation,omitempty"`
}

// EdgeRoutingRule represents a routing rule for edge
type EdgeRoutingRule struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Priority   int                    `json:"priority"`
	Conditions []map[string]interface{} `json:"conditions"`
	Action     map[string]interface{} `json:"action"`
	Status     string                 `json:"status"`
}

// EdgeABTestConfig represents A/B test config for edge
type EdgeABTestConfig struct {
	Enabled  bool                `json:"enabled"`
	Variants []EdgeABTestVariant `json:"variants"`
}

// EdgeABTestVariant represents an A/B test variant
type EdgeABTestVariant struct {
	ID         string  `json:"id"`
	URL        string  `json:"url"`
	Percentage float64 `json:"percentage"`
}

// EdgeRotationConfig represents rotation config for edge
type EdgeRotationConfig struct {
	Mode         string                   `json:"mode"`
	Destinations []EdgeWeightedDestination `json:"destinations"`
}

// EdgeWeightedDestination represents a weighted destination
type EdgeWeightedDestination struct {
	URL     string `json:"url"`
	Weight  int    `json:"weight"`
	OfferID string `json:"offer_id,omitempty"`
}

// GetOfferConfigForEdge gets offer config for edge caching
func (s *EdgeIngestService) GetOfferConfigForEdge(trackingCode string) (*EdgeOfferConfig, error) {
	// Resolve tracking code
	if s.linkService == nil {
		return nil, fmt.Errorf("link service not configured")
	}
	
	userOfferID, err := s.linkService.ResolveTrackingCode(trackingCode)
	if err != nil {
		return nil, err
	}
	
	// Get user offer
	var userOffer models.UserOffer
	if err := s.db.Preload("Offer").First(&userOffer, "id = ?", userOfferID).Error; err != nil {
		return nil, err
	}
	
	// Build config
	config := &EdgeOfferConfig{
		ID:            userOffer.OfferID.String(),
		TenantID:      models.DefaultTenantID.String(), // TODO: Get from user offer
		AdvertiserID:  userOffer.Offer.NetworkID.String(),
		LandingURL:    userOffer.Offer.DestinationURL,
		FallbackURL:   userOffer.Offer.DestinationURL,
		CurrentClicks: userOffer.TotalClicks,
		Status:        string(userOffer.Status),
	}
	
	return config, nil
}

// ============================================
// GLOBAL INSTANCE
// ============================================

var (
	edgeIngestInstance *EdgeIngestService
	edgeIngestOnce     sync.Once
)

// GetEdgeIngestService returns the global edge ingest service
func GetEdgeIngestService(db *gorm.DB) *EdgeIngestService {
	edgeIngestOnce.Do(func() {
		edgeIngestInstance = NewEdgeIngestService(db)
	})
	return edgeIngestInstance
}


package services

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// CRASH RECOVERY ENGINE
// ============================================

// CrashRecoveryEngine handles system crash recovery
type CrashRecoveryEngine struct {
	db            *gorm.DB
	walService    *WALService
	failoverQueue *FailoverQueue
	observability *ObservabilityService
	
	// State
	isRecovering  bool
	lastRecovery  time.Time
	
	// Metrics
	totalRecovered int64
	totalFailed    int64
	recoveryTimeMs int64
	
	// Handlers
	clickHandler      func(data map[string]interface{}) error
	conversionHandler func(data map[string]interface{}) error
	postbackHandler   func(data map[string]interface{}) error
}

// NewCrashRecoveryEngine creates a new crash recovery engine
func NewCrashRecoveryEngine(db *gorm.DB) *CrashRecoveryEngine {
	return &CrashRecoveryEngine{
		db:            db,
		walService:    GetWALService(),
		failoverQueue: GetFailoverQueue(),
		observability: NewObservabilityService(),
	}
}

// SetClickHandler sets the click recovery handler
func (e *CrashRecoveryEngine) SetClickHandler(handler func(data map[string]interface{}) error) {
	e.clickHandler = handler
}

// SetConversionHandler sets the conversion recovery handler
func (e *CrashRecoveryEngine) SetConversionHandler(handler func(data map[string]interface{}) error) {
	e.conversionHandler = handler
}

// SetPostbackHandler sets the postback recovery handler
func (e *CrashRecoveryEngine) SetPostbackHandler(handler func(data map[string]interface{}) error) {
	e.postbackHandler = handler
}

// ============================================
// RECOVERY OPERATIONS
// ============================================

// Recover performs full crash recovery
func (e *CrashRecoveryEngine) Recover() (*RecoveryResult, error) {
	if e.isRecovering {
		return nil, fmt.Errorf("recovery already in progress")
	}

	e.isRecovering = true
	defer func() {
		e.isRecovering = false
		e.lastRecovery = time.Now()
	}()

	startTime := time.Now()
	result := &RecoveryResult{
		StartTime: startTime,
	}

	e.observability.Log(LogEvent{
		Category: LogCategorySystemEvent,
		Level:    LogLevelInfo,
		Message:  "Starting crash recovery",
	})

	// 1. Recover from WAL
	walRecovered, walFailed, err := e.recoverFromWAL()
	if err != nil {
		e.observability.Log(LogEvent{
			Category: LogCategoryErrorEvent,
			Level:    LogLevelError,
			Message:  "WAL recovery failed",
			Metadata: map[string]interface{}{"error": err.Error()},
		})
	}
	result.WALRecovered = walRecovered
	result.WALFailed = walFailed

	// 2. Recover from failover queue
	queueRecovered, queueFailed, err := e.recoverFromQueue()
	if err != nil {
		e.observability.Log(LogEvent{
			Category: LogCategoryErrorEvent,
			Level:    LogLevelError,
			Message:  "Queue recovery failed",
			Metadata: map[string]interface{}{"error": err.Error()},
		})
	}
	result.QueueRecovered = queueRecovered
	result.QueueFailed = queueFailed

	// 3. Verify database consistency
	inconsistencies, err := e.verifyConsistency()
	if err != nil {
		e.observability.Log(LogEvent{
			Category: LogCategoryErrorEvent,
			Level:    LogLevelError,
			Message:  "Consistency check failed",
			Metadata: map[string]interface{}{"error": err.Error()},
		})
	}
	result.Inconsistencies = inconsistencies

	result.EndTime = time.Now()
	result.DurationMs = result.EndTime.Sub(startTime).Milliseconds()

	atomic.AddInt64(&e.totalRecovered, int64(walRecovered+queueRecovered))
	atomic.AddInt64(&e.totalFailed, int64(walFailed+queueFailed))
	atomic.StoreInt64(&e.recoveryTimeMs, result.DurationMs)

	e.observability.Log(LogEvent{
		Category: LogCategorySystemEvent,
		Level:    LogLevelInfo,
		Message:  "Crash recovery completed",
		Metadata: map[string]interface{}{
			"wal_recovered":    walRecovered,
			"wal_failed":       walFailed,
			"queue_recovered":  queueRecovered,
			"queue_failed":     queueFailed,
			"inconsistencies":  inconsistencies,
			"duration_ms":      result.DurationMs,
		},
	})

	return result, nil
}

// recoverFromWAL recovers events from the Write-Ahead Log
func (e *CrashRecoveryEngine) recoverFromWAL() (int, int, error) {
	if e.walService == nil {
		return 0, 0, nil
	}

	processor := func(entry *WALEntry) error {
		return e.processEntry(entry)
	}

	return e.walService.Replay(processor)
}

// recoverFromQueue recovers events from the failover queue
func (e *CrashRecoveryEngine) recoverFromQueue() (int, int, error) {
	if e.failoverQueue == nil {
		return 0, 0, nil
	}

	var recovered, failed int

	// Process all queued events
	stats := e.failoverQueue.GetStats()
	bufferSize := stats["buffer_size"].(int)

	for i := 0; i < bufferSize; i++ {
		e.failoverQueue.ProcessQueue()
	}

	recovered = int(stats["total_sent"].(int64))
	failed = int(stats["total_failed"].(int64))

	return recovered, failed, nil
}

// processEntry processes a single WAL entry
func (e *CrashRecoveryEngine) processEntry(entry *WALEntry) error {
	switch entry.EventType {
	case WALEventClick:
		if e.clickHandler != nil {
			return e.clickHandler(entry.Data)
		}
		return e.defaultClickHandler(entry.Data)

	case WALEventConversion:
		if e.conversionHandler != nil {
			return e.conversionHandler(entry.Data)
		}
		return e.defaultConversionHandler(entry.Data)

	case WALEventPostback:
		if e.postbackHandler != nil {
			return e.postbackHandler(entry.Data)
		}
		return e.defaultPostbackHandler(entry.Data)

	default:
		return nil
	}
}

// ============================================
// DEFAULT HANDLERS
// ============================================

// defaultClickHandler handles click recovery
func (e *CrashRecoveryEngine) defaultClickHandler(data map[string]interface{}) error {
	// Extract click data
	userOfferIDStr, _ := data["user_offer_id"].(string)
	if userOfferIDStr == "" {
		return fmt.Errorf("missing user_offer_id")
	}

	userOfferID, err := uuid.Parse(userOfferIDStr)
	if err != nil {
		return fmt.Errorf("invalid user_offer_id: %w", err)
	}

	// Check if click already exists (dedup)
	fingerprint, _ := data["fingerprint"].(string)
	if fingerprint != "" {
		var count int64
		e.db.Model(&models.Click{}).Where("fingerprint = ?", fingerprint).Count(&count)
		if count > 0 {
			return nil // Already processed
		}
	}

	// Create click record
	click := &models.Click{
		ID:          uuid.New(),
		UserOfferID: userOfferID,
	}

	if ip, ok := data["ip"].(string); ok {
		click.IPAddress = ip
	}
	if ua, ok := data["user_agent"].(string); ok {
		click.UserAgent = ua
	}
	if device, ok := data["device"].(string); ok {
		click.Device = device
	}
	if country, ok := data["country"].(string); ok {
		click.Country = country
	}
	if fingerprint != "" {
		click.Fingerprint = fingerprint
	}
	click.ClickedAt = time.Now()

	return e.db.Create(click).Error
}

// defaultConversionHandler handles conversion recovery
func (e *CrashRecoveryEngine) defaultConversionHandler(data map[string]interface{}) error {
	// Extract conversion data
	externalID, _ := data["external_id"].(string)
	if externalID == "" {
		return fmt.Errorf("missing external_id")
	}

	// Check if conversion already exists (dedup)
	var count int64
	e.db.Model(&models.Conversion{}).Where("external_conversion_id = ?", externalID).Count(&count)
	if count > 0 {
		return nil // Already processed
	}

	// Create conversion record
	conversion := &models.Conversion{
		ID:                   uuid.New(),
		ExternalConversionID: externalID,
		Status:               "pending",
		ConvertedAt:          time.Now(),
	}

	if userOfferIDStr, ok := data["user_offer_id"].(string); ok {
		if id, err := uuid.Parse(userOfferIDStr); err == nil {
			conversion.UserOfferID = id
		}
	}
	if clickIDStr, ok := data["click_id"].(string); ok {
		if id, err := uuid.Parse(clickIDStr); err == nil {
			conversion.ClickID = &id
		}
	}
	if amount, ok := data["amount"].(float64); ok {
		conversion.Amount = int(amount)
	}

	return e.db.Create(conversion).Error
}

// defaultPostbackHandler handles postback recovery
func (e *CrashRecoveryEngine) defaultPostbackHandler(data map[string]interface{}) error {
	// Log postback for manual review
	e.observability.Log(LogEvent{
		Category: LogCategoryPostbackEvent,
		Level:    LogLevelWarn,
		Message:  "Recovered postback requires manual review",
		Metadata: data,
	})
	return nil
}

// ============================================
// CONSISTENCY VERIFICATION
// ============================================

// verifyConsistency verifies database consistency
func (e *CrashRecoveryEngine) verifyConsistency() (int, error) {
	inconsistencies := 0

	// 1. Check for orphaned clicks (no matching user_offer)
	var orphanedClicks int64
	e.db.Raw(`
		SELECT COUNT(*) FROM clicks c
		LEFT JOIN user_offers uo ON c.user_offer_id = uo.id
		WHERE uo.id IS NULL
	`).Scan(&orphanedClicks)
	inconsistencies += int(orphanedClicks)

	// 2. Check for orphaned conversions (no matching user_offer)
	var orphanedConversions int64
	e.db.Raw(`
		SELECT COUNT(*) FROM conversions c
		LEFT JOIN user_offers uo ON c.user_offer_id = uo.id
		WHERE uo.id IS NULL
	`).Scan(&orphanedConversions)
	inconsistencies += int(orphanedConversions)

	// 3. Check for click count mismatches
	var clickMismatches int64
	e.db.Raw(`
		SELECT COUNT(*) FROM user_offers uo
		WHERE uo.total_clicks != (
			SELECT COUNT(*) FROM clicks c WHERE c.user_offer_id = uo.id
		)
	`).Scan(&clickMismatches)
	inconsistencies += int(clickMismatches)

	// 4. Check for conversion count mismatches
	var conversionMismatches int64
	e.db.Raw(`
		SELECT COUNT(*) FROM user_offers uo
		WHERE uo.total_conversions != (
			SELECT COUNT(*) FROM conversions c WHERE c.user_offer_id = uo.id
		)
	`).Scan(&conversionMismatches)
	inconsistencies += int(conversionMismatches)

	return inconsistencies, nil
}

// FixInconsistencies attempts to fix detected inconsistencies
func (e *CrashRecoveryEngine) FixInconsistencies() (int, error) {
	fixed := 0

	// Fix click count mismatches
	result := e.db.Exec(`
		UPDATE user_offers SET total_clicks = (
			SELECT COUNT(*) FROM clicks WHERE clicks.user_offer_id = user_offers.id
		)
		WHERE total_clicks != (
			SELECT COUNT(*) FROM clicks WHERE clicks.user_offer_id = user_offers.id
		)
	`)
	fixed += int(result.RowsAffected)

	// Fix conversion count mismatches
	result = e.db.Exec(`
		UPDATE user_offers SET total_conversions = (
			SELECT COUNT(*) FROM conversions WHERE conversions.user_offer_id = user_offers.id
		)
		WHERE total_conversions != (
			SELECT COUNT(*) FROM conversions WHERE conversions.user_offer_id = user_offers.id
		)
	`)
	fixed += int(result.RowsAffected)

	return fixed, nil
}

// ============================================
// METRICS
// ============================================

// RecoveryResult holds the result of a recovery operation
type RecoveryResult struct {
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	DurationMs      int64     `json:"duration_ms"`
	WALRecovered    int       `json:"wal_recovered"`
	WALFailed       int       `json:"wal_failed"`
	QueueRecovered  int       `json:"queue_recovered"`
	QueueFailed     int       `json:"queue_failed"`
	Inconsistencies int       `json:"inconsistencies"`
}

// GetStats returns recovery engine statistics
func (e *CrashRecoveryEngine) GetStats() map[string]interface{} {
	walStats := map[string]interface{}{}
	if e.walService != nil {
		walStats = e.walService.GetStats()
	}

	queueStats := map[string]interface{}{}
	if e.failoverQueue != nil {
		queueStats = e.failoverQueue.GetStats()
	}

	return map[string]interface{}{
		"is_recovering":     e.isRecovering,
		"last_recovery":     e.lastRecovery,
		"total_recovered":   atomic.LoadInt64(&e.totalRecovered),
		"total_failed":      atomic.LoadInt64(&e.totalFailed),
		"last_recovery_ms":  atomic.LoadInt64(&e.recoveryTimeMs),
		"wal_stats":         walStats,
		"queue_stats":       queueStats,
	}
}

// ============================================
// ZERO-DROP MODE
// ============================================

// ZeroDropMode represents the zero-drop mode state
type ZeroDropMode struct {
	mu              sync.RWMutex
	enabled         bool
	tenantSettings  map[string]bool
	observability   *ObservabilityService
}

// NewZeroDropMode creates a new zero-drop mode manager
func NewZeroDropMode() *ZeroDropMode {
	return &ZeroDropMode{
		tenantSettings: make(map[string]bool),
		observability:  NewObservabilityService(),
	}
}

// Enable enables zero-drop mode globally
func (z *ZeroDropMode) Enable() {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.enabled = true
	z.observability.Log(LogEvent{
		Category: LogCategorySystemEvent,
		Level:    LogLevelInfo,
		Message:  "Zero-drop mode enabled globally",
	})
}

// Disable disables zero-drop mode globally
func (z *ZeroDropMode) Disable() {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.enabled = false
	z.observability.Log(LogEvent{
		Category: LogCategorySystemEvent,
		Level:    LogLevelInfo,
		Message:  "Zero-drop mode disabled globally",
	})
}

// IsEnabled checks if zero-drop mode is enabled
func (z *ZeroDropMode) IsEnabled() bool {
	z.mu.RLock()
	defer z.mu.RUnlock()
	return z.enabled
}

// EnableForTenant enables zero-drop mode for a specific tenant
func (z *ZeroDropMode) EnableForTenant(tenantID string) {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.tenantSettings[tenantID] = true
}

// DisableForTenant disables zero-drop mode for a specific tenant
func (z *ZeroDropMode) DisableForTenant(tenantID string) {
	z.mu.Lock()
	defer z.mu.Unlock()
	z.tenantSettings[tenantID] = false
}

// IsEnabledForTenant checks if zero-drop mode is enabled for a tenant
func (z *ZeroDropMode) IsEnabledForTenant(tenantID string) bool {
	z.mu.RLock()
	defer z.mu.RUnlock()
	
	// Check tenant-specific setting first
	if enabled, exists := z.tenantSettings[tenantID]; exists {
		return enabled
	}
	
	// Fall back to global setting
	return z.enabled
}

// GetStatus returns the current zero-drop mode status
func (z *ZeroDropMode) GetStatus() map[string]interface{} {
	z.mu.RLock()
	defer z.mu.RUnlock()
	
	return map[string]interface{}{
		"enabled":         z.enabled,
		"tenant_settings": z.tenantSettings,
	}
}

// ============================================
// GLOBAL INSTANCES
// ============================================

var (
	crashRecoveryInstance *CrashRecoveryEngine
	crashRecoveryOnce     sync.Once
	zeroDropModeInstance  *ZeroDropMode
	zeroDropModeOnce      sync.Once
)

// GetCrashRecoveryEngine returns the global crash recovery engine
func GetCrashRecoveryEngine(db *gorm.DB) *CrashRecoveryEngine {
	crashRecoveryOnce.Do(func() {
		crashRecoveryInstance = NewCrashRecoveryEngine(db)
	})
	return crashRecoveryInstance
}

// GetZeroDropMode returns the global zero-drop mode manager
func GetZeroDropMode() *ZeroDropMode {
	zeroDropModeOnce.Do(func() {
		zeroDropModeInstance = NewZeroDropMode()
	})
	return zeroDropModeInstance
}

// InitZeroDrop initializes the zero-drop system
func InitZeroDrop(db *gorm.DB) error {
	ctx := context.Background()
	
	// Initialize WAL service
	walService := GetWALService()
	walService.Start()
	
	// Initialize failover queue
	failoverQueue := GetFailoverQueue()
	failoverQueue.Start()
	
	// Initialize crash recovery engine
	recoveryEngine := GetCrashRecoveryEngine(db)
	
	// Perform recovery on startup
	result, err := recoveryEngine.Recover()
	if err != nil {
		return fmt.Errorf("crash recovery failed: %w", err)
	}
	
	// Log recovery result
	if result.WALRecovered > 0 || result.QueueRecovered > 0 {
		fmt.Printf("✅ Crash recovery: WAL=%d, Queue=%d, Duration=%dms\n",
			result.WALRecovered, result.QueueRecovered, result.DurationMs)
	}
	
	// Initialize stream consumer
	streamConsumer := GetStreamConsumer()
	if err := streamConsumer.Start(); err != nil {
		// Non-fatal if Redis is not available
		fmt.Printf("⚠️ Stream consumer not started: %v\n", err)
	}
	
	// Enable zero-drop mode by default
	zeroDropMode := GetZeroDropMode()
	zeroDropMode.Enable()
	
	_ = ctx // Suppress unused variable warning
	
	return nil
}

// ShutdownZeroDrop gracefully shuts down the zero-drop system
func ShutdownZeroDrop() {
	// Stop stream consumer
	if streamConsumerInstance != nil {
		streamConsumerInstance.Stop()
	}
	
	// Stop failover queue
	if failoverQueueInstance != nil {
		failoverQueueInstance.Stop()
	}
	
	// Stop WAL service
	if walServiceInstance != nil {
		walServiceInstance.Stop()
	}
}


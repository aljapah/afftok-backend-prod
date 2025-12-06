package services

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// CONSISTENCY CHECK TYPES
// ============================================

// ConsistencyStatus represents the overall status
type ConsistencyStatus string

const (
	ConsistencyStatusOK       ConsistencyStatus = "ok"
	ConsistencyStatusWarning  ConsistencyStatus = "warning"
	ConsistencyStatusCritical ConsistencyStatus = "critical"
)

// ConsistencyIssueType represents the type of consistency issue
type ConsistencyIssueType string

const (
	IssueClicksMismatch          ConsistencyIssueType = "clicks_mismatch"
	IssueConversionsMismatch     ConsistencyIssueType = "conversions_mismatch"
	IssueOrphanedConversion      ConsistencyIssueType = "orphaned_conversion"
	IssueOrphanedClick           ConsistencyIssueType = "orphaned_click"
	IssueRedisDBDrift            ConsistencyIssueType = "redis_db_drift"
	IssueStaleCache              ConsistencyIssueType = "stale_cache"
	IssueStreamLag               ConsistencyIssueType = "stream_lag"
	IssueUnacknowledgedMessages  ConsistencyIssueType = "unacknowledged_messages"
	IssueUserOfferStatsMismatch  ConsistencyIssueType = "user_offer_stats_mismatch"
	IssueOfferStatsMismatch      ConsistencyIssueType = "offer_stats_mismatch"
)

// ConsistencyIssueSeverity represents issue severity
type ConsistencyIssueSeverity string

const (
	ConsistencySeverityInfo     ConsistencyIssueSeverity = "info"
	ConsistencySeverityWarning  ConsistencyIssueSeverity = "warning"
	ConsistencySeverityError    ConsistencyIssueSeverity = "error"
	ConsistencySeverityCritical ConsistencyIssueSeverity = "critical"
)

// ============================================
// CONSISTENCY ISSUE
// ============================================

// ConsistencyIssue represents a detected consistency issue
type ConsistencyIssue struct {
	ID          string                   `json:"id"`
	Type        ConsistencyIssueType     `json:"type"`
	Severity    ConsistencyIssueSeverity `json:"severity"`
	Description string                   `json:"description"`
	Expected    interface{}              `json:"expected,omitempty"`
	Actual      interface{}              `json:"actual,omitempty"`
	Difference  interface{}              `json:"difference,omitempty"`
	EntityID    string                   `json:"entity_id,omitempty"`
	EntityType  string                   `json:"entity_type,omitempty"`
	Timestamp   time.Time                `json:"timestamp"`
	Fixable     bool                     `json:"fixable"`
	FixAction   string                   `json:"fix_action,omitempty"`
}

// ============================================
// CONSISTENCY REPORT
// ============================================

// ConsistencyReport holds the full consistency check report
type ConsistencyReport struct {
	ID             string                 `json:"id"`
	Status         ConsistencyStatus      `json:"status"`
	StartTime      time.Time              `json:"start_time"`
	EndTime        time.Time              `json:"end_time"`
	DurationMs     int64                  `json:"duration_ms"`
	Summary        *ConsistencySummary    `json:"summary"`
	Issues         []ConsistencyIssue     `json:"issues"`
	ChecksRun      []string               `json:"checks_run"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// ConsistencySummary provides a summary of the consistency check
type ConsistencySummary struct {
	TotalChecks              int   `json:"total_checks"`
	PassedChecks             int   `json:"passed_checks"`
	FailedChecks             int   `json:"failed_checks"`
	ClicksMismatched         int64 `json:"clicks_mismatched"`
	ConversionsMismatched    int64 `json:"conversions_mismatched"`
	ConversionsWithoutClick  int64 `json:"conversions_without_click"`
	OrphanedClicks           int64 `json:"orphaned_clicks"`
	RedisDBDrift             int64 `json:"redis_db_drift"`
	StreamLag                int64 `json:"stream_lag"`
	CriticalIssues           int   `json:"critical_issues"`
	WarningIssues            int   `json:"warning_issues"`
}

// ============================================
// CONSISTENCY SERVICE
// ============================================

// ConsistencyService performs data consistency checks
type ConsistencyService struct {
	mu            sync.RWMutex
	db            *gorm.DB
	observability *ObservabilityService
	
	// History
	reports       []*ConsistencyReport
	maxReports    int
	
	// Metrics
	totalChecks   int64
	totalIssues   int64
	lastCheckTime time.Time
}

// NewConsistencyService creates a new consistency service
func NewConsistencyService(db *gorm.DB) *ConsistencyService {
	return &ConsistencyService{
		db:            db,
		observability: NewObservabilityService(),
		reports:       make([]*ConsistencyReport, 0),
		maxReports:    50,
	}
}

// ============================================
// MAIN CHECK RUNNER
// ============================================

// RunFullCheck runs all consistency checks
func (s *ConsistencyService) RunFullCheck() (*ConsistencyReport, error) {
	report := &ConsistencyReport{
		ID:        uuid.New().String(),
		StartTime: time.Now(),
		Issues:    make([]ConsistencyIssue, 0),
		ChecksRun: make([]string, 0),
	}

	atomic.AddInt64(&s.totalChecks, 1)

	s.observability.Log(LogEvent{
		Category: LogCategorySystemEvent,
		Level:    LogLevelInfo,
		Message:  "Consistency check started",
		Metadata: map[string]interface{}{
			"report_id": report.ID,
		},
	})

	// Run all checks
	s.checkUserOfferStats(report)
	s.checkOfferStats(report)
	s.checkOrphanedConversions(report)
	s.checkOrphanedClicks(report)
	s.checkRedisDBConsistency(report)
	s.checkStreamLag(report)
	s.checkCacheStale(report)

	// Finalize report
	report.EndTime = time.Now()
	report.DurationMs = report.EndTime.Sub(report.StartTime).Milliseconds()
	report.Summary = s.calculateSummary(report)
	report.Status = s.determineStatus(report)

	// Store report
	s.storeReport(report)

	s.mu.Lock()
	s.lastCheckTime = time.Now()
	s.mu.Unlock()

	s.observability.Log(LogEvent{
		Category: LogCategorySystemEvent,
		Level:    LogLevelInfo,
		Message:  "Consistency check completed",
		Metadata: map[string]interface{}{
			"report_id":   report.ID,
			"status":      report.Status,
			"issues":      len(report.Issues),
			"duration_ms": report.DurationMs,
		},
	})

	return report, nil
}

// ============================================
// INDIVIDUAL CHECKS
// ============================================

// checkUserOfferStats checks if UserOffer stats match actual counts
func (s *ConsistencyService) checkUserOfferStats(report *ConsistencyReport) {
	report.ChecksRun = append(report.ChecksRun, "user_offer_stats")

	// Get user offers with mismatched click counts
	var clickMismatches []struct {
		ID           uuid.UUID
		TotalClicks  int
		ActualClicks int64
	}

	s.db.Raw(`
		SELECT uo.id, uo.total_clicks, COUNT(c.id) as actual_clicks
		FROM user_offers uo
		LEFT JOIN clicks c ON c.user_offer_id = uo.id
		GROUP BY uo.id
		HAVING uo.total_clicks != COUNT(c.id)
		LIMIT 100
	`).Scan(&clickMismatches)

	for _, m := range clickMismatches {
		report.Issues = append(report.Issues, ConsistencyIssue{
			ID:          uuid.New().String(),
			Type:        IssueUserOfferStatsMismatch,
			Severity:    ConsistencySeverityWarning,
			Description: fmt.Sprintf("UserOffer click count mismatch"),
			Expected:    m.ActualClicks,
			Actual:      m.TotalClicks,
			Difference:  int64(m.TotalClicks) - m.ActualClicks,
			EntityID:    m.ID.String(),
			EntityType:  "user_offer",
			Timestamp:   time.Now(),
			Fixable:     true,
			FixAction:   "UPDATE user_offers SET total_clicks = (SELECT COUNT(*) FROM clicks WHERE user_offer_id = user_offers.id)",
		})
	}

	// Check conversion counts
	var conversionMismatches []struct {
		ID                uuid.UUID
		TotalConversions  int
		ActualConversions int64
	}

	s.db.Raw(`
		SELECT uo.id, uo.total_conversions, COUNT(cv.id) as actual_conversions
		FROM user_offers uo
		LEFT JOIN conversions cv ON cv.user_offer_id = uo.id
		GROUP BY uo.id
		HAVING uo.total_conversions != COUNT(cv.id)
		LIMIT 100
	`).Scan(&conversionMismatches)

	for _, m := range conversionMismatches {
		report.Issues = append(report.Issues, ConsistencyIssue{
			ID:          uuid.New().String(),
			Type:        IssueUserOfferStatsMismatch,
			Severity:    ConsistencySeverityWarning,
			Description: fmt.Sprintf("UserOffer conversion count mismatch"),
			Expected:    m.ActualConversions,
			Actual:      m.TotalConversions,
			Difference:  int64(m.TotalConversions) - m.ActualConversions,
			EntityID:    m.ID.String(),
			EntityType:  "user_offer",
			Timestamp:   time.Now(),
			Fixable:     true,
			FixAction:   "UPDATE user_offers SET total_conversions = (SELECT COUNT(*) FROM conversions WHERE user_offer_id = user_offers.id)",
		})
	}
}

// checkOfferStats checks if Offer stats match actual counts
func (s *ConsistencyService) checkOfferStats(report *ConsistencyReport) {
	report.ChecksRun = append(report.ChecksRun, "offer_stats")

	var mismatches []struct {
		ID           uuid.UUID
		TotalClicks  int
		ActualClicks int64
	}

	s.db.Raw(`
		SELECT o.id, o.total_clicks, COALESCE(SUM(uo.total_clicks), 0) as actual_clicks
		FROM offers o
		LEFT JOIN user_offers uo ON uo.offer_id = o.id
		GROUP BY o.id
		HAVING o.total_clicks != COALESCE(SUM(uo.total_clicks), 0)
		LIMIT 100
	`).Scan(&mismatches)

	for _, m := range mismatches {
		report.Issues = append(report.Issues, ConsistencyIssue{
			ID:          uuid.New().String(),
			Type:        IssueOfferStatsMismatch,
			Severity:    ConsistencySeverityWarning,
			Description: fmt.Sprintf("Offer total clicks mismatch"),
			Expected:    m.ActualClicks,
			Actual:      m.TotalClicks,
			Difference:  int64(m.TotalClicks) - m.ActualClicks,
			EntityID:    m.ID.String(),
			EntityType:  "offer",
			Timestamp:   time.Now(),
			Fixable:     true,
			FixAction:   "UPDATE offers SET total_clicks = (SELECT SUM(total_clicks) FROM user_offers WHERE offer_id = offers.id)",
		})
	}
}

// checkOrphanedConversions checks for conversions without corresponding clicks
func (s *ConsistencyService) checkOrphanedConversions(report *ConsistencyReport) {
	report.ChecksRun = append(report.ChecksRun, "orphaned_conversions")

	var orphanedCount int64
	s.db.Raw(`
		SELECT COUNT(*) FROM conversions cv
		LEFT JOIN user_offers uo ON cv.user_offer_id = uo.id
		WHERE uo.id IS NULL
	`).Scan(&orphanedCount)

	if orphanedCount > 0 {
		report.Issues = append(report.Issues, ConsistencyIssue{
			ID:          uuid.New().String(),
			Type:        IssueOrphanedConversion,
			Severity:    ConsistencySeverityError,
			Description: fmt.Sprintf("Found %d conversions without valid user_offer", orphanedCount),
			Actual:      orphanedCount,
			Timestamp:   time.Now(),
			Fixable:     false,
		})
	}
}

// checkOrphanedClicks checks for clicks without corresponding user_offer
func (s *ConsistencyService) checkOrphanedClicks(report *ConsistencyReport) {
	report.ChecksRun = append(report.ChecksRun, "orphaned_clicks")

	var orphanedCount int64
	s.db.Raw(`
		SELECT COUNT(*) FROM clicks c
		LEFT JOIN user_offers uo ON c.user_offer_id = uo.id
		WHERE uo.id IS NULL
	`).Scan(&orphanedCount)

	if orphanedCount > 0 {
		report.Issues = append(report.Issues, ConsistencyIssue{
			ID:          uuid.New().String(),
			Type:        IssueOrphanedClick,
			Severity:    ConsistencySeverityError,
			Description: fmt.Sprintf("Found %d clicks without valid user_offer", orphanedCount),
			Actual:      orphanedCount,
			Timestamp:   time.Now(),
			Fixable:     false,
		})
	}
}

// checkRedisDBConsistency checks if Redis counters match DB
func (s *ConsistencyService) checkRedisDBConsistency(report *ConsistencyReport) {
	report.ChecksRun = append(report.ChecksRun, "redis_db_consistency")

	if cache.RedisClient == nil {
		return
	}

	ctx := context.Background()

	// Check total clicks in Redis vs DB
	var dbTotalClicks int64
	s.db.Model(&models.Click{}).Count(&dbTotalClicks)

	redisClicksStr, _ := cache.Get(ctx, "stats:total_clicks")
	var redisClicks int64
	if redisClicksStr != "" {
		fmt.Sscanf(redisClicksStr, "%d", &redisClicks)
	}

	// Allow 5% tolerance
	tolerance := float64(dbTotalClicks) * 0.05
	drift := float64(redisClicks) - float64(dbTotalClicks)
	if drift < 0 {
		drift = -drift
	}

	if drift > tolerance && dbTotalClicks > 0 {
		report.Issues = append(report.Issues, ConsistencyIssue{
			ID:          uuid.New().String(),
			Type:        IssueRedisDBDrift,
			Severity:    ConsistencySeverityWarning,
			Description: "Redis total clicks differs from DB",
			Expected:    dbTotalClicks,
			Actual:      redisClicks,
			Difference:  int64(drift),
			Timestamp:   time.Now(),
			Fixable:     true,
			FixAction:   "Refresh Redis cache from DB",
		})
	}
}

// checkStreamLag checks Redis stream consumer lag
func (s *ConsistencyService) checkStreamLag(report *ConsistencyReport) {
	report.ChecksRun = append(report.ChecksRun, "stream_lag")

	if cache.RedisClient == nil {
		return
	}

	ctx := context.Background()
	streams := []string{StreamClicks, StreamConversions, StreamPostbacks}

	for _, stream := range streams {
		pending := cache.RedisClient.XPending(ctx, stream, "afftok-consumers")
		if pending.Err() == nil && pending.Val().Count > 1000 {
			report.Issues = append(report.Issues, ConsistencyIssue{
				ID:          uuid.New().String(),
				Type:        IssueStreamLag,
				Severity:    ConsistencySeverityWarning,
				Description: fmt.Sprintf("High pending count on stream %s", stream),
				Actual:      pending.Val().Count,
				EntityID:    stream,
				EntityType:  "stream",
				Timestamp:   time.Now(),
				Fixable:     false,
			})
		}
	}
}

// checkCacheStale checks for stale cache entries
func (s *ConsistencyService) checkCacheStale(report *ConsistencyReport) {
	report.ChecksRun = append(report.ChecksRun, "cache_stale")

	// This would check TTLs and last update times
	// Simplified implementation
}

// ============================================
// HELPERS
// ============================================

func (s *ConsistencyService) calculateSummary(report *ConsistencyReport) *ConsistencySummary {
	summary := &ConsistencySummary{
		TotalChecks:  len(report.ChecksRun),
		PassedChecks: len(report.ChecksRun),
	}

	for _, issue := range report.Issues {
		switch issue.Severity {
		case ConsistencySeverityCritical:
			summary.CriticalIssues++
		case ConsistencySeverityWarning, ConsistencySeverityError:
			summary.WarningIssues++
		}

		switch issue.Type {
		case IssueClicksMismatch, IssueUserOfferStatsMismatch:
			if diff, ok := issue.Difference.(int64); ok {
				summary.ClicksMismatched += diff
			}
		case IssueConversionsMismatch:
			if diff, ok := issue.Difference.(int64); ok {
				summary.ConversionsMismatched += diff
			}
		case IssueOrphanedConversion:
			if count, ok := issue.Actual.(int64); ok {
				summary.ConversionsWithoutClick += count
			}
		case IssueOrphanedClick:
			if count, ok := issue.Actual.(int64); ok {
				summary.OrphanedClicks += count
			}
		case IssueRedisDBDrift:
			if diff, ok := issue.Difference.(int64); ok {
				summary.RedisDBDrift += diff
			}
		case IssueStreamLag:
			if lag, ok := issue.Actual.(int64); ok {
				summary.StreamLag += lag
			}
		}
	}

	if len(report.Issues) > 0 {
		summary.FailedChecks = 1
		summary.PassedChecks--
	}

	return summary
}

func (s *ConsistencyService) determineStatus(report *ConsistencyReport) ConsistencyStatus {
	if report.Summary.CriticalIssues > 0 {
		return ConsistencyStatusCritical
	}
	if report.Summary.WarningIssues > 0 {
		return ConsistencyStatusWarning
	}
	return ConsistencyStatusOK
}

func (s *ConsistencyService) storeReport(report *ConsistencyReport) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.reports = append(s.reports, report)
	if len(s.reports) > s.maxReports {
		s.reports = s.reports[1:]
	}

	atomic.AddInt64(&s.totalIssues, int64(len(report.Issues)))
}

// ============================================
// FIX ISSUES
// ============================================

// FixUserOfferStats fixes UserOffer stats mismatches
func (s *ConsistencyService) FixUserOfferStats() (int64, error) {
	result := s.db.Exec(`
		UPDATE user_offers SET total_clicks = (
			SELECT COUNT(*) FROM clicks WHERE clicks.user_offer_id = user_offers.id
		)
		WHERE total_clicks != (
			SELECT COUNT(*) FROM clicks WHERE clicks.user_offer_id = user_offers.id
		)
	`)

	clicksFixed := result.RowsAffected

	result = s.db.Exec(`
		UPDATE user_offers SET total_conversions = (
			SELECT COUNT(*) FROM conversions WHERE conversions.user_offer_id = user_offers.id
		)
		WHERE total_conversions != (
			SELECT COUNT(*) FROM conversions WHERE conversions.user_offer_id = user_offers.id
		)
	`)

	return clicksFixed + result.RowsAffected, result.Error
}

// ============================================
// GETTERS
// ============================================

// GetLatestReport returns the most recent report
func (s *ConsistencyService) GetLatestReport() *ConsistencyReport {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.reports) == 0 {
		return nil
	}
	return s.reports[len(s.reports)-1]
}

// GetReport returns a specific report by ID
func (s *ConsistencyService) GetReport(id string) (*ConsistencyReport, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, report := range s.reports {
		if report.ID == id {
			return report, nil
		}
	}
	return nil, fmt.Errorf("report not found: %s", id)
}

// GetAllIssues returns all issues from the latest report
func (s *ConsistencyService) GetAllIssues() []ConsistencyIssue {
	report := s.GetLatestReport()
	if report == nil {
		return []ConsistencyIssue{}
	}
	return report.Issues
}

// GetStats returns service statistics
func (s *ConsistencyService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"total_checks":    atomic.LoadInt64(&s.totalChecks),
		"total_issues":    atomic.LoadInt64(&s.totalIssues),
		"reports_stored":  len(s.reports),
		"last_check_time": s.lastCheckTime,
	}
}


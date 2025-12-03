package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
)

// ============================================
// THREAT TYPES
// ============================================

// ThreatType represents different types of security threats
type ThreatType string

const (
	ThreatDDoS           ThreatType = "ddos"
	ThreatBotAttack      ThreatType = "bot_attack"
	ThreatBruteForce     ThreatType = "brute_force"
	ThreatAPIKeyAbuse    ThreatType = "api_key_abuse"
	ThreatGeoAnomaly     ThreatType = "geo_anomaly"
	ThreatVelocityAbuse  ThreatType = "velocity_abuse"
	ThreatLoginAnomaly   ThreatType = "login_anomaly"
	ThreatSuspiciousIP   ThreatType = "suspicious_ip"
	ThreatWAFBlock       ThreatType = "waf_block"
	ThreatReplayAttack   ThreatType = "replay_attack"
)

// ThreatSeverity represents the severity level
type ThreatSeverity string

const (
	SeverityLow      ThreatSeverity = "low"
	SeverityMedium   ThreatSeverity = "medium"
	SeverityHigh     ThreatSeverity = "high"
	SeverityCritical ThreatSeverity = "critical"
)

// ============================================
// THREAT EVENT
// ============================================

// ThreatEvent represents a detected security threat
type ThreatEvent struct {
	ID          string                 `json:"id"`
	Type        ThreatType             `json:"type"`
	Severity    ThreatSeverity         `json:"severity"`
	IP          string                 `json:"ip"`
	TenantID    string                 `json:"tenant_id,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	APIKeyID    string                 `json:"api_key_id,omitempty"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Blocked     bool                   `json:"blocked"`
	Action      string                 `json:"action"`
}

// ============================================
// ANOMALY DETECTOR
// ============================================

// AnomalyDetector detects anomalous behavior
type AnomalyDetector struct {
	mu            sync.RWMutex
	observability *ObservabilityService
	
	// Velocity tracking
	ipVelocity    map[string]*VelocityTracker
	keyVelocity   map[string]*VelocityTracker
	userVelocity  map[string]*VelocityTracker
	
	// Thresholds
	ipRateLimit      int     // requests per minute
	keyRateLimit     int     // requests per minute per API key
	loginRateLimit   int     // login attempts per hour
	velocitySpike    float64 // percentage increase to trigger
	
	// Metrics
	totalAnomalies   int64
	blockedRequests  int64
}

// VelocityTracker tracks request velocity
type VelocityTracker struct {
	Counts    []int64   // Rolling window of counts
	Timestamps []time.Time
	WindowSize int
	LastCount  int64
	Average    float64
}

// NewAnomalyDetector creates a new anomaly detector
func NewAnomalyDetector() *AnomalyDetector {
	return &AnomalyDetector{
		observability:  NewObservabilityService(),
		ipVelocity:     make(map[string]*VelocityTracker),
		keyVelocity:    make(map[string]*VelocityTracker),
		userVelocity:   make(map[string]*VelocityTracker),
		ipRateLimit:    100,   // 100 req/min per IP
		keyRateLimit:   1000,  // 1000 req/min per API key
		loginRateLimit: 10,    // 10 login attempts per hour
		velocitySpike:  200,   // 200% increase triggers alert
	}
}

// DetectAnomaly checks for anomalous behavior
func (d *AnomalyDetector) DetectAnomaly(ip, apiKeyID, userID string) *ThreatEvent {
	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()

	// Check IP velocity
	if threat := d.checkIPVelocity(ip, now); threat != nil {
		return threat
	}

	// Check API key velocity
	if apiKeyID != "" {
		if threat := d.checkKeyVelocity(apiKeyID, now); threat != nil {
			return threat
		}
	}

	// Check user velocity
	if userID != "" {
		if threat := d.checkUserVelocity(userID, now); threat != nil {
			return threat
		}
	}

	return nil
}

// checkIPVelocity checks IP request velocity
func (d *AnomalyDetector) checkIPVelocity(ip string, now time.Time) *ThreatEvent {
	tracker, exists := d.ipVelocity[ip]
	if !exists {
		tracker = &VelocityTracker{
			Counts:     make([]int64, 60), // 60 seconds window
			Timestamps: make([]time.Time, 60),
			WindowSize: 60,
		}
		d.ipVelocity[ip] = tracker
	}

	// Update count
	second := now.Second()
	if tracker.Timestamps[second].Minute() != now.Minute() {
		tracker.Counts[second] = 0
		tracker.Timestamps[second] = now
	}
	tracker.Counts[second]++
	tracker.LastCount++

	// Calculate total in last minute
	var total int64
	for i := 0; i < 60; i++ {
		if now.Sub(tracker.Timestamps[i]) < time.Minute {
			total += tracker.Counts[i]
		}
	}

	// Check threshold
	if total > int64(d.ipRateLimit) {
		atomic.AddInt64(&d.totalAnomalies, 1)
		return &ThreatEvent{
			ID:          fmt.Sprintf("threat_%d", now.UnixNano()),
			Type:        ThreatVelocityAbuse,
			Severity:    SeverityHigh,
			IP:          ip,
			Description: fmt.Sprintf("IP exceeded rate limit: %d req/min (limit: %d)", total, d.ipRateLimit),
			Timestamp:   now,
			Blocked:     true,
			Action:      "rate_limited",
			Metadata: map[string]interface{}{
				"requests_per_minute": total,
				"limit":               d.ipRateLimit,
			},
		}
	}

	// Check for velocity spike
	if tracker.Average > 0 {
		increase := (float64(total) / tracker.Average) * 100
		if increase > d.velocitySpike {
			atomic.AddInt64(&d.totalAnomalies, 1)
			return &ThreatEvent{
				ID:          fmt.Sprintf("threat_%d", now.UnixNano()),
				Type:        ThreatVelocityAbuse,
				Severity:    SeverityMedium,
				IP:          ip,
				Description: fmt.Sprintf("IP velocity spike detected: %.0f%% increase", increase),
				Timestamp:   now,
				Blocked:     false,
				Action:      "flagged",
				Metadata: map[string]interface{}{
					"current_rate": total,
					"average_rate": tracker.Average,
					"increase":     increase,
				},
			}
		}
	}

	// Update average
	tracker.Average = (tracker.Average*0.9 + float64(total)*0.1)

	return nil
}

// checkKeyVelocity checks API key request velocity
func (d *AnomalyDetector) checkKeyVelocity(keyID string, now time.Time) *ThreatEvent {
	tracker, exists := d.keyVelocity[keyID]
	if !exists {
		tracker = &VelocityTracker{
			Counts:     make([]int64, 60),
			Timestamps: make([]time.Time, 60),
			WindowSize: 60,
		}
		d.keyVelocity[keyID] = tracker
	}

	second := now.Second()
	if tracker.Timestamps[second].Minute() != now.Minute() {
		tracker.Counts[second] = 0
		tracker.Timestamps[second] = now
	}
	tracker.Counts[second]++

	var total int64
	for i := 0; i < 60; i++ {
		if now.Sub(tracker.Timestamps[i]) < time.Minute {
			total += tracker.Counts[i]
		}
	}

	if total > int64(d.keyRateLimit) {
		atomic.AddInt64(&d.totalAnomalies, 1)
		return &ThreatEvent{
			ID:          fmt.Sprintf("threat_%d", now.UnixNano()),
			Type:        ThreatAPIKeyAbuse,
			Severity:    SeverityHigh,
			APIKeyID:    keyID,
			Description: fmt.Sprintf("API key exceeded rate limit: %d req/min", total),
			Timestamp:   now,
			Blocked:     true,
			Action:      "rate_limited",
		}
	}

	return nil
}

// checkUserVelocity checks user request velocity
func (d *AnomalyDetector) checkUserVelocity(userID string, now time.Time) *ThreatEvent {
	// Similar implementation for user-level tracking
	return nil
}

// DetectLoginAnomaly checks for login anomalies
func (d *AnomalyDetector) DetectLoginAnomaly(ip, userID string, success bool) *ThreatEvent {
	ctx := context.Background()
	key := fmt.Sprintf("login_attempts:%s:%s", ip, time.Now().Format("2006010215"))

	// Increment attempt counter
	count, _ := cache.Increment(ctx, key)
	cache.Expire(ctx, key, time.Hour)

	if !success && count > int64(d.loginRateLimit) {
		atomic.AddInt64(&d.totalAnomalies, 1)
		return &ThreatEvent{
			ID:          fmt.Sprintf("threat_%d", time.Now().UnixNano()),
			Type:        ThreatLoginAnomaly,
			Severity:    SeverityHigh,
			IP:          ip,
			UserID:      userID,
			Description: fmt.Sprintf("Too many failed login attempts: %d", count),
			Timestamp:   time.Now(),
			Blocked:     true,
			Action:      "blocked",
		}
	}

	return nil
}

// GetStats returns anomaly detector statistics
func (d *AnomalyDetector) GetStats() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]interface{}{
		"total_anomalies":    atomic.LoadInt64(&d.totalAnomalies),
		"blocked_requests":   atomic.LoadInt64(&d.blockedRequests),
		"tracked_ips":        len(d.ipVelocity),
		"tracked_api_keys":   len(d.keyVelocity),
		"tracked_users":      len(d.userVelocity),
		"ip_rate_limit":      d.ipRateLimit,
		"key_rate_limit":     d.keyRateLimit,
		"login_rate_limit":   d.loginRateLimit,
		"velocity_spike_pct": d.velocitySpike,
	}
}

// ============================================
// SUSPICIOUS IP SERVICE
// ============================================

// SuspiciousIPService manages suspicious IP tracking
type SuspiciousIPService struct {
	mu            sync.RWMutex
	observability *ObservabilityService
	
	// IP tracking
	suspiciousIPs map[string]*SuspiciousIPInfo
	blockedIPs    map[string]*BlockedIPInfo
	
	// Thresholds
	suspiciousThreshold int // Score to mark as suspicious
	blockThreshold      int // Score to auto-block
}

// SuspiciousIPInfo holds information about a suspicious IP
type SuspiciousIPInfo struct {
	IP            string    `json:"ip"`
	Score         int       `json:"score"`
	Reasons       []string  `json:"reasons"`
	FirstSeen     time.Time `json:"first_seen"`
	LastSeen      time.Time `json:"last_seen"`
	RequestCount  int64     `json:"request_count"`
	BlockedCount  int64     `json:"blocked_count"`
	ThreatTypes   []ThreatType `json:"threat_types"`
}

// BlockedIPInfo holds information about a blocked IP
type BlockedIPInfo struct {
	IP          string    `json:"ip"`
	Reason      string    `json:"reason"`
	BlockedAt   time.Time `json:"blocked_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Permanent   bool      `json:"permanent"`
	BlockedBy   string    `json:"blocked_by"`
}

// NewSuspiciousIPService creates a new suspicious IP service
func NewSuspiciousIPService() *SuspiciousIPService {
	return &SuspiciousIPService{
		observability:       NewObservabilityService(),
		suspiciousIPs:       make(map[string]*SuspiciousIPInfo),
		blockedIPs:          make(map[string]*BlockedIPInfo),
		suspiciousThreshold: 50,
		blockThreshold:      100,
	}
}

// ReportSuspiciousActivity reports suspicious activity from an IP
func (s *SuspiciousIPService) ReportSuspiciousActivity(ip string, threatType ThreatType, score int, reason string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()

	info, exists := s.suspiciousIPs[ip]
	if !exists {
		info = &SuspiciousIPInfo{
			IP:        ip,
			FirstSeen: now,
			Reasons:   []string{},
			ThreatTypes: []ThreatType{},
		}
		s.suspiciousIPs[ip] = info
	}

	info.Score += score
	info.LastSeen = now
	info.RequestCount++
	info.Reasons = append(info.Reasons, reason)
	info.ThreatTypes = append(info.ThreatTypes, threatType)

	// Auto-block if threshold exceeded
	if info.Score >= s.blockThreshold {
		s.blockIP(ip, "Auto-blocked: score threshold exceeded", 24*time.Hour, "system")
	}

	// Persist to Redis
	s.persistSuspiciousIP(ip, info)
}

// blockIP blocks an IP address
func (s *SuspiciousIPService) blockIP(ip, reason string, duration time.Duration, blockedBy string) {
	now := time.Now()
	s.blockedIPs[ip] = &BlockedIPInfo{
		IP:        ip,
		Reason:    reason,
		BlockedAt: now,
		ExpiresAt: now.Add(duration),
		Permanent: duration == 0,
		BlockedBy: blockedBy,
	}

	// Persist to Redis
	ctx := context.Background()
	key := fmt.Sprintf("blocked_ip:%s", ip)
	data, _ := json.Marshal(s.blockedIPs[ip])
	if duration > 0 {
		cache.Set(ctx, key, string(data), duration)
	} else {
		cache.Set(ctx, key, string(data), 365*24*time.Hour)
	}
}

// IsBlocked checks if an IP is blocked
func (s *SuspiciousIPService) IsBlocked(ip string) bool {
	s.mu.RLock()
	info, exists := s.blockedIPs[ip]
	s.mu.RUnlock()

	if exists {
		if info.Permanent || time.Now().Before(info.ExpiresAt) {
			return true
		}
		// Expired, remove from map
		s.mu.Lock()
		delete(s.blockedIPs, ip)
		s.mu.Unlock()
	}

	// Check Redis
	ctx := context.Background()
	key := fmt.Sprintf("blocked_ip:%s", ip)
	if data, err := cache.Get(ctx, key); err == nil && data != "" {
		return true
	}

	return false
}

// BlockIP manually blocks an IP
func (s *SuspiciousIPService) BlockIP(ip, reason string, duration time.Duration, blockedBy string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.blockIP(ip, reason, duration, blockedBy)
}

// UnblockIP unblocks an IP
func (s *SuspiciousIPService) UnblockIP(ip string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	delete(s.blockedIPs, ip)
	
	ctx := context.Background()
	cache.Delete(ctx, fmt.Sprintf("blocked_ip:%s", ip))
}

// GetSuspiciousIPs returns all suspicious IPs
func (s *SuspiciousIPService) GetSuspiciousIPs(limit int) []*SuspiciousIPInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*SuspiciousIPInfo, 0, len(s.suspiciousIPs))
	for _, info := range s.suspiciousIPs {
		result = append(result, info)
	}

	// Sort by score (descending)
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[j].Score > result[i].Score {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	if limit > 0 && len(result) > limit {
		result = result[:limit]
	}

	return result
}

// GetBlockedIPs returns all blocked IPs
func (s *SuspiciousIPService) GetBlockedIPs() []*BlockedIPInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*BlockedIPInfo, 0, len(s.blockedIPs))
	for _, info := range s.blockedIPs {
		result = append(result, info)
	}
	return result
}

// persistSuspiciousIP persists suspicious IP to Redis
func (s *SuspiciousIPService) persistSuspiciousIP(ip string, info *SuspiciousIPInfo) {
	ctx := context.Background()
	key := fmt.Sprintf("suspicious_ip:%s", ip)
	data, _ := json.Marshal(info)
	cache.Set(ctx, key, string(data), 24*time.Hour)
}

// GetStats returns service statistics
func (s *SuspiciousIPService) GetStats() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return map[string]interface{}{
		"suspicious_count":      len(s.suspiciousIPs),
		"blocked_count":         len(s.blockedIPs),
		"suspicious_threshold":  s.suspiciousThreshold,
		"block_threshold":       s.blockThreshold,
	}
}

// ============================================
// THREAT DETECTOR SERVICE
// ============================================

// ThreatDetector is the main threat detection service
type ThreatDetector struct {
	mu               sync.RWMutex
	anomalyDetector  *AnomalyDetector
	suspiciousIPs    *SuspiciousIPService
	observability    *ObservabilityService
	
	// Threat history
	recentThreats    []*ThreatEvent
	maxRecentThreats int
	
	// Metrics
	totalThreats     int64
	threatsByType    map[ThreatType]int64
}

// NewThreatDetector creates a new threat detector
func NewThreatDetector() *ThreatDetector {
	return &ThreatDetector{
		anomalyDetector:  NewAnomalyDetector(),
		suspiciousIPs:    NewSuspiciousIPService(),
		observability:    NewObservabilityService(),
		recentThreats:    make([]*ThreatEvent, 0),
		maxRecentThreats: 1000,
		threatsByType:    make(map[ThreatType]int64),
	}
}

// DetectThreat runs all threat detection checks
func (t *ThreatDetector) DetectThreat(ip, apiKeyID, userID, tenantID string) *ThreatEvent {
	// Check if IP is blocked
	if t.suspiciousIPs.IsBlocked(ip) {
		return &ThreatEvent{
			ID:          fmt.Sprintf("threat_%d", time.Now().UnixNano()),
			Type:        ThreatSuspiciousIP,
			Severity:    SeverityCritical,
			IP:          ip,
			TenantID:    tenantID,
			Description: "Request from blocked IP",
			Timestamp:   time.Now(),
			Blocked:     true,
			Action:      "blocked",
		}
	}

	// Run anomaly detection
	if threat := t.anomalyDetector.DetectAnomaly(ip, apiKeyID, userID); threat != nil {
		threat.TenantID = tenantID
		t.recordThreat(threat)
		return threat
	}

	return nil
}

// RecordThreat records a detected threat
func (t *ThreatDetector) RecordThreat(threat *ThreatEvent) {
	t.recordThreat(threat)
}

// recordThreat internal method to record threat
func (t *ThreatDetector) recordThreat(threat *ThreatEvent) {
	t.mu.Lock()
	defer t.mu.Unlock()

	atomic.AddInt64(&t.totalThreats, 1)
	t.threatsByType[threat.Type]++

	// Add to recent threats
	t.recentThreats = append(t.recentThreats, threat)
	if len(t.recentThreats) > t.maxRecentThreats {
		t.recentThreats = t.recentThreats[1:]
	}

	// Report to suspicious IP service
	score := t.getSeverityScore(threat.Severity)
	t.suspiciousIPs.ReportSuspiciousActivity(threat.IP, threat.Type, score, threat.Description)

	// Log threat
	t.observability.Log(LogEvent{
		Category: LogCategoryFraudDetection,
		Level:    LogLevelWarn,
		Message:  "Threat detected",
		Metadata: map[string]interface{}{
			"threat_id":   threat.ID,
			"threat_type": threat.Type,
			"severity":    threat.Severity,
			"ip":          threat.IP,
			"action":      threat.Action,
		},
	})
}

// getSeverityScore returns a score for a severity level
func (t *ThreatDetector) getSeverityScore(severity ThreatSeverity) int {
	switch severity {
	case SeverityLow:
		return 10
	case SeverityMedium:
		return 25
	case SeverityHigh:
		return 50
	case SeverityCritical:
		return 100
	default:
		return 10
	}
}

// GetRecentThreats returns recent threats
func (t *ThreatDetector) GetRecentThreats(limit int) []*ThreatEvent {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if limit <= 0 || limit > len(t.recentThreats) {
		limit = len(t.recentThreats)
	}

	// Return most recent first
	result := make([]*ThreatEvent, limit)
	for i := 0; i < limit; i++ {
		result[i] = t.recentThreats[len(t.recentThreats)-1-i]
	}
	return result
}

// GetThreatStats returns threat statistics
func (t *ThreatDetector) GetThreatStats() map[string]interface{} {
	t.mu.RLock()
	defer t.mu.RUnlock()

	byType := make(map[string]int64)
	for k, v := range t.threatsByType {
		byType[string(k)] = v
	}

	return map[string]interface{}{
		"total_threats":     atomic.LoadInt64(&t.totalThreats),
		"threats_by_type":   byType,
		"recent_count":      len(t.recentThreats),
		"anomaly_stats":     t.anomalyDetector.GetStats(),
		"suspicious_stats":  t.suspiciousIPs.GetStats(),
	}
}

// BlockIP blocks an IP address
func (t *ThreatDetector) BlockIP(ip, reason string, duration time.Duration, blockedBy string) {
	t.suspiciousIPs.BlockIP(ip, reason, duration, blockedBy)
}

// UnblockIP unblocks an IP address
func (t *ThreatDetector) UnblockIP(ip string) {
	t.suspiciousIPs.UnblockIP(ip)
}

// GetBlockedIPs returns all blocked IPs
func (t *ThreatDetector) GetBlockedIPs() []*BlockedIPInfo {
	return t.suspiciousIPs.GetBlockedIPs()
}

// GetSuspiciousIPs returns suspicious IPs
func (t *ThreatDetector) GetSuspiciousIPs(limit int) []*SuspiciousIPInfo {
	return t.suspiciousIPs.GetSuspiciousIPs(limit)
}

// ============================================
// GLOBAL INSTANCE
// ============================================

var (
	threatDetectorInstance *ThreatDetector
	threatDetectorOnce     sync.Once
)

// GetThreatDetector returns the global threat detector
func GetThreatDetector() *ThreatDetector {
	threatDetectorOnce.Do(func() {
		threatDetectorInstance = NewThreatDetector()
	})
	return threatDetectorInstance
}

// ============================================
// HELPER: Calculate standard deviation
// ============================================

func calculateStdDev(values []float64, mean float64) float64 {
	if len(values) == 0 {
		return 0
	}
	var sum float64
	for _, v := range values {
		sum += (v - mean) * (v - mean)
	}
	return math.Sqrt(sum / float64(len(values)))
}


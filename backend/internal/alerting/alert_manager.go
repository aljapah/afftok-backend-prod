package alerting

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================
// ALERT TYPES
// ============================================

// AlertSeverity represents alert severity levels
type AlertSeverity string

const (
	AlertSeverityInfo     AlertSeverity = "info"
	AlertSeverityWarning  AlertSeverity = "warning"
	AlertSeverityError    AlertSeverity = "error"
	AlertSeverityCritical AlertSeverity = "critical"
)

// AlertType represents different alert types
type AlertType string

const (
	AlertDBLatency        AlertType = "db_latency"
	AlertRedisLatency     AlertType = "redis_latency"
	AlertDroppedClicks    AlertType = "dropped_clicks"
	AlertWALPending       AlertType = "wal_pending"
	AlertCPUHigh          AlertType = "cpu_high"
	AlertMemoryHigh       AlertType = "memory_high"
	AlertBotSpike         AlertType = "bot_spike"
	AlertGeoBlockSpike    AlertType = "geo_block_spike"
	AlertAPIKeyBruteForce AlertType = "api_key_brute_force"
	AlertIngestionBacklog AlertType = "ingestion_backlog"
	AlertEdgeDisconnect   AlertType = "edge_disconnect"
	AlertPostbackRetries  AlertType = "postback_retries"
	AlertSystemHealth     AlertType = "system_health"
)

// ============================================
// ALERT
// ============================================

// Alert represents a single alert
type Alert struct {
	ID          string                 `json:"id"`
	Type        AlertType              `json:"type"`
	Severity    AlertSeverity          `json:"severity"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Value       interface{}            `json:"value,omitempty"`
	Threshold   interface{}            `json:"threshold,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Timestamp   time.Time              `json:"timestamp"`
	Resolved    bool                   `json:"resolved"`
	ResolvedAt  *time.Time             `json:"resolved_at,omitempty"`
	Acknowledged bool                  `json:"acknowledged"`
}

// ============================================
// ALERT THRESHOLDS
// ============================================

// AlertThresholds defines thresholds for different metrics
type AlertThresholds struct {
	DBLatencyMs         int     `json:"db_latency_ms"`
	RedisLatencyMs      int     `json:"redis_latency_ms"`
	DroppedClicksMax    int     `json:"dropped_clicks_max"`
	WALPendingMax       int     `json:"wal_pending_max"`
	CPUPercentMax       float64 `json:"cpu_percent_max"`
	MemoryPercentMax    float64 `json:"memory_percent_max"`
	BotSpikePercent     float64 `json:"bot_spike_percent"`
	GeoBlockSpikePercent float64 `json:"geo_block_spike_percent"`
	IngestionBacklogMax int     `json:"ingestion_backlog_max"`
	PostbackRetriesMax  int     `json:"postback_retries_max"`
}

// DefaultThresholds returns default alert thresholds
func DefaultThresholds() *AlertThresholds {
	return &AlertThresholds{
		DBLatencyMs:         30,
		RedisLatencyMs:      10,
		DroppedClicksMax:    0,
		WALPendingMax:       1000,
		CPUPercentMax:       80,
		MemoryPercentMax:    80,
		BotSpikePercent:     200,
		GeoBlockSpikePercent: 200,
		IngestionBacklogMax: 10000,
		PostbackRetriesMax:  100,
	}
}

// ============================================
// ALERT MANAGER
// ============================================

// AlertManager manages alerts and notifications
type AlertManager struct {
	mu           sync.RWMutex
	thresholds   *AlertThresholds
	channels     []AlertChannel
	activeAlerts map[string]*Alert
	alertHistory []*Alert
	maxHistory   int
	
	// Cooldowns to prevent alert spam
	cooldowns    map[AlertType]time.Time
	cooldownDuration time.Duration
	
	// Metrics
	totalAlerts  int64
	sentAlerts   int64
	failedAlerts int64
	
	// State
	enabled      bool
}

// AlertChannel interface for notification channels
type AlertChannel interface {
	Send(alert *Alert) error
	Name() string
	IsEnabled() bool
}

// NewAlertManager creates a new alert manager
func NewAlertManager() *AlertManager {
	return &AlertManager{
		thresholds:       DefaultThresholds(),
		channels:         make([]AlertChannel, 0),
		activeAlerts:     make(map[string]*Alert),
		alertHistory:     make([]*Alert, 0),
		maxHistory:       1000,
		cooldowns:        make(map[AlertType]time.Time),
		cooldownDuration: 5 * time.Minute,
		enabled:          true,
	}
}

// RegisterChannel registers a notification channel
func (m *AlertManager) RegisterChannel(channel AlertChannel) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.channels = append(m.channels, channel)
}

// SetThresholds updates alert thresholds
func (m *AlertManager) SetThresholds(thresholds *AlertThresholds) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.thresholds = thresholds
}

// GetThresholds returns current thresholds
func (m *AlertManager) GetThresholds() *AlertThresholds {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.thresholds
}

// Enable enables the alert manager
func (m *AlertManager) Enable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = true
}

// Disable disables the alert manager
func (m *AlertManager) Disable() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.enabled = false
}

// ============================================
// ALERT CREATION
// ============================================

// CreateAlert creates and sends an alert
func (m *AlertManager) CreateAlert(alertType AlertType, severity AlertSeverity, title, message string, value, threshold interface{}, metadata map[string]interface{}) {
	m.mu.Lock()
	
	if !m.enabled {
		m.mu.Unlock()
		return
	}

	// Check cooldown
	if lastSent, exists := m.cooldowns[alertType]; exists {
		if time.Since(lastSent) < m.cooldownDuration {
			m.mu.Unlock()
			return
		}
	}

	alert := &Alert{
		ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Type:      alertType,
		Severity:  severity,
		Title:     title,
		Message:   message,
		Value:     value,
		Threshold: threshold,
		Metadata:  metadata,
		Timestamp: time.Now(),
	}

	// Add to active alerts
	m.activeAlerts[alert.ID] = alert
	
	// Add to history
	m.alertHistory = append(m.alertHistory, alert)
	if len(m.alertHistory) > m.maxHistory {
		m.alertHistory = m.alertHistory[1:]
	}

	// Update cooldown
	m.cooldowns[alertType] = time.Now()
	
	atomic.AddInt64(&m.totalAlerts, 1)
	
	m.mu.Unlock()

	// Send to all channels
	m.sendToChannels(alert)
}

// sendToChannels sends alert to all registered channels
func (m *AlertManager) sendToChannels(alert *Alert) {
	m.mu.RLock()
	channels := make([]AlertChannel, len(m.channels))
	copy(channels, m.channels)
	m.mu.RUnlock()

	for _, channel := range channels {
		if channel.IsEnabled() {
			go func(ch AlertChannel) {
				if err := ch.Send(alert); err != nil {
					atomic.AddInt64(&m.failedAlerts, 1)
				} else {
					atomic.AddInt64(&m.sentAlerts, 1)
				}
			}(channel)
		}
	}
}

// ResolveAlert marks an alert as resolved
func (m *AlertManager) ResolveAlert(alertID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if alert, exists := m.activeAlerts[alertID]; exists {
		now := time.Now()
		alert.Resolved = true
		alert.ResolvedAt = &now
		delete(m.activeAlerts, alertID)
	}
}

// AcknowledgeAlert acknowledges an alert
func (m *AlertManager) AcknowledgeAlert(alertID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if alert, exists := m.activeAlerts[alertID]; exists {
		alert.Acknowledged = true
	}
}

// ============================================
// METRIC CHECKS
// ============================================

// CheckDBLatency checks database latency
func (m *AlertManager) CheckDBLatency(latencyMs int) {
	if latencyMs > m.thresholds.DBLatencyMs {
		m.CreateAlert(
			AlertDBLatency,
			AlertSeverityWarning,
			"High Database Latency",
			fmt.Sprintf("Database latency is %dms (threshold: %dms)", latencyMs, m.thresholds.DBLatencyMs),
			latencyMs,
			m.thresholds.DBLatencyMs,
			nil,
		)
	}
}

// CheckRedisLatency checks Redis latency
func (m *AlertManager) CheckRedisLatency(latencyMs int) {
	if latencyMs > m.thresholds.RedisLatencyMs {
		m.CreateAlert(
			AlertRedisLatency,
			AlertSeverityWarning,
			"High Redis Latency",
			fmt.Sprintf("Redis latency is %dms (threshold: %dms)", latencyMs, m.thresholds.RedisLatencyMs),
			latencyMs,
			m.thresholds.RedisLatencyMs,
			nil,
		)
	}
}

// CheckDroppedClicks checks for dropped clicks
func (m *AlertManager) CheckDroppedClicks(count int) {
	if count > m.thresholds.DroppedClicksMax {
		m.CreateAlert(
			AlertDroppedClicks,
			AlertSeverityCritical,
			"🚨 DROPPED CLICKS DETECTED",
			fmt.Sprintf("Dropped clicks: %d (MUST BE 0)", count),
			count,
			m.thresholds.DroppedClicksMax,
			nil,
		)
	}
}

// CheckWALPending checks WAL pending entries
func (m *AlertManager) CheckWALPending(count int) {
	if count > m.thresholds.WALPendingMax {
		m.CreateAlert(
			AlertWALPending,
			AlertSeverityWarning,
			"High WAL Pending Entries",
			fmt.Sprintf("WAL pending: %d (threshold: %d)", count, m.thresholds.WALPendingMax),
			count,
			m.thresholds.WALPendingMax,
			nil,
		)
	}
}

// CheckCPU checks CPU usage
func (m *AlertManager) CheckCPU(percent float64) {
	if percent > m.thresholds.CPUPercentMax {
		m.CreateAlert(
			AlertCPUHigh,
			AlertSeverityWarning,
			"High CPU Usage",
			fmt.Sprintf("CPU usage: %.1f%% (threshold: %.1f%%)", percent, m.thresholds.CPUPercentMax),
			percent,
			m.thresholds.CPUPercentMax,
			nil,
		)
	}
}

// CheckMemory checks memory usage
func (m *AlertManager) CheckMemory(percent float64) {
	if percent > m.thresholds.MemoryPercentMax {
		m.CreateAlert(
			AlertMemoryHigh,
			AlertSeverityWarning,
			"High Memory Usage",
			fmt.Sprintf("Memory usage: %.1f%% (threshold: %.1f%%)", percent, m.thresholds.MemoryPercentMax),
			percent,
			m.thresholds.MemoryPercentMax,
			nil,
		)
	}
}

// CheckBotSpike checks for bot traffic spike
func (m *AlertManager) CheckBotSpike(increasePercent float64) {
	if increasePercent > m.thresholds.BotSpikePercent {
		m.CreateAlert(
			AlertBotSpike,
			AlertSeverityError,
			"Bot Traffic Spike Detected",
			fmt.Sprintf("Bot traffic increased by %.1f%% (threshold: %.1f%%)", increasePercent, m.thresholds.BotSpikePercent),
			increasePercent,
			m.thresholds.BotSpikePercent,
			nil,
		)
	}
}

// CheckIngestionBacklog checks ingestion backlog
func (m *AlertManager) CheckIngestionBacklog(count int) {
	if count > m.thresholds.IngestionBacklogMax {
		m.CreateAlert(
			AlertIngestionBacklog,
			AlertSeverityWarning,
			"High Ingestion Backlog",
			fmt.Sprintf("Ingestion backlog: %d (threshold: %d)", count, m.thresholds.IngestionBacklogMax),
			count,
			m.thresholds.IngestionBacklogMax,
			nil,
		)
	}
}

// ============================================
// GETTERS
// ============================================

// GetActiveAlerts returns all active alerts
func (m *AlertManager) GetActiveAlerts() []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Alert, 0, len(m.activeAlerts))
	for _, alert := range m.activeAlerts {
		result = append(result, alert)
	}
	return result
}

// GetAlertHistory returns alert history
func (m *AlertManager) GetAlertHistory(limit int) []*Alert {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.alertHistory) {
		limit = len(m.alertHistory)
	}

	result := make([]*Alert, limit)
	for i := 0; i < limit; i++ {
		result[i] = m.alertHistory[len(m.alertHistory)-1-i]
	}
	return result
}

// GetStats returns alert manager statistics
func (m *AlertManager) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return map[string]interface{}{
		"enabled":        m.enabled,
		"total_alerts":   atomic.LoadInt64(&m.totalAlerts),
		"sent_alerts":    atomic.LoadInt64(&m.sentAlerts),
		"failed_alerts":  atomic.LoadInt64(&m.failedAlerts),
		"active_count":   len(m.activeAlerts),
		"history_count":  len(m.alertHistory),
		"channels_count": len(m.channels),
		"thresholds":     m.thresholds,
	}
}

// ============================================
// SLACK CHANNEL
// ============================================

// SlackChannel sends alerts to Slack
type SlackChannel struct {
	webhookURL string
	channel    string
	enabled    bool
}

// NewSlackChannel creates a new Slack channel
func NewSlackChannel() *SlackChannel {
	return &SlackChannel{
		webhookURL: os.Getenv("SLACK_WEBHOOK_URL"),
		channel:    os.Getenv("SLACK_CHANNEL"),
		enabled:    os.Getenv("SLACK_ALERTS_ENABLED") == "true",
	}
}

// Name returns the channel name
func (s *SlackChannel) Name() string {
	return "slack"
}

// IsEnabled returns whether the channel is enabled
func (s *SlackChannel) IsEnabled() bool {
	return s.enabled && s.webhookURL != ""
}

// Send sends an alert to Slack
func (s *SlackChannel) Send(alert *Alert) error {
	if !s.IsEnabled() {
		return nil
	}

	emoji := s.getEmoji(alert.Severity)
	color := s.getColor(alert.Severity)

	payload := map[string]interface{}{
		"channel": s.channel,
		"attachments": []map[string]interface{}{
			{
				"color":  color,
				"title":  fmt.Sprintf("%s %s", emoji, alert.Title),
				"text":   alert.Message,
				"fields": []map[string]interface{}{
					{"title": "Severity", "value": string(alert.Severity), "short": true},
					{"title": "Type", "value": string(alert.Type), "short": true},
				},
				"footer": "AffTok Alert System",
				"ts":     alert.Timestamp.Unix(),
			},
		},
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(s.webhookURL, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack error: %s", string(body))
	}

	return nil
}

func (s *SlackChannel) getEmoji(severity AlertSeverity) string {
	switch severity {
	case AlertSeverityInfo:
		return "ℹ️"
	case AlertSeverityWarning:
		return "⚠️"
	case AlertSeverityError:
		return "🔴"
	case AlertSeverityCritical:
		return "🚨"
	default:
		return "📢"
	}
}

func (s *SlackChannel) getColor(severity AlertSeverity) string {
	switch severity {
	case AlertSeverityInfo:
		return "#36a64f"
	case AlertSeverityWarning:
		return "#ffcc00"
	case AlertSeverityError:
		return "#ff6600"
	case AlertSeverityCritical:
		return "#ff0000"
	default:
		return "#808080"
	}
}

// ============================================
// TELEGRAM CHANNEL
// ============================================

// TelegramChannel sends alerts to Telegram
type TelegramChannel struct {
	botToken string
	chatID   string
	enabled  bool
}

// NewTelegramChannel creates a new Telegram channel
func NewTelegramChannel() *TelegramChannel {
	return &TelegramChannel{
		botToken: os.Getenv("TELEGRAM_BOT_TOKEN"),
		chatID:   os.Getenv("TELEGRAM_CHAT_ID"),
		enabled:  os.Getenv("TELEGRAM_ALERTS_ENABLED") == "true",
	}
}

// Name returns the channel name
func (t *TelegramChannel) Name() string {
	return "telegram"
}

// IsEnabled returns whether the channel is enabled
func (t *TelegramChannel) IsEnabled() bool {
	return t.enabled && t.botToken != "" && t.chatID != ""
}

// Send sends an alert to Telegram
func (t *TelegramChannel) Send(alert *Alert) error {
	if !t.IsEnabled() {
		return nil
	}

	emoji := t.getEmoji(alert.Severity)
	message := fmt.Sprintf(
		"%s *%s*\n\n%s\n\n⏰ %s\n🏷 %s | %s",
		emoji,
		alert.Title,
		alert.Message,
		alert.Timestamp.Format("2006-01-02 15:04:05"),
		alert.Severity,
		alert.Type,
	)

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	
	payload := map[string]interface{}{
		"chat_id":    t.chatID,
		"text":       message,
		"parse_mode": "Markdown",
	}

	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("telegram error: %s", string(body))
	}

	return nil
}

func (t *TelegramChannel) getEmoji(severity AlertSeverity) string {
	switch severity {
	case AlertSeverityInfo:
		return "ℹ️"
	case AlertSeverityWarning:
		return "⚠️"
	case AlertSeverityError:
		return "🔴"
	case AlertSeverityCritical:
		return "🚨"
	default:
		return "📢"
	}
}

// ============================================
// GLOBAL INSTANCE
// ============================================

var (
	alertManagerInstance *AlertManager
	alertManagerOnce     sync.Once
)

// GetAlertManager returns the global alert manager
func GetAlertManager() *AlertManager {
	alertManagerOnce.Do(func() {
		alertManagerInstance = NewAlertManager()
		
		// Register default channels
		alertManagerInstance.RegisterChannel(NewSlackChannel())
		alertManagerInstance.RegisterChannel(NewTelegramChannel())
	})
	return alertManagerInstance
}


package services

import (
	"sync"
	"sync/atomic"
	"time"
)

// ============================================
// LOGGING MODES
// ============================================

// LoggingMode represents the current logging mode
type LoggingMode string

const (
	LogModeNormal   LoggingMode = "normal"   // Minimal info, optimized for low cost
	LogModeVerbose  LoggingMode = "verbose"  // Detailed click & routing logs
	LogModeCritical LoggingMode = "critical" // Full tracing, auto-rollback after 10 min
)

// LoggingModeConfig holds configuration for each mode
type LoggingModeConfig struct {
	Mode              LoggingMode
	EnableWALTracing  bool
	EnablePostbackLog bool
	EnableBotDetect   bool
	EnableGeoRuleLog  bool
	EnableFraudLog    bool
	EnableClickLog    bool
	EnableRoutingLog  bool
	EnableAPILog      bool
	LogLevel          string
}

// ============================================
// LOGGING MODE SERVICE
// ============================================

// LoggingModeService manages logging modes
type LoggingModeService struct {
	mu               sync.RWMutex
	currentMode      LoggingMode
	config           *LoggingModeConfig
	criticalTimeout  *time.Timer
	criticalStarted  time.Time
	autoRollbackMins int
	
	// Metrics
	modeChanges      int64
	criticalCount    int64
	lastModeChange   time.Time
	
	observability    *ObservabilityService
}

// NewLoggingModeService creates a new logging mode service
func NewLoggingModeService() *LoggingModeService {
	service := &LoggingModeService{
		currentMode:      LogModeNormal,
		autoRollbackMins: 10,
		observability:    NewObservabilityService(),
	}
	service.config = service.getConfigForMode(LogModeNormal)
	return service
}

// ============================================
// MODE CONFIGURATIONS
// ============================================

// getConfigForMode returns the configuration for a specific mode
func (s *LoggingModeService) getConfigForMode(mode LoggingMode) *LoggingModeConfig {
	switch mode {
	case LogModeNormal:
		return &LoggingModeConfig{
			Mode:              LogModeNormal,
			EnableWALTracing:  false,
			EnablePostbackLog: false,
			EnableBotDetect:   false,
			EnableGeoRuleLog:  false,
			EnableFraudLog:    true, // Always log fraud
			EnableClickLog:    false,
			EnableRoutingLog:  false,
			EnableAPILog:      false,
			LogLevel:          LogLevelInfo,
		}
	
	case LogModeVerbose:
		return &LoggingModeConfig{
			Mode:              LogModeVerbose,
			EnableWALTracing:  false,
			EnablePostbackLog: true,
			EnableBotDetect:   true,
			EnableGeoRuleLog:  true,
			EnableFraudLog:    true,
			EnableClickLog:    true,
			EnableRoutingLog:  true,
			EnableAPILog:      true,
			LogLevel:          LogLevelDebug,
		}
	
	case LogModeCritical:
		return &LoggingModeConfig{
			Mode:              LogModeCritical,
			EnableWALTracing:  true,
			EnablePostbackLog: true,
			EnableBotDetect:   true,
			EnableGeoRuleLog:  true,
			EnableFraudLog:    true,
			EnableClickLog:    true,
			EnableRoutingLog:  true,
			EnableAPILog:      true,
			LogLevel:          LogLevelDebug,
		}
	
	default:
		return s.getConfigForMode(LogModeNormal)
	}
}

// ============================================
// MODE OPERATIONS
// ============================================

// SetMode sets the logging mode
func (s *LoggingModeService) SetMode(mode LoggingMode) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Cancel any existing critical timeout
	if s.criticalTimeout != nil {
		s.criticalTimeout.Stop()
		s.criticalTimeout = nil
	}

	oldMode := s.currentMode
	s.currentMode = mode
	s.config = s.getConfigForMode(mode)
	s.lastModeChange = time.Now()
	atomic.AddInt64(&s.modeChanges, 1)

	// Set auto-rollback for critical mode
	if mode == LogModeCritical {
		s.criticalStarted = time.Now()
		atomic.AddInt64(&s.criticalCount, 1)
		
		s.criticalTimeout = time.AfterFunc(time.Duration(s.autoRollbackMins)*time.Minute, func() {
			s.SetMode(LogModeNormal)
			s.observability.Log(LogEvent{
				Category: LogCategorySystemEvent,
				Level:    LogLevelInfo,
				Message:  "Auto-rollback from CRITICAL to NORMAL mode",
				Metadata: map[string]interface{}{
					"duration_mins": s.autoRollbackMins,
				},
			})
		})
	}

	s.observability.Log(LogEvent{
		Category: LogCategorySystemEvent,
		Level:    LogLevelInfo,
		Message:  "Logging mode changed",
		Metadata: map[string]interface{}{
			"old_mode": oldMode,
			"new_mode": mode,
		},
	})
}

// GetMode returns the current logging mode
func (s *LoggingModeService) GetMode() LoggingMode {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.currentMode
}

// GetConfig returns the current logging configuration
func (s *LoggingModeService) GetConfig() *LoggingModeConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// SetAutoRollbackMinutes sets the auto-rollback time for critical mode
func (s *LoggingModeService) SetAutoRollbackMinutes(mins int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.autoRollbackMins = mins
}

// ============================================
// LOGGING HELPERS
// ============================================

// ShouldLogClick returns true if click logging is enabled
func (s *LoggingModeService) ShouldLogClick() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.EnableClickLog
}

// ShouldLogPostback returns true if postback logging is enabled
func (s *LoggingModeService) ShouldLogPostback() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.EnablePostbackLog
}

// ShouldLogWAL returns true if WAL tracing is enabled
func (s *LoggingModeService) ShouldLogWAL() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.EnableWALTracing
}

// ShouldLogBotDetection returns true if bot detection logging is enabled
func (s *LoggingModeService) ShouldLogBotDetection() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.EnableBotDetect
}

// ShouldLogGeoRule returns true if geo rule logging is enabled
func (s *LoggingModeService) ShouldLogGeoRule() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.EnableGeoRuleLog
}

// ShouldLogFraud returns true if fraud logging is enabled
func (s *LoggingModeService) ShouldLogFraud() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.EnableFraudLog
}

// ShouldLogRouting returns true if routing logging is enabled
func (s *LoggingModeService) ShouldLogRouting() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.EnableRoutingLog
}

// ShouldLogAPI returns true if API logging is enabled
func (s *LoggingModeService) ShouldLogAPI() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.EnableAPILog
}

// GetLogLevel returns the current log level
func (s *LoggingModeService) GetLogLevel() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config.LogLevel
}

// ============================================
// STATUS
// ============================================

// GetStatus returns the current logging mode status
func (s *LoggingModeService) GetStatus() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := map[string]interface{}{
		"current_mode":       s.currentMode,
		"mode_changes":       atomic.LoadInt64(&s.modeChanges),
		"critical_count":     atomic.LoadInt64(&s.criticalCount),
		"last_mode_change":   s.lastModeChange,
		"auto_rollback_mins": s.autoRollbackMins,
		"config": map[string]interface{}{
			"wal_tracing":   s.config.EnableWALTracing,
			"postback_log":  s.config.EnablePostbackLog,
			"bot_detect":    s.config.EnableBotDetect,
			"geo_rule_log":  s.config.EnableGeoRuleLog,
			"fraud_log":     s.config.EnableFraudLog,
			"click_log":     s.config.EnableClickLog,
			"routing_log":   s.config.EnableRoutingLog,
			"api_log":       s.config.EnableAPILog,
			"log_level":     s.config.LogLevel,
		},
	}

	// Add critical mode info if active
	if s.currentMode == LogModeCritical {
		elapsed := time.Since(s.criticalStarted)
		remaining := time.Duration(s.autoRollbackMins)*time.Minute - elapsed
		status["critical_elapsed_seconds"] = int(elapsed.Seconds())
		status["critical_remaining_seconds"] = int(remaining.Seconds())
	}

	return status
}

// ============================================
// GLOBAL INSTANCE
// ============================================

var (
	loggingModeInstance *LoggingModeService
	loggingModeOnce     sync.Once
)

// GetLoggingModeService returns the global logging mode service
func GetLoggingModeService() *LoggingModeService {
	loggingModeOnce.Do(func() {
		loggingModeInstance = NewLoggingModeService()
	})
	return loggingModeInstance
}


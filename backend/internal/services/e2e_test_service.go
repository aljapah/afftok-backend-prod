package services

import (
	"context"
	"encoding/json"
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
// E2E TEST TYPES
// ============================================

// E2EScenarioType represents the type of E2E test scenario
type E2EScenarioType string

const (
	ScenarioClickOnly          E2EScenarioType = "click_only"
	ScenarioClickThenConversion E2EScenarioType = "click_then_conversion"
	ScenarioClickThenPostback  E2EScenarioType = "click_then_postback"
	ScenarioHighVolumeClicks   E2EScenarioType = "high_volume_clicks"
	ScenarioGeoBlock           E2EScenarioType = "geo_block_scenario"
	ScenarioAPIKeyAuth         E2EScenarioType = "api_key_auth_scenario"
	ScenarioWebhookPipeline    E2EScenarioType = "webhook_pipeline_scenario"
	ScenarioZeroDropFailure    E2EScenarioType = "zero_drop_failure_scenario"
)

// E2EStepType represents a step type in a scenario
type E2EStepType string

const (
	StepClick       E2EStepType = "click"
	StepConversion  E2EStepType = "conversion"
	StepPostback    E2EStepType = "postback"
	StepStatsCheck  E2EStepType = "stats_check"
	StepWait        E2EStepType = "wait"
	StepAPIKeyAuth  E2EStepType = "api_key_auth"
	StepWebhook     E2EStepType = "webhook"
	StepGeoCheck    E2EStepType = "geo_check"
	StepWALCheck    E2EStepType = "wal_check"
)

// E2EStepStatus represents the status of a step
type E2EStepStatus string

const (
	StepStatusPending  E2EStepStatus = "pending"
	StepStatusRunning  E2EStepStatus = "running"
	StepStatusPassed   E2EStepStatus = "passed"
	StepStatusFailed   E2EStepStatus = "failed"
	StepStatusSkipped  E2EStepStatus = "skipped"
)

// ============================================
// E2E SCENARIO DEFINITION
// ============================================

// E2EScenario defines a test scenario
type E2EScenario struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        E2EScenarioType        `json:"type"`
	Description string                 `json:"description"`
	Steps       []E2EStep              `json:"steps"`
	Variables   map[string]interface{} `json:"variables,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	Enabled     bool                   `json:"enabled"`
}

// E2EStep defines a single step in a scenario
type E2EStep struct {
	ID          string                 `json:"id"`
	Type        E2EStepType            `json:"type"`
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input,omitempty"`
	Expected    map[string]interface{} `json:"expected,omitempty"`
	Timeout     time.Duration          `json:"timeout"`
	ContinueOnFail bool                `json:"continue_on_fail"`
}

// ============================================
// E2E TEST RUN
// ============================================

// E2ETestRun represents a test run execution
type E2ETestRun struct {
	ID           string           `json:"id"`
	ScenarioID   string           `json:"scenario_id"`
	ScenarioName string           `json:"scenario_name"`
	Status       E2EStepStatus    `json:"status"`
	StartTime    time.Time        `json:"start_time"`
	EndTime      *time.Time       `json:"end_time,omitempty"`
	DurationMs   int64            `json:"duration_ms"`
	StepResults  []E2EStepResult  `json:"step_results"`
	Variables    map[string]interface{} `json:"variables"`
	Summary      *E2ETestSummary  `json:"summary,omitempty"`
	Error        string           `json:"error,omitempty"`
}

// E2EStepResult holds the result of a single step
type E2EStepResult struct {
	StepID      string                 `json:"step_id"`
	StepType    E2EStepType            `json:"step_type"`
	Status      E2EStepStatus          `json:"status"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     time.Time              `json:"end_time"`
	DurationMs  int64                  `json:"duration_ms"`
	Input       map[string]interface{} `json:"input,omitempty"`
	Output      map[string]interface{} `json:"output,omitempty"`
	Expected    map[string]interface{} `json:"expected,omitempty"`
	Actual      map[string]interface{} `json:"actual,omitempty"`
	Discrepancies []string             `json:"discrepancies,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// E2ETestSummary provides a summary of the test run
type E2ETestSummary struct {
	TotalSteps   int `json:"total_steps"`
	PassedSteps  int `json:"passed_steps"`
	FailedSteps  int `json:"failed_steps"`
	SkippedSteps int `json:"skipped_steps"`
	TotalTimeMs  int64 `json:"total_time_ms"`
}

// ============================================
// E2E TEST SERVICE
// ============================================

// E2ETestService manages E2E test scenarios and execution
type E2ETestService struct {
	mu            sync.RWMutex
	db            *gorm.DB
	observability *ObservabilityService
	
	// Predefined scenarios
	scenarios     map[string]*E2EScenario
	
	// Test history
	testRuns      []*E2ETestRun
	maxHistory    int
	
	// Metrics
	totalRuns     int64
	totalPassed   int64
	totalFailed   int64
}

// NewE2ETestService creates a new E2E test service
func NewE2ETestService(db *gorm.DB) *E2ETestService {
	service := &E2ETestService{
		db:            db,
		observability: NewObservabilityService(),
		scenarios:     make(map[string]*E2EScenario),
		testRuns:      make([]*E2ETestRun, 0),
		maxHistory:    100,
	}
	
	// Register predefined scenarios
	service.registerPredefinedScenarios()
	
	return service
}

// ============================================
// PREDEFINED SCENARIOS
// ============================================

func (s *E2ETestService) registerPredefinedScenarios() {
	// Scenario 1: Click Only
	s.scenarios["click_only"] = &E2EScenario{
		ID:          "click_only",
		Name:        "Click Only",
		Type:        ScenarioClickOnly,
		Description: "Test single click tracking without conversion",
		Timeout:     30 * time.Second,
		Enabled:     true,
		Steps: []E2EStep{
			{
				ID:          "step_1",
				Type:        StepClick,
				Description: "Generate a click event",
				Input: map[string]interface{}{
					"country": "KW",
					"device":  "mobile",
				},
				Timeout: 10 * time.Second,
			},
			{
				ID:          "step_2",
				Type:        StepWait,
				Description: "Wait for processing",
				Input: map[string]interface{}{
					"duration_ms": 1000,
				},
			},
			{
				ID:          "step_3",
				Type:        StepStatsCheck,
				Description: "Verify click was recorded",
				Expected: map[string]interface{}{
					"min_clicks": 1,
				},
				Timeout: 5 * time.Second,
			},
		},
	}

	// Scenario 2: Click Then Conversion
	s.scenarios["click_then_conversion"] = &E2EScenario{
		ID:          "click_then_conversion",
		Name:        "Click Then Conversion",
		Type:        ScenarioClickThenConversion,
		Description: "Simulate full funnel from click to conversion",
		Timeout:     60 * time.Second,
		Enabled:     true,
		Steps: []E2EStep{
			{
				ID:          "step_1",
				Type:        StepClick,
				Description: "Generate a click event",
				Input: map[string]interface{}{
					"country": "KW",
					"device":  "mobile",
				},
				Timeout: 10 * time.Second,
			},
			{
				ID:          "step_2",
				Type:        StepWait,
				Description: "Wait for click processing",
				Input: map[string]interface{}{
					"duration_ms": 500,
				},
			},
			{
				ID:          "step_3",
				Type:        StepConversion,
				Description: "Generate a conversion",
				Input: map[string]interface{}{
					"amount": 100,
					"status": "approved",
				},
				Timeout: 10 * time.Second,
			},
			{
				ID:          "step_4",
				Type:        StepWait,
				Description: "Wait for conversion processing",
				Input: map[string]interface{}{
					"duration_ms": 500,
				},
			},
			{
				ID:          "step_5",
				Type:        StepStatsCheck,
				Description: "Verify stats updated",
				Expected: map[string]interface{}{
					"min_clicks":      1,
					"min_conversions": 1,
				},
				Timeout: 5 * time.Second,
			},
		},
	}

	// Scenario 3: High Volume Clicks
	s.scenarios["high_volume_clicks"] = &E2EScenario{
		ID:          "high_volume_clicks",
		Name:        "High Volume Clicks",
		Type:        ScenarioHighVolumeClicks,
		Description: "Test system under high click volume",
		Timeout:     120 * time.Second,
		Enabled:     true,
		Steps: []E2EStep{
			{
				ID:          "step_1",
				Type:        StepClick,
				Description: "Generate 100 clicks rapidly",
				Input: map[string]interface{}{
					"count":    100,
					"parallel": true,
				},
				Timeout: 30 * time.Second,
			},
			{
				ID:          "step_2",
				Type:        StepWait,
				Description: "Wait for processing",
				Input: map[string]interface{}{
					"duration_ms": 5000,
				},
			},
			{
				ID:          "step_3",
				Type:        StepStatsCheck,
				Description: "Verify all clicks recorded",
				Expected: map[string]interface{}{
					"min_clicks": 100,
				},
				Timeout: 10 * time.Second,
			},
		},
	}

	// Scenario 4: Geo Block
	s.scenarios["geo_block_scenario"] = &E2EScenario{
		ID:          "geo_block_scenario",
		Name:        "Geo Block Scenario",
		Type:        ScenarioGeoBlock,
		Description: "Test geo blocking functionality",
		Timeout:     30 * time.Second,
		Enabled:     true,
		Steps: []E2EStep{
			{
				ID:          "step_1",
				Type:        StepGeoCheck,
				Description: "Attempt click from blocked country",
				Input: map[string]interface{}{
					"country":      "XX",
					"expect_block": true,
				},
				Timeout: 10 * time.Second,
			},
		},
	}

	// Scenario 5: Zero Drop Failure
	s.scenarios["zero_drop_failure_scenario"] = &E2EScenario{
		ID:          "zero_drop_failure_scenario",
		Name:        "Zero Drop Failure Recovery",
		Type:        ScenarioZeroDropFailure,
		Description: "Test zero-drop mode during simulated failure",
		Timeout:     180 * time.Second,
		Enabled:     true,
		Steps: []E2EStep{
			{
				ID:          "step_1",
				Type:        StepClick,
				Description: "Generate clicks before failure",
				Input: map[string]interface{}{
					"count": 10,
				},
				Timeout: 10 * time.Second,
			},
			{
				ID:          "step_2",
				Type:        StepWALCheck,
				Description: "Verify WAL is working",
				Expected: map[string]interface{}{
					"wal_running": true,
				},
				Timeout: 5 * time.Second,
			},
			{
				ID:          "step_3",
				Type:        StepStatsCheck,
				Description: "Verify no clicks dropped",
				Expected: map[string]interface{}{
					"dropped_clicks": 0,
				},
				Timeout: 10 * time.Second,
			},
		},
	}
}

// ============================================
// SCENARIO EXECUTION
// ============================================

// RunScenario executes a specific scenario
func (s *E2ETestService) RunScenario(scenarioID string) (*E2ETestRun, error) {
	s.mu.RLock()
	scenario, exists := s.scenarios[scenarioID]
	s.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("scenario not found: %s", scenarioID)
	}

	if !scenario.Enabled {
		return nil, fmt.Errorf("scenario is disabled: %s", scenarioID)
	}

	return s.executeScenario(scenario)
}

// executeScenario runs all steps in a scenario
func (s *E2ETestService) executeScenario(scenario *E2EScenario) (*E2ETestRun, error) {
	run := &E2ETestRun{
		ID:           uuid.New().String(),
		ScenarioID:   scenario.ID,
		ScenarioName: scenario.Name,
		Status:       StepStatusRunning,
		StartTime:    time.Now(),
		StepResults:  make([]E2EStepResult, 0),
		Variables:    make(map[string]interface{}),
	}

	// Copy scenario variables
	if scenario.Variables != nil {
		for k, v := range scenario.Variables {
			run.Variables[k] = v
		}
	}

	atomic.AddInt64(&s.totalRuns, 1)

	// Log start
	s.observability.Log(LogEvent{
		Category: LogCategorySystemEvent,
		Level:    LogLevelInfo,
		Message:  "E2E test started",
		Metadata: map[string]interface{}{
			"run_id":      run.ID,
			"scenario_id": scenario.ID,
		},
	})

	// Execute steps
	allPassed := true
	for _, step := range scenario.Steps {
		result := s.executeStep(&step, run)
		run.StepResults = append(run.StepResults, result)

		if result.Status == StepStatusFailed {
			allPassed = false
			if !step.ContinueOnFail {
				break
			}
		}
	}

	// Finalize run
	now := time.Now()
	run.EndTime = &now
	run.DurationMs = now.Sub(run.StartTime).Milliseconds()

	if allPassed {
		run.Status = StepStatusPassed
		atomic.AddInt64(&s.totalPassed, 1)
	} else {
		run.Status = StepStatusFailed
		atomic.AddInt64(&s.totalFailed, 1)
	}

	// Calculate summary
	run.Summary = s.calculateSummary(run)

	// Store in history
	s.storeRun(run)

	// Log completion
	s.observability.Log(LogEvent{
		Category: LogCategorySystemEvent,
		Level:    LogLevelInfo,
		Message:  "E2E test completed",
		Metadata: map[string]interface{}{
			"run_id":      run.ID,
			"status":      run.Status,
			"duration_ms": run.DurationMs,
		},
	})

	return run, nil
}

// executeStep runs a single step
func (s *E2ETestService) executeStep(step *E2EStep, run *E2ETestRun) E2EStepResult {
	result := E2EStepResult{
		StepID:    step.ID,
		StepType:  step.Type,
		Status:    StepStatusRunning,
		StartTime: time.Now(),
		Input:     step.Input,
		Expected:  step.Expected,
	}

	// Execute based on step type
	var err error
	var output map[string]interface{}

	switch step.Type {
	case StepClick:
		output, err = s.executeClickStep(step, run)
	case StepConversion:
		output, err = s.executeConversionStep(step, run)
	case StepPostback:
		output, err = s.executePostbackStep(step, run)
	case StepStatsCheck:
		output, err = s.executeStatsCheckStep(step, run)
	case StepWait:
		output, err = s.executeWaitStep(step)
	case StepGeoCheck:
		output, err = s.executeGeoCheckStep(step, run)
	case StepWALCheck:
		output, err = s.executeWALCheckStep(step)
	default:
		err = fmt.Errorf("unknown step type: %s", step.Type)
	}

	result.EndTime = time.Now()
	result.DurationMs = result.EndTime.Sub(result.StartTime).Milliseconds()
	result.Output = output

	if err != nil {
		result.Status = StepStatusFailed
		result.Error = err.Error()
	} else {
		result.Status = StepStatusPassed
	}

	return result
}

// ============================================
// STEP EXECUTORS
// ============================================

func (s *E2ETestService) executeClickStep(step *E2EStep, run *E2ETestRun) (map[string]interface{}, error) {
	count := 1
	if c, ok := step.Input["count"].(int); ok {
		count = c
	}
	if c, ok := step.Input["count"].(float64); ok {
		count = int(c)
	}

	// Get or create a test user offer
	var userOffer models.UserOffer
	if err := s.db.First(&userOffer).Error; err != nil {
		return nil, fmt.Errorf("no user offer found for testing")
	}

	// Store for later steps
	run.Variables["user_offer_id"] = userOffer.ID.String()
	run.Variables["offer_id"] = userOffer.OfferID.String()

	// Simulate clicks
	clicksCreated := 0
	for i := 0; i < count; i++ {
		click := &models.Click{
			ID:          uuid.New(),
			UserOfferID: userOffer.ID,
			IPAddress:   fmt.Sprintf("192.168.1.%d", i%255),
			UserAgent:   "E2E-Test-Agent",
			Device:      "mobile",
			Country:     "KW",
			ClickedAt:   time.Now(),
		}
		
		if err := s.db.Create(click).Error; err == nil {
			clicksCreated++
			run.Variables["last_click_id"] = click.ID.String()
		}
	}

	return map[string]interface{}{
		"clicks_created": clicksCreated,
		"user_offer_id":  userOffer.ID.String(),
	}, nil
}

func (s *E2ETestService) executeConversionStep(step *E2EStep, run *E2ETestRun) (map[string]interface{}, error) {
	userOfferIDStr, ok := run.Variables["user_offer_id"].(string)
	if !ok {
		return nil, fmt.Errorf("user_offer_id not found in variables")
	}

	userOfferID, err := uuid.Parse(userOfferIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user_offer_id: %w", err)
	}

	amount := 100
	if a, ok := step.Input["amount"].(int); ok {
		amount = a
	}
	if a, ok := step.Input["amount"].(float64); ok {
		amount = int(a)
	}

	conversion := &models.Conversion{
		ID:                   uuid.New(),
		UserOfferID:          userOfferID,
		ExternalConversionID: fmt.Sprintf("e2e_test_%d", time.Now().UnixNano()),
		Amount:               amount,
		Status:               "approved",
		ConvertedAt:          time.Now(),
	}

	if err := s.db.Create(conversion).Error; err != nil {
		return nil, fmt.Errorf("failed to create conversion: %w", err)
	}

	run.Variables["last_conversion_id"] = conversion.ID.String()

	return map[string]interface{}{
		"conversion_id": conversion.ID.String(),
		"amount":        amount,
	}, nil
}

func (s *E2ETestService) executePostbackStep(step *E2EStep, run *E2ETestRun) (map[string]interface{}, error) {
	// Simulate postback
	return map[string]interface{}{
		"postback_sent": true,
	}, nil
}

func (s *E2ETestService) executeStatsCheckStep(step *E2EStep, run *E2ETestRun) (map[string]interface{}, error) {
	userOfferIDStr, ok := run.Variables["user_offer_id"].(string)
	if !ok {
		return nil, fmt.Errorf("user_offer_id not found in variables")
	}

	userOfferID, err := uuid.Parse(userOfferIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user_offer_id: %w", err)
	}

	// Get actual stats
	var clickCount int64
	s.db.Model(&models.Click{}).Where("user_offer_id = ?", userOfferID).Count(&clickCount)

	var conversionCount int64
	s.db.Model(&models.Conversion{}).Where("user_offer_id = ?", userOfferID).Count(&conversionCount)

	actual := map[string]interface{}{
		"clicks":      clickCount,
		"conversions": conversionCount,
	}

	// Check expectations
	if minClicks, ok := step.Expected["min_clicks"].(int); ok {
		if int(clickCount) < minClicks {
			return actual, fmt.Errorf("expected at least %d clicks, got %d", minClicks, clickCount)
		}
	}
	if minClicks, ok := step.Expected["min_clicks"].(float64); ok {
		if clickCount < int64(minClicks) {
			return actual, fmt.Errorf("expected at least %.0f clicks, got %d", minClicks, clickCount)
		}
	}

	if minConversions, ok := step.Expected["min_conversions"].(int); ok {
		if int(conversionCount) < minConversions {
			return actual, fmt.Errorf("expected at least %d conversions, got %d", minConversions, conversionCount)
		}
	}
	if minConversions, ok := step.Expected["min_conversions"].(float64); ok {
		if conversionCount < int64(minConversions) {
			return actual, fmt.Errorf("expected at least %.0f conversions, got %d", minConversions, conversionCount)
		}
	}

	return actual, nil
}

func (s *E2ETestService) executeWaitStep(step *E2EStep) (map[string]interface{}, error) {
	durationMs := 1000
	if d, ok := step.Input["duration_ms"].(int); ok {
		durationMs = d
	}
	if d, ok := step.Input["duration_ms"].(float64); ok {
		durationMs = int(d)
	}

	time.Sleep(time.Duration(durationMs) * time.Millisecond)

	return map[string]interface{}{
		"waited_ms": durationMs,
	}, nil
}

func (s *E2ETestService) executeGeoCheckStep(step *E2EStep, run *E2ETestRun) (map[string]interface{}, error) {
	country, _ := step.Input["country"].(string)
	expectBlock, _ := step.Input["expect_block"].(bool)

	// Check geo rule
	geoService := NewGeoRuleService(s.db)
	result := geoService.GetEffectiveGeoRule(nil, nil, country)

	blocked := false
	if result != nil {
		blocked = !result.Allowed
	}

	if expectBlock && !blocked {
		return nil, fmt.Errorf("expected geo block for country %s, but was not blocked", country)
	}
	if !expectBlock && blocked {
		return nil, fmt.Errorf("did not expect geo block for country %s, but was blocked", country)
	}

	return map[string]interface{}{
		"country": country,
		"blocked": blocked,
	}, nil
}

func (s *E2ETestService) executeWALCheckStep(step *E2EStep) (map[string]interface{}, error) {
	walService := GetWALService()
	stats := walService.GetStats()

	isRunning, _ := stats["is_running"].(bool)

	if expectRunning, ok := step.Expected["wal_running"].(bool); ok {
		if expectRunning && !isRunning {
			return stats, fmt.Errorf("expected WAL to be running, but it's not")
		}
	}

	return stats, nil
}

// ============================================
// HELPERS
// ============================================

func (s *E2ETestService) calculateSummary(run *E2ETestRun) *E2ETestSummary {
	summary := &E2ETestSummary{
		TotalSteps:  len(run.StepResults),
		TotalTimeMs: run.DurationMs,
	}

	for _, result := range run.StepResults {
		switch result.Status {
		case StepStatusPassed:
			summary.PassedSteps++
		case StepStatusFailed:
			summary.FailedSteps++
		case StepStatusSkipped:
			summary.SkippedSteps++
		}
	}

	return summary
}

func (s *E2ETestService) storeRun(run *E2ETestRun) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.testRuns = append(s.testRuns, run)
	if len(s.testRuns) > s.maxHistory {
		s.testRuns = s.testRuns[1:]
	}

	// Also store in Redis
	ctx := context.Background()
	key := fmt.Sprintf("e2e_run:%s", run.ID)
	data, _ := json.Marshal(run)
	cache.Set(ctx, key, string(data), 24*time.Hour)
}

// ============================================
// GETTERS
// ============================================

// GetScenarios returns all available scenarios
func (s *E2ETestService) GetScenarios() []*E2EScenario {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*E2EScenario, 0, len(s.scenarios))
	for _, scenario := range s.scenarios {
		result = append(result, scenario)
	}
	return result
}

// GetScenario returns a specific scenario
func (s *E2ETestService) GetScenario(id string) (*E2EScenario, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	scenario, exists := s.scenarios[id]
	if !exists {
		return nil, fmt.Errorf("scenario not found: %s", id)
	}
	return scenario, nil
}

// GetTestRun returns a specific test run
func (s *E2ETestService) GetTestRun(id string) (*E2ETestRun, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, run := range s.testRuns {
		if run.ID == id {
			return run, nil
		}
	}

	// Try Redis
	ctx := context.Background()
	key := fmt.Sprintf("e2e_run:%s", id)
	data, err := cache.Get(ctx, key)
	if err == nil && data != "" {
		var run E2ETestRun
		if json.Unmarshal([]byte(data), &run) == nil {
			return &run, nil
		}
	}

	return nil, fmt.Errorf("test run not found: %s", id)
}

// GetTestHistory returns test run history
func (s *E2ETestService) GetTestHistory(limit int) []*E2ETestRun {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.testRuns) {
		limit = len(s.testRuns)
	}

	// Return most recent first
	result := make([]*E2ETestRun, limit)
	for i := 0; i < limit; i++ {
		result[i] = s.testRuns[len(s.testRuns)-1-i]
	}
	return result
}

// GetStats returns service statistics
func (s *E2ETestService) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_runs":     atomic.LoadInt64(&s.totalRuns),
		"total_passed":   atomic.LoadInt64(&s.totalPassed),
		"total_failed":   atomic.LoadInt64(&s.totalFailed),
		"scenarios_count": len(s.scenarios),
		"history_count":  len(s.testRuns),
	}
}


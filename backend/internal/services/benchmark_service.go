package services

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/cache"
	"github.com/aljapah/afftok-backend-prod/internal/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// BENCHMARK TYPES
// ============================================

// BenchmarkType represents the type of benchmark
type BenchmarkType string

const (
	BenchmarkClickPath      BenchmarkType = "click_path"
	BenchmarkConversionPath BenchmarkType = "conversion_path"
	BenchmarkWebhookPath    BenchmarkType = "webhook_path"
	BenchmarkDBLatency      BenchmarkType = "db_latency"
	BenchmarkRedisLatency   BenchmarkType = "redis_latency"
	BenchmarkWALReplay      BenchmarkType = "wal_replay"
	BenchmarkStreamLag      BenchmarkType = "stream_lag"
	BenchmarkFull           BenchmarkType = "full"
)

// ============================================
// LATENCY STATS
// ============================================

// LatencyStats holds percentile latency statistics
type LatencyStats struct {
	P50     float64 `json:"p50"`
	P90     float64 `json:"p90"`
	P95     float64 `json:"p95"`
	P99     float64 `json:"p99"`
	Min     float64 `json:"min"`
	Max     float64 `json:"max"`
	Avg     float64 `json:"avg"`
	Count   int     `json:"count"`
	Errors  int     `json:"errors"`
}

// ============================================
// BENCHMARK RESULT
// ============================================

// BenchmarkResult holds the result of a benchmark run
type BenchmarkResult struct {
	ID           string                    `json:"id"`
	Type         BenchmarkType             `json:"type"`
	StartTime    time.Time                 `json:"start_time"`
	EndTime      time.Time                 `json:"end_time"`
	DurationMs   int64                     `json:"duration_ms"`
	Iterations   int                       `json:"iterations"`
	Concurrency  int                       `json:"concurrency"`
	Results      map[string]*LatencyStats  `json:"results"`
	Summary      *BenchmarkSummary         `json:"summary"`
	Metadata     map[string]interface{}    `json:"metadata,omitempty"`
}

// BenchmarkSummary provides a summary of the benchmark
type BenchmarkSummary struct {
	OverallStatus    string  `json:"overall_status"` // passed, warning, failed
	ClickLatencyMs   *LatencyStats `json:"click_latency_ms,omitempty"`
	ConversionLatencyMs *LatencyStats `json:"conversion_latency_ms,omitempty"`
	WebhookLatencyMs *LatencyStats `json:"webhook_latency_ms,omitempty"`
	DBLatencyMs      float64 `json:"db_latency_ms"`
	RedisLatencyMs   float64 `json:"redis_latency_ms"`
	StreamLag        int64   `json:"stream_lag"`
	WALReplayMs      float64 `json:"wal_replay_ms,omitempty"`
	ThroughputRPS    float64 `json:"throughput_rps"`
}

// ============================================
// BENCHMARK SERVICE
// ============================================

// BenchmarkService runs performance benchmarks
type BenchmarkService struct {
	mu            sync.RWMutex
	db            *gorm.DB
	observability *ObservabilityService
	
	// History
	results       []*BenchmarkResult
	maxResults    int
	
	// Metrics
	totalRuns     int64
	
	// Thresholds
	clickP99Threshold      float64
	conversionP99Threshold float64
	dbLatencyThreshold     float64
	redisLatencyThreshold  float64
}

// NewBenchmarkService creates a new benchmark service
func NewBenchmarkService(db *gorm.DB) *BenchmarkService {
	return &BenchmarkService{
		db:                     db,
		observability:          NewObservabilityService(),
		results:                make([]*BenchmarkResult, 0),
		maxResults:             50,
		clickP99Threshold:      100,  // 100ms
		conversionP99Threshold: 200,  // 200ms
		dbLatencyThreshold:     30,   // 30ms
		redisLatencyThreshold:  10,   // 10ms
	}
}

// ============================================
// MAIN BENCHMARK RUNNER
// ============================================

// RunBenchmark runs a specific benchmark type
func (s *BenchmarkService) RunBenchmark(benchType BenchmarkType, iterations, concurrency int) (*BenchmarkResult, error) {
	if iterations <= 0 {
		iterations = 100
	}
	if concurrency <= 0 {
		concurrency = 10
	}

	result := &BenchmarkResult{
		ID:          uuid.New().String(),
		Type:        benchType,
		StartTime:   time.Now(),
		Iterations:  iterations,
		Concurrency: concurrency,
		Results:     make(map[string]*LatencyStats),
	}

	atomic.AddInt64(&s.totalRuns, 1)

	s.observability.Log(LogEvent{
		Category: LogCategorySystemEvent,
		Level:    LogLevelInfo,
		Message:  "Benchmark started",
		Metadata: map[string]interface{}{
			"benchmark_id": result.ID,
			"type":         benchType,
			"iterations":   iterations,
		},
	})

	switch benchType {
	case BenchmarkClickPath:
		s.benchmarkClickPath(result)
	case BenchmarkConversionPath:
		s.benchmarkConversionPath(result)
	case BenchmarkWebhookPath:
		s.benchmarkWebhookPath(result)
	case BenchmarkDBLatency:
		s.benchmarkDBLatency(result)
	case BenchmarkRedisLatency:
		s.benchmarkRedisLatency(result)
	case BenchmarkWALReplay:
		s.benchmarkWALReplay(result)
	case BenchmarkStreamLag:
		s.benchmarkStreamLag(result)
	case BenchmarkFull:
		s.benchmarkFull(result)
	}

	result.EndTime = time.Now()
	result.DurationMs = result.EndTime.Sub(result.StartTime).Milliseconds()
	result.Summary = s.calculateSummary(result)

	// Store result
	s.storeResult(result)

	s.observability.Log(LogEvent{
		Category: LogCategorySystemEvent,
		Level:    LogLevelInfo,
		Message:  "Benchmark completed",
		Metadata: map[string]interface{}{
			"benchmark_id": result.ID,
			"status":       result.Summary.OverallStatus,
			"duration_ms":  result.DurationMs,
		},
	})

	return result, nil
}

// ============================================
// INDIVIDUAL BENCHMARKS
// ============================================

func (s *BenchmarkService) benchmarkClickPath(result *BenchmarkResult) {
	latencies := s.runConcurrentBenchmark(result.Iterations, result.Concurrency, func() (float64, error) {
		start := time.Now()
		
		// Simulate click tracking operations
		var userOffer models.UserOffer
		if err := s.db.First(&userOffer).Error; err != nil {
			return 0, err
		}
		
		// Simulate Redis counter increment
		ctx := context.Background()
		if cache.RedisClient != nil {
			cache.RedisClient.Incr(ctx, "benchmark:click_counter")
		}
		
		return float64(time.Since(start).Microseconds()) / 1000.0, nil
	})

	result.Results["click_ingestion"] = s.calculateLatencyStats(latencies)
}

func (s *BenchmarkService) benchmarkConversionPath(result *BenchmarkResult) {
	latencies := s.runConcurrentBenchmark(result.Iterations, result.Concurrency, func() (float64, error) {
		start := time.Now()
		
		// Simulate conversion processing
		var userOffer models.UserOffer
		if err := s.db.First(&userOffer).Error; err != nil {
			return 0, err
		}
		
		// Simulate stats update
		s.db.Model(&userOffer).Update("total_conversions", gorm.Expr("total_conversions + 1"))
		
		return float64(time.Since(start).Microseconds()) / 1000.0, nil
	})

	result.Results["conversion_processing"] = s.calculateLatencyStats(latencies)
}

func (s *BenchmarkService) benchmarkWebhookPath(result *BenchmarkResult) {
	latencies := s.runConcurrentBenchmark(result.Iterations, result.Concurrency, func() (float64, error) {
		start := time.Now()
		
		// Simulate webhook preparation
		time.Sleep(time.Millisecond * 5) // Simulate network delay
		
		return float64(time.Since(start).Microseconds()) / 1000.0, nil
	})

	result.Results["webhook_delivery"] = s.calculateLatencyStats(latencies)
}

func (s *BenchmarkService) benchmarkDBLatency(result *BenchmarkResult) {
	latencies := s.runConcurrentBenchmark(result.Iterations, result.Concurrency, func() (float64, error) {
		start := time.Now()
		
		var count int
		if err := s.db.Raw("SELECT 1").Scan(&count).Error; err != nil {
			return 0, err
		}
		
		return float64(time.Since(start).Microseconds()) / 1000.0, nil
	})

	result.Results["db_ping"] = s.calculateLatencyStats(latencies)

	// Also test read query
	readLatencies := s.runConcurrentBenchmark(result.Iterations/2, result.Concurrency, func() (float64, error) {
		start := time.Now()
		
		var count int64
		if err := s.db.Model(&models.Click{}).Limit(1).Count(&count).Error; err != nil {
			return 0, err
		}
		
		return float64(time.Since(start).Microseconds()) / 1000.0, nil
	})

	result.Results["db_read"] = s.calculateLatencyStats(readLatencies)
}

func (s *BenchmarkService) benchmarkRedisLatency(result *BenchmarkResult) {
	if cache.RedisClient == nil {
		return
	}

	ctx := context.Background()

	// Ping latency
	pingLatencies := s.runConcurrentBenchmark(result.Iterations, result.Concurrency, func() (float64, error) {
		start := time.Now()
		
		if err := cache.RedisClient.Ping(ctx).Err(); err != nil {
			return 0, err
		}
		
		return float64(time.Since(start).Microseconds()) / 1000.0, nil
	})

	result.Results["redis_ping"] = s.calculateLatencyStats(pingLatencies)

	// Get/Set latency
	gsLatencies := s.runConcurrentBenchmark(result.Iterations, result.Concurrency, func() (float64, error) {
		start := time.Now()
		key := "benchmark:test_key"
		
		cache.RedisClient.Set(ctx, key, "test_value", time.Minute)
		cache.RedisClient.Get(ctx, key)
		
		return float64(time.Since(start).Microseconds()) / 1000.0, nil
	})

	result.Results["redis_get_set"] = s.calculateLatencyStats(gsLatencies)
}

func (s *BenchmarkService) benchmarkWALReplay(result *BenchmarkResult) {
	walService := GetWALService()
	
	start := time.Now()
	stats := walService.GetStats()
	
	result.Results["wal_status"] = &LatencyStats{
		Avg:   float64(time.Since(start).Microseconds()) / 1000.0,
		Count: 1,
	}
	
	result.Metadata = map[string]interface{}{
		"wal_stats": stats,
	}
}

func (s *BenchmarkService) benchmarkStreamLag(result *BenchmarkResult) {
	if cache.RedisClient == nil {
		return
	}

	ctx := context.Background()
	streams := []string{StreamClicks, StreamConversions, StreamPostbacks}
	
	totalLag := int64(0)
	for _, stream := range streams {
		pending := cache.RedisClient.XPending(ctx, stream, "afftok-consumers")
		if pending.Err() == nil {
			totalLag += pending.Val().Count
		}
	}

	result.Results["stream_lag"] = &LatencyStats{
		Avg:   float64(totalLag),
		Count: len(streams),
	}
}

func (s *BenchmarkService) benchmarkFull(result *BenchmarkResult) {
	// Run all benchmarks
	s.benchmarkClickPath(result)
	s.benchmarkConversionPath(result)
	s.benchmarkDBLatency(result)
	s.benchmarkRedisLatency(result)
	s.benchmarkStreamLag(result)
}

// ============================================
// HELPERS
// ============================================

func (s *BenchmarkService) runConcurrentBenchmark(iterations, concurrency int, fn func() (float64, error)) []float64 {
	var wg sync.WaitGroup
	var mu sync.Mutex
	latencies := make([]float64, 0, iterations)
	
	iterationsPerWorker := iterations / concurrency
	
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterationsPerWorker; j++ {
				latency, err := fn()
				if err == nil {
					mu.Lock()
					latencies = append(latencies, latency)
					mu.Unlock()
				}
			}
		}()
	}
	
	wg.Wait()
	return latencies
}

func (s *BenchmarkService) calculateLatencyStats(latencies []float64) *LatencyStats {
	if len(latencies) == 0 {
		return &LatencyStats{}
	}

	// Sort for percentiles
	sorted := make([]float64, len(latencies))
	copy(sorted, latencies)
	sort.Float64s(sorted)

	stats := &LatencyStats{
		Count: len(latencies),
		Min:   sorted[0],
		Max:   sorted[len(sorted)-1],
		P50:   s.percentile(sorted, 50),
		P90:   s.percentile(sorted, 90),
		P95:   s.percentile(sorted, 95),
		P99:   s.percentile(sorted, 99),
	}

	// Calculate average
	var sum float64
	for _, v := range latencies {
		sum += v
	}
	stats.Avg = sum / float64(len(latencies))

	return stats
}

func (s *BenchmarkService) percentile(sorted []float64, p int) float64 {
	if len(sorted) == 0 {
		return 0
	}
	index := (p * len(sorted)) / 100
	if index >= len(sorted) {
		index = len(sorted) - 1
	}
	return sorted[index]
}

func (s *BenchmarkService) calculateSummary(result *BenchmarkResult) *BenchmarkSummary {
	summary := &BenchmarkSummary{
		OverallStatus: "passed",
	}

	// Click latency
	if stats, ok := result.Results["click_ingestion"]; ok {
		summary.ClickLatencyMs = stats
		if stats.P99 > s.clickP99Threshold {
			summary.OverallStatus = "warning"
		}
	}

	// Conversion latency
	if stats, ok := result.Results["conversion_processing"]; ok {
		summary.ConversionLatencyMs = stats
		if stats.P99 > s.conversionP99Threshold {
			summary.OverallStatus = "warning"
		}
	}

	// Webhook latency
	if stats, ok := result.Results["webhook_delivery"]; ok {
		summary.WebhookLatencyMs = stats
	}

	// DB latency
	if stats, ok := result.Results["db_ping"]; ok {
		summary.DBLatencyMs = stats.Avg
		if stats.Avg > s.dbLatencyThreshold {
			summary.OverallStatus = "warning"
		}
	}

	// Redis latency
	if stats, ok := result.Results["redis_ping"]; ok {
		summary.RedisLatencyMs = stats.Avg
		if stats.Avg > s.redisLatencyThreshold {
			summary.OverallStatus = "warning"
		}
	}

	// Stream lag
	if stats, ok := result.Results["stream_lag"]; ok {
		summary.StreamLag = int64(stats.Avg)
		if stats.Avg > 1000 {
			summary.OverallStatus = "warning"
		}
	}

	// Calculate throughput
	if result.DurationMs > 0 {
		summary.ThroughputRPS = float64(result.Iterations) / (float64(result.DurationMs) / 1000.0)
	}

	return summary
}

func (s *BenchmarkService) storeResult(result *BenchmarkResult) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.results = append(s.results, result)
	if len(s.results) > s.maxResults {
		s.results = s.results[1:]
	}
}

// ============================================
// GETTERS
// ============================================

// GetLatestResult returns the most recent benchmark result
func (s *BenchmarkService) GetLatestResult() *BenchmarkResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.results) == 0 {
		return nil
	}
	return s.results[len(s.results)-1]
}

// GetResultHistory returns benchmark history
func (s *BenchmarkService) GetResultHistory(limit int) []*BenchmarkResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.results) {
		limit = len(s.results)
	}

	result := make([]*BenchmarkResult, limit)
	for i := 0; i < limit; i++ {
		result[i] = s.results[len(s.results)-1-i]
	}
	return result
}

// GetStats returns service statistics
func (s *BenchmarkService) GetStats() map[string]interface{} {
	return map[string]interface{}{
		"total_runs":     atomic.LoadInt64(&s.totalRuns),
		"results_stored": len(s.results),
		"thresholds": map[string]interface{}{
			"click_p99_ms":      s.clickP99Threshold,
			"conversion_p99_ms": s.conversionP99Threshold,
			"db_latency_ms":     s.dbLatencyThreshold,
			"redis_latency_ms":  s.redisLatencyThreshold,
		},
	}
}


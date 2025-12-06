package handlers

import (
	"fmt"
	"math/rand"
	"net/http"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminStressHandler handles admin stress testing API endpoints
type AdminStressHandler struct {
	clickServiceV2 *services.ClickServiceV2
	observability  *services.ObservabilityService
}

// NewAdminStressHandler creates a new admin stress handler
func NewAdminStressHandler() *AdminStressHandler {
	return &AdminStressHandler{
		clickServiceV2: services.NewClickServiceV2(),
		observability:  services.NewObservabilityService(),
	}
}

// StressTestResponse represents stress test results
type StressTestResponse struct {
	CorrelationID  string        `json:"correlation_id"`
	Timestamp      string        `json:"timestamp"`
	TestType       string        `json:"test_type"`
	TotalRequests  int64         `json:"total_requests"`
	SuccessCount   int64         `json:"success_count"`
	FailureCount   int64         `json:"failure_count"`
	DroppedCount   int64         `json:"dropped_count"`
	Duration       string        `json:"duration"`
	DurationMs     int64         `json:"duration_ms"`
	RequestsPerSec float64       `json:"requests_per_second"`
	Latency        LatencyStats  `json:"latency"`
	Configuration  TestConfig    `json:"configuration"`
}

// LatencyStats represents latency statistics
type LatencyStats struct {
	AvgMs float64 `json:"avg_ms"`
	MinMs float64 `json:"min_ms"`
	MaxMs float64 `json:"max_ms"`
	P50Ms float64 `json:"p50_ms"`
	P95Ms float64 `json:"p95_ms"`
	P99Ms float64 `json:"p99_ms"`
}

// TestConfig represents test configuration
type TestConfig struct {
	Count       int  `json:"count"`
	Concurrency int  `json:"concurrency"`
	Burst       bool `json:"burst"`
	Async       bool `json:"async"`
}

// SimulateClicks simulates high-volume click traffic
// GET /api/admin/stress/clicks?count=10000&concurrency=100&burst=false&async=true
func (h *AdminStressHandler) SimulateClicks(c *gin.Context) {
	correlationID := generateCorrelationID()

	// Parse parameters
	config := TestConfig{
		Count:       1000,
		Concurrency: 50,
		Burst:       false,
		Async:       true,
	}

	if countStr := c.Query("count"); countStr != "" {
		if parsed, err := strconv.Atoi(countStr); err == nil && parsed > 0 && parsed <= 100000 {
			config.Count = parsed
		}
	}
	if concStr := c.Query("concurrency"); concStr != "" {
		if parsed, err := strconv.Atoi(concStr); err == nil && parsed > 0 && parsed <= 500 {
			config.Concurrency = parsed
		}
	}
	config.Burst = c.Query("burst") == "true"
	config.Async = c.Query("async") != "false"

	// Generate test UserOffer IDs
	userOfferIDs := make([]uuid.UUID, 10)
	for i := range userOfferIDs {
		userOfferIDs[i] = uuid.New()
	}

	// Run stress test
	result := h.runClickStressTest(config, userOfferIDs)

	response := StressTestResponse{
		CorrelationID:  correlationID,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		TestType:       "click_simulation",
		TotalRequests:  result.TotalRequests,
		SuccessCount:   result.SuccessCount,
		FailureCount:   result.FailureCount,
		DroppedCount:   result.DroppedCount,
		Duration:       result.Duration.String(),
		DurationMs:     result.Duration.Milliseconds(),
		RequestsPerSec: result.RequestsPerSec,
		Latency:        result.Latency,
		Configuration:  config,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}

// StressResult represents internal stress test result
type StressResult struct {
	TotalRequests  int64
	SuccessCount   int64
	FailureCount   int64
	DroppedCount   int64
	Duration       time.Duration
	RequestsPerSec float64
	Latency        LatencyStats
}

// runClickStressTest executes the click stress test
func (h *AdminStressHandler) runClickStressTest(config TestConfig, userOfferIDs []uuid.UUID) StressResult {
	var (
		successCount int64
		failureCount int64
		droppedCount int64
		totalLatency int64
		minLatency   int64 = 1<<63 - 1
		maxLatency   int64
		latencies    []int64
		latencyMutex sync.Mutex
	)

	work := make(chan int, config.Count)
	for i := 0; i < config.Count; i++ {
		work <- i
	}
	close(work)

	start := time.Now()
	var wg sync.WaitGroup

	for w := 0; w < config.Concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for range work {
				clickData := &services.ClickData{
					UserOfferID: userOfferIDs[rand.Intn(len(userOfferIDs))],
					IP:          fmt.Sprintf("192.168.%d.%d", rand.Intn(256), rand.Intn(256)),
					UserAgent:   "Mozilla/5.0 (Stress Test)",
					Device:      []string{"mobile", "desktop", "tablet"}[rand.Intn(3)],
					Browser:     []string{"Chrome", "Firefox", "Safari"}[rand.Intn(3)],
					OS:          []string{"iOS", "Android", "Windows", "macOS"}[rand.Intn(4)],
					Country:     []string{"US", "UK", "SA", "AE", "EG"}[rand.Intn(5)],
					Fingerprint: uuid.New().String()[:16],
				}

				clickStart := time.Now()
				var success bool

				if config.Async {
					success = h.clickServiceV2.TrackClickAsync(clickData)
					if !success {
						atomic.AddInt64(&droppedCount, 1)
					}
				} else {
					_, err := h.clickServiceV2.TrackClickSync(clickData)
					success = err == nil
				}

				latency := time.Since(clickStart).Microseconds()

				if success {
					atomic.AddInt64(&successCount, 1)
				} else if !config.Async {
					atomic.AddInt64(&failureCount, 1)
				}

				atomic.AddInt64(&totalLatency, latency)

				// Update min/max
				for {
					old := atomic.LoadInt64(&minLatency)
					if latency >= old || atomic.CompareAndSwapInt64(&minLatency, old, latency) {
						break
					}
				}
				for {
					old := atomic.LoadInt64(&maxLatency)
					if latency <= old || atomic.CompareAndSwapInt64(&maxLatency, old, latency) {
						break
					}
				}

				latencyMutex.Lock()
				latencies = append(latencies, latency)
				latencyMutex.Unlock()

				if !config.Burst {
					time.Sleep(time.Microsecond * 100)
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	result := StressResult{
		TotalRequests:  int64(config.Count),
		SuccessCount:   successCount,
		FailureCount:   failureCount,
		DroppedCount:   droppedCount,
		Duration:       duration,
		RequestsPerSec: float64(config.Count) / duration.Seconds(),
	}

	if len(latencies) > 0 {
		result.Latency.AvgMs = float64(totalLatency) / float64(len(latencies)) / 1000
		result.Latency.MinMs = float64(minLatency) / 1000
		result.Latency.MaxMs = float64(maxLatency) / 1000

		sortInt64Slice(latencies)
		result.Latency.P50Ms = float64(latencies[len(latencies)*50/100]) / 1000
		result.Latency.P95Ms = float64(latencies[len(latencies)*95/100]) / 1000
		result.Latency.P99Ms = float64(latencies[len(latencies)*99/100]) / 1000
	}

	return result
}

// SimulatePostbacks simulates high-volume postback traffic
// GET /api/admin/stress/postbacks?count=500&concurrency=20
func (h *AdminStressHandler) SimulatePostbacks(c *gin.Context) {
	correlationID := generateCorrelationID()

	config := TestConfig{
		Count:       500,
		Concurrency: 20,
	}

	if countStr := c.Query("count"); countStr != "" {
		if parsed, err := strconv.Atoi(countStr); err == nil && parsed > 0 && parsed <= 10000 {
			config.Count = parsed
		}
	}
	if concStr := c.Query("concurrency"); concStr != "" {
		if parsed, err := strconv.Atoi(concStr); err == nil && parsed > 0 && parsed <= 100 {
			config.Concurrency = parsed
		}
	}

	result := h.runPostbackStressTest(config)

	response := StressTestResponse{
		CorrelationID:  correlationID,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		TestType:       "postback_simulation",
		TotalRequests:  result.TotalRequests,
		SuccessCount:   result.SuccessCount,
		FailureCount:   result.FailureCount,
		Duration:       result.Duration.String(),
		DurationMs:     result.Duration.Milliseconds(),
		RequestsPerSec: result.RequestsPerSec,
		Latency:        result.Latency,
		Configuration:  config,
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           response,
	})
}

// runPostbackStressTest executes the postback stress test
func (h *AdminStressHandler) runPostbackStressTest(config TestConfig) StressResult {
	var (
		successCount int64
		totalLatency int64
		latencies    []int64
		latencyMutex sync.Mutex
	)

	work := make(chan int, config.Count)
	for i := 0; i < config.Count; i++ {
		work <- i
	}
	close(work)

	start := time.Now()
	var wg sync.WaitGroup

	for w := 0; w < config.Concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for range work {
				postbackStart := time.Now()

				h.observability.LogPostback(
					uuid.New().String(),
					"test_network",
					uuid.New().String(),
					fmt.Sprintf("10.0.%d.%d", rand.Intn(256), rand.Intn(256)),
					true,
					"stress_test",
					0,
				)

				latency := time.Since(postbackStart).Microseconds()
				atomic.AddInt64(&successCount, 1)
				atomic.AddInt64(&totalLatency, latency)

				latencyMutex.Lock()
				latencies = append(latencies, latency)
				latencyMutex.Unlock()
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	result := StressResult{
		TotalRequests:  int64(config.Count),
		SuccessCount:   successCount,
		Duration:       duration,
		RequestsPerSec: float64(config.Count) / duration.Seconds(),
	}

	if len(latencies) > 0 {
		result.Latency.AvgMs = float64(totalLatency) / float64(len(latencies)) / 1000
		sortInt64Slice(latencies)
		result.Latency.P50Ms = float64(latencies[len(latencies)*50/100]) / 1000
		result.Latency.P95Ms = float64(latencies[len(latencies)*95/100]) / 1000
		result.Latency.P99Ms = float64(latencies[len(latencies)*99/100]) / 1000
	}

	return result
}

// RunFullStressTest runs a comprehensive stress test
// GET /api/admin/stress/full
func (h *AdminStressHandler) RunFullStressTest(c *gin.Context) {
	correlationID := generateCorrelationID()

	testIDs := make([]uuid.UUID, 5)
	for i := range testIDs {
		testIDs[i] = uuid.New()
	}

	results := make(map[string]interface{})

	// Test 1: Burst clicks
	burstConfig := TestConfig{Count: 2000, Concurrency: 200, Burst: true, Async: true}
	results["burst_clicks"] = h.runClickStressTest(burstConfig, testIDs)

	// Test 2: Sustained clicks
	sustainedConfig := TestConfig{Count: 5000, Concurrency: 100, Burst: false, Async: true}
	results["sustained_clicks"] = h.runClickStressTest(sustainedConfig, testIDs)

	// Test 3: Postbacks
	postbackConfig := TestConfig{Count: 500, Concurrency: 50}
	results["postbacks"] = h.runPostbackStressTest(postbackConfig)

	// Get final stats
	results["worker_pools"] = services.GetAllPoolStats()
	results["cache"] = services.NewCacheService().GetCacheStats()
	results["metrics"] = h.observability.GetMetrics()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"results":   results,
		},
	})
}

// GetWorkerPoolStats returns worker pool statistics
// GET /api/admin/stress/pools
func (h *AdminStressHandler) GetWorkerPoolStats(c *gin.Context) {
	correlationID := generateCorrelationID()

	stats := services.GetAllPoolStats()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"pools":     stats,
		},
	})
}

// GetCacheStats returns cache statistics
// GET /api/admin/stress/cache
func (h *AdminStressHandler) GetCacheStats(c *gin.Context) {
	correlationID := generateCorrelationID()

	cacheService := services.NewCacheService()
	stats := cacheService.GetCacheStats()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"cache":     stats,
		},
	})
}

// sortInt64Slice sorts a slice of int64
func sortInt64Slice(arr []int64) {
	for i := 1; i < len(arr); i++ {
		key := arr[i]
		j := i - 1
		for j >= 0 && arr[j] > key {
			arr[j+1] = arr[j]
			j--
		}
		arr[j+1] = key
	}
}


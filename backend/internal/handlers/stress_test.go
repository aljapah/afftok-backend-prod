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

// StressTestHandler handles load testing endpoints
type StressTestHandler struct {
	clickServiceV2 *services.ClickServiceV2
	observability  *services.ObservabilityService
}

// NewStressTestHandler creates a new stress test handler
func NewStressTestHandler() *StressTestHandler {
	return &StressTestHandler{
		clickServiceV2: services.NewClickServiceV2(),
		observability:  services.NewObservabilityService(),
	}
}

// StressTestResult represents the result of a stress test
type StressTestResult struct {
	TestType        string        `json:"test_type"`
	TotalRequests   int64         `json:"total_requests"`
	SuccessCount    int64         `json:"success_count"`
	FailureCount    int64         `json:"failure_count"`
	DroppedCount    int64         `json:"dropped_count"`
	Duration        time.Duration `json:"duration"`
	RequestsPerSec  float64       `json:"requests_per_second"`
	AvgLatencyMs    float64       `json:"avg_latency_ms"`
	MinLatencyMs    float64       `json:"min_latency_ms"`
	MaxLatencyMs    float64       `json:"max_latency_ms"`
	P50LatencyMs    float64       `json:"p50_latency_ms"`
	P95LatencyMs    float64       `json:"p95_latency_ms"`
	P99LatencyMs    float64       `json:"p99_latency_ms"`
	Timestamp       time.Time     `json:"timestamp"`
}

// SimulateClicks simulates high-volume click traffic
// GET /api/internal/stress/clicks?count=10000&concurrency=100&burst=false
func (h *StressTestHandler) SimulateClicks(c *gin.Context) {
	// Parse parameters
	count := 1000
	if countStr := c.Query("count"); countStr != "" {
		if parsed, err := strconv.Atoi(countStr); err == nil && parsed > 0 && parsed <= 100000 {
			count = parsed
		}
	}

	concurrency := 50
	if concStr := c.Query("concurrency"); concStr != "" {
		if parsed, err := strconv.Atoi(concStr); err == nil && parsed > 0 && parsed <= 500 {
			concurrency = parsed
		}
	}

	burst := c.Query("burst") == "true"
	async := c.Query("async") != "false" // Default to async

	// Generate test UserOffer IDs
	userOfferIDs := make([]uuid.UUID, 10)
	for i := range userOfferIDs {
		userOfferIDs[i] = uuid.New()
	}

	// Run stress test
	result := h.runClickStressTest(count, concurrency, userOfferIDs, burst, async)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"result":  result,
	})
}

// runClickStressTest executes the click stress test
func (h *StressTestHandler) runClickStressTest(count, concurrency int, userOfferIDs []uuid.UUID, burst, async bool) StressTestResult {
	result := StressTestResult{
		TestType:  "click_simulation",
		Timestamp: time.Now(),
	}

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

	// Create work channel
	work := make(chan int, count)
	for i := 0; i < count; i++ {
		work <- i
	}
	close(work)

	start := time.Now()
	var wg sync.WaitGroup

	// Start workers
	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for range work {
				// Simulate click data
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

				if async {
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
				} else if !async {
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

				// Store latency for percentile calculation
				latencyMutex.Lock()
				latencies = append(latencies, latency)
				latencyMutex.Unlock()

				// Add small delay if not burst mode
				if !burst {
					time.Sleep(time.Microsecond * 100)
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	// Calculate results
	result.TotalRequests = int64(count)
	result.SuccessCount = successCount
	result.FailureCount = failureCount
	result.DroppedCount = droppedCount
	result.Duration = duration
	result.RequestsPerSec = float64(count) / duration.Seconds()

	if len(latencies) > 0 {
		result.AvgLatencyMs = float64(totalLatency) / float64(len(latencies)) / 1000
		result.MinLatencyMs = float64(minLatency) / 1000
		result.MaxLatencyMs = float64(maxLatency) / 1000

		// Sort for percentiles
		sortInt64s(latencies)
		result.P50LatencyMs = float64(latencies[len(latencies)*50/100]) / 1000
		result.P95LatencyMs = float64(latencies[len(latencies)*95/100]) / 1000
		result.P99LatencyMs = float64(latencies[len(latencies)*99/100]) / 1000
	}

	return result
}

// SimulatePostbacks simulates high-volume postback traffic
// GET /api/internal/stress/postbacks?count=500&concurrency=20
func (h *StressTestHandler) SimulatePostbacks(c *gin.Context) {
	count := 500
	if countStr := c.Query("count"); countStr != "" {
		if parsed, err := strconv.Atoi(countStr); err == nil && parsed > 0 && parsed <= 10000 {
			count = parsed
		}
	}

	concurrency := 20
	if concStr := c.Query("concurrency"); concStr != "" {
		if parsed, err := strconv.Atoi(concStr); err == nil && parsed > 0 && parsed <= 100 {
			concurrency = parsed
		}
	}

	result := h.runPostbackStressTest(count, concurrency)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"result":  result,
	})
}

// runPostbackStressTest executes the postback stress test
func (h *StressTestHandler) runPostbackStressTest(count, concurrency int) StressTestResult {
	result := StressTestResult{
		TestType:  "postback_simulation",
		Timestamp: time.Now(),
	}

	var (
		successCount int64
		failureCount int64
		totalLatency int64
		latencies    []int64
		latencyMutex sync.Mutex
	)

	work := make(chan int, count)
	for i := 0; i < count; i++ {
		work <- i
	}
	close(work)

	start := time.Now()
	var wg sync.WaitGroup

	for w := 0; w < concurrency; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for range work {
				postbackStart := time.Now()

				// Simulate postback processing (just logging for now)
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

	result.TotalRequests = int64(count)
	result.SuccessCount = successCount
	result.FailureCount = failureCount
	result.Duration = duration
	result.RequestsPerSec = float64(count) / duration.Seconds()

	if len(latencies) > 0 {
		result.AvgLatencyMs = float64(totalLatency) / float64(len(latencies)) / 1000
		sortInt64s(latencies)
		result.P50LatencyMs = float64(latencies[len(latencies)*50/100]) / 1000
		result.P95LatencyMs = float64(latencies[len(latencies)*95/100]) / 1000
		result.P99LatencyMs = float64(latencies[len(latencies)*99/100]) / 1000
	}

	return result
}

// GetWorkerPoolStats returns worker pool statistics
// GET /api/internal/stress/pools
func (h *StressTestHandler) GetWorkerPoolStats(c *gin.Context) {
	stats := services.GetAllPoolStats()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"pools":   stats,
	})
}

// GetCacheStats returns cache statistics
// GET /api/internal/stress/cache
func (h *StressTestHandler) GetCacheStats(c *gin.Context) {
	cacheService := services.NewCacheService()
	stats := cacheService.GetCacheStats()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"cache":   stats,
	})
}

// RunFullStressTest runs a comprehensive stress test
// GET /api/internal/stress/full
func (h *StressTestHandler) RunFullStressTest(c *gin.Context) {
	results := make(map[string]interface{})

	// Test 1: Burst clicks (2000 in 5 seconds)
	results["burst_clicks"] = h.runClickStressTest(2000, 200, generateTestIDs(5), true, true)

	// Test 2: Sustained clicks (10000 over 1 minute simulation)
	results["sustained_clicks"] = h.runClickStressTest(10000, 100, generateTestIDs(10), false, true)

	// Test 3: Postbacks
	results["postbacks"] = h.runPostbackStressTest(500, 50)

	// Get final stats
	results["worker_pools"] = services.GetAllPoolStats()
	results["cache"] = services.NewCacheService().GetCacheStats()

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"results":   results,
		"timestamp": time.Now(),
	})
}

// Helper functions

func generateTestIDs(count int) []uuid.UUID {
	ids := make([]uuid.UUID, count)
	for i := range ids {
		ids[i] = uuid.New()
	}
	return ids
}

func sortInt64s(arr []int64) {
	// Simple insertion sort for small arrays
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


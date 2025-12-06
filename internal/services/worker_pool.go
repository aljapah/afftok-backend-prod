package services

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

// ============================================
// TASK TYPES
// ============================================

// TaskType represents the type of async task
type TaskType string

const (
	TaskTypeClickTracking    TaskType = "click_tracking"
	TaskTypePostbackProcess  TaskType = "postback_process"
	TaskTypeAnalyticsUpdate  TaskType = "analytics_update"
	TaskTypeFraudCheck       TaskType = "fraud_check"
	TaskTypeStatsAggregation TaskType = "stats_aggregation"
	TaskTypeLogPersist       TaskType = "log_persist"
	TaskTypeCacheUpdate      TaskType = "cache_update"
)

// Task represents a unit of work
type Task struct {
	ID        string
	Type      TaskType
	Priority  int // 0 = low, 1 = normal, 2 = high
	Data      interface{}
	CreatedAt time.Time
	Retries   int
	MaxRetries int
}

// TaskResult represents the result of a task execution
type TaskResult struct {
	TaskID    string
	Success   bool
	Error     error
	Duration  time.Duration
	Timestamp time.Time
}

// TaskHandler is a function that processes a task
type TaskHandler func(ctx context.Context, task *Task) error

// ============================================
// WORKER POOL
// ============================================

// WorkerPool manages a pool of workers for async processing
type WorkerPool struct {
	name           string
	numWorkers     int
	taskQueue      chan *Task
	highPriorityQ  chan *Task
	results        chan *TaskResult
	handlers       map[TaskType]TaskHandler
	wg             sync.WaitGroup
	ctx            context.Context
	cancel         context.CancelFunc
	mutex          sync.RWMutex

	// Metrics
	tasksProcessed int64
	tasksSucceeded int64
	tasksFailed    int64
	tasksDropped   int64
	avgProcessTime int64
	processCount   int64

	// Config
	maxQueueSize   int
	taskTimeout    time.Duration
}

// WorkerPoolConfig holds configuration for the worker pool
type WorkerPoolConfig struct {
	Name          string
	NumWorkers    int
	MaxQueueSize  int
	TaskTimeout   time.Duration
}

// DefaultWorkerPoolConfig returns default configuration
func DefaultWorkerPoolConfig(name string) WorkerPoolConfig {
	numCPU := runtime.NumCPU()
	return WorkerPoolConfig{
		Name:         name,
		NumWorkers:   numCPU * 2,
		MaxQueueSize: 10000,
		TaskTimeout:  30 * time.Second,
	}
}

// NewWorkerPool creates a new worker pool
func NewWorkerPool(config WorkerPoolConfig) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	
	pool := &WorkerPool{
		name:          config.Name,
		numWorkers:    config.NumWorkers,
		taskQueue:     make(chan *Task, config.MaxQueueSize),
		highPriorityQ: make(chan *Task, config.MaxQueueSize/10),
		results:       make(chan *TaskResult, 1000),
		handlers:      make(map[TaskType]TaskHandler),
		ctx:           ctx,
		cancel:        cancel,
		maxQueueSize:  config.MaxQueueSize,
		taskTimeout:   config.TaskTimeout,
	}

	return pool
}

// RegisterHandler registers a handler for a task type
func (p *WorkerPool) RegisterHandler(taskType TaskType, handler TaskHandler) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	p.handlers[taskType] = handler
}

// Start starts the worker pool
func (p *WorkerPool) Start() {
	fmt.Printf("[WorkerPool:%s] Starting %d workers\n", p.name, p.numWorkers)
	
	for i := 0; i < p.numWorkers; i++ {
		p.wg.Add(1)
		go p.worker(i)
	}

	// Start results collector
	go p.collectResults()
}

// Stop gracefully stops the worker pool
func (p *WorkerPool) Stop() {
	fmt.Printf("[WorkerPool:%s] Stopping...\n", p.name)
	p.cancel()
	
	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		fmt.Printf("[WorkerPool:%s] All workers stopped\n", p.name)
	case <-time.After(10 * time.Second):
		fmt.Printf("[WorkerPool:%s] Timeout waiting for workers\n", p.name)
	}

	close(p.taskQueue)
	close(p.highPriorityQ)
}

// Submit submits a task to the pool
func (p *WorkerPool) Submit(task *Task) bool {
	if task.ID == "" {
		task.ID = fmt.Sprintf("%s-%d", task.Type, time.Now().UnixNano())
	}
	task.CreatedAt = time.Now()
	
	// Try high priority queue first for high priority tasks
	if task.Priority >= 2 {
		select {
		case p.highPriorityQ <- task:
			return true
		default:
			// Fall through to normal queue
		}
	}

	// Try normal queue
	select {
	case p.taskQueue <- task:
		return true
	default:
		// Queue is full
		atomic.AddInt64(&p.tasksDropped, 1)
		fmt.Printf("[WorkerPool:%s] Task dropped (queue full): %s\n", p.name, task.ID)
		return false
	}
}

// SubmitAsync submits a task and returns immediately
func (p *WorkerPool) SubmitAsync(taskType TaskType, data interface{}, priority int) {
	task := &Task{
		Type:       taskType,
		Priority:   priority,
		Data:       data,
		MaxRetries: 3,
	}
	p.Submit(task)
}

// worker is the main worker loop
func (p *WorkerPool) worker(id int) {
	defer p.wg.Done()
	
	for {
		select {
		case <-p.ctx.Done():
			return
		case task := <-p.highPriorityQ:
			if task != nil {
				p.processTask(id, task)
			}
		case task := <-p.taskQueue:
			if task != nil {
				p.processTask(id, task)
			}
		}
	}
}

// processTask processes a single task
func (p *WorkerPool) processTask(workerID int, task *Task) {
	start := time.Now()
	
	// Get handler
	p.mutex.RLock()
	handler, exists := p.handlers[task.Type]
	p.mutex.RUnlock()

	result := &TaskResult{
		TaskID:    task.ID,
		Timestamp: time.Now(),
	}

	if !exists {
		result.Success = false
		result.Error = fmt.Errorf("no handler for task type: %s", task.Type)
		result.Duration = time.Since(start)
		p.results <- result
		atomic.AddInt64(&p.tasksFailed, 1)
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(p.ctx, p.taskTimeout)
	defer cancel()

	// Execute handler
	err := handler(ctx, task)
	result.Duration = time.Since(start)

	if err != nil {
		result.Success = false
		result.Error = err
		
		// Retry if possible
		if task.Retries < task.MaxRetries {
			task.Retries++
			p.Submit(task)
		} else {
			atomic.AddInt64(&p.tasksFailed, 1)
		}
	} else {
		result.Success = true
		atomic.AddInt64(&p.tasksSucceeded, 1)
	}

	atomic.AddInt64(&p.tasksProcessed, 1)
	
	// Update average processing time
	atomic.AddInt64(&p.avgProcessTime, result.Duration.Microseconds())
	atomic.AddInt64(&p.processCount, 1)

	p.results <- result
}

// collectResults collects and logs results
func (p *WorkerPool) collectResults() {
	for {
		select {
		case <-p.ctx.Done():
			return
		case result := <-p.results:
			if result != nil && !result.Success && result.Error != nil {
				fmt.Printf("[WorkerPool:%s] Task %s failed: %v\n", p.name, result.TaskID, result.Error)
			}
		}
	}
}

// GetStats returns pool statistics
func (p *WorkerPool) GetStats() map[string]interface{} {
	processCount := atomic.LoadInt64(&p.processCount)
	avgTime := int64(0)
	if processCount > 0 {
		avgTime = atomic.LoadInt64(&p.avgProcessTime) / processCount
	}

	return map[string]interface{}{
		"name":              p.name,
		"num_workers":       p.numWorkers,
		"queue_size":        len(p.taskQueue),
		"high_priority_size": len(p.highPriorityQ),
		"max_queue_size":    p.maxQueueSize,
		"tasks_processed":   atomic.LoadInt64(&p.tasksProcessed),
		"tasks_succeeded":   atomic.LoadInt64(&p.tasksSucceeded),
		"tasks_failed":      atomic.LoadInt64(&p.tasksFailed),
		"tasks_dropped":     atomic.LoadInt64(&p.tasksDropped),
		"avg_process_time_us": avgTime,
	}
}

// QueueLength returns current queue length
func (p *WorkerPool) QueueLength() int {
	return len(p.taskQueue) + len(p.highPriorityQ)
}

// ============================================
// GLOBAL WORKER POOLS
// ============================================

var (
	clickPool     *WorkerPool
	postbackPool  *WorkerPool
	analyticsPool *WorkerPool
	loggingPool   *WorkerPool
	poolsOnce     sync.Once
)

// InitWorkerPools initializes all global worker pools
func InitWorkerPools() {
	poolsOnce.Do(func() {
		// Click processing pool - high throughput
		clickConfig := DefaultWorkerPoolConfig("clicks")
		clickConfig.NumWorkers = runtime.NumCPU() * 4
		clickConfig.MaxQueueSize = 50000
		clickPool = NewWorkerPool(clickConfig)

		// Postback processing pool
		postbackConfig := DefaultWorkerPoolConfig("postbacks")
		postbackConfig.NumWorkers = runtime.NumCPU() * 2
		postbackConfig.MaxQueueSize = 10000
		postbackPool = NewWorkerPool(postbackConfig)

		// Analytics aggregation pool
		analyticsConfig := DefaultWorkerPoolConfig("analytics")
		analyticsConfig.NumWorkers = runtime.NumCPU()
		analyticsConfig.MaxQueueSize = 5000
		analyticsPool = NewWorkerPool(analyticsConfig)

		// Logging pool
		loggingConfig := DefaultWorkerPoolConfig("logging")
		loggingConfig.NumWorkers = runtime.NumCPU()
		loggingConfig.MaxQueueSize = 20000
		loggingPool = NewWorkerPool(loggingConfig)

		fmt.Println("[WorkerPools] All pools initialized")
	})
}

// GetClickPool returns the click processing pool
func GetClickPool() *WorkerPool {
	return clickPool
}

// GetPostbackPool returns the postback processing pool
func GetPostbackPool() *WorkerPool {
	return postbackPool
}

// GetAnalyticsPool returns the analytics pool
func GetAnalyticsPool() *WorkerPool {
	return analyticsPool
}

// GetLoggingPool returns the logging pool
func GetLoggingPool() *WorkerPool {
	return loggingPool
}

// StartAllPools starts all worker pools
func StartAllPools() {
	InitWorkerPools()
	
	if clickPool != nil {
		clickPool.Start()
	}
	if postbackPool != nil {
		postbackPool.Start()
	}
	if analyticsPool != nil {
		analyticsPool.Start()
	}
	if loggingPool != nil {
		loggingPool.Start()
	}
}

// StopAllPools stops all worker pools
func StopAllPools() {
	if clickPool != nil {
		clickPool.Stop()
	}
	if postbackPool != nil {
		postbackPool.Stop()
	}
	if analyticsPool != nil {
		analyticsPool.Stop()
	}
	if loggingPool != nil {
		loggingPool.Stop()
	}
}

// GetAllPoolStats returns stats for all pools
func GetAllPoolStats() map[string]interface{} {
	stats := make(map[string]interface{})
	
	if clickPool != nil {
		stats["clicks"] = clickPool.GetStats()
	}
	if postbackPool != nil {
		stats["postbacks"] = postbackPool.GetStats()
	}
	if analyticsPool != nil {
		stats["analytics"] = analyticsPool.GetStats()
	}
	if loggingPool != nil {
		stats["logging"] = loggingPool.GetStats()
	}
	
	return stats
}


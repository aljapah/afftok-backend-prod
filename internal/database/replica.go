package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ============================================
// READ REPLICA SUPPORT
// ============================================

// DBRouter manages primary and replica database connections
type DBRouter struct {
	primary       *gorm.DB
	replica       *gorm.DB
	replicaReady  atomic.Bool
	
	// Metrics
	primaryReads  int64
	replicaReads  int64
	primaryWrites int64
	replicaFails  int64
	
	// Health check
	healthTicker  *time.Ticker
	ctx           context.Context
	cancel        context.CancelFunc
	
	mutex         sync.RWMutex
}

var (
	dbRouter     *DBRouter
	dbRouterOnce sync.Once
)

// GetDBRouter returns the singleton DB router
func GetDBRouter() *DBRouter {
	return dbRouter
}

// InitDBRouter initializes the DB router with primary and optional replica
func InitDBRouter(cfg *config.Config) (*DBRouter, error) {
	var initErr error
	
	dbRouterOnce.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		
		dbRouter = &DBRouter{
			ctx:    ctx,
			cancel: cancel,
		}
		
		// Connect to primary
		primary, err := connectDB(cfg.PostgresURL, "primary")
		if err != nil {
			initErr = fmt.Errorf("failed to connect to primary DB: %w", err)
			return
		}
		dbRouter.primary = primary
		DB = primary // Keep global DB for backward compatibility
		
		// Connect to replica if configured
		replicaURL := os.Getenv("POSTGRES_REPLICA_URL")
		if replicaURL != "" {
			replica, err := connectDB(replicaURL, "replica")
			if err != nil {
				log.Printf("⚠️ Failed to connect to replica DB: %v (using primary for reads)", err)
			} else {
				dbRouter.replica = replica
				dbRouter.replicaReady.Store(true)
				log.Println("✅ Read replica connected successfully")
			}
		} else {
			log.Println("ℹ️ No replica configured (POSTGRES_REPLICA_URL not set)")
		}
		
		// Start health check
		dbRouter.startHealthCheck()
	})
	
	return dbRouter, initErr
}

// connectDB creates a new database connection
func connectDB(dsn, name string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("%s DSN is empty", name)
	}

	gormLogger := logger.Default.LogMode(logger.Error)
	if os.Getenv("ENVIRONMENT") != "production" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Apply connection pool settings
	numCPU := runtime.NumCPU()
	sqlDB.SetMaxIdleConns(numCPU * 5)
	sqlDB.SetMaxOpenConns(numCPU * 25)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Printf("✅ %s DB connected (max_open=%d, max_idle=%d)", 
		name, numCPU*25, numCPU*5)
	
	return db, nil
}

// ============================================
// READ/WRITE ROUTING
// ============================================

// Write returns the primary DB for write operations
func (r *DBRouter) Write() *gorm.DB {
	atomic.AddInt64(&r.primaryWrites, 1)
	return r.primary
}

// Read returns the best available DB for read operations
// Uses replica if available, falls back to primary
func (r *DBRouter) Read() *gorm.DB {
	if r.replicaReady.Load() && r.replica != nil {
		atomic.AddInt64(&r.replicaReads, 1)
		return r.replica
	}
	
	atomic.AddInt64(&r.primaryReads, 1)
	return r.primary
}

// Primary returns the primary DB directly
func (r *DBRouter) Primary() *gorm.DB {
	return r.primary
}

// Replica returns the replica DB (may be nil)
func (r *DBRouter) Replica() *gorm.DB {
	return r.replica
}

// HasReplica returns true if a replica is configured and healthy
func (r *DBRouter) HasReplica() bool {
	return r.replicaReady.Load()
}

// ============================================
// HEALTH CHECK
// ============================================

// startHealthCheck starts periodic health checks
func (r *DBRouter) startHealthCheck() {
	r.healthTicker = time.NewTicker(30 * time.Second)
	
	go func() {
		for {
			select {
			case <-r.ctx.Done():
				return
			case <-r.healthTicker.C:
				r.checkHealth()
			}
		}
	}()
}

// checkHealth checks the health of both databases
func (r *DBRouter) checkHealth() {
	// Check primary
	if r.primary != nil {
		if sqlDB, err := r.primary.DB(); err == nil {
			if err := sqlDB.Ping(); err != nil {
				log.Printf("⚠️ Primary DB health check failed: %v", err)
			}
		}
	}
	
	// Check replica
	if r.replica != nil {
		if sqlDB, err := r.replica.DB(); err == nil {
			if err := sqlDB.Ping(); err != nil {
				log.Printf("⚠️ Replica DB health check failed: %v", err)
				r.replicaReady.Store(false)
				atomic.AddInt64(&r.replicaFails, 1)
			} else {
				if !r.replicaReady.Load() {
					log.Println("✅ Replica DB recovered")
				}
				r.replicaReady.Store(true)
			}
		}
	}
}

// ============================================
// METRICS
// ============================================

// ReplicaStatus represents the status of the replica
type ReplicaStatus struct {
	Configured    bool    `json:"configured"`
	Healthy       bool    `json:"healthy"`
	PrimaryReads  int64   `json:"primary_reads"`
	ReplicaReads  int64   `json:"replica_reads"`
	PrimaryWrites int64   `json:"primary_writes"`
	ReplicaFails  int64   `json:"replica_fails"`
	ReplicaRatio  float64 `json:"replica_ratio_percent"`
}

// GetReplicaStatus returns the current replica status
func (r *DBRouter) GetReplicaStatus() ReplicaStatus {
	primaryReads := atomic.LoadInt64(&r.primaryReads)
	replicaReads := atomic.LoadInt64(&r.replicaReads)
	totalReads := primaryReads + replicaReads
	
	replicaRatio := float64(0)
	if totalReads > 0 {
		replicaRatio = float64(replicaReads) / float64(totalReads) * 100
	}
	
	return ReplicaStatus{
		Configured:    r.replica != nil,
		Healthy:       r.replicaReady.Load(),
		PrimaryReads:  primaryReads,
		ReplicaReads:  replicaReads,
		PrimaryWrites: atomic.LoadInt64(&r.primaryWrites),
		ReplicaFails:  atomic.LoadInt64(&r.replicaFails),
		ReplicaRatio:  replicaRatio,
	}
}

// ============================================
// LATENCY TRACKING
// ============================================

// DBLatencyTracker tracks database operation latencies
type DBLatencyTracker struct {
	mutex         sync.RWMutex
	readLatencies []int64
	writeLatencies []int64
	maxSamples    int
}

var latencyTracker = &DBLatencyTracker{
	readLatencies:  make([]int64, 0, 1000),
	writeLatencies: make([]int64, 0, 1000),
	maxSamples:     1000,
}

// RecordReadLatency records a read operation latency
func RecordReadLatency(latencyMs int64) {
	latencyTracker.mutex.Lock()
	defer latencyTracker.mutex.Unlock()
	
	if len(latencyTracker.readLatencies) >= latencyTracker.maxSamples {
		latencyTracker.readLatencies = latencyTracker.readLatencies[100:]
	}
	latencyTracker.readLatencies = append(latencyTracker.readLatencies, latencyMs)
}

// RecordWriteLatency records a write operation latency
func RecordWriteLatency(latencyMs int64) {
	latencyTracker.mutex.Lock()
	defer latencyTracker.mutex.Unlock()
	
	if len(latencyTracker.writeLatencies) >= latencyTracker.maxSamples {
		latencyTracker.writeLatencies = latencyTracker.writeLatencies[100:]
	}
	latencyTracker.writeLatencies = append(latencyTracker.writeLatencies, latencyMs)
}

// LatencyStats represents latency statistics
type LatencyStats struct {
	AvgMs float64 `json:"avg_ms"`
	MinMs int64   `json:"min_ms"`
	MaxMs int64   `json:"max_ms"`
	P50Ms int64   `json:"p50_ms"`
	P95Ms int64   `json:"p95_ms"`
	P99Ms int64   `json:"p99_ms"`
	Count int     `json:"sample_count"`
}

// GetLatencyStats returns latency statistics
func GetLatencyStats() map[string]LatencyStats {
	latencyTracker.mutex.RLock()
	defer latencyTracker.mutex.RUnlock()
	
	return map[string]LatencyStats{
		"read":  calculateStats(latencyTracker.readLatencies),
		"write": calculateStats(latencyTracker.writeLatencies),
	}
}

func calculateStats(samples []int64) LatencyStats {
	if len(samples) == 0 {
		return LatencyStats{}
	}
	
	// Copy and sort
	sorted := make([]int64, len(samples))
	copy(sorted, samples)
	sortInt64s(sorted)
	
	var sum int64
	for _, v := range sorted {
		sum += v
	}
	
	return LatencyStats{
		AvgMs: float64(sum) / float64(len(sorted)),
		MinMs: sorted[0],
		MaxMs: sorted[len(sorted)-1],
		P50Ms: sorted[len(sorted)*50/100],
		P95Ms: sorted[len(sorted)*95/100],
		P99Ms: sorted[len(sorted)*99/100],
		Count: len(sorted),
	}
}

func sortInt64s(arr []int64) {
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

// ============================================
// CLEANUP
// ============================================

// Close closes all database connections
func (r *DBRouter) Close() error {
	r.cancel()
	
	if r.healthTicker != nil {
		r.healthTicker.Stop()
	}
	
	var errs []error
	
	if r.primary != nil {
		if sqlDB, err := r.primary.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	
	if r.replica != nil {
		if sqlDB, err := r.replica.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				errs = append(errs, err)
			}
		}
	}
	
	if len(errs) > 0 {
		return fmt.Errorf("errors closing DBs: %v", errs)
	}
	
	return nil
}


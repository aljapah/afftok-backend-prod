package services

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
)

// ============================================
// WAL EVENT TYPES
// ============================================

// WALEventType represents the type of WAL event
type WALEventType string

const (
	WALEventClick      WALEventType = "click"
	WALEventConversion WALEventType = "conversion"
	WALEventPostback   WALEventType = "postback"
	WALEventAPIEvent   WALEventType = "api_event"
	WALEventEdgeEvent  WALEventType = "edge_event"
)

// WALEntryStatus represents the status of a WAL entry
type WALEntryStatus string

const (
	WALStatusPending   WALEntryStatus = "pending"
	WALStatusProcessed WALEntryStatus = "processed"
	WALStatusFailed    WALEntryStatus = "failed"
	WALStatusReplayed  WALEntryStatus = "replayed"
)

// ============================================
// WAL ENTRY
// ============================================

// WALEntry represents a single entry in the Write-Ahead Log
type WALEntry struct {
	ID          string                 `json:"id"`
	Sequence    int64                  `json:"seq"`
	EventType   WALEventType           `json:"type"`
	Status      WALEntryStatus         `json:"status"`
	TenantID    string                 `json:"tenant_id,omitempty"`
	Timestamp   time.Time              `json:"ts"`
	Data        map[string]interface{} `json:"data"`
	Checksum    string                 `json:"checksum"`
	Attempts    int                    `json:"attempts"`
	LastAttempt *time.Time             `json:"last_attempt,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// ============================================
// WAL SERVICE
// ============================================

// WALService provides Write-Ahead Logging functionality
type WALService struct {
	mu            sync.RWMutex
	walDir        string
	currentFile   *os.File
	currentPath   string
	sequence      int64
	maxFileSize   int64
	maxEntries    int64
	flushInterval time.Duration
	
	// Metrics
	totalEntries    int64
	pendingEntries  int64
	processedCount  int64
	failedCount     int64
	replayedCount   int64
	corruptionCount int64
	
	// State
	isRunning     bool
	stopChan      chan struct{}
	flushChan     chan struct{}
	wg            sync.WaitGroup
	
	// Callbacks
	onProcess     func(entry *WALEntry) error
	observability *ObservabilityService
}

// WALConfig holds WAL configuration
type WALConfig struct {
	Dir           string
	MaxFileSize   int64 // bytes
	MaxEntries    int64
	FlushInterval time.Duration
}

// DefaultWALConfig returns default WAL configuration
func DefaultWALConfig() WALConfig {
	return WALConfig{
		Dir:           "./data/wal",
		MaxFileSize:   100 * 1024 * 1024, // 100MB
		MaxEntries:    500000,
		FlushInterval: 100 * time.Millisecond,
	}
}

// NewWALService creates a new WAL service
func NewWALService(config WALConfig) (*WALService, error) {
	// Create WAL directory
	if err := os.MkdirAll(config.Dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create WAL directory: %w", err)
	}

	service := &WALService{
		walDir:        config.Dir,
		maxFileSize:   config.MaxFileSize,
		maxEntries:    config.MaxEntries,
		flushInterval: config.FlushInterval,
		stopChan:      make(chan struct{}),
		flushChan:     make(chan struct{}, 1),
		observability: NewObservabilityService(),
	}

	// Initialize current WAL file
	if err := service.initCurrentFile(); err != nil {
		return nil, err
	}

	// Load sequence from existing files
	if err := service.loadSequence(); err != nil {
		return nil, err
	}

	return service, nil
}

// ============================================
// WAL OPERATIONS
// ============================================

// Append appends an entry to the WAL
func (w *WALService) Append(eventType WALEventType, tenantID string, data map[string]interface{}) (*WALEntry, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check max entries
	if atomic.LoadInt64(&w.pendingEntries) >= w.maxEntries {
		return nil, fmt.Errorf("WAL buffer full: %d entries", w.maxEntries)
	}

	// Create entry
	seq := atomic.AddInt64(&w.sequence, 1)
	entry := &WALEntry{
		ID:        uuid.New().String(),
		Sequence:  seq,
		EventType: eventType,
		Status:    WALStatusPending,
		TenantID:  tenantID,
		Timestamp: time.Now().UTC(),
		Data:      data,
		Attempts:  0,
	}

	// Calculate checksum
	entry.Checksum = w.calculateChecksum(entry)

	// Serialize entry
	line, err := json.Marshal(entry)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize WAL entry: %w", err)
	}

	// Check if we need to rotate
	if err := w.checkRotation(); err != nil {
		return nil, err
	}

	// Write to file with newline
	line = append(line, '\n')
	if _, err := w.currentFile.Write(line); err != nil {
		return nil, fmt.Errorf("failed to write WAL entry: %w", err)
	}

	// Fsync for durability
	if err := w.currentFile.Sync(); err != nil {
		return nil, fmt.Errorf("failed to sync WAL: %w", err)
	}

	atomic.AddInt64(&w.totalEntries, 1)
	atomic.AddInt64(&w.pendingEntries, 1)

	return entry, nil
}

// MarkProcessed marks an entry as processed
func (w *WALService) MarkProcessed(entryID string) error {
	return w.updateEntryStatus(entryID, WALStatusProcessed, "")
}

// MarkFailed marks an entry as failed
func (w *WALService) MarkFailed(entryID string, errorMsg string) error {
	return w.updateEntryStatus(entryID, WALStatusFailed, errorMsg)
}

// updateEntryStatus updates the status of a WAL entry
func (w *WALService) updateEntryStatus(entryID string, status WALEntryStatus, errorMsg string) error {
	// In a production system, we'd update the entry in place or use a separate status file
	// For simplicity, we track status in memory and compact periodically
	
	switch status {
	case WALStatusProcessed:
		atomic.AddInt64(&w.processedCount, 1)
		atomic.AddInt64(&w.pendingEntries, -1)
	case WALStatusFailed:
		atomic.AddInt64(&w.failedCount, 1)
	case WALStatusReplayed:
		atomic.AddInt64(&w.replayedCount, 1)
	}

	return nil
}

// ============================================
// FILE MANAGEMENT
// ============================================

// initCurrentFile initializes the current WAL file
func (w *WALService) initCurrentFile() error {
	filename := fmt.Sprintf("wal_%s.log", time.Now().Format("20060102_150405"))
	path := filepath.Join(w.walDir, filename)

	file, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open WAL file: %w", err)
	}

	w.currentFile = file
	w.currentPath = path
	return nil
}

// checkRotation checks if we need to rotate the WAL file
func (w *WALService) checkRotation() error {
	info, err := w.currentFile.Stat()
	if err != nil {
		return err
	}

	if info.Size() >= w.maxFileSize {
		return w.rotate()
	}

	return nil
}

// rotate rotates to a new WAL file
func (w *WALService) rotate() error {
	// Close current file
	if err := w.currentFile.Sync(); err != nil {
		return err
	}
	if err := w.currentFile.Close(); err != nil {
		return err
	}

	// Create new file
	return w.initCurrentFile()
}

// loadSequence loads the highest sequence number from existing WAL files
func (w *WALService) loadSequence() error {
	files, err := w.listWALFiles()
	if err != nil {
		return err
	}

	var maxSeq int64
	for _, file := range files {
		entries, err := w.readWALFile(file)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.Sequence > maxSeq {
				maxSeq = entry.Sequence
			}
		}
	}

	w.sequence = maxSeq
	return nil
}

// listWALFiles lists all WAL files in the directory
func (w *WALService) listWALFiles() ([]string, error) {
	entries, err := os.ReadDir(w.walDir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "wal_") && strings.HasSuffix(entry.Name(), ".log") {
			files = append(files, filepath.Join(w.walDir, entry.Name()))
		}
	}

	sort.Strings(files)
	return files, nil
}

// readWALFile reads all entries from a WAL file
func (w *WALService) readWALFile(path string) ([]*WALEntry, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries []*WALEntry
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 1024*1024), 10*1024*1024) // 10MB max line

	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var entry WALEntry
		if err := json.Unmarshal(line, &entry); err != nil {
			atomic.AddInt64(&w.corruptionCount, 1)
			w.observability.Log(LogEvent{
				Category: LogCategoryErrorEvent,
				Level:    LogLevelError,
				Message:  "WAL corruption detected",
				Metadata: map[string]interface{}{
					"file":    path,
					"line":    lineNum,
					"error":   err.Error(),
				},
			})
			continue
		}

		// Verify checksum
		if !w.verifyChecksum(&entry) {
			atomic.AddInt64(&w.corruptionCount, 1)
			w.observability.Log(LogEvent{
				Category: LogCategoryErrorEvent,
				Level:    LogLevelError,
				Message:  "WAL checksum mismatch",
				Metadata: map[string]interface{}{
					"file":     path,
					"entry_id": entry.ID,
				},
			})
			continue
		}

		entries = append(entries, &entry)
	}

	return entries, scanner.Err()
}

// ============================================
// CHECKSUM
// ============================================

// calculateChecksum calculates SHA256 checksum for an entry
func (w *WALService) calculateChecksum(entry *WALEntry) string {
	data := fmt.Sprintf("%s|%d|%s|%s|%s",
		entry.ID,
		entry.Sequence,
		entry.EventType,
		entry.TenantID,
		entry.Timestamp.Format(time.RFC3339Nano),
	)
	
	if entry.Data != nil {
		dataJSON, _ := json.Marshal(entry.Data)
		data += "|" + string(dataJSON)
	}

	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:8]) // First 8 bytes
}

// verifyChecksum verifies the checksum of an entry
func (w *WALService) verifyChecksum(entry *WALEntry) bool {
	savedChecksum := entry.Checksum
	entry.Checksum = ""
	calculated := w.calculateChecksum(entry)
	entry.Checksum = savedChecksum
	return savedChecksum == calculated
}

// ============================================
// REPLAY & RECOVERY
// ============================================

// Replay replays all pending entries
func (w *WALService) Replay(processor func(entry *WALEntry) error) (int, int, error) {
	files, err := w.listWALFiles()
	if err != nil {
		return 0, 0, err
	}

	var replayed, failed int

	for _, file := range files {
		entries, err := w.readWALFile(file)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.Status != WALStatusPending {
				continue
			}

			if err := processor(entry); err != nil {
				entry.Status = WALStatusFailed
				entry.Error = err.Error()
				failed++
			} else {
				entry.Status = WALStatusReplayed
				replayed++
			}
		}
	}

	atomic.AddInt64(&w.replayedCount, int64(replayed))
	return replayed, failed, nil
}

// GetPendingEntries returns all pending entries
func (w *WALService) GetPendingEntries() ([]*WALEntry, error) {
	files, err := w.listWALFiles()
	if err != nil {
		return nil, err
	}

	var pending []*WALEntry

	for _, file := range files {
		entries, err := w.readWALFile(file)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.Status == WALStatusPending {
				pending = append(pending, entry)
			}
		}
	}

	return pending, nil
}

// ============================================
// COMPACTION
// ============================================

// Compact removes processed entries from WAL files
func (w *WALService) Compact() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	files, err := w.listWALFiles()
	if err != nil {
		return err
	}

	for _, file := range files {
		// Skip current file
		if file == w.currentPath {
			continue
		}

		entries, err := w.readWALFile(file)
		if err != nil {
			continue
		}

		// Check if all entries are processed
		allProcessed := true
		for _, entry := range entries {
			if entry.Status == WALStatusPending {
				allProcessed = false
				break
			}
		}

		// Remove file if all entries are processed
		if allProcessed {
			os.Remove(file)
		}
	}

	return nil
}

// ============================================
// LIFECYCLE
// ============================================

// Start starts the WAL service background workers
func (w *WALService) Start() {
	w.isRunning = true

	// Flush worker
	w.wg.Add(1)
	go w.flushWorker()

	// Compaction worker
	w.wg.Add(1)
	go w.compactionWorker()
}

// Stop stops the WAL service
func (w *WALService) Stop() error {
	w.isRunning = false
	close(w.stopChan)
	w.wg.Wait()

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.currentFile != nil {
		w.currentFile.Sync()
		return w.currentFile.Close()
	}
	return nil
}

// flushWorker periodically flushes the WAL
func (w *WALService) flushWorker() {
	defer w.wg.Done()

	ticker := time.NewTicker(w.flushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.mu.Lock()
			if w.currentFile != nil {
				w.currentFile.Sync()
			}
			w.mu.Unlock()
		case <-w.flushChan:
			w.mu.Lock()
			if w.currentFile != nil {
				w.currentFile.Sync()
			}
			w.mu.Unlock()
		}
	}
}

// compactionWorker periodically compacts WAL files
func (w *WALService) compactionWorker() {
	defer w.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.Compact()
		}
	}
}

// Flush forces a sync
func (w *WALService) Flush() {
	select {
	case w.flushChan <- struct{}{}:
	default:
	}
}

// ============================================
// METRICS
// ============================================

// GetStats returns WAL statistics
func (w *WALService) GetStats() map[string]interface{} {
	files, _ := w.listWALFiles()
	
	var totalSize int64
	for _, file := range files {
		if info, err := os.Stat(file); err == nil {
			totalSize += info.Size()
		}
	}

	return map[string]interface{}{
		"total_entries":      atomic.LoadInt64(&w.totalEntries),
		"pending_entries":    atomic.LoadInt64(&w.pendingEntries),
		"processed_count":    atomic.LoadInt64(&w.processedCount),
		"failed_count":       atomic.LoadInt64(&w.failedCount),
		"replayed_count":     atomic.LoadInt64(&w.replayedCount),
		"corruption_count":   atomic.LoadInt64(&w.corruptionCount),
		"current_sequence":   atomic.LoadInt64(&w.sequence),
		"file_count":         len(files),
		"total_size_bytes":   totalSize,
		"current_file":       filepath.Base(w.currentPath),
		"is_running":         w.isRunning,
	}
}

// ============================================
// GLOBAL INSTANCE
// ============================================

var (
	walServiceInstance *WALService
	walServiceOnce     sync.Once
)

// GetWALService returns the global WAL service instance
func GetWALService() *WALService {
	walServiceOnce.Do(func() {
		config := DefaultWALConfig()
		var err error
		walServiceInstance, err = NewWALService(config)
		if err != nil {
			panic(fmt.Sprintf("Failed to initialize WAL service: %v", err))
		}
	})
	return walServiceInstance
}


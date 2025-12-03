package services

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

// ============================================
// TABLE PARTITIONING SERVICE
// ============================================

// PartitionService manages table partitioning
type PartitionService struct {
	db *gorm.DB
}

// NewPartitionService creates a new partition service
func NewPartitionService(db *gorm.DB) *PartitionService {
	return &PartitionService{db: db}
}

// ============================================
// PARTITION INFO
// ============================================

// PartitionInfo represents information about a partition
type PartitionInfo struct {
	PartitionName   string    `json:"partition_name"`
	ParentTable     string    `json:"parent_table"`
	RangeStart      string    `json:"range_start"`
	RangeEnd        string    `json:"range_end"`
	RowCount        int64     `json:"row_count"`
	SizeMB          float64   `json:"size_mb"`
	IndexesSizeMB   float64   `json:"indexes_size_mb"`
	CreatedAt       time.Time `json:"created_at"`
	IsCurrentMonth  bool      `json:"is_current_month"`
}

// PartitionStatus represents the overall partitioning status
type PartitionStatus struct {
	TableName          string          `json:"table_name"`
	IsPartitioned      bool            `json:"is_partitioned"`
	PartitionKey       string          `json:"partition_key"`
	PartitionCount     int             `json:"partition_count"`
	TotalRowCount      int64           `json:"total_row_count"`
	TotalSizeMB        float64         `json:"total_size_mb"`
	Partitions         []PartitionInfo `json:"partitions"`
	NextPartitionName  string          `json:"next_partition_name"`
	NextPartitionRange string          `json:"next_partition_range"`
}

// GetPartitionStatus returns the partitioning status for a table
func (s *PartitionService) GetPartitionStatus(tableName string) (*PartitionStatus, error) {
	status := &PartitionStatus{
		TableName:  tableName,
		Partitions: make([]PartitionInfo, 0),
	}

	// Check if table is partitioned
	var partitionStrategy string
	err := s.db.Raw(`
		SELECT pt.partstrat
		FROM pg_class c
		JOIN pg_partitioned_table pt ON c.oid = pt.partrelid
		WHERE c.relname = ?
	`, tableName).Scan(&partitionStrategy).Error

	if err != nil || partitionStrategy == "" {
		status.IsPartitioned = false
		
		// Get current row count
		var count int64
		s.db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)).Scan(&count)
		status.TotalRowCount = count
		
		// Generate next partition suggestion
		now := time.Now().UTC()
		status.NextPartitionName = fmt.Sprintf("%s_%d_%02d", tableName, now.Year(), now.Month())
		status.NextPartitionRange = fmt.Sprintf("%d-%02d-01 to %d-%02d-01", 
			now.Year(), now.Month(), 
			now.Year(), now.Month()+1)
		
		return status, nil
	}

	status.IsPartitioned = true
	status.PartitionKey = "clicked_at" // Known from our schema

	// Get all partitions
	rows, err := s.db.Raw(`
		SELECT 
			c.relname as partition_name,
			pg_get_expr(c.relpartbound, c.oid, true) as partition_bound,
			pg_total_relation_size(c.oid) as total_bytes,
			pg_indexes_size(c.oid) as indexes_bytes
		FROM pg_inherits i
		JOIN pg_class c ON i.inhrelid = c.oid
		JOIN pg_class p ON i.inhparent = p.oid
		WHERE p.relname = ?
		ORDER BY c.relname
	`, tableName).Rows()
	
	if err != nil {
		return status, err
	}
	defer rows.Close()

	currentMonth := time.Now().UTC().Format("2006_01")
	
	for rows.Next() {
		var info PartitionInfo
		var partitionBound string
		var totalBytes, indexesBytes int64
		
		if err := rows.Scan(&info.PartitionName, &partitionBound, &totalBytes, &indexesBytes); err != nil {
			continue
		}
		
		info.ParentTable = tableName
		info.SizeMB = float64(totalBytes) / 1024 / 1024
		info.IndexesSizeMB = float64(indexesBytes) / 1024 / 1024
		info.IsCurrentMonth = containsMonth(info.PartitionName, currentMonth)
		
		// Parse partition bound to get range
		info.RangeStart = partitionBound // Simplified
		
		// Get row count for partition
		var count int64
		s.db.Raw(fmt.Sprintf("SELECT COUNT(*) FROM %s", info.PartitionName)).Scan(&count)
		info.RowCount = count
		
		status.Partitions = append(status.Partitions, info)
		status.TotalRowCount += count
		status.TotalSizeMB += info.SizeMB
	}

	status.PartitionCount = len(status.Partitions)

	// Generate next partition
	nextMonth := time.Now().UTC().AddDate(0, 1, 0)
	status.NextPartitionName = fmt.Sprintf("%s_%d_%02d", tableName, nextMonth.Year(), int(nextMonth.Month()))
	status.NextPartitionRange = fmt.Sprintf("%d-%02d-01 to %d-%02d-01",
		nextMonth.Year(), nextMonth.Month(),
		nextMonth.Year(), nextMonth.Month()+1)

	return status, nil
}

func containsMonth(name, month string) bool {
	return len(name) >= len(month) && name[len(name)-7:] == month
}

// ============================================
// PARTITION CREATION
// ============================================

// CreatePartitionResult represents the result of partition creation
type CreatePartitionResult struct {
	Success         bool   `json:"success"`
	PartitionName   string `json:"partition_name"`
	RangeStart      string `json:"range_start"`
	RangeEnd        string `json:"range_end"`
	SQLExecuted     string `json:"sql_executed"`
	Error           string `json:"error,omitempty"`
	AlreadyExists   bool   `json:"already_exists"`
}

// CreateMonthlyPartition creates a monthly partition for clicks table
func (s *PartitionService) CreateMonthlyPartition(year int, month int) (*CreatePartitionResult, error) {
	result := &CreatePartitionResult{
		Success: false,
	}

	// Validate input
	if year < 2020 || year > 2100 {
		result.Error = "Invalid year (must be 2020-2100)"
		return result, nil
	}
	if month < 1 || month > 12 {
		result.Error = "Invalid month (must be 1-12)"
		return result, nil
	}

	partitionName := fmt.Sprintf("clicks_%d_%02d", year, month)
	result.PartitionName = partitionName

	// Calculate range
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)
	
	result.RangeStart = startDate.Format("2006-01-02")
	result.RangeEnd = endDate.Format("2006-01-02")

	// Check if partition already exists
	var exists bool
	s.db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_class 
			WHERE relname = ? AND relkind = 'r'
		)
	`, partitionName).Scan(&exists)

	if exists {
		result.AlreadyExists = true
		result.Success = true
		result.SQLExecuted = "-- Partition already exists"
		return result, nil
	}

	// Check if parent table is partitioned
	var isPartitioned bool
	s.db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_partitioned_table pt
			JOIN pg_class c ON c.oid = pt.partrelid
			WHERE c.relname = 'clicks'
		)
	`).Scan(&isPartitioned)

	if !isPartitioned {
		// Parent table needs to be converted to partitioned table first
		result.Error = "Parent table 'clicks' is not partitioned. Run ConvertToPartitionedTable first."
		result.SQLExecuted = `
-- To convert clicks table to partitioned:
-- 1. Rename current table
ALTER TABLE clicks RENAME TO clicks_old;

-- 2. Create partitioned table
CREATE TABLE clicks (
    LIKE clicks_old INCLUDING ALL
) PARTITION BY RANGE (clicked_at);

-- 3. Create default partition for old data
CREATE TABLE clicks_default PARTITION OF clicks DEFAULT;

-- 4. Migrate data
INSERT INTO clicks SELECT * FROM clicks_old;

-- 5. Drop old table (after verification)
-- DROP TABLE clicks_old;
`
		return result, nil
	}

	// Create partition SQL
	sql := fmt.Sprintf(`
		CREATE TABLE %s PARTITION OF clicks
		FOR VALUES FROM ('%s') TO ('%s')
	`, partitionName, result.RangeStart, result.RangeEnd)

	result.SQLExecuted = sql

	// Execute
	if err := s.db.Exec(sql).Error; err != nil {
		result.Error = err.Error()
		return result, nil
	}

	// Create indexes on the new partition
	indexSQLs := []string{
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_user_offer ON %s (user_offer_id)", partitionName, partitionName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_fingerprint ON %s (fingerprint)", partitionName, partitionName),
		fmt.Sprintf("CREATE INDEX IF NOT EXISTS idx_%s_ip ON %s (ip_address)", partitionName, partitionName),
	}

	for _, indexSQL := range indexSQLs {
		if err := s.db.Exec(indexSQL).Error; err != nil {
			log.Printf("Warning: failed to create index: %v", err)
		}
	}

	result.Success = true
	log.Printf("✅ Created partition %s for %s to %s", partitionName, result.RangeStart, result.RangeEnd)
	
	return result, nil
}

// ============================================
// AUTOMATIC PARTITION MAINTENANCE
// ============================================

// EnsurePartitionsExist ensures partitions exist for current and next months
func (s *PartitionService) EnsurePartitionsExist() ([]CreatePartitionResult, error) {
	results := make([]CreatePartitionResult, 0)

	now := time.Now().UTC()
	
	// Current month
	currentResult, err := s.CreateMonthlyPartition(now.Year(), int(now.Month()))
	if err != nil {
		return results, err
	}
	results = append(results, *currentResult)

	// Next month
	nextMonth := now.AddDate(0, 1, 0)
	nextResult, err := s.CreateMonthlyPartition(nextMonth.Year(), int(nextMonth.Month()))
	if err != nil {
		return results, err
	}
	results = append(results, *nextResult)

	// Month after next (for safety)
	twoMonthsAhead := now.AddDate(0, 2, 0)
	futureResult, err := s.CreateMonthlyPartition(twoMonthsAhead.Year(), int(twoMonthsAhead.Month()))
	if err != nil {
		return results, err
	}
	results = append(results, *futureResult)

	return results, nil
}

// ============================================
// PARTITION MIGRATION HELPER
// ============================================

// MigrationPlan represents a plan to convert a table to partitioned
type MigrationPlan struct {
	TableName       string   `json:"table_name"`
	RowCount        int64    `json:"row_count"`
	TableSizeMB     float64  `json:"table_size_mb"`
	EstimatedTimeMs int64    `json:"estimated_time_ms"`
	Steps           []string `json:"steps"`
	SQLCommands     []string `json:"sql_commands"`
	Warnings        []string `json:"warnings"`
}

// GetMigrationPlan returns a plan for converting clicks to partitioned table
func (s *PartitionService) GetMigrationPlan() (*MigrationPlan, error) {
	plan := &MigrationPlan{
		TableName: "clicks",
		Steps:     make([]string, 0),
		SQLCommands: make([]string, 0),
		Warnings:  make([]string, 0),
	}

	// Get current row count
	var count int64
	s.db.Raw("SELECT COUNT(*) FROM clicks").Scan(&count)
	plan.RowCount = count

	// Get table size
	var sizeBytes int64
	s.db.Raw("SELECT pg_total_relation_size('clicks')").Scan(&sizeBytes)
	plan.TableSizeMB = float64(sizeBytes) / 1024 / 1024

	// Estimate time (rough)
	plan.EstimatedTimeMs = count / 10000 * 100 // 100ms per 10k rows

	// Check if already partitioned
	var isPartitioned bool
	s.db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_partitioned_table pt
			JOIN pg_class c ON c.oid = pt.partrelid
			WHERE c.relname = 'clicks'
		)
	`).Scan(&isPartitioned)

	if isPartitioned {
		plan.Steps = append(plan.Steps, "✅ Table is already partitioned")
		return plan, nil
	}

	// Build migration plan
	plan.Steps = []string{
		"1. Create backup of clicks table",
		"2. Rename clicks to clicks_old",
		"3. Create new partitioned clicks table",
		"4. Create partitions for historical and future months",
		"5. Migrate data from clicks_old to clicks",
		"6. Verify data integrity",
		"7. Drop clicks_old (after verification)",
	}

	plan.SQLCommands = []string{
		"-- Step 1: Backup (run pg_dump separately)",
		"-- pg_dump -t clicks dbname > clicks_backup.sql",
		"",
		"-- Step 2: Rename old table",
		"ALTER TABLE clicks RENAME TO clicks_old;",
		"",
		"-- Step 3: Create partitioned table",
		`CREATE TABLE clicks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_offer_id UUID NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    device VARCHAR(50),
    browser VARCHAR(50),
    os VARCHAR(50),
    country VARCHAR(100),
    city VARCHAR(100),
    referer TEXT,
    fingerprint VARCHAR(64),
    clicked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
) PARTITION BY RANGE (clicked_at);`,
		"",
		"-- Step 4: Create partitions",
		"CREATE TABLE clicks_2024_01 PARTITION OF clicks FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');",
		"CREATE TABLE clicks_2024_02 PARTITION OF clicks FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');",
		"-- ... continue for each month ...",
		"CREATE TABLE clicks_default PARTITION OF clicks DEFAULT;",
		"",
		"-- Step 5: Migrate data",
		"INSERT INTO clicks SELECT * FROM clicks_old;",
		"",
		"-- Step 6: Create indexes on partitions",
		"-- (Indexes are automatically created on child tables if defined on parent)",
		"",
		"-- Step 7: Verify and cleanup",
		"-- SELECT COUNT(*) FROM clicks;",
		"-- SELECT COUNT(*) FROM clicks_old;",
		"-- DROP TABLE clicks_old; -- Only after verification!",
	}

	plan.Warnings = []string{
		"⚠️ This operation requires downtime or careful online migration",
		"⚠️ Always backup data before migration",
		"⚠️ Test in staging environment first",
		"⚠️ Consider using pg_partman for automated partition management",
		"⚠️ Large tables may take significant time to migrate",
	}

	return plan, nil
}

// ============================================
// CLEANUP OLD PARTITIONS
// ============================================

// CleanupOldPartitions drops partitions older than specified months
func (s *PartitionService) CleanupOldPartitions(monthsToKeep int) ([]string, error) {
	if monthsToKeep < 3 {
		return nil, fmt.Errorf("monthsToKeep must be at least 3")
	}

	cutoffDate := time.Now().UTC().AddDate(0, -monthsToKeep, 0)
	cutoffName := fmt.Sprintf("clicks_%d_%02d", cutoffDate.Year(), cutoffDate.Month())

	// Get partitions to drop
	rows, err := s.db.Raw(`
		SELECT c.relname as partition_name
		FROM pg_inherits i
		JOIN pg_class c ON i.inhrelid = c.oid
		JOIN pg_class p ON i.inhparent = p.oid
		WHERE p.relname = 'clicks'
		AND c.relname < ?
		AND c.relname != 'clicks_default'
		ORDER BY c.relname
	`, cutoffName).Rows()

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var dropped []string
	for rows.Next() {
		var partitionName string
		if err := rows.Scan(&partitionName); err != nil {
			continue
		}

		// Drop partition
		sql := fmt.Sprintf("DROP TABLE IF EXISTS %s", partitionName)
		if err := s.db.Exec(sql).Error; err != nil {
			log.Printf("Failed to drop partition %s: %v", partitionName, err)
			continue
		}
		
		dropped = append(dropped, partitionName)
		log.Printf("🗑️ Dropped old partition: %s", partitionName)
	}

	return dropped, nil
}


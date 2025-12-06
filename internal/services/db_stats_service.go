package services

import (
	"fmt"
	"sort"
	"time"

	"gorm.io/gorm"
)

// ============================================
// DATABASE STATISTICS SERVICE
// ============================================

// DBStatsService provides database statistics and diagnostics
type DBStatsService struct {
	db *gorm.DB
}

// NewDBStatsService creates a new DB stats service
func NewDBStatsService(db *gorm.DB) *DBStatsService {
	return &DBStatsService{db: db}
}

// ============================================
// TABLE STATISTICS (pg_stat_user_tables)
// ============================================

// TableStats represents statistics for a table
type TableStats struct {
	TableName        string    `json:"table_name"`
	RowCount         int64     `json:"row_count"`
	DeadTuples       int64     `json:"dead_tuples"`
	LiveTuples       int64     `json:"live_tuples"`
	DeadTupleRatio   float64   `json:"dead_tuple_ratio_percent"`
	LastVacuum       *time.Time `json:"last_vacuum"`
	LastAutoVacuum   *time.Time `json:"last_autovacuum"`
	LastAnalyze      *time.Time `json:"last_analyze"`
	LastAutoAnalyze  *time.Time `json:"last_autoanalyze"`
	TableSizeMB      float64   `json:"table_size_mb"`
	IndexesSizeMB    float64   `json:"indexes_size_mb"`
	TotalSizeMB      float64   `json:"total_size_mb"`
	SeqScans         int64     `json:"seq_scans"`
	SeqTuplesRead    int64     `json:"seq_tuples_read"`
	IdxScans         int64     `json:"idx_scans"`
	IdxTuplesRead    int64     `json:"idx_tuples_read"`
	TuplesInserted   int64     `json:"tuples_inserted"`
	TuplesUpdated    int64     `json:"tuples_updated"`
	TuplesDeleted    int64     `json:"tuples_deleted"`
	NeedsVacuum      bool      `json:"needs_vacuum"`
	NeedsAnalyze     bool      `json:"needs_analyze"`
	BloatRatio       float64   `json:"bloat_ratio_percent"`
}

// GetTableStats returns statistics for all user tables
func (s *DBStatsService) GetTableStats() ([]TableStats, error) {
	query := `
		SELECT 
			schemaname || '.' || relname as table_name,
			n_live_tup as live_tuples,
			n_dead_tup as dead_tuples,
			CASE WHEN n_live_tup > 0 
				THEN ROUND((n_dead_tup::numeric / n_live_tup * 100), 2) 
				ELSE 0 
			END as dead_tuple_ratio,
			last_vacuum,
			last_autovacuum,
			last_analyze,
			last_autoanalyze,
			seq_scan as seq_scans,
			seq_tup_read as seq_tuples_read,
			idx_scan as idx_scans,
			idx_tup_fetch as idx_tuples_read,
			n_tup_ins as tuples_inserted,
			n_tup_upd as tuples_updated,
			n_tup_del as tuples_deleted,
			pg_table_size(relid) as table_bytes,
			pg_indexes_size(relid) as indexes_bytes,
			pg_total_relation_size(relid) as total_bytes
		FROM pg_stat_user_tables
		ORDER BY n_live_tup DESC
	`

	rows, err := s.db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []TableStats
	for rows.Next() {
		var stat TableStats
		var tableBytes, indexesBytes, totalBytes int64
		
		err := rows.Scan(
			&stat.TableName,
			&stat.LiveTuples,
			&stat.DeadTuples,
			&stat.DeadTupleRatio,
			&stat.LastVacuum,
			&stat.LastAutoVacuum,
			&stat.LastAnalyze,
			&stat.LastAutoAnalyze,
			&stat.SeqScans,
			&stat.SeqTuplesRead,
			&stat.IdxScans,
			&stat.IdxTuplesRead,
			&stat.TuplesInserted,
			&stat.TuplesUpdated,
			&stat.TuplesDeleted,
			&tableBytes,
			&indexesBytes,
			&totalBytes,
		)
		if err != nil {
			continue
		}

		stat.RowCount = stat.LiveTuples
		stat.TableSizeMB = float64(tableBytes) / 1024 / 1024
		stat.IndexesSizeMB = float64(indexesBytes) / 1024 / 1024
		stat.TotalSizeMB = float64(totalBytes) / 1024 / 1024
		
		// Determine if vacuum/analyze needed
		stat.NeedsVacuum = stat.DeadTupleRatio > 10 || (stat.LastVacuum == nil && stat.LastAutoVacuum == nil)
		stat.NeedsAnalyze = stat.LastAnalyze == nil && stat.LastAutoAnalyze == nil
		
		// Calculate bloat ratio
		if stat.LiveTuples > 0 {
			stat.BloatRatio = stat.DeadTupleRatio
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

// ============================================
// INDEX STATISTICS (pg_stat_user_indexes)
// ============================================

// IndexStats represents statistics for an index
type IndexStats struct {
	IndexName       string  `json:"index_name"`
	TableName       string  `json:"table_name"`
	IndexType       string  `json:"index_type"`
	IndexDef        string  `json:"index_definition"`
	SizeMB          float64 `json:"size_mb"`
	IdxScans        int64   `json:"idx_scans"`
	TuplesRead      int64   `json:"tuples_read"`
	TuplesFetched   int64   `json:"tuples_fetched"`
	IsUnique        bool    `json:"is_unique"`
	IsPrimary       bool    `json:"is_primary"`
	IsUnused        bool    `json:"is_unused"`
	UsageRatio      float64 `json:"usage_ratio"`
	DuplicationNote string  `json:"duplication_note,omitempty"`
}

// GetIndexStats returns statistics for all user indexes
func (s *DBStatsService) GetIndexStats() ([]IndexStats, error) {
	query := `
		SELECT 
			i.indexrelname as index_name,
			i.relname as table_name,
			am.amname as index_type,
			pg_get_indexdef(i.indexrelid) as index_def,
			pg_relation_size(i.indexrelid) as size_bytes,
			i.idx_scan as idx_scans,
			i.idx_tup_read as tuples_read,
			i.idx_tup_fetch as tuples_fetched,
			ix.indisunique as is_unique,
			ix.indisprimary as is_primary
		FROM pg_stat_user_indexes i
		JOIN pg_index ix ON i.indexrelid = ix.indexrelid
		JOIN pg_class c ON i.indexrelid = c.oid
		JOIN pg_am am ON c.relam = am.oid
		ORDER BY i.idx_scan ASC, size_bytes DESC
	`

	rows, err := s.db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []IndexStats
	for rows.Next() {
		var stat IndexStats
		var sizeBytes int64
		
		err := rows.Scan(
			&stat.IndexName,
			&stat.TableName,
			&stat.IndexType,
			&stat.IndexDef,
			&sizeBytes,
			&stat.IdxScans,
			&stat.TuplesRead,
			&stat.TuplesFetched,
			&stat.IsUnique,
			&stat.IsPrimary,
		)
		if err != nil {
			continue
		}

		stat.SizeMB = float64(sizeBytes) / 1024 / 1024
		stat.IsUnused = stat.IdxScans == 0 && !stat.IsPrimary && !stat.IsUnique
		
		// Calculate usage ratio
		if stat.TuplesRead > 0 {
			stat.UsageRatio = float64(stat.TuplesFetched) / float64(stat.TuplesRead) * 100
		}

		stats = append(stats, stat)
	}

	// Check for potential duplicates
	s.detectDuplicateIndexes(stats)

	return stats, nil
}

// detectDuplicateIndexes marks potentially duplicate indexes
func (s *DBStatsService) detectDuplicateIndexes(stats []IndexStats) {
	// Group indexes by table
	tableIndexes := make(map[string][]int)
	for i, stat := range stats {
		tableIndexes[stat.TableName] = append(tableIndexes[stat.TableName], i)
	}

	// Check each table for potential duplicates
	for _, indexes := range tableIndexes {
		for i := 0; i < len(indexes); i++ {
			for j := i + 1; j < len(indexes); j++ {
				idx1 := &stats[indexes[i]]
				idx2 := &stats[indexes[j]]

				// Simple heuristic: if one index covers another
				if idx1.IdxScans == 0 && idx2.IdxScans > 0 {
					idx1.DuplicationNote = fmt.Sprintf("Possibly redundant, consider %s", idx2.IndexName)
				} else if idx2.IdxScans == 0 && idx1.IdxScans > 0 {
					idx2.DuplicationNote = fmt.Sprintf("Possibly redundant, consider %s", idx1.IndexName)
				}
			}
		}
	}
}

// GetUnusedIndexes returns indexes that have never been scanned
func (s *DBStatsService) GetUnusedIndexes() ([]IndexStats, error) {
	allIndexes, err := s.GetIndexStats()
	if err != nil {
		return nil, err
	}

	var unused []IndexStats
	for _, idx := range allIndexes {
		if idx.IsUnused {
			unused = append(unused, idx)
		}
	}

	return unused, nil
}

// ============================================
// VACUUM / ANALYZE PLAN
// ============================================

// VacuumPlan represents a recommended vacuum/analyze plan
type VacuumPlan struct {
	GeneratedAt     time.Time         `json:"generated_at"`
	TablesNeedVacuum []VacuumTask     `json:"tables_need_vacuum"`
	TablesNeedAnalyze []AnalyzeTask   `json:"tables_need_analyze"`
	CronSchedule    string            `json:"recommended_cron_schedule"`
	SQLCommands     []string          `json:"sql_commands"`
	Notes           []string          `json:"notes"`
}

// VacuumTask represents a vacuum task
type VacuumTask struct {
	TableName       string    `json:"table_name"`
	DeadTuples      int64     `json:"dead_tuples"`
	DeadTupleRatio  float64   `json:"dead_tuple_ratio_percent"`
	LastVacuum      *time.Time `json:"last_vacuum"`
	Priority        string    `json:"priority"` // high, medium, low
	EstimatedTimeMs int64     `json:"estimated_time_ms"`
}

// AnalyzeTask represents an analyze task
type AnalyzeTask struct {
	TableName       string    `json:"table_name"`
	LastAnalyze     *time.Time `json:"last_analyze"`
	RowCount        int64     `json:"row_count"`
	Priority        string    `json:"priority"`
}

// GetVacuumPlan returns a recommended vacuum/analyze plan
func (s *DBStatsService) GetVacuumPlan() (*VacuumPlan, error) {
	stats, err := s.GetTableStats()
	if err != nil {
		return nil, err
	}

	plan := &VacuumPlan{
		GeneratedAt:       time.Now().UTC(),
		TablesNeedVacuum:  make([]VacuumTask, 0),
		TablesNeedAnalyze: make([]AnalyzeTask, 0),
		CronSchedule:      "0 3 * * * (Daily at 3 AM)",
		SQLCommands:       make([]string, 0),
		Notes: []string{
			"VACUUM reclaims storage from dead tuples",
			"ANALYZE updates statistics for query planner",
			"Consider VACUUM FULL for heavily bloated tables (requires exclusive lock)",
			"Run during low-traffic periods",
		},
	}

	for _, stat := range stats {
		// Check if vacuum needed
		if stat.NeedsVacuum || stat.DeadTupleRatio > 5 {
			priority := "low"
			if stat.DeadTupleRatio > 20 {
				priority = "high"
			} else if stat.DeadTupleRatio > 10 {
				priority = "medium"
			}

			task := VacuumTask{
				TableName:       stat.TableName,
				DeadTuples:      stat.DeadTuples,
				DeadTupleRatio:  stat.DeadTupleRatio,
				LastVacuum:      stat.LastVacuum,
				Priority:        priority,
				EstimatedTimeMs: stat.DeadTuples / 10000 * 100, // rough estimate
			}
			plan.TablesNeedVacuum = append(plan.TablesNeedVacuum, task)
			plan.SQLCommands = append(plan.SQLCommands, 
				fmt.Sprintf("VACUUM (VERBOSE, ANALYZE) %s;", stat.TableName))
		}

		// Check if analyze needed
		if stat.NeedsAnalyze {
			priority := "medium"
			if stat.RowCount > 100000 {
				priority = "high"
			}

			task := AnalyzeTask{
				TableName:   stat.TableName,
				LastAnalyze: stat.LastAnalyze,
				RowCount:    stat.RowCount,
				Priority:    priority,
			}
			plan.TablesNeedAnalyze = append(plan.TablesNeedAnalyze, task)
			if !stat.NeedsVacuum {
				plan.SQLCommands = append(plan.SQLCommands,
					fmt.Sprintf("ANALYZE %s;", stat.TableName))
			}
		}
	}

	// Sort by priority
	sort.Slice(plan.TablesNeedVacuum, func(i, j int) bool {
		return priorityOrder(plan.TablesNeedVacuum[i].Priority) > 
			   priorityOrder(plan.TablesNeedVacuum[j].Priority)
	})

	return plan, nil
}

func priorityOrder(p string) int {
	switch p {
	case "high":
		return 3
	case "medium":
		return 2
	case "low":
		return 1
	default:
		return 0
	}
}

// ============================================
// BACKUP INFO (PITR Readiness)
// ============================================

// BackupInfo represents backup configuration info
type BackupInfo struct {
	WALLevel         string `json:"wal_level"`
	ArchiveMode      string `json:"archive_mode"`
	ArchiveCommand   string `json:"archive_command"`
	MaxWALSenders    int    `json:"max_wal_senders"`
	WalKeepSize      string `json:"wal_keep_size"`
	PITRReady        bool   `json:"pitr_ready"`
	PITRIssues       []string `json:"pitr_issues,omitempty"`
	Recommendations  []string `json:"recommendations"`
}

// GetBackupInfo returns backup configuration info
func (s *DBStatsService) GetBackupInfo() (*BackupInfo, error) {
	info := &BackupInfo{
		Recommendations: make([]string, 0),
		PITRIssues:      make([]string, 0),
	}

	// Get WAL level
	var walLevel string
	s.db.Raw("SHOW wal_level").Scan(&walLevel)
	info.WALLevel = walLevel

	// Get archive mode
	var archiveMode string
	s.db.Raw("SHOW archive_mode").Scan(&archiveMode)
	info.ArchiveMode = archiveMode

	// Get archive command
	var archiveCommand string
	s.db.Raw("SHOW archive_command").Scan(&archiveCommand)
	info.ArchiveCommand = archiveCommand

	// Get max_wal_senders
	var maxWALSenders int
	s.db.Raw("SHOW max_wal_senders").Scan(&maxWALSenders)
	info.MaxWALSenders = maxWALSenders

	// Get wal_keep_size
	var walKeepSize string
	s.db.Raw("SHOW wal_keep_size").Scan(&walKeepSize)
	info.WalKeepSize = walKeepSize

	// Check PITR readiness
	info.PITRReady = true

	if walLevel != "replica" && walLevel != "logical" {
		info.PITRReady = false
		info.PITRIssues = append(info.PITRIssues, 
			"wal_level must be 'replica' or 'logical' for PITR")
		info.Recommendations = append(info.Recommendations,
			"ALTER SYSTEM SET wal_level = 'replica'; -- Requires restart")
	}

	if archiveMode != "on" {
		info.PITRReady = false
		info.PITRIssues = append(info.PITRIssues,
			"archive_mode must be 'on' for PITR")
		info.Recommendations = append(info.Recommendations,
			"ALTER SYSTEM SET archive_mode = 'on'; -- Requires restart")
	}

	if archiveCommand == "" || archiveCommand == "(disabled)" {
		info.PITRReady = false
		info.PITRIssues = append(info.PITRIssues,
			"archive_command is not configured")
		info.Recommendations = append(info.Recommendations,
			"ALTER SYSTEM SET archive_command = 'cp %p /path/to/archive/%f';")
	}

	if maxWALSenders < 2 {
		info.Recommendations = append(info.Recommendations,
			"Consider increasing max_wal_senders for replication")
	}

	if info.PITRReady {
		info.Recommendations = append(info.Recommendations,
			"PITR is configured correctly. Remember to regularly test recovery!")
	}

	return info, nil
}

// ============================================
// SLOW QUERIES
// ============================================

// SlowQuery represents a slow query
type SlowQuery struct {
	QueryID          string  `json:"query_id"`
	Query            string  `json:"query"`
	Calls            int64   `json:"calls"`
	TotalTimeMs      float64 `json:"total_time_ms"`
	MeanTimeMs       float64 `json:"mean_time_ms"`
	MinTimeMs        float64 `json:"min_time_ms"`
	MaxTimeMs        float64 `json:"max_time_ms"`
	StddevTimeMs     float64 `json:"stddev_time_ms"`
	Rows             int64   `json:"rows"`
	SharedBlksHit    int64   `json:"shared_blks_hit"`
	SharedBlksRead   int64   `json:"shared_blks_read"`
	CacheHitRatio    float64 `json:"cache_hit_ratio"`
}

// GetSlowQueries returns slow queries (requires pg_stat_statements)
func (s *DBStatsService) GetSlowQueries(limit int) ([]SlowQuery, error) {
	// Check if pg_stat_statements is available
	var exists bool
	s.db.Raw(`
		SELECT EXISTS (
			SELECT 1 FROM pg_extension WHERE extname = 'pg_stat_statements'
		)
	`).Scan(&exists)

	if !exists {
		return nil, fmt.Errorf("pg_stat_statements extension not installed. Run: CREATE EXTENSION pg_stat_statements;")
	}

	if limit <= 0 {
		limit = 20
	}

	query := fmt.Sprintf(`
		SELECT 
			queryid::text as query_id,
			query,
			calls,
			total_exec_time as total_time_ms,
			mean_exec_time as mean_time_ms,
			min_exec_time as min_time_ms,
			max_exec_time as max_time_ms,
			stddev_exec_time as stddev_time_ms,
			rows,
			shared_blks_hit,
			shared_blks_read,
			CASE WHEN shared_blks_hit + shared_blks_read > 0
				THEN shared_blks_hit::float / (shared_blks_hit + shared_blks_read) * 100
				ELSE 0
			END as cache_hit_ratio
		FROM pg_stat_statements
		WHERE query NOT LIKE '%%pg_stat%%'
		ORDER BY mean_exec_time DESC
		LIMIT %d
	`, limit)

	rows, err := s.db.Raw(query).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var queries []SlowQuery
	for rows.Next() {
		var q SlowQuery
		err := rows.Scan(
			&q.QueryID,
			&q.Query,
			&q.Calls,
			&q.TotalTimeMs,
			&q.MeanTimeMs,
			&q.MinTimeMs,
			&q.MaxTimeMs,
			&q.StddevTimeMs,
			&q.Rows,
			&q.SharedBlksHit,
			&q.SharedBlksRead,
			&q.CacheHitRatio,
		)
		if err != nil {
			continue
		}
		queries = append(queries, q)
	}

	return queries, nil
}

// ============================================
// DATABASE SIZE
// ============================================

// DBSize represents database size info
type DBSize struct {
	DatabaseName   string             `json:"database_name"`
	TotalSizeMB    float64            `json:"total_size_mb"`
	TotalSizeGB    float64            `json:"total_size_gb"`
	TablesSizeMB   float64            `json:"tables_size_mb"`
	IndexesSizeMB  float64            `json:"indexes_size_mb"`
	TopTables      []TableSizeInfo    `json:"top_tables"`
}

// TableSizeInfo represents size info for a table
type TableSizeInfo struct {
	TableName     string  `json:"table_name"`
	RowCount      int64   `json:"row_count"`
	TableSizeMB   float64 `json:"table_size_mb"`
	IndexesSizeMB float64 `json:"indexes_size_mb"`
	TotalSizeMB   float64 `json:"total_size_mb"`
	Percentage    float64 `json:"percentage"`
}

// GetDBSize returns database size information
func (s *DBStatsService) GetDBSize() (*DBSize, error) {
	var dbName string
	s.db.Raw("SELECT current_database()").Scan(&dbName)

	var totalBytes int64
	s.db.Raw("SELECT pg_database_size(current_database())").Scan(&totalBytes)

	size := &DBSize{
		DatabaseName: dbName,
		TotalSizeMB:  float64(totalBytes) / 1024 / 1024,
		TotalSizeGB:  float64(totalBytes) / 1024 / 1024 / 1024,
	}

	// Get table sizes
	query := `
		SELECT 
			schemaname || '.' || relname as table_name,
			n_live_tup as row_count,
			pg_table_size(relid) as table_bytes,
			pg_indexes_size(relid) as indexes_bytes,
			pg_total_relation_size(relid) as total_bytes
		FROM pg_stat_user_tables
		ORDER BY pg_total_relation_size(relid) DESC
		LIMIT 10
	`

	rows, err := s.db.Raw(query).Rows()
	if err != nil {
		return size, nil
	}
	defer rows.Close()

	var tablesSizeBytes, indexesSizeBytes int64
	for rows.Next() {
		var info TableSizeInfo
		var tableBytes, indexesBytes, tableTotalBytes int64
		
		err := rows.Scan(
			&info.TableName,
			&info.RowCount,
			&tableBytes,
			&indexesBytes,
			&tableTotalBytes,
		)
		if err != nil {
			continue
		}

		info.TableSizeMB = float64(tableBytes) / 1024 / 1024
		info.IndexesSizeMB = float64(indexesBytes) / 1024 / 1024
		info.TotalSizeMB = float64(tableTotalBytes) / 1024 / 1024
		if totalBytes > 0 {
			info.Percentage = float64(tableTotalBytes) / float64(totalBytes) * 100
		}

		tablesSizeBytes += tableBytes
		indexesSizeBytes += indexesBytes
		size.TopTables = append(size.TopTables, info)
	}

	size.TablesSizeMB = float64(tablesSizeBytes) / 1024 / 1024
	size.IndexesSizeMB = float64(indexesSizeBytes) / 1024 / 1024

	return size, nil
}


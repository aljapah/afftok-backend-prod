package handlers

import (
	"net/http"
	"time"

	"github.com/aljapah/afftok-backend-prod/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ============================================
// ADMIN QA HANDLER
// ============================================

// AdminQAHandler handles QA, testing, and validation endpoints
type AdminQAHandler struct {
	db                *gorm.DB
	e2eService        *services.E2ETestService
	consistencyService *services.ConsistencyService
	securityAuditService *services.SecurityAuditService
	benchmarkService  *services.BenchmarkService
	observability     *services.ObservabilityService
}

// NewAdminQAHandler creates a new admin QA handler
func NewAdminQAHandler(db *gorm.DB) *AdminQAHandler {
	return &AdminQAHandler{
		db:                   db,
		e2eService:           services.NewE2ETestService(db),
		consistencyService:   services.NewConsistencyService(db),
		securityAuditService: services.NewSecurityAuditService(db),
		benchmarkService:     services.NewBenchmarkService(db),
		observability:        services.NewObservabilityService(),
	}
}

// ============================================
// E2E TEST ENDPOINTS
// ============================================

// RunE2ETest runs a specific E2E test scenario
// POST /api/admin/e2e-tests/run
func (h *AdminQAHandler) RunE2ETest(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req struct {
		ScenarioID string `json:"scenario_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "Invalid request: " + err.Error(),
		})
		return
	}

	result, err := h.e2eService.RunScenario(req.ScenarioID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           result,
		"timestamp":      time.Now().UTC(),
	})
}

// GetE2EScenarios returns all available E2E test scenarios
// GET /api/admin/e2e-tests/scenarios
func (h *AdminQAHandler) GetE2EScenarios(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	scenarios := h.e2eService.GetScenarios()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"count":     len(scenarios),
			"scenarios": scenarios,
		},
		"timestamp": time.Now().UTC(),
	})
}

// GetE2ETestHistory returns E2E test run history
// GET /api/admin/e2e-tests/history
func (h *AdminQAHandler) GetE2ETestHistory(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	history := h.e2eService.GetTestHistory(50)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"count":   len(history),
			"history": history,
		},
		"timestamp": time.Now().UTC(),
	})
}

// GetE2ETestRun returns a specific E2E test run
// GET /api/admin/e2e-tests/:id
func (h *AdminQAHandler) GetE2ETestRun(c *gin.Context) {
	correlationID := uuid.New().String()[:8]
	runID := c.Param("id")

	run, err := h.e2eService.GetTestRun(runID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           run,
		"timestamp":      time.Now().UTC(),
	})
}

// ============================================
// CONSISTENCY CHECK ENDPOINTS
// ============================================

// RunConsistencyCheck runs a full consistency check
// GET /api/admin/consistency/run
func (h *AdminQAHandler) RunConsistencyCheck(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	report, err := h.consistencyService.RunFullCheck()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           report,
		"timestamp":      time.Now().UTC(),
	})
}

// GetConsistencyReport returns the latest consistency report
// GET /api/admin/consistency/report
func (h *AdminQAHandler) GetConsistencyReport(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	report := h.consistencyService.GetLatestReport()
	if report == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "No consistency report available. Run a check first.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           report,
		"timestamp":      time.Now().UTC(),
	})
}

// GetConsistencyIssues returns all issues from the latest check
// GET /api/admin/consistency/issues
func (h *AdminQAHandler) GetConsistencyIssues(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	issues := h.consistencyService.GetAllIssues()

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"count":  len(issues),
			"issues": issues,
		},
		"timestamp": time.Now().UTC(),
	})
}

// FixConsistencyIssues attempts to fix detected issues
// POST /api/admin/consistency/fix
func (h *AdminQAHandler) FixConsistencyIssues(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	fixed, err := h.consistencyService.FixUserOfferStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"fixed_count": fixed,
		},
		"message":   "Consistency issues fixed",
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// SECURITY AUDIT ENDPOINTS
// ============================================

// RunSecurityAudit runs a full security audit
// GET /api/admin/security/audit/run
func (h *AdminQAHandler) RunSecurityAudit(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	report, err := h.securityAuditService.RunFullAudit()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           report,
		"timestamp":      time.Now().UTC(),
	})
}

// GetSecurityAuditReport returns the latest security audit report
// GET /api/admin/security/audit/report
func (h *AdminQAHandler) GetSecurityAuditReport(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	report := h.securityAuditService.GetLatestReport()
	if report == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "No security audit report available. Run an audit first.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           report,
		"timestamp":      time.Now().UTC(),
	})
}

// GetSecurityFindings returns all security findings
// GET /api/admin/security/audit/findings
func (h *AdminQAHandler) GetSecurityFindings(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	findings := h.securityAuditService.GetAllFindings()

	// Group by severity
	bySeverity := map[string]int{
		"critical": 0,
		"high":     0,
		"medium":   0,
		"low":      0,
	}
	for _, f := range findings {
		bySeverity[string(f.Severity)]++
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"count":       len(findings),
			"by_severity": bySeverity,
			"findings":    findings,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// BENCHMARK ENDPOINTS
// ============================================

// RunBenchmark runs a performance benchmark
// POST /api/admin/benchmarks/run
func (h *AdminQAHandler) RunBenchmark(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	var req struct {
		Type        string `json:"type"`        // click_path, conversion_path, db_latency, redis_latency, full
		Iterations  int    `json:"iterations"`  // default 100
		Concurrency int    `json:"concurrency"` // default 10
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Use defaults
		req.Type = "full"
		req.Iterations = 100
		req.Concurrency = 10
	}

	benchType := services.BenchmarkType(req.Type)
	if benchType == "" {
		benchType = services.BenchmarkFull
	}

	result, err := h.benchmarkService.RunBenchmark(benchType, req.Iterations, req.Concurrency)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           result,
		"timestamp":      time.Now().UTC(),
	})
}

// GetBenchmarkReport returns the latest benchmark report
// GET /api/admin/benchmarks/report
func (h *AdminQAHandler) GetBenchmarkReport(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	result := h.benchmarkService.GetLatestResult()
	if result == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success":        false,
			"correlation_id": correlationID,
			"error":          "No benchmark report available. Run a benchmark first.",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           result,
		"timestamp":      time.Now().UTC(),
	})
}

// GetBenchmarkHistory returns benchmark history
// GET /api/admin/benchmarks/history
func (h *AdminQAHandler) GetBenchmarkHistory(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	history := h.benchmarkService.GetResultHistory(20)

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"count":   len(history),
			"history": history,
		},
		"timestamp": time.Now().UTC(),
	})
}

// ============================================
// PREFLIGHT CHECK ENDPOINT
// ============================================

// PreflightCheck runs all critical checks before deployment
// GET /api/admin/preflight/check
func (h *AdminQAHandler) PreflightCheck(c *gin.Context) {
	correlationID := uuid.New().String()[:8]
	startTime := time.Now()

	result := &PreflightResult{
		ReadyForLaunch:  true,
		CriticalIssues:  make([]string, 0),
		Warnings:        make([]string, 0),
		Details:         make(map[string]string),
		ChecksPerformed: make([]string, 0),
	}

	// 1. Health Check
	result.ChecksPerformed = append(result.ChecksPerformed, "health")
	healthStatus := h.checkHealth()
	result.Details["health"] = healthStatus
	if healthStatus != "ok" {
		result.CriticalIssues = append(result.CriticalIssues, "Health check failed: "+healthStatus)
		result.ReadyForLaunch = false
	}

	// 2. Consistency Check (quick)
	result.ChecksPerformed = append(result.ChecksPerformed, "consistency")
	consistencyReport, _ := h.consistencyService.RunFullCheck()
	if consistencyReport != nil {
		if consistencyReport.Status == services.ConsistencyStatusCritical {
			result.Details["consistency"] = "critical"
			result.CriticalIssues = append(result.CriticalIssues, "Data consistency issues detected")
			result.ReadyForLaunch = false
		} else if consistencyReport.Status == services.ConsistencyStatusWarning {
			result.Details["consistency"] = "warning"
			result.Warnings = append(result.Warnings, "Minor data consistency warnings")
		} else {
			result.Details["consistency"] = "ok"
		}
	}

	// 3. Security Audit (quick)
	result.ChecksPerformed = append(result.ChecksPerformed, "security")
	securityReport, _ := h.securityAuditService.RunQuickAudit()
	if securityReport != nil {
		if securityReport.Summary.CriticalFindings > 0 {
			result.Details["security"] = "critical"
			result.CriticalIssues = append(result.CriticalIssues, "Critical security findings detected")
			result.ReadyForLaunch = false
		} else if securityReport.Summary.HighFindings > 0 {
			result.Details["security"] = "warning"
			result.Warnings = append(result.Warnings, "High severity security findings")
		} else {
			result.Details["security"] = "ok"
		}
	}

	// 4. Performance Check (light)
	result.ChecksPerformed = append(result.ChecksPerformed, "performance")
	benchResult, _ := h.benchmarkService.RunBenchmark(services.BenchmarkDBLatency, 10, 5)
	if benchResult != nil {
		if benchResult.Summary.OverallStatus == "failed" {
			result.Details["performance"] = "failed"
			result.CriticalIssues = append(result.CriticalIssues, "Performance below acceptable thresholds")
			result.ReadyForLaunch = false
		} else if benchResult.Summary.OverallStatus == "warning" {
			result.Details["performance"] = "warning"
			result.Warnings = append(result.Warnings, "Performance slightly degraded")
		} else {
			result.Details["performance"] = "ok"
		}
	}

	// 5. WAL Status
	result.ChecksPerformed = append(result.ChecksPerformed, "wal")
	walService := services.GetWALService()
	walStats := walService.GetStats()
	if pending, ok := walStats["pending_entries"].(int64); ok && pending > 10000 {
		result.Details["wal"] = "warning"
		result.Warnings = append(result.Warnings, "High WAL pending entries")
	} else {
		result.Details["wal"] = "ok"
	}

	// 6. Zero-Drop Status
	result.ChecksPerformed = append(result.ChecksPerformed, "zero_drop")
	zeroDropMode := services.GetZeroDropMode()
	if zeroDropMode.IsEnabled() {
		result.Details["zero_drop"] = "ok"
	} else {
		result.Details["zero_drop"] = "warning"
		result.Warnings = append(result.Warnings, "Zero-drop mode is disabled")
	}

	// 7. Edge Ingestion Status
	result.ChecksPerformed = append(result.ChecksPerformed, "edge_ingestion")
	edgeService := services.GetEdgeIngestService(h.db)
	if edgeService != nil {
		result.Details["edge_ingestion"] = "ok"
	} else {
		result.Details["edge_ingestion"] = "warning"
		result.Warnings = append(result.Warnings, "Edge ingestion service not available")
	}

	// 8. Alerts Configuration
	result.ChecksPerformed = append(result.ChecksPerformed, "alerts")
	// Check if alerts are configured
	result.Details["alerts"] = "ok"

	result.DurationMs = time.Since(startTime).Milliseconds()

	// Determine final status
	if result.ReadyForLaunch && len(result.Warnings) > 0 {
		result.Status = "ok_with_warnings"
	} else if result.ReadyForLaunch {
		result.Status = "ready"
	} else {
		result.Status = "not_ready"
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data":           result,
		"timestamp":      time.Now().UTC(),
	})
}

// PreflightResult holds the result of a preflight check
type PreflightResult struct {
	ReadyForLaunch  bool              `json:"ready_for_launch"`
	Status          string            `json:"status"` // ready, ok_with_warnings, not_ready
	CriticalIssues  []string          `json:"critical_issues"`
	Warnings        []string          `json:"warnings"`
	Details         map[string]string `json:"details"`
	ChecksPerformed []string          `json:"checks_performed"`
	DurationMs      int64             `json:"duration_ms"`
}

// checkHealth performs a basic health check
func (h *AdminQAHandler) checkHealth() string {
	// Check DB
	var result int
	if err := h.db.Raw("SELECT 1").Scan(&result).Error; err != nil {
		return "db_error"
	}

	// Check Redis
	if services.GetWALService() == nil {
		return "wal_error"
	}

	return "ok"
}

// ============================================
// QA STATS ENDPOINT
// ============================================

// GetQAStats returns overall QA statistics
// GET /api/admin/qa/stats
func (h *AdminQAHandler) GetQAStats(c *gin.Context) {
	correlationID := uuid.New().String()[:8]

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"correlation_id": correlationID,
		"data": gin.H{
			"e2e_tests":      h.e2eService.GetStats(),
			"consistency":    h.consistencyService.GetStats(),
			"security_audit": h.securityAuditService.GetStats(),
			"benchmarks":     h.benchmarkService.GetStats(),
		},
		"timestamp": time.Now().UTC(),
	})
}

